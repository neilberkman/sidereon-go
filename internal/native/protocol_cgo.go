//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"runtime"
	"unsafe"
)

var errNegativeIndex = errors.New("sidereon: index must not be negative")
var errTokenTooLong = errors.New("sidereon: satellite token is too long")
var errNilNativeHandle = errors.New("sidereon: native constructor returned a nil handle")

const (
	SBASWireFramed250Value = uint32(C.SIDEREON_SBAS_WIRE_FORM_FRAMED250)
	SBASWireBody226Value   = uint32(C.SIDEREON_SBAS_WIRE_FORM_BODY226)

	SBASMessageDoNotUseValue            = uint32(C.SIDEREON_SBAS_MESSAGE_KIND_DO_NOT_USE)
	SBASMessagePRNMaskValue             = uint32(C.SIDEREON_SBAS_MESSAGE_KIND_PRN_MASK)
	SBASMessageFastCorrectionsValue     = uint32(C.SIDEREON_SBAS_MESSAGE_KIND_FAST_CORRECTIONS)
	SBASMessageIntegrityValue           = uint32(C.SIDEREON_SBAS_MESSAGE_KIND_INTEGRITY)
	SBASMessageFastDegradationValue     = uint32(C.SIDEREON_SBAS_MESSAGE_KIND_FAST_DEGRADATION)
	SBASMessageGeoNavValue              = uint32(C.SIDEREON_SBAS_MESSAGE_KIND_GEO_NAV)
	SBASMessageNetworkTimeValue         = uint32(C.SIDEREON_SBAS_MESSAGE_KIND_NETWORK_TIME)
	SBASMessageGeoAlmanacValue          = uint32(C.SIDEREON_SBAS_MESSAGE_KIND_GEO_ALMANAC)
	SBASMessageIGPMaskValue             = uint32(C.SIDEREON_SBAS_MESSAGE_KIND_IGP_MASK)
	SBASMessageMixedCorrectionsValue    = uint32(C.SIDEREON_SBAS_MESSAGE_KIND_MIXED_CORRECTIONS)
	SBASMessageLongTermCorrectionsValue = uint32(C.SIDEREON_SBAS_MESSAGE_KIND_LONG_TERM_CORRECTIONS)
	SBASMessageIONODelaysValue          = uint32(C.SIDEREON_SBAS_MESSAGE_KIND_IONO_DELAYS)
	SBASMessageUnsupportedValue         = uint32(C.SIDEREON_SBAS_MESSAGE_KIND_UNSUPPORTED)

	RTCMSSROrbitValue              = uint32(C.SIDEREON_RTCM_SSR_KIND_ORBIT)
	RTCMSSRClockValue              = uint32(C.SIDEREON_RTCM_SSR_KIND_CLOCK)
	RTCMSSRCombinedOrbitClockValue = uint32(C.SIDEREON_RTCM_SSR_KIND_COMBINED_ORBIT_CLOCK)
	RTCMSSRCodeBiasValue           = uint32(C.SIDEREON_RTCM_SSR_KIND_CODE_BIAS)
	RTCMSSRPhaseBiasValue          = uint32(C.SIDEREON_RTCM_SSR_KIND_PHASE_BIAS)
	RTCMSSRURAValue                = uint32(C.SIDEREON_RTCM_SSR_KIND_URA)
	RTCMSSRHighRateClockValue      = uint32(C.SIDEREON_RTCM_SSR_KIND_HIGH_RATE_CLOCK)
	RTCMSSRVTECValue               = uint32(C.SIDEREON_RTCM_SSR_KIND_VTEC)

	RINEXClockInstantJulianDateValue = uint32(C.SIDEREON_RINEX_CLOCK_INSTANT_REPRESENTATION_JULIAN_DATE)
	RINEXClockInstantNanosValue      = uint32(C.SIDEREON_RINEX_CLOCK_INSTANT_REPRESENTATION_NANOS)
)

func withInput(data []byte, fn func(*C.uint8_t, C.size_t) uint32) error {
	var err error
	withCThread(func() {
		var pointer unsafe.Pointer
		if len(data) != 0 {
			pointer = C.CBytes(data)
			if pointer == nil {
				err = errors.New("sidereon: unable to allocate native input buffer")
				return
			}
			defer C.free(pointer)
		}
		err = statusErrorLocked(fn((*C.uint8_t)(pointer), C.size_t(len(data))))
	})
	runtime.KeepAlive(data)
	return err
}

func withString(value string, fn func(*C.char) uint32) error {
	if err := rejectEmbeddedNUL(value, "native string"); err != nil {
		return err
	}
	var err error
	withCThread(func() {
		pointer := C.CString(value)
		if pointer == nil {
			err = errors.New("sidereon: unable to allocate native string")
			return
		}
		defer C.free(unsafe.Pointer(pointer))
		err = statusErrorLocked(fn(pointer))
	})
	runtime.KeepAlive(value)
	return err
}

func tokenFromC(value C.SidereonSatelliteToken) string {
	var out []byte
	for index := 0; index < len(value.bytes); index++ {
		if value.bytes[index] == 0 {
			break
		}
		out = append(out, byte(value.bytes[index]))
	}
	return string(out)
}

func weekTowFromC(value C.SidereonGnssWeekTow) NativeGnssWeekTow {
	return NativeGnssWeekTow{System: uint32(value.system), Week: uint32(value.week), TOWSeconds: float64(value.tow_s)}
}

func clockEpochFromC(value C.SidereonClockEpoch) NativeClockEpoch {
	return NativeClockEpoch{
		Scale:          uint32(value.scale),
		Representation: uint32(value.representation),
		JulianWhole:    float64(value.jd_whole),
		JulianFraction: float64(value.jd_fraction),
		NanosHigh:      int64(value.nanos_high),
		NanosLow:       uint64(value.nanos_low),
	}
}

type NativeGnssWeekTow struct {
	System     uint32
	Week       uint32
	TOWSeconds float64
}

type NativeKeplerianElements struct {
	SqrtA    float64
	E        float64
	M0       float64
	DeltaN   float64
	Omega0   float64
	I0       float64
	Omega    float64
	OmegaDot float64
	IDot     float64
	CUC      float64
	CUS      float64
	CRC      float64
	CRS      float64
	CIC      float64
	CIS      float64
	ToeSOW   float64
}

type NativeClockPolynomial struct {
	AF0    float64
	AF1    float64
	AF2    float64
	TocSOW float64
}

type NativeBroadcastGroupDelays struct {
	HasGPSTGD          bool
	GPSTGD             float64
	HasGalileoBGDE5AE1 bool
	GalileoBGDE5AE1    float64
	HasGalileoBGDE5BE1 bool
	GalileoBGDE5BE1    float64
	HasBeiDouTGD1      bool
	BeiDouTGD1         float64
	HasBeiDouTGD2      bool
	BeiDouTGD2         float64
	HasCNAVISCL1CA     bool
	CNAVISCL1CA        float64
	HasCNAVISCL2C      bool
	CNAVISCL2C         float64
	HasCNAVISCL5I5     bool
	CNAVISCL5I5        float64
	HasCNAVISCL5Q5     bool
	CNAVISCL5Q5        float64
	HasCNAVISCL1CD     bool
	CNAVISCL1CD        float64
	HasCNAVISCL1CP     bool
	CNAVISCL1CP        float64
}

type NativeBroadcastCNAV struct {
	Present             bool
	ADOTMPerS           float64
	DeltaN0DotRadPerS2  float64
	Top                 NativeGnssWeekTow
	URAEDIndex          int8
	URANED0Index        int8
	URANED1Index        uint8
	URANED2Index        uint8
	TransmissionTimeSOW float64
	HasFlags            bool
	Flags               uint32
}

type NativeBroadcastRecord struct {
	SatelliteID    string
	Message        uint32
	Issue          uint32
	IssueMessage   uint32
	Week           uint32
	Toe            NativeGnssWeekTow
	Toc            NativeGnssWeekTow
	Elements       NativeKeplerianElements
	Clock          NativeClockPolynomial
	GroupDelays    NativeBroadcastGroupDelays
	CNAV           NativeBroadcastCNAV
	SVHealth       float64
	SVAccuracyM    float64
	HasFitInterval bool
	FitIntervalS   float64
}

