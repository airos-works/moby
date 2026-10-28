//go:build haiku

package plugin

import (
	v2 "github.com/moby/moby/v2/daemon/pkg/plugin/v2"
)

// Haiku runs no managed plugins in v1: enabling/restoring is unsupported.
func (pm *Manager) enable(p *v2.Plugin, c *controller, force bool) error {
	return errNotSupported
}

func (pm *Manager) restore(p *v2.Plugin, c *controller) error {
	return errNotSupported
}

func recursiveUnmount(target string) error {
	return nil
}

// Shutdown stops managed plugins; none run on Haiku in v1.
func (pm *Manager) Shutdown() {}
