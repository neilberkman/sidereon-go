package sidereon

import "sidereon.dev/go/v3/internal/native"

// SP3WriteErrorKind identifies a typed refusal to serialize an SP3 product.
type SP3WriteErrorKind uint32

const (
	// SP3WriteErrorNone means the product was written.
	SP3WriteErrorNone SP3WriteErrorKind = 0
	// SP3WriteErrorTextNotColumnSafe identifies unsupported text bytes.
	SP3WriteErrorTextNotColumnSafe SP3WriteErrorKind = 2
	// SP3WriteErrorTextNotColumnStable identifies text changed by trimming.
	SP3WriteErrorTextNotColumnStable SP3WriteErrorKind = 3
	// SP3WriteErrorBlankDescriptor identifies an optional blank descriptor.
	SP3WriteErrorBlankDescriptor SP3WriteErrorKind = 4
	// SP3WriteErrorEmptyComment identifies a comment that encodes as padding.
	SP3WriteErrorEmptyComment SP3WriteErrorKind = 5
	// SP3WriteErrorTextTooWide identifies text wider than its record columns.
	SP3WriteErrorTextTooWide SP3WriteErrorKind = 6
	// SP3WriteErrorIntegerTooWide identifies an integer wider than its columns.
	SP3WriteErrorIntegerTooWide SP3WriteErrorKind = 7
	// SP3WriteErrorNonFinite identifies a non-finite header number.
	SP3WriteErrorNonFinite SP3WriteErrorKind = 8
	// SP3WriteErrorNumberTooWide identifies a number wider than its field.
	SP3WriteErrorNumberTooWide SP3WriteErrorKind = 9
	// SP3WriteErrorPrecisionNotRepresentable identifies precision lost in formatting.
	SP3WriteErrorPrecisionNotRepresentable SP3WriteErrorKind = 10
	// SP3WriteErrorYearNotRepresentable identifies a calendar year outside four digits.
	SP3WriteErrorYearNotRepresentable SP3WriteErrorKind = 11
	// SP3WriteErrorEpochNotRestatable identifies an epoch that formats to another instant.
	SP3WriteErrorEpochNotRestatable SP3WriteErrorKind = 13
	// SP3WriteErrorEpochTimeScaleMismatch identifies an epoch/header scale mismatch.
	SP3WriteErrorEpochTimeScaleMismatch SP3WriteErrorKind = 14
	// SP3WriteErrorHeaderTimeScaleMismatch identifies a header scale inconsistency.
	SP3WriteErrorHeaderTimeScaleMismatch SP3WriteErrorKind = 15
	// SP3WriteErrorEpochCountMismatch identifies different declared and stored epoch counts.
	SP3WriteErrorEpochCountMismatch SP3WriteErrorKind = 16
	// SP3WriteErrorAccuracyCodeCountMismatch identifies unaligned accuracy codes.
	SP3WriteErrorAccuracyCodeCountMismatch SP3WriteErrorKind = 17
	// SP3WriteErrorDuplicateSatellite identifies a duplicate header satellite.
	SP3WriteErrorDuplicateSatellite SP3WriteErrorKind = 18
	// SP3WriteErrorEpochArrayLengthMismatch identifies a per-epoch array length mismatch.
	SP3WriteErrorEpochArrayLengthMismatch SP3WriteErrorKind = 19
	// SP3WriteErrorUndeclaredSatelliteRecord identifies a record for an undeclared satellite.
	SP3WriteErrorUndeclaredSatelliteRecord SP3WriteErrorKind = 20
	// SP3WriteErrorConflictingRecords identifies conflicting same-epoch records.
	SP3WriteErrorConflictingRecords SP3WriteErrorKind = 21
	// SP3WriteErrorVelocityInPositionProduct identifies velocity in a position-only product.
	SP3WriteErrorVelocityInPositionProduct SP3WriteErrorKind = 22
	// SP3WriteErrorRecordValueNonFinite identifies a non-finite record value.
	SP3WriteErrorRecordValueNonFinite SP3WriteErrorKind = 23
	// SP3WriteErrorRecordValueTooWide identifies a record value wider than its columns.
	SP3WriteErrorRecordValueTooWide SP3WriteErrorKind = 24
	// SP3WriteErrorRecordValueNotRepresentable identifies a record value changed by formatting.
	SP3WriteErrorRecordValueNotRepresentable SP3WriteErrorKind = 25
	// SP3WriteErrorRecordReadsAsAbsent identifies a value formatted as an absence sentinel.
	SP3WriteErrorRecordReadsAsAbsent SP3WriteErrorKind = 26
	// SP3WriteErrorRecordFieldsDisagree identifies inconsistent product/native units.
	SP3WriteErrorRecordFieldsDisagree SP3WriteErrorKind = 27
	// SP3WriteErrorSatelliteNotRepresentable identifies an invalid satellite token.
	SP3WriteErrorSatelliteNotRepresentable SP3WriteErrorKind = 28
	// SP3WriteErrorAccuracyNotRepresentable identifies an accuracy code that cannot be written.
	SP3WriteErrorAccuracyNotRepresentable SP3WriteErrorKind = 29
	// SP3WriteErrorAccuracyRecordMismatch identifies inconsistent accuracy records.
	SP3WriteErrorAccuracyRecordMismatch SP3WriteErrorKind = 30
	// SP3WriteErrorAccuracyBasisMissing identifies an absent accuracy basis.
	SP3WriteErrorAccuracyBasisMissing SP3WriteErrorKind = 31
	// SP3WriteErrorUnknown identifies a refusal not represented by a known variant.
	SP3WriteErrorUnknown SP3WriteErrorKind = 999
)

