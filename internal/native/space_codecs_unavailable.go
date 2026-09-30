//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

type BiasEpoch struct {
	Year        int32
	DayOfYear   uint16
	SecondOfDay uint32
}
type BiasRecord struct {
	Kind, TargetKind, System        uint32
	HasSatelliteID                  bool
	SatelliteID, Station, SVN, Obs1 string
	HasObs2                         bool
	Obs2                            string
	HasValidFrom                    bool
	ValidFrom                       BiasEpoch
	HasValidUntil                   bool
	ValidUntil                      BiasEpoch
	Value                           float64
	HasSigma                        bool
	Sigma                           float64
	HasSlope                        bool
	Slope                           float64
	HasSlopeSigma                   bool
	SlopeSigma                      float64
	IsPhase                         bool
}
type CodeDCBOptions struct {
	Obs1, Obs2        string
	Year              int32
	Month             uint8
	TimeScale         uint32
	HasReceiverSystem bool
	ReceiverSystem    uint32
}
type BiasSet struct{}

type BiasNotice struct {
	Kind           uint32
	Departure      uint32
	HasLine        bool
	Line           int
	First          int
	Second         int
	DeclaredCount  uint64
	SolutionRows   int
	BiasMode       uint32
	UnknownVariant string
}

func ParseBiasSINEX([]byte, bool) (*BiasSet, error)                { return nil, unavailable() }
func ParseBiasSINEXWithPolicy([]byte, uint32) (*BiasSet, error)    { return nil, unavailable() }
func ParseCodeDCB([]byte, *CodeDCBOptions, bool) (*BiasSet, error) { return nil, unavailable() }
func ParseCodeDCBWithPolicy([]byte, *CodeDCBOptions, uint32) (*BiasSet, error) {
	return nil, unavailable()
}
func (*BiasSet) Close() error                           { return nil }
func (*BiasSet) RecordCount() (int, error)              { return 0, unavailable() }
func (*BiasSet) SkippedRecordCount() (int, error)       { return 0, unavailable() }
func (*BiasSet) WarningCount() (int, error)             { return 0, unavailable() }
func (*BiasSet) NoticeCount() (int, error)              { return 0, unavailable() }
func (*BiasSet) Notice(int) (BiasNotice, error)         { return BiasNotice{}, unavailable() }
func (*BiasSet) NoticeText(int, uint32) (string, error) { return "", unavailable() }
func (*BiasSet) BiasSINEXText() (string, error)         { return "", unavailable() }
func (*BiasSet) BiasSINEXBytes() ([]byte, error)        { return nil, unavailable() }
func (*BiasSet) CodeDCBText() (string, error)           { return "", unavailable() }
func (*BiasSet) CodeDCBBytes() ([]byte, error)          { return nil, unavailable() }
func (*BiasSet) Record(int) (BiasRecord, error)         { return BiasRecord{}, unavailable() }
func (*BiasSet) CodeOSBSeconds(string, string, BiasEpoch) (float64, bool, error) {
	return 0, false, unavailable()
}
func (*BiasSet) CodeOSBLookup(string, string, BiasEpoch) (NativeBiasLookup, error) {
	return NativeBiasLookup{}, unavailable()
}
func (*BiasSet) PhaseOSBCycles(string, string, BiasEpoch) (float64, bool, error) {
	return 0, false, unavailable()
}
func (*BiasSet) PhaseOSBLookup(string, string, BiasEpoch, bool, float64) (NativeBiasLookup, error) {
	return NativeBiasLookup{}, unavailable()
}
func (*BiasSet) CodeDSBSeconds(string, string, string, BiasEpoch) (float64, bool, error) {
	return 0, false, unavailable()
}
func (*BiasSet) CodeDSBLookup(string, string, string, BiasEpoch) (NativeBiasLookup, error) {
	return NativeBiasLookup{}, unavailable()
}
func (*BiasSet) Mode() (uint32, uint32, error)         { return 0, 0, unavailable() }
func (*BiasSet) ModeInfo() (NativeBiasModeInfo, error) { return NativeBiasModeInfo{}, unavailable() }

