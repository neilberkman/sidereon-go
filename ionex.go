package sidereon

import (
	"errors"

	"sidereon.dev/go/v3/internal/native"
)

// TECSample is one native IONEX TEC sample.
type TECSample struct {
	TimeScale                             TimeScale
	EpochJ2000S, LatDeg, LonDeg, VTECTECU float64
	RMSPresent                            bool
	RMSTECU                               float64
}

// TECGridSamples describes a complete IONEX TEC grid.
type TECGridSamples struct {
	TimeScale                                     TimeScale
	MapEpochsJ2000S, LatNodesDeg, LonNodesDeg     []float64
	DLatDeg, DLonDeg, ShellHeightKm, BaseRadiusKm float64
	Exponent                                      int32
	TECMAPsTECU                                   []float64
	RMSPresent                                    bool
	RMSMAPsTECU                                   []float64
}

// TECGridSamplesInfo reports IONEX grid dimensions and metadata.
type TECGridSamplesInfo struct {
	MapEpochCount, LatNodeCount, LonNodeCount     int
	DLatDeg, DLonDeg, ShellHeightKm, BaseRadiusKm float64
	Exponent                                      int32
	RMSPresent                                    bool
	TECMAPValueCount, RMSMAPValueCount            int
}

// IONEXSlantDelayEvaluation contains a delay and its coverage status.
type IONEXSlantDelayEvaluation struct {
	DelayM                                           float64
	Status                                           IONEXSlantDelayStatus
	CoverageError                                    IONEXCoverageErrorKind
	IsValid, HasHeld, HasDegraded, HasAssumedMapping bool
	Gap                                              IONEXNodeGap
	AssumedMapping                                   IONEXAssumedMappingKind
}

// IONEXNodeGap preserves all weighted missing-node masks from both maps.
type IONEXNodeGap struct {
	HasGap         bool
	Earlier, Later [4]bool
}

// IONEXEpochErrorKind identifies why an exact instant cannot be mapped to UTC.
type IONEXEpochErrorKind uint32

const (
	IONEXEpochErrorNone                     IONEXEpochErrorKind = 0
	IONEXEpochErrorNotWholeSecond           IONEXEpochErrorKind = 1
	IONEXEpochErrorFractionalUTCSecond      IONEXEpochErrorKind = 2
	IONEXEpochErrorNoExactUTCOffset         IONEXEpochErrorKind = 3
	IONEXEpochErrorInsertedLeapSecond       IONEXEpochErrorKind = 4
	IONEXEpochErrorBeforeIntegerLeapSeconds IONEXEpochErrorKind = 5
	IONEXEpochErrorOutOfRange               IONEXEpochErrorKind = 6
	IONEXEpochErrorYearOutOfField           IONEXEpochErrorKind = 7
	IONEXEpochErrorUnknown                  IONEXEpochErrorKind = 999
)

// IONEXEpochError retains the conversion cause independently of other row errors.
type IONEXEpochError struct {
	Kind         IONEXEpochErrorKind
	Scale        TimeScale
	HasUTCJ2000S bool
	UTCJ2000S    int64
}

// IONEXSlantPolicy controls coverage, missing-node, and mapping behavior.
type IONEXSlantPolicy struct {
	Coverage     IONEXCoveragePolicy
	MissingNodes IONEXMissingNodePolicy
	Mapping      IONEXMappingPolicy
}
type IONEXMissingNodePolicy uint32

const (
	IONEXMissingNodesStrict      IONEXMissingNodePolicy = 0
	IONEXMissingNodesRenormalize IONEXMissingNodePolicy = 1
)

type IONEXMappingPolicy uint32

const (
	IONEXMappingDeclared    IONEXMappingPolicy = 0
	IONEXMappingSingleLayer IONEXMappingPolicy = 1
)

type IONEXAssumedMappingKind uint32

