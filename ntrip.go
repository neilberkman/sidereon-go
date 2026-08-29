package sidereon

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"runtime"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/neilberkman/sidereon-go/internal/native"
)

// NTRIPVersion selects the C-generated request wire version.
type NTRIPVersion uint32

const (
	// NTRIPVersionRev1 identifies the ntrip version rev1 case.
	NTRIPVersionRev1 NTRIPVersion = NTRIPVersion(native.NTRIPVersionRev1)
	// NTRIPVersionRev2 identifies the ntrip version rev2 case.
	NTRIPVersionRev2 NTRIPVersion = NTRIPVersion(native.NTRIPVersionRev2)
)

// NTRIPState is the C sans-I/O machine state.
type NTRIPState uint32

const (
	// NTRIPStateIdle identifies the ntrip state idle case.
	NTRIPStateIdle NTRIPState = NTRIPState(native.NTRIPStateIdle)
	// NTRIPStateAwaitingStatus identifies the ntrip state awaiting status case.
	NTRIPStateAwaitingStatus NTRIPState = NTRIPState(native.NTRIPStateAwaitingStatus)
	// NTRIPStateAwaitingHeaders identifies the ntrip state awaiting headers case.
	NTRIPStateAwaitingHeaders NTRIPState = NTRIPState(native.NTRIPStateAwaitingHeaders)
	// NTRIPStateStreaming identifies the ntrip state streaming case.
	NTRIPStateStreaming NTRIPState = NTRIPState(native.NTRIPStateStreaming)
	// NTRIPStateSourcetable identifies the ntrip state sourcetable case.
	NTRIPStateSourcetable NTRIPState = NTRIPState(native.NTRIPStateSourcetable)
	// NTRIPStateClosed identifies the ntrip state closed case.
	NTRIPStateClosed NTRIPState = NTRIPState(native.NTRIPStateClosed)
)

// NTRIPEventKind identifies an event emitted by the C machine.
type NTRIPEventKind uint32

const (
	// NTRIPEventConnected identifies the ntrip event connected case.
	NTRIPEventConnected NTRIPEventKind = NTRIPEventKind(native.NTRIPEventConnected)
	// NTRIPEventPayload identifies the ntrip event payload case.
	NTRIPEventPayload NTRIPEventKind = NTRIPEventKind(native.NTRIPEventPayload)
	// NTRIPEventSourcetable identifies the ntrip event sourcetable case.
	NTRIPEventSourcetable NTRIPEventKind = NTRIPEventKind(native.NTRIPEventSourcetable)
	// NTRIPEventRejected identifies the ntrip event rejected case.
	NTRIPEventRejected NTRIPEventKind = NTRIPEventKind(native.NTRIPEventRejected)
	// NTRIPEventStreamCorrupted identifies the ntrip event stream corrupted case.
	NTRIPEventStreamCorrupted NTRIPEventKind = NTRIPEventKind(native.NTRIPEventStreamCorrupted)
	// NTRIPEventStreamEnded identifies the ntrip event stream ended case.
	NTRIPEventStreamEnded NTRIPEventKind = NTRIPEventKind(native.NTRIPEventStreamEnded)
)

// NTRIPRejectionKind identifies a protocol rejection emitted by C.
type NTRIPRejectionKind uint32

const (
	// NTRIPRejectionNone identifies the ntrip rejection none case.
	NTRIPRejectionNone NTRIPRejectionKind = NTRIPRejectionKind(native.NTRIPRejectionNone)
	// NTRIPRejectionUnauthorized identifies the ntrip rejection unauthorized case.
	NTRIPRejectionUnauthorized NTRIPRejectionKind = NTRIPRejectionKind(native.NTRIPRejectionUnauthorized)
	// NTRIPRejectionMountpointNotFound identifies the ntrip rejection mountpoint not found case.
	NTRIPRejectionMountpointNotFound NTRIPRejectionKind = NTRIPRejectionKind(native.NTRIPRejectionMountpointNotFound)
	// NTRIPRejectionDigestRequired identifies the ntrip rejection digest required case.
	NTRIPRejectionDigestRequired NTRIPRejectionKind = NTRIPRejectionKind(native.NTRIPRejectionDigestRequired)
	// NTRIPRejectionCasterError identifies the ntrip rejection caster error case.
	NTRIPRejectionCasterError NTRIPRejectionKind = NTRIPRejectionKind(native.NTRIPRejectionCasterError)
	// NTRIPRejectionUnexpectedContentType identifies the ntrip rejection unexpected content type case.
	NTRIPRejectionUnexpectedContentType NTRIPRejectionKind = NTRIPRejectionKind(native.NTRIPRejectionUnexpectedContentType)
	// NTRIPRejectionHTTPError identifies the ntrip rejection http error case.
	NTRIPRejectionHTTPError NTRIPRejectionKind = NTRIPRejectionKind(native.NTRIPRejectionHTTPError)
	// NTRIPRejectionMalformedHandshake identifies the ntrip rejection malformed handshake case.
	NTRIPRejectionMalformedHandshake NTRIPRejectionKind = NTRIPRejectionKind(native.NTRIPRejectionMalformedHandshake)
)

