# Sidereon Go binding — gate-1 parity review

## Independent conclusion

**STOP.** The committed C ABI is broad and covers most of the README’s major
engine families, but it cannot express the entire documented Python v1.1.1
surface under the stated rule that Go may not add numeric/modeling behavior.

The stop is not based on naming or wrapper style. The five covariance helpers
alone are independently sufficient and indisputable. The following concrete,
independent blockers are sufficient without relying on the broader findings
later in this report:

* Five public 6-by-6 covariance helpers have no corresponding C route. The C
  ABI has covariance propagation and a 3-by-3 RTN conversion, but not the
  Python 6-by-6 unit conversions, PSD interpolation, or either 6-by-6 frame
  transform.
* Six directly verified parser gaps have no corresponding C route:
  `parse_rinex_nav_lenient`, `parse_sbas_ems_lines`,
  `parse_sbas_rtklib_lines`, `parse_rinex_glonass_records`,
  `parse_rinex_iono_corrections`, and `parse_rinex_leap_seconds`.

The broader findings remain useful review candidates, including calendar
helpers, rich record and sample-series handles, RTK arc builders, RINEX
frequency mapping, LNAV bit helpers, and DTED tile-list builders. They are
labelled as additional candidates below where semantic-equivalence review may
still matter; the gate does not depend on them. The confirmed gaps are
engine/protocol behavior, not caller-owned I/O, and a Go implementation cannot
close them while remaining within the gate rules.

The separate exhaustive 1,402-row map remains consistent with this review: it
covers the complete generated-header declaration inventory. Its exhaustiveness
does not establish that every documented Python operation has an equivalent C
route, which is the distinct parity question tested here. No conclusion below
asks that map to omit a C declaration or treats the C inventory as anything
other than 1,402 rows.

## Source and tag evidence

Only the two public read-only repositories under `../../repos` were inspected.
Both worktrees were clean at the time of review.

| Reference | Tag / commit | Evidence used |
|---|---|---|
| `../../repos/sidereon-python` | `v1.1.1`, `53be271eaf9974b3cf451065d0b0225f436bb944` | `README.md`; six public stubs under `python/sidereon/*.pyi`; Rust binding sources under `src/` to establish whether a stub is a core call, alias, or caller-owned wrapper |
| `../../repos/sidereon-c` | `v1.1.1`, `5cf4c1d796e66fff982eb658af0bd7bdba2beb24` | committed `bindings/c/include/sidereon.h`, including declarations, value structs, opaque handles, comments, and version macros |

The header identifies itself as version 1.1.1 at lines 47–50 and contains the
generated-header notice at line 1. The Python README says the package mirrors
the full engine at lines 101–103, explicitly says the binding adds no modeling
of its own at lines 227–230, and says “Full signatures live in the bundled type
stubs” at lines 229–230. The `.pyi` functions are therefore documented public
surface for this review, not merely typing metadata. That statement is
important: an apparent Python helper cannot be replaced with newly authored Go
mathematics merely because the formula is small.

## Inventory method and counts

The inventory was performed independently of the separate coverage map:

1. Enumerate every README capability bullet (20 areas, `README.md:105–198`).
2. Parse all six `.pyi` files with Python’s AST, counting module functions,
   classes, methods, properties, and constructors. Dunder methods were kept in
   the method total and separated from ordinary public methods during review.
3. Parse the C header’s public `sidereon_*` declarations and typedef families.
   The exact generated-header declaration count used here is:

   ```text
   rg -n '^[A-Za-z].*sidereon_[A-Za-z0-9_]+[[:space:]]*\(' ../../repos/sidereon-c/bindings/c/include/sidereon.h | wc -l
   ```

   It returns **1,402**. As an independent export-side check, the Rust C
   export scan is:

   ```text
   rg -n 'pub unsafe extern "C" fn sidereon_|pub extern "C" fn sidereon_' ../../repos/sidereon-c/bindings/c/src | wc -l
   ```

   It also returns **1,402**. The header count includes status-returning entry
   points and void release or mutating helpers; it is not a count of Go
   wrappers. The earlier result came from a narrower declaration regex that
   missed pointer-return and otherwise differently shaped declarations.
4. For every suspected mismatch, trace the Python implementation registration
   and core call where available, then search the whole committed header for a
   semantic/equivalent route, not merely a same-spelling symbol. This catches
   aliases, batch loops, handle composition, and caller-owned wrappers.
