package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// RTKRINEXStaticBaselineConfig configures a static RTK solve from paired
// RINEX observations and an SP3 product.
// RTKRINEXStaticBaselineConfig configures static-baseline RTK processing from RINEX observations.
type RTKRINEXStaticBaselineConfig struct {
	// BaseM contains metres.
	BaseM         [3]float64
	ArcOptions    RTKRINEXArcOptions
	ReferenceMode RTKArcReferenceMode
	// ReferenceSatellite is the selected reference-satellite identifier.
	ReferenceSatellite string
	ReferencePerSystem []RTKArcReferenceEntry
	Model              RTKMeasurementModel
	// BaselinePriorSigmaM contains metres.
	BaselinePriorSigmaM float64
	// AmbiguityPriorSigmaM contains metres.
	AmbiguityPriorSigmaM float64
	// InitialBaselineM contains metres.
	InitialBaselineM [3]float64
	UpdateOptions    RTKArcUpdateOptions
	Preprocessing    RTKArcPreprocessing
	FloatOptions     RTKFloatOptions
	FixedOptions     RTKFixedOptions
	ResidualOptions  RTKResidualValidationOptions
}

// RTKRINEXWideLaneFixedConfig configures a static dual-frequency wide-lane
// fixed solve from paired RINEX observations and an SP3 product.
// RTKRINEXWideLaneFixedConfig configures wide-lane-fixed RTK processing from RINEX observations.
type RTKRINEXWideLaneFixedConfig struct {
	// BaseM contains metres.
	BaseM         [3]float64
	ArcOptions    RTKRINEXDualArcOptions
	ReferenceMode RTKArcReferenceMode
	// ReferenceSatellite is the selected reference-satellite identifier.
	ReferenceSatellite string
	ReferencePerSystem []RTKArcReferenceEntry
	Model              RTKMeasurementModel
	// BaselinePriorSigmaM contains metres.
	BaselinePriorSigmaM float64
	// AmbiguityPriorSigmaM contains metres.
	AmbiguityPriorSigmaM float64
	// InitialBaselineM contains metres.
	InitialBaselineM [3]float64
	UpdateOptions    RTKArcUpdateOptions
	FloatOptions     RTKFloatOptions
	FixedOptions     RTKFixedOptions
	ResidualOptions  RTKResidualValidationOptions
	// ApplyTroposphere reports whether the troposphere correction is enabled.
	ApplyTroposphere bool
}

// RTKStaticArcConfig configures the static raw-arc float and fixed solve.
type RTKStaticArcConfig struct {
	Arc             RTKArcConfig
	FloatOptions    RTKFloatOptions
	FixedOptions    RTKFixedOptions
	ResidualOptions RTKResidualValidationOptions
}

// RTKWideLaneFixedRINEXMetadata summarizes a combined wide-lane fixed solve.
type RTKWideLaneFixedRINEXMetadata struct {
	// WideLaneFixed reports whether a wide-lane fixed solution was obtained.
	WideLaneFixed          bool
	WideLaneAmbiguityCount int
	DroppedCycleSlipCount  int
	SplitCycleSlipCount    int
}

// RTKStaticArcSolution owns a static RTK result and must not be copied after
// first use.
// RTKStaticArcSolution owns a native static RTK arc solution.
type RTKStaticArcSolution struct {
	_      noCopy
	handle *native.RtkStaticArcSolution
}

// RTKWideLaneFixedRINEXSolution owns a combined wide-lane fixed RINEX result
// and must not be copied after first use.
// RTKWideLaneFixedRINEXSolution owns a native wide-lane-fixed RINEX RTK solution.
type RTKWideLaneFixedRINEXSolution struct {
	_      noCopy
	handle *native.RtkWideLaneFixedRinexSolution
}

