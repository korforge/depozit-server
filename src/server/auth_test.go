package server

import (
	"context"
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"

	"depozit-server/src/config"
	pb "depozit-server/src/protobufs"
)

// The logger is global, set it once so servers left over from earlier tests never race with a new one
func TestMain(m *testing.M) {
	config.SetupLogger()

	os.Exit(m.Run())
}

func setupServer(t *testing.T) *grpc.ClientConn {
	cfg := &config.Config{APIKey: "123456", EncryptionKey: make([]byte, 32), MaxConcurrentStreams: 100}

	rootDir, err := os.OpenRoot(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { rootDir.Close() })

	srv, err := newServer(cfg, rootDir)
	require.NoError(t, err)

	grpcServer, err := newGrpcServer(cfg)
	require.NoError(t, err)

	pb.RegisterFilesServiceServer(grpcServer, srv)

	lis := bufconn.Listen(1024 * 1024)

	go grpcServer.Serve(lis)
	t.Cleanup(grpcServer.Stop)

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	return conn
}

// Calls every RPC of the service with an empty request and returns the status code
func callAll(t *testing.T, conn *grpc.ClientConn, ctx context.Context) map[string]codes.Code {
	desc := pb.FilesService_ServiceDesc
	result := map[string]codes.Code{}

	for _, m := range desc.Methods {
		method := "/" + desc.ServiceName + "/" + m.MethodName
		err := conn.Invoke(ctx, method, &emptypb.Empty{}, &emptypb.Empty{})
		result[m.MethodName] = status.Code(err)
	}

	for _, s := range desc.Streams {
		method := "/" + desc.ServiceName + "/" + s.StreamName
		stream, err := conn.NewStream(ctx, &grpc.StreamDesc{ServerStreams: s.ServerStreams, ClientStreams: s.ClientStreams}, method)
		require.NoError(t, err)

		if !s.ClientStreams {
			require.NoError(t, stream.SendMsg(&emptypb.Empty{}))
		}

		require.NoError(t, stream.CloseSend())
		result[s.StreamName] = status.Code(stream.RecvMsg(&emptypb.Empty{}))
	}

	return result
}

func TestEveryRPCRequiresAPIKey(t *testing.T) {
	conn := setupServer(t)

	contexts := map[string]context.Context{
		"no api-key":    context.Background(),
		"wrong api-key": metadata.AppendToOutgoingContext(context.Background(), "api-key", "wrong"),
	}

	for name, ctx := range contexts {
		for method, code := range callAll(t, conn, ctx) {
			assert.Equal(t, codes.Unauthenticated, code, "%s: %s", name, method)
		}
	}
}

func TestValidAPIKeyPassesAuth(t *testing.T) {
	conn := setupServer(t)
	ctx := metadata.AppendToOutgoingContext(context.Background(), "api-key", "123456")

	for method, code := range callAll(t, conn, ctx) {
		assert.NotEqual(t, codes.Unauthenticated, code, method)
	}
}
