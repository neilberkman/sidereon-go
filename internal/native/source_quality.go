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
	"unsafe"
)

const (
	SourceSolveTOAValue                         = uint32(C.SIDEREON_SOURCE_SOLVE_MODE_TOA)
	SourceSolveTDOAValue                        = uint32(C.SIDEREON_SOURCE_SOLVE_MODE_TDOA)
	SourceLossLinearValue                       = uint32(C.SIDEREON_SOURCE_LOSS_LINEAR)
	SourceLossSoftL1Value                       = uint32(C.SIDEREON_SOURCE_LOSS_SOFT_L1)
	SourceLossHuberValue                        = uint32(C.SIDEREON_SOURCE_LOSS_HUBER)
	SourceLossCauchyValue                       = uint32(C.SIDEREON_SOURCE_LOSS_CAUCHY)
	SourceLossArctanValue                       = uint32(C.SIDEREON_SOURCE_LOSS_ARCTAN)
	BroadcastReasonPreciseUnavailableValue      = uint32(C.SIDEREON_BROADCAST_REASON_KIND_PRECISE_UNAVAILABLE)
	BroadcastReasonPreciseDegradedUnusableValue = uint32(C.SIDEREON_BROADCAST_REASON_KIND_PRECISE_DEGRADED_UNUSABLE)
	FixSourcePreciseValue                       = uint32(C.SIDEREON_FIX_SOURCE_KIND_PRECISE)
	FixSourceBroadcastValue                     = uint32(C.SIDEREON_FIX_SOURCE_KIND_BROADCAST)
	DegradationExactValue                       = uint32(C.SIDEREON_DEGRADATION_KIND_EXACT)
	DegradationNearestPriorValue                = uint32(C.SIDEREON_DEGRADATION_KIND_NEAREST_PRIOR)
	DegradationDiurnalShiftValue                = uint32(C.SIDEREON_DEGRADATION_KIND_DIURNAL_SHIFT)
)

type NativeSourceSensor struct {
	Dimension           int
	PositionM           [3]float64
	HasPropagationSpeed bool
	PropagationSpeedMS  float64
}

type NativeSourceInitialGuess struct {
	Dimension     int
	PositionM     [3]float64
	HasOriginTime bool
	OriginTimeS   float64
	ResidualRMSS  float64
}

type NativeSourceLocateOptions struct {
	Mode, ReferenceSensor int
	TimingSigmaS, FScaleS float64
	Loss                  uint32
	HasFTOL, HasXTOL      bool
	FTOL, XTOL            float64
	HasGTOL               bool
	GTOL                  float64
	HasMaxNFEV            bool
	MaxNFEV               int
}

type NativeSourceCovariance struct {
	Dimension, StateDimension  int
	State                      [16]float64
	PositionM2                 [9]float64
	HasOriginTimeS2            bool
	OriginTimeS2, TimingSigmaS float64
}

type NativeSourceCrlb struct {
	DOP        Dop
	Covariance NativeSourceCovariance
}

type NativeSourceSensorInfluence struct {
	SensorIndex            int
	ResidualS              float64
	HasLeaveOneOutResidual bool
	LeaveOneOutResidualS   float64
	HasPositionDelta       bool
	PositionDeltaM         float64
	HasOriginTimeDelta     bool
	OriginTimeDeltaS       float64
	LossWeight, Score      float64
}

type NativeSourceResidual struct {
	SensorIndex, ReferenceSensorIndex int
	HasReferenceSensor                bool
	ResidualS                         float64
}

type NativeSourceSolutionSummary struct {
	Dimension, ResidualCount, InfluenceCount int
	PositionM                                [3]float64
	HasOriginTime                            bool
	OriginTimeS                              float64
	HasCovariance                            bool
	GeometryQuality                          GeometryQuality
	InitialGuess                             NativeSourceInitialGuess
	Status                                   int32
	NFEV, NJEV                               int
	Cost, Optimality                         float64
}

type SourceSolution struct {
	_      noCopy
	handle *positioningHandle
}

type NativeStalenessMetadata struct {
	Kind                                                               uint32
	RequestedEpochJ2000S, SourceEpochJ2000S, StalenessS, StalenessDays float64
}

type SourcedSolution struct {
	_      noCopy
	handle *positioningHandle
}

func sourceMode(value uint32) (C.uint32_t, error) {
	if value != SourceSolveTOAValue && value != SourceSolveTDOAValue {
		return 0, invalidArgument("source solve mode is not defined by the C ABI")
	}
	return C.uint32_t(value), nil
}

func sourceLoss(value uint32) (C.uint32_t, error) {
	if value > SourceLossArctanValue {
		return 0, invalidArgument("source loss is not defined by the C ABI")
	}
	return C.uint32_t(value), nil
}

