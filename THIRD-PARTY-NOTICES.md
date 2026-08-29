# Third-Party Notices

The Go binding is licensed under Apache-2.0 (see LICENSE). Its committed
`libsidereon` static library contains the C binding and the locked Rust engine
dependencies listed below. License choices shown for dual-licensed components
use a permitted Apache-2.0 or MIT option. No copyleft (GPL/LGPL/AGPL/MPL/EUPL/
CDDL) code or dependency is included.

The package also carries the exact license texts needed by the algorithm and
dependency attributions in `LICENSES/`. The Apache-2.0 text is reproduced in
the root `LICENSE` and applies to the Apache-licensed dependency choices.

--------------------------------------------------------------------------------
## RTKLIB (BSD 2-Clause)

The integer least-squares (MLAMBDA/LAMBDA) routine is a Rust port of RTKLIB's
`lambda.c`.

  Copyright (c) 2007-2020, T. Takasu, All rights reserved.

  Redistribution and use in source and binary forms, with or without
  modification, are permitted provided that the following conditions are met:

  1. Redistributions of source code must retain the above copyright notice,
     this list of conditions and the following disclaimer.
  2. Redistributions in binary form must reproduce the above copyright notice,
     this list of conditions and the following disclaimer in the documentation
     and/or other materials provided with the distribution.

  THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
  AND ANY EXPRESS OR IMPLIED WARRANTIES ARE DISCLAIMED. IN NO EVENT SHALL THE
  COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT,
  INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES ARISING IN ANY WAY
  OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH
  DAMAGE.

--------------------------------------------------------------------------------
## ERFA (BSD 3-Clause)

Nutation/precession coefficient tables and conventions are derived from ERFA
(Essential Routines for Fundamental Astronomy), itself derived from IAU SOFA.
The complete BSD 3-Clause license, including ERFA's SOFA-heritage terms, is
distributed as `LICENSES/ERFA-BSD-3-Clause.txt`, copied byte-for-byte from the
official ERFA 2.0.1 license.

--------------------------------------------------------------------------------
## SciPy (BSD 3-Clause)

The trust-region least-squares solver (`trust-region-least-squares`)
reimplements algorithms equivalent to SciPy's least-squares routines. The
complete BSD 3-Clause license is distributed as
`LICENSES/SciPy-BSD-3-Clause.txt`, copied byte-for-byte from the official
SciPy 1.18.0 license.

--------------------------------------------------------------------------------
## Locked Rust dependency graph

The following package/version entries are the complete `Cargo.lock` graph for
the public C static library. Entries marked build-only provide build scripts or
procedural macros and are not object-linked; platform entries are included for
the supported conditional builds.

### C binding and engine crates

- `sidereon-c` 1.2.0 — MIT; C ABI binding.
- `sidereon` 1.2.0 — MIT; ergonomic engine facade.
- `sidereon-core` 1.2.0 — MIT; numerical engine.
- `trust-region-least-squares` 0.10.0 — MIT; statically linked numerical
  dependency.

### Statically linked registry crates

The listed license option is the one used for this distribution. Copyright
holders and authors remain those recorded in each package's public manifest.

| Package | Version | License option |
| --- | --- | --- |
| `adler2` | 2.0.1 | Apache-2.0 |
| `approx` | 0.5.1 | Apache-2.0 |
| `block-buffer` | 0.10.4 | Apache-2.0 |
| `bytemuck` | 1.25.1 | Apache-2.0 |
| `cfg-if` | 1.0.4 | Apache-2.0 |
| `cpufeatures` | 0.2.17 | Apache-2.0 |
| `crc32fast` | 1.5.0 | Apache-2.0 |
| `crossbeam-deque` | 0.8.7 | Apache-2.0 |
| `crossbeam-epoch` | 0.9.20 | Apache-2.0 |
| `crossbeam-utils` | 0.8.22 | Apache-2.0 |
| `crypto-common` | 0.1.7 | Apache-2.0 |
| `digest` | 0.10.7 | Apache-2.0 |
| `either` | 1.16.0 | Apache-2.0 |
| `flate2` | 1.1.9 | Apache-2.0 |
| `fs2` | 0.4.3 | Apache-2.0 |
| `generic-array` | 0.14.7 | MIT |
| `getrandom` | 0.2.17 | Apache-2.0 |
| `itoa` | 1.0.18 | Apache-2.0 |
| `libc` | 0.2.186 | Apache-2.0 |
| `libloading` | 0.8.9 | ISC |
| `libm` | 0.2.16 | MIT |
| `matrixmultiply` | 0.3.10 | Apache-2.0 |
| `memchr` | 2.8.3 | MIT |
| `memmap2` | 0.9.11 | Apache-2.0 |
| `miniz_oxide` | 0.8.9 | Apache-2.0 |
| `nalgebra` | 0.33.3 | Apache-2.0 |
| `num-bigint` | 0.4.8 | Apache-2.0 |
| `num-complex` | 0.4.6 | Apache-2.0 |
| `num-integer` | 0.1.46 | Apache-2.0 |
| `num-rational` | 0.4.2 | Apache-2.0 |
| `num-traits` | 0.2.19 | Apache-2.0 |
| `paste` | 1.0.15 | Apache-2.0 |
| `rawpointer` | 0.2.1 | Apache-2.0 |
| `rayon` | 1.12.0 | Apache-2.0 |
| `rayon-core` | 1.13.0 | Apache-2.0 |
| `roxmltree` | 0.21.1 | Apache-2.0 |
| `safe_arch` | 0.7.4 | Apache-2.0 |
| `serde` | 1.0.228 | Apache-2.0 |
| `serde_core` | 1.0.228 | Apache-2.0 |
| `serde_json` | 1.0.150 | Apache-2.0 |
| `sha2` | 0.10.9 | Apache-2.0 |
| `shlex` | 2.0.1 | Apache-2.0 |
| `simba` | 0.9.1 | Apache-2.0 |
| `simd-adler32` | 0.3.9 | MIT |
| `thiserror` | 1.0.69 | Apache-2.0 |
| `typenum` | 1.20.1 | Apache-2.0 |
| `wide` | 0.7.33 | Apache-2.0 |
| `zmij` | 1.0.22 | MIT |

