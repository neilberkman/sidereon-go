//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"fmt"
	"runtime"
	"unsafe"
)

type NativeCompareEpoch struct {
	BroadcastTJ2000S       float64
	PreciseJDWhole         float64
	PreciseJDFraction      float64
	PrecisePlusJDWhole     float64
	PrecisePlusJDFraction  float64
	PreciseMinusJDWhole    float64
	PreciseMinusJDFraction float64
}

type NativeCompareWindow struct {
	BroadcastWindowStartJ2000S float64
	BroadcastWindowEndJ2000S   float64
	PreciseStartJDWhole        float64
	PreciseStartJDFraction     float64
	StepS                      float64
	VelocityHalfS              float64
}

type NativeCompareStats struct {
	Count                          int
	Orbit3DRMSM, Orbit3DMaxM       float64
	RadialRMSM, RadialMaxM         float64
	AlongRMSM, AlongMaxM           float64
	CrossRMSM, CrossMaxM           float64
	ClockRMSM, ClockMaxM           float64
	ClockDatumRMSM, ClockDatumMaxM float64
}

// SelectionError preserves the distinct SidereonSelectionStatus namespace.
// It must not be represented as StatusError: the numeric values overlap with
// unrelated SidereonStatus values.
type SelectionError struct {
	Status uint32
	Text   string
	Detail string
}

func (e *SelectionError) Error() string {
	if e == nil {
		return "sidereon: selection error"
	}
	if e.Detail == "" {
		return fmt.Sprintf("%s (%d)", e.Text, e.Status)
	}
	return fmt.Sprintf("%s (%d): %s", e.Text, e.Status, e.Detail)
}

type BroadcastComparison struct {
	_        noCopy
	resource *resource
	cleanup  runtime.Cleanup
}

func releaseBroadcastComparison(pointer unsafe.Pointer) {
	withCThread(func() {
		C.sidereon_broadcast_comparison_free((*C.SidereonBroadcastComparison)(pointer))
	})
}

func newBroadcastComparison(pointer *C.SidereonBroadcastComparison) (*BroadcastComparison, error) {
	if pointer == nil {
		return nil, missingNativeHandle("broadcast comparison")
	}
	h := &BroadcastComparison{resource: &resource{ptr: unsafe.Pointer(pointer), release: releaseBroadcastComparison}}
	h.cleanup = runtime.AddCleanup(h, cleanupResource, h.resource)
	return h, nil
}

func (r *BroadcastComparison) Close() error {
	if r == nil || r.resource == nil {
		return nil
	}
	r.resource.close()
	return nil
}

func withBroadcastSP3Pair(broadcast *BroadcastEphemeris, precise *SP3, fn func(unsafe.Pointer, unsafe.Pointer) error) error {
	if broadcast == nil || broadcast.resource == nil || precise == nil || precise.handle == nil {
		return ErrClosed
	}
	return withScenarioInputs([]scenarioInput{{resource: broadcast.resource}, {positioning: precise.handle}}, func(pointers []unsafe.Pointer) error {
		return fn(pointers[0], pointers[1])
	})
}

func cCompareStats(value C.SidereonCompareStats) (NativeCompareStats, error) {
	count, err := checkedNativeCount(uint64(value.count))
	if err != nil {
		return NativeCompareStats{}, err
	}
	return NativeCompareStats{
		Count:       count,
		Orbit3DRMSM: float64(value.orbit_3d_rms_m), Orbit3DMaxM: float64(value.orbit_3d_max_m),
		RadialRMSM: float64(value.radial_rms_m), RadialMaxM: float64(value.radial_max_m),
		AlongRMSM: float64(value.along_rms_m), AlongMaxM: float64(value.along_max_m),
		CrossRMSM: float64(value.cross_rms_m), CrossMaxM: float64(value.cross_max_m),
		ClockRMSM: float64(value.clock_rms_m), ClockMaxM: float64(value.clock_max_m),
		ClockDatumRMSM: float64(value.clock_datum_removed_rms_m), ClockDatumMaxM: float64(value.clock_datum_removed_max_m),
	}, nil
}

func cCompareEpoch(value NativeCompareEpoch) C.SidereonCompareEpoch {
	return C.SidereonCompareEpoch{
		broadcast_t_j2000_s: C.double(value.BroadcastTJ2000S),
		precise_jd_whole:    C.double(value.PreciseJDWhole), precise_jd_fraction: C.double(value.PreciseJDFraction),
		precise_plus_jd_whole: C.double(value.PrecisePlusJDWhole), precise_plus_jd_fraction: C.double(value.PrecisePlusJDFraction),
		precise_minus_jd_whole: C.double(value.PreciseMinusJDWhole), precise_minus_jd_fraction: C.double(value.PreciseMinusJDFraction),
	}
}

func cCompareWindow(value NativeCompareWindow) C.SidereonCompareWindow {
	return C.SidereonCompareWindow{
		broadcast_window_start_j2000_s: C.double(value.BroadcastWindowStartJ2000S),
		broadcast_window_end_j2000_s:   C.double(value.BroadcastWindowEndJ2000S),
		precise_start_jd_whole:         C.double(value.PreciseStartJDWhole), precise_start_jd_fraction: C.double(value.PreciseStartJDFraction),
		step_s: C.double(value.StepS), velocity_half_s: C.double(value.VelocityHalfS),
	}
}

