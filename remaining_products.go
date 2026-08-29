package sidereon

import (
	"fmt"
	"os"

	"github.com/neilberkman/sidereon-go/internal/native"
)

// PreciseEphemerisSamples is a C-owned sample-backed ephemeris source.
type PreciseEphemerisSamples struct {
	_      noCopy
	handle *native.PreciseEphemerisSamples
}

// BuildPreciseEphemerisSamples copies canonical samples into a native-owned source.
func BuildPreciseEphemerisSamples(values []PreciseEphemerisSample) (*PreciseEphemerisSamples, error) {
	nativeValues := make([]native.PreciseEphemerisSample, len(values))
	for i := range values {
		nativeValues[i] = nativePreciseSample(values[i])
	}
	h, err := native.PreciseEphemerisSamplesFromSamples(nativeValues)
	if err != nil {
		return nil, publicError(err)
	}
	return &PreciseEphemerisSamples{handle: h}, nil
}

// Close releases the native sample source and is safe to call repeatedly.
func (s *PreciseEphemerisSamples) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return publicError(s.handle.Close())
}

// ObservableStates evaluates per-satellite J2000 epochs through the sample source.
func (s *PreciseEphemerisSamples) ObservableStates(satellites []string, epochs []float64) ([]ObservableStateRow, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	rows, err := s.handle.ObservableStates(append([]string(nil), satellites...), append([]float64(nil), epochs...))
	if err != nil {
		return nil, publicError(err)
	}
	return publicObservableStates(rows), nil
}

// ObservableStatesShared evaluates all requested satellites at one J2000 epoch.
func (s *PreciseEphemerisSamples) ObservableStatesShared(satellites []string, epoch float64) ([]ObservableStateRow, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	rows, err := s.handle.ObservableStatesShared(append([]string(nil), satellites...), epoch)
	if err != nil {
		return nil, publicError(err)
	}
	return publicObservableStates(rows), nil
}

// Sample evaluates a regular J2000 grid through the sample source.
func (s *PreciseEphemerisSamples) Sample(satellites []string, start, stop, step float64) ([]EphemerisSampleRow, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	rows, err := s.handle.Sample(append([]string(nil), satellites...), start, stop, step)
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]EphemerisSampleRow, len(rows))
	for i, row := range rows {
		out[i] = EphemerisSampleRow{SatelliteID: row.SatelliteID, EpochJ2000S: row.EpochJ2000S, Status: EphemerisSampleStatus(row.Status), HasPosition: row.HasPosition, PositionECEFM: row.Position, HasClock: row.HasClock, ClockS: row.ClockS}
	}
	return out, nil
}

// RangePrediction contains the geometry-only result of one precise-source
// range request. Distances and positions are metres; epochs and clocks are
// seconds. The satellite clock is optional via HasSatelliteClock.
type RangePrediction struct {
	// GeometricRangeM is the geometric range m in metres.
	GeometricRangeM float64
	// HasSatelliteClock reports whether SatelliteClockS is valid.
	HasSatelliteClock bool
	// SatelliteClockS is the satellite clock s in seconds.
	SatelliteClockS float64
	// TransmitTimeJ2000S is the transmit epoch in seconds from J2000.
	TransmitTimeJ2000S float64
	// SatellitePositionECEF is the satellite's ECEF position vector in metres.
	SatellitePositionECEF [3]float64
}

// PredictRanges evaluates range-only requests through the native precise
// sample route. The request and option values are copied before entering C.
func (s *PreciseEphemerisSamples) PredictRanges(requests []PredictRequest, options *ObservablesOptions) ([]RangePrediction, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	nativeRequests := make([]native.NativePredictRequest, len(requests))
	for i, request := range requests {
		nativeRequests[i] = native.NativePredictRequest{SatelliteID: request.SatelliteID, ReceiverECEF: [3]float64{request.ReceiverECEF.X, request.ReceiverECEF.Y, request.ReceiverECEF.Z}, TRxJ2000S: request.TRxJ2000S}
	}
	var nativeOptions *native.NativeObservablesOptions
	if options != nil {
		nativeOptions = &native.NativeObservablesOptions{CarrierHz: options.CarrierHz, LightTime: options.LightTime, Sagnac: options.Sagnac}
	}
	values, err := s.handle.PredictRanges(nativeRequests, nativeOptions)
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]RangePrediction, len(values))
	for i, value := range values {
		out[i] = RangePrediction{GeometricRangeM: value.GeometricRangeM, HasSatelliteClock: value.HasSatelliteClock, SatelliteClockS: value.SatelliteClockS, TransmitTimeJ2000S: value.TransmitTimeJ2000S, SatellitePositionECEF: value.SatellitePositionECEF}
	}
	return out, nil
}

