package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// StaticPositionEpoch is one pseudorange epoch in a static-position solve.
// Weights, when present, are positive multipliers aligned with Observations.
type StaticPositionEpoch struct {
	// Inputs contains the SPP V2 observations and receiver-time/model settings for this epoch.
	Inputs SPPInputsV2
	// Weights contains positive observation multipliers aligned with Inputs.Base.Observations.
	Weights []float64
}

// StaticPositionOptions controls the shared-position static solver.
type StaticPositionOptions struct {
	// InitialPositionM is the optional initial receiver position in ECEF metres.
	InitialPositionM [3]float64
	// WithGeodetic reports whether the solution includes a geodetic coordinate.
	WithGeodetic bool
	// RobustEnabled reports whether robust estimation is enabled.
	RobustEnabled bool
	// Robust contains Huber/IRLS scale, iteration, and tolerance settings when RobustEnabled is true.
	Robust SPPRobustConfig
}

// StaticPositionErrorKind is the solver's typed outcome category.
type StaticPositionErrorKind uint32

const (
	// StaticPositionErrorNone indicates that the operation produced no error.
	StaticPositionErrorNone StaticPositionErrorKind = 0
	// StaticPositionErrorEmptyEpochs reports that no epochs were supplied.
	StaticPositionErrorEmptyEpochs StaticPositionErrorKind = 1
	// StaticPositionErrorInvalidInput reports invalid static-position input.
	StaticPositionErrorInvalidInput StaticPositionErrorKind = 2
	// StaticPositionErrorEpochInput reports invalid epoch input.
	StaticPositionErrorEpochInput StaticPositionErrorKind = 3
	// StaticPositionErrorDuplicateObservation reports a duplicate observation.
	StaticPositionErrorDuplicateObservation StaticPositionErrorKind = 4
	// StaticPositionErrorIonosphereUnsupported reports unsupported ionospheric processing.
	StaticPositionErrorIonosphereUnsupported StaticPositionErrorKind = 5
	// StaticPositionErrorTooFewMeasurements reports too few measurements.
	StaticPositionErrorTooFewMeasurements StaticPositionErrorKind = 6
	// StaticPositionErrorEphemerisLost reports unavailable ephemeris.
	StaticPositionErrorEphemerisLost StaticPositionErrorKind = 7
	// StaticPositionErrorSingular reports singular solve geometry.
	StaticPositionErrorSingular StaticPositionErrorKind = 8
)

// StaticPositionInfluenceStatus describes a leave-one-out diagnostic.
type StaticPositionInfluenceStatus uint32

const (
	// StaticInfluenceSolved selects a successful influence solve.
	StaticInfluenceSolved StaticPositionInfluenceStatus = 0
	// StaticInfluenceTooFewMeasurements selects too few measurements for influence analysis.
	StaticInfluenceTooFewMeasurements StaticPositionInfluenceStatus = 1
	// StaticInfluenceSingularGeometry selects singular geometry during influence analysis.
	StaticInfluenceSingularGeometry StaticPositionInfluenceStatus = 2
	// StaticInfluenceInvalidInput selects invalid influence-analysis input.
	StaticInfluenceInvalidInput StaticPositionInfluenceStatus = 3
	// StaticInfluenceEphemerisUnavailable selects unavailable ephemeris during influence analysis.
	StaticInfluenceEphemerisUnavailable StaticPositionInfluenceStatus = 4
	// StaticInfluenceSolveFailed selects a failed influence solve.
	StaticInfluenceSolveFailed StaticPositionInfluenceStatus = 5
)

// StaticPositionSolveStatus is the native SPP termination status in static
// position metadata.
type StaticPositionSolveStatus uint32

const (
	// StaticPositionSolveGradientTolerance selects gradient-tolerance termination.
	StaticPositionSolveGradientTolerance StaticPositionSolveStatus = 0
	// StaticPositionSolveCostTolerance selects cost-tolerance termination.
	StaticPositionSolveCostTolerance StaticPositionSolveStatus = 1
	// StaticPositionSolveStepTolerance selects step-tolerance termination.
	StaticPositionSolveStepTolerance StaticPositionSolveStatus = 2
	// StaticPositionSolveMaxEvaluations selects maximum-evaluation termination.
	StaticPositionSolveMaxEvaluations StaticPositionSolveStatus = 3
)

