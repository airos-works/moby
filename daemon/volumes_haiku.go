//go:build haiku

package daemon

import mounttypes "github.com/moby/moby/api/types/mount"

// validateBindDaemonRoot allows the daemon root bind on Haiku.
func (daemon *Daemon) validateBindDaemonRoot(m mounttypes.Mount) (bool, error) {
	return true, nil
}
