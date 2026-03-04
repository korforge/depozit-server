package create_container

import (
	"depozit-server/src/fileio"
	"depozit-server/src/resources"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateContainer(t *testing.T) {
	tDir := SetupTeardownModule(t)

	dir := fileio.Directories{}

	fname := "test"

	path := filepath.Join(tDir, fname)

	resources.CreateContainer(dir, path, 0644)

	assert.True(t, resources.FileExists(path))
}