// SP3WriteError retains the precise refused field, record location, and values.
type SP3WriteError struct {
	// Kind is the typed reason the formatter refused the product.
	Kind SP3WriteErrorKind
	// HasField says Field identifies the refused product field.
	HasField bool
	// HasTextValue says TextValue contains the rejected text.
	HasTextValue bool
	// HasSatelliteID says SatelliteID identifies the affected satellite.
	HasSatelliteID bool
	// HasEpochIndex says EpochIndex identifies the affected epoch.
	HasEpochIndex bool
	// HasCommentIndex says CommentIndex identifies the affected comment.
	HasCommentIndex bool
	// HasColumns says Columns contains the field width.
	HasColumns bool
	// HasDecimals says Decimals contains the decimal precision.
	HasDecimals bool
	// HasIntegerValue says IntegerValue contains the rejected integer.
	HasIntegerValue bool
	// HasNumber says Number contains the rejected numeric value.
	HasNumber bool
	// HasYear says Year contains the rejected calendar year.
	HasYear bool
	// HasFieldSeconds says FieldSeconds contains the rejected seconds field.
	HasFieldSeconds bool
	// HasResidualS says ResidualS contains the rejected residual.
	HasResidualS bool
	// HasEpochTimeScale says EpochTimeScale identifies the affected epoch scale.
	HasEpochTimeScale bool
	// HasHeaderTimeScale says HeaderTimeScale identifies the header scale.
	HasHeaderTimeScale bool
	// HasTimeSystem says TimeSystem identifies the affected time system.
	HasTimeSystem bool
	// HasDeclaredEpochs says DeclaredEpochs contains the header epoch count.
	HasDeclaredEpochs bool
	// HasEpochs says Epochs contains the stored epoch count.
	HasEpochs bool
	// HasEntries says Entries contains the stored record count.
	HasEntries bool
	// HasSatellites says Satellites contains the declared satellite count.
	HasSatellites bool
	// HasCodes says Codes contains the accuracy-code count.
	HasCodes bool
	// HasStored says Stored contains the product value.
	HasStored bool
	// HasNative says Native contains the corresponding native-unit value.
	HasNative bool
	// HasColumnValue says ColumnValue contains the formatted field value.
	HasColumnValue bool
	// HasExponent says Exponent contains the rejected exponent.
	HasExponent bool
	// SatelliteID is the affected satellite identifier.
	SatelliteID string
	// TimeSystem is the affected time-system label.
	TimeSystem string
	// EpochIndex is the zero-based affected epoch.
	EpochIndex int
	// CommentIndex is the zero-based affected comment.
	CommentIndex int
	// Columns is the fixed-width field width.
	Columns int
	// Decimals is the decimal precision of the field.
	Decimals int
	// IntegerValue is the rejected integer.
	IntegerValue uint64
	// Number is the rejected floating-point value.
	Number float64
	// Year is the rejected calendar year.
	Year int64
	// FieldSeconds is the rejected epoch seconds field.
	FieldSeconds float64
	// ResidualS is the rejected residual in seconds.
	ResidualS float64
	// EpochTimeScale is the affected epoch's time scale.
	EpochTimeScale TimeScale
	// HeaderTimeScale is the header's time scale.
	HeaderTimeScale TimeScale
	// DeclaredEpochs is the header's epoch count.
	DeclaredEpochs uint64
	// Epochs is the retained epoch count.
	Epochs int
	// Entries is the retained satellite-record count.
	Entries int
	// Satellites is the declared satellite count.
	Satellites int
	// Codes is the retained accuracy-code count.
	Codes int
	// Stored is the value retained by the product.
	Stored float64
	// Native is the corresponding native-unit value.
	Native float64
	// ColumnValue is the value produced by field formatting.
	ColumnValue float64
	// Exponent is the rejected decimal exponent.
	Exponent int16
	// Message is the complete diagnostic text.
	Message string
	// Field is the refused field name.
	Field string
	// TextValue is the rejected field text.
	TextValue string
}

// SP3WriteOutcome records a successful text output or typed refusal.
type SP3WriteOutcome struct {
	// IsOK reports whether serialization succeeded.
	IsOK bool
	// Status is the C status code recorded with the outcome.
	Status uint32
	// Error retains the owned typed refusal details.
	Error SP3WriteError
}

func publicSP3WriteOutcome(value native.SP3WriteOutcome) SP3WriteOutcome {
	e := value.Error
	return SP3WriteOutcome{IsOK: value.IsOK, Status: value.Status, Error: SP3WriteError{Kind: SP3WriteErrorKind(e.Kind), HasField: e.HasField, HasTextValue: e.HasTextValue, HasSatelliteID: e.HasSatelliteID, HasEpochIndex: e.HasEpochIndex, HasCommentIndex: e.HasCommentIndex, HasColumns: e.HasColumns, HasDecimals: e.HasDecimals, HasIntegerValue: e.HasIntegerValue, HasNumber: e.HasNumber, HasYear: e.HasYear, HasFieldSeconds: e.HasFieldSeconds, HasResidualS: e.HasResidualS, HasEpochTimeScale: e.HasEpochTimeScale, HasHeaderTimeScale: e.HasHeaderTimeScale, HasTimeSystem: e.HasTimeSystem, HasDeclaredEpochs: e.HasDeclaredEpochs, HasEpochs: e.HasEpochs, HasEntries: e.HasEntries, HasSatellites: e.HasSatellites, HasCodes: e.HasCodes, HasStored: e.HasStored, HasNative: e.HasNative, HasColumnValue: e.HasColumnValue, HasExponent: e.HasExponent, SatelliteID: e.SatelliteID, TimeSystem: e.TimeSystem, EpochIndex: e.EpochIndex, CommentIndex: e.CommentIndex, Columns: e.Columns, Decimals: e.Decimals, IntegerValue: e.IntegerValue, Number: e.Number, Year: e.Year, FieldSeconds: e.FieldSeconds, ResidualS: e.ResidualS, EpochTimeScale: TimeScale(e.EpochTimeScale), HeaderTimeScale: TimeScale(e.HeaderTimeScale), DeclaredEpochs: e.DeclaredEpochs, Epochs: e.Epochs, Entries: e.Entries, Satellites: e.Satellites, Codes: e.Codes, Stored: e.Stored, Native: e.Native, ColumnValue: e.ColumnValue, Exponent: e.Exponent, Message: e.Message, Field: e.Field, TextValue: e.TextValue}}
}

// Encode writes the parsed product and returns an owned typed refusal when its fields cannot be restated exactly.
func (s *SP3) Encode() ([]byte, SP3WriteOutcome, error) {
	if s == nil || s.handle == nil {
		return nil, SP3WriteOutcome{}, ErrClosed
	}
	text, outcome, err := s.handle.TextResult()
	return text, publicSP3WriteOutcome(outcome), publicError(err)
}

// ExactSP3Coverage identifies whether the declared end epoch is excluded or included.
type ExactSP3Coverage uint32

const (
	// ExactSP3CoverageHalfOpen selects half-open SP3 coverage intervals.
	ExactSP3CoverageHalfOpen ExactSP3Coverage = ExactSP3Coverage(native.ExactSp3CoverageHalfOpen)
	// ExactSP3CoverageInclusive selects inclusive SP3 coverage intervals.
	ExactSP3CoverageInclusive ExactSP3Coverage = ExactSP3Coverage(native.ExactSp3CoverageInclusive)
)

