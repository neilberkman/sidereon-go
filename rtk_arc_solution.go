package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// RTKArcReferenceMode selects the reference satellite policy for sequential
// RTK and its derived arc products.
// RTKArcReferenceMode selects how RTK ambiguity references are chosen.
type RTKArcReferenceMode uint32

const (
	// RTKArcReferenceAuto selects automatic ambiguity-reference selection.
	RTKArcReferenceAuto RTKArcReferenceMode = 0
	// RTKArcReferenceSatellite selects single-satellite ambiguity-reference selection.
	RTKArcReferenceSatellite RTKArcReferenceMode = 1
	// RTKArcReferencePerSystem selects per-constellation ambiguity-reference selection.
	RTKArcReferencePerSystem RTKArcReferenceMode = 2
)

// RTKCycleSlipPolicy selects the response to a detected cycle slip.
type RTKCycleSlipPolicy uint32

const (
	// RTKCycleSlipError reports that cycle-slip handling failed.
	RTKCycleSlipError RTKCycleSlipPolicy = 0
	// RTKCycleSlipDropSatellite selects dropping the affected satellite on a cycle slip.
	RTKCycleSlipDropSatellite RTKCycleSlipPolicy = 1
	// RTKCycleSlipSplitArc selects splitting the arc on a cycle slip.
	RTKCycleSlipSplitArc RTKCycleSlipPolicy = 2
)

// RTKCycleSlipReceiver identifies whether a cycle slip belongs to the base or rover receiver.
type RTKCycleSlipReceiver uint32

const (
	// RTKCycleSlipBase selects the base receiver as cycle-slip source.
	RTKCycleSlipBase RTKCycleSlipReceiver = 0
	// RTKCycleSlipRover selects the rover receiver as cycle-slip source.
	RTKCycleSlipRover RTKCycleSlipReceiver = 1
)

// RTKArcEpochIDList selects one per-epoch ambiguity-id list.
type RTKArcEpochIDList uint32

const (
	// RTKArcEpochNewlyFixedIDs selects newly fixed ambiguity IDs.
	RTKArcEpochNewlyFixedIDs RTKArcEpochIDList = 0
	// RTKArcEpochFixedIDs selects all fixed ambiguity IDs.
	RTKArcEpochFixedIDs RTKArcEpochIDList = 1
	// RTKArcEpochFixedDoubleDifferenceIDs selects fixed double-difference ambiguity IDs.
	RTKArcEpochFixedDoubleDifferenceIDs RTKArcEpochIDList = 2
)

// RTKArcPositionEntry is one satellite-keyed ECEF position in meters.
type RTKArcPositionEntry struct {
	// SatelliteID is the GNSS satellite identifier.
	SatelliteID string
	// PositionM contains metres.
	PositionM [3]float64
}

// RTKArcEpoch is one raw sequential RTK epoch. Empty receiver-specific
// position slices use SatellitePositions, as specified by the C ABI.
// RTKArcEpoch contains base, rover, and satellite observations for one RTK epoch.
type RTKArcEpoch struct {
	Base, Rover             []RTKArcObservation
	SatellitePositions      []RTKArcPositionEntry
	BaseSatellitePositions  []RTKArcPositionEntry
	RoverSatellitePositions []RTKArcPositionEntry
	// HasVelocityMPS reports whether VelocityMPS is supplied.
	HasVelocityMPS bool
	// VelocityMPS contains metres per second.
	VelocityMPS [3]float64
	// HasPredictionTime reports whether PredictionTimeS is valid.
	HasPredictionTime bool
	// PredictionTimeS is the epoch time coordinate in seconds when HasPredictionTime is true.
	PredictionTimeS float64
}

// RTKArcReferenceEntry specifies a per-constellation reference satellite.
type RTKArcReferenceEntry struct {
	// System identifies the GNSS constellation or constellation set.
	System GNSSSystem
	// SatelliteID is the GNSS satellite identifier.
	SatelliteID string
}

