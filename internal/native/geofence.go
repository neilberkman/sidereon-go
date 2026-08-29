//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
*/
import "C"

import (
	"errors"
	"runtime"
	"unsafe"
)

type GeofenceUncertainty struct {
	Kind       uint32
	Covariance [9]float64
	RadiusM    float64
}

type GeofenceProbabilityOptions struct {
	Method uint32
}

type GeofencePositionEstimate struct {
	Position    Geodetic
	Uncertainty GeofenceUncertainty
}

type GeofenceHysteresis struct {
	EnterConfidence float64
	LeaveConfidence float64
}

type GeofenceCrossingEvent struct {
	SampleIndex       int
	Kind              uint32
	InsideProbability float64
}

type Geofence struct {
	handle *positioningHandle
}

func DefaultGeofenceProbabilityOptions() (GeofenceProbabilityOptions, error) {
	var output C.SidereonGeofenceProbabilityOptions
	err := callStatus(func() uint32 { return C.sidereon_geofence_probability_options_init(&output) })
	return GeofenceProbabilityOptions{Method: uint32(output.method)}, err
}

func DefaultGeofenceHysteresis() (GeofenceHysteresis, error) {
	var output C.SidereonGeofenceHysteresis
	err := callStatus(func() uint32 { return C.sidereon_geofence_hysteresis_init(&output) })
	return GeofenceHysteresis{EnterConfidence: float64(output.enter_confidence), LeaveConfidence: float64(output.leave_confidence)}, err
}

func GeofenceCreate(vertices []Geodetic) (*Geofence, uint32, error) {
	if _, err := checkedNativeAllocationSize(len(vertices), unsafe.Sizeof(C.SidereonGeodetic{})); err != nil {
		return nil, 0, err
	}
	cVertices := make([]C.SidereonGeodetic, len(vertices))
	for i, vertex := range vertices {
		cVertices[i] = cGeodetic(vertex)
	}
	var pointer *C.SidereonGeofence
	var detail uint32
	var operationErr error
	withCThread(func() {
		var input *C.SidereonGeodetic
		if len(cVertices) != 0 {
			input = &cVertices[0]
		}
		operationErr = statusErrorLocked(C.sidereon_geofence_create(input, C.size_t(len(cVertices)), &detail, &pointer))
	})
	if operationErr != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_geofence_free(pointer) })
		}
		return nil, uint32(detail), operationErr
	}
	if pointer == nil {
		return nil, uint32(detail), errors.New("sidereon: native geofence constructor returned no handle")
	}
	return &Geofence{handle: newPositioningHandle(unsafe.Pointer(pointer), func(value unsafe.Pointer) {
		C.sidereon_geofence_free((*C.SidereonGeofence)(value))
	})}, uint32(detail), nil
}

func (f *Geofence) Close() error {
	if f == nil || f.handle == nil {
		return nil
	}
	return f.handle.close()
}

func (f *Geofence) Contains(position Geodetic) (bool, uint32, error) {
	if f == nil || f.handle == nil {
		return false, 0, ErrClosed
	}
	var output C.bool
	var detail uint32
	err := f.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_geofence_contains((*C.SidereonGeofence)(pointer), cGeodetic(position), &detail, &output)
		})
	})
	runtime.KeepAlive(f)
	return bool(output), uint32(detail), err
}

func (f *Geofence) DistanceToBoundary(position Geodetic) (float64, uint32, error) {
	if f == nil || f.handle == nil {
		return 0, 0, ErrClosed
	}
	var output C.double
	var detail uint32
	err := f.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_geofence_distance_to_boundary((*C.SidereonGeofence)(pointer), cGeodetic(position), &detail, &output)
		})
	})
	runtime.KeepAlive(f)
	return float64(output), uint32(detail), err
}

func cGeofenceUncertainty(value GeofenceUncertainty) C.SidereonGeofenceUncertainty {
	var output C.SidereonGeofenceUncertainty
	output.kind = C.uint32_t(value.Kind)
	output.radius_m = C.double(value.RadiusM)
	for i := range value.Covariance {
		output.covariance_m2[i] = C.double(value.Covariance[i])
	}
	return output
}

func (f *Geofence) ContainmentProbability(position Geodetic, uncertainty GeofenceUncertainty) (float64, uint32, error) {
	return f.containmentProbability(position, uncertainty, nil)
}

func (f *Geofence) ContainmentProbabilityWithOptions(position Geodetic, uncertainty GeofenceUncertainty, options GeofenceProbabilityOptions) (float64, uint32, error) {
	return f.containmentProbability(position, uncertainty, &options)
}

func (f *Geofence) containmentProbability(position Geodetic, uncertainty GeofenceUncertainty, options *GeofenceProbabilityOptions) (float64, uint32, error) {
	if f == nil || f.handle == nil {
		return 0, 0, ErrClosed
	}
	cUncertainty := cGeofenceUncertainty(uncertainty)
	var cOptions C.SidereonGeofenceProbabilityOptions
	if options != nil {
		cOptions.method = C.uint32_t(options.Method)
	}
	var output C.double
	var detail uint32
	err := f.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			if options == nil {
				return C.sidereon_geofence_containment_probability((*C.SidereonGeofence)(pointer), cGeodetic(position), &cUncertainty, &detail, &output)
			}
			return C.sidereon_geofence_containment_probability_with_options((*C.SidereonGeofence)(pointer), cGeodetic(position), &cUncertainty, &cOptions, &detail, &output)
		})
	})
	runtime.KeepAlive(f)
	return float64(output), uint32(detail), err
}

