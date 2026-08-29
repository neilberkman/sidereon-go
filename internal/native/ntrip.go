//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"fmt"
	"runtime"
	"unsafe"
)

// NTRIPVersion selects the wire request version.
type NTRIPVersion uint32

const (
	NTRIPVersionRev1 NTRIPVersion = 1
	NTRIPVersionRev2 NTRIPVersion = 2
)

// NTRIPState is the C sans-I/O machine state.
type NTRIPState uint32

const (
	NTRIPStateIdle            NTRIPState = 0
	NTRIPStateAwaitingStatus  NTRIPState = 1
	NTRIPStateAwaitingHeaders NTRIPState = 2
	NTRIPStateStreaming       NTRIPState = 3
	NTRIPStateSourcetable     NTRIPState = 4
	NTRIPStateClosed          NTRIPState = 5
)

// NTRIPEventKind identifies a copied event from the C machine.
type NTRIPEventKind uint32

const (
	NTRIPEventConnected       NTRIPEventKind = 0
	NTRIPEventPayload         NTRIPEventKind = 1
	NTRIPEventSourcetable     NTRIPEventKind = 2
	NTRIPEventRejected        NTRIPEventKind = 3
	NTRIPEventStreamCorrupted NTRIPEventKind = 4
	NTRIPEventStreamEnded     NTRIPEventKind = 5
)

// NTRIPRejectionKind identifies a C protocol rejection.
type NTRIPRejectionKind uint32

const (
	NTRIPRejectionNone                  NTRIPRejectionKind = 0
	NTRIPRejectionUnauthorized          NTRIPRejectionKind = 1
	NTRIPRejectionMountpointNotFound    NTRIPRejectionKind = 2
	NTRIPRejectionDigestRequired        NTRIPRejectionKind = 3
	NTRIPRejectionCasterError           NTRIPRejectionKind = 4
	NTRIPRejectionUnexpectedContentType NTRIPRejectionKind = 5
	NTRIPRejectionHTTPError             NTRIPRejectionKind = 6
	NTRIPRejectionMalformedHandshake    NTRIPRejectionKind = 7
)

// NTRIPSourcetableAuth identifies the authentication field of an STR record.
type NTRIPSourcetableAuth uint32

const (
	NTRIPSourcetableAuthNone   NTRIPSourcetableAuth = 0
	NTRIPSourcetableAuthBasic  NTRIPSourcetableAuth = 1
	NTRIPSourcetableAuthDigest NTRIPSourcetableAuth = 2
	NTRIPSourcetableAuthOther  NTRIPSourcetableAuth = 3
)

// NTRIPConfig is copied into C for each call. GGAIntervalS is in seconds;
// HasGGAInterval distinguishes an omitted interval from zero.
type NTRIPConfig struct {
	Host           string
	Port           uint16
	Mountpoint     string
	Version        NTRIPVersion
	Username       string
	Password       string
	HasCredentials bool
	UserAgent      string
	GGAIntervalS   float64
	HasGGAInterval bool
}

// NTRIPGGAPosition contains decimal-degree coordinates and metres.
type NTRIPGGAPosition struct {
	LatitudeDeg  float64
	LongitudeDeg float64
	HeightM      float64
	FixQuality   uint8
	Satellites   uint8
	HDOP         float64
}

// NTRIPEventInfo is the scalar portion of one C event.
type NTRIPEventInfo struct {
	Kind               NTRIPEventKind
	Version            NTRIPVersion
	Chunked            bool
	HeaderCount        int
	PayloadLength      int
	SourcetableRecords int
	Rejection          NTRIPRejectionKind
	HTTPStatus         uint16
}

// NTRIPStreamInfo is a copied typed STR record. Optional values use the Has*
// flags, matching C's explicit optional-field representation.
type NTRIPStreamInfo struct {
	Mountpoint         string
	Identifier         string
	Format             string
	FormatDetails      string
	HasCarrier         bool
	Carrier            uint8
	NavSystem          string
	Network            string
	Country            string
	HasLatitudeDeg     bool
	LatitudeDeg        float64
	HasLongitudeDeg    bool
	LongitudeDeg       float64
	HasNMEARequired    bool
	NMEARequired       bool
	HasNetworkSolution bool
	NetworkSolution    bool
	Generator          string
	Compression        string
	Authentication     NTRIPSourcetableAuth
	HasFee             bool
	Fee                bool
	HasBitrate         bool
	Bitrate            uint32
	Misc               string
}

