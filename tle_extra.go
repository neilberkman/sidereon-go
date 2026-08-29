package sidereon

import (
	"time"

	"github.com/neilberkman/sidereon-go/internal/native"
)

// TLEFile owns a parsed multi-record TLE file. Its Satellite results are
// independent TLE handles and remain valid after the file is closed.
type TLEFile struct {
	_      noCopy
	handle *native.TLEFile
}

func ParseTLEFile(text []byte) (*TLEFile, error) {
	return ParseTLEFileWithOpsMode(text, OpsModeAFSPC)
}

func ParseTLEFileWithOpsMode(text []byte, mode OpsMode) (*TLEFile, error) {
	handle, err := native.ParseTLEFile(append([]byte(nil), text...), uint32(mode))
	if err != nil {
		return nil, publicError(err)
	}
	if handle == nil {
		return nil, errNilNativeHandle
	}
	return &TLEFile{handle: handle}, nil
}

func (f *TLEFile) Close() error {
	if f == nil || f.handle == nil {
		return nil
	}
	return publicError(f.handle.Close())
}
func (f *TLEFile) Count() (int, error) {
	if f == nil || f.handle == nil {
		return 0, ErrClosed
	}
	value, err := f.handle.Count()
	return value, publicError(err)
}
func (f *TLEFile) Skipped() (int, error) {
	if f == nil || f.handle == nil {
		return 0, ErrClosed
	}
	value, err := f.handle.Skipped()
	return value, publicError(err)
}
func (f *TLEFile) Name(index int) (string, error) {
	if f == nil || f.handle == nil {
		return "", ErrClosed
	}
	value, err := f.handle.Name(index)
	return value, publicError(err)
}
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

type TLEChecksumWarning struct{ LineNumber, Expected, Computed uint8 }

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
		out[i] = TLEChecksumWarning{LineNumber: values[i].LineNumber, Expected: values[i].Expected, Computed: values[i].Computed}
	}
	return out, nil
}

// TLEBatchPropagation owns a satellite-major matrix of TEME states. Each
// state carries the corresponding UTC epoch in EpochJ2000S.
type TLEBatchPropagation struct {
	_      noCopy
	handle *native.TLEBatchPropagation
}
type TLEPair struct{ Line1, Line2 string }

func nativeTLEPairs(values []TLEPair) []native.TLEPair {
	out := make([]native.TLEPair, len(values))
	for i := range out {
		out[i] = native.TLEPair{Line1: values[i].Line1, Line2: values[i].Line2}
	}
	return out
}
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
func (b *TLEBatchPropagation) Close() error {
	if b == nil || b.handle == nil {
		return nil
	}
	return publicError(b.handle.Close())
}
func (b *TLEBatchPropagation) Shape() (satelliteCount, epochCount int, err error) {
	if b == nil || b.handle == nil {
		return 0, 0, ErrClosed
	}
	satelliteCount, epochCount, err = b.handle.Shape()
	return satelliteCount, epochCount, publicError(err)
}
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
func (b *TLEBatchLookAngles) Close() error {
	if b == nil || b.handle == nil {
		return nil
	}
	return publicError(b.handle.Close())
}
func (b *TLEBatchLookAngles) Shape() (satelliteCount, epochCount int, err error) {
	if b == nil || b.handle == nil {
		return 0, 0, ErrClosed
	}
	satelliteCount, epochCount, err = b.handle.Shape()
	return satelliteCount, epochCount, publicError(err)
}
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

type SGP4DecayLatch struct {
	_      noCopy
	handle *native.SGP4DecayLatch
}

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
func (l *SGP4DecayLatch) Close() error {
	if l == nil || l.handle == nil {
		return nil
	}
	return publicError(l.handle.Close())
}
func (l *SGP4DecayLatch) Clear() error {
	if l == nil || l.handle == nil {
		return ErrClosed
	}
	return publicError(l.handle.Clear())
}
func (l *SGP4DecayLatch) FirstFailingEpoch() (minutesSinceEpoch float64, present bool, err error) {
	if l == nil || l.handle == nil {
		return 0, false, ErrClosed
	}
	minutesSinceEpoch, present, err = l.handle.FirstFailingEpoch()
	return minutesSinceEpoch, present, publicError(err)
}
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

type VisibleSatellite struct {
	CatalogNumber                     string
	AzimuthDeg, ElevationDeg, RangeKm float64
	PositionKm                        [3]float64
}
type VisibleList struct {
	_      noCopy
	handle *native.VisibleList
}

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
func (v *VisibleList) Close() error {
	if v == nil || v.handle == nil {
		return nil
	}
	return publicError(v.handle.Close())
}
func (v *VisibleList) Count() (int, error) {
	if v == nil || v.handle == nil {
		return 0, ErrClosed
	}
	value, err := v.handle.Count()
	return value, publicError(err)
}
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