func cSourceLocateOptions(value NativeSourceLocateOptions) (C.SidereonSourceLocateOptions, error) {
	mode, err := sourceMode(uint32(value.Mode))
	if err != nil {
		return C.SidereonSourceLocateOptions{}, err
	}
	loss, err := sourceLoss(value.Loss)
	if err != nil {
		return C.SidereonSourceLocateOptions{}, err
	}
	maxNFEV, err := cSize(value.MaxNFEV, "source maximum function evaluations")
	if err != nil {
		return C.SidereonSourceLocateOptions{}, err
	}
	referenceSensor, err := cSize(value.ReferenceSensor, "source reference sensor")
	if err != nil {
		return C.SidereonSourceLocateOptions{}, err
	}
	return C.SidereonSourceLocateOptions{mode: mode, reference_sensor: referenceSensor, timing_sigma_s: C.double(value.TimingSigmaS), loss: loss, f_scale_s: C.double(value.FScaleS), has_ftol: C.bool(value.HasFTOL), ftol: C.double(value.FTOL), has_xtol: C.bool(value.HasXTOL), xtol: C.double(value.XTOL), has_gtol: C.bool(value.HasGTOL), gtol: C.double(value.GTOL), has_max_nfev: C.bool(value.HasMaxNFEV), max_nfev: maxNFEV}, nil
}

func cSourceSensors(values []NativeSourceSensor) (unsafe.Pointer, C.size_t, error) {
	count, err := cSize(len(values), "source sensor count")
	if err != nil {
		return nil, 0, err
	}
	if len(values) == 0 {
		return nil, count, nil
	}
	size, err := checkedNativeAllocationSize(len(values), unsafe.Sizeof(C.SidereonSourceSensor{}))
	if err != nil {
		return nil, 0, err
	}
	memory := C.malloc(C.size_t(size))
	if memory == nil {
		return nil, 0, errors.New("sidereon: unable to allocate native source sensors")
	}
	output := unsafe.Slice((*C.SidereonSourceSensor)(memory), len(values))
	for i, value := range values {
		if value.Dimension != 2 && value.Dimension != 3 {
			C.free(memory)
			return nil, 0, invalidArgument("source sensor dimension must be two or three")
		}
		output[i].dimension = C.size_t(value.Dimension)
		for j := range value.PositionM {
			output[i].position_m[j] = C.double(value.PositionM[j])
		}
		output[i].has_propagation_speed_m_s = C.bool(value.HasPropagationSpeed)
		output[i].propagation_speed_m_s = C.double(value.PropagationSpeedMS)
	}
	return memory, count, nil
}

func sourceInitialGuessFromC(value C.SidereonSourceInitialGuess) (NativeSourceInitialGuess, error) {
	dimension, err := sizeTToInt(value.dimension, "source initial-guess dimension")
	if err != nil || (dimension != 2 && dimension != 3) {
		if err == nil {
			err = invalidArgument("native source initial-guess dimension is invalid")
		}
		return NativeSourceInitialGuess{}, err
	}
	var position [3]float64
	for i := range position {
		position[i] = float64(value.position_m[i])
	}
	return NativeSourceInitialGuess{Dimension: dimension, PositionM: position, HasOriginTime: bool(value.has_origin_time_s), OriginTimeS: float64(value.origin_time_s), ResidualRMSS: float64(value.residual_rms_s)}, nil
}

func sourceInitialGuessCall(sensors []NativeSourceSensor, arrivals []float64, speed float64, mode uint32, reference int, closedForm bool) (NativeSourceInitialGuess, error) {
	if len(sensors) == 0 || len(sensors) != len(arrivals) {
		return NativeSourceInitialGuess{}, invalidArgument("source sensors and arrival times must be non-empty and have equal lengths")
	}
	sensorMemory, count, err := cSourceSensors(sensors)
	if err != nil {
		return NativeSourceInitialGuess{}, err
	}
	defer C.free(sensorMemory)
	arrivalMemory, _, err := cFloats(arrivals, "source arrival times")
	if err != nil {
		return NativeSourceInitialGuess{}, err
	}
	defer C.free(arrivalMemory)
	modeCode, err := sourceMode(mode)
	if err != nil {
		return NativeSourceInitialGuess{}, err
	}
	ref, err := cSize(reference, "source reference sensor")
	if err != nil {
		return NativeSourceInitialGuess{}, err
	}
	var output C.SidereonSourceInitialGuess
	err = withCThreadError(func() error {
		var status C.enum_SidereonStatus
		if closedForm {
			status = C.sidereon_closed_form_initial_guess((*C.SidereonSourceSensor)(sensorMemory), count, (*C.double)(arrivalMemory), C.double(speed), modeCode, ref, &output)
		} else {
			status = C.sidereon_chan_ho_initial_guess((*C.SidereonSourceSensor)(sensorMemory), count, (*C.double)(arrivalMemory), C.double(speed), modeCode, ref, &output)
		}
		return statusErrorLocked(uint32(status))
	})
	if err != nil {
		return NativeSourceInitialGuess{}, err
	}
	return sourceInitialGuessFromC(output)
}

