package sidereon

import (
	"errors"
	"sidereon.dev/go/v4/internal/native"
)

// AnalyticTheory identifies a AnalyticTheory option.
type AnalyticTheory uint32

const (
	// AnalyticTheoryEcksteinHechler selects eckstein hechler.
	AnalyticTheoryEcksteinHechler AnalyticTheory = 0
	// AnalyticTheoryBrouwerLyddane selects brouwer lyddane.
	AnalyticTheoryBrouwerLyddane AnalyticTheory = 1
)

// AttitudeMode identifies a AttitudeMode option.
type AttitudeMode uint32

const (
	// AttitudeModeNominal selects nominal.
	AttitudeModeNominal AttitudeMode = 0
	// AttitudeModeNoonTurn selects noon turn.
	AttitudeModeNoonTurn AttitudeMode = 1
	// AttitudeModeMidnightTurn selects midnight turn.
	AttitudeModeMidnightTurn AttitudeMode = 2
	// AttitudeModeShadow selects shadow.
	AttitudeModeShadow AttitudeMode = 3
	// AttitudeModeOrbitNormal selects orbit normal.
	AttitudeModeOrbitNormal AttitudeMode = 4
	// AttitudeModeModifiedYawSteering selects modified yaw steering.
	AttitudeModeModifiedYawSteering AttitudeMode = 5
)

// BeiDouManufacturer identifies a BeiDouManufacturer option.
type BeiDouManufacturer uint32

const (
	// BeiDouManufacturerCAST selects cast.
	BeiDouManufacturerCAST BeiDouManufacturer = 0
	// BeiDouManufacturerSECM selects secm.
	BeiDouManufacturerSECM BeiDouManufacturer = 1
)

// BeiDouOrbitClass identifies a BeiDouOrbitClass option.
type BeiDouOrbitClass uint32

const (
	// BeiDouOrbitClassGEO selects geo.
	BeiDouOrbitClassGEO BeiDouOrbitClass = 0
	// BeiDouOrbitClassIGSO selects igso.
	BeiDouOrbitClassIGSO BeiDouOrbitClass = 1
	// BeiDouOrbitClassMEO selects meo.
	BeiDouOrbitClassMEO BeiDouOrbitClass = 2
)

// ClockDatum identifies a ClockDatum option.
type ClockDatum uint32

const (
	// ClockDatumEpochSystemMedian selects epoch system median.
	ClockDatumEpochSystemMedian ClockDatum = 0
	// ClockDatumUnadjusted selects unadjusted.
	ClockDatumUnadjusted ClockDatum = 1
)

// CosineTurnStartAngle identifies a CosineTurnStartAngle option.
type CosineTurnStartAngle uint32

const (
	// CosineTurnStartAngleColinearity selects colinearity.
	CosineTurnStartAngleColinearity CosineTurnStartAngle = 0
	// CosineTurnStartAngleOrbitAngle selects orbit angle.
	CosineTurnStartAngleOrbitAngle CosineTurnStartAngle = 1
)

// DEBody identifies a DEBody option.
type DEBody uint32

const (
	// DEBodySun selects sun.
	DEBodySun DEBody = 0
	// DEBodyMercury selects mercury.
	DEBodyMercury DEBody = 1
	// DEBodyVenus selects venus.
	DEBodyVenus DEBody = 2
	// DEBodyEarthMoon selects earth moon.
	DEBodyEarthMoon DEBody = 3
	// DEBodyEarth selects earth.
	DEBodyEarth DEBody = 4
	// DEBodyMoon selects moon.
	DEBodyMoon DEBody = 5
	// DEBodyMars selects mars.
	DEBodyMars DEBody = 6
	// DEBodyJupiter selects jupiter.
	DEBodyJupiter DEBody = 7
	// DEBodySaturn selects saturn.
	DEBodySaturn DEBody = 8
	// DEBodyUranus selects uranus.
	DEBodyUranus DEBody = 9
	// DEBodyNeptune selects neptune.
	DEBodyNeptune DEBody = 10
	// DEBodyPluto selects pluto.
	DEBodyPluto DEBody = 11
	// DEBodySolarSystem selects solar system.
	DEBodySolarSystem DEBody = 12
)

// EmpiricalHarmonic identifies a EmpiricalHarmonic option.
type EmpiricalHarmonic uint32

const (
	// EmpiricalHarmonicConstant selects constant.
	EmpiricalHarmonicConstant EmpiricalHarmonic = 0
	// EmpiricalHarmonicCosine selects cosine.
	EmpiricalHarmonicCosine EmpiricalHarmonic = 1
	// EmpiricalHarmonicSine selects sine.
	EmpiricalHarmonicSine EmpiricalHarmonic = 2
)

// FrameAgreementKind identifies a FrameAgreementKind option.
type FrameAgreementKind uint32

const (
	// FrameAgreementKindSameLabel selects same label.
	FrameAgreementKindSameLabel FrameAgreementKind = 0
	// FrameAgreementKindAssertedEquivalent selects asserted equivalent.
	FrameAgreementKindAssertedEquivalent FrameAgreementKind = 1
	// FrameAgreementKindHelmert selects helmert.
	FrameAgreementKindHelmert FrameAgreementKind = 2
	// FrameAgreementKindBroadcastDatums selects broadcast datums.
	FrameAgreementKindBroadcastDatums FrameAgreementKind = 3
)

// OrbitScoreGapPolicy identifies a OrbitScoreGapPolicy option.
type OrbitScoreGapPolicy uint32

const (
	// OrbitScoreGapPolicyCount selects count.
	OrbitScoreGapPolicyCount OrbitScoreGapPolicy = 0
	// OrbitScoreGapPolicyRefuse selects refuse.
	OrbitScoreGapPolicyRefuse OrbitScoreGapPolicy = 1
)

// GNSSBlock identifies a GNSSBlock option.
type GNSSBlock uint32

const (
	// GNSSBlockGPSIIA selects gps iia.
	GNSSBlockGPSIIA GNSSBlock = 0
	// GNSSBlockGPSIIR selects gps iir.
	GNSSBlockGPSIIR GNSSBlock = 1
	// GNSSBlockGPSIIRM selects gps iirm.
	GNSSBlockGPSIIRM GNSSBlock = 2
	// GNSSBlockGPSIIF selects gps iif.
	GNSSBlockGPSIIF GNSSBlock = 3
	// GNSSBlockGPSIII selects gps iii.
	GNSSBlockGPSIII GNSSBlock = 4
	// GNSSBlockGLONASSM selects glonass m.
	GNSSBlockGLONASSM GNSSBlock = 5
	// GNSSBlockGalileoIOV selects galileo iov.
	GNSSBlockGalileoIOV GNSSBlock = 6
	// GNSSBlockGalileoFOC selects galileo foc.
	GNSSBlockGalileoFOC GNSSBlock = 7
	// GNSSBlockBeiDou2GEO selects beidou2 geo.
	GNSSBlockBeiDou2GEO GNSSBlock = 8
	// GNSSBlockBeiDou2IGSO selects beidou2 igso.
	GNSSBlockBeiDou2IGSO GNSSBlock = 9
	// GNSSBlockBeiDou2MEO selects beidou2 meo.
	GNSSBlockBeiDou2MEO GNSSBlock = 10
	// GNSSBlockBeiDou3GEO selects beidou3 geo.
	GNSSBlockBeiDou3GEO GNSSBlock = 11
	// GNSSBlockBeiDou3IGSO selects beidou3 igso.
	GNSSBlockBeiDou3IGSO GNSSBlock = 12
	// GNSSBlockBeiDou3MEOUnknown selects beidou3 meo unknown.
	GNSSBlockBeiDou3MEOUnknown GNSSBlock = 13
	// GNSSBlockBeiDou3MEOCAST selects beidou3 meo cast.
	GNSSBlockBeiDou3MEOCAST GNSSBlock = 14
	// GNSSBlockBeiDou3MEOSECM selects beidou3 meo secm.
	GNSSBlockBeiDou3MEOSECM GNSSBlock = 15
)

// ICGEMErrors identifies a ICGEMErrors option.
type ICGEMErrors uint32

const (
	// ICGEMErrorsNo selects no.
	ICGEMErrorsNo ICGEMErrors = 0
	// ICGEMErrorsCalibrated selects calibrated.
	ICGEMErrorsCalibrated ICGEMErrors = 1
	// ICGEMErrorsFormal selects formal.
	ICGEMErrorsFormal ICGEMErrors = 2
	// ICGEMErrorsCalibratedAndFormal selects calibrated and formal.
	ICGEMErrorsCalibratedAndFormal ICGEMErrors = 3
)

// ICGEMFormat identifies a ICGEMFormat option.
type ICGEMFormat uint32

const (
	// ICGEMFormatICGEM1 selects icgem1.
	ICGEMFormatICGEM1 ICGEMFormat = 0
	// ICGEMFormatICGEM2 selects icgem2.
	ICGEMFormatICGEM2 ICGEMFormat = 1
)

// ICGEMNormalization identifies a ICGEMNormalization option.
type ICGEMNormalization uint32

const (
	// ICGEMNormalizationFullyNormalized selects fully normalized.
	ICGEMNormalizationFullyNormalized ICGEMNormalization = 0
	// ICGEMNormalizationUnnormalized selects unnormalized.
	ICGEMNormalizationUnnormalized ICGEMNormalization = 1
)

// ICGEMTermKind identifies a ICGEMTermKind option.
type ICGEMTermKind uint32

const (
	// ICGEMTermKindConstant selects constant.
	ICGEMTermKindConstant ICGEMTermKind = 0
	// ICGEMTermKindTrend selects trend.
	ICGEMTermKindTrend ICGEMTermKind = 1
	// ICGEMTermKindCosine selects cosine.
	ICGEMTermKindCosine ICGEMTermKind = 2
	// ICGEMTermKindSine selects sine.
	ICGEMTermKindSine ICGEMTermKind = 3
)

// OrbitFrameKind identifies a OrbitFrameKind option.
type OrbitFrameKind uint32

const (
	// OrbitFrameKindLabel selects label.
	OrbitFrameKindLabel OrbitFrameKind = 0
	// OrbitFrameKindBroadcastDatums selects broadcast datums.
	OrbitFrameKindBroadcastDatums OrbitFrameKind = 1
)

// OrbitReferencePoint identifies a OrbitReferencePoint option.
type OrbitReferencePoint uint32

const (
	// OrbitReferencePointCenterOfMass selects center of mass.
	OrbitReferencePointCenterOfMass OrbitReferencePoint = 0
	// OrbitReferencePointAntennaPhaseCenter selects antenna phase center.
	OrbitReferencePointAntennaPhaseCenter OrbitReferencePoint = 1
)

// OrbitRegime identifies a OrbitRegime option.
type OrbitRegime uint32

const (
	// OrbitRegimeMEO selects meo.
	OrbitRegimeMEO OrbitRegime = 0
	// OrbitRegimeIGSO selects igso.
	OrbitRegimeIGSO OrbitRegime = 1
	// OrbitRegimeGEO selects geo.
	OrbitRegimeGEO OrbitRegime = 2
)