5. Review the C comments for constructor/free pairing, borrowed storage,
   mutable state, thread-local errors, and variable-length outputs.

| Artifact | Lines | Classes | Module functions | Class methods | Properties | Constructors |
|---|---:|---:|---:|---:|---:|---:|
| `__init__.pyi` | 14,529 | 701 | 465 | 4,331 | 3,117 | 237 |
| `data.pyi` | 490 | 31 | 52 | 20 | 3 | 6 |
| `distribution.pyi` | 185 | 28 | 5 | 17 | 2 | 3 |
| `exact_cache.pyi` | 70 | 7 | 3 | 9 | 0 | 2 |
| `exact_sp3.pyi` | 47 | 3 | 2 | 8 | 6 | 1 |
| `ntrip.pyi` | 86 | 12 | 0 | 11 | 0 | 4 |
| **Total** | **15,407** | **782** | **527** | **4,396** | **3,128** | **253** |

The C header is 33,803 lines and contains 1,402 public `sidereon_*` function
declarations, 141 `Sidereon*` enum typedefs, 530 value-struct definitions,
and 119 opaque handle forward declarations. These totals are inventory
sanity checks, not claims that one C function equals one Python row. In
particular, Python classes often expose a handle plus many accessors, while
some Python functions are aliases or loops over a scalar C route.

## Gate interpretation used

* A numerical, estimation, orbital, GNSS, protocol-decoding, or modeling
  operation must be delegated to a committed C route with equivalent input,
  output, and failure semantics.
* A Python class is covered only when the C side has the required constructor
  or source handle and enough accessors to reproduce its documented public
  fields/methods. A count-only or evaluation-only accessor is not a substitute
  for raw records or sample series.
* A presentation helper, alias, array/list shaping operation, or a loop over an
  existing C scalar route may be composed in Go, provided it adds no domain
  table, numerical rule, or changed error semantics.
* Caller-owned file/path handling, HTTP, decompression, cache orchestration,
  and NTRIP socket/reconnect orchestration are allowed by the task. The
  protocol/engine bytes parser still has to be available through C unless the
  operation is only shaping an already-decoded C result.

Thus simple aliases, loops over existing C scalar functions, and permitted
caller-owned transport remain valid Go composition. They are not being treated
as ABI gaps here; only the documented engine/protocol behavior that cannot be
expressed through the committed C routes contributes to the STOP basis.

## README capability-area matrix

`COVERED` means that the engine capability has an evident C route or a valid
caller-owned composition. `PARTIAL` means the named area contains one of the
confirmed gaps or additional candidates detailed below; it is not a claim that
every neighboring route is absent.

