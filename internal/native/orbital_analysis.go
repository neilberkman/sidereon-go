//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import "unsafe"

type CartesianState struct {
	EpochTDBSeconds float64
	PositionKm      [3]float64
	VelocityKmPerS  [3]float64
}

type ClassicalElements struct {
	P         float64
	A         float64
	Ecc       float64
	Incl      float64
	RAAN      float64
	ArgP      float64
	Nu        float64
	ArgLat    float64
	TrueLon   float64
	LonPer    float64
	OrbitType uint32
}

type EquinoctialElements struct {
	A          float64
	H          float64
	K          float64
	P          float64
	Q          float64
	Lambda     float64
	Retrograde uint32
}

type ModifiedEquinoctialElements struct {
	P          float64
	F          float64
	G          float64
	H          float64
	K          float64
	L          float64
	Retrograde uint32
}

type KeplerSolution struct {
	AnomalyRad float64
	Iterations int
}

type EncounterFrame struct {
	XHat                   [3]float64
	YHat                   [3]float64
	ZHat                   [3]float64
	RelativePositionKm     [3]float64
	RelativeVelocityKmPerS [3]float64
	MissKm                 float64
	RelativeSpeedKmPerS    float64
}

type ConjunctionState struct {
	PositionKm     [3]float64
	VelocityKmPerS [3]float64
	CovarianceKm2  [3][3]float64
}

type CollisionPC struct {
	PC                  float64
	MissKm              float64
	RelativeSpeedKmPerS float64
	SigmaXKm            float64
	SigmaZKm            float64
}

type TCAFinderOptions struct {
	CoarseStepSeconds    float64
	TimeToleranceSeconds float64
}

type TCACandidate struct {
	TCATimeJDWhole             float64
	TCATimeJDFraction          float64
	TCASecondsSinceWindowStart float64
	MissDistanceKm             float64
	RelativePositionKm         [3]float64
	RelativeVelocityKmPerS     [3]float64
}

type TCAPCOptions struct {
	HardBodyRadiusKm       float64
	Method                 uint32
	UseDefaultCovariance   bool
	PrimaryCovarianceKm2   [3][3]float64
	SecondaryCovarianceKm2 [3][3]float64
}

type TCAPropagatedCovariancePCOptions struct {
	HardBodyRadiusKm     float64
	Method               uint32
	PrimaryCovariance0   [6][6]float64
	SecondaryCovariance0 [6][6]float64
	ForceModel           uint32
	Integrator           uint32
	AbsTol               float64
	RelTol               float64
	InitialStepSeconds   float64
	MinStepSeconds       float64
	MaxStepSeconds       float64
	MaxSteps             uint32
	MuKm3PerS2Enabled    bool
	MuKm3PerS2           float64
}

type TCATLEPair struct {
	Line1 string
	Line2 string
}

type TCAScreeningHit struct {
	SecondaryIndex int
	Candidate      TCACandidate
}

type TCAScreeningConjunctionHit struct {
	SecondaryIndex int
	Conjunction    TCAConjunction
}

type TCAConjunction struct {
	Candidate            TCACandidate
	CollisionProbability CollisionPC
}

const (
	OrbitTypeEllipticalInclinedValue   = uint32(C.SIDEREON_ORBIT_TYPE_ELLIPTICAL_INCLINED)
	OrbitTypeEllipticalEquatorialValue = uint32(C.SIDEREON_ORBIT_TYPE_ELLIPTICAL_EQUATORIAL)
	OrbitTypeCircularInclinedValue     = uint32(C.SIDEREON_ORBIT_TYPE_CIRCULAR_INCLINED)
	OrbitTypeCircularEquatorialValue   = uint32(C.SIDEREON_ORBIT_TYPE_CIRCULAR_EQUATORIAL)

	RetrogradeFactorProgradeValue   = uint32(C.SIDEREON_RETROGRADE_FACTOR_PROGRADE)
	RetrogradeFactorRetrogradeValue = uint32(C.SIDEREON_RETROGRADE_FACTOR_RETROGRADE)

	RelativeFrameRSWValue  = uint32(C.SIDEREON_RELATIVE_FRAME_RSW)
	RelativeFrameRTNValue  = uint32(C.SIDEREON_RELATIVE_FRAME_RTN)
	RelativeFrameRICValue  = uint32(C.SIDEREON_RELATIVE_FRAME_RIC)
	RelativeFrameLVLHValue = uint32(C.SIDEREON_RELATIVE_FRAME_LVLH)

	PCMethodFosterEqualAreaValue = uint32(C.SIDEREON_PC_METHOD_FOSTER_EQUAL_AREA)
	PCMethodFosterNumericalValue = uint32(C.SIDEREON_PC_METHOD_FOSTER_NUMERICAL)
	PCMethodAlfano2005Value      = uint32(C.SIDEREON_PC_METHOD_ALFANO2005)

	PropagationForceModelTwoBodyValue     = uint32(C.SIDEREON_PROPAGATION_FORCE_MODEL_TWO_BODY)
	PropagationForceModelTwoBodyJ2Value   = uint32(C.SIDEREON_PROPAGATION_FORCE_MODEL_TWO_BODY_J2)
	PropagationForceModelCompositeValue   = uint32(C.SIDEREON_PROPAGATION_FORCE_MODEL_COMPOSITE)
	PropagationForceModelEarthPhaseAValue = uint32(C.SIDEREON_PROPAGATION_FORCE_MODEL_EARTH_PHASE_A)
	PropagationForceModelEarthPhaseBValue = uint32(C.SIDEREON_PROPAGATION_FORCE_MODEL_EARTH_PHASE_B)

	PropagationIntegratorDP54Value = uint32(C.SIDEREON_PROPAGATION_INTEGRATOR_DP54)
	PropagationIntegratorRK4Value  = uint32(C.SIDEREON_PROPAGATION_INTEGRATOR_RK4)
)

func cCartesianState(value CartesianState) C.SidereonCartesianState {
	var out C.SidereonCartesianState
	out.epoch_s = C.double(value.EpochTDBSeconds)
	for i := 0; i < 3; i++ {
		out.position_km[i] = C.double(value.PositionKm[i])
		out.velocity_km_s[i] = C.double(value.VelocityKmPerS[i])
	}
	return out
}

