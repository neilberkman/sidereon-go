//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"runtime"
	"unsafe"
)

type PppCorrectionLookup struct {
	_      noCopy
	handle *positioningHandle
}

type PppCorrectionLookupInfo struct {
	DegradedReason          uint32
	TideEnabled             bool
	PoleTideEnabled         bool
	OceanLoadingEnabled     bool
	WindupEnabled           bool
	SatelliteAntennaEnabled bool
	CodeBiasEnabled         bool
}

func releasePppCorrectionLookup(pointer unsafe.Pointer) {
	withCThread(func() { C.sidereon_ppp_correction_lookup_free((*C.SidereonPppCorrectionLookup)(pointer)) })
}

func newPppCorrectionLookup(pointer *C.SidereonPppCorrectionLookup) (*PppCorrectionLookup, error) {
	if pointer == nil {
		return nil, missingNativeHandle("PPP correction lookup")
	}
	return &PppCorrectionLookup{handle: newPositioningHandle(unsafe.Pointer(pointer), releasePppCorrectionLookup)}, nil
}

func (lookup *PppCorrectionLookup) Close() error {
	if lookup == nil || lookup.handle == nil {
		return nil
	}
	return lookup.handle.close()
}

func PppCorrectionLookupDefault() (*PppCorrectionLookup, error) {
	var output *C.SidereonPppCorrectionLookup
	err := callStatus(func() uint32 { return uint32(C.sidereon_ppp_correction_lookup_default(&output)) })
	if err != nil {
		return nil, err
	}
	return newPppCorrectionLookup(output)
}

func PppCorrectionLookupFromCorrections(corrections *PppCorrections) (*PppCorrectionLookup, error) {
	if corrections == nil || corrections.handle == nil {
		return nil, ErrClosed
	}
	var output *C.SidereonPppCorrectionLookup
	err := corrections.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			return statusErrorLocked(uint32(C.sidereon_ppp_correction_lookup_from_corrections((*C.SidereonPppCorrections)(pointer), &output)))
		})
	})
	if err != nil {
		return nil, err
	}
	return newPppCorrectionLookup(output)
}

func PppBuildLookup(sp3 *SP3, epochs []PppEpoch, receiver [3]float64, options PppCorrectionsOptions, validity *uint32) (*PppCorrectionLookup, error) {
	if sp3 == nil || sp3.handle == nil {
		return nil, ErrClosed
	}
	if validity != nil && *validity > PppValidityPermissive {
		return nil, invalidArgument("invalid PPP validity selector")
	}
	if options.CodeBias != nil && options.CodeBias.handle == nil {
		return nil, ErrClosed
	}
	var output *C.SidereonPppCorrectionLookup
	build := func(sp3Pointer, biasPointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			arena := &pppArena{}
			defer arena.close()
			cEpochs, count, err := cPppEpochs(epochs, arena)
			if err != nil {
				return err
			}
			var bias *C.SidereonBiasSet
			if biasPointer != nil {
				bias = (*C.SidereonBiasSet)(biasPointer)
			}
			cOptions, err := cPppCorrectionsOptions(options, arena, bias)
			if err != nil {
				return err
			}
			position := [3]C.double{C.double(receiver[0]), C.double(receiver[1]), C.double(receiver[2])}
			var status C.enum_SidereonStatus
			if validity == nil {
				status = C.sidereon_build_ppp_lookup((*C.SidereonSp3)(sp3Pointer), cEpochs, count, &position[0], &cOptions, &output)
			} else {
				status = C.sidereon_build_ppp_lookup_with_validity((*C.SidereonSp3)(sp3Pointer), cEpochs, count, &position[0], &cOptions, C.uint32_t(*validity), &output)
			}
			return statusErrorLocked(uint32(status))
		})
	}
	var err error
	if options.CodeBias != nil {
		err = withPositioningHandlePair(sp3.handle, options.CodeBias.handle, build)
	} else {
		err = sp3.handle.with(func(pointer unsafe.Pointer) error { return build(pointer, nil) })
	}
	runtime.KeepAlive(sp3)
	runtime.KeepAlive(options.CodeBias)
	if err != nil {
		return nil, err
	}
	return newPppCorrectionLookup(output)
}