| README area and lines | Python public evidence | C evidence / disposition |
|---|---|---|
| Orbit propagation `105–110` | TLE/OMM propagation, force models, decay latch, batch/constellation arcs, passes, look angles, coverage | `sidereon_tle_propagate` and decay-latch routes at `32894–32909`, batch propagation at `25836`, pass prediction at `32820`, and visibility/geometry at `30992`. **COVERED for the listed engine families.** |
| Orbital mechanics `111–116` | Elements, anomalies, Kepler, Lambert, IOD, fit, relative RIC/RTN/LVLH | Representative C routes: `sidereon_propagate_kepler` at `25808`, Lambert at `23531`, `sidereon_rv2coe` at `28420`, and relative-state routes at `26354–26364`. **COVERED on the ABI expressiveness test.** |
| GNSS positioning `117–124` | SPP/static, RINEX assembly, RTK/PPP, DGNSS, robust solve, DOP | SPP/PPP/RTK declarations at `30233–30296`, RINEX-SPP solve at `30344`, and DOP at `20649–20663`. However, the separately documented RINEX RTK arc builders are absent (`__init__.pyi:3482–3503, 3717–3722`; see additional candidates). **PARTIAL (additional candidate; not needed for STOP).** |
| Integrity and error bounds `125–132` | RAIM/FDE, ARAIM, SBAS protection, reliability, observability, covariance metrics | RAIM routes at `25894–25925`, ARAIM at `18348–18402`, and SBAS correction/protection routes at `28977–29004`. **COVERED for the cited capability families; 6-by-6 helper gap remains under covariance.** |
| GNSS corrections and products `133–138` | SBAS/RTCM SSR, broadcast ephemerides, biases, ionosphere, troposphere, NTRIP | RTCM message decode/SSR accessors at `27126` and `27410–27428`; SBAS block decode/store at `28743–29004`; bias/IONEX routes include `18478–18486` and `23295`. NTRIP state-machine bytes are at `24332–24358`. **COVERED for binary engine routes; SBAS text-log parsing is a Formats blocker.** |
| Ephemeris and time `139–143` | Broadcast and precise products, SPK, uniform sampling, timescales, EOP | SP3/broadcast/SPK and sampling routes are present, including broadcast parse/accessors at `18652–18719`, but the Python calendar helpers at `4941–4950` lack C equivalents and the broadcast raw-record surface is incomplete. **PARTIAL (additional candidates; not needed for STOP).** |
| Timing and clocks `144–146` | ADEV/MDEV/HDEV/TDEV and power-law clock-noise fit | C clock routes at `19159–19361` cover the Allan-family and power-law operations. The separate RINEX clock-product sample-series gap is recorded below. **COVERED for README clock stability; RINEX clock API has an additional candidate.** |
| Estimation and detection `147–149` | Kalman/alpha-beta, gating, MAD, CFAR, ToA/TDoA localization | C alpha-beta at `18207`, Kalman at `23471`, MAD at `23721–23729`, CFAR at `19063–19093`, and source solve/accessors at `30538–30616`. **COVERED.** |
| Geodesy and monitoring `150–155` | Karney geodesics, frame catalog, station velocity, trajectory/steps/network/sidereal filtering | Geodesic routes at `22509–22521`; frame catalog at `21757–21797`; the header also contains monitoring/filter families. `terrestrial_frame_catalog` is an alias to `frame_catalog` (source `frame_catalog.rs:362–376`). **COVERED / composition.** |
| Geometry and events `156–158` | Frames, ECEF/geodetic, look angles, eclipse, conjunction, angular geometry | Frame routes span `21824–22026`; eclipse routes at `20896–20928`; visibility/geometry routes at `30938–30992`. **COVERED for the listed families.** |
| Observation and almanac `159–163` | Apparent places, refraction/polar motion, visual magnitude, rise/set, seasons, planetary/lunar/solar events | Observation and event declarations are present; satellite visual magnitude is at `28721`, with astronomy/geometry domains throughout the header. **COVERED on the ABI expressiveness test.** |
| Observation quality and integrity `164–166` | RINEX QC, post-solve RAIM/ARAIM, carrier combinations, Hatch smoothing | RINEX repair/QC at `26829–26858`, RAIM/ARAIM above, carrier/combination routes in the header. **COVERED for the cited engine families.** |
| Terrain `167–170` | DTED lookup, mmap store, tile-list builders, geoid grids/PROJ GTX | DTED lookup at `20781–20846`, mmap readers and metadata at `23839–23994`, geoid routes at `22816–22950`, and tree store writer at `33782–33788`. The explicit tile-list builders have no C route (`__init__.pyi:11832–11840`). **PARTIAL (additional candidate; not needed for STOP).** |
| RF `171` | FSPL, EIRP, C/N0, dish/link gain | C RF routes at `26507–26553` and canonical frequency/wavelength routes at `22034–22041`, `33760–33765`. **COVERED except the separate RINEX frequency-catalog helpers.** |
| GNSS/INS fusion `172–178` | Mechanization, EKF/UKF, loose/tight, RTS, timing, field mode, constraints | Fusion constructor/state/update/recording routes at `22067–22336`. **COVERED.** |
| Reference-station static solve `179–181` | Multi-epoch rover/reference observations, selected fixed/DGNSS/float result | C reference-station solve at `30417–30421` and static RINEX RTK solve at `30431–30435`. **COVERED for the end-to-end solve.** |
| Scenario simulation `182–185` | Deterministic synthetic observables, error budget, ground-truth ledger | JSON scenario variants at `29061–29127`. **COVERED.** |
| Signal analysis `186–188` | BPSK/BOC spectra, separation, DLL jitter, multipath envelopes | Signal analysis family at `29477–29875`. **COVERED.** |
| Formats `189–190` | TLE/OMM, CCSDS, RINEX, CRINEX, SP3, IONEX, ANTEX, Bias-SINEX, SBAS logs, RTCM, NMEA | Many direct parsers are present: CRINEX `20113–20130`, RINEX OBS `26757`, RTCM `27109–27141`, NMEA `24222–24261`, Bias-SINEX `18478–18486`. The six confirmed parser gaps and the additional RINEX NAV/clock, frequency, and DTED candidates below make this area **PARTIAL**; the gate does not depend on the candidates. |
| Data acquisition `191–198` | Product catalogs, exact identity validation, direct/CDDIS/local/in-memory sources, caches, HTTP/decompression | C data/catalog/identity routes at `20172–20436` and exact-cache lock/single-flight/read/publish at `21239–21446`. Python’s HTTP, auth, filesystem, decompression, retries, and caller-provided bytes are explicitly allowed Go-owned behavior. **COVERED by allowed composition plus C identity/cache semantics.** |

