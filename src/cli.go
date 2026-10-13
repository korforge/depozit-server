package main

import (
	"depozit-server/src/config"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func CliLoadConfig() *config.Config {
	envPath := flag.String("env", ".env", "Path to .env file with parameters")

	flag.Parse()

	err := godotenv.Load(*envPath)

	if err != nil {
		fmt.Printf("Warning: could not load %s: %v\n", *envPath, err)
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
		host = "0.0.0.0"
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

	if len(errs) > 0 {
		return nil, errs
	}

	return config.LoadConfig(host, port, apiKey, storagePath, encryptionKey), nil
}
