package sidereon

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
)

// RINEXHeaderFloat64 is a header floating-point value. It preserves the native
// JSON spellings NaN, Infinity, and -Infinity as their IEEE-754 values.
type RINEXHeaderFloat64 float64

// Float64 returns the value as a built-in float64.
func (value RINEXHeaderFloat64) Float64() float64 { return float64(value) }

// MarshalJSON writes finite values as JSON number tokens and preserves the
// explicit string spelling used for non-finite values by the native route.
func (value RINEXHeaderFloat64) MarshalJSON() ([]byte, error) {
	number := float64(value)
	switch {
	case math.IsNaN(number):
		return []byte("\"NaN\""), nil
	case math.IsInf(number, 1):
		return []byte("\"Infinity\""), nil
	case math.IsInf(number, -1):
		return []byte("\"-Infinity\""), nil
	default:
		return strconv.AppendFloat(nil, number, 'g', -1, 64), nil
	}
}

// UnmarshalJSON accepts finite JSON numbers and the explicit non-finite strings
// used by the native C detail route.
func (value *RINEXHeaderFloat64) UnmarshalJSON(data []byte) error {
	var text string
	if len(data) > 0 && data[0] == '"' {
		if err := json.Unmarshal(data, &text); err != nil {
			return err
		}
		switch text {
		case "NaN":
			*value = RINEXHeaderFloat64(math.NaN())
			return nil
		case "Infinity":
			*value = RINEXHeaderFloat64(math.Inf(1))
			return nil
		case "-Infinity":
			*value = RINEXHeaderFloat64(math.Inf(-1))
			return nil
		default:
			return fmt.Errorf("sidereon: invalid non-finite header float %q", text)
		}
	}
	parsed, err := strconv.ParseFloat(string(data), 64)
	if err != nil {
		return err
	}
	*value = RINEXHeaderFloat64(parsed)
	return nil
}

// RINEXObservationHeaderSegmentDetails retains one complete detached header.
type RINEXObservationHeaderSegmentDetails struct {
	FirstEpochIndex uint64                        `json:"first_epoch_index"`
	Header          RINEXObservationHeaderDetails `json:"header"`
}

// RINEXObservationHeaderDetails exposes every core ObsHeader field. Optional
// values use pointers; counts retain uint64 precision; list order follows the
// source header.
type RINEXObservationHeaderDetails struct {
	Version                 RINEXHeaderFloat64                  `json:"version"`
	ApproxPositionM         *[3]RINEXHeaderFloat64              `json:"approx_position_m"`
	AntennaDeltaHENM        *[3]RINEXHeaderFloat64              `json:"antenna_delta_hen_m"`
	ObsCodes                map[string][]string                 `json:"obs_codes"`
	DeclaredObsCodes        map[string][]string                 `json:"declared_obs_codes"`
	RINEX2Types             []string                            `json:"rinex2_types"`
	RINEX2System            *string                             `json:"rinex2_system"`
	ProgramRunByDate        *RINEXProgramRunByDate              `json:"program_run_by_date"`
	Comments                []string                            `json:"comments"`
	MarkerNumber            *string                             `json:"marker_number"`
	MarkerType              *string                             `json:"marker_type"`
	Observer                *string                             `json:"observer"`
	Agency                  *string                             `json:"agency"`
	Receiver                *RINEXReceiverInfo                  `json:"receiver"`
	Antenna                 *RINEXAntennaInfo                   `json:"antenna"`
	IntervalS               *RINEXHeaderFloat64                 `json:"interval_s"`
	TimeOfFirstObs          *RINEXHeaderEpochWithScale          `json:"time_of_first_obs"`
	TimeOfLastObs           *RINEXHeaderEpochWithScale          `json:"time_of_last_obs"`
	NSatellites             *uint64                             `json:"n_satellites"`
	PRNObsCounts            map[string][]*uint64                `json:"prn_obs_counts"`
	PhaseShifts             []RINEXObservationPhaseShift        `json:"phase_shifts"`
	ScaleFactors            []RINEXObservationScaleFactor       `json:"scale_factors"`
	GLONASSSlots            map[string]int8                     `json:"glonass_slots"`
	GLONASSCodePhaseEntries *[]RINEXObservationGLONASSCodePhase `json:"glonass_cod_phs_bis"`
	SignalStrengthUnit      *string                             `json:"signal_strength_unit"`
	LeapSeconds             *RINEXObservationLeapSeconds        `json:"leap_seconds"`
	MarkerName              *string                             `json:"marker_name"`
	UnretainedHeaderLabels  []string                            `json:"unretained_header_labels"`
}

// RINEXProgramRunByDate contains one PGM / RUN BY / DATE record.
type RINEXProgramRunByDate struct {
	Program string `json:"program"`
	RunBy   string `json:"run_by"`
	Date    string `json:"date"`
}

// RINEXReceiverInfo contains one receiver serial, type, and version record.
type RINEXReceiverInfo struct {
	Number       string `json:"number"`
	ReceiverType string `json:"receiver_type"`
	Version      string `json:"version"`
}

// RINEXAntennaInfo contains one antenna serial and type record.
type RINEXAntennaInfo struct {
	Number      string `json:"number"`
	AntennaType string `json:"antenna_type"`
}

// RINEXHeaderEpoch is a civil epoch in the observation file's time scale.
type RINEXHeaderEpoch struct {
	Year   int32              `json:"year"`
	Month  uint8              `json:"month"`
	Day    uint8              `json:"day"`
	Hour   uint8              `json:"hour"`
	Minute uint8              `json:"minute"`
	Second RINEXHeaderFloat64 `json:"second"`
}