func cartesianStateFromC(value C.SidereonCartesianState) CartesianState {
	var out CartesianState
	out.EpochTDBSeconds = float64(value.epoch_s)
	for i := 0; i < 3; i++ {
		out.PositionKm[i] = float64(value.position_km[i])
		out.VelocityKmPerS[i] = float64(value.velocity_km_s[i])
	}
	return out
}

func cOrbitType(value uint32) (C.uint32_t, error) {
	switch value {
	case OrbitTypeEllipticalInclinedValue,
		OrbitTypeEllipticalEquatorialValue,
		OrbitTypeCircularInclinedValue,
		OrbitTypeCircularEquatorialValue:
		return C.uint32_t(value), nil
	default:
		return 0, invalidArgument("orbit type is not defined by the C ABI")
	}
}

func cRetrogradeFactor(value uint32) (C.uint32_t, error) {
	switch value {
	case RetrogradeFactorProgradeValue,
		RetrogradeFactorRetrogradeValue:
		return C.uint32_t(value), nil
	default:
		return 0, invalidArgument("retrograde factor is not defined by the C ABI")
	}
}

func cClassicalElements(value ClassicalElements) (C.SidereonClassicalElements, error) {
	orbitType, err := cOrbitType(value.OrbitType)
	if err != nil {
		return C.SidereonClassicalElements{}, err
	}
	return C.SidereonClassicalElements{
		p: C.double(value.P), a: C.double(value.A), ecc: C.double(value.Ecc), incl: C.double(value.Incl),
		raan: C.double(value.RAAN), argp: C.double(value.ArgP), nu: C.double(value.Nu),
		arglat: C.double(value.ArgLat), truelon: C.double(value.TrueLon), lonper: C.double(value.LonPer),
		orbit_type: orbitType,
	}, nil
}

func classicalElementsFromC(value C.SidereonClassicalElements) ClassicalElements {
	return ClassicalElements{
		P: float64(value.p), A: float64(value.a), Ecc: float64(value.ecc), Incl: float64(value.incl),
		RAAN: float64(value.raan), ArgP: float64(value.argp), Nu: float64(value.nu),
		ArgLat: float64(value.arglat), TrueLon: float64(value.truelon), LonPer: float64(value.lonper),
		OrbitType: uint32(value.orbit_type),
	}
}

func cEquinoctialElements(value EquinoctialElements) (C.SidereonEquinoctialElements, error) {
	retrograde, err := cRetrogradeFactor(value.Retrograde)
	if err != nil {
		return C.SidereonEquinoctialElements{}, err
	}
	return C.SidereonEquinoctialElements{
		a: C.double(value.A), h: C.double(value.H), k: C.double(value.K), p: C.double(value.P),
		q: C.double(value.Q), lambda: C.double(value.Lambda), retrograde: retrograde,
	}, nil
}

func equinoctialElementsFromC(value C.SidereonEquinoctialElements) EquinoctialElements {
	return EquinoctialElements{
		A: float64(value.a), H: float64(value.h), K: float64(value.k), P: float64(value.p),
		Q: float64(value.q), Lambda: float64(value.lambda), Retrograde: uint32(value.retrograde),
	}
}

func cModifiedEquinoctialElements(value ModifiedEquinoctialElements) (C.SidereonModifiedEquinoctialElements, error) {
	retrograde, err := cRetrogradeFactor(value.Retrograde)
	if err != nil {
		return C.SidereonModifiedEquinoctialElements{}, err
	}
	return C.SidereonModifiedEquinoctialElements{
		p: C.double(value.P), f: C.double(value.F), g: C.double(value.G), h: C.double(value.H),
		k: C.double(value.K), l: C.double(value.L), retrograde: retrograde,
	}, nil
}

func modifiedEquinoctialElementsFromC(value C.SidereonModifiedEquinoctialElements) ModifiedEquinoctialElements {
	return ModifiedEquinoctialElements{
		P: float64(value.p), F: float64(value.f), G: float64(value.g), H: float64(value.h),
		K: float64(value.k), L: float64(value.l), Retrograde: uint32(value.retrograde),
	}
}

func cConjunctionState(value ConjunctionState) C.SidereonConjunctionState {
	var out C.SidereonConjunctionState
	for i := 0; i < 3; i++ {
		out.position_km[i] = C.double(value.PositionKm[i])
		out.velocity_km_s[i] = C.double(value.VelocityKmPerS[i])
		for j := 0; j < 3; j++ {
			out.covariance_km2[i][j] = C.double(value.CovarianceKm2[i][j])
		}
	}
	return out
}

func encounterFrameFromC(value C.SidereonEncounterFrame) EncounterFrame {
	var out EncounterFrame
	for i := 0; i < 3; i++ {
		out.XHat[i] = float64(value.x_hat[i])
		out.YHat[i] = float64(value.y_hat[i])
		out.ZHat[i] = float64(value.z_hat[i])
		out.RelativePositionKm[i] = float64(value.relative_position_km[i])
		out.RelativeVelocityKmPerS[i] = float64(value.relative_velocity_km_s[i])
	}
	out.MissKm = float64(value.miss_km)
	out.RelativeSpeedKmPerS = float64(value.relative_speed_km_s)
	return out
}

func collisionPCFromC(value C.SidereonCollisionPc) CollisionPC {
	return CollisionPC{
		PC: float64(value.pc), MissKm: float64(value.miss_km),
		RelativeSpeedKmPerS: float64(value.relative_speed_km_s), SigmaXKm: float64(value.sigma_x_km), SigmaZKm: float64(value.sigma_z_km),
	}
}

func tcaCandidateFromC(value C.SidereonTcaCandidate) TCACandidate {
	var out TCACandidate
	out.TCATimeJDWhole = float64(value.tca_time_jd_whole)
	out.TCATimeJDFraction = float64(value.tca_time_jd_fraction)
	out.TCASecondsSinceWindowStart = float64(value.tca_seconds_since_window_start)
	out.MissDistanceKm = float64(value.miss_distance_km)
	for i := 0; i < 3; i++ {
		out.RelativePositionKm[i] = float64(value.relative_position_km[i])
		out.RelativeVelocityKmPerS[i] = float64(value.relative_velocity_km_s[i])
	}
	return out
}