## Suspected gaps, with resolution and evidence

The table separates hard engine gaps from valid composition. “No declaration”
means a whole-header search found no semantic equivalent, not merely no exact
same-spelling symbol.

### Confirmed STOP basis

| Python surface | Concrete Python evidence | C evidence and resolution |
|---|---|---|
| `covariance6_km_to_m`, `covariance6_m_to_km`, `interpolate_covariance6`, `eci_to_rtn_covariance6`, `rtn_to_eci_covariance6` | Stubs `__init__.pyi:13376–13390`; binding implementations `src/covariance.rs:556–618` call core 6-by-6 conversion, PSD interpolation, and 6-by-6 RTN transforms. | C has `SidereonCovarianceMatrix6` at `5387–5389`, `sidereon_propagate_covariance` at `25801–25806`, and `sidereon_covariance_transport` at `20005–20012`, but no 6-by-6 conversion/interpolation/transform routes. The only named RTN conversion is explicitly **3-by-3** at `28400–28409`; the 3-by-3 PSD checks at `19988–20003` do not help. These are numeric engine operations: **five hard STOP gaps**. `propagate_covariance` itself is covered by `25801–25806`. |
| Six directly verified parser functions: `parse_rinex_nav_lenient`, `parse_sbas_ems_lines`, `parse_sbas_rtklib_lines`, `parse_rinex_glonass_records`, `parse_rinex_iono_corrections`, and `parse_rinex_leap_seconds` | Stubs: `parse_rinex_nav_lenient` at `__init__.pyi:13290`; the SBAS parsers at `12553–12554`; and the three RINEX parsers at `6293–6295`. The Rust wrappers call the engine parsers at `src/rinex.rs:2821–2824, 2904–2921` and `src/sbas_ssr.rs:2224–2235`. | C has filtered NAV-store parsing at `18652–18662` and binary SBAS block decoding/ingestion beginning at `28743` and `28977–29004`, but no declaration for the lenient skipped-diagnostics result, either timestamped SBAS text-log parser, or any of the three RINEX text parsers. The available binary/store routes do not reproduce those parser inputs and outputs. **Six confirmed parser gaps.** |

### Additional candidates (not needed for STOP)

The following findings are retained for the supervisor and coverage-map review.
They are not part of the independent STOP basis above; each is an additional
candidate unless the exhaustive map establishes an exact semantic composition.

