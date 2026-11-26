package sidereon

import (
	"cmp"

	"sidereon.dev/go/v4/internal/native"
)

// AccelerationPartialsPositionOnly sets the position derivatives and zeros the velocity derivatives.
func AccelerationPartialsPositionOnly(wrtPosition [3][3]float64) (AccelerationPartials, error) {
	value, err := native.AccelerationPartialsPositionOnly(wrtPosition)
	return publicAccelerationPartials(value), publicError(err)
}

// AccelerationPartialsZeros returns zero position and velocity derivatives.
func AccelerationPartialsZeros() AccelerationPartials {
	value := native.AccelerationPartialsZeros()
	return publicAccelerationPartials(value)
}

// NewDragForce constructs a DragForce from the supplied inputs.
func NewDragForce(parameters DragParameters) (*DragForce, error) {
	value, err := native.DragForceNew(nativeDragParameters(parameters))
	if err != nil {
		return nil, publicError(err)
	}
	return publicDragForce(value), nil
}

// AccelerationWithSunMoon returns the acceleration with sun moon.
func (model *EarthRadiationPressure) AccelerationWithSunMoon(state CartesianState, sunPositionKm [3]float64, moonPositionKm [3]float64) ([3]float64, error) {
	value, err := native.EarthRadiationPressureAccelerationWithSunMoon(nativeEarthRadiationPressureHandle(model), nativeCartesian(state), sunPositionKm, moonPositionKm)
	return value, publicError(err)
}

// MoonEclipseEnabled returns the moon eclipse enabled.
func (model *EarthRadiationPressure) MoonEclipseEnabled() (bool, error) {
	value, err := native.EarthRadiationPressureMoonEclipseEnabled(nativeEarthRadiationPressureHandle(model))
	return value, publicError(err)
}

// NewEarthRadiationPressure constructs a EarthRadiationPressure from the supplied inputs.
func NewEarthRadiationPressure(cr float64, areaToMassM2Kg float64, moonEclipseEnabled bool) (*EarthRadiationPressure, error) {
	value, err := native.EarthRadiationPressureNew(cr, areaToMassM2Kg, moonEclipseEnabled)
	if err != nil {
		return nil, publicError(err)
	}
	return publicEarthRadiationPressure(value), nil
}

// WithMoonEclipse returns a copy with the requested change.
func (model *EarthRadiationPressure) WithMoonEclipse(enabled bool) (*EarthRadiationPressure, error) {
	value, err := native.EarthRadiationPressureWithMoonEclipse(nativeEarthRadiationPressureHandle(model), enabled)
	if err != nil {
		return nil, publicError(err)
	}
	return publicEarthRadiationPressure(value), nil
}

// ECOM1CoefficientCount returns the ecom1 coefficient count.
func ECOM1CoefficientCount() uint64 {
	return native.ECOM1CoefficientCount()
}

// ECOM1CoefficientsFromArray returns the from array.
func ECOM1CoefficientsFromArray(values [9]float64) (ECOM1Coefficients, error) {
	value, err := native.ECOM1CoefficientsFromArray(values)
	return publicECOM1Coefficients(value), publicError(err)
}

// ECOM1CoefficientsNames returns the names.
func ECOM1CoefficientsNames(index uint64) (string, error) {
	value, err := native.ECOM1CoefficientsNames(index)
	return value, publicError(err)
}

// ECOM1CoefficientsParameterNames returns the parameter names.
func ECOM1CoefficientsParameterNames(index uint64) (string, error) {
	value, err := native.ECOM1CoefficientsParameterNames(index)
	return value, publicError(err)
}

// ToArray returns the coefficients in their documented order.
func (coefficients ECOM1Coefficients) ToArray() ([9]float64, error) {
	value, err := native.ECOM1CoefficientsToArray(nativeECOM1Coefficients(coefficients))
	return value, publicError(err)
}

// AccelerationPartialsWithSun returns the acceleration partials with sun.
func (model *ECOM1SolarRadiationPressure) AccelerationPartialsWithSun(positionKm [3]float64, velocityKmS [3]float64, sunPositionKm [3]float64) (AccelerationPartials, error) {
	value, err := native.ECOM1SolarRadiationPressureAccelerationPartialsWithSun(nativeECOM1SolarRadiationPressureHandle(model), positionKm, velocityKmS, sunPositionKm)
	return publicAccelerationPartials(value), publicError(err)
}

// AccelerationPartialsWithSunMoon returns the acceleration partials with sun moon.
func (model *ECOM1SolarRadiationPressure) AccelerationPartialsWithSunMoon(positionKm [3]float64, velocityKmS [3]float64, sunPositionKm [3]float64, moonPositionKm [3]float64) (AccelerationPartials, error) {
	value, err := native.ECOM1SolarRadiationPressureAccelerationPartialsWithSunMoon(nativeECOM1SolarRadiationPressureHandle(model), positionKm, velocityKmS, sunPositionKm, moonPositionKm)
	return publicAccelerationPartials(value), publicError(err)
}

