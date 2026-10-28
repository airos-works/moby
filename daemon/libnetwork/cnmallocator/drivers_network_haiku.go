//go:build haiku

package cnmallocator

import (
	"github.com/moby/moby/v2/daemon/libnetwork/driverapi"
	"github.com/moby/moby/v2/daemon/libnetwork/drivers/host"
	"github.com/moby/moby/v2/daemon/libnetwork/drivers/null"
	"github.com/moby/swarmkit/v2/manager/allocator/networkallocator"
)

// Haiku has no overlay/bridge data-plane support (no netfilter, no
// namespaces), so only the host and null drivers are registered.
var globalDrivers = map[string]func(driverapi.Registerer) error{}

var localDrivers = []string{
	host.NetworkType,
	null.NetworkType,
}

// PredefinedNetworks returns the list of predefined network structures.
func (*Provider) PredefinedNetworks() []networkallocator.PredefinedNetworkData {
	return []networkallocator.PredefinedNetworkData{
		{Name: "host", Driver: "host"},
		{Name: "null", Driver: "null"},
	}
}
