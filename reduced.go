package sidereon

import (
	"errors"

	"github.com/neilberkman/sidereon-go/internal/native"
)

// CalendarEpoch is a UTC proleptic-Gregorian epoch used by reduced-orbit
// routes. The second field may contain a fractional second.
type CalendarEpoch struct {
	Year, Month, Day, Hour, Minute int
	Second                         float64
}
type ReducedOrbitModel uint32

const (
	ReducedOrbitCircularSecular  ReducedOrbitModel = ReducedOrbitModel(native.ReducedOrbitCircularSecularValue)
	ReducedOrbitEccentricSecular ReducedOrbitModel = ReducedOrbitModel(native.ReducedOrbitEccentricSecularValue)
)

type ReducedOrbitFrame uint32

const (
	ReducedOrbitGCRS ReducedOrbitFrame = ReducedOrbitFrame(native.ReducedOrbitFrameGCRSValue)
	ReducedOrbitECEF ReducedOrbitFrame = ReducedOrbitFrame(native.ReducedOrbitFrameECEFValue)
)

type ReducedOrbitElements struct {
	Model                                                                                                                                                  ReducedOrbitModel
	Epoch                                                                                                                                                  CalendarEpoch
	SemiMajorAxisM, Eccentricity, InclinationRad, RAANRad, RAANRateRadS, RAANRateJ2RadS, ArgumentOfLatitudeRad, MeanMotionRadS, H, K, ArgumentOfPerigeeRad float64
}
type ECEFSample struct {
	Epoch      CalendarEpoch
	XM, YM, ZM float64
}
type ReducedOrbitFitStats struct {
	RMSM, MaxM  float64
	SampleCount int
}
type ReducedOrbitDriftEntry struct {
	Epoch  CalendarEpoch
	ErrorM float64
}
type ReducedOrbitDriftSummary struct {
	MaxM, RMSM           float64
	HasThresholdCrossing bool
	ThresholdIndex       int
}
type ReducedOrbitDriftReport struct {
	_      noCopy
	handle *native.ReducedDriftReport
}
type ReducedOrbitPiecewise struct {
	_      noCopy
	handle *native.ReducedPiecewise
}

// ReducedOrbitSourceSampling describes an inclusive UTC/scale-aware source
// sampling window and cadence in seconds.
type ReducedOrbitSourceSampling struct {
	Start, End CalendarEpoch
	CadenceS   float64
}

// ReducedOrbitSourceFitOptions selects a reduced-orbit model for source-backed
// fitting.
type ReducedOrbitSourceFitOptions struct {
	Sampling ReducedOrbitSourceSampling
	Model    ReducedOrbitModel
}

// ReducedOrbitSourceDriftOptions selects a source sampling window and the
// first-crossing threshold in metres.
type ReducedOrbitSourceDriftOptions struct {
	Sampling   ReducedOrbitSourceSampling
	ThresholdM float64
}

type ReducedOrbitPiecewiseSourceFitStats struct {
	RequestedSamples, UsedSamples int
}

type ReducedOrbitSourceFitStats struct {
	Fit              ReducedOrbitFitStats
	RequestedSamples int
}

type ReducedOrbitPiecewiseInfo struct {
	Model          ReducedOrbitModel
	Scale          TimeScale
	Start, End     CalendarEpoch
	SegmentSeconds int64
	SegmentCount   int
}

type ReducedOrbitPiecewiseSegment struct {
	Start, End CalendarEpoch
	Elements   ReducedOrbitElements
	Stats      ReducedOrbitFitStats
}

func nativeEpoch(v CalendarEpoch) native.NativeCalendarEpoch {
	return native.NativeCalendarEpoch{Year: int32(v.Year), Month: int32(v.Month), Day: int32(v.Day), Hour: int32(v.Hour), Minute: int32(v.Minute), Second: v.Second}
}

func validateCalendarEpoch(v CalendarEpoch) error {
	if int64(v.Year) < -2147483648 || int64(v.Year) > 2147483647 || int64(v.Month) < -2147483648 || int64(v.Month) > 2147483647 || int64(v.Day) < -2147483648 || int64(v.Day) > 2147483647 || int64(v.Hour) < -2147483648 || int64(v.Hour) > 2147483647 || int64(v.Minute) < -2147483648 || int64(v.Minute) > 2147483647 {
		return errors.New("sidereon: calendar epoch field does not fit int32")
	}
	return nil
}

