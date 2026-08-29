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
	"fmt"
	"runtime"
	"unsafe"
)

type NativeLeapSecondTableInfo struct {
	FirstMJD  int32
	LastMJD   int32
	Entries   int
	SourceLen int
}

type NativeUT1CoverageInfo struct {
	FirstMJD  int32
	LastMJD   int32
	FirstJDTT float64
	LastJDTT  float64
	Entries   int
	SourceLen int
}

type LnavDecoded struct {
	WeekNumber, L2Code, URAIndex, SVHealth, IODC int64
	TGD                                          float64
	TOC                                          int64
	AF0, AF1, AF2                                float64
	IODE                                         int64
	CRS, DeltaN, M0, CUC                         float64
	Eccentricity                                 float64
	CUS, SqrtA                                   float64
	TOE, FitIntervalFlag, AODO                   int64
	CIC, Omega0, CIS, I0, CRC, Omega             float64
	OmegaDot, IDot                               float64
}

type LnavParams struct {
	WeekNumber, L2Code, L2PDataFlag, URAIndex, SVHealth, IODC int64
	TGD                                                       float64
	TOC                                                       int64
	AF0, AF1, AF2                                             float64
	IODE                                                      int64
	CRS, DeltaN, M0, CUC                                      float64
	Eccentricity                                              float64
	CUS, SqrtA                                                float64
	TOE, FitIntervalFlag, AODO                                int64
	CIC, Omega0, CIS, I0, CRC, Omega                          float64
	OmegaDot, IDot                                            float64
}

type LnavOptions struct {
	TOW, Alert, AntiSpoof, Integrity, TLMMessage int64
}

type NequickGRay struct {
	Month                                        uint8
	UTCHours                                     float64
	StationLonDeg, StationLatDeg, StationHeightM float64
	SatelliteLonDeg, SatelliteLatDeg             float64
	SatelliteHeightM                             float64
}

type Ephemeris struct {
	_      noCopy
	handle *surfaceHandle
}

func releaseEphemeris(pointer unsafe.Pointer) {
	C.sidereon_ephemeris_free((*C.SidereonEphemeris)(pointer))
}

func PropagateState(config NativePropagationConfig, times []float64) (*Ephemeris, error) {
	if len(times) == 0 {
		return nil, invalidArgument("propagation time list is empty")
	}
	timesPointer, timesLength, err := cFloats(times, "propagation times")
	if err != nil {
		return nil, err
	}
	defer C.free(timesPointer)
	cConfig := propagationConfig(config)
	var output *C.SidereonEphemeris
	var operationErr error
	withCThread(func() {
		status := C.sidereon_propagate_state(&cConfig, (*C.double)(timesPointer), timesLength, &output)
		operationErr = statusErrorLocked(uint32(status))
		if operationErr != nil && output != nil {
			C.sidereon_ephemeris_free(output)
			output = nil
		}
	})
	runtime.KeepAlive(times)
	if operationErr != nil {
		return nil, operationErr
	}
	if output == nil {
		return nil, missingNativeHandle("state propagation")
	}
	handle, err := newSurfaceHandle(unsafe.Pointer(output), releaseEphemeris)
	if err != nil {
		withCThread(func() { C.sidereon_ephemeris_free(output) })
		return nil, err
	}
	return &Ephemeris{handle: handle}, nil
}

func (e *Ephemeris) Close() error {
	if e == nil || e.handle == nil {
		return nil
	}
	return e.handle.close()
}

func (e *Ephemeris) EpochCount() (int, error) {
	if e == nil || e.handle == nil {
		return 0, ErrClosed
	}
	var output C.size_t
	err := e.handle.read(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			return statusErrorLocked(uint32(C.sidereon_ephemeris_epoch_count((*C.SidereonEphemeris)(pointer), &output)))
		})
	})
	if err != nil {
		return 0, err
	}
	return sizeTToInt(output, "ephemeris epoch count")
}