// ExactSP3Request is an owning, validated exact-artifact request.
type ExactSP3Request struct {
	_      noCopy
	handle *native.ExactSp3Request
}

// NewExactSP3Request creates an exact SP3 request. Span and sample are IGS
// period tokens; empty issue and agency select the native defaults.
func NewExactSP3Request(year int, month, day uint8, issue, span, sample, expectedAgency string) (*ExactSP3Request, error) {
	h, err := native.ExactSp3RequestNew(year, month, day, issue, span, sample, expectedAgency)
	if err != nil {
		return nil, publicError(err)
	}
	return &ExactSP3Request{handle: h}, nil
}

// ExactSP3RequestFromIdentity creates a request from a complete product identity.
func ExactSP3RequestFromIdentity(identity ProductIdentity) (*ExactSP3Request, error) {
	nativeIdentity, err := identityToNative(identity)
	if err != nil {
		return nil, publicError(err)
	}
	h, err := native.ExactSp3RequestFromIdentity(nativeIdentity)
	if err != nil {
		return nil, publicError(err)
	}
	return &ExactSP3Request{handle: h}, nil
}

// Close releases the exact request and is safe to call repeatedly.
func (r *ExactSP3Request) Close() error {
	if r == nil || r.handle == nil {
		return nil
	}
	return publicError(r.handle.Close())
}

// LoadExactSP3 parses and validates bytes against an exact request.
func LoadExactSP3(data []byte, request *ExactSP3Request, opts ...SP3Option) (*SP3, ExactSP3Coverage, error) {
	if len(opts) == 0 {
		if request == nil || request.handle == nil {
			return nil, 0, ErrClosed
		}
		h, coverage, err := native.LoadExactSP3(append([]byte(nil), data...), request.handle)
		if err != nil {
			return nil, 0, publicError(err)
		}
		return &SP3{handle: h}, ExactSP3Coverage(coverage), nil
	}
	return LoadExactSP3WithOptions(data, request, resolveSP3Options(opts))
}

// LoadExactSP3WithOptions parses and validates bytes against an exact request using explicit interpolation options.
func LoadExactSP3WithOptions(data []byte, request *ExactSP3Request, options SP3InterpolationOptions) (*SP3, ExactSP3Coverage, error) {
	if request == nil || request.handle == nil {
		return nil, 0, ErrClosed
	}
	h, coverage, err := native.LoadExactSP3WithGapThresholdFactor(append([]byte(nil), data...), request.handle, options.GapThresholdFactor)
	if err != nil {
		return nil, 0, publicError(err)
	}
	return &SP3{handle: h}, ExactSP3Coverage(coverage), nil
}

// ValidateExact validates an already-loaded SP3 against an exact request and
// returns the native boundary convention.
func (s *SP3) ValidateExact(request *ExactSP3Request) (ExactSP3Coverage, error) {
	if s == nil || s.handle == nil || request == nil || request.handle == nil {
		return 0, ErrClosed
	}
	coverage, err := native.ValidateExactSP3(s.handle, request.handle)
	return ExactSP3Coverage(coverage), publicError(err)
}

// ValidateExactSP3 is the package-level form of SP3.ValidateExact.
func ValidateExactSP3(sp3 *SP3, request *ExactSP3Request) (ExactSP3Coverage, error) {
	return sp3.ValidateExact(request)
}

// SP3MergeCombine identifies how agreeing sources are combined.
type SP3MergeCombine uint32

const (
	// SP3MergeCombineMean selects mean combination of overlapping SP3 records.
	SP3MergeCombineMean SP3MergeCombine = SP3MergeCombine(native.Sp3MergeCombineMean)
	// SP3MergeCombineMedian selects median combination of overlapping SP3 records.
	SP3MergeCombineMedian SP3MergeCombine = SP3MergeCombine(native.Sp3MergeCombineMedian)
	// SP3MergeCombinePrecedence selects precedence-based combination of overlapping SP3 records.
	SP3MergeCombinePrecedence SP3MergeCombine = SP3MergeCombine(native.Sp3MergeCombinePrecedence)
)

// SP3MergePrecedenceScope controls precedence ownership.
type SP3MergePrecedenceScope uint32

const (
	// SP3MergePrecedenceCell selects cell-level source precedence.
	SP3MergePrecedenceCell SP3MergePrecedenceScope = SP3MergePrecedenceScope(native.Sp3MergePrecedenceCell)
	// SP3MergePrecedenceSatelliteArc selects satellite-arc source precedence.
	SP3MergePrecedenceSatelliteArc SP3MergePrecedenceScope = SP3MergePrecedenceScope(native.Sp3MergePrecedenceSatelliteArc)
)

// SP3MergeFlagKind selects one merge audit flag list.
type SP3MergeFlagKind uint32

const (
	// SP3MergeFlagQuarantined selects the quarantined merge flag.
	SP3MergeFlagQuarantined SP3MergeFlagKind = SP3MergeFlagKind(native.Sp3MergeFlagQuarantined)
	// SP3MergeFlagSingleSource selects the single-source merge flag.
	SP3MergeFlagSingleSource SP3MergeFlagKind = SP3MergeFlagKind(native.Sp3MergeFlagSingleSource)
	// SP3MergeFlagPositionOutlier selects the position-outlier merge flag.
	SP3MergeFlagPositionOutlier SP3MergeFlagKind = SP3MergeFlagKind(native.Sp3MergeFlagPositionOutlier)
	// SP3MergeFlagClockOutlier selects the clock-outlier merge flag.
	SP3MergeFlagClockOutlier SP3MergeFlagKind = SP3MergeFlagKind(native.Sp3MergeFlagClockOutlier)
)

// SP3ProvenanceMode selects the merge provenance retained in its report.
type SP3ProvenanceMode uint32

const (
	// SP3ProvenanceOff omits provenance.
	SP3ProvenanceOff SP3ProvenanceMode = SP3ProvenanceMode(native.Sp3ProvenanceModeOff)
	// SP3ProvenanceSummary retains contributor coverage and source transitions.
	SP3ProvenanceSummary SP3ProvenanceMode = SP3ProvenanceMode(native.Sp3ProvenanceModeSummary)
	// SP3ProvenanceFull also retains every accepted cell's source selection.
	SP3ProvenanceFull SP3ProvenanceMode = SP3ProvenanceMode(native.Sp3ProvenanceModeFull)
)

