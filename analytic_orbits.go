package sidereon

import (
	"sidereon.dev/go/v4/internal/native"
)

// AnalyticKeplerianElementsFromCartesian converts an inertial Cartesian state to osculating Keplerian elements.
func AnalyticKeplerianElementsFromCartesian(state AnalyticCartesian, muKm3S2 float64) (AnalyticKeplerianElements, error) {
	value, err := native.AnalyticKeplerianElementsFromCartesian(nativeAnalyticCartesian(state), muKm3S2)
	return publicAnalyticKeplerianElements(value), publicError(err)
}

// ToCartesian returns the two-body inertial position and velocity in km and km/s.
func (elements AnalyticKeplerianElements) ToCartesian(muKm3S2 float64) (AnalyticCartesian, error) {
	value, err := native.AnalyticKeplerianElementsToCartesian(nativeAnalyticKeplerianElements(elements), muKm3S2)
	return publicAnalyticCartesian(value), publicError(err)
}

// AnalyticTheoryText returns the text.
func AnalyticTheoryText(theory AnalyticTheory) (string, error) {
	value, err := native.AnalyticTheoryText(native.AnalyticTheory(theory))
	return value, publicError(err)
}

// BrouwerLyddaneMeanElements converts an osculating state to Brouwer–Lyddane mean elements.
func BrouwerLyddaneMeanElements(state AnalyticCartesian, gravity ZonalHarmonics, options MeanConversionOptions) (MeanElements[AnalyticKeplerianElements], error) {
	value, err := native.ComputeBrouwerLyddaneMeanElements(nativeAnalyticCartesian(state), nativeZonalHarmonics(gravity), nativeMeanConversionOptions(options))
	return publicBrouwerLyddaneMeanElements(value), publicError(err)
}

// BrouwerLyddaneOsculatingElements returns the brouwer lyddane osculating elements.
func BrouwerLyddaneOsculatingElements(mean AnalyticKeplerianElements, gravity ZonalHarmonics) (AnalyticKeplerianElements, error) {
	value, err := native.BrouwerLyddaneOsculatingElements(nativeAnalyticKeplerianElements(mean), nativeZonalHarmonics(gravity))
	return publicAnalyticKeplerianElements(value), publicError(err)
}

// Ephemeris evaluates the orbit at the requested TDB epochs.
func (propagator *BrouwerLyddanePropagator) Ephemeris(epochsTDBSeconds []float64) ([]CartesianState, error) {
	value, err := native.BrouwerLyddanePropagatorEphemeris(nativeBrouwerLyddanePropagatorHandle(propagator), epochsTDBSeconds)
	return orbitMap(value, func(value native.CartesianState) CartesianState { return publicCartesian(value) }), publicError(err)
}

// EpochTDBSeconds returns the reference epoch in TDB seconds past J2000.
func (propagator *BrouwerLyddanePropagator) EpochTDBSeconds() (float64, error) {
	value, err := native.BrouwerLyddanePropagatorEpochTDBSeconds(nativeBrouwerLyddanePropagatorHandle(propagator))
	return value, publicError(err)
}

// BrouwerLyddanePropagatorFromMeanElements initializes a propagator from mean Keplerian elements at a TDB epoch.
func BrouwerLyddanePropagatorFromMeanElements(epochTDBSeconds float64, mean AnalyticKeplerianElements, gravity ZonalHarmonics, m2RadS2 float64) (*BrouwerLyddanePropagator, error) {
	value, err := native.BrouwerLyddanePropagatorFromMeanElements(epochTDBSeconds, nativeAnalyticKeplerianElements(mean), nativeZonalHarmonics(gravity), m2RadS2)
	if err != nil {
		return nil, publicError(err)
	}
	return publicBrouwerLyddanePropagator(value), nil
}

// Gravity returns the zonal harmonics used by the propagation theory.
func (propagator *BrouwerLyddanePropagator) Gravity() (ZonalHarmonics, error) {
	value, err := native.BrouwerLyddanePropagatorGravity(nativeBrouwerLyddanePropagatorHandle(propagator))
	return publicZonalHarmonics(value), publicError(err)
}