// AccelerationWithSun returns the acceleration with sun.
func (model *ECOM1SolarRadiationPressure) AccelerationWithSun(positionKm [3]float64, velocityKmS [3]float64, sunPositionKm [3]float64) ([3]float64, error) {
	value, err := native.ECOM1SolarRadiationPressureAccelerationWithSun(nativeECOM1SolarRadiationPressureHandle(model), positionKm, velocityKmS, sunPositionKm)
	return value, publicError(err)
}

// AccelerationWithSunMoon returns the acceleration with sun moon.
func (model *ECOM1SolarRadiationPressure) AccelerationWithSunMoon(positionKm [3]float64, velocityKmS [3]float64, sunPositionKm [3]float64, moonPositionKm [3]float64) ([3]float64, error) {
	value, err := native.ECOM1SolarRadiationPressureAccelerationWithSunMoon(nativeECOM1SolarRadiationPressureHandle(model), positionKm, velocityKmS, sunPositionKm, moonPositionKm)
	return value, publicError(err)
}

// Coefficients returns the coefficients.
func (model *ECOM1SolarRadiationPressure) Coefficients() (ECOM1Coefficients, error) {
	value, err := native.ECOM1SolarRadiationPressureCoefficients(nativeECOM1SolarRadiationPressureHandle(model))
	return publicECOM1Coefficients(value), publicError(err)
}

// EarthShadowModel returns the earth shadow model.
func (model *ECOM1SolarRadiationPressure) EarthShadowModel() (EarthShadowModel, error) {
	value, err := native.ECOM1SolarRadiationPressureEarthShadowModel(nativeECOM1SolarRadiationPressureHandle(model))
	return EarthShadowModel(value), publicError(err)
}

// Equal compares the underlying values.
func (left *ECOM1SolarRadiationPressure) Equal(right *ECOM1SolarRadiationPressure) (bool, error) {
	value, err := native.ECOM1SolarRadiationPressureEqual(nativeECOM1SolarRadiationPressureHandle(left), nativeECOM1SolarRadiationPressureHandle(right))
	return value, publicError(err)
}

// MoonEclipseEnabled returns the moon eclipse enabled.
func (model *ECOM1SolarRadiationPressure) MoonEclipseEnabled() (bool, error) {
	value, err := native.ECOM1SolarRadiationPressureMoonEclipseEnabled(nativeECOM1SolarRadiationPressureHandle(model))
	return value, publicError(err)
}

// NewECOM1SolarRadiationPressure constructs an ECOM1SolarRadiationPressure from the supplied inputs.
func NewECOM1SolarRadiationPressure(coefficients ECOM1Coefficients, shadow RadiationShadowOptions) (*ECOM1SolarRadiationPressure, error) {
	value, err := native.ECOM1SolarRadiationPressureNew(nativeECOM1Coefficients(coefficients), nativeRadiationShadowOptions(shadow))
	if err != nil {
		return nil, publicError(err)
	}
	return publicECOM1SolarRadiationPressure(value), nil
}

// ParameterPartialsWithSun returns the parameter partials with sun.
func (model *ECOM1SolarRadiationPressure) ParameterPartialsWithSun(positionKm [3]float64, velocityKmS [3]float64, sunPositionKm [3]float64) ([][3]float64, error) {
	value, err := native.ECOM1SolarRadiationPressureParameterPartialsWithSun(nativeECOM1SolarRadiationPressureHandle(model), positionKm, velocityKmS, sunPositionKm)
	return value, publicError(err)
}

// ParameterPartialsWithSunMoon returns the parameter partials with sun moon.
func (model *ECOM1SolarRadiationPressure) ParameterPartialsWithSunMoon(positionKm [3]float64, velocityKmS [3]float64, sunPositionKm [3]float64, moonPositionKm [3]float64) ([][3]float64, error) {
	value, err := native.ECOM1SolarRadiationPressureParameterPartialsWithSunMoon(nativeECOM1SolarRadiationPressureHandle(model), positionKm, velocityKmS, sunPositionKm, moonPositionKm)
	return value, publicError(err)
}

// ShadowFunction returns the shadow function.
func (model *ECOM1SolarRadiationPressure) ShadowFunction() (ShadowFunction, error) {
	value, err := native.ECOM1SolarRadiationPressureShadowFunction(nativeECOM1SolarRadiationPressureHandle(model))
	return ShadowFunction(value), publicError(err)
}

