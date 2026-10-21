package resources

import (
	"depozit-server/src/config"
	"errors"
	"io/fs"
	"os"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Delete(containerDir *os.Root, rawPath string) error {
	path, err := cleanFilePath(rawPath)

	if err != nil {
		return err
	}

	err = containerDir.Remove(path)

	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		config.Logger.Error("failed to delete file", "path", path, "error", err)
		return status.Error(codes.Internal, "failed to delete file")
	}

	return nil
}
