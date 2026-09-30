//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

type Ionex struct{}
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
	TimeScale                             uint32
	EpochJ2000S, LatDeg, LonDeg, VTECTECU float64
	RMSPresent                            bool
	RMSTECU                               float64
}
type TecGridSamples struct {
	TimeScale                                     uint32
	MapEpochsJ2000S, LatNodesDeg, LonNodesDeg     []float64
	DLatDeg, DLonDeg, ShellHeightKm, BaseRadiusKm float64
	Exponent                                      int32
	TECMAPsTECU                                   []float64
	RMSPresent                                    bool
	RMSMAPsTECU                                   []float64
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

func ParseIONEX([]byte) (*Ionex, error)                           { return nil, protocolUnavailable() }
func BuildIONEXFromTECGridSamples(TecGridSamples) (*Ionex, error) { return nil, protocolUnavailable() }
func BuildIONEXFromTECSamples([]TecSample, float64, float64, int32) (*Ionex, error) {
	return nil, protocolUnavailable()
}
func (*Ionex) Close() error                      { return nil }
func (*Ionex) EpochCount() (int, error)          { return 0, protocolUnavailable() }
func (*Ionex) Exponent() (int32, error)          { return 0, protocolUnavailable() }
func (*Ionex) LatNodesDeg() ([]float64, error)   { return nil, protocolUnavailable() }
func (*Ionex) LonNodesDeg() ([]float64, error)   { return nil, protocolUnavailable() }
func (*Ionex) MapEpochsJ2000S() ([]int64, error) { return nil, protocolUnavailable() }
func (*Ionex) ToIONEXText() ([]byte, error)      { return nil, protocolUnavailable() }
func (*Ionex) SlantDelay(float64, float64, float64, float64, int64, float64) (float64, error) {
	return 0, protocolUnavailable()
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
