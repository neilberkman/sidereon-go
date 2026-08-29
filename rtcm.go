package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// RTCMMessageKind identifies the decoded RTCM message family.
type RTCMMessageKind uint32

const (
	RTCMMultisignalMessage          RTCMMessageKind = RTCMMessageKind(native.RTCMMessageMSMValue)
	RTCMMessageStationCoordinates   RTCMMessageKind = RTCMMessageKind(native.RTCMMessageStationCoordinatesValue)
	RTCMMessageAntennaDescriptor    RTCMMessageKind = RTCMMessageKind(native.RTCMMessageAntennaDescriptorValue)
	RTCMMessageGPSEphemeris         RTCMMessageKind = RTCMMessageKind(native.RTCMMessageGPSEphemerisValue)
	RTCMMessageGLONASSEphemeris     RTCMMessageKind = RTCMMessageKind(native.RTCMMessageGLONASSEphemerisValue)
	RTCMMessageSSR                  RTCMMessageKind = RTCMMessageKind(native.RTCMMessageSSRValue)
	RTCMUnsupportedMessage          RTCMMessageKind = RTCMMessageKind(native.RTCMMessageUnsupportedValue)
	RTCMMessageBeiDouEphemeris      RTCMMessageKind = RTCMMessageKind(native.RTCMMessageBeiDouEphemerisValue)
	RTCMMessageQZSSEphemeris        RTCMMessageKind = RTCMMessageKind(native.RTCMMessageQZSSEphemerisValue)
	RTCMMessageGalileoFNavEphemeris RTCMMessageKind = RTCMMessageKind(native.RTCMMessageGalileoFNavEphemerisValue)
	RTCMMessageGalileoINavEphemeris RTCMMessageKind = RTCMMessageKind(native.RTCMMessageGalileoINavEphemerisValue)
)

// RTCMMSMKind selects an RTCM MSM 4 or MSM 7 payload.
type RTCMMSMKind uint32

const (
	RTCMMSM4 RTCMMSMKind = RTCMMSMKind(native.RTCMMSM4Value)
	RTCMMSM7 RTCMMSMKind = RTCMMSMKind(native.RTCMMSM7Value)
)

// RTCMFrameSkipReason identifies why a stream frame was skipped.
type RTCMFrameSkipReason uint32

const (
	RTCMFrameTruncated RTCMFrameSkipReason = RTCMFrameSkipReason(native.RTCMFrameTruncatedValue)
	RTCMFrameMalformed RTCMFrameSkipReason = RTCMFrameSkipReason(native.RTCMFrameMalformedValue)
)

// RTCMMSMHeader contains raw MSM header fields. EpochTime is the native
// constellation-specific time-of-week representation.
type RTCMMSMHeader struct {
	ReferenceStationID                           uint16
	EpochTime                                    uint32
	MultipleMessage                              bool
	IODS, Reserved, ClockSteering, ExternalClock uint8
	DivergenceFreeSmoothing                      bool
	SmoothingInterval                            uint8
}

// RTCMMSMInfo contains copied MSM metadata and decoded cell dimensions.
type RTCMMSMInfo struct {
	MessageNumber               uint16
	System                      GNSSSystem
	Kind                        RTCMMSMKind
	Header                      RTCMMSMHeader
	SatelliteCount, SignalCount int
}

// RTCMMSMSatellite contains one raw MSM satellite cell. Ranges are in
// milliseconds and rates in metres per second after native scaling.
type RTCMMSMSatellite struct {
	ID, RoughRangeMS       uint8
	RoughRangeMod1         uint16
	HasExtendedInfo        bool
	ExtendedInfo           uint8
	HasRoughPhaseRangeRate bool
	RoughPhaseRangeRateMS  int16
}

// RTCMMSMSignal contains one raw MSM signal cell. CNR is in the native
// quantized representation and lock time is milliseconds.
type RTCMMSMSignal struct {
	SatelliteID, SignalID           uint8
	FinePseudorange, FinePhaseRange int32
	LockTimeIndicator               uint16
	HalfCycleAmbiguity              bool
	CNR                             uint16
	HasFinePhaseRangeRate           bool
	FinePhaseRangeRate              int16
}

