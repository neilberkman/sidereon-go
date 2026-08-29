package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// SourceSolveMode selects time-of-arrival or time-difference-of-arrival
// source localization.
type SourceSolveMode uint32

const (
	// SourceSolveTOA estimates source position and origin time.
	SourceSolveTOA SourceSolveMode = SourceSolveMode(native.SourceSolveTOAValue)
	// SourceSolveTDOA estimates source position relative to a reference sensor.
	SourceSolveTDOA SourceSolveMode = SourceSolveMode(native.SourceSolveTDOAValue)
)

// SourceLoss selects the robust loss applied by the native source solver.
type SourceLoss uint32

const (
	// SourceLossLinear identifies the source loss linear case.
	SourceLossLinear SourceLoss = SourceLoss(native.SourceLossLinearValue)
	// SourceLossSoftL1 identifies the source loss soft l1 case.
	SourceLossSoftL1 SourceLoss = SourceLoss(native.SourceLossSoftL1Value)
	// SourceLossHuber identifies the source loss huber case.
	SourceLossHuber SourceLoss = SourceLoss(native.SourceLossHuberValue)
	// SourceLossCauchy identifies the source loss cauchy case.
	SourceLossCauchy SourceLoss = SourceLoss(native.SourceLossCauchyValue)
	// SourceLossArctan identifies the source loss arctan case.
	SourceLossArctan SourceLoss = SourceLoss(native.SourceLossArctanValue)
)

// SourceSensor describes one Cartesian source-localization sensor.
type SourceSensor struct {
	// Dimension is the state dimension.
	Dimension int
	// PositionM is the position m in metres.
	PositionM [3]float64
	// HasPropagationSpeed reports whether the has propagation speed field is present.
	HasPropagationSpeed bool
	// PropagationSpeedMS is the propagation speed ms value for SourceSensor.
	PropagationSpeedMS float64
}

func nativeSourceSensors(values []SourceSensor) []native.NativeSourceSensor {
	result := make([]native.NativeSourceSensor, len(values))
	for i, value := range values {
		result[i] = native.NativeSourceSensor{Dimension: value.Dimension, PositionM: value.PositionM, HasPropagationSpeed: value.HasPropagationSpeed, PropagationSpeedMS: value.PropagationSpeedMS}
	}
	return result
}

// SourceInitialGuess is a detached C-computed source-localization seed.
type SourceInitialGuess struct {
	// Dimension is the state dimension.
	Dimension int
	// PositionM is the position m in metres.
	PositionM [3]float64
	// HasOriginTime reports whether the has origin time field is present.
	HasOriginTime bool
	// OriginTimeS is the origin time s in seconds.
	OriginTimeS float64
	// ResidualRMSS is the residual rmss in seconds per second.
	ResidualRMSS float64
}

func sourceInitialGuess(value native.NativeSourceInitialGuess) SourceInitialGuess {
	return SourceInitialGuess{Dimension: value.Dimension, PositionM: value.PositionM, HasOriginTime: value.HasOriginTime, OriginTimeS: value.OriginTimeS, ResidualRMSS: value.ResidualRMSS}
}

// SourceLocateOptions controls the native nonlinear source-localization solve.
type SourceLocateOptions struct {
	// Mode is the selected mode.
	Mode SourceSolveMode
	// ReferenceSensor is the reference sensor value for SourceLocateOptions.
	ReferenceSensor int
	// TimingSigmaS is the timing sigma s in seconds; FScaleS is the f scale s in seconds.
	TimingSigmaS, FScaleS float64
	// Loss is the loss value for SourceLocateOptions.
	Loss SourceLoss
	// HasFTOL reports whether the has ftol field is present; HasXTOL reports whether the has xtol field is present.
	HasFTOL, HasXTOL bool
	// FTOL is the ftol value for SourceLocateOptions; XTOL is the xtol value for SourceLocateOptions.
	FTOL, XTOL float64
	// HasGTOL reports whether the has gtol field is present.
	HasGTOL bool
	// GTOL is the gtol value for SourceLocateOptions.
	GTOL float64
	// HasMaxNFEV reports whether the has max nfev field is present.
	HasMaxNFEV bool
	// MaxNFEV is the max nfev value for SourceLocateOptions.
	MaxNFEV int
}

