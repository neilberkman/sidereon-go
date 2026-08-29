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

type StaticPositionEpochInput struct {
	Inputs  SppInputsV2
	Weights []float64
}

type StaticPositionOptionsInput struct {
	InitialPositionM [3]float64
	WithGeodetic     bool
	RobustEnabled    bool
	Robust           NativeSPPRobustConfig
}

type StaticReferenceStationRinexConfigInput struct {
	ReferencePositionM [3]float64
	EnableCodeDGNSS    bool
	EnableCarrierRTK   bool
	WithGeodetic       bool
	Carrier            RtkRinexStaticBaselineConfig
}

type StaticPositionClockBias struct {
	EpochIndex int
	System     uint32
	ClockS     float64
}
type StaticPositionEpochInfluence struct {
	EpochIndex, OmittedMeasurements int
	Status                          uint32
	HasPositionDelta                bool
	PositionDeltaM                  [3]float64
	PositionDeltaNormM              float64
	HasResidualRMS                  bool
	ResidualRMSM                    float64
	MinRobustWeightRatio            float64
}
type StaticPositionMetadata struct {
	Iterations, OuterIterations, UsedMeasurements, Parameters int
	Converged, HasFinalRobustScale                            bool
	Status                                                    uint32
	FinalRobustScaleM                                         float64
	Redundancy                                                int64
	GeometryQuality                                           SPPGeometryQuality
}
type StaticPositionResidual struct {
	EpochIndex            int
	SatelliteID           string
	ResidualM, BaseWeight float64
	EffectiveWeight       float64
	RobustWeightRatio     float64
}
type StaticPositionRejectedSat struct {
	SatelliteID string
	Reason      uint32
}
type StaticPositionSatelliteBatchInfluence struct {
	SatelliteID          string
	OmittedMeasurements  int
	Status               uint32
	HasPositionDelta     bool
	PositionDeltaM       [3]float64
	PositionDeltaNormM   float64
	HasResidualRMS       bool
	ResidualRMSM         float64
	MinRobustWeightRatio float64
}
type StaticPositionSatelliteInfluence struct {
	EpochIndex            int
	SatelliteID           string
	Status                uint32
	HasPositionDelta      bool
	PositionDeltaM        [3]float64
	PositionDeltaNormM    float64
	HasResidualRMS        bool
	ResidualRMSM          float64
	ResidualM, BaseWeight float64
	EffectiveWeight       float64
	RobustWeightRatio     float64
}
type StaticReferenceEpochDiagnostic struct {
	Mode, EpochIndex, UsedSatelliteCount, RejectedSatelliteCount int
	HasCodeResidualRMS, HasPhaseResidualRMS, HasResidualRMS      bool
	CodeResidualRMSM, PhaseResidualRMSM, ResidualRMSM            float64
}
type StaticReferenceStationMetadata struct {
	Mode, FixStatus                             uint32
	HasGeodetic                                 bool
	Geodetic                                    Geodetic
	BaselineM                                   float64
	HasCodeSolution, HasCarrierSolution         bool
	DiagnosticCount, ModeReportCount            int
	CarrierIntegerStatus                        uint32
	HasCarrierIntegerRatio                      bool
	CarrierIntegerRatio                         float64
	CodeDiagnosticCount, CarrierDiagnosticCount int
}
type StaticReferenceModeReport struct {
	Mode, Status                                uint32
	UsedEpochs, SkippedEpochs, UsedMeasurements int
	HasError                                    bool
}

const (
	staticPositionErrorKindMax   = uint32(C.SIDEREON_STATIC_POSITION_ERROR_KIND_SINGULAR)
	staticInfluenceStatusMax     = uint32(C.SIDEREON_STATIC_POSITION_INFLUENCE_STATUS_SOLVE_FAILED)
	staticSPPSolveStatusMax      = uint32(C.SIDEREON_SPP_SOLVE_STATUS_MAX_EVALUATIONS)
	staticReferenceModeMax       = uint32(C.SIDEREON_STATIC_REFERENCE_STATION_MODE_CARRIER_FIXED)
	staticReferenceFixStatusMax  = uint32(C.SIDEREON_STATIC_REFERENCE_FIX_STATUS_CARRIER_FIXED)
	staticReferenceModeStatusMax = uint32(C.SIDEREON_STATIC_REFERENCE_MODE_STATUS_FAILED)
	staticRtkIntegerStatusMax    = uint32(C.SIDEREON_RTK_INTEGER_STATUS_NOT_FIXED)
	staticObservabilityTierMax   = uint32(C.SIDEREON_OBSERVABILITY_TIER_NOMINAL)
	staticSppRejectionReasonMax  = uint32(C.SIDEREON_SPP_REJECTION_REASON_SBAS_IONO_UNCOVERED)
)

type StaticPositionSolution struct {
	_      noCopy
	handle *surfaceHandle
}
type StaticReferenceStationSolution struct {
	_      noCopy
	handle *surfaceHandle
}

func newStaticPositionSolution(pointer *C.SidereonStaticPositionSolution) (*StaticPositionSolution, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	h, err := newSurfaceHandle(unsafe.Pointer(pointer), func(p unsafe.Pointer) {
		C.sidereon_static_position_solution_free((*C.SidereonStaticPositionSolution)(p))
	})
	if err != nil {
		return nil, err
	}
	return &StaticPositionSolution{handle: h}, nil
}
func newStaticReferenceStationSolution(pointer *C.SidereonStaticReferenceStationSolution) (*StaticReferenceStationSolution, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	h, err := newSurfaceHandle(unsafe.Pointer(pointer), func(p unsafe.Pointer) {
		C.sidereon_static_reference_station_solution_free((*C.SidereonStaticReferenceStationSolution)(p))
	})
	if err != nil {
		return nil, err
	}
	return &StaticReferenceStationSolution{handle: h}, nil
}

type staticArena struct {
	pointers []unsafe.Pointer
	alloc    cRtkAlloc
}

