package sidereon

import (
	"errors"

	"github.com/neilberkman/sidereon-go/internal/native"
)

// EarthMuKm3PerS2 is the standard gravitational parameter used by the public
// orbital examples. Orbit conversion functions accept an explicit mu so a
// caller can select another central body or convention.
const EarthMuKm3PerS2 = 398600.4418

// CartesianState is an ECI state. Position and velocity are km and km/s.
// EpochTDBSeconds is the absolute TDB epoch in seconds, exactly as defined by
// SidereonCartesianState; it is not UTC and is not route-dependent. Epoch-free
// conversion routes return zero for this field because they do not receive an
// epoch.
type CartesianState struct {
	// EpochTDBSeconds is the absolute TDB epoch in seconds.
	EpochTDBSeconds float64
	// PositionKm is the ECI position in kilometres.
	PositionKm [3]float64
	// VelocityKmPerS is the ECI velocity in kilometres per second.
	VelocityKmPerS [3]float64
}

// OrbitType is the C geometric classification for classical elements.
type OrbitType uint32

const (
	// OrbitEllipticalInclined identifies the orbit elliptical inclined case.
	OrbitEllipticalInclined OrbitType = OrbitType(native.OrbitTypeEllipticalInclinedValue)
	// OrbitEllipticalEquatorial identifies the orbit elliptical equatorial case.
	OrbitEllipticalEquatorial OrbitType = OrbitType(native.OrbitTypeEllipticalEquatorialValue)
	// OrbitCircularInclined identifies the orbit circular inclined case.
	OrbitCircularInclined OrbitType = OrbitType(native.OrbitTypeCircularInclinedValue)
	// OrbitCircularEquatorial identifies the orbit circular equatorial case.
	OrbitCircularEquatorial OrbitType = OrbitType(native.OrbitTypeCircularEquatorialValue)
)

// ClassicalElements are Keplerian elements. p and a are km, eccentricity is
// dimensionless, and all angles are radians. For circular/equatorial cases,
// undefined auxiliary angles are NaN; use OrbitType to select the defined
// replacement angle.
type ClassicalElements struct {
	// P is the semilatus rectum in kilometres.
	P float64
	// A is the semimajor axis in kilometres.
	A float64
	// Ecc is the dimensionless eccentricity.
	Ecc float64
	// Incl is the inclination in radians.
	Incl float64
	// RAAN is the right ascension of the ascending node in radians.
	RAAN float64
	// ArgP is the argument of periapsis in radians.
	ArgP float64
	// Nu is the true anomaly in radians.
	Nu float64
	// ArgLat is the argument of latitude in radians.
	ArgLat float64
	// TrueLon is the true longitude in radians.
	TrueLon float64
	// LonPer is the longitude of periapsis in radians.
	LonPer float64
	// OrbitType is the geometric orbit classification.
	OrbitType OrbitType
}

// RetrogradeFactor selects the equinoctial singularity convention.
type RetrogradeFactor uint32

const (
	// RetrogradeFactorPrograde identifies the retrograde factor prograde case.
	RetrogradeFactorPrograde RetrogradeFactor = RetrogradeFactor(native.RetrogradeFactorProgradeValue)
	// RetrogradeFactorRetrograde identifies the retrograde factor retrograde case.
	RetrogradeFactorRetrograde RetrogradeFactor = RetrogradeFactor(native.RetrogradeFactorRetrogradeValue)
)

// EquinoctialElements are nonsingular orbital elements: a is km, h/k/p/q are
// dimensionless, lambda is radians, and Retrograde selects the pole chart.
type EquinoctialElements struct {
	// A is the semimajor axis in kilometres.
	A float64
	// H is a dimensionless equinoctial component.
	H float64
	// K is a dimensionless equinoctial component.
	K float64
	// P is a dimensionless equinoctial component.
	P float64
	// Q is a dimensionless equinoctial component.
	Q float64
	// Lambda is the mean longitude in radians.
	Lambda float64
	// Retrograde is the selected equinoctial pole chart.
	Retrograde RetrogradeFactor
}

