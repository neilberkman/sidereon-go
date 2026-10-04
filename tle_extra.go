package sidereon

import (
	"time"

	"sidereon.dev/go/v3/internal/native"
)

// TLEFilePolicy selects checksum handling while parsing a TLE file.
type TLEFilePolicy uint32

const (
	// TLEFilePolicyStrict rejects checksum mismatches and missing checksum digits.
	TLEFilePolicyStrict TLEFilePolicy = 0
	// TLEFilePolicyLenient accepts checksum problems and reports warnings per satellite.
	TLEFilePolicyLenient TLEFilePolicy = 1
)

// TLERecordIssue identifies why a nonblank input record was rejected.
type TLERecordIssue uint32

const (
	// TLERecordIssueInvalid reports a line pair refused by grammar, checksum, or SGP4.
	TLERecordIssueInvalid TLERecordIssue = 0
	// TLERecordIssueMissingLine2 reports a line 1 with no following line 2.
	TLERecordIssueMissingLine2 TLERecordIssue = 1
	// TLERecordIssueOrphanLine2 reports a line 2 with no preceding line 1.
	TLERecordIssueOrphanLine2 TLERecordIssue = 2
	// TLERecordIssueOrphanName reports a name line with no following element set.
	TLERecordIssueOrphanName TLERecordIssue = 3
)

// TLERejectedRecord is an owned description of one rejected file row.
type TLERejectedRecord struct {
	// LineNumber is the one-based first line belonging to the rejected record.
	LineNumber int
	// Issue is the native rejection category; unknown future tags are retained.
	Issue TLERecordIssue
	// Name is the associated name line, if present.
	Name string
	// Error contains native refusal text for invalid element sets.
	Error string
}

// TLEFile owns a parsed multi-record TLE file. Its Satellite results are
// independent TLE handles and remain valid after the file is closed.
type TLEFile struct {
	_      noCopy
	handle *native.TLEFile
}

// ParseTLEFile parses the supplied representation as a TLE file.
func ParseTLEFile(text []byte) (*TLEFile, error) {
	return ParseTLEFileWithOpsMode(text, OpsModeAFSPC)
}

// ParseTLEFileWithOpsMode parses the supplied representation as a TLE file with the selected OpsMode.
func ParseTLEFileWithOpsMode(text []byte, mode OpsMode) (*TLEFile, error) {
	return ParseTLEFileWithOpsModeAndPolicy(text, mode, TLEFilePolicyStrict)
}

// ParseTLEFileWithPolicy parses with the selected checksum policy and default AFSPC operations mode.
func ParseTLEFileWithPolicy(text []byte, policy TLEFilePolicy) (*TLEFile, error) {
	return ParseTLEFileWithOpsModeAndPolicy(text, OpsModeAFSPC, policy)
}

// ParseTLEFileWithOpsModeAndPolicy parses with explicit operations mode and checksum policy.
func ParseTLEFileWithOpsModeAndPolicy(text []byte, mode OpsMode, policy TLEFilePolicy) (*TLEFile, error) {
	if policy != TLEFilePolicyStrict && policy != TLEFilePolicyLenient {
		return nil, invalidArgument("TLE file policy is not defined by the C ABI")
	}
	handle, err := native.ParseTLEFileWithPolicy(append([]byte(nil), text...), uint32(mode), uint32(policy))
	if err != nil {
		return nil, publicError(err)
	}
	if handle == nil {
		return nil, errNilNativeHandle
	}
	return &TLEFile{handle: handle}, nil
}

// Close releases the native TLEFile resource and is safe to call repeatedly.
func (f *TLEFile) Close() error {
	if f == nil || f.handle == nil {
		return nil
	}
	return publicError(f.handle.Close())
}

// Count returns the number of parsed TLE records in the file.
func (f *TLEFile) Count() (int, error) {
	if f == nil || f.handle == nil {
		return 0, ErrClosed
	}
	value, err := f.handle.Count()
	return value, publicError(err)
}

