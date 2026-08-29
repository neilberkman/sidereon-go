package sidereon

import (
	"errors"
	"math"
	"testing"
)

func TestValueRoutesDeterministic(t *testing.T) {
	t.Parallel()

	t.Run("coordinates", func(t *testing.T) {
		t.Parallel()
		position, err := GeodeticToECEF(Geodetic{})
		if err != nil {
			t.Fatal(err)
		}
		if position != (ECEF{X: 6378137}) {
			t.Fatalf("equator ECEF = %#v", position)
		}
		geodetic, err := ECEFToGeodetic(position)
		if err != nil {
			t.Fatal(err)
		}
		if geodetic != (Geodetic{}) {
			t.Fatalf("equator geodetic = %#v", geodetic)
		}

		receiver := Geodetic{LatitudeRad: math.Float64frombits(0x3fe921fb54442d18), LongitudeRad: math.Float64frombits(0x3fc657184ae74487)}
		los := []LineOfSight{
			{EX: math.Float64frombits(0x3fdaa2fbd945ccc0), EY: math.Float64frombits(0x3fb2c97bb8ccdf37), EZ: math.Float64frombits(0x3fed0079302dd766)},
			{EX: math.Float64frombits(0x3fe09f95aba18faf), EY: math.Float64frombits(0x3feb4d27ce94dbce), EZ: math.Float64frombits(0x3fa8408293b0d4f0)},
			{EX: math.Float64frombits(0x3fe8f55f9ae5db06), EY: math.Float64frombits(0xbfe3f829842cce52), EZ: math.Float64frombits(0x3fa8408293b0d4b8)},
			{EX: math.Float64frombits(0xbfd3493dd2a6d2aa), EY: math.Float64frombits(0x3fe97b2bf4a3bea0), EZ: math.Float64frombits(0x3fe0c8dc2e423980)},
		}
		dop, err := ComputeDOP(los, []float64{1, 1, 1, 1}, receiver)
		if err != nil {
			t.Fatal(err)
		}
		want := DOP{
			GDOP: math.Float64frombits(0x40094512188a0ee8),
			PDOP: math.Float64frombits(0x400670b4a1e5da09),
			HDOP: math.Float64frombits(0x3ff9ebd66303910a),
			VDOP: math.Float64frombits(0x400251acf780e6a6),
			TDOP: math.Float64frombits(0x3ff73cdc1aa52898),
		}
		if dop != want {
			t.Fatalf("DOP = %#v, want %#v", dop, want)
		}
	})

	t.Run("time", func(t *testing.T) {
		t.Parallel()
		civil := CivilDateTime{Year: 2020, Month: 6, Day: 25, Hour: 12}
		seconds, err := CivilToJ2000Seconds(civil)
		if err != nil || seconds != 646358400 {
			t.Fatalf("CivilToJ2000Seconds = %.17g, %v", seconds, err)
		}
		back, err := J2000SecondsToCivil(int64(seconds))
		if err != nil || back != civil {
			t.Fatalf("J2000SecondsToCivil = %#v, %v", back, err)
		}
		date, instantSeconds, err := InstantFromUTCCivil(civil)
		if err != nil {
			t.Fatal(err)
		}
		if date != (JulianDate{Whole: 2459025.5, Fraction: 0.5}) || instantSeconds != seconds {
			t.Fatalf("InstantFromUTCCivil = %#v, %.17g", date, instantSeconds)
		}
		fromSplit, err := SplitJulianDateToJ2000Seconds(date)
		if err != nil || fromSplit != seconds {
			t.Fatalf("SplitJulianDateToJ2000Seconds = %.17g, %v", fromSplit, err)
		}
		gps, available, err := CivilToGPSSeconds(CivilDateTime{Year: 2020, Month: 6, Day: 25})
		if err != nil || !available || gps <= 0 {
			t.Fatalf("CivilToGPSSeconds = %.17g, %t, %v", gps, available, err)
		}
		if secondsOfWeek, err := GNSSSecondsOfWeek(CivilDateTime{Year: 2020, Month: 6, Day: 25}); err != nil || secondsOfWeek != 345600 {
			t.Fatalf("GNSSSecondsOfWeek = %.17g, %v", secondsOfWeek, err)
		}
		if week, err := GNSSWeekAndSecondsOfWeek(gps); err != nil || week != (GNSSWeekSeconds{Week: 2111, SecondsOfWeek: 345600}) {
			t.Fatalf("GNSSWeekAndSecondsOfWeek = %#v, %v", week, err)
		}
		gpsOffset, err := GPSUTCOffset(2459025.5)
		if err != nil || gpsOffset != 18 {
			t.Fatalf("GPSUTCOffset = %.17g, %v", gpsOffset, err)
		}
		taiOffset, err := TAIUTCOffset(2459025.5)
		if err != nil || taiOffset != 37 {
			t.Fatalf("TAIUTCOffset = %.17g, %v", taiOffset, err)
		}
		leap, err := LeapSeconds(2020, 6, 25)
		if err != nil || leap != 37 {
			t.Fatalf("LeapSeconds = %.17g, %v", leap, err)
		}
		if value, err := TimeScaleOffset(GPST, BDT); err != nil || value != -14 {
			t.Fatalf("GPST to BDT offset = %.17g, %v", value, err)
		}
		if value, err := TimeScaleOffset(TAI, TT); err != nil || value != 32.184 {
			t.Fatalf("TAI to TT offset = %.17g, %v", value, err)
		}
		if value, err := TimeScaleOffsetAt(UTC, GPST, 2457754.5); err != nil || value != 18 {
			t.Fatalf("UTC to GPST offset = %.17g, %v", value, err)
		}
		if label, err := TimeScaleLabel(GPST); err != nil || label != "GPST" {
			t.Fatalf("TimeScaleLabel = %q, %v", label, err)
		}
		timescales, err := TimeScalesFromUTC(civil)
		want := TimeScales{
			JDWhole: 2459026, UT1Fraction: -2.803625000000003e-06, TTFraction: 0.0008007407407407408,
			TDBFraction: 0.0008007438069242998, JDUT1: 2459025.9999971963, JDTT: 2459026.000800741,
			JDTDB: 2459026.0008007437,
		}
		if err != nil || !sameTimeScales(timescales, want) {
			t.Fatalf("TimeScalesFromUTC = %#v, %v", timescales, err)
		}
	})

	t.Run("carrier", func(t *testing.T) {
		t.Parallel()
		f1, err := Frequency(GNSSSystemGPS, CarrierBandL1)
		if err != nil || f1 != 1575420000 {
			t.Fatalf("GPS L1 frequency = %.17g, %v", f1, err)
		}
		f2, err := Frequency(GNSSSystemGPS, CarrierBandL2)
		if err != nil || f2 != 1227600000 {
			t.Fatalf("GPS L2 frequency = %.17g, %v", f2, err)
		}
		wavelength, err := Wavelength(GNSSSystemGPS, CarrierBandL1)
		if err != nil || !(math.Abs(wavelength-0.19029367279836487) <= 1e-15) {
			t.Fatalf("GPS L1 wavelength = %.17g, %v", wavelength, err)
		}
		if frequency, err := GLONASSG1Frequency(0); err != nil || frequency != 1602000000 {
			t.Fatalf("GLONASS G1 frequency = %.17g, %v", frequency, err)
		}
		if frequency, err := DefaultSPPFrequency(GNSSSystemGPS); err != nil || frequency != f1 {
			t.Fatalf("default SPP frequency = %.17g, %v", frequency, err)
		}
		if label, err := CarrierBandLabel(CarrierBandL1); err != nil || label != "l1" {
			t.Fatalf("CarrierBandLabel = %q, %v", label, err)
		}
		if label, err := GNSSSystemLabel(GNSSSystemGPS); err != nil || label != "GPS" {
			t.Fatalf("GNSSSystemLabel = %q, %v", label, err)
		}
		if value, err := GeometryFree(2e7, 2e7); err != nil || value != 0 {
			t.Fatalf("GeometryFree = %.17g, %v", value, err)
		}
		if value, err := IonosphereFreePseudorange(2e7, 2e7, f1, f2); err != nil || math.Abs(value-2e7) > 1e-8 {
			t.Fatalf("IonosphereFreePseudorange = %.17g, %v", value, err)
		}
		if value, err := IonosphereFreePhaseMeters(2e7, 2e7, f1, f2); err != nil || math.Abs(value-2e7) > 1e-8 {
			t.Fatalf("IonosphereFreePhaseMeters = %.17g, %v", value, err)
		}
		phase, phaseErr := PhaseMeters(1e8, f1)
		melbourneWubbena, melbourneErr := MelbourneWubbena(1e8, 8e7, 2e7, 2e7, f1, f2)
		narrowLane, narrowErr := NarrowLaneCode(2e7, 2e7, f1, f2)
		wideLane, wideErr := WideLaneCycles(1e8, 8e7, 2e7, 2e7, f1, f2)
		wideWavelength, wideWavelengthErr := WideLaneWavelength(f1, f2)
		gamma, gammaErr := IonosphericGamma(f1, f2)
		noiseAmplification, noiseAmplificationErr := IonosphereFreeNoiseAmplification(f1, f2)
		phaseCycles, phaseCyclesErr := IonosphereFreePhaseCycles(1e8, 0.9e8, f1, f2)
		codeMinusCarrier, codeMinusCarrierErr := CodeMinusCarrier(2e7, 1e8, f1)
		for _, result := range []struct {
			name  string
			value float64
			want  float64
			err   error
		}{
			{"phase", phase, 19029367.279836487, phaseErr},
			{"melbourne-wubbena", melbourneWubbena, -2761631.9935598895, melbourneErr},
			{"narrow-lane", narrowLane, 20000000, narrowErr},
			{"wide-lane", wideLane, -3204052.7183642518, wideErr},
			{"wide-lane-wavelength", wideWavelength, 0.86191840032200562, wideWavelengthErr},
			{"gamma", gamma, 2.5457277801631601, gammaErr},
			{"noise-amplification", noiseAmplification, 2.9782552444447372, noiseAmplificationErr},
			{"phase-cycles", phaseCycles, 14470162.925113961, phaseCyclesErr},
			{"code-minus-carrier", codeMinusCarrier, 970632.72016351297, codeMinusCarrierErr},
		} {
			if result.err != nil || !(math.Abs(result.value-result.want) <= 1e-8) {
				t.Fatalf("%s result = %.17g, %v", result.name, result.value, result.err)
			}
		}
		if pair, present, err := DefaultIonosphereFreePair(GNSSSystemGPS); err != nil || !present || pair != (CarrierPair{Band1: CarrierBandL1, Band2: CarrierBandL2}) {
			t.Fatalf("DefaultIonosphereFreePair = %#v, %t, %v", pair, present, err)
		}
	})

	t.Run("statistics", func(t *testing.T) {
		t.Parallel()
		if value, err := NIS(2, 4); err != nil || value != 1 {
			t.Fatalf("NIS = %.17g, %v", value, err)
		}
		if value, err := ExpectedNIS(3); err != nil || value != 3 {
			t.Fatalf("ExpectedNIS = %.17g, %v", value, err)
		}
		threshold, err := NISThreshold(1, 0.95)
		if err != nil || !(math.Abs(threshold-3.8414588206941245) <= 1e-14) {
			t.Fatalf("NISThreshold = %.17g, %v", threshold, err)
		}
		gate, err := TestNISGate(2, 4, 1, 0.95)
		if err != nil || !gate.InGate || gate.NIS != 1 || gate.DOF != 1 || gate.Threshold != threshold {
			t.Fatalf("TestNISGate = %#v, %v", gate, err)
		}
		if value, err := ChiSquareInverse(0.999, 1); err != nil || !(math.Abs(value-10.827566170662628) <= 1e-14) {
			t.Fatalf("ChiSquareInverse = %.17g, %v", value, err)
		}
	})
}