// PreciseEphemerisInterpolant is a cached, immutable interpolator.
type PreciseEphemerisInterpolant struct {
	_      noCopy
	handle *native.PreciseEphemerisInterpolant
}

// BuildPreciseEphemerisInterpolantFromSP3 builds a native-owned interpolant from an SP3.
func BuildPreciseEphemerisInterpolantFromSP3(sp3 *SP3) (*PreciseEphemerisInterpolant, error) {
	if sp3 == nil || sp3.handle == nil {
		return nil, ErrClosed
	}
	h, err := native.PreciseInterpolantFromSP3(sp3.handle)
	if err != nil {
		return nil, publicError(err)
	}
	return &PreciseEphemerisInterpolant{handle: h}, nil
}

// BuildPreciseEphemerisInterpolantFromSamples builds an interpolant from a sample source.
func BuildPreciseEphemerisInterpolantFromSamples(samples *PreciseEphemerisSamples) (*PreciseEphemerisInterpolant, error) {
	if samples == nil || samples.handle == nil {
		return nil, ErrClosed
	}
	h, err := samples.handle.Interpolant()
	if err != nil {
		return nil, publicError(err)
	}
	return &PreciseEphemerisInterpolant{handle: h}, nil
}

// BuildPreciseEphemerisInterpolant builds an interpolant from copied canonical samples.
func BuildPreciseEphemerisInterpolant(values []PreciseEphemerisSample) (*PreciseEphemerisInterpolant, error) {
	nativeValues := make([]native.PreciseEphemerisSample, len(values))
	for i := range values {
		nativeValues[i] = nativePreciseSample(values[i])
	}
	h, err := native.PreciseInterpolantFromSamples(nativeValues)
	if err != nil {
		return nil, publicError(err)
	}
	return &PreciseEphemerisInterpolant{handle: h}, nil
}

// Close releases the native interpolant and is safe to call repeatedly.
func (i *PreciseEphemerisInterpolant) Close() error {
	if i == nil || i.handle == nil {
		return nil
	}
	return publicError(i.handle.Close())
}

// ObservableStates evaluates per-satellite J2000 epochs through the interpolant.
func (i *PreciseEphemerisInterpolant) ObservableStates(satellites []string, epochs []float64) ([]ObservableStateRow, error) {
	if i == nil || i.handle == nil {
		return nil, ErrClosed
	}
	rows, err := i.handle.ObservableStates(append([]string(nil), satellites...), append([]float64(nil), epochs...))
	if err != nil {
		return nil, publicError(err)
	}
	return publicObservableStates(rows), nil
}

// ObservableStatesShared evaluates all requested satellites at one J2000 epoch.
func (i *PreciseEphemerisInterpolant) ObservableStatesShared(satellites []string, epoch float64) ([]ObservableStateRow, error) {
	if i == nil || i.handle == nil {
		return nil, ErrClosed
	}
	rows, err := i.handle.ObservableStatesShared(append([]string(nil), satellites...), epoch)
	if err != nil {
		return nil, publicError(err)
	}
	return publicObservableStates(rows), nil
}

// PreciseInterpolantArtifact is opened from copied bytes. Pathname-based C
// constructors are intentionally absent: Go owns filesystem acquisition.
type PreciseInterpolantArtifact struct {
	_      noCopy
	handle *native.PreciseInterpolantArtifact
}

// OpenPreciseInterpolantArtifact opens a copied artifact byte stream.
func OpenPreciseInterpolantArtifact(data []byte) (*PreciseInterpolantArtifact, PreciseInterpolantArtifactError, error) {
	h, kind, err := native.OpenPreciseInterpolantArtifact(append([]byte(nil), data...))
	if err != nil {
		return nil, PreciseInterpolantArtifactError(kind), publicError(err)
	}
	return &PreciseInterpolantArtifact{handle: h}, PreciseInterpolantArtifactError(kind), nil
}

// OpenPreciseInterpolantArtifactBorrowed has the C ABI's borrowed spelling,
// but safely composes it onto the owned byte route because a Go slice cannot
// be retained by an owning native source handle.
func OpenPreciseInterpolantArtifactBorrowed(data []byte) (*PreciseInterpolantArtifact, PreciseInterpolantArtifactError, error) {
	return OpenPreciseInterpolantArtifact(data)
}

