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
	"runtime"
	"sort"
	"unsafe"
)

const (
	PreciseSamplesErrorNone                    = uint32(C.SIDEREON_PRECISE_SAMPLES_ERROR_KIND_NONE)
	PreciseSamplesErrorEmpty                   = uint32(C.SIDEREON_PRECISE_SAMPLES_ERROR_KIND_EMPTY)
	PreciseSamplesErrorSingleSampleSatellite   = uint32(C.SIDEREON_PRECISE_SAMPLES_ERROR_KIND_SINGLE_SAMPLE_SATELLITE)
	PreciseSamplesErrorNonMonotonicEpochs      = uint32(C.SIDEREON_PRECISE_SAMPLES_ERROR_KIND_NON_MONOTONIC_EPOCHS)
	PreciseSamplesErrorMixedTimeScales         = uint32(C.SIDEREON_PRECISE_SAMPLES_ERROR_KIND_MIXED_TIME_SCALES)
	PreciseSamplesErrorEpochNotRepresentable   = uint32(C.SIDEREON_PRECISE_SAMPLES_ERROR_KIND_EPOCH_NOT_REPRESENTABLE)
	PreciseSamplesErrorNonFiniteSample         = uint32(C.SIDEREON_PRECISE_SAMPLES_ERROR_KIND_NON_FINITE_SAMPLE)
	PreciseSamplesErrorAccuracySamplesMismatch = uint32(C.SIDEREON_PRECISE_SAMPLES_ERROR_KIND_ACCURACY_SAMPLES_MISMATCH)
	PreciseSamplesErrorInvalidAccuracyValue    = uint32(C.SIDEREON_PRECISE_SAMPLES_ERROR_KIND_INVALID_ACCURACY_VALUE)
	PreciseSamplesErrorOther                   = uint32(C.SIDEREON_PRECISE_SAMPLES_ERROR_KIND_OTHER)
)

type SP3AccuracyValue struct {
	Kind  uint32
	Value float64
}

type SP3AccuracyCodeGroup struct {
	HasAxisExponents        [3]bool
	AxisExponents           [3]int16
	HasClockExponent        bool
	ClockExponent           int16
	HasPositionVelocityBase bool
	PositionVelocityBase    float64
	HasClockRateBase        bool
	ClockRateBase           float64
}

type SP3RawRecordAccuracy struct {
	HasP bool
	P    SP3AccuracyCodeGroup
	HasV bool
	V    SP3AccuracyCodeGroup
}

type SP3PositionClockAccuracy struct {
	PositionSigmaM     [3]SP3AccuracyValue
	ClockSigmaM        SP3AccuracyValue
	PositionVarianceM2 [3]SP3AccuracyValue
	ClockVarianceM2    SP3AccuracyValue
}

type SP3VelocityAccuracy struct {
	VelocitySigmaMPerS       [3]SP3AccuracyValue
	ClockRateSigmaMPerS      SP3AccuracyValue
	VelocityVarianceM2PerS2  [3]SP3AccuracyValue
	ClockRateVarianceM2PerS2 SP3AccuracyValue
}

type SP3RecordAccuracy struct {
	HasP bool
	P    SP3PositionClockAccuracy
	HasV bool
	V    SP3VelocityAccuracy
}

type PreciseEphemerisAccuracySample struct {
	Satellite          string
	TimeScale          uint32
	EpochJ2000S        float64
	PositionVarianceM2 [3]SP3AccuracyValue
	ClockVarianceM2    SP3AccuracyValue
}

type PreciseEphemerisSampleV2 struct {
	Satellite     string
	Epoch         NativeClockEpoch
	PositionECEFM [3]float64
	HasClock      bool
	ClockS        float64
	ClockEvent    bool
}

type PreciseEphemerisAccuracySampleV2 struct {
	Satellite          string
	Epoch              NativeClockEpoch
	PositionVarianceM2 [3]SP3AccuracyValue
	ClockVarianceM2    SP3AccuracyValue
}

type PreciseSamplesError struct {
	Kind         uint32
	HasSatellite bool
	Satellite    string
}

type SSRCorrectionSize struct {
	OrbitM float64
	ClockM float64
}

type SSRCorrectedState struct {
	HasState                bool
	PositionECEFM           [3]float64
	ClockS                  float64
	HasGroupDelay           bool
	GroupDelayS             float64
	Degraded                bool
	DegradeReason           uint32
	HasSizeEvent            bool
	StrictRefusal           bool
	Size                    SSRCorrectionSize
	HasOversizedReport      bool
	Source                  uint32
	ProviderID              uint16
	SolutionID              uint8
	OrbitRefEpochJ2000S     float64
	ClockRefEpochJ2000S     float64
	FirstAppliedEpochJ2000S float64
}

