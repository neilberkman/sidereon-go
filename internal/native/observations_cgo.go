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

type NativeCalendarEpoch struct {
	Year   int32
	Month  int32
	Day    int32
	Hour   int32
	Minute int32
	Second float64
}

func calendarEpochFromC(value C.SidereonCalendarEpoch) NativeCalendarEpoch {
	return NativeCalendarEpoch{Year: int32(value.year), Month: int32(value.month), Day: int32(value.day), Hour: int32(value.hour), Minute: int32(value.minute), Second: float64(value.second)}
}

type NativeRinexObsHeader struct {
	Version             float64
	HasApproxPosition   bool
	ApproxPosition      [3]float64
	HasAntennaDelta     bool
	AntennaDelta        [3]float64
	HasInterval         bool
	Interval            float64
	HasTimeOfFirstObs   bool
	TimeOfFirstObs      NativeCalendarEpoch
	TimeOfFirstObsScale uint32
	ObsCodeCount        int
	PhaseShiftCount     int
	ScaleFactorCount    int
	GLONASSSlotCount    int
	HasMarkerName       bool
	MarkerName          string
}

type NativeRinexObsCode struct {
	System uint32
	Code   string
}
type NativeRinexObsEpoch struct {
	Epoch          NativeCalendarEpoch
	Flag           uint8
	SatelliteCount int
}
type NativeRinexObsValue struct {
	SatelliteID, Code string
	Kind              uint32
	HasValue          bool
	Value             float64
	LLI, SSI          int32
}
type NativeRinexObsPseudorange struct {
	SatelliteID  string
	PseudorangeM float64
}
type NativeClockPhaseSample struct {
	HasPhaseS bool
	PhaseS    float64
}
type NativeRinexObsCarrierPhase struct {
	SatelliteID, Code string
	HasValueCycles    bool
	ValueCycles       float64
	LLI, SSI          int32
	HasFrequency      bool
	FrequencyHz       float64
	HasWavelength     bool
	WavelengthM       float64
	HasValueM         bool
	ValueM            float64
	PhaseShiftCycles  float64
}

type RinexObs struct {
	_        noCopy
	resource *resource
	cleanup  runtime.Cleanup
}

func validateRINEXObservationKind(value uint32) error {
	if value > 4 {
		return invalidArgument("invalid RINEX observation kind returned by native code")
	}
	return nil
}

func validateGNSSSystemValue(value uint32) error {
	if value > 6 {
		return invalidArgument("invalid GNSS system returned by native code")
	}
	return nil
}

func newRinexObs(pointer *C.SidereonRinexObs) (*RinexObs, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	handle := &RinexObs{resource: &resource{ptr: unsafe.Pointer(pointer), release: func(pointer unsafe.Pointer) {
		C.sidereon_rinex_obs_free((*C.SidereonRinexObs)(pointer))
	}}}
	handle.cleanup = runtime.AddCleanup(handle, cleanupResource, handle.resource)
	return handle, nil
}

func ParseRinexObs(data []byte) (*RinexObs, error) {
	var pointer *C.SidereonRinexObs
	err := withInput(data, func(input *C.uint8_t, length C.size_t) uint32 {
		return C.sidereon_rinex_obs_parse(input, length, &pointer)
	})
	if err != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_rinex_obs_free(pointer) })
		}
		return nil, err
	}
	handle, err := newRinexObs(pointer)
	if err != nil && pointer != nil {
		withCThread(func() { C.sidereon_rinex_obs_free(pointer) })
	}
	return handle, err
}

func LoadRinexObs(path string) (*RinexObs, error) {
	var pointer *C.SidereonRinexObs
	err := withStringError(path, func(value *C.char) error {
		return statusErrorLocked(C.sidereon_rinex_obs_load(value, &pointer))
	})
	if err != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_rinex_obs_free(pointer) })
		}
		return nil, err
	}
	handle, err := newRinexObs(pointer)
	if err != nil && pointer != nil {
		withCThread(func() { C.sidereon_rinex_obs_free(pointer) })
	}
	return handle, err
}

func (obs *RinexObs) Close() error {
	if obs == nil {
		return nil
	}
	return closeProtocolResource(obs, obs.resource, &obs.cleanup)
}

func (obs *RinexObs) with(fn func(*C.SidereonRinexObs) error) error {
	if obs == nil || obs.resource == nil {
		return ErrClosed
	}
	return obs.resource.with(func(pointer unsafe.Pointer) error { return fn((*C.SidereonRinexObs)(pointer)) })
}