// StaticPositionRejectionReason identifies why native SPP excluded a row.
type StaticPositionRejectionReason uint32

const (
	// StaticPositionRejectionNoEphemeris identifies rejection because ephemeris was unavailable.
	StaticPositionRejectionNoEphemeris StaticPositionRejectionReason = 0
	// StaticPositionRejectionLowElevation identifies rejection because elevation was too low.
	StaticPositionRejectionLowElevation StaticPositionRejectionReason = 1
	// StaticPositionRejectionSBASWithdrawn identifies rejection because SBAS corrections were withdrawn.
	StaticPositionRejectionSBASWithdrawn StaticPositionRejectionReason = 2
	// StaticPositionRejectionSBASIONOUncovered identifies rejection because the SBAS ionosphere was uncovered.
	StaticPositionRejectionSBASIONOUncovered StaticPositionRejectionReason = 3
)

// StaticPositionClockBias is one detached receiver clock estimate in seconds.
type StaticPositionClockBias struct {
	// EpochIndex is the zero-based input epoch whose receiver clock bias is reported.
	EpochIndex int
	// System identifies the GNSS constellation or constellation set.
	System GNSSSystem
	// ClockS is the receiver clock bias in seconds.
	ClockS float64
}

// StaticPositionEpochInfluence is a detached leave-one-epoch diagnostic.
type StaticPositionEpochInfluence struct {
	// EpochIndex is the zero-based omitted epoch; OmittedMeasurements is its measurement count.
	EpochIndex, OmittedMeasurements int
	// Status identifies whether the leave-one-epoch solve succeeded or why it failed.
	Status StaticPositionInfluenceStatus
	// HasPositionDelta reports whether PositionDeltaM is valid.
	HasPositionDelta bool
	// PositionDeltaM is the ECEF position change from the full solve in metres.
	PositionDeltaM [3]float64
	// PositionDeltaNormM is the Euclidean ECEF position-change norm in metres.
	PositionDeltaNormM float64
	// HasResidualRMS reports whether ResidualRMSM is valid.
	HasResidualRMS bool
	// ResidualRMSM is the post-fit residual RMS in metres.
	ResidualRMSM float64
	// MinRobustWeightRatio is the smallest dimensionless robust weight ratio in the influence solve.
	MinRobustWeightRatio float64
}

// StaticPositionMetadata contains detached static solver diagnostics.
type StaticPositionMetadata struct {
	// Iterations and OuterIterations are inner and robust outer solver iteration counts.
	// UsedMeasurements and Parameters are the retained-observation and fitted-parameter counts.
	Iterations, OuterIterations, UsedMeasurements, Parameters int
	// Converged reports solver convergence; HasFinalRobustScale guards FinalRobustScaleM.
	Converged, HasFinalRobustScale bool
	// Status identifies the native termination criterion.
	Status StaticPositionSolveStatus
	// FinalRobustScaleM contains metres.
	FinalRobustScaleM float64
	// Redundancy is the number of independent residual degrees of freedom.
	Redundancy int64
	// GeometryQuality contains rank, conditioning, dilution, and covariance checks.
	GeometryQuality SPPGeometryQuality
}

// StaticPositionResidual is one detached post-fit residual in metres.
type StaticPositionResidual struct {
	// EpochIndex is the zero-based epoch containing the residual.
	EpochIndex int
	// SatelliteID is the GNSS satellite identifier.
	SatelliteID string
	// ResidualM is the post-fit pseudorange residual in metres; BaseWeight is its dimensionless initial weight.
	ResidualM, BaseWeight float64
	// EffectiveWeight is the final dimensionless weight after robust reweighting.
	EffectiveWeight float64
	// RobustWeightRatio is the dimensionless effective-to-base weight ratio.
	RobustWeightRatio float64
}

// StaticPositionRejectedSatellite preserves an excluded satellite and reason.
type StaticPositionRejectedSatellite struct {
	// SatelliteID is the GNSS satellite identifier.
	SatelliteID string
	// Reason identifies why the native solver excluded this satellite row.
	Reason StaticPositionRejectionReason
}

