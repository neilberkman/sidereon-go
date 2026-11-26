package sidereon

import (
	"sidereon.dev/go/v4/internal/native"
)

// BeiDou2OnToYsBetaRad returns the beidou2 on to ys beta rad.
func BeiDou2OnToYsBetaRad() float64 {
	return native.BeiDou2OnToYsBetaRad()
}

// BeiDou2YsToOnBetaRad returns the beidou2 ys to on beta rad.
func BeiDou2YsToOnBetaRad() float64 {
	return native.BeiDou2YsToOnBetaRad()
}

// BeiDou3CASTIGSOTurn returns the beidou3 cast igso turn.
func BeiDou3CASTIGSOTurn() CosineTurnParameters {
	value := native.BeiDou3CASTIGSOTurn()
	return publicCosineTurnParameters(value)
}

// BeiDou3CASTMEOTurn returns the beidou3 cast meo turn.
func BeiDou3CASTMEOTurn() CosineTurnParameters {
	value := native.BeiDou3CASTMEOTurn()
	return publicCosineTurnParameters(value)
}

// BeiDou3SECMBeta0Rad returns the beidou3 secm beta0 rad.
func BeiDou3SECMBeta0Rad() float64 {
	return native.BeiDou3SECMBeta0Rad()
}

// Compare returns -1, 0, or 1 according to the ordering of the two values.
func (left BeiDouManufacturer) Compare(right BeiDouManufacturer) (int32, error) {
	value, err := native.BeiDouManufacturerCompare(native.BeiDouManufacturer(left), native.BeiDouManufacturer(right))
	return value, publicError(err)
}

// Compare returns -1, 0, or 1 according to the ordering of the two values.
func (left BeiDouOrbitClass) Compare(right BeiDouOrbitClass) (int32, error) {
	value, err := native.BeiDouOrbitClassCompare(native.BeiDouOrbitClass(left), native.BeiDouOrbitClass(right))
	return value, publicError(err)
}

// ComputeDYBAxes evaluates the Sun-fixed orthonormal axes from geocentric positions in metres.
func ComputeDYBAxes(positionM [3]float64, sunPositionM [3]float64) (DYBAxes, error) {
	value, err := native.ComputeDYBAxes(positionM, sunPositionM)
	return publicDYBAxes(value), publicError(err)
}

// GalileoFOCTurn returns the galileo foc turn.
func GalileoFOCTurn() CosineTurnParameters {
	value := native.GalileoFOCTurn()
	return publicCosineTurnParameters(value)
}

// GalileoIOVBetaXRad returns the galileo iov beta x rad.
func GalileoIOVBetaXRad() float64 {
	return native.GalileoIOVBetaXRad()
}

// GalileoIOVBetaYRad returns the galileo iov beta y rad.
func GalileoIOVBetaYRad() float64 {
	return native.GalileoIOVBetaYRad()
}

// GLONASSMDefaultYawRateRadS returns the glonass m default yaw rate rad s.
func GLONASSMDefaultYawRateRadS() float64 {
	return native.GLONASSMDefaultYawRateRadS()
}

// GLONASSMShadowLimitRad returns the glonass m shadow limit rad.
func GLONASSMShadowLimitRad() float64 {
	return native.GLONASSMShadowLimitRad()
}

// GNSSAttitudeLawBeiDou3MEOUnknownManufacturer returns the beidou3 meo unknown manufacturer.
func GNSSAttitudeLawBeiDou3MEOUnknownManufacturer(cast CosineTurnParameters, secmBeta0Rad float64) (*GNSSAttitudeLaw, error) {
	value, err := native.GNSSAttitudeLawBeiDou3MEOUnknownManufacturer(nativeCosineTurnParameters(cast), secmBeta0Rad)
	if err != nil {
		return nil, publicError(err)
	}
	return publicGNSSAttitudeLaw(value), nil
}

// GNSSAttitudeLawBeiDouOrbitNormalSwitch returns the beidou orbit normal switch.
func GNSSAttitudeLawBeiDouOrbitNormalSwitch(ysToOnBetaRad float64, onToYsBetaRad float64) (*GNSSAttitudeLaw, error) {
	value, err := native.GNSSAttitudeLawBeiDouOrbitNormalSwitch(ysToOnBetaRad, onToYsBetaRad)
	if err != nil {
		return nil, publicError(err)
	}
	return publicGNSSAttitudeLaw(value), nil
}