// WithEarthShadowModel returns a copy with the requested change.
func (model *ECOM1SolarRadiationPressure) WithEarthShadowModel(earthModel EarthShadowModel) (*ECOM1SolarRadiationPressure, error) {
	value, err := native.ECOM1SolarRadiationPressureWithEarthShadowModel(nativeECOM1SolarRadiationPressureHandle(model), uint32(earthModel))
	if err != nil {
		return nil, publicError(err)
	}
	return publicECOM1SolarRadiationPressure(value), nil
}

// WithMoonEclipse returns a copy with the requested change.
func (model *ECOM1SolarRadiationPressure) WithMoonEclipse(enabled bool) (*ECOM1SolarRadiationPressure, error) {
	value, err := native.ECOM1SolarRadiationPressureWithMoonEclipse(nativeECOM1SolarRadiationPressureHandle(model), enabled)
	if err != nil {
		return nil, publicError(err)
	}
	return publicECOM1SolarRadiationPressure(value), nil
}

// WithShadowFunction returns a copy with the requested change.
func (model *ECOM1SolarRadiationPressure) WithShadowFunction(shadowFunction ShadowFunction) (*ECOM1SolarRadiationPressure, error) {
	value, err := native.ECOM1SolarRadiationPressureWithShadowFunction(nativeECOM1SolarRadiationPressureHandle(model), native.ShadowFunction(shadowFunction))
	if err != nil {
		return nil, publicError(err)
	}
	return publicECOM1SolarRadiationPressure(value), nil
}

// ECOM2D2B1CoefficientCount returns the ecom2 d2b1 coefficient count.
func ECOM2D2B1CoefficientCount() uint64 {
	return native.ECOM2D2B1CoefficientCount()
}

// ECOM2D2B1CoefficientsFromArray returns the from array.
func ECOM2D2B1CoefficientsFromArray(values [7]float64) (ECOM2D2B1Coefficients, error) {
	value, err := native.ECOM2D2B1CoefficientsFromArray(values)
	return publicECOM2D2B1Coefficients(value), publicError(err)
}

// ECOM2D2B1CoefficientsNames returns the names.
func ECOM2D2B1CoefficientsNames(index uint64) (string, error) {
	value, err := native.ECOM2D2B1CoefficientsNames(index)
	return value, publicError(err)
}

// ECOM2D2B1CoefficientsParameterNames returns the parameter names.
func ECOM2D2B1CoefficientsParameterNames(index uint64) (string, error) {
	value, err := native.ECOM2D2B1CoefficientsParameterNames(index)
	return value, publicError(err)
}

// ToArray returns the coefficients in their documented order.
func (coefficients ECOM2D2B1Coefficients) ToArray() ([7]float64, error) {
	value, err := native.ECOM2D2B1CoefficientsToArray(nativeECOM2D2B1Coefficients(coefficients))
	return value, publicError(err)
}

// ECOM2D4B1CoefficientCount returns the ecom2 d4b1 coefficient count.
func ECOM2D4B1CoefficientCount() uint64 {
	return native.ECOM2D4B1CoefficientCount()
}

// ECOM2D4B1CoefficientsFromArray returns the from array.
func ECOM2D4B1CoefficientsFromArray(values [9]float64) (ECOM2D4B1Coefficients, error) {
	value, err := native.ECOM2D4B1CoefficientsFromArray(values)
	return publicECOM2D4B1Coefficients(value), publicError(err)
}

// ECOM2D4B1CoefficientsNames returns the names.
func ECOM2D4B1CoefficientsNames(index uint64) (string, error) {
	value, err := native.ECOM2D4B1CoefficientsNames(index)
	return value, publicError(err)
}

// ECOM2D4B1CoefficientsParameterNames returns the parameter names.
func ECOM2D4B1CoefficientsParameterNames(index uint64) (string, error) {
	value, err := native.ECOM2D4B1CoefficientsParameterNames(index)
	return value, publicError(err)
}

// ToArray returns the coefficients in their documented order.
func (coefficients ECOM2D4B1Coefficients) ToArray() ([9]float64, error) {
	value, err := native.ECOM2D4B1CoefficientsToArray(nativeECOM2D4B1Coefficients(coefficients))
	return value, publicError(err)
}

// AccelerationPartialsWithSun returns the acceleration partials with sun.
func (model *ECOM2SolarRadiationPressure) AccelerationPartialsWithSun(positionKm [3]float64, velocityKmS [3]float64, sunPositionKm [3]float64) (AccelerationPartials, error) {
	value, err := native.ECOM2SolarRadiationPressureAccelerationPartialsWithSun(nativeECOM2SolarRadiationPressureHandle(model), positionKm, velocityKmS, sunPositionKm)
	return publicAccelerationPartials(value), publicError(err)
}

