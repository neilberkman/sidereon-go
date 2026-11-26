package sidereon

import "sidereon.dev/go/v4/internal/native"

// PropagationEnvironment supplies Earth orientation, ephemeris and weather data for built-in forces.
type PropagationEnvironment struct {
	// GravityTideSystem selects the Earth gravity tide convention.
	GravityTideSystem GravityTideSystem
	// Context supplies ephemeris and Earth orientation data.
	Context *PropagationContext
	// SpaceWeather supplies a time-dependent weather table and requires drag.
	// Nil uses the fixed weather in DragParameters.
	SpaceWeather *SpaceWeatherTable
}

func nativePropagationEnvironment(forces ForceModelComponents, environment PropagationEnvironment) native.PropagationEnvironment {
	return native.PropagationEnvironment{
		GravityTideSystem:      uint32(environment.GravityTideSystem),
		SolarShadow:            nativeConfiguredSolarShadow(forces.SolarRadiationPressure),
		EarthRadiationPressure: nativeConfiguredEarthRadiation(forces),
		Context:                nativePropagationContext(environment.Context),
		SpaceWeather:           nativeSpaceWeatherTableHandle(environment.SpaceWeather),
	}
}

func hasRadiationOptions(components ForceModelComponents) bool {
	return components.SolarRadiationPressure.HasShadow || components.HasEarthRadiationPressure
}

func nativeConfiguredSolarShadow(options SolarRadiationPressure) *native.RadiationShadowOptions {
	if !options.HasShadow {
		return nil
	}
	value := nativeRadiationShadowOptions(options.Shadow)
	return &value
}

func nativeConfiguredEarthRadiation(components ForceModelComponents) *native.EarthRadiationPressureOptions {
	if !components.HasEarthRadiationPressure {
		return nil
	}
	value := nativeEarthRadiationPressureOptions(components.EarthRadiationPressure)
	return &value
}

// PropagateStateWithForceModel computes a GCRF ephemeris at absolute TDB
// seconds past J2000. It evaluates the configured forces and drag, then
// sums additionalForces in slice order. Pass nil to use only the configured
// forces and drag.
func PropagateStateWithForceModel(config PropagationConfig, epochs []float64, environment PropagationEnvironment, additionalForces []Force) (*Ephemeris, error) {
	value, err := native.PropagateStateWithForceModel(nativePropagationConfig(config), epochs, nativePropagationEnvironment(config.ForceComponents, environment), orbitMap(additionalForces, nativeForceHandle))
	if err != nil {
		return nil, publicError(err)
	}
	return &Ephemeris{handle: value}, nil
}

// PropagateCovarianceWithContext propagates the state and covariance at the requested TDB epochs.
func PropagateCovarianceWithContext(config PropagationConfig, covariance Covariance6, epochs []float64, inputFrame, outputFrame CovarianceFrame, noise ProcessNoise, environment PropagationEnvironment) (*CovarianceEphemeris, error) {
	value, err := native.PropagateCovarianceWithContext(nativePropagationConfig(config), covariance.Values, epochs,
		uint32(inputFrame), uint32(outputFrame), nativeProcessNoise(noise), nativePropagationEnvironment(config.ForceComponents, environment))
	if err != nil {
		return nil, publicError(err)
	}
	return &CovarianceEphemeris{handle: value}, nil
}

// StateWithCovariance contains an inertial Cartesian state and its covariance.
// Values can be compared with ==.
type StateWithCovariance struct {
	// State is the propagated Cartesian state.
	State CartesianState
	// Covariance is the propagated inertial covariance.
	Covariance Covariance6
}

// PropagateStateWithCovarianceWithContext propagates an inertial state and covariance by a TDB duration in seconds.
func PropagateStateWithCovarianceWithContext(config PropagationConfig, covariance Covariance6, spanSeconds float64, environment PropagationEnvironment) (StateWithCovariance, error) {
	value, err := native.PropagateStateWithCovarianceWithContext(nativePropagationConfig(config), covariance.Values, spanSeconds, nativePropagationEnvironment(config.ForceComponents, environment))
	return StateWithCovariance{State: publicCartesian(value.State), Covariance: Covariance6{Values: value.Covariance}}, publicError(err)
}

// StateTransitionMatrixToWithContext returns the numerical state transition matrix at an absolute TDB epoch.
func StateTransitionMatrixToWithContext(config PropagationConfig, epochTDBSeconds float64, environment PropagationEnvironment) ([6][6]float64, error) {
	value, err := native.StateTransitionMatrixToWithContext(nativePropagationConfig(config), epochTDBSeconds, nativePropagationEnvironment(config.ForceComponents, environment))
	return value, publicError(err)
}

// EstimateDecayWithContext estimates the first reentry crossing with the supplied propagation context.
func EstimateDecayWithContext(initial CartesianState, config DecayConfig, context *PropagationContext) (DecayEstimate, error) {
	return estimateDecayWithContext(initial, config, nil, context)
}

// EstimateDecayWithSourceAndContext estimates the first reentry crossing with time-dependent space weather.
func EstimateDecayWithSourceAndContext(initial CartesianState, config DecayConfig, table *SpaceWeatherTable, context *PropagationContext) (DecayEstimate, error) {
	if table == nil || table.native == nil {
		return DecayEstimate{}, ErrClosed
	}
	return estimateDecayWithContext(initial, config, table, context)
}

func estimateDecayWithContext(initial CartesianState, config DecayConfig, table *SpaceWeatherTable, context *PropagationContext) (DecayEstimate, error) {
	value, err := native.EstimateDecayWithContext(nativeCartesian(initial), nativeDecayConfig(config), nativeSpaceWeatherTableHandle(table), nativePropagationContext(context))
	return DecayEstimate{TimeToDecayS: value.TimeToDecayS, ReentryState: publicCartesian(value.ReentryState), ReentryAltitudeKm: value.ReentryAltitudeKm}, publicError(err)
}

// PropagateStateWithDenseOutput propagates a GCRF state to an absolute epoch
// in TDB seconds past J2000 and requests continuous DP5(4) output. Close the
// dense output when done. RK4 returns nil output. A zero-duration DP5(4) arc
// returns an empty output. It sums additionalForces in slice order after
// the configured forces and drag.
func PropagateStateWithDenseOutput(config PropagationConfig, targetTDBSeconds float64, environment PropagationEnvironment, additionalForces []Force) (CartesianState, *DenseOutput, error) {
	state, dense, err := native.PropagateStateWithDenseOutput(nativePropagationConfig(config), targetTDBSeconds, nativePropagationEnvironment(config.ForceComponents, environment), orbitMap(additionalForces, nativeForceHandle))
	if err != nil || dense == nil {
		return publicCartesian(state), nil, publicError(err)
	}
	return publicCartesian(state), &DenseOutput{handle: dense}, nil
}