func broadcastRecordFromC(value C.SidereonBroadcastRecord) NativeBroadcastRecord {
	return NativeBroadcastRecord{
		SatelliteID:  tokenFromC(value.sat_id),
		Message:      uint32(value.message),
		Issue:        uint32(value.issue),
		IssueMessage: uint32(value.issue_message),
		Week:         uint32(value.week),
		Toe:          weekTowFromC(value.toe),
		Toc:          weekTowFromC(value.toc),
		Elements: NativeKeplerianElements{
			SqrtA: float64(value.elements.sqrt_a), E: float64(value.elements.e), M0: float64(value.elements.m0),
			DeltaN: float64(value.elements.delta_n), Omega0: float64(value.elements.omega0), I0: float64(value.elements.i0),
			Omega: float64(value.elements.omega), OmegaDot: float64(value.elements.omega_dot), IDot: float64(value.elements.idot),
			CUC: float64(value.elements.cuc), CUS: float64(value.elements.cus), CRC: float64(value.elements.crc),
			CRS: float64(value.elements.crs), CIC: float64(value.elements.cic), CIS: float64(value.elements.cis),
			ToeSOW: float64(value.elements.toe_sow),
		},
		Clock: NativeClockPolynomial{AF0: float64(value.clock.af0), AF1: float64(value.clock.af1), AF2: float64(value.clock.af2), TocSOW: float64(value.clock.toc_sow)},
		GroupDelays: NativeBroadcastGroupDelays{
			HasGPSTGD: bool(value.group_delays.has_gps_tgd_s), GPSTGD: float64(value.group_delays.gps_tgd_s),
			HasGalileoBGDE5AE1: bool(value.group_delays.has_galileo_bgd_e5a_e1_s), GalileoBGDE5AE1: float64(value.group_delays.galileo_bgd_e5a_e1_s),
			HasGalileoBGDE5BE1: bool(value.group_delays.has_galileo_bgd_e5b_e1_s), GalileoBGDE5BE1: float64(value.group_delays.galileo_bgd_e5b_e1_s),
			HasBeiDouTGD1: bool(value.group_delays.has_beidou_tgd1_s), BeiDouTGD1: float64(value.group_delays.beidou_tgd1_s),
			HasBeiDouTGD2: bool(value.group_delays.has_beidou_tgd2_s), BeiDouTGD2: float64(value.group_delays.beidou_tgd2_s),
			HasCNAVISCL1CA: bool(value.group_delays.has_cnav_isc_l1ca_s), CNAVISCL1CA: float64(value.group_delays.cnav_isc_l1ca_s),
			HasCNAVISCL2C: bool(value.group_delays.has_cnav_isc_l2c_s), CNAVISCL2C: float64(value.group_delays.cnav_isc_l2c_s),
			HasCNAVISCL5I5: bool(value.group_delays.has_cnav_isc_l5i5_s), CNAVISCL5I5: float64(value.group_delays.cnav_isc_l5i5_s),
			HasCNAVISCL5Q5: bool(value.group_delays.has_cnav_isc_l5q5_s), CNAVISCL5Q5: float64(value.group_delays.cnav_isc_l5q5_s),
			HasCNAVISCL1CD: bool(value.group_delays.has_cnav_isc_l1cd_s), CNAVISCL1CD: float64(value.group_delays.cnav_isc_l1cd_s),
			HasCNAVISCL1CP: bool(value.group_delays.has_cnav_isc_l1cp_s), CNAVISCL1CP: float64(value.group_delays.cnav_isc_l1cp_s),
		},
		CNAV: NativeBroadcastCNAV{
			Present: bool(value.cnav.present), ADOTMPerS: float64(value.cnav.adot_m_s), DeltaN0DotRadPerS2: float64(value.cnav.delta_n0_dot_rad_s2),
			Top:        NativeGnssWeekTow{System: uint32(value.cnav.top.system), Week: uint32(value.cnav.top.week), TOWSeconds: float64(value.cnav.top.tow_s)},
			URAEDIndex: int8(value.cnav.ura_ed_index), URANED0Index: int8(value.cnav.ura_ned0_index), URANED1Index: uint8(value.cnav.ura_ned1_index), URANED2Index: uint8(value.cnav.ura_ned2_index),
			TransmissionTimeSOW: float64(value.cnav.transmission_time_sow), HasFlags: bool(value.cnav.has_flags), Flags: uint32(value.cnav.flags),
		},
		SVHealth: float64(value.sv_health), SVAccuracyM: float64(value.sv_accuracy_m), HasFitInterval: bool(value.has_fit_interval_s), FitIntervalS: float64(value.fit_interval_s),
	}
}

type NativeSkippedNavBlock struct {
	SatelliteID string
	Message     string
}

type NativeIonoCorrections struct {
	GPSPresent     bool
	GPSAlpha       [4]float64
	GPSBeta        [4]float64
	BeiDouPresent  bool
	BeiDouAlpha    [4]float64
	BeiDouBeta     [4]float64
	GalileoPresent bool
	GalileoAI0     float64
	GalileoAI1     float64
	GalileoAI2     float64
}

func copyFloat4(value *[4]C.double) (out [4]float64) {
	for index := range out {
		out[index] = float64(value[index])
	}
	return out
}

func parseIono(data []byte) (NativeIonoCorrections, error) {
	var value C.SidereonIonoCorrections
	err := withInput(data, func(pointer *C.uint8_t, length C.size_t) uint32 {
		return C.sidereon_parse_rinex_iono_corrections(pointer, length, &value)
	})
	return NativeIonoCorrections{
		GPSPresent: bool(value.gps.present), GPSAlpha: copyFloat4(&value.gps.alpha), GPSBeta: copyFloat4(&value.gps.beta),
		BeiDouPresent: bool(value.beidou.present), BeiDouAlpha: copyFloat4(&value.beidou.alpha), BeiDouBeta: copyFloat4(&value.beidou.beta),
		GalileoPresent: bool(value.galileo.present), GalileoAI0: float64(value.galileo.ai0), GalileoAI1: float64(value.galileo.ai1), GalileoAI2: float64(value.galileo.ai2),
	}, err
}

func parseLeapSeconds(data []byte) (float64, bool, error) {
	var seconds C.double
	var present C.bool
	err := withInput(data, func(pointer *C.uint8_t, length C.size_t) uint32 {
		return C.sidereon_parse_rinex_leap_seconds(pointer, length, &seconds, &present)
	})
	return float64(seconds), bool(present), err
}

func ParseRinexIonoCorrections(data []byte) (NativeIonoCorrections, error) {
	return parseIono(data)
}

func ParseRinexLeapSeconds(data []byte) (float64, bool, error) {
	return parseLeapSeconds(data)
}

type RinexNavRecords struct {
	resource *resource
	cleanup  runtime.Cleanup
}

func newRinexNavRecords(pointer *C.SidereonRinexNavRecords) (*RinexNavRecords, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	handle := &RinexNavRecords{resource: &resource{ptr: unsafe.Pointer(pointer), release: func(pointer unsafe.Pointer) {
		C.sidereon_rinex_nav_records_free((*C.SidereonRinexNavRecords)(pointer))
	}}}
	handle.cleanup = runtime.AddCleanup(handle, cleanupResource, handle.resource)
	return handle, nil
}

func ParseRinexNavRecords(data []byte) (*RinexNavRecords, error) {
	var pointer *C.SidereonRinexNavRecords
	err := withInput(data, func(input *C.uint8_t, length C.size_t) uint32 {
		return C.sidereon_parse_rinex_nav_records(input, length, &pointer)
	})
	if err != nil {
		return nil, err
	}
	return newRinexNavRecords(pointer)
}

func (records *RinexNavRecords) Close() error {
	if records == nil {
		return nil
	}
	return closeProtocolResource(records, records.resource, &records.cleanup)
}

func (records *RinexNavRecords) Count() (int, error) {
	if records == nil || records.resource == nil {
		return 0, ErrClosed
	}
	var count C.size_t
	err := records.resource.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_rinex_nav_records_count((*C.SidereonRinexNavRecords)(pointer), &count)
		})
	})
	runtime.KeepAlive(records)
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(count))
}

func (records *RinexNavRecords) Records() ([]NativeBroadcastRecord, error) {
	if records == nil || records.resource == nil {
		return nil, ErrClosed
	}
	var out []NativeBroadcastRecord
	err := records.resource.with(func(pointer unsafe.Pointer) error {
		var count C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_rinex_nav_records_count((*C.SidereonRinexNavRecords)(pointer), &count)
		}); err != nil {
			return err
		}
		n, err := checkedNativeCount(uint64(count))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonBroadcastRecord{})); err != nil {
			return err
		}
		out = make([]NativeBroadcastRecord, n)
		for index := range out {
			var value C.SidereonBroadcastRecord
			if err := callStatus(func() uint32 {
				return C.sidereon_rinex_nav_records_item((*C.SidereonRinexNavRecords)(pointer), C.size_t(index), &value)
			}); err != nil {
				return err
			}
			out[index] = broadcastRecordFromC(value)
		}
		return nil
	})
	runtime.KeepAlive(records)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (records *RinexNavRecords) Record(index int) (NativeBroadcastRecord, error) {
	if index < 0 {
		return NativeBroadcastRecord{}, errNegativeIndex
	}
	if records == nil || records.resource == nil {
		return NativeBroadcastRecord{}, ErrClosed
	}
	var out NativeBroadcastRecord
	err := records.resource.with(func(pointer unsafe.Pointer) error {
		var value C.SidereonBroadcastRecord
		err := callStatus(func() uint32 {
			return C.sidereon_rinex_nav_records_item((*C.SidereonRinexNavRecords)(pointer), C.size_t(index), &value)
		})
		if err == nil {
			out = broadcastRecordFromC(value)
		}
		return err
	})
	runtime.KeepAlive(records)
	return out, err
}

type RinexNavParse struct {
	resource *resource
	cleanup  runtime.Cleanup
}

func newRinexNavParse(pointer *C.SidereonRinexNavParse) (*RinexNavParse, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	handle := &RinexNavParse{resource: &resource{ptr: unsafe.Pointer(pointer), release: func(pointer unsafe.Pointer) {
		C.sidereon_nav_parse_free((*C.SidereonRinexNavParse)(pointer))
	}}}
	handle.cleanup = runtime.AddCleanup(handle, cleanupResource, handle.resource)
	return handle, nil
}

func ParseRinexNavLenient(data []byte) (*RinexNavParse, error) {
	var pointer *C.SidereonRinexNavParse
	err := withInput(data, func(input *C.uint8_t, length C.size_t) uint32 {
		return C.sidereon_parse_rinex_nav_lenient(input, length, &pointer)
	})
	if err != nil {
		return nil, err
	}
	return newRinexNavParse(pointer)
}

func (parse *RinexNavParse) Close() error {
	if parse == nil {
		return nil
	}
	return closeProtocolResource(parse, parse.resource, &parse.cleanup)
}

func (parse *RinexNavParse) RecordCount() (int, error) {
	if parse == nil || parse.resource == nil {
		return 0, ErrClosed
	}
	var count C.size_t
	err := parse.resource.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_nav_parse_record_count((*C.SidereonRinexNavParse)(pointer), &count)
		})
	})
	runtime.KeepAlive(parse)
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(count))
}

