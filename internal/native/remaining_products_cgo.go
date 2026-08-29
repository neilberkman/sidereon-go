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
	"unsafe"
)

const (
	PreciseInterpolantArtifactErrorNoneValue                       = uint32(C.SIDEREON_PRECISE_INTERPOLANT_ARTIFACT_ERROR_KIND_NONE)
	PreciseInterpolantArtifactErrorTruncatedValue                  = uint32(C.SIDEREON_PRECISE_INTERPOLANT_ARTIFACT_ERROR_KIND_TRUNCATED)
	PreciseInterpolantArtifactErrorCorruptValue                    = uint32(C.SIDEREON_PRECISE_INTERPOLANT_ARTIFACT_ERROR_KIND_CORRUPT)
	PreciseInterpolantArtifactErrorParseValue                      = uint32(C.SIDEREON_PRECISE_INTERPOLANT_ARTIFACT_ERROR_KIND_PARSE)
	PreciseInterpolantArtifactErrorUnsupportedVersionValue         = uint32(C.SIDEREON_PRECISE_INTERPOLANT_ARTIFACT_ERROR_KIND_UNSUPPORTED_VERSION)
	PreciseInterpolantArtifactErrorUnsupportedTimeScaleValue       = uint32(C.SIDEREON_PRECISE_INTERPOLANT_ARTIFACT_ERROR_KIND_UNSUPPORTED_TIME_SCALE)
	PreciseInterpolantArtifactErrorUnsupportedSatelliteSystemValue = uint32(C.SIDEREON_PRECISE_INTERPOLANT_ARTIFACT_ERROR_KIND_UNSUPPORTED_SATELLITE_SYSTEM)
	PreciseInterpolantArtifactErrorDuplicateSatelliteValue         = uint32(C.SIDEREON_PRECISE_INTERPOLANT_ARTIFACT_ERROR_KIND_DUPLICATE_SATELLITE)
	PreciseInterpolantArtifactErrorIOValue                         = uint32(C.SIDEREON_PRECISE_INTERPOLANT_ARTIFACT_ERROR_KIND_IO)
	PreciseInterpolantArtifactErrorAttestedChecksumMismatchValue   = uint32(C.SIDEREON_PRECISE_INTERPOLANT_ARTIFACT_ERROR_KIND_ATTESTED_CHECKSUM_MISMATCH)
)

func validatePreciseInterpolantArtifactError(value uint32) error {
	if value > PreciseInterpolantArtifactErrorAttestedChecksumMismatchValue {
		return invalidArgument("invalid precise-interpolant artifact error kind returned by native code")
	}
	return nil
}

func validateDigestProvenance(value uint32) error {
	if value != DigestProvenanceVerifiedValue && value != DigestProvenanceAttestedValue {
		return invalidArgument("invalid digest provenance returned by native code")
	}
	return nil
}

func validateObservableStateResultStatusValue(value uint32) error {
	if value < uint32(C.SIDEREON_STATUS_OK) || value > uint32(C.SIDEREON_STATUS_TIMEOUT) {
		return invalidArgument("invalid observable-state result status returned by native code")
	}
	return nil
}

// PreciseEphemerisSamples and PreciseEphemerisInterpolant are owning handles.
// Their methods use positioningHandle so reads and Close are serialized.
type PreciseEphemerisSamples struct {
	_      noCopy
	handle *positioningHandle
}
type PreciseEphemerisInterpolant struct {
	_      noCopy
	handle *positioningHandle
}

func releasePreciseSamples(p unsafe.Pointer) {
	withCThread(func() { C.sidereon_precise_ephemeris_samples_free((*C.SidereonPreciseEphemerisSamples)(p)) })
}

func releasePreciseInterpolant(p unsafe.Pointer) {
	withCThread(func() { C.sidereon_precise_ephemeris_interpolant_free((*C.SidereonPreciseEphemerisInterpolant)(p)) })
}

