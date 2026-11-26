//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

type BrouwerLyddanePropagator struct{}

func (*BrouwerLyddanePropagator) Close() error { return nil }

type DEEphemeris struct{}

func (*DEEphemeris) Close() error { return nil }

type EcksteinHechlerPropagator struct{}

func (*EcksteinHechlerPropagator) Close() error { return nil }

type ForceModel struct{}

func (*ForceModel) Close() error { return nil }

type GNSSAttitudeLaw struct{}

func (*GNSSAttitudeLaw) Close() error { return nil }

type ICGEMGravityField struct{}

func (*ICGEMGravityField) Close() error { return nil }

type IERS2010EarthOrientationProvider struct{}

func (*IERS2010EarthOrientationProvider) Close() error { return nil }

type OrbitParameterFit struct{}

func (*OrbitParameterFit) Close() error { return nil }

type OrbitScoreReport struct{}

func (*OrbitScoreReport) Close() error { return nil }

type OrbitSource struct{}

func (*OrbitSource) Close() error { return nil }

type SphericalHarmonicGravity struct{}

func (*SphericalHarmonicGravity) Close() error { return nil }

type VariationalTrajectory struct{}

func (*VariationalTrajectory) Close() error { return nil }