func (a *staticArena) calloc(count int, size uintptr, label string) (unsafe.Pointer, error) {
	if count == 0 {
		return nil, nil
	}
	bytes, err := checkedNativeAllocationSize(count, size)
	if err != nil {
		return nil, err
	}
	p := C.calloc(1, C.size_t(bytes))
	if p == nil {
		return nil, fmt.Errorf("sidereon: unable to allocate native %s", label)
	}
	a.pointers = append(a.pointers, p)
	return p, nil
}
func (a *staticArena) close() {
	a.alloc.close()
	for i := len(a.pointers) - 1; i >= 0; i-- {
		C.free(a.pointers[i])
	}
}

func makeStaticEpochs(values []StaticPositionEpochInput, arena *staticArena) (*C.SidereonStaticPositionEpoch, C.size_t, error) {
	count, err := checkedNativeSize(len(values))
	if err != nil {
		return nil, 0, err
	}
	if len(values) == 0 {
		return nil, count, nil
	}
	memory, err := arena.calloc(len(values), unsafe.Sizeof(C.SidereonStaticPositionEpoch{}), "static position epochs")
	if err != nil {
		return nil, 0, err
	}
	rows := unsafe.Slice((*C.SidereonStaticPositionEpoch)(memory), len(values))
	for index, value := range values {
		var initErr error
		withCThread(func() { initErr = statusErrorLocked(uint32(C.sidereon_spp_inputs_v2_init(&rows[index].inputs))) })
		if initErr != nil {
			return nil, 0, initErr
		}
		if len(value.Weights) != 0 && len(value.Weights) != len(value.Inputs.Base.Observations) {
			return nil, 0, invalidArgument("static position weights must match observations")
		}
		if e := fillSppV2(&rows[index].inputs, value.Inputs, &arena.alloc); e != nil {
			return nil, 0, e
		}
		if len(value.Weights) != 0 {
			weightCount, e := checkedNativeSize(len(value.Weights))
			if e != nil {
				return nil, 0, e
			}
			weightMemory, e := arena.calloc(len(value.Weights), unsafe.Sizeof(C.double(0)), "static position weights")
			if e != nil {
				return nil, 0, e
			}
			weights := unsafe.Slice((*C.double)(weightMemory), len(value.Weights))
			for i, weight := range value.Weights {
				weights[i] = C.double(weight)
			}
			rows[index].weights = &weights[0]
			rows[index].weight_count = weightCount
		}
	}
	return &rows[0], count, nil
}

func makeStaticOptions(value StaticPositionOptionsInput, arena *staticArena) (*C.SidereonStaticPositionOptions, error) {
	memory, err := arena.calloc(1, unsafe.Sizeof(C.SidereonStaticPositionOptions{}), "static position options")
	if err != nil {
		return nil, err
	}
	options := (*C.SidereonStaticPositionOptions)(memory)
	for i := range value.InitialPositionM {
		options.initial_position_m[i] = C.double(value.InitialPositionM[i])
	}
	options.with_geodetic, options.robust_enabled = C.bool(value.WithGeodetic), C.bool(value.RobustEnabled)
	options.robust.huber_k = C.double(value.Robust.HuberK)
	options.robust.scale_floor_m = C.double(value.Robust.ScaleFloorM)
	maxOuter, err := cSize64(value.Robust.MaxOuter, "static robust max outer")
	if err != nil {
		return nil, err
	}
	options.robust.max_outer = maxOuter
	options.robust.outer_tol_m = C.double(value.Robust.OuterToleranceM)
	return options, nil
}

func solveStaticPosition(source unsafe.Pointer, broadcast bool, epochs []StaticPositionEpochInput, options *StaticPositionOptionsInput) (*StaticPositionSolution, uint32, error) {
	arena := new(staticArena)
	defer arena.close()
	epochPointer, epochCount, err := makeStaticEpochs(epochs, arena)
	if err != nil {
		return nil, 0, err
	}
	var optionPointer *C.SidereonStaticPositionOptions
	if options != nil {
		optionPointer, err = makeStaticOptions(*options, arena)
		if err != nil {
			return nil, 0, err
		}
	}
	var kind C.enum_SidereonStaticPositionErrorKind
	var pointer *C.SidereonStaticPositionSolution
	err = withCThreadError(func() error {
		var status C.enum_SidereonStatus
		if broadcast {
			status = C.sidereon_solve_static_position_broadcast((*C.SidereonBroadcastEphemeris)(source), epochPointer, epochCount, optionPointer, &kind, &pointer)
		} else {
			status = C.sidereon_solve_static_position_sp3((*C.SidereonSp3)(source), epochPointer, epochCount, optionPointer, &kind, &pointer)
		}
		if e := statusErrorLocked(uint32(status)); e != nil {
			if pointer != nil {
				C.sidereon_static_position_solution_free(pointer)
				pointer = nil
			}
			return e
		}
		return nil
	})
	if err != nil {
		return nil, uint32(kind), err
	}
	if uint32(kind) > staticPositionErrorKindMax {
		if pointer != nil {
			C.sidereon_static_position_solution_free(pointer)
		}
		return nil, 0, invalidArgument("invalid static-position error kind returned by native code")
	}
	if pointer == nil {
		return nil, uint32(kind), errors.New("sidereon: native static-position solve returned no solution")
	}
	h, err := newStaticPositionSolution(pointer)
	if err != nil {
		C.sidereon_static_position_solution_free(pointer)
		return nil, 0, err
	}
	return h, uint32(kind), nil
}

func SolveStaticPositionBroadcast(source *BroadcastEphemeris, epochs []StaticPositionEpochInput, options *StaticPositionOptionsInput) (*StaticPositionSolution, uint32, error) {
	if source == nil || source.resource == nil {
		return nil, 0, ErrClosed
	}
	var result *StaticPositionSolution
	var kind uint32
	err := source.resource.with(func(pointer unsafe.Pointer) error {
		var e error
		result, kind, e = solveStaticPosition(pointer, true, epochs, options)
		return e
	})
	runtime.KeepAlive(source)
	runtime.KeepAlive(epochs)
	return result, kind, err
}

