package sidereon

import (
	"errors"
	"math"
	"testing"
	"time"
)

func assertTrajectory(t *testing.T, name string, trajectory *OrbitAttitudeTrajectory, law *GNSSAttitudeLaw) AttitudeSample {
	t.Helper()
	pin, ok := readOrbitalFixture(t).Attitude[name]
	if !ok {
		t.Fatalf("missing attitude fixture %q", name)
	}
	sample := orbitalResult[AttitudeSample](t)(trajectory.Sample(pin.Epoch))
	orbitalEqual(t, pin.Sample)(sample, nil)
	orbitalEqual(t, pin.Attitude)(GnssAttitude(law, trajectory, pin.Epoch))
	return sample
}

func TestOrbitAttitudeSP3(t *testing.T) {
	law := orbitalResource[*GNSSAttitudeLaw](t)(GNSSAttitudeLawOrbitNormal())
	product := orbitalResource[*SP3](t)(LoadSP3(readPositioningFixture(t, "rtk/GBM0MGXRAP_20201770000_01D_05M_ORB_120epoch.sp3")))
	predicted := orbitalResource[*PredictedSP3](t)(PredictedSP3WithCutoff(product, orbitalClockEpoch(GPST, 0)))
	predictedTrajectory := orbitalResource[*OrbitAttitudeTrajectory](t)(OrbitAttitudeTrajectoryFromOrbitSource(predicted, "G03", nil))

	eop := IERS2010EopSample{
		MJDUTC: 59025, XpArcsec: 0.1, YpArcsec: 0.2, Ut1MinusUTCS: -0.2,
		LodS: 0.001, DxArcsec: 0.0001, DyArcsec: -0.0001, HasPoleRates: true,
	}
	next := eop
	next.MJDUTC++
	orientation := orbitalResource[*IERS2010EarthOrientationProvider](t)(NewIERS2010EarthOrientationProvider([]IERS2010EopSample{eop, next}))
	trajectory := orbitalResource[*OrbitAttitudeTrajectory](t)(OrbitAttitudeTrajectoryFromSP3(product, "G03", orientation))
	for _, close := range []func() error{product.Close, orientation.Close, predicted.Close} {
		if err := close(); err != nil {
			t.Fatal(err)
		}
	}

	if err := trajectory.WithVelocityHalfStepS(10); err != nil {
		t.Fatal(err)
	}
	requireOrbitalError[*OrbitScoreError](t, trajectory.WithVelocityHalfStepS(math.NaN()), StatusInvalidArgument, "invalid_input",
		`{"field":"velocity_half_step_s","reason":"not positive and finite"}`)
	requireOrbitalError[*OrbitScoreError](t, trajectory.WithVelocityHalfStepS(1e-9), StatusInvalidArgument, "invalid_input",
		`{"field":"velocity_half_step_s","reason":"below the 1e-8 s resolution of SP3 epochs"}`)
	assertTrajectory(t, "SP3", trajectory, law)
	assertTrajectory(t, "PREDICTED", predictedTrajectory, law)
}

func TestOrbitAttitudeBroadcast(t *testing.T) {
	law := orbitalResource[*GNSSAttitudeLaw](t)(GNSSAttitudeLawOrbitNormal())
	nav := orbitalResource[*BroadcastEphemeris](t)(ParseBroadcastEphemeris(readPositioningFixture(t, "nav/ESBC00DNK_R_20201770000_01D_MN.rnx")))
	trajectory := orbitalResource[*OrbitAttitudeTrajectory](t)(OrbitAttitudeTrajectoryFromBroadcast(nav, "G03", nil))
	source := orbitalResource[*BroadcastOrbit](t)(NewBroadcastOrbit(nav, orbitalClockEpoch(GPST, 0)))
	if err := nav.Close(); err != nil {
		t.Fatal(err)
	}
	scored := orbitalResource[*OrbitAttitudeTrajectory](t)(OrbitAttitudeTrajectoryFromOrbitSource(source, "G03", nil))
	if err := source.Close(); err != nil {
		t.Fatal(err)
	}
	assertTrajectory(t, "NAV", trajectory, law)
	assertTrajectory(t, "NAV", scored, law)
}

