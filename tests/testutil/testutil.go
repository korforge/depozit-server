package testutil

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "depozit-server/src/protobufs"
	"depozit-server/src/resources"
	"depozit-server/src/security"
)

const Container = "default"

// Calls the resource functions the way the gRPC handlers do: open the container, run, close it
type Store struct {
	root        *os.Root
	cipher      *security.Cipher
	maxFileSize int64
}

// Removes leftover temp files, like the server does on startup
func NewStore(root *os.Root, cipher *security.Cipher, maxFileSize int64) (*Store, error) {
	if err := resources.CleanTempFiles(root); err != nil {
		return nil, err
	}

	return &Store{root: root, cipher: cipher, maxFileSize: maxFileSize}, nil
}

// Opens a root in a temp dir with an empty "default" container, and a store on it
func SetupStore(t *testing.T) (*os.Root, *Store) {
	return SetupStoreWithLimit(t, 0)
}

func SetupStoreWithLimit(t *testing.T, maxFileSize int64) (*os.Root, *Store) {
	root, err := os.OpenRoot(t.TempDir())
	require.NoError(t, err)

	t.Cleanup(func() { root.Close() })

	require.NoError(t, root.Mkdir(Container, 0700))

	cipher, err := security.NewCipher(make([]byte, 32))
	require.NoError(t, err)

	store, err := NewStore(root, cipher, maxFileSize)
	require.NoError(t, err)

	return root, store
}

func (s *Store) open(container string) (*os.Root, error) {
	return resources.OpenContainer(s.root, container)
}

// Sends r as an upload stream. The first message carries the path, a read error fails the stream
func (s *Store) Upload(container string, path string, overwrite bool, r io.Reader) error {
	containerDir, err := s.open(container)

	if err != nil {
		return err
	}

	defer containerDir.Close()

	return resources.Upload(containerDir, s.cipher, s.maxFileSize, &uploadStream{r: r, path: path, overwrite: overwrite})
}

// Calls send with each chunk the server would stream to the client
func (s *Store) Read(container string, path string, send func([]byte) error) error {
	containerDir, err := s.open(container)

	if err != nil {
		return err
	}

	defer containerDir.Close()

	return resources.Read(containerDir, s.cipher, path, &readStream{send: send})
}

func (s *Store) Delete(container string, path string) error {
	containerDir, err := s.open(container)

	if err != nil {
		return err
	}

	defer containerDir.Close()

	return resources.Delete(containerDir, path)
}

func (s *Store) Rename(container string, path string, newName string) error {
	containerDir, err := s.open(container)

	if err != nil {
		return err
	}

	defer containerDir.Close()

	return resources.Rename(containerDir, s.cipher, path, newName)
}

func (s *Store) List(container string, prefix string, pageToken string, pageSize int) ([]string, string, error) {
	containerDir, err := s.open(container)

	if err != nil {
		return nil, "", err
	}

	defer containerDir.Close()

	return resources.List(containerDir, prefix, pageToken, pageSize)
}

type uploadStream struct {
	grpc.ServerStream
	r         io.Reader
	path      string
	overwrite bool
	sent      bool
	buf       [32 << 10]byte
}

func (u *uploadStream) Recv() (*pb.UploadBinaryRequest, error) {
	n, err := u.r.Read(u.buf[:])
	msg := &pb.UploadBinaryRequest{Data: bytes.Clone(u.buf[:n])}

	if !u.sent {
		u.sent = true
		msg.RemotePath = u.path
		msg.Overwrite = u.overwrite

		if err == io.EOF {
			err = nil
		}
	}

	if n > 0 && err == io.EOF {
		err = nil
	}

	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (u *uploadStream) SendAndClose(*emptypb.Empty) error { return nil }

func (u *uploadStream) Context() context.Context { return context.Background() }

type readStream struct {
	grpc.ServerStream
	send func([]byte) error
}

func (r *readStream) Send(resp *pb.ReadResponse) error { return r.send(resp.Data) }

func (r *readStream) Context() context.Context { return context.Background() }

func Upload(store *Store, path string, data []byte, overwrite bool) error {
	return store.Upload(Container, path, overwrite, bytes.NewReader(data))
}

func Read(store *Store, path string) ([]byte, error) {
	var out []byte

	err := store.Read(Container, path, func(p []byte) error {
		out = append(out, p...)
		return nil
	})

	return out, err
}