func SolveStaticPositionSP3(source *SP3, epochs []StaticPositionEpochInput, options *StaticPositionOptionsInput) (*StaticPositionSolution, uint32, error) {
	if source == nil || source.handle == nil {
		return nil, 0, ErrClosed
	}
	var result *StaticPositionSolution
	var kind uint32
	err := source.handle.with(func(pointer unsafe.Pointer) error {
		var e error
		result, kind, e = solveStaticPosition(pointer, false, epochs, options)
		return e
	})
	runtime.KeepAlive(source)
	runtime.KeepAlive(epochs)
	return result, kind, err
}

func StaticPositionOptionsInit() (StaticPositionOptionsInput, error) {
	var value C.SidereonStaticPositionOptions
	if err := callStatus(func() uint32 { return uint32(C.sidereon_static_position_options_init(&value)) }); err != nil {
		return StaticPositionOptionsInput{}, err
	}
	return StaticPositionOptionsInput{InitialPositionM: [3]float64{float64(value.initial_position_m[0]), float64(value.initial_position_m[1]), float64(value.initial_position_m[2])}, WithGeodetic: bool(value.with_geodetic), RobustEnabled: bool(value.robust_enabled), Robust: NativeSPPRobustConfig{HuberK: float64(value.robust.huber_k), ScaleFloorM: float64(value.robust.scale_floor_m), MaxOuter: uint64(value.robust.max_outer), OuterToleranceM: float64(value.robust.outer_tol_m)}}, nil
}

