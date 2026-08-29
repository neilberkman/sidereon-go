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

// SppInputsV2 is a Go-owned description of the V2 input graph. The C graph is
// materialized by each call and is never retained by C after that call.
type SppInputsV2 struct {
	Base            SPPConfig
	BeidouEnabled   bool
	BeidouAlpha     [4]float64
	BeidouBeta      [4]float64
	RobustEnabled   bool
	Robust          NativeSPPRobustConfig
	Policy          NativeSPPSolvePolicy
	GlonassChannels []NativeGlonassChannel
}

type NativeGlonassChannel struct {
	Slot    uint8
	Channel int8
}

type NativeSPPValidationOptions struct {
	MaxPDOPEnabled                                                              bool
	MaxPDOP, MinPlausibleRadiusM, MaxPlausibleRadiusM, MaxConvergedResidualRMSM float64
}

type NativeSPPSolvePolicy struct {
	UseValidationOptions bool
	Validation           NativeSPPValidationOptions
	CoarseSearchEnabled  bool
	CoarseSearchSeeds    int
}

type NativeRinexSPPOptions struct {
	Ionosphere, Troposphere, InitialGuessEnabled bool
	InitialGuess                                 [4]float64
	PressureHPA, TemperatureK, RelativeHumidity  float64
	RobustEnabled                                bool
	Robust                                       NativeSPPRobustConfig
}

type NativeRinexSPPEpoch struct {
	Index, ObservationCount int
	Epoch                   NativeCalendarEpoch
}

type NativeSPPRejectedSatellite struct {
	SatelliteID string
	Reason      uint32
}

type NativeSPPSystemClock struct {
	System    uint32
	ReceiverS float64
}

type NativeSPPSystemTDOP struct {
	System uint32
	TDOP   float64
}

type SppSolutionHandle struct {
	_      noCopy
	handle *surfaceHandle
}

type RinexSPPInputs struct {
	_      noCopy
	handle *surfaceHandle
}

type RinexSPPSolutions struct {
	_      noCopy
	handle *surfaceHandle
}

type SPPBatch struct {
	_      noCopy
	handle *surfaceHandle
}

// FallbackError retains the distinct status namespace returned by the
// fallback ABI. Detail is copied from the same-thread native error slot.
type FallbackError struct {
	Status uint32
	Detail string
}

func (e *FallbackError) Error() string {
	if e == nil {
		return "sidereon: fallback solve failed"
	}
	if e.Detail == "" {
		return fmt.Sprintf("sidereon: fallback solve failed (%d)", e.Status)
	}
	return fmt.Sprintf("sidereon: fallback solve failed (%d): %s", e.Status, e.Detail)
}

func fallbackStatusErrorLocked(status C.enum_SidereonFallbackStatus) error {
	if status == C.SIDEREON_FALLBACK_STATUS_OK {
		return nil
	}
	if uint32(status) > uint32(C.SIDEREON_FALLBACK_STATUS_BROADCAST_SOLVE) {
		return invalidArgument("invalid fallback status returned by native code")
	}
	required := C.sidereon_last_error_message(nil, 0)
	var detail string
	if required > 0 {
		n, err := sizeTToInt(required, "fallback error detail")
		if err != nil {
			return err
		}
		if n == int(^uint(0)>>1) {
			return errors.New("sidereon: fallback error detail is too large")
		}
		if _, err := checkedNativeAllocationSize(n+1, 1); err != nil {
			return err
		}
		buffer := make([]byte, n+1)
		written := C.sidereon_last_error_message((*C.char)(unsafe.Pointer(&buffer[0])), C.size_t(len(buffer)))
		count, err := validateTwoPassCounts("fallback error detail", n, n, uint64(written), uint64(required))
		if err != nil {
			return err
		}
		detail = string(buffer[:count])
	}
	return &FallbackError{Status: uint32(status), Detail: detail}
}

func newSppHandle(pointer *C.SidereonSppSolution) (*SppSolutionHandle, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	h, err := newSurfaceHandle(unsafe.Pointer(pointer), func(p unsafe.Pointer) {
		C.sidereon_spp_solution_free((*C.SidereonSppSolution)(p))
	})
	if err != nil {
		return nil, err
	}
	return &SppSolutionHandle{handle: h}, nil
}

func newRinexSPPInputs(pointer *C.SidereonRinexSppInputs) (*RinexSPPInputs, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	h, err := newSurfaceHandle(unsafe.Pointer(pointer), func(p unsafe.Pointer) {
		C.sidereon_rinex_spp_inputs_free((*C.SidereonRinexSppInputs)(p))
	})
	if err != nil {
		return nil, err
	}
	return &RinexSPPInputs{handle: h}, nil
}

func newRinexSPPSolutions(pointer *C.SidereonRinexSppSolutions) (*RinexSPPSolutions, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	h, err := newSurfaceHandle(unsafe.Pointer(pointer), func(p unsafe.Pointer) {
		C.sidereon_rinex_spp_solutions_free((*C.SidereonRinexSppSolutions)(p))
	})
	if err != nil {
		return nil, err
	}
	return &RinexSPPSolutions{handle: h}, nil
}

func newSppBatch(pointer *C.SidereonSppBatch) (*SPPBatch, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	h, err := newSurfaceHandle(unsafe.Pointer(pointer), func(p unsafe.Pointer) {
		C.sidereon_spp_batch_free((*C.SidereonSppBatch)(p))
	})
	if err != nil {
		return nil, err
	}
	return &SPPBatch{handle: h}, nil
}

func nativeRobust(value NativeSPPRobustConfig, dst *C.SidereonSppRobustConfig) error {
	maxOuter, err := cSize64(value.MaxOuter, "SPP robust max outer")
	if err != nil {
		return fmt.Errorf("sidereon: SPP robust max outer: %w", err)
	}
	*dst = C.SidereonSppRobustConfig{huber_k: C.double(value.HuberK), scale_floor_m: C.double(value.ScaleFloorM), max_outer: maxOuter, outer_tol_m: C.double(value.OuterToleranceM)}
	return nil
}

