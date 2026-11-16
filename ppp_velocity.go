package sidereon

import "sidereon.dev/go/v4/internal/native"

// PPPVelocityRobustConfig controls Huber/MAD reweighting. Values are comparable
// with Go's == operator.
type PPPVelocityRobustConfig struct {
	// HuberK is the Huber tuning constant.
	HuberK float64
	// ScaleFloorMS is the floor on the MAD residual scale in metres per second.
	ScaleFloorMS float64
	// MaxOuter is the maximum total number of outer solves.
	MaxOuter int
	// OuterToleranceMS is the outer-loop state-step tolerance in metres per second.
	OuterToleranceMS float64
}

// PPPVelocityConfig controls the low-level precise-positioning velocity solve.
// A value with a nil Robust field is comparable to another value by comparing
// its fields; non-nil Robust values can be compared by dereferencing the field.
type PPPVelocityConfig struct {
	// MinimumObservations is the requested minimum observation count; the engine
	// always requires at least four.
	MinimumObservations int
	// Robust enables robust reweighting when non-nil.
	Robust *PPPVelocityRobustConfig
}

// PPPReceiverVelocityState is a receiver state for range-rate prediction.
// Values are comparable with Go's == operator.
type PPPReceiverVelocityState struct {
	// PositionM is the receiver ECEF position in metres.
	PositionM [3]float64
	// VelocityMS is the receiver ECEF velocity in metres per second.
	VelocityMS [3]float64
	// ClockDriftMS is receiver clock drift as a range-rate bias in metres per second.
	ClockDriftMS float64
}

// PPPVelocityObservation is one satellite range-rate observation. Values are
// comparable with Go's == operator.
type PPPVelocityObservation struct {
	// SatelliteID is the satellite token, for example G08.
	SatelliteID string
	// SatellitePositionM is the satellite ECEF position in metres.
	SatellitePositionM [3]float64
	// SatelliteVelocityMS is the satellite ECEF velocity in metres per second.
	SatelliteVelocityMS [3]float64
	// MeasuredRangeRateMS is measured pseudorange rate in metres per second.
	MeasuredRangeRateMS float64
	// SigmaMS is the one-sigma range-rate uncertainty in metres per second.
	SigmaMS float64
	// SatelliteClockDriftMS is satellite clock drift as a range-rate bias.
	SatelliteClockDriftMS float64
}

// PPPRangeRatePrediction contains predicted range-rate geometry. Values are
// comparable with Go's == operator.
type PPPRangeRatePrediction struct {
	// LOSUnit is the unit line-of-sight vector from receiver to satellite.
	LOSUnit [3]float64
	// RangeRateMS is the predicted range rate in metres per second.
	RangeRateMS float64
}

// PPPVelocitySolutionValues contains detached velocity-solve results. Values
// are comparable with Go's == operator.
type PPPVelocitySolutionValues struct {
	// VelocityMS is receiver ECEF velocity in metres per second.
	VelocityMS [3]float64
	// ClockDriftMS is receiver clock drift in metres per second.
	ClockDriftMS float64
	// ResidualRMSMS is residual RMS in metres per second.
	ResidualRMSMS float64
	// RobustIterations is the number of robust outer iterations.
	RobustIterations int
	// HasRobustScale reports whether RobustScaleMS is present.
	HasRobustScale bool
	// RobustScaleMS is the final robust residual scale in metres per second.
	RobustScaleMS float64
}

// PPPVelocitySolution owns a native precise-positioning velocity solution.
type PPPVelocitySolution struct {
	handle *native.PppVelocitySolution
}

// Equal compares two robust configurations through the engine.
func (value PPPVelocityRobustConfig) Equal(other PPPVelocityRobustConfig) (bool, error) {
	result, err := native.PppVelocityRobustConfigEqual(nativePPPVelocityRobustConfig(value), nativePPPVelocityRobustConfig(other))
	return result, publicError(err)
}

// Equal compares two velocity configurations through the engine.
func (value PPPVelocityConfig) Equal(other PPPVelocityConfig) (bool, error) {
	result, err := native.PppVelocityConfigEqual(nativePPPVelocityConfig(value), nativePPPVelocityConfig(other))
	return result, publicError(err)
}

// Equal compares two receiver states through the engine.
func (value PPPReceiverVelocityState) Equal(other PPPReceiverVelocityState) (bool, error) {
	result, err := native.PppReceiverVelocityStateEqual(nativePPPReceiverVelocityState(value), nativePPPReceiverVelocityState(other))
	return result, publicError(err)
}

// Equal compares two velocity observations through the engine.
func (value PPPVelocityObservation) Equal(other PPPVelocityObservation) (bool, error) {
	result, err := native.PppVelocityObservationEqual(nativePPPVelocityObservation(value), nativePPPVelocityObservation(other))
	return result, publicError(err)
}

// Equal compares two range-rate predictions through the engine.
func (value PPPRangeRatePrediction) Equal(other PPPRangeRatePrediction) (bool, error) {
	result, err := native.PppRangeRatePredictionEqual(native.PppRangeRatePrediction{LOSUnit: value.LOSUnit, RangeRateMS: value.RangeRateMS}, native.PppRangeRatePrediction{LOSUnit: other.LOSUnit, RangeRateMS: other.RangeRateMS})
	return result, publicError(err)
}

func nativePPPVelocityRobustConfig(value PPPVelocityRobustConfig) native.PppVelocityRobustConfig {
	return native.PppVelocityRobustConfig{HuberK: value.HuberK, ScaleFloorMS: value.ScaleFloorMS, MaxOuter: value.MaxOuter, OuterTolMS: value.OuterToleranceMS}
}

