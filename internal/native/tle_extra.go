//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"runtime"
	"time"
	"unsafe"
)

type TLEChecksumWarning struct{ LineNumber, Expected, Computed uint8 }
type TLEFile struct {
	_      noCopy
	handle *positioningHandle
}
type TLEBatchPropagation struct {
	_          noCopy
	handle     *positioningHandle
	epochJ2000 []float64
}
type TLEBatchLookAngles struct {
	_      noCopy
	handle *positioningHandle
}
type VisibleSatellite struct {
	CatalogNumber                     string
	AzimuthDeg, ElevationDeg, RangeKm float64
	PositionKm                        [3]float64
}
type VisibleList struct {
	_      noCopy
	handle *positioningHandle
}
type SGP4DecayLatch struct {
	_      noCopy
	handle *positioningHandle
}

func validOpsMode(value uint32) error {
	if value != TLEOpsModeAFSPCValue && value != TLEOpsModeImprovedValue {
		return invalidArgument("TLE operations mode is not defined by the C ABI")
	}
	return nil
}
func releaseTLEFile(pointer unsafe.Pointer) { C.sidereon_tle_file_free((*C.SidereonTleFile)(pointer)) }
func releaseTLEBatchPropagation(pointer unsafe.Pointer) {
	C.sidereon_tle_batch_propagation_free((*C.SidereonTleBatchPropagation)(pointer))
}
func releaseTLEBatchLookAngles(pointer unsafe.Pointer) {
	C.sidereon_tle_batch_look_angles_free((*C.SidereonTleBatchLookAngles)(pointer))
}
func releaseVisibleList(pointer unsafe.Pointer) {
	C.sidereon_visible_list_free((*C.SidereonVisibleList)(pointer))
}
func releaseDecayLatch(pointer unsafe.Pointer) {
	C.sidereon_sgp4_decay_latch_free((*C.SidereonSgp4DecayLatch)(pointer))
}

func ParseTLEFile(text []byte, opsmode uint32) (*TLEFile, error) {
	if err := validOpsMode(opsmode); err != nil {
		return nil, err
	}
	if _, err := checkedNativeSize(len(text)); err != nil {
		return nil, err
	}
	data := append([]byte(nil), text...)
	var ptr unsafe.Pointer
	if len(data) > 0 {
		ptr = C.CBytes(data)
		if ptr == nil {
			return nil, errors.New("sidereon: unable to allocate native TLE file input")
		}
		defer C.free(ptr)
	}
	var out *C.SidereonTleFile
	err := callStatus(func() uint32 {
		return C.sidereon_parse_tle_file((*C.uint8_t)(ptr), C.size_t(len(data)), C.uint32_t(opsmode), &out)
	})
	if err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_tle_file_free(out) })
		}
		return nil, err
	}
	if out == nil {
		return nil, missingNativeHandle("TLE file parse")
	}
	return &TLEFile{handle: newPositioningHandle(unsafe.Pointer(out), releaseTLEFile)}, nil
}
func (f *TLEFile) Close() error {
	if f == nil || f.handle == nil {
		return nil
	}
	return f.handle.close()
}
func (f *TLEFile) count(which bool) (int, error) {
	if f == nil || f.handle == nil {
		return 0, ErrClosed
	}
	var out C.size_t
	err := f.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			if which {
				return C.sidereon_tle_file_skipped((*C.SidereonTleFile)(pointer), &out)
			}
			return C.sidereon_tle_file_count((*C.SidereonTleFile)(pointer), &out)
		})
	})
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(out))
}
func (f *TLEFile) Count() (int, error)   { return f.count(false) }
func (f *TLEFile) Skipped() (int, error) { return f.count(true) }
func (f *TLEFile) Name(index int) (string, error) {
	if f == nil || f.handle == nil {
		return "", ErrClosed
	}
	if index < 0 {
		return "", errors.New("sidereon: TLE file index must not be negative")
	}
	var result string
	err := f.handle.with(func(pointer unsafe.Pointer) error {
		var required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_tle_file_name((*C.SidereonTleFile)(pointer), C.size_t(index), nil, 0, &required)
		}); err != nil {
			return err
		}
		n, err := checkedNativeCount(uint64(required))
		if err != nil {
			return err
		}
		if n <= 0 {
			return errors.New("sidereon: native TLE file name has invalid required size")
		}
		if _, err := checkedNativeAllocationSize(n, 1); err != nil {
			return err
		}
		buffer := C.malloc(C.size_t(n))
		if buffer == nil {
			return errors.New("sidereon: unable to allocate native TLE file name")
		}
		defer C.free(buffer)
		var requiredAgain C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_tle_file_name((*C.SidereonTleFile)(pointer), C.size_t(index), (*C.char)(buffer), C.size_t(n), &requiredAgain)
		}); err != nil {
			return err
		}
		if requiredAgain != required {
			return errors.New("sidereon: native TLE file name size changed")
		}
		result = C.GoString((*C.char)(buffer))
		return nil
	})
	return result, err
}
func (f *TLEFile) Satellite(index int) (*TLE, error) {
	if f == nil || f.handle == nil {
		return nil, ErrClosed
	}
	if index < 0 {
		return nil, errors.New("sidereon: TLE file index must not be negative")
	}
	var out *C.SidereonTle
	err := f.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_tle_file_satellite((*C.SidereonTleFile)(pointer), C.size_t(index), &out)
		})
	})
	if err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_tle_free(out) })
		}
		return nil, err
	}
	if out == nil {
		return nil, missingNativeHandle("TLE file satellite")
	}
	return &TLE{handle: newPositioningHandle(unsafe.Pointer(out), releaseTLE)}, nil
}

