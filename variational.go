package sidereon

import (
	"sidereon.dev/go/v4/internal/native"
)

// DefaultIntegratorOptions returns the integrator options default.
func DefaultIntegratorOptions() IntegratorOptions {
	value := native.IntegratorOptionsDefault()
	return publicIntegratorOptions(value)
}

// VariationalEpochFromTDBSeconds returns the variational epoch from TDB seconds.
func VariationalEpochFromTDBSeconds(initialTDBJ2000S float64, tdbJ2000S float64) VariationalEpoch {
	value := native.VariationalEpochFromTDBSeconds(initialTDBJ2000S, tdbJ2000S)
	return publicVariationalEpoch(value)
}

// Count returns the count.
func (trajectory *VariationalTrajectory) Count() (uint64, error) {
	value, err := native.VariationalTrajectoryCount(nativeVariationalTrajectoryHandle(trajectory))
	return value, publicError(err)
}

// Sensitivity returns the sensitivity.
func (trajectory *VariationalTrajectory) Sensitivity(index uint64) ([]float64, error) {
	value, err := native.VariationalTrajectorySensitivity(nativeVariationalTrajectoryHandle(trajectory), index)
	return value, publicError(err)
}

// State returns the state.
func (trajectory *VariationalTrajectory) State(index uint64) (VariationalState, error) {
	value, err := native.VariationalTrajectoryState(nativeVariationalTrajectoryHandle(trajectory), index)
	return publicVariationalState(value), publicError(err)
}
