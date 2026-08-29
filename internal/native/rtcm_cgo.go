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

type NativeRTCMMessageInfo struct {
	Kind          uint32
	MessageNumber uint16
}
type NativeRTCMMSMHeader struct {
	ReferenceStationID                           uint16
	EpochTime                                    uint32
	MultipleMessage                              bool
	IODS, Reserved, ClockSteering, ExternalClock uint8
	DivergenceFreeSmoothing                      bool
	SmoothingInterval                            uint8
}
type NativeRTCMMSMInfo struct {
	MessageNumber               uint16
	System                      uint32
	Kind                        uint32
	Header                      NativeRTCMMSMHeader
	SatelliteCount, SignalCount int
}
type NativeRTCMMSMSatellite struct {
	ID, RoughRangeMS       uint8
	RoughRangeMod1         uint16
	HasExtendedInfo        bool
	ExtendedInfo           uint8
	HasRoughPhaseRangeRate bool
	RoughPhaseRangeRateMS  int16
}
type NativeRTCMMSMSignal struct {
	SatelliteID, SignalID           uint8
	FinePseudorange, FinePhaseRange int32
	LockTimeIndicator               uint16
	HalfCycleAmbiguity              bool
	CNR                             uint16
	HasFinePhaseRangeRate           bool
	FinePhaseRangeRate              int16
}
type NativeRTCMStationCoordinates struct {
	MessageNumber, ReferenceStationID                                           uint16
	ITRFRealizationYear                                                         uint8
	GPS, GLONASS, Galileo, ReferenceStation, SingleReceiverOscillator, Reserved bool
	QuarterCycleIndicator                                                       uint8
	ECEFX, ECEFY, ECEFZ                                                         int64
	XM, YM, ZM                                                                  float64
	HasAntennaHeight                                                            bool
	AntennaHeight                                                               uint16
	AntennaHeightM                                                              float64
}
type NativeRTCMAntennaDescriptor struct {
	MessageNumber, ReferenceStationID                                                            uint16
	AntennaSetupID                                                                               uint8
	HasAntennaSerialNumber, HasReceiverType, HasReceiverFirmwareVersion, HasReceiverSerialNumber bool
}
type NativeRTCMFrameSkip struct {
	Offset           int
	HasMessageNumber bool
	MessageNumber    uint16
	Reason           uint32
}
type NativeRTCMCellLLI struct {
	SatelliteID, SignalID, LLI uint8
	HasMinLockTime             bool
	MinLockTimeMS              uint32
}
type NativeRTCMPreviousLock struct {
	HasMinLockTime bool
	MinLockTimeMS  uint32
	ElapsedMS      uint64
}
type NativeRTCMGPSEphemeris struct {
	SatelliteID              uint8
	WeekNumber               uint16
	SVAccuracy, CodeOnL2     uint8
	IDOT                     int32
	IODE                     uint8
	TOC                      uint16
	AF2                      int16
	AF1, AF0                 int32
	IODC                     uint16
	CRS, DeltaN              int32
	M0                       int64
	CUC                      int32
	Eccentricity             uint64
	CUS                      int32
	SqrtA                    uint64
	TOE                      uint16
	CIC                      int32
	Omega0                   int64
	CIS                      int32
	I0                       int64
	CRC                      int32
	Omega                    int64
	OmegaDot                 int32
	TGD                      int16
	SVHealth                 uint8
	L2PDataFlag, FitInterval bool
}
type NativeRTCMQZSSEphemeris struct {
	SatelliteID    uint8
	TOC            uint16
	AF2            int16
	AF1, AF0       int32
	IODE           uint8
	CRS, DeltaN    int32
	M0             int64
	CUC            int32
	Eccentricity   uint64
	CUS            int32
	SqrtA          uint64
	TOE            uint16
	CIC            int32
	Omega0         int64
	CIS            int32
	I0             int64
	CRC            int32
	Omega          int64
	OmegaDot, IDOT int32
	CodesOnL2      uint8
	WeekNumber     uint16
	URA, SVHealth  uint8
	TGD            int16
	IODC           uint16
	FitInterval    bool
}
type NativeRTCMBeidouEphemeris struct {
	SatelliteID  uint8
	WeekNumber   uint16
	SVURAI       uint8
	IDOT         int32
	AODE         uint8
	TOC          uint32
	AF2          int16
	AF1, AF0     int32
	AODC         uint8
	CRS, DeltaN  int32
	M0           int64
	CUC          int32
	Eccentricity uint64
	CUS          int32
	SqrtA        uint64
	TOE          uint32
	CIC          int32
	Omega0       int64
	CIS          int32
	I0           int64
	CRC          int32
	Omega        int64
	OmegaDot     int32
	TGD1, TGD2   int16
	SVHealth     bool
}
type NativeRTCMGalileoFNavEphemeris struct {
	SatelliteID        uint8
	WeekNumber, IodNav uint16
	SISA               uint8
	IDOT               int32
	TOC                uint16
	AF2                int16
	AF1                int32
	AF0                int64
	CRS, DeltaN        int32
	M0                 int64
	CUC                int32
	Eccentricity       uint64
	CUS                int32
	SqrtA              uint64
	TOE                uint16
	CIC                int32
	Omega0             int64
	CIS                int32
	I0                 int64
	CRC                int32
	Omega              int64
	OmegaDot           int32
	BGDE5AE1           int16
	E5ASignalHealth    uint8
	E5ADataValidity    bool
	Reserved           uint8
}
type NativeRTCMGalileoINavEphemeris struct {
	SatelliteID        uint8
	WeekNumber, IodNav uint16
	SISAIndex          uint8
	IDOT               int32
	TOC                uint16
	AF2                int16
	AF1                int32
	AF0                int64
	CRS, DeltaN        int32
	M0                 int64
	CUC                int32
	Eccentricity       uint64
	CUS                int32
	SqrtA              uint64
	TOE                uint16
	CIC                int32
	Omega0             int64
	CIS                int32
	I0                 int64
	CRC                int32
	Omega              int64
	OmegaDot           int32
	BGDE5AE1, BGDE5BE1 int16
	E5BSignalHealth    uint8
	E5BDataValidity    bool
	E1BSignalHealth    uint8
	E1BDataValidity    bool
	Reserved           uint8
}
type NativeRTCMGLONASSEphemeris struct {
	SatelliteID, FrequencyChannel            uint8
	AlmanacHealth, AlmanacHealthAvailability bool
	P1                                       uint8
	TK                                       uint16
	BNMSB, P2                                bool
	TB                                       uint8
	XNDot, XN                                int32
	XNDotDot                                 int8
	YNDot, YN                                int32
	YNDotDot                                 int8
	ZNDot, ZN                                int32
	ZNDotDot                                 int8
	P3                                       bool
	GammaN                                   int16
	MP                                       uint8
	MLNThird                                 bool
	TauN                                     int32
	DeltaTauN                                int8
	EN                                       uint8
	MP4                                      bool
	MFT                                      uint8
	MNT                                      uint16
	MM                                       uint8
	AdditionalDataAvailable                  bool
	NA                                       uint16
	TauC                                     int64
	MN4                                      uint8
	MTauGPS                                  int32
	MLNFifth                                 bool
	Reserved                                 uint8
}

type RtcmMessages struct {
	_        noCopy
	resource *resource
	cleanup  runtime.Cleanup
}
type RtcmFrames struct {
	_        noCopy
	resource *resource
	cleanup  runtime.Cleanup
}
type RtcmDiagnostics struct {
	_        noCopy
	resource *resource
	cleanup  runtime.Cleanup
}
type RtcmLockTimeTracker struct {
	_        noCopy
	resource *resource
	cleanup  runtime.Cleanup
}

func validateRTCMMessageKindValue(value uint32) error {
	if value > RTCMMessageGalileoINavEphemerisValue {
		return invalidArgument("invalid RTCM message kind returned by native code")
	}
	return nil
}

func validateRTCMMSMKindValue(value uint32) error {
	if value != RTCMMSM4Value && value != RTCMMSM7Value {
		return invalidArgument("invalid RTCM MSM kind returned by native code")
	}
	return nil
}

func validateRTCMSSRKindValue(value uint32) error {
	if value > RTCMSSRVTECValue {
		return invalidArgument("invalid RTCM SSR kind returned by native code")
	}
	return nil
}

func validateRTCMFrameSkipReasonValue(value uint32) error {
	if value > RTCMFrameMalformedValue {
		return invalidArgument("invalid RTCM frame skip reason returned by native code")
	}
	return nil
}

func validateRTCMStringFieldValue(value uint32) error {
	if value > RTCMReceiverSerialNumberFieldValue {
		return invalidArgument("invalid RTCM antenna string field")
	}
	return nil
}

func newRtcmMessages(p *C.SidereonRtcmMessages) (*RtcmMessages, error) {
	if p == nil {
		return nil, errNilNativeHandle
	}
	h := &RtcmMessages{resource: &resource{ptr: unsafe.Pointer(p), release: func(x unsafe.Pointer) { C.sidereon_rtcm_messages_free((*C.SidereonRtcmMessages)(x)) }}}
	h.cleanup = runtime.AddCleanup(h, cleanupResource, h.resource)
	return h, nil
}
func newRtcmFrames(p *C.SidereonRtcmFrames) (*RtcmFrames, error) {
	if p == nil {
		return nil, errNilNativeHandle
	}
	h := &RtcmFrames{resource: &resource{ptr: unsafe.Pointer(p), release: func(x unsafe.Pointer) { C.sidereon_rtcm_frames_free((*C.SidereonRtcmFrames)(x)) }}}
	h.cleanup = runtime.AddCleanup(h, cleanupResource, h.resource)
	return h, nil
}
func newRtcmDiagnostics(p *C.SidereonRtcmStreamDiagnostics) (*RtcmDiagnostics, error) {
	if p == nil {
		return nil, errNilNativeHandle
	}
	h := &RtcmDiagnostics{resource: &resource{ptr: unsafe.Pointer(p), release: func(x unsafe.Pointer) { C.sidereon_rtcm_stream_diagnostics_free((*C.SidereonRtcmStreamDiagnostics)(x)) }}}
	h.cleanup = runtime.AddCleanup(h, cleanupResource, h.resource)
	return h, nil
}
func newRtcmTracker(p *C.SidereonRtcmLockTimeTracker) (*RtcmLockTimeTracker, error) {
	if p == nil {
		return nil, errNilNativeHandle
	}
	h := &RtcmLockTimeTracker{resource: &resource{ptr: unsafe.Pointer(p), release: func(x unsafe.Pointer) { C.sidereon_rtcm_lock_time_tracker_free((*C.SidereonRtcmLockTimeTracker)(x)) }}}
	h.cleanup = runtime.AddCleanup(h, cleanupResource, h.resource)
	return h, nil
}