func (s *SP3) StateAtEpochQuery(query *ExactEpochQuery, satellite string) (SP3State, error) {
	if s == nil || s.handle == nil || query == nil || query.handle == nil {
		return SP3State{}, ErrClosed
	}
	sat, err := tokenToC(satellite)
	if err != nil {
		return SP3State{}, err
	}
	var out C.SidereonSp3State
	err = withPositioningHandles([]*positioningHandle{s.handle, query.handle}, func(pointers []unsafe.Pointer) error {
		return callStatusWithSp3Diagnostics(func() uint32 {
			return uint32(C.sidereon_sp3_state_at_epoch_query(
				(*C.SidereonSp3)(pointers[0]), (*C.char)(unsafe.Pointer(&sat.bytes[0])),
				(*C.SidereonExactEpochQuery)(pointers[1]), &out,
			))
		})
	})
	runtime.KeepAlive(s)
	runtime.KeepAlive(query)
	if err != nil {
		return SP3State{}, err
	}
	return sp3StateFromC(out), nil
}

func accuracyCodeGroupFromC(value C.SidereonSp3AccuracyCodeGroup) SP3AccuracyCodeGroup {
	out := SP3AccuracyCodeGroup{
		HasClockExponent: bool(value.has_clock_exponent), ClockExponent: int16(value.clock_exponent),
		HasPositionVelocityBase: bool(value.has_position_velocity_base),
		PositionVelocityBase:    float64(value.position_velocity_base),
		HasClockRateBase:        bool(value.has_clock_rate_base), ClockRateBase: float64(value.clock_rate_base),
	}
	for axis := 0; axis < 3; axis++ {
		out.HasAxisExponents[axis] = bool(value.has_axis_exponents[axis])
		out.AxisExponents[axis] = int16(value.axis_exponents[axis])
	}
	return out
}

func (s *SP3) RecordAccuracyCodes(satellite string, epochIndex int) (SP3RawRecordAccuracy, error) {
	if s == nil || s.handle == nil {
		return SP3RawRecordAccuracy{}, ErrClosed
	}
	if epochIndex < 0 {
		return SP3RawRecordAccuracy{}, invalidArgument("SP3 epoch index must not be negative")
	}
	index, err := checkedNativeSize(epochIndex)
	if err != nil {
		return SP3RawRecordAccuracy{}, err
	}
	sat, err := tokenToC(satellite)
	if err != nil {
		return SP3RawRecordAccuracy{}, err
	}
	var value C.SidereonSp3RawRecordAccuracy
	err = s.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_sp3_record_accuracy_codes(
				(*C.SidereonSp3)(pointer), (*C.char)(unsafe.Pointer(&sat.bytes[0])), index, &value,
			))
		})
	})
	runtime.KeepAlive(s)
	if err != nil {
		return SP3RawRecordAccuracy{}, err
	}
	return SP3RawRecordAccuracy{
		HasP: bool(value.has_p), P: accuracyCodeGroupFromC(value.p),
		HasV: bool(value.has_v), V: accuracyCodeGroupFromC(value.v),
	}, nil
}

func accuracyValueFromC(value C.SidereonSp3AccuracyValue) SP3AccuracyValue {
	return SP3AccuracyValue{Kind: uint32(value.kind), Value: float64(value.value)}
}

func (s *SP3) RecordAccuracy(satellite string, epochIndex int) (SP3RecordAccuracy, error) {
	if s == nil || s.handle == nil {
		return SP3RecordAccuracy{}, ErrClosed
	}
	if epochIndex < 0 {
		return SP3RecordAccuracy{}, invalidArgument("SP3 epoch index must not be negative")
	}
	index, err := checkedNativeSize(epochIndex)
	if err != nil {
		return SP3RecordAccuracy{}, err
	}
	sat, err := tokenToC(satellite)
	if err != nil {
		return SP3RecordAccuracy{}, err
	}
	var value C.SidereonSp3RecordAccuracy
	err = s.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_sp3_record_accuracy(
				(*C.SidereonSp3)(pointer), (*C.char)(unsafe.Pointer(&sat.bytes[0])), index, &value,
			))
		})
	})
	runtime.KeepAlive(s)
	if err != nil {
		return SP3RecordAccuracy{}, err
	}
	out := SP3RecordAccuracy{HasP: bool(value.has_p), HasV: bool(value.has_v)}
	for axis := 0; axis < 3; axis++ {
		out.P.PositionSigmaM[axis] = accuracyValueFromC(value.p.position_sigma_m[axis])
		out.P.PositionVarianceM2[axis] = accuracyValueFromC(value.p.position_variance_m2[axis])
		out.V.VelocitySigmaMPerS[axis] = accuracyValueFromC(value.v.velocity_sigma_m_s[axis])
		out.V.VelocityVarianceM2PerS2[axis] = accuracyValueFromC(value.v.velocity_variance_m2_s2[axis])
	}
	out.P.ClockSigmaM = accuracyValueFromC(value.p.clock_sigma_m)
	out.P.ClockVarianceM2 = accuracyValueFromC(value.p.clock_variance_m2)
	out.V.ClockRateSigmaMPerS = accuracyValueFromC(value.v.clock_rate_sigma_m_s)
	out.V.ClockRateVarianceM2PerS2 = accuracyValueFromC(value.v.clock_rate_variance_m2_s2)
	return out, nil
}

