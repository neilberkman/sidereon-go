//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#include <sidereon.h>
#include <stdlib.h>
enum SidereonStatus sidereon_rtcm_message_trailing_bits(const struct SidereonRtcmMessages *, size_t, bool *, size_t, size_t *, size_t *);
enum SidereonStatus sidereon_rtcm_message_unsupported_body(const struct SidereonRtcmMessages *, size_t, uint8_t *, size_t, size_t *, size_t *);
enum SidereonStatus sidereon_rtcm_build_unsupported(uint16_t, const uint8_t *, size_t, struct SidereonRtcmMessages **);
enum SidereonStatus sidereon_rtcm_message_with_trailing_bits(const struct SidereonRtcmMessages *, size_t, const bool *, size_t, struct SidereonRtcmMessages **);
enum SidereonStatus sidereon_rtcm_message_ssr_info_v2(const struct SidereonRtcmMessages *, size_t, SidereonRtcmSsrInfoV2 *);
enum SidereonStatus sidereon_rtcm_build_ssr_v2(const SidereonRtcmSsrInfoV2 *, const struct SidereonRtcmSsrOrbitRecord *, size_t, const struct SidereonRtcmSsrClockRecord *, size_t, const struct SidereonRtcmSsrUraRecord *, size_t, const struct SidereonRtcmSsrCodeBiasRecord *, size_t, const struct SidereonRtcmSsrCodeBiasSignal *, size_t, const struct SidereonRtcmSsrPhaseBiasRecord *, size_t, const struct SidereonRtcmSsrPhaseBiasSignal *, size_t, const bool *, size_t, struct SidereonRtcmMessages **);
*/
import "C"

import (
	"runtime"
	"unsafe"
)

type NativeRTCMLegacyL1 struct {
	CodeIndicator                  bool
	Pseudorange                    uint32
	PhaseRangeMinusPseudorange     int32
	LockTimeIndicator              uint8
	HasPseudorangeModulusAmbiguity bool
	PseudorangeModulusAmbiguity    uint8
	HasCNR                         bool
	CNR                            uint8
}
type NativeRTCMLegacyL2 struct {
	CodeIndicator                uint8
	PseudorangeDifference        int16
	PhaseRangeMinusL1Pseudorange int32
	LockTimeIndicator            uint8
	HasCNR                       bool
	CNR                          uint8
}
type NativeRTCMLegacySatellite struct {
	SatelliteID         uint8
	HasFrequencyChannel bool
	FrequencyChannel    uint8
	L1                  NativeRTCMLegacyL1
	HasL2               bool
	L2                  NativeRTCMLegacyL2
}
type NativeRTCMLegacyObservations struct {
	MessageNumber, ReferenceStationID uint16
	EpochTime                         uint32
	SynchronousGNSS                   bool
	SatelliteCount                    uint8
	DivergenceFreeSmoothing           bool
	SmoothingInterval                 uint8
	Satellites                        []NativeRTCMLegacySatellite
	TrailingBits                      []bool
}
type NativeRTCMMessageAnnouncement struct {
	MessageNumber uint16
	Synchronous   bool
	Interval      uint16
}
type NativeRTCMSystemParameters struct {
	ReferenceStationID, MJD        uint16
	SecondsOfDay                   uint32
	AnnouncementCount, LeapSeconds uint8
	Announcements                  []NativeRTCMMessageAnnouncement
	TrailingBits                   []bool
}
type NativeRTCMText struct {
	ReferenceStationID, MJD uint16
	SecondsOfDay            uint32
	CharacterCount          uint8
	CodeUnits               []byte
	TrailingBits            []bool
}
type NativeRTCMNetworkAuxiliaryStation struct {
	NetworkID, SubnetworkID, AuxiliaryStationCount uint8
	MasterStationID, AuxiliaryStationID            uint16
	DeltaLatitude, DeltaLongitude, DeltaHeight     int32
	TrailingBits                                   []bool
}
type NativeRTCMSSRVTECInfo struct {
	MessageNumber    uint16
	HasIGSSSRVersion bool
	IGSSSRVersion    uint8
	EpochTimeS       uint32
	UpdateInterval   uint8
	MultipleMessage  bool
	IODSSR           uint8
	ProviderID       uint16
	SolutionID       uint8
	QualityIndicator uint16
	Layers           []NativeRTCMTecLayer
	TrailingBits     []bool
}
type NativeRTCMTecLayer struct {
	Height, Degree, Order uint8
	Cosine, Sine          []int16
}
type NativeRTCMNetworkDifference struct {
	SatelliteID, AmbiguityStatus, NonSyncCount uint8
	HasGeometric                               bool
	Geometric                                  int32
	HasIOD                                     bool
	IOD                                        uint8
	HasIonospheric                             bool
	Ionospheric                                int32
}
type NativeRTCMNetworkDifferences struct {
	MessageNumber                       uint16
	NetworkID, SubnetworkID             uint8
	EpochTime                           uint32
	MultipleMessage                     bool
	MasterStationID, AuxiliaryStationID uint16
	SatelliteCount                      uint8
	Satellites                          []NativeRTCMNetworkDifference
	TrailingBits                        []bool
}
type NativeRTCMNetworkResidual struct {
	SatelliteID, SOC uint8
	SOD              uint16
	SOH              uint8
	SLC, SLD         uint16
}
type NativeRTCMNetworkResiduals struct {
	MessageNumber                         uint16
	EpochTime                             uint32
	ReferenceStationID                    uint16
	ReferenceStationCount, SatelliteCount uint8
	Satellites                            []NativeRTCMNetworkResidual
	TrailingBits                          []bool
}
type NativeRTCMFKPGradient struct {
	SatelliteID, IOD                                                 uint8
	GeometricNorth, GeometricEast, IonosphericNorth, IonosphericEast int16
}
type NativeRTCMFKPGradients struct {
	MessageNumber, ReferenceStationID uint16
	EpochTime                         uint32
	SatelliteCount                    uint8
	Satellites                        []NativeRTCMFKPGradient
	TrailingBits                      []bool
}
type NativeRTCMGridResidual struct{ Horizontal1, Horizontal2, Height int16 }
type NativeRTCMResidualGrid struct {
	MessageNumber                                                                      uint16
	SystemID                                                                           uint8
	HorizontalShift, VerticalShift                                                     bool
	Origin1, Origin2                                                                   int32
	Extension1, Extension2                                                             uint16
	MeanOffset1, MeanOffset2, MeanHeightOffset                                         int16
	Residuals                                                                          [16]NativeRTCMGridResidual
	HorizontalInterpolation, VerticalInterpolation, HorizontalQuality, VerticalQuality uint8
	MJD                                                                                uint16
	TrailingBits                                                                       []bool
}
type NativeRTCMProjection struct {
	MessageNumber                                             uint16
	SystemID, ProjectionType                                  uint8
	Rectification                                             bool
	Latitude, Longitude, StandardParallel1, StandardParallel2 int64
	Azimuth                                                   uint64
	RectifiedToSkew                                           int32
	AddScale                                                  uint32
	Easting                                                   uint64
	Northing                                                  int64
	TrailingBits                                              []bool
}
type NativeRTCMPhysicalReferenceStation struct {
	NonPhysicalStationID, PhysicalStationID uint16
	ITRFRealizationYear                     uint8
	ECEFX, ECEFY, ECEFZ                     int64
	TrailingBits                            []bool
}
type NativeRTCMHelmertTransformation struct {
	MessageNumber                                         uint16
	SourceName, TargetName                                string
	SystemID                                              uint8
	UtilizedMessages                                      uint16
	PlateNumber, ComputationIndicator, HeightIndicator    uint8
	ValidityLatitude, ValidityLongitude                   int32
	ValidityExtensionLatitude, ValidityExtensionLongitude uint16
	DX, DY, DZ, R1, R2, R3, DS                            int32
	HasRotationPoint                                      bool
	RotationPointX, RotationPointY, RotationPointZ        int64
	AddAS, AddBS, AddAT, AddBT                            uint32
	HorizontalQuality, VerticalQuality                    uint8
	TrailingBits                                          []bool
}
type NativeRTCMGLONASSCodePhaseBiases struct {
	ReferenceStationID               uint16
	Aligned                          bool
	Reserved                         uint8
	HasL1CA, HasL1P, HasL2CA, HasL2P bool
	L1CA, L1P, L2CA, L2P             int16
	TrailingBits                     []bool
}

