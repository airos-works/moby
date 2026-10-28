//go:build haiku

package daemon

import (
	"context"

	containertypes "github.com/moby/moby/api/types/container"
	dconfig "github.com/moby/moby/v2/daemon/config"
	"github.com/moby/moby/v2/daemon/container"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// createSpec builds the OCI runtime specification for a container. The
// Haiku execution backend (stage 3 of the port) replaces this stub.
func (daemon *Daemon) createSpec(ctx context.Context, daemonCfg *configStore, c *container.Container, mounts []container.Mount) (retSpec *specs.Spec, _ error) {
	return nil, errNotImplemented
}

// mergeUlimits merges daemon ulimits into the host config; no platform
// defaults exist on Haiku.
func (daemon *Daemon) mergeUlimits(c *containertypes.HostConfig, daemonCfg *dconfig.Config) {}