func nativePPPVelocityConfig(value PPPVelocityConfig) native.PppVelocityConfig {
	result := native.PppVelocityConfig{MinimumObservations: value.MinimumObservations}
	if value.Robust != nil {
		robust := nativePPPVelocityRobustConfig(*value.Robust)
		result.Robust = &robust
	}
	return result
}

func nativePPPVelocityObservation(value PPPVelocityObservation) native.PppVelocityObservation {
	return native.PppVelocityObservation{SatelliteID: value.SatelliteID, SatellitePositionM: value.SatellitePositionM, SatelliteVelocityMS: value.SatelliteVelocityMS, MeasuredRangeRateMS: value.MeasuredRangeRateMS, SigmaMS: value.SigmaMS, SatelliteClockDriftMS: value.SatelliteClockDriftMS}
}

func nativePPPReceiverVelocityState(value PPPReceiverVelocityState) native.PppReceiverVelocityState {
	return native.PppReceiverVelocityState{PositionM: value.PositionM, VelocityMS: value.VelocityMS, ClockDriftMS: value.ClockDriftMS}
}

// DefaultPPPVelocityRobustConfig returns the engine's robust velocity defaults.
func DefaultPPPVelocityRobustConfig() (PPPVelocityRobustConfig, error) {
	value, err := native.PppVelocityRobustConfigInit()
	return PPPVelocityRobustConfig{HuberK: value.HuberK, ScaleFloorMS: value.ScaleFloorMS, MaxOuter: value.MaxOuter, OuterToleranceMS: value.OuterTolMS}, publicError(err)
}

// DefaultPPPVelocityConfig returns the engine's velocity-solve defaults.
func DefaultPPPVelocityConfig() (PPPVelocityConfig, error) {
	value, err := native.PppVelocityConfigInit()
	result := PPPVelocityConfig{MinimumObservations: value.MinimumObservations}
	if value.Robust != nil {
		result.Robust = &PPPVelocityRobustConfig{HuberK: value.Robust.HuberK, ScaleFloorMS: value.Robust.ScaleFloorMS, MaxOuter: value.Robust.MaxOuter, OuterToleranceMS: value.Robust.OuterTolMS}
	}
	return result, publicError(err)
}

// NewStationaryPPPReceiverVelocityState returns a stationary receiver state at positionM.
func NewStationaryPPPReceiverVelocityState(positionM [3]float64) (PPPReceiverVelocityState, error) {
	value, err := native.PppReceiverVelocityStateStationaryAt(positionM)
	return PPPReceiverVelocityState{PositionM: value.PositionM, VelocityMS: value.VelocityMS, ClockDriftMS: value.ClockDriftMS}, publicError(err)
}

// PredictPPPRangeRateMS predicts range rate for one observation and receiver state.
func PredictPPPRangeRateMS(observation PPPVelocityObservation, receiver PPPReceiverVelocityState) (PPPRangeRatePrediction, error) {
	value, err := native.PppPredictRangeRateMS(nativePPPVelocityObservation(observation), nativePPPReceiverVelocityState(receiver))
	return PPPRangeRatePrediction{LOSUnit: value.LOSUnit, RangeRateMS: value.RangeRateMS}, publicError(err)
}

// SolvePPPVelocityLeastSquares solves receiver velocity and clock drift from satellite states.
func SolvePPPVelocityLeastSquares(observations []PPPVelocityObservation, receiverPositionM [3]float64, config PPPVelocityConfig) (*PPPVelocitySolution, error) {
	values := make([]native.PppVelocityObservation, len(observations))
	for index, observation := range observations {
		values[index] = nativePPPVelocityObservation(observation)
	}
	handle, err := native.SolvePppVelocityLeastSquares(values, receiverPositionM, nativePPPVelocityConfig(config))
	if err != nil {
		return nil, publicError(err)
	}
	return &PPPVelocitySolution{handle: handle}, nil
}

// Close releases the velocity solution and is safe to call repeatedly.
func (solution *PPPVelocitySolution) Close() error {
	if solution == nil || solution.handle == nil {
		return nil
	}
	return publicError(solution.handle.Close())
}

// Values returns detached scalar and vector velocity-solve results.
func (solution *PPPVelocitySolution) Values() (PPPVelocitySolutionValues, error) {
	if solution == nil || solution.handle == nil {
		return PPPVelocitySolutionValues{}, ErrClosed
	}
	value, err := solution.handle.Values()
	return PPPVelocitySolutionValues{VelocityMS: value.VelocityMS, ClockDriftMS: value.ClockDriftMS, ResidualRMSMS: value.ResidualRMSMS, RobustIterations: value.RobustIterations, HasRobustScale: value.HasRobustScale, RobustScaleMS: value.RobustScaleMS}, publicError(err)
}

// UsedSatelliteIDs returns satellite tokens retained by the velocity solve.
func (solution *PPPVelocitySolution) UsedSatelliteIDs() ([]string, error) {
	if solution == nil || solution.handle == nil {
		return nil, ErrClosed
	}
	values, err := solution.handle.UsedSatelliteIDs()
	return values, publicError(err)
}

// DownweightedSatelliteIDs returns satellite tokens downweighted by robust solving.
func (solution *PPPVelocitySolution) DownweightedSatelliteIDs() ([]string, error) {
	if solution == nil || solution.handle == nil {
		return nil, ErrClosed
	}
	values, err := solution.handle.DownweightedSatelliteIDs()
	return values, publicError(err)
}

// Equal compares two velocity solutions by their complete engine values.
func (solution *PPPVelocitySolution) Equal(other *PPPVelocitySolution) (bool, error) {
	if solution == nil || solution.handle == nil || other == nil || other.handle == nil {
		return false, ErrClosed
	}
	equal, err := solution.handle.Equal(other.handle)
	return equal, publicError(err)
}