func copyByteOutput(label string, call func(*C.uint8_t, C.size_t, *C.size_t, *C.size_t) uint32) ([]byte, error) {
	var written, required C.size_t
	if err := callStatus(func() uint32 { return call(nil, 0, &written, &required) }); err != nil {
		return nil, err
	}
	n, err := validateNativeQuery(label, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(n, 1); err != nil {
		return nil, err
	}
	buffer := make([]byte, n)
	var output *C.uint8_t
	if n != 0 {
		output = (*C.uint8_t)(unsafe.Pointer(&buffer[0]))
	}
	if err := callStatus(func() uint32 { return call(output, C.size_t(n), &written, &required) }); err != nil {
		return nil, err
	}
	w, err := validateNativeOutput(label, n, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	return append([]byte(nil), buffer[:w]...), nil
}

func (obs *RinexObs) Version() (float64, error) {
	var out C.double
	err := obs.with(func(pointer *C.SidereonRinexObs) error {
		return callStatus(func() uint32 { return C.sidereon_rinex_obs_version(pointer, &out) })
	})
	runtime.KeepAlive(obs)
	return float64(out), err
}

func (obs *RinexObs) Header() (NativeRinexObsHeader, error) {
	var value C.SidereonRinexObsHeader
	err := obs.with(func(pointer *C.SidereonRinexObs) error {
		return callStatus(func() uint32 { return C.sidereon_rinex_obs_header(pointer, &value) })
	})
	if err != nil {
		return NativeRinexObsHeader{}, err
	}
	var marker string
	for i := 0; i < len(value.marker_name); i++ {
		if value.marker_name[i] == 0 {
			break
		}
		marker += string(byte(value.marker_name[i]))
	}
	out := NativeRinexObsHeader{Version: float64(value.version), HasApproxPosition: bool(value.has_approx_position_m), HasAntennaDelta: bool(value.has_antenna_delta_hen_m), HasInterval: bool(value.has_interval_s), Interval: float64(value.interval_s), HasTimeOfFirstObs: bool(value.has_time_of_first_obs), TimeOfFirstObs: calendarEpochFromC(value.time_of_first_obs), TimeOfFirstObsScale: uint32(value.time_of_first_obs_scale), HasMarkerName: bool(value.has_marker_name), MarkerName: marker}
	for i := range out.ApproxPosition {
		out.ApproxPosition[i] = float64(value.approx_position_m[i])
		out.AntennaDelta[i] = float64(value.antenna_delta_hen_m[i])
	}
	counts := []*int{&out.ObsCodeCount, &out.PhaseShiftCount, &out.ScaleFactorCount, &out.GLONASSSlotCount}
	nativeCounts := []C.size_t{value.obs_code_count, value.phase_shift_count, value.scale_factor_count, value.glonass_slot_count}
	for i := range counts {
		count, err := checkedNativeCount(uint64(nativeCounts[i]))
		if err != nil {
			return NativeRinexObsHeader{}, err
		}
		*counts[i] = count
	}
	runtime.KeepAlive(obs)
	return out, nil
}

func (obs *RinexObs) EpochCount() (int, error) {
	var count C.size_t
	err := obs.with(func(pointer *C.SidereonRinexObs) error {
		return callStatus(func() uint32 { return C.sidereon_rinex_obs_epoch_count(pointer, &count) })
	})
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(count))
}

func (obs *RinexObs) Codes() ([]NativeRinexObsCode, error) {
	var result []NativeRinexObsCode
	err := obs.with(func(pointer *C.SidereonRinexObs) error {
		var written, required C.size_t
		if err := callStatus(func() uint32 { return C.sidereon_rinex_obs_codes(pointer, nil, 0, &written, &required) }); err != nil {
			return err
		}
		n, err := validateNativeQuery("RINEX observation codes", uint64(written), uint64(required))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonRinexObsCode{})); err != nil {
			return err
		}
		values := make([]C.SidereonRinexObsCode, n)
		var out *C.SidereonRinexObsCode
		if n != 0 {
			out = &values[0]
		}
		if err := callStatus(func() uint32 { return C.sidereon_rinex_obs_codes(pointer, out, C.size_t(n), &written, &required) }); err != nil {
			return err
		}
		w, err := validateNativeOutput("RINEX observation codes", n, uint64(written), uint64(required))
		if err != nil {
			return err
		}
		converted := make([]NativeRinexObsCode, w)
		for i := range converted {
			if err := validateGNSSSystemValue(uint32(values[i].system)); err != nil {
				return err
			}
			converted[i] = NativeRinexObsCode{System: uint32(values[i].system), Code: observationFixedString(values[i].code[:])}
		}
		result = converted
		return nil
	})
	runtime.KeepAlive(obs)
	return result, err
}

func observationFixedString(value []C.char) string {
	out := make([]byte, 0, len(value))
	for _, item := range value {
		if item == 0 {
			break
		}
		out = append(out, byte(item))
	}
	return string(out)
}

func (obs *RinexObs) Epochs() ([]NativeRinexObsEpoch, error) {
	var result []NativeRinexObsEpoch
	err := obs.with(func(pointer *C.SidereonRinexObs) error {
		var written, required C.size_t
		if err := callStatus(func() uint32 { return C.sidereon_rinex_obs_epochs(pointer, nil, 0, &written, &required) }); err != nil {
			return err
		}
		n, err := validateNativeQuery("RINEX observation epochs", uint64(written), uint64(required))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonRinexObsEpoch{})); err != nil {
			return err
		}
		values := make([]C.SidereonRinexObsEpoch, n)
		var out *C.SidereonRinexObsEpoch
		if n != 0 {
			out = &values[0]
		}
		if err := callStatus(func() uint32 { return C.sidereon_rinex_obs_epochs(pointer, out, C.size_t(n), &written, &required) }); err != nil {
			return err
		}
		w, err := validateNativeOutput("RINEX observation epochs", n, uint64(written), uint64(required))
		if err != nil {
			return err
		}
		result = make([]NativeRinexObsEpoch, w)
		for i := range result {
			result[i] = NativeRinexObsEpoch{Epoch: calendarEpochFromC(values[i].epoch), Flag: uint8(values[i].flag)}
			result[i].SatelliteCount, err = checkedNativeCount(uint64(values[i].satellite_count))
			if err != nil {
				return err
			}
		}
		return nil
	})
	runtime.KeepAlive(obs)
	return result, err
}

