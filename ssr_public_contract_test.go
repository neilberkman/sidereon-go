package sidereon

import (
	"bytes"
	"encoding/hex"
	"reflect"
	"testing"
)

func TestSSRPublicPhaseBiasMessagePreservesEveryReturnedField(t *testing.T) {
	wantHeader := RTCMSSRHeader{
		EpochTimeS: 4321, UpdateInterval: 2, MultipleMessage: true, IODSSR: 4,
		ProviderID: 12, SolutionID: 3,
		HasDispersiveBiasConsistency: true, DispersiveBiasConsistency: true,
		HasMWConsistency: true, MWConsistency: false, SatelliteCount: 1,
	}
	wantInfoV2 := RTCMSSRInfoV2{
		MessageNumber: 4076, System: GNSSSystemSBAS, Kind: RTCMSSRPhaseBias,
		Header: wantHeader, HasIGSSSRVersion: true, IGSSSRVersion: 1,
		PhaseBiasCount: 1, PaddingBits: RTCMTrailingBits{false},
	}
	wantGroup := RTCMSSRPhaseBiasGroup{
		Record: RTCMSSRPhaseBiasRecord{SatelliteID: 48, YawAngle: 20, YawRate: -1, SignalCount: 1},
		Signals: []RTCMSSRPhaseBiasSignal{{
			SignalID: 7, IntegerIndicator: 1, WideLaneIntegerIndicator: 0,
			DiscontinuityCounter: 9, Bias: -12,
		}},
	}
	messages, err := BuildRTCMSSRMessageV2(wantInfoV2, nil, nil, nil, nil, []RTCMSSRPhaseBiasGroup{wantGroup})
	if err != nil {
		t.Fatalf("BuildRTCMSSRMessageV2() error = %v", err)
	}
	closeAfterTest(t, messages)

	gotInfoV2, err := messages.SSRInfoV2(0)
	if err != nil || !reflect.DeepEqual(gotInfoV2, wantInfoV2) {
		t.Fatalf("SSRInfoV2() = %#v, %v; want %#v", gotInfoV2, err, wantInfoV2)
	}
	bareBody, err := messages.Encode(0)
	assertSSRPublicTransportFixture(t, bareBody, "d30012fec2fc021c254000c381c0214ff3c9ffff40462b6c")
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	decoded, err := DecodeSSRMessage(bareBody)
	if err != nil {
		t.Fatalf("DecodeSSRMessage() error = %v", err)
	}
	closeAfterTest(t, decoded)
	gotBody, err := decoded.Body()
	if err != nil || !bytes.Equal(gotBody, bareBody) {
		t.Fatalf("Body() = %x, %v; want exact input %x", gotBody, err, bareBody)
	}
	gotEncoded, err := decoded.Encode()
	if err != nil || !bytes.Equal(gotEncoded, bareBody) {
		t.Fatalf("Encode() = %x, %v; want exact input %x", gotEncoded, err, bareBody)
	}
	wantInfo := RTCMSSRInfo{
		MessageNumber: wantInfoV2.MessageNumber, System: wantInfoV2.System, Kind: wantInfoV2.Kind,
		Header: wantHeader, PhaseBiasCount: 1,
	}
	gotInfo, err := decoded.Info()
	if err != nil || !reflect.DeepEqual(gotInfo, wantInfo) {
		t.Fatalf("Info() = %#v, %v; want %#v", gotInfo, err, wantInfo)
	}
	gotGroups, err := decoded.PhaseBiasGroups()
	if err != nil || !reflect.DeepEqual(gotGroups, []RTCMSSRPhaseBiasGroup{wantGroup}) {
		t.Fatalf("PhaseBiasGroups() = %#v, %v; want %#v", gotGroups, err, []RTCMSSRPhaseBiasGroup{wantGroup})
	}
	for _, check := range []struct {
		name string
		get  func() (int, error)
	}{
		{"orbits", func() (int, error) { values, err := decoded.Orbits(); return len(values), err }},
		{"clocks", func() (int, error) { values, err := decoded.Clocks(); return len(values), err }},
		{"code biases", func() (int, error) { values, err := decoded.CodeBiasGroups(); return len(values), err }},
		{"URA", func() (int, error) { values, err := decoded.URA(); return len(values), err }},
	} {
		count, err := check.get()
		if err != nil || count != 0 {
			t.Fatalf("%s() count = %d, %v; want 0", check.name, count, err)
		}
	}
}