// NTRIPMachine owns a mutable C sans-I/O state machine. Its operations are
// serialized. Close waits for an active operation, clears the native resource,
// and is idempotent. Values must not be copied after first use.
type NTRIPMachine struct {
	_       noCopy
	handle  *ntripHandle
	cleanup runtime.Cleanup
}

// NTRIPBytes owns one C-produced byte sequence until Close.
type NTRIPBytes struct {
	_       noCopy
	handle  *ntripHandle
	cleanup runtime.Cleanup
}

// NTRIPEvents owns one C-produced event batch until Close.
type NTRIPEvents struct {
	_       noCopy
	handle  *ntripHandle
	cleanup runtime.Cleanup
}

// NTRIPSourcetable owns one read-only C-parsed sourcetable until Close.
// Summary, Streams, and Text may run concurrently. Close waits for active
// reads, clears the native resource, and is idempotent. Values must not be
// copied after first use.
type NTRIPSourcetable struct {
	_       noCopy
	handle  *ntripHandle
	cleanup runtime.Cleanup
}

func nativeNTRIPConfig(config NTRIPConfig, fn func(*C.SidereonNtripConfig) error) error {
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
		if containsNUL(field.value) {
			return fmt.Errorf("sidereon: %s contains an embedded NUL byte", field.name)
		}
	}
	var callErr error
	withCThread(func() {
		host, err := cString(config.Host)
		if err != nil {
			callErr = err
			return
		}
		defer freeCString(host)
		mountpoint, err := cString(config.Mountpoint)
		if err != nil {
			callErr = err
			return
		}
		defer freeCString(mountpoint)
		var userAgent *C.char
		if config.UserAgent != "" {
			userAgent, err = cString(config.UserAgent)
			if err != nil {
				callErr = err
				return
			}
			defer freeCString(userAgent)
		}
		var username, password *C.char
		if config.HasCredentials {
			username, err = cString(config.Username)
			if err != nil {
				callErr = err
				return
			}
			defer freeCString(username)
			password, err = cString(config.Password)
			if err != nil {
				callErr = err
				return
			}
			defer freeCString(password)
		}
		value := C.SidereonNtripConfig{
			host:               host,
			port:               C.uint16_t(config.Port),
			mountpoint:         mountpoint,
			version:            C.uint32_t(config.Version),
			has_credentials:    C.bool(config.HasCredentials),
			username:           username,
			password:           password,
			user_agent_product: userAgent,
			has_gga_interval_s: C.bool(config.HasGGAInterval),
			gga_interval_s:     C.double(config.GGAIntervalS),
		}
		callErr = fn(&value)
	})
	return callErr
}

func newNTRIPHandle(pointer unsafe.Pointer, release func(unsafe.Pointer)) (*ntripHandle, error) {
	if pointer == nil {
		return nil, errors.New("sidereon: native NTRIP constructor returned a nil handle")
	}
	return &ntripHandle{ptr: pointer, release: release}, nil
}

func releaseNTRIPMachine(pointer unsafe.Pointer) {
	C.sidereon_ntrip_machine_free((*C.SidereonNtripMachine)(pointer))
}

func releaseNTRIPBytes(pointer unsafe.Pointer) {
	C.sidereon_ntrip_bytes_free((*C.SidereonNtripBytes)(pointer))
}

func releaseNTRIPEvents(pointer unsafe.Pointer) {
	C.sidereon_ntrip_events_free((*C.SidereonNtripEvents)(pointer))
}

func releaseNTRIPSourcetable(pointer unsafe.Pointer) {
	C.sidereon_ntrip_sourcetable_free((*C.SidereonNtripSourcetable)(pointer))
}

func releaseNTRIPOnError(pointer unsafe.Pointer, release func(unsafe.Pointer)) {
	if pointer != nil {
		withCThread(func() { release(pointer) })
	}
}

func newNTRIPMachine(pointer *C.SidereonNtripMachine) (*NTRIPMachine, error) {
	handle, err := newNTRIPHandle(unsafe.Pointer(pointer), releaseNTRIPMachine)
	if err != nil {
		return nil, err
	}
	owner := &NTRIPMachine{handle: handle}
	owner.cleanup = addNTRIPCleanup(owner, handle)
	return owner, nil
}

