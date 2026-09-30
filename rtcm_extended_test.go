package sidereon

import (
	"bytes"
	"testing"
)

func TestRTCMExtendedLegacySystemTextAndRetentionRoutes(t *testing.T) {
	legacyInput := RTCMLegacyObservations{
		MessageNumber:      1004,
		ReferenceStationID: 12,
		EpochTime:          345,
		SatelliteCount:     1,
		Satellites: []RTCMLegacySatellite{{
			SatelliteID: 7,
			L1:          RTCMLegacyL1{CodeIndicator: true, Pseudorange: 123, PhaseRangeMinusPseudorange: -9, LockTimeIndicator: 4, HasPseudorangeModulusAmbiguity: true, PseudorangeModulusAmbiguity: 3, HasCNR: true, CNR: 40},
			HasL2:       true,
			L2:          RTCMLegacyL2{CodeIndicator: 2, PseudorangeDifference: -5, PhaseRangeMinusL1Pseudorange: 6, LockTimeIndicator: 7, HasCNR: true, CNR: 38},
		}},
		TrailingBits: RTCMTrailingBits{},
	}
	legacy, err := BuildRTCMLegacyObservations(legacyInput)
	if err != nil {
		t.Fatal(err)
	}
	defer legacy.Close()
	legacyMessage, err := legacy.Message(0)
	if err != nil {
		t.Fatal(err)
	}
	if legacyMessage.Legacy == nil || legacyMessage.Legacy.Satellites[0].L1.Pseudorange != 123 || legacyMessage.Legacy.Satellites[0].L1.PhaseRangeMinusPseudorange != -9 || !legacyMessage.Legacy.Satellites[0].HasL2 || legacyMessage.Legacy.Satellites[0].L2.PseudorangeDifference != -5 {
		t.Fatalf("legacy payload mismatch: %+v", legacyMessage.Legacy)
	}
	if got, err := legacy.TrailingBits(0); err != nil || !bytes.Equal(boolBytes(got), boolBytes(legacyInput.TrailingBits)) {
		t.Fatalf("legacy trailing bits = %v, %v", got, err)
	}
	cloned, err := legacy.WithTrailingBits(0, RTCMTrailingBits{false, true})
	if err != nil {
		t.Fatal(err)
	}
	defer cloned.Close()
	if got, err := cloned.TrailingBits(0); err != nil || !bytes.Equal(boolBytes(got), []byte{0, 1}) {
		t.Fatalf("replacement trailing bits = %v, %v", got, err)
	}

	parameterInput := RTCMSystemParameters{ReferenceStationID: 1, MJD: 60000, SecondsOfDay: 1234, AnnouncementCount: 1, LeapSeconds: 18, Announcements: []RTCMMessageAnnouncement{{MessageNumber: 1077, Synchronous: true, Interval: 10}}}
	parameters, err := BuildRTCMSystemParameters(parameterInput)
	if err != nil {
		t.Fatal(err)
	}
	defer parameters.Close()
	parameterMessage, err := parameters.Message(0)
	if err != nil {
		t.Fatal(err)
	}
	if parameterMessage.SystemParameters == nil || parameterMessage.SystemParameters.MJD != 60000 || parameterMessage.SystemParameters.Announcements[0].Interval != 10 {
		t.Fatalf("system-parameters payload mismatch: %+v", parameterMessage.SystemParameters)
	}

	textInput := RTCMText{ReferenceStationID: 2, MJD: 60001, SecondsOfDay: 42, CharacterCount: 1, CodeUnits: []byte{0xc3, 0xa9}}
	textMessages, err := BuildRTCMText(textInput)
	if err != nil {
		t.Fatal(err)
	}
	defer textMessages.Close()
	textMessage, err := textMessages.Message(0)
	if err != nil {
		t.Fatal(err)
	}
	if textMessage.Text == nil || !bytes.Equal(textMessage.Text.CodeUnits, textInput.CodeUnits) || textMessage.Text.CharacterCount != 1 {
		t.Fatalf("text payload mismatch: %+v", textMessage.Text)
	}

	unsupportedInput := RTCMUnsupportedBody{MessageNumber: 4090, Body: []byte{0xFF, 0xA0, 0x34}}
	unsupported, err := BuildRTCMUnsupported(unsupportedInput)
	if err != nil {
		t.Fatal(err)
	}
	defer unsupported.Close()
	unsupportedMessage, err := unsupported.Message(0)
	if err != nil || unsupportedMessage.Unsupported == nil || !bytes.Equal(unsupportedMessage.Unsupported.Body, unsupportedInput.Body) {
		t.Fatalf("unsupported RTCM payload = %+v, %v", unsupportedMessage.Unsupported, err)
	}
}

