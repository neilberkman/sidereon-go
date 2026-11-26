package sidereon

import (
	"slices"

	"sidereon.dev/go/v4/internal/native"
)

// Correlation returns the correlation.
func (fit *OrbitParameterFit) Correlation(first string, second string) (OptionalOrbitScalar, error) {
	value, err := native.OrbitParameterFitCorrelation(nativeOrbitParameterFitHandle(fit), first, second)
	return publicOptionalOrbitScalar(value), publicError(err)
}

// Covariance returns the covariance.
func (fit *OrbitParameterFit) Covariance() (OrbitParameterCovariance, error) {
	value, err := native.OrbitParameterFitCovariance(nativeOrbitParameterFitHandle(fit))
	return publicOrbitParameterCovariance(value), publicError(err)
}

// CovarianceMatrix returns the covariance matrix.
func (fit *OrbitParameterFit) CovarianceMatrix(correlation bool) ([]float64, error) {
	value, err := native.OrbitParameterFitCovarianceMatrix(nativeOrbitParameterFitHandle(fit), correlation)
	return value, publicError(err)
}

// Equal compares the underlying values.
func (left *OrbitParameterFit) Equal(right *OrbitParameterFit) (bool, error) {
	value, err := native.OrbitParameterFitEqual(nativeOrbitParameterFitHandle(left), nativeOrbitParameterFitHandle(right))
	return value, publicError(err)
}

// ForceParameter returns the named estimate, or the zero value if the name was
// not fitted. Use Sigma to read its formal standard deviation.
func (fit *OrbitParameterFit) ForceParameter(name string) (EstimatedForceParameter, error) {
	value, err := native.OrbitParameterFitForceParameter(nativeOrbitParameterFitHandle(fit), name)
	return publicEstimatedForceParameter(value), publicError(err)
}

// ParameterName returns the parameter name.
func (fit *OrbitParameterFit) ParameterName(index uint64) (string, error) {
	value, err := native.OrbitParameterFitParameterName(nativeOrbitParameterFitHandle(fit), index)
	return value, publicError(err)
}

// Residuals returns the residuals.
func (fit *OrbitParameterFit) Residuals() ([]OrbitRTNResidual, error) {
	value, err := native.OrbitParameterFitResiduals(nativeOrbitParameterFitHandle(fit))
	return orbitMap(value, func(value native.OrbitRTNResidual) OrbitRTNResidual { return publicOrbitRTNResidual(value) }), publicError(err)
}

// Sigma returns the named parameter's formal standard deviation from the scaled
// covariance. The value is absent for an unknown name or unbounded covariance.
func (fit *OrbitParameterFit) Sigma(name string) (OptionalOrbitScalar, error) {
	value, err := native.OrbitParameterFitSigma(nativeOrbitParameterFitHandle(fit), name)
	return publicOptionalOrbitScalar(value), publicError(err)
}

// Summary returns the summary.
func (fit *OrbitParameterFit) Summary() (OrbitParameterFitSummary, error) {
	value, err := native.ReadOrbitParameterFitSummary(nativeOrbitParameterFitHandle(fit))
	return publicOrbitParameterFitSummary(value), publicError(err)
}

// StateParameterName returns the state parameter name.
func StateParameterName(index uint64) (string, error) {
	value, err := native.StateParameterName(index)
	return value, publicError(err)
}

// StateParameterNameCount returns the state parameter name count.
func StateParameterNameCount() uint64 {
	return native.StateParameterNameCount()
}

// InertialPositionSample is a GCRF position observation at a tagged instant.
type InertialPositionSample struct {
	// Epoch is the observation instant.
	Epoch ClockEpoch
	// PositionGCRFM is the observed position in metres.
	PositionGCRFM [3]float64
}

// OrbitParameterFitOptions controls estimation of the initial state and force parameters.
// Use Equal to compare values.
type OrbitParameterFitOptions struct {
	// ForceModel selects the built-in forces.
	ForceModel PropagationForceModel
	// ForceComponents configures the built-in forces.
	ForceComponents ForceModelComponents
	// GravityTideSystem selects the Earth gravity tide convention.
	GravityTideSystem GravityTideSystem
	// MuKm3S2 overrides the central gravitational parameter in km³/s² when present.
	MuKm3S2 *float64
	// Drag configures the atmospheric drag force when present.
	Drag *DragParameters
	// SpaceWeather supplies time-dependent atmospheric model inputs when drag is enabled.
	SpaceWeather *SpaceWeatherTable
	// AdditionalForces adds concrete forces in evaluation order.
	AdditionalForces []Force
	// EstimatedParameters lists the force parameter names in estimation order.
	EstimatedParameters []string
	// IntegratorOptions controls numerical integration.
	IntegratorOptions IntegratorOptions
	// PropagationContext supplies ephemeris and Earth orientation data during integration.
	PropagationContext *PropagationContext
	// ObservationSigmaM is the a priori standard deviation of one position component in metres.
	ObservationSigmaM float64
	// MaxIterations limits the Gauss–Newton iterations.
	MaxIterations uint64
	// ConvergenceThreshold is the stopping threshold in formal standard deviations.
	ConvergenceThreshold float64
	// MaxConditionNumber limits the normal equation condition number.
	MaxConditionNumber float64
	// MinLedgerSamples is the sample threshold for marking a residual ledger entry low-sample.
	MinLedgerSamples uint64
}