func (parse *RinexNavParse) SkippedCount() (int, error) {
	if parse == nil || parse.resource == nil {
		return 0, ErrClosed
	}
	var count C.size_t
	err := parse.resource.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_nav_parse_skipped_count((*C.SidereonRinexNavParse)(pointer), &count)
		})
	})
	runtime.KeepAlive(parse)
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(count))
}

func (parse *RinexNavParse) Records() ([]NativeBroadcastRecord, error) {
	if parse == nil || parse.resource == nil {
		return nil, ErrClosed
	}
	var out []NativeBroadcastRecord
	err := parse.resource.with(func(pointer unsafe.Pointer) error {
		var count C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_nav_parse_record_count((*C.SidereonRinexNavParse)(pointer), &count)
		}); err != nil {
			return err
		}
		n, err := checkedNativeCount(uint64(count))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonBroadcastRecord{})); err != nil {
			return err
		}
		out = make([]NativeBroadcastRecord, n)
		for index := range out {
			var value C.SidereonBroadcastRecord
			if err := callStatus(func() uint32 {
				return C.sidereon_nav_parse_record((*C.SidereonRinexNavParse)(pointer), C.size_t(index), &value)
			}); err != nil {
				return err
			}
			out[index] = broadcastRecordFromC(value)
		}
		return nil
	})
	runtime.KeepAlive(parse)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (parse *RinexNavParse) Skipped() ([]NativeSkippedNavBlock, error) {
	if parse == nil || parse.resource == nil {
		return nil, ErrClosed
	}
	var out []NativeSkippedNavBlock
	err := parse.resource.with(func(pointer unsafe.Pointer) error {
		var count C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_nav_parse_skipped_count((*C.SidereonRinexNavParse)(pointer), &count)
		}); err != nil {
			return err
		}
		n, err := checkedNativeCount(uint64(count))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonSkippedNavBlock{})); err != nil {
			return err
		}
		out = make([]NativeSkippedNavBlock, n)
		for index := range out {
			var value C.SidereonSkippedNavBlock
			if err := callStatus(func() uint32 {
				return C.sidereon_nav_parse_skipped((*C.SidereonRinexNavParse)(pointer), C.size_t(index), &value)
			}); err != nil {
				return err
			}
			var written, required C.size_t
			if err := callStatus(func() uint32 {
				return C.sidereon_nav_parse_skipped_message((*C.SidereonRinexNavParse)(pointer), C.size_t(index), nil, 0, &written, &required)
			}); err != nil {
				return err
			}
			n, err := checkedNativeCount(uint64(required))
			if err != nil {
				return err
			}
			message := make([]C.uint8_t, n)
			var output *C.uint8_t
			if len(message) != 0 {
				output = &message[0]
			}
			if err := callStatus(func() uint32 {
				return C.sidereon_nav_parse_skipped_message((*C.SidereonRinexNavParse)(pointer), C.size_t(index), output, C.size_t(len(message)), &written, &required)
			}); err != nil {
				return err
			}
			writtenInt, err := validateNativeOutput("RINEX NAV skipped diagnostic", len(message), uint64(written), uint64(required))
			if err != nil {
				return err
			}
			text := make([]byte, writtenInt)
			for position := range text {
				text[position] = byte(message[position])
			}
			out[index] = NativeSkippedNavBlock{SatelliteID: tokenFromC(value.satellite), Message: string(text)}
		}
		return nil
	})
	runtime.KeepAlive(parse)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func tokenToC(value string) (out C.SidereonSatelliteToken, err error) {
	if err := rejectEmbeddedNUL(value, "satellite token"); err != nil {
		return out, err
	}
	if len(value) > len(out.bytes)-1 {
		return out, errTokenTooLong
	}
	for index := 0; index < len(value); index++ {
		out.bytes[index] = C.char(value[index])
	}
	return out, nil
}

func weekTowToC(value NativeGnssWeekTow) C.SidereonGnssWeekTow {
	return C.SidereonGnssWeekTow{system: C.uint32_t(value.System), week: C.uint32_t(value.Week), tow_s: C.double(value.TOWSeconds)}
}

func broadcastRecordToC(value NativeBroadcastRecord) (C.SidereonBroadcastRecord, error) {
	satelliteID, err := tokenToC(value.SatelliteID)
	if err != nil {
		return C.SidereonBroadcastRecord{}, err
	}
	return C.SidereonBroadcastRecord{
		sat_id: satelliteID, message: C.uint32_t(value.Message), issue: C.uint32_t(value.Issue), issue_message: C.uint32_t(value.IssueMessage), week: C.uint32_t(value.Week),
		toe: weekTowToC(value.Toe), toc: weekTowToC(value.Toc),
		elements: C.SidereonKeplerianElements{
			sqrt_a: C.double(value.Elements.SqrtA), e: C.double(value.Elements.E), m0: C.double(value.Elements.M0), delta_n: C.double(value.Elements.DeltaN),
			omega0: C.double(value.Elements.Omega0), i0: C.double(value.Elements.I0), omega: C.double(value.Elements.Omega), omega_dot: C.double(value.Elements.OmegaDot), idot: C.double(value.Elements.IDot),
			cuc: C.double(value.Elements.CUC), cus: C.double(value.Elements.CUS), crc: C.double(value.Elements.CRC), crs: C.double(value.Elements.CRS), cic: C.double(value.Elements.CIC), cis: C.double(value.Elements.CIS), toe_sow: C.double(value.Elements.ToeSOW),
		},
		clock: C.SidereonClockPolynomial{af0: C.double(value.Clock.AF0), af1: C.double(value.Clock.AF1), af2: C.double(value.Clock.AF2), toc_sow: C.double(value.Clock.TocSOW)},
		group_delays: C.SidereonBroadcastGroupDelays{
			has_gps_tgd_s: C.bool(value.GroupDelays.HasGPSTGD), gps_tgd_s: C.double(value.GroupDelays.GPSTGD),
			has_galileo_bgd_e5a_e1_s: C.bool(value.GroupDelays.HasGalileoBGDE5AE1), galileo_bgd_e5a_e1_s: C.double(value.GroupDelays.GalileoBGDE5AE1),
			has_galileo_bgd_e5b_e1_s: C.bool(value.GroupDelays.HasGalileoBGDE5BE1), galileo_bgd_e5b_e1_s: C.double(value.GroupDelays.GalileoBGDE5BE1),
			has_beidou_tgd1_s: C.bool(value.GroupDelays.HasBeiDouTGD1), beidou_tgd1_s: C.double(value.GroupDelays.BeiDouTGD1), has_beidou_tgd2_s: C.bool(value.GroupDelays.HasBeiDouTGD2), beidou_tgd2_s: C.double(value.GroupDelays.BeiDouTGD2),
			has_cnav_isc_l1ca_s: C.bool(value.GroupDelays.HasCNAVISCL1CA), cnav_isc_l1ca_s: C.double(value.GroupDelays.CNAVISCL1CA), has_cnav_isc_l2c_s: C.bool(value.GroupDelays.HasCNAVISCL2C), cnav_isc_l2c_s: C.double(value.GroupDelays.CNAVISCL2C),
			has_cnav_isc_l5i5_s: C.bool(value.GroupDelays.HasCNAVISCL5I5), cnav_isc_l5i5_s: C.double(value.GroupDelays.CNAVISCL5I5), has_cnav_isc_l5q5_s: C.bool(value.GroupDelays.HasCNAVISCL5Q5), cnav_isc_l5q5_s: C.double(value.GroupDelays.CNAVISCL5Q5),
			has_cnav_isc_l1cd_s: C.bool(value.GroupDelays.HasCNAVISCL1CD), cnav_isc_l1cd_s: C.double(value.GroupDelays.CNAVISCL1CD), has_cnav_isc_l1cp_s: C.bool(value.GroupDelays.HasCNAVISCL1CP), cnav_isc_l1cp_s: C.double(value.GroupDelays.CNAVISCL1CP),
		},
		cnav: C.SidereonBroadcastCnavParameters{
			present: C.bool(value.CNAV.Present), adot_m_s: C.double(value.CNAV.ADOTMPerS), delta_n0_dot_rad_s2: C.double(value.CNAV.DeltaN0DotRadPerS2), top: weekTowToC(value.CNAV.Top),
			ura_ed_index: C.int8_t(value.CNAV.URAEDIndex), ura_ned0_index: C.int8_t(value.CNAV.URANED0Index), ura_ned1_index: C.uint8_t(value.CNAV.URANED1Index), ura_ned2_index: C.uint8_t(value.CNAV.URANED2Index),
			transmission_time_sow: C.double(value.CNAV.TransmissionTimeSOW), has_flags: C.bool(value.CNAV.HasFlags), flags: C.uint32_t(value.CNAV.Flags),
		},
		sv_health: C.double(value.SVHealth), sv_accuracy_m: C.double(value.SVAccuracyM), has_fit_interval_s: C.bool(value.HasFitInterval), fit_interval_s: C.double(value.FitIntervalS),
	}, nil
}

func EncodeRinexNav(records []NativeBroadcastRecord) ([]byte, error) {
	if _, err := checkedNativeAllocationSize(len(records), unsafe.Sizeof(C.SidereonBroadcastRecord{})); err != nil {
		return nil, err
	}
	values := make([]C.SidereonBroadcastRecord, len(records))
	for index, record := range records {
		value, err := broadcastRecordToC(record)
		if err != nil {
			return nil, err
		}
		values[index] = value
	}
	var out []byte
	var result error
	withCThread(func() {
		var input *C.SidereonBroadcastRecord
		if len(values) != 0 {
			input = &values[0]
		}
		var written, required C.size_t
		result = statusErrorLocked(C.sidereon_encode_rinex_nav(input, C.size_t(len(values)), nil, 0, &written, &required))
		if result != nil {
			return
		}
		n, err := checkedNativeCount(uint64(required))
		if err != nil {
			result = err
			return
		}
		buffer := make([]C.uint8_t, n)
		var output *C.uint8_t
		if len(buffer) != 0 {
			output = &buffer[0]
		}
		result = statusErrorLocked(C.sidereon_encode_rinex_nav(input, C.size_t(len(values)), output, C.size_t(len(buffer)), &written, &required))
		if result != nil {
			return
		}
		writtenInt, err := validateNativeOutput("RINEX NAV encoding", len(buffer), uint64(written), uint64(required))
		if err != nil {
			result = err
			return
		}
		out = make([]byte, writtenInt)
		for index := range out {
			out[index] = byte(buffer[index])
		}
	})
	runtime.KeepAlive(records)
	return out, result
}

