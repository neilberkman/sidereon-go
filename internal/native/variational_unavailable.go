//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

func IntegratorOptionsDefault() IntegratorOptions { return IntegratorOptions{} }

func VariationalEpochFromTDBSeconds(_ float64, _ float64) VariationalEpoch { return VariationalEpoch{} }

func VariationalTrajectoryCount(_ *VariationalTrajectory) (uint64, error) { return 0, ErrUnavailable }

func VariationalTrajectorySensitivity(_ *VariationalTrajectory, _ uint64) ([]float64, error) {
	return nil, ErrUnavailable
}

func VariationalTrajectoryState(_ *VariationalTrajectory, _ uint64) (VariationalState, error) {
	return VariationalState{}, ErrUnavailable
}