func checkedBatchShape(satellites, epochs int) (int, error) {
	if satellites < 0 || epochs < 0 {
		return 0, errors.New("sidereon: negative batch shape")
	}
	maxInt := int(^uint(0) >> 1)
	if epochs != 0 && satellites > maxInt/epochs {
		return 0, errors.New("sidereon: TLE batch shape overflows")
	}
	return satellites * epochs, nil
}
func cTLEPairs(pairs []TLEPair) (unsafe.Pointer, []unsafe.Pointer, error) {
	if len(pairs) == 0 {
		return nil, nil, nil
	}
	maxInt := int(^uint(0) >> 1)
	if len(pairs) > maxInt/2 {
		return nil, nil, errors.New("sidereon: TLE pair allocation overflows")
	}
	if _, err := checkedNativeAllocationSize(len(pairs), unsafe.Sizeof(C.SidereonTlePair{})); err != nil {
		return nil, nil, err
	}
	if _, err := checkedNativeAllocationSize(len(pairs)*2, unsafe.Sizeof(unsafe.Pointer(nil))); err != nil {
		return nil, nil, err
	}
	memory := C.malloc(C.size_t(len(pairs)) * C.size_t(unsafe.Sizeof(C.SidereonTlePair{})))
	if memory == nil {
		return nil, nil, errors.New("sidereon: unable to allocate native TLE pairs")
	}
	values := unsafe.Slice((*C.SidereonTlePair)(memory), len(pairs))
	allocations := make([]unsafe.Pointer, 0, len(pairs)*2)
	for i, pair := range pairs {
		for lineIndex, line := range []string{pair.Line1, pair.Line2} {
			p, err := copyNativeCString(line, "TLE batch line")
			if err != nil {
				C.free(memory)
				for _, q := range allocations {
					C.free(q)
				}
				return nil, nil, err
			}
			allocations = append(allocations, p)
			if lineIndex == 0 {
				values[i].line1 = (*C.char)(p)
			} else {
				values[i].line2 = (*C.char)(p)
			}
		}
	}
	return memory, allocations, nil
}

type TLEPair struct{ Line1, Line2 string }