// RTKArcPreprocessing controls optional cycle-slip, Hatch, and elevation
// preprocessing in a sequential solve.
// RTKArcPreprocessing configures cycle-slip and observation preprocessing.
type RTKArcPreprocessing struct {
	// HasCycleSlip reports whether CycleSlip is present.
	HasCycleSlip bool
	CycleSlip    RTKCycleSlipPolicy
	// HasHatchWindowCap reports whether HatchWindowCap is valid.
	HasHatchWindowCap bool
	// HatchWindowCap is the maximum Hatch smoothing window in epochs.
	HatchWindowCap int
	// HasElevationMaskDeg reports whether ElevationMaskDeg is valid.
	HasElevationMaskDeg bool
	// ElevationMaskDeg is the elevation mask in degrees.
	ElevationMaskDeg float64
}

// RTKArcConfig contains all inputs to a sequential RTK arc solve.
type RTKArcConfig struct {
	// BaseM contains metres.
	BaseM         [3]float64
	ReferenceMode RTKArcReferenceMode
	// ReferenceSatellite is the selected reference satellite identifier.
	ReferenceSatellite string
	ReferencePerSystem []RTKArcReferenceEntry
	Model              RTKMeasurementModel
	// BaselinePriorSigmaM contains metres.
	BaselinePriorSigmaM float64
	// AmbiguityPriorSigmaM contains metres.
	AmbiguityPriorSigmaM float64
	// InitialBaselineM contains metres.
	InitialBaselineM [3]float64
	Wavelengths      []RTKFloatMapEntry
	Offsets          []RTKFloatMapEntry
	UpdateOptions    RTKArcUpdateOptions
	// ReceiverAntenna refers to an optional value; nil means it is unavailable.
	ReceiverAntenna *RTKReceiverAntennaCorrections
	Preprocessing   RTKArcPreprocessing
}

// RTKArcEpochMetadata contains one copied sequential solution epoch.
type RTKArcEpochMetadata struct {
	// ReportedBaselineM and FloatBaselineM are rover-minus-base ECEF baselines in metres.
	ReportedBaselineM, FloatBaselineM [3]float64
	// IntegerFixed reports whether integer ambiguities were fixed.
	IntegerFixed bool
	// IntegerRatio is the ambiguity ratio.
	IntegerRatio                                   float64
	NewlyFixedCount, FixedIDCount                  int
	FixedDoubleDifferenceCount, UsedSatelliteCount int
	SDAmbiguityCount, ResidualCount                int
	// HasSearch reports whether integer-search diagnostics are available.
	HasSearch bool
}

// RTKArcReference is one copied per-constellation reference result.
type RTKArcReference struct {
	// System and ReferenceID identifies the GNSS constellation or constellation set.
	System, ReferenceID string
}

// RTKArcSplitArc is one copied cycle-slip split interval.
type RTKArcSplitArc struct {
	Receiver RTKCycleSlipReceiver
	// SatelliteID identifies the satellite; AmbiguityID identifies its carrier-phase ambiguity.
	SatelliteID, AmbiguityID       string
	StartEpochIndex, EndEpochIndex int
	EpochCount                     int
}

// RTKDualFrequencyArcEpoch is one dual-frequency epoch accepted by the
// wide-lane and ionosphere-free C routes.
// RTKDualFrequencyArcEpoch contains dual-frequency observations and satellite positions for one epoch.
type RTKDualFrequencyArcEpoch struct {
	// JDWhole and JDFraction are the split Julian-date day value and fractional day.
	JDWhole, JDFraction float64
	// HasEpochSortKey reports whether the epoch sort key is valid.
	HasEpochSortKey bool
	// EpochSortKey is the caller-provided key used to order epochs.
	EpochSortKey string
	// HasGapTimeS reports whether GapTimeS is valid.
	HasGapTimeS bool
	// GapTimeS is the cycle-slip gap threshold in seconds.
	GapTimeS                float64
	Observations            []RTKDualFrequencySatelliteObservation
	SatellitePositions      []RTKArcPositionEntry
	BaseSatellitePositions  []RTKArcPositionEntry
	RoverSatellitePositions []RTKArcPositionEntry
	// HasVelocityMPS reports whether VelocityMPS is supplied.
	HasVelocityMPS bool
	// VelocityMPS contains metres per second.
	VelocityMPS [3]float64
	// HasPredictionTime reports whether PredictionTimeS is valid.
	HasPredictionTime bool
	// PredictionTimeS is the epoch time coordinate in seconds when HasPredictionTime is true.
	PredictionTimeS float64
}