// SP3FrameReconciliationMethod identifies how coordinate labels were reconciled.
type SP3FrameReconciliationMethod uint32

const (
	// SP3FrameReconciliationAssertedEquivalence selects asserted frame equivalence.
	SP3FrameReconciliationAssertedEquivalence SP3FrameReconciliationMethod = SP3FrameReconciliationMethod(native.Sp3FrameReconciliationAssertedEquivalence)
	// SP3FrameReconciliationHelmert selects Helmert frame reconciliation.
	SP3FrameReconciliationHelmert SP3FrameReconciliationMethod = SP3FrameReconciliationMethod(native.Sp3FrameReconciliationHelmert)
)

// SP3MergeOptions controls native SP3 consensus merging. Distances are metres,
// clock tolerances are seconds, and target intervals are seconds.
type SP3MergeOptions struct {
	// PositionToleranceM is the maximum position disagreement accepted in metres.
	PositionToleranceM float64
	// ClockToleranceS is the clock disagreement tolerance in seconds.
	ClockToleranceS float64
	// MinAgree is the minimum number of sources that must agree on a cell.
	MinAgree int
	// ClockMinCommon is the minimum number of common clock samples for clock agreement.
	ClockMinCommon int
	// Combine selects mean, median, or precedence-based overlap combination.
	Combine SP3MergeCombine
	// PrecedenceScope selects whether precedence applies per cell or satellite arc.
	PrecedenceScope SP3MergePrecedenceScope
	// OutlierRejectEnabled reports whether position/clock outlier rejection is enabled.
	OutlierRejectEnabled bool
	// OutlierRejectPositionTolerance is the position outlier threshold in metres.
	OutlierRejectPositionTolerance float64
	// OutlierRejectClockTolerance is the clock outlier threshold in seconds.
	OutlierRejectClockTolerance float64
	// TargetEpochIntervalEnabled reports whether target-epoch interval filtering is enabled.
	TargetEpochIntervalEnabled bool
	// TargetEpochIntervalS is the desired SP3 epoch spacing in seconds when enabled.
	TargetEpochIntervalS float64
	// Systems optionally filters the merge to the listed GNSS constellations; empty disables filtering.
	Systems []GNSSSystem
	// AssertedFrameLabelSets contains caller-asserted equivalence groups of two or more coordinate-frame labels.
	AssertedFrameLabelSets [][]string
	// HelmertFrameReconciliation reports whether Helmert frame reconciliation is enabled.
	HelmertFrameReconciliation bool
	// ProvenanceMode selects whether the report retains no, summary, or full provenance.
	ProvenanceMode SP3ProvenanceMode
}

// NewSP3MergeOptions returns native merge defaults.
func NewSP3MergeOptions() (SP3MergeOptions, error) {
	v, err := native.Sp3MergeOptionsInit()
	if err != nil {
		return SP3MergeOptions{}, publicError(err)
	}
	return publicSP3MergeOptions(v), nil
}

func publicSP3MergeOptions(v native.NativeSp3MergeOptions) SP3MergeOptions {
	return SP3MergeOptions{PositionToleranceM: v.PositionToleranceM, ClockToleranceS: v.ClockToleranceS, MinAgree: v.MinAgree, ClockMinCommon: v.ClockMinCommon, Combine: SP3MergeCombine(v.Combine), PrecedenceScope: SP3MergePrecedenceScope(v.PrecedenceScope), OutlierRejectEnabled: v.OutlierRejectEnabled, OutlierRejectPositionTolerance: v.OutlierRejectPositionTolerance, OutlierRejectClockTolerance: v.OutlierRejectClockTolerance, TargetEpochIntervalEnabled: v.TargetEpochIntervalEnabled, TargetEpochIntervalS: v.TargetEpochIntervalS, HelmertFrameReconciliation: v.HelmertFrameReconciliation, ProvenanceMode: SP3ProvenanceMode(v.ProvenanceMode)}
}

func nativeSP3MergeOptions(v *SP3MergeOptions) *native.NativeSp3MergeOptions {
	if v == nil {
		return nil
	}
	systems := make([]uint32, len(v.Systems))
	for i := range v.Systems {
		systems[i] = uint32(v.Systems[i])
	}
	return &native.NativeSp3MergeOptions{PositionToleranceM: v.PositionToleranceM, ClockToleranceS: v.ClockToleranceS, MinAgree: v.MinAgree, ClockMinCommon: v.ClockMinCommon, Combine: uint32(v.Combine), PrecedenceScope: uint32(v.PrecedenceScope), OutlierRejectEnabled: v.OutlierRejectEnabled, OutlierRejectPositionTolerance: v.OutlierRejectPositionTolerance, OutlierRejectClockTolerance: v.OutlierRejectClockTolerance, TargetEpochIntervalEnabled: v.TargetEpochIntervalEnabled, TargetEpochIntervalS: v.TargetEpochIntervalS, Systems: systems, AssertedFrameLabelSets: append([][]string(nil), v.AssertedFrameLabelSets...), HelmertFrameReconciliation: v.HelmertFrameReconciliation, ProvenanceMode: uint32(v.ProvenanceMode)}
}

// SP3ArtifactIdentity describes one verified SP3 merge contributor.
type SP3ArtifactIdentity struct {
	RequestedIdentity ProductIdentity
	ResolvedIdentity  ProductIdentity
	Distribution      DistributionSource
	// OfficialFilename is the provider's canonical SP3 filename.
	OfficialFilename string
	// ProductSHA256 is the lowercase SHA-256 digest of the SP3 product bytes.
	ProductSHA256 string
	// ProductByteLength is the byte length of the uncompressed SP3 product.
	ProductByteLength uint64
	// ArchiveSHA256 is the lowercase SHA-256 digest of the compressed archive bytes.
	ArchiveSHA256 string
	// ArchiveByteLength is the byte length of the downloaded archive.
	ArchiveByteLength uint64
	Compression       ArchiveCompression
}

func nativeSP3ArtifactIdentity(v SP3ArtifactIdentity) (native.NativeSp3ArtifactIdentity, error) {
	requested, err := identityToNative(v.RequestedIdentity)
	if err != nil {
		return native.NativeSp3ArtifactIdentity{}, err
	}
	resolved, err := identityToNative(v.ResolvedIdentity)
	if err != nil {
		return native.NativeSp3ArtifactIdentity{}, err
	}
	return native.NativeSp3ArtifactIdentity{RequestedIdentity: requested, ResolvedIdentity: resolved, Distribution: uint32(v.Distribution), OfficialFilename: v.OfficialFilename, ProductSHA256: v.ProductSHA256, ProductByteLength: v.ProductByteLength, ArchiveSHA256: v.ArchiveSHA256, ArchiveByteLength: v.ArchiveByteLength, Compression: uint32(v.Compression)}, nil
}

