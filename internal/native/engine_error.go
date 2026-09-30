package native

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
)

// EngineErrorFamily identifies the core subsystem producing a schema1 generic engine error.
type EngineErrorFamily uint32

const (
	EngineErrorFamilyNone               EngineErrorFamily = 0
	EngineErrorFamilyRtk                EngineErrorFamily = 1
	EngineErrorFamilyStaticReference    EngineErrorFamily = 2
	EngineErrorFamilyTrls               EngineErrorFamily = 3
	EngineErrorFamilyIls                EngineErrorFamily = 4
	EngineErrorFamilySpk                EngineErrorFamily = 5
	EngineErrorFamilyCdm                EngineErrorFamily = 6
	EngineErrorFamilyTdm                EngineErrorFamily = 7
	EngineErrorFamilyFusion             EngineErrorFamily = 8
	EngineErrorFamilyFusionStateCodec   EngineErrorFamily = 9
	EngineErrorFamilyAllan              EngineErrorFamily = 10
	EngineErrorFamilyPowerLawNoise      EngineErrorFamily = 11
	EngineErrorFamilyFrameCatalog       EngineErrorFamily = 12
	EngineErrorFamilySidereal           EngineErrorFamily = 13
	EngineErrorFamilyAtmosphere         EngineErrorFamily = 14
	EngineErrorFamilySourceLocalization EngineErrorFamily = 15
	EngineErrorFamilyGeodeticTimeSeries EngineErrorFamily = 16
	EngineErrorFamilyNormality          EngineErrorFamily = 17
	EngineErrorFamilyTrack              EngineErrorFamily = 18
	EngineErrorFamilyPreciseSamples     EngineErrorFamily = 19
	EngineErrorFamilyPreciseInterpolant EngineErrorFamily = 20
	EngineErrorFamilySpaceWeather       EngineErrorFamily = 21
	EngineErrorFamilyAraim              EngineErrorFamily = 22
	EngineErrorFamilyReducedOrbit       EngineErrorFamily = 23
	EngineErrorFamilyReducedOrbitSource EngineErrorFamily = 24
	EngineErrorFamilyPiecewiseOrbit     EngineErrorFamily = 25
	EngineErrorFamilyOrbitFit           EngineErrorFamily = 26
	EngineErrorFamilyElements           EngineErrorFamily = 27
	EngineErrorFamilyEquinoctial        EngineErrorFamily = 28
	EngineErrorFamilyRtnFrame           EngineErrorFamily = 29
	EngineErrorFamilyAnomaly            EngineErrorFamily = 30
	EngineErrorFamilyPropagation        EngineErrorFamily = 31
	EngineErrorFamilyDecay              EngineErrorFamily = 32
	EngineErrorFamilyDgnss              EngineErrorFamily = 33
	EngineErrorFamilyScenario           EngineErrorFamily = 34
	EngineErrorFamilyCatalog            EngineErrorFamily = 35
	EngineErrorFamilyExactCache         EngineErrorFamily = 36
	EngineErrorFamilyTca                EngineErrorFamily = 37
	EngineErrorFamilyAlmanac            EngineErrorFamily = 38
	EngineErrorFamilyObserve            EngineErrorFamily = 39
	EngineErrorFamilyBodyObservation    EngineErrorFamily = 40
	EngineErrorFamilyLookAngle          EngineErrorFamily = 41
	EngineErrorFamilyPass               EngineErrorFamily = 42
	EngineErrorFamilyEventFinder        EngineErrorFamily = 43
	EngineErrorFamilyFrameTransform     EngineErrorFamily = 44
	EngineErrorFamilyConjunction        EngineErrorFamily = 45
	EngineErrorFamilyFacade             EngineErrorFamily = 46
	EngineErrorFamilySpp                EngineErrorFamily = 47
	EngineErrorFamilySppPolicy          EngineErrorFamily = 48
	EngineErrorFamilySunMoon            EngineErrorFamily = 49
	EngineErrorFamilyRinexSpp           EngineErrorFamily = 50
	EngineErrorFamilySolutionValidation EngineErrorFamily = 51
	EngineErrorFamilyRF                 EngineErrorFamily = 52
	EngineErrorFamilyIonosphereFree     EngineErrorFamily = 53
	EngineErrorFamilyDoppler            EngineErrorFamily = 54
	EngineErrorFamilyOEM                EngineErrorFamily = 55
	EngineErrorFamilyOPM                EngineErrorFamily = 56
	EngineErrorFamilyOMM                EngineErrorFamily = 57
	EngineErrorFamilyDOP                EngineErrorFamily = 58
	EngineErrorFamilyGeofence           EngineErrorFamily = 59
	EngineErrorFamilySignal             EngineErrorFamily = 60
	EngineErrorFamilyCarrierPhase       EngineErrorFamily = 61
	EngineErrorFamilySignalAnalysis     EngineErrorFamily = 62
	EngineErrorFamilyErrorMetrics       EngineErrorFamily = 63
	EngineErrorFamilyObservables        EngineErrorFamily = 64
	EngineErrorFamilyNMEA               EngineErrorFamily = 65
	EngineErrorFamilyTLEFit             EngineErrorFamily = 66
	EngineErrorFamilyIOD                EngineErrorFamily = 67
	EngineErrorFamilySelection          EngineErrorFamily = 68
	EngineErrorFamilyPppAutoInit        EngineErrorFamily = 69
	EngineErrorFamilyStaticPositioning  EngineErrorFamily = 70
	EngineErrorFamilyTimeOffset         EngineErrorFamily = 71
	EngineErrorFamilyTimeModel          EngineErrorFamily = 72
	EngineErrorFamilyUnknown            EngineErrorFamily = 999
)

