//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

const (
	PreciseSamplesErrorNone                    = uint32(0)
	PreciseSamplesErrorEmpty                   = uint32(1)
	PreciseSamplesErrorSingleSampleSatellite   = uint32(2)
	PreciseSamplesErrorNonMonotonicEpochs      = uint32(3)
	PreciseSamplesErrorMixedTimeScales         = uint32(4)
	PreciseSamplesErrorEpochNotRepresentable   = uint32(5)
	PreciseSamplesErrorNonFiniteSample         = uint32(6)
	PreciseSamplesErrorAccuracySamplesMismatch = uint32(7)
	PreciseSamplesErrorInvalidAccuracyValue    = uint32(8)
	PreciseSamplesErrorOther                   = uint32(9)
)

type SP3AccuracyValue struct {
	Kind  uint32
	Value float64
}

type SP3AccuracyCodeGroup struct {
	HasAxisExponents        [3]bool
	AxisExponents           [3]int16
	HasClockExponent        bool
	ClockExponent           int16
	HasPositionVelocityBase bool
	PositionVelocityBase    float64
	HasClockRateBase        bool
	ClockRateBase           float64
}

type SP3RawRecordAccuracy struct {
	HasP bool
	P    SP3AccuracyCodeGroup
	HasV bool
	V    SP3AccuracyCodeGroup
}

type SP3PositionClockAccuracy struct {
	PositionSigmaM     [3]SP3AccuracyValue
	ClockSigmaM        SP3AccuracyValue
	PositionVarianceM2 [3]SP3AccuracyValue
	ClockVarianceM2    SP3AccuracyValue
}

type SP3VelocityAccuracy struct {
	VelocitySigmaMPerS       [3]SP3AccuracyValue
	ClockRateSigmaMPerS      SP3AccuracyValue
	VelocityVarianceM2PerS2  [3]SP3AccuracyValue
	ClockRateVarianceM2PerS2 SP3AccuracyValue
}

type SP3RecordAccuracy struct {
	HasP bool
	P    SP3PositionClockAccuracy
	HasV bool
	V    SP3VelocityAccuracy
}

type PreciseEphemerisAccuracySample struct {
	Satellite          string
	TimeScale          uint32
	EpochJ2000S        float64
	PositionVarianceM2 [3]SP3AccuracyValue
	ClockVarianceM2    SP3AccuracyValue
}

type PreciseSamplesError struct {
	Kind         uint32
	HasSatellite bool
	Satellite    string
}

type PreciseEphemerisSample struct {
	Satellite     string
	TimeScale     uint32
	EpochJ2000S   float64
	PositionECEFM [3]float64
	HasClock      bool
	ClockS        float64
	ClockEvent    bool
}

type PreciseEphemerisSampleV2 struct {
	Satellite     string
	Epoch         NativeClockEpoch
	PositionECEFM [3]float64
	HasClock      bool
	ClockS        float64
	ClockEvent    bool
}

type EphemerisSourceState struct {
	HasState      bool
	PositionECEFM [3]float64
	ClockS        float64
	HasGroupDelay bool
	GroupDelayS   float64
	Degraded      bool
}

type TransmitEpochClock struct {
	HasClock bool
	ClockS   float64
	Degraded bool
}

type ClockRelativity struct {
	Kind  uint32
	TermS float64
}

type PreciseEphemerisAccuracySampleV2 struct {
	Satellite          string
	Epoch              NativeClockEpoch
	PositionVarianceM2 [3]SP3AccuracyValue
	ClockVarianceM2    SP3AccuracyValue
}

type PreciseEphemerisSamples struct{}
type PreciseEphemerisInterpolant struct{}
type PreciseInterpolantArtifact struct{}

func (*PreciseEphemerisSamples) Close() error     { return nil }
func (*PreciseEphemerisInterpolant) Close() error { return nil }
func (*PreciseInterpolantArtifact) Close() error  { return nil }

type SSRCorrectionSize struct {
	OrbitM float64
	ClockM float64
}

type SSRCorrectedState struct {
	HasState                bool
	PositionECEFM           [3]float64
	ClockS                  float64
	HasGroupDelay           bool
	GroupDelayS             float64
	Degraded                bool
	DegradeReason           uint32
	HasSizeEvent            bool
	StrictRefusal           bool
	Size                    SSRCorrectionSize
	HasOversizedReport      bool
	Source                  uint32
	ProviderID              uint16
	SolutionID              uint8
	OrbitRefEpochJ2000S     float64
	ClockRefEpochJ2000S     float64
	FirstAppliedEpochJ2000S float64
}