// RTNAxis identifies a RTNAxis option.
type RTNAxis uint32

const (
	// RTNAxisRadial selects radial.
	RTNAxisRadial RTNAxis = 0
	// RTNAxisAlongTrack selects along track.
	RTNAxisAlongTrack RTNAxis = 1
	// RTNAxisCrossTrack selects cross track.
	RTNAxisCrossTrack RTNAxis = 2
)

// ScoredProduct identifies a ScoredProduct option.
type ScoredProduct uint32

const (
	// ScoredProductSource selects source.
	ScoredProductSource ScoredProduct = 0
	// ScoredProductReference selects reference.
	ScoredProductReference ScoredProduct = 1
)

// ShadowFunction identifies a ShadowFunction option.
type ShadowFunction uint32

const (
	// ShadowFunctionApparentDisc selects apparent disc.
	ShadowFunctionApparentDisc ShadowFunction = 0
	// ShadowFunctionConical selects conical.
	ShadowFunctionConical ShadowFunction = 1
)

// TriadVelocity identifies a TriadVelocity option.
type TriadVelocity uint32

const (
	// TriadVelocityInertial selects inertial.
	TriadVelocityInertial TriadVelocity = 0
	// TriadVelocityEarthFixed selects earth fixed.
	TriadVelocityEarthFixed TriadVelocity = 1
)

// Ut1Degradation identifies a Ut1Degradation option.
type Ut1Degradation uint32

const (
	// Ut1DegradationNone selects none.
	Ut1DegradationNone Ut1Degradation = 0
	// Ut1DegradationBeforeCoverage selects before coverage.
	Ut1DegradationBeforeCoverage Ut1Degradation = 1
	// Ut1DegradationAfterCoverage selects after coverage.
	Ut1DegradationAfterCoverage Ut1Degradation = 2
)

// AccelerationPartials contains acceleration derivatives in row-major order. Compare elements by value.
// Compare values with ==.
type AccelerationPartials struct {
	// WrtPosition is derivatives with respect to position.
	WrtPosition [9]float64
	// WrtVelocity is derivatives with respect to velocity.
	WrtVelocity [9]float64
}

// AnalyticCartesian contains inertial Cartesian coordinates for an analytic orbit.
// Compare values with ==.
type AnalyticCartesian struct {
	// PositionKm is position in kilometres.
	PositionKm [3]float64
	// VelocityKmS is velocity in kilometres per second.
	VelocityKmS [3]float64
}

// AnalyticKeplerianElements contains analytic orbit elements.
// Compare values with ==.
type AnalyticKeplerianElements struct {
	// AKm is semi-major axis in kilometres.
	AKm float64
	// E is eccentricity.
	E float64
	// IRad is inclination in radians.
	IRad float64
	// ArgPerigeeRad is argument of perigee in radians.
	ArgPerigeeRad float64
	// RaanRad is right ascension of the ascending node in radians.
	RaanRad float64
	// MeanAnomalyRad is mean anomaly in radians.
	MeanAnomalyRad float64
}

// AttitudeSample contains satellite and Sun geometry in a common geocentric inertial frame.
// Compare values with ==.
type AttitudeSample struct {
	// PositionM is satellite position in metres.
	PositionM [3]float64
	// VelocityMS is satellite velocity in metres per second.
	VelocityMS [3]float64
	// SunPositionM is geocentric Sun position in metres.
	SunPositionM [3]float64
}

// MeanElements contains the mean elements and the number of iterations taken to obtain them.
// Values with the same element type can be compared with ==.
type MeanElements[T CircularElements | AnalyticKeplerianElements] struct {
	// Elements are the converged mean elements.
	Elements T
	// Iterations is the number of conversion iterations.
	Iterations uint32
}

// CircularElements contains analytic orbit elements.
// Compare values with ==.
type CircularElements struct {
	// AKm is semi-major axis in kilometres.
	AKm float64
	// Ex is eccentricity times cosine of argument of perigee.
	Ex float64
	// Ey is eccentricity times sine of argument of perigee.
	Ey float64
	// IRad is inclination in radians.
	IRad float64
	// RaanRad is right ascension of the ascending node in radians.
	RaanRad float64
	// MeanLatitudeArgumentRad is mean latitude argument in radians.
	MeanLatitudeArgumentRad float64
}

// CosineTurnParameters contains cosine yaw-turn configuration.
// Compare values with ==.
type CosineTurnParameters struct {
	// BetaLimitRad is absolute Sun elevation limit, radians.
	BetaLimitRad float64
	// BetaLimitInclusive is whether the elevation limit is inclusive.
	BetaLimitInclusive bool
	// StartAngle is start angle selector as SidereonCosineTurnStartAngle.
	StartAngle CosineTurnStartAngle
	// StartAngleRad is angle before noon or midnight at the start, radians.
	StartAngleRad float64
	// PeriodS is cosine period in seconds.
	PeriodS float64
}

// DECoverage contains coverage interval in absolute TDB seconds past J2000.
// Compare values with ==.
type DECoverage struct {
	// StartTDBSeconds is first covered epoch, inclusive.
	StartTDBSeconds float64
	// EndTDBSeconds is last covered epoch, inclusive.
	EndTDBSeconds float64
}

// DEGM is a DE header mass parameter that may be absent.
// Use Equal to compare values.
type DEGM struct {
	// Present is whether the header provides the parameter.
	Present bool
	// ValueM3S2 is gravitational mass parameter in m³/s², or zero when absent.
	ValueM3S2 float64
}

// Equal compares the presence and value of the mass parameter.
func (value DEGM) Equal(other DEGM) bool {
	return equalOptionalValue(value.Present, value.ValueM3S2, other.Present, other.ValueM3S2)
}

// DEState contains geocentric state in GCRF. Compare the fields by value.
// Compare values with ==.
type DEState struct {
	// PositionM is position in metres.
	PositionM [3]float64
	// VelocityMS is velocity in metres per second.
	VelocityMS [3]float64
}

// DESunMoon contains simultaneous geocentric Sun and Moon states in GCRF.
// Compare values with ==.
type DESunMoon struct {
	// Sun is sun state.
	Sun DEState
	// Moon is moon state.
	Moon DEState
}

// DYBAxes contains sun-fixed axes in the input frame.
// Compare values with ==.
type DYBAxes struct {
	// D is d unit vector.
	D [3]float64
	// Y is y unit vector.
	Y [3]float64
	// B is b unit vector.
	B [3]float64
}

// EarthOrientation contains earth orientation at one instant. Matrices use row-major order.
// Use Equal to compare values.
type EarthOrientation struct {
	// TimeScales is time scales at the instant and UT1 coverage status.
	TimeScales TimeScales
	// PolarMotionRad is polar motion X and Y in radians.
	PolarMotionRad [2]float64
	// EarthRotationVectorITRFRadS is earth spin vector in ITRF, in radians per second.
	EarthRotationVectorITRFRadS [3]float64
	// GCRFToITRFMatrix is rotation from GCRF to ITRF.
	GCRFToITRFMatrix [3][3]float64
	// ITRFToGCRFMatrix is rotation from ITRF to GCRF.
	ITRFToGCRFMatrix [3][3]float64
	// GCRFToITRFRotationRateMatrix is time derivative of the GCRF to ITRF rotation.
	GCRFToITRFRotationRateMatrix [3][3]float64
	// ITRFToGCRFRotationRateMatrix is time derivative of the ITRF to GCRF rotation.
	ITRFToGCRFRotationRateMatrix [3][3]float64
	// HasIERS2010Quaternion is whether an IERS 2010 rotation quaternion is available.
	HasIERS2010Quaternion bool
	// IERS2010Quaternion is IERS 2010 rotation quaternion in scalar-first order when available.
	IERS2010Quaternion [4]float64
}

// Equal compares Earth orientation, including the quaternion when present.
func (value EarthOrientation) Equal(other EarthOrientation) bool {
	if !value.HasIERS2010Quaternion {
		value.IERS2010Quaternion = [4]float64{}
	}
	if !other.HasIERS2010Quaternion {
		other.IERS2010Quaternion = [4]float64{}
	}
	return value == other
}

// ECOM1Coefficients contains radiation pressure coefficients in m/s².
// Compare values with ==.
type ECOM1Coefficients struct {
	// D0 is coefficient D0.
	D0 float64
	// Dc is coefficient DC.
	Dc float64
	// Ds is coefficient DS.
	Ds float64
	// Y0 is coefficient Y0.
	Y0 float64
	// Yc is coefficient YC.
	Yc float64
	// Ys is coefficient YS.
	Ys float64
	// B0 is coefficient B0.
	B0 float64
	// Bc is coefficient BC.
	Bc float64
	// Bs is coefficient BS.
	Bs float64
}

// ECOM2D2B1Coefficients contains radiation pressure coefficients in m/s².
// Compare values with ==.
type ECOM2D2B1Coefficients struct {
	// B0 is coefficient B0.
	B0 float64
	// B1c is coefficient B1C.
	B1c float64
	// B1s is coefficient B1S.
	B1s float64
	// D0 is coefficient D0.
	D0 float64
	// D2c is coefficient D2C.
	D2c float64
	// D2s is coefficient D2S.
	D2s float64
	// Y0 is coefficient Y0.
	Y0 float64
}

// ECOM2D4B1Coefficients contains radiation pressure coefficients in m/s².
// Compare values with ==.
type ECOM2D4B1Coefficients struct {
	// B0 is coefficient B0.
	B0 float64
	// B1c is coefficient B1C.
	B1c float64
	// B1s is coefficient B1S.
	B1s float64
	// D0 is coefficient D0.
	D0 float64
	// D2c is coefficient D2C.
	D2c float64
	// D4c is coefficient D4C.
	D4c float64
	// D2s is coefficient D2S.
	D2s float64
	// D4s is coefficient D4S.
	D4s float64
	// Y0 is coefficient Y0.
	Y0 float64
}

// ECOMGeometry contains sun-fixed geometry used by the ECOM radiation pressure models.
// Use Equal to compare values.
type ECOMGeometry struct {
	// DYB is sun-fixed unit vectors.
	DYB DYBAxes
	// DeltaURad is difference in latitude argument, radians, when defined.
	DeltaURad OptionalOrbitScalar
	// ArgumentOfLatitudeRad is satellite argument of latitude, radians, when defined.
	ArgumentOfLatitudeRad OptionalOrbitScalar
	// LightingRatio is fraction of incident sunlight.
	LightingRatio float64
}

// Equal compares the geometry and the available latitude angles.
func (value ECOMGeometry) Equal(other ECOMGeometry) bool {
	return value.DYB == other.DYB && value.LightingRatio == other.LightingRatio &&
		value.DeltaURad.Equal(other.DeltaURad) && value.ArgumentOfLatitudeRad.Equal(other.ArgumentOfLatitudeRad)
}

// EmpiricalRTNTerm contains one empirical RTN coefficient.
// Compare values with ==.
type EmpiricalRTNTerm struct {
	// Axis is axis as SidereonRtnAxis.
	Axis RTNAxis
	// Harmonic is harmonic as SidereonEmpiricalHarmonic.
	Harmonic EmpiricalHarmonic
	// CoefficientKmS2 is coefficient in km/s².
	CoefficientKmS2 float64
}

