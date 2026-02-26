package read_file

import (
	"depozit-server/src/resources"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadFile(t *testing.T) {
	tDir := SetupTeardownModule(t)

	fname := "test-read.txt"
	expected_content := "test"

	os.WriteFile(filepath.Join(tDir, fname), []byte(expected_content), 0644)

	content, err := resources.ReadFile(&resources.FileOperations{}, tDir, fname)

	assert.Equal(t, string(content), expected_content)
	assert.NoError(t, err)
}
