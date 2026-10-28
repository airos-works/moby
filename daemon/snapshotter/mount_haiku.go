//go:build haiku

package snapshotter

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/containerd/containerd/v2/core/mount"
	"github.com/moby/sys/mountinfo"
)

// Haiku has no bind-mount syscall, and its chroot(2) only works on mount
// points. The snapshotters used on air/OS (native) return a single bind
// mount per snapshot; we realize that by creating a small bfs image per
// container, mounting it at the target directory, and copying the
// snapshot's content into it. The result is a real mount point, so the
// runtime can chroot into it: containers get filesystem isolation.

// isMounted reports whether the target is a live (mounted) rootfs.
func isMounted(path string) bool {
	var st, parent syscall.Stat_t
	if err := syscall.Stat(path, &st); err != nil {
		return false
	}
	if err := syscall.Stat(filepath.Dir(path), &parent); err != nil {
		return false
	}
	return st.Dev != parent.Dev
}

func unmount(target string) error {
	if !isMounted(target) {
		return nil
	}
	cmd := exec.Command("unmount", target)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("unmount %s: %v (%s)", target, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// imagePathFor returns the backing image file for a mount target.
func imagePathFor(target string) string {
	return target + ".bfs.img"
}

// applyMounts realizes a mount set at target. Only the single-bind case is
// supported, which is what the native snapshotter produces.
func applyMounts(mounts []mount.Mount, target string) error {
	if len(mounts) == 0 {
		return nil
	}
	if len(mounts) != 1 || mounts[0].Type != "bind" {
		return fmt.Errorf("haiku: unsupported mount set: %v", mounts)
	}
	source := mounts[0].Source

	if isMounted(target) {
		return nil
	}

	// Filesystem isolation via bfs rootfs images is experimental: chroot
	// into them is broken on current Haiku (dynamic executables exit
	// silently even from a mount point). Enable with DOCKER_BFS_ROOTFS=1
	// on kernels where chroot works.
	if os.Getenv("DOCKER_BFS_ROOTFS") == "1" {
		if err := applyMountsBFS(source, target); err != nil {
			// Fall back to the alias rather than failing the container.
			return applyMountsSymlink(source, target)
		}
		return nil
	}
	return applyMountsSymlink(source, target)
}

// applyMountsBFS realizes the rootfs as a mounted bfs volume (requires
// <data-root>/blank.bfs.img, a pre-formatted template).
func applyMountsBFS(source, target string) error {
	template := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(target))), "blank.bfs.img")
	if _, err := os.Stat(template); err != nil {
		return fmt.Errorf("no bfs template: %w", err)
	}
	image := imagePathFor(target)
	_ = os.Remove(image)
	if err := copyFile(template, image); err != nil {
		return err
	}
	mout, merr := exec.Command("mount", "-t", "bfs", image, target).CombinedOutput()
	if merr != nil {
		_ = os.Remove(image)
		return fmt.Errorf("mount %s at %s: %v (%s)", image, target, merr, strings.TrimSpace(string(mout)))
	}
	if err := copyDir(source, target); err != nil {
		_ = exec.Command("unmount", target).Run()
		_ = os.Remove(image)
		return err
	}
	return nil
}

// applyMountsSymlink is the no-template fallback: a symlink aliases the
// snapshot directory. Containers run without filesystem isolation.
func applyMountsSymlink(source, target string) error {
	if fi, err := os.Lstat(target); err == nil {
		if fi.Mode()&os.ModeSymlink != 0 {
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

// copyFile copies a file verbatim.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// createImage allocates a zero-filled file of the given size.
func createImage(path string, size int64) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := f.Truncate(size); err != nil {
		return err
	}
	return f.Sync()
}

// formatBFS initializes a bfs volume (mkfs is interactive: answer "yes").
func formatBFS(image, label string) error {
	cmd := exec.Command("mkfs", "-t", "bfs", image, label)
	cmd.Stdin = strings.NewReader("yes\n")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("haiku: mkfs %s: %v (%s)", image, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// copyDir recursively copies the content of src into dst.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		to := filepath.Join(dst, rel)
		mode := info.Mode()
		switch {
		case mode.IsDir():
			return os.MkdirAll(to, mode.Perm())
		case mode.IsRegular():
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			return os.WriteFile(to, data, mode.Perm())
		case mode&os.ModeSymlink != 0:
			dest, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(dest, to)
		}
		return nil
	})
}

var _ = mountinfo.Mounted