// RTCMStationCoordinates contains a copied 1005/1006 station message. ECEF
// integer and floating-point coordinates are metres.
type RTCMStationCoordinates struct {
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

// RTCMAntennaDescriptor contains copied presence flags for an antenna
// descriptor message; strings are retrieved with AntennaString.
type RTCMAntennaDescriptor struct {
	MessageNumber, ReferenceStationID                                                            uint16
	AntennaSetupID                                                                               uint8
	HasAntennaSerialNumber, HasReceiverType, HasReceiverFirmwareVersion, HasReceiverSerialNumber bool
}

// RTCMGPSEphemeris is the lossless raw transmitted-integer payload of a 1019
// GPS broadcast ephemeris. Time fields retain the RTCM wire units.
type RTCMGPSEphemeris struct {
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

// RTCMQZSSEphemeris is the lossless raw transmitted-integer payload of a 1044
// QZSS broadcast ephemeris. Time fields retain the RTCM wire units.
type RTCMQZSSEphemeris struct {
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

// RTCMBeiDouEphemeris is the lossless raw transmitted-integer payload of a
// 1042 BeiDou broadcast ephemeris.
type RTCMBeiDouEphemeris struct {
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

// RTCMGalileoFNavEphemeris is the lossless raw transmitted-integer payload of
// a 1045 Galileo F/NAV broadcast ephemeris.
type RTCMGalileoFNavEphemeris struct {
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

// RTCMGalileoINavEphemeris is the lossless raw transmitted-integer payload of
// a 1046 Galileo I/NAV broadcast ephemeris.
type RTCMGalileoINavEphemeris struct {
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

// RTCMGLONASSEphemeris is the lossless raw transmitted-integer payload of a
// 1020 GLONASS broadcast ephemeris.
type RTCMGLONASSEphemeris struct {
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

// RTCMFrameSkip contains one detached stream skip offset and reason.
type RTCMFrameSkip struct {
	Offset           int
	HasMessageNumber bool
	MessageNumber    uint16
	Reason           RTCMFrameSkipReason
}

// RTCMCellLLI contains one loss-of-lock result from the lock-time tracker.
type RTCMCellLLI struct {
	SatelliteID, SignalID, LLI uint8
	HasMinLockTime             bool
	MinLockTimeMS              uint32
}

// RTCMMessage is a detached decoded message. Body and Frame are copied bytes;
// only the matching typed payload pointer is populated.
type RTCMMessage struct {
	Kind           RTCMMessageKind
	MessageNumber  uint16
	Body, Frame    []byte
	MSM            *RTCMMSMInfo
	Station        *RTCMStationCoordinates
	Antenna        *RTCMAntennaDescriptor
	GPS            *RTCMGPSEphemeris
	GLONASS        *RTCMGLONASSEphemeris
	BeiDou         *RTCMBeiDouEphemeris
	QZSS           *RTCMQZSSEphemeris
	GalileoFNav    *RTCMGalileoFNavEphemeris
	GalileoINav    *RTCMGalileoINavEphemeris
	SSR            *RTCMSSRInfo
	SSROrbits      []RTCMSSROrbitRecord
	SSRClocks      []RTCMSSRClockRecord
	SSRCodeBiases  []RTCMSSRCodeBiasGroup
	SSRPhaseBiases []RTCMSSRPhaseBiasGroup
	SSRURA         []RTCMSSRURARecord
}

// RTCMMessages owns a C-backed decoded message collection.
type RTCMMessages struct {
	_      noCopy
	handle *native.RtcmMessages
}

// RTCMFrames owns a C-backed scanned frame collection.
type RTCMFrames struct {
	_      noCopy
	handle *native.RtcmFrames
}

// RTCMStreamDiagnostics owns copied stream-resynchronization diagnostics.
type RTCMStreamDiagnostics struct {
	_      noCopy
	handle *native.RtcmDiagnostics
}

// DecodeRTCM decodes RTCM bytes through the native C parser.
func DecodeRTCM(data []byte) (*RTCMMessages, error) {
	h, e := native.DecodeRTCM(data)
	if e != nil {
		return nil, publicError(e)
	}
	return &RTCMMessages{handle: h}, nil
}

// DecodeRTCMStream decodes recoverable RTCM stream bytes and returns owning
// messages and diagnostics handles.
func DecodeRTCMStream(data []byte) (*RTCMMessages, *RTCMStreamDiagnostics, error) {
	m, d, e := native.DecodeRTCMStream(data)
	if e != nil {
		return nil, nil, publicError(e)
	}
	return &RTCMMessages{handle: m}, &RTCMStreamDiagnostics{handle: d}, nil
}

// ScanRTCMFrames scans transport frames without decoding message payloads.
func ScanRTCMFrames(data []byte) (*RTCMFrames, error) {
	h, e := native.ScanRTCMFrames(data)
	if e != nil {
		return nil, publicError(e)
	}
	return &RTCMFrames{handle: h}, nil
}

// DecodeRTCMFrame decodes the first complete RTCM 3 frame, returning a copied
// message body and the total transport-frame length.
func DecodeRTCMFrame(data []byte) ([]byte, int, error) {
	body, length, err := native.DecodeRTCMFrame(data)
	return body, length, publicError(err)
}

// EncodeRTCMFrame returns a detached RTCM 3 transport frame for a body.
func EncodeRTCMFrame(body []byte) ([]byte, error) {
	v, e := native.EncodeRTCMFrame(body)
	return v, publicError(e)
}

// BuildRTCMAntennaDescriptor constructs a 1007, 1008, or 1033 message. A
// nil optional string preserves the C ABI's absent-field distinction.
func BuildRTCMAntennaDescriptor(messageNumber, referenceStationID uint16, setup uint8, descriptor string, serial, receiverType, firmware, receiverSerial *string) (*RTCMMessages, error) {
	h, err := native.BuildRTCMAntennaDescriptor(messageNumber, referenceStationID, setup, descriptor, serial, receiverType, firmware, receiverSerial)
	if err != nil {
		return nil, publicError(err)
	}
	return &RTCMMessages{handle: h}, nil
}

// BuildRTCMMSM builds a detached RTCM MSM message from copied cell arrays.
func BuildRTCMMSM(info RTCMMSMInfo, satellites []RTCMMSMSatellite, signals []RTCMMSMSignal) (*RTCMMessages, error) {
	if err := validateGNSSSystem(info.System); err != nil {
		return nil, err
	}
	if err := validateRTCMMSMKind(info.Kind); err != nil {
		return nil, err
	}
	nativeSatellites := make([]native.NativeRTCMMSMSatellite, len(satellites))
	for i, value := range satellites {
		nativeSatellites[i] = native.NativeRTCMMSMSatellite{ID: value.ID, RoughRangeMS: value.RoughRangeMS, RoughRangeMod1: value.RoughRangeMod1, HasExtendedInfo: value.HasExtendedInfo, ExtendedInfo: value.ExtendedInfo, HasRoughPhaseRangeRate: value.HasRoughPhaseRangeRate, RoughPhaseRangeRateMS: value.RoughPhaseRangeRateMS}
	}
	nativeSignals := make([]native.NativeRTCMMSMSignal, len(signals))
	for i, value := range signals {
		nativeSignals[i] = native.NativeRTCMMSMSignal{SatelliteID: value.SatelliteID, SignalID: value.SignalID, FinePseudorange: value.FinePseudorange, FinePhaseRange: value.FinePhaseRange, LockTimeIndicator: value.LockTimeIndicator, HalfCycleAmbiguity: value.HalfCycleAmbiguity, CNR: value.CNR, HasFinePhaseRangeRate: value.HasFinePhaseRangeRate, FinePhaseRangeRate: value.FinePhaseRangeRate}
	}
	nativeInfo := native.NativeRTCMMSMInfo{MessageNumber: info.MessageNumber, System: uint32(info.System), Kind: uint32(info.Kind), Header: native.NativeRTCMMSMHeader{ReferenceStationID: info.Header.ReferenceStationID, EpochTime: info.Header.EpochTime, MultipleMessage: info.Header.MultipleMessage, IODS: info.Header.IODS, Reserved: info.Header.Reserved, ClockSteering: info.Header.ClockSteering, ExternalClock: info.Header.ExternalClock, DivergenceFreeSmoothing: info.Header.DivergenceFreeSmoothing, SmoothingInterval: info.Header.SmoothingInterval}, SatelliteCount: len(satellites), SignalCount: len(signals)}
	h, err := native.BuildRTCMMSM(nativeInfo, nativeSatellites, nativeSignals)
	if err != nil {
		return nil, publicError(err)
	}
	return &RTCMMessages{handle: h}, nil
}

// BuildRTCMStationCoordinates builds a detached RTCM station-coordinate
// message with ECEF coordinates in metres.
func BuildRTCMStationCoordinates(value RTCMStationCoordinates) (*RTCMMessages, error) {
	h, err := native.BuildRTCMStation(native.NativeRTCMStationCoordinates{MessageNumber: value.MessageNumber, ReferenceStationID: value.ReferenceStationID, ITRFRealizationYear: value.ITRFRealizationYear, GPS: value.GPS, GLONASS: value.GLONASS, Galileo: value.Galileo, ReferenceStation: value.ReferenceStation, SingleReceiverOscillator: value.SingleReceiverOscillator, Reserved: value.Reserved, QuarterCycleIndicator: value.QuarterCycleIndicator, ECEFX: value.ECEFX, ECEFY: value.ECEFY, ECEFZ: value.ECEFZ, XM: value.XM, YM: value.YM, ZM: value.ZM, HasAntennaHeight: value.HasAntennaHeight, AntennaHeight: value.AntennaHeight, AntennaHeightM: value.AntennaHeightM})
	if err != nil {
		return nil, publicError(err)
	}
	return &RTCMMessages{handle: h}, nil
}

// BuildRTCMGPSEphemeris builds a detached GPS 1019 message.
func BuildRTCMGPSEphemeris(value RTCMGPSEphemeris) (*RTCMMessages, error) {
	h, err := native.BuildRTCMGPSEphemeris(native.NativeRTCMGPSEphemeris{SatelliteID: value.SatelliteID, WeekNumber: value.WeekNumber, SVAccuracy: value.SVAccuracy, CodeOnL2: value.CodeOnL2, IDOT: value.IDOT, IODE: value.IODE, TOC: value.TOC, AF2: value.AF2, AF1: value.AF1, AF0: value.AF0, IODC: value.IODC, CRS: value.CRS, DeltaN: value.DeltaN, M0: value.M0, CUC: value.CUC, Eccentricity: value.Eccentricity, CUS: value.CUS, SqrtA: value.SqrtA, TOE: value.TOE, CIC: value.CIC, Omega0: value.Omega0, CIS: value.CIS, I0: value.I0, CRC: value.CRC, Omega: value.Omega, OmegaDot: value.OmegaDot, TGD: value.TGD, SVHealth: value.SVHealth, L2PDataFlag: value.L2PDataFlag, FitInterval: value.FitInterval})
	if err != nil {
		return nil, publicError(err)
	}
	return &RTCMMessages{handle: h}, nil
}

// BuildRTCMQZSSEphemeris builds a detached QZSS 1044 message.
func BuildRTCMQZSSEphemeris(value RTCMQZSSEphemeris) (*RTCMMessages, error) {
	h, err := native.BuildRTCMQZSSEphemeris(native.NativeRTCMQZSSEphemeris{SatelliteID: value.SatelliteID, TOC: value.TOC, AF2: value.AF2, AF1: value.AF1, AF0: value.AF0, IODE: value.IODE, CRS: value.CRS, DeltaN: value.DeltaN, M0: value.M0, CUC: value.CUC, Eccentricity: value.Eccentricity, CUS: value.CUS, SqrtA: value.SqrtA, TOE: value.TOE, CIC: value.CIC, Omega0: value.Omega0, CIS: value.CIS, I0: value.I0, CRC: value.CRC, Omega: value.Omega, OmegaDot: value.OmegaDot, IDOT: value.IDOT, CodesOnL2: value.CodesOnL2, WeekNumber: value.WeekNumber, URA: value.URA, SVHealth: value.SVHealth, TGD: value.TGD, IODC: value.IODC, FitInterval: value.FitInterval})
	if err != nil {
		return nil, publicError(err)
	}
	return &RTCMMessages{handle: h}, nil
}

// BuildRTCMBeiDouEphemeris builds a detached BeiDou 1042 message.
func BuildRTCMBeiDouEphemeris(value RTCMBeiDouEphemeris) (*RTCMMessages, error) {
	h, err := native.BuildRTCMBeiDouEphemeris(native.NativeRTCMBeidouEphemeris{SatelliteID: value.SatelliteID, WeekNumber: value.WeekNumber, SVURAI: value.SVURAI, IDOT: value.IDOT, AODE: value.AODE, TOC: value.TOC, AF2: value.AF2, AF1: value.AF1, AF0: value.AF0, AODC: value.AODC, CRS: value.CRS, DeltaN: value.DeltaN, M0: value.M0, CUC: value.CUC, Eccentricity: value.Eccentricity, CUS: value.CUS, SqrtA: value.SqrtA, TOE: value.TOE, CIC: value.CIC, Omega0: value.Omega0, CIS: value.CIS, I0: value.I0, CRC: value.CRC, Omega: value.Omega, OmegaDot: value.OmegaDot, TGD1: value.TGD1, TGD2: value.TGD2, SVHealth: value.SVHealth})
	if err != nil {
		return nil, publicError(err)
	}
	return &RTCMMessages{handle: h}, nil
}

// BuildRTCMGalileoFNavEphemeris builds a detached Galileo F/NAV 1045 message.
func BuildRTCMGalileoFNavEphemeris(value RTCMGalileoFNavEphemeris) (*RTCMMessages, error) {
	h, err := native.BuildRTCMGalileoFNavEphemeris(native.NativeRTCMGalileoFNavEphemeris{SatelliteID: value.SatelliteID, WeekNumber: value.WeekNumber, IodNav: value.IodNav, SISA: value.SISA, IDOT: value.IDOT, TOC: value.TOC, AF2: value.AF2, AF1: value.AF1, AF0: value.AF0, CRS: value.CRS, DeltaN: value.DeltaN, M0: value.M0, CUC: value.CUC, Eccentricity: value.Eccentricity, CUS: value.CUS, SqrtA: value.SqrtA, TOE: value.TOE, CIC: value.CIC, Omega0: value.Omega0, CIS: value.CIS, I0: value.I0, CRC: value.CRC, Omega: value.Omega, OmegaDot: value.OmegaDot, BGDE5AE1: value.BGDE5AE1, E5ASignalHealth: value.E5ASignalHealth, E5ADataValidity: value.E5ADataValidity, Reserved: value.Reserved})
	if err != nil {
		return nil, publicError(err)
	}
	return &RTCMMessages{handle: h}, nil
}

// BuildRTCMGalileoINavEphemeris builds a detached Galileo I/NAV 1046 message.
func BuildRTCMGalileoINavEphemeris(value RTCMGalileoINavEphemeris) (*RTCMMessages, error) {
	h, err := native.BuildRTCMGalileoINavEphemeris(native.NativeRTCMGalileoINavEphemeris{SatelliteID: value.SatelliteID, WeekNumber: value.WeekNumber, IodNav: value.IodNav, SISAIndex: value.SISAIndex, IDOT: value.IDOT, TOC: value.TOC, AF2: value.AF2, AF1: value.AF1, AF0: value.AF0, CRS: value.CRS, DeltaN: value.DeltaN, M0: value.M0, CUC: value.CUC, Eccentricity: value.Eccentricity, CUS: value.CUS, SqrtA: value.SqrtA, TOE: value.TOE, CIC: value.CIC, Omega0: value.Omega0, CIS: value.CIS, I0: value.I0, CRC: value.CRC, Omega: value.Omega, OmegaDot: value.OmegaDot, BGDE5AE1: value.BGDE5AE1, BGDE5BE1: value.BGDE5BE1, E5BSignalHealth: value.E5BSignalHealth, E5BDataValidity: value.E5BDataValidity, E1BSignalHealth: value.E1BSignalHealth, E1BDataValidity: value.E1BDataValidity, Reserved: value.Reserved})
	if err != nil {
		return nil, publicError(err)
	}
	return &RTCMMessages{handle: h}, nil
}

// BuildRTCMGLONASSEphemeris builds a detached GLONASS 1020 message.
func BuildRTCMGLONASSEphemeris(value RTCMGLONASSEphemeris) (*RTCMMessages, error) {
	h, err := native.BuildRTCMGLONASSEphemeris(native.NativeRTCMGLONASSEphemeris{SatelliteID: value.SatelliteID, FrequencyChannel: value.FrequencyChannel, AlmanacHealth: value.AlmanacHealth, AlmanacHealthAvailability: value.AlmanacHealthAvailability, P1: value.P1, TK: value.TK, BNMSB: value.BNMSB, P2: value.P2, TB: value.TB, XNDot: value.XNDot, XN: value.XN, XNDotDot: value.XNDotDot, YNDot: value.YNDot, YN: value.YN, YNDotDot: value.YNDotDot, ZNDot: value.ZNDot, ZN: value.ZN, ZNDotDot: value.ZNDotDot, P3: value.P3, GammaN: value.GammaN, MP: value.MP, MLNThird: value.MLNThird, TauN: value.TauN, DeltaTauN: value.DeltaTauN, EN: value.EN, MP4: value.MP4, MFT: value.MFT, MNT: value.MNT, MM: value.MM, AdditionalDataAvailable: value.AdditionalDataAvailable, NA: value.NA, TauC: value.TauC, MN4: value.MN4, MTauGPS: value.MTauGPS, MLNFifth: value.MLNFifth, Reserved: value.Reserved})
	if err != nil {
		return nil, publicError(err)
	}
	return &RTCMMessages{handle: h}, nil
}

// Close releases the message collection; repeated calls are safe.
func (m *RTCMMessages) Close() error {
	if m == nil || m.handle == nil {
		return nil
	}
	return publicError(m.handle.Close())
}

// Count returns the number of decoded messages.
func (m *RTCMMessages) Count() (int, error) {
	if m == nil || m.handle == nil {
		return 0, ErrClosed
	}
	v, e := m.handle.Count()
	return v, publicError(e)
}

// Message returns one detached decoded message by zero-based index.
func (m *RTCMMessages) Message(index int) (RTCMMessage, error) {
	if m == nil || m.handle == nil {
		return RTCMMessage{}, ErrClosed
	}
	info, e := m.handle.Kind(index)
	if e != nil {
		return RTCMMessage{}, publicError(e)
	}
	body, e := m.handle.Encode(index)
	if e != nil {
		return RTCMMessage{}, publicError(e)
	}
	frame, e := m.handle.Frame(index)
	if e != nil {
		return RTCMMessage{}, publicError(e)
	}
	out := RTCMMessage{Kind: RTCMMessageKind(info.Kind), MessageNumber: info.MessageNumber, Body: append([]byte(nil), body...), Frame: append([]byte(nil), frame...)}
	switch out.Kind {
	case RTCMMultisignalMessage:
		v, e := m.handle.MSMInfo(index)
		if e != nil {
			return RTCMMessage{}, publicError(e)
		}
		x := rtcmMSMInfo(v)
		out.MSM = &x
	case RTCMMessageStationCoordinates:
		v, e := m.handle.Station(index)
		if e != nil {
			return RTCMMessage{}, publicError(e)
		}
		x := rtcmStation(v)
		out.Station = &x
	case RTCMMessageAntennaDescriptor:
		v, e := m.handle.Antenna(index)
		if e != nil {
			return RTCMMessage{}, publicError(e)
		}
		x := rtcmAntenna(v)
		out.Antenna = &x
	case RTCMMessageSSR:
		info, err := m.handle.SSRInfo(index)
		if err != nil {
			return RTCMMessage{}, publicError(err)
		}
		ssr := ssrInfoFromNative(info)
		out.SSR = &ssr
		if info.OrbitCount > 0 {
			values, err := m.SSROrbits(index)
			if err != nil {
				return RTCMMessage{}, err
			}
			out.SSROrbits = values
		}
		if info.ClockCount > 0 {
			values, err := m.SSRClocks(index)
			if err != nil {
				return RTCMMessage{}, err
			}
			out.SSRClocks = values
		}
		if info.CodeBiasCount > 0 {
			values, err := m.SSRCodeBiases(index)
			if err != nil {
				return RTCMMessage{}, err
			}
			out.SSRCodeBiases = values
		}
		if info.PhaseBiasCount > 0 {
			values, err := m.SSRPhaseBiases(index)
			if err != nil {
				return RTCMMessage{}, err
			}
			out.SSRPhaseBiases = values
		}
		if info.URACount > 0 {
			values, err := m.SSRURA(index)
			if err != nil {
				return RTCMMessage{}, err
			}
			out.SSRURA = values
		}
	case RTCMMessageGPSEphemeris:
		value, err := m.handle.GPSEphemeris(index)
		if err != nil {
			return RTCMMessage{}, publicError(err)
		}
		converted := rtcmGPS(value)
		out.GPS = &converted
	case RTCMMessageGLONASSEphemeris:
		value, err := m.handle.GLONASSEphemeris(index)
		if err != nil {
			return RTCMMessage{}, publicError(err)
		}
		converted := rtcmGLONASS(value)
		out.GLONASS = &converted
	case RTCMMessageBeiDouEphemeris:
		value, err := m.handle.BeidouEphemeris(index)
		if err != nil {
			return RTCMMessage{}, publicError(err)
		}
		converted := rtcmBeiDou(value)
		out.BeiDou = &converted
	case RTCMMessageQZSSEphemeris:
		value, err := m.handle.QZSSEphemeris(index)
		if err != nil {
			return RTCMMessage{}, publicError(err)
		}
		converted := rtcmQZSS(value)
		out.QZSS = &converted
	case RTCMMessageGalileoFNavEphemeris:
		value, err := m.handle.GalileoFNavEphemeris(index)
		if err != nil {
			return RTCMMessage{}, publicError(err)
		}
		converted := rtcmGalileoFNav(value)
		out.GalileoFNav = &converted
	case RTCMMessageGalileoINavEphemeris:
		value, err := m.handle.GalileoINavEphemeris(index)
		if err != nil {
			return RTCMMessage{}, publicError(err)
		}
		converted := rtcmGalileoINav(value)
		out.GalileoINav = &converted
	}
	return out, nil
}

func (m *RTCMMessages) SSRInfo(index int) (RTCMSSRInfo, error) {
	if m == nil || m.handle == nil {
		return RTCMSSRInfo{}, ErrClosed
	}
	v, err := m.handle.SSRInfo(index)
	return ssrInfoFromNative(v), publicError(err)
}

func (m *RTCMMessages) SSROrbits(index int) ([]RTCMSSROrbitRecord, error) {
	if m == nil || m.handle == nil {
		return nil, ErrClosed
	}
	values, err := m.handle.SSROrbitRecords(index)
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]RTCMSSROrbitRecord, len(values))
	for i, value := range values {
		out[i] = RTCMSSROrbitRecord{SatelliteID: value.SatelliteID, IODE: value.IODE, DeltaRadial: value.DeltaRadial, DeltaAlong: value.DeltaAlong, DeltaCross: value.DeltaCross, DotDeltaRadial: value.DotDeltaRadial, DotDeltaAlong: value.DotDeltaAlong, DotDeltaCross: value.DotDeltaCross}
	}
	return out, nil
}

func (m *RTCMMessages) SSRClocks(index int) ([]RTCMSSRClockRecord, error) {
	if m == nil || m.handle == nil {
		return nil, ErrClosed
	}
	values, err := m.handle.SSRClockRecords(index)
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]RTCMSSRClockRecord, len(values))
	for i, value := range values {
		out[i] = RTCMSSRClockRecord{SatelliteID: value.SatelliteID, C0: value.C0, C1: value.C1, C2: value.C2}
	}
	return out, nil
}

func (m *RTCMMessages) SSRCodeBiases(index int) ([]RTCMSSRCodeBiasGroup, error) {
	if m == nil || m.handle == nil {
		return nil, ErrClosed
	}
	records, err := m.handle.SSRCodeBiasRecords(index)
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]RTCMSSRCodeBiasGroup, len(records))
	for i, record := range records {
		out[i].Record = RTCMSSRCodeBiasRecord{SatelliteID: record.SatelliteID, SignalCount: record.SignalCount}
		if record.SignalCount == 0 {
			continue
		}
		values, err := m.handle.SSRCodeBiasSignals(index, i)
		if err != nil {
			return nil, publicError(err)
		}
		out[i].Signals = make([]RTCMSSRCodeBiasSignal, len(values))
		for j, value := range values {
			out[i].Signals[j] = RTCMSSRCodeBiasSignal{SignalID: value.SignalID, Bias: value.Bias}
		}
	}
	return out, nil
}

func (m *RTCMMessages) SSRCodeBiasSignals(index, recordIndex int) ([]RTCMSSRCodeBiasSignal, error) {
	if m == nil || m.handle == nil {
		return nil, ErrClosed
	}
	values, err := m.handle.SSRCodeBiasSignals(index, recordIndex)
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]RTCMSSRCodeBiasSignal, len(values))
	for i, value := range values {
		out[i] = RTCMSSRCodeBiasSignal{SignalID: value.SignalID, Bias: value.Bias}
	}
	return out, nil
}

