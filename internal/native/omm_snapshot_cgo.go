//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
*/
import "C"

import "unsafe"

// OMMFromSnapshotJSON constructs an owned OMM from a complete detached snapshot.
func OMMFromSnapshotJSON(data []byte) (*OMM, error) {
	if _, err := checkedNativeSize(len(data)); err != nil {
		return nil, err
	}
	input, err := copyNativeInput(data)
	if err != nil {
		return nil, err
	}
	defer freeNativeInput(input)
	var out *C.SidereonOmm
	err = callStatus(func() uint32 {
		return uint32(C.sidereon_omm_from_snapshot_json((*C.uint8_t)(input), C.size_t(len(data)), &out))
	})
	if err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_omm_free(out) })
		}
		return nil, err
	}
	return newOMM(out)
}

// SnapshotJSON returns every stored OMM field in a detached JSON snapshot,
// including per-block comments and the in-memory SGP4 side channels.
func (o *OMM) SnapshotJSON() ([]byte, error) {
	return copyHandleText(o.handle, func(pointer unsafe.Pointer, out *C.uint8_t, length C.size_t, written, required *C.size_t) uint32 {
		return uint32(C.sidereon_omm_snapshot_json((*C.SidereonOmm)(pointer), out, length, written, required))
	})
}

// ToElementSet applies the core OMM-to-SGP4 bridge and copies every output
// field, retaining optional values through explicit presence flags.
func (o *OMM) ToElementSet() (OMMElementSet, error) {
	var value C.SidereonOmmElementSet
	err := o.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_omm_to_element_set((*C.SidereonOmm)(pointer), &value))
		})
	})
	if err != nil {
		return OMMElementSet{}, err
	}
	return OMMElementSet{
		EpochWhole: float64(value.epoch_whole), EpochFraction: float64(value.epoch_fraction),
		BStar:                float64(value.bstar),
		MeanMotionDotPresent: bool(value.mean_motion_dot_present), MeanMotionDot: float64(value.mean_motion_dot),
		MeanMotionDoubleDotPresent: bool(value.mean_motion_double_dot_present), MeanMotionDoubleDot: float64(value.mean_motion_double_dot),
		Eccentricity: float64(value.eccentricity), ArgumentOfPerigeeDeg: float64(value.argument_of_perigee_deg),
		InclinationDeg: float64(value.inclination_deg), MeanAnomalyDeg: float64(value.mean_anomaly_deg),
		MeanMotionRevPerDay: float64(value.mean_motion_rev_per_day), RightAscensionDeg: float64(value.right_ascension_deg),
		CatalogNumberPresent: bool(value.catalog_number_present), CatalogNumber: uint32(value.catalog_number),
		OMMEpochDaysPresent: bool(value.omm_epoch_days_present), OMMEpochDays: float64(value.omm_epoch_days),
	}, nil
}