func (f EngineErrorFamily) Name() string {
	switch f {
	case EngineErrorFamilyNone:
		return "none"
	case EngineErrorFamilyRtk:
		return "rtk"
	case EngineErrorFamilyStaticReference:
		return "static_reference"
	case EngineErrorFamilyTrls:
		return "trls"
	case EngineErrorFamilyIls:
		return "ils"
	case EngineErrorFamilySpk:
		return "spk"
	case EngineErrorFamilyCdm:
		return "cdm"
	case EngineErrorFamilyTdm:
		return "tdm"
	case EngineErrorFamilyFusion:
		return "fusion"
	case EngineErrorFamilyFusionStateCodec:
		return "fusion_state_codec"
	case EngineErrorFamilyAllan:
		return "allan"
	case EngineErrorFamilyPowerLawNoise:
		return "power_law_noise"
	case EngineErrorFamilyFrameCatalog:
		return "frame_catalog"
	case EngineErrorFamilySidereal:
		return "sidereal"
	case EngineErrorFamilyAtmosphere:
		return "atmosphere"
	case EngineErrorFamilySourceLocalization:
		return "source_localization"
	case EngineErrorFamilyGeodeticTimeSeries:
		return "geodetic_time_series"
	case EngineErrorFamilyNormality:
		return "normality"
	case EngineErrorFamilyTrack:
		return "track"
	case EngineErrorFamilyPreciseSamples:
		return "precise_samples"
	case EngineErrorFamilyPreciseInterpolant:
		return "precise_interpolant"
	case EngineErrorFamilySpaceWeather:
		return "space_weather"
	case EngineErrorFamilyAraim:
		return "araim"
	case EngineErrorFamilyReducedOrbit:
		return "reduced_orbit"
	case EngineErrorFamilyReducedOrbitSource:
		return "reduced_orbit_source"
	case EngineErrorFamilyPiecewiseOrbit:
		return "piecewise_orbit"
	case EngineErrorFamilyOrbitFit:
		return "orbit_fit"
	case EngineErrorFamilyElements:
		return "elements"
	case EngineErrorFamilyEquinoctial:
		return "equinoctial"
	case EngineErrorFamilyRtnFrame:
		return "rtn_frame"
	case EngineErrorFamilyAnomaly:
		return "anomaly"
	case EngineErrorFamilyPropagation:
		return "propagation"
	case EngineErrorFamilyDecay:
		return "decay"
	case EngineErrorFamilyDgnss:
		return "dgnss"
	case EngineErrorFamilyScenario:
		return "scenario"
	case EngineErrorFamilyCatalog:
		return "catalog"
	case EngineErrorFamilyExactCache:
		return "exact_cache"
	case EngineErrorFamilyTca:
		return "tca"
	case EngineErrorFamilyAlmanac:
		return "almanac"
	case EngineErrorFamilyObserve:
		return "observe"
	case EngineErrorFamilyBodyObservation:
		return "body_observation"
	case EngineErrorFamilyLookAngle:
		return "look_angle"
	case EngineErrorFamilyPass:
		return "pass"
	case EngineErrorFamilyEventFinder:
		return "event_finder"
	case EngineErrorFamilyFrameTransform:
		return "frame_transform"
	case EngineErrorFamilyConjunction:
		return "conjunction"
	case EngineErrorFamilyFacade:
		return "facade"
	case EngineErrorFamilySpp:
		return "spp"
	case EngineErrorFamilySppPolicy:
		return "spp_policy"
	case EngineErrorFamilySunMoon:
		return "sun_moon"
	case EngineErrorFamilyRinexSpp:
		return "rinex_spp"
	case EngineErrorFamilySolutionValidation:
		return "solution_validation"
	case EngineErrorFamilyRF:
		return "rf"
	case EngineErrorFamilyIonosphereFree:
		return "ionosphere_free"
	case EngineErrorFamilyDoppler:
		return "doppler"
	case EngineErrorFamilyOEM:
		return "oem"
	case EngineErrorFamilyOPM:
		return "opm"
	case EngineErrorFamilyOMM:
		return "omm"
	case EngineErrorFamilyDOP:
		return "dop"
	case EngineErrorFamilyGeofence:
		return "geofence"
	case EngineErrorFamilySignal:
		return "signal"
	case EngineErrorFamilyCarrierPhase:
		return "carrier_phase"
	case EngineErrorFamilySignalAnalysis:
		return "signal_analysis"
	case EngineErrorFamilyErrorMetrics:
		return "error_metrics"
	case EngineErrorFamilyObservables:
		return "observables"
	case EngineErrorFamilyNMEA:
		return "nmea"
	case EngineErrorFamilyTLEFit:
		return "tle_fit"
	case EngineErrorFamilyIOD:
		return "iod"
	case EngineErrorFamilySelection:
		return "selection"
	case EngineErrorFamilyPppAutoInit:
		return "ppp_auto_init"
	case EngineErrorFamilyStaticPositioning:
		return "static_positioning"
	case EngineErrorFamilyTimeOffset:
		return "time_offset"
	case EngineErrorFamilyTimeModel:
		return "time_model"
	case EngineErrorFamilyUnknown:
		return "unknown"
	default:
		return fmt.Sprintf("family_%d", f)
	}
}