// NTRIPSourcetableAuth identifies an STR authentication field.
type NTRIPSourcetableAuth uint32

const (
	// NTRIPSourcetableAuthNone identifies the ntrip sourcetable auth none case.
	NTRIPSourcetableAuthNone NTRIPSourcetableAuth = NTRIPSourcetableAuth(native.NTRIPSourcetableAuthNone)
	// NTRIPSourcetableAuthBasic identifies the ntrip sourcetable auth basic case.
	NTRIPSourcetableAuthBasic NTRIPSourcetableAuth = NTRIPSourcetableAuth(native.NTRIPSourcetableAuthBasic)
	// NTRIPSourcetableAuthDigest identifies the ntrip sourcetable auth digest case.
	NTRIPSourcetableAuthDigest NTRIPSourcetableAuth = NTRIPSourcetableAuth(native.NTRIPSourcetableAuthDigest)
	// NTRIPSourcetableAuthOther identifies the ntrip sourcetable auth other case.
	NTRIPSourcetableAuthOther NTRIPSourcetableAuth = NTRIPSourcetableAuth(native.NTRIPSourcetableAuthOther)
)

// NTRIPConfig is the C configuration for a caster. GGAIntervalS is in
// seconds; set HasGGAInterval when a zero interval is intentionally supplied.
type NTRIPConfig struct {
	// Host is the host value for NTRIPConfig.
	Host string
	// Port is the port value for NTRIPConfig.
	Port uint16
	// Mountpoint is the mountpoint value for NTRIPConfig.
	Mountpoint string
	// Version is the version value for NTRIPConfig.
	Version NTRIPVersion
	// Username is the username value for NTRIPConfig.
	Username string
	// Password is the password value for NTRIPConfig.
	Password string
	// HasCredentials reports whether the has credentials field is present.
	HasCredentials bool
	// UserAgent is the user agent value for NTRIPConfig.
	UserAgent string
	// GGAIntervalS is the gga interval s in seconds.
	GGAIntervalS float64
	// HasGGAInterval reports whether the has gga interval field is present.
	HasGGAInterval bool
}

func (c NTRIPConfig) native() native.NTRIPConfig {
	version := c.Version
	if version == 0 {
		version = NTRIPVersionRev2
	}
	port := c.Port
	if port == 0 {
		port = 2101
	}
	return native.NTRIPConfig{Host: c.Host, Port: port, Mountpoint: c.Mountpoint, Version: native.NTRIPVersion(version), Username: c.Username, Password: c.Password, HasCredentials: c.HasCredentials, UserAgent: c.UserAgent, GGAIntervalS: c.GGAIntervalS, HasGGAInterval: c.HasGGAInterval}
}

func validateNTRIPConfig(config NTRIPConfig) error {
	for _, field := range []struct {
		name  string
		value string
	}{
		{name: "NTRIP host", value: config.Host},
		{name: "NTRIP mountpoint", value: config.Mountpoint},
		{name: "NTRIP username", value: config.Username},
		{name: "NTRIP password", value: config.Password},
		{name: "NTRIP user agent", value: config.UserAgent},
	} {
		if strings.IndexByte(field.value, 0) >= 0 {
			return fmt.Errorf("sidereon: %s contains an embedded NUL byte", field.name)
		}
	}
	return nil
}

// NTRIPGGAPosition contains decimal-degree coordinates, height in metres, and
// the fix metadata forwarded to C's GGA formatter.
type NTRIPGGAPosition struct {
	// LatitudeDeg is the latitude deg in degrees.
	LatitudeDeg float64
	// LongitudeDeg is the longitude deg in degrees.
	LongitudeDeg float64
	// HeightM is the height m in metres.
	HeightM float64
	// FixQuality is the fix quality value for NTRIPGGAPosition.
	FixQuality uint8
	// Satellites is the satellites value for NTRIPGGAPosition.
	Satellites uint8
	// HDOP is the hdop value for NTRIPGGAPosition.
	HDOP float64
}

// NTRIPEvent is a fully copied C event. Payload and Detail never alias C or a
// network read buffer. Sourcetable is independently owned when present.
type NTRIPEvent struct {
	// Kind is the event or record kind.
	Kind NTRIPEventKind
	// Version is the version value for NTRIPEvent.
	Version NTRIPVersion
	// Chunked is the chunked value for NTRIPEvent.
	Chunked bool
	// HeaderCount identifies or counts this record.
	HeaderCount int
	// Payload contains a detached copy; nil means this field is absent.
	Payload []byte
	// SourcetableRecords is the sourcetable records value for NTRIPEvent.
	SourcetableRecords int
	// Rejection is the rejection value for NTRIPEvent.
	Rejection NTRIPRejectionKind
	// HTTPStatus is the httpstatus value for NTRIPEvent.
	HTTPStatus uint16
	// Detail contains a detached copy; nil means this field is absent.
	Detail []byte
	// Sourcetable is independently owned by the recipient and must be closed
	// when no longer needed. This obligation also applies to events delivered
	// to an NTRIPClient.Run callback.
	Sourcetable *NTRIPSourcetable
}

