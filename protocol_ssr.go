package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// RTCMSSRKind identifies the SSR message body family.
type RTCMSSRKind uint32

const (
	// RTCMSSROrbit identifies orbit corrections.
	RTCMSSROrbit RTCMSSRKind = RTCMSSRKind(native.RTCMSSROrbitValue)
	// RTCMSSRClock identifies clock corrections.
	RTCMSSRClock RTCMSSRKind = RTCMSSRKind(native.RTCMSSRClockValue)
	// RTCMSSRCombinedOrbitClock identifies combined orbit and clock corrections.
	RTCMSSRCombinedOrbitClock RTCMSSRKind = RTCMSSRKind(native.RTCMSSRCombinedOrbitClockValue)
	// RTCMSSRCodeBias identifies code-bias corrections.
	RTCMSSRCodeBias RTCMSSRKind = RTCMSSRKind(native.RTCMSSRCodeBiasValue)
	// RTCMSSRPhaseBias identifies phase-bias corrections.
	RTCMSSRPhaseBias RTCMSSRKind = RTCMSSRKind(native.RTCMSSRPhaseBiasValue)
	// RTCMSSRURA identifies user-range-accuracy corrections.
	RTCMSSRURA RTCMSSRKind = RTCMSSRKind(native.RTCMSSRURAValue)
	// RTCMSSRHighRateClock identifies high-rate clock corrections.
	RTCMSSRHighRateClock RTCMSSRKind = RTCMSSRKind(native.RTCMSSRHighRateClockValue)
	// RTCMSSRVTEC identifies vertical total-electron-content corrections.
	RTCMSSRVTEC RTCMSSRKind = RTCMSSRKind(native.RTCMSSRVTECValue)
)

// RTCMSSRHeader contains the scaled SSR message header fields.
type RTCMSSRHeader struct {
	// EpochTimeS is the native constellation epoch in seconds.
	EpochTimeS                   uint32
	UpdateInterval               uint8
	MultipleMessage              bool
	IODSSR                       uint8
	ProviderID                   uint16
	SolutionID                   uint8
	HasSatelliteReferenceDatum   bool
	SatelliteReferenceDatum      bool
	HasDispersiveBiasConsistency bool
	DispersiveBiasConsistency    bool
	HasMWConsistency             bool
	MWConsistency                bool
	SatelliteCount               uint8
}

// RTCMSSRInfo summarizes a decoded bare SSR body. Record values remain raw
// wire integers as defined by the C ABI.
type RTCMSSRInfo struct {
	// MessageNumber identifies the RTCM message type.
	MessageNumber  uint16
	System         GNSSSystem
	Kind           RTCMSSRKind
	Header         RTCMSSRHeader
	OrbitCount     int
	ClockCount     int
	URACount       int
	CodeBiasCount  int
	PhaseBiasCount int
}

// RTCMSSRClockRecord contains one raw SSR clock correction record.
type RTCMSSRClockRecord struct {
	// SatelliteID identifies the satellite; C0/C1/C2 are raw wire coefficients.
	SatelliteID uint8
	C0          int32
	C1          int32
	C2          int32
}

// RTCMSSRCodeBiasSignal contains one raw SSR code-bias signal row.
type RTCMSSRCodeBiasSignal struct {
	// SignalID identifies the signal; Bias is the raw wire bias value.
	SignalID uint8
	Bias     int16
}

// RTCMSSRCodeBiasRecord identifies one satellite's code-bias group.
type RTCMSSRCodeBiasRecord struct {
	// SatelliteID identifies the satellite; SignalCount is its native row count.
	SatelliteID uint8
	SignalCount int
}

// RTCMSSRCodeBiasGroup is a record together with its copied signal rows.
type RTCMSSRCodeBiasGroup struct {
	Record  RTCMSSRCodeBiasRecord
	Signals []RTCMSSRCodeBiasSignal
}

// RTCMSSROrbitRecord contains one raw SSR orbit correction record.
type RTCMSSROrbitRecord struct {
	// SatelliteID and IODE identify the correction record.
	SatelliteID    uint8
	IODE           uint32
	DeltaRadial    int32
	DeltaAlong     int32
	DeltaCross     int32
	DotDeltaRadial int32
	DotDeltaAlong  int32
	DotDeltaCross  int32
}

// RTCMSSRPhaseBiasSignal contains one raw SSR phase-bias signal row.
type RTCMSSRPhaseBiasSignal struct {
	// SignalID identifies the signal; indicators and Bias are raw wire values.
	SignalID                 uint8
	IntegerIndicator         uint8
	WideLaneIntegerIndicator uint8
	DiscontinuityCounter     uint8
	Bias                     int32
}

// RTCMSSRPhaseBiasRecord identifies one satellite's phase-bias group.
type RTCMSSRPhaseBiasRecord struct {
	// SatelliteID identifies the satellite; yaw values and SignalCount are native wire data.
	SatelliteID uint8
	YawAngle    uint16
	YawRate     int8
	SignalCount int
}

// RTCMSSRPhaseBiasGroup is a record together with its copied signal rows.
type RTCMSSRPhaseBiasGroup struct {
	Record  RTCMSSRPhaseBiasRecord
	Signals []RTCMSSRPhaseBiasSignal
}