// Skipped returns the number of malformed or otherwise skipped TLE records.
func (f *TLEFile) Skipped() (int, error) {
	if f == nil || f.handle == nil {
		return 0, ErrClosed
	}
	value, err := f.handle.Skipped()
	return value, publicError(err)
}

// LineNumber returns the one-based source line of the accepted record at index.
func (f *TLEFile) LineNumber(index int) (int, error) {
	if f == nil || f.handle == nil {
		return 0, ErrClosed
	}
	value, err := f.handle.LineNumber(index)
	return value, publicError(err)
}

// Name returns the detached name line associated with a parsed TLE record.
func (f *TLEFile) Name(index int) (string, error) {
	if f == nil || f.handle == nil {
		return "", ErrClosed
	}
	value, err := f.handle.Name(index)
	return value, publicError(err)
}

// Rejected returns detached location and refusal details for a skipped record.
func (f *TLEFile) Rejected(index int) (TLERejectedRecord, error) {
	if f == nil || f.handle == nil {
		return TLERejectedRecord{}, ErrClosed
	}
	value, err := f.handle.Rejected(index)
	return TLERejectedRecord{LineNumber: value.LineNumber, Issue: TLERecordIssue(value.Issue), Name: value.Name, Error: value.Error}, publicError(err)
}

// Satellite returns an independent TLE handle for the parsed record at index.
func (f *TLEFile) Satellite(index int) (*TLE, error) {
	if f == nil || f.handle == nil {
		return nil, ErrClosed
	}
	value, err := f.handle.Satellite(index)
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errNilNativeHandle
	}
	return &TLE{handle: value}, nil
}

// TLEChecksumWarning records a checksum issue for one TLE line.
type TLEChecksumWarning struct {
	LineNumber uint8
	// Kind is the native TLE checksum warning variant.
	Kind uint32
	// Found is the byte in column 69, or zero when the checksum is missing.
	Found uint8
	// Computed is the checksum recomputed from columns 1 through 68.
	Computed uint8
	// Expected is a compatibility alias for Found.
	Expected uint8
}

// ChecksumWarnings returns detached checksum-mismatch diagnostics for this TLE.
func (t *TLE) ChecksumWarnings() ([]TLEChecksumWarning, error) {
	if t == nil || t.handle == nil {
		return nil, ErrClosed
	}
	values, err := t.handle.ChecksumWarnings()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]TLEChecksumWarning, len(values))
	for i := range out {
		out[i] = TLEChecksumWarning{LineNumber: values[i].LineNumber, Kind: values[i].Kind, Found: values[i].Found, Expected: values[i].Expected, Computed: values[i].Computed}
	}
	return out, nil
}

// TLEBatchPropagation owns a satellite-major matrix of TEME states. Each
// state carries the corresponding UTC epoch in EpochJ2000S.
type TLEBatchPropagation struct {
	_      noCopy
	handle *native.TLEBatchPropagation
}

// TLEPair contains independent copies of the two TLE lines.
type TLEPair struct {
	// Line1 and Line2 are independent copies of the two TLE lines.
	Line1, Line2 string
}

func nativeTLEPairs(values []TLEPair) []native.TLEPair {
	out := make([]native.TLEPair, len(values))
	for i := range out {
		out[i] = native.TLEPair{Line1: values[i].Line1, Line2: values[i].Line2}
	}
	return out
}

// PropagateTLEBatch computes detached TEME states for each supplied TLE and epoch.
func PropagateTLEBatch(pairs []TLEPair, epochs []time.Time, mode OpsMode, parallel bool) (*TLEBatchPropagation, error) {
	value, err := native.PropagateTLEBatch(nativeTLEPairs(append([]TLEPair(nil), pairs...)), append([]time.Time(nil), epochs...), uint32(mode), parallel)
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errNilNativeHandle
	}
	return &TLEBatchPropagation{handle: value}, nil
}