func (s *SP3) PreciseAccuracySamples() ([]PreciseEphemerisAccuracySample, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var result []PreciseEphemerisAccuracySample
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return uint32(C.sidereon_sp3_precise_ephemeris_accuracy_samples(
				(*C.SidereonSp3)(pointer), nil, 0, &written, &required,
			))
		}); err != nil {
			return err
		}
		count, err := validateNativeQuery("SP3 accuracy samples", uint64(written), uint64(required))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonPreciseEphemerisAccuracySample{})); err != nil {
			return err
		}
		values := make([]C.SidereonPreciseEphemerisAccuracySample, count)
		var output *C.SidereonPreciseEphemerisAccuracySample
		if count != 0 {
			output = &values[0]
		}
		written, required = 0, 0
		if err := callStatus(func() uint32 {
			return uint32(C.sidereon_sp3_precise_ephemeris_accuracy_samples(
				(*C.SidereonSp3)(pointer), output, C.size_t(count), &written, &required,
			))
		}); err != nil {
			return err
		}
		writtenCount, err := validateNativeOutput("SP3 accuracy samples", count, uint64(written), uint64(required))
		if err != nil {
			return err
		}
		result = make([]PreciseEphemerisAccuracySample, writtenCount)
		for row := 0; row < writtenCount; row++ {
			if err := validTimeScale(uint32(values[row].time_scale)); err != nil {
				return err
			}
			result[row] = PreciseEphemerisAccuracySample{
				Satellite: tokenFromC(values[row].sat), TimeScale: uint32(values[row].time_scale),
				EpochJ2000S:     float64(values[row].epoch_j2000_s),
				ClockVarianceM2: accuracyValueFromC(values[row].clock_variance_m2),
			}
			for axis := 0; axis < 3; axis++ {
				result[row].PositionVarianceM2[axis] = accuracyValueFromC(values[row].position_variance_m2[axis])
			}
		}
		return nil
	})
	runtime.KeepAlive(s)
	return result, err
}

func preciseSampleV2FromC(value C.SidereonPreciseEphemerisSampleV2) PreciseEphemerisSampleV2 {
	out := PreciseEphemerisSampleV2{
		Satellite: tokenFromC(value.sat), Epoch: clockEpochFromC(value.epoch),
		HasClock: bool(value.has_clock_s), ClockS: float64(value.clock_s),
		ClockEvent: bool(value.clock_event),
	}
	for axis := range out.PositionECEFM {
		out.PositionECEFM[axis] = float64(value.position_ecef_m[axis])
	}
	return out
}

func preciseAccuracySampleV2FromC(value C.SidereonPreciseEphemerisAccuracySampleV2) PreciseEphemerisAccuracySampleV2 {
	out := PreciseEphemerisAccuracySampleV2{
		Satellite: tokenFromC(value.sat), Epoch: clockEpochFromC(value.epoch),
		ClockVarianceM2: accuracyValueFromC(value.clock_variance_m2),
	}
	for axis := range out.PositionVarianceM2 {
		out.PositionVarianceM2[axis] = accuracyValueFromC(value.position_variance_m2[axis])
	}
	return out
}

