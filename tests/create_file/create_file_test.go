package create_file

import (
	"depozit-server/src/resources"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateFile(t *testing.T) {
	tDir := SetupTeardownModule(t)

	fname := "test.txt"

	resources.CreateFile(&resources.FileOperations{}, tDir, fname, []byte("test"), 0644, true)

	path := filepath.Join(tDir, fname)

	fileWasCreated := resources.FileExists(path)

	assert.True(t, fileWasCreated)
}

func TestCreateFileAndCheckContent(t *testing.T) {
	tDir := SetupTeardownModule(t)

	fname := "test.txt"

	expected_content := []byte("test 001")

	resources.CreateFile(&resources.FileOperations{}, tDir, fname, expected_content, 0644, true)

	path := filepath.Join(tDir, fname)

	content, err := os.ReadFile(path)

	assert.Equal(t, nil, err)
	assert.Equal(t, content, expected_content)

}
