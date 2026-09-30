package sidereon

import (
	"math"
	"reflect"
	"strings"
	"testing"
)

func TestIONEXParseWarningsAndSkippedRecordSnapshot(t *testing.T) {
	product := newIONEXExactTestProduct(t)
	closeAfterTest(t, product)
	text, err := product.ToIONEXText()
	if err != nil {
		t.Fatal(err)
	}
	parsed, warnings, err := ParseIONEXWithWarnings(text)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, parsed)
	if warnings == nil {
		t.Fatal("warning snapshot is nil; expected a detached empty slice")
	}
	var withoutVersion strings.Builder
	for _, line := range strings.Split(string(text), "\n") {
		if !strings.Contains(line, "IONEX VERSION / TYPE") {
			withoutVersion.WriteString(line)
			withoutVersion.WriteByte('\n')
		}
	}
	warned, findings, err := ParseIONEXWithWarnings([]byte(withoutVersion.String()))
	if err != nil {
		t.Fatalf("parse with omitted mandatory record: %v", err)
	}
	closeAfterTest(t, warned)
	if err := warned.Close(); err != nil {
		t.Fatal(err)
	}
	if len(findings) == 0 || findings[0].Label == "" || findings[0].Message == "" {
		t.Fatalf("missing-record warning did not retain text: %#v", findings)
	}
	count, err := parsed.SkippedRecords()
	if err != nil {
		t.Fatal(err)
	}
	if count < 0 {
		t.Fatalf("negative skipped record count %d", count)
	}
	if _, _, err := ParseIONEXWithWarnings([]byte("not an IONEX product")); err == nil {
		t.Fatal("invalid IONEX text unexpectedly parsed with warnings")
	}
	if _, err := parsed.SkippedRecords(); err != nil {
		t.Fatalf("read skipped-record count after failed parse: %v", err)
	}
}

func TestIONEXKnownAbsentTECSampleSurvivesNativeRoundtrip(t *testing.T) {
	product, err := NewIONEXFromTECSamples([]TECSample{{
		TimeScale: UTC, EpochJ2000S: 0, LatDeg: 40, LonDeg: -20,
		VTECPresenceKnown: true, VTECPresent: false,
		EpochJ2000WholeS: 9007199254740993, EpochJ2000WholeSPresent: true,
		HeightOffsetKm: 12.5, HeightOffsetPresent: true,
	}, {
		TimeScale: UTC, EpochJ2000S: 0, LatDeg: 40, LonDeg: 20,
		VTECPresenceKnown: true, VTECPresent: true, VTECTECU: 2,
		EpochJ2000WholeS: 9007199254740993, EpochJ2000WholeSPresent: true,
		HeightOffsetKm: 12.5, HeightOffsetPresent: true,
	}, {
		TimeScale: UTC, EpochJ2000S: 0, LatDeg: -40, LonDeg: -20,
		VTECPresenceKnown: true, VTECPresent: true, VTECTECU: 3,
		EpochJ2000WholeS: 9007199254740993, EpochJ2000WholeSPresent: true,
		HeightOffsetKm: 12.5, HeightOffsetPresent: true,
	}, {
		TimeScale: UTC, EpochJ2000S: 0, LatDeg: -40, LonDeg: 20,
		VTECPresenceKnown: true, VTECPresent: true, VTECTECU: 4,
		EpochJ2000WholeS: 9007199254740993, EpochJ2000WholeSPresent: true,
		HeightOffsetKm: 12.5, HeightOffsetPresent: true,
	}}, 450, 6371, -1)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, product)
	samples, err := product.TECSamples()
	if err != nil {
		t.Fatal(err)
	}
	if len(samples) != 4 || !samples[0].VTECPresenceKnown || samples[0].VTECPresent || !samples[0].EpochJ2000WholeSPresent || samples[0].EpochJ2000WholeS != 9007199254740993 || !samples[0].HeightOffsetPresent || samples[0].HeightOffsetKm != 12.5 {
		t.Fatalf("known-absent VTEC sample changed: %#v", samples)
	}
}

