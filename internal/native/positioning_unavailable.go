//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

import "time"

type SP3State struct{}

type SP3PredictionSummary struct{}

type SP3 struct{}

func LoadSP3([]byte) (*SP3, error) { return nil, unavailable() }
func (*SP3) Close() error          { return nil }
func (*SP3) EpochCount() (int, error) {
	return 0, unavailable()
}
func (*SP3) Epochs() ([]float64, error)          { return nil, unavailable() }
func (*SP3) Satellites() ([]string, error)       { return nil, unavailable() }
func (*SP3) State(string, int) (SP3State, error) { return SP3State{}, unavailable() }
func (*SP3) PredictionSummary() (SP3PredictionSummary, error) {
	return SP3PredictionSummary{}, unavailable()
}

type SPPObservation struct {
	SatelliteID  string
	PseudorangeM float64
}

type SPPConfig struct {
	Observations    []SPPObservation
	TRxJ2000S       float64
	TRxSecondOfDayS float64
	DayOfYear       float64
	InitialGuess    [4]float64
	Ionosphere      bool
	Troposphere     bool
	WithGeodetic    bool
}

type SPPGeometryQuality struct{}
type SPPMetadata struct{}
type SPPSolution struct{}

type StaticPositionEpochInput struct {
	Inputs  SppInputsV2
	Weights []float64
}
type NativeSolutionValidationOptions struct {
	HasMaxPDOP                                                                  bool
	MaxPDOP, MinPlausibleRadiusM, MaxPlausibleRadiusM, MaxConvergedResidualRMSM float64
}
type NativeGlonassChannel struct {
	Slot    uint8
	Channel int8
}
type NativeSPPValidationOptions struct {
	MaxPDOPEnabled                                                              bool
	MaxPDOP, MinPlausibleRadiusM, MaxPlausibleRadiusM, MaxConvergedResidualRMSM float64
}
type NativeSPPSolvePolicy struct {
	UseValidationOptions bool
	Validation           NativeSPPValidationOptions
	CoarseSearchEnabled  bool
	CoarseSearchSeeds    int
}
type SppInputsV2 struct {
	Base                    SPPConfig
	BeidouEnabled           bool
	BeidouAlpha, BeidouBeta [4]float64
	RobustEnabled           bool
	Robust                  NativeSPPRobustConfig
	Policy                  NativeSPPSolvePolicy
	GlonassChannels         []NativeGlonassChannel
}
type NativeSPPRobustConfig struct {
	HuberK, ScaleFloorM float64
	MaxOuter            uint64
	OuterToleranceM     float64
}
type StaticPositionOptionsInput struct {
	InitialPositionM [3]float64
	WithGeodetic     bool
	RobustEnabled    bool
	Robust           NativeSPPRobustConfig
}
type RtkRinexStaticBaselineConfig struct{}
type StaticReferenceStationRinexConfigInput struct {
	ReferencePositionM [3]float64
	EnableCodeDGNSS    bool
	EnableCarrierRTK   bool
	WithGeodetic       bool
	Carrier            RtkRinexStaticBaselineConfig
}
type StaticPositionClockBias struct{}
type StaticPositionEpochInfluence struct{}
type StaticPositionMetadata struct{}
type StaticPositionResidual struct{}
type StaticPositionRejectedSat struct{}
type StaticPositionSatelliteBatchInfluence struct{}
type StaticPositionSatelliteInfluence struct{}
type StaticReferenceEpochDiagnostic struct{}
type StaticReferenceStationMetadata struct{}
type StaticReferenceModeReport struct{}
type StaticPositionSolution struct{}
type StaticReferenceStationSolution struct{}

