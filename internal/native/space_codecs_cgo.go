//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
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

// The native protocol handles are protected by positioningHandle. This keeps
// pointer reads, C calls, explicit Close, and cleanup callbacks serialized.

type BiasEpoch struct {
	Year        int32
	DayOfYear   uint16
	SecondOfDay uint32
}

type BiasRecord struct {
	Kind           uint32
	TargetKind     uint32
	System         uint32
	HasSatelliteID bool
	SatelliteID    string
	Station        string
	SVN            string
	Obs1           string
	HasObs2        bool
	Obs2           string
	HasValidFrom   bool
	ValidFrom      BiasEpoch
	HasValidUntil  bool
	ValidUntil     BiasEpoch
	Value          float64
	HasSigma       bool
	Sigma          float64
	HasSlope       bool
	Slope          float64
	HasSlopeSigma  bool
	SlopeSigma     float64
	IsPhase        bool
}

type CodeDCBOptions struct {
	Obs1              string
	Obs2              string
	Year              int32
	Month             uint8
	TimeScale         uint32
	HasReceiverSystem bool
	ReceiverSystem    uint32
}

type BiasSet struct{ handle *positioningHandle }

func newBiasSet(pointer *C.SidereonBiasSet) (*BiasSet, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	return &BiasSet{handle: newPositioningHandle(unsafe.Pointer(pointer), func(value unsafe.Pointer) {
		C.sidereon_bias_set_free((*C.SidereonBiasSet)(value))
	})}, nil
}

func parseBias(data []byte, lossy bool) (*BiasSet, error) {
	if err := rejectEmbeddedNULBytes(data, "bias data"); err != nil {
		return nil, err
	}
	var pointer *C.SidereonBiasSet
	err := withInput(data, func(value *C.uint8_t, length C.size_t) uint32 {
		if lossy {
			return uint32(C.sidereon_bias_sinex_parse_lossy(value, length, &pointer))
		}
		return uint32(C.sidereon_bias_sinex_parse(value, length, &pointer))
	})
	if err != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_bias_set_free(pointer) })
		}
		return nil, err
	}
	return newBiasSet(pointer)
}

func ParseBiasSINEX(data []byte, lossy bool) (*BiasSet, error) { return parseBias(data, lossy) }

func codeDCBOptions(value *CodeDCBOptions) (*C.SidereonCodeDcbOptions, func(), error) {
	if value == nil {
		return nil, func() {}, nil
	}
	for name, text := range map[string]string{"DCB observation 1": value.Obs1, "DCB observation 2": value.Obs2} {
		if err := validateNativeCString(text, name); err != nil {
			return nil, nil, err
		}
	}
	obs1 := C.CString(value.Obs1)
	obs2 := C.CString(value.Obs2)
	if obs1 == nil || obs2 == nil {
		if obs1 != nil {
			C.free(unsafe.Pointer(obs1))
		}
		if obs2 != nil {
			C.free(unsafe.Pointer(obs2))
		}
		return nil, nil, errors.New("sidereon: unable to allocate DCB observations")
	}
	optionBytes, err := checkedNativeAllocationSize(1, unsafe.Sizeof(C.SidereonCodeDcbOptions{}))
	if err != nil {
		C.free(unsafe.Pointer(obs1))
		C.free(unsafe.Pointer(obs2))
		return nil, nil, err
	}
	allocation := C.malloc(C.size_t(optionBytes))
	if allocation == nil {
		C.free(unsafe.Pointer(obs1))
		C.free(unsafe.Pointer(obs2))
		return nil, nil, errors.New("sidereon: unable to allocate DCB options")
	}
	options := (*C.SidereonCodeDcbOptions)(allocation)
	*options = C.SidereonCodeDcbOptions{
		obs1: obs1, obs2: obs2, year: C.int32_t(value.Year), month: C.uint8_t(value.Month),
		time_scale: C.uint32_t(value.TimeScale), has_receiver_system: C.bool(value.HasReceiverSystem),
		receiver_system: C.uint32_t(value.ReceiverSystem),
	}
	return options, func() {
		C.free(unsafe.Pointer(options))
		C.free(unsafe.Pointer(obs1))
		C.free(unsafe.Pointer(obs2))
	}, nil
}

func parseCodeDCB(data []byte, options *CodeDCBOptions, lossy bool) (*BiasSet, error) {
	if err := rejectEmbeddedNULBytes(data, "DCB data"); err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(len(data), 1); err != nil {
		return nil, err
	}
	var pointer *C.SidereonBiasSet
	var result error
	withCThread(func() {
		coptions, cleanup, err := codeDCBOptions(options)
		if err != nil {
			result = err
			return
		}
		defer cleanup()
		var input unsafe.Pointer
		if len(data) != 0 {
			input = C.CBytes(data)
			if input == nil {
				result = errors.New("sidereon: unable to allocate native input buffer")
				return
			}
			defer C.free(input)
		}
		var status uint32
		if lossy {
			status = uint32(C.sidereon_code_dcb_parse_lossy((*C.uint8_t)(input), C.size_t(len(data)), coptions, &pointer))
		} else {
			status = uint32(C.sidereon_code_dcb_parse((*C.uint8_t)(input), C.size_t(len(data)), coptions, &pointer))
		}
		result = statusErrorLocked(status)
		if result != nil && pointer != nil {
			C.sidereon_bias_set_free(pointer)
			pointer = nil
		}
	})
	runtime.KeepAlive(data)
	if result != nil {
		return nil, result
	}
	return newBiasSet(pointer)
}

func ParseCodeDCB(data []byte, options *CodeDCBOptions, lossy bool) (*BiasSet, error) {
	return parseCodeDCB(data, options, lossy)
}

func (s *BiasSet) Close() error { return s.handle.close() }

func (s *BiasSet) count(which func(*C.SidereonBiasSet, *C.size_t) uint32) (int, error) {
	var count C.size_t
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 { return which((*C.SidereonBiasSet)(pointer), &count) })
	})
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(count))
}

func (s *BiasSet) RecordCount() (int, error) {
	return s.count(func(pointer *C.SidereonBiasSet, count *C.size_t) uint32 {
		return uint32(C.sidereon_bias_set_record_count(pointer, count))
	})
}
func (s *BiasSet) SkippedRecordCount() (int, error) {
	return s.count(func(pointer *C.SidereonBiasSet, count *C.size_t) uint32 {
		return uint32(C.sidereon_bias_set_skipped_record_count(pointer, count))
	})
}
func (s *BiasSet) WarningCount() (int, error) {
	return s.count(func(pointer *C.SidereonBiasSet, count *C.size_t) uint32 {
		return uint32(C.sidereon_bias_set_warning_count(pointer, count))
	})
}

func cChars(pointer *C.char, length int) string {
	if pointer == nil || length <= 0 {
		return ""
	}
	bytes := unsafe.Slice((*byte)(unsafe.Pointer(pointer)), length)
	for index, value := range bytes {
		if value == 0 {
			bytes = bytes[:index]
			break
		}
	}
	return string(append([]byte(nil), bytes...))
}

func biasEpochFromC(value C.SidereonBiasEpoch) BiasEpoch {
	return BiasEpoch{Year: int32(value.year), DayOfYear: uint16(value.day_of_year), SecondOfDay: uint32(value.second_of_day)}
}