func (m *RtcmMessages) HelmertTransformation(index int) (NativeRTCMHelmertTransformation, error) {
	if index < 0 {
		return NativeRTCMHelmertTransformation{}, errNegativeIndex
	}
	var value C.SidereonRtcmHelmertTransformation
	var result NativeRTCMHelmertTransformation
	err := m.resource.with(func(pointer unsafe.Pointer) error {
		if err := callStatus(func() uint32 {
			return uint32(C.sidereon_rtcm_message_helmert_transformation((*C.SidereonRtcmMessages)(pointer), C.size_t(index), &value))
		}); err != nil {
			return err
		}
		source := make([]byte, int(value.source_name_len))
		target := make([]byte, int(value.target_name_len))
		for row := range source {
			source[row] = byte(value.source_name[row])
		}
		for row := range target {
			target[row] = byte(value.target_name[row])
		}
		result = NativeRTCMHelmertTransformation{MessageNumber: uint16(value.message_number), SourceName: string(source), TargetName: string(target), SystemID: uint8(value.system_id), UtilizedMessages: uint16(value.utilized_messages), PlateNumber: uint8(value.plate_number), ComputationIndicator: uint8(value.computation_indicator), HeightIndicator: uint8(value.height_indicator), ValidityLatitude: int32(value.validity_latitude), ValidityLongitude: int32(value.validity_longitude), ValidityExtensionLatitude: uint16(value.validity_extension_latitude), ValidityExtensionLongitude: uint16(value.validity_extension_longitude), DX: int32(value.dx), DY: int32(value.dy), DZ: int32(value.dz), R1: int32(value.r1), R2: int32(value.r2), R3: int32(value.r3), DS: int32(value.ds), HasRotationPoint: bool(value.has_rotation_point), RotationPointX: int64(value.rotation_point_x), RotationPointY: int64(value.rotation_point_y), RotationPointZ: int64(value.rotation_point_z), AddAS: uint32(value.add_as), AddBS: uint32(value.add_bs), AddAT: uint32(value.add_at), AddBT: uint32(value.add_bt), HorizontalQuality: uint8(value.horizontal_quality), VerticalQuality: uint8(value.vertical_quality)}
		return nil
	})
	if err == nil {
		result.TrailingBits, err = m.TrailingBits(index)
	}
	runtime.KeepAlive(m)
	return result, err
}

func BuildRTCMHelmertTransformation(value NativeRTCMHelmertTransformation) (*RtcmMessages, error) {
	if len(value.SourceName) > 31 || len(value.TargetName) > 31 {
		return nil, invalidArgument("Helmert name exceeds 31 bytes")
	}
	tails := make([]C.bool, len(value.TrailingBits))
	for index, bit := range value.TrailingBits {
		tails[index] = C.bool(bit)
	}
	var fields C.SidereonRtcmHelmertTransformation
	fields.message_number = C.uint16_t(value.MessageNumber)
	fields.source_name_len = C.uint8_t(len(value.SourceName))
	fields.target_name_len = C.uint8_t(len(value.TargetName))
	for index, unit := range []byte(value.SourceName) {
		fields.source_name[index] = C.uint8_t(unit)
	}
	for index, unit := range []byte(value.TargetName) {
		fields.target_name[index] = C.uint8_t(unit)
	}
	fields.system_id = C.uint8_t(value.SystemID)
	fields.utilized_messages = C.uint16_t(value.UtilizedMessages)
	fields.plate_number = C.uint8_t(value.PlateNumber)
	fields.computation_indicator = C.uint8_t(value.ComputationIndicator)
	fields.height_indicator = C.uint8_t(value.HeightIndicator)
	fields.validity_latitude = C.int32_t(value.ValidityLatitude)
	fields.validity_longitude = C.int32_t(value.ValidityLongitude)
	fields.validity_extension_latitude = C.uint16_t(value.ValidityExtensionLatitude)
	fields.validity_extension_longitude = C.uint16_t(value.ValidityExtensionLongitude)
	fields.dx = C.int32_t(value.DX)
	fields.dy = C.int32_t(value.DY)
	fields.dz = C.int32_t(value.DZ)
	fields.r1 = C.int32_t(value.R1)
	fields.r2 = C.int32_t(value.R2)
	fields.r3 = C.int32_t(value.R3)
	fields.ds = C.int32_t(value.DS)
	fields.has_rotation_point = C.bool(value.HasRotationPoint)
	fields.rotation_point_x = C.int64_t(value.RotationPointX)
	fields.rotation_point_y = C.int64_t(value.RotationPointY)
	fields.rotation_point_z = C.int64_t(value.RotationPointZ)
	fields.add_as = C.uint32_t(value.AddAS)
	fields.add_bs = C.uint32_t(value.AddBS)
	fields.add_at = C.uint32_t(value.AddAT)
	fields.add_bt = C.uint32_t(value.AddBT)
	fields.horizontal_quality = C.uint8_t(value.HorizontalQuality)
	fields.vertical_quality = C.uint8_t(value.VerticalQuality)
	fields.trailing_bit_count = C.size_t(len(tails))
	return buildRTCMMessage(func(out **C.SidereonRtcmMessages) uint32 {
		return uint32(C.sidereon_rtcm_build_helmert_transformation(&fields, (*C.bool)(unsafe.Pointer(unsafe.SliceData(tails))), C.size_t(len(tails)), out))
	})
}

func (m *RtcmMessages) GLONASSCodePhaseBiases(index int) (NativeRTCMGLONASSCodePhaseBiases, error) {
	if index < 0 {
		return NativeRTCMGLONASSCodePhaseBiases{}, errNegativeIndex
	}
	var value C.SidereonRtcmGlonassCodePhaseBiases
	var result NativeRTCMGLONASSCodePhaseBiases
	err := m.resource.with(func(pointer unsafe.Pointer) error {
		if err := callStatus(func() uint32 {
			return uint32(C.sidereon_rtcm_message_glonass_code_phase_biases((*C.SidereonRtcmMessages)(pointer), C.size_t(index), &value))
		}); err != nil {
			return err
		}
		result = NativeRTCMGLONASSCodePhaseBiases{ReferenceStationID: uint16(value.reference_station_id), Aligned: bool(value.aligned), Reserved: uint8(value.reserved), HasL1CA: bool(value.has_l1_ca), L1CA: int16(value.l1_ca), HasL1P: bool(value.has_l1_p), L1P: int16(value.l1_p), HasL2CA: bool(value.has_l2_ca), L2CA: int16(value.l2_ca), HasL2P: bool(value.has_l2_p), L2P: int16(value.l2_p)}
		return nil
	})
	if err == nil {
		result.TrailingBits, err = m.TrailingBits(index)
	}
	runtime.KeepAlive(m)
	return result, err
}

func BuildRTCMGLONASSCodePhaseBiases(value NativeRTCMGLONASSCodePhaseBiases) (*RtcmMessages, error) {
	tails := make([]C.bool, len(value.TrailingBits))
	for index, bit := range value.TrailingBits {
		tails[index] = C.bool(bit)
	}
	fields := C.SidereonRtcmGlonassCodePhaseBiases{reference_station_id: C.uint16_t(value.ReferenceStationID), aligned: C.bool(value.Aligned), reserved: C.uint8_t(value.Reserved), has_l1_ca: C.bool(value.HasL1CA), l1_ca: C.int16_t(value.L1CA), has_l1_p: C.bool(value.HasL1P), l1_p: C.int16_t(value.L1P), has_l2_ca: C.bool(value.HasL2CA), l2_ca: C.int16_t(value.L2CA), has_l2_p: C.bool(value.HasL2P), l2_p: C.int16_t(value.L2P), trailing_bit_count: C.size_t(len(tails))}
	return buildRTCMMessage(func(out **C.SidereonRtcmMessages) uint32 {
		return uint32(C.sidereon_rtcm_build_glonass_code_phase_biases(&fields, (*C.bool)(unsafe.Pointer(unsafe.SliceData(tails))), C.size_t(len(tails)), out))
	})
}

func (m *RtcmMessages) ResidualGrid(index int) (NativeRTCMResidualGrid, error) {
	if index < 0 {
		return NativeRTCMResidualGrid{}, errNegativeIndex
	}
	var value C.SidereonRtcmResidualGrid
	var result NativeRTCMResidualGrid
	err := m.resource.with(func(pointer unsafe.Pointer) error {
		if err := callStatus(func() uint32 {
			return uint32(C.sidereon_rtcm_message_residual_grid((*C.SidereonRtcmMessages)(pointer), C.size_t(index), &value))
		}); err != nil {
			return err
		}
		result = NativeRTCMResidualGrid{MessageNumber: uint16(value.message_number), SystemID: uint8(value.system_id), HorizontalShift: bool(value.horizontal_shift), VerticalShift: bool(value.vertical_shift), Origin1: int32(value.origin_1), Origin2: int32(value.origin_2), Extension1: uint16(value.extension_1), Extension2: uint16(value.extension_2), MeanOffset1: int16(value.mean_offset_1), MeanOffset2: int16(value.mean_offset_2), MeanHeightOffset: int16(value.mean_height_offset), HorizontalInterpolation: uint8(value.horizontal_interpolation), VerticalInterpolation: uint8(value.vertical_interpolation), HorizontalQuality: uint8(value.horizontal_quality), VerticalQuality: uint8(value.vertical_quality), MJD: uint16(value.mjd)}
		for row := range result.Residuals {
			result.Residuals[row] = NativeRTCMGridResidual{Horizontal1: int16(value.residuals[row].horizontal_1), Horizontal2: int16(value.residuals[row].horizontal_2), Height: int16(value.residuals[row].height)}
		}
		return nil
	})
	if err == nil {
		result.TrailingBits, err = m.TrailingBits(index)
	}
	runtime.KeepAlive(m)
	return result, err
}

