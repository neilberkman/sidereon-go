package sidereon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"

	"sidereon.dev/go/v3/internal/native"
)

var ommSnapshotKeys = []string{"ccsds_omm_vers", "classification", "creation_date", "originator", "message_id", "object_name", "object_id", "center_name", "ref_frame", "ref_frame_epoch", "time_system", "mean_element_theory", "epoch", "mean_motion", "semi_major_axis_km", "eccentricity", "inclination_deg", "ra_of_asc_node_deg", "arg_of_pericenter_deg", "mean_anomaly_deg", "gm_km3_s2", "spacecraft", "ephemeris_type", "classification_type", "norad_cat_id", "element_set_no", "rev_at_epoch", "bstar", "bterm_m2_kg", "mean_motion_dot", "mean_motion_ddot", "agom_m2_kg", "covariance", "user_defined", "comments", "exact_sgp4_epoch", "quantize_tle_derived_fields"}

// OMMData is a detached copy of all 37 fields stored by the native OMM value.
// It retains block-specific comments and the in-memory SGP4 side channels that
// are not part of the three CCSDS wire encodings.
type OMMData struct {
	// CCSDSOMMVERS is the stated CCSDS OMM version, when present.
	CCSDSOMMVERS *string `json:"ccsds_omm_vers"`
	// Classification is the optional message-level classification text.
	Classification *string `json:"classification"`
	// CreationDate is the validated creation timestamp, when present.
	CreationDate *string `json:"creation_date"`
	// Originator is the stated message originator, when present.
	Originator *string `json:"originator"`
	// MessageID is the optional message identifier.
	MessageID *string `json:"message_id"`
	// ObjectName is the optional object name.
	ObjectName *string `json:"object_name"`
	// ObjectID is the international designator, when present.
	ObjectID *string `json:"object_id"`
	// CenterName is the optional center metadata.
	CenterName *string `json:"center_name"`
	// RefFrame is the optional reference frame metadata.
	RefFrame *string `json:"ref_frame"`
	// RefFrameEpoch is the optional reference frame epoch text.
	RefFrameEpoch *string `json:"ref_frame_epoch"`
	// TimeSystem is the optional time-system label.
	TimeSystem *string `json:"time_system"`
	// MeanElementTheory is the stated mean-element theory.
	MeanElementTheory *string `json:"mean_element_theory"`
	// Epoch is the parsed civil epoch and retains its microsecond/femtosecond parts.
	Epoch OMMEpoch `json:"epoch"`
	// MeanMotion is the optional mean motion in revolutions per day.
	MeanMotion *float64 `json:"mean_motion"`
	// SemiMajorAxisKM is the optional semi-major axis in kilometres.
	SemiMajorAxisKM *float64 `json:"semi_major_axis_km"`
	// Eccentricity is the dimensionless orbital eccentricity.
	Eccentricity float64 `json:"eccentricity"`
	// InclinationDeg is the inclination in degrees.
	InclinationDeg float64 `json:"inclination_deg"`
	// RAOfAscNodeDeg is the right ascension of the ascending node in degrees.
	RAOfAscNodeDeg float64 `json:"ra_of_asc_node_deg"`
	// ArgOfPericenterDeg is the argument of pericenter in degrees.
	ArgOfPericenterDeg float64 `json:"arg_of_pericenter_deg"`
	// MeanAnomalyDeg is the mean anomaly in degrees.
	MeanAnomalyDeg float64 `json:"mean_anomaly_deg"`
	// GMKM3S2 is the optional gravitational parameter in km³/s².
	GMKM3S2 *float64 `json:"gm_km3_s2"`
	// Spacecraft contains the optional spacecraft-parameter block.
	Spacecraft *OMMSpacecraft `json:"spacecraft"`
	// EphemerisType is the optional stated ephemeris type.
	EphemerisType *int32 `json:"ephemeris_type"`
	// ClassificationType is the optional TLE classification character.
	ClassificationType *string `json:"classification_type"`
	// NORADCatID is the optional NORAD catalog identifier.
	NORADCatID *uint32 `json:"norad_cat_id"`
	// ElementSetNo is the optional element-set number.
	ElementSetNo *int32 `json:"element_set_no"`
	// RevAtEpoch is the optional revolution count at epoch.
	RevAtEpoch *int64 `json:"rev_at_epoch"`
	// BStar is the optional SGP4 drag term in inverse Earth radii.
	BStar *float64 `json:"bstar"`
	// BTermM2KG is the optional ballistic coefficient in m²/kg.
	BTermM2KG *float64 `json:"bterm_m2_kg"`
	// MeanMotionDot is the optional first derivative in rev/day².
	MeanMotionDot *float64 `json:"mean_motion_dot"`
	// MeanMotionDDot is the optional second derivative in rev/day³.
	MeanMotionDDot *float64 `json:"mean_motion_ddot"`
	// AGOMM2KG is the optional solar-radiation coefficient in m²/kg.
	AGOMM2KG *float64 `json:"agom_m2_kg"`
	// Covariance is the optional position/velocity covariance block.
	Covariance *OMMCovariance `json:"covariance"`
	// UserDefined contains user parameters in source order.
	UserDefined []OMMUserDefined `json:"user_defined"`
	// Comments preserves comments by OMM block and source order.
	Comments OMMComments `json:"comments"`
	// ExactSGP4Epoch is the optional exact split Julian date used by SGP4.
	ExactSGP4Epoch *OMMExactEpoch `json:"exact_sgp4_epoch"`
	// QuantizeTLEDerivedFields reports whether the bridge applies TLE-grid rounding.
	QuantizeTLEDerivedFields bool `json:"quantize_tle_derived_fields"`
}