func biasRecordFromC(value C.SidereonBiasRecord) BiasRecord {
	return BiasRecord{
		Kind: uint32(value.kind), TargetKind: uint32(value.target_kind), System: uint32(value.system),
		HasSatelliteID: bool(value.has_sat_id), SatelliteID: tokenFromC(value.sat_id),
		Station: cChars((*C.char)(unsafe.Pointer(&value.station[0])), len(value.station)),
		SVN:     cChars((*C.char)(unsafe.Pointer(&value.svn[0])), len(value.svn)),
		Obs1:    cChars((*C.char)(unsafe.Pointer(&value.obs1[0])), len(value.obs1)), HasObs2: bool(value.has_obs2),
		Obs2: cChars((*C.char)(unsafe.Pointer(&value.obs2[0])), len(value.obs2)), HasValidFrom: bool(value.has_valid_from),
		ValidFrom: biasEpochFromC(value.valid_from), HasValidUntil: bool(value.has_valid_until), ValidUntil: biasEpochFromC(value.valid_until),
		Value: float64(value.value), HasSigma: bool(value.has_sigma), Sigma: float64(value.sigma), HasSlope: bool(value.has_slope),
		Slope: float64(value.slope), HasSlopeSigma: bool(value.has_slope_sigma), SlopeSigma: float64(value.slope_sigma), IsPhase: bool(value.is_phase),
	}
}

func (s *BiasSet) Record(index int) (BiasRecord, error) {
	if index < 0 {
		return BiasRecord{}, errNegativeIndex
	}
	nativeIndex, err := checkedNativeSize(index)
	if err != nil {
		return BiasRecord{}, err
	}
	var value C.SidereonBiasRecord
	err = s.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_bias_set_record((*C.SidereonBiasSet)(pointer), nativeIndex, &value)
		})
	})
	return biasRecordFromC(value), err
}

func biasLookup(s *BiasSet, sat, obs string, epoch BiasEpoch, phase bool) (float64, bool, error) {
	if err := validateNativeCString(sat, "satellite identifier"); err != nil {
		return 0, false, err
	}
	if err := validateNativeCString(obs, "bias observation"); err != nil {
		return 0, false, err
	}
	var value C.double
	var present C.bool
	var result error
	withCThread(func() {
		csat, cobs := C.CString(sat), C.CString(obs)
		if csat == nil || cobs == nil {
			if csat != nil {
				C.free(unsafe.Pointer(csat))
			}
			if cobs != nil {
				C.free(unsafe.Pointer(cobs))
			}
			result = errors.New("sidereon: unable to allocate bias lookup strings")
			return
		}
		defer C.free(unsafe.Pointer(csat))
		defer C.free(unsafe.Pointer(cobs))
		result = s.handle.with(func(pointer unsafe.Pointer) error {
			cepoch := C.SidereonBiasEpoch{year: C.int32_t(epoch.Year), day_of_year: C.uint16_t(epoch.DayOfYear), second_of_day: C.uint32_t(epoch.SecondOfDay)}
			if phase {
				return statusErrorLocked(uint32(C.sidereon_bias_set_phase_osb_cycles((*C.SidereonBiasSet)(pointer), csat, cobs, cepoch, &present, &value)))
			}
			return statusErrorLocked(uint32(C.sidereon_bias_set_code_osb_seconds((*C.SidereonBiasSet)(pointer), csat, cobs, cepoch, &present, &value)))
		})
	})
	return float64(value), bool(present), result
}

func (s *BiasSet) CodeOSBSeconds(sat, obs string, epoch BiasEpoch) (float64, bool, error) {
	return biasLookup(s, sat, obs, epoch, false)
}
func (s *BiasSet) PhaseOSBCycles(sat, obs string, epoch BiasEpoch) (float64, bool, error) {
	return biasLookup(s, sat, obs, epoch, true)
}

func (s *BiasSet) CodeDSBSeconds(sat, obs1, obs2 string, epoch BiasEpoch) (float64, bool, error) {
	for name, value := range map[string]string{"satellite identifier": sat, "bias observation 1": obs1, "bias observation 2": obs2} {
		if err := validateNativeCString(value, name); err != nil {
			return 0, false, err
		}
	}
	var value C.double
	var present C.bool
	var result error
	withCThread(func() {
		csat, cobs1, cobs2 := C.CString(sat), C.CString(obs1), C.CString(obs2)
		if csat == nil || cobs1 == nil || cobs2 == nil {
			if csat != nil {
				C.free(unsafe.Pointer(csat))
			}
			if cobs1 != nil {
				C.free(unsafe.Pointer(cobs1))
			}
			if cobs2 != nil {
				C.free(unsafe.Pointer(cobs2))
			}
			result = errors.New("sidereon: unable to allocate bias lookup strings")
			return
		}
		defer C.free(unsafe.Pointer(csat))
		defer C.free(unsafe.Pointer(cobs1))
		defer C.free(unsafe.Pointer(cobs2))
		result = s.handle.with(func(pointer unsafe.Pointer) error {
			cepoch := C.SidereonBiasEpoch{year: C.int32_t(epoch.Year), day_of_year: C.uint16_t(epoch.DayOfYear), second_of_day: C.uint32_t(epoch.SecondOfDay)}
			return statusErrorLocked(uint32(C.sidereon_bias_set_code_dsb_seconds((*C.SidereonBiasSet)(pointer), csat, cobs1, cobs2, cepoch, &present, &value)))
		})
	})
	return float64(value), bool(present), result
}

func (s *BiasSet) Mode() (mode, scale uint32, err error) {
	var cmode C.enum_SidereonBiasMode
	var cscale C.uint32_t
	err = s.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 { return C.sidereon_bias_set_mode((*C.SidereonBiasSet)(pointer), &cmode, &cscale) })
	})
	return uint32(cmode), uint32(cscale), err
}

type AllanSample struct {
	Present bool
	Value   float64
}
type AllanSeries struct {
	Kind    uint32
	Samples []AllanSample
}
type AllanEstimatorSet struct{ ADEV, OverlappingADEV, MDEV, HDEV, TDEV bool }
type AllanOptions struct {
	Estimators       AllanEstimatorSet
	TauGrid          uint32
	GapPolicy        uint32
	AveragingFactors []int
}
type AllanPoint struct {
	TauS, Deviation float64
	N               int
}
type AllanDeviationCurves struct{ handle *positioningHandle }
type PowerLawNoiseOptions struct {
	MinPointsPerOctave                                                  int
	SlopeTolerance, ScatterTolerance, BasicTauS, MeasurementBandwidthHz float64
}
type PowerLawOctave struct {
	TauStartS, TauEndS             float64
	PointCount                     int
	HasADEVSlope                   bool
	ADEVSlope                      float64
	HasMDEVSlope                   bool
	MDEVSlope                      float64
	HasSlopeScatter                bool
	SlopeScatter                   float64
	DominanceKind, NoiseType, Flag uint32
}
type PowerLawNoiseRegion struct {
	NoiseType               uint32
	TauStartS, TauEndS      float64
	OctaveCount, PointCount int
	MeanSlope, Coefficient  float64
}
type PowerLawNoiseFit struct{ handle *positioningHandle }