const (
	IONEXAssumedMappingNone   IONEXAssumedMappingKind = 0
	IONEXAssumedMappingOther  IONEXAssumedMappingKind = 1
	IONEXAssumedMappingAbsent IONEXAssumedMappingKind = 2
)

// IONEXSlantErrorKind identifies the typed refusal for one slant query.
type IONEXSlantErrorKind uint32

const (
	IONEXSlantErrorNone              IONEXSlantErrorKind = 0
	IONEXSlantErrorInvalidInput      IONEXSlantErrorKind = 1
	IONEXSlantErrorOutOfCoverage     IONEXSlantErrorKind = 2
	IONEXSlantErrorNodesNotAvailable IONEXSlantErrorKind = 3
	IONEXSlantErrorUnavailable       IONEXSlantErrorKind = 4
	IONEXSlantErrorUnknown           IONEXSlantErrorKind = 999
)

// IONEXSlantError retains every typed cause field returned by the C ABI.
type IONEXSlantError struct {
	Kind                                               IONEXSlantErrorKind
	CoverageError                                      IONEXCoverageErrorKind
	HasGap                                             bool
	Gap                                                IONEXNodeGap
	Refusal                                            uint32
	RefusalMapNumber, RefusalLatIndex, RefusalLonIndex uint64
	HasMappingDeclaration                              bool
	MappingDeclaration                                 uint32
	HasMappingFunction                                 bool
	MappingFunction                                    uint32
}

// IONEXInstantSlantRequest contains one slant query at a lossless scale-tagged instant.
type IONEXInstantSlantRequest struct {
	LatDeg, LonDeg, AzimuthDeg, ElevationDeg float64
	Epoch                                    ClockEpoch
	FrequencyHz                              float64
}

// IONEXInstantSlantResult retains either the value or full typed row failure.
type IONEXInstantSlantResult struct {
	IsOK        bool
	Status      StatusCode
	Evaluation  IONEXSlantDelayEvaluation
	Error       *IONEXSlantError
	EpochError  *IONEXEpochError
	Message     string
	MappingCode string
}

// IONEXInstantSlantResultList owns detached, input-ordered exact-time results.
type IONEXInstantSlantResultList struct{ Rows []IONEXInstantSlantResult }

// IONEX owns a parsed or synthesized IONEX product.
type IONEX struct {
	_      noCopy
	handle *native.Ionex
}

func publicTECSample(v native.TecSample) TECSample {
	return TECSample{TimeScale: TimeScale(v.TimeScale), EpochJ2000S: v.EpochJ2000S, LatDeg: v.LatDeg, LonDeg: v.LonDeg, VTECTECU: v.VTECTECU, RMSPresent: v.RMSPresent, RMSTECU: v.RMSTECU}
}

// ParseIONEX parses an IONEX product with the native parser.
func ParseIONEX(data []byte) (*IONEX, error) {
	v, e := native.ParseIONEX(data)
	if e != nil {
		return nil, publicError(e)
	}
	if v == nil {
		return nil, errNilNativeHandle
	}
	return &IONEX{handle: v}, nil
}

// NewIONEXFromTECGridSamples builds an IONEX product from a complete grid.
func NewIONEXFromTECGridSamples(v TECGridSamples) (*IONEX, error) {
	x := native.TecGridSamples{TimeScale: uint32(v.TimeScale), MapEpochsJ2000S: append([]float64(nil), v.MapEpochsJ2000S...), LatNodesDeg: append([]float64(nil), v.LatNodesDeg...), LonNodesDeg: append([]float64(nil), v.LonNodesDeg...), DLatDeg: v.DLatDeg, DLonDeg: v.DLonDeg, ShellHeightKm: v.ShellHeightKm, BaseRadiusKm: v.BaseRadiusKm, Exponent: v.Exponent, TECMAPsTECU: append([]float64(nil), v.TECMAPsTECU...), RMSPresent: v.RMSPresent, RMSMAPsTECU: append([]float64(nil), v.RMSMAPsTECU...)}
	out, e := native.BuildIONEXFromTECGridSamples(x)
	if e != nil {
		return nil, publicError(e)
	}
	return &IONEX{handle: out}, nil
}

