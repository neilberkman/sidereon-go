//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

func AccelerationPartialsPositionOnly(_ [3][3]float64) (AccelerationPartials, error) {
	return AccelerationPartials{}, ErrUnavailable
}

func AccelerationPartialsZeros() AccelerationPartials { return AccelerationPartials{} }

func DragForceNew(_ DragParameters) (*ForceModel, error) { return nil, ErrUnavailable }

func EarthRadiationPressureAccelerationWithSunMoon(_ *ForceModel, _ CartesianState, _ [3]float64, _ [3]float64) ([3]float64, error) {
	return [3]float64{}, ErrUnavailable
}

func EarthRadiationPressureMoonEclipseEnabled(_ *ForceModel) (bool, error) {
	return false, ErrUnavailable
}

func EarthRadiationPressureNew(_ float64, _ float64, _ bool) (*ForceModel, error) {
	return nil, ErrUnavailable
}

func EarthRadiationPressureWithMoonEclipse(_ *ForceModel, _ bool) (*ForceModel, error) {
	return nil, ErrUnavailable
}

func ECOM1CoefficientCount() uint64 { return 0 }

func ECOM1CoefficientsFromArray(_ [9]float64) (ECOM1Coefficients, error) {
	return ECOM1Coefficients{}, ErrUnavailable
}

func ECOM1CoefficientsNames(_ uint64) (string, error) { return "", ErrUnavailable }

func ECOM1CoefficientsParameterNames(_ uint64) (string, error) { return "", ErrUnavailable }

func ECOM1CoefficientsToArray(_ ECOM1Coefficients) ([9]float64, error) {
	return [9]float64{}, ErrUnavailable
}

func ECOM1SolarRadiationPressureAccelerationPartialsWithSun(_ *ForceModel, _ [3]float64, _ [3]float64, _ [3]float64) (AccelerationPartials, error) {
	return AccelerationPartials{}, ErrUnavailable
}

func ECOM1SolarRadiationPressureAccelerationPartialsWithSunMoon(_ *ForceModel, _ [3]float64, _ [3]float64, _ [3]float64, _ [3]float64) (AccelerationPartials, error) {
	return AccelerationPartials{}, ErrUnavailable
}

func ECOM1SolarRadiationPressureAccelerationWithSun(_ *ForceModel, _ [3]float64, _ [3]float64, _ [3]float64) ([3]float64, error) {
	return [3]float64{}, ErrUnavailable
}

func ECOM1SolarRadiationPressureAccelerationWithSunMoon(_ *ForceModel, _ [3]float64, _ [3]float64, _ [3]float64, _ [3]float64) ([3]float64, error) {
	return [3]float64{}, ErrUnavailable
}

func ECOM1SolarRadiationPressureCoefficients(_ *ForceModel) (ECOM1Coefficients, error) {
	return ECOM1Coefficients{}, ErrUnavailable
}

func ECOM1SolarRadiationPressureEarthShadowModel(_ *ForceModel) (uint32, error) {
	return 0, ErrUnavailable
}

func ECOM1SolarRadiationPressureEqual(_ *ForceModel, _ *ForceModel) (bool, error) {
	return false, ErrUnavailable
}

func ECOM1SolarRadiationPressureMoonEclipseEnabled(_ *ForceModel) (bool, error) {
	return false, ErrUnavailable
}

func ECOM1SolarRadiationPressureNew(_ ECOM1Coefficients, _ RadiationShadowOptions) (*ForceModel, error) {
	return nil, ErrUnavailable
}

func ECOM1SolarRadiationPressureParameterPartialsWithSun(_ *ForceModel, _ [3]float64, _ [3]float64, _ [3]float64) ([][3]float64, error) {
	return nil, ErrUnavailable
}

func ECOM1SolarRadiationPressureParameterPartialsWithSunMoon(_ *ForceModel, _ [3]float64, _ [3]float64, _ [3]float64, _ [3]float64) ([][3]float64, error) {
	return nil, ErrUnavailable
}