func cPreciseSamples(values []PreciseEphemerisSample) (unsafe.Pointer, C.size_t, error) {
	if len(values) == 0 {
		return nil, 0, invalidArgument("precise-ephemeris samples must not be empty")
	}
	size, err := checkedNativeAllocationSize(len(values), unsafe.Sizeof(C.SidereonPreciseEphemerisSample{}))
	if err != nil {
		return nil, 0, err
	}
	p := C.malloc(C.size_t(size))
	if p == nil {
		return nil, 0, errors.New("sidereon: unable to allocate native precise samples")
	}
	out := unsafe.Slice((*C.SidereonPreciseEphemerisSample)(p), len(values))
	for i := range values {
		converted, convertErr := cPreciseSample(values[i])
		if convertErr != nil {
			C.free(p)
			return nil, 0, convertErr
		}
		out[i] = converted
	}
	return p, C.size_t(len(values)), nil
}

func newPreciseSamples(p *C.SidereonPreciseEphemerisSamples) (*PreciseEphemerisSamples, error) {
	if p == nil {
		return nil, missingNativeHandle("precise-ephemeris samples")
	}
	return &PreciseEphemerisSamples{handle: newPositioningHandle(unsafe.Pointer(p), releasePreciseSamples)}, nil
}

func newPreciseInterpolant(p *C.SidereonPreciseEphemerisInterpolant) (*PreciseEphemerisInterpolant, error) {
	if p == nil {
		return nil, missingNativeHandle("precise-ephemeris interpolant")
	}
	return &PreciseEphemerisInterpolant{handle: newPositioningHandle(unsafe.Pointer(p), releasePreciseInterpolant)}, nil
}

// PreciseEphemerisSamplesFromSamples copies every nested value into C-owned
// temporary storage before constructing the native source.
func PreciseEphemerisSamplesFromSamples(values []PreciseEphemerisSample) (*PreciseEphemerisSamples, error) {
	p, count, err := cPreciseSamples(values)
	if err != nil {
		return nil, err
	}
	defer C.free(p)
	var out *C.SidereonPreciseEphemerisSamples
	withCThread(func() {
		err = statusErrorLocked(uint32(C.sidereon_precise_ephemeris_samples_from_samples((*C.SidereonPreciseEphemerisSample)(p), count, &out)))
		if err != nil && out != nil {
			C.sidereon_precise_ephemeris_samples_free(out)
			out = nil
		}
	})
	if err != nil {
		return nil, err
	}
	return newPreciseSamples(out)
}

func (s *PreciseEphemerisSamples) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return s.handle.close()
}

func (s *PreciseEphemerisSamples) interpolant() (*PreciseEphemerisInterpolant, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var out *C.SidereonPreciseEphemerisInterpolant
	err := s.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_precise_ephemeris_interpolant_from_precise_ephemeris_samples((*C.SidereonPreciseEphemerisSamples)(p), &out))
		})
	})
	runtime.KeepAlive(s)
	if err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_precise_ephemeris_interpolant_free(out) })
		}
		return nil, err
	}
	return newPreciseInterpolant(out)
}

func (s *PreciseEphemerisSamples) Interpolant() (*PreciseEphemerisInterpolant, error) {
	return s.interpolant()
}

func PreciseInterpolantFromSamples(values []PreciseEphemerisSample) (*PreciseEphemerisInterpolant, error) {
	p, count, err := cPreciseSamples(values)
	if err != nil {
		return nil, err
	}
	defer C.free(p)
	var out *C.SidereonPreciseEphemerisInterpolant
	withCThread(func() {
		err = statusErrorLocked(uint32(C.sidereon_precise_ephemeris_interpolant_from_samples((*C.SidereonPreciseEphemerisSample)(p), count, &out)))
		if err != nil && out != nil {
			C.sidereon_precise_ephemeris_interpolant_free(out)
			out = nil
		}
	})
	if err != nil {
		return nil, err
	}
	return newPreciseInterpolant(out)
}

func PreciseInterpolantFromSP3(sp3 *SP3) (*PreciseEphemerisInterpolant, error) {
	if sp3 == nil || sp3.handle == nil {
		return nil, ErrClosed
	}
	var out *C.SidereonPreciseEphemerisInterpolant
	err := sp3.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_precise_ephemeris_interpolant_from_sp3((*C.SidereonSp3)(p), &out))
		})
	})
	runtime.KeepAlive(sp3)
	if err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_precise_ephemeris_interpolant_free(out) })
		}
		return nil, err
	}
	return newPreciseInterpolant(out)
}

func (i *PreciseEphemerisInterpolant) Close() error {
	if i == nil || i.handle == nil {
		return nil
	}
	return i.handle.close()
}