func StaticPositionOptionsInit() (StaticPositionOptionsInput, error) {
	return StaticPositionOptionsInput{}, unavailable()
}
func SolveStaticPositionBroadcast(*BroadcastEphemeris, []StaticPositionEpochInput, *StaticPositionOptionsInput) (*StaticPositionSolution, uint32, error) {
	return nil, 0, unavailable()
}
func SolveStaticPositionSP3(*SP3, []StaticPositionEpochInput, *StaticPositionOptionsInput) (*StaticPositionSolution, uint32, error) {
	return nil, 0, unavailable()
}
func (s *StaticPositionSolution) Close() error                  { return nil }
func (s *StaticPositionSolution) Position() ([3]float64, error) { return [3]float64{}, unavailable() }
func (s *StaticPositionSolution) PositionCovarianceECEFM2() ([9]float64, error) {
	return [9]float64{}, unavailable()
}
func (s *StaticPositionSolution) PositionCovarianceENUM2() ([9]float64, error) {
	return [9]float64{}, unavailable()
}
func (s *StaticPositionSolution) ClockBiases() ([]StaticPositionClockBias, error) {
	return nil, unavailable()
}
func (s *StaticPositionSolution) EpochInfluence() ([]StaticPositionEpochInfluence, error) {
	return nil, unavailable()
}
func (s *StaticPositionSolution) Geodetic() (Geodetic, bool, error) {
	return Geodetic{}, false, unavailable()
}
func (s *StaticPositionSolution) Metadata() (StaticPositionMetadata, error) {
	return StaticPositionMetadata{}, unavailable()
}
func (s *StaticPositionSolution) RejectedSats(int) ([]StaticPositionRejectedSat, error) {
	return nil, unavailable()
}
func (s *StaticPositionSolution) Residuals() ([]StaticPositionResidual, error) {
	return nil, unavailable()
}
func (s *StaticPositionSolution) SatelliteBatchInfluence() ([]StaticPositionSatelliteBatchInfluence, error) {
	return nil, unavailable()
}
func (s *StaticPositionSolution) SatelliteInfluence() ([]StaticPositionSatelliteInfluence, error) {
	return nil, unavailable()
}
func (s *StaticPositionSolution) StateCovarianceM2() ([]float64, error) { return nil, unavailable() }
func StaticReferenceStationRinexConfigInit() (StaticReferenceStationRinexConfigInput, error) {
	return StaticReferenceStationRinexConfigInput{}, unavailable()
}
func SolveStaticReferenceStationRinex(*SP3, *RinexObs, *RinexObs, StaticReferenceStationRinexConfigInput) (*StaticReferenceStationSolution, error) {
	return nil, unavailable()
}
func (s *StaticReferenceStationSolution) Close() error { return nil }
func (s *StaticReferenceStationSolution) BaselineECEF() ([3]float64, error) {
	return [3]float64{}, unavailable()
}
func (s *StaticReferenceStationSolution) PositionECEF() ([3]float64, error) {
	return [3]float64{}, unavailable()
}
func (s *StaticReferenceStationSolution) CovarianceECEF() ([9]float64, error) {
	return [9]float64{}, unavailable()
}
func (s *StaticReferenceStationSolution) CovarianceENU() ([9]float64, error) {
	return [9]float64{}, unavailable()
}
func (s *StaticReferenceStationSolution) Diagnostics() ([]StaticReferenceEpochDiagnostic, error) {
	return nil, unavailable()
}
func (s *StaticReferenceStationSolution) Metadata() (StaticReferenceStationMetadata, error) {
	return StaticReferenceStationMetadata{}, unavailable()
}
func (s *StaticReferenceStationSolution) ModeReports() ([]StaticReferenceModeReport, error) {
	return nil, unavailable()
}

func (*SP3) Solve(SPPConfig) (SPPSolution, error) { return SPPSolution{}, unavailable() }

type TLEMetadata struct{}
type TLELines struct{}
type TEMEState struct{}
type TLE struct{}

const (
	TLEOpsModeAFSPCValue    uint32 = 0
	TLEOpsModeImprovedValue uint32 = 1
)

func ParseTLE(string, string) (*TLE, error)        { return nil, unavailable() }
func LoadTLE(string, string, uint32) (*TLE, error) { return nil, unavailable() }
func (*TLE) Close() error                          { return nil }
func (*TLE) Metadata() (TLEMetadata, error)        { return TLEMetadata{}, unavailable() }
func (*TLE) Lines() (TLELines, error)              { return TLELines{}, unavailable() }
func (*TLE) Propagate([]time.Time) ([]TEMEState, error) {
	return nil, unavailable()
}

type GroundStation struct {
	LatitudeDeg  float64
	LongitudeDeg float64
	AltitudeM    float64
}

type SatellitePass struct {
	AOS, LOS, Culmination time.Time
	MaxElevationDeg       float64
	DurationS             float64
}

type PassFinderOptions struct {
	ElevationMaskDeg     float64
	StepSeconds          float64
	TimeToleranceSeconds float64
}

type PassList struct{}
type GroundTrack struct{}
type LookAngles struct{}

func PassFinderDefaults() (PassFinderOptions, error) {
	return PassFinderOptions{}, unavailable()
}
func (*TLE) FindPasses(GroundStation, time.Time, time.Time, *PassFinderOptions) (*PassList, error) {
	return nil, unavailable()
}
func (*TLE) GroundTrack([]time.Time) (*GroundTrack, error) {
	return nil, unavailable()
}
func (*TLE) LookAngles(GroundStation, []time.Time) (*LookAngles, error) {
	return nil, unavailable()
}
func (*PassList) Close() error                     { return nil }
func (*PassList) Count() (int, error)              { return 0, unavailable() }
func (*PassList) Values() ([]SatellitePass, error) { return nil, unavailable() }
func (*GroundTrack) Close() error                  { return nil }
func (*GroundTrack) Count() (int, error)           { return 0, unavailable() }
func (*GroundTrack) Values() ([]Geodetic, error)   { return nil, unavailable() }
func (*LookAngles) Close() error                   { return nil }
func (*LookAngles) Count() (int, error)            { return 0, unavailable() }
func (*LookAngles) Values() ([]LookAngle, error)   { return nil, unavailable() }
