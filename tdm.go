package sidereon

import "sidereon.dev/go/v3/internal/native"

// TDMLeniency controls whether one class of non-conforming TDM input is refused or retained.
type TDMLeniency uint32

const (
	// TDMLeniencyStrict refuses the departure.
	TDMLeniencyStrict TDMLeniency = TDMLeniency(native.TDMLeniencyStrict)
	// TDMLeniencyForgive accepts and reports the departure.
	TDMLeniencyForgive TDMLeniency = TDMLeniency(native.TDMLeniencyForgive)
)

// TDMObservable identifies a TDM measurement keyword family.
type TDMObservable uint32

const (
	// TDMObservableRange is RANGE.
	TDMObservableRange TDMObservable = TDMObservable(native.TDMObservableRange)
	// TDMObservableDopplerInstantaneous is DOPPLER_INSTANTANEOUS.
	TDMObservableDopplerInstantaneous TDMObservable = TDMObservable(native.TDMObservableDopplerInstantaneous)
	// TDMObservableDopplerIntegrated is DOPPLER_INTEGRATED.
	TDMObservableDopplerIntegrated TDMObservable = TDMObservable(native.TDMObservableDopplerIntegrated)
	// TDMObservableReceiveFrequency is RECEIVE_FREQ.
	TDMObservableReceiveFrequency TDMObservable = TDMObservable(native.TDMObservableReceiveFreq)
	// TDMObservableTransmitFrequency is TRANSMIT_FREQ.
	TDMObservableTransmitFrequency TDMObservable = TDMObservable(native.TDMObservableTransmitFreq)
	// TDMObservableTransmitFrequencyRate is TRANSMIT_FREQ_RATE.
	TDMObservableTransmitFrequencyRate TDMObservable = TDMObservable(native.TDMObservableTransmitFreqRate)
	// TDMObservableAngle1 is ANGLE_1.
	TDMObservableAngle1 TDMObservable = TDMObservable(native.TDMObservableAngle1)
	// TDMObservableAngle2 is ANGLE_2.
	TDMObservableAngle2 TDMObservable = TDMObservable(native.TDMObservableAngle2)
	// TDMObservableOther preserves an unmodeled CCSDS observable.
	TDMObservableOther TDMObservable = TDMObservable(native.TDMObservableOther)
)

// TDMUnit identifies the unit of a TDM record value.
type TDMUnit uint32

const (
	// TDMUnitKilometers is km.
	TDMUnitKilometers TDMUnit = TDMUnit(native.TDMUnitKilometersValue)
	// TDMUnitSeconds is s.
	TDMUnitSeconds TDMUnit = TDMUnit(native.TDMUnitSecondsValue)
	// TDMUnitRangeUnits is a CCSDS range unit.
	TDMUnitRangeUnits TDMUnit = TDMUnit(native.TDMUnitRangeUnitsValue)
	// TDMUnitKilometersPerSecond is km/s.
	TDMUnitKilometersPerSecond TDMUnit = TDMUnit(native.TDMUnitKilometersPerSecondValue)
	// TDMUnitHertz is Hz.
	TDMUnitHertz TDMUnit = TDMUnit(native.TDMUnitHertzValue)
	// TDMUnitHertzPerSecond is Hz/s.
	TDMUnitHertzPerSecond TDMUnit = TDMUnit(native.TDMUnitHertzPerSecondValue)
	// TDMUnitDegrees is degrees.
	TDMUnitDegrees TDMUnit = TDMUnit(native.TDMUnitDegreesValue)
	// TDMUnitDecibelWatts is dBW.
	TDMUnitDecibelWatts TDMUnit = TDMUnit(native.TDMUnitDecibelWattsValue)
	// TDMUnitDecibelHertz is dBHz.
	TDMUnitDecibelHertz TDMUnit = TDMUnit(native.TDMUnitDecibelHertzValue)
	// TDMUnitSquareMeters is m^2.
	TDMUnitSquareMeters TDMUnit = TDMUnit(native.TDMUnitSquareMetersValue)
	// TDMUnitMeters is m.
	TDMUnitMeters TDMUnit = TDMUnit(native.TDMUnitMetersValue)
	// TDMUnitSecondsPerSecond is s/s.
	TDMUnitSecondsPerSecond TDMUnit = TDMUnit(native.TDMUnitSecondsPerSecondValue)
	// TDMUnitPercent is percent.
	TDMUnitPercent TDMUnit = TDMUnit(native.TDMUnitPercentValue)
	// TDMUnitKelvin is K.
	TDMUnitKelvin TDMUnit = TDMUnit(native.TDMUnitKelvinValue)
	// TDMUnitHectopascals is hPa.
	TDMUnitHectopascals TDMUnit = TDMUnit(native.TDMUnitHectopascalsValue)
	// TDMUnitTotalElectronContentUnits is TECU.
	TDMUnitTotalElectronContentUnits TDMUnit = TDMUnit(native.TDMUnitTotalElectronContentUnitsValue)
	// TDMUnitDimensionless is a dimensionless quantity.
	TDMUnitDimensionless TDMUnit = TDMUnit(native.TDMUnitDimensionlessValue)
	// TDMUnitUnknown preserves an unmodeled unit.
	TDMUnitUnknown TDMUnit = TDMUnit(native.TDMUnitUnknownValue)
)

