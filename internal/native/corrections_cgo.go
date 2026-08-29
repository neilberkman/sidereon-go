//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"runtime"
	"unsafe"
)

type NativeAirborneModel struct{ SigmaNoiseDivergenceM float64 }
type NativeDegradationParams struct {
	DeltaUDRE, EpsFCM, EpsRRCM, EpsLTCM, EpsERM, EpsIonoM float64
	RSSUDRE                                               bool
}
type NativeSbasKMultipliers struct{ KH, KV float64 }
type NativeSbasProtectionRow struct {
	SatelliteID  string
	LineOfSight  [3]float64
	System       uint32
	ElevationRad float64
}
type NativeSbasSISError struct {
	SatelliteID                                   string
	SigmaFLTM, SigmaUIREM, SigmaAirM, SigmaTropoM float64
}
type NativeSbasProtection struct {
	HPLM, VPLM, DMajorM, SigmaUM, DEastM, DNorthM, DENM2 float64
}

func SbasAirborneModelAADA() (NativeAirborneModel, error) {
	var value C.SidereonAirborneModel
	err := callStatus(func() uint32 { return C.sidereon_sbas_airborne_model_aad_a(&value) })
	return NativeAirborneModel{SigmaNoiseDivergenceM: float64(value.sigma_noise_divergence_m)}, err
}
func SbasAirborneSigma(model NativeAirborneModel, elevation float64) (float64, error) {
	var output C.double
	var err error
	withCThread(func() {
		memory, allocationErr := checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonAirborneModel{}))
		if allocationErr != nil {
			err = allocationErr
			return
		}
		defer C.free(memory)
		value := (*C.SidereonAirborneModel)(memory)
		value.sigma_noise_divergence_m = C.double(model.SigmaNoiseDivergenceM)
		err = statusErrorLocked(C.sidereon_sbas_airborne_sigma_air_m(value, C.double(elevation), &output))
	})
	return float64(output), err
}
func SbasDegradationParamsNone() (NativeDegradationParams, error) {
	var value C.SidereonDegradationParams
	err := callStatus(func() uint32 { return C.sidereon_sbas_degradation_params_none(&value) })
	return NativeDegradationParams{DeltaUDRE: float64(value.delta_udre), EpsFCM: float64(value.eps_fc_m), EpsRRCM: float64(value.eps_rrc_m), EpsLTCM: float64(value.eps_ltc_m), EpsERM: float64(value.eps_er_m), EpsIonoM: float64(value.eps_iono_m), RSSUDRE: bool(value.rss_udre)}, err
}
func SbasKMultipliersEnRouteNPA() (NativeSbasKMultipliers, error) {
	var value C.SidereonSbasKMultipliers
	err := callStatus(func() uint32 { return C.sidereon_sbas_k_multipliers_en_route_npa(&value) })
	return NativeSbasKMultipliers{KH: float64(value.k_h), KV: float64(value.k_v)}, err
}
func SbasKMultipliersPrecisionApproach() (NativeSbasKMultipliers, error) {
	var value C.SidereonSbasKMultipliers
	err := callStatus(func() uint32 { return C.sidereon_sbas_k_multipliers_precision_approach(&value) })
	return NativeSbasKMultipliers{KH: float64(value.k_h), KV: float64(value.k_v)}, err
}
func SbasSigmaFLTM(udrei uint8, params NativeDegradationParams) (float64, error) {
	var output C.double
	var err error
	withCThread(func() {
		memory, allocationErr := checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonDegradationParams{}))
		if allocationErr != nil {
			err = allocationErr
			return
		}
		defer C.free(memory)
		value := (*C.SidereonDegradationParams)(memory)
		*value = C.SidereonDegradationParams{delta_udre: C.double(params.DeltaUDRE), eps_fc_m: C.double(params.EpsFCM), eps_rrc_m: C.double(params.EpsRRCM), eps_ltc_m: C.double(params.EpsLTCM), eps_er_m: C.double(params.EpsERM), eps_iono_m: C.double(params.EpsIonoM), rss_udre: C.bool(params.RSSUDRE)}
		err = statusErrorLocked(C.sidereon_sbas_sigma_flt_m_for_udrei(C.uint8_t(udrei), value, &output))
	})
	return float64(output), err
}
func SbasSigmaTropo(elevation float64) (float64, error) {
	var output C.double
	err := callStatus(func() uint32 { return C.sidereon_sbas_sigma_tropo_m(C.double(elevation), &output) })
	return float64(output), err
}
func SbasSISErrorSigma(value NativeSbasSISError) (float64, error) {
	var output C.double
	if err := rejectEmbeddedNUL(value.SatelliteID, "satellite token"); err != nil {
		return 0, err
	}
	if len(value.SatelliteID) >= 16 {
		return 0, errTokenTooLong
	}
	err := withStringError(value.SatelliteID, func(satellite *C.char) error {
		memory, allocationErr := checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonSbasSisError{}))
		if allocationErr != nil {
			return allocationErr
		}
		defer C.free(memory)
		row := (*C.SidereonSbasSisError)(memory)
		*row = C.SidereonSbasSisError{sat_id: satellite, sigma_flt_m: C.double(value.SigmaFLTM), sigma_uire_m: C.double(value.SigmaUIREM), sigma_air_m: C.double(value.SigmaAirM), sigma_tropo_m: C.double(value.SigmaTropoM)}
		return statusErrorLocked(C.sidereon_sbas_sis_error_sigma_m(row, &output))
	})
	return float64(output), err
}
func SbasProtectionLevels(rows []NativeSbasProtectionRow, receiver [3]float64, clockSystems []uint32, model []NativeSbasSISError, multipliers NativeSbasKMultipliers) (NativeSbasProtection, uint32, error) {
	if len(rows) != len(model) {
		return NativeSbasProtection{}, 0, errors.New("sidereon: SBAS geometry and error-model lengths differ")
	}
	var output C.SidereonSbasProtection
	var outputError C.enum_SidereonSbasPlError
	var result NativeSbasProtection
	clockSystemCount, countErr := checkedNativeSize(len(clockSystems))
	if countErr != nil {
		return result, 0, countErr
	}
	err := withCStringArray(nativeSbasIDs(rows), func(rowIDs **C.char, rowCount C.size_t) error {
		return withCStringArray(nativeSbasErrorIDs(model), func(errorIDs **C.char, _ C.size_t) error {
			rowMemory, err := checkedNativeMalloc(len(rows), unsafe.Sizeof(C.SidereonSbasProtectionRow{}))
			if err != nil {
				return err
			}
			if rowMemory != nil {
				defer C.free(rowMemory)
			}
			errorMemory, err := checkedNativeMalloc(len(model), unsafe.Sizeof(C.SidereonSbasSisError{}))
			if err != nil {
				return err
			}
			if errorMemory != nil {
				defer C.free(errorMemory)
			}
			clockMemory, err := checkedNativeMalloc(len(clockSystems), unsafe.Sizeof(C.uint32_t(0)))
			if err != nil {
				return err
			}
			if clockMemory != nil {
				defer C.free(clockMemory)
			}
			cRows := unsafe.Slice((*C.SidereonSbasProtectionRow)(rowMemory), len(rows))
			cErrors := unsafe.Slice((*C.SidereonSbasSisError)(errorMemory), len(model))
			rowPointers := unsafe.Slice(rowIDs, len(rows))
			errorPointers := unsafe.Slice(errorIDs, len(model))
			for i, row := range rows {
				cRows[i] = C.SidereonSbasProtectionRow{sat_id: rowPointers[i], line_of_sight: C.SidereonLineOfSight{e_x: C.double(row.LineOfSight[0]), e_y: C.double(row.LineOfSight[1]), e_z: C.double(row.LineOfSight[2])}, system: C.uint32_t(row.System), elevation_rad: C.double(row.ElevationRad)}
			}
			for i, row := range model {
				cErrors[i] = C.SidereonSbasSisError{sat_id: errorPointers[i], sigma_flt_m: C.double(row.SigmaFLTM), sigma_uire_m: C.double(row.SigmaUIREM), sigma_air_m: C.double(row.SigmaAirM), sigma_tropo_m: C.double(row.SigmaTropoM)}
			}
			cClockSystems := unsafe.Slice((*C.uint32_t)(clockMemory), len(clockSystems))
			for i, value := range clockSystems {
				cClockSystems[i] = C.uint32_t(value)
			}
			var clockPointer *C.uint32_t
			if len(cClockSystems) > 0 {
				clockPointer = &cClockSystems[0]
			}
			var cRowPointer *C.SidereonSbasProtectionRow
			var cErrorPointer *C.SidereonSbasSisError
			if len(cRows) > 0 {
				cRowPointer = &cRows[0]
			}
			if len(cErrors) > 0 {
				cErrorPointer = &cErrors[0]
			}
			geometryMemory, err := checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonSbasProtectionGeometry{}))
			if err != nil {
				return err
			}
			defer C.free(geometryMemory)
			geometry := (*C.SidereonSbasProtectionGeometry)(geometryMemory)
			*geometry = C.SidereonSbasProtectionGeometry{rows: cRowPointer, row_count: rowCount, receiver: C.SidereonGeodetic{lat_rad: C.double(receiver[0]), lon_rad: C.double(receiver[1]), height_m: C.double(receiver[2])}, clock_systems: clockPointer, clock_system_count: clockSystemCount}
			errorModelMemory, err := checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonSbasErrorModel{}))
			if err != nil {
				return err
			}
			defer C.free(errorModelMemory)
			errorModel := (*C.SidereonSbasErrorModel)(errorModelMemory)
			*errorModel = C.SidereonSbasErrorModel{rows: cErrorPointer, row_count: rowCount}
			k := C.SidereonSbasKMultipliers{k_h: C.double(multipliers.KH), k_v: C.double(multipliers.KV)}
			if err := callStatus(func() uint32 {
				return C.sidereon_sbas_protection_levels(geometry, errorModel, k, &output, &outputError)
			}); err != nil {
				return err
			}
			if uint32(outputError) > SBASPLInvalidErrorModelValue {
				return invalidArgument("invalid SBAS protection outcome returned by native code")
			}
			result = NativeSbasProtection{HPLM: float64(output.hpl_m), VPLM: float64(output.vpl_m), DMajorM: float64(output.d_major_m), SigmaUM: float64(output.sigma_u_m), DEastM: float64(output.d_east_m), DNorthM: float64(output.d_north_m), DENM2: float64(output.d_en_m2)}
			return nil
		})
	})
	return result, uint32(outputError), err
}

