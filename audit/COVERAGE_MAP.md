# sidereon-go coverage map

Gate result: STOP

This is the first coverage gate for the proposed public Go binding. The audit is static and is based only on the five public sibling repositories under ../../repos, each verified at tag v1.1.1. No sibling repository was edited, and no private/downstream context was inspected.

The C declaration inventory is complete, but the documented-parity gate is STOP. The authoritative brief requires stopping before implementation when any documented sibling capability cannot be expressed through the public C ABI without changing `sidereon-c` or adding a prohibited pure-Go numeric/parser implementation. The confirmed gaps are recorded below.

## Gate decision

STOP. The five public Python 6x6 covariance helpers and six public parser helpers in the hard-stop matrix below are real sibling functionality backed by public `sidereon-core` functions, but no equivalent public v1.1.1 C ABI route exists. The complete C table remains an inventory of the 1,402 C declarations; it is not evidence that the sibling surface is complete.

The propagated-covariance handle API is not a substitute for the documented 6x6 state-covariance transforms, unit scaling, or arbitrary PSD interpolation.

No Go implementation, archive pipeline, tests, tag, or publication may begin under the brief. Resumption requires a same-version public `sidereon-c` release/header exposing equivalent ABI routes for every required documented sibling capability, or an explicit change to the authoritative scope.

## Authoritative evidence

| Item | Evidence | Audit interpretation |
| --- | --- | --- |
| C header | ../../repos/sidereon-c/bindings/c/include/sidereon.h, tag v1.1.1, 33,803 lines | Generated committed ABI; the declaration scan below finds 1,402 exported sidereon_* functions. |
| C implementation | ../../repos/sidereon-c/bindings/c/src, tag v1.1.1 | Rust export scan also finds 1,402 pub extern C functions, matching the header. |
| Static library | ../../repos/sidereon-c/bindings/c/Cargo.toml:10-15; README.md:8-13 | The crate type is cdylib plus staticlib named sidereon; cgo can link the bundled static archive. |
| Python surface | ../../repos/sidereon-python/python/sidereon/__init__.pyi and auxiliary .pyi files, tag v1.1.1 | The documented typed surface is the parity baseline, including data/distribution/exact-cache/NTRIP modules. |
| Header contract | sidereon.h:1-33, 173-215 | Ownership, thread-local errors, status semantics, and variable-output copy/query rules are binding wrapper constraints. |

## Proposed Go organization and ABI policy

The following organization is planning context only and is blocked by the STOP: positioning (SP3, SPP, RTK, PPP, broadcast, observations), astro (SGP4, frames, passes, conjunction), distribution (GNSS data distribution and sans-I/O protocol helpers), geodesy, errormetrics, and a small core/support layer for status/version and format or compatibility pieces that do not fit those five domains. The one-to-one table uses provisional bridge names such as positioning.SP3Load; no façade work may start until the gate resumes.

All fallible C calls return a Go typed error. Proposed common contract:

- core.Status is the named Go representation of the C status enum: OK, NULL_POINTER, INVALID_ARGUMENT, INVALID_TOKEN, SP3_PARSE, SOLVE, PANIC, and TIMEOUT (sidereon.h:173-215).
- core.StatusError carries both Status and Message. Message is the sidereon_status_message text plus the thread-local sidereon_last_error_message detail when present. OK is never returned as an error.
- The wrapper captures the last-error text on the same OS thread as the fallible C call. Because the C message is thread-local (sidereon.h:21-23; lib.rs:416-445), a cgo bridge must lock the OS thread for the call-and-read sequence or provide an equivalent same-thread helper. This is an implementation constraint, not an ABI change.
- Handle-owning Go types are non-copying wrappers around opaque C pointers. They expose explicit Close, clear their pointer exactly once, and install a `runtime.AddCleanup` backstop. A failed constructor owns no handle, per sidereon.h:24-28.
- A Go wrapper never stores a Go pointer in C. []byte, []string, numeric slices, and structured observation arrays are copied into temporary C-owned memory before the call; variable outputs use the C query-then-copy rule (sidereon.h:29-33; lib.rs:739-754, 847-867).
- C read-only handles may be shared while alive. Go serializes mutation and Close; callers must not use a handle concurrently with Close. The concurrency gap is documented below.
- Data transport is Go-owned: net/http, TCP/TLS, reconnect, cache, decompression, and file I/O stay in Go. C receives bytes and returns parsed/validated values or protocol state; no Rust socket or callback surface is proposed.

### Representative proposed signatures (facade, not implementation)

| Package | Proposed signature | Ownership/result rule |
| --- | --- | --- |
| positioning | LoadSP3(data []byte) (*SP3, error); LoadExactSP3(data []byte, req ExactSP3Request) (*SP3, error) | Copies input bytes; SP3 owns C handle; Close required. |
| positioning | (s *SP3) State(satellite string, epoch Time) (SP3State, error); (s *SP3) Epochs() ([]Time, error) | Returned values/slices are Go-owned copies. |
| positioning | SolveSPP(in SPPInput) (SPPSolution, error); SolveRTKFloat(in RTKInput) (RTKFloatSolution, error); SolvePPPFloat(in PPPInput) (PPPFloatSolution, error) | Input observation/correction slices copied at FFI boundary; returned solution handles or values follow matching C ownership. |
| positioning | ParseRINEXObs(data []byte) (*RINEXObs, error); ParseRINEXNav(data []byte) (*BroadcastEphemeris, error); ParseRTCM(data []byte) ([]RTCMMessage, error) | Go owns bytes and decoded copies; parser handles, if any, expose Close. |
| astro | ParseTLE(line1, line2 string) (TLE, error); (t TLE) Propagate(epochs []Time) ([]State, error); FindPasses(...) ([]Pass, error) | TLE/pass handles have Close when backed by C; arrays are copied. |
| astro | TransformFrame(x State, from, to Frame, epoch Time) (State, error); FindTCA(...) ([]Conjunction, error) | Value results; conjunction/CDM handles Close where C allocates them. |
| distribution | Fetch(ctx context.Context, req ProductRequest) (Product, error); Acquire(ctx context.Context, req ProductRequest) (AcquiredProduct, error) | Entire network/cache lifecycle is Go-owned; C validates/decodes bytes and exact metadata only. |
| distribution | NewNTRIPClient(cfg NTRIPConfig) (*NTRIPClient, error); (c *NTRIPClient) Stream(ctx context.Context, w io.Writer) error | Go owns socket/TLS/reconnect/GGA; feeds received bytes to sidereon_ntrip_* and sidereon_rtcm_*. |
| geodesy | GeodeticToECEF(p Geodetic) (ECEF, error); ECEFToGeodetic(p ECEF) (Geodetic, error); GeodesicInverse(a, b Geodetic) (GeodesicResult, error) | Values are copied; no C-held Go pointers. |
| errormetrics | ComputeDOP(obs []Observation, state ReceiverState) (DOP, error); RAIM(...)(RAIMResult, error); ComputeAllanDeviations(...)(AllanCurves, error) | Input/output arrays are copied and status errors retain status plus text. |

### Required grouping anchors

| Domain | Capability | Proposed Go surface | C coverage anchor |
| --- | --- | --- | --- |
| positioning | SP3 / precise | LoadSP3([]byte), LoadExactSP3([]byte, ExactRequest), MergeSP3([]*SP3, MergeOptions); SP3.Close; copied epoch/state/clock accessors | sidereon_sp3_*, sidereon_precise_*, sidereon_solve_spp*; handle transfer/free is explicit. |
| positioning | SPP / DGNSS | SolveSPP(SPPInputs), SolveSPPBatch, SolveDGNSS; value results with []Observation copied at call boundary | sidereon_spp_*, sidereon_solve_spp*, sidereon_dgnss_*. |
| positioning | RTK / PPP | SolveRTKFloat/Fixed, SolvePPPFloat/Fixed; correction/arc/session handles Close | sidereon_rtk_*, sidereon_solve_rtk_*, sidereon_ppp_*, sidereon_solve_ppp_*. |
| positioning | broadcast / observations | ParseRINEXNav, ParseRINEXObs, ParseRTCM, ParseNMEA, broadcast ephemeris and observation value types | sidereon_broadcast_*, sidereon_rinex_*, sidereon_rtcm_*, sidereon_nmea_*, sidereon_observation_*. |
| astro | SGP4 / frames | ParseTLE, Propagate, PropagateBatch, frame catalog and explicit frame transforms | sidereon_tle_*, sidereon_sgp4_*, sidereon_frame_*, sidereon_frame_catalog_*. |
| astro | passes / conjunction | FindPasses, FindTCA, ScreenConjunctions, CDM/OMM/OEM/OPM codecs | sidereon_pass_*, sidereon_tca_*, sidereon_cdm_*, sidereon_omm_*, sidereon_oem_*, sidereon_opm_*. |
| distribution | GNSS data distribution | Go-owned Fetch/Acquire/cache APIs; C bytes/metadata/exact-cache validation; NTRIP TCP/TLS client in Go | sidereon_data_*, sidereon_exact_cache_*, sidereon_ntrip_* sans-I/O machine, sidereon_rtcm_*, sidereon_sbas_*, sidereon_ssr_*. |
| geodesy | coordinates / Earth models | Geodetic↔ECEF, geodesic, geoid, DTED/terrain, antenna/ATX, time scales | sidereon_geodetic_*, sidereon_geodesic_*, sidereon_geoid_*, sidereon_dted_*, sidereon_terrain_*, sidereon_time_*. |
| error-metrics | error, quality, estimation | Typed metrics/results for residuals, DOP, covariance, RAIM/FDE, clock, reliability, NIS and filters | sidereon_error_metrics_*, sidereon_dop_*, sidereon_covariance_*, sidereon_raim_*, sidereon_fde_*, sidereon_nis_*, sidereon_estimate_*. |
| support | ABI/core and remaining formats | core.Status, core.Version, internal error-text capture, and narrowly scoped format/legacy compatibility adapters | Every remaining sidereon_* row below remains accounted for; no C export is silently dropped. |

## Python documented-surface parity matrix

Inventory basis: the package implementation declares 1,194 unique __all__ entries in __init__.py:1134-2530, while __init__.pyi contains 701 class declarations and 465 def declarations (464 unique function names; raim is overloaded). Auxiliary stubs add 81 classes and 62 functions across data.pyi, distribution.pyi, exact_cache.pyi, exact_sp3.pyi, and ntrip.pyi. The matrix is capability-grouped so overloaded methods, properties, enum values, and presentation aliases are not falsely counted as separate numerical engines.

| Python documented capability/symbol family | Python evidence | C ABI expression | Planned Go API and disposition | Gate |
| --- | --- | --- | --- | --- |
| SP3 load, exact validation, interpolation, merge, precise products | __init__.pyi:4958-4963 and SP3/precise declarations throughout; exact_sp3.pyi:43-47 | sidereon_sp3_load, sidereon_sp3_load_exact, sidereon_sp3_merge, sidereon_sp3_*, sidereon_exact_* and precise_* | positioning.LoadSP3, LoadExactSP3, SP3 methods, MergeReport; Go byte input and copied outputs | Expressible |
| SPP, batch SPP, Doppler velocity, static, DGNSS, solution accessors | __init__.pyi:5101-5134, 6783 onward | sidereon_spp_*, sidereon_solve_spp*, sidereon_spp_*, sidereon_static_*, sidereon_dgnss_*, observation/observable functions | positioning.SolveSPP, SolveSPPBatch, SolveDGNSS, Static; map returned handles/accessors | Expressible |
| RTK float/fixed, arcs, ionosphere, wide-lane, RINEX workflows | __init__.pyi:5139 onward and RTK API declarations | sidereon_solve_rtk_*, sidereon_rtk_*, sidereon_combination_*, sidereon_ionex_* | positioning.RTK package with explicit session/arc Close and copied measurement slices | Expressible |
| PPP float/fixed, corrections, auto options, products | __init__.pyi:5145 onward and PPP/correction declarations | sidereon_solve_ppp_*, sidereon_ppp_*, sidereon_precise_* | positioning.PPP package; corrections and solution handles map to Close | Expressible |
| Broadcast ephemeris, GPS/Galileo/GLONASS/CNAV, RINEX navigation and comparisons | __init__.pyi:6287-6294 and broadcast names | sidereon_broadcast_ephemeris_parse_nav and related record accessors cover the usable broadcast store, but no lenient NAV result or raw GLONASS-record parser route exists | positioning.Broadcast and ParseRINEXNav can cover the existing C route; the documented lenient/GLONASS helpers are STOP gaps below | STOP: parser gaps |
| Observations, RINEX observation, carrier/signal, NMEA, RTCM measurements | __init__.pyi:6296 onward, 6691 onward, 9741 onward | sidereon_observation_*, observables_*, rinex_*, carrier_*, signal_*, nmea_*, rtcm_* | positioning.Observations and distribution.RTCM; decoded slices are copied | Expressible |
| TLE, SGP4, batch propagation, decay, Kepler/orbit utilities | __init__.pyi:5068-5083 and orbit/propagation families | sidereon_tle_*, sidereon_sgp4_*, sidereon_propagate_*, orbit_*, decay_*, reduced_* | astro.SGP4, Astro orbit functions; C handles Close, values copied | Expressible |
| Frames, frame catalog, time scales, Earth orientation, Sun/Moon | __init__.pyi:463-487, 4934-4954, plus frame/time/body families | sidereon_frame_*, frame_catalog_*, time/timescale/leap/ut1, sun_*, moon_*, nutation/precession | astro.Frames and geodesy.Time; direct C transforms where present, Go calendar convenience otherwise | Expressible |
| Passes, visibility, geometry, ground tracks, eclipses | __init__.pyi: pass/visibility/geometry families; package examples __init__.py:15-31 | sidereon_pass_*, sidereon_sp3_geometry_*, sidereon_tle_find_passes, visible_*, eclipse_*, ground_* | astro.Passes and geometry; input/output arrays copied | Expressible |
| Conjunction screening, TCA, covariance propagation, CDM/OMM/OEM/OPM | __init__.pyi:5609-5610, 8764-8798, 13329-13359 | sidereon_tca_*, codec families, sidereon_covariance_transport, sidereon_propagate_covariance, and the 3x3 sidereon_rtn_to_eci_covariance | astro.Conjunction and covariance propagation can use existing routes, but the documented 6x6 unit/interpolation/frame helpers are STOP gaps below | STOP: 6x6 covariance gaps |
| Geodesy, geodetic time series, network field, MIDAS, geofences, geoid, terrain | __init__.pyi:4934-4954 and geodetic families through phase/domain additions | sidereon_geodetic_*, geodesic_*, geoid_*, geofence_*, dted_*, terrain_*, velocity_* | geodesy package; C numeric engine, Go file/byte adapters for DTED | Expressible |
| Atmosphere, troposphere, ionosphere, RF, antenna/ANTEX | __init__.pyi: RF/atmosphere/antenna families | sidereon_atmosphere_*, tropo_*, iono_*, rf_*, antenna_*, antex_* | geodesy and positioning RF/antenna subpackages; values and parsed tables copied | Expressible |
| Error metrics, residuals, DOP, covariance, reliability, clock/allan | __init__.pyi:4990, 7159-7183 and metric families | sidereon_error_metrics_*, residual_*, dop_*, covariance_*, reliability_*, clock_*, sidereon_clock_compute_allan_deviations; no public C route for the five documented 6x6 helpers | errormetrics can use existing metrics, but cannot supply the missing 6x6 unit/frame/interpolation operations without a prohibited implementation | STOP: 6x6 covariance gaps |
| RAIM/FDE, robust solves, chi-square, normality, filters, NIS | __init__.pyi:7159-7183, 11961 onward and estimation families | sidereon_raim_*, fde_*, chi2_*, normal_*, alpha_*, beta_*, hessian_*, kalman_*, nis_*, ewma_*, cfar_* | errormetrics.RAIM/FDE/Estimation; no numeric reimplementation in Go | Expressible |
| IOD, orbit fitting, forces, drag, Lambert, relative/CW | __init__.pyi: orbit/iod/force/relative families | sidereon_iod_*, orbit_fit_*, force_*, drag_*, lambert_*, relative_*, cw_* | astro.Orbit and propagation helpers | Expressible |
| SBAS, SSR, RTCM scan/decode/encode, NMEA | __init__.pyi:903, 9741, protocol families | sidereon_sbas_block_decode and related SBAS block/store routes exist, but no public C line-oriented EMS/RTKLIB parser | distribution protocol codecs/state; the two documented line parsers remain STOP gaps below | STOP: parser gaps |
| NTRIP client, GGA feed, sourcetable, reconnect and stream_into | ntrip.pyi:1-86; __init__.py:920 | sidereon_ntrip_bytes, ntrip_machine_*, ntrip_request_*, ntrip_sourcetable_* provide sans-I/O protocol state and bytes; no C socket is required | distribution.NTRIPClient uses Go TCP/TLS/HTTP/reconnect and feeds C state; protocol errors map to Go errors | Go-owned convenience; C state expressible |
| GNSS data metadata/catalog, product URLs, cache and acquisition | data.pyi:1-490; distribution.pyi:1-185 | sidereon_data_*, exact_cache_*, precise_artifact_*, sourced_*, space_weather_*, navcen_* express metadata/validation/cache state; network is intentionally outside C | distribution.Fetch/Acquire, metadata, cache, decompression, checksums and file APIs are Go-owned conveniences around C byte validation | Go-owned convenience |
| exact_sp3 and exact-cache lifecycle/locking | exact_sp3.pyi:1-47; exact_cache.pyi:1-70 | sidereon_sp3_exact_request_*, sidereon_exact_cache_* and validation/status functions | distribution.ExactSP3 and ExactCache with Go filesystem locks; C handles Close; status preserved | Expressible/Go-owned I/O |
| Python paths, file-like inputs, JSON/XML/KVN codecs, enum/dataclass properties, repr/to_dict and aliases | __init__.py:1-43, 2258-2274; .pyi declarations | Byte/string parsers and C codecs cover the applicable numeric/content semantics; path, serialization, enum naming and aliases are Go-owned presentation | Go accepts []byte/io.Reader/path adapters where appropriate; typed structs, String/Marshal methods and compatibility aliases | Presentation/transport convenience |

The remaining Expressible rows are not affected by the confirmed gaps. Python network/file/path behavior remains Go-owned where the brief permits it, but that decomposition does not cure the numeric and parser gaps in the hard-stop matrix.

## Hard-stop capability matrix

The exact C-symbol absence scan for the capabilities below returned no matches in either the generated header or the Rust C binding source. The closest C surfaces are listed to make the non-equivalence explicit.

| Confirmed Python capability | Python/core evidence | Closest public C surface | Why the C surface is not equivalent |
| --- | --- | --- | --- |
| `covariance6_km_to_m` | `../../repos/sidereon-python/python/sidereon/__init__.pyi:13376-13378`; implementation delegates at `../../repos/sidereon-python/src/covariance.rs:558-567`; core function is public at `../../repos/sidereon/crates/sidereon-core/src/astro/covariance.rs:186-189` | `sidereon_covariance_transport` (`../../repos/sidereon-c/bindings/c/include/sidereon.h:20005-20011`) and `sidereon_propagate_covariance` (`../../repos/sidereon-c/bindings/c/include/sidereon.h:25801-25806`) | Both propagate/transport a covariance through supplied dynamics/process-noise inputs. Neither exposes the documented km-to-m 6x6 unit scaling. |
| `covariance6_m_to_km` | `../../repos/sidereon-python/python/sidereon/__init__.pyi:13379-13381`; implementation `../../repos/sidereon-python/src/covariance.rs:570-579`; public core `../../repos/sidereon/crates/sidereon-core/src/astro/covariance.rs:191-193` | `sidereon_covariance_transport` / `sidereon_propagate_covariance` as above | A propagated covariance handle or transport sequence cannot express the inverse documented 6x6 unit conversion. |
| `interpolate_covariance6` | `../../repos/sidereon-python/python/sidereon/__init__.pyi:13382-13384`; implementation `../../repos/sidereon-python/src/covariance.rs:582-594`; public core `../../repos/sidereon/crates/sidereon-core/src/astro/covariance.rs:204-219` (`interpolate_covariance_psd`) | `sidereon_covariance_transport` (`../../repos/sidereon-c/bindings/c/include/sidereon.h:20005-20011`) | Transport applies an STM/process-noise sequence to one covariance; it is not arbitrary PSD-safe interpolation between two same-frame 6x6 covariances. |
| `eci_to_rtn_covariance6` | `../../repos/sidereon-python/python/sidereon/__init__.pyi:13385-13387`; implementation `../../repos/sidereon-python/src/covariance.rs:597-606`; public core `../../repos/sidereon/crates/sidereon-core/src/astro/covariance.rs:161-169` | `sidereon_rtn_to_eci_covariance` (`../../repos/sidereon-c/bindings/c/include/sidereon.h:28406-28412`; `../../repos/sidereon-c/bindings/c/src/orbit.rs:1-45`) plus propagated covariance surfaces | The existing RTN function is expressly 3x3. Propagated-covariance handles do not expose this 6x6 ECI-to-RTN transform. |
| `rtn_to_eci_covariance6` | `../../repos/sidereon-python/python/sidereon/__init__.pyi:13388-13390`; implementation `../../repos/sidereon-python/src/covariance.rs:610-619`; public core `../../repos/sidereon/crates/sidereon-core/src/astro/covariance.rs:174-182` | `sidereon_rtn_to_eci_covariance` (`../../repos/sidereon-c/bindings/c/include/sidereon.h:28406-28412`; `../../repos/sidereon-c/bindings/c/src/orbit.rs:1-45`) plus propagated covariance surfaces | The existing RTN function accepts and returns nine doubles, not a 6x6 state covariance; propagation is not a substitute for the documented inverse transform. |
| `parse_rinex_nav_lenient` | `../../repos/sidereon-python/python/sidereon/__init__.pyi:13290`; implementation delegates to imported public `parse_nav_lenient` at `../../repos/sidereon-python/src/rinex.rs:31-35,2821-2824`; public core `../../repos/sidereon/crates/sidereon-core/src/rinex_nav/mod.rs:883-891` | `sidereon_broadcast_ephemeris_parse_nav` (`../../repos/sidereon-c/bindings/c/include/sidereon.h:18660-18664`; `../../repos/sidereon-c/bindings/c/src/broadcast.rs:33-77`) and `sidereon_rinex_lint_nav` (`../../repos/sidereon-c/bindings/c/include/sidereon.h:26632-26634`) | The broadcast parser returns the usable store and the lint report is not a `NavParse` result with parsed records plus skipped malformed blocks. No C route exposes the lenient result. |
| `parse_sbas_ems_lines` | `../../repos/sidereon-python/python/sidereon/__init__.pyi:12553`; implementation delegates at `../../repos/sidereon-python/src/sbas_ssr.rs:13-17,2224-2229`; public core `../../repos/sidereon/crates/sidereon-core/src/sbas/format.rs:17-25` | `sidereon_sbas_block_decode` (`../../repos/sidereon-c/bindings/c/include/sidereon.h:28743-28747`; `../../repos/sidereon-c/bindings/c/src/sbas.rs:290-316`) | C decodes one raw SBAS byte block. It has no text line parser producing timestamped `SbasLogBlock` records. |
| `parse_sbas_rtklib_lines` | `../../repos/sidereon-python/python/sidereon/__init__.pyi:12554`; implementation delegates at `../../repos/sidereon-python/src/sbas_ssr.rs:15-17,2232-2237`; public core `../../repos/sidereon/crates/sidereon-core/src/sbas/format.rs:27-35` | `sidereon_sbas_block_decode` as above | The one-block byte decoder does not parse RTKLIB text lines or return their epoch/satellite/form metadata. |
| `parse_rinex_glonass_records` | `../../repos/sidereon-python/python/sidereon/__init__.pyi:6293`; implementation delegates to imported public `parse_glonass` at `../../repos/sidereon-python/src/rinex.rs:31-35,2904-2910`; public core `../../repos/sidereon/crates/sidereon-core/src/rinex_nav/mod.rs:1478-1481` | `sidereon_broadcast_ephemeris_parse_nav` plus `sidereon_broadcast_ephemeris_records` (`../../repos/sidereon-c/bindings/c/include/sidereon.h:18660-18680`) or RTCM 1020 accessors | The broadcast record struct (`../../repos/sidereon-c/bindings/c/include/sidereon.h:4484-4500`) has Keplerian record fields, not GLONASS PZ-90.11 state/clock terms; RTCM 1020 is a different wire format. No C route parses RINEX GLONASS records. |
| `parse_rinex_iono_corrections` | `../../repos/sidereon-python/python/sidereon/__init__.pyi:6294`; implementation delegates at `../../repos/sidereon-python/src/rinex.rs:2914-2917`; public core `../../repos/sidereon/crates/sidereon-core/src/rinex_nav/mod.rs:1211-1215` | Caller-populated Klobuchar fields in `SidereonSppInputs` (`../../repos/sidereon-c/bindings/c/include/sidereon.h:5894-5906`) and `sidereon_klobuchar_native` (`../../repos/sidereon-c/bindings/c/include/sidereon.h:23477-23491`) | These consume supplied coefficients or evaluate a correction; neither parses RINEX `IONOSPHERIC CORR`/RINEX 4 `> ION` text into the documented `IonoCorrections` value. |
| `parse_rinex_leap_seconds` | `../../repos/sidereon-python/python/sidereon/__init__.pyi:6295`; implementation delegates at `../../repos/sidereon-python/src/rinex.rs:2920-2922`; public core `../../repos/sidereon/crates/sidereon-core/src/rinex_nav/mod.rs:1380-1384` | `sidereon_leap_seconds` (`../../repos/sidereon-c/bindings/c/include/sidereon.h:23593-23600`) and table metadata/source functions | The C function looks up built-in TAI−UTC for a calendar date; it does not parse the NAV header's `LEAP SECONDS` field and cannot preserve its absent/present result semantics. |