func nativeRTKRINEXStaticConfig(value RTKRINEXStaticBaselineConfig) native.RtkRinexStaticBaselineConfig {
	result := native.RtkRinexStaticBaselineConfig{BaseM: value.BaseM, ArcOptions: nativeRTKRINEXArcOptions(value.ArcOptions), ReferenceMode: uint32(value.ReferenceMode), ReferenceSatellite: value.ReferenceSatellite, ReferencePerSystem: nativeRTKArcReferences(value.ReferencePerSystem), Model: native.RtkMeasurementModel{CodeSigmaM: value.Model.CodeSigmaM, PhaseSigmaM: value.Model.PhaseSigmaM, Sagnac: value.Model.Sagnac, Stochastic: value.Model.Stochastic, ElevationWeighting: value.Model.ElevationWeighting}, BaselinePriorSigmaM: value.BaselinePriorSigmaM, AmbiguityPriorSigmaM: value.AmbiguityPriorSigmaM, InitialBaselineM: value.InitialBaselineM, UpdateOptions: nativeRTKUpdateOptions(value.UpdateOptions), Preprocessing: native.RtkArcPreprocessing{HasCycleSlip: value.Preprocessing.HasCycleSlip, CycleSlip: uint32(value.Preprocessing.CycleSlip), HasHatchWindowCap: value.Preprocessing.HasHatchWindowCap, HatchWindowCap: value.Preprocessing.HatchWindowCap, HasElevationMaskDeg: value.Preprocessing.HasElevationMaskDeg, ElevationMaskDeg: value.Preprocessing.ElevationMaskDeg}, FloatOptions: native.RtkFloatOptions{PositionTolM: value.FloatOptions.PositionTolM, AmbiguityTolM: value.FloatOptions.AmbiguityTolM, MaxIterations: value.FloatOptions.MaxIterations}, FixedOptions: native.RtkFixedOptions{PositionTolM: value.FixedOptions.PositionTolM, AmbiguityTolM: value.FixedOptions.AmbiguityTolM, MaxIterations: value.FixedOptions.MaxIterations, RatioThreshold: value.FixedOptions.RatioThreshold, PartialAmbiguityResolution: value.FixedOptions.PartialAmbiguityResolution, PartialMinAmbiguities: value.FixedOptions.PartialMinAmbiguities}, ResidualOptions: native.RtkResidualValidationOptions{ThresholdSigmaEnabled: value.ResidualOptions.ThresholdSigmaEnabled, ThresholdSigma: value.ResidualOptions.ThresholdSigma, MaxExclusions: value.ResidualOptions.MaxExclusions}}
	return result
}

func nativeRTKRINEXWideLaneFixedConfig(value RTKRINEXWideLaneFixedConfig) native.RtkRinexWideLaneFixedConfig {
	return native.RtkRinexWideLaneFixedConfig{BaseM: value.BaseM, ArcOptions: nativeRTKRINEXDualArcOptions(value.ArcOptions), ReferenceMode: uint32(value.ReferenceMode), ReferenceSatellite: value.ReferenceSatellite, ReferencePerSystem: nativeRTKArcReferences(value.ReferencePerSystem), Model: native.RtkMeasurementModel{CodeSigmaM: value.Model.CodeSigmaM, PhaseSigmaM: value.Model.PhaseSigmaM, Sagnac: value.Model.Sagnac, Stochastic: value.Model.Stochastic, ElevationWeighting: value.Model.ElevationWeighting}, BaselinePriorSigmaM: value.BaselinePriorSigmaM, AmbiguityPriorSigmaM: value.AmbiguityPriorSigmaM, InitialBaselineM: value.InitialBaselineM, UpdateOptions: nativeRTKUpdateOptions(value.UpdateOptions), FloatOptions: native.RtkFloatOptions{PositionTolM: value.FloatOptions.PositionTolM, AmbiguityTolM: value.FloatOptions.AmbiguityTolM, MaxIterations: value.FloatOptions.MaxIterations}, FixedOptions: native.RtkFixedOptions{PositionTolM: value.FixedOptions.PositionTolM, AmbiguityTolM: value.FixedOptions.AmbiguityTolM, MaxIterations: value.FixedOptions.MaxIterations, RatioThreshold: value.FixedOptions.RatioThreshold, PartialAmbiguityResolution: value.FixedOptions.PartialAmbiguityResolution, PartialMinAmbiguities: value.FixedOptions.PartialMinAmbiguities}, ResidualOptions: native.RtkResidualValidationOptions{ThresholdSigmaEnabled: value.ResidualOptions.ThresholdSigmaEnabled, ThresholdSigma: value.ResidualOptions.ThresholdSigma, MaxExclusions: value.ResidualOptions.MaxExclusions}, ApplyTroposphere: value.ApplyTroposphere}
}