// NewIONEXFromTECSamples builds an IONEX product from grid-node samples.
func NewIONEXFromTECSamples(v []TECSample, shellHeightKm, baseRadiusKm float64, exponent int32) (*IONEX, error) {
	x := make([]native.TecSample, len(v))
	for i, s := range v {
		x[i] = native.TecSample{TimeScale: uint32(s.TimeScale), EpochJ2000S: s.EpochJ2000S, LatDeg: s.LatDeg, LonDeg: s.LonDeg, VTECTECU: s.VTECTECU, RMSPresent: s.RMSPresent, RMSTECU: s.RMSTECU}
	}
	out, e := native.BuildIONEXFromTECSamples(x, shellHeightKm, baseRadiusKm, exponent)
	if e != nil {
		return nil, publicError(e)
	}
	return &IONEX{handle: out}, nil
}

// Close releases the IONEX product; it is safe to call more than once.
func (i *IONEX) Close() error {
	if i == nil || i.handle == nil {
		return nil
	}
	return publicError(i.handle.Close())
}

// EpochCount returns the number of TEC maps.
func (i *IONEX) EpochCount() (int, error) {
	if i == nil || i.handle == nil {
		return 0, ErrClosed
	}
	v, e := i.handle.EpochCount()
	return v, publicError(e)
}

// Exponent returns the IONEX TEC exponent.
func (i *IONEX) Exponent() (int32, error) {
	if i == nil || i.handle == nil {
		return 0, ErrClosed
	}
	v, e := i.handle.Exponent()
	return v, publicError(e)
}

// LatNodesDeg returns the latitude node axis in degrees.
func (i *IONEX) LatNodesDeg() ([]float64, error) {
	if i == nil || i.handle == nil {
		return nil, ErrClosed
	}
	v, e := i.handle.LatNodesDeg()
	return v, publicError(e)
}

// LonNodesDeg returns the longitude node axis in degrees.
func (i *IONEX) LonNodesDeg() ([]float64, error) {
	if i == nil || i.handle == nil {
		return nil, ErrClosed
	}
	v, e := i.handle.LonNodesDeg()
	return v, publicError(e)
}

// MapEpochsJ2000S returns map epochs as integer J2000 seconds.
func (i *IONEX) MapEpochsJ2000S() ([]int64, error) {
	if i == nil || i.handle == nil {
		return nil, ErrClosed
	}
	v, e := i.handle.MapEpochsJ2000S()
	return v, publicError(e)
}

// ToIONEXText serializes the product using the native IONEX writer.
func (i *IONEX) ToIONEXText() ([]byte, error) {
	if i == nil || i.handle == nil {
		return nil, ErrClosed
	}
	v, e := i.handle.ToIONEXText()
	return v, publicError(e)
}

// SlantDelay evaluates native IONEX slant delay with the legacy hold policy.
func (i *IONEX) SlantDelay(lat, lon, azimuth, elevation float64, epochJ2000S int64, frequencyHz float64) (float64, error) {
	if i == nil || i.handle == nil {
		return 0, ErrClosed
	}
	v, e := i.handle.SlantDelay(lat, lon, azimuth, elevation, epochJ2000S, frequencyHz)
	return v, publicError(e)
}

// SlantDelayWithPolicy evaluates slant delay with an explicit coverage policy.
func (i *IONEX) SlantDelayWithPolicy(lat, lon, azimuth, elevation float64, epochJ2000S int64, frequencyHz float64, policy IONEXCoveragePolicy) (IONEXSlantDelayEvaluation, error) {
	if i == nil || i.handle == nil {
		return IONEXSlantDelayEvaluation{}, ErrClosed
	}
	v, e := i.handle.SlantDelayWithPolicy(lat, lon, azimuth, elevation, epochJ2000S, frequencyHz, uint32(policy))
	return publicIONEXEvaluation(v), publicError(e)
}

