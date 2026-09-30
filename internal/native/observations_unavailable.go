//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

// These declarations preserve source compatibility on platforms where the
// pinned static C ABI is unavailable. They never emulate parsing or numeric
// behavior.

type NativeCalendarEpoch struct {
	Year, Month, Day, Hour, Minute int
	Second                         float64
}
type NativeRinexObsHeader struct {
	Version                                                           float64
	HasApproxPosition                                                 bool
	ApproxPosition                                                    [3]float64
	HasAntennaDelta                                                   bool
	AntennaDelta                                                      [3]float64
	HasInterval                                                       bool
	Interval                                                          float64
	HasTimeOfFirstObs                                                 bool
	TimeOfFirstObs                                                    NativeCalendarEpoch
	TimeOfFirstObsScale                                               uint32
	ObsCodeCount, PhaseShiftCount, ScaleFactorCount, GLONASSSlotCount int
	HasMarkerName                                                     bool
	MarkerName                                                        string
}
type NativeObservableRowError struct {
	Index       uint64
	Status      uint32
	Payload     []byte
	Engine      *EngineError
	DecodeError error
}
type NativeObservableStateRow struct {
	Position      [3]float64
	ClockS        float64
	HasClock      bool
	ElementStatus uint32
	ResultStatus  uint32
	Error         *NativeObservableRowError
}
type NativeRinexObsCode struct {
	System uint32
	Code   string
}
type NativeRinexObsEpoch struct {
	Epoch          NativeCalendarEpoch
	Flag           uint8
	SatelliteCount int
}
type NativeRinexObsValue struct {
	SatelliteID, Code string
	Kind              uint32
	HasValue          bool
	Value             float64
	LLI, SSI          int32
}
type NativeRinexObsPseudorange struct {
	SatelliteID  string
	PseudorangeM float64
}
type NativeClockPhaseSample struct {
	HasPhaseS bool
	PhaseS    float64
}
type NativeRinexObsCarrierPhase struct {
	SatelliteID, Code string
	HasValueCycles    bool
	ValueCycles       float64
	LLI, SSI          int32
	HasFrequency      bool
	FrequencyHz       float64
	HasWavelength     bool
	WavelengthM       float64
	HasValueM         bool
	ValueM            float64
	PhaseShiftCycles  float64
}
type RinexObs struct{}

func ParseRinexObs([]byte) (*RinexObs, error) { return nil, protocolUnavailable() }
func (*RinexObs) Close() error                { return nil }
func (*RinexObs) Version() (float64, error)   { return 0, protocolUnavailable() }
func (*RinexObs) Header() (NativeRinexObsHeader, error) {
	return NativeRinexObsHeader{}, protocolUnavailable()
}
func (*RinexObs) EpochCount() (int, error)                  { return 0, protocolUnavailable() }
func (*RinexObs) Codes() ([]NativeRinexObsCode, error)      { return nil, protocolUnavailable() }
func (*RinexObs) Epochs() ([]NativeRinexObsEpoch, error)    { return nil, protocolUnavailable() }
func (*RinexObs) Values(int) ([]NativeRinexObsValue, error) { return nil, protocolUnavailable() }
func (*RinexObs) CarrierPhase(int) ([]NativeRinexObsCarrierPhase, error) {
	return nil, protocolUnavailable()
}
func (*RinexObs) Pseudoranges(int) ([]NativeRinexObsPseudorange, error) {
	return nil, protocolUnavailable()
}
func (*RinexObs) Observation(int, string, string) (float64, bool, int32, int32, error) {
	return 0, false, -1, -1, protocolUnavailable()
}
func (*RinexObs) ReceiverClockPhase() ([]NativeClockPhaseSample, error) {
	return nil, protocolUnavailable()
}
func (*RinexObs) Text() ([]byte, error) { return nil, protocolUnavailable() }
func RinexObservationFrequency(uint32, string, float64, bool, int8) (float64, error) {
	return 0, protocolUnavailable()
}
func RinexObservationWavelength(uint32, string, float64, bool, int8) (float64, error) {
	return 0, protocolUnavailable()
}

type NativeObservationQcOptions struct {
	HasIntervalOverride                             bool
	IntervalOverride, GapFactor, ClockJumpThreshold float64
}
type NativeObservationQcSummary struct {
	TotalEpochRecords, ObservationEpochs, EventRecords, PowerFailureEpochs, SkippedRecords          int
	HasInterval                                                                                     bool
	Interval                                                                                        float64
	IntervalSource                                                                                  uint32
	MissingEpochs, DataGapCount, SatelliteCount, SatelliteSignalCount, SystemSignalCount, NoteCount int
}
type NativeObservationQcDataGap struct {
	Start, End                     NativeCalendarEpoch
	NominalInterval, ObservedDelta float64
	MissingEpochs                  int
}
type NativeObservationQcClockJump struct {
	EpochIndex int
	Epoch      NativeCalendarEpoch
	DeltaS     float64
}
type NativeObservationQcCycleSlips struct {
	Observations, TotalSlips, SystemCount int
	HasObservationsPerSlip                bool
	ObservationsPerSlip                   float64
}
type ObservationQcReport struct{}

func QcOptionsInit() (NativeObservationQcOptions, error) {
	return NativeObservationQcOptions{}, protocolUnavailable()
}
func ObservationQCFromObs(*RinexObs, *NativeObservationQcOptions) (*ObservationQcReport, error) {
	return nil, protocolUnavailable()
}
func ParseObservationQC([]byte, *NativeObservationQcOptions) (*ObservationQcReport, error) {
	return nil, protocolUnavailable()
}
func (*ObservationQcReport) Close() error { return nil }
func (*ObservationQcReport) Summary() (NativeObservationQcSummary, error) {
	return NativeObservationQcSummary{}, protocolUnavailable()
}
func (*ObservationQcReport) Gaps() ([]NativeObservationQcDataGap, error) {
	return nil, protocolUnavailable()
}
func (*ObservationQcReport) ClockJumps() ([]NativeObservationQcClockJump, error) {
	return nil, protocolUnavailable()
}
func (*ObservationQcReport) CycleSlips() (NativeObservationQcCycleSlips, error) {
	return NativeObservationQcCycleSlips{}, protocolUnavailable()
}
func (*ObservationQcReport) Text() ([]byte, error) { return nil, protocolUnavailable() }
func (*ObservationQcReport) HTML() ([]byte, error) { return nil, protocolUnavailable() }
func (*ObservationQcReport) JSON() ([]byte, error) { return nil, protocolUnavailable() }

