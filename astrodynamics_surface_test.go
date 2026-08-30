package sidereon

import (
	"errors"
	"math"
	"sync"
	"testing"
	"time"
)

func TestAstrodynamicsSurfaceFixtureDeterministic(t *testing.T) {
	drag, err := DragParametersFromAreaMass(2.2, 10, 1000, SpaceWeather{F107: 70, F107A: 70, Ap: 4}, 120)
	if err != nil || drag.BCFactorM2PerKg != 0.022 {
		t.Fatalf("drag parameters = %+v, %v", drag, err)
	}
	if weather, err := DefaultSpaceWeather(); err != nil || weather != (SpaceWeather{F107: 150, F107A: 150, Ap: 4}) {
		t.Fatalf("default space weather = %+v, %v", weather, err)
	}
	if _, err := DefaultDecayConfig(); err != nil {
		t.Fatal(err)
	}
	acceleration, err := ForceTwoBodyAcceleration([3]float64{7000, 0, 0}, [3]float64{0, 7.5, 0})
	if err != nil || math.Abs(acceleration[0]+0.00813470289387755) > 1e-15 {
		t.Fatalf("two-body acceleration = %v, %v", acceleration, err)
	}
	if _, err := ForceJ2Acceleration([3]float64{7000, 0, 0}, [3]float64{0, 7.5, 0}); err != nil {
		t.Fatal(err)
	}

	count, err := FrameCatalogCount()
	if err != nil || count < 1 {
		t.Fatalf("frame catalog count = %d, %v", count, err)
	}
	if _, err := FrameCatalogEntry(TerrestrialFrameITRF2020, TerrestrialFrameITRF2014); err != nil {
		t.Fatal(err)
	}
	if _, err := TransformTerrestrialState(TerrestrialState{Position: TerrestrialPosition{PositionM: [3]float64{1, 2, 3}}}, TerrestrialFrameITRF2020, TerrestrialFrameITRF2014, 2020); err != nil {
		t.Fatal(err)
	}

	if _, err := InitialOrbitGibbs([3]float64{7000, 0, 0}, [3]float64{6000, 3500, 0}, [3]float64{2500, 6500, 0}); err != nil {
		t.Fatal(err)
	}
	gibbs, err := InitialOrbitGibbs([3]float64{7000, 0, 0}, [3]float64{6000, 3500, 0}, [3]float64{2500, 6500, 0})
	if err != nil || math.Abs(gibbs.VelocityKmPerS[0]+3.923181748243023) > 1e-14 || math.Abs(gibbs.Theta12Rad-0.5280744484263599) > 1e-14 {
		t.Fatalf("Gibbs result = %+v, %v", gibbs, err)
	}
	if _, err := InitialOrbitHerrickGibbs([3]float64{7000, 0, 0}, [3]float64{6000, 3500, 0}, [3]float64{2500, 6500, 0}, 2459000.1, 2459000.11, 2459000.12); err != nil {
		t.Fatal(err)
	}

	filtered, err := SiderealFilter([]float64{1, 2, 1, 2, 1, 2, 1, 2}, 4, nil)
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "SiderealFilterOutput", filtered.Close)
	values, err := filtered.Filtered()
	if err != nil || len(values) != 8 {
		t.Fatalf("sidereal filtered = %v, %v", values, err)
	}
	if values[0] != 1 || values[1] != 2 || values[2] != 1 || values[3] != 2 || values[4] != 0 || values[5] != 0 || values[6] != 0 || values[7] != 0 {
		t.Fatalf("sidereal frozen result = %v", values)
	}
	if _, err := filtered.Template(); err != nil {
		t.Fatal(err)
	}
	if _, err := filtered.Coverage(); err != nil {
		t.Fatal(err)
	}
	if _, err := filtered.UnderCovered(); err != nil {
		t.Fatal(err)
	}
	if scores, err := SiderealPeriodicityStrength([]float64{1, 2, 1, 2, 1, 2, 1, 2}, []float64{4}); err != nil || len(scores) != 1 {
		t.Fatalf("sidereal periodicity = %v, %v", scores, err)
	}
	if _, err := SiderealRepeatPeriod(0); err != nil {
		t.Fatal(err)
	}

	line1, line2, ok := splitTLELines(string(readPositioningFixture(t, "iss.tle")))
	if !ok {
		t.Fatal("invalid TLE fixture")
	}
	file, err := ParseTLEFile(readPositioningFixture(t, "iss.tle"))
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "TLEFile", file.Close)
	fileCount, err := file.Count()
	if err != nil || fileCount != 1 {
		t.Fatalf("TLE file count = %d, %v", fileCount, err)
	}
	fileTLE, err := file.Satellite(0)
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "file TLE", fileTLE.Close)
	if _, err := file.Name(0); err != nil {
		t.Fatal(err)
	}
	if skipped, err := file.Skipped(); err != nil || skipped != 0 {
		t.Fatalf("TLE file skipped = %d, %v", skipped, err)
	}
	if warnings, err := fileTLE.ChecksumWarnings(); err != nil || len(warnings) != 0 {
		t.Fatalf("TLE checksum warnings = %v, %v", warnings, err)
	}

	epochs := []time.Time{time.Unix(1530619200, 0).UTC(), time.Unix(1530619260, 0).UTC()}
	batch, err := PropagateTLEBatch([]TLEPair{{Line1: line1, Line2: line2}}, epochs, OpsModeAFSPC, false)
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "TLEBatchPropagation", batch.Close)
	satellites, batchEpochs, err := batch.Shape()
	if err != nil || satellites != 1 || batchEpochs != 2 {
		t.Fatalf("TLE batch shape = %d x %d, %v", satellites, batchEpochs, err)
	}
	if values, err := batch.States(); err != nil || len(values) != 2 {
		t.Fatalf("TLE batch states = %d, %v", len(values), err)
	} else {
		want := [3]float64{-2662.376608655317, 6207.721651310084, 622.0930876103647}
		for i := range want {
			if math.Abs(values[0].PositionKm[i]-want[i]) > 1e-12 {
				t.Fatalf("TLE batch frozen position = %v", values[0].PositionKm)
			}
		}
	}

	angles, err := BatchTLELookAngles([]TLEPair{{Line1: line1, Line2: line2}}, PassStation{LatitudeDeg: 51.5074, LongitudeDeg: -0.1278, AltitudeM: 80}, epochs, OpsModeAFSPC, false)
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "TLEBatchLookAngles", angles.Close)
	if values, err := angles.Values(); err != nil || len(values) != 2 {
		t.Fatalf("TLE batch look angles = %d, %v", len(values), err)
	}

	latch, err := NewSGP4DecayLatch()
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "SGP4DecayLatch", latch.Close)
	if _, present, err := latch.FirstFailingEpoch(); err != nil || present {
		t.Fatalf("empty decay latch = %v, %v", present, err)
	}
	if err := latch.Clear(); err != nil {
		t.Fatal(err)
	}

	instant := time.Unix(1530619200, 0).UTC()
	sun, moon, err := SunMoonECEF(instant)
	if err != nil {
		t.Fatal(err)
	}
	wantSun := [3]float64{1.4004792429599945e11, 2.5968614329235005e9, 5.9271896847197105e10}
	wantMoon := [3]float64{-1.9513065601982757e8, -3.4305074416713846e8, -7.582182032577407e7}
	for i := range wantSun {
		if math.Abs(sun[i]-wantSun[i]) > 1e-3 || math.Abs(moon[i]-wantMoon[i]) > 1e-6 {
			t.Fatalf("Sun/Moon frozen result = %v, %v", sun, moon)
		}
	}
	if value, err := SunAngle([3]float64{7000, 0, 0}, [3]float64{149597870, 0, 0}); err != nil || value != 180 {
		t.Fatalf("SunAngle = %v, %v", value, err)
	}
	if value, err := MoonAngle([3]float64{7000, 0, 0}, [3]float64{384400, 0, 0}); err != nil || value != 180 {
		t.Fatalf("MoonAngle = %v, %v", value, err)
	}
	if value, err := SunElevation([3]float64{7000, 0, 0}, [3]float64{149597870, 0, 0}); err != nil || math.Abs(value-90) > 1e-14 {
		t.Fatalf("SunElevation = %v, %v", value, err)
	}
	dpsi, deps, err := NutationIAU2000A(2458300.5)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(dpsi+6.44686677972742e-05) > 1e-18 || math.Abs(deps+2.9701143763914448e-05) > 1e-18 {
		t.Fatalf("nutation frozen result = %.18g, %.18g", dpsi, deps)
	}
	sunAzEl, err := SunAzEl(AlmanacStation{LatitudeDeg: 51.5074, LongitudeDeg: -0.1278, AltitudeKm: 0.08}, instant)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(sunAzEl.AzimuthDeg-177.70896672483974) > 1e-12 || math.Abs(sunAzEl.ElevationDeg-61.41245106651939) > 1e-12 {
		t.Fatalf("Sun az/el frozen result = %+v", sunAzEl)
	}
	if _, err := MoonIlluminationAt(AlmanacStation{LatitudeDeg: 51.5074, LongitudeDeg: -0.1278, AltitudeKm: 0.08}, instant); err != nil {
		t.Fatal(err)
	}
	moonOptions, err := DefaultMoonElevationOptions()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := FindMoonElevationCrossings(AlmanacStation{LatitudeDeg: 51.5074, LongitudeDeg: -0.1278, AltitudeKm: 0.08}, instant, instant.Add(24*time.Hour), &moonOptions); err != nil {
		t.Fatal(err)
	}
	if _, err := FindMoonTransits(AlmanacStation{LatitudeDeg: 51.5074, LongitudeDeg: -0.1278, AltitudeKm: 0.08}, instant, instant.Add(24*time.Hour), moonOptions.StepSeconds, moonOptions.TimeToleranceSeconds); err != nil {
		t.Fatal(err)
	}
	if observation, err := ObserveBody(AlmanacStation{LatitudeDeg: 51.5074, LongitudeDeg: -0.1278, AltitudeKm: 0.08}, instant, ObserveSun, nil, 0, nil, nil, nil); err != nil || observation.Horizontal.RangeKm <= 0 {
		t.Fatalf("Sun observation = %+v, %v", observation, err)
	}
	if _, err := SubSolarPoint([3]float64{1, 2, 3}); err != nil {
		t.Fatal(err)
	}
	if sunBatch, moonBatch, err := SunMoonECEFBatch([]time.Time{instant, instant.Add(time.Minute)}); err != nil || len(sunBatch) != 2 || len(moonBatch) != 2 {
		t.Fatalf("Sun/Moon ECEF batch = %d/%d, %v", len(sunBatch), len(moonBatch), err)
	}

	sp3, err := LoadSP3(readPositioningFixture(t, "trimmed.sp3"))
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "SP3", sp3.Close)
	fit, err := FitSP3ECEFPreciseOrbit(sp3, "G08", nil)
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "OrbitFitReport", fit.Close)
	if solutions, err := fit.Fits(); err != nil || len(solutions) != 1 {
		t.Fatalf("orbit fit solutions = %d, %v", len(solutions), err)
	} else {
		// A least-squares orbit fit compounds per-iteration cross-arch ULP
		// divergence further than a single solve; observed up to ~1.3e-6 km
		// across platforms, so this bound (still ~1cm) is widened from 1e-9.
		// The exact iteration count is itself sensitive to that same noise near
		// the convergence threshold (observed 4/6/8 across three platforms).
		if math.Abs(solutions[0].InitialState.PositionKm[0]-22430.09371997231) > 1e-5 || solutions[0].GeometryQuality.Rank != 6 || solutions[0].Iterations < 1 || solutions[0].Iterations > 20 {
			t.Fatalf("orbit fit frozen result = %+v", solutions[0])
		}
	}

	timeline := []time.Time{time.Date(2018, 7, 3, 0, 0, 0, 0, time.UTC), time.Date(2018, 7, 3, 0, 10, 0, 0, time.UTC), time.Date(2018, 7, 3, 0, 20, 0, 0, time.UTC), time.Date(2018, 7, 3, 0, 30, 0, 0, time.UTC)}
	tle, err := ParseTLE(line1, line2)
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "TLE", tle.Close)
	if _, err := tle.PropagateWithDecayLatch(0, latch); err != nil {
		t.Fatal(err)
	}
	visible, err := VisibleSatellites([]*TLE{tle}, []string{"25544"}, PassStation{LatitudeDeg: 51.5074, LongitudeDeg: -0.1278, AltitudeM: 80}, instant, -90)
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "VisibleList", visible.Close)
	if _, err := visible.Values(); err != nil {
		t.Fatal(err)
	}
	propagated, err := tle.Propagate(timeline)
	if err != nil {
		t.Fatal(err)
	}
	fitSamples := make([]SGP4FitSample, len(timeline))
	for i, instant := range timeline {
		jd, _, err := InstantFromUTCCivil(CivilDateTime{Year: instant.Year(), Month: int(instant.Month()), Day: instant.Day(), Hour: instant.Hour(), Minute: instant.Minute(), Second: float64(instant.Second())})
		if err != nil {
			t.Fatal(err)
		}
		fitSamples[i] = SGP4FitSample{JDWhole: jd.Whole, JDFraction: jd.Fraction, PositionTEMEKm: propagated[i].PositionKm, HasVelocityTEMEKmPS: true, VelocityTEMEKmPS: propagated[i].VelocityKmPerS}
	}
	fitConfig, err := SGP4FitConfigDefaults()
	if err != nil {
		t.Fatal(err)
	}
	fitConfig.MaxNFEV = -1 // ignored by the C ABI while HasMaxNFEV is false
	sgp4Fit, err := FitSGP4TLE(fitSamples, fitConfig)
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "SGP4TLEFit", sgp4Fit.Close)
	if lines, err := sgp4Fit.Lines(); err != nil || len(lines.Line1) == 0 || len(lines.Line2) == 0 {
		t.Fatalf("SGP4 fit lines = %+v, %v", lines, err)
	}
	if stats, err := sgp4Fit.Statistics(); err != nil || stats.NFEV < 1 {
		t.Fatalf("SGP4 fit statistics = %+v, %v", stats, err)
	} else {
		// RMSPositionKm is a residual-noise-floor quantity (already ~1e-5 km,
		// i.e. sub-centimeter) from an iterative fit; cross-arch ULP divergence
		// compounds into a large *relative* swing here even though the fitted
		// position/velocity themselves stay tight (observed ~1.5e-6 km absolute
		// swing, roughly 10% relative). Pinning an exact value at the noise
		// floor is not a meaningful test; assert it stays small instead.
		// NFEV (function evaluation count) is, like iteration count, itself
		// sensitive to cross-arch ULP noise near the convergence threshold
		// (observed 20/23/30 across three platforms).
		if stats.RMSPositionKm < 0 || stats.RMSPositionKm > 5e-5 || stats.NFEV < 1 || stats.NFEV > 60 || stats.SeedRefinePasses != 2 {
			t.Fatalf("SGP4 fit frozen result = %+v", stats)
		}
	}
	fitOMM, err := sgp4Fit.OMM()
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "OMM", fitOMM.Close)
	if encoded, err := fitOMM.KVN(); err != nil || len(encoded) == 0 {
		t.Fatalf("fitted OMM KVN = %d bytes, %v", len(encoded), err)
	} else {
		parsedOMM, err := ParseOMMKVN(encoded)
		if err != nil {
			t.Fatal(err)
		}
		cleanupClose(t, "round-trip OMM", parsedOMM.Close)
		if roundTrip, err := parsedOMM.KVN(); err != nil || len(roundTrip) == 0 {
			t.Fatalf("round-trip OMM = %d bytes, %v", len(roundTrip), err)
		}
	}
}

