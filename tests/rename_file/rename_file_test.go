package rename_file

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

func TestRenameFile(t *testing.T) {
	root, store := SetupStore(t)

	data := make([]byte, 2*config.ChunkSize+3)
	require.NoError(t, Upload(store, "notes/hello.txt", data, false))

	require.NoError(t, store.Rename(Container, "notes/hello.txt", "greet.txt"))

	content, err := Read(store, "notes/greet.txt")
	require.NoError(t, err)
	assert.Equal(t, data, content)

	_, err = root.Stat("default/notes/hello.txt")
	assert.Error(t, err)
}

func TestRenameFileSameName(t *testing.T) {
	_, store := SetupStore(t)

	require.NoError(t, Upload(store, "a.txt", []byte("test"), false))
	require.NoError(t, store.Rename(Container, "a.txt", "a.txt"))

	content, err := Read(store, "a.txt")
	require.NoError(t, err)
	assert.Equal(t, "test", string(content))
}

func TestRenameFileTargetExists(t *testing.T) {
	_, store := SetupStore(t)

	require.NoError(t, Upload(store, "a.txt", []byte("a"), false))
	require.NoError(t, Upload(store, "b.txt", []byte("b"), false))

	assert.Equal(t, codes.AlreadyExists, status.Code(store.Rename(Container, "a.txt", "b.txt")))

	content, err := Read(store, "b.txt")
	require.NoError(t, err)
	assert.Equal(t, "b", string(content))
}

func TestRenameFileNotFound(t *testing.T) {
	_, store := SetupStore(t)

	assert.Equal(t, codes.NotFound, status.Code(store.Rename(Container, "missing.txt", "b.txt")))
}

func TestRenameFileInvalidName(t *testing.T) {
	_, store := SetupStore(t)

	require.NoError(t, Upload(store, "a.txt", []byte("a"), false))

	for _, name := range []string{"", "x/y.txt", "..", ".", ".tmp"} {
		assert.Equal(t, codes.InvalidArgument, status.Code(store.Rename(Container, "a.txt", name)), name)
	}
}
