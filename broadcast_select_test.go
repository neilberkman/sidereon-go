package sidereon

import (
	"bytes"
	"errors"
	"math"
	"os"
	"testing"
)

func TestBroadcastSelectAndComparisonRoutes(t *testing.T) {
	nav, err := os.ReadFile("testdata/nav/ESBC00DNK_R_20201770000_01D_MN.rnx")
	if err != nil {
		t.Fatal(err)
	}
	broadcast, err := ParseBroadcastEphemeris(nav)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := broadcast.Close(); err != nil {
			t.Errorf("broadcast.Close: %v", err)
		}
	})
	info, err := broadcast.RecordsInfo()
	if err != nil || len(info) == 0 {
		t.Fatalf("RecordsInfo len=%d err=%v", len(info), err)
	}
	var chosen BroadcastRecordInfo
	for _, value := range info {
		if len(value.SatelliteID) > 0 && value.SatelliteID[0] == 'G' {
			chosen = value
			break
		}
	}
	if chosen.SatelliteID == "" {
		t.Fatal("fixture has no GPS record")
	}
	gpsEpoch, err := CivilToJ2000Seconds(CivilDateTime{Year: 1980, Month: 1, Day: 6})
	if err != nil {
		t.Fatal(err)
	}
	epoch := gpsEpoch + float64(chosen.ToeWeek)*604800 + chosen.ToeTOWSeconds
	selected, present, err := broadcast.SelectByIssue(chosen.SatelliteID, chosen.Issue, chosen.Message, epoch)
	if err != nil || !present || selected.SatelliteID != chosen.SatelliteID {
		t.Fatalf("SelectByIssue=%+v present=%v err=%v", selected, present, err)
	}
	if _, _, err := broadcast.RecordCNavCorrection(0, 0); err != nil {
		// The fixture has no CNAV correction for every record, but the native
		// route must still be callable and return either a value or presence=false.
		t.Fatalf("RecordCNavCorrection: %v", err)
	}
	commonEpoch, err := CivilToJ2000Seconds(CivilDateTime{Year: 2020, Month: 6, Day: 25, Hour: 12})
	if err != nil {
		t.Fatal(err)
	}
	receiver := [3]float64{1.1e6, 2.2e6, 3.3e6}
	synthetic := SPPConfig{TRxJ2000S: commonEpoch, TRxSecondOfDayS: 43200, DayOfYear: 177, InitialGuess: [4]float64{1e6, 2e6, 3e6, 0}, WithGeodetic: false}
	seenSatellites := make(map[string]bool)
	for _, record := range info {
		satellite := record.SatelliteID
		if seenSatellites[satellite] {
			continue
		}
		seenSatellites[satellite] = true
		position, clock, hasClock, stateErr := broadcast.ObservableState(satellite, commonEpoch)
		if stateErr != nil {
			continue
		}
		dx, dy, dz := position[0]-receiver[0], position[1]-receiver[1], position[2]-receiver[2]
		rangeM := math.Sqrt(dx*dx + dy*dy + dz*dz)
		if hasClock {
			rangeM -= 299792458 * clock
		}
		synthetic.Observations = append(synthetic.Observations, SPPObservation{SatelliteID: satellite, PseudorangeM: rangeM})
	}
	if len(synthetic.Observations) < 4 {
		t.Fatalf("synthetic observations=%d", len(synthetic.Observations))
	}
	solution, err := SolveBroadcast(broadcast, synthetic)
	if err != nil || solution.UsedSatelliteCount < 4 {
		t.Fatalf("synthetic SolveBroadcast used=%d err=%v", solution.UsedSatelliteCount, err)
	}
	dopplers := make([]SPPDopplerObservation, len(synthetic.Observations))
	for i, observation := range synthetic.Observations {
		dopplers[i] = SPPDopplerObservation{SatelliteID: observation.SatelliteID, CarrierHz: 1.57542e9}
	}
	dopplerSolution, err := SolveBroadcastWithDopplerVelocity(broadcast, synthetic, dopplers)
	if err != nil || dopplerSolution.Receiver.UsedSatelliteCount < 4 {
		t.Fatalf("synthetic Doppler solve used=%d err=%v", dopplerSolution.Receiver.UsedSatelliteCount, err)
	}
	if !dopplerSolution.HasVelocity || dopplerSolution.Velocity == nil {
		t.Fatalf("synthetic Doppler solve hasVelocity=%v velocity=%+v", dopplerSolution.HasVelocity, dopplerSolution.Velocity)
	}
	if len(dopplerSolution.Velocity.UsedSatelliteIDs) != dopplerSolution.Velocity.UsedSatelliteCount || len(dopplerSolution.Velocity.ResidualsMPerS) != dopplerSolution.Velocity.UsedSatelliteCount {
		t.Fatalf("synthetic Doppler velocity readers counts=%d ids=%d residuals=%d", dopplerSolution.Velocity.UsedSatelliteCount, len(dopplerSolution.Velocity.UsedSatelliteIDs), len(dopplerSolution.Velocity.ResidualsMPerS))
	}
	if _, err := SolveBroadcastWithDopplerVelocity(broadcast, SPPConfig{}, nil); err == nil {
		t.Fatal("SolveBroadcastWithDopplerVelocity unexpectedly accepted empty inputs")
	}
	if _, _, err := broadcast.RecordGroupDelay(0, 0); err != nil {
		t.Fatalf("RecordGroupDelay: %v", err)
	}
	fullRecords, err := broadcast.Records()
	if err != nil || len(fullRecords) == 0 {
		t.Fatalf("Records len=%d err=%v", len(fullRecords), err)
	}
	constants := ConstellationConstants{GMM3PerS2: 3.986004418e14, OmegaERadPerS: 7.2921151467e-5, DTRF: -4.442807633e-10}
	orbitState, err := BroadcastSatellitePositionECEF(fullRecords[0].Elements, constants, fullRecords[0].Toe.TOWSeconds, false)
	if err != nil {
		t.Fatalf("BroadcastSatellitePositionECEF: %v", err)
	}
	clockOffset, err := BroadcastSatelliteClockOffset(fullRecords[0].Clock, constants, fullRecords[0].Elements, 0, fullRecords[0].Toe.TOWSeconds, 0)
	if err != nil {
		t.Fatalf("BroadcastSatelliteClockOffset: %v", err)
	}
	satelliteState, err := BroadcastSatelliteState(fullRecords[0].Elements, fullRecords[0].Clock, constants, fullRecords[0].Toe.TOWSeconds, 0, false)
	if err != nil {
		t.Fatalf("BroadcastSatelliteState: %v", err)
	}
	if orbitState.KeplerIterations != 4 || math.Float64bits(orbitState.XM) != 0x4174e5f05f72e43c || math.Float64bits(clockOffset.TotalS) != 0xbf40e4000004de97 || math.Float64bits(satelliteState.Clock.TotalS) != 0xbf40e3fde214ad05 {
		t.Fatalf("broadcast frozen values changed: orbit=%+v clock=%+v stateClock=%+v", orbitState, clockOffset, satelliteState.Clock)
	}
	if selected.SatelliteID != "G01" || selected.Issue != 58 || selected.Message != 0 || selected.FitIntervalS != 14400 {
		t.Fatalf("record selection frozen values changed: %+v", selected)
	}

	sp3Data, err := os.ReadFile("testdata/trimmed.sp3")
	if err != nil {
		t.Fatal(err)
	}
	sp3, err := LoadSP3(sp3Data)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := sp3.Close(); err != nil {
			t.Errorf("sp3.Close: %v", err)
		}
	})
	comparisonData := append([]byte(nil), sp3Data...)
	for _, replacement := range [][2][]byte{
		{[]byte("*  2020  6 24 11 45  0.00000000"), []byte("*  2020  6 25  0  0  0.00000000")},
		{[]byte("*  2020  6 24 12  0  0.00000000"), []byte("*  2020  6 25  0 15  0.00000000")},
		{[]byte("*  2020  6 24 12 15  0.00000000"), []byte("*  2020  6 25  0 30  0.00000000")},
		{[]byte("*  2020  6 24 12 30  0.00000000"), []byte("*  2020  6 25  0 45  0.00000000")},
		{[]byte("*  2020  6 24 12 45  0.00000000"), []byte("*  2020  6 25  1  0  0.00000000")},
	} {
		comparisonData = bytes.ReplaceAll(comparisonData, replacement[0], replacement[1])
	}
	comparisonSP3, err := LoadSP3(comparisonData)
	if err != nil {
		t.Fatalf("LoadSP3 comparison fixture: %v", err)
	}
	t.Cleanup(func() {
		if err := comparisonSP3.Close(); err != nil {
			t.Errorf("comparison SP3.Close: %v", err)
		}
	})
	epochs, err := sp3.Epochs()
	if err != nil || len(epochs) == 0 {
		t.Fatalf("SP3 epochs=%d err=%v", len(epochs), err)
	}
	selectedSP3, metadata, err := SelectSP3([]*SP3{sp3}, epochs[0], StalenessPolicyDefault())
	if err != nil {
		t.Fatalf("SelectSP3: %v", err)
	}
	t.Cleanup(func() {
		if err := selectedSP3.Close(); err != nil {
			t.Errorf("selected SP3.Close: %v", err)
		}
	})
	if metadata.Kind != DegradationExact {
		t.Fatalf("selection metadata=%+v", metadata)
	}
	assertConcurrentClose(t, func() error { _, err := selectedSP3.EpochCount(); return err }, selectedSP3.Close)
	selectedRange, _, err := SelectSP3OverRange([]*SP3{sp3}, epochs[0], epochs[len(epochs)-1], StalenessPolicyDefault())
	if err != nil {
		t.Fatalf("SelectSP3OverRange: %v", err)
	}
	t.Cleanup(func() {
		if err := selectedRange.Close(); err != nil {
			t.Errorf("selected range SP3.Close: %v", err)
		}
	})

	anomaly, iterations, err := BroadcastEccentricAnomaly(0.3, 0.1)
	if err != nil || math.IsNaN(anomaly) || iterations <= 0 {
		t.Fatalf("EccentricAnomaly=%v iterations=%d err=%v", anomaly, iterations, err)
	}
	comparisonEpoch, err := CivilToJ2000Seconds(CivilDateTime{Year: 2020, Month: 6, Day: 25})
	if err != nil {
		t.Fatal(err)
	}
	comparisonJD := comparisonEpoch/86400 + 2451545
	comparisonWhole := math.Floor(comparisonJD)
	comparisonFraction := comparisonJD - comparisonWhole
	comparison, err := CompareBroadcast(broadcast, comparisonSP3, []string{"G08"}, []CompareEpoch{{BroadcastTJ2000S: comparisonEpoch, PreciseJDWhole: comparisonWhole, PreciseJDFraction: comparisonFraction, PrecisePlusJDWhole: comparisonWhole, PrecisePlusJDFraction: comparisonFraction, PreciseMinusJDWhole: comparisonWhole, PreciseMinusJDFraction: comparisonFraction}}, 1)
	if err != nil {
		t.Fatalf("CompareBroadcast: %v", err)
	}
	comparisonStats, err := comparison.Overall()
	if err != nil {
		t.Fatalf("comparison overall: %v", err)
	}
	if comparisonStats.Count != 1 || math.Float64bits(comparisonStats.Orbit3DRMSM) != 0x41859b4b8bd6624f || math.Float64bits(comparisonStats.ClockRMSM) != 0x4035c57ba756fc79 {
		t.Fatalf("comparison frozen stats changed: %+v", comparisonStats)
	}
	satelliteCount, err := comparison.SatelliteCount()
	if err != nil {
		t.Fatalf("comparison satellite count: %v", err)
	}
	if satelliteCount != 1 {
		t.Fatalf("comparison satellite count=%d", satelliteCount)
	}
	satellite, satelliteStats, err := comparison.Satellite(0)
	if err != nil || satellite != "G08" || satelliteStats.Count != 1 {
		t.Fatalf("comparison satellite=%q stats=%+v err=%v", satellite, satelliteStats, err)
	}
	assertConcurrentClose(t, func() error { _, err := comparison.Overall(); return err }, comparison.Close)
	windowComparison, err := CompareBroadcastWindow(broadcast, comparisonSP3, []string{"G08"}, CompareWindow{BroadcastWindowStartJ2000S: comparisonEpoch, BroadcastWindowEndJ2000S: comparisonEpoch, PreciseStartJDWhole: comparisonWhole, PreciseStartJDFraction: comparisonFraction, StepS: 900, VelocityHalfS: 1})
	if err != nil {
		t.Fatalf("CompareBroadcastWindow: %v", err)
	}
	t.Cleanup(func() {
		if err := windowComparison.Close(); err != nil {
			t.Errorf("window comparison.Close: %v", err)
		}
	})
	if count, err := windowComparison.SatelliteCount(); err != nil || count != 1 {
		t.Fatalf("window comparison count=%d err=%v", count, err)
	}
}

