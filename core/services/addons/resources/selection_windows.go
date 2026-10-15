//go:build windows

package resources

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
)

var (
	lockMu sync.Mutex
	locks  = make(map[string]bool)
)

func openLockFile(dir string) (*os.File, error) {
	return os.OpenFile(filepath.Join(dir, selectionLock), os.O_CREATE|os.O_RDWR, 0o600)
}

func lockFile(f *os.File) error {
	path := f.Name()
	lockMu.Lock()
	defer lockMu.Unlock()
	if locks[path] {
		return errors.New("already locked")
	}
	locks[path] = true
	return nil
}

func unlockFile(f *os.File) error {
	path := f.Name()
	lockMu.Lock()
	defer lockMu.Unlock()
	delete(locks, path)
	return nil
}