func TestIONEXTECSampleConstructorWithHeader(t *testing.T) {
	header := IONEXHeaderMetadata{
		Version: 1, Date: "2026-09-29 00:00 UTC", Program: "sidereon", RunBy: "tests", SatelliteSystem: "GPS",
		IntervalS: 3600, MappingDeclaration: IONEXMappingDeclarationDeclared,
		MappingFunction: IONEXMappingFunctionOther, MappingFunctionCode: "XMAP",
		Descriptions: []string{"sample constructor"}, Comments: []string{"detached header"},
	}
	samples := []TECSample{
		{TimeScale: UTC, EpochJ2000S: 3600, EpochJ2000WholeS: 3600, EpochJ2000WholeSPresent: true, LatDeg: 40, LonDeg: -20, VTECPresenceKnown: true, VTECPresent: false, RMSPresent: true, RMSTECU: 0, HeightOffsetPresent: true, HeightOffsetKm: 12.5},
		{TimeScale: UTC, EpochJ2000S: 3600, EpochJ2000WholeS: 3600, EpochJ2000WholeSPresent: true, LatDeg: 40, LonDeg: 20, VTECPresent: true, VTECTECU: 2, RMSPresent: true, RMSTECU: 0.25, HeightOffsetPresent: true, HeightOffsetKm: 13.5},
		{TimeScale: UTC, EpochJ2000S: 3600, EpochJ2000WholeS: 3600, EpochJ2000WholeSPresent: true, LatDeg: -40, LonDeg: -20, VTECPresent: true, VTECTECU: 3},
		{TimeScale: UTC, EpochJ2000S: 3600, EpochJ2000WholeS: 3600, EpochJ2000WholeSPresent: true, LatDeg: -40, LonDeg: 20, VTECPresent: true, VTECTECU: 4},
	}
	product, err := NewIONEXFromTECSamplesWithHeader(samples, 450, 6371, 0, header)
	if err != nil {
		t.Fatalf("build TEC samples with header: %v", err)
	}
	closeAfterTest(t, product)
	gotHeader, err := product.HeaderMetadata()
	if err != nil {
		t.Fatalf("read constructor header: %v", err)
	}
	if gotHeader.MappingFunctionCode != "XMAP" || len(gotHeader.Descriptions) != 1 || gotHeader.Descriptions[0] != "sample constructor" || len(gotHeader.Comments) != 1 || gotHeader.Comments[0] != "detached header" {
		t.Fatalf("constructor header was not retained: %#v", gotHeader)
	}
	gotSamples, err := product.TECSamples()
	if err != nil {
		t.Fatalf("read constructed TEC samples: %v", err)
	}
	if len(gotSamples) != len(samples) || !gotSamples[0].VTECPresenceKnown || gotSamples[0].VTECPresent || !gotSamples[0].RMSPresent || gotSamples[0].RMSTECU != 0 || !gotSamples[0].HeightOffsetPresent || gotSamples[0].HeightOffsetKm != 12.5 || !gotSamples[1].VTECPresent || gotSamples[1].VTECTECU != 2 || gotSamples[1].RMSTECU != 0.25 || gotSamples[1].HeightOffsetKm != 13.5 {
		t.Fatalf("constructor changed sample fields or ordering: %#v", gotSamples)
	}
	if err := product.Close(); err != nil {
		t.Fatal(err)
	}
	if gotHeader.MappingFunctionCode != "XMAP" || gotHeader.Descriptions[0] != "sample constructor" || gotHeader.Comments[0] != "detached header" {
		t.Fatalf("header snapshot changed after product close: %#v", gotHeader)
	}
}

