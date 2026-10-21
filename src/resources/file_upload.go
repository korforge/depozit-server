package resources

import (
	"depozit-server/src/config"
	"depozit-server/src/security"
	"errors"
	"fmt"
	"io"
	"os"

	pb "depozit-server/src/protobufs"

	"google.golang.org/grpc/codes"
)

var errFileTooLarge = errors.New("upload over MAX_FILE_SIZE")

// maxFileSize 0 means no limit
func Upload(containerDir *os.Root, cipher *security.Cipher, maxFileSize int64, stream pb.FilesService_UploadStreamServer) error {
	msg, err := stream.Recv()
	overwrite := msg.GetOverwrite()

	path, err := UploadInitialValidations(containerDir, msg, overwrite, err)

	if err != nil {
		return err
	}

	tmpFile := NewTempFile(containerDir, path, overwrite, config.DefaultFilePermission, config.DefaultDirectoryPermission)

	f, err := tmpFile.create()

	if err != nil {
		return err
	}

	w := newChunkWriter(f, cipher, path)
	var size int64

	for {
		data := msg.GetData()
		size += int64(len(data))

		// Checked before the write, so data over the limit never reaches the disk
		if maxFileSize > 0 && size > maxFileSize {
			return tmpFile.deleteAfterError(codes.ResourceExhausted, errFileTooLarge, fmt.Sprintf("file is larger than %d bytes", maxFileSize))
		}

		if _, err = w.Write(data); err != nil {
			return tmpFile.deleteAfterError(codes.Internal, err, "failed to write file")
		}

		msg, err = stream.Recv()

		if err == io.EOF {
			break
		}

		if err != nil {
			return tmpFile.deleteAfterError(codes.Internal, err, "failed to write file")
		}
	}

	if err := w.Close(); err != nil {
		return tmpFile.deleteAfterError(codes.Internal, err, "failed to write file")
	}

	return tmpFile.commit()
}