// NTRIPStream is a copied typed STR record from a sourcetable.
type NTRIPStream struct {
	// Mountpoint is the mountpoint value for NTRIPStream.
	Mountpoint string
	// Identifier is the identifier value for NTRIPStream.
	Identifier string
	// Format is the product format.
	Format string
	// FormatDetails is the format details value for NTRIPStream.
	FormatDetails string
	// HasCarrier reports whether the has carrier field is present.
	HasCarrier bool
	// Carrier is the carrier value for NTRIPStream.
	Carrier uint8
	// NavSystem is the nav system value for NTRIPStream.
	NavSystem string
	// Network is the network value for NTRIPStream.
	Network string
	// Country identifies or counts this record.
	Country string
	// HasLatitudeDeg reports whether the has latitude deg field is present.
	HasLatitudeDeg bool
	// LatitudeDeg is the latitude deg in degrees.
	LatitudeDeg float64
	// HasLongitudeDeg reports whether the has longitude deg field is present.
	HasLongitudeDeg bool
	// LongitudeDeg is the longitude deg in degrees.
	LongitudeDeg float64
	// HasNMEARequired reports whether the has nmea required field is present.
	HasNMEARequired bool
	// NMEARequired is the nmearequired value for NTRIPStream.
	NMEARequired bool
	// HasNetworkSolution reports whether the has network solution field is present.
	HasNetworkSolution bool
	// NetworkSolution is the network solution value for NTRIPStream.
	NetworkSolution bool
	// Generator is the generator value for NTRIPStream.
	Generator string
	// Compression is the archive compression.
	Compression string
	// Authentication is the authentication value for NTRIPStream.
	Authentication NTRIPSourcetableAuth
	// HasFee reports whether the has fee field is present.
	HasFee bool
	// Fee is the fee value for NTRIPStream.
	Fee bool
	// HasBitrate reports whether the has bitrate field is present.
	HasBitrate bool
	// Bitrate is the bitrate value for NTRIPStream.
	Bitrate uint32
	// Misc is the misc value for NTRIPStream.
	Misc string
}

// NTRIPSourcetableSummary contains C's record and STR-stream counts.
type NTRIPSourcetableSummary struct {
	// RecordCount identifies or counts this record.
	RecordCount int
	// StreamCount identifies or counts this record.
	StreamCount int
}

// NTRIPSourcetable owns a read-only C-parsed sourcetable. It is independently
// owned from the event batch that produced it and must be closed by its
// recipient. Summary, Streams, and Text may run concurrently; Close waits for
// active reads, clears the native resource, and is idempotent. Values must not
// be copied after first use.
type NTRIPSourcetable struct {
	_      noCopy
	handle *native.NTRIPSourcetable
}

// NTRIPMachine is a mutable deterministic sans-I/O C protocol state machine.
// Its operations are serialized. Close waits for an active operation, clears
// the native resource, and is idempotent. Values must not be copied after first
// use.
type NTRIPMachine struct {
	_      noCopy
	handle *native.NTRIPMachine
}

// NTRIPRequestBytes returns the exact C-generated request bytes without opening
// a connection.
func NTRIPRequestBytes(config NTRIPConfig) ([]byte, error) {
	if err := validateNTRIPConfig(config); err != nil {
		return nil, err
	}
	data, err := native.NTRIPRequestBytes(config.native())
	return data, publicError(err)
}

// NewNTRIPMachine creates a mutable C protocol machine.
func NewNTRIPMachine(config NTRIPConfig) (*NTRIPMachine, error) {
	if err := validateNTRIPConfig(config); err != nil {
		return nil, err
	}
	handle, err := native.NewNTRIPMachine(config.native())
	if err != nil {
		return nil, publicError(err)
	}
	return &NTRIPMachine{handle: handle}, nil
}

// ParseNTRIPSourcetable parses bytes through the C parser. The input is copied
// into C-owned memory for the duration of the call only; the resulting Go and
// native objects do not retain caller memory.
func ParseNTRIPSourcetable(data []byte) (*NTRIPSourcetable, error) {
	handle, err := native.ParseNTRIPSourcetable(data)
	if err != nil {
		return nil, publicError(err)
	}
	return &NTRIPSourcetable{handle: handle}, nil
}

// Close releases the C machine and is idempotent.
func (m *NTRIPMachine) Close() error {
	if m == nil || m.handle == nil {
		return nil
	}
	m.handle.Close()
	return nil
}

