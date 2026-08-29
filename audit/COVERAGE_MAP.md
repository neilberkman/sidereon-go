# sidereon-go coverage map

Gate result: PASS

This static audit covers the public C ABI at exact commit `b38ecf8caf796a02f209dbb4cbebdaa4a042204c` and the documented Python v1.1.1 surface. The current header and Rust export inventories each contain 1,503 unique `sidereon_*` functions, and their sorted symbol sets are identical. All documented capabilities are expressible through this ABI; the public Go implementation gate is clear. No Go implementation code is written here.

## Scope and pins

| Item | Evidence |
| --- | --- |
| C commit | `../../repos/sidereon-c` HEAD `b38ecf8caf796a02f209dbb4cbebdaa4a042204c` |
| C header | `bindings/c/include/sidereon.h`; SHA-256 `6e79405bbce65d91958fe591279ad4c73f79f86573c64176f7c3dd0d9c29420d`; 35,760 lines |
| C version | macros `1.2.0` at header lines 66-69; crate version 1.2.0 |
| Rust source | `bindings/c/src`; 1,503 `pub extern "C" fn` exports |
| Python baseline | `../../repos/sidereon-python` v1.1.1, HEAD `53be271eaf9974b3cf451065d0b0225f436bb944` |
| Python stub | `python/sidereon/__init__.pyi`; SHA-256 `24d0712209e561c03f52148006540a151e12aa048dc3bcd7ceb845f87be848f9` |
| Public core semantic source | `../../repos/sidereon` v1.1.1, HEAD `267b28f69f7678921c11689531f8fa250e7cca78` |

The header still reports 1.2.0 although the future Go release target is 1.3.0. That is expected pre-release state of the pinned reference, not a capability gap.

## Gate decision

PASS. Every current C declaration has exactly one row below. Rust exports match the header with a zero sorted symbol-set diff. The former eleven blockers, bare SSR decode/bias access, and SBAS PRN lookup were rechecked for inputs, outputs, accessors, ownership, and status behavior. No true ABI parity gap remains.

The documented reverse helper `satellite_id_to_sbas_prn` has no reverse-named C symbol. It is expressible as a Go convenience over the exact finite C forward route `sidereon_sbas_prn_to_satellite_id`: test the documented SBAS PRN domain and compare the copied identifier. This preserves the public mapping without introducing a second numeric implementation.

## Wrapper policy

Proposed packages are positioning, astro, distribution, geodesy, errormetrics, and support. The map is a bridge inventory, not an implementation claim.

C statuses map to a typed Go `StatusError` carrying the enum, stable status text, and thread-local detail. The eight statuses are OK, NULL_POINTER, INVALID_ARGUMENT, INVALID_TOKEN, SP3_PARSE, SOLVE, PANIC, and TIMEOUT. Non-OK is an error; failed constructors transfer no handle.

Owning opaque C pointers become non-copying Go handles with explicit `Close`, pointer clearing, and a `runtime.AddCleanup` backstop. Borrowed structs and variable arrays are copied before release. Variable output uses query-then-copy: query required length, allocate Go storage, copy, and reject short buffers as invalid arguments. Inputs are copied into temporary C storage; C retains no Go pointer.

The C error string is thread-local, so the fallible call and immediate error read must remain on one locked OS thread. Live read-only handles may be shared; mutation and `Close` must be serialized, and callers must not race a read with `Close`. Go owns sockets, TLS, retries, filesystem paths, cache locks, decompression, and presentation aliases; C owns byte parsing, numeric evaluation, and protocol state.

## Python documented-surface parity

The package implementation has 1,194 unique `__all__` names from `__init__.py:1134-2530`. The main stub has 701 classes and 465 function declarations (464 unique names because `raim` is overloaded). Auxiliary stubs add 81 classes and 62 functions. Methods and properties are included in the capability review below.

| Python capability family | Python evidence | C ABI route and semantic coverage | Go disposition |
| --- | --- | --- | --- |
| SP3, exact products, interpolation, merge | SP3 declarations; `exact_sp3.pyi:1-47` | `sidereon_sp3_*`, precise artifacts, exact-cache and product routes; owned handles and copied arrays | positioning; expressible |
| SPP, batch, Doppler, static, DGNSS | SPP/DGNSS declarations | `sidereon_spp_*`, solve, DGNSS and static routes | positioning; expressible |
| RTK, PPP, ionosphere-free, wide-lane, RINEX arcs | RTK/PPP declarations; arc stubs at `__init__.pyi:3428-3717` | RTK/PPP/combination routes and current arc builders/accessors | positioning; expressible |
| Broadcast GPS/Galileo/GLONASS/CNAV and NAV | broadcast/RINEX declarations | broadcast full records, GLONASS records/channels, NAV parse records and skipped diagnostics | positioning; expressible |
| Observations, RINEX observation, carrier/signal, NMEA, RTCM | observation/protocol declarations | observation, RINEX, carrier, signal, NMEA, RTCM routes; decoded slices copied | positioning/distribution; expressible |
| TLE, SGP4, propagation, decay, orbit utilities | TLE/orbit declarations | TLE, SGP4, propagation, orbit and decay routes | astro; expressible |
| Frames, time scales, EOP, Sun/Moon, almanac | frame/time/body declarations | frame, timescale, UT1, Sun/Moon and almanac routes | astro/geodesy; expressible |
| Passes, visibility, geometry, ground tracks, eclipses | pass/visibility declarations | pass, visible, geometry, ground-track and eclipse routes | astro; expressible |
| TCA/conjunction, covariance, CDM/OMM/OEM/OPM | conjunction/codec declarations | TCA, codec, covariance transport/propagation and complete current 6x6 routes | astro/errormetrics; expressible |
| Geodesy, geoid, terrain, DTED, antenna/ANTEX | geodesy/terrain declarations | geodetic, geodesic, geoid, geofence, DTED, terrain, antenna and ANTEX routes | geodesy; file adapters in Go; expressible |
| Atmosphere, troposphere, ionosphere, RF | atmosphere/RF declarations | atmosphere, tropo, iono, RF, signal and antenna routes | geodesy/positioning; expressible |
| Error metrics, DOP, covariance, clock/Allan | metric/clock declarations | metrics, residual, DOP, covariance6, clock and Allan routes | errormetrics; expressible |
| RAIM/FDE, robust solves, filters, NIS | estimation declarations | RAIM/FDE, chi-square, filters, Hessian, NIS, EWMA and CFAR routes | errormetrics; expressible |
| IOD, fitting, forces, Lambert, relative/CW | orbit/force declarations | IOD, fitting, force, drag, Lambert, relative and CW routes | astro; expressible |
| SBAS, SSR, RTCM | SBAS/SSR declarations | SBAS blocks/store, line parsers, bare SSR message accessors and RTCM routes | distribution; expressible |
| NTRIP and sourcetable | `ntrip.pyi:1-86` | sans-I/O NTRIP state and bytes; no C socket is required | Go socket/TLS adapter; expressible |
| Data catalog, cache, acquisition, metadata | `data.pyi:1-490`, `distribution.pyi:1-185`, `exact_cache.pyi:1-70` | data, cache, product and validation routes | Go I/O/cache adapter around C; expressible |
| JSON/XML/KVN, paths, aliases | codec/path declarations | serializer and byte routes where present | Go presentation/I/O adapters; expressible |

## Former blockers and current additions

| Capability | Python source | C header / Rust source | Verified semantic completeness |
| --- | --- | --- | --- |
| `build_rinex_rtk_arc` | `__init__.pyi:3482`; `src/rtk.rs:2796` | header:19718; C Rust `bindings/c/src/rtk.rs:4287` | builder, epoch metadata/count, base/rover observations, satellite positions, wavelengths, offsets, skipped count, owned free |
| `build_dual_frequency_rinex_rtk_arc` | `__init__.pyi:3717`; `src/rtk.rs:2814` | header:19703; C Rust `bindings/c/src/rtk.rs:4342` | builder, count, metadata, sort key, observations, all/base/rover satellite positions, skipped count, owned free |
| covariance6 km↔m, diagonal, PSD interpolation | `__init__.pyi:13376-13382`; `src/covariance.rs:558-594` | header:20845-20878; C Rust `bindings/c/src/covariance.rs:293-375` | complete 6x6 matrix input/output, unit conversions, diagonal creation, PSD interpolation, validation |
| covariance6 ECI↔RTN | `__init__.pyi:13385-13388`; `src/covariance.rs:597-622` | header:20834-20901; C Rust `bindings/c/src/covariance.rs:423-489` | both transforms and validation delegate to public `Covariance6` semantics; full matrix copied |
| `parse_rinex_nav_lenient` | `__init__.pyi:13290`; `src/rinex.rs:2821` | header:26188-26200, accessors:25172-25217; C Rust `bindings/c/src/rinex.rs:867-1020` | owned result exposes count, records, skipped entries and skipped messages; no silent loss |
| `parse_sbas_ems_lines` / `parse_sbas_rtklib_lines` | `__init__.pyi:12553-12554`; `src/sbas_ssr.rs:2224-2232` | header:26210-26221; C Rust `bindings/c/src/sbas.rs:79-116` | both return owned log blocks; count/item and byte query/copy preserve bytes |
| `parse_rinex_glonass_records` | `__init__.pyi:6293`; `src/rinex.rs:2904` | header:26153-26164, accessors:27979-28017; C Rust `bindings/c/src/rinex.rs:1276-1512` | records and skipped raw tokens are separate copied outputs, including extended slots |
| `parse_rinex_iono_corrections` | `__init__.pyi:6294`; `src/rinex.rs:2914` | header:26164; C Rust `bindings/c/src/rinex.rs:1479-1527` | presence-tagged GPS/BeiDou alpha/beta values retained |
| `parse_rinex_leap_seconds` | `__init__.pyi:6295`; `src/rinex.rs:2920` | header:26175; C Rust `bindings/c/src/rinex.rs:1529-1567` | optional value and presence flag retained |
| bare SSR decode/bias/access | `__init__.pyi:12664-12723`; `src/sbas_ssr.rs:2253-2261` | header:33825-33944; C Rust `bindings/c/src/ssr.rs:403-735` | header/system/kind/epoch/update fields, orbit/clock/URA/code/phase counts, all variable record/signal accessors, free |
| `sbas_prn_to_satellite_id` | `__init__.pyi:12555`; `src/sbas_ssr.rs:2240` | header:30715; C Rust `bindings/c/src/sbas.rs:39-59` | copied token output; no-match is successful zero-length output |
| reverse `satellite_id_to_sbas_prn` | `__init__.pyi:12556`; `src/sbas_ssr.rs:2246-2248` | composed over header:30715; public core `crates/sidereon-core/src/sbas/store.rs:747-758` | finite inverse of exact forward route; Go convenience, not a direct C export |
| bare `SsrMessage.encode` | `__init__.pyi:12695`; SSR implementation | generic `sidereon_rtcm_message_encode` / `sidereon_rtcm_message_to_frame`; header:28755, 28968 | original body retained or generic message encoder used; copied frame bytes, no field loss |

No C route is “nearby only” in these checks: each listed parser has its own constructor route, each aggregate has count/item and free accessors, and each returned variable byte/value array has query/copy semantics.

## Exclusions

No current C export is excluded. Free functions map to `Close`; status/version helpers map to support. Network, TLS, retry, filesystem, cache locking, decompression, and presentation aliases are Go-owned adapters. Reverse SBAS lookup is an explicit finite composition over the direct forward ABI route. SSR bare encoding is composed through generic RTCM encoding with retained body bytes.

## Mechanical invariant

The following table is the complete current C declaration map. It has exactly one row per current header function, 1,503 unique names, and no exclusions.

