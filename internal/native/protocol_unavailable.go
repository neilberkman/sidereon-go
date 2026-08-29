//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

import "errors"

var errNegativeIndex = errors.New("sidereon: index must not be negative")

func protocolUnavailable() error { return unavailable() }

const (
	RTCMMessageMSMValue                   = uint32(0)
	RTCMMessageStationCoordinatesValue    = uint32(1)
	RTCMMessageAntennaDescriptorValue     = uint32(2)
	RTCMMessageGPSEphemerisValue          = uint32(3)
	RTCMMessageGLONASSEphemerisValue      = uint32(4)
	RTCMMessageSSRValue                   = uint32(5)
	RTCMMessageUnsupportedValue           = uint32(6)
	RTCMMessageBeiDouEphemerisValue       = uint32(7)
	RTCMMessageQZSSEphemerisValue         = uint32(8)
	RTCMMessageGalileoFNavEphemerisValue  = uint32(9)
	RTCMMessageGalileoINavEphemerisValue  = uint32(10)
	RTCMMSM4Value                         = uint32(0)
	RTCMMSM7Value                         = uint32(1)
	RTCMFrameTruncatedValue               = uint32(0)
	RTCMFrameMalformedValue               = uint32(1)
	RTCMAntennaDescriptorFieldValue       = uint32(0)
	RTCMAntennaSerialNumberFieldValue     = uint32(1)
	RTCMReceiverTypeFieldValue            = uint32(2)
	RTCMReceiverFirmwareVersionFieldValue = uint32(3)
	RTCMReceiverSerialNumberFieldValue    = uint32(4)
	EphemerisSampleValidValue             = uint32(0)
	EphemerisSampleGapValue               = uint32(1)
	ObservableStateValidValue             = uint32(0)
	ObservableStateGapValue               = uint32(1)
	ObservableStateErrorValue             = uint32(2)
	EmissionMediaValidValue               = uint32(0)
	EmissionMediaGapValue                 = uint32(1)
	EmissionMediaBelowElevationValue      = uint32(2)
	EmissionMediaErrorValue               = uint32(3)
	SignalModulationBPSKValue             = uint32(0)
	SignalModulationBOCSineValue          = uint32(1)
	SignalModulationBOCCosineValue        = uint32(2)
	SignalModulationMBOCValue             = uint32(3)
	SignalModulationTMBOCValue            = uint32(4)
	SignalModulationCBOCPlusValue         = uint32(5)
	SignalModulationCBOCMinusValue        = uint32(6)
	DLLCoherentValue                      = uint32(0)
	DLLNonCoherentValue                   = uint32(1)
	SBASPLNoErrorValue                    = uint32(0)
	SBASPLInsufficientGeometryValue       = uint32(1)
	SBASPLNumericalFailureValue           = uint32(2)
	SBASPLInvalidErrorModelValue          = uint32(3)
	SBASSolveMixedValue                   = uint32(0)
	SBASSolveSBASOnlyValue                = uint32(1)
	SSRReferencePointAntennaValue         = uint32(0)
	SSRReferencePointCenterOfMassValue    = uint32(1)
	SSRSourceRTCMValue                    = uint32(0)
	SSRSourceGalileoHASValue              = uint32(1)
	SSRMissingDeclineValue                = uint32(0)
	SSRMissingFallbackValue               = uint32(1)

	SBASWireFramed250Value = uint32(0)
	SBASWireBody226Value   = uint32(1)

	SBASMessageDoNotUseValue            = uint32(0)
	SBASMessagePRNMaskValue             = uint32(1)
	SBASMessageFastCorrectionsValue     = uint32(2)
	SBASMessageIntegrityValue           = uint32(3)
	SBASMessageFastDegradationValue     = uint32(4)
	SBASMessageGeoNavValue              = uint32(5)
	SBASMessageNetworkTimeValue         = uint32(6)
	SBASMessageGeoAlmanacValue          = uint32(7)
	SBASMessageIGPMaskValue             = uint32(8)
	SBASMessageMixedCorrectionsValue    = uint32(9)
	SBASMessageLongTermCorrectionsValue = uint32(10)
	SBASMessageIONODelaysValue          = uint32(11)
	SBASMessageUnsupportedValue         = uint32(12)

	RTCMSSROrbitValue              = uint32(0)
	RTCMSSRClockValue              = uint32(1)
	RTCMSSRCombinedOrbitClockValue = uint32(2)
	RTCMSSRCodeBiasValue           = uint32(3)
	RTCMSSRPhaseBiasValue          = uint32(4)
	RTCMSSRURAValue                = uint32(5)
	RTCMSSRHighRateClockValue      = uint32(6)
	RTCMSSRVTECValue               = uint32(7)

	RINEXClockInstantJulianDateValue = uint32(0)
	RINEXClockInstantNanosValue      = uint32(1)
)