func CompareBroadcast(broadcast *BroadcastEphemeris, precise *SP3, satellites []string, epochs []NativeCompareEpoch, velocityHalfS float64) (*BroadcastComparison, error) {
	if broadcast == nil || broadcast.resource == nil || precise == nil || precise.handle == nil {
		return nil, ErrClosed
	}
	epochCount, err := checkedNativeSize(len(epochs))
	if err != nil {
		return nil, err
	}
	var result *BroadcastComparison
	err = withBroadcastSP3Pair(broadcast, precise, func(broadcastPointer, precisePointer unsafe.Pointer) error {
		return withCStringArray(satellites, func(satellitePointer **C.char, satelliteCount C.size_t) error {
			var epochMemory unsafe.Pointer
			if len(epochs) != 0 {
				size, err := checkedNativeAllocationSize(len(epochs), unsafe.Sizeof(C.SidereonCompareEpoch{}))
				if err != nil {
					return err
				}
				epochMemory = C.calloc(1, C.size_t(size))
				if epochMemory == nil {
					return errors.New("sidereon: unable to allocate native comparison epochs")
				}
				defer C.free(epochMemory)
			}
			cEpochs := unsafe.Slice((*C.SidereonCompareEpoch)(epochMemory), len(epochs))
			for i, epoch := range epochs {
				cEpochs[i] = cCompareEpoch(epoch)
			}
			var pointer *C.SidereonBroadcastComparison
			status := C.sidereon_broadcast_comparison_compare(
				(*C.SidereonBroadcastEphemeris)(broadcastPointer), (*C.SidereonSp3)(precisePointer),
				satellitePointer, satelliteCount, (*C.SidereonCompareEpoch)(epochMemory), epochCount, C.double(velocityHalfS), &pointer,
			)
			if err := statusErrorLocked(uint32(status)); err != nil {
				if pointer != nil {
					C.sidereon_broadcast_comparison_free(pointer)
				}
				return err
			}
			var err error
			result, err = newBroadcastComparison(pointer)
			if err != nil && pointer != nil {
				C.sidereon_broadcast_comparison_free(pointer)
			}
			return err
		})
	})
	runtime.KeepAlive(broadcast)
	runtime.KeepAlive(precise)
	runtime.KeepAlive(satellites)
	runtime.KeepAlive(epochs)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func CompareBroadcastWindow(broadcast *BroadcastEphemeris, precise *SP3, satellites []string, window NativeCompareWindow) (*BroadcastComparison, error) {
	if broadcast == nil || broadcast.resource == nil || precise == nil || precise.handle == nil {
		return nil, ErrClosed
	}
	var result *BroadcastComparison
	err := withBroadcastSP3Pair(broadcast, precise, func(broadcastPointer, precisePointer unsafe.Pointer) error {
		return withCStringArray(satellites, func(satellitePointer **C.char, satelliteCount C.size_t) error {
			cWindow := cCompareWindow(window)
			var pointer *C.SidereonBroadcastComparison
			status := C.sidereon_broadcast_comparison_compare_window(
				(*C.SidereonBroadcastEphemeris)(broadcastPointer), (*C.SidereonSp3)(precisePointer),
				satellitePointer, satelliteCount, &cWindow, &pointer,
			)
			if err := statusErrorLocked(uint32(status)); err != nil {
				if pointer != nil {
					C.sidereon_broadcast_comparison_free(pointer)
				}
				return err
			}
			var err error
			result, err = newBroadcastComparison(pointer)
			if err != nil && pointer != nil {
				C.sidereon_broadcast_comparison_free(pointer)
			}
			return err
		})
	})
	runtime.KeepAlive(broadcast)
	runtime.KeepAlive(precise)
	runtime.KeepAlive(satellites)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *BroadcastComparison) Overall() (NativeCompareStats, error) {
	if r == nil || r.resource == nil {
		return NativeCompareStats{}, ErrClosed
	}
	var stats C.SidereonCompareStats
	err := r.resource.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_broadcast_comparison_overall((*C.SidereonBroadcastComparison)(pointer), &stats)
		})
	})
	result, convertErr := cCompareStats(stats)
	if err != nil {
		return NativeCompareStats{}, err
	}
	return result, convertErr
}

func (r *BroadcastComparison) SatelliteCount() (int, error) {
	if r == nil || r.resource == nil {
		return 0, ErrClosed
	}
	var count C.size_t
	err := r.resource.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_broadcast_comparison_satellite_count((*C.SidereonBroadcastComparison)(pointer), &count)
		})
	})
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(count))
}

func (r *BroadcastComparison) Satellite(index int) (string, NativeCompareStats, error) {
	if r == nil || r.resource == nil {
		return "", NativeCompareStats{}, ErrClosed
	}
	indexValue, err := checkedNativeSize(index)
	if err != nil {
		return "", NativeCompareStats{}, err
	}
	var token C.SidereonSatelliteToken
	var stats C.SidereonCompareStats
	err = r.resource.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_broadcast_comparison_satellite((*C.SidereonBroadcastComparison)(pointer), indexValue, (*C.char)(unsafe.Pointer(&token.bytes[0])), C.size_t(len(token.bytes)), &stats)
		})
	})
	if err != nil {
		return "", NativeCompareStats{}, err
	}
	converted, err := cCompareStats(stats)
	if err != nil {
		return "", NativeCompareStats{}, err
	}
	return tokenFromC(token), converted, nil
}