func PropagateTLEBatch(pairs []TLEPair, epochs []time.Time, opsmode uint32, parallel bool) (*TLEBatchPropagation, error) {
	if err := validOpsMode(opsmode); err != nil {
		return nil, err
	}
	if _, err := checkedBatchShape(len(pairs), len(epochs)); err != nil {
		return nil, err
	}
	pairCopy, epochCopy := append([]TLEPair(nil), pairs...), append([]time.Time(nil), epochs...)
	memory, allocations, err := cTLEPairs(pairCopy)
	if err != nil {
		return nil, err
	}
	if memory != nil {
		defer C.free(memory)
	}
	for _, p := range allocations {
		defer C.free(p)
	}
	epochValues, err := unixMicrosecondsSlice(epochCopy)
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(len(epochValues), unsafe.Sizeof(C.int64_t(0))); err != nil {
		return nil, err
	}
	var epochPtr *C.int64_t
	cEpochs := make([]C.int64_t, len(epochValues))
	for i, v := range epochValues {
		cEpochs[i] = C.int64_t(v)
	}
	if len(cEpochs) > 0 {
		epochPtr = &cEpochs[0]
	}
	var out *C.SidereonTleBatchPropagation
	pairPtr := (*C.SidereonTlePair)(memory)
	err = callStatus(func() uint32 {
		return C.sidereon_propagate_tle_batch(pairPtr, C.size_t(len(pairCopy)), epochPtr, C.size_t(len(cEpochs)), C.uint32_t(opsmode), C.bool(parallel), &out)
	})
	runtime.KeepAlive(cEpochs)
	if err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_tle_batch_propagation_free(out) })
		}
		return nil, err
	}
	if out == nil {
		return nil, missingNativeHandle("TLE batch propagation")
	}
	epochJ2000 := make([]float64, len(epochCopy))
	var conversionErr error
	withCThread(func() {
		for i, value := range epochCopy {
			epochJ2000[i], conversionErr = civilJ2000Locked(value)
			if conversionErr != nil {
				return
			}
		}
	})
	if conversionErr != nil {
		withCThread(func() { C.sidereon_tle_batch_propagation_free(out) })
		return nil, conversionErr
	}
	return &TLEBatchPropagation{handle: newPositioningHandle(unsafe.Pointer(out), releaseTLEBatchPropagation), epochJ2000: epochJ2000}, nil
}
func (b *TLEBatchPropagation) Close() error {
	if b == nil || b.handle == nil {
		return nil
	}
	return b.handle.close()
}
func (b *TLEBatchPropagation) Shape() (int, int, error) {
	if b == nil || b.handle == nil {
		return 0, 0, ErrClosed
	}
	var sat, epoch C.size_t
	err := b.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_tle_batch_propagation_shape((*C.SidereonTleBatchPropagation)(pointer), &sat, &epoch)
		})
	})
	if err != nil {
		return 0, 0, err
	}
	s, err := checkedNativeCount(uint64(sat))
	if err != nil {
		return 0, 0, err
	}
	e, err := checkedNativeCount(uint64(epoch))
	return s, e, err
}
func (b *TLEBatchPropagation) States() ([]TEMEState, error) {
	if b == nil || b.handle == nil {
		return nil, ErrClosed
	}
	var result []TEMEState
	err := b.handle.with(func(pointer unsafe.Pointer) error {
		var satCount, epochCount C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_tle_batch_propagation_shape((*C.SidereonTleBatchPropagation)(pointer), &satCount, &epochCount)
		}); err != nil {
			return err
		}
		sat, err := checkedNativeCount(uint64(satCount))
		if err != nil {
			return err
		}
		epoch, err := checkedNativeCount(uint64(epochCount))
		if err != nil {
			return err
		}
		count, err := checkedBatchShape(sat, epoch)
		if err != nil {
			return err
		}
		if epoch != len(b.epochJ2000) {
			return errors.New("sidereon: native TLE batch epoch shape changed")
		}
		if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonTemeState{})); err != nil {
			return err
		}
		values := make([]C.SidereonTemeState, count)
		var out *C.SidereonTemeState
		if count > 0 {
			out = &values[0]
		}
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_tle_batch_propagation_states((*C.SidereonTleBatchPropagation)(pointer), out, C.size_t(count), &written, &required)
		}); err != nil {
			return err
		}
		n, err := validateTwoPassCounts("TLE batch propagation states", count, count, uint64(written), uint64(required))
		if err != nil {
			return err
		}
		result = make([]TEMEState, n)
		for i := range result {
			result[i].EpochJ2000S = b.epochJ2000[i%epoch]
			for k := 0; k < 3; k++ {
				result[i].PositionKm[k], result[i].VelocityKmPerS[k] = float64(values[i].position_km[k]), float64(values[i].velocity_km_s[k])
			}
		}
		return nil
	})
	return result, err
}