func newAllanCurves(pointer *C.SidereonAllanDeviationCurves) (*AllanDeviationCurves, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	return &AllanDeviationCurves{handle: newPositioningHandle(unsafe.Pointer(pointer), func(value unsafe.Pointer) {
		C.sidereon_clock_allan_deviation_curves_free((*C.SidereonAllanDeviationCurves)(value))
	})}, nil
}
func newPowerLawFit(pointer *C.SidereonPowerLawNoiseFit) (*PowerLawNoiseFit, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	return &PowerLawNoiseFit{handle: newPositioningHandle(unsafe.Pointer(pointer), func(value unsafe.Pointer) {
		C.sidereon_clock_power_law_noise_fit_free((*C.SidereonPowerLawNoiseFit)(value))
	})}, nil
}

func rejectEmbeddedNULBytes(value []byte, name string) error {
	for _, byteValue := range value {
		if byteValue == 0 {
			return fmt.Errorf("sidereon: %s contains an embedded NUL byte", name)
		}
	}
	return nil
}

// validateNativeCString checks the complete allocation needed by C.CString.
// The returned C strings are always freed by the caller after the synchronous
// native route returns.
func validateNativeCString(value, name string) error {
	if err := rejectEmbeddedNUL(value, name); err != nil {
		return err
	}
	if len(value) == int(^uint(0)>>1) {
		return fmt.Errorf("sidereon: %s is too large", name)
	}
	_, err := checkedNativeAllocationSize(len(value)+1, 1)
	return err
}

func nativeSamples(series AllanSeries) (unsafe.Pointer, C.size_t, func(), error) {
	bytes, err := checkedNativeAllocationSize(len(series.Samples), unsafe.Sizeof(C.SidereonAllanSample{}))
	if err != nil {
		return nil, 0, func() {}, err
	}
	if bytes == 0 {
		return nil, 0, func() {}, nil
	}
	pointer := C.malloc(C.size_t(bytes))
	if pointer == nil {
		return nil, 0, func() {}, errors.New("sidereon: unable to allocate Allan samples")
	}
	values := unsafe.Slice((*C.SidereonAllanSample)(pointer), len(series.Samples))
	for index, sample := range series.Samples {
		values[index] = C.SidereonAllanSample{has_value: C.bool(sample.Present), value: C.double(sample.Value)}
	}
	return pointer, C.size_t(len(values)), func() { C.free(pointer) }, nil
}

func nativeFactors(factors []int) (unsafe.Pointer, C.size_t, func(), error) {
	if len(factors) == 0 {
		return nil, 0, func() {}, nil
	}
	for _, factor := range factors {
		if factor < 0 {
			return nil, 0, func() {}, errNegativeIndex
		}
		if factor == 0 {
			return nil, 0, func() {}, invalidArgument("Allan averaging factors must be positive")
		}
		if _, err := checkedNativeSize(factor); err != nil {
			return nil, 0, func() {}, err
		}
	}
	bytes, err := checkedNativeAllocationSize(len(factors), unsafe.Sizeof(C.size_t(0)))
	if err != nil {
		return nil, 0, func() {}, err
	}
	pointer := C.malloc(C.size_t(bytes))
	if pointer == nil {
		return nil, 0, func() {}, errors.New("sidereon: unable to allocate Allan averaging factors")
	}
	values := unsafe.Slice((*C.size_t)(pointer), len(factors))
	for index, factor := range factors {
		values[index] = C.size_t(factor)
	}
	return pointer, C.size_t(len(values)), func() { C.free(pointer) }, nil
}

func allanOptionsToC(value AllanOptions) (*C.SidereonAllanOptions, func(), error) {
	factors, count, cleanup, err := nativeFactors(value.AveragingFactors)
	if err != nil {
		return nil, func() {}, err
	}
	optionBytes, err := checkedNativeAllocationSize(1, unsafe.Sizeof(C.SidereonAllanOptions{}))
	if err != nil {
		cleanup()
		return nil, func() {}, err
	}
	allocation := C.malloc(C.size_t(optionBytes))
	if allocation == nil {
		cleanup()
		return nil, func() {}, errors.New("sidereon: unable to allocate Allan options")
	}
	options := (*C.SidereonAllanOptions)(allocation)
	*options = C.SidereonAllanOptions{
		estimators: C.SidereonAllanEstimatorSet{adev: C.bool(value.Estimators.ADEV), overlapping_adev: C.bool(value.Estimators.OverlappingADEV), mdev: C.bool(value.Estimators.MDEV), hdev: C.bool(value.Estimators.HDEV), tdev: C.bool(value.Estimators.TDEV)},
		tau_grid:   C.uint32_t(value.TauGrid), gap_policy: C.uint32_t(value.GapPolicy), averaging_factors: (*C.size_t)(factors), averaging_factor_count: count,
	}
	return options, func() {
		C.free(allocation)
		cleanup()
	}, nil
}

func validateAllanSeriesKind(kind uint32) error {
	switch kind {
	case AllanSeriesPhaseSecondsValue, AllanSeriesFractionalFrequencyValue, AllanSeriesPhaseSecondsWithGapsValue, AllanSeriesFractionalFrequencyWithGapsValue:
		return nil
	default:
		return invalidArgument("invalid Allan series kind")
	}
}

func validateAllanOptions(value AllanOptions) error {
	switch value.TauGrid {
	case AllanTauGridOctaveValue, AllanTauGridAllValue, AllanTauGridExplicitValue:
	default:
		return invalidArgument("invalid Allan tau grid")
	}
	switch value.GapPolicy {
	case GapPolicyRejectValue, GapPolicyOmitTermsValue:
	default:
		return invalidArgument("invalid Allan gap policy")
	}
	return nil
}

func validateAllanEstimator(estimator uint32) error {
	switch estimator {
	case AllanEstimatorADEVValue, AllanEstimatorOverlappingADEVValue, AllanEstimatorMDEVValue, AllanEstimatorHDEVValue, AllanEstimatorTDEVValue:
		return nil
	default:
		return invalidArgument("invalid Allan estimator")
	}
}

func AllanOptionsDefault() (AllanOptions, error) {
	var value C.SidereonAllanOptions
	err := callStatus(func() uint32 { return C.sidereon_clock_allan_options_init(&value) })
	return AllanOptions{Estimators: AllanEstimatorSet{ADEV: bool(value.estimators.adev), OverlappingADEV: bool(value.estimators.overlapping_adev), MDEV: bool(value.estimators.mdev), HDEV: bool(value.estimators.hdev), TDEV: bool(value.estimators.tdev)}, TauGrid: uint32(value.tau_grid), GapPolicy: uint32(value.gap_policy)}, err
}

func ComputeAllanDeviations(input AllanInputNative) (*AllanDeviationCurves, error) {
	return computeAllan(input)
}

type AllanInputNative struct {
	Series  AllanSeries
	Tau0S   float64
	Options AllanOptions
}

func computeAllan(input AllanInputNative) (*AllanDeviationCurves, error) {
	if err := validateAllanSeriesKind(input.Series.Kind); err != nil {
		return nil, err
	}
	if err := validateAllanOptions(input.Options); err != nil {
		return nil, err
	}
	samples, count, sampleCleanup, err := nativeSamples(input.Series)
	if err != nil {
		return nil, err
	}
	defer sampleCleanup()
	options, optionCleanup, err := allanOptionsToC(input.Options)
	if err != nil {
		return nil, err
	}
	defer optionCleanup()
	var pointer *C.SidereonAllanDeviationCurves
	err = callStatus(func() uint32 {
		return C.sidereon_clock_compute_allan_deviations((*C.SidereonAllanSample)(samples), count, C.uint32_t(input.Series.Kind), C.double(input.Tau0S), options, &pointer)
	})
	if err != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_clock_allan_deviation_curves_free(pointer) })
		}
		return nil, err
	}
	return newAllanCurves(pointer)
}

