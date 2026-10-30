//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package sidereon

import (
	"fmt"
	"math"
	"os"
	"reflect"
	"strings"
	"testing"
)

func accuracyFixtureWithBases(t *testing.T, name, positionBase, clockBase string) string {
	t.Helper()
	data, err := os.ReadFile("testdata/sp3_accuracy_306/" + name + ".sp3")
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	old := "%f  1.2500000  1.025000000"
	if name == "A-velocity" {
		old = "%f  2.0000000  2.000000000"
	}
	new := fmt.Sprintf("%%f  %9s  %11s", positionBase, clockBase)
	if !strings.Contains(text, old) {
		t.Fatalf("fixture %s lacks bases %q", name, old)
	}
	return strings.Replace(text, old, new, 1)
}

func accuracyFixtureWithCodes(t *testing.T, text, record string, codes [4]int) string {
	t.Helper()
	lines := strings.Split(text, "\n")
	found := 0
	for index, line := range lines {
		if !strings.HasPrefix(line, record) {
			continue
		}
		if len(line) < 73 {
			t.Fatalf("short accuracy record %q", line)
		}
		line = line[:61] + fmt.Sprintf("%2d %2d %2d %3d", codes[0], codes[1], codes[2], codes[3]) + line[73:]
		lines[index] = line
		found++
	}
	if found != 2 {
		t.Fatalf("record %s matched %d epochs, want 2", record, found)
	}
	return strings.Join(lines, "\n")
}

