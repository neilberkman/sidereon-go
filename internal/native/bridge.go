//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#cgo darwin,arm64,!sidereon_use_system_lib LDFLAGS: -L${SRCDIR}/lib -lsidereon_darwin_arm64
#cgo darwin,amd64,!sidereon_use_system_lib LDFLAGS: -L${SRCDIR}/lib -lsidereon_darwin_amd64
#cgo linux,amd64,sidereon_linux_glibc,!sidereon_linux_musl,!sidereon_use_system_lib LDFLAGS: -L${SRCDIR}/lib -lsidereon_linux_amd64_glibc -lgcc_eh -lutil -lrt -lpthread -lm -ldl
#cgo linux,arm64,sidereon_linux_glibc,!sidereon_linux_musl,!sidereon_use_system_lib LDFLAGS: -L${SRCDIR}/lib -lsidereon_linux_arm64_glibc -lgcc_eh -lutil -lrt -lpthread -lm -ldl
#cgo linux,amd64,sidereon_linux_musl,!sidereon_linux_glibc,!sidereon_use_system_lib LDFLAGS: -L${SRCDIR}/lib -lsidereon_linux_amd64_musl -lgcc_eh -lutil -lrt -lpthread -lm -ldl
#cgo linux,arm64,sidereon_linux_musl,!sidereon_linux_glibc,!sidereon_use_system_lib LDFLAGS: -L${SRCDIR}/lib -lsidereon_linux_arm64_musl -lgcc_eh -lutil -lrt -lpthread -lm -ldl
#cgo windows,amd64,!sidereon_use_system_lib LDFLAGS: -L${SRCDIR}/lib -lsidereon_windows_amd64_gnu -lgcc_eh -lws2_32 -luserenv -lbcrypt -lntdll
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
	Code         int
	Text         string
	Detail       string
	TerrainDatum *TerrainDatumError
	TerrainStore *TerrainStoreError
}

func (e *StatusError) Error() string {
	if e.Detail == "" {
		return fmt.Sprintf("%s (%d)", e.Text, e.Code)
	}
	return fmt.Sprintf("%s (%d): %s", e.Text, e.Code, e.Detail)
}

func (e *StatusError) Unwrap() error {
	var details []error
	if e.TerrainDatum != nil {
		details = append(details, e.TerrainDatum)
	}
	if e.TerrainStore != nil {
		details = append(details, e.TerrainStore)
	}
	return errors.Join(details...)
}

var ErrClosed = errors.New("sidereon: handle is closed")

func withCThread(fn func()) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	fn()
}

func copyNativeInput(data []byte) (unsafe.Pointer, error) {
	if len(data) == 0 {
		return nil, nil
	}
	pointer := C.CBytes(data)
	if pointer == nil {
		return nil, errors.New("sidereon: unable to allocate native input buffer")
	}
	return pointer, nil
}

func freeNativeInput(pointer unsafe.Pointer) {
	if pointer != nil {
		C.free(pointer)
	}
}

// withInputError copies Go input into temporary C-owned storage and runs an
// error-returning native operation on the locked OS thread.
func withInputError(data []byte, fn func(*C.uint8_t, C.size_t) error) error {
	length, err := checkedNativeSize(len(data))
	if err != nil {
		return err
	}
	var operationErr error
	withCThread(func() {
		pointer, copyErr := copyNativeInput(data)
		if copyErr != nil {
			operationErr = copyErr
			return
		}
		defer freeNativeInput(pointer)
		operationErr = fn((*C.uint8_t)(pointer), length)
	})
	runtime.KeepAlive(data)
	return operationErr
}

func callStatus(fn func() uint32) error {
	var err error
	withCThread(func() {
		err = statusErrorLocked(fn())
	})
	return err
}