// Close releases the native TLEBatchPropagation resource and is safe to call repeatedly.
func (b *TLEBatchPropagation) Close() error {
	if b == nil || b.handle == nil {
		return nil
	}
	return publicError(b.handle.Close())
}

// Shape returns the TLE-batch output dimensions.
func (b *TLEBatchPropagation) Shape() (satelliteCount, epochCount int, err error) {
	if b == nil || b.handle == nil {
		return 0, 0, ErrClosed
	}
	satelliteCount, epochCount, err = b.handle.Shape()
	return satelliteCount, epochCount, publicError(err)
}

// States returns detached TEME states in TLE-major order.
func (b *TLEBatchPropagation) States() ([]TEMEState, error) {
	if b == nil || b.handle == nil {
		return nil, ErrClosed
	}
	values, err := b.handle.States()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]TEMEState, len(values))
	for i := range out {
		out[i] = TEMEState{EpochJ2000S: values[i].EpochJ2000S, PositionKm: values[i].PositionKm, VelocityKmPerS: values[i].VelocityKmPerS}
	}
	return out, nil
}

// TLEBatchLookAngles owns flattened satellite-major topocentric rows.
type TLEBatchLookAngles struct {
	_      noCopy
	handle *native.TLEBatchLookAngles
}

// BatchTLELookAngles returns detached topocentric look-angle rows for each TLE and epoch.
func BatchTLELookAngles(pairs []TLEPair, station PassStation, epochs []time.Time, mode OpsMode, parallel bool) (*TLEBatchLookAngles, error) {
	value, err := native.LookAnglesBatch(nativeTLEPairs(append([]TLEPair(nil), pairs...)), native.GroundStation{LatitudeDeg: station.LatitudeDeg, LongitudeDeg: station.LongitudeDeg, AltitudeM: station.AltitudeM}, append([]time.Time(nil), epochs...), uint32(mode), parallel)
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errNilNativeHandle
	}
	return &TLEBatchLookAngles{handle: value}, nil
}

// Close releases the native TLEBatchLookAngles resource and is safe to call repeatedly.
func (b *TLEBatchLookAngles) Close() error {
	if b == nil || b.handle == nil {
		return nil
	}
	return publicError(b.handle.Close())
}

// Shape returns the look-angle output dimensions.
func (b *TLEBatchLookAngles) Shape() (satelliteCount, epochCount int, err error) {
	if b == nil || b.handle == nil {
		return 0, 0, ErrClosed
	}
	satelliteCount, epochCount, err = b.handle.Shape()
	return satelliteCount, epochCount, publicError(err)
}

// Values returns detached look-angle rows in TLE-major order.
func (b *TLEBatchLookAngles) Values() ([]Topocentric, error) {
	if b == nil || b.handle == nil {
		return nil, ErrClosed
	}
	values, err := b.handle.Values()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]Topocentric, len(values))
	for i := range out {
		out[i] = Topocentric{AzimuthDeg: values[i].AzimuthDeg, ElevationDeg: values[i].ElevationDeg, RangeKm: values[i].RangeKm}
	}
	return out, nil
}

// SGP4DecayLatch owns a native SGP4 decay-latch state and its failure epoch.
type SGP4DecayLatch struct {
	_      noCopy
	handle *native.SGP4DecayLatch
}

// NewSGP4DecayLatch constructs an SGP4 decay latch from the supplied configuration.
func NewSGP4DecayLatch() (*SGP4DecayLatch, error) {
	value, err := native.NewSGP4DecayLatch()
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errNilNativeHandle
	}
	return &SGP4DecayLatch{handle: value}, nil
}

// Close releases the native SGP4DecayLatch resource and is safe to call repeatedly.
func (l *SGP4DecayLatch) Close() error {
	if l == nil || l.handle == nil {
		return nil
	}
	return publicError(l.handle.Close())
}

