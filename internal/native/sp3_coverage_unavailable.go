//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

const (
	SP3ChannelPosition = uint32(0)
	SP3ChannelClock    = uint32(1)
)

type SP3CoverageGrid struct {
	HasInterval      bool
	IntervalS        float64
	AgreesWithHeader bool
	OutOfOrder       []int
	Unplaced         []int
}
type SP3ChannelCoverage struct {
	Epochs, SpanCount, GapCount int
	Complete                    bool
	Spans                       []SP3CoverageSpan
	Gaps                        []SP3CoverageGap
}
type SP3CoverageSatellite struct {
	Satellite string
	Declared  bool
	Positions SP3ChannelCoverage
	Clocks    SP3ChannelCoverage
}
type SP3CoverageSpan struct {
	FirstIndex, LastIndex int
	FirstEpoch            NativeClockEpoch
	FirstJ2000            float64
	LastEpoch             NativeClockEpoch
	LastJ2000             float64
}
type SP3CoverageGap struct {
	HasAfterIndex, HasBeforeIndex bool
	AfterIndex, BeforeIndex       int
	MissingEpochs                 int
}
type SP3Coverage struct {
	Grid       SP3CoverageGrid
	Satellites []SP3CoverageSatellite
}

func (*SP3) Coverage() (SP3Coverage, error)                            { return SP3Coverage{}, ErrUnavailable }
func (*SP3) SelectedNodes(string, float64, float64) ([]float64, error) { return nil, ErrUnavailable }
