//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

type OceanLoadingBLQ struct {
	AmplitudeM [3][11]float64
	PhaseDeg   [3][11]float64
}

const (
	StationTideConstantsConventions = uint32(0)
	StationTideConstantsIERSRoutine = uint32(1)
	StationTideValidityStrict       = uint32(0)
	StationTideValidityPermissive   = uint32(1)
)

type NativeStationTideEpoch struct {
	Year, Month, Day, Hour, Minute int32
	Second                         float64
	HasPolarMotion                 bool
	XPArcsec, YPArcsec             float64
}
type NativeStationTideOptions struct {
	SolidEarthTide, PoleTide bool
	OceanLoading             *OceanLoadingBLQ
	Constants, ValidityMode  uint32
}
type NativeStationTideDisplacement struct {
	ECEFM, SolidEarthTideECEFM, PoleTideECEFM, OceanLoadingECEFM [3]float64
	HasSolidEarthTide, HasPoleTide, HasOceanLoading              bool
	DegradeReason                                                uint32
}
type NativeStationTideError struct {
	Kind, NestedKind, SunMoonCause, InputKind, DegradeReason uint32
	HasField, HasReason                                      bool
	Field, Reason                                            string
}
type NativeStationTideBatchRow struct {
	Status       uint32
	Displacement NativeStationTideDisplacement
	Error        NativeStationTideError
}

func StationTideDisplacement([3]float64, NativeStationTideEpoch, NativeStationTideOptions) (NativeStationTideDisplacement, NativeStationTideError, error) {
	return NativeStationTideDisplacement{}, NativeStationTideError{}, unavailable()
}
func StationTideDisplacementBatch([3]float64, []NativeStationTideEpoch, NativeStationTideOptions) ([]NativeStationTideBatchRow, error) {
	return nil, unavailable()
}
func StationTideConstantsDefault() uint32 { return StationTideConstantsConventions }
func SolidEarthTideWithConstants([3]float64, int32, int32, int32, float64, [3]float64, [3]float64, uint32) ([3]float64, error) {
	return [3]float64{}, unavailable()
}