func (c *AllanDeviationCurves) Close() error { return c.handle.close() }
func (c *AllanDeviationCurves) Present(estimator uint32) (bool, error) {
	if err := validateAllanEstimator(estimator); err != nil {
		return false, err
	}
	var value C.bool
	err := c.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_clock_allan_curve_present((*C.SidereonAllanDeviationCurves)(pointer), C.uint32_t(estimator), &value)
		})
	})
	return bool(value), err
}
func (c *AllanDeviationCurves) Curve(estimator uint32) ([]AllanPoint, bool, error) {
	if err := validateAllanEstimator(estimator); err != nil {
		return nil, false, err
	}
	var values []C.SidereonAllanPoint
	var present bool
	err := c.handle.with(func(pointer unsafe.Pointer) error {
		var cpresent C.bool
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_clock_allan_curve_present((*C.SidereonAllanDeviationCurves)(pointer), C.uint32_t(estimator), &cpresent)
		}); err != nil {
			return err
		}
		present = bool(cpresent)
		if !present {
			return nil
		}
		if err := callStatus(func() uint32 {
			return C.sidereon_clock_allan_curve((*C.SidereonAllanDeviationCurves)(pointer), C.uint32_t(estimator), nil, 0, &written, &required)
		}); err != nil {
			return err
		}
		n, err := validateNativeQuery("Allan curve", uint64(written), uint64(required))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonAllanPoint{})); err != nil {
			return err
		}
		values = make([]C.SidereonAllanPoint, n)
		var output *C.SidereonAllanPoint
		if n != 0 {
			output = &values[0]
		}
		if err := callStatus(func() uint32 {
			return C.sidereon_clock_allan_curve((*C.SidereonAllanDeviationCurves)(pointer), C.uint32_t(estimator), output, C.size_t(n), &written, &required)
		}); err != nil {
			return err
		}
		_, err = validateTwoPassCounts("Allan curve", len(values), n, uint64(written), uint64(required))
		return err
	})
	if err != nil {
		return nil, false, err
	}
	if !present {
		return nil, false, nil
	}
	n := len(values)
	out := make([]AllanPoint, n)
	for index := range out {
		terms, countErr := checkedNativeCount(uint64(values[index].n))
		if countErr != nil {
			return nil, false, countErr
		}
		out[index] = AllanPoint{TauS: float64(values[index].tau_s), Deviation: float64(values[index].deviation), N: terms}
	}
	return out, true, nil
}

func explicitAllan(series AllanSeries, tau0 float64, factors []int, estimator uint32) ([]AllanPoint, error) {
	if err := validateAllanSeriesKind(series.Kind); err != nil {
		return nil, err
	}
	if err := validateAllanEstimator(estimator); err != nil {
		return nil, err
	}
	samples, count, sampleCleanup, err := nativeSamples(series)
	if err != nil {
		return nil, err
	}
	defer sampleCleanup()
	factorPointer, factorCount, factorCleanup, err := nativeFactors(factors)
	if err != nil {
		return nil, err
	}
	defer factorCleanup()
	var written, required C.size_t
	call := func(out *C.SidereonAllanPoint, length C.size_t) uint32 {
		switch estimator {
		case AllanEstimatorADEVValue:
			return uint32(C.sidereon_clock_allan_deviation((*C.SidereonAllanSample)(samples), count, C.uint32_t(series.Kind), C.double(tau0), (*C.size_t)(factorPointer), factorCount, out, length, &written, &required))
		case AllanEstimatorOverlappingADEVValue:
			return uint32(C.sidereon_clock_overlapping_adev((*C.SidereonAllanSample)(samples), count, C.uint32_t(series.Kind), C.double(tau0), (*C.size_t)(factorPointer), factorCount, out, length, &written, &required))
		case AllanEstimatorMDEVValue:
			return uint32(C.sidereon_clock_modified_adev((*C.SidereonAllanSample)(samples), count, C.uint32_t(series.Kind), C.double(tau0), (*C.size_t)(factorPointer), factorCount, out, length, &written, &required))
		case AllanEstimatorHDEVValue:
			return uint32(C.sidereon_clock_hadamard_deviation((*C.SidereonAllanSample)(samples), count, C.uint32_t(series.Kind), C.double(tau0), (*C.size_t)(factorPointer), factorCount, out, length, &written, &required))
		default:
			return uint32(C.sidereon_clock_time_deviation((*C.SidereonAllanSample)(samples), count, C.uint32_t(series.Kind), C.double(tau0), (*C.size_t)(factorPointer), factorCount, out, length, &written, &required))
		}
	}
	err = callStatus(func() uint32 { return call(nil, 0) })
	if err != nil {
		return nil, err
	}
	n, err := validateNativeQuery("Allan deviation", uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonAllanPoint{})); err != nil {
		return nil, err
	}
	values := make([]C.SidereonAllanPoint, n)
	var output *C.SidereonAllanPoint
	if n != 0 {
		output = &values[0]
	}
	err = callStatus(func() uint32 { return call(output, C.size_t(n)) })
	if err != nil {
		return nil, err
	}
	n, err = validateTwoPassCounts("Allan deviation", len(values), n, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	out := make([]AllanPoint, n)
	for index := range out {
		terms, countErr := checkedNativeCount(uint64(values[index].n))
		if countErr != nil {
			return nil, countErr
		}
		out[index] = AllanPoint{TauS: float64(values[index].tau_s), Deviation: float64(values[index].deviation), N: terms}
	}
	return out, nil
}
func AllanDeviation(series AllanSeries, tau0 float64, factors []int) ([]AllanPoint, error) {
	return explicitAllan(series, tau0, factors, AllanEstimatorADEVValue)
}
func OverlappingADEV(series AllanSeries, tau0 float64, factors []int) ([]AllanPoint, error) {
	return explicitAllan(series, tau0, factors, AllanEstimatorOverlappingADEVValue)
}
func ModifiedADEV(series AllanSeries, tau0 float64, factors []int) ([]AllanPoint, error) {
	return explicitAllan(series, tau0, factors, AllanEstimatorMDEVValue)
}
func HadamardDeviation(series AllanSeries, tau0 float64, factors []int) ([]AllanPoint, error) {
	return explicitAllan(series, tau0, factors, AllanEstimatorHDEVValue)
}
func TimeDeviation(series AllanSeries, tau0 float64, factors []int) ([]AllanPoint, error) {
	return explicitAllan(series, tau0, factors, AllanEstimatorTDEVValue)
}