type NativeBroadcastRecordInfo struct {
	SatelliteID       string
	Message, Issue    uint32
	IssueMessage      uint32
	Week, ToeWeek     uint32
	ToeTOWSeconds     float64
	TocWeek           uint32
	TocTOWSeconds     float64
	SVHealth          float64
	SVAccuracyM       float64
	HasFitInterval    bool
	FitIntervalS      float64
	DefaultGroupDelay float64
	CNAV              NativeBroadcastCNAV
}

func cnavInfoFromC(value C.SidereonCnavParameters) NativeBroadcastCNAV {
	return NativeBroadcastCNAV{
		Present: bool(value.present), ADOTMPerS: float64(value.adot_m_s), DeltaN0DotRadPerS2: float64(value.delta_n0_dot_rad_s2),
		Top:        NativeGnssWeekTow{Week: uint32(value.top_week), TOWSeconds: float64(value.top_tow_s)},
		URAEDIndex: int8(value.ura_ed_index), URANED0Index: int8(value.ura_ned0_index), URANED1Index: uint8(value.ura_ned1_index), URANED2Index: uint8(value.ura_ned2_index),
		TransmissionTimeSOW: float64(value.transmission_time_sow), HasFlags: bool(value.has_flags), Flags: uint32(value.flags),
	}
}

func recordInfoFromC(value C.SidereonBroadcastRecordInfo) NativeBroadcastRecordInfo {
	return NativeBroadcastRecordInfo{
		SatelliteID: tokenFromC(value.sat_id), Message: uint32(value.message), Issue: uint32(value.issue), IssueMessage: uint32(value.issue_message),
		Week: uint32(value.week), ToeWeek: uint32(value.toe_week), ToeTOWSeconds: float64(value.toe_tow_s), TocWeek: uint32(value.toc_week), TocTOWSeconds: float64(value.toc_tow_s),
		SVHealth: float64(value.sv_health), SVAccuracyM: float64(value.sv_accuracy_m), HasFitInterval: bool(value.has_fit_interval_s), FitIntervalS: float64(value.fit_interval_s),
		DefaultGroupDelay: float64(value.default_group_delay_s), CNAV: cnavInfoFromC(value.cnav),
	}
}

func (b *BroadcastEphemeris) RecordCNavCorrection(index int, signal uint32) (float64, bool, error) {
	if b == nil || b.resource == nil {
		return 0, false, ErrClosed
	}
	indexValue, err := checkedNativeSize(index)
	if err != nil {
		return 0, false, err
	}
	var value C.double
	var present C.bool
	err = b.resource.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_broadcast_ephemeris_record_cnav_correction((*C.SidereonBroadcastEphemeris)(pointer), indexValue, C.uint32_t(signal), &value, &present)
		})
	})
	return float64(value), bool(present), err
}

func (b *BroadcastEphemeris) RecordGroupDelay(index int, term uint32) (float64, bool, error) {
	if b == nil || b.resource == nil {
		return 0, false, ErrClosed
	}
	indexValue, err := checkedNativeSize(index)
	if err != nil {
		return 0, false, err
	}
	var value C.double
	var present C.bool
	err = b.resource.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_broadcast_ephemeris_record_group_delay((*C.SidereonBroadcastEphemeris)(pointer), indexValue, C.uint32_t(term), &value, &present)
		})
	})
	return float64(value), bool(present), err
}

func (b *BroadcastEphemeris) RecordsInfo() ([]NativeBroadcastRecordInfo, error) {
	if b == nil || b.resource == nil {
		return nil, ErrClosed
	}
	var result []NativeBroadcastRecordInfo
	err := b.resource.with(func(pointer unsafe.Pointer) error {
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_broadcast_ephemeris_records((*C.SidereonBroadcastEphemeris)(pointer), nil, 0, &written, &required)
		}); err != nil {
			return err
		}
		count, err := validateNativeQuery("broadcast record info", uint64(written), uint64(required))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonBroadcastRecordInfo{})); err != nil {
			return err
		}
		var memory unsafe.Pointer
		if count != 0 {
			memory = C.calloc(C.size_t(count), C.size_t(unsafe.Sizeof(C.SidereonBroadcastRecordInfo{})))
			if memory == nil {
				return errors.New("sidereon: unable to allocate native record info")
			}
			defer C.free(memory)
		}
		written, required = 0, 0
		if err := callStatus(func() uint32 {
			return C.sidereon_broadcast_ephemeris_records((*C.SidereonBroadcastEphemeris)(pointer), (*C.SidereonBroadcastRecordInfo)(memory), C.size_t(count), &written, &required)
		}); err != nil {
			return err
		}
		actual, err := validateTwoPassCounts("broadcast record info", count, count, uint64(written), uint64(required))
		if err != nil {
			return err
		}
		rows := unsafe.Slice((*C.SidereonBroadcastRecordInfo)(memory), actual)
		result = make([]NativeBroadcastRecordInfo, actual)
		for i := range result {
			result[i] = recordInfoFromC(rows[i])
		}
		return nil
	})
	runtime.KeepAlive(b)
	return result, err
}