func (m *RTCMMessages) SSRPhaseBiases(index int) ([]RTCMSSRPhaseBiasGroup, error) {
	if m == nil || m.handle == nil {
		return nil, ErrClosed
	}
	records, err := m.handle.SSRPhaseBiasRecords(index)
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]RTCMSSRPhaseBiasGroup, len(records))
	for i, record := range records {
		out[i].Record = RTCMSSRPhaseBiasRecord{SatelliteID: record.SatelliteID, YawAngle: record.YawAngle, YawRate: record.YawRate, SignalCount: record.SignalCount}
		if record.SignalCount == 0 {
			continue
		}
		values, err := m.handle.SSRPhaseBiasSignals(index, i)
		if err != nil {
			return nil, publicError(err)
		}
		out[i].Signals = make([]RTCMSSRPhaseBiasSignal, len(values))
		for j, value := range values {
			out[i].Signals[j] = RTCMSSRPhaseBiasSignal{SignalID: value.SignalID, IntegerIndicator: value.IntegerIndicator, WideLaneIntegerIndicator: value.WideLaneIntegerIndicator, DiscontinuityCounter: value.DiscontinuityCounter, Bias: value.Bias}
		}
	}
	return out, nil
}

func (m *RTCMMessages) SSRPhaseBiasSignals(index, recordIndex int) ([]RTCMSSRPhaseBiasSignal, error) {
	if m == nil || m.handle == nil {
		return nil, ErrClosed
	}
	values, err := m.handle.SSRPhaseBiasSignals(index, recordIndex)
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]RTCMSSRPhaseBiasSignal, len(values))
	for i, value := range values {
		out[i] = RTCMSSRPhaseBiasSignal{SignalID: value.SignalID, IntegerIndicator: value.IntegerIndicator, WideLaneIntegerIndicator: value.WideLaneIntegerIndicator, DiscontinuityCounter: value.DiscontinuityCounter, Bias: value.Bias}
	}
	return out, nil
}