func ChanHOInitialGuess(sensors []NativeSourceSensor, arrivals []float64, speed float64, mode uint32, reference int) (NativeSourceInitialGuess, error) {
	return sourceInitialGuessCall(sensors, arrivals, speed, mode, reference, false)
}

func ClosedFormInitialGuess(sensors []NativeSourceSensor, arrivals []float64, speed float64, mode uint32, reference int) (NativeSourceInitialGuess, error) {
	return sourceInitialGuessCall(sensors, arrivals, speed, mode, reference, true)
}

func SourceLocateOptionsInit() (NativeSourceLocateOptions, error) {
	var output C.SidereonSourceLocateOptions
	err := callStatus(func() uint32 { return uint32(C.sidereon_source_locate_options_init(&output)) })
	if err != nil {
		return NativeSourceLocateOptions{}, err
	}
	if _, err := sourceMode(uint32(output.mode)); err != nil {
		return NativeSourceLocateOptions{}, err
	}
	if _, err := sourceLoss(uint32(output.loss)); err != nil {
		return NativeSourceLocateOptions{}, err
	}
	referenceSensor, err := sizeTToInt(output.reference_sensor, "source reference sensor")
	if err != nil {
		return NativeSourceLocateOptions{}, err
	}
	maxNFEV, err := sizeTToInt(output.max_nfev, "source maximum function evaluations")
	if err != nil {
		return NativeSourceLocateOptions{}, err
	}
	return NativeSourceLocateOptions{Mode: int(output.mode), ReferenceSensor: referenceSensor, TimingSigmaS: float64(output.timing_sigma_s), Loss: uint32(output.loss), FScaleS: float64(output.f_scale_s), HasFTOL: bool(output.has_ftol), FTOL: float64(output.ftol), HasXTOL: bool(output.has_xtol), XTOL: float64(output.xtol), HasGTOL: bool(output.has_gtol), GTOL: float64(output.gtol), HasMaxNFEV: bool(output.has_max_nfev), MaxNFEV: maxNFEV}, nil
}

func newSourceSolution(pointer *C.SidereonSourceSolution) (*SourceSolution, error) {
	if pointer == nil {
		return nil, missingNativeHandle("source localization")
	}
	handle := newPositioningHandle(unsafe.Pointer(pointer), func(value unsafe.Pointer) { C.sidereon_source_solution_free((*C.SidereonSourceSolution)(value)) })
	return &SourceSolution{handle: handle}, nil
}

func locateSource(sensors []NativeSourceSensor, arrivals []float64, speed float64, options *NativeSourceLocateOptions, includeInfluence bool, withOptions bool) (*SourceSolution, error) {
	if len(sensors) == 0 || len(sensors) != len(arrivals) {
		return nil, invalidArgument("source sensors and arrival times must be non-empty and have equal lengths")
	}
	sensorMemory, count, err := cSourceSensors(sensors)
	if err != nil {
		return nil, err
	}
	defer C.free(sensorMemory)
	arrivalMemory, _, err := cFloats(arrivals, "source arrival times")
	if err != nil {
		return nil, err
	}
	defer C.free(arrivalMemory)
	var cOptions C.SidereonSourceLocateOptions
	var optionsPointer *C.SidereonSourceLocateOptions
	if options != nil {
		cOptions, err = cSourceLocateOptions(*options)
		if err != nil {
			return nil, err
		}
		optionsPointer = &cOptions
	} else if withOptions {
		optionsPointer = nil
	}
	var output *C.SidereonSourceSolution
	var operationErr error
	withCThread(func() {
		var status C.enum_SidereonStatus
		if withOptions {
			status = C.sidereon_locate_source_with((*C.SidereonSourceSensor)(sensorMemory), count, (*C.double)(arrivalMemory), C.double(speed), optionsPointer, C.bool(includeInfluence), &output)
		} else {
			status = C.sidereon_locate_source((*C.SidereonSourceSensor)(sensorMemory), count, (*C.double)(arrivalMemory), C.double(speed), optionsPointer, &output)
		}
		operationErr = statusErrorLocked(uint32(status))
		if operationErr != nil && output != nil {
			C.sidereon_source_solution_free(output)
			output = nil
		}
	})
	if operationErr != nil {
		return nil, operationErr
	}
	return newSourceSolution(output)
}

func LocateSource(sensors []NativeSourceSensor, arrivals []float64, speed float64, options *NativeSourceLocateOptions) (*SourceSolution, error) {
	return locateSource(sensors, arrivals, speed, options, true, false)
}