// StaticPositionSatelliteBatchInfluence is a detached all-epochs diagnostic.
type StaticPositionSatelliteBatchInfluence struct {
	// SatelliteID is the GNSS satellite identifier.
	SatelliteID string
	// OmittedMeasurements is the number of rows removed for this satellite.
	OmittedMeasurements int
	// Status identifies the result of the all-epochs leave-one-satellite solve.
	Status StaticPositionInfluenceStatus
	// HasPositionDelta reports whether PositionDeltaM is valid.
	HasPositionDelta bool
	// PositionDeltaM is the ECEF position change from the full solve in metres.
	PositionDeltaM [3]float64
	// PositionDeltaNormM is the Euclidean ECEF position-change norm in metres.
	PositionDeltaNormM float64
	// HasResidualRMS reports whether ResidualRMSM is valid.
	HasResidualRMS bool
	// ResidualRMSM is the post-fit residual RMS in metres.
	ResidualRMSM float64
	// MinRobustWeightRatio is the smallest dimensionless robust weight ratio.
	MinRobustWeightRatio float64
}

// StaticPositionSatelliteInfluence is a detached per-epoch satellite diagnostic.
type StaticPositionSatelliteInfluence struct {
	// EpochIndex is the zero-based epoch containing the satellite row.
	EpochIndex int
	// SatelliteID is the GNSS satellite identifier.
	SatelliteID string
	// Status identifies the result of omitting this satellite from the epoch.
	Status StaticPositionInfluenceStatus
	// HasPositionDelta reports whether PositionDeltaM is valid.
	HasPositionDelta bool
	// PositionDeltaM is the ECEF position change from the full solve in metres.
	PositionDeltaM [3]float64
	// PositionDeltaNormM is the Euclidean ECEF position-change norm in metres.
	PositionDeltaNormM float64
	// HasResidualRMS reports whether ResidualRMSM is valid.
	HasResidualRMS bool
	// ResidualRMSM is the post-fit residual RMS in metres.
	ResidualRMSM float64
	// ResidualM is the row residual in metres; BaseWeight is its dimensionless initial weight.
	ResidualM, BaseWeight float64
	// EffectiveWeight is the final dimensionless weight after robust reweighting.
	EffectiveWeight float64
	// RobustWeightRatio is the dimensionless effective-to-base weight ratio.
	RobustWeightRatio float64
}

// StaticPositionSolution owns a native static-position result.
type StaticPositionSolution struct {
	_      noCopy
	handle *native.StaticPositionSolution
}

func nativeStaticEpochs(values []StaticPositionEpoch) ([]native.StaticPositionEpochInput, error) {
	result := make([]native.StaticPositionEpochInput, len(values))
	for i, value := range values {
		inputs, err := nativeSppV2(value.Inputs)
		if err != nil {
			return nil, err
		}
		result[i] = native.StaticPositionEpochInput{Inputs: inputs, Weights: append([]float64(nil), value.Weights...)}
	}
	return result, nil
}

func nativeStaticOptions(value *StaticPositionOptions) (*native.StaticPositionOptionsInput, error) {
	if value == nil {
		return nil, nil
	}
	robust, err := nativeRobustConfig(value.Robust)
	if err != nil {
		return nil, err
	}
	return &native.StaticPositionOptionsInput{InitialPositionM: value.InitialPositionM, WithGeodetic: value.WithGeodetic, RobustEnabled: value.RobustEnabled, Robust: robust}, nil
}

// DefaultStaticPositionOptions returns C's static-position defaults.
func DefaultStaticPositionOptions() (StaticPositionOptions, error) {
	value, err := native.StaticPositionOptionsInit()
	if err != nil {
		return StaticPositionOptions{}, publicError(err)
	}
	maxOuter, err := nativeCountToInt(value.Robust.MaxOuter, "static position robust max outer")
	if err != nil {
		return StaticPositionOptions{}, err
	}
	return StaticPositionOptions{InitialPositionM: value.InitialPositionM, WithGeodetic: value.WithGeodetic, RobustEnabled: value.RobustEnabled, Robust: SPPRobustConfig{HuberK: value.Robust.HuberK, ScaleFloorM: value.Robust.ScaleFloorM, MaxOuter: maxOuter, OuterToleranceM: value.Robust.OuterToleranceM}}, nil
}