func TestRTCMExtendedNetworkAndVTECTypedRoutes(t *testing.T) {
	auxiliaryInput := RTCMNetworkAuxiliaryStation{NetworkID: 1, SubnetworkID: 2, AuxiliaryStationCount: 1, MasterStationID: 12, AuxiliaryStationID: 13, DeltaLatitude: -17, DeltaLongitude: 29, DeltaHeight: -4, TrailingBits: RTCMTrailingBits{true, false}}
	auxiliary, err := BuildRTCMNetworkAuxiliaryStation(auxiliaryInput)
	if err != nil {
		t.Fatal(err)
	}
	defer auxiliary.Close()
	value, err := auxiliary.NetworkAuxiliaryStation(0)
	if err != nil || value.DeltaLatitude != auxiliaryInput.DeltaLatitude || value.AuxiliaryStationID != auxiliaryInput.AuxiliaryStationID {
		t.Fatalf("network auxiliary station = %+v, %v", value, err)
	}

	biasInput := RTCMGLONASSCodePhaseBiases{ReferenceStationID: 22, Aligned: true, HasL1CA: true, L1CA: -12, HasL2P: true, L2P: 37, TrailingBits: RTCMTrailingBits{true}}
	biases, err := BuildRTCMGLONASSCodePhaseBiases(biasInput)
	if err != nil {
		t.Fatal(err)
	}
	defer biases.Close()
	bias, err := biases.GLONASSCodePhaseBiases(0)
	if err != nil || !bias.HasL1CA || bias.L1CA != -12 || !bias.HasL2P || bias.L2P != 37 {
		t.Fatalf("GLONASS code/phase biases = %+v, %v", bias, err)
	}

	vtecInput := RTCMSSRVTECInfo{MessageNumber: 1264, EpochTimeS: 45, UpdateInterval: 2, IODSSR: 3, ProviderID: 4, SolutionID: 1, QualityIndicator: 6, Layers: []RTCMTecLayer{{Height: 10, Degree: 1, Order: 1, Cosine: []int16{7, -8}, Sine: []int16{9}}}, TrailingBits: RTCMTrailingBits{true, false}}
	vtec, err := BuildRTCMSSRVTEC(vtecInput)
	if err != nil {
		t.Fatal(err)
	}
	defer vtec.Close()
	decoded, err := vtec.SSRVTEC(0)
	if err != nil || len(decoded.Layers) != 1 || decoded.Layers[0].Cosine[1] != -8 || decoded.Layers[0].Sine[0] != 9 {
		t.Fatalf("SSR VTEC = %+v, %v", decoded, err)
	}
}

