package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// OptionalNumber preserves the C CDM distinction between an absent numeric
// field and a present zero. Value is NaN when Present is false.
type OptionalNumber struct {
	// Value is the value stored in this record.
	Value float64
	// Present reports whether the present field is present.
	Present bool
}

// CDMNumbers contains the optional top-level CDM numeric scalars. Distances
// are meters, speed is meters/second, and probabilities are dimensionless.
type CDMNumbers struct {
	// MissDistanceM is the miss distance m in metres.
	MissDistanceM OptionalNumber
	// RelativeSpeedMPerS is the relative speed m per s in metres per second.
	RelativeSpeedMPerS OptionalNumber
	// CollisionProbability is the collision probability.
	CollisionProbability OptionalNumber
	// HardBodyRadiusM is the hard body radius m in metres.
	HardBodyRadiusM OptionalNumber
}

// CDMObject is the copied state and selected metadata for one CDM object.
// Position is km, velocity is km/s, and CovarianceRTN is the six-value lower
// triangle (CR_R, CT_R, CT_T, CN_R, CN_T, CN_N) in m^2. The optional velocity
// covariance is the 15-value lower triangle in the C header's order and is in
// m^2/s and m^2/s^2 as specified by the CDM.
type CDMObject struct {
	// PositionKm is the position km in kilometres.
	PositionKm [3]float64
	// VelocityKmPerS is the velocity km per s in kilometres per second.
	VelocityKmPerS [3]float64
	// CovarianceRTN is the six-element lower-triangular position covariance in RTN order (CR_R, CT_R, CT_T, CN_R, CN_T, CN_N), in square metres.
	CovarianceRTN [6]float64
	// VelocityCovarianceRTN is the fixed 15-element lower triangle in the C header's RTN order; entries use the CDM's mixed square-metre-per-second and square-metre-per-second-squared units.
	VelocityCovarianceRTN [15]float64
	// HasVelocityCovariance reports whether the has velocity covariance field is present.
	HasVelocityCovariance bool
	// ObjectDesignator is the CDM-designated identifier for this object.
	ObjectDesignator string
	// CatalogName is the catalog or catalog-source name for this object.
	CatalogName string
	// ObjectName is the human-readable name for this object.
	ObjectName string
	// InternationalDesignator is the international designator assigned to this object.
	InternationalDesignator string
	// ObjectType is the provider's classification of this object.
	ObjectType string
	// RefFrame identifies the reference frame used by this CDM object.
	RefFrame string
}

// CDMStringField selects one of the C top-level CDM string readers.
type CDMStringField uint32

const (
	// CDMCreationDate identifies the cdm creation date case.
	CDMCreationDate CDMStringField = CDMStringField(native.CDMStringFieldCreationDateValue)
	// CDMOriginator identifies the cdm originator case.
	CDMOriginator CDMStringField = CDMStringField(native.CDMStringFieldOriginatorValue)
	// CDMMessageID identifies the cdm message id case.
	CDMMessageID CDMStringField = CDMStringField(native.CDMStringFieldMessageIDValue)
	// CDMTCA identifies the cdmtca case.
	CDMTCA CDMStringField = CDMStringField(native.CDMStringFieldTCAValue)
	// CDMCollisionProbabilityMethod identifies the cdm collision probability method case.
	CDMCollisionProbabilityMethod CDMStringField = CDMStringField(native.CDMStringFieldCollisionProbabilityMethodValue)
	// CDMObject1Designator identifies the cdm object1 designator case.
	CDMObject1Designator CDMStringField = CDMStringField(native.CDMStringFieldObject1DesignatorValue)
	// CDMObject1Name identifies the cdm object1 name case.
	CDMObject1Name CDMStringField = CDMStringField(native.CDMStringFieldObject1NameValue)
	// CDMObject2Designator identifies the cdm object2 designator case.
	CDMObject2Designator CDMStringField = CDMStringField(native.CDMStringFieldObject2DesignatorValue)
	// CDMObject2Name identifies the cdm object2 name case.
	CDMObject2Name CDMStringField = CDMStringField(native.CDMStringFieldObject2NameValue)
)

// CDMObjectStringField selects a per-object metadata string.
type CDMObjectStringField uint32

