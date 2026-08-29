package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// SolarRadiationPressure contains the coefficient and area-to-mass ratio used
// by the C solar-radiation-pressure model.
type SolarRadiationPressure struct {
	// CR is the coefficient of reflectivity.
	CR float64
	// AreaToMassM2PerKg is the area to mass m2 per kg in square metres per kilogram.
	AreaToMassM2PerKg float64
}

// ForceModelComponents selects additive perturbations for composite
// propagation. Degree and order values are the native model's limits.
type ForceModelComponents struct {
	// HasTwoBody reports whether the has two body field is present.
	HasTwoBody bool
	// TwoBodyMuKm3PerS2Enabled reports whether MuKm3PerS2 overrides the native gravitational parameter.
	TwoBodyMuKm3PerS2Enabled bool
	// TwoBodyMuKm3PerS2 is the two body mu km3 per s2 in cubic kilometres per second squared.
	TwoBodyMuKm3PerS2 float64
	// HasZonal reports whether the has zonal field is present.
	HasZonal bool
	// ZonalMaxDegree is the maximum zonal-harmonic degree used when HasZonal is true.
	ZonalMaxDegree uint32
	// HasSphericalHarmonic reports whether the has spherical harmonic field is present.
	HasSphericalHarmonic bool
	// SphericalHarmonicMaxDegree is the maximum spherical-harmonic degree when HasSphericalHarmonic is true.
	SphericalHarmonicMaxDegree uint32
	// SphericalHarmonicMaxOrder is the maximum spherical-harmonic order when HasSphericalHarmonic is true.
	SphericalHarmonicMaxOrder uint32
	// HasSolidEarthTide reports whether the has solid earth tide field is present.
	HasSolidEarthTide bool
	// HasSolidEarthPoleTide reports whether the has solid earth pole tide field is present.
	HasSolidEarthPoleTide bool
	// HasThirdBody reports whether the has third body field is present.
	HasThirdBody bool
	// ThirdBodySun reports the status of this record.
	ThirdBodySun bool
	// ThirdBodyMoon reports the status of this record.
	ThirdBodyMoon bool
	// HasSolarRadiationPressure reports whether the has solar radiation pressure field is present.
	HasSolarRadiationPressure bool
	// SolarRadiationPressure contains the coefficient and area-to-mass ratio used when HasSolarRadiationPressure is true.
	SolarRadiationPressure SolarRadiationPressure
	// HasRelativity reports whether the has relativity field is present.
	HasRelativity bool
}

// DragParameters contains validated atmospheric-drag parameters. The
// ballistic factor is C_D*A/m in m²/kg and cutoff altitude is km.
type DragParameters struct {
	// BCFactorM2PerKg is the bc factor m2 per kg in square metres per kilogram.
	BCFactorM2PerKg float64
	// Weather is the space-weather table.
	Weather SpaceWeather
	// CutoffAltitudeKm is the cutoff altitude km in kilometres.
	CutoffAltitudeKm float64
}

func nativeSpaceWeather(value SpaceWeather) native.SpaceWeather {
	return native.SpaceWeather{F107: value.F107, F107A: value.F107A, Ap: value.Ap}
}

func nativeForceComponents(value ForceModelComponents) native.ForceModelComponents {
	return native.ForceModelComponents{
		HasTwoBody: value.HasTwoBody, TwoBodyMuKm3PerS2Enabled: value.TwoBodyMuKm3PerS2Enabled, TwoBodyMuKm3PerS2: value.TwoBodyMuKm3PerS2,
		HasZonal: value.HasZonal, ZonalMaxDegree: value.ZonalMaxDegree,
		HasSphericalHarmonic: value.HasSphericalHarmonic, SphericalHarmonicMaxDegree: value.SphericalHarmonicMaxDegree, SphericalHarmonicMaxOrder: value.SphericalHarmonicMaxOrder,
		HasSolidEarthTide: value.HasSolidEarthTide, HasSolidEarthPoleTide: value.HasSolidEarthPoleTide,
		HasThirdBody: value.HasThirdBody, ThirdBodySun: value.ThirdBodySun, ThirdBodyMoon: value.ThirdBodyMoon,
		HasSolarRadiationPressure: value.HasSolarRadiationPressure, SolarRadiationPressure: native.SolarRadiationPressure{CR: value.SolarRadiationPressure.CR, AreaToMassM2PerKg: value.SolarRadiationPressure.AreaToMassM2PerKg}, HasRelativity: value.HasRelativity,
	}
}