// ModifiedEquinoctialElements are nonsingular orbital elements: p is km,
// f/g/h/k are dimensionless, l is radians, and Retrograde selects the pole
// chart.
type ModifiedEquinoctialElements struct {
	// P is the semilatus rectum in kilometres.
	P float64
	// F is a dimensionless modified-equinoctial component.
	F float64
	// G is a dimensionless modified-equinoctial component.
	G float64
	// H is a dimensionless modified-equinoctial component.
	H float64
	// K is a dimensionless modified-equinoctial component.
	K float64
	// L is the true longitude in radians.
	L float64
	// Retrograde is the selected modified-equinoctial pole chart.
	Retrograde RetrogradeFactor
}

// KeplerSolution contains the solved eccentric anomaly in radians and the
// C solver's accepted iteration count.
type KeplerSolution struct {
	// AnomalyRad is the anomaly rad in radians.
	AnomalyRad float64
	// Iterations is the native solver iteration count.
	Iterations int
}

func nativeClassical(value ClassicalElements) native.ClassicalElements {
	return native.ClassicalElements{P: value.P, A: value.A, Ecc: value.Ecc, Incl: value.Incl, RAAN: value.RAAN, ArgP: value.ArgP, Nu: value.Nu, ArgLat: value.ArgLat, TrueLon: value.TrueLon, LonPer: value.LonPer, OrbitType: uint32(value.OrbitType)}
}

func publicClassical(value native.ClassicalElements) ClassicalElements {
	return ClassicalElements{P: value.P, A: value.A, Ecc: value.Ecc, Incl: value.Incl, RAAN: value.RAAN, ArgP: value.ArgP, Nu: value.Nu, ArgLat: value.ArgLat, TrueLon: value.TrueLon, LonPer: value.LonPer, OrbitType: OrbitType(value.OrbitType)}
}

func nativeEquinoctial(value EquinoctialElements) native.EquinoctialElements {
	return native.EquinoctialElements{A: value.A, H: value.H, K: value.K, P: value.P, Q: value.Q, Lambda: value.Lambda, Retrograde: uint32(value.Retrograde)}
}

func publicEquinoctial(value native.EquinoctialElements) EquinoctialElements {
	return EquinoctialElements{A: value.A, H: value.H, K: value.K, P: value.P, Q: value.Q, Lambda: value.Lambda, Retrograde: RetrogradeFactor(value.Retrograde)}
}

func nativeModifiedEquinoctial(value ModifiedEquinoctialElements) native.ModifiedEquinoctialElements {
	return native.ModifiedEquinoctialElements{P: value.P, F: value.F, G: value.G, H: value.H, K: value.K, L: value.L, Retrograde: uint32(value.Retrograde)}
}

func publicModifiedEquinoctial(value native.ModifiedEquinoctialElements) ModifiedEquinoctialElements {
	return ModifiedEquinoctialElements{P: value.P, F: value.F, G: value.G, H: value.H, K: value.K, L: value.L, Retrograde: RetrogradeFactor(value.Retrograde)}
}

func nativeCartesian(value CartesianState) native.CartesianState {
	return native.CartesianState{EpochTDBSeconds: value.EpochTDBSeconds, PositionKm: value.PositionKm, VelocityKmPerS: value.VelocityKmPerS}
}

func publicCartesian(value native.CartesianState) CartesianState {
	return CartesianState{EpochTDBSeconds: value.EpochTDBSeconds, PositionKm: value.PositionKm, VelocityKmPerS: value.VelocityKmPerS}
}

// CartesianToClassical converts an ECI state to classical elements. mu is in
// km^3/s^2.
func CartesianToClassical(state CartesianState, mu float64) (ClassicalElements, error) {
	value, err := native.CartesianToClassical(nativeCartesian(state), mu)
	return publicClassical(value), publicError(err)
}

