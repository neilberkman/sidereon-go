//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

func PropagateStateWithForceModel(NativePropagationConfig, []float64, PropagationEnvironment, []*ForceModel) (*Ephemeris, error) {
	return nil, ErrUnavailable
}
func PropagateCovarianceWithContext(NativePropagationConfig, [6][6]float64, []float64, uint32, uint32, NativeProcessNoise, PropagationEnvironment) (*CovarianceEphemeris, error) {
	return nil, ErrUnavailable
}
func PropagateStateWithCovarianceWithContext(NativePropagationConfig, [6][6]float64, float64, PropagationEnvironment) (StateWithCovariance, error) {
	return StateWithCovariance{}, ErrUnavailable
}
func StateTransitionMatrixToWithContext(NativePropagationConfig, float64, PropagationEnvironment) ([6][6]float64, error) {
	return [6][6]float64{}, ErrUnavailable
}
func EstimateDecayWithContext(CartesianState, DecayConfig, *SpaceWeatherTable, *PropagationContext) (DecayEstimate, error) {
	return DecayEstimate{}, ErrUnavailable
}

func FitSP3PreciseOrbitWithContext(*SP3, string, *OrbitFitOptions, bool, PropagationEnvironment) (*OrbitFitReport, error) {
	return nil, ErrUnavailable
}
func FitSP3ECEFPreciseOrbitsWithContext(*SP3, []string, *OrbitFitOptions, bool, PropagationEnvironment) (*OrbitFitReport, error) {
	return nil, ErrUnavailable
}
func FitPreciseEphemerisSamplesWithContext([]PreciseEphemerisSample, string, *OrbitFitOptions, PropagationEnvironment) (*OrbitFitReport, error) {
	return nil, ErrUnavailable
}