// RTKWideLaneCycle is one fixed Melbourne-Wubbena ambiguity in cycles.
type RTKWideLaneCycle struct {
	// ID identifies the associated record.
	ID string
	// Cycles is the carrier-cycle count.
	Cycles int64
}

// RTKWideLaneOptions controls wide-lane estimation.
type RTKWideLaneOptions struct {
	// MinEpochs is the minimum epoch count.
	MinEpochs int
	// ToleranceCycles is the cycle-slip tolerance in cycles.
	ToleranceCycles float64
	// SkipShort reports whether short arcs are skipped.
	SkipShort bool
}

// RTKWideLaneArcConfig configures the wide-lane solver.
type RTKWideLaneArcConfig struct {
	// BaseM contains metres.
	BaseM         [3]float64
	ReferenceMode RTKArcReferenceMode
	// ReferenceSatellite is the selected reference satellite identifier.
	ReferenceSatellite string
	ReferencePerSystem []RTKArcReferenceEntry
	Options            RTKWideLaneOptions
	// HasCycleSlip reports whether CycleSlip is present.
	HasCycleSlip     bool
	CycleSlipPolicy  RTKCycleSlipPolicy
	CycleSlipOptions CycleSlipOptions
}

// RTKIonosphereFreeArcConfig configures ionosphere-free preparation.
type RTKIonosphereFreeArcConfig struct {
	// BaseM contains metres.
	BaseM [3]float64
	// InitialBaselineM contains metres.
	InitialBaselineM [3]float64
	ReferenceMode    RTKArcReferenceMode
	// ReferenceSatellite is the selected reference satellite identifier.
	ReferenceSatellite string
	ReferencePerSystem []RTKArcReferenceEntry
	// ApplyTroposphere reports whether the troposphere correction is enabled.
	ApplyTroposphere bool
}

// RTKArcSolution owns a sequential RTK result and must not be copied after
// first use.
// RTKArcSolution owns a native sequential RTK arc solution.
type RTKArcSolution struct {
	_      noCopy
	handle *native.RtkArcSolution
}

// RTKIonosphereFreeArcSolution owns an ionosphere-free RTK result and must not
// be copied after first use.
// RTKIonosphereFreeArcSolution owns a native ionosphere-free RTK arc solution.
type RTKIonosphereFreeArcSolution struct {
	_      noCopy
	handle *native.RtkIonosphereFreeArcSolution
}

// RTKWideLaneArcSolution owns a wide-lane result and must not be copied after
// first use.
// RTKWideLaneArcSolution owns a native wide-lane RTK arc solution.
type RTKWideLaneArcSolution struct {
	_      noCopy
	handle *native.RtkWideLaneArcSolution
}

func nativeRTKArcEpoch(value RTKArcEpoch) native.RtkArcEpochInput {
	result := native.RtkArcEpochInput{HasVelocityMPS: value.HasVelocityMPS, VelocityMPS: value.VelocityMPS, HasPredictionTime: value.HasPredictionTime, PredictionTimeS: value.PredictionTimeS}
	result.Base = append([]native.RtkArcObservationInput(nil), nativeRTKArcObservations(value.Base)...)
	result.Rover = append([]native.RtkArcObservationInput(nil), nativeRTKArcObservations(value.Rover)...)
	result.SatellitePositions = nativeRTKArcPositions(value.SatellitePositions)
	result.BaseSatellitePositions = nativeRTKArcPositions(value.BaseSatellitePositions)
	result.RoverSatellitePositions = nativeRTKArcPositions(value.RoverSatellitePositions)
	return result
}

func nativeRTKArcObservations(values []RTKArcObservation) []native.RtkArcObservationInput {
	result := make([]native.RtkArcObservationInput, len(values))
	for index, value := range values {
		result[index] = native.RtkArcObservationInput{SatelliteID: value.SatelliteID, AmbiguityID: value.AmbiguityID, CodeM: value.CodeM, PhaseM: value.PhaseM, HasLLI: value.HasLLI, LLI: value.LLI}
	}
	return result
}

func nativeRTKArcPositions(values []RTKArcPositionEntry) []native.RtkArcPositionEntry {
	result := make([]native.RtkArcPositionEntry, len(values))
	for index, value := range values {
		result[index] = native.RtkArcPositionEntry{ID: value.SatelliteID, PositionM: value.PositionM}
	}
	return result
}

