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

func AnalyticKeplerianElementsFromCartesian(state AnalyticCartesian, muKm3S2 float64) (AnalyticKeplerianElements, error) {
	var value C.SidereonAnalyticKeplerianElements
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_analytic_keplerian_elements_from_cartesian(orbitAnalyticCartesianToC(state), C.double(muKm3S2), &value))
	})
	return orbitAnalyticKeplerianElementsFromC(value), err
}

func AnalyticKeplerianElementsToCartesian(elements AnalyticKeplerianElements, muKm3S2 float64) (AnalyticCartesian, error) {
	var value C.SidereonAnalyticCartesian
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_analytic_keplerian_elements_to_cartesian(orbitAnalyticKeplerianElementsToC(elements), C.double(muKm3S2), &value))
	})
	return orbitAnalyticCartesianFromC(value), err
}

func AnalyticTheoryText(theory AnalyticTheory) (string, error) {
	return orbitText("analytic theory text", func(out *C.uint8_t, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_analytic_theory_text(C.uint32_t(theory), out, capacity, written, required)
	})
}

func ComputeBrouwerLyddaneMeanElements(state AnalyticCartesian, gravity ZonalHarmonics, options MeanConversionOptions) (BrouwerLyddaneMeanElements, error) {
	var value C.SidereonBrouwerLyddaneMeanElements
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_brouwer_lyddane_mean_elements(orbitAnalyticCartesianToC(state), orbitZonalHarmonicsToC(gravity), orbitMeanConversionOptionsToC(options), &value))
	})
	return orbitBrouwerLyddaneMeanElementsFromC(value), err
}

func BrouwerLyddaneOsculatingElements(mean AnalyticKeplerianElements, gravity ZonalHarmonics) (AnalyticKeplerianElements, error) {
	var value C.SidereonAnalyticKeplerianElements
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_brouwer_lyddane_osculating_elements(orbitAnalyticKeplerianElementsToC(mean), orbitZonalHarmonicsToC(gravity), &value))
	})
	return orbitAnalyticKeplerianElementsFromC(value), err
}

func BrouwerLyddanePropagatorEphemeris(propagator *BrouwerLyddanePropagator, epochsTDBSeconds []float64) ([]CartesianState, error) {
	return withOrbitHandles([]*positioningHandle{propagator.orbitHandle()}, func(pointers []unsafe.Pointer) ([]CartesianState, error) {
		raw, err := orbitReadRows("brouwer lyddane propagator ephemeris", func(out *C.SidereonCartesianState, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_brouwer_lyddane_propagator_ephemeris((*C.SidereonBrouwerLyddanePropagator)(pointers[0]), (*C.double)(unsafe.Pointer(unsafe.SliceData(epochsTDBSeconds))), C.size_t(len(epochsTDBSeconds)), out, capacity, written, required)
		})
		if err != nil {
			return nil, err
		}
		return orbitMap(raw, func(item C.SidereonCartesianState) CartesianState { return cartesianStateFromC(item) }), nil
	})
}

func BrouwerLyddanePropagatorEpochTDBSeconds(propagator *BrouwerLyddanePropagator) (float64, error) {
	return withOrbitHandles([]*positioningHandle{propagator.orbitHandle()}, func(pointers []unsafe.Pointer) (float64, error) {
		var value C.double
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_brouwer_lyddane_propagator_epoch_tdb_seconds((*C.SidereonBrouwerLyddanePropagator)(pointers[0]), &value))
		})
		return float64(value), err
	})
}

func BrouwerLyddanePropagatorFromMeanElements(epochTDBSeconds float64, mean AnalyticKeplerianElements, gravity ZonalHarmonics, m2RadS2 float64) (*BrouwerLyddanePropagator, error) {
	var value *C.SidereonBrouwerLyddanePropagator
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_brouwer_lyddane_propagator_from_mean_elements(C.double(epochTDBSeconds), orbitAnalyticKeplerianElementsToC(mean), orbitZonalHarmonicsToC(gravity), C.double(m2RadS2), &value))
	})
	return newOrbitBrouwerLyddanePropagator(value, err)
}

func BrouwerLyddanePropagatorGravity(propagator *BrouwerLyddanePropagator) (ZonalHarmonics, error) {
	return withOrbitHandles([]*positioningHandle{propagator.orbitHandle()}, func(pointers []unsafe.Pointer) (ZonalHarmonics, error) {
		var value C.SidereonZonalHarmonics
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_brouwer_lyddane_propagator_gravity((*C.SidereonBrouwerLyddanePropagator)(pointers[0]), &value))
		})
		return orbitZonalHarmonicsFromC(value), err
	})
}

func BrouwerLyddanePropagatorM2RadS2(propagator *BrouwerLyddanePropagator) (float64, error) {
	return withOrbitHandles([]*positioningHandle{propagator.orbitHandle()}, func(pointers []unsafe.Pointer) (float64, error) {
		var value C.double
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_brouwer_lyddane_propagator_m2_rad_s2((*C.SidereonBrouwerLyddanePropagator)(pointers[0]), &value))
		})
		return float64(value), err
	})
}

