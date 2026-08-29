package sidereon

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"math"
	"os"
	"sync"
	"testing"
)

func assertConcurrentClose(t *testing.T, read func() error, closeHandle func() error) {
	t.Helper()
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			if err := read(); err != nil && !errors.Is(err, ErrClosed) {
				t.Errorf("concurrent read: %v", err)
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		if err := closeHandle(); err != nil {
			t.Errorf("concurrent close: %v", err)
		}
	}()
	close(start)
	wg.Wait()
	if err := closeHandle(); err != nil {
		t.Fatalf("repeated close: %v", err)
	}
	if err := read(); !errors.Is(err, ErrClosed) {
		t.Fatalf("post-close read = %v", err)
	}
}

func observationFixture(t *testing.T) []byte {
	t.Helper()
	b, e := os.ReadFile("testdata/obs/ESBC00DNK_R_20201770000_01D_30S_MO_trim.rnx")
	if e != nil {
		t.Fatal(e)
	}
	return b
}

func readObservationFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/obs/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestCommittedObservationFixtureHashesAndInventory(t *testing.T) {
	cases := []struct {
		name, hash, marker                          string
		version                                     float64
		epochs, codes, values, carrier, pseudorange int
		firstSatelliteCount                         int
		firstSatellite, firstCode                   string
		firstKind                                   RINEXObservationKind
		firstValue, firstPseudorange, firstCarrierM float64
		firstLLI, firstSSI                          int32
	}{
		{"ESBC00DNK_R_20201770000_01D_30S_MO_trim.rnx", "c1fbc120be90d7498b3ff138f28ceb7865ce050ddd9cd2f55ce413e553f2e7e0", "ESBC00DNK", 3.05, 2, 90, 720, 174, 39, 43, "G02", "C1C", RINEXObservationPseudorange, 25847357.745, 25847357.745, 0, -1, 3},
		{"ESBC00DNK_R_20201770000_01D_30S_MO_trim.crx", "73b2294711f317c20a043c290c5d590917b037caac8feb21d74dc7600f55f5c2", "ESBC00DNK", 3.05, 2, 90, 720, 174, 39, 43, "G02", "C1C", RINEXObservationPseudorange, 25847357.745, 25847357.745, 0, -1, 3},
		{"algo0010_2015001_v1_trim.rnx", "f2eae58b37fa267b6f64549de8eb1504473057b46fba1d039b9d2f063b536f22", "ALGO CACS-GSD 883160 Algonquin Park ON Canada", 2.11, 2, 16, 160, 40, 20, 20, "G01", "L1C", RINEXObservationCarrierPhase, 132905208.937, 25291008.79, 25291020.342655797, 4, 5},
		{"algo0010_2015001_v1_trim.crx", "acc0d16347d28fb5911798f792046b1d32b8177a73b8b8fb4e521fa1fcf0af38", "ALGO CACS-GSD 883160 Algonquin Park ON Canada", 2.11, 2, 16, 160, 40, 20, 20, "G01", "L1C", RINEXObservationCarrierPhase, 132905208.937, 25291008.79, 25291020.342655797, 4, 5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := readObservationFixture(t, tc.name)
			gotHash := fmt.Sprintf("%x", sha256.Sum256(data))
			if gotHash != tc.hash {
				t.Fatalf("hash=%s, want %s", gotHash, tc.hash)
			}
			if len(tc.name) > 4 && tc.name[len(tc.name)-4:] == ".crx" {
				var err error
				data, err = DecodeCRINEX(data)
				if err != nil {
					t.Fatal(err)
				}
			}
			obs, err := ParseRINEXObservation(data)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := obs.Close(); err != nil {
					t.Error(err)
				}
			}()
			header, err := obs.Header()
			if err != nil {
				t.Fatal(err)
			}
			if header.Version != tc.version || !header.HasMarkerName || header.MarkerName != tc.marker || !header.HasInterval || header.IntervalS != 30 {
				t.Fatalf("header version/marker=(%v,%q), want (%v,%q)", header.Version, header.MarkerName, tc.version, tc.marker)
			}
			codes, err := obs.Codes()
			if err != nil || len(codes) == 0 {
				t.Fatalf("codes=%d err=%v", len(codes), err)
			}
			epochs, err := obs.Epochs()
			if err != nil || len(epochs) == 0 {
				t.Fatalf("epochs=%d err=%v", len(epochs), err)
			}
			values, err := obs.Values(0)
			if err != nil || len(values) == 0 {
				t.Fatalf("values=%d err=%v", len(values), err)
			}
			carrier, err := obs.CarrierPhase(0)
			if err != nil || len(carrier) == 0 {
				t.Fatalf("carrier=%d err=%v", len(carrier), err)
			}
			pseudoranges, err := obs.Pseudoranges(0)
			if err != nil || len(pseudoranges) == 0 {
				t.Fatalf("pseudoranges=%d err=%v", len(pseudoranges), err)
			}
			if len(epochs) != tc.epochs || len(values) != tc.values || len(carrier) != tc.carrier || len(pseudoranges) != tc.pseudorange || len(codes) != tc.codes || header.ObservationCodeCount != tc.codes || epochs[0].SatelliteCount != tc.firstSatelliteCount {
				t.Fatalf("counts epochs=%d codes=%d values=%d carrier=%d pseudoranges=%d", len(epochs), len(codes), len(values), len(carrier), len(pseudoranges))
			}
			if values[0].SatelliteID != tc.firstSatellite || values[0].Code != tc.firstCode || values[0].Kind != tc.firstKind || values[0].Value != tc.firstValue || values[0].LLI != tc.firstLLI || values[0].SSI != tc.firstSSI {
				t.Fatalf("first observation=%+v", values[0])
			}
			if pseudoranges[0].PseudorangeM != tc.firstPseudorange || carrier[0].ValueM != tc.firstCarrierM {
				t.Fatalf("first pseudorange=%+v carrier=%+v", pseudoranges[0], carrier[0])
			}
			clock, err := obs.ReceiverClockPhaseDeviations()
			if err != nil {
				t.Fatal(err)
			}
			if len(clock) != tc.epochs || clock[0].HasPhaseS || clock[1].HasPhaseS {
				t.Fatalf("clock phase=%+v, want two absent samples", clock)
			}
		})
	}
}

