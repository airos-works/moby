//go:build haiku

package osl

import "time"

// Address advertisement knobs, matching the Linux defaults; unused on Haiku
// (no ARP/NDISC announcement plumbing in the stub sandbox).
const (
	AdvertiseAddrNMsgsMin      = 0
	AdvertiseAddrNMsgsMax      = 3
	AdvertiseAddrIntervalMin   = 100 * time.Millisecond
	AdvertiseAddrIntervalMax   = 2 * time.Second
)