func TestIONEXBuildOutcomeRetainsTypedRefusal(t *testing.T) {
	grid, outcome, err := NewIONEXFromTECGridSamplesWithOutcome(TECGridSamples{
		TimeScale: UTC, MapEpochsJ2000S: []float64{0}, LatNodesDeg: []float64{1, -1}, LonNodesDeg: []float64{0, 1},
		DLatDeg: -2, DLonDeg: 1, ShellHeightKm: 450, BaseRadiusKm: 6371,
		TECMAPsTECU: []float64{math.NaN(), 2, 3, 4},
	})
	if err != nil || grid != nil || outcome.IsOK || outcome.Status != StatusInvalidArgument || outcome.Error == nil || outcome.Error.Kind != IONEXBuildErrorNonFiniteValue || outcome.Error.Input != IONEXBuildInputTECValue || !outcome.Error.HasIndex || outcome.Error.Index != 0 || outcome.Error.Message == "" {
		t.Fatalf("typed grid refusal = product %v, outcome %+v, err %v", grid, outcome, err)
	}
	header := IONEXHeaderMetadata{Version: 1, MappingFunction: IONEXMappingFunctionOther, MappingFunctionCode: "XMAP"}
	samples := []TECSample{
		{TimeScale: UTC, EpochJ2000S: 0.5, LatDeg: 40, LonDeg: -20},
		{TimeScale: UTC, EpochJ2000S: 0.5, LatDeg: 40, LonDeg: 20},
		{TimeScale: UTC, EpochJ2000S: 0.5, LatDeg: -40, LonDeg: -20},
		{TimeScale: UTC, EpochJ2000S: 0.5, LatDeg: -40, LonDeg: 20},
	}
	product, outcome, err := NewIONEXFromTECSamplesWithOutcome(samples, 450, 6371, 0)
	if err != nil || product != nil || outcome.IsOK || outcome.Error == nil || outcome.Error.Kind != IONEXBuildErrorEpochNotRepresentable || outcome.Error.Input != IONEXBuildInputSampleEpoch || !outcome.Error.HasIndex || outcome.Error.Index != 0 || outcome.Error.Message == "" {
		t.Fatalf("typed per-node refusal without header = product %v, outcome %+v, err %v", product, outcome, err)
	}
	product, outcome, err = NewIONEXFromTECSamplesWithHeaderOutcome(samples, 450, 6371, 0, header)
	if err != nil || product != nil || outcome.IsOK || outcome.Status != StatusInvalidArgument || outcome.Error == nil || outcome.Error.Kind != IONEXBuildErrorEpochNotRepresentable || !outcome.Error.HasIndex || outcome.Error.Index != 0 || outcome.Error.Message == "" {
		t.Fatalf("typed per-node refusal = product %v, outcome %+v, err %v", product, outcome, err)
	}
}

func TestIONEXGridBuildOutcomeClassifiesNativeValidation(t *testing.T) {
	base := TECGridSamples{TimeScale: UTC, MapEpochsJ2000S: []float64{0}, LatNodesDeg: []float64{1, -1}, LonNodesDeg: []float64{0, 1}, DLatDeg: -2, DLonDeg: 1, ShellHeightKm: 450, BaseRadiusKm: 6371, TECMAPsTECU: []float64{1, 2, 3, 4}}
	cases := []struct {
		name string
		edit func(*TECGridSamples)
		kind IONEXBuildErrorKind
	}{
		{"empty", func(v *TECGridSamples) { v.MapEpochsJ2000S = nil; v.TECMAPsTECU = nil }, IONEXBuildErrorEmpty},
		{"too-few-latitude-nodes", func(v *TECGridSamples) { v.LatNodesDeg = []float64{1}; v.DLatDeg = 1; v.TECMAPsTECU = []float64{1, 2} }, IONEXBuildErrorTooFewNodes},
		{"tec-count", func(v *TECGridSamples) { v.TECMAPsTECU = []float64{1, 2, 3} }, IONEXBuildErrorShapeMismatch},
		{"rms-count", func(v *TECGridSamples) { v.RMSPresent = true; v.RMSMAPsTECU = []float64{1} }, IONEXBuildErrorRMSCountMismatch},
		{"height-count", func(v *TECGridSamples) { v.HeightPresent = true; v.HeightMapsKm = []float64{450} }, IONEXBuildErrorHeightCountMismatch},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := base
			tc.edit(&input)
			product, outcome, err := NewIONEXFromTECGridSamplesWithOutcome(input)
			if err != nil || product != nil || outcome.IsOK || outcome.Error == nil || outcome.Error.Kind != tc.kind || outcome.Error.Message == "" {
				t.Fatalf("native outcome = product %v, outcome %+v, err %v", product, outcome, err)
			}
			message := outcome.Error.Error()
			laterProduct, later, laterErr := NewIONEXFromTECGridSamplesWithOutcome(base)
			if laterErr != nil || !later.IsOK || later.Error != nil || outcome.Error.Message == "" || outcome.Error.Error() != message {
				t.Fatalf("detached refusal after later native call: original=%+v later=%+v err=%v", outcome.Error, later, laterErr)
			}
			closeAfterTest(t, laterProduct)
		})
	}
	product, outcome, err := NewIONEXFromTECGridSamplesWithOutcome(base)
	if err != nil || product == nil || !outcome.IsOK || outcome.Status != StatusOK || outcome.Error != nil {
		t.Fatalf("successful native outcome = product %v, outcome %+v, err %v", product, outcome, err)
	}
	closeAfterTest(t, product)
}

