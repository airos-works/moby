//go:build haiku

package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/netip"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/containerd/containerd/v2/core/containers"
	coci "github.com/containerd/containerd/v2/pkg/oci"
	containertypes "github.com/moby/moby/api/types/container"
	networktypes "github.com/moby/moby/api/types/network"
	"github.com/moby/moby/v2/daemon/container"
	"github.com/moby/moby/v2/daemon/pkg/oci/caps"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// Containers from Linux images run in air/OS's Linux VM (containerd's
// containerd-linuxvm, its agent and the haiku shim). Their spec is a Linux
// one, and these annotations carry what the agent resolves in the VM. They
// match containerd's pkg/linuxvm.
const (
	annotationPlatform   = "works.airos.platform"
	annotationUser       = "works.airos.user"
	annotationNetwork    = "works.airos.network"
	annotationPorts      = "works.airos.ports"
	annotationPrivileged = "works.airos.privileged"
	annotationDataRoot   = "works.airos.dataroot"
)

type linuxVMNetwork struct {
	Mode     string            `json:"mode"`
	Hostname string            `json:"hostname,omitempty"`
	Networks []linuxVMEndpoint `json:"networks,omitempty"`
}

type linuxVMEndpoint struct {
	Name    string   `json:"name"`
	Aliases []string `json:"aliases,omitempty"`
}

type linuxVMPort struct {
	Proto         string `json:"proto"`
	ContainerPort int    `json:"containerPort"`
	HostIP        string `json:"hostIP,omitempty"`
	HostPort      int    `json:"hostPort"`
}

// isLinuxVMContainer reports whether the container's image is a Linux one.
func isLinuxVMContainer(c *container.Container) bool {
	return c.ImagePlatform.OS == "linux"
}

// withLinuxVM makes the spec a Linux container's: Docker's capabilities,
// resources, read-only root, tmpfs mounts and sysctls, and the annotations
// for the VM's agent.
func (daemon *Daemon) withLinuxVM(c *container.Container) coci.SpecOpts {
	return func(_ context.Context, _ coci.Client, _ *containers.Container, s *specs.Spec) error {
		hc := c.HostConfig
		if s.Annotations == nil {
			s.Annotations = map[string]string{}
		}
		s.Annotations[annotationPlatform] = "linux"
		if c.Config.User != "" {
			s.Annotations[annotationUser] = c.Config.User
		}
		if hc.Privileged {
			s.Annotations[annotationPrivileged] = "true"
		}
		if root, err := filepath.EvalSymlinks(daemon.root); err == nil {
			s.Annotations[annotationDataRoot] = root
		}

		capabilities, err := caps.TweakCapabilities(caps.DefaultCapabilities(), hc.CapAdd, hc.CapDrop, hc.Privileged)
		if err != nil {
			return err
		}
		s.Process.Capabilities = &specs.LinuxCapabilities{
			Bounding: capabilities, Permitted: capabilities, Effective: capabilities,
		}
		if !hc.Privileged {
			s.Process.NoNewPrivileges = c.NoNewPrivileges
		}
		for _, ul := range hc.Ulimits {
			s.Process.Rlimits = append(s.Process.Rlimits, specs.POSIXRlimit{
				Type: "RLIMIT_" + strings.ToUpper(ul.Name),
				Soft: uint64(ul.Soft),
				Hard: uint64(ul.Hard),
			})
		}
		s.Root.Readonly = hc.ReadonlyRootfs

		res := linuxVMResources(hc.Resources)
		if s.Linux.Resources != nil {
			// keep the default spec's device rules
			res.Devices = s.Linux.Resources.Devices
		}
		s.Linux.Resources = res

		if len(hc.Sysctls) > 0 {
			s.Linux.Sysctl = map[string]string{}
			for k, v := range hc.Sysctls {
				s.Linux.Sysctl[k] = v
			}
		}
		for i, m := range s.Mounts {
			if m.Destination == "/dev/shm" && hc.ShmSize > 0 {
				s.Mounts[i].Options = append(s.Mounts[i].Options, "size="+strconv.FormatInt(hc.ShmSize, 10))
			}
		}
		for dest, data := range hc.Tmpfs {
			opts := []string{"nosuid", "nodev", "noexec"}
			if data != "" {
				opts = append(opts, strings.Split(data, ",")...)
			}
			s.Mounts = append(s.Mounts, specs.Mount{Destination: dest, Type: "tmpfs", Source: "tmpfs", Options: opts})
		}

		netCfg, err := daemon.linuxVMNetwork(c)
		if err != nil {
			return err
		}
		b, err := json.Marshal(netCfg)
		if err != nil {
			return err
		}
		s.Annotations[annotationNetwork] = string(b)

		ports, err := linuxVMPorts(c)
		if err != nil {
			return err
		}
		if len(ports) > 0 {
			b, err := json.Marshal(ports)
			if err != nil {
				return err
			}
			s.Annotations[annotationPorts] = string(b)
		}
		return nil
	}
}