func nativeRTKArcReferences(values []RTKArcReferenceEntry) []native.RtkArcReferenceEntry {
	result := make([]native.RtkArcReferenceEntry, len(values))
	for index, value := range values {
		result[index] = native.RtkArcReferenceEntry{System: uint32(value.System), SatelliteID: value.SatelliteID}
	}
	return result
}

func nativeRTKArcConfig(value RTKArcConfig) native.RtkArcConfigInput {
	return native.RtkArcConfigInput{BaseM: value.BaseM, ReferenceMode: uint32(value.ReferenceMode), ReferenceSatellite: value.ReferenceSatellite, ReferencePerSystem: nativeRTKArcReferences(value.ReferencePerSystem), Model: native.RtkMeasurementModel{CodeSigmaM: value.Model.CodeSigmaM, PhaseSigmaM: value.Model.PhaseSigmaM, Sagnac: value.Model.Sagnac, Stochastic: value.Model.Stochastic, ElevationWeighting: value.Model.ElevationWeighting}, BaselinePriorSigmaM: value.BaselinePriorSigmaM, AmbiguityPriorSigmaM: value.AmbiguityPriorSigmaM, InitialBaselineM: value.InitialBaselineM, Wavelengths: append([]native.RtkFloatMapEntry(nil), nativeRTKFloatMapEntries(value.Wavelengths)...), Offsets: append([]native.RtkFloatMapEntry(nil), nativeRTKFloatMapEntries(value.Offsets)...), UpdateOptions: native.RtkArcUpdateOptions{HoldSigmaM: value.UpdateOptions.HoldSigmaM, PositionTolM: value.UpdateOptions.PositionTolM, AmbiguityTolM: value.UpdateOptions.AmbiguityTolM, MaxIterations: value.UpdateOptions.MaxIterations, ProcessNoiseBaselineSigmaM: value.UpdateOptions.ProcessNoiseBaselineSigmaM, DynamicsVelocityPropagated: value.UpdateOptions.DynamicsVelocityPropagated, FloatOnlySystems: nativeRTKSystems(value.UpdateOptions.FloatOnlySystems), ReportResiduals: value.UpdateOptions.ReportResiduals, HasARArmingSigmaM: value.UpdateOptions.HasARArmingSigmaM, ARArmingSigmaM: value.UpdateOptions.ARArmingSigmaM, RatioThreshold: value.UpdateOptions.RatioThreshold, ReceiverAntenna: nativeRTKReceiverAntenna(value.UpdateOptions.ReceiverAntenna)}, ReceiverAntenna: nativeRTKReceiverAntenna(value.ReceiverAntenna), Preprocessing: native.RtkArcPreprocessing{HasCycleSlip: value.Preprocessing.HasCycleSlip, CycleSlip: uint32(value.Preprocessing.CycleSlip), HasHatchWindowCap: value.Preprocessing.HasHatchWindowCap, HatchWindowCap: value.Preprocessing.HatchWindowCap, HasElevationMaskDeg: value.Preprocessing.HasElevationMaskDeg, ElevationMaskDeg: value.Preprocessing.ElevationMaskDeg}}
}

func nativeRTKFloatMapEntries(values []RTKFloatMapEntry) []native.RtkFloatMapEntry {
	result := make([]native.RtkFloatMapEntry, len(values))
	for index, value := range values {
		result[index] = native.RtkFloatMapEntry{ID: value.ID, Value: value.Value}
	}
	return result
}

func nativeRTKSystems(values []GNSSSystem) []uint32 {
	result := make([]uint32, len(values))
	for index, value := range values {
		result[index] = uint32(value)
	}
	return result
}

