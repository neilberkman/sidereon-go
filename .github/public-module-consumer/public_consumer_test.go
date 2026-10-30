package consumer_test

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"os"
	"reflect"
	"strings"
	"testing"

	"sidereon.dev/go/v3"
)

func TestPublishedLibraryVersion(t *testing.T) {
	want := strings.TrimPrefix(os.Getenv("SIDEREON_EXPECTED_VERSION"), "v")
	got := sidereon.LibraryVersion()
	if got.String != want || fmt.Sprintf("%d.%d.%d", got.Major, got.Minor, got.Patch) != want {
		t.Fatalf("LibraryVersion = %+v, want %s", got, want)
	}
}

func TestPublishedSPPIndependentReference(t *testing.T) {
	sp3 := loadSP3(t)
	observations := []sidereon.SPPObservation{
		{SatelliteID: "G08", PseudorangeM: math.Float64frombits(0x4176b8c6fd82e861)},
		{SatelliteID: "G10", PseudorangeM: math.Float64frombits(0x4175aa4fa1a0c21f)},
		{SatelliteID: "G16", PseudorangeM: math.Float64frombits(0x417387abd6052c3b)},
		{SatelliteID: "G18", PseudorangeM: math.Float64frombits(0x4174c288f3bd1166)},
		{SatelliteID: "G20", PseudorangeM: math.Float64frombits(0x417443947bd00bd6)},
		{SatelliteID: "G21", PseudorangeM: math.Float64frombits(0x4173d8405cd09f84)},
		{SatelliteID: "G26", PseudorangeM: math.Float64frombits(0x417425d51967e798)},
		{SatelliteID: "G27", PseudorangeM: math.Float64frombits(0x41745a4b78a81707)},
	}
	solution, err := sidereon.SolveSPP(sp3, sidereon.SPPConfig{
		Observations:    observations,
		TRxJ2000S:       646272000,
		TRxSecondOfDayS: 43200,
		DayOfYear:       176.5,
		InitialGuess:    [4]float64{4.5e6, 0.5e6, 4.5e6, 0},
		WithGeodetic:    true,
	})
	if err != nil {
		t.Fatal(err)
	}
	wantPosition := [3]float64{4484137.56180868, 550578.016017378, 4487569.615357698}
	for axis, want := range wantPosition {
		if delta := math.Abs(solution.PositionM[axis] - want); math.IsNaN(delta) || math.IsInf(delta, 0) || delta > 5e-5 {
			t.Fatalf("position[%d] = %.17g, independent reference %.17g", axis, solution.PositionM[axis], want)
		}
	}
	clockDelta := math.Abs(solution.ReceiverClockS - 0.00010009082050400748)
	if math.IsNaN(clockDelta) || math.IsInf(clockDelta, 0) || clockDelta > 1e-12 {
		t.Fatalf("receiver clock = %.17g", solution.ReceiverClockS)
	}
	wantResiduals := []float64{-0.7862987704575062, -2.770254924893379, -0.6641135476529598, -0.17380670458078384, 3.825963206589222, -4.198393113911152, 2.1111478097736835, 2.6201590932905674}
	if len(solution.ResidualsM) != len(wantResiduals) {
		t.Fatalf("residual count = %d, want %d", len(solution.ResidualsM), len(wantResiduals))
	}
	for index, want := range wantResiduals {
		if delta := math.Abs(solution.ResidualsM[index] - want); math.IsNaN(delta) || math.IsInf(delta, 0) || delta > 1e-4 {
			t.Fatalf("residual[%d] = %.17g, independent reference %.17g", index, solution.ResidualsM[index], want)
		}
	}
	if !solution.Metadata.Converged || solution.DOP == nil || solution.Geodetic == nil {
		t.Fatalf("incomplete SPP solution: %+v", solution)
	}
}

