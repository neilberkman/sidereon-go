package sidereon

import (
	"sidereon.dev/go/v4/internal/native"
)

// CoefficientAt returns the coefficient at.
func (field *ICGEMGravityField) CoefficientAt(degree uint16, order uint16, epochTTSeconds float64) (OptionalHarmonicCoefficient, error) {
	value, err := native.ICGEMGravityFieldCoefficientAt(nativeICGEMGravityFieldHandle(field), degree, order, epochTTSeconds)
	return publicOptionalHarmonicCoefficient(value), publicError(err)
}

// Equal compares the underlying values.
func (left *ICGEMGravityField) Equal(right *ICGEMGravityField) (bool, error) {
	value, err := native.ICGEMGravityFieldEqual(nativeICGEMGravityFieldHandle(left), nativeICGEMGravityFieldHandle(right))
	return value, publicError(err)
}

// Gravity builds the requested degree and order using the field’s tide system.
func (field *ICGEMGravityField) Gravity(maxDegree uint16, maxOrder uint16) (*SphericalHarmonicGravity, error) {
	value, err := native.ICGEMGravityFieldGravity(nativeICGEMGravityFieldHandle(field), maxDegree, maxOrder)
	if err != nil {
		return nil, publicError(err)
	}
	return publicSphericalHarmonicGravity(value)
}

// GravityInTideSystem returns the gravity in tide system.
func (field *ICGEMGravityField) GravityInTideSystem(maxDegree uint16, maxOrder uint16, tideSystem GravityTideSystem) (*SphericalHarmonicGravity, error) {
	value, err := native.ICGEMGravityFieldGravityInTideSystem(nativeICGEMGravityFieldHandle(field), maxDegree, maxOrder, native.GravityTideSystem(tideSystem))
	if err != nil {
		return nil, publicError(err)
	}
	return publicSphericalHarmonicGravity(value)
}

// Header returns the header.
func (field *ICGEMGravityField) Header() (ICGEMHeader, error) {
	value, err := native.ICGEMGravityFieldHeader(nativeICGEMGravityFieldHandle(field))
	return publicICGEMHeader(value), publicError(err)
}

// IsTimeVariable returns the is time variable.
func (field *ICGEMGravityField) IsTimeVariable() (bool, error) {
	value, err := native.ICGEMGravityFieldIsTimeVariable(nativeICGEMGravityFieldHandle(field))
	return value, publicError(err)
}

// ModelName returns the model name.
func (field *ICGEMGravityField) ModelName() (string, error) {
	value, err := native.ICGEMGravityFieldModelName(nativeICGEMGravityFieldHandle(field))
	return value, publicError(err)
}

// ParseICGEMGravityField parses static and time-variable harmonic coefficients.
func ParseICGEMGravityField(data string) (*ICGEMGravityField, error) {
	value, err := native.ICGEMGravityFieldParse([]byte(data))
	if err != nil {
		return nil, publicError(err)
	}
	return publicICGEMGravityField(value), nil
}

// ParseICGEMGravityFieldTruncated parses the field up to the requested degree and order.
func ParseICGEMGravityFieldTruncated(data string, maxDegree uint16, maxOrder uint16) (*ICGEMGravityField, error) {
	value, err := native.ICGEMGravityFieldParseTruncated([]byte(data), maxDegree, maxOrder)
	if err != nil {
		return nil, publicError(err)
	}
	return publicICGEMGravityField(value), nil
}

// StaticCoefficient returns the static coefficient.
func (field *ICGEMGravityField) StaticCoefficient(degree uint16, order uint16) (OptionalHarmonicCoefficient, error) {
	value, err := native.ICGEMGravityFieldStaticCoefficient(nativeICGEMGravityFieldHandle(field), degree, order)
	return publicOptionalHarmonicCoefficient(value), publicError(err)
}

// StaticSigmas returns the static sigmas.
func (field *ICGEMGravityField) StaticSigmas(degree uint16, order uint16) (ICGEMSigmas, error) {
	value, err := native.ICGEMGravityFieldStaticSigmas(nativeICGEMGravityFieldHandle(field), degree, order)
	return publicICGEMSigmas(value), publicError(err)
}

// StoredMaxDegree returns the stored max degree.
func (field *ICGEMGravityField) StoredMaxDegree() (uint16, error) {
	value, err := native.ICGEMGravityFieldStoredMaxDegree(nativeICGEMGravityFieldHandle(field))
	return value, publicError(err)
}

// StoredMaxOrder returns the stored max order.
func (field *ICGEMGravityField) StoredMaxOrder() (uint16, error) {
	value, err := native.ICGEMGravityFieldStoredMaxOrder(nativeICGEMGravityFieldHandle(field))
	return value, publicError(err)
}

