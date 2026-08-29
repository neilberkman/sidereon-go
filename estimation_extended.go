package sidereon

import (
	"sort"

	"github.com/neilberkman/sidereon-go/internal/native"
)

// AlphaBetaState is a scalar level/rate state. Level and rate use caller-chosen
// units; rate is those units per second.
type AlphaBetaState struct{ Level, Rate float64 }

// AlphaBetaGains contains the dimensionless alpha and beta gains computed by C.
type AlphaBetaGains struct{ Alpha, Beta float64 }

// AlphaBetaStep contains C's prediction, innovation, and update.
type AlphaBetaStep struct {
	Predicted, Updated AlphaBetaState
	Innovation         float64
}

// AlphaBetaSteadyStateGains computes steady-state gains for a positive
// dimensionless tracking index.
func AlphaBetaSteadyStateGains(trackingIndex float64) (AlphaBetaGains, error) {
	value, err := native.AlphaBetaSteadyStateGains(trackingIndex)
	return AlphaBetaGains{value.Alpha, value.Beta}, publicError(err)
}

// AlphaBetaFilterStep performs one C alpha-beta predict/update step.
func AlphaBetaFilterStep(state AlphaBetaState, measurement, dt float64, gains AlphaBetaGains) (AlphaBetaStep, error) {
	value, err := native.AlphaBetaFilterStep(native.NativeAlphaBetaState(state), measurement, dt, native.NativeAlphaBetaGains(gains))
	return AlphaBetaStep{Predicted: AlphaBetaState(value.Predicted), Updated: AlphaBetaState(value.Updated), Innovation: value.Innovation}, publicError(err)
}

// ScalarKalmanGains are steady-state constant-velocity Kalman gains.
type ScalarKalmanGains struct{ PositionGain, RateGain float64 }

// ScalarKalmanSteadyStateGains computes C's constant-velocity gains. RateGain
// is expressed in inverse seconds when applied to a position innovation.
func ScalarKalmanSteadyStateGains(trackingIndex, dt, measurementVariance float64) (ScalarKalmanGains, error) {
	value, err := native.ScalarKalmanSteadyStateGains(trackingIndex, dt, measurementVariance)
	return ScalarKalmanGains{value.PositionGain, value.RateGain}, publicError(err)
}

// NormalizedInnovation returns innovation/sqrt(variance), a dimensionless
// signed innovation statistic computed by C.
func NormalizedInnovation(innovation, innovationVariance float64) (float64, error) {
	value, err := native.NormalizedInnovation(innovation, innovationVariance)
	return value, publicError(err)
}

// ResidualMoments contains the C-computed first four residual moments.
type ResidualMoments struct{ Mean, Variance, Skewness, KurtosisExcess float64 }
type JarqueBera struct{ Statistic, PValue float64 }
type ShapiroWilk struct{ W, PValue float64 }

func ResidualSkewness(values []float64, bias bool) (float64, error) {
	value, err := native.ResidualSkewness(values, bias)
	return value, publicError(err)
}
func ResidualKurtosis(values []float64, fisher, bias bool) (float64, error) {
	value, err := native.ResidualKurtosis(values, fisher, bias)
	return value, publicError(err)
}
func ResidualMomentsOf(values []float64, fisher, bias bool) (ResidualMoments, error) {
	value, err := native.ResidualMoments(values, fisher, bias)
	return ResidualMoments{value.Mean, value.Variance, value.Skewness, value.KurtosisExcess}, publicError(err)
}
func ResidualJarqueBera(values []float64) (JarqueBera, error) {
	value, err := native.ResidualJarqueBera(values)
	return JarqueBera{value.Statistic, value.PValue}, publicError(err)
}
func ResidualShapiroWilk(values []float64) (ShapiroWilk, error) {
	value, err := native.ResidualShapiroWilk(values)
	return ShapiroWilk{value.W, value.PValue}, publicError(err)
}