func (lookup *PppCorrectionLookup) Equal(other *PppCorrectionLookup) (equal bool, err error) {
	if lookup == nil || lookup.handle == nil || other == nil || other.handle == nil {
		return false, ErrClosed
	}
	err = withPositioningHandlePair(lookup.handle, other.handle, func(left, right unsafe.Pointer) error {
		return withCThreadError(func() error {
			var output C.bool
			if statusErr := statusErrorLocked(uint32(C.sidereon_ppp_correction_lookup_equal((*C.SidereonPppCorrectionLookup)(left), (*C.SidereonPppCorrectionLookup)(right), &output))); statusErr != nil {
				return statusErr
			}
			equal = bool(output)
			return nil
		})
	})
	return equal, err
}

func (lookup *PppCorrectionLookup) Info() (result PppCorrectionLookupInfo, err error) {
	if lookup == nil || lookup.handle == nil {
		return result, ErrClosed
	}
	err = lookup.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			var degraded C.enum_SidereonPppCorrectionsDegradeReason
			var flags C.SidereonPppCorrectionLookupFlags
			if statusErr := statusErrorLocked(uint32(C.sidereon_ppp_correction_lookup_info((*C.SidereonPppCorrectionLookup)(pointer), &degraded, &flags))); statusErr != nil {
				return statusErr
			}
			result = PppCorrectionLookupInfo{DegradedReason: uint32(degraded), TideEnabled: bool(flags.tide_enabled), PoleTideEnabled: bool(flags.pole_tide_enabled), OceanLoadingEnabled: bool(flags.ocean_loading_enabled), WindupEnabled: bool(flags.windup_enabled), SatelliteAntennaEnabled: bool(flags.satellite_antenna_enabled), CodeBiasEnabled: bool(flags.code_bias_enabled)}
			return nil
		})
	})
	return result, err
}

func pppLookupScalars(lookup *PppCorrectionLookup, kind uint32) (result []PppSatScalarCorrection, err error) {
	if lookup == nil || lookup.handle == nil {
		return nil, ErrClosed
	}
	err = lookup.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			var written, required C.size_t
			call := func(output *C.SidereonSatScalarCorrection, count C.size_t) C.enum_SidereonStatus {
				switch kind {
				case 0:
					return C.sidereon_ppp_correction_lookup_code_bias((*C.SidereonPppCorrectionLookup)(pointer), output, count, &written, &required)
				case 1:
					return C.sidereon_ppp_correction_lookup_sat_pcv((*C.SidereonPppCorrectionLookup)(pointer), output, count, &written, &required)
				default:
					return C.sidereon_ppp_correction_lookup_windup((*C.SidereonPppCorrectionLookup)(pointer), output, count, &written, &required)
				}
			}
			if statusErr := statusErrorLocked(uint32(call(nil, 0))); statusErr != nil {
				return statusErr
			}
			count, countErr := validateNativeQuery("PPP lookup scalar corrections", uint64(written), uint64(required))
			if countErr != nil {
				return countErr
			}
			values := make([]C.SidereonSatScalarCorrection, count)
			var output *C.SidereonSatScalarCorrection
			if count != 0 {
				output = &values[0]
			}
			written, required = 0, 0
			if statusErr := statusErrorLocked(uint32(call(output, C.size_t(count)))); statusErr != nil {
				return statusErr
			}
			actual, countErr := validateTwoPassCounts("PPP lookup scalar corrections", count, count, uint64(written), uint64(required))
			if countErr != nil {
				return countErr
			}
			result = make([]PppSatScalarCorrection, actual)
			for index := range result {
				epoch, conversionErr := checkedNativeCount(uint64(values[index].epoch_index))
				if conversionErr != nil {
					return conversionErr
				}
				result[index] = PppSatScalarCorrection{SatelliteID: C.GoString(&values[index].sat_id[0]), EpochIndex: epoch, Value: float64(values[index].value_m)}
			}
			return nil
		})
	})
	return result, err
}

func (lookup *PppCorrectionLookup) CodeBias() ([]PppSatScalarCorrection, error) {
	return pppLookupScalars(lookup, 0)
}
func (lookup *PppCorrectionLookup) SatellitePCV() ([]PppSatScalarCorrection, error) {
	return pppLookupScalars(lookup, 1)
}
func (lookup *PppCorrectionLookup) Windup() ([]PppSatScalarCorrection, error) {
	return pppLookupScalars(lookup, 2)
}