func preciseObservableStates(h *positioningHandle, satellites []string, epochs []float64, shared bool, epoch float64) ([]NativeObservableStateRow, error) {
	if len(satellites) != len(epochs) && !shared {
		return nil, errors.New("sidereon: satellite and epoch lengths differ")
	}
	var result []NativeObservableStateRow
	err := h.with(func(pointer unsafe.Pointer) error {
		return withCStringArray(satellites, func(satPtr **C.char, count C.size_t) error {
			n := len(satellites)
			positionCount, err := checkedNativeProduct(n, 3, "precise observable-state position")
			if err != nil {
				return err
			}
			positions := make([]C.double, positionCount)
			clocks := make([]C.double, n)
			hasClocks := make([]C.bool, n)
			elements := make([]C.enum_SidereonObservableStateElementStatus, n)
			statuses := make([]C.enum_SidereonStatus, n)
			var epochMemory unsafe.Pointer
			if !shared {
				epochMemory, err = checkedNativeMalloc(n, unsafe.Sizeof(C.double(0)))
				if err != nil {
					return err
				}
				if epochMemory != nil {
					defer C.free(epochMemory)
				}
				epochValues := unsafe.Slice((*C.double)(epochMemory), n)
				for j := range epochs {
					epochValues[j] = C.double(epochs[j])
				}
			}
			var pp, cp, ep *C.double
			var hp *C.bool
			var esp *C.enum_SidereonObservableStateElementStatus
			var sp *C.enum_SidereonStatus
			if n != 0 {
				pp, cp, hp = &positions[0], &clocks[0], &hasClocks[0]
				esp, sp = &elements[0], &statuses[0]
				if !shared {
					ep = (*C.double)(epochMemory)
				}
			}
			var status uint32
			if shared {
				status = uint32(C.sidereon_precise_ephemeris_interpolant_observable_states_at_shared_j2000_s((*C.SidereonPreciseEphemerisInterpolant)(pointer), (**C.char)(satPtr), count, C.double(epoch), pp, cp, hp, esp, sp))
			} else {
				status = uint32(C.sidereon_precise_ephemeris_interpolant_observable_states_at_j2000_s((*C.SidereonPreciseEphemerisInterpolant)(pointer), (**C.char)(satPtr), ep, count, pp, cp, hp, esp, sp))
			}
			if err := statusErrorLocked(status); err != nil {
				return err
			}
			result = make([]NativeObservableStateRow, n)
			for j := range result {
				if err := validateObservableStateElementStatusValue(uint32(elements[j])); err != nil {
					return err
				}
				if err := validateObservableStateResultStatusValue(uint32(statuses[j])); err != nil {
					return err
				}
				result[j].ClockS, result[j].HasClock = float64(clocks[j]), bool(hasClocks[j])
				result[j].ElementStatus, result[j].ResultStatus = uint32(elements[j]), uint32(statuses[j])
				for axis := range result[j].Position {
					result[j].Position[axis] = float64(positions[j*3+axis])
				}
			}
			return nil
		})
	})
	return result, err
}

func (s *PreciseEphemerisSamples) ObservableStates(satellites []string, epochs []float64) ([]NativeObservableStateRow, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	// The sample-backed route is intentionally called directly rather than
	// routing through interpolation, preserving its native error/status arrays.
	var result []NativeObservableStateRow
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		return preciseSourceObservable(pointer, satellites, epochs, false, 0, &result)
	})
	runtime.KeepAlive(s)
	return result, err
}

func (s *PreciseEphemerisSamples) ObservableStatesShared(satellites []string, epoch float64) ([]NativeObservableStateRow, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var result []NativeObservableStateRow
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		return preciseSourceObservable(pointer, satellites, nil, true, epoch, &result)
	})
	runtime.KeepAlive(s)
	return result, err
}

