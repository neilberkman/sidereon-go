# Sidereon Go audit evidence

Gate result: PASS

This evidence ledger records the reproducible public-source checks for the audit. The pinned C reference is `../../repos/sidereon-c` at `b38ecf8caf796a02f209dbb4cbebdaa4a042204c`. The documented baseline is `../../repos/sidereon-python` v1.1.1. Public `../../repos/sidereon` source is cited only where it establishes the semantics of a C route.

## Repository pins and hashes

Commands, run from the audit worktree:

```sh
git -C ../../repos/sidereon-c rev-parse HEAD
git -C ../../repos/sidereon-c describe --always --tags
git -C ../../repos/sidereon-c rev-parse HEAD^{tree}
git -C ../../repos/sidereon-c rev-parse HEAD^
git -C ../../repos/sidereon-c log -1 --format='%h %s'
```

Output:

```text
b38ecf8caf796a02f209dbb4cbebdaa4a042204c
v1.2.0-2-gb38ecf8
a726a28cdd629eb1bec7e1252f23616538e4329a
8d90c450d92680f084acc68ec72fe8bd5e3f3760
b38ecf8 Add bare SSR message decode and SBAS PRN lookup routes (#5)
```

```sh
wc -l ../../repos/sidereon-c/bindings/c/include/sidereon.h
shasum -a 256 ../../repos/sidereon-c/bindings/c/include/sidereon.h
shasum -a 256 ../../repos/sidereon-python/python/sidereon/__init__.pyi
for repo in sidereon sidereon-python sidereon-wasm sidereon-ex; do
  printf '%s ' "$repo"
  git -C "../../repos/$repo" rev-parse HEAD
done
```

Output:

```text
35760 ../../repos/sidereon-c/bindings/c/include/sidereon.h
6e79405bbce65d91958fe591279ad4c73f79f86573c64176f7c3dd0d9c29420d  ../../repos/sidereon-c/bindings/c/include/sidereon.h
24d0712209e561c03f52148006540a151e12aa048dc3bcd7ceb845f87be848f9  ../../repos/sidereon-python/python/sidereon/__init__.pyi
sidereon 267b28f69f7678921c11689531f8fa250e7cca78
sidereon-python 53be271eaf9974b3cf451065d0b0225f436bb944
sidereon-wasm c1791bc4cbeb7d51c8a516516cc62eed2ff88dc0
sidereon-ex 2716c05e7ff6aa4d6ac4749484b45f843fefd595
```

The C header version command and output are:

```sh
rg -n 'SIDEREON_VERSION_(MAJOR|MINOR|PATCH|STRING)' ../../repos/sidereon-c/bindings/c/include/sidereon.h
```

```text
66:#define SIDEREON_VERSION_MAJOR 1
67:#define SIDEREON_VERSION_MINOR 2
68:#define SIDEREON_VERSION_PATCH 0
69:#define SIDEREON_VERSION_STRING "1.2.0"
```

The future Go target is 1.3.0, but the pinned header’s 1.2.0 macros are expected pre-release state and are not a coverage gap.

## C declaration and Rust export invariants

Header declaration command:

```sh
rg -n '^[A-Za-z].*sidereon_[A-Za-z0-9_]+[[:space:]]*\('   ../../repos/sidereon-c/bindings/c/include/sidereon.h
```

Result: 1,503 declaration lines. The corresponding unique-name command returned 1,503:

```sh
rg -o 'sidereon_[A-Za-z0-9_]+' \
  <(rg -n '^[A-Za-z].*sidereon_[A-Za-z0-9_]+[[:space:]]*\('     ../../repos/sidereon-c/bindings/c/include/sidereon.h) |
  sort -u | wc -l
```

Rust export commands:

```sh
rg -n 'pub (unsafe )?extern "C" fn sidereon_'   ../../repos/sidereon-c/bindings/c/src | wc -l
rg -o 'pub (unsafe )?extern "C" fn sidereon_[A-Za-z0-9_]+'   ../../repos/sidereon-c/bindings/c/src |
  sed -E 's/.* fn //' | sort -u | wc -l
```

Results: 1,503 export lines and 1,503 unique Rust names.

Sorted set comparison:

