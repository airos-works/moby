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
1. **Compile skeleton.** **DONE (2026-10-28).** `cmd/dockerd` builds for
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
   `unix:///boot/system/cache/docker/docker.sock`, answers `/_ping`, and
   serves `docker version`/`docker info` from the air/OS CLI.
2. **Images without containers.** Wire the registry client (pure Go; needs
   the go-haiku resolver verified against Haiku DNS) and image store on vfs:
   `docker pull`, `docker images`, `docker save/load`.
3. **Container execution on Haiku primitives.** No namespaces exist, so v1
   containers are: `chroot(2)` root filesystems (Haiku is POSIX enough),
   process supervision via teams/wait4, `setrlimit` resource caps, and
   userland networking (loopback TCP proxies per published port). This is a
   documented, honest subset — containers see each other's processes, which
   is called out in `docker info`'s security options.
4. **Ecosystem.** build via BuildKit (or a v1 builder fallback), volumes on
   BFS attributes, structured logs, compose plugin.

## Toolchain prerequisites already in place

- gen2brain/go Haiku port (go-haiku) with air/OS additions:
  `syscall.{Ioctl,Tcgetattr,Tcsetattr,Openat,Fchmodat,Mknodat,Fchownat}`
  and termios constants (branch `airos` in the workstation toolchain).
- `x/sys/unix` Haiku shim: vendored in airos-works/docker and extended into
  this repo's vendor tree as stages land.
- `containerd` and `runc` get the same treatment as the CLI once stage 3
  begins; there is no 32-bit-arm Haiku Go target, but all air/OS boards
  (x86_64, rpi4/ARM64, Cubie A7S/ARM64, ROCK 5/ARM64) are covered.
