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

type Sp3ContinuityOptions struct {
	SpeedBoundKind           uint32
	OrbitClass               uint32
	ExplicitMaxSpeedMPS      float64
	ResidualToleranceEnabled bool
	ResidualToleranceM       float64
	GapThresholdFactor       float64
}

func sp3ContinuityOptionsFromC(v C.SidereonSp3ContinuityOptions) Sp3ContinuityOptions {
	return Sp3ContinuityOptions{SpeedBoundKind: uint32(v.speed_bound_kind), OrbitClass: uint32(v.orbit_class), ExplicitMaxSpeedMPS: float64(v.explicit_max_speed_m_s), ResidualToleranceEnabled: v.residual_tolerance_enabled != 0, ResidualToleranceM: float64(v.residual_tolerance_m), GapThresholdFactor: float64(v.gap_threshold_factor)}
}

func sp3ContinuityOptionsToC(v Sp3ContinuityOptions) C.SidereonSp3ContinuityOptions {
	var enabled C.uint8_t
	if v.ResidualToleranceEnabled {
		enabled = 1
	}
	return C.SidereonSp3ContinuityOptions{speed_bound_kind: C.uint32_t(v.SpeedBoundKind), orbit_class: C.uint32_t(v.OrbitClass), explicit_max_speed_m_s: C.double(v.ExplicitMaxSpeedMPS), residual_tolerance_enabled: enabled, residual_tolerance_m: C.double(v.ResidualToleranceM), gap_threshold_factor: C.double(v.GapThresholdFactor)}
}

func DefaultSp3ContinuityOptions(orbitClass uint32) (result Sp3ContinuityOptions, err error) {
	var options C.SidereonSp3ContinuityOptions
	err = callStatusWithSp3Diagnostics(func() uint32 {
		return uint32(C.sidereon_sp3_continuity_options_for_orbit_class(C.uint32_t(orbitClass), &options))
	})
	if err == nil {
		result = sp3ContinuityOptionsFromC(options)
	}
	return result, err
}

func (s *SP3) ContinuityReportJSON(options Sp3ContinuityOptions) (result []byte, err error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	cOptions := sp3ContinuityOptionsToC(options)
	err = s.handle.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			var e error
			result, e = copyNativeBytesLockedWithStatus("SP3 continuity report", func(out *C.uint8_t, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_sp3_continuity_report_json((*C.SidereonSp3)(p), &cOptions, out, n, w, r)
			}, statusSp3ErrorLocked)
			return e
		})
	})
	runtime.KeepAlive(s)
	return result, err
}

func (s *SP3) ContinuityVerdictJSONWithPolicy(options Sp3ContinuityOptions, from, through float64) (result []byte, err error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	cOptions := sp3ContinuityOptionsToC(options)
	err = s.handle.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			var e error
			result, e = copyNativeBytesLockedWithStatus("SP3 continuity verdict", func(out *C.uint8_t, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_sp3_continuity_verdict_json_with_options((*C.SidereonSp3)(p), &cOptions, C.double(from), C.double(through), out, n, w, r)
			}, statusSp3ErrorLocked)
			return e
		})
	})
	runtime.KeepAlive(s)
	return result, err
}