func BuildRTCMResidualGrid(value NativeRTCMResidualGrid) (*RtcmMessages, error) {
	var residuals [16]C.SidereonRtcmGridResidual
	for row, item := range value.Residuals {
		residuals[row] = C.SidereonRtcmGridResidual{horizontal_1: C.int16_t(item.Horizontal1), horizontal_2: C.int16_t(item.Horizontal2), height: C.int16_t(item.Height)}
	}
	tails := make([]C.bool, len(value.TrailingBits))
	for index, bit := range value.TrailingBits {
		tails[index] = C.bool(bit)
	}
	fields := C.SidereonRtcmResidualGrid{message_number: C.uint16_t(value.MessageNumber), system_id: C.uint8_t(value.SystemID), horizontal_shift: C.bool(value.HorizontalShift), vertical_shift: C.bool(value.VerticalShift), origin_1: C.int32_t(value.Origin1), origin_2: C.int32_t(value.Origin2), extension_1: C.uint16_t(value.Extension1), extension_2: C.uint16_t(value.Extension2), mean_offset_1: C.int16_t(value.MeanOffset1), mean_offset_2: C.int16_t(value.MeanOffset2), mean_height_offset: C.int16_t(value.MeanHeightOffset), horizontal_interpolation: C.uint8_t(value.HorizontalInterpolation), vertical_interpolation: C.uint8_t(value.VerticalInterpolation), horizontal_quality: C.uint8_t(value.HorizontalQuality), vertical_quality: C.uint8_t(value.VerticalQuality), mjd: C.uint16_t(value.MJD), trailing_bit_count: C.size_t(len(tails))}
	for row, item := range residuals {
		fields.residuals[row] = item
	}
	return buildRTCMMessage(func(out **C.SidereonRtcmMessages) uint32 {
		return uint32(C.sidereon_rtcm_build_residual_grid(&fields, (*C.bool)(unsafe.Pointer(unsafe.SliceData(tails))), C.size_t(len(tails)), out))
	})
}

func (m *RtcmMessages) Projection(index int) (NativeRTCMProjection, error) {
	if index < 0 {
		return NativeRTCMProjection{}, errNegativeIndex
	}
	var value C.SidereonRtcmProjection
	var result NativeRTCMProjection
	err := m.resource.with(func(pointer unsafe.Pointer) error {
		if err := callStatus(func() uint32 {
			return uint32(C.sidereon_rtcm_message_projection((*C.SidereonRtcmMessages)(pointer), C.size_t(index), &value))
		}); err != nil {
			return err
		}
		result = NativeRTCMProjection{MessageNumber: uint16(value.message_number), SystemID: uint8(value.system_id), ProjectionType: uint8(value.projection_type), Rectification: bool(value.rectification), Latitude: int64(value.latitude), Longitude: int64(value.longitude), StandardParallel1: int64(value.standard_parallel_1), StandardParallel2: int64(value.standard_parallel_2), Azimuth: uint64(value.azimuth), RectifiedToSkew: int32(value.rectified_to_skew), AddScale: uint32(value.add_scale), Easting: uint64(value.easting), Northing: int64(value.northing)}
		return nil
	})
	if err == nil {
		result.TrailingBits, err = m.TrailingBits(index)
	}
	runtime.KeepAlive(m)
	return result, err
}

func BuildRTCMProjection(value NativeRTCMProjection) (*RtcmMessages, error) {
	tails := make([]C.bool, len(value.TrailingBits))
	for index, bit := range value.TrailingBits {
		tails[index] = C.bool(bit)
	}
	fields := C.SidereonRtcmProjection{message_number: C.uint16_t(value.MessageNumber), system_id: C.uint8_t(value.SystemID), projection_type: C.uint8_t(value.ProjectionType), rectification: C.bool(value.Rectification), latitude: C.int64_t(value.Latitude), longitude: C.int64_t(value.Longitude), standard_parallel_1: C.int64_t(value.StandardParallel1), standard_parallel_2: C.int64_t(value.StandardParallel2), azimuth: C.uint64_t(value.Azimuth), rectified_to_skew: C.int32_t(value.RectifiedToSkew), add_scale: C.uint32_t(value.AddScale), easting: C.uint64_t(value.Easting), northing: C.int64_t(value.Northing), trailing_bit_count: C.size_t(len(tails))}
	return buildRTCMMessage(func(out **C.SidereonRtcmMessages) uint32 {
		return uint32(C.sidereon_rtcm_build_projection(&fields, (*C.bool)(unsafe.Pointer(unsafe.SliceData(tails))), C.size_t(len(tails)), out))
	})
}

func (m *RtcmMessages) PhysicalReferenceStation(index int) (NativeRTCMPhysicalReferenceStation, error) {
	if index < 0 {
		return NativeRTCMPhysicalReferenceStation{}, errNegativeIndex
	}
	var value C.SidereonRtcmPhysicalReferenceStation
	var result NativeRTCMPhysicalReferenceStation
	err := m.resource.with(func(pointer unsafe.Pointer) error {
		if err := callStatus(func() uint32 {
			return uint32(C.sidereon_rtcm_message_physical_reference_station((*C.SidereonRtcmMessages)(pointer), C.size_t(index), &value))
		}); err != nil {
			return err
		}
		result = NativeRTCMPhysicalReferenceStation{NonPhysicalStationID: uint16(value.non_physical_station_id), PhysicalStationID: uint16(value.physical_station_id), ITRFRealizationYear: uint8(value.itrf_realization_year), ECEFX: int64(value.ecef_x), ECEFY: int64(value.ecef_y), ECEFZ: int64(value.ecef_z)}
		return nil
	})
	if err == nil {
		result.TrailingBits, err = m.TrailingBits(index)
	}
	runtime.KeepAlive(m)
	return result, err
}

func BuildRTCMPhysicalReferenceStation(value NativeRTCMPhysicalReferenceStation) (*RtcmMessages, error) {
	tails := make([]C.bool, len(value.TrailingBits))
	for index, bit := range value.TrailingBits {
		tails[index] = C.bool(bit)
	}
	fields := C.SidereonRtcmPhysicalReferenceStation{non_physical_station_id: C.uint16_t(value.NonPhysicalStationID), physical_station_id: C.uint16_t(value.PhysicalStationID), itrf_realization_year: C.uint8_t(value.ITRFRealizationYear), ecef_x: C.int64_t(value.ECEFX), ecef_y: C.int64_t(value.ECEFY), ecef_z: C.int64_t(value.ECEFZ), trailing_bit_count: C.size_t(len(tails))}
	return buildRTCMMessage(func(out **C.SidereonRtcmMessages) uint32 {
		return uint32(C.sidereon_rtcm_build_physical_reference_station(&fields, (*C.bool)(unsafe.Pointer(unsafe.SliceData(tails))), C.size_t(len(tails)), out))
	})
}

func (m *RtcmMessages) NetworkDifferences(index int) (NativeRTCMNetworkDifferences, error) {
	if index < 0 {
		return NativeRTCMNetworkDifferences{}, errNegativeIndex
	}
	var result NativeRTCMNetworkDifferences
	err := m.resource.with(func(pointer unsafe.Pointer) error {
		messagePointer := (*C.SidereonRtcmMessages)(pointer)
		var fields C.SidereonRtcmNetworkDifferences
		if err := callStatus(func() uint32 {
			return uint32(C.sidereon_rtcm_message_network_differences(messagePointer, C.size_t(index), &fields))
		}); err != nil {
			return err
		}
		count, err := checkedNativeCount(uint64(fields.satellite_records))
		if err != nil {
			return err
		}
		records := make([]C.SidereonRtcmNetworkDifference, count)
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return uint32(C.sidereon_rtcm_message_network_difference_satellites(messagePointer, C.size_t(index), (*C.SidereonRtcmNetworkDifference)(unsafe.Pointer(unsafe.SliceData(records))), C.size_t(len(records)), &written, &required))
		}); err != nil {
			return err
		}
		result = NativeRTCMNetworkDifferences{MessageNumber: uint16(fields.message_number), NetworkID: uint8(fields.network_id), SubnetworkID: uint8(fields.subnetwork_id), EpochTime: uint32(fields.epoch_time), MultipleMessage: bool(fields.multiple_message), MasterStationID: uint16(fields.master_station_id), AuxiliaryStationID: uint16(fields.auxiliary_station_id), SatelliteCount: uint8(fields.satellite_count), Satellites: make([]NativeRTCMNetworkDifference, int(written))}
		for row, record := range records[:int(written)] {
			result.Satellites[row] = NativeRTCMNetworkDifference{SatelliteID: uint8(record.satellite_id), AmbiguityStatus: uint8(record.ambiguity_status), NonSyncCount: uint8(record.non_sync_count), HasGeometric: bool(record.has_geometric), Geometric: int32(record.geometric), HasIOD: bool(record.has_iod), IOD: uint8(record.iod), HasIonospheric: bool(record.has_ionospheric), Ionospheric: int32(record.ionospheric)}
		}
		return nil
	})
	if err == nil {
		result.TrailingBits, err = m.TrailingBits(index)
	}
	runtime.KeepAlive(m)
	return result, err
}

