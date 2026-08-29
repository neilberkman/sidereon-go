//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

type RtkReceiverAntennaNoAzimuthPCV struct {
	ZenithDeg, ValueM float64
}

type RtkReceiverAntennaAzimuthPCV struct {
	AzimuthDeg, ZenithDeg, ValueM float64
}

type RtkReceiverAntennaCalibration struct {
	PCONEUM      [3]float64
	NoAzimuthPCV []RtkReceiverAntennaNoAzimuthPCV
	AzimuthPCV   []RtkReceiverAntennaAzimuthPCV
}

type RtkReceiverAntennaCorrections struct {
	Base, Rover RtkReceiverAntennaCalibration
}

type RtkFixedConfig struct {
	Epochs              []RtkEpoch
	BaseECEFM           [3]float64
	AmbiguityIDs        []string
	AmbiguitySatellites []RtkAmbiguitySatellite
	Wavelengths         []RtkFloatMapEntry
	Offsets             []RtkFloatMapEntry
	Model               RtkMeasurementModel
	ReceiverAntenna     *RtkReceiverAntennaCorrections
	FloatOptions        RtkFloatOptions
	FixedOptions        RtkFixedOptions
	ResidualOptions     RtkResidualValidationOptions
	FloatOnlySystems    []uint32
	InitialBaselineM    [3]float64
}

type RtkFixedAmbiguity struct {
	ID     string
	Cycles int64
	ValueM float64
}

type RtkFixedMetadata struct {
	Iterations, NObservations, FreeAmbiguityCount, FixedAmbiguityCount int
	ResidualCount, UsedSatCount                                        int
	Converged                                                          bool
	Status, IntegerStatus                                              uint32
	CodeRMSM, PhaseRMSM, WeightedRMSM                                  float64
	HasIntegerRatio                                                    bool
	IntegerRatio                                                       float64
	HasIntegerBestScore                                                bool
	IntegerBestScore                                                   float64
	HasIntegerSecondBestScore                                          bool
	IntegerSecondBestScore                                             float64
	IntegerCandidates                                                  int
	GeometryQuality                                                    GeometryQuality
}

type RtkFixedSolution struct {
	_      noCopy
	handle *surfaceHandle
}

func newRtkFixedSolution(pointer *C.SidereonRtkFixedSolution) (*RtkFixedSolution, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	handle, err := newSurfaceHandle(unsafe.Pointer(pointer), func(value unsafe.Pointer) {
		C.sidereon_rtk_fixed_solution_free((*C.SidereonRtkFixedSolution)(value))
	})
	if err != nil {
		return nil, err
	}
	return &RtkFixedSolution{handle: handle}, nil
}

func copyRtkAmbiguityIDs(values []string, alloc *cRtkAlloc) (**C.char, C.size_t, error) {
	count, err := checkedNativeSize(len(values))
	if err != nil {
		return nil, 0, err
	}
	if len(values) == 0 {
		return nil, count, nil
	}
	bytes, err := checkedNativeAllocationSize(len(values), unsafe.Sizeof(uintptr(0)))
	if err != nil {
		return nil, 0, err
	}
	memory, err := alloc.malloc(bytes, "RTK fixed ambiguity IDs")
	if err != nil {
		return nil, 0, err
	}
	rows := unsafe.Slice((**C.char)(memory), len(values))
	for i, value := range values {
		rows[i], err = alloc.cstring(value, "RTK fixed ambiguity ID")
		if err != nil {
			return nil, 0, err
		}
	}
	return (**C.char)(memory), count, nil
}

func copyRtkAmbiguitySatellites(values []RtkAmbiguitySatellite, alloc *cRtkAlloc) (*C.SidereonRtkAmbiguitySatellite, C.size_t, error) {
	count, err := checkedNativeSize(len(values))
	if err != nil {
		return nil, 0, err
	}
	if len(values) == 0 {
		return nil, count, nil
	}
	bytes, err := checkedNativeAllocationSize(len(values), unsafe.Sizeof(C.SidereonRtkAmbiguitySatellite{}))
	if err != nil {
		return nil, 0, err
	}
	memory, err := alloc.malloc(bytes, "RTK fixed ambiguity-satellite map")
	if err != nil {
		return nil, 0, err
	}
	rows := unsafe.Slice((*C.SidereonRtkAmbiguitySatellite)(memory), len(values))
	for i, value := range values {
		id, e := alloc.cstring(value.ID, "RTK fixed ambiguity ID")
		if e != nil {
			return nil, 0, e
		}
		satellite, e := alloc.cstring(value.SatelliteID, "RTK fixed satellite ID")
		if e != nil {
			return nil, 0, e
		}
		rows[i] = C.SidereonRtkAmbiguitySatellite{id: id, sat_id: satellite}
	}
	return &rows[0], count, nil
}