func PowerLawNoiseOptionsDefault(basicTauS, bandwidthHz float64) (PowerLawNoiseOptions, error) {
	var value C.SidereonPowerLawNoiseOptions
	err := callStatus(func() uint32 {
		return C.sidereon_clock_power_law_noise_options_init(C.double(basicTauS), C.double(bandwidthHz), &value)
	})
	if err != nil {
		return PowerLawNoiseOptions{}, err
	}
	minimum, err := checkedNativeCount(uint64(value.min_points_per_octave))
	if err != nil {
		return PowerLawNoiseOptions{}, err
	}
	return PowerLawNoiseOptions{MinPointsPerOctave: minimum, SlopeTolerance: float64(value.slope_tolerance), ScatterTolerance: float64(value.scatter_tolerance), BasicTauS: float64(value.basic_tau_s), MeasurementBandwidthHz: float64(value.measurement_bandwidth_hz)}, nil
}
func PowerLawNoiseSlopes(noiseType uint32) (adev, mdev float64, exponent int, err error) {
	switch noiseType {
	case PowerLawRandomWalkFMValue, PowerLawFlickerFMValue, PowerLawWhiteFMValue, PowerLawFlickerPMValue, PowerLawWhitePMValue:
	default:
		return 0, 0, 0, invalidArgument("invalid power-law noise type")
	}
	var a, m C.double
	var e C.int32_t
	err = callStatus(func() uint32 { return C.sidereon_clock_power_law_noise_slopes(C.uint32_t(noiseType), &a, &m, &e) })
	return float64(a), float64(m), int(e), err
}

func nativePoints(points []AllanPoint) (unsafe.Pointer, C.size_t, func(), error) {
	for _, point := range points {
		if point.N < 0 {
			return nil, 0, func() {}, errNegativeIndex
		}
	}
	bytes, err := checkedNativeAllocationSize(len(points), unsafe.Sizeof(C.SidereonAllanPoint{}))
	if err != nil {
		return nil, 0, func() {}, err
	}
	if bytes == 0 {
		return nil, 0, func() {}, nil
	}
	pointer := C.malloc(C.size_t(bytes))
	if pointer == nil {
		return nil, 0, func() {}, errors.New("sidereon: unable to allocate Allan points")
	}
	values := unsafe.Slice((*C.SidereonAllanPoint)(pointer), len(points))
	for index, point := range points {
		terms, err := checkedNativeSize(point.N)
		if err != nil {
			C.free(pointer)
			return nil, 0, func() {}, err
		}
		values[index] = C.SidereonAllanPoint{tau_s: C.double(point.TauS), deviation: C.double(point.Deviation), n: terms}
	}
	return pointer, C.size_t(len(points)), func() { C.free(pointer) }, nil
}
func FitPowerLawNoise(adev, mdev []AllanPoint, options *PowerLawNoiseOptions) (*PowerLawNoiseFit, error) {
	ap, ac, aCleanup, err := nativePoints(adev)
	if err != nil {
		return nil, err
	}
	defer aCleanup()
	mp, mc, mCleanup, err := nativePoints(mdev)
	if err != nil {
		return nil, err
	}
	defer mCleanup()
	var optionPointer *C.SidereonPowerLawNoiseOptions
	var optionCleanup func()
	if options != nil {
		minimum, sizeErr := checkedNativeSize(options.MinPointsPerOctave)
		if sizeErr != nil {
			return nil, sizeErr
		}
		optionBytes, sizeErr := checkedNativeAllocationSize(1, unsafe.Sizeof(C.SidereonPowerLawNoiseOptions{}))
		if sizeErr != nil {
			return nil, sizeErr
		}
		allocation := C.malloc(C.size_t(optionBytes))
		if allocation == nil {
			return nil, errors.New("sidereon: unable to allocate power-law options")
		}
		optionPointer = (*C.SidereonPowerLawNoiseOptions)(allocation)
		*optionPointer = C.SidereonPowerLawNoiseOptions{min_points_per_octave: minimum, slope_tolerance: C.double(options.SlopeTolerance), scatter_tolerance: C.double(options.ScatterTolerance), basic_tau_s: C.double(options.BasicTauS), measurement_bandwidth_hz: C.double(options.MeasurementBandwidthHz)}
		optionCleanup = func() { C.free(allocation) }
	}
	if optionCleanup != nil {
		defer optionCleanup()
	}
	var pointer *C.SidereonPowerLawNoiseFit
	err = callStatus(func() uint32 {
		return C.sidereon_clock_fit_power_law_noise((*C.SidereonAllanPoint)(ap), ac, (*C.SidereonAllanPoint)(mp), mc, optionPointer, &pointer)
	})
	if err != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_clock_power_law_noise_fit_free(pointer) })
		}
		return nil, err
	}
	return newPowerLawFit(pointer)
}
func (f *PowerLawNoiseFit) Close() error { return f.handle.close() }
func (f *PowerLawNoiseFit) Coefficients() ([5]float64, error) {
	var c [5]C.double
	err := f.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_clock_power_law_noise_fit_coefficients((*C.SidereonPowerLawNoiseFit)(pointer), &c[0])
		})
	})
	var out [5]float64
	for index := range out {
		out[index] = float64(c[index])
	}
	return out, err
}
func (f *PowerLawNoiseFit) Octaves() ([]PowerLawOctave, error) {
	var values []C.SidereonPowerLawOctave
	err := f.handle.with(func(pointer unsafe.Pointer) error {
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_clock_power_law_noise_fit_octaves((*C.SidereonPowerLawNoiseFit)(pointer), nil, 0, &written, &required)
		}); err != nil {
			return err
		}
		n, err := validateNativeQuery("power-law octaves", uint64(written), uint64(required))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonPowerLawOctave{})); err != nil {
			return err
		}
		values = make([]C.SidereonPowerLawOctave, n)
		var output *C.SidereonPowerLawOctave
		if n != 0 {
			output = &values[0]
		}
		if err := callStatus(func() uint32 {
			return C.sidereon_clock_power_law_noise_fit_octaves((*C.SidereonPowerLawNoiseFit)(pointer), output, C.size_t(n), &written, &required)
		}); err != nil {
			return err
		}
		_, err = validateTwoPassCounts("power-law octaves", len(values), n, uint64(written), uint64(required))
		return err
	})
	if err != nil {
		return nil, err
	}
	out := make([]PowerLawOctave, len(values))
	for index := range out {
		value := values[index]
		pointCount, countErr := checkedNativeCount(uint64(value.point_count))
		if countErr != nil {
			return nil, countErr
		}
		out[index] = PowerLawOctave{TauStartS: float64(value.tau_start_s), TauEndS: float64(value.tau_end_s), PointCount: pointCount, HasADEVSlope: bool(value.has_adev_slope), ADEVSlope: float64(value.adev_slope), HasMDEVSlope: bool(value.has_mdev_slope), MDEVSlope: float64(value.mdev_slope), HasSlopeScatter: bool(value.has_slope_scatter), SlopeScatter: float64(value.slope_scatter), DominanceKind: uint32(value.dominance_kind), NoiseType: uint32(value.noise_type), Flag: uint32(value.flag)}
	}
	return out, nil
}
func (f *PowerLawNoiseFit) Regions() ([]PowerLawNoiseRegion, error) {
	var values []C.SidereonPowerLawNoiseRegion
	err := f.handle.with(func(pointer unsafe.Pointer) error {
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_clock_power_law_noise_fit_regions((*C.SidereonPowerLawNoiseFit)(pointer), nil, 0, &written, &required)
		}); err != nil {
			return err
		}
		n, err := validateNativeQuery("power-law regions", uint64(written), uint64(required))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonPowerLawNoiseRegion{})); err != nil {
			return err
		}
		values = make([]C.SidereonPowerLawNoiseRegion, n)
		var output *C.SidereonPowerLawNoiseRegion
		if n != 0 {
			output = &values[0]
		}
		if err := callStatus(func() uint32 {
			return C.sidereon_clock_power_law_noise_fit_regions((*C.SidereonPowerLawNoiseFit)(pointer), output, C.size_t(n), &written, &required)
		}); err != nil {
			return err
		}
		_, err = validateTwoPassCounts("power-law regions", len(values), n, uint64(written), uint64(required))
		return err
	})
	if err != nil {
		return nil, err
	}
	out := make([]PowerLawNoiseRegion, len(values))
	for index := range out {
		value := values[index]
		octaveCount, countErr := checkedNativeCount(uint64(value.octave_count))
		if countErr != nil {
			return nil, countErr
		}
		pointCount, countErr := checkedNativeCount(uint64(value.point_count))
		if countErr != nil {
			return nil, countErr
		}
		out[index] = PowerLawNoiseRegion{NoiseType: uint32(value.noise_type), TauStartS: float64(value.tau_start_s), TauEndS: float64(value.tau_end_s), OctaveCount: octaveCount, PointCount: pointCount, MeanSlope: float64(value.mean_slope), Coefficient: float64(value.coefficient)}
	}
	return out, nil
}

