# Docker Engine (dockerd) for air/OS — port scoping

The CLI port lives in [airos-works/docker](https://github.com/airos-works/docker);
this document scopes the Engine (this repo) and lists what Haiku needs.

## Why the Engine is a different scale of work

`cmd/dockerd` imports, directly or transitively: Linux netfilter
(`daemon/libnetwork/internal/nftables`, `daemon/libnetwork/iptables` are
`//go:build linux` and exclude every other GOOS), cgroups v1/v2, mount
namespaces and `pivot_root`, `overlayfs`, seccomp, capabilities,
`AF_VSOCK` hvsock, and `containerd`+`runc` as the runtime plumbing. Haiku
has none of the kernel facilities; a faithful moby port is staged, not
monolithic.

## Stages

0. **CLI, packaging, GUI** — done in airos-works/docker and
   airos-works/docker-desktop. hpkgs ship in the rpi4, Cubie A7S and ROCK 5
   image package sets today.
1. **Compile skeleton.** **DONE (2026-10-28).**
   **Runtime milestone (2026-10-28, QEMU x86_64, hrev59866):** dockerd
   starts with `storage-driver=native` and serves the API;
   `docker version/info/images/inspect` work; `docker pull` downloads and
   extracts images into the native snapshotter; the docker-compose v2
   plugin runs (`docker compose version/ps/up` to the pull stage).
   Runtime fixes: bbolt grows the file before over-EOF mmap and uses
   writable shared mappings; `mount.WithTempMount` treats single bind
   mounts as their source (Haiku has no bind syscall); containerd
   registers the native snapshotter + walking differ for haiku; the
   toolchain gained `UtimesNanoAt` (utimensat). `docker run` progressed through rootfs preparation (symlink-emulated
   bind mounts), OCI spec creation and containerd container creation,
   and now stops at the shim boundary: `containerd-shim-runc-v2` is not
   installed. Stage 3 is a haiku shim (ttrpc TaskService + chroot exec)
   plus persistence-friendly restarts (proper SIGTERM shutdown releases
   the bolt flock; kill -9 orphans hold it). `cmd/dockerd` builds for
   `haiku/amd64` and `haiku/arm64` (`tools/build-haiku.sh`). libnetwork runs
   with the null/host drivers only (bridge package is a stub; no netfilter),
   the resolver's NAT setup is a no-op, osl sandboxes hold no resources, and
   mount/xattr/containerd plumbing reports ENOSYS where Haiku has no
   equivalent. The daemon requires a containerd binary to run; porting
   containerd's content/image services (vfs snapshotter works without kernel
   mounts) is the next step, followed by a QEMU smoke test. Make `cmd/dockerd` build for `haiku/amd64` +
   `haiku/arm64` behind a `haiku` build tag: stub `libnetwork`'s
   iptables/nftables backends with a no-op firewall (Haiku has no netfilter),
   select the `vfs` graphdriver (no overlayfs), and stub the cluster
   executor. Deliverable: a `dockerd` that starts, binds
   `unix:///var/run/docker.sock`, answers `/_ping`, and
   serves `docker version`/`docker info` from the air/OS CLI.
2. **Images without containers.**
   BuildKit (buildx solves) status: the buildx plugin builds and runs, and
   the bind-mount wall is solved (localmounter bind-alias). The remaining
   blocker is below the port: buildx drives builds over a grpc *session*
   to 127.0.0.1, and Haiku's TCP stack rejects parallel localhost dials
   ("only one connection allowed" from addrConn.createTransport), so the
   session healthcheck fails fatally and the build context snapshot is
   garbage-collected mid-solve (gc error lstat snapshots/N). Needs Haiku
   netstack work or a buildkit session transport that reuses one
   connection. The legacy `docker build` builder is unaffected and is the
   supported build path. Wire the registry client (pure Go; needs
   the go-haiku resolver verified against Haiku DNS) and image store on vfs:
   `docker pull`, `docker images`, `docker save/load`.
3. **Container execution on Haiku primitives.** No namespaces exist, so v1
   containers are: `chroot(2)` root filesystems (Haiku is POSIX enough),
   process supervision via teams/wait4, `setrlimit` resource caps, and
   userland networking (loopback TCP proxies per published port). This is a
   documented, honest subset — containers see each other's processes, which
   is called out in `docker info`'s security options. Containers from
   **Linux images** do not use it: they run in the Linux VM (below), with
   Linux's namespaces.
