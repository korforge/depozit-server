package delete_file

import (
	"depozit-server/src/resources"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestDeleteFile(t *testing.T) {
	root := SetupRoot(t)

	// Paths must resolve against the root, not the process working directory
	t.Chdir(t.TempDir())

	require.NoError(t, root.WriteFile("default/test-delete.txt", []byte("test"), 0600))

	_, err := resources.DeleteFile(root, container, "test-delete.txt")
	require.NoError(t, err)

	_, err = root.Stat("default/test-delete.txt")
	assert.Error(t, err)
}

func TestDeleteMissingFile(t *testing.T) {
	root := SetupRoot(t)

	_, err := resources.DeleteFile(root, container, "missing.txt")

	assert.NoError(t, err)
}

func TestDeleteFileInvalidPath(t *testing.T) {
	root := SetupRoot(t)

	for _, path := range []string{"../x", "/etc/passwd", ""} {
		_, err := resources.DeleteFile(root, container, path)
		assert.Equal(t, codes.InvalidArgument, status.Code(err), path)
	}
}
