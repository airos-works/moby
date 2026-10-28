//go:build !linux && !windows

package libnetwork

import (
	"context"

	"github.com/moby/moby/v2/daemon/libnetwork/config"
	"github.com/moby/moby/v2/daemon/libnetwork/driverapi"
	"github.com/moby/moby/v2/daemon/libnetwork/drivers/bridge"
	"github.com/moby/moby/v2/daemon/libnetwork/drivers/host"
	"github.com/moby/moby/v2/daemon/libnetwork/drivers/null"
	"github.com/moby/moby/v2/daemon/libnetwork/datastore"
	"github.com/moby/moby/v2/daemon/libnetwork/drvregistry"
)

func registerNetworkDrivers(r driverapi.Registerer, cfg *config.Config, store *datastore.Store, pms *drvregistry.PortMappers) error {
	if err := bridge.Register(r); err != nil {
		return err
	}
	if err := host.Register(r); err != nil {
		return err
	}
	return null.Register(r)
}

// registerPortMappers is a no-op without the bridge driver: there are no
// port-mapping data-plane implementations on unsupported platforms.
func registerPortMappers(ctx context.Context, r *drvregistry.PortMappers, cfg *config.Config) error {
	return nil
}