// Clear clears the latch state.
func (l *SGP4DecayLatch) Clear() error {
	if l == nil || l.handle == nil {
		return ErrClosed
	}
	return publicError(l.handle.Clear())
}

// FirstFailingEpoch returns the first epoch at which the latch observed decay.
func (l *SGP4DecayLatch) FirstFailingEpoch() (minutesSinceEpoch float64, present bool, err error) {
	if l == nil || l.handle == nil {
		return 0, false, ErrClosed
	}
	minutesSinceEpoch, present, err = l.handle.FirstFailingEpoch()
	return minutesSinceEpoch, present, publicError(err)
}

// PropagateWithDecayLatch computes a TEME state and latches the first decay epoch.
func (t *TLE) PropagateWithDecayLatch(minutesSinceEpoch float64, latch *SGP4DecayLatch) (TEMEState, error) {
	if t == nil || t.handle == nil {
		return TEMEState{}, ErrClosed
	}
	if latch == nil || latch.handle == nil {
		return TEMEState{}, ErrClosed
	}
	value, err := t.handle.PropagateWithDecayLatch(minutesSinceEpoch, latch.handle)
	return TEMEState{PositionKm: value.PositionKm, VelocityKmPerS: value.VelocityKmPerS}, publicError(err)
}

// VisibleSatellite contains one visible satellite and its look-angle interval.
type VisibleSatellite struct {
	// CatalogNumber is the NORAD catalog identifier of the visible satellite.
	CatalogNumber string
	// AzimuthDeg is the azimuth deg in degrees; ElevationDeg is the elevation deg in degrees; RangeKm is the range km in kilometres.
	AzimuthDeg, ElevationDeg, RangeKm float64
	// PositionKm is the position km in kilometres.
	PositionKm [3]float64
}

// VisibleList owns a native list of visible satellites and returns detached copies.
type VisibleList struct {
	_      noCopy
	handle *native.VisibleList
}

// VisibleSatellites returns detached satellites whose look-angle intervals meet the visibility limits.
func VisibleSatellites(tles []*TLE, catalogNumbers []string, station PassStation, epoch time.Time, minElevationDeg float64) (*VisibleList, error) {
	nativeTLEs := make([]*native.TLE, len(tles))
	for i, tle := range tles {
		if tle == nil || tle.handle == nil {
			return nil, ErrClosed
		}
		nativeTLEs[i] = tle.handle
	}
	value, err := native.VisibleFromSatellites(nativeTLEs, append([]string(nil), catalogNumbers...), native.GroundStation{LatitudeDeg: station.LatitudeDeg, LongitudeDeg: station.LongitudeDeg, AltitudeM: station.AltitudeM}, epoch, minElevationDeg)
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errNilNativeHandle
	}
	return &VisibleList{handle: value}, nil
}

// Close releases the native VisibleList resource and is safe to call repeatedly.
func (v *VisibleList) Close() error {
	if v == nil || v.handle == nil {
		return nil
	}
	return publicError(v.handle.Close())
}

// Count returns the number of visible-satellite records in the list.
func (v *VisibleList) Count() (int, error) {
	if v == nil || v.handle == nil {
		return 0, ErrClosed
	}
	value, err := v.handle.Count()
	return value, publicError(err)
}

// Values returns detached visible-satellite records in native order.
func (v *VisibleList) Values() ([]VisibleSatellite, error) {
	if v == nil || v.handle == nil {
		return nil, ErrClosed
	}
	values, err := v.handle.Values()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]VisibleSatellite, len(values))
	for i := range out {
		out[i] = VisibleSatellite{CatalogNumber: values[i].CatalogNumber, AzimuthDeg: values[i].AzimuthDeg, ElevationDeg: values[i].ElevationDeg, RangeKm: values[i].RangeKm, PositionKm: values[i].PositionKm}
	}
	return out, nil
}
