package resources

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"depozit-server/src/config"
)

// A container is a single directory directly under the storage root
func ValidateContainer(container string) error {
	if !filepath.IsLocal(container) || container == "." || strings.ContainsRune(container, filepath.Separator) {
		return status.Error(codes.InvalidArgument, "invalid container name")
	}

	return nil
}

func ContainerExists(d *os.Root, container string) error {
	info, err := d.Stat(container)

	if errors.Is(err, fs.ErrNotExist) || (err == nil && !info.IsDir()) {
		return status.Error(codes.NotFound, "container not found")
	}

	if err != nil {
		config.Logger.Error("failed to stat container", "error", err)
		return status.Error(codes.Internal, "failed to check container")
	}

	return nil
}

func CreateContainer(d *os.Root, container string, permission os.FileMode) error {
	if err := ValidateContainer(container); err != nil {
		return err
	}

	err := d.Mkdir(container, permission)

	if errors.Is(err, fs.ErrExist) {
		return status.Error(codes.AlreadyExists, "container already exists")
	}

	if err != nil {
		config.Logger.Error("failed to create container", "error", err)
		return status.Error(codes.Internal, "failed to create container")
	}

	return nil
}

func DeleteContainer(d *os.Root, container string) error {
	if err := ValidateContainer(container); err != nil {
		return err
	}

	if err := ContainerExists(d, container); err != nil {
		return err
	}

	err := d.RemoveAll(container)

	if err != nil {
		config.Logger.Error("failed to delete container", "error", err)
		return status.Error(codes.Internal, "failed to delete container")
	}

	return nil
}

func RenameContainer(d *os.Root, oldName string, newName string) error {
	if err := ValidateContainer(oldName); err != nil {
		return err
	}

	if err := ValidateContainer(newName); err != nil {
		return err
	}

	if err := ContainerExists(d, oldName); err != nil {
		return err
	}

	// Rename replaces an empty target directory, so check first
	if _, err := d.Stat(newName); err == nil {
		return status.Error(codes.AlreadyExists, "container already exists")
	}

	err := d.Rename(oldName, newName)

	if err != nil {
		config.Logger.Error("failed to rename container", "error", err)
		return status.Error(codes.Internal, "failed to rename container")
	}

	return nil
}