func callStatusWithTerrainDiagnostics(fn func() uint32, captureDatum, captureStore bool) error {
	var err error
	withCThread(func() {
		err = statusErrorLockedWithTerrainDiagnostics(fn(), captureDatum, captureStore)
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
		count, countErr := checkedNativeCount(uint64(required))
		if countErr != nil {
			return &StatusError{Code: int(status), Text: text, Detail: countErr.Error()}
		}
		if count == int(^uint(0)>>1) {
			return &StatusError{Code: int(status), Text: text, Detail: "sidereon: native error detail is too large"}
		}
		if _, allocationErr := checkedNativeAllocationSize(count+1, 1); allocationErr != nil {
			return &StatusError{Code: int(status), Text: text, Detail: allocationErr.Error()}
		}
		buffer := make([]byte, count+1)
		outputLength, err := cSize(len(buffer), "error detail output capacity")
		if err != nil {
			return &StatusError{Code: int(status), Text: text, Detail: err.Error()}
		}
		written := C.sidereon_last_error_message(
			(*C.char)(unsafe.Pointer(&buffer[0])), outputLength)
		writtenCount, countErr := validateTwoPassCounts("error detail", count, count, uint64(written), uint64(required))
		if countErr != nil {
			return &StatusError{Code: int(status), Text: text, Detail: countErr.Error()}
		}
		detail = string(buffer[:writtenCount])
	}
	return &StatusError{Code: int(status), Text: text, Detail: detail}
}

func statusErrorLockedWithTerrainDiagnostics(status uint32, captureDatum, captureStore bool) error {
	err := statusErrorLocked(status)
	if err == nil || (!captureDatum && !captureStore) {
		return err
	}
	statusErr, ok := err.(*StatusError)
	if !ok {
		return err
	}
	if captureDatum {
		value, diagnosticErr := lastTerrainDatumErrorLocked()
		if diagnosticErr == nil && value.Kind != TerrainDatumErrorNoneValue {
			statusErr.TerrainDatum = &value
		}
	}
	if captureStore {
		value, diagnosticErr := lastTerrainStoreErrorLocked()
		if diagnosticErr == nil && value.Kind != TerrainStoreErrorNoneValue {
			statusErr.TerrainStore = &value
		}
	}
	return statusErr
}

func sizeTToInt(value C.size_t, field string) (int, error) {
	maxInt := uint64(^uint(0) >> 1)
	if uint64(value) > maxInt {
		return 0, fmt.Errorf("sidereon: native %s %d does not fit in int", field, uint64(value))
	}
	return int(value), nil
}

func writtenToInt(value C.size_t, capacity int, field string) (int, error) {
	converted, err := sizeTToInt(value, field)
	if err != nil {
		return 0, err
	}
	if converted > capacity {
		return 0, fmt.Errorf("sidereon: native %s %d exceeds allocated capacity %d", field, converted, capacity)
	}
	return converted, nil
}

type nativeBytesCall func(*C.uint8_t, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus

// copyNativeBytesLocked performs the complete variable-length output contract
// on the caller's locked C thread. The second call must reproduce the queried
// required count and fill exactly that many bytes; accepting a partial copy
// would make a native result ambiguous.
func copyNativeBytesLocked(label string, call nativeBytesCall) ([]byte, error) {
	return copyNativeBytesLockedWithStatus(label, call, statusErrorLocked)
}

func copyNativeBytesLockedWithTerrainDiagnostics(label string, call nativeBytesCall, captureDatum, captureStore bool) ([]byte, error) {
	return copyNativeBytesLockedWithStatus(label, call, func(status uint32) error {
		return statusErrorLockedWithTerrainDiagnostics(status, captureDatum, captureStore)
	})
}

func copyNativeBytesLockedWithStatus(label string, call nativeBytesCall, statusError func(uint32) error) ([]byte, error) {
	var written, required C.size_t
	if err := statusError(uint32(call(nil, 0, &written, &required))); err != nil {
		return nil, err
	}
	requiredInt, err := sizeTToInt(required, label+" required byte count")
	if err != nil {
		return nil, err
	}
	if _, err := writtenToInt(written, 0, label+" first-call written byte count"); err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(requiredInt, unsafe.Sizeof(byte(0))); err != nil {
		return nil, err
	}
	buffer := make([]byte, requiredInt)
	outputLength, err := cSize(len(buffer), label+" output byte capacity")
	if err != nil {
		return nil, err
	}
	var output *C.uint8_t
	if len(buffer) != 0 {
		output = (*C.uint8_t)(unsafe.Pointer(&buffer[0]))
	}
	written, required = 0, 0
	if err := statusError(uint32(call(output, outputLength, &written, &required))); err != nil {
		return nil, err
	}
	writtenInt, err := validateTwoPassCounts(label, len(buffer), requiredInt, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	return append([]byte(nil), buffer[:writtenInt]...), nil
}

func copyNativeBytes(label string, call nativeBytesCall) (result []byte, err error) {
	withCThread(func() {
		result, err = copyNativeBytesLocked(label, call)
	})
	return result, err
}

func copyNativeBytesWithTerrainDiagnostics(label string, call nativeBytesCall, captureDatum, captureStore bool) (result []byte, err error) {
	withCThread(func() {
		result, err = copyNativeBytesLockedWithTerrainDiagnostics(label, call, captureDatum, captureStore)
	})
	return result, err
}

// withStringError copies a Go string into temporary C-owned storage and runs
// an error-returning native operation on the locked OS thread.
func withStringError(value string, fn func(*C.char) error) error {
	if err := rejectEmbeddedNUL(value, "native string"); err != nil {
		return err
	}
	if len(value) == int(^uint(0)>>1) {
		return errors.New("sidereon: native string is too large")
	}
	if _, err := checkedNativeAllocationSize(len(value)+1, 1); err != nil {
		return err
	}
	var err error
	withCThread(func() {
		pointer := C.CString(value)
		if pointer == nil {
			err = errors.New("sidereon: unable to allocate native string")
			return
		}
		defer C.free(unsafe.Pointer(pointer))
		err = fn(pointer)
	})
	runtime.KeepAlive(value)
	return err
}

// withToken validates and copies a fixed-width satellite token for C.
func withToken(value, name string, fn func(*C.char) uint32) error {
	if err := rejectEmbeddedNUL(value, name); err != nil {
		return err
	}
	if len(value) >= 16 {
		return fmt.Errorf("sidereon: %s is too long", name)
	}
	return withString(value, fn)
}

// withTokenError validates and copies a fixed-width satellite token for a
// callback that can report allocation or native errors directly.
func withTokenError(value, name string, fn func(*C.char) error) error {
	if err := rejectEmbeddedNUL(value, name); err != nil {
		return err
	}
	if len(value) >= 16 {
		return fmt.Errorf("sidereon: %s is too long", name)
	}
	return withStringError(value, fn)
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
	checkedHour, err := checkedInt32(hour, "hour")
	if err != nil {
		return 0, err
	}
	checkedMinute, err := checkedInt32(minute, "minute")
	if err != nil {
		return 0, err
	}
	var out C.double
	err = callStatus(func() uint32 {
		return C.sidereon_second_of_day(C.int32_t(checkedHour), C.int32_t(checkedMinute), C.double(second), &out)
	})
	return float64(out), err
}

func DayOfYear(year, month, day, hour, minute int, second float64) (float64, error) {
	checkedYear, err := checkedInt32(year, "year")
	if err != nil {
		return 0, err
	}
	checkedMonth, err := checkedInt32(month, "month")
	if err != nil {
		return 0, err
	}
	checkedDay, err := checkedInt32(day, "day")
	if err != nil {
		return 0, err
	}
	checkedHour, err := checkedInt32(hour, "hour")
	if err != nil {
		return 0, err
	}
	checkedMinute, err := checkedInt32(minute, "minute")
	if err != nil {
		return 0, err
	}
	var out C.double
	err = callStatus(func() uint32 {
		return C.sidereon_day_of_year(
			C.int32_t(checkedYear), C.int32_t(checkedMonth), C.int32_t(checkedDay),
			C.int32_t(checkedHour), C.int32_t(checkedMinute), C.double(second), &out,
		)
	})
	return float64(out), err
}

func DataDayOfYear(year int, month, day uint8) (uint16, error) {
	checkedYear, err := checkedInt32(year, "year")
	if err != nil {
		return 0, err
	}
	var out C.uint16_t
	err = callStatus(func() uint32 {
		return C.sidereon_data_day_of_year(C.int32_t(checkedYear), C.uint8_t(month), C.uint8_t(day), &out)
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
	count, err := cSize(len(diagonal), "covariance diagonal length")
	if err != nil {
		return [6][6]float64{}, err
	}
	cdiagonal := make([]C.double, len(diagonal))
	for i, value := range diagonal {
		cdiagonal[i] = C.double(value)
	}
	var pointer *C.double
	if len(cdiagonal) != 0 {
		pointer = &cdiagonal[0]
	}
	err = callStatus(func() uint32 {
		return C.sidereon_covariance6_from_diagonal(pointer, count, &out)
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

func newNMEALog(pointer *C.SidereonNmeaLog) (*NMEALog, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	r := &resource{ptr: unsafe.Pointer(pointer), release: releaseNMEALog}
	log := &NMEALog{resource: r}
	log.cleanup = runtime.AddCleanup(log, cleanupNMEALog, r)
	return log, nil
}

func ParseNMEA(data []byte) (*NMEALog, error) {
	var pointer *C.SidereonNmeaLog
	var err error
	length, lengthErr := cSize(len(data), "NMEA input")
	if lengthErr != nil {
		return nil, lengthErr
	}
	withCThread(func() {
		cdata, copyErr := copyNativeInput(data)
		if copyErr != nil {
			err = copyErr
			return
		}
		defer freeNativeInput(cdata)
		err = statusErrorLocked(C.sidereon_nmea_parse(
			(*C.uint8_t)(cdata), length, &pointer,
		))
		if err != nil && pointer != nil {
			releaseNMEALog(unsafe.Pointer(pointer))
			pointer = nil
		}
	})
	if err != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_nmea_log_free(pointer) })
		}
		return nil, err
	}
	if pointer == nil {
		return nil, errors.New("sidereon: native NMEA parse returned no handle")
	}
	return newNMEALog(pointer)
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
		count, err := checkedNativeCount(uint64(required))
		if err != nil {
			return err
		}
		if _, err := writtenToInt(written, 0, "NMEA epoch first-call written count"); err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonNmeaEpochSummary{})); err != nil {
			return err
		}
		epochs = make([]C.SidereonNmeaEpochSummary, count)
		outputLength, err := cSize(len(epochs), "NMEA epoch output capacity")
		if err != nil {
			return err
		}
		var output *C.SidereonNmeaEpochSummary
		if len(epochs) != 0 {
			output = &epochs[0]
		}
		written, required = 0, 0
		err = callStatus(func() uint32 {
			return C.sidereon_nmea_log_epochs(
				(*C.SidereonNmeaLog)(pointer), output, outputLength, &written, &required,
			)
		})
		if err != nil {
			return err
		}
		writtenCount, err := validateTwoPassCounts("NMEA epochs", len(epochs), count, uint64(written), uint64(required))
		if err != nil {
			return err
		}
		epochs = epochs[:writtenCount]
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
