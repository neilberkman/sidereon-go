//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

type Ionex struct{}

// TecSamplesOutcome is the detached fixed-width and text result of a sample build.
type TecSamplesOutcome struct {
	// IsOK reports whether construction produced a product.
	IsOK bool
	// HasIndex reports whether the error names an input entry.
	HasIndex bool
	// HasAxisValue reports whether AxisValue carries a refused coordinate or step.
	HasAxisValue bool
	// Status is the native status paired with the outcome.
	Status uint32
	// Kind identifies the typed construction failure.
	Kind uint32
	// Input identifies the caller buffer named by Index.
	Input uint32
	// Index is the zero-based input index when HasIndex is true.
	Index uint64
	// NodeCount is the short-axis length for a too-few-nodes failure.
	NodeCount uint64
	// ValueCount is the supplied value count for a count mismatch.
	ValueCount uint64
	// ExpectedValueCount is the axis-implied count for a count mismatch.
	ExpectedValueCount uint64
	// AxisValue is the refused coordinate or step when HasAxisValue is true.
	AxisValue float64
	// Message is the detached native detail string.
	Message string
}
type IonexWarning struct {
	Kind                                          uint32
	Line, MapNumber, SetByLine                    uint64
	DeclaredCount                                 uint64
	TECMapCount, AllMapCount                      uint64
	DeclaredIntervalS                             uint32
	ActualSpacingS                                int64
	Exponent                                      int32
	LatDeg, LonDeg                                float64
	HasEpochs                                     bool
	DeclaredEpochJ2000S, MapEpochJ2000S           float64
	DeclaredEpochJ2000WholeS, MapEpochJ2000WholeS int64
	Label, Message                                string
}
type IonexEpochError struct {
	Kind, Scale  uint32
	HasUTCJ2000S bool
	UTCJ2000S    int64
}
type IonexSlantError struct {
	Kind, CoverageError                                uint32
	HasGap                                             bool
	Gap                                                IonexNodeGap
	Refusal                                            uint32
	RefusalMapNumber, RefusalLatIndex, RefusalLonIndex uint64
	HasMappingDeclaration                              bool
	MappingDeclaration                                 uint32
	HasMappingFunction                                 bool
	MappingFunction                                    uint32
}
type IonexNodeGap struct {
	HasGap         bool
	Earlier, Later [4]bool
}
type IonexInstantSlantRequest struct {
	LatDeg, LonDeg, AzimuthDeg, ElevationDeg float64
	Epoch                                    NativeClockEpoch
	FrequencyHz                              float64
}
type IonexInstantSlantRow struct {
	IsOK                 bool
	Status               uint32
	Evaluation           IonexSlantDelayEvaluation
	Error                IonexSlantError
	EpochError           IonexEpochError
	Message, MappingCode string
}
type IonexInstantSlantResultList struct{ Rows []IonexInstantSlantRow }
type IonexSlantPolicy struct{ Coverage, MissingNodes, Mapping uint32 }
type IonexSlantDelayEvaluation struct {
	DelayM                                           float64
	Status, CoverageError                            uint32
	IsValid, HasHeld, HasDegraded, HasAssumedMapping bool
	Gap                                              IonexNodeGap
	AssumedMapping                                   uint32
}
type TecSample struct {
	TimeScale                                  uint32
	EpochJ2000S, LatDeg, LonDeg, VTECTECU      float64
	VTECPresent, VTECPresenceKnown, RMSPresent bool
	RMSTECU                                    float64
	EpochJ2000WholeS                           int64
	EpochJ2000WholeSPresent                    bool
	HeightOffsetKm                             float64
	HeightOffsetPresent                        bool
}
type TecGridSamples struct {
	TimeScale                                     uint32
	MapEpochsJ2000S, LatNodesDeg, LonNodesDeg     []float64
	DLatDeg, DLonDeg, ShellHeightKm, BaseRadiusKm float64
	Exponent                                      int32
	TECMAPsTECU                                   []float64
	TECMAPsPresent                                []bool
	RMSPresent                                    bool
	RMSMAPsTECU                                   []float64
	RMSMAPsPresent                                []bool
	HeightPresent                                 bool
	HeightMapsKm                                  []float64
	HeightMapsPresent                             []bool
	Header                                        *IonexHeaderMetadata
}
type IonexHeaderMetadata struct {
	Version                                                float64
	Date, Program, RunBy, SatelliteSystem, ObservablesUsed string
	HasSatelliteCount, HasStationCount, HasMapsInFile      bool
	SatelliteCount, StationCount, MapsInFile               uint32
	ElevationCutoffDeg                                     float64
	IntervalS                                              uint32
	MappingDeclaration, MappingFunction                    uint32
	MappingFunctionCode                                    string
	Descriptions, Comments                                 []string
}
type TecGridSamplesInfo struct {
	MapEpochCount, LatNodeCount, LonNodeCount     int
	DLatDeg, DLonDeg, ShellHeightKm, BaseRadiusKm float64
	Exponent                                      int32
	RMSPresent                                    bool
	TECMAPValueCount, RMSMAPValueCount            int
}
type NativeStalenessMetadata struct {
	Kind                                                               uint32
	RequestedEpochJ2000S, SourceEpochJ2000S, StalenessS, StalenessDays float64
}
type StalenessPolicy struct{ MaxStalenessS float64 }

