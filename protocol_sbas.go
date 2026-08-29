package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// SBASWireForm identifies the SBAS message wire encoding.
type SBASWireForm uint32

const (
	// SBASWireFramed250 selects a 250-bit framed SBAS message.
	SBASWireFramed250 SBASWireForm = SBASWireForm(native.SBASWireFramed250Value)
	// SBASWireBody226 selects a 226-bit SBAS message body.
	SBASWireBody226 SBASWireForm = SBASWireForm(native.SBASWireBody226Value)
)

// SBASMessageKind classifies a decoded SBAS message payload.
type SBASMessageKind uint32

const (
	// SBASMessageDoNotUse classifies an SBAS message carrying do not use.
	SBASMessageDoNotUse SBASMessageKind = SBASMessageKind(native.SBASMessageDoNotUseValue)
	// SBASMessagePRNMask classifies an SBAS message carrying a PRN mask.
	SBASMessagePRNMask SBASMessageKind = SBASMessageKind(native.SBASMessagePRNMaskValue)
	// SBASMessageFastCorrections classifies an SBAS message carrying fast corrections.
	SBASMessageFastCorrections SBASMessageKind = SBASMessageKind(native.SBASMessageFastCorrectionsValue)
	// SBASMessageIntegrity classifies an SBAS message carrying integrity.
	SBASMessageIntegrity SBASMessageKind = SBASMessageKind(native.SBASMessageIntegrityValue)
	// SBASMessageFastDegradation classifies an SBAS message carrying fast degradation.
	SBASMessageFastDegradation SBASMessageKind = SBASMessageKind(native.SBASMessageFastDegradationValue)
	// SBASMessageGeoNav classifies an SBAS message carrying GEO navigation data.
	SBASMessageGeoNav SBASMessageKind = SBASMessageKind(native.SBASMessageGeoNavValue)
	// SBASMessageNetworkTime classifies an SBAS message carrying network time.
	SBASMessageNetworkTime SBASMessageKind = SBASMessageKind(native.SBASMessageNetworkTimeValue)
	// SBASMessageGeoAlmanac classifies an SBAS message carrying GEO almanac data.
	SBASMessageGeoAlmanac SBASMessageKind = SBASMessageKind(native.SBASMessageGeoAlmanacValue)
	// SBASMessageIGPMask classifies an SBAS message carrying an IGP mask.
	SBASMessageIGPMask SBASMessageKind = SBASMessageKind(native.SBASMessageIGPMaskValue)
	// SBASMessageMixedCorrections classifies an SBAS message carrying mixed corrections.
	SBASMessageMixedCorrections SBASMessageKind = SBASMessageKind(native.SBASMessageMixedCorrectionsValue)
	// SBASMessageLongTermCorrections classifies an SBAS message carrying long-term corrections.
	SBASMessageLongTermCorrections SBASMessageKind = SBASMessageKind(native.SBASMessageLongTermCorrectionsValue)
	// SBASMessageIONODelays classifies an SBAS message carrying ionospheric delays.
	SBASMessageIONODelays SBASMessageKind = SBASMessageKind(native.SBASMessageIONODelaysValue)
	// SBASMessageUnsupported classifies an SBAS message carrying unsupported.
	SBASMessageUnsupported SBASMessageKind = SBASMessageKind(native.SBASMessageUnsupportedValue)
)

// SBASMessageInfo is the copied classification and record-count summary.
type SBASMessageInfo struct {
	// Form selects the SBAS wire representation used by the message block.
	Form SBASWireForm
	// Kind identifies the decoded SBAS message payload type.
	Kind SBASMessageKind
	// MessageType is the SBAS message type.
	MessageType uint8
	// Preamble is the SBAS preamble.
	Preamble       uint8
	FastCount      int
	LongTermCount  int
	IONODelayCount int
}

// SBASRawFastCorrections contains raw fast-correction fields decoded from an SBAS message.
type SBASRawFastCorrections struct {
	// Preamble and MessageType and IODF and IODP are the SBAS preamble, the SBAS message type, the issue-of-data fast index, the issue-of-data processing index.
	Preamble, MessageType, IODF, IODP uint8
	// PRC contains 13 raw/scaled fast pseudorange-correction fields in SBAS satellite order.
	PRC [13]int16
	// UDREI contains 13 fast-correction user-differential-range-error indicators.
	UDREI [13]uint8
}

