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

func BeiDou2OnToYsBetaRad() float64 {
	value := C.sidereon_beidou2_on_to_ys_beta_rad()
	return float64(value)
}

func BeiDou2YsToOnBetaRad() float64 {
	value := C.sidereon_beidou2_ys_to_on_beta_rad()
	return float64(value)
}

func BeiDou3CASTIGSOTurn() CosineTurnParameters {
	value := C.sidereon_beidou3_cast_igso_turn()
	return orbitCosineTurnParametersFromC(value)
}

func BeiDou3CASTMEOTurn() CosineTurnParameters {
	value := C.sidereon_beidou3_cast_meo_turn()
	return orbitCosineTurnParametersFromC(value)
}

func BeiDou3SECMBeta0Rad() float64 {
	value := C.sidereon_beidou3_secm_beta0_rad()
	return float64(value)
}

func BeiDouManufacturerCompare(left BeiDouManufacturer, right BeiDouManufacturer) (int32, error) {
	var value C.int32_t
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_beidou_manufacturer_compare(C.uint32_t(left), C.uint32_t(right), &value))
	})
	return int32(value), err
}

func BeiDouOrbitClassCompare(left BeiDouOrbitClass, right BeiDouOrbitClass) (int32, error) {
	var value C.int32_t
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_beidou_orbit_class_compare(C.uint32_t(left), C.uint32_t(right), &value))
	})
	return int32(value), err
}

func ComputeDYBAxes(positionM [3]float64, sunPositionM [3]float64) (DYBAxes, error) {
	var value C.SidereonDybAxes
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_dyb_axes((*C.double)(unsafe.Pointer(&positionM)), (*C.double)(unsafe.Pointer(&sunPositionM)), &value))
	})
	return orbitDYBAxesFromC(value), err
}

func GalileoFOCTurn() CosineTurnParameters {
	value := C.sidereon_galileo_foc_turn()
	return orbitCosineTurnParametersFromC(value)
}

func GalileoIOVBetaXRad() float64 {
	value := C.sidereon_galileo_iov_beta_x_rad()
	return float64(value)
}

func GalileoIOVBetaYRad() float64 {
	value := C.sidereon_galileo_iov_beta_y_rad()
	return float64(value)
}

func GLONASSMDefaultYawRateRadS() float64 {
	value := C.sidereon_glonass_m_default_yaw_rate_rad_s()
	return float64(value)
}

func GLONASSMShadowLimitRad() float64 {
	value := C.sidereon_glonass_m_shadow_limit_rad()
	return float64(value)
}

func GNSSAttitudeLawBeiDou3MEOUnknownManufacturer(cast CosineTurnParameters, secmBeta0Rad float64) (*GNSSAttitudeLaw, error) {
	var value *C.SidereonGnssAttitudeLaw
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_gnss_attitude_law_beidou3_meo_unknown_manufacturer(orbitCosineTurnParametersToC(cast), C.double(secmBeta0Rad), &value))
	})
	return newOrbitGNSSAttitudeLaw(value, err)
}

func GNSSAttitudeLawBeiDouOrbitNormalSwitch(ysToOnBetaRad float64, onToYsBetaRad float64) (*GNSSAttitudeLaw, error) {
	var value *C.SidereonGnssAttitudeLaw
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_gnss_attitude_law_beidou_orbit_normal_switch(C.double(ysToOnBetaRad), C.double(onToYsBetaRad), &value))
	})
	return newOrbitGNSSAttitudeLaw(value, err)
}

func GNSSAttitudeLawCosineTurn(parameters CosineTurnParameters) (*GNSSAttitudeLaw, error) {
	var value *C.SidereonGnssAttitudeLaw
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_gnss_attitude_law_cosine_turn(orbitCosineTurnParametersToC(parameters), &value))
	})
	return newOrbitGNSSAttitudeLaw(value, err)
}

func GNSSAttitudeLawEqual(left *GNSSAttitudeLaw, right *GNSSAttitudeLaw) (bool, error) {
	return withOrbitHandles([]*positioningHandle{left.orbitHandle(), right.orbitHandle()}, func(pointers []unsafe.Pointer) (bool, error) {
		var value C.bool
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_gnss_attitude_law_equal((*C.SidereonGnssAttitudeLaw)(pointers[0]), (*C.SidereonGnssAttitudeLaw)(pointers[1]), &value))
		})
		return bool(value), err
	})
}

func GNSSAttitudeLawFixedBetaYawSteering(beta0Rad float64) (*GNSSAttitudeLaw, error) {
	var value *C.SidereonGnssAttitudeLaw
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_gnss_attitude_law_fixed_beta_yaw_steering(C.double(beta0Rad), &value))
	})
	return newOrbitGNSSAttitudeLaw(value, err)
}

func GNSSAttitudeLawForBlock(block GNSSBlock, prn uint8) (*GNSSAttitudeLaw, error) {
	var value *C.SidereonGnssAttitudeLaw
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_gnss_attitude_law_for_block(C.uint32_t(block), C.uint8_t(prn), &value))
	})
	return newOrbitGNSSAttitudeLaw(value, err)
}

func GNSSAttitudeLawGalileoIOV(betaXRad float64, betaYRad float64) (*GNSSAttitudeLaw, error) {
	var value *C.SidereonGnssAttitudeLaw
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_gnss_attitude_law_galileo_iov(C.double(betaXRad), C.double(betaYRad), &value))
	})
	return newOrbitGNSSAttitudeLaw(value, err)
}

func GNSSAttitudeLawGLONASSM(yawRateRadS float64, shadowLimitRad float64) (*GNSSAttitudeLaw, error) {
	var value *C.SidereonGnssAttitudeLaw
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_gnss_attitude_law_glonass_m(C.double(yawRateRadS), C.double(shadowLimitRad), &value))
	})
	return newOrbitGNSSAttitudeLaw(value, err)
}

