package sidereon

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math"
	"os"
	"reflect"
	"sync"
	"testing"
)

func readSpaceFixture(t *testing.T, name string) []byte {
	t.Helper()
	value, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func fixtureHash(t *testing.T, name, want string) []byte {
	t.Helper()
	data := readSpaceFixture(t, name)
	hash := sha256.Sum256(data)
	if got := hex.EncodeToString(hash[:]); got != want {
		t.Fatalf("%s SHA-256 = %s, want %s", name, got, want)
	}
	return data
}

func assertSerializedHash(t *testing.T, name string, data []byte, want string) {
	t.Helper()
	hash := sha256.Sum256(data)
	if got := hex.EncodeToString(hash[:]); got != want {
		t.Fatalf("%s serializer SHA-256 = %s, want %s", name, got, want)
	}
}

func TestPublicFixturesArePinned(t *testing.T) {
	fixtures := map[string]string{
		"bias/edge.bia":       "d5b5dd38bbef7096fe1f3f5f3d1cf15db9acd37541cf45d6e19a9f9afbf2758d",
		"bias/P1C1_RINEX.DCB": "c71806425b2ad8c7798e1de17ab468bc07454492c3928bee95c621e05a6e860e",
		"bias/COD0OPSFIN_20261330000_01D_01D_OSB.BIA.gz": "37cdf3f14e32118bfe3b8d12945d240767dc3b3dd76ea0842928a13993393681",
		"oem/gps.kvn":                  "94352528a735af9d941335086ab36f776192b635ed590aadf39963bf8c9784aa",
		"oem/gps.xml":                  "b62b1bddcf0b9143ecba0ba55a779d45a2b0976e3c444b10493771ad09d00354",
		"omm/24876.kvn":                "99bd2ec09bc481d292b4bdc6d8dab0fea9a2d40c7ff1277c632c276e46cd7c66",
		"opm/osprey.kvn":               "4fffacabf0b5a6455e26b9234675886817309fcc00faead3465faaf734884512",
		"opm/osprey.xml":               "4ab71092fa2a7e5b648f704cb453b46b0a5dba810e681994a6916224fd01a180",
		"spk/horizons_eros_type21.bsp": "d2b6da88f5695e262f2576142444cb94405c21ce15290f7403b0dabab60441bb",
		"tdm/annex_e_18.kvn":           "3de252f3e4641fd529bc08336bdb9b939b2a9ec55515d2382022ea3c13a9821c",
	}
	for name, want := range fixtures {
		fixtureHash(t, name, want)
	}
}

func TestClockCore012Expectations(t *testing.T) {
	values := make([]float64, 12)
	for i := range values {
		values[i] = 1e-9 * float64((i+1)*(i+3))
	}
	series := AllanSeriesPhaseSeconds(values)
	values[0] = math.MaxFloat64
	factors := []int{1, 2}

	type expectedPoint struct {
		tau, deviation float64
		n              int
	}
	expected := map[AllanEstimator][]expectedPoint{
		AllanEstimatorADEV: {
			{1, 1.4142135623730968e-9, 10}, {2, 2.82842712474619e-9, 4},
		},
		AllanEstimatorOverlappingADEV: {
			{1, 1.4142135623730968e-9, 10}, {2, 2.82842712474619e-9, 8},
		},
		AllanEstimatorMDEV: {
			{1, 1.4142135623730968e-9, 10}, {2, 2.8284271247461906e-9, 7},
		},
		AllanEstimatorHDEV: {
			{1, 1.830211747325232e-23, 9}, {2, 3.2419899345169506e-24, 6},
		},
		AllanEstimatorTDEV: {
			{1, 8.164965809277271e-10, 10}, {2, 3.265986323710905e-9, 7},
		},
	}
	check := func(name string, result AllanResult, err error, want []expectedPoint) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if len(result.TauS) != len(want) || len(result.Deviation) != len(want) || len(result.N) != len(want) {
			t.Fatalf("%s lengths: %#v", name, result)
		}
		for i, point := range want {
			if result.TauS[i] != point.tau || result.Deviation[i] != point.deviation || result.N[i] != point.n {
				t.Fatalf("%s[%d] = (%g, %g, %d), want (%g, %g, %d)", name, i, result.TauS[i], result.Deviation[i], result.N[i], point.tau, point.deviation, point.n)
			}
		}
	}
	result, err := AllanDeviation(series, 1, factors)
	check("ADEV", result, err, expected[AllanEstimatorADEV])
	result, err = OverlappingADEV(series, 1, factors)
	check("overlapping ADEV", result, err, expected[AllanEstimatorOverlappingADEV])
	result, err = ModifiedADEV(series, 1, factors)
	check("MDEV", result, err, expected[AllanEstimatorMDEV])
	result, err = HadamardDeviation(series, 1, factors)
	check("HDEV", result, err, expected[AllanEstimatorHDEV])
	result, err = TimeDeviation(series, 1, factors)
	check("TDEV", result, err, expected[AllanEstimatorTDEV])

	options := AllanOptions{Estimators: AllanEstimatorSetStandard(), TauGrid: TauGridExplicit(factors), GapPolicy: GapPolicyReject}
	input := NewAllanInput(series, 1, &options)
	factors[0] = 99
	curves, err := ComputeAllanDeviations(input)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := curves.Close(); err != nil {
			t.Error(err)
		}
	}()
	for estimator, want := range expected {
		got, present, err := curves.Curve(estimator)
		if err != nil || !present {
			t.Fatalf("combined %v: present=%v err=%v", estimator, present, err)
		}
		check("combined", got, nil, want)
	}
	if _, _, err := curves.Curve(AllanEstimator(99)); err == nil {
		t.Fatal("invalid estimator was accepted")
	}
}

