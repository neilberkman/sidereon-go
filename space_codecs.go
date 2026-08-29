package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// OEM owns a parsed CCSDS Orbit Ephemeris Message. The C handle retains all
// parsed metadata, segments, state records, and covariance blocks; this Go
// surface exposes the C-provided segment count and native serializers.
type OEM struct {
	_      noCopy
	native *native.OEM
}

// ParseOEMKVN parses a CCSDS OEM Keyword-Value Notation byte stream. Input
// bytes are copied and not retained.
func ParseOEMKVN(data []byte) (*OEM, error) {
	value, err := native.ParseOEM(data, false)
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errNilNativeHandle
	}
	return &OEM{native: value}, nil
}

// ParseOEMXML parses a CCSDS OEM XML byte stream. Input bytes are copied.
func ParseOEMXML(data []byte) (*OEM, error) {
	value, err := native.ParseOEM(data, true)
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errNilNativeHandle
	}
	return &OEM{native: value}, nil
}

// Close releases the OEM handle and is idempotent. Reads synchronize with it.
func (o *OEM) Close() error {
	if o == nil || o.native == nil {
		return nil
	}
	return publicError(o.native.Close())
}

// SegmentCount returns the number of OEM ephemeris segments.
func (o *OEM) SegmentCount() (int, error) {
	if o == nil || o.native == nil {
		return 0, ErrClosed
	}
	v, err := o.native.SegmentCount()
	return v, publicError(err)
}

// ToKVN serializes the OEM to an independent KVN byte slice.
func (o *OEM) ToKVN() ([]byte, error) {
	if o == nil || o.native == nil {
		return nil, ErrClosed
	}
	v, err := o.native.Text(false)
	return v, publicError(err)
}

// ToXML serializes the OEM to an independent XML byte slice.
func (o *OEM) ToXML() ([]byte, error) {
	if o == nil || o.native == nil {
		return nil, ErrClosed
	}
	v, err := o.native.Text(true)
	return v, publicError(err)
}

// ToKVN serializes the OMM to an independent KVN byte slice.
func (o *OMM) ToKVN() ([]byte, error) {
	if o == nil || o.handle == nil {
		return nil, ErrClosed
	}
	v, err := o.handle.Text(native.OMMFormatKVNValue)
	return v, publicError(err)
}

// ToXML serializes the OMM to an independent XML byte slice.
func (o *OMM) ToXML() ([]byte, error) {
	if o == nil || o.handle == nil {
		return nil, ErrClosed
	}
	v, err := o.handle.Text(native.OMMFormatXMLValue)
	return v, publicError(err)
}

// ToJSON serializes the OMM to an independent JSON byte slice.
func (o *OMM) ToJSON() ([]byte, error) {
	if o == nil || o.handle == nil {
		return nil, ErrClosed
	}
	v, err := o.handle.Text(native.OMMFormatJSONValue)
	return v, publicError(err)
}

// ConstellationRecord is the exact identity value returned by the OMM catalog
// adapter. It is distinct from orbital element data in an OMM message.
type ConstellationRecord struct {
	// System is the GNSS constellation.
	System GNSSSystem
	// PRN is the constellation-specific pseudorandom-noise number.
	PRN uint16
	// SVNPresent reports whether SVN is meaningful.
	SVNPresent bool
	// SVN is the optional space-vehicle number.
	SVN uint16
	// NORADID is the catalog identifier.
	NORADID uint32
	// FDMAChannelPresent reports whether FDMAChannel is meaningful.
	FDMAChannelPresent bool
	// FDMAChannel is the optional signed GLONASS FDMA channel.
	FDMAChannel int8
	// Active reports the catalog activity flag.
	Active bool
	// Usable reports whether the record is usable by the catalog.
	Usable bool
}

