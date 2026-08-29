package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// NMEASummary contains aggregate counts copied from a parsed NMEA log.
type NMEASummary struct {
	SentenceCount uint64
	EpochCount    uint64
	SkipCount     uint64
	WarningCount  uint64
}

// NMEAEpoch is one copied epoch summary from a parsed NMEA log.
type NMEAEpoch struct {
	HasPosition        bool
	LatitudeRad        float64
	LongitudeRad       float64
	HeightM            float64
	SentenceCount      uint64
	UsedSatelliteCount uint64
	SatellitesInView   uint64
	SkipCount          uint64
	WarningCount       uint64
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
	out := make([]NMEAEpoch, len(epochs))
	for i, epoch := range epochs {
		out[i] = NMEAEpoch{
			HasPosition:        epoch.HasPosition,
			LatitudeRad:        epoch.LatitudeRad,
			LongitudeRad:       epoch.LongitudeRad,
			HeightM:            epoch.HeightM,
			SentenceCount:      uint64(epoch.SentenceCount),
			UsedSatelliteCount: uint64(epoch.UsedSatelliteCount),
			SatellitesInView:   uint64(epoch.SatellitesInView),
			SkipCount:          uint64(epoch.SkipCount),
			WarningCount:       uint64(epoch.WarningCount),
		}
	}
	return out, nil
}
