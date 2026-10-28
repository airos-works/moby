//go:build haiku

package sockaddr

// Haiku has no routing socket for this; the default route info is
// unavailable until a filesystem/network-kit based implementation lands.
func getDefaultIfName() (string, error) {
	return "", ErrNoInterface
}

func NewRouteInfo() (routeInfo, error) {
	return routeInfo{}, ErrNoRoute
}

func (ri routeInfo) GetDefaultInterfaceName() (string, error) {
	return "", ErrNoInterface
}
