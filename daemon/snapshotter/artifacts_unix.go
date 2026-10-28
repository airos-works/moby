//go:build !windows && !haiku

package snapshotter

import "os"

func removeMountArtifacts(target string) {
	_ = os.Remove(target)
}
