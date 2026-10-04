package sidereon

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
	"sync"
	"testing"
)

// rtkR28Channel7Observations extends the checked-in mixed-system fixture with
// the same out-of-allocation GLONASS channel used by core's RINEX arc test.
func rtkR28Channel7Observations(t *testing.T) []byte {
	t.Helper()
	lines := bytes.Split(readObservationFixture(t, "ESBC00DNK_R_20201770000_01D_30S_MO_trim.rnx"), []byte("\n"))
	countUpdated, lastSlotLine := false, -1
	for i, line := range lines {
		text := string(line)
		if strings.Contains(text, "GLONASS SLOT / FRQ #") {
			lastSlotLine = i
			if strings.HasPrefix(text, " 23 ") {
				lines[i] = []byte(strings.Replace(text, " 23 ", " 24 ", 1))
				countUpdated = true
			}
		}
	}
	if !countUpdated || lastSlotLine < 0 {
		t.Fatal("GLONASS slot count/header lines were not found")
	}
	lines = append(lines[:lastSlotLine+1], append([][]byte{[]byte(fmt.Sprintf("%-60s%-20s", "    R28  7", "GLONASS SLOT / FRQ #"))}, lines[lastSlotLine+1:]...)...)
	var firstEpoch, insertedObservation bool
	for i, line := range lines {
		text := string(line)
		if strings.HasPrefix(text, ">") {
			if firstEpoch {
				break
			}
			firstEpoch = true
			if !strings.HasSuffix(text, " 43") {
				t.Fatalf("first epoch satellite count changed: %q", text)
			}
			lines[i] = []byte(strings.TrimSuffix(text, " 43") + " 44")
			continue
		}
		if firstEpoch && strings.HasPrefix(text, "S23") {
			fields := make([]string, 20)
			for n := range fields {
				fields[n] = strings.Repeat(" ", 16)
			}
			fields[0] = fmt.Sprintf("%14.3f  ", 20_000_000.0) // C1C
			fields[2] = fmt.Sprintf("%14.3f  ", 20_000_010.0) // C2C
			fields[10] = fmt.Sprintf("%14.3f  ", 100_000.0)   // L1C
			fields[12] = fmt.Sprintf("%14.3f  ", 90_000.0)    // L2C
			row := []byte("R28" + strings.Join(fields, ""))
			lines = append(lines[:i], append([][]byte{row}, lines[i:]...)...)
			insertedObservation = true
			break
		}
	}
	if !firstEpoch || !insertedObservation {
		t.Fatal("first epoch or insertion point for R28 was not found")
	}
	return bytes.Join(lines, []byte("\n"))
}