func (obs *RinexObs) Values(epoch int) ([]NativeRinexObsValue, error) {
	return obs.values(epoch, false)
}
func (obs *RinexObs) CarrierPhase(epoch int) ([]NativeRinexObsCarrierPhase, error) {
	if epoch < 0 {
		return nil, errNegativeIndex
	}
	var result []NativeRinexObsCarrierPhase
	err := obs.with(func(pointer *C.SidereonRinexObs) error {
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_rinex_obs_carrier_phase(pointer, C.size_t(epoch), nil, 0, &written, &required)
		}); err != nil {
			return err
		}
		n, err := validateNativeQuery("RINEX carrier phase", uint64(written), uint64(required))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonRinexObsCarrierPhase{})); err != nil {
			return err
		}
		values := make([]C.SidereonRinexObsCarrierPhase, n)
		var out *C.SidereonRinexObsCarrierPhase
		if n != 0 {
			out = &values[0]
		}
		if err := callStatus(func() uint32 {
			return C.sidereon_rinex_obs_carrier_phase(pointer, C.size_t(epoch), out, C.size_t(n), &written, &required)
		}); err != nil {
			return err
		}
		w, err := validateNativeOutput("RINEX carrier phase", n, uint64(written), uint64(required))
		if err != nil {
			return err
		}
		result = make([]NativeRinexObsCarrierPhase, w)
		for i := range result {
			v := values[i]
			result[i] = NativeRinexObsCarrierPhase{SatelliteID: tokenFromC(v.sat_id), Code: observationFixedString(v.code[:]), HasValueCycles: bool(v.has_value_cycles), ValueCycles: float64(v.value_cycles), LLI: int32(v.lli), SSI: int32(v.ssi), HasFrequency: bool(v.has_frequency_hz), FrequencyHz: float64(v.frequency_hz), HasWavelength: bool(v.has_wavelength_m), WavelengthM: float64(v.wavelength_m), HasValueM: bool(v.has_value_m), ValueM: float64(v.value_m), PhaseShiftCycles: float64(v.phase_shift_cycles)}
		}
		return nil
	})
	runtime.KeepAlive(obs)
	return result, err
}

func (obs *RinexObs) values(epoch int, pseudo bool) ([]NativeRinexObsValue, error) {
	if epoch < 0 {
		return nil, errNegativeIndex
	}
	var result []NativeRinexObsValue
	err := obs.with(func(pointer *C.SidereonRinexObs) error {
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_rinex_obs_values(pointer, C.size_t(epoch), nil, 0, &written, &required)
		}); err != nil {
			return err
		}
		n, err := validateNativeQuery("RINEX observation values", uint64(written), uint64(required))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonRinexObsValue{})); err != nil {
			return err
		}
		values := make([]C.SidereonRinexObsValue, n)
		var out *C.SidereonRinexObsValue
		if n != 0 {
			out = &values[0]
		}
		if err := callStatus(func() uint32 {
			return C.sidereon_rinex_obs_values(pointer, C.size_t(epoch), out, C.size_t(n), &written, &required)
		}); err != nil {
			return err
		}
		w, err := validateNativeOutput("RINEX observation values", n, uint64(written), uint64(required))
		if err != nil {
			return err
		}
		result = make([]NativeRinexObsValue, w)
		for i := range result {
			v := values[i]
			if err := validateRINEXObservationKind(uint32(v.kind)); err != nil {
				return err
			}
			result[i] = NativeRinexObsValue{SatelliteID: tokenFromC(v.sat_id), Code: observationFixedString(v.code[:]), Kind: uint32(v.kind), HasValue: bool(v.has_value), Value: float64(v.value), LLI: int32(v.lli), SSI: int32(v.ssi)}
		}
		return nil
	})
	runtime.KeepAlive(obs)
	return result, err
}

func (obs *RinexObs) Pseudoranges(epoch int) ([]NativeRinexObsPseudorange, error) {
	if epoch < 0 {
		return nil, errNegativeIndex
	}
	var result []NativeRinexObsPseudorange
	err := obs.with(func(pointer *C.SidereonRinexObs) error {
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_rinex_obs_pseudoranges(pointer, C.size_t(epoch), nil, 0, &written, &required)
		}); err != nil {
			return err
		}
		n, err := validateNativeQuery("RINEX pseudoranges", uint64(written), uint64(required))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonRinexObsPseudorange{})); err != nil {
			return err
		}
		values := make([]C.SidereonRinexObsPseudorange, n)
		var out *C.SidereonRinexObsPseudorange
		if n != 0 {
			out = &values[0]
		}
		if err := callStatus(func() uint32 {
			return C.sidereon_rinex_obs_pseudoranges(pointer, C.size_t(epoch), out, C.size_t(n), &written, &required)
		}); err != nil {
			return err
		}
		w, err := validateNativeOutput("RINEX pseudoranges", n, uint64(written), uint64(required))
		if err != nil {
			return err
		}
		result = make([]NativeRinexObsPseudorange, w)
		for i := range result {
			result[i] = NativeRinexObsPseudorange{SatelliteID: tokenFromC(values[i].sat_id), PseudorangeM: float64(values[i].pseudorange_m)}
		}
		return nil
	})
	runtime.KeepAlive(obs)
	return result, err
}

func (obs *RinexObs) Observation(epoch int, satellite, code string) (float64, bool, int32, int32, error) {
	if epoch < 0 {
		return 0, false, -1, -1, errNegativeIndex
	}
	if err := rejectEmbeddedNUL(satellite, "satellite token"); err != nil {
		return 0, false, -1, -1, err
	}
	if err := rejectEmbeddedNUL(code, "observation code"); err != nil {
		return 0, false, -1, -1, err
	}
	var value C.double
	var present C.bool
	var lli, ssi C.int32_t
	err := obs.with(func(pointer *C.SidereonRinexObs) error {
		return withTwoStrings(satellite, code, func(sat, obsCode *C.char) uint32 {
			return C.sidereon_rinex_obs_observation(pointer, C.size_t(epoch), sat, obsCode, &value, &present, &lli, &ssi)
		})
	})
	runtime.KeepAlive(obs)
	return float64(value), bool(present), int32(lli), int32(ssi), err
}

