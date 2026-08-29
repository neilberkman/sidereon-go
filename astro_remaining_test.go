package sidereon

import (
	"errors"
	"math"
	"sync"
	"testing"
	"time"
)

const scenarioFixture = `{
  "schema_version": 1,
  "seed": 1,
  "epochs": {"start_j2000_s": 0, "count": 1, "cadence_s": 1},
  "constellation": {"kind": "synthetic_keplerian", "satellites": [{
    "satellite_id": {"system": "Gps", "prn": 1},
    "semi_major_axis_m": 26560000,
    "eccentricity": 0,
    "inclination_rad": 0,
    "raan_rad": 0,
    "arg_perigee_rad": 0,
    "mean_anomaly_rad": 0,
    "epoch_j2000_s": 0,
    "clock_bias_s": 0,
    "clock_drift_s_s": 0
  }]},
  "receiver": {"kind": "static_geodetic", "position": {"lat_rad": 0, "lon_rad": 0, "height_m": 0}},
  "signals": [{
    "satellite_id": {"system": "Gps", "prn": 1},
    "system": "Gps",
    "code_observable": "C1C",
    "phase_observable": "L1C",
    "doppler_observable": "D1C",
    "carrier_hz": 1575420000,
    "carrier_phase_bias_cycles": 0
  }],
  "error_budget": {
    "receiver_clock": {"enabled": false, "bias_s": 0, "drift_s_s": 0, "power_law_coefficients": [0, 0, 0, 0, 0]},
    "satellite_clock": {"enabled": false, "bias_s": 0, "drift_s_s": 0, "power_law_coefficients": [0, 0, 0, 0, 0]},
    "ionosphere": {"enabled": false, "bias_s": 0, "drift_s_s": 0, "power_law_coefficients": [0, 0, 0, 0, 0], "kind": "off"},
    "troposphere": {"enabled": false, "bias_s": 0, "drift_s_s": 0, "power_law_coefficients": [0, 0, 0, 0, 0], "kind": "off"},
    "thermal_noise": {"enabled": false, "bias_s": 0, "drift_s_s": 0, "power_law_coefficients": [0, 0, 0, 0, 0], "pseudorange_sigma_m": 0, "carrier_phase_sigma_m": 0, "doppler_sigma_hz": 0},
    "multipath": {"kind": "off", "enabled": false, "amplitude_m": 0, "reflector_height_m": 0, "phase_rad": 0},
    "elevation_mask_deg": -90
  }
}`

