//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

type SatelliteConstellation struct{}
type ConstellationLookAngleArcs struct{}
type ConstellationGroundTracks struct{}
type ConstellationPasses struct{}
type FleetPass struct {
	SatelliteIndex int
	Pass           SatellitePass
}

func BuildSatelliteConstellation([]*TLE) (*SatelliteConstellation, error) {
	return nil, unavailable()
}
func (*SatelliteConstellation) Close() error                      { return nil }
func (*SatelliteConstellation) Count() (int, error)               { return 0, unavailable() }
func (*SatelliteConstellation) CatalogNumber(int) (string, error) { return "", unavailable() }
func (*SatelliteConstellation) Propagate([]int64, bool) (*TLEBatchPropagation, error) {
	return nil, unavailable()
}
func (*SatelliteConstellation) Visible(GroundStation, int64, float64) (*VisibleList, error) {
	return nil, unavailable()
}
func (*SatelliteConstellation) LookAngles(GroundStation, []int64, bool) (*ConstellationLookAngleArcs, error) {
	return nil, unavailable()
}
func (*SatelliteConstellation) GroundTracks([]int64) (*ConstellationGroundTracks, error) {
	return nil, unavailable()
}
func (*SatelliteConstellation) Passes(GroundStation, int64, int64, *PassFinderOptions) (*ConstellationPasses, error) {
	return nil, unavailable()
}

func (*ConstellationLookAngleArcs) Close() error                     { return nil }
func (*ConstellationLookAngleArcs) SatelliteCount() (int, error)     { return 0, unavailable() }
func (*ConstellationLookAngleArcs) ArcLengths() ([]int, error)       { return nil, unavailable() }
func (*ConstellationLookAngleArcs) Values() ([]LookAngle, error)     { return nil, unavailable() }
func (*ConstellationLookAngleArcs) ErrorPayload(int) ([]byte, error) { return nil, unavailable() }

func (*ConstellationGroundTracks) Close() error                     { return nil }
func (*ConstellationGroundTracks) SatelliteCount() (int, error)     { return 0, unavailable() }
func (*ConstellationGroundTracks) TrackLengths() ([]int, error)     { return nil, unavailable() }
func (*ConstellationGroundTracks) Values() ([]Geodetic, error)      { return nil, unavailable() }
func (*ConstellationGroundTracks) ErrorPayload(int) ([]byte, error) { return nil, unavailable() }

func (*ConstellationPasses) Close() error                     { return nil }
func (*ConstellationPasses) Count() (int, error)              { return 0, unavailable() }
func (*ConstellationPasses) Values() ([]FleetPass, error)     { return nil, unavailable() }
func (*ConstellationPasses) ErrorPayload(int) ([]byte, error) { return nil, unavailable() }

func SatelliteVisualMagnitude(float64, float64, float64, float64) (float64, error) {
	return 0, unavailable()
}