func tcaConjunctionFromC(value C.SidereonTcaConjunction) TCAConjunction {
	return TCAConjunction{Candidate: tcaCandidateFromC(value.candidate), CollisionProbability: collisionPCFromC(value.collision_probability)}
}

func CartesianToClassical(state CartesianState, muKm3S2 float64) (ClassicalElements, error) {
	r := [3]C.double{}
	v := [3]C.double{}
	for i := 0; i < 3; i++ {
		r[i] = C.double(state.PositionKm[i])
		v[i] = C.double(state.VelocityKmPerS[i])
	}
	var out C.SidereonClassicalElements
	err := callStatus(func() uint32 { return C.sidereon_rv2coe(&r[0], &v[0], C.double(muKm3S2), &out) })
	return classicalElementsFromC(out), err
}

func ClassicalToCartesian(elements ClassicalElements, muKm3S2 float64) (CartesianState, error) {
	coe, err := cClassicalElements(elements)
	if err != nil {
		return CartesianState{}, err
	}
	var r, v [3]C.double
	err = callStatus(func() uint32 { return C.sidereon_coe2rv(&coe, C.double(muKm3S2), &r[0], &v[0]) })
	var out CartesianState
	for i := 0; i < 3; i++ {
		out.PositionKm[i] = float64(r[i])
		out.VelocityKmPerS[i] = float64(v[i])
	}
	return out, err
}

func ClassicalToEquinoctial(elements ClassicalElements, retrograde uint32) (EquinoctialElements, error) {
	coe, err := cClassicalElements(elements)
	if err != nil {
		return EquinoctialElements{}, err
	}
	retrogradeC, err := cRetrogradeFactor(retrograde)
	if err != nil {
		return EquinoctialElements{}, err
	}
	var out C.SidereonEquinoctialElements
	err = callStatus(func() uint32 { return C.sidereon_coe2eq(&coe, retrogradeC, &out) })
	return equinoctialElementsFromC(out), err
}

func EquinoctialToClassical(elements EquinoctialElements) (ClassicalElements, error) {
	eq, err := cEquinoctialElements(elements)
	if err != nil {
		return ClassicalElements{}, err
	}
	var out C.SidereonClassicalElements
	err = callStatus(func() uint32 { return C.sidereon_eq2coe(&eq, &out) })
	return classicalElementsFromC(out), err
}

func ClassicalToModifiedEquinoctial(elements ClassicalElements, retrograde uint32) (ModifiedEquinoctialElements, error) {
	coe, err := cClassicalElements(elements)
	if err != nil {
		return ModifiedEquinoctialElements{}, err
	}
	retrogradeC, err := cRetrogradeFactor(retrograde)
	if err != nil {
		return ModifiedEquinoctialElements{}, err
	}
	var out C.SidereonModifiedEquinoctialElements
	err = callStatus(func() uint32 { return C.sidereon_coe2mee(&coe, retrogradeC, &out) })
	return modifiedEquinoctialElementsFromC(out), err
}

func ModifiedEquinoctialToClassical(elements ModifiedEquinoctialElements) (ClassicalElements, error) {
	mee, err := cModifiedEquinoctialElements(elements)
	if err != nil {
		return ClassicalElements{}, err
	}
	var out C.SidereonClassicalElements
	err = callStatus(func() uint32 { return C.sidereon_mee2coe(&mee, &out) })
	return classicalElementsFromC(out), err
}

func CartesianToEquinoctial(state CartesianState, muKm3S2 float64, retrograde uint32) (EquinoctialElements, error) {
	retrogradeC, err := cRetrogradeFactor(retrograde)
	if err != nil {
		return EquinoctialElements{}, err
	}
	r := [3]C.double{}
	v := [3]C.double{}
	for i := 0; i < 3; i++ {
		r[i] = C.double(state.PositionKm[i])
		v[i] = C.double(state.VelocityKmPerS[i])
	}
	var out C.SidereonEquinoctialElements
	err = callStatus(func() uint32 { return C.sidereon_rv2eq(&r[0], &v[0], C.double(muKm3S2), retrogradeC, &out) })
	return equinoctialElementsFromC(out), err
}

func EquinoctialToCartesian(elements EquinoctialElements, muKm3S2 float64) (CartesianState, error) {
	eq, err := cEquinoctialElements(elements)
	if err != nil {
		return CartesianState{}, err
	}
	var r, v [3]C.double
	err = callStatus(func() uint32 { return C.sidereon_eq2rv(&eq, C.double(muKm3S2), &r[0], &v[0]) })
	var out CartesianState
	for i := 0; i < 3; i++ {
		out.PositionKm[i] = float64(r[i])
		out.VelocityKmPerS[i] = float64(v[i])
	}
	return out, err
}

func CartesianToModifiedEquinoctial(state CartesianState, muKm3S2 float64, retrograde uint32) (ModifiedEquinoctialElements, error) {
	retrogradeC, err := cRetrogradeFactor(retrograde)
	if err != nil {
		return ModifiedEquinoctialElements{}, err
	}
	r := [3]C.double{}
	v := [3]C.double{}
	for i := 0; i < 3; i++ {
		r[i] = C.double(state.PositionKm[i])
		v[i] = C.double(state.VelocityKmPerS[i])
	}
	var out C.SidereonModifiedEquinoctialElements
	err = callStatus(func() uint32 { return C.sidereon_rv2mee(&r[0], &v[0], C.double(muKm3S2), retrogradeC, &out) })
	return modifiedEquinoctialElementsFromC(out), err
}

func ModifiedEquinoctialToCartesian(elements ModifiedEquinoctialElements, muKm3S2 float64) (CartesianState, error) {
	mee, err := cModifiedEquinoctialElements(elements)
	if err != nil {
		return CartesianState{}, err
	}
	var r, v [3]C.double
	err = callStatus(func() uint32 { return C.sidereon_mee2rv(&mee, C.double(muKm3S2), &r[0], &v[0]) })
	var out CartesianState
	for i := 0; i < 3; i++ {
		out.PositionKm[i] = float64(r[i])
		out.VelocityKmPerS[i] = float64(v[i])
	}
	return out, err
}