func TestPublishedMergedGridSP3WritesAndRoundTrips(t *testing.T) {
	data, err := os.ReadFile("testdata/trimmed.sp3")
	if err != nil {
		t.Fatal(err)
	}
	// Change one valid six-decimal clock column by one column unit, so the
	// two-source mean falls halfway between representable clock columns.
	secondData := bytes.Replace(data,
		[]byte("PG08   6042.284075 -25684.525952  -1323.814227    -38.640412"),
		[]byte("PG08   6042.284075 -25684.525952  -1323.814227    -38.640413"), 1)
	if bytes.Equal(secondData, data) {
		t.Fatal("fixture did not contain the expected first G08 record")
	}
	first, err := sidereon.LoadSP3(data)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = first.Close() })
	second, err := sidereon.LoadSP3(secondData)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = second.Close() })
	for label, source := range map[string]*sidereon.SP3{"first": first, "second": second} {
		if _, outcome, err := source.Encode(); err != nil || !outcome.IsOK {
			t.Fatalf("%s writable SP3 input = %+v, %v", label, outcome, err)
		}
	}

	options, err := sidereon.NewSP3MergeOptions()
	if err != nil {
		t.Fatal(err)
	}
	options.MinAgree = 1
	options.ClockMinCommon = 1
	merged, report, err := sidereon.MergeSP3([]*sidereon.SP3{first, second}, &options)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = merged.Close() })
	t.Cleanup(func() { _ = report.Close() })
	encoded, outcome, err := merged.Encode()
	if err != nil || !outcome.IsOK || len(encoded) == 0 {
		t.Fatalf("strict write of merged grid product = %+v, %v", outcome, err)
	}
	roundedText, rounding, err := merged.ToSP3StringWithRounding()
	if err != nil {
		t.Fatal(err)
	}
	if rounding != (sidereon.SP3RoundingReport{}) || roundedText != string(encoded) {
		t.Fatalf("explicit writer changed already-quantized merge: report=%+v", rounding)
	}
	reparsed, err := sidereon.LoadSP3(encoded)
	if err != nil {
		t.Fatalf("reparse merged output: %v", err)
	}
	t.Cleanup(func() { _ = reparsed.Close() })
	mergedState, err := merged.State("G08", 0)
	if err != nil {
		t.Fatal(err)
	}
	reparsedState, err := reparsed.State("G08", 0)
	if err != nil {
		t.Fatal(err)
	}
	if math.Float64bits(mergedState.ClockS) != math.Float64bits(reparsedState.ClockS) {
		t.Fatalf("merged/read-back clock differs: merged %.17g, read-back %.17g", mergedState.ClockS, reparsedState.ClockS)
	}
	if _, outcome, err := reparsed.Encode(); err != nil || !outcome.IsOK {
		t.Fatalf("strict write after merge/read-back = %+v, %v", outcome, err)
	}
}

