//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

func OpenPreciseInterpolantArtifactFile(string) (*PreciseInterpolantArtifact, uint32, error) {
	return nil, 0, protocolUnavailable()
}

func OpenPreciseInterpolantArtifactAttestedFile(string, uint64) (*PreciseInterpolantArtifact, uint32, error) {
	return nil, 0, protocolUnavailable()
}