func newNTRIPBytes(pointer *C.SidereonNtripBytes) (*NTRIPBytes, error) {
	handle, err := newNTRIPHandle(unsafe.Pointer(pointer), releaseNTRIPBytes)
	if err != nil {
		return nil, err
	}
	owner := &NTRIPBytes{handle: handle}
	owner.cleanup = addNTRIPCleanup(owner, handle)
	return owner, nil
}

func newNTRIPEvents(pointer *C.SidereonNtripEvents) (*NTRIPEvents, error) {
	handle, err := newNTRIPHandle(unsafe.Pointer(pointer), releaseNTRIPEvents)
	if err != nil {
		return nil, err
	}
	owner := &NTRIPEvents{handle: handle}
	owner.cleanup = addNTRIPCleanup(owner, handle)
	return owner, nil
}

func newNTRIPSourcetable(pointer *C.SidereonNtripSourcetable) (*NTRIPSourcetable, error) {
	handle, err := newNTRIPHandle(unsafe.Pointer(pointer), releaseNTRIPSourcetable)
	if err != nil {
		return nil, err
	}
	owner := &NTRIPSourcetable{handle: handle}
	owner.cleanup = addNTRIPCleanup(owner, handle)
	return owner, nil
}

func (m *NTRIPMachine) Close() {
	if m == nil || m.handle == nil {
		return
	}
	m.cleanup.Stop()
	m.handle.close()
}

func (b *NTRIPBytes) Close() {
	if b == nil || b.handle == nil {
		return
	}
	b.cleanup.Stop()
	b.handle.close()
}

func (e *NTRIPEvents) Close() {
	if e == nil || e.handle == nil {
		return
	}
	e.cleanup.Stop()
	e.handle.close()
}

func (t *NTRIPSourcetable) Close() {
	if t == nil || t.handle == nil {
		return
	}
	t.cleanup.Stop()
	t.handle.close()
}

func NewNTRIPMachine(config NTRIPConfig) (*NTRIPMachine, error) {
	var result *NTRIPMachine
	err := nativeNTRIPConfig(config, func(value *C.SidereonNtripConfig) error {
		var pointer *C.SidereonNtripMachine
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_ntrip_machine_new(value, &pointer))
		})
		if err != nil {
			releaseNTRIPOnError(unsafe.Pointer(pointer), releaseNTRIPMachine)
			return err
		}
		result, err = newNTRIPMachine(pointer)
		return err
	})
	return result, err
}

func NTRIPRequestBytes(config NTRIPConfig) ([]byte, error) {
	var result []byte
	err := nativeNTRIPConfig(config, func(value *C.SidereonNtripConfig) error {
		var err error
		result, err = copyNativeBytesLocked("NTRIP request", func(out *C.uint8_t, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_ntrip_request_bytes(value, out, length, written, required)
		})
		return err
	})
	return result, err
}

func (m *NTRIPMachine) ConnectionRequest() (*NTRIPBytes, error) {
	if m == nil || m.handle == nil {
		return nil, ErrClosed
	}
	var result *NTRIPBytes
	err := m.handle.with(func(pointer unsafe.Pointer) error {
		var bytes *C.SidereonNtripBytes
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_ntrip_machine_connection_request((*C.SidereonNtripMachine)(pointer), &bytes))
		})
		if err != nil {
			releaseNTRIPOnError(unsafe.Pointer(bytes), releaseNTRIPBytes)
			return err
		}
		result, err = newNTRIPBytes(bytes)
		return err
	})
	runtime.KeepAlive(m)
	return result, err
}

func (m *NTRIPMachine) Push(data []byte) (*NTRIPEvents, error) {
	if m == nil || m.handle == nil {
		return nil, ErrClosed
	}
	var result *NTRIPEvents
	err := m.handle.with(func(pointer unsafe.Pointer) error {
		cdata, copyErr := copyNativeInput(data)
		if copyErr != nil {
			return copyErr
		}
		defer freeNativeInput(cdata)
		var events *C.SidereonNtripEvents
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_ntrip_machine_push((*C.SidereonNtripMachine)(pointer), (*C.uint8_t)(cdata), C.size_t(len(data)), &events))
		})
		if err != nil {
			releaseNTRIPOnError(unsafe.Pointer(events), releaseNTRIPEvents)
			return err
		}
		result, err = newNTRIPEvents(events)
		return err
	})
	runtime.KeepAlive(data)
	runtime.KeepAlive(m)
	return result, err
}