// TDMSection identifies the header, metadata, or data section containing a returned row.
type TDMSection uint32

const (
	// TDMSectionHeader identifies a message header row.
	TDMSectionHeader TDMSection = 0
	// TDMSectionMetadata identifies a segment metadata row.
	TDMSectionMetadata TDMSection = 1
	// TDMSectionData identifies a segment data row.
	TDMSectionData TDMSection = 2
)

// TDMPolicy controls policy-aware KVN parsing.
type TDMPolicy struct {
	// NonPrintable controls non-printable character handling.
	NonPrintable TDMLeniency
	// MissingKeywords controls required-keyword handling.
	MissingKeywords TDMLeniency
	// LongLines controls overlong-line handling.
	LongLines TDMLeniency
	// EmptyDataSections controls empty data-section handling.
	EmptyDataSections TDMLeniency
	// RecordOrder controls out-of-order record handling.
	RecordOrder TDMLeniency
	// DuplicateRecords controls duplicate record handling.
	DuplicateRecords TDMLeniency
	// KeywordOrder controls out-of-order keyword handling.
	KeywordOrder TDMLeniency
	// FinalTerminator controls handling of a missing final line terminator.
	FinalTerminator TDMLeniency
}

// StrictTDMPolicy returns the strict TDM read policy.
func StrictTDMPolicy() TDMPolicy {
	return publicTDMPolicy(native.TDMPolicyStrict())
}

// DefaultTDMPolicy returns the default TDM read policy.
func DefaultTDMPolicy() TDMPolicy {
	return publicTDMPolicy(native.TDMPolicyDefault())
}

// LenientTDMPolicy returns a policy that forgives every supported departure.
func LenientTDMPolicy() TDMPolicy {
	return publicTDMPolicy(native.TDMPolicyLenient())
}

func publicTDMPolicy(policy native.TDMPolicy) TDMPolicy {
	return TDMPolicy{
		NonPrintable:      TDMLeniency(policy.NonPrintable),
		MissingKeywords:   TDMLeniency(policy.MissingKeywords),
		LongLines:         TDMLeniency(policy.LongLines),
		EmptyDataSections: TDMLeniency(policy.EmptyDataSections),
		RecordOrder:       TDMLeniency(policy.RecordOrder),
		DuplicateRecords:  TDMLeniency(policy.DuplicateRecords),
		KeywordOrder:      TDMLeniency(policy.KeywordOrder),
		FinalTerminator:   TDMLeniency(policy.FinalTerminator),
	}
}

