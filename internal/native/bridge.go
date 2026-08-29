//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && ((sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#cgo darwin,arm64,!sidereon_use_system_lib LDFLAGS: -L${SRCDIR}/lib -lsidereon_darwin_arm64
#cgo darwin,amd64,!sidereon_use_system_lib LDFLAGS: -L${SRCDIR}/lib -lsidereon_darwin_amd64
#cgo linux,amd64,sidereon_linux_glibc,!sidereon_linux_musl,!sidereon_use_system_lib LDFLAGS: -L${SRCDIR}/lib -lsidereon_linux_amd64_glibc
#cgo linux,arm64,sidereon_linux_glibc,!sidereon_linux_musl,!sidereon_use_system_lib LDFLAGS: -L${SRCDIR}/lib -lsidereon_linux_arm64_glibc
#cgo linux,amd64,sidereon_linux_musl,!sidereon_linux_glibc,!sidereon_use_system_lib LDFLAGS: -L${SRCDIR}/lib -lsidereon_linux_amd64_musl
#cgo linux,arm64,sidereon_linux_musl,!sidereon_linux_glibc,!sidereon_use_system_lib LDFLAGS: -L${SRCDIR}/lib -lsidereon_linux_arm64_musl
#cgo windows,amd64,!sidereon_use_system_lib LDFLAGS: -L${SRCDIR}/lib -lsidereon_windows_amd64
#include <sidereon.h>
*/
import "C"

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"unsafe"
)

// StatusError is the internal form translated to the public package error.
type StatusError struct {
	Code   int
	Text   string
	Detail string
}

func (e *StatusError) Error() string {
	if e.Detail == "" {
		return fmt.Sprintf("%s (%d)", e.Text, e.Code)
	}
	return fmt.Sprintf("%s (%d): %s", e.Text, e.Code, e.Detail)
}

var ErrClosed = errors.New("sidereon: handle is closed")

var cgoCallMu sync.Mutex

func withCThread(fn func()) {
	cgoCallMu.Lock()
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer cgoCallMu.Unlock()
	fn()
}

func callStatus(fn func() uint32) error {
	var err error
	withCThread(func() {
		err = statusErrorLocked(fn())
	})
	return err
}

func statusErrorLocked(status uint32) error {
	if status == C.SIDEREON_STATUS_OK {
		return nil
	}

	text := C.GoString(C.sidereon_status_message(status))
	required := C.sidereon_last_error_message(nil, 0)
	var detail string
	if required > 0 {
		buffer := make([]byte, int(required)+1)
		written := C.sidereon_last_error_message(
			(*C.char)(unsafe.Pointer(&buffer[0])), C.size_t(len(buffer)))
		if written > C.size_t(len(buffer)) {
			written = C.size_t(len(buffer))
		}
		detail = string(buffer[:int(written)])
	}
	return &StatusError{Code: int(status), Text: text, Detail: detail}
}

type Version struct {
	Major  uint32
	Minor  uint32
	Patch  uint32
	String string
}

func LibraryVersion() Version {
	var major, minor, patch C.uint32_t
	var versionString string
	withCThread(func() {
		C.sidereon_version(&major, &minor, &patch)
		versionString = C.GoString(C.sidereon_version_string())
	})
	return Version{
		Major:  uint32(major),
		Minor:  uint32(minor),
		Patch:  uint32(patch),
		String: versionString,
	}
}

func SecondOfDay(hour, minute int, second float64) (float64, error) {
	var out C.double
	err := callStatus(func() uint32 {
		return C.sidereon_second_of_day(C.int32_t(hour), C.int32_t(minute), C.double(second), &out)
	})
	return float64(out), err
}

func DayOfYear(year, month, day, hour, minute int, second float64) (float64, error) {
	var out C.double
	err := callStatus(func() uint32 {
		return C.sidereon_day_of_year(
			C.int32_t(year), C.int32_t(month), C.int32_t(day),
			C.int32_t(hour), C.int32_t(minute), C.double(second), &out,
		)
	})
	return float64(out), err
}

