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

type OptionalNumber struct {
	Value   float64
	Present bool
}

type CdmNumbers struct {
	MissDistanceM        OptionalNumber
	RelativeSpeedMPerS   OptionalNumber
	CollisionProbability OptionalNumber
	HardBodyRadiusM      OptionalNumber
}

type CdmObject struct {
	PositionKm              [3]float64
	VelocityKmPerS          [3]float64
	CovarianceRTN           [6]float64
	VelocityCovarianceRTN   [15]float64
	HasVelocityCovariance   bool
	ObjectDesignator        string
	CatalogName             string
	ObjectName              string
	InternationalDesignator string
	ObjectType              string
	RefFrame                string
}

type Cdm struct{ handle *positioningHandle }

const (
	CDMStringFieldCreationDateValue               = uint32(C.SIDEREON_CDM_STRING_FIELD_CREATION_DATE)
	CDMStringFieldOriginatorValue                 = uint32(C.SIDEREON_CDM_STRING_FIELD_ORIGINATOR)
	CDMStringFieldMessageIDValue                  = uint32(C.SIDEREON_CDM_STRING_FIELD_MESSAGE_ID)
	CDMStringFieldTCAValue                        = uint32(C.SIDEREON_CDM_STRING_FIELD_TCA)
	CDMStringFieldCollisionProbabilityMethodValue = uint32(C.SIDEREON_CDM_STRING_FIELD_COLLISION_PROBABILITY_METHOD)
	CDMStringFieldObject1DesignatorValue          = uint32(C.SIDEREON_CDM_STRING_FIELD_OBJECT1_DESIGNATOR)
	CDMStringFieldObject1NameValue                = uint32(C.SIDEREON_CDM_STRING_FIELD_OBJECT1_NAME)
	CDMStringFieldObject2DesignatorValue          = uint32(C.SIDEREON_CDM_STRING_FIELD_OBJECT2_DESIGNATOR)
	CDMStringFieldObject2NameValue                = uint32(C.SIDEREON_CDM_STRING_FIELD_OBJECT2_NAME)

	CDMObjectStringFieldObjectDesignatorValue        = uint32(C.SIDEREON_CDM_OBJECT_STRING_FIELD_OBJECT_DESIGNATOR)
	CDMObjectStringFieldCatalogNameValue             = uint32(C.SIDEREON_CDM_OBJECT_STRING_FIELD_CATALOG_NAME)
	CDMObjectStringFieldObjectNameValue              = uint32(C.SIDEREON_CDM_OBJECT_STRING_FIELD_OBJECT_NAME)
	CDMObjectStringFieldInternationalDesignatorValue = uint32(C.SIDEREON_CDM_OBJECT_STRING_FIELD_INTERNATIONAL_DESIGNATOR)
	CDMObjectStringFieldObjectTypeValue              = uint32(C.SIDEREON_CDM_OBJECT_STRING_FIELD_OBJECT_TYPE)
	CDMObjectStringFieldOperatorContactPositionValue = uint32(C.SIDEREON_CDM_OBJECT_STRING_FIELD_OPERATOR_CONTACT_POSITION)
	CDMObjectStringFieldOperatorOrganizationValue    = uint32(C.SIDEREON_CDM_OBJECT_STRING_FIELD_OPERATOR_ORGANIZATION)
	CDMObjectStringFieldOperatorPhoneValue           = uint32(C.SIDEREON_CDM_OBJECT_STRING_FIELD_OPERATOR_PHONE)
	CDMObjectStringFieldOperatorEmailValue           = uint32(C.SIDEREON_CDM_OBJECT_STRING_FIELD_OPERATOR_EMAIL)
	CDMObjectStringFieldEphemerisNameValue           = uint32(C.SIDEREON_CDM_OBJECT_STRING_FIELD_EPHEMERIS_NAME)
	CDMObjectStringFieldCovarianceMethodValue        = uint32(C.SIDEREON_CDM_OBJECT_STRING_FIELD_COVARIANCE_METHOD)
	CDMObjectStringFieldManeuverableValue            = uint32(C.SIDEREON_CDM_OBJECT_STRING_FIELD_MANEUVERABLE)
	CDMObjectStringFieldOrbitCenterValue             = uint32(C.SIDEREON_CDM_OBJECT_STRING_FIELD_ORBIT_CENTER)
	CDMObjectStringFieldRefFrameValue                = uint32(C.SIDEREON_CDM_OBJECT_STRING_FIELD_REF_FRAME)
	CDMObjectStringFieldGravityModelValue            = uint32(C.SIDEREON_CDM_OBJECT_STRING_FIELD_GRAVITY_MODEL)
	CDMObjectStringFieldAtmosphericModelValue        = uint32(C.SIDEREON_CDM_OBJECT_STRING_FIELD_ATMOSPHERIC_MODEL)
	CDMObjectStringFieldNBodyPerturbationsValue      = uint32(C.SIDEREON_CDM_OBJECT_STRING_FIELD_N_BODY_PERTURBATIONS)
	CDMObjectStringFieldSolarRadPressureValue        = uint32(C.SIDEREON_CDM_OBJECT_STRING_FIELD_SOLAR_RAD_PRESSURE)
	CDMObjectStringFieldEarthTidesValue              = uint32(C.SIDEREON_CDM_OBJECT_STRING_FIELD_EARTH_TIDES)
	CDMObjectStringFieldIntrackThrustValue           = uint32(C.SIDEREON_CDM_OBJECT_STRING_FIELD_INTRACK_THRUST)
)