func TestClockPowerLawCap015Expectations(t *testing.T) {
	adev := AllanResult{TauS: []float64{1}, Deviation: []float64{1}, N: []int{1}}
	if got, err := AllanDeviationPowerLawSlope(PowerLawWhiteFM); err != nil || got != -0.5 {
		t.Fatalf("WhiteFM ADEV slope = %v, %v", got, err)
	}
	if got, err := ModifiedAllanDeviationPowerLawSlope(PowerLawWhiteFM); err != nil || got != -0.5 {
		t.Fatalf("WhiteFM MDEV slope = %v, %v", got, err)
	}
	if got, err := AllanVariancePowerLawTauExponent(PowerLawWhiteFM); err != nil || got != -1 {
		t.Fatalf("WhiteFM variance exponent = %v, %v", got, err)
	}
	options, err := DefaultPowerLawNoiseOptions(1, 0.5)
	if err != nil {
		t.Fatal(err)
	}
	options.SlopeTolerance = 1e-12
	options.ScatterTolerance = 1e-12
	fit, err := FitPowerLawNoise(adev, adev, &options)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := fit.Close(); err != nil {
			t.Error(err)
		}
	}()
	octaves, err := fit.Octaves()
	if err != nil || len(octaves) != 1 {
		t.Fatalf("power-law octaves = %#v, %v", octaves, err)
	}
	if octaves[0].Dominance != PowerLawFlagged || octaves[0].Flag != PowerLawOctaveUnderSampled || octaves[0].PointCount != 1 {
		t.Fatalf("power-law under-sampled octave = %#v", octaves[0])
	}
	regions, err := fit.Regions()
	if err != nil || len(regions) != 0 {
		t.Fatalf("power-law regions = %#v, %v", regions, err)
	}
	if err := fit.Close(); err != nil {
		t.Fatal(err)
	}
	if err := fit.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := fit.Coefficients(); !errors.Is(err, ErrClosed) {
		t.Fatalf("power-law after close = %v", err)
	}
}

