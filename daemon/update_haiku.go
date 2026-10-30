//go:build haiku

package daemon

import (
	"reflect"

	containertypes "github.com/moby/moby/api/types/container"
	libcontainerdtypes "github.com/moby/moby/v2/daemon/internal/libcontainerd/types"
)

// toContainerdResources converts container resources to the containerd
// shape: the cgroup limits of a container in the Linux VM. A Haiku
// container's shim ignores them.
func toContainerdResources(resources containertypes.Resources) (*libcontainerdtypes.Resources, error) {
	if reflect.ValueOf(resources).IsZero() {
		return nil, nil
	}
	return linuxVMResources(resources), nil
}