func LocateSourceWith(sensors []NativeSourceSensor, arrivals []float64, speed float64, options *NativeSourceLocateOptions, includeInfluence bool) (*SourceSolution, error) {
	return locateSource(sensors, arrivals, speed, options, includeInfluence, true)
}

func sourceCovarianceFromC(value C.SidereonSourceCovariance) (NativeSourceCovariance, error) {
	dimension, err := sizeTToInt(value.dimension, "source covariance dimension")
	if err != nil {
		return NativeSourceCovariance{}, err
	}
	stateDimension, err := sizeTToInt(value.state_dimension, "source covariance state dimension")
	if err != nil {
		return NativeSourceCovariance{}, err
	}
	if (dimension != 2 && dimension != 3) || stateDimension == 0 || stateDimension > 4 {
		return NativeSourceCovariance{}, invalidArgument("native source covariance shape is invalid")
	}
	var state [16]float64
	var position [9]float64
	for i := range state {
		state[i] = float64(value.state[i])
	}
	for i := range position {
		position[i] = float64(value.position_m2[i])
	}
	return NativeSourceCovariance{Dimension: dimension, StateDimension: stateDimension, State: state, PositionM2: position, HasOriginTimeS2: bool(value.has_origin_time_s2), OriginTimeS2: float64(value.origin_time_s2), TimingSigmaS: float64(value.timing_sigma_s)}, nil
}

func sourceCovarianceResult(present bool, decode func() (NativeSourceCovariance, error)) (NativeSourceCovariance, bool, error) {
	if !present {
		return NativeSourceCovariance{}, false, nil
	}
	decoded, err := decode()
	return decoded, true, err
}

func sourceInfluenceFromC(value C.SidereonSourceSensorInfluence) (NativeSourceSensorInfluence, error) {
	index, err := sizeTToInt(value.sensor_index, "source influence sensor index")
	if err != nil {
		return NativeSourceSensorInfluence{}, err
	}
	return NativeSourceSensorInfluence{SensorIndex: index, ResidualS: float64(value.residual_s), HasLeaveOneOutResidual: bool(value.has_leave_one_out_residual_s), LeaveOneOutResidualS: float64(value.leave_one_out_residual_s), HasPositionDelta: bool(value.has_position_delta_m), PositionDeltaM: float64(value.position_delta_m), HasOriginTimeDelta: bool(value.has_origin_time_delta_s), OriginTimeDeltaS: float64(value.origin_time_delta_s), LossWeight: float64(value.loss_weight), Score: float64(value.score)}, nil
}

func sourceResidualFromC(value C.SidereonSourceResidual) (NativeSourceResidual, error) {
	index, err := sizeTToInt(value.sensor_index, "source residual sensor index")
	if err != nil {
		return NativeSourceResidual{}, err
	}
	ref, err := sizeTToInt(value.reference_sensor_index, "source residual reference index")
	if err != nil {
		return NativeSourceResidual{}, err
	}
	return NativeSourceResidual{SensorIndex: index, ReferenceSensorIndex: ref, HasReferenceSensor: bool(value.has_reference_sensor_index), ResidualS: float64(value.residual_s)}, nil
}

func sourceSummaryFromC(value C.SidereonSourceSolutionSummary) (NativeSourceSolutionSummary, error) {
	dimension, err := sizeTToInt(value.dimension, "source summary dimension")
	if err != nil {
		return NativeSourceSolutionSummary{}, err
	}
	residualCount, err := sizeTToInt(value.residual_count, "source summary residual count")
	if err != nil {
		return NativeSourceSolutionSummary{}, err
	}
	influenceCount, err := sizeTToInt(value.influence_count, "source summary influence count")
	if err != nil {
		return NativeSourceSolutionSummary{}, err
	}
	rank, err := sizeTToInt(value.geometry_quality.rank, "source summary geometry rank")
	if err != nil {
		return NativeSourceSolutionSummary{}, err
	}
	if dimension != 2 && dimension != 3 {
		return NativeSourceSolutionSummary{}, invalidArgument("native source summary dimension is invalid")
	}
	if err := validateObservabilityTier(uint32(value.geometry_quality.tier)); err != nil {
		return NativeSourceSolutionSummary{}, err
	}
	guess, err := sourceInitialGuessFromC(value.initial_guess)
	if err != nil {
		return NativeSourceSolutionSummary{}, err
	}
	var position [3]float64
	for i := range position {
		position[i] = float64(value.position_m[i])
	}
	geometry := GeometryQuality{Tier: uint32(value.geometry_quality.tier), Redundancy: int32(value.geometry_quality.redundancy), Rank: rank, ConditionNumber: float64(value.geometry_quality.condition_number), GDOP: float64(value.geometry_quality.gdop), RAIMCheckable: bool(value.geometry_quality.raim_checkable), CovarianceValidated: bool(value.geometry_quality.covariance_validated)}
	nfev, err := sizeTToInt(value.nfev, "source summary nfev")
	if err != nil {
		return NativeSourceSolutionSummary{}, err
	}
	njev, err := sizeTToInt(value.njev, "source summary njev")
	if err != nil {
		return NativeSourceSolutionSummary{}, err
	}
	return NativeSourceSolutionSummary{Dimension: dimension, ResidualCount: residualCount, InfluenceCount: influenceCount, PositionM: position, HasOriginTime: bool(value.has_origin_time_s), OriginTimeS: float64(value.origin_time_s), HasCovariance: bool(value.has_covariance), GeometryQuality: geometry, InitialGuess: guess, Status: int32(value.status), NFEV: nfev, NJEV: njev, Cost: float64(value.cost), Optimality: float64(value.optimality)}, nil
}