func (*SP3) StateAtEpochQuery(*ExactEpochQuery, string) (SP3State, error) {
	return SP3State{}, unavailable()
}
func (*SP3) SourceStateAtEpochQueries(*ExactEpochQuery, *ExactEpochQuery, string) (EphemerisSourceState, error) {
	return EphemerisSourceState{}, unavailable()
}
func (*SP3) TransmitEpochClockAtEpochQueries(*ExactEpochQuery, *ExactEpochQuery, string) (TransmitEpochClock, error) {
	return TransmitEpochClock{}, unavailable()
}
func (*SP3) ClockRelativityAtEpochQuery(*ExactEpochQuery, string, [3]float64) (ClockRelativity, error) {
	return ClockRelativity{}, unavailable()
}
func (*SP3) EphemerisVarianceAtEpochQueries(*ExactEpochQuery, *ExactEpochQuery, string) (float64, error) {
	return 0, unavailable()
}
func (*SP3) RecordAccuracyCodes(string, int) (SP3RawRecordAccuracy, error) {
	return SP3RawRecordAccuracy{}, unavailable()
}
func (*SP3) RecordAccuracy(string, int) (SP3RecordAccuracy, error) {
	return SP3RecordAccuracy{}, unavailable()
}
func (*SP3) PreciseAccuracySamples() ([]PreciseEphemerisAccuracySample, error) {
	return nil, unavailable()
}
func (*SP3) PreciseSamplesV2() ([]PreciseEphemerisSampleV2, error) {
	return nil, unavailable()
}
func (*SP3) PreciseAccuracySamplesV2() ([]PreciseEphemerisAccuracySampleV2, error) {
	return nil, unavailable()
}
func (*PreciseEphemerisSamples) RecordsV2() ([]PreciseEphemerisSampleV2, error) {
	return nil, unavailable()
}
func (*PreciseEphemerisSamples) AccuracyRecordsV2() ([]PreciseEphemerisAccuracySampleV2, error) {
	return nil, unavailable()
}
func PreciseEphemerisSamplesFromSamplesWithAccuracy([]PreciseEphemerisSample, []PreciseEphemerisAccuracySample, float64) (*PreciseEphemerisSamples, PreciseSamplesError, error) {
	return nil, PreciseSamplesError{}, unavailable()
}
func PreciseEphemerisInterpolantFromSamplesWithAccuracy([]PreciseEphemerisSample, []PreciseEphemerisAccuracySample, float64) (*PreciseEphemerisInterpolant, PreciseSamplesError, error) {
	return nil, PreciseSamplesError{}, unavailable()
}
func PreciseEphemerisSamplesFromSamplesV2([]PreciseEphemerisSampleV2, float64) (*PreciseEphemerisSamples, PreciseSamplesError, error) {
	return nil, PreciseSamplesError{}, unavailable()
}
func PreciseEphemerisInterpolantFromSamplesV2([]PreciseEphemerisSampleV2, float64) (*PreciseEphemerisInterpolant, PreciseSamplesError, error) {
	return nil, PreciseSamplesError{}, unavailable()
}
func PreciseEphemerisSamplesFromSamplesWithAccuracyV2([]PreciseEphemerisSampleV2, []PreciseEphemerisAccuracySampleV2, float64) (*PreciseEphemerisSamples, PreciseSamplesError, error) {
	return nil, PreciseSamplesError{}, unavailable()
}
func PreciseEphemerisInterpolantFromSamplesWithAccuracyV2([]PreciseEphemerisSampleV2, []PreciseEphemerisAccuracySampleV2, float64) (*PreciseEphemerisInterpolant, PreciseSamplesError, error) {
	return nil, PreciseSamplesError{}, unavailable()
}
func (*PreciseEphemerisInterpolant) StateAtEpochQuery(*ExactEpochQuery, string) (SP3State, error) {
	return SP3State{}, unavailable()
}
func (*PreciseEphemerisInterpolant) SourceStateAtEpochQueries(*ExactEpochQuery, *ExactEpochQuery, string) (EphemerisSourceState, error) {
	return EphemerisSourceState{}, unavailable()
}
func (*PreciseEphemerisInterpolant) TransmitEpochClockAtEpochQueries(*ExactEpochQuery, *ExactEpochQuery, string) (TransmitEpochClock, error) {
	return TransmitEpochClock{}, unavailable()
}
func (*PreciseEphemerisInterpolant) ClockRelativityAtEpochQuery(*ExactEpochQuery, string, [3]float64) (ClockRelativity, error) {
	return ClockRelativity{}, unavailable()
}
func (*PreciseEphemerisInterpolant) EphemerisVarianceAtEpochQueries(*ExactEpochQuery, *ExactEpochQuery, string) (float64, error) {
	return 0, unavailable()
}
func (*PreciseInterpolantArtifact) StateAtEpochQuery(*ExactEpochQuery, string) (SP3State, error) {
	return SP3State{}, unavailable()
}
func (*PreciseInterpolantArtifact) SourceStateAtEpochQueries(*ExactEpochQuery, *ExactEpochQuery, string) (EphemerisSourceState, error) {
	return EphemerisSourceState{}, unavailable()
}
func (*PreciseInterpolantArtifact) TransmitEpochClockAtEpochQueries(*ExactEpochQuery, *ExactEpochQuery, string) (TransmitEpochClock, error) {
	return TransmitEpochClock{}, unavailable()
}
func (*PreciseInterpolantArtifact) ClockRelativityAtEpochQuery(*ExactEpochQuery, string, [3]float64) (ClockRelativity, error) {
	return ClockRelativity{}, unavailable()
}
func (*PreciseInterpolantArtifact) EphemerisVarianceAtEpochQueries(*ExactEpochQuery, *ExactEpochQuery, string) (float64, error) {
	return 0, unavailable()
}
func CorrectedStateAtEpochQueries(*BroadcastEphemeris, *SSRCorrectionStore, string, *ExactEpochQuery, *ExactEpochQuery, float64, uint32, bool, uint16, uint32) (SSRCorrectedState, error) {
	return SSRCorrectedState{}, unavailable()
}