func MADGaussianConsistency() (float64, error) {
	value, err := native.MadGaussianConsistency()
	return value, publicError(err)
}
func MADSpread(values []float64, scaleFloor float64) (float64, error) {
	value, err := native.MadSpread(values, scaleFloor)
	return value, publicError(err)
}
func EWMA(previous, sample, alpha float64) (float64, error) {
	value, err := native.EWMA(previous, sample, alpha)
	return value, publicError(err)
}
func EWMAPowerOfTwo(previous, sample float64, shift uint32) (float64, error) {
	value, err := native.EWMAPowerOfTwo(previous, sample, shift)
	return value, publicError(err)
}
func CFARMultiplier(searchedCells uint64, falseAlarmProbability float64) (float64, error) {
	value, err := native.CFARMultiplier(searchedCells, falseAlarmProbability)
	return value, publicError(err)
}
func CFARPFA(searchedCells uint64, multiplier float64) (float64, error) {
	value, err := native.CFARPFA(searchedCells, multiplier)
	return value, publicError(err)
}
func CFARThreshold(searchedCells uint64, falseAlarmProbability, noiseLevel float64) (float64, error) {
	value, err := native.CFARThreshold(searchedCells, falseAlarmProbability, noiseLevel)
	return value, publicError(err)
}
func CFARFalseAlarmProbability(searchedCells uint64, threshold, noiseLevel float64) (float64, error) {
	value, err := native.CFARFalseAlarmProbability(searchedCells, threshold, noiseLevel)
	return value, publicError(err)
}

// PseudorangeVarianceModel selects the C pseudorange variance model.
type PseudorangeVarianceModel uint32

const (
	PseudorangeVarianceElevation    PseudorangeVarianceModel = PseudorangeVarianceModel(native.PseudorangeVarianceElevationValue)
	PseudorangeVarianceElevationCN0 PseudorangeVarianceModel = PseudorangeVarianceModel(native.PseudorangeVarianceElevationCN0Value)
)

type PseudorangeVarianceOptions struct {
	AM, BM              float64
	Model               PseudorangeVarianceModel
	HasCN0              bool
	CN0DBHz, CN0ScaleM2 float64
}
type WeightEntry struct {
	SatelliteID  string
	ElevationDeg float64
	HasCN0       bool
	CN0DBHz      float64
}
type SigmaValue struct {
	Value   float64
	Present bool
}

func nativeVarianceOptions(value PseudorangeVarianceOptions) native.NativePseudorangeVarianceOptions {
	return native.NativePseudorangeVarianceOptions{AM: value.AM, BM: value.BM, Model: uint32(value.Model), HasCN0: value.HasCN0, CN0DBHz: value.CN0DBHz, CN0ScaleM2: value.CN0ScaleM2}
}
func nativeWeightEntries(values []WeightEntry) []native.NativeWeightEntry {
	result := make([]native.NativeWeightEntry, len(values))
	for index, value := range values {
		result[index] = native.NativeWeightEntry{SatelliteID: value.SatelliteID, ElevationDeg: value.ElevationDeg, HasCN0: value.HasCN0, CN0DBHz: value.CN0DBHz}
	}
	return result
}

func nativeWeightMap(values map[string]float64) []native.NativeFDERaimWeight {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]native.NativeFDERaimWeight, len(keys))
	for index, key := range keys {
		result[index] = native.NativeFDERaimWeight{SatelliteID: key, Weight: values[key]}
	}
	return result
}

func WeightVector(entries []WeightEntry, options PseudorangeVarianceOptions) ([]SigmaValue, error) {
	values, present, err := native.WeightVector(nativeWeightEntries(entries), nativeVarianceOptions(options))
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]SigmaValue, len(values))
	for index := range values {
		result[index] = SigmaValue{values[index], present[index]}
	}
	return result, nil
}
func PseudorangeSigmas(entries []WeightEntry, options PseudorangeVarianceOptions) ([]SigmaValue, error) {
	values, present, err := native.Sigmas(nativeWeightEntries(entries), nativeVarianceOptions(options))
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]SigmaValue, len(values))
	for index := range values {
		result[index] = SigmaValue{values[index], present[index]}
	}
	return result, nil
}

// NormalCovariance returns the row-major n-by-n covariance produced from an
// m-by-n Jacobian. The shape is validated before entering C.
func NormalCovariance(jacobian []float64, rows, columns int, varianceScale float64) ([]float64, error) {
	value, err := native.NormalCovariance(append([]float64(nil), jacobian...), rows, columns, varianceScale)
	return value, publicError(err)
}
func HessianTrace(jacobian []float64, rows, columns int) (float64, error) {
	value, err := native.HessianTrace(append([]float64(nil), jacobian...), rows, columns)
	return value, publicError(err)
}
func CovarianceFromJacobian(jacobian []float64, rows, columns int, cost float64) ([]float64, error) {
	value, err := native.CovarianceFromJacobian(append([]float64(nil), jacobian...), rows, columns, cost)
	return value, publicError(err)
}
func CovarianceIsSymmetric(covariance [3][3]float64) (bool, error) {
	value, err := native.CovarianceIsSymmetric(covariance)
	return value, publicError(err)
}
func CovarianceIsPositiveSemidefinite(covariance [3][3]float64) (bool, error) {
	value, err := native.CovarianceIsPositiveSemidefinite(covariance)
	return value, publicError(err)
}