func anomalyCall(meanOrTrue, eccentricity float64, kind uint32) (float64, error) {
	var out C.double
	err := callStatus(func() uint32 {
		switch kind {
		case 0:
			return C.sidereon_mean_to_eccentric_anomaly(C.double(meanOrTrue), C.double(eccentricity), &out)
		case 1:
			return C.sidereon_eccentric_to_mean_anomaly(C.double(meanOrTrue), C.double(eccentricity), &out)
		case 2:
			return C.sidereon_eccentric_to_true_anomaly(C.double(meanOrTrue), C.double(eccentricity), &out)
		case 3:
			return C.sidereon_true_to_eccentric_anomaly(C.double(meanOrTrue), C.double(eccentricity), &out)
		case 4:
			return C.sidereon_mean_to_true_anomaly(C.double(meanOrTrue), C.double(eccentricity), &out)
		default:
			return C.sidereon_true_to_mean_anomaly(C.double(meanOrTrue), C.double(eccentricity), &out)
		}
	})
	return float64(out), err
}

func MeanToEccentricAnomaly(mean, eccentricity float64) (float64, error) {
	return anomalyCall(mean, eccentricity, 0)
}
func EccentricToMeanAnomaly(eccentric, eccentricity float64) (float64, error) {
	return anomalyCall(eccentric, eccentricity, 1)
}
func EccentricToTrueAnomaly(eccentric, eccentricity float64) (float64, error) {
	return anomalyCall(eccentric, eccentricity, 2)
}
func TrueToEccentricAnomaly(trueAnomaly, eccentricity float64) (float64, error) {
	return anomalyCall(trueAnomaly, eccentricity, 3)
}
func MeanToTrueAnomaly(mean, eccentricity float64) (float64, error) {
	return anomalyCall(mean, eccentricity, 4)
}
func TrueToMeanAnomaly(trueAnomaly, eccentricity float64) (float64, error) {
	return anomalyCall(trueAnomaly, eccentricity, 5)
}

func SolveKepler(mean, eccentricity float64) (KeplerSolution, error) {
	var out C.SidereonKeplerSolution
	err := callStatus(func() uint32 { return C.sidereon_solve_kepler(C.double(mean), C.double(eccentricity), &out) })
	if err != nil {
		return KeplerSolution{}, err
	}
	iterations, err := checkedNativeCount(uint64(out.iterations))
	if err != nil {
		return KeplerSolution{}, err
	}
	return KeplerSolution{AnomalyRad: float64(out.anomaly_rad), Iterations: iterations}, nil
}

func PropagateKepler(elements ClassicalElements, muKm3S2, dtS float64) (ClassicalElements, error) {
	coe, err := cClassicalElements(elements)
	if err != nil {
		return ClassicalElements{}, err
	}
	var out C.SidereonClassicalElements
	err = callStatus(func() uint32 { return C.sidereon_propagate_kepler(&coe, C.double(muKm3S2), C.double(dtS), &out) })
	return classicalElementsFromC(out), err
}

func LambertBattin(r1, r2, v1 [3]float64, dm, de uint32, nrev int32, dtS float64) ([3]float64, [3]float64, error) {
	cr1, cr2, cv1 := [3]C.double{}, [3]C.double{}, [3]C.double{}
	for i := 0; i < 3; i++ {
		cr1[i] = C.double(r1[i])
		cr2[i] = C.double(r2[i])
		cv1[i] = C.double(v1[i])
	}
	var out1, out2 [3]C.double
	err := callStatus(func() uint32 {
		return C.sidereon_lambert_battin(&cr1[0], &cr2[0], &cv1[0], C.uint32_t(dm), C.uint32_t(de), C.int32_t(nrev), C.double(dtS), &out1[0], &out2[0])
	})
	var vout1, vout2 [3]float64
	for i := 0; i < 3; i++ {
		vout1[i] = float64(out1[i])
		vout2[i] = float64(out2[i])
	}
	return vout1, vout2, err
}

func CWSTM(meanMotion, dtS float64) ([6][6]float64, error) {
	var values [36]C.double
	err := callStatus(func() uint32 {
		return C.sidereon_cw_stm(C.double(meanMotion), C.double(dtS), &values[0], C.size_t(len(values)))
	})
	var out [6][6]float64
	for i := 0; i < 6; i++ {
		for j := 0; j < 6; j++ {
			out[i][j] = float64(values[i*6+j])
		}
	}
	return out, err
}

func CWPropagate(state CartesianState, meanMotion, dtS float64) (CartesianState, error) {
	in := cCartesianState(state)
	var out C.SidereonCartesianState
	err := callStatus(func() uint32 { return C.sidereon_cw_propagate(&in, C.double(meanMotion), C.double(dtS), &out) })
	return cartesianStateFromC(out), err
}

func RelativeMeanMotionCircular(radiusKm float64) (float64, error) {
	var out C.double
	err := callStatus(func() uint32 { return C.sidereon_relative_mean_motion_circular(C.double(radiusKm), &out) })
	return float64(out), err
}

func RelativeMeanMotionFromState(state CartesianState) (float64, error) {
	in := cCartesianState(state)
	var out C.double
	err := callStatus(func() uint32 { return C.sidereon_relative_mean_motion_from_state(&in, &out) })
	return float64(out), err
}

func RelativeRotation(frame uint32, state CartesianState) ([3][3]float64, error) {
	frameC, err := cRelativeFrame(frame)
	if err != nil {
		return [3][3]float64{}, err
	}
	in := cCartesianState(state)
	var values [9]C.double
	err = callStatus(func() uint32 {
		return C.sidereon_relative_rotation(frameC, &in, &values[0], C.size_t(len(values)))
	})
	var out [3][3]float64
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			out[i][j] = float64(values[i*3+j])
		}
	}
	return out, err
}

func cRelativeFrame(value uint32) (C.uint32_t, error) {
	switch value {
	case RelativeFrameRSWValue,
		RelativeFrameRTNValue,
		RelativeFrameRICValue,
		RelativeFrameLVLHValue:
		return C.uint32_t(value), nil
	default:
		return 0, invalidArgument("relative frame is not defined by the C ABI")
	}
}

func RelativeState(chief, deputy CartesianState) (CartesianState, error) {
	c, d := cCartesianState(chief), cCartesianState(deputy)
	var out C.SidereonCartesianState
	err := callStatus(func() uint32 { return C.sidereon_relative_state(&c, &d, &out) })
	return cartesianStateFromC(out), err
}

