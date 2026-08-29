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
	"sync"
	"unsafe"
)

const (
	PseudorangeVarianceElevationValue    = uint32(C.SIDEREON_PSEUDORANGE_VARIANCE_MODEL_ELEVATION)
	PseudorangeVarianceElevationCN0Value = uint32(C.SIDEREON_PSEUDORANGE_VARIANCE_MODEL_ELEVATION_CN0)
	ReducedOrbitCircularSecularValue     = uint32(C.SIDEREON_REDUCED_ORBIT_MODEL_CIRCULAR_SECULAR)
	ReducedOrbitEccentricSecularValue    = uint32(C.SIDEREON_REDUCED_ORBIT_MODEL_ECCENTRIC_SECULAR)
	ReducedOrbitFrameGCRSValue           = uint32(C.SIDEREON_REDUCED_ORBIT_FRAME_GCRS)
	ReducedOrbitFrameECEFValue           = uint32(C.SIDEREON_REDUCED_ORBIT_FRAME_ECEF)
	FusionFilterEKFValue                 = uint32(C.SIDEREON_FUSION_FILTER_KIND_EKF)
	FusionFilterUKFValue                 = uint32(C.SIDEREON_FUSION_FILTER_KIND_UKF)
	FusionLayoutFifteenValue             = uint32(C.SIDEREON_FUSION_ERROR_STATE_LAYOUT_FIFTEEN)
	FusionLayoutTwentyOneValue           = uint32(C.SIDEREON_FUSION_ERROR_STATE_LAYOUT_TWENTY_ONE)
	FusionImuGradeMEMSValue              = uint32(C.SIDEREON_FUSION_IMU_GRADE_MEMS)
	FusionImuGradeTacticalValue          = uint32(C.SIDEREON_FUSION_IMU_GRADE_TACTICAL)
	FusionImuGradeNavigationValue        = uint32(C.SIDEREON_FUSION_IMU_GRADE_NAVIGATION)
	FusionSampleRateValue                = uint32(C.SIDEREON_FUSION_IMU_SAMPLE_KIND_RATE)
	FusionSampleIncrementValue           = uint32(C.SIDEREON_FUSION_IMU_SAMPLE_KIND_INCREMENT)
	FusionFixSingleValue                 = uint32(C.SIDEREON_FUSION_GNSS_FIX_STATUS_SINGLE)
	FusionFixFloatValue                  = uint32(C.SIDEREON_FUSION_GNSS_FIX_STATUS_FLOAT)
	FusionFixFixedValue                  = uint32(C.SIDEREON_FUSION_GNSS_FIX_STATUS_FIXED)
	TrlsLinearValue                      = uint32(C.SIDEREON_TRLS_KIND_LINEAR)
	TrlsPolynomialValue                  = uint32(C.SIDEREON_TRLS_KIND_POLYNOMIAL)
	TrlsExponentialValue                 = uint32(C.SIDEREON_TRLS_KIND_EXPONENTIAL)
	TrlsLossLinearValue                  = uint32(C.SIDEREON_TRLS_LOSS_LINEAR)
	TrlsLossSoftL1Value                  = uint32(C.SIDEREON_TRLS_LOSS_SOFT_L1)
	TrlsLossHuberValue                   = uint32(C.SIDEREON_TRLS_LOSS_HUBER)
	TrlsLossCauchyValue                  = uint32(C.SIDEREON_TRLS_LOSS_CAUCHY)
	TrlsLossArctanValue                  = uint32(C.SIDEREON_TRLS_LOSS_ARCTAN)
	TrlsXScaleUnitValue                  = uint32(C.SIDEREON_TRLS_X_SCALE_UNIT)
	TrlsXScaleJacValue                   = uint32(C.SIDEREON_TRLS_X_SCALE_JAC)
	TrlsXScaleValuesValue                = uint32(C.SIDEREON_TRLS_X_SCALE_VALUES)
	TrlsBackendNativeValue               = uint32(C.SIDEREON_TRLS_BACKEND_NATIVE)
	TrlsBackendHostLAPACKValue           = uint32(C.SIDEREON_TRLS_BACKEND_HOST_LAPACK)
	TrackFrameECEFValue                  = uint32(C.SIDEREON_TRACK_COORDINATE_FRAME_ECEF)
	TrackFrameENUValue                   = uint32(C.SIDEREON_TRACK_COORDINATE_FRAME_ENU)
	TrackFrameCallerCartesianValue       = uint32(C.SIDEREON_TRACK_COORDINATE_FRAME_CALLER_DEFINED_CARTESIAN)
	PropagationForceTwoBodyValue         = uint32(C.SIDEREON_PROPAGATION_FORCE_MODEL_TWO_BODY)
	PropagationForceTwoBodyJ2Value       = uint32(C.SIDEREON_PROPAGATION_FORCE_MODEL_TWO_BODY_J2)
	PropagationForceCompositeValue       = uint32(C.SIDEREON_PROPAGATION_FORCE_MODEL_COMPOSITE)
	PropagationForceEarthPhaseAValue     = uint32(C.SIDEREON_PROPAGATION_FORCE_MODEL_EARTH_PHASE_A)
	PropagationForceEarthPhaseBValue     = uint32(C.SIDEREON_PROPAGATION_FORCE_MODEL_EARTH_PHASE_B)
)

