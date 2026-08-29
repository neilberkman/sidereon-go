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

type RtkArcObservationInput struct {
	SatelliteID, AmbiguityID string
	CodeM, PhaseM            float64
	HasLLI                   bool
	LLI                      int64
}

type RtkArcPositionEntry struct {
	ID        string
	PositionM [3]float64
}

type RtkArcEpochInput struct {
	Base, Rover             []RtkArcObservationInput
	SatellitePositions      []RtkArcPositionEntry
	BaseSatellitePositions  []RtkArcPositionEntry
	RoverSatellitePositions []RtkArcPositionEntry
	HasVelocityMPS          bool
	VelocityMPS             [3]float64
	HasPredictionTime       bool
	PredictionTimeS         float64
}

type RtkArcReferenceEntry struct {
	System      uint32
	SatelliteID string
}

type RtkArcPreprocessing struct {
	HasCycleSlip        bool
	CycleSlip           uint32
	HasHatchWindowCap   bool
	HatchWindowCap      int
	HasElevationMaskDeg bool
	ElevationMaskDeg    float64
}

type RtkArcEpochMetadata struct {
	ReportedBaselineM, FloatBaselineM              [3]float64
	IntegerFixed                                   bool
	IntegerRatio                                   float64
	NewlyFixedCount, FixedIDCount                  int
	FixedDoubleDifferenceCount, UsedSatelliteCount int
	SDAmbiguityCount, ResidualCount                int
	HasSearch                                      bool
}

type RtkArcReference struct {
	System, ReferenceID string
}

type RtkArcSplitArc struct {
	Receiver                       uint32
	SatelliteID, AmbiguityID       string
	StartEpochIndex, EndEpochIndex int
	EpochCount                     int
}

type RtkArcConfigInput struct {
	BaseM                [3]float64
	ReferenceMode        uint32
	ReferenceSatellite   string
	ReferencePerSystem   []RtkArcReferenceEntry
	Model                RtkMeasurementModel
	BaselinePriorSigmaM  float64
	AmbiguityPriorSigmaM float64
	InitialBaselineM     [3]float64
	Wavelengths          []RtkFloatMapEntry
	Offsets              []RtkFloatMapEntry
	UpdateOptions        RtkArcUpdateOptions
	ReceiverAntenna      *RtkReceiverAntennaCorrections
	Preprocessing        RtkArcPreprocessing
}

type RtkDualFrequencyObservationInput struct {
	AmbiguityID                      string
	P1M, P2M, Phi1Cycles, Phi2Cycles float64
	F1Hz, F2Hz                       float64
	HasLLI1, HasLLI2                 bool
	LLI1, LLI2                       int64
}

type RtkDualFrequencySatelliteObservationInput struct {
	SatelliteID string
	Base, Rover RtkDualFrequencyObservationInput
}

type RtkDualFrequencyArcEpochInput struct {
	JDWhole, JDFraction     float64
	HasEpochSortKey         bool
	EpochSortKey            string
	HasGapTimeS             bool
	GapTimeS                float64
	Observations            []RtkDualFrequencySatelliteObservationInput
	SatellitePositions      []RtkArcPositionEntry
	BaseSatellitePositions  []RtkArcPositionEntry
	RoverSatellitePositions []RtkArcPositionEntry
	HasVelocityMPS          bool
	VelocityMPS             [3]float64
	HasPredictionTime       bool
	PredictionTimeS         float64
}

type RtkWideLaneCycleInput struct {
	ID     string
	Cycles int64
}

type RtkWideLaneOptions struct {
	MinEpochs       int
	ToleranceCycles float64
	SkipShort       bool
}

type RtkWideLaneConfigInput struct {
	BaseM              [3]float64
	ReferenceMode      uint32
	ReferenceSatellite string
	ReferencePerSystem []RtkArcReferenceEntry
	Options            RtkWideLaneOptions
	HasCycleSlip       bool
	CycleSlipPolicy    uint32
	CycleSlipOptions   CycleSlipOptions
}

type RtkIonosphereFreeConfigInput struct {
	BaseM              [3]float64
	InitialBaselineM   [3]float64
	ReferenceMode      uint32
	ReferenceSatellite string
	ReferencePerSystem []RtkArcReferenceEntry
	ApplyTroposphere   bool
}

type RtkArcSolution struct {
	_      noCopy
	handle *surfaceHandle
}

type RtkIonosphereFreeArcSolution struct {
	_      noCopy
	handle *surfaceHandle
}

type RtkWideLaneArcSolution struct {
	_      noCopy
	handle *surfaceHandle
}

func newRtkArcSolution(pointer *C.SidereonRtkArcSolution) (*RtkArcSolution, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	handle, err := newSurfaceHandle(unsafe.Pointer(pointer), func(value unsafe.Pointer) {
		C.sidereon_rtk_arc_solution_free((*C.SidereonRtkArcSolution)(value))
	})
	if err != nil {
		return nil, err
	}
	return &RtkArcSolution{handle: handle}, nil
}

func newRtkIonosphereFreeArcSolution(pointer *C.SidereonRtkIonosphereFreeArcSolution) (*RtkIonosphereFreeArcSolution, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	handle, err := newSurfaceHandle(unsafe.Pointer(pointer), func(value unsafe.Pointer) {
		C.sidereon_rtk_ionosphere_free_arc_solution_free((*C.SidereonRtkIonosphereFreeArcSolution)(value))
	})
	if err != nil {
		return nil, err
	}
	return &RtkIonosphereFreeArcSolution{handle: handle}, nil
}

func newRtkWideLaneArcSolution(pointer *C.SidereonRtkWideLaneArcSolution) (*RtkWideLaneArcSolution, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	handle, err := newSurfaceHandle(unsafe.Pointer(pointer), func(value unsafe.Pointer) {
		C.sidereon_rtk_wide_lane_arc_solution_free((*C.SidereonRtkWideLaneArcSolution)(value))
	})
	if err != nil {
		return nil, err
	}
	return &RtkWideLaneArcSolution{handle: handle}, nil
}

