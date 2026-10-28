//go:build haiku

package daemon

import (
	"errors"
	"reflect"

	containertypes "github.com/moby/moby/api/types/container"
	libcontainerdtypes "github.com/moby/moby/v2/daemon/internal/libcontainerd/types"
)

// toContainerdResources converts container resources to the containerd
// shape; resource limits are enforced with setrlimit in the Haiku runtime,
// so this currently carries no Linux cgroup data.
func toContainerdResources(resources containertypes.Resources) (*libcontainerdtypes.Resources, error) {
	if reflect.ValueOf(resources).IsZero() {
		return nil, nil
	}
	return nil, errors.New("updating container resources is not supported on haiku yet")
}