func TestRINEXObservationFixtureAndCRINEX(t *testing.T) {
	data := observationFixture(t)
	obs, e := ParseRINEXObservation(data)
	if e != nil {
		t.Fatal(e)
	}
	defer func() {
		if e := obs.Close(); e != nil {
			t.Fatal(e)
		}
	}()
	h, e := obs.Header()
	if e != nil {
		t.Fatal(e)
	}
	if h.Version < 3 || !h.HasMarkerName {
		t.Fatalf("header=%+v", h)
	}
	epochs, e := obs.Epochs()
	if e != nil || len(epochs) == 0 {
		t.Fatalf("epochs=%d err=%v", len(epochs), e)
	}
	codes, e := obs.Codes()
	if e != nil || len(codes) == 0 {
		t.Fatalf("codes=%d err=%v", len(codes), e)
	}
	values, e := obs.Values(0)
	if e != nil || len(values) == 0 {
		t.Fatalf("values=%d err=%v", len(values), e)
	}
	if values[0].SatelliteID != "G02" || values[0].Code != "C1C" || !values[0].HasValue {
		t.Fatalf("first observation=%+v", values[0])
	}
	if values[0].Value != 25847357.745 {
		t.Fatalf("first observation value=%.12g, want frozen fixture value", values[0].Value)
	}
	text, e := obs.RINEXText()
	if e != nil || len(text) == 0 {
		t.Fatalf("text=%d err=%v", len(text), e)
	}
	roundTrip, e := ParseRINEXObservation(append([]byte(nil), text...))
	if e != nil {
		t.Fatal(e)
	}
	if e = roundTrip.Close(); e != nil {
		t.Fatal(e)
	}
	crx, e := EncodeCRINEX(data)
	if e != nil || len(crx) == 0 {
		t.Fatalf("crx=%d err=%v", len(crx), e)
	}
	rnx, e := DecodeCRINEX(crx)
	if e != nil || len(rnx) == 0 {
		t.Fatalf("rnx=%d err=%v", len(rnx), e)
	}
	decoded, e := ParseRINEXObservation(rnx)
	if e != nil {
		t.Fatal(e)
	}
	if e = decoded.Close(); e != nil {
		t.Fatal(e)
	}
	crxFixture := readObservationFixture(t, "ESBC00DNK_R_20201770000_01D_30S_MO_trim.crx")
	decodedFixture, e := DecodeCRINEX(crxFixture)
	if e != nil || len(decodedFixture) == 0 {
		t.Fatalf("decode committed CRINEX fixture: bytes=%d err=%v", len(decodedFixture), e)
	}
	decodedFixtureObs, e := ParseRINEXObservation(decodedFixture)
	if e != nil {
		t.Fatal(e)
	}
	if e = decodedFixtureObs.Close(); e != nil {
		t.Fatal(e)
	}
}