// SkippedOMM identifies an OMM catalog object omitted from typed records.
type SkippedOMM struct {
	// NORADID is the skipped object's catalog identifier.
	NORADID uint32
	// ObjectNamePresent reports whether ObjectName is present.
	ObjectNamePresent bool
	// ObjectName is an independent copy of the skipped object's name.
	ObjectName string
}

// OMMCatalog owns the typed and diagnostic results of a lenient OMM catalog
// build. It is non-copyable and synchronizes all reads with Close.
type OMMCatalog struct {
	_      noCopy
	native *native.OMMCatalog
}

// BuildOMMCatalogLenient parses a JSON array, retaining valid constellation
// records and exposing skipped and malformed counts. Input bytes are copied.
func BuildOMMCatalogLenient(system GNSSSystem, data []byte) (*OMMCatalog, error) {
	value, err := native.BuildOMMCatalogLenient(uint32(system), data)
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errNilNativeHandle
	}
	return &OMMCatalog{native: value}, nil
}

// Close releases the catalog and is idempotent. Reads synchronize with it.
func (c *OMMCatalog) Close() error {
	if c == nil || c.native == nil {
		return nil
	}
	return publicError(c.native.Close())
}

// RecordCount returns the number of retained typed records.
func (c *OMMCatalog) RecordCount() (int, error) {
	if c == nil || c.native == nil {
		return 0, ErrClosed
	}
	v, err := c.native.RecordCount()
	return v, publicError(err)
}

// SkippedCount returns the number of valid-but-unmatched catalog objects.
func (c *OMMCatalog) SkippedCount() (int, error) {
	if c == nil || c.native == nil {
		return 0, ErrClosed
	}
	v, err := c.native.SkippedCount()
	return v, publicError(err)
}

// MalformedCount returns the number of malformed array elements.
func (c *OMMCatalog) MalformedCount() (int, error) {
	if c == nil || c.native == nil {
		return 0, ErrClosed
	}
	v, err := c.native.MalformedCount()
	return v, publicError(err)
}

// Record returns an independent typed catalog record at index.
func (c *OMMCatalog) Record(index int) (ConstellationRecord, error) {
	if c == nil || c.native == nil {
		return ConstellationRecord{}, ErrClosed
	}
	v, err := c.native.Record(index)
	return ConstellationRecord{System: GNSSSystem(v.System), PRN: v.PRN, SVNPresent: v.SVNPresent, SVN: v.SVN, NORADID: v.NORADID, FDMAChannelPresent: v.FDMAChannelPresent, FDMAChannel: v.FDMAChannel, Active: v.Active, Usable: v.Usable}, publicError(err)
}