func copyStaticFixed(label string, h *surfaceHandle, call func(unsafe.Pointer, *C.double, C.size_t) C.enum_SidereonStatus, count int) ([3]float64, error) {
	var out [3]C.double
	err := h.read(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			if count != 3 {
				return invalidArgument(label + " requires three values")
			}
			return statusErrorLocked(uint32(call(pointer, &out[0], 3)))
		})
	})
	return [3]float64{float64(out[0]), float64(out[1]), float64(out[2])}, err
}
func staticOutputMemory(n int, size uintptr, label string) (unsafe.Pointer, error) {
	bytes, err := checkedNativeAllocationSize(n, size)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, nil
	}
	memory := C.malloc(C.size_t(bytes))
	if memory == nil {
		return nil, fmt.Errorf("sidereon: unable to allocate native %s", label)
	}
	return memory, nil
}
func staticClockFromC(v C.SidereonStaticPositionClockBias) (StaticPositionClockBias, error) {
	n, e := sizeTToInt(v.epoch_index, "static clock epoch index")
	return StaticPositionClockBias{EpochIndex: n, System: uint32(v.system), ClockS: float64(v.clock_s)}, e
}
func staticEpochInfluenceFromC(v C.SidereonStaticPositionEpochInfluence) (StaticPositionEpochInfluence, error) {
	i, e := sizeTToInt(v.epoch_index, "static influence epoch index")
	if e != nil {
		return StaticPositionEpochInfluence{}, e
	}
	m, e := sizeTToInt(v.omitted_measurements, "static influence omitted measurements")
	return StaticPositionEpochInfluence{EpochIndex: i, OmittedMeasurements: m, Status: uint32(v.status), HasPositionDelta: bool(v.has_position_delta_m), PositionDeltaM: [3]float64{float64(v.position_delta_m[0]), float64(v.position_delta_m[1]), float64(v.position_delta_m[2])}, PositionDeltaNormM: float64(v.position_delta_norm_m), HasResidualRMS: bool(v.has_residual_rms_m), ResidualRMSM: float64(v.residual_rms_m), MinRobustWeightRatio: float64(v.min_robust_weight_ratio)}, e
}
func staticResidualFromC(v C.SidereonStaticPositionResidual) (StaticPositionResidual, error) {
	i, e := sizeTToInt(v.epoch_index, "static residual epoch index")
	return StaticPositionResidual{EpochIndex: i, SatelliteID: tokenFromC(v.sat_id), ResidualM: float64(v.residual_m), BaseWeight: float64(v.base_weight), EffectiveWeight: float64(v.effective_weight), RobustWeightRatio: float64(v.robust_weight_ratio)}, e
}
func staticBatchInfluenceFromC(v C.SidereonStaticPositionSatelliteBatchInfluence) (StaticPositionSatelliteBatchInfluence, error) {
	m, e := sizeTToInt(v.omitted_measurements, "static batch omitted measurements")
	return StaticPositionSatelliteBatchInfluence{SatelliteID: tokenFromC(v.sat_id), OmittedMeasurements: m, Status: uint32(v.status), HasPositionDelta: bool(v.has_position_delta_m), PositionDeltaM: [3]float64{float64(v.position_delta_m[0]), float64(v.position_delta_m[1]), float64(v.position_delta_m[2])}, PositionDeltaNormM: float64(v.position_delta_norm_m), HasResidualRMS: bool(v.has_residual_rms_m), ResidualRMSM: float64(v.residual_rms_m), MinRobustWeightRatio: float64(v.min_robust_weight_ratio)}, e
}
func staticInfluenceFromC(v C.SidereonStaticPositionSatelliteInfluence) (StaticPositionSatelliteInfluence, error) {
	i, e := sizeTToInt(v.epoch_index, "static satellite influence epoch index")
	if e != nil {
		return StaticPositionSatelliteInfluence{}, e
	}
	return StaticPositionSatelliteInfluence{EpochIndex: i, SatelliteID: tokenFromC(v.sat_id), Status: uint32(v.status), HasPositionDelta: bool(v.has_position_delta_m), PositionDeltaM: [3]float64{float64(v.position_delta_m[0]), float64(v.position_delta_m[1]), float64(v.position_delta_m[2])}, PositionDeltaNormM: float64(v.position_delta_norm_m), HasResidualRMS: bool(v.has_residual_rms_m), ResidualRMSM: float64(v.residual_rms_m), ResidualM: float64(v.residual_m), BaseWeight: float64(v.base_weight), EffectiveWeight: float64(v.effective_weight), RobustWeightRatio: float64(v.robust_weight_ratio)}, e
}
func staticDiagnosticFromC(v C.SidereonStaticReferenceEpochDiagnostic) (StaticReferenceEpochDiagnostic, error) {
	i, e := sizeTToInt(v.epoch_index, "static diagnostic epoch index")
	if e != nil {
		return StaticReferenceEpochDiagnostic{}, e
	}
	u, e := sizeTToInt(v.used_satellite_count, "static diagnostic used satellites")
	if e != nil {
		return StaticReferenceEpochDiagnostic{}, e
	}
	r, e := sizeTToInt(v.rejected_satellite_count, "static diagnostic rejected satellites")
	return StaticReferenceEpochDiagnostic{Mode: int(v.mode), EpochIndex: i, UsedSatelliteCount: u, RejectedSatelliteCount: r, HasCodeResidualRMS: bool(v.has_code_residual_rms_m), CodeResidualRMSM: float64(v.code_residual_rms_m), HasPhaseResidualRMS: bool(v.has_phase_residual_rms_m), PhaseResidualRMSM: float64(v.phase_residual_rms_m), HasResidualRMS: bool(v.has_residual_rms_m), ResidualRMSM: float64(v.residual_rms_m)}, e
}
func staticModeReportFromC(v C.SidereonStaticReferenceModeReport) (StaticReferenceModeReport, error) {
	u, e := sizeTToInt(v.used_epochs, "static mode used epochs")
	if e != nil {
		return StaticReferenceModeReport{}, e
	}
	s, e := sizeTToInt(v.skipped_epochs, "static mode skipped epochs")
	if e != nil {
		return StaticReferenceModeReport{}, e
	}
	m, e := sizeTToInt(v.used_measurements, "static mode used measurements")
	return StaticReferenceModeReport{Mode: uint32(v.mode), Status: uint32(v.status), UsedEpochs: u, SkippedEpochs: s, UsedMeasurements: m, HasError: bool(v.has_error)}, e
}
func staticMetadataFromC(v C.SidereonStaticPositionMetadata) (StaticPositionMetadata, error) {
	if uint32(v.status) > staticSPPSolveStatusMax || uint32(v.geometry_quality.tier) > staticObservabilityTierMax {
		return StaticPositionMetadata{}, invalidArgument("invalid static metadata enum")
	}
	i, e := sizeTToInt(v.iterations, "static metadata iterations")
	if e != nil {
		return StaticPositionMetadata{}, e
	}
	o, e := sizeTToInt(v.outer_iterations, "static metadata outer iterations")
	if e != nil {
		return StaticPositionMetadata{}, e
	}
	u, e := sizeTToInt(v.used_measurements, "static metadata used measurements")
	if e != nil {
		return StaticPositionMetadata{}, e
	}
	p, e := sizeTToInt(v.n_parameters, "static metadata parameters")
	if e != nil {
		return StaticPositionMetadata{}, e
	}
	r, e := sizeTToInt(v.geometry_quality.rank, "static metadata rank")
	if e != nil {
		return StaticPositionMetadata{}, e
	}
	return StaticPositionMetadata{Iterations: i, Converged: bool(v.converged), Status: uint32(v.status), OuterIterations: o, HasFinalRobustScale: bool(v.has_final_robust_scale_m), FinalRobustScaleM: float64(v.final_robust_scale_m), UsedMeasurements: u, Parameters: p, Redundancy: int64(v.redundancy), GeometryQuality: SPPGeometryQuality{Tier: uint32(v.geometry_quality.tier), Redundancy: int32(v.geometry_quality.redundancy), Rank: r, ConditionNumber: float64(v.geometry_quality.condition_number), GDOP: float64(v.geometry_quality.gdop), RAIMCheckable: bool(v.geometry_quality.raim_checkable), CovarianceValidated: bool(v.geometry_quality.covariance_validated)}}, nil
}
func staticReferenceMetadataFromC(v C.SidereonStaticReferenceStationMetadata) (StaticReferenceStationMetadata, error) {
	if uint32(v.mode) > staticReferenceModeMax || uint32(v.fix_status) > staticReferenceFixStatusMax || uint32(v.carrier_integer_status) > staticRtkIntegerStatusMax {
		return StaticReferenceStationMetadata{}, invalidArgument("invalid static reference metadata enum")
	}
	d, e := sizeTToInt(v.diagnostic_count, "static reference diagnostic count")
	if e != nil {
		return StaticReferenceStationMetadata{}, e
	}
	m, e := sizeTToInt(v.mode_report_count, "static reference mode report count")
	if e != nil {
		return StaticReferenceStationMetadata{}, e
	}
	c, e := sizeTToInt(v.code_diagnostic_count, "static reference code diagnostic count")
	if e != nil {
		return StaticReferenceStationMetadata{}, e
	}
	k, e := sizeTToInt(v.carrier_diagnostic_count, "static reference carrier diagnostic count")
	return StaticReferenceStationMetadata{Mode: uint32(v.mode), FixStatus: uint32(v.fix_status), HasGeodetic: bool(v.has_geodetic), Geodetic: geodeticFromC(v.geodetic), BaselineM: float64(v.baseline_m), HasCodeSolution: bool(v.has_code_solution), HasCarrierSolution: bool(v.has_carrier_solution), DiagnosticCount: d, ModeReportCount: m, CarrierIntegerStatus: uint32(v.carrier_integer_status), HasCarrierIntegerRatio: bool(v.has_carrier_integer_ratio), CarrierIntegerRatio: float64(v.carrier_integer_ratio), CodeDiagnosticCount: c, CarrierDiagnosticCount: k}, e
}
func copyStaticCovariance(label string, h *surfaceHandle, call func(unsafe.Pointer, *C.double, C.size_t) C.enum_SidereonStatus) ([9]float64, error) {
	var out [9]C.double
	err := h.read(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error { return statusErrorLocked(uint32(call(pointer, &out[0], 9))) })
	})
	var result [9]float64
	for i := range result {
		result[i] = float64(out[i])
	}
	return result, err
}