// AccelerationPartialsWithSunMoon returns the acceleration partials with sun moon.
func (model *ECOM2SolarRadiationPressure) AccelerationPartialsWithSunMoon(positionKm [3]float64, velocityKmS [3]float64, sunPositionKm [3]float64, moonPositionKm [3]float64) (AccelerationPartials, error) {
	value, err := native.ECOM2SolarRadiationPressureAccelerationPartialsWithSunMoon(nativeECOM2SolarRadiationPressureHandle(model), positionKm, velocityKmS, sunPositionKm, moonPositionKm)
	return publicAccelerationPartials(value), publicError(err)
}

// AccelerationWithSun returns the acceleration with sun.
func (model *ECOM2SolarRadiationPressure) AccelerationWithSun(positionKm [3]float64, velocityKmS [3]float64, sunPositionKm [3]float64) ([3]float64, error) {
	value, err := native.ECOM2SolarRadiationPressureAccelerationWithSun(nativeECOM2SolarRadiationPressureHandle(model), positionKm, velocityKmS, sunPositionKm)
	return value, publicError(err)
}

// AccelerationWithSunMoon returns the acceleration with sun moon.
func (model *ECOM2SolarRadiationPressure) AccelerationWithSunMoon(positionKm [3]float64, velocityKmS [3]float64, sunPositionKm [3]float64, moonPositionKm [3]float64) ([3]float64, error) {
	value, err := native.ECOM2SolarRadiationPressureAccelerationWithSunMoon(nativeECOM2SolarRadiationPressureHandle(model), positionKm, velocityKmS, sunPositionKm, moonPositionKm)
	return value, publicError(err)
}

// ECOM2SolarRadiationPressureCoefficientCount returns the coefficient count.
func ECOM2SolarRadiationPressureCoefficientCount(nD uint64, nB uint64) (uint64, error) {
	value, err := native.ECOM2SolarRadiationPressureCoefficientCount(nD, nB)
	return value, publicError(err)
}

// Coefficients returns the coefficients.
func (model *ECOM2SolarRadiationPressure) Coefficients() ([]float64, error) {
	value, err := native.ECOM2SolarRadiationPressureCoefficients(nativeECOM2SolarRadiationPressureHandle(model))
	return value, publicError(err)
}

// ECOM2SolarRadiationPressureD2B1 returns the d2b1.
func ECOM2SolarRadiationPressureD2B1(coefficients ECOM2D2B1Coefficients, shadow RadiationShadowOptions) (*ECOM2SolarRadiationPressure, error) {
	value, err := native.ECOM2SolarRadiationPressureD2B1(nativeECOM2D2B1Coefficients(coefficients), nativeRadiationShadowOptions(shadow))
	if err != nil {
		return nil, publicError(err)
	}
	return publicECOM2SolarRadiationPressure(value), nil
}

// ECOM2SolarRadiationPressureD4B1 returns the d4b1.
func ECOM2SolarRadiationPressureD4B1(coefficients ECOM2D4B1Coefficients, shadow RadiationShadowOptions) (*ECOM2SolarRadiationPressure, error) {
	value, err := native.ECOM2SolarRadiationPressureD4B1(nativeECOM2D4B1Coefficients(coefficients), nativeRadiationShadowOptions(shadow))
	if err != nil {
		return nil, publicError(err)
	}
	return publicECOM2SolarRadiationPressure(value), nil
}

// EarthShadowModel returns the earth shadow model.
func (model *ECOM2SolarRadiationPressure) EarthShadowModel() (EarthShadowModel, error) {
	value, err := native.ECOM2SolarRadiationPressureEarthShadowModel(nativeECOM2SolarRadiationPressureHandle(model))
	return EarthShadowModel(value), publicError(err)
}

// Equal compares the underlying values.
func (left *ECOM2SolarRadiationPressure) Equal(right *ECOM2SolarRadiationPressure) (bool, error) {
	value, err := native.ECOM2SolarRadiationPressureEqual(nativeECOM2SolarRadiationPressureHandle(left), nativeECOM2SolarRadiationPressureHandle(right))
	return value, publicError(err)
}

// MoonEclipseEnabled returns the moon eclipse enabled.
func (model *ECOM2SolarRadiationPressure) MoonEclipseEnabled() (bool, error) {
	value, err := native.ECOM2SolarRadiationPressureMoonEclipseEnabled(nativeECOM2SolarRadiationPressureHandle(model))
	return value, publicError(err)
}

// NB returns the n b.
func (model *ECOM2SolarRadiationPressure) NB() (uint64, error) {
	value, err := native.ECOM2SolarRadiationPressureNB(nativeECOM2SolarRadiationPressureHandle(model))
	return value, publicError(err)
}

// ND returns the n d.
func (model *ECOM2SolarRadiationPressure) ND() (uint64, error) {
	value, err := native.ECOM2SolarRadiationPressureND(nativeECOM2SolarRadiationPressureHandle(model))
	return value, publicError(err)
}

