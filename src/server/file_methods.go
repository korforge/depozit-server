package server

import (
	"context"
	"depozit-server/src/config"
	pb "depozit-server/src/protobufs"
	"depozit-server/src/resources"
	"depozit-server/src/security"
	"os"

	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	pb.UnimplementedFilesServiceServer
	cfg     *config.Config
	rootDir *os.Root
	cipher  *security.Cipher
}

func (s *Server) UploadStream(stream pb.FilesService_UploadStreamServer) error {
	config.Logger.Info("upload file")

	containerDir, err := resources.GetContainerDir(s.rootDir, stream.Context(), DefaultContainer)

	if err != nil {
		return err
	}

	defer containerDir.Close()

	err = resources.Upload(containerDir, s.cipher, s.cfg.MaxFileSize, stream)

	if err != nil {
		return err
	}

	return stream.SendAndClose(&emptypb.Empty{})
}

func (s *Server) ReadStream(req *pb.ReadDeleteRequest, stream pb.FilesService_ReadStreamServer) error {
	containerDir, err := resources.GetContainerDir(s.rootDir, stream.Context(), DefaultContainer)

	if err != nil {
		return err
	}

	defer containerDir.Close()

	return resources.Read(containerDir, s.cipher, req.RemotePath, stream)
}

func (s *Server) Delete(ctx context.Context, req *pb.ReadDeleteRequest) (*emptypb.Empty, error) {
	config.Logger.Info("delete file")

	containerDir, err := resources.GetContainerDir(s.rootDir, ctx, DefaultContainer)

	if err != nil {
		return nil, err
	}

	defer containerDir.Close()

	if err := resources.Delete(containerDir, req.RemotePath); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) Rename(ctx context.Context, req *pb.RenameRequest) (*emptypb.Empty, error) {
	config.Logger.Info("rename file", "path", req.RemotePath, "new_name", req.NewName)

	containerDir, err := resources.GetContainerDir(s.rootDir, ctx, DefaultContainer)

	if err != nil {
		return nil, err
	}

	defer containerDir.Close()

	if err := resources.Rename(containerDir, s.cipher, req.RemotePath, req.NewName); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) List(ctx context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
	containerDir, err := resources.GetContainerDir(s.rootDir, ctx, DefaultContainer)

	if err != nil {
		return nil, err
	}

	defer containerDir.Close()

	paths, nextPageToken, err := resources.List(containerDir, req.Prefix, req.PageToken, int(req.PageSize))

	if err != nil {
		return nil, err
	}

	return &pb.ListResponse{RemotePaths: paths, NextPageToken: nextPageToken}, nil
}