func releaseCdm(pointer unsafe.Pointer) {
	C.sidereon_cdm_free((*C.SidereonCdm)(pointer))
}

func parseCdm(data []byte, xml bool) (*Cdm, error) {
	for _, value := range data {
		if value == 0 {
			return nil, invalidArgument("CDM text contains an embedded NUL byte")
		}
	}
	var out *C.SidereonCdm
	err := withCBytes(data, func(pointer *C.uint8_t, length C.size_t) uint32 {
		if xml {
			return C.sidereon_cdm_parse_xml(pointer, length, &out)
		}
		return C.sidereon_cdm_parse_kvn(pointer, length, &out)
	})
	if err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_cdm_free(out) })
		}
		return nil, err
	}
	if out == nil {
		return nil, missingNativeHandle("CDM parse")
	}
	return &Cdm{handle: newPositioningHandle(unsafe.Pointer(out), releaseCdm)}, nil
}

func ParseCdmKVN(data []byte) (*Cdm, error) {
	return parseCdm(data, false)
}

func ParseCdmXML(data []byte) (*Cdm, error) {
	return parseCdm(data, true)
}

func withCBytes(data []byte, fn func(*C.uint8_t, C.size_t) uint32) error {
	var err error
	withCThread(func() {
		var pointer unsafe.Pointer
		if len(data) > 0 {
			pointer = C.CBytes(data)
			if pointer == nil {
				err = invalidArgument("unable to allocate native input buffer")
				return
			}
			defer C.free(pointer)
		}
		err = statusErrorLocked(fn((*C.uint8_t)(pointer), C.size_t(len(data))))
	})
	return err
}

// Close releases the native CDM handle. It is idempotent.
func (c *Cdm) Close() error {
	if c == nil {
		return nil
	}
	return c.handle.close()
}

func isNaN(value float64) bool {
	return value != value
}

// Numbers copies top-level numeric CDM fields and preserves their optional
// presence. An absent C value is represented by NaN and Present=false.
func (c *Cdm) Numbers() (CdmNumbers, error) {
	var values CdmNumbers
	if c == nil || c.handle == nil {
		return values, ErrClosed
	}
	var miss, speed, pc, radius C.double
	err := c.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_cdm_numbers((*C.SidereonCdm)(pointer), &miss, &speed, &pc, &radius)
		})
	})
	runtime.KeepAlive(c)
	if err != nil {
		return values, err
	}
	values.MissDistanceM = OptionalNumber{Value: float64(miss), Present: !isNaN(float64(miss))}
	values.RelativeSpeedMPerS = OptionalNumber{Value: float64(speed), Present: !isNaN(float64(speed))}
	values.CollisionProbability = OptionalNumber{Value: float64(pc), Present: !isNaN(float64(pc))}
	values.HardBodyRadiusM = OptionalNumber{Value: float64(radius), Present: !isNaN(float64(radius))}
	return values, nil
}