func TestOrbitAttitudeBroadcastPreference(t *testing.T) {
	law := orbitalResource[*GNSSAttitudeLaw](t)(GNSSAttitudeLawOrbitNormal())
	nav := orbitalResource[*BroadcastEphemeris](t)(ParseBroadcastEphemeris(readPositioningFixture(t, "nav/BRD400DLR_S_20261800000_01H_MN_trim.rnx")))
	orientation := orbitalResource[*IERS2010EarthOrientationProvider](t)(NewIERS2010EarthOrientationProvider(
		[]IERS2010EopSample{{MJDUTC: 61219}, {MJDUTC: 61221}}))
	const satellite = "J02"
	origin := orbitalClockEpoch(GPST, 0)
	oldSource := orbitalResource[*BroadcastOrbit](t)(NewBroadcastOrbit(nav, origin))
	oldTrajectory := orbitalResource[*OrbitAttitudeTrajectory](t)(OrbitAttitudeTrajectoryFromBroadcast(nav, satellite, orientation))
	orbitalEqual(t, uint32(0))(nav.NavMessagePreference())
	if err := nav.SetNavMessagePreference(1); err != nil {
		t.Fatal(err)
	}
	orbitalEqual(t, uint32(1))(nav.NavMessagePreference())
	newSource := orbitalResource[*BroadcastOrbit](t)(NewBroadcastOrbit(nav, origin))
	newTrajectory := orbitalResource[*OrbitAttitudeTrajectory](t)(OrbitAttitudeTrajectoryFromBroadcast(nav, satellite, orientation))
	if err := nav.Close(); err != nil {
		t.Fatal(err)
	}

	var samples [2]AttitudeSample
	for index, test := range []struct {
		name       string
		source     *BroadcastOrbit
		trajectory *OrbitAttitudeTrajectory
	}{
		{"NAV_LNAV", oldSource, oldTrajectory},
		{"NAV_CNAV", newSource, newTrajectory},
	} {
		scored := orbitalResource[*OrbitAttitudeTrajectory](t)(OrbitAttitudeTrajectoryFromOrbitSource(test.source, satellite, orientation))
		if err := test.source.Close(); err != nil {
			t.Fatal(err)
		}
		assertTrajectory(t, test.name, test.trajectory, law)
		samples[index] = assertTrajectory(t, test.name, scored, law)
	}
	if samples[0].PositionM == samples[1].PositionM {
		t.Fatal("LNAV and CNAV selected the same position")
	}
}

func TestOrbitAttitudeConcurrentFitAndClose(t *testing.T) {
	data := readPositioningFixture(t, "orbitfit-5epoch.sp3")
	options := DefaultOrbitParameterFitOptions()
	options.ForceModel = PropagationForceModelTwoBody
	for range 16 {
		product := orbitalResult[*SP3](t)(LoadSP3(data))
		provider := orbitalResult[*IERS2010EarthOrientationProvider](t)(NewIERS2010EarthOrientationProvider(
			[]IERS2010EopSample{{MJDUTC: 59023}, {MJDUTC: 59025}}))
		build := func() error {
			value, err := OrbitAttitudeTrajectoryFromSP3(product, "G08", provider)
			if err != nil {
				return err
			}
			return value.Close()
		}
		fit := func() error {
			value, err := FitSP3ECEFOrbitParameters(product, provider, "G08", nil, options)
			if err != nil {
				return err
			}
			return value.Close()
		}
		if err := build(); err != nil {
			t.Fatal(err)
		}
		if err := fit(); err != nil {
			t.Fatal(err)
		}

		start := make(chan struct{})
		ready := make(chan struct{}, 4)
		done := make(chan error, 6)
		for _, call := range []func() error{build, fit, build, fit} {
			go func() {
				<-start
				ready <- struct{}{}
				for range 8 {
					if err := call(); err != nil {
						if errors.Is(err, ErrClosed) {
							err = nil
						}
						done <- err
						return
					}
				}
				done <- nil
			}()
		}
		close(start)
		for range 4 {
			<-ready
		}
		go func() { done <- product.Close() }()
		go func() { done <- provider.Close() }()

		timeout := time.NewTimer(5 * time.Second)
		for range 6 {
			select {
			case err := <-done:
				if err != nil {
					t.Error(err)
				}
			case <-timeout.C:
				t.Fatal("trajectory construction, fit or close did not finish")
			}
		}
		timeout.Stop()
		if err := build(); !errors.Is(err, ErrClosed) {
			t.Fatalf("trajectory construction after close = %v", err)
		}
		if err := fit(); !errors.Is(err, ErrClosed) {
			t.Fatalf("orbit fit after close = %v", err)
		}
	}
}

func TestOrbitAttitudeSampled(t *testing.T) {
	law := orbitalResource[*GNSSAttitudeLaw](t)(GNSSAttitudeLawOrbitNormal())
	samples := []TrajectorySample{
		{Satellite: "G03", Epoch: orbitalClockEpoch(GPST, 646315199), PositionECEFM: [3]float64{26000000, 0, 1000000}},
		{Satellite: "G03", Epoch: orbitalClockEpoch(GPST, 646315200), PositionECEFM: [3]float64{26000000, 100, 1000000}},
		{Satellite: "G03", Epoch: orbitalClockEpoch(GPST, 646315201), PositionECEFM: [3]float64{26000000, 200, 1000000}},
	}
	source := orbitalResource[*OrbitTrajectory](t)(NewOrbitTrajectory("IGS14", orbitalClockEpoch(GPST, 0), samples))
	trajectory := orbitalResource[*OrbitAttitudeTrajectory](t)(OrbitAttitudeTrajectoryFromOrbitSource(source, "G03", nil))
	if err := trajectory.WithVelocityHalfStepS(1); err != nil {
		t.Fatal(err)
	}
	assertTrajectory(t, "SAMPLED", trajectory, law)
}