// surfaceHandle is the common lifetime boundary for the owning C handles in
// this file. Read calls share the handle; mutable calls take the exclusive
// lock. The C documentation permits concurrent reads, but does not permit a
// read or mutation to overlap release.
type surfaceHandle struct {
	mu       sync.RWMutex
	resource *resource
	cleanup  runtime.Cleanup
}

func newSurfaceHandle(pointer unsafe.Pointer, release func(unsafe.Pointer)) (*surfaceHandle, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	handle := &surfaceHandle{resource: &resource{ptr: pointer, release: release}}
	handle.cleanup = runtime.AddCleanup(handle, cleanupResource, handle.resource)
	return handle, nil
}

func (h *surfaceHandle) read(fn func(unsafe.Pointer) error) error {
	if h == nil {
		return ErrClosed
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.resource == nil {
		return ErrClosed
	}
	return h.resource.with(fn)
}

func (h *surfaceHandle) write(fn func(unsafe.Pointer) error) error {
	if h == nil {
		return ErrClosed
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.resource == nil {
		return ErrClosed
	}
	return h.resource.with(func(pointer unsafe.Pointer) error { return fn(pointer) })
}

func (h *surfaceHandle) close() error {
	if h == nil {
		return nil
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.resource == nil {
		return nil
	}
	r := h.resource
	h.resource = nil
	h.cleanup.Stop()
	r.close()
	return nil
}

func cSize(n int, label string) (C.size_t, error) {
	if n < 0 || uint64(n) > uint64(^C.size_t(0)) {
		return 0, fmt.Errorf("sidereon: %s does not fit size_t", label)
	}
	return C.size_t(n), nil
}

func cSize64(n uint64, label string) (C.size_t, error) {
	if n > uint64(^C.size_t(0)) {
		return 0, fmt.Errorf("sidereon: %s does not fit size_t", label)
	}
	return C.size_t(n), nil
}

func cInt64(value int64, label string) (C.int64_t, error) {
	// C.int64_t is fixed width; keeping this helper at the boundary documents
	// the conversion site and prevents future unsigned-to-signed truncation.
	_ = label
	return C.int64_t(value), nil
}

func cFloats(values []float64, label string) (unsafe.Pointer, C.size_t, error) {
	length, err := cSize(len(values), label+" length")
	if err != nil {
		return nil, 0, err
	}
	if len(values) == 0 {
		return nil, length, nil
	}
	size, err := checkedNativeAllocationSize(len(values), unsafe.Sizeof(C.double(0)))
	if err != nil {
		return nil, 0, fmt.Errorf("sidereon: %s: %w", label, err)
	}
	pointer := C.malloc(C.size_t(size))
	if pointer == nil {
		return nil, 0, errors.New("sidereon: unable to allocate native float buffer")
	}
	output := unsafe.Slice((*C.double)(pointer), len(values))
	for index, value := range values {
		output[index] = C.double(value)
	}
	return pointer, length, nil
}

func freeCStrings(values []*C.char) {
	for _, value := range values {
		if value != nil {
			C.free(unsafe.Pointer(value))
		}
	}
}

type doubleOutputCall func(*C.double, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus

func copyNativeDoublesLocked(label string, call doubleOutputCall) ([]float64, error) {
	var written, required C.size_t
	if err := statusErrorLocked(uint32(call(nil, 0, &written, &required))); err != nil {
		return nil, err
	}
	count, err := sizeTToInt(required, label+" required count")
	if err != nil {
		return nil, err
	}
	if _, err := writtenToInt(written, 0, label+" first-call written count"); err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.double(0))); err != nil {
		return nil, err
	}
	buffer := make([]C.double, count)
	outputLength, err := cSize(len(buffer), label+" output capacity")
	if err != nil {
		return nil, err
	}
	var output *C.double
	if len(buffer) != 0 {
		output = &buffer[0]
	}
	written, required = 0, 0
	if err := statusErrorLocked(uint32(call(output, outputLength, &written, &required))); err != nil {
		return nil, err
	}
	writtenCount, err := validateTwoPassCounts(label, len(buffer), count, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	result := make([]float64, writtenCount)
	for index := range result {
		result[index] = float64(buffer[index])
	}
	return result, nil
}

func copyNativeDoubles(call doubleOutputCall, label string) (result []float64, err error) {
	withCThread(func() { result, err = copyNativeDoublesLocked(label, call) })
	return result, err
}

func copySurfaceDoubles(handle *surfaceHandle, label string, call func(unsafe.Pointer, *C.double, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus) ([]float64, error) {
	var result []float64
	var operationErr error
	err := handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			result, operationErr = copyNativeDoublesLocked(label, func(out *C.double, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return call(pointer, out, length, written, required)
			})
		})
		return operationErr
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func copySurfaceBytes(handle *surfaceHandle, label string, call func(unsafe.Pointer, *C.uint8_t, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus) ([]byte, error) {
	var result []byte
	var operationErr error
	err := handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			result, operationErr = copyNativeBytesLocked(label, func(out *C.uint8_t, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return call(pointer, out, length, written, required)
			})
		})
		return operationErr
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

type tokenOutputCall func(*C.SidereonSatelliteToken, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus

func copyNativeTokensLocked(label string, call tokenOutputCall) ([]string, error) {
	var written, required C.size_t
	if err := statusErrorLocked(uint32(call(nil, 0, &written, &required))); err != nil {
		return nil, err
	}
	count, err := sizeTToInt(required, label+" required count")
	if err != nil {
		return nil, err
	}
	if _, err := writtenToInt(written, 0, label+" first-call written count"); err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonSatelliteToken{})); err != nil {
		return nil, err
	}
	buffer := make([]C.SidereonSatelliteToken, count)
	outputLength, err := cSize(len(buffer), label+" output capacity")
	if err != nil {
		return nil, err
	}
	var output *C.SidereonSatelliteToken
	if len(buffer) != 0 {
		output = &buffer[0]
	}
	written, required = 0, 0
	if err := statusErrorLocked(uint32(call(output, outputLength, &written, &required))); err != nil {
		return nil, err
	}
	writtenCount, err := validateTwoPassCounts(label, len(buffer), count, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	result := make([]string, writtenCount)
	for index := range result {
		result[index] = tokenFromC(buffer[index])
	}
	return result, nil
}

func copyNativeTokens(call tokenOutputCall, label string) (result []string, err error) {
	withCThread(func() { result, err = copyNativeTokensLocked(label, call) })
	return result, err
}

func copySurfaceTokens(handle *surfaceHandle, label string, call func(unsafe.Pointer, *C.SidereonSatelliteToken, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus) ([]string, error) {
	var result []string
	var operationErr error
	err := handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			result, operationErr = copyNativeTokensLocked(label, func(out *C.SidereonSatelliteToken, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return call(pointer, out, length, written, required)
			})
		})
		return operationErr
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

type nativeRtkID struct{ Value string }

func rtkIDFromC(value C.SidereonRtkId) nativeRtkID {
	var bytes []byte
	for index := 0; index < len(value.bytes); index++ {
		if value.bytes[index] == 0 {
			break
		}
		bytes = append(bytes, byte(value.bytes[index]))
	}
	return nativeRtkID{Value: string(bytes)}
}

type NativeAlphaBetaState struct{ Level, Rate float64 }
type NativeAlphaBetaGains struct{ Alpha, Beta float64 }
type NativeAlphaBetaStep struct {
	Predicted  NativeAlphaBetaState
	Updated    NativeAlphaBetaState
	Innovation float64
}
type NativeScalarKalmanGains struct{ PositionGain, RateGain float64 }
type NativeResidualMoments struct{ Mean, Variance, Skewness, KurtosisExcess float64 }
type NativeJarqueBera struct{ Statistic, PValue float64 }
type NativeShapiroWilk struct{ W, PValue float64 }
type NativeIlsResult struct {
	FixedStatus         bool
	Ratio               float64
	BestScore           float64
	SecondBestPresent   bool
	SecondBestScore     float64
	CandidatesEvaluated uint64
}

type NativeWTestNoncentrality struct{ Delta0, Lambda0 float64 }

func WTestNoncentrality(alpha, beta float64) (NativeWTestNoncentrality, error) {
	var out C.SidereonWTestNoncentrality
	err := callStatus(func() uint32 { return uint32(C.sidereon_wtest_noncentrality(C.double(alpha), C.double(beta), &out)) })
	return NativeWTestNoncentrality{float64(out.delta0), float64(out.lambda0)}, err
}

func AlphaBetaSteadyStateGains(trackingIndex float64) (NativeAlphaBetaGains, error) {
	var output C.SidereonAlphaBetaGains
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_alpha_beta_steady_state_gains(C.double(trackingIndex), &output))
	})
	return NativeAlphaBetaGains{float64(output.alpha), float64(output.beta)}, err
}

