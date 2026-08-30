package sidereon

import (
	"errors"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func velocitySP3Fixture(t *testing.T) *SP3 {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "trimmed.sp3"))
	if err != nil {
		t.Fatal(err)
	}
	sp3, err := LoadSP3(data)
	if err != nil {
		t.Fatal(err)
	}
	return sp3
}

func velocityFixtureObservations() []VelocityObservation {
	ids := []string{"G08", "G10", "G16", "G18", "G20", "G21", "G26", "G27"}
	observations := make([]VelocityObservation, len(ids))
	for i, id := range ids {
		observations[i].SatelliteID = id
	}
	return observations
}

func assertVelocityFloatBits(t *testing.T, got, want float64) {
	t.Helper()
	if !closeTol(got, want, toleranceM2) {
		t.Fatalf("value = %.17g (%#x), want %.17g (%#x)", got, math.Float64bits(got), want, math.Float64bits(want))
	}
}

func assertVelocityFloatSliceBits(t *testing.T, got, want []float64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("length = %d, want %d", len(got), len(want))
	}
	for i := range got {
		if !closeTol(got[i], want[i], toleranceM2) {
			t.Fatalf("value[%d] = %.17g (%#x), want %.17g (%#x)", i, got[i], math.Float64bits(got[i]), want[i], math.Float64bits(want[i]))
		}
	}
}

func TestVelocitySP3Fixture(t *testing.T) {
	defaults, err := VelocityOptionsDefaults()
	if err != nil {
		t.Fatal(err)
	}
	if defaults.Observable != VelocityObservableRangeRate {
		t.Fatalf("default velocity observable = %v, want range rate", defaults.Observable)
	}

	sp3 := velocitySP3Fixture(t)
	t.Cleanup(func() {
		if err := sp3.Close(); err != nil {
			t.Error(err)
		}
	})
	solution, err := SolveVelocity(sp3, velocityFixtureObservations(), [3]float64{4476130, 530371, 4481350}, 646272000, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := solution.Close(); err != nil {
			t.Error(err)
		}
	})

	velocity, err := solution.Velocity()
	if err != nil {
		t.Fatal(err)
	}
	assertVelocityFloatSliceBits(t, velocity[:], []float64{373.86547240733194, 356.43741989282614, 816.1437404497244})
	clockDrift, err := solution.ClockDrift()
	if err != nil {
		t.Fatal(err)
	}
	assertVelocityFloatBits(t, clockDrift, 2.339954490391944e-06)
	speed, err := solution.Speed()
	if err != nil {
		t.Fatal(err)
	}
	assertVelocityFloatBits(t, speed, 965.8745419739975)
	residuals, err := solution.Residuals()
	if err != nil {
		t.Fatal(err)
	}
	assertVelocityFloatSliceBits(t, residuals, []float64{-68.17584423780806, 261.35140116044, 73.08419919031576, -289.74531253342695, 284.00022050431164, 64.37404887655646, -563.4573978740209, 238.56868491363002})
	covariance, err := solution.StateCovariance()
	if err != nil {
		t.Fatal(err)
	}
	assertVelocityFloatSliceBits(t, covariance[:], []float64{2.3162279558882606, -0.2967750023080173, 1.2173682454030528, 6.139875050756155e-09, -0.2967750023080172, 0.5132897623182687, 0.029378963283489993, -3.9217849890269614e-10, 1.2173682454030543, 0.029378963283490475, 2.0338358084492016, 5.518255216980216e-09, 6.139875050756155e-09, -3.9217849890269624e-10, 5.5182552169802185e-09, 2.1450337074179046e-17})
	count, err := solution.UsedSatelliteCount()
	if err != nil || count != 8 {
		t.Fatalf("used satellite count = %d, err = %v, want 8", count, err)
	}
	used, err := solution.UsedSatelliteIDs()
	if err != nil {
		t.Fatal(err)
	}
	wantUsed := []string{"G08", "G10", "G16", "G18", "G20", "G21", "G26", "G27"}
	if len(used) != len(wantUsed) {
		t.Fatalf("used satellite IDs = %v, want %v", used, wantUsed)
	}
	for i := range used {
		if used[i] != wantUsed[i] {
			t.Fatalf("used satellite ID[%d] = %q, want %q", i, used[i], wantUsed[i])
		}
	}
}