func nativeRTKReceiverAntenna(value *RTKReceiverAntennaCorrections) *native.RtkReceiverAntennaCorrections {
	if value == nil {
		return nil
	}
	result := &native.RtkReceiverAntennaCorrections{}
	result.Base = native.RtkReceiverAntennaCalibration{PCONEUM: value.Base.PCONEUM}
	result.Base.NoAzimuthPCV = make([]native.RtkReceiverAntennaNoAzimuthPCV, len(value.Base.NoAzimuthPCV))
	for index, sample := range value.Base.NoAzimuthPCV {
		result.Base.NoAzimuthPCV[index] = native.RtkReceiverAntennaNoAzimuthPCV{ZenithDeg: sample.ZenithDeg, ValueM: sample.ValueM}
	}
	result.Base.AzimuthPCV = make([]native.RtkReceiverAntennaAzimuthPCV, len(value.Base.AzimuthPCV))
	for index, sample := range value.Base.AzimuthPCV {
		result.Base.AzimuthPCV[index] = native.RtkReceiverAntennaAzimuthPCV{AzimuthDeg: sample.AzimuthDeg, ZenithDeg: sample.ZenithDeg, ValueM: sample.ValueM}
	}
	result.Rover = native.RtkReceiverAntennaCalibration{PCONEUM: value.Rover.PCONEUM}
	result.Rover.NoAzimuthPCV = make([]native.RtkReceiverAntennaNoAzimuthPCV, len(value.Rover.NoAzimuthPCV))
	for index, sample := range value.Rover.NoAzimuthPCV {
		result.Rover.NoAzimuthPCV[index] = native.RtkReceiverAntennaNoAzimuthPCV{ZenithDeg: sample.ZenithDeg, ValueM: sample.ValueM}
	}
	result.Rover.AzimuthPCV = make([]native.RtkReceiverAntennaAzimuthPCV, len(value.Rover.AzimuthPCV))
	for index, sample := range value.Rover.AzimuthPCV {
		result.Rover.AzimuthPCV[index] = native.RtkReceiverAntennaAzimuthPCV{AzimuthDeg: sample.AzimuthDeg, ZenithDeg: sample.ZenithDeg, ValueM: sample.ValueM}
	}
	return result
}

func nativeRTKDualEpoch(value RTKDualFrequencyArcEpoch) native.RtkDualFrequencyArcEpochInput {
	result := native.RtkDualFrequencyArcEpochInput{JDWhole: value.JDWhole, JDFraction: value.JDFraction, HasEpochSortKey: value.HasEpochSortKey, EpochSortKey: value.EpochSortKey, HasGapTimeS: value.HasGapTimeS, GapTimeS: value.GapTimeS, HasVelocityMPS: value.HasVelocityMPS, VelocityMPS: value.VelocityMPS, HasPredictionTime: value.HasPredictionTime, PredictionTimeS: value.PredictionTimeS, SatellitePositions: nativeRTKArcPositions(value.SatellitePositions), BaseSatellitePositions: nativeRTKArcPositions(value.BaseSatellitePositions), RoverSatellitePositions: nativeRTKArcPositions(value.RoverSatellitePositions)}
	result.Observations = make([]native.RtkDualFrequencySatelliteObservationInput, len(value.Observations))
	for index, observation := range value.Observations {
		result.Observations[index] = native.RtkDualFrequencySatelliteObservationInput{SatelliteID: observation.SatelliteID, Base: nativeRTKDualObservation(observation.Base), Rover: nativeRTKDualObservation(observation.Rover)}
	}
	return result
}

func nativeRTKDualObservation(value RTKDualFrequencyObservation) native.RtkDualFrequencyObservationInput {
	return native.RtkDualFrequencyObservationInput{AmbiguityID: value.AmbiguityID, P1M: value.P1M, P2M: value.P2M, Phi1Cycles: value.Phi1Cycles, Phi2Cycles: value.Phi2Cycles, F1Hz: value.F1Hz, F2Hz: value.F2Hz, HasLLI1: value.HasLLI1, LLI1: value.LLI1, HasLLI2: value.HasLLI2, LLI2: value.LLI2}
}

func nativeRTKWideConfig(value RTKWideLaneArcConfig) native.RtkWideLaneConfigInput {
	return native.RtkWideLaneConfigInput{BaseM: value.BaseM, ReferenceMode: uint32(value.ReferenceMode), ReferenceSatellite: value.ReferenceSatellite, ReferencePerSystem: nativeRTKArcReferences(value.ReferencePerSystem), Options: native.RtkWideLaneOptions{MinEpochs: value.Options.MinEpochs, ToleranceCycles: value.Options.ToleranceCycles, SkipShort: value.Options.SkipShort}, HasCycleSlip: value.HasCycleSlip, CycleSlipPolicy: uint32(value.CycleSlipPolicy), CycleSlipOptions: native.CycleSlipOptions{GFThresholdM: value.CycleSlipOptions.GFThresholdM, MWThresholdCycles: value.CycleSlipOptions.MWThresholdCycles, MinArcGapS: value.CycleSlipOptions.MinArcGapS}}
}