func (e *Ephemeris) States() ([]CartesianState, error) {
	if e == nil || e.handle == nil {
		return nil, ErrClosed
	}
	var result []CartesianState
	err := e.handle.read(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			var written, required C.size_t
			status := C.sidereon_ephemeris_states((*C.SidereonEphemeris)(pointer), nil, 0, &written, &required)
			if err := statusErrorLocked(uint32(status)); err != nil {
				return err
			}
			n, err := sizeTToInt(required, "ephemeris state count")
			if err != nil {
				return err
			}
			if _, err := writtenToInt(written, 0, "ephemeris state query written count"); err != nil {
				return err
			}
			if _, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonCartesianState{})); err != nil {
				return err
			}
			buffer := make([]C.SidereonCartesianState, n)
			var output *C.SidereonCartesianState
			if n > 0 {
				output = &buffer[0]
			}
			written, required = 0, 0
			status = C.sidereon_ephemeris_states((*C.SidereonEphemeris)(pointer), output, C.size_t(n), &written, &required)
			if err := statusErrorLocked(uint32(status)); err != nil {
				return err
			}
			actual, err := validateTwoPassCounts("ephemeris states", n, n, uint64(written), uint64(required))
			if err != nil {
				return err
			}
			result = make([]CartesianState, actual)
			for i := range result {
				result[i] = cartesianStateFromC(buffer[i])
			}
			return nil
		})
	})
	return result, err
}

func (e *Ephemeris) TimesS() ([]float64, error) {
	if e == nil || e.handle == nil {
		return nil, ErrClosed
	}
	var result []float64
	err := e.handle.read(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			var written, required C.size_t
			status := C.sidereon_ephemeris_times_s((*C.SidereonEphemeris)(pointer), nil, 0, &written, &required)
			if err := statusErrorLocked(uint32(status)); err != nil {
				return err
			}
			n, err := sizeTToInt(required, "ephemeris time count")
			if err != nil {
				return err
			}
			if _, err := writtenToInt(written, 0, "ephemeris time query written count"); err != nil {
				return err
			}
			if _, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.double(0))); err != nil {
				return err
			}
			buffer := make([]C.double, n)
			var output *C.double
			if n > 0 {
				output = &buffer[0]
			}
			written, required = 0, 0
			status = C.sidereon_ephemeris_times_s((*C.SidereonEphemeris)(pointer), output, C.size_t(n), &written, &required)
			if err := statusErrorLocked(uint32(status)); err != nil {
				return err
			}
			actual, err := validateTwoPassCounts("ephemeris times", n, n, uint64(written), uint64(required))
			if err != nil {
				return err
			}
			result = make([]float64, actual)
			for i := range result {
				result[i] = float64(buffer[i])
			}
			return nil
		})
	})
	return result, err
}

func cGnssWeekTow(value NativeGnssWeekTow) C.SidereonGnssWeekTow {
	return C.SidereonGnssWeekTow{system: C.uint32_t(value.System), week: C.uint32_t(value.Week), tow_s: C.double(value.TOWSeconds)}
}

func gnssWeekTowFromC(value C.SidereonGnssWeekTow) (NativeGnssWeekTow, error) {
	if err := validTimeScale(uint32(value.system)); err != nil {
		return NativeGnssWeekTow{}, err
	}
	return NativeGnssWeekTow{System: uint32(value.system), Week: uint32(value.week), TOWSeconds: float64(value.tow_s)}, nil
}

func GNSSWeekEpochJulianDayNumber(system uint32) (int64, bool, error) {
	if err := validTimeScale(system); err != nil {
		return 0, false, err
	}
	var present C.bool
	var output C.int64_t
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_gnss_week_epoch_julian_day_number(C.uint32_t(system), &present, &output))
	})
	return int64(output), bool(present), err
}

func GNSSWeekFromCalendar(system uint32, year, month, day int64) (uint32, bool, error) {
	if err := validTimeScale(system); err != nil {
		return 0, false, err
	}
	var present C.bool
	var output C.uint32_t
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_gnss_week_from_calendar(C.uint32_t(system), C.int64_t(year), C.int64_t(month), C.int64_t(day), &present, &output))
	})
	return uint32(output), bool(present), err
}