func TestVelocityBroadcastFixture(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "nav", "ESBC00DNK_R_20201770000_01D_MN.rnx"))
	if err != nil {
		t.Fatal(err)
	}
	broadcast, err := ParseBroadcastEphemeris(data)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := broadcast.Close(); err != nil {
			t.Error(err)
		}
	})
	ids := []string{"C05", "C06", "C07", "C08", "C09", "C10", "C11", "C12"}
	observations := make([]VelocityObservation, len(ids))
	for i, id := range ids {
		observations[i].SatelliteID = id
	}
	epoch, err := CivilToJ2000Seconds(CivilDateTime{Year: 2020, Month: 6, Day: 24, Hour: 22})
	if err != nil {
		t.Fatal(err)
	}
	solution, err := SolveVelocityBroadcast(broadcast, observations, [3]float64{4476130, 530371, 4481350}, epoch, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := solution.Close(); err != nil {
			t.Error(err)
		}
	})
	velocity, err := solution.Velocity()
	if err != nil {
		t.Fatal(err)
	}
	assertVelocityFloatSliceBits(t, velocity[:], []float64{-21.29633023043516, 1056.607009835363, 83.37590764658978})
}

func TestVelocityBoundariesAndCloseRace(t *testing.T) {
	sp3 := velocitySP3Fixture(t)
	t.Cleanup(func() {
		if err := sp3.Close(); err != nil {
			t.Error(err)
		}
	})
	invalid := VelocityOptions{Observable: VelocityObservable(99)}
	if _, err := SolveVelocity(sp3, velocityFixtureObservations(), [3]float64{4476130, 530371, 4481350}, 646272000, &invalid); err == nil {
		t.Fatal("invalid velocity observable was accepted")
	}
	withNUL := velocityFixtureObservations()
	withNUL[0].SatelliteID = "G08\x00suffix"
	if _, err := SolveVelocity(sp3, withNUL, [3]float64{4476130, 530371, 4481350}, 646272000, nil); err == nil {
		t.Fatal("embedded-NUL satellite ID was accepted")
	}
	if _, err := SolveVelocity(sp3, velocityFixtureObservations()[:3], [3]float64{4476130, 530371, 4481350}, 646272000, nil); err == nil {
		t.Fatal("too few velocity observations were accepted")
	}
	var zero VelocitySolution
	if err := zero.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := zero.Velocity(); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero velocity solution error = %v, want ErrClosed", err)
	}
	solution, err := SolveVelocity(sp3, velocityFixtureObservations(), [3]float64{4476130, 530371, 4481350}, 646272000, nil)
	if err != nil {
		t.Fatal(err)
	}
	assertConcurrentClose(t, func() error { _, err := solution.Residuals(); return err }, solution.Close)
}