func TestRemainingFrameRoutesUseCResults(t *testing.T) {
	scales, err := TimeScalesFromUTC(CivilDateTime{Year: 2020, Month: 1, Day: 2, Hour: 3, Minute: 4, Second: 5})
	if err != nil {
		t.Fatal(err)
	}
	if value, err := FrameGMSTRadians(scales); err != nil || math.Abs(value-2.5700592614587316) > 1e-14 {
		t.Fatalf("GMST = %.17g, %v", value, err)
	}
	if value, err := FrameGASTRadians(scales); err != nil || math.Abs(value-2.5699856521619338) > 1e-14 {
		t.Fatalf("GAST = %.17g, %v", value, err)
	}
	for _, call := range []func() (Matrix3, error){
		func() (Matrix3, error) { return FrameGCRSToITRSMatrixWithPolarMotion(scales, 0.1, -0.2) },
		func() (Matrix3, error) { return FrameITRSToGCRSMatrixWithPolarMotion(scales, 0.1, -0.2) },
		func() (Matrix3, error) { return FrameMeanOfDateToITRSMatrix(scales) },
		func() (Matrix3, error) { return FrameMeanOfDateToITRSMatrixWithPolarMotion(scales, 0.1, -0.2) },
	} {
		matrix, err := call()
		if err != nil || matrix == (Matrix3{}) {
			t.Fatalf("frame matrix = %v, %v", matrix, err)
		}
	}
	position := [3]float64{7000, -100, 300}
	if value, err := FrameGCRSToITRSWithPolarMotion(position, scales, false, 0.1, -0.2); err != nil || value == ([3]float64{}) {
		t.Fatalf("GCRS to ITRS = %v, %v", value, err)
	}
	if value, err := FrameITRSToGCRSWithPolarMotion(position, scales, 0.1, -0.2); err != nil || value == ([3]float64{}) {
		t.Fatalf("ITRS to GCRS = %v, %v", value, err)
	}
	fromECEF, err := FrameGeodeticFromECEFProj([3]float64{6378137, 0, 0})
	if err != nil || math.Abs(fromECEF.LongitudeDeg) > 1e-12 || math.Abs(fromECEF.LatitudeDeg) > 1e-12 || math.Abs(fromECEF.AltitudeM) > 1e-8 {
		t.Fatalf("ECEF geodetic = %+v, %v", fromECEF, err)
	}
	toITRS, err := FrameGeodeticToITRS(0, 0, 0)
	if err != nil || math.Abs(toITRS[0]-6378.137) > 1e-9 || math.Abs(toITRS[1]) > 1e-12 || math.Abs(toITRS[2]) > 1e-12 {
		t.Fatalf("geodetic ITRS = %v, %v", toITRS, err)
	}
	fromITRS, err := FrameITRSToGeodetic(toITRS)
	if err != nil || math.Abs(fromITRS.LatitudeDeg) > 1e-9 || math.Abs(fromITRS.LongitudeDeg) > 1e-9 || math.Abs(fromITRS.AltitudeKm) > 1e-9 {
		t.Fatalf("ITRS geodetic = %+v, %v", fromITRS, err)
	}
	teme, err := FrameTEMEToGCRS(position, [3]float64{0, 7.5, 1}, scales, false)
	if err != nil || teme.PositionKm == ([3]float64{}) || teme.VelocityKmPerS == ([3]float64{}) {
		t.Fatalf("TEME to GCRS = %+v, %v", teme, err)
	}
}

func TestRemainingCoverageFixtureAndCloseRace(t *testing.T) {
	if _, err := CoverageLookAngles(nil, []PassStation{{}}, time.Unix(1530000000, 0)); err == nil {
		t.Fatal("empty coverage satellite list was accepted")
	}
	tle, err := ParseTLE(fixtureISSLine1, fixtureISSLine2)
	if err != nil {
		t.Fatal(err)
	}
	grid, err := CoverageLookAngles([]*TLE{tle}, []PassStation{{LatitudeDeg: 51.5, LongitudeDeg: -0.1, AltitudeM: 80}}, time.Unix(1530000000, 0))
	if err != nil {
		_ = tle.Close()
		t.Fatal(err)
	}
	if sat, station, err := grid.Dimensions(); err != nil || sat != 1 || station != 1 {
		t.Fatalf("coverage dimensions = %d,%d / %v", sat, station, err)
	}
	cell, err := grid.LookAngle(0, 0)
	if err != nil || !cell.OK || !math.IsNaN(cell.ElevationDeg) && math.IsInf(cell.ElevationDeg, 0) {
		t.Fatalf("coverage cell = %+v, %v", cell, err)
	}
	if _, err := grid.LookAngle(-1, 0); err == nil {
		t.Fatal("negative coverage index was accepted")
	}
	if values, err := CoverageGridAccessCounts(grid, 0); err != nil || len(values) != 1 {
		t.Fatalf("coverage access counts = %v, %v", values, err)
	}
	if values, err := CoverageGridMaxElevationDeg(grid); err != nil || len(values) != 1 {
		t.Fatalf("coverage maximum elevation = %v, %v", values, err)
	}
	if values, err := CoverageGridVisibleMask(grid, 0); err != nil || len(values) != 1 {
		t.Fatalf("coverage visible mask = %v, %v", values, err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 30; j++ {
				_, _, _ = grid.Dimensions()
				_, _ = grid.LookAngle(0, 0)
			}
		}()
	}
	wg.Add(1)
	go func() { defer wg.Done(); _ = grid.Close() }()
	wg.Wait()
	if err := grid.Close(); err != nil {
		t.Fatal(err)
	}
	if _, _, err := grid.Dimensions(); !errors.Is(err, ErrClosed) {
		t.Fatalf("coverage use after close = %v", err)
	}
	_ = tle.Close()
}

