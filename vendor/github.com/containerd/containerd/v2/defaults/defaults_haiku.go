//go:build haiku

package defaults

// DefaultRuntime is the containerd runtime name used when none is
// configured; no OCI runtime is wired up on Haiku yet.
const DefaultRuntime = "io.containerd.runc.v2"