// NewECOM2SolarRadiationPressure constructs an ECOM2SolarRadiationPressure from the supplied inputs.
func NewECOM2SolarRadiationPressure(nD uint64, nB uint64, coefficients []float64, shadow RadiationShadowOptions) (*ECOM2SolarRadiationPressure, error) {
	value, err := native.ECOM2SolarRadiationPressureNew(nD, nB, coefficients, nativeRadiationShadowOptions(shadow))
	if err != nil {
		return nil, publicError(err)
	}
	return publicECOM2SolarRadiationPressure(value), nil
}

// ECOM2SolarRadiationPressureParameterName returns the parameter name.
func ECOM2SolarRadiationPressureParameterName(nD uint64, nB uint64, index uint64) (string, error) {
	value, err := native.ECOM2SolarRadiationPressureParameterName(nD, nB, index)
	return value, publicError(err)
}

// ParameterPartialsWithSun returns the parameter partials with sun.
func (model *ECOM2SolarRadiationPressure) ParameterPartialsWithSun(positionKm [3]float64, velocityKmS [3]float64, sunPositionKm [3]float64) ([][3]float64, error) {
	value, err := native.ECOM2SolarRadiationPressureParameterPartialsWithSun(nativeECOM2SolarRadiationPressureHandle(model), positionKm, velocityKmS, sunPositionKm)
	return value, publicError(err)
}

// ParameterPartialsWithSunMoon returns the parameter partials with sun moon.
func (model *ECOM2SolarRadiationPressure) ParameterPartialsWithSunMoon(positionKm [3]float64, velocityKmS [3]float64, sunPositionKm [3]float64, moonPositionKm [3]float64) ([][3]float64, error) {
	value, err := native.ECOM2SolarRadiationPressureParameterPartialsWithSunMoon(nativeECOM2SolarRadiationPressureHandle(model), positionKm, velocityKmS, sunPositionKm, moonPositionKm)
	return value, publicError(err)
}

// ShadowFunction returns the shadow function.
func (model *ECOM2SolarRadiationPressure) ShadowFunction() (ShadowFunction, error) {
	value, err := native.ECOM2SolarRadiationPressureShadowFunction(nativeECOM2SolarRadiationPressureHandle(model))
	return ShadowFunction(value), publicError(err)
}

// WithEarthShadowModel returns a copy with the requested change.
func (model *ECOM2SolarRadiationPressure) WithEarthShadowModel(earthModel EarthShadowModel) (*ECOM2SolarRadiationPressure, error) {
	value, err := native.ECOM2SolarRadiationPressureWithEarthShadowModel(nativeECOM2SolarRadiationPressureHandle(model), uint32(earthModel))
	if err != nil {
		return nil, publicError(err)
	}
	return publicECOM2SolarRadiationPressure(value), nil
}

// WithMoonEclipse returns a copy with the requested change.
func (model *ECOM2SolarRadiationPressure) WithMoonEclipse(enabled bool) (*ECOM2SolarRadiationPressure, error) {
	value, err := native.ECOM2SolarRadiationPressureWithMoonEclipse(nativeECOM2SolarRadiationPressureHandle(model), enabled)
	if err != nil {
		return nil, publicError(err)
	}
	return publicECOM2SolarRadiationPressure(value), nil
}

// WithShadowFunction returns a copy with the requested change.
func (model *ECOM2SolarRadiationPressure) WithShadowFunction(shadowFunction ShadowFunction) (*ECOM2SolarRadiationPressure, error) {
	value, err := native.ECOM2SolarRadiationPressureWithShadowFunction(nativeECOM2SolarRadiationPressureHandle(model), native.ShadowFunction(shadowFunction))
	if err != nil {
		return nil, publicError(err)
	}
	return publicECOM2SolarRadiationPressure(value), nil
}

// ComputeECOMGeometry returns the ecom geometry.
func ComputeECOMGeometry(positionKm [3]float64, velocityKmS [3]float64, sunPositionKm [3]float64) (ECOMGeometry, error) {
	value, err := native.ComputeECOMGeometry(positionKm, velocityKmS, sunPositionKm)
	return publicECOMGeometry(value), publicError(err)
}

// ECOMGeometryWithShadow returns the ecom geometry with shadow.
func ECOMGeometryWithShadow(positionKm [3]float64, velocityKmS [3]float64, sunPositionKm [3]float64, moonPositionKm *[3]float64, earthShadowModel EarthShadowModel) (ECOMGeometry, error) {
	value, err := native.ECOMGeometryWithShadow(positionKm, velocityKmS, sunPositionKm, moonPositionKm, uint32(earthShadowModel))
	return publicECOMGeometry(value), publicError(err)
}

// EmpiricalHarmonicAll returns the empirical harmonic all.
func EmpiricalHarmonicAll() ([3]EmpiricalHarmonic, error) {
	value, err := native.EmpiricalHarmonicAll()
	return convertOrbitArrayEmpiricalHarmonic(value), publicError(err)
}

