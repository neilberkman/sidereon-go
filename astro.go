package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// EclipseStatus is the C eclipse classification.
type EclipseStatus uint32

const (
	// EclipseSunlit indicates full solar illumination.
	EclipseSunlit EclipseStatus = 0
	// EclipsePenumbra indicates partial solar illumination in the penumbra.
	EclipsePenumbra EclipseStatus = 1
	// EclipseUmbra indicates that the satellite is in the umbra.
	EclipseUmbra EclipseStatus = 2
)

// EarthShadowModel selects the Earth model used by eclipse helpers.
type EarthShadowModel uint32

const (
	// EarthShadowSpherical selects a spherical-Earth shadow model.
	EarthShadowSpherical EarthShadowModel = 0
	// EarthShadowWGS84Oblate selects an oblate WGS 84 Earth-shadow model.
	EarthShadowWGS84Oblate EarthShadowModel = 1
)

// AngularSeparationCoords returns the on-sky separation of two (longitude,
// latitude) pairs in degrees. The latitude inputs are degrees.
func AngularSeparationCoords(aLongitudeDeg, aLatitudeDeg, bLongitudeDeg, bLatitudeDeg float64) (float64, error) {
	value, err := native.AngularSeparationCoords(aLongitudeDeg, aLatitudeDeg, bLongitudeDeg, bLatitudeDeg)
	return value, publicError(err)
}

// AngularSeparation returns the separation in degrees between two non-zero
// direction vectors.
func AngularSeparation(a, b [3]float64) (float64, error) {
	value, err := native.AngularSeparation(a, b)
	return value, publicError(err)
}

// EarthAngularRadius returns the apparent Earth angular radius in degrees from
// a satellite position in kilometres.
func EarthAngularRadius(satellitePositionKm [3]float64) (float64, error) {
	value, err := native.EarthAngularRadius(satellitePositionKm)
	return value, publicError(err)
}

// EclipseShadowFraction returns fractional solar illumination at a satellite
// position, both positions in kilometres.
func EclipseShadowFraction(satellitePositionKm, sunPositionKm [3]float64) (float64, error) {
	value, err := native.EclipseShadowFraction(satellitePositionKm, sunPositionKm, nil)
	return value, publicError(err)
}

// EclipseShadowFractionWithModel returns fractional solar illumination using
// an explicit Earth shadow model.
func EclipseShadowFractionWithModel(satellitePositionKm, sunPositionKm [3]float64, model EarthShadowModel) (float64, error) {
	modelValue := uint32(model)
	value, err := native.EclipseShadowFraction(satellitePositionKm, sunPositionKm, &modelValue)
	return value, publicError(err)
}

// EclipseStatusFor returns the C sunlit, penumbra, or umbra classification.
func EclipseStatusFor(satellitePositionKm, sunPositionKm [3]float64) (EclipseStatus, error) {
	value, err := native.EclipseStatus(satellitePositionKm, sunPositionKm, nil)
	return EclipseStatus(value), publicError(err)
}

// EclipseStatusWithModel returns the C eclipse classification using an
// explicit Earth shadow model.
func EclipseStatusWithModel(satellitePositionKm, sunPositionKm [3]float64, model EarthShadowModel) (EclipseStatus, error) {
	modelValue := uint32(model)
	value, err := native.EclipseStatus(satellitePositionKm, sunPositionKm, &modelValue)
	return EclipseStatus(value), publicError(err)
}
