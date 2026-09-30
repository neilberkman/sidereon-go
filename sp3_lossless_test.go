package sidereon

import (
	"encoding/json"
	"errors"
	"math"
	"testing"

	"sidereon.dev/go/v3/internal/native"
)

func TestLosslessPreciseSampleEpochConversionPreservesTaggedInstant(t *testing.T) {
	want := ClockEpoch{
		Scale: GPST, Representation: RINEXClockInstantNanos,
		NanosHigh: -1, NanosLow: 0xfedcba9876543210,
	}
	got := publicClockEpoch(nativeClockEpoch(want))
	if got != want {
		t.Fatalf("clock epoch round trip = %#v, want %#v", got, want)
	}

	sample := PreciseEphemerisSampleV2{
		Satellite: "G01", Epoch: want, PositionECEFM: [3]float64{1, 2, 3},
		HasClock: true, ClockS: 4e-6, ClockEvent: true,
	}
	nativeSample := nativePreciseSampleV2(sample)
	if converted := publicPreciseSampleV2(nativeSample); converted != sample {
		t.Fatalf("precise sample V2 round trip = %#v, want %#v", converted, sample)
	}

	accuracy := PreciseEphemerisAccuracySampleV2{
		Satellite: "G01", Epoch: want,
		PositionVarianceM2: [3]SP3AccuracyValue{{Kind: SP3AccuracyKnown, Value: 1}, {Kind: SP3AccuracyUnknown}, {Kind: SP3AccuracyTooLarge}},
		ClockVarianceM2:    SP3AccuracyValue{Kind: SP3AccuracyKnown, Value: 4},
	}
	if converted := publicPreciseAccuracySampleV2(nativePreciseAccuracySampleV2(accuracy)); converted != accuracy {
		t.Fatalf("precise accuracy V2 round trip = %#v, want %#v", converted, accuracy)
	}
}

func TestNativeLosslessSampleEpochConversionPreservesAllWords(t *testing.T) {
	value := native.NativeClockEpoch{
		Scale: uint32(BDT), Representation: uint32(RINEXClockInstantNanos),
		NanosHigh: -0x123456789, NanosLow: 0x0123456789abcdef,
	}
	got := nativeClockEpoch(publicClockEpoch(value))
	if got != value {
		t.Fatalf("native clock epoch round trip = %#v, want %#v", got, value)
	}
}

