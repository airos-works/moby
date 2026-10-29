package local

import (
	"os"
	"syscall"
	"time"

	"github.com/moby/moby/v2/errdefs"
	"github.com/pkg/errors"
)

// Local volumes on Haiku are plain directories under the daemon root, as on
// Windows: the daemon cannot mount nfs/cifs/tmpfs devices (mount.Mount is a
// no-op on Haiku, so accepting "type"/"device" would silently give an empty
// directory instead), and Haiku has no project quotas for "size".
type optsConfig struct{}

func (r *Root) validateOpts(opts map[string]string) error {
	if len(opts) == 0 {
		return nil
	}
	return errdefs.InvalidParameter(errors.New("options are not supported on this platform"))
}

func (v *localVolume) setOpts(opts map[string]string) error {
	return nil
}

func (v *localVolume) needsMount() bool {
	return false
}

func (v *localVolume) mount() error {
	return nil
}

func (v *localVolume) unmount() error {
	return nil
}

func (v *localVolume) postMount() error {
	return nil
}

// restoreIfMounted is a no-op on Haiku (because mounts are not supported).
func (v *localVolume) restoreIfMounted() error {
	return nil
}

func (v *localVolume) CreatedAt() (time.Time, error) {
	fileInfo, err := os.Stat(v.rootPath)
	if err != nil {
		return time.Time{}, err
	}
	sec, nsec := fileInfo.Sys().(*syscall.Stat_t).Crtim.Unix()
	return time.Unix(sec, nsec), nil
}
