//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

const (
	OrbitTypeEllipticalInclinedValue   uint32 = 0
	OrbitTypeEllipticalEquatorialValue uint32 = 1
	OrbitTypeCircularInclinedValue     uint32 = 2
	OrbitTypeCircularEquatorialValue   uint32 = 3

	RetrogradeFactorProgradeValue   uint32 = 0
	RetrogradeFactorRetrogradeValue uint32 = 1

	RelativeFrameRSWValue  uint32 = 0
	RelativeFrameRTNValue  uint32 = 1
	RelativeFrameRICValue  uint32 = 2
	RelativeFrameLVLHValue uint32 = 3

	PCMethodFosterEqualAreaValue uint32 = 0
	PCMethodFosterNumericalValue uint32 = 1
	PCMethodAlfano2005Value      uint32 = 2

	PropagationForceModelTwoBodyValue     uint32 = 0
	PropagationForceModelTwoBodyJ2Value   uint32 = 1
	PropagationForceModelCompositeValue   uint32 = 2
	PropagationForceModelEarthPhaseAValue uint32 = 3
	PropagationForceModelEarthPhaseBValue uint32 = 4

	PropagationIntegratorDP54Value uint32 = 0
	PropagationIntegratorRK4Value  uint32 = 1

	CDMStringFieldCreationDateValue               uint32 = 0
	CDMStringFieldOriginatorValue                 uint32 = 1
	CDMStringFieldMessageIDValue                  uint32 = 2
	CDMStringFieldTCAValue                        uint32 = 3
	CDMStringFieldCollisionProbabilityMethodValue uint32 = 4
	CDMStringFieldObject1DesignatorValue          uint32 = 5
	CDMStringFieldObject1NameValue                uint32 = 6
	CDMStringFieldObject2DesignatorValue          uint32 = 7
	CDMStringFieldObject2NameValue                uint32 = 8

	CDMObjectStringFieldObjectDesignatorValue        uint32 = 0
	CDMObjectStringFieldCatalogNameValue             uint32 = 1
	CDMObjectStringFieldObjectNameValue              uint32 = 2
	CDMObjectStringFieldInternationalDesignatorValue uint32 = 3
	CDMObjectStringFieldObjectTypeValue              uint32 = 4
	CDMObjectStringFieldOperatorContactPositionValue uint32 = 5
	CDMObjectStringFieldOperatorOrganizationValue    uint32 = 6
	CDMObjectStringFieldOperatorPhoneValue           uint32 = 7
	CDMObjectStringFieldOperatorEmailValue           uint32 = 8
	CDMObjectStringFieldEphemerisNameValue           uint32 = 9
	CDMObjectStringFieldCovarianceMethodValue        uint32 = 10
	CDMObjectStringFieldManeuverableValue            uint32 = 11
	CDMObjectStringFieldOrbitCenterValue             uint32 = 12
	CDMObjectStringFieldRefFrameValue                uint32 = 13
	CDMObjectStringFieldGravityModelValue            uint32 = 14
	CDMObjectStringFieldAtmosphericModelValue        uint32 = 15
	CDMObjectStringFieldNBodyPerturbationsValue      uint32 = 16
	CDMObjectStringFieldSolarRadPressureValue        uint32 = 17
	CDMObjectStringFieldEarthTidesValue              uint32 = 18
	CDMObjectStringFieldIntrackThrustValue           uint32 = 19
)