// EmpiricalRTNAccelerationAllTerms returns the all terms.
func EmpiricalRTNAccelerationAllTerms() (*EmpiricalRTNAcceleration, error) {
	value, err := native.EmpiricalRTNAccelerationAllTerms()
	if err != nil {
		return nil, publicError(err)
	}
	return publicEmpiricalRTNAcceleration(value), nil
}

// CoefficientKmS2 returns the coefficient km s2.
func (model *EmpiricalRTNAcceleration) CoefficientKmS2(axis RTNAxis, harmonic EmpiricalHarmonic) (OptionalOrbitScalar, error) {
	value, err := native.EmpiricalRTNAccelerationCoefficientKmS2(nativeEmpiricalRTNAccelerationHandle(model), native.RTNAxis(axis), native.EmpiricalHarmonic(harmonic))
	return publicOptionalOrbitScalar(value), publicError(err)
}

// EmpiricalRTNAccelerationConstantTerms returns the constant terms.
func EmpiricalRTNAccelerationConstantTerms() (*EmpiricalRTNAcceleration, error) {
	value, err := native.EmpiricalRTNAccelerationConstantTerms()
	if err != nil {
		return nil, publicError(err)
	}
	return publicEmpiricalRTNAcceleration(value), nil
}

// Equal compares the underlying values.
func (left *EmpiricalRTNAcceleration) Equal(right *EmpiricalRTNAcceleration) (bool, error) {
	value, err := native.EmpiricalRTNAccelerationEqual(nativeEmpiricalRTNAccelerationHandle(left), nativeEmpiricalRTNAccelerationHandle(right))
	return value, publicError(err)
}

// NewEmpiricalRTNAcceleration constructs an EmpiricalRTNAcceleration from the supplied inputs.
func NewEmpiricalRTNAcceleration(terms []EmpiricalRTNTerm) (*EmpiricalRTNAcceleration, error) {
	value, err := native.EmpiricalRTNAccelerationNew(orbitMap(terms, func(item EmpiricalRTNTerm) native.EmpiricalRTNTerm { return nativeEmpiricalRTNTerm(item) }))
	if err != nil {
		return nil, publicError(err)
	}
	return publicEmpiricalRTNAcceleration(value), nil
}

// EmpiricalRTNAccelerationParameterName returns the parameter name.
func EmpiricalRTNAccelerationParameterName(axis RTNAxis, harmonic EmpiricalHarmonic) (string, error) {
	value, err := native.EmpiricalRTNAccelerationParameterName(native.RTNAxis(axis), native.EmpiricalHarmonic(harmonic))
	return value, publicError(err)
}

// Terms returns the terms.
func (model *EmpiricalRTNAcceleration) Terms() ([]EmpiricalRTNTerm, error) {
	value, err := native.EmpiricalRTNAccelerationTerms(nativeEmpiricalRTNAccelerationHandle(model))
	return orbitMap(value, func(value native.EmpiricalRTNTerm) EmpiricalRTNTerm { return publicEmpiricalRTNTerm(value) }), publicError(err)
}

// WithTerm returns a copy with the requested change.
func (model *EmpiricalRTNAcceleration) WithTerm(term EmpiricalRTNTerm) (*EmpiricalRTNAcceleration, error) {
	value, err := native.EmpiricalRTNAccelerationWithTerm(nativeEmpiricalRTNAccelerationHandle(model), nativeEmpiricalRTNTerm(term))
	if err != nil {
		return nil, publicError(err)
	}
	return publicEmpiricalRTNAcceleration(value), nil
}

// EmpiricalRTNAccelerationWithZeroTerms enables the specified terms with zero coefficients.
func EmpiricalRTNAccelerationWithZeroTerms(terms []EmpiricalRTNTermSelector) (*EmpiricalRTNAcceleration, error) {
	value, err := native.EmpiricalRTNAccelerationWithZeroTerms(orbitMap(terms, func(item EmpiricalRTNTermSelector) native.EmpiricalRTNTermSelector {
		return nativeEmpiricalRTNTermSelector(item)
	}))
	if err != nil {
		return nil, publicError(err)
	}
	return publicEmpiricalRTNAcceleration(value), nil
}

// Acceleration returns the acceleration.
func (model *ForceModel) Acceleration(state CartesianState, context *PropagationContext) ([3]float64, error) {
	value, err := native.ForceModelAcceleration(nativeForceModelValue(model), nativeCartesian(state), nativePropagationContext(context))
	return value, publicError(err)
}

// AccelerationPartials returns the acceleration partials.
func (model *ForceModel) AccelerationPartials(state CartesianState, parameters []float64, context *PropagationContext) (AccelerationPartials, error) {
	value, err := native.ForceModelAccelerationPartials(nativeForceModelValue(model), nativeCartesian(state), parameters, nativePropagationContext(context))
	return publicAccelerationPartials(value), publicError(err)
}