func TestIONEXPresenceMaskCountsAreCheckedWithoutHeight(t *testing.T) {
	base := TECGridSamples{TimeScale: UTC, MapEpochsJ2000S: []float64{0}, LatNodesDeg: []float64{1, -1}, LonNodesDeg: []float64{0, 1}, DLatDeg: -2, DLonDeg: 1, ShellHeightKm: 450, BaseRadiusKm: 6371, TECMAPsTECU: []float64{1, 2, 3, 4}}
	for _, tc := range []struct {
		name string
		edit func(*TECGridSamples)
	}{
		{"vtec-short", func(v *TECGridSamples) { v.TECMAPsPresent = []bool{true} }},
		{"vtec-long", func(v *TECGridSamples) { v.TECMAPsPresent = []bool{true, true, true, true, true} }},
		{"rms-short", func(v *TECGridSamples) {
			v.RMSPresent = true
			v.RMSMAPsTECU = []float64{1, 2, 3, 4}
			v.RMSMAPsPresent = []bool{true}
		}},
		{"rms-long", func(v *TECGridSamples) {
			v.RMSPresent = true
			v.RMSMAPsTECU = []float64{1, 2, 3, 4}
			v.RMSMAPsPresent = []bool{true, true, true, true, true}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := base
			tc.edit(&input)
			if product, err := NewIONEXFromTECGridSamples(input); err == nil || product != nil {
				if product != nil {
					closeAfterTest(t, product)
				}
				t.Fatalf("convenience constructor accepted malformed presence mask: product=%v err=%v", product, err)
			}
			if product, outcome, err := NewIONEXFromTECGridSamplesWithOutcome(input); err == nil || product != nil || outcome.IsOK {
				if product != nil {
					closeAfterTest(t, product)
				}
				t.Fatalf("outcome constructor accepted malformed presence mask: product=%v outcome=%+v err=%v", product, outcome, err)
			}
		})
	}
	badRMS := base
	badRMS.RMSPresent = true
	badRMS.RMSMAPsTECU = []float64{1}
	badRMS.RMSMAPsPresent = []bool{true, true}
	product, outcome, err := NewIONEXFromTECGridSamplesWithOutcome(badRMS)
	if err != nil || product != nil || outcome.IsOK || outcome.Error == nil || outcome.Error.Kind != IONEXBuildErrorRMSCountMismatch {
		t.Fatalf("value-count refusal must precede mask access: product=%v outcome=%+v err=%v", product, outcome, err)
	}
}

