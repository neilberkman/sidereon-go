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

type SiderealFilterOptions struct {
	SampleIntervalS           float64
	PriorPeriods, MinCoverage int
	TemplateMethod            uint32
	EWMAAlpha                 float64
}
type SiderealPeriodicityStrengthValue struct{ PeriodS, Strength float64 }

func cSiderealOptions(value SiderealFilterOptions) (C.SidereonSiderealFilterOptions, error) {
	if value.PriorPeriods < 0 || value.MinCoverage < 0 {
		return C.SidereonSiderealFilterOptions{}, invalidArgument("sidereal coverage counts must not be negative")
	}
	if value.TemplateMethod > 2 {
		return C.SidereonSiderealFilterOptions{}, invalidArgument("sidereal template method is not defined by the C ABI")
	}
	prior, err := checkedNativeSize(value.PriorPeriods)
	if err != nil {
		return C.SidereonSiderealFilterOptions{}, err
	}
	coverage, err := checkedNativeSize(value.MinCoverage)
	if err != nil {
		return C.SidereonSiderealFilterOptions{}, err
	}
	return C.SidereonSiderealFilterOptions{sample_interval_s: C.double(value.SampleIntervalS), prior_periods: prior, min_coverage: coverage, template_method: C.uint32_t(value.TemplateMethod), ewma_alpha: C.double(value.EWMAAlpha)}, nil
}
func siderealOptionsFromC(value C.SidereonSiderealFilterOptions) (SiderealFilterOptions, error) {
	priorPeriods, err := checkedNativeCount(uint64(value.prior_periods))
	if err != nil {
		return SiderealFilterOptions{}, err
	}
	minCoverage, err := checkedNativeCount(uint64(value.min_coverage))
	if err != nil {
		return SiderealFilterOptions{}, err
	}
	return SiderealFilterOptions{SampleIntervalS: float64(value.sample_interval_s), PriorPeriods: priorPeriods, MinCoverage: minCoverage, TemplateMethod: uint32(value.template_method), EWMAAlpha: float64(value.ewma_alpha)}, nil
}

type SiderealFilterOutput struct {
	_      noCopy
	handle *positioningHandle
}

func releaseSiderealFilterOutput(pointer unsafe.Pointer) {
	C.sidereon_sidereal_filter_output_free((*C.SidereonSiderealFilterOutput)(pointer))
}

func SiderealFilterOptionsDefaults() (SiderealFilterOptions, error) {
	var out C.SidereonSiderealFilterOptions
	err := callStatus(func() uint32 { return C.sidereon_sidereal_filter_options_init(&out) })
	if err != nil {
		return SiderealFilterOptions{}, err
	}
	return siderealOptionsFromC(out)
}

func SiderealFilter(series []float64, periodS float64, options *SiderealFilterOptions) (*SiderealFilterOutput, error) {
	if _, err := checkedNativeAllocationSize(len(series), unsafe.Sizeof(C.double(0))); err != nil {
		return nil, err
	}
	seriesCopy := append([]float64(nil), series...)
	if _, err := checkedNativeAllocationSize(len(seriesCopy), unsafe.Sizeof(C.double(0))); err != nil {
		return nil, err
	}
	var optionsC C.SidereonSiderealFilterOptions
	var optionsPtr *C.SidereonSiderealFilterOptions
	var err error
	if options != nil {
		optionsC, err = cSiderealOptions(*options)
		if err != nil {
			return nil, err
		}
		optionsPtr = &optionsC
	}
	var output *C.SidereonSiderealFilterOutput
	var input *C.double
	if len(seriesCopy) != 0 {
		input = (*C.double)(unsafe.Pointer(&seriesCopy[0]))
	}
	err = callStatus(func() uint32 {
		return C.sidereon_sidereal_filter(input, C.size_t(len(seriesCopy)), C.double(periodS), optionsPtr, &output)
	})
	runtime.KeepAlive(seriesCopy)
	if err != nil {
		if output != nil {
			withCThread(func() { C.sidereon_sidereal_filter_output_free(output) })
		}
		return nil, err
	}
	if output == nil {
		return nil, missingNativeHandle("sidereal filter")
	}
	return &SiderealFilterOutput{handle: newPositioningHandle(unsafe.Pointer(output), releaseSiderealFilterOutput)}, nil
}
func (o *SiderealFilterOutput) Close() error {
	if o == nil || o.handle == nil {
		return nil
	}
	return o.handle.close()
}