func nativePolicy(value NativeSPPSolvePolicy, dst *C.SidereonSppSolvePolicy) error {
	seeds, err := checkedNativeSize(value.CoarseSearchSeeds)
	if err != nil {
		return fmt.Errorf("sidereon: SPP coarse-search seeds: %w", err)
	}
	*dst = C.SidereonSppSolvePolicy{use_validation_options: C.bool(value.UseValidationOptions), coarse_search_enabled: C.bool(value.CoarseSearchEnabled), coarse_search_seeds: seeds}
	dst.validation = C.SidereonSppValidationOptions{max_pdop_enabled: C.bool(value.Validation.MaxPDOPEnabled), max_pdop: C.double(value.Validation.MaxPDOP), min_plausible_radius_m: C.double(value.Validation.MinPlausibleRadiusM), max_plausible_radius_m: C.double(value.Validation.MaxPlausibleRadiusM), max_converged_residual_rms_m: C.double(value.Validation.MaxConvergedResidualRMSM)}
	return nil
}

func fillSppBase(dst *C.SidereonSppInputs, value SPPConfig, alloc *cRtkAlloc) error {
	count, err := checkedNativeSize(len(value.Observations))
	if err != nil {
		return err
	}
	for _, row := range value.Observations {
		if err := validateSPPSatelliteID(row.SatelliteID); err != nil {
			return err
		}
	}
	var observations *C.SidereonObservation
	if len(value.Observations) != 0 {
		bytes, err := checkedNativeAllocationSize(len(value.Observations), unsafe.Sizeof(C.SidereonObservation{}))
		if err != nil {
			return err
		}
		memory, err := alloc.malloc(bytes, "SPP observations")
		if err != nil {
			return err
		}
		rows := unsafe.Slice((*C.SidereonObservation)(memory), len(value.Observations))
		for i, row := range value.Observations {
			id, err := alloc.cstring(row.SatelliteID, "SPP satellite ID")
			if err != nil {
				return err
			}
			rows[i].sat_id = id
			rows[i].pseudorange_m = C.double(row.PseudorangeM)
		}
		observations = &rows[0]
	}
	*dst = C.SidereonSppInputs{observations: observations, observation_count: count, t_rx_j2000_s: C.double(value.TRxJ2000S), t_rx_second_of_day_s: C.double(value.TRxSecondOfDayS), day_of_year: C.double(value.DayOfYear), ionosphere: C.bool(value.Ionosphere), troposphere: C.bool(value.Troposphere), with_geodetic: C.bool(value.WithGeodetic)}
	for i := 0; i < 4; i++ {
		dst.initial_guess[i] = C.double(value.InitialGuess[i])
		dst.klobuchar_alpha[i] = C.double(value.KlobucharAlpha[i])
		dst.klobuchar_beta[i] = C.double(value.KlobucharBeta[i])
	}
	dst.pressure_hpa = C.double(value.PressureHPA)
	dst.temperature_k = C.double(value.TemperatureK)
	dst.relative_humidity = C.double(value.RelativeHumidity)
	return nil
}

func fillSppV2(dst *C.SidereonSppInputsV2, value SppInputsV2, alloc *cRtkAlloc) error {
	if err := nativePolicy(value.Policy, &dst.policy); err != nil {
		return err
	}
	if err := nativeRobust(value.Robust, &dst.robust); err != nil {
		return err
	}
	if err := fillSppBase(&dst.base, value.Base, alloc); err != nil {
		return err
	}
	dst.beidou_klobuchar_enabled, dst.robust_enabled = C.bool(value.BeidouEnabled), C.bool(value.RobustEnabled)
	for i := 0; i < 4; i++ {
		dst.beidou_klobuchar_alpha[i], dst.beidou_klobuchar_beta[i] = C.double(value.BeidouAlpha[i]), C.double(value.BeidouBeta[i])
	}
	if len(value.GlonassChannels) != 0 {
		bytes, err := checkedNativeAllocationSize(len(value.GlonassChannels), unsafe.Sizeof(C.SidereonGlonassChannel{}))
		if err != nil {
			return err
		}
		memory, err := alloc.malloc(bytes, "GLONASS channels")
		if err != nil {
			return err
		}
		rows := unsafe.Slice((*C.SidereonGlonassChannel)(memory), len(value.GlonassChannels))
		for i, row := range value.GlonassChannels {
			rows[i] = C.SidereonGlonassChannel{slot: C.uint8_t(row.Slot), channel: C.int8_t(row.Channel)}
		}
		dst.glonass_channels = &rows[0]
	}
	count, err := checkedNativeSize(len(value.GlonassChannels))
	if err != nil {
		return err
	}
	dst.glonass_channel_count = count
	return nil
}

func makeSppV2(value SppInputsV2, alloc *cRtkAlloc) (*C.SidereonSppInputsV2, error) {
	dst, err := makeDgnssBaseInputs(value.Base, alloc)
	if err != nil {
		return nil, err
	}
	return dst, fillSppV2(dst, value, alloc)
}

func makeSppV2Array(values []SppInputsV2, alloc *cRtkAlloc) (*C.SidereonSppInputsV2, C.size_t, error) {
	count, err := checkedNativeSize(len(values))
	if err != nil {
		return nil, 0, err
	}
	if len(values) == 0 {
		return nil, count, nil
	}
	bytes, err := checkedNativeAllocationSize(len(values), unsafe.Sizeof(C.SidereonSppInputsV2{}))
	if err != nil {
		return nil, 0, err
	}
	memory, err := alloc.malloc(bytes, "SPP V2 input array")
	if err != nil {
		return nil, 0, err
	}
	rows := unsafe.Slice((*C.SidereonSppInputsV2)(memory), len(values))
	defaults, err := makeDgnssBaseInputs(SPPConfig{}, alloc)
	if err != nil {
		return nil, 0, err
	}
	for i := range values {
		rows[i] = *defaults
		if err := fillSppV2(&rows[i], values[i], alloc); err != nil {
			return nil, 0, err
		}
	}
	return &rows[0], count, nil
}