func DecodeRTCM(data []byte) (*RtcmMessages, error) {
	var p *C.SidereonRtcmMessages
	err := withInput(data, func(b *C.uint8_t, n C.size_t) uint32 { return C.sidereon_rtcm_decode_messages(b, n, &p) })
	if err != nil {
		if p != nil {
			withCThread(func() { C.sidereon_rtcm_messages_free(p) })
		}
		return nil, err
	}
	handle, err := newRtcmMessages(p)
	if err != nil && p != nil {
		withCThread(func() { C.sidereon_rtcm_messages_free(p) })
	}
	return handle, err
}
func DecodeRTCMStream(data []byte) (*RtcmMessages, *RtcmDiagnostics, error) {
	length, err := checkedNativeSize(len(data))
	if err != nil {
		return nil, nil, err
	}
	var m *C.SidereonRtcmMessages
	var d *C.SidereonRtcmStreamDiagnostics
	var operationErr error
	withCThread(func() {
		var p unsafe.Pointer
		if len(data) > 0 {
			p = C.CBytes(data)
			if p == nil {
				operationErr = errors.New("sidereon: unable to allocate native input buffer")
				return
			}
			defer C.free(p)
		}
		operationErr = statusErrorLocked(C.sidereon_rtcm_decode_stream((*C.uint8_t)(p), length, &m, &d))
		if operationErr != nil {
			if m != nil {
				C.sidereon_rtcm_messages_free(m)
				m = nil
			}
			if d != nil {
				C.sidereon_rtcm_stream_diagnostics_free(d)
				d = nil
			}
		}
	})
	runtime.KeepAlive(data)
	if operationErr != nil {
		return nil, nil, operationErr
	}
	messages, err := newRtcmMessages(m)
	if err != nil {
		if d != nil {
			withCThread(func() { C.sidereon_rtcm_stream_diagnostics_free(d) })
		}
		return nil, nil, err
	}
	diag, err := newRtcmDiagnostics(d)
	if err != nil {
		if closeErr := messages.Close(); closeErr != nil {
			return nil, nil, errors.Join(err, closeErr)
		}
		return nil, nil, err
	}
	return messages, diag, nil
}
func ScanRTCMFrames(data []byte) (*RtcmFrames, error) {
	var p *C.SidereonRtcmFrames
	err := withInput(data, func(b *C.uint8_t, n C.size_t) uint32 { return C.sidereon_rtcm_scan_frames(b, n, &p) })
	if err != nil {
		if p != nil {
			withCThread(func() { C.sidereon_rtcm_frames_free(p) })
		}
		return nil, err
	}
	handle, err := newRtcmFrames(p)
	if err != nil && p != nil {
		withCThread(func() { C.sidereon_rtcm_frames_free(p) })
	}
	return handle, err
}

func DecodeRTCMFrame(data []byte) ([]byte, int, error) {
	inputLength, inputErr := checkedNativeSize(len(data))
	if inputErr != nil {
		return nil, 0, inputErr
	}
	var body []byte
	var frameLength C.size_t
	var err error
	withCThread(func() {
		var input unsafe.Pointer
		if len(data) > 0 {
			input = C.CBytes(data)
			if input == nil {
				err = errors.New("sidereon: unable to allocate native input buffer")
				return
			}
			defer C.free(input)
		}
		var written, required C.size_t
		if err = statusErrorLocked(C.sidereon_rtcm_decode_frame((*C.uint8_t)(input), inputLength, nil, 0, &written, &required, &frameLength)); err != nil {
			return
		}
		n, queryErr := validateNativeQuery("RTCM decoded frame body", uint64(written), uint64(required))
		if queryErr != nil {
			err = queryErr
			return
		}
		if _, queryErr = checkedNativeAllocationSize(n, 1); queryErr != nil {
			err = queryErr
			return
		}
		buffer := make([]byte, n)
		outputLength, queryErr := checkedNativeSize(n)
		if queryErr != nil {
			err = queryErr
			return
		}
		var output *C.uint8_t
		if n > 0 {
			output = (*C.uint8_t)(unsafe.Pointer(&buffer[0]))
		}
		if err = statusErrorLocked(C.sidereon_rtcm_decode_frame((*C.uint8_t)(input), inputLength, output, outputLength, &written, &required, &frameLength)); err != nil {
			return
		}
		count, validationErr := validateNativeOutput("RTCM decoded frame body", n, uint64(written), uint64(required))
		if validationErr != nil {
			err = validationErr
			return
		}
		body = append([]byte(nil), buffer[:count]...)
	})
	runtime.KeepAlive(data)
	frameCount, countErr := checkedNativeCount(uint64(frameLength))
	if err == nil && countErr != nil {
		err = countErr
	}
	return body, frameCount, err
}
func EncodeRTCMFrame(body []byte) ([]byte, error) {
	var out []byte
	err := withInputError(body, func(b *C.uint8_t, n C.size_t) error {
		var w, r C.size_t
		if err := callStatus(func() uint32 { return C.sidereon_rtcm_encode_frame(b, n, nil, 0, &w, &r) }); err != nil {
			return err
		}
		need, err := validateNativeQuery("RTCM frame", uint64(w), uint64(r))
		if err != nil {
			return err
		}
		mem, err := checkedNativeMalloc(need, 1)
		if err != nil {
			return err
		}
		if mem != nil {
			defer C.free(mem)
		}
		if err := callStatus(func() uint32 { return C.sidereon_rtcm_encode_frame(b, n, (*C.uint8_t)(mem), C.size_t(need), &w, &r) }); err != nil {
			return err
		}
		z, err := validateNativeOutput("RTCM frame", need, uint64(w), uint64(r))
		if err != nil {
			return err
		}
		out = make([]byte, z)
		copy(out, unsafe.Slice((*byte)(mem), z))
		return nil
	})
	return out, err
}