func (f EngineErrorFamily) String() string {
	return f.Name()
}

func EngineErrorFamilyFromName(name string) EngineErrorFamily {
	switch name {
	case "none":
		return EngineErrorFamilyNone
	case "rtk":
		return EngineErrorFamilyRtk
	case "static_reference":
		return EngineErrorFamilyStaticReference
	case "trls":
		return EngineErrorFamilyTrls
	case "ils":
		return EngineErrorFamilyIls
	case "spk":
		return EngineErrorFamilySpk
	case "cdm":
		return EngineErrorFamilyCdm
	case "tdm":
		return EngineErrorFamilyTdm
	case "fusion":
		return EngineErrorFamilyFusion
	case "fusion_state_codec":
		return EngineErrorFamilyFusionStateCodec
	case "allan":
		return EngineErrorFamilyAllan
	case "power_law_noise":
		return EngineErrorFamilyPowerLawNoise
	case "frame_catalog":
		return EngineErrorFamilyFrameCatalog
	case "sidereal":
		return EngineErrorFamilySidereal
	case "atmosphere":
		return EngineErrorFamilyAtmosphere
	case "source_localization":
		return EngineErrorFamilySourceLocalization
	case "geodetic_time_series":
		return EngineErrorFamilyGeodeticTimeSeries
	case "normality":
		return EngineErrorFamilyNormality
	case "track":
		return EngineErrorFamilyTrack
	case "precise_samples":
		return EngineErrorFamilyPreciseSamples
	case "precise_interpolant":
		return EngineErrorFamilyPreciseInterpolant
	case "space_weather":
		return EngineErrorFamilySpaceWeather
	case "araim":
		return EngineErrorFamilyAraim
	case "reduced_orbit":
		return EngineErrorFamilyReducedOrbit
	case "reduced_orbit_source":
		return EngineErrorFamilyReducedOrbitSource
	case "piecewise_orbit":
		return EngineErrorFamilyPiecewiseOrbit
	case "orbit_fit":
		return EngineErrorFamilyOrbitFit
	case "elements":
		return EngineErrorFamilyElements
	case "equinoctial":
		return EngineErrorFamilyEquinoctial
	case "rtn_frame":
		return EngineErrorFamilyRtnFrame
	case "anomaly":
		return EngineErrorFamilyAnomaly
	case "propagation":
		return EngineErrorFamilyPropagation
	case "decay":
		return EngineErrorFamilyDecay
	case "dgnss":
		return EngineErrorFamilyDgnss
	case "scenario":
		return EngineErrorFamilyScenario
	case "catalog":
		return EngineErrorFamilyCatalog
	case "exact_cache":
		return EngineErrorFamilyExactCache
	case "tca":
		return EngineErrorFamilyTca
	case "almanac":
		return EngineErrorFamilyAlmanac
	case "observe":
		return EngineErrorFamilyObserve
	case "body_observation":
		return EngineErrorFamilyBodyObservation
	case "look_angle":
		return EngineErrorFamilyLookAngle
	case "pass":
		return EngineErrorFamilyPass
	case "event_finder":
		return EngineErrorFamilyEventFinder
	case "frame_transform":
		return EngineErrorFamilyFrameTransform
	case "conjunction":
		return EngineErrorFamilyConjunction
	case "facade":
		return EngineErrorFamilyFacade
	case "spp":
		return EngineErrorFamilySpp
	case "spp_policy":
		return EngineErrorFamilySppPolicy
	case "sun_moon":
		return EngineErrorFamilySunMoon
	case "rinex_spp":
		return EngineErrorFamilyRinexSpp
	case "solution_validation":
		return EngineErrorFamilySolutionValidation
	case "rf":
		return EngineErrorFamilyRF
	case "ionosphere_free":
		return EngineErrorFamilyIonosphereFree
	case "doppler":
		return EngineErrorFamilyDoppler
	case "oem":
		return EngineErrorFamilyOEM
	case "opm":
		return EngineErrorFamilyOPM
	case "omm":
		return EngineErrorFamilyOMM
	case "dop":
		return EngineErrorFamilyDOP
	case "geofence":
		return EngineErrorFamilyGeofence
	case "signal":
		return EngineErrorFamilySignal
	case "carrier_phase":
		return EngineErrorFamilyCarrierPhase
	case "signal_analysis":
		return EngineErrorFamilySignalAnalysis
	case "error_metrics":
		return EngineErrorFamilyErrorMetrics
	case "observables":
		return EngineErrorFamilyObservables
	case "nmea":
		return EngineErrorFamilyNMEA
	case "tle_fit":
		return EngineErrorFamilyTLEFit
	case "iod":
		return EngineErrorFamilyIOD
	case "selection":
		return EngineErrorFamilySelection
	case "ppp_auto_init":
		return EngineErrorFamilyPppAutoInit
	case "static_positioning":
		return EngineErrorFamilyStaticPositioning
	case "time_offset":
		return EngineErrorFamilyTimeOffset
	case "time_model":
		return EngineErrorFamilyTimeModel
	default:
		return EngineErrorFamilyUnknown
	}
}