func TestAstrodynamicsSurfaceInvalidBounds(t *testing.T) {
	if _, err := FrameCatalogEntry(TerrestrialFrame(99), TerrestrialFrameITRF2020); err == nil {
		t.Fatal("invalid terrestrial frame accepted")
	}
	if _, err := PropagateTLEBatch([]TLEPair{{Line1: "a\x00", Line2: "b"}}, nil, OpsModeAFSPC, false); err == nil {
		t.Fatal("embedded NUL accepted")
	}
	if _, err := ParseTLEFileWithOpsMode([]byte(""), OpsMode(99)); err == nil {
		t.Fatal("invalid TLE mode accepted")
	}
	if _, err := FitPreciseEphemerisSamples([]PreciseEphemerisSample{{Satellite: "G01", TimeScale: TimeScale(99)}}, "G01", nil); err == nil {
		t.Fatal("invalid precise sample time scale accepted")
	}
	if _, err := EclipseShadowFractionWithModel([3]float64{7000, 0, 0}, [3]float64{149597870, 0, 0}, EarthShadowModel(99)); err == nil {
		t.Fatal("invalid Earth shadow model accepted")
	}
	if _, err := SiderealRepeatPeriod(99); err == nil {
		t.Fatal("invalid GNSS system accepted")
	}
	decayConfig, err := DefaultDecayConfig()
	if err != nil {
		t.Fatal(err)
	}
	decayConfig.MaxDurationS = -1
	if _, err := EstimateDecay(CartesianState{PositionKm: [3]float64{7000, 0, 0}}, decayConfig); err == nil {
		t.Fatal("negative decay duration accepted")
	}
	sgp4Config, err := SGP4FitConfigDefaults()
	if err != nil {
		t.Fatal(err)
	}
	sgp4Config.EpochKind = SGP4FitEpochSample
	sgp4Config.EpochSampleIndex = -1
	if _, err := FitSGP4TLE([]SGP4FitSample{{PositionTEMEKm: [3]float64{1, 0, 0}}}, sgp4Config); err == nil {
		t.Fatal("negative SGP4 sample index accepted")
	}
	if _, err := FilterSidereal([]float64{1, 2}, 4, &SiderealFilterOptions{PriorPeriods: -1}); err == nil {
		t.Fatal("negative sidereal coverage accepted")
	}
	if _, err := ParseOMMJSON([]byte("{}")); err == nil {
		t.Fatal("invalid OMM accepted")
	}
	if _, err := ParseSpaceWeatherTable([]byte("not a space-weather table")); err == nil {
		t.Fatal("invalid space-weather table accepted")
	}
	if _, err := ParseBroadcastEphemeris(nil); err == nil {
		t.Fatal("empty broadcast ephemeris accepted")
	}
	start := time.Unix(1530619200, 0).UTC()
	end := start.Add(time.Hour)
	if _, err := AlmanacSeasons(nil, start, end, 300, 1); err == nil {
		t.Fatal("almanac seasons accepted a nil SPK")
	}
	if _, err := AlmanacMoonPhases(nil, start, end, 300, 1); err == nil {
		t.Fatal("almanac moon phases accepted a nil SPK")
	}
	if _, err := AlmanacPlanetaryEvents(nil, Mercury, PlanetaryConjunction, start, end, 300, 1); err == nil {
		t.Fatal("almanac planetary events accepted a nil SPK")
	}
	if _, err := AlmanacEclipses(nil, start, end, 300, 1); err == nil {
		t.Fatal("almanac eclipses accepted a nil SPK")
	}
	if _, err := AlmanacMeridianTransits(nil, AlmanacStation{}, TransitBodySun, Mercury, start, end, 300, 1); err == nil {
		t.Fatal("almanac meridian transits accepted a nil SPK")
	}
}