func makeRinexOptions(value NativeRinexSPPOptions, alloc *cRtkAlloc) (*C.SidereonRinexSppOptions, error) {
	bytes, err := checkedNativeAllocationSize(1, unsafe.Sizeof(C.SidereonRinexSppOptions{}))
	if err != nil {
		return nil, err
	}
	memory, err := alloc.malloc(bytes, "RINEX SPP options")
	if err != nil {
		return nil, err
	}
	dst := (*C.SidereonRinexSppOptions)(memory)
	*dst = C.SidereonRinexSppOptions{ionosphere: C.bool(value.Ionosphere), troposphere: C.bool(value.Troposphere), initial_guess_enabled: C.bool(value.InitialGuessEnabled), pressure_hpa: C.double(value.PressureHPA), temperature_k: C.double(value.TemperatureK), relative_humidity: C.double(value.RelativeHumidity), robust_enabled: C.bool(value.RobustEnabled)}
	for i := 0; i < 4; i++ {
		dst.initial_guess[i] = C.double(value.InitialGuess[i])
	}
	if err := nativeRobust(value.Robust, &dst.robust); err != nil {
		return nil, err
	}
	return dst, nil
}

func makeDopplerRows(values []NativeSppDopplerObservation, alloc *cRtkAlloc) (*C.SidereonSppDopplerObservation, C.size_t, error) {
	count, err := checkedNativeSize(len(values))
	if err != nil {
		return nil, 0, err
	}
	if len(values) == 0 {
		return nil, count, nil
	}
	bytes, err := checkedNativeAllocationSize(len(values), unsafe.Sizeof(C.SidereonSppDopplerObservation{}))
	if err != nil {
		return nil, 0, err
	}
	memory, err := alloc.malloc(bytes, "SPP Doppler observations")
	if err != nil {
		return nil, 0, err
	}
	rows := unsafe.Slice((*C.SidereonSppDopplerObservation)(memory), len(values))
	for i, value := range values {
		if err := validateSPPSatelliteID(value.SatelliteID); err != nil {
			return nil, 0, err
		}
		id, err := alloc.cstring(value.SatelliteID, "SPP Doppler satellite ID")
		if err != nil {
			return nil, 0, err
		}
		rows[i] = C.SidereonSppDopplerObservation{sat_id: id, doppler_hz: C.double(value.DopplerHz), carrier_hz: C.double(value.CarrierHz), sat_clock_drift_s_s: C.double(value.SatelliteClockDriftSS)}
	}
	return &rows[0], count, nil
}

func makeLegacySpp(value SPPConfig, alloc *cRtkAlloc) (*C.SidereonSppInputs, error) {
	bytes, err := checkedNativeAllocationSize(1, unsafe.Sizeof(C.SidereonSppInputs{}))
	if err != nil {
		return nil, err
	}
	memory, err := alloc.malloc(bytes, "SPP inputs")
	if err != nil {
		return nil, err
	}
	dst := (*C.SidereonSppInputs)(memory)
	return dst, fillSppBase(dst, value, alloc)
}

func statusCall(fn func() C.enum_SidereonStatus) error {
	var err error
	withCThread(func() { err = statusErrorLocked(uint32(fn())) })
	return err
}

func (s *SppSolutionHandle) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return s.handle.close()
}
func (r *RinexSPPInputs) Close() error {
	if r == nil || r.handle == nil {
		return nil
	}
	return r.handle.close()
}
func (r *RinexSPPSolutions) Close() error {
	if r == nil || r.handle == nil {
		return nil
	}
	return r.handle.close()
}
func (b *SPPBatch) Close() error {
	if b == nil || b.handle == nil {
		return nil
	}
	return b.handle.close()
}
func (r *RinexSPPInputs) Count() (int, error) {
	if r == nil || r.handle == nil {
		return 0, ErrClosed
	}
	var out C.size_t
	var op error
	err := r.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			op = statusErrorLocked(uint32(C.sidereon_rinex_spp_inputs_count((*C.SidereonRinexSppInputs)(p), &out)))
		})
		return op
	})
	if err != nil {
		return 0, err
	}
	return sizeTToInt(out, "RINEX SPP input count")
}

func rinexEpochFromC(value C.SidereonRinexSppEpoch) (NativeRinexSPPEpoch, error) {
	count, err := sizeTToInt(value.observation_count, "RINEX SPP observation count")
	if err != nil {
		return NativeRinexSPPEpoch{}, err
	}
	index, err := sizeTToInt(value.epoch_index, "RINEX SPP epoch index")
	if err != nil {
		return NativeRinexSPPEpoch{}, err
	}
	return NativeRinexSPPEpoch{Index: index, ObservationCount: count, Epoch: calendarEpochFromC(value.epoch)}, nil
}

func (r *RinexSPPInputs) Epoch(index int) (NativeRinexSPPEpoch, error) {
	if r == nil || r.handle == nil {
		return NativeRinexSPPEpoch{}, ErrClosed
	}
	idx, err := cSize(index, "RINEX SPP epoch index")
	if err != nil {
		return NativeRinexSPPEpoch{}, err
	}
	var out C.SidereonRinexSppEpoch
	var op error
	err = r.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			op = statusErrorLocked(uint32(C.sidereon_rinex_spp_inputs_epoch((*C.SidereonRinexSppInputs)(p), idx, &out)))
		})
		return op
	})
	if err != nil {
		return NativeRinexSPPEpoch{}, err
	}
	return rinexEpochFromC(out)
}

