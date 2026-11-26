//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

func IERS2010EarthOrientationProviderEqual(_ *IERS2010EarthOrientationProvider, _ *IERS2010EarthOrientationProvider) (bool, error) {
	return false, ErrUnavailable
}

func IERS2010EarthOrientationProviderNew(_ []IERS2010EopSample) (*IERS2010EarthOrientationProvider, error) {
	return nil, ErrUnavailable
}

func IERS2010EarthOrientationProviderOrientationAtTDBSeconds(_ *IERS2010EarthOrientationProvider, _ float64) (EarthOrientation, error) {
	return EarthOrientation{}, ErrUnavailable
}

func IERS2010EarthOrientationProviderOrientationAtUTCMJD(_ *IERS2010EarthOrientationProvider, _ int32, _ float64) (EarthOrientation, error) {
	return EarthOrientation{}, ErrUnavailable
}

func IERS2010EarthOrientationProviderWithValidity(_ *IERS2010EarthOrientationProvider, _ bool) (*IERS2010EarthOrientationProvider, error) {
	return nil, ErrUnavailable
}