func nativeSourceLocateOptions(value SourceLocateOptions) native.NativeSourceLocateOptions {
	return native.NativeSourceLocateOptions{Mode: int(value.Mode), ReferenceSensor: value.ReferenceSensor, TimingSigmaS: value.TimingSigmaS, FScaleS: value.FScaleS, Loss: uint32(value.Loss), HasFTOL: value.HasFTOL, FTOL: value.FTOL, HasXTOL: value.HasXTOL, XTOL: value.XTOL, HasGTOL: value.HasGTOL, GTOL: value.GTOL, HasMaxNFEV: value.HasMaxNFEV, MaxNFEV: value.MaxNFEV}
}

func sourceLocateOptions(value native.NativeSourceLocateOptions) SourceLocateOptions {
	return SourceLocateOptions{Mode: SourceSolveMode(value.Mode), ReferenceSensor: value.ReferenceSensor, TimingSigmaS: value.TimingSigmaS, FScaleS: value.FScaleS, Loss: SourceLoss(value.Loss), HasFTOL: value.HasFTOL, FTOL: value.FTOL, HasXTOL: value.HasXTOL, XTOL: value.XTOL, HasGTOL: value.HasGTOL, GTOL: value.GTOL, HasMaxNFEV: value.HasMaxNFEV, MaxNFEV: value.MaxNFEV}
}

// SourceLocateOptionsInit returns the native source-solver defaults.
func SourceLocateOptionsInit() (SourceLocateOptions, error) {
	value, err := native.SourceLocateOptionsInit()
	return sourceLocateOptions(value), publicError(err)
}

// ChanHOInitialGuess computes the deprecated Chan-Ho seed through C.
func ChanHOInitialGuess(sensors []SourceSensor, arrivalsS []float64, speedMS float64, mode SourceSolveMode, referenceSensor int) (SourceInitialGuess, error) {
	value, err := native.ChanHOInitialGuess(nativeSourceSensors(sensors), append([]float64(nil), arrivalsS...), speedMS, uint32(mode), referenceSensor)
	return sourceInitialGuess(value), publicError(err)
}

// ChanHoInitialGuess is the conventional mixed-case alias for
// ChanHOInitialGuess.
func ChanHoInitialGuess(sensors []SourceSensor, arrivalsS []float64, speedMS float64, mode SourceSolveMode, referenceSensor int) (SourceInitialGuess, error) {
	return ChanHOInitialGuess(sensors, arrivalsS, speedMS, mode, referenceSensor)
}

// ClosedFormInitialGuess computes the closed-form native source seed.
func ClosedFormInitialGuess(sensors []SourceSensor, arrivalsS []float64, speedMS float64, mode SourceSolveMode, referenceSensor int) (SourceInitialGuess, error) {
	value, err := native.ClosedFormInitialGuess(nativeSourceSensors(sensors), append([]float64(nil), arrivalsS...), speedMS, uint32(mode), referenceSensor)
	return sourceInitialGuess(value), publicError(err)
}

// SourceCovariance is a detached source-state covariance returned by C.
type SourceCovariance struct {
	// Dimension is the state dimension; StateDimension is the state dimension.
	Dimension, StateDimension int
	// State contains the fixed-size array for this record.
	State [16]float64
	// PositionM2 is the position m2 in square metres.
	PositionM2 [9]float64
	// HasOriginTimeS2 reports whether the has origin time s2 field is present.
	HasOriginTimeS2 bool
	// OriginTimeS2 is the origin time s2 in seconds squared; TimingSigmaS is the timing sigma s in seconds.
	OriginTimeS2, TimingSigmaS float64
}