func (m *RtcmMessages) Close() error {
	if m == nil {
		return nil
	}
	return closeProtocolResource(m, m.resource, &m.cleanup)
}
func (f *RtcmFrames) Close() error {
	if f == nil {
		return nil
	}
	return closeProtocolResource(f, f.resource, &f.cleanup)
}
func (d *RtcmDiagnostics) Close() error {
	if d == nil {
		return nil
	}
	return closeProtocolResource(d, d.resource, &d.cleanup)
}
func (t *RtcmLockTimeTracker) Close() error {
	if t == nil {
		return nil
	}
	return closeProtocolResource(t, t.resource, &t.cleanup)
}
func (m *RtcmMessages) Count() (int, error) {
	var c C.size_t
	err := m.resource.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 { return C.sidereon_rtcm_messages_count((*C.SidereonRtcmMessages)(p), &c) })
	})
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(c))
}
func (m *RtcmMessages) Kind(index int) (NativeRTCMMessageInfo, error) {
	if index < 0 {
		return NativeRTCMMessageInfo{}, errNegativeIndex
	}
	var k uint32
	var n C.uint16_t
	err := m.resource.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_rtcm_message_kind((*C.SidereonRtcmMessages)(p), C.size_t(index), &k, &n)
		})
	})
	if err == nil {
		err = validateRTCMMessageKindValue(k)
	}
	return NativeRTCMMessageInfo{Kind: k, MessageNumber: uint16(n)}, err
}
func rtcmBytesCall(m *RtcmMessages, index int, frame bool) ([]byte, error) {
	if index < 0 {
		return nil, errNegativeIndex
	}
	var out []byte
	err := m.resource.with(func(p unsafe.Pointer) error {
		var e error
		out, e = copyByteOutput("RTCM encoded message", func(b *C.uint8_t, n C.size_t, w, r *C.size_t) uint32 {
			if frame {
				return C.sidereon_rtcm_message_to_frame((*C.SidereonRtcmMessages)(p), C.size_t(index), b, n, w, r)
			}
			return C.sidereon_rtcm_message_encode((*C.SidereonRtcmMessages)(p), C.size_t(index), b, n, w, r)
		})
		return e
	})
	runtime.KeepAlive(m)
	return out, err
}
func (m *RtcmMessages) Encode(index int) ([]byte, error) { return rtcmBytesCall(m, index, false) }
func (m *RtcmMessages) Frame(index int) ([]byte, error)  { return rtcmBytesCall(m, index, true) }
func (m *RtcmMessages) MSMInfo(index int) (NativeRTCMMSMInfo, error) {
	if index < 0 {
		return NativeRTCMMSMInfo{}, errNegativeIndex
	}
	var x C.SidereonRtcmMsmInfo
	err := m.resource.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_rtcm_message_msm_info((*C.SidereonRtcmMessages)(p), C.size_t(index), &x)
		})
	})
	if err != nil {
		return NativeRTCMMSMInfo{}, err
	}
	v := NativeRTCMMSMInfo{MessageNumber: uint16(x.message_number), System: uint32(x.system), Kind: uint32(x.kind), Header: NativeRTCMMSMHeader{ReferenceStationID: uint16(x.header.reference_station_id), EpochTime: uint32(x.header.epoch_time), MultipleMessage: bool(x.header.multiple_message), IODS: uint8(x.header.iods), Reserved: uint8(x.header.reserved), ClockSteering: uint8(x.header.clock_steering), ExternalClock: uint8(x.header.external_clock), DivergenceFreeSmoothing: bool(x.header.divergence_free_smoothing), SmoothingInterval: uint8(x.header.smoothing_interval)}}
	if err := validateGNSSSystemValue(v.System); err != nil {
		return NativeRTCMMSMInfo{}, err
	}
	if err := validateRTCMMSMKindValue(v.Kind); err != nil {
		return NativeRTCMMSMInfo{}, err
	}
	v.SatelliteCount, err = checkedNativeCount(uint64(x.satellite_count))
	if err != nil {
		return NativeRTCMMSMInfo{}, err
	}
	v.SignalCount, err = checkedNativeCount(uint64(x.signal_count))
	if err != nil {
		return NativeRTCMMSMInfo{}, err
	}
	return v, nil
}
func (m *RtcmMessages) MSMSatellites(index int) ([]NativeRTCMMSMSatellite, error) {
	if index < 0 {
		return nil, errNegativeIndex
	}
	var result []NativeRTCMMSMSatellite
	err := m.resource.with(func(p unsafe.Pointer) error {
		var w, r C.size_t
		if e := callStatus(func() uint32 {
			return C.sidereon_rtcm_message_msm_satellites((*C.SidereonRtcmMessages)(p), C.size_t(index), nil, 0, &w, &r)
		}); e != nil {
			return e
		}
		n, e := validateNativeQuery("RTCM MSM satellites", uint64(w), uint64(r))
		if e != nil {
			return e
		}
		if _, e := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonRtcmMsmSatellite{})); e != nil {
			return e
		}
		v := make([]C.SidereonRtcmMsmSatellite, n)
		var o *C.SidereonRtcmMsmSatellite
		if n > 0 {
			o = &v[0]
		}
		if e := callStatus(func() uint32 {
			return C.sidereon_rtcm_message_msm_satellites((*C.SidereonRtcmMessages)(p), C.size_t(index), o, C.size_t(n), &w, &r)
		}); e != nil {
			return e
		}
		z, e := validateNativeOutput("RTCM MSM satellites", n, uint64(w), uint64(r))
		if e != nil {
			return e
		}
		result = make([]NativeRTCMMSMSatellite, z)
		for i := range result {
			x := v[i]
			result[i] = NativeRTCMMSMSatellite{ID: uint8(x.id), RoughRangeMS: uint8(x.rough_range_ms), RoughRangeMod1: uint16(x.rough_range_mod1), HasExtendedInfo: bool(x.has_extended_info), ExtendedInfo: uint8(x.extended_info), HasRoughPhaseRangeRate: bool(x.has_rough_phase_range_rate), RoughPhaseRangeRateMS: int16(x.rough_phase_range_rate_m_s)}
		}
		return nil
	})
	runtime.KeepAlive(m)
	return result, err
}
func (m *RtcmMessages) MSMSignals(index int) ([]NativeRTCMMSMSignal, error) {
	if index < 0 {
		return nil, errNegativeIndex
	}
	var result []NativeRTCMMSMSignal
	err := m.resource.with(func(p unsafe.Pointer) error {
		var w, r C.size_t
		if e := callStatus(func() uint32 {
			return C.sidereon_rtcm_message_msm_signals((*C.SidereonRtcmMessages)(p), C.size_t(index), nil, 0, &w, &r)
		}); e != nil {
			return e
		}
		n, e := validateNativeQuery("RTCM MSM signals", uint64(w), uint64(r))
		if e != nil {
			return e
		}
		if _, e := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonRtcmMsmSignal{})); e != nil {
			return e
		}
		v := make([]C.SidereonRtcmMsmSignal, n)
		var o *C.SidereonRtcmMsmSignal
		if n > 0 {
			o = &v[0]
		}
		if e := callStatus(func() uint32 {
			return C.sidereon_rtcm_message_msm_signals((*C.SidereonRtcmMessages)(p), C.size_t(index), o, C.size_t(n), &w, &r)
		}); e != nil {
			return e
		}
		z, e := validateNativeOutput("RTCM MSM signals", n, uint64(w), uint64(r))
		if e != nil {
			return e
		}
		result = make([]NativeRTCMMSMSignal, z)
		for i := range result {
			x := v[i]
			result[i] = NativeRTCMMSMSignal{SatelliteID: uint8(x.satellite_id), SignalID: uint8(x.signal_id), FinePseudorange: int32(x.fine_pseudorange), FinePhaseRange: int32(x.fine_phase_range), LockTimeIndicator: uint16(x.lock_time_indicator), HalfCycleAmbiguity: bool(x.half_cycle_ambiguity), CNR: uint16(x.cnr), HasFinePhaseRangeRate: bool(x.has_fine_phase_range_rate), FinePhaseRangeRate: int16(x.fine_phase_range_rate)}
		}
		return nil
	})
	runtime.KeepAlive(m)
	return result, err
}
func (m *RtcmMessages) Station(index int) (NativeRTCMStationCoordinates, error) {
	if index < 0 {
		return NativeRTCMStationCoordinates{}, errNegativeIndex
	}
	var x C.SidereonRtcmStationCoordinates
	err := m.resource.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_rtcm_message_station_coordinates((*C.SidereonRtcmMessages)(p), C.size_t(index), &x)
		})
	})
	return NativeRTCMStationCoordinates{MessageNumber: uint16(x.message_number), ReferenceStationID: uint16(x.reference_station_id), ITRFRealizationYear: uint8(x.itrf_realization_year), GPS: bool(x.gps_indicator), GLONASS: bool(x.glonass_indicator), Galileo: bool(x.galileo_indicator), ReferenceStation: bool(x.reference_station_indicator), SingleReceiverOscillator: bool(x.single_receiver_oscillator), Reserved: bool(x.reserved), QuarterCycleIndicator: uint8(x.quarter_cycle_indicator), ECEFX: int64(x.ecef_x), ECEFY: int64(x.ecef_y), ECEFZ: int64(x.ecef_z), XM: float64(x.x_m), YM: float64(x.y_m), ZM: float64(x.z_m), HasAntennaHeight: bool(x.has_antenna_height), AntennaHeight: uint16(x.antenna_height), AntennaHeightM: float64(x.antenna_height_m)}, err
}
func (m *RtcmMessages) Antenna(index int) (NativeRTCMAntennaDescriptor, error) {
	if index < 0 {
		return NativeRTCMAntennaDescriptor{}, errNegativeIndex
	}
	var x C.SidereonRtcmAntennaDescriptor
	err := m.resource.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_rtcm_message_antenna_descriptor((*C.SidereonRtcmMessages)(p), C.size_t(index), &x)
		})
	})
	return NativeRTCMAntennaDescriptor{MessageNumber: uint16(x.message_number), ReferenceStationID: uint16(x.reference_station_id), AntennaSetupID: uint8(x.antenna_setup_id), HasAntennaSerialNumber: bool(x.has_antenna_serial_number), HasReceiverType: bool(x.has_receiver_type), HasReceiverFirmwareVersion: bool(x.has_receiver_firmware_version), HasReceiverSerialNumber: bool(x.has_receiver_serial_number)}, err
}
func (m *RtcmMessages) AntennaString(index int, field uint32) ([]byte, error) {
	if index < 0 {
		return nil, errNegativeIndex
	}
	if err := validateRTCMStringFieldValue(field); err != nil {
		return nil, err
	}
	var out []byte
	err := m.resource.with(func(p unsafe.Pointer) error {
		var e error
		out, e = copyByteOutput("RTCM antenna string", func(b *C.uint8_t, n C.size_t, w, r *C.size_t) uint32 {
			return C.sidereon_rtcm_message_antenna_string((*C.SidereonRtcmMessages)(p), C.size_t(index), C.uint32_t(field), b, n, w, r)
		})
		return e
	})
	return out, err
}

func (m *RtcmMessages) SSRInfo(index int) (NativeSsrInfo, error) {
	if index < 0 {
		return NativeSsrInfo{}, errNegativeIndex
	}
	var value C.SidereonRtcmSsrInfo
	err := m.resource.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_rtcm_message_ssr_info((*C.SidereonRtcmMessages)(pointer), C.size_t(index), &value)
		})
	})
	if err != nil {
		return NativeSsrInfo{}, err
	}
	if err := validateGNSSSystemValue(uint32(value.system)); err != nil {
		return NativeSsrInfo{}, err
	}
	if err := validateRTCMSSRKindValue(uint32(value.kind)); err != nil {
		return NativeSsrInfo{}, err
	}
	orbits, err := checkedNativeCount(uint64(value.orbit_count))
	if err != nil {
		return NativeSsrInfo{}, err
	}
	clocks, err := checkedNativeCount(uint64(value.clock_count))
	if err != nil {
		return NativeSsrInfo{}, err
	}
	ura, err := checkedNativeCount(uint64(value.ura_count))
	if err != nil {
		return NativeSsrInfo{}, err
	}
	codeBiases, err := checkedNativeCount(uint64(value.code_bias_count))
	if err != nil {
		return NativeSsrInfo{}, err
	}
	phaseBiases, err := checkedNativeCount(uint64(value.phase_bias_count))
	if err != nil {
		return NativeSsrInfo{}, err
	}
	return NativeSsrInfo{
		MessageNumber:  uint16(value.message_number),
		System:         uint32(value.system),
		Kind:           uint32(value.kind),
		Header:         ssrHeaderFromC(value.header),
		OrbitCount:     orbits,
		ClockCount:     clocks,
		URACount:       ura,
		CodeBiasCount:  codeBiases,
		PhaseBiasCount: phaseBiases,
	}, nil
}

func (m *RtcmMessages) SSROrbitRecords(index int) ([]NativeSsrOrbitRecord, error) {
	info, err := m.SSRInfo(index)
	if err != nil {
		return nil, err
	}
	return m.copySSROrbitRecords(index, info.OrbitCount)
}