func (m *RTCMMessages) SSRURA(index int) ([]RTCMSSRURARecord, error) {
	if m == nil || m.handle == nil {
		return nil, ErrClosed
	}
	values, err := m.handle.SSRURARecords(index)
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]RTCMSSRURARecord, len(values))
	for i, value := range values {
		out[i] = RTCMSSRURARecord{SatelliteID: value.SatelliteID, URAIndex: value.URAIndex}
	}
	return out, nil
}
func (m *RTCMMessages) Encode(index int) ([]byte, error) {
	if m == nil || m.handle == nil {
		return nil, ErrClosed
	}
	v, e := m.handle.Encode(index)
	return v, publicError(e)
}
func (m *RTCMMessages) Frame(index int) ([]byte, error) {
	if m == nil || m.handle == nil {
		return nil, ErrClosed
	}
	v, e := m.handle.Frame(index)
	return v, publicError(e)
}
func (m *RTCMMessages) MSMInfo(index int) (RTCMMSMInfo, error) {
	if m == nil || m.handle == nil {
		return RTCMMSMInfo{}, ErrClosed
	}
	v, e := m.handle.MSMInfo(index)
	return rtcmMSMInfo(v), publicError(e)
}
func (m *RTCMMessages) MSMSatellites(index int) ([]RTCMMSMSatellite, error) {
	if m == nil || m.handle == nil {
		return nil, ErrClosed
	}
	v, e := m.handle.MSMSatellites(index)
	if e != nil {
		return nil, publicError(e)
	}
	out := make([]RTCMMSMSatellite, len(v))
	for i, x := range v {
		out[i] = RTCMMSMSatellite{ID: x.ID, RoughRangeMS: x.RoughRangeMS, RoughRangeMod1: x.RoughRangeMod1, HasExtendedInfo: x.HasExtendedInfo, ExtendedInfo: x.ExtendedInfo, HasRoughPhaseRangeRate: x.HasRoughPhaseRangeRate, RoughPhaseRangeRateMS: x.RoughPhaseRangeRateMS}
	}
	return out, nil
}
func (m *RTCMMessages) MSMSignals(index int) ([]RTCMMSMSignal, error) {
	if m == nil || m.handle == nil {
		return nil, ErrClosed
	}
	v, e := m.handle.MSMSignals(index)
	if e != nil {
		return nil, publicError(e)
	}
	out := make([]RTCMMSMSignal, len(v))
	for i, x := range v {
		out[i] = RTCMMSMSignal{SatelliteID: x.SatelliteID, SignalID: x.SignalID, FinePseudorange: x.FinePseudorange, FinePhaseRange: x.FinePhaseRange, LockTimeIndicator: x.LockTimeIndicator, HalfCycleAmbiguity: x.HalfCycleAmbiguity, CNR: x.CNR, HasFinePhaseRangeRate: x.HasFinePhaseRangeRate, FinePhaseRangeRate: x.FinePhaseRangeRate}
	}
	return out, nil
}
func (m *RTCMMessages) StationCoordinates(index int) (RTCMStationCoordinates, error) {
	if m == nil || m.handle == nil {
		return RTCMStationCoordinates{}, ErrClosed
	}
	v, e := m.handle.Station(index)
	return rtcmStation(v), publicError(e)
}

