package resources

import "os"

func FileExists(rootDir *os.Root, path string) bool {
	_, err := rootDir.Stat(path)

	return err == nil
}