func BuildRTCMNetworkDifferences(value NativeRTCMNetworkDifferences) (*RtcmMessages, error) {
	records := make([]C.SidereonRtcmNetworkDifference, len(value.Satellites))
	for index, row := range value.Satellites {
		records[index] = C.SidereonRtcmNetworkDifference{satellite_id: C.uint8_t(row.SatelliteID), ambiguity_status: C.uint8_t(row.AmbiguityStatus), non_sync_count: C.uint8_t(row.NonSyncCount), has_geometric: C.bool(row.HasGeometric), geometric: C.int32_t(row.Geometric), has_iod: C.bool(row.HasIOD), iod: C.uint8_t(row.IOD), has_ionospheric: C.bool(row.HasIonospheric), ionospheric: C.int32_t(row.Ionospheric)}
	}
	tails := make([]C.bool, len(value.TrailingBits))
	for index, bit := range value.TrailingBits {
		tails[index] = C.bool(bit)
	}
	fields := C.SidereonRtcmNetworkDifferences{message_number: C.uint16_t(value.MessageNumber), network_id: C.uint8_t(value.NetworkID), subnetwork_id: C.uint8_t(value.SubnetworkID), epoch_time: C.uint32_t(value.EpochTime), multiple_message: C.bool(value.MultipleMessage), master_station_id: C.uint16_t(value.MasterStationID), auxiliary_station_id: C.uint16_t(value.AuxiliaryStationID), satellite_count: C.uint8_t(value.SatelliteCount), satellite_records: C.size_t(len(records)), trailing_bit_count: C.size_t(len(tails))}
	return buildRTCMMessage(func(out **C.SidereonRtcmMessages) uint32 {
		return uint32(C.sidereon_rtcm_build_network_differences(&fields, (*C.SidereonRtcmNetworkDifference)(unsafe.Pointer(unsafe.SliceData(records))), C.size_t(len(records)), (*C.bool)(unsafe.Pointer(unsafe.SliceData(tails))), C.size_t(len(tails)), out))
	})
}

func (m *RtcmMessages) NetworkResiduals(index int) (NativeRTCMNetworkResiduals, error) {
	if index < 0 {
		return NativeRTCMNetworkResiduals{}, errNegativeIndex
	}
	var result NativeRTCMNetworkResiduals
	err := m.resource.with(func(pointer unsafe.Pointer) error {
		messagePointer := (*C.SidereonRtcmMessages)(pointer)
		var fields C.SidereonRtcmNetworkResiduals
		if err := callStatus(func() uint32 {
			return uint32(C.sidereon_rtcm_message_network_residuals(messagePointer, C.size_t(index), &fields))
		}); err != nil {
			return err
		}
		count, err := checkedNativeCount(uint64(fields.satellite_records))
		if err != nil {
			return err
		}
		records := make([]C.SidereonRtcmNetworkResidual, count)
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return uint32(C.sidereon_rtcm_message_network_residual_satellites(messagePointer, C.size_t(index), (*C.SidereonRtcmNetworkResidual)(unsafe.Pointer(unsafe.SliceData(records))), C.size_t(len(records)), &written, &required))
		}); err != nil {
			return err
		}
		result = NativeRTCMNetworkResiduals{MessageNumber: uint16(fields.message_number), EpochTime: uint32(fields.epoch_time), ReferenceStationID: uint16(fields.reference_station_id), ReferenceStationCount: uint8(fields.reference_station_count), SatelliteCount: uint8(fields.satellite_count), Satellites: make([]NativeRTCMNetworkResidual, int(written))}
		for row, record := range records[:int(written)] {
			result.Satellites[row] = NativeRTCMNetworkResidual{SatelliteID: uint8(record.satellite_id), SOC: uint8(record.s_oc), SOD: uint16(record.s_od), SOH: uint8(record.s_oh), SLC: uint16(record.s_lc), SLD: uint16(record.s_ld)}
		}
		return nil
	})
	if err == nil {
		result.TrailingBits, err = m.TrailingBits(index)
	}
	runtime.KeepAlive(m)
	return result, err
}

func (m *RtcmMessages) FKPGradients(index int) (NativeRTCMFKPGradients, error) {
	if index < 0 {
		return NativeRTCMFKPGradients{}, errNegativeIndex
	}
	var result NativeRTCMFKPGradients
	err := m.resource.with(func(pointer unsafe.Pointer) error {
		messagePointer := (*C.SidereonRtcmMessages)(pointer)
		var fields C.SidereonRtcmFkpGradients
		if err := callStatus(func() uint32 {
			return uint32(C.sidereon_rtcm_message_fkp_gradients(messagePointer, C.size_t(index), &fields))
		}); err != nil {
			return err
		}
		count, err := checkedNativeCount(uint64(fields.satellite_records))
		if err != nil {
			return err
		}
		records := make([]C.SidereonRtcmFkpGradient, count)
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return uint32(C.sidereon_rtcm_message_fkp_gradient_satellites(messagePointer, C.size_t(index), (*C.SidereonRtcmFkpGradient)(unsafe.Pointer(unsafe.SliceData(records))), C.size_t(len(records)), &written, &required))
		}); err != nil {
			return err
		}
		result = NativeRTCMFKPGradients{MessageNumber: uint16(fields.message_number), ReferenceStationID: uint16(fields.reference_station_id), EpochTime: uint32(fields.epoch_time), SatelliteCount: uint8(fields.satellite_count), Satellites: make([]NativeRTCMFKPGradient, int(written))}
		for row, record := range records[:int(written)] {
			result.Satellites[row] = NativeRTCMFKPGradient{SatelliteID: uint8(record.satellite_id), IOD: uint8(record.iod), GeometricNorth: int16(record.geometric_north), GeometricEast: int16(record.geometric_east), IonosphericNorth: int16(record.ionospheric_north), IonosphericEast: int16(record.ionospheric_east)}
		}
		return nil
	})
	if err == nil {
		result.TrailingBits, err = m.TrailingBits(index)
	}
	runtime.KeepAlive(m)
	return result, err
}

func BuildRTCMNetworkResiduals(value NativeRTCMNetworkResiduals) (*RtcmMessages, error) {
	records := make([]C.SidereonRtcmNetworkResidual, len(value.Satellites))
	for index, row := range value.Satellites {
		records[index] = C.SidereonRtcmNetworkResidual{satellite_id: C.uint8_t(row.SatelliteID), s_oc: C.uint8_t(row.SOC), s_od: C.uint16_t(row.SOD), s_oh: C.uint8_t(row.SOH), s_lc: C.uint16_t(row.SLC), s_ld: C.uint16_t(row.SLD)}
	}
	tails := make([]C.bool, len(value.TrailingBits))
	for index, bit := range value.TrailingBits {
		tails[index] = C.bool(bit)
	}
	fields := C.SidereonRtcmNetworkResiduals{message_number: C.uint16_t(value.MessageNumber), epoch_time: C.uint32_t(value.EpochTime), reference_station_id: C.uint16_t(value.ReferenceStationID), reference_station_count: C.uint8_t(value.ReferenceStationCount), satellite_count: C.uint8_t(value.SatelliteCount), satellite_records: C.size_t(len(records)), trailing_bit_count: C.size_t(len(tails))}
	return buildRTCMMessage(func(out **C.SidereonRtcmMessages) uint32 {
		return uint32(C.sidereon_rtcm_build_network_residuals(&fields, (*C.SidereonRtcmNetworkResidual)(unsafe.Pointer(unsafe.SliceData(records))), C.size_t(len(records)), (*C.bool)(unsafe.Pointer(unsafe.SliceData(tails))), C.size_t(len(tails)), out))
	})
}

