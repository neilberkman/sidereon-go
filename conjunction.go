package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// EncounterFrame is the C B-plane frame. Unit vectors are dimensionless;
// relative position and miss distance are km; relative velocity and speed are
// km/s.
type EncounterFrame struct {
	XHat                   [3]float64
	YHat                   [3]float64
	ZHat                   [3]float64
	RelativePositionKm     [3]float64
	RelativeVelocityKmPerS [3]float64
	MissKm                 float64
	RelativeSpeedKmPerS    float64
}

// ConjunctionState contains an ECI position/velocity and a 3x3 position
// covariance. Position is km, velocity is km/s, and covariance is km^2.
type ConjunctionState struct {
	PositionKm     [3]float64
	VelocityKmPerS [3]float64
	CovarianceKm2  [3][3]float64
}

// CollisionPC contains collision probability and encounter-plane diagnostics.
// Miss distance and sigmas are km; relative speed is km/s; PC is a probability.
type CollisionPC struct {
	PC                  float64
	MissKm              float64
	RelativeSpeedKmPerS float64
	SigmaXKm            float64
	SigmaZKm            float64
}

// PCMethod selects the C collision-probability model.
type PCMethod uint32

const (
	PCMethodFosterEqualArea PCMethod = PCMethod(native.PCMethodFosterEqualAreaValue)
	PCMethodFosterNumerical PCMethod = PCMethod(native.PCMethodFosterNumericalValue)
	PCMethodAlfano2005      PCMethod = PCMethod(native.PCMethodAlfano2005Value)
)

// PropagationForceModel selects the C force model used to transport TCA
// covariances.
type PropagationForceModel uint32

const (
	PropagationForceModelTwoBody     PropagationForceModel = PropagationForceModel(native.PropagationForceModelTwoBodyValue)
	PropagationForceModelTwoBodyJ2   PropagationForceModel = PropagationForceModel(native.PropagationForceModelTwoBodyJ2Value)
	PropagationForceModelComposite   PropagationForceModel = PropagationForceModel(native.PropagationForceModelCompositeValue)
	PropagationForceModelEarthPhaseA PropagationForceModel = PropagationForceModel(native.PropagationForceModelEarthPhaseAValue)
	PropagationForceModelEarthPhaseB PropagationForceModel = PropagationForceModel(native.PropagationForceModelEarthPhaseBValue)
)

// PropagationIntegrator selects the C covariance-transport integrator.
type PropagationIntegrator uint32

const (
	PropagationIntegratorDP54 PropagationIntegrator = PropagationIntegrator(native.PropagationIntegratorDP54Value)
	PropagationIntegratorRK4  PropagationIntegrator = PropagationIntegrator(native.PropagationIntegratorRK4Value)
)

func nativeConjunctionState(value ConjunctionState) native.ConjunctionState {
	return native.ConjunctionState{PositionKm: value.PositionKm, VelocityKmPerS: value.VelocityKmPerS, CovarianceKm2: value.CovarianceKm2}
}

func publicEncounterFrame(value native.EncounterFrame) EncounterFrame {
	return EncounterFrame{XHat: value.XHat, YHat: value.YHat, ZHat: value.ZHat, RelativePositionKm: value.RelativePositionKm, RelativeVelocityKmPerS: value.RelativeVelocityKmPerS, MissKm: value.MissKm, RelativeSpeedKmPerS: value.RelativeSpeedKmPerS}
}

func publicCollisionPC(value native.CollisionPC) CollisionPC {
	return CollisionPC{PC: value.PC, MissKm: value.MissKm, RelativeSpeedKmPerS: value.RelativeSpeedKmPerS, SigmaXKm: value.SigmaXKm, SigmaZKm: value.SigmaZKm}
}

// BuildEncounterFrame computes the B-plane frame and relative geometry from
// two ECI states. Vectors use km and km/s.
func BuildEncounterFrame(position1Km, velocity1KmPerS, position2Km, velocity2KmPerS [3]float64) (EncounterFrame, error) {
	value, err := native.BuildEncounterFrame(position1Km, velocity1KmPerS, position2Km, velocity2KmPerS)
	return publicEncounterFrame(value), publicError(err)
}

// EncounterPlaneCovariance projects a position covariance in km^2 onto the
// frame's x/z encounter plane and returns a row-major 2x2 covariance.
func EncounterPlaneCovariance(frame EncounterFrame, covarianceKm2 [3][3]float64) ([2][2]float64, error) {
	value, err := native.EncounterPlaneCovariance(native.EncounterFrame{XHat: frame.XHat, YHat: frame.YHat, ZHat: frame.ZHat, RelativePositionKm: frame.RelativePositionKm, RelativeVelocityKmPerS: frame.RelativeVelocityKmPerS, MissKm: frame.MissKm, RelativeSpeedKmPerS: frame.RelativeSpeedKmPerS}, covarianceKm2)
	return value, publicError(err)
}

