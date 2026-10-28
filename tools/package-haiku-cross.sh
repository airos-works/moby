#!/usr/bin/env bash
# Build the dockerd .hpkg(s) on Linux with the host Haiku `package` tool.
# Usage: tools/package-haiku-cross.sh [arch ...]   (default: amd64 arm64)
set -euo pipefail
ROOT=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
cd "$ROOT"
PACKAGE_TOOL=${PACKAGE_TOOL:-/mnt/HaikuWork/x399/build/x86_64/objects/linux/x86_64/release/tools/package/package}
BASETAG=$(git tag -l 'v*' --sort=-v:refname | head -1)
[ -z "$BASETAG" ] && BASETAG=v0.0.0
AHEAD=$(git rev-list --count "$BASETAG"..HEAD 2>/dev/null || echo 0)
VERSION=$(echo "$BASETAG" | sed 's/^v//')~git.$AHEAD.$(git rev-parse --short HEAD)
STAGE=$(mktemp -d /tmp/dockerd-pkg-XXXXXX)
trap 'rm -rf -- "$STAGE"' EXIT
mkdir -p "$ROOT/artifacts" "$STAGE/bin" "$STAGE/documentation/packages/docker_engine" \
    "$STAGE/boot/home/config/settings/boot/launch"

for arch in "${@:-amd64 arm64}"; do
    haiku_arch=$arch
    [ "$arch" = "amd64" ] && haiku_arch=x86_64
    bin="build/haiku-$arch/dockerd"
    [ -f "$bin" ] || { echo ">> building $arch first"; bash tools/build-haiku.sh "$arch"; }
    rm -rf "$STAGE/bin/dockerd" "$STAGE/.PackageInfo"
    cp "$bin" "$STAGE/bin/dockerd"
    cp docs/airos.md "$STAGE/documentation/packages/docker_engine/"
    cat > "$STAGE/boot/home/config/settings/boot/launch/dockerd" <<'LAUNCHEOF'
#!/bin/sh
# Start the Docker Engine for this user session.
PATH=/boot/system/bin:/boot/home/config/non-packaged/bin:$PATH
exec /boot/system/bin/dockerd --userland-proxy=false
LAUNCHEOF
    cat > "$STAGE/.PackageInfo" <<INFOEOF
name docker_engine
version $VERSION-1
architecture $haiku_arch
summary "Docker Engine (dockerd) for air/OS"
description "The Docker Engine daemon for air/OS, ported from moby. Uses the containerd native snapshotter (no kernel mounts), host/null networking, and serves the standard Docker API on unix:///var/run/docker.sock. Container execution requires the stage-3 runtime."
packager "air/OS contributors"
vendor "air/OS"
copyrights { "2026 air/OS contributors" "Docker, Inc." }
licenses { "Apache v2" }
provides {
    docker_engine = $VERSION
    cmd:dockerd = $VERSION
}
requires {
    haiku >= r1~beta5
    containerd
}
INFOEOF
    out="$ROOT/artifacts/docker_engine-$VERSION-1-${haiku_arch}.hpkg"
    "$PACKAGE_TOOL" create -C "$STAGE" "$out"
    echo ">> $out"
done