func ECOM1SolarRadiationPressureShadowFunction(_ *ForceModel) (ShadowFunction, error) {
	return 0, ErrUnavailable
}

func ECOM1SolarRadiationPressureWithEarthShadowModel(_ *ForceModel, _ uint32) (*ForceModel, error) {
	return nil, ErrUnavailable
}

func ECOM1SolarRadiationPressureWithMoonEclipse(_ *ForceModel, _ bool) (*ForceModel, error) {
	return nil, ErrUnavailable
}

func ECOM1SolarRadiationPressureWithShadowFunction(_ *ForceModel, _ ShadowFunction) (*ForceModel, error) {
	return nil, ErrUnavailable
}

func ECOM2D2B1CoefficientCount() uint64 { return 0 }

func ECOM2D2B1CoefficientsFromArray(_ [7]float64) (ECOM2D2B1Coefficients, error) {
	return ECOM2D2B1Coefficients{}, ErrUnavailable
}

func ECOM2D2B1CoefficientsNames(_ uint64) (string, error) { return "", ErrUnavailable }

func ECOM2D2B1CoefficientsParameterNames(_ uint64) (string, error) { return "", ErrUnavailable }

func ECOM2D2B1CoefficientsToArray(_ ECOM2D2B1Coefficients) ([7]float64, error) {
	return [7]float64{}, ErrUnavailable
}

func ECOM2D4B1CoefficientCount() uint64 { return 0 }

func ECOM2D4B1CoefficientsFromArray(_ [9]float64) (ECOM2D4B1Coefficients, error) {
	return ECOM2D4B1Coefficients{}, ErrUnavailable
}

func ECOM2D4B1CoefficientsNames(_ uint64) (string, error) { return "", ErrUnavailable }

func ECOM2D4B1CoefficientsParameterNames(_ uint64) (string, error) { return "", ErrUnavailable }

func ECOM2D4B1CoefficientsToArray(_ ECOM2D4B1Coefficients) ([9]float64, error) {
	return [9]float64{}, ErrUnavailable
}

func ECOM2SolarRadiationPressureAccelerationPartialsWithSun(_ *ForceModel, _ [3]float64, _ [3]float64, _ [3]float64) (AccelerationPartials, error) {
	return AccelerationPartials{}, ErrUnavailable
}

func ECOM2SolarRadiationPressureAccelerationPartialsWithSunMoon(_ *ForceModel, _ [3]float64, _ [3]float64, _ [3]float64, _ [3]float64) (AccelerationPartials, error) {
	return AccelerationPartials{}, ErrUnavailable
}

func ECOM2SolarRadiationPressureAccelerationWithSun(_ *ForceModel, _ [3]float64, _ [3]float64, _ [3]float64) ([3]float64, error) {
	return [3]float64{}, ErrUnavailable
}

func ECOM2SolarRadiationPressureAccelerationWithSunMoon(_ *ForceModel, _ [3]float64, _ [3]float64, _ [3]float64, _ [3]float64) ([3]float64, error) {
	return [3]float64{}, ErrUnavailable
}

func ECOM2SolarRadiationPressureCoefficientCount(_ uint64, _ uint64) (uint64, error) {
	return 0, ErrUnavailable
}

func ECOM2SolarRadiationPressureCoefficients(_ *ForceModel) ([]float64, error) {
	return nil, ErrUnavailable
}

func ECOM2SolarRadiationPressureD2B1(_ ECOM2D2B1Coefficients, _ RadiationShadowOptions) (*ForceModel, error) {
	return nil, ErrUnavailable
}

func ECOM2SolarRadiationPressureD4B1(_ ECOM2D4B1Coefficients, _ RadiationShadowOptions) (*ForceModel, error) {
	return nil, ErrUnavailable
}

func ECOM2SolarRadiationPressureEarthShadowModel(_ *ForceModel) (uint32, error) {
	return 0, ErrUnavailable
}

func ECOM2SolarRadiationPressureEqual(_ *ForceModel, _ *ForceModel) (bool, error) {
	return false, ErrUnavailable
}