func nativeRTKUpdateOptions(value RTKArcUpdateOptions) native.RtkArcUpdateOptions {
	return native.RtkArcUpdateOptions{HoldSigmaM: value.HoldSigmaM, PositionTolM: value.PositionTolM, AmbiguityTolM: value.AmbiguityTolM, MaxIterations: value.MaxIterations, ProcessNoiseBaselineSigmaM: value.ProcessNoiseBaselineSigmaM, DynamicsVelocityPropagated: value.DynamicsVelocityPropagated, FloatOnlySystems: nativeRTKSystems(value.FloatOnlySystems), ReportResiduals: value.ReportResiduals, HasARArmingSigmaM: value.HasARArmingSigmaM, ARArmingSigmaM: value.ARArmingSigmaM, RatioThreshold: value.RatioThreshold, ReceiverAntenna: nativeRTKReceiverAntenna(value.ReceiverAntenna)}
}

func nativeRTKStaticArcConfig(value RTKStaticArcConfig) native.RtkStaticArcConfigInput {
	return native.RtkStaticArcConfigInput{Arc: nativeRTKArcConfig(value.Arc), FloatOptions: native.RtkFloatOptions{PositionTolM: value.FloatOptions.PositionTolM, AmbiguityTolM: value.FloatOptions.AmbiguityTolM, MaxIterations: value.FloatOptions.MaxIterations}, FixedOptions: native.RtkFixedOptions{PositionTolM: value.FixedOptions.PositionTolM, AmbiguityTolM: value.FixedOptions.AmbiguityTolM, MaxIterations: value.FixedOptions.MaxIterations, RatioThreshold: value.FixedOptions.RatioThreshold, PartialAmbiguityResolution: value.FixedOptions.PartialAmbiguityResolution, PartialMinAmbiguities: value.FixedOptions.PartialMinAmbiguities}, ResidualOptions: native.RtkResidualValidationOptions{ThresholdSigmaEnabled: value.ResidualOptions.ThresholdSigmaEnabled, ThresholdSigma: value.ResidualOptions.ThresholdSigma, MaxExclusions: value.ResidualOptions.MaxExclusions}}
}

// DefaultRTKRINEXStaticBaselineConfig returns C's static RINEX RTK defaults.
func DefaultRTKRINEXStaticBaselineConfig() (RTKRINEXStaticBaselineConfig, error) {
	value, err := native.RtkRinexStaticBaselineConfigInit()
	return publicRTKRINEXStaticConfig(value), publicError(err)
}

func publicRTKRINEXStaticConfig(value native.RtkRinexStaticBaselineConfig) RTKRINEXStaticBaselineConfig {
	result := RTKRINEXStaticBaselineConfig{BaseM: value.BaseM, ArcOptions: publicRTKRINEXArcOptions(value.ArcOptions), ReferenceMode: RTKArcReferenceMode(value.ReferenceMode), ReferenceSatellite: value.ReferenceSatellite, Model: RTKMeasurementModel{CodeSigmaM: value.Model.CodeSigmaM, PhaseSigmaM: value.Model.PhaseSigmaM, Sagnac: value.Model.Sagnac, Stochastic: value.Model.Stochastic, ElevationWeighting: value.Model.ElevationWeighting}, BaselinePriorSigmaM: value.BaselinePriorSigmaM, AmbiguityPriorSigmaM: value.AmbiguityPriorSigmaM, InitialBaselineM: value.InitialBaselineM, UpdateOptions: publicRTKUpdateOptions(value.UpdateOptions), Preprocessing: RTKArcPreprocessing{HasCycleSlip: value.Preprocessing.HasCycleSlip, CycleSlip: RTKCycleSlipPolicy(value.Preprocessing.CycleSlip), HasHatchWindowCap: value.Preprocessing.HasHatchWindowCap, HatchWindowCap: value.Preprocessing.HatchWindowCap, HasElevationMaskDeg: value.Preprocessing.HasElevationMaskDeg, ElevationMaskDeg: value.Preprocessing.ElevationMaskDeg}, FloatOptions: RTKFloatOptions{PositionTolM: value.FloatOptions.PositionTolM, AmbiguityTolM: value.FloatOptions.AmbiguityTolM, MaxIterations: value.FloatOptions.MaxIterations}, FixedOptions: RTKFixedOptions{PositionTolM: value.FixedOptions.PositionTolM, AmbiguityTolM: value.FixedOptions.AmbiguityTolM, MaxIterations: value.FixedOptions.MaxIterations, RatioThreshold: value.FixedOptions.RatioThreshold, PartialAmbiguityResolution: value.FixedOptions.PartialAmbiguityResolution, PartialMinAmbiguities: value.FixedOptions.PartialMinAmbiguities}, ResidualOptions: RTKResidualValidationOptions{ThresholdSigmaEnabled: value.ResidualOptions.ThresholdSigmaEnabled, ThresholdSigma: value.ResidualOptions.ThresholdSigma, MaxExclusions: value.ResidualOptions.MaxExclusions}}
	for _, reference := range value.ReferencePerSystem {
		result.ReferencePerSystem = append(result.ReferencePerSystem, RTKArcReferenceEntry{System: GNSSSystem(reference.System), SatelliteID: reference.SatelliteID})
	}
	return result
}

