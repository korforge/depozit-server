package resources

import (
	"depozit-server/src/config"
	"depozit-server/src/security"
	"io"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Rename(containerDir *os.Root, cipher *security.Cipher, rawPath string, newName string) error {
	path, err := cleanFilePath(rawPath)

	if err != nil {
		return err
	}

	if newName == "" || strings.ContainsRune(newName, filepath.Separator) {
		return status.Error(codes.InvalidArgument, "new name must be a file name, without directories")
	}

	newPath, err := cleanFilePath(filepath.Join(filepath.Dir(path), newName))

	if err != nil {
		return err
	}

	if newPath == path {
		return nil
	}

	// Fail before encrypting the whole file again, the commit checks again without a race
	if _, err := containerDir.Lstat(newPath); err == nil {
		return status.Errorf(codes.AlreadyExists, "%s already exists", newPath)
	}

	src, err := openFile(containerDir, path)

	if err != nil {
		return err
	}

	defer src.Close()

	tmpFile := NewTempFile(containerDir, newPath, false, config.DefaultFilePermission, config.DefaultDirectoryPermission)

	f, err := tmpFile.create()

	if err != nil {
		return err
	}

	// Each chunk is bound to its path, so the file is encrypted again under the new path
	r := newChunkReader(src, cipher, path)
	w := newChunkWriter(f, cipher, newPath)

	for {
		var data []byte

		if data, err = r.Next(); err != nil {
			if err == io.EOF {
				err = w.Close()
			}

			break
		}

		if _, err = w.Write(data); err != nil {
			break
		}
	}

	if err != nil {
		return tmpFile.deleteAfterError(codes.Internal, err, "failed to rename file")
	}

	if err := tmpFile.commit(); err != nil {
		return err
	}

	if err := containerDir.Remove(path); err != nil {
		config.Logger.Error("failed to remove renamed file", "path", path, "error", err)
		return status.Error(codes.Internal, "file copied to the new name but the old file was not removed")
	}

	return nil
}
