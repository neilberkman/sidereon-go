//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"runtime"
	"unsafe"
)

type SpaceWeather struct {
	F107  float64
	F107A float64
	Ap    float64
}

// SpaceWeatherObservationClass mirrors the C source enum. The generated C
// header carries these fields as uint32_t for ABI compatibility.
type SpaceWeatherObservationClass uint32

const (
	SpaceWeatherObservationObservedValue         SpaceWeatherObservationClass = 0
	SpaceWeatherObservationInterpolatedValue     SpaceWeatherObservationClass = 1
	SpaceWeatherObservationDailyPredictedValue   SpaceWeatherObservationClass = 2
	SpaceWeatherObservationMonthlyPredictedValue SpaceWeatherObservationClass = 3
)

type SpaceWeatherCoverage struct {
	FirstJ2000S              float64
	HasLastObservedJ2000S    bool
	LastObservedJ2000S       float64
	HasLastDailyPredictedS   bool
	LastDailyPredictedJ2000S float64
	EndJ2000S                float64
}

type SpaceWeatherDay struct {
	Year               int
	Month              uint8
	Day                uint8
	Class              SpaceWeatherObservationClass
	HasBSRN            bool
	BSRN               uint16
	HasND              bool
	ND                 uint8
	HasKp              [8]bool
	Kp10               [8]uint16
	HasKpSum10         bool
	KpSum10            uint16
	HasAp              [8]bool
	Ap8                [8]uint16
	HasApAvg           bool
	ApAvg              uint16
	HasCP10            bool
	CP10               uint8
	HasC9              bool
	C9                 uint8
	HasISN             bool
	ISN                uint16
	HasFluxQualifier   bool
	FluxQualifier      uint8
	HasF107Obs         bool
	F107Obs            float64
	HasF107Adj         bool
	F107Adj            float64
	HasF107ObsCenter81 bool
	F107ObsCenter81    float64
	HasF107ObsLast81   bool
	F107ObsLast81      float64
	HasF107AdjCenter81 bool
	F107AdjCenter81    float64
	HasF107AdjLast81   bool
	F107AdjLast81      float64
}

type SpaceWeatherSample struct {
	Weather     SpaceWeather
	Class       SpaceWeatherObservationClass
	ApDefaulted bool
}

type SpaceWeatherPolicy struct {
	AllowInterpolated     bool
	AllowDailyPredicted   bool
	AllowMonthlyPredicted bool
	RequireGeomagnetic    bool
}

type SpaceWeatherTableSummary struct {
	DayCount     int
	MonthlyCount int
	SkipCount    int
	WarningCount int
}

type SpaceWeatherTable struct {
	handle *positioningHandle
}

func spaceWeatherFromC(value C.SidereonSpaceWeather) SpaceWeather {
	return SpaceWeather{F107: float64(value.f107), F107A: float64(value.f107a), Ap: float64(value.ap)}
}

func DefaultSpaceWeather() (SpaceWeather, error) {
	var output C.SidereonSpaceWeather
	err := callStatus(func() uint32 { return C.sidereon_space_weather_default(&output) })
	return spaceWeatherFromC(output), err
}

func spaceWeatherDayFromC(value C.SidereonSpaceWeatherDay) SpaceWeatherDay {
	var output SpaceWeatherDay
	output.Year = int(value.year)
	output.Month = uint8(value.month)
	output.Day = uint8(value.day)
	output.Class = SpaceWeatherObservationClass(value.class_)
	output.HasBSRN = bool(value.has_bsrn)
	output.BSRN = uint16(value.bsrn)
	output.HasND = bool(value.has_nd)
	output.ND = uint8(value.nd)
	for i := range output.HasKp {
		output.HasKp[i] = bool(value.has_kp[i])
		output.Kp10[i] = uint16(value.kp_10[i])
		output.HasAp[i] = bool(value.has_ap[i])
		output.Ap8[i] = uint16(value.ap[i])
	}
	output.HasKpSum10 = bool(value.has_kp_sum_10)
	output.KpSum10 = uint16(value.kp_sum_10)
	output.HasApAvg = bool(value.has_ap_avg)
	output.ApAvg = uint16(value.ap_avg)
	output.HasCP10 = bool(value.has_cp_10)
	output.CP10 = uint8(value.cp_10)
	output.HasC9 = bool(value.has_c9)
	output.C9 = uint8(value.c9)
	output.HasISN = bool(value.has_isn)
	output.ISN = uint16(value.isn)
	output.HasFluxQualifier = bool(value.has_flux_qualifier)
	output.FluxQualifier = uint8(value.flux_qualifier)
	output.HasF107Obs = bool(value.has_f107_obs)
	output.F107Obs = float64(value.f107_obs)
	output.HasF107Adj = bool(value.has_f107_adj)
	output.F107Adj = float64(value.f107_adj)
	output.HasF107ObsCenter81 = bool(value.has_f107_obs_center81)
	output.F107ObsCenter81 = float64(value.f107_obs_center81)
	output.HasF107ObsLast81 = bool(value.has_f107_obs_last81)
	output.F107ObsLast81 = float64(value.f107_obs_last81)
	output.HasF107AdjCenter81 = bool(value.has_f107_adj_center81)
	output.F107AdjCenter81 = float64(value.f107_adj_center81)
	output.HasF107AdjLast81 = bool(value.has_f107_adj_last81)
	output.F107AdjLast81 = float64(value.f107_adj_last81)
	return output
}

