package resources

import (
	"depozit-server/src/config"
	"depozit-server/src/security"
	"errors"
	"io"
	"os"

	pb "depozit-server/src/protobufs"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Read(containerDir *os.Root, cipher *security.Cipher, remotePath string, stream pb.FilesService_ReadStreamServer) error {
	path, err := cleanFilePath(remotePath)

	if err != nil {
		return err
	}

	f, err := openFile(containerDir, path)

	if err != nil {
		return err
	}

	defer f.Close()

	// One response reused for every chunk, Send serializes the message before returning
	resp := &pb.ReadResponse{}
	r := newChunkReader(f, cipher, path)

	for {
		data, err := r.Next()

		if err == io.EOF {
			return nil
		}

		if errors.Is(err, errCorrupt) {
			config.Logger.Error("failed to read file", "path", path, "error", err)
			return status.Error(codes.DataLoss, "file is corrupted")
		}

		if err != nil {
			config.Logger.Error("failed to read file", "path", path, "error", err)
			return status.Error(codes.Internal, "failed to read file")
		}

		// The empty last chunk carries no data
		if len(data) == 0 {
			continue
		}

		resp.Data = data

		// Stop reading when the client is gone or the stream failed
		if err := stream.Send(resp); err != nil {
			return err
		}
	}
}