func publicRTKRINEXArcOptions(value native.RtkRinexArcOptions) RTKRINEXArcOptions {
	result := RTKRINEXArcOptions{HasMaxEpochs: value.HasMaxEpochs, MaxEpochs: value.MaxEpochs, MinCommonSatellites: value.MinCommonSatellites, IncludePredictionTime: value.IncludePredictionTime}
	for _, pair := range value.SignalPairs {
		result.SignalPairs = append(result.SignalPairs, RTKRINEXSignalPair{System: GNSSSystem(pair.System), CodeObservable: pair.CodeObservable, PhaseObservable: pair.PhaseObservable})
	}
	return result
}

func publicRTKUpdateOptions(value native.RtkArcUpdateOptions) RTKArcUpdateOptions {
	result := RTKArcUpdateOptions{HoldSigmaM: value.HoldSigmaM, PositionTolM: value.PositionTolM, AmbiguityTolM: value.AmbiguityTolM, MaxIterations: value.MaxIterations, ProcessNoiseBaselineSigmaM: value.ProcessNoiseBaselineSigmaM, DynamicsVelocityPropagated: value.DynamicsVelocityPropagated, ReportResiduals: value.ReportResiduals, HasARArmingSigmaM: value.HasARArmingSigmaM, ARArmingSigmaM: value.ARArmingSigmaM, RatioThreshold: value.RatioThreshold}
	if value.ReceiverAntenna != nil {
		result.ReceiverAntenna = publicRTKReceiverAntenna(value.ReceiverAntenna)
	}
	for _, system := range value.FloatOnlySystems {
		result.FloatOnlySystems = append(result.FloatOnlySystems, GNSSSystem(system))
	}
	return result
}

func publicRTKReceiverAntenna(value *native.RtkReceiverAntennaCorrections) *RTKReceiverAntennaCorrections {
	if value == nil {
		return nil
	}
	result := &RTKReceiverAntennaCorrections{Base: RTKReceiverAntennaCalibration{PCONEUM: value.Base.PCONEUM}, Rover: RTKReceiverAntennaCalibration{PCONEUM: value.Rover.PCONEUM}}
	for _, sample := range value.Base.NoAzimuthPCV {
		result.Base.NoAzimuthPCV = append(result.Base.NoAzimuthPCV, RTKReceiverAntennaNoAzimuthPCV{ZenithDeg: sample.ZenithDeg, ValueM: sample.ValueM})
	}
	for _, sample := range value.Base.AzimuthPCV {
		result.Base.AzimuthPCV = append(result.Base.AzimuthPCV, RTKReceiverAntennaAzimuthPCV{AzimuthDeg: sample.AzimuthDeg, ZenithDeg: sample.ZenithDeg, ValueM: sample.ValueM})
	}
	for _, sample := range value.Rover.NoAzimuthPCV {
		result.Rover.NoAzimuthPCV = append(result.Rover.NoAzimuthPCV, RTKReceiverAntennaNoAzimuthPCV{ZenithDeg: sample.ZenithDeg, ValueM: sample.ValueM})
	}
	for _, sample := range value.Rover.AzimuthPCV {
		result.Rover.AzimuthPCV = append(result.Rover.AzimuthPCV, RTKReceiverAntennaAzimuthPCV{AzimuthDeg: sample.AzimuthDeg, ZenithDeg: sample.ZenithDeg, ValueM: sample.ValueM})
	}
	return result
}

// DefaultRTKRINEXWideLaneFixedConfig returns C's dual-frequency static
// wide-lane fixed defaults.
// DefaultRTKRINEXWideLaneFixedConfig returns native defaults for wide-lane-fixed RINEX processing.
func DefaultRTKRINEXWideLaneFixedConfig() (RTKRINEXWideLaneFixedConfig, error) {
	value, err := native.RtkRinexWideLaneFixedConfigInit()
	return publicRTKRINEXWideLaneFixedConfig(value), publicError(err)
}