// EmpiricalRTNTermSelector contains empirical RTN term selected by axis and harmonic.
// Compare values with ==.
type EmpiricalRTNTermSelector struct {
	// Axis is axis as SidereonRtnAxis.
	Axis RTNAxis
	// Harmonic is harmonic as SidereonEmpiricalHarmonic.
	Harmonic EmpiricalHarmonic
}

// ErrorStats contains statistics of a range error series.
// Use Equal to compare values.
type ErrorStats struct {
	// Present is whether the series contains data.
	Present bool
	// RMSM is root mean square in metres.
	RMSM float64
	// MeanM is mean error in metres.
	MeanM float64
	// MaxAbsM is maximum absolute error in metres.
	MaxAbsM float64
}

// Equal compares the statistics when the series is present.
func (value ErrorStats) Equal(other ErrorStats) bool {
	return value.Present == other.Present && (!value.Present || value == other)
}

// EstimatedForceParameter contains a force parameter estimate. Compare values
// with == and use OrbitParameterFit.Sigma for the formal standard deviation.
type EstimatedForceParameter struct {
	// Name is the fitted force parameter name.
	Name string
	// InitialValue is configured value before fitting.
	InitialValue float64
	// Value is the fitted value.
	Value float64
}

// FrameAgreement contains frame reconciliation applied in a score.
// Use Equal to compare values.
type FrameAgreement struct {
	// Kind is frame agreement kind.
	Kind FrameAgreementKind
	// From is source frame code for Helmert agreement.
	From TerrestrialFrame
	// To is reference frame code for Helmert agreement.
	To TerrestrialFrame
}

// Equal compares the reconciliation kind and any Helmert frames.
func (value FrameAgreement) Equal(other FrameAgreement) bool {
	return value.Kind == other.Kind &&
		(value.Kind != FrameAgreementKindHelmert || (value.From == other.From && value.To == other.To))
}

// GNSSAttitude contains satellite body axes and yaw from a GNSS attitude law.
// Use Equal to compare values.
type GNSSAttitude struct {
	// Mode is yaw law mode.
	Mode AttitudeMode
	// YawRad is yaw in radians.
	YawRad float64
	// NominalYawRad is nominal yaw in radians.
	NominalYawRad float64
	// BetaRad is sun elevation above the orbit plane, radians.
	BetaRad float64
	// OrbitAngleRad is in-plane orbit angle, radians.
	OrbitAngleRad float64
	// BodyX is body X unit vector.
	BodyX [3]float64
	// BodyY is body Y unit vector.
	BodyY [3]float64
	// BodyZ is body Z unit vector.
	BodyZ [3]float64
	// HasDYB is whether the DYB frame is defined.
	HasDYB bool
	// DYB is DYB unit vectors when defined.
	DYB DYBAxes
}

// Equal compares the attitude, including the DYB axes when defined.
func (value GNSSAttitude) Equal(other GNSSAttitude) bool {
	if !value.HasDYB {
		value.DYB = DYBAxes{}
	}
	if !other.HasDYB {
		other.DYB = DYBAxes{}
	}
	return value == other
}

// HorizonBinStats contains scores for a prediction horizon bin.
// Use Equal to compare values.
type HorizonBinStats struct {
	// HorizonStartS is inclusive bin start in seconds.
	HorizonStartS float64
	// HorizonEndS is bin end in seconds.
	HorizonEndS float64
	// Count is number of orbit scores in the bin.
	Count uint64
	// Radial is radial position errors.
	Radial ErrorStats
	// Transverse is transverse position errors.
	Transverse ErrorStats
	// Normal is normal position errors.
	Normal ErrorStats
	// Orbit3d is three-dimensional position errors.
	Orbit3d ErrorStats
	// SISREOrbit is orbit-only SISRE.
	SISREOrbit ErrorStats
	// Clock is clock errors expressed in metres.
	Clock ErrorStats
	// SISRE is SISRE including clock offsets.
	SISRE ErrorStats
	// ClockCount is number of clock scores in the bin.
	ClockCount uint64
}

// Equal compares the bin limits, counts and available statistics.
func (value HorizonBinStats) Equal(other HorizonBinStats) bool {
	return value.HorizonStartS == other.HorizonStartS && value.HorizonEndS == other.HorizonEndS &&
		value.Count == other.Count && value.ClockCount == other.ClockCount &&
		value.Radial.Equal(other.Radial) && value.Transverse.Equal(other.Transverse) &&
		value.Normal.Equal(other.Normal) && value.Orbit3d.Equal(other.Orbit3d) &&
		value.SISREOrbit.Equal(other.SISREOrbit) && value.Clock.Equal(other.Clock) && value.SISRE.Equal(other.SISRE)
}

// ICGEMHeader contains the iCGEM gravity field header metadata.
// Use Equal to compare values.
type ICGEMHeader struct {
	// GravityConstantM3S2 is gravity constant in m³/s².
	GravityConstantM3S2 float64
	// RadiusM is reference radius in metres.
	RadiusM float64
	// MaxDegree is maximum degree stated in the file.
	MaxDegree uint16
	// Errors is uncertainty columns as SidereonIcgemErrors.
	Errors ICGEMErrors
	// Normalization is file normalization as SidereonIcgemNormalization.
	Normalization ICGEMNormalization
	// HasTideSystem is whether a tide system is stated.
	HasTideSystem bool
	// TideSystem is declared tide system as SidereonGravityTideSystem when present.
	TideSystem GravityTideSystem
	// Format is file format as SidereonIcgemFormat.
	Format ICGEMFormat
}

// Equal compares the header fields, including the tide system when present.
// Compare model names separately with ICGEMGravityField.ModelName.
func (value ICGEMHeader) Equal(other ICGEMHeader) bool {
	if !value.HasTideSystem {
		value.TideSystem = 0
	}
	if !other.HasTideSystem {
		other.TideSystem = 0
	}
	return value == other
}

// ICGEMSigmas contains uncertainties of one normalized coefficient pair.
// Use Equal to compare values.
type ICGEMSigmas struct {
	// HasCalibrated is whether calibrated uncertainties are present.
	HasCalibrated bool
	// Calibrated is calibrated C and S uncertainties when present.
	Calibrated [2]float64
	// HasFormal is whether formal uncertainties are present.
	HasFormal bool
	// Formal is formal C and S uncertainties when present.
	Formal [2]float64
}

// Equal compares the presence and values of calibrated and formal uncertainties.
func (value ICGEMSigmas) Equal(other ICGEMSigmas) bool {
	return equalOptionalValue(value.HasCalibrated, value.Calibrated, other.HasCalibrated, other.Calibrated) &&
		equalOptionalValue(value.HasFormal, value.Formal, other.HasFormal, other.Formal)
}

// ICGEMTimeVariableTerm contains terms of a time-dependent gravity coefficient.
// Use Equal to compare values.
type ICGEMTimeVariableTerm struct {
	// Line is one-based source line.
	Line uint64
	// Kind is record kind as SidereonIcgemTermKind.
	Kind ICGEMTermKind
	// PeriodYears is period in Julian years for sine and cosine records, zero otherwise.
	PeriodYears float64
	// Degree is harmonic degree.
	Degree uint16
	// Order is harmonic order.
	Order uint16
	// C is fully normalized C term.
	C float64
	// S is fully normalized S term.
	S float64
	// Sigmas is coefficient uncertainties.
	Sigmas ICGEMSigmas
	// ReferenceTTSeconds is reference epoch in TT seconds past J2000.
	ReferenceTTSeconds float64
	// ValidFromTTSeconds is inclusive start in TT seconds past J2000.
	ValidFromTTSeconds float64
	// ValidUntilTTSeconds is exclusive end in TT seconds past J2000.
	ValidUntilTTSeconds float64
}

// Equal compares the term, with a period for sine and cosine terms.
func (value ICGEMTimeVariableTerm) Equal(other ICGEMTimeVariableTerm) bool {
	if !value.Sigmas.Equal(other.Sigmas) {
		return false
	}
	value.Sigmas, other.Sigmas = ICGEMSigmas{}, ICGEMSigmas{}
	if value.Kind == ICGEMTermKindConstant || value.Kind == ICGEMTermKindTrend {
		value.PeriodYears, other.PeriodYears = 0, 0
	}
	return value == other
}

// IERS2010EopSample is a daily IERS Earth orientation sample.
// Use Equal to compare values.
type IERS2010EopSample struct {
	// MJDUTC is UTC modified Julian date.
	MJDUTC int32
	// XpArcsec is pole X in arcseconds.
	XpArcsec float64
	// YpArcsec is pole Y in arcseconds.
	YpArcsec float64
	// Ut1MinusUTCS is UT1 minus UTC in seconds.
	Ut1MinusUTCS float64
	// LodS is excess length of day in seconds.
	LodS float64
	// DxArcsec is celestial pole X correction in arcseconds.
	DxArcsec float64
	// DyArcsec is celestial pole Y correction in arcseconds.
	DyArcsec float64
	// HasPoleRates is whether both pole rates are supplied.
	HasPoleRates bool
	// PoleRatesRadS is pole X and Y rates in radians per second when supplied.
	PoleRatesRadS [2]float64
}

// Equal compares the sample, including the pole rates when supplied.
func (value IERS2010EopSample) Equal(other IERS2010EopSample) bool {
	if !value.HasPoleRates {
		value.PoleRatesRadS = [2]float64{}
	}
	if !other.HasPoleRates {
		other.PoleRatesRadS = [2]float64{}
	}
	return value == other
}

// IntegratorOptions contains numerical integrator and controls for state and variational propagation.
// Compare values with ==.
type IntegratorOptions struct {
	// Integrator is integrator as SidereonPropagationIntegrator.
	Integrator PropagationIntegrator
	// AbsTol is absolute error tolerance.
	AbsTol float64
	// RelTol is relative error tolerance.
	RelTol float64
	// InitialStepS is initial step in seconds.
	InitialStepS float64
	// MinStepS is minimum step in seconds.
	MinStepS float64
	// MaxStepS is maximum step in seconds.
	MaxStepS float64
	// MaxSteps is maximum integration steps.
	MaxSteps uint32
}

// MeanConversionOptions contains mean element iteration settings.
// Compare values with ==.
type MeanConversionOptions struct {
	// Tolerance is relative convergence tolerance.
	Tolerance float64
	// MaxIterations is maximum number of iterations.
	MaxIterations uint32
}

// NominalYaw contains nominal yaw and body axes.
// Compare values with ==.
type NominalYaw struct {
	// YawRad is nominal yaw angle, radians.
	YawRad float64
	// BetaRad is sun elevation above the orbit plane, radians.
	BetaRad float64
	// BodyX is body X unit vector.
	BodyX [3]float64
	// BodyY is body Y unit vector.
	BodyY [3]float64
	// BodyZ is body Z unit vector.
	BodyZ [3]float64
}

// OptionalHarmonicCoefficient contains harmonic coefficient lookup result.
// Use Equal to compare values.
type OptionalHarmonicCoefficient struct {
	// Present is whether the coefficient exists.
	Present bool
	// Coefficient is coefficient when present.
	Coefficient SphericalHarmonicCoefficient
}