func (obs *RinexObs) ReceiverClockPhase() ([]NativeClockPhaseSample, error) {
	var result []NativeClockPhaseSample
	err := obs.with(func(pointer *C.SidereonRinexObs) error {
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_rinex_obs_receiver_clock_phase_deviations(pointer, nil, 0, &written, &required)
		}); err != nil {
			return err
		}
		n, err := validateNativeQuery("RINEX receiver clock phase", uint64(written), uint64(required))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonClockPhaseSample{})); err != nil {
			return err
		}
		values := make([]C.SidereonClockPhaseSample, n)
		var out *C.SidereonClockPhaseSample
		if n != 0 {
			out = &values[0]
		}
		if err := callStatus(func() uint32 {
			return C.sidereon_rinex_obs_receiver_clock_phase_deviations(pointer, out, C.size_t(n), &written, &required)
		}); err != nil {
			return err
		}
		w, err := validateNativeOutput("RINEX receiver clock phase", n, uint64(written), uint64(required))
		if err != nil {
			return err
		}
		result = make([]NativeClockPhaseSample, w)
		for i := range result {
			result[i] = NativeClockPhaseSample{HasPhaseS: bool(values[i].has_phase_s), PhaseS: float64(values[i].phase_s)}
		}
		return nil
	})
	runtime.KeepAlive(obs)
	return result, err
}

func (obs *RinexObs) Text() ([]byte, error) {
	var result []byte
	err := obs.with(func(pointer *C.SidereonRinexObs) error {
		var e error
		result, e = copyByteOutput("RINEX observation text", func(out *C.uint8_t, len C.size_t, written, required *C.size_t) uint32 {
			return C.sidereon_rinex_obs_to_rinex_text(pointer, out, len, written, required)
		})
		return e
	})
	runtime.KeepAlive(obs)
	return result, err
}

func RinexObservationFrequency(system uint32, code string, version float64, hasChannel bool, channel int8) (float64, error) {
	if err := validateGNSSSystemValue(system); err != nil {
		return 0, err
	}
	if err := validateRinexObservationCode(code); err != nil {
		return 0, err
	}
	var out C.double
	err := withString(code, func(value *C.char) uint32 {
		return C.sidereon_rinex_observation_frequency_hz(C.uint32_t(system), value, C.double(version), C.bool(hasChannel), C.int8_t(channel), &out)
	})
	return float64(out), err
}
func RinexObservationWavelength(system uint32, code string, version float64, hasChannel bool, channel int8) (float64, error) {
	if err := validateGNSSSystemValue(system); err != nil {
		return 0, err
	}
	if err := validateRinexObservationCode(code); err != nil {
		return 0, err
	}
	var out C.double
	err := withString(code, func(value *C.char) uint32 {
		return C.sidereon_rinex_observation_wavelength_m(C.uint32_t(system), value, C.double(version), C.bool(hasChannel), C.int8_t(channel), &out)
	})
	return float64(out), err
}

func withTwoStrings(first, second string, fn func(*C.char, *C.char) uint32) error {
	if err := rejectEmbeddedNUL(first, "satellite token"); err != nil {
		return err
	}
	if len(first) >= 16 {
		return errTokenTooLong
	}
	if err := validateRinexObservationCode(second); err != nil {
		return err
	}
	return withTwoStringsUnchecked(first, second, fn)
}

func withTwoTokens(first, second string, firstName, secondName string, fn func(*C.char, *C.char) uint32) error {
	if err := rejectEmbeddedNUL(first, firstName); err != nil {
		return err
	}
	if len(first) >= 16 {
		return errTokenTooLong
	}
	if err := rejectEmbeddedNUL(second, secondName); err != nil {
		return err
	}
	if len(second) >= 16 {
		return errTokenTooLong
	}
	return withTwoStringsUnchecked(first, second, fn)
}

func validateRinexObservationCode(value string) error {
	if err := rejectEmbeddedNUL(value, "observation code"); err != nil {
		return err
	}
	if len(value) >= 9 {
		return errors.New("sidereon: RINEX observation code is too long")
	}
	return nil
}

func withTwoStringsUnchecked(first, second string, fn func(*C.char, *C.char) uint32) error {
	var err error
	withCThread(func() {
		one := C.CString(first)
		two := C.CString(second)
		if one == nil || two == nil {
			if one != nil {
				C.free(unsafe.Pointer(one))
			}
			if two != nil {
				C.free(unsafe.Pointer(two))
			}
			err = errors.New("sidereon: unable to allocate native string")
			return
		}
		defer C.free(unsafe.Pointer(one))
		defer C.free(unsafe.Pointer(two))
		err = statusErrorLocked(fn(one, two))
	})
	return err
}