const (
	IONEXCoveragePolicyStrictValue             uint32 = 0
	IONEXCoveragePolicyHoldValue               uint32 = 1
	IONEXSlantDelayStatusValidValue            uint32 = 0
	IONEXSlantDelayStatusHeldValue             uint32 = 1
	IONEXCoverageErrorNoneValue                uint32 = 0
	IONEXCoverageErrorEpochBeforeFirstMapValue uint32 = 1
	IONEXCoverageErrorEpochAfterLastMapValue   uint32 = 2
	IONEXCoverageErrorLatitudeValue            uint32 = 3
	IONEXCoverageErrorLongitudeValue           uint32 = 4
)

func (*Ionex) SlantDelayAtInstant(float64, float64, float64, float64, NativeClockEpoch, float64) (float64, IonexEpochError, error) {
	return 0, IonexEpochError{}, protocolUnavailable()
}
func (*Ionex) SlantDelayAtInstantWithPolicy(float64, float64, float64, float64, NativeClockEpoch, float64, IonexSlantPolicy) (IonexInstantSlantRow, error) {
	return IonexInstantSlantRow{}, protocolUnavailable()
}
func (*Ionex) SlantDelayResultsAtInstants([]IonexInstantSlantRequest, IonexSlantPolicy) ([]IonexInstantSlantRow, error) {
	return nil, protocolUnavailable()
}
func (*Ionex) SlantDelayResultsAtInstantsOwned([]IonexInstantSlantRequest, IonexSlantPolicy) (IonexInstantSlantResultList, error) {
	return IonexInstantSlantResultList{}, protocolUnavailable()
}
func SelectIONEXAtInstant([]*Ionex, NativeClockEpoch, StalenessPolicy) (*Ionex, NativeStalenessMetadata, IonexEpochError, error) {
	return nil, NativeStalenessMetadata{}, IonexEpochError{}, protocolUnavailable()
}
func SelectIONEXOverInstantRange([]*Ionex, NativeClockEpoch, NativeClockEpoch, StalenessPolicy) (*Ionex, NativeStalenessMetadata, IonexEpochError, error) {
	return nil, NativeStalenessMetadata{}, IonexEpochError{}, protocolUnavailable()
}

func ParseIONEX([]byte) (*Ionex, error) { return nil, protocolUnavailable() }
func ParseIONEXWithWarnings([]byte) (*Ionex, []IonexWarning, error) {
	return nil, nil, protocolUnavailable()
}
func BuildIONEXFromTECGridSamples(TecGridSamples) (*Ionex, error) { return nil, protocolUnavailable() }

