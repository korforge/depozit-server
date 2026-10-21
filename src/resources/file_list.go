package resources

import (
	"depozit-server/src/config"
	"errors"
	"io/fs"
	"os"
	"slices"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func comparePaths(a string, b string) int {
	return slices.Compare(strings.Split(a, "/"), strings.Split(b, "/"))
}

func List(containerDir *os.Root, prefix string, pageToken string, pageSize int) ([]string, string, error) {
	if pageSize <= 0 {
		pageSize = config.DefaultPageSize
	}

	pageSize = min(pageSize, config.MaxPageSize)

	var result []string
	nextToken := ""
	errFull := errors.New("page full")

	err := fs.WalkDir(containerDir.FS(), ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if p == "." {
			return nil
		}

		if d.IsDir() {
			dir := p + "/"

			// Skip uploads in progress, directories that cannot hold the prefix, and directories already listed
			if p == TmpDir ||
				!(strings.HasPrefix(dir, prefix) || strings.HasPrefix(prefix, dir)) ||
				(pageToken != "" && comparePaths(p, pageToken) < 0 && !strings.HasPrefix(pageToken, dir)) {
				return fs.SkipDir
			}

			return nil
		}

		if !strings.HasPrefix(p, prefix) || (pageToken != "" && comparePaths(p, pageToken) <= 0) {
			return nil
		}

		if len(result) == pageSize {
			nextToken = result[len(result)-1]
			return errFull
		}

		result = append(result, p)

		return nil
	})

	if err != nil && err != errFull {
		config.Logger.Error("failed to list files", "error", err)
		return nil, "", status.Error(codes.Internal, "failed to list files")
	}

	return result, nextToken, nil
}
