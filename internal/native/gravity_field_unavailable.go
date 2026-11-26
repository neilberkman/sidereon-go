//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

func ICGEMGravityFieldCoefficientAt(_ *ICGEMGravityField, _ uint16, _ uint16, _ float64) (OptionalHarmonicCoefficient, error) {
	return OptionalHarmonicCoefficient{}, ErrUnavailable
}

func ICGEMGravityFieldEqual(_ *ICGEMGravityField, _ *ICGEMGravityField) (bool, error) {
	return false, ErrUnavailable
}

func ICGEMGravityFieldGravity(_ *ICGEMGravityField, _ uint16, _ uint16) (*SphericalHarmonicGravity, error) {
	return nil, ErrUnavailable
}

func ICGEMGravityFieldGravityInTideSystem(_ *ICGEMGravityField, _ uint16, _ uint16, _ GravityTideSystem) (*SphericalHarmonicGravity, error) {
	return nil, ErrUnavailable
}

func ICGEMGravityFieldHeader(_ *ICGEMGravityField) (ICGEMHeader, error) {
	return ICGEMHeader{}, ErrUnavailable
}

func ICGEMGravityFieldIsTimeVariable(_ *ICGEMGravityField) (bool, error) {
	return false, ErrUnavailable
}

func ICGEMGravityFieldModelName(_ *ICGEMGravityField) (string, error) { return "", ErrUnavailable }

func ICGEMGravityFieldParse(_ []byte) (*ICGEMGravityField, error) { return nil, ErrUnavailable }

func ICGEMGravityFieldParseTruncated(_ []byte, _ uint16, _ uint16) (*ICGEMGravityField, error) {
	return nil, ErrUnavailable
}

func ICGEMGravityFieldStaticCoefficient(_ *ICGEMGravityField, _ uint16, _ uint16) (OptionalHarmonicCoefficient, error) {
	return OptionalHarmonicCoefficient{}, ErrUnavailable
}

func ICGEMGravityFieldStaticSigmas(_ *ICGEMGravityField, _ uint16, _ uint16) (ICGEMSigmas, error) {
	return ICGEMSigmas{}, ErrUnavailable
}

func ICGEMGravityFieldStoredMaxDegree(_ *ICGEMGravityField) (uint16, error) { return 0, ErrUnavailable }

func ICGEMGravityFieldStoredMaxOrder(_ *ICGEMGravityField) (uint16, error) { return 0, ErrUnavailable }

func ICGEMGravityFieldTimeVariableTerms(_ *ICGEMGravityField) ([]ICGEMTimeVariableTerm, error) {
	return nil, ErrUnavailable
}

func ICGEMGravityFieldWithAssumedTideSystem(_ *ICGEMGravityField, _ GravityTideSystem) (*ICGEMGravityField, error) {
	return nil, ErrUnavailable
}

func SphericalHarmonicGravityAsForceModel(_ *SphericalHarmonicGravity) (*ForceModel, error) {
	return nil, ErrUnavailable
}

func SphericalHarmonicGravityBodyFixedAccelerationAtKmS2(_ *SphericalHarmonicGravity, _ [3]float64, _ float64) ([3]float64, error) {
	return [3]float64{}, ErrUnavailable
}

func SphericalHarmonicGravityBodyFixedAccelerationKmS2(_ *SphericalHarmonicGravity, _ [3]float64) ([3]float64, error) {
	return [3]float64{}, ErrUnavailable
}

func SphericalHarmonicGravityBodyFixedPotentialAtKm2S2(_ *SphericalHarmonicGravity, _ [3]float64, _ float64) (float64, error) {
	return 0, ErrUnavailable
}

func SphericalHarmonicGravityBodyFixedPotentialKm2S2(_ *SphericalHarmonicGravity, _ [3]float64) (float64, error) {
	return 0, ErrUnavailable
}

func SphericalHarmonicGravityCoefficientAt(_ *SphericalHarmonicGravity, _ uint16, _ uint16, _ float64) (OptionalHarmonicCoefficient, error) {
	return OptionalHarmonicCoefficient{}, ErrUnavailable
}

func SphericalHarmonicGravityEgm96(_ uint16, _ uint16) (*SphericalHarmonicGravity, error) {
	return nil, ErrUnavailable
}

func SphericalHarmonicGravityEqual(_ *SphericalHarmonicGravity, _ *SphericalHarmonicGravity) (bool, error) {
	return false, ErrUnavailable
}

func SphericalHarmonicGravityIsTimeVariable(_ *SphericalHarmonicGravity) (bool, error) {
	return false, ErrUnavailable
}