func TestRINEXRTKUnresolvedCarriersRetainNonemptyNativeFixture(t *testing.T) {
	sp3, originalBase, originalRover := rinexRTKFixture(t)
	closeAfterTest(t, originalBase)
	closeAfterTest(t, originalRover)
	observationBytes := rtkR28Channel7Observations(t)
	base, err := ParseRINEXObservation(observationBytes)
	if err != nil {
		t.Fatal(err)
	}
	parsedValues, err := base.Values(0)
	if err != nil {
		t.Fatal(err)
	}
	var r28Values []RINEXObservationValue
	for _, value := range parsedValues {
		if value.SatelliteID == "R28" {
			r28Values = append(r28Values, value)
		}
	}
	if len(r28Values) == 0 {
		for _, line := range bytes.Split(observationBytes, []byte("\n")) {
			if bytes.HasPrefix(line, []byte("R28")) {
				t.Fatalf("fixture row parsed without R28 values; raw row length=%d prefix=%q", len(line), line[:min(len(line), 48)])
			}
		}
		t.Fatal("fixture has no R28 row")
	}
	closeAfterTest(t, base)
	rover, err := ParseRINEXObservation(observationBytes)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, rover)
	wantSingle := []RTKRINEXUnresolvedCarrier{
		{Receiver: RTKRINEXReceiverBase, EpochIndex: 0, SatelliteID: "R28", ObservableCode: "L1C"},
		{Receiver: RTKRINEXReceiverRover, EpochIndex: 0, SatelliteID: "R28", ObservableCode: "L1C"},
	}
	single, err := BuildRINEXRTKArc(sp3, base, rover, &RTKRINEXArcOptions{
		SignalPairs: []RTKRINEXSignalPair{
			{System: GNSSSystemGPS, CodeObservable: "C1C", PhaseObservable: "L1C"},
			{System: GNSSSystemGLONASS, CodeObservable: "C1C", PhaseObservable: "L1C"},
		},
		HasMaxEpochs: true, MaxEpochs: 1, MinCommonSatellites: 4,
	})
	if err != nil {
		t.Fatalf("single-frequency mixed-constellation arc: %v; R28 values=%+v", err, r28Values)
	}
	closeAfterTest(t, single)
	if rows, err := single.UnresolvedCarriers(); err != nil || !reflect.DeepEqual(rows, wantSingle) {
		t.Fatalf("single unresolved carriers = %+v, %v; want %+v", rows, err, wantSingle)
	} else {
		wantSingle = append([]RTKRINEXUnresolvedCarrier(nil), rows...)
	}
	wantDual := []RTKRINEXUnresolvedCarrier{
		{Receiver: RTKRINEXReceiverBase, EpochIndex: 0, SatelliteID: "R28", ObservableCode: "L1C"},
		{Receiver: RTKRINEXReceiverBase, EpochIndex: 0, SatelliteID: "R28", ObservableCode: "L2C"},
		{Receiver: RTKRINEXReceiverRover, EpochIndex: 0, SatelliteID: "R28", ObservableCode: "L1C"},
		{Receiver: RTKRINEXReceiverRover, EpochIndex: 0, SatelliteID: "R28", ObservableCode: "L2C"},
	}
	dual, err := BuildDualFrequencyRINEXRTKArc(sp3, base, rover, &RTKRINEXDualArcOptions{
		SignalPairs: []RTKRINEXDualSignalPair{
			{System: GNSSSystemGPS, Code1Observable: "C1C", Phase1Observable: "L1C", Code2Observable: "C2W", Phase2Observable: "L2W"},
			{System: GNSSSystemGLONASS, Code1Observable: "C1C", Phase1Observable: "L1C", Code2Observable: "C2C", Phase2Observable: "L2C"},
		},
		HasMaxEpochs: true, MaxEpochs: 1, MinCommonSatellites: 4,
	})
	if err != nil {
		t.Fatalf("dual-frequency GLONASS arc: %v", err)
	}
	closeAfterTest(t, dual)
	if rows, err := dual.UnresolvedCarriers(); err != nil || !reflect.DeepEqual(rows, wantDual) {
		t.Fatalf("dual unresolved carriers = %+v, %v; want %+v", rows, err, wantDual)
	} else {
		wantDual = append([]RTKRINEXUnresolvedCarrier(nil), rows...)
	}
	if err := single.Close(); err != nil {
		t.Fatal(err)
	}
	if err := dual.Close(); err != nil {
		t.Fatal(err)
	}
	if err := base.Close(); err != nil {
		t.Fatal(err)
	}
	if err := rover.Close(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(wantSingle, []RTKRINEXUnresolvedCarrier{
		{Receiver: RTKRINEXReceiverBase, EpochIndex: 0, SatelliteID: "R28", ObservableCode: "L1C"},
		{Receiver: RTKRINEXReceiverRover, EpochIndex: 0, SatelliteID: "R28", ObservableCode: "L1C"},
	}) || !reflect.DeepEqual(wantDual, []RTKRINEXUnresolvedCarrier{
		{Receiver: RTKRINEXReceiverBase, EpochIndex: 0, SatelliteID: "R28", ObservableCode: "L1C"},
		{Receiver: RTKRINEXReceiverBase, EpochIndex: 0, SatelliteID: "R28", ObservableCode: "L2C"},
		{Receiver: RTKRINEXReceiverRover, EpochIndex: 0, SatelliteID: "R28", ObservableCode: "L1C"},
		{Receiver: RTKRINEXReceiverRover, EpochIndex: 0, SatelliteID: "R28", ObservableCode: "L2C"},
	}) {
		t.Fatalf("detached unresolved records changed after closing product and observation handles: single=%+v dual=%+v", wantSingle, wantDual)
	}
}

