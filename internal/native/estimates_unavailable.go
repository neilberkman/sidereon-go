//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

func ForceModelParameters(*ForceModel) ([]ForceParameter, error)    { return nil, ErrUnavailable }
func DEEphemerisConstants(*DEEphemeris) (map[string]float64, error) { return nil, ErrUnavailable }
func CompositeForceModelNew([]*ForceModel) (*ForceModel, error)     { return nil, ErrUnavailable }
func PropagateVariational(CartesianState, []VariationalEpoch, VariationalEquations, IntegratorOptions, *PropagationContext) (*VariationalTrajectory, error) {
	return nil, ErrUnavailable
}