func TestBiasPhaseBRecordsAndOwnership(t *testing.T) {
	data := fixtureHash(t, "bias/edge.bia", "d5b5dd38bbef7096fe1f3f5f3d1cf15db9acd37541cf45d6e19a9f9afbf2758d")
	set, err := ParseBiasSINEX(data)
	if err != nil {
		t.Fatal(err)
	}
	data[0] = 'x'
	if count, err := set.RecordCount(); err != nil || count != 11 {
		t.Fatalf("bias record count = %d, %v", count, err)
	}
	if count, err := set.SkippedRecordCount(); err != nil || count != 0 {
		t.Fatalf("bias skipped count = %d, %v", count, err)
	}
	if count, err := set.WarningCount(); err != nil || count != 2 {
		t.Fatalf("bias warning count = %d, %v", count, err)
	}
	mode, scale, err := set.Mode()
	if err != nil || mode != BiasModeAbsolute || scale != GPST {
		t.Fatalf("bias mode = %v, %v, %v", mode, scale, err)
	}
	records, err := set.Records()
	if err != nil {
		t.Fatal(err)
	}
	want := []struct {
		kind                                            BiasKind
		target                                          BiasTargetKind
		system                                          GNSSSystem
		sat, station, svn, obs1, obs2                   string
		hasSat, hasObs2, phase, hasSlope, hasSlopeSigma bool
		value, sigma, slope, slopeSigma                 float64
	}{
		{BiasKindOSB, BiasTargetSystem, GNSSSystemGPS, "", "", "G063", "C1C", "", false, false, false, false, false, 1.0000000000000002e-10, 1.0000000000000001e-11, 0, 0},
		{BiasKindOSB, BiasTargetSatellite, GNSSSystemGPS, "G01", "", "G063", "C1C", "", true, false, false, true, true, -1.23456789e-9, 2.0000000000000002e-11, 8.64e-10, 1.0000000000000001e-11},
		{BiasKindOSB, BiasTargetSatellite, GNSSSystemGPS, "G01", "", "G063", "C1W", "", true, false, false, false, false, 5.600000000000001e-10, 2.0000000000000002e-11, 0, 0},
		{BiasKindDSB, BiasTargetSatellite, GNSSSystemGPS, "G01", "", "G063", "C1C", "C1W", true, true, false, false, false, -1.79456789e-9, 3e-11, 0, 0},
		{BiasKindISB, BiasTargetSatellite, GNSSSystemGPS, "G01", "", "G063", "C1C", "C2W", true, true, false, false, false, 2.5e-10, 4.0000000000000004e-11, 0, 0},
		{BiasKindOSB, BiasTargetSatellite, GNSSSystemGPS, "G01", "", "G063", "L1C", "", true, false, true, false, false, -0.105, 0.01, 0, 0},
		{BiasKindOSB, BiasTargetReceiver, GNSSSystemGPS, "", "ALGO", "", "C1C", "", false, false, false, false, false, 3.1000000000000005e-9, 5.000000000000001e-11, 0, 0},
		{BiasKindOSB, BiasTargetReceiver, GNSSSystemGalileo, "", "ALGO", "", "C1C", "", false, false, false, false, false, 4.2e-9, 6e-11, 0, 0},
		{BiasKindOSB, BiasTargetSatelliteReceiver, GNSSSystemGPS, "G01", "ALGO", "G063", "C1C", "", true, false, false, false, false, 9.900000000000001e-9, 7.000000000000002e-11, 0, 0},
		{BiasKindOSB, BiasTargetSatellite, GNSSSystemGalileo, "E11", "", "E011", "C1C", "", true, false, false, false, false, 1.5000000000000002e-9, 2.0000000000000002e-11, 0, 0},
		{BiasKindOSB, BiasTargetSatellite, GNSSSystemGPS, "G01", "", "G063", "C2W", "", true, false, false, false, false, -3e-10, 2.0000000000000002e-11, 0, 0},
	}
	if len(records) != len(want) {
		t.Fatalf("records length = %d", len(records))
	}
	for i, expected := range want {
		got := records[i]
		if got.Kind != expected.kind || got.TargetKind != expected.target || got.System != expected.system || got.SatelliteID != expected.sat || got.Station != expected.station || got.SVN != expected.svn || got.Obs1 != expected.obs1 || got.Obs2 != expected.obs2 || got.HasSatelliteID != expected.hasSat || got.HasObs2 != expected.hasObs2 || got.IsPhase != expected.phase || got.HasSlope != expected.hasSlope || got.HasSlopeSigma != expected.hasSlopeSigma || got.Value != expected.value || got.Sigma != expected.sigma || got.Slope != expected.slope || got.SlopeSigma != expected.slopeSigma || !got.HasValidFrom || !got.HasValidUntil {
			t.Fatalf("bias record %d = %#v", i, got)
		}
		from, until := BiasEpoch{Year: 2020, DayOfYear: 1}, BiasEpoch{Year: 2020, DayOfYear: 2}
		if i == 10 {
			from, until = BiasEpoch{Year: 2020, DayOfYear: 2}, BiasEpoch{Year: 2020, DayOfYear: 4}
		}
		if got.ValidFrom != from || got.ValidUntil != until {
			t.Fatalf("bias record %d epochs = %#v", i, got)
		}
	}
	value, present, err := set.CodeOSBSeconds("G01", "C1C", BiasEpoch{Year: 2020, DayOfYear: 1})
	if err != nil || !present || value != -1.234567890000e-9 {
		t.Fatalf("code OSB lookup = %v, %v, %v", value, present, err)
	}
	value, present, err = set.PhaseOSBCycles("G01", "L1C", BiasEpoch{Year: 2020, DayOfYear: 1})
	if err != nil || !present || value != -0.105 {
		t.Fatalf("phase OSB lookup = %v, %v, %v", value, present, err)
	}
	value, present, err = set.CodeDSBSeconds("G01", "C1C", "C1W", BiasEpoch{Year: 2020, DayOfYear: 1})
	if err != nil || !present || value != -1.794567890000e-9 {
		t.Fatalf("code DSB lookup = %v, %v, %v", value, present, err)
	}
	_, present, err = set.CodeOSBSeconds("G01", "C1C", BiasEpoch{Year: 2020, DayOfYear: 10})
	if err != nil || present {
		t.Fatalf("missing lookup = present %v, err %v", present, err)
	}
	copyOfRecords := append([]BiasRecord(nil), records...)
	copyOfRecords[0].Station = "changed"
	again, err := set.Record(0)
	if err != nil || again.Station != "" {
		t.Fatalf("record output alias: %#v, %v", again, err)
	}

	lossy, err := ParseBiasSINEXLossy(readSpaceFixture(t, "bias/edge.bia"))
	if err != nil || lossy.SkipCount != 0 || lossy.WarningCount != 2 || lossy.Value == nil {
		t.Fatalf("lossy bias = %#v, %v", lossy, err)
	}
	if err := lossy.Close(); err != nil {
		t.Fatal(err)
	}
	if err := lossy.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := lossy.Value.RecordCount(); !errors.Is(err, ErrClosed) {
		t.Fatalf("lossy value after Close = %v", err)
	}
	if err := set.Close(); err != nil {
		t.Fatal(err)
	}
	if err := set.Close(); err != nil {
		t.Fatal(err)
	}
	if err := set.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := set.RecordCount(); !errors.Is(err, ErrClosed) {
		t.Fatalf("bias after Close = %v", err)
	}
}