// OMMExactEpoch preserves the producer's exact split Julian date.
type OMMExactEpoch struct {
	// Whole is the integral Julian-date component.
	Whole float64 `json:"whole"`
	// Fraction is the fractional-day component.
	Fraction float64 `json:"fraction"`
}

// OMMEpoch is a detached civil timestamp including sub-microsecond precision.
type OMMEpoch struct {
	// Year is the full civil year.
	Year int32 `json:"year"`
	// Month is the civil month.
	Month uint32 `json:"month"`
	// Day is the civil day.
	Day uint32 `json:"day"`
	// Hour is the hour of day.
	Hour uint32 `json:"hour"`
	// Minute is the minute of hour.
	Minute uint32 `json:"minute"`
	// Second is the second field, including a retained leap second where valid.
	Second uint32 `json:"second"`
	// Microsecond is the fractional second's whole-microsecond component.
	Microsecond uint32 `json:"microsecond"`
	// Femtosecond is the sub-microsecond remainder in femtoseconds.
	Femtosecond uint32 `json:"femtosecond"`
}

// OMMComments stores the five core comment groups independently.
type OMMComments struct {
	// Header contains comments following the OMM version.
	Header []string `json:"header"`
	// Metadata contains comments before object metadata.
	Metadata []string `json:"metadata"`
	// MeanElements contains comments before mean-element fields.
	MeanElements []string `json:"mean_elements"`
	// TLEParameters contains comments before TLE-related fields.
	TLEParameters []string `json:"tle_parameters"`
	// UserDefined contains comments preceding user-defined parameters.
	UserDefined []string `json:"user_defined"`
}

// OMMSpacecraft holds one optional spacecraft-parameter block.
type OMMSpacecraft struct {
	// Comments contains comments within the spacecraft block.
	Comments []string `json:"comments"`
	// MassKG is the optional spacecraft mass in kilograms.
	MassKG *float64 `json:"mass_kg"`
	// SolarRadAreaM2 is the optional solar-radiation area in m².
	SolarRadAreaM2 *float64 `json:"solar_rad_area_m2"`
	// SolarRadCoeff is the optional dimensionless solar-radiation coefficient.
	SolarRadCoeff *float64 `json:"solar_rad_coeff"`
	// DragAreaM2 is the optional drag area in m².
	DragAreaM2 *float64 `json:"drag_area_m2"`
	// DragCoeff is the optional dimensionless drag coefficient.
	DragCoeff *float64 `json:"drag_coeff"`
}

// OMMCovariance holds the optional covariance frame and 21 lower-triangle values.
type OMMCovariance struct {
	// Comments contains comments within the covariance block.
	Comments []string `json:"comments"`
	// CovRefFrame is the optional covariance reference frame.
	CovRefFrame *string `json:"cov_ref_frame"`
	// LowerTriangle contains the 21 matrix values in CCSDS keyword order.
	LowerTriangle [21]float64 `json:"lower_triangle"`
}

// OMMUserDefined is one user-defined OMM parameter in source order.
type OMMUserDefined struct {
	// Parameter is the suffix of the USER_DEFINED keyword.
	Parameter string `json:"parameter"`
	// Value is its verbatim text value.
	Value string `json:"value"`
}