func movingBaselineFixture(t *testing.T) MovingBaselineConfig {
	t.Helper()
	base := [3]float64{4_000_000, 1_000_000, 4_500_000}
	rover := [3]float64{base[0] + 1, base[1] - 1, base[2] + 2}
	ids := []string{"G01", "G02", "G03", "G05", "G06"}
	satellites := [][3]float64{{15_600_000, 0, 20_180_000}, {18_700_000, 2_300_000, 19_400_000}, {21_100_000, -8_000_000, 15_300_000}, {23_000_000, 11_000_000, 12_700_000}, {12_000_000, 18_000_000, 21_000_000}}
	rows := make([]RTKSatMeasurement, len(ids))
	for i, satellite := range satellites {
		baseRange := math.Sqrt((satellite[0]-base[0])*(satellite[0]-base[0]) + (satellite[1]-base[1])*(satellite[1]-base[1]) + (satellite[2]-base[2])*(satellite[2]-base[2]))
		roverRange := math.Sqrt((satellite[0]-rover[0])*(satellite[0]-rover[0]) + (satellite[1]-rover[1])*(satellite[1]-rover[1]) + (satellite[2]-rover[2])*(satellite[2]-rover[2]))
		rows[i] = RTKSatMeasurement{SatelliteID: ids[i], SDAmbiguityID: ids[i], BaseCodeM: baseRange, BasePhaseM: baseRange, RoverCodeM: roverRange, RoverPhaseM: roverRange, BaseTXPos: satellite, RoverTXPos: satellite, Pos: satellite}
	}
	model, err := DefaultRTKMeasurementModel()
	if err != nil {
		t.Fatal(err)
	}
	floatOptions, err := DefaultRTKFloatOptions()
	if err != nil {
		t.Fatal(err)
	}
	fixedOptions, err := DefaultRTKFixedOptions()
	if err != nil {
		t.Fatal(err)
	}
	ambiguous := ids[1:]
	mapEntries := make([]RTKFloatMapEntry, len(ambiguous))
	ambiguitySatellites := make([]RTKAmbiguitySatellite, len(ambiguous))
	for i, id := range ambiguous {
		mapEntries[i] = RTKFloatMapEntry{ID: id, Value: 0.190293672798365}
		ambiguitySatellites[i] = RTKAmbiguitySatellite{ID: id, SatelliteID: id}
	}
	epoch := MovingBaselineEpoch{BasePositionM: base, Epoch: RTKEpoch{References: rows[:1], NonReference: rows[1:], DTS: 30}, AmbiguityIDs: append([]string(nil), ambiguous...), AmbiguitySatellites: ambiguitySatellites, WavelengthsM: mapEntries, OffsetsM: make([]RTKFloatMapEntry, len(ambiguous))}
	for i, id := range ambiguous {
		epoch.OffsetsM[i].ID = id
	}
	return MovingBaselineConfig{Epochs: []MovingBaselineEpoch{epoch, epoch, epoch}, Model: model, FloatOptions: floatOptions, FixedOptions: fixedOptions, InitialBaselineM: [3]float64{1, -1, 2}}
}

func TestMovingBaselineFixture(t *testing.T) {
	solution, err := SolveMovingBaseline(movingBaselineFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := solution.Close(); err != nil {
			t.Error(err)
		}
	})
	count, err := solution.EpochCount()
	if err != nil || count != 3 {
		t.Fatalf("epoch count = %d, err = %v, want 3", count, err)
	}
	value, err := solution.Epoch(0)
	if err != nil {
		t.Fatal(err)
	}
	if value.Status != MovingBaselineStatusFixed || value.Float.UsedSatCount != 4 || value.Fixed.FixedAmbiguityCount != 4 || !value.Float.Converged || !value.Fixed.Converged {
		t.Fatalf("unexpected moving-baseline summary: %+v", value)
	}
	assertVelocityFloatSliceBits(t, value.BaselineM[:], []float64{0.9999966267483605, -1.0000055280187778, 2.000002162929948})
	assertVelocityFloatBits(t, value.BaselineLengthM, 2.449492388496173)
	if value.Float.NObservations != 8 || value.Fixed.NObservations != 8 || value.Float.AmbiguityCount != 4 || value.Fixed.FreeAmbiguityCount != 0 {
		t.Fatalf("unexpected moving-baseline counts: %+v", value)
	}
}

func TestMovingBaselineBoundariesAndCloseRace(t *testing.T) {
	var zero MovingBaselineSolution
	if err := zero.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := zero.EpochCount(); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero moving epoch count error = %v, want ErrClosed", err)
	}
	if _, err := zero.Epoch(-1); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero moving epoch error = %v, want ErrClosed", err)
	}
	config := movingBaselineFixture(t)
	config.FixedOptions.MaxIterations = -1
	if _, err := SolveMovingBaseline(config); err == nil {
		t.Fatal("negative fixed iteration count was accepted")
	}
	config = movingBaselineFixture(t)
	config.Epochs[0].AmbiguityIDs[0] = "G02\x00suffix"
	if _, err := SolveMovingBaseline(config); err == nil {
		t.Fatal("embedded-NUL ambiguity ID was accepted")
	}
	solution, err := SolveMovingBaseline(movingBaselineFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	assertConcurrentClose(t, func() error { _, err := solution.EpochCount(); return err }, solution.Close)
}
