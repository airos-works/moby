//go:build haiku

package buildkit

import (
	"context"

	"github.com/containerd/log"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// The builder-next bridge provider has no data plane on Haiku; builders run
// through the stub executor. The symbols below keep the shared executor
// code compiling.
const networkName = "host"

func (iface *lnInterface) Set(s *specs.Spec) error {
	<-iface.ready
	if iface.err != nil {
		log.G(context.TODO()).WithError(iface.err).Error("failed to set networking spec")
		return iface.err
	}
	return nil
}

var _ = context.Background