func TestRemainingCoverageInverseOrder(t *testing.T) {
	left, err := ParseTLE(fixtureISSLine1, fixtureISSLine2)
	if err != nil {
		t.Fatal(err)
	}
	right, err := ParseTLE(fixtureTCAPrimaryLine1, fixtureTCAPrimaryLine2)
	if err != nil {
		_ = left.Close()
		t.Fatal(err)
	}
	defer func() { _ = left.Close() }()
	defer func() { _ = right.Close() }()
	var wg sync.WaitGroup
	for _, tles := range [][]*TLE{{left, right, left}, {right, left, right}} {
		wg.Add(1)
		go func(values []*TLE) {
			defer wg.Done()
			for i := 0; i < 20; i++ {
				grid, err := CoverageLookAngles(values, []PassStation{{LatitudeDeg: 0, LongitudeDeg: 0, AltitudeM: 0}}, time.Unix(1530000000, 0))
				if err == nil {
					_ = grid.Close()
				}
			}
		}(tles)
	}
	finished := make(chan struct{})
	go func() { wg.Wait(); close(finished) }()
	select {
	case <-finished:
	case <-time.After(2 * time.Second):
		t.Fatal("inverse-order coverage calls did not finish")
	}
}

func TestRemainingScenarioFixtureAndCloseRace(t *testing.T) {
	simulation, err := ScenarioSimulateJSON([]byte(scenarioFixture))
	if err != nil {
		t.Fatal(err)
	}
	if summary, err := ScenarioSimulationSummary(simulation); err != nil || summary.SchemaVersion != 1 || summary.Seed != 1 || summary.ReceiverTruthCount != 1 || summary.ObservationCount != 1 || summary.EpochOffsetCount != 2 || summary.JSONLength == 0 {
		t.Fatalf("scenario summary = %+v, %v", summary, err)
	}
	if offsets, err := ScenarioEpochOffsets(simulation); err != nil || len(offsets) != 2 || offsets[0] != 0 || offsets[1] != 1 {
		t.Fatalf("scenario offsets = %v, %v", offsets, err)
	}
	observations, err := ScenarioObservations(simulation)
	if err != nil || len(observations) != 1 || observations[0].SatelliteID != "G01" || observations[0].CodeObservable != "C1C" || observations[0].CarrierHz != 1575420000 {
		t.Fatalf("scenario observations = %+v, %v", observations, err)
	}
	truth, err := ScenarioReceiverTruth(simulation)
	if err != nil || len(truth) != 1 || math.Abs(truth[0].PositionECEFM[0]-6378137) > 1e-6 {
		t.Fatalf("scenario truth = %+v, %v", truth, err)
	}
	terms, err := ScenarioTerms(simulation)
	if err != nil || len(terms) != 1 || terms[0].GeometricRangeM < 2e7 {
		t.Fatalf("scenario terms = %+v, %v", terms, err)
	}
	serialized, err := ScenarioSimulationJSON(simulation)
	if err != nil || len(serialized) == 0 {
		t.Fatalf("scenario JSON = %d, %v", len(serialized), err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 30; j++ {
				_, _ = simulation.Summary()
				_, _ = simulation.JSON()
			}
		}()
	}
	wg.Add(1)
	go func() { defer wg.Done(); _ = simulation.Close() }()
	wg.Wait()
	if _, err := simulation.Summary(); !errors.Is(err, ErrClosed) {
		t.Fatalf("scenario use after close = %v", err)
	}
	if err := simulation.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestRemainingScenarioRejectsInvalidJSON(t *testing.T) {
	if _, err := ScenarioSimulateJSON([]byte("{}")); err == nil {
		t.Fatal("invalid scenario JSON was accepted")
	}
	if _, err := ScenarioSimulateJSON(append([]byte(scenarioFixture), 0)); err == nil {
		t.Fatal("embedded NUL scenario JSON was accepted")
	}
}