func copyRtkArcObservations(values []RtkArcObservationInput, alloc *cRtkAlloc) (*C.SidereonRtkArcObservation, C.size_t, error) {
	count, err := checkedNativeSize(len(values))
	if err != nil {
		return nil, 0, err
	}
	if len(values) == 0 {
		return nil, count, nil
	}
	bytes, err := checkedNativeAllocationSize(len(values), unsafe.Sizeof(C.SidereonRtkArcObservation{}))
	if err != nil {
		return nil, 0, err
	}
	memory, err := alloc.malloc(bytes, "RTK arc observations")
	if err != nil {
		return nil, 0, err
	}
	rows := unsafe.Slice((*C.SidereonRtkArcObservation)(memory), len(values))
	for index, value := range values {
		sat, e := alloc.cstring(value.SatelliteID, "RTK arc satellite ID")
		if e != nil {
			return nil, 0, e
		}
		ambiguity, e := alloc.cstring(value.AmbiguityID, "RTK arc ambiguity ID")
		if e != nil {
			return nil, 0, e
		}
		rows[index] = C.SidereonRtkArcObservation{sat_id: sat, ambiguity_id: ambiguity, code_m: C.double(value.CodeM), phase_m: C.double(value.PhaseM), has_lli: C.bool(value.HasLLI), lli: C.int64_t(value.LLI)}
	}
	return &rows[0], count, nil
}

func copyRtkArcPositions(values []RtkArcPositionEntry, alloc *cRtkAlloc, label string) (*C.SidereonRtkArcPositionEntry, C.size_t, error) {
	count, err := checkedNativeSize(len(values))
	if err != nil {
		return nil, 0, err
	}
	if len(values) == 0 {
		return nil, count, nil
	}
	bytes, err := checkedNativeAllocationSize(len(values), unsafe.Sizeof(C.SidereonRtkArcPositionEntry{}))
	if err != nil {
		return nil, 0, err
	}
	memory, err := alloc.malloc(bytes, label)
	if err != nil {
		return nil, 0, err
	}
	rows := unsafe.Slice((*C.SidereonRtkArcPositionEntry)(memory), len(values))
	for index, value := range values {
		id, e := alloc.cstring(value.ID, "RTK arc position satellite ID")
		if e != nil {
			return nil, 0, e
		}
		rows[index] = C.SidereonRtkArcPositionEntry{id: id, pos: [3]C.double{C.double(value.PositionM[0]), C.double(value.PositionM[1]), C.double(value.PositionM[2])}}
	}
	return &rows[0], count, nil
}

func copyRtkArcEpochInputs(values []RtkArcEpochInput, alloc *cRtkAlloc) (*C.SidereonRtkArcEpoch, C.size_t, error) {
	count, err := checkedNativeSize(len(values))
	if err != nil {
		return nil, 0, err
	}
	if len(values) == 0 {
		return nil, count, nil
	}
	bytes, err := checkedNativeAllocationSize(len(values), unsafe.Sizeof(C.SidereonRtkArcEpoch{}))
	if err != nil {
		return nil, 0, err
	}
	memory, err := alloc.malloc(bytes, "RTK arc epochs")
	if err != nil {
		return nil, 0, err
	}
	rows := unsafe.Slice((*C.SidereonRtkArcEpoch)(memory), len(values))
	for index, value := range values {
		base, baseCount, e := copyRtkArcObservations(value.Base, alloc)
		if e != nil {
			return nil, 0, e
		}
		rover, roverCount, e := copyRtkArcObservations(value.Rover, alloc)
		if e != nil {
			return nil, 0, e
		}
		positions, positionCount, e := copyRtkArcPositions(value.SatellitePositions, alloc, "RTK arc satellite positions")
		if e != nil {
			return nil, 0, e
		}
		basePositions, basePositionCount, e := copyRtkArcPositions(value.BaseSatellitePositions, alloc, "RTK arc base satellite positions")
		if e != nil {
			return nil, 0, e
		}
		roverPositions, roverPositionCount, e := copyRtkArcPositions(value.RoverSatellitePositions, alloc, "RTK arc rover satellite positions")
		if e != nil {
			return nil, 0, e
		}
		rows[index] = C.SidereonRtkArcEpoch{base: base, base_count: baseCount, rover: rover, rover_count: roverCount, satellite_positions: positions, satellite_position_count: positionCount, base_satellite_positions: basePositions, base_satellite_position_count: basePositionCount, rover_satellite_positions: roverPositions, rover_satellite_position_count: roverPositionCount, has_velocity_mps: C.bool(value.HasVelocityMPS), velocity_mps: [3]C.double{C.double(value.VelocityMPS[0]), C.double(value.VelocityMPS[1]), C.double(value.VelocityMPS[2])}, has_prediction_time: C.bool(value.HasPredictionTime), prediction_time_s: C.double(value.PredictionTimeS)}
	}
	return &rows[0], count, nil
}

func copyRtkArcReferences(values []RtkArcReferenceEntry, alloc *cRtkAlloc, label string) (*C.SidereonRtkArcReferenceEntry, C.size_t, error) {
	count, err := checkedNativeSize(len(values))
	if err != nil {
		return nil, 0, err
	}
	if len(values) == 0 {
		return nil, count, nil
	}
	bytes, err := checkedNativeAllocationSize(len(values), unsafe.Sizeof(C.SidereonRtkArcReferenceEntry{}))
	if err != nil {
		return nil, 0, err
	}
	memory, err := alloc.malloc(bytes, label)
	if err != nil {
		return nil, 0, err
	}
	rows := unsafe.Slice((*C.SidereonRtkArcReferenceEntry)(memory), len(values))
	for index, value := range values {
		if err := validateGNSSSystemValue(value.System); err != nil {
			return nil, 0, err
		}
		sat, e := alloc.cstring(value.SatelliteID, "RTK reference satellite ID")
		if e != nil {
			return nil, 0, e
		}
		rows[index] = C.SidereonRtkArcReferenceEntry{system: C.enum_SidereonGnssSystem(value.System), sat_id: sat}
	}
	return &rows[0], count, nil
}

func copyRtkFloatOnlySystemsForArc(values []uint32, alloc *cRtkAlloc) (*C.uint32_t, C.size_t, error) {
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
	memory, err := alloc.malloc(bytes, "RTK arc float-only systems")
	if err != nil {
		return nil, 0, err
	}
	rows := unsafe.Slice((*C.uint32_t)(memory), len(values))
	for index, value := range values {
		rows[index] = C.uint32_t(value)
	}
	return &rows[0], count, nil
}