func TestSP3AccuracyRoundingPreservesConservativeSigmas(t *testing.T) {
	read := func(name string) []byte {
		t.Helper()
		data, err := os.ReadFile("testdata/sp3_accuracy_306/" + name + ".sp3")
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	a, err := LoadSP3(read("A-velocity"))
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, a)
	b, err := LoadSP3(read("B-velocity"))
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, b)

	// The source files are independently writable and decode to the fixture's
	// analytically defined position sigmas at both epochs.
	for name, source := range map[string]*SP3{"A": a, "B": b} {
		text, outcome, err := source.Encode()
		if err != nil || !outcome.IsOK || len(text) == 0 {
			t.Fatalf("source %s strict write = %+v, %v", name, outcome, err)
		}
		for epoch := 0; epoch < 2; epoch++ {
			id, want := "G01", 0.002
			base, exponent := 2.0, 1
			clockBase := 2.0
			if name == "B" {
				id, want = "G12", math.Pow(1.25, 13)*1e-3
				base, exponent = 1.25, 13
				clockBase = 1.025
			}
			accuracy, err := source.RecordAccuracy(id, epoch)
			if err != nil || !accuracy.HasP || accuracy.P.PositionSigmaM[0].Kind != SP3AccuracyKnown || math.Abs(accuracy.P.PositionSigmaM[0].Value-want) > 1e-15 {
				t.Fatalf("source %s %s epoch %d sigma = %+v, err %v; want %.17g", name, id, epoch, accuracy, err, want)
			}
			velocity, err := source.RecordAccuracy(id, epoch)
			if err != nil || !velocity.HasV {
				t.Fatalf("source %s %s epoch %d velocity accuracy = %+v, err %v", name, id, epoch, velocity, err)
			}
			wantVelocity := math.Pow(base, float64(exponent)) * 1e-7
			wantClock := math.Pow(clockBase, float64(exponent)) * 299792458e-12
			wantRate := math.Pow(clockBase, float64(exponent)) * 299792458e-16
			for axis := 0; axis < 3; axis++ {
				if velocity.P.PositionSigmaM[axis].Kind != SP3AccuracyKnown || velocity.V.VelocitySigmaMPerS[axis].Kind != SP3AccuracyKnown {
					t.Fatalf("source %s %s epoch %d axis %d accuracy is not known: %+v", name, id, epoch, axis, velocity)
				}
			}
			if velocity.P.ClockSigmaM.Kind != SP3AccuracyKnown || velocity.V.ClockRateSigmaMPerS.Kind != SP3AccuracyKnown || math.Abs(velocity.V.VelocitySigmaMPerS[0].Value-wantVelocity) > 1e-15 || math.Abs(velocity.P.ClockSigmaM.Value-wantClock) > 1e-15 || math.Abs(velocity.V.ClockRateSigmaMPerS.Value-wantRate) > 1e-15 {
				t.Fatalf("source %s %s epoch %d decoded all-field accuracy = %+v", name, id, epoch, velocity)
			}
		}
	}
	directText, directReport, err := b.ToSP3StringWithAccuracyRounding(SP3AccuracyRoundingConservative)
	if err != nil || directText == "" || directReport.AccuracySigmas != (SP3AccuracyRoundingReport{}) {
		t.Fatalf("direct all-field accuracy roundtrip report = %+v, err %v", directReport.AccuracySigmas, err)
	}
	directReparsed, err := LoadSP3([]byte(directText))
	if err != nil {
		t.Fatalf("parse direct all-field output: %v", err)
	}
	closeAfterTest(t, directReparsed)
	for epoch := 0; epoch < 2; epoch++ {
		before, err := b.RecordAccuracy("G12", epoch)
		if err != nil {
			t.Fatal(err)
		}
		after, err := directReparsed.RecordAccuracy("G12", epoch)
		if err != nil || !reflect.DeepEqual(before, after) {
			t.Fatalf("direct writer changed all-field G12 accuracy epoch %d: before %+v after %+v err %v", epoch, before, after, err)
		}
	}

	options, err := NewSP3MergeOptions()
	if err != nil {
		t.Fatal(err)
	}
	options.MinAgree = 1
	options.ClockMinCommon = 1
	merged, report, err := MergeSP3([]*SP3{a, b}, &options)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, merged)
	closeAfterTest(t, report)
	_, strictOutcome, err := merged.Encode()
	if err != nil || strictOutcome.IsOK || strictOutcome.Error.Kind != SP3WriteErrorAccuracyNotRepresentable {
		t.Fatalf("strict merged write = %+v, %v; want accuracy representability refusal", strictOutcome, err)
	}
	if _, _, err := merged.ToSP3StringWithAccuracyRounding(SP3AccuracyRoundingPolicy(99)); err == nil {
		t.Fatal("unknown accuracy rounding policy unexpectedly selected a conversion")
	}

	text, detailed, err := merged.ToSP3StringWithAccuracyRounding(SP3AccuracyRoundingConservative)
	if err != nil || text == "" {
		t.Fatalf("conservative accuracy write = %q, %+v, %v", text, detailed, err)
	}
	wantSource := math.Pow(1.25, 13) * 1e-3
	wantSigma := 0.032
	wantClockSource := math.Pow(1.025, 13) * 299792458e-12
	wantClockSigma := 2 * 299792458e-12
	wantReport := SP3AccuracyRoundingReport{
		PositionSigmaXM: wantSigma - wantSource, PositionSigmaYM: wantSigma - wantSource, PositionSigmaZM: wantSigma - wantSource,
		ClockSigmaM: wantClockSigma - wantClockSource,
	}
	gotReport := detailed.AccuracySigmas
	gotValues := []float64{gotReport.PositionSigmaXM, gotReport.PositionSigmaYM, gotReport.PositionSigmaZM, gotReport.VelocitySigmaXMS, gotReport.VelocitySigmaYMS, gotReport.VelocitySigmaZMS, gotReport.ClockSigmaM, gotReport.ClockRateSigmaMS}
	wantValues := []float64{wantReport.PositionSigmaXM, wantReport.PositionSigmaYM, wantReport.PositionSigmaZM, wantReport.VelocitySigmaXMS, wantReport.VelocitySigmaYMS, wantReport.VelocitySigmaZMS, wantReport.ClockSigmaM, wantReport.ClockRateSigmaMS}
	for index, want := range wantValues {
		if math.Abs(gotValues[index]-want) > 1e-15 {
			t.Fatalf("accuracy report field %d = %.17g, want %.17g (%+v)", index, gotValues[index], want, gotReport)
		}
	}
	if detailed.RecordValues != (SP3RoundingReport{}) {
		t.Fatalf("grid-aligned records changed during accuracy conversion: %+v", detailed.RecordValues)
	}
	if detailed.AccuracySigmas == (SP3AccuracyRoundingReport{}) {
		t.Fatalf("unrelated accuracy report fields changed: %+v", detailed.AccuracySigmas)
	}

	reparsed, err := LoadSP3([]byte(text))
	if err != nil {
		t.Fatalf("parse accuracy-rounded output: %v", err)
	}
	closeAfterTest(t, reparsed)
	for epoch := 0; epoch < 2; epoch++ {
		for _, id := range []string{"G01", "G12"} {
			before, err := merged.RecordAccuracy(id, epoch)
			if err != nil {
				t.Fatal(err)
			}
			after, err := reparsed.RecordAccuracy(id, epoch)
			if err != nil || !after.HasP || after.HasV || after.P.PositionSigmaM[0].Kind != SP3AccuracyKnown {
				t.Fatalf("reparsed %s epoch %d accuracy = %+v, err %v", id, epoch, after, err)
			}
			wantBase, exponent := 2.0, 1
			if id == "G12" {
				wantBase, exponent = 2.0, 5
			}
			wantPos := math.Pow(wantBase, float64(exponent)) * 1e-3
			wantClock := 2 * 299792458e-12
			for axis := 0; axis < 3; axis++ {
				if math.Abs(after.P.PositionSigmaM[axis].Value-wantPos) > 1e-15 {
					t.Fatalf("rounded %s epoch %d axis %d position sigma = %.17g", id, epoch, axis, after.P.PositionSigmaM[axis].Value)
				}
			}
			if math.Abs(after.P.ClockSigmaM.Value-wantClock) > 1e-15 {
				t.Fatalf("rounded %s epoch %d clock sigma = %.17g", id, epoch, after.P.ClockSigmaM.Value)
			}
			if id == "G12" {
				sourcePos := math.Pow(1.25, 13) * 1e-3
				sourceClock := math.Pow(1.025, 13) * 299792458e-12
				if math.Abs(before.P.PositionSigmaM[0].Value-sourcePos) > 1e-15 || math.Abs(before.P.ClockSigmaM.Value-sourceClock) > 1e-15 {
					t.Fatalf("merge altered G12 source sigmas epoch %d: %+v", epoch, before)
				}
			}
		}
		anchor, err := reparsed.State("G02", epoch)
		wantClock := -20e-6
		if epoch == 1 {
			wantClock = -19e-6
		}
		if err != nil || !anchor.HasClock || math.Abs(anchor.ClockS-wantClock) > 1e-18 {
			t.Fatalf("shared G02 clock anchor epoch %d = %+v, err %v", epoch, anchor, err)
		}
		anchorAccuracy, err := reparsed.RecordAccuracy("G02", epoch)
		if err != nil || !anchorAccuracy.HasP || anchorAccuracy.HasV || anchorAccuracy.P.ClockSigmaM.Kind != SP3AccuracyUnknown {
			t.Fatalf("rounded G02 clock sigma epoch %d = %+v, err %v", epoch, anchorAccuracy.P.ClockSigmaM, err)
		}
		for _, id := range []string{"G01", "G02", "G12"} {
			before, err := merged.State(id, epoch)
			if err != nil {
				t.Fatal(err)
			}
			after, err := reparsed.State(id, epoch)
			if err != nil || !reflect.DeepEqual(before, after) {
				t.Fatalf("record values changed for %s epoch %d: before %+v after %+v err %v", id, epoch, before, after, err)
			}
		}
	}
	if _, outcome, err := reparsed.Encode(); err != nil || !outcome.IsOK {
		t.Fatalf("strict write after conservative accuracy conversion = %+v, %v", outcome, err)
	}
}