func AlphaBetaFilterStep(state NativeAlphaBetaState, measurement, dt float64, gains NativeAlphaBetaGains) (NativeAlphaBetaStep, error) {
	inputState := C.SidereonAlphaBetaState{level: C.double(state.Level), rate: C.double(state.Rate)}
	inputGains := C.SidereonAlphaBetaGains{alpha: C.double(gains.Alpha), beta: C.double(gains.Beta)}
	var output C.SidereonAlphaBetaStep
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_alpha_beta_filter_step(&inputState, C.double(measurement), C.double(dt), &inputGains, &output))
	})
	return NativeAlphaBetaStep{
		Predicted:  NativeAlphaBetaState{float64(output.predicted.level), float64(output.predicted.rate)},
		Updated:    NativeAlphaBetaState{float64(output.updated.level), float64(output.updated.rate)},
		Innovation: float64(output.innovation),
	}, err
}

func ScalarKalmanSteadyStateGains(trackingIndex, dt, measurementVariance float64) (NativeScalarKalmanGains, error) {
	var output C.SidereonScalarKalmanGains
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_kalman_cv_steady_state_gains(C.double(trackingIndex), C.double(dt), C.double(measurementVariance), &output))
	})
	return NativeScalarKalmanGains{float64(output.position_gain), float64(output.rate_gain)}, err
}

func HessianTrace(jacobian []float64, rows, columns int) (float64, error) {
	if rows < 0 || columns < 0 {
		return 0, errors.New("sidereon: matrix shape must not be negative")
	}
	expected, productErr := checkedProduct(rows, columns, "Jacobian shape")
	if productErr != nil {
		return 0, productErr
	}
	if len(jacobian) != expected {
		return 0, fmt.Errorf("sidereon: Jacobian shape is %d x %d but has %d values", rows, columns, len(jacobian))
	}
	rowCount, err := cSize(rows, "Jacobian row count")
	if err != nil {
		return 0, err
	}
	columnCount, err := cSize(columns, "Jacobian column count")
	if err != nil {
		return 0, err
	}
	pointer, _, err := cFloats(jacobian, "Jacobian")
	if err != nil {
		return 0, err
	}
	if pointer != nil {
		defer C.free(pointer)
	}
	var output C.double
	err = callStatus(func() uint32 {
		return uint32(C.sidereon_hessian_trace((*C.double)(pointer), rowCount, columnCount, &output))
	})
	return float64(output), err
}