type AllanSample struct {
	Present bool
	Value   float64
}
type AllanSeries struct {
	Kind    uint32
	Samples []AllanSample
}
type AllanEstimatorSet struct{ ADEV, OverlappingADEV, MDEV, HDEV, TDEV bool }
type AllanOptions struct {
	Estimators       AllanEstimatorSet
	TauGrid          uint32
	GapPolicy        uint32
	AveragingFactors []int
}
type AllanPoint struct {
	TauS, Deviation float64
	N               int
}
type AllanInputNative struct {
	Series  AllanSeries
	Tau0S   float64
	Options AllanOptions
}
type AllanDeviationCurves struct{}
type PowerLawNoiseOptions struct {
	MinPointsPerOctave                                                  int
	SlopeTolerance, ScatterTolerance, BasicTauS, MeasurementBandwidthHz float64
}
type PowerLawOctave struct {
	TauStartS, TauEndS             float64
	PointCount                     int
	HasADEVSlope                   bool
	ADEVSlope                      float64
	HasMDEVSlope                   bool
	MDEVSlope                      float64
	HasSlopeScatter                bool
	SlopeScatter                   float64
	DominanceKind, NoiseType, Flag uint32
}
type PowerLawNoiseRegion struct {
	NoiseType               uint32
	TauStartS, TauEndS      float64
	OctaveCount, PointCount int
	MeanSlope, Coefficient  float64
}
type PowerLawNoiseFit struct{}

func AllanOptionsDefault() (AllanOptions, error) { return AllanOptions{}, unavailable() }
func ComputeAllanDeviations(AllanInputNative) (*AllanDeviationCurves, error) {
	return nil, unavailable()
}
func (*AllanDeviationCurves) Close() error                 { return nil }
func (*AllanDeviationCurves) Present(uint32) (bool, error) { return false, unavailable() }
func (*AllanDeviationCurves) Curve(uint32) ([]AllanPoint, bool, error) {
	return nil, false, unavailable()
}
func AllanDeviation(AllanSeries, float64, []int) ([]AllanPoint, error)    { return nil, unavailable() }
func OverlappingADEV(AllanSeries, float64, []int) ([]AllanPoint, error)   { return nil, unavailable() }
func ModifiedADEV(AllanSeries, float64, []int) ([]AllanPoint, error)      { return nil, unavailable() }
func HadamardDeviation(AllanSeries, float64, []int) ([]AllanPoint, error) { return nil, unavailable() }
func TimeDeviation(AllanSeries, float64, []int) ([]AllanPoint, error)     { return nil, unavailable() }
func PowerLawNoiseOptionsDefault(float64, float64) (PowerLawNoiseOptions, error) {
	return PowerLawNoiseOptions{}, unavailable()
}
func PowerLawNoiseSlopes(uint32) (float64, float64, int, error) { return 0, 0, 0, unavailable() }
func FitPowerLawNoise([]AllanPoint, []AllanPoint, *PowerLawNoiseOptions) (*PowerLawNoiseFit, error) {
	return nil, unavailable()
}
func (*PowerLawNoiseFit) Close() error                            { return nil }
func (*PowerLawNoiseFit) Coefficients() ([5]float64, error)       { return [5]float64{}, unavailable() }
func (*PowerLawNoiseFit) Octaves() ([]PowerLawOctave, error)      { return nil, unavailable() }
func (*PowerLawNoiseFit) Regions() ([]PowerLawNoiseRegion, error) { return nil, unavailable() }

type OEM struct{}
type OMM struct{}
type OPM struct{}
type SPK struct{}
type TDM struct{}