// SourceCRLBResult contains timing DOP and the corresponding native covariance.
type SourceCRLBResult struct {
	// DOP is the dop value for SourceCRLBResult.
	DOP DOP
	// Covariance is the covariance value for SourceCRLBResult.
	Covariance SourceCovariance
}

// SourceSensorInfluence is a detached leave-one-out sensor diagnostic.
type SourceSensorInfluence struct {
	// SensorIndex identifies or counts this record.
	SensorIndex int
	// ResidualS is the residual s value for SourceSensorInfluence.
	ResidualS float64
	// HasLeaveOneOutResidual reports whether the has leave one out residual field is present.
	HasLeaveOneOutResidual bool
	// LeaveOneOutResidualS is the leave one out residual s value for SourceSensorInfluence.
	LeaveOneOutResidualS float64
	// HasPositionDelta reports whether the has position delta field is present.
	HasPositionDelta bool
	// PositionDeltaM is the position delta m in metres.
	PositionDeltaM float64
	// HasOriginTimeDelta reports whether the has origin time delta field is present.
	HasOriginTimeDelta bool
	// OriginTimeDeltaS is the origin time delta s value for SourceSensorInfluence.
	OriginTimeDeltaS float64
	// LossWeight is the loss weight value for SourceSensorInfluence; Score is the score value for SourceSensorInfluence.
	LossWeight, Score float64
}

// SourceResidual is one detached source-solve residual row.
type SourceResidual struct {
	// SensorIndex identifies or counts this record; ReferenceSensorIndex identifies or counts this record.
	SensorIndex, ReferenceSensorIndex int
	// HasReferenceSensor reports whether the has reference sensor field is present.
	HasReferenceSensor bool
	// ResidualS is the residual s value for SourceResidual.
	ResidualS float64
}

// SourceSolutionSummary contains detached native solve diagnostics.
type SourceSolutionSummary struct {
	// Dimension is the state dimension; ResidualCount identifies or counts this record; InfluenceCount identifies or counts this record.
	Dimension, ResidualCount, InfluenceCount int
	// PositionM is the position m in metres.
	PositionM [3]float64
	// HasOriginTime reports whether the has origin time field is present.
	HasOriginTime bool
	// OriginTimeS is the origin time s in seconds.
	OriginTimeS float64
	// HasCovariance reports whether the has covariance field is present.
	HasCovariance bool
	// GeometryQuality is the geometry-quality diagnostics.
	GeometryQuality GeometryQuality
	// InitialGuess is the initial state guess.
	InitialGuess SourceInitialGuess
	// Status is the native status code.
	Status int32
	// NFEV is the nfev value for SourceSolutionSummary; NJEV is the njev value for SourceSolutionSummary.
	NFEV, NJEV int
	// Cost is the cost value for SourceSolutionSummary; Optimality is the optimality value for SourceSolutionSummary.
	Cost, Optimality float64
}

func sourceCovariance(value native.NativeSourceCovariance) SourceCovariance {
	return SourceCovariance{Dimension: value.Dimension, StateDimension: value.StateDimension, State: value.State, PositionM2: value.PositionM2, HasOriginTimeS2: value.HasOriginTimeS2, OriginTimeS2: value.OriginTimeS2, TimingSigmaS: value.TimingSigmaS}
}

func sourceCRLB(value native.NativeSourceCrlb) SourceCRLBResult {
	return SourceCRLBResult{DOP: DOP{GDOP: value.DOP.GDOP, PDOP: value.DOP.PDOP, HDOP: value.DOP.HDOP, VDOP: value.DOP.VDOP, TDOP: value.DOP.TDOP}, Covariance: sourceCovariance(value.Covariance)}
}

