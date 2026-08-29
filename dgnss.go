package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// DGNSSObservation is one code-only pseudorange observation in meters.
type DGNSSObservation struct {
	// SatelliteID identifies or counts this record.
	SatelliteID string
	// PseudorangeM is the pseudorange m in metres.
	PseudorangeM float64
}

// DGNSSAppliedObservation is one observation after applying a base correction.
type DGNSSAppliedObservation struct {
	// SatelliteID identifies or counts this record.
	SatelliteID string
	// PseudorangeM is the pseudorange m in metres.
	PseudorangeM float64
}

// DGNSSCorrections owns a native per-satellite correction table. It must not
// be copied after first use.
type DGNSSCorrections struct {
	_      noCopy
	handle *native.DgnssCorrections
}

// DGNSSApplied owns the native result of applying DGNSS corrections. It must
// not be copied after first use.
type DGNSSApplied struct {
	_      noCopy
	handle *native.DgnssApplied
}

// DGNSSSolution owns a native corrected-rover position solution. It must not
// be copied after first use.
type DGNSSSolution struct {
	_      noCopy
	handle *native.DgnssSolution
}

func nativeDGNSSObservations(values []DGNSSObservation) []native.DgnssCodeObservation {
	result := make([]native.DgnssCodeObservation, len(values))
	for i, value := range values {
		result[i] = native.DgnssCodeObservation{SatelliteID: value.SatelliteID, PseudorangeM: value.PseudorangeM}
	}
	return result
}

// DGNSSPseudorangeCorrections computes base-station pseudorange corrections
// through the native C ABI.
func DGNSSPseudorangeCorrections(sp3 *SP3, basePosition [3]float64, observations []DGNSSObservation, receiveTimeJ2000S float64) (*DGNSSCorrections, error) {
	if sp3 == nil || sp3.handle == nil {
		return nil, ErrClosed
	}
	handle, err := native.DgnssPseudorangeCorrections(sp3.handle, basePosition, nativeDGNSSObservations(observations), receiveTimeJ2000S)
	if err != nil {
		return nil, publicError(err)
	}
	return &DGNSSCorrections{handle: handle}, nil
}

// Close releases the correction table. It is idempotent and safe to call
// concurrently with its read methods.
func (c *DGNSSCorrections) Close() error {
	if c == nil || c.handle == nil {
		return nil
	}
	return publicError(c.handle.Close())
}

// Count returns the number of correction entries.
func (c *DGNSSCorrections) Count() (int, error) {
	if c == nil || c.handle == nil {
		return 0, ErrClosed
	}
	count, err := c.handle.Count()
	return count, publicError(err)
}

// Correction returns a satellite's correction and whether an entry exists.
func (c *DGNSSCorrections) Correction(satelliteID string) (float64, bool, error) {
	if c == nil || c.handle == nil {
		return 0, false, ErrClosed
	}
	value, present, err := c.handle.Correction(satelliteID)
	return value, present, publicError(err)
}

// ApplyDGNSSCorrections applies a correction table to rover observations.
func ApplyDGNSSCorrections(observations []DGNSSObservation, corrections *DGNSSCorrections) (*DGNSSApplied, error) {
	if corrections == nil || corrections.handle == nil {
		return nil, ErrClosed
	}
	handle, err := native.ApplyDgnssCorrections(nativeDGNSSObservations(observations), corrections.handle)
	if err != nil {
		return nil, publicError(err)
	}
	return &DGNSSApplied{handle: handle}, nil
}

// Close releases the applied-observations result. It is idempotent and safe
// to call concurrently with its read methods.
func (a *DGNSSApplied) Close() error {
	if a == nil || a.handle == nil {
		return nil
	}
	return publicError(a.handle.Close())
}

// Counts returns the numbers of corrected and dropped observations.
func (a *DGNSSApplied) Counts() (corrected, dropped int, err error) {
	if a == nil || a.handle == nil {
		return 0, 0, ErrClosed
	}
	corrected, dropped, err = a.handle.Counts()
	return corrected, dropped, publicError(err)
}

// Corrected returns one copied corrected observation by index.
func (a *DGNSSApplied) Corrected(index int) (DGNSSAppliedObservation, error) {
	if a == nil || a.handle == nil {
		return DGNSSAppliedObservation{}, ErrClosed
	}
	value, err := a.handle.Corrected(index)
	return DGNSSAppliedObservation{SatelliteID: value.SatelliteID, PseudorangeM: value.PseudorangeM}, publicError(err)
}

// Dropped returns one copied dropped satellite identifier by index.
func (a *DGNSSApplied) Dropped(index int) (string, error) {
	if a == nil || a.handle == nil {
		return "", ErrClosed
	}
	value, err := a.handle.Dropped(index)
	return value, publicError(err)
}

// SolveDGNSSPosition computes corrections, applies them to rover
// observations, and solves the corrected rover position through the C ABI.
func SolveDGNSSPosition(sp3 *SP3, basePosition [3]float64, base, rover []DGNSSObservation, config SPPConfig) (*DGNSSSolution, error) {
	if sp3 == nil || sp3.handle == nil {
		return nil, ErrClosed
	}
	handle, err := native.SolveDgnssPosition(sp3.handle, basePosition, nativeDGNSSObservations(base), nativeDGNSSObservations(rover), nativeSPPConfig(config))
	if err != nil {
		return nil, publicError(err)
	}
	return &DGNSSSolution{handle: handle}, nil
}

// Close releases the DGNSS solution. It is idempotent and safe to call
// concurrently with its read methods.
func (s *DGNSSSolution) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return publicError(s.handle.Close())
}

// Baseline returns the rover-minus-base ECEF baseline and its length.
func (s *DGNSSSolution) Baseline() ([3]float64, float64, error) {
	if s == nil || s.handle == nil {
		return [3]float64{}, 0, ErrClosed
	}
	vector, length, err := s.handle.Baseline()
	return vector, length, publicError(err)
}

// DroppedSatellites returns copied rover satellite identifiers that lacked a
// matching base correction.
func (s *DGNSSSolution) DroppedSatellites() ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.DroppedSatellites()
	return append([]string(nil), value...), publicError(err)
}

// Solution returns the embedded corrected-rover SPP solution as detached Go
// memory.
func (s *DGNSSSolution) Solution() (SPPSolution, error) {
	if s == nil || s.handle == nil {
		return SPPSolution{}, ErrClosed
	}
	value, err := s.handle.Solution()
	return publicSPPSolution(value), publicError(err)
}

// SPPSolution is an explicit alias for Solution for callers that prefer the
// result type in the method name.
func (s *DGNSSSolution) SPPSolution() (SPPSolution, error) {
	return s.Solution()
}