func BuildRTCMFKPGradients(value NativeRTCMFKPGradients) (*RtcmMessages, error) {
	records := make([]C.SidereonRtcmFkpGradient, len(value.Satellites))
	for index, row := range value.Satellites {
		records[index] = C.SidereonRtcmFkpGradient{satellite_id: C.uint8_t(row.SatelliteID), iod: C.uint8_t(row.IOD), geometric_north: C.int16_t(row.GeometricNorth), geometric_east: C.int16_t(row.GeometricEast), ionospheric_north: C.int16_t(row.IonosphericNorth), ionospheric_east: C.int16_t(row.IonosphericEast)}
	}
	tails := make([]C.bool, len(value.TrailingBits))
	for index, bit := range value.TrailingBits {
		tails[index] = C.bool(bit)
	}
	fields := C.SidereonRtcmFkpGradients{message_number: C.uint16_t(value.MessageNumber), reference_station_id: C.uint16_t(value.ReferenceStationID), epoch_time: C.uint32_t(value.EpochTime), satellite_count: C.uint8_t(value.SatelliteCount), satellite_records: C.size_t(len(records)), trailing_bit_count: C.size_t(len(tails))}
	return buildRTCMMessage(func(out **C.SidereonRtcmMessages) uint32 {
		return uint32(C.sidereon_rtcm_build_fkp_gradients(&fields, (*C.SidereonRtcmFkpGradient)(unsafe.Pointer(unsafe.SliceData(records))), C.size_t(len(records)), (*C.bool)(unsafe.Pointer(unsafe.SliceData(tails))), C.size_t(len(tails)), out))
	})
}

func (m *RtcmMessages) SSRVTEC(index int) (NativeRTCMSSRVTECInfo, error) {
	if index < 0 {
		return NativeRTCMSSRVTECInfo{}, errNegativeIndex
	}
	var result NativeRTCMSSRVTECInfo
	err := m.resource.with(func(pointer unsafe.Pointer) error {
		messagePointer := (*C.SidereonRtcmMessages)(pointer)
		var info C.SidereonRtcmSsrVtecInfo
		if err := callStatus(func() uint32 {
			return uint32(C.sidereon_rtcm_message_ssr_vtec_info(messagePointer, C.size_t(index), &info))
		}); err != nil {
			return err
		}
		count, err := checkedNativeCount(uint64(info.layer_count))
		if err != nil {
			return err
		}
		result = NativeRTCMSSRVTECInfo{MessageNumber: uint16(info.message_number), HasIGSSSRVersion: bool(info.has_igs_ssr_version), IGSSSRVersion: uint8(info.igs_ssr_version), EpochTimeS: uint32(info.epoch_time_s), UpdateInterval: uint8(info.update_interval), MultipleMessage: bool(info.multiple_message), IODSSR: uint8(info.iod_ssr), ProviderID: uint16(info.provider_id), SolutionID: uint8(info.solution_id), QualityIndicator: uint16(info.quality_indicator), Layers: make([]NativeRTCMTecLayer, count)}
		layers := make([]C.SidereonRtcmSsrVtecLayer, count)
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return uint32(C.sidereon_rtcm_message_ssr_vtec_layers(messagePointer, C.size_t(index), (*C.SidereonRtcmSsrVtecLayer)(unsafe.Pointer(unsafe.SliceData(layers))), C.size_t(len(layers)), &written, &required))
		}); err != nil {
			return err
		}
		for layerIndex, layer := range layers[:int(written)] {
			cosineCount, err := checkedNativeCount(uint64(layer.cosine_count))
			if err != nil {
				return err
			}
			sineCount, err := checkedNativeCount(uint64(layer.sine_count))
			if err != nil {
				return err
			}
			cosine := make([]C.int16_t, cosineCount)
			sine := make([]C.int16_t, sineCount)
			var coefficientWritten, coefficientRequired C.size_t
			if err := callStatus(func() uint32 {
				return uint32(C.sidereon_rtcm_message_ssr_vtec_coefficients(messagePointer, C.size_t(index), C.size_t(layerIndex), C.bool(true), (*C.int16_t)(unsafe.Pointer(unsafe.SliceData(cosine))), C.size_t(len(cosine)), &coefficientWritten, &coefficientRequired))
			}); err != nil {
				return err
			}
			cosineValues := make([]int16, int(coefficientWritten))
			for row := range cosineValues {
				cosineValues[row] = int16(cosine[row])
			}
			if err := callStatus(func() uint32 {
				return uint32(C.sidereon_rtcm_message_ssr_vtec_coefficients(messagePointer, C.size_t(index), C.size_t(layerIndex), C.bool(false), (*C.int16_t)(unsafe.Pointer(unsafe.SliceData(sine))), C.size_t(len(sine)), &coefficientWritten, &coefficientRequired))
			}); err != nil {
				return err
			}
			sineValues := make([]int16, int(coefficientWritten))
			for row := range sineValues {
				sineValues[row] = int16(sine[row])
			}
			result.Layers[layerIndex] = NativeRTCMTecLayer{Height: uint8(layer.height), Degree: uint8(layer.degree), Order: uint8(layer.order), Cosine: cosineValues, Sine: sineValues}
		}
		return nil
	})
	if err == nil {
		result.TrailingBits, err = m.TrailingBits(index)
	}
	runtime.KeepAlive(m)
	return result, err
}

func (m *RtcmMessages) NetworkAuxiliaryStation(index int) (NativeRTCMNetworkAuxiliaryStation, error) {
	if index < 0 {
		return NativeRTCMNetworkAuxiliaryStation{}, errNegativeIndex
	}
	var value C.SidereonRtcmNetworkAuxiliaryStation
	var result NativeRTCMNetworkAuxiliaryStation
	err := m.resource.with(func(pointer unsafe.Pointer) error {
		if err := callStatus(func() uint32 {
			return uint32(C.sidereon_rtcm_message_network_auxiliary_station((*C.SidereonRtcmMessages)(pointer), C.size_t(index), &value))
		}); err != nil {
			return err
		}
		result = NativeRTCMNetworkAuxiliaryStation{NetworkID: uint8(value.network_id), SubnetworkID: uint8(value.subnetwork_id), AuxiliaryStationCount: uint8(value.auxiliary_station_count), MasterStationID: uint16(value.master_station_id), AuxiliaryStationID: uint16(value.auxiliary_station_id), DeltaLatitude: int32(value.delta_latitude), DeltaLongitude: int32(value.delta_longitude), DeltaHeight: int32(value.delta_height)}
		return nil
	})
	if err == nil {
		result.TrailingBits, err = m.TrailingBits(index)
	}
	runtime.KeepAlive(m)
	return result, err
}

func BuildRTCMNetworkAuxiliaryStation(value NativeRTCMNetworkAuxiliaryStation) (*RtcmMessages, error) {
	tails := make([]C.bool, len(value.TrailingBits))
	for index, bit := range value.TrailingBits {
		tails[index] = C.bool(bit)
	}
	fields := C.SidereonRtcmNetworkAuxiliaryStation{network_id: C.uint8_t(value.NetworkID), subnetwork_id: C.uint8_t(value.SubnetworkID), auxiliary_station_count: C.uint8_t(value.AuxiliaryStationCount), master_station_id: C.uint16_t(value.MasterStationID), auxiliary_station_id: C.uint16_t(value.AuxiliaryStationID), delta_latitude: C.int32_t(value.DeltaLatitude), delta_longitude: C.int32_t(value.DeltaLongitude), delta_height: C.int32_t(value.DeltaHeight), trailing_bit_count: C.size_t(len(tails))}
	return buildRTCMMessage(func(out **C.SidereonRtcmMessages) uint32 {
		return uint32(C.sidereon_rtcm_build_network_auxiliary_station(&fields, (*C.bool)(unsafe.Pointer(unsafe.SliceData(tails))), C.size_t(len(tails)), out))
	})
}