func (m *RTCMMessages) GPSEphemeris(index int) (RTCMGPSEphemeris, error) {
	if m == nil || m.handle == nil {
		return RTCMGPSEphemeris{}, ErrClosed
	}
	v, err := m.handle.GPSEphemeris(index)
	return rtcmGPS(v), publicError(err)
}
func (m *RTCMMessages) GLONASSEphemeris(index int) (RTCMGLONASSEphemeris, error) {
	if m == nil || m.handle == nil {
		return RTCMGLONASSEphemeris{}, ErrClosed
	}
	v, err := m.handle.GLONASSEphemeris(index)
	return rtcmGLONASS(v), publicError(err)
}
func (m *RTCMMessages) BeiDouEphemeris(index int) (RTCMBeiDouEphemeris, error) {
	if m == nil || m.handle == nil {
		return RTCMBeiDouEphemeris{}, ErrClosed
	}
	v, err := m.handle.BeidouEphemeris(index)
	return rtcmBeiDou(v), publicError(err)
}
func (m *RTCMMessages) QZSSEphemeris(index int) (RTCMQZSSEphemeris, error) {
	if m == nil || m.handle == nil {
		return RTCMQZSSEphemeris{}, ErrClosed
	}
	v, err := m.handle.QZSSEphemeris(index)
	return rtcmQZSS(v), publicError(err)
}
func (m *RTCMMessages) GalileoFNavEphemeris(index int) (RTCMGalileoFNavEphemeris, error) {
	if m == nil || m.handle == nil {
		return RTCMGalileoFNavEphemeris{}, ErrClosed
	}
	v, err := m.handle.GalileoFNavEphemeris(index)
	return rtcmGalileoFNav(v), publicError(err)
}
func (m *RTCMMessages) GalileoINavEphemeris(index int) (RTCMGalileoINavEphemeris, error) {
	if m == nil || m.handle == nil {
		return RTCMGalileoINavEphemeris{}, ErrClosed
	}
	v, err := m.handle.GalileoINavEphemeris(index)
	return rtcmGalileoINav(v), publicError(err)
}

