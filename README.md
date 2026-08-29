# sidereon-go

`sidereon-go` is the Go binding for Sidereon GNSS and astrodynamics. It uses
cgo to call the committed `libsidereon` static library through the committed,
verbatim `sidereon.h` interface. Go owns the surrounding I/O and transport
work; the C library owns byte parsing and numerical evaluation.

The package name is `sidereon` and the module path is
`github.com/neilberkman/sidereon-go`.

## Pre-release status

This repository is being prepared for the lockstep `1.3.0` release. The public
`v1.3.0` tag does not exist yet, so the Go module is not available from the
module proxy under that version. The current pinned C reference still reports
`1.2.0` in its header; that is the expected local pre-release state and is not
renamed to `1.3.0` here.

Do not tag, publish, or request `@v1.3.0` until the C reference has a public
`v1.3.0` tag with matching `1.3.0` header macros. Until then, a checkout can be
tested with a local replacement or a Git branch/commit reference.

## Install

The binding requires cgo and a C compiler on the host. On macOS, install the
Xcode Command Line Tools. On Linux, install GCC or Clang and the development
headers for the selected libc. On Windows, use the GNU toolchain; the MSVC
ABI is not a supported target for the bundled library.

For a branch or commit that is available from Git, use `go get` with cgo
enabled:

```sh
CGO_ENABLED=1 go get github.com/neilberkman/sidereon-go@main
```

The command above names a Git reference, not an already-published release. In
a local checkout, `go get ./...` and `go test ./...` use the files in that
checkout.

## Supported targets

The distribution includes one static archive for each of these target
families:

| OS | Architectures | libc / ABI |
| --- | --- | --- |
| macOS | `arm64`, `amd64` | Darwin |
| Linux | `amd64`, `arm64` | glibc and musl, separately |
| Windows | `amd64` | GNU |

Linux libc is part of target selection. A glibc program must use the glibc
archive, and a musl program must use the musl archive; the two archives are not
interchangeable. The archive selector uses both `GOOS`/`GOARCH` and the libc
variant. A Linux build must not silently fall back from one libc family to the
other.

Cross-compilation still needs a C compiler that targets the same OS, CPU, and
libc as the Go build. Zig can provide that compiler when it has the required
target support:

```sh
CGO_ENABLED=1 GOOS=linux GOARCH=arm64 \
  CC='zig cc -target aarch64-linux-musl' \
  go build ./...

CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 \
  CC='zig cc -target aarch64-macos' \
  go build ./...

CGO_ENABLED=1 GOOS=windows GOARCH=amd64 \
  CC='zig cc -target x86_64-windows-gnu' \
  go build ./...
```

The archive, compiler target, and any SDK/sysroot must agree. A cross compiler
does not make an archive for a different libc usable.

### Using a system library

The `sidereon_use_system_lib` build tag disables the bundled-archive path. It
is intended for users who provide their own compatible `libsidereon` and
header. Supply the library search and link flags yourself through
`CGO_LDFLAGS`:

```sh
CGO_ENABLED=1 \
  CGO_LDFLAGS='-L./lib -lsidereon' \
  go build -tags sidereon_use_system_lib ./...
```

The system library must expose the same header ABI and version as the Go
binding. The default build does not require a system installation.

## Quickstarts

The Go surface follows the C binding's byte-oriented contracts. Names below
show the package-level API used by the release examples; all input bytes may
come from a file, an HTTP response, or another Go-owned reader.

Parse a TLE and propagate one epoch:

```go
package main

import (
	"fmt"
	"time"

	"github.com/neilberkman/sidereon-go"
)

func main() {
	tle, err := sidereon.ParseTLE(
		"1 25544U 98067A   24001.50000000  .00016717  00000-0  10270-3 0  9009",
		"2 25544  51.6400 208.8657 0002644 250.3037 109.7782 15.49560812999990",
	)
	if err != nil {
		panic(err)
	}
	states, err := tle.Propagate([]time.Time{time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)})
	if err != nil {
		panic(err)
	}
	fmt.Println(states[0].Position)
}
```

Load SP3 bytes and run a single-point positioning solve. The observations and
initial guess use the same units and shape as the public C and Python
examples: pseudoranges in metres, receiver time in seconds, and an ECEF
initial guess in metres.

```go
package main

import (
	"os"

	"github.com/neilberkman/sidereon-go"
)

func main() {
	sp3Bytes, err := os.ReadFile("trimmed.sp3")
	if err != nil {
		panic(err)
	}
	sp3, err := sidereon.LoadSP3(sp3Bytes)
	if err != nil {
		panic(err)
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
		panic(err)
	}
	println(solution.PositionM)
}
```

These examples demonstrate the public operation and units; they do not imply
that every sibling-language convenience name is present in Go. Consult the Go
package documentation for the routes included in the current module.

## Ownership, concurrency, and I/O

Go owns HTTP clients, sockets, TLS, retries, filesystem paths, cache locking,
transport decompression, and presentation adapters. Pass downloaded or read
data to the binding as byte slices. The binding copies data needed for a C
call; the C library does not retain Go pointers after the call returns.

Owning C handles have an explicit, idempotent `Close` method. Call `Close` when
the handle is no longer needed, even though the binding also has a garbage
collection cleanup backstop. Do not use a handle after `Close` or race a live
operation with `Close`.

Read-only operations on a live handle may be shared when the type documents
that behavior. Mutating operations and `Close` are serialized per handle.
Errors are read immediately on the same OS thread as the fallible C call
because the C error detail is thread-local.

## Versioning

Go, C, and the canonical Sidereon engine release in lockstep. A published
`v1.3.0` Go module must use the matching C header macros and static library for
`1.3.0`. A future breaking Go API will use the module suffix
`github.com/neilberkman/sidereon-go/v2`; the current major version has no
suffix.

This module does not claim full parity with the Python, WebAssembly, Elixir, or
C interfaces unless a specific Go route and its tests establish that parity.

## License

The Go module is Apache-2.0; see [LICENSE](LICENSE). The statically linked C
library and its dependencies have additional attributions in
[THIRD-PARTY-NOTICES.md](THIRD-PARTY-NOTICES.md).
