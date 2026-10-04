//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

const (
	SourceSolveTOAValue                         = uint32(0)
	SourceSolveTDOAValue                        = uint32(1)
	SourceLossLinearValue                       = uint32(0)
	SourceLossSoftL1Value                       = uint32(1)
	SourceLossHuberValue                        = uint32(2)
	SourceLossCauchyValue                       = uint32(3)
	SourceLossArctanValue                       = uint32(4)
	BroadcastReasonPreciseUnavailableValue      = uint32(0)
	BroadcastReasonPreciseDegradedUnusableValue = uint32(1)
	FixSourcePreciseValue                       = uint32(0)
	FixSourceBroadcastValue                     = uint32(1)
	DegradationExactValue                       = uint32(0)
	DegradationNearestPriorValue                = uint32(1)
	DegradationDiurnalShiftValue                = uint32(2)
)

type NativeSourceSensor struct {
	Dimension           int
	PositionM           [3]float64
	HasPropagationSpeed bool
	PropagationSpeedMS  float64
}

type NativeSourceInitialGuess struct {
	Dimension     int
	PositionM     [3]float64
	HasOriginTime bool
	OriginTimeS   float64
	ResidualRMSS  float64
}

type NativeSourceLocateOptions struct {
	Mode, ReferenceSensor int
	TimingSigmaS, FScaleS float64
	Loss                  uint32
	HasFTOL, HasXTOL      bool
	FTOL, XTOL            float64
	HasGTOL               bool
	GTOL                  float64
	HasMaxNFEV            bool
	MaxNFEV               int
}

type NativeSourceCovariance struct {
	Dimension, StateDimension  int
	State                      [16]float64
	PositionM2                 [9]float64
	HasOriginTimeS2            bool
	OriginTimeS2, TimingSigmaS float64
}

type NativeSourceCrlb struct {
	DOP        Dop
	Covariance NativeSourceCovariance
}

type NativeSourceSensorInfluence struct {
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

type NativeSourceResidual struct {
	SensorIndex, ReferenceSensorIndex int
	HasReferenceSensor                bool
	ResidualS                         float64
}

type NativeSourceSolutionSummary struct {
	Dimension, ResidualCount, InfluenceCount int
	PositionM                                [3]float64
	HasOriginTime                            bool
	OriginTimeS                              float64
	HasCovariance                            bool
	GeometryQuality                          GeometryQuality
	InitialGuess                             NativeSourceInitialGuess
	Status                                   int32
	NFEV, NJEV                               int
	Cost, Optimality                         float64
}

type GeometryQuality struct {
	Tier                               uint32
	Redundancy                         int32
	Rank                               int
	ConditionNumber, GDOP              float64
	RAIMCheckable, CovarianceValidated bool
}

type NativePseudorangeVarianceOptions struct {
	AM, BM              float64
	Model               uint32
	HasCN0              bool
	CN0DBHz, CN0ScaleM2 float64
}

type SourceSolution struct{}
type SourcedSolution struct{}

type NativeCnavParameters struct {
	Present                    bool
	ADOTMS, DeltaN0DotRadS2    float64
	TopWeek                    uint32
	TopTOWS                    float64
	URAEDIndex, URANED0Index   int8
	URANED1Index, URANED2Index uint8
	TransmissionTimeSOW        float64
	HasFlags                   bool
	Flags                      uint32
}

func ChanHOInitialGuess([]NativeSourceSensor, []float64, float64, uint32, int) (NativeSourceInitialGuess, error) {
	return NativeSourceInitialGuess{}, unavailable()
}
func ClosedFormInitialGuess([]NativeSourceSensor, []float64, float64, uint32, int) (NativeSourceInitialGuess, error) {
	return NativeSourceInitialGuess{}, unavailable()
}
func SourceLocateOptionsInit() (NativeSourceLocateOptions, error) {
	return NativeSourceLocateOptions{}, unavailable()
}
func LocateSource([]NativeSourceSensor, []float64, float64, *NativeSourceLocateOptions) (*SourceSolution, error) {
	return nil, unavailable()
}
func LocateSourceWith([]NativeSourceSensor, []float64, float64, *NativeSourceLocateOptions, bool) (*SourceSolution, error) {
	return nil, unavailable()
}
func (*SourceSolution) Close() error { return nil }
func (*SourceSolution) Covariance() (NativeSourceCovariance, bool, error) {
	return NativeSourceCovariance{}, false, unavailable()
}
func (*SourceSolution) Influences() ([]NativeSourceSensorInfluence, error) {
	return nil, unavailable()
}
func (*SourceSolution) Residuals() ([]NativeSourceResidual, error) { return nil, unavailable() }
func (*SourceSolution) Summary() (NativeSourceSolutionSummary, error) {
	return NativeSourceSolutionSummary{}, unavailable()
}
func (*SourcedSolution) Close() error { return nil }
func (*SourcedSolution) BroadcastReason() (uint32, uint32, NativeStalenessMetadata, bool, error) {
	return 0, 0, NativeStalenessMetadata{}, false, unavailable()
}
func (*SourcedSolution) BroadcastReasonDetail() ([]byte, error) { return nil, unavailable() }
func (*SourcedSolution) IsPreciseExact() (bool, error)          { return false, unavailable() }
func (*SourcedSolution) SourceKind() (uint32, error)            { return 0, unavailable() }
func (*SourcedSolution) Staleness() (NativeStalenessMetadata, bool, error) {
	return NativeStalenessMetadata{}, false, unavailable()
}
func (*SourcedSolution) Solution() (SPPSolution, error) { return SPPSolution{}, unavailable() }
func SourceCRLB([]NativeSourceSensor, []float64, float64, float64) (NativeSourceCrlb, error) {
	return NativeSourceCrlb{}, unavailable()
}
func SourceDOP([]NativeSourceSensor, []float64, float64) (Dop, error) { return Dop{}, unavailable() }
func CNAVURANEDM(NativeCnavParameters, uint32, float64) (float64, bool, error) {
	return 0, false, unavailable()
}
func CNAVURANominalM(int8) (float64, bool, error) { return 0, false, unavailable() }
func PseudorangeVarianceOptionsInit() (NativePseudorangeVarianceOptions, error) {
	return NativePseudorangeVarianceOptions{}, unavailable()
}
func PseudorangeVariance(float64, NativePseudorangeVarianceOptions) (float64, error) {
	return 0, unavailable()
}
func SolutionValidationOptionsInit() (NativeSolutionValidationOptions, error) {
	return NativeSolutionValidationOptions{}, unavailable()
}