func nativeRTKIonosphereFreeConfig(value RTKIonosphereFreeArcConfig) native.RtkIonosphereFreeConfigInput {
	return native.RtkIonosphereFreeConfigInput{BaseM: value.BaseM, InitialBaselineM: value.InitialBaselineM, ReferenceMode: uint32(value.ReferenceMode), ReferenceSatellite: value.ReferenceSatellite, ReferencePerSystem: nativeRTKArcReferences(value.ReferencePerSystem), ApplyTroposphere: value.ApplyTroposphere}
}

func nativeRTKDualEpochs(values []RTKDualFrequencyArcEpoch) []native.RtkDualFrequencyArcEpochInput {
	result := make([]native.RtkDualFrequencyArcEpochInput, len(values))
	for index, value := range values {
		result[index] = nativeRTKDualEpoch(value)
	}
	return result
}

// SolveRTKArc delegates sequential RTK solving to C.
func SolveRTKArc(epochs []RTKArcEpoch, config RTKArcConfig) (*RTKArcSolution, error) {
	nativeEpochs := make([]native.RtkArcEpochInput, len(epochs))
	for index, value := range epochs {
		nativeEpochs[index] = nativeRTKArcEpoch(value)
	}
	result, err := native.SolveRtkArc(nativeEpochs, nativeRTKArcConfig(config))
	if err != nil {
		return nil, publicError(err)
	}
	return &RTKArcSolution{handle: result}, nil
}

// FixWideLaneRTKArc delegates wide-lane fixing to C.
func FixWideLaneRTKArc(epochs []RTKDualFrequencyArcEpoch, config RTKWideLaneArcConfig) (*RTKWideLaneArcSolution, error) {
	result, err := native.FixWideLaneRtkArc(nativeRTKDualEpochs(epochs), nativeRTKWideConfig(config))
	if err != nil {
		return nil, publicError(err)
	}
	return &RTKWideLaneArcSolution{handle: result}, nil
}

// PrepareIonosphereFreeRTKArc delegates ionosphere-free preparation to C.
func PrepareIonosphereFreeRTKArc(epochs []RTKDualFrequencyArcEpoch, cycles []RTKWideLaneCycle, config RTKIonosphereFreeArcConfig) (*RTKIonosphereFreeArcSolution, error) {
	nativeCycles := make([]native.RtkWideLaneCycleInput, len(cycles))
	for index, value := range cycles {
		nativeCycles[index] = native.RtkWideLaneCycleInput{ID: value.ID, Cycles: value.Cycles}
	}
	result, err := native.PrepareIonosphereFreeRtkArc(nativeRTKDualEpochs(epochs), nativeCycles, nativeRTKIonosphereFreeConfig(config))
	if err != nil {
		return nil, publicError(err)
	}
	return &RTKIonosphereFreeArcSolution{handle: result}, nil
}

// Close releases a sequential RTK solution.
func (s *RTKArcSolution) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return publicError(s.handle.Close())
}

// Close releases an ionosphere-free RTK solution.
func (s *RTKIonosphereFreeArcSolution) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return publicError(s.handle.Close())
}

// Close releases a wide-lane RTK solution.
func (s *RTKWideLaneArcSolution) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return publicError(s.handle.Close())
}

// EpochCount returns the number of recorded epochs.
func (s *RTKArcSolution) EpochCount() (int, error) {
	if s == nil || s.handle == nil {
		return 0, ErrClosed
	}
	value, err := s.handle.EpochCount()
	return value, publicError(err)
}

// FinalEpochCount returns the number of epochs retained in the final RTK arc.
func (s *RTKArcSolution) FinalEpochCount() (int, error) {
	if s == nil || s.handle == nil {
		return 0, ErrClosed
	}
	value, err := s.handle.FinalEpochCount()
	return value, publicError(err)
}