func validateReducedSamples(values []ECEFSample) error {
	for _, value := range values {
		if err := validateCalendarEpoch(value.Epoch); err != nil {
			return err
		}
	}
	return nil
}

func validateReducedSampling(value ReducedOrbitSourceSampling) error {
	if err := validateCalendarEpoch(value.Start); err != nil {
		return err
	}
	return validateCalendarEpoch(value.End)
}
func calendarEpoch(v native.NativeCalendarEpoch) CalendarEpoch {
	return CalendarEpoch{Year: int(v.Year), Month: int(v.Month), Day: int(v.Day), Hour: int(v.Hour), Minute: int(v.Minute), Second: v.Second}
}
func nativeElements(v ReducedOrbitElements) native.NativeReducedElements {
	return native.NativeReducedElements{Model: uint32(v.Model), Epoch: nativeEpoch(v.Epoch), AM: v.SemiMajorAxisM, E: v.Eccentricity, I: v.InclinationRad, RAAN: v.RAANRad, RAANRate: v.RAANRateRadS, RAANRateJ2: v.RAANRateJ2RadS, ArgLat: v.ArgumentOfLatitudeRad, MeanMotion: v.MeanMotionRadS, H: v.H, K: v.K, ArgPerigee: v.ArgumentOfPerigeeRad}
}
func reducedElements(v native.NativeReducedElements) ReducedOrbitElements {
	return ReducedOrbitElements{ReducedOrbitModel(v.Model), calendarEpoch(v.Epoch), v.AM, v.E, v.I, v.RAAN, v.RAANRate, v.RAANRateJ2, v.ArgLat, v.MeanMotion, v.H, v.K, v.ArgPerigee}
}
func nativeSamples(v []ECEFSample) []native.NativeECEFSample {
	out := make([]native.NativeECEFSample, len(v))
	for i, x := range v {
		out[i] = native.NativeECEFSample{Epoch: nativeEpoch(x.Epoch), X: x.XM, Y: x.YM, Z: x.ZM}
	}
	return out
}

func nativeReducedSampling(v ReducedOrbitSourceSampling) native.NativeReducedSampling {
	return native.NativeReducedSampling{T0: nativeEpoch(v.Start), T1: nativeEpoch(v.End), Cadence: v.CadenceS}
}

func nativeReducedFitOptions(v ReducedOrbitSourceFitOptions) native.NativeReducedSourceFitOptions {
	return native.NativeReducedSourceFitOptions{Sampling: nativeReducedSampling(v.Sampling), Model: uint32(v.Model)}
}

func nativeReducedDriftOptions(v ReducedOrbitSourceDriftOptions) native.NativeReducedSourceDriftOptions {
	return native.NativeReducedSourceDriftOptions{Sampling: nativeReducedSampling(v.Sampling), Threshold: v.ThresholdM}
}
func ReducedOrbitFit(samples []ECEFSample, scale TimeScale, model ReducedOrbitModel) (ReducedOrbitElements, ReducedOrbitFitStats, error) {
	if err := validateReducedSamples(samples); err != nil {
		return ReducedOrbitElements{}, ReducedOrbitFitStats{}, err
	}
	e, s, err := native.ReducedOrbitFit(nativeSamples(samples), uint32(scale), uint32(model))
	if err != nil {
		return ReducedOrbitElements{}, ReducedOrbitFitStats{}, publicError(err)
	}
	sampleCount, conversionErr := nativeCountToInt(s.NSamples, "reduced fit sample count")
	if conversionErr != nil {
		return ReducedOrbitElements{}, ReducedOrbitFitStats{}, conversionErr
	}
	return reducedElements(e), ReducedOrbitFitStats{s.RMS, s.Max, sampleCount}, nil
}
func (e ReducedOrbitElements) Position(epoch CalendarEpoch, scale TimeScale, frame ReducedOrbitFrame) ([3]float64, error) {
	if err := validateCalendarEpoch(e.Epoch); err != nil {
		return [3]float64{}, err
	}
	if err := validateCalendarEpoch(epoch); err != nil {
		return [3]float64{}, err
	}
	v, err := native.ReducedOrbitPosition(nativeElements(e), nativeEpoch(epoch), uint32(scale), uint32(frame))
	return v, publicError(err)
}
func (e ReducedOrbitElements) PositionVelocity(epoch CalendarEpoch, scale TimeScale, frame ReducedOrbitFrame) ([3]float64, [3]float64, error) {
	if err := validateCalendarEpoch(e.Epoch); err != nil {
		return [3]float64{}, [3]float64{}, err
	}
	if err := validateCalendarEpoch(epoch); err != nil {
		return [3]float64{}, [3]float64{}, err
	}
	p, v, err := native.ReducedOrbitPositionVelocity(nativeElements(e), nativeEpoch(epoch), uint32(scale), uint32(frame))
	return p, v, publicError(err)
}
func ReducedOrbitDrift(elements ReducedOrbitElements, truth []ECEFSample, scale TimeScale, thresholdM float64) (*ReducedOrbitDriftReport, error) {
	if err := validateCalendarEpoch(elements.Epoch); err != nil {
		return nil, err
	}
	if err := validateReducedSamples(truth); err != nil {
		return nil, err
	}
	h, e := native.ReducedOrbitDrift(nativeElements(elements), nativeSamples(truth), uint32(scale), thresholdM)
	if e != nil {
		return nil, publicError(e)
	}
	return &ReducedOrbitDriftReport{handle: h}, nil
}