// CollisionProbability computes Pc from two ECI states, their 3x3 position
// covariances in km^2, a hard-body radius in km, and a C PCMethod.
func CollisionProbability(object1, object2 ConjunctionState, hardBodyRadiusKm float64, method PCMethod) (CollisionPC, error) {
	value, err := native.CollisionProbability(nativeConjunctionState(object1), nativeConjunctionState(object2), hardBodyRadiusKm, uint32(method))
	return publicCollisionPC(value), publicError(err)
}

// TCAFinderOptions controls the C coarse bracketing and time refinement.
// Both fields are seconds.
type TCAFinderOptions struct {
	CoarseStepSeconds    float64
	TimeToleranceSeconds float64
}

// DefaultTCAFinderOptions returns the C engine defaults.
func DefaultTCAFinderOptions() (TCAFinderOptions, error) {
	value, err := native.TCAFinderDefaults()
	return TCAFinderOptions{CoarseStepSeconds: value.CoarseStepSeconds, TimeToleranceSeconds: value.TimeToleranceSeconds}, publicError(err)
}

// TCACandidate is a local time-of-closest-approach result. The absolute time
// is a split Julian date, while the relative time is seconds from the search
// window start. Relative position is primary minus secondary in the C TEME
// convention; geometry uses km and km/s.
type TCACandidate struct {
	TCATimeJDWhole             float64
	TCATimeJDFraction          float64
	TCASecondsSinceWindowStart float64
	MissDistanceKm             float64
	RelativePositionKm         [3]float64
	RelativeVelocityKmPerS     [3]float64
}

// TCA is the absolute split Julian date represented by a candidate.
func (c TCACandidate) TCA() JulianDate {
	return JulianDate{Whole: c.TCATimeJDWhole, Fraction: c.TCATimeJDFraction}
}

// TCAConjunction pairs a refined candidate with C collision-probability
// diagnostics.
type TCAConjunction struct {
	Candidate            TCACandidate
	CollisionProbability CollisionPC
}

func publicTCACandidate(value native.TCACandidate) TCACandidate {
	return TCACandidate{TCATimeJDWhole: value.TCATimeJDWhole, TCATimeJDFraction: value.TCATimeJDFraction, TCASecondsSinceWindowStart: value.TCASecondsSinceWindowStart, MissDistanceKm: value.MissDistanceKm, RelativePositionKm: value.RelativePositionKm, RelativeVelocityKmPerS: value.RelativeVelocityKmPerS}
}

func publicTCAConjunction(value native.TCAConjunction) TCAConjunction {
	return TCAConjunction{Candidate: publicTCACandidate(value.Candidate), CollisionProbability: publicCollisionPC(value.CollisionProbability)}
}

// FindTCACandidates searches two TLEs over a split-Julian-date window. TLE
// parsing, propagation, bracketing, and refinement all occur in C. A nil
// options pointer requests the C defaults.
func FindTCACandidates(primaryLine1, primaryLine2, secondaryLine1, secondaryLine2 string, start, end JulianDate, options *TCAFinderOptions) ([]TCACandidate, error) {
	var nativeOptions *native.TCAFinderOptions
	if options != nil {
		nativeOptions = &native.TCAFinderOptions{CoarseStepSeconds: options.CoarseStepSeconds, TimeToleranceSeconds: options.TimeToleranceSeconds}
	}
	values, err := native.FindTCACandidates([4]string{primaryLine1, primaryLine2, secondaryLine1, secondaryLine2}, start.Whole, start.Fraction, end.Whole, end.Fraction, nativeOptions)
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]TCACandidate, len(values))
	for i := range out {
		out[i] = publicTCACandidate(values[i])
	}
	return out, nil
}

// TCAPCOptions supplies the C TCA collision-probability inputs. When
// UseDefaultCovariance is true, the two covariance matrices are ignored.
type TCAPCOptions struct {
	HardBodyRadiusKm       float64
	Method                 PCMethod
	UseDefaultCovariance   bool
	PrimaryCovarianceKm2   [3][3]float64
	SecondaryCovarianceKm2 [3][3]float64
}

