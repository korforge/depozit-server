package list_files

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"depozit-server/src/config"
	. "depozit-server/tests/testutil"
)

func TestMain(m *testing.M) {
	config.SetupLogger()

	os.Exit(m.Run())
}

func upload(t *testing.T, store *Store, paths ...string) {
	for _, p := range paths {
		require.NoError(t, Upload(store, p, []byte("test"), false), p)
	}
}

func list(t *testing.T, store *Store, prefix string) []string {
	files, next, err := store.List(Container, prefix, "", 0)
	require.NoError(t, err)
	assert.Empty(t, next)

	return files
}

func TestListFiles(t *testing.T) {
	_, store := SetupStore(t)

	upload(t, store, "expected/test001.txt", "test002.txt")

	assert.Equal(t, []string{"expected/test001.txt"}, list(t, store, "expected"))
}

func TestListFilesAll(t *testing.T) {
	_, store := SetupStore(t)

	upload(t, store, "a/b.txt", "c.txt")

	files := list(t, store, "")
	assert.Equal(t, []string{"a/b.txt", "c.txt"}, files)

	// Listed names must work as read paths
	for _, f := range files {
		_, err := Read(store, f)
		assert.NoError(t, err, f)
	}
}

func TestListFilesPrefixInsideName(t *testing.T) {
	_, store := SetupStore(t)

	upload(t, store, "photos/a.jpg", "photos-old/b.jpg", "pho.txt", "docs/x.txt")

	assert.Equal(t, []string{"pho.txt", "photos/a.jpg", "photos-old/b.jpg"}, list(t, store, "pho"))
	assert.Equal(t, []string{"photos/a.jpg"}, list(t, store, "photos/"))
	assert.Equal(t, []string{"photos/a.jpg"}, list(t, store, "photos/a"))
}

func TestListFilesHidesUploadsInProgress(t *testing.T) {
	root, store := SetupStore(t)

	upload(t, store, "a.txt")
	require.NoError(t, root.MkdirAll("default/.tmp", 0700))
	require.NoError(t, root.WriteFile("default/.tmp/x", []byte("x"), 0600))

	assert.Equal(t, []string{"a.txt"}, list(t, store, ""))
}

func TestListFilesPrefixDoesNotMatchContainer(t *testing.T) {
	_, store := SetupStore(t)

	upload(t, store, "test.txt")

	assert.Empty(t, list(t, store, "default"))
}

func TestListFilesOtherContainer(t *testing.T) {
	root, store := SetupStore(t)

	require.NoError(t, root.Mkdir("other", 0700))
	require.NoError(t, root.WriteFile("other/x.txt", []byte("test"), 0600))
	upload(t, store, "y.txt")

	assert.Equal(t, []string{"y.txt"}, list(t, store, ""))
}

func TestListFilesContainerNotFound(t *testing.T) {
	_, store := SetupStore(t)

	_, _, err := store.List("missing", "", "", 0)

	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestListFilesPages(t *testing.T) {
	_, store := SetupStore(t)

	var want []string

	// "a-b" sorts before "a/..." as a string but after it in walk order, pages must not skip or repeat it
	for _, p := range []string{"a/1.txt", "a/2.txt", "a-b.txt", "b/c/d.txt", "b/e.txt", "z.txt"} {
		upload(t, store, p)
	}

	want = list(t, store, "")
	require.Len(t, want, 6)

	for pageSize := 1; pageSize <= 7; pageSize++ {
		var got []string
		token := ""

		for range 10 {
			files, next, err := store.List(Container, "", token, pageSize)
			require.NoError(t, err)
			assert.LessOrEqual(t, len(files), pageSize)

			got = append(got, files...)
			token = next

			if token == "" {
				break
			}
		}

		assert.Equal(t, want, got, fmt.Sprint("page size ", pageSize))
	}
}
