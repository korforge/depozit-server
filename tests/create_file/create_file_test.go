package create_file

import (
	"depozit-server/src/resources"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateFile(t *testing.T) {
	root := SetupRoot(t)

	_, err := resources.CreateFile(root, container, "test.txt", []byte("test 001"), 0600, 0700, false)
	require.NoError(t, err)

	content, err := root.ReadFile("default/test.txt")
	require.NoError(t, err)
	assert.Equal(t, []byte("test 001"), content)
}

func TestCreateFileNested(t *testing.T) {
	root := SetupRoot(t)

	_, err := resources.CreateFile(root, container, "a/b/c.txt", []byte("test"), 0600, 0700, false)
	require.NoError(t, err)

	_, err = root.Stat("default/a/b/c.txt")
	assert.NoError(t, err)
}

func TestCreateFileAlreadyExists(t *testing.T) {
	root := SetupRoot(t)

	// Paths must resolve against the root, not the process working directory
	t.Chdir(t.TempDir())

	_, err := resources.CreateFile(root, container, "test.txt", []byte("first"), 0600, 0700, false)
	require.NoError(t, err)

	_, err = resources.CreateFile(root, container, "test.txt", []byte("second"), 0600, 0700, false)
	assert.Equal(t, codes.AlreadyExists, status.Code(err))

	content, err := root.ReadFile("default/test.txt")
	require.NoError(t, err)
	assert.Equal(t, []byte("first"), content)
}

func TestCreateFileOverwrite(t *testing.T) {
	root := SetupRoot(t)

	_, err := resources.CreateFile(root, container, "test.txt", []byte("first"), 0600, 0700, false)
	require.NoError(t, err)

	_, err = resources.CreateFile(root, container, "test.txt", []byte("second"), 0600, 0700, true)
	require.NoError(t, err)

	content, err := root.ReadFile("default/test.txt")
	require.NoError(t, err)
	assert.Equal(t, []byte("second"), content)
}

func TestCreateFileInvalidPath(t *testing.T) {
	root := SetupRoot(t)

	for _, path := range []string{"../x", "/etc/passwd", "", "a/../../x"} {
		_, err := resources.CreateFile(root, container, path, []byte("test"), 0600, 0700, true)
		assert.Equal(t, codes.InvalidArgument, status.Code(err), path)
	}
}

func TestCreateFileContainerNotFound(t *testing.T) {
	root := SetupRoot(t)

	_, err := resources.CreateFile(root, "missing", "test.txt", []byte("test"), 0600, 0700, false)
	assert.Equal(t, codes.NotFound, status.Code(err))

	_, err = root.Stat("missing")
	assert.Error(t, err)
}

func TestCreateFileMkdirAllFail(t *testing.T) {
	root := SetupRoot(t)

	// A file where a parent directory should be makes MkdirAll fail
	require.NoError(t, root.WriteFile("default/a", []byte("file"), 0600))

	_, err := resources.CreateFile(root, container, "a/test.txt", []byte("test"), 0600, 0700, true)
	assert.Equal(t, codes.Internal, status.Code(err))
}

func TestCreateFileWriteFileFail(t *testing.T) {
	root := SetupRoot(t)

	// A directory in place of the file makes WriteFile fail
	require.NoError(t, root.Mkdir("default/test.txt", 0700))

	_, err := resources.CreateFile(root, container, "test.txt", []byte("test"), 0600, 0700, true)
	assert.Equal(t, codes.Internal, status.Code(err))
}
