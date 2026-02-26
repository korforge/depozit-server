package delete_file

import (
	"depozit-server/src/resources"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleteFile(t *testing.T) {
	tDir := SetupTeardownModule(t)

	fname := "test-delete.txt"

	os.WriteFile(filepath.Join(tDir, fname), []byte("test"), 0644)

	resources.DeleteFile(tDir, fname)

	assert.Equal(t, resources.FileExists(filepath.Join(tDir, fname)), false)
}
