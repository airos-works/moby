//go:build haiku

package config

// PlatformConfig defines platform-specific configuration.
type PlatformConfig struct {
	// NetnsDir is kept for API parity; Haiku has no network namespaces.
	NetnsDir string
}