func copyRtkArcConfig(value RtkArcConfigInput, alloc *cRtkAlloc) (*C.SidereonRtkArcConfig, error) {
	if value.ReferenceMode > 2 {
		return nil, invalidArgument("RTK arc reference mode is out of range")
	}
	if value.Model.Stochastic > 1 {
		return nil, invalidArgument("RTK arc measurement stochastic model is out of range")
	}
	if value.Preprocessing.CycleSlip > 2 {
		return nil, invalidArgument("RTK arc cycle-slip policy is out of range")
	}
	referenceSatellite, err := (*C.char)(nil), error(nil)
	if value.ReferenceSatellite != "" {
		referenceSatellite, err = alloc.cstring(value.ReferenceSatellite, "RTK arc reference satellite ID")
		if err != nil {
			return nil, err
		}
	}
	references, referenceCount, err := copyRtkArcReferences(value.ReferencePerSystem, alloc, "RTK arc per-system references")
	if err != nil {
		return nil, err
	}
	wavelengths, wavelengthCount, err := copyRtkFloatMapEntries(value.Wavelengths, alloc, "RTK arc wavelengths")
	if err != nil {
		return nil, err
	}
	offsets, offsetCount, err := copyRtkFloatMapEntries(value.Offsets, alloc, "RTK arc offsets")
	if err != nil {
		return nil, err
	}
	floatOnly, floatOnlyCount, err := copyRtkFloatOnlySystemsForArc(value.UpdateOptions.FloatOnlySystems, alloc)
	if err != nil {
		return nil, err
	}
	antennaValue := value.ReceiverAntenna
	if antennaValue == nil {
		antennaValue = value.UpdateOptions.ReceiverAntenna
	}
	antenna, err := copyRtkReceiverAntenna(antennaValue, alloc)
	if err != nil {
		return nil, err
	}
	maxIterations, err := checkedNativeSize(value.UpdateOptions.MaxIterations)
	if err != nil {
		return nil, fmt.Errorf("sidereon: RTK arc max iterations: %w", err)
	}
	hatchWindow, err := checkedNativeSize(value.Preprocessing.HatchWindowCap)
	if err != nil {
		return nil, fmt.Errorf("sidereon: RTK arc hatch window cap: %w", err)
	}
	update := C.SidereonRtkArcUpdateOptions{hold_sigma_m: C.double(value.UpdateOptions.HoldSigmaM), position_tol_m: C.double(value.UpdateOptions.PositionTolM), ambiguity_tol_m: C.double(value.UpdateOptions.AmbiguityTolM), max_iterations: maxIterations, process_noise_baseline_sigma_m: C.double(value.UpdateOptions.ProcessNoiseBaselineSigmaM), dynamics_velocity_propagated: C.bool(value.UpdateOptions.DynamicsVelocityPropagated), float_only_systems: floatOnly, float_only_system_count: floatOnlyCount, report_residuals: C.bool(value.UpdateOptions.ReportResiduals), has_ar_arming_sigma_m: C.bool(value.UpdateOptions.HasARArmingSigmaM), ar_arming_sigma_m: C.double(value.UpdateOptions.ARArmingSigmaM), ratio_threshold: C.double(value.UpdateOptions.RatioThreshold), receiver_antenna: antenna}
	preprocessing := C.SidereonRtkArcPreprocessing{has_cycle_slip: C.bool(value.Preprocessing.HasCycleSlip), cycle_slip: C.uint32_t(value.Preprocessing.CycleSlip), has_hatch_window_cap: C.bool(value.Preprocessing.HasHatchWindowCap), hatch_window_cap: hatchWindow, has_elevation_mask_deg: C.bool(value.Preprocessing.HasElevationMaskDeg), elevation_mask_deg: C.double(value.Preprocessing.ElevationMaskDeg)}
	bytes, err := checkedNativeAllocationSize(1, unsafe.Sizeof(C.SidereonRtkArcConfig{}))
	if err != nil {
		return nil, err
	}
	memory, err := alloc.malloc(bytes, "RTK arc config")
	if err != nil {
		return nil, err
	}
	config := (*C.SidereonRtkArcConfig)(memory)
	*config = C.SidereonRtkArcConfig{base_m: [3]C.double{C.double(value.BaseM[0]), C.double(value.BaseM[1]), C.double(value.BaseM[2])}, reference_mode: C.uint32_t(value.ReferenceMode), reference_satellite: referenceSatellite, reference_per_system: references, reference_per_system_count: referenceCount, model: C.SidereonRtkMeasurementModel{code_sigma_m: C.double(value.Model.CodeSigmaM), phase_sigma_m: C.double(value.Model.PhaseSigmaM), sagnac: C.bool(value.Model.Sagnac), stochastic: C.uint32_t(value.Model.Stochastic), elevation_weighting: C.bool(value.Model.ElevationWeighting)}, baseline_prior_sigma_m: C.double(value.BaselinePriorSigmaM), ambiguity_prior_sigma_m: C.double(value.AmbiguityPriorSigmaM), initial_baseline_m: [3]C.double{C.double(value.InitialBaselineM[0]), C.double(value.InitialBaselineM[1]), C.double(value.InitialBaselineM[2])}, wavelengths_m: wavelengths, wavelength_count: wavelengthCount, offsets_m: offsets, offset_count: offsetCount, update_options: update, preprocessing: preprocessing}
	return config, nil
}

func copyRtkDualObservation(value RtkDualFrequencyObservationInput, alloc *cRtkAlloc, label string) (C.SidereonRtkDualFrequencyObservation, error) {
	id, err := alloc.cstring(value.AmbiguityID, label+" ambiguity ID")
	if err != nil {
		return C.SidereonRtkDualFrequencyObservation{}, err
	}
	return C.SidereonRtkDualFrequencyObservation{ambiguity_id: id, p1_m: C.double(value.P1M), p2_m: C.double(value.P2M), phi1_cycles: C.double(value.Phi1Cycles), phi2_cycles: C.double(value.Phi2Cycles), f1_hz: C.double(value.F1Hz), f2_hz: C.double(value.F2Hz), has_lli1: C.bool(value.HasLLI1), lli1: C.int64_t(value.LLI1), has_lli2: C.bool(value.HasLLI2), lli2: C.int64_t(value.LLI2)}, nil
}

