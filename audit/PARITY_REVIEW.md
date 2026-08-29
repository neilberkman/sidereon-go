# Sidereon Go parity review

Gate result: PASS

This review is a static audit of public sources only. It pins `../../repos/sidereon-c` at `b38ecf8caf796a02f209dbb4cbebdaa4a042204c`, uses `../../repos/sidereon-python` v1.1.1 as the documented baseline, and consults public `../../repos/sidereon` source only to verify semantics of routes that were previously questioned. No public reference repository was modified. No Go code is part of this change.

## Independent inventory result

The current C header contains 1,503 declaration lines with 1,503 unique function names. The Rust source contains 1,503 `pub extern "C" fn` exports with 1,503 unique names. The sorted header/Rust set difference is zero; the shared sorted-name SHA-256 is `a8eb09a0dd3a569f2ebf08406c461f882a8207ac12984bf450456730ce3a84b3`.

The refreshed coverage map contains 1,503 rows and 1,503 unique C names. It has zero missing and zero obsolete names relative to the current header. The prior 1,402-row map was an input to this review, not an authority; all 101 current additions were re-enumerated and added.

The C header SHA-256 is `6e79405bbce65d91958fe591279ad4c73f79f86573c64176f7c3dd0d9c29420d`. Its version macros remain `1.2.0` at lines 66-69. The future Go release target of 1.3.0 does not change this audit result: 1.2.0 is the expected pre-release state of the pinned C reference.

## Documented Python surface

The baseline is `../../repos/sidereon-python` at HEAD `53be271eaf9974b3cf451065d0b0225f436bb944`, with `python/sidereon/__init__.pyi` SHA-256 `24d0712209e561c03f52148006540a151e12aa048dc3bcd7ceb845f87be848f9`. The package implementation exposes 1,194 unique names in its final `__all__` construction. The typed surface has 701 classes and 465 function declarations, or 464 unique function names because `raim` is overloaded. Auxiliary stubs contain 81 classes and 62 functions across `data.pyi`, `distribution.pyi`, `exact_cache.pyi`, `exact_sp3.pyi`, and `ntrip.pyi`.

The complete family review is in `audit/COVERAGE_MAP.md`. It covers the positioning families (SP3, exact products, SPP, DGNSS, RTK, PPP, broadcast, observations, RINEX, RTCM, NMEA), astro families (TLE/SGP4, frames, time, passes, conjunction/TCA, orbit utilities and codecs), distribution families (SBAS, SSR, NTRIP, data/catalog/cache), geodesy families (coordinates, geoid, terrain, DTED, antenna/ANTEX, atmospheric and RF helpers), estimation families (metrics, covariance, DOP, RAIM/FDE, filters, NIS, clock/Allan), and presentation/path adapters. Each family has a C route or an explicitly Go-owned adapter; no documented numeric/parser family is left without an expression.

## Former blockers and current additions

