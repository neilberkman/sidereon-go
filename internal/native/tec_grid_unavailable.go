//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

type TecGrid struct{}

func NewTecGrid(TecGridInput) (*TecGrid, *NativeTecGridError, error) {
	return nil, nil, protocolUnavailable()
}
func (*TecGrid) Close() error { return nil }
func (*TecGrid) Dimensions() (NativeTecGridInfo, error) {
	return NativeTecGridInfo{}, protocolUnavailable()
}
func (*TecGrid) Epochs() ([]float64, error)     { return nil, protocolUnavailable() }
func (*TecGrid) Latitudes() ([]float64, error)  { return nil, protocolUnavailable() }
func (*TecGrid) Longitudes() ([]float64, error) { return nil, protocolUnavailable() }
func (*TecGrid) Values() ([]float64, error)     { return nil, protocolUnavailable() }
func (*TecGrid) Presence() ([]bool, error)      { return nil, protocolUnavailable() }
func (*TecGrid) VTECAt(int64, float64, float64, uint32) (NativeTecGridVTEC, error) {
	return NativeTecGridVTEC{}, protocolUnavailable()
}