// SBASFastDegradation contains SBAS fast-correction degradation parameters.
type SBASFastDegradation struct {
	// Preamble and SystemLatencyS and IODP are the SBAS preamble, the SBAS system latency in seconds, the issue-of-data processing index.
	Preamble, SystemLatencyS, IODP uint8
	// AI contains 51 fast-correction degradation indicators in SBAS satellite order.
	AI [51]uint8
}

// SBASGeoNavMessage contains raw/scaled SBAS GEO NAV position, rate, acceleration, and clock fields.
type SBASGeoNavMessage struct {
	// Preamble is the SBAS preamble.
	Preamble uint8
	// TimeOfDayS is the SBAS time-of-day field in seconds when present.
	TimeOfDayS uint16
	// URA is the decoded SBAS user-range-accuracy indicator.
	URA uint8
	// XM, YM, and ZM are raw/scaled position fields; XRateMPerS, YRateMPerS, and ZRateMPerS are raw/scaled velocity fields from the GEO NAV payload.
	XM, YM, ZM, XRateMPerS, YRateMPerS, ZRateMPerS int32
	// XAccelMPerS2, YAccelMPerS2, and ZAccelMPerS2 are raw/scaled acceleration fields; AGF0S and AGF1SPerS are raw/scaled clock fields from the GEO NAV payload.
	XAccelMPerS2, YAccelMPerS2, ZAccelMPerS2, AGF0S, AGF1SPerS int16
}

// SBASIGPMask contains the SBAS ionospheric-grid-point mask.
type SBASIGPMask struct {
	// Preamble and BandNumber and IODI are the SBAS preamble, the SBAS ionospheric-grid band number, the ionospheric-grid issue index.
	Preamble, BandNumber, IODI uint8
	// Mask contains 201 IGP mask bits in SBAS broadcast order.
	Mask [201]bool
}

// SBASIntegrity contains SBAS integrity and degradation indicators.
type SBASIntegrity struct {
	// Preamble is the SBAS preamble.
	Preamble uint8
	// IODF is the issue-of-data fast index.
	IODF [4]uint8
	// UDREI contains 51 integrity user-differential-range-error indicators.
	UDREI [51]uint8
}

// SBASIGPDelay contains one SBAS ionospheric-grid-point delay record.
type SBASIGPDelay struct {
	// VerticalDelay is the raw vertical-delay field from the SBAS payload.
	VerticalDelay uint16
	// GIVEI is the grid ionospheric vertical-error indicator.
	GIVEI uint8
}

// SBASIONODelays contains SBAS ionospheric delay corrections.
type SBASIONODelays struct {
	// Preamble is the SBAS preamble; BandNumber selects the IGP band; BlockID identifies the IGP block; IODI is the IGP issue index.
	Preamble, BandNumber, BlockID, IODI uint8
	// Entries contains 15 decoded IGP delay records in SBAS order.
	Entries [15]SBASIGPDelay
}

// SBASLongTermHalfInfo contains metadata for one long-term correction half.
type SBASLongTermHalfInfo struct {
	// VelocityCode reports whether the long-term correction uses velocity coding.
	VelocityCode bool
	// IODP is the issue-of-data processing index.
	IODP        uint8
	RecordCount int
}

// SBASLongTermRecord contains one SBAS long-term satellite correction record.
type SBASLongTermRecord struct {
	MonitoredIndex uint8
	// IODE is the ephemeris issue-of-data index.
	IODE uint8
	// DeltaX, DeltaY, and DeltaZ are raw/scaled X, Y, and Z correction fields.
	DeltaX, DeltaY, DeltaZ int32
	// DeltaXRate, DeltaYRate, and DeltaZRate are raw/scaled X, Y, and Z rate-correction fields.
	DeltaXRate, DeltaYRate, DeltaZRate int32
	// DeltaAF0 and DeltaAF1 are raw/scaled clock-bias and clock-drift correction fields.
	DeltaAF0, DeltaAF1 int32
	// HasTimeOfDayS reports whether TimeOfDayS is valid.
	HasTimeOfDayS bool
	// TimeOfDayS is the SBAS time-of-day field in seconds when present.
	TimeOfDayS uint32
}

// SBASMixedFastCorrections contains mixed SBAS fast-correction records.
type SBASMixedFastCorrections struct {
	// IODF is the fast-correction issue index; IODP is the processing issue index; BlockID identifies the mixed-correction block.
	IODF, IODP, BlockID uint8
	// PRC contains six raw/scaled fast pseudorange-correction fields.
	PRC [6]int16
	// UDREI contains six fast-correction user-differential-range-error indicators.
	UDREI [6]uint8
}