func copyRtkDualFrequencyEpochs(values []RtkDualFrequencyArcEpochInput, alloc *cRtkAlloc) (*C.SidereonRtkDualFrequencyArcEpoch, C.size_t, error) {
	count, err := checkedNativeSize(len(values))
	if err != nil {
		return nil, 0, err
	}
	if len(values) == 0 {
		return nil, count, nil
	}
	bytes, err := checkedNativeAllocationSize(len(values), unsafe.Sizeof(C.SidereonRtkDualFrequencyArcEpoch{}))
	if err != nil {
		return nil, 0, err
	}
	memory, err := alloc.malloc(bytes, "RTK dual-frequency arc epochs")
	if err != nil {
		return nil, 0, err
	}
	rows := unsafe.Slice((*C.SidereonRtkDualFrequencyArcEpoch)(memory), len(values))
	for index, value := range values {
		var sortKey *C.char
		if value.HasEpochSortKey {
			sortKey, err = alloc.cstring(value.EpochSortKey, "RTK dual-frequency epoch sort key")
			if err != nil {
				return nil, 0, err
			}
		}
		observationsCount, e := checkedNativeSize(len(value.Observations))
		if e != nil {
			return nil, 0, e
		}
		var observations *C.SidereonRtkDualFrequencySatelliteObservation
		if len(value.Observations) != 0 {
			rowBytes, e := checkedNativeAllocationSize(len(value.Observations), unsafe.Sizeof(C.SidereonRtkDualFrequencySatelliteObservation{}))
			if e != nil {
				return nil, 0, e
			}
			rowMemory, e := alloc.malloc(rowBytes, "RTK dual-frequency observations")
			if e != nil {
				return nil, 0, e
			}
			row := unsafe.Slice((*C.SidereonRtkDualFrequencySatelliteObservation)(rowMemory), len(value.Observations))
			for observationIndex, observation := range value.Observations {
				satellite, e := alloc.cstring(observation.SatelliteID, "RTK dual-frequency satellite ID")
				if e != nil {
					return nil, 0, e
				}
				base, e := copyRtkDualObservation(observation.Base, alloc, "RTK dual-frequency base")
				if e != nil {
					return nil, 0, e
				}
				rover, e := copyRtkDualObservation(observation.Rover, alloc, "RTK dual-frequency rover")
				if e != nil {
					return nil, 0, e
				}
				row[observationIndex] = C.SidereonRtkDualFrequencySatelliteObservation{sat_id: satellite, base: base, rover: rover}
			}
			observations = &row[0]
		}
		satellitePositions, satellitePositionCount, e := copyRtkArcPositions(value.SatellitePositions, alloc, "RTK dual-frequency satellite positions")
		if e != nil {
			return nil, 0, e
		}
		basePositions, basePositionCount, e := copyRtkArcPositions(value.BaseSatellitePositions, alloc, "RTK dual-frequency base positions")
		if e != nil {
			return nil, 0, e
		}
		roverPositions, roverPositionCount, e := copyRtkArcPositions(value.RoverSatellitePositions, alloc, "RTK dual-frequency rover positions")
		if e != nil {
			return nil, 0, e
		}
		rows[index] = C.SidereonRtkDualFrequencyArcEpoch{jd_whole: C.double(value.JDWhole), jd_fraction: C.double(value.JDFraction), epoch_sort_key: sortKey, has_gap_time_s: C.bool(value.HasGapTimeS), gap_time_s: C.double(value.GapTimeS), observations: observations, observation_count: observationsCount, satellite_positions: satellitePositions, satellite_position_count: satellitePositionCount, base_satellite_positions: basePositions, base_satellite_position_count: basePositionCount, rover_satellite_positions: roverPositions, rover_satellite_position_count: roverPositionCount, has_velocity_mps: C.bool(value.HasVelocityMPS), velocity_mps: [3]C.double{C.double(value.VelocityMPS[0]), C.double(value.VelocityMPS[1]), C.double(value.VelocityMPS[2])}, has_prediction_time: C.bool(value.HasPredictionTime), prediction_time_s: C.double(value.PredictionTimeS)}
	}
	return &rows[0], count, nil
}

func copyRtkWideLaneCycles(values []RtkWideLaneCycleInput, alloc *cRtkAlloc) (*C.SidereonRtkWideLaneCycle, C.size_t, error) {
	count, err := checkedNativeSize(len(values))
	if err != nil {
		return nil, 0, err
	}
	if len(values) == 0 {
		return nil, count, nil
	}
	bytes, err := checkedNativeAllocationSize(len(values), unsafe.Sizeof(C.SidereonRtkWideLaneCycle{}))
	if err != nil {
		return nil, 0, err
	}
	memory, err := alloc.malloc(bytes, "RTK wide-lane cycles")
	if err != nil {
		return nil, 0, err
	}
	rows := unsafe.Slice((*C.SidereonRtkWideLaneCycle)(memory), len(values))
	for index, value := range values {
		if err := copyRtkID(rows[index].id.bytes[:], value.ID, "RTK wide-lane ambiguity ID"); err != nil {
			return nil, 0, err
		}
		rows[index].cycles = C.int64_t(value.Cycles)
	}
	return &rows[0], count, nil
}

func copyRtkID(dst []C.char, value, label string) error {
	if err := rejectEmbeddedNUL(value, label); err != nil {
		return err
	}
	if len(value) >= len(dst) {
		return fmt.Errorf("sidereon: %s is too long", label)
	}
	for index := range dst {
		dst[index] = 0
	}
	for index, item := range []byte(value) {
		dst[index] = C.char(item)
	}
	return nil
}

func copyRtkWideLaneConfig(value RtkWideLaneConfigInput, alloc *cRtkAlloc) (*C.SidereonRtkWideLaneArcConfig, error) {
	if value.ReferenceMode > 2 {
		return nil, invalidArgument("RTK wide-lane reference mode is out of range")
	}
	if value.CycleSlipPolicy > 2 {
		return nil, invalidArgument("RTK wide-lane cycle-slip policy is out of range")
	}
	minEpochs, err := checkedNativeSize(value.Options.MinEpochs)
	if err != nil {
		return nil, fmt.Errorf("sidereon: RTK wide-lane minimum epochs: %w", err)
	}
	var referenceSatellite *C.char
	if value.ReferenceSatellite != "" {
		referenceSatellite, err = alloc.cstring(value.ReferenceSatellite, "RTK wide-lane reference satellite ID")
		if err != nil {
			return nil, err
		}
	}
	references, referenceCount, err := copyRtkArcReferences(value.ReferencePerSystem, alloc, "RTK wide-lane per-system references")
	if err != nil {
		return nil, err
	}
	options := C.SidereonRtkWideLaneOptions{min_epochs: minEpochs, tolerance_cycles: C.double(value.Options.ToleranceCycles), skip_short_fragments: C.bool(value.Options.SkipShort)}
	cycleSlip := C.SidereonRtkDualCycleSlipConfig{policy: C.uint32_t(value.CycleSlipPolicy), options: C.SidereonCycleSlipOptions{gf_threshold_m: C.double(value.CycleSlipOptions.GFThresholdM), mw_threshold_cycles: C.double(value.CycleSlipOptions.MWThresholdCycles), min_arc_gap_s: C.double(value.CycleSlipOptions.MinArcGapS)}}
	bytes, err := checkedNativeAllocationSize(1, unsafe.Sizeof(C.SidereonRtkWideLaneArcConfig{}))
	if err != nil {
		return nil, err
	}
	memory, err := alloc.malloc(bytes, "RTK wide-lane arc config")
	if err != nil {
		return nil, err
	}
	config := (*C.SidereonRtkWideLaneArcConfig)(memory)
	*config = C.SidereonRtkWideLaneArcConfig{base_m: [3]C.double{C.double(value.BaseM[0]), C.double(value.BaseM[1]), C.double(value.BaseM[2])}, reference_mode: C.uint32_t(value.ReferenceMode), reference_satellite: referenceSatellite, reference_per_system: references, reference_per_system_count: referenceCount, options: options, has_cycle_slip: C.bool(value.HasCycleSlip), cycle_slip: cycleSlip}
	return config, nil
}