func LookAnglesBatch(pairs []TLEPair, station GroundStation, epochs []time.Time, opsmode uint32, parallel bool) (*TLEBatchLookAngles, error) {
	if err := validOpsMode(opsmode); err != nil {
		return nil, err
	}
	if _, err := checkedBatchShape(len(pairs), len(epochs)); err != nil {
		return nil, err
	}
	pairCopy, epochCopy := append([]TLEPair(nil), pairs...), append([]time.Time(nil), epochs...)
	memory, allocations, err := cTLEPairs(pairCopy)
	if err != nil {
		return nil, err
	}
	if memory != nil {
		defer C.free(memory)
	}
	for _, p := range allocations {
		defer C.free(p)
	}
	epochValues, err := unixMicrosecondsSlice(epochCopy)
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(len(epochValues), unsafe.Sizeof(C.int64_t(0))); err != nil {
		return nil, err
	}
	cEpochs := make([]C.int64_t, len(epochValues))
	for i, value := range epochValues {
		cEpochs[i] = C.int64_t(value)
	}
	var epochPtr *C.int64_t
	if len(cEpochs) > 0 {
		epochPtr = &cEpochs[0]
	}
	stationC := cGroundStation(station)
	var out *C.SidereonTleBatchLookAngles
	err = callStatus(func() uint32 {
		return C.sidereon_tle_batch_look_angles((*C.SidereonTlePair)(memory), C.size_t(len(pairCopy)), &stationC, epochPtr, C.size_t(len(cEpochs)), C.uint32_t(opsmode), C.bool(parallel), &out)
	})
	runtime.KeepAlive(cEpochs)
	if err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_tle_batch_look_angles_free(out) })
		}
		return nil, err
	}
	if out == nil {
		return nil, missingNativeHandle("TLE batch look angles")
	}
	return &TLEBatchLookAngles{handle: newPositioningHandle(unsafe.Pointer(out), releaseTLEBatchLookAngles)}, nil
}
func (b *TLEBatchLookAngles) Close() error {
	if b == nil || b.handle == nil {
		return nil
	}
	return b.handle.close()
}
func (b *TLEBatchLookAngles) Shape() (int, int, error) {
	if b == nil || b.handle == nil {
		return 0, 0, ErrClosed
	}
	var sat, epoch C.size_t
	err := b.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_tle_batch_look_angles_shape((*C.SidereonTleBatchLookAngles)(pointer), &sat, &epoch)
		})
	})
	if err != nil {
		return 0, 0, err
	}
	s, err := checkedNativeCount(uint64(sat))
	if err != nil {
		return 0, 0, err
	}
	e, err := checkedNativeCount(uint64(epoch))
	return s, e, err
}
func (b *TLEBatchLookAngles) Values() ([]LookAngle, error) {
	if b == nil || b.handle == nil {
		return nil, ErrClosed
	}
	var result []LookAngle
	err := b.handle.with(func(pointer unsafe.Pointer) error {
		var satCount, epochCount C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_tle_batch_look_angles_shape((*C.SidereonTleBatchLookAngles)(pointer), &satCount, &epochCount)
		}); err != nil {
			return err
		}
		sat, err := checkedNativeCount(uint64(satCount))
		if err != nil {
			return err
		}
		epoch, err := checkedNativeCount(uint64(epochCount))
		if err != nil {
			return err
		}
		count, err := checkedBatchShape(sat, epoch)
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonLookAngle{})); err != nil {
			return err
		}
		values := make([]C.SidereonLookAngle, count)
		var out *C.SidereonLookAngle
		if count > 0 {
			out = &values[0]
		}
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_tle_batch_look_angles_values((*C.SidereonTleBatchLookAngles)(pointer), out, C.size_t(count), &written, &required)
		}); err != nil {
			return err
		}
		n, err := validateTwoPassCounts("TLE batch look angles", count, count, uint64(written), uint64(required))
		if err != nil {
			return err
		}
		result = make([]LookAngle, n)
		for i := range result {
			result[i] = LookAngle{AzimuthDeg: float64(values[i].azimuth_deg), ElevationDeg: float64(values[i].elevation_deg), RangeKm: float64(values[i].range_km)}
		}
		return nil
	})
	return result, err
}

func (t *TLE) ChecksumWarnings() ([]TLEChecksumWarning, error) {
	if t == nil || t.handle == nil {
		return nil, ErrClosed
	}
	var result []TLEChecksumWarning
	err := t.handle.with(func(pointer unsafe.Pointer) error {
		var written, required C.size_t
		invoke := func(out *C.SidereonTleChecksumWarning, n C.size_t, w, r *C.size_t) uint32 {
			return C.sidereon_tle_checksum_warnings((*C.SidereonTle)(pointer), out, n, w, r)
		}
		if err := callStatus(func() uint32 { return invoke(nil, 0, &written, &required) }); err != nil {
			return err
		}
		count, err := validateNativeQuery("TLE checksum warnings", uint64(written), uint64(required))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonTleChecksumWarning{})); err != nil {
			return err
		}
		values := make([]C.SidereonTleChecksumWarning, count)
		var out *C.SidereonTleChecksumWarning
		if count > 0 {
			out = &values[0]
		}
		written, required = 0, 0
		if err := callStatus(func() uint32 { return invoke(out, C.size_t(count), &written, &required) }); err != nil {
			return err
		}
		n, err := validateTwoPassCounts("TLE checksum warnings", count, count, uint64(written), uint64(required))
		if err != nil {
			return err
		}
		result = make([]TLEChecksumWarning, n)
		for i := range result {
			result[i] = TLEChecksumWarning{LineNumber: uint8(values[i].line_number), Expected: uint8(values[i].expected), Computed: uint8(values[i].computed)}
		}
		return nil
	})
	return result, err
}