func (m *NTRIPMachine) Finish() (*NTRIPEvents, error) {
	if m == nil || m.handle == nil {
		return nil, ErrClosed
	}
	var result *NTRIPEvents
	err := m.handle.with(func(pointer unsafe.Pointer) error {
		var events *C.SidereonNtripEvents
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_ntrip_machine_finish((*C.SidereonNtripMachine)(pointer), &events))
		})
		if err != nil {
			releaseNTRIPOnError(unsafe.Pointer(events), releaseNTRIPEvents)
			return err
		}
		result, err = newNTRIPEvents(events)
		return err
	})
	runtime.KeepAlive(m)
	return result, err
}

func (m *NTRIPMachine) Reset() error {
	if m == nil || m.handle == nil {
		return ErrClosed
	}
	err := m.handle.with(func(pointer unsafe.Pointer) error {
		withCThread(func() { C.sidereon_ntrip_machine_reset((*C.SidereonNtripMachine)(pointer)) })
		return nil
	})
	runtime.KeepAlive(m)
	return err
}

func (m *NTRIPMachine) State() (NTRIPState, error) {
	if m == nil || m.handle == nil {
		return 0, ErrClosed
	}
	var result NTRIPState
	err := m.handle.with(func(pointer unsafe.Pointer) error {
		var value C.uint32_t
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_ntrip_machine_state((*C.SidereonNtripMachine)(pointer), &value))
		})
		result = NTRIPState(value)
		return err
	})
	runtime.KeepAlive(m)
	return result, err
}

func (m *NTRIPMachine) TryGGAMessage(nowS float64, position NTRIPGGAPosition, utcSecondsOfDay float64) (*NTRIPBytes, bool, error) {
	if m == nil || m.handle == nil {
		return nil, false, ErrClosed
	}
	var result *NTRIPBytes
	var present bool
	err := m.handle.with(func(pointer unsafe.Pointer) error {
		input := C.SidereonNtripGgaPosition{
			lat_deg:        C.double(position.LatitudeDeg),
			lon_deg:        C.double(position.LongitudeDeg),
			height_m:       C.double(position.HeightM),
			fix_quality:    C.uint8_t(position.FixQuality),
			num_satellites: C.uint8_t(position.Satellites),
			hdop:           C.double(position.HDOP),
		}
		var outputPresent C.bool
		var bytes *C.SidereonNtripBytes
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_ntrip_machine_try_gga_message((*C.SidereonNtripMachine)(pointer), C.double(nowS), &input, C.double(utcSecondsOfDay), &outputPresent, &bytes))
		})
		present = bool(outputPresent)
		if err != nil {
			releaseNTRIPOnError(unsafe.Pointer(bytes), releaseNTRIPBytes)
			return err
		}
		if present {
			result, err = newNTRIPBytes(bytes)
			return err
		}
		if bytes != nil {
			releaseNTRIPOnError(unsafe.Pointer(bytes), releaseNTRIPBytes)
			return errors.New("sidereon: NTRIP GGA returned bytes without a message")
		}
		return nil
	})
	runtime.KeepAlive(m)
	return result, present, err
}

func (b *NTRIPBytes) Bytes() ([]byte, error) {
	if b == nil || b.handle == nil {
		return nil, ErrClosed
	}
	var result []byte
	err := b.handle.with(func(pointer unsafe.Pointer) error {
		var callErr error
		result, callErr = copyNativeBytes("NTRIP bytes", func(out *C.uint8_t, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_ntrip_bytes((*C.SidereonNtripBytes)(pointer), out, length, written, required)
		})
		return callErr
	})
	runtime.KeepAlive(b)
	return result, err
}

func (e *NTRIPEvents) Count() (int, error) {
	if e == nil || e.handle == nil {
		return 0, ErrClosed
	}
	var result C.size_t
	err := e.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_ntrip_events_count((*C.SidereonNtripEvents)(pointer), &result))
		})
	})
	count, countErr := sizeTToInt(result, "NTRIP event count")
	if err == nil && countErr != nil {
		err = countErr
	}
	runtime.KeepAlive(e)
	return count, err
}

func validateNTRIPIndex(index int) error {
	if index < 0 {
		return errors.New("sidereon: negative NTRIP event index")
	}
	return nil
}