func BrouwerLyddanePropagatorMeanElements(propagator *BrouwerLyddanePropagator) (AnalyticKeplerianElements, error) {
	return withOrbitHandles([]*positioningHandle{propagator.orbitHandle()}, func(pointers []unsafe.Pointer) (AnalyticKeplerianElements, error) {
		var value C.SidereonAnalyticKeplerianElements
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_brouwer_lyddane_propagator_mean_elements((*C.SidereonBrouwerLyddanePropagator)(pointers[0]), &value))
		})
		return orbitAnalyticKeplerianElementsFromC(value), err
	})
}

func BrouwerLyddanePropagatorNew(initial CartesianState, gravity ZonalHarmonics, m2RadS2 float64, options MeanConversionOptions) (*BrouwerLyddanePropagator, error) {
	var value *C.SidereonBrouwerLyddanePropagator
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_brouwer_lyddane_propagator_new(cCartesianState(initial), orbitZonalHarmonicsToC(gravity), C.double(m2RadS2), orbitMeanConversionOptionsToC(options), &value))
	})
	return newOrbitBrouwerLyddanePropagator(value, err)
}

func BrouwerLyddanePropagatorOsculatingElements(propagator *BrouwerLyddanePropagator, epochTDBSeconds float64) (AnalyticKeplerianElements, error) {
	return withOrbitHandles([]*positioningHandle{propagator.orbitHandle()}, func(pointers []unsafe.Pointer) (AnalyticKeplerianElements, error) {
		var value C.SidereonAnalyticKeplerianElements
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_brouwer_lyddane_propagator_osculating_elements((*C.SidereonBrouwerLyddanePropagator)(pointers[0]), C.double(epochTDBSeconds), &value))
		})
		return orbitAnalyticKeplerianElementsFromC(value), err
	})
}

func BrouwerLyddanePropagatorPropagate(propagator *BrouwerLyddanePropagator, epochTDBSeconds float64) (CartesianState, error) {
	return withOrbitHandles([]*positioningHandle{propagator.orbitHandle()}, func(pointers []unsafe.Pointer) (CartesianState, error) {
		var value C.SidereonCartesianState
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_brouwer_lyddane_propagator_propagate((*C.SidereonBrouwerLyddanePropagator)(pointers[0]), C.double(epochTDBSeconds), &value))
		})
		return cartesianStateFromC(value), err
	})
}

func CircularElementsEccentricity(elements CircularElements) float64 {
	value := C.sidereon_circular_elements_eccentricity(orbitCircularElementsToC(elements))
	return float64(value)
}

func CircularElementsFromCartesian(state AnalyticCartesian, muKm3S2 float64) (CircularElements, error) {
	var value C.SidereonCircularElements
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_circular_elements_from_cartesian(orbitAnalyticCartesianToC(state), C.double(muKm3S2), &value))
	})
	return orbitCircularElementsFromC(value), err
}

func CircularElementsToCartesian(elements CircularElements, muKm3S2 float64) (AnalyticCartesian, error) {
	var value C.SidereonAnalyticCartesian
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_circular_elements_to_cartesian(orbitCircularElementsToC(elements), C.double(muKm3S2), &value))
	})
	return orbitAnalyticCartesianFromC(value), err
}

func ComputeEcksteinHechlerMeanElements(state AnalyticCartesian, gravity ZonalHarmonics, options MeanConversionOptions) (EcksteinHechlerMeanElements, error) {
	var value C.SidereonEcksteinHechlerMeanElements
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_eckstein_hechler_mean_elements(orbitAnalyticCartesianToC(state), orbitZonalHarmonicsToC(gravity), orbitMeanConversionOptionsToC(options), &value))
	})
	return orbitEcksteinHechlerMeanElementsFromC(value), err
}

func EcksteinHechlerOsculatingElements(mean CircularElements, gravity ZonalHarmonics) (CircularElements, error) {
	var value C.SidereonCircularElements
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_eckstein_hechler_osculating_elements(orbitCircularElementsToC(mean), orbitZonalHarmonicsToC(gravity), &value))
	})
	return orbitCircularElementsFromC(value), err
}

func EcksteinHechlerPropagatorEphemeris(propagator *EcksteinHechlerPropagator, epochsTDBSeconds []float64) ([]CartesianState, error) {
	return withOrbitHandles([]*positioningHandle{propagator.orbitHandle()}, func(pointers []unsafe.Pointer) ([]CartesianState, error) {
		raw, err := orbitReadRows("eckstein hechler propagator ephemeris", func(out *C.SidereonCartesianState, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_eckstein_hechler_propagator_ephemeris((*C.SidereonEcksteinHechlerPropagator)(pointers[0]), (*C.double)(unsafe.Pointer(unsafe.SliceData(epochsTDBSeconds))), C.size_t(len(epochsTDBSeconds)), out, capacity, written, required)
		})
		if err != nil {
			return nil, err
		}
		return orbitMap(raw, func(item C.SidereonCartesianState) CartesianState { return cartesianStateFromC(item) }), nil
	})
}