func nativeTDMPolicy(policy TDMPolicy) native.TDMPolicy {
	return native.TDMPolicy{
		NonPrintable:      uint32(policy.NonPrintable),
		MissingKeywords:   uint32(policy.MissingKeywords),
		LongLines:         uint32(policy.LongLines),
		EmptyDataSections: uint32(policy.EmptyDataSections),
		RecordOrder:       uint32(policy.RecordOrder),
		DuplicateRecords:  uint32(policy.DuplicateRecords),
		KeywordOrder:      uint32(policy.KeywordOrder),
		FinalTerminator:   uint32(policy.FinalTerminator),
	}
}

// TDMWritePolicy controls policy-aware validation and KVN encoding.
type TDMWritePolicy struct {
	// NonPrintable controls non-printable character handling.
	NonPrintable TDMLeniency
	// MissingKeywords controls required-keyword handling.
	MissingKeywords TDMLeniency
	// LongLines controls overlong-line handling.
	LongLines TDMLeniency
	// EmptyDataSections controls empty data-section handling.
	EmptyDataSections TDMLeniency
	// RecordOrder controls out-of-order record handling.
	RecordOrder TDMLeniency
	// DuplicateRecords controls duplicate record handling.
	DuplicateRecords TDMLeniency
	// KeywordOrder controls out-of-order keyword handling.
	KeywordOrder TDMLeniency
	// FinalTerminator controls final line-terminator handling.
	FinalTerminator TDMLeniency
	// RepeatedKeywords controls repeated-keyword handling.
	RepeatedKeywords TDMLeniency
}

// StrictTDMWritePolicy returns the strict TDM write policy.
func StrictTDMWritePolicy() TDMWritePolicy {
	return publicTDMWritePolicy(native.TDMWritePolicyStrict())
}

// DefaultTDMWritePolicy returns the default TDM write policy.
func DefaultTDMWritePolicy() TDMWritePolicy {
	return publicTDMWritePolicy(native.TDMWritePolicyDefault())
}

// LenientTDMWritePolicy returns a policy that forgives every supported departure.
func LenientTDMWritePolicy() TDMWritePolicy {
	return publicTDMWritePolicy(native.TDMWritePolicyLenient())
}

func publicTDMWritePolicy(policy native.TDMWritePolicy) TDMWritePolicy {
	return TDMWritePolicy{
		NonPrintable:      TDMLeniency(policy.NonPrintable),
		MissingKeywords:   TDMLeniency(policy.MissingKeywords),
		LongLines:         TDMLeniency(policy.LongLines),
		EmptyDataSections: TDMLeniency(policy.EmptyDataSections),
		RecordOrder:       TDMLeniency(policy.RecordOrder),
		DuplicateRecords:  TDMLeniency(policy.DuplicateRecords),
		KeywordOrder:      TDMLeniency(policy.KeywordOrder),
		FinalTerminator:   TDMLeniency(policy.FinalTerminator),
		RepeatedKeywords:  TDMLeniency(policy.RepeatedKeywords),
	}
}

func nativeTDMWritePolicy(policy TDMWritePolicy) native.TDMWritePolicy {
	return native.TDMWritePolicy{
		NonPrintable:      uint32(policy.NonPrintable),
		MissingKeywords:   uint32(policy.MissingKeywords),
		LongLines:         uint32(policy.LongLines),
		EmptyDataSections: uint32(policy.EmptyDataSections),
		RecordOrder:       uint32(policy.RecordOrder),
		DuplicateRecords:  uint32(policy.DuplicateRecords),
		KeywordOrder:      uint32(policy.KeywordOrder),
		FinalTerminator:   uint32(policy.FinalTerminator),
		RepeatedKeywords:  uint32(policy.RepeatedKeywords),
	}
}

// AsRead converts a write policy to its matching read policy.
func (policy TDMWritePolicy) AsRead() (TDMPolicy, error) {
	value, err := nativeTDMWritePolicy(policy).AsRead()
	return publicTDMPolicy(value), publicError(err)
}

// TDMComment is a lossless positioned comment returned by a TDM read operation.
type TDMComment struct {
	// SegmentIndex is the zero-based segment index; header comments use zero.
	SegmentIndex int
	// Section identifies the section containing the comment.
	Section TDMSection
	// BeforeRecord is the zero-based field or record index the comment precedes.
	BeforeRecord int
	// Text is the comment text without the COMMENT keyword.
	Text string
}