func sppBaseFromC(value C.SidereonSppInputs) (SPPConfig, error) {
	count, err := sizeTToInt(value.observation_count, "SPP observation count")
	if err != nil {
		return SPPConfig{}, err
	}
	if count > 0 && value.observations == nil {
		return SPPConfig{}, errors.New("sidereon: native SPP observations pointer is null")
	}
	out := SPPConfig{TRxJ2000S: float64(value.t_rx_j2000_s), TRxSecondOfDayS: float64(value.t_rx_second_of_day_s), DayOfYear: float64(value.day_of_year), Ionosphere: bool(value.ionosphere), Troposphere: bool(value.troposphere), WithGeodetic: bool(value.with_geodetic)}
	for i := 0; i < 4; i++ {
		out.InitialGuess[i] = float64(value.initial_guess[i])
		out.KlobucharAlpha[i] = float64(value.klobuchar_alpha[i])
		out.KlobucharBeta[i] = float64(value.klobuchar_beta[i])
	}
	out.PressureHPA = float64(value.pressure_hpa)
	out.TemperatureK = float64(value.temperature_k)
	out.RelativeHumidity = float64(value.relative_humidity)
	if count != 0 {
		rows := unsafe.Slice(value.observations, count)
		out.Observations = make([]SPPObservation, count)
		for i := range rows {
			if rows[i].sat_id == nil {
				return SPPConfig{}, errors.New("sidereon: native SPP satellite ID pointer is null")
			}
			out.Observations[i] = SPPObservation{SatelliteID: C.GoString(rows[i].sat_id), PseudorangeM: float64(rows[i].pseudorange_m)}
		}
	}
	return out, nil
}

func (r *RinexSPPInputs) EpochInputs(index int) (SppInputsV2, error) {
	if r == nil || r.handle == nil {
		return SppInputsV2{}, ErrClosed
	}
	idx, err := cSize(index, "RINEX SPP epoch index")
	if err != nil {
		return SppInputsV2{}, err
	}
	var out C.SidereonSppInputsV2
	var result SppInputsV2
	var op error
	err = r.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			op = statusErrorLocked(uint32(C.sidereon_rinex_spp_inputs_epoch_inputs((*C.SidereonRinexSppInputs)(p), idx, &out)))
			if op == nil {
				result, op = sppInputsV2FromC(out)
			}
		})
		return op
	})
	if err != nil {
		return SppInputsV2{}, err
	}
	return result, nil
}

func sppInputsV2FromC(out C.SidereonSppInputsV2) (SppInputsV2, error) {
	base, err := sppBaseFromC(out.base)
	if err != nil {
		return SppInputsV2{}, err
	}
	maxOuter, err := sizeTToInt(out.robust.max_outer, "SPP robust max outer")
	if err != nil {
		return SppInputsV2{}, err
	}
	seeds, err := sizeTToInt(out.policy.coarse_search_seeds, "SPP coarse-search seeds")
	if err != nil {
		return SppInputsV2{}, err
	}
	result := SppInputsV2{Base: base, BeidouEnabled: bool(out.beidou_klobuchar_enabled), RobustEnabled: bool(out.robust_enabled), Robust: NativeSPPRobustConfig{HuberK: float64(out.robust.huber_k), ScaleFloorM: float64(out.robust.scale_floor_m), MaxOuter: uint64(maxOuter), OuterToleranceM: float64(out.robust.outer_tol_m)}, Policy: NativeSPPSolvePolicy{UseValidationOptions: bool(out.policy.use_validation_options), CoarseSearchEnabled: bool(out.policy.coarse_search_enabled), CoarseSearchSeeds: seeds, Validation: NativeSPPValidationOptions{MaxPDOPEnabled: bool(out.policy.validation.max_pdop_enabled), MaxPDOP: float64(out.policy.validation.max_pdop), MinPlausibleRadiusM: float64(out.policy.validation.min_plausible_radius_m), MaxPlausibleRadiusM: float64(out.policy.validation.max_plausible_radius_m), MaxConvergedResidualRMSM: float64(out.policy.validation.max_converged_residual_rms_m)}}}
	for i := 0; i < 4; i++ {
		result.BeidouAlpha[i], result.BeidouBeta[i] = float64(out.beidou_klobuchar_alpha[i]), float64(out.beidou_klobuchar_beta[i])
	}
	channelCount, err := sizeTToInt(out.glonass_channel_count, "GLONASS channel count")
	if err != nil {
		return SppInputsV2{}, err
	}
	if channelCount > 0 && out.glonass_channels == nil {
		return SppInputsV2{}, errors.New("sidereon: native GLONASS channels pointer is null")
	}
	if channelCount != 0 {
		rows := unsafe.Slice(out.glonass_channels, channelCount)
		result.GlonassChannels = make([]NativeGlonassChannel, channelCount)
		for i := range rows {
			result.GlonassChannels[i] = NativeGlonassChannel{Slot: uint8(rows[i].slot), Channel: int8(rows[i].channel)}
		}
	}
	return result, nil
}