// ClassicalToCartesian converts classical elements to an ECI state. mu is in
// km^3/s^2; the returned EpochTDBSeconds is zero because the C route is
// epoch-free.
func ClassicalToCartesian(elements ClassicalElements, mu float64) (CartesianState, error) {
	value, err := native.ClassicalToCartesian(nativeClassical(elements), mu)
	return publicCartesian(value), publicError(err)
}

// ClassicalToEquinoctial converts classical elements using the selected pole
// chart.
func ClassicalToEquinoctial(elements ClassicalElements, factor RetrogradeFactor) (EquinoctialElements, error) {
	value, err := native.ClassicalToEquinoctial(nativeClassical(elements), uint32(factor))
	return publicEquinoctial(value), publicError(err)
}

// EquinoctialToClassical converts equinoctial elements to classical elements.
func EquinoctialToClassical(elements EquinoctialElements) (ClassicalElements, error) {
	value, err := native.EquinoctialToClassical(nativeEquinoctial(elements))
	return publicClassical(value), publicError(err)
}

// ClassicalToModifiedEquinoctial converts classical to modified equinoctial
// elements using the selected pole chart.
func ClassicalToModifiedEquinoctial(elements ClassicalElements, factor RetrogradeFactor) (ModifiedEquinoctialElements, error) {
	value, err := native.ClassicalToModifiedEquinoctial(nativeClassical(elements), uint32(factor))
	return publicModifiedEquinoctial(value), publicError(err)
}

// ModifiedEquinoctialToClassical converts modified equinoctial elements.
func ModifiedEquinoctialToClassical(elements ModifiedEquinoctialElements) (ClassicalElements, error) {
	value, err := native.ModifiedEquinoctialToClassical(nativeModifiedEquinoctial(elements))
	return publicClassical(value), publicError(err)
}

// CartesianToEquinoctial converts an ECI state to equinoctial elements.
func CartesianToEquinoctial(state CartesianState, mu float64, factor RetrogradeFactor) (EquinoctialElements, error) {
	value, err := native.CartesianToEquinoctial(nativeCartesian(state), mu, uint32(factor))
	return publicEquinoctial(value), publicError(err)
}

// EquinoctialToCartesian converts equinoctial elements to an ECI state. The
// returned EpochTDBSeconds is zero because the C route is epoch-free.
func EquinoctialToCartesian(elements EquinoctialElements, mu float64) (CartesianState, error) {
	value, err := native.EquinoctialToCartesian(nativeEquinoctial(elements), mu)
	return publicCartesian(value), publicError(err)
}

// CartesianToModifiedEquinoctial converts an ECI state to modified
// equinoctial elements.
func CartesianToModifiedEquinoctial(state CartesianState, mu float64, factor RetrogradeFactor) (ModifiedEquinoctialElements, error) {
	value, err := native.CartesianToModifiedEquinoctial(nativeCartesian(state), mu, uint32(factor))
	return publicModifiedEquinoctial(value), publicError(err)
}

// ModifiedEquinoctialToCartesian converts modified equinoctial elements to an
// ECI state. The returned EpochTDBSeconds is zero because the C route is
// epoch-free.
func ModifiedEquinoctialToCartesian(elements ModifiedEquinoctialElements, mu float64) (CartesianState, error) {
	value, err := native.ModifiedEquinoctialToCartesian(nativeModifiedEquinoctial(elements), mu)
	return publicCartesian(value), publicError(err)
}

// MeanToEccentricAnomaly converts mean anomaly to eccentric anomaly. Angles
// are radians and eccentricity is dimensionless.
func MeanToEccentricAnomaly(mean, eccentricity float64) (float64, error) {
	value, err := native.MeanToEccentricAnomaly(mean, eccentricity)
	return value, publicError(err)
}

// EccentricToMeanAnomaly converts eccentric anomaly to mean anomaly. Angles
// are radians and eccentricity is dimensionless.
func EccentricToMeanAnomaly(eccentric, eccentricity float64) (float64, error) {
	value, err := native.EccentricToMeanAnomaly(eccentric, eccentricity)
	return value, publicError(err)
}

