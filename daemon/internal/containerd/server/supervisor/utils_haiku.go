//go:build haiku

package supervisor

import "syscall"

// containerdSysProcAttr returns the SysProcAttr to use when exec'ing
// containerd. Haiku has no Pdeathsig.
func containerdSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{}
}