func copyRtkFloatMapEntries(values []RtkFloatMapEntry, alloc *cRtkAlloc, label string) (*C.SidereonRtkFloatMapEntry, C.size_t, error) {
	count, err := checkedNativeSize(len(values))
	if err != nil {
		return nil, 0, err
	}
	if len(values) == 0 {
		return nil, count, nil
	}
	bytes, err := checkedNativeAllocationSize(len(values), unsafe.Sizeof(C.SidereonRtkFloatMapEntry{}))
	if err != nil {
		return nil, 0, err
	}
	memory, err := alloc.malloc(bytes, label)
	if err != nil {
		return nil, 0, err
	}
	rows := unsafe.Slice((*C.SidereonRtkFloatMapEntry)(memory), len(values))
	for i, value := range values {
		id, e := alloc.cstring(value.ID, "RTK fixed map ID")
		if e != nil {
			return nil, 0, e
		}
		rows[i] = C.SidereonRtkFloatMapEntry{id: id, value: C.double(value.Value)}
	}
	return &rows[0], count, nil
}

func copyRtkReceiverAntennaCalibration(value RtkReceiverAntennaCalibration, alloc *cRtkAlloc) (C.SidereonReceiverAntennaCalibration, error) {
	noAzimuthCount, err := checkedNativeSize(len(value.NoAzimuthPCV))
	if err != nil {
		return C.SidereonReceiverAntennaCalibration{}, err
	}
	azimuthCount, err := checkedNativeSize(len(value.AzimuthPCV))
	if err != nil {
		return C.SidereonReceiverAntennaCalibration{}, err
	}
	var noAzimuth *C.SidereonReceiverAntennaNoaziPcvSample
	if len(value.NoAzimuthPCV) != 0 {
		bytes, e := checkedNativeAllocationSize(len(value.NoAzimuthPCV), unsafe.Sizeof(C.SidereonReceiverAntennaNoaziPcvSample{}))
		if e != nil {
			return C.SidereonReceiverAntennaCalibration{}, e
		}
		memory, e := alloc.malloc(bytes, "RTK receiver no-azimuth PCV samples")
		if e != nil {
			return C.SidereonReceiverAntennaCalibration{}, e
		}
		rows := unsafe.Slice((*C.SidereonReceiverAntennaNoaziPcvSample)(memory), len(value.NoAzimuthPCV))
		for i, sample := range value.NoAzimuthPCV {
			rows[i] = C.SidereonReceiverAntennaNoaziPcvSample{zenith_deg: C.double(sample.ZenithDeg), value_m: C.double(sample.ValueM)}
		}
		noAzimuth = &rows[0]
	}
	var azimuth *C.SidereonReceiverAntennaAzimuthPcvSample
	if len(value.AzimuthPCV) != 0 {
		bytes, e := checkedNativeAllocationSize(len(value.AzimuthPCV), unsafe.Sizeof(C.SidereonReceiverAntennaAzimuthPcvSample{}))
		if e != nil {
			return C.SidereonReceiverAntennaCalibration{}, e
		}
		memory, e := alloc.malloc(bytes, "RTK receiver azimuth PCV samples")
		if e != nil {
			return C.SidereonReceiverAntennaCalibration{}, e
		}
		rows := unsafe.Slice((*C.SidereonReceiverAntennaAzimuthPcvSample)(memory), len(value.AzimuthPCV))
		for i, sample := range value.AzimuthPCV {
			rows[i] = C.SidereonReceiverAntennaAzimuthPcvSample{azimuth_deg: C.double(sample.AzimuthDeg), zenith_deg: C.double(sample.ZenithDeg), value_m: C.double(sample.ValueM)}
		}
		azimuth = &rows[0]
	}
	return C.SidereonReceiverAntennaCalibration{pco_neu_m: [3]C.double{C.double(value.PCONEUM[0]), C.double(value.PCONEUM[1]), C.double(value.PCONEUM[2])}, noazi_pcv_m: noAzimuth, noazi_pcv_count: noAzimuthCount, azimuth_pcv_m: azimuth, azimuth_pcv_count: azimuthCount}, nil
}

