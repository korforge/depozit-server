package resources

import (
	"depozit-server/src/fileio"
	"os"
)

func CreateContainer(d fileio.Directories, container string, permission os.FileMode) error {
	return d.MkdirAll(container, permission)
}

func DeleteContainer(d fileio.Directories, container string) error {
	return d.RemoveAll(container)
}

func RenameContainer(d fileio.Directories, oldPath string, newPath string) error {
	return os.Rename(oldPath, newPath)
}
