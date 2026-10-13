package list_files

import (
	"depozit-server/src/resources"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestListFiles(t *testing.T) {
	root := SetupRoot(t)

	require.NoError(t, root.MkdirAll("default/expected", 0700))
	require.NoError(t, root.WriteFile("default/expected/test001.txt", []byte("test"), 0600))
	require.NoError(t, root.WriteFile("default/test002.txt", []byte("test"), 0600))

	files, err := resources.ListFiles(root, container, "expected")

	require.NoError(t, err)
	assert.Equal(t, []string{"expected/test001.txt"}, files)
}

func TestListFilesAll(t *testing.T) {
	root := SetupRoot(t)

	require.NoError(t, root.MkdirAll("default/a", 0700))
	require.NoError(t, root.WriteFile("default/a/b.txt", []byte("test"), 0600))
	require.NoError(t, root.WriteFile("default/c.txt", []byte("test"), 0600))

	files, err := resources.ListFiles(root, container, "")

	require.NoError(t, err)
	assert.Equal(t, []string{"a/b.txt", "c.txt"}, files)

	// Listed names must work as read paths
	for _, f := range files {
		_, err := resources.ReadFile(root, container, f)
		assert.NoError(t, err, f)
	}
}

func TestListFilesPrefixDoesNotMatchContainer(t *testing.T) {
	root := SetupRoot(t)

	require.NoError(t, root.WriteFile("default/test.txt", []byte("test"), 0600))

	files, err := resources.ListFiles(root, container, "default")

	require.NoError(t, err)
	assert.Empty(t, files)
}

func TestListFilesOtherContainer(t *testing.T) {
	root := SetupRoot(t)

	require.NoError(t, root.Mkdir("other", 0700))
	require.NoError(t, root.WriteFile("other/x.txt", []byte("test"), 0600))
	require.NoError(t, root.WriteFile("default/y.txt", []byte("test"), 0600))

	files, err := resources.ListFiles(root, container, "")

	require.NoError(t, err)
	assert.Equal(t, []string{"y.txt"}, files)
}

func TestListFilesContainerNotFound(t *testing.T) {
	root := SetupRoot(t)

	_, err := resources.ListFiles(root, "missing", "")

	assert.Equal(t, codes.NotFound, status.Code(err))
}
