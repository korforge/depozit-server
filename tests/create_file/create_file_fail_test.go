package create_file

import (
	"depozit-server/src/resources"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateFileAlreadyExists(t *testing.T) {
	tDir := SetupTeardownModule(t)

	fname := "test.txt"

	expected_content := []byte("test 001")

	resources.CreateFile(&resources.FileOperations{}, tDir, fname, expected_content, 0644, false)

	resources.CreateFile(&resources.FileOperations{}, tDir, fname, expected_content, 0644, false)

	path := filepath.Join(tDir, fname)

	assert.Equal(t, resources.FileExists(path), true)
}

func TestCreateFileMkdirAllFail(t *testing.T) {
	tDir := SetupTeardownModule(t)

	fname := "test.txt"

	expected_content := []byte("test 001")

	resources.CreateFile(&FileOperationsFail{}, tDir, fname, expected_content, 0644, true)

	path := filepath.Join(tDir, fname)

	assert.Equal(t, resources.FileExists(path), false)
}

func TestCreateFileWriteFileFail(t *testing.T) {
	tDir := SetupTeardownModule(t)

	fname := "test.txt"

	expected_content := []byte("test 001")

	resources.CreateFile(&FileOperationsWriteFileFail{}, tDir, fname, expected_content, 0644, true)

	path := filepath.Join(tDir, fname)

	assert.Equal(t, resources.FileExists(path), false)
}