func (e *NTRIPEvents) Event(index int) (NTRIPEventInfo, error) {
	if e == nil || e.handle == nil {
		return NTRIPEventInfo{}, ErrClosed
	}
	if err := validateNTRIPIndex(index); err != nil {
		return NTRIPEventInfo{}, err
	}
	var value C.SidereonNtripEventInfo
	err := e.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_ntrip_events_event((*C.SidereonNtripEvents)(pointer), C.size_t(index), &value))
		})
	})
	var result NTRIPEventInfo
	if err == nil {
		headerCount, headerErr := sizeTToInt(value.header_count, "NTRIP event header count")
		payloadLength, payloadErr := sizeTToInt(value.payload_len, "NTRIP event payload length")
		sourcetableRecords, sourcetableErr := sizeTToInt(value.sourcetable_record_count, "NTRIP event sourcetable record count")
		if headerErr != nil || payloadErr != nil || sourcetableErr != nil {
			err = errors.Join(headerErr, payloadErr, sourcetableErr)
		} else {
			result = NTRIPEventInfo{Kind: NTRIPEventKind(value.kind), Version: NTRIPVersion(value.version), Chunked: bool(value.chunked), HeaderCount: headerCount, PayloadLength: payloadLength, SourcetableRecords: sourcetableRecords, Rejection: NTRIPRejectionKind(value.rejection), HTTPStatus: uint16(value.http_status)}
		}
	}
	runtime.KeepAlive(e)
	return result, err
}

func (e *NTRIPEvents) Payload(index int) ([]byte, error) {
	if e == nil || e.handle == nil {
		return nil, ErrClosed
	}
	if err := validateNTRIPIndex(index); err != nil {
		return nil, err
	}
	var result []byte
	err := e.handle.with(func(pointer unsafe.Pointer) error {
		var callErr error
		result, callErr = copyNativeBytes("NTRIP event payload", func(out *C.uint8_t, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_ntrip_events_payload((*C.SidereonNtripEvents)(pointer), C.size_t(index), out, length, written, required)
		})
		return callErr
	})
	runtime.KeepAlive(e)
	return result, err
}

func (e *NTRIPEvents) Detail(index int) ([]byte, error) {
	if e == nil || e.handle == nil {
		return nil, ErrClosed
	}
	if err := validateNTRIPIndex(index); err != nil {
		return nil, err
	}
	var result []byte
	err := e.handle.with(func(pointer unsafe.Pointer) error {
		var callErr error
		result, callErr = copyNativeBytes("NTRIP event detail", func(out *C.uint8_t, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_ntrip_events_detail((*C.SidereonNtripEvents)(pointer), C.size_t(index), out, length, written, required)
		})
		return callErr
	})
	runtime.KeepAlive(e)
	return result, err
}

func (e *NTRIPEvents) Sourcetable(index int) (*NTRIPSourcetable, error) {
	if e == nil || e.handle == nil {
		return nil, ErrClosed
	}
	if err := validateNTRIPIndex(index); err != nil {
		return nil, err
	}
	var result *NTRIPSourcetable
	err := e.handle.with(func(pointer unsafe.Pointer) error {
		var table *C.SidereonNtripSourcetable
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_ntrip_events_sourcetable((*C.SidereonNtripEvents)(pointer), C.size_t(index), &table))
		})
		if err != nil {
			releaseNTRIPOnError(unsafe.Pointer(table), releaseNTRIPSourcetable)
			return err
		}
		result, err = newNTRIPSourcetable(table)
		return err
	})
	runtime.KeepAlive(e)
	return result, err
}

func ParseNTRIPSourcetable(data []byte) (*NTRIPSourcetable, error) {
	var result *NTRIPSourcetable
	var err error
	withCThread(func() {
		cdata, copyErr := copyNativeInput(data)
		if copyErr != nil {
			err = copyErr
			return
		}
		defer freeNativeInput(cdata)
		var table *C.SidereonNtripSourcetable
		err = callStatus(func() uint32 {
			return uint32(C.sidereon_ntrip_sourcetable_parse((*C.uint8_t)(cdata), C.size_t(len(data)), &table))
		})
		if err != nil {
			releaseNTRIPOnError(unsafe.Pointer(table), releaseNTRIPSourcetable)
			return
		}
		result, err = newNTRIPSourcetable(table)
	})
	runtime.KeepAlive(data)
	return result, err
}