// AccelerationWithParameters returns the acceleration with parameters.
func (model *ForceModel) AccelerationWithParameters(state CartesianState, parameters []float64, context *PropagationContext) ([3]float64, error) {
	value, err := native.ForceModelAccelerationWithParameters(nativeForceModelValue(model), nativeCartesian(state), parameters, nativePropagationContext(context))
	return value, publicError(err)
}

// ParameterCount returns the parameter count.
func (model *ForceModel) ParameterCount() (uint64, error) {
	value, err := native.ForceModelParameterCount(nativeForceModelValue(model))
	return value, publicError(err)
}

// ParameterPartials returns the parameter partials.
func (model *ForceModel) ParameterPartials(state CartesianState, parameters []float64, context *PropagationContext) ([][3]float64, error) {
	value, err := native.ForceModelParameterPartials(nativeForceModelValue(model), nativeCartesian(state), parameters, nativePropagationContext(context))
	return value, publicError(err)
}

// NewJ2Gravity constructs a J2Gravity from the supplied inputs.
func NewJ2Gravity(muKm3S2 float64, referenceRadiusKm float64, j2 float64) (*J2Gravity, error) {
	value, err := native.J2GravityNew(muKm3S2, referenceRadiusKm, j2)
	if err != nil {
		return nil, publicError(err)
	}
	return publicJ2Gravity(value), nil
}

// DefaultRadiationShadowOptions selects apparent-disc lighting, WGS84 and lunar occultation.
func DefaultRadiationShadowOptions() RadiationShadowOptions {
	value := native.RadiationShadowOptionsDefault()
	return publicRadiationShadowOptions(value)
}

// RTNAxisAll returns the RTN axis all.
func RTNAxisAll() ([3]RTNAxis, error) {
	value, err := native.RTNAxisAll()
	return convertOrbitArrayRTNAxis(value), publicError(err)
}

// NewSchwarzschildRelativity constructs a SchwarzschildRelativity from the supplied inputs.
func NewSchwarzschildRelativity(muKm3S2 float64, cKmS float64) (*SchwarzschildRelativity, error) {
	value, err := native.SchwarzschildRelativityNew(muKm3S2, cKmS)
	if err != nil {
		return nil, publicError(err)
	}
	return publicSchwarzschildRelativity(value), nil
}

// AccelerationWithSunMoon returns the acceleration with sun moon.
func (model *SolarRadiationPressureForce) AccelerationWithSunMoon(positionKm [3]float64, sunPositionKm [3]float64, moonPositionKm [3]float64) ([3]float64, error) {
	value, err := native.SolarRadiationPressureAccelerationWithSunMoon(nativeSolarRadiationPressureForceHandle(model), positionKm, sunPositionKm, moonPositionKm)
	return value, publicError(err)
}

// AreaToMassM2Kg returns the area to mass m2 kg.
func (model *SolarRadiationPressureForce) AreaToMassM2Kg() (float64, error) {
	value, err := native.SolarRadiationPressureAreaToMassM2Kg(nativeSolarRadiationPressureForceHandle(model))
	return value, publicError(err)
}

// AUKm returns the au km.
func (model *SolarRadiationPressureForce) AUKm() (float64, error) {
	value, err := native.SolarRadiationPressureAUKm(nativeSolarRadiationPressureForceHandle(model))
	return value, publicError(err)
}

// CR returns the cr.
func (model *SolarRadiationPressureForce) CR() (float64, error) {
	value, err := native.SolarRadiationPressureCR(nativeSolarRadiationPressureForceHandle(model))
	return value, publicError(err)
}

// SolarRadiationPressureCRParameter returns the cr parameter.
func SolarRadiationPressureCRParameter() (string, error) {
	value, err := native.SolarRadiationPressureCRParameter()
	return value, publicError(err)
}

// EarthShadowModel returns the earth shadow model.
func (model *SolarRadiationPressureForce) EarthShadowModel() (EarthShadowModel, error) {
	value, err := native.SolarRadiationPressureEarthShadowModel(nativeSolarRadiationPressureForceHandle(model))
	return EarthShadowModel(value), publicError(err)
}

// MoonEclipseEnabled returns the moon eclipse enabled.
func (model *SolarRadiationPressureForce) MoonEclipseEnabled() (bool, error) {
	value, err := native.SolarRadiationPressureMoonEclipseEnabled(nativeSolarRadiationPressureForceHandle(model))
	return value, publicError(err)
}

// NewSolarRadiationPressureForce constructs a SolarRadiationPressureForce from the supplied inputs.
func NewSolarRadiationPressureForce(cr float64, areaToMassM2Kg float64, shadow RadiationShadowOptions) (*SolarRadiationPressureForce, error) {
	value, err := native.SolarRadiationPressureNew(cr, areaToMassM2Kg, nativeRadiationShadowOptions(shadow))
	if err != nil {
		return nil, publicError(err)
	}
	return publicSolarRadiationPressureForce(value), nil
}