func TestPublishedSP3ConservativeAccuracyRounding(t *testing.T) {
	load := func(name string) *sidereon.SP3 {
		t.Helper()
		data, err := os.ReadFile("testdata/sp3_accuracy_306/" + name + ".sp3")
		if err != nil {
			t.Fatal(err)
		}
		product, err := sidereon.LoadSP3(data)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = product.Close() })
		if _, outcome, err := product.Encode(); err != nil || !outcome.IsOK {
			t.Fatalf("source %s strict write = %+v, %v", name, outcome, err)
		}
		return product
	}
	a, b := load("A-velocity"), load("B-velocity")
	directText, directReport, err := b.ToSP3StringWithAccuracyRounding(sidereon.SP3AccuracyRoundingConservative)
	if err != nil || directText == "" || directReport.AccuracySigmas != (sidereon.SP3AccuracyRoundingReport{}) {
		t.Fatalf("direct all-field accuracy report = %+v, err %v", directReport.AccuracySigmas, err)
	}
	directReparsed, err := sidereon.LoadSP3([]byte(directText))
	if err != nil {
		t.Fatalf("parse direct all-field output: %v", err)
	}
	t.Cleanup(func() { _ = directReparsed.Close() })
	for epoch := 0; epoch < 2; epoch++ {
		before, err := b.RecordAccuracy("G12", epoch)
		if err != nil {
			t.Fatal(err)
		}
		after, err := directReparsed.RecordAccuracy("G12", epoch)
		if err != nil || !reflect.DeepEqual(before, after) {
			t.Fatalf("direct writer changed all-field accuracy epoch %d: before %+v after %+v err %v", epoch, before, after, err)
		}
	}
	options, err := sidereon.NewSP3MergeOptions()
	if err != nil {
		t.Fatal(err)
	}
	options.MinAgree = 1
	options.ClockMinCommon = 1
	merged, report, err := sidereon.MergeSP3([]*sidereon.SP3{a, b}, &options)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = merged.Close() })
	t.Cleanup(func() { _ = report.Close() })
	_, strict, err := merged.Encode()
	if err != nil || strict.IsOK || strict.Error.Kind != sidereon.SP3WriteErrorAccuracyNotRepresentable {
		t.Fatalf("strict merged write = %+v, %v", strict, err)
	}
	text, detailed, err := merged.ToSP3StringWithAccuracyRounding(sidereon.SP3AccuracyRoundingConservative)
	if err != nil {
		t.Fatalf("conservative accuracy write: %v", err)
	}
	wantSigma := 0.032
	wantSource := math.Pow(1.25, 13) * 1e-3
	wantClock := 2*299792458e-12 - math.Pow(1.025, 13)*299792458e-12
	gotReport := detailed.AccuracySigmas
	wantReport := sidereon.SP3AccuracyRoundingReport{
		PositionSigmaXM: wantSigma - wantSource, PositionSigmaYM: wantSigma - wantSource, PositionSigmaZM: wantSigma - wantSource,
		ClockSigmaM: wantClock,
	}
	gotFields := []float64{gotReport.PositionSigmaXM, gotReport.PositionSigmaYM, gotReport.PositionSigmaZM, gotReport.VelocitySigmaXMS, gotReport.VelocitySigmaYMS, gotReport.VelocitySigmaZMS, gotReport.ClockSigmaM, gotReport.ClockRateSigmaMS}
	wantFields := []float64{wantReport.PositionSigmaXM, wantReport.PositionSigmaYM, wantReport.PositionSigmaZM, wantReport.VelocitySigmaXMS, wantReport.VelocitySigmaYMS, wantReport.VelocitySigmaZMS, wantReport.ClockSigmaM, wantReport.ClockRateSigmaMS}
	for index, want := range wantFields {
		if math.Abs(gotFields[index]-want) > 1e-15 {
			t.Fatalf("accuracy report = %+v, want x delta %.17g", detailed.AccuracySigmas, wantSigma-wantSource)
		}
	}
	reparsed, err := sidereon.LoadSP3([]byte(text))
	if err != nil {
		t.Fatalf("parse conservative output: %v", err)
	}
	t.Cleanup(func() { _ = reparsed.Close() })
	for epoch := 0; epoch < 2; epoch++ {
		got, err := reparsed.RecordAccuracy("G12", epoch)
		if err != nil || !got.HasP || got.HasV || got.P.PositionSigmaM[0].Kind != sidereon.SP3AccuracyKnown || got.P.PositionSigmaM[0].Value != wantSigma {
			t.Fatalf("G12 accuracy epoch %d = %+v, err %v", epoch, got, err)
		}
		for axis := 0; axis < 3; axis++ {
			if math.Abs(got.P.PositionSigmaM[axis].Value-wantSigma) > 1e-15 {
				t.Fatalf("G12 rounded axis %d epoch %d = %.17g", axis, epoch, got.P.PositionSigmaM[axis].Value)
			}
		}
		if math.Abs(got.P.ClockSigmaM.Value-2*299792458e-12) > 1e-15 {
			t.Fatalf("G12 rounded clock sigma epoch %d = %.17g", epoch, got.P.ClockSigmaM.Value)
		}
		anchor, err := reparsed.State("G02", epoch)
		wantAnchorClock := -20e-6
		if epoch == 1 {
			wantAnchorClock = -19e-6
		}
		if err != nil || !anchor.HasClock || math.Abs(anchor.ClockS-wantAnchorClock) > 1e-18 {
			t.Fatalf("G02 clock anchor epoch %d = %+v, err %v", epoch, anchor, err)
		}
		clock, err := reparsed.RecordAccuracy("G02", epoch)
		if err != nil || !clock.HasP || clock.HasV || clock.P.ClockSigmaM.Kind != sidereon.SP3AccuracyUnknown {
			t.Fatalf("G02 rounded clock sigma epoch %d = %+v, err %v", epoch, clock.P.ClockSigmaM, err)
		}
	}
	if _, outcome, err := reparsed.Encode(); err != nil || !outcome.IsOK {
		t.Fatalf("strict write of conservative output = %+v, %v", outcome, err)
	}

	reverseOptions, err := sidereon.NewSP3MergeOptions()
	if err != nil {
		t.Fatal(err)
	}
	reverseOptions.MinAgree = 1
	reverseOptions.ClockMinCommon = 1
	reverse, reverseReport, err := sidereon.MergeSP3([]*sidereon.SP3{b, a}, &reverseOptions)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = reverse.Close() })
	t.Cleanup(func() { _ = reverseReport.Close() })
	_, reverseStrict, err := reverse.Encode()
	if err != nil || reverseStrict.IsOK || reverseStrict.Error.Kind != sidereon.SP3WriteErrorAccuracyNotRepresentable {
		t.Fatalf("reverse strict merge write = %+v, %v", reverseStrict, err)
	}
	reverseText, reverseDetailed, err := reverse.ToSP3StringWithAccuracyRounding(sidereon.SP3AccuracyRoundingConservative)
	if err != nil || reverseText == "" {
		t.Fatalf("reverse conservative merge write = %+v, %v", reverseDetailed, err)
	}
	reverseReparsed, err := sidereon.LoadSP3([]byte(reverseText))
	if err != nil {
		t.Fatalf("parse reverse conservative output: %v", err)
	}
	t.Cleanup(func() { _ = reverseReparsed.Close() })
	before, err := reverse.RecordAccuracy("G01", 0)
	if err != nil {
		t.Fatal(err)
	}
	after, err := reverseReparsed.RecordAccuracy("G01", 0)
	if err != nil {
		t.Fatal(err)
	}
	wantPositionDelta := after.P.PositionSigmaM[0].Value - before.P.PositionSigmaM[0].Value
	wantClockDelta := after.P.ClockSigmaM.Value - before.P.ClockSigmaM.Value
	if wantPositionDelta <= 0 || wantClockDelta <= 0 ||
		math.Abs(reverseDetailed.AccuracySigmas.PositionSigmaXM-wantPositionDelta) > 1e-15 ||
		math.Abs(reverseDetailed.AccuracySigmas.ClockSigmaM-wantClockDelta) > 1e-15 {
		t.Fatalf("reverse report %+v differs from reparsed G01 changes pos %.17g clock %.17g, err %v", reverseDetailed.AccuracySigmas, wantPositionDelta, wantClockDelta, err)
	}
}

