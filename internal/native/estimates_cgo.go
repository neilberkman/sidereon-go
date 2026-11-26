//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
*/
import "C"

import (
	"runtime"
	"unsafe"
)

func orbitNamedValue(label string, call func(*C.uint8_t, C.size_t, *C.size_t, *C.size_t, *C.double) C.enum_SidereonStatus) (ForceParameter, error) {
	var value C.double
	name, err := orbitText(label, func(out *C.uint8_t, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return call(out, capacity, written, required, &value)
	})
	if err != nil {
		return ForceParameter{}, err
	}
	return ForceParameter{Name: name, Value: float64(value)}, nil
}

func ForceModelParameters(model *ForceModel) ([]ForceParameter, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) ([]ForceParameter, error) {
		model := (*C.SidereonForceModel)(pointers[0])
		var count C.size_t
		if err := callStatus(func() uint32 {
			return uint32(C.sidereon_force_model_parameter_count(model, &count))
		}); err != nil {
			return nil, err
		}
		length, err := sizeTToInt(count, "force parameter count")
		if err != nil {
			return nil, err
		}
		if _, err := checkedNativeAllocationSize(length, unsafe.Sizeof(ForceParameter{})); err != nil {
			return nil, err
		}
		values := make([]ForceParameter, length)
		for i := range values {
			value, err := orbitNamedValue("force parameter", func(out *C.uint8_t, capacity C.size_t, written, required *C.size_t, value *C.double) C.enum_SidereonStatus {
				return C.sidereon_force_model_parameter(model, C.size_t(i), out, capacity, written, required, value)
			})
			if err != nil {
				return nil, err
			}
			values[i] = value
		}
		return values, nil
	})
}

func DEEphemerisConstants(ephemeris *DEEphemeris) (map[string]float64, error) {
	return withOrbitHandles([]*positioningHandle{ephemeris.orbitHandle()}, func(pointers []unsafe.Pointer) (map[string]float64, error) {
		ephemeris := (*C.SidereonDeEphemeris)(pointers[0])
		var count C.size_t
		if err := callStatus(func() uint32 {
			return uint32(C.sidereon_de_ephemeris_constant_count(ephemeris, &count))
		}); err != nil {
			return nil, err
		}
		length, err := sizeTToInt(count, "DE constant count")
		if err != nil {
			return nil, err
		}
		values := make(map[string]float64, length)
		for i := range length {
			value, err := orbitNamedValue("DE constant", func(out *C.uint8_t, capacity C.size_t, written, required *C.size_t, value *C.double) C.enum_SidereonStatus {
				return C.sidereon_de_ephemeris_constant(ephemeris, C.size_t(i), out, capacity, written, required, value)
			})
			if err != nil {
				return nil, err
			}
			values[value.Name] = value.Value
		}
		return values, nil
	})
}

func CompositeForceModelNew(models []*ForceModel) (*ForceModel, error) {
	handles := orbitMap(models, (*ForceModel).orbitHandle)
	return withOrbitHandles(handles, func(pointers []unsafe.Pointer) (*ForceModel, error) {
		raw := orbitMap(pointers, func(pointer unsafe.Pointer) *C.SidereonForceModel {
			return (*C.SidereonForceModel)(pointer)
		})
		var value *C.SidereonForceModel
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_composite_force_model_new(unsafe.SliceData(raw), C.size_t(len(raw)), &value))
		})
		return newOrbitForceModel(value, err)
	})
}

func PropagateVariational(initial CartesianState, epochs []VariationalEpoch, equations VariationalEquations, options IntegratorOptions, context *PropagationContext) (*VariationalTrajectory, error) {
	rawEpochs := orbitMap(epochs, orbitVariationalEpochToC)
	indices := orbitMap(equations.SensitivityParameters, func(index uint64) C.size_t { return C.size_t(index) })
	var pins runtime.Pinner
	defer pins.Unpin()
	parameters := (*C.double)(unsafe.Pointer(unsafe.SliceData(equations.Parameters)))
	if parameters != nil {
		pins.Pin(parameters)
	}
	indexData := unsafe.SliceData(indices)
	if indexData != nil {
		pins.Pin(indexData)
	}
	return withOrbitContext([]*positioningHandle{equations.Force.orbitHandle()}, context, func(pointers []unsafe.Pointer, context *C.SidereonPropagationContext) (*VariationalTrajectory, error) {
		raw := C.SidereonVariationalEquations{
			force:                  (*C.SidereonForceModel)(pointers[0]),
			parameters:             parameters,
			parameter_count:        C.size_t(len(equations.Parameters)),
			sensitivity_parameters: indexData,
			sensitivity_count:      C.size_t(len(indices)),
		}
		var value *C.SidereonVariationalTrajectory
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_propagate_variational(cCartesianState(initial), unsafe.SliceData(rawEpochs), C.size_t(len(rawEpochs)), raw, orbitIntegratorOptionsToC(options), context, &value))
		})
		return newOrbitVariationalTrajectory(value, err)
	})
}

func orbitVariationalEpochToC(value VariationalEpoch) C.SidereonVariationalEpoch {
	return C.SidereonVariationalEpoch{elapsed_tt_s: C.double(value.ElapsedTTS), tdb_j2000_s: C.double(value.TDBJ2000S)}
}