func (m *RtcmMessages) copySSROrbitRecords(index, count int) ([]NativeSsrOrbitRecord, error) {
	nativeLength, err := checkedNativeSize(count)
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonRtcmSsrOrbitRecord{})); err != nil {
		return nil, err
	}
	values := make([]C.SidereonRtcmSsrOrbitRecord, count)
	var result []NativeSsrOrbitRecord
	err = m.resource.with(func(pointer unsafe.Pointer) error {
		var output *C.SidereonRtcmSsrOrbitRecord
		if len(values) > 0 {
			output = &values[0]
		}
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_rtcm_message_ssr_orbits((*C.SidereonRtcmMessages)(pointer), C.size_t(index), output, nativeLength, &written, &required)
		}); err != nil {
			return err
		}
		n, err := validateNativeOutput("RTCM SSR orbit records", len(values), uint64(written), uint64(required))
		if err != nil {
			return err
		}
		result = make([]NativeSsrOrbitRecord, n)
		for i := range result {
			value := values[i]
			result[i] = NativeSsrOrbitRecord{SatelliteID: uint8(value.satellite_id), IODE: uint32(value.iode), DeltaRadial: int32(value.delta_radial), DeltaAlong: int32(value.delta_along), DeltaCross: int32(value.delta_cross), DotDeltaRadial: int32(value.dot_delta_radial), DotDeltaAlong: int32(value.dot_delta_along), DotDeltaCross: int32(value.dot_delta_cross)}
		}
		return nil
	})
	runtime.KeepAlive(m)
	return result, err
}

func (m *RtcmMessages) SSRClockRecords(index int) ([]NativeSsrClockRecord, error) {
	info, err := m.SSRInfo(index)
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(info.ClockCount, unsafe.Sizeof(C.SidereonRtcmSsrClockRecord{})); err != nil {
		return nil, err
	}
	nativeLength, err := checkedNativeSize(info.ClockCount)
	if err != nil {
		return nil, err
	}
	values := make([]C.SidereonRtcmSsrClockRecord, info.ClockCount)
	var result []NativeSsrClockRecord
	err = m.resource.with(func(pointer unsafe.Pointer) error {
		var output *C.SidereonRtcmSsrClockRecord
		if len(values) > 0 {
			output = &values[0]
		}
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_rtcm_message_ssr_clocks((*C.SidereonRtcmMessages)(pointer), C.size_t(index), output, nativeLength, &written, &required)
		}); err != nil {
			return err
		}
		n, err := validateNativeOutput("RTCM SSR clock records", len(values), uint64(written), uint64(required))
		if err != nil {
			return err
		}
		result = make([]NativeSsrClockRecord, n)
		for i := range result {
			value := values[i]
			result[i] = NativeSsrClockRecord{SatelliteID: uint8(value.satellite_id), C0: int32(value.c0), C1: int32(value.c1), C2: int32(value.c2)}
		}
		return nil
	})
	runtime.KeepAlive(m)
	return result, err
}

func (m *RtcmMessages) SSRCodeBiasRecords(index int) ([]NativeSsrCodeBiasRecord, error) {
	info, err := m.SSRInfo(index)
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(info.CodeBiasCount, unsafe.Sizeof(C.SidereonRtcmSsrCodeBiasRecord{})); err != nil {
		return nil, err
	}
	nativeLength, err := checkedNativeSize(info.CodeBiasCount)
	if err != nil {
		return nil, err
	}
	values := make([]C.SidereonRtcmSsrCodeBiasRecord, info.CodeBiasCount)
	var result []NativeSsrCodeBiasRecord
	err = m.resource.with(func(pointer unsafe.Pointer) error {
		var output *C.SidereonRtcmSsrCodeBiasRecord
		if len(values) > 0 {
			output = &values[0]
		}
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_rtcm_message_ssr_code_biases((*C.SidereonRtcmMessages)(pointer), C.size_t(index), output, nativeLength, &written, &required)
		}); err != nil {
			return err
		}
		n, err := validateNativeOutput("RTCM SSR code-bias records", len(values), uint64(written), uint64(required))
		if err != nil {
			return err
		}
		result = make([]NativeSsrCodeBiasRecord, n)
		for i := range result {
			value := values[i]
			count, err := checkedNativeCount(uint64(value.signal_count))
			if err != nil {
				return err
			}
			result[i] = NativeSsrCodeBiasRecord{SatelliteID: uint8(value.satellite_id), SignalCount: count}
		}
		return nil
	})
	runtime.KeepAlive(m)
	return result, err
}

func (m *RtcmMessages) SSRCodeBiasSignals(index, recordIndex int) ([]NativeSsrCodeBiasSignal, error) {
	if index < 0 || recordIndex < 0 {
		return nil, errNegativeIndex
	}
	records, err := m.SSRCodeBiasRecords(index)
	if err != nil {
		return nil, err
	}
	if recordIndex >= len(records) {
		return nil, errors.New("sidereon: RTCM SSR code-bias record index out of range")
	}
	count := records[recordIndex].SignalCount
	if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonRtcmSsrCodeBiasSignal{})); err != nil {
		return nil, err
	}
	nativeLength, err := checkedNativeSize(count)
	if err != nil {
		return nil, err
	}
	values := make([]C.SidereonRtcmSsrCodeBiasSignal, count)
	var result []NativeSsrCodeBiasSignal
	err = m.resource.with(func(pointer unsafe.Pointer) error {
		var output *C.SidereonRtcmSsrCodeBiasSignal
		if len(values) > 0 {
			output = &values[0]
		}
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_rtcm_message_ssr_code_bias_signals((*C.SidereonRtcmMessages)(pointer), C.size_t(index), C.size_t(recordIndex), output, nativeLength, &written, &required)
		}); err != nil {
			return err
		}
		n, err := validateNativeOutput("RTCM SSR code-bias signals", len(values), uint64(written), uint64(required))
		if err != nil {
			return err
		}
		result = make([]NativeSsrCodeBiasSignal, n)
		for i := range result {
			result[i] = NativeSsrCodeBiasSignal{SignalID: uint8(values[i].signal_id), Bias: int16(values[i].bias)}
		}
		return nil
	})
	runtime.KeepAlive(m)
	return result, err
}

func (m *RtcmMessages) SSRPhaseBiasRecords(index int) ([]NativeSsrPhaseBiasRecord, error) {
	info, err := m.SSRInfo(index)
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(info.PhaseBiasCount, unsafe.Sizeof(C.SidereonRtcmSsrPhaseBiasRecord{})); err != nil {
		return nil, err
	}
	nativeLength, err := checkedNativeSize(info.PhaseBiasCount)
	if err != nil {
		return nil, err
	}
	values := make([]C.SidereonRtcmSsrPhaseBiasRecord, info.PhaseBiasCount)
	var result []NativeSsrPhaseBiasRecord
	err = m.resource.with(func(pointer unsafe.Pointer) error {
		var output *C.SidereonRtcmSsrPhaseBiasRecord
		if len(values) > 0 {
			output = &values[0]
		}
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_rtcm_message_ssr_phase_biases((*C.SidereonRtcmMessages)(pointer), C.size_t(index), output, nativeLength, &written, &required)
		}); err != nil {
			return err
		}
		n, err := validateNativeOutput("RTCM SSR phase-bias records", len(values), uint64(written), uint64(required))
		if err != nil {
			return err
		}
		result = make([]NativeSsrPhaseBiasRecord, n)
		for i := range result {
			value := values[i]
			count, err := checkedNativeCount(uint64(value.signal_count))
			if err != nil {
				return err
			}
			result[i] = NativeSsrPhaseBiasRecord{SatelliteID: uint8(value.satellite_id), YawAngle: uint16(value.yaw_angle), YawRate: int8(value.yaw_rate), SignalCount: count}
		}
		return nil
	})
	runtime.KeepAlive(m)
	return result, err
}

func (m *RtcmMessages) SSRPhaseBiasSignals(index, recordIndex int) ([]NativeSsrPhaseBiasSignal, error) {
	if index < 0 || recordIndex < 0 {
		return nil, errNegativeIndex
	}
	records, err := m.SSRPhaseBiasRecords(index)
	if err != nil {
		return nil, err
	}
	if recordIndex >= len(records) {
		return nil, errors.New("sidereon: RTCM SSR phase-bias record index out of range")
	}
	count := records[recordIndex].SignalCount
	if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonRtcmSsrPhaseBiasSignal{})); err != nil {
		return nil, err
	}
	nativeLength, err := checkedNativeSize(count)
	if err != nil {
		return nil, err
	}
	values := make([]C.SidereonRtcmSsrPhaseBiasSignal, count)
	var result []NativeSsrPhaseBiasSignal
	err = m.resource.with(func(pointer unsafe.Pointer) error {
		var output *C.SidereonRtcmSsrPhaseBiasSignal
		if len(values) > 0 {
			output = &values[0]
		}
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_rtcm_message_ssr_phase_bias_signals((*C.SidereonRtcmMessages)(pointer), C.size_t(index), C.size_t(recordIndex), output, nativeLength, &written, &required)
		}); err != nil {
			return err
		}
		n, err := validateNativeOutput("RTCM SSR phase-bias signals", len(values), uint64(written), uint64(required))
		if err != nil {
			return err
		}
		result = make([]NativeSsrPhaseBiasSignal, n)
		for i := range result {
			value := values[i]
			result[i] = NativeSsrPhaseBiasSignal{SignalID: uint8(value.signal_id), IntegerIndicator: uint8(value.integer_indicator), WideLaneIntegerIndicator: uint8(value.wide_lane_integer_indicator), DiscontinuityCounter: uint8(value.discontinuity_counter), Bias: int32(value.bias)}
		}
		return nil
	})
	runtime.KeepAlive(m)
	return result, err
}

func (m *RtcmMessages) SSRURARecords(index int) ([]NativeSsrUraRecord, error) {
	info, err := m.SSRInfo(index)
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(info.URACount, unsafe.Sizeof(C.SidereonRtcmSsrUraRecord{})); err != nil {
		return nil, err
	}
	nativeLength, err := checkedNativeSize(info.URACount)
	if err != nil {
		return nil, err
	}
	values := make([]C.SidereonRtcmSsrUraRecord, info.URACount)
	var result []NativeSsrUraRecord
	err = m.resource.with(func(pointer unsafe.Pointer) error {
		var output *C.SidereonRtcmSsrUraRecord
		if len(values) > 0 {
			output = &values[0]
		}
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_rtcm_message_ssr_ura((*C.SidereonRtcmMessages)(pointer), C.size_t(index), output, nativeLength, &written, &required)
		}); err != nil {
			return err
		}
		n, err := validateNativeOutput("RTCM SSR URA records", len(values), uint64(written), uint64(required))
		if err != nil {
			return err
		}
		result = make([]NativeSsrUraRecord, n)
		for i := range result {
			result[i] = NativeSsrUraRecord{SatelliteID: uint8(values[i].satellite_id), URAIndex: uint8(values[i].ura_index)}
		}
		return nil
	})
	runtime.KeepAlive(m)
	return result, err
}

