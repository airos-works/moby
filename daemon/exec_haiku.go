//go:build haiku

package daemon

import (
	"context"

	"github.com/moby/moby/v2/daemon/config"
	"github.com/moby/moby/v2/daemon/container"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

func (daemon *Daemon) execSetPlatformOpt(ctx context.Context, daemonCfg *config.Config, ec *container.ExecConfig, p *specs.Process) error {
	if isLinuxVMContainer(ec.Container) && ec.User != "" {
		// The Linux VM's agent resolves it against the container's
		// /etc/passwd and /etc/group.
		p.User = specs.User{Username: ec.User}
	}
	return nil
}

// getUser resolves a username/uid to an OCI user; Haiku keeps the literal
// value until the execution backend lands.
func getUser(c *container.Container, username string) (specs.User, error) {
	return specs.User{Username: username}, nil
}