func GNSSWeekTowNew(system, week uint32, tow float64) (NativeGnssWeekTow, error) {
	if err := validTimeScale(system); err != nil {
		return NativeGnssWeekTow{}, err
	}
	var output C.SidereonGnssWeekTow
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_gnss_week_tow_new(C.uint32_t(system), C.uint32_t(week), C.double(tow), &output))
	})
	if err != nil {
		return NativeGnssWeekTow{}, err
	}
	return gnssWeekTowFromC(output)
}

func GNSSWeekTowNormalized(value NativeGnssWeekTow) (NativeGnssWeekTow, error) {
	if err := validTimeScale(value.System); err != nil {
		return NativeGnssWeekTow{}, err
	}
	var output C.SidereonGnssWeekTow
	input := cGnssWeekTow(value)
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_gnss_week_tow_normalized(&input, &output))
	})
	if err != nil {
		return NativeGnssWeekTow{}, err
	}
	return gnssWeekTowFromC(output)
}

func GNSSWeekTowUnrolledWeek(value NativeGnssWeekTow, rollovers uint32) (uint32, error) {
	if err := validTimeScale(value.System); err != nil {
		return 0, err
	}
	input := cGnssWeekTow(value)
	var output C.uint32_t
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_gnss_week_tow_unrolled_week(&input, C.uint32_t(rollovers), &output))
	})
	return uint32(output), err
}

func LeapSecondTableInfo() (NativeLeapSecondTableInfo, error) {
	var output C.SidereonLeapSecondTableInfo
	err := callStatus(func() uint32 { return uint32(C.sidereon_leap_second_table_info(&output)) })
	if err != nil {
		return NativeLeapSecondTableInfo{}, err
	}
	entries, err := sizeTToInt(output.entries, "leap-second table entries")
	if err != nil {
		return NativeLeapSecondTableInfo{}, err
	}
	sourceLen, err := sizeTToInt(output.source_len, "leap-second table source length")
	if err != nil {
		return NativeLeapSecondTableInfo{}, err
	}
	return NativeLeapSecondTableInfo{FirstMJD: int32(output.first_mjd), LastMJD: int32(output.last_mjd), Entries: entries, SourceLen: sourceLen}, nil
}

func LeapSecondTableSource() ([]byte, error) {
	var output []byte
	err := withCThreadError(func() error {
		var callErr error
		output, callErr = copyNativeBytesLocked("leap-second table source", func(out *C.uint8_t, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_leap_second_table_source(out, length, written, required)
		})
		return callErr
	})
	return output, err
}

func UT1CoverageCoversJDTT(jdTT float64) (bool, error) {
	var output C.bool
	err := callStatus(func() uint32 { return uint32(C.sidereon_ut1_coverage_covers_jd_tt(C.double(jdTT), &output)) })
	return bool(output), err
}

func UT1CoverageInfo() (NativeUT1CoverageInfo, error) {
	var output C.SidereonUt1CoverageInfo
	err := callStatus(func() uint32 { return uint32(C.sidereon_ut1_coverage_info(&output)) })
	if err != nil {
		return NativeUT1CoverageInfo{}, err
	}
	entries, err := sizeTToInt(output.entries, "UT1 coverage entries")
	if err != nil {
		return NativeUT1CoverageInfo{}, err
	}
	sourceLen, err := sizeTToInt(output.source_len, "UT1 coverage source length")
	if err != nil {
		return NativeUT1CoverageInfo{}, err
	}
	return NativeUT1CoverageInfo{FirstMJD: int32(output.first_mjd), LastMJD: int32(output.last_mjd), FirstJDTT: float64(output.first_jd_tt), LastJDTT: float64(output.last_jd_tt), Entries: entries, SourceLen: sourceLen}, nil
}

func UT1CoverageSource() ([]byte, error) {
	var output []byte
	err := withCThreadError(func() error {
		var callErr error
		output, callErr = copyNativeBytesLocked("UT1 coverage source", func(out *C.uint8_t, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_ut1_coverage_source(out, length, written, required)
		})
		return callErr
	})
	return output, err
}