func publicRTKRINEXWideLaneFixedConfig(value native.RtkRinexWideLaneFixedConfig) RTKRINEXWideLaneFixedConfig {
	result := RTKRINEXWideLaneFixedConfig{BaseM: value.BaseM, ReferenceMode: RTKArcReferenceMode(value.ReferenceMode), ReferenceSatellite: value.ReferenceSatellite, Model: RTKMeasurementModel{CodeSigmaM: value.Model.CodeSigmaM, PhaseSigmaM: value.Model.PhaseSigmaM, Sagnac: value.Model.Sagnac, Stochastic: value.Model.Stochastic, ElevationWeighting: value.Model.ElevationWeighting}, BaselinePriorSigmaM: value.BaselinePriorSigmaM, AmbiguityPriorSigmaM: value.AmbiguityPriorSigmaM, InitialBaselineM: value.InitialBaselineM, UpdateOptions: publicRTKUpdateOptions(value.UpdateOptions), FloatOptions: RTKFloatOptions{PositionTolM: value.FloatOptions.PositionTolM, AmbiguityTolM: value.FloatOptions.AmbiguityTolM, MaxIterations: value.FloatOptions.MaxIterations}, FixedOptions: RTKFixedOptions{PositionTolM: value.FixedOptions.PositionTolM, AmbiguityTolM: value.FixedOptions.AmbiguityTolM, MaxIterations: value.FixedOptions.MaxIterations, RatioThreshold: value.FixedOptions.RatioThreshold, PartialAmbiguityResolution: value.FixedOptions.PartialAmbiguityResolution, PartialMinAmbiguities: value.FixedOptions.PartialMinAmbiguities}, ResidualOptions: RTKResidualValidationOptions{ThresholdSigmaEnabled: value.ResidualOptions.ThresholdSigmaEnabled, ThresholdSigma: value.ResidualOptions.ThresholdSigma, MaxExclusions: value.ResidualOptions.MaxExclusions}, ApplyTroposphere: value.ApplyTroposphere}
	for _, reference := range value.ReferencePerSystem {
		result.ReferencePerSystem = append(result.ReferencePerSystem, RTKArcReferenceEntry{System: GNSSSystem(reference.System), SatelliteID: reference.SatelliteID})
	}
	for _, pair := range value.ArcOptions.SignalPairs {
		result.ArcOptions.SignalPairs = append(result.ArcOptions.SignalPairs, RTKRINEXDualSignalPair{System: GNSSSystem(pair.System), Code1Observable: pair.Code1Observable, Phase1Observable: pair.Phase1Observable, Code2Observable: pair.Code2Observable, Phase2Observable: pair.Phase2Observable})
	}
	result.ArcOptions.HasMaxEpochs, result.ArcOptions.MaxEpochs, result.ArcOptions.MinCommonSatellites, result.ArcOptions.IncludePredictionTime = value.ArcOptions.HasMaxEpochs, value.ArcOptions.MaxEpochs, value.ArcOptions.MinCommonSatellites, value.ArcOptions.IncludePredictionTime
	return result
}

// SolveStaticRTKArc solves a static raw RTK arc through C.
func SolveStaticRTKArc(epochs []RTKArcEpoch, config RTKStaticArcConfig) (*RTKStaticArcSolution, error) {
	values := make([]native.RtkArcEpochInput, len(epochs))
	for i, epoch := range epochs {
		values[i] = nativeRTKArcEpoch(epoch)
	}
	result, err := native.SolveStaticRtkArc(values, nativeRTKStaticArcConfig(config))
	if err != nil {
		return nil, publicError(err)
	}
	return &RTKStaticArcSolution{handle: result}, nil
}

// SolveStaticRINEXRTKBaseline solves a static RTK baseline from RINEX OBS and
// SP3 handles.
// SolveStaticRINEXRTKBaseline solves a static RTK baseline from RINEX observations and SP3 positions.
func SolveStaticRINEXRTKBaseline(sp3 *SP3, base, rover *RINEXObservation, config RTKRINEXStaticBaselineConfig) (*RTKStaticArcSolution, error) {
	result, err := native.SolveStaticRinexRtkBaseline(nativeSP3(sp3), nativeRinexObs(base), nativeRinexObs(rover), nativeRTKRINEXStaticConfig(config))
	if err != nil {
		return nil, publicError(err)
	}
	return &RTKStaticArcSolution{handle: result}, nil
}

