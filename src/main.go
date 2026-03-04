package main

import (
	"context"
	"log"
	"net"
	"path/filepath"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"depozit-server/src/config"
	"depozit-server/src/fileio"
	pb "depozit-server/src/protobufs"
	"depozit-server/src/resources"
)

type server struct {
	pb.UnimplementedFilesServiceServer
	cfg *config.Config
}

func (s *server) UploadBinary(ctx context.Context, req *pb.UploadBinaryRequest) (*emptypb.Empty, error) {
	config.Logger.Info("Create file", req.Name)

	fop := resources.FileOperations{}

	resp, err := resources.CreateFile(&fop, s.cfg.DefaultContainer, req.Name, req.Data, s.cfg.Permission, req.Overwrite)

	return resp, err
}

func (s *server) Read(ctx context.Context, req *pb.ReadDeleteRequest) (*pb.ReadResponse, error) {
	config.Logger.Info("Read file", req.Name)

	fop := resources.FileOperations{}

	data, err := resources.ReadFile(&fop, s.cfg.DefaultContainer, req.Name)

	return &pb.ReadResponse{Data: data}, err
}

func (s *server) Delete(ctx context.Context, req *pb.ReadDeleteRequest) (*emptypb.Empty, error) {
	config.Logger.Info("Delete file", req.Name)

	resp, err := resources.DeleteFile(s.cfg.DefaultContainer, req.Name)

	return resp, err
}

func (s *server) List(ctx context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
	result, err := resources.ListFiles(s.cfg.DefaultContainer, req.Prefix)

	if err != nil {
		return nil, err
	}

	return &pb.ListResponse{Names: result}, nil
}

func (s *server) CreateContainer(ctx context.Context, req *pb.ContainerRequest) (*emptypb.Empty, error) {
	config.Logger.Info("CreateContainer", req.Name)

	dir := fileio.Directories{}

	err := resources.CreateContainer(dir, req.Name, s.cfg.Permission)

	if err != nil {
		return nil, err
	}

	config.Logger.Info("Success")

	return &emptypb.Empty{}, nil
}

func (s *server) DeleteContainer(ctx context.Context, req *pb.ContainerRequest) (*emptypb.Empty, error) {
	config.Logger.Info("DeleteContainer", req.Name)

	dir := fileio.Directories{}

	err := resources.DeleteContainer(dir, req.Name)

	if err != nil {
		return nil, err
	}

	config.Logger.Info("Success")

	return &emptypb.Empty{}, nil
}

func (s *server) RenameContainer(ctx context.Context, req *pb.RenameRequest) (*emptypb.Empty, error) {
	config.Logger.Info("RenameContainer called", "old_name", req.OldName, "new_name", req.NewName)

	dir := fileio.Directories{}

	err := resources.RenameContainer(dir, req.OldName, req.NewName)

	if err != nil {
		return nil, err
	}

	config.Logger.Info("Success")

	return &emptypb.Empty{}, nil
}

func runServer(cfg *config.Config) {
	lis, err := net.Listen("tcp", ":"+cfg.Port)

	if err != nil {
		log.Fatal(err)
	}

	authInterceptor := AuthInterceptor(cfg)

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(authInterceptor))

	pb.RegisterFilesServiceServer(grpcServer, &server{cfg: cfg})

	resources.CreateContainer(filepath.Join(cfg.StoragePath, cfg.DefaultContainer), cfg.Permission)

	config.Logger.Info("Server running on :" + cfg.Port)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}

func main() {
	config.SetupLogger()
	cfg := CliLoadConfig()
	runServer(cfg)
}
