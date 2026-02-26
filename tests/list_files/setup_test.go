package list_files

import (
	"depozit-server/src/config"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	config.SetupLogger()

	m.Run()
}

func SetupTeardownModule(t *testing.T) string {
	tDir := t.TempDir()

	t.Cleanup(func() {
		os.RemoveAll(tDir)
	})

	return tDir
}