func DataDayOfYear(year int, month, day uint8) (uint16, error) {
	var out C.uint16_t
	err := callStatus(func() uint32 {
		return C.sidereon_data_day_of_year(C.int32_t(year), C.uint8_t(month), C.uint8_t(day), &out)
	})
	return uint16(out), err
}

type CovarianceValidation struct {
	Symmetric            bool
	PositiveSemidefinite bool
}

func covarianceToC(values [6][6]float64) C.SidereonCovarianceMatrix6 {
	var out C.SidereonCovarianceMatrix6
	for row := 0; row < 6; row++ {
		for column := 0; column < 6; column++ {
			out.values[row][column] = C.double(values[row][column])
		}
	}
	return out
}

func covarianceFromC(values *C.SidereonCovarianceMatrix6) [6][6]float64 {
	var out [6][6]float64
	for row := 0; row < 6; row++ {
		for column := 0; column < 6; column++ {
			out[row][column] = float64(values.values[row][column])
		}
	}
	return out
}

func CovarianceFromDiagonal(diagonal []float64) ([6][6]float64, error) {
	var out C.SidereonCovarianceMatrix6
	cdiagonal := make([]C.double, len(diagonal))
	for i, value := range diagonal {
		cdiagonal[i] = C.double(value)
	}
	var pointer *C.double
	if len(cdiagonal) != 0 {
		pointer = &cdiagonal[0]
	}
	err := callStatus(func() uint32 {
		return C.sidereon_covariance6_from_diagonal(pointer, C.size_t(len(cdiagonal)), &out)
	})
	if err != nil {
		return [6][6]float64{}, err
	}
	return covarianceFromC(&out), nil
}

func CovarianceValidate(values [6][6]float64) (CovarianceValidation, error) {
	input := covarianceToC(values)
	var out C.SidereonCovariance6Validation
	err := callStatus(func() uint32 {
		return C.sidereon_covariance6_validate(&input, &out)
	})
	return CovarianceValidation{
		Symmetric:            bool(out.symmetric),
		PositiveSemidefinite: bool(out.positive_semidefinite),
	}, err
}

func CovarianceKmToM(values [6][6]float64) ([6][6]float64, error) {
	input := covarianceToC(values)
	var out C.SidereonCovarianceMatrix6
	err := callStatus(func() uint32 {
		return C.sidereon_covariance6_km_to_m(&input, &out)
	})
	if err != nil {
		return [6][6]float64{}, err
	}
	return covarianceFromC(&out), nil
}

func CovarianceMToKm(values [6][6]float64) ([6][6]float64, error) {
	input := covarianceToC(values)
	var out C.SidereonCovarianceMatrix6
	err := callStatus(func() uint32 {
		return C.sidereon_covariance6_m_to_km(&input, &out)
	})
	if err != nil {
		return [6][6]float64{}, err
	}
	return covarianceFromC(&out), nil
}

func CovarianceInterpolate(a, b [6][6]float64, u float64) ([6][6]float64, error) {
	left := covarianceToC(a)
	right := covarianceToC(b)
	var out C.SidereonCovarianceMatrix6
	err := callStatus(func() uint32 {
		return C.sidereon_covariance6_interpolate_psd(&left, &right, C.double(u), &out)
	})
	if err != nil {
		return [6][6]float64{}, err
	}
	return covarianceFromC(&out), nil
}

type NMEASummary struct {
	SentenceCount uint64
	EpochCount    uint64
	SkipCount     uint64
	WarningCount  uint64
}

type NMEAEpoch struct {
	HasPosition        bool
	LatitudeRad        float64
	LongitudeRad       float64
	HeightM            float64
	SentenceCount      uint64
	UsedSatelliteCount uint64
	SatellitesInView   uint64
	SkipCount          uint64
	WarningCount       uint64
}

type resource struct {
	mu      sync.RWMutex
	ptr     unsafe.Pointer
	release func(unsafe.Pointer)
}

func (r *resource) with(fn func(unsafe.Pointer) error) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.ptr == nil {
		return ErrClosed
	}
	return fn(r.ptr)
}

func (r *resource) close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.ptr == nil {
		return
	}
	pointer := r.ptr
	r.ptr = nil
	withCThread(func() { r.release(pointer) })
}

