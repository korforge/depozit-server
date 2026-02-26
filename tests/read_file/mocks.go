package read_file

import (
	"errors"
	"os"
)

type FileOperationsFail struct{}

func (m *FileOperationsFail) MkdirAll(path string, permission os.FileMode) error {
	return nil
}

func (m *FileOperationsFail) WriteFile(path string, data []byte, permission os.FileMode) error {
	return nil
}

func (m *FileOperationsFail) ReadFile(path string) ([]byte, error) {
	return nil, errors.New("Read error")
}