func NormalCovariance(jacobian []float64, rows, columns int, varianceScale float64) ([]float64, error) {
	if rows < 0 || columns < 0 {
		return nil, errors.New("sidereon: matrix shape must not be negative")
	}
	expected, productErr := checkedProduct(rows, columns, "Jacobian shape")
	if productErr != nil {
		return nil, productErr
	}
	if len(jacobian) != expected {
		return nil, fmt.Errorf("sidereon: Jacobian shape is %d x %d but has %d values", rows, columns, len(jacobian))
	}
	rowCount, err := cSize(rows, "Jacobian row count")
	if err != nil {
		return nil, err
	}
	columnCount, err := cSize(columns, "Jacobian column count")
	if err != nil {
		return nil, err
	}
	pointer, _, err := cFloats(jacobian, "Jacobian")
	if err != nil {
		return nil, err
	}
	if pointer != nil {
		defer C.free(pointer)
	}
	result, err := copyNativeDoubles(func(output *C.double, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_normal_covariance((*C.double)(pointer), rowCount, columnCount, C.double(varianceScale), output, length, written, required)
	}, "normal covariance")
	return result, err
}

func ResidualSkewness(values []float64, bias bool) (float64, error) {
	pointer, valueCount, err := cFloats(values, "residual values")
	if err != nil {
		return 0, err
	}
	if pointer != nil {
		defer C.free(pointer)
	}
	var output C.double
	err = callStatus(func() uint32 {
		return uint32(C.sidereon_residual_skewness((*C.double)(pointer), valueCount, C.bool(bias), &output))
	})
	return float64(output), err
}

func ResidualKurtosis(values []float64, fisher, bias bool) (float64, error) {
	pointer, valueCount, err := cFloats(values, "residual values")
	if err != nil {
		return 0, err
	}
	if pointer != nil {
		defer C.free(pointer)
	}
	var output C.double
	err = callStatus(func() uint32 {
		return uint32(C.sidereon_residual_kurtosis((*C.double)(pointer), valueCount, C.bool(fisher), C.bool(bias), &output))
	})
	return float64(output), err
}

func ResidualMoments(values []float64, fisher, bias bool) (NativeResidualMoments, error) {
	pointer, valueCount, err := cFloats(values, "residual values")
	if err != nil {
		return NativeResidualMoments{}, err
	}
	if pointer != nil {
		defer C.free(pointer)
	}
	var output C.SidereonResidualMoments
	err = callStatus(func() uint32 {
		return uint32(C.sidereon_residual_moments((*C.double)(pointer), valueCount, C.bool(fisher), C.bool(bias), &output))
	})
	return NativeResidualMoments{float64(output.mean), float64(output.variance), float64(output.skewness), float64(output.kurtosis_excess)}, err
}

func ResidualJarqueBera(values []float64) (NativeJarqueBera, error) {
	pointer, valueCount, err := cFloats(values, "residual values")
	if err != nil {
		return NativeJarqueBera{}, err
	}
	if pointer != nil {
		defer C.free(pointer)
	}
	var output C.SidereonJarqueBera
	err = callStatus(func() uint32 {
		return uint32(C.sidereon_residual_jarque_bera((*C.double)(pointer), valueCount, &output))
	})
	return NativeJarqueBera{float64(output.statistic), float64(output.p_value)}, err
}

func ResidualShapiroWilk(values []float64) (NativeShapiroWilk, error) {
	pointer, valueCount, err := cFloats(values, "residual values")
	if err != nil {
		return NativeShapiroWilk{}, err
	}
	if pointer != nil {
		defer C.free(pointer)
	}
	var output C.SidereonShapiroWilk
	err = callStatus(func() uint32 {
		return uint32(C.sidereon_residual_shapiro_wilk((*C.double)(pointer), valueCount, &output))
	})
	return NativeShapiroWilk{float64(output.w), float64(output.p_value)}, err
}

func MadGaussianConsistency() (float64, error) {
	var output C.double
	err := callStatus(func() uint32 { return uint32(C.sidereon_mad_gaussian_consistency(&output)) })
	return float64(output), err
}

func MadSpread(values []float64, scaleFloor float64) (float64, error) {
	pointer, valueCount, err := cFloats(values, "MAD values")
	if err != nil {
		return 0, err
	}
	if pointer != nil {
		defer C.free(pointer)
	}
	var output C.double
	err = callStatus(func() uint32 {
		return uint32(C.sidereon_mad_spread((*C.double)(pointer), valueCount, C.double(scaleFloor), &output))
	})
	return float64(output), err
}

func EWMA(previous, sample, alpha float64) (float64, error) {
	var output C.double
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_ewma_update(C.double(previous), C.double(sample), C.double(alpha), &output))
	})
	return float64(output), err
}

func EWMAPowerOfTwo(previous, sample float64, shift uint32) (float64, error) {
	var output C.double
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_ewma_update_power_of_two(C.double(previous), C.double(sample), C.uint32_t(shift), &output))
	})
	return float64(output), err
}

func CFARMultiplier(searchedCells uint64, falseAlarmProbability float64) (float64, error) {
	cellCount, err := cSize64(searchedCells, "searched-cell count")
	if err != nil {
		return 0, err
	}
	var output C.double
	err = callStatus(func() uint32 {
		return uint32(C.sidereon_cfar_ca_multiplier_from_pfa(cellCount, C.double(falseAlarmProbability), &output))
	})
	return float64(output), err
}