func ECOM2SolarRadiationPressureMoonEclipseEnabled(_ *ForceModel) (bool, error) {
	return false, ErrUnavailable
}

func ECOM2SolarRadiationPressureNB(_ *ForceModel) (uint64, error) { return 0, ErrUnavailable }

func ECOM2SolarRadiationPressureND(_ *ForceModel) (uint64, error) { return 0, ErrUnavailable }

func ECOM2SolarRadiationPressureNew(_ uint64, _ uint64, _ []float64, _ RadiationShadowOptions) (*ForceModel, error) {
	return nil, ErrUnavailable
}

func ECOM2SolarRadiationPressureParameterName(_ uint64, _ uint64, _ uint64) (string, error) {
	return "", ErrUnavailable
}

func ECOM2SolarRadiationPressureParameterPartialsWithSun(_ *ForceModel, _ [3]float64, _ [3]float64, _ [3]float64) ([][3]float64, error) {
	return nil, ErrUnavailable
}

func ECOM2SolarRadiationPressureParameterPartialsWithSunMoon(_ *ForceModel, _ [3]float64, _ [3]float64, _ [3]float64, _ [3]float64) ([][3]float64, error) {
	return nil, ErrUnavailable
}

func ECOM2SolarRadiationPressureShadowFunction(_ *ForceModel) (ShadowFunction, error) {
	return 0, ErrUnavailable
}

func ECOM2SolarRadiationPressureWithEarthShadowModel(_ *ForceModel, _ uint32) (*ForceModel, error) {
	return nil, ErrUnavailable
}

func ECOM2SolarRadiationPressureWithMoonEclipse(_ *ForceModel, _ bool) (*ForceModel, error) {
	return nil, ErrUnavailable
}

func ECOM2SolarRadiationPressureWithShadowFunction(_ *ForceModel, _ ShadowFunction) (*ForceModel, error) {
	return nil, ErrUnavailable
}

func ComputeECOMGeometry(_ [3]float64, _ [3]float64, _ [3]float64) (ECOMGeometry, error) {
	return ECOMGeometry{}, ErrUnavailable
}

func ECOMGeometryWithShadow(_ [3]float64, _ [3]float64, _ [3]float64, _ *[3]float64, _ uint32) (ECOMGeometry, error) {
	return ECOMGeometry{}, ErrUnavailable
}

func EmpiricalHarmonicAll() ([3]EmpiricalHarmonic, error) {
	return [3]EmpiricalHarmonic{}, ErrUnavailable
}

func EmpiricalRTNAccelerationAllTerms() (*ForceModel, error) { return nil, ErrUnavailable }

func EmpiricalRTNAccelerationCoefficientKmS2(_ *ForceModel, _ RTNAxis, _ EmpiricalHarmonic) (OptionalOrbitScalar, error) {
	return OptionalOrbitScalar{}, ErrUnavailable
}

func EmpiricalRTNAccelerationConstantTerms() (*ForceModel, error) { return nil, ErrUnavailable }

func EmpiricalRTNAccelerationEqual(_ *ForceModel, _ *ForceModel) (bool, error) {
	return false, ErrUnavailable
}

func EmpiricalRTNAccelerationNew(_ []EmpiricalRTNTerm) (*ForceModel, error) {
	return nil, ErrUnavailable
}

func EmpiricalRTNAccelerationParameterName(_ RTNAxis, _ EmpiricalHarmonic) (string, error) {
	return "", ErrUnavailable
}

func EmpiricalRTNAccelerationTerms(_ *ForceModel) ([]EmpiricalRTNTerm, error) {
	return nil, ErrUnavailable
}

func EmpiricalRTNAccelerationWithTerm(_ *ForceModel, _ EmpiricalRTNTerm) (*ForceModel, error) {
	return nil, ErrUnavailable
}

func EmpiricalRTNAccelerationWithZeroTerms(_ []EmpiricalRTNTermSelector) (*ForceModel, error) {
	return nil, ErrUnavailable
}

