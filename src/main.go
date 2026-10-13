package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"depozit-server/src/config"
	pb "depozit-server/src/protobufs"
	"depozit-server/src/resources"
)

type server struct {
	pb.UnimplementedFilesServiceServer
	cfg     *config.Config
	cipher  *Cipher
	rootDir *os.Root
}

const DefaultContainer = "default"

func containerFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		return DefaultContainer, nil
	}

	values := md.Get("container")

	if len(values) == 0 || values[0] == "" {
		return DefaultContainer, nil
	}

	if err := resources.ValidateContainer(values[0]); err != nil {
		return "", err
	}

	return values[0], nil
}

func (s *server) UploadBinary(ctx context.Context, req *pb.UploadBinaryRequest) (*emptypb.Empty, error) {
	config.Logger.Info("create file", "path", req.RemotePath)

	container, err := containerFromContext(ctx)

	if err != nil {
		return nil, err
	}

	data := s.cipher.Encrypt(req.Data, req.RemotePath)

	resp, err := resources.CreateFile(s.rootDir, container, req.RemotePath, data, s.cfg.Permission, s.cfg.DirectoryPermission, req.Overwrite)

	return resp, err
}

func (s *server) Read(ctx context.Context, req *pb.ReadDeleteRequest) (*pb.ReadResponse, error) {
	config.Logger.Info("read file", "path", req.RemotePath)

	container, err := containerFromContext(ctx)

	if err != nil {
		return nil, err
	}

	encData, err := resources.ReadFile(s.rootDir, container, req.RemotePath)

	if err != nil {
		return nil, err
	}

	data, err := s.cipher.Decrypt(encData, req.RemotePath)

	if err != nil {
		return nil, err
	}

	return &pb.ReadResponse{Data: data}, nil
}

func (s *server) Delete(ctx context.Context, req *pb.ReadDeleteRequest) (*emptypb.Empty, error) {
	config.Logger.Info("delete file", "path", req.RemotePath)

	container, err := containerFromContext(ctx)

	if err != nil {
		return nil, err
	}

	resp, err := resources.DeleteFile(s.rootDir, container, req.RemotePath)

	return resp, err
}

func (s *server) List(ctx context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
	container, err := containerFromContext(ctx)

	if err != nil {
		return nil, err
	}

	result, err := resources.ListFiles(s.rootDir, container, req.Prefix)

	if err != nil {
		return nil, err
	}

	return &pb.ListResponse{RemotePaths: result}, nil
}

func (s *server) CreateContainer(ctx context.Context, req *pb.ContainerRequest) (*emptypb.Empty, error) {
	config.Logger.Info("create container", "name", req.Name)

	err := resources.CreateContainer(s.rootDir, req.Name, s.cfg.DirectoryPermission)

	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *server) DeleteContainer(ctx context.Context, req *pb.ContainerRequest) (*emptypb.Empty, error) {
	config.Logger.Info("delete container", "name", req.Name)

	if req.Name == DefaultContainer {
		return nil, status.Error(codes.PermissionDenied, "default container cannot be deleted")
	}

	err := resources.DeleteContainer(s.rootDir, req.Name)

	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *server) RenameContainer(ctx context.Context, req *pb.RenameRequest) (*emptypb.Empty, error) {
	config.Logger.Info("rename container", "old_name", req.RemotePath, "new_name", req.NewName)

	if req.RemotePath == DefaultContainer {
		return nil, status.Error(codes.PermissionDenied, "default container cannot be renamed")
	}

	err := resources.RenameContainer(s.rootDir, req.RemotePath, req.NewName)

	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func runServer(cfg *config.Config) {
	addr := net.JoinHostPort(cfg.Host, cfg.Port)

	lis, err := net.Listen("tcp", addr)

	if err != nil {
		log.Fatal(err)
	}

	cipher, err := NewCipher(cfg.EncryptionKey)

	if err != nil {
		log.Fatal(err)
	}

	rootDir, err := os.OpenRoot(cfg.StoragePath)

	if err != nil {
		log.Fatal(err)
	}

	err = resources.CreateContainer(rootDir, DefaultContainer, cfg.DirectoryPermission)

	if err != nil && status.Code(err) != codes.AlreadyExists {
		log.Fatal(err)
	}

	defer rootDir.Close()

	// Logging runs first so rejected requests are logged too
	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(LoggingInterceptor(), AuthInterceptor(cfg)))

	pb.RegisterFilesServiceServer(grpcServer, &server{cfg: cfg, cipher: cipher, rootDir: rootDir})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		config.Logger.Info("shutting down")
		grpcServer.GracefulStop()
	}()

	config.Logger.Info("server running", "addr", addr)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}

func main() {
	config.SetupLogger()
	cfg := CliLoadConfig()
	runServer(cfg)
}