func (r *RinexSPPSolutions) Count() (int, error) {
	if r == nil || r.handle == nil {
		return 0, ErrClosed
	}
	var out C.size_t
	var op error
	err := r.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			op = statusErrorLocked(uint32(C.sidereon_rinex_spp_solutions_count((*C.SidereonRinexSppSolutions)(p), &out)))
		})
		return op
	})
	if err != nil {
		return 0, err
	}
	return sizeTToInt(out, "RINEX SPP solution count")
}
func (r *RinexSPPSolutions) Epoch(index int) (NativeRinexSPPEpoch, error) {
	if r == nil || r.handle == nil {
		return NativeRinexSPPEpoch{}, ErrClosed
	}
	idx, err := cSize(index, "RINEX SPP epoch index")
	if err != nil {
		return NativeRinexSPPEpoch{}, err
	}
	var out C.SidereonRinexSppEpoch
	var op error
	err = r.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			op = statusErrorLocked(uint32(C.sidereon_rinex_spp_solutions_epoch((*C.SidereonRinexSppSolutions)(p), idx, &out)))
		})
		return op
	})
	if err != nil {
		return NativeRinexSPPEpoch{}, err
	}
	return rinexEpochFromC(out)
}
func (r *RinexSPPSolutions) SolutionOK(index int) (bool, error) {
	if r == nil || r.handle == nil {
		return false, ErrClosed
	}
	idx, err := cSize(index, "RINEX SPP epoch index")
	if err != nil {
		return false, err
	}
	var out C.bool
	var op error
	err = r.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			op = statusErrorLocked(uint32(C.sidereon_rinex_spp_solution_ok((*C.SidereonRinexSppSolutions)(p), idx, &out)))
		})
		return op
	})
	return bool(out), err
}
func (r *RinexSPPSolutions) SolutionError(index int) (string, error) {
	if r == nil || r.handle == nil {
		return "", ErrClosed
	}
	idx, err := cSize(index, "RINEX SPP epoch index")
	if err != nil {
		return "", err
	}
	var out string
	var op error
	err = r.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			var w, req C.size_t
			status := C.sidereon_rinex_spp_solution_error((*C.SidereonRinexSppSolutions)(p), idx, nil, 0, &w, &req)
			if e := statusErrorLocked(uint32(status)); e != nil {
				op = e
				return
			}
			n, e := validateNativeQuery("RINEX SPP solution error", uint64(w), uint64(req))
			if e != nil {
				op = e
				return
			}
			buf := make([]byte, n)
			var ptr *C.uint8_t
			if n > 0 {
				ptr = (*C.uint8_t)(unsafe.Pointer(&buf[0]))
			}
			// The native fill call owns both output counters; never let the
			// query values masquerade as the fill result if an implementation
			// fails to write them.
			w, req = 0, 0
			status = C.sidereon_rinex_spp_solution_error((*C.SidereonRinexSppSolutions)(p), idx, ptr, C.size_t(n), &w, &req)
			if e := statusErrorLocked(uint32(status)); e != nil {
				op = e
				return
			}
			m, e := validateNativeOutput("RINEX SPP solution error", n, uint64(w), uint64(req))
			if e != nil {
				op = e
				return
			}
			out = string(buf[:m])
		})
		return op
	})
	return out, err
}
func (r *RinexSPPSolutions) Solution(index int) (*SppSolutionHandle, error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	idx, err := cSize(index, "RINEX SPP epoch index")
	if err != nil {
		return nil, err
	}
	var out *C.SidereonSppSolution
	var op error
	err = r.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			status := C.sidereon_rinex_spp_solution((*C.SidereonRinexSppSolutions)(p), idx, &out)
			op = statusErrorLocked(uint32(status))
			if op != nil && out != nil {
				C.sidereon_spp_solution_free(out)
				out = nil
			}
		})
		return op
	})
	if err != nil {
		return nil, err
	}
	return newSppHandle(out)
}

func (b *SPPBatch) Count() (int, error) {
	if b == nil || b.handle == nil {
		return 0, ErrClosed
	}
	var out C.size_t
	var op error
	err := b.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() { op = statusErrorLocked(uint32(C.sidereon_spp_batch_count((*C.SidereonSppBatch)(p), &out))) })
		return op
	})
	if err != nil {
		return 0, err
	}
	return sizeTToInt(out, "SPP batch count")
}
func (b *SPPBatch) EpochOK(index int) (bool, error) {
	if b == nil || b.handle == nil {
		return false, ErrClosed
	}
	idx, err := cSize(index, "SPP batch epoch index")
	if err != nil {
		return false, err
	}
	var out C.bool
	var op error
	err = b.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			op = statusErrorLocked(uint32(C.sidereon_spp_batch_epoch_ok((*C.SidereonSppBatch)(p), idx, &out)))
		})
		return op
	})
	return bool(out), err
}
func (b *SPPBatch) Error(index int) (string, error) {
	if b == nil || b.handle == nil {
		return "", ErrClosed
	}
	idx, err := cSize(index, "SPP batch epoch index")
	if err != nil {
		return "", err
	}
	var out string
	var op error
	err = b.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			var w, req C.size_t
			status := C.sidereon_spp_batch_error((*C.SidereonSppBatch)(p), idx, nil, 0, &w, &req)
			if e := statusErrorLocked(uint32(status)); e != nil {
				op = e
				return
			}
			n, e := validateNativeQuery("SPP batch error", uint64(w), uint64(req))
			if e != nil {
				op = e
				return
			}
			buf := make([]byte, n)
			var ptr *C.uint8_t
			if n > 0 {
				ptr = (*C.uint8_t)(unsafe.Pointer(&buf[0]))
			}
			w, req = 0, 0
			status = C.sidereon_spp_batch_error((*C.SidereonSppBatch)(p), idx, ptr, C.size_t(n), &w, &req)
			if e := statusErrorLocked(uint32(status)); e != nil {
				op = e
				return
			}
			m, e := validateNativeOutput("SPP batch error", n, uint64(w), uint64(req))
			if e != nil {
				op = e
				return
			}
			out = string(buf[:m])
		})
		return op
	})
	return out, err
}
func (b *SPPBatch) Solution(index int) (*SppSolutionHandle, error) {
	if b == nil || b.handle == nil {
		return nil, ErrClosed
	}
	idx, err := cSize(index, "SPP batch epoch index")
	if err != nil {
		return nil, err
	}
	var out *C.SidereonSppSolution
	var op error
	err = b.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			status := C.sidereon_spp_batch_solution((*C.SidereonSppBatch)(p), idx, &out)
			op = statusErrorLocked(uint32(status))
			if op != nil && out != nil {
				C.sidereon_spp_solution_free(out)
				out = nil
			}
		})
		return op
	})
	if err != nil {
		return nil, err
	}
	return newSppHandle(out)
}

