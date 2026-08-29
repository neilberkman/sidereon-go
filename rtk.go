package sidereon

import (
	"github.com/neilberkman/sidereon-go/internal/native"
)

// RTKMeasurementModel selects the stochastic model for RTK code and carrier measurements.
type RTKMeasurementModel struct {
	// CodeSigmaM and PhaseSigmaM are code and carrier-phase standard deviations in metres.
	CodeSigmaM, PhaseSigmaM float64
	// Sagnac reports whether the Sagnac correction is enabled.
	Sagnac     bool
	Stochastic uint32
	// ElevationWeighting reports whether elevation weighting is enabled.
	ElevationWeighting bool
}

// RTKResidualValidationOptions configures residual validation for RTK solutions.
type RTKResidualValidationOptions struct {
	// ThresholdSigmaEnabled reports whether ThresholdSigma is enabled.
	ThresholdSigmaEnabled bool
	// ThresholdSigma is the residual-rejection threshold in standard deviations.
	ThresholdSigma float64
	// MaxExclusions is the maximum exclusion count.
	MaxExclusions int
}

// RTKFloatOptions configures iteration limits and tolerances for the RTK float solver.
type RTKFloatOptions struct {
	// PositionTolM and AmbiguityTolM are position and ambiguity tolerances in metres.
	PositionTolM, AmbiguityTolM float64
	MaxIterations               int
}

// RTKFixedOptions configures integer ambiguity fixing and ratio-test thresholds.
type RTKFixedOptions struct {
	// PositionTolM and AmbiguityTolM are position and ambiguity tolerances in metres.
	PositionTolM, AmbiguityTolM float64
	MaxIterations               int
	RatioThreshold              float64
	// PartialAmbiguityResolution reports whether partial ambiguity resolution is enabled.
	PartialAmbiguityResolution bool
	// PartialMinAmbiguities is the minimum ambiguities for partial fixing.
	PartialMinAmbiguities int
}

// RTKArcUpdateOptions configures iterative RTK arc updates and ratio testing.
type RTKArcUpdateOptions struct {
	// HoldSigmaM, PositionTolM, and AmbiguityTolM are RTK update tolerances in metres.
	HoldSigmaM, PositionTolM, AmbiguityTolM float64
	MaxIterations                           int
	// ProcessNoiseBaselineSigmaM contains metres.
	ProcessNoiseBaselineSigmaM float64
	// DynamicsVelocityPropagated reports whether velocity is propagated by the dynamics model.
	DynamicsVelocityPropagated bool
	// FloatOnlySystems identifies the GNSS constellation or constellation set.
	FloatOnlySystems []GNSSSystem
	// ReportResiduals reports whether residuals are included in the report.
	ReportResiduals bool
	// HasARArmingSigmaM reports whether ARArmingSigmaM is supplied.
	HasARArmingSigmaM bool
	// ARArmingSigmaM contains metres.
	ARArmingSigmaM float64
	RatioThreshold float64
	// ReceiverAntenna refers to an optional value; nil means it is unavailable.
	ReceiverAntenna *RTKReceiverAntennaCorrections
}

// DefaultRTKMeasurementModel returns the native default RTK measurement model.
func DefaultRTKMeasurementModel() (RTKMeasurementModel, error) {
	v, err := native.RtkMeasurementModelInit()
	return RTKMeasurementModel{CodeSigmaM: v.CodeSigmaM, PhaseSigmaM: v.PhaseSigmaM, Sagnac: v.Sagnac, Stochastic: v.Stochastic, ElevationWeighting: v.ElevationWeighting}, publicError(err)
}

// DefaultRTKResidualValidationOptions returns native default residual-validation settings.
func DefaultRTKResidualValidationOptions() (RTKResidualValidationOptions, error) {
	v, err := native.RtkResidualValidationOptionsInit()
	return RTKResidualValidationOptions{ThresholdSigmaEnabled: v.ThresholdSigmaEnabled, ThresholdSigma: v.ThresholdSigma, MaxExclusions: v.MaxExclusions}, publicError(err)
}

// DefaultRTKFloatOptions returns native default RTK-float solver settings.
func DefaultRTKFloatOptions() (RTKFloatOptions, error) {
	v, err := native.RtkFloatOptionsInit()
	return RTKFloatOptions{PositionTolM: v.PositionTolM, AmbiguityTolM: v.AmbiguityTolM, MaxIterations: v.MaxIterations}, publicError(err)
}