type OEM struct{ handle *positioningHandle }
type OPM struct{ handle *positioningHandle }
type TDM struct{ handle *positioningHandle }

func ParseOEM(data []byte, xml bool) (*OEM, error) {
	var pointer *C.SidereonOem
	err := withTextInput(data, func(value *C.uint8_t, length C.size_t) uint32 {
		if xml {
			return uint32(C.sidereon_oem_parse_xml(value, length, &pointer))
		}
		return uint32(C.sidereon_oem_parse_kvn(value, length, &pointer))
	})
	if err != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_oem_free(pointer) })
		}
		return nil, err
	}
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	return &OEM{handle: newPositioningHandle(unsafe.Pointer(pointer), func(value unsafe.Pointer) { C.sidereon_oem_free((*C.SidereonOem)(value)) })}, nil
}
func ParseOPM(data []byte, xml bool) (*OPM, error) {
	var pointer *C.SidereonOpm
	err := withTextInput(data, func(value *C.uint8_t, length C.size_t) uint32 {
		if xml {
			return uint32(C.sidereon_opm_parse_xml(value, length, &pointer))
		}
		return uint32(C.sidereon_opm_parse_kvn(value, length, &pointer))
	})
	if err != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_opm_free(pointer) })
		}
		return nil, err
	}
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	return &OPM{handle: newPositioningHandle(unsafe.Pointer(pointer), func(value unsafe.Pointer) { C.sidereon_opm_free((*C.SidereonOpm)(value)) })}, nil
}
func withTextInput(data []byte, fn func(*C.uint8_t, C.size_t) uint32) error {
	if err := rejectEmbeddedNULBytes(data, "space message"); err != nil {
		return err
	}
	return withInput(data, fn)
}
func (o *OEM) Close() error { return o.handle.close() }
func (o *OPM) Close() error { return o.handle.close() }
func (o *OEM) SegmentCount() (int, error) {
	var n C.size_t
	err := o.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 { return C.sidereon_oem_segment_count((*C.SidereonOem)(p), &n) })
	})
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(n))
}

func copyHandleText(handle *positioningHandle, call func(unsafe.Pointer, *C.uint8_t, C.size_t, *C.size_t, *C.size_t) uint32) ([]byte, error) {
	var buffer []C.uint8_t
	err := handle.with(func(pointer unsafe.Pointer) error {
		var written, required C.size_t
		if err := callStatus(func() uint32 { return call(pointer, nil, 0, &written, &required) }); err != nil {
			return err
		}
		n, err := validateNativeQuery("serialized text", uint64(written), uint64(required))
		if err != nil {
			return err
		}
		buffer = make([]C.uint8_t, n)
		var output *C.uint8_t
		if n != 0 {
			output = &buffer[0]
		}
		if err := callStatus(func() uint32 { return call(pointer, output, C.size_t(n), &written, &required) }); err != nil {
			return err
		}
		_, err = validateTwoPassCounts("serialized text", len(buffer), n, uint64(written), uint64(required))
		return err
	})
	if err != nil {
		return nil, err
	}
	n := len(buffer)
	out := make([]byte, n)
	for i := range out {
		out[i] = byte(buffer[i])
	}
	return out, nil
}

func (o *OEM) Text(xml bool) ([]byte, error) {
	return copyHandleText(o.handle, func(pointer unsafe.Pointer, out *C.uint8_t, length C.size_t, written, required *C.size_t) uint32 {
		if xml {
			return uint32(C.sidereon_oem_to_xml((*C.SidereonOem)(pointer), out, length, written, required))
		}
		return uint32(C.sidereon_oem_to_kvn((*C.SidereonOem)(pointer), out, length, written, required))
	})
}

func (o *OMM) Text(format uint32) ([]byte, error) {
	switch format {
	case OMMFormatKVNValue, OMMFormatXMLValue, OMMFormatJSONValue:
	default:
		return nil, invalidArgument("invalid OMM format")
	}
	return copyHandleText(o.handle, func(pointer unsafe.Pointer, out *C.uint8_t, length C.size_t, written, required *C.size_t) uint32 {
		switch format {
		case OMMFormatKVNValue:
			return uint32(C.sidereon_omm_to_kvn((*C.SidereonOmm)(pointer), out, length, written, required))
		case OMMFormatXMLValue:
			return uint32(C.sidereon_omm_to_xml((*C.SidereonOmm)(pointer), out, length, written, required))
		case OMMFormatJSONValue:
			return uint32(C.sidereon_omm_to_json((*C.SidereonOmm)(pointer), out, length, written, required))
		default:
			return uint32(C.SIDEREON_STATUS_INVALID_ARGUMENT)
		}
	})
}

func (o *OPM) Text(xml bool) ([]byte, error) {
	return copyHandleText(o.handle, func(pointer unsafe.Pointer, out *C.uint8_t, length C.size_t, written, required *C.size_t) uint32 {
		if xml {
			return uint32(C.sidereon_opm_to_xml((*C.SidereonOpm)(pointer), out, length, written, required))
		}
		return uint32(C.sidereon_opm_to_kvn((*C.SidereonOpm)(pointer), out, length, written, required))
	})
}

type ConstellationRecord struct {
	System             uint32
	PRN                uint16
	SVNPresent         bool
	SVN                uint16
	NORADID            uint32
	FDMAChannelPresent bool
	FDMAChannel        int8
	Active             bool
	Usable             bool
}

type SkippedOMM struct {
	NORADID           uint32
	ObjectNamePresent bool
	ObjectName        string
}
type OMMCatalog struct{ handle *positioningHandle }

func newOMMCatalog(pointer *C.SidereonOmmCatalog) (*OMMCatalog, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	return &OMMCatalog{handle: newPositioningHandle(unsafe.Pointer(pointer), func(value unsafe.Pointer) { C.sidereon_omm_catalog_free((*C.SidereonOmmCatalog)(value)) })}, nil
}

