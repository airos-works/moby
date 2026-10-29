package platform

import (
	"github.com/containerd/platforms"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// HostMatcher returns the matcher for images that run on host when no
// platform was requested. It is [platforms.Only] of host, except on haiku:
// almost no images are published for haiku, so a haiku host also accepts the
// linux images of its architecture, ranked after its own (as containerd's
// default matcher does for freebsd).
func HostMatcher(host ocispec.Platform) platforms.MatchComparer {
	only := platforms.Only(host)
	if host.OS != "haiku" {
		return only
	}
	linux := host
	linux.OS = "linux"
	linux.OSVersion = ""
	linux.OSFeatures = nil
	return orderedMatcher{only, platforms.Only(linux)}
}

// orderedMatcher matches what any of its matchers match, ranking platforms
// by the first matcher that matches them, then by that matcher's order.
type orderedMatcher []platforms.MatchComparer

func (m orderedMatcher) rank(p ocispec.Platform) int {
	for i, matcher := range m {
		if matcher.Match(p) {
			return i
		}
	}
	return len(m)
}

func (m orderedMatcher) Match(p ocispec.Platform) bool {
	return m.rank(p) < len(m)
}

func (m orderedMatcher) Less(p1, p2 ocispec.Platform) bool {
	r1, r2 := m.rank(p1), m.rank(p2)
	if r1 != r2 {
		return r1 < r2
	}
	if r1 == len(m) {
		return false
	}
	return m[r1].Less(p1, p2)
}
