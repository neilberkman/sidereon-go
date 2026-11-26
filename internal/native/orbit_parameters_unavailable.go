//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

func OrbitParameterFitCorrelation(_ *OrbitParameterFit, _ string, _ string) (OptionalOrbitScalar, error) {
	return OptionalOrbitScalar{}, ErrUnavailable
}

func OrbitParameterFitCovariance(_ *OrbitParameterFit) (OrbitParameterCovariance, error) {
	return OrbitParameterCovariance{}, ErrUnavailable
}

func OrbitParameterFitCovarianceMatrix(_ *OrbitParameterFit, _ bool) ([]float64, error) {
	return nil, ErrUnavailable
}

func OrbitParameterFitEqual(_ *OrbitParameterFit, _ *OrbitParameterFit) (bool, error) {
	return false, ErrUnavailable
}

func OrbitParameterFitForceParameter(_ *OrbitParameterFit, _ string) (EstimatedForceParameter, error) {
	return EstimatedForceParameter{}, ErrUnavailable
}

func OrbitParameterFitParameterName(_ *OrbitParameterFit, _ uint64) (string, error) {
	return "", ErrUnavailable
}

func OrbitParameterFitResiduals(_ *OrbitParameterFit) ([]OrbitRTNResidual, error) {
	return nil, ErrUnavailable
}

func OrbitParameterFitSigma(_ *OrbitParameterFit, _ string) (OptionalOrbitScalar, error) {
	return OptionalOrbitScalar{}, ErrUnavailable
}

func ReadOrbitParameterFitSummary(_ *OrbitParameterFit) (OrbitParameterFitSummary, error) {
	return OrbitParameterFitSummary{}, ErrUnavailable
}

func StateParameterName(_ uint64) (string, error) { return "", ErrUnavailable }

func StateParameterNameCount() uint64 { return 0 }

func OrbitParameterFitOptionsDefault() OrbitParameterFitOptions { return OrbitParameterFitOptions{} }
func FitInertialOrbitParameters([]InertialPositionSample, string, *CartesianState, OrbitParameterFitOptions, *bool) (*OrbitParameterFit, error) {
	return nil, ErrUnavailable
}
func FitSP3ECEFOrbitParameters(*SP3, *IERS2010EarthOrientationProvider, string, *CartesianState, OrbitParameterFitOptions, *bool) (*OrbitParameterFit, error) {
	return nil, ErrUnavailable
}

func OrbitParameterFitForceParameterAt(*OrbitParameterFit, uint64) (EstimatedForceParameter, error) {
	return EstimatedForceParameter{}, ErrUnavailable
}