func preciseSourceObservable(source unsafe.Pointer, satellites []string, epochs []float64, shared bool, epoch float64, result *[]NativeObservableStateRow) error {
	if !shared && len(satellites) != len(epochs) {
		return errors.New("sidereon: satellite and epoch lengths differ")
	}
	return withCStringArray(satellites, func(satPtr **C.char, count C.size_t) error {
		n := len(satellites)
		positionCount, err := checkedNativeProduct(n, 3, "precise observable-state position")
		if err != nil {
			return err
		}
		positions := make([]C.double, positionCount)
		clocks := make([]C.double, n)
		hasClocks := make([]C.bool, n)
		elements := make([]C.enum_SidereonObservableStateElementStatus, n)
		statuses := make([]C.enum_SidereonStatus, n)
		var epochMemory unsafe.Pointer
		if !shared {
			epochMemory, err = checkedNativeMalloc(n, unsafe.Sizeof(C.double(0)))
			if err != nil {
				return err
			}
			if epochMemory != nil {
				defer C.free(epochMemory)
			}
			epochValues := unsafe.Slice((*C.double)(epochMemory), n)
			for j := range epochs {
				epochValues[j] = C.double(epochs[j])
			}
		}
		var pp, cp, ep *C.double
		var hp *C.bool
		var esp *C.enum_SidereonObservableStateElementStatus
		var sp *C.enum_SidereonStatus
		if n != 0 {
			pp, cp, hp, esp, sp = &positions[0], &clocks[0], &hasClocks[0], &elements[0], &statuses[0]
			if !shared {
				ep = (*C.double)(epochMemory)
			}
		}
		var status C.enum_SidereonStatus
		if shared {
			status = C.sidereon_precise_ephemeris_samples_observable_states_at_shared_j2000_s((*C.SidereonPreciseEphemerisSamples)(source), (**C.char)(satPtr), count, C.double(epoch), pp, cp, hp, esp, sp)
		} else {
			status = C.sidereon_precise_ephemeris_samples_observable_states_at_j2000_s((*C.SidereonPreciseEphemerisSamples)(source), (**C.char)(satPtr), ep, count, pp, cp, hp, esp, sp)
		}
		if err := statusErrorLocked(uint32(status)); err != nil {
			return err
		}
		out := make([]NativeObservableStateRow, n)
		for j := range out {
			if err := validateObservableStateElementStatusValue(uint32(elements[j])); err != nil {
				return err
			}
			if err := validateObservableStateResultStatusValue(uint32(statuses[j])); err != nil {
				return err
			}
			out[j].ClockS, out[j].HasClock = float64(clocks[j]), bool(hasClocks[j])
			out[j].ElementStatus, out[j].ResultStatus = uint32(elements[j]), uint32(statuses[j])
			for axis := range out[j].Position {
				out[j].Position[axis] = float64(positions[j*3+axis])
			}
		}
		*result = out
		return nil
	})
}

func (i *PreciseEphemerisInterpolant) ObservableStates(satellites []string, epochs []float64) ([]NativeObservableStateRow, error) {
	if i == nil || i.handle == nil {
		return nil, ErrClosed
	}
	return preciseObservableStates(i.handle, satellites, epochs, false, 0)
}

func (i *PreciseEphemerisInterpolant) ObservableStatesShared(satellites []string, epoch float64) ([]NativeObservableStateRow, error) {
	if i == nil || i.handle == nil {
		return nil, ErrClosed
	}
	return preciseObservableStates(i.handle, satellites, nil, true, epoch)
}

func (s *PreciseEphemerisSamples) Sample(satellites []string, start, stop, step float64) ([]NativeEphemerisSampleRow, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	return preciseSampleRows(s.handle, satellites, start, stop, step)
}

type NativeRangePrediction struct {
	GeometricRangeM       float64
	HasSatelliteClock     bool
	SatelliteClockS       float64
	TransmitTimeJ2000S    float64
	SatellitePositionECEF [3]float64
}

func (s *PreciseEphemerisSamples) PredictRanges(requests []NativePredictRequest, options *NativeObservablesOptions) ([]NativeRangePrediction, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	return precisePredictRanges(s.handle, false, requests, options)
}