// TCAPropagatedCovariancePCOptions supplies initial state covariances and C
// propagation controls for covariance-aware TCA conjunctions. Covariances are
// row-major position/velocity state covariances in the C km-based convention;
// tolerances are dimensionless, integration steps are seconds, and MaxSteps
// is a count. When MuKm3PerS2Enabled is false, C uses its Earth default.
type TCAPropagatedCovariancePCOptions struct {
	HardBodyRadiusKm     float64
	Method               PCMethod
	PrimaryCovariance0   [6][6]float64
	SecondaryCovariance0 [6][6]float64
	ForceModel           PropagationForceModel
	Integrator           PropagationIntegrator
	AbsTol               float64
	RelTol               float64
	InitialStepSeconds   float64
	MinStepSeconds       float64
	MaxStepSeconds       float64
	MaxSteps             uint32
	MuKm3PerS2Enabled    bool
	MuKm3PerS2           float64
}

// TCATLEPair is one caller-owned TLE pair in a serial screening catalog. C
// copies the lines during each screening call; the Go strings are not retained.
type TCATLEPair struct {
	Line1 string
	Line2 string
}

// TCAScreeningHit identifies a threshold TCA candidate by its original
// secondary catalog index. Results retain C's screening order.
type TCAScreeningHit struct {
	SecondaryIndex int
	Candidate      TCACandidate
}

// TCAScreeningConjunctionHit identifies a threshold TCA conjunction by its
// original secondary catalog index. Results retain C's screening order.
type TCAScreeningConjunctionHit struct {
	SecondaryIndex int
	Conjunction    TCAConjunction
}

// TCACollisionProbability computes Pc for a candidate already refined by C.
func TCACollisionProbability(candidate TCACandidate, options TCAPCOptions) (TCAConjunction, error) {
	value, err := native.TCACollisionProbability(native.TCACandidate{TCATimeJDWhole: candidate.TCATimeJDWhole, TCATimeJDFraction: candidate.TCATimeJDFraction, TCASecondsSinceWindowStart: candidate.TCASecondsSinceWindowStart, MissDistanceKm: candidate.MissDistanceKm, RelativePositionKm: candidate.RelativePositionKm, RelativeVelocityKmPerS: candidate.RelativeVelocityKmPerS}, native.TCAPCOptions{HardBodyRadiusKm: options.HardBodyRadiusKm, Method: uint32(options.Method), UseDefaultCovariance: options.UseDefaultCovariance, PrimaryCovarianceKm2: options.PrimaryCovarianceKm2, SecondaryCovarianceKm2: options.SecondaryCovarianceKm2})
	return publicTCAConjunction(value), publicError(err)
}

// FindTCAConjunctions combines C TCA refinement and collision-probability
// evaluation for every local minimum in a TLE pair's search window.
func FindTCAConjunctions(primaryLine1, primaryLine2, secondaryLine1, secondaryLine2 string, start, end JulianDate, tcaOptions *TCAFinderOptions, pcOptions TCAPCOptions) ([]TCAConjunction, error) {
	var nativeOptions *native.TCAFinderOptions
	if tcaOptions != nil {
		nativeOptions = &native.TCAFinderOptions{CoarseStepSeconds: tcaOptions.CoarseStepSeconds, TimeToleranceSeconds: tcaOptions.TimeToleranceSeconds}
	}
	values, err := native.FindTCAConjunctions([4]string{primaryLine1, primaryLine2, secondaryLine1, secondaryLine2}, start.Whole, start.Fraction, end.Whole, end.Fraction, nativeOptions, native.TCAPCOptions{HardBodyRadiusKm: pcOptions.HardBodyRadiusKm, Method: uint32(pcOptions.Method), UseDefaultCovariance: pcOptions.UseDefaultCovariance, PrimaryCovarianceKm2: pcOptions.PrimaryCovarianceKm2, SecondaryCovarianceKm2: pcOptions.SecondaryCovarianceKm2})
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]TCAConjunction, len(values))
	for i := range out {
		out[i] = publicTCAConjunction(values[i])
	}
	return out, nil
}