func copyRtkReceiverAntenna(value *RtkReceiverAntennaCorrections, alloc *cRtkAlloc) (*C.SidereonRtkReceiverAntennaCorrections, error) {
	if value == nil {
		return nil, nil
	}
	base, err := copyRtkReceiverAntennaCalibration(value.Base, alloc)
	if err != nil {
		return nil, err
	}
	rover, err := copyRtkReceiverAntennaCalibration(value.Rover, alloc)
	if err != nil {
		return nil, err
	}
	bytes, err := checkedNativeAllocationSize(1, unsafe.Sizeof(C.SidereonRtkReceiverAntennaCorrections{}))
	if err != nil {
		return nil, err
	}
	memory, err := alloc.malloc(bytes, "RTK receiver antenna corrections")
	if err != nil {
		return nil, err
	}
	corrections := (*C.SidereonRtkReceiverAntennaCorrections)(memory)
	*corrections = C.SidereonRtkReceiverAntennaCorrections{base: base, rover: rover}
	return corrections, nil
}

func copyRtkFloatOnlySystems(values []uint32, alloc *cRtkAlloc) (*C.uint32_t, C.size_t, error) {
	count, err := checkedNativeSize(len(values))
	if err != nil {
		return nil, 0, err
	}
	for _, value := range values {
		if err := validateGNSSSystemValue(value); err != nil {
			return nil, 0, err
		}
	}
	if len(values) == 0 {
		return nil, count, nil
	}
	bytes, err := checkedNativeAllocationSize(len(values), unsafe.Sizeof(C.uint32_t(0)))
	if err != nil {
		return nil, 0, err
	}
	memory, err := alloc.malloc(bytes, "RTK fixed float-only systems")
	if err != nil {
		return nil, 0, err
	}
	rows := unsafe.Slice((*C.uint32_t)(memory), len(values))
	for i, value := range values {
		rows[i] = C.uint32_t(value)
	}
	return &rows[0], count, nil
}