func sourceSummary(value native.NativeSourceSolutionSummary) SourceSolutionSummary {
	return SourceSolutionSummary{Dimension: value.Dimension, ResidualCount: value.ResidualCount, InfluenceCount: value.InfluenceCount, PositionM: value.PositionM, HasOriginTime: value.HasOriginTime, OriginTimeS: value.OriginTimeS, HasCovariance: value.HasCovariance, GeometryQuality: GeometryQuality{Tier: ObservabilityTier(value.GeometryQuality.Tier), Redundancy: value.GeometryQuality.Redundancy, Rank: value.GeometryQuality.Rank, ConditionNumber: value.GeometryQuality.ConditionNumber, GDOP: value.GeometryQuality.GDOP, RAIMCheckable: value.GeometryQuality.RAIMCheckable, CovarianceValidated: value.GeometryQuality.CovarianceValidated}, InitialGuess: sourceInitialGuess(value.InitialGuess), Status: value.Status, NFEV: value.NFEV, NJEV: value.NJEV, Cost: value.Cost, Optimality: value.Optimality}
}

// SourceSolution is a synchronized, non-copyable native source solve handle.
type SourceSolution struct {
	_      noCopy
	handle *native.SourceSolution
}

// LocateSource solves a source position with the native default influence
// behavior. Inputs are copied before entering C.
func LocateSource(sensors []SourceSensor, arrivalsS []float64, speedMS float64, options *SourceLocateOptions) (*SourceSolution, error) {
	var nativeOptions *native.NativeSourceLocateOptions
	if options != nil {
		value := nativeSourceLocateOptions(*options)
		nativeOptions = &value
	}
	value, err := native.LocateSource(nativeSourceSensors(sensors), append([]float64(nil), arrivalsS...), speedMS, nativeOptions)
	if err != nil {
		return nil, publicError(err)
	}
	return &SourceSolution{handle: value}, nil
}

// LocateSourceWith solves a source position and controls influence diagnostics.
func LocateSourceWith(sensors []SourceSensor, arrivalsS []float64, speedMS float64, options *SourceLocateOptions, includeInfluence bool) (*SourceSolution, error) {
	var nativeOptions *native.NativeSourceLocateOptions
	if options != nil {
		value := nativeSourceLocateOptions(*options)
		nativeOptions = &value
	}
	value, err := native.LocateSourceWith(nativeSourceSensors(sensors), append([]float64(nil), arrivalsS...), speedMS, nativeOptions, includeInfluence)
	if err != nil {
		return nil, publicError(err)
	}
	return &SourceSolution{handle: value}, nil
}

// Close releases the native source solution and is idempotent.
func (s *SourceSolution) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return publicError(s.handle.Close())
}

// Covariance returns the detached covariance and its native availability bit.
func (s *SourceSolution) Covariance() (SourceCovariance, bool, error) {
	if s == nil || s.handle == nil {
		return SourceCovariance{}, false, ErrClosed
	}
	value, present, err := s.handle.Covariance()
	return sourceCovariance(value), present, publicError(err)
}

// Influences returns detached per-sensor diagnostics.
func (s *SourceSolution) Influences() ([]SourceSensorInfluence, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	values, err := s.handle.Influences()
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]SourceSensorInfluence, len(values))
	for i, value := range values {
		result[i] = SourceSensorInfluence{SensorIndex: value.SensorIndex, ResidualS: value.ResidualS, HasLeaveOneOutResidual: value.HasLeaveOneOutResidual, LeaveOneOutResidualS: value.LeaveOneOutResidualS, HasPositionDelta: value.HasPositionDelta, PositionDeltaM: value.PositionDeltaM, HasOriginTimeDelta: value.HasOriginTimeDelta, OriginTimeDeltaS: value.OriginTimeDeltaS, LossWeight: value.LossWeight, Score: value.Score}
	}
	return result, nil
}