func cLnavDecoded(value *C.SidereonLnavDecoded) LnavDecoded {
	return LnavDecoded{WeekNumber: int64(value.week_number), L2Code: int64(value.l2_code), URAIndex: int64(value.ura_index), SVHealth: int64(value.sv_health), IODC: int64(value.iodc), TGD: float64(value.tgd), TOC: int64(value.toc), AF0: float64(value.af0), AF1: float64(value.af1), AF2: float64(value.af2), IODE: int64(value.iode), CRS: float64(value.crs), DeltaN: float64(value.delta_n), M0: float64(value.m0), CUC: float64(value.cuc), Eccentricity: float64(value.eccentricity), CUS: float64(value.cus), SqrtA: float64(value.sqrt_a), TOE: int64(value.toe), FitIntervalFlag: int64(value.fit_interval_flag), AODO: int64(value.aodo), CIC: float64(value.cic), Omega0: float64(value.omega0), CIS: float64(value.cis), I0: float64(value.i0), CRC: float64(value.crc), Omega: float64(value.omega), OmegaDot: float64(value.omega_dot), IDot: float64(value.idot)}
}

func cLnavParams(value LnavParams) C.SidereonLnavParams {
	return C.SidereonLnavParams{week_number: C.int64_t(value.WeekNumber), l2_code: C.int64_t(value.L2Code), l2_p_data_flag: C.int64_t(value.L2PDataFlag), ura_index: C.int64_t(value.URAIndex), sv_health: C.int64_t(value.SVHealth), iodc: C.int64_t(value.IODC), tgd: C.double(value.TGD), toc: C.int64_t(value.TOC), af0: C.double(value.AF0), af1: C.double(value.AF1), af2: C.double(value.AF2), iode: C.int64_t(value.IODE), crs: C.double(value.CRS), delta_n: C.double(value.DeltaN), m0: C.double(value.M0), cuc: C.double(value.CUC), eccentricity: C.double(value.Eccentricity), cus: C.double(value.CUS), sqrt_a: C.double(value.SqrtA), toe: C.int64_t(value.TOE), fit_interval_flag: C.int64_t(value.FitIntervalFlag), aodo: C.int64_t(value.AODO), cic: C.double(value.CIC), omega0: C.double(value.Omega0), cis: C.double(value.CIS), i0: C.double(value.I0), crc: C.double(value.CRC), omega: C.double(value.Omega), omega_dot: C.double(value.OmegaDot), idot: C.double(value.IDot)}
}

func cLnavOptions(value LnavOptions) C.SidereonLnavOptions {
	return C.SidereonLnavOptions{tow: C.int64_t(value.TOW), alert: C.int64_t(value.Alert), anti_spoof: C.int64_t(value.AntiSpoof), integrity: C.int64_t(value.Integrity), tlm_message: C.int64_t(value.TLMMessage)}
}

func lnavInput(value []byte, label string) (unsafe.Pointer, C.size_t, error) {
	length, err := checkedNativeSize(len(value))
	if err != nil {
		return nil, 0, fmt.Errorf("sidereon: %s: %w", label, err)
	}
	if len(value) == 0 {
		return nil, length, nil
	}
	pointer := C.CBytes(value)
	if pointer == nil {
		return nil, 0, errors.New("sidereon: unable to allocate native LNAV input")
	}
	return pointer, length, nil
}

func requireBits(value []byte, allowed ...int) error {
	for _, n := range allowed {
		if len(value) == n {
			return nil
		}
	}
	return invalidArgument("LNAV bit array has an invalid length")
}