func TestBiasGoOwnedPathAdaptersAndDCB(t *testing.T) {
	dcb := fixtureHash(t, "bias/P1C1_RINEX.DCB", "c71806425b2ad8c7798e1de17ab468bc07454492c3928bee95c621e05a6e860e")
	options := &CodeDCBOptions{Obs1: "P1", Obs2: "C1", Year: 2026, Month: 6, TimeScale: GPST}
	set, err := ParseCodeDCB(dcb, options)
	if err != nil {
		t.Fatal(err)
	}
	options.Obs1 = "mutated"
	value, present, err := set.CodeDSBSeconds("G01", "C1W", "C1C", BiasEpoch{Year: 2026, DayOfYear: 153})
	if err != nil || !present || value != 0.626e-9 {
		t.Fatalf("DCB lookup = %v, %v, %v", value, present, err)
	}
	_, present, err = set.CodeDSBSeconds("G01", "C1W", "C1C", BiasEpoch{Year: 2026, DayOfYear: 1})
	if err != nil || present {
		t.Fatalf("missing DCB lookup = %v, %v", present, err)
	}
	if err := set.Close(); err != nil {
		t.Fatal(err)
	}

	parsed, err := LoadBiasSINEXLossy("testdata/bias/COD0OPSFIN_20261330000_01D_01D_OSB.BIA.gz")
	if err != nil || parsed.Value == nil {
		t.Fatalf("gzip Go-owned path adapter = %#v, %v", parsed, err)
	}
	if count, err := parsed.Value.RecordCount(); err != nil || count == 0 {
		t.Fatalf("gzip record count = %d, %v", count, err)
	}
	if err := parsed.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := ParseCodeDCB(dcb, &CodeDCBOptions{TimeScale: TimeScale(99)}); err == nil {
		t.Fatal("invalid DCB time scale accepted")
	}
}

