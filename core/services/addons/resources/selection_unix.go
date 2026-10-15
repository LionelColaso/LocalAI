//go:build !windows

package resources

import (
	"os"
	"path/filepath"
	"syscall"
)

func openLockFile(dir string) (*os.File, error) {
	return os.OpenFile(filepath.Join(dir, selectionLock), os.O_CREATE|os.O_RDWR|syscall.O_NOFOLLOW, 0o600)
}

func lockFile(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
}

func unlockFile(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
}