func LnavDecode(sf1, sf2, sf3 []byte) (LnavDecoded, error) {
	if err := requireBits(sf1, 300); err != nil {
		return LnavDecoded{}, err
	}
	if err := requireBits(sf2, 300); err != nil {
		return LnavDecoded{}, err
	}
	if err := requireBits(sf3, 300); err != nil {
		return LnavDecoded{}, err
	}
	if err := validateBits(sf1, "LNAV subframe 1"); err != nil {
		return LnavDecoded{}, err
	}
	if err := validateBits(sf2, "LNAV subframe 2"); err != nil {
		return LnavDecoded{}, err
	}
	if err := validateBits(sf3, "LNAV subframe 3"); err != nil {
		return LnavDecoded{}, err
	}
	var output C.SidereonLnavDecoded
	var err error
	withCThread(func() {
		p1, n1, e := lnavInput(sf1, "LNAV subframe 1")
		if e != nil {
			err = e
			return
		}
		defer C.free(p1)
		p2, n2, e := lnavInput(sf2, "LNAV subframe 2")
		if e != nil {
			err = e
			return
		}
		defer C.free(p2)
		p3, n3, e := lnavInput(sf3, "LNAV subframe 3")
		if e != nil {
			err = e
			return
		}
		defer C.free(p3)
		err = statusErrorLocked(uint32(C.sidereon_lnav_decode((*C.uint8_t)(p1), n1, (*C.uint8_t)(p2), n2, (*C.uint8_t)(p3), n3, &output)))
	})
	runtime.KeepAlive(sf1)
	runtime.KeepAlive(sf2)
	runtime.KeepAlive(sf3)
	if err != nil {
		return LnavDecoded{}, err
	}
	return cLnavDecoded(&output), nil
}

func LnavEncode(params LnavParams, options LnavOptions) (sf1, sf2, sf3 []byte, err error) {
	var output1, output2, output3 unsafe.Pointer
	withCThread(func() {
		p := cLnavParams(params)
		o := cLnavOptions(options)
		output1 = C.malloc(C.size_t(300))
		output2 = C.malloc(C.size_t(300))
		output3 = C.malloc(C.size_t(300))
		if output1 == nil || output2 == nil || output3 == nil {
			err = errors.New("sidereon: unable to allocate native LNAV output")
			return
		}
		err = statusErrorLocked(uint32(C.sidereon_lnav_encode(&p, &o, (*C.uint8_t)(output1), (*C.uint8_t)(output2), (*C.uint8_t)(output3), C.size_t(300))))
		if err == nil {
			sf1 = append([]byte(nil), unsafe.Slice((*byte)(output1), 300)...)
			sf2 = append([]byte(nil), unsafe.Slice((*byte)(output2), 300)...)
			sf3 = append([]byte(nil), unsafe.Slice((*byte)(output3), 300)...)
		}
	})
	if output1 != nil {
		C.free(output1)
	}
	if output2 != nil {
		C.free(output2)
	}
	if output3 != nil {
		C.free(output3)
	}
	return sf1, sf2, sf3, err
}

func LnavParity(data24 []byte, d29Prev, d30Prev byte) ([]byte, error) {
	if err := requireBits(data24, 24); err != nil {
		return nil, err
	}
	if d29Prev > 1 || d30Prev > 1 {
		return nil, invalidArgument("LNAV dependency bits must be zero or one")
	}
	var input unsafe.Pointer
	var output unsafe.Pointer
	var err error
	var result []byte
	withCThread(func() {
		input, _, err = lnavInput(data24, "LNAV parity data")
		if err != nil {
			return
		}
		output = C.malloc(C.size_t(6))
		if output == nil {
			err = errors.New("sidereon: unable to allocate native LNAV parity output")
			return
		}
		err = statusErrorLocked(uint32(C.sidereon_lnav_parity((*C.uint8_t)(input), C.size_t(24), C.uint8_t(d29Prev), C.uint8_t(d30Prev), (*C.uint8_t)(output), C.size_t(6))))
		if err == nil {
			bits := unsafe.Slice((*byte)(output), 6)
			err = validateBits(bits, "LNAV parity output")
			if err == nil {
				result = append([]byte(nil), bits...)
			}
		}
	})
	if input != nil {
		C.free(input)
	}
	if output != nil {
		C.free(output)
	}
	return result, err
}

func validateBits(value []byte, label string) error {
	for _, bit := range value {
		if bit > 1 {
			return fmt.Errorf("sidereon: %s contains a non-bit value", label)
		}
	}
	return nil
}

