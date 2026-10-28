//go:build !linux && !darwin && !freebsd && !windows && !haiku

package daemon

func (daemon *Daemon) setupDumpStackTrap(_ string) {
	return
}