// EccentricToTrueAnomaly converts eccentric anomaly to true anomaly. Angles
// are radians and eccentricity is dimensionless.
func EccentricToTrueAnomaly(eccentric, eccentricity float64) (float64, error) {
	value, err := native.EccentricToTrueAnomaly(eccentric, eccentricity)
	return value, publicError(err)
}

// TrueToEccentricAnomaly converts true anomaly to eccentric anomaly. Angles
// are radians and eccentricity is dimensionless.
func TrueToEccentricAnomaly(trueAnomaly, eccentricity float64) (float64, error) {
	value, err := native.TrueToEccentricAnomaly(trueAnomaly, eccentricity)
	return value, publicError(err)
}

// MeanToTrueAnomaly converts mean anomaly to true anomaly. Angles are radians
// and eccentricity is dimensionless.
func MeanToTrueAnomaly(mean, eccentricity float64) (float64, error) {
	value, err := native.MeanToTrueAnomaly(mean, eccentricity)
	return value, publicError(err)
}

// TrueToMeanAnomaly converts true anomaly to mean anomaly. Angles are radians
// and eccentricity is dimensionless.
func TrueToMeanAnomaly(trueAnomaly, eccentricity float64) (float64, error) {
	value, err := native.TrueToMeanAnomaly(trueAnomaly, eccentricity)
	return value, publicError(err)
}

// SolveKepler solves Kepler's equation and returns the C solver diagnostics.
func SolveKepler(mean, eccentricity float64) (KeplerSolution, error) {
	value, err := native.SolveKepler(mean, eccentricity)
	return KeplerSolution{AnomalyRad: value.AnomalyRad, Iterations: value.Iterations}, publicError(err)
}

// PropagateKepler advances the mean anomaly in classical elements by dt
// seconds. mu is in km^3/s^2.
func PropagateKepler(elements ClassicalElements, mu, dt float64) (ClassicalElements, error) {
	value, err := native.PropagateKepler(nativeClassical(elements), mu, dt)
	return publicClassical(value), publicError(err)
}

// LambertDirection selects the short-way or long-way transfer arc.
type LambertDirection uint32

const (
	// LambertShortWay identifies the lambert short way case.
	LambertShortWay LambertDirection = 0
	// LambertLongWay identifies the lambert long way case.
	LambertLongWay LambertDirection = 1
)

// LambertEnergy selects the low-energy or high-energy Lambert branch.
type LambertEnergy uint32

const (
	// LambertLowEnergy identifies the lambert low energy case.
	LambertLowEnergy LambertEnergy = 0
	// LambertHighEnergy identifies the lambert high energy case.
	LambertHighEnergy LambertEnergy = 1
)

// LambertBattin solves the Battin Lambert boundary-value problem. Positions
// are km, the supplied initial velocity and returned transfer velocities are
// km/s, and dt is seconds. revolutions must fit the native signed 32-bit
// parameter.
func LambertBattin(r1, r2, v1 [3]float64, direction LambertDirection, energy LambertEnergy, revolutions int, dt float64) ([3]float64, [3]float64, error) {
	if direction != LambertShortWay && direction != LambertLongWay {
		return [3]float64{}, [3]float64{}, errors.New("sidereon: Lambert direction must be 0 or 1")
	}
	if energy != LambertLowEnergy && energy != LambertHighEnergy {
		return [3]float64{}, [3]float64{}, errors.New("sidereon: Lambert energy must be 0 or 1")
	}
	if int64(revolutions) < -int64(1<<31) || int64(revolutions) > int64(1<<31-1) {
		return [3]float64{}, [3]float64{}, errors.New("sidereon: Lambert revolutions exceed native int32 range")
	}
	departure, arrival, err := native.LambertBattin(r1, r2, v1, uint32(direction), uint32(energy), int32(revolutions), dt)
	return departure, arrival, publicError(err)
}

// BetaAngleDeg returns the Sun-to-orbital-plane beta angle in degrees. The
// orbit normal and Sun vectors are dimensionless directions.
func BetaAngleDeg(orbitNormal, sun [3]float64) (float64, error) {
	value, err := native.BetaAngleDeg(orbitNormal, sun)
	return value, publicError(err)
}