// ConnectionRequest returns the exact C-generated request bytes.
func (m *NTRIPMachine) ConnectionRequest() ([]byte, error) {
	if m == nil || m.handle == nil {
		return nil, ErrClosed
	}
	bytes, err := m.handle.ConnectionRequest()
	if err != nil {
		return nil, publicError(err)
	}
	defer bytes.Close()
	data, err := bytes.Bytes()
	return data, publicError(err)
}

// Push feeds one arbitrary network chunk to C and returns copied events. The
// input is copied into C-owned memory for the duration of the call only; the
// resulting Go and native objects do not retain caller memory.
func (m *NTRIPMachine) Push(data []byte) ([]NTRIPEvent, error) {
	if m == nil || m.handle == nil {
		return nil, ErrClosed
	}
	events, err := m.handle.Push(data)
	if err != nil {
		return nil, publicError(err)
	}
	defer events.Close()
	return copyNTRIPEvents(events)
}

// Finish tells C that the byte stream has ended and returns copied terminal
// events.
func (m *NTRIPMachine) Finish() ([]NTRIPEvent, error) {
	if m == nil || m.handle == nil {
		return nil, ErrClosed
	}
	events, err := m.handle.Finish()
	if err != nil {
		return nil, publicError(err)
	}
	defer events.Close()
	return copyNTRIPEvents(events)
}

// Reset resets the C machine for another connection.
func (m *NTRIPMachine) Reset() error {
	if m == nil || m.handle == nil {
		return ErrClosed
	}
	return publicError(m.handle.Reset())
}

// State returns the current C machine state.
func (m *NTRIPMachine) State() (NTRIPState, error) {
	if m == nil || m.handle == nil {
		return 0, ErrClosed
	}
	state, err := m.handle.State()
	return NTRIPState(state), publicError(err)
}

// TryGGAMessage asks C whether an interval-eligible exact GGA sentence should
// be sent. nowS and utcSecondsOfDay are seconds, with the latter in [0,86400).
func (m *NTRIPMachine) TryGGAMessage(nowS, utcSecondsOfDay float64, position NTRIPGGAPosition) ([]byte, bool, error) {
	if m == nil || m.handle == nil {
		return nil, false, ErrClosed
	}
	bytes, present, err := m.handle.TryGGAMessage(nowS, native.NTRIPGGAPosition{LatitudeDeg: position.LatitudeDeg, LongitudeDeg: position.LongitudeDeg, HeightM: position.HeightM, FixQuality: position.FixQuality, Satellites: position.Satellites, HDOP: position.HDOP}, utcSecondsOfDay)
	if err != nil {
		return nil, false, publicError(err)
	}
	if !present {
		return nil, false, nil
	}
	defer bytes.Close()
	data, err := bytes.Bytes()
	return data, true, publicError(err)
}

func copyNTRIPEvents(events *native.NTRIPEvents) (result []NTRIPEvent, err error) {
	count, err := events.Count()
	if err != nil {
		return nil, publicError(err)
	}
	if count < 0 || uint64(count) > uint64(^uint(0)>>1)/uint64(unsafe.Sizeof(NTRIPEvent{})) {
		return nil, errors.New("sidereon: NTRIP event count is too large to allocate")
	}
	result = make([]NTRIPEvent, count)
	ownedTables := make([]*NTRIPSourcetable, 0)
	defer func() {
		if err == nil {
			return
		}
		for _, table := range ownedTables {
			if table != nil {
				err = errors.Join(err, table.Close())
			}
		}
		for i := range result {
			if result[i].Sourcetable != nil {
				result[i].Sourcetable = nil
			}
		}
		result = nil
	}()
	for i := range result {
		info, infoErr := events.Event(i)
		if infoErr != nil {
			return result, publicError(infoErr)
		}
		result[i] = NTRIPEvent{Kind: NTRIPEventKind(info.Kind), Version: NTRIPVersion(info.Version), Chunked: info.Chunked, HeaderCount: info.HeaderCount, SourcetableRecords: info.SourcetableRecords, Rejection: NTRIPRejectionKind(info.Rejection), HTTPStatus: info.HTTPStatus}
		if info.Kind == native.NTRIPEventPayload {
			result[i].Payload, err = events.Payload(i)
			if err != nil {
				return result, publicError(err)
			}
			if len(result[i].Payload) != info.PayloadLength {
				return result, fmt.Errorf("sidereon: NTRIP payload length changed from %d to %d", info.PayloadLength, len(result[i].Payload))
			}
		}
		result[i].Detail, err = events.Detail(i)
		if err != nil {
			return result, publicError(err)
		}
		if info.Kind == native.NTRIPEventSourcetable {
			table, tableErr := events.Sourcetable(i)
			if tableErr != nil {
				return result, publicError(tableErr)
			}
			result[i].Sourcetable = &NTRIPSourcetable{handle: table}
			ownedTables = append(ownedTables, result[i].Sourcetable)
		}
	}
	return result, nil
}