// OMMElementSet contains the exact core SGP4 bridge result.
type OMMElementSet struct {
	// Epoch is the exact split Julian date used for SGP4 initialization.
	Epoch JulianDate
	// BStar is the SGP4 drag term after the selected compatibility policy.
	BStar float64
	// MeanMotionDot is present when the OMM states it.
	MeanMotionDot *float64
	// MeanMotionDoubleDot is present when the OMM states it.
	MeanMotionDoubleDot *float64
	// Eccentricity is dimensionless.
	Eccentricity float64
	// ArgumentOfPerigeeDeg is the argument of perigee in degrees.
	ArgumentOfPerigeeDeg float64
	// InclinationDeg is the inclination in degrees.
	InclinationDeg float64
	// MeanAnomalyDeg is the mean anomaly in degrees.
	MeanAnomalyDeg float64
	// MeanMotionRevPerDay is mean motion in revolutions per day.
	MeanMotionRevPerDay float64
	// RightAscensionDeg is the ascending-node longitude in degrees.
	RightAscensionDeg float64
	// CatalogNumber is the optional NORAD identifier.
	CatalogNumber *uint32
	// OMMEpochDays is the optional python-sgp4 initialization epoch.
	OMMEpochDays *float64
}

// Snapshot returns a detached, lossless copy of every stored OMM field.
func (o *OMM) Snapshot() (OMMData, error) {
	if o == nil || o.handle == nil {
		return OMMData{}, ErrClosed
	}
	data, err := o.handle.SnapshotJSON()
	if err != nil {
		return OMMData{}, publicError(err)
	}
	if err := validateUniqueJSONKeys(data); err != nil {
		return OMMData{}, fmt.Errorf("sidereon: invalid native OMM snapshot: %w", err)
	}
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(data, &keys); err != nil {
		return OMMData{}, fmt.Errorf("sidereon: invalid native OMM snapshot: %w", err)
	}
	if len(keys) != len(ommSnapshotKeys) {
		return OMMData{}, fmt.Errorf("sidereon: incomplete native OMM snapshot: got %d fields, want %d", len(keys), len(ommSnapshotKeys))
	}
	for _, key := range ommSnapshotKeys {
		if _, ok := keys[key]; !ok {
			return OMMData{}, fmt.Errorf("sidereon: native OMM snapshot is missing %q", key)
		}
	}
	for key := range keys {
		if !containsOMMKey(key) {
			return OMMData{}, fmt.Errorf("sidereon: native OMM snapshot has unknown field %q", key)
		}
	}
	if err := validateOMMRequiredJSON(keys); err != nil {
		return OMMData{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var value OMMData
	if err := decoder.Decode(&value); err != nil {
		return OMMData{}, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return OMMData{}, fmt.Errorf("sidereon: trailing data in OMM snapshot")
	}
	if err := validateOMMData(value); err != nil {
		return OMMData{}, err
	}
	return value, nil
}

func validateOMMRequiredJSON(root map[string]json.RawMessage) error {
	isNull := func(raw json.RawMessage) bool { return bytes.Equal(bytes.TrimSpace(raw), []byte("null")) }
	for _, key := range []string{"epoch", "eccentricity", "inclination_deg", "ra_of_asc_node_deg", "arg_of_pericenter_deg", "mean_anomaly_deg", "user_defined", "comments", "quantize_tle_derived_fields"} {
		if isNull(root[key]) {
			return fmt.Errorf("sidereon: native OMM snapshot field %q cannot be null", key)
		}
	}
	object := func(raw json.RawMessage, field string, required []string) (map[string]json.RawMessage, error) {
		var values map[string]json.RawMessage
		if err := json.Unmarshal(raw, &values); err != nil || values == nil {
			return nil, fmt.Errorf("sidereon: native OMM snapshot %s must be an object", field)
		}
		if len(values) != len(required) {
			return nil, fmt.Errorf("sidereon: native OMM snapshot %s has incomplete fields", field)
		}
		for _, key := range required {
			if _, ok := values[key]; !ok {
				return nil, fmt.Errorf("sidereon: native OMM snapshot %s.%s is missing", field, key)
			}
		}
		return values, nil
	}
	if values, err := object(root["epoch"], "epoch", []string{"year", "month", "day", "hour", "minute", "second", "microsecond", "femtosecond"}); err != nil {
		return err
	} else {
		for key, value := range values {
			if isNull(value) {
				return fmt.Errorf("sidereon: native OMM snapshot epoch.%s cannot be null", key)
			}
		}
	}
	comments, err := object(root["comments"], "comments", []string{"header", "metadata", "mean_elements", "tle_parameters", "user_defined"})
	if err != nil {
		return err
	}
	for key, value := range comments {
		if isNull(value) {
			return fmt.Errorf("sidereon: native OMM snapshot comments.%s cannot be null", key)
		}
	}
	var rows []json.RawMessage
	if err := json.Unmarshal(root["user_defined"], &rows); err != nil || rows == nil {
		return fmt.Errorf("sidereon: native OMM snapshot user_defined must be an array")
	}
	for index, row := range rows {
		values, err := object(row, fmt.Sprintf("user_defined[%d]", index), []string{"parameter", "value"})
		if err != nil {
			return err
		}
		for key, value := range values {
			if isNull(value) {
				return fmt.Errorf("sidereon: native OMM snapshot user_defined[%d].%s cannot be null", index, key)
			}
		}
	}
	for _, pair := range []struct {
		name string
		keys []string
	}{
		{"spacecraft", []string{"comments", "mass_kg", "solar_rad_area_m2", "solar_rad_coeff", "drag_area_m2", "drag_coeff"}},
		{"covariance", []string{"comments", "cov_ref_frame", "lower_triangle"}},
	} {
		if isNull(root[pair.name]) {
			continue
		}
		values, err := object(root[pair.name], pair.name, pair.keys)
		if err != nil {
			return err
		}
		if isNull(values["comments"]) {
			return fmt.Errorf("sidereon: native OMM snapshot %s.comments cannot be null", pair.name)
		}
		if pair.name == "covariance" && isNull(values["lower_triangle"]) {
			return fmt.Errorf("sidereon: native OMM snapshot covariance.lower_triangle cannot be null")
		}
	}
	if !isNull(root["exact_sgp4_epoch"]) {
		values, err := object(root["exact_sgp4_epoch"], "exact_sgp4_epoch", []string{"whole", "fraction"})
		if err != nil {
			return err
		}
		for key, value := range values {
			if isNull(value) {
				return fmt.Errorf("sidereon: native OMM snapshot exact_sgp4_epoch.%s cannot be null", key)
			}
		}
	}
	return nil
}

func validateUniqueJSONKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := consumeUniqueJSONValue(decoder); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON value")
		}
		return err
	}
	return nil
}

func consumeUniqueJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delimiter {
	case '{':
		seen := make(map[string]struct{})
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return fmt.Errorf("object key is not a string")
			}
			if _, exists := seen[key]; exists {
				return fmt.Errorf("duplicate JSON object key %q", key)
			}
			seen[key] = struct{}{}
			if err := consumeUniqueJSONValue(decoder); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil {
			return err
		}
		if end != json.Delim('}') {
			return fmt.Errorf("unterminated JSON object")
		}
	case '[':
		for decoder.More() {
			if err := consumeUniqueJSONValue(decoder); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil {
			return err
		}
		if end != json.Delim(']') {
			return fmt.Errorf("unterminated JSON array")
		}
	default:
		return fmt.Errorf("unexpected JSON delimiter %q", delimiter)
	}
	return nil
}

func containsOMMKey(key string) bool {
	for _, candidate := range ommSnapshotKeys {
		if key == candidate {
			return true
		}
	}
	return false
}

// NewOMMFromData validates and constructs an owned OMM from a complete detached snapshot.
func NewOMMFromData(data OMMData) (*OMM, error) {
	if err := validateOMMData(data); err != nil {
		return nil, err
	}
	data = normalizeOMMDataSlices(data)
	encoded, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &object); err != nil {
		return nil, err
	}
	if data.ExactSGP4Epoch == nil {
		object["exact_sgp4_epoch"] = json.RawMessage("null")
	} else {
		exact, err := json.Marshal(struct {
			Whole    float64 `json:"whole"`
			Fraction float64 `json:"fraction"`
		}{data.ExactSGP4Epoch.Whole, data.ExactSGP4Epoch.Fraction})
		if err != nil {
			return nil, err
		}
		object["exact_sgp4_epoch"] = exact
	}
	for _, key := range ommSnapshotKeys {
		if _, ok := object[key]; !ok {
			object[key] = json.RawMessage("null")
		}
	}
	encoded, err = json.Marshal(object)
	if err != nil {
		return nil, err
	}
	value, err := native.OMMFromSnapshotJSON(encoded)
	if err != nil {
		return nil, publicError(err)
	}
	return &OMM{handle: value}, nil
}