// SolveWideLaneFixedRINEXRTKBaseline solves a static dual-frequency wide-lane
// fixed baseline from RINEX OBS and SP3 handles.
// SolveWideLaneFixedRINEXRTKBaseline solves a dual-frequency wide-lane-fixed RINEX baseline.
func SolveWideLaneFixedRINEXRTKBaseline(sp3 *SP3, base, rover *RINEXObservation, config RTKRINEXWideLaneFixedConfig) (*RTKWideLaneFixedRINEXSolution, error) {
	result, err := native.SolveWideLaneFixedRinexRtkBaseline(nativeSP3(sp3), nativeRinexObs(base), nativeRinexObs(rover), nativeRTKRINEXWideLaneFixedConfig(config))
	if err != nil {
		return nil, publicError(err)
	}
	return &RTKWideLaneFixedRINEXSolution{handle: result}, nil
}

func publicRTKGeometry(value native.GeometryQuality) SPPGeometryQuality {
	return SPPGeometryQuality{Tier: value.Tier, Redundancy: value.Redundancy, Rank: value.Rank, ConditionNumber: value.ConditionNumber, GDOP: value.GDOP, RAIMCheckable: value.RAIMCheckable, CovarianceValidated: value.CovarianceValidated}
}

func publicRTKFixedMetadata(value native.RtkFixedMetadata) RTKFixedMetadata {
	return RTKFixedMetadata{Iterations: value.Iterations, NObservations: value.NObservations, FreeAmbiguityCount: value.FreeAmbiguityCount, FixedAmbiguityCount: value.FixedAmbiguityCount, ResidualCount: value.ResidualCount, UsedSatCount: value.UsedSatCount, Converged: value.Converged, Status: value.Status, CodeRMSM: value.CodeRMSM, PhaseRMSM: value.PhaseRMSM, WeightedRMSM: value.WeightedRMSM, IntegerStatus: RTKIntegerStatus(value.IntegerStatus), HasIntegerRatio: value.HasIntegerRatio, IntegerRatio: value.IntegerRatio, HasIntegerBestScore: value.HasIntegerBestScore, IntegerBestScore: value.IntegerBestScore, HasIntegerSecondBestScore: value.HasIntegerSecondBestScore, IntegerSecondBestScore: value.IntegerSecondBestScore, IntegerCandidates: value.IntegerCandidates, GeometryQuality: publicRTKGeometry(value.GeometryQuality)}
}

func publicRTKFloatMetadata(value native.RtkFloatMetadata) RTKFloatMetadata {
	return RTKFloatMetadata{Iterations: value.Iterations, NObservations: value.NObservations, AmbiguityCount: value.AmbiguityCount, ResidualCount: value.ResidualCount, UsedSatCount: value.UsedSatCount, Converged: value.Converged, Status: value.Status, CodeRMSM: value.CodeRMSM, PhaseRMSM: value.PhaseRMSM, WeightedRMSM: value.WeightedRMSM, GeometryQuality: publicRTKGeometry(value.GeometryQuality)}
}

// Close releases the native static RTK arc solution; repeated calls are safe.
func (s *RTKStaticArcSolution) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return publicError(s.handle.Close())
}

// AmbiguityIDs returns detached satellite IDs associated with RTK ambiguities.
func (s *RTKStaticArcSolution) AmbiguityIDs() ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.AmbiguityIDs()
	return append([]string(nil), value...), publicError(err)
}

// AmbiguitySatellites returns detached satellite metadata for RTK ambiguity entries.
func (s *RTKStaticArcSolution) AmbiguitySatellites() ([]RTKAmbiguitySatellite, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.AmbiguitySatellites()
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]RTKAmbiguitySatellite, len(value))
	for i, row := range value {
		result[i] = RTKAmbiguitySatellite{ID: row.ID, SatelliteID: row.SatelliteID}
	}
	return result, nil
}

// DroppedSatellites returns detached IDs of satellites dropped from the solution.
func (s *RTKStaticArcSolution) DroppedSatellites() ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.DroppedSatellites()
	return append([]string(nil), value...), publicError(err)
}

// ElevationMaskedSatellites returns detached IDs rejected by the elevation mask.
func (s *RTKStaticArcSolution) ElevationMaskedSatellites() ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.ElevationMaskedSatellites()
	return append([]string(nil), value...), publicError(err)
}

// FixedAmbiguities returns detached ambiguities accepted by integer fixing.
func (s *RTKStaticArcSolution) FixedAmbiguities() ([]RTKFixedAmbiguity, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.FixedAmbiguities()
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]RTKFixedAmbiguity, len(value))
	for i, row := range value {
		result[i] = RTKFixedAmbiguity{ID: row.ID, Cycles: row.Cycles, ValueM: row.ValueM}
	}
	return result, nil
}