// Close releases the C sourcetable and is idempotent.
func (t *NTRIPSourcetable) Close() error {
	if t == nil || t.handle == nil {
		return nil
	}
	t.handle.Close()
	return nil
}

// Summary returns C's sourcetable counts and may run concurrently with
// Streams, Text, and other Summary calls.
func (t *NTRIPSourcetable) Summary() (NTRIPSourcetableSummary, error) {
	if t == nil || t.handle == nil {
		return NTRIPSourcetableSummary{}, ErrClosed
	}
	records, streams, err := t.handle.Summary()
	return NTRIPSourcetableSummary{RecordCount: records, StreamCount: streams}, publicError(err)
}

// Streams returns the C-parsed STR records and may run concurrently with
// Summary, Text, and other Streams calls.
func (t *NTRIPSourcetable) Streams() ([]NTRIPStream, error) {
	if t == nil || t.handle == nil {
		return nil, ErrClosed
	}
	values, err := t.handle.Streams()
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]NTRIPStream, len(values))
	for i, value := range values {
		result[i] = NTRIPStream{Mountpoint: value.Mountpoint, Identifier: value.Identifier, Format: value.Format, FormatDetails: value.FormatDetails, HasCarrier: value.HasCarrier, Carrier: value.Carrier, NavSystem: value.NavSystem, Network: value.Network, Country: value.Country, HasLatitudeDeg: value.HasLatitudeDeg, LatitudeDeg: value.LatitudeDeg, HasLongitudeDeg: value.HasLongitudeDeg, LongitudeDeg: value.LongitudeDeg, HasNMEARequired: value.HasNMEARequired, NMEARequired: value.NMEARequired, HasNetworkSolution: value.HasNetworkSolution, NetworkSolution: value.NetworkSolution, Generator: value.Generator, Compression: value.Compression, Authentication: NTRIPSourcetableAuth(value.Authentication), HasFee: value.HasFee, Fee: value.Fee, HasBitrate: value.HasBitrate, Bitrate: value.Bitrate, Misc: value.Misc}
	}
	return result, nil
}

// Text serializes the sourcetable through C and may run concurrently with
// Summary, Streams, and other Text calls.
func (t *NTRIPSourcetable) Text() ([]byte, error) {
	if t == nil || t.handle == nil {
		return nil, ErrClosed
	}
	data, err := t.handle.Text()
	return data, publicError(err)
}

// NTRIPDialer is injectable for deterministic transport tests.
type NTRIPDialer func(context.Context, string, string) (net.Conn, error)

// NTRIPClient owns socket/TLS setup and forwards all received bytes to C.
// Dialer, TLSConfig, Timeout, and ReadSize are transport policy; protocol
// bytes and state transitions remain in NTRIPMachine. Timeout may be zero for
// no transport deadline but must not be negative. Negative ReadSize values and
// values above maxNTRIPReadSize (16 MiB) are invalid; zero selects the safe
// defaultNTRIPReadSize value, and positive values up to the maximum are accepted.
type NTRIPClient struct {
	// Config is the config value for NTRIPClient.
	Config NTRIPConfig
	// TLS is the tls value for NTRIPClient.
	TLS bool
	// TLSConfig contains a detached copy; nil means this field is absent.
	TLSConfig *tls.Config
	// Dialer is the dialer value for NTRIPClient.
	Dialer NTRIPDialer
	// Timeout is the timestamp for this record.
	Timeout time.Duration
	// ReadSize is the read size value for NTRIPClient.
	ReadSize int
}

const (
	defaultNTRIPReadSize = 32 << 10

	// maxNTRIPReadSize prevents an option typo from causing a single enormous
	// allocation for every network read.
	maxNTRIPReadSize = 16 << 20
)

type ntripConnectionState struct {
	mu        sync.Mutex
	opMu      sync.Mutex
	conn      net.Conn
	machine   ntripMachine
	closed    bool
	closeErr  error
	closeDone chan struct{}
	timeout   time.Duration
	readSize  int
}

// NTRIPConnection owns a network connection and mutable C machine. Read and
// SendGGAMessage are serialized because they mutate the machine. Close first
// interrupts socket I/O, then waits for active operations and releases the C
// machine. Values must not be copied after first use.
type NTRIPConnection struct {
	_       noCopy
	state   *ntripConnectionState
	cleanup runtime.Cleanup
}

type ntripMachine interface {
	Push([]byte) ([]NTRIPEvent, error)
	Finish() ([]NTRIPEvent, error)
	TryGGAMessage(float64, float64, NTRIPGGAPosition) ([]byte, bool, error)
	Close() error
}

func newNTRIPConnection(conn net.Conn, machine ntripMachine, timeout time.Duration, readSize int) *NTRIPConnection {
	state := &ntripConnectionState{conn: conn, machine: machine, timeout: timeout, readSize: readSize, closeDone: make(chan struct{})}
	owner := &NTRIPConnection{state: state}
	owner.cleanup = runtime.AddCleanup(owner, cleanupNTRIPConnection, state)
	return owner
}

