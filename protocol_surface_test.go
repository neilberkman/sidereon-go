package sidereon

import (
	"bytes"
	"encoding/hex"
	"errors"
	"math"
	"os"
	"testing"
)

func acceptanceBytes(t *testing.T, value string) []byte {
	t.Helper()
	result, err := hex.DecodeString(value)
	if err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	return result
}

func TestCommittedRTCMCaproundSurface(t *testing.T) {
	stream := acceptanceBytes(t, "d300153ee7d30302aa3c6d183e4605ff0c02ef2b54843a98d8b487d3001b3f07d30b54524d35393830302e3030010a3134343038313233343563b90fd3003d3fb207b07ffb2a03e800fffd00c0e42aff9100deffed2979ffce003d0900003ca0eebb0003e8000700087a23fff7000f423f014dfff421cffffc00fd0076821cd3002d3fc151a0c89e8004d20002c5c10010e1800447b10000de800029a000c50001b8874578d2c4000f120160002a0020e36bd300244357d3000789000000008000000000000020000000518200fe700c0e7ffdb5c644b00014432c82")
	if len(stream) != 220 {
		t.Fatalf("RTCM fixture length = %d, want 220", len(stream))
	}
	messages, err := DecodeRTCM(stream)
	if err != nil {
		t.Fatal(err)
	}
	if count, err := messages.Count(); err != nil || count != 5 {
		t.Fatalf("message count = %d, %v", count, err)
	}
	station, err := messages.StationCoordinates(0)
	if err != nil {
		t.Fatal(err)
	}
	if station.MessageNumber != 1006 || station.ReferenceStationID != 2003 || !station.HasAntennaHeight || station.AntennaHeightM != 1.5 || station.ECEFX == 0 {
		t.Fatalf("1006 fields = %+v", station)
	}
	antenna, err := messages.AntennaDescriptor(1)
	if err != nil {
		t.Fatal(err)
	}
	if antenna.MessageNumber != 1008 || antenna.ReferenceStationID != 2003 || !antenna.HasAntennaSerialNumber || antenna.HasReceiverType {
		t.Fatalf("1008 fields = %+v", antenna)
	}
	if descriptor, err := messages.AntennaString(1, RTCMAntennaDescriptorField); err != nil || descriptor != "TRM59800.00" {
		t.Fatalf("1008 descriptor = %q, %v", descriptor, err)
	}
	if serial, err := messages.AntennaString(1, RTCMAntennaSerialNumberField); err != nil || serial != "1440812345" {
		t.Fatalf("1008 serial = %q, %v", serial, err)
	}
	gps, err := messages.GPSEphemeris(2)
	if err != nil {
		t.Fatal(err)
	}
	if gps.SatelliteID != 8 || gps.WeekNumber != 123 || gps.AF0 != 12345 {
		t.Fatalf("1019 fields = %+v", gps)
	}
	glonass, err := messages.GLONASSEphemeris(3)
	if err != nil {
		t.Fatal(err)
	}
	if glonass.SatelliteID != 5 || glonass.FrequencyChannel != 8 || glonass.MNT != 700 {
		t.Fatalf("1020 fields = %+v", glonass)
	}
	msm, err := messages.MSMInfo(4)
	if err != nil {
		t.Fatal(err)
	}
	if msm.MessageNumber != 1077 || msm.System != GNSSSystemGPS || msm.Kind != RTCMMSM7 || msm.Header.ReferenceStationID != 2003 || msm.SatelliteCount != 1 || msm.SignalCount != 1 {
		t.Fatalf("1077 info = %+v", msm)
	}
	satellites, err := messages.MSMSatellites(4)
	if err != nil || len(satellites) != 1 || satellites[0] != (RTCMMSMSatellite{ID: 8, RoughRangeMS: 70, RoughRangeMod1: 512, HasExtendedInfo: true, ExtendedInfo: 0, HasRoughPhaseRangeRate: true, RoughPhaseRangeRateMS: -100}) {
		t.Fatalf("1077 satellites = %+v, %v", satellites, err)
	}
	signals, err := messages.MSMSignals(4)
	if err != nil || len(signals) != 1 || signals[0].SatelliteID != 8 || signals[0].SignalID != 2 || signals[0].FinePseudorange != 12345 || signals[0].FinePhaseRange != -2345 || signals[0].LockTimeIndicator != 100 || signals[0].CNR != 600 || !signals[0].HasFinePhaseRangeRate || signals[0].FinePhaseRangeRate != 5 {
		t.Fatalf("1077 signals = %+v, %v", signals, err)
	}
	if lossBit, halfBit, err := RTCMLLIBits(); err != nil || lossBit != 1 || halfBit != 2 {
		t.Fatalf("LLI bits = %d, %d, %v", lossBit, halfBit, err)
	}
	if minimum, present, err := MinimumLockTimeMS(RTCMMSM7, 64); err != nil || !present || minimum != 64 {
		t.Fatalf("lock time = %d, %v, %v", minimum, present, err)
	}
	if minimum, present, err := MinimumLockTimeMS(RTCMMSM7, 705); err != nil || present || minimum != 0 {
		t.Fatalf("reserved lock time = %d, %v, %v", minimum, present, err)
	}
	previous := RTCMPreviousLock{HasMinLockTime: true, MinLockTimeMS: 512, ElapsedMS: 1000}
	current := uint32(512)
	lli, err := DeriveRTCMILLI(&previous, &current, true)
	if err != nil || lli != 3 {
		t.Fatalf("derived LLI = %d, %v", lli, err)
	}
	if delta, err := MSMEpochDeltaMS(GNSSSystemGPS, 604799000, 500); err != nil || delta != 1500 {
		t.Fatalf("MSM epoch rollover = %d, %v", delta, err)
	}
	if code, err := MSMSignalRINEXCode(GNSSSystemGPS, 2); err != nil || code != "1C" {
		t.Fatalf("MSM signal code = %q, %v", code, err)
	}
	tracker, err := NewRTCMLockTimeTracker()
	if err != nil {
		t.Fatal(err)
	}
	lliRows, err := tracker.Observe(messages, 4)
	if err != nil || len(lliRows) != 1 || lliRows[0] != (RTCMCellLLI{SatelliteID: 8, SignalID: 2, HasMinLockTime: true, MinLockTimeMS: 144}) {
		t.Fatalf("tracker rows = %+v, %v", lliRows, err)
	}
	if err := tracker.Reset(); err != nil {
		t.Fatal(err)
	}
	if err := tracker.Close(); err != nil {
		t.Fatal(err)
	}
	if err := tracker.Reset(); !errors.Is(err, ErrClosed) {
		t.Fatalf("tracker post-close reset = %v", err)
	}
	if err := tracker.Close(); err != nil {
		t.Fatal(err)
	}
	frame, err := messages.Frame(0)
	if err != nil || !bytes.Equal(frame, stream[:27]) {
		t.Fatalf("1006 frame round trip = %x, %v", frame, err)
	}
	body, err := messages.Encode(0)
	if err != nil || len(body) != 21 {
		t.Fatalf("1006 body = %d, %v", len(body), err)
	}
	if err := messages.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := messages.Count(); !errors.Is(err, ErrClosed) {
		t.Fatalf("messages post-close count = %v", err)
	}
	if err := messages.Close(); err != nil {
		t.Fatal(err)
	}

	truncatedFrame, err := EncodeRTCMFrame([]byte{0x43, 0x50})
	if err != nil {
		t.Fatal(err)
	}
	noisy := append([]byte{0x11, 0x22}, truncatedFrame...)
	noisy = append(noisy, stream...)
	streamMessages, diagnostics, err := DecodeRTCMStream(noisy)
	if err != nil {
		t.Fatal(err)
	}
	if count, err := streamMessages.Count(); err != nil || count != 5 {
		t.Fatalf("reframed stream count = %d, %v", count, err)
	}
	if value, err := diagnostics.ResyncBytes(); err != nil || value != 2 {
		t.Fatalf("resync bytes = %d, %v", value, err)
	}
	if value, err := diagnostics.SkippedCount(); err != nil || value != 1 {
		t.Fatalf("skipped count = %d, %v", value, err)
	}
	skip, err := diagnostics.Skipped(0)
	if err != nil || skip.Offset != 2 || !skip.HasMessageNumber || skip.MessageNumber != 1077 || skip.Reason != RTCMFrameTruncated {
		t.Fatalf("skip = %+v, %v", skip, err)
	}
	if value, err := diagnostics.SkippedMessage(0); err != nil || value != "" {
		t.Fatalf("skip message = %q, %v", value, err)
	}
	assertConcurrentClose(t, func() error { _, err := streamMessages.Count(); return err }, streamMessages.Close)
	assertConcurrentClose(t, func() error { _, err := diagnostics.SkippedCount(); return err }, diagnostics.Close)

	frames, err := ScanRTCMFrames(stream)
	if err != nil {
		t.Fatal(err)
	}
	if count, err := frames.Count(); err != nil || count != 5 {
		t.Fatalf("frame count = %d, %v", count, err)
	}
	if length, err := frames.Length(0); err != nil || length != 27 {
		t.Fatalf("first frame length = %d, %v", length, err)
	}
	if body, err := frames.Body(0); err != nil || !bytes.Equal(body, stream[3:24]) {
		t.Fatalf("first frame body = %x, %v", body, err)
	}
	assertConcurrentClose(t, func() error { _, err := frames.Count(); return err }, frames.Close)
}

