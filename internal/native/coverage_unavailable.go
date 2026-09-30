//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

type CoverageLookAngle struct {
	OK           bool
	AzimuthDeg   float64
	ElevationDeg float64
	RangeKm      float64
}

type CoverageGrid struct{}

func CoverageLookAngles([]*TLE, []GroundStation, int64) (*CoverageGrid, error) {
	return nil, unavailable()
}
func (*CoverageGrid) Close() error { return nil }
func (*CoverageGrid) Dimensions() (int, int, error) {
	return 0, 0, unavailable()
}
func (*CoverageGrid) LookAngle(int, int) (CoverageLookAngle, error) {
	return CoverageLookAngle{}, unavailable()
}
func (*CoverageGrid) LookAngleErrorPayload(int, int) ([]byte, error) {
	return nil, unavailable()
}
func (*CoverageGrid) AccessCounts(float64) ([]int, error) { return nil, unavailable() }
func (*CoverageGrid) MaxElevationDeg() ([]float64, error) { return nil, unavailable() }
func (*CoverageGrid) VisibleMask(float64) ([]bool, error) { return nil, unavailable() }