// FindTCAConjunctionsWithPropagatedCovarianceFromTLEs combines C TCA
// refinement with covariance propagation to each candidate TCA.
func FindTCAConjunctionsWithPropagatedCovarianceFromTLEs(primaryLine1, primaryLine2, secondaryLine1, secondaryLine2 string, start, end JulianDate, tcaOptions *TCAFinderOptions, pcOptions TCAPropagatedCovariancePCOptions) ([]TCAConjunction, error) {
	var nativeOptions *native.TCAFinderOptions
	if tcaOptions != nil {
		nativeOptions = &native.TCAFinderOptions{CoarseStepSeconds: tcaOptions.CoarseStepSeconds, TimeToleranceSeconds: tcaOptions.TimeToleranceSeconds}
	}
	values, err := native.FindTCAConjunctionsWithPropagatedCovarianceFromTLEs(
		[4]string{primaryLine1, primaryLine2, secondaryLine1, secondaryLine2},
		start.Whole, start.Fraction, end.Whole, end.Fraction, nativeOptions,
		native.TCAPropagatedCovariancePCOptions{
			HardBodyRadiusKm:     pcOptions.HardBodyRadiusKm,
			Method:               uint32(pcOptions.Method),
			PrimaryCovariance0:   pcOptions.PrimaryCovariance0,
			SecondaryCovariance0: pcOptions.SecondaryCovariance0,
			ForceModel:           uint32(pcOptions.ForceModel),
			Integrator:           uint32(pcOptions.Integrator),
			AbsTol:               pcOptions.AbsTol,
			RelTol:               pcOptions.RelTol,
			InitialStepSeconds:   pcOptions.InitialStepSeconds,
			MinStepSeconds:       pcOptions.MinStepSeconds,
			MaxStepSeconds:       pcOptions.MaxStepSeconds,
			MaxSteps:             pcOptions.MaxSteps,
			MuKm3PerS2Enabled:    pcOptions.MuKm3PerS2Enabled,
			MuKm3PerS2:           pcOptions.MuKm3PerS2,
		},
	)
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]TCAConjunction, len(values))
	for i := range out {
		out[i] = publicTCAConjunction(values[i])
	}
	return out, nil
}

// ScreenTCACandidatesFromTLECatalog serially screens a primary TLE against
// the supplied catalog in C. Secondary indices and result ordering are copied
// exactly from the C screening output.
func ScreenTCACandidatesFromTLECatalog(primaryLine1, primaryLine2 string, secondaries []TCATLEPair, start, end JulianDate, missDistanceThresholdKm float64, options *TCAFinderOptions) ([]TCAScreeningHit, error) {
	nativeSecondaries := make([]native.TCATLEPair, len(secondaries))
	for i, pair := range secondaries {
		nativeSecondaries[i] = native.TCATLEPair{Line1: pair.Line1, Line2: pair.Line2}
	}
	values, err := native.ScreenTCACandidatesFromTLECatalog(primaryLine1, primaryLine2, nativeSecondaries, start.Whole, start.Fraction, end.Whole, end.Fraction, missDistanceThresholdKm, nativeOptions(options))
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]TCAScreeningHit, len(values))
	for i, value := range values {
		out[i] = TCAScreeningHit{SecondaryIndex: value.SecondaryIndex, Candidate: publicTCACandidate(value.Candidate)}
	}
	return out, nil
}

// ScreenTCAConjunctionsFromTLECatalog serially screens a primary TLE against
// the supplied catalog and computes C collision probabilities for threshold
// breaches. Secondary indices and result ordering are preserved.
func ScreenTCAConjunctionsFromTLECatalog(primaryLine1, primaryLine2 string, secondaries []TCATLEPair, start, end JulianDate, missDistanceThresholdKm float64, tcaOptions *TCAFinderOptions, pcOptions TCAPCOptions) ([]TCAScreeningConjunctionHit, error) {
	nativeSecondaries := make([]native.TCATLEPair, len(secondaries))
	for i, pair := range secondaries {
		nativeSecondaries[i] = native.TCATLEPair{Line1: pair.Line1, Line2: pair.Line2}
	}
	values, err := native.ScreenTCAConjunctionsFromTLECatalog(primaryLine1, primaryLine2, nativeSecondaries, start.Whole, start.Fraction, end.Whole, end.Fraction, missDistanceThresholdKm, nativeOptions(tcaOptions), native.TCAPCOptions{HardBodyRadiusKm: pcOptions.HardBodyRadiusKm, Method: uint32(pcOptions.Method), UseDefaultCovariance: pcOptions.UseDefaultCovariance, PrimaryCovarianceKm2: pcOptions.PrimaryCovarianceKm2, SecondaryCovarianceKm2: pcOptions.SecondaryCovarianceKm2})
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]TCAScreeningConjunctionHit, len(values))
	for i, value := range values {
		out[i] = TCAScreeningConjunctionHit{SecondaryIndex: value.SecondaryIndex, Conjunction: publicTCAConjunction(value.Conjunction)}
	}
	return out, nil
}

func nativeOptions(value *TCAFinderOptions) *native.TCAFinderOptions {
	if value == nil {
		return nil
	}
	return &native.TCAFinderOptions{CoarseStepSeconds: value.CoarseStepSeconds, TimeToleranceSeconds: value.TimeToleranceSeconds}
}