func BuildOMMCatalogLenient(system uint32, data []byte) (*OMMCatalog, error) {
	if err := rejectEmbeddedNULBytes(data, "OMM catalog JSON"); err != nil {
		return nil, err
	}
	var pointer *C.SidereonOmmCatalog
	err := withInput(data, func(value *C.uint8_t, length C.size_t) uint32 {
		return uint32(C.sidereon_omm_catalog_build_lenient(C.uint32_t(system), value, length, &pointer))
	})
	if err != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_omm_catalog_free(pointer) })
		}
		return nil, err
	}
	return newOMMCatalog(pointer)
}
func (c *OMMCatalog) Close() error { return c.handle.close() }
func (c *OMMCatalog) count(which func(*C.SidereonOmmCatalog, *C.size_t) uint32) (int, error) {
	var n C.size_t
	err := c.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 { return which((*C.SidereonOmmCatalog)(p), &n) })
	})
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(n))
}
func (c *OMMCatalog) RecordCount() (int, error) {
	return c.count(func(pointer *C.SidereonOmmCatalog, count *C.size_t) uint32 {
		return uint32(C.sidereon_omm_catalog_record_count(pointer, count))
	})
}
func (c *OMMCatalog) SkippedCount() (int, error) {
	return c.count(func(pointer *C.SidereonOmmCatalog, count *C.size_t) uint32 {
		return uint32(C.sidereon_omm_catalog_skipped_count(pointer, count))
	})
}
func (c *OMMCatalog) MalformedCount() (int, error) {
	return c.count(func(pointer *C.SidereonOmmCatalog, count *C.size_t) uint32 {
		return uint32(C.sidereon_omm_catalog_malformed_count(pointer, count))
	})
}
func constellationRecordFromC(v C.SidereonConstellationRecord) ConstellationRecord {
	return ConstellationRecord{System: uint32(v.system), PRN: uint16(v.prn), SVNPresent: bool(v.svn_present), SVN: uint16(v.svn), NORADID: uint32(v.norad_id), FDMAChannelPresent: bool(v.fdma_channel_present), FDMAChannel: int8(v.fdma_channel), Active: bool(v.active), Usable: bool(v.usable)}
}
func (c *OMMCatalog) Record(index int) (ConstellationRecord, error) {
	if index < 0 {
		return ConstellationRecord{}, errNegativeIndex
	}
	nativeIndex, err := checkedNativeSize(index)
	if err != nil {
		return ConstellationRecord{}, err
	}
	var v C.SidereonConstellationRecord
	err = c.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 { return C.sidereon_omm_catalog_record((*C.SidereonOmmCatalog)(p), nativeIndex, &v) })
	})
	return constellationRecordFromC(v), err
}
func cSkippedObjectName(catalog *OMMCatalog, index int) (string, error) {
	if index < 0 {
		return "", errNegativeIndex
	}
	nativeIndex, err := checkedNativeSize(index)
	if err != nil {
		return "", err
	}
	var buffer []C.uint8_t
	err = catalog.handle.with(func(p unsafe.Pointer) error {
		var w, r C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_omm_catalog_skipped_object_name((*C.SidereonOmmCatalog)(p), nativeIndex, nil, 0, &w, &r)
		}); err != nil {
			return err
		}
		n, err := validateNativeQuery("OMM skipped object name", uint64(w), uint64(r))
		if err != nil {
			return err
		}
		buffer = make([]C.uint8_t, n)
		var out *C.uint8_t
		if n != 0 {
			out = &buffer[0]
		}
		if err := callStatus(func() uint32 {
			return C.sidereon_omm_catalog_skipped_object_name((*C.SidereonOmmCatalog)(p), nativeIndex, out, C.size_t(n), &w, &r)
		}); err != nil {
			return err
		}
		_, err = validateTwoPassCounts("OMM skipped object name", len(buffer), n, uint64(w), uint64(r))
		return err
	})
	if err != nil {
		return "", err
	}
	out := make([]byte, len(buffer))
	for i := range out {
		out[i] = byte(buffer[i])
	}
	return string(out), nil
}
func (c *OMMCatalog) Skipped(index int) (SkippedOMM, error) {
	if index < 0 {
		return SkippedOMM{}, errNegativeIndex
	}
	nativeIndex, err := checkedNativeSize(index)
	if err != nil {
		return SkippedOMM{}, err
	}
	var v C.SidereonSkippedOmm
	err = c.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 { return C.sidereon_omm_catalog_skipped((*C.SidereonOmmCatalog)(p), nativeIndex, &v) })
	})
	if err != nil {
		return SkippedOMM{}, err
	}
	name, err := cSkippedObjectName(c, index)
	return SkippedOMM{NORADID: uint32(v.norad_id), ObjectNamePresent: bool(v.object_name_present), ObjectName: name}, err
}

type TDMParticipant struct {
	SegmentIndex int
	Index        uint8
	Name         string
}
type TDMPath struct {
	SegmentIndex int
	Key          string
	HasIndex     bool
	Index        uint8
	Participants []uint8
}
type TDMDataRecord struct {
	SegmentIndex              int
	Observable                uint32
	HasObservableParticipant  bool
	ObservableParticipant     uint8
	Unit                      uint32
	Keyword, Epoch, ValueText string
	Value                     float64
}
type TDMStringField struct {
	Present bool
	Value   string
}
type TDMSegmentSummary struct {
	SegmentIndex                             int
	Mode, TimetagRef, TimeSystem             TDMStringField
	RangeUnit                                uint32
	ParticipantCount, PathCount, RecordCount int
}

