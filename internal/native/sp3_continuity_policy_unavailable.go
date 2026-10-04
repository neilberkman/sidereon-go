//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

type Sp3ContinuityOptions struct {
	SpeedBoundKind           uint32
	OrbitClass               uint32
	ExplicitMaxSpeedMPS      float64
	ResidualToleranceEnabled bool
	ResidualToleranceM       float64
	GapThresholdFactor       float64
}

func DefaultSp3ContinuityOptions(uint32) (Sp3ContinuityOptions, error) {
	return Sp3ContinuityOptions{}, ErrUnavailable
}
func (*SP3) ContinuityReportJSON(Sp3ContinuityOptions) ([]byte, error) { return nil, ErrUnavailable }
func (*SP3) ContinuityVerdictJSONWithPolicy(Sp3ContinuityOptions, float64, float64) ([]byte, error) {
	return nil, ErrUnavailable
}