func siderealOutputValues(o *SiderealFilterOutput, label string, kind uint32) ([]float64, []int, []bool, error) {
	if o == nil || o.handle == nil {
		return nil, nil, nil, ErrClosed
	}
	var floats []float64
	var counts []int
	var flags []bool
	err := o.handle.with(func(pointer unsafe.Pointer) error {
		var written, required C.size_t
		invoke := func(out unsafe.Pointer, len C.size_t, w, r *C.size_t) uint32 {
			switch kind {
			case 0:
				return C.sidereon_sidereal_filter_output_filtered((*C.SidereonSiderealFilterOutput)(pointer), (*C.double)(out), len, w, r)
			case 1:
				return C.sidereon_sidereal_filter_output_template((*C.SidereonSiderealFilterOutput)(pointer), (*C.double)(out), len, w, r)
			case 2:
				return C.sidereon_sidereal_filter_output_coverage((*C.SidereonSiderealFilterOutput)(pointer), (*C.size_t)(out), len, w, r)
			default:
				return C.sidereon_sidereal_filter_output_under_covered((*C.SidereonSiderealFilterOutput)(pointer), (*C.bool)(out), len, w, r)
			}
		}
		if err := callStatus(func() uint32 { return invoke(nil, 0, &written, &required) }); err != nil {
			return err
		}
		n, err := validateNativeQuery(label, uint64(written), uint64(required))
		if err != nil {
			return err
		}
		if kind < 2 {
			if _, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.double(0))); err != nil {
				return err
			}
			values := make([]C.double, n)
			var out *C.double
			if n > 0 {
				out = &values[0]
			}
			written, required = 0, 0
			if err := callStatus(func() uint32 { return invoke(unsafe.Pointer(out), C.size_t(n), &written, &required) }); err != nil {
				return err
			}
			count, err := validateTwoPassCounts(label, n, n, uint64(written), uint64(required))
			if err != nil {
				return err
			}
			floats = make([]float64, count)
			for i := range floats {
				floats[i] = float64(values[i])
			}
			return nil
		}
		if kind == 2 {
			if _, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.size_t(0))); err != nil {
				return err
			}
			values := make([]C.size_t, n)
			var out *C.size_t
			if n > 0 {
				out = &values[0]
			}
			written, required = 0, 0
			if err := callStatus(func() uint32 { return invoke(unsafe.Pointer(out), C.size_t(n), &written, &required) }); err != nil {
				return err
			}
			count, err := validateTwoPassCounts(label, n, n, uint64(written), uint64(required))
			if err != nil {
				return err
			}
			counts = make([]int, count)
			for i := range counts {
				counts[i], err = checkedNativeCount(uint64(values[i]))
				if err != nil {
					return err
				}
			}
			return nil
		}
		if _, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.bool(false))); err != nil {
			return err
		}
		values := make([]C.bool, n)
		var out *C.bool
		if n > 0 {
			out = &values[0]
		}
		written, required = 0, 0
		if err := callStatus(func() uint32 { return invoke(unsafe.Pointer(out), C.size_t(n), &written, &required) }); err != nil {
			return err
		}
		count, err := validateTwoPassCounts(label, n, n, uint64(written), uint64(required))
		if err != nil {
			return err
		}
		flags = make([]bool, count)
		for i := range flags {
			flags[i] = bool(values[i])
		}
		return nil
	})
	return floats, counts, flags, err
}

func (o *SiderealFilterOutput) Filtered() ([]float64, error) {
	values, _, _, err := siderealOutputValues(o, "sidereal filtered values", 0)
	return values, err
}
func (o *SiderealFilterOutput) Template() ([]float64, error) {
	values, _, _, err := siderealOutputValues(o, "sidereal template", 1)
	return values, err
}
func (o *SiderealFilterOutput) Coverage() ([]int, error) {
	_, values, _, err := siderealOutputValues(o, "sidereal coverage", 2)
	return values, err
}
func (o *SiderealFilterOutput) UnderCovered() ([]bool, error) {
	_, _, values, err := siderealOutputValues(o, "sidereal under-covered", 3)
	return values, err
}

func SiderealPeriodicityStrength(series, candidates []float64) ([]SiderealPeriodicityStrengthValue, error) {
	if _, err := checkedNativeAllocationSize(len(series), unsafe.Sizeof(C.double(0))); err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(len(candidates), unsafe.Sizeof(C.double(0))); err != nil {
		return nil, err
	}
	seriesCopy, candidatesCopy := append([]float64(nil), series...), append([]float64(nil), candidates...)
	if _, err := checkedNativeAllocationSize(len(seriesCopy), unsafe.Sizeof(C.double(0))); err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(len(candidatesCopy), unsafe.Sizeof(C.double(0))); err != nil {
		return nil, err
	}
	var seriesPtr, candidatePtr *C.double
	if len(seriesCopy) > 0 {
		seriesPtr = (*C.double)(unsafe.Pointer(&seriesCopy[0]))
	}
	if len(candidatesCopy) > 0 {
		candidatePtr = (*C.double)(unsafe.Pointer(&candidatesCopy[0]))
	}
	var written, required C.size_t
	err := callStatus(func() uint32 {
		return C.sidereon_sidereal_periodicity_strength(seriesPtr, C.size_t(len(seriesCopy)), candidatePtr, C.size_t(len(candidatesCopy)), nil, 0, &written, &required)
	})
	if err != nil {
		return nil, err
	}
	n, err := validateNativeQuery("sidereal periodicity strength", uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	if n != len(candidatesCopy) {
		return nil, errors.New("sidereon: sidereal periodicity count differs from candidate count")
	}
	if _, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonSiderealPeriodicityStrength{})); err != nil {
		return nil, err
	}
	values := make([]C.SidereonSiderealPeriodicityStrength, n)
	var out *C.SidereonSiderealPeriodicityStrength
	if n > 0 {
		out = &values[0]
	}
	written, required = 0, 0
	err = callStatus(func() uint32 {
		return C.sidereon_sidereal_periodicity_strength(seriesPtr, C.size_t(len(seriesCopy)), candidatePtr, C.size_t(len(candidatesCopy)), out, C.size_t(n), &written, &required)
	})
	runtime.KeepAlive(seriesCopy)
	runtime.KeepAlive(candidatesCopy)
	if err != nil {
		return nil, err
	}
	count, err := validateTwoPassCounts("sidereal periodicity strength", n, n, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	result := make([]SiderealPeriodicityStrengthValue, count)
	for i := range result {
		result[i] = SiderealPeriodicityStrengthValue{PeriodS: float64(values[i].period_s), Strength: float64(values[i].strength)}
	}
	return result, nil
}

func SiderealRepeatPeriod(system uint32) (float64, error) {
	if system > 6 {
		return 0, invalidArgument("GNSS system is not defined by the C ABI")
	}
	var out C.double
	err := callStatus(func() uint32 { return C.sidereon_sidereal_repeat_period(C.uint32_t(system), &out) })
	return float64(out), err
}