// EngineFloat represents an exact binary64 value serialized in the C schema1 engine error envelope.
type EngineFloat struct {
	Decimal string `json:"decimal"`
	BitsHex string `json:"bits_hex"`
}

// Float64 parses the bit pattern into an exact float64 value (preserving non-finite values and -0.0).
func (ef EngineFloat) Float64() (float64, error) {
	bits, err := ef.Bits()
	if err != nil {
		return 0, err
	}
	return math.Float64frombits(bits), nil
}

// Bits parses the raw uint64 bit pattern.
func (ef EngineFloat) Bits() (uint64, error) {
	if len(ef.BitsHex) != 16 {
		return 0, fmt.Errorf("sidereon: engine float bits_hex must be exactly 16 hex characters, got %d", len(ef.BitsHex))
	}
	for _, c := range ef.BitsHex {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return 0, fmt.Errorf("sidereon: invalid hex in bits_hex %q", ef.BitsHex)
		}
	}
	return strconv.ParseUint(ef.BitsHex, 16, 64)
}

// ParseEngineFloat extracts an EngineFloat from raw JSON bytes.
func ParseEngineFloat(data []byte) (EngineFloat, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return EngineFloat{}, errors.New("sidereon: empty engine float data")
	}
	if string(trimmed) == "null" {
		return EngineFloat{}, errors.New("sidereon: engine float cannot be null")
	}
	if trimmed[0] != '{' {
		return EngineFloat{}, errors.New("sidereon: engine float must be a JSON object")
	}
	if err := checkDuplicateKeys(trimmed); err != nil {
		return EngineFloat{}, fmt.Errorf("sidereon: engine float duplicate key: %w", err)
	}
	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &rawMap); err != nil {
		return EngineFloat{}, fmt.Errorf("sidereon: invalid engine float JSON: %w", err)
	}
	decRaw, okDec := rawMap["decimal"]
	if !okDec || bytes.Equal(bytes.TrimSpace(decRaw), []byte("null")) {
		return EngineFloat{}, errors.New("sidereon: engine float missing decimal field")
	}
	bitsRaw, okBits := rawMap["bits_hex"]
	if !okBits || bytes.Equal(bytes.TrimSpace(bitsRaw), []byte("null")) {
		return EngineFloat{}, errors.New("sidereon: engine float missing bits_hex field")
	}
	var decStr, bitsStr string
	if err := json.Unmarshal(decRaw, &decStr); err != nil {
		return EngineFloat{}, fmt.Errorf("sidereon: invalid decimal field: %w", err)
	}
	if err := json.Unmarshal(bitsRaw, &bitsStr); err != nil {
		return EngineFloat{}, fmt.Errorf("sidereon: invalid bits_hex field: %w", err)
	}
	ef := EngineFloat{
		Decimal: decStr,
		BitsHex: bitsStr,
	}
	if _, err := ef.Bits(); err != nil {
		return EngineFloat{}, err
	}
	return ef, nil
}