// ILSResult preserves C's search diagnostics, including absent runner-up
// state. Fixed is a copied integer ambiguity vector in input order.
type ILSResult struct {
	Fixed               []int64
	FixedStatus         bool
	Ratio, BestScore    float64
	SecondBestPresent   bool
	SecondBestScore     float64
	CandidatesEvaluated uint64
}
type BoundedILSOptions struct {
	Radius         int64
	CandidateLimit uint64
}

func searchILS(floatCycles, covariance []float64, ratioThreshold float64, bounded *BoundedILSOptions) (ILSResult, error) {
	var nativeBounded *native.NativeBoundedILS
	if bounded != nil {
		nativeBounded = &native.NativeBoundedILS{Radius: bounded.Radius, CandidateLimit: bounded.CandidateLimit}
	}
	fixed, value, err := native.LambdaILS(append([]float64(nil), floatCycles...), append([]float64(nil), covariance...), len(floatCycles), ratioThreshold, nativeBounded)
	return ILSResult{Fixed: append([]int64(nil), fixed...), FixedStatus: value.FixedStatus, Ratio: value.Ratio, BestScore: value.BestScore, SecondBestPresent: value.SecondBestPresent, SecondBestScore: value.SecondBestScore, CandidatesEvaluated: value.CandidatesEvaluated}, publicError(err)
}
func LambdaILS(floatCycles, covariance []float64, ratioThreshold float64) (ILSResult, error) {
	return searchILS(floatCycles, covariance, ratioThreshold, nil)
}
func BoundedILS(floatCycles, covariance []float64, options BoundedILSOptions, ratioThreshold float64) (ILSResult, error) {
	return searchILS(floatCycles, covariance, ratioThreshold, &options)
}

// RAIMOptions controls the C standalone chi-square integrity test.
type RAIMOptions struct {
	PFA            float64
	UnitWeights    bool
	Weights        map[string]float64
	SystemsEnabled bool
	Systems        int64
}
type RAIMInput struct {
	SatelliteIDs []string
	ResidualsM   []float64
}
type RAIMResult struct {
	FaultDetected           bool
	TestStatistic           float64
	HasThreshold            bool
	Threshold               float64
	HasReducedChiSquare     bool
	ReducedChiSquare        float64
	RMSM                    float64
	DOF                     int64
	Testable                bool
	NormalizedResidualCount int
	HasWorstSatellite       bool
	WorstSatellite          string
}
type RAIMNormalizedResidual struct {
	SatelliteID        string
	NormalizedResidual float64
}

func RAIM(input RAIMInput, options RAIMOptions) (RAIMResult, []RAIMNormalizedResidual, error) {
	weights := nativeWeightMap(options.Weights)
	value, rows, err := native.RAIM(input.SatelliteIDs, input.ResidualsM, weights, options.PFA, options.UnitWeights, options.SystemsEnabled, options.Systems)
	if err != nil {
		return RAIMResult{}, nil, publicError(err)
	}
	normalizedCount, conversionErr := nativeCountToInt(value.NormalizedResidualCount, "RAIM normalized residual count")
	if conversionErr != nil {
		return RAIMResult{}, nil, conversionErr
	}
	result := RAIMResult{FaultDetected: value.FaultDetected, TestStatistic: value.TestStatistic, HasThreshold: value.HasThreshold, Threshold: value.Threshold, HasReducedChiSquare: value.HasReducedChiSquare, ReducedChiSquare: value.ReducedChiSquare, RMSM: value.RMSM, DOF: value.DOF, Testable: value.Testable, NormalizedResidualCount: normalizedCount, HasWorstSatellite: value.HasWorstSatellite, WorstSatellite: value.WorstSatellite}
	out := make([]RAIMNormalizedResidual, len(rows))
	for index, row := range rows {
		out[index] = RAIMNormalizedResidual{row.SatelliteID, row.NormalizedResidual}
	}
	return result, out, nil
}
