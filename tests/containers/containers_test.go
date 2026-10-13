package containers

import (
	"depozit-server/src/resources"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var invalidNames = []string{"", ".", "..", "a/b", "../x", "/abs"}

func TestValidateContainer(t *testing.T) {
	assert.NoError(t, resources.ValidateContainer("photos"))

	for _, name := range invalidNames {
		assert.Equal(t, codes.InvalidArgument, status.Code(resources.ValidateContainer(name)), name)
	}
}

func TestCreateContainer(t *testing.T) {
	root := SetupRoot(t)

	require.NoError(t, resources.CreateContainer(root, "test", 0700))

	info, err := root.Stat("test")
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestCreateContainerAlreadyExists(t *testing.T) {
	root := SetupRoot(t)

	err := resources.CreateContainer(root, container, 0700)

	assert.Equal(t, codes.AlreadyExists, status.Code(err))
}

func TestCreateContainerInvalidName(t *testing.T) {
	root := SetupRoot(t)

	for _, name := range invalidNames {
		err := resources.CreateContainer(root, name, 0700)
		assert.Equal(t, codes.InvalidArgument, status.Code(err), name)
	}
}

func TestDeleteContainer(t *testing.T) {
	root := SetupRoot(t)

	require.NoError(t, root.MkdirAll("test/sub", 0700))
	require.NoError(t, root.WriteFile("test/sub/a.txt", []byte("a"), 0600))

	require.NoError(t, resources.DeleteContainer(root, "test"))

	_, err := root.Stat("test")
	assert.Error(t, err)
}

func TestDeleteContainerNotFound(t *testing.T) {
	root := SetupRoot(t)

	err := resources.DeleteContainer(root, "missing")

	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestDeleteContainerInvalidName(t *testing.T) {
	root := SetupRoot(t)

	err := resources.DeleteContainer(root, ".")

	assert.Equal(t, codes.InvalidArgument, status.Code(err))

	_, err = root.Stat(container)
	assert.NoError(t, err)
}

func TestRenameContainer(t *testing.T) {
	root := SetupRoot(t)

	require.NoError(t, root.Mkdir("test", 0700))

	require.NoError(t, resources.RenameContainer(root, "test", "new-test"))

	_, err := root.Stat("test")
	assert.Error(t, err)

	_, err = root.Stat("new-test")
	assert.NoError(t, err)
}

func TestRenameContainerTargetExists(t *testing.T) {
	root := SetupRoot(t)

	require.NoError(t, root.Mkdir("test", 0700))

	// The target is empty, so a plain rename would replace it
	err := resources.RenameContainer(root, "test", container)

	assert.Equal(t, codes.AlreadyExists, status.Code(err))

	_, err = root.Stat("test")
	assert.NoError(t, err)
}

func TestRenameContainerNotFound(t *testing.T) {
	root := SetupRoot(t)

	err := resources.RenameContainer(root, "missing", "new")

	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestRenameContainerInvalidName(t *testing.T) {
	root := SetupRoot(t)

	err := resources.RenameContainer(root, container, "a/b")

	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}