// SBASPRNMask contains the SBAS monitored-satellite mask.
type SBASPRNMask struct {
	// Preamble and IODP are the SBAS preamble, the issue-of-data processing index.
	Preamble, IODP uint8
	// Mask contains 210 monitored-satellite mask bits in SBAS broadcast order.
	Mask [210]bool
}

func sbasMessageInfoFromNative(value native.NativeSbasMessageInfo) SBASMessageInfo {
	return SBASMessageInfo{Form: SBASWireForm(value.Form), Kind: SBASMessageKind(value.Kind), MessageType: value.MessageType, Preamble: value.Preamble, FastCount: value.FastCount, LongTermCount: value.LongTermCount, IONODelayCount: value.IONODelayCount}
}

// SBASBlock owns a decoded SBAS message block. All payload accessors return
// copies; C remains responsible for bit decoding and raw scaling. Read-only
// methods may run concurrently with Close; Close waits for active calls to
// finish before releasing the native resource.
type SBASBlock struct {
	_      noCopy
	handle *native.SbasBlock
}

// DecodeSBASBlock parses the supplied representation and returns a decoded SBAS message block.
func DecodeSBASBlock(data []byte, form SBASWireForm) (*SBASBlock, error) {
	if err := validateSBASWireForm(form); err != nil {
		return nil, err
	}
	handle, err := native.DecodeSbasBlock(data, uint32(form))
	if err != nil {
		return nil, publicError(err)
	}
	return &SBASBlock{handle: handle}, nil
}

// Close releases the native SBAS block; repeated calls are safe.
func (block *SBASBlock) Close() error {
	if block == nil || block.handle == nil {
		return nil
	}
	return publicError(block.handle.Close())
}

// Info returns the SBAS wire form, message kind, type, preamble, and payload counts.
func (block *SBASBlock) Info() (SBASMessageInfo, error) {
	if block == nil || block.handle == nil {
		return SBASMessageInfo{}, ErrClosed
	}
	value, err := block.handle.Info()
	return sbasMessageInfoFromNative(value), publicError(err)
}

// Form returns the decoded SBAS wire form.
func (block *SBASBlock) Form() (SBASWireForm, error) {
	value, err := block.Info()
	return value.Form, err
}

// Kind returns the decoded message kind.
func (block *SBASBlock) Kind() (SBASMessageKind, error) {
	value, err := block.Info()
	return value.Kind, err
}

// Encode serializes the decoded SBAS block back to its wire representation.
func (block *SBASBlock) Encode() ([]byte, error) {
	if block == nil || block.handle == nil {
		return nil, ErrClosed
	}
	value, err := block.handle.Encode()
	return value, publicError(err)
}

// FastCorrections returns decoded SBAS fast-correction records.
func (block *SBASBlock) FastCorrections() (*SBASRawFastCorrections, error) {
	if block == nil || block.handle == nil {
		return nil, ErrClosed
	}
	present, value, err := block.handle.FastCorrections()
	if err != nil {
		return nil, publicError(err)
	}
	if !present {
		return nil, nil
	}
	return &SBASRawFastCorrections{Preamble: value.Preamble, MessageType: value.MessageType, IODF: value.IODF, IODP: value.IODP, PRC: value.PRC, UDREI: value.UDREI}, nil
}

// FastDegradation returns decoded SBAS fast-correction degradation data.
func (block *SBASBlock) FastDegradation() (*SBASFastDegradation, error) {
	if block == nil || block.handle == nil {
		return nil, ErrClosed
	}
	present, value, err := block.handle.FastDegradation()
	if err != nil {
		return nil, publicError(err)
	}
	if !present {
		return nil, nil
	}
	return &SBASFastDegradation{Preamble: value.Preamble, SystemLatencyS: value.SystemLatencyS, IODP: value.IODP, AI: value.AI}, nil
}

// GeoNav returns the decoded SBAS GEO navigation payload.
func (block *SBASBlock) GeoNav() (*SBASGeoNavMessage, error) {
	if block == nil || block.handle == nil {
		return nil, ErrClosed
	}
	present, value, err := block.handle.GeoNav()
	if err != nil {
		return nil, publicError(err)
	}
	if !present {
		return nil, nil
	}
	return &SBASGeoNavMessage{Preamble: value.Preamble, TimeOfDayS: value.TimeOfDayS, URA: value.URA, XM: value.XM, YM: value.YM, ZM: value.ZM, XRateMPerS: value.XRateMPerS, YRateMPerS: value.YRateMPerS, ZRateMPerS: value.ZRateMPerS, XAccelMPerS2: value.XAccelMPerS2, YAccelMPerS2: value.YAccelMPerS2, ZAccelMPerS2: value.ZAccelMPerS2, AGF0S: value.AGF0S, AGF1SPerS: value.AGF1SPerS}, nil
}