type NativeRINEXLintSummary struct {
	FindingCount, FatalCount, ErrorCount, WarningCount, InfoCount int
	IsClean, DecodedFromCRINEX                                    bool
}
type NativeRINEXLintFinding struct {
	Code                      string
	Severity                  uint32
	Repairable, HasEpochIndex bool
	EpochIndex                int
	HasSatellite              bool
	Satellite                 string
	HasField                  bool
	Field                     string
}
type RinexLintReport struct{}

func ParseRINEXLint([]byte, bool) (*RinexLintReport, error) { return nil, protocolUnavailable() }
func (*RinexLintReport) Close() error                       { return nil }
func (*RinexLintReport) Summary() (NativeRINEXLintSummary, error) {
	return NativeRINEXLintSummary{}, protocolUnavailable()
}
func (*RinexLintReport) Findings() ([]NativeRINEXLintFinding, error) {
	return nil, protocolUnavailable()
}

type NativeRINEXRepairOptions struct {
	HasFileStamp                                                                                        bool
	Program, RunBy, Date                                                                                string
	SetInterval, SetTimeOfLastObs, SetObservationCounts, DropEmptyRecords, SortRecords, DropUnsupported bool
}
type NativeRINEXRepairAction struct{ ID, Message string }
type RinexRepair struct{}

func NewRINEXRepairOptions() (NativeRINEXRepairOptions, error) {
	return NativeRINEXRepairOptions{}, protocolUnavailable()
}
func ParseRINEXRepair([]byte, bool, *NativeRINEXRepairOptions) (*RinexRepair, error) {
	return nil, protocolUnavailable()
}
func (*RinexRepair) Close() error { return nil }
func (*RinexRepair) Summary() (NativeRINEXLintSummary, error) {
	return NativeRINEXLintSummary{}, protocolUnavailable()
}
func (*RinexRepair) Actions() ([]NativeRINEXRepairAction, error) { return nil, protocolUnavailable() }
func (*RinexRepair) Text() ([]byte, error)                       { return nil, protocolUnavailable() }
func (*RinexRepair) CRINEXText() ([]byte, error)                 { return nil, protocolUnavailable() }
func DecodeCRINEX([]byte) ([]byte, error)                        { return nil, protocolUnavailable() }
func EncodeCRINEX([]byte) ([]byte, error)                        { return nil, protocolUnavailable() }
func RINEXBandFrequency(uint32, string, bool, int8) (float64, error) {
	return 0, protocolUnavailable()
}
func RINEXBandWavelength(uint32, string, bool, int8) (float64, error) {
	return 0, protocolUnavailable()
}

type NativeIQSample struct{ I, Q float64 }
type NativeAcquisitionOptions struct{ SampleRateHz, DopplerMinHz, DopplerMaxHz, DopplerStepHz float64 }
type NativeAcquisitionResult struct {
	CodePhaseChips, DopplerHz, PeakMetric, Metric, PeakPower float64
	GridCodePhaseBins                                        int
	GridDopplerStepHz, GridSamplesPerChip                    float64
	GridDopplerBinCount                                      int
}
type NativeSignalModulation struct {
	Kind        uint32
	Order, M, N float64
}
type NativeDLLTrackingOptions struct{ CN0DBHz, LoopBandwidthHz, IntegrationTimeS, CorrelatorSpacingChips, ReceiverBandwidthHz float64 }
type NativeDLLJitter struct{ Seconds, Chips, Meters, SquaringLoss float64 }
type NativeSignalInterference struct {
	Modulation          NativeSignalModulation
	PowerRatioToCarrier float64
}
type NativeCN0Degradation struct{ EffectiveCN0Hz, EffectiveCN0DBHz, DegradationDB float64 }
type NativeMultipathOptions struct{ MultipathToDirectRatio, CorrelatorSpacingChips, ReceiverBandwidthHz float64 }
type NativeMultipathPoint struct{ DelayChips, DelayS, InPhaseChips, InPhaseS, InPhaseM, AntiPhaseChips, AntiPhaseS, AntiPhaseM, RunningAverageChips, RunningAverageS, RunningAverageM float64 }
type NativeSpectralSeparation struct{ Hz, DBHz float64 }
type NativeCorrelateOptions struct{ SampleRateHz, DopplerHz, CodePhaseChips, CodeDopplerHz float64 }
type NativeCorrelationResult struct{ I, Q, Power float64 }
type NativeReplicaOptions struct {
	SampleRateHz                  float64
	NumSamples                    int
	CodePhaseChips, CodeDopplerHz float64
}

