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
	SourceLossLinear SourceLoss = SourceLoss(native.SourceLossLinearValue)
	SourceLossSoftL1 SourceLoss = SourceLoss(native.SourceLossSoftL1Value)
	SourceLossHuber  SourceLoss = SourceLoss(native.SourceLossHuberValue)
	SourceLossCauchy SourceLoss = SourceLoss(native.SourceLossCauchyValue)
	SourceLossArctan SourceLoss = SourceLoss(native.SourceLossArctanValue)
)

// SourceSensor describes one Cartesian source-localization sensor.
type SourceSensor struct {
	Dimension           int
	PositionM           [3]float64
	HasPropagationSpeed bool
	PropagationSpeedMS  float64
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
	Dimension     int
	PositionM     [3]float64
	HasOriginTime bool
	OriginTimeS   float64
	ResidualRMSS  float64
}

func sourceInitialGuess(value native.NativeSourceInitialGuess) SourceInitialGuess {
	return SourceInitialGuess{Dimension: value.Dimension, PositionM: value.PositionM, HasOriginTime: value.HasOriginTime, OriginTimeS: value.OriginTimeS, ResidualRMSS: value.ResidualRMSS}
}

// SourceLocateOptions controls the native nonlinear source-localization solve.
type SourceLocateOptions struct {
	Mode                  SourceSolveMode
	ReferenceSensor       int
	TimingSigmaS, FScaleS float64
	Loss                  SourceLoss
	HasFTOL, HasXTOL      bool
	FTOL, XTOL            float64
	HasGTOL               bool
	GTOL                  float64
	HasMaxNFEV            bool
	MaxNFEV               int
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
	Dimension, StateDimension  int
	State                      [16]float64
	PositionM2                 [9]float64
	HasOriginTimeS2            bool
	OriginTimeS2, TimingSigmaS float64
}

// SourceCRLBResult contains timing DOP and the corresponding native covariance.
type SourceCRLBResult struct {
	DOP        DOP
	Covariance SourceCovariance
}

// SourceSensorInfluence is a detached leave-one-out sensor diagnostic.
type SourceSensorInfluence struct {
	SensorIndex            int
	ResidualS              float64
	HasLeaveOneOutResidual bool
	LeaveOneOutResidualS   float64
	HasPositionDelta       bool
	PositionDeltaM         float64
	HasOriginTimeDelta     bool
	OriginTimeDeltaS       float64
	LossWeight, Score      float64
}

// SourceResidual is one detached source-solve residual row.
type SourceResidual struct {
	SensorIndex, ReferenceSensorIndex int
	HasReferenceSensor                bool
	ResidualS                         float64
}

// SourceSolutionSummary contains detached native solve diagnostics.
type SourceSolutionSummary struct {
	Dimension, ResidualCount, InfluenceCount int
	PositionM                                [3]float64
	HasOriginTime                            bool
	OriginTimeS                              float64
	HasCovariance                            bool
	GeometryQuality                          GeometryQuality
	InitialGuess                             SourceInitialGuess
	Status                                   int32
	NFEV, NJEV                               int
	Cost, Optimality                         float64
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
	BroadcastReasonPreciseUnavailable      BroadcastReasonKind = BroadcastReasonKind(native.BroadcastReasonPreciseUnavailableValue)
	BroadcastReasonPreciseDegradedUnusable BroadcastReasonKind = BroadcastReasonKind(native.BroadcastReasonPreciseDegradedUnusableValue)
)

// FixSourceKind identifies the ephemeris source of a fallback result.
type FixSourceKind uint32

const (
	FixSourcePrecise   FixSourceKind = FixSourceKind(native.FixSourcePreciseValue)
	FixSourceBroadcast FixSourceKind = FixSourceKind(native.FixSourceBroadcastValue)
)

// DegradationKind describes native product staleness degradation.
type DegradationKind uint32

const (
	DegradationExact        DegradationKind = DegradationKind(native.DegradationExactValue)
	DegradationNearestPrior DegradationKind = DegradationKind(native.DegradationNearestPriorValue)
	DegradationDiurnalShift DegradationKind = DegradationKind(native.DegradationDiurnalShiftValue)
)

// SelectionStatus is the native precise-product selection result.
type SelectionStatus uint32

const (
	SelectionOK                 SelectionStatus = 0
	SelectionNullPointer        SelectionStatus = 1
	SelectionInvalidArgument    SelectionStatus = 2
	SelectionInvalidToken       SelectionStatus = 3
	SelectionPanic              SelectionStatus = 4
	SelectionEmptyProductSet    SelectionStatus = 5
	SelectionInvalidRange       SelectionStatus = 6
	SelectionNoPriorProduct     SelectionStatus = 7
	SelectionBeyondStalenessCap SelectionStatus = 8
	SelectionInvalidProduct     SelectionStatus = 9
	SelectionInvalidPolicy      SelectionStatus = 10
	SelectionOverflow           SelectionStatus = 11
)

// StalenessMetadata is detached fallback provenance.
type StalenessMetadata struct {
	Kind                                                               DegradationKind
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
	HasMaxPDOP                                                                  bool
	MaxPDOP, MinPlausibleRadiusM, MaxPlausibleRadiusM, MaxConvergedResidualRMSM float64
}

// SolutionValidationOptionsInit returns native receiver-validation defaults.
func SolutionValidationOptionsInit() (SolutionValidationOptions, error) {
	value, err := native.SolutionValidationOptionsInit()
	return SolutionValidationOptions{HasMaxPDOP: value.HasMaxPDOP, MaxPDOP: value.MaxPDOP, MinPlausibleRadiusM: value.MinPlausibleRadiusM, MaxPlausibleRadiusM: value.MaxPlausibleRadiusM, MaxConvergedResidualRMSM: value.MaxConvergedResidualRMSM}, publicError(err)
}