func rinexRTKFixture(t *testing.T) (*SP3, *RINEXObservation, *RINEXObservation) {
	t.Helper()
	sp3Data := readPositioningFixture(t, "trimmed.sp3")
	// Retiming the complete precise product by +12 h 15 min keeps the 13
	// source epochs strictly increasing while placing original row 5 at the
	// RINEX start epoch. The prior fixture edit retimed only the middle five
	// rows, leaving the trailing rows earlier than their predecessors.
	for _, replacement := range []struct{ from, to string }{
		{"#cP2020  6 24 10 30  0.00000000", "#cP2020  6 24 22 45  0.00000000"},
		{"## 2111 297000.00000000   900.00000000 59024 0.4375000000000", "## 2111 341100.00000000   900.00000000 59024 0.9479166666667"},
		{"  2020  6 24 10 30  0.00000000", "  2020  6 24 22 45  0.00000000"},
		{"  2020  6 24 10 45  0.00000000", "  2020  6 24 23  0  0.00000000"},
		{"  2020  6 24 11  0  0.00000000", "  2020  6 24 23 15  0.00000000"},
		{"  2020  6 24 11 15  0.00000000", "  2020  6 24 23 30  0.00000000"},
		{"  2020  6 24 11 30  0.00000000", "  2020  6 24 23 45  0.00000000"},
		{"  2020  6 24 11 45  0.00000000", "  2020  6 25  0  0  0.00000000"},
		{"  2020  6 24 12  0  0.00000000", "  2020  6 25  0 15  0.00000000"},
		{"  2020  6 24 12 15  0.00000000", "  2020  6 25  0 30  0.00000000"},
		{"  2020  6 24 12 30  0.00000000", "  2020  6 25  0 45  0.00000000"},
		{"  2020  6 24 12 45  0.00000000", "  2020  6 25  1  0  0.00000000"},
		{"  2020  6 24 13  0  0.00000000", "  2020  6 25  1 15  0.00000000"},
		{"  2020  6 24 13 15  0.00000000", "  2020  6 25  1 30  0.00000000"},
		{"  2020  6 24 13 30  0.00000000", "  2020  6 25  1 45  0.00000000"},
	} {
		if count := bytes.Count(sp3Data, []byte(replacement.from)); count != 1 {
			t.Fatalf("SP3 timestamp source occurrence count for %q = %d, want 1", replacement.from, count)
		}
		sp3Data = bytes.ReplaceAll(sp3Data, []byte(replacement.from), []byte(replacement.to))
	}
	sp3, err := LoadSP3(sp3Data)
	if err != nil {
		t.Fatal(err)
	}
	epochs, err := sp3.Epochs()
	if err != nil || len(epochs) != 13 || epochs[5] != 646315200 {
		_ = sp3.Close()
		t.Fatalf("retimed SP3 epochs = %v, %v", epochs, err)
	}
	for index := 1; index < len(epochs); index++ {
		if epochs[index]-epochs[index-1] != 900 {
			_ = sp3.Close()
			t.Fatalf("retimed SP3 interval[%d] = %v, want 900 seconds", index, epochs[index]-epochs[index-1])
		}
	}
	obsData := readObservationFixture(t, "ESBC00DNK_R_20201770000_01D_30S_MO_trim.rnx")
	base, err := ParseRINEXObservation(obsData)
	if err != nil {
		_ = sp3.Close()
		t.Fatal(err)
	}
	rover, err := ParseRINEXObservation(obsData)
	if err != nil {
		_ = base.Close()
		_ = sp3.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rover.Close() })
	t.Cleanup(func() { _ = base.Close() })
	t.Cleanup(func() { _ = sp3.Close() })
	return sp3, base, rover
}