func publicSP3ArtifactIdentity(v native.NativeSp3ArtifactIdentity) SP3ArtifactIdentity {
	return SP3ArtifactIdentity{RequestedIdentity: publicIdentity(v.RequestedIdentity), ResolvedIdentity: publicIdentity(v.ResolvedIdentity), Distribution: DistributionSource(v.Distribution), OfficialFilename: v.OfficialFilename, ProductSHA256: v.ProductSHA256, ProductByteLength: v.ProductByteLength, ArchiveSHA256: v.ArchiveSHA256, ArchiveByteLength: v.ArchiveByteLength, Compression: ArchiveCompression(v.Compression)}
}

// SP3MergeInputIdentity is an owning canonical identity for merge contributors.
type SP3MergeInputIdentity struct {
	_      noCopy
	handle *native.Sp3MergeInputIdentity
}

// BuildSP3MergeInputIdentity validates and canonicalizes merge contributors and policy.
func BuildSP3MergeInputIdentity(contributors []SP3ArtifactIdentity, options *SP3MergeOptions) (*SP3MergeInputIdentity, error) {
	values := make([]native.NativeSp3ArtifactIdentity, len(contributors))
	for i := range contributors {
		var err error
		values[i], err = nativeSP3ArtifactIdentity(contributors[i])
		if err != nil {
			return nil, publicError(err)
		}
	}
	h, err := native.MergeInputIdentity(values, nativeSP3MergeOptions(options))
	if err != nil {
		return nil, publicError(err)
	}
	return &SP3MergeInputIdentity{handle: h}, nil
}

// Close releases the merge input identity and is safe to call repeatedly.
func (i *SP3MergeInputIdentity) Close() error {
	if i == nil || i.handle == nil {
		return nil
	}
	return publicError(i.handle.Close())
}

// ContributorCount returns the canonical contributor count.
func (i *SP3MergeInputIdentity) ContributorCount() (int, error) {
	if i == nil || i.handle == nil {
		return 0, ErrClosed
	}
	v, err := i.handle.ContributorCount()
	return v, publicError(err)
}

// Contributor returns one canonical contributor.
func (i *SP3MergeInputIdentity) Contributor(index int) (SP3ArtifactIdentity, error) {
	if i == nil || i.handle == nil {
		return SP3ArtifactIdentity{}, ErrClosed
	}
	v, err := i.handle.Contributor(index)
	return publicSP3ArtifactIdentity(v), publicError(err)
}

// PrecedenceContributorCount reports whether ordered precedence contributors exist.
func (i *SP3MergeInputIdentity) PrecedenceContributorCount() (bool, int, error) {
	if i == nil || i.handle == nil {
		return false, 0, ErrClosed
	}
	present, count, err := i.handle.PrecedenceContributorCount()
	return present, count, publicError(err)
}

// PrecedenceContributor returns one ordered precedence contributor.
func (i *SP3MergeInputIdentity) PrecedenceContributor(index int) (SP3ArtifactIdentity, error) {
	if i == nil || i.handle == nil {
		return SP3ArtifactIdentity{}, ErrClosed
	}
	v, err := i.handle.PrecedenceContributor(index)
	return publicSP3ArtifactIdentity(v), publicError(err)
}

// SchemaVersion returns the native merge identity schema version.
func (i *SP3MergeInputIdentity) SchemaVersion() (uint8, error) {
	if i == nil || i.handle == nil {
		return 0, ErrClosed
	}
	v, err := i.handle.SchemaVersion()
	return v, publicError(err)
}

// StableID returns detached canonical identity bytes.
func (i *SP3MergeInputIdentity) StableID() ([]byte, error) {
	if i == nil || i.handle == nil {
		return nil, ErrClosed
	}
	v, err := i.handle.StableID()
	return append([]byte(nil), v...), publicError(err)
}

// SP3MergeReport is an owning native merge audit report.
type SP3MergeReport struct {
	_      noCopy
	handle *native.Sp3MergeReport
}

// Close releases the merge report and is safe to call repeatedly.
func (r *SP3MergeReport) Close() error {
	if r == nil || r.handle == nil {
		return nil
	}
	return publicError(r.handle.Close())
}

// SP3AgreementSummary is a whole-product merge agreement summary.
type SP3AgreementSummary struct {
	// PositionRMSPresent reports whether PositionRMSM is valid.
	PositionRMSPresent bool
	// PositionRMSM is the RMS position disagreement in metres.
	PositionRMSM float64
	// PositionMaxPresent reports whether PositionMaxM is valid.
	PositionMaxPresent bool
	// PositionMaxM is the maximum position disagreement in metres.
	PositionMaxM float64
	// ClockRMSPresent reports whether ClockRMSS is valid.
	ClockRMSPresent bool
	ClockRMSS       float64
	// ClockMaxPresent reports whether ClockMaxS is valid.
	ClockMaxPresent bool
	ClockMaxS       float64
}

// AgreementSummary returns detached position and clock agreement metrics.
func (r *SP3MergeReport) AgreementSummary() (SP3AgreementSummary, error) {
	if r == nil || r.handle == nil {
		return SP3AgreementSummary{}, ErrClosed
	}
	v, err := r.handle.AgreementSummary()
	return SP3AgreementSummary{PositionRMSPresent: v.PositionRMSPresent, PositionRMSM: v.PositionRMSM, PositionMaxPresent: v.PositionMaxPresent, PositionMaxM: v.PositionMaxM, ClockRMSPresent: v.ClockRMSPresent, ClockRMSS: v.ClockRMSS, ClockMaxPresent: v.ClockMaxPresent, ClockMaxS: v.ClockMaxS}, publicError(err)
}

// SP3EpochAgreement is one per-epoch merge agreement row.
type SP3EpochAgreement struct {
	// EpochJ2000S is the agreement epoch in seconds from J2000.
	EpochJ2000S float64
	// Satellites is the number of satellite records contributing to this epoch row.
	Satellites int
	// PositionRMSM is the RMS position disagreement in metres.
	PositionRMSM float64
	// PositionMaxM is the maximum position disagreement in metres.
	PositionMaxM float64
	// ClockRMSPresent reports whether ClockRMSS is valid.
	ClockRMSPresent bool
	ClockRMSS       float64
	// ClockMaxPresent reports whether ClockMaxS is valid.
	ClockMaxPresent bool
	ClockMaxS       float64
}