// EpochMetadata returns ambiguity, residual, and used-satellite counts for the selected epoch.
func (s *RTKArcSolution) EpochMetadata(index int) (RTKArcEpochMetadata, error) {
	if s == nil || s.handle == nil {
		return RTKArcEpochMetadata{}, ErrClosed
	}
	value, err := s.handle.EpochMetadata(index)
	if err != nil {
		return RTKArcEpochMetadata{}, publicError(err)
	}
	return RTKArcEpochMetadata{ReportedBaselineM: value.ReportedBaselineM, FloatBaselineM: value.FloatBaselineM, IntegerFixed: value.IntegerFixed, IntegerRatio: value.IntegerRatio, NewlyFixedCount: value.NewlyFixedCount, FixedIDCount: value.FixedIDCount, FixedDoubleDifferenceCount: value.FixedDoubleDifferenceCount, UsedSatelliteCount: value.UsedSatelliteCount, SDAmbiguityCount: value.SDAmbiguityCount, ResidualCount: value.ResidualCount, HasSearch: value.HasSearch}, nil
}

// EpochSDAmbiguities returns detached single-difference ambiguities for the selected epoch.
func (s *RTKArcSolution) EpochSDAmbiguities(index int) ([]RTKAmbiguity, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.EpochSDAmbiguities(index)
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]RTKAmbiguity, len(value))
	for index, row := range value {
		result[index] = RTKAmbiguity{ID: row.ID, ValueM: row.ValueM}
	}
	return result, nil
}

// EpochStringIDs returns the selected detached satellite-ID list for the epoch.
func (s *RTKArcSolution) EpochStringIDs(index int, which RTKArcEpochIDList) ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.EpochStringIDs(index, uint32(which))
	return append([]string(nil), value...), publicError(err)
}

// EpochUsedSatellites returns detached IDs of satellites used at the selected epoch.
func (s *RTKArcSolution) EpochUsedSatellites(index int) ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.EpochUsedSatellites(index)
	return append([]string(nil), value...), publicError(err)
}

// DroppedSatellites returns detached IDs of satellites dropped from the solution.
func (s *RTKArcSolution) DroppedSatellites() ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.DroppedSatellites()
	return append([]string(nil), value...), publicError(err)
}

// ElevationMaskedSatellites returns detached IDs rejected by the elevation mask.
func (s *RTKArcSolution) ElevationMaskedSatellites() ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.ElevationMaskedSatellites()
	return append([]string(nil), value...), publicError(err)
}

// FinalBaseline returns the final rover-minus-base baseline in ECEF metres.
func (s *RTKArcSolution) FinalBaseline() ([3]float64, error) {
	if s == nil || s.handle == nil {
		return [3]float64{}, ErrClosed
	}
	value, err := s.handle.FinalBaseline()
	return value, publicError(err)
}

// MeasurementCovariance returns detached final measurement covariance in row-major order.
func (s *RTKArcSolution) MeasurementCovariance() ([]float64, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.MeasurementCovariance()
	return append([]float64(nil), value...), publicError(err)
}

// References returns detached ambiguity-reference selections used by the solution.
func (s *RTKArcSolution) References() ([]RTKArcReference, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.References()
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]RTKArcReference, len(value))
	for index, row := range value {
		result[index] = RTKArcReference{System: row.System, ReferenceID: row.ReferenceID}
	}
	return result, nil
}

// SplitCycleSlipArcs returns detached arc ranges created by cycle-slip splitting.
func (s *RTKArcSolution) SplitCycleSlipArcs() ([]RTKArcSplitArc, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.SplitCycleSlipArcs()
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]RTKArcSplitArc, len(value))
	for index, row := range value {
		result[index] = RTKArcSplitArc{Receiver: RTKCycleSlipReceiver(row.Receiver), SatelliteID: row.SatelliteID, AmbiguityID: row.AmbiguityID, StartEpochIndex: row.StartEpochIndex, EndEpochIndex: row.EndEpochIndex, EpochCount: row.EpochCount}
	}
	return result, nil
}

// EpochCount returns the number of recorded epochs.
func (s *RTKIonosphereFreeArcSolution) EpochCount() (int, error) {
	if s == nil || s.handle == nil {
		return 0, ErrClosed
	}
	value, err := s.handle.EpochCount()
	return value, publicError(err)
}

// EpochBaseObservations returns detached base observations for the selected epoch.
func (s *RTKIonosphereFreeArcSolution) EpochBaseObservations(index int) ([]RTKArcObservation, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.EpochBaseObservations(index)
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]RTKArcObservation, len(value))
	for index, row := range value {
		result[index] = RTKArcObservation{SatelliteID: row.SatelliteID, AmbiguityID: row.AmbiguityID, CodeM: row.CodeM, PhaseM: row.PhaseM, HasLLI: row.HasLLI, LLI: row.LLI}
	}
	return result, nil
}