func SignalCAChip(int64, int64) (int8, error)                   { return 0, protocolUnavailable() }
func SignalCACode(int64) ([]int8, error)                        { return nil, protocolUnavailable() }
func SignalReplica(int64, NativeReplicaOptions) ([]int8, error) { return nil, protocolUnavailable() }
func SignalCorrelationAt([]int8, []int8, int64) (int32, error)  { return 0, protocolUnavailable() }
func SignalCrossCorrelation([]int8, []int8) ([]int32, error)    { return nil, protocolUnavailable() }
func SignalAutocorrelation([]int8) ([]int32, error)             { return nil, protocolUnavailable() }
func SignalCorrelate([]NativeIQSample, int64, NativeCorrelateOptions) (NativeCorrelationResult, error) {
	return NativeCorrelationResult{}, protocolUnavailable()
}
func SignalCorrelateAgainst([]NativeIQSample, []int8, float64, float64) (float64, float64, error) {
	return 0, 0, protocolUnavailable()
}
func SignalAcquire([]NativeIQSample, int64, NativeAcquisitionOptions) (NativeAcquisitionResult, []float64, error) {
	return NativeAcquisitionResult{}, nil, protocolUnavailable()
}
func SignalReferenceChipRate() (float64, error)                    { return 0, protocolUnavailable() }
func SignalBetzBandwidth() (float64, error)                        { return 0, protocolUnavailable() }
func SignalCoherentLoss(float64, float64) (float64, error)         { return 0, protocolUnavailable() }
func SignalCoherentLossDB(float64, float64) (float64, error)       { return 0, protocolUnavailable() }
func SignalSNRPost(float64, float64) (float64, error)              { return 0, protocolUnavailable() }
func SignalModulationLabel(NativeSignalModulation) (string, error) { return "", protocolUnavailable() }
func SignalModulationCodeRate(NativeSignalModulation) (float64, error) {
	return 0, protocolUnavailable()
}
func SignalPSD(NativeSignalModulation, float64) (float64, error) { return 0, protocolUnavailable() }
func SignalAnalysisPSD(NativeSignalModulation, float64) (float64, error) {
	return 0, protocolUnavailable()
}
func SignalPowerInBand(NativeSignalModulation, float64) (float64, error) {
	return 0, protocolUnavailable()
}
func SignalFractionPowerInBand(NativeSignalModulation, float64) (float64, error) {
	return 0, protocolUnavailable()
}
func SignalRMSBandwidth(NativeSignalModulation, float64) (float64, error) {
	return 0, protocolUnavailable()
}
func SignalAnalysisRMSBandwidth(NativeSignalModulation, float64) (float64, error) {
	return 0, protocolUnavailable()
}
func SignalDLL(NativeSignalModulation, NativeDLLTrackingOptions, uint32) (NativeDLLJitter, error) {
	return NativeDLLJitter{}, protocolUnavailable()
}
func SignalDLLAnalysis(NativeSignalModulation, NativeDLLTrackingOptions, uint32) (NativeDLLJitter, error) {
	return NativeDLLJitter{}, protocolUnavailable()
}
func SignalDLLLowerBound(NativeSignalModulation, NativeDLLTrackingOptions) (NativeDLLJitter, error) {
	return NativeDLLJitter{}, protocolUnavailable()
}
func SignalAnalysisDLLLowerBound(NativeSignalModulation, NativeDLLTrackingOptions) (NativeDLLJitter, error) {
	return NativeDLLJitter{}, protocolUnavailable()
}
func SignalSpectralSeparation(NativeSignalModulation, NativeSignalModulation, float64) (NativeSpectralSeparation, error) {
	return NativeSpectralSeparation{}, protocolUnavailable()
}
func SignalSpectralSeparationHz(NativeSignalModulation, NativeSignalModulation, float64) (float64, error) {
	return 0, protocolUnavailable()
}
func SignalSpectralSeparationDBHz(NativeSignalModulation, NativeSignalModulation, float64) (float64, error) {
	return 0, protocolUnavailable()
}
func SignalWhiteNoiseSeparation(NativeSignalModulation, float64) (float64, error) {
	return 0, protocolUnavailable()
}
func SignalCN0(NativeSignalModulation, float64, float64, []NativeSignalInterference) (NativeCN0Degradation, error) {
	return NativeCN0Degradation{}, protocolUnavailable()
}
func SignalMultipath(NativeSignalModulation, NativeMultipathOptions, []float64) ([]NativeMultipathPoint, error) {
	return nil, protocolUnavailable()
}