// linuxVMResources is the cgroup configuration of a container's resource
// limits (docker run and docker update).
func linuxVMResources(hc containertypes.Resources) *specs.LinuxResources {
	r := &specs.LinuxResources{}
	if hc.Memory > 0 || hc.MemoryReservation > 0 || hc.MemorySwap != 0 {
		r.Memory = &specs.LinuxMemory{}
		if hc.Memory > 0 {
			r.Memory.Limit = &hc.Memory
		}
		if hc.MemoryReservation > 0 {
			r.Memory.Reservation = &hc.MemoryReservation
		}
		if hc.MemorySwap != 0 {
			r.Memory.Swap = &hc.MemorySwap
		}
	}
	if hc.NanoCPUs > 0 || hc.CPUShares > 0 || hc.CpusetCpus != "" || hc.CpusetMems != "" || hc.CPUQuota > 0 {
		r.CPU = &specs.LinuxCPU{Cpus: hc.CpusetCpus, Mems: hc.CpusetMems}
		if hc.CPUShares > 0 {
			shares := uint64(hc.CPUShares)
			r.CPU.Shares = &shares
		}
		switch {
		case hc.NanoCPUs > 0:
			period := uint64(100000)
			quota := hc.NanoCPUs * int64(period) / 1e9
			r.CPU.Period, r.CPU.Quota = &period, &quota
		case hc.CPUQuota > 0:
			quota := hc.CPUQuota
			r.CPU.Quota = &quota
			if hc.CPUPeriod > 0 {
				period := uint64(hc.CPUPeriod)
				r.CPU.Period = &period
			}
		}
	}
	if hc.PidsLimit != nil && *hc.PidsLimit > 0 {
		r.Pids = &specs.LinuxPids{Limit: hc.PidsLimit}
	}
	return r
}

func (daemon *Daemon) linuxVMNetwork(c *container.Container) (linuxVMNetwork, error) {
	n := linuxVMNetwork{Hostname: c.Config.Hostname}
	nm := c.HostConfig.NetworkMode
	switch {
	case nm.IsHost():
		n.Mode = "host"
		return n, nil
	case nm.IsNone():
		n.Mode = "none"
		return n, nil
	case nm.IsContainer():
		other, err := daemon.GetContainer(nm.ConnectedContainer())
		if err != nil {
			return n, err
		}
		n.Mode = "container:" + other.ID
		return n, nil
	}
	n.Mode = "bridge"
	var names []string
	if c.NetworkSettings != nil {
		for name := range c.NetworkSettings.Networks {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		names = []string{nm.NetworkName()}
		if nm.IsDefault() || names[0] == "" {
			names = []string{"bridge"}
		}
	}
	// The network the container was created on first; the rest in order.
	sort.SliceStable(names, func(i, j int) bool {
		return names[i] == nm.NetworkName() || (names[j] != nm.NetworkName() && names[i] < names[j])
	})
	ctrName := strings.TrimPrefix(c.Name, "/")
	for _, name := range names {
		ep := linuxVMEndpoint{Name: name}
		seen := map[string]bool{}
		add := func(a string) {
			if a != "" && !seen[a] {
				seen[a] = true
				ep.Aliases = append(ep.Aliases, a)
			}
		}
		// Docker's default bridge has no name resolution; user networks
		// resolve container names and aliases (Compose service names).
		if name != "bridge" {
			add(ctrName)
			if len(c.ID) >= 12 {
				add(c.ID[:12])
			}
			if es := c.NetworkSettings.Networks[name]; es != nil && es.EndpointSettings != nil {
				for _, a := range es.Aliases {
					add(a)
				}
				for _, a := range es.DNSNames {
					add(a)
				}
			}
		}
		n.Networks = append(n.Networks, ep)
	}
	return n, nil
}

// linuxVMPorts resolves the published ports, choosing free host ports where
// none was asked for, and records them in the container's network settings
// (docker ps, docker port).
func linuxVMPorts(c *container.Container) ([]linuxVMPort, error) {
	bindings := networktypes.PortMap{}
	for p, b := range c.HostConfig.PortBindings {
		bindings[p] = append([]networktypes.PortBinding(nil), b...)
	}
	if c.HostConfig.PublishAllPorts {
		for p := range c.Config.ExposedPorts {
			if len(bindings[p]) == 0 {
				bindings[p] = []networktypes.PortBinding{{}}
			}
		}
	}
	var out []linuxVMPort
	published := networktypes.PortMap{}
	for p, bs := range bindings {
		proto := string(p.Proto())
		if proto != "tcp" && proto != "udp" {
			continue
		}
		for _, b := range bs {
			hostIP := ""
			if b.HostIP.IsValid() && !b.HostIP.IsUnspecified() {
				hostIP = b.HostIP.String()
			}
			hostPort := 0
			if b.HostPort != "" {
				first, _, _ := strings.Cut(b.HostPort, "-")
				v, err := strconv.Atoi(first)
				if err != nil {
					return nil, fmt.Errorf("host port %q: %w", b.HostPort, err)
				}
				hostPort = v
			}
			if hostPort == 0 {
				v, err := freePort(proto, hostIP)
				if err != nil {
					return nil, err
				}
				hostPort = v
			}
			out = append(out, linuxVMPort{Proto: proto, ContainerPort: int(p.Num()), HostIP: hostIP, HostPort: hostPort})
			ip := netip.IPv4Unspecified()
			if hostIP != "" {
				ip, _ = netip.ParseAddr(hostIP)
			}
			published[p] = append(published[p], networktypes.PortBinding{HostIP: ip, HostPort: strconv.Itoa(hostPort)})
		}
	}
	if c.NetworkSettings != nil {
		c.NetworkSettings.Ports = published
	}
	return out, nil
}

func freePort(proto, ip string) (int, error) {
	addr := net.JoinHostPort(ip, "0")
	if proto == "udp" {
		pc, err := net.ListenPacket("udp", addr)
		if err != nil {
			return 0, err
		}
		defer pc.Close()
		return pc.LocalAddr().(*net.UDPAddr).Port, nil
	}
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