func AbsoluteFromRelative(chief, relative CartesianState) (CartesianState, error) {
	c, r := cCartesianState(chief), cCartesianState(relative)
	var out C.SidereonCartesianState
	err := callStatus(func() uint32 { return C.sidereon_absolute_from_relative(&c, &r, &out) })
	return cartesianStateFromC(out), err
}

func BetaAngleDeg(orbitNormal, sun [3]float64) (float64, error) {
	var normal, sunC [3]C.double
	for i := 0; i < 3; i++ {
		normal[i] = C.double(orbitNormal[i])
		sunC[i] = C.double(sun[i])
	}
	var out C.double
	err := callStatus(func() uint32 {
		return C.sidereon_beta_angle_deg(&normal[0], &sunC[0], &out)
	})
	return float64(out), err
}

func BetaAngleFromStateDeg(positionKm, velocityKmPerS, sunKm [3]float64) (float64, error) {
	var position, velocity, sunC [3]C.double
	for i := 0; i < 3; i++ {
		position[i] = C.double(positionKm[i])
		velocity[i] = C.double(velocityKmPerS[i])
		sunC[i] = C.double(sunKm[i])
	}
	var out C.double
	err := callStatus(func() uint32 {
		return C.sidereon_beta_angle_from_state_deg(&position[0], &velocity[0], &sunC[0], &out)
	})
	return float64(out), err
}

func RTNToECICovariance(covarianceRTN [3][3]float64, positionKm, velocityKmPerS [3]float64) ([3][3]float64, error) {
	var covariance [3][3]C.double
	var positionVector, velocityVector [3]C.double
	for i := 0; i < 3; i++ {
		positionVector[i] = C.double(positionKm[i])
		velocityVector[i] = C.double(velocityKmPerS[i])
		for j := 0; j < 3; j++ {
			covariance[i][j] = C.double(covarianceRTN[i][j])
		}
	}
	var flattened [9]C.double
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			flattened[i*3+j] = covariance[i][j]
		}
	}
	var out [9]C.double
	err := callStatus(func() uint32 {
		return C.sidereon_rtn_to_eci_covariance(&flattened[0], &positionVector[0], &velocityVector[0], &out[0])
	})
	var result [3][3]float64
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			result[i][j] = float64(out[i*3+j])
		}
	}
	return result, err
}

func BuildEncounterFrame(r1, v1, r2, v2 [3]float64) (EncounterFrame, error) {
	cr1, cv1, cr2, cv2 := [3]C.double{}, [3]C.double{}, [3]C.double{}, [3]C.double{}
	for i := 0; i < 3; i++ {
		cr1[i] = C.double(r1[i])
		cv1[i] = C.double(v1[i])
		cr2[i] = C.double(r2[i])
		cv2[i] = C.double(v2[i])
	}
	var out C.SidereonEncounterFrame
	err := callStatus(func() uint32 { return C.sidereon_encounter_frame(&cr1[0], &cv1[0], &cr2[0], &cv2[0], &out) })
	return encounterFrameFromC(out), err
}

func EncounterPlaneCovariance(frame EncounterFrame, covariance [3][3]float64) ([2][2]float64, error) {
	f := C.SidereonEncounterFrame{}
	for i := 0; i < 3; i++ {
		f.x_hat[i] = C.double(frame.XHat[i])
		f.y_hat[i] = C.double(frame.YHat[i])
		f.z_hat[i] = C.double(frame.ZHat[i])
		f.relative_position_km[i] = C.double(frame.RelativePositionKm[i])
		f.relative_velocity_km_s[i] = C.double(frame.RelativeVelocityKmPerS[i])
	}
	f.miss_km, f.relative_speed_km_s = C.double(frame.MissKm), C.double(frame.RelativeSpeedKmPerS)
	var cov [9]C.double
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			cov[i*3+j] = C.double(covariance[i][j])
		}
	}
	var out [4]C.double
	err := callStatus(func() uint32 { return C.sidereon_encounter_plane_covariance(&f, &cov[0], &out[0]) })
	return [2][2]float64{{float64(out[0]), float64(out[1])}, {float64(out[2]), float64(out[3])}}, err
}

func cPCMethod(value uint32) (C.uint32_t, error) {
	switch value {
	case PCMethodFosterEqualAreaValue,
		PCMethodFosterNumericalValue,
		PCMethodAlfano2005Value:
		return C.uint32_t(value), nil
	default:
		return 0, invalidArgument("PC method is not defined by the C ABI")
	}
}

func CollisionProbability(object1, object2 ConjunctionState, hardBodyRadiusKm float64, method uint32) (CollisionPC, error) {
	methodC, err := cPCMethod(method)
	if err != nil {
		return CollisionPC{}, err
	}
	o1, o2 := cConjunctionState(object1), cConjunctionState(object2)
	var out C.SidereonCollisionPc
	err = callStatus(func() uint32 {
		return C.sidereon_collision_probability(&o1, &o2, C.double(hardBodyRadiusKm), methodC, &out)
	})
	return collisionPCFromC(out), err
}

func TCAFinderDefaults() (TCAFinderOptions, error) {
	var out C.SidereonTcaFinderOptions
	err := callStatus(func() uint32 { return C.sidereon_tca_finder_options_init(&out) })
	return TCAFinderOptions{CoarseStepSeconds: float64(out.coarse_step_seconds), TimeToleranceSeconds: float64(out.time_tolerance_seconds)}, err
}

func cTCAFinderOptions(value TCAFinderOptions) C.SidereonTcaFinderOptions {
	return C.SidereonTcaFinderOptions{coarse_step_seconds: C.double(value.CoarseStepSeconds), time_tolerance_seconds: C.double(value.TimeToleranceSeconds)}
}

func cTLEStrings(lines [4]string) ([4]*C.char, func(), error) {
	var pointers [4]*C.char
	for _, line := range lines {
		if err := rejectEmbeddedNUL(line, "TLE line"); err != nil {
			return pointers, func() {}, err
		}
	}
	for i, line := range lines {
		pointers[i] = C.CString(line)
		if pointers[i] == nil {
			for j := 0; j < i; j++ {
				C.free(unsafe.Pointer(pointers[j]))
			}
			return [4]*C.char{}, func() {}, invalidArgument("unable to allocate native TLE string")
		}
	}
	return pointers, func() {
		for _, pointer := range pointers {
			C.free(unsafe.Pointer(pointer))
		}
	}, nil
}

