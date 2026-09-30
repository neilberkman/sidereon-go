package sidereon

import "sidereon.dev/go/v3/internal/native"

// RTCMPolicy selects whether RTCM operations reject or preserve recognized
// format departures.
type RTCMPolicy uint32

const (
	// RTCMPolicyStrict rejects a message that departs from the RTCM format.
	RTCMPolicyStrict RTCMPolicy = iota
	// RTCMPolicyLenient preserves readable data and reports format departures.
	RTCMPolicyLenient
)

// RTCMDepartureKind identifies a format departure retained by lenient RTCM
// encoding. Unknown numeric values remain available for forward compatibility.
type RTCMDepartureKind uint32

const (
	// RTCMDepartureFrameReservedBits identifies nonzero reserved frame bits.
	RTCMDepartureFrameReservedBits RTCMDepartureKind = iota
	// RTCMDepartureTrailingBits identifies data after a message's defined fields.
	RTCMDepartureTrailingBits
	// RTCMDepartureMSMCellMaskOver64 identifies an MSM cell mask beyond 64 cells.
	RTCMDepartureMSMCellMaskOver64
	// RTCMDepartureOrderExceedsDegree identifies a polynomial order above its degree.
	RTCMDepartureOrderExceedsDegree
	// RTCMDepartureSSRRecordsShort identifies an SSR record count mismatch.
	RTCMDepartureSSRRecordsShort
	// RTCMDepartureRecordsShort identifies a non-SSR record count mismatch.
	RTCMDepartureRecordsShort
	// RTCMDepartureUnknown identifies an unrecognized future departure kind.
	RTCMDepartureUnknown RTCMDepartureKind = 999
)

// RTCMDeparture retains the typed fields describing one encoding departure.
type RTCMDeparture struct {
	// Kind identifies the departure category, including unknown future values.
	Kind RTCMDepartureKind
	// MessageNumber is the RTCM message number associated with the departure.
	MessageNumber uint16
	// Reserved retains the reserved bits reported by the native parser.
	Reserved uint8
	// LayerIndex identifies the affected layer or record position when applicable.
	LayerIndex int
	// Degree is the reported polynomial degree when applicable.
	Degree uint8
	// Order is the reported polynomial order when applicable.
	Order uint8
	// Declared is the record or cell count declared by the message.
	Declared int
	// Read is the number of records or cells actually read.
	Read int
	// Cells is the associated cell count when applicable.
	Cells int
	// BitCount is the number of extra bits when applicable.
	BitCount int
}

// RTCMStreamDeparture retains the offset and description of one leniently
// decoded frame departure.
type RTCMStreamDeparture struct {
	// Offset is the byte offset of the frame preamble in the input stream.
	Offset int
	// Description is an owned native description of the departure.
	Description string
}

// DecodeRTCMStreamWithPolicy scans a byte stream under policy and returns
// owning messages and diagnostics. Lenient mode retains recoverable format
// departures; strict mode skips such frames and records skipped-frame and CRC
// diagnostics while continuing to scan the stream.
func DecodeRTCMStreamWithPolicy(data []byte, policy RTCMPolicy) (*RTCMMessages, *RTCMStreamDiagnostics, error) {
	m, d, err := native.DecodeRTCMStreamWithPolicy(data, uint32(policy))
	if err != nil {
		return nil, nil, publicError(err)
	}
	return &RTCMMessages{handle: m}, &RTCMStreamDiagnostics{handle: d}, nil
}

// EncodeWithPolicy returns the detached message body and any departures
// accepted under policy.
func (m *RTCMMessages) EncodeWithPolicy(index int, policy RTCMPolicy) ([]byte, []RTCMDeparture, error) {
	if m == nil || m.handle == nil {
		return nil, nil, ErrClosed
	}
	b, d, err := m.handle.EncodeWithPolicy(index, uint32(policy))
	return b, publicRTCMDepartures(d), publicError(err)
}

// FrameWithPolicy returns a detached transport frame and any departures
// accepted under policy.
func (m *RTCMMessages) FrameWithPolicy(index int, policy RTCMPolicy) ([]byte, []RTCMDeparture, error) {
	if m == nil || m.handle == nil {
		return nil, nil, ErrClosed
	}
	b, d, err := m.handle.FrameWithPolicy(index, uint32(policy))
	return b, publicRTCMDepartures(d), publicError(err)
}

func publicRTCMDepartures(values []native.NativeRTCMDeparture) []RTCMDeparture {
	result := make([]RTCMDeparture, len(values))
	for i, value := range values {
		kind := RTCMDepartureKind(value.Kind)
		result[i] = RTCMDeparture{Kind: kind, MessageNumber: value.MessageNumber, Reserved: value.Reserved, LayerIndex: value.LayerIndex, Degree: value.Degree, Order: value.Order, Declared: value.Declared, Read: value.Read, Cells: value.Cells, BitCount: value.BitCount}
	}
	return result
}

// CRCFailures returns the number of complete frames rejected for a CRC-24Q
// mismatch during stream decoding.
func (d *RTCMStreamDiagnostics) CRCFailures() (int, error) {
	if d == nil || d.handle == nil {
		return 0, ErrClosed
	}
	v, err := d.handle.CRCFailures()
	return v, publicError(err)
}

// DepartureCount returns the number of leniently accepted frame departures.
func (d *RTCMStreamDiagnostics) DepartureCount() (int, error) {
	if d == nil || d.handle == nil {
		return 0, ErrClosed
	}
	v, err := d.handle.DepartureCount()
	return v, publicError(err)
}

// Departure returns an owned description for one leniently accepted frame.
func (d *RTCMStreamDiagnostics) Departure(index int) (RTCMStreamDeparture, error) {
	if d == nil || d.handle == nil {
		return RTCMStreamDeparture{}, ErrClosed
	}
	v, err := d.handle.Departure(index)
	return RTCMStreamDeparture{Offset: v.Offset, Description: v.Description}, publicError(err)
}