// M2RadS2 returns the m2 rad s2.
func (propagator *BrouwerLyddanePropagator) M2RadS2() (float64, error) {
	value, err := native.BrouwerLyddanePropagatorM2RadS2(nativeBrouwerLyddanePropagatorHandle(propagator))
	return value, publicError(err)
}

// MeanElements returns the mean elements.
func (propagator *BrouwerLyddanePropagator) MeanElements() (AnalyticKeplerianElements, error) {
	value, err := native.BrouwerLyddanePropagatorMeanElements(nativeBrouwerLyddanePropagatorHandle(propagator))
	return publicAnalyticKeplerianElements(value), publicError(err)
}

// NewBrouwerLyddanePropagator constructs a BrouwerLyddanePropagator from the supplied inputs.
func NewBrouwerLyddanePropagator(initial CartesianState, gravity ZonalHarmonics, m2RadS2 float64, options MeanConversionOptions) (*BrouwerLyddanePropagator, error) {
	value, err := native.BrouwerLyddanePropagatorNew(nativeCartesian(initial), nativeZonalHarmonics(gravity), m2RadS2, nativeMeanConversionOptions(options))
	if err != nil {
		return nil, publicError(err)
	}
	return publicBrouwerLyddanePropagator(value), nil
}

// OsculatingElements returns the osculating elements.
func (propagator *BrouwerLyddanePropagator) OsculatingElements(epochTDBSeconds float64) (AnalyticKeplerianElements, error) {
	value, err := native.BrouwerLyddanePropagatorOsculatingElements(nativeBrouwerLyddanePropagatorHandle(propagator), epochTDBSeconds)
	return publicAnalyticKeplerianElements(value), publicError(err)
}

// Propagate returns the propagate.
func (propagator *BrouwerLyddanePropagator) Propagate(epochTDBSeconds float64) (CartesianState, error) {
	value, err := native.BrouwerLyddanePropagatorPropagate(nativeBrouwerLyddanePropagatorHandle(propagator), epochTDBSeconds)
	return publicCartesian(value), publicError(err)
}

// Eccentricity returns the eccentricity.
func (elements CircularElements) Eccentricity() float64 {
	return native.CircularElementsEccentricity(nativeCircularElements(elements))
}

// CircularElementsFromCartesian converts an inertial Cartesian state to osculating circular elements.
func CircularElementsFromCartesian(state AnalyticCartesian, muKm3S2 float64) (CircularElements, error) {
	value, err := native.CircularElementsFromCartesian(nativeAnalyticCartesian(state), muKm3S2)
	return publicCircularElements(value), publicError(err)
}

// ToCartesian returns the two-body inertial position and velocity in km and km/s.
func (elements CircularElements) ToCartesian(muKm3S2 float64) (AnalyticCartesian, error) {
	value, err := native.CircularElementsToCartesian(nativeCircularElements(elements), muKm3S2)
	return publicAnalyticCartesian(value), publicError(err)
}

// EcksteinHechlerMeanElements converts an osculating state to Eckstein–Hechler mean elements.
func EcksteinHechlerMeanElements(state AnalyticCartesian, gravity ZonalHarmonics, options MeanConversionOptions) (MeanElements[CircularElements], error) {
	value, err := native.ComputeEcksteinHechlerMeanElements(nativeAnalyticCartesian(state), nativeZonalHarmonics(gravity), nativeMeanConversionOptions(options))
	return publicEcksteinHechlerMeanElements(value), publicError(err)
}

// EcksteinHechlerOsculatingElements returns the eckstein hechler osculating elements.
func EcksteinHechlerOsculatingElements(mean CircularElements, gravity ZonalHarmonics) (CircularElements, error) {
	value, err := native.EcksteinHechlerOsculatingElements(nativeCircularElements(mean), nativeZonalHarmonics(gravity))
	return publicCircularElements(value), publicError(err)
}

// Ephemeris evaluates the orbit at the requested TDB epochs.
func (propagator *EcksteinHechlerPropagator) Ephemeris(epochsTDBSeconds []float64) ([]CartesianState, error) {
	value, err := native.EcksteinHechlerPropagatorEphemeris(nativeEcksteinHechlerPropagatorHandle(propagator), epochsTDBSeconds)
	return orbitMap(value, func(value native.CartesianState) CartesianState { return publicCartesian(value) }), publicError(err)
}

