//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import "unsafe"

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
	ECEFM               [3]float64
	HasSolidEarthTide   bool
	SolidEarthTideECEFM [3]float64
	HasPoleTide         bool
	PoleTideECEFM       [3]float64
	HasOceanLoading     bool
	OceanLoadingECEFM   [3]float64
	DegradeReason       uint32
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

func stationTideTextLocked(part uint32) (string, error) {
	var written, required C.size_t
	status := C.sidereon_station_tide_last_error_text(C.uint32_t(part), nil, 0, &written, &required)
	if err := statusErrorLocked(uint32(status)); err != nil {
		return "", err
	}
	n, err := sizeTToInt(required, "station-tide error text")
	if err != nil || n == 0 {
		return "", err
	}
	buffer := make([]byte, n)
	status = C.sidereon_station_tide_last_error_text(C.uint32_t(part), (*C.uint8_t)(unsafe.Pointer(&buffer[0])), C.size_t(n), &written, &required)
	if err := statusErrorLocked(uint32(status)); err != nil {
		return "", err
	}
	count, err := validateTwoPassCounts("station-tide error text", n, n, uint64(written), uint64(required))
	if err != nil {
		return "", err
	}
	return string(buffer[:count]), nil
}

func readStationTideErrorLocked(status C.enum_SidereonStatus) (NativeStationTideError, error) {
	var raw C.SidereonStationTideError
	if readStatus := C.sidereon_station_tide_last_error(&raw); readStatus != C.SIDEREON_STATUS_OK {
		return NativeStationTideError{}, statusErrorLocked(uint32(readStatus))
	}
	field, err := stationTideTextLocked(0)
	if err != nil {
		return NativeStationTideError{}, err
	}
	reason, err := stationTideTextLocked(1)
	if err != nil {
		return NativeStationTideError{}, err
	}
	return NativeStationTideError{Kind: uint32(raw.kind), NestedKind: uint32(raw.nested_kind), SunMoonCause: uint32(raw.sun_moon_cause), InputKind: uint32(raw.input_kind), DegradeReason: uint32(raw.degrade_reason), HasField: bool(raw.has_field), HasReason: bool(raw.has_reason), Field: field, Reason: reason}, statusErrorLocked(uint32(status))
}

func stationTideOptionsToC(value NativeStationTideOptions) C.SidereonStationTideOptions {
	out := C.SidereonStationTideOptions{solid_earth_tide: C.bool(value.SolidEarthTide), pole_tide: C.bool(value.PoleTide), constants: C.uint32_t(value.Constants), validity_mode: C.uint32_t(value.ValidityMode)}
	if value.OceanLoading != nil {
		out.has_ocean_loading = true
		out.ocean_loading = cOcean(*value.OceanLoading)
	}
	return out
}

func stationTideDisplacementFromC(value C.SidereonStationTideDisplacement) NativeStationTideDisplacement {
	out := NativeStationTideDisplacement{HasSolidEarthTide: bool(value.has_solid_earth_tide), HasPoleTide: bool(value.has_pole_tide), HasOceanLoading: bool(value.has_ocean_loading), DegradeReason: uint32(value.degrade_reason)}
	for axis := 0; axis < 3; axis++ {
		out.ECEFM[axis] = float64(value.ecef_m[axis])
		out.SolidEarthTideECEFM[axis] = float64(value.solid_earth_tide_ecef_m[axis])
		out.PoleTideECEFM[axis] = float64(value.pole_tide_ecef_m[axis])
		out.OceanLoadingECEFM[axis] = float64(value.ocean_loading_ecef_m[axis])
	}
	return out
}

func stationTideBatchText(row int, part uint32) (string, error) {
	var written, required C.size_t
	status := C.sidereon_station_tide_batch_error_text(C.size_t(row), C.uint32_t(part), nil, 0, &written, &required)
	if err := statusErrorLocked(uint32(status)); err != nil {
		return "", err
	}
	n, err := sizeTToInt(required, "station-tide batch error text")
	if err != nil || n == 0 {
		return "", err
	}
	buffer := make([]byte, n)
	status = C.sidereon_station_tide_batch_error_text(C.size_t(row), C.uint32_t(part), (*C.uint8_t)(unsafe.Pointer(&buffer[0])), C.size_t(n), &written, &required)
	if err := statusErrorLocked(uint32(status)); err != nil {
		return "", err
	}
	count, err := validateTwoPassCounts("station-tide batch error text", n, n, uint64(written), uint64(required))
	if err != nil {
		return "", err
	}
	return string(buffer[:count]), nil
}

