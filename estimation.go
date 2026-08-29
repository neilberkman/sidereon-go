package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// NISGate is the normalized-innovation-squared gate result from C.
type NISGate struct {
	NIS       float64
	Threshold float64
	InGate    bool
	DOF       uint64
}

// NIS computes the scalar normalized innovation squared statistic.
func NIS(innovation, innovationVariance float64) (float64, error) {
	value, err := native.NIS(innovation, innovationVariance)
	return value, publicError(err)
}

// ExpectedNIS returns the C expected NIS for a measurement degrees of freedom.
func ExpectedNIS(dof uint64) (float64, error) {
	value, err := native.NISExpected(dof)
	return value, publicError(err)
}

// NISThreshold returns the C chi-square gate threshold for a confidence in
// (0, 1) and positive degrees of freedom.
func NISThreshold(dof uint64, confidence float64) (float64, error) {
	value, err := native.NISThreshold(dof, confidence)
	return value, publicError(err)
}

// TestNISGate tests one innovation against a C chi-square NIS gate.
func TestNISGate(innovation, innovationVariance float64, dof uint64, confidence float64) (NISGate, error) {
	value, err := native.ComputeNISGate(innovation, innovationVariance, dof, confidence)
	return NISGate{NIS: value.NIS, Threshold: value.Threshold, InGate: value.InGate, DOF: value.DOF}, publicError(err)
}

// ChiSquareInverse returns the C inverse chi-square quantile.
func ChiSquareInverse(probability float64, dof uint64) (float64, error) {
	value, err := native.ChiSquareInverse(probability, dof)
	return value, publicError(err)
}
