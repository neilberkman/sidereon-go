//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

type PppResidualScreenRemoval struct {
	EpochIndex  int
	AmbiguityID string
}

func (*PppFloatSolution) SolvedEpochIndices() ([]int, error) { return nil, ErrUnavailable }
func (*PppFixedSolution) SolvedEpochIndices() ([]int, error) { return nil, ErrUnavailable }
func (*PppFloatSolution) EpochClocksM() ([]float64, error)   { return nil, ErrUnavailable }
func (*PppFixedSolution) EpochClocksM() ([]float64, error)   { return nil, ErrUnavailable }
func (*PppFloatSolution) ResidualScreenRemovals() ([]PppResidualScreenRemoval, error) {
	return nil, ErrUnavailable
}
