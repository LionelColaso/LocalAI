//go:build windows

package catalog

import (
	"fmt"
	"os"
)

// readSourcesFile reads the sources file on Windows using native paths.
func readSourcesFile(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	// On Windows, skip the Unix-style permission check since Windows
	// doesn't use Unix-style file permissions (0600). The file is
	// already protected by being inside the user's data directory.
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("sources file %s must be a regular file", path)
	}
	return os.ReadFile(path)
}