func TestIONEXGridPresenceMasksRetainMissingAndPresentZero(t *testing.T) {
	product, outcome, err := NewIONEXFromTECGridSamplesWithOutcome(TECGridSamples{
		TimeScale: UTC, MapEpochsJ2000S: []float64{0}, LatNodesDeg: []float64{1, -1}, LonNodesDeg: []float64{0, 1}, DLatDeg: -2, DLonDeg: 1,
		ShellHeightKm: 450, BaseRadiusKm: 6371, TECMAPsTECU: []float64{99, 0, 2, 3}, TECMAPsPresent: []bool{false, true, true, true},
		RMSPresent: true, RMSMAPsTECU: []float64{99, 0, 0.5, 0.75}, RMSMAPsPresent: []bool{false, true, true, true},
	})
	if err != nil || product == nil || !outcome.IsOK || outcome.Error != nil {
		t.Fatalf("presence-mask build outcome=%+v product=%v err=%v", outcome, product, err)
	}
	closeAfterTest(t, product)
	assertMasks := func(label string, value *IONEX) {
		t.Helper()
		vtec, err := value.TECMAPsTECU()
		if err != nil || len(vtec) != 4 || !math.IsNaN(vtec[0]) || vtec[1] != 0 || vtec[2] != 2 || vtec[3] != 3 {
			t.Fatalf("%s VTEC values=%v err=%v", label, vtec, err)
		}
		vtecPresence, err := value.TECMapPresence()
		if err != nil || !reflect.DeepEqual(vtecPresence, []bool{false, true, true, true}) {
			t.Fatalf("%s VTEC presence=%v err=%v", label, vtecPresence, err)
		}
		rms, err := value.RMSMAPsTECU()
		if err != nil || len(rms) != 4 || !math.IsNaN(rms[0]) || rms[1] != 0 || rms[2] != 0.5 || rms[3] != 0.75 {
			t.Fatalf("%s RMS values=%v err=%v", label, rms, err)
		}
		rmsPresence, err := value.RMSMapPresence()
		if err != nil || !reflect.DeepEqual(rmsPresence, []bool{false, true, true, true}) {
			t.Fatalf("%s RMS presence=%v err=%v", label, rmsPresence, err)
		}
	}
	assertMasks("constructed", product)
	text, err := product.ToIONEXText()
	if err != nil {
		t.Fatal(err)
	}
	roundTrip, err := ParseIONEX(text)
	if err != nil {
		t.Fatalf("parse writer output: %v", err)
	}
	closeAfterTest(t, roundTrip)
	assertMasks("writer roundtrip", roundTrip)
}

const ionexExactFirstMapJ2000S int64 = 631108800

func ionexExactClock(scale TimeScale, j2000Nanos int64) ClockEpoch {
	return ClockEpoch{Scale: scale, Representation: RINEXClockInstantNanos, NanosLow: uint64(j2000Nanos)}
}