// EpochAgreementCount returns the number of per-epoch rows.
func (r *SP3MergeReport) EpochAgreementCount() (int, error) {
	if r == nil || r.handle == nil {
		return 0, ErrClosed
	}
	v, err := r.handle.EpochAgreementCount()
	return v, publicError(err)
}

// EpochAgreement returns one detached per-epoch row.
func (r *SP3MergeReport) EpochAgreement(index int) (SP3EpochAgreement, error) {
	if r == nil || r.handle == nil {
		return SP3EpochAgreement{}, ErrClosed
	}
	v, err := r.handle.EpochAgreement(index)
	return SP3EpochAgreement{EpochJ2000S: v.EpochJ2000S, Satellites: v.Satellites, PositionRMSM: v.PositionRMSM, PositionMaxM: v.PositionMaxM, ClockRMSPresent: v.ClockRMSPresent, ClockRMSS: v.ClockRMSS, ClockMaxPresent: v.ClockMaxPresent, ClockMaxS: v.ClockMaxS}, publicError(err)
}

// SP3MergeFlag is one merge audit flag.
type SP3MergeFlag struct {
	// EpochJ2000S is the merge-flag epoch in seconds from J2000.
	EpochJ2000S float64
	// SatelliteID is the GNSS satellite identifier.
	SatelliteID string
	// SourceCount is the number of source products contributing to the flagged cell.
	SourceCount int
}

// FlagCount returns the size of one merge audit list.
func (r *SP3MergeReport) FlagCount(kind SP3MergeFlagKind) (int, error) {
	if r == nil || r.handle == nil {
		return 0, ErrClosed
	}
	v, err := r.handle.FlagCount(uint32(kind))
	return v, publicError(err)
}

// Flag returns one detached merge audit flag.
func (r *SP3MergeReport) Flag(kind SP3MergeFlagKind, index int) (SP3MergeFlag, error) {
	if r == nil || r.handle == nil {
		return SP3MergeFlag{}, ErrClosed
	}
	v, err := r.handle.Flag(uint32(kind), index)
	return SP3MergeFlag{EpochJ2000S: v.EpochJ2000S, SatelliteID: v.Satellite, SourceCount: v.SourceCount}, publicError(err)
}

// FlagSources returns detached source indices for one merge audit flag.
func (r *SP3MergeReport) FlagSources(kind SP3MergeFlagKind, index int) ([]int, error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	v, err := r.handle.FlagSources(uint32(kind), index)
	return append([]int(nil), v...), publicError(err)
}

// SP3FrameReconciliation is one merge coordinate-label reconciliation row.
type SP3FrameReconciliation struct {
	// SourceIndex is the zero-based source-product index.
	SourceIndex int
	// SourceLabelLen is the byte length of the source frame label.
	SourceLabelLen int
	// TargetLabelLen is the byte length of the target frame label.
	TargetLabelLen int
	// Method identifies asserted equivalence or Helmert reconciliation.
	Method SP3FrameReconciliationMethod
	// AssertedLabelCount is the number of caller-asserted labels considered.
	AssertedLabelCount int
	// SourceFramePresent reports whether SourceFrame is valid.
	SourceFramePresent bool
	// SourceFrame identifies the source product's terrestrial coordinate frame.
	SourceFrame TerrestrialFrame
	// TargetFramePresent reports whether TargetFrame is valid.
	TargetFramePresent bool
	// TargetFrame identifies the reconciled terrestrial coordinate frame.
	TargetFrame TerrestrialFrame
	// CatalogFramePresent reports whether catalog frame metadata is valid.
	CatalogFramePresent bool
	// CatalogSourceFrame identifies the source frame named by the frame catalog.
	CatalogSourceFrame TerrestrialFrame
	// CatalogTargetFrame identifies the target frame named by the frame catalog.
	CatalogTargetFrame TerrestrialFrame
	// CatalogInverse reports whether the catalog transform is inverse.
	CatalogInverse bool
	// ReferenceEpochYearPresent reports whether ReferenceEpochYear is valid.
	ReferenceEpochYearPresent bool
	// ReferenceEpochYear is the decimal year at which Helmert parameters apply.
	ReferenceEpochYear float64
	// ParametersPresent reports whether Helmert parameters are valid.
	ParametersPresent bool
	// TranslationMM contains translation components in millimetres.
	TranslationMM [3]float64
	// ScalePPB contains parts per billion.
	ScalePPB float64
	// RotationMAS contains milliarcseconds.
	RotationMAS [3]float64
	// RatesPresent reports whether Helmert rates are valid.
	RatesPresent bool
	// TranslationMMPerYear contains translation rates in millimetres per year.
	TranslationMMPerYear [3]float64
	// ScalePPBPerYear is the scale rate in parts per billion per year.
	ScalePPBPerYear float64
	// RotationMASPerYear contains rotation rates in milliarcseconds per year.
	RotationMASPerYear [3]float64
	// ProvenanceLen is the byte length of the transform provenance text.
	ProvenanceLen int
	// EpochYearSpanPresent reports whether the epoch-year span is valid.
	EpochYearSpanPresent bool
	// EpochYearStart is the start of the transform validity span in decimal years.
	EpochYearStart float64
	// EpochYearEnd is the end of the transform validity span in decimal years.
	EpochYearEnd float64
	// RecordsAffected is the number of merged records covered by the reconciliation.
	RecordsAffected int
	// Identity reports whether the frame transform is the identity.
	Identity bool
}

