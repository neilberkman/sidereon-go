package sidereon

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"math"
	"os"
	"testing"
)

// gappedSP3FixtureB64 contains the gzipped filtered SP3 fixture containing G01
// with a gap between 07:30 and 10:00.
const gappedSP3FixtureB64 = "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"

func loadGappedFixture(t *testing.T) []byte {
	t.Helper()
	if raw, err := os.ReadFile("testdata/GAP_G01_20201760000_15M.sp3"); err == nil && len(raw) > 0 {
		return raw
	}
	compressed, err := base64.StdEncoding.DecodeString(gappedSP3FixtureB64)
	if err != nil {
		t.Fatalf("decode embedded fixture base64: %v", err)
	}
	reader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatalf("decompress embedded fixture: %v", err)
	}
	defer func() { _ = reader.Close() }()
	decompressed, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read decompressed fixture: %v", err)
	}
	return decompressed
}

func TestSP3InterpolationPolicyFixture(t *testing.T) {
	fixture := loadGappedFixture(t)
	const holeMidpointJ2000S = 646260300.0

	// 1. Invalid gap threshold factor (1.0) must return error.
	if _, err := LoadSP3(fixture, WithGapThresholdFactor(1.0)); err == nil {
		t.Fatal("LoadSP3 with factor 1.0 unexpectedly succeeded")
	} else {
		var statusErr *StatusError
		if !errors.As(err, &statusErr) || statusErr.Code != StatusInvalidArgument {
			t.Fatalf("expected StatusInvalidArgument, got %v", err)
		}
	}
	if _, err := LoadSP3WithOptions(fixture, SP3LoadOptions{GapThresholdFactor: 0.5}); err == nil {
		t.Fatal("LoadSP3 with factor 0.5 unexpectedly succeeded")
	}
	if _, err := LoadSP3(fixture, WithGapThresholdFactor(math.NaN())); err == nil {
		t.Fatal("LoadSP3 with NaN factor unexpectedly succeeded")
	}
	if _, err := LoadSP3(fixture, WithGapThresholdFactor(math.Inf(1))); err == nil {
		t.Fatal("LoadSP3 with Inf factor unexpectedly succeeded")
	}

	// 2. Default factor (0.0 or omitted) selects engine default of 1.5 and refuses hole midpoint.
	sp3Default, err := LoadSP3(fixture)
	if err != nil {
		t.Fatalf("LoadSP3 with default factor: %v", err)
	}
	t.Cleanup(func() { _ = sp3Default.Close() })

	defFactor, err := sp3Default.GapThresholdFactor()
	if err != nil {
		t.Fatalf("sp3Default.GapThresholdFactor: %v", err)
	}
	if math.Abs(defFactor-1.5) > 1e-9 {
		t.Fatalf("expected default factor 1.5, got %f", defFactor)
	}

	// Midpoint query must fail on default factor.
	if _, _, _, err := sp3Default.ObservableState("G01", holeMidpointJ2000S); err == nil {
		t.Fatal("default factor unexpectedly served hole midpoint state")
	}
	if _, _, _, err := sp3Default.Interpolate("G01", []float64{holeMidpointJ2000S}); err == nil {
		t.Fatal("default factor unexpectedly interpolated hole midpoint")
	}

	// 3. Wide factor (13.0) serves hole midpoint with finite position.
	sp3Wide, err := LoadSP3(fixture, WithGapThresholdFactor(13.0))
	if err != nil {
		t.Fatalf("LoadSP3 with factor 13.0: %v", err)
	}
	t.Cleanup(func() { _ = sp3Wide.Close() })

	wideFactor, err := sp3Wide.GapThresholdFactor()
	if err != nil {
		t.Fatalf("sp3Wide.GapThresholdFactor: %v", err)
	}
	if math.Abs(wideFactor-13.0) > 1e-9 {
		t.Fatalf("expected factor 13.0, got %f", wideFactor)
	}

	pos, clk, hasClk, err := sp3Wide.ObservableState("G01", holeMidpointJ2000S)
	if err != nil {
		t.Fatalf("wide factor failed to serve hole midpoint: %v", err)
	}
	if math.IsNaN(pos[0]) || math.IsInf(pos[0], 0) || pos[0] == 0 {
		t.Fatalf("expected finite non-zero position X, got %f", pos[0])
	}
	_ = clk
	_ = hasClk

	interpolatedPos, interpolatedClk, written, err := sp3Wide.Interpolate("G01", []float64{holeMidpointJ2000S})
	if err != nil {
		t.Fatalf("sp3Wide.Interpolate: %v", err)
	}
	if written != 1 || len(interpolatedPos) != 1 || math.IsNaN(interpolatedPos[0][0]) || interpolatedPos[0][0] == 0 {
		t.Fatalf("expected finite position from interpolate, got %v", interpolatedPos)
	}
	_ = interpolatedClk

	// 4. Physical continuity and residual check options.
	// Factor 1.0 must fail.
	if _, err := sp3Default.ContinuityWithOptions(1, 1.0, SP3ContinuityOptions{GapThresholdFactor: 1.0}); err == nil {
		t.Fatal("ContinuityWithOptions with factor 1.0 unexpectedly succeeded")
	}
	if _, err := sp3Default.CheckContinuityWithOptions(1, 1.0, SP3ContinuityOptions{GapThresholdFactor: 1.0}); err == nil {
		t.Fatal("CheckContinuityWithOptions with factor 1.0 unexpectedly succeeded")
	}

	// Default continuity vs factor 13.0 continuity:
	contDef, err := sp3Default.Continuity(1, 1.0)
	if err != nil {
		t.Fatalf("sp3Default.Continuity default: %v", err)
	}
	contWide, err := sp3Default.ContinuityWithOptions(1, 1.0, SP3ContinuityOptions{GapThresholdFactor: 13.0})
	if err != nil {
		t.Fatalf("sp3Default.ContinuityWithOptions factor 13.0: %v", err)
	}
	if contWide.Defects >= contDef.Defects {
		t.Fatalf("expected factor 13.0 defects (%d) < default defects (%d)", contWide.Defects, contDef.Defects)
	}
	// Check that functional option variant behaves identically.
	contWideFunc, err := sp3Default.Continuity(1, 1.0, WithGapThresholdFactor(13.0))
	if err != nil {
		t.Fatalf("sp3Default.Continuity functional opt 13.0: %v", err)
	}
	if contWideFunc.Defects != contWide.Defects {
		t.Fatalf("defects mismatch between functional and struct options: %d != %d", contWideFunc.Defects, contWide.Defects)
	}

	// 5. Continuity verdict JSON with options.
	if _, err := sp3Default.ContinuityVerdictJSONWithOptions(1, 1.0, holeMidpointJ2000S-100, holeMidpointJ2000S+100, SP3ContinuityOptions{GapThresholdFactor: 1.0}); err == nil {
		t.Fatal("ContinuityVerdictJSONWithOptions with factor 1.0 unexpectedly succeeded")
	}
	jsonBytes, err := sp3Default.ContinuityVerdictJSONWithOptions(1, 1.0, holeMidpointJ2000S-100, holeMidpointJ2000S+100, SP3ContinuityOptions{GapThresholdFactor: 13.0})
	if err != nil {
		t.Fatalf("ContinuityVerdictJSONWithOptions factor 13.0: %v", err)
	}
	var verdictParsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &verdictParsed); err != nil {
		t.Fatalf("invalid json from verdict: %v (raw: %s)", err, string(jsonBytes))
	}
	// Also test functional option form.
	jsonBytesFunc, err := sp3Default.ContinuityVerdictJSON(1, 1.0, holeMidpointJ2000S-100, holeMidpointJ2000S+100, WithGapThresholdFactor(13.0))
	if err != nil {
		t.Fatalf("ContinuityVerdictJSON functional opt 13.0: %v", err)
	}
	if len(jsonBytesFunc) == 0 {
		t.Fatal("empty json bytes from functional ContinuityVerdictJSON")
	}

	// 6. PreciseEphemerisSamples with options and GapThresholdFactor getter.
	canonicalSamples, err := sp3Default.PreciseSamples()
	if err != nil {
		t.Fatalf("sp3Default.PreciseSamples: %v", err)
	}
	if len(canonicalSamples) == 0 {
		t.Fatal("no canonical samples extracted")
	}

	if _, err := BuildPreciseEphemerisSamplesWithOptions(canonicalSamples, SP3InterpolationOptions{GapThresholdFactor: 1.0}); err == nil {
		t.Fatal("BuildPreciseEphemerisSamplesWithOptions with factor 1.0 unexpectedly succeeded")
	}

	samplesDef, err := BuildPreciseEphemerisSamples(canonicalSamples)
	if err != nil {
		t.Fatalf("BuildPreciseEphemerisSamples default: %v", err)
	}
	t.Cleanup(func() { _ = samplesDef.Close() })
	sDefFactor, err := samplesDef.GapThresholdFactor()
	if err != nil {
		t.Fatalf("samplesDef.GapThresholdFactor: %v", err)
	}
	if math.Abs(sDefFactor-1.5) > 1e-9 {
		t.Fatalf("expected samples default factor 1.5, got %f", sDefFactor)
	}
	// At hole midpoint, default samples query should report Gap.
	statesDef, err := samplesDef.ObservableStatesShared([]string{"G01"}, holeMidpointJ2000S)
	if err != nil {
		t.Fatalf("samplesDef.ObservableStatesShared: %v", err)
	}
	if len(statesDef) != 1 || statesDef[0].ElementStatus != ObservableStateGap {
		t.Fatalf("expected ObservableStateGap for default samples, got status %v", statesDef[0].ElementStatus)
	}

	// Build samples with factor 13.0.
	samplesWide, err := BuildPreciseEphemerisSamplesWithOptions(canonicalSamples, SP3InterpolationOptions{GapThresholdFactor: 13.0})
	if err != nil {
		t.Fatalf("BuildPreciseEphemerisSamplesWithOptions factor 13.0: %v", err)
	}
	t.Cleanup(func() { _ = samplesWide.Close() })
	sWideFactor, err := samplesWide.GapThresholdFactor()
	if err != nil {
		t.Fatalf("samplesWide.GapThresholdFactor: %v", err)
	}
	if math.Abs(sWideFactor-13.0) > 1e-9 {
		t.Fatalf("expected samples factor 13.0, got %f", sWideFactor)
	}
	// At hole midpoint, wide samples query should report Valid.
	statesWide, err := samplesWide.ObservableStatesShared([]string{"G01"}, holeMidpointJ2000S)
	if err != nil {
		t.Fatalf("samplesWide.ObservableStatesShared: %v", err)
	}
	if len(statesWide) != 1 || statesWide[0].ElementStatus != ObservableStateValid || math.IsNaN(statesWide[0].PositionECEFM[0]) {
		t.Fatalf("expected ObservableStateValid for wide samples, got status %v", statesWide[0].ElementStatus)
	}

	// 7. PreciseEphemerisInterpolant with options and GapThresholdFactor getter.
	if _, err := BuildPreciseEphemerisInterpolantWithOptions(canonicalSamples, SP3InterpolationOptions{GapThresholdFactor: 1.0}); err == nil {
		t.Fatal("BuildPreciseEphemerisInterpolantWithOptions factor 1.0 unexpectedly succeeded")
	}

	interpDef, err := BuildPreciseEphemerisInterpolant(canonicalSamples)
	if err != nil {
		t.Fatalf("BuildPreciseEphemerisInterpolant default: %v", err)
	}
	t.Cleanup(func() { _ = interpDef.Close() })
	iDefFactor, err := interpDef.GapThresholdFactor()
	if err != nil {
		t.Fatalf("interpDef.GapThresholdFactor: %v", err)
	}
	if math.Abs(iDefFactor-1.5) > 1e-9 {
		t.Fatalf("expected interpolant default factor 1.5, got %f", iDefFactor)
	}
	interpStatesDef, err := interpDef.ObservableStates([]string{"G01"}, []float64{holeMidpointJ2000S})
	if err != nil {
		t.Fatalf("interpDef.ObservableStates: %v", err)
	}
	if len(interpStatesDef) != 1 || interpStatesDef[0].ElementStatus != ObservableStateGap {
		t.Fatalf("expected ObservableStateGap for default interpolant, got status %v", interpStatesDef[0].ElementStatus)
	}

	interpWide, err := BuildPreciseEphemerisInterpolantWithOptions(canonicalSamples, SP3InterpolationOptions{GapThresholdFactor: 13.0})
	if err != nil {
		t.Fatalf("BuildPreciseEphemerisInterpolantWithOptions factor 13.0: %v", err)
	}
	t.Cleanup(func() { _ = interpWide.Close() })
	iWideFactor, err := interpWide.GapThresholdFactor()
	if err != nil {
		t.Fatalf("interpWide.GapThresholdFactor: %v", err)
	}
	if math.Abs(iWideFactor-13.0) > 1e-9 {
		t.Fatalf("expected interpolant factor 13.0, got %f", iWideFactor)
	}
	interpStatesWide, err := interpWide.ObservableStates([]string{"G01"}, []float64{holeMidpointJ2000S})
	if err != nil {
		t.Fatalf("interpWide.ObservableStates: %v", err)
	}
	if len(interpStatesWide) != 1 || interpStatesWide[0].ElementStatus != ObservableStateValid || math.IsNaN(interpStatesWide[0].PositionECEFM[0]) {
		t.Fatalf("expected ObservableStateValid for wide interpolant, got status %v", interpStatesWide[0].ElementStatus)
	}

	// 8. PreciseInterpolantArtifact GapThresholdFactor getter.
	artWideBytes, _, err := sp3Wide.ArtifactBytes()
	if err != nil {
		t.Fatalf("sp3Wide.ArtifactBytes: %v", err)
	}
	artWide, _, err := OpenPreciseInterpolantArtifact(artWideBytes)
	if err != nil {
		t.Fatalf("OpenPreciseInterpolantArtifact wide: %v", err)
	}
	t.Cleanup(func() { _ = artWide.Close() })
	artWideFactor, err := artWide.GapThresholdFactor()
	if err != nil {
		t.Fatalf("artWide.GapThresholdFactor: %v", err)
	}
	if math.Abs(artWideFactor-13.0) > 1e-9 {
		t.Fatalf("expected artifact factor 13.0, got %f", artWideFactor)
	}

	artDefBytes, _, err := sp3Default.ArtifactBytes()
	if err != nil {
		t.Fatalf("sp3Default.ArtifactBytes: %v", err)
	}
	artDef, _, err := OpenPreciseInterpolantArtifact(artDefBytes)
	if err != nil {
		t.Fatalf("OpenPreciseInterpolantArtifact default: %v", err)
	}
	t.Cleanup(func() { _ = artDef.Close() })
	artDefFactor, err := artDef.GapThresholdFactor()
	if err != nil {
		t.Fatalf("artDef.GapThresholdFactor: %v", err)
	}
	if math.Abs(artDefFactor-1.5) > 1e-9 {
		t.Fatalf("expected default artifact factor 1.5, got %f", artDefFactor)
	}

	// 9. LoadExactSP3WithOptions with invalid factor vs valid factor.
	request, err := NewExactSP3Request(2020, 6, 24, "", "01D", "15M", "GRGS")
	if err != nil {
		t.Fatalf("NewExactSP3Request: %v", err)
	}
	t.Cleanup(func() { _ = request.Close() })

	if _, _, err := LoadExactSP3WithOptions(fixture, request, SP3LoadOptions{GapThresholdFactor: 1.0}); err == nil {
		t.Fatal("LoadExactSP3WithOptions factor 1.0 unexpectedly succeeded")
	} else {
		var statusErr *StatusError
		if !errors.As(err, &statusErr) || statusErr.Code != StatusInvalidArgument {
			t.Fatalf("expected StatusInvalidArgument for factor 1.0, got %v", err)
		}
	}

	// With factor 13.0, factor check passes and exact validation is evaluated.
	// Since the fixture contains a gap, exact validation correctly detects the gap/sequence mismatch.
	if _, _, err := LoadExactSP3WithOptions(fixture, request, SP3LoadOptions{GapThresholdFactor: 13.0}); err == nil {
		t.Fatal("LoadExactSP3WithOptions on gapped fixture unexpectedly passed exact validation")
	} else {
		var statusErr *StatusError
		if errors.As(err, &statusErr) && statusErr.Code == StatusInvalidArgument && statusErr.Detail == "invalid input: gap_threshold_factor must be finite and greater than 1.0" {
			t.Fatalf("factor 13.0 unexpectedly failed gap threshold factor validation: %v", err)
		}
	}

	// Functional option form of LoadExactSP3.
	if _, _, err := LoadExactSP3(fixture, request, WithGapThresholdFactor(1.0)); err == nil {
		t.Fatal("LoadExactSP3 with functional opt factor 1.0 unexpectedly succeeded")
	}

	// 10. Nil and closed handle safety for all new getters.
	if _, err := (*SP3)(nil).GapThresholdFactor(); !errors.Is(err, ErrClosed) {
		t.Fatalf("expected ErrClosed on nil SP3, got %v", err)
	}
	if _, err := (*PreciseEphemerisSamples)(nil).GapThresholdFactor(); !errors.Is(err, ErrClosed) {
		t.Fatalf("expected ErrClosed on nil PreciseEphemerisSamples, got %v", err)
	}
	if _, err := (*PreciseEphemerisInterpolant)(nil).GapThresholdFactor(); !errors.Is(err, ErrClosed) {
		t.Fatalf("expected ErrClosed on nil PreciseEphemerisInterpolant, got %v", err)
	}
	if _, err := (*PreciseInterpolantArtifact)(nil).GapThresholdFactor(); !errors.Is(err, ErrClosed) {
		t.Fatalf("expected ErrClosed on nil PreciseInterpolantArtifact, got %v", err)
	}

	sp3Temp, err := LoadSP3(fixture)
	if err != nil {
		t.Fatal(err)
	}
	if err := sp3Temp.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := sp3Temp.GapThresholdFactor(); !errors.Is(err, ErrClosed) {
		t.Fatalf("expected ErrClosed on closed SP3, got %v", err)
	}
}