// ReducedOrbitDriftSP3Source evaluates a fitted model against samples from one
// SP3 satellite. Sampling and drift calculations are performed in C.
func ReducedOrbitDriftSP3Source(elements ReducedOrbitElements, sp3 *SP3, satelliteID string, options ReducedOrbitSourceDriftOptions) (*ReducedOrbitDriftReport, error) {
	if sp3 == nil || sp3.handle == nil {
		return nil, ErrClosed
	}
	if err := validateCalendarEpoch(elements.Epoch); err != nil {
		return nil, err
	}
	if err := validateReducedSampling(options.Sampling); err != nil {
		return nil, err
	}
	h, err := native.ReducedOrbitDriftSP3Source(nativeElements(elements), sp3.handle, satelliteID, nativeReducedDriftOptions(options))
	if err != nil {
		return nil, publicError(err)
	}
	return &ReducedOrbitDriftReport{handle: h}, nil
}

// ReducedOrbitDriftTLESource evaluates a fitted model against samples from a
// TLE/SGP4 source in UTC. Sampling and drift calculations are performed in C.
func ReducedOrbitDriftTLESource(elements ReducedOrbitElements, tle *TLE, options ReducedOrbitSourceDriftOptions) (*ReducedOrbitDriftReport, error) {
	if tle == nil || tle.handle == nil {
		return nil, ErrClosed
	}
	if err := validateCalendarEpoch(elements.Epoch); err != nil {
		return nil, err
	}
	if err := validateReducedSampling(options.Sampling); err != nil {
		return nil, err
	}
	h, err := native.ReducedOrbitDriftTLESource(nativeElements(elements), tle.handle, nativeReducedDriftOptions(options))
	if err != nil {
		return nil, publicError(err)
	}
	return &ReducedOrbitDriftReport{handle: h}, nil
}

// ReducedOrbitFitPiecewise delegates piecewise fitting to C. segmentSeconds is
// the requested segment duration in seconds.
func ReducedOrbitFitPiecewise(samples []ECEFSample, scale TimeScale, model ReducedOrbitModel, start, end CalendarEpoch, segmentSeconds int64) (*ReducedOrbitPiecewise, error) {
	if err := validateReducedSamples(samples); err != nil {
		return nil, err
	}
	if err := validateCalendarEpoch(start); err != nil {
		return nil, err
	}
	if err := validateCalendarEpoch(end); err != nil {
		return nil, err
	}
	h, err := native.ReducedOrbitFitPiecewise(nativeSamples(samples), uint32(scale), uint32(model), nativeEpoch(start), nativeEpoch(end), segmentSeconds)
	if err != nil {
		return nil, publicError(err)
	}
	return &ReducedOrbitPiecewise{handle: h}, nil
}

func (p *ReducedOrbitPiecewise) Close() error {
	if p == nil || p.handle == nil {
		return nil
	}
	return publicError(p.handle.Close())
}