// FixedBaselineECEF returns the fixed rover-minus-base baseline in ECEF metres.
func (s *RTKStaticArcSolution) FixedBaselineECEF() ([3]float64, error) {
	if s == nil || s.handle == nil {
		return [3]float64{}, ErrClosed
	}
	value, err := s.handle.FixedBaselineECEF()
	return value, publicError(err)
}

// FixedFreeAmbiguities returns detached ambiguities left free after integer fixing.
func (s *RTKStaticArcSolution) FixedFreeAmbiguities() ([]RTKAmbiguity, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.FixedFreeAmbiguities()
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]RTKAmbiguity, len(value))
	for i, row := range value {
		result[i] = RTKAmbiguity{ID: row.ID, ValueM: row.ValueM}
	}
	return result, nil
}

// FixedMetadata returns metadata from the fixed-ambiguity RTK solution.
func (s *RTKStaticArcSolution) FixedMetadata() (RTKFixedMetadata, error) {
	if s == nil || s.handle == nil {
		return RTKFixedMetadata{}, ErrClosed
	}
	value, err := s.handle.FixedMetadata()
	return publicRTKFixedMetadata(value), publicError(err)
}

// FloatAmbiguities returns detached ambiguities from the float solution.
func (s *RTKStaticArcSolution) FloatAmbiguities() ([]RTKAmbiguity, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.FloatAmbiguities()
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]RTKAmbiguity, len(value))
	for i, row := range value {
		result[i] = RTKAmbiguity{ID: row.ID, ValueM: row.ValueM}
	}
	return result, nil
}

// FloatBaselineECEF returns the float rover-minus-base baseline in ECEF metres.
func (s *RTKStaticArcSolution) FloatBaselineECEF() ([3]float64, error) {
	if s == nil || s.handle == nil {
		return [3]float64{}, ErrClosed
	}
	value, err := s.handle.FloatBaselineECEF()
	return value, publicError(err)
}

// FloatMetadata returns metadata from the RTK float solution.
func (s *RTKStaticArcSolution) FloatMetadata() (RTKFloatMetadata, error) {
	if s == nil || s.handle == nil {
		return RTKFloatMetadata{}, ErrClosed
	}
	value, err := s.handle.FloatMetadata()
	return publicRTKFloatMetadata(value), publicError(err)
}

// GeometryQuality returns satellite-geometry quality metrics for the RTK solution.
func (s *RTKStaticArcSolution) GeometryQuality() (SPPGeometryQuality, error) {
	if s == nil || s.handle == nil {
		return SPPGeometryQuality{}, ErrClosed
	}
	value, err := s.handle.GeometryQuality()
	return publicRTKGeometry(value), publicError(err)
}

// References returns detached ambiguity-reference selections used by the solution.
func (s *RTKStaticArcSolution) References() ([]RTKArcReference, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.References()
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]RTKArcReference, len(value))
	for i, row := range value {
		result[i] = RTKArcReference{System: row.System, ReferenceID: row.ReferenceID}
	}
	return result, nil
}

// SplitCycleSlipArcs returns detached arc ranges created by cycle-slip splitting.
func (s *RTKStaticArcSolution) SplitCycleSlipArcs() ([]RTKArcSplitArc, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.SplitCycleSlipArcs()
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]RTKArcSplitArc, len(value))
	for i, row := range value {
		result[i] = RTKArcSplitArc{Receiver: RTKCycleSlipReceiver(row.Receiver), SatelliteID: row.SatelliteID, AmbiguityID: row.AmbiguityID, StartEpochIndex: row.StartEpochIndex, EndEpochIndex: row.EndEpochIndex, EpochCount: row.EpochCount}
	}
	return result, nil
}

// DroppedSatellites returns detached IDs of satellites dropped from the solution.
func (s *RTKWideLaneArcSolution) DroppedSatellites() ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.DroppedSatellites()
	return append([]string(nil), value...), publicError(err)
}

// EpochCount returns the number of recorded epochs.
func (s *RTKWideLaneArcSolution) EpochCount() (int, error) {
	if s == nil || s.handle == nil {
		return 0, ErrClosed
	}
	value, err := s.handle.EpochCount()
	return value, publicError(err)
}

// GeometryQuality returns satellite-geometry quality metrics for the RTK solution.
func (s *RTKWideLaneArcSolution) GeometryQuality() (SPPGeometryQuality, error) {
	if s == nil || s.handle == nil {
		return SPPGeometryQuality{}, ErrClosed
	}
	value, err := s.handle.GeometryQuality()
	return publicRTKGeometry(value), publicError(err)
}