// Residuals returns detached source-solve residual rows.
func (s *SourceSolution) Residuals() ([]SourceResidual, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	values, err := s.handle.Residuals()
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]SourceResidual, len(values))
	for i, value := range values {
		result[i] = SourceResidual{SensorIndex: value.SensorIndex, ReferenceSensorIndex: value.ReferenceSensorIndex, HasReferenceSensor: value.HasReferenceSensor, ResidualS: value.ResidualS}
	}
	return result, nil
}

// Summary returns detached native source-solve diagnostics.
func (s *SourceSolution) Summary() (SourceSolutionSummary, error) {
	if s == nil || s.handle == nil {
		return SourceSolutionSummary{}, ErrClosed
	}
	value, err := s.handle.Summary()
	return sourceSummary(value), publicError(err)
}

// SourceCRLB computes the native timing Cramér–Rao bound and DOP.
func SourceCRLB(sensors []SourceSensor, sourcePositionM []float64, speedMS, timingSigmaS float64) (SourceCRLBResult, error) {
	value, err := native.SourceCRLB(nativeSourceSensors(sensors), append([]float64(nil), sourcePositionM...), speedMS, timingSigmaS)
	return sourceCRLB(value), publicError(err)
}

// SourceCrlb is the mixed-case alias for SourceCRLB.
func SourceCrlb(sensors []SourceSensor, sourcePositionM []float64, speedMS, timingSigmaS float64) (SourceCRLBResult, error) {
	return SourceCRLB(sensors, sourcePositionM, speedMS, timingSigmaS)
}

// SourceDOP computes the native timing dilution of precision.
func SourceDOP(sensors []SourceSensor, sourcePositionM []float64, speedMS float64) (DOP, error) {
	value, err := native.SourceDOP(nativeSourceSensors(sensors), append([]float64(nil), sourcePositionM...), speedMS)
	return DOP{GDOP: value.GDOP, PDOP: value.PDOP, HDOP: value.HDOP, VDOP: value.VDOP, TDOP: value.TDOP}, publicError(err)
}

// BroadcastReasonKind explains why a fallback used broadcast ephemeris.
type BroadcastReasonKind uint32

const (
	// BroadcastReasonPreciseUnavailable identifies the broadcast reason precise unavailable case.
	BroadcastReasonPreciseUnavailable BroadcastReasonKind = BroadcastReasonKind(native.BroadcastReasonPreciseUnavailableValue)
	// BroadcastReasonPreciseDegradedUnusable identifies the broadcast reason precise degraded unusable case.
	BroadcastReasonPreciseDegradedUnusable BroadcastReasonKind = BroadcastReasonKind(native.BroadcastReasonPreciseDegradedUnusableValue)
)

// FixSourceKind identifies the ephemeris source of a fallback result.
type FixSourceKind uint32

const (
	// FixSourcePrecise identifies the fix source precise case.
	FixSourcePrecise FixSourceKind = FixSourceKind(native.FixSourcePreciseValue)
	// FixSourceBroadcast identifies the fix source broadcast case.
	FixSourceBroadcast FixSourceKind = FixSourceKind(native.FixSourceBroadcastValue)
)

// DegradationKind describes native product staleness degradation.
type DegradationKind uint32

const (
	// DegradationExact identifies the degradation exact case.
	DegradationExact DegradationKind = DegradationKind(native.DegradationExactValue)
	// DegradationNearestPrior identifies the degradation nearest prior case.
	DegradationNearestPrior DegradationKind = DegradationKind(native.DegradationNearestPriorValue)
	// DegradationDiurnalShift identifies the degradation diurnal shift case.
	DegradationDiurnalShift DegradationKind = DegradationKind(native.DegradationDiurnalShiftValue)
)

// SelectionStatus is the native precise-product selection record.
type SelectionStatus uint32

