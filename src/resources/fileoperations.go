package resources

import (
	"os"
)

type IFileOperations interface {
	MkdirAll(path string, permission os.FileMode) error
	WriteFile(path string, data []byte, permission os.FileMode) error
	ReadFile(path string) ([]byte, error)
}

type FileOperations struct{}

func (fs *FileOperations) MkdirAll(path string, permission os.FileMode) error {
	return os.MkdirAll(path, permission)
}

func (fs *FileOperations) WriteFile(path string, data []byte, permission os.FileMode) error {
	return os.WriteFile(path, data, permission)
}

func (fs *FileOperations) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}