// OpenPreciseInterpolantArtifactFile keeps filesystem acquisition in Go and
// passes only the detached bytes through the C ABI.
func OpenPreciseInterpolantArtifactFile(path string) (*PreciseInterpolantArtifact, PreciseInterpolantArtifactError, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, err
	}
	return OpenPreciseInterpolantArtifact(data)
}

// Close releases the native artifact and is safe to call concurrently and repeatedly.
func (a *PreciseInterpolantArtifact) Close() error {
	if a == nil || a.handle == nil {
		return nil
	}
	return publicError(a.handle.Close())
}

// Checksum64 returns the checksum carried by the opened artifact.
func (a *PreciseInterpolantArtifact) Checksum64() (uint64, error) {
	if a == nil || a.handle == nil {
		return 0, ErrClosed
	}
	v, err := a.handle.Checksum64()
	return v, publicError(err)
}

// PreciseInterpolantArtifactChecksum64 computes an artifact checksum from bytes.
func PreciseInterpolantArtifactChecksum64(data []byte) (uint64, error) {
	v, err := native.PreciseArtifactChecksum64(append([]byte(nil), data...))
	return v, publicError(err)
}

// DigestProvenance reports whether the artifact digest is verified or attested.
func (a *PreciseInterpolantArtifact) DigestProvenance() (DigestProvenance, error) {
	if a == nil || a.handle == nil {
		return 0, ErrClosed
	}
	v, err := a.handle.DigestProvenance()
	return DigestProvenance(v), publicError(err)
}

// Verify verifies an opened artifact payload and returns its native error kind.
func (a *PreciseInterpolantArtifact) Verify() (PreciseInterpolantArtifactError, error) {
	if a == nil || a.handle == nil {
		return 0, ErrClosed
	}
	v, err := a.handle.Verify()
	return PreciseInterpolantArtifactError(v), publicError(err)
}

// State evaluates one satellite at one J2000 epoch from the artifact.
func (a *PreciseInterpolantArtifact) State(satellite string, epoch float64) (SP3State, error) {
	if a == nil || a.handle == nil {
		return SP3State{}, ErrClosed
	}
	v, err := a.handle.State(satellite, epoch)
	if err != nil {
		return SP3State{}, publicError(err)
	}
	return SP3State{PositionM: v.PositionM, HasClock: v.HasClock, ClockS: v.ClockS, HasVelocity: v.HasVelocity, VelocityMPerS: v.VelocityMPerS, HasClockRate: v.HasClockRate, ClockRateSPerS: v.ClockRateSPerS, ClockEvent: v.ClockEvent, ClockPredicted: v.ClockPredicted, Maneuver: v.Maneuver, OrbitPredicted: v.OrbitPredicted}, nil
}

// Satellites returns detached satellite identifiers present in the artifact.
func (a *PreciseInterpolantArtifact) Satellites() ([]string, error) {
	if a == nil || a.handle == nil {
		return nil, ErrClosed
	}
	v, err := a.handle.Satellites()
	return append([]string(nil), v...), publicError(err)
}

// PreciseInterpolantArtifactError identifies a native artifact validation result.
type PreciseInterpolantArtifactError uint32