func TestSP3AccuracyConservativeRoundingRefusesOverflowingHeaderRange(t *testing.T) {
	aBytes, err := os.ReadFile("testdata/sp3_accuracy_306/A-velocity.sp3")
	if err != nil {
		t.Fatal(err)
	}
	bBytes, err := os.ReadFile("testdata/sp3_accuracy_306/B-velocity.sp3")
	if err != nil {
		t.Fatal(err)
	}
	// Reverse the bases and make B's G12 exponents 98. A's known exponent 1
	// makes the bases incompatible, while the source sigma exceeds every
	// finite accuracy value representable at A's retained base.
	aText := strings.Replace(string(aBytes), "%f  2.0000000", "%f  1.2500000", 1)
	bText := strings.Replace(string(bBytes), "%f  1.2500000", "%f  2.0000000", 1)
	bText = strings.ReplaceAll(bText, "13 13 13  13", "98 98 98  98")
	if aText == string(aBytes) || bText == string(bBytes) {
		t.Fatal("boundary fixtures did not contain expected bases and exponents")
	}
	a, err := LoadSP3([]byte(aText))
	if err != nil {
		t.Fatalf("parse target-base fixture: %v", err)
	}
	closeAfterTest(t, a)
	b, err := LoadSP3([]byte(bText))
	if err != nil {
		t.Fatalf("parse high-exponent fixture: %v", err)
	}
	closeAfterTest(t, b)
	options, err := NewSP3MergeOptions()
	if err != nil {
		t.Fatal(err)
	}
	options.MinAgree = 1
	options.ClockMinCommon = 1
	merged, report, err := MergeSP3([]*SP3{a, b}, &options)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, merged)
	closeAfterTest(t, report)
	_, strict, err := merged.Encode()
	if err != nil || strict.IsOK || strict.Error.Kind != SP3WriteErrorAccuracyNotRepresentable {
		t.Fatalf("strict boundary write = %+v, %v", strict, err)
	}
	text, _, err := merged.ToSP3StringWithAccuracyRounding(SP3AccuracyRoundingConservative)
	refusal, ok := err.(SP3WriteRefusalError)
	if !ok || text != "" || refusal.Outcome.Error.Kind != SP3WriteErrorAccuracyNotRepresentable {
		t.Fatalf("overflowing conservative accuracy write = %q, %T %v", text, err, err)
	}
}

