package daemon

import (
	"runtime"
	"time"
)

// statsEpoch is the origin of the system CPU usage reported on Haiku.
var statsEpoch = time.Now()

// getSystemCPUUsage returns the system's CPU time in nanoseconds and the
// number of CPUs. Haiku has no /proc/stat; clients only use the change
// between two samples (container CPU time over system CPU time), so the
// wall time since the daemon started times the CPU count serves, as that is
// what /proc/stat's busy and idle times add up to.
func getSystemCPUUsage() (cpuUsage uint64, cpuNum uint32, _ error) {
	n := runtime.NumCPU()
	return uint64(time.Since(statsEpoch).Nanoseconds()) * uint64(n), uint32(n), nil
}