func TestBroadcastSelectionCloseReadRace(t *testing.T) {
	nav, err := os.ReadFile("testdata/nav/ESBC00DNK_R_20201770000_01D_MN.rnx")
	if err != nil {
		t.Fatal(err)
	}
	broadcast, err := ParseBroadcastEphemeris(nav)
	if err != nil {
		t.Fatal(err)
	}
	assertConcurrentClose(t, func() error { _, err := broadcast.RecordsInfo(); return err }, broadcast.Close)
}

func TestBroadcastSelectionInvalidBoundaries(t *testing.T) {
	if _, _, err := BroadcastEccentricAnomaly(0, 2); err == nil {
		t.Fatal("invalid eccentricity accepted")
	}
	if _, _, err := SelectSP3(nil, 0, StalenessPolicyDefault()); err == nil {
		t.Fatal("empty SP3 selection accepted")
	} else {
		var selectionErr *SelectionError
		if !errors.As(err, &selectionErr) || selectionErr.Status != SelectionEmptyProductSet {
			t.Fatalf("empty SP3 selection error=%T %#v", err, err)
		}
	}
	if _, _, err := (*BroadcastEphemeris)(nil).RecordGroupDelay(-1, 0); err == nil {
		t.Fatal("negative broadcast record index accepted")
	}
}