func ecksteinHechlerAttitude(t *testing.T) *OrbitAttitudeTrajectory {
	t.Helper()
	elements := CircularElements{AKm: 7200, Ex: 0.001, Ey: 0.002, IRad: 0.5, RaanRad: 0.25, MeanLatitudeArgumentRad: 0.125}
	propagator := orbitalResource[*EcksteinHechlerPropagator](t)(EcksteinHechlerPropagatorFromMeanElements(orbitalEpoch, elements, ZonalHarmonicsEarth()))
	trajectory := orbitalResource[*OrbitAttitudeTrajectory](t)(OrbitAttitudeTrajectoryFromEcksteinHechler(propagator))
	if err := propagator.Close(); err != nil {
		t.Fatal(err)
	}
	return trajectory
}

func TestOrbitAttitudeAnalytic(t *testing.T) {
	law := orbitalResource[*GNSSAttitudeLaw](t)(GNSSAttitudeLawOrbitNormal())
	eh := ecksteinHechlerAttitude(t)
	assertTrajectory(t, "EH", eh, law)
	elements := AnalyticKeplerianElements{AKm: 7200, E: 0.01, IRad: 0.5, RaanRad: 0.25, ArgPerigeeRad: 0.75, MeanAnomalyRad: 0.125}
	propagator := orbitalResource[*BrouwerLyddanePropagator](t)(BrouwerLyddanePropagatorFromMeanElements(orbitalEpoch, elements, ZonalHarmonicsEarth(), 0))
	bl := orbitalResource[*OrbitAttitudeTrajectory](t)(OrbitAttitudeTrajectoryFromBrouwerLyddane(propagator))
	if err := propagator.Close(); err != nil {
		t.Fatal(err)
	}
	assertTrajectory(t, "BL", bl, law)
	de := orbitalResource[*DEEphemeris](t)(DEEphemerisFromBytes(readPositioningFixture(t, "orbital/lnxp1990.440")))
	if err := eh.WithDEEphemeris(de); err != nil {
		t.Fatal(err)
	}
	if err := de.Close(); err != nil {
		t.Fatal(err)
	}
	assertTrajectory(t, "DE", eh, law)
	if err := eh.WithDEEphemeris(nil); err != nil {
		t.Fatal(err)
	}
	assertTrajectory(t, "EH", eh, law)
	if err := eh.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := eh.Sample(orbitalEpoch); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed trajectory sample error = %v", err)
	}
}

func TestOrbitAttitudeDense(t *testing.T) {
	normal := orbitalResource[*GNSSAttitudeLaw](t)(GNSSAttitudeLawOrbitNormal())
	config := orbitalResult[PropagationConfig](t)(DefaultPropagationConfig())
	config.EpochTDBSeconds, config.PositionKm, config.VelocityKmPerS = orbitalEpoch, orbitalPosition, orbitalVelocity
	config.ForceModel = PropagationForceModelTwoBody
	end := orbitalEpoch + 600
	state, dense, err := PropagateStateWithDenseOutput(config, end, PropagationEnvironment{}, nil)
	dense = orbitalResource[*DenseOutput](t)(dense, err)
	orbitalEqual(t, readOrbitalFixture(t).AttitudeFinal)(state, nil)
	trajectory := orbitalResource[*OrbitAttitudeTrajectory](t)(OrbitAttitudeTrajectoryFromDenseOutput(dense))
	if err := dense.Close(); err != nil {
		t.Fatal(err)
	}
	assertTrajectory(t, "DENSE", trajectory, normal)

	const unavailable = `{"t_s":{"decimal":"0","bits_hex":"0000000000000000"}}`
	_, err = GnssAttitude(normal, trajectory, 0)
	requireOrbitalError[*GNSSAttitudeError](t, err, StatusInvalidArgument, "trajectory_unavailable", unavailable)
	_, err = trajectory.Sample(0)
	requireOrbitalError[*GNSSAttitudeError](t, err, StatusInvalidArgument, "trajectory_unavailable", unavailable)

	state, dense, err = PropagateStateWithDenseOutput(config, orbitalEpoch, PropagationEnvironment{}, nil)
	dense = orbitalResource[*DenseOutput](t)(dense, err)
	orbitalEqual(t, CartesianState{EpochTDBSeconds: orbitalEpoch, PositionKm: orbitalPosition, VelocityKmPerS: orbitalVelocity})(state, nil)
	empty := orbitalResource[*OrbitAttitudeTrajectory](t)(OrbitAttitudeTrajectoryFromDenseOutput(dense))
	_, err = empty.Sample(orbitalEpoch)
	requireOrbitalError[*GNSSAttitudeError](t, err, StatusInvalidArgument, "trajectory_unavailable",
		`{"t_s":{"decimal":"646315200.25","bits_hex":"41c342fe60200000"}}`)

	config.Integrator = PropagationIntegratorRK4
	state, dense, err = PropagateStateWithDenseOutput(config, end, PropagationEnvironment{}, nil)
	orbitalEqual(t, (*DenseOutput)(nil))(dense, err)
	orbitalEqual(t, end)(state.EpochTDBSeconds, nil)
	_, dense, err = PropagateStateWithDenseOutput(config, math.NaN(), PropagationEnvironment{}, nil)
	requireOrbitalFamily(t, err, StatusInvalidArgument, EngineErrorFamilyPropagation, "invalid_input",
		`{"message":"t_end_tdb_seconds not finite","reason":"t_end_tdb_seconds not finite"}`)
	orbitalEqual(t, (*DenseOutput)(nil))(dense, nil)
}