func StationTideDisplacement(position [3]float64, epoch NativeStationTideEpoch, options NativeStationTideOptions) (NativeStationTideDisplacement, NativeStationTideError, error) {
	var cEpoch C.SidereonStationTideEpoch
	cEpoch.year, cEpoch.month, cEpoch.day = C.int32_t(epoch.Year), C.uint8_t(epoch.Month), C.uint8_t(epoch.Day)
	cEpoch.hour, cEpoch.minute, cEpoch.second = C.uint8_t(epoch.Hour), C.uint8_t(epoch.Minute), C.double(epoch.Second)
	cEpoch.has_polar_motion, cEpoch.xp_arcsec, cEpoch.yp_arcsec = C.bool(epoch.HasPolarMotion), C.double(epoch.XPArcsec), C.double(epoch.YPArcsec)
	cPosition := [3]C.double{C.double(position[0]), C.double(position[1]), C.double(position[2])}
	cOptions := stationTideOptionsToC(options)
	var result C.SidereonStationTideDisplacement
	var typed NativeStationTideError
	err := withCThreadError(func() error {
		status := C.sidereon_station_tide_displacement(&cPosition[0], &cEpoch, &cOptions, &result)
		if status != C.SIDEREON_STATUS_OK {
			var readErr error
			typed, readErr = readStationTideErrorLocked(status)
			return readErr
		}
		return nil
	})
	return stationTideDisplacementFromC(result), typed, err
}

func StationTideDisplacementBatch(position [3]float64, epochs []NativeStationTideEpoch, options NativeStationTideOptions) ([]NativeStationTideBatchRow, error) {
	rows := make([]C.SidereonStationTideEpoch, len(epochs))
	for index, epoch := range epochs {
		rows[index] = C.SidereonStationTideEpoch{year: C.int32_t(epoch.Year), month: C.uint8_t(epoch.Month), day: C.uint8_t(epoch.Day), hour: C.uint8_t(epoch.Hour), minute: C.uint8_t(epoch.Minute), second: C.double(epoch.Second), has_polar_motion: C.bool(epoch.HasPolarMotion), xp_arcsec: C.double(epoch.XPArcsec), yp_arcsec: C.double(epoch.YPArcsec)}
	}
	output := make([]C.SidereonStationTideBatchRow, len(epochs))
	var inputPointer *C.SidereonStationTideEpoch
	if len(rows) != 0 {
		inputPointer = &rows[0]
	}
	var outputRows *C.SidereonStationTideBatchRow
	if len(output) != 0 {
		outputRows = &output[0]
	}
	cPosition := [3]C.double{C.double(position[0]), C.double(position[1]), C.double(position[2])}
	cOptions := stationTideOptionsToC(options)
	result := make([]NativeStationTideBatchRow, len(output))
	err := withCThreadError(func() error {
		status := C.sidereon_station_tide_displacement_batch(&cPosition[0], inputPointer, C.size_t(len(rows)), &cOptions, outputRows)
		if err := statusErrorLocked(uint32(status)); err != nil {
			return err
		}
		for index, row := range output {
			result[index] = NativeStationTideBatchRow{Status: uint32(row.status), Displacement: stationTideDisplacementFromC(row.displacement)}
			if row.status != C.SIDEREON_STATUS_OK {
				item := NativeStationTideError{Kind: uint32(row.error.kind), NestedKind: uint32(row.error.nested_kind), SunMoonCause: uint32(row.error.sun_moon_cause), InputKind: uint32(row.error.input_kind), DegradeReason: uint32(row.error.degrade_reason), HasField: bool(row.error.has_field), HasReason: bool(row.error.has_reason)}
				if item.HasField {
					text, textErr := stationTideBatchText(index, 0)
					if textErr != nil {
						return textErr
					}
					item.Field = text
				}
				if item.HasReason {
					text, textErr := stationTideBatchText(index, 1)
					if textErr != nil {
						return textErr
					}
					item.Reason = text
				}
				result[index].Error = item
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func StationTideConstantsDefault() uint32 {
	return uint32(C.sidereon_station_tide_constants_default())
}

func SolidEarthTideWithConstants(station [3]float64, year, month, day int32, fractionalHour float64, sun, moon [3]float64, constants uint32) ([3]float64, error) {
	var output [3]C.double
	cStation := [3]C.double{C.double(station[0]), C.double(station[1]), C.double(station[2])}
	cSun := [3]C.double{C.double(sun[0]), C.double(sun[1]), C.double(sun[2])}
	cMoon := [3]C.double{C.double(moon[0]), C.double(moon[1]), C.double(moon[2])}
	err := withCThreadError(func() error {
		return statusErrorLocked(uint32(C.sidereon_solid_earth_tide_with_constants(&cStation[0], C.int32_t(year), C.int32_t(month), C.int32_t(day), C.double(fractionalHour), &cSun[0], &cMoon[0], C.uint32_t(constants), &output[0])))
	})
	return [3]float64{float64(output[0]), float64(output[1]), float64(output[2])}, err
}
