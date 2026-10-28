//go:build haiku

package libnetwork

import "context"

// Haiku has no netfilter; the embedded resolver runs without NAT rules
// until the port of libnetwork's firewall backends lands.
func (r *Resolver) setupNAT(ctx context.Context) error {
	return nil
}
