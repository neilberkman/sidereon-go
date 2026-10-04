package sidereon

import (
	"testing"

	"sidereon.dev/go/v3/internal/native"
)

func TestSSRSizeV2GoConversionsPreserveTypedReasonAndSignedClock(t *testing.T) {
	nativeRows := []native.NativeSPPRejectedSatelliteV2{
		{SatelliteID: "G07", Reason: 5, HasSize: true, OrbitM: 12.5, ClockM: -0.75},
		{SatelliteID: "E11", Reason: 1, HasSize: false, OrbitM: 0, ClockM: 0},
	}
	sppRows := sppRejectedSatellitesV2FromNative(nativeRows)
	if sppRows[0].Reason != SPPRejectionSsrCorrectionExceedsLimit || !sppRows[0].HasSize || sppRows[0].ClockM != -0.75 {
		t.Fatalf("SPP V2 conversion lost refusal payload: %+v", sppRows[0])
	}
	staticRow := staticRejectedSatelliteV2FromNative(nativeRows[0])
	if staticRow.Reason != StaticPositionRejectionSsrCorrectionExceedsLimit || !staticRow.HasSize || staticRow.OrbitM != 12.5 || staticRow.ClockM != -0.75 {
		t.Fatalf("static V2 conversion lost refusal payload: %+v", staticRow)
	}

	pppRows := pppUnplacedV2FromNative([]native.PppUnplacedObservationV2{{EpochIndex: 3, SatelliteID: "G07", AmbiguityID: "amb-7", Reason: 1, HasSize: true, OrbitM: 8.25, ClockM: -0.5}})
	if len(pppRows) != 1 || pppRows[0].Reason != PPPUnplacedObservationSsrCorrectionExceedsLimit || pppRows[0].EpochIndex != 3 || pppRows[0].ClockM != -0.5 {
		t.Fatalf("PPP V2 conversion lost refusal payload: %+v", pppRows)
	}
}