func publicForceComponents(value native.ForceModelComponents) ForceModelComponents {
	return ForceModelComponents{
		HasTwoBody: value.HasTwoBody, TwoBodyMuKm3PerS2Enabled: value.TwoBodyMuKm3PerS2Enabled, TwoBodyMuKm3PerS2: value.TwoBodyMuKm3PerS2,
		HasZonal: value.HasZonal, ZonalMaxDegree: value.ZonalMaxDegree,
		HasSphericalHarmonic: value.HasSphericalHarmonic, SphericalHarmonicMaxDegree: value.SphericalHarmonicMaxDegree, SphericalHarmonicMaxOrder: value.SphericalHarmonicMaxOrder,
		HasSolidEarthTide: value.HasSolidEarthTide, HasSolidEarthPoleTide: value.HasSolidEarthPoleTide,
		HasThirdBody: value.HasThirdBody, ThirdBodySun: value.ThirdBodySun, ThirdBodyMoon: value.ThirdBodyMoon,
		HasSolarRadiationPressure: value.HasSolarRadiationPressure, SolarRadiationPressure: SolarRadiationPressure{CR: value.SolarRadiationPressure.CR, AreaToMassM2PerKg: value.SolarRadiationPressure.AreaToMassM2PerKg}, HasRelativity: value.HasRelativity,
	}
}

func publicDragParameters(value native.DragParameters) DragParameters {
	return DragParameters{BCFactorM2PerKg: value.BCFactorM2PerKg, Weather: SpaceWeather{F107: value.Weather.F107, F107A: value.Weather.F107A, Ap: value.Weather.Ap}, CutoffAltitudeKm: value.CutoffAltitudeKm}
}

// DragParametersFromAreaMass derives C_D*A/m from coefficient, area (m²),
// and mass (kg), entirely through the C validation and model.
func DragParametersFromAreaMass(cd, areaM2, massKg float64, weather SpaceWeather, cutoffAltitudeKm float64) (DragParameters, error) {
	value, err := native.DragParametersFromAreaMass(cd, areaM2, massKg, nativeSpaceWeather(weather), cutoffAltitudeKm)
	return publicDragParameters(value), publicError(err)
}

// DragParametersFromBCFactor validates a directly supplied C_D*A/m factor.
func DragParametersFromBCFactor(factorM2PerKg float64, weather SpaceWeather, cutoffAltitudeKm float64) (DragParameters, error) {
	value, err := native.DragParametersFromBCFactor(factorM2PerKg, nativeSpaceWeather(weather), cutoffAltitudeKm)
	return publicDragParameters(value), publicError(err)
}

// DragParametersFromBallisticCoefficient validates the reciprocal ballistic
// coefficient in kg/m².
func DragParametersFromBallisticCoefficient(bcKgPerM2 float64, weather SpaceWeather, cutoffAltitudeKm float64) (DragParameters, error) {
	value, err := native.DragParametersFromBallisticCoefficient(bcKgPerM2, nativeSpaceWeather(weather), cutoffAltitudeKm)
	return publicDragParameters(value), publicError(err)
}

// DragForceAcceleration returns atmospheric drag acceleration in km/s² for an
// ECI Cartesian state.
func DragForceAcceleration(drag DragParameters, state CartesianState) ([3]float64, error) {
	value, err := native.DragForceAcceleration(native.DragParameters{BCFactorM2PerKg: drag.BCFactorM2PerKg, Weather: nativeSpaceWeather(drag.Weather), CutoffAltitudeKm: drag.CutoffAltitudeKm}, nativeCartesian(state))
	return value, publicError(err)
}

// ForceTwoBodyAcceleration returns central-gravity acceleration in km/s².
func ForceTwoBodyAcceleration(positionKm, velocityKmPerS [3]float64) ([3]float64, error) {
	value, err := native.ForceTwoBodyAcceleration(positionKm, velocityKmPerS)
	return value, publicError(err)
}

// ForceJ2Acceleration returns the J2 perturbation in km/s².
func ForceJ2Acceleration(positionKm, velocityKmPerS [3]float64) ([3]float64, error) {
	value, err := native.ForceJ2Acceleration(positionKm, velocityKmPerS)
	return value, publicError(err)
}

