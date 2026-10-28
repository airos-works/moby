//go:build haiku

package fs

import (
	"io"
	"os"

	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

// copyFileInfo/copyFileTimestamp come from copy_unix.go; Haiku provides the
// copy and mknod helpers here.

func copyFile(source, target string) error {
	src, err := os.Open(source)
	if err != nil {
		return errors.Wrapf(err, "failed to open source %s", source)
	}
	defer src.Close()

	dst, err := os.Create(target)
	if err != nil {
		return errors.Wrapf(err, "failed to open target %s", target)
	}

	err = copyFileContent(dst, src)
	cerr := dst.Close()
	if err != nil {
		return err
	}
	return cerr
}

func copyFileContent(dst, src *os.File) error {
	_, err := io.Copy(dst, src)
	return err
}

func mknod(dst string, mode uint32, rDev int) error {
	return unix.Mknod(dst, mode, rDev)
}
