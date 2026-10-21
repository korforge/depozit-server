package resources

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"depozit-server/src/config"
	pb "depozit-server/src/protobufs"
)

// Returns the cleaned path, the only path Upload may use
func UploadInitialValidations(containerDir *os.Root, chunk *pb.UploadBinaryRequest, overwrite bool, err error) (string, error) {
	if err == io.EOF {
		return "", status.Error(codes.InvalidArgument, "empty stream")
	}

	if err != nil {
		return "", err
	}

	if chunk.GetRemotePath() == "" {
		return "", status.Error(codes.InvalidArgument, "remotePath required on first message")
	}

	path, err := cleanFilePath(chunk.GetRemotePath())

	if err != nil {
		return "", err
	}

	if !overwrite {
		if _, err := containerDir.Lstat(path); err == nil {
			return "", status.Errorf(codes.AlreadyExists, "%s already exists", path)
		}
	}

	return path, nil
}

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