func publicSP3FrameReconciliation(v native.NativeSp3FrameReconciliation) SP3FrameReconciliation {
	return SP3FrameReconciliation{SourceIndex: v.SourceIndex, SourceLabelLen: v.SourceLabelLen, TargetLabelLen: v.TargetLabelLen, Method: SP3FrameReconciliationMethod(v.Method), AssertedLabelCount: v.AssertedLabelCount, SourceFramePresent: v.SourceFramePresent, SourceFrame: TerrestrialFrame(v.SourceFrame), TargetFramePresent: v.TargetFramePresent, TargetFrame: TerrestrialFrame(v.TargetFrame), CatalogFramePresent: v.CatalogFramePresent, CatalogSourceFrame: TerrestrialFrame(v.CatalogSourceFrame), CatalogTargetFrame: TerrestrialFrame(v.CatalogTargetFrame), CatalogInverse: v.CatalogInverse, ReferenceEpochYearPresent: v.ReferenceEpochYearPresent, ReferenceEpochYear: v.ReferenceEpochYear, ParametersPresent: v.ParametersPresent, TranslationMM: v.TranslationMM, ScalePPB: v.ScalePPB, RotationMAS: v.RotationMAS, RatesPresent: v.RatesPresent, TranslationMMPerYear: v.TranslationMMPerYear, ScalePPBPerYear: v.ScalePPBPerYear, RotationMASPerYear: v.RotationMASPerYear, ProvenanceLen: v.ProvenanceLen, EpochYearSpanPresent: v.EpochYearSpanPresent, EpochYearStart: v.EpochYearStart, EpochYearEnd: v.EpochYearEnd, RecordsAffected: v.RecordsAffected, Identity: v.Identity}
}

// FrameReconciliationCount returns the number of reconciliation rows.
func (r *SP3MergeReport) FrameReconciliationCount() (int, error) {
	if r == nil || r.handle == nil {
		return 0, ErrClosed
	}
	v, err := r.handle.FrameReconciliationCount()
	return v, publicError(err)
}

// FrameReconciliation returns one detached reconciliation row.
func (r *SP3MergeReport) FrameReconciliation(index int) (SP3FrameReconciliation, error) {
	if r == nil || r.handle == nil {
		return SP3FrameReconciliation{}, ErrClosed
	}
	v, err := r.handle.FrameReconciliation(index)
	return publicSP3FrameReconciliation(v), publicError(err)
}

// AssertedLabel returns one detached asserted coordinate label.
func (r *SP3MergeReport) AssertedLabel(index, labelIndex int) (string, error) {
	if r == nil || r.handle == nil {
		return "", ErrClosed
	}
	v, err := r.handle.AssertedLabel(index, labelIndex)
	return string(v), publicError(err)
}

// Provenance returns detached catalog provenance text.
func (r *SP3MergeReport) Provenance(index int) (string, error) {
	if r == nil || r.handle == nil {
		return "", ErrClosed
	}
	v, err := r.handle.Provenance(index)
	return string(v), publicError(err)
}

// SourceLabel returns a detached source coordinate label.
func (r *SP3MergeReport) SourceLabel(index int) (string, error) {
	if r == nil || r.handle == nil {
		return "", ErrClosed
	}
	v, err := r.handle.SourceLabel(index)
	return string(v), publicError(err)
}

// TargetLabel returns a detached target coordinate label.
func (r *SP3MergeReport) TargetLabel(index int) (string, error) {
	if r == nil || r.handle == nil {
		return "", ErrClosed
	}
	v, err := r.handle.TargetLabel(index)
	return string(v), publicError(err)
}

// ContinuityVerdictJSON returns detached merge continuity JSON bytes.
func (r *SP3MergeReport) ContinuityVerdictJSON(merged *SP3, fromJ2000S, throughJ2000S float64) ([]byte, error) {
	if r == nil || r.handle == nil || merged == nil || merged.handle == nil {
		return nil, ErrClosed
	}
	v, err := r.handle.ContinuityVerdictJSON(merged.handle, fromJ2000S, throughJ2000S)
	return append([]byte(nil), v...), publicError(err)
}

// MergeSP3 combines loaded SP3 sources through the native consensus engine.
func MergeSP3(sources []*SP3, options *SP3MergeOptions) (*SP3, *SP3MergeReport, error) {
	nativeSources := make([]*native.SP3, len(sources))
	for i, source := range sources {
		if source == nil || source.handle == nil {
			return nil, nil, ErrClosed
		}
		nativeSources[i] = source.handle
	}
	merged, report, err := native.MergeSP3(nativeSources, nativeSP3MergeOptions(options))
	if err != nil {
		return nil, nil, publicError(err)
	}
	return &SP3{handle: merged}, &SP3MergeReport{handle: report}, nil
}

// VisibilityPass describes one sampled rise/set/peak pass.
type VisibilityPass struct {
	// Satellite identifies the GNSS satellite whose visibility interval was sampled.
	Satellite string
	// RiseStepIndex and SetStepIndex are zero-based sample indices at the visibility boundaries.
	RiseStepIndex int
	SetStepIndex  int
	// PeakElevation is the maximum elevation in degrees within the sampled pass.
	PeakElevation float64
	// PeakStepIndex is the zero-based sample index at PeakElevation.
	PeakStepIndex int
}

// VisibilitySeriesPoint contains the visible count at one sampled step.
type VisibilitySeriesPoint struct {
	// StepIndex is the zero-based sample index in the requested visibility window.
	StepIndex int
	// Visible is the number of satellites above the elevation mask at StepIndex.
	Visible int
}

// GeometryVisible describes one satellite above an elevation mask.
type GeometryVisible struct {
	// Satellite identifies the GNSS satellite above the elevation mask.
	Satellite string
	// ElevationDeg is the satellite elevation above the local horizon in degrees.
	ElevationDeg float64
	// AzimuthDeg is the satellite azimuth in degrees in the local horizon frame.
	AzimuthDeg float64
}

func geometrySystems(values []GNSSSystem) ([]uint32, error) {
	result := make([]uint32, len(values))
	for i, value := range values {
		if err := validateGNSSSystem(value); err != nil {
			return nil, err
		}
		result[i] = uint32(value)
	}
	return result, nil
}

// GeometryPasses computes sampled visibility passes for an ECEF receiver in metres.
func (s *SP3) GeometryPasses(receiver ECEF, windowStart, windowEnd float64, stepSeconds uint64, elevationMaskDeg float64, systems []GNSSSystem) ([]VisibilityPass, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	nativeSystems, err := geometrySystems(systems)
	if err != nil {
		return nil, err
	}
	values, err := s.handle.GeometryPasses([3]float64{receiver.X, receiver.Y, receiver.Z}, windowStart, windowEnd, stepSeconds, elevationMaskDeg, nativeSystems)
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]VisibilityPass, len(values))
	for i, value := range values {
		result[i] = VisibilityPass{Satellite: value.Satellite, RiseStepIndex: value.RiseStepIndex, SetStepIndex: value.SetStepIndex, PeakElevation: value.PeakElevation, PeakStepIndex: value.PeakStepIndex}
	}
	return result, nil
}

