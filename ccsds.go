package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// OptionalNumber preserves the C CDM distinction between an absent numeric
// field and a present zero. Value is NaN when Present is false.
type OptionalNumber struct {
	Value   float64
	Present bool
}

// CDMNumbers contains the optional top-level CDM numeric scalars. Distances
// are meters, speed is meters/second, and probabilities are dimensionless.
type CDMNumbers struct {
	MissDistanceM        OptionalNumber
	RelativeSpeedMPerS   OptionalNumber
	CollisionProbability OptionalNumber
	HardBodyRadiusM      OptionalNumber
}

// CDMObject is the copied state and selected metadata for one CDM object.
// Position is km, velocity is km/s, and CovarianceRTN is the six-value lower
// triangle (CR_R, CT_R, CT_T, CN_R, CN_T, CN_N) in m^2. The optional velocity
// covariance is the 15-value lower triangle in the C header's order and is in
// m^2/s and m^2/s^2 as specified by the CDM.
type CDMObject struct {
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

// CDMStringField selects one of the C top-level CDM string readers.
type CDMStringField uint32

const (
	CDMCreationDate               CDMStringField = CDMStringField(native.CDMStringFieldCreationDateValue)
	CDMOriginator                 CDMStringField = CDMStringField(native.CDMStringFieldOriginatorValue)
	CDMMessageID                  CDMStringField = CDMStringField(native.CDMStringFieldMessageIDValue)
	CDMTCA                        CDMStringField = CDMStringField(native.CDMStringFieldTCAValue)
	CDMCollisionProbabilityMethod CDMStringField = CDMStringField(native.CDMStringFieldCollisionProbabilityMethodValue)
	CDMObject1Designator          CDMStringField = CDMStringField(native.CDMStringFieldObject1DesignatorValue)
	CDMObject1Name                CDMStringField = CDMStringField(native.CDMStringFieldObject1NameValue)
	CDMObject2Designator          CDMStringField = CDMStringField(native.CDMStringFieldObject2DesignatorValue)
	CDMObject2Name                CDMStringField = CDMStringField(native.CDMStringFieldObject2NameValue)
)

// CDMObjectStringField selects a per-object metadata string.
type CDMObjectStringField uint32

const (
	CDMObjectDesignator        CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldObjectDesignatorValue)
	CDMCatalogName             CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldCatalogNameValue)
	CDMObjectName              CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldObjectNameValue)
	CDMInternationalDesignator CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldInternationalDesignatorValue)
	CDMObjectType              CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldObjectTypeValue)
	CDMOperatorContactPosition CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldOperatorContactPositionValue)
	CDMOperatorOrganization    CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldOperatorOrganizationValue)
	CDMOperatorPhone           CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldOperatorPhoneValue)
	CDMOperatorEmail           CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldOperatorEmailValue)
	CDMEphemerisName           CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldEphemerisNameValue)
	CDMCovarianceMethod        CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldCovarianceMethodValue)
	CDMManeuverable            CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldManeuverableValue)
	CDMOrbitCenter             CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldOrbitCenterValue)
	CDMRefFrame                CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldRefFrameValue)
	CDMGravityModel            CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldGravityModelValue)
	CDMAtmosphericModel        CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldAtmosphericModelValue)
	CDMNBodyPerturbations      CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldNBodyPerturbationsValue)
	CDMSolarRadiationPressure  CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldSolarRadPressureValue)
	CDMEarthTides              CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldEarthTidesValue)
	CDMIntrackThrust           CDMObjectStringField = CDMObjectStringField(native.CDMObjectStringFieldIntrackThrustValue)
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
