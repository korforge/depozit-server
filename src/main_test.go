package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const validKey = "62cc8e8ad7741788874fb42d493709eb2e66811403f12966e2814c72f9cb1946"

func TestCipherRoundTrip(t *testing.T) {
	c, err := NewCipher(make([]byte, 32))
	require.NoError(t, err)

	enc := c.Encrypt([]byte("hello"), "a.txt")

	data, err := c.Decrypt(enc, "a.txt")
	require.NoError(t, err)
	assert.Equal(t, []byte("hello"), data)
}

func TestCipherWrongAAD(t *testing.T) {
	c, err := NewCipher(make([]byte, 32))
	require.NoError(t, err)

	enc := c.Encrypt([]byte("hello"), "a.txt")

	_, err = c.Decrypt(enc, "b.txt")
	assert.Error(t, err)
}

func TestCipherTampered(t *testing.T) {
	c, err := NewCipher(make([]byte, 32))
	require.NoError(t, err)

	enc := c.Encrypt([]byte("hello"), "a.txt")
	enc[len(enc)-1] ^= 1

	_, err = c.Decrypt(enc, "a.txt")
	assert.Error(t, err)
}

func setEnv(t *testing.T, port string, apiKey string, storagePath string, encryptionKey string) {
	t.Setenv("HOST", "")
	t.Setenv("PORT", port)
	t.Setenv("ROOT_API_KEY", apiKey)
	t.Setenv("STORAGE_PATH", storagePath)
	t.Setenv("ENCRYPTION_KEY", encryptionKey)
}

func TestConfigFromEnv(t *testing.T) {
	setEnv(t, "5051", "123456", "/data", validKey)

	cfg, errs := ConfigFromEnv()

	require.Empty(t, errs)
	assert.Equal(t, "0.0.0.0", cfg.Host)
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

func TestContainerFromContext(t *testing.T) {
	container, err := containerFromContext(context.Background())
	require.NoError(t, err)
	assert.Equal(t, DefaultContainer, container)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("container", "photos"))
	container, err = containerFromContext(ctx)
	require.NoError(t, err)
	assert.Equal(t, "photos", container)

	for _, name := range []string{".", "a/b", "../x"} {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("container", name))
		_, err := containerFromContext(ctx)
		assert.Equal(t, codes.InvalidArgument, status.Code(err), name)
	}
}