func precisePredictRanges(h *positioningHandle, _ bool, requests []NativePredictRequest, options *NativeObservablesOptions) ([]NativeRangePrediction, error) {
	var result []NativeRangePrediction
	err := h.with(func(pointer unsafe.Pointer) error {
		n := len(requests)
		count, err := checkedNativeSize(n)
		if err != nil {
			return err
		}
		requestMemory, err := checkedNativeMalloc(n, unsafe.Sizeof(C.SidereonRangePredictionRequest{}))
		if err != nil {
			return err
		}
		if requestMemory != nil {
			defer C.free(requestMemory)
		}
		rows := unsafe.Slice((*C.SidereonRangePredictionRequest)(requestMemory), n)
		ids := make([]string, n)
		for j := range requests {
			ids[j] = requests[j].SatelliteID
		}
		return withCStringArray(ids, func(idPtr **C.char, _ C.size_t) error {
			idValues := unsafe.Slice(idPtr, n)
			for j, request := range requests {
				rows[j].sat_id = idValues[j]
				for axis := 0; axis < 3; axis++ {
					rows[j].receiver_ecef_m[axis] = C.double(request.ReceiverECEF[axis])
				}
				rows[j].t_rx_j2000_s = C.double(request.TRxJ2000S)
			}
			outputs := make([]C.SidereonRangePrediction, n)
			var optionsPointer *C.SidereonObservablesOptions
			var optionsMemory unsafe.Pointer
			if options != nil {
				optionsMemory, err = checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonObservablesOptions{}))
				if err != nil {
					return err
				}
				defer C.free(optionsMemory)
				optionsPointer = (*C.SidereonObservablesOptions)(optionsMemory)
				*optionsPointer = cObservablesOptions(*options)
			}
			var requestPointer *C.SidereonRangePredictionRequest
			var output *C.SidereonRangePrediction
			if n != 0 {
				requestPointer, output = &rows[0], &outputs[0]
			}
			if err := callStatus(func() uint32 {
				return uint32(C.sidereon_precise_ephemeris_samples_predict_ranges((*C.SidereonPreciseEphemerisSamples)(pointer), requestPointer, count, optionsPointer, output))
			}); err != nil {
				return err
			}
			result = make([]NativeRangePrediction, n)
			for j, row := range outputs {
				result[j].GeometricRangeM = float64(row.geometric_range_m)
				result[j].HasSatelliteClock, result[j].SatelliteClockS = bool(row.has_sat_clock_s), float64(row.sat_clock_s)
				result[j].TransmitTimeJ2000S = float64(row.transmit_time_j2000_s)
				for axis := range result[j].SatellitePositionECEF {
					result[j].SatellitePositionECEF[axis] = float64(row.sat_pos_ecef_m[axis])
				}
			}
			return nil
		})
	})
	runtime.KeepAlive(requests)
	return result, err
}

func preciseSampleRows(h *positioningHandle, satellites []string, start, stop, step float64) ([]NativeEphemerisSampleRow, error) {
	if len(satellites) == 0 {
		return nil, invalidArgument("precise sample satellite list must not be empty")
	}
	var result []NativeEphemerisSampleRow
	err := h.with(func(pointer unsafe.Pointer) error {
		return withCStringArray(satellites, func(satPtr **C.char, count C.size_t) error {
			var written, required C.size_t
			invoke := func(out *C.SidereonEphemerisSampleRow, length C.size_t) uint32 {
				return uint32(C.sidereon_precise_ephemeris_samples_sample((*C.SidereonPreciseEphemerisSamples)(pointer), (**C.char)(satPtr), count, C.double(start), C.double(stop), C.double(step), out, length, &written, &required))
			}
			if err := callStatus(func() uint32 { return invoke(nil, 0) }); err != nil {
				return err
			}
			n, err := validateNativeQuery("precise ephemeris sample", uint64(written), uint64(required))
			if err != nil {
				return err
			}
			if _, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonEphemerisSampleRow{})); err != nil {
				return err
			}
			rows := make([]C.SidereonEphemerisSampleRow, n)
			var out *C.SidereonEphemerisSampleRow
			if n != 0 {
				out = &rows[0]
			}
			written, required = 0, 0
			if err := callStatus(func() uint32 { return invoke(out, C.size_t(n)) }); err != nil {
				return err
			}
			// The query and fill are one locked operation, and both counts must
			// describe the same shape.
			if _, err := validateTwoPassCounts("precise ephemeris sample", n, n, uint64(written), uint64(required)); err != nil {
				return err
			}
			writtenCount, err := checkedNativeCount(uint64(written))
			if err != nil {
				return err
			}
			result = make([]NativeEphemerisSampleRow, writtenCount)
			for j := range result {
				row := rows[j]
				if err := validateEphemerisSampleStatusValue(uint32(row.status)); err != nil {
					return err
				}
				result[j] = NativeEphemerisSampleRow{SatelliteID: tokenFromC(row.sat_id), EpochJ2000S: float64(row.epoch_j2000_s), Status: uint32(row.status), HasPosition: bool(row.has_position_ecef_m), HasClock: bool(row.has_clock_s), ClockS: float64(row.clock_s)}
				for axis := range result[j].Position {
					result[j].Position[axis] = float64(row.position_ecef_m[axis])
				}
			}
			return nil
		})
	})
	runtime.KeepAlive(h)
	return result, err
}