func TestRINEXObservationValidationAndConcurrentClose(t *testing.T) {
	obs, e := ParseRINEXObservation(observationFixture(t))
	if e != nil {
		t.Fatal(e)
	}
	if _, e = obs.Values(-1); e == nil {
		t.Fatal("negative epoch accepted")
	}
	if _, _, _, _, e = obs.Observation(0, "G01\x00", "C1C"); e == nil {
		t.Fatal("embedded NUL accepted")
	}
	if e = obs.Close(); e != nil {
		t.Fatal(e)
	}
	if e = obs.Close(); e != nil {
		t.Fatal(e)
	}
	if _, e = obs.Version(); e == nil {
		t.Fatal("use after close accepted")
	}
	obs, e = ParseRINEXObservation(observationFixture(t))
	if e != nil {
		t.Fatal(e)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, _ = obs.EpochCount() }()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := obs.Close(); err != nil {
			t.Error(err)
		}
	}()
	wg.Wait()
}

func TestRINEXTextIsIndependent(t *testing.T) {
	data := observationFixture(t)
	obs, e := ParseRINEXObservation(data)
	if e != nil {
		t.Fatal(e)
	}
	data[0] = 'x'
	if version, err := obs.Version(); err != nil || version != 3.05 {
		t.Fatalf("input mutation changed parsed handle: version=%v err=%v", version, err)
	}
	defer func() {
		if err := obs.Close(); err != nil {
			t.Error(err)
		}
	}()
	a, e := obs.RINEXText()
	if e != nil {
		t.Fatal(e)
	}
	b, e := obs.RINEXText()
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(a, b) {
		t.Fatal("copied text differs")
	}
	a[0] = 'x'
	if bytes.Equal(a, b) {
		t.Fatal("text aliases")
	}
	valuesA, err := obs.Values(0)
	if err != nil {
		t.Fatal(err)
	}
	valuesB, err := obs.Values(0)
	if err != nil {
		t.Fatal(err)
	}
	valuesA[0].Value++
	if valuesA[0].Value == valuesB[0].Value {
		t.Fatal("observation values alias")
	}
}

