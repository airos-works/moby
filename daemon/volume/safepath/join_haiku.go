//go:build haiku

package safepath

import (
	"context"
	"os"

	"github.com/pkg/errors"
)

// Join on Haiku cannot use the Linux bind-mount-of-fd trick (/proc/self/fd
// and MS_BIND do not exist), so v1 validates the joined path and tracks it
// without a backing mount. The path is resolved with O_NOFOLLOW per
// component by the caller-visible semantics of evaluatePath plus an
// explicit lstat check.
func Join(ctx context.Context, path, subpath string) (*SafePath, error) {
	base, subpart, err := evaluatePath(path, subpath)
	if err != nil {
		return nil, err
	}

	joined := base
	if subpart != "" {
		joined = base + "/" + subpart
	}
	if st, err := os.Lstat(joined); err != nil {
		return nil, errors.Wrapf(err, "safe join %q", joined)
	} else if st.Mode()&os.ModeSymlink != 0 {
		return nil, errors.Errorf("safe join %q resolves to a symlink", joined)
	}

	s := &SafePath{
		path:          joined,
		sourceBase:    base,
		sourceSubpath: subpart,
	}
	s.cleanup = func(ctx context.Context) error { return nil }
	return s, nil
}
