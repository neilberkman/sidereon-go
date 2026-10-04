# Changelog

All notable changes to this module are documented here.

## Unreleased

## 3.0.0 - 2026-10-04

- Module path is now `sidereon.dev/go/v3`, as Go semantic
  import versioning requires for major version 3. Import the package as
  `sidereon.dev/go/v3`; the package name is unchanged.
- Structured engine error inspection: `StatusError.Engine` (`*EngineError`) and
  `StatusError.EngineError()` provide access to the core engine's Schema 1 generic
  error details. `errors.As(err, &statusErr)` and `errors.As(err, &engineErr)`
  unwrap and inspect engine failures.
- Lossless payload and forward compatibility: `EngineError` preserves `Family`
  (`EngineErrorFamily`), `FamilyName`, `Schema` (uint32 1), `Operation`, `Kind`,
  `Fields` (`json.RawMessage`), `Payload` (`json.RawMessage`), and `CaptureError`.
  `UnmarshalFields(target any)` decodes `Fields` using `json.Number` to preserve
  integer lexemes without lossy `float64` conversion. Raw JSON fields retain future
  error codes and unknown variants without truncation. Exact floating-point values
  are preserved via `EngineFloat` with `decimal` string and `bits_hex` string (IEEE-754 hex).
- Partial capture contract: when native diagnostic capture, payload extraction, or
  JSON decoding encounters an error, partial error data (such as `Family`,
  `FamilyName`, and raw `Payload`) is preserved on `EngineError` alongside the
  non-nil diagnostic refusal in `CaptureError`.
- Explicit generic error reset: `ClearEngineError()` explicitly clears the generic
  engine error thread-local storage slot on the calling OS thread.
- SPP batch row-owned error handling:
  - `SPPBatch.EngineError(i int) (*EngineError, error)` returns `nil, nil` for an
    epoch that solved successfully, and returns the row-owned `*EngineError` for a
    failed epoch. If secondary diagnostic capture encounters an error, a partial
    `EngineError` is returned alongside `CaptureError`.
  - `SPPBatch.Solution(i int)`: when an epoch failed, returns `StatusError` with
    code `StatusSolve`, attaching the row's engine error cause to `statusErr.Engine`.
  - `SPPBatch.Error(i int)` preserves human-readable diagnostic error text as a
    compatibility method.
- Major surface additions verified in public source types:
  - FDE and RAIM: `SolveFDE`, `FDEOptions` (`PFA`, `MaxExclusions`,
    `MaxExclusionRMSM`, `WeightsMode`, `Weights`, `Systems`), `FDEResult`,
    `FDEUnresolvedError` (retaining last faulted solution, exclusions, iterations,
    RAIM result, normalized residuals, and `CaptureError`), `RAIMResult`,
    `RangeFDEOptions`, `RangeFDEResult`, and `DefaultRangeFDEOptions`.
  - Exact time and source queries: `ExactEpoch` (attosecond resolution),
    `ExactEpochQuery`, `NewExactEpoch`, `ExactEpochComponents`, and `ExactOrdering`.
    Precise ephemeris sources (`SP3`, `PreciseEphemerisInterpolant`,
    `PreciseInterpolantArtifact`) support exact time and source queries:
    `SourceStateAtEpochQueries`, `TransmitEpochClockAtEpochQueries`,
    `ClockRelativityAtEpochQuery`, and `EphemerisVarianceAtEpochQueries`, returning
    `EphemerisSourceState`, `TransmitEpochClock`, `ClockRelativity`, and `UT1DegradeReason`.
- Native toolchain and static archives:
  - Build scripts (`scripts/build-native-archives.sh`) and `rust-toolchain.toml`
    pin Rust toolchain 1.98.1.
  - Builds from a source checkout without matching prebuilt archives require
    rebuilding native static archives (`libsidereon.a`) using Rust 1.98.1 via
    `scripts/build-native-archives.sh`, or specifying an external library with
    `-tags sidereon_use_system_lib`.

## 2.1.0 - 2026-09-06