func (b *BroadcastEphemeris) SelectByIssue(satellite string, issue, message uint32, epoch float64) (NativeBroadcastRecordInfo, bool, error) {
	if b == nil || b.resource == nil {
		return NativeBroadcastRecordInfo{}, false, ErrClosed
	}
	var record C.SidereonBroadcastRecordInfo
	var present C.bool
	var result NativeBroadcastRecordInfo
	err := b.resource.with(func(pointer unsafe.Pointer) error {
		return withString(satellite, func(value *C.char) uint32 {
			return C.sidereon_broadcast_ephemeris_select_by_issue((*C.SidereonBroadcastEphemeris)(pointer), value, C.uint32_t(issue), C.uint32_t(message), C.double(epoch), &record, &present)
		})
	})
	if err == nil && bool(present) {
		result = recordInfoFromC(record)
	}
	runtime.KeepAlive(b)
	return result, bool(present), err
}

func (b *BroadcastEphemeris) SetNavMessagePreference(preference uint32) error {
	if b == nil || b.resource == nil {
		return ErrClosed
	}
	b.resource.mu.Lock()
	defer b.resource.mu.Unlock()
	if b.resource.ptr == nil {
		return ErrClosed
	}
	return callStatus(func() uint32 {
		return C.sidereon_broadcast_ephemeris_set_nav_message_preference((*C.SidereonBroadcastEphemeris)(b.resource.ptr), C.uint32_t(preference))
	})
}

func EccentricAnomaly(meanAnomaly, eccentricity float64) (float64, int, error) {
	var anomaly C.double
	var iterations C.size_t
	err := callStatus(func() uint32 {
		return C.sidereon_broadcast_eccentric_anomaly(C.double(meanAnomaly), C.double(eccentricity), &anomaly, &iterations)
	})
	if err != nil {
		return 0, 0, err
	}
	count, err := checkedNativeCount(uint64(iterations))
	return float64(anomaly), count, err
}

type NativeConstellationConstants struct{ GMM3PerS2, OmegaERadPerS, DTRF float64 }
type NativeClockOffset struct{ ClockPolynomialS, RelativisticS, GroupDelayS, TotalS float64 }
type NativeOrbitState struct {
	A, N0, N, TK, MK, EccentricAnomaly                                           float64
	KeplerIterations                                                             int
	SinE, CosE, Nu, Phi, S2, C2, DU, DR, DI, U, R, I, XP, YP, OmegaK, XM, YM, ZM float64
}
type NativeSatelliteState struct {
	Orbit NativeOrbitState
	Clock NativeClockOffset
}

func cConstellation(value NativeConstellationConstants) C.SidereonConstellationConstants {
	return C.SidereonConstellationConstants{gm_m3_s2: C.double(value.GMM3PerS2), omega_e_rad_s: C.double(value.OmegaERadPerS), dtr_f: C.double(value.DTRF)}
}

func cElements(value NativeKeplerianElements) C.SidereonKeplerianElements {
	return C.SidereonKeplerianElements{sqrt_a: C.double(value.SqrtA), e: C.double(value.E), m0: C.double(value.M0), delta_n: C.double(value.DeltaN), omega0: C.double(value.Omega0), i0: C.double(value.I0), omega: C.double(value.Omega), omega_dot: C.double(value.OmegaDot), idot: C.double(value.IDot), cuc: C.double(value.CUC), cus: C.double(value.CUS), crc: C.double(value.CRC), crs: C.double(value.CRS), cic: C.double(value.CIC), cis: C.double(value.CIS), toe_sow: C.double(value.ToeSOW)}
}

func cClock(value NativeClockPolynomial) C.SidereonClockPolynomial {
	return C.SidereonClockPolynomial{af0: C.double(value.AF0), af1: C.double(value.AF1), af2: C.double(value.AF2), toc_sow: C.double(value.TocSOW)}
}

func orbitStateFromC(value C.SidereonOrbitState) (NativeOrbitState, error) {
	iterations, err := checkedNativeCount(uint64(value.kepler_iterations))
	if err != nil {
		return NativeOrbitState{}, err
	}
	return NativeOrbitState{A: float64(value.a), N0: float64(value.n0), N: float64(value.n), TK: float64(value.tk), MK: float64(value.mk), EccentricAnomaly: float64(value.eccentric_anomaly), KeplerIterations: iterations, SinE: float64(value.sin_e), CosE: float64(value.cos_e), Nu: float64(value.nu), Phi: float64(value.phi), S2: float64(value.s2), C2: float64(value.c2), DU: float64(value.du), DR: float64(value.dr), DI: float64(value.di), U: float64(value.u), R: float64(value.r), I: float64(value.i), XP: float64(value.xp), YP: float64(value.yp), OmegaK: float64(value.omega_k), XM: float64(value.x_m), YM: float64(value.y_m), ZM: float64(value.z_m)}, nil
}

func clockOffsetFromC(value C.SidereonClockOffset) NativeClockOffset {
	return NativeClockOffset{ClockPolynomialS: float64(value.dt_clock_poly_s), RelativisticS: float64(value.dt_rel_s), GroupDelayS: float64(value.tgd_s), TotalS: float64(value.dt_clock_total_s)}
}