const (
	// PreciseInterpolantArtifactErrorNone indicates no artifact error.
	PreciseInterpolantArtifactErrorNone PreciseInterpolantArtifactError = PreciseInterpolantArtifactError(native.PreciseInterpolantArtifactErrorNoneValue)
	// PreciseInterpolantArtifactErrorTruncated indicates incomplete artifact bytes.
	PreciseInterpolantArtifactErrorTruncated PreciseInterpolantArtifactError = PreciseInterpolantArtifactError(native.PreciseInterpolantArtifactErrorTruncatedValue)
	// PreciseInterpolantArtifactErrorCorrupt indicates a checksum mismatch.
	PreciseInterpolantArtifactErrorCorrupt PreciseInterpolantArtifactError = PreciseInterpolantArtifactError(native.PreciseInterpolantArtifactErrorCorruptValue)
	// PreciseInterpolantArtifactErrorParse indicates another parse failure.
	PreciseInterpolantArtifactErrorParse PreciseInterpolantArtifactError = PreciseInterpolantArtifactError(native.PreciseInterpolantArtifactErrorParseValue)
	// PreciseInterpolantArtifactErrorUnsupportedVersion indicates an unsupported version.
	PreciseInterpolantArtifactErrorUnsupportedVersion PreciseInterpolantArtifactError = PreciseInterpolantArtifactError(native.PreciseInterpolantArtifactErrorUnsupportedVersionValue)
	// PreciseInterpolantArtifactErrorUnsupportedTimeScale indicates an unsupported time scale.
	PreciseInterpolantArtifactErrorUnsupportedTimeScale PreciseInterpolantArtifactError = PreciseInterpolantArtifactError(native.PreciseInterpolantArtifactErrorUnsupportedTimeScaleValue)
	// PreciseInterpolantArtifactErrorUnsupportedSatelliteSystem indicates an unsupported system.
	PreciseInterpolantArtifactErrorUnsupportedSatelliteSystem PreciseInterpolantArtifactError = PreciseInterpolantArtifactError(native.PreciseInterpolantArtifactErrorUnsupportedSatelliteSystemValue)
	// PreciseInterpolantArtifactErrorDuplicateSatellite indicates a duplicate index entry.
	PreciseInterpolantArtifactErrorDuplicateSatellite PreciseInterpolantArtifactError = PreciseInterpolantArtifactError(native.PreciseInterpolantArtifactErrorDuplicateSatelliteValue)
	// PreciseInterpolantArtifactErrorIO indicates a native artifact I/O failure.
	PreciseInterpolantArtifactErrorIO PreciseInterpolantArtifactError = PreciseInterpolantArtifactError(native.PreciseInterpolantArtifactErrorIOValue)
	// PreciseInterpolantArtifactErrorAttestedChecksumMismatch indicates an attestation mismatch.
	PreciseInterpolantArtifactErrorAttestedChecksumMismatch PreciseInterpolantArtifactError = PreciseInterpolantArtifactError(native.PreciseInterpolantArtifactErrorAttestedChecksumMismatchValue)
)

func publicObservableStates(values []native.NativeObservableStateRow) []ObservableStateRow {
	out := make([]ObservableStateRow, len(values))
	for i, v := range values {
		out[i] = ObservableStateRow{PositionECEFM: v.Position, ClockS: v.ClockS, HasClock: v.HasClock, ElementStatus: ObservableStateElementStatus(v.ElementStatus), ResultStatus: StatusCode(v.ResultStatus)}
	}
	return out
}

// DeclaredEpochCount returns the epoch count declared in the SP3 header.
func (s *SP3) DeclaredEpochCount() (uint64, error) {
	if s == nil || s.handle == nil {
		return 0, ErrClosed
	}
	v, err := s.handle.DeclaredEpochCount()
	return v, publicError(err)
}

// DeclaredStart returns whether the header declares a start epoch and its J2000 seconds.
func (s *SP3) DeclaredStart() (bool, float64, error) {
	if s == nil || s.handle == nil {
		return false, 0, ErrClosed
	}
	p, v, err := s.handle.DeclaredStart()
	return p, v, publicError(err)
}

// EpochPrediction returns observed/predicted metadata for one parsed epoch.
func (s *SP3) EpochPrediction(index int) (SP3EpochPrediction, error) {
	if s == nil || s.handle == nil {
		return SP3EpochPrediction{}, ErrClosed
	}
	v, err := s.handle.EpochPrediction(index)
	if err != nil {
		return SP3EpochPrediction{}, publicError(err)
	}
	return SP3EpochPrediction{EpochJ2000S: v.EpochJ2000S, Observed: v.Observed, OrbitPredictedSatelliteCount: v.OrbitPredictedSatelliteCount, ClockPredictedSatelliteCount: v.ClockPredictedSatelliteCount}, nil
}

// SP3Continuity contains native continuity-check counts.
type SP3Continuity struct {
	// Defects is the number of continuity defects detected.
	Defects int
	// ResidualsChecked is the number of residual rows checked.
	ResidualsChecked int
	// ResidualsSkipped is the number of residual rows skipped.
	ResidualsSkipped int
}

// Continuity reports native physical-continuity and residual-check counts.
func (s *SP3) Continuity(orbitClass int, residualToleranceM float64) (SP3Continuity, error) {
	if s == nil || s.handle == nil {
		return SP3Continuity{}, ErrClosed
	}
	v, err := s.handle.Continuity(orbitClass, residualToleranceM)
	if err != nil {
		return SP3Continuity{}, publicError(err)
	}
	return SP3Continuity{Defects: v.Defects, ResidualsChecked: v.ResidualsChecked, ResidualsSkipped: v.ResidualsSkipped}, nil
}