func buildRTCMMessage(call func(**C.SidereonRtcmMessages) uint32) (*RtcmMessages, error) {
	var pointer *C.SidereonRtcmMessages
	var err error
	withCThread(func() {
		err = statusErrorLocked(call(&pointer))
		if err != nil && pointer != nil {
			C.sidereon_rtcm_messages_free(pointer)
			pointer = nil
		}
	})
	if err != nil {
		return nil, err
	}
	return newRtcmMessages(pointer)
}

func BuildRTCMAntennaDescriptor(messageNumber, referenceStationID uint16, setup uint8, descriptor string, serial, receiverType, firmware, receiverSerial *string) (*RtcmMessages, error) {
	if err := validateRTCMString(descriptor); err != nil {
		return nil, err
	}
	for _, value := range []*string{serial, receiverType, firmware, receiverSerial} {
		if value != nil {
			if err := validateRTCMString(*value); err != nil {
				return nil, err
			}
		}
	}
	var err error
	var result *RtcmMessages
	withCThread(func() {
		cDescriptor := C.CString(descriptor)
		if cDescriptor == nil {
			err = errors.New("sidereon: unable to allocate native string")
			return
		}
		defer C.free(unsafe.Pointer(cDescriptor))
		strings := make([]*C.char, 4)
		values := []*string{serial, receiverType, firmware, receiverSerial}
		for i, value := range values {
			if value == nil {
				continue
			}
			strings[i] = C.CString(*value)
			if strings[i] == nil {
				err = errors.New("sidereon: unable to allocate native string")
				for _, pointer := range strings {
					if pointer != nil {
						C.free(unsafe.Pointer(pointer))
					}
				}
				return
			}
			defer C.free(unsafe.Pointer(strings[i]))
		}
		var pointer *C.SidereonRtcmMessages
		status := C.sidereon_rtcm_build_antenna_descriptor(C.uint16_t(messageNumber), C.uint16_t(referenceStationID), C.uint8_t(setup), cDescriptor, strings[0], strings[1], strings[2], strings[3], &pointer)
		err = statusErrorLocked(status)
		if err != nil {
			if pointer != nil {
				C.sidereon_rtcm_messages_free(pointer)
			}
			return
		}
		result, err = newRtcmMessages(pointer)
		if err != nil && pointer != nil {
			C.sidereon_rtcm_messages_free(pointer)
		}
	})
	return result, err
}

func validateRTCMString(value string) error {
	if err := rejectEmbeddedNUL(value, "RTCM string"); err != nil {
		return err
	}
	if len(value) > 255 {
		return errors.New("sidereon: RTCM string exceeds 255 bytes")
	}
	return nil
}

func BuildRTCMMSM(info NativeRTCMMSMInfo, satellites []NativeRTCMMSMSatellite, signals []NativeRTCMMSMSignal) (*RtcmMessages, error) {
	if err := validateGNSSSystemValue(info.System); err != nil {
		return nil, err
	}
	if err := validateRTCMMSMKindValue(info.Kind); err != nil {
		return nil, err
	}
	satelliteCount, err := checkedNativeSize(len(satellites))
	if err != nil {
		return nil, err
	}
	signalCount, err := checkedNativeSize(len(signals))
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(len(satellites), unsafe.Sizeof(C.SidereonRtcmMsmSatellite{})); err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(len(signals), unsafe.Sizeof(C.SidereonRtcmMsmSignal{})); err != nil {
		return nil, err
	}
	satelliteMemory, err := checkedNativeMalloc(len(satellites), unsafe.Sizeof(C.SidereonRtcmMsmSatellite{}))
	if err != nil {
		return nil, err
	}
	if satelliteMemory != nil {
		defer C.free(satelliteMemory)
	}
	csatellites := unsafe.Slice((*C.SidereonRtcmMsmSatellite)(satelliteMemory), len(satellites))
	for i, value := range satellites {
		csatellites[i] = C.SidereonRtcmMsmSatellite{id: C.uint8_t(value.ID), rough_range_ms: C.uint8_t(value.RoughRangeMS), rough_range_mod1: C.uint16_t(value.RoughRangeMod1), has_extended_info: C.bool(value.HasExtendedInfo), extended_info: C.uint8_t(value.ExtendedInfo), has_rough_phase_range_rate: C.bool(value.HasRoughPhaseRangeRate), rough_phase_range_rate_m_s: C.int16_t(value.RoughPhaseRangeRateMS)}
	}
	signalMemory, err := checkedNativeMalloc(len(signals), unsafe.Sizeof(C.SidereonRtcmMsmSignal{}))
	if err != nil {
		return nil, err
	}
	if signalMemory != nil {
		defer C.free(signalMemory)
	}
	csignals := unsafe.Slice((*C.SidereonRtcmMsmSignal)(signalMemory), len(signals))
	for i, value := range signals {
		csignals[i] = C.SidereonRtcmMsmSignal{satellite_id: C.uint8_t(value.SatelliteID), signal_id: C.uint8_t(value.SignalID), fine_pseudorange: C.int32_t(value.FinePseudorange), fine_phase_range: C.int32_t(value.FinePhaseRange), lock_time_indicator: C.uint16_t(value.LockTimeIndicator), half_cycle_ambiguity: C.bool(value.HalfCycleAmbiguity), cnr: C.uint16_t(value.CNR), has_fine_phase_range_rate: C.bool(value.HasFinePhaseRangeRate), fine_phase_range_rate: C.int16_t(value.FinePhaseRangeRate)}
	}
	infoMemory, err := checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonRtcmMsmInfo{}))
	if err != nil {
		return nil, err
	}
	defer C.free(infoMemory)
	cinfo := (*C.SidereonRtcmMsmInfo)(infoMemory)
	*cinfo = C.SidereonRtcmMsmInfo{message_number: C.uint16_t(info.MessageNumber), system: C.enum_SidereonGnssSystem(info.System), kind: C.enum_SidereonRtcmMsmKind(info.Kind), header: C.SidereonRtcmMsmHeader{reference_station_id: C.uint16_t(info.Header.ReferenceStationID), epoch_time: C.uint32_t(info.Header.EpochTime), multiple_message: C.bool(info.Header.MultipleMessage), iods: C.uint8_t(info.Header.IODS), reserved: C.uint8_t(info.Header.Reserved), clock_steering: C.uint8_t(info.Header.ClockSteering), external_clock: C.uint8_t(info.Header.ExternalClock), divergence_free_smoothing: C.bool(info.Header.DivergenceFreeSmoothing), smoothing_interval: C.uint8_t(info.Header.SmoothingInterval)}, satellite_count: satelliteCount, signal_count: signalCount}
	return buildRTCMMessage(func(out **C.SidereonRtcmMessages) uint32 {
		var satellitePointer *C.SidereonRtcmMsmSatellite
		var signalPointer *C.SidereonRtcmMsmSignal
		if len(csatellites) > 0 {
			satellitePointer = (*C.SidereonRtcmMsmSatellite)(satelliteMemory)
		}
		if len(csignals) > 0 {
			signalPointer = (*C.SidereonRtcmMsmSignal)(signalMemory)
		}
		return C.sidereon_rtcm_build_msm(cinfo, satellitePointer, satelliteCount, signalPointer, signalCount, out)
	})
}

func BuildRTCMStation(value NativeRTCMStationCoordinates) (*RtcmMessages, error) {
	memory, err := checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonRtcmStationCoordinates{}))
	if err != nil {
		return nil, err
	}
	defer C.free(memory)
	station := (*C.SidereonRtcmStationCoordinates)(memory)
	*station = C.SidereonRtcmStationCoordinates{message_number: C.uint16_t(value.MessageNumber), reference_station_id: C.uint16_t(value.ReferenceStationID), itrf_realization_year: C.uint8_t(value.ITRFRealizationYear), gps_indicator: C.bool(value.GPS), glonass_indicator: C.bool(value.GLONASS), galileo_indicator: C.bool(value.Galileo), reference_station_indicator: C.bool(value.ReferenceStation), single_receiver_oscillator: C.bool(value.SingleReceiverOscillator), reserved: C.bool(value.Reserved), quarter_cycle_indicator: C.uint8_t(value.QuarterCycleIndicator), ecef_x: C.int64_t(value.ECEFX), ecef_y: C.int64_t(value.ECEFY), ecef_z: C.int64_t(value.ECEFZ), x_m: C.double(value.XM), y_m: C.double(value.YM), z_m: C.double(value.ZM), has_antenna_height: C.bool(value.HasAntennaHeight), antenna_height: C.uint16_t(value.AntennaHeight), antenna_height_m: C.double(value.AntennaHeightM)}
	return buildRTCMMessage(func(out **C.SidereonRtcmMessages) uint32 {
		return C.sidereon_rtcm_build_station_coordinates(station, out)
	})
}

