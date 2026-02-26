package resources

import "os"

func CreateContainer(container string, permission os.FileMode) error {
	return os.MkdirAll(container, permission)
}

func DeleteContainer(container string) error {
	return os.RemoveAll(container)
}

func RenameContainer(oldName string, newName string) error {
	return os.Rename(oldName, newName)
}
