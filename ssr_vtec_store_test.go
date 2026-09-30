package sidereon

import (
	"math"
	"testing"
)

func TestSSRStoreVTECEvaluationAndAgePolicy(t *testing.T) {
	messages, err := BuildRTCMSSRVTEC(RTCMSSRVTECInfo{
		MessageNumber: 4076, HasIGSSSRVersion: true, IGSSSRVersion: 1,
		EpochTimeS: 100000, UpdateInterval: 3, IODSSR: 4, ProviderID: 7,
		SolutionID: 2, QualityIndicator: 123,
		Layers: []RTCMTecLayer{{Height: 45, Degree: 1, Order: 1, Cosine: []int16{100, 200, 0}, Sine: []int16{0}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, messages)
	store, err := NewSSRCorrectionStore(SSRReferencePointAntennaPhaseCenter)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, store)
	week := GNSSWeekTow{System: GPST, Week: 2400, TOWSeconds: 100000}
	if err := store.Ingest(messages, week); err != nil {
		t.Fatal(err)
	}
	got, layers, err := store.EvaluateVTEC([3]float64{0, 0, 6370000}, [3]float64{0, 0, 26000000}, week, 1e9)
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != SSRVTECQueryEvaluated || got.LayerCount != 1 || len(layers) != 1 || got.AgeS != 0 {
		t.Fatalf("SSR VTEC query=%+v layers=%+v", got, layers)
	}
	wantSTEC := 0.5 + math.Sqrt(3)
	if math.Abs(got.STECTECU-wantSTEC) > 1e-12 || math.Abs(layers[0].STECTECU-wantSTEC) > 1e-12 {
		t.Fatalf("SSR VTEC STEC query=%g layer=%g, want %g", got.STECTECU, layers[0].STECTECU, wantSTEC)
	}
	if err := store.SetVTECMaxAge(5); err != nil {
		t.Fatal(err)
	}
	week.TOWSeconds += 6
	got, layers, err = store.EvaluateVTEC([3]float64{0, 0, 6370000}, [3]float64{0, 0, 26000000}, week, 1e9)
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != SSRVTECQueryStale || got.AgeS != 6 || got.MaxAgeS != 5 || len(layers) != 0 {
		t.Fatalf("stale SSR VTEC query=%+v layers=%+v", got, layers)
	}
}

func TestSSRStoreRTCMReadingRetainsResyncAndVTEC(t *testing.T) {
	messages, err := BuildRTCMSSRVTEC(RTCMSSRVTECInfo{
		MessageNumber: 4076, HasIGSSSRVersion: true, IGSSSRVersion: 1,
		EpochTimeS: 100000, UpdateInterval: 3, IODSSR: 4, ProviderID: 7,
		SolutionID: 2, QualityIndicator: 123,
		Layers: []RTCMTecLayer{{Height: 45, Degree: 1, Order: 1, Cosine: []int16{100, 200, 0}, Sine: []int16{0}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	frame, err := messages.Frame(0)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, messages)

	stream := append([]byte{0x00, 0x00, 0x00}, frame...)
	stream = append(stream, frame[:3]...)
	week := GNSSWeekTow{System: GPST, Week: 2400, TOWSeconds: 100000}
	store, diagnostics, trailing, refusals, err := NewSSRCorrectionStoreFromRTCMReading(stream, week)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, store)
	closeAfterTest(t, diagnostics)
	closeAfterTest(t, refusals)
	if trailing != 3 {
		t.Fatalf("trailing partial bytes = %d, want 3", trailing)
	}
	resync, err := diagnostics.ResyncBytes()
	if err != nil || resync != 6 {
		t.Fatalf("stream resync bytes = %d, %v; want 6 (3 leading noise + 3-byte partial prefix)", resync, err)
	}
	refusalCount, err := refusals.Count()
	if err != nil || refusalCount != 0 {
		t.Fatalf("SSR ingest refusals = %d, %v; want 0", refusalCount, err)
	}
	query, layers, err := store.EvaluateVTEC([3]float64{0, 0, 6370000}, [3]float64{0, 0, 26000000}, week, 1e9)
	if err != nil {
		t.Fatal(err)
	}
	if query.Kind != SSRVTECQueryEvaluated || query.LayerCount != 1 || len(layers) != 1 || math.Abs(query.STECTECU-(0.5+math.Sqrt(3))) > 1e-12 {
		t.Fatalf("lenient stream SSR VTEC result = %+v layers=%+v", query, layers)
	}
}
