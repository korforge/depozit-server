package resources

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"depozit-server/src/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	config.SetupLogger()

	os.Exit(m.Run())
}

func TestCleanStaleTempFiles(t *testing.T) {
	dir := t.TempDir()
	root, err := os.OpenRoot(dir)
	require.NoError(t, err)
	t.Cleanup(func() { root.Close() })

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "c1", TmpDir), 0700))
	require.NoError(t, os.Mkdir(filepath.Join(dir, "c2"), 0700))

	stale := filepath.Join(dir, "c1", TmpDir, "stale")
	fresh := filepath.Join(dir, "c1", TmpDir, "fresh")
	require.NoError(t, os.WriteFile(stale, []byte("x"), 0600))
	require.NoError(t, os.WriteFile(fresh, []byte("x"), 0600))

	old := time.Now().Add(-2 * time.Hour)
	require.NoError(t, os.Chtimes(stale, old, old))

	require.NoError(t, CleanStaleTempFiles(root, time.Hour))

	assert.NoFileExists(t, stale)
	assert.FileExists(t, fresh)
}