// SP3ClockReferenceOffset contains one J2000 clock-datum offset estimate.
type SP3ClockReferenceOffset struct {
	// EpochJ2000S is the epoch in seconds from J2000.
	EpochJ2000S float64
	// OffsetS is the offset s in seconds.
	OffsetS float64
	// Satellites is the number of satellites contributing to the clock-reference offset.
	Satellites int
}

// ClockReferenceOffsets estimates the other product's clock offset by epoch.
func (s *SP3) ClockReferenceOffsets(other *SP3, minCommon int) ([]SP3ClockReferenceOffset, error) {
	if s == nil || s.handle == nil || other == nil || other.handle == nil {
		return nil, ErrClosed
	}
	values, err := s.handle.ClockReferenceOffsets(other.handle, minCommon)
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]SP3ClockReferenceOffset, len(values))
	for i, value := range values {
		out[i] = SP3ClockReferenceOffset{EpochJ2000S: value.EpochJ2000S, OffsetS: value.OffsetS, Satellites: value.Satellites}
	}
	return out, nil
}

// AlignClockReference returns a copied SP3 with clocks shifted to the reference datum.
func (s *SP3) AlignClockReference(other *SP3, minCommon int) (*SP3, error) {
	if s == nil || s.handle == nil || other == nil || other.handle == nil {
		return nil, ErrClosed
	}
	h, err := s.handle.AlignClockReference(other.handle, minCommon)
	if err != nil {
		return nil, publicError(err)
	}
	return &SP3{handle: h}, nil
}

// Interpolate evaluates one satellite at J2000 epochs, returning positions,
// clocks, and the native number of written rows.
func (s *SP3) Interpolate(satellite string, epochs []float64) ([][3]float64, []float64, int, error) {
	if s == nil || s.handle == nil {
		return nil, nil, 0, ErrClosed
	}
	positions, clocks, written, err := s.handle.Interpolate(satellite, append([]float64(nil), epochs...))
	return positions, clocks, written, publicError(err)
}

// ContinuityVerdictJSON returns the native continuity-window decision as JSON bytes.
func (s *SP3) ContinuityVerdictJSON(orbitClass int, residualToleranceM, fromJ2000S, throughJ2000S float64) ([]byte, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.ContinuityVerdictJSON(orbitClass, residualToleranceM, fromJ2000S, throughJ2000S)
	return append([]byte(nil), value...), publicError(err)
}

// Text returns detached canonical SP3 text bytes.
func (s *SP3) Text() ([]byte, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	v, err := s.handle.Text()
	return v, publicError(err)
}

// EphemerisSample samples the SP3 source on a regular J2000 grid.
func (s *SP3) EphemerisSample(satellites []string, start, stop, step float64) ([]EphemerisSampleRow, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	rows, err := s.handle.EphemerisSample(append([]string(nil), satellites...), start, stop, step)
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]EphemerisSampleRow, len(rows))
	for i, row := range rows {
		out[i] = EphemerisSampleRow{SatelliteID: row.SatelliteID, EpochJ2000S: row.EpochJ2000S, Status: EphemerisSampleStatus(row.Status), HasPosition: row.HasPosition, PositionECEFM: row.Position, HasClock: row.HasClock, ClockS: row.ClockS}
	}
	return out, nil
}

// PreciseSamples returns detached canonical precise samples from the SP3 source.
func (s *SP3) PreciseSamples() ([]PreciseEphemerisSample, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	values, err := s.handle.PreciseSamples()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]PreciseEphemerisSample, len(values))
	for i, v := range values {
		out[i] = PreciseEphemerisSample{Satellite: v.Satellite, TimeScale: TimeScale(v.TimeScale), EpochJ2000S: v.EpochJ2000S, PositionECEFM: v.PositionECEFM, HasClock: v.HasClock, ClockS: v.ClockS, ClockEvent: v.ClockEvent}
	}
	return out, nil
}

// ArtifactBytes returns detached precise-interpolant artifact bytes and its native kind.
func (s *SP3) ArtifactBytes() ([]byte, PreciseInterpolantArtifactError, error) {
	if s == nil || s.handle == nil {
		return nil, 0, ErrClosed
	}
	v, kind, err := s.handle.ArtifactBytes()
	return append([]byte(nil), v...), PreciseInterpolantArtifactError(kind), publicError(err)
}