func CFARPFA(searchedCells uint64, multiplier float64) (float64, error) {
	cellCount, err := cSize64(searchedCells, "searched-cell count")
	if err != nil {
		return 0, err
	}
	var output C.double
	err = callStatus(func() uint32 {
		return uint32(C.sidereon_cfar_ca_pfa_from_multiplier(cellCount, C.double(multiplier), &output))
	})
	return float64(output), err
}

func CFARThreshold(searchedCells uint64, pfa, noiseLevel float64) (float64, error) {
	cellCount, err := cSize64(searchedCells, "searched-cell count")
	if err != nil {
		return 0, err
	}
	var output C.double
	err = callStatus(func() uint32 {
		return uint32(C.sidereon_cfar_ca_threshold(cellCount, C.double(pfa), C.double(noiseLevel), &output))
	})
	return float64(output), err
}

func CFARFalseAlarmProbability(searchedCells uint64, threshold, noiseLevel float64) (float64, error) {
	cellCount, err := cSize64(searchedCells, "searched-cell count")
	if err != nil {
		return 0, err
	}
	var output C.double
	err = callStatus(func() uint32 {
		return uint32(C.sidereon_cfar_ca_false_alarm_probability(cellCount, C.double(threshold), C.double(noiseLevel), &output))
	})
	return float64(output), err
}

func NormalizedInnovation(innovation, innovationVariance float64) (float64, error) {
	var output C.double
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_normalized_innovation(C.double(innovation), C.double(innovationVariance), &output))
	})
	return float64(output), err
}

func WeightVector(entries []NativeWeightEntry, options NativePseudorangeVarianceOptions) ([]float64, []bool, error) {
	return weightVectorOrSigmas(entries, options, false)
}

func Sigmas(entries []NativeWeightEntry, options NativePseudorangeVarianceOptions) ([]float64, []bool, error) {
	return weightVectorOrSigmas(entries, options, true)
}

type NativeWeightEntry struct {
	SatelliteID  string
	ElevationDeg float64
	HasCN0       bool
	CN0DBHz      float64
}
type NativePseudorangeVarianceOptions struct {
	AM, BM              float64
	Model               uint32
	HasCN0              bool
	CN0DBHz, CN0ScaleM2 float64
}

func weightVectorOrSigmas(entries []NativeWeightEntry, options NativePseudorangeVarianceOptions, sigmas bool) ([]float64, []bool, error) {
	if len(entries) > int(^uint(0)>>1) {
		return nil, nil, errors.New("sidereon: too many weight entries")
	}
	entryCount, err := cSize(len(entries), "weight entry count")
	if err != nil {
		return nil, nil, err
	}
	memorySize, err := checkedNativeAllocationSize(len(entries), unsafe.Sizeof(C.SidereonWeightEntry{}))
	if err != nil {
		return nil, nil, err
	}
	var memory unsafe.Pointer
	if len(entries) != 0 {
		memory = C.malloc(C.size_t(memorySize))
		if memory == nil {
			return nil, nil, errors.New("sidereon: unable to allocate native weight entries")
		}
		defer C.free(memory)
	}
	cEntries := unsafe.Slice((*C.SidereonWeightEntry)(memory), len(entries))
	for index, entry := range entries {
		id, release, err := cStrings([]string{entry.SatelliteID}, "weight satellite ID")
		if err != nil {
			return nil, nil, err
		}
		defer release()
		cEntries[index].sat_id = id[0]
		cEntries[index].elevation_deg = C.double(entry.ElevationDeg)
		cEntries[index].has_cn0 = C.bool(entry.HasCN0)
		cEntries[index].cn0_dbhz = C.double(entry.CN0DBHz)
	}
	cOptions := C.SidereonPseudorangeVarianceOptions{a_m: C.double(options.AM), b_m: C.double(options.BM), model: C.uint32_t(options.Model), has_cn0: C.bool(options.HasCN0), cn0_dbhz: C.double(options.CN0DBHz), cn0_scale_m2: C.double(options.CN0ScaleM2)}
	values := make([]C.double, len(entries))
	present := make([]C.bool, len(entries))
	var entryPointer *C.SidereonWeightEntry
	if len(entries) != 0 {
		entryPointer = &cEntries[0]
	}
	var valuePointer *C.double
	if len(values) != 0 {
		valuePointer = &values[0]
	}
	var presentPointer *C.bool
	if len(present) != 0 {
		presentPointer = &present[0]
	}
	var operationErr error
	withCThread(func() {
		var status C.enum_SidereonStatus
		if sigmas {
			status = C.sidereon_sigmas(entryPointer, entryCount, &cOptions, valuePointer, presentPointer)
		} else {
			status = C.sidereon_weight_vector(entryPointer, entryCount, &cOptions, valuePointer, presentPointer)
		}
		operationErr = statusErrorLocked(uint32(status))
	})
	if operationErr != nil {
		return nil, nil, operationErr
	}
	resultValues := make([]float64, len(values))
	resultPresent := make([]bool, len(present))
	for index := range values {
		resultValues[index] = float64(values[index])
		resultPresent[index] = bool(present[index])
	}
	return resultValues, resultPresent, nil
}

func CovarianceIsSymmetric(values [3][3]float64) (bool, error) {
	flat := matrix3ToCFlat(values)
	var output C.bool
	err := callStatus(func() uint32 { return uint32(C.sidereon_covariance_is_symmetric(&flat[0], &output)) })
	return bool(output), err
}

