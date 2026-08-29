package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// GNSSSystem identifies a constellation in the C frequency tables.
type GNSSSystem uint32

const (
	GNSSSystemGPS     GNSSSystem = 0
	GNSSSystemGLONASS GNSSSystem = 1
	GNSSSystemGalileo GNSSSystem = 2
	GNSSSystemBeiDou  GNSSSystem = 3
	GNSSSystemQZSS    GNSSSystem = 4
	GNSSSystemNavIC   GNSSSystem = 5
	GNSSSystemSBAS    GNSSSystem = 6
)

// CarrierBand identifies a constellation carrier band in the C frequency
// tables.
type CarrierBand uint32

const (
	CarrierBandL1  CarrierBand = 0
	CarrierBandL2  CarrierBand = 1
	CarrierBandL5  CarrierBand = 2
	CarrierBandE1  CarrierBand = 3
	CarrierBandE5A CarrierBand = 4
	CarrierBandE5B CarrierBand = 5
	CarrierBandE5  CarrierBand = 6
	CarrierBandE6  CarrierBand = 7
	CarrierBandB1C CarrierBand = 8
	CarrierBandB1I CarrierBand = 9
	CarrierBandB2A CarrierBand = 10
	CarrierBandB2B CarrierBand = 11
	CarrierBandB2  CarrierBand = 12
	CarrierBandB3I CarrierBand = 13
	CarrierBandG1  CarrierBand = 14
	CarrierBandG2  CarrierBand = 15
)

// CarrierPair is a standard two-band ionosphere-free pair.
type CarrierPair struct {
	Band1 CarrierBand
	Band2 CarrierBand
}

// Frequency returns the C table's carrier frequency in hertz.
func Frequency(system GNSSSystem, band CarrierBand) (float64, error) {
	value, err := native.Frequency(uint32(system), uint32(band))
	return value, publicError(err)
}

// Wavelength returns the C table's carrier wavelength in metres.
func Wavelength(system GNSSSystem, band CarrierBand) (float64, error) {
	value, err := native.Wavelength(uint32(system), uint32(band))
	return value, publicError(err)
}

// GLONASSG1Frequency returns the FDMA G1 frequency in hertz for a channel.
func GLONASSG1Frequency(channel int8) (float64, error) {
	value, err := native.GLONASSG1Frequency(channel)
	return value, publicError(err)
}

// DefaultSPPFrequency returns the default single-point-positioning frequency
// in hertz for a GNSS system.
func DefaultSPPFrequency(system GNSSSystem) (float64, error) {
	value, err := native.DefaultSPPFrequency(uint32(system))
	return value, publicError(err)
}

// DefaultIonosphereFreePair returns the C policy's default pair and whether a
// constellation-wide default is present.
func DefaultIonosphereFreePair(system GNSSSystem) (CarrierPair, bool, error) {
	value, present, err := native.DefaultIonosphereFreePair(uint32(system))
	return CarrierPair{Band1: CarrierBand(value.Band1), Band2: CarrierBand(value.Band2)}, present, publicError(err)
}

// CarrierBandLabel returns the C label such as "l1".
func CarrierBandLabel(band CarrierBand) (string, error) {
	value, err := native.CarrierBandLabel(uint32(band))
	return string(value), publicError(err)
}

// GNSSSystemLabel returns the C constellation label such as "GPS".
func GNSSSystemLabel(system GNSSSystem) (string, error) {
	value, err := native.GNSSSystemLabel(uint32(system))
	return string(value), publicError(err)
}

// PhaseMeters converts carrier phase cycles to metres using the supplied
// frequency in hertz.
func PhaseMeters(phaseCycles, frequencyHz float64) (float64, error) {
	value, err := native.PhaseMeters(phaseCycles, frequencyHz)
	return value, publicError(err)
}

// CodeMinusCarrier returns code minus carrier in metres.
func CodeMinusCarrier(codeMeters, phaseCycles, frequencyHz float64) (float64, error) {
	value, err := native.CodeMinusCarrier(codeMeters, phaseCycles, frequencyHz)
	return value, publicError(err)
}

// GeometryFree returns the L1-L2 carrier-phase combination in metres.
func GeometryFree(l1Meters, l2Meters float64) (float64, error) {
	value, err := native.GeometryFree(l1Meters, l2Meters)
	return value, publicError(err)
}

// MelbourneWubbena returns the Melbourne-Wubbena combination in metres.
func MelbourneWubbena(phi1Cycles, phi2Cycles, p1Meters, p2Meters, f1Hz, f2Hz float64) (float64, error) {
	value, err := native.MelbourneWubbena(phi1Cycles, phi2Cycles, p1Meters, p2Meters, f1Hz, f2Hz)
	return value, publicError(err)
}

// NarrowLaneCode returns the narrow-lane code combination in metres.
func NarrowLaneCode(p1Meters, p2Meters, f1Hz, f2Hz float64) (float64, error) {
	value, err := native.NarrowLaneCode(p1Meters, p2Meters, f1Hz, f2Hz)
	return value, publicError(err)
}

// WideLaneCycles returns the wide-lane ambiguity estimate in cycles.
func WideLaneCycles(phi1Cycles, phi2Cycles, p1Meters, p2Meters, f1Hz, f2Hz float64) (float64, error) {
	value, err := native.WideLaneCycles(phi1Cycles, phi2Cycles, p1Meters, p2Meters, f1Hz, f2Hz)
	return value, publicError(err)
}

// WideLaneWavelength returns the two-frequency wide-lane wavelength in
// metres.
func WideLaneWavelength(f1Hz, f2Hz float64) (float64, error) {
	value, err := native.WideLaneWavelength(f1Hz, f2Hz)
	return value, publicError(err)
}

// IonosphericGamma returns (f1/f2)^2.
func IonosphericGamma(f1Hz, f2Hz float64) (float64, error) {
	value, err := native.IonosphericGamma(f1Hz, f2Hz)
	return value, publicError(err)
}

// IonosphereFreeNoiseAmplification returns the C ionosphere-free noise
// amplification factor.
func IonosphereFreeNoiseAmplification(f1Hz, f2Hz float64) (float64, error) {
	value, err := native.CombinationNoiseAmplification(f1Hz, f2Hz)
	return value, publicError(err)
}

// IonosphereFreePseudorange returns the ionosphere-free code combination in
// metres.
func IonosphereFreePseudorange(obs1Meters, obs2Meters, f1Hz, f2Hz float64) (float64, error) {
	value, err := native.IonosphereFreeCode(obs1Meters, obs2Meters, f1Hz, f2Hz)
	return value, publicError(err)
}

// IonosphereFreePhaseCycles returns an ionosphere-free carrier combination in
// metres from phase cycles.
func IonosphereFreePhaseCycles(phi1Cycles, phi2Cycles, f1Hz, f2Hz float64) (float64, error) {
	value, err := native.IonosphereFreePhaseCycles(phi1Cycles, phi2Cycles, f1Hz, f2Hz)
	return value, publicError(err)
}

// IonosphereFreePhaseMeters returns an ionosphere-free carrier combination in
// metres from phase ranges.
func IonosphereFreePhaseMeters(phase1Meters, phase2Meters, f1Hz, f2Hz float64) (float64, error) {
	value, err := native.IonosphereFreePhaseMeters(phase1Meters, phase2Meters, f1Hz, f2Hz)
	return value, publicError(err)
}