func cTLECatalog(primaryLine1, primaryLine2 string, secondaries []TCATLEPair) (*C.char, *C.char, *C.SidereonTcaTlePair, func(), error) {
	if len(secondaries) > (int(^uint(0)>>1)-2)/2 {
		return nil, nil, nil, func() {}, invalidArgument("TLE catalog is too large")
	}
	lineValues := make([]string, 2+2*len(secondaries))
	lineValues[0], lineValues[1] = primaryLine1, primaryLine2
	for i, pair := range secondaries {
		lineValues[2+2*i], lineValues[3+2*i] = pair.Line1, pair.Line2
	}
	pointers, releaseStrings, err := cStrings(lineValues, "TLE catalog line")
	if err != nil {
		return nil, nil, nil, func() {}, err
	}
	pairSize, err := checkedNativeAllocationSize(len(secondaries), unsafe.Sizeof(C.SidereonTcaTlePair{}))
	if err != nil {
		releaseStrings()
		return nil, nil, nil, func() {}, err
	}
	var pairMemory unsafe.Pointer
	var pairPointer *C.SidereonTcaTlePair
	if pairSize > 0 {
		pairMemory = C.malloc(C.size_t(pairSize))
		if pairMemory == nil {
			releaseStrings()
			return nil, nil, nil, func() {}, invalidArgument("unable to allocate native TLE catalog")
		}
		pairPointer = (*C.SidereonTcaTlePair)(pairMemory)
		pairs := unsafe.Slice(pairPointer, len(secondaries))
		for i := range pairs {
			pairs[i].line1 = pointers[2+2*i]
			pairs[i].line2 = pointers[3+2*i]
		}
	}
	release := func() {
		if pairMemory != nil {
			C.free(pairMemory)
		}
		releaseStrings()
	}
	return pointers[0], pointers[1], pairPointer, release, nil
}

func cStrings(lines []string, label string) ([]*C.char, func(), error) {
	for _, line := range lines {
		if err := rejectEmbeddedNUL(line, label); err != nil {
			return nil, func() {}, err
		}
	}
	if _, err := checkedNativeAllocationSize(len(lines), unsafe.Sizeof(uintptr(0))); err != nil {
		return nil, func() {}, err
	}
	var memory unsafe.Pointer
	if len(lines) > 0 {
		memory = C.malloc(C.size_t(len(lines)) * C.size_t(unsafe.Sizeof(uintptr(0))))
		if memory == nil {
			return nil, func() {}, invalidArgument("unable to allocate native string pointer array")
		}
	}
	pointers := unsafe.Slice((**C.char)(memory), len(lines))
	for i, line := range lines {
		pointers[i] = C.CString(line)
		if pointers[i] == nil {
			for j := 0; j < i; j++ {
				C.free(unsafe.Pointer(pointers[j]))
			}
			if memory != nil {
				C.free(memory)
			}
			return nil, func() {}, invalidArgument("unable to allocate native string")
		}
	}
	return pointers, func() {
		for _, pointer := range pointers {
			C.free(unsafe.Pointer(pointer))
		}
		if memory != nil {
			C.free(memory)
		}
	}, nil
}

func cTCAFinderOptionsOrDefault(value *TCAFinderOptions) (C.SidereonTcaFinderOptions, error) {
	if value != nil {
		return cTCAFinderOptions(*value), nil
	}
	defaults, err := TCAFinderDefaults()
	if err != nil {
		return C.SidereonTcaFinderOptions{}, err
	}
	return cTCAFinderOptions(defaults), nil
}

func FindTCACandidates(lines [4]string, startWhole, startFraction, endWhole, endFraction float64, options *TCAFinderOptions) ([]TCACandidate, error) {
	cLines, release, err := cTLEStrings(lines)
	if err != nil {
		return nil, err
	}
	defer release()
	var cOptions C.SidereonTcaFinderOptions
	var optionsPointer *C.SidereonTcaFinderOptions
	if options != nil {
		cOptions = cTCAFinderOptions(*options)
		optionsPointer = &cOptions
	} else {
		defaults, err := TCAFinderDefaults()
		if err != nil {
			return nil, err
		}
		cOptions = cTCAFinderOptions(defaults)
		optionsPointer = &cOptions
	}
	var written, required C.size_t
	err = callStatus(func() uint32 {
		return C.sidereon_find_tca_candidates_from_tles(cLines[0], cLines[1], cLines[2], cLines[3], C.double(startWhole), C.double(startFraction), C.double(endWhole), C.double(endFraction), optionsPointer, nil, 0, &written, &required)
	})
	if err != nil {
		return nil, err
	}
	expected, err := validateNativeQuery("TCA candidates", uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(expected, unsafe.Sizeof(C.SidereonTcaCandidate{})); err != nil {
		return nil, err
	}
	values := make([]C.SidereonTcaCandidate, expected)
	var pointer *C.SidereonTcaCandidate
	if len(values) > 0 {
		pointer = &values[0]
	}
	written, required = 0, 0
	err = callStatus(func() uint32 {
		return C.sidereon_find_tca_candidates_from_tles(cLines[0], cLines[1], cLines[2], cLines[3], C.double(startWhole), C.double(startFraction), C.double(endWhole), C.double(endFraction), optionsPointer, pointer, C.size_t(len(values)), &written, &required)
	})
	if err != nil {
		return nil, err
	}
	writtenLength, err := validateTwoPassCounts("TCA candidates", len(values), expected, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	out := make([]TCACandidate, writtenLength)
	for i := range out {
		out[i] = tcaCandidateFromC(values[i])
	}
	return out, nil
}

func cTCAPCOptions(value TCAPCOptions) (C.SidereonTcaPcOptions, error) {
	var out C.SidereonTcaPcOptions
	method, err := cPCMethod(value.Method)
	if err != nil {
		return out, err
	}
	out.hard_body_radius_km, out.method, out.use_default_covariance = C.double(value.HardBodyRadiusKm), method, C.bool(value.UseDefaultCovariance)
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			out.primary_covariance_km2[i][j] = C.double(value.PrimaryCovarianceKm2[i][j])
			out.secondary_covariance_km2[i][j] = C.double(value.SecondaryCovarianceKm2[i][j])
		}
	}
	return out, nil
}