// IGPMask returns the decoded SBAS IGP mask payload.
func (block *SBASBlock) IGPMask() (*SBASIGPMask, error) {
	if block == nil || block.handle == nil {
		return nil, ErrClosed
	}
	present, value, err := block.handle.IGPMask()
	if err != nil {
		return nil, publicError(err)
	}
	if !present {
		return nil, nil
	}
	return &SBASIGPMask{Preamble: value.Preamble, BandNumber: value.BandNumber, IODI: value.IODI, Mask: value.Mask}, nil
}

// Integrity returns the decoded SBAS integrity payload.
func (block *SBASBlock) Integrity() (*SBASIntegrity, error) {
	if block == nil || block.handle == nil {
		return nil, ErrClosed
	}
	present, value, err := block.handle.Integrity()
	if err != nil {
		return nil, publicError(err)
	}
	if !present {
		return nil, nil
	}
	return &SBASIntegrity{Preamble: value.Preamble, IODF: value.IODF, UDREI: value.UDREI}, nil
}

// IONODelays returns the decoded SBAS ionospheric delays payload.
func (block *SBASBlock) IONODelays() (*SBASIONODelays, error) {
	if block == nil || block.handle == nil {
		return nil, ErrClosed
	}
	present, value, err := block.handle.IonoDelays()
	if err != nil {
		return nil, publicError(err)
	}
	if !present {
		return nil, nil
	}
	out := &SBASIONODelays{Preamble: value.Preamble, BandNumber: value.BandNumber, BlockID: value.BlockID, IODI: value.IODI}
	for index := range out.Entries {
		out.Entries[index] = SBASIGPDelay{VerticalDelay: value.Entries[index].VerticalDelay, GIVEI: value.Entries[index].GIVEI}
	}
	return out, nil
}

// MixedFastCorrections returns decoded SBAS mixed fast-correction records.
func (block *SBASBlock) MixedFastCorrections() (*SBASMixedFastCorrections, error) {
	if block == nil || block.handle == nil {
		return nil, ErrClosed
	}
	present, value, err := block.handle.MixedFastCorrections()
	if err != nil {
		return nil, publicError(err)
	}
	if !present {
		return nil, nil
	}
	return &SBASMixedFastCorrections{IODF: value.IODF, IODP: value.IODP, BlockID: value.BlockID, PRC: value.PRC, UDREI: value.UDREI}, nil
}

// PRNMask returns the decoded SBAS PRN mask payload.
func (block *SBASBlock) PRNMask() (*SBASPRNMask, error) {
	if block == nil || block.handle == nil {
		return nil, ErrClosed
	}
	present, value, err := block.handle.PrnMask()
	if err != nil {
		return nil, publicError(err)
	}
	if !present {
		return nil, nil
	}
	return &SBASPRNMask{Preamble: value.Preamble, IODP: value.IODP, Mask: value.Mask}, nil
}

// LongTermHalf returns the decoded SBAS long-term correction metadata.
func (block *SBASBlock) LongTermHalf(index int) (*SBASLongTermHalfInfo, error) {
	if block == nil || block.handle == nil {
		return nil, ErrClosed
	}
	present, value, err := block.handle.LongTermHalf(index)
	if err != nil {
		return nil, publicError(err)
	}
	if !present {
		return nil, nil
	}
	return &SBASLongTermHalfInfo{VelocityCode: value.VelocityCode, IODP: value.IODP, RecordCount: value.RecordCount}, nil
}

// LongTermRecords returns decoded SBAS long-term correction records.
func (block *SBASBlock) LongTermRecords(index int) ([]SBASLongTermRecord, error) {
	if block == nil || block.handle == nil {
		return nil, ErrClosed
	}
	values, err := block.handle.LongTermRecords(index)
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]SBASLongTermRecord, len(values))
	for i, value := range values {
		out[i] = SBASLongTermRecord{MonitoredIndex: value.MonitoredIndex, IODE: value.IODE, DeltaX: value.DeltaX, DeltaY: value.DeltaY, DeltaZ: value.DeltaZ, DeltaXRate: value.DeltaXRate, DeltaYRate: value.DeltaYRate, DeltaZRate: value.DeltaZRate, DeltaAF0: value.DeltaAF0, DeltaAF1: value.DeltaAF1, HasTimeOfDayS: value.HasTimeOfDayS, TimeOfDayS: value.TimeOfDayS}
	}
	return out, nil
}

