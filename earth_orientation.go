package sidereon

import (
	"sidereon.dev/go/v4/internal/native"
)

// Equal compares the underlying values.
func (left *IERS2010EarthOrientationProvider) Equal(right *IERS2010EarthOrientationProvider) (bool, error) {
	value, err := native.IERS2010EarthOrientationProviderEqual(nativeIERS2010EarthOrientationProviderHandle(left), nativeIERS2010EarthOrientationProviderHandle(right))
	return value, publicError(err)
}

// NewIERS2010EarthOrientationProvider constructs a IERS2010EarthOrientationProvider from the supplied inputs.
func NewIERS2010EarthOrientationProvider(samples []IERS2010EopSample) (*IERS2010EarthOrientationProvider, error) {
	value, err := native.IERS2010EarthOrientationProviderNew(orbitMap(samples, func(item IERS2010EopSample) native.IERS2010EopSample { return nativeIERS2010EopSample(item) }))
	if err != nil {
		return nil, publicError(err)
	}
	return publicIERS2010EarthOrientationProvider(value), nil
}

// OrientationAtTDBSeconds returns the orientation at TDB seconds.
func (provider *IERS2010EarthOrientationProvider) OrientationAtTDBSeconds(epochTDBSeconds float64) (EarthOrientation, error) {
	value, err := native.IERS2010EarthOrientationProviderOrientationAtTDBSeconds(nativeIERS2010EarthOrientationProviderHandle(provider), epochTDBSeconds)
	return publicEarthOrientation(value), publicError(err)
}

// OrientationAtUTCMJD returns the orientation at UTC MJD.
func (provider *IERS2010EarthOrientationProvider) OrientationAtUTCMJD(mjdUTC int32, secondsInDay float64) (EarthOrientation, error) {
	value, err := native.IERS2010EarthOrientationProviderOrientationAtUTCMJD(nativeIERS2010EarthOrientationProviderHandle(provider), mjdUTC, secondsInDay)
	return publicEarthOrientation(value), publicError(err)
}

// WithValidity returns a copy with the requested change.
func (provider *IERS2010EarthOrientationProvider) WithValidity(permissive bool) (*IERS2010EarthOrientationProvider, error) {
	value, err := native.IERS2010EarthOrientationProviderWithValidity(nativeIERS2010EarthOrientationProviderHandle(provider), permissive)
	if err != nil {
		return nil, publicError(err)
	}
	return publicIERS2010EarthOrientationProvider(value), nil
}
