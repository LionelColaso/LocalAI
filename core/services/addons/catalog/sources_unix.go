//go:build !windows

package catalog

import (
	"fmt"
	"os"
	"path/filepath"
)

// readSourcesFile reads the sources file on Unix using filepath.ToSlash for consistency.
func readSourcesFile(path string) ([]byte, error) {
	info, err := os.Lstat(filepath.ToSlash(path))
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 {
		return nil, fmt.Errorf("sources file %s must be a regular file with mode 0600 or less", path)
	}
	return os.ReadFile(path)
}
