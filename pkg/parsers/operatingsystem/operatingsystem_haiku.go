//go:build haiku

package operatingsystem

import (
	"errors"

	"golang.org/x/sys/unix"
)

// GetOperatingSystem gets the name of the current operating system.
func GetOperatingSystem() (string, error) {
	utsname := &unix.Utsname{}
	if err := unix.Uname(utsname); err != nil {
		return "", err
	}
	name := make([]byte, len(utsname.Sysname))
	for i, c := range utsname.Sysname {
		name[i] = byte(c)
	}
	return unix.ByteSliceToString(name), nil
}

// GetOperatingSystemVersion gets the version of the current operating system, as a string.
func GetOperatingSystemVersion() (string, error) {
	// there's no standard interface way of getting this, sadly...
	return "", errors.New("Unsupported on generic unix")
}

// IsContainerized returns true if we are running inside a container.
// No-op on Haiku, always returns false.
func IsContainerized() (bool, error) {
	// TODO: Implement jail detection for freeBSD
	return false, errors.New("Cannot detect if we are in container")
}