type NativeRTCMMessageInfo struct {
	Kind          uint32
	MessageNumber uint16
}
type NativeRTCMMSMHeader struct {
	ReferenceStationID                           uint16
	EpochTime                                    uint32
	MultipleMessage                              bool
	IODS, Reserved, ClockSteering, ExternalClock uint8
	DivergenceFreeSmoothing                      bool
	SmoothingInterval                            uint8
}
type NativeRTCMMSMInfo struct {
	MessageNumber               uint16
	System                      uint32
	Kind                        uint32
	Header                      NativeRTCMMSMHeader
	SatelliteCount, SignalCount int
}
type NativeRTCMMSMSatellite struct {
	ID, RoughRangeMS       uint8
	RoughRangeMod1         uint16
	HasExtendedInfo        bool
	ExtendedInfo           uint8
	HasRoughPhaseRangeRate bool
	RoughPhaseRangeRateMS  int16
}
type NativeRTCMMSMSignal struct {
	SatelliteID, SignalID           uint8
	FinePseudorange, FinePhaseRange int32
	LockTimeIndicator               uint16
	HalfCycleAmbiguity              bool
	CNR                             uint16
	HasFinePhaseRangeRate           bool
	FinePhaseRangeRate              int16
}
type NativeRTCMStationCoordinates struct {
	MessageNumber, ReferenceStationID                                           uint16
	ITRFRealizationYear                                                         uint8
	GPS, GLONASS, Galileo, ReferenceStation, SingleReceiverOscillator, Reserved bool
	QuarterCycleIndicator                                                       uint8
	ECEFX, ECEFY, ECEFZ                                                         int64
	XM, YM, ZM                                                                  float64
	HasAntennaHeight                                                            bool
	AntennaHeight                                                               uint16
	AntennaHeightM                                                              float64
}
type NativeRTCMAntennaDescriptor struct {
	MessageNumber, ReferenceStationID                                                            uint16
	AntennaSetupID                                                                               uint8
	HasAntennaSerialNumber, HasReceiverType, HasReceiverFirmwareVersion, HasReceiverSerialNumber bool
}
type NativeRTCMFrameSkip struct {
	Offset           int
	HasMessageNumber bool
	MessageNumber    uint16
	Reason           uint32
}
type NativeRTCMCellLLI struct {
	SatelliteID, SignalID, LLI uint8
	HasMinLockTime             bool
	MinLockTimeMS              uint32
}
type NativeRTCMNavICEphemeris struct {
	SatelliteID                        uint8
	WeekNumber                         uint16
	AF0, AF1                           int32
	AF2                                int16
	URA                                uint8
	TOC                                uint16
	TGD                                int16
	DeltaN                             int32
	IODEC                              uint8
	Reserved                           uint16
	L5Flag, SFlag                      bool
	CUC, CUS, CIC, CIS, CRC, CRS, IDOT int32
	M0                                 int64
	TOE                                uint16
	Eccentricity, SqrtA                uint64
	Omega0, Omega, I0                  int64
	OmegaDot                           int32
	SpareDF544, SpareDF545             uint8
}
type NativeRTCMSSRInfoV2 struct {
	MessageNumber                                                                    uint16
	System, Kind                                                                     uint32
	Header                                                                           NativeSsrHeader
	HasIGSSSRVersion                                                                 bool
	IGSSSRVersion                                                                    uint8
	OrbitCount, ClockCount, URACount, CodeBiasCount, PhaseBiasCount, PaddingBitCount int
}
type NativeRTCMSSRMessageV2 struct {
	Info             NativeRTCMSSRInfoV2
	Orbits           []NativeSsrOrbitRecord
	Clocks           []NativeSsrClockRecord
	URA              []NativeSsrUraRecord
	CodeBiases       []NativeSsrCodeBiasRecord
	CodeBiasSignals  []NativeSsrCodeBiasSignal
	PhaseBiases      []NativeSsrPhaseBiasRecord
	PhaseBiasSignals []NativeSsrPhaseBiasSignal
	PaddingBits      []bool
}
type NativeRTCMLegacyL1 struct {
	CodeIndicator                  bool
	Pseudorange                    uint32
	PhaseRangeMinusPseudorange     int32
	LockTimeIndicator              uint8
	HasPseudorangeModulusAmbiguity bool
	PseudorangeModulusAmbiguity    uint8
	HasCNR                         bool
	CNR                            uint8
}
type NativeRTCMLegacyL2 struct {
	CodeIndicator                uint8
	PseudorangeDifference        int16
	PhaseRangeMinusL1Pseudorange int32
	LockTimeIndicator            uint8
	HasCNR                       bool
	CNR                          uint8
}
type NativeRTCMLegacySatellite struct {
	SatelliteID         uint8
	HasFrequencyChannel bool
	FrequencyChannel    uint8
	L1                  NativeRTCMLegacyL1
	HasL2               bool
	L2                  NativeRTCMLegacyL2
}
type NativeRTCMLegacyObservations struct {
	MessageNumber, ReferenceStationID uint16
	EpochTime                         uint32
	SynchronousGNSS                   bool
	SatelliteCount                    uint8
	DivergenceFreeSmoothing           bool
	SmoothingInterval                 uint8
	Satellites                        []NativeRTCMLegacySatellite
	TrailingBits                      []bool
}
type NativeRTCMMessageAnnouncement struct {
	MessageNumber uint16
	Synchronous   bool
	Interval      uint16
}
type NativeRTCMSystemParameters struct {
	ReferenceStationID, MJD        uint16
	SecondsOfDay                   uint32
	AnnouncementCount, LeapSeconds uint8
	Announcements                  []NativeRTCMMessageAnnouncement
	TrailingBits                   []bool
}
type NativeRTCMText struct {
	ReferenceStationID, MJD uint16
	SecondsOfDay            uint32
	CharacterCount          uint8
	CodeUnits               []byte
	TrailingBits            []bool
}
type NativeRTCMNetworkAuxiliaryStation struct {
	NetworkID, SubnetworkID, AuxiliaryStationCount uint8
	MasterStationID, AuxiliaryStationID            uint16
	DeltaLatitude, DeltaLongitude, DeltaHeight     int32
	TrailingBits                                   []bool
}
type NativeRTCMSSRVTECInfo struct {
	MessageNumber    uint16
	HasIGSSSRVersion bool
	IGSSSRVersion    uint8
	EpochTimeS       uint32
	UpdateInterval   uint8
	MultipleMessage  bool
	IODSSR           uint8
	ProviderID       uint16
	SolutionID       uint8
	QualityIndicator uint16
	Layers           []NativeRTCMTecLayer
	TrailingBits     []bool
}
type NativeRTCMTecLayer struct {
	Height, Degree, Order uint8
	Cosine, Sine          []int16
}
type NativeRTCMNetworkDifference struct {
	SatelliteID, AmbiguityStatus, NonSyncCount uint8
	HasGeometric                               bool
	Geometric                                  int32
	HasIOD                                     bool
	IOD                                        uint8
	HasIonospheric                             bool
	Ionospheric                                int32
}
type NativeRTCMNetworkDifferences struct {
	MessageNumber                       uint16
	NetworkID, SubnetworkID             uint8
	EpochTime                           uint32
	MultipleMessage                     bool
	MasterStationID, AuxiliaryStationID uint16
	SatelliteCount                      uint8
	Satellites                          []NativeRTCMNetworkDifference
	TrailingBits                        []bool
}
type NativeRTCMNetworkResidual struct {
	SatelliteID, SOC uint8
	SOD              uint16
	SOH              uint8
	SLC, SLD         uint16
}
type NativeRTCMNetworkResiduals struct {
	MessageNumber                         uint16
	EpochTime                             uint32
	ReferenceStationID                    uint16
	ReferenceStationCount, SatelliteCount uint8
	Satellites                            []NativeRTCMNetworkResidual
	TrailingBits                          []bool
}
type NativeRTCMFKPGradient struct {
	SatelliteID, IOD                                                 uint8
	GeometricNorth, GeometricEast, IonosphericNorth, IonosphericEast int16
}
type NativeRTCMFKPGradients struct {
	MessageNumber, ReferenceStationID uint16
	EpochTime                         uint32
	SatelliteCount                    uint8
	Satellites                        []NativeRTCMFKPGradient
	TrailingBits                      []bool
}
type NativeRTCMGridResidual struct{ Horizontal1, Horizontal2, Height int16 }
type NativeRTCMResidualGrid struct {
	MessageNumber                                                                      uint16
	SystemID                                                                           uint8
	HorizontalShift, VerticalShift                                                     bool
	Origin1, Origin2                                                                   int32
	Extension1, Extension2                                                             uint16
	MeanOffset1, MeanOffset2, MeanHeightOffset                                         int16
	Residuals                                                                          [16]NativeRTCMGridResidual
	HorizontalInterpolation, VerticalInterpolation, HorizontalQuality, VerticalQuality uint8
	MJD                                                                                uint16
	TrailingBits                                                                       []bool
}
type NativeRTCMProjection struct {
	MessageNumber                                             uint16
	SystemID, ProjectionType                                  uint8
	Rectification                                             bool
	Latitude, Longitude, StandardParallel1, StandardParallel2 int64
	Azimuth                                                   uint64
	RectifiedToSkew                                           int32
	AddScale                                                  uint32
	Easting                                                   uint64
	Northing                                                  int64
	TrailingBits                                              []bool
}
type NativeRTCMPhysicalReferenceStation struct {
	NonPhysicalStationID, PhysicalStationID uint16
	ITRFRealizationYear                     uint8
	ECEFX, ECEFY, ECEFZ                     int64
	TrailingBits                            []bool
}
type NativeRTCMHelmertTransformation struct {
	MessageNumber                                         uint16
	SourceName, TargetName                                string
	SystemID                                              uint8
	UtilizedMessages                                      uint16
	PlateNumber, ComputationIndicator, HeightIndicator    uint8
	ValidityLatitude, ValidityLongitude                   int32
	ValidityExtensionLatitude, ValidityExtensionLongitude uint16
	DX, DY, DZ, R1, R2, R3, DS                            int32
	HasRotationPoint                                      bool
	RotationPointX, RotationPointY, RotationPointZ        int64
	AddAS, AddBS, AddAT, AddBT                            uint32
	HorizontalQuality, VerticalQuality                    uint8
	TrailingBits                                          []bool
}
type NativeRTCMGLONASSCodePhaseBiases struct {
	ReferenceStationID               uint16
	Aligned                          bool
	Reserved                         uint8
	HasL1CA, HasL1P, HasL2CA, HasL2P bool
	L1CA, L1P, L2CA, L2P             int16
	TrailingBits                     []bool
}
type NativeRTCMPreviousLock struct {
	HasMinLockTime bool
	MinLockTimeMS  uint32
	ElapsedMS      uint64
}
type PppFloatSolution struct{}
type PppFixedSolution struct{}
type PppUnplacedObservationV2 struct {
	EpochIndex               int
	SatelliteID, AmbiguityID string
	Reason                   uint32
	UnknownVariant           string
	HasSize                  bool
	OrbitM, ClockM           float64
}