func TestAstrodynamicsSurfaceOwnership(t *testing.T) {
	output, err := SiderealFilter([]float64{1, 2, 1, 2}, 4, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := output.Close(); err != nil {
		t.Fatal(err)
	}
	if err := output.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := output.Filtered(); !errors.Is(err, ErrClosed) {
		t.Fatalf("sidereal filtered after close = %v, want ErrClosed", err)
	}

	concurrent, err := SiderealFilter([]float64{1, 2, 1, 2}, 4, nil)
	if err != nil {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	for i := 0; i < 12; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			if _, readErr := concurrent.Filtered(); readErr != nil && !errors.Is(readErr, ErrClosed) {
				t.Errorf("concurrent sidereal read: %v", readErr)
			}
		}()
	}
	group.Add(1)
	go func() {
		defer group.Done()
		if closeErr := concurrent.Close(); closeErr != nil {
			t.Errorf("concurrent sidereal close: %v", closeErr)
		}
	}()
	group.Wait()
	if err := concurrent.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestAstrodynamicsDecayLatchConcurrentClose(t *testing.T) {
	line1, line2, ok := splitTLELines(string(readPositioningFixture(t, "iss.tle")))
	if !ok {
		t.Fatal("invalid TLE fixture")
	}
	tle, err := ParseTLE(line1, line2)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if closeErr := tle.Close(); closeErr != nil {
			t.Errorf("close TLE: %v", closeErr)
		}
	})
	latch, err := NewSGP4DecayLatch()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if closeErr := latch.Close(); closeErr != nil {
			t.Errorf("close decay latch: %v", closeErr)
		}
	})

	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			if _, _, callErr := latch.FirstFailingEpoch(); callErr != nil && !errors.Is(callErr, ErrClosed) {
				t.Errorf("concurrent latch read: %v", callErr)
			}
		}()
		group.Add(1)
		go func() {
			defer group.Done()
			if _, callErr := tle.PropagateWithDecayLatch(0, latch); callErr != nil && !errors.Is(callErr, ErrClosed) {
				t.Errorf("concurrent latch propagation: %v", callErr)
			}
		}()
		group.Add(1)
		go func() {
			defer group.Done()
			if callErr := latch.Clear(); callErr != nil && !errors.Is(callErr, ErrClosed) {
				t.Errorf("concurrent latch clear: %v", callErr)
			}
		}()
	}
	group.Add(1)
	go func() {
		defer group.Done()
		if callErr := latch.Close(); callErr != nil {
			t.Errorf("concurrent latch close: %v", callErr)
		}
	}()
	group.Wait()
}