func (s *SppSolutionHandle) Solution() (SPPSolution, error) {
	if s == nil || s.handle == nil {
		return SPPSolution{}, ErrClosed
	}
	var out SPPSolution
	err := s.handle.read(func(p unsafe.Pointer) error {
		var e error
		withCThread(func() { out, e = readSPPSolutionLocked((*C.SidereonSppSolution)(p)) })
		return e
	})
	return out, err
}
func (s *SppSolutionHandle) covariance(enu bool) ([9]float64, error) {
	if s == nil || s.handle == nil {
		return [9]float64{}, ErrClosed
	}
	var out [9]float64
	var op error
	err := s.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			var ptr *C.double
			if len(out) > 0 {
				ptr = (*C.double)(unsafe.Pointer(&out[0]))
			}
			if enu {
				op = statusErrorLocked(uint32(C.sidereon_spp_solution_position_covariance_enu_m2((*C.SidereonSppSolution)(p), ptr, 9)))
			} else {
				op = statusErrorLocked(uint32(C.sidereon_spp_solution_position_covariance_ecef_m2((*C.SidereonSppSolution)(p), ptr, 9)))
			}
		})
		return op
	})
	return out, err
}
func (s *SppSolutionHandle) CovarianceECEFM2() ([9]float64, error) { return s.covariance(false) }
func (s *SppSolutionHandle) CovarianceENUM2() ([9]float64, error)  { return s.covariance(true) }
func (s *SppSolutionHandle) RejectedSatellites() ([]NativeSPPRejectedSatellite, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var out []NativeSPPRejectedSatellite
	var op error
	err := s.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			var w, req C.size_t
			status := C.sidereon_spp_solution_rejected_sats((*C.SidereonSppSolution)(p), nil, 0, &w, &req)
			if e := statusErrorLocked(uint32(status)); e != nil {
				op = e
				return
			}
			n, e := validateNativeQuery("SPP rejected satellites", uint64(w), uint64(req))
			if e != nil {
				op = e
				return
			}
			rows := make([]C.SidereonSppRejectedSat, n)
			var ptr *C.SidereonSppRejectedSat
			if n > 0 {
				ptr = &rows[0]
			}
			w, req = 0, 0
			status = C.sidereon_spp_solution_rejected_sats((*C.SidereonSppSolution)(p), ptr, C.size_t(n), &w, &req)
			if e := statusErrorLocked(uint32(status)); e != nil {
				op = e
				return
			}
			m, e := validateNativeOutput("SPP rejected satellites", n, uint64(w), uint64(req))
			if e != nil {
				op = e
				return
			}
			out = make([]NativeSPPRejectedSatellite, m)
			for i := range out {
				out[i] = NativeSPPRejectedSatellite{SatelliteID: tokenFromC(rows[i].sat_id), Reason: uint32(rows[i].reason)}
				if out[i].Reason > 3 {
					op = invalidArgument("invalid SPP rejection reason returned by native code")
					return
				}
			}
		})
		return op
	})
	return out, err
}
func (s *SppSolutionHandle) ReceiverClockDrift() (float64, bool, error) {
	if s == nil || s.handle == nil {
		return 0, false, ErrClosed
	}
	var present C.bool
	var value C.double
	var op error
	err := s.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			op = statusErrorLocked(uint32(C.sidereon_spp_solution_rx_clock_drift_s_s((*C.SidereonSppSolution)(p), &present, &value)))
		})
		return op
	})
	return float64(value), bool(present), err
}
func (s *SppSolutionHandle) SystemClocks() ([]NativeSPPSystemClock, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var out []NativeSPPSystemClock
	var op error
	err := s.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			var w, req C.size_t
			status := C.sidereon_spp_solution_system_clocks((*C.SidereonSppSolution)(p), nil, 0, &w, &req)
			if e := statusErrorLocked(uint32(status)); e != nil {
				op = e
				return
			}
			n, e := validateNativeQuery("SPP system clocks", uint64(w), uint64(req))
			if e != nil {
				op = e
				return
			}
			rows := make([]C.SidereonSppSystemClock, n)
			var ptr *C.SidereonSppSystemClock
			if n > 0 {
				ptr = &rows[0]
			}
			w, req = 0, 0
			status = C.sidereon_spp_solution_system_clocks((*C.SidereonSppSolution)(p), ptr, C.size_t(n), &w, &req)
			if e := statusErrorLocked(uint32(status)); e != nil {
				op = e
				return
			}
			m, e := validateNativeOutput("SPP system clocks", n, uint64(w), uint64(req))
			if e != nil {
				op = e
				return
			}
			out = make([]NativeSPPSystemClock, m)
			for i := range out {
				out[i] = NativeSPPSystemClock{System: uint32(rows[i].system), ReceiverS: float64(rows[i].rx_clock_s)}
				if out[i].System > 6 {
					op = invalidArgument("invalid GNSS system returned by native code")
					return
				}
			}
		})
		return op
	})
	return out, err
}
func (s *SppSolutionHandle) SystemTDOPs() ([]NativeSPPSystemTDOP, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var out []NativeSPPSystemTDOP
	var op error
	err := s.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			var w, req C.size_t
			status := C.sidereon_spp_solution_system_tdops((*C.SidereonSppSolution)(p), nil, 0, &w, &req)
			if e := statusErrorLocked(uint32(status)); e != nil {
				op = e
				return
			}
			n, e := validateNativeQuery("SPP system TDOPs", uint64(w), uint64(req))
			if e != nil {
				op = e
				return
			}
			rows := make([]C.SidereonSppSystemTdop, n)
			var ptr *C.SidereonSppSystemTdop
			if n > 0 {
				ptr = &rows[0]
			}
			w, req = 0, 0
			status = C.sidereon_spp_solution_system_tdops((*C.SidereonSppSolution)(p), ptr, C.size_t(n), &w, &req)
			if e := statusErrorLocked(uint32(status)); e != nil {
				op = e
				return
			}
			m, e := validateNativeOutput("SPP system TDOPs", n, uint64(w), uint64(req))
			if e != nil {
				op = e
				return
			}
			out = make([]NativeSPPSystemTDOP, m)
			for i := range out {
				out[i] = NativeSPPSystemTDOP{System: uint32(rows[i].system), TDOP: float64(rows[i].tdop)}
				if out[i].System > 6 {
					op = invalidArgument("invalid GNSS system returned by native code")
					return
				}
			}
		})
		return op
	})
	return out, err
}

