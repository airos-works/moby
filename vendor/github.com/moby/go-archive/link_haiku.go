package archive

import (
	"errors"
	"io"
	"os"
	"syscall"
)

// rootLink creates newname as a hard link to oldname within root. BFS,
// Haiku's native file system, has no hard links and refuses them with EPERM,
// so there newname becomes a copy of oldname instead (a symlink stays a
// symlink, as link(2) does not follow it). The files read the same; unlike
// hard links, writing one later does not change the other.
func rootLink(root *os.Root, oldname, newname string) error {
	err := root.Link(oldname, newname)
	if err == nil || !errors.Is(err, syscall.EPERM) && !errors.Is(err, syscall.ENOTSUP) {
		return err
	}
	if cerr := copyLinkTarget(root, oldname, newname); cerr != nil {
		return &os.LinkError{Op: "link (copy)", Old: oldname, New: newname, Err: cerr}
	}
	return nil
}

func copyLinkTarget(root *os.Root, oldname, newname string) error {
	fi, err := root.Lstat(oldname)
	if err != nil {
		return err
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		target, err := root.Readlink(oldname)
		if err != nil {
			return err
		}
		return root.Symlink(target, newname)
	}
	if !fi.Mode().IsRegular() {
		return syscall.EPERM
	}
	src, err := root.Open(oldname)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := root.OpenFile(newname, os.O_WRONLY|os.O_CREATE|os.O_EXCL, fi.Mode().Perm())
	if err != nil {
		return err
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		root.Remove(newname)
		return err
	}
	return dst.Close()
}