func (s *StaticPositionSolution) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return s.handle.close()
}
func (s *StaticPositionSolution) Position() ([3]float64, error) {
	if s == nil || s.handle == nil {
		return [3]float64{}, ErrClosed
	}
	return copyStaticFixed("static position", s.handle, func(p unsafe.Pointer, o *C.double, n C.size_t) C.enum_SidereonStatus {
		return C.sidereon_static_position_solution_position((*C.SidereonStaticPositionSolution)(p), o, n)
	}, 3)
}
func (s *StaticPositionSolution) PositionCovarianceECEFM2() ([9]float64, error) {
	if s == nil || s.handle == nil {
		return [9]float64{}, ErrClosed
	}
	return copyStaticCovariance("static position ECEF covariance", s.handle, func(p unsafe.Pointer, o *C.double, n C.size_t) C.enum_SidereonStatus {
		return C.sidereon_static_position_solution_position_covariance_ecef_m2((*C.SidereonStaticPositionSolution)(p), o, n)
	})
}
func (s *StaticPositionSolution) PositionCovarianceENUM2() ([9]float64, error) {
	if s == nil || s.handle == nil {
		return [9]float64{}, ErrClosed
	}
	return copyStaticCovariance("static position ENU covariance", s.handle, func(p unsafe.Pointer, o *C.double, n C.size_t) C.enum_SidereonStatus {
		return C.sidereon_static_position_solution_position_covariance_enu_m2((*C.SidereonStaticPositionSolution)(p), o, n)
	})
}

// The row conversion functions intentionally use typed C arrays instead of
// unsafe generic layout; this keeps every C value detached before returning.
func (s *StaticPositionSolution) ClockBiases() ([]StaticPositionClockBias, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var w, r C.size_t
	var result []StaticPositionClockBias
	err := s.handle.read(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			if e := statusErrorLocked(uint32(C.sidereon_static_position_solution_clock_biases((*C.SidereonStaticPositionSolution)(p), nil, 0, &w, &r))); e != nil {
				return e
			}
			n, e := validateNativeQuery("static clock biases", uint64(w), uint64(r))
			if e != nil {
				return e
			}
			if _, e = checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonStaticPositionClockBias{})); e != nil {
				return e
			}
			if n == 0 {
				result = []StaticPositionClockBias{}
				return nil
			}
			mem, e := staticOutputMemory(n, unsafe.Sizeof(C.SidereonStaticPositionClockBias{}), "static clock biases")
			if e != nil {
				return e
			}
			defer C.free(mem)
			w, r = 0, 0
			if e = statusErrorLocked(uint32(C.sidereon_static_position_solution_clock_biases((*C.SidereonStaticPositionSolution)(p), (*C.SidereonStaticPositionClockBias)(mem), C.size_t(n), &w, &r))); e != nil {
				return e
			}
			rows, e := validateTwoPassCounts("static clock biases", n, n, uint64(w), uint64(r))
			if e != nil {
				return e
			}
			raw := unsafe.Slice((*C.SidereonStaticPositionClockBias)(mem), rows)
			result = make([]StaticPositionClockBias, rows)
			for i, v := range raw {
				if e = validateGNSSSystemValue(uint32(v.system)); e != nil {
					return e
				}
				result[i], e = staticClockFromC(v)
				if e != nil {
					return e
				}
			}
			return nil
		})
	})
	return result, err
}

func (s *StaticPositionSolution) EpochInfluence() ([]StaticPositionEpochInfluence, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var w, r C.size_t
	var result []StaticPositionEpochInfluence
	err := s.handle.read(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			if e := statusErrorLocked(uint32(C.sidereon_static_position_solution_epoch_influence((*C.SidereonStaticPositionSolution)(p), nil, 0, &w, &r))); e != nil {
				return e
			}
			n, e := validateNativeQuery("static epoch influence", uint64(w), uint64(r))
			if e != nil {
				return e
			}
			mem, e := staticOutputMemory(n, unsafe.Sizeof(C.SidereonStaticPositionEpochInfluence{}), "static epoch influence")
			if e != nil {
				return e
			}
			if mem != nil {
				defer C.free(mem)
			}
			w, r = 0, 0
			if e = statusErrorLocked(uint32(C.sidereon_static_position_solution_epoch_influence((*C.SidereonStaticPositionSolution)(p), (*C.SidereonStaticPositionEpochInfluence)(mem), C.size_t(n), &w, &r))); e != nil {
				return e
			}
			rows, e := validateTwoPassCounts("static epoch influence", n, n, uint64(w), uint64(r))
			if e != nil {
				return e
			}
			raw := unsafe.Slice((*C.SidereonStaticPositionEpochInfluence)(mem), rows)
			result = make([]StaticPositionEpochInfluence, rows)
			for i, v := range raw {
				if uint32(v.status) > staticInfluenceStatusMax {
					return invalidArgument("invalid static influence status")
				}
				result[i], e = staticEpochInfluenceFromC(v)
				if e != nil {
					return e
				}
			}
			return nil
		})
	})
	return result, err
}

func (s *StaticPositionSolution) RejectedSats(epoch int) ([]StaticPositionRejectedSat, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	if epoch < 0 {
		return nil, invalidArgument("static position epoch index must not be negative")
	}
	var w, r C.size_t
	var result []StaticPositionRejectedSat
	err := s.handle.read(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			if e := statusErrorLocked(uint32(C.sidereon_static_position_solution_rejected_sats((*C.SidereonStaticPositionSolution)(p), C.size_t(epoch), nil, 0, &w, &r))); e != nil {
				return e
			}
			n, e := validateNativeQuery("static rejected satellites", uint64(w), uint64(r))
			if e != nil {
				return e
			}
			mem, e := staticOutputMemory(n, unsafe.Sizeof(C.SidereonSppRejectedSat{}), "static rejected satellites")
			if e != nil {
				return e
			}
			if mem != nil {
				defer C.free(mem)
			}
			w, r = 0, 0
			if e = statusErrorLocked(uint32(C.sidereon_static_position_solution_rejected_sats((*C.SidereonStaticPositionSolution)(p), C.size_t(epoch), (*C.SidereonSppRejectedSat)(mem), C.size_t(n), &w, &r))); e != nil {
				return e
			}
			rows, e := validateTwoPassCounts("static rejected satellites", n, n, uint64(w), uint64(r))
			if e != nil {
				return e
			}
			raw := unsafe.Slice((*C.SidereonSppRejectedSat)(mem), rows)
			result = make([]StaticPositionRejectedSat, rows)
			for i, v := range raw {
				if uint32(v.reason) > staticSppRejectionReasonMax {
					return invalidArgument("invalid static rejection reason")
				}
				result[i] = StaticPositionRejectedSat{SatelliteID: tokenFromC(v.sat_id), Reason: uint32(v.reason)}
			}
			return nil
		})
	})
	return result, err
}