func BroadcastSatelliteClockOffset(clock NativeClockPolynomial, constants NativeConstellationConstants, elements NativeKeplerianElements, sinE, tSOW, tGD float64) (NativeClockOffset, error) {
	cClockValue, cConstants, cElements := cClock(clock), cConstellation(constants), cElements(elements)
	var result C.SidereonClockOffset
	err := callStatus(func() uint32 {
		return C.sidereon_broadcast_satellite_clock_offset_s(&cClockValue, &cConstants, &cElements, C.double(sinE), C.double(tSOW), C.double(tGD), &result)
	})
	if err != nil {
		return NativeClockOffset{}, err
	}
	return clockOffsetFromC(result), nil
}

func BroadcastSatellitePosition(elements NativeKeplerianElements, constants NativeConstellationConstants, tSOW float64, isGeo bool) (NativeOrbitState, error) {
	cElementsValue, cConstants := cElements(elements), cConstellation(constants)
	var result C.SidereonOrbitState
	err := callStatus(func() uint32 {
		return C.sidereon_broadcast_satellite_position_ecef(&cElementsValue, &cConstants, C.double(tSOW), C.bool(isGeo), &result)
	})
	if err != nil {
		return NativeOrbitState{}, err
	}
	return orbitStateFromC(result)
}

func BroadcastSatelliteState(elements NativeKeplerianElements, clock NativeClockPolynomial, constants NativeConstellationConstants, tSOW, tGD float64, isGeo bool) (NativeSatelliteState, error) {
	cElementsValue, cClockValue, cConstants := cElements(elements), cClock(clock), cConstellation(constants)
	var result C.SidereonSatelliteState
	err := callStatus(func() uint32 {
		return C.sidereon_broadcast_satellite_state(&cElementsValue, &cClockValue, &cConstants, C.double(tSOW), C.double(tGD), C.bool(isGeo), &result)
	})
	if err != nil {
		return NativeSatelliteState{}, err
	}
	orbit, err := orbitStateFromC(result.orbit)
	if err != nil {
		return NativeSatelliteState{}, err
	}
	return NativeSatelliteState{Orbit: orbit, Clock: clockOffsetFromC(result.clock)}, nil
}

func validateSelectionStatus(status uint32) error {
	if status > uint32(C.SIDEREON_SELECTION_STATUS_OVERFLOW) {
		return invalidArgument("invalid selection status returned by native code")
	}
	return nil
}

func selectionStatusErrorLocked(status uint32) error {
	if status == uint32(C.SIDEREON_SELECTION_STATUS_OK) {
		return nil
	}
	names := [...]string{"ok", "null pointer", "invalid argument", "invalid token", "panic", "empty product set", "invalid range", "no prior product", "beyond staleness cap", "invalid product", "invalid policy", "overflow"}
	text := "selection status"
	if status < uint32(len(names)) {
		text = names[status]
	}
	required := C.sidereon_last_error_message(nil, 0)
	detail := ""
	if required != 0 {
		count, err := checkedNativeCount(uint64(required))
		if err != nil {
			return &SelectionError{Status: status, Text: text, Detail: err.Error()}
		}
		if count == int(^uint(0)>>1) {
			return &SelectionError{Status: status, Text: text, Detail: "sidereon: native error detail is too large"}
		}
		if _, err := checkedNativeAllocationSize(count+1, 1); err != nil {
			return &SelectionError{Status: status, Text: text, Detail: err.Error()}
		}
		buffer := make([]byte, count+1)
		length, err := cSize(len(buffer), "selection error detail output capacity")
		if err != nil {
			return &SelectionError{Status: status, Text: text, Detail: err.Error()}
		}
		written := C.sidereon_last_error_message((*C.char)(unsafe.Pointer(&buffer[0])), length)
		writtenCount, err := validateTwoPassCounts("selection error detail", count, count, uint64(written), uint64(required))
		if err != nil {
			return &SelectionError{Status: status, Text: text, Detail: err.Error()}
		}
		detail = string(buffer[:writtenCount])
	}
	return &SelectionError{Status: status, Text: text, Detail: detail}
}

func validateDegradationKind(value uint32) error {
	if value > uint32(C.SIDEREON_DEGRADATION_KIND_DIURNAL_SHIFT) {
		return invalidArgument("invalid staleness degradation kind returned by native code")
	}
	return nil
}

func stalenessMetadataFromC(value C.SidereonStalenessMetadata) (NativeStalenessMetadata, error) {
	metadata := NativeStalenessMetadata{Kind: uint32(value.kind), RequestedEpochJ2000S: float64(value.requested_epoch_j2000_s), SourceEpochJ2000S: float64(value.source_epoch_j2000_s), StalenessS: float64(value.staleness_s), StalenessDays: float64(value.staleness_days)}
	if err := validateDegradationKind(metadata.Kind); err != nil {
		return NativeStalenessMetadata{}, err
	}
	return metadata, nil
}