// References returns detached ambiguity-reference selections used by the solution.
func (s *RTKWideLaneArcSolution) References() ([]RTKArcReference, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.References()
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]RTKArcReference, len(value))
	for i, row := range value {
		result[i] = RTKArcReference{System: row.System, ReferenceID: row.ReferenceID}
	}
	return result, nil
}

// SplitCycleSlipArcs returns detached arc ranges created by cycle-slip splitting.
func (s *RTKWideLaneArcSolution) SplitCycleSlipArcs() ([]RTKArcSplitArc, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.SplitCycleSlipArcs()
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]RTKArcSplitArc, len(value))
	for i, row := range value {
		result[i] = RTKArcSplitArc{Receiver: RTKCycleSlipReceiver(row.Receiver), SatelliteID: row.SatelliteID, AmbiguityID: row.AmbiguityID, StartEpochIndex: row.StartEpochIndex, EndEpochIndex: row.EndEpochIndex, EpochCount: row.EpochCount}
	}
	return result, nil
}

// WideLaneCycles returns detached wide-lane ambiguity cycles.
func (s *RTKWideLaneArcSolution) WideLaneCycles() ([]RTKWideLaneCycle, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.WideLaneCycles()
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]RTKWideLaneCycle, len(value))
	for i, row := range value {
		result[i] = RTKWideLaneCycle{ID: row.ID, Cycles: row.Cycles}
	}
	return result, nil
}

// Close releases the native wide-lane-fixed RINEX RTK solution; repeated calls are safe.
func (s *RTKWideLaneFixedRINEXSolution) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return publicError(s.handle.Close())
}

// FixedBaselineECEF returns the fixed rover-minus-base baseline in ECEF metres.
func (s *RTKWideLaneFixedRINEXSolution) FixedBaselineECEF() ([3]float64, error) {
	if s == nil || s.handle == nil {
		return [3]float64{}, ErrClosed
	}
	value, err := s.handle.FixedBaselineECEF()
	return value, publicError(err)
}

// FixedMetadata returns metadata from the fixed-ambiguity RTK solution.
func (s *RTKWideLaneFixedRINEXSolution) FixedMetadata() (RTKFixedMetadata, error) {
	if s == nil || s.handle == nil {
		return RTKFixedMetadata{}, ErrClosed
	}
	value, err := s.handle.FixedMetadata()
	return publicRTKFixedMetadata(value), publicError(err)
}

// FloatBaselineECEF returns the float rover-minus-base baseline in ECEF metres.
func (s *RTKWideLaneFixedRINEXSolution) FloatBaselineECEF() ([3]float64, error) {
	if s == nil || s.handle == nil {
		return [3]float64{}, ErrClosed
	}
	value, err := s.handle.FloatBaselineECEF()
	return value, publicError(err)
}

// FloatMetadata returns metadata from the RTK float solution.
func (s *RTKWideLaneFixedRINEXSolution) FloatMetadata() (RTKFloatMetadata, error) {
	if s == nil || s.handle == nil {
		return RTKFloatMetadata{}, ErrClosed
	}
	value, err := s.handle.FloatMetadata()
	return publicRTKFloatMetadata(value), publicError(err)
}

// Metadata returns solver iteration, geometry, and status metadata.
func (s *RTKWideLaneFixedRINEXSolution) Metadata() (RTKWideLaneFixedRINEXMetadata, error) {
	if s == nil || s.handle == nil {
		return RTKWideLaneFixedRINEXMetadata{}, ErrClosed
	}
	value, err := s.handle.Metadata()
	return RTKWideLaneFixedRINEXMetadata{WideLaneFixed: value.WideLaneFixed, WideLaneAmbiguityCount: value.WideLaneAmbiguityCount, DroppedCycleSlipCount: value.DroppedCycleSlipCount, SplitCycleSlipCount: value.SplitCycleSlipCount}, publicError(err)
}

// WideLaneCycles returns detached wide-lane ambiguity cycles.
func (s *RTKWideLaneFixedRINEXSolution) WideLaneCycles() ([]RTKWideLaneCycle, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.WideLaneCycles()
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]RTKWideLaneCycle, len(value))
	for i, row := range value {
		result[i] = RTKWideLaneCycle{ID: row.ID, Cycles: row.Cycles}
	}
	return result, nil
}
