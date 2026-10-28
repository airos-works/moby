//go:build haiku

package libnetwork

import (
	"context"

	"github.com/moby/moby/v2/daemon/libnetwork/osl"
	"github.com/moby/moby/v2/daemon/libnetwork/types"
)

// Haiku's sandbox has no separate network namespace resources to release.
func releaseOSSboxResources(ns *osl.Namespace, ep *Endpoint) {}

// updateGateway is a no-op without a gateway network on Haiku.
func (sb *Sandbox) updateGateway(ep4, ep6 *Endpoint) error { return nil }

// canPopulateNetworkResources reports whether the sandbox can hold network
// resources; Haiku sandboxes hold no osl resources.
func (sb *Sandbox) canPopulateNetworkResources() bool { return false }

// ExecFunc is part of the DNSBackend surface; there is no sandbox namespace
// to execute in, so the function runs directly.
func (sb *Sandbox) ExecFunc(f func()) error {
	f()
	return nil
}

// IPv6Enabled reports the sandbox's IPv6 support: ok=false on Haiku so
// callers keep the host behaviour.
func (sb *Sandbox) IPv6Enabled() (enabled, ok bool) { return false, false }

func (sb *Sandbox) restoreOslSandbox(ctx context.Context) error { return nil }

func (sb *Sandbox) populateNetworkResourcesOS(ctx context.Context, ep *Endpoint) error {
	return nil
}

// Statistics returns per-interface sandbox statistics; Haiku sandboxes hold
// no interfaces, so the map is empty.
func (sb *Sandbox) Statistics() (map[string]*types.InterfaceStatistics, error) {
	return nil, nil
}

// NetnsPath reports that no network namespace path exists.
func (sb *Sandbox) NetnsPath() (path string, ok bool) { return "", false }