func selectSP3(products []*SP3, start, end float64, policy StalenessPolicy, overRange bool) (*SP3, NativeStalenessMetadata, error) {
	handles := make([]*positioningHandle, len(products))
	for i, product := range products {
		if product == nil || product.handle == nil {
			return nil, NativeStalenessMetadata{}, ErrClosed
		}
		handles[i] = product.handle
	}
	productCount, err := checkedNativeSize(len(products))
	if err != nil {
		return nil, NativeStalenessMetadata{}, err
	}
	var result *SP3
	var metadata NativeStalenessMetadata
	err = withPositioningHandleSet(handles, func(pointers []unsafe.Pointer) error {
		return withCThreadError(func() error {
			arraySize, err := checkedNativeAllocationSize(len(pointers), unsafe.Sizeof(uintptr(0)))
			if err != nil {
				return err
			}
			var memory unsafe.Pointer
			if arraySize != 0 {
				memory = C.calloc(1, C.size_t(arraySize))
				if memory == nil {
					return errors.New("sidereon: unable to allocate native SP3 selection array")
				}
				defer C.free(memory)
			}
			array := unsafe.Slice((**C.SidereonSp3)(memory), len(pointers))
			for i, pointer := range pointers {
				array[i] = (*C.SidereonSp3)(pointer)
			}
			var selected *C.SidereonSp3
			var cMetadata C.SidereonStalenessMetadata
			var status C.SidereonSelectionStatus
			if !overRange {
				status = C.SidereonSelectionStatus(C.sidereon_select_sp3((**C.SidereonSp3)(memory), productCount, C.double(start), C.SidereonStalenessPolicy{max_staleness_s: C.double(policy.MaxStalenessS)}, &selected, &cMetadata))
			} else {
				status = C.SidereonSelectionStatus(C.sidereon_select_sp3_over_range((**C.SidereonSp3)(memory), productCount, C.double(start), C.double(end), C.SidereonStalenessPolicy{max_staleness_s: C.double(policy.MaxStalenessS)}, &selected, &cMetadata))
			}
			if err := validateSelectionStatus(uint32(status)); err != nil {
				if selected != nil {
					C.sidereon_sp3_free(selected)
				}
				return err
			}
			if status != C.SIDEREON_SELECTION_STATUS_OK {
				err := selectionStatusErrorLocked(uint32(status))
				if selected != nil {
					C.sidereon_sp3_free(selected)
				}
				return err
			}
			if selected == nil {
				return errors.New("sidereon: native SP3 selection returned no handle")
			}
			converted, err := stalenessMetadataFromC(cMetadata)
			if err != nil {
				C.sidereon_sp3_free(selected)
				return err
			}
			result, err = newSP3FromPointer(selected)
			if err != nil {
				C.sidereon_sp3_free(selected)
				return err
			}
			metadata = converted
			return nil
		})
	})
	for _, product := range products {
		runtime.KeepAlive(product)
	}
	if err != nil {
		return nil, NativeStalenessMetadata{}, err
	}
	return result, metadata, nil
}

func newSP3FromPointer(pointer *C.SidereonSp3) (*SP3, error) {
	if pointer == nil {
		return nil, missingNativeHandle("SP3 selection")
	}
	return &SP3{handle: newPositioningHandle(unsafe.Pointer(pointer), releaseSP3)}, nil
}

func SelectSP3(products []*SP3, requested float64, policy StalenessPolicy) (*SP3, NativeStalenessMetadata, error) {
	return selectSP3(products, requested, requested, policy, false)
}

func SelectSP3OverRange(products []*SP3, start, end float64, policy StalenessPolicy) (*SP3, NativeStalenessMetadata, error) {
	return selectSP3(products, start, end, policy, true)
}

func (b *BroadcastEphemeris) SolveBroadcast(config SPPConfig) (SPPSolution, error) {
	if b == nil || b.resource == nil {
		return SPPSolution{}, ErrClosed
	}
	var result SPPSolution
	err := b.resource.with(func(pointer unsafe.Pointer) error {
		return withSPPInputs(config, func(inputs *C.SidereonSppInputs) error {
			var solution *C.SidereonSppSolution
			status := C.sidereon_solve_broadcast((*C.SidereonBroadcastEphemeris)(pointer), inputs, &solution)
			if err := statusErrorLocked(uint32(status)); err != nil {
				if solution != nil {
					C.sidereon_spp_solution_free(solution)
				}
				return err
			}
			if solution == nil {
				return errors.New("sidereon: native broadcast solve returned no solution")
			}
			defer C.sidereon_spp_solution_free(solution)
			var err error
			result, err = readSPPSolutionLocked(solution)
			return err
		})
	})
	runtime.KeepAlive(b)
	return result, err
}

type NativeSppDopplerObservation struct {
	SatelliteID                                 string
	DopplerHz, CarrierHz, SatelliteClockDriftSS float64
}

// NativeSppDopplerVelocity is a detached copy of the related native velocity
// solution. Residuals are metres per second; the covariance is the native
// row-major unit-variance 4x4 matrix.
type NativeSppDopplerVelocity struct {
	VelocityMPerS      [3]float64
	ClockDriftSPerS    float64
	SpeedMPerS         float64
	StateCovariance    [16]float64
	UsedSatelliteCount int
	UsedSatelliteIDs   []string
	ResidualsMPerS     []float64
}