| # | Header line | C function | Proposed Go bridge | Disposition |
| --- | ---: | --- | --- | --- |
| 1 | 18913 | sidereon_absolute_from_relative | support.AbsoluteFromRelative | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 2 | 18917 | sidereon_almanac_lunar_solar_eclipses | astro.AlmanacLunarSolarEclipses | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 3 | 18927 | sidereon_almanac_meridian_transits | astro.AlmanacMeridianTransits | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 4 | 18940 | sidereon_almanac_moon_phases | astro.AlmanacMoonPhases | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 5 | 18950 | sidereon_almanac_planetary_events | astro.AlmanacPlanetaryEvents | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 6 | 18962 | sidereon_almanac_seasons | astro.AlmanacSeasons | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 7 | 18977 | sidereon_alpha_beta_filter_step | errormetrics.AlphaBetaFilterStep | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 8 | 18989 | sidereon_alpha_beta_steady_state_gains | errormetrics.AlphaBetaSteadyStateGains | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 9 | 18999 | sidereon_angular_separation_coords_deg | support.AngularSeparationCoordsDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 10 | 19011 | sidereon_angular_separation_deg | support.AngularSeparationDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 11 | 19020 | sidereon_antenna_free | geodesy.Antenna.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 12 | 19031 | sidereon_antenna_pco | geodesy.AntennaPco | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 13 | 19045 | sidereon_antenna_pcv | geodesy.AntennaPcv | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 14 | 19062 | sidereon_antex_antenna | geodesy.AntexAntenna | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 15 | 19072 | sidereon_antex_antenna_count | geodesy.AntexAntennaCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 16 | 19084 | sidereon_antex_encode | geodesy.AntexEncode | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 17 | 19097 | sidereon_antex_free | geodesy.Antex.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 18 | 19107 | sidereon_antex_parse | geodesy.AntexParse | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 19 | 19118 | sidereon_araim | support.Araim | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 20 | 19128 | sidereon_araim_allocation_lpv_200 | support.AraimAllocationLpv200 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 21 | 19138 | sidereon_araim_result_fault_mode_excluded_sats | support.AraimResultFaultModeExcludedSats | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 22 | 19152 | sidereon_araim_result_fault_modes | support.AraimResultFaultModes | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 23 | 19163 | sidereon_araim_result_free | support.AraimResult.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 24 | 19172 | sidereon_araim_result_summary | support.AraimResultSummary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 25 | 19182 | sidereon_atmosphere_input_default | geodesy.AtmosphereInputDefault | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 26 | 19195 | sidereon_atmosphere_nrlmsise00 | geodesy.AtmosphereNrlmsise00 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 27 | 19198 | sidereon_beta_angle_deg | errormetrics.BetaAngleDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 28 | 19202 | sidereon_beta_angle_from_state_deg | errormetrics.BetaAngleFromStateDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 29 | 19207 | sidereon_bias_set_code_dsb_seconds | errormetrics.BiasSetCodeDsbSeconds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 30 | 19215 | sidereon_bias_set_code_osb_seconds | errormetrics.BiasSetCodeOsbSeconds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 31 | 19222 | sidereon_bias_set_free | errormetrics.BiasSet.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 32 | 19224 | sidereon_bias_set_mode | errormetrics.BiasSetMode | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 33 | 19228 | sidereon_bias_set_phase_osb_cycles | errormetrics.BiasSetPhaseOsbCycles | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 34 | 19235 | sidereon_bias_set_record | errormetrics.BiasSetRecord | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 35 | 19239 | sidereon_bias_set_record_count | errormetrics.BiasSetRecordCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 36 | 19242 | sidereon_bias_set_skipped_record_count | errormetrics.BiasSetSkippedRecordCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 37 | 19245 | sidereon_bias_set_warning_count | errormetrics.BiasSetWarningCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 38 | 19248 | sidereon_bias_sinex_load | errormetrics.BiasSinexLoad | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 39 | 19250 | sidereon_bias_sinex_load_lossy | errormetrics.BiasSinexLoadLossy | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 40 | 19252 | sidereon_bias_sinex_parse | errormetrics.BiasSinexParse | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 41 | 19256 | sidereon_bias_sinex_parse_lossy | errormetrics.BiasSinexParseLossy | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 42 | 19275 | sidereon_bounded_ils_search | support.BoundedIlsSearch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 43 | 19294 | sidereon_broadcast_comparison_compare | positioning.BroadcastComparisonCompare | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 44 | 19314 | sidereon_broadcast_comparison_compare_window | positioning.BroadcastComparisonCompareWindow | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 45 | 19327 | sidereon_broadcast_comparison_free | positioning.BroadcastComparison.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 46 | 19334 | sidereon_broadcast_comparison_overall | positioning.BroadcastComparisonOverall | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 47 | 19344 | sidereon_broadcast_comparison_satellite | positioning.BroadcastComparisonSatellite | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 48 | 19355 | sidereon_broadcast_comparison_satellite_count | positioning.BroadcastComparisonSatelliteCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 49 | 19366 | sidereon_broadcast_eccentric_anomaly | positioning.BroadcastEccentricAnomaly | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 50 | 19380 | sidereon_broadcast_emission_media_batch_at_j2000_s | positioning.BroadcastEmissionMediaBatchAtJ2000S | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 51 | 19404 | sidereon_broadcast_ephemeris_free | positioning.BroadcastEphemeris.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 52 | 19411 | sidereon_broadcast_ephemeris_glonass_frequency_channel_count | positioning.BroadcastEphemerisGLONASSFrequencyChannelCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 53 | 19420 | sidereon_broadcast_ephemeris_glonass_frequency_channels | positioning.BroadcastEphemerisGLONASSFrequencyChannels | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 54 | 19431 | sidereon_broadcast_ephemeris_glonass_record_count | positioning.BroadcastEphemerisGLONASSRecordCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 55 | 19441 | sidereon_broadcast_ephemeris_glonass_records | positioning.BroadcastEphemerisGLONASSRecords | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 56 | 19452 | sidereon_broadcast_ephemeris_iono_corrections | positioning.BroadcastEphemerisIonoCorrections | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 57 | 19462 | sidereon_broadcast_ephemeris_leap_seconds | positioning.BroadcastEphemerisLeapSeconds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 58 | 19477 | sidereon_broadcast_ephemeris_load_nav | positioning.BroadcastEphemerisLoadNav | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 59 | 19480 | sidereon_broadcast_ephemeris_nav_message_preference | positioning.BroadcastEphemerisNavMessagePreference | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 60 | 19492 | sidereon_broadcast_ephemeris_parse_nav | positioning.BroadcastEphemerisParseNav | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 61 | 19496 | sidereon_broadcast_ephemeris_record_cnav_correction | positioning.BroadcastEphemerisRecordCNAVCorrection | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 62 | 19502 | sidereon_broadcast_ephemeris_record_count | positioning.BroadcastEphemerisRecordCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 63 | 19505 | sidereon_broadcast_ephemeris_record_group_delay | positioning.BroadcastEphemerisRecordGroupDelay | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 64 | 19511 | sidereon_broadcast_ephemeris_records | positioning.BroadcastEphemerisRecords | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 65 | 19527 | sidereon_broadcast_ephemeris_records_full | positioning.BroadcastEphemerisRecordsFull | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 66 | 19540 | sidereon_broadcast_ephemeris_sample | positioning.BroadcastEphemerisSample | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 67 | 19551 | sidereon_broadcast_ephemeris_select_by_issue | positioning.BroadcastEphemerisSelectByIssue | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 68 | 19559 | sidereon_broadcast_ephemeris_set_nav_message_preference | positioning.BroadcastEphemerisSetNavMessagePreference | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 69 | 19573 | sidereon_broadcast_observable_state | positioning.BroadcastObservableState | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 70 | 19588 | sidereon_broadcast_observable_states_at_j2000_s | positioning.BroadcastObservableStatesAtJ2000S | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 71 | 19604 | sidereon_broadcast_observable_states_at_shared_j2000_s | positioning.BroadcastObservableStatesAtSharedJ2000S | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 72 | 19625 | sidereon_broadcast_observables | positioning.BroadcastObservables | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 73 | 19643 | sidereon_broadcast_observables_batch | positioning.BroadcastObservablesBatch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 74 | 19658 | sidereon_broadcast_satellite_clock_offset_s | positioning.BroadcastSatelliteClockOffsetS | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 75 | 19674 | sidereon_broadcast_satellite_position_ecef | positioning.BroadcastSatellitePositionECEF | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 76 | 19687 | sidereon_broadcast_satellite_state | positioning.BroadcastSatelliteState | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 77 | 19703 | sidereon_build_dual_frequency_rinex_rtk_arc | positioning.BuildDualFrequencyRINEXRTKArc | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 78 | 19718 | sidereon_build_rinex_rtk_arc | positioning.BuildRINEXRTKArc | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 79 | 19727 | sidereon_carrier_band_label | positioning.CarrierBandLabel | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 80 | 19739 | sidereon_carrier_code_minus_carrier | positioning.CarrierCodeMinusCarrier | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 81 | 19750 | sidereon_carrier_geometry_free | positioning.CarrierGeometry.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 82 | 19758 | sidereon_carrier_melbourne_wubbena | positioning.CarrierMelbourneWubbena | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 83 | 19772 | sidereon_carrier_narrow_lane_code | positioning.CarrierNarrowLaneCode | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 84 | 19784 | sidereon_carrier_phase_meters | positioning.CarrierPhaseMeters | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 85 | 19792 | sidereon_carrier_wide_lane_cycles | positioning.CarrierWideLaneCycles | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 86 | 19806 | sidereon_carrier_wide_lane_wavelength | positioning.CarrierWideLaneWavelength | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 87 | 19813 | sidereon_cdm_free | astro.CDM.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 88 | 19822 | sidereon_cdm_numbers | astro.CDMNumbers | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 89 | 19836 | sidereon_cdm_object_state | astro.CDMObjectState | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 90 | 19851 | sidereon_cdm_object_string_field | astro.CDMObjectStringField | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 91 | 19868 | sidereon_cdm_object_velocity_covariance | astro.CDMObjectVelocityCovariance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 92 | 19879 | sidereon_cdm_parse_kvn | astro.CDMParseKVN | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 93 | 19889 | sidereon_cdm_parse_xml | astro.CDMParseXML | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 94 | 19901 | sidereon_cdm_string_field | astro.CDMStringField | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 95 | 19915 | sidereon_cdm_to_kvn | astro.CDMToKVN | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 96 | 19928 | sidereon_cdm_to_xml | astro.CDMToXML | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 97 | 19940 | sidereon_cfar_ca_false_alarm_probability | errormetrics.CFARCAFalseAlarmProbability | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 98 | 19951 | sidereon_cfar_ca_multiplier_from_pfa | errormetrics.CFARCAMultiplierFromPfa | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 99 | 19960 | sidereon_cfar_ca_pfa_from_multiplier | errormetrics.CFARCAPfaFromMultiplier | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 100 | 19970 | sidereon_cfar_ca_threshold | errormetrics.CFARCAThreshold | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 101 | 19982 | sidereon_chan_ho_initial_guess | support.ChanHoInitialGuess | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 102 | 19996 | sidereon_chi2_inv | errormetrics.Chi2Inv | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 103 | 20005 | sidereon_civil_to_gps_seconds | support.CivilToGPSSeconds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 104 | 20021 | sidereon_civil_to_j2000_seconds | geodesy.CivilToJ2000Seconds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 105 | 20036 | sidereon_clock_allan_curve | errormetrics.ClockAllanCurve | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 106 | 20049 | sidereon_clock_allan_curve_present | errormetrics.ClockAllanCurvePresent | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 107 | 20060 | sidereon_clock_allan_deviation | errormetrics.ClockAllanDeviation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 108 | 20077 | sidereon_clock_allan_deviation_curves_free | errormetrics.ClockAllanDeviationCurves.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 109 | 20085 | sidereon_clock_allan_options_init | errormetrics.ClockAllanOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 110 | 20097 | sidereon_clock_compute_allan_deviations | errormetrics.ClockComputeAllanDeviations | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 111 | 20112 | sidereon_clock_fit_power_law_noise | errormetrics.ClockFitPowerLawNoise | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 112 | 20126 | sidereon_clock_hadamard_deviation | errormetrics.ClockHadamardDeviation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 113 | 20144 | sidereon_clock_modified_adev | errormetrics.ClockModifiedAdev | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 114 | 20162 | sidereon_clock_overlapping_adev | errormetrics.ClockOverlappingAdev | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 115 | 20178 | sidereon_clock_power_law_noise_fit_coefficients | errormetrics.ClockPowerLawNoiseFitCoefficients | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 116 | 20187 | sidereon_clock_power_law_noise_fit_free | errormetrics.ClockPowerLawNoiseFit.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 117 | 20195 | sidereon_clock_power_law_noise_fit_octaves | errormetrics.ClockPowerLawNoiseFitOctaves | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 118 | 20206 | sidereon_clock_power_law_noise_fit_regions | errormetrics.ClockPowerLawNoiseFitRegions | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 119 | 20217 | sidereon_clock_power_law_noise_options_init | errormetrics.ClockPowerLawNoiseOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 120 | 20227 | sidereon_clock_power_law_noise_slopes | errormetrics.ClockPowerLawNoiseSlopes | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 121 | 20238 | sidereon_clock_time_deviation | geodesy.ClockTimeDeviation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 122 | 20255 | sidereon_closed_form_initial_guess | support.ClosedFormInitialGuess | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 123 | 20263 | sidereon_cnav_ura_ned_m | support.CNAVUraNedM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 124 | 20269 | sidereon_cnav_ura_nominal_m | support.CNAVUraNominalM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 125 | 20273 | sidereon_code_dcb_load | support.CodeDcbLoad | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 126 | 20277 | sidereon_code_dcb_load_lossy | support.CodeDcbLoadLossy | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 127 | 20281 | sidereon_code_dcb_parse | support.CodeDcbParse | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 128 | 20286 | sidereon_code_dcb_parse_lossy | support.CodeDcbParseLossy | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 129 | 20291 | sidereon_coe2eq | support.Coe2eq | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 130 | 20295 | sidereon_coe2mee | support.Coe2mee | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 131 | 20307 | sidereon_coe2rv | support.Coe2rv | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 132 | 20319 | sidereon_collision_probability | astro.CollisionProbability | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 133 | 20331 | sidereon_combination_gamma | positioning.CombinationGamma | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 134 | 20339 | sidereon_combination_ionosphere_free | positioning.CombinationIonosphere.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 135 | 20352 | sidereon_combination_ionosphere_free_phase_cycles | positioning.CombinationIonosphereFreePhaseCycles | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 136 | 20364 | sidereon_combination_ionosphere_free_phase_m | positioning.CombinationIonosphereFreePhaseM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 137 | 20382 | sidereon_combination_ionosphere_free_pseudoranges | positioning.CombinationIonosphereFreePseudoranges | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 138 | 20396 | sidereon_combination_noise_amplification | positioning.CombinationNoiseAmplification | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 139 | 20414 | sidereon_constellation_build | astro.ConstellationBuild | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 140 | 20435 | sidereon_constellation_build_at | astro.ConstellationBuildAt | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 141 | 20450 | sidereon_constellation_diff | astro.ConstellationDiff | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 142 | 20460 | sidereon_constellation_diff_activity_changed | astro.ConstellationDiffActivityChanged | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 143 | 20472 | sidereon_constellation_diff_added | astro.ConstellationDiffAdded | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 144 | 20484 | sidereon_constellation_diff_changed | astro.ConstellationDiffChanged | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 145 | 20492 | sidereon_constellation_diff_counts | astro.ConstellationDiffCounts | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 146 | 20501 | sidereon_constellation_diff_fdma_channel_changed | astro.ConstellationDiffFdmaChannelChanged | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 147 | 20513 | sidereon_constellation_diff_free | astro.ConstellationDiff.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 148 | 20521 | sidereon_constellation_diff_norad_reassigned | astro.ConstellationDiffNoradReassigned | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 149 | 20533 | sidereon_constellation_diff_removed | astro.ConstellationDiffRemoved | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 150 | 20542 | sidereon_constellation_diff_sp3_id_changed_from | positioning.ConstellationDiffSP3IdChangedFrom | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 151 | 20554 | sidereon_constellation_diff_sp3_id_changed_meta | positioning.ConstellationDiffSP3IdChangedMeta | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 152 | 20561 | sidereon_constellation_diff_sp3_id_changed_to | positioning.ConstellationDiffSP3IdChangedTo | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 153 | 20574 | sidereon_constellation_diff_svn_changed | astro.ConstellationDiffSvnChanged | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 154 | 20586 | sidereon_constellation_diff_usability_changed | astro.ConstellationDiffUsabilityChanged | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 155 | 20599 | sidereon_constellation_free | astro.Constellation.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 156 | 20607 | sidereon_constellation_galileo_prn_for_gsat | astro.ConstellationGalileoPrnForGsat | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 157 | 20617 | sidereon_constellation_glonass_fdma_channel | positioning.ConstellationGLONASSFdmaChannel | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 158 | 20627 | sidereon_constellation_glonass_slot_for_number | positioning.ConstellationGLONASSSlotForNumber | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 159 | 20640 | sidereon_constellation_gnss_sp3_id | positioning.ConstellationGNSSSP3Id | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 160 | 20656 | sidereon_constellation_record | astro.ConstellationRecord | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 161 | 20666 | sidereon_constellation_record_count | astro.ConstellationRecordCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 162 | 20679 | sidereon_constellation_to_csv | astro.ConstellationToCsv | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 163 | 20695 | sidereon_constellation_validate | astro.ConstellationValidate | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 164 | 20705 | sidereon_constellation_validate_against_sp3 | positioning.ConstellationValidateAgainstSP3 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 165 | 20720 | sidereon_constellation_validate_against_sp3_ids | positioning.ConstellationValidateAgainstSP3Ids | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 166 | 20732 | sidereon_constellation_validate_against_sp3_ids_strict | positioning.ConstellationValidateAgainstSP3IdsStrict | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 167 | 20744 | sidereon_constellation_validation_duplicate_norad_ids | astro.ConstellationValidationDuplicateNoradIds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 168 | 20759 | sidereon_constellation_validation_duplicate_prns | astro.ConstellationValidationDuplicatePrns | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 169 | 20774 | sidereon_constellation_validation_extra_sp3_ids | positioning.ConstellationValidationExtraSP3Ids | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 170 | 20787 | sidereon_constellation_validation_free | astro.ConstellationValidation.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 171 | 20798 | sidereon_constellation_validation_inactive_unusable_prns | astro.ConstellationValidationInactiveUnusablePrns | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 172 | 20809 | sidereon_constellation_validation_is_valid | astro.ConstellationValidationIsValid | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 173 | 20821 | sidereon_constellation_validation_missing_sp3_ids | positioning.ConstellationValidationMissingSP3Ids | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 174 | 20834 | sidereon_covariance6_eci_to_rtn | errormetrics.Covariance6ECIToRTN | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 175 | 20845 | sidereon_covariance6_from_diagonal | errormetrics.Covariance6FromDiagonal | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 176 | 20856 | sidereon_covariance6_interpolate_psd | errormetrics.Covariance6InterpolatePSD | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 177 | 20868 | sidereon_covariance6_km_to_m | errormetrics.Covariance6KmToM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 178 | 20878 | sidereon_covariance6_m_to_km | errormetrics.Covariance6MToKm | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 179 | 20888 | sidereon_covariance6_rtn_to_eci | errormetrics.Covariance6RTNToECI | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 180 | 20901 | sidereon_covariance6_validate | errormetrics.Covariance6Validate | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 181 | 20904 | sidereon_covariance_ephemeris_count | astro.CovarianceEphemerisCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 182 | 20907 | sidereon_covariance_ephemeris_covariance_at | astro.CovarianceEphemerisCovarianceAt | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 183 | 20911 | sidereon_covariance_ephemeris_free | astro.CovarianceEphemeris.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 184 | 20913 | sidereon_covariance_ephemeris_nodes | astro.CovarianceEphemerisNodes | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 185 | 20933 | sidereon_covariance_from_jacobian | errormetrics.CovarianceFromJacobian | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 186 | 20948 | sidereon_covariance_is_positive_semidefinite | errormetrics.CovarianceIsPositiveSemidefinite | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 187 | 20957 | sidereon_covariance_is_symmetric | errormetrics.CovarianceIsSymmetric | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 188 | 20959 | sidereon_covariance_transport | errormetrics.CovarianceTransport | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 189 | 20976 | sidereon_coverage_grid_access_counts | support.CoverageGridAccessCounts | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 190 | 20989 | sidereon_coverage_grid_dimensions | support.CoverageGridDimensions | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 191 | 20999 | sidereon_coverage_grid_free | support.CoverageGrid.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 192 | 21007 | sidereon_coverage_grid_look_angle | support.CoverageGridLookAngle | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 193 | 21021 | sidereon_coverage_grid_max_elevation_deg | support.CoverageGridMaxElevationDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 194 | 21034 | sidereon_coverage_grid_visible_mask | astro.CoverageGridVisibleMask | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 195 | 21049 | sidereon_coverage_look_angles | support.CoverageLookAngles | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 196 | 21067 | sidereon_crinex_decode | support.CrinexDecode | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 197 | 21084 | sidereon_crinex_encode | support.CrinexEncode | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 198 | 21091 | sidereon_cw_propagate | astro.CwPropagate | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 199 | 21096 | sidereon_cw_stm | astro.CwStm | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 200 | 21107 | sidereon_cycle_slip_options_init | support.CycleSlipOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 201 | 21116 | sidereon_data_day_of_year | distribution.DataDayOfYear | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 202 | 21138 | sidereon_data_default_sample_for_date | distribution.DataDefaultSampleForDate | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 203 | 21159 | sidereon_data_distribution_location | distribution.DataDistributionLocation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 204 | 21186 | sidereon_data_newest_published_product_json | distribution.DataNewestPublishedProductJSON | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 205 | 21210 | sidereon_data_next_issue_due_json | distribution.DataNextIssueDueJSON | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 206 | 21242 | sidereon_data_predicted_ionex_line_candidates_json | positioning.DataPredictedIonexLineCandidatesJSON | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 207 | 21260 | sidereon_data_problem_init | distribution.DataProblemInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 208 | 21276 | sidereon_data_product_identity | distribution.DataProductIdentity | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 209 | 21294 | sidereon_data_product_identity_cache_key | distribution.DataProductIdentityCacheKey | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 210 | 21300 | sidereon_data_product_solution_class | distribution.DataProductSolutionClass | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 211 | 21327 | sidereon_data_publication_listing_urls_json | distribution.DataPublicationListingUrlsJSON | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 212 | 21351 | sidereon_data_sp3_content_start_convention | positioning.DataSP3ContentStartConvention | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 213 | 21379 | sidereon_data_supported_samples | distribution.DataSupportedSamples | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 214 | 21402 | sidereon_data_validate_exact_product_set | distribution.DataValidateExactProductSet | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 215 | 21415 | sidereon_day_of_year | geodesy.DayOfYear | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 216 | 21428 | sidereon_decay_config_init | astro.DecayConfigInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 217 | 21438 | sidereon_default_iono_free_pair | positioning.DefaultIonoFreePair | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 218 | 21448 | sidereon_default_spp_frequency_hz | positioning.DefaultSPPFrequencyHz | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 219 | 21459 | sidereon_detect_cycle_slips | support.DetectCycleSlips | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 220 | 21474 | sidereon_dgnss_applied_corrected | positioning.DGNSSAppliedCorrected | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 221 | 21486 | sidereon_dgnss_applied_counts | positioning.DGNSSAppliedCounts | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 222 | 21495 | sidereon_dgnss_applied_dropped | positioning.DGNSSAppliedDropped | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 223 | 21505 | sidereon_dgnss_applied_free | positioning.DGNSSApplied.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 224 | 21515 | sidereon_dgnss_apply_corrections | positioning.DGNSSApplyCorrections | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 225 | 21527 | sidereon_dgnss_correction | positioning.DGNSSCorrection | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 226 | 21537 | sidereon_dgnss_corrections_count | positioning.DGNSSCorrectionsCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 227 | 21546 | sidereon_dgnss_corrections_free | positioning.DGNSSCorrections.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 228 | 21561 | sidereon_dgnss_position_solve | positioning.DGNSSPositionSolve | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 229 | 21580 | sidereon_dgnss_pseudorange_corrections | positioning.DGNSSPseudorangeCorrections | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 230 | 21594 | sidereon_dgnss_solution_baseline | positioning.DGNSSSolutionBaseline | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 231 | 21607 | sidereon_dgnss_solution_dropped_sats | positioning.DGNSSSolutionDroppedSats | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 232 | 21618 | sidereon_dgnss_solution_free | positioning.DGNSSSolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 233 | 21627 | sidereon_dgnss_solution_solution | positioning.DGNSSSolutionSolution | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 234 | 21643 | sidereon_dop | errormetrics.DOP | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 235 | 21657 | sidereon_dop_with_convention | errormetrics.DOPWithConvention | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 236 | 21671 | sidereon_doppler_range_rate_and_ratio | positioning.DopplerRangeRateAndRatio | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 237 | 21686 | sidereon_doppler_shift | positioning.DopplerShift | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 238 | 21701 | sidereon_drag_force_acceleration | astro.DragForceAcceleration | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 239 | 21710 | sidereon_drag_parameters_from_area_mass | astro.DragParametersFromAreaMass | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 240 | 21722 | sidereon_drag_parameters_from_ballistic_coefficient | astro.DragParametersFromBallisticCoefficient | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 241 | 21732 | sidereon_drag_parameters_from_bc_factor | astro.DragParametersFromBcFactor | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 242 | 21743 | sidereon_dted_interpolation_label | geodesy.DTEDInterpolationLabel | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 243 | 21755 | sidereon_dted_lookup_options_init | geodesy.DTEDLookupOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 244 | 21762 | sidereon_dted_terrain_free | geodesy.DTEDTerrain.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 245 | 21775 | sidereon_dted_terrain_height_batch_m | geodesy.DTEDTerrainHeightBatchM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 246 | 21787 | sidereon_dted_terrain_height_m | geodesy.DTEDTerrainHeightM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 247 | 21799 | sidereon_dted_terrain_height_m_with_options | geodesy.DTEDTerrainHeightMWithOptions | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 248 | 21812 | sidereon_dted_terrain_new | geodesy.DTEDTerrainNew | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 249 | 21820 | sidereon_dted_tile_free | geodesy.DTEDTile.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 250 | 21829 | sidereon_dted_tile_get_elevation | geodesy.DTEDTileGetElevation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 251 | 21844 | sidereon_dted_tile_list_to_mmap_store | geodesy.DTEDTileListToMMapStore | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 252 | 21857 | sidereon_dted_tile_load | geodesy.DTEDTileLoad | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 253 | 21867 | sidereon_dted_tree_to_mmap_store | distribution.DTEDTreeToMmapStore | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 254 | 21879 | sidereon_earth_angular_radius_deg | geodesy.EarthAngularRadiusDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 255 | 21881 | sidereon_eccentric_to_mean_anomaly | astro.EccentricToMeanAnomaly | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 256 | 21885 | sidereon_eccentric_to_true_anomaly | support.EccentricToTrueAnomaly | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 257 | 21898 | sidereon_ecef_to_geodetic | geodesy.ECEFToGeodetic | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 258 | 21907 | sidereon_eclipse_shadow_fraction | astro.EclipseShadowFraction | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 259 | 21917 | sidereon_eclipse_shadow_fraction_with_model | astro.EclipseShadowFractionWithModel | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 260 | 21929 | sidereon_eclipse_status | astro.EclipseStatus | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 261 | 21939 | sidereon_eclipse_status_with_model | astro.EclipseStatusWithModel | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 262 | 21949 | sidereon_egm96_15m_geoid_free | geodesy.EGM9615mGeoid.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 263 | 21958 | sidereon_egm96_15m_geoid_from_ww15mgh_dac_bytes | geodesy.EGM9615mGeoidFromWw15mghDacBytes | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 264 | 21971 | sidereon_egm96_15m_geoid_from_ww15mgh_dac_path | geodesy.EGM9615mGeoidFromWw15mghDacPath | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 265 | 21982 | sidereon_egm96_ellipsoidal_height_m | geodesy.EGM96EllipsoidalHeightM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 266 | 21995 | sidereon_egm96_orthometric_height_m | geodesy.EGM96OrthometricHeightM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 267 | 22008 | sidereon_egm96_undulation | geodesy.EGM96Undulation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 268 | 22010 | sidereon_egm96_undulations_deg | geodesy.EGM96UndulationsDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 269 | 22017 | sidereon_egm96_undulations_rad | geodesy.EGM96UndulationsRad | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 270 | 22031 | sidereon_ellipsoidal_height_m | geodesy.EllipsoidalHeightM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 271 | 22041 | sidereon_emission_media_options_init | support.EmissionMediaOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 272 | 22052 | sidereon_encode_rinex_nav | positioning.EncodeRINEXNAV | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 273 | 22065 | sidereon_encounter_frame | astro.EncounterFrame | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 274 | 22080 | sidereon_encounter_plane_covariance | astro.EncounterPlaneCovariance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 275 | 22089 | sidereon_ephemeris_epoch_count | astro.EphemerisEpochCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 276 | 22101 | sidereon_ephemeris_free | astro.Ephemeris.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 277 | 22111 | sidereon_ephemeris_states | astro.EphemerisStates | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 278 | 22125 | sidereon_ephemeris_times_s | astro.EphemerisTimesS | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 279 | 22131 | sidereon_eq2coe | support.Eq2coe | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 280 | 22134 | sidereon_eq2rv | support.Eq2rv | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 281 | 22149 | sidereon_error_ellipse_2x2 | errormetrics.ErrorEllipse2x2 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 282 | 22159 | sidereon_error_metrics_error_ellipse_from_enu_m2 | errormetrics.ErrorMetricsErrorEllipseFromEnuM2 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 283 | 22169 | sidereon_error_metrics_from_ecef_covariance_m2 | errormetrics.ErrorMetricsFromECEFCovarianceM2 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 284 | 22180 | sidereon_error_metrics_from_enu_covariance_m2 | errormetrics.ErrorMetricsFromEnuCovarianceM2 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 285 | 22189 | sidereon_error_metrics_from_kinematic_solution | errormetrics.ErrorMetricsFromKinematicSolution | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 286 | 22198 | sidereon_error_metrics_from_position_covariance | errormetrics.ErrorMetricsFromPositionCovariance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 287 | 22208 | sidereon_error_metrics_horizontal_radius_at | errormetrics.ErrorMetricsHorizontalRadiusAt | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 288 | 22219 | sidereon_error_metrics_spherical_radius_at | errormetrics.ErrorMetricsSphericalRadiusAt | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 289 | 22229 | sidereon_error_metrics_vertical_radius_at | geodesy.ErrorMetricsVerticalRadiusAt | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 290 | 22240 | sidereon_estimate_decay | astro.EstimateDecay | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 291 | 22244 | sidereon_estimate_decay_with_space_weather_table | astro.EstimateDecayWithSpaceWeatherTable | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 292 | 22254 | sidereon_ewma_update | errormetrics.EWMAUpdate | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 293 | 22261 | sidereon_ewma_update_power_of_two | errormetrics.EWMAUpdatePowerOfTwo | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 294 | 22271 | sidereon_exact_cache_cleanup | distribution.ExactCacheCleanup | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 295 | 22284 | sidereon_exact_cache_entry_copy_bytes | distribution.ExactCacheEntryCopyBytes | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 296 | 22300 | sidereon_exact_cache_entry_copy_id | distribution.ExactCacheEntryCopyId | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 297 | 22314 | sidereon_exact_cache_entry_copy_path | distribution.ExactCacheEntryCopyPath | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 298 | 22324 | sidereon_exact_cache_entry_free | distribution.ExactCacheEntry.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 299 | 22329 | sidereon_exact_cache_free | distribution.ExactCache.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 300 | 22343 | sidereon_exact_cache_open | distribution.ExactCacheOpen | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 301 | 22366 | sidereon_exact_cache_open_single_flight | distribution.ExactCacheOpenSingleFlight | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 302 | 22380 | sidereon_exact_cache_owner_free | distribution.ExactCacheOwner.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 303 | 22387 | sidereon_exact_cache_owner_heartbeat | distribution.ExactCacheOwnerHeartbeat | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 304 | 22402 | sidereon_exact_cache_owner_publish | distribution.ExactCacheOwnerPublish | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 305 | 22421 | sidereon_exact_cache_publish | distribution.ExactCachePublish | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 306 | 22439 | sidereon_exact_cache_read | distribution.ExactCacheRead | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 307 | 22457 | sidereon_exact_cache_read_unlocked | distribution.ExactCacheReadUnlocked | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 308 | 22473 | sidereon_exact_cache_single_flight_options_init | distribution.ExactCacheSingleFlightOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 309 | 22482 | sidereon_fde_options_init | errormetrics.FDEOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 310 | 22492 | sidereon_fde_solution_excluded_sats | errormetrics.FDESolutionExcludedSats | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 311 | 22505 | sidereon_fde_solution_free | errormetrics.FDESolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 312 | 22512 | sidereon_fde_solution_iterations | errormetrics.FDESolutionIterations | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 313 | 22523 | sidereon_fde_solution_solution | errormetrics.FDESolutionSolution | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 314 | 22538 | sidereon_fde_solve_broadcast | positioning.FDESolveBroadcast | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 315 | 22557 | sidereon_fde_solve_spp | positioning.FDESolveSPP | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 316 | 22573 | sidereon_find_moon_elevation_crossings | astro.FindMoonElevationCrossings | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 317 | 22592 | sidereon_find_moon_transits | astro.FindMoonTransits | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 318 | 22612 | sidereon_find_tca_candidates_from_tles | astro.FindTCACandidatesFromTles | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 319 | 22636 | sidereon_find_tca_conjunctions_from_tles | astro.FindTCAConjunctionsFromTles | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 320 | 22663 | sidereon_find_tca_conjunctions_with_propagated_covariance_from_tles | astro.FindTCAConjunctionsWithPropagatedCovarianceFromTles | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 321 | 22684 | sidereon_fit_all_sp3_ecef_precise_orbits | positioning.FitAllSP3ECEFPreciseOrbits | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 322 | 22696 | sidereon_fit_precise_ephemeris_sample_orbit | positioning.FitPreciseEphemerisSampleOrbit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 323 | 22712 | sidereon_fit_sp3_ecef_precise_orbit | positioning.FitSP3ECEFPreciseOrbit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 324 | 22723 | sidereon_fit_sp3_ecef_precise_orbits | positioning.FitSP3ECEFPreciseOrbits | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 325 | 22736 | sidereon_fit_sp3_precise_orbit | positioning.FitSP3PreciseOrbit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 326 | 22751 | sidereon_fix_wide_lane_rtk_arc | positioning.FixWideLaneRTKArc | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 327 | 22763 | sidereon_force_j2_acceleration | astro.ForceJ2Acceleration | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 328 | 22774 | sidereon_force_twobody_acceleration | astro.ForceTwobodyAcceleration | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 329 | 22784 | sidereon_frame_catalog_count | astro.FrameCatalogCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 330 | 22792 | sidereon_frame_catalog_entries | astro.FrameCatalogEntries | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 331 | 22803 | sidereon_frame_catalog_entry | astro.FrameCatalogEntry | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 332 | 22813 | sidereon_frame_catalog_propagate_position | astro.FrameCatalogPropagatePosition | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 333 | 22824 | sidereon_frame_catalog_transform | astro.FrameCatalogTransform | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 334 | 22836 | sidereon_frame_catalog_transform_from_epoch | astro.FrameCatalogTransformFromEpoch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 335 | 22851 | sidereon_frame_gast_radians | astro.FrameGastRadians | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 336 | 22860 | sidereon_frame_gcrs_to_itrs | astro.FrameGCRSToITRS | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 337 | 22871 | sidereon_frame_gcrs_to_itrs_matrix | astro.FrameGCRSToITRSMatrix | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 338 | 22881 | sidereon_frame_gcrs_to_itrs_matrix_with_polar_motion | astro.FrameGCRSToITRSMatrixWithPolarMotion | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 339 | 22894 | sidereon_frame_gcrs_to_itrs_with_polar_motion | astro.FrameGCRSToITRSWithPolarMotion | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 340 | 22909 | sidereon_frame_gcrs_to_topocentric | astro.FrameGCRSToTopocentric | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 341 | 22924 | sidereon_frame_geodetic_from_ecef_proj | astro.FrameGeodeticFromECEFProj | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 342 | 22933 | sidereon_frame_geodetic_to_itrs | astro.FrameGeodeticToITRS | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 343 | 22945 | sidereon_frame_gmst_radians | astro.FrameGmstRadians | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 344 | 22954 | sidereon_frame_itrs_to_gcrs | astro.FrameITRSToGCRS | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 345 | 22964 | sidereon_frame_itrs_to_gcrs_matrix | astro.FrameITRSToGCRSMatrix | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 346 | 22974 | sidereon_frame_itrs_to_gcrs_matrix_with_polar_motion | astro.FrameITRSToGCRSMatrixWithPolarMotion | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 347 | 22987 | sidereon_frame_itrs_to_gcrs_with_polar_motion | astro.FrameITRSToGCRSWithPolarMotion | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 348 | 23000 | sidereon_frame_itrs_to_geodetic | astro.FrameITRSToGeodetic | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 349 | 23010 | sidereon_frame_mat3_vec3_mul | astro.FrameMat3Vec3Mul | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 350 | 23019 | sidereon_frame_mean_of_date_to_itrs_matrix | astro.FrameMeanOfDateToITRSMatrix | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 351 | 23029 | sidereon_frame_mean_of_date_to_itrs_matrix_with_polar_motion | astro.FrameMeanOfDateToITRSMatrixWithPolarMotion | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 352 | 23040 | sidereon_frame_polar_motion_matrix | astro.FramePolarMotionMatrix | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 353 | 23053 | sidereon_frame_teme_to_gcrs | astro.FrameTEMEToGCRS | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 354 | 23068 | sidereon_frequency_hz | support.FrequencyHz | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 355 | 23075 | sidereon_fusion_error_state_layout_dimension | errormetrics.FusionErrorStateLayoutDimension | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 356 | 23083 | sidereon_fusion_error_state_layout_label | errormetrics.FusionErrorStateLayoutLabel | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 357 | 23094 | sidereon_fusion_filter_config_init | errormetrics.FusionFilterConfigInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 358 | 23101 | sidereon_fusion_filter_configure_time_sync | geodesy.FusionFilterConfigureTimeSync | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 359 | 23111 | sidereon_fusion_filter_covariance | errormetrics.FusionFilterCovariance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 360 | 23124 | sidereon_fusion_filter_create | errormetrics.FusionFilterCreate | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 361 | 23137 | sidereon_fusion_filter_encode_state | errormetrics.FusionFilterEncodeState | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 362 | 23150 | sidereon_fusion_filter_free | errormetrics.FusionFilter.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 363 | 23158 | sidereon_fusion_filter_kind_label | errormetrics.FusionFilterKindLabel | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 364 | 23170 | sidereon_fusion_filter_propagate | astro.FusionFilterPropagate | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 365 | 23179 | sidereon_fusion_filter_propagate_recorded | astro.FusionFilterPropagateRecorded | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 366 | 23190 | sidereon_fusion_filter_restore_state | errormetrics.FusionFilterRestoreState | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 367 | 23200 | sidereon_fusion_filter_state | errormetrics.FusionFilterState | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 368 | 23209 | sidereon_fusion_filter_time_sync_status | geodesy.FusionFilterTimeSyncStatus | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 369 | 23219 | sidereon_fusion_filter_update_loose | errormetrics.FusionFilterUpdateLoose | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 370 | 23230 | sidereon_fusion_filter_update_loose_recorded | errormetrics.FusionFilterUpdateLooseRecorded | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 371 | 23243 | sidereon_fusion_filter_update_loose_time_sync | geodesy.FusionFilterUpdateLooseTimeSync | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 372 | 23254 | sidereon_fusion_filter_update_non_holonomic | errormetrics.FusionFilterUpdateNonHolonomic | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 373 | 23266 | sidereon_fusion_filter_update_non_holonomic_recorded | errormetrics.FusionFilterUpdateNonHolonomicRecorded | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 374 | 23278 | sidereon_fusion_filter_update_stationary | errormetrics.FusionFilterUpdateStationary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 375 | 23289 | sidereon_fusion_filter_update_stationary_recorded | errormetrics.FusionFilterUpdateStationaryRecorded | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 376 | 23301 | sidereon_fusion_filter_update_tight_broadcast | positioning.FusionFilterUpdateTightBroadcast | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 377 | 23313 | sidereon_fusion_filter_update_tight_broadcast_recorded | positioning.FusionFilterUpdateTightBroadcastRecorded | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 378 | 23327 | sidereon_fusion_filter_update_tight_broadcast_time_sync | positioning.FusionFilterUpdateTightBroadcastTimeSync | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 379 | 23338 | sidereon_fusion_filter_update_tight_sp3 | positioning.FusionFilterUpdateTightSP3 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 380 | 23350 | sidereon_fusion_filter_update_tight_sp3_recorded | positioning.FusionFilterUpdateTightSP3Recorded | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 381 | 23363 | sidereon_fusion_filter_update_tight_sp3_time_sync | positioning.FusionFilterUpdateTightSP3TimeSync | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 382 | 23374 | sidereon_fusion_gnss_fix_status_label | positioning.FusionGNSSFixStatusLabel | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 383 | 23385 | sidereon_fusion_imu_spec_preset | errormetrics.FusionImuSpecPreset | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 384 | 23394 | sidereon_fusion_rts_history_builder_finish | errormetrics.FusionRtsHistoryBuilderFinish | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 385 | 23403 | sidereon_fusion_rts_history_builder_free | errormetrics.FusionRtsHistoryBuilder.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 386 | 23411 | sidereon_fusion_rts_history_builder_from_filter | errormetrics.FusionRtsHistoryBuilderFromFilter | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 387 | 23420 | sidereon_fusion_rts_history_builder_new | errormetrics.FusionRtsHistoryBuilderNew | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 388 | 23428 | sidereon_fusion_rts_history_epoch | errormetrics.FusionRtsHistoryEpoch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 389 | 23437 | sidereon_fusion_rts_history_epoch_count | errormetrics.FusionRtsHistoryEpochCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 390 | 23446 | sidereon_fusion_rts_history_epoch_predicted_position_ecef_m | errormetrics.FusionRtsHistoryEpochPredictedPositionECEFM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 391 | 23459 | sidereon_fusion_rts_history_epoch_transition_from_previous | errormetrics.FusionRtsHistoryEpochTransitionFromPrevious | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 392 | 23472 | sidereon_fusion_rts_history_epoch_updated_position_ecef_m | errormetrics.FusionRtsHistoryEpochUpdatedPositionECEFM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 393 | 23485 | sidereon_fusion_rts_history_free | errormetrics.FusionRtsHistory.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 394 | 23497 | sidereon_fusion_velocity_match_outage | errormetrics.FusionVelocityMatchOutage | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 395 | 23519 | sidereon_galileo_nequick_g_native | geodesy.GalileoNequickGNative | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 396 | 23536 | sidereon_geodesic_direct | geodesy.GeodesicDirect | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 397 | 23548 | sidereon_geodesic_inverse | geodesy.GeodesicInverse | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 398 | 23561 | sidereon_geodetic_detect_steps | geodesy.GeodeticDetectSteps | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 399 | 23575 | sidereon_geodetic_fit_trajectory | geodesy.GeodeticFitTrajectory | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 400 | 23585 | sidereon_geodetic_midas_options_init | geodesy.GeodeticMidasOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 401 | 23592 | sidereon_geodetic_motion_field_common_mode | geodesy.GeodeticMotionFieldCommonMode | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 402 | 23601 | sidereon_geodetic_motion_field_free | geodesy.GeodeticMotionField.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 403 | 23609 | sidereon_geodetic_motion_field_stations | geodesy.GeodeticMotionFieldStations | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 404 | 23621 | sidereon_geodetic_network_field | geodesy.GeodeticNetworkField | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 405 | 23631 | sidereon_geodetic_step_detection_options_init | geodesy.GeodeticStepDetectionOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 406 | 23642 | sidereon_geodetic_to_ecef | geodesy.GeodeticToECEF | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 407 | 23650 | sidereon_geodetic_trajectory_components | geodesy.GeodeticTrajectoryComponents | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 408 | 23658 | sidereon_geodetic_trajectory_fit_options_init | geodesy.GeodeticTrajectoryFitOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 409 | 23666 | sidereon_geodetic_trajectory_free | geodesy.GeodeticTrajectory.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 410 | 23674 | sidereon_geodetic_trajectory_offsets | geodesy.GeodeticTrajectoryOffsets | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 411 | 23686 | sidereon_geodetic_trajectory_parameter_covariance | geodesy.GeodeticTrajectoryParameterCovariance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 412 | 23698 | sidereon_geodetic_trajectory_summary | geodesy.GeodeticTrajectorySummary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 413 | 23706 | sidereon_geodetic_trajectory_terms | geodesy.GeodeticTrajectoryTerms | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 414 | 23718 | sidereon_geodetic_velocity_midas | geodesy.GeodeticVelocityMidas | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 415 | 23728 | sidereon_geofence_containment_probability | geodesy.GeofenceContainmentProbability | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 416 | 23741 | sidereon_geofence_containment_probability_with_options | geodesy.GeofenceContainmentProbabilityWithOptions | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 417 | 23754 | sidereon_geofence_contains | geodesy.GeofenceContains | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 418 | 23766 | sidereon_geofence_create | geodesy.GeofenceCreate | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 419 | 23779 | sidereon_geofence_crossing_probability | geodesy.GeofenceCrossingProbability | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 420 | 23798 | sidereon_geofence_crossing_probability_with_options | geodesy.GeofenceCrossingProbabilityWithOptions | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 421 | 23816 | sidereon_geofence_distance_to_boundary | geodesy.GeofenceDistanceToBoundary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 422 | 23827 | sidereon_geofence_free | geodesy.Geofence.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 423 | 23834 | sidereon_geofence_hysteresis_init | geodesy.GeofenceHysteresisInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 424 | 23841 | sidereon_geofence_probability_options_init | geodesy.GeofenceProbabilityOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 425 | 23843 | sidereon_geoid_grid_ellipsoidal_height_rad | geodesy.GeoidGridEllipsoidalHeightRad | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 426 | 23854 | sidereon_geoid_grid_free | geodesy.GeoidGrid.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 427 | 23862 | sidereon_geoid_grid_from_egm2008_raster | geodesy.GeoidGridFromEgm2008Raster | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 428 | 23873 | sidereon_geoid_grid_from_egm2008_raster_window | geodesy.GeoidGridFromEgm2008RasterWindow | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 429 | 23878 | sidereon_geoid_grid_from_egm96_dac | geodesy.GeoidGridFromEGM96Dac | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 430 | 23890 | sidereon_geoid_grid_from_proj_egm96_gtx | geodesy.GeoidGridFromProjEGM96Gtx | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 431 | 23902 | sidereon_geoid_grid_from_text | geodesy.GeoidGridFromText | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 432 | 23915 | sidereon_geoid_grid_new | geodesy.GeoidGridNew | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 433 | 23925 | sidereon_geoid_grid_orthometric_height_rad | geodesy.GeoidGridOrthometricHeightRad | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 434 | 23937 | sidereon_geoid_grid_undulation_deg | geodesy.GeoidGridUndulationDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 435 | 23951 | sidereon_geoid_grid_undulation_proj_rad | geodesy.GeoidGridUndulationProjRad | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 436 | 23964 | sidereon_geoid_grid_undulation_rad | geodesy.GeoidGridUndulationRad | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 437 | 23969 | sidereon_geoid_grid_undulations_deg | geodesy.GeoidGridUndulationsDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 438 | 23977 | sidereon_geoid_grid_undulations_rad | geodesy.GeoidGridUndulationsRad | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 439 | 23992 | sidereon_geoid_undulation | geodesy.GeoidUndulation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 440 | 23996 | sidereon_geoid_undulations_deg | geodesy.GeoidUndulationsDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 441 | 24003 | sidereon_geoid_undulations_rad | geodesy.GeoidUndulationsRad | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 442 | 24016 | sidereon_glonass_g1_frequency_hz | positioning.GLONASSG1FrequencyHz | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 443 | 24023 | sidereon_gnss_seconds_of_week_from_calendar | positioning.GNSSSecondsOfWeekFromCalendar | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 444 | 24034 | sidereon_gnss_system_label | positioning.GNSSSystemLabel | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 445 | 24046 | sidereon_gnss_week_and_seconds_of_week | positioning.GNSSWeekAndSecondsOfWeek | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 446 | 24055 | sidereon_gnss_week_epoch_julian_day_number | positioning.GNSSWeekEpochJulianDayNumber | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 447 | 24064 | sidereon_gnss_week_from_calendar | positioning.GNSSWeekFromCalendar | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 448 | 24076 | sidereon_gnss_week_tow_new | positioning.GNSSWeekTowNew | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 449 | 24086 | sidereon_gnss_week_tow_normalized | positioning.GNSSWeekTowNormalized | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 450 | 24094 | sidereon_gnss_week_tow_unrolled_week | positioning.GNSSWeekTowUnrolledWeek | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 451 | 24106 | sidereon_gps_utc_offset_s | support.GPSUtcOffsetS | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 452 | 24113 | sidereon_ground_track_count | astro.GroundTrackCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 453 | 24124 | sidereon_ground_track_free | astro.GroundTrack.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 454 | 24134 | sidereon_ground_track_values | astro.GroundTrackValues | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 455 | 24148 | sidereon_hessian_trace | errormetrics.HessianTrace | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 456 | 24157 | sidereon_instant_from_utc_civil | support.InstantFromUtcCivil | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 457 | 24178 | sidereon_iod_gauss_angles | astro.IODGaussAngles | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 458 | 24194 | sidereon_iod_gibbs | astro.IODGibbs | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 459 | 24209 | sidereon_iod_hgibbs | astro.IODHgibbs | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 460 | 24226 | sidereon_ionex_epoch_count | positioning.IonexEpochCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 461 | 24236 | sidereon_ionex_exponent | positioning.IonexExponent | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 462 | 24246 | sidereon_ionex_free | positioning.Ionex.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 463 | 24255 | sidereon_ionex_from_tec_grid_samples | positioning.IonexFromTecGridSamples | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 464 | 24266 | sidereon_ionex_from_tec_samples | positioning.IonexFromTecSamples | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 465 | 24281 | sidereon_ionex_lat_nodes_deg | positioning.IonexLatNodesDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 466 | 24295 | sidereon_ionex_lon_nodes_deg | positioning.IonexLonNodesDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 467 | 24309 | sidereon_ionex_map_epochs_j2000_s | positioning.IonexMapEpochsJ2000S | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 468 | 24322 | sidereon_ionex_parse | positioning.IonexParse | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 469 | 24340 | sidereon_ionex_slant_delay | positioning.IonexSlantDelay | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 470 | 24358 | sidereon_ionex_slant_delay_with_policy | positioning.IonexSlantDelayWithPolicy | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 471 | 24375 | sidereon_ionex_tec_grid_samples_epochs_j2000_s | positioning.IonexTecGridSamplesEpochsJ2000S | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 472 | 24388 | sidereon_ionex_tec_grid_samples_info | positioning.IonexTecGridSamplesInfo | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 473 | 24398 | sidereon_ionex_tec_grid_samples_rms_maps_tecu | positioning.IonexTecGridSamplesRmsMapsTecu | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 474 | 24411 | sidereon_ionex_tec_grid_samples_tec_maps_tecu | positioning.IonexTecGridSamplesTecMapsTecu | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 475 | 24424 | sidereon_ionex_tec_samples | positioning.IonexTecSamples | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 476 | 24440 | sidereon_ionex_to_ionex_text | positioning.IonexToIonexText | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 477 | 24453 | sidereon_iono_free_pseudoranges_combined | support.IonoFreePseudorangesCombined | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 478 | 24465 | sidereon_iono_free_pseudoranges_dropped | support.IonoFreePseudorangesDropped | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 479 | 24477 | sidereon_iono_free_pseudoranges_free | support.IonoFreePseudoranges.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 480 | 24485 | sidereon_j2000_seconds_to_civil | geodesy.J2000SecondsToCivil | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 481 | 24498 | sidereon_kalman_cv_steady_state_gains | errormetrics.KalmanCvSteadyStateGains | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 482 | 24516 | sidereon_klobuchar_native | support.KlobucharNative | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 483 | 24542 | sidereon_lambda_ils_search | support.LambdaIlsSearch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 484 | 24558 | sidereon_lambert_battin | support.LambertBattin | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 485 | 24578 | sidereon_last_error_message | core.ErrorDetail (internal) | internal same-thread error detail capture; text retained in StatusError |
| 486 | 24586 | sidereon_last_terrain_datum_error | geodesy.LastTerrainDatumError | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 487 | 24594 | sidereon_last_terrain_store_error | geodesy.LastTerrainStoreError | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 488 | 24601 | sidereon_leap_second_table_info | geodesy.LeapSecondTableInfo | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 489 | 24609 | sidereon_leap_second_table_source | geodesy.LeapSecondTableSource | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 490 | 24620 | sidereon_leap_seconds | geodesy.LeapSeconds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 491 | 24632 | sidereon_line_of_sight_from_az_el_deg | support.LineOfSightFromAzElDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 492 | 24646 | sidereon_lnav_decode | astro.LNAVDecode | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 493 | 24665 | sidereon_lnav_encode | astro.LNAVEncode | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 494 | 24681 | sidereon_lnav_parity | positioning.LNAVParity | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 495 | 24697 | sidereon_lnav_parity_valid | positioning.LNAVParityValid | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 496 | 24711 | sidereon_lnav_subframe_id | positioning.LNAVSubframeID | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 497 | 24723 | sidereon_lnav_tow | positioning.LNAVTOW | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 498 | 24737 | sidereon_locate_source | support.LocateSource | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 499 | 24755 | sidereon_locate_source_with | support.LocateSourceWith | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 500 | 24768 | sidereon_look_angles_epoch_count | support.LookAnglesEpochCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 501 | 24780 | sidereon_look_angles_free | support.LookAngles.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 502 | 24790 | sidereon_look_angles_values | support.LookAnglesValues | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 503 | 24801 | sidereon_mad_gaussian_consistency | errormetrics.MADGaussianConsistency | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 504 | 24809 | sidereon_mad_spread | errormetrics.MADSpread | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 505 | 24814 | sidereon_mean_to_eccentric_anomaly | astro.MeanToEccentricAnomaly | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 506 | 24818 | sidereon_mean_to_true_anomaly | astro.MeanToTrueAnomaly | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 507 | 24822 | sidereon_mee2coe | support.Mee2coe | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 508 | 24825 | sidereon_mee2rv | support.Mee2rv | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 509 | 24838 | sidereon_met_init | support.MetInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 510 | 24846 | sidereon_mmap_terrain_checksum64 | distribution.MmapTerrainChecksum64 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 511 | 24856 | sidereon_mmap_terrain_digest_provenance | distribution.MmapTerrainDigestProvenance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 512 | 24866 | sidereon_mmap_terrain_ellipsoidal_height_m | distribution.MmapTerrainEllipsoidalHeightM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 513 | 24881 | sidereon_mmap_terrain_ellipsoidal_height_m_with_model | distribution.MmapTerrainEllipsoidalHeightMWithModel | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 514 | 24898 | sidereon_mmap_terrain_ellipsoidal_height_m_with_options | distribution.MmapTerrainEllipsoidalHeightMWithOptions | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 515 | 24909 | sidereon_mmap_terrain_free | distribution.MmapTerrain.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 516 | 24919 | sidereon_mmap_terrain_from_bytes | distribution.MmapTerrainFromBytes | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 517 | 24930 | sidereon_mmap_terrain_from_path | distribution.MmapTerrainFromPath | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 518 | 24942 | sidereon_mmap_terrain_from_path_attested | distribution.MmapTerrainFromPathAttested | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 519 | 24953 | sidereon_mmap_terrain_from_vec | distribution.MmapTerrainFromVec | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 520 | 24967 | sidereon_mmap_terrain_height_batch | distribution.MmapTerrainHeightBatch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 521 | 24980 | sidereon_mmap_terrain_height_m | distribution.MmapTerrainHeightM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 522 | 24993 | sidereon_mmap_terrain_height_m_with_options | distribution.MmapTerrainHeightMWithOptions | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 523 | 25009 | sidereon_mmap_terrain_orthometric_height_batch | distribution.MmapTerrainOrthometricHeightBatch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 524 | 25022 | sidereon_mmap_terrain_orthometric_height_m | distribution.MmapTerrainOrthometricHeightM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 525 | 25036 | sidereon_mmap_terrain_orthometric_height_m_with_options | distribution.MmapTerrainOrthometricHeightMWithOptions | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 526 | 25049 | sidereon_mmap_terrain_tile_index | distribution.MmapTerrainTileIndex | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 527 | 25062 | sidereon_mmap_terrain_to_bytes | distribution.MmapTerrainToBytes | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 528 | 25074 | sidereon_mmap_terrain_verify | distribution.MmapTerrainVerify | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 529 | 25083 | sidereon_mmap_terrain_vertical_datum | distribution.MmapTerrainVerticalDatum | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 530 | 25092 | sidereon_moon_angle_deg | astro.MoonAngleDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 531 | 25104 | sidereon_moon_az_el | astro.MoonAzEl | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 532 | 25118 | sidereon_moon_elevation_deg | astro.MoonElevationDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 533 | 25128 | sidereon_moon_elevation_options_init | astro.MoonElevationOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 534 | 25137 | sidereon_moon_illumination | astro.MoonIllumination | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 535 | 25147 | sidereon_moving_baseline_solution_epoch | support.MovingBaselineSolutionEpoch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 536 | 25156 | sidereon_moving_baseline_solution_epoch_count | support.MovingBaselineSolutionEpochCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 537 | 25164 | sidereon_moving_baseline_solution_free | support.MovingBaselineSolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 538 | 25172 | sidereon_nav_parse_free | positioning.NAVParse.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 539 | 25180 | sidereon_nav_parse_record | positioning.NAVParseRecord | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 540 | 25189 | sidereon_nav_parse_record_count | positioning.NAVParseRecordCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 541 | 25197 | sidereon_nav_parse_skipped | positioning.NAVParseSkipped | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 542 | 25206 | sidereon_nav_parse_skipped_count | positioning.NAVParseSkippedCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 543 | 25217 | sidereon_nav_parse_skipped_message | positioning.NAVParseSkippedMessage | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 544 | 25230 | sidereon_navcen_assessment | distribution.NAVCENAssessment | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 545 | 25239 | sidereon_navcen_assessment_count | distribution.NAVCENAssessmentCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 546 | 25246 | sidereon_navcen_assessment_nanu_subject | distribution.NAVCENAssessmentNanuSubject | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 547 | 25257 | sidereon_navcen_assessment_nanu_type | distribution.NAVCENAssessmentNanuType | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 548 | 25269 | sidereon_navcen_assessment_outage_start | distribution.NAVCENAssessmentOutageStart | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 549 | 25282 | sidereon_navcen_assessments_free | distribution.NAVCENAssessments.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 550 | 25292 | sidereon_navcen_parse_at | distribution.NAVCENParseAt | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 551 | 25304 | sidereon_nequick_g_delay_m | geodesy.NequickGDelayM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 552 | 25320 | sidereon_nequick_g_stec_tecu | geodesy.NequickGStecTecu | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 553 | 25331 | sidereon_nis | errormetrics.NIS | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 554 | 25338 | sidereon_nis_expected_value | errormetrics.NISExpectedValue | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 555 | 25345 | sidereon_nis_gate_test | errormetrics.NISGateTest | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 556 | 25356 | sidereon_nis_gate_threshold | errormetrics.NISGateThreshold | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 557 | 25360 | sidereon_nmea_accumulator_epochs | support.NMEAAccumulatorEpochs | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 558 | 25366 | sidereon_nmea_accumulator_finish | support.NMEAAccumulatorFinish | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 559 | 25369 | sidereon_nmea_accumulator_free | support.NMEAAccumulator.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 560 | 25371 | sidereon_nmea_accumulator_new | support.NMEAAccumulatorNew | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 561 | 25373 | sidereon_nmea_accumulator_push | support.NMEAAccumulatorPush | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 562 | 25378 | sidereon_nmea_accumulator_retained_len | support.NMEAAccumulatorRetainedLen | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 563 | 25381 | sidereon_nmea_accumulator_summary | support.NMEAAccumulatorSummary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 564 | 25384 | sidereon_nmea_log_epochs | support.NMEALogEpochs | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 565 | 25390 | sidereon_nmea_log_free | support.NMEALog.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 566 | 25392 | sidereon_nmea_log_summary | support.NMEALogSummary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 567 | 25395 | sidereon_nmea_parse | support.NMEAParse | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 568 | 25399 | sidereon_nmea_write_gga | support.NMEAWriteGga | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 569 | 25417 | sidereon_normal_covariance | errormetrics.NormalCovariance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 570 | 25431 | sidereon_normalized_innovation | support.NormalizedInnovation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 571 | 25435 | sidereon_ntrip_bytes | distribution.NTRIPBytes | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 572 | 25441 | sidereon_ntrip_bytes_free | distribution.NTRIPBytes.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 573 | 25443 | sidereon_ntrip_events_count | distribution.NTRIPEventsCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 574 | 25446 | sidereon_ntrip_events_detail | distribution.NTRIPEventsDetail | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 575 | 25453 | sidereon_ntrip_events_event | distribution.NTRIPEventsEvent | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 576 | 25457 | sidereon_ntrip_events_free | distribution.NTRIPEvents.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 577 | 25459 | sidereon_ntrip_events_payload | distribution.NTRIPEventsPayload | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 578 | 25466 | sidereon_ntrip_events_sourcetable | distribution.NTRIPEventsSourcetable | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 579 | 25470 | sidereon_ntrip_machine_connection_request | distribution.NTRIPMachineConnectionRequest | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 580 | 25473 | sidereon_ntrip_machine_finish | distribution.NTRIPMachineFinish | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 581 | 25476 | sidereon_ntrip_machine_free | distribution.NTRIPMachine.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 582 | 25478 | sidereon_ntrip_machine_new | distribution.NTRIPMachineNew | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 583 | 25481 | sidereon_ntrip_machine_push | distribution.NTRIPMachinePush | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 584 | 25486 | sidereon_ntrip_machine_reset | distribution.NTRIPMachineReset | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 585 | 25488 | sidereon_ntrip_machine_state | distribution.NTRIPMachineState | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 586 | 25491 | sidereon_ntrip_machine_try_gga_message | distribution.NTRIPMachineTryGgaMessage | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 587 | 25498 | sidereon_ntrip_request_bytes | distribution.NTRIPRequestBytes | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 588 | 25504 | sidereon_ntrip_sourcetable_free | distribution.NTRIPSourcetable.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 589 | 25506 | sidereon_ntrip_sourcetable_parse | distribution.NTRIPSourcetableParse | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 590 | 25510 | sidereon_ntrip_sourcetable_streams | distribution.NTRIPSourcetableStreams | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 591 | 25516 | sidereon_ntrip_sourcetable_summary | distribution.NTRIPSourcetableSummary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 592 | 25519 | sidereon_ntrip_sourcetable_to_text | distribution.NTRIPSourcetableToText | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 593 | 25532 | sidereon_nutation_equation_of_equinoxes_terms | geodesy.NutationEquationOfEquinoxesTerms | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 594 | 25541 | sidereon_nutation_fundamental_arguments | geodesy.NutationFundamentalArguments | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 595 | 25550 | sidereon_nutation_iau2000a_radians | geodesy.NutationIau2000aRadians | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 596 | 25561 | sidereon_nutation_matrix | geodesy.NutationMatrix | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 597 | 25572 | sidereon_nutation_mean_obliquity_radians | astro.NutationMeanObliquityRadians | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 598 | 25582 | sidereon_observability_tier_label | errormetrics.ObservabilityTierLabel | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 599 | 25595 | sidereon_observable_state_missing_position_ecef_m | errormetrics.ObservableStateMissingPositionECEFM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 600 | 25604 | sidereon_observables_options_init | positioning.ObservablesOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 601 | 25606 | sidereon_observation_qc_clock_jumps | positioning.ObservationQcClockJumps | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 602 | 25612 | sidereon_observation_qc_cycle_slip_systems | positioning.ObservationQcCycleSlipSystems | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 603 | 25618 | sidereon_observation_qc_cycle_slips | positioning.ObservationQcCycleSlips | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 604 | 25621 | sidereon_observation_qc_from_obs | positioning.ObservationQcFromObs | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 605 | 25625 | sidereon_observation_qc_gaps | positioning.ObservationQcGaps | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 606 | 25631 | sidereon_observation_qc_multipath_satellites | positioning.ObservationQcMultipathSatellites | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 607 | 25637 | sidereon_observation_qc_multipath_systems | positioning.ObservationQcMultipathSystems | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 608 | 25643 | sidereon_observation_qc_options_init | positioning.ObservationQcOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 609 | 25645 | sidereon_observation_qc_parse | positioning.ObservationQcParse | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 610 | 25657 | sidereon_observation_qc_render_html | positioning.ObservationQcRenderHtml | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 611 | 25670 | sidereon_observation_qc_render_text | positioning.ObservationQcRenderText | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 612 | 25676 | sidereon_observation_qc_report_free | positioning.ObservationQcReport.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 613 | 25678 | sidereon_observation_qc_satellite_signals | positioning.ObservationQcSatelliteSignals | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 614 | 25684 | sidereon_observation_qc_satellites | positioning.ObservationQcSatellites | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 615 | 25690 | sidereon_observation_qc_summary | positioning.ObservationQcSummary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 616 | 25693 | sidereon_observation_qc_system_signals | positioning.ObservationQcSystemSignals | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 617 | 25705 | sidereon_observation_qc_to_json | positioning.ObservationQcToJSON | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 618 | 25711 | sidereon_observe | support.Observe | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 619 | 25721 | sidereon_observe_options_init | support.ObserveOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 620 | 25723 | sidereon_observe_spk_body | astro.ObserveSPKBody | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 621 | 25736 | sidereon_ocean_tide_loading | geodesy.OceanTideLoading | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 622 | 25750 | sidereon_oem_free | astro.OEM.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 623 | 25759 | sidereon_oem_parse_kvn | astro.OEMParseKVN | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 624 | 25770 | sidereon_oem_parse_xml | astro.OEMParseXML | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 625 | 25780 | sidereon_oem_segment_count | astro.OEMSegmentCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 626 | 25792 | sidereon_oem_to_kvn | astro.OEMToKVN | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 627 | 25808 | sidereon_oem_to_xml | astro.OEMToXML | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 628 | 25833 | sidereon_omm_catalog_build_lenient | astro.OMMCatalogBuildLenient | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 629 | 25845 | sidereon_omm_catalog_free | astro.OMMCatalog.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 630 | 25855 | sidereon_omm_catalog_malformed_count | astro.OMMCatalogMalformedCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 631 | 25867 | sidereon_omm_catalog_record | astro.OMMCatalogRecord | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 632 | 25877 | sidereon_omm_catalog_record_count | astro.OMMCatalogRecordCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 633 | 25889 | sidereon_omm_catalog_skipped | astro.OMMCatalogSkipped | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 634 | 25898 | sidereon_omm_catalog_skipped_count | astro.OMMCatalogSkippedCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 635 | 25913 | sidereon_omm_catalog_skipped_object_name | astro.OMMCatalogSkippedObjectName | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 636 | 25925 | sidereon_omm_free | astro.OMM.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 637 | 25933 | sidereon_omm_parse_json | astro.OMMParseJSON | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 638 | 25943 | sidereon_omm_parse_kvn | astro.OMMParseKVN | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 639 | 25953 | sidereon_omm_parse_xml | astro.OMMParseXML | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 640 | 25965 | sidereon_omm_to_json | astro.OMMToJSON | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 641 | 25979 | sidereon_omm_to_kvn | astro.OMMToKVN | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 642 | 25993 | sidereon_omm_to_xml | astro.OMMToXML | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 643 | 26005 | sidereon_opm_free | astro.OPM.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 644 | 26014 | sidereon_opm_parse_kvn | astro.OPMParseKVN | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 645 | 26025 | sidereon_opm_parse_xml | astro.OPMParseXML | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 646 | 26039 | sidereon_opm_to_kvn | astro.OPMToKVN | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 647 | 26055 | sidereon_opm_to_xml | astro.OPMToXML | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 648 | 26066 | sidereon_orbit_fit_options_init | astro.OrbitFitOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 649 | 26073 | sidereon_orbit_fit_report_arc_span | astro.OrbitFitReportArcSpan | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 650 | 26082 | sidereon_orbit_fit_report_constellation_ledger | astro.OrbitFitReportConstellationLedger | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 651 | 26094 | sidereon_orbit_fit_report_fits | astro.OrbitFitReportFits | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 652 | 26105 | sidereon_orbit_fit_report_free | astro.OrbitFitReport.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 653 | 26113 | sidereon_orbit_fit_report_satellite_ledger | astro.OrbitFitReportSatelliteLedger | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 654 | 26126 | sidereon_orthometric_height_m | geodesy.OrthometricHeightM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 655 | 26137 | sidereon_parallactic_angle_deg | geodesy.ParallacticAngleDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 656 | 26153 | sidereon_parse_rinex_glonass_records | positioning.ParseRINEXGLONASSRecords | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 657 | 26164 | sidereon_parse_rinex_iono_corrections | positioning.ParseRINEXIonoCorrections | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 658 | 26175 | sidereon_parse_rinex_leap_seconds | positioning.ParseRINEXLeapSeconds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 659 | 26188 | sidereon_parse_rinex_nav_lenient | positioning.ParseRINEXNAVLenient | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 660 | 26199 | sidereon_parse_rinex_nav_records | positioning.ParseRINEXNAVRecords | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 661 | 26210 | sidereon_parse_sbas_ems_lines | distribution.ParseSBASEmsLines | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 662 | 26221 | sidereon_parse_sbas_rtklib_lines | distribution.ParseSBASRtklibLines | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 663 | 26239 | sidereon_parse_tle_file | astro.ParseTLEFile | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 664 | 26249 | sidereon_pass_finder_options_init | astro.PassFinderOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 665 | 26256 | sidereon_pass_list_count | astro.PassListCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 666 | 26267 | sidereon_pass_list_free | astro.PassList.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 667 | 26277 | sidereon_pass_list_values | astro.PassListValues | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 668 | 26289 | sidereon_phase_angle_deg | support.PhaseAngleDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 669 | 26301 | sidereon_position_angle_deg | support.PositionAngleDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 670 | 26314 | sidereon_ppp_auto_init_options_init | positioning.PPPAutoInitOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 671 | 26327 | sidereon_ppp_corrections_build | positioning.PPPCorrectionsBuild | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 672 | 26342 | sidereon_ppp_corrections_code_bias | positioning.PPPCorrectionsCodeBias | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 673 | 26354 | sidereon_ppp_corrections_free | positioning.PPPCorrections.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 674 | 26364 | sidereon_ppp_corrections_ocean_loading | positioning.PPPCorrectionsOceanLoading | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 675 | 26378 | sidereon_ppp_corrections_pole_tide | positioning.PPPCorrectionsPoleTide | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 676 | 26392 | sidereon_ppp_corrections_sat_pco_ecef | positioning.PPPCorrectionsSatPcoECEF | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 677 | 26406 | sidereon_ppp_corrections_sat_pcv | positioning.PPPCorrectionsSatPcv | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 678 | 26420 | sidereon_ppp_corrections_tide | positioning.PPPCorrectionsTide | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 679 | 26434 | sidereon_ppp_corrections_windup | positioning.PPPCorrectionsWindup | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 680 | 26445 | sidereon_ppp_fixed_ambiguity_options_init | positioning.PPPFixedAmbiguityOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 681 | 26455 | sidereon_ppp_fixed_solution_fixed_ambiguities | positioning.PPPFixedSolutionFixedAmbiguities | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 682 | 26468 | sidereon_ppp_fixed_solution_float_position | positioning.PPPFixedSolutionFloatPosition | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 683 | 26480 | sidereon_ppp_fixed_solution_free | positioning.PPPFixedSolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 684 | 26488 | sidereon_ppp_fixed_solution_metadata | positioning.PPPFixedSolutionMetadata | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 685 | 26497 | sidereon_ppp_fixed_solution_position | positioning.PPPFixedSolutionPosition | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 686 | 26507 | sidereon_ppp_fixed_solution_position_covariances | positioning.PPPFixedSolutionPositionCovariances | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 687 | 26516 | sidereon_ppp_fixed_solution_temporal_correlation | positioning.PPPFixedSolutionTemporalCorrelation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 688 | 26525 | sidereon_ppp_fixed_solution_tropo_gradient | positioning.PPPFixedSolutionTropoGradient | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 689 | 26537 | sidereon_ppp_fixed_solution_used_ids | positioning.PPPFixedSolutionUsedIds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 690 | 26554 | sidereon_ppp_fixed_solution_used_sat_ids | positioning.PPPFixedSolutionUsedSatIds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 691 | 26565 | sidereon_ppp_float_options_init | positioning.PPPFloatOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 692 | 26575 | sidereon_ppp_float_solution_ambiguities | positioning.PPPFloatSolutionAmbiguities | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 693 | 26589 | sidereon_ppp_float_solution_free | positioning.PPPFloatSolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 694 | 26597 | sidereon_ppp_float_solution_metadata | positioning.PPPFloatSolutionMetadata | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 695 | 26606 | sidereon_ppp_float_solution_position | positioning.PPPFloatSolutionPosition | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 696 | 26616 | sidereon_ppp_float_solution_position_covariances | positioning.PPPFloatSolutionPositionCovariances | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 697 | 26625 | sidereon_ppp_float_solution_temporal_correlation | positioning.PPPFloatSolutionTemporalCorrelation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 698 | 26634 | sidereon_ppp_float_solution_tropo_gradient | positioning.PPPFloatSolutionTropoGradient | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 699 | 26646 | sidereon_ppp_float_solution_used_ids | positioning.PPPFloatSolutionUsedIds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 700 | 26663 | sidereon_ppp_float_solution_used_sat_ids | positioning.PPPFloatSolutionUsedSatIds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 701 | 26674 | sidereon_ppp_measurement_weights_init | positioning.PPPMeasurementWeightsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 702 | 26681 | sidereon_ppp_range_corrections_init | positioning.PPPRangeCorrectionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 703 | 26688 | sidereon_ppp_troposphere_options_init | positioning.PPPTroposphereOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 704 | 26696 | sidereon_precession_icrs_to_j2000_matrix | geodesy.PrecessionIcrsToJ2000Matrix | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 705 | 26705 | sidereon_precession_matrix | geodesy.PrecessionMatrix | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 706 | 26713 | sidereon_precise_ephemeris_interpolant_free | positioning.PreciseEphemerisInterpolant.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 707 | 26723 | sidereon_precise_ephemeris_interpolant_from_precise_ephemeris_samples | positioning.PreciseEphemerisInterpolantFromPreciseEphemerisSamples | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 708 | 26734 | sidereon_precise_ephemeris_interpolant_from_samples | positioning.PreciseEphemerisInterpolantFromSamples | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 709 | 26746 | sidereon_precise_ephemeris_interpolant_from_sp3 | positioning.PreciseEphemerisInterpolantFromSP3 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 710 | 26757 | sidereon_precise_ephemeris_interpolant_observable_states_at_j2000_s | positioning.PreciseEphemerisInterpolantObservableStatesAtJ2000S | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 711 | 26774 | sidereon_precise_ephemeris_interpolant_observable_states_at_shared_j2000_s | positioning.PreciseEphemerisInterpolantObservableStatesAtSharedJ2000S | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 712 | 26793 | sidereon_precise_ephemeris_samples_free | positioning.PreciseEphemerisSamples.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 713 | 26807 | sidereon_precise_ephemeris_samples_from_samples | positioning.PreciseEphemerisSamplesFromSamples | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 714 | 26819 | sidereon_precise_ephemeris_samples_observable_states_at_j2000_s | positioning.PreciseEphemerisSamplesObservableStatesAtJ2000S | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 715 | 26836 | sidereon_precise_ephemeris_samples_observable_states_at_shared_j2000_s | positioning.PreciseEphemerisSamplesObservableStatesAtSharedJ2000S | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 716 | 26858 | sidereon_precise_ephemeris_samples_predict_ranges | positioning.PreciseEphemerisSamplesPredictRanges | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 717 | 26871 | sidereon_precise_ephemeris_samples_sample | positioning.PreciseEphemerisSamplesSample | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 718 | 26888 | sidereon_precise_interpolant_artifact_checksum64 | positioning.PreciseInterpolantArtifactChecksum64 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 719 | 26899 | sidereon_precise_interpolant_artifact_digest_provenance | positioning.PreciseInterpolantArtifactDigestProvenance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 720 | 26908 | sidereon_precise_interpolant_artifact_free | positioning.PreciseInterpolantArtifact.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 721 | 26918 | sidereon_precise_interpolant_artifact_from_path | positioning.PreciseInterpolantArtifactFromPath | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 722 | 26933 | sidereon_precise_interpolant_artifact_from_path_attested | positioning.PreciseInterpolantArtifactFromPathAttested | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 723 | 26943 | sidereon_precise_interpolant_artifact_handle_checksum64 | positioning.PreciseInterpolantArtifactHandleChecksum64 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 724 | 26954 | sidereon_precise_interpolant_artifact_open_borrowed | positioning.PreciseInterpolantArtifactOpenBorrowed | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 725 | 26965 | sidereon_precise_interpolant_artifact_open_owned | positioning.PreciseInterpolantArtifactOpenOwned | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 726 | 26977 | sidereon_precise_interpolant_artifact_satellites | positioning.PreciseInterpolantArtifactSatellites | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 727 | 26989 | sidereon_precise_interpolant_artifact_state | positioning.PreciseInterpolantArtifactState | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 728 | 27001 | sidereon_precise_interpolant_artifact_verify | positioning.PreciseInterpolantArtifactVerify | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 729 | 27015 | sidereon_prepare_ionosphere_free_rtk_arc | positioning.PrepareIonosphereFreeRTKArc | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 730 | 27022 | sidereon_propagate_covariance | astro.PropagateCovariance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 731 | 27029 | sidereon_propagate_kepler | astro.PropagateKepler | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 732 | 27042 | sidereon_propagate_state | astro.PropagateState | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 733 | 27057 | sidereon_propagate_tle_batch | astro.PropagateTLEBatch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 734 | 27072 | sidereon_pseudorange_variance | positioning.PseudorangeVariance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 735 | 27081 | sidereon_pseudorange_variance_options_init | positioning.PseudorangeVarianceOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 736 | 27094 | sidereon_raim | support.RAIM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 737 | 27115 | sidereon_raim_fde_design | errormetrics.RAIMFDEDesign | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 738 | 27129 | sidereon_raim_for_solution | support.RAIMForSolution | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 739 | 27146 | sidereon_raim_normalized_residuals | support.RAIMNormalizedResiduals | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 740 | 27166 | sidereon_range_fde_options_init | positioning.RangeFDEOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 741 | 27175 | sidereon_range_fde_result_covariance | positioning.RangeFDEResultCovariance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 742 | 27188 | sidereon_range_fde_result_diagnostics | positioning.RangeFDEResultDiagnostics | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 743 | 27201 | sidereon_range_fde_result_excluded | positioning.RangeFDEResultExcluded | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 744 | 27212 | sidereon_range_fde_result_free | positioning.RangeFDEResult.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 745 | 27219 | sidereon_range_fde_result_global_test | positioning.RangeFDEResultGlobalTest | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 746 | 27227 | sidereon_range_fde_result_iterations | positioning.RangeFDEResultIterations | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 747 | 27237 | sidereon_range_fde_result_state_correction | positioning.RangeFDEResultStateCorrection | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 748 | 27248 | sidereon_range_fde_result_state_dim | positioning.RangeFDEResultStateDim | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 749 | 27262 | sidereon_reduced_orbit_drift | astro.ReducedOrbitDrift | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 750 | 27277 | sidereon_reduced_orbit_drift_report_entries | astro.ReducedOrbitDriftReportEntries | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 751 | 27290 | sidereon_reduced_orbit_drift_report_free | astro.ReducedOrbitDriftReport.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 752 | 27298 | sidereon_reduced_orbit_drift_report_requested_samples | astro.ReducedOrbitDriftReportRequestedSamples | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 753 | 27308 | sidereon_reduced_orbit_drift_report_summary | astro.ReducedOrbitDriftReportSummary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 754 | 27321 | sidereon_reduced_orbit_drift_sp3_source | positioning.ReducedOrbitDriftSP3Source | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 755 | 27336 | sidereon_reduced_orbit_drift_tle_source | astro.ReducedOrbitDriftTLESource | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 756 | 27352 | sidereon_reduced_orbit_fit | astro.ReducedOrbitFit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 757 | 27368 | sidereon_reduced_orbit_fit_piecewise | astro.ReducedOrbitFitPiecewise | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 758 | 27385 | sidereon_reduced_orbit_fit_piecewise_sp3_source | positioning.ReducedOrbitFitPiecewiseSP3Source | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 759 | 27399 | sidereon_reduced_orbit_fit_piecewise_tle_source | astro.ReducedOrbitFitPiecewiseTLESource | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 760 | 27413 | sidereon_reduced_orbit_fit_sp3_source | positioning.ReducedOrbitFitSP3Source | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 761 | 27426 | sidereon_reduced_orbit_fit_tle_source | astro.ReducedOrbitFitTLESource | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 762 | 27439 | sidereon_reduced_orbit_piecewise_drift | astro.ReducedOrbitPiecewiseDrift | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 763 | 27454 | sidereon_reduced_orbit_piecewise_drift_sp3_source | positioning.ReducedOrbitPiecewiseDriftSP3Source | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 764 | 27467 | sidereon_reduced_orbit_piecewise_drift_tle_source | astro.ReducedOrbitPiecewiseDriftTLESource | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 765 | 27478 | sidereon_reduced_orbit_piecewise_free | astro.ReducedOrbitPiecewise.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 766 | 27486 | sidereon_reduced_orbit_piecewise_info | astro.ReducedOrbitPiecewiseInfo | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 767 | 27496 | sidereon_reduced_orbit_piecewise_position | astro.ReducedOrbitPiecewisePosition | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 768 | 27509 | sidereon_reduced_orbit_piecewise_position_velocity | astro.ReducedOrbitPiecewisePositionVelocity | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 769 | 27523 | sidereon_reduced_orbit_piecewise_segments | astro.ReducedOrbitPiecewiseSegments | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 770 | 27536 | sidereon_reduced_orbit_piecewise_select_segment | astro.ReducedOrbitPiecewiseSelectSegment | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 771 | 27551 | sidereon_reduced_orbit_position | astro.ReducedOrbitPosition | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 772 | 27568 | sidereon_reduced_orbit_position_velocity | astro.ReducedOrbitPositionVelocity | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 773 | 27575 | sidereon_relative_mean_motion_circular | astro.RelativeMeanMotionCircular | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 774 | 27577 | sidereon_relative_mean_motion_from_state | astro.RelativeMeanMotionFromState | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 775 | 27580 | sidereon_relative_rotation | support.RelativeRotation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 776 | 27585 | sidereon_relative_state | support.RelativeState | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 777 | 27598 | sidereon_reliability_araim | errormetrics.ReliabilityAraim | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 778 | 27613 | sidereon_reliability_design | errormetrics.ReliabilityDesign | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 779 | 27623 | sidereon_reliability_options_init | errormetrics.ReliabilityOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 780 | 27630 | sidereon_reliability_report_free | errormetrics.ReliabilityReport.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 781 | 27641 | sidereon_reliability_report_observations | errormetrics.ReliabilityReportObservations | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 782 | 27653 | sidereon_reliability_report_summary | errormetrics.ReliabilityReportSummary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 783 | 27664 | sidereon_residual_jarque_bera | errormetrics.ResidualJarqueBera | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 784 | 27677 | sidereon_residual_kurtosis | errormetrics.ResidualKurtosis | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 785 | 27692 | sidereon_residual_moments | errormetrics.ResidualMoments | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 786 | 27707 | sidereon_residual_shapiro_wilk | errormetrics.ResidualShapiroWilk | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 787 | 27720 | sidereon_residual_skewness | errormetrics.ResidualSkewness | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 788 | 27728 | sidereon_rf_cn0 | support.RFCn0 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 789 | 27740 | sidereon_rf_dish_gain | support.RFDishGain | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 790 | 27751 | sidereon_rf_eirp | support.RFEirp | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 791 | 27758 | sidereon_rf_fspl | support.RFFspl | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 792 | 27766 | sidereon_rf_link_margin | support.RFLinkMargin | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 793 | 27774 | sidereon_rf_wavelength | support.RFWavelength | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 794 | 27785 | sidereon_rinex_band_frequency_hz | positioning.RINEXBandFrequencyHz | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 795 | 27799 | sidereon_rinex_band_wavelength_m | positioning.RINEXBandWavelengthM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 796 | 27814 | sidereon_rinex_clock_bias_at_gps_seconds | positioning.RINEXClockBiasAtGPSSeconds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 797 | 27825 | sidereon_rinex_clock_free | positioning.RINEXClock.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 798 | 27834 | sidereon_rinex_clock_parse | positioning.RINEXClockParse | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 799 | 27845 | sidereon_rinex_clock_parse_lossy | positioning.RINEXClockParseLossy | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 800 | 27854 | sidereon_rinex_clock_sample_count | positioning.RINEXClockSampleCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 801 | 27862 | sidereon_rinex_clock_satellite_count | positioning.RINEXClockSatelliteCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 802 | 27873 | sidereon_rinex_clock_satellites | positioning.RINEXClockSatellites | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 803 | 27885 | sidereon_rinex_clock_series | positioning.RINEXClockSeries | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 804 | 27894 | sidereon_rinex_clock_series_count | positioning.RINEXClockSeriesCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 805 | 27904 | sidereon_rinex_clock_series_for | positioning.RINEXClockSeriesFor | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 806 | 27913 | sidereon_rinex_clock_series_free | positioning.RINEXClockSeries.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 807 | 27920 | sidereon_rinex_clock_series_sample_count | positioning.RINEXClockSeriesSampleCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 808 | 27930 | sidereon_rinex_clock_series_samples | positioning.RINEXClockSeriesSamples | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 809 | 27941 | sidereon_rinex_clock_series_satellite | positioning.RINEXClockSeriesSatellite | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 810 | 27952 | sidereon_rinex_clock_to_text | positioning.RINEXClockToText | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 811 | 27968 | sidereon_rinex_encode_nav | positioning.RINEXEncodeNav | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 812 | 27979 | sidereon_rinex_glonass_records_count | positioning.RINEXGLONASSRecordsCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 813 | 27988 | sidereon_rinex_glonass_records_free | positioning.RINEXGLONASSRecords.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 814 | 27995 | sidereon_rinex_glonass_records_item | positioning.RINEXGLONASSRecordsItem | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 815 | 28005 | sidereon_rinex_glonass_records_skipped_count | positioning.RINEXGLONASSRecordsSkippedCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 816 | 28015 | sidereon_rinex_glonass_records_skipped_item | positioning.RINEXGLONASSRecordsSkippedItem | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 817 | 28019 | sidereon_rinex_lint_findings | positioning.RINEXLintFindings | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 818 | 28025 | sidereon_rinex_lint_nav | positioning.RINEXLintNav | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 819 | 28029 | sidereon_rinex_lint_obs | positioning.RINEXLintObs | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 820 | 28033 | sidereon_rinex_lint_report_free | positioning.RINEXLintReport.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 821 | 28035 | sidereon_rinex_lint_summary | positioning.RINEXLintSummary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 822 | 28043 | sidereon_rinex_nav_records_count | positioning.RINEXNAVRecordsCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 823 | 28052 | sidereon_rinex_nav_records_free | positioning.RINEXNAVRecords.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 824 | 28060 | sidereon_rinex_nav_records_item | positioning.RINEXNAVRecordsItem | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 825 | 28071 | sidereon_rinex_obs_carrier_phase | positioning.RINEXObsCarrierPhase | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 826 | 28086 | sidereon_rinex_obs_codes | positioning.RINEXObsCodes | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 827 | 28099 | sidereon_rinex_obs_epoch_count | positioning.RINEXObsEpochCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 828 | 28110 | sidereon_rinex_obs_epochs | positioning.RINEXObsEpochs | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 829 | 28123 | sidereon_rinex_obs_free | positioning.RINEXObs.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 830 | 28131 | sidereon_rinex_obs_header | positioning.RINEXObsHeader | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 831 | 28142 | sidereon_rinex_obs_load | positioning.RINEXObsLoad | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 832 | 28160 | sidereon_rinex_obs_observation | positioning.RINEXObsObservation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 833 | 28176 | sidereon_rinex_obs_parse | positioning.RINEXObsParse | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 834 | 28187 | sidereon_rinex_obs_pseudoranges | positioning.RINEXObsPseudoranges | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 835 | 28202 | sidereon_rinex_obs_receiver_clock_phase_deviations | positioning.RINEXObsReceiverClockPhaseDeviations | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 836 | 28218 | sidereon_rinex_obs_to_rinex_text | positioning.RINEXObsToRINEXText | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 837 | 28232 | sidereon_rinex_obs_values | positioning.RINEXObsValues | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 838 | 28245 | sidereon_rinex_obs_version | positioning.RINEXObsVersion | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 839 | 28257 | sidereon_rinex_observation_frequency_hz | positioning.RINEXObservationFrequencyHz | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 840 | 28273 | sidereon_rinex_observation_wavelength_m | positioning.RINEXObservationWavelengthM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 841 | 28280 | sidereon_rinex_repair_actions | positioning.RINEXRepairActions | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 842 | 28286 | sidereon_rinex_repair_crinex_text | positioning.RINEXRepairCrinexText | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 843 | 28292 | sidereon_rinex_repair_free | positioning.RINEXRepair.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 844 | 28294 | sidereon_rinex_repair_nav | positioning.RINEXRepairNav | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 845 | 28299 | sidereon_rinex_repair_obs | positioning.RINEXRepairObs | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 846 | 28304 | sidereon_rinex_repair_options_init | positioning.RINEXRepairOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 847 | 28306 | sidereon_rinex_repair_summary | positioning.RINEXRepairSummary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 848 | 28309 | sidereon_rinex_repair_text | positioning.RINEXRepairText | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 849 | 28320 | sidereon_rinex_spp_inputs_count | positioning.RINEXSPPInputsCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 850 | 28329 | sidereon_rinex_spp_inputs_epoch | positioning.RINEXSPPInputsEpoch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 851 | 28341 | sidereon_rinex_spp_inputs_epoch_inputs | positioning.RINEXSPPInputsEpochInputs | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 852 | 28351 | sidereon_rinex_spp_inputs_free | positioning.RINEXSPPInputs.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 853 | 28360 | sidereon_rinex_spp_options_init | positioning.RINEXSPPOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 854 | 28369 | sidereon_rinex_spp_solution | positioning.RINEXSPPSolution | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 855 | 28381 | sidereon_rinex_spp_solution_error | positioning.RINEXSPPSolutionError | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 856 | 28393 | sidereon_rinex_spp_solution_ok | positioning.RINEXSPPSolutionOk | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 857 | 28402 | sidereon_rinex_spp_solutions_count | positioning.RINEXSPPSolutionsCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 858 | 28411 | sidereon_rinex_spp_solutions_epoch | positioning.RINEXSPPSolutionsEpoch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 859 | 28421 | sidereon_rinex_spp_solutions_free | positioning.RINEXSPPSolutions.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 860 | 28423 | sidereon_robust_fde_solve_broadcast | positioning.RobustFDESolveBroadcast | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 861 | 28429 | sidereon_robust_fde_solve_spp | positioning.RobustFDESolveSPP | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 862 | 28445 | sidereon_rtcm_build_antenna_descriptor | positioning.RTCMBuildAntennaDescriptor | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 863 | 28463 | sidereon_rtcm_build_beidou_ephemeris | positioning.RTCMBuildBeidouEphemeris | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 864 | 28474 | sidereon_rtcm_build_galileo_fnav_ephemeris | positioning.RTCMBuildGalileoFnavEphemeris | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 865 | 28485 | sidereon_rtcm_build_galileo_inav_ephemeris | positioning.RTCMBuildGalileoInavEphemeris | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 866 | 28496 | sidereon_rtcm_build_glonass_ephemeris | positioning.RTCMBuildGLONASSEphemeris | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 867 | 28507 | sidereon_rtcm_build_gps_ephemeris | positioning.RTCMBuildGPSEphemeris | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 868 | 28521 | sidereon_rtcm_build_msm | positioning.RTCMBuildMsm | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 869 | 28536 | sidereon_rtcm_build_qzss_ephemeris | positioning.RTCMBuildQzssEphemeris | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 870 | 28548 | sidereon_rtcm_build_station_coordinates | positioning.RTCMBuildStationCoordinates | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 871 | 28560 | sidereon_rtcm_decode_frame | positioning.RTCMDecodeFrame | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 872 | 28577 | sidereon_rtcm_decode_messages | positioning.RTCMDecodeMessages | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 873 | 28592 | sidereon_rtcm_decode_stream | positioning.RTCMDecodeStream | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 874 | 28605 | sidereon_rtcm_derive_lli | positioning.RTCMDeriveLli | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 875 | 28620 | sidereon_rtcm_encode_frame | positioning.RTCMEncodeFrame | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 876 | 28634 | sidereon_rtcm_frame_body | positioning.RTCMFrameBody | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 877 | 28646 | sidereon_rtcm_frame_len | positioning.RTCMFrameLen | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 878 | 28655 | sidereon_rtcm_frames_count | positioning.RTCMFramesCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 879 | 28663 | sidereon_rtcm_frames_free | positioning.RTCMFrames.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 880 | 28671 | sidereon_rtcm_lli_bits | positioning.RTCMLliBits | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 881 | 28678 | sidereon_rtcm_lock_time_tracker_free | positioning.RTCMLockTimeTracker.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 882 | 28685 | sidereon_rtcm_lock_time_tracker_new | positioning.RTCMLockTimeTrackerNew | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 883 | 28695 | sidereon_rtcm_lock_time_tracker_observe | positioning.RTCMLockTimeTrackerObserve | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 884 | 28708 | sidereon_rtcm_lock_time_tracker_reset | positioning.RTCMLockTimeTrackerReset | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 885 | 28717 | sidereon_rtcm_message_antenna_descriptor | positioning.RTCMMessageAntennaDescriptor | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 886 | 28730 | sidereon_rtcm_message_antenna_string | positioning.RTCMMessageAntennaString | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 887 | 28743 | sidereon_rtcm_message_beidou_ephemeris | positioning.RTCMMessageBeidouEphemeris | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 888 | 28755 | sidereon_rtcm_message_encode | positioning.RTCMMessageEncode | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 889 | 28768 | sidereon_rtcm_message_galileo_fnav_ephemeris | positioning.RTCMMessageGalileoFnavEphemeris | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 890 | 28778 | sidereon_rtcm_message_galileo_inav_ephemeris | positioning.RTCMMessageGalileoInavEphemeris | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 891 | 28788 | sidereon_rtcm_message_glonass_ephemeris | positioning.RTCMMessageGLONASSEphemeris | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 892 | 28797 | sidereon_rtcm_message_gps_ephemeris | positioning.RTCMMessageGPSEphemeris | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 893 | 28807 | sidereon_rtcm_message_kind | positioning.RTCMMessageKind | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 894 | 28818 | sidereon_rtcm_message_msm_info | positioning.RTCMMessageMsmInfo | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 895 | 28830 | sidereon_rtcm_message_msm_satellites | positioning.RTCMMessageMsmSatellites | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 896 | 28845 | sidereon_rtcm_message_msm_signals | positioning.RTCMMessageMsmSignals | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 897 | 28857 | sidereon_rtcm_message_qzss_ephemeris | positioning.RTCMMessageQzssEphemeris | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 898 | 28861 | sidereon_rtcm_message_ssr_clocks | positioning.RTCMMessageSSRClocks | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 899 | 28876 | sidereon_rtcm_message_ssr_code_bias_signals | distribution.RTCMMessageSSRCodeBiasSignals | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 900 | 28893 | sidereon_rtcm_message_ssr_code_biases | distribution.RTCMMessageSSRCodeBiases | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 901 | 28900 | sidereon_rtcm_message_ssr_info | positioning.RTCMMessageSSRInfo | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 902 | 28904 | sidereon_rtcm_message_ssr_orbits | positioning.RTCMMessageSSROrbits | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 903 | 28919 | sidereon_rtcm_message_ssr_phase_bias_signals | distribution.RTCMMessageSSRPhaseBiasSignals | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 904 | 28936 | sidereon_rtcm_message_ssr_phase_biases | distribution.RTCMMessageSSRPhaseBiases | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 905 | 28943 | sidereon_rtcm_message_ssr_ura | positioning.RTCMMessageSSRUra | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 906 | 28956 | sidereon_rtcm_message_station_coordinates | positioning.RTCMMessageStationCoordinates | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 907 | 28968 | sidereon_rtcm_message_to_frame | positioning.RTCMMessageToFrame | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 908 | 28980 | sidereon_rtcm_messages_count | positioning.RTCMMessagesCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 909 | 28988 | sidereon_rtcm_messages_free | positioning.RTCMMessages.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 910 | 28998 | sidereon_rtcm_minimum_lock_time_ms | positioning.RTCMMinimumLockTimeMs | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 911 | 29009 | sidereon_rtcm_msm_epoch_dt_ms | positioning.RTCMMsmEpochDtMs | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 912 | 29023 | sidereon_rtcm_msm_signal_rinex_code | positioning.RTCMMsmSignalRINEXCode | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 913 | 29039 | sidereon_rtcm_scan_frames | positioning.RTCMScanFrames | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 914 | 29049 | sidereon_rtcm_stream_diagnostics_free | positioning.RTCMStreamDiagnostics.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 915 | 29056 | sidereon_rtcm_stream_diagnostics_resync_bytes | positioning.RTCMStreamDiagnosticsResyncBytes | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 916 | 29065 | sidereon_rtcm_stream_diagnostics_skipped_frame | positioning.RTCMStreamDiagnosticsSkippedFrame | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 917 | 29076 | sidereon_rtcm_stream_diagnostics_skipped_frame_message | positioning.RTCMStreamDiagnosticsSkippedFrameMessage | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 918 | 29089 | sidereon_rtcm_stream_diagnostics_skipped_frames_count | positioning.RTCMStreamDiagnosticsSkippedFramesCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 919 | 29099 | sidereon_rtk_arc_solution_dropped_sats | positioning.RTKArcSolutionDroppedSats | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 920 | 29112 | sidereon_rtk_arc_solution_elevation_masked_sats | positioning.RTKArcSolutionElevationMaskedSats | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 921 | 29123 | sidereon_rtk_arc_solution_epoch_count | positioning.RTKArcSolutionEpochCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 922 | 29131 | sidereon_rtk_arc_solution_epoch_metadata | positioning.RTKArcSolutionEpochMetadata | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 923 | 29142 | sidereon_rtk_arc_solution_epoch_sd_ambiguities | positioning.RTKArcSolutionEpochSdAmbiguities | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 924 | 29156 | sidereon_rtk_arc_solution_epoch_string_ids | positioning.RTKArcSolutionEpochStringIds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 925 | 29171 | sidereon_rtk_arc_solution_epoch_used_satellites | positioning.RTKArcSolutionEpochUsedSatellites | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 926 | 29183 | sidereon_rtk_arc_solution_final_baseline | positioning.RTKArcSolutionFinalBaseline | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 927 | 29192 | sidereon_rtk_arc_solution_final_epoch_count | positioning.RTKArcSolutionFinalEpochCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 928 | 29200 | sidereon_rtk_arc_solution_free | positioning.RTKArcSolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 929 | 29209 | sidereon_rtk_arc_solution_measurement_covariance | positioning.RTKArcSolutionMeasurementCovariance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 930 | 29222 | sidereon_rtk_arc_solution_references | positioning.RTKArcSolutionReferences | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 931 | 29235 | sidereon_rtk_arc_solution_split_cycle_slip_arcs | positioning.RTKArcSolutionSplitCycleSlipArcs | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 932 | 29246 | sidereon_rtk_arc_update_options_init | positioning.RTKArcUpdateOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 933 | 29253 | sidereon_rtk_fixed_options_init | positioning.RTKFixedOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 934 | 29263 | sidereon_rtk_fixed_solution_fixed_ambiguities | positioning.RTKFixedSolutionFixedAmbiguities | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 935 | 29275 | sidereon_rtk_fixed_solution_fixed_baseline_ecef | positioning.RTKFixedSolutionFixedBaselineECEF | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 936 | 29286 | sidereon_rtk_fixed_solution_fixed_baseline_enu | positioning.RTKFixedSolutionFixedBaselineEnu | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 937 | 29296 | sidereon_rtk_fixed_solution_float_baseline_ecef | positioning.RTKFixedSolutionFloatBaselineECEF | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 938 | 29308 | sidereon_rtk_fixed_solution_free | positioning.RTKFixedSolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 939 | 29318 | sidereon_rtk_fixed_solution_free_ambiguities | positioning.RTKFixedSolutionFreeAmbiguities | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 940 | 29330 | sidereon_rtk_fixed_solution_metadata | positioning.RTKFixedSolutionMetadata | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 941 | 29341 | sidereon_rtk_fixed_solution_used_sat_ids | positioning.RTKFixedSolutionUsedSatIds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 942 | 29352 | sidereon_rtk_float_options_init | positioning.RTKFloatOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 943 | 29362 | sidereon_rtk_float_solution_ambiguities | positioning.RTKFloatSolutionAmbiguities | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 944 | 29374 | sidereon_rtk_float_solution_baseline_ecef | positioning.RTKFloatSolutionBaselineECEF | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 945 | 29385 | sidereon_rtk_float_solution_baseline_enu | positioning.RTKFloatSolutionBaselineEnu | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 946 | 29397 | sidereon_rtk_float_solution_free | positioning.RTKFloatSolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 947 | 29405 | sidereon_rtk_float_solution_metadata | positioning.RTKFloatSolutionMetadata | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 948 | 29416 | sidereon_rtk_float_solution_used_sat_ids | positioning.RTKFloatSolutionUsedSatIds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 949 | 29430 | sidereon_rtk_ionosphere_free_arc_solution_epoch_base_observations | positioning.RTKIonosphereFreeArcSolutionEpochBaseObservations | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 950 | 29445 | sidereon_rtk_ionosphere_free_arc_solution_epoch_base_satellite_positions | positioning.RTKIonosphereFreeArcSolutionEpochBaseSatellitePositions | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 951 | 29457 | sidereon_rtk_ionosphere_free_arc_solution_epoch_count | positioning.RTKIonosphereFreeArcSolutionEpochCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 952 | 29466 | sidereon_rtk_ionosphere_free_arc_solution_epoch_metadata | positioning.RTKIonosphereFreeArcSolutionEpochMetadata | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 953 | 29478 | sidereon_rtk_ionosphere_free_arc_solution_epoch_rover_observations | positioning.RTKIonosphereFreeArcSolutionEpochRoverObservations | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 954 | 29493 | sidereon_rtk_ionosphere_free_arc_solution_epoch_rover_satellite_positions | positioning.RTKIonosphereFreeArcSolutionEpochRoverSatellitePositions | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 955 | 29508 | sidereon_rtk_ionosphere_free_arc_solution_epoch_satellite_positions | positioning.RTKIonosphereFreeArcSolutionEpochSatellitePositions | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 956 | 29521 | sidereon_rtk_ionosphere_free_arc_solution_free | positioning.RTKIonosphereFreeArcSolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 957 | 29529 | sidereon_rtk_ionosphere_free_arc_solution_offsets_m | positioning.RTKIonosphereFreeArcSolutionOffsetsM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 958 | 29542 | sidereon_rtk_ionosphere_free_arc_solution_references | positioning.RTKIonosphereFreeArcSolutionReferences | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 959 | 29554 | sidereon_rtk_ionosphere_free_arc_solution_wavelengths_m | positioning.RTKIonosphereFreeArcSolutionWavelengthsM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 960 | 29565 | sidereon_rtk_measurement_model_init | positioning.RTKMeasurementModelInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 961 | 29572 | sidereon_rtk_residual_validation_options_init | positioning.RTKResidualValidationOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 962 | 29580 | sidereon_rtk_rinex_arc_epoch_base_observations | positioning.RTKRINEXArcEpochBaseObservations | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 963 | 29594 | sidereon_rtk_rinex_arc_epoch_base_satellite_positions | positioning.RTKRINEXArcEpochBaseSatellitePositions | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 964 | 29606 | sidereon_rtk_rinex_arc_epoch_count | positioning.RTKRINEXArcEpochCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 965 | 29615 | sidereon_rtk_rinex_arc_epoch_metadata | positioning.RTKRINEXArcEpochMetadata | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 966 | 29625 | sidereon_rtk_rinex_arc_epoch_rover_observations | positioning.RTKRINEXArcEpochRoverObservations | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 967 | 29639 | sidereon_rtk_rinex_arc_epoch_rover_satellite_positions | positioning.RTKRINEXArcEpochRoverSatellitePositions | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 968 | 29652 | sidereon_rtk_rinex_arc_epoch_satellite_positions | positioning.RTKRINEXArcEpochSatellitePositions | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 969 | 29665 | sidereon_rtk_rinex_arc_free | positioning.RTKRINEXArc.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 970 | 29674 | sidereon_rtk_rinex_arc_offsets_m | positioning.RTKRINEXArcOffsetsM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 971 | 29685 | sidereon_rtk_rinex_arc_options_init | positioning.RTKRINEXArcOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 972 | 29693 | sidereon_rtk_rinex_arc_skipped_epoch_count | positioning.RTKRINEXArcSkippedEpochCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 973 | 29703 | sidereon_rtk_rinex_arc_wavelengths_m | positioning.RTKRINEXArcWavelengthsM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 974 | 29714 | sidereon_rtk_rinex_dual_arc_options_init | positioning.RTKRINEXDualArcOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 975 | 29723 | sidereon_rtk_rinex_dual_frequency_arc_epoch_base_satellite_positions | positioning.RTKRINEXDualFrequencyArcEpochBaseSatellitePositions | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 976 | 29735 | sidereon_rtk_rinex_dual_frequency_arc_epoch_count | positioning.RTKRINEXDualFrequencyArcEpochCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 977 | 29745 | sidereon_rtk_rinex_dual_frequency_arc_epoch_metadata | positioning.RTKRINEXDualFrequencyArcEpochMetadata | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 978 | 29755 | sidereon_rtk_rinex_dual_frequency_arc_epoch_observations | positioning.RTKRINEXDualFrequencyArcEpochObservations | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 979 | 29769 | sidereon_rtk_rinex_dual_frequency_arc_epoch_rover_satellite_positions | positioning.RTKRINEXDualFrequencyArcEpochRoverSatellitePositions | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 980 | 29782 | sidereon_rtk_rinex_dual_frequency_arc_epoch_satellite_positions | positioning.RTKRINEXDualFrequencyArcEpochSatellitePositions | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 981 | 29796 | sidereon_rtk_rinex_dual_frequency_arc_epoch_sort_key | positioning.RTKRINEXDualFrequencyArcEpochSortKey | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 982 | 29809 | sidereon_rtk_rinex_dual_frequency_arc_free | positioning.RTKRINEXDualFrequencyArc.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 983 | 29816 | sidereon_rtk_rinex_dual_frequency_arc_skipped_epoch_count | positioning.RTKRINEXDualFrequencyArcSkippedEpochCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 984 | 29824 | sidereon_rtk_rinex_static_baseline_config_init | positioning.RTKRINEXStaticBaselineConfigInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 985 | 29831 | sidereon_rtk_rinex_wide_lane_fixed_config_init | positioning.RTKRINEXWideLaneFixedConfigInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 986 | 29839 | sidereon_rtk_static_arc_solution_ambiguity_ids | positioning.RTKStaticArcSolutionAmbiguityIds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 987 | 29853 | sidereon_rtk_static_arc_solution_ambiguity_satellites | positioning.RTKStaticArcSolutionAmbiguitySatellites | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 988 | 29866 | sidereon_rtk_static_arc_solution_dropped_sats | positioning.RTKStaticArcSolutionDroppedSats | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 989 | 29879 | sidereon_rtk_static_arc_solution_elevation_masked_sats | positioning.RTKStaticArcSolutionElevationMaskedSats | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 990 | 29892 | sidereon_rtk_static_arc_solution_fixed_ambiguities | positioning.RTKStaticArcSolutionFixedAmbiguities | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 991 | 29903 | sidereon_rtk_static_arc_solution_fixed_baseline_ecef | positioning.RTKStaticArcSolutionFixedBaselineECEF | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 992 | 29914 | sidereon_rtk_static_arc_solution_fixed_free_ambiguities | positioning.RTKStaticArcSolutionFixedFreeAmbiguities | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 993 | 29926 | sidereon_rtk_static_arc_solution_fixed_metadata | positioning.RTKStaticArcSolutionFixedMetadata | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 994 | 29936 | sidereon_rtk_static_arc_solution_float_ambiguities | positioning.RTKStaticArcSolutionFloatAmbiguities | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 995 | 29947 | sidereon_rtk_static_arc_solution_float_baseline_ecef | positioning.RTKStaticArcSolutionFloatBaselineECEF | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 996 | 29957 | sidereon_rtk_static_arc_solution_float_metadata | positioning.RTKStaticArcSolutionFloatMetadata | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 997 | 29965 | sidereon_rtk_static_arc_solution_free | positioning.RTKStaticArcSolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 998 | 29974 | sidereon_rtk_static_arc_solution_geometry_quality | positioning.RTKStaticArcSolutionGeometryQuality | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 999 | 29984 | sidereon_rtk_static_arc_solution_references | positioning.RTKStaticArcSolutionReferences | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1000 | 29996 | sidereon_rtk_static_arc_solution_split_cycle_slip_arcs | positioning.RTKStaticArcSolutionSplitCycleSlipArcs | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1001 | 30009 | sidereon_rtk_wide_lane_arc_solution_dropped_sats | positioning.RTKWideLaneArcSolutionDroppedSats | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1002 | 30020 | sidereon_rtk_wide_lane_arc_solution_epoch_count | positioning.RTKWideLaneArcSolutionEpochCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1003 | 30028 | sidereon_rtk_wide_lane_arc_solution_free | positioning.RTKWideLaneArcSolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1004 | 30037 | sidereon_rtk_wide_lane_arc_solution_geometry_quality | positioning.RTKWideLaneArcSolutionGeometryQuality | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1005 | 30047 | sidereon_rtk_wide_lane_arc_solution_references | positioning.RTKWideLaneArcSolutionReferences | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1006 | 30059 | sidereon_rtk_wide_lane_arc_solution_split_cycle_slip_arcs | positioning.RTKWideLaneArcSolutionSplitCycleSlipArcs | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1007 | 30072 | sidereon_rtk_wide_lane_arc_solution_wide_lane_cycles | positioning.RTKWideLaneArcSolutionWideLaneCycles | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1008 | 30083 | sidereon_rtk_wide_lane_fixed_rinex_solution_fixed_baseline_ecef | positioning.RTKWideLaneFixedRINEXSolutionFixedBaselineECEF | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1009 | 30093 | sidereon_rtk_wide_lane_fixed_rinex_solution_fixed_metadata | positioning.RTKWideLaneFixedRINEXSolutionFixedMetadata | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1010 | 30101 | sidereon_rtk_wide_lane_fixed_rinex_solution_float_baseline_ecef | positioning.RTKWideLaneFixedRINEXSolutionFloatBaselineECEF | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1011 | 30111 | sidereon_rtk_wide_lane_fixed_rinex_solution_float_metadata | positioning.RTKWideLaneFixedRINEXSolutionFloatMetadata | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1012 | 30121 | sidereon_rtk_wide_lane_fixed_rinex_solution_free | positioning.RTKWideLaneFixedRINEXSolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1013 | 30129 | sidereon_rtk_wide_lane_fixed_rinex_solution_metadata | positioning.RTKWideLaneFixedRINEXSolutionMetadata | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1014 | 30139 | sidereon_rtk_wide_lane_fixed_rinex_solution_wide_lane_cycles | positioning.RTKWideLaneFixedRINEXSolutionWideLaneCycles | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1015 | 30152 | sidereon_rtn_to_eci_covariance | errormetrics.RtnToECICovariance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1016 | 30166 | sidereon_rv2coe | support.Rv2coe | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1017 | 30171 | sidereon_rv2eq | support.Rv2eq | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1018 | 30177 | sidereon_rv2mee | support.Rv2mee | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1019 | 30196 | sidereon_satellite_constellation_build | astro.SatelliteConstellationBuild | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1020 | 30211 | sidereon_satellite_constellation_catalog_number | astro.SatelliteConstellationCatalogNumber | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1021 | 30227 | sidereon_satellite_constellation_free | astro.SatelliteConstellation.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1022 | 30240 | sidereon_satellite_constellation_ground_tracks | astro.SatelliteConstellationGroundTracks | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1023 | 30254 | sidereon_satellite_constellation_ground_tracks_free | astro.SatelliteConstellationGroundTracks.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1024 | 30262 | sidereon_satellite_constellation_ground_tracks_satellite_count | astro.SatelliteConstellationGroundTracksSatelliteCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1025 | 30273 | sidereon_satellite_constellation_ground_tracks_track_len | astro.SatelliteConstellationGroundTracksTrackLen | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1026 | 30288 | sidereon_satellite_constellation_ground_tracks_values | astro.SatelliteConstellationGroundTracksValues | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1027 | 30307 | sidereon_satellite_constellation_look_angle_arcs | astro.SatelliteConstellationLookAngleArcs | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1028 | 30322 | sidereon_satellite_constellation_look_angles_arc_len | astro.SatelliteConstellationLookAnglesArcLen | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1029 | 30335 | sidereon_satellite_constellation_look_angles_free | astro.SatelliteConstellationLookAngles.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1030 | 30343 | sidereon_satellite_constellation_look_angles_satellite_count | astro.SatelliteConstellationLookAnglesSatelliteCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1031 | 30357 | sidereon_satellite_constellation_look_angles_values | astro.SatelliteConstellationLookAnglesValues | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1032 | 30374 | sidereon_satellite_constellation_passes | astro.SatelliteConstellationPasses | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1033 | 30386 | sidereon_satellite_constellation_passes_count | astro.SatelliteConstellationPassesCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1034 | 30398 | sidereon_satellite_constellation_passes_free | astro.SatelliteConstellationPasses.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1035 | 30408 | sidereon_satellite_constellation_passes_values | astro.SatelliteConstellationPassesValues | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1036 | 30428 | sidereon_satellite_constellation_propagate | astro.SatelliteConstellationPropagate | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1037 | 30440 | sidereon_satellite_constellation_satellite_count | astro.SatelliteConstellationSatelliteCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1038 | 30455 | sidereon_satellite_constellation_visible | astro.SatelliteConstellationVisible | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1039 | 30467 | sidereon_satellite_visual_magnitude | astro.SatelliteVisualMagnitude | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1040 | 30478 | sidereon_sbas_airborne_model_aad_a | positioning.SBASAirborneModelAadA | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1041 | 30485 | sidereon_sbas_airborne_sigma_air_m | positioning.SBASAirborneSigmaAirM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1042 | 30489 | sidereon_sbas_block_decode | positioning.SBASBlockDecode | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1043 | 30494 | sidereon_sbas_block_encode | positioning.SBASBlockEncode | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1044 | 30507 | sidereon_sbas_block_fast_corrections | positioning.SBASBlockFastCorrections | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1045 | 30518 | sidereon_sbas_block_fast_degradation | positioning.SBASBlockFastDegradation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1046 | 30522 | sidereon_sbas_block_free | positioning.SBASBlock.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1047 | 30531 | sidereon_sbas_block_geo_nav | positioning.SBASBlockGeoNav | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1048 | 30542 | sidereon_sbas_block_igp_mask | positioning.SBASBlockIgpMask | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1049 | 30546 | sidereon_sbas_block_info | positioning.SBASBlockInfo | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1050 | 30556 | sidereon_sbas_block_integrity | positioning.SBASBlockIntegrity | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1051 | 30567 | sidereon_sbas_block_iono_delays | positioning.SBASBlockIonoDelays | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1052 | 30579 | sidereon_sbas_block_long_term_half_info | positioning.SBASBlockLongTermHalfInfo | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1053 | 30592 | sidereon_sbas_block_long_term_records | positioning.SBASBlockLongTermRecords | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1054 | 30607 | sidereon_sbas_block_mixed_fast_corrections | positioning.SBASBlockMixedFastCorrections | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1055 | 30618 | sidereon_sbas_block_prn_mask | positioning.SBASBlockPrnMask | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1056 | 30630 | sidereon_sbas_block_raw_data | positioning.SBASBlockRawData | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1057 | 30636 | sidereon_sbas_corrected_state | positioning.SBASCorrectedState | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1058 | 30651 | sidereon_sbas_degradation_params_none | positioning.SBASDegradationParamsNone | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1059 | 30658 | sidereon_sbas_k_multipliers_en_route_npa | positioning.SBASKMultipliersEnRouteNpa | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1060 | 30665 | sidereon_sbas_k_multipliers_precision_approach | positioning.SBASKMultipliersPrecisionApproach | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1061 | 30673 | sidereon_sbas_log_blocks_bytes | distribution.SBASLogBlocksBytes | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1062 | 30685 | sidereon_sbas_log_blocks_count | distribution.SBASLogBlocksCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1063 | 30693 | sidereon_sbas_log_blocks_free | distribution.SBASLogBlocks.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1064 | 30701 | sidereon_sbas_log_blocks_item | distribution.SBASLogBlocksItem | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1065 | 30715 | sidereon_sbas_prn_to_satellite_id | distribution.SBASPRNToSatelliteID | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1066 | 30727 | sidereon_sbas_protection_levels | positioning.SBASProtectionLevels | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1067 | 30738 | sidereon_sbas_sigma_flt_m_for_udrei | positioning.SBASSigmaFltMForUdrei | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1068 | 30747 | sidereon_sbas_sigma_tropo_m | positioning.SBASSigmaTropoM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1069 | 30754 | sidereon_sbas_sis_error_sigma_m | positioning.SBASSisErrorSigmaM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1070 | 30757 | sidereon_sbas_solve_broadcast | positioning.SBASSolveBroadcast | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1071 | 30764 | sidereon_sbas_store_fast_correction | positioning.SBASStoreFastCorrection | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1072 | 30770 | sidereon_sbas_store_free | positioning.SBASStore.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1073 | 30772 | sidereon_sbas_store_geo_nav | positioning.SBASStoreGeoNav | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1074 | 30777 | sidereon_sbas_store_ingest | positioning.SBASStoreIngest | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1075 | 30782 | sidereon_sbas_store_iono_grid_igps | positioning.SBASStoreIonoGridIgps | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1076 | 30789 | sidereon_sbas_store_iono_slant_delay_m | positioning.SBASStoreIonoSlantDelayM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1077 | 30798 | sidereon_sbas_store_long_term_correction | positioning.SBASStoreLongTermCorrection | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1078 | 30804 | sidereon_sbas_store_new | positioning.SBASStoreNew | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1079 | 30806 | sidereon_sbas_store_preferred_geo | positioning.SBASStorePreferredGeo | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1080 | 30811 | sidereon_sbas_store_ready_geos | positioning.SBASStoreReadyGeos | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1081 | 30825 | sidereon_scenario_epoch_offsets | distribution.ScenarioEpochOffsets | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1082 | 30837 | sidereon_scenario_observations | distribution.ScenarioObservations | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1083 | 30849 | sidereon_scenario_receiver_truth | distribution.ScenarioReceiverTruth | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1084 | 30861 | sidereon_scenario_simulate_json | distribution.ScenarioSimulateJSON | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1085 | 30873 | sidereon_scenario_simulate_json_with_broadcast | positioning.ScenarioSimulateJSONWithBroadcast | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1086 | 30886 | sidereon_scenario_simulate_json_with_broadcast_and_ionex | positioning.ScenarioSimulateJSONWithBroadcastAndIonex | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1087 | 30900 | sidereon_scenario_simulate_json_with_ionex | positioning.ScenarioSimulateJSONWithIonex | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1088 | 30914 | sidereon_scenario_simulate_json_with_sp3 | positioning.ScenarioSimulateJSONWithSP3 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1089 | 30927 | sidereon_scenario_simulate_json_with_sp3_and_ionex | positioning.ScenarioSimulateJSONWithSP3AndIonex | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1090 | 30940 | sidereon_scenario_simulation_free | distribution.ScenarioSimulation.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1091 | 30948 | sidereon_scenario_simulation_json | distribution.ScenarioSimulationJSON | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1092 | 30960 | sidereon_scenario_simulation_summary | distribution.ScenarioSimulationSummary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1093 | 30969 | sidereon_scenario_terms | distribution.ScenarioTerms | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1094 | 30985 | sidereon_screen_tca_candidates_from_tle_catalog | astro.ScreenTCACandidatesFromTLECatalog | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1095 | 31011 | sidereon_screen_tca_conjunctions_from_tle_catalog | astro.ScreenTCAConjunctionsFromTLECatalog | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1096 | 31034 | sidereon_second_of_day | geodesy.SecondOfDay | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1097 | 31044 | sidereon_select_ionex | positioning.SelectIonex | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1098 | 31064 | sidereon_select_ionex_over_range | positioning.SelectIonexOverRange | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1099 | 31077 | sidereon_select_sp3 | positioning.SelectSP3 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1100 | 31097 | sidereon_select_sp3_over_range | positioning.SelectSP3OverRange | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1101 | 31110 | sidereon_sgp4_decay_latch_clear | astro.SGP4DecayLatchClear | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1102 | 31118 | sidereon_sgp4_decay_latch_first_failing_epoch | astro.SGP4DecayLatchFirstFailingEpoch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1103 | 31128 | sidereon_sgp4_decay_latch_free | astro.SGP4DecayLatch.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1104 | 31135 | sidereon_sgp4_decay_latch_new | astro.SGP4DecayLatchNew | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1105 | 31137 | sidereon_sgp4_fit_config_init | astro.SGP4FitConfigInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1106 | 31139 | sidereon_sgp4_fit_tle | astro.SGP4FitTLE | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1107 | 31144 | sidereon_sgp4_tle_fit_free | astro.SGP4TLEFit.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1108 | 31146 | sidereon_sgp4_tle_fit_lines | astro.SGP4TLEFitLines | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1109 | 31149 | sidereon_sgp4_tle_fit_omm | astro.SGP4TLEFitOMM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1110 | 31152 | sidereon_sgp4_tle_fit_statistics | astro.SGP4TLEFitStatistics | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1111 | 31163 | sidereon_sidereal_filter | astro.SiderealFilter | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1112 | 31174 | sidereon_sidereal_filter_options_init | astro.SiderealFilterOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1113 | 31181 | sidereon_sidereal_filter_output_coverage | astro.SiderealFilterOutputCoverage | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1114 | 31192 | sidereon_sidereal_filter_output_filtered | astro.SiderealFilterOutputFiltered | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1115 | 31203 | sidereon_sidereal_filter_output_free | astro.SiderealFilterOutput.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1116 | 31211 | sidereon_sidereal_filter_output_template | astro.SiderealFilterOutputTemplate | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1117 | 31222 | sidereon_sidereal_filter_output_under_covered | astro.SiderealFilterOutputUnderCovered | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1118 | 31234 | sidereon_sidereal_orbit_repeat_lag | astro.SiderealOrbitRepeatLag | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1119 | 31246 | sidereon_sidereal_periodicity_strength | astro.SiderealPeriodicityStrength | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1120 | 31260 | sidereon_sidereal_repeat_period | astro.SiderealRepeatPeriod | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1121 | 31272 | sidereon_sigmas | errormetrics.Sigmas | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1122 | 31289 | sidereon_signal_acquire | positioning.SignalAcquire | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1123 | 31305 | sidereon_signal_analysis_dll_jitter | positioning.SignalAnalysisDllJitter | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1124 | 31316 | sidereon_signal_analysis_dll_lower_bound | positioning.SignalAnalysisDllLowerBound | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1125 | 31327 | sidereon_signal_analysis_effective_cn0_degradation | positioning.SignalAnalysisEffectiveCn0Degradation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1126 | 31339 | sidereon_signal_analysis_fraction_power | positioning.SignalAnalysisFractionPower | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1127 | 31351 | sidereon_signal_analysis_multipath_envelope | positioning.SignalAnalysisMultipathEnvelope | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1128 | 31365 | sidereon_signal_analysis_psd | positioning.SignalAnalysisPsd | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1129 | 31374 | sidereon_signal_analysis_rms_bandwidth_hz | positioning.SignalAnalysisRmsBandwidthHz | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1130 | 31384 | sidereon_signal_analysis_spectral_separation | positioning.SignalAnalysisSpectralSeparation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1131 | 31397 | sidereon_signal_autocorrelation | positioning.SignalAutocorrelation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1132 | 31409 | sidereon_signal_betz_l1_receiver_bandwidth_hz | positioning.SignalBetzL1ReceiverBandwidthHz | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1133 | 31417 | sidereon_signal_ca_chip | positioning.SignalCAChip | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1134 | 31426 | sidereon_signal_ca_code | positioning.SignalCACode | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1135 | 31438 | sidereon_signal_coherent_loss | positioning.SignalCoherentLoss | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1136 | 31448 | sidereon_signal_coherent_loss_db | positioning.SignalCoherentLossDb | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1137 | 31459 | sidereon_signal_correlate | positioning.SignalCorrelate | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1138 | 31473 | sidereon_signal_correlate_against | positioning.SignalCorrelateAgainst | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1139 | 31489 | sidereon_signal_correlation_at | positioning.SignalCorrelationAt | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1140 | 31503 | sidereon_signal_cross_correlation | positioning.SignalCrossCorrelation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1141 | 31518 | sidereon_signal_dll_lower_bound | positioning.SignalDllLowerBound | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1142 | 31529 | sidereon_signal_dll_thermal_noise_jitter | positioning.SignalDllThermalNoiseJitter | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1143 | 31542 | sidereon_signal_effective_cn0_degradation | positioning.SignalEffectiveCn0Degradation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1144 | 31554 | sidereon_signal_fraction_power_in_band | positioning.SignalFractionPowerInBand | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1145 | 31563 | sidereon_signal_modulation_code_rate_hz | positioning.SignalModulationCodeRateHz | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1146 | 31573 | sidereon_signal_modulation_label | positioning.SignalModulationLabel | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1147 | 31588 | sidereon_signal_multipath_error_envelope | positioning.SignalMultipathErrorEnvelope | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1148 | 31602 | sidereon_signal_power_in_band | positioning.SignalPowerInBand | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1149 | 31612 | sidereon_signal_psd_hz | positioning.SignalPsdHz | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1150 | 31621 | sidereon_signal_reference_chip_rate_hz | positioning.SignalReferenceChipRateHz | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1151 | 31630 | sidereon_signal_replica | positioning.SignalReplica | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1152 | 31642 | sidereon_signal_rms_bandwidth_hz | positioning.SignalRmsBandwidthHz | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1153 | 31652 | sidereon_signal_snr_post_db | positioning.SignalSnrPostDb | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1154 | 31663 | sidereon_signal_spectral_separation_coefficient_db_hz | positioning.SignalSpectralSeparationCoefficientDbHz | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1155 | 31676 | sidereon_signal_spectral_separation_coefficient_hz | positioning.SignalSpectralSeparationCoefficientHz | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1156 | 31687 | sidereon_signal_white_noise_spectral_separation_hz | positioning.SignalWhiteNoiseSpectralSeparationHz | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1157 | 31700 | sidereon_smooth_code | support.SmoothCode | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1158 | 31715 | sidereon_smooth_fusion_rts | errormetrics.SmoothFusionRts | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1159 | 31727 | sidereon_smooth_iono_free_code | support.SmoothIonoFreeCode | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1160 | 31741 | sidereon_smooth_track_rts | errormetrics.SmoothTrackRts | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1161 | 31750 | sidereon_smoothed_fusion_trajectory_epoch | errormetrics.SmoothedFusionTrajectoryEpoch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1162 | 31759 | sidereon_smoothed_fusion_trajectory_epoch_count | errormetrics.SmoothedFusionTrajectoryEpochCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1163 | 31768 | sidereon_smoothed_fusion_trajectory_epoch_covariance | errormetrics.SmoothedFusionTrajectoryEpochCovariance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1164 | 31781 | sidereon_smoothed_fusion_trajectory_epoch_error_state_correction | errormetrics.SmoothedFusionTrajectoryEpochErrorStateCorrection | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1165 | 31794 | sidereon_smoothed_fusion_trajectory_epoch_position_ecef_m | errormetrics.SmoothedFusionTrajectoryEpochPositionECEFM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1166 | 31807 | sidereon_smoothed_fusion_trajectory_epoch_rts_gain_to_next | errormetrics.SmoothedFusionTrajectoryEpochRtsGainToNext | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1167 | 31820 | sidereon_smoothed_fusion_trajectory_free | errormetrics.SmoothedFusionTrajectory.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1168 | 31827 | sidereon_smoothed_track_epoch | errormetrics.SmoothedTrackEpoch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1169 | 31836 | sidereon_smoothed_track_epoch_count | errormetrics.SmoothedTrackEpochCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1170 | 31845 | sidereon_smoothed_track_epoch_covariance | errormetrics.SmoothedTrackEpochCovariance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1171 | 31858 | sidereon_smoothed_track_epoch_position_m | errormetrics.SmoothedTrackEpochPositionM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1172 | 31872 | sidereon_smoothed_track_epoch_rts_gain_to_next | errormetrics.SmoothedTrackEpochRtsGainToNext | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1173 | 31884 | sidereon_smoothed_track_free | errormetrics.SmoothedTrack.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1174 | 31892 | sidereon_solid_earth_pole_tide | geodesy.SolidEarthPoleTide | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1175 | 31908 | sidereon_solid_earth_tide | geodesy.SolidEarthTide | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1176 | 31922 | sidereon_solution_validation_options_init | support.SolutionValidationOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1177 | 31937 | sidereon_solve_broadcast | positioning.SolveBroadcast | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1178 | 31949 | sidereon_solve_broadcast_with_doppler_velocity | positioning.SolveBroadcastWithDopplerVelocity | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1179 | 31967 | sidereon_solve_data_problem | positioning.SolveDataProblem | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1180 | 31981 | sidereon_solve_data_problem_drop_one | positioning.SolveDataProblemDropOne | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1181 | 31984 | sidereon_solve_kepler | positioning.SolveKepler | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1182 | 31998 | sidereon_solve_moving_baseline | positioning.SolveMovingBaseline | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1183 | 32015 | sidereon_solve_ppp_auto_init_fixed | positioning.SolvePPPAutoInitFixed | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1184 | 32032 | sidereon_solve_ppp_auto_init_float | positioning.SolvePPPAutoInitFloat | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1185 | 32045 | sidereon_solve_ppp_fixed | positioning.SolvePPPFixed | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1186 | 32058 | sidereon_solve_ppp_float | positioning.SolvePPPFloat | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1187 | 32072 | sidereon_solve_rtk_arc | positioning.SolveRTKArc | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1188 | 32085 | sidereon_solve_rtk_fixed | positioning.SolveRTKFixed | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1189 | 32096 | sidereon_solve_rtk_float | positioning.SolveRTKFloat | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1190 | 32108 | sidereon_solve_spp | positioning.SolveSPP | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1191 | 32122 | sidereon_solve_spp_batch_parallel | positioning.SolveSPPBatchParallel | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1192 | 32138 | sidereon_solve_spp_batch_serial | positioning.SolveSPPBatchSerial | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1193 | 32156 | sidereon_solve_spp_from_rinex_obs | positioning.SolveSPPFromRINEXObs | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1194 | 32174 | sidereon_solve_spp_v2 | positioning.SolveSPPV2 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1195 | 32186 | sidereon_solve_spp_with_doppler_velocity | positioning.SolveSPPWithDopplerVelocity | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1196 | 32199 | sidereon_solve_static_position_broadcast | positioning.SolveStaticPositionBroadcast | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1197 | 32213 | sidereon_solve_static_position_sp3 | positioning.SolveStaticPositionSP3 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1198 | 32229 | sidereon_solve_static_reference_station_rinex | positioning.SolveStaticReferenceStationRINEX | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1199 | 32243 | sidereon_solve_static_rinex_rtk_baseline | positioning.SolveStaticRINEXRTKBaseline | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1200 | 32259 | sidereon_solve_static_rtk_arc | positioning.SolveStaticRTKArc | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1201 | 32281 | sidereon_solve_velocity | positioning.SolveVelocity | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1202 | 32303 | sidereon_solve_velocity_broadcast | positioning.SolveVelocityBroadcast | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1203 | 32319 | sidereon_solve_wide_lane_fixed_rinex_rtk_baseline | positioning.SolveWideLaneFixedRINEXRTKBaseline | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1204 | 32337 | sidereon_solve_with_fallback | positioning.SolveWithFallback | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1205 | 32350 | sidereon_source_crlb | support.SourceCrlb | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1206 | 32365 | sidereon_source_dop | errormetrics.SourceDOP | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1207 | 32378 | sidereon_source_locate_options_init | support.SourceLocateOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1208 | 32386 | sidereon_source_solution_covariance | errormetrics.SourceSolutionCovariance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1209 | 32396 | sidereon_source_solution_free | support.SourceSolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1210 | 32405 | sidereon_source_solution_influences | support.SourceSolutionInfluences | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1211 | 32417 | sidereon_source_solution_residuals | support.SourceSolutionResiduals | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1212 | 32428 | sidereon_source_solution_summary | support.SourceSolutionSummary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1213 | 32444 | sidereon_sourced_solution_broadcast_reason | positioning.SourcedSolutionBroadcastReason | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1214 | 32457 | sidereon_sourced_solution_free | distribution.SourcedSolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1215 | 32465 | sidereon_sourced_solution_is_precise_exact | positioning.SourcedSolutionIsPreciseExact | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1216 | 32476 | sidereon_sourced_solution_solution | distribution.SourcedSolutionSolution | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1217 | 32485 | sidereon_sourced_solution_source_kind | distribution.SourcedSolutionSourceKind | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1218 | 32499 | sidereon_sourced_solution_staleness | distribution.SourcedSolutionStaleness | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1219 | 32510 | sidereon_sp3_align_clock_reference | positioning.SP3AlignClockReference | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1220 | 32534 | sidereon_sp3_check_continuity | positioning.SP3CheckContinuity | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1221 | 32551 | sidereon_sp3_clock_reference_offsets | positioning.SP3ClockReferenceOffsets | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1222 | 32576 | sidereon_sp3_continuity_verdict_json | positioning.SP3ContinuityVerdictJSON | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1223 | 32596 | sidereon_sp3_declared_epoch_count | positioning.SP3DeclaredEpochCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1224 | 32611 | sidereon_sp3_declared_start_j2000_seconds | positioning.SP3DeclaredStartJ2000Seconds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1225 | 32624 | sidereon_sp3_emission_media_batch_at_j2000_s | positioning.SP3EmissionMediaBatchAtJ2000S | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1226 | 32648 | sidereon_sp3_ephemeris_sample | positioning.SP3EphemerisSample | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1227 | 32664 | sidereon_sp3_epoch_count | positioning.SP3EpochCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1228 | 32672 | sidereon_sp3_epoch_prediction | positioning.SP3EpochPrediction | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1229 | 32685 | sidereon_sp3_epochs_j2000_seconds | positioning.SP3EpochsJ2000Seconds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1230 | 32694 | sidereon_sp3_exact_request_free | positioning.SP3ExactRequest.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1231 | 32707 | sidereon_sp3_exact_request_from_identity | positioning.SP3ExactRequestFromIdentity | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1232 | 32724 | sidereon_sp3_exact_request_new | positioning.SP3ExactRequestNew | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1233 | 32742 | sidereon_sp3_free | positioning.SP3.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1234 | 32754 | sidereon_sp3_geometry_passes | positioning.SP3GeometryPasses | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1235 | 32777 | sidereon_sp3_geometry_visibility_series | positioning.SP3GeometryVisibilitySeries | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1236 | 32804 | sidereon_sp3_geometry_visible | positioning.SP3GeometryVisible | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1237 | 32824 | sidereon_sp3_interpolate | positioning.SP3Interpolate | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1238 | 32842 | sidereon_sp3_load | positioning.SP3Load | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1239 | 32859 | sidereon_sp3_load_exact | positioning.SP3LoadExact | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1240 | 32874 | sidereon_sp3_merge | positioning.SP3Merge | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1241 | 32897 | sidereon_sp3_merge_input_identity | positioning.SP3MergeInputIdentity | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1242 | 32905 | sidereon_sp3_merge_input_identity_contributor | positioning.SP3MergeInputIdentityContributor | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1243 | 32912 | sidereon_sp3_merge_input_identity_contributor_count | positioning.SP3MergeInputIdentityContributorCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1244 | 32918 | sidereon_sp3_merge_input_identity_free | positioning.SP3MergeInputIdentity.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1245 | 32923 | sidereon_sp3_merge_input_identity_precedence_contributor | positioning.SP3MergeInputIdentityPrecedenceContributor | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1246 | 32931 | sidereon_sp3_merge_input_identity_precedence_contributor_count | positioning.SP3MergeInputIdentityPrecedenceContributorCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1247 | 32938 | sidereon_sp3_merge_input_identity_schema_version | positioning.SP3MergeInputIdentitySchemaVersion | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1248 | 32945 | sidereon_sp3_merge_input_identity_stable_id | positioning.SP3MergeInputIdentityStableId | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1249 | 32956 | sidereon_sp3_merge_options_init | positioning.SP3MergeOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1250 | 32967 | sidereon_sp3_merge_report_agreement_summary | positioning.SP3MergeReportAgreementSummary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1251 | 32987 | sidereon_sp3_merge_report_continuity_verdict_json | positioning.SP3MergeReportContinuityVerdictJSON | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1252 | 33004 | sidereon_sp3_merge_report_epoch_agreement | positioning.SP3MergeReportEpochAgreement | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1253 | 33018 | sidereon_sp3_merge_report_epoch_agreement_count | positioning.SP3MergeReportEpochAgreementCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1254 | 33028 | sidereon_sp3_merge_report_flag | positioning.SP3MergeReportFlag | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1255 | 33040 | sidereon_sp3_merge_report_flag_count | positioning.SP3MergeReportFlagCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1256 | 33053 | sidereon_sp3_merge_report_flag_sources | positioning.SP3MergeReportFlagSources | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1257 | 33067 | sidereon_sp3_merge_report_frame_reconciliation | positioning.SP3MergeReportFrameReconciliation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1258 | 33077 | sidereon_sp3_merge_report_frame_reconciliation_asserted_label | positioning.SP3MergeReportFrameReconciliationAssertedLabel | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1259 | 33091 | sidereon_sp3_merge_report_frame_reconciliation_count | positioning.SP3MergeReportFrameReconciliationCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1260 | 33100 | sidereon_sp3_merge_report_frame_reconciliation_provenance | positioning.SP3MergeReportFrameReconciliationProvenance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1261 | 33113 | sidereon_sp3_merge_report_frame_reconciliation_source_label | positioning.SP3MergeReportFrameReconciliationSourceLabel | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1262 | 33126 | sidereon_sp3_merge_report_frame_reconciliation_target_label | positioning.SP3MergeReportFrameReconciliationTargetLabel | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1263 | 33141 | sidereon_sp3_merge_report_free | positioning.SP3MergeReport.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1264 | 33150 | sidereon_sp3_observable_state | positioning.SP3ObservableState | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1265 | 33171 | sidereon_sp3_observable_states_at_j2000_s | positioning.SP3ObservableStatesAtJ2000S | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1266 | 33187 | sidereon_sp3_observable_states_at_shared_j2000_s | positioning.SP3ObservableStatesAtSharedJ2000S | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1267 | 33207 | sidereon_sp3_observables | positioning.SP3Observables | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1268 | 33228 | sidereon_sp3_observables_batch | positioning.SP3ObservablesBatch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1269 | 33246 | sidereon_sp3_precise_ephemeris_samples | positioning.SP3PreciseEphemerisSamples | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1270 | 33260 | sidereon_sp3_precise_interpolant_artifact_bytes | positioning.SP3PreciseInterpolantArtifactBytes | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1271 | 33278 | sidereon_sp3_predict_ranges | positioning.SP3PredictRanges | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1272 | 33291 | sidereon_sp3_prediction_summary | positioning.SP3PredictionSummary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1273 | 33302 | sidereon_sp3_satellites | positioning.SP3Satellites | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1274 | 33316 | sidereon_sp3_state | positioning.SP3State | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1275 | 33331 | sidereon_sp3_stencil_extent | positioning.SP3StencilExtent | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1276 | 33344 | sidereon_sp3_to_sp3_text | positioning.SP3ToSP3Text | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1277 | 33359 | sidereon_sp3_validate_exact | positioning.SP3ValidateExact | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1278 | 33368 | sidereon_space_weather_default | distribution.SpaceWeatherDefault | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1279 | 33370 | sidereon_space_weather_table_ap_array_at | distribution.SpaceWeatherTableApArrayAt | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1280 | 33374 | sidereon_space_weather_table_coverage | distribution.SpaceWeatherTableCoverage | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1281 | 33377 | sidereon_space_weather_table_day | distribution.SpaceWeatherTableDay | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1282 | 33384 | sidereon_space_weather_table_days | distribution.SpaceWeatherTableDays | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1283 | 33390 | sidereon_space_weather_table_free | distribution.SpaceWeatherTable.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1284 | 33392 | sidereon_space_weather_table_monthly | distribution.SpaceWeatherTableMonthly | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1285 | 33398 | sidereon_space_weather_table_parse | distribution.SpaceWeatherTableParse | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1286 | 33402 | sidereon_space_weather_table_parse_csv | distribution.SpaceWeatherTableParseCsv | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1287 | 33406 | sidereon_space_weather_table_parse_txt | distribution.SpaceWeatherTableParseTxt | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1288 | 33410 | sidereon_space_weather_table_sample_at | distribution.SpaceWeatherTableSampleAt | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1289 | 33414 | sidereon_space_weather_table_sample_at_with_policy | distribution.SpaceWeatherTableSampleAtWithPolicy | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1290 | 33419 | sidereon_space_weather_table_space_weather_at | distribution.SpaceWeatherTableSpaceWeatherAt | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1291 | 33423 | sidereon_space_weather_table_summary | distribution.SpaceWeatherTableSummary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1292 | 33426 | sidereon_space_weather_table_to_csv | distribution.SpaceWeatherTableToCsv | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1293 | 33432 | sidereon_space_weather_table_to_txt | distribution.SpaceWeatherTableToTxt | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1294 | 33445 | sidereon_spk_free | astro.SPK.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1295 | 33455 | sidereon_spk_load | astro.SPKLoad | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1296 | 33469 | sidereon_spk_state | astro.SPKState | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1297 | 33481 | sidereon_split_jd_to_j2000_seconds | geodesy.SplitJdToJ2000Seconds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1298 | 33490 | sidereon_spp_batch_count | positioning.SPPBatchCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1299 | 33498 | sidereon_spp_batch_epoch_ok | positioning.SPPBatchEpochOk | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1300 | 33510 | sidereon_spp_batch_error | positioning.SPPBatchError | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1301 | 33522 | sidereon_spp_batch_free | positioning.SPPBatch.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1302 | 33533 | sidereon_spp_batch_solution | positioning.SPPBatchSolution | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1303 | 33543 | sidereon_spp_doppler_solution_free | positioning.SPPDopplerSolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1304 | 33551 | sidereon_spp_doppler_solution_has_velocity | positioning.SPPDopplerSolutionHasVelocity | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1305 | 33561 | sidereon_spp_doppler_solution_receiver | positioning.SPPDopplerSolutionReceiver | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1306 | 33571 | sidereon_spp_doppler_solution_velocity | positioning.SPPDopplerSolutionVelocity | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1307 | 33581 | sidereon_spp_doppler_solution_velocity_error_kind | positioning.SPPDopplerSolutionVelocityErrorKind | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1308 | 33594 | sidereon_spp_inputs_from_rinex_obs | positioning.SPPInputsFromRINEXObs | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1309 | 33606 | sidereon_spp_inputs_v2_init | positioning.SPPInputsV2Init | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1310 | 33616 | sidereon_spp_solution_dop | positioning.SPPSolutionDOP | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1311 | 33628 | sidereon_spp_solution_free | positioning.SPPSolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1312 | 33638 | sidereon_spp_solution_geodetic | positioning.SPPSolutionGeodetic | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1313 | 33648 | sidereon_spp_solution_metadata | positioning.SPPSolutionMetadata | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1314 | 33658 | sidereon_spp_solution_position | positioning.SPPSolutionPosition | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1315 | 33668 | sidereon_spp_solution_position_covariance_ecef_m2 | positioning.SPPSolutionPositionCovarianceECEFM2 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1316 | 33678 | sidereon_spp_solution_position_covariance_enu_m2 | positioning.SPPSolutionPositionCovarianceEnuM2 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1317 | 33690 | sidereon_spp_solution_rejected_sats | positioning.SPPSolutionRejectedSats | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1318 | 33709 | sidereon_spp_solution_residuals | positioning.SPPSolutionResiduals | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1319 | 33722 | sidereon_spp_solution_rx_clock_drift_s_s | positioning.SPPSolutionRxClockDriftSS | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1320 | 33732 | sidereon_spp_solution_rx_clock_s | positioning.SPPSolutionRxClockS | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1321 | 33743 | sidereon_spp_solution_system_clocks | positioning.SPPSolutionSystemClocks | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1322 | 33760 | sidereon_spp_solution_system_tdops | positioning.SPPSolutionSystemTdops | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1323 | 33772 | sidereon_spp_solution_used_sat_count | positioning.SPPSolutionUsedSatCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1324 | 33783 | sidereon_spp_solution_used_sat_ids | positioning.SPPSolutionUsedSatIds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1325 | 33789 | sidereon_ssr_corrected_state | positioning.SSRCorrectedState | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1326 | 33801 | sidereon_ssr_ephemeris_sample | positioning.SSREphemerisSample | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1327 | 33825 | sidereon_ssr_message_clocks | distribution.SSRMessageClocks | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1328 | 33839 | sidereon_ssr_message_code_bias_signals | distribution.SSRMessageCodeBiasSignals | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1329 | 33855 | sidereon_ssr_message_code_biases | distribution.SSRMessageCodeBiases | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1330 | 33870 | sidereon_ssr_message_decode | distribution.SSRMessageDecode | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1331 | 33880 | sidereon_ssr_message_free | distribution.SSRMessage.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1332 | 33889 | sidereon_ssr_message_info | distribution.SSRMessageInfo | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1333 | 33900 | sidereon_ssr_message_orbits | distribution.SSRMessageOrbits | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1334 | 33914 | sidereon_ssr_message_phase_bias_signals | distribution.SSRMessagePhaseBiasSignals | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1335 | 33930 | sidereon_ssr_message_phase_biases | distribution.SSRMessagePhaseBiases | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1336 | 33944 | sidereon_ssr_message_ura | distribution.SSRMessageURA | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1337 | 33950 | sidereon_ssr_solve_broadcast | positioning.SSRSolveBroadcast | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1338 | 33959 | sidereon_ssr_store_clock | positioning.SSRStoreClock | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1339 | 33964 | sidereon_ssr_store_code_bias_m | positioning.SSRStoreCodeBiasM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1340 | 33970 | sidereon_ssr_store_free | positioning.SSRStore.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1341 | 33972 | sidereon_ssr_store_from_rtcm | positioning.SSRStoreFromRTCM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1342 | 33977 | sidereon_ssr_store_ingest_messages | positioning.SSRStoreIngestMessages | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1343 | 33981 | sidereon_ssr_store_new | positioning.SSRStoreNew | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1344 | 33984 | sidereon_ssr_store_orbit | positioning.SSRStoreOrbit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1345 | 33989 | sidereon_ssr_store_phase_bias_m | positioning.SSRStorePhaseBiasM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1346 | 33995 | sidereon_ssr_store_ura_index | positioning.SSRStoreUraIndex | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1347 | 34003 | sidereon_staleness_policy_days | distribution.StalenessPolicyDays | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1348 | 34008 | sidereon_staleness_policy_default | distribution.StalenessPolicyDefault | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1349 | 34013 | sidereon_staleness_policy_seconds | distribution.StalenessPolicySeconds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1350 | 34020 | sidereon_state_propagation_config_init | support.StatePropagationConfigInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1351 | 34027 | sidereon_static_position_options_init | positioning.StaticPositionOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1352 | 34035 | sidereon_static_position_solution_clock_biases | positioning.StaticPositionSolutionClockBiases | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1353 | 34047 | sidereon_static_position_solution_epoch_influence | positioning.StaticPositionSolutionEpochInfluence | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1354 | 34059 | sidereon_static_position_solution_free | positioning.StaticPositionSolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1355 | 34067 | sidereon_static_position_solution_geodetic | positioning.StaticPositionSolutionGeodetic | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1356 | 34077 | sidereon_static_position_solution_metadata | positioning.StaticPositionSolutionMetadata | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1357 | 34086 | sidereon_static_position_solution_position | positioning.StaticPositionSolutionPosition | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1358 | 34096 | sidereon_static_position_solution_position_covariance_ecef_m2 | positioning.StaticPositionSolutionPositionCovarianceECEFM2 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1359 | 34106 | sidereon_static_position_solution_position_covariance_enu_m2 | positioning.StaticPositionSolutionPositionCovarianceEnuM2 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1360 | 34117 | sidereon_static_position_solution_rejected_sats | positioning.StaticPositionSolutionRejectedSats | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1361 | 34130 | sidereon_static_position_solution_residuals | positioning.StaticPositionSolutionResiduals | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1362 | 34143 | sidereon_static_position_solution_satellite_batch_influence | positioning.StaticPositionSolutionSatelliteBatchInfluence | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1363 | 34156 | sidereon_static_position_solution_satellite_influence | positioning.StaticPositionSolutionSatelliteInfluence | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1364 | 34169 | sidereon_static_position_solution_state_covariance_m2 | positioning.StaticPositionSolutionStateCovarianceM2 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1365 | 34181 | sidereon_static_reference_station_rinex_config_init | positioning.StaticReferenceStationRINEXConfigInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1366 | 34188 | sidereon_static_reference_station_solution_baseline_ecef | positioning.StaticReferenceStationSolutionBaselineECEF | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1367 | 34197 | sidereon_static_reference_station_solution_covariance_ecef | positioning.StaticReferenceStationSolutionCovarianceECEF | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1368 | 34206 | sidereon_static_reference_station_solution_covariance_enu | positioning.StaticReferenceStationSolutionCovarianceEnu | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1369 | 34218 | sidereon_static_reference_station_solution_diagnostics | positioning.StaticReferenceStationSolutionDiagnostics | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1370 | 34230 | sidereon_static_reference_station_solution_free | positioning.StaticReferenceStationSolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1371 | 34238 | sidereon_static_reference_station_solution_metadata | positioning.StaticReferenceStationSolutionMetadata | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1372 | 34249 | sidereon_static_reference_station_solution_mode_reports | positioning.StaticReferenceStationSolutionModeReports | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1373 | 34260 | sidereon_static_reference_station_solution_position_ecef | positioning.StaticReferenceStationSolutionPositionECEF | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1374 | 34273 | sidereon_status_message | core.Status.String | core status/version API; no omission |
| 1375 | 34284 | sidereon_sub_observer_point | support.SubObserverPoint | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1376 | 34298 | sidereon_sub_solar_point | support.SubSolarPoint | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1377 | 34307 | sidereon_sun_angle_deg | astro.SunAngleDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1378 | 34318 | sidereon_sun_az_el | astro.SunAzEl | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1379 | 34328 | sidereon_sun_elevation_deg | astro.SunElevationDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1380 | 34339 | sidereon_sun_moon_ecef | astro.SunMoonECEF | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1381 | 34351 | sidereon_sun_moon_ecef_batch | astro.SunMoonECEFBatch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1382 | 34365 | sidereon_sun_moon_eci | astro.SunMoonECI | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1383 | 34377 | sidereon_sun_moon_eci_batch | astro.SunMoonECIBatch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1384 | 34392 | sidereon_tai_utc_offset_s | geodesy.TAIUtcOffsetS | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1385 | 34401 | sidereon_tca_collision_probability | astro.TCACollisionProbability | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1386 | 34410 | sidereon_tca_finder_options_init | astro.TCAFinderOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1387 | 34417 | sidereon_tdm_free | support.Tdm.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1388 | 34426 | sidereon_tdm_parse_kvn | support.TdmParseKVN | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1389 | 34436 | sidereon_tdm_participants | support.TdmParticipants | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1390 | 34448 | sidereon_tdm_paths | support.TdmPaths | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1391 | 34459 | sidereon_tdm_record_count | support.TdmRecordCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1392 | 34467 | sidereon_tdm_records | support.TdmRecords | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1393 | 34478 | sidereon_tdm_segment_count | support.TdmSegmentCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1394 | 34486 | sidereon_tdm_segments | support.TdmSegments | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1395 | 34499 | sidereon_tdm_to_kvn | support.TdmToKVN | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1396 | 34512 | sidereon_terminator_latitude_deg | support.TerminatorLatitudeDeg | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1397 | 34523 | sidereon_terrain_geoid_model_label | geodesy.TerrainGeoidModelLabel | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1398 | 34535 | sidereon_terrain_store_checksum64 | geodesy.TerrainStoreChecksum64 | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1399 | 34546 | sidereon_time_scale_abbrev | geodesy.TimeScaleAbbrev | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1400 | 34562 | sidereon_timescale_offset_at_s | geodesy.TimescaleOffsetAtS | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1401 | 34577 | sidereon_timescale_offset_s | geodesy.TimescaleOffsetS | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1402 | 34585 | sidereon_timescales_from_utc | geodesy.TimescalesFromUtc | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1403 | 34603 | sidereon_tle_batch_look_angles | astro.TLEBatchLookAngles | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1404 | 34621 | sidereon_tle_batch_look_angles_free | astro.TLEBatchLookAngles.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1405 | 34630 | sidereon_tle_batch_look_angles_shape | astro.TLEBatchLookAnglesShape | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1406 | 34642 | sidereon_tle_batch_look_angles_values | astro.TLEBatchLookAnglesValues | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1407 | 34657 | sidereon_tle_batch_propagation_free | astro.TLEBatchPropagation.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1408 | 34665 | sidereon_tle_batch_propagation_shape | astro.TLEBatchPropagationShape | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1409 | 34677 | sidereon_tle_batch_propagation_states | astro.TLEBatchPropagationStates | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1410 | 34691 | sidereon_tle_checksum_warnings | astro.TLEChecksumWarnings | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1411 | 34703 | sidereon_tle_file_count | astro.TLEFileCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1412 | 34714 | sidereon_tle_file_free | astro.TLEFile.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1413 | 34727 | sidereon_tle_file_name | astro.TLEFileName | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1414 | 34742 | sidereon_tle_file_satellite | astro.TLEFileSatellite | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1415 | 34754 | sidereon_tle_file_skipped | astro.TLEFileSkipped | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1416 | 34765 | sidereon_tle_find_passes | astro.TLEFindPasses | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1417 | 34779 | sidereon_tle_free | astro.TLE.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1418 | 34789 | sidereon_tle_ground_track | astro.TLEGroundTrack | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1419 | 34802 | sidereon_tle_load | astro.TLELoad | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1420 | 34816 | sidereon_tle_look_angles | astro.TLELookAngles | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1421 | 34828 | sidereon_tle_metadata | astro.TLEMetadata | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1422 | 34839 | sidereon_tle_propagate | astro.TLEPropagate | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1423 | 34854 | sidereon_tle_propagate_with_decay_latch | astro.TLEPropagateWithDecayLatch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1424 | 34864 | sidereon_tle_propagation_epoch_count | astro.TLEPropagationEpochCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1425 | 34876 | sidereon_tle_propagation_free | astro.TLEPropagation.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1426 | 34886 | sidereon_tle_propagation_states | astro.TLEPropagationStates | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1427 | 34898 | sidereon_tle_to_lines | astro.TLEToLines | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1428 | 34906 | sidereon_track_filter_config_dimension | errormetrics.TrackFilterConfigDimension | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1429 | 34914 | sidereon_track_filter_config_frame | astro.TrackFilterConfigFrame | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1430 | 34922 | sidereon_track_filter_config_free | errormetrics.TrackFilterConfig.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1431 | 34931 | sidereon_track_filter_config_from_position | errormetrics.TrackFilterConfigFromPosition | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1432 | 34949 | sidereon_track_filter_config_from_position_velocity | errormetrics.TrackFilterConfigFromPositionVelocity | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1433 | 34965 | sidereon_track_filter_covariance | errormetrics.TrackFilterCovariance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1434 | 34976 | sidereon_track_filter_free | errormetrics.TrackFilter.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1435 | 34983 | sidereon_track_filter_new | errormetrics.TrackFilterNew | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1436 | 34993 | sidereon_track_filter_new_from_position | errormetrics.TrackFilterNewFromPosition | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1437 | 35012 | sidereon_track_filter_position_innovation | errormetrics.TrackFilterPositionInnovation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1438 | 35029 | sidereon_track_filter_position_m | errormetrics.TrackFilterPositionM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1439 | 35040 | sidereon_track_filter_predict | errormetrics.TrackFilterPredict | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1440 | 35049 | sidereon_track_filter_predict_recorded | errormetrics.TrackFilterPredictRecorded | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1441 | 35059 | sidereon_track_filter_record_prediction_only | errormetrics.TrackFilterRecordPredictionOnly | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1442 | 35067 | sidereon_track_filter_state | errormetrics.TrackFilterState | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1443 | 35078 | sidereon_track_filter_state_innovation | errormetrics.TrackFilterStateInnovation | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1444 | 35095 | sidereon_track_filter_state_vector | errormetrics.TrackFilterStateVector | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1445 | 35107 | sidereon_track_filter_update_position | errormetrics.TrackFilterUpdatePosition | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1446 | 35120 | sidereon_track_filter_update_position_gated | errormetrics.TrackFilterUpdatePositionGated | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1447 | 35135 | sidereon_track_filter_update_position_gated_recorded | errormetrics.TrackFilterUpdatePositionGatedRecorded | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1448 | 35150 | sidereon_track_filter_update_position_recorded | errormetrics.TrackFilterUpdatePositionRecorded | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1449 | 35164 | sidereon_track_filter_update_state | errormetrics.TrackFilterUpdateState | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1450 | 35177 | sidereon_track_filter_velocity_m_s | errormetrics.TrackFilterVelocityMS | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1451 | 35188 | sidereon_track_rts_history_builder_finish | errormetrics.TrackRtsHistoryBuilderFinish | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1452 | 35196 | sidereon_track_rts_history_builder_free | errormetrics.TrackRtsHistoryBuilder.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1453 | 35203 | sidereon_track_rts_history_builder_from_filter | errormetrics.TrackRtsHistoryBuilderFromFilter | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1454 | 35211 | sidereon_track_rts_history_builder_new | errormetrics.TrackRtsHistoryBuilderNew | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1455 | 35218 | sidereon_track_rts_history_epoch | errormetrics.TrackRtsHistoryEpoch | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1456 | 35227 | sidereon_track_rts_history_epoch_count | errormetrics.TrackRtsHistoryEpochCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1457 | 35236 | sidereon_track_rts_history_epoch_predicted_position_m | errormetrics.TrackRtsHistoryEpochPredictedPositionM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1458 | 35250 | sidereon_track_rts_history_epoch_transition_from_previous | errormetrics.TrackRtsHistoryEpochTransitionFromPrevious | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1459 | 35263 | sidereon_track_rts_history_epoch_updated_position_m | errormetrics.TrackRtsHistoryEpochUpdatedPositionM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1460 | 35275 | sidereon_track_rts_history_free | errormetrics.TrackRtsHistory.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1461 | 35283 | sidereon_trls_drop_one_base_summary | support.TrlsDropOneBaseSummary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1462 | 35296 | sidereon_trls_drop_one_cost_delta | support.TrlsDropOneCostDelta | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1463 | 35308 | sidereon_trls_drop_one_count | support.TrlsDropOneCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1464 | 35318 | sidereon_trls_drop_one_drop_summary | support.TrlsDropOneDropSummary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1465 | 35330 | sidereon_trls_drop_one_drop_x | support.TrlsDropOneDropX | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1466 | 35343 | sidereon_trls_drop_one_free | support.TrlsDropOne.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1467 | 35351 | sidereon_trls_solution_free | support.TrlsSolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1468 | 35361 | sidereon_trls_solution_gradient | support.TrlsSolutionGradient | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1469 | 35375 | sidereon_trls_solution_jacobian | support.TrlsSolutionJacobian | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1470 | 35389 | sidereon_trls_solution_residuals | support.TrlsSolutionResiduals | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1471 | 35402 | sidereon_trls_solution_summary | support.TrlsSolutionSummary | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1472 | 35413 | sidereon_trls_solution_x | support.TrlsSolutionX | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1473 | 35427 | sidereon_tropo_mapping_factors | geodesy.TropoMappingFactors | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1474 | 35440 | sidereon_tropo_mapping_factors_checked | geodesy.TropoMappingFactorsChecked | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1475 | 35455 | sidereon_tropo_slant_delay | geodesy.TropoSlantDelay | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1476 | 35472 | sidereon_tropo_zenith_delay | geodesy.TropoZenithDelay | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1477 | 35476 | sidereon_true_to_eccentric_anomaly | support.TrueToEccentricAnomaly | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1478 | 35480 | sidereon_true_to_mean_anomaly | astro.TrueToMeanAnomaly | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1479 | 35489 | sidereon_ut1_coverage_covers_jd_tt | geodesy.UT1CoverageCoversJdTt | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1480 | 35496 | sidereon_ut1_coverage_info | geodesy.UT1CoverageInfo | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1481 | 35504 | sidereon_ut1_coverage_source | geodesy.UT1CoverageSource | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1482 | 35518 | sidereon_validate_receiver_solution | support.ValidateReceiverSolution | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1483 | 35527 | sidereon_velocity_options_init | support.VelocityOptionsInit | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1484 | 35535 | sidereon_velocity_solution_clock_drift | errormetrics.VelocitySolutionClockDrift | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1485 | 35545 | sidereon_velocity_solution_free | support.VelocitySolution.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1486 | 35556 | sidereon_velocity_solution_residuals | support.VelocitySolutionResiduals | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1487 | 35568 | sidereon_velocity_solution_speed | support.VelocitySolutionSpeed | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1488 | 35580 | sidereon_velocity_solution_state_covariance | errormetrics.VelocitySolutionStateCovariance | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1489 | 35590 | sidereon_velocity_solution_used_sat_count | support.VelocitySolutionUsedSatCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1490 | 35601 | sidereon_velocity_solution_used_sat_ids | support.VelocitySolutionUsedSatIds | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1491 | 35614 | sidereon_velocity_solution_velocity | support.VelocitySolutionVelocity | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1492 | 35625 | sidereon_version | core.Version | core status/version API; no omission |
| 1493 | 35631 | sidereon_version_string | core.VersionString | core status/version API; no omission |
| 1494 | 35639 | sidereon_vertical_datum_label | geodesy.VerticalDatumLabel | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1495 | 35662 | sidereon_visible_from_satellites | astro.VisibleFromSatellites | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1496 | 35676 | sidereon_visible_list_count | astro.VisibleListCount | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1497 | 35688 | sidereon_visible_list_free | astro.VisibleList.Close | matching owning handle Close; explicit release plus runtime.AddCleanup backstop |
| 1498 | 35698 | sidereon_visible_list_values | astro.VisibleListValues | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1499 | 35710 | sidereon_wavelength_m | support.WavelengthM | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1500 | 35721 | sidereon_weight_vector | errormetrics.WeightVector | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1501 | 35735 | sidereon_write_dted_tile_list_to_mmap_store | geodesy.WriteDTEDTileListToMMapStore | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1502 | 35745 | sidereon_write_dted_tree_to_mmap_store | distribution.WriteDTEDTreeToMmapStore | C call bridged; non-OK maps to typed StatusError; outputs copied |
| 1503 | 35752 | sidereon_wtest_noncentrality | errormetrics.WtestNoncentrality | C call bridged; non-OK maps to typed StatusError; outputs copied |
End of map: 1,503 rows; 1,503 unique function names; zero exclusions.