type preciseSamplesV2ReadFunc func(unsafe.Pointer, unsafe.Pointer, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus

func readPreciseSamplesV2(handle *positioningHandle, sampleRead, accuracyRead preciseSamplesV2ReadFunc, accuracy bool) (samples []PreciseEphemerisSampleV2, accuracies []PreciseEphemerisAccuracySampleV2, err error) {
	if handle == nil {
		return nil, nil, ErrClosed
	}
	err = handle.with(func(pointer unsafe.Pointer) error {
		var written, required C.size_t
		invoke := sampleRead
		if accuracy {
			invoke = accuracyRead
		}
		if callErr := callStatus(func() uint32 { return uint32(invoke(pointer, nil, 0, &written, &required)) }); callErr != nil {
			return callErr
		}
		count, countErr := validateNativeQuery("SP3 lossless samples", uint64(written), uint64(required))
		if countErr != nil {
			return countErr
		}
		var output unsafe.Pointer
		if count != 0 {
			typeSize := unsafe.Sizeof(C.SidereonPreciseEphemerisSampleV2{})
			if accuracy {
				typeSize = unsafe.Sizeof(C.SidereonPreciseEphemerisAccuracySampleV2{})
			}
			allocation, allocationErr := checkedNativeAllocationSize(count, typeSize)
			if allocationErr != nil {
				return allocationErr
			}
			output = C.malloc(C.size_t(allocation))
			if output == nil {
				return errors.New("sidereon: unable to allocate lossless SP3 sample output")
			}
			defer C.free(output)
		}
		written, required = 0, 0
		if callErr := callStatus(func() uint32 {
			return uint32(invoke(pointer, output, C.size_t(count), &written, &required))
		}); callErr != nil {
			return callErr
		}
		writtenCount, countErr := validateNativeOutput("SP3 lossless samples", count, uint64(written), uint64(required))
		if countErr != nil {
			return countErr
		}
		if accuracy {
			rows := unsafe.Slice((*C.SidereonPreciseEphemerisAccuracySampleV2)(output), writtenCount)
			accuracies = make([]PreciseEphemerisAccuracySampleV2, writtenCount)
			for row := range rows {
				if scaleErr := validTimeScale(uint32(rows[row].epoch.scale)); scaleErr != nil {
					return scaleErr
				}
				accuracies[row] = preciseAccuracySampleV2FromC(rows[row])
			}
		} else {
			rows := unsafe.Slice((*C.SidereonPreciseEphemerisSampleV2)(output), writtenCount)
			samples = make([]PreciseEphemerisSampleV2, writtenCount)
			for row := range rows {
				if scaleErr := validTimeScale(uint32(rows[row].epoch.scale)); scaleErr != nil {
					return scaleErr
				}
				samples[row] = preciseSampleV2FromC(rows[row])
			}
		}
		return nil
	})
	return samples, accuracies, err
}

func (s *SP3) PreciseSamplesV2() ([]PreciseEphemerisSampleV2, error) {
	if s == nil {
		return nil, ErrClosed
	}
	samples, _, err := readPreciseSamplesV2(s.handle,
		func(pointer, output unsafe.Pointer, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_sp3_precise_ephemeris_samples_v2((*C.SidereonSp3)(pointer), (*C.SidereonPreciseEphemerisSampleV2)(output), capacity, written, required)
		},
		func(pointer, output unsafe.Pointer, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_sp3_precise_ephemeris_accuracy_samples_v2((*C.SidereonSp3)(pointer), (*C.SidereonPreciseEphemerisAccuracySampleV2)(output), capacity, written, required)
		}, false)
	runtime.KeepAlive(s)
	return samples, err
}

func (s *SP3) PreciseAccuracySamplesV2() ([]PreciseEphemerisAccuracySampleV2, error) {
	if s == nil {
		return nil, ErrClosed
	}
	_, accuracies, err := readPreciseSamplesV2(s.handle,
		func(pointer, output unsafe.Pointer, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_sp3_precise_ephemeris_samples_v2((*C.SidereonSp3)(pointer), (*C.SidereonPreciseEphemerisSampleV2)(output), capacity, written, required)
		},
		func(pointer, output unsafe.Pointer, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_sp3_precise_ephemeris_accuracy_samples_v2((*C.SidereonSp3)(pointer), (*C.SidereonPreciseEphemerisAccuracySampleV2)(output), capacity, written, required)
		}, true)
	runtime.KeepAlive(s)
	return accuracies, err
}

func (s *PreciseEphemerisSamples) RecordsV2() ([]PreciseEphemerisSampleV2, error) {
	if s == nil {
		return nil, ErrClosed
	}
	samples, _, err := readPreciseSamplesV2(s.handle,
		func(pointer, output unsafe.Pointer, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_precise_ephemeris_samples_records_v2((*C.SidereonPreciseEphemerisSamples)(pointer), (*C.SidereonPreciseEphemerisSampleV2)(output), capacity, written, required)
		},
		func(pointer, output unsafe.Pointer, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_precise_ephemeris_samples_accuracy_records_v2((*C.SidereonPreciseEphemerisSamples)(pointer), (*C.SidereonPreciseEphemerisAccuracySampleV2)(output), capacity, written, required)
		}, false)
	runtime.KeepAlive(s)
	return samples, err
}

func (s *PreciseEphemerisSamples) AccuracyRecordsV2() ([]PreciseEphemerisAccuracySampleV2, error) {
	if s == nil {
		return nil, ErrClosed
	}
	_, accuracies, err := readPreciseSamplesV2(s.handle,
		func(pointer, output unsafe.Pointer, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_precise_ephemeris_samples_records_v2((*C.SidereonPreciseEphemerisSamples)(pointer), (*C.SidereonPreciseEphemerisSampleV2)(output), capacity, written, required)
		},
		func(pointer, output unsafe.Pointer, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_precise_ephemeris_samples_accuracy_records_v2((*C.SidereonPreciseEphemerisSamples)(pointer), (*C.SidereonPreciseEphemerisAccuracySampleV2)(output), capacity, written, required)
		}, true)
	runtime.KeepAlive(s)
	return accuracies, err
}

func cPreciseSampleV2(value PreciseEphemerisSampleV2) (C.SidereonPreciseEphemerisSampleV2, error) {
	if err := validTimeScale(value.Epoch.Scale); err != nil {
		return C.SidereonPreciseEphemerisSampleV2{}, err
	}
	satellite, err := tokenToC(value.Satellite)
	if err != nil {
		return C.SidereonPreciseEphemerisSampleV2{}, err
	}
	out := C.SidereonPreciseEphemerisSampleV2{
		sat: satellite,
		epoch: C.SidereonClockEpoch{
			scale: C.uint32_t(value.Epoch.Scale), representation: C.uint32_t(value.Epoch.Representation),
			jd_whole: C.double(value.Epoch.JulianWhole), jd_fraction: C.double(value.Epoch.JulianFraction),
			nanos_high: C.int64_t(value.Epoch.NanosHigh), nanos_low: C.uint64_t(value.Epoch.NanosLow),
		},
		has_clock_s: C.bool(value.HasClock), clock_s: C.double(value.ClockS), clock_event: C.bool(value.ClockEvent),
	}
	for axis := range value.PositionECEFM {
		out.position_ecef_m[axis] = C.double(value.PositionECEFM[axis])
	}
	return out, nil
}

func cPreciseAccuracySampleV2(value PreciseEphemerisAccuracySampleV2) (C.SidereonPreciseEphemerisAccuracySampleV2, error) {
	if err := validTimeScale(value.Epoch.Scale); err != nil {
		return C.SidereonPreciseEphemerisAccuracySampleV2{}, err
	}
	satellite, err := tokenToC(value.Satellite)
	if err != nil {
		return C.SidereonPreciseEphemerisAccuracySampleV2{}, err
	}
	out := C.SidereonPreciseEphemerisAccuracySampleV2{
		sat: satellite,
		epoch: C.SidereonClockEpoch{
			scale: C.uint32_t(value.Epoch.Scale), representation: C.uint32_t(value.Epoch.Representation),
			jd_whole: C.double(value.Epoch.JulianWhole), jd_fraction: C.double(value.Epoch.JulianFraction),
			nanos_high: C.int64_t(value.Epoch.NanosHigh), nanos_low: C.uint64_t(value.Epoch.NanosLow),
		},
		clock_variance_m2: C.SidereonSp3AccuracyValue{
			kind: C.uint32_t(value.ClockVarianceM2.Kind), value: C.double(value.ClockVarianceM2.Value),
		},
	}
	for axis := range value.PositionVarianceM2 {
		out.position_variance_m2[axis] = C.SidereonSp3AccuracyValue{
			kind:  C.uint32_t(value.PositionVarianceM2[axis].Kind),
			value: C.double(value.PositionVarianceM2[axis].Value),
		}
	}
	return out, nil
}

func buildPreciseSamplesV2(samples []PreciseEphemerisSampleV2, accuracies []PreciseEphemerisAccuracySampleV2, gap float64, withAccuracy, interpolant bool) (*PreciseEphemerisSamples, *PreciseEphemerisInterpolant, PreciseSamplesError, error) {
	if withAccuracy && len(samples) != len(accuracies) {
		return nil, nil, PreciseSamplesError{Kind: PreciseSamplesErrorAccuracySamplesMismatch}, invalidArgument("precise sample and accuracy counts differ")
	}
	sampleSize, err := checkedNativeAllocationSize(len(samples), unsafe.Sizeof(C.SidereonPreciseEphemerisSampleV2{}))
	if err != nil {
		return nil, nil, PreciseSamplesError{}, err
	}
	var sampleMemory unsafe.Pointer
	if sampleSize != 0 {
		sampleMemory = C.malloc(C.size_t(sampleSize))
		if sampleMemory == nil {
			return nil, nil, PreciseSamplesError{}, errors.New("sidereon: unable to allocate lossless precise samples")
		}
	}
	defer C.free(sampleMemory)
	sampleRows := unsafe.Slice((*C.SidereonPreciseEphemerisSampleV2)(sampleMemory), len(samples))
	for row := range samples {
		converted, convertErr := cPreciseSampleV2(samples[row])
		if convertErr != nil {
			return nil, nil, PreciseSamplesError{}, convertErr
		}
		sampleRows[row] = converted
	}
	var accuracyMemory unsafe.Pointer
	if withAccuracy {
		accuracySize, sizeErr := checkedNativeAllocationSize(len(accuracies), unsafe.Sizeof(C.SidereonPreciseEphemerisAccuracySampleV2{}))
		if sizeErr != nil {
			return nil, nil, PreciseSamplesError{}, sizeErr
		}
		if accuracySize != 0 {
			accuracyMemory = C.malloc(C.size_t(accuracySize))
			if accuracyMemory == nil {
				return nil, nil, PreciseSamplesError{}, errors.New("sidereon: unable to allocate lossless precise accuracy samples")
			}
		}
		defer C.free(accuracyMemory)
		accuracyRows := unsafe.Slice((*C.SidereonPreciseEphemerisAccuracySampleV2)(accuracyMemory), len(accuracies))
		for row := range accuracies {
			converted, convertErr := cPreciseAccuracySampleV2(accuracies[row])
			if convertErr != nil {
				return nil, nil, PreciseSamplesError{}, convertErr
			}
			accuracyRows[row] = converted
		}
	}
	count, err := checkedNativeSize(len(samples))
	if err != nil {
		return nil, nil, PreciseSamplesError{}, err
	}
	var nativeError C.SidereonPreciseSamplesError
	var samplesOut *C.SidereonPreciseEphemerisSamples
	var interpolantOut *C.SidereonPreciseEphemerisInterpolant
	err = callStatus(func() uint32 {
		if interpolant {
			if withAccuracy {
				return uint32(C.sidereon_precise_ephemeris_interpolant_from_samples_with_accuracy_v2(
					(*C.SidereonPreciseEphemerisSampleV2)(sampleMemory), (*C.SidereonPreciseEphemerisAccuracySampleV2)(accuracyMemory),
					C.size_t(count), C.double(gap), &interpolantOut, &nativeError,
				))
			}
			return uint32(C.sidereon_precise_ephemeris_interpolant_from_samples_v2(
				(*C.SidereonPreciseEphemerisSampleV2)(sampleMemory), C.size_t(count), C.double(gap), &interpolantOut,
			))
		}
		if withAccuracy {
			return uint32(C.sidereon_precise_ephemeris_samples_from_samples_with_accuracy_v2(
				(*C.SidereonPreciseEphemerisSampleV2)(sampleMemory), (*C.SidereonPreciseEphemerisAccuracySampleV2)(accuracyMemory),
				C.size_t(count), C.double(gap), &samplesOut, &nativeError,
			))
		}
		return uint32(C.sidereon_precise_ephemeris_samples_from_samples_v2(
			(*C.SidereonPreciseEphemerisSampleV2)(sampleMemory), C.size_t(count), C.double(gap), &samplesOut,
		))
	})
	if err != nil {
		if samplesOut != nil {
			withCThread(func() { C.sidereon_precise_ephemeris_samples_free(samplesOut) })
		}
		if interpolantOut != nil {
			withCThread(func() { C.sidereon_precise_ephemeris_interpolant_free(interpolantOut) })
		}
		return nil, nil, preciseSamplesErrorFromC(nativeError), err
	}
	if interpolant {
		interpolantHandle, handleErr := newPreciseInterpolant(interpolantOut)
		return nil, interpolantHandle, preciseSamplesErrorFromC(nativeError), handleErr
	}
	samplesHandle, handleErr := newPreciseSamples(samplesOut)
	return samplesHandle, nil, preciseSamplesErrorFromC(nativeError), handleErr
}

func PreciseEphemerisSamplesFromSamplesV2(samples []PreciseEphemerisSampleV2, gap float64) (*PreciseEphemerisSamples, PreciseSamplesError, error) {
	handle, _, nativeError, err := buildPreciseSamplesV2(samples, nil, gap, false, false)
	return handle, nativeError, err
}

func PreciseEphemerisInterpolantFromSamplesV2(samples []PreciseEphemerisSampleV2, gap float64) (*PreciseEphemerisInterpolant, PreciseSamplesError, error) {
	_, handle, nativeError, err := buildPreciseSamplesV2(samples, nil, gap, false, true)
	return handle, nativeError, err
}

func PreciseEphemerisSamplesFromSamplesWithAccuracyV2(samples []PreciseEphemerisSampleV2, accuracies []PreciseEphemerisAccuracySampleV2, gap float64) (*PreciseEphemerisSamples, PreciseSamplesError, error) {
	handle, _, nativeError, err := buildPreciseSamplesV2(samples, accuracies, gap, true, false)
	return handle, nativeError, err
}

func PreciseEphemerisInterpolantFromSamplesWithAccuracyV2(samples []PreciseEphemerisSampleV2, accuracies []PreciseEphemerisAccuracySampleV2, gap float64) (*PreciseEphemerisInterpolant, PreciseSamplesError, error) {
	_, handle, nativeError, err := buildPreciseSamplesV2(samples, accuracies, gap, true, true)
	return handle, nativeError, err
}

func cAccuracySample(value PreciseEphemerisAccuracySample) (C.SidereonPreciseEphemerisAccuracySample, error) {
	if err := validTimeScale(value.TimeScale); err != nil {
		return C.SidereonPreciseEphemerisAccuracySample{}, err
	}
	satellite, err := tokenToC(value.Satellite)
	if err != nil {
		return C.SidereonPreciseEphemerisAccuracySample{}, err
	}
	out := C.SidereonPreciseEphemerisAccuracySample{
		sat: satellite, time_scale: C.uint32_t(value.TimeScale), epoch_j2000_s: C.double(value.EpochJ2000S),
		clock_variance_m2: C.SidereonSp3AccuracyValue{kind: C.uint32_t(value.ClockVarianceM2.Kind), value: C.double(value.ClockVarianceM2.Value)},
	}
	for axis := 0; axis < 3; axis++ {
		out.position_variance_m2[axis] = C.SidereonSp3AccuracyValue{
			kind: C.uint32_t(value.PositionVarianceM2[axis].Kind), value: C.double(value.PositionVarianceM2[axis].Value),
		}
	}
	return out, nil
}

func cAccuracies(values []PreciseEphemerisAccuracySample, count int) (unsafe.Pointer, C.size_t, error) {
	if len(values) != count {
		return nil, 0, invalidArgument("precise sample and accuracy counts differ")
	}
	allocation, err := checkedNativeAllocationSize(len(values), unsafe.Sizeof(C.SidereonPreciseEphemerisAccuracySample{}))
	if err != nil {
		return nil, 0, err
	}
	if allocation == 0 {
		return nil, 0, nil
	}
	pointer := C.malloc(C.size_t(allocation))
	if pointer == nil {
		return nil, 0, errors.New("sidereon: unable to allocate native precise accuracy samples")
	}
	rows := unsafe.Slice((*C.SidereonPreciseEphemerisAccuracySample)(pointer), len(values))
	for row := range values {
		converted, convertErr := cAccuracySample(values[row])
		if convertErr != nil {
			C.free(pointer)
			return nil, 0, convertErr
		}
		rows[row] = converted
	}
	return pointer, C.size_t(len(values)), nil
}

func preciseSamplesErrorFromC(value C.SidereonPreciseSamplesError) PreciseSamplesError {
	return PreciseSamplesError{
		Kind: uint32(value.kind), HasSatellite: bool(value.has_satellite), Satellite: tokenFromC(value.satellite),
	}
}

func buildPreciseWithAccuracy(samples []PreciseEphemerisSample, accuracy []PreciseEphemerisAccuracySample, gap float64, interpolant bool) (*PreciseEphemerisSamples, *PreciseEphemerisInterpolant, PreciseSamplesError, error) {
	if len(samples) != len(accuracy) {
		return nil, nil, PreciseSamplesError{Kind: PreciseSamplesErrorAccuracySamplesMismatch}, invalidArgument("precise sample and accuracy counts differ")
	}
	allocation, err := checkedNativeAllocationSize(len(samples), unsafe.Sizeof(C.SidereonPreciseEphemerisSample{}))
	if err != nil {
		return nil, nil, PreciseSamplesError{}, err
	}
	var samplePointer unsafe.Pointer
	if allocation != 0 {
		samplePointer = C.malloc(C.size_t(allocation))
		if samplePointer == nil {
			return nil, nil, PreciseSamplesError{}, errors.New("sidereon: unable to allocate native precise samples")
		}
		rows := unsafe.Slice((*C.SidereonPreciseEphemerisSample)(samplePointer), len(samples))
		for row := range samples {
			converted, convertErr := cPreciseSample(samples[row])
			if convertErr != nil {
				C.free(samplePointer)
				return nil, nil, PreciseSamplesError{}, convertErr
			}
			rows[row] = converted
		}
	}
	defer C.free(samplePointer)
	accuracyPointer, _, err := cAccuracies(accuracy, len(samples))
	if err != nil {
		return nil, nil, PreciseSamplesError{}, err
	}
	defer C.free(accuracyPointer)
	count, err := checkedNativeSize(len(samples))
	if err != nil {
		return nil, nil, PreciseSamplesError{}, err
	}
	var nativeError C.SidereonPreciseSamplesError
	if interpolant {
		var out *C.SidereonPreciseEphemerisInterpolant
		err = callStatus(func() uint32 {
			return uint32(C.sidereon_precise_ephemeris_interpolant_from_samples_with_accuracy(
				(*C.SidereonPreciseEphemerisSample)(samplePointer),
				(*C.SidereonPreciseEphemerisAccuracySample)(accuracyPointer), C.size_t(count),
				C.double(gap), &out, &nativeError,
			))
		})
		if err != nil {
			if out != nil {
				withCThread(func() { C.sidereon_precise_ephemeris_interpolant_free(out) })
			}
			return nil, nil, preciseSamplesErrorFromC(nativeError), err
		}
		handle, err := newPreciseInterpolant(out)
		return nil, handle, preciseSamplesErrorFromC(nativeError), err
	}
	var out *C.SidereonPreciseEphemerisSamples
	err = callStatus(func() uint32 {
		return uint32(C.sidereon_precise_ephemeris_samples_from_samples_with_accuracy(
			(*C.SidereonPreciseEphemerisSample)(samplePointer),
			(*C.SidereonPreciseEphemerisAccuracySample)(accuracyPointer), C.size_t(count),
			C.double(gap), &out, &nativeError,
		))
	})
	if err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_precise_ephemeris_samples_free(out) })
		}
		return nil, nil, preciseSamplesErrorFromC(nativeError), err
	}
	handle, err := newPreciseSamples(out)
	return handle, nil, preciseSamplesErrorFromC(nativeError), err
}