func RINEXSPPOptionsInit() (NativeRinexSPPOptions, error) {
	var c C.SidereonRinexSppOptions
	var err error
	withCThread(func() { err = statusErrorLocked(uint32(C.sidereon_rinex_spp_options_init(&c))) })
	if err != nil {
		return NativeRinexSPPOptions{}, err
	}
	out := NativeRinexSPPOptions{Ionosphere: bool(c.ionosphere), Troposphere: bool(c.troposphere), InitialGuessEnabled: bool(c.initial_guess_enabled), PressureHPA: float64(c.pressure_hpa), TemperatureK: float64(c.temperature_k), RelativeHumidity: float64(c.relative_humidity), RobustEnabled: bool(c.robust_enabled), Robust: NativeSPPRobustConfig{HuberK: float64(c.robust.huber_k), ScaleFloorM: float64(c.robust.scale_floor_m), MaxOuter: uint64(c.robust.max_outer), OuterToleranceM: float64(c.robust.outer_tol_m)}}
	for i := 0; i < 4; i++ {
		out.InitialGuess[i] = float64(c.initial_guess[i])
	}
	return out, nil
}
func SPPInputsV2Init() (SppInputsV2, error) {
	alloc := new(cRtkAlloc)
	defer alloc.close()
	c, err := makeDgnssBaseInputs(SPPConfig{}, alloc)
	if err != nil {
		return SppInputsV2{}, err
	}
	return sppInputsV2FromC(*c)
}

func SolveSPPV2(sp3 *SP3, input SppInputsV2) (*SppSolutionHandle, error) {
	if sp3 == nil || sp3.handle == nil {
		return nil, ErrClosed
	}
	alloc := new(cRtkAlloc)
	defer alloc.close()
	cinput, err := makeSppV2(input, alloc)
	if err != nil {
		return nil, err
	}
	var out *C.SidereonSppSolution
	err = sp3.handle.with(func(p unsafe.Pointer) error {
		return statusCall(func() C.enum_SidereonStatus { return C.sidereon_solve_spp_v2((*C.SidereonSp3)(p), cinput, &out) })
	})
	runtime.KeepAlive(input.Base.Observations)
	if err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_spp_solution_free(out) })
		}
		return nil, err
	}
	return newSppHandle(out)
}

func solveSPPBatch(sp3 *SP3, inputs []SppInputsV2, withGeodetic, parallel bool, policy NativeSPPSolvePolicy) (*SPPBatch, error) {
	if sp3 == nil || sp3.handle == nil {
		return nil, ErrClosed
	}
	alloc := new(cRtkAlloc)
	defer alloc.close()
	cinputs, count, err := makeSppV2Array(inputs, alloc)
	if err != nil {
		return nil, err
	}
	var cpolicy C.SidereonSppSolvePolicy
	if err := nativePolicy(policy, &cpolicy); err != nil {
		return nil, err
	}
	var out *C.SidereonSppBatch
	err = sp3.handle.with(func(p unsafe.Pointer) error {
		return statusCall(func() C.enum_SidereonStatus {
			if parallel {
				return C.sidereon_solve_spp_batch_parallel((*C.SidereonSp3)(p), cinputs, count, C.bool(withGeodetic), &cpolicy, &out)
			}
			return C.sidereon_solve_spp_batch_serial((*C.SidereonSp3)(p), cinputs, count, C.bool(withGeodetic), &cpolicy, &out)
		})
	})
	if err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_spp_batch_free(out) })
		}
		return nil, err
	}
	return newSppBatch(out)
}
func SolveSPPBatchSerial(sp3 *SP3, inputs []SppInputsV2, withGeodetic bool, policy NativeSPPSolvePolicy) (*SPPBatch, error) {
	return solveSPPBatch(sp3, inputs, withGeodetic, false, policy)
}
func SolveSPPBatchParallel(sp3 *SP3, inputs []SppInputsV2, withGeodetic bool, policy NativeSPPSolvePolicy) (*SPPBatch, error) {
	return solveSPPBatch(sp3, inputs, withGeodetic, true, policy)
}

func SolveSPPWithDoppler(sp3 *SP3, input SppInputsV2, observations []NativeSppDopplerObservation) (NativeSppDopplerResult, error) {
	if sp3 == nil || sp3.handle == nil {
		return NativeSppDopplerResult{}, ErrClosed
	}
	alloc := new(cRtkAlloc)
	defer alloc.close()
	cinput, err := makeSppV2(input, alloc)
	if err != nil {
		return NativeSppDopplerResult{}, err
	}
	rows, count, err := makeDopplerRows(observations, alloc)
	if err != nil {
		return NativeSppDopplerResult{}, err
	}
	var out *C.SidereonSppDopplerSolution
	var result NativeSppDopplerResult
	err = sp3.handle.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			status := C.sidereon_solve_spp_with_doppler_velocity((*C.SidereonSp3)(p), cinput, rows, count, &out)
			if err := statusErrorLocked(uint32(status)); err != nil {
				if out != nil {
					C.sidereon_spp_doppler_solution_free(out)
					out = nil
				}
				return err
			}
			if out == nil {
				return errors.New("sidereon: native SPP Doppler solve returned no solution")
			}
			defer C.sidereon_spp_doppler_solution_free(out)
			var err error
			result, err = readSPPDopplerResultLocked(out)
			return err
		})
	})
	if err != nil {
		return NativeSppDopplerResult{}, err
	}
	return result, nil
}

