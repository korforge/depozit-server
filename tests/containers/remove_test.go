package create_container

import (
	"depozit-server/src/fileio"
	"depozit-server/src/resources"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveContainer(t *testing.T) {
	tDir := SetupTeardownModule(t)

	fname := "test"

	path := filepath.Join(tDir, fname)

	os.MkdirAll(path, 0644)

	dir := fileio.Directories{}

	resources.DeleteContainer(dir, path)

	assert.False(t, resources.FileExists(path))
}