// Equal compares fit settings, including ordered parameter names and forces.
// Ephemerides, Earth orientation, space-weather tables and additional forces
// are compared by handle identity.
func (value OrbitParameterFitOptions) Equal(other OrbitParameterFitOptions) bool {
	return value.ForceModel == other.ForceModel &&
		value.ForceComponents.Equal(other.ForceComponents) &&
		value.GravityTideSystem == other.GravityTideSystem &&
		equalPointerValue(value.MuKm3S2, other.MuKm3S2) &&
		equalPointerValue(value.Drag, other.Drag) &&
		nativeSpaceWeatherTableHandle(value.SpaceWeather) == nativeSpaceWeatherTableHandle(other.SpaceWeather) &&
		slices.EqualFunc(value.AdditionalForces, other.AdditionalForces, func(left, right Force) bool {
			return nativeForceHandle(left) == nativeForceHandle(right)
		}) &&
		slices.Equal(value.EstimatedParameters, other.EstimatedParameters) &&
		value.IntegratorOptions == other.IntegratorOptions &&
		equalPropagationContext(value.PropagationContext, other.PropagationContext) &&
		value.ObservationSigmaM == other.ObservationSigmaM &&
		value.MaxIterations == other.MaxIterations &&
		value.ConvergenceThreshold == other.ConvergenceThreshold &&
		value.MaxConditionNumber == other.MaxConditionNumber &&
		value.MinLedgerSamples == other.MinLedgerSamples
}

func equalPointerValue[T comparable](value, other *T) bool {
	if value == nil || other == nil {
		return value == other
	}
	return *value == *other
}

func equalPropagationContext(value, other *PropagationContext) bool {
	left, right := PropagationContext{}, PropagationContext{}
	if value != nil {
		left = *value
	}
	if other != nil {
		right = *other
	}
	return nativeDEEphemerisHandle(left.DEEphemeris) == nativeDEEphemerisHandle(right.DEEphemeris) &&
		nativeIERS2010EarthOrientationProviderHandle(left.EarthOrientation) == nativeIERS2010EarthOrientationProviderHandle(right.EarthOrientation)
}

// DefaultOrbitParameterFitOptions returns the default initial-state estimation settings.
func DefaultOrbitParameterFitOptions() OrbitParameterFitOptions {
	value := native.OrbitParameterFitOptionsDefault()
	return OrbitParameterFitOptions{
		ForceModel: PropagationForceModel(value.ForceModel), ForceComponents: publicForceComponents(value.ForceComponents),
		GravityTideSystem: GravityTideSystem(value.GravityTideSystem),
		IntegratorOptions: publicIntegratorOptions(value.IntegratorOptions), ObservationSigmaM: value.ObservationSigmaM,
		MaxIterations: value.MaxIterations, ConvergenceThreshold: value.ConvergenceThreshold,
		MaxConditionNumber: value.MaxConditionNumber, MinLedgerSamples: value.MinLedgerSamples,
	}
}

func nativeOrbitParameterFitOptions(value OrbitParameterFitOptions) native.OrbitParameterFitOptions {
	return native.OrbitParameterFitOptions{
		ForceModel: uint32(value.ForceModel), ForceComponents: nativeForceComponents(value.ForceComponents),
		GravityTideSystem:      uint32(value.GravityTideSystem),
		SolarShadow:            nativeConfiguredSolarShadow(value.ForceComponents.SolarRadiationPressure),
		EarthRadiationPressure: nativeConfiguredEarthRadiation(value.ForceComponents),
		MuKm3S2:                value.MuKm3S2, Drag: orbitOptional(value.Drag, nativeDragParameters),
		SpaceWeather:        nativeSpaceWeatherTableHandle(value.SpaceWeather),
		AdditionalForces:    orbitMap(value.AdditionalForces, nativeForceHandle),
		EstimatedParameters: value.EstimatedParameters, IntegratorOptions: nativeIntegratorOptions(value.IntegratorOptions),
		Context: nativePropagationContext(value.PropagationContext), ObservationSigmaM: value.ObservationSigmaM,
		MaxIterations: value.MaxIterations, ConvergenceThreshold: value.ConvergenceThreshold,
		MaxConditionNumber: value.MaxConditionNumber, MinLedgerSamples: value.MinLedgerSamples,
	}
}