// Records returns independent copies of all typed records.
func (c *OMMCatalog) Records() ([]ConstellationRecord, error) {
	n, err := c.RecordCount()
	if err != nil {
		return nil, err
	}
	out := make([]ConstellationRecord, n)
	for i := range out {
		out[i], err = c.Record(i)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

// Skipped returns an independent skipped-object diagnostic at index.
func (c *OMMCatalog) Skipped(index int) (SkippedOMM, error) {
	if c == nil || c.native == nil {
		return SkippedOMM{}, ErrClosed
	}
	v, err := c.native.Skipped(index)
	return SkippedOMM{NORADID: v.NORADID, ObjectNamePresent: v.ObjectNamePresent, ObjectName: v.ObjectName}, publicError(err)
}

// SkippedEntries returns independent copies of all skipped-object diagnostics.
func (c *OMMCatalog) SkippedEntries() ([]SkippedOMM, error) {
	n, err := c.SkippedCount()
	if err != nil {
		return nil, err
	}
	out := make([]SkippedOMM, n)
	for i := range out {
		out[i], err = c.Skipped(i)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

// OPM owns a parsed CCSDS Orbit Parameter Message and delegates KVN/XML
// parsing and serialization to C.
type OPM struct {
	_      noCopy
	native *native.OPM
}

// ParseOPMKVN parses a CCSDS OPM KVN byte stream. Input bytes are copied.
func ParseOPMKVN(data []byte) (*OPM, error) {
	value, err := native.ParseOPM(data, false)
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errNilNativeHandle
	}
	return &OPM{native: value}, nil
}

// ParseOPMXML parses a CCSDS OPM XML byte stream. Input bytes are copied.
func ParseOPMXML(data []byte) (*OPM, error) {
	value, err := native.ParseOPM(data, true)
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errNilNativeHandle
	}
	return &OPM{native: value}, nil
}

// Close releases the OPM handle and is idempotent. Reads synchronize with it.
func (o *OPM) Close() error {
	if o == nil || o.native == nil {
		return nil
	}
	return publicError(o.native.Close())
}

// ToKVN serializes the OPM to an independent KVN byte slice.
func (o *OPM) ToKVN() ([]byte, error) {
	if o == nil || o.native == nil {
		return nil, ErrClosed
	}
	v, err := o.native.Text(false)
	return v, publicError(err)
}

// ToXML serializes the OPM to an independent XML byte slice.
func (o *OPM) ToXML() ([]byte, error) {
	if o == nil || o.native == nil {
		return nil, ErrClosed
	}
	v, err := o.native.Text(true)
	return v, publicError(err)
}

// TDMObservable identifies a TDM measurement keyword family.
type TDMObservable uint32

const (
	// TDMObservableRange is RANGE.
	TDMObservableRange TDMObservable = TDMObservable(native.TDMObservableRangeValue)
	// TDMObservableDopplerInstantaneous is DOPPLER_INSTANTANEOUS.
	TDMObservableDopplerInstantaneous TDMObservable = TDMObservable(native.TDMObservableDopplerInstantaneousValue)
	// TDMObservableDopplerIntegrated is DOPPLER_INTEGRATED.
	TDMObservableDopplerIntegrated TDMObservable = TDMObservable(native.TDMObservableDopplerIntegratedValue)
	// TDMObservableReceiveFrequency is RECEIVE_FREQ.
	TDMObservableReceiveFrequency TDMObservable = TDMObservable(native.TDMObservableReceiveFrequencyValue)
	// TDMObservableTransmitFrequency is TRANSMIT_FREQ.
	TDMObservableTransmitFrequency TDMObservable = TDMObservable(native.TDMObservableTransmitFrequencyValue)
	// TDMObservableTransmitFrequencyRate is TRANSMIT_FREQ_RATE.
	TDMObservableTransmitFrequencyRate TDMObservable = TDMObservable(native.TDMObservableTransmitRateValue)
	// TDMObservableAngle1 is ANGLE_1.
	TDMObservableAngle1 TDMObservable = TDMObservable(native.TDMObservableAngle1Value)
	// TDMObservableAngle2 is ANGLE_2.
	TDMObservableAngle2 TDMObservable = TDMObservable(native.TDMObservableAngle2Value)
	// TDMObservableOther preserves an unmodeled CCSDS observable.
	TDMObservableOther TDMObservable = TDMObservable(native.TDMObservableOtherValue)
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

// TDMDataRecord is one time-tagged measurement with both raw and numeric value.
type TDMDataRecord struct {
	// SegmentIndex is the zero-based parse-order segment.
	SegmentIndex int
	// Observable identifies the parsed measurement family.
	Observable TDMObservable
	// HasObservableParticipant reports whether ObservableParticipant is present.
	HasObservableParticipant bool
	// ObservableParticipant is the optional indexed participant suffix.
	ObservableParticipant uint8
	// Unit identifies the measurement unit.
	Unit TDMUnit
	// Keyword is the original data keyword.
	Keyword string
	// Epoch is the raw epoch token.
	Epoch string
	// ValueText is the exact decimal value token.
	ValueText string
	// Value is the parsed numeric value.
	Value float64
}

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

// TDM owns a parsed CCSDS Tracking Data Message. It is non-copyable; reads
// synchronize with Close, which is idempotent and returns cleanup errors.
type TDM struct {
	_      noCopy
	native *native.TDM
}

// ParseTDMKVN parses a CCSDS TDM KVN byte stream. Input bytes are copied.
func ParseTDMKVN(data []byte) (*TDM, error) {
	value, err := native.ParseTDM(data)
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errNilNativeHandle
	}
	return &TDM{native: value}, nil
}

// Close releases the TDM handle and is idempotent.
func (t *TDM) Close() error {
	if t == nil || t.native == nil {
		return nil
	}
	return publicError(t.native.Close())
}

// ToKVN serializes the TDM to an independent KVN byte slice.
func (t *TDM) ToKVN() ([]byte, error) {
	if t == nil || t.native == nil {
		return nil, ErrClosed
	}
	v, err := t.native.Text()
	return v, publicError(err)
}

// SegmentCount returns the number of parsed TDM segments.
func (t *TDM) SegmentCount() (int, error) {
	if t == nil || t.native == nil {
		return 0, ErrClosed
	}
	v, err := t.native.SegmentCount()
	return v, publicError(err)
}

// RecordCount returns the number of parsed TDM records.
func (t *TDM) RecordCount() (int, error) {
	if t == nil || t.native == nil {
		return 0, ErrClosed
	}
	v, err := t.native.RecordCount()
	return v, publicError(err)
}

// Segments returns independent segment metadata copies.
func (t *TDM) Segments() ([]TDMSegmentSummary, error) {
	if t == nil || t.native == nil {
		return nil, ErrClosed
	}
	values, err := t.native.Segments()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]TDMSegmentSummary, len(values))
	for i, v := range values {
		out[i] = TDMSegmentSummary{SegmentIndex: v.SegmentIndex, Mode: TDMStringField{Present: v.Mode.Present, Value: v.Mode.Value}, TimetagRef: TDMStringField{Present: v.TimetagRef.Present, Value: v.TimetagRef.Value}, TimeSystem: TDMStringField{Present: v.TimeSystem.Present, Value: v.TimeSystem.Value}, RangeUnit: TDMUnit(v.RangeUnit), ParticipantCount: v.ParticipantCount, PathCount: v.PathCount, RecordCount: v.RecordCount}
	}
	return out, nil
}

