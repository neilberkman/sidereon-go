package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// NMEASummary contains aggregate counts detached from a parsed NMEA log.
type NMEASummary struct {
	// SentenceCount is the number of NMEA sentences parsed.
	SentenceCount uint64
	// EpochCount is the number of epoch summaries produced.
	EpochCount uint64
	// SkipCount is the number of malformed or skipped sentences.
	SkipCount uint64
	// WarningCount is the number of parser warnings.
	WarningCount uint64
}

// NMEAEpoch is one detached epoch summary from a parsed NMEA log.
type NMEAEpoch struct {
	// HasCalendarEpoch reports whether CalendarEpoch contains a parsed UTC epoch.
	HasCalendarEpoch bool
	// CalendarEpoch is the calendar epoch.
	CalendarEpoch CalendarEpoch
	// HasPosition reports whether Position contains a parsed latitude/longitude fix.
	HasPosition bool
	// LatitudeRad is the latitude rad in radians.
	LatitudeRad float64
	// LongitudeRad is the longitude rad in radians.
	LongitudeRad float64
	// HeightM is the height m in metres.
	HeightM float64
	// HasInstantJ2000S reports whether InstantJ2000S contains a parsed epoch.
	HasInstantJ2000S bool
	// InstantJ2000S is the parsed epoch in seconds from J2000.
	InstantJ2000S float64
	// HasPDOP reports whether PDOP is populated from a dilution report.
	HasPDOP bool
	// PDOP is the dimensionless position dilution of precision when HasPDOP is true.
	PDOP float64
	// HasHDOP reports whether HDOP is populated from a dilution report.
	HasHDOP bool
	// HDOP is the dimensionless horizontal dilution of precision when HasHDOP is true.
	HDOP float64
	// HasVDOP reports whether VDOP is populated from a dilution report.
	HasVDOP bool
	// VDOP is the dimensionless vertical dilution of precision when HasVDOP is true.
	VDOP float64
	// SentenceCount is the number of sentences contributing to this epoch.
	SentenceCount uint64
	// UsedSatelliteCount is the number of satellites used in this epoch.
	UsedSatelliteCount uint64
	// SatellitesInView is the number of satellites reported in view.
	SatellitesInView uint64
	// SkipCount is the number of skipped sentences in this epoch.
	SkipCount uint64
	// WarningCount is the number of warnings associated with this epoch.
	WarningCount uint64
	// HasGGA reports whether a GGA sentence was parsed for the epoch.
	HasGGA bool
	// HasRMC reports whether an RMC sentence was parsed for the epoch.
	HasRMC bool
	// HasGLL reports whether a GLL sentence was parsed for the epoch.
	HasGLL bool
	// GSACount is the number of GSA sentences in this epoch.
	GSACount uint64
	// GSVGroupCount is the number of GSV groups in this epoch.
	GSVGroupCount uint64
}

// NMEAChunkSummary contains the output counts from one incremental Push or
// Finish call. RetainedLength is the number of partial-line bytes held for a
// later chunk.
type NMEAChunkSummary struct {
	// SentenceCount is the number of sentences consumed by the call.
	SentenceCount uint64
	// CompletedEpochCount is the number of complete epochs emitted by the call.
	CompletedEpochCount uint64
	// SkipCount is the number of skipped sentences reported by the call.
	SkipCount uint64
	// WarningCount is the number of parser warnings reported by the call.
	WarningCount uint64
	// RetainedLength is the number of bytes held from an incomplete trailing line.
	RetainedLength uint64
}

// NMEAGGAOptions contains validated fields for one NMEA 0183 GGA sentence.
// Position is WGS84 geodetic radians/metres and UTCSecondsOfDay is rounded
// down to centisecond precision by the engine.
type NMEAGGAOptions struct {
	// Talker is the NMEA talker identifier.
	Talker string
	// UTCSecondsOfDay is the utc seconds of day in seconds since midnight.
	UTCSecondsOfDay float64
	// Position is the position value in the containing frame.
	Position Geodetic
	// Quality is the NMEA GGA fix-quality code.
	Quality uint32
	// SatellitesUsed is the number of satellites used in the GGA solution.
	SatellitesUsed uint8
	// HDOP is the dimensionless horizontal dilution of precision in the GGA sentence.
	HDOP float64
	// CoordinateDecimals is the number of decimal places emitted for coordinates.
	CoordinateDecimals uint8
}

