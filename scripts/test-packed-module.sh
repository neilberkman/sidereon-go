#!/bin/sh
# Test the packed, tracked module from an outside consumer module.
#
# Integration seam: the generated program below assumes these root exports:
#   sidereon.LoadSP3([]byte) and sidereon.SolveSPP(*sidereon.SP3, sidereon.SPPConfig)
# Update only the generated program if the implementation chooses different
# names. Keep the input as bytes and retain a real numerical solve.

set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$ROOT"

if [ ! -f go.mod ]; then
	echo "test-packed-module: go.mod is not present; the Go implementation change is required" >&2
	exit 1
fi

TEMP_PARENT=${TMPDIR:-/tmp}
TEMP_ROOT=$(mktemp -d "$TEMP_PARENT/sidereon-go-packed.XXXXXX")
trap 'rm -rf "$TEMP_ROOT"' EXIT HUP INT TERM

PACKED=$TEMP_ROOT/module
CONSUMER=$TEMP_ROOT/consumer
mkdir -p "$PACKED" "$CONSUMER"

PACKED_GOFLAGS=${GOFLAGS:-}
if [ "$(go env GOOS)" = linux ]; then
	PACKED_GOFLAGS="${PACKED_GOFLAGS:+$PACKED_GOFLAGS }-tags=sidereon_linux_glibc"
fi

# git archive contains only committed/tracked module content and has no .git
# directory or untracked build output.
git archive --format=tar HEAD | tar -xf - -C "$PACKED"

go -C "$CONSUMER" mod init example.com/sidereon-consumer
go -C "$CONSUMER" mod edit -replace github.com/neilberkman/sidereon-go="$PACKED"
GOTOOLCHAIN=local GOPROXY=off CGO_ENABLED=1 GOFLAGS="$PACKED_GOFLAGS" \
	go -C "$CONSUMER" get github.com/neilberkman/sidereon-go@v0.0.0

cp "$PACKED/scripts/testdata/packed-consumer/main.go" "$CONSUMER/main.go"
cp "$PACKED/testdata/trimmed.sp3" "$CONSUMER/trimmed.sp3"

GOTOOLCHAIN=local GOPROXY=off CGO_ENABLED=1 GOFLAGS="$PACKED_GOFLAGS" \
	go -C "$CONSUMER" run .