func NewSGP4DecayLatch() (*SGP4DecayLatch, error) {
	var out *C.SidereonSgp4DecayLatch
	err := callStatus(func() uint32 { return C.sidereon_sgp4_decay_latch_new(&out) })
	if err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_sgp4_decay_latch_free(out) })
		}
		return nil, err
	}
	if out == nil {
		return nil, missingNativeHandle("SGP4 decay latch")
	}
	return &SGP4DecayLatch{handle: newPositioningHandle(unsafe.Pointer(out), releaseDecayLatch)}, nil
}
func (l *SGP4DecayLatch) Close() error {
	if l == nil || l.handle == nil {
		return nil
	}
	return l.handle.close()
}
func (l *SGP4DecayLatch) Clear() error {
	if l == nil || l.handle == nil {
		return ErrClosed
	}
	return l.handle.withExclusive(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 { return C.sidereon_sgp4_decay_latch_clear((*C.SidereonSgp4DecayLatch)(pointer)) })
	})
}
func (l *SGP4DecayLatch) FirstFailingEpoch() (float64, bool, error) {
	if l == nil || l.handle == nil {
		return 0, false, ErrClosed
	}
	var present C.bool
	var minutes C.double
	err := l.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_sgp4_decay_latch_first_failing_epoch((*C.SidereonSgp4DecayLatch)(pointer), &present, &minutes)
		})
	})
	return float64(minutes), bool(present), err
}
func (t *TLE) PropagateWithDecayLatch(minutes float64, latch *SGP4DecayLatch) (TEMEState, error) {
	if t == nil || t.handle == nil {
		return TEMEState{}, ErrClosed
	}
	if latch == nil || latch.handle == nil {
		return TEMEState{}, ErrClosed
	}
	var out C.SidereonTemeState
	err := t.handle.with(func(tlePointer unsafe.Pointer) error {
		return latch.handle.withExclusive(func(latchPointer unsafe.Pointer) error {
			return callStatus(func() uint32 {
				return C.sidereon_tle_propagate_with_decay_latch((*C.SidereonTle)(tlePointer), C.double(minutes), (*C.SidereonSgp4DecayLatch)(latchPointer), &out)
			})
		})
	})
	return TEMEState{PositionKm: vectorsFromC(out.position_km), VelocityKmPerS: vectorsFromC(out.velocity_km_s)}, err
}