// DefaultRTKFixedOptions returns native default integer-fixing settings.
func DefaultRTKFixedOptions() (RTKFixedOptions, error) {
	v, err := native.RtkFixedOptionsInit()
	return RTKFixedOptions{PositionTolM: v.PositionTolM, AmbiguityTolM: v.AmbiguityTolM, MaxIterations: v.MaxIterations, RatioThreshold: v.RatioThreshold, PartialAmbiguityResolution: v.PartialAmbiguityResolution, PartialMinAmbiguities: v.PartialMinAmbiguities}, publicError(err)
}

// DefaultRTKArcUpdateOptions returns native default RTK arc-update settings.
func DefaultRTKArcUpdateOptions() (RTKArcUpdateOptions, error) {
	v, err := native.RtkArcUpdateOptionsInit()
	systems := make([]GNSSSystem, len(v.FloatOnlySystems))
	for i, system := range v.FloatOnlySystems {
		systems[i] = GNSSSystem(system)
	}
	return RTKArcUpdateOptions{HoldSigmaM: v.HoldSigmaM, PositionTolM: v.PositionTolM, AmbiguityTolM: v.AmbiguityTolM, MaxIterations: v.MaxIterations, ProcessNoiseBaselineSigmaM: v.ProcessNoiseBaselineSigmaM, DynamicsVelocityPropagated: v.DynamicsVelocityPropagated, FloatOnlySystems: systems, ReportResiduals: v.ReportResiduals, HasARArmingSigmaM: v.HasARArmingSigmaM, ARArmingSigmaM: v.ARArmingSigmaM, RatioThreshold: v.RatioThreshold}, publicError(err)
}

// RTKSatMeasurement contains one satellite's code, phase, Doppler, and frequency measurements.
type RTKSatMeasurement struct {
	// SatelliteID identifies the observed satellite; SDAmbiguityID identifies its single-difference ambiguity.
	SatelliteID, SDAmbiguityID string
	// BaseCodeM and BasePhaseM are base code/phase values in metres; RoverCodeM and RoverPhaseM are rover values in metres.
	BaseCodeM, BasePhaseM, RoverCodeM, RoverPhaseM float64
	// BaseTXPos, RoverTXPos, and Pos are ECEF positions in metres at the transmit/receive epochs.
	BaseTXPos, RoverTXPos, Pos [3]float64
}

// RTKEpoch contains base/reference and rover/non-reference measurements for one RTK epoch.
type RTKEpoch struct {
	References, NonReference []RTKSatMeasurement
	// HasVelocityMPS reports whether VelocityMPS is supplied.
	HasVelocityMPS bool
	// VelocityMPS contains metres per second.
	VelocityMPS [3]float64
	// DTS is the transmit-to-receive time difference in seconds.
	DTS float64
}

// RTKAmbiguity contains one carrier-phase ambiguity and its satellite association.
type RTKAmbiguity struct {
	// ID identifies the associated record.
	ID string
	// ValueM contains metres.
	ValueM float64
}

// RTKFloatConfig configures an RTK float solve from epochs and ambiguity identifiers.
type RTKFloatConfig struct {
	Epochs []RTKEpoch
	// BaseECEFM contains metres.
	BaseECEFM    [3]float64
	AmbiguityIDs []string
	Model        RTKMeasurementModel
	// InitialBaselineM contains metres.
	InitialBaselineM [3]float64
	Options          RTKFloatOptions
}

// RTKFloatMetadata contains RTK float iteration, observation, ambiguity, residual, status, and geometry metadata.
type RTKFloatMetadata struct {
	Iterations, NObservations, AmbiguityCount, ResidualCount, UsedSatCount int
	// Converged reports whether the solve converged.
	Converged bool
	Status    uint32
	// CodeRMSM, PhaseRMSM, and WeightedRMSM are code, phase, and weighted residual RMS values in metres.
	CodeRMSM, PhaseRMSM, WeightedRMSM float64
	GeometryQuality                   SPPGeometryQuality
}

// RTKFloatSolution owns the native RTK float solution and its detached readers.
type RTKFloatSolution struct {
	_      noCopy
	handle *native.RtkFloatSolution
}

