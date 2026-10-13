package resources

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"depozit-server/src/config"
)

// Cleans a file path inside a container. "." is rejected: it is the container itself
func cleanFilePath(rawFname string) (string, error) {
	fname := filepath.Clean(rawFname)

	if !filepath.IsLocal(fname) || fname == "." {
		return "", status.Error(codes.InvalidArgument, "invalid remote path")
	}

	return fname, nil
}

func CreateFile(rootDir *os.Root, container string, rawFname string, data []byte, permission os.FileMode, directoryPermission os.FileMode, overwrite bool) (*emptypb.Empty, error) {
	fname, err := cleanFilePath(rawFname)

	if err != nil {
		return nil, err
	}

	if err := ContainerExists(rootDir, container); err != nil {
		return nil, err
	}

	path := filepath.Join(container, fname)
	exists := FileExists(rootDir, path)

	if !exists || overwrite {
		err := rootDir.MkdirAll(filepath.Dir(path), directoryPermission)

		if err != nil {
			config.Logger.Error("failed to create directory", "error", err)

			return nil, status.Error(codes.Internal, "failed to create directories")
		}

		err = rootDir.WriteFile(path, data, permission)

		if err != nil {
			config.Logger.Error("failed to write file", "error", err)
			return nil, status.Error(codes.Internal, "failed to write file")
		}

		return &emptypb.Empty{}, nil
	}

	return nil, status.Error(codes.AlreadyExists, "file already exists")
}

func ReadFile(rootDir *os.Root, container string, rawFname string) ([]byte, error) {
	fname, err := cleanFilePath(rawFname)

	if err != nil {
		return nil, err
	}

	path := filepath.Join(container, fname)

	data, err := rootDir.ReadFile(path)

	if errors.Is(err, fs.ErrNotExist) {
		return nil, status.Error(codes.NotFound, "file not found")
	}

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to read file")
	}

	return data, nil
}

func DeleteFile(rootDir *os.Root, container string, rawFname string) (*emptypb.Empty, error) {
	fname, err := cleanFilePath(rawFname)

	if err != nil {
		return nil, err
	}

	path := filepath.Join(container, fname)

	if FileExists(rootDir, path) {
		err := rootDir.Remove(path)

		if err != nil {
			config.Logger.Error("failed to delete file", "error", err)
			return &emptypb.Empty{}, status.Error(codes.Internal, "failed to delete file")
		}
	}

	return &emptypb.Empty{}, nil
}

func ListFiles(rootDir *os.Root, container string, prefix string) ([]string, error) {
	var result []string

	err := fs.WalkDir(rootDir.FS(), container, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// WalkDir paths start with the container, the prefix is relative to it
		resultPath := strings.TrimPrefix(path, container+"/")

		if strings.HasPrefix(resultPath, prefix) {
			result = append(result, resultPath)
		}

		return nil
	})

	if errors.Is(err, fs.ErrNotExist) {
		return nil, status.Error(codes.NotFound, "container not found")
	}

	if err != nil {
		config.Logger.Error("failed to list files", "error", err)
		return nil, status.Error(codes.Internal, "failed to list files")
	}

	return result, nil
}