// RawData returns detached raw bytes from the SBAS block.
func (block *SBASBlock) RawData() ([]byte, error) {
	if block == nil || block.handle == nil {
		return nil, ErrClosed
	}
	value, err := block.handle.RawData()
	return value, publicError(err)
}

// SBASLogBlock is copied metadata for one text-log row. Payload bytes are
// obtained separately so callers never retain native memory.
type SBASLogBlock struct {
	// SatelliteID is the GNSS satellite identifier.
	SatelliteID string
	// Epoch is the decoded GNSS week and time-of-week for the log block.
	Epoch GNSSWeekTow
	// Form selects the SBAS wire representation used by the log block.
	Form      SBASWireForm
	ByteCount int
}

// SBASLogBlocks owns parsed EMS or RTKLIB log blocks. Read-only methods may
// run concurrently with Close; Close waits for active calls to finish.
type SBASLogBlocks struct {
	_      noCopy
	handle *native.SbasLogBlocks
}

// ParseSBASEMSLines parses EMS text-log lines into detached SBAS log-block metadata.
func ParseSBASEMSLines(data []byte) (*SBASLogBlocks, error) {
	handle, err := native.ParseSbasEMSLines(data)
	if err != nil {
		return nil, publicError(err)
	}
	return &SBASLogBlocks{handle: handle}, nil
}

// ParseSBASRTKLIBLines parses RTKLIB text-log lines into detached SBAS log-block metadata.
func ParseSBASRTKLIBLines(data []byte) (*SBASLogBlocks, error) {
	handle, err := native.ParseSbasRTKLIBLines(data)
	if err != nil {
		return nil, publicError(err)
	}
	return &SBASLogBlocks{handle: handle}, nil
}

// Close releases the native SBAS log-block collection; repeated calls are safe.
func (blocks *SBASLogBlocks) Close() error {
	if blocks == nil || blocks.handle == nil {
		return nil
	}
	return publicError(blocks.handle.Close())
}

// Items returns detached metadata rows for each SBAS log block.
func (blocks *SBASLogBlocks) Items() ([]SBASLogBlock, error) {
	if blocks == nil || blocks.handle == nil {
		return nil, ErrClosed
	}
	values, err := blocks.handle.Items()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]SBASLogBlock, len(values))
	for index, value := range values {
		out[index] = SBASLogBlock{SatelliteID: value.SatelliteID, Epoch: GNSSWeekTow{System: TimeScale(value.Epoch.System), Week: value.Epoch.Week, TOWSeconds: value.Epoch.TOWSeconds}, Form: SBASWireForm(value.Form), ByteCount: value.ByteCount}
	}
	return out, nil
}

// Bytes returns detached raw bytes for the selected SBAS log block.
func (blocks *SBASLogBlocks) Bytes(index int) ([]byte, error) {
	if blocks == nil || blocks.handle == nil {
		return nil, ErrClosed
	}
	value, err := blocks.handle.Bytes(index)
	return value, publicError(err)
}

// Count returns the number of SBAS log blocks.
func (blocks *SBASLogBlocks) Count() (int, error) {
	if blocks == nil || blocks.handle == nil {
		return 0, ErrClosed
	}
	value, err := blocks.handle.Count()
	return value, publicError(err)
}

// SBASPRNToSatelliteID delegates the mapping to C. Present is false for the
// successful empty mapping result used by the C ABI.
// SBASPRNToSatelliteID maps an SBAS PRN to its public satellite identifier.
func SBASPRNToSatelliteID(prn uint16) (satelliteID string, present bool, err error) {
	satelliteID, present, err = native.SbasPRNToSatelliteID(prn)
	return satelliteID, present, publicError(err)
}

// SatelliteIDToSBASPRN is the finite inverse of the C-owned forward mapping.
func SatelliteIDToSBASPRN(satelliteID string) (uint16, bool, error) {
	for prn := uint16(120); prn <= 158; prn++ {
		mapped, present, err := SBASPRNToSatelliteID(prn)
		if err != nil {
			return 0, false, err
		}
		if present && mapped == satelliteID {
			return prn, true, nil
		}
	}
	return 0, false, nil
}