type NativeSppDopplerResult struct {
	Receiver          SPPSolution
	HasVelocity       bool
	VelocityErrorKind uint32
	Velocity          *NativeSppDopplerVelocity
}

func readVelocitySolutionLocked(solution *C.SidereonVelocitySolution) (NativeSppDopplerVelocity, error) {
	if solution == nil {
		return NativeSppDopplerVelocity{}, errors.New("sidereon: native velocity solution is nil")
	}
	var result NativeSppDopplerVelocity
	var velocity [3]C.double
	if err := statusErrorLocked(C.sidereon_velocity_solution_velocity(solution, &velocity[0], C.size_t(len(velocity)))); err != nil {
		return NativeSppDopplerVelocity{}, err
	}
	for index := range result.VelocityMPerS {
		result.VelocityMPerS[index] = float64(velocity[index])
	}
	var clockDrift C.double
	if err := statusErrorLocked(C.sidereon_velocity_solution_clock_drift(solution, &clockDrift)); err != nil {
		return NativeSppDopplerVelocity{}, err
	}
	result.ClockDriftSPerS = float64(clockDrift)
	var speed C.double
	if err := statusErrorLocked(C.sidereon_velocity_solution_speed(solution, &speed)); err != nil {
		return NativeSppDopplerVelocity{}, err
	}
	result.SpeedMPerS = float64(speed)
	var covariance [16]C.double
	if err := statusErrorLocked(C.sidereon_velocity_solution_state_covariance(solution, &covariance[0], C.size_t(len(covariance)))); err != nil {
		return NativeSppDopplerVelocity{}, err
	}
	for index := range result.StateCovariance {
		result.StateCovariance[index] = float64(covariance[index])
	}
	var count C.size_t
	if err := statusErrorLocked(C.sidereon_velocity_solution_used_sat_count(solution, &count)); err != nil {
		return NativeSppDopplerVelocity{}, err
	}
	usedCount, err := checkedNativeCount(uint64(count))
	if err != nil {
		return NativeSppDopplerVelocity{}, err
	}
	result.UsedSatelliteCount = usedCount
	var written, required C.size_t
	if err := statusErrorLocked(C.sidereon_velocity_solution_used_sat_ids(solution, nil, 0, &written, &required)); err != nil {
		return NativeSppDopplerVelocity{}, err
	}
	idCount, err := validateNativeQuery("velocity used satellite IDs", uint64(written), uint64(required))
	if err != nil {
		return NativeSppDopplerVelocity{}, err
	}
	if idCount != usedCount {
		return NativeSppDopplerVelocity{}, fmt.Errorf("sidereon: velocity used satellite count %d differs from ID count %d", usedCount, idCount)
	}
	if _, err := checkedNativeAllocationSize(idCount, unsafe.Sizeof(C.SidereonSatelliteToken{})); err != nil {
		return NativeSppDopplerVelocity{}, err
	}
	ids := make([]C.SidereonSatelliteToken, idCount)
	idLength, err := checkedNativeSize(len(ids))
	if err != nil {
		return NativeSppDopplerVelocity{}, err
	}
	var idOutput *C.SidereonSatelliteToken
	if len(ids) != 0 {
		idOutput = &ids[0]
	}
	written, required = 0, 0
	if err := statusErrorLocked(C.sidereon_velocity_solution_used_sat_ids(solution, idOutput, idLength, &written, &required)); err != nil {
		return NativeSppDopplerVelocity{}, err
	}
	actualIDs, err := validateTwoPassCounts("velocity used satellite IDs", len(ids), idCount, uint64(written), uint64(required))
	if err != nil {
		return NativeSppDopplerVelocity{}, err
	}
	result.UsedSatelliteIDs = make([]string, actualIDs)
	for index := range result.UsedSatelliteIDs {
		result.UsedSatelliteIDs[index] = tokenFromC(ids[index])
	}
	written, required = 0, 0
	if err := statusErrorLocked(C.sidereon_velocity_solution_residuals(solution, nil, 0, &written, &required)); err != nil {
		return NativeSppDopplerVelocity{}, err
	}
	residualCount, err := validateNativeQuery("velocity residuals", uint64(written), uint64(required))
	if err != nil {
		return NativeSppDopplerVelocity{}, err
	}
	if _, err := checkedNativeAllocationSize(residualCount, unsafe.Sizeof(C.double(0))); err != nil {
		return NativeSppDopplerVelocity{}, err
	}
	residuals := make([]C.double, residualCount)
	residualLength, err := checkedNativeSize(len(residuals))
	if err != nil {
		return NativeSppDopplerVelocity{}, err
	}
	var residualOutput *C.double
	if len(residuals) != 0 {
		residualOutput = &residuals[0]
	}
	written, required = 0, 0
	if err := statusErrorLocked(C.sidereon_velocity_solution_residuals(solution, residualOutput, residualLength, &written, &required)); err != nil {
		return NativeSppDopplerVelocity{}, err
	}
	actualResiduals, err := validateTwoPassCounts("velocity residuals", len(residuals), residualCount, uint64(written), uint64(required))
	if err != nil {
		return NativeSppDopplerVelocity{}, err
	}
	result.ResidualsMPerS = make([]float64, actualResiduals)
	for index := range result.ResidualsMPerS {
		result.ResidualsMPerS[index] = float64(residuals[index])
	}
	return result, nil
}

