//go:build !windows && !haiku

package daemon

import "os"

// getSystemCPUUsage reads the system's CPU usage from /proc/stat and returns
// the total CPU usage in nanoseconds and the number of CPUs.
func getSystemCPUUsage() (cpuUsage uint64, cpuNum uint32, _ error) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	return readSystemCPUUsage(f)
}
