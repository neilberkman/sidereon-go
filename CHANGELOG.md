# Changelog

All notable changes to this module are documented here.

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