func (m *RtcmMessages) LegacyObservations(index int) (NativeRTCMLegacyObservations, error) {
	if index < 0 {
		return NativeRTCMLegacyObservations{}, errNegativeIndex
	}
	var header C.SidereonRtcmLegacyHeader
	var result NativeRTCMLegacyObservations
	err := m.resource.with(func(pointer unsafe.Pointer) error {
		messagePointer := (*C.SidereonRtcmMessages)(pointer)
		if err := callStatus(func() uint32 {
			return uint32(C.sidereon_rtcm_message_legacy_header(messagePointer, C.size_t(index), &header))
		}); err != nil {
			return err
		}
		count, err := checkedNativeCount(uint64(header.satellite_records))
		if err != nil {
			return err
		}
		result = NativeRTCMLegacyObservations{MessageNumber: uint16(header.message_number), ReferenceStationID: uint16(header.reference_station_id), EpochTime: uint32(header.epoch_time), SynchronousGNSS: bool(header.synchronous_gnss), SatelliteCount: uint8(header.satellite_count), DivergenceFreeSmoothing: bool(header.divergence_free_smoothing), SmoothingInterval: uint8(header.smoothing_interval), Satellites: make([]NativeRTCMLegacySatellite, count)}
		for row := range result.Satellites {
			var value C.SidereonRtcmLegacySatellite
			if err := callStatus(func() uint32 {
				return uint32(C.sidereon_rtcm_message_legacy_satellite(messagePointer, C.size_t(index), C.size_t(row), &value))
			}); err != nil {
				return err
			}
			result.Satellites[row] = NativeRTCMLegacySatellite{SatelliteID: uint8(value.satellite_id), HasFrequencyChannel: bool(value.has_frequency_channel), FrequencyChannel: uint8(value.frequency_channel), L1: NativeRTCMLegacyL1{CodeIndicator: bool(value.l1.code_indicator), Pseudorange: uint32(value.l1.pseudorange), PhaseRangeMinusPseudorange: int32(value.l1.phase_range_minus_pseudorange), LockTimeIndicator: uint8(value.l1.lock_time_indicator), HasPseudorangeModulusAmbiguity: bool(value.l1.has_pseudorange_modulus_ambiguity), PseudorangeModulusAmbiguity: uint8(value.l1.pseudorange_modulus_ambiguity), HasCNR: bool(value.l1.has_cnr), CNR: uint8(value.l1.cnr)}, HasL2: bool(value.has_l2), L2: NativeRTCMLegacyL2{CodeIndicator: uint8(value.l2.code_indicator), PseudorangeDifference: int16(value.l2.pseudorange_difference), PhaseRangeMinusL1Pseudorange: int32(value.l2.phase_range_minus_l1_pseudorange), LockTimeIndicator: uint8(value.l2.lock_time_indicator), HasCNR: bool(value.l2.has_cnr), CNR: uint8(value.l2.cnr)}}
		}
		return nil
	})
	if err == nil {
		result.TrailingBits, err = m.TrailingBits(index)
	}
	runtime.KeepAlive(m)
	return result, err
}

func BuildRTCMLegacyObservations(value NativeRTCMLegacyObservations) (*RtcmMessages, error) {
	rows := make([]C.SidereonRtcmLegacySatellite, len(value.Satellites))
	for index, row := range value.Satellites {
		rows[index] = C.SidereonRtcmLegacySatellite{satellite_id: C.uint8_t(row.SatelliteID), has_frequency_channel: C.bool(row.HasFrequencyChannel), frequency_channel: C.uint8_t(row.FrequencyChannel), l1: C.SidereonRtcmLegacyL1{code_indicator: C.bool(row.L1.CodeIndicator), pseudorange: C.uint32_t(row.L1.Pseudorange), phase_range_minus_pseudorange: C.int32_t(row.L1.PhaseRangeMinusPseudorange), lock_time_indicator: C.uint8_t(row.L1.LockTimeIndicator), has_pseudorange_modulus_ambiguity: C.bool(row.L1.HasPseudorangeModulusAmbiguity), pseudorange_modulus_ambiguity: C.uint8_t(row.L1.PseudorangeModulusAmbiguity), has_cnr: C.bool(row.L1.HasCNR), cnr: C.uint8_t(row.L1.CNR)}, has_l2: C.bool(row.HasL2), l2: C.SidereonRtcmLegacyL2{code_indicator: C.uint8_t(row.L2.CodeIndicator), pseudorange_difference: C.int16_t(row.L2.PseudorangeDifference), phase_range_minus_l1_pseudorange: C.int32_t(row.L2.PhaseRangeMinusL1Pseudorange), lock_time_indicator: C.uint8_t(row.L2.LockTimeIndicator), has_cnr: C.bool(row.L2.HasCNR), cnr: C.uint8_t(row.L2.CNR)}}
	}
	tails := make([]C.bool, len(value.TrailingBits))
	for index, bit := range value.TrailingBits {
		tails[index] = C.bool(bit)
	}
	return buildRTCMMessage(func(out **C.SidereonRtcmMessages) uint32 {
		return uint32(C.sidereon_rtcm_build_legacy(C.uint16_t(value.MessageNumber), C.uint16_t(value.ReferenceStationID), C.uint32_t(value.EpochTime), C.bool(value.SynchronousGNSS), C.uint8_t(value.SatelliteCount), C.bool(value.DivergenceFreeSmoothing), C.uint8_t(value.SmoothingInterval), (*C.SidereonRtcmLegacySatellite)(unsafe.Pointer(unsafe.SliceData(rows))), C.size_t(len(rows)), (*C.bool)(unsafe.Pointer(unsafe.SliceData(tails))), C.size_t(len(tails)), out))
	})
}

func (m *RtcmMessages) SystemParameters(index int) (NativeRTCMSystemParameters, error) {
	if index < 0 {
		return NativeRTCMSystemParameters{}, errNegativeIndex
	}
	var value C.SidereonRtcmSystemParameters
	var result NativeRTCMSystemParameters
	err := m.resource.with(func(pointer unsafe.Pointer) error {
		messagePointer := (*C.SidereonRtcmMessages)(pointer)
		if err := callStatus(func() uint32 {
			return uint32(C.sidereon_rtcm_message_system_parameters(messagePointer, C.size_t(index), &value))
		}); err != nil {
			return err
		}
		count, err := checkedNativeCount(uint64(value.announcements))
		if err != nil {
			return err
		}
		result = NativeRTCMSystemParameters{ReferenceStationID: uint16(value.reference_station_id), MJD: uint16(value.mjd), SecondsOfDay: uint32(value.seconds_of_day), AnnouncementCount: uint8(value.announcement_count), LeapSeconds: uint8(value.leap_seconds), Announcements: make([]NativeRTCMMessageAnnouncement, count)}
		for row := range result.Announcements {
			var announcement C.SidereonRtcmMessageAnnouncement
			if err := callStatus(func() uint32 {
				return uint32(C.sidereon_rtcm_message_announcement(messagePointer, C.size_t(index), C.size_t(row), &announcement))
			}); err != nil {
				return err
			}
			result.Announcements[row] = NativeRTCMMessageAnnouncement{MessageNumber: uint16(announcement.message_number), Synchronous: bool(announcement.synchronous), Interval: uint16(announcement.interval)}
		}
		return nil
	})
	if err == nil {
		result.TrailingBits, err = m.TrailingBits(index)
	}
	runtime.KeepAlive(m)
	return result, err
}

func (m *RtcmMessages) Text(index int) (NativeRTCMText, error) {
	if index < 0 {
		return NativeRTCMText{}, errNegativeIndex
	}
	var result NativeRTCMText
	err := m.resource.with(func(pointer unsafe.Pointer) error {
		var value C.SidereonRtcmTextMessage
		var written, required C.size_t
		status := callStatus(func() uint32 {
			return uint32(C.sidereon_rtcm_message_text((*C.SidereonRtcmMessages)(pointer), C.size_t(index), &value, nil, 0, &written, &required))
		})
		if status != nil {
			return status
		}
		count, err := checkedNativeCount(uint64(required))
		if err != nil {
			return err
		}
		units := make([]C.uint8_t, count)
		if err := callStatus(func() uint32 {
			return uint32(C.sidereon_rtcm_message_text((*C.SidereonRtcmMessages)(pointer), C.size_t(index), &value, (*C.uint8_t)(unsafe.Pointer(unsafe.SliceData(units))), C.size_t(len(units)), &written, &required))
		}); err != nil {
			return err
		}
		copied := make([]byte, int(written))
		for row := range copied {
			copied[row] = byte(units[row])
		}
		result = NativeRTCMText{ReferenceStationID: uint16(value.reference_station_id), MJD: uint16(value.mjd), SecondsOfDay: uint32(value.seconds_of_day), CharacterCount: uint8(value.character_count), CodeUnits: copied}
		return nil
	})
	if err == nil {
		result.TrailingBits, err = m.TrailingBits(index)
	}
	runtime.KeepAlive(m)
	return result, err
}

