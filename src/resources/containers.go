package resources

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"depozit-server/src/config"
)

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

// Removes uploads left behind by a crash. Call it before the server accepts requests,
// while no upload can be running
func CleanTempFiles(d *os.Root) error {
	entries, err := fs.ReadDir(d.FS(), ".")

	if err != nil {
		return err
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		if err := d.RemoveAll(filepath.Join(e.Name(), TmpDir)); err != nil {
			return err
		}
	}

	return nil
}

// Removes temp files not written for maxAge, left by uploads that stopped without an error.
// Every write updates the modification time, so a running upload keeps its temp file
func CleanStaleTempFiles(d *os.Root, maxAge time.Duration) error {
	entries, err := fs.ReadDir(d.FS(), ".")

	if err != nil {
		return err
	}

	cutoff := time.Now().Add(-maxAge)

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		tmpDir := filepath.Join(e.Name(), TmpDir)
		files, err := fs.ReadDir(d.FS(), tmpDir)

		if errors.Is(err, fs.ErrNotExist) {
			continue
		}

		if err != nil {
			config.Logger.Error("failed to list temp files", "path", tmpDir, "error", err)
			continue
		}

		for _, f := range files {
			info, err := f.Info()

			// The upload finished or failed since the listing
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}

			if err != nil {
				config.Logger.Error("failed to stat temp file", "path", tmpDir, "error", err)
				continue
			}

			if info.ModTime().After(cutoff) {
				continue
			}

			path := filepath.Join(tmpDir, f.Name())

			if err := d.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
				config.Logger.Error("failed to delete stale temp file", "path", path, "error", err)
				continue
			}

			config.Logger.Info("deleted stale temp file", "path", path)
		}
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