func SolveSPPFromRINEXObs(broadcast *BroadcastEphemeris, obs *RinexObs, options *NativeRinexSPPOptions, withGeodetic bool, policy *NativeSPPSolvePolicy) (*RinexSPPSolutions, error) {
	if broadcast == nil || broadcast.resource == nil || obs == nil || obs.resource == nil {
		return nil, ErrClosed
	}
	alloc := new(cRtkAlloc)
	defer alloc.close()
	var coptions *C.SidereonRinexSppOptions
	var err error
	if options != nil {
		coptions, err = makeRinexOptions(*options, alloc)
		if err != nil {
			return nil, err
		}
	}
	var cpolicy *C.SidereonSppSolvePolicy
	if policy != nil {
		bytes, e := checkedNativeAllocationSize(1, unsafe.Sizeof(C.SidereonSppSolvePolicy{}))
		if e != nil {
			return nil, e
		}
		memory, e := alloc.malloc(bytes, "SPP solve policy")
		if e != nil {
			return nil, e
		}
		cpolicy = (*C.SidereonSppSolvePolicy)(memory)
		if e = nativePolicy(*policy, cpolicy); e != nil {
			return nil, e
		}
	}
	var out *C.SidereonRinexSppSolutions
	err = withScenarioInputs([]scenarioInput{{resource: broadcast.resource}, {resource: obs.resource}}, func(pointers []unsafe.Pointer) error {
		var operationErr error
		withCThread(func() {
			status := C.sidereon_solve_spp_from_rinex_obs((*C.SidereonBroadcastEphemeris)(pointers[0]), (*C.SidereonRinexObs)(pointers[1]), coptions, C.bool(withGeodetic), cpolicy, &out)
			operationErr = statusErrorLocked(uint32(status))
		})
		return operationErr
	})
	if err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_rinex_spp_solutions_free(out) })
		}
		return nil, err
	}
	return newRinexSPPSolutions(out)
}

func SPPInputsFromRINEXObs(obs *RinexObs, broadcast *BroadcastEphemeris, options *NativeRinexSPPOptions) (*RinexSPPInputs, error) {
	if obs == nil || obs.resource == nil || broadcast == nil || broadcast.resource == nil {
		return nil, ErrClosed
	}
	alloc := new(cRtkAlloc)
	defer alloc.close()
	var coptions *C.SidereonRinexSppOptions
	var err error
	if options != nil {
		coptions, err = makeRinexOptions(*options, alloc)
		if err != nil {
			return nil, err
		}
	}
	var out *C.SidereonRinexSppInputs
	err = withScenarioInputs([]scenarioInput{{resource: obs.resource}, {resource: broadcast.resource}}, func(pointers []unsafe.Pointer) error {
		var operationErr error
		withCThread(func() {
			status := C.sidereon_spp_inputs_from_rinex_obs((*C.SidereonRinexObs)(pointers[0]), (*C.SidereonBroadcastEphemeris)(pointers[1]), coptions, &out)
			operationErr = statusErrorLocked(uint32(status))
		})
		return operationErr
	})
	if err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_rinex_spp_inputs_free(out) })
		}
		return nil, err
	}
	return newRinexSPPInputs(out)
}

func SolveWithFallback(precise []*SP3, broadcast *BroadcastEphemeris, input SPPConfig, policy StalenessPolicy) (*SourcedSolution, error) {
	if broadcast == nil || broadcast.resource == nil {
		return nil, ErrClosed
	}
	handles := make([]*positioningHandle, len(precise))
	for i, item := range precise {
		if item == nil || item.handle == nil {
			return nil, ErrClosed
		}
		handles[i] = item.handle
	}
	preciseCount, err := checkedNativeSize(len(handles))
	if err != nil {
		return nil, err
	}
	alloc := new(cRtkAlloc)
	defer alloc.close()
	cinput, err := makeLegacySpp(input, alloc)
	if err != nil {
		return nil, err
	}
	var out *C.SidereonSourcedSolution
	inputs := make([]scenarioInput, 0, len(handles)+1)
	for _, handle := range handles {
		inputs = append(inputs, scenarioInput{positioning: handle})
	}
	inputs = append(inputs, scenarioInput{resource: broadcast.resource})
	err = withScenarioInputs(inputs, func(pointers []unsafe.Pointer) error {
		bytes, e := checkedNativeAllocationSize(len(handles), unsafe.Sizeof((*C.SidereonSp3)(nil)))
		if e != nil {
			return e
		}
		var array **C.SidereonSp3
		if len(handles) != 0 {
			memory, e := alloc.malloc(bytes, "precise SP3 pointers")
			if e != nil {
				return e
			}
			rows := unsafe.Slice((**C.SidereonSp3)(memory), len(handles))
			for i, p := range pointers[:len(handles)] {
				rows[i] = (*C.SidereonSp3)(p)
			}
			array = (**C.SidereonSp3)(memory)
		}
		var operationErr error
		withCThread(func() {
			status := C.sidereon_solve_with_fallback(array, preciseCount, (*C.SidereonBroadcastEphemeris)(pointers[len(handles)]), cinput, C.SidereonStalenessPolicy{max_staleness_s: C.double(policy.MaxStalenessS)}, &out)
			operationErr = fallbackStatusErrorLocked(status)
		})
		return operationErr
	})
	if err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_sourced_solution_free(out) })
		}
		return nil, err
	}
	if out == nil {
		return nil, missingNativeHandle("fallback solution")
	}
	return &SourcedSolution{handle: newPositioningHandle(unsafe.Pointer(out), releaseSourcedSolution)}, nil
}