func solveStaticPositionBroadcast(source *BroadcastEphemeris, epochs []StaticPositionEpoch, options *StaticPositionOptions) (*StaticPositionSolution, StaticPositionErrorKind, error) {
	if source == nil || source.handle == nil {
		return nil, 0, ErrClosed
	}
	cOptions, err := nativeStaticOptions(options)
	if err != nil {
		return nil, 0, err
	}
	nativeEpochs, err := nativeStaticEpochs(epochs)
	if err != nil {
		return nil, 0, err
	}
	h, kind, err := native.SolveStaticPositionBroadcast(source.handle, nativeEpochs, cOptions)
	if err != nil {
		return nil, StaticPositionErrorKind(kind), publicError(err)
	}
	return &StaticPositionSolution{handle: h}, StaticPositionErrorKind(kind), nil
}

// SolveStaticPositionBroadcast solves shared receiver position from broadcast ephemerides.
func SolveStaticPositionBroadcast(source *BroadcastEphemeris, epochs []StaticPositionEpoch, options *StaticPositionOptions) (*StaticPositionSolution, StaticPositionErrorKind, error) {
	return solveStaticPositionBroadcast(source, epochs, options)
}

// SolveStaticPositionSP3 solves shared receiver position from an SP3 product.
func SolveStaticPositionSP3(source *SP3, epochs []StaticPositionEpoch, options *StaticPositionOptions) (*StaticPositionSolution, StaticPositionErrorKind, error) {
	if source == nil || source.handle == nil {
		return nil, 0, ErrClosed
	}
	cOptions, err := nativeStaticOptions(options)
	if err != nil {
		return nil, 0, err
	}
	nativeEpochs, err := nativeStaticEpochs(epochs)
	if err != nil {
		return nil, 0, err
	}
	h, kind, err := native.SolveStaticPositionSP3(source.handle, nativeEpochs, cOptions)
	if err != nil {
		return nil, StaticPositionErrorKind(kind), publicError(err)
	}
	return &StaticPositionSolution{handle: h}, StaticPositionErrorKind(kind), nil
}

// Close releases the static-position result. It is idempotent and race-safe.
func (s *StaticPositionSolution) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return publicError(s.handle.Close())
}

// Position returns the ECEF receiver position in metres.
func (s *StaticPositionSolution) Position() ([3]float64, error) {
	if s == nil || s.handle == nil {
		return [3]float64{}, ErrClosed
	}
	v, err := s.handle.Position()
	return v, publicError(err)
}

// PositionCovarianceECEFM2 returns the 3x3 ECEF covariance in m².
func (s *StaticPositionSolution) PositionCovarianceECEFM2() ([9]float64, error) {
	if s == nil || s.handle == nil {
		return [9]float64{}, ErrClosed
	}
	v, err := s.handle.PositionCovarianceECEFM2()
	return v, publicError(err)
}

// PositionCovarianceENUM2 returns the 3x3 ENU covariance in m².
func (s *StaticPositionSolution) PositionCovarianceENUM2() ([9]float64, error) {
	if s == nil || s.handle == nil {
		return [9]float64{}, ErrClosed
	}
	v, err := s.handle.PositionCovarianceENUM2()
	return v, publicError(err)
}

// ClockBiases returns detached receiver clock estimates.
func (s *StaticPositionSolution) ClockBiases() ([]StaticPositionClockBias, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	values, err := s.handle.ClockBiases()
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]StaticPositionClockBias, len(values))
	for i, value := range values {
		result[i] = StaticPositionClockBias{EpochIndex: value.EpochIndex, System: GNSSSystem(value.System), ClockS: value.ClockS}
	}
	return result, nil
}

// EpochInfluence returns detached leave-one-epoch diagnostics.
func (s *StaticPositionSolution) EpochInfluence() ([]StaticPositionEpochInfluence, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	values, err := s.handle.EpochInfluence()
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]StaticPositionEpochInfluence, len(values))
	for i, v := range values {
		result[i] = StaticPositionEpochInfluence{EpochIndex: v.EpochIndex, OmittedMeasurements: v.OmittedMeasurements, Status: StaticPositionInfluenceStatus(v.Status), HasPositionDelta: v.HasPositionDelta, PositionDeltaM: v.PositionDeltaM, PositionDeltaNormM: v.PositionDeltaNormM, HasResidualRMS: v.HasResidualRMS, ResidualRMSM: v.ResidualRMSM, MinRobustWeightRatio: v.MinRobustWeightRatio}
	}
	return result, nil
}