func CovarianceIsPositiveSemidefinite(values [3][3]float64) (bool, error) {
	flat := matrix3ToCFlat(values)
	var output C.bool
	err := callStatus(func() uint32 { return uint32(C.sidereon_covariance_is_positive_semidefinite(&flat[0], &output)) })
	return bool(output), err
}

func CovarianceFromJacobian(jacobian []float64, rows, columns int, cost float64) ([]float64, error) {
	expected, productErr := checkedProduct(rows, columns, "Jacobian shape")
	if productErr != nil {
		return nil, productErr
	}
	if len(jacobian) != expected {
		return nil, errors.New("sidereon: Jacobian shape does not match values")
	}
	rowCount, err := cSize(rows, "Jacobian row count")
	if err != nil {
		return nil, err
	}
	columnCount, err := cSize(columns, "Jacobian column count")
	if err != nil {
		return nil, err
	}
	pointer, _, err := cFloats(jacobian, "Jacobian")
	if err != nil {
		return nil, err
	}
	if pointer != nil {
		defer C.free(pointer)
	}
	return copyNativeDoubles(func(out *C.double, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_covariance_from_jacobian((*C.double)(pointer), rowCount, columnCount, C.double(cost), out, length, written, required)
	}, "covariance from Jacobian")
}

func CovarianceTransport(covariance [6][6]float64, segments []NativeCovarianceTransportSegment, processNoise NativeProcessNoise) ([][6][6]float64, error) {
	input := covarianceToC(covariance)
	segmentCount, err := cSize(len(segments), "covariance segment count")
	if err != nil {
		return nil, err
	}
	size, err := checkedNativeAllocationSize(len(segments), unsafe.Sizeof(C.SidereonCovarianceTransportSegment{}))
	if err != nil {
		return nil, err
	}
	var memory unsafe.Pointer
	if len(segments) != 0 {
		memory = C.malloc(C.size_t(size))
		if memory == nil {
			return nil, errors.New("sidereon: unable to allocate covariance segments")
		}
		defer C.free(memory)
	}
	cSegments := unsafe.Slice((*C.SidereonCovarianceTransportSegment)(memory), len(segments))
	for index, segment := range segments {
		cSegments[index] = covarianceTransportSegmentToC(segment)
	}
	var segmentPointer *C.SidereonCovarianceTransportSegment
	if len(segments) != 0 {
		segmentPointer = &cSegments[0]
	}
	noise := C.SidereonProcessNoise{kind: C.uint32_t(processNoise.Kind), q_radial_km2_s3: C.double(processNoise.RadialKm2S3), q_transverse_km2_s3: C.double(processNoise.TransverseKm2S3), q_normal_km2_s3: C.double(processNoise.NormalKm2S3)}
	var written, required C.size_t
	var operationErr error
	withCThread(func() {
		status := C.sidereon_covariance_transport(&input, segmentPointer, segmentCount, noise, nil, 0, &written, &required)
		operationErr = statusErrorLocked(uint32(status))
	})
	if operationErr != nil {
		return nil, operationErr
	}
	count, err := sizeTToInt(required, "covariance transport required count")
	if err != nil {
		return nil, err
	}
	if _, err := writtenToInt(written, 0, "covariance transport first-call written count"); err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonCovarianceMatrix6{})); err != nil {
		return nil, err
	}
	buffer := make([]C.SidereonCovarianceMatrix6, count)
	outputLength, err := cSize(len(buffer), "covariance transport output capacity")
	if err != nil {
		return nil, err
	}
	var output *C.SidereonCovarianceMatrix6
	if len(buffer) != 0 {
		output = &buffer[0]
	}
	written, required = 0, 0
	withCThread(func() {
		status := C.sidereon_covariance_transport(&input, segmentPointer, segmentCount, noise, output, outputLength, &written, &required)
		operationErr = statusErrorLocked(uint32(status))
	})
	if operationErr != nil {
		return nil, operationErr
	}
	actual, err := validateTwoPassCounts("covariance transport", len(buffer), count, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	result := make([][6][6]float64, actual)
	for index := range result {
		result[index] = covarianceFromC(&buffer[index])
	}
	return result, nil
}

func CovarianceECIToRTN(covariance [6][6]float64, state NativeCartesianState) ([6][6]float64, error) {
	input, cstate := covarianceToC(covariance), cartesianStateToC(state)
	var output C.SidereonCovarianceMatrix6
	err := callStatus(func() uint32 { return uint32(C.sidereon_covariance6_eci_to_rtn(&input, &cstate, &output)) })
	if err != nil {
		return [6][6]float64{}, err
	}
	return covarianceFromC(&output), nil
}
func CovarianceRTNToECI(covariance [6][6]float64, state NativeCartesianState) ([6][6]float64, error) {
	input, cstate := covarianceToC(covariance), cartesianStateToC(state)
	var output C.SidereonCovarianceMatrix6
	err := callStatus(func() uint32 { return uint32(C.sidereon_covariance6_rtn_to_eci(&input, &cstate, &output)) })
	if err != nil {
		return [6][6]float64{}, err
	}
	return covarianceFromC(&output), nil
}

type NativeProcessNoise struct {
	Kind                                      uint32
	RadialKm2S3, TransverseKm2S3, NormalKm2S3 float64
}
type NativeCovarianceTransportSegment struct {
	STM            [6][6]float64
	DTSeconds      float64
	QRotationState NativeCartesianState
}
type NativeCartesianState struct {
	EpochS                  float64
	PositionKm, VelocityKmS [3]float64
}