func (b *BroadcastEphemeris) SolveBroadcastWithDopplerVelocity(config SPPConfig, dopplers []NativeSppDopplerObservation) (NativeSppDopplerResult, error) {
	if b == nil || b.resource == nil {
		return NativeSppDopplerResult{}, ErrClosed
	}
	for _, observation := range dopplers {
		if err := validateSPPSatelliteID(observation.SatelliteID); err != nil {
			return NativeSppDopplerResult{}, err
		}
	}
	var result NativeSppDopplerResult
	err := b.resource.with(func(pointer unsafe.Pointer) error {
		return withSPPInputs(config, func(inputs *C.SidereonSppInputs) error {
			v2Memory := C.calloc(1, C.size_t(unsafe.Sizeof(C.SidereonSppInputsV2{})))
			if v2Memory == nil {
				return errors.New("sidereon: unable to allocate native SPP V2 inputs")
			}
			defer C.free(v2Memory)
			v2 := (*C.SidereonSppInputsV2)(v2Memory)
			v2.base = *inputs
			var memory unsafe.Pointer
			if len(dopplers) != 0 {
				size, err := checkedNativeAllocationSize(len(dopplers), unsafe.Sizeof(C.SidereonSppDopplerObservation{}))
				if err != nil {
					return err
				}
				memory = C.calloc(1, C.size_t(size))
				if memory == nil {
					return errors.New("sidereon: unable to allocate native Doppler observations")
				}
				defer C.free(memory)
			}
			rows := unsafe.Slice((*C.SidereonSppDopplerObservation)(memory), len(dopplers))
			ids := make([]unsafe.Pointer, 0, len(dopplers))
			defer func() {
				for _, id := range ids {
					C.free(id)
				}
			}()
			for i, observation := range dopplers {
				id := C.CString(observation.SatelliteID)
				if id == nil {
					return errors.New("sidereon: unable to allocate native Doppler satellite ID")
				}
				ids = append(ids, unsafe.Pointer(id))
				rows[i].sat_id = id
				rows[i].doppler_hz, rows[i].carrier_hz, rows[i].sat_clock_drift_s_s = C.double(observation.DopplerHz), C.double(observation.CarrierHz), C.double(observation.SatelliteClockDriftSS)
			}
			count, err := checkedNativeSize(len(dopplers))
			if err != nil {
				return err
			}
			var solution *C.SidereonSppDopplerSolution
			status := C.sidereon_solve_broadcast_with_doppler_velocity((*C.SidereonBroadcastEphemeris)(pointer), v2, (*C.SidereonSppDopplerObservation)(memory), count, &solution)
			if err := statusErrorLocked(uint32(status)); err != nil {
				if solution != nil {
					C.sidereon_spp_doppler_solution_free(solution)
				}
				return err
			}
			if solution == nil {
				return errors.New("sidereon: native broadcast Doppler solve returned no solution")
			}
			defer C.sidereon_spp_doppler_solution_free(solution)
			var hasVelocity C.bool
			if err := statusErrorLocked(uint32(C.sidereon_spp_doppler_solution_has_velocity(solution, &hasVelocity))); err != nil {
				return err
			}
			result.HasVelocity = bool(hasVelocity)
			var errorKind uint32
			if err := statusErrorLocked(uint32(C.sidereon_spp_doppler_solution_velocity_error_kind(solution, &errorKind))); err != nil {
				return err
			}
			if uint32(errorKind) > uint32(C.SIDEREON_SPP_DOPPLER_VELOCITY_ERROR_KIND_INVALID_RECEIVER_STATE) {
				return invalidArgument("invalid Doppler velocity error kind returned by native code")
			}
			result.VelocityErrorKind = uint32(errorKind)
			var receiver *C.SidereonSppSolution
			if err := statusErrorLocked(uint32(C.sidereon_spp_doppler_solution_receiver(solution, &receiver))); err != nil {
				if receiver != nil {
					C.sidereon_spp_solution_free(receiver)
				}
				return err
			}
			if receiver == nil {
				return errors.New("sidereon: native broadcast Doppler solve returned no receiver solution")
			}
			defer C.sidereon_spp_solution_free(receiver)
			result.Receiver, err = readSPPSolutionLocked(receiver)
			if err != nil {
				return err
			}
			var velocity *C.SidereonVelocitySolution
			velocityStatus := C.sidereon_spp_doppler_solution_velocity(solution, &velocity)
			if velocityStatus != C.SIDEREON_STATUS_OK {
				if velocity != nil {
					C.sidereon_velocity_solution_free(velocity)
				}
				if velocityStatus == C.SIDEREON_STATUS_SOLVE {
					return nil
				}
				return statusErrorLocked(uint32(velocityStatus))
			}
			if velocity == nil {
				return errors.New("sidereon: native broadcast Doppler solve returned no velocity solution")
			}
			defer C.sidereon_velocity_solution_free(velocity)
			value, readErr := readVelocitySolutionLocked(velocity)
			if readErr != nil {
				return readErr
			}
			result.Velocity = &value
			return nil
		})
	})
	runtime.KeepAlive(b)
	runtime.KeepAlive(dopplers)
	return result, err
}