type NativeGnssWeekTow struct {
	System     uint32
	Week       uint32
	TOWSeconds float64
}
type NativeKeplerianElements struct{ SqrtA, E, M0, DeltaN, Omega0, I0, Omega, OmegaDot, IDot, CUC, CUS, CRC, CRS, CIC, CIS, ToeSOW float64 }
type NativeClockPolynomial struct{ AF0, AF1, AF2, TocSOW float64 }
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
	Present                       bool
	ADOTMPerS, DeltaN0DotRadPerS2 float64
	Top                           NativeGnssWeekTow
	URAEDIndex, URANED0Index      int8
	URANED1Index, URANED2Index    uint8
	TransmissionTimeSOW           float64
	HasFlags                      bool
	Flags                         uint32
}
type NativeBroadcastRecord struct {
	SatelliteID                        string
	Message, Issue, IssueMessage, Week uint32
	Toe, Toc                           NativeGnssWeekTow
	Elements                           NativeKeplerianElements
	Clock                              NativeClockPolynomial
	GroupDelays                        NativeBroadcastGroupDelays
	CNAV                               NativeBroadcastCNAV
	SVHealth, SVAccuracyM              float64
	HasFitInterval                     bool
	FitIntervalS                       float64
}
type NativeSkippedNavBlock struct{ SatelliteID, Message string }
type NativeIonoCorrections struct {
	GPSPresent                         bool
	GPSAlpha, GPSBeta                  [4]float64
	BeiDouPresent                      bool
	BeiDouAlpha, BeiDouBeta            [4]float64
	GalileoPresent                     bool
	GalileoAI0, GalileoAI1, GalileoAI2 float64
}

type RinexNavRecords struct{}

func ParseRinexNavRecords([]byte) (*RinexNavRecords, error)        { return nil, protocolUnavailable() }
func (*RinexNavRecords) Close() error                              { return nil }
func (*RinexNavRecords) Records() ([]NativeBroadcastRecord, error) { return nil, protocolUnavailable() }
func (*RinexNavRecords) Record(int) (NativeBroadcastRecord, error) {
	return NativeBroadcastRecord{}, protocolUnavailable()
}
func (*RinexNavRecords) Count() (int, error)                 { return 0, protocolUnavailable() }
func EncodeRinexNav([]NativeBroadcastRecord) ([]byte, error) { return nil, protocolUnavailable() }
func ParseRinexIonoCorrections([]byte) (NativeIonoCorrections, error) {
	return NativeIonoCorrections{}, protocolUnavailable()
}
func ParseRinexLeapSeconds([]byte) (float64, bool, error) { return 0, false, protocolUnavailable() }

type RinexNavParse struct{}

func ParseRinexNavLenient([]byte) (*RinexNavParse, error)        { return nil, protocolUnavailable() }
func (*RinexNavParse) Close() error                              { return nil }
func (*RinexNavParse) Records() ([]NativeBroadcastRecord, error) { return nil, protocolUnavailable() }
func (*RinexNavParse) Skipped() ([]NativeSkippedNavBlock, error) { return nil, protocolUnavailable() }
func (*RinexNavParse) RecordCount() (int, error)                 { return 0, protocolUnavailable() }
func (*RinexNavParse) SkippedCount() (int, error)                { return 0, protocolUnavailable() }

type NativeClockEpoch struct {
	Scale                       uint32
	Representation              uint32
	JulianWhole, JulianFraction float64
	NanosHigh                   int64
	NanosLow                    uint64
}
type NativeClockPoint struct {
	Epoch NativeClockEpoch
	BiasS float64
}
type RinexClock struct{}
type ClockSeries struct{}

func ParseRinexClock([]byte, bool) (*RinexClock, error)    { return nil, protocolUnavailable() }
func (*RinexClock) Close() error                           { return nil }
func (*RinexClock) Satellites() ([]string, error)          { return nil, protocolUnavailable() }
func (*RinexClock) SatelliteCount() (int, error)           { return 0, protocolUnavailable() }
func (*RinexClock) SeriesCount() (int, error)              { return 0, protocolUnavailable() }
func (*RinexClock) SampleCount() (int, error)              { return 0, protocolUnavailable() }
func (*RinexClock) Series(int) (*ClockSeries, error)       { return nil, protocolUnavailable() }
func (*RinexClock) SeriesFor(string) (*ClockSeries, error) { return nil, protocolUnavailable() }
func (*RinexClock) BiasAtGPSSeconds(string, float64) (float64, bool, error) {
	return 0, false, protocolUnavailable()
}
func (*RinexClock) Text() ([]byte, error)                 { return nil, protocolUnavailable() }
func (*ClockSeries) Close() error                         { return nil }
func (*ClockSeries) Satellite() (string, error)           { return "", protocolUnavailable() }
func (*ClockSeries) Samples() ([]NativeClockPoint, error) { return nil, protocolUnavailable() }
func (*ClockSeries) SampleCount() (int, error)            { return 0, protocolUnavailable() }