func covarianceTransportSegmentToC(value NativeCovarianceTransportSegment) C.SidereonCovarianceTransportSegment {
	return C.SidereonCovarianceTransportSegment{stm: covarianceToC(value.STM), dt_seconds: C.double(value.DTSeconds), q_rotation_state: cartesianStateToC(value.QRotationState)}
}
func cartesianStateToC(value NativeCartesianState) C.SidereonCartesianState {
	result := C.SidereonCartesianState{epoch_s: C.double(value.EpochS)}
	for index := 0; index < 3; index++ {
		result.position_km[index] = C.double(value.PositionKm[index])
		result.velocity_km_s[index] = C.double(value.VelocityKmS[index])
	}
	return result
}
func matrix3ToCFlat(values [3][3]float64) [9]C.double {
	var result [9]C.double
	for row := 0; row < 3; row++ {
		for column := 0; column < 3; column++ {
			result[row*3+column] = C.double(values[row][column])
		}
	}
	return result
}

func LambdaILS(floatCycles, covariance []float64, n int, ratioThreshold float64, bounded *NativeBoundedILS) ([]int64, NativeIlsResult, error) {
	if n < 1 || len(floatCycles) != n {
		return nil, NativeIlsResult{}, errors.New("sidereon: ILS dimension does not match float ambiguities")
	}
	expected, productErr := checkedProduct(n, n, "ILS covariance shape")
	if productErr != nil {
		return nil, NativeIlsResult{}, productErr
	}
	if len(covariance) != expected {
		return nil, NativeIlsResult{}, errors.New("sidereon: ILS covariance shape must be n by n")
	}
	nativeN, err := cSize(n, "ILS dimension")
	if err != nil {
		return nil, NativeIlsResult{}, err
	}
	floatPointer, floatLength, err := cFloats(floatCycles, "ILS ambiguities")
	if err != nil {
		return nil, NativeIlsResult{}, err
	}
	defer C.free(floatPointer)
	covPointer, covarianceLength, err := cFloats(covariance, "ILS covariance")
	if err != nil {
		return nil, NativeIlsResult{}, err
	}
	defer C.free(covPointer)
	if _, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.int64_t(0))); err != nil {
		return nil, NativeIlsResult{}, err
	}
	fixed := make([]C.int64_t, n)
	var result C.SidereonIlsResult
	var status C.enum_SidereonStatus
	var operationErr error
	var candidateLimit C.size_t
	if bounded != nil {
		candidateLimit, err = cSize64(bounded.CandidateLimit, "ILS candidate limit")
		if err != nil {
			return nil, NativeIlsResult{}, err
		}
	}
	withCThread(func() {
		if bounded == nil {
			status = C.sidereon_lambda_ils_search((*C.double)(floatPointer), nativeN, (*C.double)(covPointer), covarianceLength, C.double(ratioThreshold), &fixed[0], &result)
		} else {
			status = C.sidereon_bounded_ils_search((*C.double)(floatPointer), nativeN, (*C.double)(covPointer), covarianceLength, C.int64_t(bounded.Radius), candidateLimit, C.double(ratioThreshold), &fixed[0], &result)
		}
		operationErr = statusErrorLocked(uint32(status))
	})
	if operationErr != nil {
		return nil, NativeIlsResult{}, operationErr
	}
	if floatLength != nativeN {
		return nil, NativeIlsResult{}, errors.New("sidereon: ILS ambiguity count changed during conversion")
	}
	output := make([]int64, n)
	for index := range fixed {
		output[index] = int64(fixed[index])
	}
	return output, NativeIlsResult{bool(result.fixed_status), float64(result.ratio), float64(result.best_score), bool(result.second_best_present), float64(result.second_best_score), uint64(result.candidates_evaluated)}, nil
}

type NativeBoundedILS struct {
	Radius         int64
	CandidateLimit uint64
}

type NativeFDERaimWeight struct {
	SatelliteID string
	Weight      float64
}
type NativeRaimResult struct {
	FaultDetected           bool
	TestStatistic           float64
	HasThreshold            bool
	Threshold               float64
	HasReducedChiSquare     bool
	ReducedChiSquare        float64
	RMSM                    float64
	DOF                     int64
	Testable                bool
	NormalizedResidualCount uint64
	HasWorstSatellite       bool
	WorstSatellite          string
}
type NativeRaimNormalizedResidual struct {
	SatelliteID        string
	NormalizedResidual float64
}

func makeFDERaimWeights(values []NativeFDERaimWeight) (unsafe.Pointer, C.size_t, [](*C.char), error) {
	length, err := cSize(len(values), "RAIM weight count")
	if err != nil {
		return nil, 0, nil, err
	}
	size, err := checkedNativeAllocationSize(len(values), unsafe.Sizeof(C.SidereonFdeRaimWeight{}))
	if err != nil {
		return nil, 0, nil, err
	}
	var memory unsafe.Pointer
	if len(values) != 0 {
		memory = C.malloc(C.size_t(size))
		if memory == nil {
			return nil, 0, nil, errors.New("sidereon: unable to allocate RAIM weights")
		}
	}
	cValues := unsafe.Slice((*C.SidereonFdeRaimWeight)(memory), len(values))
	ids := make([]*C.char, len(values))
	for index, value := range values {
		if err := rejectEmbeddedNUL(value.SatelliteID, "RAIM satellite ID"); err != nil {
			freeCStrings(ids)
			C.free(memory)
			return nil, 0, nil, err
		}
		ids[index] = C.CString(value.SatelliteID)
		if ids[index] == nil {
			freeCStrings(ids)
			C.free(memory)
			return nil, 0, nil, errors.New("sidereon: unable to allocate native string")
		}
		cValues[index].sat_id = ids[index]
		cValues[index].weight = C.double(value.Weight)
	}
	return memory, length, ids, nil
}

