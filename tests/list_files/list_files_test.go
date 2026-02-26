package list_files

import (
	"depozit-server/src/resources"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListFiles(t *testing.T) {
	tDir := SetupTeardownModule(t)

	expected_fname := "expected/test001.txt"
	fname := "test002.txt"

	os.Mkdir(filepath.Join(tDir, "expected"), 0755)
	os.WriteFile(filepath.Join(tDir, expected_fname), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tDir, fname), []byte("test"), 0644)

	expected := []string{filepath.Join(tDir, expected_fname)}

	files, _ := resources.ListFiles(tDir, "test001")

	assert.Equal(t, files, expected)
}