func TestOwningHandlesCloseAndReadRace(t *testing.T) {
	obs, err := ParseRINEXObservation(observationFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	assertConcurrentClose(t, func() error { _, err := obs.EpochCount(); return err }, obs.Close)

	data := readObservationFixture(t, "algo0010_2015001_v1_trim.rnx")
	quality, err := ParseObservationQC(data, nil)
	if err != nil {
		t.Fatal(err)
	}
	assertConcurrentClose(t, func() error { _, err := quality.Summary(); return err }, quality.Close)
	lint, err := LintRINEXObservation(data)
	if err != nil {
		t.Fatal(err)
	}
	assertConcurrentClose(t, func() error { _, err := lint.Summary(); return err }, lint.Close)
	repairOptions := RINEXRepairOptions{DropUnsupported: true}
	repair, err := RepairRINEXObservation(data, &repairOptions)
	if err != nil {
		t.Fatal(err)
	}
	assertConcurrentClose(t, func() error { _, err := repair.Summary(); return err }, repair.Close)

	nav, err := os.ReadFile("testdata/nav/ESBC00DNK_R_20201770000_01D_MN.rnx")
	if err != nil {
		t.Fatal(err)
	}
	broadcast, err := ParseBroadcastEphemeris(nav)
	if err != nil {
		t.Fatal(err)
	}
	assertConcurrentClose(t, func() error { _, err := broadcast.RecordCount(); return err }, broadcast.Close)

	body := []byte{0x00, 0x01, 0x02, 0x03}
	frame, err := EncodeRTCMFrame(body)
	if err != nil {
		t.Fatal(err)
	}
	frames, err := ScanRTCMFrames(frame)
	if err != nil {
		t.Fatal(err)
	}
	assertConcurrentClose(t, func() error { _, err := frames.Count(); return err }, frames.Close)
	messages, diagnostics, err := DecodeRTCMStream(frame)
	if err != nil {
		t.Fatal(err)
	}
	assertConcurrentClose(t, func() error { _, err := messages.Count(); return err }, messages.Close)
	assertConcurrentClose(t, func() error { _, err := diagnostics.SkippedCount(); return err }, diagnostics.Close)
	tracker, err := NewRTCMLockTimeTracker()
	if err != nil {
		t.Fatal(err)
	}
	assertConcurrentClose(t, tracker.Reset, tracker.Close)

	sbas, err := NewSBASCorrectionStore()
	if err != nil {
		t.Fatal(err)
	}
	assertConcurrentClose(t, func() error { _, _, err := sbas.PreferredGEO(0); return err }, sbas.Close)
	ssr, err := NewSSRCorrectionStore(SSRReferencePointAntennaPhaseCenter)
	if err != nil {
		t.Fatal(err)
	}
	assertConcurrentClose(t, func() error { _, _, err := ssr.Clock("G01"); return err }, ssr.Close)

	glonass, err := ParseRINEXGLONASSRecords(nav)
	if err != nil {
		t.Fatal(err)
	}
	assertConcurrentClose(t, func() error { _, err := glonass.Count(); return err }, glonass.Close)
}

func TestTypedSelectorValidation(t *testing.T) {
	if err := validateGNSSSystem(GNSSSystem(99)); err == nil {
		t.Fatal("invalid GNSS system accepted")
	}
	if err := validateTimeScale(TimeScale(99)); err == nil {
		t.Fatal("invalid time scale accepted")
	}
	if err := validateRTCMStringField(RTCMAntennaStringField(99)); err == nil {
		t.Fatal("invalid RTCM antenna string field accepted")
	}
	if err := validateSignalModulationKind(SignalModulationKind(99)); err == nil {
		t.Fatal("invalid signal modulation accepted")
	}
	if _, err := ModulationLabel(SignalModulation{Kind: SignalModulationKind(99)}); err == nil {
		t.Fatal("invalid modulation reached native call")
	}
	if _, err := DLLThermalNoiseJitter(BPSK(1), DLLTrackingOptions{}, DLLProcessing(99)); err == nil {
		t.Fatal("invalid DLL processing accepted")
	}
	if _, err := BuildRTCMMSM(RTCMMSMInfo{System: GNSSSystemGPS, Kind: RTCMMSMKind(99)}, nil, nil); err == nil {
		t.Fatal("invalid RTCM MSM kind accepted")
	}
	if _, _, err := MinimumLockTimeMS(RTCMMSMKind(99), 0); err == nil {
		t.Fatal("invalid RTCM MSM selector accepted")
	}
	if _, err := NewSSRCorrectionStore(SSRReferencePoint(99)); err == nil {
		t.Fatal("invalid SSR reference point accepted")
	}
	if err := validateSBASSolveMode(SBASSolveMode(99)); err == nil {
		t.Fatal("invalid SBAS solve mode accepted")
	}
	if err := validateSSRMissingAction(SSRMissingCorrectionAction(99)); err == nil {
		t.Fatal("invalid SSR missing action accepted")
	}
	if err := validateRINEXLintSeverity(RINEXLintSeverity(99)); err == nil {
		t.Fatal("invalid RINEX lint severity accepted")
	}
}

func TestObservablesRoutesWithCommittedProducts(t *testing.T) {
	nav, err := os.ReadFile("testdata/nav/ESBC00DNK_R_20201770000_01D_MN.rnx")
	if err != nil {
		t.Fatal(err)
	}
	broadcast, err := ParseBroadcastEphemeris(nav)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := broadcast.Close(); err != nil {
			t.Error(err)
		}
	}()
	records, err := broadcast.Records()
	if err != nil || len(records) == 0 {
		t.Fatalf("broadcast records=%d err=%v", len(records), err)
	}
	broadcastEpoch, err := CivilToJ2000Seconds(CivilDateTime{Year: 2020, Month: 6, Day: 24, Hour: 22})
	if err != nil {
		t.Fatal(err)
	}
	sp3Data, err := os.ReadFile("testdata/trimmed.sp3")
	if err != nil {
		t.Fatal(err)
	}
	sp3, err := LoadSP3(sp3Data)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := sp3.Close(); err != nil {
			t.Error(err)
		}
	}()
	sp3Epochs, err := sp3.Epochs()
	if err != nil || len(sp3Epochs) == 0 {
		t.Fatalf("SP3 epochs=%d err=%v", len(sp3Epochs), err)
	}
	sp3Satellites, err := sp3.Satellites()
	if err != nil || len(sp3Satellites) == 0 {
		t.Fatalf("SP3 satellites=%d err=%v", len(sp3Satellites), err)
	}
	position, err := MissingObservablePositionECEF()
	if err != nil || !math.IsNaN(position[0]) || !math.IsNaN(position[1]) || !math.IsNaN(position[2]) {
		t.Fatalf("missing position=%v err=%v", position, err)
	}
	sp3Rows, err := sp3.EmissionMediaBatch([]string{sp3Satellites[0]}, []float64{sp3Epochs[0]}, ECEF{}, nil)
	if err != nil || len(sp3Rows) != 1 {
		t.Fatalf("SP3 emission rows=%d err=%v", len(sp3Rows), err)
	}
	broadcastRows, err := broadcast.EmissionMediaBatch([]string{records[0].SatelliteID}, []float64{broadcastEpoch}, ECEF{}, nil)
	if err != nil || len(broadcastRows) != 1 {
		t.Fatalf("broadcast emission rows=%d err=%v", len(broadcastRows), err)
	}
	sample, err := broadcast.EphemerisSample([]string{records[0].SatelliteID}, broadcastEpoch, broadcastEpoch+3600, 3600)
	if err != nil || len(sample) != 2 {
		t.Fatalf("broadcast sample rows=%d err=%v", len(sample), err)
	}
	statePosition, stateClock, statePresent, err := broadcast.ObservableState(records[0].SatelliteID, broadcastEpoch)
	if err != nil {
		t.Fatalf("broadcast observable state err=%v", err)
	}
	states, err := broadcast.ObservableStates([]string{records[0].SatelliteID}, []float64{broadcastEpoch})
	if err != nil || len(states) != 1 {
		t.Fatalf("broadcast states=%d err=%v", len(states), err)
	}
	shared, err := broadcast.ObservableStatesShared([]string{records[0].SatelliteID}, broadcastEpoch)
	if err != nil || len(shared) != 1 {
		t.Fatalf("broadcast shared states=%d err=%v", len(shared), err)
	}
	predicted, err := broadcast.PredictObservables(records[0].SatelliteID, ECEF{}, broadcastEpoch, nil)
	if err != nil {
		t.Fatalf("broadcast predicted observables err=%v", err)
	}
	batch, accepted, err := broadcast.PredictObservablesBatch([]PredictRequest{{SatelliteID: records[0].SatelliteID, ReceiverECEF: ECEF{}, TRxJ2000S: broadcastEpoch}}, nil)
	if err != nil || len(batch) != 1 || len(accepted) != 1 {
		t.Fatalf("broadcast predicted batch=%d accepted=%d err=%v", len(batch), len(accepted), err)
	}
	if len(sp3Rows) == 0 || len(broadcastRows) == 0 || len(sample) == 0 || len(states) == 0 || len(shared) == 0 || len(batch) == 0 {
		t.Fatal("observable route returned no rows")
	}
	missingRows, err := broadcast.EmissionMediaBatch([]string{"G32"}, []float64{broadcastEpoch}, ECEF{}, nil)
	if err != nil || len(missingRows) != 1 {
		t.Fatalf("missing broadcast emission rows=%d err=%v", len(missingRows), err)
	}
	missing := missingRows[0]
	if missing.HasPosition || !math.IsNaN(missing.PositionECEFM[0]) || !math.IsNaN(missing.PositionECEFM[1]) || !math.IsNaN(missing.PositionECEFM[2]) || missing.Status != EmissionMediaGap || missing.ResultStatus != StatusSolve {
		t.Fatalf("missing emission sentinel/status = %+v", missing)
	}
	assertBits := func(label string, actual float64, expected uint64) {
		t.Helper()
		if bits := math.Float64bits(actual); bits != expected {
			t.Fatalf("%s bits = %016x, want %016x", label, bits, expected)
		}
	}
	assertMedia := func(label string, row EmissionMediaRow, expected [4]uint64) {
		t.Helper()
		if !row.HasPosition || !row.HasClock || !row.HasIonosphereSlantDelay || !row.HasTroposphereDelay || row.Status != EmissionMediaValid || row.ResultStatus != StatusOK {
			t.Fatalf("%s presence/status = %+v", label, row)
		}
		for index := range row.PositionECEFM {
			assertBits(label+" position", row.PositionECEFM[index], expected[index])
		}
		assertBits(label+" clock", row.ClockS, expected[3])
	}
	assertMedia("SP3", sp3Rows[0], [4]uint64{0x415b0f8f0f9db22d, 0xc17540ec987ef9db, 0x41678ed0e05a1cac, 0xbf0442e1be8b9d32})
	assertMedia("broadcast", broadcastRows[0], [4]uint64{0x4174e5f16a0c82aa, 0x41812ab508561ee9, 0xc12c036525890f1c, 0xbf40e3fc147a882d})
	if sample[0].SatelliteID != records[0].SatelliteID || sample[0].Status != EphemerisSampleValid || !sample[0].HasPosition || !sample[0].HasClock || sample[1].Status != EphemerisSampleValid || !sample[1].HasPosition || !sample[1].HasClock {
		t.Fatalf("sample presence/status = %+v", sample)
	}
	for index, expected := range [][4]uint64{
		{0x4174e5f16a0c82aa, 0x41812ab508561ee9, 0xc12c036525890f1c, 0xbf40e3fc147a882d},
		{0x4174e38491022c05, 0x41812a81ddbb90d7, 0xc1300399cef2c792, 0xbf40e603ade91dcc},
	} {
		if index >= len(sample) {
			break
		}
		for coordinate := range sample[index].PositionECEFM {
			assertBits(fmt.Sprintf("sample[%d] position[%d]", index, coordinate), sample[index].PositionECEFM[coordinate], expected[coordinate])
		}
		assertBits(fmt.Sprintf("sample[%d] clock", index), sample[index].ClockS, expected[3])
	}
	if !statePresent || statePosition != broadcastRows[0].PositionECEFM || math.Float64bits(stateClock) != math.Float64bits(broadcastRows[0].ClockS) {
		t.Fatalf("observable state = %v, %v, %v", statePosition, stateClock, statePresent)
	}
	for label, state := range map[string]ObservableStateRow{"states": states[0], "shared": shared[0]} {
		if state.ElementStatus != ObservableStateValid || state.ResultStatus != StatusOK || !state.HasClock {
			t.Fatalf("%s presence/status = %+v", label, state)
		}
		if state.PositionECEFM != broadcastRows[0].PositionECEFM || math.Float64bits(state.ClockS) != math.Float64bits(broadcastRows[0].ClockS) {
			t.Fatalf("%s values = %+v", label, state)
		}
	}
	if !predicted.HasSatelliteClock || predicted.TransmitOffsetUS != 140618 || predicted != batch[0] || !accepted[0] {
		t.Fatalf("predicted/batch = %+v, %+v, accepted=%v", predicted, batch[0], accepted)
	}
	for index, value := range []float64{predicted.GeometricRangeM, predicted.RangeRateMPerS, predicted.DopplerHz, predicted.SatelliteClockS, predicted.ElevationDeg, predicted.AzimuthDeg, predicted.TransmitTimeJ2000S} {
		assertBits(fmt.Sprintf("predicted scalar[%d]", index), value, [...]uint64{0x41841a04123953f0, 0xbff125202b8c08b6, 0x40168640b4ecfa6a, 0xbf40e3fc0f4a41e9, 0x403f5203b2874415, 0x4056dd79f47ba28e, 0x41c342f04fee003b}[index])
	}
	for index, expected := range [][3]uint64{{0x3fe0a26383e1be53, 0x3feb53f08a085920, 0xbf964c12613259d1}, {0x4174e608821ba770, 0x41812aae03c39c3e, 0xc12c035843c6f30c}} {
		for coordinate := range expected {
			var actual float64
			if index == 0 {
				actual = predicted.LOSUnit[coordinate]
			} else {
				actual = predicted.SatellitePositionECEFM[coordinate]
			}
			assertBits(fmt.Sprintf("predicted vector[%d]", coordinate), actual, expected[coordinate])
		}
	}
}

