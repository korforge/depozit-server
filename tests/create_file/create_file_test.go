package create_file

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"depozit-server/src/config"
	"depozit-server/src/security"
	. "depozit-server/tests/testutil"
)

func TestMain(m *testing.M) {
	config.SetupLogger()

	os.Exit(m.Run())
}

func TestUpload(t *testing.T) {
	root, store := SetupStore(t)

	require.NoError(t, Upload(store, "test.txt", []byte("test 001"), false))

	content, err := Read(store, "test.txt")
	require.NoError(t, err)
	assert.Equal(t, []byte("test 001"), content)

	// Stored encrypted
	raw, err := root.ReadFile("default/test.txt")
	require.NoError(t, err)
	assert.NotContains(t, string(raw), "test 001")
}

func TestUploadSizes(t *testing.T) {
	_, store := SetupStore(t)

	for _, size := range []int{0, 1, config.ChunkSize, 3*config.ChunkSize + 7} {
		data := bytes.Repeat([]byte{'x'}, size)
		require.NoError(t, Upload(store, "test.txt", data, true), size)

		content, err := Read(store, "test.txt")
		require.NoError(t, err, size)
		assert.Equal(t, size, len(content), size)
	}
}

func TestUploadNested(t *testing.T) {
	root, store := SetupStore(t)

	require.NoError(t, Upload(store, "a/b/c.txt", []byte("test"), false))

	_, err := root.Stat("default/a/b/c.txt")
	assert.NoError(t, err)
}

func TestUploadAlreadyExists(t *testing.T) {
	_, store := SetupStore(t)

	// Paths must resolve against the root, not the process working directory
	t.Chdir(t.TempDir())

	require.NoError(t, Upload(store, "test.txt", []byte("first"), false))

	err := Upload(store, "test.txt", []byte("second"), false)
	assert.Equal(t, codes.AlreadyExists, status.Code(err))

	content, err := Read(store, "test.txt")
	require.NoError(t, err)
	assert.Equal(t, []byte("first"), content)
}

func TestUploadOverwrite(t *testing.T) {
	_, store := SetupStore(t)

	require.NoError(t, Upload(store, "test.txt", []byte("first"), false))
	require.NoError(t, Upload(store, "test.txt", []byte("second"), true))

	content, err := Read(store, "test.txt")
	require.NoError(t, err)
	assert.Equal(t, []byte("second"), content)
}

// Two uploads of the same path: the one that started first holds the path, the second gets AlreadyExists
func TestUploadNoOverwriteRace(t *testing.T) {
	_, store := SetupStore(t)

	pr, pw := io.Pipe()
	done := make(chan error)

	go func() { done <- store.Upload(Container, "test.txt", false, pr) }()

	// The second write returns only after the first chunk is written, so the temp file exists by then
	pw.Write([]byte("slow"))
	pw.Write([]byte("more"))

	assert.Equal(t, codes.AlreadyExists, status.Code(Upload(store, "test.txt", []byte("fast"), false)))
	pw.Close()

	require.NoError(t, <-done)

	content, err := Read(store, "test.txt")
	require.NoError(t, err)
	assert.Equal(t, []byte("slowmore"), content)
}

func TestUploadInvalidPath(t *testing.T) {
	_, store := SetupStore(t)

	for _, path := range []string{"../x", "/etc/passwd", "", ".", "a/../../x", ".tmp/x", ".tmp"} {
		err := Upload(store, path, []byte("test"), true)
		assert.Equal(t, codes.InvalidArgument, status.Code(err), path)
	}
}

func TestUploadContainerNotFound(t *testing.T) {
	root, store := SetupStore(t)

	err := store.Upload("missing", "test.txt", true, bytes.NewReader([]byte("test")))
	assert.Equal(t, codes.NotFound, status.Code(err))

	_, err = root.Stat("missing")
	assert.Error(t, err)
}

func TestUploadMkdirAllFail(t *testing.T) {
	root, store := SetupStore(t)

	// A file where a parent directory should be makes MkdirAll fail
	require.NoError(t, root.WriteFile("default/a", []byte("file"), 0600))

	err := Upload(store, "a/test.txt", []byte("test"), true)
	assert.Equal(t, codes.Internal, status.Code(err))
}

func TestUploadTooLarge(t *testing.T) {
	root, store := SetupStoreWithLimit(t, 10)

	require.NoError(t, Upload(store, "ok.txt", make([]byte, 10), false))

	err := Upload(store, "big.txt", make([]byte, 11), false)
	assert.Equal(t, codes.ResourceExhausted, status.Code(err))

	_, err = root.Stat("default/big.txt")
	assert.Error(t, err)
}

// A failed stream leaves neither the file nor a temporary file
func TestUploadStreamError(t *testing.T) {
	root, store := SetupStore(t)

	r := io.MultiReader(bytes.NewReader([]byte("partial")), errReader{})

	err := store.Upload(Container, "test.txt", false, r)
	assert.Error(t, err)

	_, err = root.Stat("default/test.txt")
	assert.Error(t, err)

	entries, err := fs.ReadDir(root.FS(), "default/.tmp")
	require.NoError(t, err)
	assert.Empty(t, entries)
}

// Leftover uploads from a crash are removed on startup
func TestNewStoreCleansTempFiles(t *testing.T) {
	root, _ := SetupStore(t)

	require.NoError(t, root.MkdirAll("default/.tmp", 0700))
	require.NoError(t, root.WriteFile("default/.tmp/leftover", []byte("x"), 0600))

	cipher, err := security.NewCipher(make([]byte, 32))
	require.NoError(t, err)

	_, err = NewStore(root, cipher, 0)
	require.NoError(t, err)

	_, err = root.Stat("default/.tmp")
	assert.Error(t, err)
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("stream broken") }