// Geodetic returns the optional geodetic coordinate and its presence flag.
func (s *StaticPositionSolution) Geodetic() (Geodetic, bool, error) {
	if s == nil || s.handle == nil {
		return Geodetic{}, false, ErrClosed
	}
	v, present, err := s.handle.Geodetic()
	return Geodetic{LatitudeRad: v.LatitudeRad, LongitudeRad: v.LongitudeRad, HeightM: v.HeightM}, present, publicError(err)
}

// Metadata returns detached solver metadata.
func (s *StaticPositionSolution) Metadata() (StaticPositionMetadata, error) {
	if s == nil || s.handle == nil {
		return StaticPositionMetadata{}, ErrClosed
	}
	v, err := s.handle.Metadata()
	if err != nil {
		return StaticPositionMetadata{}, publicError(err)
	}
	return StaticPositionMetadata{Iterations: v.Iterations, OuterIterations: v.OuterIterations, UsedMeasurements: v.UsedMeasurements, Parameters: v.Parameters, Converged: v.Converged, HasFinalRobustScale: v.HasFinalRobustScale, Status: StaticPositionSolveStatus(v.Status), FinalRobustScaleM: v.FinalRobustScaleM, Redundancy: v.Redundancy, GeometryQuality: SPPGeometryQuality{Tier: v.GeometryQuality.Tier, Redundancy: v.GeometryQuality.Redundancy, Rank: v.GeometryQuality.Rank, ConditionNumber: v.GeometryQuality.ConditionNumber, GDOP: v.GeometryQuality.GDOP, RAIMCheckable: v.GeometryQuality.RAIMCheckable, CovarianceValidated: v.GeometryQuality.CovarianceValidated}}, nil
}

// RejectedSatellites returns detached excluded satellites for one epoch.
func (s *StaticPositionSolution) RejectedSatellites(epoch int) ([]StaticPositionRejectedSatellite, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	values, err := s.handle.RejectedSats(epoch)
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]StaticPositionRejectedSatellite, len(values))
	for i, v := range values {
		result[i] = StaticPositionRejectedSatellite{SatelliteID: v.SatelliteID, Reason: StaticPositionRejectionReason(v.Reason)}
	}
	return result, nil
}

// Residuals returns detached post-fit residual rows in metres.
func (s *StaticPositionSolution) Residuals() ([]StaticPositionResidual, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	values, err := s.handle.Residuals()
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]StaticPositionResidual, len(values))
	for i, v := range values {
		result[i] = StaticPositionResidual{EpochIndex: v.EpochIndex, SatelliteID: v.SatelliteID, ResidualM: v.ResidualM, BaseWeight: v.BaseWeight, EffectiveWeight: v.EffectiveWeight, RobustWeightRatio: v.RobustWeightRatio}
	}
	return result, nil
}

// SatelliteBatchInfluence returns detached all-epochs satellite diagnostics.
func (s *StaticPositionSolution) SatelliteBatchInfluence() ([]StaticPositionSatelliteBatchInfluence, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	values, err := s.handle.SatelliteBatchInfluence()
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]StaticPositionSatelliteBatchInfluence, len(values))
	for i, v := range values {
		result[i] = StaticPositionSatelliteBatchInfluence{SatelliteID: v.SatelliteID, OmittedMeasurements: v.OmittedMeasurements, Status: StaticPositionInfluenceStatus(v.Status), HasPositionDelta: v.HasPositionDelta, PositionDeltaM: v.PositionDeltaM, PositionDeltaNormM: v.PositionDeltaNormM, HasResidualRMS: v.HasResidualRMS, ResidualRMSM: v.ResidualRMSM, MinRobustWeightRatio: v.MinRobustWeightRatio}
	}
	return result, nil
}