func (s *StaticPositionSolution) Residuals() ([]StaticPositionResidual, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var w, r C.size_t
	var result []StaticPositionResidual
	err := s.handle.read(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			if e := statusErrorLocked(uint32(C.sidereon_static_position_solution_residuals((*C.SidereonStaticPositionSolution)(p), nil, 0, &w, &r))); e != nil {
				return e
			}
			n, e := validateNativeQuery("static residuals", uint64(w), uint64(r))
			if e != nil {
				return e
			}
			mem, e := staticOutputMemory(n, unsafe.Sizeof(C.SidereonStaticPositionResidual{}), "static residuals")
			if e != nil {
				return e
			}
			if mem != nil {
				defer C.free(mem)
			}
			w, r = 0, 0
			if e = statusErrorLocked(uint32(C.sidereon_static_position_solution_residuals((*C.SidereonStaticPositionSolution)(p), (*C.SidereonStaticPositionResidual)(mem), C.size_t(n), &w, &r))); e != nil {
				return e
			}
			rows, e := validateTwoPassCounts("static residuals", n, n, uint64(w), uint64(r))
			if e != nil {
				return e
			}
			raw := unsafe.Slice((*C.SidereonStaticPositionResidual)(mem), rows)
			result = make([]StaticPositionResidual, rows)
			for i, v := range raw {
				result[i], e = staticResidualFromC(v)
				if e != nil {
					return e
				}
			}
			return nil
		})
	})
	return result, err
}

func (s *StaticPositionSolution) SatelliteBatchInfluence() ([]StaticPositionSatelliteBatchInfluence, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var w, r C.size_t
	var result []StaticPositionSatelliteBatchInfluence
	err := s.handle.read(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			if e := statusErrorLocked(uint32(C.sidereon_static_position_solution_satellite_batch_influence((*C.SidereonStaticPositionSolution)(p), nil, 0, &w, &r))); e != nil {
				return e
			}
			n, e := validateNativeQuery("static satellite batch influence", uint64(w), uint64(r))
			if e != nil {
				return e
			}
			mem, e := staticOutputMemory(n, unsafe.Sizeof(C.SidereonStaticPositionSatelliteBatchInfluence{}), "static satellite batch influence")
			if e != nil {
				return e
			}
			if mem != nil {
				defer C.free(mem)
			}
			w, r = 0, 0
			if e = statusErrorLocked(uint32(C.sidereon_static_position_solution_satellite_batch_influence((*C.SidereonStaticPositionSolution)(p), (*C.SidereonStaticPositionSatelliteBatchInfluence)(mem), C.size_t(n), &w, &r))); e != nil {
				return e
			}
			rows, e := validateTwoPassCounts("static satellite batch influence", n, n, uint64(w), uint64(r))
			if e != nil {
				return e
			}
			raw := unsafe.Slice((*C.SidereonStaticPositionSatelliteBatchInfluence)(mem), rows)
			result = make([]StaticPositionSatelliteBatchInfluence, rows)
			for i, v := range raw {
				if uint32(v.status) > staticInfluenceStatusMax {
					return invalidArgument("invalid static influence status")
				}
				result[i], e = staticBatchInfluenceFromC(v)
				if e != nil {
					return e
				}
			}
			return nil
		})
	})
	return result, err
}

func (s *StaticPositionSolution) SatelliteInfluence() ([]StaticPositionSatelliteInfluence, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var w, r C.size_t
	var result []StaticPositionSatelliteInfluence
	err := s.handle.read(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			if e := statusErrorLocked(uint32(C.sidereon_static_position_solution_satellite_influence((*C.SidereonStaticPositionSolution)(p), nil, 0, &w, &r))); e != nil {
				return e
			}
			n, e := validateNativeQuery("static satellite influence", uint64(w), uint64(r))
			if e != nil {
				return e
			}
			mem, e := staticOutputMemory(n, unsafe.Sizeof(C.SidereonStaticPositionSatelliteInfluence{}), "static satellite influence")
			if e != nil {
				return e
			}
			if mem != nil {
				defer C.free(mem)
			}
			w, r = 0, 0
			if e = statusErrorLocked(uint32(C.sidereon_static_position_solution_satellite_influence((*C.SidereonStaticPositionSolution)(p), (*C.SidereonStaticPositionSatelliteInfluence)(mem), C.size_t(n), &w, &r))); e != nil {
				return e
			}
			rows, e := validateTwoPassCounts("static satellite influence", n, n, uint64(w), uint64(r))
			if e != nil {
				return e
			}
			raw := unsafe.Slice((*C.SidereonStaticPositionSatelliteInfluence)(mem), rows)
			result = make([]StaticPositionSatelliteInfluence, rows)
			for i, v := range raw {
				if uint32(v.status) > staticInfluenceStatusMax {
					return invalidArgument("invalid static influence status")
				}
				result[i], e = staticInfluenceFromC(v)
				if e != nil {
					return e
				}
			}
			return nil
		})
	})
	return result, err
}