func spaceWeatherSampleFromC(value C.SidereonSpaceWeatherSample) SpaceWeatherSample {
	return SpaceWeatherSample{Weather: spaceWeatherFromC(value.weather), Class: SpaceWeatherObservationClass(value.class_), ApDefaulted: bool(value.ap_defaulted)}
}

func spaceWeatherTableFromPointer(pointer *C.SidereonSpaceWeatherTable) (*SpaceWeatherTable, error) {
	if pointer == nil {
		return nil, errors.New("sidereon: native space-weather constructor returned no handle")
	}
	return &SpaceWeatherTable{handle: newPositioningHandle(unsafe.Pointer(pointer), func(value unsafe.Pointer) {
		C.sidereon_space_weather_table_free((*C.SidereonSpaceWeatherTable)(value))
	})}, nil
}

func spaceWeatherParse(data []byte, parser func(*C.uint8_t, C.size_t, **C.SidereonSpaceWeatherTable) uint32) (*SpaceWeatherTable, error) {
	var pointer *C.SidereonSpaceWeatherTable
	operationErr := withInput(data, func(input *C.uint8_t, length C.size_t) uint32 {
		return parser(input, length, &pointer)
	})
	if operationErr != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_space_weather_table_free(pointer) })
		}
		return nil, operationErr
	}
	return spaceWeatherTableFromPointer(pointer)
}

func ParseSpaceWeatherTable(data []byte) (*SpaceWeatherTable, error) {
	return spaceWeatherParse(data, func(input *C.uint8_t, length C.size_t, out **C.SidereonSpaceWeatherTable) uint32 {
		return C.sidereon_space_weather_table_parse(input, length, out)
	})
}

func ParseSpaceWeatherCSV(data []byte) (*SpaceWeatherTable, error) {
	return spaceWeatherParse(data, func(input *C.uint8_t, length C.size_t, out **C.SidereonSpaceWeatherTable) uint32 {
		return C.sidereon_space_weather_table_parse_csv(input, length, out)
	})
}

func ParseSpaceWeatherTXT(data []byte) (*SpaceWeatherTable, error) {
	return spaceWeatherParse(data, func(input *C.uint8_t, length C.size_t, out **C.SidereonSpaceWeatherTable) uint32 {
		return C.sidereon_space_weather_table_parse_txt(input, length, out)
	})
}

func (t *SpaceWeatherTable) Close() error {
	if t == nil || t.handle == nil {
		return nil
	}
	return t.handle.close()
}