const (
	// SelectionOK identifies the selection ok case.
	SelectionOK SelectionStatus = 0
	// SelectionNullPointer identifies the selection null pointer case.
	SelectionNullPointer SelectionStatus = 1
	// SelectionInvalidArgument identifies the selection invalid argument case.
	SelectionInvalidArgument SelectionStatus = 2
	// SelectionInvalidToken identifies the selection invalid token case.
	SelectionInvalidToken SelectionStatus = 3
	// SelectionPanic identifies the selection panic case.
	SelectionPanic SelectionStatus = 4
	// SelectionEmptyProductSet identifies the selection empty product set case.
	SelectionEmptyProductSet SelectionStatus = 5
	// SelectionInvalidRange identifies the selection invalid range case.
	SelectionInvalidRange SelectionStatus = 6
	// SelectionNoPriorProduct identifies the selection no prior product case.
	SelectionNoPriorProduct SelectionStatus = 7
	// SelectionBeyondStalenessCap identifies the selection beyond staleness cap case.
	SelectionBeyondStalenessCap SelectionStatus = 8
	// SelectionInvalidProduct identifies the selection invalid product case.
	SelectionInvalidProduct SelectionStatus = 9
	// SelectionInvalidPolicy identifies the selection invalid policy case.
	SelectionInvalidPolicy SelectionStatus = 10
	// SelectionOverflow identifies the selection overflow case.
	SelectionOverflow SelectionStatus = 11
)

// StalenessMetadata is detached fallback provenance.
type StalenessMetadata struct {
	// Kind is the event or record kind.
	Kind DegradationKind
	// RequestedEpochJ2000S is the requested epoch j2000 s in seconds; SourceEpochJ2000S is the source epoch j2000 s in seconds; StalenessS is the staleness s in seconds; StalenessDays is the staleness days in days.
	RequestedEpochJ2000S, SourceEpochJ2000S, StalenessS, StalenessDays float64
}

// SourcedSolution is a synchronized native fallback result handle.
type SourcedSolution struct {
	_      noCopy
	handle *native.SourcedSolution
}

// Close releases the native sourced solution and is idempotent.
func (s *SourcedSolution) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return publicError(s.handle.Close())
}

// BroadcastReason returns C's fallback reason, selection status, metadata and presence.
func (s *SourcedSolution) BroadcastReason() (BroadcastReasonKind, SelectionStatus, StalenessMetadata, bool, error) {
	if s == nil || s.handle == nil {
		return 0, 0, StalenessMetadata{}, false, ErrClosed
	}
	reason, selection, metadata, present, err := s.handle.BroadcastReason()
	return BroadcastReasonKind(reason), SelectionStatus(selection), stalenessMetadata(metadata), present, publicError(err)
}

// IsPreciseExact reports whether C used an exact precise product.
func (s *SourcedSolution) IsPreciseExact() (bool, error) {
	if s == nil || s.handle == nil {
		return false, ErrClosed
	}
	value, err := s.handle.IsPreciseExact()
	return value, publicError(err)
}

// SourceKind reports which ephemeris source produced the fix.
func (s *SourcedSolution) SourceKind() (FixSourceKind, error) {
	if s == nil || s.handle == nil {
		return 0, ErrClosed
	}
	value, err := s.handle.SourceKind()
	return FixSourceKind(value), publicError(err)
}

// Staleness returns C's detached source staleness metadata.
func (s *SourcedSolution) Staleness() (StalenessMetadata, bool, error) {
	if s == nil || s.handle == nil {
		return StalenessMetadata{}, false, ErrClosed
	}
	value, present, err := s.handle.Staleness()
	return stalenessMetadata(value), present, publicError(err)
}

// Solution returns a detached receiver solution copied by C.
func (s *SourcedSolution) Solution() (SPPSolution, error) {
	if s == nil || s.handle == nil {
		return SPPSolution{}, ErrClosed
	}
	value, err := s.handle.Solution()
	return publicSPPSolution(value), publicError(err)
}