func TestCCSDSFixturesRoundTrip(t *testing.T) {
	oemKVN, err := ParseOEMKVN(fixtureHash(t, "oem/gps.kvn", "94352528a735af9d941335086ab36f776192b635ed590aadf39963bf8c9784aa"))
	if err != nil {
		t.Fatal(err)
	}
	if count, err := oemKVN.SegmentCount(); err != nil || count != 1 {
		t.Fatalf("OEM count = %d, %v", count, err)
	}
	oemOutput, err := oemKVN.ToKVN()
	if err != nil || len(oemOutput) == 0 {
		t.Fatalf("OEM output = %d, %v", len(oemOutput), err)
	}
	assertSerializedHash(t, "OEM KVN", oemOutput, "5ed7ed33eba24eaa0224521190185050c7c7abe58bc9aa673ed1aa388aac158c")
	oemRoundTrip, err := ParseOEMKVN(oemOutput)
	if err != nil {
		t.Fatal(err)
	}
	if err := oemRoundTrip.Close(); err != nil {
		t.Fatal(err)
	}
	oemOutput[0] ^= 0xff
	independentOEMOutput, err := oemKVN.ToKVN()
	if err != nil {
		t.Fatal(err)
	}
	assertSerializedHash(t, "independent OEM KVN", independentOEMOutput, "5ed7ed33eba24eaa0224521190185050c7c7abe58bc9aa673ed1aa388aac158c")
	if err := oemKVN.Close(); err != nil {
		t.Fatal(err)
	}
	if err := oemKVN.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := oemKVN.ToKVN(); !errors.Is(err, ErrClosed) {
		t.Fatalf("OEM after close: %v", err)
	}
	oemXML, err := ParseOEMXML(fixtureHash(t, "oem/gps.xml", "b62b1bddcf0b9143ecba0ba55a779d45a2b0976e3c444b10493771ad09d00354"))
	if err != nil {
		t.Fatal(err)
	}
	if output, err := oemXML.ToXML(); err != nil || len(output) == 0 {
		t.Fatalf("OEM XML output = %d, %v", len(output), err)
	} else {
		assertSerializedHash(t, "OEM XML", output, "d303c46f29fd6f53d7e2204251913ab6516b05a76e5e11b2daf38054c0aa376c")
		roundTrip, parseErr := ParseOEMXML(output)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		if closeErr := roundTrip.Close(); closeErr != nil {
			t.Fatal(closeErr)
		}
	}
	if err := oemXML.Close(); err != nil {
		t.Fatal(err)
	}

	opmKVN, err := ParseOPMKVN(fixtureHash(t, "opm/osprey.kvn", "4fffacabf0b5a6455e26b9234675886817309fcc00faead3465faaf734884512"))
	if err != nil {
		t.Fatal(err)
	}
	if output, err := opmKVN.ToXML(); err != nil || len(output) == 0 {
		t.Fatalf("OPM XML output = %d, %v", len(output), err)
	} else {
		assertSerializedHash(t, "OPM XML", output, "5845ee087531b9ab9dbf44e80ef66d80b913902f5e43b06780fefb40f217e362")
		roundTrip, parseErr := ParseOPMXML(output)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		if closeErr := roundTrip.Close(); closeErr != nil {
			t.Fatal(closeErr)
		}
	}
	opmOutput, err := opmKVN.ToKVN()
	if err != nil {
		t.Fatal(err)
	}
	assertSerializedHash(t, "OPM KVN", opmOutput, "25e9f6e96c9dbf50046bfbafb43628561b0717051fe2ab76ccadd564e5f754bd")
	opmRoundTrip, err := ParseOPMKVN(opmOutput)
	if err != nil {
		t.Fatal(err)
	}
	if err := opmRoundTrip.Close(); err != nil {
		t.Fatal(err)
	}
	if err := opmKVN.Close(); err != nil {
		t.Fatal(err)
	}
	if err := opmKVN.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := opmKVN.ToKVN(); !errors.Is(err, ErrClosed) {
		t.Fatalf("OPM after close: %v", err)
	}
	opmXML, err := ParseOPMXML(fixtureHash(t, "opm/osprey.xml", "4ab71092fa2a7e5b648f704cb453b46b0a5dba810e681994a6916224fd01a180"))
	if err != nil {
		t.Fatal(err)
	}
	if output, err := opmXML.ToKVN(); err != nil || len(output) == 0 {
		t.Fatalf("OPM KVN output = %d, %v", len(output), err)
	} else {
		assertSerializedHash(t, "OPM XML fixture to KVN", output, "25e9f6e96c9dbf50046bfbafb43628561b0717051fe2ab76ccadd564e5f754bd")
	}
	if err := opmXML.Close(); err != nil {
		t.Fatal(err)
	}

	omm, err := ParseOMMKVN(fixtureHash(t, "omm/24876.kvn", "99bd2ec09bc481d292b4bdc6d8dab0fea9a2d40c7ff1277c632c276e46cd7c66"))
	if err != nil {
		t.Fatal(err)
	}
	for name := range map[string]bool{"KVN": true, "XML": true, "JSON": true} {
		var output []byte
		switch name {
		case "KVN":
			output, err = omm.ToKVN()
		case "XML":
			output, err = omm.ToXML()
		case "JSON":
			output, err = omm.ToJSON()
		}
		if err != nil || len(output) == 0 {
			t.Fatalf("OMM %s output = %d, %v", name, len(output), err)
		}
		wantHash := map[string]string{"KVN": "e20c3848984d9df2556d78f43040092a352699cce5b9397da71a0615d271e51c", "XML": "3f6ab03c6ce8aa09396399d7081b88388014d1b6e46940c5293fb57dc1b745fc", "JSON": "79c3c753fb68ac76c4cca7ae53d3ac1eed35b21baa2a2a3f1b7502689151d014"}
		assertSerializedHash(t, "OMM "+name, output, wantHash[name])
		var roundTrip *OMM
		switch name {
		case "KVN":
			roundTrip, err = ParseOMMKVN(output)
		case "XML":
			roundTrip, err = ParseOMMXML(output)
		case "JSON":
			roundTrip, err = ParseOMMJSON(output)
		}
		if err != nil {
			t.Fatalf("OMM %s round trip: %v", name, err)
		}
		if err := roundTrip.Close(); err != nil {
			t.Fatal(err)
		}
	}
	if err := omm.Close(); err != nil {
		t.Fatal(err)
	}
	if err := omm.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := omm.ToJSON(); !errors.Is(err, ErrClosed) {
		t.Fatalf("OMM after close: %v", err)
	}

	tdm, err := ParseTDMKVN(fixtureHash(t, "tdm/annex_e_18.kvn", "3de252f3e4641fd529bc08336bdb9b939b2a9ec55515d2382022ea3c13a9821c"))
	if err != nil {
		t.Fatal(err)
	}
	if segments, err := tdm.SegmentCount(); err != nil || segments != 2 {
		t.Fatalf("TDM segments = %d, %v", segments, err)
	}
	if records, err := tdm.RecordCount(); err != nil || records != 20 {
		t.Fatalf("TDM records = %d, %v", records, err)
	}
	participants, err := tdm.Participants()
	if err != nil || len(participants) != 4 {
		t.Fatalf("TDM participants = %d, %v", len(participants), err)
	}
	paths, err := tdm.Paths()
	if err != nil || len(paths) != 2 {
		t.Fatalf("TDM paths = %d, %v", len(paths), err)
	}
	rows, err := tdm.Records()
	if err != nil || len(rows) != 20 {
		t.Fatalf("TDM rows = %d, %v", len(rows), err)
	}
	if rows[0].Keyword != "TRANSMIT_PHASE_CT_1" || rows[0].Observable != TDMObservableOther || rows[0].ValueText == "" {
		t.Fatalf("TDM first record = %#v", rows[0])
	}
	if participants[0] != (TDMParticipant{SegmentIndex: 0, Index: 1, Name: "DSS-55"}) || participants[1] != (TDMParticipant{SegmentIndex: 0, Index: 2, Name: "yyyy-nnnA"}) {
		t.Fatalf("TDM participants = %#v", participants)
	}
	if !reflect.DeepEqual(paths[0], TDMPath{SegmentIndex: 0, Key: "PATH", Participants: []uint8{1, 2, 1}}) {
		t.Fatalf("TDM path = %#v", paths[0])
	}
	paths[0].Participants[0] = 9
	pathsAgain, err := tdm.Paths()
	if err != nil || !reflect.DeepEqual(pathsAgain[0], TDMPath{SegmentIndex: 0, Key: "PATH", Participants: []uint8{1, 2, 1}}) {
		t.Fatalf("TDM path output alias = %#v, %v", pathsAgain, err)
	}
	segments, err := tdm.Segments()
	if err != nil {
		t.Fatal(err)
	}
	wantSegments := []TDMSegmentSummary{{SegmentIndex: 0, Mode: TDMStringField{Present: true, Value: "SEQUENTIAL"}, TimetagRef: TDMStringField{}, TimeSystem: TDMStringField{Present: true, Value: "UTC"}, RangeUnit: TDMUnitKilometers, ParticipantCount: 2, PathCount: 1, RecordCount: 10}, {SegmentIndex: 1, Mode: TDMStringField{Present: true, Value: "SEQUENTIAL"}, TimetagRef: TDMStringField{}, TimeSystem: TDMStringField{Present: true, Value: "UTC"}, RangeUnit: TDMUnitKilometers, ParticipantCount: 2, PathCount: 1, RecordCount: 10}}
	if !reflect.DeepEqual(segments, wantSegments) {
		t.Fatalf("TDM segments = %#v", segments)
	}
	if rows[0] != (TDMDataRecord{SegmentIndex: 0, Observable: TDMObservableOther, Unit: TDMUnitDimensionless, Keyword: "TRANSMIT_PHASE_CT_1", Epoch: "2005-184T11:12:23", ValueText: "7175173383.615373", Value: 7175173383.615373}) {
		t.Fatalf("TDM first record = %#v", rows[0])
	}
	if output, err := tdm.ToKVN(); err != nil || len(output) == 0 {
		t.Fatalf("TDM output = %d, %v", len(output), err)
	} else {
		assertSerializedHash(t, "TDM KVN", output, "53b5837c284d017e59a8b6fabdb8665f77bd8f3a469d3958f732bf9ac03bcb38")
		roundTrip, parseErr := ParseTDMKVN(output)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		if count, countErr := roundTrip.RecordCount(); countErr != nil || count != 20 {
			t.Fatalf("TDM round-trip count = %d, %v", count, countErr)
		}
		if closeErr := roundTrip.Close(); closeErr != nil {
			t.Fatal(closeErr)
		}
	}
	if err := tdm.Close(); err != nil {
		t.Fatal(err)
	}
	if err := tdm.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := tdm.RecordCount(); !errors.Is(err, ErrClosed) {
		t.Fatalf("TDM after close: %v", err)
	}
}