func pppLookupVectors(lookup *PppCorrectionLookup, kind uint32) (result []PppEpochVectorCorrection, err error) {
	if lookup == nil || lookup.handle == nil {
		return nil, ErrClosed
	}
	err = lookup.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			var written, required C.size_t
			call := func(output *C.SidereonEpochVectorCorrection, count C.size_t) C.enum_SidereonStatus {
				switch kind {
				case 0:
					return C.sidereon_ppp_correction_lookup_tide((*C.SidereonPppCorrectionLookup)(pointer), output, count, &written, &required)
				case 1:
					return C.sidereon_ppp_correction_lookup_pole_tide((*C.SidereonPppCorrectionLookup)(pointer), output, count, &written, &required)
				default:
					return C.sidereon_ppp_correction_lookup_ocean_loading((*C.SidereonPppCorrectionLookup)(pointer), output, count, &written, &required)
				}
			}
			if statusErr := statusErrorLocked(uint32(call(nil, 0))); statusErr != nil {
				return statusErr
			}
			count, countErr := validateNativeQuery("PPP lookup vector corrections", uint64(written), uint64(required))
			if countErr != nil {
				return countErr
			}
			values := make([]C.SidereonEpochVectorCorrection, count)
			var output *C.SidereonEpochVectorCorrection
			if count != 0 {
				output = &values[0]
			}
			written, required = 0, 0
			if statusErr := statusErrorLocked(uint32(call(output, C.size_t(count)))); statusErr != nil {
				return statusErr
			}
			actual, countErr := validateTwoPassCounts("PPP lookup vector corrections", count, count, uint64(written), uint64(required))
			if countErr != nil {
				return countErr
			}
			result = make([]PppEpochVectorCorrection, actual)
			for index := range result {
				epoch, conversionErr := checkedNativeCount(uint64(values[index].epoch_index))
				if conversionErr != nil {
					return conversionErr
				}
				result[index] = PppEpochVectorCorrection{EpochIndex: epoch, ValueM: [3]float64{float64(values[index].vector_m[0]), float64(values[index].vector_m[1]), float64(values[index].vector_m[2])}}
			}
			return nil
		})
	})
	return result, err
}

func (lookup *PppCorrectionLookup) Tide() ([]PppEpochVectorCorrection, error) {
	return pppLookupVectors(lookup, 0)
}
func (lookup *PppCorrectionLookup) PoleTide() ([]PppEpochVectorCorrection, error) {
	return pppLookupVectors(lookup, 1)
}
func (lookup *PppCorrectionLookup) OceanLoading() ([]PppEpochVectorCorrection, error) {
	return pppLookupVectors(lookup, 2)
}

func (lookup *PppCorrectionLookup) SatellitePCOECEF() (result []PppSatVectorCorrection, err error) {
	if lookup == nil || lookup.handle == nil {
		return nil, ErrClosed
	}
	err = lookup.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			var written, required C.size_t
			call := func(output *C.SidereonSatVectorCorrection, count C.size_t) C.enum_SidereonStatus {
				return C.sidereon_ppp_correction_lookup_sat_pco_ecef((*C.SidereonPppCorrectionLookup)(pointer), output, count, &written, &required)
			}
			if statusErr := statusErrorLocked(uint32(call(nil, 0))); statusErr != nil {
				return statusErr
			}
			count, countErr := validateNativeQuery("PPP lookup satellite PCO", uint64(written), uint64(required))
			if countErr != nil {
				return countErr
			}
			values := make([]C.SidereonSatVectorCorrection, count)
			var output *C.SidereonSatVectorCorrection
			if count != 0 {
				output = &values[0]
			}
			written, required = 0, 0
			if statusErr := statusErrorLocked(uint32(call(output, C.size_t(count)))); statusErr != nil {
				return statusErr
			}
			actual, countErr := validateTwoPassCounts("PPP lookup satellite PCO", count, count, uint64(written), uint64(required))
			if countErr != nil {
				return countErr
			}
			result = make([]PppSatVectorCorrection, actual)
			for index := range result {
				epoch, conversionErr := checkedNativeCount(uint64(values[index].epoch_index))
				if conversionErr != nil {
					return conversionErr
				}
				result[index] = PppSatVectorCorrection{SatelliteID: C.GoString(&values[index].sat_id[0]), EpochIndex: epoch, ValueM: [3]float64{float64(values[index].vector_m[0]), float64(values[index].vector_m[1]), float64(values[index].vector_m[2])}}
			}
			return nil
		})
	})
	return result, err
}