// RTCMAntennaStringField selects one optional antenna/receiver string field.
type RTCMAntennaStringField uint32

const (
	RTCMAntennaDescriptorField       RTCMAntennaStringField = RTCMAntennaStringField(native.RTCMAntennaDescriptorFieldValue)
	RTCMAntennaSerialNumberField     RTCMAntennaStringField = RTCMAntennaStringField(native.RTCMAntennaSerialNumberFieldValue)
	RTCMReceiverTypeField            RTCMAntennaStringField = RTCMAntennaStringField(native.RTCMReceiverTypeFieldValue)
	RTCMReceiverFirmwareVersionField RTCMAntennaStringField = RTCMAntennaStringField(native.RTCMReceiverFirmwareVersionFieldValue)
	RTCMReceiverSerialNumberField    RTCMAntennaStringField = RTCMAntennaStringField(native.RTCMReceiverSerialNumberFieldValue)
)

// AntennaDescriptor returns copied antenna descriptor presence metadata.
func (m *RTCMMessages) AntennaDescriptor(index int) (RTCMAntennaDescriptor, error) {
	if m == nil || m.handle == nil {
		return RTCMAntennaDescriptor{}, ErrClosed
	}
	v, e := m.handle.Antenna(index)
	return rtcmAntenna(v), publicError(e)
}

// AntennaString returns one selected optional antenna/receiver string.
func (m *RTCMMessages) AntennaString(index int, field RTCMAntennaStringField) (string, error) {
	if m == nil || m.handle == nil {
		return "", ErrClosed
	}
	if err := validateRTCMStringField(field); err != nil {
		return "", err
	}
	v, e := m.handle.AntennaString(index, uint32(field))
	return string(v), publicError(e)
}
func rtcmMSMInfo(v native.NativeRTCMMSMInfo) RTCMMSMInfo {
	return RTCMMSMInfo{MessageNumber: v.MessageNumber, System: GNSSSystem(v.System), Kind: RTCMMSMKind(v.Kind), Header: RTCMMSMHeader{ReferenceStationID: v.Header.ReferenceStationID, EpochTime: v.Header.EpochTime, MultipleMessage: v.Header.MultipleMessage, IODS: v.Header.IODS, Reserved: v.Header.Reserved, ClockSteering: v.Header.ClockSteering, ExternalClock: v.Header.ExternalClock, DivergenceFreeSmoothing: v.Header.DivergenceFreeSmoothing, SmoothingInterval: v.Header.SmoothingInterval}, SatelliteCount: v.SatelliteCount, SignalCount: v.SignalCount}
}
func rtcmStation(v native.NativeRTCMStationCoordinates) RTCMStationCoordinates {
	return RTCMStationCoordinates{MessageNumber: v.MessageNumber, ReferenceStationID: v.ReferenceStationID, ITRFRealizationYear: v.ITRFRealizationYear, GPS: v.GPS, GLONASS: v.GLONASS, Galileo: v.Galileo, ReferenceStation: v.ReferenceStation, SingleReceiverOscillator: v.SingleReceiverOscillator, Reserved: v.Reserved, QuarterCycleIndicator: v.QuarterCycleIndicator, ECEFX: v.ECEFX, ECEFY: v.ECEFY, ECEFZ: v.ECEFZ, XM: v.XM, YM: v.YM, ZM: v.ZM, HasAntennaHeight: v.HasAntennaHeight, AntennaHeight: v.AntennaHeight, AntennaHeightM: v.AntennaHeightM}
}
func rtcmAntenna(v native.NativeRTCMAntennaDescriptor) RTCMAntennaDescriptor {
	return RTCMAntennaDescriptor{MessageNumber: v.MessageNumber, ReferenceStationID: v.ReferenceStationID, AntennaSetupID: v.AntennaSetupID, HasAntennaSerialNumber: v.HasAntennaSerialNumber, HasReceiverType: v.HasReceiverType, HasReceiverFirmwareVersion: v.HasReceiverFirmwareVersion, HasReceiverSerialNumber: v.HasReceiverSerialNumber}
}

