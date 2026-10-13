package read_file

import (
	"depozit-server/src/resources"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestReadFile(t *testing.T) {
	root := SetupRoot(t)

	require.NoError(t, root.WriteFile("default/test-read.txt", []byte("test"), 0600))

	content, err := resources.ReadFile(root, container, "test-read.txt")

	require.NoError(t, err)
	assert.Equal(t, "test", string(content))
}

func TestReadFileNotFound(t *testing.T) {
	root := SetupRoot(t)

	content, err := resources.ReadFile(root, container, "unexisting.txt")

	assert.Nil(t, content)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestReadFileInvalidPath(t *testing.T) {
	root := SetupRoot(t)

	for _, path := range []string{"../x", "/etc/passwd", ""} {
		_, err := resources.ReadFile(root, container, path)
		assert.Equal(t, codes.InvalidArgument, status.Code(err), path)
	}
}

func TestReadFileFail(t *testing.T) {
	root := SetupRoot(t)

	// Reading a directory fails with an error other than not found
	require.NoError(t, root.Mkdir("default/dir", 0700))

	content, err := resources.ReadFile(root, container, "dir")

	assert.Nil(t, content)
	assert.Equal(t, codes.Internal, status.Code(err))
}