func (c RtkFixedConfig) toC(alloc *cRtkAlloc) (*C.SidereonRtkFixedConfig, error) {
	epochCount, err := checkedNativeSize(len(c.Epochs))
	if err != nil {
		return nil, err
	}
	epochs, err := copyRtkEpochs(c.Epochs, alloc)
	if err != nil {
		return nil, err
	}
	ids, idCount, err := copyRtkAmbiguityIDs(c.AmbiguityIDs, alloc)
	if err != nil {
		return nil, err
	}
	ambiguitySatellites, ambiguitySatelliteCount, err := copyRtkAmbiguitySatellites(c.AmbiguitySatellites, alloc)
	if err != nil {
		return nil, err
	}
	wavelengths, wavelengthCount, err := copyRtkFloatMapEntries(c.Wavelengths, alloc, "RTK fixed wavelengths")
	if err != nil {
		return nil, err
	}
	offsets, offsetCount, err := copyRtkFloatMapEntries(c.Offsets, alloc, "RTK fixed offsets")
	if err != nil {
		return nil, err
	}
	antenna, err := copyRtkReceiverAntenna(c.ReceiverAntenna, alloc)
	if err != nil {
		return nil, err
	}
	floatOnlySystems, floatOnlySystemCount, err := copyRtkFloatOnlySystems(c.FloatOnlySystems, alloc)
	if err != nil {
		return nil, err
	}
	floatMaxIterations, err := checkedNativeSize(c.FloatOptions.MaxIterations)
	if err != nil {
		return nil, fmt.Errorf("sidereon: RTK fixed float max iterations: %w", err)
	}
	fixedMaxIterations, err := checkedNativeSize(c.FixedOptions.MaxIterations)
	if err != nil {
		return nil, fmt.Errorf("sidereon: RTK fixed max iterations: %w", err)
	}
	partialMinAmbiguities, err := checkedNativeSize(c.FixedOptions.PartialMinAmbiguities)
	if err != nil {
		return nil, fmt.Errorf("sidereon: RTK fixed partial ambiguities: %w", err)
	}
	maxExclusions, err := checkedNativeSize(c.ResidualOptions.MaxExclusions)
	if err != nil {
		return nil, fmt.Errorf("sidereon: RTK fixed residual exclusions: %w", err)
	}
	if c.Model.Stochastic > 1 {
		return nil, invalidArgument("RTK measurement stochastic model is out of range")
	}
	model := C.SidereonRtkMeasurementModel{code_sigma_m: C.double(c.Model.CodeSigmaM), phase_sigma_m: C.double(c.Model.PhaseSigmaM), sagnac: C.bool(c.Model.Sagnac), stochastic: C.uint32_t(c.Model.Stochastic), elevation_weighting: C.bool(c.Model.ElevationWeighting)}
	floatOptions := C.SidereonRtkFloatOptions{position_tol_m: C.double(c.FloatOptions.PositionTolM), ambiguity_tol_m: C.double(c.FloatOptions.AmbiguityTolM), max_iterations: floatMaxIterations}
	fixedOptions := C.SidereonRtkFixedOptions{position_tol_m: C.double(c.FixedOptions.PositionTolM), ambiguity_tol_m: C.double(c.FixedOptions.AmbiguityTolM), max_iterations: fixedMaxIterations, ratio_threshold: C.double(c.FixedOptions.RatioThreshold), partial_ambiguity_resolution: C.bool(c.FixedOptions.PartialAmbiguityResolution), partial_min_ambiguities: partialMinAmbiguities}
	residualOptions := C.SidereonRtkResidualValidationOptions{threshold_sigma_enabled: C.bool(c.ResidualOptions.ThresholdSigmaEnabled), threshold_sigma: C.double(c.ResidualOptions.ThresholdSigma), max_exclusions: maxExclusions}
	configBytes, err := checkedNativeAllocationSize(1, unsafe.Sizeof(C.SidereonRtkFixedConfig{}))
	if err != nil {
		return nil, err
	}
	memory, err := alloc.malloc(configBytes, "RTK fixed config")
	if err != nil {
		return nil, err
	}
	config := (*C.SidereonRtkFixedConfig)(memory)
	*config = C.SidereonRtkFixedConfig{epochs: epochs, epoch_count: epochCount, base_ecef_m: [3]C.double{C.double(c.BaseECEFM[0]), C.double(c.BaseECEFM[1]), C.double(c.BaseECEFM[2])}, ambiguity_ids: ids, ambiguity_id_count: idCount, ambiguity_satellites: ambiguitySatellites, ambiguity_satellite_count: ambiguitySatelliteCount, wavelengths_m: wavelengths, wavelength_count: wavelengthCount, offsets_m: offsets, offset_count: offsetCount, model: model, receiver_antenna: antenna, float_options: floatOptions, fixed_options: fixedOptions, residual_options: residualOptions, float_only_systems: floatOnlySystems, float_only_system_count: floatOnlySystemCount, initial_baseline_m: [3]C.double{C.double(c.InitialBaselineM[0]), C.double(c.InitialBaselineM[1]), C.double(c.InitialBaselineM[2])}}
	return config, nil
}

func SolveRtkFixed(config RtkFixedConfig) (*RtkFixedSolution, error) {
	alloc := new(cRtkAlloc)
	defer alloc.close()
	cConfig, err := config.toC(alloc)
	if err != nil {
		return nil, err
	}
	var pointer *C.SidereonRtkFixedSolution
	var operationErr error
	withCThread(func() {
		status := C.sidereon_solve_rtk_fixed(cConfig, &pointer)
		operationErr = statusErrorLocked(uint32(status))
		if operationErr != nil && pointer != nil {
			C.sidereon_rtk_fixed_solution_free(pointer)
			pointer = nil
		}
	})
	if operationErr != nil {
		return nil, operationErr
	}
	return newRtkFixedSolution(pointer)
}

func (s *RtkFixedSolution) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return s.handle.close()
}

