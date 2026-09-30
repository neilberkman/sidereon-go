//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package sidereon

import (
	"bytes"
	"testing"
)

func TestRTCMPolicyEncodingStreamAndOwnedDiagnostics(t *testing.T) {
	referenceFrame := []byte{0xd3, 0x00, 0x13, 0x3e, 0xd0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf2, 0x4b, 0xf4}
	if got := testRTCMCRC24Q(referenceFrame[:len(referenceFrame)-3]); got != uint32(referenceFrame[len(referenceFrame)-3])<<16|uint32(referenceFrame[len(referenceFrame)-2])<<8|uint32(referenceFrame[len(referenceFrame)-1]) {
		t.Fatalf("independent reference-frame CRC does not match: %06x", got)
	}
	known, knownDiagnostics, err := DecodeRTCMStreamWithPolicy(referenceFrame, RTCMPolicyStrict)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, known)
	closeAfterTest(t, knownDiagnostics)
	if count, err := known.Count(); err != nil || count != 1 {
		t.Fatalf("literal reference frame decoded count=%d, err=%v", count, err)
	}
	knownMessage, err := known.Message(0)
	if err != nil || knownMessage.Kind != RTCMMessageStationCoordinates || knownMessage.MessageNumber != 1005 || knownMessage.Station == nil || knownMessage.Station.ReferenceStationID != 0 {
		t.Fatalf("literal reference frame message=%+v, err=%v", knownMessage, err)
	}
	if count, err := knownDiagnostics.CRCFailures(); err != nil || count != 0 {
		t.Fatalf("literal reference CRC failures=%d, err=%v", count, err)
	}

	base, err := BuildRTCMStationCoordinates(RTCMStationCoordinates{
		MessageNumber: 1006, ReferenceStationID: 17, GPS: true,
		ECEFX: 111, ECEFY: -222, ECEFZ: 333, HasAntennaHeight: true,
		AntennaHeight: 42,
	})
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, base)
	departing, err := base.WithTrailingBits(0, RTCMTrailingBits{true, false, true})
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, departing)

	if _, _, err := departing.EncodeWithPolicy(0, RTCMPolicyStrict); err == nil {
		t.Fatal("strict policy accepted trailing bits")
	}
	body, departures, err := departing.EncodeWithPolicy(0, RTCMPolicyLenient)
	if err != nil {
		t.Fatal(err)
	}
	if len(departures) != 1 || departures[0].Kind != RTCMDepartureTrailingBits || departures[0].MessageNumber != 1006 || departures[0].BitCount != 8 {
		t.Fatalf("encoding departures = %+v", departures)
	}
	framed, frameDepartures, err := departing.FrameWithPolicy(0, RTCMPolicyLenient)
	if err != nil {
		t.Fatal(err)
	}
	if len(frameDepartures) != 1 || frameDepartures[0].Kind != departures[0].Kind || frameDepartures[0].BitCount != departures[0].BitCount || !bytes.Equal(framed[3:len(framed)-3], body) {
		t.Fatalf("frame result/departures do not preserve encoded body: frame=%x body=%x departures=%+v", framed, body, frameDepartures)
	}
	if _, _, err := departing.FrameWithPolicy(0, RTCMPolicy(99)); err == nil {
		t.Fatal("unknown policy accepted")
	}

	badCRC := append([]byte(nil), referenceFrame...)
	badCRC[len(badCRC)-1] ^= 1
	stream := append(append([]byte(nil), badCRC...), framed...)
	messages, diagnostics, err := DecodeRTCMStreamWithPolicy(stream, RTCMPolicyLenient)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, messages)
	closeAfterTest(t, diagnostics)
	count, err := messages.Count()
	if err != nil || count != 1 {
		t.Fatalf("decoded message count=%d, err=%v", count, err)
	}
	crcFailures, err := diagnostics.CRCFailures()
	if err != nil || crcFailures != 1 {
		t.Fatalf("CRC failures=%d, err=%v", crcFailures, err)
	}
	departureCount, err := diagnostics.DepartureCount()
	if err != nil || departureCount != 1 {
		t.Fatalf("departure count=%d, err=%v", departureCount, err)
	}
	row, err := diagnostics.Departure(0)
	if err != nil || row.Offset != len(badCRC) || row.Description == "" {
		t.Fatalf("stream departure=%+v, err=%v", row, err)
	}
	if _, err := diagnostics.Departure(1); err == nil {
		t.Fatal("out-of-range departure index accepted")
	}
	if err := diagnostics.Close(); err != nil {
		t.Fatal(err)
	}
	if got, err := diagnostics.Departure(0); err == nil || got.Description != "" {
		t.Fatalf("closed diagnostics returned %+v, %v", got, err)
	}
	if row.Offset != len(badCRC) || row.Description == "" {
		t.Fatalf("detached diagnostic changed after close: %+v", row)
	}
	strictMessages, strictDiagnostics, err := DecodeRTCMStreamWithPolicy(stream, RTCMPolicyStrict)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, strictMessages)
	closeAfterTest(t, strictDiagnostics)
	if count, err := strictMessages.Count(); err != nil || count != 0 {
		t.Fatalf("strict stream emitted %d departing messages: %v", count, err)
	}
	if count, err := strictDiagnostics.CRCFailures(); err != nil || count != 1 {
		t.Fatalf("strict CRC failures=%d, err=%v", count, err)
	}
	if count, err := strictDiagnostics.SkippedCount(); err != nil || count != 1 {
		t.Fatalf("strict skipped frames=%d, err=%v", count, err)
	}
	if count, err := strictDiagnostics.DepartureCount(); err != nil || count != 0 {
		t.Fatalf("strict departures=%d, err=%v", count, err)
	}
}

func testRTCMCRC24Q(data []byte) uint32 {
	var crc uint32
	for _, value := range data {
		crc ^= uint32(value) << 16
		for bit := 0; bit < 8; bit++ {
			crc <<= 1
			if crc&0x1000000 != 0 {
				crc ^= 0x1864cfb
			}
		}
	}
	return crc & 0xffffff
}
