package create_container

import (
	"depozit-server/src/fileio"
	"depozit-server/src/resources"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenameContainer(t *testing.T) {
	tDir := SetupTeardownModule(t)

	oldFname := "test"
	newFname := "new-test"

	oldPath := filepath.Join(tDir, oldFname)
	newPath := filepath.Join(tDir, newFname)

	os.MkdirAll(oldPath, 0644)

	dir := fileio.Directories{}

	resources.RenameContainer(dir, oldPath, newPath)

	assert.False(t, resources.FileExists(oldPath))
	assert.True(t, resources.FileExists(newPath))
}