func (s *StaticPositionSolution) StateCovarianceM2() ([]float64, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var w, r C.size_t
	var result []float64
	err := s.handle.read(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			if e := statusErrorLocked(uint32(C.sidereon_static_position_solution_state_covariance_m2((*C.SidereonStaticPositionSolution)(p), nil, 0, &w, &r))); e != nil {
				return e
			}
			n, e := validateNativeQuery("static state covariance", uint64(w), uint64(r))
			if e != nil {
				return e
			}
			if _, e = checkedNativeAllocationSize(n, unsafe.Sizeof(C.double(0))); e != nil {
				return e
			}
			if n == 0 {
				result = []float64{}
				return nil
			}
			mem, e := staticOutputMemory(n, unsafe.Sizeof(C.double(0)), "static state covariance")
			if e != nil {
				return e
			}
			defer C.free(mem)
			w, r = 0, 0
			if e = statusErrorLocked(uint32(C.sidereon_static_position_solution_state_covariance_m2((*C.SidereonStaticPositionSolution)(p), (*C.double)(mem), C.size_t(n), &w, &r))); e != nil {
				return e
			}
			rows, e := validateTwoPassCounts("static state covariance", n, n, uint64(w), uint64(r))
			if e != nil {
				return e
			}
			raw := unsafe.Slice((*C.double)(mem), rows)
			result = make([]float64, rows)
			for i := range raw {
				result[i] = float64(raw[i])
			}
			return nil
		})
	})
	return result, err
}

func (s *StaticPositionSolution) Geodetic() (Geodetic, bool, error) {
	if s == nil || s.handle == nil {
		return Geodetic{}, false, ErrClosed
	}
	var out C.SidereonGeodetic
	var present C.bool
	err := s.handle.read(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			return statusErrorLocked(uint32(C.sidereon_static_position_solution_geodetic((*C.SidereonStaticPositionSolution)(p), &out, &present)))
		})
	})
	return geodeticFromC(out), bool(present), err
}
func (s *StaticPositionSolution) Metadata() (StaticPositionMetadata, error) {
	if s == nil || s.handle == nil {
		return StaticPositionMetadata{}, ErrClosed
	}
	var out C.SidereonStaticPositionMetadata
	err := s.handle.read(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			return statusErrorLocked(uint32(C.sidereon_static_position_solution_metadata((*C.SidereonStaticPositionSolution)(p), &out)))
		})
	})
	if err != nil {
		return StaticPositionMetadata{}, err
	}
	return staticMetadataFromC(out)
}

func StaticReferenceStationRinexConfigInit() (StaticReferenceStationRinexConfigInput, error) {
	var out C.SidereonStaticReferenceStationRinexConfig
	if err := callStatus(func() uint32 { return uint32(C.sidereon_static_reference_station_rinex_config_init(&out)) }); err != nil {
		return StaticReferenceStationRinexConfigInput{}, err
	}
	carrier, err := rtkRinexStaticBaselineConfigFromC(out.carrier)
	if err != nil {
		return StaticReferenceStationRinexConfigInput{}, err
	}
	return StaticReferenceStationRinexConfigInput{ReferencePositionM: [3]float64{float64(out.reference_position_m[0]), float64(out.reference_position_m[1]), float64(out.reference_position_m[2])}, EnableCodeDGNSS: bool(out.enable_code_dgnss), EnableCarrierRTK: bool(out.enable_carrier_rtk), WithGeodetic: bool(out.with_geodetic), Carrier: carrier}, nil
}

func SolveStaticReferenceStationRinex(sp3 *SP3, reference, rover *RinexObs, value StaticReferenceStationRinexConfigInput) (*StaticReferenceStationSolution, error) {
	if sp3 == nil || sp3.handle == nil || reference == nil || rover == nil || reference.resource == nil || rover.resource == nil {
		return nil, ErrClosed
	}
	alloc := new(cRtkAlloc)
	defer alloc.close()
	carrier, err := copyRtkRinexStaticBaselineConfig(value.Carrier, alloc)
	if err != nil {
		return nil, err
	}
	mem, err := alloc.malloc(unsafe.Sizeof(C.SidereonStaticReferenceStationRinexConfig{}), "static reference station config")
	if err != nil {
		return nil, err
	}
	config := (*C.SidereonStaticReferenceStationRinexConfig)(mem)
	*config = C.SidereonStaticReferenceStationRinexConfig{reference_position_m: [3]C.double{C.double(value.ReferencePositionM[0]), C.double(value.ReferencePositionM[1]), C.double(value.ReferencePositionM[2])}, enable_code_dgnss: C.bool(value.EnableCodeDGNSS), enable_carrier_rtk: C.bool(value.EnableCarrierRTK), with_geodetic: C.bool(value.WithGeodetic), carrier: *carrier}
	var out *C.SidereonStaticReferenceStationSolution
	err = withScenarioInputs([]scenarioInput{{positioning: sp3.handle}, {resource: reference.resource}, {resource: rover.resource}}, func(pointers []unsafe.Pointer) error {
		return withCThreadError(func() error {
			status := C.sidereon_solve_static_reference_station_rinex((*C.SidereonSp3)(pointers[0]), (*C.SidereonRinexObs)(pointers[1]), (*C.SidereonRinexObs)(pointers[2]), config, &out)
			e := statusErrorLocked(uint32(status))
			if e != nil && out != nil {
				C.sidereon_static_reference_station_solution_free(out)
				out = nil
			}
			return e
		})
	})
	runtime.KeepAlive(sp3)
	runtime.KeepAlive(reference)
	runtime.KeepAlive(rover)
	if err != nil {
		return nil, err
	}
	return newStaticReferenceStationSolution(out)
}