func TestSP3AccuracyMergeBasesAndFiniteExponentBoundaries(t *testing.T) {
	baseA, err := os.ReadFile("testdata/sp3_accuracy_306/A-velocity.sp3")
	if err != nil {
		t.Fatal(err)
	}
	baseB, err := os.ReadFile("testdata/sp3_accuracy_306/B-velocity.sp3")
	if err != nil {
		t.Fatal(err)
	}
	load := func(text string) *SP3 {
		t.Helper()
		product, err := LoadSP3([]byte(text))
		if err != nil {
			t.Fatal(err)
		}
		closeAfterTest(t, product)
		return product
	}
	merge := func(a, b *SP3) *SP3 {
		t.Helper()
		options, err := NewSP3MergeOptions()
		if err != nil {
			t.Fatal(err)
		}
		options.MinAgree, options.ClockMinCommon = 1, 1
		product, report, err := MergeSP3([]*SP3{a, b}, &options)
		if err != nil {
			t.Fatal(err)
		}
		closeAfterTest(t, product)
		closeAfterTest(t, report)
		return product
	}
	// A common basis permits strict writing without changing either source's
	// decoded values.
	compatibleA := accuracyFixtureWithCodes(t, string(baseA), "PG01", [4]int{2, 2, 2, 2})
	compatibleB := accuracyFixtureWithBases(t, "B-velocity", "4.0000000", "4.000000000")
	compatibleB = accuracyFixtureWithCodes(t, compatibleB, "PG12", [4]int{1, 1, 1, 1})
	compatible := merge(load(compatibleA), load(compatibleB))
	if _, outcome, err := compatible.Encode(); err != nil || !outcome.IsOK {
		t.Fatalf("compatible-basis strict merge = %+v, %v", outcome, err)
	}

	// Reversing incompatible sources changes which basis is retained. In both
	// orders, the explicit writer must conservatively preserve all known source
	// sigmas and disclose a nonzero adjustment.
	a := load(string(baseA))
	b := load(string(baseB))
	for _, sources := range [][]*SP3{{a, b}, {b, a}} {
		merged := merge(sources[0], sources[1])
		if _, outcome, err := merged.Encode(); err != nil || outcome.IsOK || outcome.Error.Kind != SP3WriteErrorAccuracyNotRepresentable {
			t.Fatalf("incompatible-basis strict merge = %+v, %v", outcome, err)
		}
		text, detailed, err := merged.ToSP3StringWithAccuracyRounding(SP3AccuracyRoundingConservative)
		if err != nil || text == "" || detailed.AccuracySigmas == (SP3AccuracyRoundingReport{}) {
			t.Fatalf("incompatible-basis conservative merge report = %+v, %v", detailed, err)
		}
		reparsed, err := LoadSP3([]byte(text))
		if err != nil {
			t.Fatal(err)
		}
		closeAfterTest(t, reparsed)
		if sources[0] == b {
			before, e1 := merged.RecordAccuracy("G01", 0)
			after, e2 := reparsed.RecordAccuracy("G01", 0)
			wantPosition := after.P.PositionSigmaM[0].Value - before.P.PositionSigmaM[0].Value
			wantClock := after.P.ClockSigmaM.Value - before.P.ClockSigmaM.Value
			if e1 != nil || e2 != nil || math.Abs(detailed.AccuracySigmas.PositionSigmaXM-wantPosition) > math.Max(math.Abs(wantPosition)*1e-12, 1e-300) || math.Abs(detailed.AccuracySigmas.ClockSigmaM-wantClock) > math.Max(math.Abs(wantClock)*1e-12, 1e-300) {
				t.Fatalf("reverse-order reported deltas %+v differ from G01 decoded changes pos %.17g clock %.17g: %v / %v", detailed.AccuracySigmas, wantPosition, wantClock, e1, e2)
			}
		}
	}

	// Position codes are I2: B's base-4 exponent -5 rounds upward to A's
	// base-2 lower bound -9.
	lowPosA := accuracyFixtureWithCodes(t, string(baseA), "PG01", [4]int{-9, -9, -9, 1})
	lowPosB := accuracyFixtureWithBases(t, "B-velocity", "4.0000000", "1.025000000")
	lowPosB = accuracyFixtureWithCodes(t, lowPosB, "PG12", [4]int{-5, -5, -5, 13})
	lowPos := merge(load(lowPosA), load(lowPosB))
	lowPosText, lowPosReport, err := lowPos.ToSP3StringWithAccuracyRounding(SP3AccuracyRoundingConservative)
	lowPosReparsed, parseErr := LoadSP3([]byte(lowPosText))
	if parseErr != nil {
		t.Fatal(parseErr)
	}
	closeAfterTest(t, lowPosReparsed)
	lowPosBefore, beforeErr := lowPos.RecordAccuracy("G12", 0)
	lowPosAfter, afterErr := lowPosReparsed.RecordAccuracy("G12", 0)
	lowPosDelta := lowPosAfter.P.PositionSigmaM[0].Value - lowPosBefore.P.PositionSigmaM[0].Value
	if err != nil || beforeErr != nil || afterErr != nil || lowPosReport.AccuracySigmas.PositionSigmaXM <= 0 || math.Abs(lowPosReport.AccuracySigmas.PositionSigmaXM-lowPosDelta) > math.Max(math.Abs(lowPosDelta)*1e-12, 1e-300) {
		t.Fatalf("I2 lower-bound conservative report = %+v, %v", lowPosReport, err)
	}

	// Clock codes are I3: B's base-4 exponent -50 rounds upward to A's
	// base-2 lower bound -99.
	lowClockA := accuracyFixtureWithCodes(t, string(baseA), "PG01", [4]int{1, 1, 1, -99})
	lowClockB := accuracyFixtureWithBases(t, "B-velocity", "1.2500000", "4.000000000")
	lowClockB = accuracyFixtureWithCodes(t, lowClockB, "PG12", [4]int{13, 13, 13, -50})
	lowClock := merge(load(lowClockA), load(lowClockB))
	lowClockText, lowClockReport, err := lowClock.ToSP3StringWithAccuracyRounding(SP3AccuracyRoundingConservative)
	lowClockReparsed, parseErr := LoadSP3([]byte(lowClockText))
	if parseErr != nil {
		t.Fatal(parseErr)
	}
	closeAfterTest(t, lowClockReparsed)
	lowClockBefore, beforeErr := lowClock.RecordAccuracy("G12", 0)
	lowClockAfter, afterErr := lowClockReparsed.RecordAccuracy("G12", 0)
	lowClockDelta := lowClockAfter.P.ClockSigmaM.Value - lowClockBefore.P.ClockSigmaM.Value
	if err != nil || beforeErr != nil || afterErr != nil || lowClockReport.AccuracySigmas.ClockSigmaM <= 0 || math.Abs(lowClockReport.AccuracySigmas.ClockSigmaM-lowClockDelta) > math.Max(math.Abs(lowClockDelta)*1e-12, 1e-300) {
		t.Fatalf("I3 lower-bound conservative report = %+v, %v", lowClockReport, err)
	}

	// The upper I2/I3 cases cannot be represented at the retained target base.
	upperPosABases := accuracyFixtureWithBases(t, "A-velocity", "1.2500000", "2.000000000")
	upperPosA := accuracyFixtureWithCodes(t, upperPosABases, "PG01", [4]int{1, 1, 1, 1})
	upperPosB := accuracyFixtureWithBases(t, "B-velocity", "2.0000000", "2.000000000")
	upperPosB = accuracyFixtureWithCodes(t, upperPosB, "PG12", [4]int{98, 98, 98, 13})
	for _, tc := range []struct {
		name string
		a, b string
	}{
		{"I2 upper", upperPosA, upperPosB},
		{"I3 upper", accuracyFixtureWithCodes(t, accuracyFixtureWithBases(t, "A-velocity", "2.0000000", "1.025000000"), "PG01", [4]int{1, 1, 1, 1}), func() string {
			text := accuracyFixtureWithBases(t, "B-velocity", "2.0000000", "2.000000000")
			return accuracyFixtureWithCodes(t, text, "PG12", [4]int{13, 13, 13, 998})
		}()},
	} {
		merged := merge(load(tc.a), load(tc.b))
		if _, strict, err := merged.Encode(); err != nil || strict.IsOK || strict.Error.Kind != SP3WriteErrorAccuracyNotRepresentable {
			t.Errorf("%s strict write = %+v, %v", tc.name, strict, err)
		}
		text, detailed, err := merged.ToSP3StringWithAccuracyRounding(SP3AccuracyRoundingConservative)
		refusal, ok := err.(SP3WriteRefusalError)
		if !ok || text != "" || refusal.Outcome.Error.Kind != SP3WriteErrorAccuracyNotRepresentable {
			t.Errorf("%s rounding = %q, %+v, %T %v", tc.name, text, detailed, err, err)
		}
	}
}