func (s *SourceSolution) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return s.handle.close()
}

func (s *SourceSolution) Covariance() (NativeSourceCovariance, bool, error) {
	if s == nil || s.handle == nil {
		return NativeSourceCovariance{}, false, ErrClosed
	}
	var output C.SidereonSourceCovariance
	var present C.bool
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			return statusErrorLocked(uint32(C.sidereon_source_solution_covariance((*C.SidereonSourceSolution)(pointer), &output, &present)))
		})
	})
	if err != nil {
		return NativeSourceCovariance{}, false, err
	}
	return sourceCovarianceResult(bool(present), func() (NativeSourceCovariance, error) { return sourceCovarianceFromC(output) })
}

func (s *SourceSolution) Influences() ([]NativeSourceSensorInfluence, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var result []NativeSourceSensorInfluence
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			var written, required C.size_t
			status := C.sidereon_source_solution_influences((*C.SidereonSourceSolution)(pointer), nil, 0, &written, &required)
			if err := statusErrorLocked(uint32(status)); err != nil {
				return err
			}
			n, err := sizeTToInt(required, "source influence count")
			if err != nil {
				return err
			}
			if _, err = writtenToInt(written, 0, "source influence query written count"); err != nil {
				return err
			}
			if _, err = checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonSourceSensorInfluence{})); err != nil {
				return err
			}
			buffer := make([]C.SidereonSourceSensorInfluence, n)
			var output *C.SidereonSourceSensorInfluence
			if n > 0 {
				output = &buffer[0]
			}
			written, required = 0, 0
			status = C.sidereon_source_solution_influences((*C.SidereonSourceSolution)(pointer), output, C.size_t(n), &written, &required)
			if err := statusErrorLocked(uint32(status)); err != nil {
				return err
			}
			actual, err := validateTwoPassCounts("source influences", n, n, uint64(written), uint64(required))
			if err != nil {
				return err
			}
			result = make([]NativeSourceSensorInfluence, actual)
			for i := range result {
				result[i], err = sourceInfluenceFromC(buffer[i])
				if err != nil {
					return err
				}
			}
			return nil
		})
	})
	return result, err
}

func (s *SourceSolution) Residuals() ([]NativeSourceResidual, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var result []NativeSourceResidual
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			var written, required C.size_t
			status := C.sidereon_source_solution_residuals((*C.SidereonSourceSolution)(pointer), nil, 0, &written, &required)
			if err := statusErrorLocked(uint32(status)); err != nil {
				return err
			}
			n, err := sizeTToInt(required, "source residual count")
			if err != nil {
				return err
			}
			if _, err = writtenToInt(written, 0, "source residual query written count"); err != nil {
				return err
			}
			if _, err = checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonSourceResidual{})); err != nil {
				return err
			}
			buffer := make([]C.SidereonSourceResidual, n)
			var output *C.SidereonSourceResidual
			if n > 0 {
				output = &buffer[0]
			}
			written, required = 0, 0
			status = C.sidereon_source_solution_residuals((*C.SidereonSourceSolution)(pointer), output, C.size_t(n), &written, &required)
			if err := statusErrorLocked(uint32(status)); err != nil {
				return err
			}
			actual, err := validateTwoPassCounts("source residuals", n, n, uint64(written), uint64(required))
			if err != nil {
				return err
			}
			result = make([]NativeSourceResidual, actual)
			for i := range result {
				result[i], err = sourceResidualFromC(buffer[i])
				if err != nil {
					return err
				}
			}
			return nil
		})
	})
	return result, err
}