func cGeofenceEstimate(value GeofencePositionEstimate) C.SidereonGeofencePositionEstimate {
	return C.SidereonGeofencePositionEstimate{position: cGeodetic(value.Position), uncertainty: cGeofenceUncertainty(value.Uncertainty)}
}

func copyGeofenceEvents(call func(*C.SidereonGeofenceCrossingEvent, C.size_t, *C.size_t, *C.size_t) uint32) ([]GeofenceCrossingEvent, error) {
	var result []GeofenceCrossingEvent
	var operationErr error
	withCThread(func() {
		var written, required C.size_t
		if operationErr = statusErrorLocked(call(nil, 0, &written, &required)); operationErr != nil {
			return
		}
		queryWritten, err := checkedNativeCount(uint64(written))
		if err != nil {
			operationErr = err
			return
		}
		if queryWritten != 0 {
			operationErr = errors.New("sidereon: native geofence event query wrote data")
			return
		}
		requiredCount, err := checkedNativeCount(uint64(required))
		if err != nil {
			operationErr = err
			return
		}
		if _, err := checkedNativeAllocationSize(requiredCount, unsafe.Sizeof(C.SidereonGeofenceCrossingEvent{})); err != nil {
			operationErr = err
			return
		}
		buffer := make([]C.SidereonGeofenceCrossingEvent, requiredCount)
		var output *C.SidereonGeofenceCrossingEvent
		if len(buffer) != 0 {
			output = &buffer[0]
		}
		if operationErr = statusErrorLocked(call(output, C.size_t(len(buffer)), &written, &required)); operationErr != nil {
			return
		}
		writtenCount, err := validateTwoPassCounts("geofence event output", len(buffer), len(buffer), uint64(written), uint64(required))
		if err != nil {
			operationErr = err
			return
		}
		result = make([]GeofenceCrossingEvent, writtenCount)
		for i := range result {
			sampleIndex, err := checkedNativeCount(uint64(buffer[i].sample_index))
			if err != nil {
				operationErr = err
				return
			}
			result[i] = GeofenceCrossingEvent{SampleIndex: sampleIndex, Kind: uint32(buffer[i].kind), InsideProbability: float64(buffer[i].inside_probability)}
		}
	})
	return result, operationErr
}

func (f *Geofence) CrossingProbability(samples []GeofencePositionEstimate, hysteresis GeofenceHysteresis) ([]GeofenceCrossingEvent, uint32, error) {
	return f.crossingProbability(samples, hysteresis, nil)
}

func (f *Geofence) CrossingProbabilityWithOptions(samples []GeofencePositionEstimate, hysteresis GeofenceHysteresis, options GeofenceProbabilityOptions) ([]GeofenceCrossingEvent, uint32, error) {
	return f.crossingProbability(samples, hysteresis, &options)
}

func (f *Geofence) crossingProbability(samples []GeofencePositionEstimate, hysteresis GeofenceHysteresis, options *GeofenceProbabilityOptions) ([]GeofenceCrossingEvent, uint32, error) {
	if f == nil || f.handle == nil {
		return nil, 0, ErrClosed
	}
	if _, err := checkedNativeAllocationSize(len(samples), unsafe.Sizeof(C.SidereonGeofencePositionEstimate{})); err != nil {
		return nil, 0, err
	}
	cSamples := make([]C.SidereonGeofencePositionEstimate, len(samples))
	for i, sample := range samples {
		cSamples[i] = cGeofenceEstimate(sample)
	}
	cHysteresis := C.SidereonGeofenceHysteresis{enter_confidence: C.double(hysteresis.EnterConfidence), leave_confidence: C.double(hysteresis.LeaveConfidence)}
	var cOptions C.SidereonGeofenceProbabilityOptions
	if options != nil {
		cOptions.method = C.uint32_t(options.Method)
	}
	var detail uint32
	var output []GeofenceCrossingEvent
	err := f.handle.with(func(pointer unsafe.Pointer) error {
		var operationErr error
		output, operationErr = copyGeofenceEvents(func(out *C.SidereonGeofenceCrossingEvent, length C.size_t, written, required *C.size_t) uint32 {
			var samplePointer *C.SidereonGeofencePositionEstimate
			if len(cSamples) != 0 {
				samplePointer = &cSamples[0]
			}
			if options == nil {
				return C.sidereon_geofence_crossing_probability((*C.SidereonGeofence)(pointer), samplePointer, C.size_t(len(cSamples)), &cHysteresis, &detail, out, length, written, required)
			}
			return C.sidereon_geofence_crossing_probability_with_options((*C.SidereonGeofence)(pointer), samplePointer, C.size_t(len(cSamples)), &cHysteresis, &cOptions, &detail, out, length, written, required)
		})
		return operationErr
	})
	runtime.KeepAlive(f)
	return output, uint32(detail), err
}

func ObservabilityTierLabel(tier uint32) ([]byte, error) {
	return copyLabel(func(out *C.uint8_t, length C.size_t, written, required *C.size_t) uint32 {
		return C.sidereon_observability_tier_label(C.uint32_t(tier), out, length, written, required)
	})
}