func copySpaceWeatherDays(call func(*C.SidereonSpaceWeatherDay, C.size_t, *C.size_t, *C.size_t) uint32) ([]SpaceWeatherDay, error) {
	var result []SpaceWeatherDay
	var operationErr error
	withCThread(func() {
		var written, required C.size_t
		if operationErr = statusErrorLocked(call(nil, 0, &written, &required)); operationErr != nil {
			return
		}
		queryWritten, err := checkedNativeCount(uint64(written))
		if err != nil {
			operationErr = err
			return
		}
		if queryWritten != 0 {
			operationErr = errors.New("sidereon: native day output query wrote data")
			return
		}
		requiredCount, err := checkedNativeCount(uint64(required))
		if err != nil {
			operationErr = err
			return
		}
		if _, err := checkedNativeAllocationSize(requiredCount, unsafe.Sizeof(C.SidereonSpaceWeatherDay{})); err != nil {
			operationErr = err
			return
		}
		buffer := make([]C.SidereonSpaceWeatherDay, requiredCount)
		var pointer *C.SidereonSpaceWeatherDay
		if len(buffer) != 0 {
			pointer = &buffer[0]
		}
		if operationErr = statusErrorLocked(call(pointer, C.size_t(len(buffer)), &written, &required)); operationErr != nil {
			return
		}
		writtenCount, err := validateTwoPassCounts("day output", len(buffer), len(buffer), uint64(written), uint64(required))
		if err != nil {
			operationErr = err
			return
		}
		result = make([]SpaceWeatherDay, writtenCount)
		for i := range result {
			result[i] = spaceWeatherDayFromC(buffer[i])
		}
	})
	return result, operationErr
}

func (t *SpaceWeatherTable) Days() ([]SpaceWeatherDay, error) {
	if t == nil || t.handle == nil {
		return nil, ErrClosed
	}
	var result []SpaceWeatherDay
	err := t.handle.with(func(pointer unsafe.Pointer) error {
		var operationErr error
		result, operationErr = copySpaceWeatherDays(func(out *C.SidereonSpaceWeatherDay, length C.size_t, written, required *C.size_t) uint32 {
			return C.sidereon_space_weather_table_days((*C.SidereonSpaceWeatherTable)(pointer), out, length, written, required)
		})
		return operationErr
	})
	runtime.KeepAlive(t)
	return result, err
}

func (t *SpaceWeatherTable) Monthly() ([]SpaceWeatherDay, error) {
	if t == nil || t.handle == nil {
		return nil, ErrClosed
	}
	var result []SpaceWeatherDay
	err := t.handle.with(func(pointer unsafe.Pointer) error {
		var operationErr error
		result, operationErr = copySpaceWeatherDays(func(out *C.SidereonSpaceWeatherDay, length C.size_t, written, required *C.size_t) uint32 {
			return C.sidereon_space_weather_table_monthly((*C.SidereonSpaceWeatherTable)(pointer), out, length, written, required)
		})
		return operationErr
	})
	runtime.KeepAlive(t)
	return result, err
}

func (t *SpaceWeatherTable) Day(year int, month, day uint8) (SpaceWeatherDay, bool, error) {
	if t == nil || t.handle == nil {
		return SpaceWeatherDay{}, false, ErrClosed
	}
	checkedYear, err := checkedInt32(year, "space-weather year")
	if err != nil {
		return SpaceWeatherDay{}, false, err
	}
	var output C.SidereonSpaceWeatherDay
	var present C.bool
	err = t.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_space_weather_table_day((*C.SidereonSpaceWeatherTable)(pointer), C.int32_t(checkedYear), C.uint8_t(month), C.uint8_t(day), &present, &output)
		})
	})
	runtime.KeepAlive(t)
	return spaceWeatherDayFromC(output), bool(present), err
}

func (t *SpaceWeatherTable) Coverage() (SpaceWeatherCoverage, error) {
	if t == nil || t.handle == nil {
		return SpaceWeatherCoverage{}, ErrClosed
	}
	var output C.SidereonSpaceWeatherCoverage
	err := t.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_space_weather_table_coverage((*C.SidereonSpaceWeatherTable)(pointer), &output)
		})
	})
	runtime.KeepAlive(t)
	return SpaceWeatherCoverage{FirstJ2000S: float64(output.first_j2000_s), HasLastObservedJ2000S: bool(output.has_last_observed_j2000_s), LastObservedJ2000S: float64(output.last_observed_j2000_s), HasLastDailyPredictedS: bool(output.has_last_daily_predicted_j2000_s), LastDailyPredictedJ2000S: float64(output.last_daily_predicted_j2000_s), EndJ2000S: float64(output.end_j2000_s)}, err
}