4. **Ecosystem.** build via BuildKit (or a v1 builder fallback), volumes on
   BFS attributes, structured logs, compose plugin.

## Engine fixes (2026-10-30, QEMU x86_64 hrev59866)

- **Pulls:** with no `--platform`, a haiku host matches haiku images first and
  then the linux images of its architecture (`platform.HostMatcher`, used by
  the containerd image service and the graphdriver puller), as containerd's
  default matcher does for freebsd. An explicit `--platform` stays strict.
- **Volumes:** the `local` driver is registered on haiku. Volumes are plain
  directories under the data root (`daemon/volume/local/local_haiku.go`);
  driver options (`type`, `device`, `o`, `size`) are refused, since Haiku's
  `mount.Mount` is a no-op and has no quotas. The haiku shim links them into
  containers (airos-works/containerd).
- **Terminals:** `createSpec` sets `Process.Terminal` from `Config.Tty`, so
  `docker run -t` gets the shim's pty (`docker exec -t` already passed it).
- **Stats:** the system CPU usage is the wall time since the daemon started
  times the CPU count (there is no `/proc/stat`), which is what the CPU
  percentage needs; the shim reports the container's usage.
- **docker top:** Haiku's `ps` has no Linux options or PID column, so
  `ContainerTop` (`daemon/top_haiku.go`) runs it with its defaults and keeps
  the container's teams (the shim reports the whole tree), titled UID, GID,
  PID, THREADS and CMD; `ps` arguments other than the default are refused.
- **Hard links:** layer and archive extraction copy a file when BFS refuses a
  hard link (vendored `containerd/v2/pkg/archive` and `moby/go-archive`).

## Toolchain prerequisites already in place

- gen2brain/go Haiku port (go-haiku) with air/OS additions:
  `syscall.{Ioctl,Tcgetattr,Tcsetattr,Openat,Fchmodat,Mknodat,Fchownat}`
  and termios constants (branch `airos` in the workstation toolchain).
- `x/sys/unix` Haiku shim: vendored in airos-works/docker and extended into
  this repo's vendor tree as stages land.
- `containerd` and `runc` get the same treatment as the CLI once stage 3
  begins; there is no 32-bit-arm Haiku Go target, but all air/OS boards
  (x86_64, rpi4/ARM64, Cubie A7S/ARM64, ROCK 5/ARM64) are covered.

## Linux images run in the Linux VM (2026-10-30)

A container whose image is a Linux one (`ImagePlatform.OS == "linux"`) runs
in air/OS's Linux VM (the `linuxvm` package, airos-works/linuxvm; the
haiku shim, `containerd-linuxvm` and its agent in airos-works/containerd).
The Engine's part (`daemon/oci_linuxvm_haiku.go`):

- **Spec:** the Linux section of the default spec stays (namespaces,
  mounts, masked and read-only paths, devices); Docker's capabilities with
  `--cap-add`/`--cap-drop`/`--privileged` (`caps` knows the VM kernel's
  capabilities on haiku), the resources (`--memory`, `--cpus`,
  `--cpu-shares`, `--cpuset-cpus`, `--pids-limit`; also `docker update`,
  `update_haiku.go`), ulimits, `--read-only`, `--tmpfs`, `--shm-size`,
  sysctls, and PATH, HOSTNAME and TERM as Docker sets them on Linux.
- **Annotations** for what the VM resolves: `works.airos.platform=linux`,
  the image's or the run's USER (resolved against the container's
  `/etc/passwd`; `docker exec -u` passes it the same way), the networks
  with the container's DNS names and aliases on each (Compose service
  names), the published ports (free host ports chosen for `-P` and empty
  host ports, recorded in the container's port map for `docker ps` and
  `docker port`), `--privileged`, and the Engine's data root (the VM caches
  what is under it).
- **docker top** reads the VM's `ps` through `containerd-linuxvm`'s socket
  and keeps the container's processes.
- **Signals:** `moby/sys/signal` has Haiku's signal table (vendored, as in
  airos-works/docker), so `docker kill -s USR1` and images' stop signals
  parse; the shim translates the numbers to Linux's.
- No "platform does not match" warning for Linux images of the host's
  architecture.