func copyRtkIonosphereFreeConfig(value RtkIonosphereFreeConfigInput, alloc *cRtkAlloc) (*C.SidereonRtkIonosphereFreeArcConfig, error) {
	if value.ReferenceMode > 2 {
		return nil, invalidArgument("RTK ionosphere-free reference mode is out of range")
	}
	var referenceSatellite *C.char
	var err error
	if value.ReferenceSatellite != "" {
		referenceSatellite, err = alloc.cstring(value.ReferenceSatellite, "RTK ionosphere-free reference satellite ID")
		if err != nil {
			return nil, err
		}
	}
	references, referenceCount, err := copyRtkArcReferences(value.ReferencePerSystem, alloc, "RTK ionosphere-free per-system references")
	if err != nil {
		return nil, err
	}
	bytes, err := checkedNativeAllocationSize(1, unsafe.Sizeof(C.SidereonRtkIonosphereFreeArcConfig{}))
	if err != nil {
		return nil, err
	}
	memory, err := alloc.malloc(bytes, "RTK ionosphere-free arc config")
	if err != nil {
		return nil, err
	}
	config := (*C.SidereonRtkIonosphereFreeArcConfig)(memory)
	*config = C.SidereonRtkIonosphereFreeArcConfig{base_m: [3]C.double{C.double(value.BaseM[0]), C.double(value.BaseM[1]), C.double(value.BaseM[2])}, initial_baseline_m: [3]C.double{C.double(value.InitialBaselineM[0]), C.double(value.InitialBaselineM[1]), C.double(value.InitialBaselineM[2])}, reference_mode: C.uint32_t(value.ReferenceMode), reference_satellite: referenceSatellite, reference_per_system: references, reference_per_system_count: referenceCount, apply_troposphere: C.bool(value.ApplyTroposphere)}
	return config, nil
}

func SolveRtkArc(epochs []RtkArcEpochInput, config RtkArcConfigInput) (*RtkArcSolution, error) {
	alloc := new(cRtkAlloc)
	defer alloc.close()
	cEpochs, epochCount, err := copyRtkArcEpochInputs(epochs, alloc)
	if err != nil {
		return nil, err
	}
	cConfig, err := copyRtkArcConfig(config, alloc)
	if err != nil {
		return nil, err
	}
	var pointer *C.SidereonRtkArcSolution
	var operationErr error
	withCThread(func() {
		status := C.sidereon_solve_rtk_arc(cEpochs, epochCount, cConfig, &pointer)
		operationErr = statusErrorLocked(uint32(status))
		if operationErr != nil && pointer != nil {
			C.sidereon_rtk_arc_solution_free(pointer)
			pointer = nil
		}
	})
	if operationErr != nil {
		return nil, operationErr
	}
	return newRtkArcSolution(pointer)
}

func FixWideLaneRtkArc(epochs []RtkDualFrequencyArcEpochInput, config RtkWideLaneConfigInput) (*RtkWideLaneArcSolution, error) {
	alloc := new(cRtkAlloc)
	defer alloc.close()
	cEpochs, epochCount, err := copyRtkDualFrequencyEpochs(epochs, alloc)
	if err != nil {
		return nil, err
	}
	cConfig, err := copyRtkWideLaneConfig(config, alloc)
	if err != nil {
		return nil, err
	}
	var pointer *C.SidereonRtkWideLaneArcSolution
	var operationErr error
	withCThread(func() {
		status := C.sidereon_fix_wide_lane_rtk_arc(cEpochs, epochCount, cConfig, &pointer)
		operationErr = statusErrorLocked(uint32(status))
		if operationErr != nil && pointer != nil {
			C.sidereon_rtk_wide_lane_arc_solution_free(pointer)
			pointer = nil
		}
	})
	if operationErr != nil {
		return nil, operationErr
	}
	return newRtkWideLaneArcSolution(pointer)
}

func PrepareIonosphereFreeRtkArc(epochs []RtkDualFrequencyArcEpochInput, cycles []RtkWideLaneCycleInput, config RtkIonosphereFreeConfigInput) (*RtkIonosphereFreeArcSolution, error) {
	alloc := new(cRtkAlloc)
	defer alloc.close()
	cEpochs, epochCount, err := copyRtkDualFrequencyEpochs(epochs, alloc)
	if err != nil {
		return nil, err
	}
	cCycles, cycleCount, err := copyRtkWideLaneCycles(cycles, alloc)
	if err != nil {
		return nil, err
	}
	cConfig, err := copyRtkIonosphereFreeConfig(config, alloc)
	if err != nil {
		return nil, err
	}
	var pointer *C.SidereonRtkIonosphereFreeArcSolution
	var operationErr error
	withCThread(func() {
		status := C.sidereon_prepare_ionosphere_free_rtk_arc(cEpochs, epochCount, cCycles, cycleCount, cConfig, &pointer)
		operationErr = statusErrorLocked(uint32(status))
		if operationErr != nil && pointer != nil {
			C.sidereon_rtk_ionosphere_free_arc_solution_free(pointer)
			pointer = nil
		}
	})
	if operationErr != nil {
		return nil, operationErr
	}
	return newRtkIonosphereFreeArcSolution(pointer)
}

func (s *RtkArcSolution) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return s.handle.close()
}

func (s *RtkIonosphereFreeArcSolution) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return s.handle.close()
}

func (s *RtkWideLaneArcSolution) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return s.handle.close()
}

func checkedRtkArcSolutionIndex(index int) (C.size_t, error) {
	return checkedRtkRinexIndex(index, "RTK arc solution epoch")
}