func RAIM(ids []string, residuals []float64, weights []NativeFDERaimWeight, pfa float64, unitWeights, systemsEnabled bool, systems int64) (NativeRaimResult, []NativeRaimNormalizedResidual, error) {
	if len(ids) != len(residuals) {
		return NativeRaimResult{}, nil, errors.New("sidereon: RAIM satellite and residual lengths differ")
	}
	idCount, err := cSize(len(ids), "RAIM observation count")
	if err != nil {
		return NativeRaimResult{}, nil, err
	}
	idPointers, releaseIDs, err := cStrings(ids, "RAIM satellite ID")
	if err != nil {
		return NativeRaimResult{}, nil, err
	}
	defer releaseIDs()
	residualPointer, residualCount, err := cFloats(residuals, "RAIM residuals")
	if err != nil {
		return NativeRaimResult{}, nil, err
	}
	if residualPointer != nil {
		defer C.free(residualPointer)
	}
	weightMemory, weightLength, weightIDs, err := makeFDERaimWeights(weights)
	if err != nil {
		return NativeRaimResult{}, nil, err
	}
	if weightMemory != nil {
		defer C.free(weightMemory)
	}
	defer freeCStrings(weightIDs)
	var weightPointer *C.SidereonFdeRaimWeight
	if weightMemory != nil {
		weightPointer = (*C.SidereonFdeRaimWeight)(weightMemory)
	}
	var idPointer **C.char
	if len(idPointers) != 0 {
		idPointer = &idPointers[0]
	}
	var result C.SidereonRaimResult
	var status C.enum_SidereonStatus
	var operationErr error
	withCThread(func() {
		status = C.sidereon_raim(idPointer, (*C.double)(residualPointer), idCount, C.double(pfa), C.bool(unitWeights), weightPointer, weightLength, C.bool(systemsEnabled), C.int64_t(systems), &result)
		operationErr = statusErrorLocked(uint32(status))
	})
	if operationErr != nil {
		return NativeRaimResult{}, nil, operationErr
	}
	nativeResult := NativeRaimResult{bool(result.fault_detected), float64(result.test_statistic), bool(result.has_threshold), float64(result.threshold), bool(result.has_reduced_chi_square), float64(result.reduced_chi_square), float64(result.rms_m), int64(result.dof), bool(result.testable), uint64(result.normalized_residual_count), bool(result.has_worst_sat), tokenChars(result.worst_sat[:])}
	var written, required C.size_t
	var status2 C.enum_SidereonStatus
	withCThread(func() {
		status2 = C.sidereon_raim_normalized_residuals(idPointer, (*C.double)(residualPointer), idCount, C.double(pfa), C.bool(unitWeights), weightPointer, weightLength, C.bool(systemsEnabled), C.int64_t(systems), nil, 0, &written, &required)
		operationErr = statusErrorLocked(uint32(status2))
	})
	if operationErr != nil {
		return nativeResult, nil, operationErr
	}
	count, err := sizeTToInt(required, "RAIM normalized residual count")
	if err != nil {
		return nativeResult, nil, err
	}
	if _, err := writtenToInt(written, 0, "RAIM normalized residual first-call written count"); err != nil {
		return nativeResult, nil, err
	}
	if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonRaimNormalizedResidual{})); err != nil {
		return nativeResult, nil, err
	}
	buffer := make([]C.SidereonRaimNormalizedResidual, count)
	outputLength, err := cSize(len(buffer), "RAIM normalized residual output capacity")
	if err != nil {
		return nativeResult, nil, err
	}
	var output *C.SidereonRaimNormalizedResidual
	if len(buffer) != 0 {
		output = &buffer[0]
	}
	written, required = 0, 0
	withCThread(func() {
		status2 = C.sidereon_raim_normalized_residuals(idPointer, (*C.double)(residualPointer), idCount, C.double(pfa), C.bool(unitWeights), weightPointer, weightLength, C.bool(systemsEnabled), C.int64_t(systems), output, outputLength, &written, &required)
		operationErr = statusErrorLocked(uint32(status2))
	})
	if operationErr != nil {
		return nativeResult, nil, operationErr
	}
	if residualCount != idCount {
		return NativeRaimResult{}, nil, errors.New("sidereon: RAIM residual count changed during conversion")
	}
	actual, err := validateTwoPassCounts("RAIM normalized residuals", len(buffer), count, uint64(written), uint64(required))
	if err != nil {
		return nativeResult, nil, err
	}
	rows := make([]NativeRaimNormalizedResidual, actual)
	for index := range rows {
		rows[index] = NativeRaimNormalizedResidual{tokenFromC(buffer[index].sat_id), float64(buffer[index].normalized_residual)}
	}
	return nativeResult, rows, nil
}

func tokenChars(values []C.char) string {
	bytes := make([]byte, 0, len(values))
	for _, value := range values {
		if value == 0 {
			break
		}
		bytes = append(bytes, byte(value))
	}
	return string(bytes)
}
