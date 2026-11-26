package sidereon

import (
	"sidereon.dev/go/v4/internal/native"
)

// MoonShadowFraction returns the eclipse moon shadow fraction.
func MoonShadowFraction(satPosKm [3]float64, sunPosKm [3]float64, moonPosKm [3]float64) (float64, error) {
	value, err := native.EclipseMoonShadowFraction(satPosKm, sunPosKm, moonPosKm)
	return value, publicError(err)
}

// ShadowFractionWithGradient returns the eclipse shadow fraction with gradient.
func ShadowFractionWithGradient(satPosKm [3]float64, sunPosKm [3]float64) (ShadowGradient, error) {
	value, err := native.EclipseShadowFractionWithGradient(satPosKm, sunPosKm)
	return publicShadowGradient(value), publicError(err)
}

// ShadowFractionWithModelAndGradient returns the eclipse shadow fraction with model and gradient.
func ShadowFractionWithModelAndGradient(satPosKm [3]float64, sunPosKm [3]float64, model EarthShadowModel) (ShadowGradient, error) {
	value, err := native.EclipseShadowFractionWithModelAndGradient(satPosKm, sunPosKm, uint32(model))
	return publicShadowGradient(value), publicError(err)
}

// ShadowFractionWithModelAndMoon returns the eclipse shadow fraction with model and moon.
func ShadowFractionWithModelAndMoon(satPosKm [3]float64, sunPosKm [3]float64, moonPosKm [3]float64, model EarthShadowModel) (float64, error) {
	value, err := native.EclipseShadowFractionWithModelAndMoon(satPosKm, sunPosKm, moonPosKm, uint32(model))
	return value, publicError(err)
}

// ShadowFractionWithModelMoonAndGradient returns the eclipse shadow fraction with model moon and gradient.
func ShadowFractionWithModelMoonAndGradient(satPosKm [3]float64, sunPosKm [3]float64, moonPosKm [3]float64, model EarthShadowModel) (ShadowGradient, error) {
	value, err := native.EclipseShadowFractionWithModelMoonAndGradient(satPosKm, sunPosKm, moonPosKm, uint32(model))
	return publicShadowGradient(value), publicError(err)
}

// ShadowFractionWithMoon returns the eclipse shadow fraction with moon.
func ShadowFractionWithMoon(satPosKm [3]float64, sunPosKm [3]float64, moonPosKm [3]float64) (float64, error) {
	value, err := native.EclipseShadowFractionWithMoon(satPosKm, sunPosKm, moonPosKm)
	return value, publicError(err)
}

// EclipseStatusWithModelAndMoon returns the eclipse status with model and moon.
func EclipseStatusWithModelAndMoon(satPosKm [3]float64, sunPosKm [3]float64, moonPosKm [3]float64, model EarthShadowModel) (EclipseStatus, error) {
	value, err := native.EclipseStatusWithModelAndMoon(satPosKm, sunPosKm, moonPosKm, uint32(model))
	return EclipseStatus(value), publicError(err)
}

// EclipseStatusWithMoon returns the eclipse status with moon.
func EclipseStatusWithMoon(satPosKm [3]float64, sunPosKm [3]float64, moonPosKm [3]float64) (EclipseStatus, error) {
	value, err := native.EclipseStatusWithMoon(satPosKm, sunPosKm, moonPosKm)
	return EclipseStatus(value), publicError(err)
}

// IAU2015SolarRadiusKm returns the iau 2015 solar radius km.
func IAU2015SolarRadiusKm() float64 {
	return native.IAU2015SolarRadiusKm()
}