// Equal compares the presence and value of the coefficient.
func (value OptionalHarmonicCoefficient) Equal(other OptionalHarmonicCoefficient) bool {
	return equalOptionalValue(value.Present, value.Coefficient, other.Present, other.Coefficient)
}

func equalOptionalValue[T comparable](present bool, value T, otherPresent bool, other T) bool {
	return present == otherPresent && (!present || value == other)
}

// OptionalOrbitScalar contains optional scalar result from an orbit operation.
// Use Equal to compare values.
type OptionalOrbitScalar struct {
	// Present is whether a value is present.
	Present bool
	// Value is value when present.
	Value float64
}

// Equal compares the presence and value of the scalar.
func (value OptionalOrbitScalar) Equal(other OptionalOrbitScalar) bool {
	return equalOptionalValue(value.Present, value.Value, other.Present, other.Value)
}

// OrbitParameterCovariance contains metadata for the scaled covariance and correlation matrices.
// Use Equal to compare values.
type OrbitParameterCovariance struct {
	// Estimated is whether estimated covariance and correlation matrices are available.
	Estimated bool
	// Dimension is number of rows and columns in each matrix, when estimated.
	Dimension uint64
	// VarianceFactor is a-posteriori variance factor, when estimated.
	VarianceFactor float64
	// DegreesOfFreedom is residual degrees of freedom.
	DegreesOfFreedom uint64
}

// Equal compares covariance metadata when estimated. Use OrbitParameterFit.Equal
// to compare fits including the covariance and correlation matrices.
func (value OrbitParameterCovariance) Equal(other OrbitParameterCovariance) bool {
	return value.Estimated == other.Estimated && (!value.Estimated || value == other)
}

// OrbitParameterFitSummary contains summary of a force-parameter fit.
// Compare values with ==.
type OrbitParameterFitSummary struct {
	// Satellite is fitted satellite.
	Satellite string
	// InitialState is estimated initial state at the first observation.
	InitialState CartesianState
	// ForceParameterCount is number of estimated force parameters.
	ForceParameterCount uint64
	// ParameterCount is number of initial-state and force parameters.
	ParameterCount uint64
	// ResidualStats is post-fit residual statistics.
	ResidualStats OrbitResidualStats
	// SeedRMS3dM is initial guess three-dimensional RMS residual, m.
	SeedRMS3dM float64
	// Iterations is iteration count.
	Iterations uint64
	// ConditionNumber is column-scaled normal equation condition number.
	ConditionNumber float64
	// Ut1Degraded is first UT1 degradation, if any.
	Ut1Degraded Ut1Degradation
}

// OrbitRTNResidual contains one post-fit RTN residual.
// Compare values with ==.
type OrbitRTNResidual struct {
	// VariationalEpoch is observation epoch in TDB and elapsed TT since the estimated initial state.
	VariationalEpoch VariationalEpoch
	// RadialM is radial residual in metres.
	RadialM float64
	// AlongM is along-track residual in metres.
	AlongM float64
	// CrossM is cross-track residual in metres.
	CrossM float64
}

// OrbitScoreSummary contains aggregate score report and row counts.
// Use Equal to compare values.
type OrbitScoreSummary struct {
	// Frame is frame reconciliation applied.
	Frame FrameAgreement
	// SourceReferencePoint is spacecraft reference point of the orbit source.
	SourceReferencePoint OrbitReferencePoint
	// SatelliteCount is number of satellite summaries.
	SatelliteCount uint64
	// SystemCount is number of constellation summaries.
	SystemCount uint64
	// Coverage is coverage over all requested satellites and epochs.
	Coverage ScoreCoverage
}

// Equal compares the frame reconciliation and aggregate report fields.
func (value OrbitScoreSummary) Equal(other OrbitScoreSummary) bool {
	return value.Frame.Equal(other.Frame) && value.SourceReferencePoint == other.SourceReferencePoint &&
		value.SatelliteCount == other.SatelliteCount && value.SystemCount == other.SystemCount && value.Coverage == other.Coverage
}

// OrbitSourceInfo contains orbit source metadata. Compare metadata fields with ==.
// Read the frame and SP3 time system labels separately.
type OrbitSourceInfo struct {
	// TimeScale is time scale as SidereonTimeScale.
	TimeScale TimeScale
	// HasSP3TimeSystem is whether the source specifically declares an SP3 time system label.
	HasSP3TimeSystem bool
	// Frame is coordinate frame kind.
	Frame OrbitFrameKind
	// ReferencePoint is spacecraft reference point.
	ReferencePoint OrbitReferencePoint
	// ClockIncludesRelativity is whether the clock includes periodic relativity.
	ClockIncludesRelativity bool
}

// OrbitSourceState contains one Earth-fixed state from an orbit source.
// Use Equal to compare values.
type OrbitSourceState struct {
	// PositionECEFM is earth-fixed position in metres.
	PositionECEFM [3]float64
	// ClockS is clock offset in seconds, when available.
	ClockS OptionalOrbitScalar
}

// Equal compares Earth-fixed position and the optional clock offset.
func (value OrbitSourceState) Equal(other OrbitSourceState) bool {
	return value.PositionECEFM == other.PositionECEFM && value.ClockS.Equal(other.ClockS)
}

// RadiationShadowOptions contains solar radiation pressure lighting configuration.
// Compare values with ==.
type RadiationShadowOptions struct {
	// ShadowFunction is lighting function as SidereonShadowFunction.
	ShadowFunction ShadowFunction
	// EarthShadowModel is earth figure as SidereonEarthShadowModel.
	EarthShadowModel EarthShadowModel
	// MoonEclipseEnabled is whether the Moon is included under apparent-disc occultation.
	MoonEclipseEnabled bool
}

// RTNComponents contains radial, transverse and normal components of a vector.
// Compare values with ==.
type RTNComponents struct {
	// RadialM is radial component in metres.
	RadialM float64
	// TransverseM is transverse component in metres.
	TransverseM float64
	// NormalM is normal component in metres.
	NormalM float64
}

// SatelliteScore contains summary for one satellite.
// Use Equal to compare values.
type SatelliteScore struct {
	// Satellite is satellite token.
	Satellite string
	// PredictionOriginJ2000S is prediction origin in the source time scale.
	PredictionOriginJ2000S OptionalOrbitScalar
	// HasOrbitRegime is whether the orbit regime is known.
	HasOrbitRegime bool
	// OrbitRegime is orbit regime when known.
	OrbitRegime OrbitRegime
	// HasSISREWeights is whether SISRE weights were determined.
	HasSISREWeights bool
	// SISREWeights is SISRE weights when determined.
	SISREWeights SISREWeights
	// BinCount is number of prediction horizon bins.
	BinCount uint64
	// Coverage is coverage for this satellite.
	Coverage ScoreCoverage
}

// Equal compares the satellite summary. Use OrbitScoreReport.Equal to compare
// reports including their bin statistics.
func (value SatelliteScore) Equal(other SatelliteScore) bool {
	return value.Satellite == other.Satellite && value.BinCount == other.BinCount && value.Coverage == other.Coverage &&
		value.PredictionOriginJ2000S.Equal(other.PredictionOriginJ2000S) &&
		equalOptionalValue(value.HasOrbitRegime, value.OrbitRegime, other.HasOrbitRegime, other.OrbitRegime) &&
		equalOptionalValue(value.HasSISREWeights, value.SISREWeights, other.HasSISREWeights, other.SISREWeights)
}

// ScoreCoverage contains counts of requested satellite and epoch pairs.
// Compare values with ==.
type ScoreCoverage struct {
	// Epochs is number of requested pairs.
	Epochs uint64
	// Scored is number of pairs scored.
	Scored uint64
	// Clocked is number of pairs with clock scores.
	Clocked uint64
	// OutsideHorizonBins is number of pairs outside the horizon bins.
	OutsideHorizonBins uint64
	// NoPredictionOrigin is number of pairs without a prediction origin.
	NoPredictionOrigin uint64
	// SourceSatelliteMissing is number of pairs whose satellite is absent from the source.
	SourceSatelliteMissing uint64
	// SourceEpochMissing is number of pairs whose epoch is absent from the source.
	SourceEpochMissing uint64
	// ReferenceSatelliteMissing is number of pairs whose satellite is absent from the reference.
	ReferenceSatelliteMissing uint64
	// ReferenceEpochMissing is number of pairs whose epoch is absent from the reference.
	ReferenceEpochMissing uint64
}

// ShadowGradient contains fraction in shadow and its derivative with respect to satellite position.
// Compare values with ==.
type ShadowGradient struct {
	// Fraction is fraction of the solar disc in shadow.
	Fraction float64
	// GradientPerKm is partial derivatives with respect to satellite coordinates, per km.
	GradientPerKm [3]float64
}

// SISREWeights contains sISRE radial weight and squared transverse/normal weight.
// Compare values with ==.
type SISREWeights struct {
	// Radial is radial weight.
	Radial float64
	// TransverseNormalSquared is squared transverse and normal weight.
	TransverseNormalSquared float64
}

// SphericalHarmonicCoefficient contains fully normalized harmonic coefficient pair.
// Compare values with ==.
type SphericalHarmonicCoefficient struct {
	// Degree is harmonic degree.
	Degree uint16
	// Order is harmonic order.
	Order uint16
	// C is normalized C coefficient.
	C float64
	// S is normalized S coefficient.
	S float64
}

// SunMoon contains analytic geocentric positions of the Sun and Moon on GCRS axes.
// Compare values with ==.
type SunMoon struct {
	// SunM is sun position in metres.
	SunM [3]float64
	// MoonM is moon position in metres.
	MoonM [3]float64
}

// SystemScore contains summary for one constellation.
// Compare values with ==.
type SystemScore struct {
	// System is constellation.
	System GNSSSystem
	// BinCount is number of prediction horizon bins.
	BinCount uint64
	// Coverage is coverage for this constellation.
	Coverage ScoreCoverage
}

// VariationalEpoch contains force evaluation epoch and integration time.
// Compare values with ==.
type VariationalEpoch struct {
	// TDBJ2000S is absolute TDB seconds past J2000 for force evaluation.
	TDBJ2000S float64
	// ElapsedTTS is elapsed TT seconds from the initial state.
	ElapsedTTS float64
}

// VariationalState contains state and state-transition matrix at one variational output epoch.
// Compare values with ==.
type VariationalState struct {
	// State is cartesian state.
	State CartesianState
	// StateTransition is state-transition matrix in row-major order.
	StateTransition [36]float64
	// SensitivityColumns is number of columns in the six-row parameter-sensitivity matrix.
	SensitivityColumns uint64
}

// ZonalHarmonics contains unnormalized zonal gravity coefficients.
// Compare values with ==.
type ZonalHarmonics struct {
	// MuKm3S2 is gravitational parameter in cubic kilometres per second squared.
	MuKm3S2 float64
	// ReKm is reference radius in kilometres.
	ReKm float64
	// C20 is unnormalized degree-2 zonal coefficient.
	C20 float64
	// C30 is unnormalized degree-3 zonal coefficient.
	C30 float64
	// C40 is unnormalized degree-4 zonal coefficient.
	C40 float64
	// C50 is unnormalized degree-5 zonal coefficient.
	C50 float64
	// C60 is unnormalized degree-6 zonal coefficient.
	C60 float64
}