func TestCommittedRTCMBuildersAnd1046(t *testing.T) {
	station, err := BuildRTCMStationCoordinates(RTCMStationCoordinates{MessageNumber: 1006, ReferenceStationID: 2003, ITRFRealizationYear: 1, GPS: true, GLONASS: true, Galileo: true, ReferenceStation: true, ECEFX: 38403690, ECEFY: 6863060, ECEFZ: 50208700, HasAntennaHeight: true, AntennaHeight: 15000, AntennaHeightM: 1.5})
	if err != nil {
		t.Fatal(err)
	}
	stationValue, err := station.StationCoordinates(0)
	if err != nil || stationValue.ReferenceStationID != 2003 || stationValue.ECEFX != 38403690 || !stationValue.HasAntennaHeight || stationValue.AntennaHeight != 15000 {
		t.Fatalf("built 1006 = %+v, %v", stationValue, err)
	}
	if err := station.Close(); err != nil {
		t.Fatal(err)
	}
	descriptor := "TRM59800.00"
	serial := "1440812345"
	antenna, err := BuildRTCMAntennaDescriptor(1008, 2003, 1, descriptor, &serial, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	antennaValue, err := antenna.AntennaDescriptor(0)
	if err != nil || antennaValue.MessageNumber != 1008 || !antennaValue.HasAntennaSerialNumber || antennaValue.HasReceiverType {
		t.Fatalf("built 1008 = %+v, %v", antennaValue, err)
	}
	if value, err := antenna.AntennaString(0, RTCMAntennaDescriptorField); err != nil || value != descriptor {
		t.Fatalf("built descriptor = %q, %v", value, err)
	}
	if err := antenna.Close(); err != nil {
		t.Fatal(err)
	}
	gps, err := BuildRTCMGPSEphemeris(RTCMGPSEphemeris{SatelliteID: 8, WeekNumber: 123, AF0: 12345, TOE: 7200, SqrtA: 2702336448})
	if err != nil {
		t.Fatal(err)
	}
	gpsValue, err := gps.GPSEphemeris(0)
	if err != nil || gpsValue.SatelliteID != 8 || gpsValue.WeekNumber != 123 || gpsValue.AF0 != 12345 || gpsValue.SqrtA != 2702336448 {
		t.Fatalf("built 1019 = %+v, %v", gpsValue, err)
	}
	if err := gps.Close(); err != nil {
		t.Fatal(err)
	}
	glonass, err := BuildRTCMGLONASSEphemeris(RTCMGLONASSEphemeris{SatelliteID: 5, FrequencyChannel: 8, TB: 30, MNT: 700})
	if err != nil {
		t.Fatal(err)
	}
	glonassValue, err := glonass.GLONASSEphemeris(0)
	if err != nil || glonassValue.SatelliteID != 5 || glonassValue.FrequencyChannel != 8 || glonassValue.TB != 30 || glonassValue.MNT != 700 {
		t.Fatalf("built 1020 = %+v, %v", glonassValue, err)
	}
	if err := glonass.Close(); err != nil {
		t.Fatal(err)
	}
	beiDou, err := BuildRTCMBeiDouEphemeris(RTCMBeiDouEphemeris{SatelliteID: 19, WeekNumber: 902, AODE: 17, TOC: 12000, AF1: 12345, AF0: -45678, SqrtA: 2852448983, TOE: 12000})
	if err != nil {
		t.Fatal(err)
	}
	beiDouValue, err := beiDou.BeiDouEphemeris(0)
	if err != nil || beiDouValue.SatelliteID != 19 || beiDouValue.WeekNumber != 902 || beiDouValue.AODE != 17 || beiDouValue.AF0 != -45678 || beiDouValue.SqrtA != 2852448983 {
		t.Fatalf("built 1042 = %+v, %v", beiDouValue, err)
	}
	if err := beiDou.Close(); err != nil {
		t.Fatal(err)
	}
	qzss, err := BuildRTCMQZSSEphemeris(RTCMQZSSEphemeris{SatelliteID: 3, WeekNumber: 123, IODE: 11, TOC: 7200, AF0: 23456, SqrtA: 2702336448, TOE: 3600, CodesOnL2: 1})
	if err != nil {
		t.Fatal(err)
	}
	qzssValue, err := qzss.QZSSEphemeris(0)
	if err != nil || qzssValue.SatelliteID != 3 || qzssValue.WeekNumber != 123 || qzssValue.IODE != 11 || qzssValue.CodesOnL2 != 1 || qzssValue.SqrtA != 2702336448 {
		t.Fatalf("built 1044 = %+v, %v", qzssValue, err)
	}
	if err := qzss.Close(); err != nil {
		t.Fatal(err)
	}
	fnav, err := BuildRTCMGalileoFNavEphemeris(RTCMGalileoFNavEphemeris{SatelliteID: 12, WeekNumber: 1402, IodNav: 7, SISA: 42, TOC: 5150, AF1: -151, AF0: -471483, SqrtA: 2852448983, TOE: 5150})
	if err != nil {
		t.Fatal(err)
	}
	fnavValue, err := fnav.GalileoFNavEphemeris(0)
	if err != nil || fnavValue.SatelliteID != 12 || fnavValue.WeekNumber != 1402 || fnavValue.IodNav != 7 || fnavValue.SISA != 42 || fnavValue.AF0 != -471483 || fnavValue.SqrtA != 2852448983 {
		t.Fatalf("built 1045 = %+v, %v", fnavValue, err)
	}
	if err := fnav.Close(); err != nil {
		t.Fatal(err)
	}
	inav, err := BuildRTCMGalileoINavEphemeris(RTCMGalileoINavEphemeris{SatelliteID: 3, WeekNumber: 1402, IodNav: 7, SISAIndex: 107, TOC: 5150, AF1: -151, AF0: -471483, SqrtA: 2852448983, TOE: 5150, BGDE5BE1: 7})
	if err != nil {
		t.Fatal(err)
	}
	inavValue, err := inav.GalileoINavEphemeris(0)
	if err != nil || inavValue.SatelliteID != 3 || inavValue.WeekNumber != 1402 || inavValue.IodNav != 7 || inavValue.SISAIndex != 107 || inavValue.AF0 != -471483 || inavValue.SqrtA != 2852448983 || inavValue.BGDE5BE1 != 7 {
		t.Fatalf("built 1046 = %+v, %v", inavValue, err)
	}
	if err := inav.Close(); err != nil {
		t.Fatal(err)
	}
	msm, err := BuildRTCMMSM(RTCMMSMInfo{MessageNumber: 1077, System: GNSSSystemGPS, Kind: RTCMMSM7, Header: RTCMMSMHeader{ReferenceStationID: 2003, EpochTime: 100000}}, []RTCMMSMSatellite{{ID: 8, RoughRangeMS: 75, RoughRangeMod1: 512, HasExtendedInfo: true, ExtendedInfo: 3, HasRoughPhaseRangeRate: true, RoughPhaseRangeRateMS: -100}}, []RTCMMSMSignal{{SatelliteID: 8, SignalID: 2, FinePseudorange: 1234, FinePhaseRange: -5678, LockTimeIndicator: 200, CNR: 720, HasFinePhaseRangeRate: true, FinePhaseRangeRate: 42}})
	if err != nil {
		t.Fatal(err)
	}
	msmInfo, err := msm.MSMInfo(0)
	if err != nil || msmInfo.MessageNumber != 1077 || msmInfo.SatelliteCount != 1 || msmInfo.SignalCount != 1 {
		t.Fatalf("built MSM info = %+v, %v", msmInfo, err)
	}
	if err := msm.Close(); err != nil {
		t.Fatal(err)
	}
	real1046 := acceptanceBytes(t, "d3003f4160d5e8076b06c941e03ffed3ffe33917f3a490e984d2089bf4f4011030b0343aa813ab5d41efffb7e44fe8cfff5277d0b011a2416397fffffc2280140700800a8e")
	decoded, err := DecodeRTCM(real1046)
	if err != nil {
		t.Fatal(err)
	}
	info, err := decoded.Message(0)
	if err != nil || info.Kind != RTCMMessageGalileoINavEphemeris || info.MessageNumber != 1046 || info.GalileoINav == nil || info.GalileoINav.SatelliteID != 3 || info.GalileoINav.WeekNumber != 1402 || info.GalileoINav.IodNav != 7 || info.GalileoINav.SqrtA != 2852448983 || info.GalileoINav.Eccentricity != 4459564 {
		t.Fatalf("real 1046 = %+v, %v", info, err)
	}
	frame, err := decoded.Frame(0)
	if err != nil || !bytes.Equal(frame, real1046) {
		t.Fatalf("real 1046 frame = %x, %v", frame, err)
	}
	assertConcurrentClose(t, func() error { _, err := decoded.Count(); return err }, decoded.Close)
}

func TestCommittedSSRMessageNestedSurface(t *testing.T) {
	codeBody := acceptanceBytes(t, "423546002c803da021883d97249290")
	code, err := DecodeSSRMessage(codeBody)
	if err != nil {
		t.Fatal(err)
	}
	if body, err := code.Body(); err != nil || !bytes.Equal(body, codeBody) {
		t.Fatalf("code body = %x, %v", body, err)
	}
	codeInfo, err := code.Info()
	if err != nil || codeInfo.MessageNumber != 1059 || codeInfo.System != GNSSSystemGPS || codeInfo.Kind != RTCMSSRCodeBias || codeInfo.CodeBiasCount != 1 {
		t.Fatalf("code info = %+v, %v", codeInfo, err)
	}
	codeGroups, err := code.CodeBiasGroups()
	if err != nil || len(codeGroups) != 1 || codeGroups[0].Record.SatelliteID != 3 || len(codeGroups[0].Signals) != 2 || codeGroups[0].Signals[0] != (RTCMSSRCodeBiasSignal{SignalID: 1, Bias: -1234}) || codeGroups[0].Signals[1] != (RTCMSSRCodeBiasSignal{SignalID: 9, Bias: 2345}) {
		t.Fatalf("code groups = %+v, %v", codeGroups, err)
	}
	if _, err := code.CodeBiasSignals(4); err == nil {
		t.Fatal("invalid code-bias record index accepted")
	}
	if err := code.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := code.Info(); !errors.Is(err, ErrClosed) {
		t.Fatalf("code post-close info = %v", err)
	}
	if err := code.Close(); err != nil {
		t.Fatal(err)
	}

	phaseBody := acceptanceBytes(t, "4f1546002c803da408623ffa071f0ee024a1ca2380")
	phase, err := DecodeSSRMessage(phaseBody)
	if err != nil {
		t.Fatal(err)
	}
	phaseInfo, err := phase.Info()
	if err != nil || phaseInfo.MessageNumber != 1265 || phaseInfo.System != GNSSSystemGPS || phaseInfo.Kind != RTCMSSRPhaseBias || phaseInfo.PhaseBiasCount != 1 || !phaseInfo.Header.HasDispersiveBiasConsistency || !phaseInfo.Header.DispersiveBiasConsistency || !phaseInfo.Header.HasMWConsistency || phaseInfo.Header.MWConsistency {
		t.Fatalf("phase info = %+v, %v", phaseInfo, err)
	}
	phaseGroups, err := phase.PhaseBiasGroups()
	if err != nil || len(phaseGroups) != 1 || phaseGroups[0].Record.SatelliteID != 3 || phaseGroups[0].Record.YawAngle != 127 || phaseGroups[0].Record.YawRate != -12 || len(phaseGroups[0].Signals) != 2 {
		t.Fatalf("phase groups = %+v, %v", phaseGroups, err)
	}
	if phaseGroups[0].Signals[0] != (RTCMSSRPhaseBiasSignal{SignalID: 1, IntegerIndicator: 1, WideLaneIntegerIndicator: 2, DiscontinuityCounter: 3, Bias: -123456}) || phaseGroups[0].Signals[1] != (RTCMSSRPhaseBiasSignal{SignalID: 9, IntegerIndicator: 0, WideLaneIntegerIndicator: 1, DiscontinuityCounter: 4, Bias: 234567}) {
		t.Fatalf("phase signals = %+v", phaseGroups[0].Signals)
	}
	if err := phase.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := phase.Info(); !errors.Is(err, ErrClosed) {
		t.Fatalf("phase post-close info = %v", err)
	}
	if err := phase.Close(); err != nil {
		t.Fatal(err)
	}

	uraBody := acceptanceBytes(t, "425546002c803da021d2")
	ura, err := DecodeSSRMessage(uraBody)
	if err != nil {
		t.Fatal(err)
	}
	uraInfo, err := ura.Info()
	if err != nil || uraInfo.MessageNumber != 1061 || uraInfo.System != GNSSSystemGPS || uraInfo.Kind != RTCMSSRURA || uraInfo.URACount != 1 {
		t.Fatalf("URA info = %+v, %v", uraInfo, err)
	}
	uraRows, err := ura.URA()
	if err != nil || len(uraRows) != 1 || uraRows[0] != (RTCMSSRURARecord{SatelliteID: 3, URAIndex: 41}) {
		t.Fatalf("URA rows = %+v, %v", uraRows, err)
	}
	if err := ura.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := ura.Info(); !errors.Is(err, ErrClosed) {
		t.Fatalf("URA post-close info = %v", err)
	}
	if err := ura.Close(); err != nil {
		t.Fatal(err)
	}

	combinedFrame := acceptanceBytes(t, "d3003c4245438a3040000827968003270026dffea30000f7fff6ffff0000530000000000003e87fff8effc94002c7ffff57fffc80003004128000000000000625cf0")
	combined, err := DecodeRTCM(combinedFrame)
	if err != nil {
		t.Fatal(err)
	}
	combinedInfo, err := combined.SSRInfo(0)
	if err != nil || combinedInfo.MessageNumber != 1060 || combinedInfo.System != GNSSSystemGPS || combinedInfo.Kind != RTCMSSRCombinedOrbitClock || combinedInfo.OrbitCount == 0 || combinedInfo.ClockCount == 0 {
		t.Fatalf("combined info = %+v, %v", combinedInfo, err)
	}
	combinedMessage, err := combined.Message(0)
	if err != nil || combinedMessage.SSR == nil || len(combinedMessage.SSROrbits) != combinedInfo.OrbitCount || len(combinedMessage.SSRClocks) != combinedInfo.ClockCount {
		t.Fatalf("combined message = %+v, %v", combinedMessage, err)
	}
	bareBody, err := combined.Encode(0)
	if err != nil {
		t.Fatal(err)
	}
	bare, err := DecodeSSRMessage(bareBody)
	if err != nil {
		t.Fatal(err)
	}
	bareInfo, err := bare.Info()
	if err != nil || bareInfo != combinedInfo {
		t.Fatalf("bare/framed info = %+v / %+v, %v", bareInfo, combinedInfo, err)
	}
	bareOrbits, err := bare.Orbits()
	if err != nil {
		t.Fatal(err)
	}
	framedOrbits, err := combined.SSROrbits(0)
	if err != nil || len(framedOrbits) != len(bareOrbits) {
		t.Fatalf("orbit counts = %d / %d, %v", len(framedOrbits), len(bareOrbits), err)
	}
	for index := range framedOrbits {
		if framedOrbits[index] != bareOrbits[index] {
			t.Fatalf("orbit[%d] framed=%+v bare=%+v", index, framedOrbits[index], bareOrbits[index])
		}
	}
	bareClocks, err := bare.Clocks()
	if err != nil {
		t.Fatal(err)
	}
	framedClocks, err := combined.SSRClocks(0)
	if err != nil || len(framedClocks) != len(bareClocks) {
		t.Fatalf("clock counts = %d / %d, %v", len(framedClocks), len(bareClocks), err)
	}
	for index := range framedClocks {
		if framedClocks[index] != bareClocks[index] {
			t.Fatalf("clock[%d] framed=%+v bare=%+v", index, framedClocks[index], bareClocks[index])
		}
	}
	assertConcurrentClose(t, func() error { _, err := bare.Info(); return err }, bare.Close)
	assertConcurrentClose(t, func() error { _, err := combined.Count(); return err }, combined.Close)
}

func TestCommittedSignalAcquisitionCorrelationAndAnalysis(t *testing.T) {
	bpsk := BPSK(1)
	boc := BOCSine(1, 1)
	if label, err := ModulationLabel(bpsk); err != nil || label != "BPSK(n)" {
		t.Fatalf("BPSK label = %q, %v", label, err)
	}
	if value, err := ReferenceChipRateHz(); err != nil || value != 1023000 {
		t.Fatalf("reference chip rate = %v, %v", value, err)
	}
	if value, err := BetzL1ReceiverBandwidthHz(); err != nil || value != 24000000 {
		t.Fatalf("Betz bandwidth = %v, %v", value, err)
	}
	if value, err := ModulationCodeRateHz(bpsk); err != nil || value != 1023000 {
		t.Fatalf("BPSK code rate = %v, %v", value, err)
	}
	for label, value := range map[string]float64{
		"BPSK PSD": 9.775171065493646e-7,
		"BOC PSD":  3.9617276106485926e-7,
	} {
		var actual float64
		var err error
		if label == "BPSK PSD" {
			actual, err = AnalysisPSDHz(bpsk, 0)
		} else {
			actual, err = AnalysisPSDHz(boc, 0.5*1023000)
		}
		if err != nil || math.Abs(actual-value) > 1e-20 {
			t.Fatalf("%s = %.17g, %v", label, actual, err)
		}
	}
	if actual, err := FractionPowerInBand(bpsk, 24000000); err != nil || math.Abs(actual-0.99147813722178968) > 1e-15 {
		t.Fatalf("BPSK fraction = %.17g, %v", actual, err)
	}
	if actual, err := AnalysisRMSBandwidthHz(boc, 24000000); err != nil || math.Abs(actual-1978624.6068839289) > 1e-9 {
		t.Fatalf("BOC RMS bandwidth = %.17g, %v", actual, err)
	}
	cn0, err := AnalysisEffectiveCN0Degradation(bpsk, 45, 24000000, []SignalInterference{{Modulation: boc, PowerRatioToCarrier: 0.01}})
	if err != nil || math.Abs(cn0.EffectiveCN0Hz-31621.13351302073) > 1e-8 || math.Abs(cn0.EffectiveCN0DBHz-44.999774338955795) > 1e-12 || math.Abs(cn0.DegradationDB-0.00022566104420462807) > 1e-15 {
		t.Fatalf("analysis C/N0 = %+v, %v", cn0, err)
	}
	dll := DLLTrackingOptions{CN0DBHz: 45, LoopBandwidthHz: 1, IntegrationTimeS: 0.02, CorrelatorSpacingChips: 0.5, ReceiverBandwidthHz: 100000000}
	jitter, err := AnalysisDLLThermalNoiseJitter(bpsk, dll, DLLCoherent)
	if err != nil || math.Abs(jitter.Chips-0.0027925349810969391) > 1e-15 || math.Abs(jitter.Meters-0.8183586764751074) > 1e-12 {
		t.Fatalf("analysis DLL jitter = %+v, %v", jitter, err)
	}
	lower, err := DLLLowerBound(boc, dll)
	if err != nil || math.Abs(lower.Seconds-2.2630212065471776e-10) > 1e-21 {
		t.Fatalf("DLL lower bound = %+v, %v", lower, err)
	}
	multipath, err := AnalysisMultipathErrorEnvelope(bpsk, MultipathOptions{MultipathToDirectRatio: 0.5, CorrelatorSpacingChips: 1, ReceiverBandwidthHz: 100000000}, []float64{0, 0.5, 1})
	if err != nil || len(multipath) != 3 || math.Abs(multipath[1].InPhaseChips-0.16666443790427826) > 1e-12 || math.Abs(multipath[1].AntiPhaseChips+0.20000709850498244) > 1e-12 || math.Abs(multipath[1].RunningAverageChips-0.10000354925249122) > 1e-12 {
		t.Fatalf("analysis multipath = %+v, %v", multipath, err)
	}

	replica, err := SignalReplica(1, ReplicaOptions{SampleRateHz: 1023000, NumSamples: 1023})
	if err != nil || len(replica) != 1023 {
		t.Fatalf("replica = %d, %v", len(replica), err)
	}
	code, err := CACode(1)
	if err != nil || len(code) != 1023 || code[0] != replica[0] {
		t.Fatalf("C/A code = %d, %v", len(code), err)
	}
	samples := make([]IQSample, len(replica))
	for index, value := range replica {
		samples[index] = IQSample{I: float64(value)}
	}
	correlation, err := CorrelateSignal(samples, 1, CorrelateOptions{SampleRateHz: 1023000})
	if err != nil || correlation.I <= 1000 || math.Abs(correlation.Q) > 1e-9 || correlation.Power <= 1000000 {
		t.Fatalf("PRN correlation = %+v, %v", correlation, err)
	}
	i, q, err := CorrelateAgainst(samples, replica, 1023000, 0)
	if err != nil || math.Abs(i-correlation.I) > 1e-9 || math.Abs(q-correlation.Q) > 1e-9 {
		t.Fatalf("explicit correlation = %.17g, %.17g, %v", i, q, err)
	}
	if value, err := CorrelationAt(replica, replica, 0); err != nil || value != 1023 {
		t.Fatalf("correlation at zero = %d, %v", value, err)
	}
	autocorrelation, err := Autocorrelation(replica)
	if err != nil || len(autocorrelation) != 1023 || autocorrelation[0] != 1023 {
		t.Fatalf("autocorrelation = %d, first=%d, %v", len(autocorrelation), autocorrelation[0], err)
	}
	acquisition, metrics, err := AcquireSignal(samples, 1, AcquisitionOptions{SampleRateHz: 1023000, DopplerMinHz: 0, DopplerMaxHz: 0, DopplerStepHz: 1})
	if err != nil || acquisition.GridDopplerBinCount != 1 || acquisition.PeakPower <= 1000000 || len(metrics) == 0 {
		t.Fatalf("acquisition = %+v metrics=%d, %v", acquisition, len(metrics), err)
	}
}

func TestCommittedCorrectionStoreSurface(t *testing.T) {
	sbasBody := acceptanceBytes(t, "5366819010029ee7ed83018202819bbe1a08bf8008ffa00000004066c0")
	block, err := DecodeSBASBlock(sbasBody, SBASWireBody226)
	if err != nil {
		t.Fatal(err)
	}
	info, err := block.Info()
	if err != nil || info.Form != SBASWireBody226 || info.Kind != SBASMessageLongTermCorrections || info.MessageType != 25 || info.LongTermCount != 2 {
		t.Fatalf("SBAS info = %+v, %v", info, err)
	}
	encoded, err := block.Encode()
	if err != nil || !bytes.Equal(encoded, sbasBody) {
		t.Fatalf("SBAS encode = %x, %v", encoded, err)
	}
	half, err := block.LongTermHalf(0)
	if err != nil {
		t.Fatal(err)
	}
	longTerm, err := block.LongTermRecords(0)
	if err != nil {
		t.Fatal(err)
	}
	if half.RecordCount != len(longTerm) || len(longTerm) == 0 {
		t.Fatalf("SBAS long-term counts = %+v records=%+v", half, longTerm)
	}
	secondHalf, err := block.LongTermHalf(1)
	if err != nil {
		t.Fatal(err)
	}
	secondRecords, err := block.LongTermRecords(1)
	if err != nil {
		t.Fatal(err)
	}
	if half.VelocityCode != true || half.IODP != 3 || half.RecordCount != 1 || longTerm[0] != (SBASLongTermRecord{MonitoredIndex: 16, IODE: 50, DeltaX: 16, DeltaY: 20, DeltaZ: -71, DeltaXRate: 6, DeltaYRate: 3, DeltaZRate: 4, DeltaAF0: -37, DeltaAF1: 5, HasTimeOfDayS: true, TimeOfDayS: 102}) || secondHalf.VelocityCode != true || secondHalf.IODP != 3 || secondHalf.RecordCount != 1 || secondRecords[0] != (SBASLongTermRecord{MonitoredIndex: 31, IODE: 13, DeltaX: 34, DeltaY: -16, DeltaZ: 8, DeltaAF0: -3, DeltaAF1: 2, HasTimeOfDayS: true, TimeOfDayS: 102}) {
		t.Fatalf("SBAS long-term fields = half=%+v records=%+v second=%+v records=%+v", half, longTerm, secondHalf, secondRecords)
	}
	store, err := NewSBASCorrectionStore()
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Ingest(block, "S20", GNSSWeekTow{System: GPST, Week: 2400, TOWSeconds: 20}); err != nil {
		t.Fatal(err)
	}
	preferred, present, err := store.PreferredGEO(0)
	if err != nil {
		t.Fatal(err)
	}
	ready, err := store.ReadyGEOs(0)
	if err != nil {
		t.Fatal(err)
	}
	if present || preferred != "" || len(ready) != 0 {
		t.Fatalf("SBAS store selection = %q, %v, %v", preferred, present, ready)
	}
	if _, present, err := store.FastCorrection("S20", "G01"); err != nil || present {
		t.Fatalf("absent SBAS fast correction = present=%v err=%v", present, err)
	}
	if _, present, err := store.LongTermCorrection("S20", "G01"); err != nil || present {
		t.Fatalf("absent SBAS long-term correction = present=%v err=%v", present, err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := store.ReadyGEOs(0); !errors.Is(err, ErrClosed) {
		t.Fatalf("SBAS store post-close = %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	if err := block.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := block.Info(); !errors.Is(err, ErrClosed) {
		t.Fatalf("SBAS block post-close = %v", err)
	}
	if err := block.Close(); err != nil {
		t.Fatal(err)
	}
	racingBlock, err := DecodeSBASBlock(sbasBody, SBASWireBody226)
	if err != nil {
		t.Fatal(err)
	}
	assertConcurrentClose(t, func() error { _, err := racingBlock.Info(); return err }, racingBlock.Close)

	combinedFrame := acceptanceBytes(t, "d3003c4245438a3040000827968003270026dffea30000f7fff6ffff0000530000000000003e87fff8effc94002c7ffff57fffc80003004128000000000000625cf0")
	ssr, err := NewSSRCorrectionStoreFromRTCM(combinedFrame, GNSSWeekTow{System: GPST, Week: 2425, TOWSeconds: 344970})
	if err != nil {
		t.Fatal(err)
	}
	orbit, present, err := ssr.Orbit("G30")
	if err != nil || !present || math.IsNaN(orbit.RadialM) {
		t.Fatalf("SSR orbit = %+v present=%v err=%v", orbit, present, err)
	}
	clock, present, err := ssr.Clock("G30")
	if err != nil || !present || math.IsNaN(clock.C0M) {
		t.Fatalf("SSR clock = %+v present=%v err=%v", clock, present, err)
	}
	if _, present, err := ssr.Orbit("G32"); err != nil || present {
		t.Fatalf("absent SSR orbit = present=%v err=%v", present, err)
	}
	if _, present, err := ssr.Clock("G32"); err != nil || present {
		t.Fatalf("absent SSR clock = present=%v err=%v", present, err)
	}
	if _, present, err := ssr.CodeBias("G30", 1); err != nil || present {
		t.Fatalf("absent SSR code bias = present=%v err=%v", present, err)
	}
	if _, present, err := ssr.PhaseBias("G30", 1); err != nil || present {
		t.Fatalf("absent SSR phase bias = present=%v err=%v", present, err)
	}
	if _, present, err := ssr.URAIndex("G30"); err != nil || present {
		t.Fatalf("absent SSR URA = present=%v err=%v", present, err)
	}
	if err := ssr.Close(); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ssr.Clock("G30"); !errors.Is(err, ErrClosed) {
		t.Fatalf("SSR store post-close = %v", err)
	}
	if err := ssr.Close(); err != nil {
		t.Fatal(err)
	}
	empty, err := NewSSRCorrectionStore(SSRReferencePointCenterOfMass)
	if err != nil {
		t.Fatal(err)
	}
	if _, present, err := empty.Orbit("G30"); err != nil || present {
		t.Fatalf("empty SSR orbit = present=%v err=%v", present, err)
	}
	if err := empty.Close(); err != nil {
		t.Fatal(err)
	}
	if err := empty.Close(); err != nil {
		t.Fatal(err)
	}
	racingSSR, err := NewSSRCorrectionStore(SSRReferencePointAntennaPhaseCenter)
	if err != nil {
		t.Fatal(err)
	}
	assertConcurrentClose(t, func() error { _, _, err := racingSSR.Clock("G30"); return err }, racingSSR.Close)
}

func TestCommittedGLONASSRecordsAndSkips(t *testing.T) {
	fixture := "     3.05           NAVIGATION DATA     M                   RINEX VERSION / TYPE\n" +
		"     XXX                                                         END OF HEADER\n" +
		"R01 2020 06 24 23 15 00 6.355904042721e-05 0.000000000000e+00 3.420000000000e+05\n" +
		"     1.090894238281e+04 1.407806396484e+00-1.862645149231e-09 0.000000000000e+00\n" +
		"    -2.885726074219e+03 2.795855522156e+00-0.000000000000e+00 1.000000000000e+00\n" +
		"     2.288353955078e+04-3.169984817505e-01-2.793967723846e-09 0.000000000000e+00\n"
	records, err := ParseRINEXGLONASSRecords([]byte(fixture))
	if err != nil {
		t.Fatal(err)
	}
	count, err := records.Count()
	if err != nil || count != 1 {
		t.Fatalf("GLONASS count = %d, %v", count, err)
	}
	value, err := records.Record(0)
	if err != nil {
		t.Fatal(err)
	}
	assertFloat := func(label string, actual float64, expected uint64) {
		t.Helper()
		if bits := math.Float64bits(actual); bits != expected {
			t.Fatalf("%s bits = %016x, want %016x", label, bits, expected)
		}
	}
	if value.SatelliteID != "R01" || value.FrequencyChannel != 1 {
		t.Fatalf("GLONASS identity = %+v", value)
	}
	for index, expected := range [][3]uint64{{0x4164cea1cc3ffac2, 0xc146042f09800219, 0x4175d2cd38cffeb0}, {0x4095ff39bffff98f, 0x40a5d7b60700020c, 0xc073cff9c80000ce}, {0xbebf4000000000cb, 0x8000000000000000, 0xbec76ffffffffbfc}} {
		var actual [3]float64
		switch index {
		case 0:
			actual = value.PositionM
		case 1:
			actual = value.VelocityMPerS
		case 2:
			actual = value.AccelerationMPerS2
		}
		for coordinate := range expected {
			assertFloat("GLONASS vector", actual[coordinate], expected[coordinate])
		}
	}
	assertFloat("GLONASS toe", value.ToeUTCJ2000S, 0x41c342f91a000000)
	assertFloat("GLONASS clock", value.ClockBiasS, 0x3f10a96000000098)
	assertFloat("GLONASS gamma", value.GammaN, 0)
	assertFloat("GLONASS health", value.SVHealth, 0)

	extended := "     3.05           NAVIGATION DATA     M                   RINEX VERSION / TYPE\n" +
		"     XXX                                                         END OF HEADER\n" +
		"R28 2020 06 24 23 15 00 6.355904042721e-05 0.000000000000e+00 3.420000000000e+05\n" +
		"     1.090894238281e+04 1.407806396484e+00-1.862645149231e-09 0.000000000000e+00\n" +
		"    -2.885726074219e+03 2.795855522156e+00-0.000000000000e+00 1.000000000000e+00\n" +
		"     2.288353955078e+04-3.169984817505e-01-2.793967723846e-09 0.000000000000e+00\n"
	extendedRecords, err := ParseRINEXGLONASSRecords([]byte(extended))
	if err != nil {
		t.Fatal(err)
	}
	if count, err := extendedRecords.Count(); err != nil || count != 0 {
		t.Fatalf("extended GLONASS count = %d, %v", count, err)
	}
	if count, err := extendedRecords.SkippedCount(); err != nil || count != 1 {
		t.Fatalf("extended GLONASS skips = %d, %v", count, err)
	}
	skip, err := extendedRecords.Skipped(0)
	if err != nil || skip.SatelliteID != "R28" {
		t.Fatalf("extended GLONASS skip = %+v, %v", skip, err)
	}
	assertConcurrentClose(t, func() error { _, err := extendedRecords.SkippedCount(); return err }, extendedRecords.Close)
	if err := records.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := records.Record(0); !errors.Is(err, ErrClosed) {
		t.Fatalf("GLONASS post-close = %v", err)
	}
	if err := records.Close(); err != nil {
		t.Fatal(err)
	}

	nav, err := os.ReadFile("testdata/nav/ESBC00DNK_R_20201770000_01D_MN.rnx")
	if err != nil {
		t.Fatal(err)
	}
	combined := append(append([]byte(nil), nav...), []byte(fixture)...)
	broadcast, err := ParseBroadcastEphemeris(combined)
	if err != nil {
		t.Fatal(err)
	}
	broadcastRecords, err := broadcast.GLONASSRecords()
	if err != nil || len(broadcastRecords) != 1 || broadcastRecords[0] != value {
		t.Fatalf("broadcast GLONASS records = %+v, %v", broadcastRecords, err)
	}
	channels, err := broadcast.GLONASSFrequencyChannels()
	if err != nil || len(channels) != 1 || channels[0] != (FrequencyChannel{Slot: 1, Channel: 1}) {
		t.Fatalf("broadcast GLONASS channels = %+v, %v", channels, err)
	}
	if err := broadcast.Close(); err != nil {
		t.Fatal(err)
	}
}
