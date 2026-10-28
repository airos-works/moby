//go:build haiku

package supervisor

import (
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/moby/moby/v2/pkg/process"
)

const (
	binaryName    = "containerd"
	sockFile      = "containerd.sock"
	debugSockFile = "containerd-debug.sock"
)

func defaultGRPCAddress(stateDir string) string {
	return filepath.Join(stateDir, sockFile)
}

func defaultDebugAddress(stateDir string) string {
	return filepath.Join(stateDir, debugSockFile)
}

func (r *Daemon) stopDaemon() {
	_ = syscall.Kill(r.daemonPid, syscall.SIGTERM)
	for i := time.Duration(0); i < shutdownTimeout; i += time.Second {
		if !process.Alive(r.daemonPid) {
			break
		}
		time.Sleep(time.Second)
	}
	if process.Alive(r.daemonPid) {
		r.logger.WithField("pid", r.daemonPid).Warn("containerd didn't stop within 15 secs, killing it")
		_ = syscall.Kill(r.daemonPid, syscall.SIGKILL)
	}
}

func (r *Daemon) killDaemon() {
	_ = syscall.Kill(r.daemonPid, syscall.SIGUSR1)
	<-time.After(100 * time.Millisecond)
	_ = process.Kill(r.daemonPid)
}

func (r *Daemon) platformCleanup() {
	_ = os.Remove(r.Address())
}