func (s *SourceSolution) Summary() (NativeSourceSolutionSummary, error) {
	if s == nil || s.handle == nil {
		return NativeSourceSolutionSummary{}, ErrClosed
	}
	var output C.SidereonSourceSolutionSummary
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			return statusErrorLocked(uint32(C.sidereon_source_solution_summary((*C.SidereonSourceSolution)(pointer), &output)))
		})
	})
	if err != nil {
		return NativeSourceSolutionSummary{}, err
	}
	return sourceSummaryFromC(output)
}

func validateDegradation(value uint32) error {
	if value > DegradationDiurnalShiftValue {
		return invalidArgument("degradation kind is not defined by the C ABI")
	}
	return nil
}
func validateSelection(value uint32) error {
	if value > 11 {
		return invalidArgument("selection status is not defined by the C ABI")
	}
	return nil
}
func validateFixSource(value uint32) error {
	if value > FixSourceBroadcastValue {
		return invalidArgument("fix source kind is not defined by the C ABI")
	}
	return nil
}
func validateBroadcastReason(value uint32) error {
	if value > BroadcastReasonPreciseDegradedUnusableValue {
		return invalidArgument("broadcast reason is not defined by the C ABI")
	}
	return nil
}

func releaseSourcedSolution(pointer unsafe.Pointer) {
	C.sidereon_sourced_solution_free((*C.SidereonSourcedSolution)(pointer))
}
func (s *SourcedSolution) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return s.handle.close()
}

func (s *SourcedSolution) BroadcastReason() (uint32, uint32, NativeStalenessMetadata, bool, error) {
	if s == nil || s.handle == nil {
		return 0, 0, NativeStalenessMetadata{}, false, ErrClosed
	}
	var reason C.enum_SidereonBroadcastReasonKind
	var selection C.enum_SidereonSelectionStatus
	var metadata C.SidereonStalenessMetadata
	var present C.bool
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			return statusErrorLocked(uint32(C.sidereon_sourced_solution_broadcast_reason((*C.SidereonSourcedSolution)(pointer), &reason, &selection, &metadata, &present)))
		})
	})
	if err != nil {
		return 0, 0, NativeStalenessMetadata{}, false, err
	}
	if err := validateBroadcastReason(uint32(reason)); err != nil {
		return 0, 0, NativeStalenessMetadata{}, false, err
	}
	if err := validateSelection(uint32(selection)); err != nil {
		return 0, 0, NativeStalenessMetadata{}, false, err
	}
	if bool(present) {
		if err := validateDegradation(uint32(metadata.kind)); err != nil {
			return 0, 0, NativeStalenessMetadata{}, false, err
		}
	} else {
		return uint32(reason), uint32(selection), NativeStalenessMetadata{}, false, nil
	}
	return uint32(reason), uint32(selection), NativeStalenessMetadata{Kind: uint32(metadata.kind), RequestedEpochJ2000S: float64(metadata.requested_epoch_j2000_s), SourceEpochJ2000S: float64(metadata.source_epoch_j2000_s), StalenessS: float64(metadata.staleness_s), StalenessDays: float64(metadata.staleness_days)}, bool(present), nil
}

func (s *SourcedSolution) IsPreciseExact() (bool, error) {
	if s == nil || s.handle == nil {
		return false, ErrClosed
	}
	var output C.bool
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			return statusErrorLocked(uint32(C.sidereon_sourced_solution_is_precise_exact((*C.SidereonSourcedSolution)(pointer), &output)))
		})
	})
	return bool(output), err
}

func (s *SourcedSolution) SourceKind() (uint32, error) {
	if s == nil || s.handle == nil {
		return 0, ErrClosed
	}
	var output C.enum_SidereonFixSourceKind
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			return statusErrorLocked(uint32(C.sidereon_sourced_solution_source_kind((*C.SidereonSourcedSolution)(pointer), &output)))
		})
	})
	if err != nil {
		return 0, err
	}
	if err := validateFixSource(uint32(output)); err != nil {
		return 0, err
	}
	return uint32(output), nil
}

func (s *SourcedSolution) Staleness() (NativeStalenessMetadata, bool, error) {
	if s == nil || s.handle == nil {
		return NativeStalenessMetadata{}, false, ErrClosed
	}
	var output C.SidereonStalenessMetadata
	var present C.bool
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			return statusErrorLocked(uint32(C.sidereon_sourced_solution_staleness((*C.SidereonSourcedSolution)(pointer), &output, &present)))
		})
	})
	if err != nil {
		return NativeStalenessMetadata{}, false, err
	}
	if bool(present) {
		if err := validateDegradation(uint32(output.kind)); err != nil {
			return NativeStalenessMetadata{}, false, err
		}
	} else {
		return NativeStalenessMetadata{}, false, nil
	}
	return NativeStalenessMetadata{Kind: uint32(output.kind), RequestedEpochJ2000S: float64(output.requested_epoch_j2000_s), SourceEpochJ2000S: float64(output.source_epoch_j2000_s), StalenessS: float64(output.staleness_s), StalenessDays: float64(output.staleness_days)}, bool(present), nil
}

