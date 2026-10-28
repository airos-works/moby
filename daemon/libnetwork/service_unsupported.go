//go:build !linux && !windows

package libnetwork

import "net"

// Service discovery and load balancing have no data plane on unsupported
// platforms (no netfilter / service mesh plumbing): the methods are no-ops
// matching the signatures the agent and endpoint code expect.

func (c *Controller) cleanupServiceDiscovery(cleanupNID string) {}

func (c *Controller) cleanupServiceBindings(nid string) {}

func (c *Controller) addContainerNameResolution(nID, eID, containerName string, taskAliases []string, ip net.IP, method string) error {
	return nil
}

func (c *Controller) delContainerNameResolution(nID, eID, containerName string, taskAliases []string, ip net.IP, method string) error {
	return nil
}

func (c *Controller) addServiceBinding(svcName, svcID, nID, eID, containerName string, vip net.IP, ingressPorts []*PortConfig, serviceAliases, taskAliases []string, ip net.IP, method string) error {
	return nil
}

func (c *Controller) rmServiceBinding(svcName, svcID, nID, eID, containerName string, vip net.IP, ingressPorts []*PortConfig, serviceAliases, taskAliases []string, ip net.IP, method string, deleteSvcRecords bool, fullRemove bool) error {
	return nil
}

func (c *Controller) getLBIndex(sid, nid string, ingressPorts []*PortConfig) int {
	return 0
}

func (sb *Sandbox) populateLoadBalancers(*Endpoint) {}

func arrangeIngressFilterRule() {}