// SP3EpochPrediction contains one epoch's observed/predicted metadata.
type SP3EpochPrediction struct {
	// EpochJ2000S is the epoch in seconds from J2000.
	EpochJ2000S float64
	// Observed is the observed epoch value in the product time scale.
	Observed bool
	// OrbitPredictedSatelliteCount is the number of satellites with predicted orbit values.
	OrbitPredictedSatelliteCount int
	// ClockPredictedSatelliteCount is the number of satellites with predicted clock values.
	ClockPredictedSatelliteCount int
}

// CatalogIdentityCacheKey derives the native stable key for an exact identity.
func CatalogIdentityCacheKey(identity ProductIdentity) ([]byte, error) {
	nativeIdentity, err := identityToNative(identity)
	if err != nil {
		return nil, err
	}
	v, err := native.CatalogIdentityCacheKey(nativeIdentity)
	return v, publicError(err)
}

// ValidateExactProductSet checks that available identities exactly match expected.
func ValidateExactProductSet(expected, available []ProductIdentity) error {
	a, b := make([]native.ProductIdentity, len(expected)), make([]native.ProductIdentity, len(available))
	for i := range expected {
		value, err := identityToNative(expected[i])
		if err != nil {
			return err
		}
		a[i] = value
	}
	for i := range available {
		value, err := identityToNative(available[i])
		if err != nil {
			return err
		}
		b[i] = value
	}
	return publicError(native.CatalogValidateExactProductSet(a, b))
}

// SP3ContentStartConventionKind identifies a native SP3 content-start convention.
type SP3ContentStartConventionKind uint32

// SP3ContentStartConventionFor returns the native convention and byte offset.
func SP3ContentStartConventionFor(center string, year int, month, day uint8, issue string) (SP3ContentStartConventionKind, int64, error) {
	c, o, err := native.CatalogSP3ContentStartConvention(center, year, month, day, issue)
	return SP3ContentStartConventionKind(c), o, publicError(err)
}

// SP3ContentStartConvention is an alias for SP3ContentStartConventionFor.
func SP3ContentStartConvention(center string, year int, month, day uint8, issue string) (SP3ContentStartConventionKind, int64, error) {
	return SP3ContentStartConventionFor(center, year, month, day, issue)
}

// StalenessPolicy contains the maximum allowed age in seconds.
type StalenessPolicy struct {
	// MaxStalenessS is the maximum allowed age in seconds.
	MaxStalenessS float64
}

// StalenessPolicyDays constructs a native staleness policy from days.
func StalenessPolicyDays(days float64) StalenessPolicy {
	v := native.StalenessPolicyDays(days)
	return StalenessPolicy{MaxStalenessS: v.MaxStalenessS}
}

// StalenessPolicySeconds constructs a native staleness policy from seconds.
func StalenessPolicySeconds(seconds float64) StalenessPolicy {
	v := native.StalenessPolicySeconds(seconds)
	return StalenessPolicy{MaxStalenessS: v.MaxStalenessS}
}

// StalenessPolicyDefault returns the native default staleness policy.
func StalenessPolicyDefault() StalenessPolicy {
	v := native.StalenessPolicyDefault()
	return StalenessPolicy{MaxStalenessS: v.MaxStalenessS}
}

func identityToNative(value ProductIdentity) (native.ProductIdentity, error) {
	year := value.Date.Year()
	if year < -1<<31 || year > 1<<31-1 {
		return native.ProductIdentity{}, fmt.Errorf("sidereon: product identity year %d does not fit in int32", year)
	}
	return native.ProductIdentity{Family: native.ProductFamily(value.Family), AnalysisCenter: value.AnalysisCenter, Publisher: native.ProductPublisher(value.Publisher), SolutionClass: native.SolutionClass(value.SolutionClass), Campaign: native.ProductCampaign(value.Campaign), FilenameVersion: value.FilenameVersion, Year: int32(year), Month: uint8(value.Date.Month()), Day: uint8(value.Date.Day()), HasIssue: value.HasIssue, Issue: value.Issue, Span: value.Span, Sample: value.Sample, OfficialFilename: value.OfficialFilename, Format: native.ProductFormat(value.Format), HasFormatVersion: value.HasFormatVersion, FormatVersion: value.FormatVersion, HasPredictionHorizonDays: value.HasPredictionHorizonDays, PredictionHorizonDays: value.PredictionHorizonDays}, nil
}