type NativeSbasMessageInfo struct {
	Form, Kind                               uint32
	MessageType, Preamble                    uint8
	FastCount, LongTermCount, IONODelayCount int
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
	Preamble                                                   uint8
	TimeOfDayS                                                 uint16
	URA                                                        uint8
	XM, YM, ZM, XRateMPerS, YRateMPerS, ZRateMPerS             int32
	XAccelMPerS2, YAccelMPerS2, ZAccelMPerS2, AGF0S, AGF1SPerS int16
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
type SbasBlock struct{}

func DecodeSbasBlock([]byte, uint32) (*SbasBlock, error) { return nil, protocolUnavailable() }
func (*SbasBlock) Close() error                          { return nil }
func (*SbasBlock) Info() (NativeSbasMessageInfo, error) {
	return NativeSbasMessageInfo{}, protocolUnavailable()
}
func (*SbasBlock) FastCorrections() (bool, NativeSbasRawFastCorrections, error) {
	return false, NativeSbasRawFastCorrections{}, protocolUnavailable()
}
func (*SbasBlock) FastDegradation() (bool, NativeSbasFastDegradation, error) {
	return false, NativeSbasFastDegradation{}, protocolUnavailable()
}
func (*SbasBlock) GeoNav() (bool, NativeSbasGeoNavMessage, error) {
	return false, NativeSbasGeoNavMessage{}, protocolUnavailable()
}
func (*SbasBlock) IGPMask() (bool, NativeSbasIgpMask, error) {
	return false, NativeSbasIgpMask{}, protocolUnavailable()
}
func (*SbasBlock) Integrity() (bool, NativeSbasIntegrity, error) {
	return false, NativeSbasIntegrity{}, protocolUnavailable()
}
func (*SbasBlock) IonoDelays() (bool, NativeSbasIonoDelays, error) {
	return false, NativeSbasIonoDelays{}, protocolUnavailable()
}
func (*SbasBlock) MixedFastCorrections() (bool, NativeSbasMixedFastCorrections, error) {
	return false, NativeSbasMixedFastCorrections{}, protocolUnavailable()
}
func (*SbasBlock) PrnMask() (bool, NativeSbasPrnMask, error) {
	return false, NativeSbasPrnMask{}, protocolUnavailable()
}
func (*SbasBlock) LongTermHalf(int) (bool, NativeSbasLongTermHalfInfo, error) {
	return false, NativeSbasLongTermHalfInfo{}, protocolUnavailable()
}
func (*SbasBlock) LongTermRecords(int) ([]NativeSbasLongTermRecord, error) {
	return nil, protocolUnavailable()
}
func (*SbasBlock) RawData() ([]byte, error) { return nil, protocolUnavailable() }
func (*SbasBlock) Encode() ([]byte, error)  { return nil, protocolUnavailable() }

type NativeSbasLogBlock struct {
	SatelliteID string
	Epoch       NativeGnssWeekTow
	Form        uint32
	ByteCount   int
}
type SbasLogBlocks struct{}

func ParseSbasEMSLines([]byte) (*SbasLogBlocks, error)      { return nil, protocolUnavailable() }
func ParseSbasRTKLIBLines([]byte) (*SbasLogBlocks, error)   { return nil, protocolUnavailable() }
func (*SbasLogBlocks) Close() error                         { return nil }
func (*SbasLogBlocks) Items() ([]NativeSbasLogBlock, error) { return nil, protocolUnavailable() }
func (*SbasLogBlocks) Bytes(int) ([]byte, error)            { return nil, protocolUnavailable() }
func (*SbasLogBlocks) Count() (int, error)                  { return 0, protocolUnavailable() }
func SbasPRNToSatelliteID(uint16) (string, bool, error)     { return "", false, protocolUnavailable() }

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
type SsrMessage struct{}

func DecodeSsrMessage([]byte) (*SsrMessage, error)                 { return nil, protocolUnavailable() }
func (*SsrMessage) Close() error                                   { return nil }
func (*SsrMessage) Body() ([]byte, error)                          { return nil, protocolUnavailable() }
func (*SsrMessage) Encode() ([]byte, error)                        { return nil, protocolUnavailable() }
func (*SsrMessage) Info() (NativeSsrInfo, error)                   { return NativeSsrInfo{}, protocolUnavailable() }
func (*SsrMessage) Orbits() ([]NativeSsrOrbitRecord, error)        { return nil, protocolUnavailable() }
func (*SsrMessage) Clocks() ([]NativeSsrClockRecord, error)        { return nil, protocolUnavailable() }
func (*SsrMessage) CodeBiases() ([]NativeSsrCodeBiasRecord, error) { return nil, protocolUnavailable() }
func (*SsrMessage) CodeBiasSignals(int) ([]NativeSsrCodeBiasSignal, error) {
	return nil, protocolUnavailable()
}
func (*SsrMessage) PhaseBiases() ([]NativeSsrPhaseBiasRecord, error) {
	return nil, protocolUnavailable()
}
func (*SsrMessage) PhaseBiasSignals(int) ([]NativeSsrPhaseBiasSignal, error) {
	return nil, protocolUnavailable()
}
func (*SsrMessage) URA() ([]NativeSsrUraRecord, error) { return nil, protocolUnavailable() }