func nativeSbasIDs(rows []NativeSbasProtectionRow) []string {
	out := make([]string, len(rows))
	for i, row := range rows {
		out[i] = row.SatelliteID
	}
	return out
}
func nativeSbasErrorIDs(rows []NativeSbasSISError) []string {
	out := make([]string, len(rows))
	for i, row := range rows {
		out[i] = row.SatelliteID
	}
	return out
}

type NativeSBASFastCorrection struct {
	PRCM, RRCMPerS float64
	UDREI          uint8
	TOfJ2000S      float64
	IODF           uint8
}
type NativeSBASLongTermCorrection struct {
	IODE                               uint8
	DeltaECEFM, DeltaECEFRateMPerS     [3]float64
	DeltaAF0S, DeltaAF1SPerS, T0J2000S float64
}
type NativeSBASGeoState struct {
	Position, Velocity, Acceleration     [3]float64
	ClockOffsetS, ClockDriftSS, T0J2000S float64
}
type NativeSBASIGP struct {
	LatDeg, LonDeg, VerticalDelayM float64
	HasGIVEVariance                bool
	GIVEVarianceM2                 float64
}
type SBASCorrectionStore struct {
	_        noCopy
	resource *resource
	cleanup  runtime.Cleanup
}
type NativeSSRClockCorrection struct {
	Source                                                       uint32
	ProviderID                                                   uint16
	SolutionID, IODSSR                                           uint8
	C0M, C1MPerS, C2MPerS2, RefEpochJ2000S, UpdateIntervalS      float64
	HasHighRate                                                  bool
	HighRateC0M, HighRateRefEpochJ2000S, HighRateUpdateIntervalS float64
}
type NativeSSROrbitCorrection struct {
	Source                                                                                                    uint32
	ProviderID                                                                                                uint16
	SolutionID                                                                                                uint8
	IODE                                                                                                      uint32
	IODSSR                                                                                                    uint8
	CRSRegional                                                                                               bool
	ReferencePoint                                                                                            uint32
	RadialM, AlongM, CrossM, RadialRateMPerS, AlongRateMPerS, CrossRateMPerS, RefEpochJ2000S, UpdateIntervalS float64
}
type SSRCorrectionStore struct {
	_        noCopy
	resource *resource
	cleanup  runtime.Cleanup
}