// DecayConfig controls the C drag-perturbed reentry estimate.
type DecayConfig struct {
	// ForceModel is the propagation force model.
	ForceModel PropagationForceModel
	// Integrator is the propagation integrator.
	Integrator PropagationIntegrator
	// AbsTol is the absolute solver tolerance.
	AbsTol float64
	// RelTol is the relative solver tolerance.
	RelTol float64
	// InitialStepS is the initial integration step in seconds.
	InitialStepS float64
	// MinStepS is the minimum integration step in seconds.
	MinStepS float64
	// MaxStepS is the maximum integration step in seconds.
	MaxStepS float64
	// MaxSteps is the maximum solver step count.
	MaxSteps uint32
	// MuKm3PerS2Enabled reports whether MuKm3PerS2 overrides the default gravitational parameter.
	MuKm3PerS2Enabled bool
	// MuKm3PerS2 is the mu km3 per s2 in cubic kilometres per second squared.
	MuKm3PerS2 float64
	// Drag is the drag model.
	Drag DragParameters
	// ReentryAltitudeKm is the reentry altitude km in kilometres.
	ReentryAltitudeKm float64
	// ScanStepS is the interval between re-entry crossing samples in seconds.
	ScanStepS float64
	// CrossingToleranceS is the crossing tolerance s in seconds.
	CrossingToleranceS float64
	// MaxDurationS is the max duration s in seconds.
	MaxDurationS float64
	// MaxScanSamples is the maximum number of samples used to bracket a re-entry crossing.
	MaxScanSamples uint32
}

// DecayEstimate contains the C-computed reentry duration and state.
type DecayEstimate struct {
	// TimeToDecayS is the estimated time to re-entry in seconds.
	TimeToDecayS float64
	// ReentryState is the re-entry state classification.
	ReentryState CartesianState
	// ReentryAltitudeKm is the reentry altitude km in kilometres.
	ReentryAltitudeKm float64
}

func publicDecayConfig(value native.DecayConfig) DecayConfig {
	return DecayConfig{ForceModel: PropagationForceModel(value.ForceModel), Integrator: PropagationIntegrator(value.Integrator), AbsTol: value.AbsTol, RelTol: value.RelTol, InitialStepS: value.InitialStepS, MinStepS: value.MinStepS, MaxStepS: value.MaxStepS, MaxSteps: value.MaxSteps, MuKm3PerS2Enabled: value.MuKm3PerS2Enabled, MuKm3PerS2: value.MuKm3PerS2, Drag: publicDragParameters(value.Drag), ReentryAltitudeKm: value.ReentryAltitudeKm, ScanStepS: value.ScanStepS, CrossingToleranceS: value.CrossingToleranceS, MaxDurationS: value.MaxDurationS, MaxScanSamples: value.MaxScanSamples}
}

func nativeDecayConfig(value DecayConfig) native.DecayConfig {
	return native.DecayConfig{ForceModel: uint32(value.ForceModel), Integrator: uint32(value.Integrator), AbsTol: value.AbsTol, RelTol: value.RelTol, InitialStepS: value.InitialStepS, MinStepS: value.MinStepS, MaxStepS: value.MaxStepS, MaxSteps: value.MaxSteps, MuKm3PerS2Enabled: value.MuKm3PerS2Enabled, MuKm3PerS2: value.MuKm3PerS2, Drag: native.DragParameters{BCFactorM2PerKg: value.Drag.BCFactorM2PerKg, Weather: nativeSpaceWeather(value.Drag.Weather), CutoffAltitudeKm: value.Drag.CutoffAltitudeKm}, ReentryAltitudeKm: value.ReentryAltitudeKm, ScanStepS: value.ScanStepS, CrossingToleranceS: value.CrossingToleranceS, MaxDurationS: value.MaxDurationS, MaxScanSamples: value.MaxScanSamples}
}

// DefaultDecayConfig returns the C engine defaults.
func DefaultDecayConfig() (DecayConfig, error) {
	value, err := native.DecayConfigDefaults()
	return publicDecayConfig(value), publicError(err)
}

// EstimateDecay finds the first C-defined reentry crossing.
func EstimateDecay(initial CartesianState, config DecayConfig) (DecayEstimate, error) {
	value, err := native.EstimateDecay(nativeCartesian(initial), nativeDecayConfig(config))
	return DecayEstimate{TimeToDecayS: value.TimeToDecayS, ReentryState: publicCartesian(value.ReentryState), ReentryAltitudeKm: value.ReentryAltitudeKm}, publicError(err)
}

// EstimateDecayWithSpaceWeatherTable uses a parsed native space-weather table
// for the atmospheric model.
func EstimateDecayWithSpaceWeatherTable(initial CartesianState, config DecayConfig, table *SpaceWeatherTable) (DecayEstimate, error) {
	if table == nil || table.native == nil {
		return DecayEstimate{}, ErrClosed
	}
	value, err := native.EstimateDecayWithSpaceWeatherTable(nativeCartesian(initial), nativeDecayConfig(config), table.native)
	return DecayEstimate{TimeToDecayS: value.TimeToDecayS, ReentryState: publicCartesian(value.ReentryState), ReentryAltitudeKm: value.ReentryAltitudeKm}, publicError(err)
}