These 11 capabilities trigger the brief's ABI stop condition. They are not rows in the C table because they are missing C exports; adding them to Go would require either a changed public `sidereon-c` ABI or prohibited pure-Go numeric/parser implementations.

## Explicit exclusions and why they are safe

- No C export is excluded from the audit. The complete one-row-per-export table follows.
- Raw *_free functions are not public free-standing Go functions. Each is the matching owning type's Close method plus runtime.AddCleanup backstop; this preserves the C ownership contract and prevents users from passing opaque pointers around.
- sidereon_last_error_message is an internal error-capture operation, represented in core.StatusError rather than exposed as an unsafe buffer API. The message text is still retained on every failure.
- sidereon_status_message, sidereon_version, and sidereon_version_string become core.Status.String and core.Version APIs. They are not omitted.
- Go network/socket/callback implementations are deliberately outside C because the binding requirement is Go-owned transport. sidereon_ntrip_* is used as a sans-I/O parser/state machine, not as a socket API.
- Pure Python presentation, calendar, path, JSON/XML/KVN, cache, checksum, decompression, and HTTP conveniences are planned as Go standard-library or Go-owned APIs. They add no modeling or numeric behavior and do not justify an ABI patch.

## Concurrency and thread-safety findings

The header preamble (sidereon.h:7-19) is the only authoritative general handle statement: newly allocated opaque handles transfer through out-parameters; matching _free releases them; NULL free is a no-op; opaque handles are read-only shareable after creation; concurrent free invalidates the handle and borrowed pointers. lib.rs:416-445 confirms last-error storage is thread-local and lib.rs:799-803 confirms NULL-safe boxed free. README.md:87-93 states reader outputs are copied/caller-owned and every handle must be released.

Resolved wrapper policy: share immutable/read-only handles, serialize mutation and Close per handle, and reject use after Close. No C handle may outlive the Go owner or retain Go memory.

Unresolved contract gap for supervisor review: the header does not classify each handle or operation as immutable, mutable-but-internally-synchronized, or externally synchronized. It also does not promise concurrent read-call safety for every read-only handle, nor define races between accessors and Close. The Go binding must conservatively serialize mutable handles and Close and document read-sharing only for handles whose C docs explicitly permit it. The thread-local error requires same-OS-thread status/message capture in cgo. These are wrapper constraints and documentation gaps, not evidence of an unexpressible Python capability.

## Complete exported C-function map

Every row below is a declaration found by the exact scan documented in AUDIT_EVIDENCE.md. The provisional bridge name is a one-to-one implementation anchor, not an invitation to expose unsafe C naming. The public façade uses the idiomatic signatures above. Fallible calls apply StatusError; lifecycle rows apply Close; variable outputs apply query-then-copy.

