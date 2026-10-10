package sidereon

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"math"
	"os"
	"reflect"
	"strings"
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
		"bias/edge.bia":       "314d1699ae152a0155970d6ad4633d8859702b7a3dfa2cf12c88ffff3f25f886",
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
	data := fixtureHash(t, "bias/edge.bia", "314d1699ae152a0155970d6ad4633d8859702b7a3dfa2cf12c88ffff3f25f886")
	set, err := ParseBiasSINEX(data)
	if err != nil {
		t.Fatal(err)
	}
	textOutput, err := set.BiasSINEXText()
	if err != nil {
		t.Fatalf("write Bias-SINEX text: %v", err)
	}
	byteOutput, err := set.BiasSINEXBytes()
	if err != nil || !bytes.Equal([]byte(textOutput), byteOutput) {
		t.Fatalf("Bias-SINEX text/bytes differ: %v", err)
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
	if count, err := set.NoticeCount(); err != nil || count != 0 {
		t.Fatalf("bias notice count = %d, %v", count, err)
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
		if !got.HasLine || got.Line != uint64(16+i) {
			t.Fatalf("bias record %d source line = %d, present=%v", i, got.Line, got.HasLine)
		}
		expectedFamily, expectedUnit := BiasObservableCode, BiasUnitNanoseconds
		if i == 5 {
			expectedFamily, expectedUnit = BiasObservablePhase, BiasUnitCycles
		}
		if got.Family != expectedFamily || got.Unit != expectedUnit {
			t.Fatalf("bias record %d family/unit = %v/%v", i, got.Family, got.Unit)
		}
	}
	value, present, err := set.CodeOSBSeconds("G01", "C1C", BiasEpoch{Year: 2020, DayOfYear: 1})
	if err != nil || !present || value != -3.732603456789e-5 {
		t.Fatalf("code OSB lookup = %v, %v, %v", value, present, err)
	}
	value, present, err = set.CodeOSBSeconds("G01", "C1C", BiasEpoch{Year: 2020, DayOfYear: 1, SecondOfDay: 43200})
	if err != nil || !present || value != -1.234567890000e-9 {
		t.Fatalf("code OSB midpoint lookup = %v, %v, %v", value, present, err)
	}
	value, present, err = set.PhaseOSBCycles("G01", "L1C", BiasEpoch{Year: 2020, DayOfYear: 1})
	if err != nil || !present || value != -0.105 {
		t.Fatalf("phase OSB lookup = %v, %v, %v", value, present, err)
	}
	value, present, err = set.CodeDSBSeconds("G01", "C1C", "C1W", BiasEpoch{Year: 2020, DayOfYear: 1})
	if err != nil || !present || value != -1.794567890000e-9 {
		t.Fatalf("code DSB lookup = %v, %v, %v", value, present, err)
	}
	fullCode, err := set.CodeOSBLookup("G01", "C1C", BiasEpoch{Year: 2020, DayOfYear: 1})
	if err != nil || fullCode.Status != BiasLookupAvailable || fullCode.Value != -3.732603456789e-5 || len(fullCode.RecordIndices) == 0 {
		t.Fatalf("full code OSB result = %+v, %v", fullCode, err)
	}
	fullCodeMidpoint, err := set.CodeOSBLookup("G01", "C1C", BiasEpoch{Year: 2020, DayOfYear: 1, SecondOfDay: 43200})
	if err != nil || fullCodeMidpoint.Status != BiasLookupAvailable || fullCodeMidpoint.Value != -1.234567890000e-9 || len(fullCodeMidpoint.RecordIndices) == 0 {
		t.Fatalf("full code OSB midpoint result = %+v, %v", fullCodeMidpoint, err)
	}
	fullPhase, err := set.PhaseOSBLookup("G01", "L1C", BiasEpoch{Year: 2020, DayOfYear: 1}, false, 0)
	if err != nil || fullPhase.Status != BiasLookupAvailable || fullPhase.Value != -0.105 || len(fullPhase.RecordIndices) == 0 {
		t.Fatalf("full phase OSB result = %+v, %v", fullPhase, err)
	}
	fullDSB, err := set.CodeDSBLookup("G01", "C1C", "C1W", BiasEpoch{Year: 2020, DayOfYear: 1})
	if err != nil || fullDSB.Status != BiasLookupAvailable || fullDSB.Value != -1.794567890000e-9 || len(fullDSB.RecordIndices) == 0 {
		t.Fatalf("full code DSB result = %+v, %v", fullDSB, err)
	}
	fullMissing, err := set.CodeOSBLookup("G01", "C1C", BiasEpoch{Year: 2020, DayOfYear: 10})
	if err != nil || fullMissing.Status != BiasLookupAbsent {
		t.Fatalf("full missing result = %+v, %v", fullMissing, err)
	}
	modeInfo, err := set.ModeInfo()
	if err != nil || !modeInfo.HasTimeScale || modeInfo.Mode == BiasModeUnspecified {
		t.Fatalf("bias mode presence = %+v, %v", modeInfo, err)
	}
	_, present, err = set.CodeOSBSeconds("G01", "C1C", BiasEpoch{Year: 2020, DayOfYear: 10})
	if err != nil || present {
		t.Fatalf("missing lookup = present %v, err %v", present, err)
	}
	_, present, err = set.CodeOSBSeconds("", "C1C", BiasEpoch{Year: 2020, DayOfYear: 1})
	if err == nil || present {
		t.Fatalf("invalid satellite lookup = present %v, err %v", present, err)
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
	if _, present, err := set.CodeOSBSeconds("G01", "C1C", BiasEpoch{Year: 2020, DayOfYear: 1}); present || !errors.Is(err, ErrClosed) {
		t.Fatalf("legacy bias lookup after Close = present %v, err %v", present, err)
	}
	if fullCode.Status != BiasLookupAvailable || fullCode.Value != -3.732603456789e-5 || len(fullCode.RecordIndices) == 0 {
		t.Fatalf("owned code OSB start result after Close = %+v", fullCode)
	}
	if fullCodeMidpoint.Status != BiasLookupAvailable || fullCodeMidpoint.Value != -1.234567890000e-9 || len(fullCodeMidpoint.RecordIndices) == 0 {
		t.Fatalf("owned code OSB midpoint result after Close = %+v", fullCodeMidpoint)
	}
}

func TestBiasLookupRetainsOverriddenRecordIndicesAfterClose(t *testing.T) {
	data := string(readSpaceFixture(t, "bias/edge.bia"))
	data = strings.Replace(data, "00000011", "00000013", 1)
	rows := " OSB  G063 G01           C1C       2020:001:00000 2020:004:00000 ns      1.000000000000E+00 2.00000E-02\n" +
		" OSB  G063 G01           C1C       2020:002:00000 2020:004:00000 ns      2.000000000000E+00 2.00000E-02\n"
	data = strings.Replace(data, "-BIAS/SOLUTION", rows+"-BIAS/SOLUTION", 1)
	set, err := ParseBiasSINEX([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	lookup, err := set.CodeOSBLookup("G01", "C1C", BiasEpoch{Year: 2020, DayOfYear: 2})
	if err != nil || lookup.Status != BiasLookupAvailable || lookup.Value != 2e-9 || !reflect.DeepEqual(lookup.RecordIndices, []uint64{12}) || !reflect.DeepEqual(lookup.OverriddenRecordIndices, []uint64{11}) {
		t.Fatalf("overridden lookup = %+v, %v", lookup, err)
	}
	if err := set.Close(); err != nil {
		t.Fatal(err)
	}
	if lookup.Status != BiasLookupAvailable || lookup.Value != 2e-9 || !reflect.DeepEqual(lookup.RecordIndices, []uint64{12}) || !reflect.DeepEqual(lookup.OverriddenRecordIndices, []uint64{11}) {
		t.Fatalf("owned overridden lookup after Close = %+v", lookup)
	}
}

func TestBiasLookupRetainsAmbiguousRecordIndices(t *testing.T) {
	data := string(readSpaceFixture(t, "bias/edge.bia"))
	data = strings.Replace(data, "00000011", "00000013", 1)
	rows := " OSB  G063 G01           C1C       2020:002:00000 2020:004:00000 ns      1.000000000000E+00 2.00000E-02\n" +
		" OSB  G063 G01           C1C       2020:002:00000 2020:004:00000 ns      2.000000000000E+00 2.00000E-02\n"
	data = strings.Replace(data, "-BIAS/SOLUTION", rows+"-BIAS/SOLUTION", 1)
	set, err := ParseBiasSINEX([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, set)
	lookup, err := set.CodeOSBLookup("G01", "C1C", BiasEpoch{Year: 2020, DayOfYear: 2})
	if err != nil || lookup.Status != BiasLookupAmbiguous || !reflect.DeepEqual(lookup.RecordIndices, []uint64{11, 12}) {
		t.Fatalf("ambiguous lookup = %+v, %v", lookup, err)
	}
}

func TestBiasDepartureNoticeHasTypedFields(t *testing.T) {
	data := readSpaceFixture(t, "bias/edge.bia")
	data = bytes.Replace(data, []byte("%=BIA 1.00"), []byte("%=BIA 1.01"), 1)
	_, strictErr := ParseBiasSINEX(data)
	var typedErr *BiasError
	if !errors.As(strictErr, &typedErr) || typedErr.Kind != BiasErrorUnsupportedVersion {
		t.Fatalf("typed unsupported-version error = %#v, %v", typedErr, strictErr)
	}
	_, policyStrictErr := ParseBiasSINEXWithPolicy(data, BiasReadPolicyStrict)
	var policyTypedErr *BiasError
	if !errors.As(policyStrictErr, &policyTypedErr) || policyTypedErr.Kind != BiasErrorUnsupportedVersion {
		t.Fatalf("explicit strict policy error = %#v, %v", policyTypedErr, policyStrictErr)
	}
	if _, err := ParseBiasSINEXWithPolicy(data, BiasReadPolicy(99)); err == nil {
		t.Fatal("unknown Bias-SINEX policy accepted")
	}
	version, err := typedErr.TextBytes(BiasErrorTextVersion, 0)
	if err != nil || string(version) != "1.01" {
		t.Fatalf("unsupported-version bytes = %q, %v", version, err)
	}
	parsed, err := ParseBiasSINEXWithPolicy(data, BiasReadPolicyLenient)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, parsed)
	if parsed.Value == nil {
		t.Fatal("lossy parser returned no bias set")
	}
	noticeCount, err := parsed.Value.NoticeCount()
	if err != nil || noticeCount == 0 {
		t.Fatalf("departure notice count = %d, %v", noticeCount, err)
	}
	var found bool
	for index := 0; index < noticeCount; index++ {
		notice, noticeErr := parsed.Value.Notice(index)
		if noticeErr != nil {
			t.Fatal(noticeErr)
		}
		if notice.Kind != BiasNoticeDeparture || notice.Departure != BiasDepartureOtherVersion {
			continue
		}
		version, textErr := parsed.Value.NoticeText(index, BiasNoticeTextVersion)
		if textErr != nil || version != "1.01" {
			t.Fatalf("departure version text = %q, %v", version, textErr)
		}
		if notice.HasLine || notice.Line != 0 {
			t.Fatalf("departure line fields = %+v", notice)
		}
		found = true
		break
	}
	if !found {
		t.Fatal("typed OtherVersion departure notice was not exposed")
	}
}

func TestBiasGoOwnedPathAdaptersAndDCB(t *testing.T) {
	dcb := fixtureHash(t, "bias/P1C1_RINEX.DCB", "c71806425b2ad8c7798e1de17ab468bc07454492c3928bee95c621e05a6e860e")
	options := &CodeDCBOptions{Obs1: "P1", Obs2: "C1", Year: 2026, Month: 6, TimeScale: GPST}
	set, err := ParseCodeDCB(dcb, options)
	if err != nil {
		t.Fatal(err)
	}
	dcbText, err := set.CodeDCBText()
	if err != nil {
		t.Fatalf("write CODE DCB text: %v", err)
	}
	dcbBytes, err := set.CodeDCBBytes()
	if err != nil || !bytes.Equal([]byte(dcbText), dcbBytes) {
		t.Fatalf("CODE DCB text/bytes differ: %v", err)
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
	policyOptions := &CodeDCBOptions{Obs1: "P1", Obs2: "C1", Year: 2026, Month: 6, TimeScale: GPST}
	policySet, err := ParseCodeDCBWithPolicy(dcb, policyOptions, BiasReadPolicyStrict)
	if err != nil || policySet.Value == nil {
		t.Fatalf("explicit strict CODE DCB policy = %#v, %v", policySet, err)
	}
	if count, err := policySet.Value.RecordCount(); err != nil || count == 0 {
		t.Fatalf("explicit-policy CODE DCB records = %d, %v", count, err)
	}
	if err := policySet.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := ParseCodeDCBWithPolicy(dcb, policyOptions, BiasReadPolicy(99)); err == nil {
		t.Fatal("unknown CODE DCB read policy accepted")
	}
	loaded, err := LoadBiasSINEXWithPolicy("testdata/bias/edge.bia", BiasReadPolicyStrict)
	if err != nil || loaded.Value == nil {
		t.Fatalf("explicit-policy Bias-SINEX load = %#v, %v", loaded, err)
	}
	if count, err := loaded.Value.RecordCount(); err != nil || count == 0 {
		t.Fatalf("explicit-policy loaded records = %d, %v", count, err)
	}
	if err := loaded.Close(); err != nil {
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

	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	if _, err := writer.Write(dcb); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	compressedPath := t.TempDir() + "/sample.dcb.gz"
	if err := os.WriteFile(compressedPath, compressed.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	loadOptions := &CodeDCBOptions{Obs1: "P1", Obs2: "C1", Year: 2026, Month: 6, TimeScale: GPST}
	for _, load := range []struct {
		name string
		fn   func(*testing.T) (int, error)
	}{
		{"strict", func(t *testing.T) (int, error) {
			value, err := LoadCodeDCB(compressedPath, loadOptions)
			if err != nil {
				return 0, err
			}
			closeAfterTest(t, value)
			return value.RecordCount()
		}},
		{"policy", func(t *testing.T) (int, error) {
			value, err := LoadCodeDCBWithPolicy(compressedPath, loadOptions, BiasReadPolicyStrict)
			if err != nil {
				return 0, err
			}
			closeAfterTest(t, value)
			return value.Value.RecordCount()
		}},
		{"lossy", func(t *testing.T) (int, error) {
			value, err := LoadCodeDCBLossy(compressedPath, loadOptions)
			if err != nil {
				return 0, err
			}
			closeAfterTest(t, value)
			return value.Value.RecordCount()
		}},
	} {
		t.Run(load.name, func(t *testing.T) {
			count, err := load.fn(t)
			if err != nil || count == 0 {
				t.Fatalf("gzip CODE DCB load: records=%d err=%v", count, err)
			}
		})
	}
}

func TestReadCodeDCBPathBoundsAndGzipMembers(t *testing.T) {
	t.Run("plain input bound", func(t *testing.T) {
		path := t.TempDir() + "/input.dcb"
		if err := os.WriteFile(path, []byte("12345"), 0o600); err != nil {
			t.Fatal(err)
		}
		_, err := readCodeDCBPathWithLimits(path, 10, 4)
		var limitErr *SizeLimitError
		if !errors.As(err, &limitErr) || limitErr.Kind != "CODE DCB product" || limitErr.Limit != 4 {
			t.Fatalf("plain size error = %#v, %v", limitErr, err)
		}
	})

	t.Run("compressed input bound", func(t *testing.T) {
		path := t.TempDir() + "/input.dcb.gz"
		if err := os.WriteFile(path, bytes.Repeat([]byte{'x'}, 11), 0o600); err != nil {
			t.Fatal(err)
		}
		_, err := readCodeDCBPathWithLimits(path, 10, 100)
		var limitErr *SizeLimitError
		if !errors.As(err, &limitErr) || limitErr.Kind != "compressed CODE DCB product" || limitErr.Limit != 10 {
			t.Fatalf("compressed size error = %#v, %v", limitErr, err)
		}
	})

	t.Run("expanded input bound", func(t *testing.T) {
		var archive bytes.Buffer
		writer := gzip.NewWriter(&archive)
		if _, err := writer.Write([]byte("more than eight")); err != nil {
			t.Fatal(err)
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		path := t.TempDir() + "/input.dcb.gz"
		if err := os.WriteFile(path, archive.Bytes(), 0o600); err != nil {
			t.Fatal(err)
		}
		_, err := readCodeDCBPathWithLimits(path, 100, 8)
		var limitErr *SizeLimitError
		if !errors.As(err, &limitErr) || limitErr.Kind != "decompressed CODE DCB product" || limitErr.Limit != 8 {
			t.Fatalf("decompressed size error = %#v, %v", limitErr, err)
		}
	})

	t.Run("concatenated members", func(t *testing.T) {
		var archive bytes.Buffer
		for _, part := range []string{"first", "second"} {
			writer := gzip.NewWriter(&archive)
			if _, err := io.WriteString(writer, part); err != nil {
				t.Fatal(err)
			}
			if err := writer.Close(); err != nil {
				t.Fatal(err)
			}
		}
		path := t.TempDir() + "/input.dcb.gz"
		if err := os.WriteFile(path, archive.Bytes(), 0o600); err != nil {
			t.Fatal(err)
		}
		got, err := readCodeDCBPathWithLimits(path, 100, 100)
		if err != nil || string(got) != "firstsecond" {
			t.Fatalf("concatenated gzip output = %q, %v", got, err)
		}
	})

	t.Run("truncated gzip", func(t *testing.T) {
		var archive bytes.Buffer
		writer := gzip.NewWriter(&archive)
		if _, err := writer.Write([]byte("payload")); err != nil {
			t.Fatal(err)
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		path := t.TempDir() + "/input.dcb.gz"
		truncated := archive.Bytes()[:archive.Len()-4]
		if err := os.WriteFile(path, truncated, 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := readCodeDCBPathWithLimits(path, 100, 100); err == nil {
			t.Fatal("truncated gzip member was accepted")
		}
	})
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
	// The current writer preserves both input COMMENT records and serializes
	// the six lower-triangular covariance rows required by the parsed matrix.
	// The former output dropped comments and emitted a keyword expansion; the
	// canonical bytes were independently diffed against the pinned fixture.
	if !bytes.Contains(oemOutput, []byte("COMMENT Annotated OEM fixture for a GPS navigation spacecraft.")) ||
		!bytes.Contains(oemOutput, []byte("COMMENT Epoch X Y Z X_DOT Y_DOT Z_DOT with one acceleration-bearing sample.")) {
		t.Fatalf("OEM output lost retained comments: %s", oemOutput)
	}
	covarianceRows := []byte("0.0001\n0 0.0002\n0 0 0.0003\n0 0 0 0.00000001\n0 0 0 0 0.00000002\n0 0 0 0 0 0.00000003\n")
	if !bytes.Contains(oemOutput, covarianceRows) {
		t.Fatalf("OEM covariance lower triangle was not fully serialized: %s", oemOutput)
	}
	assertSerializedHash(t, "OEM KVN", oemOutput, "e42826faae852e2c861cb777b50e4c646f27ed503781826d945104edb5023606")
	oemRoundTrip, err := ParseOEMKVN(oemOutput)
	if err != nil {
		t.Fatal(err)
	}
	roundTripBytes, err := oemRoundTrip.ToKVN()
	if err != nil || !bytes.Equal(roundTripBytes, oemOutput) {
		t.Fatalf("OEM retained comments/covariance changed across parse-serialize round trip: %v", err)
	}
	if err := oemRoundTrip.Close(); err != nil {
		t.Fatal(err)
	}
	oemOutput[0] ^= 0xff
	independentOEMOutput, err := oemKVN.ToKVN()
	if err != nil {
		t.Fatal(err)
	}
	assertSerializedHash(t, "independent OEM KVN", independentOEMOutput, "e42826faae852e2c861cb777b50e4c646f27ed503781826d945104edb5023606")
	if err := oemKVN.Close(); err != nil {
		t.Fatal(err)
	}
	if err := oemKVN.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := oemKVN.ToKVN(); !errors.Is(err, ErrClosed) {
		t.Fatalf("OEM after close: %v", err)
	}
	oemKVNXML, err := ParseOEMKVN(fixtureHash(t, "oem/gps.kvn", "94352528a735af9d941335086ab36f776192b635ed590aadf39963bf8c9784aa"))
	if err != nil {
		t.Fatal(err)
	}
	oemXMLBytes, err := oemKVNXML.ToXML()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(oemXMLBytes, []byte("<COMMENT>Annotated OEM fixture for a GPS navigation spacecraft.</COMMENT>")) ||
		!bytes.Contains(oemXMLBytes, []byte("<COMMENT>Epoch X Y Z X_DOT Y_DOT Z_DOT with one acceleration-bearing sample.</COMMENT>")) ||
		!bytes.Contains(oemXMLBytes, []byte("<CZ_DOT_Z_DOT>0.00000003</CZ_DOT_Z_DOT>")) {
		t.Fatalf("OEM XML lost comments or covariance fields: %s", oemXMLBytes)
	}
	assertSerializedHash(t, "OEM KVN source to XML", oemXMLBytes, "8efe3df6901143766ce001c998b5263256e88f2c54e3952a0582bd1598cd7f40")
	if err := oemKVNXML.Close(); err != nil {
		t.Fatal(err)
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
		if !bytes.Contains(output, []byte("<COMMENT>Annotated OPM fixture for a low Earth orbit servicing spacecraft.</COMMENT>")) ||
			!bytes.Contains(output, []byte("<COMMENT>Two planned trim burns.</COMMENT>")) {
			t.Fatalf("OPM XML lost retained comments: %s", output)
		}
		assertSerializedHash(t, "OPM KVN source to XML", output, "4bafa8ee0c4f3d51cba94a3a311be5d1fd694afdf890c976cdef479ab419839c")
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
	if !bytes.Contains(opmOutput, []byte("COMMENT Annotated OPM fixture for a low Earth orbit servicing spacecraft.")) ||
		!bytes.Contains(opmOutput, []byte("COMMENT Two planned trim burns.")) {
		t.Fatalf("OPM KVN lost retained comments: %s", opmOutput)
	}
	assertSerializedHash(t, "OPM KVN", opmOutput, "1ee924750ce4d0e9151c6197c660c707eaf33980265e569602f367e73458ac6d")
	opmRoundTrip, err := ParseOPMKVN(opmOutput)
	if err != nil {
		t.Fatal(err)
	}
	opmRoundTripOutput, err := opmRoundTrip.ToKVN()
	if err != nil || !bytes.Equal(opmRoundTripOutput, opmOutput) {
		t.Fatalf("OPM comments changed across KVN round trip: %v", err)
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
		if name == "JSON" {
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(output, &fields); err != nil {
				t.Fatalf("OMM JSON decode: %v", err)
			}
			for key, want := range map[string]string{
				"CCSDS_OMM_VERS": `"2.0"`, "CREATION_DATE": "null", "ORIGINATOR": "null",
				"OBJECT_NAME": `"NAVSTAR 43 (USA 132)"`, "CENTER_NAME": `"EARTH"`,
				"REF_FRAME": `"TEME"`, "TIME_SYSTEM": `"UTC"`,
				"MEAN_ELEMENT_THEORY": `"SGP/SGP4"`,
			} {
				if string(fields[key]) != want {
					t.Fatalf("OMM JSON field %s = %s, want %s", key, fields[key], want)
				}
			}
		}
		// The prior JSON pin is the same field/value tokens sorted by key. The
		// merged core enables serde_json preserve_order, so this hash pins the
		// current insertion order without changing numeric or string values.
		wantHash := map[string]string{"KVN": "e20c3848984d9df2556d78f43040092a352699cce5b9397da71a0615d271e51c", "XML": "3f6ab03c6ce8aa09396399d7081b88388014d1b6e46940c5293fb57dc1b745fc", "JSON": "0158cebc281e0a0c614a057223441798f55017b95390288692360fcd003d274b"}
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
		if name == "JSON" {
			roundTripOutput, encodeErr := roundTrip.ToJSON()
			if encodeErr != nil || !bytes.Equal(roundTripOutput, output) {
				t.Fatalf("OMM JSON metadata changed across round trip: %v", encodeErr)
			}
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
		// Strict TDM KVN terminates its last line; the old canonical bytes were
		// identical except for that final LF.
		if !bytes.HasSuffix(output, []byte("\n")) || bytes.HasSuffix(output, []byte("\n\n")) {
			tail := output
			if len(tail) > 8 {
				tail = tail[len(tail)-8:]
			}
			t.Fatalf("TDM strict writer must emit exactly one final LF: tail=%q", tail)
		}
		if !bytes.Contains(output, []byte("CCSDS_TDM_VERS = 2.0\n")) ||
			!bytes.Contains(output, []byte("TRANSMIT_PHASE_CT_1 = 2005-184T11:12:23 7175173383.615373\n")) {
			t.Fatalf("TDM writer did not emit canonical spaced assignments: %s", output)
		}
		assertSerializedHash(t, "TDM KVN", output, "25b438d10334c787a9d88f24b7682dfa43fd850a1f4d2fc33255229ebaa11a3f")
		roundTrip, parseErr := ParseTDMKVN(output)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		if count, countErr := roundTrip.RecordCount(); countErr != nil || count != 20 {
			t.Fatalf("TDM round-trip count = %d, %v", count, countErr)
		}
		roundTripBytes, encodeErr := roundTrip.ToKVN()
		if encodeErr != nil || !bytes.Equal(roundTripBytes, output) {
			t.Fatalf("TDM canonical KVN changed across round trip: %v", encodeErr)
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
	inSameFrame, err := spk.StateInFrame(20000433, 10, 757339200, state.Frame)
	if err != nil || inSameFrame.Frame != state.Frame || inSameFrame.PositionKm != state.PositionKm || inSameFrame.VelocityKmPerS != state.VelocityKmPerS {
		t.Fatalf("SPK requested-frame identity = %#v, %v; default=%#v", inSameFrame, err, state)
	}
	inECLIPJ2000, err := spk.StateInFrame(20000433, 10, 757339200, 17)
	if err != nil || inECLIPJ2000.Frame != 17 || inECLIPJ2000.PositionKm == state.PositionKm {
		t.Fatalf("SPK ECLIPJ2000 transform = %#v, %v; default=%#v", inECLIPJ2000, err, state)
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
	if !skipped[0].NORADIDPresent {
		t.Fatalf("stated catalog id lost its presence bit: %#v", skipped[0])
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
	malformed, err := badCatalog.MalformedRecords()
	if err != nil || len(malformed) != 1 {
		t.Fatalf("bad catalog malformed records = %#v, %v", malformed, err)
	}
	if malformed[0].Index != 0 || malformed[0].Error.Kind != "field" || malformed[0].Error.Fields.Message == nil || *malformed[0].Error.Fields.Message != "expected a JSON object" || len(malformed[0].Raw) == 0 || len(malformed[0].Error.Raw) == 0 {
		t.Fatalf("typed parser-level rejection = %#v", malformed[0])
	}
	if _, err := badCatalog.MalformedRecord(-1); err == nil {
		t.Fatal("negative malformed-record index accepted")
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
	if _, err := badCatalog.MalformedRecord(0); !errors.Is(err, ErrClosed) {
		t.Fatalf("malformed record after close: %v", err)
	}
	for _, tc := range []struct {
		name    string
		feed    string
		present bool
	}{
		{name: "absent identifier", feed: strings.Replace(ommCatalogFeed, `,"NORAD_CAT_ID":42738`, "", 1), present: false},
		{name: "stated zero identifier", feed: strings.Replace(ommCatalogFeed, `"NORAD_CAT_ID":42738`, `"NORAD_CAT_ID":0`, 1), present: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			value, err := BuildOMMCatalogLenient(GNSSSystemGPS, []byte(tc.feed))
			if err != nil {
				t.Fatal(err)
			}
			closeAfterTest(t, value)
			entry, err := value.Skipped(0)
			if err != nil {
				t.Fatal(err)
			}
			if entry.NORADIDPresent != tc.present || entry.NORADID != 0 {
				t.Fatalf("identity presence/value = %v/%d, want %v/0", entry.NORADIDPresent, entry.NORADID, tc.present)
			}
		})
	}
	if _, err := BuildOMMCatalogLenient(GNSSSystem(99), []byte(ommCatalogFeed)); err == nil {
		t.Fatal("invalid catalog system accepted")
	}
}

func TestOMMParseErrorExpectedFieldRetainsVariantTypeAndNestedCause(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		kind    string
	}{
		{name: "unit mismatch text", payload: `{"kind":"unit_mismatch","fields":{"field":"MEAN_MOTION","unit":"rev/s","expected":"rev/day"}}`, kind: "unit_mismatch"},
		{name: "unit mismatch null", payload: `{"kind":"unit_mismatch","fields":{"field":"X","unit":"m","expected":null}}`, kind: "unit_mismatch"},
		{name: "csv column count", payload: `{"kind":"csv_column_count","fields":{"found":3,"expected":24}}`, kind: "csv_column_count"},
		{name: "nested source", payload: `{"kind":"in_record","fields":{"index":7,"source":{"kind":"csv_column_count","fields":{"found":3,"expected":24}}}}`, kind: "in_record"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var value OMMParseError
			if err := json.Unmarshal([]byte(tc.payload), &value); err != nil {
				t.Fatal(err)
			}
			if value.Kind != tc.kind {
				t.Fatalf("kind=%q", value.Kind)
			}
			target := &value
			if tc.kind == "in_record" {
				if value.Fields.Index == nil || value.Fields.Index.String() != "7" || value.Fields.Source == nil || len(value.Raw) == 0 || len(value.Fields.Source.Raw) == 0 {
					t.Fatalf("nested record error=%+v", value.Fields)
				}
				target = value.Fields.Source
			}
			if tc.name == "unit mismatch text" {
				if !target.Fields.Expected.Present || target.Fields.Expected.Text == nil || *target.Fields.Expected.Text != "rev/day" || target.Fields.Expected.Count != nil || target.Fields.Expected.IsNull {
					t.Fatalf("typed text expectation=%+v", target.Fields.Expected)
				}
			} else if tc.kind == "unit_mismatch" {
				if !target.Fields.Expected.Present || !target.Fields.Expected.IsNull || target.Fields.Expected.Text != nil || target.Fields.Expected.Count != nil {
					t.Fatalf("typed null expectation=%+v", target.Fields.Expected)
				}
			} else if !target.Fields.Expected.Present || target.Fields.Expected.Count == nil || target.Fields.Expected.Count.String() != "24" || target.Fields.Expected.Text != nil || target.Fields.Found == nil || target.Fields.Found.String() != "3" {
				t.Fatalf("typed CSV count fields=%+v", target.Fields)
			}
		})
	}
}

func TestOEMRetainsTypedSkippedStateAlongsideValidSegment(t *testing.T) {
	fixture := readSpaceFixture(t, "oem/gps.kvn")
	const validLine = "2026-06-28T00:15:00.000 17450"
	if !bytes.Contains(fixture, []byte(validLine)) {
		t.Fatal("OEM fixture no longer contains the expected valid state line")
	}
	modified := bytes.Replace(fixture, []byte(validLine), []byte("2026-06-28T00:07:30.000 1 2\n"+validLine), 1)
	oem, err := ParseOEMKVN(modified)
	if err != nil {
		t.Fatal(err)
	}
	if count, err := oem.SegmentCount(); err != nil || count != 1 {
		t.Fatalf("retained OEM segment count=%d err=%v", count, err)
	}
	if count, err := oem.SkippedStateCount(); err != nil || count != 1 {
		t.Fatalf("OEM skipped-state count=%d err=%v", count, err)
	}
	rows, err := oem.SkippedStates()
	if err != nil || len(rows) != 1 {
		t.Fatalf("OEM skipped states=%+v err=%v", rows, err)
	}
	row := rows[0]
	if row.Line == 0 || row.Segment != 0 || row.Text != "2026-06-28T00:07:30.000 1 2" || row.Reason.Kind != "item_count" || row.Reason.Fields.Found == nil || *row.Reason.Fields.Found != "3" || len(row.Raw) == 0 || len(row.Reason.Raw) == 0 {
		t.Fatalf("typed OEM skipped state=%+v", row)
	}
	if _, err := oem.SkippedState(-1); err == nil {
		t.Fatal("negative OEM skipped-state index accepted")
	}
	if err := oem.Close(); err != nil {
		t.Fatal(err)
	}
	if err := oem.Close(); err != nil {
		t.Fatal(err)
	}
	if rows[0].Reason.Fields.Found == nil || *rows[0].Reason.Fields.Found != "3" || row.Text == "" {
		t.Fatalf("detached OEM state changed after close: %+v", rows)
	}
	if _, err := oem.SkippedStateCount(); !errors.Is(err, ErrClosed) {
		t.Fatalf("OEM after close: %v", err)
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
