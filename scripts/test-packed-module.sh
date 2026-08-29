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
	"fmt"
	"os"

	"github.com/neilberkman/sidereon-go"
)

// This is the only implementation-dependent part of the packed consumer.
// It deliberately performs a numerical SPP solve over in-memory SP3 bytes.
func runSolve() error {
	sp3Bytes, err := os.ReadFile("trimmed.sp3")
	if err != nil {
		return err
	}
	sp3, err := sidereon.LoadSP3(sp3Bytes)
	if err != nil {
		return err
	}
	defer sp3.Close()

	solution, err := sidereon.SolveSPP(sp3, sidereon.SPPConfig{
		Observations: []sidereon.SPPObservation{
			{SatelliteID: "G08", PseudorangeM: 23825519.8},
			{SatelliteID: "G10", PseudorangeM: 22717690.1},
			{SatelliteID: "G16", PseudorangeM: 20478653.4},
			{SatelliteID: "G18", PseudorangeM: 21768335.2},
			{SatelliteID: "G20", PseudorangeM: 21248327.7},
			{SatelliteID: "G21", PseudorangeM: 20808709.8},
		},
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
	fmt.Printf("numerical solve: %v\n", solution)
	return nil
}

func main() {
	if err := runSolve(); err != nil {
		panic(err)
	}
}
EOF

# The public C example's one-epoch slice is enough for this SPP smoke. The
# parser receives exactly the bytes that a Go-owned file or HTTP reader would
# provide.
cat > "$CONSUMER/trimmed.sp3" <<'EOF'
#cP2020  6 24 12  0  0.00000000      19 TEST IGb14 FIT TEST
## 2111 267300.00000000   900.00000000 59024 0.0000000000000
+    6   G08G10G16G18G20G21  0  0  0  0  0  0  0  0  0  0
+          0  0  0  0  0  0  0  0  0  0  0  0  0  0  0  0
+          0  0  0  0  0  0  0  0  0  0  0  0  0  0  0
+          0  0  0  0  0  0  0  0  0  0  0  0  0  0  0
+          0  0  0  0  0  0  0  0  0  0  0  0  0  0  0
++         4  4  4  4  4  4  0  0  0  0  0  0  0  0  0  0
++         0  0  0  0  0  0  0  0  0  0  0  0  0  0  0  0
++         0  0  0  0  0  0  0  0  0  0  0  0  0  0  0
++         0  0  0  0  0  0  0  0  0  0  0  0  0  0  0
++         0  0  0  0  0  0  0  0  0  0  0  0  0  0  0
%c M  cc GPS ccc cccc cccc cccc ccccc ccccc ccccc
%c cc cc ccc ccc cccc cccc cccc cccc ccccc ccccc ccccc
%f  0.0000000  0.000000000  0.00000000000  0.000000000000000
%f  0.0000000  0.000000000  0.00000000000  0.000000000000000
%i    0    0    0    0      0      0      0      0         0
%i    0    0    0    0      0      0      0      0         0
/* C static-library consumer smoke
*  2020  6 24 12  0  0.00000000
PG08   7438.042916 -20762.704119  14621.192800    -38.647002
PG10  23919.560682  11717.183156   1816.071269   -380.567404
PG16  18784.088401  -3819.088286  18359.430195   -174.389524
PG18   6670.210753  13729.937789  21723.310519    228.892428
PG20  17932.623680  14929.762015  12814.016152    527.447354
PG21  16999.913180   4708.560403  20550.418405     15.546299
EOF

GOTOOLCHAIN=local GOPROXY=off CGO_ENABLED=1 GOFLAGS="$PACKED_GOFLAGS" \
	go -C "$CONSUMER" run .