```sh
comm -3 \
  <(rg -n '^[A-Za-z].*sidereon_[A-Za-z0-9_]+[[:space:]]*\('       ../../repos/sidereon-c/bindings/c/include/sidereon.h |
      sed -E 's/^([0-9]+):.*(sidereon_[A-Za-z0-9_]+)[[:space:]]*\(.*/\2/' |
      sort -u) \
  <(rg -o 'pub (unsafe )?extern "C" fn sidereon_[A-Za-z0-9_]+'       ../../repos/sidereon-c/bindings/c/src |
      sed -E 's/.* fn //' | sort -u) | wc -l
```

Result: `0`. The shared sorted-name SHA-256 is:

```text
a8eb09a0dd3a569f2ebf08406c461f882a8207ac12984bf450456730ce3a84b3
```

The prior committed map had 1,402 rows. Comparing its names with the current header found 101 current additions. The C source contains 85 tracked Rust files; the pinned HEAD changes 10 files.

## Coverage-map invariants

Commands:

```sh
rg '^\| [0-9]+ \| [0-9]+ \| sidereon_' audit/COVERAGE_MAP.md | wc -l
rg '^\| [0-9]+ \| [0-9]+ \| sidereon_' audit/COVERAGE_MAP.md |
  sed -E 's/^\| [0-9]+ \| [0-9]+ \| ([^ ]+).*/\1/' |
  sort -u | wc -l
comm -23 \
  <(rg -n '^[A-Za-z].*sidereon_[A-Za-z0-9_]+[[:space:]]*\(' \
      ../../repos/sidereon-c/bindings/c/include/sidereon.h |
      sed -E 's/^([0-9]+):.*(sidereon_[A-Za-z0-9_]+)[[:space:]]*\(.*/\2/' |
      sort -u) \
  <(rg '^\| [0-9]+ \| [0-9]+ \| sidereon_' audit/COVERAGE_MAP.md |
      sed -E 's/^\| [0-9]+ \| [0-9]+ \| ([^ ]+).*/\1/' |
      sort -u) | wc -l
comm -13 \
  <(rg -n '^[A-Za-z].*sidereon_[A-Za-z0-9_]+[[:space:]]*\(' \
      ../../repos/sidereon-c/bindings/c/include/sidereon.h |
      sed -E 's/^([0-9]+):.*(sidereon_[A-Za-z0-9_]+)[[:space:]]*\(.*/\2/' |
      sort -u) \
  <(rg '^\| [0-9]+ \| [0-9]+ \| sidereon_' audit/COVERAGE_MAP.md |
      sed -E 's/^\| [0-9]+ \| [0-9]+ \| ([^ ]+).*/\1/' |
      sort -u) | wc -l
```

Results: map rows `1,503`; unique map names `1,503`; missing `0`; obsolete `0`.

The map ends with:

```text
End of map: 1,503 rows; 1,503 unique function names; zero exclusions.
```

## Python baseline inventory

Commands:

```sh
for f in __init__.pyi data.pyi distribution.pyi exact_cache.pyi exact_sp3.pyi ntrip.pyi; do
  p="../../repos/sidereon-python/python/sidereon/$f"
  printf '%s ' "$f"; wc -l < "$p"
  printf 'classes='; rg -c '^class ' "$p"
  printf 'top_level_defs='; rg -c '^def ' "$p"
  printf 'unique_top_level_def_names='
  rg -o '^def [A-Za-z_][A-Za-z0-9_]*' "$p" |
    sed -E 's/^def //' | sort -u | wc -l
done
sed -n '1134,2530p' ../../repos/sidereon-python/python/sidereon/__init__.py |
  rg -o '"[A-Za-z_][A-Za-z0-9_]*"' | sort -u | wc -l
find ../../repos/sidereon-python/python/sidereon -maxdepth 1 -name '*.pyi' -print0 |
  xargs -0 cat | wc -l
```

Results:

```text
__init__.pyi 14529 lines; classes=701; top_level_defs=465; unique=464
data.pyi 490 lines; classes=31; top_level_defs=52; unique=52
distribution.pyi 185 lines; classes=28; top_level_defs=5; unique=5
exact_cache.pyi 70 lines; classes=7; top_level_defs=3; unique=3
exact_sp3.pyi 47 lines; classes=3; top_level_defs=2; unique=2
ntrip.pyi 86 lines; classes=12; top_level_defs=0; unique=0
final __all__ names=1194
all .pyi lines=15407
```

The broad capability grouping and routes are in `audit/COVERAGE_MAP.md`; the individual former-blocker checks are repeated in `audit/PARITY_REVIEW.md`.

## Source anchors for ABI contract