func TestSPKPhaseBFixture(t *testing.T) {
	spk, err := LoadSPK(fixtureHash(t, "spk/horizons_eros_type21.bsp", "d2b6da88f5695e262f2576142444cb94405c21ce15290f7403b0dabab60441bb"))
	if err != nil {
		t.Fatal(err)
	}
	state, err := spk.State(20000433, 10, 757339200)
	if err != nil {
		t.Fatal(err)
	}
	if state.Target != 20000433 || state.Center != 10 || !state.HasVelocityKmPerS {
		t.Fatalf("SPK metadata = %#v", state)
	}
	wantPosition := [3]float64{198083634.33689928, 56306354.00566181, 67761020.0290685}
	wantVelocity := [3]float64{-14.136880898003753, 18.729945253375007, 8.080580941541488}
	for i := range wantPosition {
		if math.Abs(state.PositionKm[i]-wantPosition[i]) > 5e-8 || math.Abs(state.VelocityKmPerS[i]-wantVelocity[i]) > 1e-14 {
			t.Fatalf("SPK state[%d] = %#v", i, state)
		}
	}
	if err := spk.Close(); err != nil {
		t.Fatal(err)
	}
	if err := spk.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := spk.State(20000433, 10, 757339200); !errors.Is(err, ErrClosed) {
		t.Fatalf("SPK after close: %v", err)
	}
}

const ommCatalogFeed = `[{"OBJECT_NAME":"GPS BIIF-8  (PRN 03)","EPOCH":"2020-06-25T00:00:00.000000","MEAN_MOTION":2.0056,"ECCENTRICITY":0.0001,"INCLINATION":55.0,"RA_OF_ASC_NODE":100.0,"ARG_OF_PERICENTER":50.0,"MEAN_ANOMALY":10.0,"NORAD_CAT_ID":40294,"BSTAR":0.0,"MEAN_MOTION_DOT":0.0,"MEAN_MOTION_DDOT":0.0},{"OBJECT_NAME":"GPS BIII-1  (PRN 04)","EPOCH":"2020-06-25T00:00:00.000000","MEAN_MOTION":2.0056,"ECCENTRICITY":0.0001,"INCLINATION":55.0,"RA_OF_ASC_NODE":100.0,"ARG_OF_PERICENTER":50.0,"MEAN_ANOMALY":10.0,"NORAD_CAT_ID":43873,"BSTAR":0.0,"MEAN_MOTION_DOT":0.0,"MEAN_MOTION_DDOT":0.0},{"OBJECT_NAME":"QZS-2 (QZSS/PRN 194)","EPOCH":"2020-06-25T00:00:00.000000","MEAN_MOTION":2.0056,"ECCENTRICITY":0.0001,"INCLINATION":55.0,"RA_OF_ASC_NODE":100.0,"ARG_OF_PERICENTER":50.0,"MEAN_ANOMALY":10.0,"NORAD_CAT_ID":42738,"BSTAR":0.0,"MEAN_MOTION_DOT":0.0,"MEAN_MOTION_DDOT":0.0}]`