type NativeClockEpoch struct {
	Scale          uint32
	Representation uint32
	JulianWhole    float64
	JulianFraction float64
	NanosHigh      int64
	NanosLow       uint64
}

type NativeClockPoint struct {
	Epoch NativeClockEpoch
	BiasS float64
}

type RinexClock struct {
	resource *resource
	cleanup  runtime.Cleanup
}

type ClockSeries struct {
	resource *resource
	cleanup  runtime.Cleanup
}

func newRinexClock(pointer *C.SidereonRinexClock) (*RinexClock, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	handle := &RinexClock{resource: &resource{ptr: unsafe.Pointer(pointer), release: func(pointer unsafe.Pointer) {
		C.sidereon_rinex_clock_free((*C.SidereonRinexClock)(pointer))
	}}}
	handle.cleanup = runtime.AddCleanup(handle, cleanupResource, handle.resource)
	return handle, nil
}

func newClockSeries(pointer *C.SidereonClockSeries) (*ClockSeries, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	handle := &ClockSeries{resource: &resource{ptr: unsafe.Pointer(pointer), release: func(pointer unsafe.Pointer) {
		C.sidereon_rinex_clock_series_free((*C.SidereonClockSeries)(pointer))
	}}}
	handle.cleanup = runtime.AddCleanup(handle, cleanupResource, handle.resource)
	return handle, nil
}

func ParseRinexClock(data []byte, lossy bool) (*RinexClock, error) {
	var pointer *C.SidereonRinexClock
	err := withInput(data, func(input *C.uint8_t, length C.size_t) uint32 {
		if lossy {
			return C.sidereon_rinex_clock_parse_lossy(input, length, &pointer)
		}
		return C.sidereon_rinex_clock_parse(input, length, &pointer)
	})
	if err != nil {
		return nil, err
	}
	return newRinexClock(pointer)
}

func (clock *RinexClock) Close() error {
	if clock == nil {
		return nil
	}
	return closeProtocolResource(clock, clock.resource, &clock.cleanup)
}

func (clock *RinexClock) SatelliteCount() (int, error) {
	if clock == nil || clock.resource == nil {
		return 0, ErrClosed
	}
	var count C.size_t
	err := clock.resource.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_rinex_clock_satellite_count((*C.SidereonRinexClock)(pointer), &count)
		})
	})
	runtime.KeepAlive(clock)
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(count))
}

func (clock *RinexClock) SeriesCount() (int, error) {
	if clock == nil || clock.resource == nil {
		return 0, ErrClosed
	}
	var count C.size_t
	err := clock.resource.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_rinex_clock_series_count((*C.SidereonRinexClock)(pointer), &count)
		})
	})
	runtime.KeepAlive(clock)
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(count))
}

func (clock *RinexClock) SampleCount() (int, error) {
	if clock == nil || clock.resource == nil {
		return 0, ErrClosed
	}
	var count C.size_t
	err := clock.resource.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_rinex_clock_sample_count((*C.SidereonRinexClock)(pointer), &count)
		})
	})
	runtime.KeepAlive(clock)
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(count))
}

func (clock *RinexClock) Satellites() ([]string, error) {
	if clock == nil || clock.resource == nil {
		return nil, ErrClosed
	}
	var out []string
	err := clock.resource.with(func(pointer unsafe.Pointer) error {
		var count C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_rinex_clock_satellite_count((*C.SidereonRinexClock)(pointer), &count)
		}); err != nil {
			return err
		}
		n, err := checkedNativeCount(uint64(count))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonSatelliteToken{})); err != nil {
			return err
		}
		values := make([]C.SidereonSatelliteToken, n)
		var output *C.SidereonSatelliteToken
		if len(values) != 0 {
			output = &values[0]
		}
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_rinex_clock_satellites((*C.SidereonRinexClock)(pointer), output, C.size_t(len(values)), &written, &required)
		}); err != nil {
			return err
		}
		writtenInt, err := validateNativeOutput("RINEX clock satellites", len(values), uint64(written), uint64(required))
		if err != nil {
			return err
		}
		out = make([]string, writtenInt)
		for index := range out {
			out[index] = tokenFromC(values[index])
		}
		return nil
	})
	runtime.KeepAlive(clock)
	return out, err
}

func (clock *RinexClock) Series(index int) (*ClockSeries, error) {
	if index < 0 {
		return nil, errNegativeIndex
	}
	if clock == nil || clock.resource == nil {
		return nil, ErrClosed
	}
	var pointer *C.SidereonClockSeries
	err := clock.resource.with(func(owner unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_rinex_clock_series((*C.SidereonRinexClock)(owner), C.size_t(index), &pointer)
		})
	})
	runtime.KeepAlive(clock)
	if err != nil {
		return nil, err
	}
	return newClockSeries(pointer)
}

func (clock *RinexClock) SeriesFor(satelliteID string) (*ClockSeries, error) {
	if clock == nil || clock.resource == nil {
		return nil, ErrClosed
	}
	var pointer *C.SidereonClockSeries
	err := clock.resource.with(func(owner unsafe.Pointer) error {
		return withString(satelliteID, func(value *C.char) uint32 {
			return C.sidereon_rinex_clock_series_for((*C.SidereonRinexClock)(owner), value, &pointer)
		})
	})
	runtime.KeepAlive(clock)
	if err != nil {
		return nil, err
	}
	if pointer == nil {
		return nil, nil
	}
	return newClockSeries(pointer)
}

func (clock *RinexClock) BiasAtGPSSeconds(satelliteID string, seconds float64) (float64, bool, error) {
	if clock == nil || clock.resource == nil {
		return 0, false, ErrClosed
	}
	var bias C.double
	var available C.bool
	err := clock.resource.with(func(owner unsafe.Pointer) error {
		return withString(satelliteID, func(value *C.char) uint32 {
			return C.sidereon_rinex_clock_bias_at_gps_seconds((*C.SidereonRinexClock)(owner), value, C.double(seconds), &bias, &available)
		})
	})
	runtime.KeepAlive(clock)
	return float64(bias), bool(available), err
}

func (clock *RinexClock) Text() ([]byte, error) {
	if clock == nil || clock.resource == nil {
		return nil, ErrClosed
	}
	var out []byte
	err := clock.resource.with(func(pointer unsafe.Pointer) error {
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_rinex_clock_to_text((*C.SidereonRinexClock)(pointer), nil, 0, &written, &required)
		}); err != nil {
			return err
		}
		n, err := checkedNativeCount(uint64(required))
		if err != nil {
			return err
		}
		values := make([]C.uint8_t, n)
		var output *C.uint8_t
		if len(values) != 0 {
			output = &values[0]
		}
		if err := callStatus(func() uint32 {
			return C.sidereon_rinex_clock_to_text((*C.SidereonRinexClock)(pointer), output, C.size_t(len(values)), &written, &required)
		}); err != nil {
			return err
		}
		writtenInt, err := validateNativeOutput("RINEX clock text", len(values), uint64(written), uint64(required))
		if err != nil {
			return err
		}
		out = make([]byte, writtenInt)
		for index := range out {
			out[index] = byte(values[index])
		}
		return nil
	})
	runtime.KeepAlive(clock)
	return out, err
}

func (series *ClockSeries) Close() error {
	if series == nil {
		return nil
	}
	return closeProtocolResource(series, series.resource, &series.cleanup)
}

func (series *ClockSeries) SampleCount() (int, error) {
	if series == nil || series.resource == nil {
		return 0, ErrClosed
	}
	var count C.size_t
	err := series.resource.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_rinex_clock_series_sample_count((*C.SidereonClockSeries)(pointer), &count)
		})
	})
	runtime.KeepAlive(series)
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(count))
}

func (series *ClockSeries) Satellite() (string, error) {
	if series == nil || series.resource == nil {
		return "", ErrClosed
	}
	var value C.SidereonSatelliteToken
	err := series.resource.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_rinex_clock_series_satellite((*C.SidereonClockSeries)(pointer), &value)
		})
	})
	runtime.KeepAlive(series)
	return tokenFromC(value), err
}

func (series *ClockSeries) Samples() ([]NativeClockPoint, error) {
	if series == nil || series.resource == nil {
		return nil, ErrClosed
	}
	var out []NativeClockPoint
	err := series.resource.with(func(pointer unsafe.Pointer) error {
		var count C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_rinex_clock_series_sample_count((*C.SidereonClockSeries)(pointer), &count)
		}); err != nil {
			return err
		}
		n, err := checkedNativeCount(uint64(count))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonClockPoint{})); err != nil {
			return err
		}
		values := make([]C.SidereonClockPoint, n)
		var output *C.SidereonClockPoint
		if len(values) != 0 {
			output = &values[0]
		}
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_rinex_clock_series_samples((*C.SidereonClockSeries)(pointer), output, C.size_t(len(values)), &written, &required)
		}); err != nil {
			return err
		}
		writtenInt, err := validateNativeOutput("RINEX clock samples", len(values), uint64(written), uint64(required))
		if err != nil {
			return err
		}
		out = make([]NativeClockPoint, writtenInt)
		for index := range out {
			out[index] = NativeClockPoint{Epoch: clockEpochFromC(values[index].epoch), BiasS: float64(values[index].bias_s)}
		}
		return nil
	})
	runtime.KeepAlive(series)
	return out, err
}