func BuildRTCMSystemParameters(value NativeRTCMSystemParameters) (*RtcmMessages, error) {
	announcements := make([]C.SidereonRtcmMessageAnnouncement, len(value.Announcements))
	for index, row := range value.Announcements {
		announcements[index] = C.SidereonRtcmMessageAnnouncement{message_number: C.uint16_t(row.MessageNumber), synchronous: C.bool(row.Synchronous), interval: C.uint16_t(row.Interval)}
	}
	tails := make([]C.bool, len(value.TrailingBits))
	for index, bit := range value.TrailingBits {
		tails[index] = C.bool(bit)
	}
	return buildRTCMMessage(func(out **C.SidereonRtcmMessages) uint32 {
		return uint32(C.sidereon_rtcm_build_system_parameters(C.uint16_t(value.ReferenceStationID), C.uint16_t(value.MJD), C.uint32_t(value.SecondsOfDay), C.uint8_t(value.AnnouncementCount), C.uint8_t(value.LeapSeconds), (*C.SidereonRtcmMessageAnnouncement)(unsafe.Pointer(unsafe.SliceData(announcements))), C.size_t(len(announcements)), (*C.bool)(unsafe.Pointer(unsafe.SliceData(tails))), C.size_t(len(tails)), out))
	})
}

func BuildRTCMText(value NativeRTCMText) (*RtcmMessages, error) {
	units := make([]C.uint8_t, len(value.CodeUnits))
	for index, unit := range value.CodeUnits {
		units[index] = C.uint8_t(unit)
	}
	tails := make([]C.bool, len(value.TrailingBits))
	for index, bit := range value.TrailingBits {
		tails[index] = C.bool(bit)
	}
	return buildRTCMMessage(func(out **C.SidereonRtcmMessages) uint32 {
		return uint32(C.sidereon_rtcm_build_text(C.uint16_t(value.ReferenceStationID), C.uint16_t(value.MJD), C.uint32_t(value.SecondsOfDay), C.uint8_t(value.CharacterCount), (*C.uint8_t)(unsafe.Pointer(unsafe.SliceData(units))), C.size_t(len(units)), (*C.bool)(unsafe.Pointer(unsafe.SliceData(tails))), C.size_t(len(tails)), out))
	})
}

func BuildRTCMSSRVTEC(value NativeRTCMSSRVTECInfo) (*RtcmMessages, error) {
	layers := make([]C.SidereonRtcmSsrVtecLayer, len(value.Layers))
	var cosineValues, sineValues []C.int16_t
	for index, layer := range value.Layers {
		layers[index] = C.SidereonRtcmSsrVtecLayer{height: C.uint8_t(layer.Height), degree: C.uint8_t(layer.Degree), order: C.uint8_t(layer.Order), cosine_count: C.size_t(len(layer.Cosine)), sine_count: C.size_t(len(layer.Sine))}
		for _, coefficient := range layer.Cosine {
			cosineValues = append(cosineValues, C.int16_t(coefficient))
		}
		for _, coefficient := range layer.Sine {
			sineValues = append(sineValues, C.int16_t(coefficient))
		}
	}
	tails := make([]C.bool, len(value.TrailingBits))
	for index, bit := range value.TrailingBits {
		tails[index] = C.bool(bit)
	}
	info := C.SidereonRtcmSsrVtecInfo{message_number: C.uint16_t(value.MessageNumber), has_igs_ssr_version: C.bool(value.HasIGSSSRVersion), igs_ssr_version: C.uint8_t(value.IGSSSRVersion), epoch_time_s: C.uint32_t(value.EpochTimeS), update_interval: C.uint8_t(value.UpdateInterval), multiple_message: C.bool(value.MultipleMessage), iod_ssr: C.uint8_t(value.IODSSR), provider_id: C.uint16_t(value.ProviderID), solution_id: C.uint8_t(value.SolutionID), quality_indicator: C.uint16_t(value.QualityIndicator), layer_count: C.size_t(len(layers)), trailing_bit_count: C.size_t(len(tails))}
	return buildRTCMMessage(func(out **C.SidereonRtcmMessages) uint32 {
		return uint32(C.sidereon_rtcm_build_ssr_vtec(&info, (*C.SidereonRtcmSsrVtecLayer)(unsafe.Pointer(unsafe.SliceData(layers))), C.size_t(len(layers)), (*C.int16_t)(unsafe.Pointer(unsafe.SliceData(cosineValues))), C.size_t(len(cosineValues)), (*C.int16_t)(unsafe.Pointer(unsafe.SliceData(sineValues))), C.size_t(len(sineValues)), (*C.bool)(unsafe.Pointer(unsafe.SliceData(tails))), C.size_t(len(tails)), out))
	})
}

func (m *RtcmMessages) TrailingBits(index int) ([]bool, error) {
	if index < 0 {
		return nil, errNegativeIndex
	}
	var values []bool
	err := m.resource.with(func(pointer unsafe.Pointer) error {
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return uint32(C.sidereon_rtcm_message_trailing_bits((*C.SidereonRtcmMessages)(pointer), C.size_t(index), nil, 0, &written, &required))
		}); err != nil {
			return err
		}
		count, countErr := checkedNativeCount(uint64(required))
		if countErr != nil {
			return countErr
		}
		buffer := make([]C.bool, count)
		if err := callStatus(func() uint32 {
			return uint32(C.sidereon_rtcm_message_trailing_bits((*C.SidereonRtcmMessages)(pointer), C.size_t(index), (*C.bool)(unsafe.Pointer(unsafe.SliceData(buffer))), C.size_t(len(buffer)), &written, &required))
		}); err != nil {
			return err
		}
		values = make([]bool, int(written))
		for offset, value := range buffer[:int(written)] {
			values[offset] = bool(value)
		}
		return nil
	})
	runtime.KeepAlive(m)
	return values, err
}

func (m *RtcmMessages) UnsupportedBody(index int) ([]byte, error) {
	if index < 0 {
		return nil, errNegativeIndex
	}
	var body []byte
	err := m.resource.with(func(pointer unsafe.Pointer) error {
		var bodyErr error
		body, bodyErr = copyByteOutput("RTCM unsupported body", func(output *C.uint8_t, capacity C.size_t, written, required *C.size_t) uint32 {
			return uint32(C.sidereon_rtcm_message_unsupported_body((*C.SidereonRtcmMessages)(pointer), C.size_t(index), output, capacity, written, required))
		})
		return bodyErr
	})
	runtime.KeepAlive(m)
	return body, err
}

func BuildRTCMUnsupported(messageNumber uint16, body []byte) (*RtcmMessages, error) {
	return buildRTCMMessage(func(output **C.SidereonRtcmMessages) uint32 {
		return uint32(C.sidereon_rtcm_build_unsupported(C.uint16_t(messageNumber), (*C.uint8_t)(unsafe.Pointer(unsafe.SliceData(body))), C.size_t(len(body)), output))
	})
}

func (m *RtcmMessages) WithTrailingBits(index int, bits []bool) (*RtcmMessages, error) {
	if index < 0 {
		return nil, errNegativeIndex
	}
	buffer := make([]C.bool, len(bits))
	for offset, value := range bits {
		buffer[offset] = C.bool(value)
	}
	var result *RtcmMessages
	err := m.resource.with(func(pointer unsafe.Pointer) error {
		var output *C.SidereonRtcmMessages
		if err := callStatus(func() uint32 {
			return uint32(C.sidereon_rtcm_message_with_trailing_bits((*C.SidereonRtcmMessages)(pointer), C.size_t(index), (*C.bool)(unsafe.Pointer(unsafe.SliceData(buffer))), C.size_t(len(buffer)), &output))
		}); err != nil {
			return err
		}
		created, err := newRtcmMessages(output)
		if err != nil {
			return err
		}
		result = created
		return nil
	})
	runtime.KeepAlive(m)
	return result, err
}

type NativeRTCMSSRInfoV2 struct {
	MessageNumber                                                                    uint16
	System, Kind                                                                     uint32
	Header                                                                           NativeSsrHeader
	HasIGSSSRVersion                                                                 bool
	IGSSSRVersion                                                                    uint8
	OrbitCount, ClockCount, URACount, CodeBiasCount, PhaseBiasCount, PaddingBitCount int
}

type NativeRTCMSSRMessageV2 struct {
	Info             NativeRTCMSSRInfoV2
	Orbits           []NativeSsrOrbitRecord
	Clocks           []NativeSsrClockRecord
	URA              []NativeSsrUraRecord
	CodeBiases       []NativeSsrCodeBiasRecord
	CodeBiasSignals  []NativeSsrCodeBiasSignal
	PhaseBiases      []NativeSsrPhaseBiasRecord
	PhaseBiasSignals []NativeSsrPhaseBiasSignal
	PaddingBits      []bool
}

