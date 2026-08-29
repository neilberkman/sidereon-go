package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// RTKArcReferenceMode selects the reference satellite policy for sequential
// RTK and its derived arc products.
type RTKArcReferenceMode uint32

const (
	RTKArcReferenceAuto      RTKArcReferenceMode = 0
	RTKArcReferenceSatellite RTKArcReferenceMode = 1
	RTKArcReferencePerSystem RTKArcReferenceMode = 2
)

// RTKCycleSlipPolicy selects the response to a detected cycle slip.
type RTKCycleSlipPolicy uint32

const (
	RTKCycleSlipError         RTKCycleSlipPolicy = 0
	RTKCycleSlipDropSatellite RTKCycleSlipPolicy = 1
	RTKCycleSlipSplitArc      RTKCycleSlipPolicy = 2
)

// RTKCycleSlipReceiver identifies the receiver side of a split arc.
type RTKCycleSlipReceiver uint32

const (
	RTKCycleSlipBase  RTKCycleSlipReceiver = 0
	RTKCycleSlipRover RTKCycleSlipReceiver = 1
)

// RTKArcEpochIDList selects one per-epoch ambiguity-id list.
type RTKArcEpochIDList uint32

const (
	RTKArcEpochNewlyFixedIDs            RTKArcEpochIDList = 0
	RTKArcEpochFixedIDs                 RTKArcEpochIDList = 1
	RTKArcEpochFixedDoubleDifferenceIDs RTKArcEpochIDList = 2
)

// RTKArcPositionEntry is one satellite-keyed ECEF position in meters.
type RTKArcPositionEntry struct {
	SatelliteID string
	PositionM   [3]float64
}

// RTKArcEpoch is one raw sequential RTK epoch. Empty receiver-specific
// position slices use SatellitePositions, as specified by the C ABI.
type RTKArcEpoch struct {
	Base, Rover             []RTKArcObservation
	SatellitePositions      []RTKArcPositionEntry
	BaseSatellitePositions  []RTKArcPositionEntry
	RoverSatellitePositions []RTKArcPositionEntry
	HasVelocityMPS          bool
	VelocityMPS             [3]float64
	HasPredictionTime       bool
	PredictionTimeS         float64
}

// RTKArcReferenceEntry specifies a per-constellation reference satellite.
type RTKArcReferenceEntry struct {
	System      GNSSSystem
	SatelliteID string
}

// RTKArcPreprocessing controls optional cycle-slip, Hatch, and elevation
// preprocessing in a sequential solve.
type RTKArcPreprocessing struct {
	HasCycleSlip        bool
	CycleSlip           RTKCycleSlipPolicy
	HasHatchWindowCap   bool
	HatchWindowCap      int
	HasElevationMaskDeg bool
	ElevationMaskDeg    float64
}

// RTKArcConfig contains all inputs to a sequential RTK arc solve.
type RTKArcConfig struct {
	BaseM                [3]float64
	ReferenceMode        RTKArcReferenceMode
	ReferenceSatellite   string
	ReferencePerSystem   []RTKArcReferenceEntry
	Model                RTKMeasurementModel
	BaselinePriorSigmaM  float64
	AmbiguityPriorSigmaM float64
	InitialBaselineM     [3]float64
	Wavelengths          []RTKFloatMapEntry
	Offsets              []RTKFloatMapEntry
	UpdateOptions        RTKArcUpdateOptions
	ReceiverAntenna      *RTKReceiverAntennaCorrections
	Preprocessing        RTKArcPreprocessing
}

// RTKArcEpochMetadata contains one copied sequential solution epoch.
type RTKArcEpochMetadata struct {
	ReportedBaselineM, FloatBaselineM              [3]float64
	IntegerFixed                                   bool
	IntegerRatio                                   float64
	NewlyFixedCount, FixedIDCount                  int
	FixedDoubleDifferenceCount, UsedSatelliteCount int
	SDAmbiguityCount, ResidualCount                int
	HasSearch                                      bool
}

// RTKArcReference is one copied per-constellation reference result.
type RTKArcReference struct {
	System, ReferenceID string
}

// RTKArcSplitArc is one copied cycle-slip split interval.
type RTKArcSplitArc struct {
	Receiver                       RTKCycleSlipReceiver
	SatelliteID, AmbiguityID       string
	StartEpochIndex, EndEpochIndex int
	EpochCount                     int
}

// RTKDualFrequencyArcEpoch is one dual-frequency epoch accepted by the
// wide-lane and ionosphere-free C routes.
type RTKDualFrequencyArcEpoch struct {
	JDWhole, JDFraction     float64
	HasEpochSortKey         bool
	EpochSortKey            string
	HasGapTimeS             bool
	GapTimeS                float64
	Observations            []RTKDualFrequencySatelliteObservation
	SatellitePositions      []RTKArcPositionEntry
	BaseSatellitePositions  []RTKArcPositionEntry
	RoverSatellitePositions []RTKArcPositionEntry
	HasVelocityMPS          bool
	VelocityMPS             [3]float64
	HasPredictionTime       bool
	PredictionTimeS         float64
}

// RTKWideLaneCycle is one fixed Melbourne-Wubbena ambiguity in cycles.
type RTKWideLaneCycle struct {
	ID     string
	Cycles int64
}

// RTKWideLaneOptions controls wide-lane estimation.
type RTKWideLaneOptions struct {
	MinEpochs       int
	ToleranceCycles float64
	SkipShort       bool
}

