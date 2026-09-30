//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

const (
	Sp3ProvenanceModeOff     = uint32(0)
	Sp3ProvenanceModeSummary = uint32(1)
	Sp3ProvenanceModeFull    = uint32(2)
)

type NativeSp3AgreementMetric struct {
	Epoch              NativeClockEpoch
	EpochJ2000S        float64
	Satellite          string
	PositionMembers    int
	PositionRMSPresent bool
	PositionRMSM       float64
	PositionMaxPresent bool
	PositionMaxM       float64
	ClockMembers       int
	ClockRMSPresent    bool
	ClockRMSS          float64
	ClockMaxPresent    bool
	ClockMaxS          float64
}
type NativeSp3ClockOmission struct {
	Epoch        NativeClockEpoch
	EpochJ2000S  float64
	Satellite    string
	Source       int
	Reason       uint32
	HasPreferred bool
	Preferred    int
	CellHasClock bool
}
type NativeSp3DroppedInputEpoch struct {
	Source      int
	EpochIndex  int
	Epoch       NativeClockEpoch
	EpochJ2000S float64
	Reason      uint32
}
type NativeSp3MergeEpoch struct {
	Epoch       NativeClockEpoch
	EpochJ2000S float64
}
type NativeSp3ProvenanceInfo struct {
	Recorded        bool
	Mode            uint32
	CellCount       int
	TransitionCount int
	CoverageCount   int
}
type NativeSp3CellSelection struct {
	Kind        uint32
	HasSource   bool
	Source      int
	HasRule     bool
	Rule        uint32
	MemberCount int
}
type NativeSp3CellProvenance struct {
	Epoch       NativeClockEpoch
	EpochJ2000S float64
	Satellite   string
	HasPosition bool
	Position    NativeSp3CellSelection
	HasClock    bool
	Clock       NativeSp3CellSelection
}
type NativeSp3ContributorCoverage struct {
	Source           int
	CellsContributed int
	CellsSelected    int
	HasFirstEpoch    bool
	FirstEpoch       NativeClockEpoch
	FirstEpochJ2000S float64
	HasLastEpoch     bool
	LastEpoch        NativeClockEpoch
	LastEpochJ2000S  float64
	CellsAbsent      int
}
type NativeSp3PrecedenceTransition struct {
	Satellite     string
	Epoch         NativeClockEpoch
	EpochJ2000S   float64
	HasFromSource bool
	FromSource    int
	HasToSource   bool
	ToSource      int
	Reason        uint32
}

type Sp3MergeReport struct{}

func (*Sp3MergeReport) AgreementMetrics() ([]NativeSp3AgreementMetric, error) {
	return nil, ErrUnavailable
}
func (*Sp3MergeReport) ClockOmissions() ([]NativeSp3ClockOmission, error) { return nil, ErrUnavailable }
func (*Sp3MergeReport) DroppedInputEpochs() ([]NativeSp3DroppedInputEpoch, error) {
	return nil, ErrUnavailable
}
func (*Sp3MergeReport) OmittedEpochs() ([]NativeSp3MergeEpoch, error) { return nil, ErrUnavailable }
func (*Sp3MergeReport) ProvenanceInfo() (NativeSp3ProvenanceInfo, error) {
	return NativeSp3ProvenanceInfo{}, ErrUnavailable
}
func (*Sp3MergeReport) ProvenanceCells() ([]NativeSp3CellProvenance, error) {
	return nil, ErrUnavailable
}
func (*Sp3MergeReport) ProvenanceCellMembers(int, uint32) ([]int, error) { return nil, ErrUnavailable }
func (*Sp3MergeReport) ProvenanceCoverage() ([]NativeSp3ContributorCoverage, error) {
	return nil, ErrUnavailable
}
func (*Sp3MergeReport) ProvenanceTransitions() ([]NativeSp3PrecedenceTransition, error) {
	return nil, ErrUnavailable
}
func (*Sp3MergeReport) ContinuityJSON() ([]byte, error) { return nil, ErrUnavailable }
func (*Sp3MergeReport) ContinuitySelectedNodes(string, float64, float64) (bool, []float64, error) {
	return false, nil, ErrUnavailable
}