func DecodeSbasBlock(data []byte, form uint32) (*SbasBlock, error) {
	var pointer *C.SidereonSbasBlock
	err := withInput(data, func(input *C.uint8_t, length C.size_t) uint32 {
		return C.sidereon_sbas_block_decode(input, length, C.uint32_t(form), &pointer)
	})
	if err != nil {
		return nil, err
	}
	return newSbasBlock(pointer)
}

type SbasBlock struct {
	resource *resource
	cleanup  runtime.Cleanup
}

func newSbasBlock(pointer *C.SidereonSbasBlock) (*SbasBlock, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	handle := &SbasBlock{resource: &resource{ptr: unsafe.Pointer(pointer), release: func(pointer unsafe.Pointer) {
		C.sidereon_sbas_block_free((*C.SidereonSbasBlock)(pointer))
	}}}
	handle.cleanup = runtime.AddCleanup(handle, cleanupResource, handle.resource)
	return handle, nil
}

type NativeSbasMessageInfo struct {
	Form           uint32
	Kind           uint32
	MessageType    uint8
	Preamble       uint8
	FastCount      int
	LongTermCount  int
	IONODelayCount int
}

type NativeSbasRawFastCorrections struct {
	Preamble, MessageType, IODF, IODP uint8
	PRC                               [13]int16
	UDREI                             [13]uint8
}
type NativeSbasFastDegradation struct {
	Preamble, SystemLatencyS, IODP uint8
	AI                             [51]uint8
}
type NativeSbasGeoNavMessage struct {
	Preamble                                 uint8
	TimeOfDayS                               uint16
	URA                                      uint8
	XM, YM, ZM                               int32
	XRateMPerS, YRateMPerS, ZRateMPerS       int32
	XAccelMPerS2, YAccelMPerS2, ZAccelMPerS2 int16
	AGF0S, AGF1SPerS                         int16
}
type NativeSbasIgpMask struct {
	Preamble, BandNumber, IODI uint8
	Mask                       [201]bool
}
type NativeSbasIntegrity struct {
	Preamble uint8
	IODF     [4]uint8
	UDREI    [51]uint8
}
type NativeSbasIgpDelay struct {
	VerticalDelay uint16
	GIVEI         uint8
}
type NativeSbasIonoDelays struct {
	Preamble, BandNumber, BlockID, IODI uint8
	Entries                             [15]NativeSbasIgpDelay
}
type NativeSbasLongTermHalfInfo struct {
	VelocityCode bool
	IODP         uint8
	RecordCount  int
}
type NativeSbasLongTermRecord struct {
	MonitoredIndex, IODE                                                           uint8
	DeltaX, DeltaY, DeltaZ, DeltaXRate, DeltaYRate, DeltaZRate, DeltaAF0, DeltaAF1 int32
	HasTimeOfDayS                                                                  bool
	TimeOfDayS                                                                     uint32
}
type NativeSbasMixedFastCorrections struct {
	IODF, IODP, BlockID uint8
	PRC                 [6]int16
	UDREI               [6]uint8
}
type NativeSbasPrnMask struct {
	Preamble, IODP uint8
	Mask           [210]bool
}

func (block *SbasBlock) Close() error {
	if block == nil {
		return nil
	}
	return closeProtocolResource(block, block.resource, &block.cleanup)
}

func (block *SbasBlock) Info() (NativeSbasMessageInfo, error) {
	if block == nil || block.resource == nil {
		return NativeSbasMessageInfo{}, ErrClosed
	}
	var value C.SidereonSbasMessageInfo
	err := block.resource.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 { return C.sidereon_sbas_block_info((*C.SidereonSbasBlock)(pointer), &value) })
	})
	runtime.KeepAlive(block)
	if err != nil {
		return NativeSbasMessageInfo{}, err
	}
	fastCount, err := checkedNativeCount(uint64(value.fast_count))
	if err != nil {
		return NativeSbasMessageInfo{}, err
	}
	longTermCount, err := checkedNativeCount(uint64(value.long_term_count))
	if err != nil {
		return NativeSbasMessageInfo{}, err
	}
	ionoDelayCount, err := checkedNativeCount(uint64(value.iono_delay_count))
	if err != nil {
		return NativeSbasMessageInfo{}, err
	}
	return NativeSbasMessageInfo{Form: uint32(value.form), Kind: uint32(value.kind), MessageType: uint8(value.message_type), Preamble: uint8(value.preamble), FastCount: fastCount, LongTermCount: longTermCount, IONODelayCount: ionoDelayCount}, nil
}

func (block *SbasBlock) FastCorrections() (bool, NativeSbasRawFastCorrections, error) {
	if block == nil || block.resource == nil {
		return false, NativeSbasRawFastCorrections{}, ErrClosed
	}
	var present C.bool
	var value C.SidereonSbasRawFastCorrections
	err := block.resource.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_sbas_block_fast_corrections((*C.SidereonSbasBlock)(pointer), &present, &value)
		})
	})
	runtime.KeepAlive(block)
	out := NativeSbasRawFastCorrections{Preamble: uint8(value.preamble), MessageType: uint8(value.message_type), IODF: uint8(value.iodf), IODP: uint8(value.iodp)}
	for i := range out.PRC {
		out.PRC[i] = int16(value.prc[i])
		out.UDREI[i] = uint8(value.udrei[i])
	}
	return bool(present), out, err
}

func (block *SbasBlock) FastDegradation() (bool, NativeSbasFastDegradation, error) {
	if block == nil || block.resource == nil {
		return false, NativeSbasFastDegradation{}, ErrClosed
	}
	var present C.bool
	var value C.SidereonSbasFastDegradation
	err := block.resource.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_sbas_block_fast_degradation((*C.SidereonSbasBlock)(pointer), &present, &value)
		})
	})
	runtime.KeepAlive(block)
	out := NativeSbasFastDegradation{Preamble: uint8(value.preamble), SystemLatencyS: uint8(value.system_latency_s), IODP: uint8(value.iodp)}
	for i := range out.AI {
		out.AI[i] = uint8(value.ai[i])
	}
	return bool(present), out, err
}

func (block *SbasBlock) GeoNav() (bool, NativeSbasGeoNavMessage, error) {
	if block == nil || block.resource == nil {
		return false, NativeSbasGeoNavMessage{}, ErrClosed
	}
	var present C.bool
	var value C.SidereonSbasGeoNavMessage
	err := block.resource.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 { return C.sidereon_sbas_block_geo_nav((*C.SidereonSbasBlock)(pointer), &present, &value) })
	})
	runtime.KeepAlive(block)
	return bool(present), NativeSbasGeoNavMessage{Preamble: uint8(value.preamble), TimeOfDayS: uint16(value.time_of_day_s), URA: uint8(value.ura), XM: int32(value.x_m), YM: int32(value.y_m), ZM: int32(value.z_m), XRateMPerS: int32(value.x_rate_m_s), YRateMPerS: int32(value.y_rate_m_s), ZRateMPerS: int32(value.z_rate_m_s), XAccelMPerS2: int16(value.x_accel_m_s2), YAccelMPerS2: int16(value.y_accel_m_s2), ZAccelMPerS2: int16(value.z_accel_m_s2), AGF0S: int16(value.a_gf0_s), AGF1SPerS: int16(value.a_gf1_s_s)}, err
}

func (block *SbasBlock) IGPMask() (bool, NativeSbasIgpMask, error) {
	if block == nil || block.resource == nil {
		return false, NativeSbasIgpMask{}, ErrClosed
	}
	var present C.bool
	var value C.SidereonSbasIgpMask
	err := block.resource.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_sbas_block_igp_mask((*C.SidereonSbasBlock)(pointer), &present, &value)
		})
	})
	runtime.KeepAlive(block)
	out := NativeSbasIgpMask{Preamble: uint8(value.preamble), BandNumber: uint8(value.band_number), IODI: uint8(value.iodi)}
	for i := range out.Mask {
		out.Mask[i] = bool(value.mask[i])
	}
	return bool(present), out, err
}

func (block *SbasBlock) Integrity() (bool, NativeSbasIntegrity, error) {
	if block == nil || block.resource == nil {
		return false, NativeSbasIntegrity{}, ErrClosed
	}
	var present C.bool
	var value C.SidereonSbasIntegrity
	err := block.resource.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_sbas_block_integrity((*C.SidereonSbasBlock)(pointer), &present, &value)
		})
	})
	runtime.KeepAlive(block)
	out := NativeSbasIntegrity{Preamble: uint8(value.preamble)}
	for i := range out.IODF {
		out.IODF[i] = uint8(value.iodf[i])
	}
	for i := range out.UDREI {
		out.UDREI[i] = uint8(value.udrei[i])
	}
	return bool(present), out, err
}

func (block *SbasBlock) IonoDelays() (bool, NativeSbasIonoDelays, error) {
	if block == nil || block.resource == nil {
		return false, NativeSbasIonoDelays{}, ErrClosed
	}
	var present C.bool
	var value C.SidereonSbasIonoDelays
	err := block.resource.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_sbas_block_iono_delays((*C.SidereonSbasBlock)(pointer), &present, &value)
		})
	})
	runtime.KeepAlive(block)
	out := NativeSbasIonoDelays{Preamble: uint8(value.preamble), BandNumber: uint8(value.band_number), BlockID: uint8(value.block_id), IODI: uint8(value.iodi)}
	for i := range out.Entries {
		out.Entries[i] = NativeSbasIgpDelay{VerticalDelay: uint16(value.entries[i].vertical_delay), GIVEI: uint8(value.entries[i].givei)}
	}
	return bool(present), out, err
}