type rtkIDOutputCall func(*C.SidereonRtkId, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus

func copyRtkIDs(label string, call rtkIDOutputCall) ([]string, error) {
	var written, required C.size_t
	if err := statusErrorLocked(uint32(call(nil, 0, &written, &required))); err != nil {
		return nil, err
	}
	count, err := validateNativeQuery(label, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonRtkId{})); err != nil {
		return nil, err
	}
	capacity, err := cSize(count, label+" output capacity")
	if err != nil {
		return nil, err
	}
	rows := make([]C.SidereonRtkId, count)
	var output *C.SidereonRtkId
	if count != 0 {
		output = &rows[0]
	}
	written, required = 0, 0
	if err := statusErrorLocked(uint32(call(output, capacity, &written, &required))); err != nil {
		return nil, err
	}
	writtenCount, err := validateTwoPassCounts(label, count, count, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	result := make([]string, writtenCount)
	for index := range result {
		result[index] = rtkIDFromC(rows[index]).Value
	}
	return result, nil
}

type rtkFloatOutputCall func(*C.double, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus

func copyRtkFloatValues(label string, call rtkFloatOutputCall) ([]float64, error) {
	var written, required C.size_t
	if err := statusErrorLocked(uint32(call(nil, 0, &written, &required))); err != nil {
		return nil, err
	}
	count, err := validateNativeQuery(label, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.double(0))); err != nil {
		return nil, err
	}
	capacity, err := cSize(count, label+" output capacity")
	if err != nil {
		return nil, err
	}
	rows := make([]C.double, count)
	var output *C.double
	if count != 0 {
		output = &rows[0]
	}
	written, required = 0, 0
	if err := statusErrorLocked(uint32(call(output, capacity, &written, &required))); err != nil {
		return nil, err
	}
	writtenCount, err := validateTwoPassCounts(label, count, count, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	result := make([]float64, writtenCount)
	for index := range result {
		result[index] = float64(rows[index])
	}
	return result, nil
}

func (s *RtkArcSolution) EpochCount() (int, error) {
	if s == nil || s.handle == nil {
		return 0, ErrClosed
	}
	var count C.size_t
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_rtk_arc_solution_epoch_count((*C.SidereonRtkArcSolution)(pointer), &count))
		})
	})
	if err != nil {
		return 0, err
	}
	return sizeTToInt(count, "RTK arc solution epoch count")
}

func (s *RtkArcSolution) FinalEpochCount() (int, error) {
	if s == nil || s.handle == nil {
		return 0, ErrClosed
	}
	var count C.size_t
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_rtk_arc_solution_final_epoch_count((*C.SidereonRtkArcSolution)(pointer), &count))
		})
	})
	if err != nil {
		return 0, err
	}
	return sizeTToInt(count, "RTK arc solution final epoch count")
}

func (s *RtkArcSolution) EpochMetadata(index int) (RtkArcEpochMetadata, error) {
	if s == nil || s.handle == nil {
		return RtkArcEpochMetadata{}, ErrClosed
	}
	nativeIndex, err := checkedRtkArcSolutionIndex(index)
	if err != nil {
		return RtkArcEpochMetadata{}, err
	}
	var value C.SidereonRtkArcEpochMetadata
	err = s.handle.read(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_rtk_arc_solution_epoch_metadata((*C.SidereonRtkArcSolution)(pointer), nativeIndex, &value))
		})
	})
	if err != nil {
		return RtkArcEpochMetadata{}, err
	}
	counts := []*struct {
		value C.size_t
		label string
	}{
		{value.newly_fixed_count, "RTK arc metadata newly fixed count"},
		{value.fixed_id_count, "RTK arc metadata fixed ID count"},
		{value.fixed_double_difference_count, "RTK arc metadata fixed double-difference count"},
		{value.used_satellite_count, "RTK arc metadata used satellite count"},
		{value.sd_ambiguity_count, "RTK arc metadata ambiguity count"},
		{value.residual_count, "RTK arc metadata residual count"},
	}
	converted := make([]int, len(counts))
	for index, count := range counts {
		converted[index], err = sizeTToInt(count.value, count.label)
		if err != nil {
			return RtkArcEpochMetadata{}, err
		}
	}
	return RtkArcEpochMetadata{ReportedBaselineM: [3]float64{float64(value.reported_baseline_m[0]), float64(value.reported_baseline_m[1]), float64(value.reported_baseline_m[2])}, FloatBaselineM: [3]float64{float64(value.float_baseline_m[0]), float64(value.float_baseline_m[1]), float64(value.float_baseline_m[2])}, IntegerFixed: bool(value.integer_fixed), IntegerRatio: float64(value.integer_ratio), NewlyFixedCount: converted[0], FixedIDCount: converted[1], FixedDoubleDifferenceCount: converted[2], UsedSatelliteCount: converted[3], SDAmbiguityCount: converted[4], ResidualCount: converted[5], HasSearch: bool(value.has_search)}, nil
}

func (s *RtkArcSolution) EpochSDAmbiguities(index int) ([]RtkAmbiguity, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	nativeIndex, err := checkedRtkArcSolutionIndex(index)
	if err != nil {
		return nil, err
	}
	var result []RtkAmbiguity
	err = s.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			result, err = copyRtkArcAmbiguities("RTK arc epoch SD ambiguities", func(out *C.SidereonRtkAmbiguity, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_rtk_arc_solution_epoch_sd_ambiguities((*C.SidereonRtkArcSolution)(pointer), nativeIndex, out, length, written, required)
			})
		})
		return err
	})
	return result, err
}

type rtkAmbiguityOutputCall func(*C.SidereonRtkAmbiguity, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus

func copyRtkArcAmbiguities(label string, call rtkAmbiguityOutputCall) ([]RtkAmbiguity, error) {
	var written, required C.size_t
	if err := statusErrorLocked(uint32(call(nil, 0, &written, &required))); err != nil {
		return nil, err
	}
	count, err := validateNativeQuery(label, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonRtkAmbiguity{})); err != nil {
		return nil, err
	}
	capacity, err := cSize(count, label+" output capacity")
	if err != nil {
		return nil, err
	}
	rows := make([]C.SidereonRtkAmbiguity, count)
	var output *C.SidereonRtkAmbiguity
	if count != 0 {
		output = &rows[0]
	}
	written, required = 0, 0
	if err := statusErrorLocked(uint32(call(output, capacity, &written, &required))); err != nil {
		return nil, err
	}
	writtenCount, err := validateTwoPassCounts(label, count, count, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	result := make([]RtkAmbiguity, writtenCount)
	for index := range result {
		result[index] = RtkAmbiguity{ID: rtkIDFromC(rows[index].id).Value, ValueM: float64(rows[index].value_m)}
	}
	return result, nil
}

func (s *RtkArcSolution) EpochStringIDs(index int, which uint32) ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	if which > 2 {
		return nil, invalidArgument("RTK arc epoch ID list is out of range")
	}
	nativeIndex, err := checkedRtkArcSolutionIndex(index)
	if err != nil {
		return nil, err
	}
	var result []string
	err = s.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			result, err = copyRtkIDs("RTK arc epoch string IDs", func(out *C.SidereonRtkId, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_rtk_arc_solution_epoch_string_ids((*C.SidereonRtkArcSolution)(pointer), nativeIndex, C.uint32_t(which), out, length, written, required)
			})
		})
		return err
	})
	return result, err
}

