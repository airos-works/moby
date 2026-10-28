//go:build haiku

package snapshotter

import (
	"fmt"
	"os"

	"github.com/containerd/containerd/v2/core/mount"
	"github.com/moby/sys/mountinfo"
)

// Haiku has no bind-mount syscall. The snapshotters used on air/OS (native)
// return a single bind mount per snapshot, which on Linux would alias the
// snapshot directory at the target. Emulate that with a symlink: file
// operations through the target resolve to the snapshot content, which is
// all the Engine needs to build a container rootfs.

// isMounted reports whether the target is a live (symlinked) mount point.
func isMounted(path string) bool {
	fi, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeSymlink != 0
}

func unmount(target string) error {
	if fi, err := os.Lstat(target); err == nil && fi.Mode()&os.ModeSymlink != 0 {
		return os.Remove(target)
	}
	return nil
}

// applyMounts emulates a mount set at target. Only the single-bind case is
// supported, which is what the native snapshotter produces.
func applyMounts(mounts []mount.Mount, target string) error {
	if len(mounts) == 0 {
		return nil
	}
	if len(mounts) != 1 || mounts[0].Type != "bind" {
		return fmt.Errorf("haiku: unsupported mount set: %v", mounts)
	}
	source := mounts[0].Source
	if fi, err := os.Lstat(target); err == nil {
		if fi.Mode()&os.ModeSymlink != 0 {
			// Already linked; verify it points at the right place.
			if dest, err := os.Readlink(target); err == nil && dest == source {
				return nil
			}
			if err := os.Remove(target); err != nil {
				return err
			}
		} else if fi.IsDir() {
			if err := os.Remove(target); err != nil {
				return fmt.Errorf("haiku: replacing mount dir with symlink: %w", err)
			}
		} else {
			return fmt.Errorf("haiku: mount target %s exists and is not a symlink", target)
		}
	}
	return os.Symlink(source, target)
}

var _ = mountinfo.Mounted
