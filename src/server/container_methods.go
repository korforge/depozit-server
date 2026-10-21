package server

import (
	"context"
	"depozit-server/src/config"
	"depozit-server/src/resources"

	pb "depozit-server/src/protobufs"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) CreateContainer(ctx context.Context, req *pb.ContainerRequest) (*emptypb.Empty, error) {
	config.Logger.Info("create container", "name", req.Name)

	if err := resources.CreateContainer(s.rootDir, req.Name, config.DefaultDirectoryPermission); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) DeleteContainer(ctx context.Context, req *pb.ContainerRequest) (*emptypb.Empty, error) {
	config.Logger.Info("delete container", "name", req.Name)

	if req.Name == DefaultContainer {
		return nil, status.Error(codes.PermissionDenied, "default container cannot be deleted")
	}

	if err := resources.DeleteContainer(s.rootDir, req.Name); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) RenameContainer(ctx context.Context, req *pb.RenameRequest) (*emptypb.Empty, error) {
	config.Logger.Info("rename container", "old_name", req.RemotePath, "new_name", req.NewName)

	if req.RemotePath == DefaultContainer {
		return nil, status.Error(codes.PermissionDenied, "default container cannot be renamed")
	}

	if err := resources.RenameContainer(s.rootDir, req.RemotePath, req.NewName); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