type NMEALog struct {
	mu       sync.RWMutex
	resource *resource
	cleanup  runtime.Cleanup
}

func releaseNMEALog(pointer unsafe.Pointer) {
	C.sidereon_nmea_log_free((*C.SidereonNmeaLog)(pointer))
}

func cleanupNMEALog(r *resource) {
	r.close()
}

func newNMEALog(pointer *C.SidereonNmeaLog) *NMEALog {
	r := &resource{ptr: unsafe.Pointer(pointer), release: releaseNMEALog}
	log := &NMEALog{resource: r}
	log.cleanup = runtime.AddCleanup(log, cleanupNMEALog, r)
	return log
}

func ParseNMEA(data []byte) (*NMEALog, error) {
	var pointer *C.SidereonNmeaLog
	var err error
	withCThread(func() {
		var cdata unsafe.Pointer
		if len(data) != 0 {
			cdata = C.CBytes(data)
			if cdata == nil {
				err = errors.New("sidereon: unable to allocate native input buffer")
				return
			}
			defer C.free(cdata)
		}
		err = statusErrorLocked(C.sidereon_nmea_parse(
			(*C.uint8_t)(cdata), C.size_t(len(data)), &pointer,
		))
	})
	if err != nil {
		return nil, err
	}
	return newNMEALog(pointer), nil
}

func (l *NMEALog) Close() error {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.resource == nil {
		return nil
	}
	r := l.resource
	l.resource = nil
	l.cleanup.Stop()
	r.close()
	return nil
}

func (l *NMEALog) Summary() (NMEASummary, error) {
	if l == nil {
		return NMEASummary{}, ErrClosed
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.resource == nil {
		return NMEASummary{}, ErrClosed
	}
	var out C.SidereonNmeaSummary
	err := l.resource.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_nmea_log_summary((*C.SidereonNmeaLog)(pointer), &out)
		})
	})
	runtime.KeepAlive(l)
	return NMEASummary{
		SentenceCount: uint64(out.sentence_count),
		EpochCount:    uint64(out.epoch_count),
		SkipCount:     uint64(out.skip_count),
		WarningCount:  uint64(out.warning_count),
	}, err
}

func (l *NMEALog) Epochs() ([]NMEAEpoch, error) {
	if l == nil {
		return nil, ErrClosed
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.resource == nil {
		return nil, ErrClosed
	}
	var epochs []C.SidereonNmeaEpochSummary
	err := l.resource.with(func(pointer unsafe.Pointer) error {
		var written, required C.size_t
		err := callStatus(func() uint32 {
			return C.sidereon_nmea_log_epochs(
				(*C.SidereonNmeaLog)(pointer), nil, 0, &written, &required,
			)
		})
		if err != nil {
			return err
		}
		epochs = make([]C.SidereonNmeaEpochSummary, int(required))
		var output *C.SidereonNmeaEpochSummary
		if len(epochs) != 0 {
			output = &epochs[0]
		}
		err = callStatus(func() uint32 {
			return C.sidereon_nmea_log_epochs(
				(*C.SidereonNmeaLog)(pointer), output, C.size_t(len(epochs)), &written, &required,
			)
		})
		if err != nil {
			return err
		}
		epochs = epochs[:int(written)]
		return nil
	})
	runtime.KeepAlive(l)
	if err != nil {
		return nil, err
	}
	out := make([]NMEAEpoch, len(epochs))
	for i := range epochs {
		epoch := &epochs[i]
		out[i] = NMEAEpoch{
			HasPosition:        bool(epoch.has_position),
			LatitudeRad:        float64(epoch.position.lat_rad),
			LongitudeRad:       float64(epoch.position.lon_rad),
			HeightM:            float64(epoch.position.height_m),
			SentenceCount:      uint64(epoch.sentence_count),
			UsedSatelliteCount: uint64(epoch.used_satellite_count),
			SatellitesInView:   uint64(epoch.satellites_in_view),
			SkipCount:          uint64(epoch.skip_count),
			WarningCount:       uint64(epoch.warning_count),
		}
	}
	return out, nil
}