// PreciseInterpolantArtifact owns a C artifact reader. Go always enters the
// byte-backed owned route; the borrowed route is deliberately not used because
// a returned handle may retain its input beyond the call.
type PreciseInterpolantArtifact struct {
	_      noCopy
	handle *positioningHandle
}

func releasePreciseArtifact(p unsafe.Pointer) {
	withCThread(func() { C.sidereon_precise_interpolant_artifact_free((*C.SidereonPreciseInterpolantArtifact)(p)) })
}

func OpenPreciseInterpolantArtifact(data []byte) (*PreciseInterpolantArtifact, uint32, error) {
	if _, err := checkedNativeSize(len(data)); err != nil {
		return nil, 0, err
	}
	var out *C.SidereonPreciseInterpolantArtifact
	var artifactError C.enum_SidereonPreciseInterpolantArtifactErrorKind
	var err error
	withCThread(func() {
		var input unsafe.Pointer
		if len(data) != 0 {
			input = C.CBytes(data)
			if input == nil {
				err = errors.New("sidereon: unable to allocate native artifact input")
				return
			}
			defer C.free(input)
		}
		nativeErr := statusErrorLocked(uint32(C.sidereon_precise_interpolant_artifact_open_owned((*C.uint8_t)(input), C.size_t(len(data)), &artifactError, &out)))
		if enumErr := validatePreciseInterpolantArtifactError(uint32(artifactError)); enumErr != nil {
			artifactError = C.enum_SidereonPreciseInterpolantArtifactErrorKind(PreciseInterpolantArtifactErrorNoneValue)
			if nativeErr == nil {
				err = enumErr
			} else {
				err = nativeErr
			}
		} else {
			err = nativeErr
		}
		if err != nil && out != nil {
			C.sidereon_precise_interpolant_artifact_free(out)
			out = nil
		}
	})
	runtime.KeepAlive(data)
	if err != nil {
		return nil, uint32(artifactError), err
	}
	if out == nil {
		return nil, uint32(artifactError), missingNativeHandle("precise artifact")
	}
	return &PreciseInterpolantArtifact{handle: newPositioningHandle(unsafe.Pointer(out), releasePreciseArtifact)}, uint32(artifactError), nil
}

func (a *PreciseInterpolantArtifact) Close() error {
	if a == nil || a.handle == nil {
		return nil
	}
	return a.handle.close()
}

func (a *PreciseInterpolantArtifact) Checksum64() (uint64, error) {
	if a == nil || a.handle == nil {
		return 0, ErrClosed
	}
	var out C.uint64_t
	err := a.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_precise_interpolant_artifact_handle_checksum64((*C.SidereonPreciseInterpolantArtifact)(p), &out))
		})
	})
	runtime.KeepAlive(a)
	return uint64(out), err
}

func PreciseArtifactChecksum64(data []byte) (uint64, error) {
	var out C.uint64_t
	err := withInputError(data, func(p *C.uint8_t, n C.size_t) error {
		return callStatus(func() uint32 { return uint32(C.sidereon_precise_interpolant_artifact_checksum64(p, n, &out)) })
	})
	return uint64(out), err
}

func (a *PreciseInterpolantArtifact) DigestProvenance() (uint32, error) {
	if a == nil || a.handle == nil {
		return 0, ErrClosed
	}
	var out C.enum_SidereonDigestProvenance
	err := a.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_precise_interpolant_artifact_digest_provenance((*C.SidereonPreciseInterpolantArtifact)(p), &out))
		})
	})
	runtime.KeepAlive(a)
	if enumErr := validateDigestProvenance(uint32(out)); enumErr != nil {
		out = 0
		if err == nil {
			err = enumErr
		}
	}
	return uint32(out), err
}

