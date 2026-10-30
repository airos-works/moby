//go:build !linux && !haiku

package caps

func initCaps() {
	// no capabilities on Windows
}
