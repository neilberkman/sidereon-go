package sidereon

import "sidereon.dev/go/v3/internal/native"

// SP3OrbitClass selects a standard earth-fixed satellite-speed bound.
type SP3OrbitClass uint32

const (
	// SP3OrbitClassMEOGNSS covers GNSS medium-earth-orbit satellites.
	SP3OrbitClassMEOGNSS SP3OrbitClass = iota
	// SP3OrbitClassGeosynchronous covers geostationary and inclined-geosynchronous satellites.
	SP3OrbitClassGeosynchronous
	// SP3OrbitClassLEO covers low-earth-orbit satellites.
	SP3OrbitClassLEO
)

// SP3SpeedBoundKind selects the adjacent-pair speed check.
type SP3SpeedBoundKind uint32

const (
	// SP3SpeedBoundNone disables the adjacent-pair speed check.
	SP3SpeedBoundNone SP3SpeedBoundKind = iota
	// SP3SpeedBoundOrbitClass uses the selected orbit class's physical bound.
	SP3SpeedBoundOrbitClass
	// SP3SpeedBoundExplicit uses ExplicitMaxSpeedMPS.
	SP3SpeedBoundExplicit
)

// SP3ContinuityPolicy contains all native continuity-check options.
type SP3ContinuityPolicy struct {
	// SpeedBoundKind selects whether and how adjacent-pair speed is checked.
	SpeedBoundKind SP3SpeedBoundKind
	// OrbitClass selects the class bound when SpeedBoundKind is SP3SpeedBoundOrbitClass.
	OrbitClass SP3OrbitClass
	// ExplicitMaxSpeedMPS is the maximum speed when SpeedBoundKind is SP3SpeedBoundExplicit.
	ExplicitMaxSpeedMPS float64
	// ResidualToleranceEnabled controls whether hold-out residuals are checked.
	ResidualToleranceEnabled bool
	// ResidualToleranceM is the maximum held-out interpolation residual in metres.
	ResidualToleranceM float64
	// GapThresholdFactor controls when a node gap splits interpolation coverage.
	GapThresholdFactor float64
}

func nativeSP3ContinuityPolicy(value SP3ContinuityPolicy) native.Sp3ContinuityOptions {
	return native.Sp3ContinuityOptions{SpeedBoundKind: uint32(value.SpeedBoundKind), OrbitClass: uint32(value.OrbitClass), ExplicitMaxSpeedMPS: value.ExplicitMaxSpeedMPS, ResidualToleranceEnabled: value.ResidualToleranceEnabled, ResidualToleranceM: value.ResidualToleranceM, GapThresholdFactor: value.GapThresholdFactor}
}

func publicSP3ContinuityPolicy(value native.Sp3ContinuityOptions) SP3ContinuityPolicy {
	return SP3ContinuityPolicy{SpeedBoundKind: SP3SpeedBoundKind(value.SpeedBoundKind), OrbitClass: SP3OrbitClass(value.OrbitClass), ExplicitMaxSpeedMPS: value.ExplicitMaxSpeedMPS, ResidualToleranceEnabled: value.ResidualToleranceEnabled, ResidualToleranceM: value.ResidualToleranceM, GapThresholdFactor: value.GapThresholdFactor}
}

// DefaultSP3ContinuityPolicy returns native defaults for an orbit class.
func DefaultSP3ContinuityPolicy(orbitClass SP3OrbitClass) (SP3ContinuityPolicy, error) {
	value, err := native.DefaultSp3ContinuityOptions(uint32(orbitClass))
	if err != nil {
		return SP3ContinuityPolicy{}, publicError(err)
	}
	return publicSP3ContinuityPolicy(value), nil
}

// ContinuityReportJSON checks the product with the complete policy and returns its full report.
func (s *SP3) ContinuityReportJSON(policy SP3ContinuityPolicy) ([]byte, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.ContinuityReportJSON(nativeSP3ContinuityPolicy(policy))
	return append([]byte(nil), value...), publicError(err)
}

// ContinuityVerdictJSONWithPolicy returns the window verdict under the complete policy.
func (s *SP3) ContinuityVerdictJSONWithPolicy(policy SP3ContinuityPolicy, fromJ2000S, throughJ2000S float64) ([]byte, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.ContinuityVerdictJSONWithPolicy(nativeSP3ContinuityPolicy(policy), fromJ2000S, throughJ2000S)
	return append([]byte(nil), value...), publicError(err)
}
