//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

func AnalyticKeplerianElementsFromCartesian(_ AnalyticCartesian, _ float64) (AnalyticKeplerianElements, error) {
	return AnalyticKeplerianElements{}, ErrUnavailable
}

func AnalyticKeplerianElementsToCartesian(_ AnalyticKeplerianElements, _ float64) (AnalyticCartesian, error) {
	return AnalyticCartesian{}, ErrUnavailable
}

func AnalyticTheoryText(_ AnalyticTheory) (string, error) { return "", ErrUnavailable }

func ComputeBrouwerLyddaneMeanElements(_ AnalyticCartesian, _ ZonalHarmonics, _ MeanConversionOptions) (BrouwerLyddaneMeanElements, error) {
	return BrouwerLyddaneMeanElements{}, ErrUnavailable
}

func BrouwerLyddaneOsculatingElements(_ AnalyticKeplerianElements, _ ZonalHarmonics) (AnalyticKeplerianElements, error) {
	return AnalyticKeplerianElements{}, ErrUnavailable
}

func BrouwerLyddanePropagatorEphemeris(_ *BrouwerLyddanePropagator, _ []float64) ([]CartesianState, error) {
	return nil, ErrUnavailable
}

func BrouwerLyddanePropagatorEpochTDBSeconds(_ *BrouwerLyddanePropagator) (float64, error) {
	return 0, ErrUnavailable
}

func BrouwerLyddanePropagatorFromMeanElements(_ float64, _ AnalyticKeplerianElements, _ ZonalHarmonics, _ float64) (*BrouwerLyddanePropagator, error) {
	return nil, ErrUnavailable
}

func BrouwerLyddanePropagatorGravity(_ *BrouwerLyddanePropagator) (ZonalHarmonics, error) {
	return ZonalHarmonics{}, ErrUnavailable
}

func BrouwerLyddanePropagatorM2RadS2(_ *BrouwerLyddanePropagator) (float64, error) {
	return 0, ErrUnavailable
}

func BrouwerLyddanePropagatorMeanElements(_ *BrouwerLyddanePropagator) (AnalyticKeplerianElements, error) {
	return AnalyticKeplerianElements{}, ErrUnavailable
}

func BrouwerLyddanePropagatorNew(_ CartesianState, _ ZonalHarmonics, _ float64, _ MeanConversionOptions) (*BrouwerLyddanePropagator, error) {
	return nil, ErrUnavailable
}

func BrouwerLyddanePropagatorOsculatingElements(_ *BrouwerLyddanePropagator, _ float64) (AnalyticKeplerianElements, error) {
	return AnalyticKeplerianElements{}, ErrUnavailable
}

func BrouwerLyddanePropagatorPropagate(_ *BrouwerLyddanePropagator, _ float64) (CartesianState, error) {
	return CartesianState{}, ErrUnavailable
}

func CircularElementsEccentricity(_ CircularElements) float64 { return 0 }

func CircularElementsFromCartesian(_ AnalyticCartesian, _ float64) (CircularElements, error) {
	return CircularElements{}, ErrUnavailable
}

func CircularElementsToCartesian(_ CircularElements, _ float64) (AnalyticCartesian, error) {
	return AnalyticCartesian{}, ErrUnavailable
}

func ComputeEcksteinHechlerMeanElements(_ AnalyticCartesian, _ ZonalHarmonics, _ MeanConversionOptions) (EcksteinHechlerMeanElements, error) {
	return EcksteinHechlerMeanElements{}, ErrUnavailable
}

func EcksteinHechlerOsculatingElements(_ CircularElements, _ ZonalHarmonics) (CircularElements, error) {
	return CircularElements{}, ErrUnavailable
}

func EcksteinHechlerPropagatorEphemeris(_ *EcksteinHechlerPropagator, _ []float64) ([]CartesianState, error) {
	return nil, ErrUnavailable
}

func EcksteinHechlerPropagatorEpochTDBSeconds(_ *EcksteinHechlerPropagator) (float64, error) {
	return 0, ErrUnavailable
}

func EcksteinHechlerPropagatorFromMeanElements(_ float64, _ CircularElements, _ ZonalHarmonics) (*EcksteinHechlerPropagator, error) {
	return nil, ErrUnavailable
}

func EcksteinHechlerPropagatorGravity(_ *EcksteinHechlerPropagator) (ZonalHarmonics, error) {
	return ZonalHarmonics{}, ErrUnavailable
}

func EcksteinHechlerPropagatorMeanElements(_ *EcksteinHechlerPropagator) (CircularElements, error) {
	return CircularElements{}, ErrUnavailable
}

func EcksteinHechlerPropagatorNew(_ CartesianState, _ ZonalHarmonics, _ MeanConversionOptions) (*EcksteinHechlerPropagator, error) {
	return nil, ErrUnavailable
}

func EcksteinHechlerPropagatorOsculatingElements(_ *EcksteinHechlerPropagator, _ float64) (CircularElements, error) {
	return CircularElements{}, ErrUnavailable
}

func EcksteinHechlerPropagatorPropagate(_ *EcksteinHechlerPropagator, _ float64) (CartesianState, error) {
	return CartesianState{}, ErrUnavailable
}

func MeanConversionOptionsBrouwerLyddaneDefault() MeanConversionOptions {
	return MeanConversionOptions{}
}

func MeanConversionOptionsEcksteinHechlerDefault() MeanConversionOptions {
	return MeanConversionOptions{}
}

func ZonalHarmonicsDefault() ZonalHarmonics { return ZonalHarmonics{} }

func ZonalHarmonicsEarth() ZonalHarmonics { return ZonalHarmonics{} }
