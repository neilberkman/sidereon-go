//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

func EclipseMoonShadowFraction(_ [3]float64, _ [3]float64, _ [3]float64) (float64, error) {
	return 0, ErrUnavailable
}

func EclipseShadowFractionWithGradient(_ [3]float64, _ [3]float64) (ShadowGradient, error) {
	return ShadowGradient{}, ErrUnavailable
}

func EclipseShadowFractionWithModelAndGradient(_ [3]float64, _ [3]float64, _ uint32) (ShadowGradient, error) {
	return ShadowGradient{}, ErrUnavailable
}

func EclipseShadowFractionWithModelAndMoon(_ [3]float64, _ [3]float64, _ [3]float64, _ uint32) (float64, error) {
	return 0, ErrUnavailable
}

func EclipseShadowFractionWithModelMoonAndGradient(_ [3]float64, _ [3]float64, _ [3]float64, _ uint32) (ShadowGradient, error) {
	return ShadowGradient{}, ErrUnavailable
}

func EclipseShadowFractionWithMoon(_ [3]float64, _ [3]float64, _ [3]float64) (float64, error) {
	return 0, ErrUnavailable
}

func EclipseStatusWithModelAndMoon(_ [3]float64, _ [3]float64, _ [3]float64, _ uint32) (uint32, error) {
	return 0, ErrUnavailable
}

func EclipseStatusWithMoon(_ [3]float64, _ [3]float64, _ [3]float64) (uint32, error) {
	return 0, ErrUnavailable
}

func IAU2015SolarRadiusKm() float64 { return 0 }