func newTDM(pointer *C.SidereonTdm) (*TDM, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	return &TDM{handle: newPositioningHandle(unsafe.Pointer(pointer), func(value unsafe.Pointer) { C.sidereon_tdm_free((*C.SidereonTdm)(value)) })}, nil
}
func ParseTDM(data []byte) (*TDM, error) {
	if err := rejectEmbeddedNULBytes(data, "TDM data"); err != nil {
		return nil, err
	}
	var pointer *C.SidereonTdm
	err := withInput(data, func(value *C.uint8_t, length C.size_t) uint32 {
		return uint32(C.sidereon_tdm_parse_kvn(value, length, &pointer))
	})
	if err != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_tdm_free(pointer) })
		}
		return nil, err
	}
	return newTDM(pointer)
}
func (t *TDM) Close() error { return t.handle.close() }
func (t *TDM) Text() ([]byte, error) {
	return copyHandleText(t.handle, func(pointer unsafe.Pointer, out *C.uint8_t, length C.size_t, w, r *C.size_t) uint32 {
		return uint32(C.sidereon_tdm_to_kvn((*C.SidereonTdm)(pointer), out, length, w, r))
	})
}
func (t *TDM) count(which func(*C.SidereonTdm, *C.size_t) uint32) (int, error) {
	var n C.size_t
	err := t.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 { return which((*C.SidereonTdm)(p), &n) })
	})
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(n))
}
func (t *TDM) SegmentCount() (int, error) {
	return t.count(func(pointer *C.SidereonTdm, count *C.size_t) uint32 {
		return uint32(C.sidereon_tdm_segment_count(pointer, count))
	})
}
func (t *TDM) RecordCount() (int, error) {
	return t.count(func(pointer *C.SidereonTdm, count *C.size_t) uint32 {
		return uint32(C.sidereon_tdm_record_count(pointer, count))
	})
}
func (t *TDM) Segments() ([]TDMSegmentSummary, error) {
	var values []C.SidereonTdmSegmentSummary
	err := t.handle.with(func(p unsafe.Pointer) error {
		var w, r C.size_t
		if err := callStatus(func() uint32 { return C.sidereon_tdm_segments((*C.SidereonTdm)(p), nil, 0, &w, &r) }); err != nil {
			return err
		}
		n, err := validateNativeQuery("TDM segments", uint64(w), uint64(r))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonTdmSegmentSummary{})); err != nil {
			return err
		}
		values = make([]C.SidereonTdmSegmentSummary, n)
		var out *C.SidereonTdmSegmentSummary
		if n != 0 {
			out = &values[0]
		}
		if err := callStatus(func() uint32 { return C.sidereon_tdm_segments((*C.SidereonTdm)(p), out, C.size_t(n), &w, &r) }); err != nil {
			return err
		}
		_, err = validateTwoPassCounts("TDM segments", len(values), n, uint64(w), uint64(r))
		return err
	})
	if err != nil {
		return nil, err
	}
	out := make([]TDMSegmentSummary, len(values))
	for i := range out {
		v := values[i]
		segmentIndex, countErr := checkedNativeCount(uint64(v.segment_index))
		if countErr != nil {
			return nil, countErr
		}
		participantCount, countErr := checkedNativeCount(uint64(v.participant_count))
		if countErr != nil {
			return nil, countErr
		}
		pathCount, countErr := checkedNativeCount(uint64(v.path_count))
		if countErr != nil {
			return nil, countErr
		}
		recordCount, countErr := checkedNativeCount(uint64(v.record_count))
		if countErr != nil {
			return nil, countErr
		}
		out[i] = TDMSegmentSummary{SegmentIndex: segmentIndex, Mode: TDMStringField{Present: bool(v.mode.has_value), Value: cChars((*C.char)(unsafe.Pointer(&v.mode.value[0])), len(v.mode.value))}, TimetagRef: TDMStringField{Present: bool(v.timetag_ref.has_value), Value: cChars((*C.char)(unsafe.Pointer(&v.timetag_ref.value[0])), len(v.timetag_ref.value))}, TimeSystem: TDMStringField{Present: bool(v.time_system.has_value), Value: cChars((*C.char)(unsafe.Pointer(&v.time_system.value[0])), len(v.time_system.value))}, RangeUnit: uint32(v.range_unit), ParticipantCount: participantCount, PathCount: pathCount, RecordCount: recordCount}
	}
	return out, nil
}
func (t *TDM) Participants() ([]TDMParticipant, error) {
	var values []C.SidereonTdmParticipant
	err := t.handle.with(func(p unsafe.Pointer) error {
		var w, r C.size_t
		if err := callStatus(func() uint32 { return C.sidereon_tdm_participants((*C.SidereonTdm)(p), nil, 0, &w, &r) }); err != nil {
			return err
		}
		n, err := validateNativeQuery("TDM participants", uint64(w), uint64(r))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonTdmParticipant{})); err != nil {
			return err
		}
		values = make([]C.SidereonTdmParticipant, n)
		var out *C.SidereonTdmParticipant
		if n != 0 {
			out = &values[0]
		}
		if err := callStatus(func() uint32 { return C.sidereon_tdm_participants((*C.SidereonTdm)(p), out, C.size_t(n), &w, &r) }); err != nil {
			return err
		}
		_, err = validateTwoPassCounts("TDM participants", len(values), n, uint64(w), uint64(r))
		return err
	})
	if err != nil {
		return nil, err
	}
	out := make([]TDMParticipant, len(values))
	for i := range out {
		v := values[i]
		segmentIndex, countErr := checkedNativeCount(uint64(v.segment_index))
		if countErr != nil {
			return nil, countErr
		}
		out[i] = TDMParticipant{SegmentIndex: segmentIndex, Index: uint8(v.index), Name: cChars((*C.char)(unsafe.Pointer(&v.name[0])), len(v.name))}
	}
	return out, nil
}
func (t *TDM) Paths() ([]TDMPath, error) {
	var values []C.SidereonTdmPath
	err := t.handle.with(func(p unsafe.Pointer) error {
		var w, r C.size_t
		if err := callStatus(func() uint32 { return C.sidereon_tdm_paths((*C.SidereonTdm)(p), nil, 0, &w, &r) }); err != nil {
			return err
		}
		n, err := validateNativeQuery("TDM paths", uint64(w), uint64(r))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonTdmPath{})); err != nil {
			return err
		}
		values = make([]C.SidereonTdmPath, n)
		var out *C.SidereonTdmPath
		if n != 0 {
			out = &values[0]
		}
		if err := callStatus(func() uint32 { return C.sidereon_tdm_paths((*C.SidereonTdm)(p), out, C.size_t(n), &w, &r) }); err != nil {
			return err
		}
		_, err = validateTwoPassCounts("TDM paths", len(values), n, uint64(w), uint64(r))
		return err
	})
	if err != nil {
		return nil, err
	}
	out := make([]TDMPath, len(values))
	for i := range out {
		v := values[i]
		participantCount, countErr := checkedNativeCount(uint64(v.participant_count))
		if countErr != nil {
			return nil, countErr
		}
		if participantCount > len(v.participants) {
			return nil, errors.New("sidereon: native TDM path participant count exceeds fixed shape")
		}
		segmentIndex, countErr := checkedNativeCount(uint64(v.segment_index))
		if countErr != nil {
			return nil, countErr
		}
		parts := append([]uint8(nil), unsafe.Slice((*uint8)(unsafe.Pointer(&v.participants[0])), participantCount)...)
		out[i] = TDMPath{SegmentIndex: segmentIndex, Key: cChars((*C.char)(unsafe.Pointer(&v.key[0])), len(v.key)), HasIndex: bool(v.has_index), Index: uint8(v.index), Participants: parts}
	}
	return out, nil
}
func (t *TDM) Records() ([]TDMDataRecord, error) {
	var values []C.SidereonTdmDataRecord
	err := t.handle.with(func(p unsafe.Pointer) error {
		var w, r C.size_t
		if err := callStatus(func() uint32 { return C.sidereon_tdm_records((*C.SidereonTdm)(p), nil, 0, &w, &r) }); err != nil {
			return err
		}
		n, err := validateNativeQuery("TDM records", uint64(w), uint64(r))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonTdmDataRecord{})); err != nil {
			return err
		}
		values = make([]C.SidereonTdmDataRecord, n)
		var out *C.SidereonTdmDataRecord
		if n != 0 {
			out = &values[0]
		}
		if err := callStatus(func() uint32 { return C.sidereon_tdm_records((*C.SidereonTdm)(p), out, C.size_t(n), &w, &r) }); err != nil {
			return err
		}
		_, err = validateTwoPassCounts("TDM records", len(values), n, uint64(w), uint64(r))
		return err
	})
	if err != nil {
		return nil, err
	}
	out := make([]TDMDataRecord, len(values))
	for i := range out {
		v := values[i]
		segmentIndex, countErr := checkedNativeCount(uint64(v.segment_index))
		if countErr != nil {
			return nil, countErr
		}
		out[i] = TDMDataRecord{SegmentIndex: segmentIndex, Observable: uint32(v.observable), HasObservableParticipant: bool(v.has_observable_participant), ObservableParticipant: uint8(v.observable_participant), Unit: uint32(v.unit), Keyword: cChars((*C.char)(unsafe.Pointer(&v.keyword[0])), len(v.keyword)), Epoch: cChars((*C.char)(unsafe.Pointer(&v.epoch[0])), len(v.epoch)), ValueText: cChars((*C.char)(unsafe.Pointer(&v.value_text[0])), len(v.value_text)), Value: float64(v.value)}
	}
	return out, nil
}