func PreciseEphemerisSamplesFromSamplesWithAccuracy(samples []PreciseEphemerisSample, accuracy []PreciseEphemerisAccuracySample, gap float64) (*PreciseEphemerisSamples, PreciseSamplesError, error) {
	samplesHandle, _, nativeError, err := buildPreciseWithAccuracy(samples, accuracy, gap, false)
	return samplesHandle, nativeError, err
}

func PreciseEphemerisInterpolantFromSamplesWithAccuracy(samples []PreciseEphemerisSample, accuracy []PreciseEphemerisAccuracySample, gap float64) (*PreciseEphemerisInterpolant, PreciseSamplesError, error) {
	_, interpolant, nativeError, err := buildPreciseWithAccuracy(samples, accuracy, gap, true)
	return interpolant, nativeError, err
}

func (i *PreciseEphemerisInterpolant) StateAtEpochQuery(query *ExactEpochQuery, satellite string) (SP3State, error) {
	if i == nil {
		return SP3State{}, ErrClosed
	}
	return preciseStateAtEpochQuery(i.handle, query, satellite, true, func(pointer unsafe.Pointer, sat *C.char, q *C.SidereonExactEpochQuery, out *C.SidereonSp3State) uint32 {
		return uint32(C.sidereon_precise_ephemeris_interpolant_state_at_epoch_query((*C.SidereonPreciseEphemerisInterpolant)(pointer), sat, q, out))
	})
}