// SlantDelayAtInstant evaluates a delay without reducing the supplied epoch to whole seconds.
func (i *IONEX) SlantDelayAtInstant(lat, lon, azimuth, elevation float64, epoch ClockEpoch, frequencyHz float64) (float64, *IONEXEpochError, error) {
	if i == nil || i.handle == nil {
		return 0, nil, ErrClosed
	}
	delay, detail, err := i.handle.SlantDelayAtInstant(lat, lon, azimuth, elevation, nativeClockEpoch(epoch), frequencyHz)
	return delay, publicIONEXEpochError(detail), publicError(err)
}

// SlantDelayAtInstantWithPolicy applies the exact composite policy and returns complete typed detail.
func (i *IONEX) SlantDelayAtInstantWithPolicy(lat, lon, azimuth, elevation float64, epoch ClockEpoch, frequencyHz float64, policy IONEXSlantPolicy) (IONEXInstantSlantResult, error) {
	if i == nil || i.handle == nil {
		return IONEXInstantSlantResult{}, ErrClosed
	}
	row, err := i.handle.SlantDelayAtInstantWithPolicy(lat, lon, azimuth, elevation, nativeClockEpoch(epoch), frequencyHz, nativeIONEXPolicy(policy))
	return publicIONEXInstantRow(row), publicError(err)
}

// SlantDelayResultsAtInstants fills caller-owned Go rows from exact-time requests.
func (i *IONEX) SlantDelayResultsAtInstants(requests []IONEXInstantSlantRequest, out []IONEXInstantSlantResult, policy IONEXSlantPolicy) error {
	if i == nil || i.handle == nil {
		return ErrClosed
	}
	if len(out) != len(requests) {
		return invalidArgument("IONEX request and output lengths differ")
	}
	nativeRequests := nativeIONEXRequests(requests)
	rows, err := i.handle.SlantDelayResultsAtInstants(nativeRequests, nativeIONEXPolicy(policy))
	if err != nil {
		return publicError(err)
	}
	if len(rows) != len(out) {
		return errors.New("sidereon: native IONEX result count differs from request count")
	}
	for index := range out {
		out[index] = publicIONEXInstantRow(rows[index])
	}
	return nil
}

// SlantDelayResultsAtInstantsOwned returns independent rows with their messages and mapping codes.
func (i *IONEX) SlantDelayResultsAtInstantsOwned(requests []IONEXInstantSlantRequest, policy IONEXSlantPolicy) (IONEXInstantSlantResultList, error) {
	if i == nil || i.handle == nil {
		return IONEXInstantSlantResultList{}, ErrClosed
	}
	values, err := i.handle.SlantDelayResultsAtInstantsOwned(nativeIONEXRequests(requests), nativeIONEXPolicy(policy))
	if err != nil {
		return IONEXInstantSlantResultList{}, publicError(err)
	}
	out := IONEXInstantSlantResultList{Rows: make([]IONEXInstantSlantResult, len(values.Rows))}
	for index := range values.Rows {
		out.Rows[index] = publicIONEXInstantRow(values.Rows[index])
	}
	return out, nil
}

// DefaultIONEXSlantPolicy returns strict coverage/nodes and single-layer mapping.
func DefaultIONEXSlantPolicy() IONEXSlantPolicy {
	return IONEXSlantPolicy{Coverage: IONEXCoveragePolicyStrict, MissingNodes: IONEXMissingNodesStrict, Mapping: IONEXMappingSingleLayer}
}