func ParseOEM([]byte, bool) (*OEM, error)            { return nil, unavailable() }
func ParseOPM([]byte, bool) (*OPM, error)            { return nil, unavailable() }
func (*OEM) Close() error                            { return nil }
func (*OMM) Close() error                            { return nil }
func (*OPM) Close() error                            { return nil }
func (*OEM) SegmentCount() (int, error)              { return 0, unavailable() }
func (*OEM) SkippedStateCount() (int, error)         { return 0, unavailable() }
func (*OEM) SkippedStatePayload(int) ([]byte, error) { return nil, unavailable() }
func (*OEM) Text(bool) ([]byte, error)               { return nil, unavailable() }
func (*OMM) Text(uint32) ([]byte, error)             { return nil, unavailable() }
func (*OMM) SnapshotJSON() ([]byte, error)           { return nil, unavailable() }
func OMMFromSnapshotJSON([]byte) (*OMM, error)       { return nil, unavailable() }
func (*OMM) ToElementSet() (OMMElementSet, error)    { return OMMElementSet{}, unavailable() }

// OMMArray is the unavailable-build placeholder for the owned OMM collection.
type OMMArray struct{}

// ParseOMMJSONArray is unavailable without the supported native build.
func ParseOMMJSONArray([]byte) (*OMMArray, error) { return nil, unavailable() }

// ParseOMMXMLAll is unavailable without the supported native build.
func ParseOMMXMLAll([]byte) (*OMMArray, error) { return nil, unavailable() }

// ParseOMMCSVArray is unavailable without the supported native build.
func ParseOMMCSVArray([]byte) (*OMMArray, error) { return nil, unavailable() }

// NewOMMArray is unavailable without the supported native build.
func NewOMMArray() (*OMMArray, error) { return nil, unavailable() }

// Append is unavailable without the supported native build.
func (*OMMArray) Append(*OMM) error { return unavailable() }

// Count is unavailable without the supported native build.
func (*OMMArray) Count() (int, error) { return 0, unavailable() }

// Record is unavailable without the supported native build.
func (*OMMArray) Record(int) (*OMM, error) { return nil, unavailable() }

// Skipped is unavailable without the supported native build.
func (*OMMArray) Skipped() ([]OMMSkippedRecord, error) { return nil, unavailable() }

// JSON is unavailable without the supported native build.
func (*OMMArray) JSON() ([]byte, error) { return nil, unavailable() }

// JSONDiscardingComments is unavailable without the supported native build.
func (*OMMArray) JSONDiscardingComments() ([]byte, error) { return nil, unavailable() }

// CSV is unavailable without the supported native build.
func (*OMMArray) CSV() ([]byte, error) { return nil, unavailable() }

// CSVDiscardingComments is unavailable without the supported native build.
func (*OMMArray) CSVDiscardingComments() ([]byte, error) { return nil, unavailable() }

// Close releases the unavailable placeholder.
func (*OMMArray) Close() error { return nil }

// ParseOMM is unavailable without the supported native build.
func ParseOMM([]byte) (*OMM, error) { return nil, unavailable() }

// ParseOMMCSV is unavailable without the supported native build.
func ParseOMMCSV([]byte) (*OMM, error) { return nil, unavailable() }

// ParseOMMEpoch is unavailable without the supported native build.
func ParseOMMEpoch(string) (OMMEpoch, error) { return OMMEpoch{}, unavailable() }

// Sgp4Satellite is the unavailable-build placeholder for an OMM SGP4 handle.
type Sgp4Satellite struct{}

// Sgp4SatelliteFromOMM is unavailable without the supported native build.
func Sgp4SatelliteFromOMM(*OMM) (*Sgp4Satellite, error) { return nil, unavailable() }

// PropagateMinutes is unavailable without the supported native build.
func (*Sgp4Satellite) PropagateMinutes(float64) (TEMEState, error) { return TEMEState{}, unavailable() }

// PropagateJD is unavailable without the supported native build.
func (*Sgp4Satellite) PropagateJD(JulianDate) (TEMEState, error) { return TEMEState{}, unavailable() }

