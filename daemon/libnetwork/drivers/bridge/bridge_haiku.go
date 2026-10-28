//go:build haiku

// Package bridge on Haiku is a compile-time stub: the bridge driver needs
// netfilter and network namespaces, which Haiku does not have. The driver
// package exposes the small external surface the rest of the daemon
// references so the Engine can build and run with host/null networking.
package bridge

import (
	"fmt"
	"net"
)

const (
	// NetworkType is the driver name.
	NetworkType = "bridge"

	// DefaultBridgeName is the default bridge device name.
	DefaultBridgeName = "docker0"
)

// Configuration is the driver-level configuration. On Haiku every option
// is accepted for compatibility but has no effect.
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