func TestLosslessPreciseSamplesV2BridgePreservesEpochsAndQueries(t *testing.T) {
	baseNanos := uint64(1) << 63
	const intervalNanos = uint64(30_000_000_000)
	if float64(baseNanos)/1e9 != float64(baseNanos+1)/1e9 {
		t.Fatal("chosen adjacent tagged epochs do not collide after legacy float conversion")
	}
	collisionSamples := []PreciseEphemerisSampleV2{
		{Satellite: "G02", Epoch: ClockEpoch{Scale: GPST, Representation: RINEXClockInstantNanos, NanosLow: baseNanos}, PositionECEFM: [3]float64{1, 2, 3}},
		{Satellite: "G02", Epoch: ClockEpoch{Scale: GPST, Representation: RINEXClockInstantNanos, NanosLow: baseNanos + 1}, PositionECEFM: [3]float64{4, 5, 6}},
	}
	_, err := BuildPreciseEphemerisSamplesV2(collisionSamples)
	var statusErr *StatusError
	if !errors.As(err, &statusErr) || statusErr.Code != StatusInvalidArgument {
		t.Fatalf("same-whole-second V2 nodes = %T %v, want InvalidArgument with captured precise-samples details", err, err)
	}
	engine := statusErr.EngineError()
	if engine == nil || engine.Family != EngineErrorFamilyPreciseSamples || engine.FamilyName != "precise_samples" || engine.Operation != "sidereon_precise_ephemeris_samples_from_samples_v2" || engine.Kind != "non_monotonic_epochs" || len(engine.Payload) == 0 || len(engine.Fields) == 0 || engine.CaptureError != nil {
		t.Fatalf("same-whole-second V2 engine error lost its complete producer record: %#v", engine)
	}
	satellite, ok := engine.TypedFields["satellite"]
	if !ok || satellite.Kind != EngineJSONObject || satellite.Object["system"].Kind != EngineJSONString || satellite.Object["system"].String != "gps" || satellite.Object["prn"].Kind != EngineJSONNumber || satellite.Object["prn"].Number.String() != "2" {
		t.Fatalf("same-whole-second V2 typed satellite cause = %#v, want gps PRN 2 with its exact integer token", satellite)
	}
	var rawSatellite struct {
		System string `json:"system"`
		PRN    int    `json:"prn"`
	}
	if err := json.Unmarshal(engine.Fields, &struct {
		Satellite *struct {
			System string `json:"system"`
			PRN    int    `json:"prn"`
		} `json:"satellite"`
	}{Satellite: &rawSatellite}); err != nil || rawSatellite.System != "gps" || rawSatellite.PRN != 2 {
		t.Fatalf("same-whole-second V2 raw satellite fields = %+v, decode error %v", rawSatellite, err)
	}

	samples := make([]PreciseEphemerisSampleV2, 11)
	accuracy := make([]PreciseEphemerisAccuracySampleV2, len(samples))
	for index := range samples {
		offsetNanos := int64(index-5) * int64(intervalNanos)
		epochNanos := baseNanos + uint64(offsetNanos)
		epoch := ClockEpoch{Scale: GPST, Representation: RINEXClockInstantNanos, NanosLow: epochNanos}
		samples[index] = PreciseEphemerisSampleV2{
			Satellite: "G01", Epoch: epoch,
			PositionECEFM: [3]float64{20_000_000, 1_000_000, 2_000_000},
			HasClock:      true, ClockS: 1e-6,
		}
		accuracy[index] = PreciseEphemerisAccuracySampleV2{
			Satellite: "G01", Epoch: epoch,
			PositionVarianceM2: [3]SP3AccuracyValue{{Kind: SP3AccuracyKnown, Value: 1}, {Kind: SP3AccuracyKnown, Value: 4}, {Kind: SP3AccuracyKnown, Value: 9}},
			ClockVarianceM2:    SP3AccuracyValue{Kind: SP3AccuracyKnown, Value: 16},
		}
	}

	sampleSet, err := BuildPreciseEphemerisSamplesWithAccuracyV2(samples, accuracy)
	if err != nil {
		t.Fatalf("build V2 samples: %v", err)
	}
	defer sampleSet.Close()
	gotSamples, err := sampleSet.RecordsV2()
	if err != nil {
		t.Fatalf("read V2 sample records: %v", err)
	}
	if len(gotSamples) != len(samples) {
		t.Fatalf("V2 sample count = %d, want %d", len(gotSamples), len(samples))
	}
	for index := range samples {
		if gotSamples[index] != samples[index] {
			t.Fatalf("V2 sample[%d] = %#v, want %#v", index, gotSamples[index], samples[index])
		}
	}
	gotAccuracy, err := sampleSet.AccuracyRecordsV2()
	if err != nil {
		t.Fatalf("read V2 accuracy records: %v", err)
	}
	if len(gotAccuracy) != len(accuracy) {
		t.Fatalf("V2 accuracy count = %d, want %d", len(gotAccuracy), len(accuracy))
	}
	for index := range accuracy {
		if gotAccuracy[index] != accuracy[index] {
			t.Fatalf("V2 accuracy[%d] = %#v, want %#v", index, gotAccuracy[index], accuracy[index])
		}
	}

	interpolant, err := BuildPreciseEphemerisInterpolantWithAccuracyV2(samples, accuracy)
	if err != nil {
		t.Fatalf("build V2 interpolant: %v", err)
	}
	defer interpolant.Close()
	seconds := int64(baseNanos / 1_000_000_000)
	attoseconds := (baseNanos % 1_000_000_000) * 1_000_000_000
	epoch, err := NewExactEpoch(seconds, attoseconds)
	if err != nil {
		t.Fatalf("construct exact tagged sample epoch: %v", err)
	}
	defer epoch.Close()
	query, err := epoch.Query()
	if err != nil {
		t.Fatalf("query exact tagged sample epoch: %v", err)
	}
	defer query.Close()
	state, err := interpolant.SourceStateAtEpochQueries(query, query, "G01")
	if err != nil {
		t.Fatalf("evaluate V2 interpolant at exact epoch: %v", err)
	}
	if !state.HasState {
		t.Fatalf("exact V2 source state = %+v", state)
	}
	wantCenter := [3]float64{20_000_000, 1_000_000, 2_000_000}
	for index, want := range wantCenter {
		if math.IsNaN(state.PositionECEFM[index]) || math.IsInf(state.PositionECEFM[index], 0) || math.Abs(state.PositionECEFM[index]-want) > 1e-6 {
			t.Fatalf("linear V2 center position[%d] = %.17g, want %.17g within 1e-6 m", index, state.PositionECEFM[index], want)
		}
	}
	if math.IsNaN(state.ClockS) || math.IsInf(state.ClockS, 0) || math.Abs(state.ClockS-1e-6) > 1e-12 {
		t.Fatalf("linear V2 center clock = %.17g, want 1e-6 s", state.ClockS)
	}
	for _, component := range state.PositionECEFM {
		if math.IsNaN(component) || math.IsInf(component, 0) {
			t.Fatalf("exact V2 source state has non-finite position: %+v", state.PositionECEFM)
		}
	}
	variance, err := interpolant.EphemerisVarianceAtEpochQueries(query, query, "G01")
	if err != nil {
		t.Fatalf("evaluate V2 variance at exact epoch: %v", err)
	}
	if variance < 0 || math.IsNaN(variance) || math.IsInf(variance, 0) {
		t.Fatalf("exact V2 variance = %g, want finite and nonnegative", variance)
	}
}
