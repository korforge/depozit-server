package read_file

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"depozit-server/src/config"
	. "depozit-server/tests/testutil"
)

func TestMain(m *testing.M) {
	config.SetupLogger()

	os.Exit(m.Run())
}

func TestReadFile(t *testing.T) {
	_, store := SetupStore(t)

	require.NoError(t, Upload(store, "test-read.txt", []byte("test"), false))

	content, err := Read(store, "test-read.txt")

	require.NoError(t, err)
	assert.Equal(t, "test", string(content))
}

// Different spellings of the same path read the same file
func TestReadFileUncleanPath(t *testing.T) {
	_, store := SetupStore(t)

	require.NoError(t, Upload(store, "a//b/./c.txt", []byte("test"), false))

	for _, path := range []string{"a/b/c.txt", "a//b/c.txt", "a/x/../b/c.txt"} {
		content, err := Read(store, path)
		require.NoError(t, err, path)
		assert.Equal(t, "test", string(content), path)
	}
}

func TestReadFileNotFound(t *testing.T) {
	_, store := SetupStore(t)

	content, err := Read(store, "unexisting.txt")

	assert.Nil(t, content)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestReadFileContainerNotFound(t *testing.T) {
	_, store := SetupStore(t)

	err := store.Read("missing", "test.txt", func([]byte) error { return nil })

	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestReadFileInvalidPath(t *testing.T) {
	_, store := SetupStore(t)

	for _, path := range []string{"../x", "/etc/passwd", "", ".tmp/x"} {
		_, err := Read(store, path)
		assert.Equal(t, codes.InvalidArgument, status.Code(err), path)
	}
}

func TestReadFileFail(t *testing.T) {
	root, store := SetupStore(t)

	// Reading a directory fails with an error other than not found
	require.NoError(t, root.Mkdir("default/dir", 0700))

	content, err := Read(store, "dir")

	assert.Nil(t, content)
	assert.Equal(t, codes.Internal, status.Code(err))
}

func TestReadFileCorrupted(t *testing.T) {
	root, store := SetupStore(t)

	require.NoError(t, Upload(store, "test.txt", []byte("test"), false))

	raw, err := root.ReadFile("default/test.txt")
	require.NoError(t, err)

	// Truncated, and copied under another name
	require.NoError(t, root.WriteFile("default/test.txt", raw[:len(raw)-1], 0600))
	require.NoError(t, root.WriteFile("default/copy.txt", raw, 0600))

	for _, path := range []string{"test.txt", "copy.txt"} {
		_, err = Read(store, path)
		assert.Equal(t, codes.DataLoss, status.Code(err), path)
	}
}
