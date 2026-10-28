//go:build haiku

package command

import "github.com/moby/moby/v2/daemon/config"

// preNotifyReady sends a message to the host when the API is active; no
// notification mechanism exists on Haiku.
func preNotifyReady() error { return nil }

// notifyReady signals the daemon is ready to serve requests.
func notifyReady() {}

// notifyReloading signals a configuration reload.
func notifyReloading() func() { return func() {} }

// notifyStopping signals the daemon is shutting down.
func notifyStopping() {}

func validateCPURealtimeOptions(_ *config.Config) error { return nil }

func setPlatformOptions(conf *config.Config) error { return nil }
