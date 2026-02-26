package read_file

import (
	"depozit-server/src/resources"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadFileFail(t *testing.T) {
	tDir := SetupTeardownModule(t)

	fname := "test.txt"

	os.WriteFile(filepath.Join(tDir, fname), []byte("test"), 0644)

	fop := FileOperationsFail{}

	content, err := resources.ReadFile(&fop, tDir, fname)

	assert.Nil(t, content)
	assert.Error(t, err)
}

func TestReadUnexistingFileFail(t *testing.T) {
	tDir := SetupTeardownModule(t)

	fname := "unexisting.txt"

	content, err := resources.ReadFile(&resources.FileOperations{}, tDir, fname)

	assert.Nil(t, content)
	assert.Error(t, err)
}