const (
	// CDMObjectDesignator identifies the cdm object designator case.
	CDMObjectDesignator CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldObjectDesignatorValue)
	// CDMCatalogName identifies the cdm catalog name case.
	CDMCatalogName CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldCatalogNameValue)
	// CDMObjectName identifies the cdm object name case.
	CDMObjectName CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldObjectNameValue)
	// CDMInternationalDesignator identifies the cdm international designator case.
	CDMInternationalDesignator CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldInternationalDesignatorValue)
	// CDMObjectType identifies the cdm object type case.
	CDMObjectType CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldObjectTypeValue)
	// CDMOperatorContactPosition identifies the cdm operator contact position case.
	CDMOperatorContactPosition CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldOperatorContactPositionValue)
	// CDMOperatorOrganization identifies the cdm operator organization case.
	CDMOperatorOrganization CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldOperatorOrganizationValue)
	// CDMOperatorPhone identifies the cdm operator phone case.
	CDMOperatorPhone CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldOperatorPhoneValue)
	// CDMOperatorEmail identifies the cdm operator email case.
	CDMOperatorEmail CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldOperatorEmailValue)
	// CDMEphemerisName identifies the cdm ephemeris name case.
	CDMEphemerisName CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldEphemerisNameValue)
	// CDMCovarianceMethod identifies the cdm covariance method case.
	CDMCovarianceMethod CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldCovarianceMethodValue)
	// CDMManeuverable identifies the cdm maneuverable case.
	CDMManeuverable CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldManeuverableValue)
	// CDMOrbitCenter identifies the cdm orbit center case.
	CDMOrbitCenter CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldOrbitCenterValue)
	// CDMRefFrame identifies the cdm ref frame case.
	CDMRefFrame CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldRefFrameValue)
	// CDMGravityModel identifies the cdm gravity model case.
	CDMGravityModel CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldGravityModelValue)
	// CDMAtmosphericModel identifies the cdm atmospheric model case.
	CDMAtmosphericModel CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldAtmosphericModelValue)
	// CDMNBodyPerturbations identifies the cdmn body perturbations case.
	CDMNBodyPerturbations CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldNBodyPerturbationsValue)
	// CDMSolarRadiationPressure identifies the cdm solar radiation pressure case.
	CDMSolarRadiationPressure CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldSolarRadPressureValue)
	// CDMEarthTides identifies the cdm earth tides case.
	CDMEarthTides CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldEarthTidesValue)
	// CDMIntrackThrust identifies the cdm intrack thrust case.
	CDMIntrackThrust CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldIntrackThrustValue)
)

// CDM owns a parsed CCSDS Conjunction Data Message. Its C handle is
// non-copyable. Read-only methods may run concurrently with one another and
// with Close. Close waits for active calls, clears the native resource, and is
// idempotent; methods called after Close return ErrClosed.
type CDM struct {
	_      noCopy
	handle *native.Cdm
}

// ParseCDMKVN parses a public CCSDS CDM KVN message through C.
func ParseCDMKVN(data []byte) (*CDM, error) {
	value, err := native.ParseCdmKVN(data)
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errNilNativeHandle
	}
	return &CDM{handle: value}, nil
}

// ParseCDMXML parses a public CCSDS CDM XML message through C.
func ParseCDMXML(data []byte) (*CDM, error) {
	value, err := native.ParseCdmXML(data)
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errNilNativeHandle
	}
	return &CDM{handle: value}, nil
}

// Close releases the native CDM handle. It is idempotent.
func (c *CDM) Close() error {
	if c != nil && c.handle != nil {
		return publicError(c.handle.Close())
	}
	return nil
}

// Numbers copies top-level CDM numeric fields, retaining presence information.
func (c *CDM) Numbers() (CDMNumbers, error) {
	if c == nil || c.handle == nil {
		return CDMNumbers{}, ErrClosed
	}
	value, err := c.handle.Numbers()
	return CDMNumbers{MissDistanceM: publicOptional(value.MissDistanceM), RelativeSpeedMPerS: publicOptional(value.RelativeSpeedMPerS), CollisionProbability: publicOptional(value.CollisionProbability), HardBodyRadiusM: publicOptional(value.HardBodyRadiusM)}, publicError(err)
}

func publicOptional(value native.OptionalNumber) OptionalNumber {
	return OptionalNumber{Value: value.Value, Present: value.Present}
}

// StringField copies a selected top-level field. The boolean reports whether
// the returned text is nonempty; the C selector has no separate presence bit.
func (c *CDM) StringField(field CDMStringField) (string, bool, error) {
	if c == nil || c.handle == nil {
		return "", false, ErrClosed
	}
	value, present, err := c.handle.StringField(uint32(field))
	return value, present, publicError(err)
}

// ObjectStringField copies a selected per-object field. The boolean reports
// whether the returned text is nonempty; the C selector has no separate
// presence bit.
// objectIndex must be 1 or 2.
func (c *CDM) ObjectStringField(objectIndex int, field CDMObjectStringField) (string, bool, error) {
	if c == nil || c.handle == nil {
		return "", false, ErrClosed
	}
	value, present, err := c.handle.ObjectStringField(objectIndex, uint32(field))
	return value, present, publicError(err)
}

// Object copies state, covariance, and a useful fixed metadata subset for
// objectIndex 1 or 2. Optional velocity covariance is indicated by
// HasVelocityCovariance.
func (c *CDM) Object(objectIndex int) (CDMObject, error) {
	if c == nil || c.handle == nil {
		return CDMObject{}, ErrClosed
	}
	value, err := c.handle.Object(objectIndex)
	return CDMObject{PositionKm: value.PositionKm, VelocityKmPerS: value.VelocityKmPerS, CovarianceRTN: value.CovarianceRTN, VelocityCovarianceRTN: value.VelocityCovarianceRTN, HasVelocityCovariance: value.HasVelocityCovariance, ObjectDesignator: value.ObjectDesignator, CatalogName: value.CatalogName, ObjectName: value.ObjectName, InternationalDesignator: value.InternationalDesignator, ObjectType: value.ObjectType, RefFrame: value.RefFrame}, publicError(err)
}

// ToKVN serializes the CDM through C and returns an independent byte slice.
func (c *CDM) ToKVN() ([]byte, error) {
	if c == nil || c.handle == nil {
		return nil, ErrClosed
	}
	value, err := c.handle.ToKVN()
	return value, publicError(err)
}

// ToXML serializes the CDM through C and returns an independent byte slice.
func (c *CDM) ToXML() ([]byte, error) {
	if c == nil || c.handle == nil {
		return nil, ErrClosed
	}
	value, err := c.handle.ToXML()
	return value, publicError(err)
}