func nativeInertialPositionSample(value InertialPositionSample) native.InertialPositionSample {
	return native.InertialPositionSample{Epoch: nativeClockEpoch(value.Epoch), PositionGCRFM: value.PositionGCRFM}
}

// FitInertialOrbitParameters estimates the initial state and selected force parameters from GCRF positions.
func FitInertialOrbitParameters(samples []InertialPositionSample, satellite string, initialState *CartesianState, options OrbitParameterFitOptions) (*OrbitParameterFit, error) {
	return fitInertialOrbitParameters(samples, satellite, initialState, options, nil)
}

// FitInertialOrbitParametersWithValidity estimates from GCRF positions with the specified UT1 coverage behavior.
func FitInertialOrbitParametersWithValidity(samples []InertialPositionSample, satellite string, initialState *CartesianState, options OrbitParameterFitOptions, permissive bool) (*OrbitParameterFit, error) {
	return fitInertialOrbitParameters(samples, satellite, initialState, options, &permissive)
}

func fitInertialOrbitParameters(samples []InertialPositionSample, satellite string, initialState *CartesianState, options OrbitParameterFitOptions, permissive *bool) (*OrbitParameterFit, error) {
	value, err := native.FitInertialOrbitParameters(orbitMap(samples, nativeInertialPositionSample), satellite,
		orbitOptional(initialState, nativeCartesian), nativeOrbitParameterFitOptions(options), permissive)
	if err != nil {
		return nil, publicError(err)
	}
	return publicOrbitParameterFit(value), nil
}

// FitSP3ECEFOrbitParameters estimates the initial state and selected force parameters from SP3 positions.
func FitSP3ECEFOrbitParameters(product *SP3, provider *IERS2010EarthOrientationProvider, satellite string, initialState *CartesianState, options OrbitParameterFitOptions) (*OrbitParameterFit, error) {
	return fitSP3ECEFOrbitParameters(product, provider, satellite, initialState, options, nil)
}

// FitSP3ECEFOrbitParametersWithValidity estimates from SP3 positions with the specified UT1 coverage behavior.
func FitSP3ECEFOrbitParametersWithValidity(product *SP3, provider *IERS2010EarthOrientationProvider, satellite string, initialState *CartesianState, options OrbitParameterFitOptions, permissive bool) (*OrbitParameterFit, error) {
	return fitSP3ECEFOrbitParameters(product, provider, satellite, initialState, options, &permissive)
}

func fitSP3ECEFOrbitParameters(product *SP3, provider *IERS2010EarthOrientationProvider, satellite string, initialState *CartesianState, options OrbitParameterFitOptions, permissive *bool) (*OrbitParameterFit, error) {
	value, err := native.FitSP3ECEFOrbitParameters(nativeSP3Handle(product), nativeIERS2010EarthOrientationProviderHandle(provider), satellite,
		orbitOptional(initialState, nativeCartesian), nativeOrbitParameterFitOptions(options), permissive)
	if err != nil {
		return nil, publicError(err)
	}
	return publicOrbitParameterFit(value), nil
}

// StateParameterNames returns the six initial-state parameter names in covariance order.
func StateParameterNames() ([]string, error) {
	return orbitParameterNames(StateParameterNameCount(), StateParameterName)
}

// ParameterNames returns the fitted parameter names in covariance order.
func (fit *OrbitParameterFit) ParameterNames() ([]string, error) {
	summary, err := fit.Summary()
	if err != nil {
		return nil, err
	}
	return orbitParameterNames(summary.ParameterCount, fit.ParameterName)
}

func orbitParameterNames(count uint64, nameAt func(uint64) (string, error)) ([]string, error) {
	names := make([]string, 0)
	for i := range count {
		name, err := nameAt(i)
		if err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, nil
}

// ForceParameters returns the fitted force parameters in estimation order.
func (fit *OrbitParameterFit) ForceParameters() ([]EstimatedForceParameter, error) {
	summary, err := fit.Summary()
	if err != nil {
		return nil, err
	}
	parameters := make([]EstimatedForceParameter, 0)
	for index := range summary.ForceParameterCount {
		value, err := native.OrbitParameterFitForceParameterAt(nativeOrbitParameterFitHandle(fit), index)
		if err != nil {
			return nil, publicError(err)
		}
		parameters = append(parameters, publicEstimatedForceParameter(value))
	}
	return parameters, nil
}