func BuildRTCMGPSEphemeris(value NativeRTCMGPSEphemeris) (*RtcmMessages, error) {
	memory, err := checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonRtcmGpsEphemeris{}))
	if err != nil {
		return nil, err
	}
	defer C.free(memory)
	eph := (*C.SidereonRtcmGpsEphemeris)(memory)
	*eph = C.SidereonRtcmGpsEphemeris{satellite_id: C.uint8_t(value.SatelliteID), week_number: C.uint16_t(value.WeekNumber), sv_accuracy: C.uint8_t(value.SVAccuracy), code_on_l2: C.uint8_t(value.CodeOnL2), idot: C.int32_t(value.IDOT), iode: C.uint8_t(value.IODE), t_oc: C.uint16_t(value.TOC), a_f2: C.int16_t(value.AF2), a_f1: C.int32_t(value.AF1), a_f0: C.int32_t(value.AF0), iodc: C.uint16_t(value.IODC), c_rs: C.int32_t(value.CRS), delta_n: C.int32_t(value.DeltaN), m0: C.int64_t(value.M0), c_uc: C.int32_t(value.CUC), eccentricity: C.uint64_t(value.Eccentricity), c_us: C.int32_t(value.CUS), sqrt_a: C.uint64_t(value.SqrtA), t_oe: C.uint16_t(value.TOE), c_ic: C.int32_t(value.CIC), omega0: C.int64_t(value.Omega0), c_is: C.int32_t(value.CIS), i0: C.int64_t(value.I0), c_rc: C.int32_t(value.CRC), omega: C.int64_t(value.Omega), omega_dot: C.int32_t(value.OmegaDot), t_gd: C.int16_t(value.TGD), sv_health: C.uint8_t(value.SVHealth), l2_p_data_flag: C.bool(value.L2PDataFlag), fit_interval: C.bool(value.FitInterval)}
	return buildRTCMMessage(func(out **C.SidereonRtcmMessages) uint32 { return C.sidereon_rtcm_build_gps_ephemeris(eph, out) })
}

func BuildRTCMQZSSEphemeris(value NativeRTCMQZSSEphemeris) (*RtcmMessages, error) {
	memory, err := checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonRtcmQzssEphemeris{}))
	if err != nil {
		return nil, err
	}
	defer C.free(memory)
	eph := (*C.SidereonRtcmQzssEphemeris)(memory)
	*eph = C.SidereonRtcmQzssEphemeris{satellite_id: C.uint8_t(value.SatelliteID), t_oc: C.uint16_t(value.TOC), a_f2: C.int16_t(value.AF2), a_f1: C.int32_t(value.AF1), a_f0: C.int32_t(value.AF0), iode: C.uint8_t(value.IODE), c_rs: C.int32_t(value.CRS), delta_n: C.int32_t(value.DeltaN), m0: C.int64_t(value.M0), c_uc: C.int32_t(value.CUC), eccentricity: C.uint64_t(value.Eccentricity), c_us: C.int32_t(value.CUS), sqrt_a: C.uint64_t(value.SqrtA), t_oe: C.uint16_t(value.TOE), c_ic: C.int32_t(value.CIC), omega0: C.int64_t(value.Omega0), c_is: C.int32_t(value.CIS), i0: C.int64_t(value.I0), c_rc: C.int32_t(value.CRC), omega: C.int64_t(value.Omega), omega_dot: C.int32_t(value.OmegaDot), idot: C.int32_t(value.IDOT), codes_on_l2: C.uint8_t(value.CodesOnL2), week_number: C.uint16_t(value.WeekNumber), ura: C.uint8_t(value.URA), sv_health: C.uint8_t(value.SVHealth), t_gd: C.int16_t(value.TGD), iodc: C.uint16_t(value.IODC), fit_interval: C.bool(value.FitInterval)}
	return buildRTCMMessage(func(out **C.SidereonRtcmMessages) uint32 { return C.sidereon_rtcm_build_qzss_ephemeris(eph, out) })
}

func BuildRTCMBeiDouEphemeris(value NativeRTCMBeidouEphemeris) (*RtcmMessages, error) {
	memory, err := checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonRtcmBeidouEphemeris{}))
	if err != nil {
		return nil, err
	}
	defer C.free(memory)
	eph := C.SidereonRtcmBeidouEphemeris{satellite_id: C.uint8_t(value.SatelliteID), week_number: C.uint16_t(value.WeekNumber), sv_urai: C.uint8_t(value.SVURAI), idot: C.int32_t(value.IDOT), aode: C.uint8_t(value.AODE), t_oc: C.uint32_t(value.TOC), a_f2: C.int16_t(value.AF2), a_f1: C.int32_t(value.AF1), a_f0: C.int32_t(value.AF0), aodc: C.uint8_t(value.AODC), c_rs: C.int32_t(value.CRS), delta_n: C.int32_t(value.DeltaN), m0: C.int64_t(value.M0), c_uc: C.int32_t(value.CUC), eccentricity: C.uint64_t(value.Eccentricity), c_us: C.int32_t(value.CUS), sqrt_a: C.uint64_t(value.SqrtA), t_oe: C.uint32_t(value.TOE), c_ic: C.int32_t(value.CIC), omega0: C.int64_t(value.Omega0), c_is: C.int32_t(value.CIS), i0: C.int64_t(value.I0), c_rc: C.int32_t(value.CRC), omega: C.int64_t(value.Omega), omega_dot: C.int32_t(value.OmegaDot), t_gd1: C.int16_t(value.TGD1), t_gd2: C.int16_t(value.TGD2), sv_health: C.bool(value.SVHealth)}
	*(*C.SidereonRtcmBeidouEphemeris)(memory) = eph
	return buildRTCMMessage(func(out **C.SidereonRtcmMessages) uint32 {
		return C.sidereon_rtcm_build_beidou_ephemeris((*C.SidereonRtcmBeidouEphemeris)(memory), out)
	})
}

func BuildRTCMGalileoFNavEphemeris(value NativeRTCMGalileoFNavEphemeris) (*RtcmMessages, error) {
	memory, err := checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonRtcmGalileoFnavEphemeris{}))
	if err != nil {
		return nil, err
	}
	defer C.free(memory)
	eph := C.SidereonRtcmGalileoFnavEphemeris{satellite_id: C.uint8_t(value.SatelliteID), week_number: C.uint16_t(value.WeekNumber), iod_nav: C.uint16_t(value.IodNav), sisa: C.uint8_t(value.SISA), idot: C.int32_t(value.IDOT), t_oc: C.uint16_t(value.TOC), a_f2: C.int16_t(value.AF2), a_f1: C.int32_t(value.AF1), a_f0: C.int64_t(value.AF0), c_rs: C.int32_t(value.CRS), delta_n: C.int32_t(value.DeltaN), m0: C.int64_t(value.M0), c_uc: C.int32_t(value.CUC), eccentricity: C.uint64_t(value.Eccentricity), c_us: C.int32_t(value.CUS), sqrt_a: C.uint64_t(value.SqrtA), t_oe: C.uint16_t(value.TOE), c_ic: C.int32_t(value.CIC), omega0: C.int64_t(value.Omega0), c_is: C.int32_t(value.CIS), i0: C.int64_t(value.I0), c_rc: C.int32_t(value.CRC), omega: C.int64_t(value.Omega), omega_dot: C.int32_t(value.OmegaDot), bgd_e5a_e1: C.int16_t(value.BGDE5AE1), e5a_signal_health: C.uint8_t(value.E5ASignalHealth), e5a_data_validity: C.bool(value.E5ADataValidity), reserved: C.uint8_t(value.Reserved)}
	*(*C.SidereonRtcmGalileoFnavEphemeris)(memory) = eph
	return buildRTCMMessage(func(out **C.SidereonRtcmMessages) uint32 {
		return C.sidereon_rtcm_build_galileo_fnav_ephemeris((*C.SidereonRtcmGalileoFnavEphemeris)(memory), out)
	})
}

func BuildRTCMGalileoINavEphemeris(value NativeRTCMGalileoINavEphemeris) (*RtcmMessages, error) {
	memory, err := checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonRtcmGalileoInavEphemeris{}))
	if err != nil {
		return nil, err
	}
	defer C.free(memory)
	eph := C.SidereonRtcmGalileoInavEphemeris{satellite_id: C.uint8_t(value.SatelliteID), week_number: C.uint16_t(value.WeekNumber), iod_nav: C.uint16_t(value.IodNav), sisa_index: C.uint8_t(value.SISAIndex), idot: C.int32_t(value.IDOT), t_oc: C.uint16_t(value.TOC), a_f2: C.int16_t(value.AF2), a_f1: C.int32_t(value.AF1), a_f0: C.int64_t(value.AF0), c_rs: C.int32_t(value.CRS), delta_n: C.int32_t(value.DeltaN), m0: C.int64_t(value.M0), c_uc: C.int32_t(value.CUC), eccentricity: C.uint64_t(value.Eccentricity), c_us: C.int32_t(value.CUS), sqrt_a: C.uint64_t(value.SqrtA), t_oe: C.uint16_t(value.TOE), c_ic: C.int32_t(value.CIC), omega0: C.int64_t(value.Omega0), c_is: C.int32_t(value.CIS), i0: C.int64_t(value.I0), c_rc: C.int32_t(value.CRC), omega: C.int64_t(value.Omega), omega_dot: C.int32_t(value.OmegaDot), bgd_e5a_e1: C.int16_t(value.BGDE5AE1), bgd_e5b_e1: C.int16_t(value.BGDE5BE1), e5b_signal_health: C.uint8_t(value.E5BSignalHealth), e5b_data_validity: C.bool(value.E5BDataValidity), e1b_signal_health: C.uint8_t(value.E1BSignalHealth), e1b_data_validity: C.bool(value.E1BDataValidity), reserved: C.uint8_t(value.Reserved)}
	*(*C.SidereonRtcmGalileoInavEphemeris)(memory) = eph
	return buildRTCMMessage(func(out **C.SidereonRtcmMessages) uint32 {
		return C.sidereon_rtcm_build_galileo_inav_ephemeris((*C.SidereonRtcmGalileoInavEphemeris)(memory), out)
	})
}

