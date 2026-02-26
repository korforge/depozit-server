package create_file

import (
	"errors"
	"os"
)

type FileOperationsFail struct{}

func (m *FileOperationsFail) MkdirAll(path string, perm os.FileMode) error {
	return errors.New("mkdir error")
}

func (m *FileOperationsFail) WriteFile(path string, data []byte, perm os.FileMode) error {
	return errors.New("write file error")
}

func (m *FileOperationsFail) ReadFile(path string) ([]byte, error) {
	return nil, errors.New("read file error")
}

type FileOperationsWriteFileFail struct{}

func (m *FileOperationsWriteFileFail) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (m *FileOperationsWriteFileFail) WriteFile(path string, data []byte, perm os.FileMode) error {
	return errors.New("write file error")
}

func (m *FileOperationsWriteFileFail) ReadFile(path string) ([]byte, error) {
	return nil, errors.New("read file error")
}