func (block *SbasBlock) MixedFastCorrections() (bool, NativeSbasMixedFastCorrections, error) {
	if block == nil || block.resource == nil {
		return false, NativeSbasMixedFastCorrections{}, ErrClosed
	}
	var present C.bool
	var value C.SidereonSbasMixedFastCorrections
	err := block.resource.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_sbas_block_mixed_fast_corrections((*C.SidereonSbasBlock)(pointer), &present, &value)
		})
	})
	runtime.KeepAlive(block)
	out := NativeSbasMixedFastCorrections{IODF: uint8(value.iodf), IODP: uint8(value.iodp), BlockID: uint8(value.block_id)}
	for i := range out.PRC {
		out.PRC[i] = int16(value.prc[i])
		out.UDREI[i] = uint8(value.udrei[i])
	}
	return bool(present), out, err
}

func (block *SbasBlock) PrnMask() (bool, NativeSbasPrnMask, error) {
	if block == nil || block.resource == nil {
		return false, NativeSbasPrnMask{}, ErrClosed
	}
	var present C.bool
	var value C.SidereonSbasPrnMask
	err := block.resource.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_sbas_block_prn_mask((*C.SidereonSbasBlock)(pointer), &present, &value)
		})
	})
	runtime.KeepAlive(block)
	out := NativeSbasPrnMask{Preamble: uint8(value.preamble), IODP: uint8(value.iodp)}
	for i := range out.Mask {
		out.Mask[i] = bool(value.mask[i])
	}
	return bool(present), out, err
}

func (block *SbasBlock) LongTermHalf(index int) (bool, NativeSbasLongTermHalfInfo, error) {
	if index < 0 {
		return false, NativeSbasLongTermHalfInfo{}, errNegativeIndex
	}
	if block == nil || block.resource == nil {
		return false, NativeSbasLongTermHalfInfo{}, ErrClosed
	}
	var present C.bool
	var value C.SidereonSbasLongTermHalfInfo
	err := block.resource.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_sbas_block_long_term_half_info((*C.SidereonSbasBlock)(pointer), C.size_t(index), &present, &value)
		})
	})
	runtime.KeepAlive(block)
	if err != nil {
		return false, NativeSbasLongTermHalfInfo{}, err
	}
	count, err := checkedNativeCount(uint64(value.record_count))
	if err != nil {
		return false, NativeSbasLongTermHalfInfo{}, err
	}
	return bool(present), NativeSbasLongTermHalfInfo{VelocityCode: bool(value.velocity_code), IODP: uint8(value.iodp), RecordCount: count}, nil
}

func (block *SbasBlock) LongTermRecords(index int) ([]NativeSbasLongTermRecord, error) {
	if index < 0 {
		return nil, errNegativeIndex
	}
	if block == nil || block.resource == nil {
		return nil, ErrClosed
	}
	var out []NativeSbasLongTermRecord
	err := block.resource.with(func(pointer unsafe.Pointer) error {
		var present C.bool
		var info C.SidereonSbasLongTermHalfInfo
		if err := callStatus(func() uint32 {
			return C.sidereon_sbas_block_long_term_half_info((*C.SidereonSbasBlock)(pointer), C.size_t(index), &present, &info)
		}); err != nil {
			return err
		}
		if !bool(present) {
			out = []NativeSbasLongTermRecord{}
			return nil
		}
		n, err := checkedNativeCount(uint64(info.record_count))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonSbasLongTermRecord{})); err != nil {
			return err
		}
		values := make([]C.SidereonSbasLongTermRecord, n)
		var output *C.SidereonSbasLongTermRecord
		if len(values) != 0 {
			output = &values[0]
		}
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_sbas_block_long_term_records((*C.SidereonSbasBlock)(pointer), C.size_t(index), output, C.size_t(len(values)), &written, &required)
		}); err != nil {
			return err
		}
		writtenInt, err := validateNativeOutput("SBAS long-term records", len(values), uint64(written), uint64(required))
		if err != nil {
			return err
		}
		out = make([]NativeSbasLongTermRecord, writtenInt)
		for i := range out {
			value := values[i]
			out[i] = NativeSbasLongTermRecord{MonitoredIndex: uint8(value.monitored_index), IODE: uint8(value.iode), DeltaX: int32(value.delta_x), DeltaY: int32(value.delta_y), DeltaZ: int32(value.delta_z), DeltaXRate: int32(value.delta_x_rate), DeltaYRate: int32(value.delta_y_rate), DeltaZRate: int32(value.delta_z_rate), DeltaAF0: int32(value.delta_a_f0), DeltaAF1: int32(value.delta_a_f1), HasTimeOfDayS: bool(value.has_time_of_day_s), TimeOfDayS: uint32(value.time_of_day_s)}
		}
		return nil
	})
	runtime.KeepAlive(block)
	return out, err
}

func (block *SbasBlock) RawData() ([]byte, error) {
	if block == nil || block.resource == nil {
		return nil, ErrClosed
	}
	var out []byte
	err := block.resource.with(func(pointer unsafe.Pointer) error {
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_sbas_block_raw_data((*C.SidereonSbasBlock)(pointer), nil, 0, &written, &required)
		}); err != nil {
			return err
		}
		n, err := checkedNativeCount(uint64(required))
		if err != nil {
			return err
		}
		values := make([]C.uint8_t, n)
		var output *C.uint8_t
		if len(values) != 0 {
			output = &values[0]
		}
		if err := callStatus(func() uint32 {
			return C.sidereon_sbas_block_raw_data((*C.SidereonSbasBlock)(pointer), output, C.size_t(len(values)), &written, &required)
		}); err != nil {
			return err
		}
		writtenInt, err := validateNativeOutput("SBAS raw data", len(values), uint64(written), uint64(required))
		if err != nil {
			return err
		}
		out = make([]byte, writtenInt)
		for i := range out {
			out[i] = byte(values[i])
		}
		return nil
	})
	runtime.KeepAlive(block)
	return out, err
}

func (block *SbasBlock) Encode() ([]byte, error) {
	if block == nil || block.resource == nil {
		return nil, ErrClosed
	}
	var out []byte
	err := block.resource.with(func(pointer unsafe.Pointer) error {
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_sbas_block_encode((*C.SidereonSbasBlock)(pointer), nil, 0, &written, &required)
		}); err != nil {
			return err
		}
		n, err := checkedNativeCount(uint64(required))
		if err != nil {
			return err
		}
		values := make([]C.uint8_t, n)
		var output *C.uint8_t
		if len(values) != 0 {
			output = &values[0]
		}
		if err := callStatus(func() uint32 {
			return C.sidereon_sbas_block_encode((*C.SidereonSbasBlock)(pointer), output, C.size_t(len(values)), &written, &required)
		}); err != nil {
			return err
		}
		writtenInt, err := validateNativeOutput("SBAS encoding", len(values), uint64(written), uint64(required))
		if err != nil {
			return err
		}
		out = make([]byte, writtenInt)
		for i := range out {
			out[i] = byte(values[i])
		}
		return nil
	})
	runtime.KeepAlive(block)
	return out, err
}

type NativeSbasLogBlock struct {
	SatelliteID string
	Epoch       NativeGnssWeekTow
	Form        uint32
	ByteCount   int
}
type SbasLogBlocks struct {
	resource *resource
	cleanup  runtime.Cleanup
}

func newSbasLogBlocks(pointer *C.SidereonSbasLogBlocks) (*SbasLogBlocks, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	handle := &SbasLogBlocks{resource: &resource{ptr: unsafe.Pointer(pointer), release: func(pointer unsafe.Pointer) { C.sidereon_sbas_log_blocks_free((*C.SidereonSbasLogBlocks)(pointer)) }}}
	handle.cleanup = runtime.AddCleanup(handle, cleanupResource, handle.resource)
	return handle, nil
}

func parseSbasLog(data []byte, rtklib bool) (*SbasLogBlocks, error) {
	var pointer *C.SidereonSbasLogBlocks
	err := withInput(data, func(input *C.uint8_t, length C.size_t) uint32 {
		if rtklib {
			return C.sidereon_parse_sbas_rtklib_lines(input, length, &pointer)
		}
		return C.sidereon_parse_sbas_ems_lines(input, length, &pointer)
	})
	if err != nil {
		return nil, err
	}
	return newSbasLogBlocks(pointer)
}

func ParseSbasEMSLines(data []byte) (*SbasLogBlocks, error) {
	return parseSbasLog(data, false)
}

func ParseSbasRTKLIBLines(data []byte) (*SbasLogBlocks, error) {
	return parseSbasLog(data, true)
}
func (blocks *SbasLogBlocks) Close() error {
	if blocks == nil {
		return nil
	}
	return closeProtocolResource(blocks, blocks.resource, &blocks.cleanup)
}