func cleanupNTRIPConnection(state *ntripConnectionState) {
	_ = state.close()
}

func (c *NTRIPClient) validate() (int, error) {
	if c == nil {
		return 0, errors.New("sidereon: nil NTRIP client")
	}
	if c.Timeout < 0 {
		return 0, errors.New("sidereon: NTRIP timeout must not be negative")
	}
	readSize := c.ReadSize
	if readSize == 0 {
		readSize = defaultNTRIPReadSize
	}
	if readSize < 0 {
		return 0, errors.New("sidereon: NTRIP read size must not be negative")
	}
	if readSize > maxNTRIPReadSize {
		return 0, fmt.Errorf("sidereon: NTRIP read size must not exceed %d", maxNTRIPReadSize)
	}
	return readSize, nil
}

// Connect dials, optionally performs TLS, creates the C machine, and writes
// its exact connection request before returning.
func (c *NTRIPClient) Connect(ctx context.Context) (*NTRIPConnection, error) {
	if ctx == nil {
		return nil, errors.New("sidereon: nil context")
	}
	readSize, err := c.validate()
	if err != nil {
		return nil, err
	}
	if err := validateNTRIPConfig(c.Config); err != nil {
		return nil, err
	}
	dialer := c.Dialer
	if dialer == nil {
		d := &net.Dialer{}
		dialer = d.DialContext
	}
	port := c.Config.Port
	if port == 0 {
		port = 2101
	}
	dialCtx := ctx
	var cancel context.CancelFunc
	if c.Timeout > 0 {
		dialCtx, cancel = context.WithTimeout(ctx, c.Timeout)
		defer cancel()
	}
	conn, err := dialer(dialCtx, "tcp", net.JoinHostPort(c.Config.Host, fmt.Sprint(port)))
	if err != nil {
		if ctx.Err() != nil {
			return nil, errors.Join(ctx.Err(), err, closeConn(conn))
		}
		return nil, errors.Join(err, closeConn(conn))
	}
	if conn == nil {
		return nil, errors.New("sidereon: NTRIP dialer returned a nil connection")
	}
	if c.TLS {
		config := c.TLSConfig
		if config == nil {
			config = &tls.Config{}
		} else {
			config = config.Clone()
		}
		if config.MinVersion < tls.VersionTLS12 {
			config.MinVersion = tls.VersionTLS12
		}
		if config.ServerName == "" {
			config.ServerName = c.Config.Host
		}
		tlsConn := tls.Client(conn, config)
		if err := tlsConn.HandshakeContext(dialCtx); err != nil {
			closeErr := conn.Close()
			if ctx.Err() != nil {
				return nil, errors.Join(ctx.Err(), closeErr)
			}
			return nil, errors.Join(err, closeErr)
		}
		conn = tlsConn
	}
	machine, err := NewNTRIPMachine(c.Config)
	if err != nil {
		return nil, errors.Join(err, conn.Close())
	}
	request, err := machine.ConnectionRequest()
	if err != nil {
		return nil, errors.Join(err, machine.Close(), conn.Close())
	}
	if err := writeWithContext(ctx, conn, request, c.Timeout); err != nil {
		return nil, errors.Join(err, machine.Close(), conn.Close())
	}
	return newNTRIPConnection(conn, machine, c.Timeout, readSize), nil
}

// Read reads one network chunk, feeds it to C, and returns copied events. EOF
// is converted into C Finish events and returned alongside io.EOF.
func (c *NTRIPConnection) Read(ctx context.Context) ([]NTRIPEvent, error) {
	if c == nil {
		return nil, ErrClosed
	}
	state := c.state
	if state == nil {
		return nil, ErrClosed
	}
	if ctx == nil {
		return nil, errors.New("sidereon: nil context")
	}
	state.opMu.Lock()
	defer state.opMu.Unlock()
	state.mu.Lock()
	if state.closed || state.conn == nil {
		state.mu.Unlock()
		return nil, ErrClosed
	}
	conn := state.conn
	machine := state.machine
	timeout := state.timeout
	readSize := state.readSize
	state.mu.Unlock()
	if machine == nil {
		return nil, ErrClosed
	}
	buffer := make([]byte, readSize)
	n, err := readWithContext(ctx, conn, buffer, timeout)
	transportErr := withoutEOF(err)
	if n > 0 {
		events, pushErr := machine.Push(buffer[:n])
		if pushErr != nil {
			return nil, ntripJoinErrors(ntripMachineError("push", pushErr), transportErr, closeUndeliveredNTRIPEvents(events, 0))
		}
		if errors.Is(err, io.EOF) {
			return finishNTRIPRead(machine, events, transportErr)
		}
		return events, transportErr
	}
	if errors.Is(err, io.EOF) {
		return finishNTRIPRead(machine, nil, transportErr)
	}
	state.mu.Lock()
	closed := state.closed
	state.mu.Unlock()
	if closed {
		return nil, ErrClosed
	}
	return nil, transportErr
}