func checkDuplicateKeys(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	return checkDuplicateTokens(dec, tok)
}

func checkDuplicateTokens(dec *json.Decoder, tok json.Token) error {
	delim, ok := tok.(json.Delim)
	if !ok {
		return nil
	}
	if delim == '{' {
		seen := make(map[string]bool)
		for dec.More() {
			kTok, err := dec.Token()
			if err != nil {
				return err
			}
			key, ok := kTok.(string)
			if !ok {
				return fmt.Errorf("expected string key in object, got %T", kTok)
			}
			if seen[key] {
				return fmt.Errorf("duplicate key %q in object", key)
			}
			seen[key] = true
			vTok, err := dec.Token()
			if err != nil {
				return err
			}
			if err := checkDuplicateTokens(dec, vTok); err != nil {
				return err
			}
		}
		endTok, err := dec.Token()
		if err != nil {
			return err
		}
		if endDelim, ok := endTok.(json.Delim); !ok || endDelim != '}' {
			return fmt.Errorf("expected '}', got %v", endTok)
		}
	} else if delim == '[' {
		for dec.More() {
			vTok, err := dec.Token()
			if err != nil {
				return err
			}
			if err := checkDuplicateTokens(dec, vTok); err != nil {
				return err
			}
		}
		endTok, err := dec.Token()
		if err != nil {
			return err
		}
		if endDelim, ok := endTok.(json.Delim); !ok || endDelim != ']' {
			return fmt.Errorf("expected ']', got %v", endTok)
		}
	}
	return nil
}