func nativeRTKFloatConfig(value RTKFloatConfig) native.RtkFloatConfig {
	result := native.RtkFloatConfig{BaseECEFM: value.BaseECEFM, AmbiguityIDs: append([]string(nil), value.AmbiguityIDs...), InitialBaselineM: value.InitialBaselineM, Model: native.RtkMeasurementModel{CodeSigmaM: value.Model.CodeSigmaM, PhaseSigmaM: value.Model.PhaseSigmaM, Sagnac: value.Model.Sagnac, Stochastic: value.Model.Stochastic, ElevationWeighting: value.Model.ElevationWeighting}, Options: native.RtkFloatOptions{PositionTolM: value.Options.PositionTolM, AmbiguityTolM: value.Options.AmbiguityTolM, MaxIterations: value.Options.MaxIterations}}
	result.Epochs = make([]native.RtkEpoch, len(value.Epochs))
	for i, epoch := range value.Epochs {
		result.Epochs[i] = native.RtkEpoch{HasVelocityMPS: epoch.HasVelocityMPS, VelocityMPS: epoch.VelocityMPS, DTS: epoch.DTS}
		result.Epochs[i].References = make([]native.RtkSatMeasurement, len(epoch.References))
		result.Epochs[i].NonReference = make([]native.RtkSatMeasurement, len(epoch.NonReference))
		for j, row := range epoch.References {
			result.Epochs[i].References[j] = native.RtkSatMeasurement{SatelliteID: row.SatelliteID, SDAmbiguityID: row.SDAmbiguityID, BaseCodeM: row.BaseCodeM, BasePhaseM: row.BasePhaseM, RoverCodeM: row.RoverCodeM, RoverPhaseM: row.RoverPhaseM, BaseTXPos: row.BaseTXPos, RoverTXPos: row.RoverTXPos, Pos: row.Pos}
		}
		for j, row := range epoch.NonReference {
			result.Epochs[i].NonReference[j] = native.RtkSatMeasurement{SatelliteID: row.SatelliteID, SDAmbiguityID: row.SDAmbiguityID, BaseCodeM: row.BaseCodeM, BasePhaseM: row.BasePhaseM, RoverCodeM: row.RoverCodeM, RoverPhaseM: row.RoverPhaseM, BaseTXPos: row.BaseTXPos, RoverTXPos: row.RoverTXPos, Pos: row.Pos}
		}
	}
	return result
}

// SolveRTKFloat solves the configured RTK float ambiguity problem.
func SolveRTKFloat(config RTKFloatConfig) (*RTKFloatSolution, error) {
	handle, err := native.SolveRtkFloat(nativeRTKFloatConfig(config))
	if err != nil {
		return nil, publicError(err)
	}
	return &RTKFloatSolution{handle: handle}, nil
}

// Close releases the native RTK float solution; repeated calls are safe.
func (s *RTKFloatSolution) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return publicError(s.handle.Close())
}

// BaselineECEF returns the RTK float baseline in ECEF metres.
func (s *RTKFloatSolution) BaselineECEF() ([3]float64, error) {
	if s == nil || s.handle == nil {
		return [3]float64{}, ErrClosed
	}
	value, err := s.handle.BaselineECEF()
	return value, publicError(err)
}

// BaselineENU returns the RTK float baseline in local ENU metres.
func (s *RTKFloatSolution) BaselineENU() ([3]float64, error) {
	if s == nil || s.handle == nil {
		return [3]float64{}, ErrClosed
	}
	value, err := s.handle.BaselineENU()
	return value, publicError(err)
}

// Metadata returns solver iteration, geometry, and status metadata.
func (s *RTKFloatSolution) Metadata() (RTKFloatMetadata, error) {
	if s == nil || s.handle == nil {
		return RTKFloatMetadata{}, ErrClosed
	}
	v, err := s.handle.Metadata()
	return RTKFloatMetadata{Iterations: v.Iterations, NObservations: v.NObservations, AmbiguityCount: v.AmbiguityCount, ResidualCount: v.ResidualCount, UsedSatCount: v.UsedSatCount, Converged: v.Converged, Status: v.Status, CodeRMSM: v.CodeRMSM, PhaseRMSM: v.PhaseRMSM, WeightedRMSM: v.WeightedRMSM, GeometryQuality: SPPGeometryQuality{Tier: v.GeometryQuality.Tier, Redundancy: v.GeometryQuality.Redundancy, Rank: v.GeometryQuality.Rank, ConditionNumber: v.GeometryQuality.ConditionNumber, GDOP: v.GeometryQuality.GDOP, RAIMCheckable: v.GeometryQuality.RAIMCheckable, CovarianceValidated: v.GeometryQuality.CovarianceValidated}}, publicError(err)
}

// Ambiguities returns detached float carrier-phase ambiguities.
func (s *RTKFloatSolution) Ambiguities() ([]RTKAmbiguity, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	v, err := s.handle.Ambiguities()
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]RTKAmbiguity, len(v))
	for i, row := range v {
		result[i] = RTKAmbiguity{ID: row.ID, ValueM: row.ValueM}
	}
	return result, nil
}

// UsedSatelliteIDs returns detached identifiers of satellites used by the solution.
func (s *RTKFloatSolution) UsedSatelliteIDs() ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	v, err := s.handle.UsedSatelliteIDs()
	return append([]string(nil), v...), publicError(err)
}
