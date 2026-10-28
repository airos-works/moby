//go:build haiku

package platform

import (
	"runtime"
	"sync"

	"golang.org/x/sys/unix"
)

// Haiku has no /sys/devices/system/cpu; fall back to the online CPU count.
var possibleCPUs = sync.OnceValue(func() []int {
	n := runtime.NumCPU()
	cpus := make([]int, n)
	for i := range cpus {
		cpus[i] = i
	}
	return cpus
})

// runtimeArchitecture gets the name of the current architecture.
func runtimeArchitecture() (string, error) {
	utsname := &unix.Utsname{}
	if err := unix.Uname(utsname); err != nil {
		return "", err
	}
	machine := make([]byte, len(utsname.Machine))
	for i, c := range utsname.Machine {
		machine[i] = byte(c)
	}
	return unix.ByteSliceToString(machine), nil
}