type NativeObservationQcOptions struct {
	HasIntervalOverride                             bool
	IntervalOverride, GapFactor, ClockJumpThreshold float64
}
type NativeObservationQcSummary struct {
	TotalEpochRecords, ObservationEpochs, EventRecords, PowerFailureEpochs, SkippedRecords          int
	HasInterval                                                                                     bool
	Interval                                                                                        float64
	IntervalSource                                                                                  uint32
	MissingEpochs, DataGapCount, SatelliteCount, SatelliteSignalCount, SystemSignalCount, NoteCount int
}
type NativeObservationQcDataGap struct {
	Start, End                     NativeCalendarEpoch
	NominalInterval, ObservedDelta float64
	MissingEpochs                  int
}
type NativeObservationQcClockJump struct {
	EpochIndex int
	Epoch      NativeCalendarEpoch
	DeltaS     float64
}
type NativeObservationQcCycleSlips struct {
	Observations, TotalSlips, SystemCount int
	HasObservationsPerSlip                bool
	ObservationsPerSlip                   float64
}
type NativeObservationQcSystemCycleSlip struct {
	System                 uint32
	Observations, Slips    int
	HasObservationsPerSlip bool
	ObservationsPerSlip    float64
}
type NativeObservationQcMpStats struct {
	N    int
	RMSM float64
}
type NativeObservationQcSatelliteMultipath struct {
	SatelliteID string
	HasMP1      bool
	MP1         NativeObservationQcMpStats
	HasMP2      bool
	MP2         NativeObservationQcMpStats
}
type NativeObservationQcSystemMultipath struct {
	System uint32
	HasMP1 bool
	MP1    NativeObservationQcMpStats
	HasMP2 bool
	MP2    NativeObservationQcMpStats
}
type NativeObservationQcSatellite struct {
	SatelliteID                               string
	EpochsWithObservations, ValueObservations int
}
type NativeObservationQcSignal struct {
	SatelliteID             string
	System                  uint32
	Code                    string
	ValueObservations       int
	HasSSI                  bool
	SSICounts               [10]uint64
	HasSNR                  bool
	SNRN                    int
	SNRMean, SNRMin, SNRMax float64
	HasSNRStd               bool
	SNRStd                  float64
}

type ObservationQcReport struct {
	_        noCopy
	resource *resource
	cleanup  runtime.Cleanup
}

func newObservationQcReport(pointer *C.SidereonObservationQcReport) (*ObservationQcReport, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	h := &ObservationQcReport{resource: &resource{ptr: unsafe.Pointer(pointer), release: func(p unsafe.Pointer) { C.sidereon_observation_qc_report_free((*C.SidereonObservationQcReport)(p)) }}}
	h.cleanup = runtime.AddCleanup(h, cleanupResource, h.resource)
	return h, nil
}
func qcOptions(value NativeObservationQcOptions) C.SidereonObservationQcOptions {
	return C.SidereonObservationQcOptions{has_interval_override_s: C.bool(value.HasIntervalOverride), interval_override_s: C.double(value.IntervalOverride), gap_factor: C.double(value.GapFactor), clock_jump_threshold_s: C.double(value.ClockJumpThreshold)}
}
func QcOptionsInit() (NativeObservationQcOptions, error) {
	var v C.SidereonObservationQcOptions
	err := callStatus(func() uint32 { return C.sidereon_observation_qc_options_init(&v) })
	return NativeObservationQcOptions{HasIntervalOverride: bool(v.has_interval_override_s), IntervalOverride: float64(v.interval_override_s), GapFactor: float64(v.gap_factor), ClockJumpThreshold: float64(v.clock_jump_threshold_s)}, err
}
func ObservationQCFromObs(obs *RinexObs, options *NativeObservationQcOptions) (*ObservationQcReport, error) {
	if obs == nil || obs.resource == nil {
		return nil, ErrClosed
	}
	var out *C.SidereonObservationQcReport
	err := obs.with(func(pointer *C.SidereonRinexObs) error {
		err := withCThreadError(func() error {
			var p *C.SidereonObservationQcOptions
			if options != nil {
				memory, allocationErr := checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonObservationQcOptions{}))
				if allocationErr != nil {
					return allocationErr
				}
				defer C.free(memory)
				p = (*C.SidereonObservationQcOptions)(memory)
				*p = qcOptions(*options)
			}
			return statusErrorLocked(C.sidereon_observation_qc_from_obs(pointer, p, &out))
		})
		if err != nil && out != nil {
			withCThread(func() { C.sidereon_observation_qc_report_free(out) })
			out = nil
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return newObservationQcReport(out)
}
func ParseObservationQC(data []byte, options *NativeObservationQcOptions) (*ObservationQcReport, error) {
	inputLength, inputErr := checkedNativeSize(len(data))
	if inputErr != nil {
		return nil, inputErr
	}
	var out *C.SidereonObservationQcReport
	var err error
	withCThread(func() {
		var p *C.SidereonObservationQcOptions
		if options != nil {
			memory, allocationErr := checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonObservationQcOptions{}))
			if allocationErr != nil {
				err = allocationErr
				return
			}
			defer C.free(memory)
			p = (*C.SidereonObservationQcOptions)(memory)
			*p = qcOptions(*options)
		}
		if err != nil {
			return
		}
		var input unsafe.Pointer
		if len(data) != 0 {
			input = C.CBytes(data)
			if input == nil {
				err = errors.New("sidereon: unable to allocate native input buffer")
				return
			}
			defer C.free(input)
		}
		err = statusErrorLocked(C.sidereon_observation_qc_parse((*C.uint8_t)(input), inputLength, p, &out))
		if err != nil && out != nil {
			C.sidereon_observation_qc_report_free(out)
			out = nil
		}
	})
	runtime.KeepAlive(data)
	if err != nil {
		return nil, err
	}
	handle, err := newObservationQcReport(out)
	if err != nil && out != nil {
		withCThread(func() { C.sidereon_observation_qc_report_free(out) })
	}
	return handle, err
}
func (report *ObservationQcReport) Close() error {
	if report == nil {
		return nil
	}
	return closeProtocolResource(report, report.resource, &report.cleanup)
}
func (report *ObservationQcReport) Summary() (NativeObservationQcSummary, error) {
	var v C.SidereonObservationQcSummary
	err := report.resource.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 { return C.sidereon_observation_qc_summary((*C.SidereonObservationQcReport)(p), &v) })
	})
	if err != nil {
		return NativeObservationQcSummary{}, err
	}
	out := NativeObservationQcSummary{HasInterval: bool(v.has_interval_s), Interval: float64(v.interval_s), IntervalSource: uint32(v.interval_source)}
	counts := []*int{&out.TotalEpochRecords, &out.ObservationEpochs, &out.EventRecords, &out.PowerFailureEpochs, &out.SkippedRecords, &out.MissingEpochs, &out.DataGapCount, &out.SatelliteCount, &out.SatelliteSignalCount, &out.SystemSignalCount, &out.NoteCount}
	nativeCounts := []C.size_t{v.total_epoch_records, v.observation_epochs, v.event_records, v.power_failure_epochs, v.skipped_records, v.missing_epochs, v.data_gap_count, v.satellite_count, v.satellite_signal_count, v.system_signal_count, v.note_count}
	for i := range counts {
		count, err := checkedNativeCount(uint64(nativeCounts[i]))
		if err != nil {
			return NativeObservationQcSummary{}, err
		}
		*counts[i] = count
	}
	runtime.KeepAlive(report)
	return out, nil
}