func TestValueRoutesFixtureGeodesic(t *testing.T) {
	latitude1, longitude1, azimuth1 := 36.530042355041, 0.0, 176.125875162171
	latitude2, longitude2, azimuth2 := -48.164270779097768864, 5.762344694676510456, 175.334308316285410561
	distance := 9398502.0434687
	inverse, err := GeodesicInverse(latitude1, longitude1, latitude2, longitude2)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(inverse.DistanceM-distance) > 1e-8 || math.Abs(inverse.InitialAzimuthDeg-azimuth1) > 5e-13 || math.Abs(inverse.FinalAzimuthDeg-azimuth2) > 5e-13 {
		t.Fatalf("GeodesicInverse = %#v", inverse)
	}
	direct, err := GeodesicDirect(latitude1, longitude1, azimuth1, distance)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(direct.LatitudeDeg-latitude2) > 2e-13 || math.Abs(direct.LongitudeDeg-longitude2) > 2e-13 || math.Abs(direct.FinalAzimuthDeg-azimuth2) > 5e-13 {
		t.Fatalf("GeodesicDirect = %#v", direct)
	}
}

func TestValueRoutesFixtureMetrics(t *testing.T) {
	covariance := [3][3]float64{{9, 0, 0}, {0, 9, 0}, {0, 0, 9}}
	metrics, kind, err := ErrorMetricsFromENU(covariance)
	if err != nil || kind != ErrorMetricsNone {
		t.Fatalf("ErrorMetricsFromENU = %#v, %v, %v", metrics, kind, err)
	}
	if metrics.SigmaEM != 3 || metrics.SigmaNM != 3 || metrics.SigmaUM != 3 || math.Abs(metrics.DRMS-4.242640687119285) > 1e-12 {
		t.Fatalf("isotropic metrics = %#v", metrics)
	}
	ellipse, kind, err := ErrorEllipseFromENU(covariance)
	if err != nil || kind != ErrorMetricsNone || ellipse != (ErrorEllipse{SemiMajorM: 3, SemiMinorM: 3}) {
		t.Fatalf("ErrorEllipseFromENU = %#v, %v, %v", ellipse, kind, err)
	}
	radius, kind, err := HorizontalProtectionRadius(covariance, 0.95)
	if err != nil || kind != ErrorMetricsNone || radius.RadiusM != metrics.R95.RadiusM || radius.Probability != 0.95 {
		t.Fatalf("HorizontalProtectionRadius = %#v, %v, %v", radius, kind, err)
	}
	vertical, kind, err := VerticalProtectionRadius(9, 0.5)
	if err != nil || kind != ErrorMetricsNone || math.Abs(vertical-2.023469250588245) > 1e-12 {
		t.Fatalf("VerticalProtectionRadius = %.17g, %v, %v", vertical, kind, err)
	}
	ellipse2, err := ErrorEllipse2x2([4]float64{9, 2, 2, 4}, 0.95)
	if err != nil || ellipse2.SemiMajor <= ellipse2.SemiMinor || ellipse2.Confidence != 0.95 {
		t.Fatalf("ErrorEllipse2x2 = %#v, %v", ellipse2, err)
	}

	_, kind, err = ErrorMetricsFromENU([3][3]float64{{1, 0, 0}, {0, -1, 0}, {0, 0, 1}})
	if err == nil || kind != ErrorMetricsNotPositiveSemidefinite {
		t.Fatalf("non-PSD metrics = kind %v, err %v", kind, err)
	}
	var statusErr *StatusError
	if !errors.As(err, &statusErr) || statusErr.Code != StatusInvalidArgument || statusErr.Detail == "" {
		t.Fatalf("non-PSD status error = %T, %v", err, err)
	}
	_, kind, err = HorizontalProtectionRadius(covariance, 1)
	if err == nil || kind != ErrorMetricsInvalidProbability {
		t.Fatalf("invalid probability = kind %v, err %v", kind, err)
	}
}