### Build-only and conditional entries

These entries are retained in the lockfile but are not object-linked into the
static archive: `autocfg` 1.5.1 (Apache-2.0 OR MIT), `cc` 1.2.67
(MIT OR Apache-2.0), `find-msvc-tools` 0.1.9 (MIT OR Apache-2.0),
`nalgebra-macros` 0.2.2 (Apache-2.0), `proc-macro2` 1.0.106 (MIT OR
Apache-2.0), `quote` 1.0.46 (MIT OR Apache-2.0), `serde_derive` 1.0.228
(MIT OR Apache-2.0), `syn` 2.0.118 (MIT OR Apache-2.0),
`thiserror-impl` 1.0.69 (MIT OR Apache-2.0), `version_check` 0.9.5
(MIT/Apache-2.0), and `unicode-ident` 1.0.24
((MIT OR Apache-2.0) AND Unicode-3.0). The supported Windows GNU build may
also resolve `winapi` 0.3.9, `winapi-i686-pc-windows-gnu` 0.4.0,
`winapi-x86_64-pc-windows-gnu` 0.4.0, and `windows-link` 0.2.1; all permit
Apache-2.0 or MIT. `wasi` 0.11.1+wasi-snapshot-preview1 is a conditional
lock entry for non-supported targets and permits Apache-2.0, Apache-2.0 with
the LLVM exception, or MIT.

The exact ISC notice for `libloading` is in `LICENSES/ISC-libloading.txt`.
The MIT-only package copyright notices are: `generic-array` 0.14.7, Bartłomiej
Kamiński; `libm` 0.2.16, the Rust Project Developers; `memchr` 2.8.3, Andrew
Gallant; `simd-adler32` 0.3.9, Marvin Countryman; and `zmij` 1.0.22, David
Tolnay. The C binding, engine crates, and trust-region solver carry the Neil
Berkman copyright in their distributed MIT licenses. The MIT text used by
these components is reproduced below; each package's copyright notice remains
associated with its entry above.

  MIT License

  Permission is hereby granted, free of charge, to any person obtaining a copy
  of this software and associated documentation files (the "Software"), to deal
  in the Software without restriction, including without limitation the rights
  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
  copies of the Software, and to permit persons to whom the Software is
  furnished to do so, subject to the following conditions:

  The above copyright notice and this permission notice shall be included in all
  copies or substantial portions of the Software.

  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
  SOFTWARE.

--------------------------------------------------------------------------------
## IERS Conventions Software

The solid-earth / ocean / pole tide displacement follows the IERS Conventions
reference routines (e.g. DEHANTTIDEINEL), used under the IERS Conventions
Software License. This is Sidereon-derived Rust code, not software distributed
or endorsed by the IERS Conventions Center. The routines were renamed, and the
source describes how the derived implementation differs from the original.

The full official notice is reproduced in
`LICENSES/IERS-Conventions-Software-License.txt` from the official
[`DEHANTTIDEINEL.F`](https://iers-conventions.obspm.fr/content/chapter7/software/dehanttideinel/DEHANTTIDEINEL.F)
source. The exact public non-test tide sources from the statically linked
[sidereon-core 1.2.0 crate](https://crates.io/crates/sidereon-core/1.2.0)
are distributed under `third_party_source/sidereon-core-1.2.0/tides/`.
The source comments retain the derivation, renamed routines, IERS
acknowledgment, and difference-from-original text.
Published results obtained with these routines should acknowledge use of the
IERS Conventions software.

--------------------------------------------------------------------------------
## Reference algorithms (no code copied)

The following informed reimplementations from public specifications/literature;
no source code was copied:

- SGP4 / SDP4: D. Vallado et al., "Revisiting Spacetrack Report #3" (AIAA), and
  the CelesTrak reference vectors (validation only).
- Frame/time-scale conventions cross-checked against Skyfield (MIT) and the IAU
  conventions.
- Galileo NeQuick-G: reimplemented from the Galileo OS SIS ICD "Ionospheric
  Correction Algorithm for Galileo Single Frequency Users"; MODIP and CCIR data
  tables transcribed as ITU-R / EU-JRC reference data (facts).
- NRLMSISE-00: U.S. Naval Research Laboratory (public domain).