func stalenessMetadata(value native.NativeStalenessMetadata) StalenessMetadata {
	return StalenessMetadata{Kind: DegradationKind(value.Kind), RequestedEpochJ2000S: value.RequestedEpochJ2000S, SourceEpochJ2000S: value.SourceEpochJ2000S, StalenessS: value.StalenessS, StalenessDays: value.StalenessDays}
}

// CNAVURANEDM evaluates the native CNAV time-dependent NED URA value.
func CNAVURANEDM(params BroadcastCNAV, queryWeek uint32, queryTOW float64) (float64, bool, error) {
	value, present, err := native.CNAVURANEDM(native.NativeCnavParameters{Present: params.Present, ADOTMS: params.ADOTMPerS, DeltaN0DotRadS2: params.DeltaN0DotRadPerS2, TopWeek: params.Top.Week, TopTOWS: params.Top.TOWSeconds, URAEDIndex: params.URAEDIndex, URANED0Index: params.URANED0Index, URANED1Index: params.URANED1Index, URANED2Index: params.URANED2Index, TransmissionTimeSOW: params.TransmissionTimeSOW, HasFlags: params.HasFlags, Flags: params.Flags}, queryWeek, queryTOW)
	return value, present, publicError(err)
}

// CNAVUraNedM is the conventional mixed-case alias for CNAVURANEDM.
func CNAVUraNedM(params BroadcastCNAV, queryWeek uint32, queryTOW float64) (float64, bool, error) {
	return CNAVURANEDM(params, queryWeek, queryTOW)
}

// CNAVURANominalM returns the native nominal CNAV URA for an index.
func CNAVURANominalM(index int8) (float64, bool, error) {
	value, present, err := native.CNAVURANominalM(index)
	return value, present, publicError(err)
}

// CNAVUraNominalM is the conventional mixed-case alias for CNAVURANominalM.
func CNAVUraNominalM(index int8) (float64, bool, error) {
	return CNAVURANominalM(index)
}

// PseudorangeVarianceOptionsInit returns native pseudorange-variance defaults.
func PseudorangeVarianceOptionsInit() (PseudorangeVarianceOptions, error) {
	value, err := native.PseudorangeVarianceOptionsInit()
	return PseudorangeVarianceOptions{AM: value.AM, BM: value.BM, Model: PseudorangeVarianceModel(value.Model), HasCN0: value.HasCN0, CN0DBHz: value.CN0DBHz, CN0ScaleM2: value.CN0ScaleM2}, publicError(err)
}

// PseudorangeVariance evaluates the selected native pseudorange model.
func PseudorangeVariance(elevationDeg float64, options PseudorangeVarianceOptions) (float64, error) {
	value, err := native.PseudorangeVariance(elevationDeg, nativeVarianceOptions(options))
	return value, publicError(err)
}

// SolutionValidationOptions configures native receiver plausibility gates.
type SolutionValidationOptions struct {
	// HasMaxPDOP reports whether the has max pdop field is present.
	HasMaxPDOP bool
	// MaxPDOP is the max pdop value for SolutionValidationOptions; MinPlausibleRadiusM is the min plausible radius m value for SolutionValidationOptions; MaxPlausibleRadiusM is the max plausible radius m value for SolutionValidationOptions; MaxConvergedResidualRMSM is the max converged residual rmsm value for SolutionValidationOptions.
	MaxPDOP, MinPlausibleRadiusM, MaxPlausibleRadiusM, MaxConvergedResidualRMSM float64
}

// SolutionValidationOptionsInit returns native receiver-validation defaults.
func SolutionValidationOptionsInit() (SolutionValidationOptions, error) {
	value, err := native.SolutionValidationOptionsInit()
	return SolutionValidationOptions{HasMaxPDOP: value.HasMaxPDOP, MaxPDOP: value.MaxPDOP, MinPlausibleRadiusM: value.MinPlausibleRadiusM, MaxPlausibleRadiusM: value.MaxPlausibleRadiusM, MaxConvergedResidualRMSM: value.MaxConvergedResidualRMSM}, publicError(err)
}
