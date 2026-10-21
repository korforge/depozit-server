package resources

import (
	"crypto/sha256"
	"depozit-server/src/config"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Uploads in progress, hidden from clients
const TmpDir = ".tmp"

type TmpFile struct {
	path                string
	tmpPath             string
	overwrite           bool
	filePermission      os.FileMode
	directoryPermission os.FileMode
	containerDir        *os.Root
	file                *os.File
}

func NewTempFile(containerDir *os.Root, path string, overwrite bool, filePermission os.FileMode, directoryPermission os.FileMode) *TmpFile {
	// Same path gives the same temp name, so O_EXCL lets only one upload per path run at a time
	sum := sha256.Sum256([]byte(path))

	return &TmpFile{
		path:                path,
		tmpPath:             filepath.Join(TmpDir, hex.EncodeToString(sum[:])),
		overwrite:           overwrite,
		filePermission:      filePermission,
		directoryPermission: directoryPermission,
		containerDir:        containerDir,
	}
}

func (s *TmpFile) create() (*os.File, error) {
	if err := s.containerDir.MkdirAll(TmpDir, s.directoryPermission); err != nil {
		config.Logger.Error("failed to create temp dir", "error", err)
		return nil, status.Error(codes.Internal, "failed to create file")
	}

	f, err := s.containerDir.OpenFile(s.tmpPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, s.filePermission)

	if errors.Is(err, fs.ErrExist) {
		return nil, status.Errorf(codes.AlreadyExists, "%s is being uploaded", s.path)
	}

	if err != nil {
		config.Logger.Error("failed to create temp file", "error", err)
		return nil, status.Error(codes.Internal, "failed to create file")
	}

	s.file = f

	return f, nil
}

// Removes the temp file even when closing it fails, a close error does not matter for a discarded file
func (s *TmpFile) delete() error {
	s.file.Close()

	return s.containerDir.Remove(s.tmpPath)
}

func (s *TmpFile) deleteAfterError(code codes.Code, srcErr error, errorMsg string) error {
	config.Logger.Error(errorMsg, "error", srcErr)

	if err := s.delete(); err != nil {
		config.Logger.Error("failed to delete temp file", "path", s.tmpPath, "error", err)
	}

	return status.Error(code, errorMsg)
}

func (s *TmpFile) commit() error {
	if err := s.file.Sync(); err != nil {
		return s.deleteAfterError(codes.Internal, err, "failed to write file")
	}

	if err := s.file.Close(); err != nil {
		return s.deleteAfterError(codes.Internal, err, "failed to write file")
	}

	defer s.containerDir.Remove(s.tmpPath)

	if dir := filepath.Dir(s.path); dir != "." {
		if err := s.containerDir.MkdirAll(dir, s.directoryPermission); err != nil {
			config.Logger.Error("failed to create directories", "error", err)
			return status.Error(codes.Internal, "failed to create directories")
		}
	}

	var err error

	if s.overwrite {
		err = s.containerDir.Rename(s.tmpPath, s.path)
	} else {
		err = s.containerDir.Link(s.tmpPath, s.path)
	}

	if errors.Is(err, fs.ErrExist) {
		return status.Errorf(codes.AlreadyExists, "%s already exists", s.path)
	}

	if err != nil {
		config.Logger.Error("failed to commit file", "path", s.path, "error", err)
		return status.Error(codes.Internal, "failed to create file")
	}

	return nil
}
