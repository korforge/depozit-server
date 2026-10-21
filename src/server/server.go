package server

import (
	"context"
	"depozit-server/src/config"
	pb "depozit-server/src/protobufs"
	"depozit-server/src/resources"
	"depozit-server/src/security"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

const DefaultContainer = "default"

// Largest message a client may send. The Python client sends 64 KiB chunks
const maxRecvMsgSize = 4 * 1024 * 1024

// How long shutdown waits for open streams before it closes them
const shutdownTimeout = 30 * time.Second

func newGrpcServer(cfg *config.Config) (*grpc.Server, error) {
	options := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(LoggingInterceptor(), AuthInterceptor(cfg)),
		grpc.ChainStreamInterceptor(LoggingStreamInterceptor(), AuthStreamInterceptor(cfg)),
		grpc.MaxRecvMsgSize(maxRecvMsgSize),
		grpc.MaxConcurrentStreams(cfg.MaxConcurrentStreams),

		// Drop clients that ping too often, and close idle connections
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{MinTime: 10 * time.Second, PermitWithoutStream: true}),
		grpc.KeepaliveParams(keepalive.ServerParameters{MaxConnectionIdle: 15 * time.Minute, Time: time.Minute, Timeout: 20 * time.Second}),
	}

	if cfg.TLSCert != "" {
		creds, err := credentials.NewServerTLSFromFile(cfg.TLSCert, cfg.TLSKey)

		if err != nil {
			return nil, err
		}

		options = append(options, grpc.Creds(creds))
	} else if !isLoopback(cfg.Host) {
		config.Logger.Warn("TLS disabled, the API key and files are sent in plaintext", "host", cfg.Host)
	}

	return grpc.NewServer(options...), nil
}

func isLoopback(host string) bool {
	ip := net.ParseIP(host)

	return host == "localhost" || (ip != nil && ip.IsLoopback())
}

func newServer(cfg *config.Config, rootDir *os.Root) (*Server, error) {
	cipher, err := security.NewCipher(cfg.EncryptionKey)

	if err != nil {
		return nil, err
	}

	err = resources.CreateContainer(rootDir, config.DefaultContainer, config.DefaultDirectoryPermission)

	if err != nil && status.Code(err) != codes.AlreadyExists {
		return nil, err
	}

	if err := resources.CleanTempFiles(rootDir); err != nil {
		return nil, err
	}

	return &Server{cfg: cfg, rootDir: rootDir, cipher: cipher}, nil
}

func RunServer(cfg *config.Config) {
	addr := net.JoinHostPort(cfg.Host, cfg.Port)

	lis, err := net.Listen("tcp", addr)

	if err != nil {
		log.Fatal(err)
	}

	rootDir, err := os.OpenRoot(cfg.StoragePath)

	if err != nil {
		log.Fatal(err)
	}

	defer rootDir.Close()

	srv, err := newServer(cfg, rootDir)

	if err != nil {
		log.Fatal(err)
	}

	grpcServer, err := newGrpcServer(cfg)

	if err != nil {
		log.Fatal(err)
	}

	pb.RegisterFilesServiceServer(grpcServer, srv)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				log.Println("Implement a go routine to remove file that had the upload interrupted")
			}
		}
	}()

	defer stop()

	go func() {
		<-ctx.Done()
		config.Logger.Info("shutting down")

		done := make(chan struct{})

		go func() {
			grpcServer.GracefulStop()
			close(done)
		}()

		// GracefulStop waits for every stream, so a stuck upload would block shutdown forever
		select {
		case <-done:
		case <-time.After(shutdownTimeout):
			config.Logger.Warn("graceful shutdown timed out, closing open streams")
			grpcServer.Stop()
		}
	}()

	config.Logger.Info("server running", "addr", addr, "tls", cfg.TLSCert != "")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