func GNSSAttitudeLawGPSIIA(yawRateRadS float64, yawBiasRad float64, shadowLimitRad float64) (*GNSSAttitudeLaw, error) {
	var value *C.SidereonGnssAttitudeLaw
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_gnss_attitude_law_gps_iia(C.double(yawRateRadS), C.double(yawBiasRad), C.double(shadowLimitRad), &value))
	})
	return newOrbitGNSSAttitudeLaw(value, err)
}

func GNSSAttitudeLawGPSIIF(yawRateRadS float64, yawBiasRad float64, shadowLimitRad float64) (*GNSSAttitudeLaw, error) {
	var value *C.SidereonGnssAttitudeLaw
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_gnss_attitude_law_gps_iif(C.double(yawRateRadS), C.double(yawBiasRad), C.double(shadowLimitRad), &value))
	})
	return newOrbitGNSSAttitudeLaw(value, err)
}

func GNSSAttitudeLawGPSIIR(yawRateRadS float64) (*GNSSAttitudeLaw, error) {
	var value *C.SidereonGnssAttitudeLaw
	err := callStatus(func() uint32 { return uint32(C.sidereon_gnss_attitude_law_gps_iir(C.double(yawRateRadS), &value)) })
	return newOrbitGNSSAttitudeLaw(value, err)
}

func GNSSAttitudeLawOrbitNormal() (*GNSSAttitudeLaw, error) {
	var value *C.SidereonGnssAttitudeLaw
	err := callStatus(func() uint32 { return uint32(C.sidereon_gnss_attitude_law_orbit_normal(&value)) })
	return newOrbitGNSSAttitudeLaw(value, err)
}

func GNSSBlockBeiDou3MEO(manufacturer *BeiDouManufacturer) (GNSSBlock, error) {
	var manufacturerC *C.uint32_t
	if manufacturer != nil {
		value := C.uint32_t(*manufacturer)
		manufacturerC = &value
	}
	var value C.enum_SidereonGnssBlock
	err := callStatus(func() uint32 { return uint32(C.sidereon_gnss_block_beidou3_meo(manufacturerC, &value)) })
	return GNSSBlock(value), err
}

func GNSSBlockCompare(left GNSSBlock, right GNSSBlock) (int32, error) {
	var value C.int32_t
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_gnss_block_compare(C.uint32_t(left), C.uint32_t(right), &value))
	})
	return int32(value), err
}

func GNSSBlockFromConstellationBlockType(system uint32, data []byte, beidouOrbit *BeiDouOrbitClass) (GNSSBlock, error) {
	var beidouOrbitC *C.uint32_t
	if beidouOrbit != nil {
		value := C.uint32_t(*beidouOrbit)
		beidouOrbitC = &value
	}
	var value C.enum_SidereonGnssBlock
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_gnss_block_from_constellation_block_type(C.uint32_t(system), (*C.uint8_t)(unsafe.Pointer(unsafe.SliceData(data))), C.size_t(len(data)), beidouOrbitC, &value))
	})
	return GNSSBlock(value), err
}

func GNSSBlockFromIGSName(data []byte) (GNSSBlock, error) {
	var value C.enum_SidereonGnssBlock
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_gnss_block_from_igs_name((*C.uint8_t)(unsafe.Pointer(unsafe.SliceData(data))), C.size_t(len(data)), &value))
	})
	return GNSSBlock(value), err
}

func GNSSBlockIGSName(block GNSSBlock) (string, error) {
	return orbitText("gnss block igs name", func(out *C.uint8_t, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_gnss_block_igs_name(C.uint32_t(block), out, capacity, written, required)
	})
}

func GNSSBlockSystem(block GNSSBlock) (uint32, error) {
	var value C.enum_SidereonGnssSystem
	err := callStatus(func() uint32 { return uint32(C.sidereon_gnss_block_system(C.uint32_t(block), &value)) })
	return uint32(value), err
}

func GPSIIADefaultYawBiasRad() float64 {
	value := C.sidereon_gps_iia_default_yaw_bias_rad()
	return float64(value)
}

func GPSIIADefaultYawRateRadS(prn uint8) (float64, error) {
	var value C.double
	err := callStatus(func() uint32 { return uint32(C.sidereon_gps_iia_default_yaw_rate_rad_s(C.uint8_t(prn), &value)) })
	return float64(value), err
}

func GPSIIFDefaultYawBiasRad() float64 {
	value := C.sidereon_gps_iif_default_yaw_bias_rad()
	return float64(value)
}

func GPSIIFDefaultYawRateRadS() float64 {
	value := C.sidereon_gps_iif_default_yaw_rate_rad_s()
	return float64(value)
}

func GPSIIIBetaXRad() float64 {
	value := C.sidereon_gps_iii_beta_x_rad()
	return float64(value)
}

func GPSIIIBetaYRad() float64 {
	value := C.sidereon_gps_iii_beta_y_rad()
	return float64(value)
}

func GPSIIRDefaultYawRateRadS() float64 {
	value := C.sidereon_gps_iir_default_yaw_rate_rad_s()
	return float64(value)
}

func GPSShadowLimitRad() float64 {
	value := C.sidereon_gps_shadow_limit_rad()
	return float64(value)
}

func ComputeNominalYaw(sample AttitudeSample) (NominalYaw, error) {
	sampleC := orbitAttitudeSampleToC(sample)
	var value C.SidereonNominalYaw
	err := callStatus(func() uint32 { return uint32(C.sidereon_nominal_yaw(&sampleC, &value)) })
	return orbitNominalYawFromC(value), err
}