func normalizeOMMDataSlices(v OMMData) OMMData {
	if v.UserDefined == nil {
		v.UserDefined = []OMMUserDefined{}
	}
	if v.Comments.Header == nil {
		v.Comments.Header = []string{}
	}
	if v.Comments.Metadata == nil {
		v.Comments.Metadata = []string{}
	}
	if v.Comments.MeanElements == nil {
		v.Comments.MeanElements = []string{}
	}
	if v.Comments.TLEParameters == nil {
		v.Comments.TLEParameters = []string{}
	}
	if v.Comments.UserDefined == nil {
		v.Comments.UserDefined = []string{}
	}
	if v.Spacecraft != nil && v.Spacecraft.Comments == nil {
		v.Spacecraft.Comments = []string{}
	}
	if v.Covariance != nil && v.Covariance.Comments == nil {
		v.Covariance.Comments = []string{}
	}
	return v
}

func validateOMMData(v OMMData) error {
	finite := func(p *float64) bool { return p == nil || (!math.IsNaN(*p) && !math.IsInf(*p, 0)) }
	values := []float64{v.Eccentricity, v.InclinationDeg, v.RAOfAscNodeDeg, v.ArgOfPericenterDeg, v.MeanAnomalyDeg}
	for _, x := range values {
		if math.IsNaN(x) || math.IsInf(x, 0) {
			return fmt.Errorf("sidereon: OMM snapshot has non-finite required numeric field")
		}
	}
	for _, p := range []*float64{v.MeanMotion, v.SemiMajorAxisKM, v.GMKM3S2, v.BStar, v.BTermM2KG, v.MeanMotionDot, v.MeanMotionDDot, v.AGOMM2KG} {
		if !finite(p) {
			return fmt.Errorf("sidereon: OMM snapshot has non-finite optional numeric field")
		}
	}
	if v.ExactSGP4Epoch != nil && (math.IsNaN(v.ExactSGP4Epoch.Whole) || math.IsInf(v.ExactSGP4Epoch.Whole, 0) || math.IsNaN(v.ExactSGP4Epoch.Fraction) || math.IsInf(v.ExactSGP4Epoch.Fraction, 0)) {
		return fmt.Errorf("sidereon: OMM exact epoch must be finite")
	}
	if v.Spacecraft != nil {
		for _, p := range []*float64{v.Spacecraft.MassKG, v.Spacecraft.SolarRadAreaM2, v.Spacecraft.SolarRadCoeff, v.Spacecraft.DragAreaM2, v.Spacecraft.DragCoeff} {
			if !finite(p) {
				return fmt.Errorf("sidereon: OMM spacecraft value must be finite")
			}
		}
	}
	if v.Covariance != nil {
		for _, x := range v.Covariance.LowerTriangle {
			if math.IsNaN(x) || math.IsInf(x, 0) {
				return fmt.Errorf("sidereon: OMM covariance value must be finite")
			}
		}
	}
	return nil
}

// ToElementSet applies the core OMM-to-SGP4 bridge without encoding and
// reparsing, so exact in-memory epoch and quantization settings are retained.
func (o *OMM) ToElementSet() (OMMElementSet, error) {
	if o == nil || o.handle == nil {
		return OMMElementSet{}, ErrClosed
	}
	value, err := o.handle.ToElementSet()
	if err != nil {
		return OMMElementSet{}, publicError(err)
	}
	result := OMMElementSet{
		Epoch: JulianDate{Whole: value.EpochWhole, Fraction: value.EpochFraction},
		BStar: value.BStar, Eccentricity: value.Eccentricity,
		ArgumentOfPerigeeDeg: value.ArgumentOfPerigeeDeg, InclinationDeg: value.InclinationDeg,
		MeanAnomalyDeg: value.MeanAnomalyDeg, MeanMotionRevPerDay: value.MeanMotionRevPerDay,
		RightAscensionDeg: value.RightAscensionDeg,
	}
	if value.MeanMotionDotPresent {
		result.MeanMotionDot = &value.MeanMotionDot
	}
	if value.MeanMotionDoubleDotPresent {
		result.MeanMotionDoubleDot = &value.MeanMotionDoubleDot
	}
	if value.CatalogNumberPresent {
		result.CatalogNumber = &value.CatalogNumber
	}
	if value.OMMEpochDaysPresent {
		result.OMMEpochDays = &value.OMMEpochDays
	}
	return result, nil
}
