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

cat > "$CONSUMER/main.go" <<'EOF'
package main

import (
	"errors"
	"fmt"
	"math"
	"os"

	"github.com/neilberkman/sidereon-go"
)

// This is the only implementation-dependent part of the packed consumer.
// It performs a numerical SPP solve over in-memory bytes from the archived
// module fixture.
func runSolve() (err error) {
	sp3Bytes, err := os.ReadFile("trimmed.sp3")
	if err != nil {
		return err
	}
	sp3, err := sidereon.LoadSP3(sp3Bytes)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := sp3.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close SP3: %w", closeErr))
		}
	}()

	observations := []struct {
		id   string
		bits uint64
	}{
		{"G08", 0x4176b8c6fd82e861},
		{"G10", 0x4175aa4fa1a0c21f},
		{"G16", 0x417387abd6052c3b},
		{"G18", 0x4174c288f3bd1166},
		{"G20", 0x417443947bd00bd6},
		{"G21", 0x4173d8405cd09f84},
		{"G26", 0x417425d51967e798},
		{"G27", 0x41745a4b78a81707},
	}
	input := make([]sidereon.SPPObservation, len(observations))
	for i, observation := range observations {
		input[i] = sidereon.SPPObservation{
			SatelliteID:  observation.id,
			PseudorangeM: math.Float64frombits(observation.bits),
		}
	}
	solution, err := sidereon.SolveSPP(sp3, sidereon.SPPConfig{
		Observations:    input,
		TRxJ2000S:       646272000.0,
		TRxSecondOfDayS: 43200.0,
		DayOfYear:       176.5,
		InitialGuess:    [4]float64{4.5e6, 0.5e6, 4.5e6, 0},
		Ionosphere:      false,
		Troposphere:     false,
		WithGeodetic:    true,
	})
	if err != nil {
		return err
	}
	wantPosition := [3]uint64{0x41511b07ff83c7f1, 0x4120cd6b5ee8cafe, 0x41511e62229db724}
	for axis, bits := range wantPosition {
		if math.Float64bits(solution.PositionM[axis]) != bits {
			return fmt.Errorf("position[%d] = %.17g, want frozen C value", axis, solution.PositionM[axis])
		}
	}
	if math.Float64bits(solution.ReceiverClockS) != 0x3f1a3b88360a8d78 {
		return fmt.Errorf("receiver clock = %.17g, want frozen C value", solution.ReceiverClockS)
	}
	if solution.UsedSatelliteCount != len(observations) {
		return fmt.Errorf("used satellite count = %d, want %d", solution.UsedSatelliteCount, len(observations))
	}
	fmt.Printf("numerical solve: %v\n", solution)
	return nil
}

func main() {
	if err := runSolve(); err != nil {
		panic(err)
	}
}
EOF

cp "$PACKED/testdata/trimmed.sp3" "$CONSUMER/trimmed.sp3"

GOTOOLCHAIN=local GOPROXY=off CGO_ENABLED=1 GOFLAGS="$PACKED_GOFLAGS" \
	go -C "$CONSUMER" run .
