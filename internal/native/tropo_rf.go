//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
*/
import "C"

type Met struct {
	PressureHPa      float64
	TemperatureK     float64
	RelativeHumidity float64
}

type MappingFactors struct {
	Dry float64
	Wet float64
}

type TropoMappingError struct {
	Kind            uint32
	ElevationRad    float64
	MinElevationRad float64
}

type ZenithDelay struct {
	DryM float64
	WetM float64
}

func DefaultMet() (Met, error) {
	var output C.SidereonMet
	err := callStatus(func() uint32 { return C.sidereon_met_init(&output) })
	return Met{PressureHPa: float64(output.pressure_hpa), TemperatureK: float64(output.temperature_k), RelativeHumidity: float64(output.relative_humidity)}, err
}

func cMet(value Met) C.SidereonMet {
	return C.SidereonMet{pressure_hpa: C.double(value.PressureHPa), temperature_k: C.double(value.TemperatureK), relative_humidity: C.double(value.RelativeHumidity)}
}

func TropoMappingFactors(elevationRad float64, receiver Geodetic, scale uint32, jdWhole, jdFraction float64) (MappingFactors, error) {
	var output C.SidereonMappingFactors
	err := callStatus(func() uint32 {
		return C.sidereon_tropo_mapping_factors(C.double(elevationRad), cGeodetic(receiver), C.uint32_t(scale), C.double(jdWhole), C.double(jdFraction), &output)
	})
	return MappingFactors{Dry: float64(output.dry), Wet: float64(output.wet)}, err
}

func TropoMappingFactorsChecked(elevationRad float64, receiver Geodetic, scale uint32, jdWhole, jdFraction float64) (MappingFactors, TropoMappingError, error) {
	var output C.SidereonMappingFactors
	var detail C.SidereonTropoMappingError
	err := callStatus(func() uint32 {
		return C.sidereon_tropo_mapping_factors_checked(C.double(elevationRad), cGeodetic(receiver), C.uint32_t(scale), C.double(jdWhole), C.double(jdFraction), &output, &detail)
	})
	return MappingFactors{Dry: float64(output.dry), Wet: float64(output.wet)}, TropoMappingError{Kind: uint32(detail.kind), ElevationRad: float64(detail.elevation_rad), MinElevationRad: float64(detail.min_elevation_rad)}, err
}

func TropoZenithDelay(receiver Geodetic, met Met) (ZenithDelay, error) {
	cMetValue := cMet(met)
	var output C.SidereonZenithDelay
	err := callStatus(func() uint32 { return C.sidereon_tropo_zenith_delay(cGeodetic(receiver), &cMetValue, &output) })
	return ZenithDelay{DryM: float64(output.dry_m), WetM: float64(output.wet_m)}, err
}

func TropoSlantDelay(elevationRad float64, receiver Geodetic, met Met, scale uint32, jdWhole, jdFraction float64) (float64, error) {
	cMetValue := cMet(met)
	var output C.double
	err := callStatus(func() uint32 {
		return C.sidereon_tropo_slant_delay(C.double(elevationRad), cGeodetic(receiver), &cMetValue, C.uint32_t(scale), C.double(jdWhole), C.double(jdFraction), &output)
	})
	return float64(output), err
}

type LinkBudget struct {
	EIRPdBW         float64
	FSPLdB          float64
	ReceiverGTdBK   float64
	OtherLossesdB   float64
	RequiredCN0dBHz float64
}

func scalarRF(fn func(*C.double) uint32) (float64, error) {
	var output C.double
	err := callStatus(func() uint32 { return fn(&output) })
	return float64(output), err
}

func RFFSPL(distanceKm, frequencyMHz float64) (float64, error) {
	return scalarRF(func(out *C.double) uint32 {
		return C.sidereon_rf_fspl(C.double(distanceKm), C.double(frequencyMHz), out)
	})
}

func RFEIRP(txPowerDBM, txAntennaGainDBI float64) (float64, error) {
	return scalarRF(func(out *C.double) uint32 {
		return C.sidereon_rf_eirp(C.double(txPowerDBM), C.double(txAntennaGainDBI), out)
	})
}

func RFCN0(eirpDBW, fsplDB, receiverGTDBK, otherLossesDB float64) (float64, error) {
	return scalarRF(func(out *C.double) uint32 {
		return C.sidereon_rf_cn0(C.double(eirpDBW), C.double(fsplDB), C.double(receiverGTDBK), C.double(otherLossesDB), out)
	})
}

func RFLinkMargin(budget LinkBudget) (float64, error) {
	input := C.SidereonLinkBudget{eirp_dbw: C.double(budget.EIRPdBW), fspl_db: C.double(budget.FSPLdB), receiver_gt_dbk: C.double(budget.ReceiverGTdBK), other_losses_db: C.double(budget.OtherLossesdB), required_cn0_dbhz: C.double(budget.RequiredCN0dBHz)}
	return scalarRF(func(out *C.double) uint32 { return C.sidereon_rf_link_margin(&input, out) })
}

func RFWavelength(frequencyHz float64) (float64, error) {
	return scalarRF(func(out *C.double) uint32 { return C.sidereon_rf_wavelength(C.double(frequencyHz), out) })
}

func RFDishGain(diameterM, frequencyHz, efficiency float64) (float64, error) {
	return scalarRF(func(out *C.double) uint32 {
		return C.sidereon_rf_dish_gain(C.double(diameterM), C.double(frequencyHz), C.double(efficiency), out)
	})
}