func TestSSRPublicOrbitCorrectionPreservesEveryField(t *testing.T) {
	const week = uint32(2400)
	const tow = 100000.0
	wantInfo := RTCMSSRInfoV2{
		MessageNumber: 1057, System: GNSSSystemGPS, Kind: RTCMSSROrbit,
		Header: RTCMSSRHeader{
			EpochTimeS: uint32(tow), UpdateInterval: 1, IODSSR: 4,
			ProviderID: 12, SolutionID: 3, HasSatelliteReferenceDatum: true,
			SatelliteReferenceDatum: true, SatelliteCount: 1,
		},
		OrbitCount: 1,
	}
	wantRecord := RTCMSSROrbitRecord{
		SatelliteID: 1, IODE: 17, DeltaRadial: -3, DeltaAlong: 4, DeltaCross: -5,
		DotDeltaRadial: 6, DotDeltaAlong: -7, DotDeltaCross: 8,
	}
	messages, err := BuildRTCMSSRMessageV2(wantInfo, []RTCMSSROrbitRecord{wantRecord}, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("BuildRTCMSSRMessageV2() error = %v", err)
	}
	closeAfterTest(t, messages)
	store, err := NewSSRCorrectionStore(SSRReferencePointCenterOfMass)
	if err != nil {
		t.Fatalf("NewSSRCorrectionStore() error = %v", err)
	}
	closeAfterTest(t, store)
	fixtureBody, err := messages.Encode(0)
	if err != nil {
		t.Fatalf("Encode(0) error = %v", err)
	}
	assertSSRPublicTransportFixture(t, fixtureBody, "d3001a421186a01500030c10447ffffd00004ffffb000037fff9000100f41ef3")
	epoch := GNSSWeekTow{System: GPST, Week: week, TOWSeconds: tow}
	if err := store.Ingest(messages, epoch); err != nil {
		t.Fatalf("Ingest() error = %v", err)
	}
	got, present, err := store.Orbit("G01")
	want := SSROrbitCorrection{
		Source: SSRSourceRTCM, ProviderID: 12, SolutionID: 3, IODE: 17, IODSSR: 4,
		CRSRegional: true, ReferencePoint: SSRReferencePointCenterOfMass,
		RadialM:         -float64(wantRecord.DeltaRadial) * 1e-4,
		AlongM:          -float64(wantRecord.DeltaAlong) * 4e-4,
		CrossM:          -float64(wantRecord.DeltaCross) * 4e-4,
		RadialRateMPerS: -float64(wantRecord.DotDeltaRadial) * 1e-6,
		AlongRateMPerS:  -float64(wantRecord.DotDeltaAlong) * 4e-6,
		CrossRateMPerS:  -float64(wantRecord.DotDeltaCross) * 4e-6,
		RefEpochJ2000S:  float64(week)*604800 + tow - 630763200 + 1,
		UpdateIntervalS: 2,
	}
	if err != nil || !present || !reflect.DeepEqual(got, want) {
		t.Fatalf("Orbit(G01) = %#v, present=%t, err=%v; want %#v, present=true", got, present, err, want)
	}
}

func TestSSRPublicVTECMessagePreservesEveryField(t *testing.T) {
	want := RTCMSSRVTECInfo{
		MessageNumber: 4076, HasIGSSSRVersion: true, IGSSSRVersion: 1,
		EpochTimeS: 100000, UpdateInterval: 3, MultipleMessage: true,
		IODSSR: 4, ProviderID: 7, SolutionID: 2, QualityIndicator: 123,
		Layers: []RTCMTecLayer{{Height: 45, Degree: 2, Order: 1,
			Cosine: []int16{100, 200, -300, 600, -700}, Sine: []int16{400, -500}}},
		TrailingBits: RTCMTrailingBits{},
	}
	messages, err := BuildRTCMSSRVTEC(want)
	if err != nil {
		t.Fatalf("BuildRTCMSSRVTEC() error = %v", err)
	}
	closeAfterTest(t, messages)
	fixtureBody, err := messages.Encode(0)
	if err != nil {
		t.Fatalf("Encode(0) error = %v", err)
	}
	assertSSRPublicTransportFixture(t, fixtureBody, "d3001bfec39230d4074000723d85a2000c80191fda804b1fa880321fc180d1d44d")
	got, err := messages.SSRVTEC(0)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("SSRVTEC() = %#v, %v; want %#v", got, err, want)
	}
}

func assertSSRPublicTransportFixture(t *testing.T, body []byte, wantFrameHex string) {
	t.Helper()
	wantFrame, err := hex.DecodeString(wantFrameHex)
	if err != nil {
		t.Fatalf("DecodeString(fixture) error = %v", err)
	}
	frame, err := EncodeRTCMFrame(body)
	if err != nil || !bytes.Equal(frame, wantFrame) {
		t.Fatalf("EncodeRTCMFrame() = %x, %v; want %x", frame, err, wantFrame)
	}
	gotBody, consumed, err := DecodeRTCMFrame(frame)
	if err != nil || consumed != len(frame) || !bytes.Equal(gotBody, body) {
		t.Fatalf("DecodeRTCMFrame() = body %x, consumed %d, err %v; want body %x and consumed %d", gotBody, consumed, err, body, len(frame))
	}
}