// DefaultNMEAGGAOptions returns the sibling-interface VRS defaults. The
// caller supplies Position and UTCSecondsOfDay before calling WriteNMEAGGA.
func DefaultNMEAGGAOptions() NMEAGGAOptions {
	return NMEAGGAOptions{Talker: "GP", Quality: 1, SatellitesUsed: 10, HDOP: 1, CoordinateDecimals: 7}
}

// NMEAAccumulator owns an incremental native parser. Mutating Push and Finish
// calls are serialized with read-only Summary, Epochs, RetainedLength, and
// Close calls, so the handle may be shared by goroutines.
type NMEAAccumulator struct {
	_      noCopy
	handle *native.NMEAAccumulator
}

// NewNMEAAccumulator creates an empty incremental NMEA parser.
func NewNMEAAccumulator() (*NMEAAccumulator, error) {
	handle, err := native.NewNMEAAccumulator()
	if err != nil {
		return nil, publicError(err)
	}
	if handle == nil {
		return nil, errNilNativeHandle
	}
	return &NMEAAccumulator{handle: handle}, nil
}

// Close releases the native accumulator. It is idempotent.
func (accumulator *NMEAAccumulator) Close() error {
	if accumulator == nil || accumulator.handle == nil {
		return nil
	}
	return publicError(accumulator.handle.Close())
}

// Push parses one arbitrary byte chunk and retains an incomplete trailing
// line for a later call. The input is copied into C-owned memory for the call.
func (accumulator *NMEAAccumulator) Push(data []byte) (NMEAChunkSummary, error) {
	if accumulator == nil || accumulator.handle == nil {
		return NMEAChunkSummary{}, ErrClosed
	}
	value, err := accumulator.handle.Push(data)
	return publicNMEAChunkSummary(value), publicError(err)
}

// Finish flushes the pending epoch after the final byte chunk. A
// CompletedEpochCount of zero means there was no pending epoch.
func (accumulator *NMEAAccumulator) Finish() (NMEAChunkSummary, error) {
	if accumulator == nil || accumulator.handle == nil {
		return NMEAChunkSummary{}, ErrClosed
	}
	value, err := accumulator.handle.Finish()
	return publicNMEAChunkSummary(value), publicError(err)
}

// Summary returns cumulative counts for all chunks accepted so far.
func (accumulator *NMEAAccumulator) Summary() (NMEASummary, error) {
	if accumulator == nil || accumulator.handle == nil {
		return NMEASummary{}, ErrClosed
	}
	value, err := accumulator.handle.Summary()
	return NMEASummary{
		SentenceCount: value.SentenceCount,
		EpochCount:    value.EpochCount,
		SkipCount:     value.SkipCount,
		WarningCount:  value.WarningCount,
	}, publicError(err)
}

// RetainedLength returns the number of partial-line bytes waiting for a later
// Push or Finish call.
func (accumulator *NMEAAccumulator) RetainedLength() (uint64, error) {
	if accumulator == nil || accumulator.handle == nil {
		return 0, ErrClosed
	}
	value, err := accumulator.handle.RetainedLength()
	return value, publicError(err)
}

// Epochs copies every completed accumulated epoch into Go-owned values.
func (accumulator *NMEAAccumulator) Epochs() ([]NMEAEpoch, error) {
	if accumulator == nil || accumulator.handle == nil {
		return nil, ErrClosed
	}
	values, err := accumulator.handle.Epochs()
	if err != nil {
		return nil, publicError(err)
	}
	return publicNMEAEpochs(values), nil
}

// WriteNMEAGGA formats one checksummed NMEA 0183 GGA sentence, including its
// CRLF terminator, through the engine's validated writer.
func WriteNMEAGGA(options NMEAGGAOptions) ([]byte, error) {
	value, err := native.WriteNMEAGGA(native.NMEAGGAOptions{
		Talker:             options.Talker,
		UTCSecondsOfDay:    options.UTCSecondsOfDay,
		Position:           native.Geodetic{LatitudeRad: options.Position.LatitudeRad, LongitudeRad: options.Position.LongitudeRad, HeightM: options.Position.HeightM},
		Quality:            options.Quality,
		SatellitesUsed:     options.SatellitesUsed,
		HDOP:               options.HDOP,
		CoordinateDecimals: options.CoordinateDecimals,
	})
	return value, publicError(err)
}

