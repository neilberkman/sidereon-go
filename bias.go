package sidereon

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"os"

	"github.com/neilberkman/sidereon-go/internal/native"
)

// BiasMode identifies whether a bias product is absolute, relative, or
// unspecified. It has no unit; values are the C ABI discriminants.
type BiasMode uint32

const (
	// BiasModeAbsolute reports product values in an absolute reference frame.
	BiasModeAbsolute BiasMode = BiasMode(native.BiasModeAbsoluteValue)
	// BiasModeRelative reports product values relative to the product reference.
	BiasModeRelative BiasMode = BiasMode(native.BiasModeRelativeValue)
	// BiasModeUnspecified reports that the product did not declare its mode.
	BiasModeUnspecified BiasMode = BiasMode(native.BiasModeUnspecifiedValue)
)

// BiasKind identifies an observable-bias relationship. Values are C ABI
// discriminants; the bias values themselves are seconds for code and cycles
// for phase records.
type BiasKind uint32

const (
	// BiasKindOSB is a one-observable signal bias.
	BiasKindOSB BiasKind = BiasKind(native.BiasKindOSBValue)
	// BiasKindDSB is a between-observable differential signal bias.
	BiasKindDSB BiasKind = BiasKind(native.BiasKindDSBValue)
	// BiasKindISB is an inter-system bias.
	BiasKindISB BiasKind = BiasKind(native.BiasKindISBValue)
)

// BiasTargetKind identifies the entity to which a bias record applies.
type BiasTargetKind uint32

const (
	// BiasTargetSystem applies to a whole GNSS system.
	BiasTargetSystem BiasTargetKind = BiasTargetKind(native.BiasTargetSystemValue)
	// BiasTargetSatellite applies to a satellite.
	BiasTargetSatellite BiasTargetKind = BiasTargetKind(native.BiasTargetSatelliteValue)
	// BiasTargetReceiver applies to a receiver station.
	BiasTargetReceiver BiasTargetKind = BiasTargetKind(native.BiasTargetReceiverValue)
	// BiasTargetSatelliteReceiver applies to a satellite and receiver pair.
	BiasTargetSatelliteReceiver BiasTargetKind = BiasTargetKind(native.BiasTargetSatelliteReceiverValue)
)

// BiasEpoch is a GNSS calendar epoch used by SINEX and code-DCB products.
type BiasEpoch struct {
	// Year is the proleptic calendar year.
	Year int32
	// DayOfYear is one-based, in the inclusive range 1..366.
	DayOfYear uint16
	// SecondOfDay is whole SI seconds since midnight.
	SecondOfDay uint32
}

// BiasRecord is a lossless copy of one C bias record. Presence fields
// distinguish absent values from present zero values.
type BiasRecord struct {
	// Kind is the OSB, DSB, or ISB relationship.
	Kind BiasKind
	// TargetKind is the record's system/satellite/receiver scope.
	TargetKind BiasTargetKind
	// System is the GNSS constellation.
	System GNSSSystem
	// HasSatelliteID reports whether SatelliteID is present.
	HasSatelliteID bool
	// SatelliteID is the canonical satellite token when present.
	SatelliteID string
	// Station is the receiver/station token when present.
	Station string
	// SVN is the space-vehicle number token when present.
	SVN string
	// Obs1 is the first RINEX observable token.
	Obs1 string
	// HasObs2 reports whether Obs2 is present.
	HasObs2 bool
	// Obs2 is the second RINEX observable token when present.
	Obs2 string
	// HasValidFrom reports whether ValidFrom is present.
	HasValidFrom bool
	// ValidFrom is the inclusive start epoch when present.
	ValidFrom BiasEpoch
	// HasValidUntil reports whether ValidUntil is present.
	HasValidUntil bool
	// ValidUntil is the inclusive end epoch when present.
	ValidUntil BiasEpoch
	// Value is seconds for code records and cycles for phase records.
	Value float64
	// HasSigma reports whether Sigma is present.
	HasSigma bool
	// Sigma is the standard deviation in the same unit as Value.
	Sigma float64
	// HasSlope reports whether Slope is present.
	HasSlope bool
	// Slope is the change in Value per second when supplied by the product.
	Slope float64
	// HasSlopeSigma reports whether SlopeSigma is present.
	HasSlopeSigma bool
	// SlopeSigma is the slope standard deviation in Value per second.
	SlopeSigma float64
	// IsPhase reports whether Value and Sigma are in carrier cycles.
	IsPhase bool
}

// CodeDCBOptions selects the observation pair and epoch policy for a code-DCB
// product. A nil options pointer selects the native parser defaults.
type CodeDCBOptions struct {
	// Obs1 is the first code observable token.
	Obs1 string
	// Obs2 is the second code observable token.
	Obs2 string
	// Year selects the DCB validity year.
	Year int32
	// Month selects the one-based validity month.
	Month uint8
	// TimeScale identifies the lookup epoch scale.
	TimeScale TimeScale
	// HasReceiverSystem reports whether ReceiverSystem filters the product.
	HasReceiverSystem bool
	// ReceiverSystem is the optional GNSS system filter.
	ReceiverSystem GNSSSystem
}