func TestRTCMFrameRoundTripAndDiagnostics(t *testing.T) {
	body := []byte{0x00, 0x01, 0x02, 0x03}
	frame, e := EncodeRTCMFrame(body)
	if e != nil {
		t.Fatal(e)
	}
	frames, e := ScanRTCMFrames(frame)
	if e != nil {
		t.Fatal(e)
	}
	defer func() {
		if err := frames.Close(); err != nil {
			t.Error(err)
		}
	}()
	if n, e := frames.Count(); e != nil || n != 1 {
		t.Fatalf("frames=%d err=%v", n, e)
	}
	got, e := frames.Body(0)
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(got, body) {
		t.Fatalf("body=%x want %x", got, body)
	}
	messages, e := DecodeRTCM(frame)
	if e != nil {
		t.Fatal(e)
	}
	defer func() {
		if err := messages.Close(); err != nil {
			t.Error(err)
		}
	}()
	if n, e := messages.Count(); e != nil || n != 1 {
		t.Fatalf("messages=%d err=%v", n, e)
	}
	bad := append([]byte(nil), frame...)
	bad[len(bad)-1] ^= 1
	if _, _, e = DecodeRTCMFrame(bad); e == nil {
		t.Fatal("strict frame decoder accepted bad CRC")
	}
	if _, _, e = DecodeRTCMFrame(frame[:len(frame)-1]); e == nil {
		t.Fatal("strict frame decoder accepted truncation")
	}
	if _, e = DecodeRTCM(bad); e != nil {
		t.Fatal("forgiving decoder rejected bad CRC:", e)
	}
	m, diag, e := DecodeRTCMStream(append(bad, frame...))
	if e != nil {
		t.Fatal(e)
	}
	defer func() {
		if err := m.Close(); err != nil {
			t.Error(err)
		}
	}()
	defer func() {
		if err := diag.Close(); err != nil {
			t.Error(err)
		}
	}()
	if n, e := m.Count(); e != nil || n != 1 {
		t.Fatalf("stream messages=%d err=%v", n, e)
	}
	resync, e := diag.ResyncBytes()
	if e != nil || resync == 0 {
		t.Fatalf("stream resync bytes=%d err=%v", resync, e)
	}
	malformedBody := []byte{0x3e, 0xe0} // RTCM message 1006 with a truncated body.
	malformedFrame, e := EncodeRTCMFrame(malformedBody)
	if e != nil {
		t.Fatal(e)
	}
	_, malformedDiag, e := DecodeRTCMStream(malformedFrame)
	if e != nil {
		t.Fatal(e)
	}
	defer func() {
		if err := malformedDiag.Close(); err != nil {
			t.Error(err)
		}
	}()
	if skipped, e := malformedDiag.SkippedCount(); e != nil || skipped != 1 {
		t.Fatalf("malformed skipped=%d err=%v", skipped, e)
	}
	skip, e := malformedDiag.Skipped(0)
	if e != nil || !skip.HasMessageNumber || skip.MessageNumber != 1006 || skip.Reason != RTCMFrameTruncated {
		t.Fatalf("malformed skip=%+v err=%v", skip, e)
	}
	if message, e := malformedDiag.SkippedMessage(0); e != nil || message != "" {
		t.Fatalf("truncated skip message=%q err=%v", message, e)
	}
}

