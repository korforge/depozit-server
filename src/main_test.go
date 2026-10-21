package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"depozit-server/src/config"
)

const validKey = "62cc8e8ad7741788874fb42d493709eb2e66811403f12966e2814c72f9cb1946"

// The logger is global, set it once so servers left over from earlier tests never race with a new one
func TestMain(m *testing.M) {
	config.SetupLogger()

	os.Exit(m.Run())
}

func setEnv(t *testing.T, port string, apiKey string, storagePath string, encryptionKey string) {
	t.Setenv("HOST", "")
	t.Setenv("PORT", port)
	t.Setenv("ROOT_API_KEY", apiKey)
	t.Setenv("STORAGE_PATH", storagePath)
	t.Setenv("ENCRYPTION_KEY", encryptionKey)
	t.Setenv("KEYS_FILE", "")
	t.Setenv("TLS_CERT", "")
	t.Setenv("TLS_KEY", "")
	t.Setenv("MAX_FILE_SIZE", "")
	t.Setenv("MAX_CONCURRENT_STREAMS", "")
}

func TestConfigFromEnv(t *testing.T) {
	setEnv(t, "5051", "123456", "/data", validKey)

	cfg, errs := ConfigFromEnv()

	require.Empty(t, errs)
	assert.Equal(t, "127.0.0.1", cfg.Host)
	assert.Equal(t, "5051", cfg.Port)
	assert.Len(t, cfg.EncryptionKey, 32)
}

func TestConfigFromEnvShortKey(t *testing.T) {
	setEnv(t, "5051", "123456", "/data", validKey[:32])

	_, errs := ConfigFromEnv()

	assert.Len(t, errs, 1)
}

func TestConfigFromEnvMissing(t *testing.T) {
	setEnv(t, "", "", "", "")

	_, errs := ConfigFromEnv()

	// One error per variable, not two for the empty key
	assert.Len(t, errs, 4)
}

func TestConfigFromEnvInvalidPort(t *testing.T) {
	for _, port := range []string{"abc", "0", "70000"} {
		setEnv(t, port, "123456", "/data", validKey)

		_, errs := ConfigFromEnv()

		assert.Len(t, errs, 1, port)
	}
}

func TestConfigFromEnvOptional(t *testing.T) {
	setEnv(t, "5051", "123456", "/data", validKey)
	t.Setenv("TLS_CERT", "cert.pem")
	t.Setenv("TLS_KEY", "key.pem")
	t.Setenv("MAX_FILE_SIZE", "1024")
	t.Setenv("MAX_CONCURRENT_STREAMS", "")

	cfg, errs := ConfigFromEnv()

	require.Empty(t, errs)
	assert.Equal(t, "cert.pem", cfg.TLSCert)
	assert.Equal(t, int64(1024), cfg.MaxFileSize)
	assert.Equal(t, uint32(100), cfg.MaxConcurrentStreams)
}

func TestConfigFromEnvInvalidOptional(t *testing.T) {
	setEnv(t, "5051", "123456", "/data", validKey)
	t.Setenv("TLS_CERT", "cert.pem")
	t.Setenv("TLS_KEY", "")
	t.Setenv("MAX_FILE_SIZE", "-1")
	t.Setenv("MAX_CONCURRENT_STREAMS", "0")

	_, errs := ConfigFromEnv()

	assert.Len(t, errs, 3)
}

func writeKeysFile(t *testing.T, mode os.FileMode) string {
	path := t.TempDir() + "/keys.env"

	require.NoError(t, os.WriteFile(path, []byte("ROOT_API_KEY=123456\nENCRYPTION_KEY="+validKey+"\n"), mode))

	return path
}

func TestLoadKeysFile(t *testing.T) {
	setEnv(t, "5051", "", "/data", "")
	t.Setenv("KEYS_FILE", writeKeysFile(t, 0600))

	require.NoError(t, loadKeysFile())

	cfg, errs := ConfigFromEnv()

	require.Empty(t, errs)
	assert.Equal(t, "123456", cfg.APIKey)
	assert.Len(t, cfg.EncryptionKey, 32)
}

func TestLoadKeysFileReadableByOthers(t *testing.T) {
	setEnv(t, "5051", "", "/data", "")
	t.Setenv("KEYS_FILE", writeKeysFile(t, 0644))

	assert.ErrorContains(t, loadKeysFile(), "readable by other users")
}

func TestLoadKeysFileKeyInBoth(t *testing.T) {
	setEnv(t, "5051", "other", "/data", "")
	t.Setenv("KEYS_FILE", writeKeysFile(t, 0600))

	assert.ErrorContains(t, loadKeysFile(), "ROOT_API_KEY is set in both")
}

func TestLoadKeysFileMissing(t *testing.T) {
	setEnv(t, "5051", "", "/data", "")
	t.Setenv("KEYS_FILE", t.TempDir()+"/missing")

	assert.Error(t, loadKeysFile())
}