func finishNTRIPRead(machine ntripMachine, events []NTRIPEvent, transportErr error) ([]NTRIPEvent, error) {
	finished, finishErr := machine.Finish()
	if finishErr != nil {
		return events, ntripJoinErrors(ntripMachineError("finish", finishErr), transportErr, closeUndeliveredNTRIPEvents(finished, 0))
	}
	if transportErr != nil {
		return append(events, finished...), transportErr
	}
	return append(events, finished...), io.EOF
}

func ntripJoinErrors(values ...error) error {
	var nonNil []error
	for _, value := range values {
		if value != nil {
			nonNil = append(nonNil, value)
		}
	}
	switch len(nonNil) {
	case 0:
		return nil
	case 1:
		return nonNil[0]
	default:
		return errors.Join(nonNil...)
	}
}

func withoutEOF(err error) error {
	if err == nil || !errors.Is(err, io.EOF) {
		return err
	}
	return withoutEOFNode(err)
}

func ntripMachineError(operation string, err error) error {
	if cleaned := withoutEOF(err); cleaned != nil {
		return cleaned
	}
	return fmt.Errorf("sidereon: NTRIP machine %s failed", operation)
}

func withoutEOFNode(err error) error {
	if err == nil {
		return nil
	}
	if !errors.Is(err, io.EOF) {
		return err
	}
	if unwrapper, ok := err.(interface{ Unwrap() []error }); ok {
		children := make([]error, 0, len(unwrapper.Unwrap()))
		for _, child := range unwrapper.Unwrap() {
			if child = withoutEOFNode(child); child != nil {
				children = append(children, child)
			}
		}
		return ntripJoinErrors(children...)
	}
	if unwrapper, ok := err.(interface{ Unwrap() error }); ok {
		return withoutEOFNode(unwrapper.Unwrap())
	}
	return nil
}

func closeConn(conn net.Conn) error {
	if conn == nil {
		return nil
	}
	return conn.Close()
}

// SendGGAMessage asks C to generate and, when due, writes one exact GGA
// sentence. nowS is in seconds on the caller's chosen monotonic scale.
func (c *NTRIPConnection) SendGGAMessage(ctx context.Context, nowS, utcSecondsOfDay float64, position NTRIPGGAPosition) (bool, error) {
	if c == nil {
		return false, ErrClosed
	}
	state := c.state
	if state == nil {
		return false, ErrClosed
	}
	if ctx == nil {
		return false, errors.New("sidereon: nil context")
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}
	state.opMu.Lock()
	defer state.opMu.Unlock()
	state.mu.Lock()
	if state.closed || state.conn == nil {
		state.mu.Unlock()
		return false, ErrClosed
	}
	conn := state.conn
	machine := state.machine
	timeout := state.timeout
	state.mu.Unlock()
	if machine == nil {
		return false, ErrClosed
	}
	data, present, err := machine.TryGGAMessage(nowS, utcSecondsOfDay, position)
	if err != nil || !present {
		return present, err
	}
	return true, writeWithContext(ctx, conn, data, timeout)
}

// Close closes the socket and machine once. It is safe to call repeatedly.
func (c *NTRIPConnection) Close() error {
	if c == nil {
		return nil
	}
	c.cleanup.Stop()
	if c.state == nil {
		return nil
	}
	return c.state.close()
}

func (c *ntripConnectionState) close() error {
	c.mu.Lock()
	if c.closeDone == nil {
		c.closeDone = make(chan struct{})
	}
	if c.closed {
		done := c.closeDone
		c.mu.Unlock()
		<-done
		c.mu.Lock()
		err := c.closeErr
		c.mu.Unlock()
		return err
	}
	c.closed = true
	conn := c.conn
	c.conn = nil
	done := c.closeDone
	c.mu.Unlock()
	var closeErr error
	if conn != nil {
		closeErr = conn.Close()
	}
	c.opMu.Lock()
	c.mu.Lock()
	var machine ntripMachine
	if c.machine != nil {
		machine = c.machine
		c.machine = nil
	}
	c.mu.Unlock()
	if machine != nil {
		closeErr = errors.Join(closeErr, machine.Close())
	}
	c.opMu.Unlock()
	c.mu.Lock()
	c.closeErr = closeErr
	close(done)
	c.mu.Unlock()
	return closeErr
}

// Run connects and delivers copied events until the context ends, the peer
// closes, or callback returns an error.
func (c *NTRIPClient) Run(ctx context.Context, callback func(NTRIPEvent) error) (runErr error) {
	if ctx == nil {
		return errors.New("sidereon: nil context")
	}
	if callback == nil {
		return errors.New("sidereon: nil NTRIP event callback")
	}
	connection, err := c.Connect(ctx)
	if err != nil {
		return err
	}
	return c.runConnection(ctx, callback, connection)
}