// RTCMSSRURARecord contains one raw SSR user-range-accuracy row.
type RTCMSSRURARecord struct {
	// SatelliteID identifies the satellite; URAIndex is the raw wire index.
	SatelliteID uint8
	URAIndex    uint8
}

func ssrInfoFromNative(value native.NativeSsrInfo) RTCMSSRInfo {
	h := value.Header
	return RTCMSSRInfo{
		MessageNumber: value.MessageNumber, System: GNSSSystem(value.System), Kind: RTCMSSRKind(value.Kind),
		Header: RTCMSSRHeader{
			EpochTimeS: h.EpochTimeS, UpdateInterval: h.UpdateInterval, MultipleMessage: h.MultipleMessage, IODSSR: h.IODSSR,
			ProviderID: h.ProviderID, SolutionID: h.SolutionID, HasSatelliteReferenceDatum: h.HasSatelliteReferenceDatum,
			SatelliteReferenceDatum: h.SatelliteReferenceDatum, HasDispersiveBiasConsistency: h.HasDispersiveBiasConsistency,
			DispersiveBiasConsistency: h.DispersiveBiasConsistency, HasMWConsistency: h.HasMWConsistency, MWConsistency: h.MWConsistency,
			SatelliteCount: h.SatelliteCount,
		},
		OrbitCount: value.OrbitCount, ClockCount: value.ClockCount, URACount: value.URACount,
		CodeBiasCount: value.CodeBiasCount, PhaseBiasCount: value.PhaseBiasCount,
	}
}

// SSRMessage owns a decoded bare RTCM SSR body. The C decoder consumes the
// caller-provided bytes only during decoding; after successful decoding, the
// Go binding retains its own independent copy of the bare body in the
// binding-owned handle, and C does not retain that byte copy. It deliberately
// does not decode RTCM transport framing; callers pass the body after the
// transport preamble, length, and CRC have been removed. The caller's input
// slice is not retained. All read-only methods may run concurrently with one
// another and with Close; Close waits for active calls to finish before
// releasing the binding-owned handle.
// Body and Encode return new independent copies of the binding-retained bytes,
// synchronize with Close, and return ErrClosed after Close.
type SSRMessage struct {
	_      noCopy
	handle *native.SsrMessage
}

// DecodeSSRMessage decodes a bare RTCM SSR body. The C decoder consumes the
// caller-provided bytes only during decoding. After successful decoding, the
// binding retains an independent copy of the bare body in the returned handle;
// the caller's input slice is not retained and C does not retain that byte
// copy.
func DecodeSSRMessage(body []byte) (*SSRMessage, error) {
	handle, err := native.DecodeSsrMessage(body)
	if err != nil {
		return nil, publicError(err)
	}
	return &SSRMessage{handle: handle}, nil
}

// Close releases the decoded SSR message and is idempotent.
func (message *SSRMessage) Close() error {
	if message == nil || message.handle == nil {
		return nil
	}
	return publicError(message.handle.Close())
}

// Body returns a new independent copy of the exact bare body successfully
// decoded and retained by the binding. It synchronizes with Close and returns
// ErrClosed after Close.
func (message *SSRMessage) Body() ([]byte, error) {
	if message == nil || message.handle == nil {
		return nil, ErrClosed
	}
	value, err := message.handle.Body()
	return value, publicError(err)
}

// Encode returns a new independent copy of the exact bare body successfully
// decoded and retained by the binding. It is not a native encoder and does not
// re-encode the message in Go. Encode synchronizes with Close and returns
// ErrClosed after Close.
func (message *SSRMessage) Encode() ([]byte, error) {
	if message == nil || message.handle == nil {
		return nil, ErrClosed
	}
	value, err := message.handle.Encode()
	return value, publicError(err)
}

// Info returns detached SSR message metadata and record counts.
func (message *SSRMessage) Info() (RTCMSSRInfo, error) {
	if message == nil || message.handle == nil {
		return RTCMSSRInfo{}, ErrClosed
	}
	value, err := message.handle.Info()
	return ssrInfoFromNative(value), publicError(err)
}

// Orbits returns detached SSR orbit correction records.
func (message *SSRMessage) Orbits() ([]RTCMSSROrbitRecord, error) {
	if message == nil || message.handle == nil {
		return nil, ErrClosed
	}
	values, err := message.handle.Orbits()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]RTCMSSROrbitRecord, len(values))
	for index, value := range values {
		out[index] = RTCMSSROrbitRecord{SatelliteID: value.SatelliteID, IODE: value.IODE, DeltaRadial: value.DeltaRadial, DeltaAlong: value.DeltaAlong, DeltaCross: value.DeltaCross, DotDeltaRadial: value.DotDeltaRadial, DotDeltaAlong: value.DotDeltaAlong, DotDeltaCross: value.DotDeltaCross}
	}
	return out, nil
}