// EngineError preserves the native generic schema1 engine error transport record.
// If diagnostic capture or payload decoding encounters an error, partial data
// (such as Family, FamilyName, and raw Payload) is preserved alongside the
// diagnostic refusal in CaptureError according to the partial-record contract.
type EngineError struct {
	// Family is the numeric family identifier from the C ABI.
	Family EngineErrorFamily
	// FamilyName is the string name of the error family.
	FamilyName string
	// Schema is the schema version from the envelope (schema_version=1).
	Schema uint32
	// Operation is the logical operation name reported by the core engine.
	Operation string
	// Kind is the primary error discriminant string from the error node.
	Kind string
	// Fields contains the raw JSON bytes of the error fields.
	// Preserves arbitrary future fields/codes and all integer lexemes without float64 conversion.
	Fields json.RawMessage
	// TypedFields is the recursively typed view of Fields. Numbers remain
	// json.Number values so large integers and exact decimal tokens are preserved.
	TypedFields map[string]EngineJSONValue
	// Payload is the complete versioned JSON payload returned by the C ABI.
	Payload json.RawMessage
	// CaptureError reports an explicit error when C info, payload query, or JSON decoding fails.
	CaptureError error
}

// EngineJSONKind identifies one exact JSON value category in a typed engine
// error payload.
type EngineJSONKind uint8

const (
	EngineJSONNull EngineJSONKind = iota
	EngineJSONString
	EngineJSONNumber
	EngineJSONBoolean
	EngineJSONObject
	EngineJSONArray
)

// EngineJSONValue is a lossless typed JSON value. Object keys retain the
// serializer's field names, and Number stores the original JSON numeric token.
type EngineJSONValue struct {
	Kind   EngineJSONKind
	String string
	Number json.Number
	Bool   bool
	Object map[string]EngineJSONValue
	Array  []EngineJSONValue
}

func engineJSONValue(value any) (EngineJSONValue, error) {
	switch value := value.(type) {
	case nil:
		return EngineJSONValue{Kind: EngineJSONNull}, nil
	case string:
		return EngineJSONValue{Kind: EngineJSONString, String: value}, nil
	case json.Number:
		return EngineJSONValue{Kind: EngineJSONNumber, Number: value}, nil
	case bool:
		return EngineJSONValue{Kind: EngineJSONBoolean, Bool: value}, nil
	case map[string]any:
		object := make(map[string]EngineJSONValue, len(value))
		for key, item := range value {
			converted, err := engineJSONValue(item)
			if err != nil {
				return EngineJSONValue{}, err
			}
			object[key] = converted
		}
		return EngineJSONValue{Kind: EngineJSONObject, Object: object}, nil
	case []any:
		array := make([]EngineJSONValue, len(value))
		for index, item := range value {
			converted, err := engineJSONValue(item)
			if err != nil {
				return EngineJSONValue{}, err
			}
			array[index] = converted
		}
		return EngineJSONValue{Kind: EngineJSONArray, Array: array}, nil
	default:
		return EngineJSONValue{}, fmt.Errorf("sidereon: unsupported decoded engine JSON value %T", value)
	}
}

func decodeEngineJSONFields(raw []byte) (map[string]EngineJSONValue, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var decoded map[string]any
	if err := decoder.Decode(&decoded); err != nil {
		return nil, err
	}
	fields := make(map[string]EngineJSONValue, len(decoded))
	for key, value := range decoded {
		converted, err := engineJSONValue(value)
		if err != nil {
			return nil, err
		}
		fields[key] = converted
	}
	return fields, nil
}

// DecodeEngineJSONFields decodes an object into lossless recursive engine JSON
// values. Numeric tokens remain json.Number values.
func DecodeEngineJSONFields(raw []byte) (map[string]EngineJSONValue, error) {
	return decodeEngineJSONFields(raw)
}

func (e *EngineError) Error() string {
	if e == nil {
		return "sidereon: engine error"
	}
	name := e.FamilyName
	if name == "" {
		name = e.Family.Name()
	}
	var msg string
	if e.Operation != "" && e.Kind != "" {
		msg = fmt.Sprintf("sidereon: engine error family=%s (%d) op=%s kind=%s", name, e.Family, e.Operation, e.Kind)
	} else if e.Operation != "" {
		msg = fmt.Sprintf("sidereon: engine error family=%s (%d) op=%s", name, e.Family, e.Operation)
	} else if e.Kind != "" {
		msg = fmt.Sprintf("sidereon: engine error family=%s (%d) kind=%s", name, e.Family, e.Kind)
	} else {
		msg = fmt.Sprintf("sidereon: engine error family=%s (%d)", name, e.Family)
	}
	if e.CaptureError != nil {
		msg += fmt.Sprintf("; capture error: %v", e.CaptureError)
	}
	return msg
}