func nativeIONEXPolicy(value IONEXSlantPolicy) native.IonexSlantPolicy {
	return native.IonexSlantPolicy{Coverage: uint32(value.Coverage), MissingNodes: uint32(value.MissingNodes), Mapping: uint32(value.Mapping)}
}
func nativeIONEXRequests(values []IONEXInstantSlantRequest) []native.IonexInstantSlantRequest {
	out := make([]native.IonexInstantSlantRequest, len(values))
	for i, v := range values {
		out[i] = native.IonexInstantSlantRequest{LatDeg: v.LatDeg, LonDeg: v.LonDeg, AzimuthDeg: v.AzimuthDeg, ElevationDeg: v.ElevationDeg, Epoch: nativeClockEpoch(v.Epoch), FrequencyHz: v.FrequencyHz}
	}
	return out
}
func publicIONEXEvaluation(v native.IonexSlantDelayEvaluation) IONEXSlantDelayEvaluation {
	return IONEXSlantDelayEvaluation{DelayM: v.DelayM, Status: IONEXSlantDelayStatus(v.Status), CoverageError: IONEXCoverageErrorKind(v.CoverageError), IsValid: v.IsValid, HasHeld: v.HasHeld, HasDegraded: v.HasDegraded, HasAssumedMapping: v.HasAssumedMapping, Gap: IONEXNodeGap{HasGap: v.Gap.HasGap, Earlier: v.Gap.Earlier, Later: v.Gap.Later}, AssumedMapping: IONEXAssumedMappingKind(v.AssumedMapping)}
}
func publicIONEXEpochError(v native.IonexEpochError) *IONEXEpochError {
	if v.Kind == 0 {
		return nil
	}
	return &IONEXEpochError{Kind: IONEXEpochErrorKind(v.Kind), Scale: TimeScale(v.Scale), HasUTCJ2000S: v.HasUTCJ2000S, UTCJ2000S: v.UTCJ2000S}
}
func publicIONEXSlantError(v native.IonexSlantError) *IONEXSlantError {
	if v.Kind == 0 {
		return nil
	}
	return &IONEXSlantError{Kind: IONEXSlantErrorKind(v.Kind), CoverageError: IONEXCoverageErrorKind(v.CoverageError), HasGap: v.HasGap, Gap: IONEXNodeGap{HasGap: v.Gap.HasGap, Earlier: v.Gap.Earlier, Later: v.Gap.Later}, Refusal: v.Refusal, RefusalMapNumber: v.RefusalMapNumber, RefusalLatIndex: v.RefusalLatIndex, RefusalLonIndex: v.RefusalLonIndex, HasMappingDeclaration: v.HasMappingDeclaration, MappingDeclaration: v.MappingDeclaration, HasMappingFunction: v.HasMappingFunction, MappingFunction: v.MappingFunction}
}
func publicIONEXInstantRow(v native.IonexInstantSlantRow) IONEXInstantSlantResult {
	return IONEXInstantSlantResult{IsOK: v.IsOK, Status: StatusCode(v.Status), Evaluation: publicIONEXEvaluation(v.Evaluation), Error: publicIONEXSlantError(v.Error), EpochError: publicIONEXEpochError(v.EpochError), Message: v.Message, MappingCode: v.MappingCode}
}

// TECSamples returns detached TEC samples from the product.
func (i *IONEX) TECSamples() ([]TECSample, error) {
	if i == nil || i.handle == nil {
		return nil, ErrClosed
	}
	v, e := i.handle.TECSamples()
	if e != nil {
		return nil, publicError(e)
	}
	out := make([]TECSample, len(v))
	for j := range v {
		out[j] = publicTECSample(v[j])
	}
	return out, nil
}

// GridInfo returns native grid dimensions and metadata.
func (i *IONEX) GridInfo() (TECGridSamplesInfo, error) {
	if i == nil || i.handle == nil {
		return TECGridSamplesInfo{}, ErrClosed
	}
	v, e := i.handle.GridInfo()
	return TECGridSamplesInfo{MapEpochCount: v.MapEpochCount, LatNodeCount: v.LatNodeCount, LonNodeCount: v.LonNodeCount, DLatDeg: v.DLatDeg, DLonDeg: v.DLonDeg, ShellHeightKm: v.ShellHeightKm, BaseRadiusKm: v.BaseRadiusKm, Exponent: v.Exponent, RMSPresent: v.RMSPresent, TECMAPValueCount: v.TECMAPValueCount, RMSMAPValueCount: v.RMSMAPValueCount}, publicError(e)
}