// BetaAngleFromStateDeg returns the beta angle in degrees from an ECI state
// and Sun position. Position is km and velocity is km/s.
func BetaAngleFromStateDeg(positionKm, velocityKmPerS, sunKm [3]float64) (float64, error) {
	value, err := native.BetaAngleFromStateDeg(positionKm, velocityKmPerS, sunKm)
	return value, publicError(err)
}

// RTNToECICovariance rotates a row-major RTN covariance in km^2 into ECI
// coordinates using the supplied position in km and velocity in km/s.
func RTNToECICovariance(covarianceRTN [3][3]float64, positionKm, velocityKmPerS [3]float64) ([3][3]float64, error) {
	value, err := native.RTNToECICovariance(covarianceRTN, positionKm, velocityKmPerS)
	return value, publicError(err)
}

// CWStateTransitionMatrix returns the 6x6 row-major Clohessy-Wiltshire state
// transition matrix. Mean motion is rad/s and dt is seconds.
func CWStateTransitionMatrix(meanMotion, dt float64) ([6][6]float64, error) {
	value, err := native.CWSTM(meanMotion, dt)
	return value, publicError(err)
}

// CWPropagate propagates a relative Cartesian state with the Clohessy-Wiltshire
// model. Position/velocity are km and km/s; mean motion is rad/s and dt is s.
func CWPropagate(state CartesianState, meanMotion, dt float64) (CartesianState, error) {
	value, err := native.CWPropagate(nativeCartesian(state), meanMotion, dt)
	return publicCartesian(value), publicError(err)
}

// RelativeMeanMotionCircular returns rad/s for a circular orbit of radius km.
func RelativeMeanMotionCircular(radiusKm float64) (float64, error) {
	value, err := native.RelativeMeanMotionCircular(radiusKm)
	return value, publicError(err)
}

// RelativeMeanMotionFromState derives rad/s from a chief ECI state.
func RelativeMeanMotionFromState(chief CartesianState) (float64, error) {
	value, err := native.RelativeMeanMotionFromState(nativeCartesian(chief))
	return value, publicError(err)
}

// RelativeFrame selects the local frame used by RelativeRotation.
type RelativeFrame uint32

const (
	// RelativeFrameRSW identifies the relative frame rsw case.
	RelativeFrameRSW RelativeFrame = RelativeFrame(native.RelativeFrameRSWValue)
	// RelativeFrameRTN identifies the relative frame rtn case.
	RelativeFrameRTN RelativeFrame = RelativeFrame(native.RelativeFrameRTNValue)
	// RelativeFrameRIC identifies the relative frame ric case.
	RelativeFrameRIC RelativeFrame = RelativeFrame(native.RelativeFrameRICValue)
	// RelativeFrameLVLH identifies the relative frame lvlh case.
	RelativeFrameLVLH RelativeFrame = RelativeFrame(native.RelativeFrameLVLHValue)
)

// RelativeRotation returns a row-major 3x3 local-to-inertial rotation for a
// chief state.
func RelativeRotation(frame RelativeFrame, chief CartesianState) ([3][3]float64, error) {
	value, err := native.RelativeRotation(uint32(frame), nativeCartesian(chief))
	return value, publicError(err)
}

// RelativeState returns deputy minus chief in the C relative-frame state
// convention. The returned EpochTDBSeconds is zero because the C route is
// epoch-free.
func RelativeState(chief, deputy CartesianState) (CartesianState, error) {
	value, err := native.RelativeState(nativeCartesian(chief), nativeCartesian(deputy))
	return publicCartesian(value), publicError(err)
}

// AbsoluteFromRelative reconstructs a deputy state from a chief and relative
// state.
func AbsoluteFromRelative(chief, relative CartesianState) (CartesianState, error) {
	value, err := native.AbsoluteFromRelative(nativeCartesian(chief), nativeCartesian(relative))
	return publicCartesian(value), publicError(err)
}
