package delete_file

import (
	"depozit-server/src/config"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const container = "default"

func TestMain(m *testing.M) {
	config.SetupLogger()

	os.Exit(m.Run())
}

// Opens a root in a temp dir with an empty "default" container
func SetupRoot(t *testing.T) *os.Root {
	root, err := os.OpenRoot(t.TempDir())
	require.NoError(t, err)

	t.Cleanup(func() { root.Close() })

	require.NoError(t, root.Mkdir(container, 0700))

	return root
}