func (c *NTRIPClient) runConnection(ctx context.Context, callback func(NTRIPEvent) error, connection *NTRIPConnection) (runErr error) {
	defer func() { runErr = errors.Join(runErr, connection.Close()) }()
	for {
		events, readErr := connection.Read(ctx)
		for index, event := range events {
			if callbackErr := callback(event); callbackErr != nil {
				return errors.Join(callbackErr, closeUndeliveredNTRIPEvents(events, index+1))
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				return nil
			}
			return readErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
	}
}

func closeUndeliveredNTRIPEvents(events []NTRIPEvent, start int) (err error) {
	for i := start; i < len(events); i++ {
		if events[i].Sourcetable != nil {
			err = errors.Join(err, events[i].Sourcetable.Close())
			events[i].Sourcetable = nil
		}
	}
	return err
}

func writeWithContext(ctx context.Context, conn net.Conn, data []byte, timeout time.Duration) (err error) {
	if ctx == nil {
		return errors.New("sidereon: nil context")
	}
	if conn == nil {
		return errors.New("sidereon: nil NTRIP connection")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	deadline := time.Time{}
	if timeout > 0 {
		deadline = time.Now().Add(timeout)
	}
	if ctxDeadline, ok := ctx.Deadline(); ok && (deadline.IsZero() || ctxDeadline.Before(deadline)) {
		deadline = ctxDeadline
	}
	if !deadline.IsZero() {
		if err := conn.SetWriteDeadline(deadline); err != nil {
			return err
		}
	}
	stop := interruptConnOnCancel(ctx, conn, true)
	defer func() {
		cleanupErr := stop()
		if !deadline.IsZero() {
			cleanupErr = errors.Join(cleanupErr, conn.SetWriteDeadline(time.Time{}))
		}
		err = errors.Join(err, cleanupErr)
	}()
	for len(data) != 0 {
		n, writeErr := conn.Write(data)
		if n < 0 || n > len(data) {
			return io.ErrShortWrite
		}
		if writeErr != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return writeErr
		}
		if n == 0 {
			return io.ErrNoProgress
		}
		data = data[n:]
	}
	runtime.KeepAlive(conn)
	return nil
}

func readWithContext(ctx context.Context, conn net.Conn, buffer []byte, timeout time.Duration) (n int, err error) {
	if ctx == nil {
		return 0, errors.New("sidereon: nil context")
	}
	if conn == nil {
		return 0, errors.New("sidereon: nil NTRIP connection")
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	deadline := time.Time{}
	if timeout > 0 {
		deadline = time.Now().Add(timeout)
	}
	if ctxDeadline, ok := ctx.Deadline(); ok && (deadline.IsZero() || ctxDeadline.Before(deadline)) {
		deadline = ctxDeadline
	}
	if !deadline.IsZero() {
		if err := conn.SetReadDeadline(deadline); err != nil {
			return 0, err
		}
	}
	stop := interruptConnOnCancel(ctx, conn, false)
	defer func() {
		cleanupErr := stop()
		if !deadline.IsZero() {
			cleanupErr = errors.Join(cleanupErr, conn.SetReadDeadline(time.Time{}))
		}
		err = errors.Join(err, cleanupErr)
	}()
	n, err = conn.Read(buffer)
	if n < 0 || n > len(buffer) {
		return 0, io.ErrShortWrite
	}
	if err != nil && ctx.Err() != nil {
		return n, ctx.Err()
	}
	if n == 0 && err == nil {
		return 0, io.ErrNoProgress
	}
	runtime.KeepAlive(conn)
	return n, err
}

func interruptConnOnCancel(ctx context.Context, conn net.Conn, write bool) func() error {
	if ctx.Done() == nil {
		return func() error { return nil }
	}
	done := make(chan struct{})
	type interruptResult struct {
		err            error
		deadlineForced bool
	}
	stopped := make(chan interruptResult, 1)
	setForcedDeadline := func() interruptResult {
		var err error
		if write {
			err = conn.SetWriteDeadline(time.Now())
		} else {
			err = conn.SetReadDeadline(time.Now())
		}
		return interruptResult{err: err, deadlineForced: err == nil}
	}
	go func() {
		select {
		case <-ctx.Done():
			stopped <- setForcedDeadline()
		case <-done:
			select {
			case <-ctx.Done():
				stopped <- setForcedDeadline()
			default:
				stopped <- interruptResult{}
			}
		}
	}()
	var stopOnce sync.Once
	var stopResult interruptResult
	return func() error {
		stopOnce.Do(func() {
			close(done)
			stopResult = <-stopped
			if stopResult.deadlineForced {
				var clearErr error
				if write {
					clearErr = conn.SetWriteDeadline(time.Time{})
				} else {
					clearErr = conn.SetReadDeadline(time.Time{})
				}
				stopResult.err = errors.Join(stopResult.err, clearErr)
			}
		})
		return stopResult.err
	}
}