func (t *NTRIPSourcetable) Summary() (int, int, error) {
	if t == nil || t.handle == nil {
		return 0, 0, ErrClosed
	}
	var value C.SidereonNtripSourcetableSummary
	err := t.handle.read(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_ntrip_sourcetable_summary((*C.SidereonNtripSourcetable)(pointer), &value))
		})
	})
	recordCount, recordErr := sizeTToInt(value.record_count, "NTRIP sourcetable record count")
	streamCount, streamErr := sizeTToInt(value.stream_count, "NTRIP sourcetable stream count")
	if err == nil {
		err = errors.Join(recordErr, streamErr)
	}
	runtime.KeepAlive(t)
	return recordCount, streamCount, err
}

func streamFromC(value *C.SidereonNtripStreamInfo) NTRIPStreamInfo {
	return NTRIPStreamInfo{
		Mountpoint: fixedCString((*C.char)(unsafe.Pointer(&value.mountpoint[0]))), Identifier: fixedCString((*C.char)(unsafe.Pointer(&value.identifier[0]))), Format: fixedCString((*C.char)(unsafe.Pointer(&value.format[0]))), FormatDetails: fixedCString((*C.char)(unsafe.Pointer(&value.format_details[0]))), HasCarrier: bool(value.has_carrier), Carrier: uint8(value.carrier), NavSystem: fixedCString((*C.char)(unsafe.Pointer(&value.nav_system[0]))), Network: fixedCString((*C.char)(unsafe.Pointer(&value.network[0]))), Country: fixedCString((*C.char)(unsafe.Pointer(&value.country[0]))), HasLatitudeDeg: bool(value.has_lat_deg), LatitudeDeg: float64(value.lat_deg), HasLongitudeDeg: bool(value.has_lon_deg), LongitudeDeg: float64(value.lon_deg), HasNMEARequired: bool(value.has_nmea_required), NMEARequired: bool(value.nmea_required), HasNetworkSolution: bool(value.has_network_solution), NetworkSolution: bool(value.network_solution), Generator: fixedCString((*C.char)(unsafe.Pointer(&value.generator[0]))), Compression: fixedCString((*C.char)(unsafe.Pointer(&value.compression[0]))), Authentication: NTRIPSourcetableAuth(value.authentication), HasFee: bool(value.has_fee), Fee: bool(value.fee), HasBitrate: bool(value.has_bitrate), Bitrate: uint32(value.bitrate), Misc: fixedCString((*C.char)(unsafe.Pointer(&value.misc[0]))),
	}
}

func (t *NTRIPSourcetable) Streams() ([]NTRIPStreamInfo, error) {
	if t == nil || t.handle == nil {
		return nil, ErrClosed
	}
	var result []NTRIPStreamInfo
	err := t.handle.read(func(pointer unsafe.Pointer) error {
		var written, required C.size_t
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_ntrip_sourcetable_streams((*C.SidereonNtripSourcetable)(pointer), nil, 0, &written, &required))
		})
		if err != nil {
			return err
		}
		if _, err := writtenToInt(written, 0, "NTRIP stream first-call written count"); err != nil {
			return err
		}
		requiredInt, err := sizeTToInt(required, "NTRIP stream required count")
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(requiredInt, unsafe.Sizeof(C.SidereonNtripStreamInfo{})); err != nil {
			return err
		}
		values := make([]C.SidereonNtripStreamInfo, requiredInt)
		var output *C.SidereonNtripStreamInfo
		if len(values) != 0 {
			output = &values[0]
		}
		err = callStatus(func() uint32 {
			return uint32(C.sidereon_ntrip_sourcetable_streams((*C.SidereonNtripSourcetable)(pointer), output, C.size_t(len(values)), &written, &required))
		})
		if err == nil {
			writtenInt, countErr := validateTwoPassCounts(
				"NTRIP streams", len(values), requiredInt, uint64(written), uint64(required),
			)
			if countErr != nil {
				return countErr
			}
			result = make([]NTRIPStreamInfo, writtenInt)
			for i := range result {
				result[i] = streamFromC(&values[i])
			}
		}
		return err
	})
	runtime.KeepAlive(t)
	return result, err
}

func (t *NTRIPSourcetable) Text() ([]byte, error) {
	if t == nil || t.handle == nil {
		return nil, ErrClosed
	}
	var result []byte
	err := t.handle.read(func(pointer unsafe.Pointer) error {
		var callErr error
		result, callErr = copyNativeBytes("NTRIP sourcetable", func(out *C.uint8_t, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_ntrip_sourcetable_to_text((*C.SidereonNtripSourcetable)(pointer), out, length, written, required)
		})
		return callErr
	})
	runtime.KeepAlive(t)
	return result, err
}