func (a *PreciseInterpolantArtifact) StateAtEpochQuery(query *ExactEpochQuery, satellite string) (SP3State, error) {
	if a == nil {
		return SP3State{}, ErrClosed
	}
	return preciseStateAtEpochQuery(a.handle, query, satellite, false, func(pointer unsafe.Pointer, sat *C.char, q *C.SidereonExactEpochQuery, out *C.SidereonSp3State) uint32 {
		return uint32(C.sidereon_precise_interpolant_artifact_state_at_epoch_query((*C.SidereonPreciseInterpolantArtifact)(pointer), sat, q, out))
	})
}

func preciseStateAtEpochQuery(handle *positioningHandle, query *ExactEpochQuery, satellite string, captureSp3Error bool, invoke func(unsafe.Pointer, *C.char, *C.SidereonExactEpochQuery, *C.SidereonSp3State) uint32) (SP3State, error) {
	if handle == nil || query == nil || query.handle == nil {
		return SP3State{}, ErrClosed
	}
	sat, err := tokenToC(satellite)
	if err != nil {
		return SP3State{}, err
	}
	var out C.SidereonSp3State
	err = withPositioningHandles([]*positioningHandle{handle, query.handle}, func(pointers []unsafe.Pointer) error {
		capture := callStatus
		if captureSp3Error {
			capture = callStatusWithSp3Diagnostics
		}
		return capture(func() uint32 {
			return invoke(pointers[0], (*C.char)(unsafe.Pointer(&sat.bytes[0])), (*C.SidereonExactEpochQuery)(pointers[1]), &out)
		})
	})
	runtime.KeepAlive(query)
	if err != nil {
		return SP3State{}, err
	}
	return sp3StateFromC(out), nil
}