// SatelliteInfluence returns detached per-epoch satellite diagnostics.
func (s *StaticPositionSolution) SatelliteInfluence() ([]StaticPositionSatelliteInfluence, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	values, err := s.handle.SatelliteInfluence()
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]StaticPositionSatelliteInfluence, len(values))
	for i, v := range values {
		result[i] = StaticPositionSatelliteInfluence{EpochIndex: v.EpochIndex, SatelliteID: v.SatelliteID, Status: StaticPositionInfluenceStatus(v.Status), HasPositionDelta: v.HasPositionDelta, PositionDeltaM: v.PositionDeltaM, PositionDeltaNormM: v.PositionDeltaNormM, HasResidualRMS: v.HasResidualRMS, ResidualRMSM: v.ResidualRMSM, ResidualM: v.ResidualM, BaseWeight: v.BaseWeight, EffectiveWeight: v.EffectiveWeight, RobustWeightRatio: v.RobustWeightRatio}
	}
	return result, nil
}

// StateCovarianceM2 returns the detached row-major 3x3 ECEF position covariance in square metres.
func (s *StaticPositionSolution) StateCovarianceM2() ([]float64, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	v, err := s.handle.StateCovarianceM2()
	return append([]float64(nil), v...), publicError(err)
}

// StaticReferenceStationRinexConfig selects static reference-station modes.
type StaticReferenceStationRinexConfig struct {
	// ReferencePositionM is the fixed reference-station ECEF position in metres.
	ReferencePositionM [3]float64
	// EnableCodeDGNSS reports whether code-DGNSS processing is enabled.
	EnableCodeDGNSS bool
	// EnableCarrierRTK reports whether carrier-RTK processing is enabled.
	EnableCarrierRTK bool
	// WithGeodetic reports whether the solution includes a geodetic coordinate.
	WithGeodetic bool
	// Carrier contains the RTK static-baseline configuration used when carrier processing is enabled.
	Carrier RTKRINEXStaticBaselineConfig
}

// StaticReferenceStationMode identifies the static-reference processing mode.
type StaticReferenceStationMode uint32

const (
	// StaticReferenceModeCodeDGNSS selects code-based DGNSS processing.
	StaticReferenceModeCodeDGNSS StaticReferenceStationMode = 0
	// StaticReferenceModeCarrierFloat selects carrier-phase float processing.
	StaticReferenceModeCarrierFloat StaticReferenceStationMode = 1
	// StaticReferenceModeCarrierFixed selects carrier-phase fixed processing.
	StaticReferenceModeCarrierFixed StaticReferenceStationMode = 2
)

// StaticReferenceFixStatus identifies the fix status of a static-reference solution.
type StaticReferenceFixStatus uint32

const (
	// StaticReferenceFixCodeDGNSS selects a code-based DGNSS fix.
	StaticReferenceFixCodeDGNSS StaticReferenceFixStatus = 0
	// StaticReferenceFixCarrierFloat selects a carrier-phase float fix.
	StaticReferenceFixCarrierFloat StaticReferenceFixStatus = 1
	// StaticReferenceFixCarrierFixed selects a carrier-phase fixed fix.
	StaticReferenceFixCarrierFixed StaticReferenceFixStatus = 2
)

// StaticReferenceModeStatus identifies an attempted reference-station mode.
type StaticReferenceModeStatus uint32

const (
	// StaticReferenceModeSolved selects a solved reference-station mode.
	StaticReferenceModeSolved StaticReferenceModeStatus = 0
	// StaticReferenceModeFailed selects a failed reference-station mode.
	StaticReferenceModeFailed StaticReferenceModeStatus = 1
)

// StaticReferenceEpochDiagnostic is one detached reference-station row.
type StaticReferenceEpochDiagnostic struct {
	// Mode identifies the attempted code-DGNSS, carrier-float, or carrier-fixed mode.
	Mode StaticReferenceStationMode
	// EpochIndex is the zero-based RINEX epoch; the two counts are used and rejected satellites.
	EpochIndex, UsedSatelliteCount, RejectedSatelliteCount int
	// HasCodeResidualRMS, HasPhaseResidualRMS, and HasResidualRMS guard the code, phase, and combined RMS fields below.
	HasCodeResidualRMS, HasPhaseResidualRMS, HasResidualRMS bool
	// CodeResidualRMSM, PhaseResidualRMSM, and ResidualRMSM are code, phase, and combined residual RMS values in metres.
	CodeResidualRMSM, PhaseResidualRMSM, ResidualRMSM float64
}