func (s *RtkFixedSolution) FixedAmbiguities() ([]RtkFixedAmbiguity, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var result []RtkFixedAmbiguity
	var operationErr error
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			var written, required C.size_t
			status := C.sidereon_rtk_fixed_solution_fixed_ambiguities((*C.SidereonRtkFixedSolution)(pointer), nil, 0, &written, &required)
			if operationErr = statusErrorLocked(uint32(status)); operationErr != nil {
				return
			}
			count, e := validateNativeQuery("RTK fixed ambiguities", uint64(written), uint64(required))
			if e != nil {
				operationErr = e
				return
			}
			if _, e = checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonRtkFixedAmbiguity{})); e != nil {
				operationErr = e
				return
			}
			rows := make([]C.SidereonRtkFixedAmbiguity, count)
			var output *C.SidereonRtkFixedAmbiguity
			if count != 0 {
				output = &rows[0]
			}
			written, required = 0, 0
			status = C.sidereon_rtk_fixed_solution_fixed_ambiguities((*C.SidereonRtkFixedSolution)(pointer), output, C.size_t(count), &written, &required)
			if operationErr = statusErrorLocked(uint32(status)); operationErr != nil {
				return
			}
			writtenCount, e := validateTwoPassCounts("RTK fixed ambiguities", count, count, uint64(written), uint64(required))
			if e != nil {
				operationErr = e
				return
			}
			result = make([]RtkFixedAmbiguity, writtenCount)
			for i := range result {
				result[i] = RtkFixedAmbiguity{ID: rtkIDFromC(rows[i].id).Value, Cycles: int64(rows[i].cycles), ValueM: float64(rows[i].value_m)}
			}
		})
		return operationErr
	})
	return result, err
}

func (s *RtkFixedSolution) baseline(call func(unsafe.Pointer, *C.double, C.size_t) uint32) ([3]float64, error) {
	var result [3]float64
	if s == nil || s.handle == nil {
		return result, ErrClosed
	}
	var values [3]C.double
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 { return call(pointer, &values[0], 3) })
	})
	if err != nil {
		return result, err
	}
	for i := range result {
		result[i] = float64(values[i])
	}
	return result, nil
}

func (s *RtkFixedSolution) FixedBaselineECEF() ([3]float64, error) {
	return s.baseline(func(pointer unsafe.Pointer, out *C.double, length C.size_t) uint32 {
		return uint32(C.sidereon_rtk_fixed_solution_fixed_baseline_ecef((*C.SidereonRtkFixedSolution)(pointer), out, length))
	})
}

func (s *RtkFixedSolution) FixedBaselineENU() ([3]float64, error) {
	return s.baseline(func(pointer unsafe.Pointer, out *C.double, length C.size_t) uint32 {
		return uint32(C.sidereon_rtk_fixed_solution_fixed_baseline_enu((*C.SidereonRtkFixedSolution)(pointer), out, length))
	})
}

func (s *RtkFixedSolution) FloatBaselineECEF() ([3]float64, error) {
	return s.baseline(func(pointer unsafe.Pointer, out *C.double, length C.size_t) uint32 {
		return uint32(C.sidereon_rtk_fixed_solution_float_baseline_ecef((*C.SidereonRtkFixedSolution)(pointer), out, length))
	})
}

func (s *RtkFixedSolution) FreeAmbiguities() ([]RtkAmbiguity, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var result []RtkAmbiguity
	var operationErr error
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			var written, required C.size_t
			status := C.sidereon_rtk_fixed_solution_free_ambiguities((*C.SidereonRtkFixedSolution)(pointer), nil, 0, &written, &required)
			if operationErr = statusErrorLocked(uint32(status)); operationErr != nil {
				return
			}
			count, e := validateNativeQuery("RTK fixed free ambiguities", uint64(written), uint64(required))
			if e != nil {
				operationErr = e
				return
			}
			if _, e = checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonRtkAmbiguity{})); e != nil {
				operationErr = e
				return
			}
			rows := make([]C.SidereonRtkAmbiguity, count)
			var output *C.SidereonRtkAmbiguity
			if count != 0 {
				output = &rows[0]
			}
			written, required = 0, 0
			status = C.sidereon_rtk_fixed_solution_free_ambiguities((*C.SidereonRtkFixedSolution)(pointer), output, C.size_t(count), &written, &required)
			if operationErr = statusErrorLocked(uint32(status)); operationErr != nil {
				return
			}
			writtenCount, e := validateTwoPassCounts("RTK fixed free ambiguities", count, count, uint64(written), uint64(required))
			if e != nil {
				operationErr = e
				return
			}
			result = make([]RtkAmbiguity, writtenCount)
			for i := range result {
				result[i] = RtkAmbiguity{ID: rtkIDFromC(rows[i].id).Value, ValueM: float64(rows[i].value_m)}
			}
		})
		return operationErr
	})
	return result, err
}