func (report *ObservationQcReport) Gaps() ([]NativeObservationQcDataGap, error) {
	var result []NativeObservationQcDataGap
	err := report.resource.with(func(p unsafe.Pointer) error {
		var w, r C.size_t
		f := func(out *C.SidereonObservationQcDataGap, n C.size_t) uint32 {
			return C.sidereon_observation_qc_gaps((*C.SidereonObservationQcReport)(p), out, n, &w, &r)
		}
		if err := callStatus(func() uint32 { return f(nil, 0) }); err != nil {
			return err
		}
		n, e := validateNativeQuery("QC gaps", uint64(w), uint64(r))
		if e != nil {
			return e
		}
		if _, e := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonObservationQcDataGap{})); e != nil {
			return e
		}
		v := make([]C.SidereonObservationQcDataGap, n)
		var o *C.SidereonObservationQcDataGap
		if n > 0 {
			o = &v[0]
		}
		if err := callStatus(func() uint32 { return f(o, C.size_t(n)) }); err != nil {
			return err
		}
		z, e := validateNativeOutput("QC gaps", n, uint64(w), uint64(r))
		if e != nil {
			return e
		}
		result = make([]NativeObservationQcDataGap, z)
		for i := range result {
			x := v[i]
			result[i] = NativeObservationQcDataGap{Start: calendarEpochFromC(x.start_epoch), End: calendarEpochFromC(x.end_epoch), NominalInterval: float64(x.nominal_interval_s), ObservedDelta: float64(x.observed_delta_s)}
			result[i].MissingEpochs, e = checkedNativeCount(uint64(x.missing_epochs))
			if e != nil {
				return e
			}
		}
		return nil
	})
	runtime.KeepAlive(report)
	return result, err
}
func (report *ObservationQcReport) ClockJumps() ([]NativeObservationQcClockJump, error) {
	var result []NativeObservationQcClockJump
	err := report.resource.with(func(p unsafe.Pointer) error {
		var w, r C.size_t
		f := func(o *C.SidereonObservationQcClockJump, n C.size_t) uint32 {
			return C.sidereon_observation_qc_clock_jumps((*C.SidereonObservationQcReport)(p), o, n, &w, &r)
		}
		if e := callStatus(func() uint32 { return f(nil, 0) }); e != nil {
			return e
		}
		n, e := validateNativeQuery("QC clock jumps", uint64(w), uint64(r))
		if e != nil {
			return e
		}
		if _, e := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonObservationQcClockJump{})); e != nil {
			return e
		}
		v := make([]C.SidereonObservationQcClockJump, n)
		var o *C.SidereonObservationQcClockJump
		if n > 0 {
			o = &v[0]
		}
		if e := callStatus(func() uint32 { return f(o, C.size_t(n)) }); e != nil {
			return e
		}
		z, e := validateNativeOutput("QC clock jumps", n, uint64(w), uint64(r))
		if e != nil {
			return e
		}
		result = make([]NativeObservationQcClockJump, z)
		for i := range result {
			epochIndex, err := checkedNativeCount(uint64(v[i].epoch_index))
			if err != nil {
				return err
			}
			result[i] = NativeObservationQcClockJump{EpochIndex: epochIndex, Epoch: calendarEpochFromC(v[i].epoch), DeltaS: float64(v[i].delta_s)}
		}
		return nil
	})
	runtime.KeepAlive(report)
	return result, err
}
func (report *ObservationQcReport) CycleSlips() (NativeObservationQcCycleSlips, error) {
	var v C.SidereonObservationQcCycleSlips
	err := report.resource.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 { return C.sidereon_observation_qc_cycle_slips((*C.SidereonObservationQcReport)(p), &v) })
	})
	observations, countErr := checkedNativeCount(uint64(v.observations))
	if err == nil {
		err = countErr
	}
	totalSlips, countErr := checkedNativeCount(uint64(v.total_slips))
	if err == nil {
		err = countErr
	}
	systemCount, countErr := checkedNativeCount(uint64(v.system_count))
	if err == nil {
		err = countErr
	}
	return NativeObservationQcCycleSlips{Observations: observations, TotalSlips: totalSlips, SystemCount: systemCount, HasObservationsPerSlip: bool(v.has_observations_per_slip), ObservationsPerSlip: float64(v.observations_per_slip)}, err
}