func copyVariable(call func(*C.uint8_t, C.size_t, *C.size_t, *C.size_t) uint32) ([]byte, error) {
	var written, required C.size_t
	if err := callStatus(func() uint32 { return call(nil, 0, &written, &required) }); err != nil {
		return nil, err
	}
	expected, err := validateNativeQuery("CDM variable output", uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(expected, unsafe.Sizeof(C.uint8_t(0))); err != nil {
		return nil, err
	}
	buffer := make([]C.uint8_t, expected)
	var output *C.uint8_t
	if len(buffer) > 0 {
		output = &buffer[0]
	}
	written, required = 0, 0
	if err := callStatus(func() uint32 { return call(output, C.size_t(len(buffer)), &written, &required) }); err != nil {
		return nil, err
	}
	count, err := validateTwoPassCounts("CDM variable output", len(buffer), expected, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	value := make([]byte, count)
	for i := range value {
		value[i] = byte(buffer[i])
	}
	return value, nil
}

func cCDMStringField(field uint32) (C.uint32_t, error) {
	switch field {
	case CDMStringFieldCreationDateValue,
		CDMStringFieldOriginatorValue,
		CDMStringFieldMessageIDValue,
		CDMStringFieldTCAValue,
		CDMStringFieldCollisionProbabilityMethodValue,
		CDMStringFieldObject1DesignatorValue,
		CDMStringFieldObject1NameValue,
		CDMStringFieldObject2DesignatorValue,
		CDMStringFieldObject2NameValue:
		return C.uint32_t(field), nil
	default:
		return 0, invalidArgument("invalid CDM string-field selector")
	}
}

func cCDMObjectStringField(field uint32) (C.uint32_t, error) {
	switch field {
	case CDMObjectStringFieldObjectDesignatorValue,
		CDMObjectStringFieldCatalogNameValue,
		CDMObjectStringFieldObjectNameValue,
		CDMObjectStringFieldInternationalDesignatorValue,
		CDMObjectStringFieldObjectTypeValue,
		CDMObjectStringFieldOperatorContactPositionValue,
		CDMObjectStringFieldOperatorOrganizationValue,
		CDMObjectStringFieldOperatorPhoneValue,
		CDMObjectStringFieldOperatorEmailValue,
		CDMObjectStringFieldEphemerisNameValue,
		CDMObjectStringFieldCovarianceMethodValue,
		CDMObjectStringFieldManeuverableValue,
		CDMObjectStringFieldOrbitCenterValue,
		CDMObjectStringFieldRefFrameValue,
		CDMObjectStringFieldGravityModelValue,
		CDMObjectStringFieldAtmosphericModelValue,
		CDMObjectStringFieldNBodyPerturbationsValue,
		CDMObjectStringFieldSolarRadPressureValue,
		CDMObjectStringFieldEarthTidesValue,
		CDMObjectStringFieldIntrackThrustValue:
		return C.uint32_t(field), nil
	default:
		return 0, invalidArgument("invalid CDM object string-field selector")
	}
}

func (c *Cdm) StringField(field uint32) (string, bool, error) {
	if c == nil || c.handle == nil {
		return "", false, ErrClosed
	}
	cField, err := cCDMStringField(field)
	if err != nil {
		return "", false, err
	}
	var value []byte
	err = c.handle.with(func(pointer unsafe.Pointer) error {
		value, err = copyVariable(func(out *C.uint8_t, length C.size_t, written, required *C.size_t) uint32 {
			return C.sidereon_cdm_string_field((*C.SidereonCdm)(pointer), cField, out, length, written, required)
		})
		return err
	})
	runtime.KeepAlive(c)
	return string(value), len(value) != 0, err
}

func (c *Cdm) objectStringField(pointer unsafe.Pointer, index int, field C.uint32_t) (string, bool, error) {
	value, err := copyVariable(func(out *C.uint8_t, length C.size_t, written, required *C.size_t) uint32 {
		return C.sidereon_cdm_object_string_field((*C.SidereonCdm)(pointer), C.uint32_t(index), field, out, length, written, required)
	})
	return string(value), len(value) != 0, err
}

// ObjectStringField copies a selected per-object string. objectIndex is 1 or
// 2. The boolean reports nonempty text; the C selector has no separate
// presence bit for these fields.
func (c *Cdm) ObjectStringField(objectIndex int, field uint32) (string, bool, error) {
	if c == nil || c.handle == nil {
		return "", false, ErrClosed
	}
	if objectIndex != 1 && objectIndex != 2 {
		return "", false, invalidArgument("CDM object index must be 1 or 2")
	}
	cField, err := cCDMObjectStringField(field)
	if err != nil {
		return "", false, err
	}
	var value string
	var present bool
	err = c.handle.with(func(pointer unsafe.Pointer) error {
		value, present, err = c.objectStringField(pointer, objectIndex, cField)
		return err
	})
	runtime.KeepAlive(c)
	return value, present, err
}

// Object copies state, covariance, and the fixed metadata subset for object 1
// or 2. Optional velocity covariance is indicated by HasVelocityCovariance.
func (c *Cdm) Object(index int) (CdmObject, error) {
	var out CdmObject
	if c == nil || c.handle == nil {
		return out, ErrClosed
	}
	if index != 1 && index != 2 {
		return out, invalidArgument("CDM object index must be 1 or 2")
	}
	var position, velocity [3]C.double
	var covariance [6]C.double
	var velocityCovariance [15]C.double
	var hasVelocityCovariance C.bool
	err := c.handle.with(func(pointer unsafe.Pointer) error {
		if err := callStatus(func() uint32 {
			return C.sidereon_cdm_object_state((*C.SidereonCdm)(pointer), C.uint32_t(index), &position[0], &velocity[0], &covariance[0])
		}); err != nil {
			return err
		}
		if err := callStatus(func() uint32 {
			return C.sidereon_cdm_object_velocity_covariance((*C.SidereonCdm)(pointer), C.uint32_t(index), &velocityCovariance[0], &hasVelocityCovariance)
		}); err != nil {
			return err
		}
		fields := []struct {
			field       C.uint32_t
			destination *string
		}{
			{C.SIDEREON_CDM_OBJECT_STRING_FIELD_OBJECT_DESIGNATOR, &out.ObjectDesignator},
			{C.SIDEREON_CDM_OBJECT_STRING_FIELD_CATALOG_NAME, &out.CatalogName},
			{C.SIDEREON_CDM_OBJECT_STRING_FIELD_OBJECT_NAME, &out.ObjectName},
			{C.SIDEREON_CDM_OBJECT_STRING_FIELD_INTERNATIONAL_DESIGNATOR, &out.InternationalDesignator},
			{C.SIDEREON_CDM_OBJECT_STRING_FIELD_OBJECT_TYPE, &out.ObjectType},
			{C.SIDEREON_CDM_OBJECT_STRING_FIELD_REF_FRAME, &out.RefFrame},
		}
		for _, item := range fields {
			value, _, err := c.objectStringField(pointer, index, item.field)
			if err != nil {
				return err
			}
			*item.destination = value
		}
		return nil
	})
	runtime.KeepAlive(c)
	if err != nil {
		return out, err
	}
	for i := 0; i < 3; i++ {
		out.PositionKm[i] = float64(position[i])
		out.VelocityKmPerS[i] = float64(velocity[i])
	}
	for i := 0; i < 6; i++ {
		out.CovarianceRTN[i] = float64(covariance[i])
	}
	for i := 0; i < 15; i++ {
		out.VelocityCovarianceRTN[i] = float64(velocityCovariance[i])
	}
	out.HasVelocityCovariance = bool(hasVelocityCovariance)
	return out, nil
}

// ToKVN serializes the CDM through C into Go-owned bytes.
func (c *Cdm) ToKVN() ([]byte, error) {
	if c == nil || c.handle == nil {
		return nil, ErrClosed
	}
	var value []byte
	var copyErr error
	err := c.handle.with(func(pointer unsafe.Pointer) error {
		value, copyErr = copyVariable(func(out *C.uint8_t, length C.size_t, written, required *C.size_t) uint32 {
			return C.sidereon_cdm_to_kvn((*C.SidereonCdm)(pointer), out, length, written, required)
		})
		return copyErr
	})
	runtime.KeepAlive(c)
	return value, err
}

// ToXML serializes the CDM through C into Go-owned bytes.
func (c *Cdm) ToXML() ([]byte, error) {
	if c == nil || c.handle == nil {
		return nil, ErrClosed
	}
	var value []byte
	var copyErr error
	err := c.handle.with(func(pointer unsafe.Pointer) error {
		value, copyErr = copyVariable(func(out *C.uint8_t, length C.size_t, written, required *C.size_t) uint32 {
			return C.sidereon_cdm_to_xml((*C.SidereonCdm)(pointer), out, length, written, required)
		})
		return copyErr
	})
	runtime.KeepAlive(c)
	return value, err
}