func TCACollisionProbability(candidate TCACandidate, options TCAPCOptions) (TCAConjunction, error) {
	c := C.SidereonTcaCandidate{tca_time_jd_whole: C.double(candidate.TCATimeJDWhole), tca_time_jd_fraction: C.double(candidate.TCATimeJDFraction), tca_seconds_since_window_start: C.double(candidate.TCASecondsSinceWindowStart), miss_distance_km: C.double(candidate.MissDistanceKm)}
	for i := 0; i < 3; i++ {
		c.relative_position_km[i] = C.double(candidate.RelativePositionKm[i])
		c.relative_velocity_km_s[i] = C.double(candidate.RelativeVelocityKmPerS[i])
	}
	pc, err := cTCAPCOptions(options)
	if err != nil {
		return TCAConjunction{}, err
	}
	var out C.SidereonTcaConjunction
	err = callStatus(func() uint32 { return C.sidereon_tca_collision_probability(&c, &pc, &out) })
	return tcaConjunctionFromC(out), err
}

func FindTCAConjunctions(lines [4]string, startWhole, startFraction, endWhole, endFraction float64, tcaOptions *TCAFinderOptions, pcOptions TCAPCOptions) ([]TCAConjunction, error) {
	cLines, release, err := cTLEStrings(lines)
	if err != nil {
		return nil, err
	}
	defer release()
	var cTcaOptions C.SidereonTcaFinderOptions
	var tcaOptionsPointer *C.SidereonTcaFinderOptions
	if tcaOptions != nil {
		cTcaOptions, tcaOptionsPointer = cTCAFinderOptions(*tcaOptions), &cTcaOptions
	} else {
		defaults, err := TCAFinderDefaults()
		if err != nil {
			return nil, err
		}
		cTcaOptions, tcaOptionsPointer = cTCAFinderOptions(defaults), &cTcaOptions
	}
	cPcOptions, err := cTCAPCOptions(pcOptions)
	if err != nil {
		return nil, err
	}
	var written, required C.size_t
	err = callStatus(func() uint32 {
		return C.sidereon_find_tca_conjunctions_from_tles(cLines[0], cLines[1], cLines[2], cLines[3], C.double(startWhole), C.double(startFraction), C.double(endWhole), C.double(endFraction), tcaOptionsPointer, &cPcOptions, nil, 0, &written, &required)
	})
	if err != nil {
		return nil, err
	}
	expected, err := validateNativeQuery("TCA conjunctions", uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(expected, unsafe.Sizeof(C.SidereonTcaConjunction{})); err != nil {
		return nil, err
	}
	values := make([]C.SidereonTcaConjunction, expected)
	var pointer *C.SidereonTcaConjunction
	if len(values) > 0 {
		pointer = &values[0]
	}
	written, required = 0, 0
	err = callStatus(func() uint32 {
		return C.sidereon_find_tca_conjunctions_from_tles(cLines[0], cLines[1], cLines[2], cLines[3], C.double(startWhole), C.double(startFraction), C.double(endWhole), C.double(endFraction), tcaOptionsPointer, &cPcOptions, pointer, C.size_t(len(values)), &written, &required)
	})
	if err != nil {
		return nil, err
	}
	writtenLength, err := validateTwoPassCounts("TCA conjunctions", len(values), expected, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	out := make([]TCAConjunction, writtenLength)
	for i := range out {
		out[i] = tcaConjunctionFromC(values[i])
	}
	return out, nil
}

func cPropagationForceModel(value uint32) (C.uint32_t, error) {
	switch value {
	case PropagationForceModelTwoBodyValue,
		PropagationForceModelTwoBodyJ2Value,
		PropagationForceModelCompositeValue,
		PropagationForceModelEarthPhaseAValue,
		PropagationForceModelEarthPhaseBValue:
		return C.uint32_t(value), nil
	default:
		return 0, invalidArgument("propagation force model is not defined by the C ABI")
	}
}

func cPropagationIntegrator(value uint32) (C.uint32_t, error) {
	switch value {
	case PropagationIntegratorDP54Value,
		PropagationIntegratorRK4Value:
		return C.uint32_t(value), nil
	default:
		return 0, invalidArgument("propagation integrator is not defined by the C ABI")
	}
}

func cTCAPropagatedCovariancePCOptions(value TCAPropagatedCovariancePCOptions) (C.SidereonTcaPropagatedCovariancePcOptions, error) {
	var out C.SidereonTcaPropagatedCovariancePcOptions
	method, err := cPCMethod(value.Method)
	if err != nil {
		return out, err
	}
	forceModel, err := cPropagationForceModel(value.ForceModel)
	if err != nil {
		return out, err
	}
	integrator, err := cPropagationIntegrator(value.Integrator)
	if err != nil {
		return out, err
	}
	out.hard_body_radius_km = C.double(value.HardBodyRadiusKm)
	out.method = method
	out.force_model = forceModel
	out.integrator = integrator
	out.abs_tol = C.double(value.AbsTol)
	out.rel_tol = C.double(value.RelTol)
	out.initial_step_s = C.double(value.InitialStepSeconds)
	out.min_step_s = C.double(value.MinStepSeconds)
	out.max_step_s = C.double(value.MaxStepSeconds)
	out.max_steps = C.uint32_t(value.MaxSteps)
	out.mu_km3_s2_enabled = C.bool(value.MuKm3PerS2Enabled)
	out.mu_km3_s2 = C.double(value.MuKm3PerS2)
	for i := 0; i < 6; i++ {
		for j := 0; j < 6; j++ {
			out.primary_covariance0[i][j] = C.double(value.PrimaryCovariance0[i][j])
			out.secondary_covariance0[i][j] = C.double(value.SecondaryCovariance0[i][j])
		}
	}
	return out, nil
}

func FindTCAConjunctionsWithPropagatedCovarianceFromTLEs(lines [4]string, startWhole, startFraction, endWhole, endFraction float64, tcaOptions *TCAFinderOptions, pcOptions TCAPropagatedCovariancePCOptions) ([]TCAConjunction, error) {
	cLines, release, err := cTLEStrings(lines)
	if err != nil {
		return nil, err
	}
	defer release()
	cTcaOptions, err := cTCAFinderOptionsOrDefault(tcaOptions)
	if err != nil {
		return nil, err
	}
	cPcOptions, err := cTCAPropagatedCovariancePCOptions(pcOptions)
	if err != nil {
		return nil, err
	}
	var written, required C.size_t
	call := func(output *C.SidereonTcaConjunction, length C.size_t) error {
		return callStatus(func() uint32 {
			return C.sidereon_find_tca_conjunctions_with_propagated_covariance_from_tles(
				cLines[0], cLines[1], cLines[2], cLines[3], C.double(startWhole), C.double(startFraction),
				C.double(endWhole), C.double(endFraction), &cTcaOptions, &cPcOptions, output, length, &written, &required,
			)
		})
	}
	if err := call(nil, 0); err != nil {
		return nil, err
	}
	expected, err := validateNativeQuery("propagated-covariance TCA conjunctions", uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(expected, unsafe.Sizeof(C.SidereonTcaConjunction{})); err != nil {
		return nil, err
	}
	values := make([]C.SidereonTcaConjunction, expected)
	var output *C.SidereonTcaConjunction
	if len(values) > 0 {
		output = &values[0]
	}
	written, required = 0, 0
	if err := call(output, C.size_t(len(values))); err != nil {
		return nil, err
	}
	count, err := validateTwoPassCounts("propagated-covariance TCA conjunctions", len(values), expected, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	result := make([]TCAConjunction, count)
	for i := range result {
		result[i] = tcaConjunctionFromC(values[i])
	}
	return result, nil
}

func screeningHitFromC(value C.SidereonTcaScreeningHit) (TCAScreeningHit, error) {
	index, err := checkedNativeCount(uint64(value.secondary_index))
	if err != nil {
		return TCAScreeningHit{}, err
	}
	return TCAScreeningHit{SecondaryIndex: index, Candidate: tcaCandidateFromC(value.candidate)}, nil
}

func screeningConjunctionHitFromC(value C.SidereonTcaScreeningConjunctionHit) (TCAScreeningConjunctionHit, error) {
	index, err := checkedNativeCount(uint64(value.secondary_index))
	if err != nil {
		return TCAScreeningConjunctionHit{}, err
	}
	return TCAScreeningConjunctionHit{SecondaryIndex: index, Conjunction: tcaConjunctionFromC(value.conjunction)}, nil
}

func ScreenTCACandidatesFromTLECatalog(primaryLine1, primaryLine2 string, secondaries []TCATLEPair, startWhole, startFraction, endWhole, endFraction, missDistanceThresholdKm float64, options *TCAFinderOptions) ([]TCAScreeningHit, error) {
	primary1, primary2, catalog, release, err := cTLECatalog(primaryLine1, primaryLine2, secondaries)
	if err != nil {
		return nil, err
	}
	defer release()
	cOptions, err := cTCAFinderOptionsOrDefault(options)
	if err != nil {
		return nil, err
	}
	var written, required C.size_t
	call := func(output *C.SidereonTcaScreeningHit, length C.size_t) error {
		return callStatus(func() uint32 {
			return C.sidereon_screen_tca_candidates_from_tle_catalog(primary1, primary2, catalog, C.size_t(len(secondaries)), C.double(startWhole), C.double(startFraction), C.double(endWhole), C.double(endFraction), C.double(missDistanceThresholdKm), &cOptions, output, length, &written, &required)
		})
	}
	if err := call(nil, 0); err != nil {
		return nil, err
	}
	expected, err := validateNativeQuery("TCA catalog candidate screening", uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(expected, unsafe.Sizeof(C.SidereonTcaScreeningHit{})); err != nil {
		return nil, err
	}
	values := make([]C.SidereonTcaScreeningHit, expected)
	var output *C.SidereonTcaScreeningHit
	if len(values) > 0 {
		output = &values[0]
	}
	written, required = 0, 0
	if err := call(output, C.size_t(len(values))); err != nil {
		return nil, err
	}
	count, err := validateTwoPassCounts("TCA catalog candidate screening", len(values), expected, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	result := make([]TCAScreeningHit, count)
	for i := range result {
		result[i], err = screeningHitFromC(values[i])
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func ScreenTCAConjunctionsFromTLECatalog(primaryLine1, primaryLine2 string, secondaries []TCATLEPair, startWhole, startFraction, endWhole, endFraction, missDistanceThresholdKm float64, tcaOptions *TCAFinderOptions, pcOptions TCAPCOptions) ([]TCAScreeningConjunctionHit, error) {
	primary1, primary2, catalog, release, err := cTLECatalog(primaryLine1, primaryLine2, secondaries)
	if err != nil {
		return nil, err
	}
	defer release()
	cTcaOptions, err := cTCAFinderOptionsOrDefault(tcaOptions)
	if err != nil {
		return nil, err
	}
	cPcOptions, err := cTCAPCOptions(pcOptions)
	if err != nil {
		return nil, err
	}
	var written, required C.size_t
	call := func(output *C.SidereonTcaScreeningConjunctionHit, length C.size_t) error {
		return callStatus(func() uint32 {
			return C.sidereon_screen_tca_conjunctions_from_tle_catalog(primary1, primary2, catalog, C.size_t(len(secondaries)), C.double(startWhole), C.double(startFraction), C.double(endWhole), C.double(endFraction), C.double(missDistanceThresholdKm), &cTcaOptions, &cPcOptions, output, length, &written, &required)
		})
	}
	if err := call(nil, 0); err != nil {
		return nil, err
	}
	expected, err := validateNativeQuery("TCA catalog conjunction screening", uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(expected, unsafe.Sizeof(C.SidereonTcaScreeningConjunctionHit{})); err != nil {
		return nil, err
	}
	values := make([]C.SidereonTcaScreeningConjunctionHit, expected)
	var output *C.SidereonTcaScreeningConjunctionHit
	if len(values) > 0 {
		output = &values[0]
	}
	written, required = 0, 0
	if err := call(output, C.size_t(len(values))); err != nil {
		return nil, err
	}
	count, err := validateTwoPassCounts("TCA catalog conjunction screening", len(values), expected, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	result := make([]TCAScreeningConjunctionHit, count)
	for i := range result {
		result[i], err = screeningConjunctionHitFromC(values[i])
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}
