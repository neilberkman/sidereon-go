//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

import "errors"

// This build is intentionally link-free. Use cgo and, on Linux, select one of
// sidereon_linux_glibc or sidereon_linux_musl. Use sidereon_use_system_lib
// with CGO_LDFLAGS when linking a system installation.
var ErrClosed = errors.New("sidereon: handle is closed")

type StatusError struct {
	Code         int
	Text         string
	Detail       string
	TerrainDatum *TerrainDatumError
	TerrainStore *TerrainStoreError
}

func (e *StatusError) Error() string { return e.Text }

func (e *StatusError) Unwrap() error {
	var details []error
	if e.TerrainDatum != nil {
		details = append(details, e.TerrainDatum)
	}
	if e.TerrainStore != nil {
		details = append(details, e.TerrainStore)
	}
	return errors.Join(details...)
}

type Version struct {
	Major  uint32
	Minor  uint32
	Patch  uint32
	String string
}

func LibraryVersion() Version { return Version{} }

func unavailable() error {
	return errors.New("sidereon: this platform or build configuration requires cgo and a supported native ABI selection")
}

func SecondOfDay(int, int, float64) (float64, error)              { return 0, unavailable() }
func DayOfYear(int, int, int, int, int, float64) (float64, error) { return 0, unavailable() }
func DataDayOfYear(int, uint8, uint8) (uint16, error)             { return 0, unavailable() }

type CovarianceValidation struct {
	Symmetric            bool
	PositiveSemidefinite bool
}

func CovarianceFromDiagonal([]float64) ([6][6]float64, error) {
	return [6][6]float64{}, unavailable()
}
func CovarianceValidate([6][6]float64) (CovarianceValidation, error) {
	return CovarianceValidation{}, unavailable()
}
func CovarianceKmToM([6][6]float64) ([6][6]float64, error) {
	return [6][6]float64{}, unavailable()
}
func CovarianceMToKm([6][6]float64) ([6][6]float64, error) {
	return [6][6]float64{}, unavailable()
}
func CovarianceInterpolate([6][6]float64, [6][6]float64, float64) ([6][6]float64, error) {
	return [6][6]float64{}, unavailable()
}

type NMEASummary struct {
	SentenceCount uint64
	EpochCount    uint64
	SkipCount     uint64
	WarningCount  uint64
}

type NMEAEpoch struct {
	HasPosition        bool
	LatitudeRad        float64
	LongitudeRad       float64
	HeightM            float64
	SentenceCount      uint64
	UsedSatelliteCount uint64
	SatellitesInView   uint64
	SkipCount          uint64
	WarningCount       uint64
}

type NMEALog struct{}

func ParseNMEA([]byte) (*NMEALog, error) { return nil, unavailable() }
func (*NMEALog) Close() error            { return nil }
func (*NMEALog) Summary() (NMEASummary, error) {
	return NMEASummary{}, unavailable()
}
func (*NMEALog) Epochs() ([]NMEAEpoch, error) { return nil, unavailable() }
