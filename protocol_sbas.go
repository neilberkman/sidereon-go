package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// SBASWireForm identifies the input/output wire shape of an SBAS message.
type SBASWireForm uint32

const (
	SBASWireFramed250 SBASWireForm = SBASWireForm(native.SBASWireFramed250Value)
	SBASWireBody226   SBASWireForm = SBASWireForm(native.SBASWireBody226Value)
)

// SBASMessageKind classifies a decoded SBAS message payload.
type SBASMessageKind uint32

const (
	SBASMessageDoNotUse            SBASMessageKind = SBASMessageKind(native.SBASMessageDoNotUseValue)
	SBASMessagePRNMask             SBASMessageKind = SBASMessageKind(native.SBASMessagePRNMaskValue)
	SBASMessageFastCorrections     SBASMessageKind = SBASMessageKind(native.SBASMessageFastCorrectionsValue)
	SBASMessageIntegrity           SBASMessageKind = SBASMessageKind(native.SBASMessageIntegrityValue)
	SBASMessageFastDegradation     SBASMessageKind = SBASMessageKind(native.SBASMessageFastDegradationValue)
	SBASMessageGeoNav              SBASMessageKind = SBASMessageKind(native.SBASMessageGeoNavValue)
	SBASMessageNetworkTime         SBASMessageKind = SBASMessageKind(native.SBASMessageNetworkTimeValue)
	SBASMessageGeoAlmanac          SBASMessageKind = SBASMessageKind(native.SBASMessageGeoAlmanacValue)
	SBASMessageIGPMask             SBASMessageKind = SBASMessageKind(native.SBASMessageIGPMaskValue)
	SBASMessageMixedCorrections    SBASMessageKind = SBASMessageKind(native.SBASMessageMixedCorrectionsValue)
	SBASMessageLongTermCorrections SBASMessageKind = SBASMessageKind(native.SBASMessageLongTermCorrectionsValue)
	SBASMessageIONODelays          SBASMessageKind = SBASMessageKind(native.SBASMessageIONODelaysValue)
	SBASMessageUnsupported         SBASMessageKind = SBASMessageKind(native.SBASMessageUnsupportedValue)
)

// SBASMessageInfo is the copied classification and record-count summary.
type SBASMessageInfo struct {
	Form           SBASWireForm
	Kind           SBASMessageKind
	MessageType    uint8
	Preamble       uint8
	FastCount      int
	LongTermCount  int
	IONODelayCount int
}

type SBASRawFastCorrections struct {
	Preamble, MessageType, IODF, IODP uint8
	PRC                               [13]int16
	UDREI                             [13]uint8
}

type SBASFastDegradation struct {
	Preamble, SystemLatencyS, IODP uint8
	AI                             [51]uint8
}

// SBASGeoNavMessage contains raw SBAS GEO navigation fields. Position values
// are meters, rates are meters per second, accelerations are meters per
// second squared, and clock fields are seconds or seconds per second.
type SBASGeoNavMessage struct {
	Preamble                                                   uint8
	TimeOfDayS                                                 uint16
	URA                                                        uint8
	XM, YM, ZM, XRateMPerS, YRateMPerS, ZRateMPerS             int32
	XAccelMPerS2, YAccelMPerS2, ZAccelMPerS2, AGF0S, AGF1SPerS int16
}

type SBASIGPMask struct {
	Preamble, BandNumber, IODI uint8
	Mask                       [201]bool
}

type SBASIntegrity struct {
	Preamble uint8
	IODF     [4]uint8
	UDREI    [51]uint8
}

type SBASIGPDelay struct {
	VerticalDelay uint16
	GIVEI         uint8
}

type SBASIONODelays struct {
	Preamble, BandNumber, BlockID, IODI uint8
	Entries                             [15]SBASIGPDelay
}

type SBASLongTermHalfInfo struct {
	VelocityCode bool
	IODP         uint8
	RecordCount  int
}

type SBASLongTermRecord struct {
	MonitoredIndex                     uint8
	IODE                               uint8
	DeltaX, DeltaY, DeltaZ             int32
	DeltaXRate, DeltaYRate, DeltaZRate int32
	DeltaAF0, DeltaAF1                 int32
	HasTimeOfDayS                      bool
	TimeOfDayS                         uint32
}

type SBASMixedFastCorrections struct {
	IODF, IODP, BlockID uint8
	PRC                 [6]int16
	UDREI               [6]uint8
}

type SBASPRNMask struct {
	Preamble, IODP uint8
	Mask           [210]bool
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

func (block *SBASBlock) Close() error {
	if block == nil || block.handle == nil {
		return nil
	}
	return publicError(block.handle.Close())
}

func (block *SBASBlock) Info() (SBASMessageInfo, error) {
	if block == nil || block.handle == nil {
		return SBASMessageInfo{}, ErrClosed
	}
	value, err := block.handle.Info()
	return sbasMessageInfoFromNative(value), publicError(err)
}

func (block *SBASBlock) Form() (SBASWireForm, error) {
	value, err := block.Info()
	return value.Form, err
}

func (block *SBASBlock) Kind() (SBASMessageKind, error) {
	value, err := block.Info()
	return value.Kind, err
}

func (block *SBASBlock) Encode() ([]byte, error) {
	if block == nil || block.handle == nil {
		return nil, ErrClosed
	}
	value, err := block.handle.Encode()
	return value, publicError(err)
}

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
	SatelliteID string
	Epoch       GNSSWeekTow
	Form        SBASWireForm
	ByteCount   int
}

// SBASLogBlocks owns parsed EMS or RTKLIB log blocks. Read-only methods may
// run concurrently with Close; Close waits for active calls to finish.
type SBASLogBlocks struct {
	_      noCopy
	handle *native.SbasLogBlocks
}

func ParseSBASEMSLines(data []byte) (*SBASLogBlocks, error) {
	handle, err := native.ParseSbasEMSLines(data)
	if err != nil {
		return nil, publicError(err)
	}
	return &SBASLogBlocks{handle: handle}, nil
}

func ParseSBASRTKLIBLines(data []byte) (*SBASLogBlocks, error) {
	handle, err := native.ParseSbasRTKLIBLines(data)
	if err != nil {
		return nil, publicError(err)
	}
	return &SBASLogBlocks{handle: handle}, nil
}

func (blocks *SBASLogBlocks) Close() error {
	if blocks == nil || blocks.handle == nil {
		return nil
	}
	return publicError(blocks.handle.Close())
}

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

func (blocks *SBASLogBlocks) Bytes(index int) ([]byte, error) {
	if blocks == nil || blocks.handle == nil {
		return nil, ErrClosed
	}
	value, err := blocks.handle.Bytes(index)
	return value, publicError(err)
}

func (blocks *SBASLogBlocks) Count() (int, error) {
	if blocks == nil || blocks.handle == nil {
		return 0, ErrClosed
	}
	value, err := blocks.handle.Count()
	return value, publicError(err)
}

// SBASPRNToSatelliteID delegates the mapping to C. Present is false for the
// successful empty mapping result used by the C ABI.
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
