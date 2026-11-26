//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"unsafe"
)

func IntegratorOptionsDefault() IntegratorOptions {
	value := C.sidereon_integrator_options_default()
	return orbitIntegratorOptionsFromC(value)
}

func VariationalEpochFromTDBSeconds(initialTDBJ2000S float64, tdbJ2000S float64) VariationalEpoch {
	value := C.sidereon_variational_epoch_from_tdb_seconds(C.double(initialTDBJ2000S), C.double(tdbJ2000S))
	return orbitVariationalEpochFromC(value)
}

func VariationalTrajectoryCount(trajectory *VariationalTrajectory) (uint64, error) {
	return withOrbitHandles([]*positioningHandle{trajectory.orbitHandle()}, func(pointers []unsafe.Pointer) (uint64, error) {
		var value C.size_t
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_variational_trajectory_count((*C.SidereonVariationalTrajectory)(pointers[0]), &value))
		})
		return uint64(value), err
	})
}

func VariationalTrajectorySensitivity(trajectory *VariationalTrajectory, index uint64) ([]float64, error) {
	return withOrbitHandles([]*positioningHandle{trajectory.orbitHandle()}, func(pointers []unsafe.Pointer) ([]float64, error) {
		raw, err := orbitReadRows("variational trajectory sensitivity", func(out *C.double, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_variational_trajectory_sensitivity((*C.SidereonVariationalTrajectory)(pointers[0]), C.size_t(index), out, capacity, written, required)
		})
		if err != nil {
			return nil, err
		}
		return orbitMap(raw, func(item C.double) float64 { return float64(item) }), nil
	})
}

func VariationalTrajectoryState(trajectory *VariationalTrajectory, index uint64) (VariationalState, error) {
	return withOrbitHandles([]*positioningHandle{trajectory.orbitHandle()}, func(pointers []unsafe.Pointer) (VariationalState, error) {
		var value C.SidereonVariationalState
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_variational_trajectory_state((*C.SidereonVariationalTrajectory)(pointers[0]), C.size_t(index), &value))
		})
		return orbitVariationalStateFromC(value), err
	})
}