func TestSP3AccuracySentinelAndInvalidBaseSemantics(t *testing.T) {
	bBytes, err := os.ReadFile("testdata/sp3_accuracy_306/B-velocity.sp3")
	if err != nil {
		t.Fatal(err)
	}
	// Sentinel codes remain TooLarge through explicit conversion and parsing.
	sentinelText := accuracyFixtureWithCodes(t, string(bBytes), "PG12", [4]int{99, 99, 99, 999})
	sentinel, err := LoadSP3([]byte(sentinelText))
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, sentinel)
	text, _, err := sentinel.ToSP3StringWithAccuracyRounding(SP3AccuracyRoundingConservative)
	if err != nil {
		t.Fatalf("sentinel conversion: %v", err)
	}
	reparsed, err := LoadSP3([]byte(text))
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, reparsed)
	for epoch := 0; epoch < 2; epoch++ {
		accuracy, err := reparsed.RecordAccuracy("G12", epoch)
		if err != nil || accuracy.P.PositionSigmaM[0].Kind != SP3AccuracyTooLarge || accuracy.P.ClockSigmaM.Kind != SP3AccuracyTooLarge {
			t.Fatalf("sentinel semantics changed at epoch %d: %+v, %v", epoch, accuracy, err)
		}
	}

	// A finite exponent with a zero base remains InvalidBase when its base is unchanged.
	invalidBaseText := strings.Replace(string(bBytes), "%f  1.2500000", "%f  0.0000000", 1)
	invalidBase, err := LoadSP3([]byte(invalidBaseText))
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, invalidBase)
	before, err := invalidBase.RecordAccuracy("G12", 0)
	if err != nil || before.P.PositionSigmaM[0].Kind != SP3AccuracyInvalidBase {
		t.Fatalf("invalid-base decoded value = %+v, %v", before, err)
	}
	text, detailed, err := invalidBase.ToSP3StringWithAccuracyRounding(SP3AccuracyRoundingConservative)
	if err != nil || text == "" || detailed.AccuracySigmas != (SP3AccuracyRoundingReport{}) {
		t.Fatalf("standalone invalid-base conversion did not preserve the source: %+v, %v", detailed, err)
	}
	reparsed, err = LoadSP3([]byte(text))
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, reparsed)
	after, err := reparsed.RecordAccuracy("G12", 0)
	if err != nil || after.P.PositionSigmaM[0].Kind != SP3AccuracyInvalidBase {
		t.Fatalf("standalone invalid-base meaning changed: %+v, %v", after, err)
	}
	originalAfter, err := invalidBase.RecordAccuracy("G12", 0)
	if err != nil || originalAfter.P.PositionSigmaM[0].Kind != SP3AccuracyInvalidBase {
		t.Fatalf("standalone invalid-base input changed: %+v, %v", originalAfter, err)
	}

	knownBytes, err := os.ReadFile("testdata/sp3_accuracy_306/A-velocity.sp3")
	if err != nil {
		t.Fatal(err)
	}
	known, err := LoadSP3(knownBytes)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, known)
	for _, sources := range [][]*SP3{{invalidBase, known}, {known, invalidBase}} {
		options, err := NewSP3MergeOptions()
		if err != nil {
			t.Fatal(err)
		}
		options.MinAgree, options.ClockMinCommon = 1, 1
		merged, report, err := MergeSP3(sources, &options)
		if err != nil {
			t.Fatal(err)
		}
		closeAfterTest(t, merged)
		closeAfterTest(t, report)
		if output, _, err := merged.ToSP3StringWithAccuracyRounding(SP3AccuracyRoundingConservative); output != "" {
			t.Fatalf("mixed invalid/known bases converted in source order %+v: %q", sources, output)
		} else if refusal, ok := err.(SP3WriteRefusalError); !ok || refusal.Outcome.Error.Kind != SP3WriteErrorAccuracyNotRepresentable {
			t.Fatalf("mixed invalid/known bases returned %T %v, want typed accuracy refusal", err, err)
		}
	}
}