func TestPublishedNMEAEpochInstantValidation(t *testing.T) {
	readEpoch := func(body string) sidereon.NMEAEpoch {
		t.Helper()
		log, err := sidereon.ParseNMEA([]byte(nmeaSentence(body) + "\n"))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = log.Close() })
		epochs, err := log.Epochs()
		if err != nil || len(epochs) != 1 {
			t.Fatalf("epochs = %+v, err = %v", epochs, err)
		}
		return epochs[0]
	}

	ordinary := readEpoch("GPRMC,120000.123456789,A,4807.038,N,01131.000,E,0.0,0.0,010100,,,A")
	if !ordinary.HasInstantJ2000S || ordinary.InstantJ2000S != 0.12345678900000001 {
		t.Fatalf("ordinary instant = present %v, %.17g", ordinary.HasInstantJ2000S, ordinary.InstantJ2000S)
	}
	wholeLeap := readEpoch("GPRMC,235960,A,4807.038,N,01131.000,E,0.0,0.0,010100,,,A")
	if !wholeLeap.HasInstantJ2000S || wholeLeap.InstantJ2000S != 43200 {
		t.Fatalf("whole leap instant = present %v, %.17g", wholeLeap.HasInstantJ2000S, wholeLeap.InstantJ2000S)
	}
	fractionalLeap := readEpoch("GPRMC,235960.123456789,A,4807.038,N,01131.000,E,0.0,0.0,010100,,,A")
	if !fractionalLeap.HasCalendarEpoch || fractionalLeap.CalendarEpoch.Second != 60.123456789 || fractionalLeap.HasInstantJ2000S || fractionalLeap.InstantJ2000S != 0 {
		t.Fatalf("fractional leap epoch = %+v", fractionalLeap)
	}
	missingDate := readEpoch("GPGGA,120000,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,")
	if missingDate.HasInstantJ2000S || missingDate.InstantJ2000S != 0 {
		t.Fatalf("missing-date instant = present %v, %.17g", missingDate.HasInstantJ2000S, missingDate.InstantJ2000S)
	}
	missingTime := readEpoch("GPRMC,,A,4807.038,N,01131.000,E,0.0,0.0,010100,,,A")
	if missingTime.HasInstantJ2000S || missingTime.InstantJ2000S != 0 {
		t.Fatalf("missing-time instant = present %v, %.17g", missingTime.HasInstantJ2000S, missingTime.InstantJ2000S)
	}
}