func TestOMMCatalogNewGapsExpectations(t *testing.T) {
	feed := []byte(ommCatalogFeed)
	catalog, err := BuildOMMCatalogLenient(GNSSSystemGPS, feed)
	if err != nil {
		t.Fatal(err)
	}
	feed[0] = 'x'
	if n, err := catalog.RecordCount(); err != nil || n != 2 {
		t.Fatalf("catalog records = %d, %v", n, err)
	}
	if n, err := catalog.SkippedCount(); err != nil || n != 1 {
		t.Fatalf("catalog skipped = %d, %v", n, err)
	}
	if n, err := catalog.MalformedCount(); err != nil || n != 0 {
		t.Fatalf("catalog malformed = %d, %v", n, err)
	}
	records, err := catalog.Records()
	if err != nil {
		t.Fatal(err)
	}
	want := []ConstellationRecord{{System: GNSSSystemGPS, PRN: 3, NORADID: 40294, Active: true, Usable: true}, {System: GNSSSystemGPS, PRN: 4, NORADID: 43873, Active: true, Usable: true}}
	if !reflect.DeepEqual(records, want) {
		t.Fatalf("catalog records = %#v, want %#v", records, want)
	}
	skipped, err := catalog.SkippedEntries()
	if err != nil || len(skipped) != 1 {
		t.Fatalf("catalog skipped entries = %#v, %v", skipped, err)
	}
	if skipped[0].NORADID != 42738 || !skipped[0].ObjectNamePresent || skipped[0].ObjectName != "QZS-2 (QZSS/PRN 194)" {
		t.Fatalf("catalog skipped = %#v", skipped[0])
	}
	if _, err := catalog.Record(-1); err == nil {
		t.Fatal("negative catalog index accepted")
	}
	if err := catalog.Close(); err != nil {
		t.Fatal(err)
	}
	if err := catalog.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := catalog.RecordCount(); !errors.Is(err, ErrClosed) {
		t.Fatalf("catalog after close: %v", err)
	}

	bad := []byte(`[42,{"OBJECT_NAME":"GPS BIIF-8  (PRN 03)","EPOCH":"2020-06-25T00:00:00.000000","MEAN_MOTION":2.0056,"ECCENTRICITY":0.0001,"INCLINATION":55.0,"RA_OF_ASC_NODE":100.0,"ARG_OF_PERICENTER":50.0,"MEAN_ANOMALY":10.0,"NORAD_CAT_ID":40294,"BSTAR":0.0,"MEAN_MOTION_DOT":0.0,"MEAN_MOTION_DDOT":0.0}]`)
	badCatalog, err := BuildOMMCatalogLenient(GNSSSystemGPS, bad)
	if err != nil {
		t.Fatal(err)
	}
	if n, err := badCatalog.RecordCount(); err != nil || n != 1 {
		t.Fatalf("bad catalog records = %d, %v", n, err)
	}
	if n, err := badCatalog.MalformedCount(); err != nil || n != 1 {
		t.Fatalf("bad catalog malformed = %d, %v", n, err)
	}
	if err := badCatalog.Close(); err != nil {
		t.Fatal(err)
	}
	if err := badCatalog.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := badCatalog.RecordCount(); !errors.Is(err, ErrClosed) {
		t.Fatalf("bad catalog after close: %v", err)
	}
	if _, err := BuildOMMCatalogLenient(GNSSSystem(99), []byte(ommCatalogFeed)); err == nil {
		t.Fatal("invalid catalog system accepted")
	}
}