// Clocks returns detached SSR clock correction records.
func (message *SSRMessage) Clocks() ([]RTCMSSRClockRecord, error) {
	if message == nil || message.handle == nil {
		return nil, ErrClosed
	}
	values, err := message.handle.Clocks()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]RTCMSSRClockRecord, len(values))
	for index, value := range values {
		out[index] = RTCMSSRClockRecord{SatelliteID: value.SatelliteID, C0: value.C0, C1: value.C1, C2: value.C2}
	}
	return out, nil
}

// CodeBiases returns detached SSR code-bias group records.
func (message *SSRMessage) CodeBiases() ([]RTCMSSRCodeBiasRecord, error) {
	if message == nil || message.handle == nil {
		return nil, ErrClosed
	}
	values, err := message.handle.CodeBiases()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]RTCMSSRCodeBiasRecord, len(values))
	for index, value := range values {
		out[index] = RTCMSSRCodeBiasRecord{SatelliteID: value.SatelliteID, SignalCount: value.SignalCount}
	}
	return out, nil
}

// CodeBiasSignals returns detached signal rows for one code-bias record.
func (message *SSRMessage) CodeBiasSignals(recordIndex int) ([]RTCMSSRCodeBiasSignal, error) {
	if message == nil || message.handle == nil {
		return nil, ErrClosed
	}
	values, err := message.handle.CodeBiasSignals(recordIndex)
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]RTCMSSRCodeBiasSignal, len(values))
	for index, value := range values {
		out[index] = RTCMSSRCodeBiasSignal{SignalID: value.SignalID, Bias: value.Bias}
	}
	return out, nil
}

// CodeBiasGroups returns grouped per-satellite code-bias rows in message order.
func (message *SSRMessage) CodeBiasGroups() ([]RTCMSSRCodeBiasGroup, error) {
	records, err := message.CodeBiases()
	if err != nil {
		return nil, err
	}
	out := make([]RTCMSSRCodeBiasGroup, len(records))
	for index, record := range records {
		signals, err := message.CodeBiasSignals(index)
		if err != nil {
			return nil, err
		}
		out[index] = RTCMSSRCodeBiasGroup{Record: record, Signals: signals}
	}
	return out, nil
}

// PhaseBiases returns detached SSR phase-bias group records.
func (message *SSRMessage) PhaseBiases() ([]RTCMSSRPhaseBiasRecord, error) {
	if message == nil || message.handle == nil {
		return nil, ErrClosed
	}
	values, err := message.handle.PhaseBiases()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]RTCMSSRPhaseBiasRecord, len(values))
	for index, value := range values {
		out[index] = RTCMSSRPhaseBiasRecord{SatelliteID: value.SatelliteID, YawAngle: value.YawAngle, YawRate: value.YawRate, SignalCount: value.SignalCount}
	}
	return out, nil
}

// PhaseBiasSignals returns detached signal rows for one phase-bias record.
func (message *SSRMessage) PhaseBiasSignals(recordIndex int) ([]RTCMSSRPhaseBiasSignal, error) {
	if message == nil || message.handle == nil {
		return nil, ErrClosed
	}
	values, err := message.handle.PhaseBiasSignals(recordIndex)
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]RTCMSSRPhaseBiasSignal, len(values))
	for index, value := range values {
		out[index] = RTCMSSRPhaseBiasSignal{SignalID: value.SignalID, IntegerIndicator: value.IntegerIndicator, WideLaneIntegerIndicator: value.WideLaneIntegerIndicator, DiscontinuityCounter: value.DiscontinuityCounter, Bias: value.Bias}
	}
	return out, nil
}

// PhaseBiasGroups returns grouped per-satellite phase-bias rows in message order.
func (message *SSRMessage) PhaseBiasGroups() ([]RTCMSSRPhaseBiasGroup, error) {
	records, err := message.PhaseBiases()
	if err != nil {
		return nil, err
	}
	out := make([]RTCMSSRPhaseBiasGroup, len(records))
	for index, record := range records {
		signals, err := message.PhaseBiasSignals(index)
		if err != nil {
			return nil, err
		}
		out[index] = RTCMSSRPhaseBiasGroup{Record: record, Signals: signals}
	}
	return out, nil
}

// URA returns detached SSR user-range-accuracy records.
func (message *SSRMessage) URA() ([]RTCMSSRURARecord, error) {
	if message == nil || message.handle == nil {
		return nil, ErrClosed
	}
	values, err := message.handle.URA()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]RTCMSSRURARecord, len(values))
	for index, value := range values {
		out[index] = RTCMSSRURARecord{SatelliteID: value.SatelliteID, URAIndex: value.URAIndex}
	}
	return out, nil
}

// UpdateInterval and IODSSR are native wire header values.
// MultipleMessage reports the RTCM multiple-message flag.
// ProviderID and SolutionID identify the SSR provider and solution.
// HasSatelliteReferenceDatum controls SatelliteReferenceDatum.
// HasDispersiveBiasConsistency controls DispersiveBiasConsistency.
// HasMWConsistency controls MWConsistency.
// SatelliteCount is the native record count.
// System and Kind identify the constellation and SSR body family.
// Header contains the copied SSR header.
// The count fields are native record counts for each correction family.
// Delta fields are raw radial/along/cross orbit corrections.