func (s *SourcedSolution) Solution() (SPPSolution, error) {
	if s == nil || s.handle == nil {
		return SPPSolution{}, ErrClosed
	}
	var result SPPSolution
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			var output *C.SidereonSppSolution
			status := C.sidereon_sourced_solution_solution((*C.SidereonSourcedSolution)(pointer), &output)
			if err := statusErrorLocked(uint32(status)); err != nil {
				if output != nil {
					C.sidereon_spp_solution_free(output)
				}
				return err
			}
			if output == nil {
				return missingNativeHandle("sourced solution")
			}
			defer C.sidereon_spp_solution_free(output)
			var err error
			result, err = readSPPSolutionLocked(output)
			return err
		})
	})
	return result, err
}

func cSourcePosition(value []float64, label string) (unsafe.Pointer, C.size_t, error) {
	return cFloats(value, label)
}

func SourceCRLB(sensors []NativeSourceSensor, sourcePosition []float64, speed, timingSigma float64) (NativeSourceCrlb, error) {
	return sourceCRLBOrDOP(sensors, sourcePosition, speed, timingSigma, true)
}
func SourceDOP(sensors []NativeSourceSensor, sourcePosition []float64, speed float64) (Dop, error) {
	value, err := sourceCRLBOrDOP(sensors, sourcePosition, speed, 0, false)
	return value.DOP, err
}

func sourceCRLBOrDOP(sensors []NativeSourceSensor, sourcePosition []float64, speed, timingSigma float64, crlb bool) (NativeSourceCrlb, error) {
	if len(sensors) == 0 || (len(sourcePosition) != 2 && len(sourcePosition) != 3) {
		return NativeSourceCrlb{}, invalidArgument("source geometry dimensions are invalid")
	}
	sensorMemory, count, err := cSourceSensors(sensors)
	if err != nil {
		return NativeSourceCrlb{}, err
	}
	defer C.free(sensorMemory)
	positionMemory, positionCount, err := cSourcePosition(sourcePosition, "source position")
	if err != nil {
		return NativeSourceCrlb{}, err
	}
	defer C.free(positionMemory)
	var output NativeSourceCrlb
	var operationErr error
	withCThread(func() {
		if crlb {
			var value C.SidereonSourceCrlb
			operationErr = statusErrorLocked(uint32(C.sidereon_source_crlb((*C.SidereonSourceSensor)(sensorMemory), count, (*C.double)(positionMemory), positionCount, C.double(speed), C.double(timingSigma), &value)))
			if operationErr == nil {
				output.DOP = Dop{GDOP: float64(value.dop.gdop), PDOP: float64(value.dop.pdop), HDOP: float64(value.dop.hdop), VDOP: float64(value.dop.vdop), TDOP: float64(value.dop.tdop)}
				output.Covariance, operationErr = sourceCovarianceFromC(value.covariance)
			}
		} else {
			var value C.SidereonDop
			operationErr = statusErrorLocked(uint32(C.sidereon_source_dop((*C.SidereonSourceSensor)(sensorMemory), count, (*C.double)(positionMemory), positionCount, C.double(speed), &value)))
			output.DOP = Dop{GDOP: float64(value.gdop), PDOP: float64(value.pdop), HDOP: float64(value.hdop), VDOP: float64(value.vdop), TDOP: float64(value.tdop)}
		}
	})
	return output, operationErr
}

type NativeCnavParameters struct {
	Present                    bool
	ADOTMS, DeltaN0DotRadS2    float64
	TopWeek                    uint32
	TopTOWS                    float64
	URAEDIndex, URANED0Index   int8
	URANED1Index, URANED2Index uint8
	TransmissionTimeSOW        float64
	HasFlags                   bool
	Flags                      uint32
}

func cCnavParameters(value NativeCnavParameters) C.SidereonCnavParameters {
	return C.SidereonCnavParameters{present: C.bool(value.Present), adot_m_s: C.double(value.ADOTMS), delta_n0_dot_rad_s2: C.double(value.DeltaN0DotRadS2), top_week: C.uint32_t(value.TopWeek), top_tow_s: C.double(value.TopTOWS), ura_ed_index: C.int8_t(value.URAEDIndex), ura_ned0_index: C.int8_t(value.URANED0Index), ura_ned1_index: C.uint8_t(value.URANED1Index), ura_ned2_index: C.uint8_t(value.URANED2Index), transmission_time_sow: C.double(value.TransmissionTimeSOW), has_flags: C.bool(value.HasFlags), flags: C.uint32_t(value.Flags)}
}
func CNAVURANEDM(params NativeCnavParameters, queryWeek uint32, queryTOW float64) (float64, bool, error) {
	input := cCnavParameters(params)
	var output C.double
	var present C.bool
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_cnav_ura_ned_m(&input, C.uint32_t(queryWeek), C.double(queryTOW), &output, &present))
	})
	return float64(output), bool(present), err
}
func CNAVURANominalM(index int8) (float64, bool, error) {
	var output C.double
	var present C.bool
	err := callStatus(func() uint32 { return uint32(C.sidereon_cnav_ura_nominal_m(C.int8_t(index), &output, &present)) })
	return float64(output), bool(present), err
}