// BrouwerLyddanePropagator owns a native BrouwerLyddanePropagator resource. Its methods may run concurrently with Close.
// A BrouwerLyddanePropagator must not be copied after first use.
type BrouwerLyddanePropagator struct {
	_      noCopy
	handle *native.BrouwerLyddanePropagator
}

func nativeBrouwerLyddanePropagatorHandle(v *BrouwerLyddanePropagator) *native.BrouwerLyddanePropagator {
	if v == nil {
		return nil
	}
	return v.handle
}
func publicBrouwerLyddanePropagator(handle *native.BrouwerLyddanePropagator) *BrouwerLyddanePropagator {
	return &BrouwerLyddanePropagator{handle: handle}
}

// Close releases the BrouwerLyddanePropagator resource. It is safe to call more than once.
func (v *BrouwerLyddanePropagator) Close() error {
	if handle := nativeBrouwerLyddanePropagatorHandle(v); handle != nil {
		return publicError(handle.Close())
	}
	return nil
}

// DEEphemeris owns a native DEEphemeris resource. Its methods may run concurrently with Close.
// A DEEphemeris must not be copied after first use.
type DEEphemeris struct {
	_      noCopy
	handle *native.DEEphemeris
}

func nativeDEEphemerisHandle(v *DEEphemeris) *native.DEEphemeris {
	if v == nil {
		return nil
	}
	if v.handle == nil {
		return &native.DEEphemeris{}
	}
	return v.handle
}
func publicDEEphemeris(handle *native.DEEphemeris) *DEEphemeris {
	return &DEEphemeris{handle: handle}
}

// Close releases the DEEphemeris resource. It is safe to call more than once.
func (v *DEEphemeris) Close() error {
	if handle := nativeDEEphemerisHandle(v); handle != nil {
		return publicError(handle.Close())
	}
	return nil
}

// EcksteinHechlerPropagator owns a native EcksteinHechlerPropagator resource. Its methods may run concurrently with Close.
// A EcksteinHechlerPropagator must not be copied after first use.
type EcksteinHechlerPropagator struct {
	_      noCopy
	handle *native.EcksteinHechlerPropagator
}

func nativeEcksteinHechlerPropagatorHandle(v *EcksteinHechlerPropagator) *native.EcksteinHechlerPropagator {
	if v == nil {
		return nil
	}
	return v.handle
}
func publicEcksteinHechlerPropagator(handle *native.EcksteinHechlerPropagator) *EcksteinHechlerPropagator {
	return &EcksteinHechlerPropagator{handle: handle}
}

// Close releases the EcksteinHechlerPropagator resource. It is safe to call more than once.
func (v *EcksteinHechlerPropagator) Close() error {
	if handle := nativeEcksteinHechlerPropagatorHandle(v); handle != nil {
		return publicError(handle.Close())
	}
	return nil
}

// ForceModel owns a native ForceModel resource. Its methods may run concurrently with Close.
// A ForceModel must not be copied after first use.
type ForceModel struct {
	_      noCopy
	handle *native.ForceModel
}

func nativeForceModelValue(v *ForceModel) *native.ForceModel {
	if v == nil {
		return nil
	}
	return v.handle
}
func publicForceModel(handle *native.ForceModel) *ForceModel {
	return &ForceModel{handle: handle}
}

// Close releases the ForceModel resource. It is safe to call more than once.
func (v *ForceModel) Close() error {
	if handle := nativeForceModelValue(v); handle != nil {
		return publicError(handle.Close())
	}
	return nil
}

// GNSSAttitudeLaw owns a native GNSSAttitudeLaw resource. Its methods may run concurrently with Close.
// A GNSSAttitudeLaw must not be copied after first use.
type GNSSAttitudeLaw struct {
	_      noCopy
	handle *native.GNSSAttitudeLaw
}

func nativeGNSSAttitudeLawHandle(v *GNSSAttitudeLaw) *native.GNSSAttitudeLaw {
	if v == nil {
		return nil
	}
	return v.handle
}
func publicGNSSAttitudeLaw(handle *native.GNSSAttitudeLaw) *GNSSAttitudeLaw {
	return &GNSSAttitudeLaw{handle: handle}
}

// Close releases the GNSSAttitudeLaw resource. It is safe to call more than once.
func (v *GNSSAttitudeLaw) Close() error {
	if handle := nativeGNSSAttitudeLawHandle(v); handle != nil {
		return publicError(handle.Close())
	}
	return nil
}

// ICGEMGravityField owns a native ICGEMGravityField resource. Its methods may run concurrently with Close.
// A ICGEMGravityField must not be copied after first use.
type ICGEMGravityField struct {
	_      noCopy
	handle *native.ICGEMGravityField
}

func nativeICGEMGravityFieldHandle(v *ICGEMGravityField) *native.ICGEMGravityField {
	if v == nil {
		return nil
	}
	return v.handle
}
func publicICGEMGravityField(handle *native.ICGEMGravityField) *ICGEMGravityField {
	return &ICGEMGravityField{handle: handle}
}

// Close releases the ICGEMGravityField resource. It is safe to call more than once.
func (v *ICGEMGravityField) Close() error {
	if handle := nativeICGEMGravityFieldHandle(v); handle != nil {
		return publicError(handle.Close())
	}
	return nil
}

// IERS2010EarthOrientationProvider owns a native IERS2010EarthOrientationProvider resource. Its methods may run concurrently with Close.
// A IERS2010EarthOrientationProvider must not be copied after first use.
type IERS2010EarthOrientationProvider struct {
	_      noCopy
	handle *native.IERS2010EarthOrientationProvider
}

func nativeIERS2010EarthOrientationProviderHandle(v *IERS2010EarthOrientationProvider) *native.IERS2010EarthOrientationProvider {
	if v == nil {
		return nil
	}
	if v.handle == nil {
		return &native.IERS2010EarthOrientationProvider{}
	}
	return v.handle
}
func publicIERS2010EarthOrientationProvider(handle *native.IERS2010EarthOrientationProvider) *IERS2010EarthOrientationProvider {
	return &IERS2010EarthOrientationProvider{handle: handle}
}

// Close releases the IERS2010EarthOrientationProvider resource. It is safe to call more than once.
func (v *IERS2010EarthOrientationProvider) Close() error {
	if handle := nativeIERS2010EarthOrientationProviderHandle(v); handle != nil {
		return publicError(handle.Close())
	}
	return nil
}

// OrbitParameterFit owns a native OrbitParameterFit resource. Its methods may run concurrently with Close.
// A OrbitParameterFit must not be copied after first use.
type OrbitParameterFit struct {
	_      noCopy
	handle *native.OrbitParameterFit
}

func nativeOrbitParameterFitHandle(v *OrbitParameterFit) *native.OrbitParameterFit {
	if v == nil {
		return nil
	}
	return v.handle
}
func publicOrbitParameterFit(handle *native.OrbitParameterFit) *OrbitParameterFit {
	return &OrbitParameterFit{handle: handle}
}

// Close releases the OrbitParameterFit resource. It is safe to call more than once.
func (v *OrbitParameterFit) Close() error {
	if handle := nativeOrbitParameterFitHandle(v); handle != nil {
		return publicError(handle.Close())
	}
	return nil
}

// OrbitScoreReport owns a native OrbitScoreReport resource. Its methods may run concurrently with Close.
// A OrbitScoreReport must not be copied after first use.
type OrbitScoreReport struct {
	_      noCopy
	handle *native.OrbitScoreReport
}

func nativeOrbitScoreReportHandle(v *OrbitScoreReport) *native.OrbitScoreReport {
	if v == nil {
		return nil
	}
	return v.handle
}
func publicOrbitScoreReport(handle *native.OrbitScoreReport) *OrbitScoreReport {
	return &OrbitScoreReport{handle: handle}
}

// Close releases the OrbitScoreReport resource. It is safe to call more than once.
func (v *OrbitScoreReport) Close() error {
	if handle := nativeOrbitScoreReportHandle(v); handle != nil {
		return publicError(handle.Close())
	}
	return nil
}

// OrbitSource supplies Earth-fixed states evaluated by orbit scoring.
type OrbitSource interface {
	// Close releases the orbit source.
	Close() error
	// Info describes the source time scale, frame and clock.
	Info() (OrbitSourceInfo, error)
	// FrameLabel returns the frame label when the source uses a named frame.
	FrameLabel() (string, error)
	// Frame returns the terrestrial reference frame of the source.
	Frame() (OrbitFrame, error)
	// TimeScale returns the time scale used to evaluate source states.
	TimeScale() (TimeScale, error)
	// ReferencePoint returns the spacecraft reference point of source positions.
	ReferencePoint() (OrbitReferencePoint, error)
	// ClockIncludesRelativity reports whether the clock includes periodic relativity.
	ClockIncludesRelativity() (bool, error)
	// SP3TimeSystem returns the SP3 time system label when present.
	SP3TimeSystem() (string, error)
	// PredictionOriginJ2000S returns the satellite prediction origin in the source time scale.
	PredictionOriginJ2000S(satellite string) (OptionalOrbitScalar, error)
	// StateAtJ2000S interpolates the satellite state at the given epoch in the source time scale.
	StateAtJ2000S(satellite string, epochJ2000S float64) (OrbitSourceState, error)
	orbitSource() *orbitSourceHandle
}

type orbitSourceHandle struct{ handle *native.OrbitSource }

func (v *orbitSourceHandle) orbitSource() *orbitSourceHandle { return v }
func nativeOrbitSourceHandle(v OrbitSource) *native.OrbitSource {
	if v == nil {
		return nil
	}
	return nativeOrbitSourceValue(v.orbitSource())
}
func nativeOrbitSourceValue(v *orbitSourceHandle) *native.OrbitSource {
	if v == nil {
		return nil
	}
	return v.handle
}

// Close releases the orbit source. It is safe to call more than once.
func (v *orbitSourceHandle) Close() error {
	if handle := nativeOrbitSourceValue(v); handle != nil {
		return publicError(handle.Close())
	}
	return nil
}

// SphericalHarmonicGravity owns a native SphericalHarmonicGravity resource. Its methods may run concurrently with Close.
// A SphericalHarmonicGravity must not be copied after first use.
type SphericalHarmonicGravity struct {
	*ForceModel
	_      noCopy
	handle *native.SphericalHarmonicGravity
}

func nativeSphericalHarmonicGravityHandle(v *SphericalHarmonicGravity) *native.SphericalHarmonicGravity {
	if v == nil {
		return nil
	}
	return v.handle
}
func publicSphericalHarmonicGravity(handle *native.SphericalHarmonicGravity) (*SphericalHarmonicGravity, error) {
	force, err := native.SphericalHarmonicGravityAsForceModel(handle)
	if err != nil {
		return nil, publicError(errors.Join(err, handle.Close()))
	}
	return &SphericalHarmonicGravity{handle: handle, ForceModel: publicForceModel(force)}, nil
}