func TestSP3MergePreservesKnownAccuracyFromUnbasedSourceInBothOrders(t *testing.T) {
	read := func(name string) []byte {
		t.Helper()
		data, err := os.ReadFile("testdata/sp3_accuracy_306/" + name + ".sp3")
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	a, err := LoadSP3(read("A-unbased"))
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, a)
	b, err := LoadSP3(read("B-based"))
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, b)
	for name, source := range map[string]*SP3{"unbased A": a, "based B": b} {
		if _, outcome, err := source.Encode(); err != nil || !outcome.IsOK {
			t.Fatalf("%s source strict writer = %+v, %v", name, outcome, err)
		}
	}
	wantSigma := math.Pow(1.25, 13) * 1e-3
	for order, sources := range [][]*SP3{{a, b}, {b, a}} {
		options, err := NewSP3MergeOptions()
		if err != nil {
			t.Fatal(err)
		}
		options.Combine = SP3MergeCombinePrecedence
		options.PrecedenceScope = SP3MergePrecedenceCell
		options.MinAgree = 1
		options.ClockMinCommon = 1
		options.PositionToleranceM = 5
		options.ClockToleranceS = 50e-9
		options.OutlierRejectEnabled = true
		options.OutlierRejectPositionTolerance = 5
		options.OutlierRejectClockTolerance = 50e-9
		options.TargetEpochIntervalEnabled = true
		options.TargetEpochIntervalS = 300
		options.Systems = []GNSSSystem{GNSSSystemGPS}
		merged, report, err := MergeSP3(sources, &options)
		if err != nil {
			t.Fatalf("merge order %d: %v", order, err)
		}
		closeAfterTest(t, merged)
		closeAfterTest(t, report)
		text, outcome, err := merged.Encode()
		if err != nil || !outcome.IsOK {
			t.Fatalf("strict merge write = %+v, %v", outcome, err)
		}
		reparsed, err := LoadSP3(text)
		if err != nil {
			t.Fatalf("reparse merged strict output: %v", err)
		}
		closeAfterTest(t, reparsed)
		for epoch := 0; epoch < 2; epoch++ {
			got, err := reparsed.RecordAccuracy("G12", epoch)
			if err != nil || !got.HasP || got.P.PositionSigmaM[0].Kind != SP3AccuracyKnown || math.Abs(got.P.PositionSigmaM[0].Value-wantSigma) > 1e-15 {
				t.Fatalf("merged G12 sigma epoch %d = %+v, err %v; want %.17g", epoch, got, err, wantSigma)
			}
			unknown, err := reparsed.RecordAccuracy("G01", epoch)
			if err != nil || unknown.P.PositionSigmaM[0].Kind != SP3AccuracyUnknown {
				t.Fatalf("unbased A G01 accuracy epoch %d = %+v, err %v", epoch, unknown, err)
			}
		}
	}
}
