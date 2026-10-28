//go:build haiku

package daemon

import "github.com/moby/moby/v2/daemon/container"

// saveAppArmorConfig is a no-op: Haiku has no AppArmor.
func (daemon *Daemon) saveAppArmorConfig(container *container.Container) error {
	return nil
}