// StaticReferenceStationMetadata contains detached reference-station metadata.
type StaticReferenceStationMetadata struct {
	// Mode and FixStatus identify the selected processing mode and its code/float/fixed outcome.
	Mode      StaticReferenceStationMode
	FixStatus StaticReferenceFixStatus
	// HasGeodetic guards the returned geodetic coordinate.
	HasGeodetic bool
	// Geodetic is the station coordinate when HasGeodetic is true.
	Geodetic Geodetic
	// BaselineM is the rover-minus-reference baseline length in metres.
	BaselineM float64
	// HasCodeSolution and HasCarrierSolution report which nested solve families produced results.
	HasCodeSolution, HasCarrierSolution bool
	// DiagnosticCount and ModeReportCount count epoch diagnostics and attempted-mode reports.
	DiagnosticCount, ModeReportCount int
	// CarrierIntegerStatus identifies the carrier ambiguity resolution outcome.
	CarrierIntegerStatus RTKIntegerStatus
	// HasCarrierIntegerRatio reports whether CarrierIntegerRatio is valid.
	HasCarrierIntegerRatio bool
	// CarrierIntegerRatio is the carrier ambiguity ratio when HasCarrierIntegerRatio is true.
	CarrierIntegerRatio float64
	// CodeDiagnosticCount and CarrierDiagnosticCount count the mode-specific diagnostic rows.
	CodeDiagnosticCount, CarrierDiagnosticCount int
}

// StaticReferenceModeReport records one attempted static mode.
type StaticReferenceModeReport struct {
	// Mode identifies the attempted reference-station processing mode.
	Mode StaticReferenceStationMode
	// Status reports whether that mode produced a solution.
	Status StaticReferenceModeStatus
	// UsedEpochs, SkippedEpochs, and UsedMeasurements count accepted epochs, rejected epochs, and retained measurement rows.
	UsedEpochs, SkippedEpochs, UsedMeasurements int
	// HasError reports whether an error was recorded.
	HasError bool
}

// StaticReferenceStationSolution owns a native static reference-station result.
type StaticReferenceStationSolution struct {
	_      noCopy
	handle *native.StaticReferenceStationSolution
}

// DefaultStaticReferenceStationRinexConfig returns C's defaults.
func DefaultStaticReferenceStationRinexConfig() (StaticReferenceStationRinexConfig, error) {
	v, err := native.StaticReferenceStationRinexConfigInit()
	return StaticReferenceStationRinexConfig{ReferencePositionM: v.ReferencePositionM, EnableCodeDGNSS: v.EnableCodeDGNSS, EnableCarrierRTK: v.EnableCarrierRTK, WithGeodetic: v.WithGeodetic, Carrier: publicRTKRINEXStaticConfig(v.Carrier)}, publicError(err)
}

// SolveStaticReferenceStationRINEX solves a static reference-station coordinate.
func SolveStaticReferenceStationRINEX(sp3 *SP3, reference, rover *RINEXObservation, config StaticReferenceStationRinexConfig) (*StaticReferenceStationSolution, error) {
	if sp3 == nil || sp3.handle == nil || reference == nil || reference.handle == nil || rover == nil || rover.handle == nil {
		return nil, ErrClosed
	}
	h, err := native.SolveStaticReferenceStationRinex(nativeSP3(sp3), nativeRinexObs(reference), nativeRinexObs(rover), native.StaticReferenceStationRinexConfigInput{ReferencePositionM: config.ReferencePositionM, EnableCodeDGNSS: config.EnableCodeDGNSS, EnableCarrierRTK: config.EnableCarrierRTK, WithGeodetic: config.WithGeodetic, Carrier: nativeRTKRINEXStaticConfig(config.Carrier)})
	if err != nil {
		return nil, publicError(err)
	}
	return &StaticReferenceStationSolution{handle: h}, nil
}

// Close releases the reference-station result. It is idempotent and race-safe.
func (s *StaticReferenceStationSolution) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return publicError(s.handle.Close())
}

// BaselineECEF returns rover-minus-reference ECEF metres.
func (s *StaticReferenceStationSolution) BaselineECEF() ([3]float64, error) {
	if s == nil || s.handle == nil {
		return [3]float64{}, ErrClosed
	}
	v, err := s.handle.BaselineECEF()
	return v, publicError(err)
}

