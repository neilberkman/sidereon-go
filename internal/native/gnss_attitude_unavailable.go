//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

func BeiDou2OnToYsBetaRad() float64 { return 0 }

func BeiDou2YsToOnBetaRad() float64 { return 0 }

func BeiDou3CASTIGSOTurn() CosineTurnParameters { return CosineTurnParameters{} }

func BeiDou3CASTMEOTurn() CosineTurnParameters { return CosineTurnParameters{} }

func BeiDou3SECMBeta0Rad() float64 { return 0 }

func BeiDouManufacturerCompare(_ BeiDouManufacturer, _ BeiDouManufacturer) (int32, error) {
	return 0, ErrUnavailable
}

func BeiDouOrbitClassCompare(_ BeiDouOrbitClass, _ BeiDouOrbitClass) (int32, error) {
	return 0, ErrUnavailable
}

func ComputeDYBAxes(_ [3]float64, _ [3]float64) (DYBAxes, error) { return DYBAxes{}, ErrUnavailable }

func GalileoFOCTurn() CosineTurnParameters { return CosineTurnParameters{} }

func GalileoIOVBetaXRad() float64 { return 0 }

func GalileoIOVBetaYRad() float64 { return 0 }

func GLONASSMDefaultYawRateRadS() float64 { return 0 }

func GLONASSMShadowLimitRad() float64 { return 0 }

func GNSSAttitudeLawBeiDou3MEOUnknownManufacturer(_ CosineTurnParameters, _ float64) (*GNSSAttitudeLaw, error) {
	return nil, ErrUnavailable
}

func GNSSAttitudeLawBeiDouOrbitNormalSwitch(_ float64, _ float64) (*GNSSAttitudeLaw, error) {
	return nil, ErrUnavailable
}

func GNSSAttitudeLawCosineTurn(_ CosineTurnParameters) (*GNSSAttitudeLaw, error) {
	return nil, ErrUnavailable
}

func GNSSAttitudeLawEqual(_ *GNSSAttitudeLaw, _ *GNSSAttitudeLaw) (bool, error) {
	return false, ErrUnavailable
}

func GNSSAttitudeLawFixedBetaYawSteering(_ float64) (*GNSSAttitudeLaw, error) {
	return nil, ErrUnavailable
}

func GNSSAttitudeLawForBlock(_ GNSSBlock, _ uint8) (*GNSSAttitudeLaw, error) {
	return nil, ErrUnavailable
}

func GNSSAttitudeLawGalileoIOV(_ float64, _ float64) (*GNSSAttitudeLaw, error) {
	return nil, ErrUnavailable
}

func GNSSAttitudeLawGLONASSM(_ float64, _ float64) (*GNSSAttitudeLaw, error) {
	return nil, ErrUnavailable
}

func GNSSAttitudeLawGPSIIA(_ float64, _ float64, _ float64) (*GNSSAttitudeLaw, error) {
	return nil, ErrUnavailable
}

func GNSSAttitudeLawGPSIIF(_ float64, _ float64, _ float64) (*GNSSAttitudeLaw, error) {
	return nil, ErrUnavailable
}

func GNSSAttitudeLawGPSIIR(_ float64) (*GNSSAttitudeLaw, error) { return nil, ErrUnavailable }

func GNSSAttitudeLawOrbitNormal() (*GNSSAttitudeLaw, error) { return nil, ErrUnavailable }

func GNSSBlockBeiDou3MEO(_ *BeiDouManufacturer) (GNSSBlock, error) { return 0, ErrUnavailable }

func GNSSBlockCompare(_ GNSSBlock, _ GNSSBlock) (int32, error) { return 0, ErrUnavailable }

func GNSSBlockFromConstellationBlockType(_ uint32, _ []byte, _ *BeiDouOrbitClass) (GNSSBlock, error) {
	return 0, ErrUnavailable
}

func GNSSBlockFromIGSName(_ []byte) (GNSSBlock, error) { return 0, ErrUnavailable }

func GNSSBlockIGSName(_ GNSSBlock) (string, error) { return "", ErrUnavailable }

func GNSSBlockSystem(_ GNSSBlock) (uint32, error) { return 0, ErrUnavailable }

func GPSIIADefaultYawBiasRad() float64 { return 0 }

func GPSIIADefaultYawRateRadS(_ uint8) (float64, error) { return 0, ErrUnavailable }

func GPSIIFDefaultYawBiasRad() float64 { return 0 }

func GPSIIFDefaultYawRateRadS() float64 { return 0 }

func GPSIIIBetaXRad() float64 { return 0 }

func GPSIIIBetaYRad() float64 { return 0 }

func GPSIIRDefaultYawRateRadS() float64 { return 0 }

func GPSShadowLimitRad() float64 { return 0 }

func ComputeNominalYaw(_ AttitudeSample) (NominalYaw, error) { return NominalYaw{}, ErrUnavailable }