func TestObservationQualityLintRepairAndSignalValidation(t *testing.T) {
	data := readObservationFixture(t, "algo0010_2015001_v1_trim.rnx")
	obs, err := ParseRINEXObservation(data)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := obs.Close(); err != nil {
			t.Fatal(err)
		}
	}()
	options, err := NewObservationQCOptions()
	if err != nil {
		t.Fatal(err)
	}
	report, err := obs.Quality(&options)
	if err != nil {
		t.Fatal(err)
	}
	qcSummary, err := report.Summary()
	if err != nil {
		t.Fatal(err)
	}
	wantQCSummary := ObservationQCSummary{TotalEpochRecords: 2, ObservationEpochs: 2, IntervalS: 30, HasInterval: true, IntervalSource: ObservationQCIntervalSource(1), SatelliteCount: 20, SatelliteSignalCount: 153, SystemSignalCount: 16}
	if qcSummary != wantQCSummary {
		t.Fatalf("QC summary=%+v, want %+v", qcSummary, wantQCSummary)
	}
	if _, err := report.Gaps(); err != nil {
		t.Fatal(err)
	}
	if _, err := report.ClockJumps(); err != nil {
		t.Fatal(err)
	}
	if _, err := report.CycleSlips(); err != nil {
		t.Fatal(err)
	}
	if _, err := report.CycleSlipSystems(); err != nil {
		t.Fatal(err)
	}
	if _, err := report.Satellites(); err != nil {
		t.Fatal(err)
	}
	if _, err := report.SatelliteSignals(); err != nil {
		t.Fatal(err)
	}
	if _, err := report.SystemSignals(); err != nil {
		t.Fatal(err)
	}
	if _, err := report.SatelliteMultipath(); err != nil {
		t.Fatal(err)
	}
	if _, err := report.SystemMultipath(); err != nil {
		t.Fatal(err)
	}
	textRender, err := report.Text()
	if err != nil || len(textRender) == 0 {
		t.Fatalf("quality text bytes=%d err=%v", len(textRender), err)
	}
	htmlRender, err := report.HTML()
	if err != nil || len(htmlRender) == 0 {
		t.Fatalf("quality HTML bytes=%d err=%v", len(htmlRender), err)
	}
	jsonA, err := report.JSON()
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(textRender)); got != "701babc38230d7236883eea8090d18a7b7fc382c21b217b18968b9add6ba15e5" {
		t.Fatalf("quality text hash = %s", got)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(htmlRender)); got != "7c36369028e2ff7e362ff1d6c89af306c46a272663da031045a2ae58e90518b4" {
		t.Fatalf("quality HTML hash = %s", got)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(jsonA)); got != "8d2a92b82022f217a7405b9780c571492d36b3d088fc4c97ece088020bf13e43" {
		t.Fatalf("quality JSON hash = %s", got)
	}
	jsonB, err := report.JSON()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(jsonA, jsonB) {
		t.Fatal("quality JSON copies differ")
	}
	jsonA[0] ^= 1
	if bytes.Equal(jsonA, jsonB) {
		t.Fatal("quality JSON output aliases")
	}
	if err := report.Close(); err != nil {
		t.Fatal(err)
	}
	if err := report.Close(); err != nil {
		t.Fatal(err)
	}

	lint, err := LintRINEXObservation(data)
	if err != nil {
		t.Fatal(err)
	}
	lintSummary, err := lint.Summary()
	if err != nil {
		t.Fatal(err)
	}
	if lintSummary != (RINEXLintSummary{FindingCount: 9, ErrorCount: 8, InfoCount: 1, IsClean: false}) {
		t.Fatalf("lint summary=%+v", lintSummary)
	}
	findings, err := lint.Findings()
	if err != nil {
		t.Fatal(err)
	}
	wantFindings := []RINEXLintFinding{
		{Code: "OBS-H90", Severity: RINEXLintInfo, HasField: true, Field: "header"},
		{Code: "OBS-H12", Severity: RINEXLintError, HasSatellite: true, Satellite: "R05", HasField: true, Field: "GLONASS SLOT / FRQ #"},
		{Code: "OBS-H12", Severity: RINEXLintError, HasSatellite: true, Satellite: "R06", HasField: true, Field: "GLONASS SLOT / FRQ #"},
		{Code: "OBS-H12", Severity: RINEXLintError, HasSatellite: true, Satellite: "R07", HasField: true, Field: "GLONASS SLOT / FRQ #"},
		{Code: "OBS-H12", Severity: RINEXLintError, HasSatellite: true, Satellite: "R09", HasField: true, Field: "GLONASS SLOT / FRQ #"},
		{Code: "OBS-H12", Severity: RINEXLintError, HasSatellite: true, Satellite: "R15", HasField: true, Field: "GLONASS SLOT / FRQ #"},
		{Code: "OBS-H12", Severity: RINEXLintError, HasSatellite: true, Satellite: "R16", HasField: true, Field: "GLONASS SLOT / FRQ #"},
		{Code: "OBS-H12", Severity: RINEXLintError, HasSatellite: true, Satellite: "R17", HasField: true, Field: "GLONASS SLOT / FRQ #"},
		{Code: "OBS-H12", Severity: RINEXLintError, HasSatellite: true, Satellite: "R24", HasField: true, Field: "GLONASS SLOT / FRQ #"},
	}
	if len(findings) != len(wantFindings) {
		t.Fatalf("lint findings=%+v", findings)
	}
	for index := range wantFindings {
		if findings[index] != wantFindings[index] {
			t.Fatalf("lint finding[%d]=%+v want %+v", index, findings[index], wantFindings[index])
		}
	}
	if err := lint.Close(); err != nil {
		t.Fatal(err)
	}
	repairOptions, err := NewRINEXRepairOptions()
	if err != nil {
		t.Fatal(err)
	}
	repairOptions.DropUnsupported = true
	repair, err := RepairRINEXObservation(data, &repairOptions)
	if err != nil {
		t.Fatal(err)
	}
	repairSummary, err := repair.Summary()
	if err != nil {
		t.Fatal(err)
	}
	if repairSummary != (RINEXLintSummary{FindingCount: 8, ErrorCount: 8, IsClean: false}) {
		t.Fatalf("repair summary=%+v", repairSummary)
	}
	actions, err := repair.Actions()
	if err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 || actions[0] != (RINEXRepairAction{ID: "OBS-H90", Message: "dropped 1 unretained header records"}) {
		t.Fatalf("repair actions=%+v", actions)
	}
	repairedText, err := repair.RINEXText()
	if err != nil {
		t.Fatal(err)
	}
	if len(repairedText) == 0 {
		t.Fatal("empty repaired RINEX output")
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(repairedText)); got != "7af542b522ea990ee4bcb9cf08dafa5afc06b8f7e4ef71e647d4adf57f89474b" {
		t.Fatalf("repair text hash = %s", got)
	}
	if err := repair.Close(); err != nil {
		t.Fatal(err)
	}
	crxRepair, err := RepairRINEXObservation(readObservationFixture(t, "ESBC00DNK_R_20201770000_01D_30S_MO_trim.crx"), &repairOptions)
	if err != nil {
		t.Fatal(err)
	}
	repairedCRX, err := crxRepair.CRINEXText()
	if err != nil {
		t.Fatal(err)
	}
	repairedRNX, err := DecodeCRINEX(repairedCRX)
	if err != nil {
		t.Fatal(err)
	}
	reparsedCRX, err := ParseRINEXObservation(repairedRNX)
	if err != nil {
		t.Fatal(err)
	}
	if err := reparsedCRX.Close(); err != nil {
		t.Fatal(err)
	}
	if err := crxRepair.Close(); err != nil {
		t.Fatal(err)
	}

	if _, err := RINEXObservationFrequency(GNSSSystem(99), "C1C", 3.04, nil); err == nil {
		t.Fatal("invalid signal mapping accepted")
	}
	if _, err := RINEXObservationFrequency(GNSSSystem(1), "C1C\x00", 3.04, nil); err == nil {
		t.Fatal("embedded NUL accepted by signal mapping")
	}
	if _, err := RINEXObservationFrequency(GNSSSystem(1), "C1234567", 3.04, nil); err == nil {
		t.Fatal("overlong observation code accepted")
	}
	code, err := CACode(1)
	if err != nil || len(code) != 1023 {
		t.Fatalf("C/A code length=%d err=%v", len(code), err)
	}
	chip, err := CAChip(1, 0)
	if err != nil || chip != code[0] {
		t.Fatalf("C/A first chip=%d code[0]=%d err=%v", chip, code[0], err)
	}
	label, err := ModulationLabel(BPSK(1))
	if err != nil || label == "" {
		t.Fatalf("modulation label=%q err=%v", label, err)
	}
}
