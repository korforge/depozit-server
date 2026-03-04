package fileio

import (
	"os"
)

type IDirectories interface {
	MkdirAll(path string, permission os.FileMode) error
	RemoveAll(path string) error
	Rename(path string, newName string) error
}

type Directories struct{}

func (d *Directories) MkdirAll(path string, permission os.FileMode) error {
	return os.MkdirAll(path, permission)
}

func (d *Directories) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (d *Directories) Rename(path string, newName string) error {
	return os.Rename(path, newName)
}