| Capability | Python evidence | Current C/Rust evidence | Finding |
| --- | --- | --- | --- |
| `build_rinex_rtk_arc` | `__init__.pyi:3482`; `src/rtk.rs:2796` | header:19718; `bindings/c/src/rtk.rs:4287` | Dedicated builder. Count, epoch metadata, base/rover observations, satellite positions, wavelengths, offsets, skipped count, ownership and free are present. |
| `build_dual_frequency_rinex_rtk_arc` | `__init__.pyi:3717`; `src/rtk.rs:2814` | header:19703; `bindings/c/src/rtk.rs:4342` | Dedicated builder. Count, metadata, sort key, observations, all/base/rover satellite positions, skipped count, ownership and free are present. |
| covariance6 km↔m, diagonal, PSD interpolation | `__init__.pyi:13376-13382`; `src/covariance.rs:558-594` | header:20845-20878; `bindings/c/src/covariance.rs:293-375` | Complete 6x6 matrix input/output and unit/interpolation/validation operations are present. |
| covariance6 ECI↔RTN | `__init__.pyi:13385-13388`; `src/covariance.rs:597-622` | header:20834-20901; `bindings/c/src/covariance.rs:423-489` | Both transforms and validation delegate to public `Covariance6` semantics; full matrix output is copied. |
| `parse_rinex_nav_lenient` | `__init__.pyi:13290`; `src/rinex.rs:2821` | header:26188-26200; `bindings/c/src/rinex.rs:867-1020` | Dedicated owned result with record count/items and skipped count/items/messages. Diagnostics are not discarded. |
| `parse_sbas_ems_lines` | `__init__.pyi:12553`; `src/sbas_ssr.rs:2224` | header:26210; `bindings/c/src/sbas.rs:79-116` | Dedicated line parser returns owned blocks; count/item and byte query/copy accessors are present. |
| `parse_sbas_rtklib_lines` | `__init__.pyi:12554`; `src/sbas_ssr.rs:2232` | header:26221; `bindings/c/src/sbas.rs:99-116` | Dedicated line parser returns owned blocks with the same complete access pattern. |
| `parse_rinex_glonass_records` | `__init__.pyi:6293`; `src/rinex.rs:2904` | header:26153-26164, 27979-28017; `bindings/c/src/rinex.rs:1276-1512` | Representable records and skipped raw tokens are separate copied outputs, including extended slots that cannot become a typed record. |
| `parse_rinex_iono_corrections` | `__init__.pyi:6294`; `src/rinex.rs:2914` | header:26164; `bindings/c/src/rinex.rs:1479-1527` | Presence-tagged GPS/BeiDou alpha/beta values are retained. |
| `parse_rinex_leap_seconds` | `__init__.pyi:6295`; `src/rinex.rs:2920` | header:26175; `bindings/c/src/rinex.rs:1529-1567` | Optional value and presence flag are retained. |
| bare SSR decode and bias/accessor surface | `__init__.pyi:12664-12723`; `src/sbas_ssr.rs:2253-2261` | header:33825-33944; `bindings/c/src/ssr.rs:403-735` | Header fields plus orbit, clock, URA, code-bias and phase-bias counts and records are exposed. Nested signal arrays have count and query/copy accessors; free is present. |
| `sbas_prn_to_satellite_id` | `__init__.pyi:12555`; `src/sbas_ssr.rs:2240` | header:30715; `bindings/c/src/sbas.rs:39-59` | Exact token mapping is returned as copied bytes; no-match is successful zero-length output. |
| reverse `satellite_id_to_sbas_prn` | `__init__.pyi:12556`; `src/sbas_ssr.rs:2246-2248` | forward route at header:30715; core semantics at `crates/sidereon-core/src/sbas/store.rs:747-758` | No direct reverse C name exists, but the finite inverse is fully expressible by testing the documented PRN domain against the exact C forward route. It is a Go convenience, not a second numeric implementation. |
| bare `SsrMessage.encode` | `__init__.pyi:12695`; SSR implementation | generic `sidereon_rtcm_message_encode` and `sidereon_rtcm_message_to_frame`; header:28755, 28968 | Retain the decoded body or pass the generic message to the encoder; return copied frame bytes. This preserves the documented operation without a bare-message-only encoder symbol. |

These checks verify semantic output and accessors, not just similarly named functions. For the RINEX and SBAS parsers, the result container and diagnostic/raw-byte routes were inspected. For covariance, the C Rust functions delegate to the public core six-by-six operations. For SSR, the C Rust mappings include every documented record group and nested signal group.

## ABI ownership, errors, query/copy, and concurrency

The C header contract at lines 7-48 identifies owning handles, matching frees, thread-local error text, status behavior, and variable output rules. Rust implementation anchors are `bindings/c/src/lib.rs:417-468` for thread-local error and panic containment, `:740-755` for slice validation, `:800-804` for release, and `:836-871` for query/copy output.

Go wrapper decisions:

- C-owned pointers become non-copying Go handles. Constructors transfer ownership only on OK; `Close` is idempotent, clears the pointer, and has a `runtime.AddCleanup` backstop.
- Every non-OK status becomes a typed `StatusError` with the numeric status, stable status text, and same-call thread-local detail. The call and error read stay on one locked OS thread.
- Borrowed structs, strings, bytes, and arrays are copied into Go memory before release. Variable outputs use the C query-then-copy protocol and treat short buffers as invalid arguments.
- Live read-only handles may be shared. Mutating calls and `Close` are serialized; callers cannot use a handle concurrently with `Close`. C does not retain Go pointers.
- Go owns sockets, TLS, retries, filesystem paths, cache locking, decompression, and presentation aliases. The C ABI is used for byte parsing, numerical evaluation, validation, and protocol state.

## Exclusions and gate

No current C function is excluded from the one-to-one map. C free functions are represented by `Close`; status and version helpers are support APIs. Go-owned I/O and protocol transport are composition decisions, not missing C numerical capability. The reverse SBAS lookup and SSR encoding compositions are the only places where the documented Python name is not a direct C function name, and both have been verified as lossless expressions of public routes.

PASS. The C declaration count, Rust export count, map row count, unique-name count, and set comparison all satisfy the audit invariants. All documented capabilities are expressible. The implementation gate is clear.