func newSBASStore(p *C.SidereonSbasCorrectionStore) (*SBASCorrectionStore, error) {
	if p == nil {
		return nil, errNilNativeHandle
	}
	h := &SBASCorrectionStore{resource: &resource{ptr: unsafe.Pointer(p), release: func(x unsafe.Pointer) { C.sidereon_sbas_store_free((*C.SidereonSbasCorrectionStore)(x)) }}}
	h.cleanup = runtime.AddCleanup(h, cleanupResource, h.resource)
	return h, nil
}
func newSSRStore(p *C.SidereonSsrCorrectionStore) (*SSRCorrectionStore, error) {
	if p == nil {
		return nil, errNilNativeHandle
	}
	h := &SSRCorrectionStore{resource: &resource{ptr: unsafe.Pointer(p), release: func(x unsafe.Pointer) { C.sidereon_ssr_store_free((*C.SidereonSsrCorrectionStore)(x)) }}}
	h.cleanup = runtime.AddCleanup(h, cleanupResource, h.resource)
	return h, nil
}
func cWeek(v NativeGnssWeekTow) C.SidereonGnssWeekTow {
	return C.SidereonGnssWeekTow{system: C.uint32_t(v.System), week: C.uint32_t(v.Week), tow_s: C.double(v.TOWSeconds)}
}
func cGeo(v Geodetic) C.SidereonGeodetic {
	return C.SidereonGeodetic{lat_rad: C.double(v.LatitudeRad), lon_rad: C.double(v.LongitudeRad), height_m: C.double(v.HeightM)}
}