// TDMCommentInput describes a positioned comment accepted by a TDM edit operation.
type TDMCommentInput struct {
	// BeforeRecord is the zero-based field or record index the comment precedes.
	BeforeRecord int
	// Text is the comment text without the COMMENT keyword.
	Text string
}

// TDMField is a lossless ordered raw field returned by a TDM read operation.
type TDMField struct {
	// SegmentIndex is the zero-based segment index; header fields use zero.
	SegmentIndex int
	// Section identifies the section containing the field.
	Section TDMSection
	// Key is the original KVN keyword.
	Key string
	// Value is the original KVN value text.
	Value string
}

// TDMFieldInput describes an ordered raw field accepted by a TDM edit operation.
type TDMFieldInput struct {
	// Key is the KVN keyword.
	Key string
	// Value is the KVN value text.
	Value string
}

// TDMHeader contains scalar header values and raw header rows returned by Header.
type TDMHeader struct {
	// Version is the CCSDS TDM version.
	Version string
	// CreationDate is the optional raw CREATION_DATE value.
	CreationDate TDMStringField
	// Originator is the optional raw ORIGINATOR value.
	Originator TDMStringField
	// MessageID is the optional raw MESSAGE_ID value.
	MessageID TDMStringField
	// Comments contains lossless positioned header comments.
	Comments []TDMComment
	// Fields contains lossless raw header fields.
	Fields []TDMField
	// SegmentCount is the number of metadata/data segments.
	SegmentCount int
}

// TDMHeaderValues contains only the scalar header values accepted by SetHeader.
type TDMHeaderValues struct {
	// Version is the CCSDS TDM version.
	Version string
	// CreationDate is the optional raw CREATION_DATE value.
	CreationDate TDMStringField
	// Originator is the optional raw ORIGINATOR value.
	Originator TDMStringField
	// MessageID is the optional raw MESSAGE_ID value.
	MessageID TDMStringField
}

// TDMParticipant is one participant declaration copied from a TDM segment.
type TDMParticipant struct {
	// SegmentIndex is the zero-based parse-order segment.
	SegmentIndex int
	// Index is the PARTICIPANT_n suffix.
	Index uint8
	// Name is an independent participant name copy.
	Name string
}

// TDMPath preserves one segment PATH declaration and its participant order.
type TDMPath struct {
	// SegmentIndex is the zero-based parse-order segment.
	SegmentIndex int
	// Key is the original PATH keyword.
	Key string
	// HasIndex reports whether Index was present in PATH_n.
	HasIndex bool
	// Index is the optional PATH_n suffix.
	Index uint8
	// Participants contains copied participant indices in path order.
	Participants []uint8
}

// TDMRecord is one lossless time-tagged measurement returned by Records.
type TDMRecord struct {
	// SegmentIndex is the zero-based parse-order segment.
	SegmentIndex int
	// Observable identifies the parsed measurement family.
	Observable TDMObservable
	// HasObservableParticipant reports whether ObservableParticipant is present.
	HasObservableParticipant bool
	// ObservableParticipant is the optional indexed participant suffix.
	ObservableParticipant uint8
	// ObservableName preserves the name of an unmodeled observable.
	ObservableName string
	// Unit identifies the measurement unit.
	Unit TDMUnit
	// UnitLabel preserves the canonical or unmodeled unit label.
	UnitLabel string
	// Keyword is the original data keyword.
	Keyword string
	// Epoch is the raw epoch token.
	Epoch string
	// ValueText is the exact decimal value token.
	ValueText string
	// Value is the parsed numeric value.
	Value float64
}

