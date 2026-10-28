//go:build haiku

package daemon

import (
	"context"
	"os"
	"path/filepath"

	containertypes "github.com/moby/moby/api/types/container"
	"github.com/moby/moby/v2/daemon/container"
	"github.com/moby/sys/symlink"
)

// containerFSView on Haiku has no mount-namespace isolation: operations run
// directly against the container's mounted root filesystem.
type containerFSView struct {
	d   *Daemon
	ctr *container.Container
}

func (daemon *Daemon) openContainerFS(ctr *container.Container) (_ *containerFSView, retErr error) {
	if err := daemon.Mount(ctr); err != nil {
		return nil, err
	}
	return &containerFSView{d: daemon, ctr: ctr}, nil
}

func (vw *containerFSView) RunInFS(ctx context.Context, fn func() error) error {
	return fn()
}

func (vw *containerFSView) GoInFS(ctx context.Context, fn func()) error {
	fn()
	return nil
}

func (vw *containerFSView) Close() error {
	return vw.d.Unmount(vw.ctr)
}

func (vw *containerFSView) Stat(ctx context.Context, path string) (*containertypes.PathStat, error) {
	full := filepath.Join(vw.ctr.BaseFS, path)
	lstat, err := os.Lstat(full)
	if err != nil {
		return nil, err
	}
	var target string
	if lstat.Mode()&os.ModeSymlink != 0 {
		target, err = symlink.FollowSymlinkInScope(full, vw.ctr.BaseFS)
		if err != nil {
			return nil, err
		}
	}
	return &containertypes.PathStat{
		Name:       filepath.Base(path),
		Size:       lstat.Size(),
		Mode:       lstat.Mode(),
		Mtime:      lstat.ModTime(),
		LinkTarget: target,
	}, nil
}