func (p *ReducedOrbitPiecewise) Info() (ReducedOrbitPiecewiseInfo, error) {
	if p == nil || p.handle == nil {
		return ReducedOrbitPiecewiseInfo{}, ErrClosed
	}
	v, err := p.handle.Info()
	if err != nil {
		return ReducedOrbitPiecewiseInfo{}, publicError(err)
	}
	segmentCount, conversionErr := nativeCountToInt(v.NSegments, "reduced piecewise segment count")
	if conversionErr != nil {
		return ReducedOrbitPiecewiseInfo{}, conversionErr
	}
	return ReducedOrbitPiecewiseInfo{ReducedOrbitModel(v.Model), TimeScale(v.Scale), calendarEpoch(v.T0), calendarEpoch(v.T1), v.SegmentSeconds, segmentCount}, nil
}

func (p *ReducedOrbitPiecewise) Segments() ([]ReducedOrbitPiecewiseSegment, error) {
	if p == nil || p.handle == nil {
		return nil, ErrClosed
	}
	v, err := p.handle.Segments()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]ReducedOrbitPiecewiseSegment, len(v))
	for i, segment := range v {
		sampleCount, conversionErr := nativeCountToInt(segment.Stats.NSamples, "reduced piecewise sample count")
		if conversionErr != nil {
			return nil, conversionErr
		}
		out[i] = ReducedOrbitPiecewiseSegment{calendarEpoch(segment.T0), calendarEpoch(segment.T1), reducedElements(segment.Elements), ReducedOrbitFitStats{segment.Stats.RMS, segment.Stats.Max, sampleCount}}
	}
	return out, nil
}

func (p *ReducedOrbitPiecewise) SelectSegment(epoch CalendarEpoch) (int, ReducedOrbitPiecewiseSegment, error) {
	if p == nil || p.handle == nil {
		return 0, ReducedOrbitPiecewiseSegment{}, ErrClosed
	}
	if err := validateCalendarEpoch(epoch); err != nil {
		return 0, ReducedOrbitPiecewiseSegment{}, err
	}
	index, segment, err := p.handle.SelectSegment(nativeEpoch(epoch))
	if err != nil {
		return 0, ReducedOrbitPiecewiseSegment{}, publicError(err)
	}
	indexInt, conversionErr := nativeCountToInt(index, "reduced selected segment index")
	if conversionErr != nil {
		return 0, ReducedOrbitPiecewiseSegment{}, conversionErr
	}
	sampleCount, conversionErr := nativeCountToInt(segment.Stats.NSamples, "reduced selected segment sample count")
	if conversionErr != nil {
		return 0, ReducedOrbitPiecewiseSegment{}, conversionErr
	}
	return indexInt, ReducedOrbitPiecewiseSegment{calendarEpoch(segment.T0), calendarEpoch(segment.T1), reducedElements(segment.Elements), ReducedOrbitFitStats{segment.Stats.RMS, segment.Stats.Max, sampleCount}}, nil
}

func (p *ReducedOrbitPiecewise) Position(epoch CalendarEpoch, frame ReducedOrbitFrame) ([3]float64, error) {
	if p == nil || p.handle == nil {
		return [3]float64{}, ErrClosed
	}
	if err := validateCalendarEpoch(epoch); err != nil {
		return [3]float64{}, err
	}
	v, err := p.handle.Position(nativeEpoch(epoch), uint32(frame))
	return v, publicError(err)
}

func (p *ReducedOrbitPiecewise) PositionVelocity(epoch CalendarEpoch, frame ReducedOrbitFrame) ([3]float64, [3]float64, error) {
	if p == nil || p.handle == nil {
		return [3]float64{}, [3]float64{}, ErrClosed
	}
	if err := validateCalendarEpoch(epoch); err != nil {
		return [3]float64{}, [3]float64{}, err
	}
	position, velocity, err := p.handle.PositionVelocity(nativeEpoch(epoch), uint32(frame))
	return position, velocity, publicError(err)
}

func (p *ReducedOrbitPiecewise) Drift(truth []ECEFSample, thresholdM float64) (*ReducedOrbitDriftReport, error) {
	if p == nil || p.handle == nil {
		return nil, ErrClosed
	}
	if err := validateReducedSamples(truth); err != nil {
		return nil, err
	}
	h, err := p.handle.Drift(nativeSamples(truth), thresholdM)
	if err != nil {
		return nil, publicError(err)
	}
	return &ReducedOrbitDriftReport{handle: h}, nil
}