func (blocks *SbasLogBlocks) Count() (int, error) {
	if blocks == nil || blocks.resource == nil {
		return 0, ErrClosed
	}
	var count C.size_t
	err := blocks.resource.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_sbas_log_blocks_count((*C.SidereonSbasLogBlocks)(pointer), &count)
		})
	})
	runtime.KeepAlive(blocks)
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(count))
}
func (blocks *SbasLogBlocks) Items() ([]NativeSbasLogBlock, error) {
	if blocks == nil || blocks.resource == nil {
		return nil, ErrClosed
	}
	var out []NativeSbasLogBlock
	err := blocks.resource.with(func(pointer unsafe.Pointer) error {
		var count C.size_t
		if err := callStatus(func() uint32 { return C.sidereon_sbas_log_blocks_count((*C.SidereonSbasLogBlocks)(pointer), &count) }); err != nil {
			return err
		}
		n, err := checkedNativeCount(uint64(count))
		if err != nil {
			return err
		}
		out = make([]NativeSbasLogBlock, n)
		for i := range out {
			var value C.SidereonSbasLogBlock
			if err := callStatus(func() uint32 {
				return C.sidereon_sbas_log_blocks_item((*C.SidereonSbasLogBlocks)(pointer), C.size_t(i), &value)
			}); err != nil {
				return err
			}
			byteCount, err := checkedNativeCount(uint64(value.byte_count))
			if err != nil {
				return err
			}
			out[i] = NativeSbasLogBlock{SatelliteID: tokenFromC(value.sat_id), Epoch: weekTowFromC(value.epoch), Form: uint32(value.form), ByteCount: byteCount}
		}
		return nil
	})
	runtime.KeepAlive(blocks)
	return out, err
}
func (blocks *SbasLogBlocks) Bytes(index int) ([]byte, error) {
	if index < 0 {
		return nil, errNegativeIndex
	}
	if blocks == nil || blocks.resource == nil {
		return nil, ErrClosed
	}
	var out []byte
	err := blocks.resource.with(func(pointer unsafe.Pointer) error {
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_sbas_log_blocks_bytes((*C.SidereonSbasLogBlocks)(pointer), C.size_t(index), nil, 0, &written, &required)
		}); err != nil {
			return err
		}
		n, err := checkedNativeCount(uint64(required))
		if err != nil {
			return err
		}
		values := make([]C.uint8_t, n)
		var output *C.uint8_t
		if len(values) != 0 {
			output = &values[0]
		}
		if err := callStatus(func() uint32 {
			return C.sidereon_sbas_log_blocks_bytes((*C.SidereonSbasLogBlocks)(pointer), C.size_t(index), output, C.size_t(len(values)), &written, &required)
		}); err != nil {
			return err
		}
		writtenInt, err := validateNativeOutput("SBAS log bytes", len(values), uint64(written), uint64(required))
		if err != nil {
			return err
		}
		out = make([]byte, writtenInt)
		for i := range out {
			out[i] = byte(values[i])
		}
		return nil
	})
	runtime.KeepAlive(blocks)
	return out, err
}

func SbasPRNToSatelliteID(prn uint16) (string, bool, error) {
	var out string
	var result error
	withCThread(func() {
		var written, required C.size_t
		result = statusErrorLocked(C.sidereon_sbas_prn_to_satellite_id(C.uint16_t(prn), nil, 0, &written, &required))
		if result != nil {
			return
		}
		n, sizeErr := checkedNativeCount(uint64(required))
		if sizeErr != nil {
			result = sizeErr
			return
		}
		values := make([]C.uint8_t, n)
		var output *C.uint8_t
		if len(values) != 0 {
			output = &values[0]
		}
		result = statusErrorLocked(C.sidereon_sbas_prn_to_satellite_id(C.uint16_t(prn), output, C.size_t(len(values)), &written, &required))
		if result != nil {
			return
		}
		n, sizeErr = validateNativeOutput("SBAS PRN mapping", len(values), uint64(written), uint64(required))
		if sizeErr != nil {
			result = sizeErr
			return
		}
		bytes := make([]byte, n)
		for i := range bytes {
			bytes[i] = byte(values[i])
		}
		out = string(bytes)
	})
	return out, out != "", result
}

type NativeSsrClockRecord struct {
	SatelliteID uint8
	C0, C1, C2  int32
}
type NativeSsrCodeBiasSignal struct {
	SignalID uint8
	Bias     int16
}
type NativeSsrCodeBiasRecord struct {
	SatelliteID uint8
	SignalCount int
}
type NativeSsrHeader struct {
	EpochTimeS                                              uint32
	UpdateInterval                                          uint8
	MultipleMessage                                         bool
	IODSSR                                                  uint8
	ProviderID                                              uint16
	SolutionID                                              uint8
	HasSatelliteReferenceDatum, SatelliteReferenceDatum     bool
	HasDispersiveBiasConsistency, DispersiveBiasConsistency bool
	HasMWConsistency, MWConsistency                         bool
	SatelliteCount                                          uint8
}
type NativeSsrInfo struct {
	MessageNumber                                                   uint16
	System                                                          uint32
	Kind                                                            uint32
	Header                                                          NativeSsrHeader
	OrbitCount, ClockCount, URACount, CodeBiasCount, PhaseBiasCount int
}
type NativeSsrOrbitRecord struct {
	SatelliteID                                                                       uint8
	IODE                                                                              uint32
	DeltaRadial, DeltaAlong, DeltaCross, DotDeltaRadial, DotDeltaAlong, DotDeltaCross int32
}
type NativeSsrPhaseBiasSignal struct {
	SignalID, IntegerIndicator, WideLaneIntegerIndicator, DiscontinuityCounter uint8
	Bias                                                                       int32
}
type NativeSsrPhaseBiasRecord struct {
	SatelliteID uint8
	YawAngle    uint16
	YawRate     int8
	SignalCount int
}
type NativeSsrUraRecord struct{ SatelliteID, URAIndex uint8 }

type SsrMessage struct {
	resource *resource
	cleanup  runtime.Cleanup
	body     []byte
}

func newSsrMessage(pointer *C.SidereonSsrMessage, body []byte) (*SsrMessage, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	handle := &SsrMessage{resource: &resource{ptr: unsafe.Pointer(pointer), release: func(pointer unsafe.Pointer) { C.sidereon_ssr_message_free((*C.SidereonSsrMessage)(pointer)) }}, body: append([]byte(nil), body...)}
	handle.cleanup = runtime.AddCleanup(handle, cleanupResource, handle.resource)
	return handle, nil
}
func DecodeSsrMessage(data []byte) (*SsrMessage, error) {
	var pointer *C.SidereonSsrMessage
	err := withInput(data, func(input *C.uint8_t, length C.size_t) uint32 {
		return C.sidereon_ssr_message_decode(input, length, &pointer)
	})
	if err != nil {
		return nil, err
	}
	return newSsrMessage(pointer, data)
}
func (message *SsrMessage) Close() error {
	if message == nil {
		return nil
	}
	return closeProtocolResource(message, message.resource, &message.cleanup)
}
func (message *SsrMessage) Body() ([]byte, error) {
	if message == nil || message.resource == nil {
		return nil, ErrClosed
	}
	var out []byte
	err := message.resource.with(func(unsafe.Pointer) error {
		out = append([]byte(nil), message.body...)
		return nil
	})
	runtime.KeepAlive(message)
	return out, err
}

func (message *SsrMessage) Encode() ([]byte, error) {
	return message.Body()
}

func ssrHeaderFromC(value C.SidereonRtcmSsrHeader) NativeSsrHeader {
	return NativeSsrHeader{EpochTimeS: uint32(value.epoch_time_s), UpdateInterval: uint8(value.update_interval), MultipleMessage: bool(value.multiple_message), IODSSR: uint8(value.iod_ssr), ProviderID: uint16(value.provider_id), SolutionID: uint8(value.solution_id), HasSatelliteReferenceDatum: bool(value.has_satellite_reference_datum), SatelliteReferenceDatum: bool(value.satellite_reference_datum), HasDispersiveBiasConsistency: bool(value.has_dispersive_bias_consistency), DispersiveBiasConsistency: bool(value.dispersive_bias_consistency), HasMWConsistency: bool(value.has_mw_consistency), MWConsistency: bool(value.mw_consistency), SatelliteCount: uint8(value.satellite_count)}
}
func (message *SsrMessage) Info() (NativeSsrInfo, error) {
	if message == nil || message.resource == nil {
		return NativeSsrInfo{}, ErrClosed
	}
	var value C.SidereonRtcmSsrInfo
	err := message.resource.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 { return C.sidereon_ssr_message_info((*C.SidereonSsrMessage)(pointer), &value) })
	})
	runtime.KeepAlive(message)
	if err != nil {
		return NativeSsrInfo{}, err
	}
	orbitCount, err := checkedNativeCount(uint64(value.orbit_count))
	if err != nil {
		return NativeSsrInfo{}, err
	}
	clockCount, err := checkedNativeCount(uint64(value.clock_count))
	if err != nil {
		return NativeSsrInfo{}, err
	}
	uraCount, err := checkedNativeCount(uint64(value.ura_count))
	if err != nil {
		return NativeSsrInfo{}, err
	}
	codeBiasCount, err := checkedNativeCount(uint64(value.code_bias_count))
	if err != nil {
		return NativeSsrInfo{}, err
	}
	phaseBiasCount, err := checkedNativeCount(uint64(value.phase_bias_count))
	if err != nil {
		return NativeSsrInfo{}, err
	}
	return NativeSsrInfo{MessageNumber: uint16(value.message_number), System: uint32(value.system), Kind: uint32(value.kind), Header: ssrHeaderFromC(value.header), OrbitCount: orbitCount, ClockCount: clockCount, URACount: uraCount, CodeBiasCount: codeBiasCount, PhaseBiasCount: phaseBiasCount}, nil
}