// Participants returns independent participant metadata copies.
func (t *TDM) Participants() ([]TDMParticipant, error) {
	if t == nil || t.native == nil {
		return nil, ErrClosed
	}
	values, err := t.native.Participants()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]TDMParticipant, len(values))
	for i, v := range values {
		out[i] = TDMParticipant{SegmentIndex: v.SegmentIndex, Index: v.Index, Name: v.Name}
	}
	return out, nil
}

// Paths returns independent path metadata and participant-index copies.
func (t *TDM) Paths() ([]TDMPath, error) {
	if t == nil || t.native == nil {
		return nil, ErrClosed
	}
	values, err := t.native.Paths()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]TDMPath, len(values))
	for i, v := range values {
		out[i] = TDMPath{SegmentIndex: v.SegmentIndex, Key: v.Key, HasIndex: v.HasIndex, Index: v.Index, Participants: append([]uint8(nil), v.Participants...)}
	}
	return out, nil
}

// Records returns independent raw and numeric record copies.
func (t *TDM) Records() ([]TDMDataRecord, error) {
	if t == nil || t.native == nil {
		return nil, ErrClosed
	}
	values, err := t.native.Records()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]TDMDataRecord, len(values))
	for i, v := range values {
		out[i] = TDMDataRecord{SegmentIndex: v.SegmentIndex, Observable: TDMObservable(v.Observable), HasObservableParticipant: v.HasObservableParticipant, ObservableParticipant: v.ObservableParticipant, Unit: TDMUnit(v.Unit), Keyword: v.Keyword, Epoch: v.Epoch, ValueText: v.ValueText, Value: v.Value}
	}
	return out, nil
}