// TDMRecordInput describes one lossless time-tagged measurement accepted by an edit operation.
type TDMRecordInput struct {
	// Observable identifies the measurement family.
	Observable TDMObservable
	// HasObservableParticipant reports whether ObservableParticipant is present.
	HasObservableParticipant bool
	// ObservableParticipant is the optional indexed participant suffix.
	ObservableParticipant uint8
	// ObservableName supplies the name of an unmodeled observable.
	ObservableName string
	// Unit identifies the measurement unit.
	Unit TDMUnit
	// UnitLabel supplies the canonical or unmodeled unit label.
	UnitLabel string
	// Keyword is the original data keyword.
	Keyword string
	// Epoch is the raw epoch token.
	Epoch string
	// ValueText is the exact decimal value token.
	ValueText string
	// Value is the parsed numeric value.
	Value float64
}

// TDMDataRecord is retained as an alias for TDMRecord.
type TDMDataRecord = TDMRecord

// TDMStringField preserves presence separately from a possibly empty value.
type TDMStringField struct {
	// Present reports whether Value was declared.
	Present bool
	// Value is an independent copy of the text field.
	Value string
}

// TDMSegmentSummary contains metadata and counts for one TDM segment.
type TDMSegmentSummary struct {
	// SegmentIndex is the zero-based parse-order segment.
	SegmentIndex int
	// Mode is the optional CCSDS mode field.
	Mode TDMStringField
	// TimetagRef is the optional time-tag reference field.
	TimetagRef TDMStringField
	// TimeSystem is the optional time scale field.
	TimeSystem TDMStringField
	// RangeUnit is the declared range unit.
	RangeUnit TDMUnit
	// ParticipantCount is the segment participant count.
	ParticipantCount int
	// PathCount is the segment path count.
	PathCount int
	// RecordCount is the segment record count.
	RecordCount int
}

// TDMWarning describes one forgiven parse departure.
type TDMWarning struct {
	// Kind is the stable warning discriminator.
	Kind string
	// Fields retains every typed JSON detail field.
	Fields map[string]EngineJSONValue
	// Message is the native diagnostic text.
	Message string
}

// TDMDeparture describes one forgiven validation or write departure.
type TDMDeparture struct {
	// Kind is the stable departure discriminator.
	Kind string
	// Fields retains every typed JSON detail field.
	Fields map[string]EngineJSONValue
	// Message is the native diagnostic text.
	Message string
}

// TDM owns a parsed or constructed CCSDS Tracking Data Message. It is non-copyable; reads synchronize with Close.
type TDM struct {
	_      noCopy
	handle *native.TDMFull
}

func publicTDM(value *native.TDMFull) *TDM {
	if value == nil {
		return nil
	}
	return &TDM{handle: value}
}

// NewTDM constructs an empty mutable TDM with the supplied version.
func NewTDM(version string) (*TDM, error) {
	value, err := native.CreateTDMFull(version)
	if err != nil {
		return nil, publicError(err)
	}
	return publicTDM(value), nil
}

// ParseTDMKVN parses strict CCSDS TDM KVN text. Input bytes are copied.
func ParseTDMKVN(data []byte) (*TDM, error) {
	value, err := native.ParseTDMFull(append([]byte(nil), data...))
	if err != nil {
		return nil, publicError(err)
	}
	return publicTDM(value), nil
}

// ParseTDMKVNWithPolicy parses CCSDS TDM KVN text and returns forgiven departures as warnings.
func ParseTDMKVNWithPolicy(data []byte, policy TDMPolicy) (*TDM, []TDMWarning, error) {
	value, warnings, err := native.ParseTDMFullWithPolicy(append([]byte(nil), data...), nativeTDMPolicy(policy))
	if err != nil {
		return nil, nil, publicError(err)
	}
	return publicTDM(value), publicWarnings(warnings), nil
}

// Close releases the TDM handle and is idempotent.
func (tdm *TDM) Close() error {
	if tdm == nil || tdm.handle == nil {
		return nil
	}
	return publicError(tdm.handle.Close())
}

func (tdm *TDM) valid() error {
	if tdm == nil || tdm.handle == nil {
		return ErrClosed
	}
	return nil
}

// ToKVN serializes the TDM under the strict policy to an independent KVN byte slice.
func (tdm *TDM) ToKVN() ([]byte, error) {
	if err := tdm.valid(); err != nil {
		return nil, err
	}
	value, err := tdm.handle.KVN()
	return append([]byte(nil), value...), publicError(err)
}