// BiasSet owns a parsed bias product. It must not be copied after first use.
// Read methods may run concurrently with Close: each read takes a shared
// native-handle lock, while Close waits for in-flight reads and then prevents
// later use. Close is idempotent and returns any native cleanup error.
type BiasSet struct {
	_      noCopy
	native *native.BiasSet
}

func newBiasSet(value *native.BiasSet, err error) (*BiasSet, error) {
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errNilNativeHandle
	}
	return &BiasSet{native: value}, nil
}

// ParseBiasSINEX strictly parses a SINEX bias byte stream. The byte slice is
// copied before entering C and is not retained.
func ParseBiasSINEX(data []byte) (*BiasSet, error) {
	return newBiasSet(native.ParseBiasSINEX(data, false))
}

// ParseBiasSINEXLossy parses a SINEX bias byte stream while retaining native
// skipped-record and warning diagnostics. Input bytes are copied.
func ParseBiasSINEXLossy(data []byte) (*BiasParsed, error) {
	set, err := newBiasSet(native.ParseBiasSINEX(data, true))
	if err != nil {
		return nil, err
	}
	return newBiasParsed(set)
}

// readBiasPath is the Go-owned filesystem adapter. Native C filesystem routes
// are intentionally not used: path reads, gzip transport, and the byte-parser
// ownership boundary remain in one place.
func readBiasPath(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if !bytes.HasPrefix(data, []byte{0x1f, 0x8b}) {
		return data, nil
	}
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	decoded, readErr := io.ReadAll(reader)
	closeErr := reader.Close()
	return decoded, errors.Join(readErr, closeErr)
}

// LoadBiasSINEX reads a plain or gzip-compressed SINEX file in Go and passes
// an owned byte copy to the C parser. The input path is never retained.
func LoadBiasSINEX(path string) (*BiasSet, error) {
	data, err := readBiasPath(path)
	if err != nil {
		return nil, err
	}
	return ParseBiasSINEX(data)
}

// LoadBiasSINEXLossy reads a plain or gzip-compressed SINEX file through the
// Go-owned path adapter and returns lossy parse diagnostics.
func LoadBiasSINEXLossy(path string) (*BiasParsed, error) {
	data, err := readBiasPath(path)
	if err != nil {
		return nil, err
	}
	return ParseBiasSINEXLossy(data)
}

// ParseCodeDCB strictly parses a code differential-code-bias byte stream.
// Options are copied into native-owned temporary storage.
func ParseCodeDCB(data []byte, options *CodeDCBOptions) (*BiasSet, error) {
	value, err := nativeCodeDCBOptions(options)
	if err != nil {
		return nil, err
	}
	return newBiasSet(native.ParseCodeDCB(data, value, false))
}

// ParseCodeDCBLossy parses a code DCB stream and retains native diagnostics.
func ParseCodeDCBLossy(data []byte, options *CodeDCBOptions) (*BiasParsed, error) {
	value, err := nativeCodeDCBOptions(options)
	if err != nil {
		return nil, err
	}
	set, err := newBiasSet(native.ParseCodeDCB(data, value, true))
	if err != nil {
		return nil, err
	}
	return newBiasParsed(set)
}

// LoadCodeDCB reads a code DCB file in Go and strictly parses its bytes.
func LoadCodeDCB(path string, options *CodeDCBOptions) (*BiasSet, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseCodeDCB(data, options)
}

// LoadCodeDCBLossy reads a code DCB file in Go and returns lossy diagnostics.
func LoadCodeDCBLossy(path string, options *CodeDCBOptions) (*BiasParsed, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseCodeDCBLossy(data, options)
}

func nativeCodeDCBOptions(value *CodeDCBOptions) (*native.CodeDCBOptions, error) {
	if value == nil {
		return nil, nil
	}
	return &native.CodeDCBOptions{Obs1: value.Obs1, Obs2: value.Obs2, Year: value.Year, Month: value.Month, TimeScale: uint32(value.TimeScale), HasReceiverSystem: value.HasReceiverSystem, ReceiverSystem: uint32(value.ReceiverSystem)}, nil
}

// BiasParsed contains a lossy parse result and its native diagnostics.
type BiasParsed struct {
	_ noCopy
	// Value is the owned parsed product; it is closed by Close.
	Value *BiasSet
	// SkipCount is the number of records omitted by lossy parsing.
	SkipCount int
	// WarningCount is the number of non-fatal native diagnostics.
	WarningCount int
}

func newBiasParsed(set *BiasSet) (*BiasParsed, error) {
	skipped, err := set.native.SkippedRecordCount()
	if err != nil {
		_ = set.Close()
		return nil, publicError(err)
	}
	warnings, err := set.native.WarningCount()
	if err != nil {
		_ = set.Close()
		return nil, publicError(err)
	}
	return &BiasParsed{Value: set, SkipCount: skipped, WarningCount: warnings}, nil
}

// Close releases Value. It is safe and idempotent, including on a nil result.
func (p *BiasParsed) Close() error {
	if p == nil || p.Value == nil {
		return nil
	}
	return p.Value.Close()
}