// GNSSAttitudeLawCosineTurn returns the cosine turn.
func GNSSAttitudeLawCosineTurn(parameters CosineTurnParameters) (*GNSSAttitudeLaw, error) {
	value, err := native.GNSSAttitudeLawCosineTurn(nativeCosineTurnParameters(parameters))
	if err != nil {
		return nil, publicError(err)
	}
	return publicGNSSAttitudeLaw(value), nil
}

// Equal compares the underlying values.
func (left *GNSSAttitudeLaw) Equal(right *GNSSAttitudeLaw) (bool, error) {
	value, err := native.GNSSAttitudeLawEqual(nativeGNSSAttitudeLawHandle(left), nativeGNSSAttitudeLawHandle(right))
	return value, publicError(err)
}

// GNSSAttitudeLawFixedBetaYawSteering returns the fixed beta yaw steering.
func GNSSAttitudeLawFixedBetaYawSteering(beta0Rad float64) (*GNSSAttitudeLaw, error) {
	value, err := native.GNSSAttitudeLawFixedBetaYawSteering(beta0Rad)
	if err != nil {
		return nil, publicError(err)
	}
	return publicGNSSAttitudeLaw(value), nil
}

// GNSSAttitudeLawForBlock selects the default law for a block. A PRN of zero omits the PRN.
func GNSSAttitudeLawForBlock(block GNSSBlock, prn uint8) (*GNSSAttitudeLaw, error) {
	value, err := native.GNSSAttitudeLawForBlock(native.GNSSBlock(block), prn)
	if err != nil {
		return nil, publicError(err)
	}
	return publicGNSSAttitudeLaw(value), nil
}

// GNSSAttitudeLawGalileoIOV returns the galileo iov.
func GNSSAttitudeLawGalileoIOV(betaXRad float64, betaYRad float64) (*GNSSAttitudeLaw, error) {
	value, err := native.GNSSAttitudeLawGalileoIOV(betaXRad, betaYRad)
	if err != nil {
		return nil, publicError(err)
	}
	return publicGNSSAttitudeLaw(value), nil
}

// GNSSAttitudeLawGLONASSM returns the glonass m.
func GNSSAttitudeLawGLONASSM(yawRateRadS float64, shadowLimitRad float64) (*GNSSAttitudeLaw, error) {
	value, err := native.GNSSAttitudeLawGLONASSM(yawRateRadS, shadowLimitRad)
	if err != nil {
		return nil, publicError(err)
	}
	return publicGNSSAttitudeLaw(value), nil
}

// GNSSAttitudeLawGPSIIA returns the gps iia.
func GNSSAttitudeLawGPSIIA(yawRateRadS float64, yawBiasRad float64, shadowLimitRad float64) (*GNSSAttitudeLaw, error) {
	value, err := native.GNSSAttitudeLawGPSIIA(yawRateRadS, yawBiasRad, shadowLimitRad)
	if err != nil {
		return nil, publicError(err)
	}
	return publicGNSSAttitudeLaw(value), nil
}

// GNSSAttitudeLawGPSIIF returns the gps iif.
func GNSSAttitudeLawGPSIIF(yawRateRadS float64, yawBiasRad float64, shadowLimitRad float64) (*GNSSAttitudeLaw, error) {
	value, err := native.GNSSAttitudeLawGPSIIF(yawRateRadS, yawBiasRad, shadowLimitRad)
	if err != nil {
		return nil, publicError(err)
	}
	return publicGNSSAttitudeLaw(value), nil
}

// GNSSAttitudeLawGPSIIR returns the gps iir.
func GNSSAttitudeLawGPSIIR(yawRateRadS float64) (*GNSSAttitudeLaw, error) {
	value, err := native.GNSSAttitudeLawGPSIIR(yawRateRadS)
	if err != nil {
		return nil, publicError(err)
	}
	return publicGNSSAttitudeLaw(value), nil
}