func (s *RtkArcSolution) EpochUsedSatellites(index int) ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	nativeIndex, err := checkedRtkArcSolutionIndex(index)
	if err != nil {
		return nil, err
	}
	var result []string
	err = s.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			result, err = copyNativeTokensLocked("RTK arc epoch used satellites", func(out *C.SidereonSatelliteToken, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_rtk_arc_solution_epoch_used_satellites((*C.SidereonRtkArcSolution)(pointer), nativeIndex, out, length, written, required)
			})
		})
		return err
	})
	return result, err
}

func (s *RtkArcSolution) DroppedSatellites() ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var result []string
	var operationErr error
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			result, operationErr = copyNativeTokensLocked("RTK arc dropped satellites", func(out *C.SidereonSatelliteToken, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_rtk_arc_solution_dropped_sats((*C.SidereonRtkArcSolution)(pointer), out, length, written, required)
			})
		})
		return operationErr
	})
	return result, err
}

func (s *RtkArcSolution) ElevationMaskedSatellites() ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var result []string
	var operationErr error
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			result, operationErr = copyNativeTokensLocked("RTK arc elevation-masked satellites", func(out *C.SidereonSatelliteToken, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_rtk_arc_solution_elevation_masked_sats((*C.SidereonRtkArcSolution)(pointer), out, length, written, required)
			})
		})
		return operationErr
	})
	return result, err
}

func (s *RtkArcSolution) FinalBaseline() ([3]float64, error) {
	var result [3]float64
	if s == nil || s.handle == nil {
		return result, ErrClosed
	}
	var values [3]C.double
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_rtk_arc_solution_final_baseline((*C.SidereonRtkArcSolution)(pointer), &values[0], 3))
		})
	})
	if err != nil {
		return result, err
	}
	for index := range result {
		result[index] = float64(values[index])
	}
	return result, nil
}

func (s *RtkArcSolution) MeasurementCovariance() ([]float64, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var result []float64
	var operationErr error
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			result, operationErr = copyRtkFloatValues("RTK arc measurement covariance", func(out *C.double, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_rtk_arc_solution_measurement_covariance((*C.SidereonRtkArcSolution)(pointer), out, length, written, required)
			})
		})
		return operationErr
	})
	return result, err
}

type rtkReferenceOutputCall func(*C.SidereonRtkArcReferenceOut, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus

func copyRtkReferences(label string, call rtkReferenceOutputCall) ([]RtkArcReference, error) {
	var written, required C.size_t
	if err := statusErrorLocked(uint32(call(nil, 0, &written, &required))); err != nil {
		return nil, err
	}
	count, err := validateNativeQuery(label, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonRtkArcReferenceOut{})); err != nil {
		return nil, err
	}
	capacity, err := cSize(count, label+" output capacity")
	if err != nil {
		return nil, err
	}
	rows := make([]C.SidereonRtkArcReferenceOut, count)
	var output *C.SidereonRtkArcReferenceOut
	if count != 0 {
		output = &rows[0]
	}
	written, required = 0, 0
	if err := statusErrorLocked(uint32(call(output, capacity, &written, &required))); err != nil {
		return nil, err
	}
	writtenCount, err := validateTwoPassCounts(label, count, count, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	result := make([]RtkArcReference, writtenCount)
	for index := range result {
		result[index] = RtkArcReference{System: rtkIDFromC(rows[index].system).Value, ReferenceID: rtkIDFromC(rows[index].reference_id).Value}
	}
	return result, nil
}

func (s *RtkArcSolution) References() ([]RtkArcReference, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var result []RtkArcReference
	var operationErr error
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			result, operationErr = copyRtkReferences("RTK arc references", func(out *C.SidereonRtkArcReferenceOut, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_rtk_arc_solution_references((*C.SidereonRtkArcSolution)(pointer), out, length, written, required)
			})
		})
		return operationErr
	})
	return result, err
}

func (s *RtkArcSolution) SplitCycleSlipArcs() ([]RtkArcSplitArc, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var result []RtkArcSplitArc
	var operationErr error
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			result, operationErr = copyRtkSplitArcs("RTK arc split cycle-slip arcs", func(out *C.SidereonRtkArcSplitArc, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_rtk_arc_solution_split_cycle_slip_arcs((*C.SidereonRtkArcSolution)(pointer), out, length, written, required)
			})
		})
		return operationErr
	})
	return result, err
}

type rtkSplitArcOutputCall func(*C.SidereonRtkArcSplitArc, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus

func copyRtkSplitArcs(label string, call rtkSplitArcOutputCall) ([]RtkArcSplitArc, error) {
	var written, required C.size_t
	if err := statusErrorLocked(uint32(call(nil, 0, &written, &required))); err != nil {
		return nil, err
	}
	count, err := validateNativeQuery(label, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonRtkArcSplitArc{})); err != nil {
		return nil, err
	}
	capacity, err := cSize(count, label+" output capacity")
	if err != nil {
		return nil, err
	}
	rows := make([]C.SidereonRtkArcSplitArc, count)
	var output *C.SidereonRtkArcSplitArc
	if count != 0 {
		output = &rows[0]
	}
	written, required = 0, 0
	if err := statusErrorLocked(uint32(call(output, capacity, &written, &required))); err != nil {
		return nil, err
	}
	writtenCount, err := validateTwoPassCounts(label, count, count, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	result := make([]RtkArcSplitArc, writtenCount)
	for index := range result {
		if rows[index].receiver > 1 {
			return nil, invalidArgument("native RTK split arc receiver is out of range")
		}
		start, e := sizeTToInt(rows[index].start_epoch_index, "RTK split arc start epoch")
		if e != nil {
			return nil, e
		}
		end, e := sizeTToInt(rows[index].end_epoch_index, "RTK split arc end epoch")
		if e != nil {
			return nil, e
		}
		n, e := sizeTToInt(rows[index].n_epochs, "RTK split arc epoch count")
		if e != nil {
			return nil, e
		}
		result[index] = RtkArcSplitArc{Receiver: uint32(rows[index].receiver), SatelliteID: tokenFromC(rows[index].satellite_id), AmbiguityID: rtkIDFromC(rows[index].ambiguity_id).Value, StartEpochIndex: start, EndEpochIndex: end, EpochCount: n}
	}
	return result, nil
}

func (s *RtkIonosphereFreeArcSolution) EpochCount() (int, error) {
	if s == nil || s.handle == nil {
		return 0, ErrClosed
	}
	var count C.size_t
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_rtk_ionosphere_free_arc_solution_epoch_count((*C.SidereonRtkIonosphereFreeArcSolution)(pointer), &count))
		})
	})
	if err != nil {
		return 0, err
	}
	return sizeTToInt(count, "RTK ionosphere-free epoch count")
}

