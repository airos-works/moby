//go:build !haiku

package archive

import "os"

func rootLink(root *os.Root, oldname, newname string) error {
	return root.Link(oldname, newname)
}