func rtcmGPS(v native.NativeRTCMGPSEphemeris) RTCMGPSEphemeris {
	return RTCMGPSEphemeris{SatelliteID: v.SatelliteID, WeekNumber: v.WeekNumber, SVAccuracy: v.SVAccuracy, CodeOnL2: v.CodeOnL2, IDOT: v.IDOT, IODE: v.IODE, TOC: v.TOC, AF2: v.AF2, AF1: v.AF1, AF0: v.AF0, IODC: v.IODC, CRS: v.CRS, DeltaN: v.DeltaN, M0: v.M0, CUC: v.CUC, Eccentricity: v.Eccentricity, CUS: v.CUS, SqrtA: v.SqrtA, TOE: v.TOE, CIC: v.CIC, Omega0: v.Omega0, CIS: v.CIS, I0: v.I0, CRC: v.CRC, Omega: v.Omega, OmegaDot: v.OmegaDot, TGD: v.TGD, SVHealth: v.SVHealth, L2PDataFlag: v.L2PDataFlag, FitInterval: v.FitInterval}
}
func rtcmQZSS(v native.NativeRTCMQZSSEphemeris) RTCMQZSSEphemeris {
	return RTCMQZSSEphemeris{SatelliteID: v.SatelliteID, TOC: v.TOC, AF2: v.AF2, AF1: v.AF1, AF0: v.AF0, IODE: v.IODE, CRS: v.CRS, DeltaN: v.DeltaN, M0: v.M0, CUC: v.CUC, Eccentricity: v.Eccentricity, CUS: v.CUS, SqrtA: v.SqrtA, TOE: v.TOE, CIC: v.CIC, Omega0: v.Omega0, CIS: v.CIS, I0: v.I0, CRC: v.CRC, Omega: v.Omega, OmegaDot: v.OmegaDot, IDOT: v.IDOT, CodesOnL2: v.CodesOnL2, WeekNumber: v.WeekNumber, URA: v.URA, SVHealth: v.SVHealth, TGD: v.TGD, IODC: v.IODC, FitInterval: v.FitInterval}
}
func rtcmBeiDou(v native.NativeRTCMBeidouEphemeris) RTCMBeiDouEphemeris {
	return RTCMBeiDouEphemeris{SatelliteID: v.SatelliteID, WeekNumber: v.WeekNumber, SVURAI: v.SVURAI, IDOT: v.IDOT, AODE: v.AODE, TOC: v.TOC, AF2: v.AF2, AF1: v.AF1, AF0: v.AF0, AODC: v.AODC, CRS: v.CRS, DeltaN: v.DeltaN, M0: v.M0, CUC: v.CUC, Eccentricity: v.Eccentricity, CUS: v.CUS, SqrtA: v.SqrtA, TOE: v.TOE, CIC: v.CIC, Omega0: v.Omega0, CIS: v.CIS, I0: v.I0, CRC: v.CRC, Omega: v.Omega, OmegaDot: v.OmegaDot, TGD1: v.TGD1, TGD2: v.TGD2, SVHealth: v.SVHealth}
}
func rtcmGalileoFNav(v native.NativeRTCMGalileoFNavEphemeris) RTCMGalileoFNavEphemeris {
	return RTCMGalileoFNavEphemeris{SatelliteID: v.SatelliteID, WeekNumber: v.WeekNumber, IodNav: v.IodNav, SISA: v.SISA, IDOT: v.IDOT, TOC: v.TOC, AF2: v.AF2, AF1: v.AF1, AF0: v.AF0, CRS: v.CRS, DeltaN: v.DeltaN, M0: v.M0, CUC: v.CUC, Eccentricity: v.Eccentricity, CUS: v.CUS, SqrtA: v.SqrtA, TOE: v.TOE, CIC: v.CIC, Omega0: v.Omega0, CIS: v.CIS, I0: v.I0, CRC: v.CRC, Omega: v.Omega, OmegaDot: v.OmegaDot, BGDE5AE1: v.BGDE5AE1, E5ASignalHealth: v.E5ASignalHealth, E5ADataValidity: v.E5ADataValidity, Reserved: v.Reserved}
}
func rtcmGalileoINav(v native.NativeRTCMGalileoINavEphemeris) RTCMGalileoINavEphemeris {
	return RTCMGalileoINavEphemeris{SatelliteID: v.SatelliteID, WeekNumber: v.WeekNumber, IodNav: v.IodNav, SISAIndex: v.SISAIndex, IDOT: v.IDOT, TOC: v.TOC, AF2: v.AF2, AF1: v.AF1, AF0: v.AF0, CRS: v.CRS, DeltaN: v.DeltaN, M0: v.M0, CUC: v.CUC, Eccentricity: v.Eccentricity, CUS: v.CUS, SqrtA: v.SqrtA, TOE: v.TOE, CIC: v.CIC, Omega0: v.Omega0, CIS: v.CIS, I0: v.I0, CRC: v.CRC, Omega: v.Omega, OmegaDot: v.OmegaDot, BGDE5AE1: v.BGDE5AE1, BGDE5BE1: v.BGDE5BE1, E5BSignalHealth: v.E5BSignalHealth, E5BDataValidity: v.E5BDataValidity, E1BSignalHealth: v.E1BSignalHealth, E1BDataValidity: v.E1BDataValidity, Reserved: v.Reserved}
}
func rtcmGLONASS(v native.NativeRTCMGLONASSEphemeris) RTCMGLONASSEphemeris {
	return RTCMGLONASSEphemeris{SatelliteID: v.SatelliteID, FrequencyChannel: v.FrequencyChannel, AlmanacHealth: v.AlmanacHealth, AlmanacHealthAvailability: v.AlmanacHealthAvailability, P1: v.P1, TK: v.TK, BNMSB: v.BNMSB, P2: v.P2, TB: v.TB, XNDot: v.XNDot, XN: v.XN, XNDotDot: v.XNDotDot, YNDot: v.YNDot, YN: v.YN, YNDotDot: v.YNDotDot, ZNDot: v.ZNDot, ZN: v.ZN, ZNDotDot: v.ZNDotDot, P3: v.P3, GammaN: v.GammaN, MP: v.MP, MLNThird: v.MLNThird, TauN: v.TauN, DeltaTauN: v.DeltaTauN, EN: v.EN, MP4: v.MP4, MFT: v.MFT, MNT: v.MNT, MM: v.MM, AdditionalDataAvailable: v.AdditionalDataAvailable, NA: v.NA, TauC: v.TauC, MN4: v.MN4, MTauGPS: v.MTauGPS, MLNFifth: v.MLNFifth, Reserved: v.Reserved}
}

