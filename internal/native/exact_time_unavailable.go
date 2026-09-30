//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

type ExactEpochComponents struct {
	Seconds       int64
	Attoseconds   uint64
	ResidueDigits int64
	ResiduePlaces uint16
}

type ExactOrdering int32

const (
	ExactOrderingLess    ExactOrdering = -1
	ExactOrderingEqual   ExactOrdering = 0
	ExactOrderingGreater ExactOrdering = 1
)

type ExactEpoch struct{}
type ExactEpochQuery struct{}

func NewExactEpoch(int64, uint64) (*ExactEpoch, error)        { return nil, unavailableValue() }
func ExactEpochJ2000() (*ExactEpoch, error)                   { return nil, unavailableValue() }
func ExactAttosecondsPerSecond() (uint64, error)              { return 0, unavailableValue() }
func ExactEpochFromJ2000Seconds(float64) (*ExactEpoch, error) { return nil, unavailableValue() }
func ExactEpochFromCivil(CivilDateTime) (*ExactEpoch, error)  { return nil, unavailableValue() }
func ExactEpochQueryFromBinaryJ2000Seconds(float64) (*ExactEpochQuery, error) {
	return nil, unavailableValue()
}
func (*ExactEpoch) Query() (*ExactEpochQuery, error) { return nil, unavailableValue() }
func (*ExactEpochQuery) Epoch() (*ExactEpoch, error) { return nil, unavailableValue() }
func (*ExactEpoch) Compare(*ExactEpoch) (ExactOrdering, error) {
	return ExactOrderingEqual, unavailableValue()
}
func (*ExactEpoch) Equal(*ExactEpoch) (bool, error)                { return false, unavailableValue() }
func (*ExactEpochQuery) Equal(*ExactEpochQuery) (bool, error)      { return false, unavailableValue() }
func (*ExactEpoch) CheckedAddSeconds(float64) (*ExactEpoch, error) { return nil, unavailableValue() }
func (*ExactEpoch) CheckedSubSeconds(float64) (*ExactEpoch, error) { return nil, unavailableValue() }
func (*ExactEpochQuery) CheckedAddBinarySeconds(float64) (*ExactEpochQuery, error) {
	return nil, unavailableValue()
}
func (*ExactEpochQuery) CheckedSubBinarySeconds(float64) (*ExactEpochQuery, error) {
	return nil, unavailableValue()
}
func (*ExactEpoch) Components() (ExactEpochComponents, error) {
	return ExactEpochComponents{}, unavailableValue()
}
func (*ExactEpoch) J2000Seconds() (float64, error)                 { return 0, unavailableValue() }
func (*ExactEpoch) SplitJulianDate() (JulianDate, error)           { return JulianDate{}, unavailableValue() }
func (*ExactEpoch) SecondsSince(*ExactEpoch) (float64, error)      { return 0, unavailableValue() }
func (*ExactEpochQuery) SecondsSince(*ExactEpoch) (float64, error) { return 0, unavailableValue() }
func (*ExactEpochQuery) SecondsSinceQuery(*ExactEpochQuery) (float64, error) {
	return 0, unavailableValue()
}
func (*ExactEpochQuery) J2000Seconds() (float64, error) { return 0, unavailableValue() }
func (*ExactEpoch) Close() error                        { return nil }
func (*ExactEpochQuery) Close() error                   { return nil }
