package sidereon

import (
	"github.com/neilberkman/sidereon-go/internal/native"
)

type RTKMeasurementModel struct {
	CodeSigmaM, PhaseSigmaM float64
	Sagnac                  bool
	Stochastic              uint32
	ElevationWeighting      bool
}

type RTKResidualValidationOptions struct {
	ThresholdSigmaEnabled bool
	ThresholdSigma        float64
	MaxExclusions         int
}

type RTKFloatOptions struct {
	PositionTolM, AmbiguityTolM float64
	MaxIterations               int
}

type RTKFixedOptions struct {
	PositionTolM, AmbiguityTolM float64
	MaxIterations               int
	RatioThreshold              float64
	PartialAmbiguityResolution  bool
	PartialMinAmbiguities       int
}

type RTKArcUpdateOptions struct {
	HoldSigmaM, PositionTolM, AmbiguityTolM float64
	MaxIterations                           int
	ProcessNoiseBaselineSigmaM              float64
	DynamicsVelocityPropagated              bool
	FloatOnlySystems                        []GNSSSystem
	ReportResiduals                         bool
	HasARArmingSigmaM                       bool
	ARArmingSigmaM                          float64
	RatioThreshold                          float64
	ReceiverAntenna                         *RTKReceiverAntennaCorrections
}

func DefaultRTKMeasurementModel() (RTKMeasurementModel, error) {
	v, err := native.RtkMeasurementModelInit()
	return RTKMeasurementModel{CodeSigmaM: v.CodeSigmaM, PhaseSigmaM: v.PhaseSigmaM, Sagnac: v.Sagnac, Stochastic: v.Stochastic, ElevationWeighting: v.ElevationWeighting}, publicError(err)
}

func DefaultRTKResidualValidationOptions() (RTKResidualValidationOptions, error) {
	v, err := native.RtkResidualValidationOptionsInit()
	return RTKResidualValidationOptions{ThresholdSigmaEnabled: v.ThresholdSigmaEnabled, ThresholdSigma: v.ThresholdSigma, MaxExclusions: v.MaxExclusions}, publicError(err)
}

func DefaultRTKFloatOptions() (RTKFloatOptions, error) {
	v, err := native.RtkFloatOptionsInit()
	return RTKFloatOptions{PositionTolM: v.PositionTolM, AmbiguityTolM: v.AmbiguityTolM, MaxIterations: v.MaxIterations}, publicError(err)
}

func DefaultRTKFixedOptions() (RTKFixedOptions, error) {
	v, err := native.RtkFixedOptionsInit()
	return RTKFixedOptions{PositionTolM: v.PositionTolM, AmbiguityTolM: v.AmbiguityTolM, MaxIterations: v.MaxIterations, RatioThreshold: v.RatioThreshold, PartialAmbiguityResolution: v.PartialAmbiguityResolution, PartialMinAmbiguities: v.PartialMinAmbiguities}, publicError(err)
}

func DefaultRTKArcUpdateOptions() (RTKArcUpdateOptions, error) {
	v, err := native.RtkArcUpdateOptionsInit()
	systems := make([]GNSSSystem, len(v.FloatOnlySystems))
	for i, system := range v.FloatOnlySystems {
		systems[i] = GNSSSystem(system)
	}
	return RTKArcUpdateOptions{HoldSigmaM: v.HoldSigmaM, PositionTolM: v.PositionTolM, AmbiguityTolM: v.AmbiguityTolM, MaxIterations: v.MaxIterations, ProcessNoiseBaselineSigmaM: v.ProcessNoiseBaselineSigmaM, DynamicsVelocityPropagated: v.DynamicsVelocityPropagated, FloatOnlySystems: systems, ReportResiduals: v.ReportResiduals, HasARArmingSigmaM: v.HasARArmingSigmaM, ARArmingSigmaM: v.ARArmingSigmaM, RatioThreshold: v.RatioThreshold}, publicError(err)
}

type RTKSatMeasurement struct {
	SatelliteID, SDAmbiguityID                     string
	BaseCodeM, BasePhaseM, RoverCodeM, RoverPhaseM float64
	BaseTXPos, RoverTXPos, Pos                     [3]float64
}

type RTKEpoch struct {
	References, NonReference []RTKSatMeasurement
	HasVelocityMPS           bool
	VelocityMPS              [3]float64
	DTS                      float64
}

type RTKAmbiguity struct {
	ID     string
	ValueM float64
}

type RTKFloatConfig struct {
	Epochs           []RTKEpoch
	BaseECEFM        [3]float64
	AmbiguityIDs     []string
	Model            RTKMeasurementModel
	InitialBaselineM [3]float64
	Options          RTKFloatOptions
}

type RTKFloatMetadata struct {
	Iterations, NObservations, AmbiguityCount, ResidualCount, UsedSatCount int
	Converged                                                              bool
	Status                                                                 uint32
	CodeRMSM, PhaseRMSM, WeightedRMSM                                      float64
	GeometryQuality                                                        SPPGeometryQuality
}

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

func SolveRTKFloat(config RTKFloatConfig) (*RTKFloatSolution, error) {
	handle, err := native.SolveRtkFloat(nativeRTKFloatConfig(config))
	if err != nil {
		return nil, publicError(err)
	}
	return &RTKFloatSolution{handle: handle}, nil
}

func (s *RTKFloatSolution) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return publicError(s.handle.Close())
}
func (s *RTKFloatSolution) BaselineECEF() ([3]float64, error) {
	if s == nil || s.handle == nil {
		return [3]float64{}, ErrClosed
	}
	value, err := s.handle.BaselineECEF()
	return value, publicError(err)
}
func (s *RTKFloatSolution) BaselineENU() ([3]float64, error) {
	if s == nil || s.handle == nil {
		return [3]float64{}, ErrClosed
	}
	value, err := s.handle.BaselineENU()
	return value, publicError(err)
}
func (s *RTKFloatSolution) Metadata() (RTKFloatMetadata, error) {
	if s == nil || s.handle == nil {
		return RTKFloatMetadata{}, ErrClosed
	}
	v, err := s.handle.Metadata()
	return RTKFloatMetadata{Iterations: v.Iterations, NObservations: v.NObservations, AmbiguityCount: v.AmbiguityCount, ResidualCount: v.ResidualCount, UsedSatCount: v.UsedSatCount, Converged: v.Converged, Status: v.Status, CodeRMSM: v.CodeRMSM, PhaseRMSM: v.PhaseRMSM, WeightedRMSM: v.WeightedRMSM, GeometryQuality: SPPGeometryQuality{Tier: v.GeometryQuality.Tier, Redundancy: v.GeometryQuality.Redundancy, Rank: v.GeometryQuality.Rank, ConditionNumber: v.GeometryQuality.ConditionNumber, GDOP: v.GeometryQuality.GDOP, RAIMCheckable: v.GeometryQuality.RAIMCheckable, CovarianceValidated: v.GeometryQuality.CovarianceValidated}}, publicError(err)
}
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
func (s *RTKFloatSolution) UsedSatelliteIDs() ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	v, err := s.handle.UsedSatelliteIDs()
	return append([]string(nil), v...), publicError(err)
}