func BuildRTCMGLONASSEphemeris(value NativeRTCMGLONASSEphemeris) (*RtcmMessages, error) {
	memory, err := checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonRtcmGlonassEphemeris{}))
	if err != nil {
		return nil, err
	}
	defer C.free(memory)
	eph := C.SidereonRtcmGlonassEphemeris{satellite_id: C.uint8_t(value.SatelliteID), frequency_channel: C.uint8_t(value.FrequencyChannel), almanac_health: C.bool(value.AlmanacHealth), almanac_health_availability: C.bool(value.AlmanacHealthAvailability), p1: C.uint8_t(value.P1), t_k: C.uint16_t(value.TK), b_n_msb: C.bool(value.BNMSB), p2: C.bool(value.P2), t_b: C.uint8_t(value.TB), xn_dot: C.int32_t(value.XNDot), xn: C.int32_t(value.XN), xn_dot_dot: C.int8_t(value.XNDotDot), yn_dot: C.int32_t(value.YNDot), yn: C.int32_t(value.YN), yn_dot_dot: C.int8_t(value.YNDotDot), zn_dot: C.int32_t(value.ZNDot), zn: C.int32_t(value.ZN), zn_dot_dot: C.int8_t(value.ZNDotDot), p3: C.bool(value.P3), gamma_n: C.int16_t(value.GammaN), m_p: C.uint8_t(value.MP), m_l_n_third: C.bool(value.MLNThird), tau_n: C.int32_t(value.TauN), delta_tau_n: C.int8_t(value.DeltaTauN), e_n: C.uint8_t(value.EN), m_p4: C.bool(value.MP4), m_f_t: C.uint8_t(value.MFT), m_n_t: C.uint16_t(value.MNT), m_m: C.uint8_t(value.MM), additional_data_available: C.bool(value.AdditionalDataAvailable), n_a: C.uint16_t(value.NA), tau_c: C.int64_t(value.TauC), m_n4: C.uint8_t(value.MN4), m_tau_gps: C.int32_t(value.MTauGPS), m_l_n_fifth: C.bool(value.MLNFifth), reserved: C.uint8_t(value.Reserved)}
	*(*C.SidereonRtcmGlonassEphemeris)(memory) = eph
	return buildRTCMMessage(func(out **C.SidereonRtcmMessages) uint32 {
		return C.sidereon_rtcm_build_glonass_ephemeris((*C.SidereonRtcmGlonassEphemeris)(memory), out)
	})
}