func LnavParityValid(word30 []byte, d29Prev, d30Prev byte) (bool, error) {
	if err := requireBits(word30, 30); err != nil {
		return false, err
	}
	if err := validateBits(word30, "LNAV word"); err != nil {
		return false, err
	}
	if d29Prev > 1 || d30Prev > 1 {
		return false, invalidArgument("LNAV dependency bits must be zero or one")
	}
	var valid C.bool
	var err error
	withCThread(func() {
		input, _, e := lnavInput(word30, "LNAV word")
		if e != nil {
			err = e
			return
		}
		defer C.free(input)
		err = statusErrorLocked(uint32(C.sidereon_lnav_parity_valid((*C.uint8_t)(input), C.size_t(30), C.uint8_t(d29Prev), C.uint8_t(d30Prev), &valid)))
	})
	runtime.KeepAlive(word30)
	return bool(valid), err
}

func LnavSubframeID(bits []byte) (uint64, error) { return lnavWordValue(bits, true) }
func LnavTOW(bits []byte) (uint64, error)        { return lnavWordValue(bits, false) }

func lnavWordValue(bits []byte, subframe bool) (uint64, error) {
	if err := requireBits(bits, 30, 300); err != nil {
		return 0, err
	}
	if err := validateBits(bits, "LNAV bits"); err != nil {
		return 0, err
	}
	var output C.uint64_t
	var err error
	withCThread(func() {
		input, length, e := lnavInput(bits, "LNAV bits")
		if e != nil {
			err = e
			return
		}
		defer C.free(input)
		var status C.enum_SidereonStatus
		if subframe {
			status = C.sidereon_lnav_subframe_id((*C.uint8_t)(input), length, &output)
		} else {
			status = C.sidereon_lnav_tow((*C.uint8_t)(input), length, &output)
		}
		err = statusErrorLocked(uint32(status))
	})
	runtime.KeepAlive(bits)
	return uint64(output), err
}

func cNequickGRay(value NequickGRay) C.SidereonNequickGRay {
	return C.SidereonNequickGRay{month: C.uint8_t(value.Month), utc_hours: C.double(value.UTCHours), station_lon_deg: C.double(value.StationLonDeg), station_lat_deg: C.double(value.StationLatDeg), station_height_m: C.double(value.StationHeightM), satellite_lon_deg: C.double(value.SatelliteLonDeg), satellite_lat_deg: C.double(value.SatelliteLatDeg), satellite_height_m: C.double(value.SatelliteHeightM)}
}

func GalileoNequickGNative(ai0, ai1, ai2, latDeg, lonDeg, elDeg, tGalS, dayOfYear, frequencyHz float64) (float64, error) {
	var output C.double
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_galileo_nequick_g_native(C.double(ai0), C.double(ai1), C.double(ai2), C.double(latDeg), C.double(lonDeg), C.double(elDeg), C.double(tGalS), C.double(dayOfYear), C.double(frequencyHz), &output))
	})
	return float64(output), err
}

func NequickGDelayM(ai0, ai1, ai2 float64, ray NequickGRay, frequencyHz float64) (float64, error) {
	input := cNequickGRay(ray)
	var output C.double
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_nequick_g_delay_m(C.double(ai0), C.double(ai1), C.double(ai2), &input, C.double(frequencyHz), &output))
	})
	return float64(output), err
}

func NequickGStecTecu(ai0, ai1, ai2 float64, ray NequickGRay) (float64, error) {
	input := cNequickGRay(ray)
	var output C.double
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_nequick_g_stec_tecu(C.double(ai0), C.double(ai1), C.double(ai2), &input, &output))
	})
	return float64(output), err
}

func KlobucharNative(alpha, beta [4]float64, latDeg, lonDeg, azDeg, elDeg, tGPS, frequencyHz float64) (float64, error) {
	var a, b [4]C.double
	for i := range alpha {
		a[i], b[i] = C.double(alpha[i]), C.double(beta[i])
	}
	var output C.double
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_klobuchar_native(&a[0], &b[0], C.double(latDeg), C.double(lonDeg), C.double(azDeg), C.double(elDeg), C.double(tGPS), C.double(frequencyHz), &output))
	})
	return float64(output), err
}