// EpochJ2000S is unavailable without the supported native build.
func (*Sgp4Satellite) EpochJ2000S() (float64, error) { return 0, unavailable() }
func (*Sgp4Satellite) Epoch() (JulianDate, error)    { return JulianDate{}, unavailable() }

// Close releases the unavailable placeholder.
func (*Sgp4Satellite) Close() error    { return nil }
func (*OPM) Text(bool) ([]byte, error) { return nil, unavailable() }

type ConstellationRecord struct {
	System             uint32
	PRN                uint16
	SVNPresent         bool
	SVN                uint16
	NORADID            uint32
	FDMAChannelPresent bool
	FDMAChannel        int8
	Active             bool
	Usable             bool
}
type SkippedOMM struct {
	NORADIDPresent    bool
	NORADID           uint32
	ObjectNamePresent bool
	ObjectName        string
}
type OMMCatalog struct{}

func BuildOMMCatalogLenient(uint32, []byte) (*OMMCatalog, error) { return nil, unavailable() }
func (*OMMCatalog) Close() error                                 { return nil }
func (*OMMCatalog) RecordCount() (int, error)                    { return 0, unavailable() }
func (*OMMCatalog) SkippedCount() (int, error)                   { return 0, unavailable() }
func (*OMMCatalog) MalformedCount() (int, error)                 { return 0, unavailable() }
func (*OMMCatalog) MalformedRecord(int) (uint64, []byte, error)  { return 0, nil, unavailable() }
func (*OMMCatalog) Record(int) (ConstellationRecord, error) {
	return ConstellationRecord{}, unavailable()
}
func (*OMMCatalog) Skipped(int) (SkippedOMM, error) { return SkippedOMM{}, unavailable() }

type SPKState struct {
	Target, Center    int32
	PositionKm        [3]float64
	HasVelocity       bool
	HasVelocityKmPerS bool
	VelocityKmPerS    [3]float64
	Frame             int32
}

func LoadSPK([]byte) (*SPK, error)                         { return nil, unavailable() }
func (*SPK) Close() error                                  { return nil }
func (*SPK) State(int32, int32, float64) (SPKState, error) { return SPKState{}, unavailable() }
func (*SPK) StateInFrame(int32, int32, float64, int32) (SPKState, error) {
	return SPKState{}, unavailable()
}

type TDMParticipant struct {
	SegmentIndex int
	Index        uint8
	Name         string
}
type TDMPath struct {
	SegmentIndex int
	Key          string
	HasIndex     bool
	Index        uint8
	Participants []uint8
}
type TDMDataRecord struct {
	SegmentIndex              int
	Observable                uint32
	HasObservableParticipant  bool
	ObservableParticipant     uint8
	Unit                      uint32
	Keyword, Epoch, ValueText string
	Value                     float64
}
type TDMStringField struct {
	Present bool
	Value   string
}
type TDMSegmentSummary struct {
	SegmentIndex                             int
	Mode, TimetagRef, TimeSystem             TDMStringField
	RangeUnit                                uint32
	ParticipantCount, PathCount, RecordCount int
}

func ParseTDM([]byte) (*TDM, error)                  { return nil, unavailable() }
func (*TDM) Close() error                            { return nil }
func (*TDM) Text() ([]byte, error)                   { return nil, unavailable() }
func (*TDM) SegmentCount() (int, error)              { return 0, unavailable() }
func (*TDM) RecordCount() (int, error)               { return 0, unavailable() }
func (*TDM) Segments() ([]TDMSegmentSummary, error)  { return nil, unavailable() }
func (*TDM) Participants() ([]TDMParticipant, error) { return nil, unavailable() }
func (*TDM) Paths() ([]TDMPath, error)               { return nil, unavailable() }
func (*TDM) Records() ([]TDMDataRecord, error)       { return nil, unavailable() }