func (*PppFloatSolution) UnplacedObservationsV2() ([]PppUnplacedObservationV2, error) {
	return nil, protocolUnavailable()
}
func (*PppFixedSolution) UnplacedObservationsV2() ([]PppUnplacedObservationV2, error) {
	return nil, protocolUnavailable()
}

type RtcmMessages struct{}
type RtcmFrames struct{}
type RtcmDiagnostics struct{}
type RtcmLockTimeTracker struct{}

func DecodeRTCM([]byte) (*RtcmMessages, error) { return nil, protocolUnavailable() }
func DecodeRTCMStream([]byte) (*RtcmMessages, *RtcmDiagnostics, error) {
	return nil, nil, protocolUnavailable()
}
func ScanRTCMFrames([]byte) (*RtcmFrames, error) { return nil, protocolUnavailable() }
func EncodeRTCMFrame([]byte) ([]byte, error)     { return nil, protocolUnavailable() }
func (*RtcmMessages) Close() error               { return nil }
func (*RtcmMessages) Count() (int, error)        { return 0, protocolUnavailable() }
func (*RtcmMessages) Kind(int) (NativeRTCMMessageInfo, error) {
	return NativeRTCMMessageInfo{}, protocolUnavailable()
}
func (*RtcmMessages) Encode(int) ([]byte, error) { return nil, protocolUnavailable() }
func (*RtcmMessages) Frame(int) ([]byte, error)  { return nil, protocolUnavailable() }
func (*RtcmMessages) MSMInfo(int) (NativeRTCMMSMInfo, error) {
	return NativeRTCMMSMInfo{}, protocolUnavailable()
}
func (*RtcmMessages) MSMSatellites(int) ([]NativeRTCMMSMSatellite, error) {
	return nil, protocolUnavailable()
}
func (*RtcmMessages) MSMSignals(int) ([]NativeRTCMMSMSignal, error) {
	return nil, protocolUnavailable()
}
func (*RtcmMessages) Station(int) (NativeRTCMStationCoordinates, error) {
	return NativeRTCMStationCoordinates{}, protocolUnavailable()
}
func (*RtcmMessages) Antenna(int) (NativeRTCMAntennaDescriptor, error) {
	return NativeRTCMAntennaDescriptor{}, protocolUnavailable()
}
func BuildRTCMNavICEphemeris(NativeRTCMNavICEphemeris) (*RtcmMessages, error) {
	return nil, protocolUnavailable()
}
func (*RtcmMessages) NavICEphemeris(int) (NativeRTCMNavICEphemeris, error) {
	return NativeRTCMNavICEphemeris{}, protocolUnavailable()
}
func (*RtcmMessages) TrailingBits(int) ([]bool, error)    { return nil, protocolUnavailable() }
func (*RtcmMessages) UnsupportedBody(int) ([]byte, error) { return nil, protocolUnavailable() }
func (*RtcmMessages) SSRInfoV2(int) (NativeRTCMSSRInfoV2, error) {
	return NativeRTCMSSRInfoV2{}, protocolUnavailable()
}
func BuildRTCMSSRMessageV2(NativeRTCMSSRMessageV2) (*RtcmMessages, error) {
	return nil, protocolUnavailable()
}
func (*RtcmMessages) WithTrailingBits(int, []bool) (*RtcmMessages, error) {
	return nil, protocolUnavailable()
}
func (*RtcmMessages) LegacyObservations(int) (NativeRTCMLegacyObservations, error) {
	return NativeRTCMLegacyObservations{}, protocolUnavailable()
}
func BuildRTCMLegacyObservations(NativeRTCMLegacyObservations) (*RtcmMessages, error) {
	return nil, protocolUnavailable()
}
func (*RtcmMessages) SystemParameters(int) (NativeRTCMSystemParameters, error) {
	return NativeRTCMSystemParameters{}, protocolUnavailable()
}
func (*RtcmMessages) Text(int) (NativeRTCMText, error) {
	return NativeRTCMText{}, protocolUnavailable()
}
func BuildRTCMSystemParameters(NativeRTCMSystemParameters) (*RtcmMessages, error) {
	return nil, protocolUnavailable()
}
func BuildRTCMText(NativeRTCMText) (*RtcmMessages, error) { return nil, protocolUnavailable() }
func (*RtcmMessages) NetworkAuxiliaryStation(int) (NativeRTCMNetworkAuxiliaryStation, error) {
	return NativeRTCMNetworkAuxiliaryStation{}, protocolUnavailable()
}
func BuildRTCMNetworkAuxiliaryStation(NativeRTCMNetworkAuxiliaryStation) (*RtcmMessages, error) {
	return nil, protocolUnavailable()
}
func (*RtcmMessages) SSRVTEC(int) (NativeRTCMSSRVTECInfo, error) {
	return NativeRTCMSSRVTECInfo{}, protocolUnavailable()
}
func BuildRTCMSSRVTEC(NativeRTCMSSRVTECInfo) (*RtcmMessages, error) {
	return nil, protocolUnavailable()
}
func (*RtcmMessages) NetworkDifferences(int) (NativeRTCMNetworkDifferences, error) {
	return NativeRTCMNetworkDifferences{}, protocolUnavailable()
}
func BuildRTCMNetworkDifferences(NativeRTCMNetworkDifferences) (*RtcmMessages, error) {
	return nil, protocolUnavailable()
}
func (*RtcmMessages) NetworkResiduals(int) (NativeRTCMNetworkResiduals, error) {
	return NativeRTCMNetworkResiduals{}, protocolUnavailable()
}
func BuildRTCMNetworkResiduals(NativeRTCMNetworkResiduals) (*RtcmMessages, error) {
	return nil, protocolUnavailable()
}
func (*RtcmMessages) FKPGradients(int) (NativeRTCMFKPGradients, error) {
	return NativeRTCMFKPGradients{}, protocolUnavailable()
}
func BuildRTCMFKPGradients(NativeRTCMFKPGradients) (*RtcmMessages, error) {
	return nil, protocolUnavailable()
}
func (*RtcmMessages) ResidualGrid(int) (NativeRTCMResidualGrid, error) {
	return NativeRTCMResidualGrid{}, protocolUnavailable()
}
func BuildRTCMResidualGrid(NativeRTCMResidualGrid) (*RtcmMessages, error) {
	return nil, protocolUnavailable()
}
func (*RtcmMessages) Projection(int) (NativeRTCMProjection, error) {
	return NativeRTCMProjection{}, protocolUnavailable()
}
func BuildRTCMProjection(NativeRTCMProjection) (*RtcmMessages, error) {
	return nil, protocolUnavailable()
}
func (*RtcmMessages) PhysicalReferenceStation(int) (NativeRTCMPhysicalReferenceStation, error) {
	return NativeRTCMPhysicalReferenceStation{}, protocolUnavailable()
}
func BuildRTCMPhysicalReferenceStation(NativeRTCMPhysicalReferenceStation) (*RtcmMessages, error) {
	return nil, protocolUnavailable()
}
func (*RtcmMessages) HelmertTransformation(int) (NativeRTCMHelmertTransformation, error) {
	return NativeRTCMHelmertTransformation{}, protocolUnavailable()
}
func BuildRTCMHelmertTransformation(NativeRTCMHelmertTransformation) (*RtcmMessages, error) {
	return nil, protocolUnavailable()
}
func (*RtcmMessages) GLONASSCodePhaseBiases(int) (NativeRTCMGLONASSCodePhaseBiases, error) {
	return NativeRTCMGLONASSCodePhaseBiases{}, protocolUnavailable()
}
func BuildRTCMGLONASSCodePhaseBiases(NativeRTCMGLONASSCodePhaseBiases) (*RtcmMessages, error) {
	return nil, protocolUnavailable()
}
func BuildRTCMUnsupported(uint16, []byte) (*RtcmMessages, error) { return nil, protocolUnavailable() }
func (*RtcmMessages) AntennaString(int, uint32) ([]byte, error)  { return nil, protocolUnavailable() }
func (*RtcmFrames) Close() error                                 { return nil }
func (*RtcmFrames) Count() (int, error)                          { return 0, protocolUnavailable() }
func (*RtcmFrames) Len(int) (int, error)                         { return 0, protocolUnavailable() }
func (*RtcmFrames) Body(int) ([]byte, error)                     { return nil, protocolUnavailable() }
func (*RtcmDiagnostics) Close() error                            { return nil }
func (*RtcmDiagnostics) ResyncBytes() (int, error)               { return 0, protocolUnavailable() }
func (*RtcmDiagnostics) SkippedCount() (int, error)              { return 0, protocolUnavailable() }
func (*RtcmDiagnostics) Skipped(int) (NativeRTCMFrameSkip, error) {
	return NativeRTCMFrameSkip{}, protocolUnavailable()
}
func (*RtcmDiagnostics) SkippedMessage(int) ([]byte, error) { return nil, protocolUnavailable() }
func NewRTCMTracker() (*RtcmLockTimeTracker, error)         { return nil, protocolUnavailable() }
func (*RtcmLockTimeTracker) Close() error                   { return nil }
func (*RtcmLockTimeTracker) Reset() error                   { return protocolUnavailable() }
func (*RtcmLockTimeTracker) Observe(*RtcmMessages, int) ([]NativeRTCMCellLLI, error) {
	return nil, protocolUnavailable()
}
func MinimumLockTime(uint32, uint16) (bool, uint32, error) { return false, 0, protocolUnavailable() }
func DeriveLLI(*NativeRTCMPreviousLock, bool, uint32, bool) (uint8, error) {
	return 0, protocolUnavailable()
}
func MSMEpochDelta(uint32, uint32, uint32) (uint64, error) { return 0, protocolUnavailable() }
func MSMSignalRINEXCode(uint32, uint8) ([]byte, error)     { return nil, protocolUnavailable() }
func RTCMLLIBits() (uint8, uint8, error)                   { return 0, 0, protocolUnavailable() }