func TestOrbitAttitudeLawResults(t *testing.T) {
	trajectory := ecksteinHechlerAttitude(t)
	newLaw := orbitalResource[*GNSSAttitudeLaw](t)
	for _, law := range []*GNSSAttitudeLaw{
		newLaw(GNSSAttitudeLawGPSIIA(orbitalResult[float64](t)(GPSIIADefaultYawRateRadS(1)), GPSIIADefaultYawBiasRad(), GPSShadowLimitRad())),
		newLaw(GNSSAttitudeLawGPSIIF(GPSIIFDefaultYawRateRadS(), GPSIIFDefaultYawBiasRad(), GPSShadowLimitRad())),
		newLaw(GNSSAttitudeLawGPSIIR(GPSIIRDefaultYawRateRadS())),
		newLaw(GNSSAttitudeLawGLONASSM(GLONASSMDefaultYawRateRadS(), GLONASSMShadowLimitRad())),
		newLaw(GNSSAttitudeLawGalileoIOV(GalileoIOVBetaXRad(), GalileoIOVBetaYRad())),
		newLaw(GNSSAttitudeLawCosineTurn(GalileoFOCTurn())),
		newLaw(GNSSAttitudeLawBeiDouOrbitNormalSwitch(BeiDou2YsToOnBetaRad(), BeiDou2OnToYsBetaRad())),
		newLaw(GNSSAttitudeLawBeiDou3MEOUnknownManufacturer(BeiDou3CASTMEOTurn(), BeiDou3SECMBeta0Rad())),
	} {
		assertTrajectory(t, "NOMINAL", trajectory, law)
	}
	assertTrajectory(t, "FIXED_BETA", trajectory, newLaw(GNSSAttitudeLawFixedBetaYawSteering(1.5)))
}

func TestOrbitAttitudeLawRefusals(t *testing.T) {
	trajectory := ecksteinHechlerAttitude(t)
	newLaw := orbitalResource[*GNSSAttitudeLaw](t)
	const epoch = `"t_s":{"decimal":"646315800.25","bits_hex":"41c342ff8c200000"}`
	tests := []struct {
		law    *GNSSAttitudeLaw
		kind   string
		detail string
	}{
		{newLaw(GNSSAttitudeLawFixedBetaYawSteering(0)), "invalid_parameter", `{"field":"beta0_rad","value":{"decimal":"0","bits_hex":"0000000000000000"}}`},
		{newLaw(GNSSAttitudeLawBeiDou3MEOUnknownManufacturer(BeiDou3CASTMEOTurn(), 1.5)), "beidou_manufacturer_required", "{" + epoch + "}"},
		{newLaw(GNSSAttitudeLawCosineTurn(CosineTurnParameters{BetaLimitRad: 1e-9, StartAngle: CosineTurnStartAngleOrbitAngle, StartAngleRad: 0.1, PeriodS: 1e7})), "turn_end_not_found", "{" + epoch + "}"},
		{newLaw(GNSSAttitudeLawGPSIIA(1000, 0, 1.6)), "outside_validated_domain", "{" + epoch + `,"reason":"GPS IIA maximum yaw rate above 1e5 times the orbital rate, where the shadow crossing's yaw is set only to the rate times the uncertainty of the shadow entry"}`},
	}
	for _, test := range tests {
		t.Run(test.kind, func(t *testing.T) {
			_, err := GnssAttitude(test.law, trajectory, orbitalEpoch+600)
			requireOrbitalError[*GNSSAttitudeError](t, err, StatusInvalidArgument, test.kind, test.detail)
		})
	}
}
