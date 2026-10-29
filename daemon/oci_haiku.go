//go:build haiku

package daemon

import (
	"context"
	"path"
	"strconv"
	"strings"

	"github.com/containerd/containerd/v2/core/containers"
	coci "github.com/containerd/containerd/v2/pkg/oci"
	containertypes "github.com/moby/moby/api/types/container"
	dconfig "github.com/moby/moby/v2/daemon/config"
	"github.com/moby/moby/v2/daemon/container"
	mobyoci "github.com/moby/moby/v2/daemon/pkg/oci"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// createSpec builds the OCI runtime specification for a container on
// air/OS. Haiku has no namespaces, cgroups or capability sets, so the spec
// carries only what a chroot-based runtime needs: the root filesystem
// (resolved from the containerd snapshot), the process description, and
// the mount table.
func (daemon *Daemon) createSpec(ctx context.Context, daemonCfg *configStore, c *container.Container, mounts []container.Mount) (retSpec *specs.Spec, _ error) {
	s := mobyoci.DefaultSpec()
	// DefaultSpec targets Linux; the haiku runtime has no Linux section.
	s.Linux = nil

	var args []string
	if ep := c.Config.Entrypoint; len(ep) > 0 {
		args = append(args, ep...)
	}
	args = append(args, c.Config.Cmd...)

	cwd := c.Config.WorkingDir
	if cwd == "" {
		cwd = "/"
	}

	opts := []coci.SpecOpts{
		coci.WithProcessArgs(args...),
		coci.WithEnv(c.Config.Env),
		coci.WithProcessCwd(cwd),
		coci.WithHostname(c.Config.Hostname),
		coci.WithDomainname(c.Config.Domainname),
		func(_ context.Context, _ coci.Client, _ *containers.Container, s *specs.Spec) error {
			// The haiku shim gives the process a pty when this is set.
			s.Process.Terminal = c.Config.Tty
			if c.Config.Tty && len(c.HostConfig.ConsoleSize) == 2 {
				s.Process.ConsoleSize = &specs.Box{
					Height: c.HostConfig.ConsoleSize[0],
					Width:  c.HostConfig.ConsoleSize[1],
				}
			}
			return nil
		},
	}

	var snapshotter, snapshotKey string
	if daemon.UsesSnapshotter() {
		snapshotter = daemon.imageService.StorageDriver()
		snapshotKey = c.ID
		opts = append(opts, coci.WithRootFSPath(c.BaseFS))
	} else {
		opts = append(opts, coci.WithRootFSPath(c.BaseFS))
	}

	// User: numeric uid/gid is all the haiku runtime supports for now.
	var user specs.User
	if c.Config.User != "" {
		parts := strings.SplitN(c.Config.User, ":", 2)
		if uid, err := strconv.Atoi(parts[0]); err == nil {
			user.UID = uint32(uid)
		}
		if len(parts) == 2 {
			if gid, err := strconv.Atoi(parts[1]); err == nil {
				user.GID = uint32(gid)
			}
		}
	}
	opts = append(opts, func(_ context.Context, _ coci.Client, _ *containers.Container, s *specs.Spec) error {
		s.Process.User = user
		return nil
	})

	// Mounts: bind each source at its destination. Haiku resolves them
	// through symlinks set up by the snapshotter mounter.
	opts = append(opts, func(_ context.Context, _ coci.Client, _ *containers.Container, s *specs.Spec) error {
		for _, m := range mounts {
			if m.Source == "" {
				continue
			}
			spec := specs.Mount{
				Destination: path.Clean(m.Destination),
				Type:        "bind",
				Source:      m.Source,
			}
			if m.Writable {
				spec.Options = []string{"rbind", "rw"}
			} else {
				spec.Options = []string{"rbind", "ro"}
			}
			if m.Propagation != "" {
				spec.Options = append(spec.Options, string(m.Propagation))
			}
			s.Mounts = append(s.Mounts, spec)
		}
		return nil
	})

	if err := coci.ApplyOpts(ctx, daemon.containerdClient, &containers.Container{
		ID:          c.ID,
		Snapshotter: snapshotter,
		SnapshotKey: snapshotKey,
	}, &s, opts...); err != nil {
		return nil, err
	}
	return &s, nil
}

// mergeUlimits merges daemon ulimits into the host config; Haiku has no
// platform-specific defaults.
func (daemon *Daemon) mergeUlimits(c *containertypes.HostConfig, daemonCfg *dconfig.Config) {
	ulimits := c.Ulimits
	ulIdx := make(map[string]struct{})
	for _, ul := range ulimits {
		ulIdx[ul.Name] = struct{}{}
	}
	for name, ul := range daemonCfg.Ulimits {
		if _, exists := ulIdx[name]; !exists {
			ulimits = append(ulimits, ul)
		}
	}
	c.Ulimits = ulimits
}