func (m *RtcmMessages) GPSEphemeris(index int) (NativeRTCMGPSEphemeris, error) {
	if index < 0 {
		return NativeRTCMGPSEphemeris{}, errNegativeIndex
	}
	var x C.SidereonRtcmGpsEphemeris
	err := m.resource.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_rtcm_message_gps_ephemeris((*C.SidereonRtcmMessages)(p), C.size_t(index), &x)
		})
	})
	return NativeRTCMGPSEphemeris{SatelliteID: uint8(x.satellite_id), WeekNumber: uint16(x.week_number), SVAccuracy: uint8(x.sv_accuracy), CodeOnL2: uint8(x.code_on_l2), IDOT: int32(x.idot), IODE: uint8(x.iode), TOC: uint16(x.t_oc), AF2: int16(x.a_f2), AF1: int32(x.a_f1), AF0: int32(x.a_f0), IODC: uint16(x.iodc), CRS: int32(x.c_rs), DeltaN: int32(x.delta_n), M0: int64(x.m0), CUC: int32(x.c_uc), Eccentricity: uint64(x.eccentricity), CUS: int32(x.c_us), SqrtA: uint64(x.sqrt_a), TOE: uint16(x.t_oe), CIC: int32(x.c_ic), Omega0: int64(x.omega0), CIS: int32(x.c_is), I0: int64(x.i0), CRC: int32(x.c_rc), Omega: int64(x.omega), OmegaDot: int32(x.omega_dot), TGD: int16(x.t_gd), SVHealth: uint8(x.sv_health), L2PDataFlag: bool(x.l2_p_data_flag), FitInterval: bool(x.fit_interval)}, err
}
func (m *RtcmMessages) QZSSEphemeris(index int) (NativeRTCMQZSSEphemeris, error) {
	if index < 0 {
		return NativeRTCMQZSSEphemeris{}, errNegativeIndex
	}
	var x C.SidereonRtcmQzssEphemeris
	err := m.resource.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_rtcm_message_qzss_ephemeris((*C.SidereonRtcmMessages)(p), C.size_t(index), &x)
		})
	})
	return NativeRTCMQZSSEphemeris{SatelliteID: uint8(x.satellite_id), TOC: uint16(x.t_oc), AF2: int16(x.a_f2), AF1: int32(x.a_f1), AF0: int32(x.a_f0), IODE: uint8(x.iode), CRS: int32(x.c_rs), DeltaN: int32(x.delta_n), M0: int64(x.m0), CUC: int32(x.c_uc), Eccentricity: uint64(x.eccentricity), CUS: int32(x.c_us), SqrtA: uint64(x.sqrt_a), TOE: uint16(x.t_oe), CIC: int32(x.c_ic), Omega0: int64(x.omega0), CIS: int32(x.c_is), I0: int64(x.i0), CRC: int32(x.c_rc), Omega: int64(x.omega), OmegaDot: int32(x.omega_dot), IDOT: int32(x.idot), CodesOnL2: uint8(x.codes_on_l2), WeekNumber: uint16(x.week_number), URA: uint8(x.ura), SVHealth: uint8(x.sv_health), TGD: int16(x.t_gd), IODC: uint16(x.iodc), FitInterval: bool(x.fit_interval)}, err
}
func (m *RtcmMessages) BeidouEphemeris(index int) (NativeRTCMBeidouEphemeris, error) {
	if index < 0 {
		return NativeRTCMBeidouEphemeris{}, errNegativeIndex
	}
	var x C.SidereonRtcmBeidouEphemeris
	err := m.resource.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_rtcm_message_beidou_ephemeris((*C.SidereonRtcmMessages)(p), C.size_t(index), &x)
		})
	})
	return NativeRTCMBeidouEphemeris{SatelliteID: uint8(x.satellite_id), WeekNumber: uint16(x.week_number), SVURAI: uint8(x.sv_urai), IDOT: int32(x.idot), AODE: uint8(x.aode), TOC: uint32(x.t_oc), AF2: int16(x.a_f2), AF1: int32(x.a_f1), AF0: int32(x.a_f0), AODC: uint8(x.aodc), CRS: int32(x.c_rs), DeltaN: int32(x.delta_n), M0: int64(x.m0), CUC: int32(x.c_uc), Eccentricity: uint64(x.eccentricity), CUS: int32(x.c_us), SqrtA: uint64(x.sqrt_a), TOE: uint32(x.t_oe), CIC: int32(x.c_ic), Omega0: int64(x.omega0), CIS: int32(x.c_is), I0: int64(x.i0), CRC: int32(x.c_rc), Omega: int64(x.omega), OmegaDot: int32(x.omega_dot), TGD1: int16(x.t_gd1), TGD2: int16(x.t_gd2), SVHealth: bool(x.sv_health)}, err
}
func (m *RtcmMessages) GalileoFNavEphemeris(index int) (NativeRTCMGalileoFNavEphemeris, error) {
	if index < 0 {
		return NativeRTCMGalileoFNavEphemeris{}, errNegativeIndex
	}
	var x C.SidereonRtcmGalileoFnavEphemeris
	err := m.resource.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_rtcm_message_galileo_fnav_ephemeris((*C.SidereonRtcmMessages)(p), C.size_t(index), &x)
		})
	})
	return NativeRTCMGalileoFNavEphemeris{SatelliteID: uint8(x.satellite_id), WeekNumber: uint16(x.week_number), IodNav: uint16(x.iod_nav), SISA: uint8(x.sisa), IDOT: int32(x.idot), TOC: uint16(x.t_oc), AF2: int16(x.a_f2), AF1: int32(x.a_f1), AF0: int64(x.a_f0), CRS: int32(x.c_rs), DeltaN: int32(x.delta_n), M0: int64(x.m0), CUC: int32(x.c_uc), Eccentricity: uint64(x.eccentricity), CUS: int32(x.c_us), SqrtA: uint64(x.sqrt_a), TOE: uint16(x.t_oe), CIC: int32(x.c_ic), Omega0: int64(x.omega0), CIS: int32(x.c_is), I0: int64(x.i0), CRC: int32(x.c_rc), Omega: int64(x.omega), OmegaDot: int32(x.omega_dot), BGDE5AE1: int16(x.bgd_e5a_e1), E5ASignalHealth: uint8(x.e5a_signal_health), E5ADataValidity: bool(x.e5a_data_validity), Reserved: uint8(x.reserved)}, err
}
func (m *RtcmMessages) GalileoINavEphemeris(index int) (NativeRTCMGalileoINavEphemeris, error) {
	if index < 0 {
		return NativeRTCMGalileoINavEphemeris{}, errNegativeIndex
	}
	var x C.SidereonRtcmGalileoInavEphemeris
	err := m.resource.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_rtcm_message_galileo_inav_ephemeris((*C.SidereonRtcmMessages)(p), C.size_t(index), &x)
		})
	})
	return NativeRTCMGalileoINavEphemeris{SatelliteID: uint8(x.satellite_id), WeekNumber: uint16(x.week_number), IodNav: uint16(x.iod_nav), SISAIndex: uint8(x.sisa_index), IDOT: int32(x.idot), TOC: uint16(x.t_oc), AF2: int16(x.a_f2), AF1: int32(x.a_f1), AF0: int64(x.a_f0), CRS: int32(x.c_rs), DeltaN: int32(x.delta_n), M0: int64(x.m0), CUC: int32(x.c_uc), Eccentricity: uint64(x.eccentricity), CUS: int32(x.c_us), SqrtA: uint64(x.sqrt_a), TOE: uint16(x.t_oe), CIC: int32(x.c_ic), Omega0: int64(x.omega0), CIS: int32(x.c_is), I0: int64(x.i0), CRC: int32(x.c_rc), Omega: int64(x.omega), OmegaDot: int32(x.omega_dot), BGDE5AE1: int16(x.bgd_e5a_e1), BGDE5BE1: int16(x.bgd_e5b_e1), E5BSignalHealth: uint8(x.e5b_signal_health), E5BDataValidity: bool(x.e5b_data_validity), E1BSignalHealth: uint8(x.e1b_signal_health), E1BDataValidity: bool(x.e1b_data_validity), Reserved: uint8(x.reserved)}, err
}
func (m *RtcmMessages) GLONASSEphemeris(index int) (NativeRTCMGLONASSEphemeris, error) {
	if index < 0 {
		return NativeRTCMGLONASSEphemeris{}, errNegativeIndex
	}
	var x C.SidereonRtcmGlonassEphemeris
	err := m.resource.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_rtcm_message_glonass_ephemeris((*C.SidereonRtcmMessages)(p), C.size_t(index), &x)
		})
	})
	return NativeRTCMGLONASSEphemeris{SatelliteID: uint8(x.satellite_id), FrequencyChannel: uint8(x.frequency_channel), AlmanacHealth: bool(x.almanac_health), AlmanacHealthAvailability: bool(x.almanac_health_availability), P1: uint8(x.p1), TK: uint16(x.t_k), BNMSB: bool(x.b_n_msb), P2: bool(x.p2), TB: uint8(x.t_b), XNDot: int32(x.xn_dot), XN: int32(x.xn), XNDotDot: int8(x.xn_dot_dot), YNDot: int32(x.yn_dot), YN: int32(x.yn), YNDotDot: int8(x.yn_dot_dot), ZNDot: int32(x.zn_dot), ZN: int32(x.zn), ZNDotDot: int8(x.zn_dot_dot), P3: bool(x.p3), GammaN: int16(x.gamma_n), MP: uint8(x.m_p), MLNThird: bool(x.m_l_n_third), TauN: int32(x.tau_n), DeltaTauN: int8(x.delta_tau_n), EN: uint8(x.e_n), MP4: bool(x.m_p4), MFT: uint8(x.m_f_t), MNT: uint16(x.m_n_t), MM: uint8(x.m_m), AdditionalDataAvailable: bool(x.additional_data_available), NA: uint16(x.n_a), TauC: int64(x.tau_c), MN4: uint8(x.m_n4), MTauGPS: int32(x.m_tau_gps), MLNFifth: bool(x.m_l_n_fifth), Reserved: uint8(x.reserved)}, err
}
func (f *RtcmFrames) Count() (int, error) {
	var c C.size_t
	err := f.resource.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 { return C.sidereon_rtcm_frames_count((*C.SidereonRtcmFrames)(p), &c) })
	})
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(c))
}
func (f *RtcmFrames) Len(index int) (int, error) {
	if index < 0 {
		return 0, errNegativeIndex
	}
	var c C.size_t
	err := f.resource.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 { return C.sidereon_rtcm_frame_len((*C.SidereonRtcmFrames)(p), C.size_t(index), &c) })
	})
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(c))
}
func (f *RtcmFrames) Body(index int) ([]byte, error) {
	if index < 0 {
		return nil, errNegativeIndex
	}
	var out []byte
	err := f.resource.with(func(p unsafe.Pointer) error {
		var e error
		out, e = copyByteOutput("RTCM frame body", func(b *C.uint8_t, n C.size_t, w, r *C.size_t) uint32 {
			return C.sidereon_rtcm_frame_body((*C.SidereonRtcmFrames)(p), C.size_t(index), b, n, w, r)
		})
		return e
	})
	return out, err
}
func (d *RtcmDiagnostics) ResyncBytes() (int, error) {
	var c C.size_t
	err := d.resource.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_rtcm_stream_diagnostics_resync_bytes((*C.SidereonRtcmStreamDiagnostics)(p), &c)
		})
	})
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(c))
}
func (d *RtcmDiagnostics) SkippedCount() (int, error) {
	var c C.size_t
	err := d.resource.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_rtcm_stream_diagnostics_skipped_frames_count((*C.SidereonRtcmStreamDiagnostics)(p), &c)
		})
	})
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(c))
}
func (d *RtcmDiagnostics) Skipped(index int) (NativeRTCMFrameSkip, error) {
	if index < 0 {
		return NativeRTCMFrameSkip{}, errNegativeIndex
	}
	var x C.SidereonRtcmFrameSkip
	err := d.resource.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_rtcm_stream_diagnostics_skipped_frame((*C.SidereonRtcmStreamDiagnostics)(p), C.size_t(index), &x)
		})
	})
	if err == nil {
		if skipErr := validateRTCMFrameSkipReasonValue(uint32(x.reason)); skipErr != nil {
			return NativeRTCMFrameSkip{}, skipErr
		}
	}
	v := NativeRTCMFrameSkip{HasMessageNumber: bool(x.has_message_number), MessageNumber: uint16(x.message_number), Reason: uint32(x.reason)}
	if err == nil {
		v.Offset, err = checkedNativeCount(uint64(x.offset))
	}
	return v, err
}
func (d *RtcmDiagnostics) SkippedMessage(index int) ([]byte, error) {
	if index < 0 {
		return nil, errNegativeIndex
	}
	var out []byte
	err := d.resource.with(func(p unsafe.Pointer) error {
		var e error
		out, e = copyByteOutput("RTCM skipped frame message", func(b *C.uint8_t, n C.size_t, w, r *C.size_t) uint32 {
			return C.sidereon_rtcm_stream_diagnostics_skipped_frame_message((*C.SidereonRtcmStreamDiagnostics)(p), C.size_t(index), b, n, w, r)
		})
		return e
	})
	return out, err
}
func NewRTCMTracker() (*RtcmLockTimeTracker, error) {
	var p *C.SidereonRtcmLockTimeTracker
	err := callStatus(func() uint32 { return C.sidereon_rtcm_lock_time_tracker_new(&p) })
	if err != nil {
		if p != nil {
			withCThread(func() { C.sidereon_rtcm_lock_time_tracker_free(p) })
		}
		return nil, err
	}
	handle, err := newRtcmTracker(p)
	if err != nil && p != nil {
		withCThread(func() { C.sidereon_rtcm_lock_time_tracker_free(p) })
	}
	return handle, err
}
func (t *RtcmLockTimeTracker) Reset() error {
	return t.resource.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 { return C.sidereon_rtcm_lock_time_tracker_reset((*C.SidereonRtcmLockTimeTracker)(p)) })
	})
}
func (t *RtcmLockTimeTracker) Observe(m *RtcmMessages, index int) ([]NativeRTCMCellLLI, error) {
	if index < 0 {
		return nil, errNegativeIndex
	}
	if m == nil || m.resource == nil {
		return nil, ErrClosed
	}
	var result []NativeRTCMCellLLI
	err := m.resource.with(func(mp unsafe.Pointer) error {
		return t.resource.with(func(tp unsafe.Pointer) error {
			var w, r C.size_t
			if e := callStatus(func() uint32 {
				return C.sidereon_rtcm_lock_time_tracker_observe((*C.SidereonRtcmLockTimeTracker)(tp), (*C.SidereonRtcmMessages)(mp), C.size_t(index), nil, 0, &w, &r)
			}); e != nil {
				return e
			}
			n, e := validateNativeQuery("RTCM lock-time LLI", uint64(w), uint64(r))
			if e != nil {
				return e
			}
			if _, e := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonRtcmCellLli{})); e != nil {
				return e
			}
			v := make([]C.SidereonRtcmCellLli, n)
			var o *C.SidereonRtcmCellLli
			if n > 0 {
				o = &v[0]
			}
			if e := callStatus(func() uint32 {
				return C.sidereon_rtcm_lock_time_tracker_observe((*C.SidereonRtcmLockTimeTracker)(tp), (*C.SidereonRtcmMessages)(mp), C.size_t(index), o, C.size_t(n), &w, &r)
			}); e != nil {
				return e
			}
			z, e := validateNativeOutput("RTCM lock-time LLI", n, uint64(w), uint64(r))
			if e != nil {
				return e
			}
			result = make([]NativeRTCMCellLLI, z)
			for i := range result {
				result[i] = NativeRTCMCellLLI{SatelliteID: uint8(v[i].satellite_id), SignalID: uint8(v[i].signal_id), LLI: uint8(v[i].lli), HasMinLockTime: bool(v[i].has_min_lock_time_ms), MinLockTimeMS: uint32(v[i].min_lock_time_ms)}
			}
			return nil
		})
	})
	runtime.KeepAlive(t)
	runtime.KeepAlive(m)
	return result, err
}
func MinimumLockTime(kind uint32, indicator uint16) (bool, uint32, error) {
	if err := validateRTCMMSMKindValue(kind); err != nil {
		return false, 0, err
	}
	var p C.bool
	var v C.uint32_t
	err := callStatus(func() uint32 {
		return C.sidereon_rtcm_minimum_lock_time_ms(C.uint32_t(kind), C.uint16_t(indicator), &p, &v)
	})
	return bool(p), uint32(v), err
}
func DeriveLLI(previous *NativeRTCMPreviousLock, currentPresent bool, current uint32, half bool) (uint8, error) {
	var pp *C.SidereonRtcmPreviousLock
	if previous != nil {
		memory, err := checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonRtcmPreviousLock{}))
		if err != nil {
			return 0, err
		}
		defer C.free(memory)
		pp = (*C.SidereonRtcmPreviousLock)(memory)
		*pp = C.SidereonRtcmPreviousLock{has_min_lock_time_ms: C.bool(previous.HasMinLockTime), min_lock_time_ms: C.uint32_t(previous.MinLockTimeMS), elapsed_ms: C.uint64_t(previous.ElapsedMS)}
	}
	var out C.uint8_t
	err := callStatus(func() uint32 {
		return C.sidereon_rtcm_derive_lli(pp, C.bool(currentPresent), C.uint32_t(current), C.bool(half), &out)
	})
	return uint8(out), err
}
func MSMEpochDelta(system, previous, current uint32) (uint64, error) {
	if err := validateGNSSSystemValue(system); err != nil {
		return 0, err
	}
	var out C.uint64_t
	err := callStatus(func() uint32 {
		return C.sidereon_rtcm_msm_epoch_dt_ms(C.uint32_t(system), C.uint32_t(previous), C.uint32_t(current), &out)
	})
	return uint64(out), err
}
func MSMSignalRINEXCode(system uint32, signal uint8) ([]byte, error) {
	if err := validateGNSSSystemValue(system); err != nil {
		return nil, err
	}
	return copyByteOutput("RTCM MSM signal RINEX code", func(b *C.uint8_t, n C.size_t, w, r *C.size_t) uint32 {
		return C.sidereon_rtcm_msm_signal_rinex_code(C.uint32_t(system), C.uint8_t(signal), b, n, w, r)
	})
}
func RTCMLLIBits() (uint8, uint8, error) {
	var a, b C.uint8_t
	err := callStatus(func() uint32 { return C.sidereon_rtcm_lli_bits(&a, &b) })
	return uint8(a), uint8(b), err
}