func ForceModelAcceleration(_ *ForceModel, _ CartesianState, _ *PropagationContext) ([3]float64, error) {
	return [3]float64{}, ErrUnavailable
}

func ForceModelAccelerationPartials(_ *ForceModel, _ CartesianState, _ []float64, _ *PropagationContext) (AccelerationPartials, error) {
	return AccelerationPartials{}, ErrUnavailable
}

func ForceModelAccelerationWithParameters(_ *ForceModel, _ CartesianState, _ []float64, _ *PropagationContext) ([3]float64, error) {
	return [3]float64{}, ErrUnavailable
}

func ForceModelParameterCount(_ *ForceModel) (uint64, error) { return 0, ErrUnavailable }

func ForceModelParameterPartials(_ *ForceModel, _ CartesianState, _ []float64, _ *PropagationContext) ([][3]float64, error) {
	return nil, ErrUnavailable
}

func J2GravityNew(_ float64, _ float64, _ float64) (*ForceModel, error) { return nil, ErrUnavailable }

func RadiationShadowOptionsDefault() RadiationShadowOptions { return RadiationShadowOptions{} }

func RTNAxisAll() ([3]RTNAxis, error) { return [3]RTNAxis{}, ErrUnavailable }

func SchwarzschildRelativityNew(_ float64, _ float64) (*ForceModel, error) {
	return nil, ErrUnavailable
}

func SolarRadiationPressureAccelerationWithSunMoon(_ *ForceModel, _ [3]float64, _ [3]float64, _ [3]float64) ([3]float64, error) {
	return [3]float64{}, ErrUnavailable
}

func SolarRadiationPressureAreaToMassM2Kg(_ *ForceModel) (float64, error) { return 0, ErrUnavailable }

func SolarRadiationPressureAUKm(_ *ForceModel) (float64, error) { return 0, ErrUnavailable }

func SolarRadiationPressureCR(_ *ForceModel) (float64, error) { return 0, ErrUnavailable }

func SolarRadiationPressureCRParameter() (string, error) { return "", ErrUnavailable }

func SolarRadiationPressureEarthShadowModel(_ *ForceModel) (uint32, error) { return 0, ErrUnavailable }

func SolarRadiationPressureMoonEclipseEnabled(_ *ForceModel) (bool, error) {
	return false, ErrUnavailable
}

func SolarRadiationPressureNew(_ float64, _ float64, _ RadiationShadowOptions) (*ForceModel, error) {
	return nil, ErrUnavailable
}

func SolarRadiationPressureShadowFunction(_ *ForceModel) (ShadowFunction, error) {
	return 0, ErrUnavailable
}

func SolarRadiationPressureWithEarthShadowModel(_ *ForceModel, _ uint32) (*ForceModel, error) {
	return nil, ErrUnavailable
}

func SolarRadiationPressureWithMoonEclipse(_ *ForceModel, _ bool) (*ForceModel, error) {
	return nil, ErrUnavailable
}

func SolarRadiationPressureWithShadowFunction(_ *ForceModel, _ ShadowFunction) (*ForceModel, error) {
	return nil, ErrUnavailable
}

func SolidEarthPoleTideGravityNew() (*ForceModel, error) { return nil, ErrUnavailable }

func SolidEarthTideGravityNew(_ GravityTideSystem) (*ForceModel, error) { return nil, ErrUnavailable }

func SourcedDragForceNew(_ float64, _ float64, _ *SpaceWeatherTable) (*ForceModel, error) {
	return nil, ErrUnavailable
}

func ThirdBodyGravityNew(_ bool, _ bool) (*ForceModel, error) { return nil, ErrUnavailable }

func TwoBodyGravityNew(_ float64) (*ForceModel, error) { return nil, ErrUnavailable }

func ZonalGravityNew(_ float64, _ float64, _ uint8, _ GravityTideSystem) (*ForceModel, error) {
	return nil, ErrUnavailable
}
