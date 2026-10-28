//go:build haiku

package osl

// InvokeFunc runs f directly: Haiku sandboxes have no separate network
// namespace to switch into.
func (n *Namespace) InvokeFunc(f func()) error {
	f()
	return nil
}
