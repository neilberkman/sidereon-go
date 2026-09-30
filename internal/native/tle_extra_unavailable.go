//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

import "time"

type TLEFile struct{}
type NativeTLERejectedRecord struct {
	LineNumber  int
	Issue       uint32
	Name, Error string
}
type TLEBatchPropagation struct{}
type TLEBatchLookAngles struct{}
type VisibleSatellite struct {
	CatalogNumber                     string
	AzimuthDeg, ElevationDeg, RangeKm float64
	PositionKm                        [3]float64
}
type VisibleList struct{}
type SGP4DecayLatch struct{}
type TLEPair struct{ Line1, Line2 string }

func ParseTLEFile([]byte, uint32) (*TLEFile, error)                   { return nil, unavailable() }
func ParseTLEFileWithPolicy([]byte, uint32, uint32) (*TLEFile, error) { return nil, unavailable() }
func (*TLEFile) Close() error                                         { return nil }
func (*TLEFile) Count() (int, error)                                  { return 0, unavailable() }
func (*TLEFile) LineNumber(int) (int, error)                          { return 0, unavailable() }
func (*TLEFile) Skipped() (int, error)                                { return 0, unavailable() }
func (*TLEFile) Rejected(int) (NativeTLERejectedRecord, error) {
	return NativeTLERejectedRecord{}, unavailable()
}
func (*TLEFile) Name(int) (string, error)    { return "", unavailable() }
func (*TLEFile) Satellite(int) (*TLE, error) { return nil, unavailable() }

func PropagateTLEBatch([]TLEPair, []time.Time, uint32, bool) (*TLEBatchPropagation, error) {
	return nil, unavailable()
}
func (*TLEBatchPropagation) Close() error                 { return nil }
func (*TLEBatchPropagation) Shape() (int, int, error)     { return 0, 0, unavailable() }
func (*TLEBatchPropagation) States() ([]TEMEState, error) { return nil, unavailable() }

func LookAnglesBatch([]TLEPair, GroundStation, []time.Time, uint32, bool) (*TLEBatchLookAngles, error) {
	return nil, unavailable()
}
func (*TLEBatchLookAngles) Close() error                 { return nil }
func (*TLEBatchLookAngles) Shape() (int, int, error)     { return 0, 0, unavailable() }
func (*TLEBatchLookAngles) Values() ([]LookAngle, error) { return nil, unavailable() }
func (*TLE) PropagateWithDecayLatch(float64, *SGP4DecayLatch) (TEMEState, error) {
	return TEMEState{}, unavailable()
}
func NewSGP4DecayLatch() (*SGP4DecayLatch, error) { return nil, unavailable() }
func (*SGP4DecayLatch) Close() error              { return nil }
func (*SGP4DecayLatch) Clear() error              { return unavailable() }
func (*SGP4DecayLatch) FirstFailingEpoch() (float64, bool, error) {
	return 0, false, unavailable()
}
func VisibleFromSatellites([]*TLE, []string, GroundStation, time.Time, float64) (*VisibleList, error) {
	return nil, unavailable()
}
func (*VisibleList) Close() error                        { return nil }
func (*VisibleList) Values() ([]VisibleSatellite, error) { return nil, unavailable() }
func (*VisibleList) Count() (int, error)                 { return 0, unavailable() }