func TestPublishedPPPUT1RefusalAndPermissiveResult(t *testing.T) {
	sp3 := loadSP3(t)
	epochs := []sidereon.PPPCorrectionEpoch{{
		Epoch:     sidereon.CivilDateTime{Year: 1900, Month: 1, Day: 1},
		TRxJ2000S: -3_155_716_800,
	}}
	receiver := [3]float64{4.5e6, 0.5e6, 4.5e6}
	options := sidereon.PPPCorrectionsOptions{SolidEarthTide: true}
	strictBuilders := []func() (*sidereon.PPPCorrections, error){
		func() (*sidereon.PPPCorrections, error) {
			return sidereon.BuildPPPCorrections(sp3, epochs, receiver, options)
		},
		func() (*sidereon.PPPCorrections, error) {
			return sidereon.BuildPPPCorrectionsWithValidityAndTideConstants(sp3, epochs, receiver, options, sidereon.PPPValidityStrict, sidereon.StationTideConventions)
		},
	}
	for index, build := range strictBuilders {
		corrections, err := build()
		if err == nil || corrections != nil {
			t.Fatalf("strict builder %d = %v, %v; want nil handle and refusal", index, corrections, err)
		}
		var buildErr *sidereon.PPPCorrectionsBuildError
		if !errors.As(err, &buildErr) || buildErr.Kind != 2 {
			t.Fatalf("strict builder %d correction error = %T %+v", index, err, err)
		}
		var statusErr *sidereon.StatusError
		if !errors.As(err, &statusErr) || statusErr.Code != sidereon.StatusCode(8) {
			t.Fatalf("strict builder %d status = %T %+v, want status 8", index, err, err)
		}
		if !strings.Contains(statusErr.Detail, "precedes") || !strings.Contains(statusErr.Detail, "UT1") || !strings.Contains(statusErr.Detail, "coverage") {
			t.Fatalf("strict builder %d detail = %q", index, statusErr.Detail)
		}
	}

	permissive, err := sidereon.BuildPPPCorrectionsWithValidityAndTideConstants(sp3, epochs, receiver, options, sidereon.PPPValidityPermissive, sidereon.StationTideConventions)
	if err != nil || permissive == nil {
		t.Fatalf("permissive result = %v, %v", permissive, err)
	}
	t.Cleanup(func() { _ = permissive.Close() })
	degraded, err := permissive.DegradedReason()
	if err != nil || degraded != sidereon.StationTideBeforeCoverage {
		t.Fatalf("permissive degradation = %v, %v", degraded, err)
	}
	tide, err := permissive.Tide()
	if err != nil || len(tide) != 1 {
		t.Fatalf("permissive tide = %+v, %v", tide, err)
	}
	for axis, value := range tide[0].ValueM {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			t.Fatalf("permissive tide axis %d = %v", axis, value)
		}
	}
}

func loadSP3(t *testing.T) *sidereon.SP3 {
	t.Helper()
	data, err := os.ReadFile("testdata/trimmed.sp3")
	if err != nil {
		t.Fatal(err)
	}
	sp3, err := sidereon.LoadSP3(data)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sp3.Close() })
	return sp3
}

func nmeaSentence(body string) string {
	var checksum byte
	for _, value := range []byte(body) {
		checksum ^= value
	}
	return fmt.Sprintf("$%s*%02X", body, checksum)
}