func (s *StaticReferenceStationSolution) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return s.handle.close()
}
func (s *StaticReferenceStationSolution) fixed(call func(*C.SidereonStaticReferenceStationSolution, *C.double, C.size_t) C.enum_SidereonStatus) ([3]float64, error) {
	if s == nil || s.handle == nil {
		return [3]float64{}, ErrClosed
	}
	var out [3]C.double
	err := s.handle.read(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			return statusErrorLocked(uint32(call((*C.SidereonStaticReferenceStationSolution)(p), &out[0], 3)))
		})
	})
	return [3]float64{float64(out[0]), float64(out[1]), float64(out[2])}, err
}
func (s *StaticReferenceStationSolution) BaselineECEF() ([3]float64, error) {
	return s.fixed(func(solution *C.SidereonStaticReferenceStationSolution, out *C.double, n C.size_t) C.enum_SidereonStatus {
		return C.sidereon_static_reference_station_solution_baseline_ecef(solution, out, n)
	})
}
func (s *StaticReferenceStationSolution) PositionECEF() ([3]float64, error) {
	return s.fixed(func(solution *C.SidereonStaticReferenceStationSolution, out *C.double, n C.size_t) C.enum_SidereonStatus {
		return C.sidereon_static_reference_station_solution_position_ecef(solution, out, n)
	})
}
func (s *StaticReferenceStationSolution) CovarianceECEF() ([9]float64, error) {
	if s == nil || s.handle == nil {
		return [9]float64{}, ErrClosed
	}
	var out [9]C.double
	err := s.handle.read(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			return statusErrorLocked(uint32(C.sidereon_static_reference_station_solution_covariance_ecef((*C.SidereonStaticReferenceStationSolution)(p), &out[0], 9)))
		})
	})
	var result [9]float64
	for i := range result {
		result[i] = float64(out[i])
	}
	return result, err
}
func (s *StaticReferenceStationSolution) CovarianceENU() ([9]float64, error) {
	if s == nil || s.handle == nil {
		return [9]float64{}, ErrClosed
	}
	var out [9]C.double
	err := s.handle.read(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			return statusErrorLocked(uint32(C.sidereon_static_reference_station_solution_covariance_enu((*C.SidereonStaticReferenceStationSolution)(p), &out[0], 9)))
		})
	})
	var result [9]float64
	for i := range result {
		result[i] = float64(out[i])
	}
	return result, err
}
func (s *StaticReferenceStationSolution) Metadata() (StaticReferenceStationMetadata, error) {
	if s == nil || s.handle == nil {
		return StaticReferenceStationMetadata{}, ErrClosed
	}
	var out C.SidereonStaticReferenceStationMetadata
	err := s.handle.read(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			return statusErrorLocked(uint32(C.sidereon_static_reference_station_solution_metadata((*C.SidereonStaticReferenceStationSolution)(p), &out)))
		})
	})
	if err != nil {
		return StaticReferenceStationMetadata{}, err
	}
	return staticReferenceMetadataFromC(out)
}

func (s *StaticReferenceStationSolution) Diagnostics() ([]StaticReferenceEpochDiagnostic, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var w, r C.size_t
	var result []StaticReferenceEpochDiagnostic
	err := s.handle.read(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			if e := statusErrorLocked(uint32(C.sidereon_static_reference_station_solution_diagnostics((*C.SidereonStaticReferenceStationSolution)(p), nil, 0, &w, &r))); e != nil {
				return e
			}
			n, e := validateNativeQuery("static reference diagnostics", uint64(w), uint64(r))
			if e != nil {
				return e
			}
			mem, e := staticOutputMemory(n, unsafe.Sizeof(C.SidereonStaticReferenceEpochDiagnostic{}), "static reference diagnostics")
			if e != nil {
				return e
			}
			if mem != nil {
				defer C.free(mem)
			}
			w, r = 0, 0
			if e = statusErrorLocked(uint32(C.sidereon_static_reference_station_solution_diagnostics((*C.SidereonStaticReferenceStationSolution)(p), (*C.SidereonStaticReferenceEpochDiagnostic)(mem), C.size_t(n), &w, &r))); e != nil {
				return e
			}
			rows, e := validateTwoPassCounts("static reference diagnostics", n, n, uint64(w), uint64(r))
			if e != nil {
				return e
			}
			raw := unsafe.Slice((*C.SidereonStaticReferenceEpochDiagnostic)(mem), rows)
			result = make([]StaticReferenceEpochDiagnostic, rows)
			for i, v := range raw {
				if uint32(v.mode) > staticReferenceModeMax {
					return invalidArgument("invalid static reference mode")
				}
				result[i], e = staticDiagnosticFromC(v)
				if e != nil {
					return e
				}
			}
			return nil
		})
	})
	return result, err
}
func (s *StaticReferenceStationSolution) ModeReports() ([]StaticReferenceModeReport, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var w, r C.size_t
	var result []StaticReferenceModeReport
	err := s.handle.read(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			if e := statusErrorLocked(uint32(C.sidereon_static_reference_station_solution_mode_reports((*C.SidereonStaticReferenceStationSolution)(p), nil, 0, &w, &r))); e != nil {
				return e
			}
			n, e := validateNativeQuery("static reference mode reports", uint64(w), uint64(r))
			if e != nil {
				return e
			}
			mem, e := staticOutputMemory(n, unsafe.Sizeof(C.SidereonStaticReferenceModeReport{}), "static reference mode reports")
			if e != nil {
				return e
			}
			if mem != nil {
				defer C.free(mem)
			}
			w, r = 0, 0
			if e = statusErrorLocked(uint32(C.sidereon_static_reference_station_solution_mode_reports((*C.SidereonStaticReferenceStationSolution)(p), (*C.SidereonStaticReferenceModeReport)(mem), C.size_t(n), &w, &r))); e != nil {
				return e
			}
			rows, e := validateTwoPassCounts("static reference mode reports", n, n, uint64(w), uint64(r))
			if e != nil {
				return e
			}
			raw := unsafe.Slice((*C.SidereonStaticReferenceModeReport)(mem), rows)
			result = make([]StaticReferenceModeReport, rows)
			for i, v := range raw {
				if uint32(v.mode) > staticReferenceModeMax || uint32(v.status) > staticReferenceModeStatusMax {
					return invalidArgument("invalid static reference mode report enum")
				}
				result[i], e = staticModeReportFromC(v)
				if e != nil {
					return e
				}
			}
			return nil
		})
	})
	return result, err
}