func cPseudorangeOptions(value NativePseudorangeVarianceOptions) (C.SidereonPseudorangeVarianceOptions, error) {
	if value.Model != PseudorangeVarianceElevationValue && value.Model != PseudorangeVarianceElevationCN0Value {
		return C.SidereonPseudorangeVarianceOptions{}, invalidArgument("pseudorange variance model is not defined by the C ABI")
	}
	return C.SidereonPseudorangeVarianceOptions{a_m: C.double(value.AM), b_m: C.double(value.BM), model: C.uint32_t(value.Model), has_cn0: C.bool(value.HasCN0), cn0_dbhz: C.double(value.CN0DBHz), cn0_scale_m2: C.double(value.CN0ScaleM2)}, nil
}
func PseudorangeVarianceOptionsInit() (NativePseudorangeVarianceOptions, error) {
	var output C.SidereonPseudorangeVarianceOptions
	err := callStatus(func() uint32 { return uint32(C.sidereon_pseudorange_variance_options_init(&output)) })
	if err != nil {
		return NativePseudorangeVarianceOptions{}, err
	}
	if output.model != C.SIDEREON_PSEUDORANGE_VARIANCE_MODEL_ELEVATION && output.model != C.SIDEREON_PSEUDORANGE_VARIANCE_MODEL_ELEVATION_CN0 {
		return NativePseudorangeVarianceOptions{}, invalidArgument("native pseudorange variance model is not defined by the C ABI")
	}
	return NativePseudorangeVarianceOptions{AM: float64(output.a_m), BM: float64(output.b_m), Model: uint32(output.model), HasCN0: bool(output.has_cn0), CN0DBHz: float64(output.cn0_dbhz), CN0ScaleM2: float64(output.cn0_scale_m2)}, nil
}
func PseudorangeVariance(elevation float64, options NativePseudorangeVarianceOptions) (float64, error) {
	input, err := cPseudorangeOptions(options)
	if err != nil {
		return 0, err
	}
	var output C.double
	err = callStatus(func() uint32 { return uint32(C.sidereon_pseudorange_variance(C.double(elevation), &input, &output)) })
	return float64(output), err
}

type NativeSolutionValidationOptions struct {
	HasMaxPDOP                                                                  bool
	MaxPDOP, MinPlausibleRadiusM, MaxPlausibleRadiusM, MaxConvergedResidualRMSM float64
}

func cSolutionValidationOptions(value NativeSolutionValidationOptions) C.SidereonSolutionValidationOptions {
	return C.SidereonSolutionValidationOptions{has_max_pdop: C.bool(value.HasMaxPDOP), max_pdop: C.double(value.MaxPDOP), min_plausible_radius_m: C.double(value.MinPlausibleRadiusM), max_plausible_radius_m: C.double(value.MaxPlausibleRadiusM), max_converged_residual_rms_m: C.double(value.MaxConvergedResidualRMSM)}
}

func validateReceiverSolutionLocked(solution *C.SidereonSppSolution, options NativeSolutionValidationOptions) error {
	if solution == nil {
		return ErrClosed
	}
	input := cSolutionValidationOptions(options)
	return statusErrorLocked(uint32(C.sidereon_validate_receiver_solution(solution, &input)))
}

func SolutionValidationOptionsInit() (NativeSolutionValidationOptions, error) {
	var output C.SidereonSolutionValidationOptions
	err := callStatus(func() uint32 { return uint32(C.sidereon_solution_validation_options_init(&output)) })
	return NativeSolutionValidationOptions{HasMaxPDOP: bool(output.has_max_pdop), MaxPDOP: float64(output.max_pdop), MinPlausibleRadiusM: float64(output.min_plausible_radius_m), MaxPlausibleRadiusM: float64(output.max_plausible_radius_m), MaxConvergedResidualRMSM: float64(output.max_converged_residual_rms_m)}, err
}

// ValidateReceiverSolution is intentionally kept at the native boundary: the
// C ABI requires a live SidereonSppSolution handle. Callers with such a handle
// can use this internal helper from the owning solve path.
func ValidateReceiverSolution(solution *C.SidereonSppSolution, options NativeSolutionValidationOptions) error {
	return withCThreadError(func() error { return validateReceiverSolutionLocked(solution, options) })
}
