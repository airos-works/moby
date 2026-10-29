package platform

import (
	"testing"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"gotest.tools/v3/assert"
)

func TestHostMatcher(t *testing.T) {
	haiku := ocispec.Platform{OS: "haiku", Architecture: "amd64", Variant: "v3"}
	linux := ocispec.Platform{OS: "linux", Architecture: "amd64"}
	linuxV2 := ocispec.Platform{OS: "linux", Architecture: "amd64", Variant: "v2"}
	linuxArm := ocispec.Platform{OS: "linux", Architecture: "arm64"}
	windows := ocispec.Platform{OS: "windows", Architecture: "amd64"}

	m := HostMatcher(haiku)
	assert.Check(t, m.Match(haiku))
	assert.Check(t, m.Match(ocispec.Platform{OS: "haiku", Architecture: "amd64"}))
	assert.Check(t, m.Match(linux))
	assert.Check(t, m.Match(linuxV2))
	assert.Check(t, !m.Match(linuxArm))
	assert.Check(t, !m.Match(windows))

	// The host's own OS ranks first, then linux by the usual variant order.
	assert.Check(t, m.Less(haiku, linux))
	assert.Check(t, !m.Less(linux, haiku))
	assert.Check(t, m.Less(linuxV2, linux))
	assert.Check(t, m.Less(linux, windows))

	// Other hosts keep platforms.Only.
	l := HostMatcher(linux)
	assert.Check(t, l.Match(linux))
	assert.Check(t, !l.Match(haiku))
}
