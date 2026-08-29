package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// StaticPositionEpoch is one pseudorange epoch in a static-position solve.
// Weights, when present, are positive multipliers aligned with Observations.
type StaticPositionEpoch struct {
	Inputs  SPPInputsV2
	Weights []float64
}

// StaticPositionOptions controls the shared-position static solver.
type StaticPositionOptions struct {
	InitialPositionM [3]float64
	WithGeodetic     bool
	RobustEnabled    bool
	Robust           SPPRobustConfig
}

// StaticPositionErrorKind is the solver's typed outcome category.
type StaticPositionErrorKind uint32

const (
	StaticPositionErrorNone                  StaticPositionErrorKind = 0
	StaticPositionErrorEmptyEpochs           StaticPositionErrorKind = 1
	StaticPositionErrorInvalidInput          StaticPositionErrorKind = 2
	StaticPositionErrorEpochInput            StaticPositionErrorKind = 3
	StaticPositionErrorDuplicateObservation  StaticPositionErrorKind = 4
	StaticPositionErrorIonosphereUnsupported StaticPositionErrorKind = 5
	StaticPositionErrorTooFewMeasurements    StaticPositionErrorKind = 6
	StaticPositionErrorEphemerisLost         StaticPositionErrorKind = 7
	StaticPositionErrorSingular              StaticPositionErrorKind = 8
)

// StaticPositionInfluenceStatus describes a leave-one-out diagnostic.
type StaticPositionInfluenceStatus uint32

const (
	StaticInfluenceSolved               StaticPositionInfluenceStatus = 0
	StaticInfluenceTooFewMeasurements   StaticPositionInfluenceStatus = 1
	StaticInfluenceSingularGeometry     StaticPositionInfluenceStatus = 2
	StaticInfluenceInvalidInput         StaticPositionInfluenceStatus = 3
	StaticInfluenceEphemerisUnavailable StaticPositionInfluenceStatus = 4
	StaticInfluenceSolveFailed          StaticPositionInfluenceStatus = 5
)

// StaticPositionSolveStatus is the native SPP termination status in static
// position metadata.
type StaticPositionSolveStatus uint32

const (
	StaticPositionSolveGradientTolerance StaticPositionSolveStatus = 0
	StaticPositionSolveCostTolerance     StaticPositionSolveStatus = 1
	StaticPositionSolveStepTolerance     StaticPositionSolveStatus = 2
	StaticPositionSolveMaxEvaluations    StaticPositionSolveStatus = 3
)

// StaticPositionRejectionReason identifies why native SPP excluded a row.
type StaticPositionRejectionReason uint32

const (
	StaticPositionRejectionNoEphemeris       StaticPositionRejectionReason = 0
	StaticPositionRejectionLowElevation      StaticPositionRejectionReason = 1
	StaticPositionRejectionSBASWithdrawn     StaticPositionRejectionReason = 2
	StaticPositionRejectionSBASIONOUncovered StaticPositionRejectionReason = 3
)

// StaticPositionClockBias is one detached receiver clock estimate in seconds.
type StaticPositionClockBias struct {
	EpochIndex int
	System     GNSSSystem
	ClockS     float64
}

// StaticPositionEpochInfluence is a detached leave-one-epoch diagnostic.
type StaticPositionEpochInfluence struct {
	EpochIndex, OmittedMeasurements int
	Status                          StaticPositionInfluenceStatus
	HasPositionDelta                bool
	PositionDeltaM                  [3]float64
	PositionDeltaNormM              float64
	HasResidualRMS                  bool
	ResidualRMSM                    float64
	MinRobustWeightRatio            float64
}

// StaticPositionMetadata contains detached static solver diagnostics.
type StaticPositionMetadata struct {
	Iterations, OuterIterations, UsedMeasurements, Parameters int
	Converged, HasFinalRobustScale                            bool
	Status                                                    StaticPositionSolveStatus
	FinalRobustScaleM                                         float64
	Redundancy                                                int64
	GeometryQuality                                           SPPGeometryQuality
}

// StaticPositionResidual is one detached post-fit residual in metres.
type StaticPositionResidual struct {
	EpochIndex            int
	SatelliteID           string
	ResidualM, BaseWeight float64
	EffectiveWeight       float64
	RobustWeightRatio     float64
}

