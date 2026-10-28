//go:build haiku

package snapshotter

import (
	"os"
)

// removeMountArtifacts removes the mount point and its backing image.
func removeMountArtifacts(target string) {
	if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
		// The mount point directory itself.
		_ = err
	}
	_ = os.Remove(imagePathFor(target))
}