// ToKVNWithPolicy serializes the TDM under policy and returns its departures.
func (tdm *TDM) ToKVNWithPolicy(policy TDMWritePolicy) ([]byte, []TDMDeparture, error) {
	if err := tdm.valid(); err != nil {
		return nil, nil, err
	}
	value, departures, err := tdm.handle.KVNWithPolicy(nativeTDMWritePolicy(policy))
	return append([]byte(nil), value...), publicDepartures(departures), publicError(err)
}

// Header returns the complete TDM header as independent scalar and raw-row copies.
func (tdm *TDM) Header() (TDMHeader, error) {
	if err := tdm.valid(); err != nil {
		return TDMHeader{}, err
	}
	value, err := tdm.handle.Header()
	return publicHeader(value), publicError(err)
}

// SetHeader atomically replaces the scalar header values while retaining raw rows and segments.
func (tdm *TDM) SetHeader(values TDMHeaderValues) error {
	if err := tdm.valid(); err != nil {
		return err
	}
	return publicError(tdm.handle.SetHeader(nativeHeaderValues(values)))
}

// ReplaceHeaderRaw replaces header comments and raw header fields while retaining scalar values and segments.
func (tdm *TDM) ReplaceHeaderRaw(comments []TDMCommentInput, fields []TDMFieldInput) error {
	if err := tdm.valid(); err != nil {
		return err
	}
	return publicError(tdm.handle.ReplaceHeaderRaw(nativeCommentInputs(comments), nativeFieldInputs(fields)))
}

// Comments returns independent copies of all header, metadata, and data comments.
func (tdm *TDM) Comments() ([]TDMComment, error) {
	if err := tdm.valid(); err != nil {
		return nil, err
	}
	values, err := tdm.handle.Comments()
	return publicComments(values), publicError(err)
}

// Fields returns independent copies of all header and metadata raw fields.
func (tdm *TDM) Fields() ([]TDMField, error) {
	if err := tdm.valid(); err != nil {
		return nil, err
	}
	values, err := tdm.handle.Fields()
	return publicFields(values), publicError(err)
}

// Participants returns independent participant metadata copies.
func (tdm *TDM) Participants() ([]TDMParticipant, error) {
	if err := tdm.valid(); err != nil {
		return nil, err
	}
	values, err := tdm.handle.Participants()
	result := make([]TDMParticipant, len(values))
	for index, value := range values {
		result[index] = TDMParticipant{
			SegmentIndex: value.SegmentIndex,
			Index:        value.Index,
			Name:         value.Name,
		}
	}
	return result, publicError(err)
}

// Paths returns independent path metadata and participant-index copies.
func (tdm *TDM) Paths() ([]TDMPath, error) {
	if err := tdm.valid(); err != nil {
		return nil, err
	}
	values, err := tdm.handle.Paths()
	result := make([]TDMPath, len(values))
	for index, value := range values {
		result[index] = TDMPath{
			SegmentIndex: value.SegmentIndex,
			Key:          value.Key,
			HasIndex:     value.HasIndex,
			Index:        value.Index,
			Participants: append([]uint8(nil), value.Participants...),
		}
	}
	return result, publicError(err)
}

// Segments returns independent segment metadata summaries.
func (tdm *TDM) Segments() ([]TDMSegmentSummary, error) {
	if err := tdm.valid(); err != nil {
		return nil, err
	}
	values, err := tdm.handle.Segments()
	result := make([]TDMSegmentSummary, len(values))
	for index, value := range values {
		result[index] = TDMSegmentSummary{
			SegmentIndex:     value.SegmentIndex,
			Mode:             publicTDMStringField(value.Mode),
			TimetagRef:       publicTDMStringField(value.TimetagRef),
			TimeSystem:       publicTDMStringField(value.TimeSystem),
			RangeUnit:        TDMUnit(value.RangeUnit),
			ParticipantCount: value.ParticipantCount,
			PathCount:        value.PathCount,
			RecordCount:      value.RecordCount,
		}
	}
	return result, publicError(err)
}

