package image

import (
	"errors"
	"runtime"
	"strings"

	"github.com/moby/moby/v2/errdefs"
)

// CheckOS checks if the given OS matches the host's platform, and
// returns an error otherwise. On air/OS (haiku), Linux images are supported
// too: they run in the Linux VM.
func CheckOS(os string) error {
	if runtime.GOOS == "haiku" && strings.EqualFold(os, "linux") {
		return nil
	}
	if !strings.EqualFold(runtime.GOOS, os) {
		return errdefs.InvalidParameter(errors.New("operating system is not supported"))
	}
	return nil
}