func withResources(resources []*resource, fn func([]unsafe.Pointer) error) error {
	unique := make([]*resource, 0, len(resources))
	seen := make(map[*resource]struct{}, len(resources))
	for _, value := range resources {
		if value == nil {
			return ErrClosed
		}
		if _, ok := seen[value]; !ok {
			seen[value] = struct{}{}
			unique = append(unique, value)
		}
	}
	sort.Slice(unique, func(i, j int) bool {
		return uintptr(unsafe.Pointer(unique[i])) < uintptr(unsafe.Pointer(unique[j]))
	})
	for _, value := range unique {
		value.mu.RLock()
	}
	defer func() {
		for index := len(unique) - 1; index >= 0; index-- {
			unique[index].mu.RUnlock()
		}
	}()
	pointers := make([]unsafe.Pointer, len(resources))
	for index, value := range resources {
		if value.ptr == nil {
			return ErrClosed
		}
		pointers[index] = value.ptr
	}
	return fn(pointers)
}

func CorrectedStateAtEpochQueries(broadcast *BroadcastEphemeris, store *SSRCorrectionStore, satellite string, stateEpoch, selectionEpoch *ExactEpochQuery, staleness float64, missing uint32, allowRegional bool, provider uint16, sizePolicy uint32) (SSRCorrectedState, error) {
	if broadcast == nil || store == nil || stateEpoch == nil || selectionEpoch == nil || stateEpoch.handle == nil || selectionEpoch.handle == nil {
		return SSRCorrectedState{}, ErrClosed
	}
	sat, err := tokenToC(satellite)
	if err != nil {
		return SSRCorrectedState{}, err
	}
	var out C.SidereonSsrCorrectedStateResult
	err = withResources([]*resource{broadcast.resource, store.resource}, func(pointers []unsafe.Pointer) error {
		return withPositioningHandles([]*positioningHandle{stateEpoch.handle, selectionEpoch.handle}, func(epochPointers []unsafe.Pointer) error {
			return rtcmCallStatus(func() uint32 {
				return uint32(C.sidereon_ssr_corrected_state_at_epoch_queries(
					(*C.SidereonBroadcastEphemeris)(pointers[0]), (*C.SidereonSsrCorrectionStore)(pointers[1]),
					(*C.char)(unsafe.Pointer(&sat.bytes[0])),
					(*C.SidereonExactEpochQuery)(epochPointers[0]), (*C.SidereonExactEpochQuery)(epochPointers[1]),
					C.double(staleness), C.uint32_t(missing), C.bool(allowRegional), C.uint16_t(provider),
					C.uint32_t(sizePolicy), &out,
				))
			})
		})
	})
	runtime.KeepAlive(broadcast)
	runtime.KeepAlive(store)
	runtime.KeepAlive(stateEpoch)
	runtime.KeepAlive(selectionEpoch)
	if err != nil {
		return SSRCorrectedState{}, err
	}
	result := SSRCorrectedState{
		HasState: bool(out.has_state), ClockS: float64(out.clock_s), HasGroupDelay: bool(out.has_group_delay), GroupDelayS: float64(out.group_delay_s), Degraded: bool(out.degraded), HasSizeEvent: bool(out.has_size_event),
		StrictRefusal: bool(out.strict_refusal), Size: SSRCorrectionSize{OrbitM: float64(out.size.orbit_m), ClockM: float64(out.size.clock_m)},
		HasOversizedReport: bool(out.has_oversized_report), Source: uint32(out.source),
		ProviderID: uint16(out.provider_id), SolutionID: uint8(out.solution_id),
		OrbitRefEpochJ2000S:     float64(out.orbit_ref_epoch_j2000_s),
		ClockRefEpochJ2000S:     float64(out.clock_ref_epoch_j2000_s),
		FirstAppliedEpochJ2000S: float64(out.first_applied_epoch_j2000_s),
	}
	for axis := 0; axis < 3; axis++ {
		result.PositionECEFM[axis] = float64(out.position_ecef_m[axis])
	}
	return result, nil
}