func TestRINEXRTKFixtureRoutes(t *testing.T) {
	sp3, base, rover := rinexRTKFixture(t)
	if options, err := DefaultRTKRINEXArcOptions(); err != nil || options.MinCommonSatellites == 0 {
		t.Fatalf("single RINEX defaults = %+v, %v", options, err)
	}
	if options, err := DefaultRTKRINEXDualArcOptions(); err != nil || options.MinCommonSatellites == 0 {
		t.Fatalf("dual RINEX defaults = %+v, %v", options, err)
	}
	single, err := BuildRINEXRTKArc(sp3, base, rover, &RTKRINEXArcOptions{SignalPairs: []RTKRINEXSignalPair{{System: GNSSSystemGPS, CodeObservable: "C1C", PhaseObservable: "L1C"}}, HasMaxEpochs: true, MaxEpochs: 2, MinCommonSatellites: 4, IncludePredictionTime: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = single.Close() })
	if count, err := single.EpochCount(); err != nil || count != 2 {
		t.Fatalf("single epoch count = %d, %v", count, err)
	}
	if skipped, err := single.SkippedEpochCount(); err != nil || skipped != 0 {
		t.Fatalf("single skipped count = %d, %v", skipped, err)
	}
	if unresolved, err := single.UnresolvedCarriers(); err != nil || len(unresolved) != 0 {
		t.Fatalf("single unresolved carriers = %+v, %v", unresolved, err)
	}
	metadata, err := single.EpochMetadata(0)
	if err != nil {
		t.Fatal(err)
	}
	if metadata.BaseCount != 4 || metadata.RoverCount != 4 || metadata.SatellitePositionCount != 4 || metadata.BaseSatellitePositionCount != 4 || metadata.RoverSatellitePositionCount != 4 || !metadata.HasPredictionTime || metadata.PredictionTimeS != 646315200 {
		t.Fatalf("single metadata = %+v", metadata)
	}
	observations, err := single.EpochBaseObservations(0)
	if err != nil || len(observations) != 4 {
		t.Fatalf("single observations = %d, %v", len(observations), err)
	}
	if observations[0].SatelliteID != "G08" || observations[0].AmbiguityID != "G08" || observations[0].CodeM != 24985914.282 || math.Abs(observations[0].PhaseM-24985914.387503017) > 1e-9 || !observations[0].HasLLI || observations[0].LLI != 0 {
		t.Fatalf("single first observation = %+v", observations[0])
	}
	if roverObservations, err := single.EpochRoverObservations(0); err != nil || len(roverObservations) != 4 || roverObservations[0] != observations[0] {
		t.Fatalf("single rover observations = %+v, %v", roverObservations, err)
	}
	positions, err := single.EpochSatellitePositions(0)
	if err != nil || len(positions) != 4 || positions[0].SatelliteID != "G08" || positions[0].PositionM != [3]float64{7093820.244, -22286025.531, 12351111.011} {
		t.Fatalf("single shared positions = %+v, %v", positions, err)
	}
	basePositions, err := single.EpochBaseSatellitePositions(0)
	// Pinned RTKLIB demo5 peph2pos reference for the same retimed SP3 and
	// first-epoch G08 C1C pseudorange: 7093793.018806088 m. The measured
	// Go/C46 difference is 8.91e-6 m; the 1 mm bound allows interpolation
	// arithmetic variation while rejecting the stale 7093002.97 m golden.
	if err != nil || len(basePositions) != 4 || math.Abs(basePositions[0].PositionM[0]-7093793.018806088) > 1e-3 {
		t.Fatalf("single base positions = %+v, %v", basePositions, err)
	}
	roverPositions, err := single.EpochRoverSatellitePositions(0)
	if err != nil || len(roverPositions) != 4 || math.Abs(roverPositions[0].PositionM[0]-basePositions[0].PositionM[0]) > 1e-12 {
		t.Fatalf("single rover positions = %+v, %v", roverPositions, err)
	}
	offsets, err := single.OffsetsM()
	if err != nil || len(offsets) != 4 || offsets[0].ID != "G08" || offsets[0].Value != 0 {
		t.Fatalf("single offsets = %+v, %v", offsets, err)
	}
	wavelengths, err := single.WavelengthsM()
	if err != nil || len(wavelengths) != 4 || wavelengths[0].ID != "G08" || math.Abs(wavelengths[0].Value-0.19029367279836487) > 1e-15 {
		t.Fatalf("single wavelengths = %+v, %v", wavelengths, err)
	}

	dual, err := BuildDualFrequencyRINEXRTKArc(sp3, base, rover, &RTKRINEXDualArcOptions{SignalPairs: []RTKRINEXDualSignalPair{{System: GNSSSystemGPS, Code1Observable: "C1C", Phase1Observable: "L1C", Code2Observable: "C2W", Phase2Observable: "L2W"}}, HasMaxEpochs: true, MaxEpochs: 2, MinCommonSatellites: 4, IncludePredictionTime: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = dual.Close() })
	if count, err := dual.EpochCount(); err != nil || count != 2 {
		t.Fatalf("dual epoch count = %d, %v", count, err)
	}
	if skipped, err := dual.SkippedEpochCount(); err != nil || skipped != 0 {
		t.Fatalf("dual skipped count = %d, %v", skipped, err)
	}
	if unresolved, err := dual.UnresolvedCarriers(); err != nil || len(unresolved) != 0 {
		t.Fatalf("dual unresolved carriers = %+v, %v", unresolved, err)
	}
	dualMetadata, err := dual.EpochMetadata(0)
	if err != nil {
		t.Fatal(err)
	}
	if dualMetadata.JDWhole != 2459025.5 || dualMetadata.JDFraction != 0 || !dualMetadata.HasGapTimeS || dualMetadata.GapTimeS != 646315200 || dualMetadata.ObservationCount != 4 || dualMetadata.SatellitePositionCount != 4 || dualMetadata.BaseSatellitePositionCount != 4 || dualMetadata.RoverSatellitePositionCount != 4 || !dualMetadata.HasPredictionTime || dualMetadata.PredictionTimeS != 646315200 {
		t.Fatalf("dual metadata = %+v", dualMetadata)
	}
	dualObservations, err := dual.EpochObservations(0)
	if err != nil || len(dualObservations) != 4 {
		t.Fatalf("dual observations = %d, %v", len(dualObservations), err)
	}
	first := dualObservations[0]
	if first.SatelliteID != "G08" || first.Base.AmbiguityID != "G08" || first.Base.P1M != 24985914.282 || first.Base.P2M != 24985917.497 || first.Base.F1Hz != 1575420000 || first.Base.F2Hz != 1227600000 || !first.Base.HasLLI1 || !first.Base.HasLLI2 || first.Base.LLI1 != 0 || first.Base.LLI2 != 0 || first.Rover != first.Base {
		t.Fatalf("dual first observation = %+v", first)
	}
	if positions, err := dual.EpochSatellitePositions(0); err != nil || len(positions) != 4 {
		t.Fatalf("dual shared positions = %+v, %v", positions, err)
	}
	if positions, err := dual.EpochBaseSatellitePositions(0); err != nil || len(positions) != 4 {
		t.Fatalf("dual base positions = %+v, %v", positions, err)
	}
	if positions, err := dual.EpochRoverSatellitePositions(0); err != nil || len(positions) != 4 {
		t.Fatalf("dual rover positions = %+v, %v", positions, err)
	}
	if key, err := dual.EpochSortKey(0); err != nil || key != "2020-06-25T00:00:0.000000000" {
		t.Fatalf("dual sort key = %q, %v", key, err)
	}
}