func (gravity *SphericalHarmonicGravity) forceModel() *ForceModel {
	if gravity == nil {
		return nil
	}
	return gravity.ForceModel
}

// Close releases the spherical harmonic gravity model. It is safe to call more than once.
func (gravity *SphericalHarmonicGravity) Close() error {
	if gravity == nil {
		return nil
	}
	return errors.Join(gravity.ForceModel.Close(), publicError(gravity.handle.Close()))
}

// VariationalTrajectory owns a native VariationalTrajectory resource. Its methods may run concurrently with Close.
// A VariationalTrajectory must not be copied after first use.
type VariationalTrajectory struct {
	_      noCopy
	handle *native.VariationalTrajectory
}

func nativeVariationalTrajectoryHandle(v *VariationalTrajectory) *native.VariationalTrajectory {
	if v == nil {
		return nil
	}
	return v.handle
}
func publicVariationalTrajectory(handle *native.VariationalTrajectory) *VariationalTrajectory {
	return &VariationalTrajectory{handle: handle}
}

// Close releases the VariationalTrajectory resource. It is safe to call more than once.
func (v *VariationalTrajectory) Close() error {
	if handle := nativeVariationalTrajectoryHandle(v); handle != nil {
		return publicError(handle.Close())
	}
	return nil
}

// ECOM1SolarRadiationPressure supplies a force to numerical propagation.
type ECOM1SolarRadiationPressure struct{ *ForceModel }

func (v *ECOM1SolarRadiationPressure) forceModel() *ForceModel {
	if v == nil {
		return nil
	}
	return v.ForceModel
}
func nativeECOM1SolarRadiationPressureHandle(v *ECOM1SolarRadiationPressure) *native.ForceModel {
	return nativeForceModelValue(v.forceModel())
}
func publicECOM1SolarRadiationPressure(handle *native.ForceModel) *ECOM1SolarRadiationPressure {
	return &ECOM1SolarRadiationPressure{ForceModel: publicForceModel(handle)}
}

// ECOM2SolarRadiationPressure supplies a force to numerical propagation.
type ECOM2SolarRadiationPressure struct{ *ForceModel }

func (v *ECOM2SolarRadiationPressure) forceModel() *ForceModel {
	if v == nil {
		return nil
	}
	return v.ForceModel
}
func nativeECOM2SolarRadiationPressureHandle(v *ECOM2SolarRadiationPressure) *native.ForceModel {
	return nativeForceModelValue(v.forceModel())
}
func publicECOM2SolarRadiationPressure(handle *native.ForceModel) *ECOM2SolarRadiationPressure {
	return &ECOM2SolarRadiationPressure{ForceModel: publicForceModel(handle)}
}

// EmpiricalRTNAcceleration supplies a force to numerical propagation.
type EmpiricalRTNAcceleration struct{ *ForceModel }

func (v *EmpiricalRTNAcceleration) forceModel() *ForceModel {
	if v == nil {
		return nil
	}
	return v.ForceModel
}
func nativeEmpiricalRTNAccelerationHandle(v *EmpiricalRTNAcceleration) *native.ForceModel {
	return nativeForceModelValue(v.forceModel())
}
func publicEmpiricalRTNAcceleration(handle *native.ForceModel) *EmpiricalRTNAcceleration {
	return &EmpiricalRTNAcceleration{ForceModel: publicForceModel(handle)}
}

// EarthRadiationPressure supplies a force to numerical propagation.
type EarthRadiationPressure struct{ *ForceModel }

func (v *EarthRadiationPressure) forceModel() *ForceModel {
	if v == nil {
		return nil
	}
	return v.ForceModel
}
func nativeEarthRadiationPressureHandle(v *EarthRadiationPressure) *native.ForceModel {
	return nativeForceModelValue(v.forceModel())
}
func publicEarthRadiationPressure(handle *native.ForceModel) *EarthRadiationPressure {
	return &EarthRadiationPressure{ForceModel: publicForceModel(handle)}
}

// TwoBodyGravity supplies a force to numerical propagation.
type TwoBodyGravity struct{ *ForceModel }

func (v *TwoBodyGravity) forceModel() *ForceModel {
	if v == nil {
		return nil
	}
	return v.ForceModel
}
func publicTwoBodyGravity(handle *native.ForceModel) *TwoBodyGravity {
	return &TwoBodyGravity{ForceModel: publicForceModel(handle)}
}

// J2Gravity supplies a force to numerical propagation.
type J2Gravity struct{ *ForceModel }

func (v *J2Gravity) forceModel() *ForceModel {
	if v == nil {
		return nil
	}
	return v.ForceModel
}
func publicJ2Gravity(handle *native.ForceModel) *J2Gravity {
	return &J2Gravity{ForceModel: publicForceModel(handle)}
}

// ZonalGravity supplies a force to numerical propagation.
type ZonalGravity struct{ *ForceModel }

func (v *ZonalGravity) forceModel() *ForceModel {
	if v == nil {
		return nil
	}
	return v.ForceModel
}
func publicZonalGravity(handle *native.ForceModel) *ZonalGravity {
	return &ZonalGravity{ForceModel: publicForceModel(handle)}
}

// DragForce supplies a force to numerical propagation.
type DragForce struct{ *ForceModel }

func (v *DragForce) forceModel() *ForceModel {
	if v == nil {
		return nil
	}
	return v.ForceModel
}
func publicDragForce(handle *native.ForceModel) *DragForce {
	return &DragForce{ForceModel: publicForceModel(handle)}
}

// SourcedDragForce supplies a force to numerical propagation.
type SourcedDragForce struct{ *ForceModel }

func (v *SourcedDragForce) forceModel() *ForceModel {
	if v == nil {
		return nil
	}
	return v.ForceModel
}
func publicSourcedDragForce(handle *native.ForceModel) *SourcedDragForce {
	return &SourcedDragForce{ForceModel: publicForceModel(handle)}
}

// ThirdBodyGravity supplies a force to numerical propagation.
type ThirdBodyGravity struct{ *ForceModel }

func (v *ThirdBodyGravity) forceModel() *ForceModel {
	if v == nil {
		return nil
	}
	return v.ForceModel
}
func publicThirdBodyGravity(handle *native.ForceModel) *ThirdBodyGravity {
	return &ThirdBodyGravity{ForceModel: publicForceModel(handle)}
}

// SolidEarthTideGravity supplies a force to numerical propagation.
type SolidEarthTideGravity struct{ *ForceModel }

func (v *SolidEarthTideGravity) forceModel() *ForceModel {
	if v == nil {
		return nil
	}
	return v.ForceModel
}
func publicSolidEarthTideGravity(handle *native.ForceModel) *SolidEarthTideGravity {
	return &SolidEarthTideGravity{ForceModel: publicForceModel(handle)}
}

// SolidEarthPoleTideGravity supplies a force to numerical propagation.
type SolidEarthPoleTideGravity struct{ *ForceModel }

func (v *SolidEarthPoleTideGravity) forceModel() *ForceModel {
	if v == nil {
		return nil
	}
	return v.ForceModel
}
func publicSolidEarthPoleTideGravity(handle *native.ForceModel) *SolidEarthPoleTideGravity {
	return &SolidEarthPoleTideGravity{ForceModel: publicForceModel(handle)}
}

// SchwarzschildRelativity supplies a force to numerical propagation.
type SchwarzschildRelativity struct{ *ForceModel }

func (v *SchwarzschildRelativity) forceModel() *ForceModel {
	if v == nil {
		return nil
	}
	return v.ForceModel
}
func publicSchwarzschildRelativity(handle *native.ForceModel) *SchwarzschildRelativity {
	return &SchwarzschildRelativity{ForceModel: publicForceModel(handle)}
}

// CompositeForceModel supplies a force to numerical propagation.
type CompositeForceModel struct{ *ForceModel }

func (v *CompositeForceModel) forceModel() *ForceModel {
	if v == nil {
		return nil
	}
	return v.ForceModel
}
func publicCompositeForceModel(handle *native.ForceModel) *CompositeForceModel {
	return &CompositeForceModel{ForceModel: publicForceModel(handle)}
}

// SolarRadiationPressureForce supplies a force to numerical propagation.
type SolarRadiationPressureForce struct{ *ForceModel }

func (v *SolarRadiationPressureForce) forceModel() *ForceModel {
	if v == nil {
		return nil
	}
	return v.ForceModel
}
func nativeSolarRadiationPressureForceHandle(v *SolarRadiationPressureForce) *native.ForceModel {
	return nativeForceModelValue(v.forceModel())
}
func publicSolarRadiationPressureForce(handle *native.ForceModel) *SolarRadiationPressureForce {
	return &SolarRadiationPressureForce{ForceModel: publicForceModel(handle)}
}

// PredictedSP3 supplies Earth-fixed states for orbit scoring.
type PredictedSP3 struct{ *orbitSourceHandle }

func (v *PredictedSP3) orbitSource() *orbitSourceHandle {
	if v == nil {
		return nil
	}
	return v.orbitSourceHandle
}
func nativePredictedSP3Handle(v *PredictedSP3) *native.OrbitSource {
	return nativeOrbitSourceValue(v.orbitSource())
}
func publicPredictedSP3(handle *native.OrbitSource) *PredictedSP3 {
	return &PredictedSP3{orbitSourceHandle: &orbitSourceHandle{handle: handle}}
}

// BroadcastOrbit supplies Earth-fixed states for orbit scoring.
type BroadcastOrbit struct{ *orbitSourceHandle }

func (v *BroadcastOrbit) orbitSource() *orbitSourceHandle {
	if v == nil {
		return nil
	}
	return v.orbitSourceHandle
}
func publicBroadcastOrbit(handle *native.OrbitSource) *BroadcastOrbit {
	return &BroadcastOrbit{orbitSourceHandle: &orbitSourceHandle{handle: handle}}
}

// OrbitTrajectory supplies Earth-fixed states for orbit scoring.
type OrbitTrajectory struct{ *orbitSourceHandle }

func (v *OrbitTrajectory) orbitSource() *orbitSourceHandle {
	if v == nil {
		return nil
	}
	return v.orbitSourceHandle
}
func nativeOrbitTrajectoryHandle(v *OrbitTrajectory) *native.OrbitSource {
	return nativeOrbitSourceValue(v.orbitSource())
}
func publicOrbitTrajectory(handle *native.OrbitSource) *OrbitTrajectory {
	return &OrbitTrajectory{orbitSourceHandle: &orbitSourceHandle{handle: handle}}
}

func publicAccelerationPartials(value native.AccelerationPartials) AccelerationPartials {
	return AccelerationPartials{
		WrtPosition: value.WrtPosition,
		WrtVelocity: value.WrtVelocity,
	}
}

func publicAnalyticCartesian(value native.AnalyticCartesian) AnalyticCartesian {
	return AnalyticCartesian{
		PositionKm:  value.PositionKm,
		VelocityKmS: value.VelocityKmS,
	}
}

func nativeAnalyticCartesian(value AnalyticCartesian) native.AnalyticCartesian {
	return native.AnalyticCartesian{
		PositionKm:  value.PositionKm,
		VelocityKmS: value.VelocityKmS,
	}
}

