package delete_file

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

func TestDeleteFile(t *testing.T) {
	root, store := SetupStore(t)

	require.NoError(t, Upload(store, "test-delete.txt", []byte("test"), false))

	require.NoError(t, store.Delete(Container, "test-delete.txt"))

	_, err := root.Stat("default/test-delete.txt")
	assert.Error(t, err)
}

func TestDeleteMissingFile(t *testing.T) {
	_, store := SetupStore(t)

	assert.NoError(t, store.Delete(Container, "missing.txt"))
}

func TestDeleteFileContainerNotFound(t *testing.T) {
	_, store := SetupStore(t)

	assert.Equal(t, codes.NotFound, status.Code(store.Delete("missing", "test.txt")))
}

func TestDeleteFileInvalidPath(t *testing.T) {
	_, store := SetupStore(t)

	for _, path := range []string{"../x", "/etc/passwd", "", "."} {
		err := store.Delete(Container, path)
		assert.Equal(t, codes.InvalidArgument, status.Code(err), path)
	}
}