// PositionECEF returns the selected rover ECEF coordinate in metres.
func (s *StaticReferenceStationSolution) PositionECEF() ([3]float64, error) {
	if s == nil || s.handle == nil {
		return [3]float64{}, ErrClosed
	}
	v, err := s.handle.PositionECEF()
	return v, publicError(err)
}

// CovarianceECEF returns the selected 3x3 covariance in m².
func (s *StaticReferenceStationSolution) CovarianceECEF() ([9]float64, error) {
	if s == nil || s.handle == nil {
		return [9]float64{}, ErrClosed
	}
	v, err := s.handle.CovarianceECEF()
	return v, publicError(err)
}

// CovarianceENU returns the selected ENU covariance in m².
func (s *StaticReferenceStationSolution) CovarianceENU() ([9]float64, error) {
	if s == nil || s.handle == nil {
		return [9]float64{}, ErrClosed
	}
	v, err := s.handle.CovarianceENU()
	return v, publicError(err)
}

// Diagnostics returns detached selected-mode diagnostics.
func (s *StaticReferenceStationSolution) Diagnostics() ([]StaticReferenceEpochDiagnostic, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	v, err := s.handle.Diagnostics()
	if err != nil {
		return nil, publicError(err)
	}
	r := make([]StaticReferenceEpochDiagnostic, len(v))
	for i, x := range v {
		r[i] = StaticReferenceEpochDiagnostic{Mode: StaticReferenceStationMode(x.Mode), EpochIndex: x.EpochIndex, UsedSatelliteCount: x.UsedSatelliteCount, RejectedSatelliteCount: x.RejectedSatelliteCount, HasCodeResidualRMS: x.HasCodeResidualRMS, CodeResidualRMSM: x.CodeResidualRMSM, HasPhaseResidualRMS: x.HasPhaseResidualRMS, PhaseResidualRMSM: x.PhaseResidualRMSM, HasResidualRMS: x.HasResidualRMS, ResidualRMSM: x.ResidualRMSM}
	}
	return r, nil
}

// Metadata returns detached reference-station metadata.
func (s *StaticReferenceStationSolution) Metadata() (StaticReferenceStationMetadata, error) {
	if s == nil || s.handle == nil {
		return StaticReferenceStationMetadata{}, ErrClosed
	}
	v, err := s.handle.Metadata()
	if err != nil {
		return StaticReferenceStationMetadata{}, publicError(err)
	}
	return StaticReferenceStationMetadata{Mode: StaticReferenceStationMode(v.Mode), FixStatus: StaticReferenceFixStatus(v.FixStatus), HasGeodetic: v.HasGeodetic, Geodetic: Geodetic{LatitudeRad: v.Geodetic.LatitudeRad, LongitudeRad: v.Geodetic.LongitudeRad, HeightM: v.Geodetic.HeightM}, BaselineM: v.BaselineM, HasCodeSolution: v.HasCodeSolution, HasCarrierSolution: v.HasCarrierSolution, DiagnosticCount: v.DiagnosticCount, ModeReportCount: v.ModeReportCount, CarrierIntegerStatus: RTKIntegerStatus(v.CarrierIntegerStatus), HasCarrierIntegerRatio: v.HasCarrierIntegerRatio, CarrierIntegerRatio: v.CarrierIntegerRatio, CodeDiagnosticCount: v.CodeDiagnosticCount, CarrierDiagnosticCount: v.CarrierDiagnosticCount}, nil
}

// ModeReports returns detached per-mode reports.
func (s *StaticReferenceStationSolution) ModeReports() ([]StaticReferenceModeReport, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	v, err := s.handle.ModeReports()
	if err != nil {
		return nil, publicError(err)
	}
	r := make([]StaticReferenceModeReport, len(v))
	for i, x := range v {
		r[i] = StaticReferenceModeReport{Mode: StaticReferenceStationMode(x.Mode), Status: StaticReferenceModeStatus(x.Status), UsedEpochs: x.UsedEpochs, SkippedEpochs: x.SkippedEpochs, UsedMeasurements: x.UsedMeasurements, HasError: x.HasError}
	}
	return r, nil
}