func TestRINEXRTKZeroValueAndBoundaries(t *testing.T) {
	var single RTKRINEXArc
	if err := single.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := single.EpochCount(); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero single EpochCount = %v", err)
	}
	if _, err := single.EpochBaseObservations(0); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero single observations = %v", err)
	}
	if _, err := single.EpochMetadata(0); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero single metadata = %v", err)
	}
	if _, err := single.OffsetsM(); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero single offsets = %v", err)
	}
	var dual RTKRINEXDualFrequencyArc
	if err := dual.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := dual.EpochSortKey(0); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero dual sort key = %v", err)
	}
	if _, err := dual.EpochCount(); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero dual count = %v", err)
	}
	if _, err := dual.EpochMetadata(0); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero dual metadata = %v", err)
	}
	sp3, base, rover := rinexRTKFixture(t)
	if _, err := BuildRINEXRTKArc(nil, base, rover, nil); !errors.Is(err, ErrClosed) {
		t.Fatalf("nil SP3 builder error = %v", err)
	}
	options := &RTKRINEXArcOptions{MaxEpochs: -1, MinCommonSatellites: 4}
	if _, err := BuildRINEXRTKArc(sp3, base, rover, options); err == nil {
		t.Fatal("negative max epochs accepted")
	}
	options = &RTKRINEXArcOptions{MinCommonSatellites: 4, SignalPairs: []RTKRINEXSignalPair{{System: GNSSSystem(99), CodeObservable: "C1C", PhaseObservable: "L1C"}}}
	if _, err := BuildRINEXRTKArc(sp3, base, rover, options); err == nil {
		t.Fatal("invalid single signal system accepted")
	}
	options = &RTKRINEXArcOptions{MinCommonSatellites: 4, SignalPairs: []RTKRINEXSignalPair{{System: GNSSSystemGPS, CodeObservable: "C1\x00", PhaseObservable: "L1C"}}}
	if _, err := BuildRINEXRTKArc(sp3, base, rover, options); err == nil {
		t.Fatal("embedded NUL single signal accepted")
	}
	dualOptions := &RTKRINEXDualArcOptions{MinCommonSatellites: 4, SignalPairs: []RTKRINEXDualSignalPair{{System: GNSSSystemGPS, Code1Observable: "C1C", Phase1Observable: "L1C", Code2Observable: "C2W", Phase2Observable: "L2W\x00"}}}
	if _, err := BuildDualFrequencyRINEXRTKArc(sp3, base, rover, dualOptions); err == nil {
		t.Fatal("embedded NUL dual signal accepted")
	}
	validSingle, err := BuildRINEXRTKArc(sp3, base, rover, &RTKRINEXArcOptions{HasMaxEpochs: true, MaxEpochs: 2, MinCommonSatellites: 4})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = validSingle.Close() })
	if _, err := validSingle.EpochMetadata(-1); err == nil {
		t.Fatal("negative single epoch accepted")
	}
	if _, err := validSingle.EpochMetadata(2); err == nil {
		t.Fatal("out-of-range single epoch accepted")
	}
	validDual, err := BuildDualFrequencyRINEXRTKArc(sp3, base, rover, &RTKRINEXDualArcOptions{HasMaxEpochs: true, MaxEpochs: 2, MinCommonSatellites: 4})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = validDual.Close() })
	if _, err := validDual.EpochObservations(-1); err == nil {
		t.Fatal("negative dual epoch accepted")
	}
	if _, err := validDual.EpochObservations(2); err == nil {
		t.Fatal("out-of-range dual epoch accepted")
	}
}