type BroadcastEphemeris struct{}

func ParseBroadcast([]byte) (*BroadcastEphemeris, error) { return nil, protocolUnavailable() }

func (*BroadcastEphemeris) Close() error { return nil }
func (*BroadcastEphemeris) RecordCNavCorrection(int, uint32) (float64, bool, error) {
	return 0, false, protocolUnavailable()
}
func (*BroadcastEphemeris) RecordGroupDelay(int, uint32) (float64, bool, error) {
	return 0, false, protocolUnavailable()
}
func (*BroadcastEphemeris) RecordsInfo() ([]NativeBroadcastRecordInfo, error) {
	return nil, protocolUnavailable()
}
func (*BroadcastEphemeris) SelectByIssue(string, uint32, uint32, float64) (NativeBroadcastRecordInfo, bool, error) {
	return NativeBroadcastRecordInfo{}, false, protocolUnavailable()
}
func (*BroadcastEphemeris) SetNavMessagePreference(uint32) error { return protocolUnavailable() }
func (*BroadcastEphemeris) SourceStateAtEpochQueries(*ExactEpochQuery, *ExactEpochQuery, string) (EphemerisSourceState, error) {
	return EphemerisSourceState{}, protocolUnavailable()
}
func (*BroadcastEphemeris) TransmitEpochClockAtEpochQueries(*ExactEpochQuery, *ExactEpochQuery, string) (TransmitEpochClock, error) {
	return TransmitEpochClock{}, protocolUnavailable()
}
func (*BroadcastEphemeris) ClockRelativityAtEpochQuery(*ExactEpochQuery, string, [3]float64) (ClockRelativity, error) {
	return ClockRelativity{}, protocolUnavailable()
}
func (*BroadcastEphemeris) EphemerisVarianceAtEpochQueries(*ExactEpochQuery, *ExactEpochQuery, string) (float64, error) {
	return 0, protocolUnavailable()
}