// GNSSAttitudeLawOrbitNormal returns the orbit normal.
func GNSSAttitudeLawOrbitNormal() (*GNSSAttitudeLaw, error) {
	value, err := native.GNSSAttitudeLawOrbitNormal()
	if err != nil {
		return nil, publicError(err)
	}
	return publicGNSSAttitudeLaw(value), nil
}

// GNSSBlockBeiDou3MEO returns the beidou3 meo.
func GNSSBlockBeiDou3MEO(manufacturer *BeiDouManufacturer) (GNSSBlock, error) {
	value, err := native.GNSSBlockBeiDou3MEO((*native.BeiDouManufacturer)(manufacturer))
	return GNSSBlock(value), publicError(err)
}

// Compare returns -1, 0, or 1 according to the ordering of the two values.
func (left GNSSBlock) Compare(right GNSSBlock) (int32, error) {
	value, err := native.GNSSBlockCompare(native.GNSSBlock(left), native.GNSSBlock(right))
	return value, publicError(err)
}

// GNSSBlockFromConstellationBlockType identifies a satellite block from constellation metadata.
func GNSSBlockFromConstellationBlockType(system GNSSSystem, data string, beidouOrbit *BeiDouOrbitClass) (GNSSBlock, error) {
	value, err := native.GNSSBlockFromConstellationBlockType(uint32(system), []byte(data), (*native.BeiDouOrbitClass)(beidouOrbit))
	return GNSSBlock(value), publicError(err)
}

// GNSSBlockFromIGSName identifies a satellite block by its IGS name.
func GNSSBlockFromIGSName(data string) (GNSSBlock, error) {
	value, err := native.GNSSBlockFromIGSName([]byte(data))
	return GNSSBlock(value), publicError(err)
}

// IGSName returns the IGS name.
func (block GNSSBlock) IGSName() (string, error) {
	value, err := native.GNSSBlockIGSName(native.GNSSBlock(block))
	return value, publicError(err)
}

// System returns the system.
func (block GNSSBlock) System() (GNSSSystem, error) {
	value, err := native.GNSSBlockSystem(native.GNSSBlock(block))
	return GNSSSystem(value), publicError(err)
}

// GPSIIADefaultYawBiasRad returns the gps iia default yaw bias rad.
func GPSIIADefaultYawBiasRad() float64 {
	return native.GPSIIADefaultYawBiasRad()
}

// GPSIIADefaultYawRateRadS returns the gps iia default yaw rate rad s.
func GPSIIADefaultYawRateRadS(prn uint8) (float64, error) {
	value, err := native.GPSIIADefaultYawRateRadS(prn)
	return value, publicError(err)
}

// GPSIIFDefaultYawBiasRad returns the gps iif default yaw bias rad.
func GPSIIFDefaultYawBiasRad() float64 {
	return native.GPSIIFDefaultYawBiasRad()
}

// GPSIIFDefaultYawRateRadS returns the gps iif default yaw rate rad s.
func GPSIIFDefaultYawRateRadS() float64 {
	return native.GPSIIFDefaultYawRateRadS()
}

// GPSIIIBetaXRad returns the gps iii beta x rad.
func GPSIIIBetaXRad() float64 {
	return native.GPSIIIBetaXRad()
}

// GPSIIIBetaYRad returns the gps iii beta y rad.
func GPSIIIBetaYRad() float64 {
	return native.GPSIIIBetaYRad()
}

// GPSIIRDefaultYawRateRadS returns the gps iir default yaw rate rad s.
func GPSIIRDefaultYawRateRadS() float64 {
	return native.GPSIIRDefaultYawRateRadS()
}

// GPSShadowLimitRad returns the gps shadow limit rad.
func GPSShadowLimitRad() float64 {
	return native.GPSShadowLimitRad()
}

// ComputeNominalYaw evaluates nominal yaw, beta and body axes for the sample.
func ComputeNominalYaw(sample AttitudeSample) (NominalYaw, error) {
	value, err := native.ComputeNominalYaw(nativeAttitudeSample(sample))
	return publicNominalYaw(value), publicError(err)
}

// String returns the IGS block name.
func (block GNSSBlock) String() string {
	value, _ := block.IGSName()
	return value
}