func ReducedOrbitFitSP3Source(sp3 *SP3, satelliteID string, options ReducedOrbitSourceFitOptions) (ReducedOrbitElements, ReducedOrbitSourceFitStats, error) {
	if sp3 == nil || sp3.handle == nil {
		return ReducedOrbitElements{}, ReducedOrbitSourceFitStats{}, ErrClosed
	}
	if err := validateReducedSampling(options.Sampling); err != nil {
		return ReducedOrbitElements{}, ReducedOrbitSourceFitStats{}, err
	}
	elements, stats, err := native.ReducedOrbitFitSP3Source(sp3.handle, satelliteID, nativeReducedFitOptions(options))
	if err != nil {
		return ReducedOrbitElements{}, ReducedOrbitSourceFitStats{}, publicError(err)
	}
	sampleCount, conversionErr := nativeCountToInt(stats.Fit.NSamples, "reduced source fit sample count")
	if conversionErr != nil {
		return ReducedOrbitElements{}, ReducedOrbitSourceFitStats{}, conversionErr
	}
	requested, conversionErr := nativeCountToInt(stats.Requested, "reduced source requested sample count")
	if conversionErr != nil {
		return ReducedOrbitElements{}, ReducedOrbitSourceFitStats{}, conversionErr
	}
	return reducedElements(elements), ReducedOrbitSourceFitStats{ReducedOrbitFitStats{stats.Fit.RMS, stats.Fit.Max, sampleCount}, requested}, nil
}

func ReducedOrbitFitTLESource(tle *TLE, options ReducedOrbitSourceFitOptions) (ReducedOrbitElements, ReducedOrbitSourceFitStats, error) {
	if tle == nil || tle.handle == nil {
		return ReducedOrbitElements{}, ReducedOrbitSourceFitStats{}, ErrClosed
	}
	if err := validateReducedSampling(options.Sampling); err != nil {
		return ReducedOrbitElements{}, ReducedOrbitSourceFitStats{}, err
	}
	elements, stats, err := native.ReducedOrbitFitTLESource(tle.handle, nativeReducedFitOptions(options))
	if err != nil {
		return ReducedOrbitElements{}, ReducedOrbitSourceFitStats{}, publicError(err)
	}
	sampleCount, conversionErr := nativeCountToInt(stats.Fit.NSamples, "reduced source fit sample count")
	if conversionErr != nil {
		return ReducedOrbitElements{}, ReducedOrbitSourceFitStats{}, conversionErr
	}
	requested, conversionErr := nativeCountToInt(stats.Requested, "reduced source requested sample count")
	if conversionErr != nil {
		return ReducedOrbitElements{}, ReducedOrbitSourceFitStats{}, conversionErr
	}
	return reducedElements(elements), ReducedOrbitSourceFitStats{ReducedOrbitFitStats{stats.Fit.RMS, stats.Fit.Max, sampleCount}, requested}, nil
}

func ReducedOrbitFitPiecewiseSP3Source(sp3 *SP3, satelliteID string, options ReducedOrbitSourceFitOptions, segmentSeconds float64) (*ReducedOrbitPiecewise, ReducedOrbitPiecewiseSourceFitStats, error) {
	if sp3 == nil || sp3.handle == nil {
		return nil, ReducedOrbitPiecewiseSourceFitStats{}, ErrClosed
	}
	if err := validateReducedSampling(options.Sampling); err != nil {
		return nil, ReducedOrbitPiecewiseSourceFitStats{}, err
	}
	h, stats, err := native.ReducedOrbitFitPiecewiseSP3Source(sp3.handle, satelliteID, nativeReducedFitOptions(options), segmentSeconds)
	if err != nil {
		return nil, ReducedOrbitPiecewiseSourceFitStats{}, publicError(err)
	}
	requested, conversionErr := nativeCountToInt(stats.Requested, "reduced piecewise requested sample count")
	if conversionErr != nil {
		return nil, ReducedOrbitPiecewiseSourceFitStats{}, conversionErr
	}
	used, conversionErr := nativeCountToInt(stats.Used, "reduced piecewise used sample count")
	if conversionErr != nil {
		return nil, ReducedOrbitPiecewiseSourceFitStats{}, conversionErr
	}
	return &ReducedOrbitPiecewise{handle: h}, ReducedOrbitPiecewiseSourceFitStats{requested, used}, nil
}