func (e *EngineError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.CaptureError
}

// UnmarshalFields decodes e.Fields into target using json.Number so that integer
// lexemes are preserved without lossy float64 conversion.
func (e *EngineError) UnmarshalFields(target any) error {
	if len(e.Fields) == 0 {
		return errors.New("sidereon: engine error has no fields")
	}
	dec := json.NewDecoder(bytes.NewReader(e.Fields))
	dec.UseNumber()
	return dec.Decode(target)
}

// DecodeSchema1EnginePayload decodes a schema1 versioned JSON engine error payload.
// Follows non-loss policy: preserves arbitrary future fields/codes and all integer lexemes.
// If malformed or unknown schema version, retains raw payload and explicit diagnostic error.
func DecodeSchema1EnginePayload(family EngineErrorFamily, payload []byte) (*EngineError, error) {
	if len(payload) == 0 {
		err := errors.New("sidereon: empty engine error payload")
		return &EngineError{
			Family:       family,
			FamilyName:   family.Name(),
			CaptureError: err,
		}, err
	}

	result := &EngineError{
		Family:     family,
		FamilyName: family.Name(),
		Payload:    append(json.RawMessage(nil), payload...),
	}

	trimmed := bytes.TrimSpace(payload)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		err := errors.New("sidereon: engine error payload must be a JSON object")
		result.CaptureError = err
		return result, err
	}

	if !json.Valid(trimmed) {
		err := errors.New("sidereon: engine error payload is not valid JSON")
		result.CaptureError = err
		return result, err
	}

	if err := checkDuplicateKeys(trimmed); err != nil {
		captureErr := fmt.Errorf("sidereon: engine error payload contains duplicate key: %w", err)
		result.CaptureError = captureErr
		return result, captureErr
	}

	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &rawMap); err != nil {
		captureErr := fmt.Errorf("sidereon: engine error envelope unmarshal failed: %w", err)
		result.CaptureError = captureErr
		return result, captureErr
	}

	schemaRaw, okSchema := rawMap["schema_version"]
	if !okSchema || bytes.Equal(bytes.TrimSpace(schemaRaw), []byte("null")) {
		err := errors.New("sidereon: engine error payload missing schema_version")
		result.CaptureError = err
		return result, err
	}

	var schema uint32
	dec := json.NewDecoder(bytes.NewReader(schemaRaw))
	dec.UseNumber()
	if err := dec.Decode(&schema); err != nil {
		err := fmt.Errorf("sidereon: invalid schema_version: %w", err)
		result.CaptureError = err
		return result, err
	}

	result.Schema = schema
	if schema != 1 {
		err := fmt.Errorf("sidereon: unsupported engine error schema version %d", schema)
		result.CaptureError = err
		return result, err
	}

	familyRaw, okFamily := rawMap["family"]
	if !okFamily || bytes.Equal(bytes.TrimSpace(familyRaw), []byte("null")) {
		err := errors.New("sidereon: engine error envelope missing family field")
		result.CaptureError = err
		return result, err
	}
	var envFamily string
	if err := json.Unmarshal(familyRaw, &envFamily); err != nil {
		err := fmt.Errorf("sidereon: invalid family field: %w", err)
		result.CaptureError = err
		return result, err
	}
	if envFamily == "" {
		err := errors.New("sidereon: engine error envelope family must not be empty")
		result.CaptureError = err
		return result, err
	}

	opRaw, okOp := rawMap["operation"]
	if !okOp || bytes.Equal(bytes.TrimSpace(opRaw), []byte("null")) {
		err := errors.New("sidereon: engine error envelope missing operation field")
		result.CaptureError = err
		return result, err
	}
	var envOp string
	if err := json.Unmarshal(opRaw, &envOp); err != nil {
		err := fmt.Errorf("sidereon: invalid operation field: %w", err)
		result.CaptureError = err
		return result, err
	}
	if envOp == "" {
		err := errors.New("sidereon: engine error envelope operation must not be empty")
		result.CaptureError = err
		return result, err
	}
	result.Operation = envOp

	errRaw, okErr := rawMap["error"]
	if !okErr || len(errRaw) == 0 || bytes.Equal(bytes.TrimSpace(errRaw), []byte("null")) {
		err := errors.New("sidereon: engine error envelope missing error object")
		result.CaptureError = err
		return result, err
	}
	trimmedError := bytes.TrimSpace(errRaw)
	if len(trimmedError) == 0 || trimmedError[0] != '{' {
		result.Fields = append(json.RawMessage(nil), errRaw...)
		err := errors.New("sidereon: engine error envelope error field must be a JSON object")
		result.CaptureError = err
		return result, err
	}

	var bodyMap map[string]json.RawMessage
	if err := json.Unmarshal(trimmedError, &bodyMap); err != nil {
		result.Fields = append(json.RawMessage(nil), errRaw...)
		captureErr := fmt.Errorf("sidereon: malformed engine error body: %w", err)
		result.CaptureError = captureErr
		return result, captureErr
	}

	kindRaw, okKind := bodyMap["kind"]
	if !okKind || bytes.Equal(bytes.TrimSpace(kindRaw), []byte("null")) {
		err := errors.New("sidereon: engine error body missing kind field")
		result.CaptureError = err
		return result, err
	}
	var kind string
	if err := json.Unmarshal(kindRaw, &kind); err != nil {
		err := fmt.Errorf("sidereon: invalid kind field: %w", err)
		result.CaptureError = err
		return result, err
	}
	if kind == "" {
		err := errors.New("sidereon: engine error body kind must not be empty")
		result.CaptureError = err
		return result, err
	}
	result.Kind = kind

	fieldsRaw, okFields := bodyMap["fields"]
	if !okFields || len(fieldsRaw) == 0 || bytes.Equal(bytes.TrimSpace(fieldsRaw), []byte("null")) {
		err := errors.New("sidereon: engine error body missing fields object")
		result.CaptureError = err
		return result, err
	}
	trimmedFields := bytes.TrimSpace(fieldsRaw)
	if len(trimmedFields) == 0 || trimmedFields[0] != '{' {
		result.Fields = append(json.RawMessage(nil), fieldsRaw...)
		err := errors.New("sidereon: engine error body fields must be a JSON object")
		result.CaptureError = err
		return result, err
	}
	result.Fields = append(json.RawMessage(nil), fieldsRaw...)
	typedFields, typedErr := decodeEngineJSONFields(fieldsRaw)
	if typedErr != nil {
		captureErr := fmt.Errorf("sidereon: typed engine error fields decode failed: %w", typedErr)
		result.CaptureError = captureErr
		return result, captureErr
	}
	result.TypedFields = typedFields

	// Family consistency & future numeric family behavior
	isKnownNumeric := (family != EngineErrorFamilyNone && family != EngineErrorFamilyUnknown && family.Name() != fmt.Sprintf("family_%d", family))
	knownByName := EngineErrorFamilyFromName(envFamily)

	if isKnownNumeric {
		expectedName := family.Name()
		if envFamily != expectedName {
			result.FamilyName = envFamily
			err := fmt.Errorf("sidereon: engine error family mismatch: summary=%s (%d) envelope=%s", expectedName, family, envFamily)
			result.CaptureError = err
			return result, err
		}
		result.FamilyName = expectedName
	} else {
		// Unknown future numeric family or family == EngineErrorFamilyUnknown
		if knownByName != EngineErrorFamilyUnknown && knownByName != EngineErrorFamilyNone {
			result.FamilyName = envFamily
			err := fmt.Errorf("sidereon: engine error family mismatch: envelope name %q expects numeric %d, got %d", envFamily, knownByName, family)
			result.CaptureError = err
			return result, err
		}
		result.Family = family
		result.FamilyName = envFamily
	}

	return result, nil
}
