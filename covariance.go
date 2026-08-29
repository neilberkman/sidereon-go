package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// Covariance6 is a value-owned row-major six-by-six covariance matrix.
// Methods delegate validation and transformations to the C library.
type Covariance6 struct {
	Values [6][6]float64
}

// CovarianceValidation contains the C library's matrix validation results.
type CovarianceValidation struct {
	Symmetric            bool
	PositiveSemidefinite bool
}

// CovarianceFromDiagonal builds a covariance from exactly six diagonal
// variances. The input is copied before crossing the cgo boundary.
func CovarianceFromDiagonal(diagonal []float64) (Covariance6, error) {
	values, err := native.CovarianceFromDiagonal(diagonal)
	return Covariance6{Values: values}, publicError(err)
}

// Validate checks symmetry and positive semidefiniteness.
func (c Covariance6) Validate() (CovarianceValidation, error) {
	validation, err := native.CovarianceValidate(c.Values)
	return CovarianceValidation{
		Symmetric:            validation.Symmetric,
		PositiveSemidefinite: validation.PositiveSemidefinite,
	}, publicError(err)
}

// ToMeters converts a kilometre-based covariance to metre units.
func (c Covariance6) ToMeters() (Covariance6, error) {
	values, err := native.CovarianceKmToM(c.Values)
	return Covariance6{Values: values}, publicError(err)
}

// ToKilometers converts a metre-based covariance to kilometre units.
func (c Covariance6) ToKilometers() (Covariance6, error) {
	values, err := native.CovarianceMToKm(c.Values)
	return Covariance6{Values: values}, publicError(err)
}

// Interpolate returns the C library's PSD-safe interpolation between c and
// other at parameter u.
func (c Covariance6) Interpolate(other Covariance6, u float64) (Covariance6, error) {
	values, err := native.CovarianceInterpolate(c.Values, other.Values, u)
	return Covariance6{Values: values}, publicError(err)
}