func TestRTCMExtendedSSRRetentionAndNetworkBuilders(t *testing.T) {
	ssrInfo := RTCMSSRInfoV2{MessageNumber: 1058, System: GNSSSystemGPS, Kind: RTCMSSRClock, Header: RTCMSSRHeader{EpochTimeS: 90, UpdateInterval: 2, IODSSR: 3, ProviderID: 4, SolutionID: 1, SatelliteCount: 1}, PaddingBits: RTCMTrailingBits{true, false}}
	ssr, err := BuildRTCMSSRMessageV2(ssrInfo, nil, []RTCMSSRClockRecord{{SatelliteID: 1, C0: -7, C1: 8, C2: -9}}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ssr.Close()
	retained, err := ssr.SSRInfoV2(0)
	if err != nil || retained.HasIGSSSRVersion != ssrInfo.HasIGSSSRVersion || retained.PaddingBits[0] != true || retained.PaddingBits[1] != false {
		t.Fatalf("SSR V2 metadata = %+v, %v", retained, err)
	}
	clocks, err := ssr.SSRClocks(0)
	if err != nil || len(clocks) != 1 || clocks[0].C0 != -7 || clocks[0].C2 != -9 {
		t.Fatalf("SSR clock records = %+v, %v", clocks, err)
	}

	differencesInput := RTCMNetworkDifferences{MessageNumber: 1015, NetworkID: 1, SubnetworkID: 2, EpochTime: 30, MasterStationID: 12, AuxiliaryStationID: 13, SatelliteCount: 1, Satellites: []RTCMNetworkDifference{{SatelliteID: 7, AmbiguityStatus: 1, NonSyncCount: 2, HasGeometric: true, Geometric: -13, HasIOD: true, IOD: 4, HasIonospheric: true, Ionospheric: 17}}}
	differences, err := BuildRTCMNetworkDifferences(differencesInput)
	if err != nil {
		t.Fatal(err)
	}
	defer differences.Close()
	decodedDifferences, err := differences.NetworkDifferences(0)
	if err != nil || len(decodedDifferences.Satellites) != 1 || decodedDifferences.Satellites[0].Ionospheric != 17 {
		t.Fatalf("network differences = %+v, %v", decodedDifferences, err)
	}

	residualInput := RTCMNetworkResiduals{MessageNumber: 1017, EpochTime: 40, ReferenceStationID: 12, ReferenceStationCount: 2, SatelliteCount: 1, Satellites: []RTCMNetworkResidual{{SatelliteID: 7, SOC: 1, SOD: 2, SOH: 3, SLC: 4, SLD: 5}}}
	residuals, err := BuildRTCMNetworkResiduals(residualInput)
	if err != nil {
		t.Fatal(err)
	}
	defer residuals.Close()
	decodedResiduals, err := residuals.NetworkResiduals(0)
	if err != nil || len(decodedResiduals.Satellites) != 1 || decodedResiduals.Satellites[0].SLD != 5 {
		t.Fatalf("network residuals = %+v, %v", decodedResiduals, err)
	}
}

func TestRTCMExtendedFKPGridAndTransformationRoutes(t *testing.T) {
	fkpInput := RTCMFKPGradients{MessageNumber: 1034, ReferenceStationID: 3, EpochTime: 19, SatelliteCount: 1, Satellites: []RTCMFKPGradient{{SatelliteID: 5, IOD: 6, GeometricNorth: -7, GeometricEast: 8, IonosphericNorth: -9, IonosphericEast: 10}}}
	fkp, err := BuildRTCMFKPGradients(fkpInput)
	if err != nil {
		t.Fatal(err)
	}
	defer fkp.Close()
	fkpMessage, err := fkp.Message(0)
	if err != nil || fkpMessage.FKPGradients == nil || fkpMessage.FKPGradients.Satellites[0].IonosphericEast != 10 {
		t.Fatalf("FKP message = %+v, %v", fkpMessage.FKPGradients, err)
	}

	gridInput := RTCMResidualGrid{MessageNumber: 1023, SystemID: 2, Origin1: -11, Origin2: 12, MeanHeightOffset: -13, Residuals: [16]RTCMGridResidual{{Horizontal1: -14, Horizontal2: 15, Height: -16}}}
	grid, err := BuildRTCMResidualGrid(gridInput)
	if err != nil {
		t.Fatal(err)
	}
	defer grid.Close()
	gridMessage, err := grid.Message(0)
	if err != nil || gridMessage.ResidualGrid == nil || gridMessage.ResidualGrid.Residuals[0].Height != -16 {
		t.Fatalf("residual grid = %+v, %v", gridMessage.ResidualGrid, err)
	}

	projectionInput := RTCMProjection{MessageNumber: 1025, SystemID: 1, ProjectionType: 2, Latitude: -17, Longitude: 18, AddScale: 19, Easting: 20, Northing: -21}
	projection, err := BuildRTCMProjection(projectionInput)
	if err != nil {
		t.Fatal(err)
	}
	defer projection.Close()
	projectionMessage, err := projection.Message(0)
	if err != nil || projectionMessage.Projection == nil || projectionMessage.Projection.Longitude != 18 {
		t.Fatalf("projection = %+v, %v", projectionMessage.Projection, err)
	}

	physicalInput := RTCMPhysicalReferenceStation{NonPhysicalStationID: 7, PhysicalStationID: 8, ITRFRealizationYear: 9, ECEFX: -10, ECEFY: 11, ECEFZ: -12}
	physical, err := BuildRTCMPhysicalReferenceStation(physicalInput)
	if err != nil {
		t.Fatal(err)
	}
	defer physical.Close()
	physicalMessage, err := physical.Message(0)
	if err != nil || physicalMessage.PhysicalReferenceStation == nil || physicalMessage.PhysicalReferenceStation.ECEFX != -10 {
		t.Fatalf("physical reference station = %+v, %v", physicalMessage.PhysicalReferenceStation, err)
	}

	helmertInput := RTCMHelmertTransformation{MessageNumber: 1022, SourceName: "A", TargetName: "B", DX: -22, HasRotationPoint: true, RotationPointX: 23, RotationPointY: -24, RotationPointZ: 25}
	helmert, err := BuildRTCMHelmertTransformation(helmertInput)
	if err != nil {
		t.Fatal(err)
	}
	defer helmert.Close()
	helmertMessage, err := helmert.Message(0)
	if err != nil || helmertMessage.HelmertTransformation == nil || helmertMessage.HelmertTransformation.RotationPointY != -24 {
		t.Fatalf("Helmert transformation = %+v, %v", helmertMessage.HelmertTransformation, err)
	}
}

func boolBytes(values []bool) []byte {
	result := make([]byte, len(values))
	for index, value := range values {
		if value {
			result[index] = 1
		}
	}
	return result
}