- Module path is now `github.com/neilberkman/sidereon-go/v2`, as Go semantic
  import versioning requires from major version 2 onward. The `v2.0.0` tag
  was never resolvable as a module version for that reason; `v2.1.0` is the
  first installable 2.x release (`go get github.com/neilberkman/sidereon-go/v2@v2.1.0`).
  Import the package as `github.com/neilberkman/sidereon-go/v2`; the package
  name is unchanged.
- Fixed: the ionosphere-free override band strings passed to the RTK surface
  were never freed, because the deferred release captured the slice before
  anything was appended to it.
- Engine update: sidereon-c 2.1.0 (sidereon-core 2.1.0), additive. The SP3
  coverage-gap threshold is now a validated, product-carried policy with the
  1.5 default bit-identical to before; the window-scoped continuity reach
  follows the interpolator's actual selectable node spans; RINEX 4 CNAV
  week/TOW round trips are stable at the week boundary.
- Native archives are built with Rust 1.98.0 (was 1.94.0); a
  `rust-toolchain.toml` pins the same compiler for local rebuilds.
- Expose the new sidereon-c 2.1.0 SP3 interpolation-policy entry points:
  - Options types `SP3InterpolationOptions`, `SP3LoadOptions`, `SP3ContinuityOptions`,
    and functional option `WithGapThresholdFactor`.
  - Option-aware loaders `LoadSP3WithOptions` and `LoadExactSP3WithOptions`, with
    backward-compatible variadic `opts ...SP3Option` on `LoadSP3` and `LoadExactSP3`.
  - Gap threshold factor getters on `SP3`, `PreciseEphemerisSamples`,
    `PreciseEphemerisInterpolant`, and `PreciseInterpolantArtifact`.
  - Option-aware continuity check routes `ContinuityWithOptions`, `CheckContinuity`,
    `CheckContinuityWithOptions`, and `ContinuityVerdictJSONWithOptions`.
  - Option-aware ephemeris builders `BuildPreciseEphemerisSamplesWithOptions` and
    `BuildPreciseEphemerisInterpolantWithOptions`.

## 2.0.0 - 2026-09-03

- Engine update: sidereon-c 2.0.0 (sidereon-core 2.0.0). The upstream engine
  has breaking changes on its Rust API (public input structs are non_exhaustive;
  terrain, IONEX TEC grid, and dense-output return typed error enums), but the C
  ABI and this Go API are unchanged, so nothing a Go caller writes has to move.

## 1.4.1 - 2026-08-31

- Engine update: sidereon-c 1.4.1 (sidereon-core 1.4.1; 1.4.0 is skipped, it
  reported a different first-order optimality for fits). Every
  transcendental and fused multiply-add now goes through portable kernels and
  decompositions run on a portable scalar, so results are bit-identical across
  the seven bundled targets; SVD-derived covariance and geometry diagnostics
  take the values 1.3.3 produced on x86_64/glibc. The static-position
  fixture pins move accordingly; positions, clocks, and residuals do not. No
  Go API changes.

## 1.3.3 - 2026-08-30

- Engine update: sidereon-c 1.3.3 (sidereon-core 1.3.3). Archive-listing
  parsing in the engine is no longer quadratic, and transcendental math is
  bit-identical across x86_64 and arm64. 1.3.2 is skipped: its Moon
  ephemeris lost the parallax sine and placed the Moon ~17 km off; 1.3.3
  restores it. No Go API changes.

## 1.3.1 - 2026-08-29

- Relicense the module from Apache-2.0 to the MIT License, matching the engine
  and every other Sidereon language interface. The Apache-2.0 text remains in
  `LICENSES/Apache-2.0.txt` for the Apache-licensed dependency choices.
- Engine update: sidereon 1.3.1 / sidereon-core 1.3.1. Coordination release
  keeping the shared release number across the language interfaces. No API
  changes.

## 1.3.0 - 2026-08-29

- Prepare the cgo binding for the lockstep Sidereon 1.3.0 release.
- Document supported targets, libc-specific archive selection, system-library
  builds, ownership, concurrency, and Go-owned I/O.
- Add release/version checks, packed-module consumer coverage, and ordinary CI.
- Include the applicable third-party notices and IERS-derived tide sources.

This section has no release date because `v1.3.0` has not been published.