func TestRINEXRTKCloseReadRace(t *testing.T) {
	sp3, base, rover := rinexRTKFixture(t)
	single, err := BuildRINEXRTKArc(sp3, base, rover, &RTKRINEXArcOptions{HasMaxEpochs: true, MaxEpochs: 2, MinCommonSatellites: 4})
	if err != nil {
		t.Fatal(err)
	}
	dual, err := BuildDualFrequencyRINEXRTKArc(sp3, base, rover, &RTKRINEXDualArcOptions{HasMaxEpochs: true, MaxEpochs: 2, MinCommonSatellites: 4})
	if err != nil {
		_ = single.Close()
		t.Fatal(err)
	}
	var group sync.WaitGroup
	group.Add(3)
	go func() {
		defer group.Done()
		for i := 0; i < 30; i++ {
			_, _ = single.EpochCount()
			_, _ = single.EpochBaseObservations(0)
			_, _ = single.EpochBaseSatellitePositions(0)
			_, _ = single.EpochMetadata(0)
			_, _ = single.OffsetsM()
		}
	}()
	go func() {
		defer group.Done()
		for i := 0; i < 30; i++ {
			_, _ = dual.EpochCount()
			_, _ = dual.EpochObservations(0)
			_, _ = dual.EpochSatellitePositions(0)
			_, _ = dual.EpochMetadata(0)
			_, _ = dual.EpochSortKey(0)
		}
	}()
	go func() {
		defer group.Done()
		for i := 0; i < 10; i++ {
			_ = single.Close()
			_ = dual.Close()
		}
	}()
	group.Wait()
	if err := single.Close(); err != nil {
		t.Fatal(err)
	}
	if err := dual.Close(); err != nil {
		t.Fatal(err)
	}
}