func publicAnalyticKeplerianElements(value native.AnalyticKeplerianElements) AnalyticKeplerianElements {
	return AnalyticKeplerianElements{
		AKm:            value.AKm,
		E:              value.E,
		IRad:           value.IRad,
		ArgPerigeeRad:  value.ArgPerigeeRad,
		RaanRad:        value.RaanRad,
		MeanAnomalyRad: value.MeanAnomalyRad,
	}
}

func nativeAnalyticKeplerianElements(value AnalyticKeplerianElements) native.AnalyticKeplerianElements {
	return native.AnalyticKeplerianElements{
		AKm:            value.AKm,
		E:              value.E,
		IRad:           value.IRad,
		ArgPerigeeRad:  value.ArgPerigeeRad,
		RaanRad:        value.RaanRad,
		MeanAnomalyRad: value.MeanAnomalyRad,
	}
}

func nativeAttitudeSample(value AttitudeSample) native.AttitudeSample {
	return native.AttitudeSample{
		PositionM:    value.PositionM,
		VelocityMS:   value.VelocityMS,
		SunPositionM: value.SunPositionM,
	}
}

func publicBrouwerLyddaneMeanElements(value native.BrouwerLyddaneMeanElements) MeanElements[AnalyticKeplerianElements] {
	return MeanElements[AnalyticKeplerianElements]{
		Elements:   publicAnalyticKeplerianElements(value.Elements),
		Iterations: value.Iterations,
	}
}

func publicCircularElements(value native.CircularElements) CircularElements {
	return CircularElements{
		AKm:                     value.AKm,
		Ex:                      value.Ex,
		Ey:                      value.Ey,
		IRad:                    value.IRad,
		RaanRad:                 value.RaanRad,
		MeanLatitudeArgumentRad: value.MeanLatitudeArgumentRad,
	}
}

func nativeCircularElements(value CircularElements) native.CircularElements {
	return native.CircularElements{
		AKm:                     value.AKm,
		Ex:                      value.Ex,
		Ey:                      value.Ey,
		IRad:                    value.IRad,
		RaanRad:                 value.RaanRad,
		MeanLatitudeArgumentRad: value.MeanLatitudeArgumentRad,
	}
}

func publicCosineTurnParameters(value native.CosineTurnParameters) CosineTurnParameters {
	return CosineTurnParameters{
		BetaLimitRad:       value.BetaLimitRad,
		BetaLimitInclusive: value.BetaLimitInclusive,
		StartAngle:         CosineTurnStartAngle(value.StartAngle),
		StartAngleRad:      value.StartAngleRad,
		PeriodS:            value.PeriodS,
	}
}

func nativeCosineTurnParameters(value CosineTurnParameters) native.CosineTurnParameters {
	return native.CosineTurnParameters{
		BetaLimitRad:       value.BetaLimitRad,
		BetaLimitInclusive: value.BetaLimitInclusive,
		StartAngle:         native.CosineTurnStartAngle(value.StartAngle),
		StartAngleRad:      value.StartAngleRad,
		PeriodS:            value.PeriodS,
	}
}

func publicDECoverage(value native.DECoverage) DECoverage {
	return DECoverage{
		StartTDBSeconds: value.StartTDBSeconds,
		EndTDBSeconds:   value.EndTDBSeconds,
	}
}

func publicDEGM(value native.DEGM) DEGM {
	return DEGM{
		Present:   value.Present,
		ValueM3S2: value.ValueM3S2,
	}
}

func publicDESunMoon(value native.DESunMoon) DESunMoon {
	return DESunMoon{
		Sun:  publicDEState(value.Sun),
		Moon: publicDEState(value.Moon),
	}
}

func publicDEState(value native.DEState) DEState {
	return DEState{
		PositionM:  value.PositionM,
		VelocityMS: value.VelocityMS,
	}
}

func publicDYBAxes(value native.DYBAxes) DYBAxes {
	return DYBAxes{
		D: value.D,
		Y: value.Y,
		B: value.B,
	}
}

func publicEarthOrientation(value native.EarthOrientation) EarthOrientation {
	return EarthOrientation{
		TimeScales:                   publicTimeScales(value.TimeScales),
		PolarMotionRad:               value.PolarMotionRad,
		EarthRotationVectorITRFRadS:  value.EarthRotationVectorITRFRadS,
		GCRFToITRFMatrix:             value.GCRFToITRFMatrix,
		ITRFToGCRFMatrix:             value.ITRFToGCRFMatrix,
		GCRFToITRFRotationRateMatrix: value.GCRFToITRFRotationRateMatrix,
		ITRFToGCRFRotationRateMatrix: value.ITRFToGCRFRotationRateMatrix,
		HasIERS2010Quaternion:        value.HasIERS2010Quaternion,
		IERS2010Quaternion:           value.IERS2010Quaternion,
	}
}

func publicEcksteinHechlerMeanElements(value native.EcksteinHechlerMeanElements) MeanElements[CircularElements] {
	return MeanElements[CircularElements]{
		Elements:   publicCircularElements(value.Elements),
		Iterations: value.Iterations,
	}
}

func publicECOM1Coefficients(value native.ECOM1Coefficients) ECOM1Coefficients {
	return ECOM1Coefficients{
		D0: value.D0,
		Dc: value.Dc,
		Ds: value.Ds,
		Y0: value.Y0,
		Yc: value.Yc,
		Ys: value.Ys,
		B0: value.B0,
		Bc: value.Bc,
		Bs: value.Bs,
	}
}

func nativeECOM1Coefficients(value ECOM1Coefficients) native.ECOM1Coefficients {
	return native.ECOM1Coefficients{
		D0: value.D0,
		Dc: value.Dc,
		Ds: value.Ds,
		Y0: value.Y0,
		Yc: value.Yc,
		Ys: value.Ys,
		B0: value.B0,
		Bc: value.Bc,
		Bs: value.Bs,
	}
}

func publicECOM2D2B1Coefficients(value native.ECOM2D2B1Coefficients) ECOM2D2B1Coefficients {
	return ECOM2D2B1Coefficients{
		B0:  value.B0,
		B1c: value.B1c,
		B1s: value.B1s,
		D0:  value.D0,
		D2c: value.D2c,
		D2s: value.D2s,
		Y0:  value.Y0,
	}
}

func nativeECOM2D2B1Coefficients(value ECOM2D2B1Coefficients) native.ECOM2D2B1Coefficients {
	return native.ECOM2D2B1Coefficients{
		B0:  value.B0,
		B1c: value.B1c,
		B1s: value.B1s,
		D0:  value.D0,
		D2c: value.D2c,
		D2s: value.D2s,
		Y0:  value.Y0,
	}
}

func publicECOM2D4B1Coefficients(value native.ECOM2D4B1Coefficients) ECOM2D4B1Coefficients {
	return ECOM2D4B1Coefficients{
		B0:  value.B0,
		B1c: value.B1c,
		B1s: value.B1s,
		D0:  value.D0,
		D2c: value.D2c,
		D4c: value.D4c,
		D2s: value.D2s,
		D4s: value.D4s,
		Y0:  value.Y0,
	}
}

func nativeECOM2D4B1Coefficients(value ECOM2D4B1Coefficients) native.ECOM2D4B1Coefficients {
	return native.ECOM2D4B1Coefficients{
		B0:  value.B0,
		B1c: value.B1c,
		B1s: value.B1s,
		D0:  value.D0,
		D2c: value.D2c,
		D4c: value.D4c,
		D2s: value.D2s,
		D4s: value.D4s,
		Y0:  value.Y0,
	}
}

func publicECOMGeometry(value native.ECOMGeometry) ECOMGeometry {
	return ECOMGeometry{
		DYB:                   publicDYBAxes(value.DYB),
		DeltaURad:             publicOptionalOrbitScalar(value.DeltaURad),
		ArgumentOfLatitudeRad: publicOptionalOrbitScalar(value.ArgumentOfLatitudeRad),
		LightingRatio:         value.LightingRatio,
	}
}

func publicEmpiricalRTNTerm(value native.EmpiricalRTNTerm) EmpiricalRTNTerm {
	return EmpiricalRTNTerm{
		Axis:            RTNAxis(value.Axis),
		Harmonic:        EmpiricalHarmonic(value.Harmonic),
		CoefficientKmS2: value.CoefficientKmS2,
	}
}

func nativeEmpiricalRTNTerm(value EmpiricalRTNTerm) native.EmpiricalRTNTerm {
	return native.EmpiricalRTNTerm{
		Axis:            native.RTNAxis(value.Axis),
		Harmonic:        native.EmpiricalHarmonic(value.Harmonic),
		CoefficientKmS2: value.CoefficientKmS2,
	}
}

func nativeEmpiricalRTNTermSelector(value EmpiricalRTNTermSelector) native.EmpiricalRTNTermSelector {
	return native.EmpiricalRTNTermSelector{
		Axis:     native.RTNAxis(value.Axis),
		Harmonic: native.EmpiricalHarmonic(value.Harmonic),
	}
}

func publicEstimatedForceParameter(value native.EstimatedForceParameter) EstimatedForceParameter {
	return EstimatedForceParameter{
		Name:         value.Name,
		InitialValue: value.InitialValue,
		Value:        value.Value,
	}
}

func publicHorizonBinStats(value native.HorizonBinStats) HorizonBinStats {
	return HorizonBinStats{
		HorizonStartS: value.HorizonStartS,
		HorizonEndS:   value.HorizonEndS,
		Count:         value.Count,
		Radial:        publicErrorStats(value.Radial),
		Transverse:    publicErrorStats(value.Transverse),
		Normal:        publicErrorStats(value.Normal),
		Orbit3d:       publicErrorStats(value.Orbit3d),
		SISREOrbit:    publicErrorStats(value.SISREOrbit),
		Clock:         publicErrorStats(value.Clock),
		SISRE:         publicErrorStats(value.SISRE),
		ClockCount:    value.ClockCount,
	}
}

func publicErrorStats(value native.ErrorStats) ErrorStats {
	return ErrorStats{
		Present: value.Present,
		RMSM:    value.RMSM,
		MeanM:   value.MeanM,
		MaxAbsM: value.MaxAbsM,
	}
}

func publicICGEMHeader(value native.ICGEMHeader) ICGEMHeader {
	return ICGEMHeader{
		GravityConstantM3S2: value.GravityConstantM3S2,
		RadiusM:             value.RadiusM,
		MaxDegree:           value.MaxDegree,
		Errors:              ICGEMErrors(value.Errors),
		Normalization:       ICGEMNormalization(value.Normalization),
		HasTideSystem:       value.HasTideSystem,
		TideSystem:          GravityTideSystem(value.TideSystem),
		Format:              ICGEMFormat(value.Format),
	}
}

func publicICGEMSigmas(value native.ICGEMSigmas) ICGEMSigmas {
	return ICGEMSigmas{
		HasCalibrated: value.HasCalibrated,
		Calibrated:    value.Calibrated,
		HasFormal:     value.HasFormal,
		Formal:        value.Formal,
	}
}