// ShadowFunction returns the shadow function.
func (model *SolarRadiationPressureForce) ShadowFunction() (ShadowFunction, error) {
	value, err := native.SolarRadiationPressureShadowFunction(nativeSolarRadiationPressureForceHandle(model))
	return ShadowFunction(value), publicError(err)
}

// WithEarthShadowModel returns a copy with the requested change.
func (model *SolarRadiationPressureForce) WithEarthShadowModel(earthModel EarthShadowModel) (*SolarRadiationPressureForce, error) {
	value, err := native.SolarRadiationPressureWithEarthShadowModel(nativeSolarRadiationPressureForceHandle(model), uint32(earthModel))
	if err != nil {
		return nil, publicError(err)
	}
	return publicSolarRadiationPressureForce(value), nil
}

// WithMoonEclipse returns a copy with the requested change.
func (model *SolarRadiationPressureForce) WithMoonEclipse(enabled bool) (*SolarRadiationPressureForce, error) {
	value, err := native.SolarRadiationPressureWithMoonEclipse(nativeSolarRadiationPressureForceHandle(model), enabled)
	if err != nil {
		return nil, publicError(err)
	}
	return publicSolarRadiationPressureForce(value), nil
}

// WithShadowFunction returns a copy with the requested change.
func (model *SolarRadiationPressureForce) WithShadowFunction(shadowFunction ShadowFunction) (*SolarRadiationPressureForce, error) {
	value, err := native.SolarRadiationPressureWithShadowFunction(nativeSolarRadiationPressureForceHandle(model), native.ShadowFunction(shadowFunction))
	if err != nil {
		return nil, publicError(err)
	}
	return publicSolarRadiationPressureForce(value), nil
}

// NewSolidEarthPoleTideGravity constructs a SolidEarthPoleTideGravity from the supplied inputs.
func NewSolidEarthPoleTideGravity() (*SolidEarthPoleTideGravity, error) {
	value, err := native.SolidEarthPoleTideGravityNew()
	if err != nil {
		return nil, publicError(err)
	}
	return publicSolidEarthPoleTideGravity(value), nil
}

// NewSolidEarthTideGravity constructs a SolidEarthTideGravity from the supplied inputs.
func NewSolidEarthTideGravity(tideSystem GravityTideSystem) (*SolidEarthTideGravity, error) {
	value, err := native.SolidEarthTideGravityNew(native.GravityTideSystem(tideSystem))
	if err != nil {
		return nil, publicError(err)
	}
	return publicSolidEarthTideGravity(value), nil
}

// NewSourcedDragForce constructs a SourcedDragForce from the supplied inputs.
func NewSourcedDragForce(bcFactorM2Kg float64, cutoffAltitudeKm float64, table *SpaceWeatherTable) (*SourcedDragForce, error) {
	value, err := native.SourcedDragForceNew(bcFactorM2Kg, cutoffAltitudeKm, nativeSpaceWeatherTableHandle(table))
	if err != nil {
		return nil, publicError(err)
	}
	return publicSourcedDragForce(value), nil
}

// NewThirdBodyGravity constructs a ThirdBodyGravity from the supplied inputs.
func NewThirdBodyGravity(sun bool, moon bool) (*ThirdBodyGravity, error) {
	value, err := native.ThirdBodyGravityNew(sun, moon)
	if err != nil {
		return nil, publicError(err)
	}
	return publicThirdBodyGravity(value), nil
}

// NewTwoBodyGravity constructs a TwoBodyGravity from the supplied inputs.
func NewTwoBodyGravity(muKm3S2 float64) (*TwoBodyGravity, error) {
	value, err := native.TwoBodyGravityNew(muKm3S2)
	if err != nil {
		return nil, publicError(err)
	}
	return publicTwoBodyGravity(value), nil
}

// NewZonalGravity constructs a ZonalGravity from the supplied inputs.
func NewZonalGravity(muKm3S2 float64, referenceRadiusKm float64, maxDegree uint8, tideSystem GravityTideSystem) (*ZonalGravity, error) {
	value, err := native.ZonalGravityNew(muKm3S2, referenceRadiusKm, maxDegree, native.GravityTideSystem(tideSystem))
	if err != nil {
		return nil, publicError(err)
	}
	return publicZonalGravity(value), nil
}

// Compare orders RTN axes as radial, along-track and cross-track.
func (axis RTNAxis) Compare(other RTNAxis) int {
	return cmp.Compare(axis, other)
}

// Compare orders empirical harmonics as constant, cosine and sine.
func (harmonic EmpiricalHarmonic) Compare(other EmpiricalHarmonic) int {
	return cmp.Compare(harmonic, other)
}