// Close releases the native bias product and is safe to repeat.
func (s *BiasSet) Close() error {
	if s == nil || s.native == nil {
		return nil
	}
	return publicError(s.native.Close())
}

// RecordCount returns the number of retained bias records.
func (s *BiasSet) RecordCount() (int, error) {
	if s == nil || s.native == nil {
		return 0, ErrClosed
	}
	v, err := s.native.RecordCount()
	return v, publicError(err)
}

// SkippedRecordCount returns lossy-parser records omitted from the result.
func (s *BiasSet) SkippedRecordCount() (int, error) {
	if s == nil || s.native == nil {
		return 0, ErrClosed
	}
	v, err := s.native.SkippedRecordCount()
	return v, publicError(err)
}

// WarningCount returns non-fatal diagnostics recorded by the native parser.
func (s *BiasSet) WarningCount() (int, error) {
	if s == nil || s.native == nil {
		return 0, ErrClosed
	}
	v, err := s.native.WarningCount()
	return v, publicError(err)
}

// Record returns an independent copy of the retained record at index.
func (s *BiasSet) Record(index int) (BiasRecord, error) {
	if s == nil || s.native == nil {
		return BiasRecord{}, ErrClosed
	}
	v, err := s.native.Record(index)
	return BiasRecord{Kind: BiasKind(v.Kind), TargetKind: BiasTargetKind(v.TargetKind), System: GNSSSystem(v.System), HasSatelliteID: v.HasSatelliteID, SatelliteID: v.SatelliteID, Station: v.Station, SVN: v.SVN, Obs1: v.Obs1, HasObs2: v.HasObs2, Obs2: v.Obs2, HasValidFrom: v.HasValidFrom, ValidFrom: BiasEpoch{Year: v.ValidFrom.Year, DayOfYear: v.ValidFrom.DayOfYear, SecondOfDay: v.ValidFrom.SecondOfDay}, HasValidUntil: v.HasValidUntil, ValidUntil: BiasEpoch{Year: v.ValidUntil.Year, DayOfYear: v.ValidUntil.DayOfYear, SecondOfDay: v.ValidUntil.SecondOfDay}, Value: v.Value, HasSigma: v.HasSigma, Sigma: v.Sigma, HasSlope: v.HasSlope, Slope: v.Slope, HasSlopeSigma: v.HasSlopeSigma, SlopeSigma: v.SlopeSigma, IsPhase: v.IsPhase}, publicError(err)
}

// Records returns independent copies of all retained records in native order.
func (s *BiasSet) Records() ([]BiasRecord, error) {
	count, err := s.RecordCount()
	if err != nil {
		return nil, err
	}
	out := make([]BiasRecord, count)
	for index := range out {
		out[index], err = s.Record(index)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

// Mode returns the product bias mode and declared time scale.
func (s *BiasSet) Mode() (BiasMode, TimeScale, error) {
	if s == nil || s.native == nil {
		return 0, 0, ErrClosed
	}
	mode, scale, err := s.native.Mode()
	return BiasMode(mode), TimeScale(scale), publicError(err)
}

// TimeScale returns the product's declared time scale.
func (s *BiasSet) TimeScale() (TimeScale, error) {
	_, scale, err := s.Mode()
	return scale, err
}

// CodeOSBSeconds looks up a code OSB in seconds. The bool reports presence;
// a present zero is distinct from an absent value.
func (s *BiasSet) CodeOSBSeconds(satellite, observation string, epoch BiasEpoch) (float64, bool, error) {
	if s == nil || s.native == nil {
		return 0, false, ErrClosed
	}
	v, present, err := s.native.CodeOSBSeconds(satellite, observation, native.BiasEpoch{Year: epoch.Year, DayOfYear: epoch.DayOfYear, SecondOfDay: epoch.SecondOfDay})
	return v, present, publicError(err)
}

// PhaseOSBCycles looks up a phase OSB in carrier cycles with an explicit
// presence flag.
func (s *BiasSet) PhaseOSBCycles(satellite, observation string, epoch BiasEpoch) (float64, bool, error) {
	if s == nil || s.native == nil {
		return 0, false, ErrClosed
	}
	v, present, err := s.native.PhaseOSBCycles(satellite, observation, native.BiasEpoch{Year: epoch.Year, DayOfYear: epoch.DayOfYear, SecondOfDay: epoch.SecondOfDay})
	return v, present, publicError(err)
}

// CodeDSBSeconds looks up a code DSB in seconds with an explicit presence
// flag. Returned values are independent scalars.
func (s *BiasSet) CodeDSBSeconds(satellite, observation1, observation2 string, epoch BiasEpoch) (float64, bool, error) {
	if s == nil || s.native == nil {
		return 0, false, ErrClosed
	}
	v, present, err := s.native.CodeDSBSeconds(satellite, observation1, observation2, native.BiasEpoch{Year: epoch.Year, DayOfYear: epoch.DayOfYear, SecondOfDay: epoch.SecondOfDay})
	return v, present, publicError(err)
}
