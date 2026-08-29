package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// FrequencyChannel maps a GLONASS slot to its FDMA channel number.
type FrequencyChannel struct {
	// Slot is the slot value for FrequencyChannel.
	Slot uint8
	// Channel is the channel value for FrequencyChannel.
	Channel int8
}

// GLONASSRecord is one copied RINEX GLONASS state-vector record. Position is
// in metres, velocity in metres per second, acceleration in metres per second
// squared, and ToeUTCJ2000S is UTC J2000 seconds.
type GLONASSRecord struct {
	// SatelliteID identifies or counts this record.
	SatelliteID string
	// ToeUTCJ2000S is the toe utcj2000 s in seconds.
	ToeUTCJ2000S float64
	// PositionM is the position m in metres.
	PositionM [3]float64
	// VelocityMPerS is the velocity m per s in metres per second.
	VelocityMPerS [3]float64
	// AccelerationMPerS2 is the acceleration m per s2 in metres per second squared.
	AccelerationMPerS2 [3]float64
	// ClockBiasS is the clock bias s value for GLONASSRecord.
	ClockBiasS float64
	// GammaN is the gamma n value for GLONASSRecord.
	GammaN float64
	// SVHealth is the svhealth value for GLONASSRecord.
	SVHealth float64
	// FrequencyChannel is the frequency channel value for GLONASSRecord.
	FrequencyChannel int32
}

// SkippedGLONASSRecord preserves the raw token of an extended GLONASS slot
// that the core satellite identifier cannot represent.
type SkippedGLONASSRecord struct {
	// SatelliteID identifies or counts this record.
	SatelliteID string
}

// RINEXGLONASSRecords owns a parsed list of GLONASS state vectors and skipped
// extended-slot diagnostics. It may be read concurrently with Close.
type RINEXGLONASSRecords struct {
	_      noCopy
	handle *native.RinexGlonassRecords
}

// ParseRINEXGLONASSRecords parses RINEX GLONASS bytes through the C ABI.
func ParseRINEXGLONASSRecords(data []byte) (*RINEXGLONASSRecords, error) {
	handle, err := native.ParseRinexGlonassRecords(data)
	if err != nil {
		return nil, publicError(err)
	}
	return &RINEXGLONASSRecords{handle: handle}, nil
}

// Close releases the GLONASS record handle; repeated calls are safe.
func (r *RINEXGLONASSRecords) Close() error {
	if r == nil || r.handle == nil {
		return nil
	}
	return publicError(r.handle.Close())
}

// Count returns the number of copied GLONASS records.
func (r *RINEXGLONASSRecords) Count() (int, error) {
	if r == nil || r.handle == nil {
		return 0, ErrClosed
	}
	value, err := r.handle.Count()
	return value, publicError(err)
}

// Record returns one detached GLONASS record by zero-based index.
func (r *RINEXGLONASSRecords) Record(index int) (GLONASSRecord, error) {
	if r == nil || r.handle == nil {
		return GLONASSRecord{}, ErrClosed
	}
	value, err := r.handle.Record(index)
	return glonassRecordFromNative(value), publicError(err)
}

// SkippedCount returns the number of unsupported extended-slot diagnostics.
func (r *RINEXGLONASSRecords) SkippedCount() (int, error) {
	if r == nil || r.handle == nil {
		return 0, ErrClosed
	}
	value, err := r.handle.SkippedCount()
	return value, publicError(err)
}

// Skipped returns one detached skipped-record diagnostic by zero-based index.
func (r *RINEXGLONASSRecords) Skipped(index int) (SkippedGLONASSRecord, error) {
	if r == nil || r.handle == nil {
		return SkippedGLONASSRecord{}, ErrClosed
	}
	value, err := r.handle.Skipped(index)
	return SkippedGLONASSRecord{SatelliteID: value.SatelliteID}, publicError(err)
}

func glonassRecordFromNative(value native.NativeGlonassRecord) GLONASSRecord {
	return GLONASSRecord{SatelliteID: value.SatelliteID, ToeUTCJ2000S: value.ToeUTCJ2000S, PositionM: value.PositionM, VelocityMPerS: value.VelocityMPerS, AccelerationMPerS2: value.AccelerationMPerS2, ClockBiasS: value.ClockBiasS, GammaN: value.GammaN, SVHealth: value.SVHealth, FrequencyChannel: value.FrequencyChannel}
}

// GLONASSFrequencyChannels returns the native broadcast slot/channel map.
func (b *BroadcastEphemeris) GLONASSFrequencyChannels() ([]FrequencyChannel, error) {
	if b == nil || b.handle == nil {
		return nil, ErrClosed
	}
	values, err := b.handle.GlonassFrequencyChannels()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]FrequencyChannel, len(values))
	for i, value := range values {
		out[i] = FrequencyChannel{Slot: value.Slot, Channel: value.Channel}
	}
	return out, nil
}

// GLONASSRecords returns copied broadcast GLONASS state-vector records.
func (b *BroadcastEphemeris) GLONASSRecords() ([]GLONASSRecord, error) {
	if b == nil || b.handle == nil {
		return nil, ErrClosed
	}
	values, err := b.handle.GlonassRecords()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]GLONASSRecord, len(values))
	for i, value := range values {
		out[i] = glonassRecordFromNative(value)
	}
	return out, nil
}
