//go:build haiku

// Package bridge on air/OS provides a compatibility network driver named
// "bridge" so that default compose files and `docker network create` work
// out of the box. Haiku has no network namespaces or bridges: containers
// attached to this driver share the host network stack, so every service a
// container listens on is reachable directly on the host's addresses. Port
// publishing options are accepted and ignored.
package bridge

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/moby/moby/v2/daemon/libnetwork/driverapi"
	"github.com/moby/moby/v2/daemon/libnetwork/scope"
)

const (
	// NetworkType is the driver name.
	NetworkType = "bridge"

	// DefaultBridgeName is the default bridge network name.
	DefaultBridgeName = "docker0"
)

// Configuration is the driver-level configuration, accepted for
// compatibility and ignored.
type Configuration struct {
	EnableIPForwarding       bool
	DisableFilterForwardDrop bool
	EnableIPTables           bool
	EnableIP6Tables          bool
	EnableProxy              bool
	ProxyPath                string
	AllowDirectRouting       bool
	AcceptFwMark             string
}

// ValidateFixedCIDRV6 checks the fixed IPv6 CIDR is syntactically valid.
func ValidateFixedCIDRV6(val string) error {
	if val == "" {
		return nil
	}
	_, _, err := net.ParseCIDR(val)
	if err != nil {
		return fmt.Errorf("invalid fixed-cidr-v6 %q: %w", val, err)
	}
	return nil
}

// LegacyContainerLinkOptions returns the bridge label for legacy --link
// endpoints; links are unsupported on Haiku, so the labels are empty.
func LegacyContainerLinkOptions(parentEndpoints, childEndpoints []string) map[string]any {
	return nil
}

type driver struct {
	mu       sync.Mutex
	networks map[string]bool
}

// Register registers the bridge driver with the libnetwork registry.
func Register(r driverapi.Registerer) error {
	return r.RegisterDriver(NetworkType, &driver{networks: make(map[string]bool)}, driverapi.Capability{
		DataScope:         scope.Local,
		ConnectivityScope: scope.Local,
	})
}

func (d *driver) CreateNetwork(ctx context.Context, id string, option map[string]any, nInfo driverapi.NetworkInfo, ipV4Data, ipV6Data []driverapi.IPAMData) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.networks[id] = true
	return nil
}

func (d *driver) DeleteNetwork(nid string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.networks, nid)
	return nil
}

func (d *driver) CreateEndpoint(_ context.Context, nid, eid string, ifInfo driverapi.InterfaceInfo, epOptions map[string]any) error {
	return nil
}

func (d *driver) DeleteEndpoint(nid, eid string) error {
	return nil
}

func (d *driver) EndpointOperInfo(nid, eid string) (map[string]any, error) {
	return map[string]any{}, nil
}

func (d *driver) Join(_ context.Context, nid, eid string, sboxKey string, jinfo driverapi.JoinInfo, _, _ map[string]any) error {
	return nil
}

func (d *driver) Leave(nid, eid string) error {
	return nil
}

func (d *driver) ProgramExternalConnectivity(nid, eid string, options map[string]any) error {
	return nil
}

func (d *driver) RevokeExternalConnectivity(nid, eid string) error {
	return nil
}

func (d *driver) IsBuiltIn() bool {
	return true
}

var _ driverapi.Driver = (*driver)(nil)