func newIONEXExactTestProduct(t *testing.T) *IONEX {
	t.Helper()
	value, err := NewIONEXFromTECGridSamples(TECGridSamples{
		TimeScale:       UTC,
		MapEpochsJ2000S: []float64{float64(ionexExactFirstMapJ2000S), float64(ionexExactFirstMapJ2000S + 3600)},
		LatNodesDeg:     []float64{40, -40}, LonNodesDeg: []float64{-20, 20}, DLatDeg: -80, DLonDeg: 40,
		ShellHeightKm: 450, BaseRadiusKm: 6371, Exponent: 0,
		TECMAPsTECU: []float64{10, 10, 10, 10, 20, 20, 20, 20},
	})
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func TestIONEXExactInstantFractionScaleBatchAndOwnership(t *testing.T) {
	product := newIONEXExactTestProduct(t)
	t.Cleanup(func() { _ = product.Close() })
	first := ionexExactClock(UTC, ionexExactFirstMapJ2000S*1_000_000_000)
	last := ionexExactClock(UTC, (ionexExactFirstMapJ2000S+3600)*1_000_000_000)
	middle := ionexExactClock(UTC, ionexExactFirstMapJ2000S*1_000_000_000+1_800_500_000_000)
	gpstEquivalent := ionexExactClock(GPST, ionexExactFirstMapJ2000S*1_000_000_000+1_818_500_000_000)
	frequency := 1_575_420_000.0
	query := func(epoch ClockEpoch) float64 {
		value, epochError, err := product.SlantDelayAtInstant(0, 0, 0, 90, epoch, frequency)
		if err != nil || epochError != nil {
			t.Fatalf("exact scalar epoch error=%+v err=%v", epochError, err)
		}
		return value
	}
	d0, d1, dm := query(first), query(last), query(middle)
	want := d0 + (d1-d0)*(1800.5/3600.0)
	if math.Abs(dm-want) > math.Max(1e-12, math.Abs(want)*1e-12) {
		t.Fatalf("fractional interpolation %.17g, independent linear reference %.17g", dm, want)
	}
	requests := []IONEXInstantSlantRequest{
		{LatDeg: 0, LonDeg: 0, ElevationDeg: 90, Epoch: middle, FrequencyHz: frequency},
		{LatDeg: 0, LonDeg: 0, ElevationDeg: 90, Epoch: gpstEquivalent, FrequencyHz: frequency},
	}
	out := make([]IONEXInstantSlantResult, len(requests))
	if err := product.SlantDelayResultsAtInstants(requests, out, DefaultIONEXSlantPolicy()); err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 || !out[0].IsOK || !out[1].IsOK || math.Float64bits(out[0].Evaluation.DelayM) != math.Float64bits(out[1].Evaluation.DelayM) {
		t.Fatalf("UTC/GPST batch results=%+v", out)
	}
	if math.Float64bits(out[0].Evaluation.DelayM) != math.Float64bits(dm) {
		t.Fatalf("batch delay bits differ from exact scalar: %.17g %.17g", out[0].Evaluation.DelayM, dm)
	}
	tdb := ionexExactClock(TDB, ionexExactFirstMapJ2000S*1_000_000_000+1_800_500_000_000)
	epochRows := make([]IONEXInstantSlantResult, 1)
	epochErr := product.SlantDelayResultsAtInstants([]IONEXInstantSlantRequest{{LatDeg: 0, LonDeg: 0, ElevationDeg: 90, Epoch: tdb, FrequencyHz: frequency}}, epochRows, DefaultIONEXSlantPolicy())
	if epochErr != nil || len(epochRows) != 1 || epochRows[0].IsOK || epochRows[0].EpochError == nil || epochRows[0].EpochError.Kind != IONEXEpochErrorNoExactUTCOffset || epochRows[0].Error != nil {
		t.Fatalf("typed TDB epoch refusal=%+v err=%v", epochRows, epochErr)
	}
	owned, err := product.SlantDelayResultsAtInstantsOwned(requests, DefaultIONEXSlantPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if len(owned.Rows) != 2 || owned.Rows[0].Message != "" {
		t.Fatalf("owned successful rows=%+v", owned.Rows)
	}
	if err := product.Close(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(owned.Rows[0].Evaluation, out[0].Evaluation) {
		t.Fatal("owned row changed after source close")
	}
}

func TestIONEXExactInstantKeepsStrictFractionalBoundaryAndSelection(t *testing.T) {
	product := newIONEXExactTestProduct(t)
	closeAfterTest(t, product)
	after := ionexExactClock(UTC, (ionexExactFirstMapJ2000S+3600)*1_000_000_000+500_000_000)
	row, err := product.SlantDelayAtInstantWithPolicy(0, 0, 0, 90, after, 1_575_420_000, DefaultIONEXSlantPolicy())
	if err == nil || row.IsOK || row.EpochError != nil || row.Error == nil || row.Error.Kind != IONEXSlantErrorOutOfCoverage || row.Error.CoverageError != IONEXCoverageErrorEpochAfterLastMap {
		t.Fatalf("strict fractional boundary row=%+v err=%v", row, err)
	}
	owned, ownedErr := product.SlantDelayResultsAtInstantsOwned([]IONEXInstantSlantRequest{{LatDeg: 0, LonDeg: 0, ElevationDeg: 90, Epoch: after, FrequencyHz: 1_575_420_000}}, DefaultIONEXSlantPolicy())
	if ownedErr != nil || len(owned.Rows) != 1 || owned.Rows[0].Error == nil || owned.Rows[0].Error.Kind != IONEXSlantErrorOutOfCoverage || owned.Rows[0].Message == "" {
		t.Fatalf("owned strict refusal=%+v err=%v", owned, ownedErr)
	}
	requested := ionexExactClock(UTC, ionexExactFirstMapJ2000S*1_000_000_000+1_800_500_000_000)
	selected, metadata, epochError, err := SelectIONEXAtInstant([]*IONEX{product}, requested, StalenessPolicyDefault())
	if err != nil || selected == nil || epochError != nil || metadata.RequestedEpochJ2000S != float64(ionexExactFirstMapJ2000S)+1800.5 {
		t.Fatalf("exact selection=%v metadata=%+v epoch=%+v err=%v", selected, metadata, epochError, err)
	}
	closeAfterTest(t, selected)
	start := ionexExactClock(UTC, ionexExactFirstMapJ2000S*1_000_000_000+1_800_250_000_000)
	end := ionexExactClock(GPST, ionexExactFirstMapJ2000S*1_000_000_000+1_836_750_000_000)
	rangeProduct, rangeMetadata, epochError, err := SelectIONEXOverInstantRange([]*IONEX{product}, start, end, StalenessPolicyDefault())
	if err != nil || rangeProduct == nil || epochError != nil || rangeMetadata.RequestedEpochJ2000S != float64(ionexExactFirstMapJ2000S)+1818.75 {
		t.Fatalf("exact range selection=%v metadata=%+v epoch=%+v err=%v", rangeProduct, rangeMetadata, epochError, err)
	}
	closeAfterTest(t, rangeProduct)
	if err := product.Close(); err != nil {
		t.Fatal(err)
	}
	if len(owned.Rows) != 1 || owned.Rows[0].Error == nil || owned.Rows[0].Message == "" {
		t.Fatalf("owned strict refusal lost details after source close: %+v", owned.Rows)
	}
}

func sameIONEXSampleRowsBits(a, b []TECSample) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		left, right := reflect.ValueOf(a[i]), reflect.ValueOf(b[i])
		for field := 0; field < left.NumField(); field++ {
			if left.Field(field).Kind() == reflect.Float64 {
				if math.Float64bits(left.Field(field).Float()) != math.Float64bits(right.Field(field).Float()) {
					return false
				}
			} else if !reflect.DeepEqual(left.Field(field).Interface(), right.Field(field).Interface()) {
				return false
			}
		}
	}
	return true
}

func TestIONEXHeaderCloneAndClearRecipeRetainsPresenceAwareSamples(t *testing.T) {
	const epoch int64 = 9007199254740993
	header := IONEXHeaderMetadata{
		Version: 1, Date: "2026-09-29 00:00 UTC", Program: "sidereon", RunBy: "tests", SatelliteSystem: "GPS",
		HasSatelliteCount: true, SatelliteCount: 7, HasStationCount: true, StationCount: 3,
		ElevationCutoffDeg: 12.5, IntervalS: 3600, HasMapsInFile: true, MapsInFile: 1,
		ObservablesUsed: "L1 L2", MappingDeclaration: IONEXMappingDeclarationDeclared,
		MappingFunction: IONEXMappingFunctionOther, MappingFunctionCode: "XMAP",
		Descriptions: []string{"first description", "second description"}, Comments: []string{"first comment", "second comment"},
	}
	samples := []TECSample{
		{TimeScale: UTC, EpochJ2000S: 0, EpochJ2000WholeS: epoch, EpochJ2000WholeSPresent: true, LatDeg: 40, LonDeg: -20, VTECPresenceKnown: true, VTECPresent: false, RMSPresent: true, RMSTECU: 0, HeightOffsetPresent: true, HeightOffsetKm: 12.5},
		{TimeScale: UTC, EpochJ2000S: 0, EpochJ2000WholeS: epoch, EpochJ2000WholeSPresent: true, LatDeg: 40, LonDeg: 20, VTECPresenceKnown: true, VTECPresent: true, VTECTECU: 2},
		{TimeScale: UTC, EpochJ2000S: 0, EpochJ2000WholeS: epoch, EpochJ2000WholeSPresent: true, LatDeg: -40, LonDeg: -20, VTECPresenceKnown: true, VTECPresent: true, VTECTECU: 3, RMSPresent: true, RMSTECU: 0.25, HeightOffsetPresent: true, HeightOffsetKm: 13.5},
		{TimeScale: UTC, EpochJ2000S: 0, EpochJ2000WholeS: epoch, EpochJ2000WholeSPresent: true, LatDeg: -40, LonDeg: 20, VTECPresenceKnown: true, VTECPresent: true, VTECTECU: 4, HeightOffsetPresent: true, HeightOffsetKm: 14.5},
	}
	product, err := NewIONEXFromTECSamplesWithHeader(samples, 450, 6371, -1, header)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, product)
	originalHeader, err := product.HeaderMetadata()
	if err != nil {
		t.Fatal(err)
	}
	originalSamples, err := product.TECSamples()
	if err != nil {
		t.Fatal(err)
	}
	grid, err := product.GridInfo()
	if err != nil {
		t.Fatal(err)
	}

	clonedHeader := originalHeader
	clonedHeader.Descriptions = append([]string(nil), originalHeader.Descriptions...)
	clonedHeader.Comments = append([]string(nil), originalHeader.Comments...)
	clone, err := NewIONEXFromTECSamplesWithHeader(originalSamples, grid.ShellHeightKm, grid.BaseRadiusKm, grid.Exponent, clonedHeader)
	if err != nil {
		t.Fatalf("rebuild cloned header: %v", err)
	}
	closeAfterTest(t, clone)
	cloneSamples, err := clone.TECSamples()
	if err != nil {
		t.Fatal(err)
	}
	cloneGrid, err := clone.GridInfo()
	if err != nil {
		t.Fatal(err)
	}
	cloneHeader, err := clone.HeaderMetadata()
	if err != nil {
		t.Fatal(err)
	}
	if !sameIONEXSampleRowsBits(cloneSamples, originalSamples) || !reflect.DeepEqual(cloneGrid, grid) || !reflect.DeepEqual(cloneHeader, originalHeader) {
		t.Fatalf("detached clone changed header/grid/sample values:\nheader=%#v\ngrid=%#v\nsamples=%#v", cloneHeader, cloneGrid, cloneSamples)
	}

	checkRebuild := func(label string, rebuiltHeader, wantHeader IONEXHeaderMetadata) *IONEX {
		t.Helper()
		rebuilt, err := NewIONEXFromTECSamplesWithHeader(originalSamples, grid.ShellHeightKm, grid.BaseRadiusKm, grid.Exponent, rebuiltHeader)
		if err != nil {
			t.Fatalf("rebuild with %s: %v", label, err)
		}
		closeAfterTest(t, rebuilt)
		gotHeader, err := rebuilt.HeaderMetadata()
		if err != nil {
			t.Fatal(err)
		}
		gotSamples, err := rebuilt.TECSamples()
		if err != nil {
			t.Fatal(err)
		}
		gotGrid, err := rebuilt.GridInfo()
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(gotHeader, wantHeader) || !sameIONEXSampleRowsBits(gotSamples, originalSamples) || !reflect.DeepEqual(gotGrid, grid) {
			t.Fatalf("%s changed remaining metadata/data: header=%#v grid=%#v samples=%#v", label, gotHeader, gotGrid, gotSamples)
		}
		if len(gotSamples) != 4 || !gotSamples[0].EpochJ2000WholeSPresent || gotSamples[0].EpochJ2000WholeS != epoch || !gotSamples[0].VTECPresenceKnown || gotSamples[0].VTECPresent || !gotSamples[0].RMSPresent || !gotSamples[0].HeightOffsetPresent || gotSamples[0].HeightOffsetKm != 12.5 || !gotSamples[1].VTECPresent || gotSamples[1].VTECTECU != 2 || gotSamples[2].RMSTECU != 0.25 || !gotGrid.RMSPresent || gotGrid.ShellHeightKm != 450 || gotGrid.BaseRadiusKm != 6371 || gotGrid.Exponent != -1 {
			t.Fatalf("%s changed presence, exact epoch, or grid scalars: grid=%#v samples=%#v", label, gotGrid, gotSamples)
		}
		return rebuilt
	}

	commentsCleared := clonedHeader
	commentsCleared.Comments = nil
	wantCommentsCleared := originalHeader
	wantCommentsCleared.Comments = nil
	commentProduct := checkRebuild("comment clear", commentsCleared, wantCommentsCleared)
	descriptionsCleared := clonedHeader
	descriptionsCleared.Descriptions = nil
	wantDescriptionsCleared := originalHeader
	wantDescriptionsCleared.Descriptions = nil
	descriptionProduct := checkRebuild("description clear", descriptionsCleared, wantDescriptionsCleared)
	bothCleared := clonedHeader
	bothCleared.Comments = nil
	bothCleared.Descriptions = nil
	wantBothCleared := originalHeader
	wantBothCleared.Comments = nil
	wantBothCleared.Descriptions = nil
	bothProduct := checkRebuild("comment and description clear", bothCleared, wantBothCleared)
	for label, product := range map[string]*IONEX{"comments": commentProduct, "descriptions": descriptionProduct, "both": bothProduct} {
		if err := product.Close(); err != nil {
			t.Errorf("close %s rebuild: %v", label, err)
		}
	}
	if len(cloneHeader.Comments) != 2 || len(cloneHeader.Descriptions) != 2 {
		t.Fatalf("clearing copied slices mutated the cloned product: %#v", cloneHeader)
	}
}
