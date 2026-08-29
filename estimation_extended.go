package sidereon

import (
	"sort"

	"github.com/neilberkman/sidereon-go/internal/native"
)

// AlphaBetaState is a scalar level/rate state. Level uses caller-chosen units;
// Rate uses those units per second.
type AlphaBetaState struct{ Level, Rate float64 }

// AlphaBetaGains contains the dimensionless Alpha level and Beta rate gains computed by C.
type AlphaBetaGains struct{ Alpha, Beta float64 }

// AlphaBetaStep contains C's prediction, innovation, and update.
type AlphaBetaStep struct {
	// Predicted and Updated are the pre- and post-measurement states.
	Predicted, Updated AlphaBetaState
	// Innovation is the measurement residual in level units.
	Innovation float64
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

// ScalarKalmanGains are steady-state constant-velocity Kalman gains. PositionGain
// is dimensionless; RateGain is in inverse seconds.
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

// ResidualMoments contains the C-computed first four residual moments: Mean in
// input units, Variance in squared input units, and dimensionless Skewness and
// KurtosisExcess.
type ResidualMoments struct{ Mean, Variance, Skewness, KurtosisExcess float64 }

// JarqueBera contains a Jarque-Bera statistic and its p-value.
type JarqueBera struct{ Statistic, PValue float64 }

// ShapiroWilk contains the Shapiro-Wilk W statistic and its p-value.
type ShapiroWilk struct{ W, PValue float64 }

// ResidualSkewness computes the C residual skewness estimate.
func ResidualSkewness(values []float64, bias bool) (float64, error) {
	value, err := native.ResidualSkewness(values, bias)
	return value, publicError(err)
}

// ResidualKurtosis computes the C residual kurtosis estimate.
func ResidualKurtosis(values []float64, fisher, bias bool) (float64, error) {
	value, err := native.ResidualKurtosis(values, fisher, bias)
	return value, publicError(err)
}

// ResidualMomentsOf computes mean, variance, skewness, and excess kurtosis.
func ResidualMomentsOf(values []float64, fisher, bias bool) (ResidualMoments, error) {
	value, err := native.ResidualMoments(values, fisher, bias)
	return ResidualMoments{value.Mean, value.Variance, value.Skewness, value.KurtosisExcess}, publicError(err)
}

// ResidualJarqueBera computes the C Jarque-Bera normality statistic.
func ResidualJarqueBera(values []float64) (JarqueBera, error) {
	value, err := native.ResidualJarqueBera(values)
	return JarqueBera{value.Statistic, value.PValue}, publicError(err)
}

// ResidualShapiroWilk computes the C Shapiro-Wilk normality statistic.
func ResidualShapiroWilk(values []float64) (ShapiroWilk, error) {
	value, err := native.ResidualShapiroWilk(values)
	return ShapiroWilk{value.W, value.PValue}, publicError(err)
}

// MADGaussianConsistency returns C's Gaussian consistency factor for MAD.
func MADGaussianConsistency() (float64, error) {
	value, err := native.MadGaussianConsistency()
	return value, publicError(err)
}

// MADSpread computes a median-absolute-deviation spread with a scale floor.
func MADSpread(values []float64, scaleFloor float64) (float64, error) {
	value, err := native.MadSpread(values, scaleFloor)
	return value, publicError(err)
}

// EWMA computes one exponentially weighted moving-average update.
func EWMA(previous, sample, alpha float64) (float64, error) {
	value, err := native.EWMA(previous, sample, alpha)
	return value, publicError(err)
}

// EWMAPowerOfTwo computes an EWMA update with alpha equal to a power of two.
func EWMAPowerOfTwo(previous, sample float64, shift uint32) (float64, error) {
	value, err := native.EWMAPowerOfTwo(previous, sample, shift)
	return value, publicError(err)
}

// CFARMultiplier computes the constant-false-alarm-rate multiplier.
func CFARMultiplier(searchedCells uint64, falseAlarmProbability float64) (float64, error) {
	value, err := native.CFARMultiplier(searchedCells, falseAlarmProbability)
	return value, publicError(err)
}

// CFARPFA computes the false-alarm probability for a CFAR multiplier.
func CFARPFA(searchedCells uint64, multiplier float64) (float64, error) {
	value, err := native.CFARPFA(searchedCells, multiplier)
	return value, publicError(err)
}

// CFARThreshold computes a CFAR threshold from noise level and probability.
func CFARThreshold(searchedCells uint64, falseAlarmProbability, noiseLevel float64) (float64, error) {
	value, err := native.CFARThreshold(searchedCells, falseAlarmProbability, noiseLevel)
	return value, publicError(err)
}

// CFARFalseAlarmProbability computes the probability at a CFAR threshold.
func CFARFalseAlarmProbability(searchedCells uint64, threshold, noiseLevel float64) (float64, error) {
	value, err := native.CFARFalseAlarmProbability(searchedCells, threshold, noiseLevel)
	return value, publicError(err)
}

// PseudorangeVarianceModel selects the C pseudorange variance model.
type PseudorangeVarianceModel uint32

const (
	// PseudorangeVarianceElevation uses elevation-angle weighting.
	PseudorangeVarianceElevation PseudorangeVarianceModel = PseudorangeVarianceModel(native.PseudorangeVarianceElevationValue)
	// PseudorangeVarianceElevationCN0 combines elevation and carrier-to-noise weighting.
	PseudorangeVarianceElevationCN0 PseudorangeVarianceModel = PseudorangeVarianceModel(native.PseudorangeVarianceElevationCN0Value)
)

// PseudorangeVarianceOptions selects pseudorange variance coefficients and model.
type PseudorangeVarianceOptions struct {
	// AM and BM are native variance-model coefficients.
	AM, BM float64
	// Model selects the elevation/CN0 variance model.
	Model PseudorangeVarianceModel
	// HasCN0 controls whether CN0 fields are used.
	HasCN0 bool
	// CN0DBHz is carrier-to-noise density in dB-Hz; CN0ScaleM2 is m².
	CN0DBHz, CN0ScaleM2 float64
}

// WeightEntry supplies one satellite's elevation and optional C/N0 value.
type WeightEntry struct {
	// SatelliteID identifies the observation.
	SatelliteID string
	// ElevationDeg is satellite elevation in degrees.
	ElevationDeg float64
	// HasCN0 controls use of CN0DBHz.
	HasCN0 bool
	// CN0DBHz is carrier-to-noise density in dB-Hz when present.
	CN0DBHz float64
}

// SigmaValue contains a C-computed sigma and whether that sigma is present.
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

// WeightVector computes C's pseudorange weight vector and presence flags.
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

// PseudorangeSigmas computes C's pseudorange standard deviations and presence flags.
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

// HessianTrace computes the trace of the normal Hessian for a row-major Jacobian.
func HessianTrace(jacobian []float64, rows, columns int) (float64, error) {
	value, err := native.HessianTrace(append([]float64(nil), jacobian...), rows, columns)
	return value, publicError(err)
}

// CovarianceFromJacobian computes a covariance from a row-major Jacobian.
func CovarianceFromJacobian(jacobian []float64, rows, columns int, cost float64) ([]float64, error) {
	value, err := native.CovarianceFromJacobian(append([]float64(nil), jacobian...), rows, columns, cost)
	return value, publicError(err)
}

// CovarianceIsSymmetric reports whether a 3x3 covariance is symmetric.
func CovarianceIsSymmetric(covariance [3][3]float64) (bool, error) {
	value, err := native.CovarianceIsSymmetric(covariance)
	return value, publicError(err)
}

// CovarianceIsPositiveSemidefinite reports whether a 3x3 covariance is PSD.
func CovarianceIsPositiveSemidefinite(covariance [3][3]float64) (bool, error) {
	value, err := native.CovarianceIsPositiveSemidefinite(covariance)
	return value, publicError(err)
}

// ILSResult preserves C's search diagnostics, including absent runner-up
// state. Fixed is a copied integer ambiguity vector in input order.
type ILSResult struct {
	// Fixed is the detached integer ambiguity vector.
	Fixed []int64
	// FixedStatus reports whether integer fixing succeeded.
	FixedStatus bool
	// Ratio and scores contain native integer-search diagnostics.
	Ratio, BestScore float64
	// SecondBestPresent controls whether SecondBestScore is valid.
	SecondBestPresent bool
	// SecondBestScore is valid only when SecondBestPresent is true.
	SecondBestScore float64
	// CandidatesEvaluated is the native search count.
	CandidatesEvaluated uint64
}

// BoundedILSOptions limits the integer least-squares search radius and count.
type BoundedILSOptions struct {
	// Radius bounds each integer candidate.
	Radius int64
	// CandidateLimit bounds the native search count.
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

// LambdaILS performs the unbounded integer least-squares search.
func LambdaILS(floatCycles, covariance []float64, ratioThreshold float64) (ILSResult, error) {
	return searchILS(floatCycles, covariance, ratioThreshold, nil)
}

// BoundedILS performs integer least-squares search with explicit bounds.
func BoundedILS(floatCycles, covariance []float64, options BoundedILSOptions, ratioThreshold float64) (ILSResult, error) {
	return searchILS(floatCycles, covariance, ratioThreshold, &options)
}

// RAIMOptions controls the C standalone chi-square integrity test.
type RAIMOptions struct {
	// PFA is the requested false-alarm probability.
	PFA float64
	// UnitWeights selects unit residual weights.
	UnitWeights bool
	// Weights optionally maps satellite IDs to residual weights.
	Weights map[string]float64
	// SystemsEnabled controls constellation grouping.
	SystemsEnabled bool
	// Systems is the native constellation mask.
	Systems int64
}

// RAIMInput supplies satellite identifiers and metre residuals to RAIM.
type RAIMInput struct {
	// SatelliteIDs names residual rows in matching order.
	SatelliteIDs []string
	// ResidualsM contains residuals in metres in matching order.
	ResidualsM []float64
}

// RAIMResult contains C's receiver-autonomous integrity-monitoring summary.
type RAIMResult struct {
	// FaultDetected is the native integrity verdict.
	FaultDetected bool
	// TestStatistic is the native chi-square statistic.
	TestStatistic float64
	// HasThreshold controls whether Threshold is valid.
	HasThreshold bool
	// Threshold is the native test threshold.
	Threshold float64
	// HasReducedChiSquare controls whether ReducedChiSquare is valid.
	HasReducedChiSquare bool
	// ReducedChiSquare is the native reduced statistic.
	ReducedChiSquare float64
	// RMSM is residual RMS in metres.
	RMSM float64
	// DOF is degrees of freedom.
	DOF int64
	// Testable reports whether enough rows were available.
	Testable bool
	// NormalizedResidualCount is the native row count.
	NormalizedResidualCount int
	// HasWorstSatellite controls whether WorstSatellite is valid.
	HasWorstSatellite bool
	// WorstSatellite identifies the largest normalized residual.
	WorstSatellite string
}

// RAIMNormalizedResidual is one detached normalized residual by satellite.
type RAIMNormalizedResidual struct {
	// SatelliteID identifies the residual row.
	SatelliteID string
	// NormalizedResidual is dimensionless.
	NormalizedResidual float64
}

// RAIM performs the C standalone receiver-autonomous integrity test.
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