// Close releases the frame collection; repeated calls are safe.
func (f *RTCMFrames) Close() error {
	if f == nil || f.handle == nil {
		return nil
	}
	return publicError(f.handle.Close())
}

// Count returns the number of scanned frames.
func (f *RTCMFrames) Count() (int, error) {
	if f == nil || f.handle == nil {
		return 0, ErrClosed
	}
	v, e := f.handle.Count()
	return v, publicError(e)
}

// Length returns one transport-frame length in bytes.
func (f *RTCMFrames) Length(index int) (int, error) {
	if f == nil || f.handle == nil {
		return 0, ErrClosed
	}
	v, e := f.handle.Len(index)
	return v, publicError(e)
}

// Body returns a detached copy of one scanned frame body.
func (f *RTCMFrames) Body(index int) ([]byte, error) {
	if f == nil || f.handle == nil {
		return nil, ErrClosed
	}
	v, e := f.handle.Body(index)
	return v, publicError(e)
}

// Close releases stream diagnostics; repeated calls are safe.
func (d *RTCMStreamDiagnostics) Close() error {
	if d == nil || d.handle == nil {
		return nil
	}
	return publicError(d.handle.Close())
}

// ResyncBytes returns the number of bytes consumed while resynchronizing.
func (d *RTCMStreamDiagnostics) ResyncBytes() (int, error) {
	if d == nil || d.handle == nil {
		return 0, ErrClosed
	}
	v, e := d.handle.ResyncBytes()
	return v, publicError(e)
}

// SkippedCount returns the number of skipped stream frames.
func (d *RTCMStreamDiagnostics) SkippedCount() (int, error) {
	if d == nil || d.handle == nil {
		return 0, ErrClosed
	}
	v, e := d.handle.SkippedCount()
	return v, publicError(e)
}

// Skipped returns one detached skipped-frame diagnostic.
func (d *RTCMStreamDiagnostics) Skipped(index int) (RTCMFrameSkip, error) {
	if d == nil || d.handle == nil {
		return RTCMFrameSkip{}, ErrClosed
	}
	v, e := d.handle.Skipped(index)
	return RTCMFrameSkip{Offset: v.Offset, HasMessageNumber: v.HasMessageNumber, MessageNumber: v.MessageNumber, Reason: RTCMFrameSkipReason(v.Reason)}, publicError(e)
}

// SkippedMessage returns the native diagnostic message for a skipped frame.
func (d *RTCMStreamDiagnostics) SkippedMessage(index int) (string, error) {
	if d == nil || d.handle == nil {
		return "", ErrClosed
	}
	v, e := d.handle.SkippedMessage(index)
	return string(v), publicError(e)
}

// RTCMLockTimeTracker owns state used to derive per-cell RTCM LLI values.
type RTCMLockTimeTracker struct {
	_      noCopy
	handle *native.RtcmLockTimeTracker
}

// NewRTCMLockTimeTracker creates an empty lock-time tracker.
func NewRTCMLockTimeTracker() (*RTCMLockTimeTracker, error) {
	h, e := native.NewRTCMTracker()
	if e != nil {
		return nil, publicError(e)
	}
	return &RTCMLockTimeTracker{handle: h}, nil
}

// Close releases the lock-time tracker; repeated calls are safe.
func (t *RTCMLockTimeTracker) Close() error {
	if t == nil || t.handle == nil {
		return nil
	}
	return publicError(t.handle.Close())
}

// Reset clears all tracked satellite/signal lock state.
func (t *RTCMLockTimeTracker) Reset() error {
	if t == nil || t.handle == nil {
		return ErrClosed
	}
	return publicError(t.handle.Reset())
}

// Observe derives detached LLI cells from one decoded MSM message.
func (t *RTCMLockTimeTracker) Observe(m *RTCMMessages, index int) ([]RTCMCellLLI, error) {
	if t == nil || t.handle == nil || m == nil || m.handle == nil {
		return nil, ErrClosed
	}
	v, e := t.handle.Observe(m.handle, index)
	if e != nil {
		return nil, publicError(e)
	}
	out := make([]RTCMCellLLI, len(v))
	for i, x := range v {
		out[i] = RTCMCellLLI{SatelliteID: x.SatelliteID, SignalID: x.SignalID, LLI: x.LLI, HasMinLockTime: x.HasMinLockTime, MinLockTimeMS: x.MinLockTimeMS}
	}
	return out, nil
}

// MinimumLockTimeMS returns an optional lock-time value in milliseconds.
func MinimumLockTimeMS(kind RTCMMSMKind, indicator uint16) (uint32, bool, error) {
	if err := validateRTCMMSMKind(kind); err != nil {
		return 0, false, err
	}
	present, v, e := native.MinimumLockTime(uint32(kind), indicator)
	return v, present, publicError(e)
}

// DeriveRTCMILLI derives the standard RTCM loss-of-lock indicator. Nil
// pointers preserve absent previous/current lock times.
func DeriveRTCMILLI(previous *RTCMPreviousLock, currentMinLockTimeMS *uint32, halfCycle bool) (uint8, error) {
	var p *native.NativeRTCMPreviousLock
	if previous != nil {
		p = &native.NativeRTCMPreviousLock{HasMinLockTime: previous.HasMinLockTime, MinLockTimeMS: previous.MinLockTimeMS, ElapsedMS: previous.ElapsedMS}
	}
	var current uint32
	if currentMinLockTimeMS != nil {
		current = *currentMinLockTimeMS
	}
	v, e := native.DeriveLLI(p, currentMinLockTimeMS != nil, current, halfCycle)
	return v, publicError(e)
}

// RTCMPreviousLock contains optional prior lock time in milliseconds and an
// elapsed interval in milliseconds.
type RTCMPreviousLock struct {
	HasMinLockTime bool
	MinLockTimeMS  uint32
	ElapsedMS      uint64
}

// MSMSignalRINEXCode maps an MSM signal ID to its RINEX code.
func MSMSignalRINEXCode(system GNSSSystem, signalID uint8) (string, error) {
	if err := validateGNSSSystem(system); err != nil {
		return "", err
	}
	v, e := native.MSMSignalRINEXCode(uint32(system), signalID)
	return string(v), publicError(e)
}

// MSMEpochDeltaMS returns a constellation-specific epoch delta in
// milliseconds.
func MSMEpochDeltaMS(system GNSSSystem, previous, current uint32) (uint64, error) {
	if err := validateGNSSSystem(system); err != nil {
		return 0, err
	}
	v, e := native.MSMEpochDelta(uint32(system), previous, current)
	return v, publicError(e)
}

// RTCMLLIBits returns the bit widths of RTCM loss-of-lock and half-cycle
// indicators.
func RTCMLLIBits() (lossOfLock, halfCycle uint8, err error) {
	lossOfLock, halfCycle, err = native.RTCMLLIBits()
	return lossOfLock, halfCycle, publicError(err)
}