func validateSBASSolveModeValue(value uint32) error {
	if value != SBASSolveMixedValue && value != SBASSolveSBASOnlyValue {
		return invalidArgument("invalid SBAS solve mode")
	}
	return nil
}

func validateSSRReferencePointValue(value uint32) error {
	if value != SSRReferencePointAntennaValue && value != SSRReferencePointCenterOfMassValue {
		return invalidArgument("invalid SSR reference point")
	}
	return nil
}

func validateSSRMissingActionValue(value uint32) error {
	if value != SSRMissingDeclineValue && value != SSRMissingFallbackValue {
		return invalidArgument("invalid SSR missing-correction action")
	}
	return nil
}

func ParseBroadcast(data []byte) (*BroadcastEphemeris, error) { return ParseBroadcastEphemeris(data) }

func NewSBASStore() (*SBASCorrectionStore, error) {
	var p *C.SidereonSbasCorrectionStore
	err := callStatus(func() uint32 { return C.sidereon_sbas_store_new(&p) })
	if err != nil {
		if p != nil {
			withCThread(func() { C.sidereon_sbas_store_free(p) })
		}
		return nil, err
	}
	handle, err := newSBASStore(p)
	if err != nil && p != nil {
		withCThread(func() { C.sidereon_sbas_store_free(p) })
	}
	return handle, err
}
func (s *SBASCorrectionStore) Close() error {
	if s == nil {
		return nil
	}
	return closeProtocolResource(s, s.resource, &s.cleanup)
}
func (s *SBASCorrectionStore) Ingest(block *SbasBlock, geo string, epoch NativeGnssWeekTow) error {
	if s == nil || s.resource == nil || block == nil || block.resource == nil {
		return ErrClosed
	}
	return block.resource.with(func(bp unsafe.Pointer) error {
		return s.resource.with(func(sp unsafe.Pointer) error {
			return withTokenError(geo, "GEO satellite token", func(g *C.char) error {
				memory, err := checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonGnssWeekTow{}))
				if err != nil {
					return err
				}
				defer C.free(memory)
				week := (*C.SidereonGnssWeekTow)(memory)
				*week = cWeek(epoch)
				return statusErrorLocked(C.sidereon_sbas_store_ingest((*C.SidereonSbasCorrectionStore)(sp), (*C.SidereonSbasBlock)(bp), g, week))
			})
		})
	})
}
func (s *SBASCorrectionStore) PreferredGeo(t float64) (string, bool, error) {
	var p C.bool
	var tok C.SidereonSatelliteToken
	err := s.resource.with(func(sp unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_sbas_store_preferred_geo((*C.SidereonSbasCorrectionStore)(sp), C.double(t), &p, &tok)
		})
	})
	return tokenFromC(tok), bool(p), err
}
func (s *SBASCorrectionStore) ReadyGeos(t float64) ([]string, error) {
	var out []string
	err := s.resource.with(func(sp unsafe.Pointer) error {
		var w, r C.size_t
		if e := callStatus(func() uint32 {
			return C.sidereon_sbas_store_ready_geos((*C.SidereonSbasCorrectionStore)(sp), C.double(t), nil, 0, &w, &r)
		}); e != nil {
			return e
		}
		n, e := validateNativeQuery("SBAS ready GEOs", uint64(w), uint64(r))
		if e != nil {
			return e
		}
		if _, e := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonSatelliteToken{})); e != nil {
			return e
		}
		v := make([]C.SidereonSatelliteToken, n)
		var q *C.SidereonSatelliteToken
		if n > 0 {
			q = &v[0]
		}
		if e := callStatus(func() uint32 {
			return C.sidereon_sbas_store_ready_geos((*C.SidereonSbasCorrectionStore)(sp), C.double(t), q, C.size_t(n), &w, &r)
		}); e != nil {
			return e
		}
		z, e := validateNativeOutput("SBAS ready GEOs", n, uint64(w), uint64(r))
		if e != nil {
			return e
		}
		out = make([]string, z)
		for i := range out {
			out[i] = tokenFromC(v[i])
		}
		return nil
	})
	return out, err
}
func (s *SBASCorrectionStore) Fast(geo, sat string) (NativeSBASFastCorrection, bool, error) {
	var p C.bool
	var v C.SidereonSbasFastCorrection
	err := s.resource.with(func(sp unsafe.Pointer) error {
		return withTwoTokens(geo, sat, "GEO satellite token", "satellite token", func(g, x *C.char) uint32 {
			return C.sidereon_sbas_store_fast_correction((*C.SidereonSbasCorrectionStore)(sp), g, x, &p, &v)
		})
	})
	return NativeSBASFastCorrection{PRCM: float64(v.prc_m), RRCMPerS: float64(v.rrc_m_s), UDREI: uint8(v.udrei), TOfJ2000S: float64(v.t_of_j2000_s), IODF: uint8(v.iodf)}, bool(p), err
}
func (s *SBASCorrectionStore) LongTerm(geo, sat string) (NativeSBASLongTermCorrection, bool, error) {
	var p C.bool
	var v C.SidereonSbasLongTermCorrection
	err := s.resource.with(func(sp unsafe.Pointer) error {
		return withTwoTokens(geo, sat, "GEO satellite token", "satellite token", func(g, x *C.char) uint32 {
			return C.sidereon_sbas_store_long_term_correction((*C.SidereonSbasCorrectionStore)(sp), g, x, &p, &v)
		})
	})
	out := NativeSBASLongTermCorrection{IODE: uint8(v.iode), DeltaAF0S: float64(v.delta_af0_s), DeltaAF1SPerS: float64(v.delta_af1_s_s), T0J2000S: float64(v.t0_j2000_s)}
	for i := range out.DeltaECEFM {
		out.DeltaECEFM[i] = float64(v.delta_ecef_m[i])
		out.DeltaECEFRateMPerS[i] = float64(v.delta_ecef_rate_m_s[i])
	}
	return out, bool(p), err
}
func (s *SBASCorrectionStore) GeoNav(geo string) (NativeSBASGeoState, bool, error) {
	var p C.bool
	var v C.SidereonSbasGeoState
	err := s.resource.with(func(sp unsafe.Pointer) error {
		return withToken(geo, "GEO satellite token", func(g *C.char) uint32 {
			return C.sidereon_sbas_store_geo_nav((*C.SidereonSbasCorrectionStore)(sp), g, &p, &v)
		})
	})
	out := NativeSBASGeoState{ClockOffsetS: float64(v.clock_offset_s), ClockDriftSS: float64(v.clock_drift_s_s), T0J2000S: float64(v.t0_j2000_s)}
	for i := range out.Position {
		out.Position[i] = float64(v.position_ecef_m[i])
		out.Velocity[i] = float64(v.velocity_ecef_m_s[i])
		out.Acceleration[i] = float64(v.acceleration_ecef_m_s2[i])
	}
	return out, bool(p), err
}
func (s *SBASCorrectionStore) IGPs(geo string) ([]NativeSBASIGP, error) {
	var out []NativeSBASIGP
	err := s.resource.with(func(sp unsafe.Pointer) error {
		return withStringError(geo, func(g *C.char) error {
			var w, r C.size_t
			if err := callStatus(func() uint32 {
				return C.sidereon_sbas_store_iono_grid_igps((*C.SidereonSbasCorrectionStore)(sp), g, nil, 0, &w, &r)
			}); err != nil {
				return err
			}
			n, err := validateNativeQuery("SBAS IGPs", uint64(w), uint64(r))
			if err != nil {
				return err
			}
			if _, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonSbasIgp{})); err != nil {
				return err
			}
			v := make([]C.SidereonSbasIgp, n)
			var q *C.SidereonSbasIgp
			if n > 0 {
				q = &v[0]
			}
			if err := callStatus(func() uint32 {
				return C.sidereon_sbas_store_iono_grid_igps((*C.SidereonSbasCorrectionStore)(sp), g, q, C.size_t(n), &w, &r)
			}); err != nil {
				return err
			}
			z, err := validateNativeOutput("SBAS IGPs", n, uint64(w), uint64(r))
			if err != nil {
				return err
			}
			out = make([]NativeSBASIGP, z)
			for i := range out {
				out[i] = NativeSBASIGP{LatDeg: float64(v[i].lat_deg), LonDeg: float64(v[i].lon_deg), VerticalDelayM: float64(v[i].vertical_delay_m), HasGIVEVariance: bool(v[i].has_give_variance_m2), GIVEVarianceM2: float64(v[i].give_variance_m2)}
			}
			return nil
		})
	})
	return out, err
}
func (s *SBASCorrectionStore) SlantDelay(geo string, receiver Geodetic, elevation, azimuth, frequency float64) (float64, bool, error) {
	var p C.bool
	var v C.double
	err := s.resource.with(func(sp unsafe.Pointer) error {
		return withTokenError(geo, "GEO satellite token", func(x *C.char) error {
			memory, err := checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonGeodetic{}))
			if err != nil {
				return err
			}
			defer C.free(memory)
			g := (*C.SidereonGeodetic)(memory)
			*g = cGeo(receiver)
			return statusErrorLocked(C.sidereon_sbas_store_iono_slant_delay_m((*C.SidereonSbasCorrectionStore)(sp), x, g, C.double(elevation), C.double(azimuth), C.double(frequency), &p, &v))
		})
	})
	return float64(v), bool(p), err
}
func (s *SBASCorrectionStore) CorrectedState(b *BroadcastEphemeris, geo string, mode uint32, sat string, t float64) ([3]float64, float64, bool, error) {
	var pos [3]C.double
	var clk C.double
	var present C.bool
	var out [3]float64
	if s == nil || s.resource == nil || b == nil || b.resource == nil {
		return out, 0, false, ErrClosed
	}
	if err := validateSBASSolveModeValue(mode); err != nil {
		return out, 0, false, err
	}
	err := b.resource.with(func(bp unsafe.Pointer) error {
		return s.resource.with(func(sp unsafe.Pointer) error {
			return withTwoTokens(geo, sat, "GEO satellite token", "satellite token", func(g, x *C.char) uint32 {
				return C.sidereon_sbas_corrected_state((*C.SidereonBroadcastEphemeris)(bp), (*C.SidereonSbasCorrectionStore)(sp), g, C.uint32_t(mode), x, C.double(t), &present, &pos[0], &clk)
			})
		})
	})
	for i := range out {
		out[i] = float64(pos[i])
	}
	return out, float64(clk), bool(present), err
}

