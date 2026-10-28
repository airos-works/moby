//go:build haiku

package mount

import "github.com/containerd/errdefs"

// Mount is not implemented on Haiku yet; the Engine's mount backend lands
// with the graphdriver work.
func (m *Mount) mount(target string) error {
	return errdefs.ErrNotImplemented
}