func TestOwningHandlesReadCloseRace(t *testing.T) {
	opm, err := ParseOPMKVN(readSpaceFixture(t, "opm/osprey.kvn"))
	if err != nil {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	group.Add(2)
	go func() {
		defer group.Done()
		for i := 0; i < 100; i++ {
			_, _ = opm.ToKVN()
		}
	}()
	go func() { defer group.Done(); _ = opm.Close(); _ = opm.Close() }()
	group.Wait()
	if _, err := opm.ToKVN(); !errors.Is(err, ErrClosed) {
		t.Fatalf("race handle after close: %v", err)
	}

	curves, err := ComputeAllanDeviations(NewAllanInput(AllanSeriesPhaseSeconds([]float64{0, 1, 2, 3, 4, 5}), 1, &AllanOptions{Estimators: AllanEstimatorSetStandard(), TauGrid: TauGridExplicit([]int{1}), GapPolicy: GapPolicyReject}))
	if err != nil {
		t.Fatal(err)
	}
	group.Add(2)
	go func() {
		defer group.Done()
		for i := 0; i < 100; i++ {
			_, _, _ = curves.Curve(AllanEstimatorADEV)
		}
	}()
	go func() { defer group.Done(); _ = curves.Close(); _ = curves.Close() }()
	group.Wait()
	if _, _, err := curves.Curve(AllanEstimatorADEV); !errors.Is(err, ErrClosed) {
		t.Fatalf("curves after close: %v", err)
	}
}

func assertReadCloseRace(t *testing.T, name string, read func() error, close func() error) {
	t.Helper()
	var group sync.WaitGroup
	group.Add(2)
	go func() {
		defer group.Done()
		for i := 0; i < 100; i++ {
			if err := read(); err != nil && !errors.Is(err, ErrClosed) {
				t.Errorf("%s read: %v", name, err)
			}
		}
	}()
	go func() {
		defer group.Done()
		if err := close(); err != nil {
			t.Errorf("%s close: %v", name, err)
		}
		if err := close(); err != nil {
			t.Errorf("%s repeated close: %v", name, err)
		}
	}()
	group.Wait()
	if err := close(); err != nil {
		t.Fatalf("%s final close: %v", name, err)
	}
	if err := read(); !errors.Is(err, ErrClosed) {
		t.Fatalf("%s read after close = %v", name, err)
	}
}

func TestAllOwningHandleReadCloseContracts(t *testing.T) {
	bias, err := ParseBiasSINEX(readSpaceFixture(t, "bias/edge.bia"))
	if err != nil {
		t.Fatal(err)
	}
	assertReadCloseRace(t, "bias", func() error { _, err := bias.RecordCount(); return err }, bias.Close)

	adev := AllanResult{TauS: []float64{1}, Deviation: []float64{1}, N: []int{1}}
	fit, err := FitPowerLawNoise(adev, adev, nil)
	if err != nil {
		t.Fatal(err)
	}
	assertReadCloseRace(t, "power-law fit", func() error { _, err := fit.Coefficients(); return err }, fit.Close)

	oem, err := ParseOEMKVN(readSpaceFixture(t, "oem/gps.kvn"))
	if err != nil {
		t.Fatal(err)
	}
	assertReadCloseRace(t, "OEM", func() error { _, err := oem.SegmentCount(); return err }, oem.Close)

	omm, err := ParseOMMKVN(readSpaceFixture(t, "omm/24876.kvn"))
	if err != nil {
		t.Fatal(err)
	}
	assertReadCloseRace(t, "OMM", func() error { _, err := omm.ToJSON(); return err }, omm.Close)

	opm, err := ParseOPMKVN(readSpaceFixture(t, "opm/osprey.kvn"))
	if err != nil {
		t.Fatal(err)
	}
	assertReadCloseRace(t, "OPM", func() error { _, err := opm.ToKVN(); return err }, opm.Close)

	spk, err := LoadSPK(readSpaceFixture(t, "spk/horizons_eros_type21.bsp"))
	if err != nil {
		t.Fatal(err)
	}
	assertReadCloseRace(t, "SPK", func() error { _, err := spk.State(20000433, 10, 757339200); return err }, spk.Close)

	tdm, err := ParseTDMKVN(readSpaceFixture(t, "tdm/annex_e_18.kvn"))
	if err != nil {
		t.Fatal(err)
	}
	assertReadCloseRace(t, "TDM", func() error { _, err := tdm.RecordCount(); return err }, tdm.Close)

	catalog, err := BuildOMMCatalogLenient(GNSSSystemGPS, []byte(ommCatalogFeed))
	if err != nil {
		t.Fatal(err)
	}
	assertReadCloseRace(t, "OMM catalog", func() error { _, err := catalog.RecordCount(); return err }, catalog.Close)
}

func TestCodecBoundaryValidation(t *testing.T) {
	if _, err := ParseOMMKVN([]byte{'x', 0, 'y'}); err == nil {
		t.Fatal("embedded NUL accepted")
	}
	bias, err := ParseBiasSINEX(readSpaceFixture(t, "bias/edge.bia"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if closeErr := bias.Close(); closeErr != nil {
			t.Errorf("close bias set: %v", closeErr)
		}
	})
	if _, _, err := bias.CodeOSBSeconds("G\x0001", "C1C", BiasEpoch{}); err == nil {
		t.Fatal("embedded NUL in bias lookup accepted")
	}
	if _, err := ParseCodeDCB([]byte("x"), &CodeDCBOptions{Obs1: "C\x001C"}); err == nil {
		t.Fatal("embedded NUL in DCB options accepted")
	}
	if _, err := AllanDeviation(AllanSeriesFractionalFrequency([]float64{1, 2}), 1, []int{-1}); err == nil {
		t.Fatal("negative averaging factor accepted")
	}
	if _, err := AllanDeviation(AllanSeriesFractionalFrequency([]float64{1, 2}), 1, []int{0}); err == nil {
		t.Fatal("zero averaging factor accepted")
	}
	if _, err := LoadSPK([]byte("not a kernel")); err == nil {
		t.Fatal("malformed SPK accepted")
	}
	if _, err := FitPowerLawNoise(AllanResult{TauS: []float64{1}, Deviation: []float64{1}, N: []int{-1}}, AllanResult{TauS: []float64{1}, Deviation: []float64{1}, N: []int{1}}, nil); err == nil {
		t.Fatal("negative Allan term count accepted")
	}
	if _, err := ParseCodeDCB([]byte("x"), &CodeDCBOptions{TimeScale: TimeScale(99)}); err == nil {
		t.Fatal("invalid time scale accepted")
	}
	if _, err := AllanDeviationPowerLawSlope(PowerLawNoiseType(99)); err == nil {
		t.Fatal("invalid power-law type accepted")
	}
}