// TECMAPsTECU returns flattened VTEC map values.
func (i *IONEX) TECMAPsTECU() ([]float64, error) {
	if i == nil || i.handle == nil {
		return nil, ErrClosed
	}
	v, e := i.handle.TECMAPsTECU()
	return v, publicError(e)
}

// RMSMAPsTECU returns flattened RMS map values.
func (i *IONEX) RMSMAPsTECU() ([]float64, error) {
	if i == nil || i.handle == nil {
		return nil, ErrClosed
	}
	v, e := i.handle.RMSMAPsTECU()
	return v, publicError(e)
}

// GridEpochsJ2000S returns grid map epochs as J2000 seconds.
func (i *IONEX) GridEpochsJ2000S() ([]float64, error) {
	if i == nil || i.handle == nil {
		return nil, ErrClosed
	}
	v, e := i.handle.GridEpochsJ2000S()
	return v, publicError(e)
}

// IONEXCoveragePolicy controls behavior outside map coverage.
type IONEXCoveragePolicy uint32

const (
	// IONEXCoveragePolicyStrict rejects epochs outside map coverage.
	IONEXCoveragePolicyStrict IONEXCoveragePolicy = IONEXCoveragePolicy(native.IONEXCoveragePolicyStrictValue)
	// IONEXCoveragePolicyHold holds the nearest map outside coverage.
	IONEXCoveragePolicyHold IONEXCoveragePolicy = IONEXCoveragePolicy(native.IONEXCoveragePolicyHoldValue)
)

// IONEXSlantDelayStatus identifies the returned delay status.
type IONEXSlantDelayStatus uint32

const (
	// IONEXSlantDelayStatusValid indicates a directly interpolated delay.
	IONEXSlantDelayStatusValid IONEXSlantDelayStatus = IONEXSlantDelayStatus(native.IONEXSlantDelayStatusValidValue)
	// IONEXSlantDelayStatusHeld indicates a held boundary delay.
	IONEXSlantDelayStatusHeld IONEXSlantDelayStatus = IONEXSlantDelayStatus(native.IONEXSlantDelayStatusHeldValue)
)

// IONEXCoverageErrorKind identifies a held-value coverage miss.
type IONEXCoverageErrorKind uint32

const (
	// IONEXCoverageErrorNone indicates no coverage error.
	IONEXCoverageErrorNone IONEXCoverageErrorKind = IONEXCoverageErrorKind(native.IONEXCoverageErrorNoneValue)
	// IONEXCoverageErrorEpochBeforeFirstMap indicates an early epoch.
	IONEXCoverageErrorEpochBeforeFirstMap IONEXCoverageErrorKind = IONEXCoverageErrorKind(native.IONEXCoverageErrorEpochBeforeFirstMapValue)
	// IONEXCoverageErrorEpochAfterLastMap indicates a late epoch.
	IONEXCoverageErrorEpochAfterLastMap IONEXCoverageErrorKind = IONEXCoverageErrorKind(native.IONEXCoverageErrorEpochAfterLastMapValue)
	// IONEXCoverageErrorLatitude indicates an invalid latitude.
	IONEXCoverageErrorLatitude IONEXCoverageErrorKind = IONEXCoverageErrorKind(native.IONEXCoverageErrorLatitudeValue)
	// IONEXCoverageErrorLongitude indicates an invalid longitude.
	IONEXCoverageErrorLongitude IONEXCoverageErrorKind = IONEXCoverageErrorKind(native.IONEXCoverageErrorLongitudeValue)
)