func (a *PreciseInterpolantArtifact) Verify() (uint32, error) {
	if a == nil || a.handle == nil {
		return 0, ErrClosed
	}
	var out C.enum_SidereonPreciseInterpolantArtifactErrorKind
	err := a.handle.withExclusive(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_precise_interpolant_artifact_verify((*C.SidereonPreciseInterpolantArtifact)(p), &out))
		})
	})
	runtime.KeepAlive(a)
	if enumErr := validatePreciseInterpolantArtifactError(uint32(out)); enumErr != nil {
		out = 0
		if err == nil {
			err = enumErr
		}
	}
	return uint32(out), err
}

func (a *PreciseInterpolantArtifact) State(satellite string, epoch float64) (SP3State, error) {
	if a == nil || a.handle == nil {
		return SP3State{}, ErrClosed
	}
	var out C.SidereonSp3State
	err := a.handle.with(func(p unsafe.Pointer) error {
		return withToken(satellite, "artifact satellite ID", func(id *C.char) uint32 {
			return uint32(C.sidereon_precise_interpolant_artifact_state((*C.SidereonPreciseInterpolantArtifact)(p), id, C.double(epoch), &out))
		})
	})
	runtime.KeepAlive(a)
	if err != nil {
		return SP3State{}, err
	}
	return sp3StateFromC(out), nil
}

func (a *PreciseInterpolantArtifact) Satellites() ([]string, error) {
	if a == nil || a.handle == nil {
		return nil, ErrClosed
	}
	var result []string
	err := a.handle.with(func(pointer unsafe.Pointer) error {
		var written, required C.size_t
		invoke := func(out *C.SidereonSatelliteToken, n C.size_t) uint32 {
			return uint32(C.sidereon_precise_interpolant_artifact_satellites((*C.SidereonPreciseInterpolantArtifact)(pointer), out, n, &written, &required))
		}
		if err := callStatus(func() uint32 { return invoke(nil, 0) }); err != nil {
			return err
		}
		count, err := validateNativeQuery("precise artifact satellites", uint64(written), uint64(required))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonSatelliteToken{})); err != nil {
			return err
		}
		values := make([]C.SidereonSatelliteToken, count)
		var out *C.SidereonSatelliteToken
		if count != 0 {
			out = &values[0]
		}
		written, required = 0, 0
		if err := callStatus(func() uint32 { return invoke(out, C.size_t(count)) }); err != nil {
			return err
		}
		writtenCount, err := validateNativeOutput("precise artifact satellites", count, uint64(written), uint64(required))
		if err != nil {
			return err
		}
		result = make([]string, writtenCount)
		for i := range result {
			result[i] = tokenFromC(values[i])
		}
		return nil
	})
	runtime.KeepAlive(a)
	return result, err
}

func sp3StateFromC(value C.SidereonSp3State) SP3State {
	var out SP3State
	for j := range out.PositionM {
		out.PositionM[j] = float64(value.position_m[j])
		out.VelocityMPerS[j] = float64(value.velocity_m_s[j])
	}
	out.HasClock, out.ClockS = bool(value.has_clock_s), float64(value.clock_s)
	out.HasVelocity, out.HasClockRate = bool(value.has_velocity_m_s), bool(value.has_clock_rate_s_s)
	out.ClockRateSPerS = float64(value.clock_rate_s_s)
	out.ClockEvent, out.ClockPredicted, out.Maneuver, out.OrbitPredicted = bool(value.clock_event), bool(value.clock_predicted), bool(value.maneuver), bool(value.orbit_predicted)
	return out
}

// StalenessPolicy is a value type exactly matching the C ABI. These helpers
// intentionally perform no local normalization; C owns all numeric semantics.
type StalenessPolicy struct{ MaxStalenessS float64 }

func StalenessPolicyDays(days float64) StalenessPolicy {
	v := C.sidereon_staleness_policy_days(C.double(days))
	return StalenessPolicy{MaxStalenessS: float64(v.max_staleness_s)}
}

func StalenessPolicySeconds(seconds float64) StalenessPolicy {
	v := C.sidereon_staleness_policy_seconds(C.double(seconds))
	return StalenessPolicy{MaxStalenessS: float64(v.max_staleness_s)}
}

func StalenessPolicyDefault() StalenessPolicy {
	v := C.sidereon_staleness_policy_default()
	return StalenessPolicy{MaxStalenessS: float64(v.max_staleness_s)}
}