| # | Header line | C export | Proposed Go bridge/facade | Disposition |
| ---: | ---: | --- | --- | --- |
| 1 | 18143 | sidereon_absolute_from_relative | support.AbsoluteFromRelative | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 2 | 18147 | sidereon_almanac_lunar_solar_eclipses | astro.AlmanacLunarSolarEclipses | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 3 | 18157 | sidereon_almanac_meridian_transits | astro.AlmanacMeridianTransits | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 4 | 18170 | sidereon_almanac_moon_phases | astro.AlmanacMoonPhases | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 5 | 18180 | sidereon_almanac_planetary_events | astro.AlmanacPlanetaryEvents | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 6 | 18192 | sidereon_almanac_seasons | astro.AlmanacSeasons | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 7 | 18207 | sidereon_alpha_beta_filter_step | errormetrics.AlphaBetaFilterStep | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 8 | 18219 | sidereon_alpha_beta_steady_state_gains | errormetrics.AlphaBetaSteadyStateGains | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 9 | 18229 | sidereon_angular_separation_coords_deg | support.AngularSeparationCoordsDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 10 | 18241 | sidereon_angular_separation_deg | support.AngularSeparationDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 11 | 18250 | sidereon_antenna_free | geodesy.Antenna.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 12 | 18261 | sidereon_antenna_pco | geodesy.AntennaPco | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 13 | 18275 | sidereon_antenna_pcv | geodesy.AntennaPcv | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 14 | 18292 | sidereon_antex_antenna | geodesy.AntexAntenna | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 15 | 18302 | sidereon_antex_antenna_count | geodesy.AntexAntennaCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 16 | 18314 | sidereon_antex_encode | geodesy.AntexEncode | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 17 | 18327 | sidereon_antex_free | geodesy.Antex.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 18 | 18337 | sidereon_antex_parse | geodesy.AntexParse | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 19 | 18348 | sidereon_araim | support.Araim | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 20 | 18358 | sidereon_araim_allocation_lpv_200 | support.AraimAllocationLpv200 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 21 | 18368 | sidereon_araim_result_fault_mode_excluded_sats | support.AraimResultFaultModeExcludedSats | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 22 | 18382 | sidereon_araim_result_fault_modes | support.AraimResultFaultModes | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 23 | 18393 | sidereon_araim_result_free | support.AraimResult.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 24 | 18402 | sidereon_araim_result_summary | support.AraimResultSummary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 25 | 18412 | sidereon_atmosphere_input_default | geodesy.AtmosphereInputDefault | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 26 | 18425 | sidereon_atmosphere_nrlmsise00 | geodesy.AtmosphereNrlmsise00 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 27 | 18428 | sidereon_beta_angle_deg | errormetrics.BetaAngleDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 28 | 18432 | sidereon_beta_angle_from_state_deg | errormetrics.BetaAngleFromStateDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 29 | 18437 | sidereon_bias_set_code_dsb_seconds | errormetrics.BiasSetCodeDsbSeconds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 30 | 18445 | sidereon_bias_set_code_osb_seconds | errormetrics.BiasSetCodeOsbSeconds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 31 | 18452 | sidereon_bias_set_free | errormetrics.BiasSet.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 32 | 18454 | sidereon_bias_set_mode | errormetrics.BiasSetMode | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 33 | 18458 | sidereon_bias_set_phase_osb_cycles | errormetrics.BiasSetPhaseOsbCycles | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 34 | 18465 | sidereon_bias_set_record | errormetrics.BiasSetRecord | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 35 | 18469 | sidereon_bias_set_record_count | errormetrics.BiasSetRecordCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 36 | 18472 | sidereon_bias_set_skipped_record_count | errormetrics.BiasSetSkippedRecordCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 37 | 18475 | sidereon_bias_set_warning_count | errormetrics.BiasSetWarningCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 38 | 18478 | sidereon_bias_sinex_load | errormetrics.BiasSinexLoad | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 39 | 18480 | sidereon_bias_sinex_load_lossy | errormetrics.BiasSinexLoadLossy | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 40 | 18482 | sidereon_bias_sinex_parse | errormetrics.BiasSinexParse | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 41 | 18486 | sidereon_bias_sinex_parse_lossy | errormetrics.BiasSinexParseLossy | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 42 | 18505 | sidereon_bounded_ils_search | support.BoundedIlsSearch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 43 | 18524 | sidereon_broadcast_comparison_compare | positioning.BroadcastComparisonCompare | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 44 | 18544 | sidereon_broadcast_comparison_compare_window | positioning.BroadcastComparisonCompareWindow | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 45 | 18557 | sidereon_broadcast_comparison_free | positioning.BroadcastComparison.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 46 | 18564 | sidereon_broadcast_comparison_overall | positioning.BroadcastComparisonOverall | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 47 | 18574 | sidereon_broadcast_comparison_satellite | positioning.BroadcastComparisonSatellite | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 48 | 18585 | sidereon_broadcast_comparison_satellite_count | positioning.BroadcastComparisonSatelliteCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 49 | 18596 | sidereon_broadcast_eccentric_anomaly | positioning.BroadcastEccentricAnomaly | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 50 | 18610 | sidereon_broadcast_emission_media_batch_at_j2000_s | positioning.BroadcastEmissionMediaBatchAtJ2000S | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 51 | 18634 | sidereon_broadcast_ephemeris_free | positioning.BroadcastEphemeris.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 52 | 18645 | sidereon_broadcast_ephemeris_load_nav | positioning.BroadcastEphemerisLoadNav | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 53 | 18648 | sidereon_broadcast_ephemeris_nav_message_preference | positioning.BroadcastEphemerisNavMessagePreference | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 54 | 18660 | sidereon_broadcast_ephemeris_parse_nav | positioning.BroadcastEphemerisParseNav | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 55 | 18664 | sidereon_broadcast_ephemeris_record_cnav_correction | positioning.BroadcastEphemerisRecordCNAVCorrection | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 56 | 18670 | sidereon_broadcast_ephemeris_record_count | positioning.BroadcastEphemerisRecordCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 57 | 18673 | sidereon_broadcast_ephemeris_record_group_delay | positioning.BroadcastEphemerisRecordGroupDelay | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 58 | 18679 | sidereon_broadcast_ephemeris_records | positioning.BroadcastEphemerisRecords | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 59 | 18692 | sidereon_broadcast_ephemeris_sample | positioning.BroadcastEphemerisSample | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 60 | 18703 | sidereon_broadcast_ephemeris_select_by_issue | positioning.BroadcastEphemerisSelectByIssue | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 61 | 18711 | sidereon_broadcast_ephemeris_set_nav_message_preference | positioning.BroadcastEphemerisSetNavMessagePreference | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 62 | 18725 | sidereon_broadcast_observable_state | positioning.BroadcastObservableState | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 63 | 18740 | sidereon_broadcast_observable_states_at_j2000_s | positioning.BroadcastObservableStatesAtJ2000S | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 64 | 18756 | sidereon_broadcast_observable_states_at_shared_j2000_s | positioning.BroadcastObservableStatesAtSharedJ2000S | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 65 | 18777 | sidereon_broadcast_observables | positioning.BroadcastObservables | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 66 | 18795 | sidereon_broadcast_observables_batch | positioning.BroadcastObservablesBatch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 67 | 18810 | sidereon_broadcast_satellite_clock_offset_s | positioning.BroadcastSatelliteClockOffsetS | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 68 | 18826 | sidereon_broadcast_satellite_position_ecef | positioning.BroadcastSatellitePositionECEF | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 69 | 18839 | sidereon_broadcast_satellite_state | positioning.BroadcastSatelliteState | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 70 | 18850 | sidereon_carrier_band_label | positioning.CarrierBandLabel | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 71 | 18862 | sidereon_carrier_code_minus_carrier | positioning.CarrierCodeMinusCarrier | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 72 | 18873 | sidereon_carrier_geometry_free | positioning.CarrierGeometry.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 73 | 18881 | sidereon_carrier_melbourne_wubbena | positioning.CarrierMelbourneWubbena | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 74 | 18895 | sidereon_carrier_narrow_lane_code | positioning.CarrierNarrowLaneCode | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 75 | 18907 | sidereon_carrier_phase_meters | positioning.CarrierPhaseMeters | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 76 | 18915 | sidereon_carrier_wide_lane_cycles | positioning.CarrierWideLaneCycles | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 77 | 18929 | sidereon_carrier_wide_lane_wavelength | positioning.CarrierWideLaneWavelength | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 78 | 18936 | sidereon_cdm_free | astro.CDM.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 79 | 18945 | sidereon_cdm_numbers | astro.CDMNumbers | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 80 | 18959 | sidereon_cdm_object_state | astro.CDMObjectState | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 81 | 18974 | sidereon_cdm_object_string_field | astro.CDMObjectStringField | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 82 | 18991 | sidereon_cdm_object_velocity_covariance | astro.CDMObjectVelocityCovariance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 83 | 19002 | sidereon_cdm_parse_kvn | astro.CDMParseKVN | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 84 | 19012 | sidereon_cdm_parse_xml | astro.CDMParseXML | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 85 | 19024 | sidereon_cdm_string_field | astro.CDMStringField | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 86 | 19038 | sidereon_cdm_to_kvn | astro.CDMToKVN | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 87 | 19051 | sidereon_cdm_to_xml | astro.CDMToXML | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 88 | 19063 | sidereon_cfar_ca_false_alarm_probability | errormetrics.CFARCAFalseAlarmProbability | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 89 | 19074 | sidereon_cfar_ca_multiplier_from_pfa | errormetrics.CFARCAMultiplierFromPfa | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 90 | 19083 | sidereon_cfar_ca_pfa_from_multiplier | errormetrics.CFARCAPfaFromMultiplier | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 91 | 19093 | sidereon_cfar_ca_threshold | errormetrics.CFARCAThreshold | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 92 | 19105 | sidereon_chan_ho_initial_guess | support.ChanHoInitialGuess | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 93 | 19119 | sidereon_chi2_inv | errormetrics.Chi2Inv | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 94 | 19128 | sidereon_civil_to_gps_seconds | support.CivilToGPSSeconds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 95 | 19144 | sidereon_civil_to_j2000_seconds | geodesy.CivilToJ2000Seconds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 96 | 19159 | sidereon_clock_allan_curve | errormetrics.ClockAllanCurve | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 97 | 19172 | sidereon_clock_allan_curve_present | errormetrics.ClockAllanCurvePresent | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 98 | 19183 | sidereon_clock_allan_deviation | errormetrics.ClockAllanDeviation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 99 | 19200 | sidereon_clock_allan_deviation_curves_free | errormetrics.ClockAllanDeviationCurves.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 100 | 19208 | sidereon_clock_allan_options_init | errormetrics.ClockAllanOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 101 | 19220 | sidereon_clock_compute_allan_deviations | errormetrics.ClockComputeAllanDeviations | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 102 | 19235 | sidereon_clock_fit_power_law_noise | errormetrics.ClockFitPowerLawNoise | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 103 | 19249 | sidereon_clock_hadamard_deviation | errormetrics.ClockHadamardDeviation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 104 | 19267 | sidereon_clock_modified_adev | errormetrics.ClockModifiedAdev | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 105 | 19285 | sidereon_clock_overlapping_adev | errormetrics.ClockOverlappingAdev | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 106 | 19301 | sidereon_clock_power_law_noise_fit_coefficients | errormetrics.ClockPowerLawNoiseFitCoefficients | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 107 | 19310 | sidereon_clock_power_law_noise_fit_free | errormetrics.ClockPowerLawNoiseFit.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 108 | 19318 | sidereon_clock_power_law_noise_fit_octaves | errormetrics.ClockPowerLawNoiseFitOctaves | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 109 | 19329 | sidereon_clock_power_law_noise_fit_regions | errormetrics.ClockPowerLawNoiseFitRegions | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 110 | 19340 | sidereon_clock_power_law_noise_options_init | errormetrics.ClockPowerLawNoiseOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 111 | 19350 | sidereon_clock_power_law_noise_slopes | errormetrics.ClockPowerLawNoiseSlopes | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 112 | 19361 | sidereon_clock_time_deviation | geodesy.ClockTimeDeviation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 113 | 19378 | sidereon_closed_form_initial_guess | support.ClosedFormInitialGuess | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 114 | 19386 | sidereon_cnav_ura_ned_m | support.CNAVUraNedM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 115 | 19392 | sidereon_cnav_ura_nominal_m | support.CNAVUraNominalM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 116 | 19396 | sidereon_code_dcb_load | support.CodeDcbLoad | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 117 | 19400 | sidereon_code_dcb_load_lossy | support.CodeDcbLoadLossy | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 118 | 19404 | sidereon_code_dcb_parse | support.CodeDcbParse | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 119 | 19409 | sidereon_code_dcb_parse_lossy | support.CodeDcbParseLossy | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 120 | 19414 | sidereon_coe2eq | support.Coe2eq | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 121 | 19418 | sidereon_coe2mee | support.Coe2mee | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 122 | 19430 | sidereon_coe2rv | support.Coe2rv | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 123 | 19442 | sidereon_collision_probability | astro.CollisionProbability | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 124 | 19454 | sidereon_combination_gamma | positioning.CombinationGamma | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 125 | 19462 | sidereon_combination_ionosphere_free | positioning.CombinationIonosphere.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 126 | 19475 | sidereon_combination_ionosphere_free_phase_cycles | positioning.CombinationIonosphereFreePhaseCycles | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 127 | 19487 | sidereon_combination_ionosphere_free_phase_m | positioning.CombinationIonosphereFreePhaseM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 128 | 19505 | sidereon_combination_ionosphere_free_pseudoranges | positioning.CombinationIonosphereFreePseudoranges | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 129 | 19519 | sidereon_combination_noise_amplification | positioning.CombinationNoiseAmplification | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 130 | 19537 | sidereon_constellation_build | astro.ConstellationBuild | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 131 | 19558 | sidereon_constellation_build_at | astro.ConstellationBuildAt | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 132 | 19573 | sidereon_constellation_diff | astro.ConstellationDiff | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 133 | 19583 | sidereon_constellation_diff_activity_changed | astro.ConstellationDiffActivityChanged | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 134 | 19595 | sidereon_constellation_diff_added | astro.ConstellationDiffAdded | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 135 | 19607 | sidereon_constellation_diff_changed | astro.ConstellationDiffChanged | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 136 | 19615 | sidereon_constellation_diff_counts | astro.ConstellationDiffCounts | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 137 | 19624 | sidereon_constellation_diff_fdma_channel_changed | astro.ConstellationDiffFdmaChannelChanged | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 138 | 19636 | sidereon_constellation_diff_free | astro.ConstellationDiff.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 139 | 19644 | sidereon_constellation_diff_norad_reassigned | astro.ConstellationDiffNoradReassigned | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 140 | 19656 | sidereon_constellation_diff_removed | astro.ConstellationDiffRemoved | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 141 | 19665 | sidereon_constellation_diff_sp3_id_changed_from | positioning.ConstellationDiffSP3IdChangedFrom | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 142 | 19677 | sidereon_constellation_diff_sp3_id_changed_meta | positioning.ConstellationDiffSP3IdChangedMeta | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 143 | 19684 | sidereon_constellation_diff_sp3_id_changed_to | positioning.ConstellationDiffSP3IdChangedTo | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 144 | 19697 | sidereon_constellation_diff_svn_changed | astro.ConstellationDiffSvnChanged | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 145 | 19709 | sidereon_constellation_diff_usability_changed | astro.ConstellationDiffUsabilityChanged | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 146 | 19722 | sidereon_constellation_free | astro.Constellation.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 147 | 19730 | sidereon_constellation_galileo_prn_for_gsat | astro.ConstellationGalileoPrnForGsat | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 148 | 19740 | sidereon_constellation_glonass_fdma_channel | positioning.ConstellationGLONASSFdmaChannel | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 149 | 19750 | sidereon_constellation_glonass_slot_for_number | positioning.ConstellationGLONASSSlotForNumber | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 150 | 19763 | sidereon_constellation_gnss_sp3_id | positioning.ConstellationGNSSSP3Id | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 151 | 19779 | sidereon_constellation_record | astro.ConstellationRecord | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 152 | 19789 | sidereon_constellation_record_count | astro.ConstellationRecordCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 153 | 19802 | sidereon_constellation_to_csv | astro.ConstellationToCsv | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 154 | 19818 | sidereon_constellation_validate | astro.ConstellationValidate | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 155 | 19828 | sidereon_constellation_validate_against_sp3 | positioning.ConstellationValidateAgainstSP3 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 156 | 19843 | sidereon_constellation_validate_against_sp3_ids | positioning.ConstellationValidateAgainstSP3Ids | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 157 | 19855 | sidereon_constellation_validate_against_sp3_ids_strict | positioning.ConstellationValidateAgainstSP3IdsStrict | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 158 | 19867 | sidereon_constellation_validation_duplicate_norad_ids | astro.ConstellationValidationDuplicateNoradIds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 159 | 19882 | sidereon_constellation_validation_duplicate_prns | astro.ConstellationValidationDuplicatePrns | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 160 | 19897 | sidereon_constellation_validation_extra_sp3_ids | positioning.ConstellationValidationExtraSP3Ids | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 161 | 19910 | sidereon_constellation_validation_free | astro.ConstellationValidation.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 162 | 19921 | sidereon_constellation_validation_inactive_unusable_prns | astro.ConstellationValidationInactiveUnusablePrns | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 163 | 19932 | sidereon_constellation_validation_is_valid | astro.ConstellationValidationIsValid | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 164 | 19944 | sidereon_constellation_validation_missing_sp3_ids | positioning.ConstellationValidationMissingSP3Ids | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 165 | 19950 | sidereon_covariance_ephemeris_count | astro.CovarianceEphemerisCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 166 | 19953 | sidereon_covariance_ephemeris_covariance_at | astro.CovarianceEphemerisCovarianceAt | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 167 | 19957 | sidereon_covariance_ephemeris_free | astro.CovarianceEphemeris.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 168 | 19959 | sidereon_covariance_ephemeris_nodes | astro.CovarianceEphemerisNodes | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 169 | 19979 | sidereon_covariance_from_jacobian | errormetrics.CovarianceFromJacobian | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 170 | 19994 | sidereon_covariance_is_positive_semidefinite | errormetrics.CovarianceIsPositiveSemidefinite | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 171 | 20003 | sidereon_covariance_is_symmetric | errormetrics.CovarianceIsSymmetric | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 172 | 20005 | sidereon_covariance_transport | errormetrics.CovarianceTransport | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 173 | 20022 | sidereon_coverage_grid_access_counts | support.CoverageGridAccessCounts | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 174 | 20035 | sidereon_coverage_grid_dimensions | support.CoverageGridDimensions | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 175 | 20045 | sidereon_coverage_grid_free | support.CoverageGrid.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 176 | 20053 | sidereon_coverage_grid_look_angle | support.CoverageGridLookAngle | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 177 | 20067 | sidereon_coverage_grid_max_elevation_deg | support.CoverageGridMaxElevationDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 178 | 20080 | sidereon_coverage_grid_visible_mask | astro.CoverageGridVisibleMask | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 179 | 20095 | sidereon_coverage_look_angles | support.CoverageLookAngles | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 180 | 20113 | sidereon_crinex_decode | support.CrinexDecode | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 181 | 20130 | sidereon_crinex_encode | support.CrinexEncode | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 182 | 20137 | sidereon_cw_propagate | astro.CwPropagate | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 183 | 20142 | sidereon_cw_stm | astro.CwStm | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 184 | 20153 | sidereon_cycle_slip_options_init | support.CycleSlipOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 185 | 20172 | sidereon_data_default_sample_for_date | distribution.DataDefaultSampleForDate | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 186 | 20193 | sidereon_data_distribution_location | distribution.DataDistributionLocation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 187 | 20220 | sidereon_data_newest_published_product_json | distribution.DataNewestPublishedProductJSON | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 188 | 20244 | sidereon_data_next_issue_due_json | distribution.DataNextIssueDueJSON | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 189 | 20276 | sidereon_data_predicted_ionex_line_candidates_json | positioning.DataPredictedIonexLineCandidatesJSON | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 190 | 20294 | sidereon_data_problem_init | distribution.DataProblemInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 191 | 20310 | sidereon_data_product_identity | distribution.DataProductIdentity | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 192 | 20328 | sidereon_data_product_identity_cache_key | distribution.DataProductIdentityCacheKey | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 193 | 20334 | sidereon_data_product_solution_class | distribution.DataProductSolutionClass | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 194 | 20361 | sidereon_data_publication_listing_urls_json | distribution.DataPublicationListingUrlsJSON | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 195 | 20385 | sidereon_data_sp3_content_start_convention | positioning.DataSP3ContentStartConvention | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 196 | 20413 | sidereon_data_supported_samples | distribution.DataSupportedSamples | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 197 | 20436 | sidereon_data_validate_exact_product_set | distribution.DataValidateExactProductSet | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 198 | 20446 | sidereon_decay_config_init | astro.DecayConfigInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 199 | 20454 | sidereon_default_spp_frequency_hz | positioning.DefaultSPPFrequencyHz | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 200 | 20465 | sidereon_detect_cycle_slips | support.DetectCycleSlips | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 201 | 20480 | sidereon_dgnss_applied_corrected | positioning.DGNSSAppliedCorrected | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 202 | 20492 | sidereon_dgnss_applied_counts | positioning.DGNSSAppliedCounts | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 203 | 20501 | sidereon_dgnss_applied_dropped | positioning.DGNSSAppliedDropped | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 204 | 20511 | sidereon_dgnss_applied_free | positioning.DGNSSApplied.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 205 | 20521 | sidereon_dgnss_apply_corrections | positioning.DGNSSApplyCorrections | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 206 | 20533 | sidereon_dgnss_correction | positioning.DGNSSCorrection | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 207 | 20543 | sidereon_dgnss_corrections_count | positioning.DGNSSCorrectionsCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 208 | 20552 | sidereon_dgnss_corrections_free | positioning.DGNSSCorrections.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 209 | 20567 | sidereon_dgnss_position_solve | positioning.DGNSSPositionSolve | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 210 | 20586 | sidereon_dgnss_pseudorange_corrections | positioning.DGNSSPseudorangeCorrections | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 211 | 20600 | sidereon_dgnss_solution_baseline | positioning.DGNSSSolutionBaseline | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 212 | 20613 | sidereon_dgnss_solution_dropped_sats | positioning.DGNSSSolutionDroppedSats | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 213 | 20624 | sidereon_dgnss_solution_free | positioning.DGNSSSolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 214 | 20633 | sidereon_dgnss_solution_solution | positioning.DGNSSSolutionSolution | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 215 | 20649 | sidereon_dop | errormetrics.DOP | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 216 | 20663 | sidereon_dop_with_convention | errormetrics.DOPWithConvention | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 217 | 20677 | sidereon_doppler_range_rate_and_ratio | positioning.DopplerRangeRateAndRatio | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 218 | 20692 | sidereon_doppler_shift | positioning.DopplerShift | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 219 | 20707 | sidereon_drag_force_acceleration | astro.DragForceAcceleration | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 220 | 20716 | sidereon_drag_parameters_from_area_mass | astro.DragParametersFromAreaMass | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 221 | 20728 | sidereon_drag_parameters_from_ballistic_coefficient | astro.DragParametersFromBallisticCoefficient | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 222 | 20738 | sidereon_drag_parameters_from_bc_factor | astro.DragParametersFromBcFactor | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 223 | 20749 | sidereon_dted_interpolation_label | geodesy.DTEDInterpolationLabel | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 224 | 20761 | sidereon_dted_lookup_options_init | geodesy.DTEDLookupOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 225 | 20768 | sidereon_dted_terrain_free | geodesy.DTEDTerrain.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 226 | 20781 | sidereon_dted_terrain_height_batch_m | geodesy.DTEDTerrainHeightBatchM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 227 | 20793 | sidereon_dted_terrain_height_m | geodesy.DTEDTerrainHeightM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 228 | 20805 | sidereon_dted_terrain_height_m_with_options | geodesy.DTEDTerrainHeightMWithOptions | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 229 | 20818 | sidereon_dted_terrain_new | geodesy.DTEDTerrainNew | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 230 | 20826 | sidereon_dted_tile_free | geodesy.DTEDTile.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 231 | 20835 | sidereon_dted_tile_get_elevation | geodesy.DTEDTileGetElevation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 232 | 20846 | sidereon_dted_tile_load | geodesy.DTEDTileLoad | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 233 | 20856 | sidereon_dted_tree_to_mmap_store | distribution.DTEDTreeToMmapStore | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 234 | 20868 | sidereon_earth_angular_radius_deg | geodesy.EarthAngularRadiusDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 235 | 20870 | sidereon_eccentric_to_mean_anomaly | astro.EccentricToMeanAnomaly | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 236 | 20874 | sidereon_eccentric_to_true_anomaly | support.EccentricToTrueAnomaly | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 237 | 20887 | sidereon_ecef_to_geodetic | geodesy.ECEFToGeodetic | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 238 | 20896 | sidereon_eclipse_shadow_fraction | astro.EclipseShadowFraction | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 239 | 20906 | sidereon_eclipse_shadow_fraction_with_model | astro.EclipseShadowFractionWithModel | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 240 | 20918 | sidereon_eclipse_status | astro.EclipseStatus | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 241 | 20928 | sidereon_eclipse_status_with_model | astro.EclipseStatusWithModel | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 242 | 20938 | sidereon_egm96_15m_geoid_free | geodesy.EGM9615mGeoid.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 243 | 20947 | sidereon_egm96_15m_geoid_from_ww15mgh_dac_bytes | geodesy.EGM9615mGeoidFromWw15mghDacBytes | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 244 | 20960 | sidereon_egm96_15m_geoid_from_ww15mgh_dac_path | geodesy.EGM9615mGeoidFromWw15mghDacPath | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 245 | 20971 | sidereon_egm96_ellipsoidal_height_m | geodesy.EGM96EllipsoidalHeightM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 246 | 20984 | sidereon_egm96_orthometric_height_m | geodesy.EGM96OrthometricHeightM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 247 | 20997 | sidereon_egm96_undulation | geodesy.EGM96Undulation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 248 | 20999 | sidereon_egm96_undulations_deg | geodesy.EGM96UndulationsDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 249 | 21006 | sidereon_egm96_undulations_rad | geodesy.EGM96UndulationsRad | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 250 | 21020 | sidereon_ellipsoidal_height_m | geodesy.EllipsoidalHeightM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 251 | 21030 | sidereon_emission_media_options_init | support.EmissionMediaOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 252 | 21038 | sidereon_encounter_frame | astro.EncounterFrame | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 253 | 21053 | sidereon_encounter_plane_covariance | astro.EncounterPlaneCovariance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 254 | 21062 | sidereon_ephemeris_epoch_count | astro.EphemerisEpochCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 255 | 21074 | sidereon_ephemeris_free | astro.Ephemeris.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 256 | 21084 | sidereon_ephemeris_states | astro.EphemerisStates | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 257 | 21098 | sidereon_ephemeris_times_s | astro.EphemerisTimesS | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 258 | 21104 | sidereon_eq2coe | support.Eq2coe | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 259 | 21107 | sidereon_eq2rv | support.Eq2rv | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 260 | 21122 | sidereon_error_ellipse_2x2 | errormetrics.ErrorEllipse2x2 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 261 | 21132 | sidereon_error_metrics_error_ellipse_from_enu_m2 | errormetrics.ErrorMetricsErrorEllipseFromEnuM2 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 262 | 21142 | sidereon_error_metrics_from_ecef_covariance_m2 | errormetrics.ErrorMetricsFromECEFCovarianceM2 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 263 | 21153 | sidereon_error_metrics_from_enu_covariance_m2 | errormetrics.ErrorMetricsFromEnuCovarianceM2 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 264 | 21162 | sidereon_error_metrics_from_kinematic_solution | errormetrics.ErrorMetricsFromKinematicSolution | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 265 | 21171 | sidereon_error_metrics_from_position_covariance | errormetrics.ErrorMetricsFromPositionCovariance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 266 | 21181 | sidereon_error_metrics_horizontal_radius_at | errormetrics.ErrorMetricsHorizontalRadiusAt | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 267 | 21192 | sidereon_error_metrics_spherical_radius_at | errormetrics.ErrorMetricsSphericalRadiusAt | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 268 | 21202 | sidereon_error_metrics_vertical_radius_at | geodesy.ErrorMetricsVerticalRadiusAt | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 269 | 21213 | sidereon_estimate_decay | astro.EstimateDecay | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 270 | 21217 | sidereon_estimate_decay_with_space_weather_table | astro.EstimateDecayWithSpaceWeatherTable | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 271 | 21227 | sidereon_ewma_update | errormetrics.EWMAUpdate | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 272 | 21234 | sidereon_ewma_update_power_of_two | errormetrics.EWMAUpdatePowerOfTwo | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 273 | 21244 | sidereon_exact_cache_cleanup | distribution.ExactCacheCleanup | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 274 | 21257 | sidereon_exact_cache_entry_copy_bytes | distribution.ExactCacheEntryCopyBytes | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 275 | 21273 | sidereon_exact_cache_entry_copy_id | distribution.ExactCacheEntryCopyId | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 276 | 21287 | sidereon_exact_cache_entry_copy_path | distribution.ExactCacheEntryCopyPath | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 277 | 21297 | sidereon_exact_cache_entry_free | distribution.ExactCacheEntry.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 278 | 21302 | sidereon_exact_cache_free | distribution.ExactCache.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 279 | 21316 | sidereon_exact_cache_open | distribution.ExactCacheOpen | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 280 | 21339 | sidereon_exact_cache_open_single_flight | distribution.ExactCacheOpenSingleFlight | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 281 | 21353 | sidereon_exact_cache_owner_free | distribution.ExactCacheOwner.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 282 | 21360 | sidereon_exact_cache_owner_heartbeat | distribution.ExactCacheOwnerHeartbeat | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 283 | 21375 | sidereon_exact_cache_owner_publish | distribution.ExactCacheOwnerPublish | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 284 | 21394 | sidereon_exact_cache_publish | distribution.ExactCachePublish | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 285 | 21412 | sidereon_exact_cache_read | distribution.ExactCacheRead | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 286 | 21430 | sidereon_exact_cache_read_unlocked | distribution.ExactCacheReadUnlocked | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 287 | 21446 | sidereon_exact_cache_single_flight_options_init | distribution.ExactCacheSingleFlightOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 288 | 21455 | sidereon_fde_options_init | errormetrics.FDEOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 289 | 21465 | sidereon_fde_solution_excluded_sats | errormetrics.FDESolutionExcludedSats | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 290 | 21478 | sidereon_fde_solution_free | errormetrics.FDESolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 291 | 21485 | sidereon_fde_solution_iterations | errormetrics.FDESolutionIterations | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 292 | 21496 | sidereon_fde_solution_solution | errormetrics.FDESolutionSolution | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 293 | 21511 | sidereon_fde_solve_broadcast | positioning.FDESolveBroadcast | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 294 | 21530 | sidereon_fde_solve_spp | positioning.FDESolveSPP | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 295 | 21546 | sidereon_find_moon_elevation_crossings | astro.FindMoonElevationCrossings | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 296 | 21565 | sidereon_find_moon_transits | astro.FindMoonTransits | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 297 | 21585 | sidereon_find_tca_candidates_from_tles | astro.FindTCACandidatesFromTles | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 298 | 21609 | sidereon_find_tca_conjunctions_from_tles | astro.FindTCAConjunctionsFromTles | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 299 | 21636 | sidereon_find_tca_conjunctions_with_propagated_covariance_from_tles | astro.FindTCAConjunctionsWithPropagatedCovarianceFromTles | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 300 | 21657 | sidereon_fit_all_sp3_ecef_precise_orbits | positioning.FitAllSP3ECEFPreciseOrbits | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 301 | 21669 | sidereon_fit_precise_ephemeris_sample_orbit | positioning.FitPreciseEphemerisSampleOrbit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 302 | 21685 | sidereon_fit_sp3_ecef_precise_orbit | positioning.FitSP3ECEFPreciseOrbit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 303 | 21696 | sidereon_fit_sp3_ecef_precise_orbits | positioning.FitSP3ECEFPreciseOrbits | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 304 | 21709 | sidereon_fit_sp3_precise_orbit | positioning.FitSP3PreciseOrbit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 305 | 21724 | sidereon_fix_wide_lane_rtk_arc | positioning.FixWideLaneRTKArc | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 306 | 21736 | sidereon_force_j2_acceleration | astro.ForceJ2Acceleration | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 307 | 21747 | sidereon_force_twobody_acceleration | astro.ForceTwobodyAcceleration | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 308 | 21757 | sidereon_frame_catalog_count | astro.FrameCatalogCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 309 | 21765 | sidereon_frame_catalog_entries | astro.FrameCatalogEntries | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 310 | 21776 | sidereon_frame_catalog_entry | astro.FrameCatalogEntry | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 311 | 21786 | sidereon_frame_catalog_propagate_position | astro.FrameCatalogPropagatePosition | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 312 | 21797 | sidereon_frame_catalog_transform | astro.FrameCatalogTransform | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 313 | 21809 | sidereon_frame_catalog_transform_from_epoch | astro.FrameCatalogTransformFromEpoch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 314 | 21824 | sidereon_frame_gast_radians | astro.FrameGastRadians | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 315 | 21833 | sidereon_frame_gcrs_to_itrs | astro.FrameGCRSToITRS | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 316 | 21844 | sidereon_frame_gcrs_to_itrs_matrix | astro.FrameGCRSToITRSMatrix | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 317 | 21854 | sidereon_frame_gcrs_to_itrs_matrix_with_polar_motion | astro.FrameGCRSToITRSMatrixWithPolarMotion | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 318 | 21867 | sidereon_frame_gcrs_to_itrs_with_polar_motion | astro.FrameGCRSToITRSWithPolarMotion | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 319 | 21882 | sidereon_frame_gcrs_to_topocentric | astro.FrameGCRSToTopocentric | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 320 | 21897 | sidereon_frame_geodetic_from_ecef_proj | astro.FrameGeodeticFromECEFProj | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 321 | 21906 | sidereon_frame_geodetic_to_itrs | astro.FrameGeodeticToITRS | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 322 | 21918 | sidereon_frame_gmst_radians | astro.FrameGmstRadians | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 323 | 21927 | sidereon_frame_itrs_to_gcrs | astro.FrameITRSToGCRS | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 324 | 21937 | sidereon_frame_itrs_to_gcrs_matrix | astro.FrameITRSToGCRSMatrix | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 325 | 21947 | sidereon_frame_itrs_to_gcrs_matrix_with_polar_motion | astro.FrameITRSToGCRSMatrixWithPolarMotion | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 326 | 21960 | sidereon_frame_itrs_to_gcrs_with_polar_motion | astro.FrameITRSToGCRSWithPolarMotion | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 327 | 21973 | sidereon_frame_itrs_to_geodetic | astro.FrameITRSToGeodetic | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 328 | 21983 | sidereon_frame_mat3_vec3_mul | astro.FrameMat3Vec3Mul | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 329 | 21992 | sidereon_frame_mean_of_date_to_itrs_matrix | astro.FrameMeanOfDateToITRSMatrix | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 330 | 22002 | sidereon_frame_mean_of_date_to_itrs_matrix_with_polar_motion | astro.FrameMeanOfDateToITRSMatrixWithPolarMotion | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 331 | 22013 | sidereon_frame_polar_motion_matrix | astro.FramePolarMotionMatrix | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 332 | 22026 | sidereon_frame_teme_to_gcrs | astro.FrameTEMEToGCRS | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 333 | 22041 | sidereon_frequency_hz | support.FrequencyHz | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 334 | 22048 | sidereon_fusion_error_state_layout_dimension | errormetrics.FusionErrorStateLayoutDimension | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 335 | 22056 | sidereon_fusion_error_state_layout_label | errormetrics.FusionErrorStateLayoutLabel | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 336 | 22067 | sidereon_fusion_filter_config_init | errormetrics.FusionFilterConfigInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 337 | 22074 | sidereon_fusion_filter_configure_time_sync | geodesy.FusionFilterConfigureTimeSync | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 338 | 22084 | sidereon_fusion_filter_covariance | errormetrics.FusionFilterCovariance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 339 | 22097 | sidereon_fusion_filter_create | errormetrics.FusionFilterCreate | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 340 | 22110 | sidereon_fusion_filter_encode_state | errormetrics.FusionFilterEncodeState | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 341 | 22123 | sidereon_fusion_filter_free | errormetrics.FusionFilter.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 342 | 22131 | sidereon_fusion_filter_kind_label | errormetrics.FusionFilterKindLabel | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 343 | 22143 | sidereon_fusion_filter_propagate | astro.FusionFilterPropagate | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 344 | 22152 | sidereon_fusion_filter_propagate_recorded | astro.FusionFilterPropagateRecorded | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 345 | 22163 | sidereon_fusion_filter_restore_state | errormetrics.FusionFilterRestoreState | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 346 | 22173 | sidereon_fusion_filter_state | errormetrics.FusionFilterState | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 347 | 22182 | sidereon_fusion_filter_time_sync_status | geodesy.FusionFilterTimeSyncStatus | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 348 | 22192 | sidereon_fusion_filter_update_loose | errormetrics.FusionFilterUpdateLoose | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 349 | 22203 | sidereon_fusion_filter_update_loose_recorded | errormetrics.FusionFilterUpdateLooseRecorded | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 350 | 22216 | sidereon_fusion_filter_update_loose_time_sync | geodesy.FusionFilterUpdateLooseTimeSync | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 351 | 22227 | sidereon_fusion_filter_update_non_holonomic | errormetrics.FusionFilterUpdateNonHolonomic | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 352 | 22239 | sidereon_fusion_filter_update_non_holonomic_recorded | errormetrics.FusionFilterUpdateNonHolonomicRecorded | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 353 | 22251 | sidereon_fusion_filter_update_stationary | errormetrics.FusionFilterUpdateStationary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 354 | 22262 | sidereon_fusion_filter_update_stationary_recorded | errormetrics.FusionFilterUpdateStationaryRecorded | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 355 | 22274 | sidereon_fusion_filter_update_tight_broadcast | positioning.FusionFilterUpdateTightBroadcast | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 356 | 22286 | sidereon_fusion_filter_update_tight_broadcast_recorded | positioning.FusionFilterUpdateTightBroadcastRecorded | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 357 | 22300 | sidereon_fusion_filter_update_tight_broadcast_time_sync | positioning.FusionFilterUpdateTightBroadcastTimeSync | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 358 | 22311 | sidereon_fusion_filter_update_tight_sp3 | positioning.FusionFilterUpdateTightSP3 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 359 | 22323 | sidereon_fusion_filter_update_tight_sp3_recorded | positioning.FusionFilterUpdateTightSP3Recorded | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 360 | 22336 | sidereon_fusion_filter_update_tight_sp3_time_sync | positioning.FusionFilterUpdateTightSP3TimeSync | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 361 | 22347 | sidereon_fusion_gnss_fix_status_label | positioning.FusionGNSSFixStatusLabel | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 362 | 22358 | sidereon_fusion_imu_spec_preset | errormetrics.FusionImuSpecPreset | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 363 | 22367 | sidereon_fusion_rts_history_builder_finish | errormetrics.FusionRtsHistoryBuilderFinish | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 364 | 22376 | sidereon_fusion_rts_history_builder_free | errormetrics.FusionRtsHistoryBuilder.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 365 | 22384 | sidereon_fusion_rts_history_builder_from_filter | errormetrics.FusionRtsHistoryBuilderFromFilter | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 366 | 22393 | sidereon_fusion_rts_history_builder_new | errormetrics.FusionRtsHistoryBuilderNew | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 367 | 22401 | sidereon_fusion_rts_history_epoch | errormetrics.FusionRtsHistoryEpoch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 368 | 22410 | sidereon_fusion_rts_history_epoch_count | errormetrics.FusionRtsHistoryEpochCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 369 | 22419 | sidereon_fusion_rts_history_epoch_predicted_position_ecef_m | errormetrics.FusionRtsHistoryEpochPredictedPositionECEFM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 370 | 22432 | sidereon_fusion_rts_history_epoch_transition_from_previous | errormetrics.FusionRtsHistoryEpochTransitionFromPrevious | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 371 | 22445 | sidereon_fusion_rts_history_epoch_updated_position_ecef_m | errormetrics.FusionRtsHistoryEpochUpdatedPositionECEFM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 372 | 22458 | sidereon_fusion_rts_history_free | errormetrics.FusionRtsHistory.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 373 | 22470 | sidereon_fusion_velocity_match_outage | errormetrics.FusionVelocityMatchOutage | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 374 | 22492 | sidereon_galileo_nequick_g_native | geodesy.GalileoNequickGNative | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 375 | 22509 | sidereon_geodesic_direct | geodesy.GeodesicDirect | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 376 | 22521 | sidereon_geodesic_inverse | geodesy.GeodesicInverse | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 377 | 22534 | sidereon_geodetic_detect_steps | geodesy.GeodeticDetectSteps | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 378 | 22548 | sidereon_geodetic_fit_trajectory | geodesy.GeodeticFitTrajectory | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 379 | 22558 | sidereon_geodetic_midas_options_init | geodesy.GeodeticMidasOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 380 | 22565 | sidereon_geodetic_motion_field_common_mode | geodesy.GeodeticMotionFieldCommonMode | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 381 | 22574 | sidereon_geodetic_motion_field_free | geodesy.GeodeticMotionField.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 382 | 22582 | sidereon_geodetic_motion_field_stations | geodesy.GeodeticMotionFieldStations | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 383 | 22594 | sidereon_geodetic_network_field | geodesy.GeodeticNetworkField | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 384 | 22604 | sidereon_geodetic_step_detection_options_init | geodesy.GeodeticStepDetectionOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 385 | 22615 | sidereon_geodetic_to_ecef | geodesy.GeodeticToECEF | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 386 | 22623 | sidereon_geodetic_trajectory_components | geodesy.GeodeticTrajectoryComponents | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 387 | 22631 | sidereon_geodetic_trajectory_fit_options_init | geodesy.GeodeticTrajectoryFitOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 388 | 22639 | sidereon_geodetic_trajectory_free | geodesy.GeodeticTrajectory.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 389 | 22647 | sidereon_geodetic_trajectory_offsets | geodesy.GeodeticTrajectoryOffsets | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 390 | 22659 | sidereon_geodetic_trajectory_parameter_covariance | geodesy.GeodeticTrajectoryParameterCovariance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 391 | 22671 | sidereon_geodetic_trajectory_summary | geodesy.GeodeticTrajectorySummary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 392 | 22679 | sidereon_geodetic_trajectory_terms | geodesy.GeodeticTrajectoryTerms | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 393 | 22691 | sidereon_geodetic_velocity_midas | geodesy.GeodeticVelocityMidas | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 394 | 22701 | sidereon_geofence_containment_probability | geodesy.GeofenceContainmentProbability | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 395 | 22714 | sidereon_geofence_containment_probability_with_options | geodesy.GeofenceContainmentProbabilityWithOptions | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 396 | 22727 | sidereon_geofence_contains | geodesy.GeofenceContains | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 397 | 22739 | sidereon_geofence_create | geodesy.GeofenceCreate | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 398 | 22752 | sidereon_geofence_crossing_probability | geodesy.GeofenceCrossingProbability | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 399 | 22771 | sidereon_geofence_crossing_probability_with_options | geodesy.GeofenceCrossingProbabilityWithOptions | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 400 | 22789 | sidereon_geofence_distance_to_boundary | geodesy.GeofenceDistanceToBoundary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 401 | 22800 | sidereon_geofence_free | geodesy.Geofence.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 402 | 22807 | sidereon_geofence_hysteresis_init | geodesy.GeofenceHysteresisInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 403 | 22814 | sidereon_geofence_probability_options_init | geodesy.GeofenceProbabilityOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 404 | 22816 | sidereon_geoid_grid_ellipsoidal_height_rad | geodesy.GeoidGridEllipsoidalHeightRad | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 405 | 22827 | sidereon_geoid_grid_free | geodesy.GeoidGrid.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 406 | 22835 | sidereon_geoid_grid_from_egm2008_raster | geodesy.GeoidGridFromEgm2008Raster | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 407 | 22846 | sidereon_geoid_grid_from_egm2008_raster_window | geodesy.GeoidGridFromEgm2008RasterWindow | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 408 | 22851 | sidereon_geoid_grid_from_egm96_dac | geodesy.GeoidGridFromEGM96Dac | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 409 | 22863 | sidereon_geoid_grid_from_proj_egm96_gtx | geodesy.GeoidGridFromProjEGM96Gtx | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 410 | 22875 | sidereon_geoid_grid_from_text | geodesy.GeoidGridFromText | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 411 | 22888 | sidereon_geoid_grid_new | geodesy.GeoidGridNew | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 412 | 22898 | sidereon_geoid_grid_orthometric_height_rad | geodesy.GeoidGridOrthometricHeightRad | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 413 | 22910 | sidereon_geoid_grid_undulation_deg | geodesy.GeoidGridUndulationDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 414 | 22924 | sidereon_geoid_grid_undulation_proj_rad | geodesy.GeoidGridUndulationProjRad | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 415 | 22937 | sidereon_geoid_grid_undulation_rad | geodesy.GeoidGridUndulationRad | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 416 | 22942 | sidereon_geoid_grid_undulations_deg | geodesy.GeoidGridUndulationsDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 417 | 22950 | sidereon_geoid_grid_undulations_rad | geodesy.GeoidGridUndulationsRad | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 418 | 22965 | sidereon_geoid_undulation | geodesy.GeoidUndulation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 419 | 22969 | sidereon_geoid_undulations_deg | geodesy.GeoidUndulationsDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 420 | 22976 | sidereon_geoid_undulations_rad | geodesy.GeoidUndulationsRad | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 421 | 22989 | sidereon_glonass_g1_frequency_hz | positioning.GLONASSG1FrequencyHz | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 422 | 22996 | sidereon_gnss_seconds_of_week_from_calendar | positioning.GNSSSecondsOfWeekFromCalendar | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 423 | 23007 | sidereon_gnss_system_label | positioning.GNSSSystemLabel | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 424 | 23019 | sidereon_gnss_week_and_seconds_of_week | positioning.GNSSWeekAndSecondsOfWeek | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 425 | 23028 | sidereon_gnss_week_epoch_julian_day_number | positioning.GNSSWeekEpochJulianDayNumber | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 426 | 23037 | sidereon_gnss_week_from_calendar | positioning.GNSSWeekFromCalendar | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 427 | 23049 | sidereon_gnss_week_tow_new | positioning.GNSSWeekTowNew | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 428 | 23059 | sidereon_gnss_week_tow_normalized | positioning.GNSSWeekTowNormalized | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 429 | 23067 | sidereon_gnss_week_tow_unrolled_week | positioning.GNSSWeekTowUnrolledWeek | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 430 | 23079 | sidereon_gps_utc_offset_s | support.GPSUtcOffsetS | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 431 | 23086 | sidereon_ground_track_count | astro.GroundTrackCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 432 | 23097 | sidereon_ground_track_free | astro.GroundTrack.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 433 | 23107 | sidereon_ground_track_values | astro.GroundTrackValues | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 434 | 23121 | sidereon_hessian_trace | errormetrics.HessianTrace | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 435 | 23130 | sidereon_instant_from_utc_civil | support.InstantFromUtcCivil | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 436 | 23151 | sidereon_iod_gauss_angles | astro.IODGaussAngles | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 437 | 23167 | sidereon_iod_gibbs | astro.IODGibbs | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 438 | 23182 | sidereon_iod_hgibbs | astro.IODHgibbs | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 439 | 23199 | sidereon_ionex_epoch_count | positioning.IonexEpochCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 440 | 23209 | sidereon_ionex_exponent | positioning.IonexExponent | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 441 | 23219 | sidereon_ionex_free | positioning.Ionex.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 442 | 23228 | sidereon_ionex_from_tec_grid_samples | positioning.IonexFromTecGridSamples | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 443 | 23239 | sidereon_ionex_from_tec_samples | positioning.IonexFromTecSamples | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 444 | 23254 | sidereon_ionex_lat_nodes_deg | positioning.IonexLatNodesDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 445 | 23268 | sidereon_ionex_lon_nodes_deg | positioning.IonexLonNodesDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 446 | 23282 | sidereon_ionex_map_epochs_j2000_s | positioning.IonexMapEpochsJ2000S | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 447 | 23295 | sidereon_ionex_parse | positioning.IonexParse | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 448 | 23313 | sidereon_ionex_slant_delay | positioning.IonexSlantDelay | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 449 | 23331 | sidereon_ionex_slant_delay_with_policy | positioning.IonexSlantDelayWithPolicy | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 450 | 23348 | sidereon_ionex_tec_grid_samples_epochs_j2000_s | positioning.IonexTecGridSamplesEpochsJ2000S | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 451 | 23361 | sidereon_ionex_tec_grid_samples_info | positioning.IonexTecGridSamplesInfo | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 452 | 23371 | sidereon_ionex_tec_grid_samples_rms_maps_tecu | positioning.IonexTecGridSamplesRmsMapsTecu | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 453 | 23384 | sidereon_ionex_tec_grid_samples_tec_maps_tecu | positioning.IonexTecGridSamplesTecMapsTecu | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 454 | 23397 | sidereon_ionex_tec_samples | positioning.IonexTecSamples | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 455 | 23413 | sidereon_ionex_to_ionex_text | positioning.IonexToIonexText | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 456 | 23426 | sidereon_iono_free_pseudoranges_combined | support.IonoFreePseudorangesCombined | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 457 | 23438 | sidereon_iono_free_pseudoranges_dropped | support.IonoFreePseudorangesDropped | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 458 | 23450 | sidereon_iono_free_pseudoranges_free | support.IonoFreePseudoranges.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 459 | 23458 | sidereon_j2000_seconds_to_civil | geodesy.J2000SecondsToCivil | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 460 | 23471 | sidereon_kalman_cv_steady_state_gains | errormetrics.KalmanCvSteadyStateGains | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 461 | 23489 | sidereon_klobuchar_native | support.KlobucharNative | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 462 | 23515 | sidereon_lambda_ils_search | support.LambdaIlsSearch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 463 | 23531 | sidereon_lambert_battin | support.LambertBattin | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 464 | 23551 | sidereon_last_error_message | core.ErrorDetail (internal) | internal same-thread error detail capture; text retained in StatusError |
| 465 | 23559 | sidereon_last_terrain_datum_error | geodesy.LastTerrainDatumError | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 466 | 23567 | sidereon_last_terrain_store_error | geodesy.LastTerrainStoreError | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 467 | 23574 | sidereon_leap_second_table_info | geodesy.LeapSecondTableInfo | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 468 | 23582 | sidereon_leap_second_table_source | geodesy.LeapSecondTableSource | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 469 | 23593 | sidereon_leap_seconds | geodesy.LeapSeconds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 470 | 23605 | sidereon_line_of_sight_from_az_el_deg | support.LineOfSightFromAzElDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 471 | 23619 | sidereon_lnav_decode | astro.LNAVDecode | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 472 | 23638 | sidereon_lnav_encode | astro.LNAVEncode | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 473 | 23657 | sidereon_locate_source | support.LocateSource | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 474 | 23675 | sidereon_locate_source_with | support.LocateSourceWith | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 475 | 23688 | sidereon_look_angles_epoch_count | support.LookAnglesEpochCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 476 | 23700 | sidereon_look_angles_free | support.LookAngles.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 477 | 23710 | sidereon_look_angles_values | support.LookAnglesValues | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 478 | 23721 | sidereon_mad_gaussian_consistency | errormetrics.MADGaussianConsistency | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 479 | 23729 | sidereon_mad_spread | errormetrics.MADSpread | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 480 | 23734 | sidereon_mean_to_eccentric_anomaly | astro.MeanToEccentricAnomaly | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 481 | 23738 | sidereon_mean_to_true_anomaly | astro.MeanToTrueAnomaly | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 482 | 23742 | sidereon_mee2coe | support.Mee2coe | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 483 | 23745 | sidereon_mee2rv | support.Mee2rv | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 484 | 23758 | sidereon_met_init | support.MetInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 485 | 23766 | sidereon_mmap_terrain_checksum64 | distribution.MmapTerrainChecksum64 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 486 | 23776 | sidereon_mmap_terrain_digest_provenance | distribution.MmapTerrainDigestProvenance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 487 | 23786 | sidereon_mmap_terrain_ellipsoidal_height_m | distribution.MmapTerrainEllipsoidalHeightM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 488 | 23801 | sidereon_mmap_terrain_ellipsoidal_height_m_with_model | distribution.MmapTerrainEllipsoidalHeightMWithModel | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 489 | 23818 | sidereon_mmap_terrain_ellipsoidal_height_m_with_options | distribution.MmapTerrainEllipsoidalHeightMWithOptions | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 490 | 23829 | sidereon_mmap_terrain_free | distribution.MmapTerrain.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 491 | 23839 | sidereon_mmap_terrain_from_bytes | distribution.MmapTerrainFromBytes | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 492 | 23850 | sidereon_mmap_terrain_from_path | distribution.MmapTerrainFromPath | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 493 | 23862 | sidereon_mmap_terrain_from_path_attested | distribution.MmapTerrainFromPathAttested | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 494 | 23873 | sidereon_mmap_terrain_from_vec | distribution.MmapTerrainFromVec | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 495 | 23887 | sidereon_mmap_terrain_height_batch | distribution.MmapTerrainHeightBatch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 496 | 23900 | sidereon_mmap_terrain_height_m | distribution.MmapTerrainHeightM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 497 | 23913 | sidereon_mmap_terrain_height_m_with_options | distribution.MmapTerrainHeightMWithOptions | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 498 | 23929 | sidereon_mmap_terrain_orthometric_height_batch | distribution.MmapTerrainOrthometricHeightBatch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 499 | 23942 | sidereon_mmap_terrain_orthometric_height_m | distribution.MmapTerrainOrthometricHeightM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 500 | 23956 | sidereon_mmap_terrain_orthometric_height_m_with_options | distribution.MmapTerrainOrthometricHeightMWithOptions | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 501 | 23969 | sidereon_mmap_terrain_tile_index | distribution.MmapTerrainTileIndex | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 502 | 23982 | sidereon_mmap_terrain_to_bytes | distribution.MmapTerrainToBytes | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 503 | 23994 | sidereon_mmap_terrain_verify | distribution.MmapTerrainVerify | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 504 | 24003 | sidereon_mmap_terrain_vertical_datum | distribution.MmapTerrainVerticalDatum | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 505 | 24012 | sidereon_moon_angle_deg | astro.MoonAngleDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 506 | 24024 | sidereon_moon_az_el | astro.MoonAzEl | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 507 | 24038 | sidereon_moon_elevation_deg | astro.MoonElevationDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 508 | 24048 | sidereon_moon_elevation_options_init | astro.MoonElevationOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 509 | 24057 | sidereon_moon_illumination | astro.MoonIllumination | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 510 | 24067 | sidereon_moving_baseline_solution_epoch | support.MovingBaselineSolutionEpoch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 511 | 24076 | sidereon_moving_baseline_solution_epoch_count | support.MovingBaselineSolutionEpochCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 512 | 24084 | sidereon_moving_baseline_solution_free | support.MovingBaselineSolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 513 | 24092 | sidereon_navcen_assessment | distribution.NAVCENAssessment | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 514 | 24101 | sidereon_navcen_assessment_count | distribution.NAVCENAssessmentCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 515 | 24108 | sidereon_navcen_assessment_nanu_subject | distribution.NAVCENAssessmentNanuSubject | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 516 | 24119 | sidereon_navcen_assessment_nanu_type | distribution.NAVCENAssessmentNanuType | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 517 | 24131 | sidereon_navcen_assessment_outage_start | distribution.NAVCENAssessmentOutageStart | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 518 | 24144 | sidereon_navcen_assessments_free | distribution.NAVCENAssessments.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 519 | 24154 | sidereon_navcen_parse_at | distribution.NAVCENParseAt | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 520 | 24166 | sidereon_nequick_g_delay_m | geodesy.NequickGDelayM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 521 | 24182 | sidereon_nequick_g_stec_tecu | geodesy.NequickGStecTecu | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 522 | 24193 | sidereon_nis | errormetrics.NIS | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 523 | 24200 | sidereon_nis_expected_value | errormetrics.NISExpectedValue | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 524 | 24207 | sidereon_nis_gate_test | errormetrics.NISGateTest | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 525 | 24218 | sidereon_nis_gate_threshold | errormetrics.NISGateThreshold | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 526 | 24222 | sidereon_nmea_accumulator_epochs | support.NMEAAccumulatorEpochs | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 527 | 24228 | sidereon_nmea_accumulator_finish | support.NMEAAccumulatorFinish | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 528 | 24231 | sidereon_nmea_accumulator_free | support.NMEAAccumulator.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 529 | 24233 | sidereon_nmea_accumulator_new | support.NMEAAccumulatorNew | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 530 | 24235 | sidereon_nmea_accumulator_push | support.NMEAAccumulatorPush | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 531 | 24240 | sidereon_nmea_accumulator_retained_len | support.NMEAAccumulatorRetainedLen | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 532 | 24243 | sidereon_nmea_accumulator_summary | support.NMEAAccumulatorSummary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 533 | 24246 | sidereon_nmea_log_epochs | support.NMEALogEpochs | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 534 | 24252 | sidereon_nmea_log_free | support.NMEALog.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 535 | 24254 | sidereon_nmea_log_summary | support.NMEALogSummary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 536 | 24257 | sidereon_nmea_parse | support.NMEAParse | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 537 | 24261 | sidereon_nmea_write_gga | support.NMEAWriteGga | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 538 | 24279 | sidereon_normal_covariance | errormetrics.NormalCovariance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 539 | 24293 | sidereon_normalized_innovation | support.NormalizedInnovation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 540 | 24297 | sidereon_ntrip_bytes | distribution.NTRIPBytes | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 541 | 24303 | sidereon_ntrip_bytes_free | distribution.NTRIPBytes.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 542 | 24305 | sidereon_ntrip_events_count | distribution.NTRIPEventsCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 543 | 24308 | sidereon_ntrip_events_detail | distribution.NTRIPEventsDetail | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 544 | 24315 | sidereon_ntrip_events_event | distribution.NTRIPEventsEvent | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 545 | 24319 | sidereon_ntrip_events_free | distribution.NTRIPEvents.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 546 | 24321 | sidereon_ntrip_events_payload | distribution.NTRIPEventsPayload | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 547 | 24328 | sidereon_ntrip_events_sourcetable | distribution.NTRIPEventsSourcetable | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 548 | 24332 | sidereon_ntrip_machine_connection_request | distribution.NTRIPMachineConnectionRequest | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 549 | 24335 | sidereon_ntrip_machine_finish | distribution.NTRIPMachineFinish | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 550 | 24338 | sidereon_ntrip_machine_free | distribution.NTRIPMachine.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 551 | 24340 | sidereon_ntrip_machine_new | distribution.NTRIPMachineNew | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 552 | 24343 | sidereon_ntrip_machine_push | distribution.NTRIPMachinePush | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 553 | 24348 | sidereon_ntrip_machine_reset | distribution.NTRIPMachineReset | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 554 | 24350 | sidereon_ntrip_machine_state | distribution.NTRIPMachineState | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 555 | 24353 | sidereon_ntrip_machine_try_gga_message | distribution.NTRIPMachineTryGgaMessage | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 556 | 24360 | sidereon_ntrip_request_bytes | distribution.NTRIPRequestBytes | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 557 | 24366 | sidereon_ntrip_sourcetable_free | distribution.NTRIPSourcetable.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 558 | 24368 | sidereon_ntrip_sourcetable_parse | distribution.NTRIPSourcetableParse | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 559 | 24372 | sidereon_ntrip_sourcetable_streams | distribution.NTRIPSourcetableStreams | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 560 | 24378 | sidereon_ntrip_sourcetable_summary | distribution.NTRIPSourcetableSummary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 561 | 24381 | sidereon_ntrip_sourcetable_to_text | distribution.NTRIPSourcetableToText | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 562 | 24394 | sidereon_nutation_equation_of_equinoxes_terms | geodesy.NutationEquationOfEquinoxesTerms | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 563 | 24403 | sidereon_nutation_fundamental_arguments | geodesy.NutationFundamentalArguments | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 564 | 24412 | sidereon_nutation_iau2000a_radians | geodesy.NutationIau2000aRadians | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 565 | 24423 | sidereon_nutation_matrix | geodesy.NutationMatrix | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 566 | 24434 | sidereon_nutation_mean_obliquity_radians | astro.NutationMeanObliquityRadians | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 567 | 24444 | sidereon_observability_tier_label | errormetrics.ObservabilityTierLabel | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 568 | 24457 | sidereon_observable_state_missing_position_ecef_m | errormetrics.ObservableStateMissingPositionECEFM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 569 | 24466 | sidereon_observables_options_init | positioning.ObservablesOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 570 | 24468 | sidereon_observation_qc_clock_jumps | positioning.ObservationQcClockJumps | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 571 | 24474 | sidereon_observation_qc_cycle_slip_systems | positioning.ObservationQcCycleSlipSystems | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 572 | 24480 | sidereon_observation_qc_cycle_slips | positioning.ObservationQcCycleSlips | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 573 | 24483 | sidereon_observation_qc_from_obs | positioning.ObservationQcFromObs | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 574 | 24487 | sidereon_observation_qc_gaps | positioning.ObservationQcGaps | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 575 | 24493 | sidereon_observation_qc_multipath_satellites | positioning.ObservationQcMultipathSatellites | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 576 | 24499 | sidereon_observation_qc_multipath_systems | positioning.ObservationQcMultipathSystems | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 577 | 24505 | sidereon_observation_qc_options_init | positioning.ObservationQcOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 578 | 24507 | sidereon_observation_qc_parse | positioning.ObservationQcParse | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 579 | 24519 | sidereon_observation_qc_render_html | positioning.ObservationQcRenderHtml | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 580 | 24532 | sidereon_observation_qc_render_text | positioning.ObservationQcRenderText | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 581 | 24538 | sidereon_observation_qc_report_free | positioning.ObservationQcReport.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 582 | 24540 | sidereon_observation_qc_satellite_signals | positioning.ObservationQcSatelliteSignals | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 583 | 24546 | sidereon_observation_qc_satellites | positioning.ObservationQcSatellites | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 584 | 24552 | sidereon_observation_qc_summary | positioning.ObservationQcSummary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 585 | 24555 | sidereon_observation_qc_system_signals | positioning.ObservationQcSystemSignals | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 586 | 24567 | sidereon_observation_qc_to_json | positioning.ObservationQcToJSON | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 587 | 24573 | sidereon_observe | support.Observe | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 588 | 24583 | sidereon_observe_options_init | support.ObserveOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 589 | 24585 | sidereon_observe_spk_body | astro.ObserveSPKBody | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 590 | 24598 | sidereon_ocean_tide_loading | geodesy.OceanTideLoading | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 591 | 24612 | sidereon_oem_free | astro.OEM.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 592 | 24621 | sidereon_oem_parse_kvn | astro.OEMParseKVN | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 593 | 24632 | sidereon_oem_parse_xml | astro.OEMParseXML | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 594 | 24642 | sidereon_oem_segment_count | astro.OEMSegmentCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 595 | 24654 | sidereon_oem_to_kvn | astro.OEMToKVN | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 596 | 24670 | sidereon_oem_to_xml | astro.OEMToXML | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 597 | 24695 | sidereon_omm_catalog_build_lenient | astro.OMMCatalogBuildLenient | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 598 | 24707 | sidereon_omm_catalog_free | astro.OMMCatalog.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 599 | 24717 | sidereon_omm_catalog_malformed_count | astro.OMMCatalogMalformedCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 600 | 24729 | sidereon_omm_catalog_record | astro.OMMCatalogRecord | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 601 | 24739 | sidereon_omm_catalog_record_count | astro.OMMCatalogRecordCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 602 | 24751 | sidereon_omm_catalog_skipped | astro.OMMCatalogSkipped | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 603 | 24760 | sidereon_omm_catalog_skipped_count | astro.OMMCatalogSkippedCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 604 | 24775 | sidereon_omm_catalog_skipped_object_name | astro.OMMCatalogSkippedObjectName | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 605 | 24787 | sidereon_omm_free | astro.OMM.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 606 | 24795 | sidereon_omm_parse_json | astro.OMMParseJSON | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 607 | 24805 | sidereon_omm_parse_kvn | astro.OMMParseKVN | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 608 | 24815 | sidereon_omm_parse_xml | astro.OMMParseXML | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 609 | 24827 | sidereon_omm_to_json | astro.OMMToJSON | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 610 | 24841 | sidereon_omm_to_kvn | astro.OMMToKVN | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 611 | 24855 | sidereon_omm_to_xml | astro.OMMToXML | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 612 | 24867 | sidereon_opm_free | astro.OPM.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 613 | 24876 | sidereon_opm_parse_kvn | astro.OPMParseKVN | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 614 | 24887 | sidereon_opm_parse_xml | astro.OPMParseXML | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 615 | 24901 | sidereon_opm_to_kvn | astro.OPMToKVN | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 616 | 24917 | sidereon_opm_to_xml | astro.OPMToXML | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 617 | 24928 | sidereon_orbit_fit_options_init | astro.OrbitFitOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 618 | 24935 | sidereon_orbit_fit_report_arc_span | astro.OrbitFitReportArcSpan | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 619 | 24944 | sidereon_orbit_fit_report_constellation_ledger | astro.OrbitFitReportConstellationLedger | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 620 | 24956 | sidereon_orbit_fit_report_fits | astro.OrbitFitReportFits | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 621 | 24967 | sidereon_orbit_fit_report_free | astro.OrbitFitReport.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 622 | 24975 | sidereon_orbit_fit_report_satellite_ledger | astro.OrbitFitReportSatelliteLedger | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 623 | 24988 | sidereon_orthometric_height_m | geodesy.OrthometricHeightM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 624 | 24999 | sidereon_parallactic_angle_deg | geodesy.ParallacticAngleDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 625 | 25018 | sidereon_parse_tle_file | astro.ParseTLEFile | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 626 | 25028 | sidereon_pass_finder_options_init | astro.PassFinderOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 627 | 25035 | sidereon_pass_list_count | astro.PassListCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 628 | 25046 | sidereon_pass_list_free | astro.PassList.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 629 | 25056 | sidereon_pass_list_values | astro.PassListValues | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 630 | 25068 | sidereon_phase_angle_deg | support.PhaseAngleDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 631 | 25080 | sidereon_position_angle_deg | support.PositionAngleDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 632 | 25093 | sidereon_ppp_auto_init_options_init | positioning.PPPAutoInitOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 633 | 25106 | sidereon_ppp_corrections_build | positioning.PPPCorrectionsBuild | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 634 | 25121 | sidereon_ppp_corrections_code_bias | positioning.PPPCorrectionsCodeBias | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 635 | 25133 | sidereon_ppp_corrections_free | positioning.PPPCorrections.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 636 | 25143 | sidereon_ppp_corrections_ocean_loading | positioning.PPPCorrectionsOceanLoading | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 637 | 25157 | sidereon_ppp_corrections_pole_tide | positioning.PPPCorrectionsPoleTide | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 638 | 25171 | sidereon_ppp_corrections_sat_pco_ecef | positioning.PPPCorrectionsSatPcoECEF | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 639 | 25185 | sidereon_ppp_corrections_sat_pcv | positioning.PPPCorrectionsSatPcv | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 640 | 25199 | sidereon_ppp_corrections_tide | positioning.PPPCorrectionsTide | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 641 | 25213 | sidereon_ppp_corrections_windup | positioning.PPPCorrectionsWindup | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 642 | 25224 | sidereon_ppp_fixed_ambiguity_options_init | positioning.PPPFixedAmbiguityOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 643 | 25234 | sidereon_ppp_fixed_solution_fixed_ambiguities | positioning.PPPFixedSolutionFixedAmbiguities | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 644 | 25247 | sidereon_ppp_fixed_solution_float_position | positioning.PPPFixedSolutionFloatPosition | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 645 | 25259 | sidereon_ppp_fixed_solution_free | positioning.PPPFixedSolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 646 | 25267 | sidereon_ppp_fixed_solution_metadata | positioning.PPPFixedSolutionMetadata | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 647 | 25276 | sidereon_ppp_fixed_solution_position | positioning.PPPFixedSolutionPosition | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 648 | 25286 | sidereon_ppp_fixed_solution_position_covariances | positioning.PPPFixedSolutionPositionCovariances | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 649 | 25295 | sidereon_ppp_fixed_solution_temporal_correlation | positioning.PPPFixedSolutionTemporalCorrelation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 650 | 25304 | sidereon_ppp_fixed_solution_tropo_gradient | positioning.PPPFixedSolutionTropoGradient | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 651 | 25316 | sidereon_ppp_fixed_solution_used_ids | positioning.PPPFixedSolutionUsedIds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 652 | 25333 | sidereon_ppp_fixed_solution_used_sat_ids | positioning.PPPFixedSolutionUsedSatIds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 653 | 25344 | sidereon_ppp_float_options_init | positioning.PPPFloatOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 654 | 25354 | sidereon_ppp_float_solution_ambiguities | positioning.PPPFloatSolutionAmbiguities | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 655 | 25368 | sidereon_ppp_float_solution_free | positioning.PPPFloatSolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 656 | 25376 | sidereon_ppp_float_solution_metadata | positioning.PPPFloatSolutionMetadata | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 657 | 25385 | sidereon_ppp_float_solution_position | positioning.PPPFloatSolutionPosition | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 658 | 25395 | sidereon_ppp_float_solution_position_covariances | positioning.PPPFloatSolutionPositionCovariances | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 659 | 25404 | sidereon_ppp_float_solution_temporal_correlation | positioning.PPPFloatSolutionTemporalCorrelation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 660 | 25413 | sidereon_ppp_float_solution_tropo_gradient | positioning.PPPFloatSolutionTropoGradient | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 661 | 25425 | sidereon_ppp_float_solution_used_ids | positioning.PPPFloatSolutionUsedIds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 662 | 25442 | sidereon_ppp_float_solution_used_sat_ids | positioning.PPPFloatSolutionUsedSatIds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 663 | 25453 | sidereon_ppp_measurement_weights_init | positioning.PPPMeasurementWeightsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 664 | 25460 | sidereon_ppp_range_corrections_init | positioning.PPPRangeCorrectionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 665 | 25467 | sidereon_ppp_troposphere_options_init | positioning.PPPTroposphereOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 666 | 25475 | sidereon_precession_icrs_to_j2000_matrix | geodesy.PrecessionIcrsToJ2000Matrix | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 667 | 25484 | sidereon_precession_matrix | geodesy.PrecessionMatrix | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 668 | 25492 | sidereon_precise_ephemeris_interpolant_free | positioning.PreciseEphemerisInterpolant.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 669 | 25502 | sidereon_precise_ephemeris_interpolant_from_precise_ephemeris_samples | positioning.PreciseEphemerisInterpolantFromPreciseEphemerisSamples | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 670 | 25513 | sidereon_precise_ephemeris_interpolant_from_samples | positioning.PreciseEphemerisInterpolantFromSamples | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 671 | 25525 | sidereon_precise_ephemeris_interpolant_from_sp3 | positioning.PreciseEphemerisInterpolantFromSP3 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 672 | 25536 | sidereon_precise_ephemeris_interpolant_observable_states_at_j2000_s | positioning.PreciseEphemerisInterpolantObservableStatesAtJ2000S | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 673 | 25553 | sidereon_precise_ephemeris_interpolant_observable_states_at_shared_j2000_s | positioning.PreciseEphemerisInterpolantObservableStatesAtSharedJ2000S | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 674 | 25572 | sidereon_precise_ephemeris_samples_free | positioning.PreciseEphemerisSamples.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 675 | 25586 | sidereon_precise_ephemeris_samples_from_samples | positioning.PreciseEphemerisSamplesFromSamples | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 676 | 25598 | sidereon_precise_ephemeris_samples_observable_states_at_j2000_s | positioning.PreciseEphemerisSamplesObservableStatesAtJ2000S | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 677 | 25615 | sidereon_precise_ephemeris_samples_observable_states_at_shared_j2000_s | positioning.PreciseEphemerisSamplesObservableStatesAtSharedJ2000S | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 678 | 25637 | sidereon_precise_ephemeris_samples_predict_ranges | positioning.PreciseEphemerisSamplesPredictRanges | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 679 | 25650 | sidereon_precise_ephemeris_samples_sample | positioning.PreciseEphemerisSamplesSample | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 680 | 25667 | sidereon_precise_interpolant_artifact_checksum64 | positioning.PreciseInterpolantArtifactChecksum64 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 681 | 25678 | sidereon_precise_interpolant_artifact_digest_provenance | positioning.PreciseInterpolantArtifactDigestProvenance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 682 | 25687 | sidereon_precise_interpolant_artifact_free | positioning.PreciseInterpolantArtifact.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 683 | 25697 | sidereon_precise_interpolant_artifact_from_path | positioning.PreciseInterpolantArtifactFromPath | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 684 | 25712 | sidereon_precise_interpolant_artifact_from_path_attested | positioning.PreciseInterpolantArtifactFromPathAttested | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 685 | 25722 | sidereon_precise_interpolant_artifact_handle_checksum64 | positioning.PreciseInterpolantArtifactHandleChecksum64 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 686 | 25733 | sidereon_precise_interpolant_artifact_open_borrowed | positioning.PreciseInterpolantArtifactOpenBorrowed | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 687 | 25744 | sidereon_precise_interpolant_artifact_open_owned | positioning.PreciseInterpolantArtifactOpenOwned | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 688 | 25756 | sidereon_precise_interpolant_artifact_satellites | positioning.PreciseInterpolantArtifactSatellites | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 689 | 25768 | sidereon_precise_interpolant_artifact_state | positioning.PreciseInterpolantArtifactState | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 690 | 25780 | sidereon_precise_interpolant_artifact_verify | positioning.PreciseInterpolantArtifactVerify | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 691 | 25794 | sidereon_prepare_ionosphere_free_rtk_arc | positioning.PrepareIonosphereFreeRTKArc | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 692 | 25801 | sidereon_propagate_covariance | astro.PropagateCovariance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 693 | 25808 | sidereon_propagate_kepler | astro.PropagateKepler | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 694 | 25821 | sidereon_propagate_state | astro.PropagateState | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 695 | 25836 | sidereon_propagate_tle_batch | astro.PropagateTLEBatch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 696 | 25851 | sidereon_pseudorange_variance | positioning.PseudorangeVariance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 697 | 25860 | sidereon_pseudorange_variance_options_init | positioning.PseudorangeVarianceOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 698 | 25873 | sidereon_raim | support.RAIM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 699 | 25894 | sidereon_raim_fde_design | errormetrics.RAIMFDEDesign | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 700 | 25908 | sidereon_raim_for_solution | support.RAIMForSolution | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 701 | 25925 | sidereon_raim_normalized_residuals | support.RAIMNormalizedResiduals | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 702 | 25945 | sidereon_range_fde_options_init | positioning.RangeFDEOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 703 | 25954 | sidereon_range_fde_result_covariance | positioning.RangeFDEResultCovariance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 704 | 25967 | sidereon_range_fde_result_diagnostics | positioning.RangeFDEResultDiagnostics | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 705 | 25980 | sidereon_range_fde_result_excluded | positioning.RangeFDEResultExcluded | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 706 | 25991 | sidereon_range_fde_result_free | positioning.RangeFDEResult.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 707 | 25998 | sidereon_range_fde_result_global_test | positioning.RangeFDEResultGlobalTest | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 708 | 26006 | sidereon_range_fde_result_iterations | positioning.RangeFDEResultIterations | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 709 | 26016 | sidereon_range_fde_result_state_correction | positioning.RangeFDEResultStateCorrection | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 710 | 26027 | sidereon_range_fde_result_state_dim | positioning.RangeFDEResultStateDim | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 711 | 26041 | sidereon_reduced_orbit_drift | astro.ReducedOrbitDrift | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 712 | 26056 | sidereon_reduced_orbit_drift_report_entries | astro.ReducedOrbitDriftReportEntries | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 713 | 26069 | sidereon_reduced_orbit_drift_report_free | astro.ReducedOrbitDriftReport.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 714 | 26077 | sidereon_reduced_orbit_drift_report_requested_samples | astro.ReducedOrbitDriftReportRequestedSamples | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 715 | 26087 | sidereon_reduced_orbit_drift_report_summary | astro.ReducedOrbitDriftReportSummary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 716 | 26100 | sidereon_reduced_orbit_drift_sp3_source | positioning.ReducedOrbitDriftSP3Source | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 717 | 26115 | sidereon_reduced_orbit_drift_tle_source | astro.ReducedOrbitDriftTLESource | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 718 | 26131 | sidereon_reduced_orbit_fit | astro.ReducedOrbitFit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 719 | 26147 | sidereon_reduced_orbit_fit_piecewise | astro.ReducedOrbitFitPiecewise | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 720 | 26164 | sidereon_reduced_orbit_fit_piecewise_sp3_source | positioning.ReducedOrbitFitPiecewiseSP3Source | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 721 | 26178 | sidereon_reduced_orbit_fit_piecewise_tle_source | astro.ReducedOrbitFitPiecewiseTLESource | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 722 | 26192 | sidereon_reduced_orbit_fit_sp3_source | positioning.ReducedOrbitFitSP3Source | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 723 | 26205 | sidereon_reduced_orbit_fit_tle_source | astro.ReducedOrbitFitTLESource | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 724 | 26218 | sidereon_reduced_orbit_piecewise_drift | astro.ReducedOrbitPiecewiseDrift | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 725 | 26233 | sidereon_reduced_orbit_piecewise_drift_sp3_source | positioning.ReducedOrbitPiecewiseDriftSP3Source | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 726 | 26246 | sidereon_reduced_orbit_piecewise_drift_tle_source | astro.ReducedOrbitPiecewiseDriftTLESource | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 727 | 26257 | sidereon_reduced_orbit_piecewise_free | astro.ReducedOrbitPiecewise.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 728 | 26265 | sidereon_reduced_orbit_piecewise_info | astro.ReducedOrbitPiecewiseInfo | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 729 | 26275 | sidereon_reduced_orbit_piecewise_position | astro.ReducedOrbitPiecewisePosition | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 730 | 26288 | sidereon_reduced_orbit_piecewise_position_velocity | astro.ReducedOrbitPiecewisePositionVelocity | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 731 | 26302 | sidereon_reduced_orbit_piecewise_segments | astro.ReducedOrbitPiecewiseSegments | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 732 | 26315 | sidereon_reduced_orbit_piecewise_select_segment | astro.ReducedOrbitPiecewiseSelectSegment | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 733 | 26330 | sidereon_reduced_orbit_position | astro.ReducedOrbitPosition | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 734 | 26347 | sidereon_reduced_orbit_position_velocity | astro.ReducedOrbitPositionVelocity | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 735 | 26354 | sidereon_relative_mean_motion_circular | astro.RelativeMeanMotionCircular | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 736 | 26356 | sidereon_relative_mean_motion_from_state | astro.RelativeMeanMotionFromState | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 737 | 26359 | sidereon_relative_rotation | support.RelativeRotation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 738 | 26364 | sidereon_relative_state | support.RelativeState | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 739 | 26377 | sidereon_reliability_araim | errormetrics.ReliabilityAraim | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 740 | 26392 | sidereon_reliability_design | errormetrics.ReliabilityDesign | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 741 | 26402 | sidereon_reliability_options_init | errormetrics.ReliabilityOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 742 | 26409 | sidereon_reliability_report_free | errormetrics.ReliabilityReport.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 743 | 26420 | sidereon_reliability_report_observations | errormetrics.ReliabilityReportObservations | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 744 | 26432 | sidereon_reliability_report_summary | errormetrics.ReliabilityReportSummary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 745 | 26443 | sidereon_residual_jarque_bera | errormetrics.ResidualJarqueBera | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 746 | 26456 | sidereon_residual_kurtosis | errormetrics.ResidualKurtosis | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 747 | 26471 | sidereon_residual_moments | errormetrics.ResidualMoments | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 748 | 26486 | sidereon_residual_shapiro_wilk | errormetrics.ResidualShapiroWilk | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 749 | 26499 | sidereon_residual_skewness | errormetrics.ResidualSkewness | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 750 | 26507 | sidereon_rf_cn0 | support.RFCn0 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 751 | 26519 | sidereon_rf_dish_gain | support.RFDishGain | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 752 | 26530 | sidereon_rf_eirp | support.RFEirp | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 753 | 26537 | sidereon_rf_fspl | support.RFFspl | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 754 | 26545 | sidereon_rf_link_margin | support.RFLinkMargin | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 755 | 26553 | sidereon_rf_wavelength | support.RFWavelength | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 756 | 26564 | sidereon_rinex_clock_bias_at_gps_seconds | positioning.RINEXClockBiasAtGPSSeconds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 757 | 26575 | sidereon_rinex_clock_free | positioning.RINEXClock.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 758 | 26584 | sidereon_rinex_clock_parse | positioning.RINEXClockParse | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 759 | 26593 | sidereon_rinex_clock_satellite_count | positioning.RINEXClockSatelliteCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 760 | 26604 | sidereon_rinex_clock_to_text | positioning.RINEXClockToText | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 761 | 26620 | sidereon_rinex_encode_nav | positioning.RINEXEncodeNav | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 762 | 26626 | sidereon_rinex_lint_findings | positioning.RINEXLintFindings | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 763 | 26632 | sidereon_rinex_lint_nav | positioning.RINEXLintNav | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 764 | 26636 | sidereon_rinex_lint_obs | positioning.RINEXLintObs | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 765 | 26640 | sidereon_rinex_lint_report_free | positioning.RINEXLintReport.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 766 | 26642 | sidereon_rinex_lint_summary | positioning.RINEXLintSummary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 767 | 26652 | sidereon_rinex_obs_carrier_phase | positioning.RINEXObsCarrierPhase | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 768 | 26667 | sidereon_rinex_obs_codes | positioning.RINEXObsCodes | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 769 | 26680 | sidereon_rinex_obs_epoch_count | positioning.RINEXObsEpochCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 770 | 26691 | sidereon_rinex_obs_epochs | positioning.RINEXObsEpochs | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 771 | 26704 | sidereon_rinex_obs_free | positioning.RINEXObs.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 772 | 26712 | sidereon_rinex_obs_header | positioning.RINEXObsHeader | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 773 | 26723 | sidereon_rinex_obs_load | positioning.RINEXObsLoad | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 774 | 26741 | sidereon_rinex_obs_observation | positioning.RINEXObsObservation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 775 | 26757 | sidereon_rinex_obs_parse | positioning.RINEXObsParse | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 776 | 26768 | sidereon_rinex_obs_pseudoranges | positioning.RINEXObsPseudoranges | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 777 | 26783 | sidereon_rinex_obs_receiver_clock_phase_deviations | positioning.RINEXObsReceiverClockPhaseDeviations | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 778 | 26799 | sidereon_rinex_obs_to_rinex_text | positioning.RINEXObsToRINEXText | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 779 | 26813 | sidereon_rinex_obs_values | positioning.RINEXObsValues | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 780 | 26826 | sidereon_rinex_obs_version | positioning.RINEXObsVersion | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 781 | 26829 | sidereon_rinex_repair_actions | positioning.RINEXRepairActions | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 782 | 26835 | sidereon_rinex_repair_crinex_text | positioning.RINEXRepairCrinexText | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 783 | 26841 | sidereon_rinex_repair_free | positioning.RINEXRepair.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 784 | 26843 | sidereon_rinex_repair_nav | positioning.RINEXRepairNav | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 785 | 26848 | sidereon_rinex_repair_obs | positioning.RINEXRepairObs | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 786 | 26853 | sidereon_rinex_repair_options_init | positioning.RINEXRepairOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 787 | 26855 | sidereon_rinex_repair_summary | positioning.RINEXRepairSummary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 788 | 26858 | sidereon_rinex_repair_text | positioning.RINEXRepairText | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 789 | 26869 | sidereon_rinex_spp_inputs_count | positioning.RINEXSPPInputsCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 790 | 26878 | sidereon_rinex_spp_inputs_epoch | positioning.RINEXSPPInputsEpoch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 791 | 26890 | sidereon_rinex_spp_inputs_epoch_inputs | positioning.RINEXSPPInputsEpochInputs | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 792 | 26900 | sidereon_rinex_spp_inputs_free | positioning.RINEXSPPInputs.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 793 | 26909 | sidereon_rinex_spp_options_init | positioning.RINEXSPPOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 794 | 26918 | sidereon_rinex_spp_solution | positioning.RINEXSPPSolution | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 795 | 26930 | sidereon_rinex_spp_solution_error | positioning.RINEXSPPSolutionError | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 796 | 26942 | sidereon_rinex_spp_solution_ok | positioning.RINEXSPPSolutionOk | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 797 | 26951 | sidereon_rinex_spp_solutions_count | positioning.RINEXSPPSolutionsCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 798 | 26960 | sidereon_rinex_spp_solutions_epoch | positioning.RINEXSPPSolutionsEpoch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 799 | 26970 | sidereon_rinex_spp_solutions_free | positioning.RINEXSPPSolutions.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 800 | 26972 | sidereon_robust_fde_solve_broadcast | positioning.RobustFDESolveBroadcast | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 801 | 26978 | sidereon_robust_fde_solve_spp | positioning.RobustFDESolveSPP | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 802 | 26994 | sidereon_rtcm_build_antenna_descriptor | positioning.RTCMBuildAntennaDescriptor | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 803 | 27012 | sidereon_rtcm_build_beidou_ephemeris | positioning.RTCMBuildBeidouEphemeris | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 804 | 27023 | sidereon_rtcm_build_galileo_fnav_ephemeris | positioning.RTCMBuildGalileoFnavEphemeris | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 805 | 27034 | sidereon_rtcm_build_galileo_inav_ephemeris | positioning.RTCMBuildGalileoInavEphemeris | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 806 | 27045 | sidereon_rtcm_build_glonass_ephemeris | positioning.RTCMBuildGLONASSEphemeris | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 807 | 27056 | sidereon_rtcm_build_gps_ephemeris | positioning.RTCMBuildGPSEphemeris | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 808 | 27070 | sidereon_rtcm_build_msm | positioning.RTCMBuildMsm | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 809 | 27085 | sidereon_rtcm_build_qzss_ephemeris | positioning.RTCMBuildQzssEphemeris | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 810 | 27097 | sidereon_rtcm_build_station_coordinates | positioning.RTCMBuildStationCoordinates | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 811 | 27109 | sidereon_rtcm_decode_frame | positioning.RTCMDecodeFrame | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 812 | 27126 | sidereon_rtcm_decode_messages | positioning.RTCMDecodeMessages | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 813 | 27141 | sidereon_rtcm_decode_stream | positioning.RTCMDecodeStream | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 814 | 27154 | sidereon_rtcm_derive_lli | positioning.RTCMDeriveLli | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 815 | 27169 | sidereon_rtcm_encode_frame | positioning.RTCMEncodeFrame | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 816 | 27183 | sidereon_rtcm_frame_body | positioning.RTCMFrameBody | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 817 | 27195 | sidereon_rtcm_frame_len | positioning.RTCMFrameLen | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 818 | 27204 | sidereon_rtcm_frames_count | positioning.RTCMFramesCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 819 | 27212 | sidereon_rtcm_frames_free | positioning.RTCMFrames.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 820 | 27220 | sidereon_rtcm_lli_bits | positioning.RTCMLliBits | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 821 | 27227 | sidereon_rtcm_lock_time_tracker_free | positioning.RTCMLockTimeTracker.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 822 | 27234 | sidereon_rtcm_lock_time_tracker_new | positioning.RTCMLockTimeTrackerNew | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 823 | 27244 | sidereon_rtcm_lock_time_tracker_observe | positioning.RTCMLockTimeTrackerObserve | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 824 | 27257 | sidereon_rtcm_lock_time_tracker_reset | positioning.RTCMLockTimeTrackerReset | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 825 | 27266 | sidereon_rtcm_message_antenna_descriptor | positioning.RTCMMessageAntennaDescriptor | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 826 | 27279 | sidereon_rtcm_message_antenna_string | positioning.RTCMMessageAntennaString | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 827 | 27292 | sidereon_rtcm_message_beidou_ephemeris | positioning.RTCMMessageBeidouEphemeris | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 828 | 27304 | sidereon_rtcm_message_encode | positioning.RTCMMessageEncode | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 829 | 27317 | sidereon_rtcm_message_galileo_fnav_ephemeris | positioning.RTCMMessageGalileoFnavEphemeris | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 830 | 27327 | sidereon_rtcm_message_galileo_inav_ephemeris | positioning.RTCMMessageGalileoInavEphemeris | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 831 | 27337 | sidereon_rtcm_message_glonass_ephemeris | positioning.RTCMMessageGLONASSEphemeris | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 832 | 27346 | sidereon_rtcm_message_gps_ephemeris | positioning.RTCMMessageGPSEphemeris | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 833 | 27356 | sidereon_rtcm_message_kind | positioning.RTCMMessageKind | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 834 | 27367 | sidereon_rtcm_message_msm_info | positioning.RTCMMessageMsmInfo | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 835 | 27379 | sidereon_rtcm_message_msm_satellites | positioning.RTCMMessageMsmSatellites | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 836 | 27394 | sidereon_rtcm_message_msm_signals | positioning.RTCMMessageMsmSignals | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 837 | 27406 | sidereon_rtcm_message_qzss_ephemeris | positioning.RTCMMessageQzssEphemeris | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 838 | 27410 | sidereon_rtcm_message_ssr_clocks | positioning.RTCMMessageSSRClocks | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 839 | 27417 | sidereon_rtcm_message_ssr_info | positioning.RTCMMessageSSRInfo | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 840 | 27421 | sidereon_rtcm_message_ssr_orbits | positioning.RTCMMessageSSROrbits | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 841 | 27428 | sidereon_rtcm_message_ssr_ura | positioning.RTCMMessageSSRUra | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 842 | 27441 | sidereon_rtcm_message_station_coordinates | positioning.RTCMMessageStationCoordinates | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 843 | 27453 | sidereon_rtcm_message_to_frame | positioning.RTCMMessageToFrame | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 844 | 27465 | sidereon_rtcm_messages_count | positioning.RTCMMessagesCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 845 | 27473 | sidereon_rtcm_messages_free | positioning.RTCMMessages.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 846 | 27483 | sidereon_rtcm_minimum_lock_time_ms | positioning.RTCMMinimumLockTimeMs | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 847 | 27494 | sidereon_rtcm_msm_epoch_dt_ms | positioning.RTCMMsmEpochDtMs | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 848 | 27508 | sidereon_rtcm_msm_signal_rinex_code | positioning.RTCMMsmSignalRINEXCode | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 849 | 27524 | sidereon_rtcm_scan_frames | positioning.RTCMScanFrames | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 850 | 27534 | sidereon_rtcm_stream_diagnostics_free | positioning.RTCMStreamDiagnostics.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 851 | 27541 | sidereon_rtcm_stream_diagnostics_resync_bytes | positioning.RTCMStreamDiagnosticsResyncBytes | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 852 | 27550 | sidereon_rtcm_stream_diagnostics_skipped_frame | positioning.RTCMStreamDiagnosticsSkippedFrame | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 853 | 27561 | sidereon_rtcm_stream_diagnostics_skipped_frame_message | positioning.RTCMStreamDiagnosticsSkippedFrameMessage | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 854 | 27574 | sidereon_rtcm_stream_diagnostics_skipped_frames_count | positioning.RTCMStreamDiagnosticsSkippedFramesCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 855 | 27584 | sidereon_rtk_arc_solution_dropped_sats | positioning.RTKArcSolutionDroppedSats | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 856 | 27597 | sidereon_rtk_arc_solution_elevation_masked_sats | positioning.RTKArcSolutionElevationMaskedSats | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 857 | 27608 | sidereon_rtk_arc_solution_epoch_count | positioning.RTKArcSolutionEpochCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 858 | 27616 | sidereon_rtk_arc_solution_epoch_metadata | positioning.RTKArcSolutionEpochMetadata | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 859 | 27627 | sidereon_rtk_arc_solution_epoch_sd_ambiguities | positioning.RTKArcSolutionEpochSdAmbiguities | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 860 | 27641 | sidereon_rtk_arc_solution_epoch_string_ids | positioning.RTKArcSolutionEpochStringIds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 861 | 27656 | sidereon_rtk_arc_solution_epoch_used_satellites | positioning.RTKArcSolutionEpochUsedSatellites | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 862 | 27668 | sidereon_rtk_arc_solution_final_baseline | positioning.RTKArcSolutionFinalBaseline | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 863 | 27677 | sidereon_rtk_arc_solution_final_epoch_count | positioning.RTKArcSolutionFinalEpochCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 864 | 27685 | sidereon_rtk_arc_solution_free | positioning.RTKArcSolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 865 | 27694 | sidereon_rtk_arc_solution_measurement_covariance | positioning.RTKArcSolutionMeasurementCovariance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 866 | 27707 | sidereon_rtk_arc_solution_references | positioning.RTKArcSolutionReferences | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 867 | 27720 | sidereon_rtk_arc_solution_split_cycle_slip_arcs | positioning.RTKArcSolutionSplitCycleSlipArcs | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 868 | 27731 | sidereon_rtk_arc_update_options_init | positioning.RTKArcUpdateOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 869 | 27738 | sidereon_rtk_fixed_options_init | positioning.RTKFixedOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 870 | 27748 | sidereon_rtk_fixed_solution_fixed_ambiguities | positioning.RTKFixedSolutionFixedAmbiguities | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 871 | 27760 | sidereon_rtk_fixed_solution_fixed_baseline_ecef | positioning.RTKFixedSolutionFixedBaselineECEF | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 872 | 27771 | sidereon_rtk_fixed_solution_fixed_baseline_enu | positioning.RTKFixedSolutionFixedBaselineEnu | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 873 | 27781 | sidereon_rtk_fixed_solution_float_baseline_ecef | positioning.RTKFixedSolutionFloatBaselineECEF | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 874 | 27793 | sidereon_rtk_fixed_solution_free | positioning.RTKFixedSolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 875 | 27803 | sidereon_rtk_fixed_solution_free_ambiguities | positioning.RTKFixedSolutionFreeAmbiguities | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 876 | 27815 | sidereon_rtk_fixed_solution_metadata | positioning.RTKFixedSolutionMetadata | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 877 | 27826 | sidereon_rtk_fixed_solution_used_sat_ids | positioning.RTKFixedSolutionUsedSatIds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 878 | 27837 | sidereon_rtk_float_options_init | positioning.RTKFloatOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 879 | 27847 | sidereon_rtk_float_solution_ambiguities | positioning.RTKFloatSolutionAmbiguities | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 880 | 27859 | sidereon_rtk_float_solution_baseline_ecef | positioning.RTKFloatSolutionBaselineECEF | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 881 | 27870 | sidereon_rtk_float_solution_baseline_enu | positioning.RTKFloatSolutionBaselineEnu | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 882 | 27882 | sidereon_rtk_float_solution_free | positioning.RTKFloatSolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 883 | 27890 | sidereon_rtk_float_solution_metadata | positioning.RTKFloatSolutionMetadata | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 884 | 27901 | sidereon_rtk_float_solution_used_sat_ids | positioning.RTKFloatSolutionUsedSatIds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 885 | 27915 | sidereon_rtk_ionosphere_free_arc_solution_epoch_base_observations | positioning.RTKIonosphereFreeArcSolutionEpochBaseObservations | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 886 | 27930 | sidereon_rtk_ionosphere_free_arc_solution_epoch_base_satellite_positions | positioning.RTKIonosphereFreeArcSolutionEpochBaseSatellitePositions | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 887 | 27942 | sidereon_rtk_ionosphere_free_arc_solution_epoch_count | positioning.RTKIonosphereFreeArcSolutionEpochCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 888 | 27951 | sidereon_rtk_ionosphere_free_arc_solution_epoch_metadata | positioning.RTKIonosphereFreeArcSolutionEpochMetadata | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 889 | 27963 | sidereon_rtk_ionosphere_free_arc_solution_epoch_rover_observations | positioning.RTKIonosphereFreeArcSolutionEpochRoverObservations | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 890 | 27978 | sidereon_rtk_ionosphere_free_arc_solution_epoch_rover_satellite_positions | positioning.RTKIonosphereFreeArcSolutionEpochRoverSatellitePositions | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 891 | 27993 | sidereon_rtk_ionosphere_free_arc_solution_epoch_satellite_positions | positioning.RTKIonosphereFreeArcSolutionEpochSatellitePositions | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 892 | 28006 | sidereon_rtk_ionosphere_free_arc_solution_free | positioning.RTKIonosphereFreeArcSolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 893 | 28014 | sidereon_rtk_ionosphere_free_arc_solution_offsets_m | positioning.RTKIonosphereFreeArcSolutionOffsetsM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 894 | 28027 | sidereon_rtk_ionosphere_free_arc_solution_references | positioning.RTKIonosphereFreeArcSolutionReferences | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 895 | 28039 | sidereon_rtk_ionosphere_free_arc_solution_wavelengths_m | positioning.RTKIonosphereFreeArcSolutionWavelengthsM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 896 | 28050 | sidereon_rtk_measurement_model_init | positioning.RTKMeasurementModelInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 897 | 28057 | sidereon_rtk_residual_validation_options_init | positioning.RTKResidualValidationOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 898 | 28064 | sidereon_rtk_rinex_arc_options_init | positioning.RTKRINEXArcOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 899 | 28071 | sidereon_rtk_rinex_dual_arc_options_init | positioning.RTKRINEXDualArcOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 900 | 28078 | sidereon_rtk_rinex_static_baseline_config_init | positioning.RTKRINEXStaticBaselineConfigInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 901 | 28085 | sidereon_rtk_rinex_wide_lane_fixed_config_init | positioning.RTKRINEXWideLaneFixedConfigInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 902 | 28093 | sidereon_rtk_static_arc_solution_ambiguity_ids | positioning.RTKStaticArcSolutionAmbiguityIds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 903 | 28107 | sidereon_rtk_static_arc_solution_ambiguity_satellites | positioning.RTKStaticArcSolutionAmbiguitySatellites | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 904 | 28120 | sidereon_rtk_static_arc_solution_dropped_sats | positioning.RTKStaticArcSolutionDroppedSats | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 905 | 28133 | sidereon_rtk_static_arc_solution_elevation_masked_sats | positioning.RTKStaticArcSolutionElevationMaskedSats | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 906 | 28146 | sidereon_rtk_static_arc_solution_fixed_ambiguities | positioning.RTKStaticArcSolutionFixedAmbiguities | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 907 | 28157 | sidereon_rtk_static_arc_solution_fixed_baseline_ecef | positioning.RTKStaticArcSolutionFixedBaselineECEF | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 908 | 28168 | sidereon_rtk_static_arc_solution_fixed_free_ambiguities | positioning.RTKStaticArcSolutionFixedFreeAmbiguities | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 909 | 28180 | sidereon_rtk_static_arc_solution_fixed_metadata | positioning.RTKStaticArcSolutionFixedMetadata | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 910 | 28190 | sidereon_rtk_static_arc_solution_float_ambiguities | positioning.RTKStaticArcSolutionFloatAmbiguities | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 911 | 28201 | sidereon_rtk_static_arc_solution_float_baseline_ecef | positioning.RTKStaticArcSolutionFloatBaselineECEF | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 912 | 28211 | sidereon_rtk_static_arc_solution_float_metadata | positioning.RTKStaticArcSolutionFloatMetadata | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 913 | 28219 | sidereon_rtk_static_arc_solution_free | positioning.RTKStaticArcSolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 914 | 28228 | sidereon_rtk_static_arc_solution_geometry_quality | positioning.RTKStaticArcSolutionGeometryQuality | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 915 | 28238 | sidereon_rtk_static_arc_solution_references | positioning.RTKStaticArcSolutionReferences | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 916 | 28250 | sidereon_rtk_static_arc_solution_split_cycle_slip_arcs | positioning.RTKStaticArcSolutionSplitCycleSlipArcs | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 917 | 28263 | sidereon_rtk_wide_lane_arc_solution_dropped_sats | positioning.RTKWideLaneArcSolutionDroppedSats | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 918 | 28274 | sidereon_rtk_wide_lane_arc_solution_epoch_count | positioning.RTKWideLaneArcSolutionEpochCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 919 | 28282 | sidereon_rtk_wide_lane_arc_solution_free | positioning.RTKWideLaneArcSolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 920 | 28291 | sidereon_rtk_wide_lane_arc_solution_geometry_quality | positioning.RTKWideLaneArcSolutionGeometryQuality | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 921 | 28301 | sidereon_rtk_wide_lane_arc_solution_references | positioning.RTKWideLaneArcSolutionReferences | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 922 | 28313 | sidereon_rtk_wide_lane_arc_solution_split_cycle_slip_arcs | positioning.RTKWideLaneArcSolutionSplitCycleSlipArcs | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 923 | 28326 | sidereon_rtk_wide_lane_arc_solution_wide_lane_cycles | positioning.RTKWideLaneArcSolutionWideLaneCycles | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 924 | 28337 | sidereon_rtk_wide_lane_fixed_rinex_solution_fixed_baseline_ecef | positioning.RTKWideLaneFixedRINEXSolutionFixedBaselineECEF | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 925 | 28347 | sidereon_rtk_wide_lane_fixed_rinex_solution_fixed_metadata | positioning.RTKWideLaneFixedRINEXSolutionFixedMetadata | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 926 | 28355 | sidereon_rtk_wide_lane_fixed_rinex_solution_float_baseline_ecef | positioning.RTKWideLaneFixedRINEXSolutionFloatBaselineECEF | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 927 | 28365 | sidereon_rtk_wide_lane_fixed_rinex_solution_float_metadata | positioning.RTKWideLaneFixedRINEXSolutionFloatMetadata | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 928 | 28375 | sidereon_rtk_wide_lane_fixed_rinex_solution_free | positioning.RTKWideLaneFixedRINEXSolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 929 | 28383 | sidereon_rtk_wide_lane_fixed_rinex_solution_metadata | positioning.RTKWideLaneFixedRINEXSolutionMetadata | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 930 | 28393 | sidereon_rtk_wide_lane_fixed_rinex_solution_wide_lane_cycles | positioning.RTKWideLaneFixedRINEXSolutionWideLaneCycles | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 931 | 28406 | sidereon_rtn_to_eci_covariance | errormetrics.RtnToECICovariance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 932 | 28420 | sidereon_rv2coe | support.Rv2coe | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 933 | 28425 | sidereon_rv2eq | support.Rv2eq | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 934 | 28431 | sidereon_rv2mee | support.Rv2mee | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 935 | 28450 | sidereon_satellite_constellation_build | astro.SatelliteConstellationBuild | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 936 | 28465 | sidereon_satellite_constellation_catalog_number | astro.SatelliteConstellationCatalogNumber | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 937 | 28481 | sidereon_satellite_constellation_free | astro.SatelliteConstellation.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 938 | 28494 | sidereon_satellite_constellation_ground_tracks | astro.SatelliteConstellationGroundTracks | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 939 | 28508 | sidereon_satellite_constellation_ground_tracks_free | astro.SatelliteConstellationGroundTracks.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 940 | 28516 | sidereon_satellite_constellation_ground_tracks_satellite_count | astro.SatelliteConstellationGroundTracksSatelliteCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 941 | 28527 | sidereon_satellite_constellation_ground_tracks_track_len | astro.SatelliteConstellationGroundTracksTrackLen | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 942 | 28542 | sidereon_satellite_constellation_ground_tracks_values | astro.SatelliteConstellationGroundTracksValues | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 943 | 28561 | sidereon_satellite_constellation_look_angle_arcs | astro.SatelliteConstellationLookAngleArcs | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 944 | 28576 | sidereon_satellite_constellation_look_angles_arc_len | astro.SatelliteConstellationLookAnglesArcLen | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 945 | 28589 | sidereon_satellite_constellation_look_angles_free | astro.SatelliteConstellationLookAngles.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 946 | 28597 | sidereon_satellite_constellation_look_angles_satellite_count | astro.SatelliteConstellationLookAnglesSatelliteCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 947 | 28611 | sidereon_satellite_constellation_look_angles_values | astro.SatelliteConstellationLookAnglesValues | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 948 | 28628 | sidereon_satellite_constellation_passes | astro.SatelliteConstellationPasses | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 949 | 28640 | sidereon_satellite_constellation_passes_count | astro.SatelliteConstellationPassesCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 950 | 28652 | sidereon_satellite_constellation_passes_free | astro.SatelliteConstellationPasses.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 951 | 28662 | sidereon_satellite_constellation_passes_values | astro.SatelliteConstellationPassesValues | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 952 | 28682 | sidereon_satellite_constellation_propagate | astro.SatelliteConstellationPropagate | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 953 | 28694 | sidereon_satellite_constellation_satellite_count | astro.SatelliteConstellationSatelliteCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 954 | 28709 | sidereon_satellite_constellation_visible | astro.SatelliteConstellationVisible | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 955 | 28721 | sidereon_satellite_visual_magnitude | astro.SatelliteVisualMagnitude | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 956 | 28732 | sidereon_sbas_airborne_model_aad_a | positioning.SBASAirborneModelAadA | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 957 | 28739 | sidereon_sbas_airborne_sigma_air_m | positioning.SBASAirborneSigmaAirM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 958 | 28743 | sidereon_sbas_block_decode | positioning.SBASBlockDecode | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 959 | 28748 | sidereon_sbas_block_encode | positioning.SBASBlockEncode | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 960 | 28761 | sidereon_sbas_block_fast_corrections | positioning.SBASBlockFastCorrections | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 961 | 28772 | sidereon_sbas_block_fast_degradation | positioning.SBASBlockFastDegradation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 962 | 28776 | sidereon_sbas_block_free | positioning.SBASBlock.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 963 | 28785 | sidereon_sbas_block_geo_nav | positioning.SBASBlockGeoNav | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 964 | 28796 | sidereon_sbas_block_igp_mask | positioning.SBASBlockIgpMask | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 965 | 28800 | sidereon_sbas_block_info | positioning.SBASBlockInfo | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 966 | 28810 | sidereon_sbas_block_integrity | positioning.SBASBlockIntegrity | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 967 | 28821 | sidereon_sbas_block_iono_delays | positioning.SBASBlockIonoDelays | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 968 | 28833 | sidereon_sbas_block_long_term_half_info | positioning.SBASBlockLongTermHalfInfo | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 969 | 28846 | sidereon_sbas_block_long_term_records | positioning.SBASBlockLongTermRecords | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 970 | 28861 | sidereon_sbas_block_mixed_fast_corrections | positioning.SBASBlockMixedFastCorrections | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 971 | 28872 | sidereon_sbas_block_prn_mask | positioning.SBASBlockPrnMask | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 972 | 28884 | sidereon_sbas_block_raw_data | positioning.SBASBlockRawData | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 973 | 28890 | sidereon_sbas_corrected_state | positioning.SBASCorrectedState | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 974 | 28905 | sidereon_sbas_degradation_params_none | positioning.SBASDegradationParamsNone | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 975 | 28912 | sidereon_sbas_k_multipliers_en_route_npa | positioning.SBASKMultipliersEnRouteNpa | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 976 | 28919 | sidereon_sbas_k_multipliers_precision_approach | positioning.SBASKMultipliersPrecisionApproach | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 977 | 28927 | sidereon_sbas_protection_levels | positioning.SBASProtectionLevels | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 978 | 28938 | sidereon_sbas_sigma_flt_m_for_udrei | positioning.SBASSigmaFltMForUdrei | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 979 | 28947 | sidereon_sbas_sigma_tropo_m | positioning.SBASSigmaTropoM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 980 | 28954 | sidereon_sbas_sis_error_sigma_m | positioning.SBASSisErrorSigmaM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 981 | 28957 | sidereon_sbas_solve_broadcast | positioning.SBASSolveBroadcast | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 982 | 28964 | sidereon_sbas_store_fast_correction | positioning.SBASStoreFastCorrection | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 983 | 28970 | sidereon_sbas_store_free | positioning.SBASStore.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 984 | 28972 | sidereon_sbas_store_geo_nav | positioning.SBASStoreGeoNav | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 985 | 28977 | sidereon_sbas_store_ingest | positioning.SBASStoreIngest | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 986 | 28982 | sidereon_sbas_store_iono_grid_igps | positioning.SBASStoreIonoGridIgps | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 987 | 28989 | sidereon_sbas_store_iono_slant_delay_m | positioning.SBASStoreIonoSlantDelayM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 988 | 28998 | sidereon_sbas_store_long_term_correction | positioning.SBASStoreLongTermCorrection | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 989 | 29004 | sidereon_sbas_store_new | positioning.SBASStoreNew | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 990 | 29006 | sidereon_sbas_store_preferred_geo | positioning.SBASStorePreferredGeo | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 991 | 29011 | sidereon_sbas_store_ready_geos | positioning.SBASStoreReadyGeos | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 992 | 29025 | sidereon_scenario_epoch_offsets | distribution.ScenarioEpochOffsets | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 993 | 29037 | sidereon_scenario_observations | distribution.ScenarioObservations | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 994 | 29049 | sidereon_scenario_receiver_truth | distribution.ScenarioReceiverTruth | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 995 | 29061 | sidereon_scenario_simulate_json | distribution.ScenarioSimulateJSON | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 996 | 29073 | sidereon_scenario_simulate_json_with_broadcast | positioning.ScenarioSimulateJSONWithBroadcast | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 997 | 29086 | sidereon_scenario_simulate_json_with_broadcast_and_ionex | positioning.ScenarioSimulateJSONWithBroadcastAndIonex | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 998 | 29100 | sidereon_scenario_simulate_json_with_ionex | positioning.ScenarioSimulateJSONWithIonex | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 999 | 29114 | sidereon_scenario_simulate_json_with_sp3 | positioning.ScenarioSimulateJSONWithSP3 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1000 | 29127 | sidereon_scenario_simulate_json_with_sp3_and_ionex | positioning.ScenarioSimulateJSONWithSP3AndIonex | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1001 | 29140 | sidereon_scenario_simulation_free | distribution.ScenarioSimulation.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1002 | 29148 | sidereon_scenario_simulation_json | distribution.ScenarioSimulationJSON | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1003 | 29160 | sidereon_scenario_simulation_summary | distribution.ScenarioSimulationSummary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1004 | 29169 | sidereon_scenario_terms | distribution.ScenarioTerms | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1005 | 29185 | sidereon_screen_tca_candidates_from_tle_catalog | astro.ScreenTCACandidatesFromTLECatalog | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1006 | 29211 | sidereon_screen_tca_conjunctions_from_tle_catalog | astro.ScreenTCAConjunctionsFromTLECatalog | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1007 | 29232 | sidereon_select_ionex | positioning.SelectIonex | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1008 | 29252 | sidereon_select_ionex_over_range | positioning.SelectIonexOverRange | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1009 | 29265 | sidereon_select_sp3 | positioning.SelectSP3 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1010 | 29285 | sidereon_select_sp3_over_range | positioning.SelectSP3OverRange | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1011 | 29298 | sidereon_sgp4_decay_latch_clear | astro.SGP4DecayLatchClear | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1012 | 29306 | sidereon_sgp4_decay_latch_first_failing_epoch | astro.SGP4DecayLatchFirstFailingEpoch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1013 | 29316 | sidereon_sgp4_decay_latch_free | astro.SGP4DecayLatch.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1014 | 29323 | sidereon_sgp4_decay_latch_new | astro.SGP4DecayLatchNew | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1015 | 29325 | sidereon_sgp4_fit_config_init | astro.SGP4FitConfigInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1016 | 29327 | sidereon_sgp4_fit_tle | astro.SGP4FitTLE | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1017 | 29332 | sidereon_sgp4_tle_fit_free | astro.SGP4TLEFit.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1018 | 29334 | sidereon_sgp4_tle_fit_lines | astro.SGP4TLEFitLines | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1019 | 29337 | sidereon_sgp4_tle_fit_omm | astro.SGP4TLEFitOMM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1020 | 29340 | sidereon_sgp4_tle_fit_statistics | astro.SGP4TLEFitStatistics | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1021 | 29351 | sidereon_sidereal_filter | astro.SiderealFilter | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1022 | 29362 | sidereon_sidereal_filter_options_init | astro.SiderealFilterOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1023 | 29369 | sidereon_sidereal_filter_output_coverage | astro.SiderealFilterOutputCoverage | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1024 | 29380 | sidereon_sidereal_filter_output_filtered | astro.SiderealFilterOutputFiltered | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1025 | 29391 | sidereon_sidereal_filter_output_free | astro.SiderealFilterOutput.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1026 | 29399 | sidereon_sidereal_filter_output_template | astro.SiderealFilterOutputTemplate | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1027 | 29410 | sidereon_sidereal_filter_output_under_covered | astro.SiderealFilterOutputUnderCovered | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1028 | 29422 | sidereon_sidereal_orbit_repeat_lag | astro.SiderealOrbitRepeatLag | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1029 | 29434 | sidereon_sidereal_periodicity_strength | astro.SiderealPeriodicityStrength | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1030 | 29448 | sidereon_sidereal_repeat_period | astro.SiderealRepeatPeriod | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1031 | 29460 | sidereon_sigmas | errormetrics.Sigmas | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1032 | 29477 | sidereon_signal_acquire | positioning.SignalAcquire | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1033 | 29493 | sidereon_signal_analysis_dll_jitter | positioning.SignalAnalysisDllJitter | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1034 | 29504 | sidereon_signal_analysis_dll_lower_bound | positioning.SignalAnalysisDllLowerBound | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1035 | 29515 | sidereon_signal_analysis_effective_cn0_degradation | positioning.SignalAnalysisEffectiveCn0Degradation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1036 | 29527 | sidereon_signal_analysis_fraction_power | positioning.SignalAnalysisFractionPower | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1037 | 29539 | sidereon_signal_analysis_multipath_envelope | positioning.SignalAnalysisMultipathEnvelope | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1038 | 29553 | sidereon_signal_analysis_psd | positioning.SignalAnalysisPsd | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1039 | 29562 | sidereon_signal_analysis_rms_bandwidth_hz | positioning.SignalAnalysisRmsBandwidthHz | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1040 | 29572 | sidereon_signal_analysis_spectral_separation | positioning.SignalAnalysisSpectralSeparation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1041 | 29585 | sidereon_signal_autocorrelation | positioning.SignalAutocorrelation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1042 | 29597 | sidereon_signal_betz_l1_receiver_bandwidth_hz | positioning.SignalBetzL1ReceiverBandwidthHz | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1043 | 29605 | sidereon_signal_ca_chip | positioning.SignalCAChip | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1044 | 29614 | sidereon_signal_ca_code | positioning.SignalCACode | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1045 | 29626 | sidereon_signal_coherent_loss | positioning.SignalCoherentLoss | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1046 | 29636 | sidereon_signal_coherent_loss_db | positioning.SignalCoherentLossDb | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1047 | 29647 | sidereon_signal_correlate | positioning.SignalCorrelate | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1048 | 29661 | sidereon_signal_correlate_against | positioning.SignalCorrelateAgainst | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1049 | 29677 | sidereon_signal_correlation_at | positioning.SignalCorrelationAt | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1050 | 29691 | sidereon_signal_cross_correlation | positioning.SignalCrossCorrelation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1051 | 29706 | sidereon_signal_dll_lower_bound | positioning.SignalDllLowerBound | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1052 | 29717 | sidereon_signal_dll_thermal_noise_jitter | positioning.SignalDllThermalNoiseJitter | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1053 | 29730 | sidereon_signal_effective_cn0_degradation | positioning.SignalEffectiveCn0Degradation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1054 | 29742 | sidereon_signal_fraction_power_in_band | positioning.SignalFractionPowerInBand | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1055 | 29751 | sidereon_signal_modulation_code_rate_hz | positioning.SignalModulationCodeRateHz | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1056 | 29761 | sidereon_signal_modulation_label | positioning.SignalModulationLabel | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1057 | 29776 | sidereon_signal_multipath_error_envelope | positioning.SignalMultipathErrorEnvelope | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1058 | 29790 | sidereon_signal_power_in_band | positioning.SignalPowerInBand | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1059 | 29800 | sidereon_signal_psd_hz | positioning.SignalPsdHz | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1060 | 29809 | sidereon_signal_reference_chip_rate_hz | positioning.SignalReferenceChipRateHz | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1061 | 29818 | sidereon_signal_replica | positioning.SignalReplica | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1062 | 29830 | sidereon_signal_rms_bandwidth_hz | positioning.SignalRmsBandwidthHz | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1063 | 29840 | sidereon_signal_snr_post_db | positioning.SignalSnrPostDb | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1064 | 29851 | sidereon_signal_spectral_separation_coefficient_db_hz | positioning.SignalSpectralSeparationCoefficientDbHz | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1065 | 29864 | sidereon_signal_spectral_separation_coefficient_hz | positioning.SignalSpectralSeparationCoefficientHz | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1066 | 29875 | sidereon_signal_white_noise_spectral_separation_hz | positioning.SignalWhiteNoiseSpectralSeparationHz | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1067 | 29888 | sidereon_smooth_code | support.SmoothCode | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1068 | 29903 | sidereon_smooth_fusion_rts | errormetrics.SmoothFusionRts | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1069 | 29915 | sidereon_smooth_iono_free_code | support.SmoothIonoFreeCode | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1070 | 29929 | sidereon_smooth_track_rts | errormetrics.SmoothTrackRts | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1071 | 29938 | sidereon_smoothed_fusion_trajectory_epoch | errormetrics.SmoothedFusionTrajectoryEpoch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1072 | 29947 | sidereon_smoothed_fusion_trajectory_epoch_count | errormetrics.SmoothedFusionTrajectoryEpochCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1073 | 29956 | sidereon_smoothed_fusion_trajectory_epoch_covariance | errormetrics.SmoothedFusionTrajectoryEpochCovariance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1074 | 29969 | sidereon_smoothed_fusion_trajectory_epoch_error_state_correction | errormetrics.SmoothedFusionTrajectoryEpochErrorStateCorrection | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1075 | 29982 | sidereon_smoothed_fusion_trajectory_epoch_position_ecef_m | errormetrics.SmoothedFusionTrajectoryEpochPositionECEFM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1076 | 29995 | sidereon_smoothed_fusion_trajectory_epoch_rts_gain_to_next | errormetrics.SmoothedFusionTrajectoryEpochRtsGainToNext | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1077 | 30008 | sidereon_smoothed_fusion_trajectory_free | errormetrics.SmoothedFusionTrajectory.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1078 | 30015 | sidereon_smoothed_track_epoch | errormetrics.SmoothedTrackEpoch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1079 | 30024 | sidereon_smoothed_track_epoch_count | errormetrics.SmoothedTrackEpochCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1080 | 30033 | sidereon_smoothed_track_epoch_covariance | errormetrics.SmoothedTrackEpochCovariance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1081 | 30046 | sidereon_smoothed_track_epoch_position_m | errormetrics.SmoothedTrackEpochPositionM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1082 | 30060 | sidereon_smoothed_track_epoch_rts_gain_to_next | errormetrics.SmoothedTrackEpochRtsGainToNext | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1083 | 30072 | sidereon_smoothed_track_free | errormetrics.SmoothedTrack.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1084 | 30080 | sidereon_solid_earth_pole_tide | geodesy.SolidEarthPoleTide | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1085 | 30096 | sidereon_solid_earth_tide | geodesy.SolidEarthTide | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1086 | 30110 | sidereon_solution_validation_options_init | support.SolutionValidationOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1087 | 30125 | sidereon_solve_broadcast | positioning.SolveBroadcast | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1088 | 30137 | sidereon_solve_broadcast_with_doppler_velocity | positioning.SolveBroadcastWithDopplerVelocity | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1089 | 30155 | sidereon_solve_data_problem | positioning.SolveDataProblem | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1090 | 30169 | sidereon_solve_data_problem_drop_one | positioning.SolveDataProblemDropOne | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1091 | 30172 | sidereon_solve_kepler | positioning.SolveKepler | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1092 | 30186 | sidereon_solve_moving_baseline | positioning.SolveMovingBaseline | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1093 | 30203 | sidereon_solve_ppp_auto_init_fixed | positioning.SolvePPPAutoInitFixed | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1094 | 30220 | sidereon_solve_ppp_auto_init_float | positioning.SolvePPPAutoInitFloat | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1095 | 30233 | sidereon_solve_ppp_fixed | positioning.SolvePPPFixed | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1096 | 30246 | sidereon_solve_ppp_float | positioning.SolvePPPFloat | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1097 | 30260 | sidereon_solve_rtk_arc | positioning.SolveRTKArc | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1098 | 30273 | sidereon_solve_rtk_fixed | positioning.SolveRTKFixed | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1099 | 30284 | sidereon_solve_rtk_float | positioning.SolveRTKFloat | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1100 | 30296 | sidereon_solve_spp | positioning.SolveSPP | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1101 | 30310 | sidereon_solve_spp_batch_parallel | positioning.SolveSPPBatchParallel | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1102 | 30326 | sidereon_solve_spp_batch_serial | positioning.SolveSPPBatchSerial | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1103 | 30344 | sidereon_solve_spp_from_rinex_obs | positioning.SolveSPPFromRINEXObs | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1104 | 30362 | sidereon_solve_spp_v2 | positioning.SolveSPPV2 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1105 | 30374 | sidereon_solve_spp_with_doppler_velocity | positioning.SolveSPPWithDopplerVelocity | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1106 | 30387 | sidereon_solve_static_position_broadcast | positioning.SolveStaticPositionBroadcast | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1107 | 30401 | sidereon_solve_static_position_sp3 | positioning.SolveStaticPositionSP3 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1108 | 30417 | sidereon_solve_static_reference_station_rinex | positioning.SolveStaticReferenceStationRINEX | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1109 | 30431 | sidereon_solve_static_rinex_rtk_baseline | positioning.SolveStaticRINEXRTKBaseline | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1110 | 30447 | sidereon_solve_static_rtk_arc | positioning.SolveStaticRTKArc | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1111 | 30469 | sidereon_solve_velocity | positioning.SolveVelocity | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1112 | 30491 | sidereon_solve_velocity_broadcast | positioning.SolveVelocityBroadcast | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1113 | 30507 | sidereon_solve_wide_lane_fixed_rinex_rtk_baseline | positioning.SolveWideLaneFixedRINEXRTKBaseline | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1114 | 30525 | sidereon_solve_with_fallback | positioning.SolveWithFallback | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1115 | 30538 | sidereon_source_crlb | support.SourceCrlb | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1116 | 30553 | sidereon_source_dop | errormetrics.SourceDOP | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1117 | 30566 | sidereon_source_locate_options_init | support.SourceLocateOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1118 | 30574 | sidereon_source_solution_covariance | errormetrics.SourceSolutionCovariance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1119 | 30584 | sidereon_source_solution_free | support.SourceSolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1120 | 30593 | sidereon_source_solution_influences | support.SourceSolutionInfluences | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1121 | 30605 | sidereon_source_solution_residuals | support.SourceSolutionResiduals | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1122 | 30616 | sidereon_source_solution_summary | support.SourceSolutionSummary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1123 | 30632 | sidereon_sourced_solution_broadcast_reason | positioning.SourcedSolutionBroadcastReason | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1124 | 30645 | sidereon_sourced_solution_free | distribution.SourcedSolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1125 | 30653 | sidereon_sourced_solution_is_precise_exact | positioning.SourcedSolutionIsPreciseExact | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1126 | 30664 | sidereon_sourced_solution_solution | distribution.SourcedSolutionSolution | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1127 | 30673 | sidereon_sourced_solution_source_kind | distribution.SourcedSolutionSourceKind | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1128 | 30687 | sidereon_sourced_solution_staleness | distribution.SourcedSolutionStaleness | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1129 | 30698 | sidereon_sp3_align_clock_reference | positioning.SP3AlignClockReference | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1130 | 30722 | sidereon_sp3_check_continuity | positioning.SP3CheckContinuity | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1131 | 30739 | sidereon_sp3_clock_reference_offsets | positioning.SP3ClockReferenceOffsets | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1132 | 30764 | sidereon_sp3_continuity_verdict_json | positioning.SP3ContinuityVerdictJSON | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1133 | 30784 | sidereon_sp3_declared_epoch_count | positioning.SP3DeclaredEpochCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1134 | 30799 | sidereon_sp3_declared_start_j2000_seconds | positioning.SP3DeclaredStartJ2000Seconds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1135 | 30812 | sidereon_sp3_emission_media_batch_at_j2000_s | positioning.SP3EmissionMediaBatchAtJ2000S | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1136 | 30836 | sidereon_sp3_ephemeris_sample | positioning.SP3EphemerisSample | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1137 | 30852 | sidereon_sp3_epoch_count | positioning.SP3EpochCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1138 | 30860 | sidereon_sp3_epoch_prediction | positioning.SP3EpochPrediction | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1139 | 30873 | sidereon_sp3_epochs_j2000_seconds | positioning.SP3EpochsJ2000Seconds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1140 | 30882 | sidereon_sp3_exact_request_free | positioning.SP3ExactRequest.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1141 | 30895 | sidereon_sp3_exact_request_from_identity | positioning.SP3ExactRequestFromIdentity | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1142 | 30912 | sidereon_sp3_exact_request_new | positioning.SP3ExactRequestNew | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1143 | 30930 | sidereon_sp3_free | positioning.SP3.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1144 | 30942 | sidereon_sp3_geometry_passes | positioning.SP3GeometryPasses | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1145 | 30965 | sidereon_sp3_geometry_visibility_series | positioning.SP3GeometryVisibilitySeries | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1146 | 30992 | sidereon_sp3_geometry_visible | positioning.SP3GeometryVisible | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1147 | 31012 | sidereon_sp3_interpolate | positioning.SP3Interpolate | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1148 | 31030 | sidereon_sp3_load | positioning.SP3Load | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1149 | 31047 | sidereon_sp3_load_exact | positioning.SP3LoadExact | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1150 | 31062 | sidereon_sp3_merge | positioning.SP3Merge | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1151 | 31085 | sidereon_sp3_merge_input_identity | positioning.SP3MergeInputIdentity | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1152 | 31093 | sidereon_sp3_merge_input_identity_contributor | positioning.SP3MergeInputIdentityContributor | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1153 | 31100 | sidereon_sp3_merge_input_identity_contributor_count | positioning.SP3MergeInputIdentityContributorCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1154 | 31106 | sidereon_sp3_merge_input_identity_free | positioning.SP3MergeInputIdentity.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1155 | 31111 | sidereon_sp3_merge_input_identity_precedence_contributor | positioning.SP3MergeInputIdentityPrecedenceContributor | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1156 | 31119 | sidereon_sp3_merge_input_identity_precedence_contributor_count | positioning.SP3MergeInputIdentityPrecedenceContributorCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1157 | 31126 | sidereon_sp3_merge_input_identity_schema_version | positioning.SP3MergeInputIdentitySchemaVersion | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1158 | 31133 | sidereon_sp3_merge_input_identity_stable_id | positioning.SP3MergeInputIdentityStableId | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1159 | 31144 | sidereon_sp3_merge_options_init | positioning.SP3MergeOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1160 | 31155 | sidereon_sp3_merge_report_agreement_summary | positioning.SP3MergeReportAgreementSummary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1161 | 31175 | sidereon_sp3_merge_report_continuity_verdict_json | positioning.SP3MergeReportContinuityVerdictJSON | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1162 | 31192 | sidereon_sp3_merge_report_epoch_agreement | positioning.SP3MergeReportEpochAgreement | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1163 | 31206 | sidereon_sp3_merge_report_epoch_agreement_count | positioning.SP3MergeReportEpochAgreementCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1164 | 31216 | sidereon_sp3_merge_report_flag | positioning.SP3MergeReportFlag | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1165 | 31228 | sidereon_sp3_merge_report_flag_count | positioning.SP3MergeReportFlagCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1166 | 31241 | sidereon_sp3_merge_report_flag_sources | positioning.SP3MergeReportFlagSources | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1167 | 31255 | sidereon_sp3_merge_report_frame_reconciliation | positioning.SP3MergeReportFrameReconciliation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1168 | 31265 | sidereon_sp3_merge_report_frame_reconciliation_asserted_label | positioning.SP3MergeReportFrameReconciliationAssertedLabel | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1169 | 31279 | sidereon_sp3_merge_report_frame_reconciliation_count | positioning.SP3MergeReportFrameReconciliationCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1170 | 31288 | sidereon_sp3_merge_report_frame_reconciliation_provenance | positioning.SP3MergeReportFrameReconciliationProvenance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1171 | 31301 | sidereon_sp3_merge_report_frame_reconciliation_source_label | positioning.SP3MergeReportFrameReconciliationSourceLabel | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1172 | 31314 | sidereon_sp3_merge_report_frame_reconciliation_target_label | positioning.SP3MergeReportFrameReconciliationTargetLabel | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1173 | 31329 | sidereon_sp3_merge_report_free | positioning.SP3MergeReport.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1174 | 31338 | sidereon_sp3_observable_state | positioning.SP3ObservableState | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1175 | 31359 | sidereon_sp3_observable_states_at_j2000_s | positioning.SP3ObservableStatesAtJ2000S | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1176 | 31375 | sidereon_sp3_observable_states_at_shared_j2000_s | positioning.SP3ObservableStatesAtSharedJ2000S | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1177 | 31395 | sidereon_sp3_observables | positioning.SP3Observables | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1178 | 31416 | sidereon_sp3_observables_batch | positioning.SP3ObservablesBatch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1179 | 31434 | sidereon_sp3_precise_ephemeris_samples | positioning.SP3PreciseEphemerisSamples | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1180 | 31448 | sidereon_sp3_precise_interpolant_artifact_bytes | positioning.SP3PreciseInterpolantArtifactBytes | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1181 | 31466 | sidereon_sp3_predict_ranges | positioning.SP3PredictRanges | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1182 | 31479 | sidereon_sp3_prediction_summary | positioning.SP3PredictionSummary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1183 | 31490 | sidereon_sp3_satellites | positioning.SP3Satellites | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1184 | 31504 | sidereon_sp3_state | positioning.SP3State | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1185 | 31519 | sidereon_sp3_stencil_extent | positioning.SP3StencilExtent | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1186 | 31532 | sidereon_sp3_to_sp3_text | positioning.SP3ToSP3Text | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1187 | 31547 | sidereon_sp3_validate_exact | positioning.SP3ValidateExact | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1188 | 31556 | sidereon_space_weather_default | distribution.SpaceWeatherDefault | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1189 | 31558 | sidereon_space_weather_table_ap_array_at | distribution.SpaceWeatherTableApArrayAt | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1190 | 31562 | sidereon_space_weather_table_coverage | distribution.SpaceWeatherTableCoverage | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1191 | 31565 | sidereon_space_weather_table_day | distribution.SpaceWeatherTableDay | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1192 | 31572 | sidereon_space_weather_table_days | distribution.SpaceWeatherTableDays | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1193 | 31578 | sidereon_space_weather_table_free | distribution.SpaceWeatherTable.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1194 | 31580 | sidereon_space_weather_table_monthly | distribution.SpaceWeatherTableMonthly | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1195 | 31586 | sidereon_space_weather_table_parse | distribution.SpaceWeatherTableParse | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1196 | 31590 | sidereon_space_weather_table_parse_csv | distribution.SpaceWeatherTableParseCsv | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1197 | 31594 | sidereon_space_weather_table_parse_txt | distribution.SpaceWeatherTableParseTxt | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1198 | 31598 | sidereon_space_weather_table_sample_at | distribution.SpaceWeatherTableSampleAt | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1199 | 31602 | sidereon_space_weather_table_sample_at_with_policy | distribution.SpaceWeatherTableSampleAtWithPolicy | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1200 | 31607 | sidereon_space_weather_table_space_weather_at | distribution.SpaceWeatherTableSpaceWeatherAt | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1201 | 31611 | sidereon_space_weather_table_summary | distribution.SpaceWeatherTableSummary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1202 | 31614 | sidereon_space_weather_table_to_csv | distribution.SpaceWeatherTableToCsv | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1203 | 31620 | sidereon_space_weather_table_to_txt | distribution.SpaceWeatherTableToTxt | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1204 | 31633 | sidereon_spk_free | astro.SPK.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1205 | 31643 | sidereon_spk_load | astro.SPKLoad | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1206 | 31657 | sidereon_spk_state | astro.SPKState | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1207 | 31669 | sidereon_split_jd_to_j2000_seconds | geodesy.SplitJdToJ2000Seconds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1208 | 31678 | sidereon_spp_batch_count | positioning.SPPBatchCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1209 | 31686 | sidereon_spp_batch_epoch_ok | positioning.SPPBatchEpochOk | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1210 | 31698 | sidereon_spp_batch_error | positioning.SPPBatchError | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1211 | 31710 | sidereon_spp_batch_free | positioning.SPPBatch.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1212 | 31721 | sidereon_spp_batch_solution | positioning.SPPBatchSolution | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1213 | 31731 | sidereon_spp_doppler_solution_free | positioning.SPPDopplerSolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1214 | 31739 | sidereon_spp_doppler_solution_has_velocity | positioning.SPPDopplerSolutionHasVelocity | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1215 | 31749 | sidereon_spp_doppler_solution_receiver | positioning.SPPDopplerSolutionReceiver | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1216 | 31759 | sidereon_spp_doppler_solution_velocity | positioning.SPPDopplerSolutionVelocity | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1217 | 31769 | sidereon_spp_doppler_solution_velocity_error_kind | positioning.SPPDopplerSolutionVelocityErrorKind | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1218 | 31782 | sidereon_spp_inputs_from_rinex_obs | positioning.SPPInputsFromRINEXObs | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1219 | 31794 | sidereon_spp_inputs_v2_init | positioning.SPPInputsV2Init | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1220 | 31804 | sidereon_spp_solution_dop | positioning.SPPSolutionDOP | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1221 | 31816 | sidereon_spp_solution_free | positioning.SPPSolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1222 | 31826 | sidereon_spp_solution_geodetic | positioning.SPPSolutionGeodetic | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1223 | 31836 | sidereon_spp_solution_metadata | positioning.SPPSolutionMetadata | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1224 | 31846 | sidereon_spp_solution_position | positioning.SPPSolutionPosition | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1225 | 31856 | sidereon_spp_solution_position_covariance_ecef_m2 | positioning.SPPSolutionPositionCovarianceECEFM2 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1226 | 31866 | sidereon_spp_solution_position_covariance_enu_m2 | positioning.SPPSolutionPositionCovarianceEnuM2 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1227 | 31878 | sidereon_spp_solution_rejected_sats | positioning.SPPSolutionRejectedSats | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1228 | 31897 | sidereon_spp_solution_residuals | positioning.SPPSolutionResiduals | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1229 | 31910 | sidereon_spp_solution_rx_clock_drift_s_s | positioning.SPPSolutionRxClockDriftSS | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1230 | 31920 | sidereon_spp_solution_rx_clock_s | positioning.SPPSolutionRxClockS | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1231 | 31931 | sidereon_spp_solution_system_clocks | positioning.SPPSolutionSystemClocks | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1232 | 31948 | sidereon_spp_solution_system_tdops | positioning.SPPSolutionSystemTdops | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1233 | 31960 | sidereon_spp_solution_used_sat_count | positioning.SPPSolutionUsedSatCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1234 | 31971 | sidereon_spp_solution_used_sat_ids | positioning.SPPSolutionUsedSatIds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1235 | 31977 | sidereon_ssr_corrected_state | positioning.SSRCorrectedState | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1236 | 31989 | sidereon_ssr_ephemeris_sample | positioning.SSREphemerisSample | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1237 | 32005 | sidereon_ssr_solve_broadcast | positioning.SSRSolveBroadcast | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1238 | 32014 | sidereon_ssr_store_clock | positioning.SSRStoreClock | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1239 | 32019 | sidereon_ssr_store_code_bias_m | positioning.SSRStoreCodeBiasM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1240 | 32025 | sidereon_ssr_store_free | positioning.SSRStore.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1241 | 32027 | sidereon_ssr_store_from_rtcm | positioning.SSRStoreFromRTCM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1242 | 32032 | sidereon_ssr_store_ingest_messages | positioning.SSRStoreIngestMessages | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1243 | 32036 | sidereon_ssr_store_new | positioning.SSRStoreNew | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1244 | 32039 | sidereon_ssr_store_orbit | positioning.SSRStoreOrbit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1245 | 32044 | sidereon_ssr_store_phase_bias_m | positioning.SSRStorePhaseBiasM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1246 | 32050 | sidereon_ssr_store_ura_index | positioning.SSRStoreUraIndex | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1247 | 32058 | sidereon_staleness_policy_days | distribution.StalenessPolicyDays | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1248 | 32063 | sidereon_staleness_policy_default | distribution.StalenessPolicyDefault | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1249 | 32068 | sidereon_staleness_policy_seconds | distribution.StalenessPolicySeconds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1250 | 32075 | sidereon_state_propagation_config_init | support.StatePropagationConfigInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1251 | 32082 | sidereon_static_position_options_init | positioning.StaticPositionOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1252 | 32090 | sidereon_static_position_solution_clock_biases | positioning.StaticPositionSolutionClockBiases | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1253 | 32102 | sidereon_static_position_solution_epoch_influence | positioning.StaticPositionSolutionEpochInfluence | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1254 | 32114 | sidereon_static_position_solution_free | positioning.StaticPositionSolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1255 | 32122 | sidereon_static_position_solution_geodetic | positioning.StaticPositionSolutionGeodetic | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1256 | 32132 | sidereon_static_position_solution_metadata | positioning.StaticPositionSolutionMetadata | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1257 | 32141 | sidereon_static_position_solution_position | positioning.StaticPositionSolutionPosition | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1258 | 32151 | sidereon_static_position_solution_position_covariance_ecef_m2 | positioning.StaticPositionSolutionPositionCovarianceECEFM2 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1259 | 32161 | sidereon_static_position_solution_position_covariance_enu_m2 | positioning.StaticPositionSolutionPositionCovarianceEnuM2 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1260 | 32172 | sidereon_static_position_solution_rejected_sats | positioning.StaticPositionSolutionRejectedSats | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1261 | 32185 | sidereon_static_position_solution_residuals | positioning.StaticPositionSolutionResiduals | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1262 | 32198 | sidereon_static_position_solution_satellite_batch_influence | positioning.StaticPositionSolutionSatelliteBatchInfluence | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1263 | 32211 | sidereon_static_position_solution_satellite_influence | positioning.StaticPositionSolutionSatelliteInfluence | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1264 | 32224 | sidereon_static_position_solution_state_covariance_m2 | positioning.StaticPositionSolutionStateCovarianceM2 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1265 | 32236 | sidereon_static_reference_station_rinex_config_init | positioning.StaticReferenceStationRINEXConfigInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1266 | 32243 | sidereon_static_reference_station_solution_baseline_ecef | positioning.StaticReferenceStationSolutionBaselineECEF | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1267 | 32252 | sidereon_static_reference_station_solution_covariance_ecef | positioning.StaticReferenceStationSolutionCovarianceECEF | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1268 | 32261 | sidereon_static_reference_station_solution_covariance_enu | positioning.StaticReferenceStationSolutionCovarianceEnu | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1269 | 32273 | sidereon_static_reference_station_solution_diagnostics | positioning.StaticReferenceStationSolutionDiagnostics | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1270 | 32285 | sidereon_static_reference_station_solution_free | positioning.StaticReferenceStationSolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1271 | 32293 | sidereon_static_reference_station_solution_metadata | positioning.StaticReferenceStationSolutionMetadata | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1272 | 32304 | sidereon_static_reference_station_solution_mode_reports | positioning.StaticReferenceStationSolutionModeReports | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1273 | 32315 | sidereon_static_reference_station_solution_position_ecef | positioning.StaticReferenceStationSolutionPositionECEF | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1274 | 32328 | sidereon_status_message | core.Status.String | core status/version API; no omission |
| 1275 | 32339 | sidereon_sub_observer_point | support.SubObserverPoint | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1276 | 32353 | sidereon_sub_solar_point | support.SubSolarPoint | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1277 | 32362 | sidereon_sun_angle_deg | astro.SunAngleDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1278 | 32373 | sidereon_sun_az_el | astro.SunAzEl | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1279 | 32383 | sidereon_sun_elevation_deg | astro.SunElevationDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1280 | 32394 | sidereon_sun_moon_ecef | astro.SunMoonECEF | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1281 | 32406 | sidereon_sun_moon_ecef_batch | astro.SunMoonECEFBatch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1282 | 32420 | sidereon_sun_moon_eci | astro.SunMoonECI | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1283 | 32432 | sidereon_sun_moon_eci_batch | astro.SunMoonECIBatch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1284 | 32447 | sidereon_tai_utc_offset_s | geodesy.TAIUtcOffsetS | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1285 | 32456 | sidereon_tca_collision_probability | astro.TCACollisionProbability | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1286 | 32465 | sidereon_tca_finder_options_init | astro.TCAFinderOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1287 | 32472 | sidereon_tdm_free | support.Tdm.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1288 | 32481 | sidereon_tdm_parse_kvn | support.TdmParseKVN | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1289 | 32491 | sidereon_tdm_participants | support.TdmParticipants | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1290 | 32503 | sidereon_tdm_paths | support.TdmPaths | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1291 | 32514 | sidereon_tdm_record_count | support.TdmRecordCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1292 | 32522 | sidereon_tdm_records | support.TdmRecords | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1293 | 32533 | sidereon_tdm_segment_count | support.TdmSegmentCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1294 | 32541 | sidereon_tdm_segments | support.TdmSegments | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1295 | 32554 | sidereon_tdm_to_kvn | support.TdmToKVN | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1296 | 32567 | sidereon_terminator_latitude_deg | support.TerminatorLatitudeDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1297 | 32578 | sidereon_terrain_geoid_model_label | geodesy.TerrainGeoidModelLabel | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1298 | 32590 | sidereon_terrain_store_checksum64 | geodesy.TerrainStoreChecksum64 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1299 | 32601 | sidereon_time_scale_abbrev | geodesy.TimeScaleAbbrev | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1300 | 32617 | sidereon_timescale_offset_at_s | geodesy.TimescaleOffsetAtS | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1301 | 32632 | sidereon_timescale_offset_s | geodesy.TimescaleOffsetS | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1302 | 32640 | sidereon_timescales_from_utc | geodesy.TimescalesFromUtc | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1303 | 32658 | sidereon_tle_batch_look_angles | astro.TLEBatchLookAngles | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1304 | 32676 | sidereon_tle_batch_look_angles_free | astro.TLEBatchLookAngles.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1305 | 32685 | sidereon_tle_batch_look_angles_shape | astro.TLEBatchLookAnglesShape | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1306 | 32697 | sidereon_tle_batch_look_angles_values | astro.TLEBatchLookAnglesValues | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1307 | 32712 | sidereon_tle_batch_propagation_free | astro.TLEBatchPropagation.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1308 | 32720 | sidereon_tle_batch_propagation_shape | astro.TLEBatchPropagationShape | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1309 | 32732 | sidereon_tle_batch_propagation_states | astro.TLEBatchPropagationStates | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1310 | 32746 | sidereon_tle_checksum_warnings | astro.TLEChecksumWarnings | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1311 | 32758 | sidereon_tle_file_count | astro.TLEFileCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1312 | 32769 | sidereon_tle_file_free | astro.TLEFile.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1313 | 32782 | sidereon_tle_file_name | astro.TLEFileName | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1314 | 32797 | sidereon_tle_file_satellite | astro.TLEFileSatellite | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1315 | 32809 | sidereon_tle_file_skipped | astro.TLEFileSkipped | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1316 | 32820 | sidereon_tle_find_passes | astro.TLEFindPasses | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1317 | 32834 | sidereon_tle_free | astro.TLE.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1318 | 32844 | sidereon_tle_ground_track | astro.TLEGroundTrack | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1319 | 32857 | sidereon_tle_load | astro.TLELoad | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1320 | 32871 | sidereon_tle_look_angles | astro.TLELookAngles | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1321 | 32883 | sidereon_tle_metadata | astro.TLEMetadata | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1322 | 32894 | sidereon_tle_propagate | astro.TLEPropagate | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1323 | 32909 | sidereon_tle_propagate_with_decay_latch | astro.TLEPropagateWithDecayLatch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1324 | 32919 | sidereon_tle_propagation_epoch_count | astro.TLEPropagationEpochCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1325 | 32931 | sidereon_tle_propagation_free | astro.TLEPropagation.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1326 | 32941 | sidereon_tle_propagation_states | astro.TLEPropagationStates | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1327 | 32953 | sidereon_tle_to_lines | astro.TLEToLines | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1328 | 32961 | sidereon_track_filter_config_dimension | errormetrics.TrackFilterConfigDimension | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1329 | 32969 | sidereon_track_filter_config_frame | astro.TrackFilterConfigFrame | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1330 | 32977 | sidereon_track_filter_config_free | errormetrics.TrackFilterConfig.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1331 | 32986 | sidereon_track_filter_config_from_position | errormetrics.TrackFilterConfigFromPosition | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1332 | 33004 | sidereon_track_filter_config_from_position_velocity | errormetrics.TrackFilterConfigFromPositionVelocity | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1333 | 33020 | sidereon_track_filter_covariance | errormetrics.TrackFilterCovariance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1334 | 33031 | sidereon_track_filter_free | errormetrics.TrackFilter.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1335 | 33038 | sidereon_track_filter_new | errormetrics.TrackFilterNew | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1336 | 33048 | sidereon_track_filter_new_from_position | errormetrics.TrackFilterNewFromPosition | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1337 | 33067 | sidereon_track_filter_position_innovation | errormetrics.TrackFilterPositionInnovation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1338 | 33084 | sidereon_track_filter_position_m | errormetrics.TrackFilterPositionM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1339 | 33095 | sidereon_track_filter_predict | errormetrics.TrackFilterPredict | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1340 | 33104 | sidereon_track_filter_predict_recorded | errormetrics.TrackFilterPredictRecorded | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1341 | 33114 | sidereon_track_filter_record_prediction_only | errormetrics.TrackFilterRecordPredictionOnly | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1342 | 33122 | sidereon_track_filter_state | errormetrics.TrackFilterState | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1343 | 33133 | sidereon_track_filter_state_innovation | errormetrics.TrackFilterStateInnovation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1344 | 33150 | sidereon_track_filter_state_vector | errormetrics.TrackFilterStateVector | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1345 | 33162 | sidereon_track_filter_update_position | errormetrics.TrackFilterUpdatePosition | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1346 | 33175 | sidereon_track_filter_update_position_gated | errormetrics.TrackFilterUpdatePositionGated | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1347 | 33190 | sidereon_track_filter_update_position_gated_recorded | errormetrics.TrackFilterUpdatePositionGatedRecorded | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1348 | 33205 | sidereon_track_filter_update_position_recorded | errormetrics.TrackFilterUpdatePositionRecorded | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1349 | 33219 | sidereon_track_filter_update_state | errormetrics.TrackFilterUpdateState | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1350 | 33232 | sidereon_track_filter_velocity_m_s | errormetrics.TrackFilterVelocityMS | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1351 | 33243 | sidereon_track_rts_history_builder_finish | errormetrics.TrackRtsHistoryBuilderFinish | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1352 | 33251 | sidereon_track_rts_history_builder_free | errormetrics.TrackRtsHistoryBuilder.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1353 | 33258 | sidereon_track_rts_history_builder_from_filter | errormetrics.TrackRtsHistoryBuilderFromFilter | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1354 | 33266 | sidereon_track_rts_history_builder_new | errormetrics.TrackRtsHistoryBuilderNew | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1355 | 33273 | sidereon_track_rts_history_epoch | errormetrics.TrackRtsHistoryEpoch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1356 | 33282 | sidereon_track_rts_history_epoch_count | errormetrics.TrackRtsHistoryEpochCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1357 | 33291 | sidereon_track_rts_history_epoch_predicted_position_m | errormetrics.TrackRtsHistoryEpochPredictedPositionM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1358 | 33305 | sidereon_track_rts_history_epoch_transition_from_previous | errormetrics.TrackRtsHistoryEpochTransitionFromPrevious | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1359 | 33318 | sidereon_track_rts_history_epoch_updated_position_m | errormetrics.TrackRtsHistoryEpochUpdatedPositionM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1360 | 33330 | sidereon_track_rts_history_free | errormetrics.TrackRtsHistory.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1361 | 33338 | sidereon_trls_drop_one_base_summary | support.TrlsDropOneBaseSummary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1362 | 33351 | sidereon_trls_drop_one_cost_delta | support.TrlsDropOneCostDelta | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1363 | 33363 | sidereon_trls_drop_one_count | support.TrlsDropOneCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1364 | 33373 | sidereon_trls_drop_one_drop_summary | support.TrlsDropOneDropSummary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1365 | 33385 | sidereon_trls_drop_one_drop_x | support.TrlsDropOneDropX | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1366 | 33398 | sidereon_trls_drop_one_free | support.TrlsDropOne.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1367 | 33406 | sidereon_trls_solution_free | support.TrlsSolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1368 | 33416 | sidereon_trls_solution_gradient | support.TrlsSolutionGradient | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1369 | 33430 | sidereon_trls_solution_jacobian | support.TrlsSolutionJacobian | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1370 | 33444 | sidereon_trls_solution_residuals | support.TrlsSolutionResiduals | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1371 | 33457 | sidereon_trls_solution_summary | support.TrlsSolutionSummary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1372 | 33468 | sidereon_trls_solution_x | support.TrlsSolutionX | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1373 | 33482 | sidereon_tropo_mapping_factors | geodesy.TropoMappingFactors | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1374 | 33495 | sidereon_tropo_mapping_factors_checked | geodesy.TropoMappingFactorsChecked | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1375 | 33510 | sidereon_tropo_slant_delay | geodesy.TropoSlantDelay | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1376 | 33527 | sidereon_tropo_zenith_delay | geodesy.TropoZenithDelay | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1377 | 33531 | sidereon_true_to_eccentric_anomaly | support.TrueToEccentricAnomaly | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1378 | 33535 | sidereon_true_to_mean_anomaly | astro.TrueToMeanAnomaly | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1379 | 33544 | sidereon_ut1_coverage_covers_jd_tt | geodesy.UT1CoverageCoversJdTt | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1380 | 33551 | sidereon_ut1_coverage_info | geodesy.UT1CoverageInfo | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1381 | 33559 | sidereon_ut1_coverage_source | geodesy.UT1CoverageSource | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1382 | 33573 | sidereon_validate_receiver_solution | support.ValidateReceiverSolution | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1383 | 33582 | sidereon_velocity_options_init | support.VelocityOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1384 | 33590 | sidereon_velocity_solution_clock_drift | errormetrics.VelocitySolutionClockDrift | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1385 | 33600 | sidereon_velocity_solution_free | support.VelocitySolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1386 | 33611 | sidereon_velocity_solution_residuals | support.VelocitySolutionResiduals | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1387 | 33623 | sidereon_velocity_solution_speed | support.VelocitySolutionSpeed | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1388 | 33635 | sidereon_velocity_solution_state_covariance | errormetrics.VelocitySolutionStateCovariance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1389 | 33645 | sidereon_velocity_solution_used_sat_count | support.VelocitySolutionUsedSatCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1390 | 33656 | sidereon_velocity_solution_used_sat_ids | support.VelocitySolutionUsedSatIds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1391 | 33669 | sidereon_velocity_solution_velocity | support.VelocitySolutionVelocity | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1392 | 33680 | sidereon_version | core.Version | core status/version API; no omission |
| 1393 | 33686 | sidereon_version_string | core.VersionString | core status/version API; no omission |
| 1394 | 33694 | sidereon_vertical_datum_label | geodesy.VerticalDatumLabel | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1395 | 33717 | sidereon_visible_from_satellites | astro.VisibleFromSatellites | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1396 | 33731 | sidereon_visible_list_count | astro.VisibleListCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1397 | 33743 | sidereon_visible_list_free | astro.VisibleList.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1398 | 33753 | sidereon_visible_list_values | astro.VisibleListValues | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1399 | 33765 | sidereon_wavelength_m | support.WavelengthM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1400 | 33776 | sidereon_weight_vector | errormetrics.WeightVector | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1401 | 33788 | sidereon_write_dted_tree_to_mmap_store | distribution.WriteDTEDTreeToMmapStore | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1402 | 33795 | sidereon_wtest_noncentrality | errormetrics.WtestNoncentrality | C call bridged; non-OK maps to typed StatusError; outputs copied |

End of map: 1,402 C declarations accounted for; the C inventory is complete, but the documented-parity gate remains STOP because of the confirmed ABI gaps above. See AUDIT_EVIDENCE.md for reproducible commands, totals, and review ambiguities.