type NativeObservablesOptions struct {
	CarrierHz         float64
	LightTime, Sagnac bool
}
type NativeEmissionMediaOptions struct {
	CarrierHz                                   float64
	MinElevationEnabled                         bool
	MinElevationRad                             float64
	TroposphereEnabled                          bool
	PressureHPA, TemperatureK, RelativeHumidity float64
}
type NativeEmissionMediaRow struct {
	Position                [3]float64
	HasPosition             bool
	ClockS                  float64
	HasClock                bool
	IonosphereSlantDelayM   float64
	HasIonosphereSlantDelay bool
	TroposphereDelayM       float64
	HasTroposphereDelay     bool
	Status, ResultStatus    uint32
	Error                   *NativeObservableRowError
}

func ObservablesOptionsInit() (NativeObservablesOptions, error) {
	return NativeObservablesOptions{}, protocolUnavailable()
}
func EmissionMediaOptionsInit() (NativeEmissionMediaOptions, error) {
	return NativeEmissionMediaOptions{}, protocolUnavailable()
}
func MissingObservablePosition() ([3]float64, error) { return [3]float64{}, protocolUnavailable() }
func (*BroadcastEphemeris) EmissionBatch([]string, []float64, [3]float64, *NativeEmissionMediaOptions) ([]NativeEmissionMediaRow, error) {
	return nil, protocolUnavailable()
}
func (*SP3) EmissionBatch([]string, []float64, [3]float64, *NativeEmissionMediaOptions) ([]NativeEmissionMediaRow, error) {
	return nil, protocolUnavailable()
}

type NativeFrequencyChannel struct {
	Slot    uint8
	Channel int8
}
type NativeGlonassRecord struct {
	SatelliteID             string
	ToeUTCJ2000S            float64
	EpochUTCJ2000S          float64
	PositionM               [3]float64
	VelocityMPerS           [3]float64
	AccelerationMPerS2      [3]float64
	ClockBiasS              float64
	GammaN                  float64
	SVHealth                float64
	FrequencyChannel        int32
	StatedFrequencyChannel  int32
	HasMessageFrameTime     bool
	MessageFrameTimeS       float64
	HasAgeDays              bool
	AgeDays                 float64
	HasStatusFlags          bool
	StatusFlags             float64
	HasL1L2GroupDelayFieldS bool
	L1L2GroupDelayFieldS    float64
	HasURAI                 bool
	URAI                    float64
	HasHealthFlags          bool
	HealthFlags             float64
}
type NativeSkippedGlonassRecord struct{ SatelliteID string }
type RinexGlonassRecords struct{}

func ParseRinexGlonassRecords([]byte) (*RinexGlonassRecords, error) {
	return nil, protocolUnavailable()
}
func (*RinexGlonassRecords) Close() error        { return nil }
func (*RinexGlonassRecords) Count() (int, error) { return 0, protocolUnavailable() }
func (*RinexGlonassRecords) Record(int) (NativeGlonassRecord, error) {
	return NativeGlonassRecord{}, protocolUnavailable()
}
func (*RinexGlonassRecords) SkippedCount() (int, error) { return 0, protocolUnavailable() }
func (*RinexGlonassRecords) Skipped(int) (NativeSkippedGlonassRecord, error) {
	return NativeSkippedGlonassRecord{}, protocolUnavailable()
}
func (*BroadcastEphemeris) GlonassFrequencyChannels() ([]NativeFrequencyChannel, error) {
	return nil, protocolUnavailable()
}
func (*BroadcastEphemeris) GlonassRecords() ([]NativeGlonassRecord, error) {
	return nil, protocolUnavailable()
}

type NativeSBASFastCorrection struct {
	PRCM, RRCMPerS float64
	UDREI          uint8
	TOfJ2000S      float64
	IODF           uint8
}
type NativeSBASLongTermCorrection struct {
	IODE                               uint8
	DeltaECEFM, DeltaECEFRateMPerS     [3]float64
	DeltaAF0S, DeltaAF1SPerS, T0J2000S float64
}
type NativeSBASGeoState struct {
	Position, Velocity, Acceleration     [3]float64
	ClockOffsetS, ClockDriftSS, T0J2000S float64
}
type NativeSBASIGP struct {
	LatDeg, LonDeg, VerticalDelayM float64
	HasGIVEVariance                bool
	GIVEVarianceM2                 float64
}
type SBASCorrectionStore struct{}
type NativeSSRClockCorrection struct {
	Source                                                       uint32
	ProviderID                                                   uint16
	SolutionID, IODSSR                                           uint8
	C0M, C1MPerS, C2MPerS2, RefEpochJ2000S, UpdateIntervalS      float64
	HasHighRate                                                  bool
	HighRateC0M, HighRateRefEpochJ2000S, HighRateUpdateIntervalS float64
}
type NativeSSROrbitCorrection struct {
	Source                                                                                                    uint32
	ProviderID                                                                                                uint16
	SolutionID                                                                                                uint8
	IODE                                                                                                      uint32
	IODSSR                                                                                                    uint8
	CRSRegional                                                                                               bool
	ReferencePoint                                                                                            uint32
	RadialM, AlongM, CrossM, RadialRateMPerS, AlongRateMPerS, CrossRateMPerS, RefEpochJ2000S, UpdateIntervalS float64
}
type SSRCorrectionStore struct{}