// RINEXHeaderEpochWithScale pairs an observation header epoch with its scale.
type RINEXHeaderEpochWithScale struct {
	Epoch     RINEXHeaderEpoch `json:"epoch"`
	TimeScale string           `json:"time_scale"`
}

// RINEXObservationPhaseShift retains one SYS / PHASE SHIFT record.
type RINEXObservationPhaseShift struct {
	System                    string              `json:"system"`
	Code                      *string             `json:"code"`
	CorrectionCycles          *RINEXHeaderFloat64 `json:"correction_cycles"`
	Satellites                []string            `json:"satellites"`
	UnrepresentableSatellites []string            `json:"unrepresentable_satellites"`
}

// RINEXObservationScaleFactor retains one SYS / SCALE FACTOR record.
type RINEXObservationScaleFactor struct {
	System string             `json:"system"`
	Factor RINEXHeaderFloat64 `json:"factor"`
	Codes  []string           `json:"codes"`
}

// RINEXObservationGLONASSCodePhase is one raw GLONASS COD/PHS/BIS entry.
type RINEXObservationGLONASSCodePhase struct {
	Code string              `json:"code"`
	Bias *RINEXHeaderFloat64 `json:"bias"`
}

// MarshalJSON preserves the core's ordered [code, bias] tuple representation.
func (entry RINEXObservationGLONASSCodePhase) MarshalJSON() ([]byte, error) {
	var bias any
	if entry.Bias != nil {
		bias = *entry.Bias
	}
	return json.Marshal([2]any{entry.Code, bias})
}

// UnmarshalJSON retains the core's ordered [code, bias] tuple representation.
func (entry *RINEXObservationGLONASSCodePhase) UnmarshalJSON(data []byte) error {
	var pair []json.RawMessage
	if err := json.Unmarshal(data, &pair); err != nil {
		return err
	}
	if len(pair) != 2 {
		return fmt.Errorf("sidereon: GLONASS code-phase entry must contain two values")
	}
	if err := json.Unmarshal(pair[0], &entry.Code); err != nil {
		return err
	}
	if bytes.Equal(pair[1], []byte("null")) {
		entry.Bias = nil
		return nil
	}
	var bias RINEXHeaderFloat64
	if err := json.Unmarshal(pair[1], &bias); err != nil {
		return err
	}
	entry.Bias = &bias
	return nil
}

// RINEXObservationLeapSeconds retains one LEAP SECONDS record.
type RINEXObservationLeapSeconds struct {
	Current     int64   `json:"current"`
	DeltaFuture *int64  `json:"delta_future"`
	Week        *int64  `json:"week"`
	Day         *int64  `json:"day"`
	TimeSystem  *string `json:"time_system"`
}

// RINEXCorrectionUnavailableError reports an unknown or ambiguous header bias.
type RINEXCorrectionUnavailableError struct {
	Kind        string
	Corrections []*RINEXHeaderFloat64
}

// Error formats the correction-unavailable category.
func (err *RINEXCorrectionUnavailableError) Error() string {
	if err == nil {
		return "sidereon: correction unavailable"
	}
	return fmt.Sprintf("sidereon: correction unavailable (%s)", err.Kind)
}

// GLONASSCodePhaseBias applies the core header lookup semantics to the copied
// GLONASS COD/PHS/BIS entries. A false presence result means the header states
// no correction; unknown and conflicting blank/nonblank values are typed errors.
func (header RINEXObservationHeaderDetails) GLONASSCodePhaseBias(code string) (RINEXHeaderFloat64, bool, error) {
	if header.Version.Float64() >= 4 || header.GLONASSCodePhaseEntries == nil {
		return 0, false, nil
	}
	entries := *header.GLONASSCodePhaseEntries
	if len(entries) == 0 {
		return 0, false, &RINEXCorrectionUnavailableError{Kind: "Unknown"}
	}
	var distinct []*RINEXHeaderFloat64
	for _, entry := range entries {
		if entry.Code != code {
			continue
		}
		seen := false
		for _, prior := range distinct {
			if prior == nil && entry.Bias == nil || prior != nil && entry.Bias != nil && *prior == *entry.Bias {
				seen = true
				break
			}
		}
		if !seen {
			distinct = append(distinct, entry.Bias)
		}
	}
	switch len(distinct) {
	case 0:
		return 0, false, nil
	case 1:
		if distinct[0] == nil {
			return 0, false, &RINEXCorrectionUnavailableError{Kind: "Unknown"}
		}
		return *distinct[0], true, nil
	default:
		return 0, false, &RINEXCorrectionUnavailableError{Kind: "Ambiguous", Corrections: distinct}
	}
}

// HeaderDetails returns complete copied headers for the file and each effective
// event segment. Integer fields decode without float64 conversion.
func (obs *RINEXObservation) HeaderDetails() ([]RINEXObservationHeaderSegmentDetails, error) {
	if obs == nil || obs.handle == nil {
		return nil, ErrClosed
	}
	data, err := obs.handle.HeaderDetailsJSON()
	if err != nil {
		return nil, publicError(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var payload struct {
		Segments []RINEXObservationHeaderSegmentDetails `json:"segments"`
	}
	if err := decoder.Decode(&payload); err != nil {
		return nil, fmt.Errorf("sidereon: decode observation header details: %w", err)
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		return nil, errors.New("sidereon: trailing observation header details JSON")
	}
	return payload.Segments, nil
}