func NewSSRStore(referencePoint uint32) (*SSRCorrectionStore, error) {
	if err := validateSSRReferencePointValue(referencePoint); err != nil {
		return nil, err
	}
	var p *C.SidereonSsrCorrectionStore
	err := callStatus(func() uint32 { return C.sidereon_ssr_store_new(C.uint32_t(referencePoint), &p) })
	if err != nil {
		if p != nil {
			withCThread(func() { C.sidereon_ssr_store_free(p) })
		}
		return nil, err
	}
	handle, err := newSSRStore(p)
	if err != nil && p != nil {
		withCThread(func() { C.sidereon_ssr_store_free(p) })
	}
	return handle, err
}
func NewSSRStoreFromRTCM(data []byte, epoch NativeGnssWeekTow) (*SSRCorrectionStore, error) {
	var p *C.SidereonSsrCorrectionStore
	err := withInputError(data, func(b *C.uint8_t, n C.size_t) error {
		memory, allocationErr := checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonGnssWeekTow{}))
		if allocationErr != nil {
			return allocationErr
		}
		defer C.free(memory)
		week := (*C.SidereonGnssWeekTow)(memory)
		*week = cWeek(epoch)
		return statusErrorLocked(C.sidereon_ssr_store_from_rtcm(b, n, week, &p))
	})
	if err != nil {
		if p != nil {
			withCThread(func() { C.sidereon_ssr_store_free(p) })
		}
		return nil, err
	}
	handle, err := newSSRStore(p)
	if err != nil && p != nil {
		withCThread(func() { C.sidereon_ssr_store_free(p) })
	}
	return handle, err
}
func (s *SSRCorrectionStore) Close() error {
	if s == nil {
		return nil
	}
	return closeProtocolResource(s, s.resource, &s.cleanup)
}
func (s *SSRCorrectionStore) Ingest(messages *RtcmMessages, epoch NativeGnssWeekTow) error {
	if s == nil || s.resource == nil || messages == nil || messages.resource == nil {
		return ErrClosed
	}
	return messages.resource.with(func(mp unsafe.Pointer) error {
		return s.resource.with(func(sp unsafe.Pointer) error {
			memory, err := checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonGnssWeekTow{}))
			if err != nil {
				return err
			}
			defer C.free(memory)
			week := (*C.SidereonGnssWeekTow)(memory)
			*week = cWeek(epoch)
			return statusErrorLocked(C.sidereon_ssr_store_ingest_messages((*C.SidereonSsrCorrectionStore)(sp), (*C.SidereonRtcmMessages)(mp), week))
		})
	})
}
func (s *SSRCorrectionStore) Orbit(sat string) (NativeSSROrbitCorrection, bool, error) {
	var p C.bool
	var v C.SidereonSsrOrbitCorrection
	err := s.resource.with(func(sp unsafe.Pointer) error {
		return withToken(sat, "satellite token", func(x *C.char) uint32 {
			return C.sidereon_ssr_store_orbit((*C.SidereonSsrCorrectionStore)(sp), x, &p, &v)
		})
	})
	out := NativeSSROrbitCorrection{Source: uint32(v.source), ProviderID: uint16(v.provider_id), SolutionID: uint8(v.solution_id), IODE: uint32(v.iode), IODSSR: uint8(v.iod_ssr), CRSRegional: bool(v.crs_regional), ReferencePoint: uint32(v.reference_point), RadialM: float64(v.radial_m), AlongM: float64(v.along_m), CrossM: float64(v.cross_m), RadialRateMPerS: float64(v.radial_rate_m_s), AlongRateMPerS: float64(v.along_rate_m_s), CrossRateMPerS: float64(v.cross_rate_m_s), RefEpochJ2000S: float64(v.ref_epoch_j2000_s), UpdateIntervalS: float64(v.update_interval_s)}
	if err == nil {
		if err = validateSSRReferencePointValue(out.ReferencePoint); err != nil {
			return NativeSSROrbitCorrection{}, false, err
		}
		if out.Source != SSRSourceRTCMValue && out.Source != SSRSourceGalileoHASValue {
			return NativeSSROrbitCorrection{}, false, invalidArgument("invalid SSR source returned by native code")
		}
	}
	return out, bool(p), err
}
func (s *SSRCorrectionStore) Clock(sat string) (NativeSSRClockCorrection, bool, error) {
	var p C.bool
	var v C.SidereonSsrClockCorrection
	err := s.resource.with(func(sp unsafe.Pointer) error {
		return withToken(sat, "satellite token", func(x *C.char) uint32 {
			return C.sidereon_ssr_store_clock((*C.SidereonSsrCorrectionStore)(sp), x, &p, &v)
		})
	})
	out := NativeSSRClockCorrection{Source: uint32(v.source), ProviderID: uint16(v.provider_id), SolutionID: uint8(v.solution_id), IODSSR: uint8(v.iod_ssr), C0M: float64(v.c0_m), C1MPerS: float64(v.c1_m_s), C2MPerS2: float64(v.c2_m_s2), RefEpochJ2000S: float64(v.ref_epoch_j2000_s), UpdateIntervalS: float64(v.update_interval_s), HasHighRate: bool(v.has_high_rate), HighRateC0M: float64(v.high_rate_c0_m), HighRateRefEpochJ2000S: float64(v.high_rate_ref_epoch_j2000_s), HighRateUpdateIntervalS: float64(v.high_rate_update_interval_s)}
	if err == nil && out.Source != SSRSourceRTCMValue && out.Source != SSRSourceGalileoHASValue {
		return NativeSSRClockCorrection{}, false, invalidArgument("invalid SSR source returned by native code")
	}
	return out, bool(p), err
}
func (s *SSRCorrectionStore) CodeBias(sat string, signal uint8) (float64, bool, error) {
	var p C.bool
	var v C.double
	err := s.resource.with(func(sp unsafe.Pointer) error {
		return withToken(sat, "satellite token", func(x *C.char) uint32 {
			return C.sidereon_ssr_store_code_bias_m((*C.SidereonSsrCorrectionStore)(sp), x, C.uint8_t(signal), &p, &v)
		})
	})
	return float64(v), bool(p), err
}
func (s *SSRCorrectionStore) PhaseBias(sat string, signal uint8) (float64, bool, error) {
	var p C.bool
	var v C.double
	err := s.resource.with(func(sp unsafe.Pointer) error {
		return withToken(sat, "satellite token", func(x *C.char) uint32 {
			return C.sidereon_ssr_store_phase_bias_m((*C.SidereonSsrCorrectionStore)(sp), x, C.uint8_t(signal), &p, &v)
		})
	})
	return float64(v), bool(p), err
}
func (s *SSRCorrectionStore) URA(sat string) (uint8, bool, error) {
	var p C.bool
	var v C.uint8_t
	err := s.resource.with(func(sp unsafe.Pointer) error {
		return withToken(sat, "satellite token", func(x *C.char) uint32 {
			return C.sidereon_ssr_store_ura_index((*C.SidereonSsrCorrectionStore)(sp), x, &p, &v)
		})
	})
	return uint8(v), bool(p), err
}
func (s *SSRCorrectionStore) CorrectedState(b *BroadcastEphemeris, sat string, t, staleness float64, missing uint32, allow bool, provider uint16) ([3]float64, float64, bool, error) {
	var pos [3]C.double
	var clk C.double
	var present C.bool
	var out [3]float64
	if s == nil || s.resource == nil || b == nil || b.resource == nil {
		return out, 0, false, ErrClosed
	}
	if err := validateSSRMissingActionValue(missing); err != nil {
		return out, 0, false, err
	}
	err := b.resource.with(func(bp unsafe.Pointer) error {
		return s.resource.with(func(sp unsafe.Pointer) error {
			return withToken(sat, "satellite token", func(x *C.char) uint32 {
				return C.sidereon_ssr_corrected_state((*C.SidereonBroadcastEphemeris)(bp), (*C.SidereonSsrCorrectionStore)(sp), x, C.double(t), C.double(staleness), C.uint32_t(missing), C.bool(allow), C.uint16_t(provider), &present, &pos[0], &clk)
			})
		})
	})
	for i := range out {
		out[i] = float64(pos[i])
	}
	return out, float64(clk), bool(present), err
}