func NewSBASStore() (*SBASCorrectionStore, error) { return nil, protocolUnavailable() }
func (*SBASCorrectionStore) Close() error         { return nil }
func (*SBASCorrectionStore) Ingest(*SbasBlock, string, NativeGnssWeekTow) error {
	return protocolUnavailable()
}
func (*SBASCorrectionStore) PreferredGeo(float64) (string, bool, error) {
	return "", false, protocolUnavailable()
}
func (*SBASCorrectionStore) ReadyGeos(float64) ([]string, error) { return nil, protocolUnavailable() }
func (*SBASCorrectionStore) Fast(string, string) (NativeSBASFastCorrection, bool, error) {
	return NativeSBASFastCorrection{}, false, protocolUnavailable()
}
func (*SBASCorrectionStore) LongTerm(string, string) (NativeSBASLongTermCorrection, bool, error) {
	return NativeSBASLongTermCorrection{}, false, protocolUnavailable()
}
func (*SBASCorrectionStore) GeoNav(string) (NativeSBASGeoState, bool, error) {
	return NativeSBASGeoState{}, false, protocolUnavailable()
}
func (*SBASCorrectionStore) IGPs(string) ([]NativeSBASIGP, error) { return nil, protocolUnavailable() }
func (*SBASCorrectionStore) SlantDelay(string, Geodetic, float64, float64, float64) (float64, bool, error) {
	return 0, false, protocolUnavailable()
}
func (*SBASCorrectionStore) CorrectedState(*BroadcastEphemeris, string, uint32, string, float64) ([3]float64, float64, bool, error) {
	return [3]float64{}, 0, false, protocolUnavailable()
}
func (*SBASCorrectionStore) SolveBroadcastAtExactEpoch(*BroadcastEphemeris, string, uint32, SPPConfig, *ExactEpoch) (SPPSolution, error) {
	return SPPSolution{}, protocolUnavailable()
}
func (*SBASCorrectionStore) SolveBroadcastV2AtExactEpoch(*BroadcastEphemeris, string, uint32, SppInputsV2, *ExactEpoch) (SPPSolution, error) {
	return SPPSolution{}, protocolUnavailable()
}
func (*SBASCorrectionStore) SourceStateAtEpochQueries(*BroadcastEphemeris, string, uint32, string, *ExactEpochQuery, *ExactEpochQuery) (EphemerisSourceState, error) {
	return EphemerisSourceState{}, protocolUnavailable()
}
func (*SBASCorrectionStore) TransmitEpochClockAtEpochQueries(*BroadcastEphemeris, string, uint32, string, *ExactEpochQuery, *ExactEpochQuery) (TransmitEpochClock, error) {
	return TransmitEpochClock{}, protocolUnavailable()
}
func (*SBASCorrectionStore) ClockRelativityAtEpochQuery(*BroadcastEphemeris, string, uint32, string, *ExactEpochQuery, [3]float64) (ClockRelativity, error) {
	return ClockRelativity{}, protocolUnavailable()
}
func (*SBASCorrectionStore) EphemerisVarianceAtEpochQueries(*BroadcastEphemeris, string, uint32, string, *ExactEpochQuery, *ExactEpochQuery) (float64, error) {
	return 0, protocolUnavailable()
}
func NewSSRStore(uint32) (*SSRCorrectionStore, error) { return nil, protocolUnavailable() }
func NewSSRStoreFromRTCM([]byte, NativeGnssWeekTow) (*SSRCorrectionStore, error) {
	return nil, protocolUnavailable()
}
func (*SSRCorrectionStore) Close() error { return nil }
func (*SSRCorrectionStore) Ingest(*RtcmMessages, NativeGnssWeekTow) error {
	return protocolUnavailable()
}
func (*SSRCorrectionStore) Orbit(string) (NativeSSROrbitCorrection, bool, error) {
	return NativeSSROrbitCorrection{}, false, protocolUnavailable()
}
func (*SSRCorrectionStore) Clock(string) (NativeSSRClockCorrection, bool, error) {
	return NativeSSRClockCorrection{}, false, protocolUnavailable()
}
func (*SSRCorrectionStore) CodeBias(string, uint32, uint8) (float64, bool, error) {
	return 0, false, protocolUnavailable()
}
func (*SSRCorrectionStore) PhaseBias(string, uint32, uint8) (float64, bool, error) {
	return 0, false, protocolUnavailable()
}
func (*SSRCorrectionStore) URA(string) (uint8, bool, error) { return 0, false, protocolUnavailable() }
func (*SSRCorrectionStore) CorrectedState(*BroadcastEphemeris, string, float64, float64, uint32, bool, uint16) ([3]float64, float64, bool, error) {
	return [3]float64{}, 0, false, protocolUnavailable()
}
func (*SSRCorrectionStore) SolveBroadcastAtExactEpoch(*BroadcastEphemeris, SPPConfig, float64, uint32, bool, uint16, uint32, *ExactEpoch) (SPPSolution, error) {
	return SPPSolution{}, protocolUnavailable()
}
func (*SSRCorrectionStore) SolveBroadcastV2AtExactEpoch(*BroadcastEphemeris, SppInputsV2, float64, uint32, bool, uint16, uint32, *ExactEpoch) (SPPSolution, error) {
	return SPPSolution{}, protocolUnavailable()
}
func (*SSRCorrectionStore) SourceStateAtEpochQueries(*BroadcastEphemeris, string, *ExactEpochQuery, *ExactEpochQuery, float64, uint32, bool, uint16, uint32) (SSRCorrectedState, error) {
	return SSRCorrectedState{}, protocolUnavailable()
}
func (*SSRCorrectionStore) TransmitEpochClockAtEpochQueries(*BroadcastEphemeris, string, *ExactEpochQuery, *ExactEpochQuery, float64, uint32, bool, uint16, uint32) (TransmitEpochClock, SSRCorrectedState, error) {
	return TransmitEpochClock{}, SSRCorrectedState{}, protocolUnavailable()
}
func (*SSRCorrectionStore) ClockRelativityAtEpochQuery(*BroadcastEphemeris, string, *ExactEpochQuery, [3]float64, float64, uint32, bool, uint16, uint32) (ClockRelativity, SSRCorrectedState, error) {
	return ClockRelativity{}, SSRCorrectedState{}, protocolUnavailable()
}
func (*SSRCorrectionStore) EphemerisVarianceAtEpochQueries(*BroadcastEphemeris, string, *ExactEpochQuery, *ExactEpochQuery, float64, uint32, bool, uint16, uint32) (float64, SSRCorrectedState, error) {
	return 0, SSRCorrectedState{}, protocolUnavailable()
}
