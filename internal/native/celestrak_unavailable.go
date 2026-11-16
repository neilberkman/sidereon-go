//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

type CelestrakGpResponse struct{}

func CelestrakText(uint32, uint32, string, uint64, uint32) (string, error) { return "", unavailable() }
func ParseCelestrakGpResponse([]byte, uint32) (*CelestrakGpResponse, error) {
	return nil, unavailable()
}
func (*CelestrakGpResponse) Close() error                                { return nil }
func (*CelestrakGpResponse) Counts() (int, int, error)                   { return 0, 0, unavailable() }
func (*CelestrakGpResponse) RecordBytes(int, bool, bool) ([]byte, error) { return nil, unavailable() }
func (*CelestrakGpResponse) RecordOMM(int) (*OMM, error)                 { return nil, unavailable() }
func (*CelestrakGpResponse) Equal(*CelestrakGpResponse) (bool, error)    { return false, unavailable() }
func (*CelestrakGpResponse) RecordTLE(int, uint32, uint32) (*TLE, error) { return nil, unavailable() }
func (*CelestrakGpResponse) RecordElementSet(int) (OMMElementSet, error) {
	return OMMElementSet{}, unavailable()
}
func (*CelestrakGpResponse) CompareRecord(*CelestrakGpResponse, int, int, uint32) (bool, error) {
	return false, unavailable()
}