type CartesianState struct {
	EpochTDBSeconds            float64
	PositionKm, VelocityKmPerS [3]float64
}
type ClassicalElements struct {
	P, A, Ecc, Incl, RAAN, ArgP, Nu, ArgLat, TrueLon, LonPer float64
	OrbitType                                                uint32
}
type EquinoctialElements struct {
	A, H, K, P, Q, Lambda float64
	Retrograde            uint32
}
type ModifiedEquinoctialElements struct {
	P, F, G, H, K, L float64
	Retrograde       uint32
}
type KeplerSolution struct {
	AnomalyRad float64
	Iterations int
}
type EncounterFrame struct {
	XHat, YHat, ZHat, RelativePositionKm, RelativeVelocityKmPerS [3]float64
	MissKm, RelativeSpeedKmPerS                                  float64
}
type ConjunctionState struct {
	PositionKm, VelocityKmPerS [3]float64
	CovarianceKm2              [3][3]float64
}
type CollisionPC struct{ PC, MissKm, RelativeSpeedKmPerS, SigmaXKm, SigmaZKm float64 }
type TCAFinderOptions struct{ CoarseStepSeconds, TimeToleranceSeconds float64 }
type TCACandidate struct {
	TCATimeJDWhole, TCATimeJDFraction, TCASecondsSinceWindowStart, MissDistanceKm float64
	RelativePositionKm, RelativeVelocityKmPerS                                    [3]float64
}
type TCAPCOptions struct {
	HardBodyRadiusKm                             float64
	Method                                       uint32
	UseDefaultCovariance                         bool
	PrimaryCovarianceKm2, SecondaryCovarianceKm2 [3][3]float64
}
type TCAConjunction struct {
	Candidate            TCACandidate
	CollisionProbability CollisionPC
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

func orbitalUnavailable() error { return unavailable() }
func CartesianToClassical(CartesianState, float64) (ClassicalElements, error) {
	return ClassicalElements{}, orbitalUnavailable()
}
func ClassicalToCartesian(ClassicalElements, float64) (CartesianState, error) {
	return CartesianState{}, orbitalUnavailable()
}
func ClassicalToEquinoctial(ClassicalElements, uint32) (EquinoctialElements, error) {
	return EquinoctialElements{}, orbitalUnavailable()
}
func EquinoctialToClassical(EquinoctialElements) (ClassicalElements, error) {
	return ClassicalElements{}, orbitalUnavailable()
}
func ClassicalToModifiedEquinoctial(ClassicalElements, uint32) (ModifiedEquinoctialElements, error) {
	return ModifiedEquinoctialElements{}, orbitalUnavailable()
}
func ModifiedEquinoctialToClassical(ModifiedEquinoctialElements) (ClassicalElements, error) {
	return ClassicalElements{}, orbitalUnavailable()
}
func CartesianToEquinoctial(CartesianState, float64, uint32) (EquinoctialElements, error) {
	return EquinoctialElements{}, orbitalUnavailable()
}
func EquinoctialToCartesian(EquinoctialElements, float64) (CartesianState, error) {
	return CartesianState{}, orbitalUnavailable()
}
func CartesianToModifiedEquinoctial(CartesianState, float64, uint32) (ModifiedEquinoctialElements, error) {
	return ModifiedEquinoctialElements{}, orbitalUnavailable()
}
func ModifiedEquinoctialToCartesian(ModifiedEquinoctialElements, float64) (CartesianState, error) {
	return CartesianState{}, orbitalUnavailable()
}
func MeanToEccentricAnomaly(float64, float64) (float64, error) { return 0, orbitalUnavailable() }
func EccentricToMeanAnomaly(float64, float64) (float64, error) { return 0, orbitalUnavailable() }
func EccentricToTrueAnomaly(float64, float64) (float64, error) { return 0, orbitalUnavailable() }
func TrueToEccentricAnomaly(float64, float64) (float64, error) { return 0, orbitalUnavailable() }
func MeanToTrueAnomaly(float64, float64) (float64, error)      { return 0, orbitalUnavailable() }
func TrueToMeanAnomaly(float64, float64) (float64, error)      { return 0, orbitalUnavailable() }
func SolveKepler(float64, float64) (KeplerSolution, error) {
	return KeplerSolution{}, orbitalUnavailable()
}
func PropagateKepler(ClassicalElements, float64, float64) (ClassicalElements, error) {
	return ClassicalElements{}, orbitalUnavailable()
}
func LambertBattin([3]float64, [3]float64, [3]float64, uint32, uint32, int32, float64) ([3]float64, [3]float64, error) {
	return [3]float64{}, [3]float64{}, orbitalUnavailable()
}
func CWSTM(float64, float64) ([6][6]float64, error) { return [6][6]float64{}, orbitalUnavailable() }
func CWPropagate(CartesianState, float64, float64) (CartesianState, error) {
	return CartesianState{}, orbitalUnavailable()
}
func RelativeMeanMotionCircular(float64) (float64, error)         { return 0, orbitalUnavailable() }
func RelativeMeanMotionFromState(CartesianState) (float64, error) { return 0, orbitalUnavailable() }
func RelativeRotation(uint32, CartesianState) ([3][3]float64, error) {
	return [3][3]float64{}, orbitalUnavailable()
}
func RelativeState(CartesianState, CartesianState) (CartesianState, error) {
	return CartesianState{}, orbitalUnavailable()
}
func AbsoluteFromRelative(CartesianState, CartesianState) (CartesianState, error) {
	return CartesianState{}, orbitalUnavailable()
}
func BetaAngleDeg([3]float64, [3]float64) (float64, error) { return 0, orbitalUnavailable() }
func BetaAngleFromStateDeg([3]float64, [3]float64, [3]float64) (float64, error) {
	return 0, orbitalUnavailable()
}
func RTNToECICovariance([3][3]float64, [3]float64, [3]float64) ([3][3]float64, error) {
	return [3][3]float64{}, orbitalUnavailable()
}
func BuildEncounterFrame([3]float64, [3]float64, [3]float64, [3]float64) (EncounterFrame, error) {
	return EncounterFrame{}, orbitalUnavailable()
}
func EncounterPlaneCovariance(EncounterFrame, [3][3]float64) ([2][2]float64, error) {
	return [2][2]float64{}, orbitalUnavailable()
}
func CollisionProbability(ConjunctionState, ConjunctionState, float64, uint32) (CollisionPC, error) {
	return CollisionPC{}, orbitalUnavailable()
}
func TCAFinderDefaults() (TCAFinderOptions, error) { return TCAFinderOptions{}, orbitalUnavailable() }
func FindTCACandidates([4]string, float64, float64, float64, float64, *TCAFinderOptions) ([]TCACandidate, error) {
	return nil, orbitalUnavailable()
}
func TCACollisionProbability(TCACandidate, TCAPCOptions) (TCAConjunction, error) {
	return TCAConjunction{}, orbitalUnavailable()
}
func FindTCAConjunctions([4]string, float64, float64, float64, float64, *TCAFinderOptions, TCAPCOptions) ([]TCAConjunction, error) {
	return nil, orbitalUnavailable()
}
func FindTCAConjunctionsWithPropagatedCovarianceFromTLEs([4]string, float64, float64, float64, float64, *TCAFinderOptions, TCAPropagatedCovariancePCOptions) ([]TCAConjunction, error) {
	return nil, orbitalUnavailable()
}
func ScreenTCACandidatesFromTLECatalog(string, string, []TCATLEPair, float64, float64, float64, float64, float64, *TCAFinderOptions) ([]TCAScreeningHit, error) {
	return nil, orbitalUnavailable()
}
func ScreenTCAConjunctionsFromTLECatalog(string, string, []TCATLEPair, float64, float64, float64, float64, float64, *TCAFinderOptions, TCAPCOptions) ([]TCAScreeningConjunctionHit, error) {
	return nil, orbitalUnavailable()
}

type LookAngle struct{ AzimuthDeg, ElevationDeg, RangeKm float64 }
type OptionalNumber struct {
	Value   float64
	Present bool
}
type CdmNumbers struct{ MissDistanceM, RelativeSpeedMPerS, CollisionProbability, HardBodyRadiusM OptionalNumber }
type CdmObject struct {
	PositionKm, VelocityKmPerS                                                               [3]float64
	CovarianceRTN                                                                            [6]float64
	VelocityCovarianceRTN                                                                    [15]float64
	HasVelocityCovariance                                                                    bool
	ObjectDesignator, CatalogName, ObjectName, InternationalDesignator, ObjectType, RefFrame string
}
type Cdm struct{}

func ParseCdmKVN([]byte) (*Cdm, error)                { return nil, orbitalUnavailable() }
func ParseCdmXML([]byte) (*Cdm, error)                { return nil, orbitalUnavailable() }
func (*Cdm) Close() error                             { return nil }
func (*Cdm) Numbers() (CdmNumbers, error)             { return CdmNumbers{}, orbitalUnavailable() }
func (*Cdm) StringField(uint32) (string, bool, error) { return "", false, orbitalUnavailable() }
func (*Cdm) ObjectStringField(int, uint32) (string, bool, error) {
	return "", false, orbitalUnavailable()
}
func (*Cdm) Object(int) (CdmObject, error) { return CdmObject{}, orbitalUnavailable() }
func (*Cdm) ToKVN() ([]byte, error)        { return nil, orbitalUnavailable() }
func (*Cdm) ToXML() ([]byte, error)        { return nil, orbitalUnavailable() }
