//go:build haiku

package git

import "syscall"

// No namespace or credential-daemon plumbing on Haiku for reexec.
var reexecSysProcAttr = syscall.SysProcAttr{}