func publicICGEMTimeVariableTerm(value native.ICGEMTimeVariableTerm) ICGEMTimeVariableTerm {
	return ICGEMTimeVariableTerm{
		Line:                value.Line,
		Kind:                ICGEMTermKind(value.Kind),
		PeriodYears:         value.PeriodYears,
		Degree:              value.Degree,
		Order:               value.Order,
		C:                   value.C,
		S:                   value.S,
		Sigmas:              publicICGEMSigmas(value.Sigmas),
		ReferenceTTSeconds:  value.ReferenceTTSeconds,
		ValidFromTTSeconds:  value.ValidFromTTSeconds,
		ValidUntilTTSeconds: value.ValidUntilTTSeconds,
	}
}

func nativeIERS2010EopSample(value IERS2010EopSample) native.IERS2010EopSample {
	return native.IERS2010EopSample{
		MJDUTC:        value.MJDUTC,
		XpArcsec:      value.XpArcsec,
		YpArcsec:      value.YpArcsec,
		Ut1MinusUTCS:  value.Ut1MinusUTCS,
		LodS:          value.LodS,
		DxArcsec:      value.DxArcsec,
		DyArcsec:      value.DyArcsec,
		HasPoleRates:  value.HasPoleRates,
		PoleRatesRadS: value.PoleRatesRadS,
	}
}

func publicIntegratorOptions(value native.IntegratorOptions) IntegratorOptions {
	return IntegratorOptions{
		Integrator:   PropagationIntegrator(value.Integrator),
		AbsTol:       value.AbsTol,
		RelTol:       value.RelTol,
		InitialStepS: value.InitialStepS,
		MinStepS:     value.MinStepS,
		MaxStepS:     value.MaxStepS,
		MaxSteps:     value.MaxSteps,
	}
}

func nativeIntegratorOptions(value IntegratorOptions) native.IntegratorOptions {
	return native.IntegratorOptions{
		Integrator:   uint32(value.Integrator),
		AbsTol:       value.AbsTol,
		RelTol:       value.RelTol,
		InitialStepS: value.InitialStepS,
		MinStepS:     value.MinStepS,
		MaxStepS:     value.MaxStepS,
		MaxSteps:     value.MaxSteps,
	}
}

func publicMeanConversionOptions(value native.MeanConversionOptions) MeanConversionOptions {
	return MeanConversionOptions{
		Tolerance:     value.Tolerance,
		MaxIterations: value.MaxIterations,
	}
}

func nativeMeanConversionOptions(value MeanConversionOptions) native.MeanConversionOptions {
	return native.MeanConversionOptions{
		Tolerance:     value.Tolerance,
		MaxIterations: value.MaxIterations,
	}
}

func publicNominalYaw(value native.NominalYaw) NominalYaw {
	return NominalYaw{
		YawRad:  value.YawRad,
		BetaRad: value.BetaRad,
		BodyX:   value.BodyX,
		BodyY:   value.BodyY,
		BodyZ:   value.BodyZ,
	}
}

func publicOptionalHarmonicCoefficient(value native.OptionalHarmonicCoefficient) OptionalHarmonicCoefficient {
	return OptionalHarmonicCoefficient{
		Present:     value.Present,
		Coefficient: publicSphericalHarmonicCoefficient(value.Coefficient),
	}
}

func publicOptionalOrbitScalar(value native.OptionalOrbitScalar) OptionalOrbitScalar {
	return OptionalOrbitScalar{
		Present: value.Present,
		Value:   value.Value,
	}
}

func publicOrbitParameterCovariance(value native.OrbitParameterCovariance) OrbitParameterCovariance {
	return OrbitParameterCovariance{
		Estimated:        value.Estimated,
		Dimension:        value.Dimension,
		VarianceFactor:   value.VarianceFactor,
		DegreesOfFreedom: value.DegreesOfFreedom,
	}
}

func publicOrbitParameterFitSummary(value native.OrbitParameterFitSummary) OrbitParameterFitSummary {
	return OrbitParameterFitSummary{
		Satellite:           value.Satellite,
		InitialState:        publicCartesian(value.InitialState),
		ForceParameterCount: value.ForceParameterCount,
		ParameterCount:      value.ParameterCount,
		ResidualStats:       publicOrbitResidualStats(value.ResidualStats),
		SeedRMS3dM:          value.SeedRMS3dM,
		Iterations:          value.Iterations,
		ConditionNumber:     value.ConditionNumber,
		Ut1Degraded:         Ut1Degradation(value.Ut1Degraded),
	}
}

func publicOrbitRTNResidual(value native.OrbitRTNResidual) OrbitRTNResidual {
	return OrbitRTNResidual{
		VariationalEpoch: publicVariationalEpoch(value.VariationalEpoch),
		RadialM:          value.RadialM,
		AlongM:           value.AlongM,
		CrossM:           value.CrossM,
	}
}

func publicOrbitScoreSummary(value native.OrbitScoreSummary) OrbitScoreSummary {
	return OrbitScoreSummary{
		Frame:                publicFrameAgreement(value.Frame),
		SourceReferencePoint: OrbitReferencePoint(value.SourceReferencePoint),
		SatelliteCount:       value.SatelliteCount,
		SystemCount:          value.SystemCount,
		Coverage:             publicScoreCoverage(value.Coverage),
	}
}

func publicFrameAgreement(value native.FrameAgreement) FrameAgreement {
	return FrameAgreement{
		Kind: FrameAgreementKind(value.Kind),
		From: TerrestrialFrame(value.From),
		To:   TerrestrialFrame(value.To),
	}
}

func publicOrbitSourceInfo(value native.OrbitSourceInfo) OrbitSourceInfo {
	return OrbitSourceInfo{
		TimeScale:               TimeScale(value.TimeScale),
		HasSP3TimeSystem:        value.HasSP3TimeSystem,
		Frame:                   OrbitFrameKind(value.Frame),
		ReferencePoint:          OrbitReferencePoint(value.ReferencePoint),
		ClockIncludesRelativity: value.ClockIncludesRelativity,
	}
}

func publicOrbitSourceState(value native.OrbitSourceState) OrbitSourceState {
	return OrbitSourceState{
		PositionECEFM: value.PositionECEFM,
		ClockS:        publicOptionalOrbitScalar(value.ClockS),
	}
}

func publicRadiationShadowOptions(value native.RadiationShadowOptions) RadiationShadowOptions {
	return RadiationShadowOptions{
		ShadowFunction:     ShadowFunction(value.ShadowFunction),
		EarthShadowModel:   EarthShadowModel(value.EarthShadowModel),
		MoonEclipseEnabled: value.MoonEclipseEnabled,
	}
}

func nativeRadiationShadowOptions(value RadiationShadowOptions) native.RadiationShadowOptions {
	return native.RadiationShadowOptions{
		ShadowFunction:     native.ShadowFunction(value.ShadowFunction),
		EarthShadowModel:   uint32(value.EarthShadowModel),
		MoonEclipseEnabled: value.MoonEclipseEnabled,
	}
}

func publicRTNComponents(value native.RTNComponents) RTNComponents {
	return RTNComponents{
		RadialM:     value.RadialM,
		TransverseM: value.TransverseM,
		NormalM:     value.NormalM,
	}
}

func nativeRTNComponents(value RTNComponents) native.RTNComponents {
	return native.RTNComponents{
		RadialM:     value.RadialM,
		TransverseM: value.TransverseM,
		NormalM:     value.NormalM,
	}
}

func publicSatelliteScore(value native.SatelliteScore) SatelliteScore {
	return SatelliteScore{
		Satellite:              value.Satellite,
		PredictionOriginJ2000S: publicOptionalOrbitScalar(value.PredictionOriginJ2000S),
		HasOrbitRegime:         value.HasOrbitRegime,
		OrbitRegime:            OrbitRegime(value.OrbitRegime),
		HasSISREWeights:        value.HasSISREWeights,
		SISREWeights:           publicSISREWeights(value.SISREWeights),
		BinCount:               value.BinCount,
		Coverage:               publicScoreCoverage(value.Coverage),
	}
}

func publicScoreCoverage(value native.ScoreCoverage) ScoreCoverage {
	return ScoreCoverage{
		Epochs:                    value.Epochs,
		Scored:                    value.Scored,
		Clocked:                   value.Clocked,
		OutsideHorizonBins:        value.OutsideHorizonBins,
		NoPredictionOrigin:        value.NoPredictionOrigin,
		SourceSatelliteMissing:    value.SourceSatelliteMissing,
		SourceEpochMissing:        value.SourceEpochMissing,
		ReferenceSatelliteMissing: value.ReferenceSatelliteMissing,
		ReferenceEpochMissing:     value.ReferenceEpochMissing,
	}
}

func publicShadowGradient(value native.ShadowGradient) ShadowGradient {
	return ShadowGradient{
		Fraction:      value.Fraction,
		GradientPerKm: value.GradientPerKm,
	}
}

func publicSISREWeights(value native.SISREWeights) SISREWeights {
	return SISREWeights{
		Radial:                  value.Radial,
		TransverseNormalSquared: value.TransverseNormalSquared,
	}
}

func nativeSISREWeights(value SISREWeights) native.SISREWeights {
	return native.SISREWeights{
		Radial:                  value.Radial,
		TransverseNormalSquared: value.TransverseNormalSquared,
	}
}

func publicSphericalHarmonicCoefficient(value native.SphericalHarmonicCoefficient) SphericalHarmonicCoefficient {
	return SphericalHarmonicCoefficient{
		Degree: value.Degree,
		Order:  value.Order,
		C:      value.C,
		S:      value.S,
	}
}

func publicSunMoon(value native.SunMoon) SunMoon {
	return SunMoon{
		SunM:  value.SunM,
		MoonM: value.MoonM,
	}
}

func publicSystemScore(value native.SystemScore) SystemScore {
	return SystemScore{
		System:   GNSSSystem(value.System),
		BinCount: value.BinCount,
		Coverage: publicScoreCoverage(value.Coverage),
	}
}

func publicVariationalEpoch(value native.VariationalEpoch) VariationalEpoch {
	return VariationalEpoch{
		TDBJ2000S:  value.TDBJ2000S,
		ElapsedTTS: value.ElapsedTTS,
	}
}

func publicVariationalState(value native.VariationalState) VariationalState {
	return VariationalState{
		State:              publicCartesian(value.State),
		StateTransition:    value.StateTransition,
		SensitivityColumns: value.SensitivityColumns,
	}
}

func publicZonalHarmonics(value native.ZonalHarmonics) ZonalHarmonics {
	return ZonalHarmonics{
		MuKm3S2: value.MuKm3S2,
		ReKm:    value.ReKm,
		C20:     value.C20,
		C30:     value.C30,
		C40:     value.C40,
		C50:     value.C50,
		C60:     value.C60,
	}
}

func nativeZonalHarmonics(value ZonalHarmonics) native.ZonalHarmonics {
	return native.ZonalHarmonics{
		MuKm3S2: value.MuKm3S2,
		ReKm:    value.ReKm,
		C20:     value.C20,
		C30:     value.C30,
		C40:     value.C40,
		C50:     value.C50,
		C60:     value.C60,
	}
}