func (report *ObservationQcReport) CycleSlipSystems() ([]NativeObservationQcSystemCycleSlip, error) {
	var result []NativeObservationQcSystemCycleSlip
	err := report.resource.with(func(p unsafe.Pointer) error {
		var w, r C.size_t
		if e := callStatus(func() uint32 {
			return C.sidereon_observation_qc_cycle_slip_systems((*C.SidereonObservationQcReport)(p), nil, 0, &w, &r)
		}); e != nil {
			return e
		}
		n, e := validateNativeQuery("QC system cycle slips", uint64(w), uint64(r))
		if e != nil {
			return e
		}
		if _, e := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonObservationQcSystemCycleSlip{})); e != nil {
			return e
		}
		v := make([]C.SidereonObservationQcSystemCycleSlip, n)
		var q *C.SidereonObservationQcSystemCycleSlip
		if n > 0 {
			q = &v[0]
		}
		if e = callStatus(func() uint32 {
			return C.sidereon_observation_qc_cycle_slip_systems((*C.SidereonObservationQcReport)(p), q, C.size_t(n), &w, &r)
		}); e != nil {
			return e
		}
		z, e := validateNativeOutput("QC system cycle slips", n, uint64(w), uint64(r))
		if e != nil {
			return e
		}
		result = make([]NativeObservationQcSystemCycleSlip, z)
		for i := range result {
			if err := validateGNSSSystemValue(uint32(v[i].system)); err != nil {
				return err
			}
			result[i] = NativeObservationQcSystemCycleSlip{System: uint32(v[i].system), HasObservationsPerSlip: bool(v[i].has_observations_per_slip), ObservationsPerSlip: float64(v[i].observations_per_slip)}
			result[i].Observations, e = checkedNativeCount(uint64(v[i].observations))
			if e != nil {
				return e
			}
			result[i].Slips, e = checkedNativeCount(uint64(v[i].slips))
			if e != nil {
				return e
			}
		}
		return nil
	})
	runtime.KeepAlive(report)
	return result, err
}

func (report *ObservationQcReport) Satellites() ([]NativeObservationQcSatellite, error) {
	var result []NativeObservationQcSatellite
	err := report.resource.with(func(p unsafe.Pointer) error {
		var w, r C.size_t
		if e := callStatus(func() uint32 {
			return C.sidereon_observation_qc_satellites((*C.SidereonObservationQcReport)(p), nil, 0, &w, &r)
		}); e != nil {
			return e
		}
		n, e := validateNativeQuery("QC satellites", uint64(w), uint64(r))
		if e != nil {
			return e
		}
		if _, e := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonObservationQcSatellite{})); e != nil {
			return e
		}
		v := make([]C.SidereonObservationQcSatellite, n)
		var q *C.SidereonObservationQcSatellite
		if n > 0 {
			q = &v[0]
		}
		if e = callStatus(func() uint32 {
			return C.sidereon_observation_qc_satellites((*C.SidereonObservationQcReport)(p), q, C.size_t(n), &w, &r)
		}); e != nil {
			return e
		}
		z, e := validateNativeOutput("QC satellites", n, uint64(w), uint64(r))
		if e != nil {
			return e
		}
		result = make([]NativeObservationQcSatellite, z)
		for i := range result {
			result[i] = NativeObservationQcSatellite{SatelliteID: tokenFromC(v[i].sat_id)}
			result[i].EpochsWithObservations, e = checkedNativeCount(uint64(v[i].epochs_with_observations))
			if e != nil {
				return e
			}
			result[i].ValueObservations, e = checkedNativeCount(uint64(v[i].value_observations))
			if e != nil {
				return e
			}
		}
		return nil
	})
	runtime.KeepAlive(report)
	return result, err
}

func (report *ObservationQcReport) Signals(system bool) ([]NativeObservationQcSignal, error) {
	var result []NativeObservationQcSignal
	err := report.resource.with(func(p unsafe.Pointer) error {
		var w, r C.size_t
		call := func(q *C.SidereonObservationQcSignal, n C.size_t) uint32 {
			if system {
				return C.sidereon_observation_qc_system_signals((*C.SidereonObservationQcReport)(p), q, n, &w, &r)
			}
			return C.sidereon_observation_qc_satellite_signals((*C.SidereonObservationQcReport)(p), q, n, &w, &r)
		}
		if e := callStatus(func() uint32 { return call(nil, 0) }); e != nil {
			return e
		}
		n, e := validateNativeQuery("QC signals", uint64(w), uint64(r))
		if e != nil {
			return e
		}
		if _, e := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonObservationQcSignal{})); e != nil {
			return e
		}
		v := make([]C.SidereonObservationQcSignal, n)
		var q *C.SidereonObservationQcSignal
		if n > 0 {
			q = &v[0]
		}
		if e = callStatus(func() uint32 { return call(q, C.size_t(n)) }); e != nil {
			return e
		}
		z, e := validateNativeOutput("QC signals", n, uint64(w), uint64(r))
		if e != nil {
			return e
		}
		result = make([]NativeObservationQcSignal, z)
		for i := range result {
			x := v[i]
			if err := validateGNSSSystemValue(uint32(x.system)); err != nil {
				return err
			}
			result[i] = NativeObservationQcSignal{SatelliteID: tokenFromC(x.sat_id), System: uint32(x.system), Code: observationFixedString(x.code[:]), HasSSI: bool(x.has_ssi), HasSNR: bool(x.has_snr), SNRMean: float64(x.snr_mean), SNRMin: float64(x.snr_min), SNRMax: float64(x.snr_max), HasSNRStd: bool(x.has_snr_std), SNRStd: float64(x.snr_std)}
			result[i].ValueObservations, e = checkedNativeCount(uint64(x.value_observations))
			if e != nil {
				return e
			}
			result[i].SNRN, e = checkedNativeCount(uint64(x.snr_n))
			if e != nil {
				return e
			}
			for j := range result[i].SSICounts {
				result[i].SSICounts[j] = uint64(x.ssi_counts[j])
			}
		}
		return nil
	})
	runtime.KeepAlive(report)
	return result, err
}

