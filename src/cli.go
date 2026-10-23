package main

import (
	"depozit-server/src/config"
	"encoding/hex"
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

func CliLoadConfig() *config.Config {
	envPath := flag.String("env", ".env", "Path to .env file with parameters")
	maxFileSize := flag.String("max_file_size", "", "Largest upload in bytes, or in MB with an MB suffix (100MB), 0 for no limit (default 100MB). Overrides MAX_FILE_SIZE")

	flag.Parse()

	err := godotenv.Load(*envPath)

	if err != nil {
		fmt.Printf("Warning: could not load %s: %v\n", *envPath, err)
	}

	if err := loadKeysFile(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	// Set after the env files load, so the flag wins over them and goes through the same validation
	if *maxFileSize != "" {
		os.Setenv("MAX_FILE_SIZE", *maxFileSize)
	}

	cfg, errs := ConfigFromEnv()

	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Println("Error:", e)
		}

		fmt.Println()
		flag.Usage()
		os.Exit(1)
	}

	return cfg
}

// Reads the config from environment variables and returns every validation error
func ConfigFromEnv() (*config.Config, []string) {
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	apiKey := os.Getenv("ROOT_API_KEY")
	rawEncryptionKey := os.Getenv("ENCRYPTION_KEY")
	storagePath := os.Getenv("STORAGE_PATH")

	var errs []string

	if host == "" {
		host = "127.0.0.1"
	}

	portNumber, err := strconv.Atoi(port)

	if err != nil || portNumber < 1 || portNumber > 65535 {
		errs = append(errs, "PORT must be a number between 1 and 65535")
	}

	if apiKey == "" {
		errs = append(errs, "ROOT_API_KEY is required")
	}

	if storagePath == "" {
		errs = append(errs, "STORAGE_PATH is required")
	}

	var encryptionKey []byte

	if rawEncryptionKey == "" {
		errs = append(errs, "ENCRYPTION_KEY is required. You can use 'openssl rand -hex 32' to generate a key")
	} else {
		encryptionKey, err = hex.DecodeString(rawEncryptionKey)

		if err != nil || len(encryptionKey) != 32 {
			errs = append(errs, "ENCRYPTION_KEY must be 64 hex chars (32 bytes). You can use 'openssl rand -hex 32' to generate a key")
		}
	}

	rawMaxFileSize := os.Getenv("MAX_FILE_SIZE")

	if rawMaxFileSize == "" {
		rawMaxFileSize = "100MB"
	}

	maxFileSize, err := parseSize(rawMaxFileSize)

	if err != nil || maxFileSize < 0 {
		errs = append(errs, "MAX_FILE_SIZE must be a number of bytes, or of MB with an MB suffix (100MB), 0 for no limit")
	}

	tlsCert := os.Getenv("TLS_CERT")
	tlsKey := os.Getenv("TLS_KEY")

	if (tlsCert == "") != (tlsKey == "") {
		errs = append(errs, "TLS_CERT and TLS_KEY must be set together, or both left empty to run without TLS")
	}

	maxConcurrentStreams, err := envInt("MAX_CONCURRENT_STREAMS", 100)

	if err != nil || maxConcurrentStreams < 1 || maxConcurrentStreams > math.MaxUint32 {
		errs = append(errs, "MAX_CONCURRENT_STREAMS must be a positive number")
	}

	if len(errs) > 0 {
		return nil, errs
	}

	cfg := config.LoadConfig(host, port, apiKey, storagePath, encryptionKey)

	cfg.MaxFileSize = maxFileSize
	cfg.MaxConcurrentStreams = uint32(maxConcurrentStreams)
	cfg.TLSCert = tlsCert
	cfg.TLSKey = tlsKey

	return cfg, nil
}

// Loads the secrets from the file named by KEYS_FILE, so they live apart from the config.
// Like ssh with private keys, it refuses a file that other users can read
func loadKeysFile() error {
	path := os.Getenv("KEYS_FILE")

	if path == "" {
		return nil
	}

	info, err := os.Stat(path)

	if err != nil {
		return fmt.Errorf("KEYS_FILE: %v", err)
	}

	if info.Mode().Perm()&0077 != 0 {
		return fmt.Errorf("KEYS_FILE %s is readable by other users, run: chmod 600 %s", path, path)
	}

	keys, err := godotenv.Read(path)

	if err != nil {
		return fmt.Errorf("KEYS_FILE: %v", err)
	}

	for name, value := range keys {
		// A key in both files is ambiguous, and one of them is likely stale
		if os.Getenv(name) != "" {
			return fmt.Errorf("%s is set in both KEYS_FILE and the environment, keep it only in %s", name, path)
		}

		os.Setenv(name, value)
	}

	return nil
}

// Parses a size in bytes, or in MB (1 MB = 1048576 bytes) when it ends in MB. Empty means 0
func parseSize(value string) (int64, error) {
	if value == "" {
		return 0, nil
	}

	multiplier := int64(1)

	if n, ok := strings.CutSuffix(strings.ToUpper(value), "MB"); ok {
		value = strings.TrimSpace(n)
		multiplier = 1 << 20
	}

	size, err := strconv.ParseInt(value, 10, 64)

	if err != nil {
		return 0, err
	}

	if size > math.MaxInt64/multiplier {
		return 0, strconv.ErrRange
	}

	return size * multiplier, nil
}

// Reads an optional integer variable
func envInt(name string, fallback int64) (int64, error) {
	value := os.Getenv(name)

	if value == "" {
		return fallback, nil
	}

	return strconv.ParseInt(value, 10, 64)
}