| Python surface | Concrete Python evidence | C evidence and resolution |
|---|---|---|
| `split_julian_date`, `second_of_day`, `day_of_year` | Stubs `__init__.pyi:4941–4950`; binding source `src/frames.rs:546–572` directly calls `core_split_julian_date`, `core_second_of_day`, and `core_day_of_year`. | C has `civil_to_j2000_seconds` at `19138–19150`, `split_jd_to_j2000_seconds` at `31663–31671`, and leap-second/timescale routes at `23588–23596`, `32617–32646`, but no split-JD constructor, second-of-day, or day-of-year route. **Additional calendar-helper candidate**; it is not needed for STOP. |
| `parse_rinex_nav_records` and `encode_rinex_nav` | Stubs `__init__.pyi:6291–6292`; source `src/rinex.rs:2811–2817, 2891–2899` parses raw `BroadcastRecord` values and serializes them. | C parses the filtered store at `18652–18662` and serializes that store at `26610–26624`. Its raw record output at `18679–18683` is `SidereonBroadcastRecordInfo`, whose definition `4484–4500` has metadata only. **Additional raw-record/record-list candidate**; the store route is not obviously an equivalent raw-record route. |
| `BroadcastEphemeris` rich handle surface | Stubs `__init__.pyi:2002–2023`; implementation `src/rinex.rs:2258–2339` exposes full `records`, `glonass_records`, `iono_corrections`, header leap seconds, frequency-channel map, counts, and setter. | C has parse/count/metadata/select/message preference at `18660–18712`, but no GLONASS-record count/accessor, ionosphere accessor, leap-second accessor, frequency-channel map accessor, or full `BroadcastRecord`/`GlonassRecord` construction. **Additional rich-handle candidate.** |
| `parse_rinex_clock_lossy` and `RinexClock` sample series | Stubs `__init__.pyi:1818–1837` and functions `6300–6307`; implementation `src/rinex.rs:908–975` exposes satellite list, full `ClockSeries` samples, `series_for`, civil-epoch lookup, GPS-seconds lookup, and text serialization. Lossy parsing calls `CoreRinexClock::parse_lossy` at `143–146` and is registered at `2960–2975`. | C has strict parse `26577–26586`, satellite count `26588–26594`, point lookup `26555–26568`, and text serialization `26596–26608`. It has no lossy parser, satellite-list accessor, series/sample accessor, or `ClockSeries` construction route. **Additional RINEX clock rich-surface candidate.** File loading around it remains allowed caller-owned I/O. |
| RINEX RTK arc builders | Stubs `__init__.pyi:3469–3487` and `3708–3722`; source `src/rtk.rs:2469–2509` exposes arc epochs/wavelengths/offsets/skipped count, and `2793–2831` calls the two core builders. | C exposes option value structs at `13711–13740` and `13768–13785`, plus downstream solve routes at `30431–30435` and raw arc solve at `30447–30450`. It has no `RinexRtkArc`/dual-arc handle, builder declaration, or arc accessors. **Additional RTK arc-builder and handle candidate.** |
| `rinex_band_frequency_hz`, `rinex_band_wavelength_m`, `rinex_observation_frequency_hz`, `rinex_observation_wavelength_m` | Stubs `__init__.pyi:6691–6710`; source `src/observables.rs:1647–1697` calls core RINEX band/code frequency mapping, including GLONASS channel and RINEX-version policy. | C has canonical `(system, CarrierBand)` frequency/wavelength at `22034–22041` and `33760–33765`, plus GLONASS G1 at `22984–22989`, but no RINEX band/code/version mapping declarations. **Additional RINEX frequency-policy candidate.** |
| `default_pair` | Stub `__init__.pyi:6711`; source `src/observables.rs:1699–1703` calls core `default_iono_free_pair`. | The header has combination math and canonical carrier enums, but no obvious default-pair selection route. **Additional unresolved catalog candidate**; no gate conclusion depends on it. |
| `lnav_tow`, `lnav_subframe_id`, `lnav_parity`, `lnav_parity_valid` | Stubs `__init__.pyi:8256–8266`; binding source `src/lnav.rs:310–355` delegates these bit operations to core. | C has only full three-subframe `sidereon_lnav_decode` at `23610–23625` and full parameter `sidereon_lnav_encode` at `23627–23643`. A full decode is not obviously an equivalent one-subframe TOW/ID/parity API. **Additional LNAV helper candidate.** |
| DTED explicit tile-list store builders | Stubs `__init__.pyi:11831–11840`; source `src/terrain_store.rs:1020–1053` calls core tile-list serialization. | C has tree-to-store at `20848–20860` and `33782–33788`, plus mmap readers/metadata at `23839–23994`, but no tile-list input struct/builder route. **Additional DTED tile-list candidate.** |

### Resolved apparent gaps and allowed composition