func (s *SSRCorrectionStore) EphemerisSample(b *BroadcastEphemeris, satellites []string, staleness float64, missing uint32, allow bool, provider uint16, start, stop, step float64) ([]NativeEphemerisSampleRow, error) {
	if s == nil || s.resource == nil || b == nil || b.resource == nil {
		return nil, ErrClosed
	}
	if err := validateSSRMissingActionValue(missing); err != nil {
		return nil, err
	}
	var result []NativeEphemerisSampleRow
	err := b.resource.with(func(broadcastPointer unsafe.Pointer) error {
		return s.resource.with(func(storePointer unsafe.Pointer) error {
			return withCStringArray(satellites, func(satellitePointer **C.char, count C.size_t) error {
				var written, required C.size_t
				call := func(out *C.SidereonEphemerisSampleRow, length C.size_t) uint32 {
					return C.sidereon_ssr_ephemeris_sample((*C.SidereonBroadcastEphemeris)(broadcastPointer), (*C.SidereonSsrCorrectionStore)(storePointer), C.double(staleness), C.uint32_t(missing), C.bool(allow), C.uint16_t(provider), satellitePointer, count, C.double(start), C.double(stop), C.double(step), out, length, &written, &required)
				}
				if err := callStatus(func() uint32 { return call(nil, 0) }); err != nil {
					return err
				}
				n, err := validateNativeQuery("SSR ephemeris sample", uint64(written), uint64(required))
				if err != nil {
					return err
				}
				if _, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonEphemerisSampleRow{})); err != nil {
					return err
				}
				values := make([]C.SidereonEphemerisSampleRow, n)
				var output *C.SidereonEphemerisSampleRow
				if n > 0 {
					output = &values[0]
				}
				if err := callStatus(func() uint32 { return call(output, C.size_t(n)) }); err != nil {
					return err
				}
				writtenCount, err := validateNativeOutput("SSR ephemeris sample", n, uint64(written), uint64(required))
				if err != nil {
					return err
				}
				result = make([]NativeEphemerisSampleRow, writtenCount)
				for i := range result {
					value := values[i]
					if err := validateEphemerisSampleStatusValue(uint32(value.status)); err != nil {
						return err
					}
					row := NativeEphemerisSampleRow{SatelliteID: tokenFromC(value.sat_id), EpochJ2000S: float64(value.epoch_j2000_s), Status: uint32(value.status), HasPosition: bool(value.has_position_ecef_m), HasClock: bool(value.has_clock_s), ClockS: float64(value.clock_s)}
					for axis := range row.Position {
						row.Position[axis] = float64(value.position_ecef_m[axis])
					}
					result[i] = row
				}
				return nil
			})
		})
	})
	runtime.KeepAlive(s)
	runtime.KeepAlive(b)
	return result, err
}