func VisibleFromSatellites(tles []*TLE, ids []string, station GroundStation, epoch time.Time, minElevationDeg float64) (*VisibleList, error) {
	if len(tles) != len(ids) {
		return nil, invalidArgument("TLE and visible-id counts differ")
	}
	for _, id := range ids {
		if err := rejectEmbeddedNUL(id, "visible satellite ID"); err != nil {
			return nil, err
		}
		if len(id) == 0 || len(id) > 64 {
			return nil, invalidArgument("visible satellite ID must contain 1 through 64 bytes")
		}
	}
	unix, err := unixMicroseconds(epoch)
	if err != nil {
		return nil, err
	}
	if len(tles) == 0 {
		return nil, invalidArgument("visible satellite list must not be empty")
	}
	tleBytes, err := checkedNativeAllocationSize(len(tles), unsafe.Sizeof((*C.SidereonTle)(nil)))
	if err != nil {
		return nil, err
	}
	idBytes, err := checkedNativeAllocationSize(len(ids), unsafe.Sizeof((*C.char)(nil)))
	if err != nil {
		return nil, err
	}
	tleMemory := C.malloc(C.size_t(tleBytes))
	if tleMemory == nil {
		return nil, errors.New("sidereon: unable to allocate visible TLE pointers")
	}
	defer C.free(tleMemory)
	idMemory := C.malloc(C.size_t(idBytes))
	if idMemory == nil {
		return nil, errors.New("sidereon: unable to allocate visible ID pointers")
	}
	defer C.free(idMemory)
	tlePointers := unsafe.Slice((**C.SidereonTle)(tleMemory), len(tles))
	idPointers := unsafe.Slice((**C.char)(idMemory), len(ids))
	allocations := make([]unsafe.Pointer, len(ids))
	defer func() {
		for _, p := range allocations {
			if p != nil {
				C.free(p)
			}
		}
	}()
	for i, id := range ids {
		p, err := copyNativeCString(id, "visible satellite ID")
		if err != nil {
			return nil, err
		}
		allocations[i] = p
		idPointers[i] = (*C.char)(p)
	}
	var result *VisibleList
	var operationErr error
	// Hold every TLE read lock while C consumes the borrowed pointer array.
	var recurse func(int) error
	recurse = func(i int) error {
		if i == len(tles) {
			stationC := cGroundStation(station)
			var out *C.SidereonVisibleList
			operationErr = callStatus(func() uint32 {
				return C.sidereon_visible_from_satellites((**C.SidereonTle)(tleMemory), (**C.char)(idMemory), C.size_t(len(tles)), &stationC, C.int64_t(unix), C.double(minElevationDeg), &out)
			})
			if out != nil {
				if operationErr != nil {
					withCThread(func() { C.sidereon_visible_list_free(out) })
				} else {
					result = &VisibleList{handle: newPositioningHandle(unsafe.Pointer(out), releaseVisibleList)}
				}
			}
			return operationErr
		}
		if tles[i] == nil || tles[i].handle == nil {
			return ErrClosed
		}
		return tles[i].handle.with(func(pointer unsafe.Pointer) error { tlePointers[i] = (*C.SidereonTle)(pointer); return recurse(i + 1) })
	}
	err = recurse(0)
	if err != nil {
		if result != nil && result.handle != nil {
			if closeErr := result.handle.close(); closeErr != nil {
				return nil, errors.Join(err, closeErr)
			}
		}
		return nil, err
	}
	if result == nil || result.handle == nil {
		return nil, missingNativeHandle("visible satellite list")
	}
	return result, nil
}
func (v *VisibleList) Close() error {
	if v == nil || v.handle == nil {
		return nil
	}
	return v.handle.close()
}
func (v *VisibleList) Values() ([]VisibleSatellite, error) {
	if v == nil || v.handle == nil {
		return nil, ErrClosed
	}
	var result []VisibleSatellite
	err := v.handle.with(func(pointer unsafe.Pointer) error {
		var written, required C.size_t
		invoke := func(out *C.SidereonVisibleSatellite, n C.size_t, w, r *C.size_t) uint32 {
			return C.sidereon_visible_list_values((*C.SidereonVisibleList)(pointer), out, n, w, r)
		}
		if err := callStatus(func() uint32 { return invoke(nil, 0, &written, &required) }); err != nil {
			return err
		}
		count, err := validateNativeQuery("visible satellite values", uint64(written), uint64(required))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonVisibleSatellite{})); err != nil {
			return err
		}
		values := make([]C.SidereonVisibleSatellite, count)
		var out *C.SidereonVisibleSatellite
		if count > 0 {
			out = &values[0]
		}
		written, required = 0, 0
		if err := callStatus(func() uint32 { return invoke(out, C.size_t(count), &written, &required) }); err != nil {
			return err
		}
		n, err := validateTwoPassCounts("visible satellite values", count, count, uint64(written), uint64(required))
		if err != nil {
			return err
		}
		result = make([]VisibleSatellite, n)
		for i := range result {
			result[i].CatalogNumber = C.GoString(&values[i].catalog_number[0])
			result[i].AzimuthDeg, result[i].ElevationDeg, result[i].RangeKm = float64(values[i].azimuth_deg), float64(values[i].elevation_deg), float64(values[i].range_km)
			for k := 0; k < 3; k++ {
				result[i].PositionKm[k] = float64(values[i].position_km[k])
			}
		}
		return nil
	})
	return result, err
}
func (v *VisibleList) Count() (int, error) {
	if v == nil || v.handle == nil {
		return 0, ErrClosed
	}
	var out C.size_t
	err := v.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 { return C.sidereon_visible_list_count((*C.SidereonVisibleList)(pointer), &out) })
	})
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(out))
}
