//go:build haiku

package mount

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// Unmount releases a mount on Haiku. Symlink aliases are removed; bfs
// volumes are released with the unmount command; plain directories (like
// the container shm dir, which has no tmpfs on Haiku) need nothing.
func Unmount(target string) error {
	fi, err := os.Lstat(target)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		return os.Remove(target)
	}
	if fi.IsDir() {
		var st, parent syscall.Stat_t
		if err := syscall.Stat(target, &st); err == nil {
			if err := syscall.Stat(filepath.Dir(target), &parent); err == nil && st.Dev != parent.Dev {
				cmd := exec.Command("unmount", target)
				return cmd.Run()
			}
		}
	}
	return nil
}

// RecursiveUnmount on Haiku only the root itself is realized.
func RecursiveUnmount(target string) error {
	return Unmount(target)
}