// RTKWideLaneArcConfig configures the wide-lane solver.
type RTKWideLaneArcConfig struct {
	BaseM              [3]float64
	ReferenceMode      RTKArcReferenceMode
	ReferenceSatellite string
	ReferencePerSystem []RTKArcReferenceEntry
	Options            RTKWideLaneOptions
	HasCycleSlip       bool
	CycleSlipPolicy    RTKCycleSlipPolicy
	CycleSlipOptions   CycleSlipOptions
}

// RTKIonosphereFreeArcConfig configures ionosphere-free preparation.
type RTKIonosphereFreeArcConfig struct {
	BaseM              [3]float64
	InitialBaselineM   [3]float64
	ReferenceMode      RTKArcReferenceMode
	ReferenceSatellite string
	ReferencePerSystem []RTKArcReferenceEntry
	ApplyTroposphere   bool
}

// RTKArcSolution owns a sequential RTK result and must not be copied after
// first use.
type RTKArcSolution struct {
	_      noCopy
	handle *native.RtkArcSolution
}

// RTKIonosphereFreeArcSolution owns an ionosphere-free RTK result and must not
// be copied after first use.
type RTKIonosphereFreeArcSolution struct {
	_      noCopy
	handle *native.RtkIonosphereFreeArcSolution
}

// RTKWideLaneArcSolution owns a wide-lane result and must not be copied after
// first use.
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
	return native.RtkArcConfigInput{BaseM: value.BaseM, ReferenceMode: uint32(value.ReferenceMode), ReferenceSatellite: value.ReferenceSatellite, ReferencePerSystem: nativeRTKArcReferences(value.ReferencePerSystem), Model: native.RtkMeasurementModel{CodeSigmaM: value.Model.CodeSigmaM, PhaseSigmaM: value.Model.PhaseSigmaM, Sagnac: value.Model.Sagnac, Stochastic: value.Model.Stochastic, ElevationWeighting: value.Model.ElevationWeighting}, BaselinePriorSigmaM: value.BaselinePriorSigmaM, AmbiguityPriorSigmaM: value.AmbiguityPriorSigmaM, InitialBaselineM: value.InitialBaselineM, Wavelengths: append([]native.RtkFloatMapEntry(nil), nativeRTKFloatMapEntries(value.Wavelengths)...), Offsets: append([]native.RtkFloatMapEntry(nil), nativeRTKFloatMapEntries(value.Offsets)...), UpdateOptions: native.RtkArcUpdateOptions{HoldSigmaM: value.UpdateOptions.HoldSigmaM, PositionTolM: value.UpdateOptions.PositionTolM, AmbiguityTolM: value.UpdateOptions.AmbiguityTolM, MaxIterations: value.UpdateOptions.MaxIterations, ProcessNoiseBaselineSigmaM: value.UpdateOptions.ProcessNoiseBaselineSigmaM, DynamicsVelocityPropagated: value.UpdateOptions.DynamicsVelocityPropagated, FloatOnlySystems: nativeRTKSystems(value.UpdateOptions.FloatOnlySystems), ReportResiduals: value.UpdateOptions.ReportResiduals, HasARArmingSigmaM: value.UpdateOptions.HasARArmingSigmaM, ARArmingSigmaM: value.UpdateOptions.ARArmingSigmaM, RatioThreshold: value.UpdateOptions.RatioThreshold}, ReceiverAntenna: nativeRTKReceiverAntenna(value.ReceiverAntenna), Preprocessing: native.RtkArcPreprocessing{HasCycleSlip: value.Preprocessing.HasCycleSlip, CycleSlip: uint32(value.Preprocessing.CycleSlip), HasHatchWindowCap: value.Preprocessing.HasHatchWindowCap, HatchWindowCap: value.Preprocessing.HatchWindowCap, HasElevationMaskDeg: value.Preprocessing.HasElevationMaskDeg, ElevationMaskDeg: value.Preprocessing.ElevationMaskDeg}}
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

func (s *RTKArcSolution) EpochCount() (int, error) {
	if s == nil || s.handle == nil {
		return 0, ErrClosed
	}
	value, err := s.handle.EpochCount()
	return value, publicError(err)
}

func (s *RTKArcSolution) FinalEpochCount() (int, error) {
	if s == nil || s.handle == nil {
		return 0, ErrClosed
	}
	value, err := s.handle.FinalEpochCount()
	return value, publicError(err)
}

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

func (s *RTKArcSolution) EpochStringIDs(index int, which RTKArcEpochIDList) ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.EpochStringIDs(index, uint32(which))
	return append([]string(nil), value...), publicError(err)
}

func (s *RTKArcSolution) EpochUsedSatellites(index int) ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.EpochUsedSatellites(index)
	return append([]string(nil), value...), publicError(err)
}

func (s *RTKArcSolution) DroppedSatellites() ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.DroppedSatellites()
	return append([]string(nil), value...), publicError(err)
}

func (s *RTKArcSolution) ElevationMaskedSatellites() ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.ElevationMaskedSatellites()
	return append([]string(nil), value...), publicError(err)
}

func (s *RTKArcSolution) FinalBaseline() ([3]float64, error) {
	if s == nil || s.handle == nil {
		return [3]float64{}, ErrClosed
	}
	value, err := s.handle.FinalBaseline()
	return value, publicError(err)
}

func (s *RTKArcSolution) MeasurementCovariance() ([]float64, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.MeasurementCovariance()
	return append([]float64(nil), value...), publicError(err)
}

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

func (s *RTKIonosphereFreeArcSolution) EpochCount() (int, error) {
	if s == nil || s.handle == nil {
		return 0, ErrClosed
	}
	value, err := s.handle.EpochCount()
	return value, publicError(err)
}

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