// BuildIONEXFromTECGridSamplesWithOutcome preserves typed data refusals.
func BuildIONEXFromTECGridSamplesWithOutcome(TecGridSamples) (*Ionex, TecSamplesOutcome, error) {
	return nil, TecSamplesOutcome{}, protocolUnavailable()
}
func (*Ionex) HeaderMetadata() (IonexHeaderMetadata, error) {
	return IonexHeaderMetadata{}, protocolUnavailable()
}
func BuildIONEXFromTECSamples([]TecSample, float64, float64, int32) (*Ionex, error) {
	return nil, protocolUnavailable()
}
func BuildIONEXFromTECSamplesWithHeader([]TecSample, float64, float64, int32, *IonexHeaderMetadata) (*Ionex, error) {
	return nil, protocolUnavailable()
}

// BuildIONEXFromTECSamplesWithOutcome preserves typed per-node data refusals.
func BuildIONEXFromTECSamplesWithOutcome([]TecSample, float64, float64, int32) (*Ionex, TecSamplesOutcome, error) {
	return nil, TecSamplesOutcome{}, protocolUnavailable()
}

// BuildIONEXFromTECSamplesWithHeaderOutcome preserves typed data refusals with header input.
func BuildIONEXFromTECSamplesWithHeaderOutcome([]TecSample, float64, float64, int32, *IonexHeaderMetadata) (*Ionex, TecSamplesOutcome, error) {
	return nil, TecSamplesOutcome{}, protocolUnavailable()
}
func (*Ionex) Close() error                       { return nil }
func (*Ionex) EpochCount() (int, error)           { return 0, protocolUnavailable() }
func (*Ionex) SkippedRecords() (int, error)       { return 0, protocolUnavailable() }
func (*Ionex) TECMapPresence() ([]bool, error)    { return nil, protocolUnavailable() }
func (*Ionex) RMSMapPresence() ([]bool, error)    { return nil, protocolUnavailable() }
func (*Ionex) HeightMapPresence() ([]bool, error) { return nil, protocolUnavailable() }
func (*Ionex) HeightMapsKm() ([]float64, error)   { return nil, protocolUnavailable() }
func (*Ionex) Exponent() (int32, error)           { return 0, protocolUnavailable() }
func (*Ionex) LatNodesDeg() ([]float64, error)    { return nil, protocolUnavailable() }
func (*Ionex) LonNodesDeg() ([]float64, error)    { return nil, protocolUnavailable() }
func (*Ionex) MapEpochsJ2000S() ([]int64, error)  { return nil, protocolUnavailable() }
func (*Ionex) ToIONEXText() ([]byte, error)       { return nil, protocolUnavailable() }
func (*Ionex) SlantDelay(float64, float64, float64, float64, int64, float64) (float64, error) {
	return 0, protocolUnavailable()
}
func (*Ionex) SlantDelays([]IonexSlantRequest) ([]float64, error) {
	return nil, protocolUnavailable()
}
func (*Ionex) SlantDelayWithPolicy(float64, float64, float64, float64, int64, float64, uint32) (IonexSlantDelayEvaluation, error) {
	return IonexSlantDelayEvaluation{}, protocolUnavailable()
}
func (*Ionex) TECSamples() ([]TecSample, error) { return nil, protocolUnavailable() }
func (*Ionex) GridInfo() (TecGridSamplesInfo, error) {
	return TecGridSamplesInfo{}, protocolUnavailable()
}
func (*Ionex) TECMAPsTECU() ([]float64, error)      { return nil, protocolUnavailable() }
func (*Ionex) RMSMAPsTECU() ([]float64, error)      { return nil, protocolUnavailable() }
func (*Ionex) GridEpochsJ2000S() ([]float64, error) { return nil, protocolUnavailable() }
func SelectIONEX([]*Ionex, int64, StalenessPolicy) (*Ionex, NativeStalenessMetadata, error) {
	return nil, NativeStalenessMetadata{}, protocolUnavailable()
}
func SelectIONEXOverRange([]*Ionex, int64, int64, StalenessPolicy) (*Ionex, NativeStalenessMetadata, error) {
	return nil, NativeStalenessMetadata{}, protocolUnavailable()
}