func (s *RtkIonosphereFreeArcSolution) epochObservations(index int, base bool) ([]RtkRinexArcObservation, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	nativeIndex, err := checkedRtkArcSolutionIndex(index)
	if err != nil {
		return nil, err
	}
	var result []RtkRinexArcObservation
	err = s.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			result, err = copyRtkRinexArcObservations("RTK ionosphere-free observations", func(out *C.SidereonRtkArcObservationOut, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				if base {
					return C.sidereon_rtk_ionosphere_free_arc_solution_epoch_base_observations((*C.SidereonRtkIonosphereFreeArcSolution)(pointer), nativeIndex, out, length, written, required)
				}
				return C.sidereon_rtk_ionosphere_free_arc_solution_epoch_rover_observations((*C.SidereonRtkIonosphereFreeArcSolution)(pointer), nativeIndex, out, length, written, required)
			})
		})
		return err
	})
	return result, err
}

func (s *RtkIonosphereFreeArcSolution) EpochBaseObservations(index int) ([]RtkRinexArcObservation, error) {
	return s.epochObservations(index, true)
}

func (s *RtkIonosphereFreeArcSolution) EpochRoverObservations(index int) ([]RtkRinexArcObservation, error) {
	return s.epochObservations(index, false)
}

func (s *RtkIonosphereFreeArcSolution) epochPositions(index int, base, rover bool) ([]RtkRinexArcPosition, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	nativeIndex, err := checkedRtkArcSolutionIndex(index)
	if err != nil {
		return nil, err
	}
	var result []RtkRinexArcPosition
	err = s.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			result, err = copyRtkRinexArcPositions("RTK ionosphere-free positions", func(out *C.SidereonRtkArcPositionOut, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				switch {
				case base:
					return C.sidereon_rtk_ionosphere_free_arc_solution_epoch_base_satellite_positions((*C.SidereonRtkIonosphereFreeArcSolution)(pointer), nativeIndex, out, length, written, required)
				case rover:
					return C.sidereon_rtk_ionosphere_free_arc_solution_epoch_rover_satellite_positions((*C.SidereonRtkIonosphereFreeArcSolution)(pointer), nativeIndex, out, length, written, required)
				default:
					return C.sidereon_rtk_ionosphere_free_arc_solution_epoch_satellite_positions((*C.SidereonRtkIonosphereFreeArcSolution)(pointer), nativeIndex, out, length, written, required)
				}
			})
		})
		return err
	})
	return result, err
}

func (s *RtkIonosphereFreeArcSolution) EpochBaseSatellitePositions(index int) ([]RtkRinexArcPosition, error) {
	return s.epochPositions(index, true, false)
}

func (s *RtkIonosphereFreeArcSolution) EpochRoverSatellitePositions(index int) ([]RtkRinexArcPosition, error) {
	return s.epochPositions(index, false, true)
}

func (s *RtkIonosphereFreeArcSolution) EpochSatellitePositions(index int) ([]RtkRinexArcPosition, error) {
	return s.epochPositions(index, false, false)
}

func (s *RtkIonosphereFreeArcSolution) EpochMetadata(index int) (RtkRinexArcEpochMetadata, error) {
	if s == nil || s.handle == nil {
		return RtkRinexArcEpochMetadata{}, ErrClosed
	}
	nativeIndex, err := checkedRtkArcSolutionIndex(index)
	if err != nil {
		return RtkRinexArcEpochMetadata{}, err
	}
	var value C.SidereonRtkArcEpochOutMetadata
	err = s.handle.read(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_rtk_ionosphere_free_arc_solution_epoch_metadata((*C.SidereonRtkIonosphereFreeArcSolution)(pointer), nativeIndex, &value))
		})
	})
	if err != nil {
		return RtkRinexArcEpochMetadata{}, err
	}
	baseCount, err := sizeTToInt(value.base_count, "RTK ionosphere-free base count")
	if err != nil {
		return RtkRinexArcEpochMetadata{}, err
	}
	roverCount, err := sizeTToInt(value.rover_count, "RTK ionosphere-free rover count")
	if err != nil {
		return RtkRinexArcEpochMetadata{}, err
	}
	positions, err := sizeTToInt(value.satellite_position_count, "RTK ionosphere-free position count")
	if err != nil {
		return RtkRinexArcEpochMetadata{}, err
	}
	basePositions, err := sizeTToInt(value.base_satellite_position_count, "RTK ionosphere-free base position count")
	if err != nil {
		return RtkRinexArcEpochMetadata{}, err
	}
	roverPositions, err := sizeTToInt(value.rover_satellite_position_count, "RTK ionosphere-free rover position count")
	if err != nil {
		return RtkRinexArcEpochMetadata{}, err
	}
	return RtkRinexArcEpochMetadata{BaseCount: baseCount, RoverCount: roverCount, SatellitePositionCount: positions, BaseSatellitePositionCount: basePositions, RoverSatellitePositionCount: roverPositions, HasVelocityMPS: bool(value.has_velocity_mps), VelocityMPS: [3]float64{float64(value.velocity_mps[0]), float64(value.velocity_mps[1]), float64(value.velocity_mps[2])}, HasPredictionTime: bool(value.has_prediction_time), PredictionTimeS: float64(value.prediction_time_s)}, nil
}

func (s *RtkIonosphereFreeArcSolution) mapValues(offsets bool) ([]RtkRinexMapValue, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var result []RtkRinexMapValue
	var operationErr error
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			result, operationErr = copyRtkRinexMapValues("RTK ionosphere-free map values", func(out *C.SidereonRtkMapValue, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				if offsets {
					return C.sidereon_rtk_ionosphere_free_arc_solution_offsets_m((*C.SidereonRtkIonosphereFreeArcSolution)(pointer), out, length, written, required)
				}
				return C.sidereon_rtk_ionosphere_free_arc_solution_wavelengths_m((*C.SidereonRtkIonosphereFreeArcSolution)(pointer), out, length, written, required)
			})
		})
		return operationErr
	})
	return result, err
}

func (s *RtkIonosphereFreeArcSolution) OffsetsM() ([]RtkRinexMapValue, error) {
	return s.mapValues(true)
}

func (s *RtkIonosphereFreeArcSolution) WavelengthsM() ([]RtkRinexMapValue, error) {
	return s.mapValues(false)
}

func (s *RtkIonosphereFreeArcSolution) References() ([]RtkArcReference, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var result []RtkArcReference
	var operationErr error
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			result, operationErr = copyRtkReferences("RTK ionosphere-free references", func(out *C.SidereonRtkArcReferenceOut, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_rtk_ionosphere_free_arc_solution_references((*C.SidereonRtkIonosphereFreeArcSolution)(pointer), out, length, written, required)
			})
		})
		return operationErr
	})
	return result, err
}