func ReducedOrbitFitPiecewiseTLESource(tle *TLE, options ReducedOrbitSourceFitOptions, segmentSeconds float64) (*ReducedOrbitPiecewise, ReducedOrbitPiecewiseSourceFitStats, error) {
	if tle == nil || tle.handle == nil {
		return nil, ReducedOrbitPiecewiseSourceFitStats{}, ErrClosed
	}
	if err := validateReducedSampling(options.Sampling); err != nil {
		return nil, ReducedOrbitPiecewiseSourceFitStats{}, err
	}
	h, stats, err := native.ReducedOrbitFitPiecewiseTLESource(tle.handle, nativeReducedFitOptions(options), segmentSeconds)
	if err != nil {
		return nil, ReducedOrbitPiecewiseSourceFitStats{}, publicError(err)
	}
	requested, conversionErr := nativeCountToInt(stats.Requested, "reduced piecewise requested sample count")
	if conversionErr != nil {
		return nil, ReducedOrbitPiecewiseSourceFitStats{}, conversionErr
	}
	used, conversionErr := nativeCountToInt(stats.Used, "reduced piecewise used sample count")
	if conversionErr != nil {
		return nil, ReducedOrbitPiecewiseSourceFitStats{}, conversionErr
	}
	return &ReducedOrbitPiecewise{handle: h}, ReducedOrbitPiecewiseSourceFitStats{requested, used}, nil
}

func (p *ReducedOrbitPiecewise) DriftSP3Source(sp3 *SP3, satelliteID string, options ReducedOrbitSourceDriftOptions) (*ReducedOrbitDriftReport, error) {
	if p == nil || p.handle == nil || sp3 == nil || sp3.handle == nil {
		return nil, ErrClosed
	}
	if err := validateReducedSampling(options.Sampling); err != nil {
		return nil, err
	}
	h, err := p.handle.DriftSP3Source(sp3.handle, satelliteID, nativeReducedDriftOptions(options))
	if err != nil {
		return nil, publicError(err)
	}
	return &ReducedOrbitDriftReport{handle: h}, nil
}

func (p *ReducedOrbitPiecewise) DriftTLESource(tle *TLE, options ReducedOrbitSourceDriftOptions) (*ReducedOrbitDriftReport, error) {
	if p == nil || p.handle == nil || tle == nil || tle.handle == nil {
		return nil, ErrClosed
	}
	if err := validateReducedSampling(options.Sampling); err != nil {
		return nil, err
	}
	h, err := p.handle.DriftTLESource(tle.handle, nativeReducedDriftOptions(options))
	if err != nil {
		return nil, publicError(err)
	}
	return &ReducedOrbitDriftReport{handle: h}, nil
}
func (r *ReducedOrbitDriftReport) Close() error {
	if r == nil || r.handle == nil {
		return nil
	}
	return publicError(r.handle.Close())
}
func (r *ReducedOrbitDriftReport) Output() ([]ReducedOrbitDriftEntry, ReducedOrbitDriftSummary, int, error) {
	if r == nil || r.handle == nil {
		return nil, ReducedOrbitDriftSummary{}, 0, ErrClosed
	}
	v, s, n, e := r.handle.Output()
	if e != nil {
		return nil, ReducedOrbitDriftSummary{}, 0, publicError(e)
	}
	out := make([]ReducedOrbitDriftEntry, len(v))
	for i, x := range v {
		out[i] = ReducedOrbitDriftEntry{calendarEpoch(x.Epoch), x.Error}
	}
	thresholdIndex, conversionErr := nativeCountToInt(s.ThresholdIndex, "reduced drift threshold index")
	if conversionErr != nil {
		return nil, ReducedOrbitDriftSummary{}, 0, conversionErr
	}
	requested, conversionErr := nativeCountToInt(n, "reduced drift requested sample count")
	if conversionErr != nil {
		return nil, ReducedOrbitDriftSummary{}, 0, conversionErr
	}
	return out, ReducedOrbitDriftSummary{s.Max, s.RMS, s.HasCrossing, thresholdIndex}, requested, nil
}
