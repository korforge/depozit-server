package server

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "depozit-server/src/protobufs"
	"depozit-server/src/resources"
)

func authContext() context.Context {
	return metadata.AppendToOutgoingContext(context.Background(), "api-key", "123456")
}

func upload(client pb.FilesServiceClient, path string, chunks ...[]byte) error {
	stream, err := client.UploadStream(authContext())

	if err != nil {
		return err
	}

	first := &pb.UploadBinaryRequest{RemotePath: path}

	if len(chunks) > 0 {
		first.Data = chunks[0]
		chunks = chunks[1:]
	}

	if err := stream.Send(first); err != nil {
		return err
	}

	for _, c := range chunks {
		if err := stream.Send(&pb.UploadBinaryRequest{Data: c}); err != nil {
			return err
		}
	}

	_, err = stream.CloseAndRecv()

	return err
}

func read(client pb.FilesServiceClient, path string) ([]byte, error) {
	stream, err := client.ReadStream(authContext(), &pb.ReadDeleteRequest{RemotePath: path})

	if err != nil {
		return nil, err
	}

	var out []byte

	for {
		resp, err := stream.Recv()

		if err == io.EOF {
			return out, nil
		}

		if err != nil {
			return nil, err
		}

		out = append(out, resp.Data...)
	}
}

func TestStreamRoundTrip(t *testing.T) {
	client := pb.NewFilesServiceClient(setupServer(t))

	// Client chunks do not line up with the 64 KiB storage chunks
	var chunks [][]byte
	var want []byte

	for i := range 10 {
		c := bytes.Repeat([]byte{byte(i)}, 30000)
		chunks = append(chunks, c)
		want = append(want, c...)
	}

	require.NoError(t, upload(client, "a/big.bin", chunks...))

	got, err := read(client, "a/big.bin")
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestStreamEmptyFile(t *testing.T) {
	client := pb.NewFilesServiceClient(setupServer(t))

	require.NoError(t, upload(client, "empty.txt"))

	got, err := read(client, "empty.txt")
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestStreamEmptyStream(t *testing.T) {
	client := pb.NewFilesServiceClient(setupServer(t))

	stream, err := client.UploadStream(authContext())
	require.NoError(t, err)

	_, err = stream.CloseAndRecv()
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// A client that cancels mid upload leaves no file behind
func TestStreamCancelledUpload(t *testing.T) {
	client := pb.NewFilesServiceClient(setupServer(t))

	ctx, cancel := context.WithCancel(authContext())
	stream, err := client.UploadStream(ctx)
	require.NoError(t, err)

	require.NoError(t, stream.Send(&pb.UploadBinaryRequest{RemotePath: "x.txt", Data: []byte("partial")}))
	cancel()

	_, err = stream.CloseAndRecv()
	assert.Equal(t, codes.Canceled, status.Code(err))

	_, err = read(client, "x.txt")
	assert.Equal(t, codes.NotFound, status.Code(err))

	list, err := client.List(authContext(), &pb.ListRequest{})
	require.NoError(t, err)
	assert.Empty(t, list.RemotePaths)
}

func TestRenameAndListPages(t *testing.T) {
	client := pb.NewFilesServiceClient(setupServer(t))

	require.NoError(t, upload(client, "notes/hello.txt", []byte("hello")))
	require.NoError(t, upload(client, "z.txt", []byte("z")))

	_, err := client.Rename(authContext(), &pb.RenameRequest{RemotePath: "notes/hello.txt", NewName: "greet.txt"})
	require.NoError(t, err)

	got, err := read(client, "notes/greet.txt")
	require.NoError(t, err)
	assert.Equal(t, []byte("hello"), got)

	page, err := client.List(authContext(), &pb.ListRequest{PageSize: 1})
	require.NoError(t, err)
	assert.Equal(t, []string{"notes/greet.txt"}, page.RemotePaths)

	page, err = client.List(authContext(), &pb.ListRequest{PageSize: 1, PageToken: page.NextPageToken})
	require.NoError(t, err)
	assert.Equal(t, []string{"z.txt"}, page.RemotePaths)
	assert.Empty(t, page.NextPageToken)
}

func TestMissingContainer(t *testing.T) {
	client := pb.NewFilesServiceClient(setupServer(t))

	ctx := metadata.AppendToOutgoingContext(authContext(), "container", "nope")

	_, err := client.List(ctx, &pb.ListRequest{})
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestContainerFromContext(t *testing.T) {
	container, err := resources.ContainerFromContext(context.Background(), DefaultContainer)
	require.NoError(t, err)
	assert.Equal(t, DefaultContainer, container)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("container", "photos"))
	container, err = resources.ContainerFromContext(ctx, DefaultContainer)
	require.NoError(t, err)
	assert.Equal(t, "photos", container)

	for _, name := range []string{".", "a/b", "../x"} {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("container", name))
		_, err := resources.ContainerFromContext(ctx, DefaultContainer)
		assert.Equal(t, codes.InvalidArgument, status.Code(err), name)
	}
}

func TestIsLoopback(t *testing.T) {
	for host, want := range map[string]bool{"localhost": true, "127.0.0.1": true, "::1": true, "0.0.0.0": false, "10.0.0.1": false} {
		assert.Equal(t, want, isLoopback(host), host)
	}
}
