package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// SiderealOrbitRepeatLag returns the period of the selected broadcast orbit
// in seconds near the supplied TDB/J2000 epoch.
func SiderealOrbitRepeatLag(b *BroadcastEphemeris, satellite string, nearEpochJ2000S float64) (float64, error) {
	if b == nil || b.handle == nil {
		return 0, ErrClosed
	}
	value, err := native.SiderealOrbitRepeatLag(b.handle, satellite, nearEpochJ2000S)
	return value, publicError(err)
}
