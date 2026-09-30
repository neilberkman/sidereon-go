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

func pppSsrSolutionFromC(v C.SidereonSsrSolutionId) PppSsrSolutionID {
	return PppSsrSolutionID{Source: uint32(v.source), ProviderID: uint16(v.provider_id), SolutionID: uint8(v.solution_id)}
}
func pppSsrSignalFromC(v C.SidereonSsrSignalKey) PppSsrSignalKey {
	return PppSsrSignalKey{IsPhysical: bool(v.is_physical), System: uint32(v.system), Code: C.GoString(&v.code[0]), Source: uint32(v.source), Index: uint8(v.index)}
}
func pppSsrSignalReportFromC(v C.SidereonPppSsrSignalReport) PppSsrSignalReport {
	return PppSsrSignalReport{Present: bool(v.present), Signal: pppSsrSignalFromC(v.signal), Status: uint32(v.status), HasSourceSignal: bool(v.has_source_signal), SourceSignal: pppSsrSignalFromC(v.source_signal), HasBiasM: bool(v.has_bias_m), BiasM: float64(v.bias_m), HasBiasCycles: bool(v.has_bias_cycles), BiasCycles: float64(v.bias_cycles), HasSolution: bool(v.has_solution), Solution: pppSsrSolutionFromC(v.solution), HasIODSSR: bool(v.has_iod_ssr), IODSSR: uint8(v.iod_ssr), UnknownVariant: C.GoString(&v.unknown_variant[0])}
}
func pppTransmitTimeFailureFromC(v C.SidereonPppTransmitTimeFailure) PppTransmitTimeFailure {
	return PppTransmitTimeFailure{Kind: uint32(v.kind), HasTransmitTime: bool(v.has_transmit_time), TransmitTimeJ2000S: float64(v.transmit_time_j2000_s), HasApplied: bool(v.has_applied), Applied: pppSsrSolutionFromC(v.applied), HasSignal: bool(v.has_signal), Signal: pppSsrSignalFromC(v.signal), BiasStatus: uint32(v.bias_status), UnknownVariant: C.GoString(&v.unknown_variant[0])}
}
func pppSsrApplicationFromC(v C.SidereonPppSsrApplication) PppSsrApplication {
	return PppSsrApplication{Present: bool(v.present), HasTransmitTime: bool(v.has_transmit_time), TransmitTimeJ2000S: float64(v.transmit_time_j2000_s), HasAppliedOrbitClockSolution: bool(v.has_applied_orbit_clock_solution), AppliedOrbitClockSolution: pppSsrSolutionFromC(v.applied_orbit_clock_solution), HasObservationSignals: bool(v.has_observation_signals), Code1Signal: C.GoString(&v.code1_signal[0]), Code2Signal: C.GoString(&v.code2_signal[0]), Phase1Signal: C.GoString(&v.phase1_signal[0]), Phase2Signal: C.GoString(&v.phase2_signal[0]), CodeStatus: uint32(v.code_status), CodeStatusUT1: uint32(v.code_status_ut1), HasAppliedCodeIFM: bool(v.has_applied_code_if_m), AppliedCodeIFM: float64(v.applied_code_if_m), Code1: pppSsrSignalReportFromC(v.code1), Code2: pppSsrSignalReportFromC(v.code2), PhaseStatus: uint32(v.phase_status), PhaseStatusUT1: uint32(v.phase_status_ut1), HasAppliedPhaseIFM: bool(v.has_applied_phase_if_m), AppliedPhaseIFM: float64(v.applied_phase_if_m), Phase1: pppSsrSignalReportFromC(v.phase1), Phase2: pppSsrSignalReportFromC(v.phase2), UnknownVariant: C.GoString(&v.unknown_variant[0])}
}

func pppSSRExclusions(handle *positioningHandle, fixed bool) (result []PppSsrBiasExclusion, err error) {
	if handle == nil {
		return nil, ErrClosed
	}
	err = handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			var rows []C.SidereonPppSsrBiasExclusion
			var e error
			if fixed {
				rows, e = pppCopySlice[C.SidereonPppSsrBiasExclusion]("PPP fixed SSR-bias exclusion", func(out unsafe.Pointer, n C.size_t, w, r *C.size_t) uint32 {
					return uint32(C.sidereon_ppp_fixed_solution_ssr_bias_exclusions((*C.SidereonPppFixedSolution)(pointer), (*C.SidereonPppSsrBiasExclusion)(out), n, w, r))
				})
			} else {
				rows, e = pppCopySlice[C.SidereonPppSsrBiasExclusion]("PPP float SSR-bias exclusion", func(out unsafe.Pointer, n C.size_t, w, r *C.size_t) uint32 {
					return uint32(C.sidereon_ppp_float_solution_ssr_bias_exclusions((*C.SidereonPppFloatSolution)(pointer), (*C.SidereonPppSsrBiasExclusion)(out), n, w, r))
				})
			}
			if e != nil {
				return e
			}
			result = make([]PppSsrBiasExclusion, len(rows))
			for i, row := range rows {
				index, e := sizeTToInt(row.epoch_index, "PPP SSR-bias exclusion epoch index")
				if e != nil {
					return e
				}
				var text []byte
				text, e = copyNativeBytesLocked("PPP SSR-bias exclusion error text", func(out *C.uint8_t, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
					if fixed {
						return C.sidereon_ppp_fixed_solution_ssr_bias_exclusion_error_text((*C.SidereonPppFixedSolution)(pointer), C.size_t(i), out, n, w, r)
					}
					return C.sidereon_ppp_float_solution_ssr_bias_exclusion_error_text((*C.SidereonPppFloatSolution)(pointer), C.size_t(i), out, n, w, r)
				})
				if e != nil {
					return e
				}
				result[i] = PppSsrBiasExclusion{EpochIndex: index, SatelliteID: C.GoString(&row.satellite_id.bytes[0]), AmbiguityID: C.GoString(&row.ambiguity_id.bytes[0]), CodeBiasMissing: bool(row.code_bias_missing), PhaseBiasMissing: bool(row.phase_bias_missing), TransmitTimeFailure: pppTransmitTimeFailureFromC(row.transmit_time_failure), Application: pppSsrApplicationFromC(row.application), ErrorText: string(text)}
			}
			return nil
		})
	})
	runtime.KeepAlive(handle)
	return result, err
}

func (s *PppFloatSolution) SSRBiasExclusions() ([]PppSsrBiasExclusion, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	return pppSSRExclusions(s.handle, false)
}
func (s *PppFixedSolution) SSRBiasExclusions() ([]PppSsrBiasExclusion, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	return pppSSRExclusions(s.handle, true)
}