// EpochTDBSeconds returns the reference epoch in TDB seconds past J2000.
func (propagator *EcksteinHechlerPropagator) EpochTDBSeconds() (float64, error) {
	value, err := native.EcksteinHechlerPropagatorEpochTDBSeconds(nativeEcksteinHechlerPropagatorHandle(propagator))
	return value, publicError(err)
}

// EcksteinHechlerPropagatorFromMeanElements initializes a propagator from mean circular elements at a TDB epoch.
func EcksteinHechlerPropagatorFromMeanElements(epochTDBSeconds float64, mean CircularElements, gravity ZonalHarmonics) (*EcksteinHechlerPropagator, error) {
	value, err := native.EcksteinHechlerPropagatorFromMeanElements(epochTDBSeconds, nativeCircularElements(mean), nativeZonalHarmonics(gravity))
	if err != nil {
		return nil, publicError(err)
	}
	return publicEcksteinHechlerPropagator(value), nil
}

// Gravity returns the zonal harmonics used by the propagation theory.
func (propagator *EcksteinHechlerPropagator) Gravity() (ZonalHarmonics, error) {
	value, err := native.EcksteinHechlerPropagatorGravity(nativeEcksteinHechlerPropagatorHandle(propagator))
	return publicZonalHarmonics(value), publicError(err)
}

// MeanElements returns the mean elements.
func (propagator *EcksteinHechlerPropagator) MeanElements() (CircularElements, error) {
	value, err := native.EcksteinHechlerPropagatorMeanElements(nativeEcksteinHechlerPropagatorHandle(propagator))
	return publicCircularElements(value), publicError(err)
}

// NewEcksteinHechlerPropagator constructs a EcksteinHechlerPropagator from the supplied inputs.
func NewEcksteinHechlerPropagator(initial CartesianState, gravity ZonalHarmonics, options MeanConversionOptions) (*EcksteinHechlerPropagator, error) {
	value, err := native.EcksteinHechlerPropagatorNew(nativeCartesian(initial), nativeZonalHarmonics(gravity), nativeMeanConversionOptions(options))
	if err != nil {
		return nil, publicError(err)
	}
	return publicEcksteinHechlerPropagator(value), nil
}

// OsculatingElements returns the osculating elements.
func (propagator *EcksteinHechlerPropagator) OsculatingElements(epochTDBSeconds float64) (CircularElements, error) {
	value, err := native.EcksteinHechlerPropagatorOsculatingElements(nativeEcksteinHechlerPropagatorHandle(propagator), epochTDBSeconds)
	return publicCircularElements(value), publicError(err)
}

// Propagate returns the propagate.
func (propagator *EcksteinHechlerPropagator) Propagate(epochTDBSeconds float64) (CartesianState, error) {
	value, err := native.EcksteinHechlerPropagatorPropagate(nativeEcksteinHechlerPropagatorHandle(propagator), epochTDBSeconds)
	return publicCartesian(value), publicError(err)
}

// DefaultMeanConversionOptionsBrouwerLyddane returns the Brouwer–Lyddane conversion tolerance and iteration limit.
func DefaultMeanConversionOptionsBrouwerLyddane() MeanConversionOptions {
	value := native.MeanConversionOptionsBrouwerLyddaneDefault()
	return publicMeanConversionOptions(value)
}

// DefaultMeanConversionOptionsEcksteinHechler returns the Eckstein–Hechler conversion tolerance and iteration limit.
func DefaultMeanConversionOptionsEcksteinHechler() MeanConversionOptions {
	value := native.MeanConversionOptionsEcksteinHechlerDefault()
	return publicMeanConversionOptions(value)
}

// DefaultZonalHarmonics returns the zonal harmonics default.
func DefaultZonalHarmonics() ZonalHarmonics {
	value := native.ZonalHarmonicsDefault()
	return publicZonalHarmonics(value)
}

// ZonalHarmonicsEarth returns the zonal harmonics earth.
func ZonalHarmonicsEarth() ZonalHarmonics {
	value := native.ZonalHarmonicsEarth()
	return publicZonalHarmonics(value)
}

// String returns the display name of the analytic theory.
func (theory AnalyticTheory) String() string {
	value, _ := AnalyticTheoryText(theory)
	return value
}