func (t *SpaceWeatherTable) Summary() (SpaceWeatherTableSummary, error) {
	if t == nil || t.handle == nil {
		return SpaceWeatherTableSummary{}, ErrClosed
	}
	var output C.SidereonSpaceWeatherTableSummary
	err := t.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_space_weather_table_summary((*C.SidereonSpaceWeatherTable)(pointer), &output)
		})
	})
	runtime.KeepAlive(t)
	if err != nil {
		return SpaceWeatherTableSummary{}, err
	}
	days, err := checkedNativeCount(uint64(output.day_count))
	if err != nil {
		return SpaceWeatherTableSummary{}, err
	}
	monthly, err := checkedNativeCount(uint64(output.monthly_count))
	if err != nil {
		return SpaceWeatherTableSummary{}, err
	}
	skips, err := checkedNativeCount(uint64(output.skip_count))
	if err != nil {
		return SpaceWeatherTableSummary{}, err
	}
	warnings, err := checkedNativeCount(uint64(output.warning_count))
	if err != nil {
		return SpaceWeatherTableSummary{}, err
	}
	return SpaceWeatherTableSummary{DayCount: days, MonthlyCount: monthly, SkipCount: skips, WarningCount: warnings}, nil
}

func (t *SpaceWeatherTable) APArrayAt(epochJ2000S float64) ([7]float64, error) {
	if t == nil || t.handle == nil {
		return [7]float64{}, ErrClosed
	}
	var output [7]C.double
	err := t.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_space_weather_table_ap_array_at((*C.SidereonSpaceWeatherTable)(pointer), C.double(epochJ2000S), &output[0])
		})
	})
	runtime.KeepAlive(t)
	var result [7]float64
	for i := range result {
		result[i] = float64(output[i])
	}
	return result, err
}

func (t *SpaceWeatherTable) SampleAt(epochJ2000S float64) (SpaceWeatherSample, error) {
	if t == nil || t.handle == nil {
		return SpaceWeatherSample{}, ErrClosed
	}
	var output C.SidereonSpaceWeatherSample
	err := t.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_space_weather_table_sample_at((*C.SidereonSpaceWeatherTable)(pointer), C.double(epochJ2000S), &output)
		})
	})
	runtime.KeepAlive(t)
	return spaceWeatherSampleFromC(output), err
}

func (t *SpaceWeatherTable) SampleAtWithPolicy(epochJ2000S float64, policy SpaceWeatherPolicy) (SpaceWeatherSample, error) {
	if t == nil || t.handle == nil {
		return SpaceWeatherSample{}, ErrClosed
	}
	cPolicy := C.SidereonSpaceWeatherPolicy{allow_interpolated: C.bool(policy.AllowInterpolated), allow_daily_predicted: C.bool(policy.AllowDailyPredicted), allow_monthly_predicted: C.bool(policy.AllowMonthlyPredicted), require_geomagnetic: C.bool(policy.RequireGeomagnetic)}
	var output C.SidereonSpaceWeatherSample
	err := t.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_space_weather_table_sample_at_with_policy((*C.SidereonSpaceWeatherTable)(pointer), C.double(epochJ2000S), &cPolicy, &output)
		})
	})
	runtime.KeepAlive(t)
	return spaceWeatherSampleFromC(output), err
}

func (t *SpaceWeatherTable) SpaceWeatherAt(epochJ2000S float64) (SpaceWeather, error) {
	if t == nil || t.handle == nil {
		return SpaceWeather{}, ErrClosed
	}
	var output C.SidereonSpaceWeather
	err := t.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_space_weather_table_space_weather_at((*C.SidereonSpaceWeatherTable)(pointer), C.double(epochJ2000S), &output)
		})
	})
	runtime.KeepAlive(t)
	return spaceWeatherFromC(output), err
}

func (t *SpaceWeatherTable) ToCSV() ([]byte, error) {
	if t == nil || t.handle == nil {
		return nil, ErrClosed
	}
	var result []byte
	err := t.handle.with(func(pointer unsafe.Pointer) error {
		var operationErr error
		result, operationErr = copyNativeBytes("space-weather CSV", func(out *C.uint8_t, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_space_weather_table_to_csv((*C.SidereonSpaceWeatherTable)(pointer), out, length, written, required)
		})
		return operationErr
	})
	runtime.KeepAlive(t)
	return result, err
}

func (t *SpaceWeatherTable) ToTXT() ([]byte, error) {
	if t == nil || t.handle == nil {
		return nil, ErrClosed
	}
	var result []byte
	err := t.handle.with(func(pointer unsafe.Pointer) error {
		var operationErr error
		result, operationErr = copyNativeBytes("space-weather TXT", func(out *C.uint8_t, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_space_weather_table_to_txt((*C.SidereonSpaceWeatherTable)(pointer), out, length, written, required)
		})
		return operationErr
	})
	runtime.KeepAlive(t)
	return result, err
}