func publicTDMStringField(value *string) TDMStringField {
	if value == nil {
		return TDMStringField{}
	}
	return TDMStringField{Present: true, Value: *value}
}

// SegmentCount returns the number of parsed TDM segments.
func (tdm *TDM) SegmentCount() (int, error) {
	if err := tdm.valid(); err != nil {
		return 0, err
	}
	value, err := tdm.handle.SegmentCount()
	return value, publicError(err)
}

// Records returns independent raw and numeric record copies.
func (tdm *TDM) Records() ([]TDMRecord, error) {
	if err := tdm.valid(); err != nil {
		return nil, err
	}
	values, err := tdm.handle.Records()
	return publicRecords(values), publicError(err)
}

// RecordCount returns the number of parsed TDM records.
func (tdm *TDM) RecordCount() (int, error) {
	if err := tdm.valid(); err != nil {
		return 0, err
	}
	value, err := tdm.handle.RecordCount()
	return value, publicError(err)
}

// AppendSegment appends a validated segment and returns policy departures.
func (tdm *TDM) AppendSegment(fields []TDMFieldInput, metadataComments, dataComments []TDMCommentInput, records []TDMRecordInput, policy TDMWritePolicy) ([]TDMDeparture, error) {
	if err := tdm.valid(); err != nil {
		return nil, err
	}
	values, err := tdm.handle.AppendSegment(
		nativeFieldInputs(fields),
		nativeCommentInputs(metadataComments),
		nativeCommentInputs(dataComments),
		nativeRecordInputs(records),
		nativeTDMWritePolicy(policy),
	)
	return publicDepartures(values), publicError(err)
}

// ReplaceMetadata atomically replaces one segment's raw metadata and returns policy departures.
func (tdm *TDM) ReplaceMetadata(segmentIndex int, fields []TDMFieldInput, comments []TDMCommentInput, policy TDMWritePolicy) ([]TDMDeparture, error) {
	if err := tdm.valid(); err != nil {
		return nil, err
	}
	values, err := tdm.handle.ReplaceMetadata(
		segmentIndex,
		nativeFieldInputs(fields),
		nativeCommentInputs(comments),
		nativeTDMWritePolicy(policy),
	)
	return publicDepartures(values), publicError(err)
}

// ReplaceData atomically replaces one segment's data comments and records.
func (tdm *TDM) ReplaceData(segmentIndex int, comments []TDMCommentInput, records []TDMRecordInput) error {
	if err := tdm.valid(); err != nil {
		return err
	}
	return publicError(tdm.handle.ReplaceData(segmentIndex, nativeCommentInputs(comments), nativeRecordInputs(records)))
}

// InsertRecord inserts a lossless record at recordIndex in one segment.
func (tdm *TDM) InsertRecord(segmentIndex, recordIndex int, record TDMRecordInput) error {
	if err := tdm.valid(); err != nil {
		return err
	}
	return publicError(tdm.handle.InsertRecord(segmentIndex, recordIndex, nativeRecordInput(record)))
}

// RemoveRecord removes the record at recordIndex from one segment.
func (tdm *TDM) RemoveRecord(segmentIndex, recordIndex int) error {
	if err := tdm.valid(); err != nil {
		return err
	}
	return publicError(tdm.handle.RemoveRecord(segmentIndex, recordIndex))
}

// RemoveSegment removes the complete segment at segmentIndex.
func (tdm *TDM) RemoveSegment(segmentIndex int) error {
	if err := tdm.valid(); err != nil {
		return err
	}
	return publicError(tdm.handle.RemoveSegment(segmentIndex))
}

func publicComments(values []native.TDMComment) []TDMComment {
	result := make([]TDMComment, len(values))
	for index, value := range values {
		result[index] = TDMComment{
			SegmentIndex: value.SegmentIndex,
			Section:      TDMSection(value.Section),
			BeforeRecord: value.BeforeRecord,
			Text:         value.Text,
		}
	}
	return result
}

