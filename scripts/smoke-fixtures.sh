#!/bin/sh
# Run the implementation's deterministic fixture smoke tests.

set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$ROOT"

if [ ! -f go.mod ]; then
	echo "smoke-fixtures: go.mod is not present; the Go implementation change is required" >&2
	exit 1
fi

GO_TEST_FLAGS=
if [ "$(go env GOOS)" = linux ]; then
	GO_TEST_FLAGS=-tags=sidereon_linux_glibc
fi

TEST_REGEX=${SIDEREON_FIXTURE_TEST_REGEX:-'^Test.*(Fixture|Deterministic|Smoke)'}
if ! go test $GO_TEST_FLAGS ./... -list . 2>/dev/null | grep -Eq "$TEST_REGEX"; then
	echo "smoke-fixtures: no fixture, deterministic, or smoke test matched $TEST_REGEX" >&2
	exit 1
fi

go test $GO_TEST_FLAGS ./... -run "$TEST_REGEX" -count=1
go test $GO_TEST_FLAGS ./... -run "$TEST_REGEX" -count=1