func TestValueRoutesSmokeFramesAndAstro(t *testing.T) {
	t.Parallel()
	identity, err := PolarMotionMatrix(0, 0)
	if err != nil || identity != (Matrix3{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}}) {
		t.Fatalf("PolarMotionMatrix(0, 0) = %#v, %v", identity, err)
	}
	vector, err := MatrixVectorMultiply(identity, [3]float64{1, 2, 3})
	if err != nil || vector != [3]float64{1, 2, 3} {
		t.Fatalf("MatrixVectorMultiply = %#v, %v", vector, err)
	}
	timescales, err := TimeScalesFromUTC(CivilDateTime{Year: 2020, Month: 6, Day: 24})
	if err != nil {
		t.Fatal(err)
	}
	rotation, err := GCRSToITRSMatrix(timescales)
	wantRotation := Matrix3{
		{0.040946277150422414, -0.9991613457952838, -8.634858331273963e-05},
		{0.9991594329170512, 0.04094636772494656, -0.001955142708403866},
		{0.0019570386805978345, -6.220186334629785e-06, 0.9999980849786221},
	}
	if err != nil || !sameMatrix(rotation, wantRotation, 1e-14) {
		t.Fatalf("GCRSToITRSMatrix = %#v, %v", rotation, err)
	}
	wantInverseRotation := Matrix3{
		{0.040946277150422414, 0.9991594329170512, 0.0019570386805978345},
		{-0.9991613457952838, 0.04094636772494656, -6.220186334629785e-06},
		{-8.634858331273963e-05, -0.001955142708403866, 0.9999980849786221},
	}
	if inverseRotation, err := ITRSToGCRSMatrix(timescales); err != nil || !sameMatrix(inverseRotation, wantInverseRotation, 1e-14) {
		t.Fatalf("ITRSToGCRSMatrix = %#v, %v", inverseRotation, err)
	}
	wantPosition := [3]float64{1495.4969153069437, 6942.029239951248, 1313.7043076618584}
	if position, err := GCRSToITRS([3]float64{7000, -1210, 1300}, timescales, false); err != nil || !sameVector(position, wantPosition, 1e-9) {
		t.Fatalf("GCRSToITRS = %#v, %v", position, err)
	}
	wantInversePosition := [3]float64{-919.814823491898, -7043.682611756407, 1301.7587930661882}
	if position, err := ITRSToGCRS([3]float64{7000, -1210, 1300}, timescales); err != nil || !sameVector(position, wantInversePosition, 1e-9) {
		t.Fatalf("ITRSToGCRS = %#v, %v", position, err)
	}
	wantTopocentric := Topocentric{AzimuthDeg: 92.65685539936756, ElevationDeg: -32.40968954077646, RangeKm: 8234.772247413492}
	if topocentric, err := GCRSToTopocentric([3]float64{7000, -1210, 1300}, GroundStation{LatitudeDeg: 51.5, LongitudeDeg: -0.1, AltitudeKm: 0.08}, timescales, false); err != nil || !(math.Abs(topocentric.AzimuthDeg-wantTopocentric.AzimuthDeg) <= 1e-12) || !(math.Abs(topocentric.ElevationDeg-wantTopocentric.ElevationDeg) <= 1e-12) || !(math.Abs(topocentric.RangeKm-wantTopocentric.RangeKm) <= 1e-9) {
		t.Fatalf("GCRSToTopocentric = %#v, %v", topocentric, err)
	}
	wantRangeRate := DopplerRangeRate{RangeRateKmS: 1.8940111566276965, DopplerRatio: -6.31774117755723e-06}
	rangeRate, err := ComputeDopplerRangeRate([3]float64{7000, -1210, 1300}, [3]float64{0.2, 7.2, 1}, GroundStation{LatitudeDeg: 51.5, LongitudeDeg: -0.1, AltitudeKm: 0.08}, timescales)
	if err != nil || !(math.Abs(rangeRate.RangeRateKmS-wantRangeRate.RangeRateKmS) <= 1e-12) || !(math.Abs(rangeRate.DopplerRatio-wantRangeRate.DopplerRatio) <= 1e-18) {
		t.Fatalf("ComputeDopplerRangeRate = %#v, %v", rangeRate, err)
	}
	wantShift := DopplerShift{RangeRateKmS: 1.8940111566276965, DopplerHz: -9953.095805947212, DopplerRatio: -6.31774117755723e-06}
	shift, err := ComputeDopplerShift([3]float64{7000, -1210, 1300}, [3]float64{0.2, 7.2, 1}, GroundStation{LatitudeDeg: 51.5, LongitudeDeg: -0.1, AltitudeKm: 0.08}, timescales, 1575.42e6)
	if err != nil || !(math.Abs(shift.RangeRateKmS-wantShift.RangeRateKmS) <= 1e-12) || !(math.Abs(shift.DopplerHz-wantShift.DopplerHz) <= 1e-9) || !(math.Abs(shift.DopplerRatio-wantShift.DopplerRatio) <= 1e-18) {
		t.Fatalf("ComputeDopplerShift = %#v, %v", shift, err)
	}
	satellite := [3]float64{7000, 0, 0}
	sun := [3]float64{1.5e8, 0, 0}
	if status, err := EclipseStatusFor(satellite, sun); err != nil || status != EclipseSunlit {
		t.Fatalf("EclipseStatusFor = %v, %v", status, err)
	}
	legacy, err := EclipseShadowFraction(satellite, sun)
	if err != nil {
		t.Fatal(err)
	}
	spherical, err := EclipseShadowFractionWithModel(satellite, sun, EarthShadowSpherical)
	if err != nil || legacy != spherical {
		t.Fatalf("eclipse fractions = %.17g, %.17g, %v", legacy, spherical, err)
	}
	if radius, err := EarthAngularRadius(satellite); err != nil || !(math.Abs(radius-65.66648805589493) <= 1e-12) {
		t.Fatalf("EarthAngularRadius = %.17g, %v", radius, err)
	}
	if separation, err := AngularSeparationCoords(0, 0, 90, 0); err != nil || separation != 90 {
		t.Fatalf("AngularSeparationCoords = %.17g, %v", separation, err)
	}
	if separation, err := AngularSeparation([3]float64{1, 0, 0}, [3]float64{0, 1, 0}); err != nil || separation != 90 {
		t.Fatalf("AngularSeparation = %.17g, %v", separation, err)
	}
}

