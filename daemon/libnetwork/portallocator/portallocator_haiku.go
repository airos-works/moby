//go:build haiku

package portallocator

// Haiku's ephemeral port range follows IANA (49152-65535); the kernel
// does not publish a reservable-range interface, so none are reserved.
func getDynamicPortRange() (start int, end int, _ error) {
	return 49152, 65535, nil
}

func getReservedPorts(begin, end int) (map[uint16]struct{}, error) {
	return nil, nil
}
