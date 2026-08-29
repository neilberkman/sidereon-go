package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

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
	DelayM        float64
	Status        IONEXSlantDelayStatus
	CoverageError IONEXCoverageErrorKind
}

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
	return IONEXSlantDelayEvaluation{DelayM: v.DelayM, Status: IONEXSlantDelayStatus(v.Status), CoverageError: IONEXCoverageErrorKind(v.CoverageError)}, publicError(e)
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