func TestValueRoutesSmokeInvalidInput(t *testing.T) {
	_, err := LineOfSightFromAzEl(0, 91, Geodetic{LatitudeRad: 0.6, LongitudeRad: 0.1})
	if err == nil {
		t.Fatal("expected invalid elevation error")
	}
	var statusErr *StatusError
	if !errors.As(err, &statusErr) || statusErr.Code != StatusInvalidArgument || statusErr.Detail == "" {
		t.Fatalf("invalid elevation error = %T, %v", err, err)
	}
	if _, err := ComputeDOP(nil, nil, Geodetic{}); err == nil {
		t.Fatal("expected empty DOP error")
	}
	if _, err := Frequency(GNSSSystem(999), CarrierBandL1); err == nil {
		t.Fatal("expected invalid frequency error")
	}
	if _, err := TimeScaleOffset(GPST, UTC); err == nil {
		t.Fatal("expected epoch-dependent offset error")
	}
}

func TestValueRoutesParallel(t *testing.T) {
	for i := 0; i < 16; i++ {
		i := i
		t.Run(string(rune('a'+i)), func(t *testing.T) {
			t.Parallel()
			value, err := CivilToJ2000Seconds(CivilDateTime{Year: 2020, Month: 6, Day: 25, Hour: i})
			if err != nil || value != 646315200+float64(i*3600) {
				t.Fatalf("CivilToJ2000Seconds = %.17g, %v", value, err)
			}
		})
	}
}

func sameTimeScales(got, want TimeScales) bool {
	return got.JDWhole == want.JDWhole &&
		math.Abs(got.UT1Fraction-want.UT1Fraction) <= 1e-15 &&
		math.Abs(got.TTFraction-want.TTFraction) <= 1e-15 &&
		math.Abs(got.TDBFraction-want.TDBFraction) <= 1e-15 &&
		math.Abs(got.JDUT1-want.JDUT1) <= 1e-9 &&
		math.Abs(got.JDTT-want.JDTT) <= 1e-9 &&
		math.Abs(got.JDTDB-want.JDTDB) <= 1e-9
}

func sameMatrix(got, want Matrix3, tolerance float64) bool {
	for i := range got {
		for j := range got[i] {
			if !(math.Abs(got[i][j]-want[i][j]) <= tolerance) {
				return false
			}
		}
	}
	return true
}

func sameVector(got, want [3]float64, tolerance float64) bool {
	for i := range got {
		if !(math.Abs(got[i]-want[i]) <= tolerance) {
			return false
		}
	}
	return true
}