| Apparent mismatch | Evidence and resolution |
|---|---|
| `j2000_seconds` looked like the missing calendar route | Python calls core `j2000_seconds` at `src/frames.rs:559–562`; C’s equivalent `sidereon_civil_to_j2000_seconds` is documented at `19138–19150`. **Covered.** |
| `leap_seconds_batch` looked like a missing batch ABI | Python explicitly documents it as a convenience loop over scalar `leap_seconds` at `src/frames.rs:916–932`; C scalar route is `23588–23596`. A Go loop adds no numeric rule. **Allowed composition.** |
| `propagate_covariance` looked absent with the other covariance helpers | The Python signature is `__init__.pyi:13391–13409`; C has the matching propagation handle route at `25801–25806`. **Covered; only the five independent helper functions are missing.** |
| `decode_sbas_message` looked like a second decoder | Python implementation is a direct call to `decode_sbas_block` at `src/sbas_ssr.rs:2216–2220`; C block decode is available. **Alias/composition.** |
| `sbas_prn_to_satellite_id` and reverse mapping looked like missing SBAS engine routes | Python implementation is a token/catalog conversion at `src/sbas_ssr.rs:2238–2248`; C has constellation token/catalog routes, including record access at `19770–19790`. **Presentation/catalog composition, not a new correction model.** |
| `decode_ssr` looked like a second missing SSR decoder | Python implementation delegates directly to `decode_ssr_message` at `src/sbas_ssr.rs:2251–2263`; C RTCM SSR message accessors are at `27410–27428`. **Alias; no independent engine gap.** |
| `decode_crinex_lines` looked like a missing C function | Python first invokes the core CRINEX decoder and only collects lines at `src/rinex.rs:2984–2992`; C has byte/text decode at `20113–20130`. **C route plus presentation shaping.** |
| `gnss_dop_series` looked like a missing series solver | Python source validates the epoch grid then repeatedly derives receiver geometry and DOP (`src/events.rs:1044–1090`); C has geometry visibility/pass routes at `30938–30992`, line-of-sight conversion at `23605–23608`, and DOP at `20649–20663`. **Valid composition, subject to using the C routes for each numeric result.** |
| `enumerate_fault_modes` looked like a missing ARAIM engine | Python source calls core enumeration at `src/araim.rs:791–804`; C ARAIM result/fault-mode routes are at `18348–18402`. **Covered by existing ARAIM handle/result routes.** |
| `terrestrial_frame_catalog` looked like a second frame catalog | It directly calls `frame_catalog()` at `src/frame_catalog.rs:362–376`; C frame catalog count/entries are at `21757–21778`. **Deprecated-name-free alias/composition.** |
| `carrier_frequency_hz`, `wavelength_m`, `gamma`, and combination math | Canonical frequency/wavelength are direct C routes at `22034–22041` and `33760–33765`; combination declarations are in the C combination family around `19449–19519`. **Covered.** The RINEX code/band mapping functions are not covered by this finding. |
| Data acquisition, HTTP, compression, and caller-provided bytes | README scope is `191–198`; stubs expose source choices, retries, auth, cache, decompression, and in-memory inputs at `data.pyi:276–308`, `350–417`, `distribution.pyi:13–18,161–183`, and `exact_cache.pyi:34–70`. C supplies product identity/catalog/validation at `20172–20436` and exact cache coordination at `21239–21446`. **Allowed Go-owned composition.** |
| NTRIP client | Python `ntrip.pyi:49–86` owns sockets, TLS, reconnect, stream iteration, and store feeding. C supplies the byte state machine, GGA, events, and reset at `24332–24358`; Python’s `classify_http_response` is a small helper around core at `src/ntrip.rs:1010–1017`. **Allowed socket/HTTP orchestration with C parser/state-machine calls.** |

## Deprecated aliases requiring an explicit Go decision

Deprecated Python names remain part of the documented public surface until the
binding’s parity policy says otherwise. They cannot be silently omitted from a
coverage count.

| Python name | Evidence | C route / required parity decision |
|---|---|---|
| `chan_ho_initial_guess` | Stub documents deprecation since 1.1.0 and points to `closed_form_initial_guess` at `__init__.pyi:10840–10851`. | C retains the explicitly deprecated `sidereon_chan_ho_initial_guess` at `19098–19111` and the canonical `sidereon_closed_form_initial_guess` at `19378`. **Decision required:** preserve the old Go export and warning semantics, or record a deliberate documented omission. It is not a numeric ABI gap. |
| `spp_robust_fde_driver` | Stub is at `__init__.pyi:11971–11980`; implementation emits `DeprecationWarning` and calls the shared implementation at `src/qc.rs:509–541`. | C has robust FDE SPP/broadcast routes at `26972–26982`. **Decision required:** keep the deprecated Go name as an alias with its documented warning, or explicitly record the omission. The underlying engine route exists. |

## Handle constructors, accessors, ownership, and concurrency

### What the header documents sufficiently

The top-level header contract is unusually clear for immutable handles:

* Newly allocated opaque handles are returned through out-parameters and each
  non-null handle must be released exactly once with its matching `_free`; NULL
  free is a no-op (`sidereon.h:7–18`).
* Opaque handles are “read-only shareable after creation,” but concurrent free
  invalidates the handle and borrowed pointers (`sidereon.h:16–18`).
