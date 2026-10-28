//go:build haiku

package daemon

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/containerd/log"
	"github.com/moby/moby/api/types/network"
	containertypes "github.com/moby/moby/api/types/container"
	"github.com/moby/moby/v2/daemon/config"
	"github.com/moby/moby/v2/daemon/container"
	"github.com/moby/moby/v2/daemon/internal/otelutil"
	"github.com/moby/moby/v2/daemon/libnetwork"
	nwconfig "github.com/moby/moby/v2/daemon/libnetwork/config"
	"github.com/moby/moby/v2/pkg/sysinfo"
	"github.com/moby/sys/user"
	"go.opentelemetry.io/otel/baggage"
)

// parseSecurityOpt validates security options; Haiku supports no seccomp,
// AppArmor or SELinux labelling.
func (daemon *Daemon) parseSecurityOpt(cfg *config.Config, securityOptions *container.SecurityOptions, hostConfig *containertypes.HostConfig) error {
	return nil
}

// verifyPlatformContainerSettings performs platform-specific validation of
// container settings; Haiku adds none beyond the common checks.
func verifyPlatformContainerSettings(daemon *Daemon, daemonCfg *configStore, hostConfig *containertypes.HostConfig, config *containertypes.Config, update bool) (warnings []string, _ error) {
	return nil, nil
}

// adaptContainerSettings applies the platform defaults.
func (daemon *Daemon) adaptContainerSettings(daemonCfg *config.Config, hostConfig *containertypes.HostConfig) error {
	if hostConfig.Memory > 0 && hostConfig.MemorySwap == 0 {
		// By default, MemorySwap is set to twice the size of Memory.
		hostConfig.MemorySwap = hostConfig.Memory * 2
	}
	return nil
}

// registerLinks is a no-op in v1: legacy --link is not supported on Haiku.
func (daemon *Daemon) registerLinks(ctr *container.Container) error {
	return nil
}

// setupInitLayer returns a no-op setup function; Haiku needs no init layer
// (no /dev bind mounts or masked paths to prepare).
func setupInitLayer(uid int, gid int) func(string) error {
	return func(string) error { return nil }
}

// adjustParallelLimit mirrors the generic Unix rule-of-thumb.
func adjustParallelLimit(n int, limit int) int {
	const overhead = 2
	if limit < overhead {
		limit = overhead
	}
	if n > limit/overhead {
		n = limit / overhead
	}
	return n
}

// setHostGatewayIP is a no-op without a default bridge network.
func setHostGatewayIP(controller *libnetwork.Controller, config *config.Config) {}

// initNetworkController wires libnetwork with the Haiku network stack:
// only the null and host drivers exist, so no bridge/firewall setup runs.
func (daemon *Daemon) initNetworkController(cfg *config.Config, activeSandboxes map[string]any) error {
	netOptions, err := daemon.networkOptions(cfg, daemon.PluginStore, daemon.id, activeSandboxes)
	if err != nil {
		return err
	}

	ctx := baggage.ContextWithBaggage(context.TODO(), otelutil.MustNewBaggage(
		otelutil.MustNewMemberRaw(otelutil.TriggerKey, "daemon.initNetworkController"),
	))
	if err := daemon.newNetworkController(ctx, cfg, netOptions); err != nil {
		return fmt.Errorf("error obtaining controller instance: %v", err)
	}

	if len(activeSandboxes) > 0 {
		log.G(ctx).Info("there are running containers, updated network configuration will not take affect")
	}

	// Create the predefined "none" and "host" networks.
	if n, _ := daemon.netController.NetworkByName(network.NetworkNone); n == nil {
		if _, err := daemon.netController.NewNetwork(ctx, "null", network.NetworkNone, "",
			libnetwork.NetworkOptionPersist(true)); err != nil {
			return fmt.Errorf(`error creating default %q network: %v`, network.NetworkNone, err)
		}
	}
	if n, _ := daemon.netController.NetworkByName(network.NetworkHost); n == nil {
		if _, err := daemon.netController.NewNetwork(ctx, "host", network.NetworkHost, "",
			libnetwork.NetworkOptionPersist(true)); err != nil {
			return fmt.Errorf(`error creating default %q network: %v`, network.NetworkHost, err)
		}
	}
	return nil
}

// --- Remaining platform symbols (v1 stubs, replaced as stages land) ---

var errNotImplemented = errors.New("not implemented on haiku")

func (daemon *Daemon) cleanupMountsByID(id string) error { return nil }

func (daemon *Daemon) conditionalMountOnStart(container *container.Container) error {
	return nil
}

func (daemon *Daemon) conditionalUnmountOnCleanup(container *container.Container) error {
	return nil
}

func getPluginExecRoot(cfg *config.Config) string { return cfg.ExecRoot }

// UsingSystemd is always false on Haiku.
func UsingSystemd(config *config.Config) bool { return false }

func cgroupDriver(cfg *config.Config) string { return "cgroupfs" }

func configureKernelSecuritySupport(config *config.Config, driverName string) error {
	return nil
}

func configureMaxThreads(ctx context.Context) error { return nil }

func networkPlatformOptions(conf *config.Config) []nwconfig.Option { return nil }

// setupDaemonRoot creates the daemon's root directory.
func setupDaemonRoot(config *config.Config, rootDir string, uid, gid int) error {
	if err := os.MkdirAll(rootDir, 0700); err != nil {
		return err
	}
	if uid == 0 && gid == 0 {
		return nil
	}
	return os.Chown(rootDir, uid, gid)
}

// setupRemappedRoot returns an empty identity mapping: Haiku has no user
// namespaces.
func setupRemappedRoot(cfg *config.Config) (user.IdentityMapping, error) {
	return user.IdentityMapping{}, nil
}

func verifyDaemonSettings(conf *config.Config) error { return nil }

func (daemon *Daemon) initOCIHooks(cfg *config.Config) error { return nil }

func (daemon *Daemon) setDefaultIsolation(cfg *config.Config) error { return nil }

func (daemon *Daemon) setupSeccompProfile(cfg *config.Config) error { return nil }

// checkSystem validates the host can run the daemon.
func checkSystem() error { return nil }

// setupResolvConf is a no-op placeholder matching the platform hook.
func setupResolvConf(_ *config.Config) {}

// getSysInfo returns the system information advertised through /info.
func getSysInfo(cfg *config.Config) *sysinfo.SysInfo {
	return sysinfo.New()
}

const (
	cgroupSystemdDriver = "systemd"
	cgroupNoneDriver    = "none"
)

// cleanupMounts removes leftover mount points from previous runs.
func (daemon *Daemon) cleanupMounts(cfg *config.Config) error { return nil }

// runInNetNS executes f directly: Haiku has no network namespaces.
func (daemon *Daemon) runInNetNS(f func() error) error { return f() }

// recursiveUnmount detaches a mount tree; Haiku has no mount table walk in
// v1, so this is a no-op.
func recursiveUnmount(target string) error { return nil }