// EpochRoverObservations returns detached rover observations for the selected epoch.
func (s *RTKIonosphereFreeArcSolution) EpochRoverObservations(index int) ([]RTKArcObservation, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.EpochRoverObservations(index)
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]RTKArcObservation, len(value))
	for index, row := range value {
		result[index] = RTKArcObservation{SatelliteID: row.SatelliteID, AmbiguityID: row.AmbiguityID, CodeM: row.CodeM, PhaseM: row.PhaseM, HasLLI: row.HasLLI, LLI: row.LLI}
	}
	return result, nil
}

func publicRTKArcPositions(value []native.RtkRinexArcPosition) []RTKArcPosition {
	result := make([]RTKArcPosition, len(value))
	for index, row := range value {
		result[index] = RTKArcPosition{SatelliteID: row.SatelliteID, PositionM: row.PositionM}
	}
	return result
}

// EpochSatellitePositions returns detached satellite ECEF positions for the selected epoch.
func (s *RTKIonosphereFreeArcSolution) EpochSatellitePositions(index int) ([]RTKArcPosition, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.EpochSatellitePositions(index)
	if err != nil {
		return nil, publicError(err)
	}
	return publicRTKArcPositions(value), nil
}

// EpochBaseSatellitePositions returns detached base-side satellite positions for the selected epoch.
func (s *RTKIonosphereFreeArcSolution) EpochBaseSatellitePositions(index int) ([]RTKArcPosition, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.EpochBaseSatellitePositions(index)
	if err != nil {
		return nil, publicError(err)
	}
	return publicRTKArcPositions(value), nil
}

// EpochRoverSatellitePositions returns detached rover-side satellite positions for the selected epoch.
func (s *RTKIonosphereFreeArcSolution) EpochRoverSatellitePositions(index int) ([]RTKArcPosition, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.EpochRoverSatellitePositions(index)
	if err != nil {
		return nil, publicError(err)
	}
	return publicRTKArcPositions(value), nil
}

// EpochMetadata returns ambiguity, residual, and used-satellite counts for the selected epoch.
func (s *RTKIonosphereFreeArcSolution) EpochMetadata(index int) (RTKRINEXArcEpochMetadata, error) {
	if s == nil || s.handle == nil {
		return RTKRINEXArcEpochMetadata{}, ErrClosed
	}
	value, err := s.handle.EpochMetadata(index)
	if err != nil {
		return RTKRINEXArcEpochMetadata{}, publicError(err)
	}
	return RTKRINEXArcEpochMetadata{BaseCount: value.BaseCount, RoverCount: value.RoverCount, SatellitePositionCount: value.SatellitePositionCount, BaseSatellitePositionCount: value.BaseSatellitePositionCount, RoverSatellitePositionCount: value.RoverSatellitePositionCount, HasVelocityMPS: value.HasVelocityMPS, VelocityMPS: value.VelocityMPS, HasPredictionTime: value.HasPredictionTime, PredictionTimeS: value.PredictionTimeS}, nil
}

// OffsetsM returns detached ionosphere-free code offsets in metres.
func (s *RTKIonosphereFreeArcSolution) OffsetsM() ([]RTKFloatMapEntry, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.OffsetsM()
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]RTKFloatMapEntry, len(value))
	for index, row := range value {
		result[index] = RTKFloatMapEntry{ID: row.ID, Value: row.Value}
	}
	return result, nil
}

// WavelengthsM returns detached ionosphere-free carrier wavelengths in metres.
func (s *RTKIonosphereFreeArcSolution) WavelengthsM() ([]RTKFloatMapEntry, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.WavelengthsM()
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]RTKFloatMapEntry, len(value))
	for index, row := range value {
		result[index] = RTKFloatMapEntry{ID: row.ID, Value: row.Value}
	}
	return result, nil
}

// References returns detached ambiguity-reference selections used by the solution.
func (s *RTKIonosphereFreeArcSolution) References() ([]RTKArcReference, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.References()
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]RTKArcReference, len(value))
	for index, row := range value {
		result[index] = RTKArcReference{System: row.System, ReferenceID: row.ReferenceID}
	}
	return result, nil
}
