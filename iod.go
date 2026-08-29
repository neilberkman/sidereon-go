package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// GibbsResult contains the middle-observation velocity in km/s and angular
// diagnostics in radians.
type GibbsResult struct {
	VelocityKmPerS [3]float64
	Theta12Rad     float64
	Theta23Rad     float64
	CoplanarRad    float64
}

type GaussAnglesResult struct {
	PositionKm     [3]float64
	VelocityKmPerS [3]float64
}

func publicGibbs(value native.GibbsResult) GibbsResult {
	return GibbsResult{VelocityKmPerS: value.VelocityKmPerS, Theta12Rad: value.Theta12Rad, Theta23Rad: value.Theta23Rad, CoplanarRad: value.CoplanarRad}
}

func InitialOrbitGibbs(r1Km, r2Km, r3Km [3]float64) (GibbsResult, error) {
	value, err := native.IODGibbs(r1Km, r2Km, r3Km)
	return publicGibbs(value), publicError(err)
}

func InitialOrbitHerrickGibbs(r1Km, r2Km, r3Km [3]float64, jd1, jd2, jd3 float64) (GibbsResult, error) {
	value, err := native.IODHGibbs(r1Km, r2Km, r3Km, jd1, jd2, jd3)
	return publicGibbs(value), publicError(err)
}

// InitialOrbitGaussAngles takes three right-ascension/declination pairs in
// radians, split Julian dates, and ECI site positions in km.
func InitialOrbitGaussAngles(declinationRad, rightAscensionRad, jd, jdFraction [3]float64, siteECIKm [3][3]float64) (GaussAnglesResult, error) {
	value, err := native.IODGaussAngles(declinationRad, rightAscensionRad, jd, jdFraction, siteECIKm)
	return GaussAnglesResult{PositionKm: value.PositionKm, VelocityKmPerS: value.VelocityKmPerS}, publicError(err)
}