// StaticPositionRejectedSatellite preserves an excluded satellite and reason.
type StaticPositionRejectedSatellite struct {
	SatelliteID string
	Reason      StaticPositionRejectionReason
}

// StaticPositionSatelliteBatchInfluence is a detached all-epochs diagnostic.
type StaticPositionSatelliteBatchInfluence struct {
	SatelliteID          string
	OmittedMeasurements  int
	Status               StaticPositionInfluenceStatus
	HasPositionDelta     bool
	PositionDeltaM       [3]float64
	PositionDeltaNormM   float64
	HasResidualRMS       bool
	ResidualRMSM         float64
	MinRobustWeightRatio float64
}

// StaticPositionSatelliteInfluence is a detached per-epoch satellite diagnostic.
type StaticPositionSatelliteInfluence struct {
	EpochIndex            int
	SatelliteID           string
	Status                StaticPositionInfluenceStatus
	HasPositionDelta      bool
	PositionDeltaM        [3]float64
	PositionDeltaNormM    float64
	HasResidualRMS        bool
	ResidualRMSM          float64
	ResidualM, BaseWeight float64
	EffectiveWeight       float64
	RobustWeightRatio     float64
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

// StateCovarianceM2 returns detached state covariance entries in m².
func (s *StaticPositionSolution) StateCovarianceM2() ([]float64, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	v, err := s.handle.StateCovarianceM2()
	return append([]float64(nil), v...), publicError(err)
}

// StaticReferenceStationRinexConfig selects static reference-station modes.
type StaticReferenceStationRinexConfig struct {
	ReferencePositionM [3]float64
	EnableCodeDGNSS    bool
	EnableCarrierRTK   bool
	WithGeodetic       bool
	Carrier            RTKRINEXStaticBaselineConfig
}

// StaticReferenceStationMode identifies the selected reference-station mode.
type StaticReferenceStationMode uint32

const (
	StaticReferenceModeCodeDGNSS    StaticReferenceStationMode = 0
	StaticReferenceModeCarrierFloat StaticReferenceStationMode = 1
	StaticReferenceModeCarrierFixed StaticReferenceStationMode = 2
)

// StaticReferenceFixStatus identifies the selected reference-station fix.
type StaticReferenceFixStatus uint32

const (
	StaticReferenceFixCodeDGNSS    StaticReferenceFixStatus = 0
	StaticReferenceFixCarrierFloat StaticReferenceFixStatus = 1
	StaticReferenceFixCarrierFixed StaticReferenceFixStatus = 2
)

// StaticReferenceModeStatus identifies an attempted reference-station mode.
type StaticReferenceModeStatus uint32

const (
	StaticReferenceModeSolved StaticReferenceModeStatus = 0
	StaticReferenceModeFailed StaticReferenceModeStatus = 1
)

// StaticReferenceEpochDiagnostic is one detached reference-station row.
type StaticReferenceEpochDiagnostic struct {
	Mode                                                    StaticReferenceStationMode
	EpochIndex, UsedSatelliteCount, RejectedSatelliteCount  int
	HasCodeResidualRMS, HasPhaseResidualRMS, HasResidualRMS bool
	CodeResidualRMSM, PhaseResidualRMSM, ResidualRMSM       float64
}

// StaticReferenceStationMetadata contains detached reference-station metadata.
type StaticReferenceStationMetadata struct {
	Mode                                        StaticReferenceStationMode
	FixStatus                                   StaticReferenceFixStatus
	HasGeodetic                                 bool
	Geodetic                                    Geodetic
	BaselineM                                   float64
	HasCodeSolution, HasCarrierSolution         bool
	DiagnosticCount, ModeReportCount            int
	CarrierIntegerStatus                        RTKIntegerStatus
	HasCarrierIntegerRatio                      bool
	CarrierIntegerRatio                         float64
	CodeDiagnosticCount, CarrierDiagnosticCount int
}

// StaticReferenceModeReport records one attempted static mode.
type StaticReferenceModeReport struct {
	Mode                                        StaticReferenceStationMode
	Status                                      StaticReferenceModeStatus
	UsedEpochs, SkippedEpochs, UsedMeasurements int
	HasError                                    bool
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
