package resources

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"depozit-server/src/config"
)

func CreateFile(fop IFileOperations, container string, fname string, data []byte, permission os.FileMode, overwrite bool) (*emptypb.Empty, error) {
	path := filepath.Join(container, fname)
	exists := FileExists(path)

	if !exists || overwrite {
		err := fop.MkdirAll(filepath.Dir(path), permission)

		if err != nil {
			config.Logger.Error("failed to create directory", "error", err)

			return nil, status.Error(codes.Internal, "failed to create directories")
		}

		err = fop.WriteFile(path, data, permission)

		if err != nil {
			config.Logger.Error("failed to write file", "error", err)
			return nil, status.Error(codes.Internal, "failed to write file")
		}

		return &emptypb.Empty{}, nil
	}

	return nil, status.Error(codes.AlreadyExists, "file already exists")
}

func ReadFile(fop IFileOperations, container string, fname string) ([]byte, error) {
	path := filepath.Join(container, fname)

	data, err := fop.ReadFile(path)
	exists := FileExists(path)

	if !exists {
		return nil, status.Error(codes.NotFound, "file not found")
	}

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to read file")
	}

	return data, nil
}

func DeleteFile(container string, fname string) (*emptypb.Empty, error) {
	path := filepath.Join(container, fname)

	if FileExists(path) {
		os.Remove(path)
	}

	return &emptypb.Empty{}, nil
}

func ListFiles(container string, prefix string) ([]string, error) {
	var result []string

	err := filepath.WalkDir(container, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.Contains(path, prefix) {
			result_path := strings.Replace(path, container+"/", "", 1)
			result = append(result, result_path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
