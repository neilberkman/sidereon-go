package sidereon

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"os"

	"sidereon.dev/go/v3/internal/native"
)

// BiasReadPolicy selects the validation policy for bias product titles and records.
type BiasReadPolicy uint32

const (
	BiasReadPolicyStrict  BiasReadPolicy = BiasReadPolicy(native.BiasReadPolicyStrictValue)
	BiasReadPolicyLenient BiasReadPolicy = BiasReadPolicy(native.BiasReadPolicyLenientValue)
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

// BiasLookupStatus is the exact outcome of one bias lookup.
type BiasLookupStatus uint32

const (
	BiasLookupAvailable        BiasLookupStatus = BiasLookupStatus(native.BiasLookupAvailableValue)
	BiasLookupAbsent           BiasLookupStatus = BiasLookupStatus(native.BiasLookupAbsentValue)
	BiasLookupUnsupportedScale BiasLookupStatus = BiasLookupStatus(native.BiasLookupUnsupportedScaleValue)
	BiasLookupAmbiguous        BiasLookupStatus = BiasLookupStatus(native.BiasLookupAmbiguousValue)
	BiasLookupCarrierRequired  BiasLookupStatus = BiasLookupStatus(native.BiasLookupCarrierRequiredValue)
	BiasLookupInvalidCarrier   BiasLookupStatus = BiasLookupStatus(native.BiasLookupInvalidCarrierValue)
	BiasLookupCarrierUnknown   BiasLookupStatus = BiasLookupStatus(native.BiasLookupCarrierUnknownValue)
	BiasLookupUndefinedSlope   BiasLookupStatus = BiasLookupStatus(native.BiasLookupUndefinedSlopeValue)
	BiasLookupInvalidEpoch     BiasLookupStatus = BiasLookupStatus(native.BiasLookupInvalidEpochValue)
	BiasLookupUnknown          BiasLookupStatus = BiasLookupStatus(native.BiasLookupUnknownValue)
)

// BiasLookup retains every field returned by the native OSB/DSB query.
type BiasLookup struct {
	Status                  BiasLookupStatus
	Value                   float64
	RecordIndices           []uint64
	OverriddenRecordIndices []uint64
	RecordIndex             uint64
	HasProductTimeScale     bool
	ProductTimeScale        TimeScale
	HasQueryTimeScale       bool
	QueryTimeScale          TimeScale
	Observable              string
	UnknownVariant          string
}

// BiasSetModeInfo retains whether the source product actually declared a time scale.
type BiasSetModeInfo struct {
	Mode         BiasMode
	HasTimeScale bool
	TimeScale    TimeScale
}

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

type BiasNoticeKind uint32

const (
	BiasNoticeDeparture BiasNoticeKind = iota
	BiasNoticeInvalidUTF8
	BiasNoticeRepeatedDeclaration
	BiasNoticeConflictingDeclaration
	BiasNoticeOverlap
	BiasNoticeDcbTimeSystemAssumed
	BiasNoticeDcbTimeSystemAlias
	BiasNoticeUnknown BiasNoticeKind = 999
)

type BiasDepartureKind uint32

const (
	BiasDepartureNone BiasDepartureKind = iota
	BiasDepartureHeaderLayout
	BiasDepartureOtherVersion
	BiasDepartureMissingFooter
	BiasDepartureContentAfterFooter
	BiasDepartureUnexpectedControlLine
	BiasDepartureUnclosedBlock
	BiasDepartureUnopenedBlockEnd
	BiasDepartureMismatchedBlockEnd
	BiasDepartureNestedBlock
	BiasDepartureMissingBlock
	BiasDepartureUnknownBlock
	BiasDepartureBlockStartSuffix
	BiasDepartureDataOutsideBlock
	BiasDepartureMissingDeclaration
	BiasDepartureUnsupportedBiasMode
	BiasDepartureNonStandardTimeSystem
	BiasDepartureHeaderModeMismatch
	BiasDepartureUnknownDcbTimeSystem
	BiasDepartureEstimateCountMismatch
	BiasDepartureUnknown BiasDepartureKind = 999
)

type BiasNoticeText uint32

const (
	BiasNoticeTextKeyword BiasNoticeText = iota
	BiasNoticeTextLabel
	BiasNoticeTextName
	BiasNoticeTextOpen
	BiasNoticeTextClose
	BiasNoticeTextInner
	BiasNoticeTextVersion
	BiasNoticeTextReason
	BiasNoticeTextHeader
)

type BiasNotice struct {
	Kind           BiasNoticeKind
	Departure      BiasDepartureKind
	HasLine        bool
	Line           int
	First          int
	Second         int
	DeclaredCount  uint64
	SolutionRows   int
	BiasMode       BiasMode
	UnknownVariant string
}

// BiasErrorKind identifies a typed Bias-SINEX or CODE DCB failure.
type BiasErrorKind uint32

const (
	BiasErrorNone BiasErrorKind = iota
	BiasErrorInvalidInput
	BiasErrorInvalidEpoch
	BiasErrorUnknownObservable
	BiasErrorUnsupportedVersion
	BiasErrorMissingDcbMetadata
	BiasErrorMissingClockReference
	BiasErrorMissingWriterMetadata
	BiasErrorUTF8
	BiasErrorDeparture
	BiasErrorInvalidUTF8Line
	BiasErrorUnsupportedTimeSystem
	BiasErrorDcbRecordMismatch
	BiasErrorUnknown BiasErrorKind = 999
)

// BiasErrorText selects a text field copied from a typed bias failure.
type BiasErrorText uint32

const (
	BiasErrorTextMessage BiasErrorText = iota
	BiasErrorTextField
	BiasErrorTextReason
	BiasErrorTextCode
	BiasErrorTextVersion
	BiasErrorTextDepartureNotice
)

// BiasError preserves a native bias failure's structured fields and exact text bytes.
type BiasError struct {
	Kind          BiasErrorKind
	Line          int
	Record        int
	HasTimeScale  bool
	TimeScale     TimeScale
	Departure     BiasNotice
	Message       []byte
	Field         []byte
	Reason        []byte
	Code          []byte
	Version       []byte
	DepartureText [9][]byte
}

func (err *BiasError) Error() string {
	if err == nil {
		return "sidereon: bias error"
	}
	if len(err.Message) != 0 {
		return string(err.Message)
	}
	return "sidereon: bias error"
}

// TextBytes returns an independent copy of one error text field.
func (err *BiasError) TextBytes(part BiasErrorText, departurePart BiasNoticeText) ([]byte, error) {
	if err == nil {
		return nil, ErrClosed
	}
	var value []byte
	switch part {
	case BiasErrorTextMessage:
		value = err.Message
	case BiasErrorTextField:
		value = err.Field
	case BiasErrorTextReason:
		value = err.Reason
	case BiasErrorTextCode:
		value = err.Code
	case BiasErrorTextVersion:
		value = err.Version
	case BiasErrorTextDepartureNotice:
		if departurePart > BiasNoticeTextHeader {
			return nil, invalidArgument("invalid bias departure text part")
		}
		value = err.DepartureText[departurePart]
	default:
		return nil, invalidArgument("invalid bias error text part")
	}
	return append([]byte(nil), value...), nil
}

// Text returns one error text field as a byte-preserving Go string.
func (err *BiasError) Text(part BiasErrorText, departurePart BiasNoticeText) (string, error) {
	value, errText := err.TextBytes(part, departurePart)
	return string(value), errText
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

// ParseBiasSINEXWithPolicy parses a SINEX bias stream using an explicit native read policy.
// The strict policy is equivalent to ParseBiasSINEX; lenient departures are returned in BiasParsed.
func ParseBiasSINEXWithPolicy(data []byte, policy BiasReadPolicy) (*BiasParsed, error) {
	if policy != BiasReadPolicyStrict && policy != BiasReadPolicyLenient {
		return nil, errors.New("sidereon: invalid bias read policy")
	}
	set, err := newBiasSet(native.ParseBiasSINEXWithPolicy(data, uint32(policy)))
	if err != nil {
		return nil, err
	}
	return newBiasParsed(set)
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

// LoadBiasSINEXWithPolicy reads a plain or gzip-compressed file using an explicit read policy.
func LoadBiasSINEXWithPolicy(path string, policy BiasReadPolicy) (*BiasParsed, error) {
	data, err := readBiasPath(path)
	if err != nil {
		return nil, err
	}
	return ParseBiasSINEXWithPolicy(data, policy)
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

// ParseCodeDCBWithPolicy parses a code DCB byte stream under an explicit native read policy.
func ParseCodeDCBWithPolicy(data []byte, options *CodeDCBOptions, policy BiasReadPolicy) (*BiasParsed, error) {
	if policy != BiasReadPolicyStrict && policy != BiasReadPolicyLenient {
		return nil, errors.New("sidereon: invalid bias read policy")
	}
	value, err := nativeCodeDCBOptions(options)
	if err != nil {
		return nil, err
	}
	set, err := newBiasSet(native.ParseCodeDCBWithPolicy(data, value, uint32(policy)))
	if err != nil {
		return nil, err
	}
	return newBiasParsed(set)
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

// LoadCodeDCBWithPolicy reads and parses a code DCB file under an explicit read policy.
func LoadCodeDCBWithPolicy(path string, options *CodeDCBOptions, policy BiasReadPolicy) (*BiasParsed, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseCodeDCBWithPolicy(data, options, policy)
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

// BiasParsed contains a parsed bias product and its native diagnostics.
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

// NoticeCount returns the number of structured lenient-parse notices.
func (s *BiasSet) NoticeCount() (int, error) {
	if s == nil || s.native == nil {
		return 0, ErrClosed
	}
	value, err := s.native.NoticeCount()
	return value, publicError(err)
}

// Notice returns structured notice fields, including typed Bias-SINEX departure reasons.
func (s *BiasSet) Notice(index int) (BiasNotice, error) {
	if s == nil || s.native == nil {
		return BiasNotice{}, ErrClosed
	}
	value, err := s.native.Notice(index)
	if err != nil {
		return BiasNotice{}, publicError(err)
	}
	return BiasNotice{Kind: BiasNoticeKind(value.Kind), Departure: BiasDepartureKind(value.Departure), HasLine: value.HasLine, Line: value.Line, First: value.First, Second: value.Second, DeclaredCount: value.DeclaredCount, SolutionRows: value.SolutionRows, BiasMode: BiasMode(value.BiasMode), UnknownVariant: value.UnknownVariant}, nil
}

// NoticeText copies one structured notice text component without parsing debug output.
func (s *BiasSet) NoticeText(index int, part BiasNoticeText) (string, error) {
	if s == nil || s.native == nil {
		return "", ErrClosed
	}
	value, err := s.native.NoticeText(index, uint32(part))
	return value, publicError(err)
}

// BiasSINEXText writes the set as Bias-SINEX UTF-8 text.
func (s *BiasSet) BiasSINEXText() (string, error) {
	if s == nil || s.native == nil {
		return "", ErrClosed
	}
	value, err := s.native.BiasSINEXText()
	return value, publicError(err)
}

// BiasSINEXBytes writes Bias-SINEX bytes, preserving retained non-UTF-8 source lines.
func (s *BiasSet) BiasSINEXBytes() ([]byte, error) {
	if s == nil || s.native == nil {
		return nil, ErrClosed
	}
	value, err := s.native.BiasSINEXBytes()
	return value, publicError(err)
}

// CodeDCBText writes the set as CODE DCB UTF-8 text.
func (s *BiasSet) CodeDCBText() (string, error) {
	if s == nil || s.native == nil {
		return "", ErrClosed
	}
	value, err := s.native.CodeDCBText()
	return value, publicError(err)
}

// CodeDCBBytes writes CODE DCB bytes, preserving retained non-UTF-8 source lines.
func (s *BiasSet) CodeDCBBytes() ([]byte, error) {
	if s == nil || s.native == nil {
		return nil, ErrClosed
	}
	value, err := s.native.CodeDCBBytes()
	return value, publicError(err)
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

// ModeInfo reports product mode and the exact presence of its time scale.
func (s *BiasSet) ModeInfo() (BiasSetModeInfo, error) {
	if s == nil || s.native == nil {
		return BiasSetModeInfo{}, ErrClosed
	}
	value, err := s.native.ModeInfo()
	return BiasSetModeInfo{Mode: BiasMode(value.Mode), HasTimeScale: value.HasTimeScale, TimeScale: TimeScale(value.TimeScale)}, publicError(err)
}

func publicBiasLookup(value native.NativeBiasLookup) BiasLookup {
	return BiasLookup{
		Status: BiasLookupStatus(value.Status), Value: value.Value,
		RecordIndices: value.RecordIndices, OverriddenRecordIndices: value.OverriddenRecordIndices,
		RecordIndex: value.RecordIndex, HasProductTimeScale: value.HasProductTimeScale,
		ProductTimeScale: TimeScale(value.ProductTimeScale), HasQueryTimeScale: value.HasQueryTimeScale,
		QueryTimeScale: TimeScale(value.QueryTimeScale), Observable: value.Observable, UnknownVariant: value.UnknownVariant,
	}
}

// CodeOSBLookup returns the complete outcome and record provenance for a code OSB query.
func (s *BiasSet) CodeOSBLookup(satellite, observation string, epoch BiasEpoch) (BiasLookup, error) {
	if s == nil || s.native == nil {
		return BiasLookup{}, ErrClosed
	}
	value, err := s.native.CodeOSBLookup(satellite, observation, native.BiasEpoch{Year: epoch.Year, DayOfYear: epoch.DayOfYear, SecondOfDay: epoch.SecondOfDay})
	return publicBiasLookup(value), publicError(err)
}

// PhaseOSBLookup returns the complete outcome and record provenance. Supply an explicit carrier frequency only when one is known.
func (s *BiasSet) PhaseOSBLookup(satellite, observation string, epoch BiasEpoch, hasCarrier bool, carrierHz float64) (BiasLookup, error) {
	if s == nil || s.native == nil {
		return BiasLookup{}, ErrClosed
	}
	value, err := s.native.PhaseOSBLookup(satellite, observation, native.BiasEpoch{Year: epoch.Year, DayOfYear: epoch.DayOfYear, SecondOfDay: epoch.SecondOfDay}, hasCarrier, carrierHz)
	return publicBiasLookup(value), publicError(err)
}

// CodeDSBLookup returns the complete outcome and record provenance for a code DSB query.
func (s *BiasSet) CodeDSBLookup(satellite, observation1, observation2 string, epoch BiasEpoch) (BiasLookup, error) {
	if s == nil || s.native == nil {
		return BiasLookup{}, ErrClosed
	}
	value, err := s.native.CodeDSBLookup(satellite, observation1, observation2, native.BiasEpoch{Year: epoch.Year, DayOfYear: epoch.DayOfYear, SecondOfDay: epoch.SecondOfDay})
	return publicBiasLookup(value), publicError(err)
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
