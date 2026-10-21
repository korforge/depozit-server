package resources

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"depozit-server/src/config"
)

func cleanFilePath(rawFname string) (string, error) {
	fname := filepath.Clean(rawFname)

	// The temp dir is reserved for uploads in progress
	if !filepath.IsLocal(fname) || fname == "." || fname == TmpDir || strings.HasPrefix(fname, TmpDir+"/") {
		return "", status.Error(codes.InvalidArgument, "invalid remote path")
	}

	return fname, nil
}

func openFile(containerDir *os.Root, path string) (*os.File, error) {
	f, err := containerDir.Open(path)

	if errors.Is(err, fs.ErrNotExist) {
		return nil, status.Error(codes.NotFound, "file not found")
	}

	if err != nil {
		config.Logger.Error("failed to open file", "path", path, "error", err)
		return nil, status.Error(codes.Internal, "failed to open file")
	}

	return f, nil
}

func OpenContainer(rootDir *os.Root, container string) (*os.Root, error) {
	containerDir, err := rootDir.OpenRoot(container)

	if errors.Is(err, fs.ErrNotExist) || errors.Is(err, fs.ErrInvalid) {
		return nil, status.Error(codes.NotFound, "container not found")
	}

	if err != nil {
		config.Logger.Error("failed to open container", "container", container, "error", err)
		return nil, status.Error(codes.Internal, "failed to open container")
	}

	return containerDir, nil
}

func ContainerFromContext(ctx context.Context, defaultContainer string) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		return defaultContainer, nil
	}

	values := md.Get("container")

	if len(values) == 0 || values[0] == "" {
		return defaultContainer, nil
	}

	if err := ValidateContainer(values[0]); err != nil {
		return "", err
	}

	return values[0], nil
}

func GetContainerDir(rootDir *os.Root, ctx context.Context, defaultContainer string) (*os.Root, error) {
	container, err := ContainerFromContext(ctx, defaultContainer)

	if err != nil {
		return &os.Root{}, err
	}

	return OpenContainer(rootDir, container)
}