func nativeCommentInputs(values []TDMCommentInput) []native.TDMComment {
	result := make([]native.TDMComment, len(values))
	for index, value := range values {
		result[index] = native.TDMComment{
			BeforeRecord: value.BeforeRecord,
			Text:         value.Text,
		}
	}
	return result
}

func publicFields(values []native.TDMField) []TDMField {
	result := make([]TDMField, len(values))
	for index, value := range values {
		result[index] = TDMField{
			SegmentIndex: value.SegmentIndex,
			Section:      TDMSection(value.Section),
			Key:          value.Key,
			Value:        value.Value,
		}
	}
	return result
}

func nativeFieldInputs(values []TDMFieldInput) []native.TDMField {
	result := make([]native.TDMField, len(values))
	for index, value := range values {
		result[index] = native.TDMField{Key: value.Key, Value: value.Value}
	}
	return result
}

func nativeRecordInput(value TDMRecordInput) native.TDMRecord {
	return native.TDMRecord{
		Observable:               uint32(value.Observable),
		HasObservableParticipant: value.HasObservableParticipant,
		ObservableParticipant:    value.ObservableParticipant,
		ObservableName:           value.ObservableName,
		Unit:                     uint32(value.Unit),
		UnitLabel:                value.UnitLabel,
		Keyword:                  value.Keyword,
		Epoch:                    value.Epoch,
		ValueText:                value.ValueText,
		Value:                    value.Value,
	}
}

func nativeRecordInputs(values []TDMRecordInput) []native.TDMRecord {
	result := make([]native.TDMRecord, len(values))
	for index, value := range values {
		result[index] = nativeRecordInput(value)
	}
	return result
}

func publicRecords(values []native.TDMRecord) []TDMRecord {
	result := make([]TDMRecord, len(values))
	for index, value := range values {
		result[index] = TDMRecord{
			SegmentIndex:             value.SegmentIndex,
			Observable:               TDMObservable(value.Observable),
			HasObservableParticipant: value.HasObservableParticipant,
			ObservableParticipant:    value.ObservableParticipant,
			ObservableName:           value.ObservableName,
			Unit:                     TDMUnit(value.Unit),
			UnitLabel:                value.UnitLabel,
			Keyword:                  value.Keyword,
			Epoch:                    value.Epoch,
			ValueText:                value.ValueText,
			Value:                    value.Value,
		}
	}
	return result
}

func nativeHeaderValues(values TDMHeaderValues) native.TDMHeader {
	return native.TDMHeader{
		Version:      values.Version,
		CreationDate: nativeOptionalTDMString(values.CreationDate),
		Originator:   nativeOptionalTDMString(values.Originator),
		MessageID:    nativeOptionalTDMString(values.MessageID),
	}
}

func nativeOptionalTDMString(value TDMStringField) *string {
	if !value.Present {
		return nil
	}
	result := value.Value
	return &result
}

func publicHeader(value native.TDMHeader) TDMHeader {
	return TDMHeader{
		Version:      value.Version,
		CreationDate: publicTDMStringField(value.CreationDate),
		Originator:   publicTDMStringField(value.Originator),
		MessageID:    publicTDMStringField(value.MessageID),
		Comments:     publicComments(value.Comments),
		Fields:       publicFields(value.Fields),
		SegmentCount: value.SegmentCount,
	}
}

func publicWarnings(values []native.TDMNotice) []TDMWarning {
	result := make([]TDMWarning, len(values))
	for index, value := range values {
		result[index] = TDMWarning{
			Kind:    value.Kind,
			Fields:  publicEngineJSONFields(value.Fields),
			Message: value.Message,
		}
	}
	return result
}

func publicDepartures(values []native.TDMNotice) []TDMDeparture {
	result := make([]TDMDeparture, len(values))
	for index, value := range values {
		result[index] = TDMDeparture{
			Kind:    value.Kind,
			Fields:  publicEngineJSONFields(value.Fields),
			Message: value.Message,
		}
	}
	return result
}
