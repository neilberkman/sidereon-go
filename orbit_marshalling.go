package sidereon

import "sidereon.dev/go/v4/internal/native"

func orbitMap[S, T any](values []S, convert func(S) T) []T {
	if values == nil {
		return nil
	}
	result := make([]T, len(values))
	for i, value := range values {
		result[i] = convert(value)
	}
	return result
}

func orbitOptional[S, T any](value *S, convert func(S) T) *T {
	if value == nil {
		return nil
	}
	result := convert(*value)
	return &result
}

func nativeBroadcastEphemerisHandle(value *BroadcastEphemeris) *native.BroadcastEphemeris {
	if value == nil {
		return nil
	}
	return value.handle
}

func nativeSpaceWeatherTableHandle(value *SpaceWeatherTable) *native.SpaceWeatherTable {
	if value == nil {
		return nil
	}
	if value.native == nil {
		return &native.SpaceWeatherTable{}
	}
	return value.native
}

func publicSP3Handle(value *native.SP3) *SP3 {
	return &SP3{handle: value}
}

func publicTimeScales(value native.TimeScales) TimeScales { return TimeScales(value) }

func nativeDragParameters(value DragParameters) native.DragParameters {
	return native.DragParameters{
		BCFactorM2PerKg: value.BCFactorM2PerKg, Weather: nativeSpaceWeather(value.Weather),
		CutoffAltitudeKm: value.CutoffAltitudeKm,
	}
}

// PropagationContext selects optional planetary ephemeris and Earth orientation data.
type PropagationContext struct {
	// DEEphemeris supplies Sun and Moon positions. A nil value uses the analytic series.
	DEEphemeris *DEEphemeris
	// EarthOrientation supplies Earth orientation. A nil value uses the operation's default.
	EarthOrientation *IERS2010EarthOrientationProvider
}

func nativePropagationContext(value *PropagationContext) *native.PropagationContext {
	if value == nil {
		return nil
	}
	return &native.PropagationContext{
		DEEphemeris:      nativeDEEphemerisHandle(value.DEEphemeris),
		EarthOrientation: nativeIERS2010EarthOrientationProviderHandle(value.EarthOrientation),
	}
}

// Force is a model accepted by state and variational propagation.
type Force interface {
	forceModel() *ForceModel
}

func (model *ForceModel) forceModel() *ForceModel { return model }

func nativeForceHandle(model Force) *native.ForceModel {
	if model == nil {
		return nil
	}
	return nativeForceModelValue(model.forceModel())
}

func convertOrbitArrayRTNAxis(value [3]native.RTNAxis) [3]RTNAxis {
	return [3]RTNAxis{RTNAxis(value[0]), RTNAxis(value[1]), RTNAxis(value[2])}
}

func convertOrbitArrayEmpiricalHarmonic(value [3]native.EmpiricalHarmonic) [3]EmpiricalHarmonic {
	return [3]EmpiricalHarmonic{EmpiricalHarmonic(value[0]), EmpiricalHarmonic(value[1]), EmpiricalHarmonic(value[2])}
}

// WithDEEphemeris returns a context with the supplied planetary ephemeris.
func (context PropagationContext) WithDEEphemeris(ephemeris *DEEphemeris) PropagationContext {
	context.DEEphemeris = ephemeris
	return context
}