* Only `SIDEREON_STATUS_OK` is success; failures set thread-local error state,
  outputs may be initialized/cleared before validation, and no new handle is
  transferred on failure (`sidereon.h:20–26`, `168–178`).
* Variable-length outputs use a size-query call with NULL/zero length; too-small
  buffers return `INVALID_ARGUMENT`, write zero elements, and report the
  required count (`sidereon.h:28–32`).
* The error must be copied immediately from the current thread using
  `sidereon_last_error_message` (`23540–23551`). A Go wrapper must not treat the
  last error as process-global or read it later on another goroutine.

This is enough to define conservative immutable-handle behavior: a Go value
may share an immutable C handle across readers while it remains live; every
borrowed output must be copied before the parent handle is released; no call
may race with release.

The borrowed-storage rule is explicit for assembled RINEX SPP inputs:
`sidereon_rinex_spp_inputs_epoch_inputs` returns pointer fields that remain
valid only until `sidereon_rinex_spp_inputs_free` (`26882–26900`). This same
copy-before-release discipline must be applied to every pointer-returning
accessor, even where an individual comment is shorter.

### Mutable handles are not documented as concurrently safe

The header does not grant a positive concurrent-mutation guarantee. The
following are visibly stateful because the API mutates them or has an explicit
mutable safety comment:

| Handle/state | Mutating evidence | Safe Go interpretation |
|---|---|---|
| `SidereonNtripMachine` | `new`, `push`, `finish`, `reset`, and GGA generation at `24332–24358` | Exclusive access for all state-changing calls and free; a mutex may serialize a shared wrapper, but concurrent mutation must not be assumed safe. |
| `SidereonNmeaAccumulator` | `push`/`finish`/summary family at `24222–24257` | Serialize pushes, finish, reads that depend on accumulated state, and free. |
| `SidereonRtcmLockTimeTracker` | `observe` advances history and `reset` clears it at `27230–27257` | Serialize observe/reset/free. |
| SBAS/SSR correction stores | `sidereon_sbas_store_ingest` at `28977–28987`; `sidereon_ssr_store_ingest_messages` at `32025–32037` | Serialize ingestion and free; read-only correction queries are safe only while no mutation/free races. |
| `SidereonFusionFilter` and `SidereonTrackFilter` | Fusion updates at `22192–22336`; track predict/update at `33090–33232` | Serialize predict/update/restore/configuration/free; state/covariance reads must not race mutation. |
| Exact-cache owner/cache | Cross-process lock and owner lifecycle at `21305–21345`, owner publish/close semantics at `21347–21401` | Treat owner publish/heartbeat/close as exclusive and obey the documented transaction lifecycle. The cross-process semantics are sufficient for the Python exact-cache behavior. |
| Mmap terrain/artifact verification | Verification is explicitly mutable at `23988–23994` and artifact verification at `25773–25781` | Serialize verify/free with all reads that depend on digest provenance. |
| Broadcast ephemeris preference | Setter at `18703–18712` | Serialize preference mutation with reads/solve calls unless the Go wrapper owns a separate immutable snapshot. |

Thus the header is sufficient to define safe **conservative** Go behavior, but
not sufficient to claim that mutable handles are `Sync` or safe for concurrent
method calls. A public Go API that exposes such handles must make exclusive
ownership or internal serialization part of its documented behavior. Free is
never concurrent with any use, immutable or mutable.

## Final gate decision

The committed C ABI is a strong foundation: the header exposes 1,402 entry
points, typed value records, opaque handle lifecycles, output-size contracts,
thread-local errors, exact-cache coordination, NTRIP byte state, and broad
routes for the README’s major numerical families. Many apparent mismatches are
legitimate aliases, result shaping, or caller-owned transport/cache work.

That breadth does not satisfy the authoritative question. The five missing
6-by-6 covariance helpers alone are independently sufficient and indisputable;
the six directly verified parser gaps provide an additional confirmed basis.
All eleven are documented public behavior with no equivalent committed C
route, and numeric/modeling or parser behavior cannot be assigned to Go under
the gate rules. The calendar, rich-handle,
clock, RTK arc, RINEX frequency, LNAV, and DTED findings remain explicitly
labelled additional candidates above; none is required for this decision.

The exhaustive 1,402-row coverage map and this review answer different
questions: the map accounts for every C declaration, while this review tests
whether the documented Python surface can be expressed through those
declarations. They are therefore not contradictory.

**Independent gate result: STOP.**