func (message *SsrMessage) Orbits() ([]NativeSsrOrbitRecord, error) {
	if message == nil || message.resource == nil {
		return nil, ErrClosed
	}
	info, err := message.Info()
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(info.OrbitCount, unsafe.Sizeof(C.SidereonRtcmSsrOrbitRecord{})); err != nil {
		return nil, err
	}
	values := make([]C.SidereonRtcmSsrOrbitRecord, info.OrbitCount)
	var out []NativeSsrOrbitRecord
	err = message.resource.with(func(pointer unsafe.Pointer) error {
		var output *C.SidereonRtcmSsrOrbitRecord
		if len(values) != 0 {
			output = &values[0]
		}
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_ssr_message_orbits((*C.SidereonSsrMessage)(pointer), output, C.size_t(len(values)), &written, &required)
		}); err != nil {
			return err
		}
		n, err := validateNativeOutput("SSR orbit records", len(values), uint64(written), uint64(required))
		if err != nil {
			return err
		}
		out = make([]NativeSsrOrbitRecord, n)
		for i := range out {
			v := values[i]
			out[i] = NativeSsrOrbitRecord{SatelliteID: uint8(v.satellite_id), IODE: uint32(v.iode), DeltaRadial: int32(v.delta_radial), DeltaAlong: int32(v.delta_along), DeltaCross: int32(v.delta_cross), DotDeltaRadial: int32(v.dot_delta_radial), DotDeltaAlong: int32(v.dot_delta_along), DotDeltaCross: int32(v.dot_delta_cross)}
		}
		return nil
	})
	runtime.KeepAlive(message)
	return out, err
}
func (message *SsrMessage) Clocks() ([]NativeSsrClockRecord, error) {
	if message == nil || message.resource == nil {
		return nil, ErrClosed
	}
	info, err := message.Info()
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(info.ClockCount, unsafe.Sizeof(C.SidereonRtcmSsrClockRecord{})); err != nil {
		return nil, err
	}
	values := make([]C.SidereonRtcmSsrClockRecord, info.ClockCount)
	var out []NativeSsrClockRecord
	err = message.resource.with(func(pointer unsafe.Pointer) error {
		var output *C.SidereonRtcmSsrClockRecord
		if len(values) != 0 {
			output = &values[0]
		}
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_ssr_message_clocks((*C.SidereonSsrMessage)(pointer), output, C.size_t(len(values)), &written, &required)
		}); err != nil {
			return err
		}
		n, err := validateNativeOutput("SSR clock records", len(values), uint64(written), uint64(required))
		if err != nil {
			return err
		}
		out = make([]NativeSsrClockRecord, n)
		for i := range out {
			v := values[i]
			out[i] = NativeSsrClockRecord{SatelliteID: uint8(v.satellite_id), C0: int32(v.c0), C1: int32(v.c1), C2: int32(v.c2)}
		}
		return nil
	})
	runtime.KeepAlive(message)
	return out, err
}
func (message *SsrMessage) CodeBiases() ([]NativeSsrCodeBiasRecord, error) {
	if message == nil || message.resource == nil {
		return nil, ErrClosed
	}
	info, err := message.Info()
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(info.CodeBiasCount, unsafe.Sizeof(C.SidereonRtcmSsrCodeBiasRecord{})); err != nil {
		return nil, err
	}
	values := make([]C.SidereonRtcmSsrCodeBiasRecord, info.CodeBiasCount)
	var out []NativeSsrCodeBiasRecord
	err = message.resource.with(func(pointer unsafe.Pointer) error {
		var output *C.SidereonRtcmSsrCodeBiasRecord
		if len(values) != 0 {
			output = &values[0]
		}
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_ssr_message_code_biases((*C.SidereonSsrMessage)(pointer), output, C.size_t(len(values)), &written, &required)
		}); err != nil {
			return err
		}
		n, err := validateNativeOutput("SSR code-bias records", len(values), uint64(written), uint64(required))
		if err != nil {
			return err
		}
		out = make([]NativeSsrCodeBiasRecord, n)
		for i := range out {
			v := values[i]
			count, err := checkedNativeCount(uint64(v.signal_count))
			if err != nil {
				return err
			}
			out[i] = NativeSsrCodeBiasRecord{SatelliteID: uint8(v.satellite_id), SignalCount: count}
		}
		return nil
	})
	runtime.KeepAlive(message)
	return out, err
}
func (message *SsrMessage) CodeBiasSignals(index int) ([]NativeSsrCodeBiasSignal, error) {
	if index < 0 {
		return nil, errNegativeIndex
	}
	if message == nil || message.resource == nil {
		return nil, ErrClosed
	}
	records, err := message.CodeBiases()
	if err != nil {
		return nil, err
	}
	if index >= len(records) {
		return nil, errors.New("sidereon: code-bias record index out of range")
	}
	if _, err := checkedNativeAllocationSize(records[index].SignalCount, unsafe.Sizeof(C.SidereonRtcmSsrCodeBiasSignal{})); err != nil {
		return nil, err
	}
	values := make([]C.SidereonRtcmSsrCodeBiasSignal, records[index].SignalCount)
	var out []NativeSsrCodeBiasSignal
	err = message.resource.with(func(pointer unsafe.Pointer) error {
		var output *C.SidereonRtcmSsrCodeBiasSignal
		if len(values) != 0 {
			output = &values[0]
		}
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_ssr_message_code_bias_signals((*C.SidereonSsrMessage)(pointer), C.size_t(index), output, C.size_t(len(values)), &written, &required)
		}); err != nil {
			return err
		}
		n, err := validateNativeOutput("SSR code-bias signals", len(values), uint64(written), uint64(required))
		if err != nil {
			return err
		}
		out = make([]NativeSsrCodeBiasSignal, n)
		for i := range out {
			out[i] = NativeSsrCodeBiasSignal{SignalID: uint8(values[i].signal_id), Bias: int16(values[i].bias)}
		}
		return nil
	})
	runtime.KeepAlive(message)
	return out, err
}
func (message *SsrMessage) PhaseBiases() ([]NativeSsrPhaseBiasRecord, error) {
	if message == nil || message.resource == nil {
		return nil, ErrClosed
	}
	info, err := message.Info()
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(info.PhaseBiasCount, unsafe.Sizeof(C.SidereonRtcmSsrPhaseBiasRecord{})); err != nil {
		return nil, err
	}
	values := make([]C.SidereonRtcmSsrPhaseBiasRecord, info.PhaseBiasCount)
	var out []NativeSsrPhaseBiasRecord
	err = message.resource.with(func(pointer unsafe.Pointer) error {
		var output *C.SidereonRtcmSsrPhaseBiasRecord
		if len(values) != 0 {
			output = &values[0]
		}
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_ssr_message_phase_biases((*C.SidereonSsrMessage)(pointer), output, C.size_t(len(values)), &written, &required)
		}); err != nil {
			return err
		}
		n, err := validateNativeOutput("SSR phase-bias records", len(values), uint64(written), uint64(required))
		if err != nil {
			return err
		}
		out = make([]NativeSsrPhaseBiasRecord, n)
		for i := range out {
			v := values[i]
			count, err := checkedNativeCount(uint64(v.signal_count))
			if err != nil {
				return err
			}
			out[i] = NativeSsrPhaseBiasRecord{SatelliteID: uint8(v.satellite_id), YawAngle: uint16(v.yaw_angle), YawRate: int8(v.yaw_rate), SignalCount: count}
		}
		return nil
	})
	runtime.KeepAlive(message)
	return out, err
}
func (message *SsrMessage) PhaseBiasSignals(index int) ([]NativeSsrPhaseBiasSignal, error) {
	if index < 0 {
		return nil, errNegativeIndex
	}
	if message == nil || message.resource == nil {
		return nil, ErrClosed
	}
	records, err := message.PhaseBiases()
	if err != nil {
		return nil, err
	}
	if index >= len(records) {
		return nil, errors.New("sidereon: phase-bias record index out of range")
	}
	if _, err := checkedNativeAllocationSize(records[index].SignalCount, unsafe.Sizeof(C.SidereonRtcmSsrPhaseBiasSignal{})); err != nil {
		return nil, err
	}
	values := make([]C.SidereonRtcmSsrPhaseBiasSignal, records[index].SignalCount)
	var out []NativeSsrPhaseBiasSignal
	err = message.resource.with(func(pointer unsafe.Pointer) error {
		var output *C.SidereonRtcmSsrPhaseBiasSignal
		if len(values) != 0 {
			output = &values[0]
		}
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_ssr_message_phase_bias_signals((*C.SidereonSsrMessage)(pointer), C.size_t(index), output, C.size_t(len(values)), &written, &required)
		}); err != nil {
			return err
		}
		n, err := validateNativeOutput("SSR phase-bias signals", len(values), uint64(written), uint64(required))
		if err != nil {
			return err
		}
		out = make([]NativeSsrPhaseBiasSignal, n)
		for i := range out {
			v := values[i]
			out[i] = NativeSsrPhaseBiasSignal{SignalID: uint8(v.signal_id), IntegerIndicator: uint8(v.integer_indicator), WideLaneIntegerIndicator: uint8(v.wide_lane_integer_indicator), DiscontinuityCounter: uint8(v.discontinuity_counter), Bias: int32(v.bias)}
		}
		return nil
	})
	runtime.KeepAlive(message)
	return out, err
}
func (message *SsrMessage) URA() ([]NativeSsrUraRecord, error) {
	if message == nil || message.resource == nil {
		return nil, ErrClosed
	}
	info, err := message.Info()
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(info.URACount, unsafe.Sizeof(C.SidereonRtcmSsrUraRecord{})); err != nil {
		return nil, err
	}
	values := make([]C.SidereonRtcmSsrUraRecord, info.URACount)
	var out []NativeSsrUraRecord
	err = message.resource.with(func(pointer unsafe.Pointer) error {
		var output *C.SidereonRtcmSsrUraRecord
		if len(values) != 0 {
			output = &values[0]
		}
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_ssr_message_ura((*C.SidereonSsrMessage)(pointer), output, C.size_t(len(values)), &written, &required)
		}); err != nil {
			return err
		}
		n, err := validateNativeOutput("SSR URA records", len(values), uint64(written), uint64(required))
		if err != nil {
			return err
		}
		out = make([]NativeSsrUraRecord, n)
		for i := range out {
			v := values[i]
			out[i] = NativeSsrUraRecord{SatelliteID: uint8(v.satellite_id), URAIndex: uint8(v.ura_index)}
		}
		return nil
	})
	runtime.KeepAlive(message)
	return out, err
}

func cleanupResource(value *resource) { value.close() }
func closeProtocolResource(owner any, value *resource, cleanup *runtime.Cleanup) error {
	if owner == nil || value == nil {
		return nil
	}
	cleanup.Stop()
	value.close()
	return nil
}