// TimeVariableTerms returns the time variable terms.
func (field *ICGEMGravityField) TimeVariableTerms() ([]ICGEMTimeVariableTerm, error) {
	value, err := native.ICGEMGravityFieldTimeVariableTerms(nativeICGEMGravityFieldHandle(field))
	return orbitMap(value, func(value native.ICGEMTimeVariableTerm) ICGEMTimeVariableTerm {
		return publicICGEMTimeVariableTerm(value)
	}), publicError(err)
}

// WithAssumedTideSystem returns a copy with the requested change.
func (field *ICGEMGravityField) WithAssumedTideSystem(tideSystem GravityTideSystem) (*ICGEMGravityField, error) {
	value, err := native.ICGEMGravityFieldWithAssumedTideSystem(nativeICGEMGravityFieldHandle(field), native.GravityTideSystem(tideSystem))
	if err != nil {
		return nil, publicError(err)
	}
	return publicICGEMGravityField(value), nil
}

// AsForceModel returns the as force model.
func (gravity *SphericalHarmonicGravity) AsForceModel() (*ForceModel, error) {
	value, err := native.SphericalHarmonicGravityAsForceModel(nativeSphericalHarmonicGravityHandle(gravity))
	if err != nil {
		return nil, publicError(err)
	}
	return publicForceModel(value), nil
}

// BodyFixedAccelerationAtKmS2 returns the body fixed acceleration at km s2.
func (gravity *SphericalHarmonicGravity) BodyFixedAccelerationAtKmS2(positionKm [3]float64, epochTTSeconds float64) ([3]float64, error) {
	value, err := native.SphericalHarmonicGravityBodyFixedAccelerationAtKmS2(nativeSphericalHarmonicGravityHandle(gravity), positionKm, epochTTSeconds)
	return value, publicError(err)
}

// BodyFixedAccelerationKmS2 returns the body fixed acceleration km s2.
func (gravity *SphericalHarmonicGravity) BodyFixedAccelerationKmS2(positionKm [3]float64) ([3]float64, error) {
	value, err := native.SphericalHarmonicGravityBodyFixedAccelerationKmS2(nativeSphericalHarmonicGravityHandle(gravity), positionKm)
	return value, publicError(err)
}

// BodyFixedPotentialAtKm2S2 returns the body fixed potential at km2 s2.
func (gravity *SphericalHarmonicGravity) BodyFixedPotentialAtKm2S2(positionKm [3]float64, epochTTSeconds float64) (float64, error) {
	value, err := native.SphericalHarmonicGravityBodyFixedPotentialAtKm2S2(nativeSphericalHarmonicGravityHandle(gravity), positionKm, epochTTSeconds)
	return value, publicError(err)
}

// BodyFixedPotentialKm2S2 returns the body fixed potential km2 s2.
func (gravity *SphericalHarmonicGravity) BodyFixedPotentialKm2S2(positionKm [3]float64) (float64, error) {
	value, err := native.SphericalHarmonicGravityBodyFixedPotentialKm2S2(nativeSphericalHarmonicGravityHandle(gravity), positionKm)
	return value, publicError(err)
}

// CoefficientAt returns the coefficient at.
func (gravity *SphericalHarmonicGravity) CoefficientAt(degree uint16, order uint16, epochTTSeconds float64) (OptionalHarmonicCoefficient, error) {
	value, err := native.SphericalHarmonicGravityCoefficientAt(nativeSphericalHarmonicGravityHandle(gravity), degree, order, epochTTSeconds)
	return publicOptionalHarmonicCoefficient(value), publicError(err)
}

// SphericalHarmonicGravityEGM96 returns the egm96.
func SphericalHarmonicGravityEGM96(maxDegree uint16, maxOrder uint16) (*SphericalHarmonicGravity, error) {
	value, err := native.SphericalHarmonicGravityEgm96(maxDegree, maxOrder)
	if err != nil {
		return nil, publicError(err)
	}
	return publicSphericalHarmonicGravity(value)
}

// Equal compares the underlying values.
func (left *SphericalHarmonicGravity) Equal(right *SphericalHarmonicGravity) (bool, error) {
	value, err := native.SphericalHarmonicGravityEqual(nativeSphericalHarmonicGravityHandle(left), nativeSphericalHarmonicGravityHandle(right))
	return value, publicError(err)
}

// IsTimeVariable returns the is time variable.
func (gravity *SphericalHarmonicGravity) IsTimeVariable() (bool, error) {
	value, err := native.SphericalHarmonicGravityIsTimeVariable(nativeSphericalHarmonicGravityHandle(gravity))
	return value, publicError(err)
}