func EcksteinHechlerPropagatorEpochTDBSeconds(propagator *EcksteinHechlerPropagator) (float64, error) {
	return withOrbitHandles([]*positioningHandle{propagator.orbitHandle()}, func(pointers []unsafe.Pointer) (float64, error) {
		var value C.double
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_eckstein_hechler_propagator_epoch_tdb_seconds((*C.SidereonEcksteinHechlerPropagator)(pointers[0]), &value))
		})
		return float64(value), err
	})
}

func EcksteinHechlerPropagatorFromMeanElements(epochTDBSeconds float64, mean CircularElements, gravity ZonalHarmonics) (*EcksteinHechlerPropagator, error) {
	var value *C.SidereonEcksteinHechlerPropagator
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_eckstein_hechler_propagator_from_mean_elements(C.double(epochTDBSeconds), orbitCircularElementsToC(mean), orbitZonalHarmonicsToC(gravity), &value))
	})
	return newOrbitEcksteinHechlerPropagator(value, err)
}

func EcksteinHechlerPropagatorGravity(propagator *EcksteinHechlerPropagator) (ZonalHarmonics, error) {
	return withOrbitHandles([]*positioningHandle{propagator.orbitHandle()}, func(pointers []unsafe.Pointer) (ZonalHarmonics, error) {
		var value C.SidereonZonalHarmonics
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_eckstein_hechler_propagator_gravity((*C.SidereonEcksteinHechlerPropagator)(pointers[0]), &value))
		})
		return orbitZonalHarmonicsFromC(value), err
	})
}

func EcksteinHechlerPropagatorMeanElements(propagator *EcksteinHechlerPropagator) (CircularElements, error) {
	return withOrbitHandles([]*positioningHandle{propagator.orbitHandle()}, func(pointers []unsafe.Pointer) (CircularElements, error) {
		var value C.SidereonCircularElements
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_eckstein_hechler_propagator_mean_elements((*C.SidereonEcksteinHechlerPropagator)(pointers[0]), &value))
		})
		return orbitCircularElementsFromC(value), err
	})
}

func EcksteinHechlerPropagatorNew(initial CartesianState, gravity ZonalHarmonics, options MeanConversionOptions) (*EcksteinHechlerPropagator, error) {
	var value *C.SidereonEcksteinHechlerPropagator
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_eckstein_hechler_propagator_new(cCartesianState(initial), orbitZonalHarmonicsToC(gravity), orbitMeanConversionOptionsToC(options), &value))
	})
	return newOrbitEcksteinHechlerPropagator(value, err)
}

func EcksteinHechlerPropagatorOsculatingElements(propagator *EcksteinHechlerPropagator, epochTDBSeconds float64) (CircularElements, error) {
	return withOrbitHandles([]*positioningHandle{propagator.orbitHandle()}, func(pointers []unsafe.Pointer) (CircularElements, error) {
		var value C.SidereonCircularElements
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_eckstein_hechler_propagator_osculating_elements((*C.SidereonEcksteinHechlerPropagator)(pointers[0]), C.double(epochTDBSeconds), &value))
		})
		return orbitCircularElementsFromC(value), err
	})
}

func EcksteinHechlerPropagatorPropagate(propagator *EcksteinHechlerPropagator, epochTDBSeconds float64) (CartesianState, error) {
	return withOrbitHandles([]*positioningHandle{propagator.orbitHandle()}, func(pointers []unsafe.Pointer) (CartesianState, error) {
		var value C.SidereonCartesianState
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_eckstein_hechler_propagator_propagate((*C.SidereonEcksteinHechlerPropagator)(pointers[0]), C.double(epochTDBSeconds), &value))
		})
		return cartesianStateFromC(value), err
	})
}

func MeanConversionOptionsBrouwerLyddaneDefault() MeanConversionOptions {
	value := C.sidereon_mean_conversion_options_brouwer_lyddane_default()
	return orbitMeanConversionOptionsFromC(value)
}

func MeanConversionOptionsEcksteinHechlerDefault() MeanConversionOptions {
	value := C.sidereon_mean_conversion_options_eckstein_hechler_default()
	return orbitMeanConversionOptionsFromC(value)
}

func ZonalHarmonicsDefault() ZonalHarmonics {
	value := C.sidereon_zonal_harmonics_default()
	return orbitZonalHarmonicsFromC(value)
}

func ZonalHarmonicsEarth() ZonalHarmonics {
	value := C.sidereon_zonal_harmonics_earth()
	return orbitZonalHarmonicsFromC(value)
}