// GeometryVisibilitySeries counts visible satellites over sampled epochs.
func (s *SP3) GeometryVisibilitySeries(receiver ECEF, windowStart, windowEnd float64, stepSeconds uint64, elevationMaskDeg float64, systems []GNSSSystem) ([]VisibilitySeriesPoint, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	nativeSystems, err := geometrySystems(systems)
	if err != nil {
		return nil, err
	}
	values, err := s.handle.GeometryVisibilitySeries([3]float64{receiver.X, receiver.Y, receiver.Z}, windowStart, windowEnd, stepSeconds, elevationMaskDeg, nativeSystems)
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]VisibilitySeriesPoint, len(values))
	for i, value := range values {
		result[i] = VisibilitySeriesPoint{StepIndex: value.StepIndex, Visible: value.Visible}
	}
	return result, nil
}

// GeometryVisible lists satellites above an elevation mask at one epoch.
func (s *SP3) GeometryVisible(receiver ECEF, epochJ2000S, elevationMaskDeg float64, systems []GNSSSystem) ([]GeometryVisible, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	nativeSystems, err := geometrySystems(systems)
	if err != nil {
		return nil, err
	}
	values, err := s.handle.GeometryVisible([3]float64{receiver.X, receiver.Y, receiver.Z}, epochJ2000S, elevationMaskDeg, nativeSystems)
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]GeometryVisible, len(values))
	for i, value := range values {
		result[i] = GeometryVisible{Satellite: value.Satellite, ElevationDeg: value.ElevationDeg, AzimuthDeg: value.AzimuthDeg}
	}
	return result, nil
}

// ObservableState evaluates one precise SP3 ECEF state in metres and its
// optional satellite clock in seconds.
func (s *SP3) ObservableState(satelliteID string, epochJ2000S float64) ([3]float64, float64, bool, error) {
	if s == nil || s.handle == nil {
		return [3]float64{}, 0, false, ErrClosed
	}
	position, clock, present, err := s.handle.ObservableState(satelliteID, epochJ2000S)
	return position, clock, present, publicError(err)
}

// ObservableStates evaluates per-satellite SP3 states.
func (s *SP3) ObservableStates(satellites []string, epochsJ2000S []float64) ([]ObservableStateRow, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	values, err := s.handle.ObservableStates(satellites, epochsJ2000S)
	if err != nil {
		return nil, publicError(err)
	}
	return publicObservableStates(values), nil
}

// ObservableStatesShared evaluates all requested satellites at one epoch.
func (s *SP3) ObservableStatesShared(satellites []string, epochJ2000S float64) ([]ObservableStateRow, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	values, err := s.handle.ObservableStatesShared(satellites, epochJ2000S)
	if err != nil {
		return nil, publicError(err)
	}
	return publicObservableStates(values), nil
}

// PredictObservables computes one SP3 geometric/signal prediction.
func (s *SP3) PredictObservables(satelliteID string, receiver ECEF, epochJ2000S float64, options *ObservablesOptions) (PredictedObservables, error) {
	if s == nil || s.handle == nil {
		return PredictedObservables{}, ErrClosed
	}
	var nativeOptions *native.NativeObservablesOptions
	if options != nil {
		nativeOptions = &native.NativeObservablesOptions{CarrierHz: options.CarrierHz, LightTime: options.LightTime, Sagnac: options.Sagnac}
	}
	value, err := s.handle.PredictObservables(satelliteID, [3]float64{receiver.X, receiver.Y, receiver.Z}, epochJ2000S, nativeOptions)
	return fromNativePredicted(value), publicError(err)
}

// PredictObservablesBatch computes detached SP3 predictions for many requests.
func (s *SP3) PredictObservablesBatch(requests []PredictRequest, options *ObservablesOptions) ([]PredictedObservables, []bool, error) {
	if s == nil || s.handle == nil {
		return nil, nil, ErrClosed
	}
	nativeRequests := make([]native.NativePredictRequest, len(requests))
	for i, request := range requests {
		nativeRequests[i] = native.NativePredictRequest{SatelliteID: request.SatelliteID, ReceiverECEF: [3]float64{request.ReceiverECEF.X, request.ReceiverECEF.Y, request.ReceiverECEF.Z}, TRxJ2000S: request.TRxJ2000S}
	}
	var nativeOptions *native.NativeObservablesOptions
	if options != nil {
		nativeOptions = &native.NativeObservablesOptions{CarrierHz: options.CarrierHz, LightTime: options.LightTime, Sagnac: options.Sagnac}
	}
	values, accepted, err := s.handle.PredictObservablesBatch(nativeRequests, nativeOptions)
	if err != nil {
		return nil, nil, publicError(err)
	}
	result := make([]PredictedObservables, len(values))
	for i := range values {
		result[i] = fromNativePredicted(values[i])
	}
	return result, append([]bool(nil), accepted...), nil
}

// PredictRanges computes geometry-only range predictions through SP3.
func (s *SP3) PredictRanges(requests []PredictRequest, options *ObservablesOptions) ([]RangePrediction, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	nativeRequests := make([]native.NativePredictRequest, len(requests))
	for i, request := range requests {
		nativeRequests[i] = native.NativePredictRequest{SatelliteID: request.SatelliteID, ReceiverECEF: [3]float64{request.ReceiverECEF.X, request.ReceiverECEF.Y, request.ReceiverECEF.Z}, TRxJ2000S: request.TRxJ2000S}
	}
	var nativeOptions *native.NativeObservablesOptions
	if options != nil {
		nativeOptions = &native.NativeObservablesOptions{CarrierHz: options.CarrierHz, LightTime: options.LightTime, Sagnac: options.Sagnac}
	}
	values, err := s.handle.PredictRanges(nativeRequests, nativeOptions)
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]RangePrediction, len(values))
	for i, value := range values {
		result[i] = RangePrediction{GeometricRangeM: value.GeometricRangeM, HasSatelliteClock: value.HasSatelliteClock, SatelliteClockS: value.SatelliteClockS, TransmitTimeJ2000S: value.TransmitTimeJ2000S, SatellitePositionECEF: value.SatellitePositionECEF}
	}
	return result, nil
}

// StencilExtent returns the SP3 interpolation reach before and after a query, in seconds.
func (s *SP3) StencilExtent() (float64, float64, error) {
	if s == nil || s.handle == nil {
		return 0, 0, ErrClosed
	}
	before, after, err := s.handle.StencilExtent()
	return before, after, publicError(err)
}