func BuildRTCMSSRMessageV2(value NativeRTCMSSRMessageV2) (*RtcmMessages, error) {
	orbit := make([]C.SidereonRtcmSsrOrbitRecord, len(value.Orbits))
	for index, row := range value.Orbits {
		orbit[index] = C.SidereonRtcmSsrOrbitRecord{satellite_id: C.uint8_t(row.SatelliteID), iode: C.uint32_t(row.IODE), delta_radial: C.int32_t(row.DeltaRadial), delta_along: C.int32_t(row.DeltaAlong), delta_cross: C.int32_t(row.DeltaCross), dot_delta_radial: C.int32_t(row.DotDeltaRadial), dot_delta_along: C.int32_t(row.DotDeltaAlong), dot_delta_cross: C.int32_t(row.DotDeltaCross)}
	}
	clock := make([]C.SidereonRtcmSsrClockRecord, len(value.Clocks))
	for index, row := range value.Clocks {
		clock[index] = C.SidereonRtcmSsrClockRecord{satellite_id: C.uint8_t(row.SatelliteID), c0: C.int32_t(row.C0), c1: C.int32_t(row.C1), c2: C.int32_t(row.C2)}
	}
	ura := make([]C.SidereonRtcmSsrUraRecord, len(value.URA))
	for index, row := range value.URA {
		ura[index] = C.SidereonRtcmSsrUraRecord{satellite_id: C.uint8_t(row.SatelliteID), ura_index: C.uint8_t(row.URAIndex)}
	}
	codeRecords := make([]C.SidereonRtcmSsrCodeBiasRecord, len(value.CodeBiases))
	for index, row := range value.CodeBiases {
		count, err := checkedNativeSize(row.SignalCount)
		if err != nil {
			return nil, err
		}
		codeRecords[index] = C.SidereonRtcmSsrCodeBiasRecord{satellite_id: C.uint8_t(row.SatelliteID), signal_count: C.size_t(count)}
	}
	codeSignals := make([]C.SidereonRtcmSsrCodeBiasSignal, len(value.CodeBiasSignals))
	for index, row := range value.CodeBiasSignals {
		codeSignals[index] = C.SidereonRtcmSsrCodeBiasSignal{signal_id: C.uint8_t(row.SignalID), bias: C.int16_t(row.Bias)}
	}
	phaseRecords := make([]C.SidereonRtcmSsrPhaseBiasRecord, len(value.PhaseBiases))
	for index, row := range value.PhaseBiases {
		count, err := checkedNativeSize(row.SignalCount)
		if err != nil {
			return nil, err
		}
		phaseRecords[index] = C.SidereonRtcmSsrPhaseBiasRecord{satellite_id: C.uint8_t(row.SatelliteID), yaw_angle: C.uint16_t(row.YawAngle), yaw_rate: C.int8_t(row.YawRate), signal_count: C.size_t(count)}
	}
	phaseSignals := make([]C.SidereonRtcmSsrPhaseBiasSignal, len(value.PhaseBiasSignals))
	for index, row := range value.PhaseBiasSignals {
		phaseSignals[index] = C.SidereonRtcmSsrPhaseBiasSignal{signal_id: C.uint8_t(row.SignalID), integer_indicator: C.uint8_t(row.IntegerIndicator), wide_lane_integer_indicator: C.uint8_t(row.WideLaneIntegerIndicator), discontinuity_counter: C.uint8_t(row.DiscontinuityCounter), bias: C.int32_t(row.Bias)}
	}
	padding := make([]C.bool, len(value.PaddingBits))
	for index, bit := range value.PaddingBits {
		padding[index] = C.bool(bit)
	}
	source := value.Info
	header := source.Header
	info := C.SidereonRtcmSsrInfoV2{message_number: C.uint16_t(source.MessageNumber), system: C.enum_SidereonGnssSystem(source.System), kind: C.enum_SidereonRtcmSsrKind(source.Kind), has_igs_ssr_version: C.bool(source.HasIGSSSRVersion), igs_ssr_version: C.uint8_t(source.IGSSSRVersion), orbit_count: C.size_t(len(orbit)), clock_count: C.size_t(len(clock)), ura_count: C.size_t(len(ura)), code_bias_count: C.size_t(len(codeRecords)), phase_bias_count: C.size_t(len(phaseRecords)), padding_bit_count: C.size_t(len(padding))}
	info.header = C.SidereonRtcmSsrHeader{epoch_time_s: C.uint32_t(header.EpochTimeS), update_interval: C.uint8_t(header.UpdateInterval), multiple_message: C.bool(header.MultipleMessage), iod_ssr: C.uint8_t(header.IODSSR), provider_id: C.uint16_t(header.ProviderID), solution_id: C.uint8_t(header.SolutionID), has_satellite_reference_datum: C.bool(header.HasSatelliteReferenceDatum), satellite_reference_datum: C.bool(header.SatelliteReferenceDatum), has_dispersive_bias_consistency: C.bool(header.HasDispersiveBiasConsistency), dispersive_bias_consistency: C.bool(header.DispersiveBiasConsistency), has_mw_consistency: C.bool(header.HasMWConsistency), mw_consistency: C.bool(header.MWConsistency), satellite_count: C.uint8_t(header.SatelliteCount)}
	return buildRTCMMessage(func(out **C.SidereonRtcmMessages) uint32 {
		return uint32(C.sidereon_rtcm_build_ssr_v2(&info, (*C.SidereonRtcmSsrOrbitRecord)(unsafe.Pointer(unsafe.SliceData(orbit))), C.size_t(len(orbit)), (*C.SidereonRtcmSsrClockRecord)(unsafe.Pointer(unsafe.SliceData(clock))), C.size_t(len(clock)), (*C.SidereonRtcmSsrUraRecord)(unsafe.Pointer(unsafe.SliceData(ura))), C.size_t(len(ura)), (*C.SidereonRtcmSsrCodeBiasRecord)(unsafe.Pointer(unsafe.SliceData(codeRecords))), C.size_t(len(codeRecords)), (*C.SidereonRtcmSsrCodeBiasSignal)(unsafe.Pointer(unsafe.SliceData(codeSignals))), C.size_t(len(codeSignals)), (*C.SidereonRtcmSsrPhaseBiasRecord)(unsafe.Pointer(unsafe.SliceData(phaseRecords))), C.size_t(len(phaseRecords)), (*C.SidereonRtcmSsrPhaseBiasSignal)(unsafe.Pointer(unsafe.SliceData(phaseSignals))), C.size_t(len(phaseSignals)), (*C.bool)(unsafe.Pointer(unsafe.SliceData(padding))), C.size_t(len(padding)), out))
	})
}

func (m *RtcmMessages) SSRInfoV2(index int) (NativeRTCMSSRInfoV2, error) {
	if index < 0 {
		return NativeRTCMSSRInfoV2{}, errNegativeIndex
	}
	var result NativeRTCMSSRInfoV2
	var countErr error
	err := m.resource.with(func(pointer unsafe.Pointer) error {
		var value C.SidereonRtcmSsrInfoV2
		if err := callStatus(func() uint32 {
			return uint32(C.sidereon_rtcm_message_ssr_info_v2((*C.SidereonRtcmMessages)(pointer), C.size_t(index), &value))
		}); err != nil {
			return err
		}
		result = NativeRTCMSSRInfoV2{MessageNumber: uint16(value.message_number), System: uint32(value.system), Kind: uint32(value.kind), HasIGSSSRVersion: bool(value.has_igs_ssr_version), IGSSSRVersion: uint8(value.igs_ssr_version)}
		result.OrbitCount, countErr = checkedNativeCount(uint64(value.orbit_count))
		if countErr != nil {
			return countErr
		}
		result.ClockCount, countErr = checkedNativeCount(uint64(value.clock_count))
		if countErr != nil {
			return countErr
		}
		result.URACount, countErr = checkedNativeCount(uint64(value.ura_count))
		if countErr != nil {
			return countErr
		}
		result.CodeBiasCount, countErr = checkedNativeCount(uint64(value.code_bias_count))
		if countErr != nil {
			return countErr
		}
		result.PhaseBiasCount, countErr = checkedNativeCount(uint64(value.phase_bias_count))
		if countErr != nil {
			return countErr
		}
		result.PaddingBitCount, countErr = checkedNativeCount(uint64(value.padding_bit_count))
		if countErr != nil {
			return countErr
		}
		result.Header = NativeSsrHeader{EpochTimeS: uint32(value.header.epoch_time_s), UpdateInterval: uint8(value.header.update_interval), MultipleMessage: bool(value.header.multiple_message), IODSSR: uint8(value.header.iod_ssr), ProviderID: uint16(value.header.provider_id), SolutionID: uint8(value.header.solution_id), HasSatelliteReferenceDatum: bool(value.header.has_satellite_reference_datum), SatelliteReferenceDatum: bool(value.header.satellite_reference_datum), HasDispersiveBiasConsistency: bool(value.header.has_dispersive_bias_consistency), DispersiveBiasConsistency: bool(value.header.dispersive_bias_consistency), HasMWConsistency: bool(value.header.has_mw_consistency), MWConsistency: bool(value.header.mw_consistency), SatelliteCount: uint8(value.header.satellite_count)}
		return nil
	})
	runtime.KeepAlive(m)
	return result, err
}