func publicNMEAChunkSummary(value native.NMEAChunkSummary) NMEAChunkSummary {
	return NMEAChunkSummary{
		SentenceCount:       value.SentenceCount,
		CompletedEpochCount: value.CompletedEpochCount,
		SkipCount:           value.SkipCount,
		WarningCount:        value.WarningCount,
		RetainedLength:      value.RetainedLength,
	}
}

// NMEALog owns a parsed C NMEA log. A log may be shared by goroutines for
// read-only Summary and Epochs calls. Close may run concurrently with reads;
// it waits for an active read, clears the native pointer, and is idempotent.
type NMEALog struct {
	_      noCopy
	handle *native.NMEALog
}

// ParseNMEA parses data through the C ABI and returns an owning log handle.
// The input is copied only for the duration of the C call; the C library does
// not retain a Go pointer.
func ParseNMEA(data []byte) (*NMEALog, error) {
	handle, err := native.ParseNMEA(data)
	if err != nil {
		return nil, publicError(err)
	}
	if handle == nil {
		return nil, errNilNativeHandle
	}
	return &NMEALog{handle: handle}, nil
}

// Close releases the native log. It is safe to call more than once and from a
// goroutine concurrent with read-only methods.
func (l *NMEALog) Close() error {
	if l == nil || l.handle == nil {
		return nil
	}
	err := l.handle.Close()
	return publicError(err)
}

// Summary copies aggregate log counts into Go-owned values.
func (l *NMEALog) Summary() (NMEASummary, error) {
	if l == nil || l.handle == nil {
		return NMEASummary{}, ErrClosed
	}
	summary, err := l.handle.Summary()
	return NMEASummary{
		SentenceCount: uint64(summary.SentenceCount),
		EpochCount:    uint64(summary.EpochCount),
		SkipCount:     uint64(summary.SkipCount),
		WarningCount:  uint64(summary.WarningCount),
	}, publicError(err)
}

// Epochs performs the C size query and copy, returning Go-owned summaries.
func (l *NMEALog) Epochs() ([]NMEAEpoch, error) {
	if l == nil || l.handle == nil {
		return nil, ErrClosed
	}
	epochs, err := l.handle.Epochs()
	if err != nil {
		return nil, publicError(err)
	}
	return publicNMEAEpochs(epochs), nil
}

func publicNMEAEpochs(values []native.NMEAEpoch) []NMEAEpoch {
	out := make([]NMEAEpoch, len(values))
	for i, epoch := range values {
		out[i] = NMEAEpoch{
			HasCalendarEpoch:   epoch.HasCalendarEpoch,
			CalendarEpoch:      calendarEpoch(epoch.CalendarEpoch),
			HasPosition:        epoch.HasPosition,
			LatitudeRad:        epoch.LatitudeRad,
			LongitudeRad:       epoch.LongitudeRad,
			HeightM:            epoch.HeightM,
			HasInstantJ2000S:   epoch.HasInstantJ2000S,
			InstantJ2000S:      epoch.InstantJ2000S,
			HasPDOP:            epoch.HasPDOP,
			PDOP:               epoch.PDOP,
			HasHDOP:            epoch.HasHDOP,
			HDOP:               epoch.HDOP,
			HasVDOP:            epoch.HasVDOP,
			VDOP:               epoch.VDOP,
			SentenceCount:      epoch.SentenceCount,
			UsedSatelliteCount: epoch.UsedSatelliteCount,
			SatellitesInView:   epoch.SatellitesInView,
			SkipCount:          epoch.SkipCount,
			WarningCount:       epoch.WarningCount,
			HasGGA:             epoch.HasGGA,
			HasRMC:             epoch.HasRMC,
			HasGLL:             epoch.HasGLL,
			GSACount:           epoch.GSACount,
			GSVGroupCount:      epoch.GSVGroupCount,
		}
	}
	return out
}