func (s *RtkFixedSolution) Metadata() (RtkFixedMetadata, error) {
	if s == nil || s.handle == nil {
		return RtkFixedMetadata{}, ErrClosed
	}
	var value C.SidereonRtkFixedMetadata
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_rtk_fixed_solution_metadata((*C.SidereonRtkFixedSolution)(pointer), &value))
		})
	})
	if err != nil {
		return RtkFixedMetadata{}, err
	}
	iterations, err := sizeTToInt(value.iterations, "RTK fixed metadata iterations")
	if err != nil {
		return RtkFixedMetadata{}, err
	}
	nObservations, err := sizeTToInt(value.n_observations, "RTK fixed metadata observations")
	if err != nil {
		return RtkFixedMetadata{}, err
	}
	freeAmbiguities, err := sizeTToInt(value.free_ambiguity_count, "RTK fixed metadata free ambiguities")
	if err != nil {
		return RtkFixedMetadata{}, err
	}
	fixedAmbiguities, err := sizeTToInt(value.fixed_ambiguity_count, "RTK fixed metadata fixed ambiguities")
	if err != nil {
		return RtkFixedMetadata{}, err
	}
	residuals, err := sizeTToInt(value.residual_count, "RTK fixed metadata residuals")
	if err != nil {
		return RtkFixedMetadata{}, err
	}
	usedSatellites, err := sizeTToInt(value.used_sat_count, "RTK fixed metadata satellites")
	if err != nil {
		return RtkFixedMetadata{}, err
	}
	candidates, err := sizeTToInt(value.integer_candidates, "RTK fixed metadata integer candidates")
	if err != nil {
		return RtkFixedMetadata{}, err
	}
	if value.status != C.SIDEREON_RTK_SOLVE_STATUS_STATE_TOLERANCE && value.status != C.SIDEREON_RTK_SOLVE_STATUS_MAX_ITERATIONS {
		return RtkFixedMetadata{}, fmt.Errorf("sidereon: native RTK fixed metadata solve status %d is invalid", uint32(value.status))
	}
	if value.integer_status != C.SIDEREON_RTK_INTEGER_STATUS_FIXED && value.integer_status != C.SIDEREON_RTK_INTEGER_STATUS_NOT_FIXED {
		return RtkFixedMetadata{}, fmt.Errorf("sidereon: native RTK fixed metadata integer status %d is invalid", uint32(value.integer_status))
	}
	geometry, err := geometryFromC(value.geometry_quality)
	if err != nil {
		return RtkFixedMetadata{}, err
	}
	return RtkFixedMetadata{Iterations: iterations, NObservations: nObservations, FreeAmbiguityCount: freeAmbiguities, FixedAmbiguityCount: fixedAmbiguities, ResidualCount: residuals, UsedSatCount: usedSatellites, Converged: bool(value.converged), Status: uint32(value.status), IntegerStatus: uint32(value.integer_status), CodeRMSM: float64(value.code_rms_m), PhaseRMSM: float64(value.phase_rms_m), WeightedRMSM: float64(value.weighted_rms_m), HasIntegerRatio: bool(value.has_integer_ratio), IntegerRatio: float64(value.integer_ratio), HasIntegerBestScore: bool(value.has_integer_best_score), IntegerBestScore: float64(value.integer_best_score), HasIntegerSecondBestScore: bool(value.has_integer_second_best_score), IntegerSecondBestScore: float64(value.integer_second_best_score), IntegerCandidates: candidates, GeometryQuality: geometry}, nil
}

func (s *RtkFixedSolution) UsedSatelliteIDs() ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var result []string
	var operationErr error
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			result, operationErr = copyNativeTokensLocked("RTK fixed used satellites", func(out *C.SidereonSatelliteToken, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_rtk_fixed_solution_used_sat_ids((*C.SidereonRtkFixedSolution)(pointer), out, length, written, required)
			})
		})
		return operationErr
	})
	return result, err
}