- C header ownership, status, error, and variable-output preamble: `bindings/c/include/sidereon.h:7-48`.
- C status enum: `bindings/c/include/sidereon.h:199-234`; last-error and status-message declarations: lines 24578 and 34273; version functions: lines 35625-35631.
- Rust thread-local error and panic boundary: `bindings/c/src/lib.rs:417-468`.
- Rust slice validation: `bindings/c/src/lib.rs:740-755`.
- Rust release helper: `bindings/c/src/lib.rs:800-804`.
- Rust variable-output query/copy: `bindings/c/src/lib.rs:836-871`.
- Public C crate kind and current route summary: `bindings/c/README.md:11`, `:95-142`.
- C status enum values are OK 0, NULL_POINTER 1, INVALID_ARGUMENT 2, INVALID_TOKEN 3, SP3_PARSE 4, SOLVE 5, PANIC 6, TIMEOUT 7.

These anchors support the ownership, error, query/copy, and concurrency decisions in the two audit documents.

## Source anchors for former blockers

- RTK builders: C header lines 19703-19726; C Rust `bindings/c/src/rtk.rs:4287-4380`; Python `__init__.pyi:3428-3717`, implementation `src/rtk.rs:2796-2814`.
- Covariance6: C header lines 20834-20914 and 20959; C Rust `bindings/c/src/covariance.rs:199-489`; Python `__init__.pyi:13376-13388`, implementation `src/covariance.rs:558-622`; public core `crates/sidereon-core/src/astro/covariance.rs:161-219`.
- Lenient NAV: C header lines 25172-25217 and 26188-26200; C Rust `bindings/c/src/rinex.rs:867-1020`; Python `__init__.pyi:13290`, implementation `src/rinex.rs:2821`; public core `crates/sidereon-core/src/rinex_nav/mod.rs:883`.
- GLONASS records: C header lines 26153-26164 and 27979-28017; C Rust `bindings/c/src/rinex.rs:1276-1512`; Python `__init__.pyi:6293`, implementation `src/rinex.rs:2904`; public core `crates/sidereon-core/src/rinex_nav/mod.rs:1478-1490`.
- RINEX ionosphere/leap: C header lines 26164-26175; C Rust `bindings/c/src/rinex.rs:1479-1567`; Python `__init__.pyi:6294-6295`, implementation `src/rinex.rs:2914-2920`; public core `crates/sidereon-core/src/rinex_nav/mod.rs:1211,1380`.
- SBAS line parsers and PRN lookup: C header lines 26210-26221 and 30715; C Rust `bindings/c/src/sbas.rs:39-116`; Python `__init__.pyi:12553-12556`, implementation `src/sbas_ssr.rs:2224-2248`; public core inverse semantics `crates/sidereon-core/src/sbas/store.rs:747-758`.
- Bare SSR: C header lines 33825-33944; C Rust `bindings/c/src/ssr.rs:403-735`; Python `__init__.pyi:12664-12723`, implementation `src/sbas_ssr.rs:2253-2261`.
- Generic SSR encoding composition: C header lines 28755 and 28968; C Rust generic RTCM message routes; Python `SsrMessage.encode` at `__init__.pyi:12695`.

## Evidence interpretation

The C README describes dedicated routes for covariance6, calendar/signal/LNAV helpers, full NAV and GLONASS records with skipped raw tokens, RINEX clocks, SBAS logs, RTK arcs, DTED tile lists, owning handles, and query/fill accessors. The current C Rust implementation confirms those are actual `extern "C"` functions and not documentation-only names.

For ownership, every constructor’s success path yields an owning handle and every current map row for a free function is assigned to an explicit Go `Close`. For errors, the Rust boundary catches panics and updates thread-local detail before returning a status. For arrays and byte strings, the C API requires query then copy; the Go design copies both inputs and outputs. Read-only handles are shareable while alive, but a Go wrapper must prevent concurrent mutation or release, and must keep status reads on the same OS thread.

No C or Python tests were run as part of this static audit. No public reference repository was edited. The result is based on committed header/source/stub text and reproducible inventories, not on private fixtures or downstream behavior.

## Final evidence decision

PASS. Header declarations: 1,503. Rust exports: 1,503. Sorted symbol-set diff: 0. Coverage-map rows: 1,503. Unique map names: 1,503. Missing map names: 0. Obsolete map names: 0. All documented capabilities, including the former eleven blockers, bare SSR decode/bias access, and SBAS PRN lookup, are expressible. The implementation gate is clear.