func (report *ObservationQcReport) Multipath(systems bool) (any, error) {
	if systems {
		var result []NativeObservationQcSystemMultipath
		err := report.resource.with(func(p unsafe.Pointer) error {
			var w, r C.size_t
			if e := callStatus(func() uint32 {
				return C.sidereon_observation_qc_multipath_systems((*C.SidereonObservationQcReport)(p), nil, 0, &w, &r)
			}); e != nil {
				return e
			}
			n, e := validateNativeQuery("QC system multipath", uint64(w), uint64(r))
			if e != nil {
				return e
			}
			if _, e := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonObservationQcSystemMultipath{})); e != nil {
				return e
			}
			v := make([]C.SidereonObservationQcSystemMultipath, n)
			var q *C.SidereonObservationQcSystemMultipath
			if n > 0 {
				q = &v[0]
			}
			if e = callStatus(func() uint32 {
				return C.sidereon_observation_qc_multipath_systems((*C.SidereonObservationQcReport)(p), q, C.size_t(n), &w, &r)
			}); e != nil {
				return e
			}
			z, e := validateNativeOutput("QC system multipath", n, uint64(w), uint64(r))
			if e != nil {
				return e
			}
			result = make([]NativeObservationQcSystemMultipath, z)
			for i := range result {
				x := v[i]
				mp1, e := checkedNativeCount(uint64(x.mp1.n))
				if e != nil {
					return e
				}
				mp2, e := checkedNativeCount(uint64(x.mp2.n))
				if e != nil {
					return e
				}
				if err := validateGNSSSystemValue(uint32(x.system)); err != nil {
					return err
				}
				result[i] = NativeObservationQcSystemMultipath{System: uint32(x.system), HasMP1: bool(x.has_mp1), HasMP2: bool(x.has_mp2), MP1: NativeObservationQcMpStats{N: mp1, RMSM: float64(x.mp1.rms_m)}, MP2: NativeObservationQcMpStats{N: mp2, RMSM: float64(x.mp2.rms_m)}}
			}
			return nil
		})
		runtime.KeepAlive(report)
		return result, err
	}
	var result []NativeObservationQcSatelliteMultipath
	err := report.resource.with(func(p unsafe.Pointer) error {
		var w, r C.size_t
		if e := callStatus(func() uint32 {
			return C.sidereon_observation_qc_multipath_satellites((*C.SidereonObservationQcReport)(p), nil, 0, &w, &r)
		}); e != nil {
			return e
		}
		n, e := validateNativeQuery("QC satellite multipath", uint64(w), uint64(r))
		if e != nil {
			return e
		}
		if _, e := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonObservationQcSatelliteMultipath{})); e != nil {
			return e
		}
		v := make([]C.SidereonObservationQcSatelliteMultipath, n)
		var q *C.SidereonObservationQcSatelliteMultipath
		if n > 0 {
			q = &v[0]
		}
		if e = callStatus(func() uint32 {
			return C.sidereon_observation_qc_multipath_satellites((*C.SidereonObservationQcReport)(p), q, C.size_t(n), &w, &r)
		}); e != nil {
			return e
		}
		z, e := validateNativeOutput("QC satellite multipath", n, uint64(w), uint64(r))
		if e != nil {
			return e
		}
		result = make([]NativeObservationQcSatelliteMultipath, z)
		for i := range result {
			x := v[i]
			mp1, e := checkedNativeCount(uint64(x.mp1.n))
			if e != nil {
				return e
			}
			mp2, e := checkedNativeCount(uint64(x.mp2.n))
			if e != nil {
				return e
			}
			result[i] = NativeObservationQcSatelliteMultipath{SatelliteID: tokenFromC(x.sat_id), HasMP1: bool(x.has_mp1), HasMP2: bool(x.has_mp2), MP1: NativeObservationQcMpStats{N: mp1, RMSM: float64(x.mp1.rms_m)}, MP2: NativeObservationQcMpStats{N: mp2, RMSM: float64(x.mp2.rms_m)}}
		}
		return nil
	})
	runtime.KeepAlive(report)
	return result, err
}
func (report *ObservationQcReport) Text() ([]byte, error) {
	var result []byte
	err := report.resource.with(func(p unsafe.Pointer) error {
		var e error
		result, e = copyByteOutput("QC text", func(o *C.uint8_t, n C.size_t, w, r *C.size_t) uint32 {
			return C.sidereon_observation_qc_render_text((*C.SidereonObservationQcReport)(p), o, n, w, r)
		})
		return e
	})
	runtime.KeepAlive(report)
	return result, err
}
func (report *ObservationQcReport) HTML() ([]byte, error) {
	var result []byte
	err := report.resource.with(func(p unsafe.Pointer) error {
		var e error
		result, e = copyByteOutput("QC HTML", func(o *C.uint8_t, n C.size_t, w, r *C.size_t) uint32 {
			return C.sidereon_observation_qc_render_html((*C.SidereonObservationQcReport)(p), o, n, w, r)
		})
		return e
	})
	runtime.KeepAlive(report)
	return result, err
}
func (report *ObservationQcReport) JSON() ([]byte, error) {
	var result []byte
	err := report.resource.with(func(p unsafe.Pointer) error {
		var e error
		result, e = copyByteOutput("QC JSON", func(o *C.uint8_t, n C.size_t, w, r *C.size_t) uint32 {
			return C.sidereon_observation_qc_to_json((*C.SidereonObservationQcReport)(p), o, n, w, r)
		})
		return e
	})
	runtime.KeepAlive(report)
	return result, err
}
