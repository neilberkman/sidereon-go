# sidereon-go

`sidereon-go` is the Go binding for Sidereon GNSS and astrodynamics. It uses
cgo to call the committed `libsidereon` static library through the committed,
verbatim `sidereon.h` interface. Go owns the surrounding I/O and transport
work; the C library owns byte parsing and numerical evaluation.

The package name is `sidereon` and the module path is
`sidereon.dev/go/v4`.

OMM KVN and XML writers require complete CCSDS messages. Use
`OMM.ToKVNLenient` or `OMM.ToXMLLenient` to write partial messages.
The catalog API provides CelesTrak GP and NAVCEN source metadata and URL
builders. `ParseCelestrakGPResponse` parses supplied response bytes without
performing a network request.
Record TLE handles support AFSPC or improved operations mode and strict or
lenient checksum policy through `TLEWithOpsModeAndPolicy`.
Detached record, rejection and error details can be compared with
`reflect.DeepEqual`; catalog and record conversion errors also provide
`Equal`. Response and record references use native equality methods.

## Install

The binding requires cgo and a C compiler on the host. On macOS, install the
Xcode Command Line Tools. On Linux, install GCC or Clang and the development
headers for the selected libc. On Windows, use the GNU toolchain; the MSVC
ABI is not a supported target for the bundled library.

Install a released version with `go get` and cgo enabled (a branch or commit
reference such as `@main` works the same way):

```sh
CGO_ENABLED=1 go get sidereon.dev/go/v4@v4.2.0
```

On Linux, select the matching bundled libc explicitly. For the usual glibc
distribution:

```sh
CGO_ENABLED=1 GOFLAGS='-tags=sidereon_linux_glibc' \
  go get sidereon.dev/go/v4@v4.2.0
```

Use `sidereon_linux_musl` instead when building with a musl C toolchain.

In a local checkout, `go get ./...` and `go test ./...` use the files in that
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
  GOFLAGS='-tags=sidereon_linux_musl' \
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
is intended for users who provide their own compatible `libsidereon`. Supply
the library search and link flags yourself through `CGO_LDFLAGS`:

```sh
CGO_ENABLED=1 \
  CGO_LDFLAGS='-L./lib -lsidereon' \
  go build -tags sidereon_use_system_lib ./...
```

The system library must expose the same header ABI and version as the Go
binding. The default build does not require a system installation.

### Native archive rebuilding

The native static archives for version 3 require Rust 1.98.1 (pinned in `rust-toolchain.toml`). When building from a source checkout without pre-compiled archives, rebuild them using:

```sh
./scripts/build-native-archives.sh
```

Alternatively, link against an external compatible `libsidereon` using `-tags sidereon_use_system_lib`.

## Quickstarts

### Precise positioning

The PPP API supports float, fixed, single-epoch, and strategy-selected
auto-initialized solves. Additive configuration fields enable residual-ionosphere
estimation without changing existing call sites. `BuildPPPLookup` creates owned,
indexed correction tables, and fixed solutions can clone owned ambiguity-search
metadata. Float and fixed solutions expose posterior variance and residual
ionosphere; float solutions also expose weighted residuals and decorrelation time.
The low-level range-rate API predicts pseudorange rate and solves receiver
velocity and clock drift, with optional robust reweighting.

SSR bias records, application reports, aggregate status, phase-continuity
tokens, `SsrPppBiasOptions`, and `with_ssr_biases` are not yet exposed by the C
library and therefore are not available in this Go binding.

The Go surface follows the C binding's byte-oriented contracts. Names below
show the package-level API used by the release examples; all input bytes may
come from a file, an HTTP response, or another Go-owned reader.

### CCSDS TDM

`ParseTDMKVN` and `ParseTDMKVNWithPolicy` read CCSDS Tracking Data Messages.
`NewTDM` constructs a message from raw header, metadata, comment, and record
values. A `TDM` retains lossless KVN text, exposes participants, paths,
segments, and records, and supports atomic header, metadata, and data edits.
`TDM.Equal` delegates complete message value equality to the engine. Plain TDM
value structs use Go's `==` where comparable and provide `Equal` otherwise.
Strict operations reject non-conforming messages; explicit read and write
policies return typed warnings and departures when a rule is forgiven.
Edits use `TDMHeaderValues`, `TDMCommentInput`, `TDMFieldInput`, and
`TDMRecordInput`; read methods return the corresponding lossless output types.
Optional header text uses `TDMStringField`, for example
`TDMStringField{Present: true, Value: "2026-001T00:00:00"}`.

```go
tdm, warnings, err := sidereon.ParseTDMKVNWithPolicy(data, sidereon.LenientTDMPolicy())
if err != nil {
	return err
}
defer tdm.Close()

kvn, departures, err := tdm.ToKVNWithPolicy(sidereon.StrictTDMWritePolicy())
```

Parse a TLE and propagate one epoch:

```go
package main

import (
	"fmt"
	"time"

	"sidereon.dev/go/v4"
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

	"sidereon.dev/go/v4"
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

### SP3 interpolation options and continuity checks

When loading SP3 products or evaluating continuity, interpolation policy can be
configured using functional options like `WithGapThresholdFactor` or explicit
options structs:

```go
// Load an SP3 product with a custom gap threshold factor (e.g. 13.0)
sp3, err := sidereon.LoadSP3(sp3Bytes, sidereon.WithGapThresholdFactor(13.0))
if err != nil {
	panic(err)
}
defer sp3.Close()

// Inspect the configured factor
factor, err := sp3.GapThresholdFactor()

// Evaluate physical continuity and holdout defects with custom options
continuity, err := sp3.ContinuityWithOptions(1, 1.0, sidereon.SP3ContinuityOptions{
	GapThresholdFactor: 13.0,
})

// Or obtain the continuity verdict JSON
verdictJSON, err := sp3.ContinuityVerdictJSON(1, 1.0, fromJ2000S, throughJ2000S, sidereon.WithGapThresholdFactor(13.0))
```

To opt into record-column quantization while writing, call
`SP3.ToSP3StringWithRounding()`. It writes a copy, returns the maximum change
per position, velocity, clock, and clock-rate component in SI units, and keeps
a typed writer refusal in `SP3WriteRefusalError` when the product still cannot
be written. The refusal error retains its `SP3WriteOutcome`; the source product
remains unchanged.

Merging or aligning inputs already on SP3 record columns returns records on
those columns. Merge agreement metrics describe contributor dispersion before
that final rounding; higher-precision inputs retain their precision.

These examples demonstrate the public operation and units; they do not imply
that every sibling-language convenience name is present in Go. Consult the Go
package documentation for the routes included in the current module.

## Orbital dynamics

`PropagateState` and the covariance and orbit-fit operations use elapsed
terrestrial time for integration. Cartesian propagation epochs remain
absolute TDB seconds past J2000. Drag uses UT1 Earth orientation and UTC
for the atmospheric calendar and space-weather lookup.

`PropagateStateWithForceModel(config, epochs, environment, additionalForces)`
samples a GCRF ephemeris. `PropagateStateWithDenseOutput(config, target, environment, additionalForces)`
returns the final GCRF state and continuous DP5(4) output when available.
Both use the built-in force model and optional drag in `PropagationConfig`.
Set `PropagationEnvironment.SpaceWeather` for table-driven drag, or use
the fixed weather in `DragParameters`. The engine adds the `[]Force` slice
in order after the configured forces and drag. Pass nil for no additional
forces. Use `DenseOutput.Coverage` to read the inclusive TDB interval.
Close the returned ephemeris or dense output when done.

Use `Equal` to compare `PropagationConfig`, `OrbitFitOptions`,
`OrbitParameterFitOptions`, and `TCAForceOptions`. It compares enabled
force components and ignores the payloads of disabled components.

`NewCompositeForceModel` combines concrete force handles in evaluation order.
Force handles provide parameter names, replacement values and derivatives.
Use `PropagateVariational` for state-transition matrices and force-parameter
sensitivities. `FitInertialOrbitParameters` and `FitSP3ECEFOrbitParameters`
estimate selected force parameters together with the initial state. The fit
options take the same additional force list and an optional gravity tide
convention, which defaults to tide-free. `ForceParameters` returns estimates
in fit order. Compare `EstimatedForceParameter` values with `==` and use
`OrbitParameterFit.Sigma(name)` to read a formal standard deviation from the
scaled covariance.

Solar radiation pressure uses the apparent solar disc, an oblate WGS84
Earth and lunar occultation by default. To override it in a built-in force
configuration, set `SolarRadiationPressure.HasShadow` and supply `Shadow`.
Set `ForceModelComponents.HasEarthRadiationPressure` and supply
`EarthRadiationPressure` to enable albedo and infrared pressure. Both
configurations provide `Equal` methods that compare enabled settings. The
radiation and third-body forces use GCRS Sun and Moon positions. Supply
a `DEEphemeris` through `PropagationContext` to use
a JPL planetary ephemeris. An `IERS2010EarthOrientationProvider` supplies
Earth orientation from daily observations.

`ParseICGEMGravityField` reads static and time-variable gravity
coefficients and their uncertainty columns. The geopotential uses
Holmes–Featherstone recursion for embedded EGM96 and ICGEM fields.
Eckstein–Hechler and Brouwer–Lyddane propagators convert between mean and
osculating elements and evaluate analytic ephemerides.

## Orbit scoring and GNSS attitude

`ScoreOrbit` compares predicted SP3, broadcast or sampled trajectories
with a reference SP3 product. Reports include radial, transverse and
normal errors, clock and SISRE statistics, prediction-horizon bins and
coverage counts by satellite and constellation. Scoring sources retain their
backing products. Broadcast sources keep the navigation-message preference
active at construction. `Equal` compares scoring options, frame reconciliation
options, epoch selections and trajectory samples using engine value semantics.
Use `Equal` on optional orbit scalars, source states, score summaries and
attitude values to compare only the payloads that are present.

`GNSSBlock` identifies satellite blocks from IGS names or constellation
metadata. `GNSSAttitudeLawForBlock` selects the block's yaw law. Use
`ComputeNominalYaw` and `ComputeDYBAxes` for geocentric position and
velocity data. The eclipse functions provide lunar and combined
Earth–Moon shadow fractions, classifications and position gradients.

`GnssAttitude` evaluates a yaw law along an `OrbitAttitudeTrajectory` at TDB
seconds past J2000. Its body and DYB axes are in GCRF. `Sample` returns GCRF
satellite and geocentric Sun positions in metres and satellite velocity in
metres per second.

Create a trajectory from SP3, broadcast navigation, an orbit scoring source,
a GCRF Eckstein–Hechler or Brouwer–Lyddane propagator, or continuous DP5(4)
output from `PropagateStateWithDenseOutput`. Earth-fixed sources must use
positions aligned with the orientation provider's ITRF. Pass nil for the
default TDB provider, or supply an `IERS2010EarthOrientationProvider`. Set the
velocity difference half step in source time-scale seconds with
`WithVelocityHalfStepS`. The Sun comes from the analytic ephemeris by default
and can be set to a JPL DE ephemeris with `WithDEEphemeris`.

Close the trajectory when done. It retains the source data after the
original handles close. Requests beyond orbit, Earth orientation or Sun
coverage return a typed `GNSSAttitudeError` with kind
`trajectory_unavailable`. Some laws also sample around the requested epoch.

## Ownership, concurrency, and I/O

Go owns HTTP clients, sockets, TLS, retries, filesystem paths, cache locking,
transport decompression, and presentation adapters. Pass downloaded or read
data to the binding as byte slices. The binding copies data needed for a C
call; the C library does not retain Go pointers after the call returns.

Owning C handles have an explicit, idempotent `Close` method. Call `Close` when
the handle is no longer needed, even though the binding also has a garbage
collection cleanup backstop. Where a handle documents concurrent use with
`Close`, an operation can finish with its retained data or return `ErrClosed`
if the handle closes first.

Read-only operations on a live handle may be shared when the type documents
that behavior. Mutating operations and `Close` are serialized per handle.
Errors are read immediately on the same OS thread as the fallible C call
because the C error detail is thread-local.

### Error handling

Fallible operations using the general C status contract return a `*StatusError` on native failure; operation-specific failures may return dedicated types such as `*SelectionError` or `*FDEUnresolvedError`. Use `errors.As` to inspect supported error types:

```go
var statusErr *sidereon.StatusError
if errors.As(err, &statusErr) {
	// Status code, stable status name, and same-call diagnostic detail:
	fmt.Printf("status %s (%d): %s\n", statusErr.Text, statusErr.Code, statusErr.Detail)

	// Access Schema 1 generic engine error details when present:
	if engErr := statusErr.Engine; engErr != nil {
		fmt.Printf("engine error family=%s (%d) op=%s kind=%s\n",
			engErr.FamilyName, engErr.Family, engErr.Operation, engErr.Kind)
		// Access raw payload bytes or unmarshal structured fields:
		// engErr.UnmarshalFields(&customFields)
	}
}
```

- **Lossless payload & partial capture**: `EngineError` preserves raw JSON fields (`Fields`) and complete payload bytes (`Payload`) without truncating future variants or rounding integer lexemes. If diagnostic capture or payload decoding encounters an issue, partial error data is retained on `EngineError` alongside `CaptureError`.
- **Batch SPP row errors**: `batch.EngineError(i)` returns a row-owned `*EngineError` for failed epochs (and `nil, nil` for successful epochs) directly from the batch handle, avoiding transient thread-local storage races. Calling `batch.Solution(i)` on a failed epoch returns `StatusSolve` with the row cause attached to `statusErr.Engine`.
- **Thread-local reset**: Generic error state is stored per OS thread, and `sidereon.ClearEngineError()` only clears the executing OS thread. Ordinary wrapper calls already capture errors and reset generic state at operation entry, and owned error records remain valid. For a deliberate cross-call native TLS sequence, the caller must hold `runtime.LockOSThread()` and `defer runtime.UnlockOSThread()`.

## Versioning

Go, C, and the canonical Sidereon engine release in lockstep: a published
`v4.2.0` Go module uses the matching C header macros and static library for
`4.2.0`. Go semantic import versioning requires the module path to carry the
major version from 2 onward, so the module path is
`sidereon.dev/go/v4` and each later major version moves
the suffix with it (`/v5`, ...), whether or not the Go API itself changed.
The package name stays `sidereon`.

This module does not bind PPP snapshot RAIM or carried RTK filter state.

## License

The Go module is MIT licensed; see [LICENSE](LICENSE). The statically linked C
library and its dependencies have additional attributions in
[THIRD-PARTY-NOTICES.md](THIRD-PARTY-NOTICES.md).
