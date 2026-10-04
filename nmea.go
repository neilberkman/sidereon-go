package sidereon

import (
	"encoding/json"
	"errors"

	"sidereon.dev/go/v3/internal/native"
)

// NMEATalker is the exact parsed talker identifier.
type NMEATalker struct {
	Kind   string  `json:"kind"`
	System string  `json:"system,omitempty"`
	Code   string  `json:"code"`
	Bytes  []uint8 `json:"bytes,omitempty"`
}

// NMEATime retains parsed clock fields and fractional precision.
type NMEATime struct {
	Hour     uint8  `json:"hour"`
	Minute   uint8  `json:"minute"`
	Second   uint8  `json:"second"`
	Nanos    uint32 `json:"nanos"`
	Decimals uint8  `json:"decimals"`
}

// NMEADate is a validated calendar date.
type NMEADate struct {
	Year  uint16 `json:"year"`
	Month uint8  `json:"month"`
	Day   uint8  `json:"day"`
}

// NMEACoordinate retains the source coordinate integer and decimal value.
type NMEACoordinate struct {
	Degrees       uint16      `json:"degrees"`
	MinutesScaled json.Number `json:"minutes_scaled"`
	Decimals      uint8       `json:"decimals"`
	Negative      bool        `json:"negative"`
	DegreesF64    float64     `json:"degrees_f64"`
}

// NMEAQuality is the typed GGA fix-quality code.
type NMEAQuality struct {
	Kind  string `json:"kind"`
	Value uint8  `json:"value"`
}

// NMEASatelliteNumber retains a raw satellite number and optional resolved ID.
type NMEASatelliteNumber struct {
	Raw      uint16  `json:"raw"`
	Resolved *string `json:"resolved"`
}

// NMEASignal retains the parsed signal code and optional carrier band.
type NMEASignal struct {
	System      *string `json:"system"`
	ID          uint8   `json:"id"`
	CarrierBand *string `json:"carrier_band"`
}

// NMEAStatus is a typed RMC or GLL status value.
type NMEAStatus struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

// NMEAGGA contains all typed GGA sentence fields.
type NMEAGGA struct {
	Kind                  string          `json:"kind"`
	Time                  *NMEATime       `json:"time"`
	Latitude              *NMEACoordinate `json:"latitude"`
	Longitude             *NMEACoordinate `json:"longitude"`
	Quality               *NMEAQuality    `json:"quality"`
	SatellitesUsed        *uint8          `json:"satellites_used"`
	HDOP                  *json.Number    `json:"hdop"`
	AltitudeMSLM          *json.Number    `json:"altitude_msl_m"`
	GeoidSeparationM      *json.Number    `json:"geoid_separation_m"`
	DifferentialAgeS      *json.Number    `json:"differential_age_s"`
	DifferentialStationID *uint16         `json:"differential_station_id"`
}

// NMEARMC contains all typed RMC sentence fields.
type NMEARMC struct {
	Kind                 string          `json:"kind"`
	Time                 *NMEATime       `json:"time"`
	Status               *NMEAStatus     `json:"status"`
	Latitude             *NMEACoordinate `json:"latitude"`
	Longitude            *NMEACoordinate `json:"longitude"`
	SpeedOverGroundKn    *json.Number    `json:"speed_over_ground_kn"`
	CourseOverGroundDeg  *json.Number    `json:"course_over_ground_deg"`
	Date                 *NMEADate       `json:"date"`
	MagneticVariationDeg *json.Number    `json:"magnetic_variation_deg"`
	FAAMode              *string         `json:"faa_mode"`
	NavigationalStatus   *string         `json:"navigational_status"`
}

// NMEAGSA contains all typed GSA sentence fields.
type NMEAGSA struct {
	Kind          string      `json:"kind"`
	SelectionMode *NMEAStatus `json:"selection_mode"`
	FixMode       *struct {
		Kind  string      `json:"kind"`
		Value json.Number `json:"value"`
	} `json:"fix_mode"`
	Satellites []NMEASatelliteNumber `json:"satellites"`
	PDOP       *json.Number          `json:"pdop"`
	HDOP       *json.Number          `json:"hdop"`
	VDOP       *json.Number          `json:"vdop"`
	SystemID   *uint8                `json:"system_id"`
	System     *string               `json:"system"`
}

// NMEAGSVSatellite contains one optional four-field GSV slot.
type NMEAGSVSatellite struct {
	SatNumber    *NMEASatelliteNumber `json:"sat_number"`
	ElevationDeg *int16               `json:"elevation_deg"`
	AzimuthDeg   *uint16              `json:"azimuth_deg"`
	CN0DBHz      *uint8               `json:"cn0_db_hz"`
}

// NMEAGSV contains the typed fields from one GSV page.
type NMEAGSV struct {
	Kind             string             `json:"kind"`
	TotalMessages    uint8              `json:"total_messages"`
	MessageNumber    uint8              `json:"message_number"`
	SatellitesInView *uint16            `json:"satellites_in_view"`
	Satellites       []NMEAGSVSatellite `json:"satellites"`
	Signal           *NMEASignal        `json:"signal"`
}

// NMEAGST contains all typed GST sentence fields.
type NMEAGST struct {
	Kind              string       `json:"kind"`
	Time              *NMEATime    `json:"time"`
	RMSRangeResidualM *json.Number `json:"rms_range_residual_m"`
	SemiMajorErrorM   *json.Number `json:"semi_major_error_m"`
	SemiMinorErrorM   *json.Number `json:"semi_minor_error_m"`
	OrientationDeg    *json.Number `json:"orientation_deg"`
	LatitudeSigmaM    *json.Number `json:"latitude_sigma_m"`
	LongitudeSigmaM   *json.Number `json:"longitude_sigma_m"`
	AltitudeSigmaM    *json.Number `json:"altitude_sigma_m"`
}

// NMEAVTG contains all typed VTG sentence fields.
type NMEAVTG struct {
	Kind              string       `json:"kind"`
	CourseTrueDeg     *json.Number `json:"course_true_deg"`
	CourseMagneticDeg *json.Number `json:"course_magnetic_deg"`
	SpeedKn           *json.Number `json:"speed_kn"`
	SpeedKMH          *json.Number `json:"speed_kmh"`
	FAAMode           *string      `json:"faa_mode"`
}

// NMEAGLL contains all typed GLL sentence fields.
type NMEAGLL struct {
	Kind      string          `json:"kind"`
	Latitude  *NMEACoordinate `json:"latitude"`
	Longitude *NMEACoordinate `json:"longitude"`
	Time      *NMEATime       `json:"time"`
	Status    *NMEAStatus     `json:"status"`
	FAAMode   *string         `json:"faa_mode"`
}

// NMEAZDA contains all typed ZDA sentence fields.
type NMEAZDA struct {
	Kind             string    `json:"kind"`
	Time             *NMEATime `json:"time"`
	Date             *NMEADate `json:"date"`
	LocalZoneHours   *int8     `json:"local_zone_hours"`
	LocalZoneMinutes *int8     `json:"local_zone_minutes"`
}

// NMEASentenceBody is a tagged union. Exactly one variant pointer is populated.
type NMEASentenceBody struct {
	Kind string          `json:"kind"`
	GGA  *NMEAGGA        `json:"-"`
	RMC  *NMEARMC        `json:"-"`
	GSA  *NMEAGSA        `json:"-"`
	GSV  *NMEAGSV        `json:"-"`
	GST  *NMEAGST        `json:"-"`
	VTG  *NMEAVTG        `json:"-"`
	GLL  *NMEAGLL        `json:"-"`
	ZDA  *NMEAZDA        `json:"-"`
	Raw  json.RawMessage `json:"-"`
}

// NMEASentenceRecord retains one accepted sentence with its typed body.
type NMEASentenceRecord struct {
	Kind   string           `json:"kind"`
	Index  uint64           `json:"index"`
	Talker NMEATalker       `json:"talker"`
	Body   NMEASentenceBody `json:"body"`
	Raw    json.RawMessage  `json:"-"`
}

// NMEAEpochRecord contains all accepted sentence fields for one epoch.
type NMEAEpochRecord struct {
	Kind             string         `json:"kind"`
	Index            uint64         `json:"index"`
	Time             *NMEATime      `json:"time"`
	Date             *NMEADate      `json:"date"`
	GGA              *NMEAGGA       `json:"gga"`
	RMC              *NMEARMC       `json:"rmc"`
	GLL              *NMEAGLL       `json:"gll"`
	GST              *NMEAGST       `json:"gst"`
	VTG              *NMEAVTG       `json:"vtg"`
	ZDA              *NMEAZDA       `json:"zda"`
	GSA              []NMEAEpochGSA `json:"gsa"`
	GSV              []NMEAEpochGSV `json:"gsv"`
	SentenceCount    uint64         `json:"sentence_count"`
	DiagnosticCounts struct {
		Skips    uint64 `json:"skips"`
		Warnings uint64 `json:"warnings"`
	} `json:"diagnostic_counts"`
	Raw json.RawMessage `json:"-"`
}

// NMEAEpochGSA retains one grouped GSA body and system context.
type NMEAEpochGSA struct {
	System *string `json:"system"`
	Body   NMEAGSA `json:"body"`
}

// NMEAEpochGSV retains one grouped GSV view and its complete satellite slots.
type NMEAEpochGSV struct {
	Talker        NMEATalker         `json:"talker"`
	Signal        *NMEASignal        `json:"signal"`
	ClaimedInView *uint16            `json:"claimed_in_view"`
	Complete      bool               `json:"complete"`
	Satellites    []NMEAGSVSatellite `json:"satellites"`
}

// NMEADiagnosticSource identifies parser or epoch assembly diagnostics.
type NMEADiagnosticSource uint32

const (
	// NMEADiagnosticParser marks a parser-level diagnostic.
	NMEADiagnosticParser NMEADiagnosticSource = iota
	// NMEADiagnosticEpochAssembly marks an epoch assembly diagnostic.
	NMEADiagnosticEpochAssembly
)

// NMEADiagnosticKind identifies a skipped record or warning.
type NMEADiagnosticKind uint32

const (
	// NMEADiagnosticSkip marks a skipped input record.
	NMEADiagnosticSkip NMEADiagnosticKind = iota
	// NMEADiagnosticWarning marks a nonfatal parser or assembly warning.
	NMEADiagnosticWarning
)

// NMEARecordReference preserves optional input line and record indices.
type NMEARecordReference struct {
	Line        *uint64 `json:"line"`
	RecordIndex *uint64 `json:"record_index"`
	Satellite   *string `json:"satellite"`
}

// NMEAFieldError retains a typed field-validation refusal.
type NMEAFieldError struct {
	Kind   string `json:"kind"`
	Fields struct {
		Field          *string      `json:"field"`
		Reason         *string      `json:"reason"`
		Min            *EngineFloat `json:"min"`
		Max            *EngineFloat `json:"max"`
		UpperInclusive *bool        `json:"upper_inclusive"`
		Value          *string      `json:"value"`
		Year           *int64       `json:"year"`
		Month          *int64       `json:"month"`
		Day            *int64       `json:"day"`
		Hour           *int64       `json:"hour"`
		Minute         *int64       `json:"minute"`
		Second         *EngineFloat `json:"second"`
	} `json:"fields"`
	Raw json.RawMessage `json:"-"`
}

// NMEASkipReasonFields contains variant-specific skip data.
type NMEASkipReasonFields struct {
	RecordType *string         `json:"record_type"`
	Cause      *NMEAFieldError `json:"cause"`
	Unit       *string         `json:"unit"`
	Block      *string         `json:"block"`
	Reason     *string         `json:"reason"`
}

// NMEASkipReason retains the typed parser skip variant.
type NMEASkipReason struct {
	Kind   string               `json:"kind"`
	Fields NMEASkipReasonFields `json:"fields"`
	Raw    json.RawMessage      `json:"-"`
}

// NMEASkipDiagnosticFields groups the input reference and reason.
type NMEASkipDiagnosticFields struct {
	At     NMEARecordReference `json:"at"`
	Reason NMEASkipReason      `json:"reason"`
}

// NMEASkipDiagnostic retains one typed parser skip.
type NMEASkipDiagnostic struct {
	Kind   string                   `json:"kind"`
	Fields NMEASkipDiagnosticFields `json:"fields"`
	Raw    json.RawMessage          `json:"-"`
}

// NMEAWarningDiagnosticFields groups the input reference and warning kind.
type NMEAWarningDiagnosticFields struct {
	At          NMEARecordReference `json:"at"`
	WarningKind string              `json:"warning_kind"`
}

// NMEAWarningDiagnostic retains one typed parser or assembly warning.
type NMEAWarningDiagnostic struct {
	Kind   string                      `json:"kind"`
	Fields NMEAWarningDiagnosticFields `json:"fields"`
	Raw    json.RawMessage             `json:"-"`
}

// NMEADiagnostic retains the typed scope, warning or skip value, and exact payload.
type NMEADiagnostic struct {
	Source        NMEADiagnosticSource   `json:"source"`
	Kind          NMEADiagnosticKind     `json:"kind"`
	HasEpochIndex bool                   `json:"has_epoch_index"`
	EpochIndex    uint64                 `json:"epoch_index"`
	Skip          *NMEASkipDiagnostic    `json:"skip,omitempty"`
	Warning       *NMEAWarningDiagnostic `json:"warning,omitempty"`
	Payload       json.RawMessage        `json:"payload"`
	DecodeError   error                  `json:"-"`
}

// UnmarshalJSON decodes JSON while preserving the typed field representation.
func (body *NMEASentenceBody) UnmarshalJSON(data []byte) error {
	var head struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(data, &head); err != nil {
		return err
	}
	*body = NMEASentenceBody{}
	body.Kind = head.Kind
	body.Raw = append(json.RawMessage(nil), data...)
	switch head.Kind {
	case "gga":
		body.GGA = new(NMEAGGA)
		return json.Unmarshal(data, body.GGA)
	case "rmc":
		body.RMC = new(NMEARMC)
		return json.Unmarshal(data, body.RMC)
	case "gsa":
		body.GSA = new(NMEAGSA)
		return json.Unmarshal(data, body.GSA)
	case "gsv":
		body.GSV = new(NMEAGSV)
		return json.Unmarshal(data, body.GSV)
	case "gst":
		body.GST = new(NMEAGST)
		return json.Unmarshal(data, body.GST)
	case "vtg":
		body.VTG = new(NMEAVTG)
		return json.Unmarshal(data, body.VTG)
	case "gll":
		body.GLL = new(NMEAGLL)
		return json.Unmarshal(data, body.GLL)
	case "zda":
		body.ZDA = new(NMEAZDA)
		return json.Unmarshal(data, body.ZDA)
	default:
		return nil
	}
}

// MarshalJSON encodes this value as JSON.
func (body NMEASentenceBody) MarshalJSON() ([]byte, error) {
	switch body.Kind {
	case "gga":
		if body.GGA != nil {
			return json.Marshal(body.GGA)
		}
	case "rmc":
		if body.RMC != nil {
			return json.Marshal(body.RMC)
		}
	case "gsa":
		if body.GSA != nil {
			return json.Marshal(body.GSA)
		}
	case "gsv":
		if body.GSV != nil {
			return json.Marshal(body.GSV)
		}
	case "gst":
		if body.GST != nil {
			return json.Marshal(body.GST)
		}
	case "vtg":
		if body.VTG != nil {
			return json.Marshal(body.VTG)
		}
	case "gll":
		if body.GLL != nil {
			return json.Marshal(body.GLL)
		}
	case "zda":
		if body.ZDA != nil {
			return json.Marshal(body.ZDA)
		}
	}
	if len(body.Raw) != 0 {
		return append([]byte(nil), body.Raw...), nil
	}
	return nil, errors.New("sidereon: NMEA sentence body has no matching variant")
}

// NMEASummary contains aggregate counts detached from a parsed NMEA log.
type NMEASummary struct {
	// SentenceCount is the number of NMEA sentences parsed.
	SentenceCount uint64
	// EpochCount is the number of epoch summaries produced.
	EpochCount uint64
	// SkipCount is the number of malformed or skipped sentences.
	SkipCount uint64
	// WarningCount is the number of parser warnings.
	WarningCount uint64
}

// NMEAEpoch is one detached epoch summary from a parsed NMEA log.
type NMEAEpoch struct {
	// HasCalendarEpoch reports whether CalendarEpoch contains a parsed UTC epoch.
	HasCalendarEpoch bool
	// CalendarEpoch is the calendar epoch.
	CalendarEpoch CalendarEpoch
	// HasPosition reports whether Position contains a parsed latitude/longitude fix.
	HasPosition bool
	// LatitudeRad is the latitude rad in radians.
	LatitudeRad float64
	// LongitudeRad is the longitude rad in radians.
	LongitudeRad float64
	// HeightM is the height m in metres.
	HeightM float64
	// HasInstantJ2000S reports whether InstantJ2000S contains a parsed epoch.
	HasInstantJ2000S bool
	// InstantJ2000S is the parsed epoch in seconds from J2000.
	InstantJ2000S float64
	// HasPDOP reports whether PDOP is populated from a dilution report.
	HasPDOP bool
	// PDOP is the dimensionless position dilution of precision when HasPDOP is true.
	PDOP float64
	// HasHDOP reports whether HDOP is populated from a dilution report.
	HasHDOP bool
	// HDOP is the dimensionless horizontal dilution of precision when HasHDOP is true.
	HDOP float64
	// HasVDOP reports whether VDOP is populated from a dilution report.
	HasVDOP bool
	// VDOP is the dimensionless vertical dilution of precision when HasVDOP is true.
	VDOP float64
	// SentenceCount is the number of sentences contributing to this epoch.
	SentenceCount uint64
	// UsedSatelliteCount is the number of satellites used in this epoch.
	UsedSatelliteCount uint64
	// SatellitesInView is the number of satellites reported in view.
	SatellitesInView uint64
	// SkipCount is the number of skipped sentences in this epoch.
	SkipCount uint64
	// WarningCount is the number of warnings associated with this epoch.
	WarningCount uint64
	// HasGGA reports whether a GGA sentence was parsed for the epoch.
	HasGGA bool
	// HasRMC reports whether an RMC sentence was parsed for the epoch.
	HasRMC bool
	// HasGLL reports whether a GLL sentence was parsed for the epoch.
	HasGLL bool
	// GSACount is the number of GSA sentences in this epoch.
	GSACount uint64
	// GSVGroupCount is the number of GSV groups in this epoch.
	GSVGroupCount uint64
}

// NMEAChunkSummary contains the output counts from one incremental Push or
// Finish call. RetainedLength is the number of partial-line bytes held for a
// later chunk.
type NMEAChunkSummary struct {
	// SentenceCount is the number of sentences consumed by the call.
	SentenceCount uint64
	// CompletedEpochCount is the number of complete epochs emitted by the call.
	CompletedEpochCount uint64
	// SkipCount is the number of skipped sentences reported by the call.
	SkipCount uint64
	// WarningCount is the number of parser warnings reported by the call.
	WarningCount uint64
	// RetainedLength is the number of bytes held from an incomplete trailing line.
	RetainedLength uint64
}

// NMEAGGAOptions contains validated fields for one NMEA 0183 GGA sentence.
// Position is WGS84 geodetic radians/metres and UTCSecondsOfDay is rounded
// down to centisecond precision by the engine.
type NMEAGGAOptions struct {
	// Talker is the NMEA talker identifier.
	Talker string
	// UTCSecondsOfDay is the utc seconds of day in seconds since midnight.
	UTCSecondsOfDay float64
	// Position is the position value in the containing frame.
	Position Geodetic
	// Quality is the NMEA GGA fix-quality code.
	Quality uint32
	// SatellitesUsed is the number of satellites used in the GGA solution.
	SatellitesUsed uint8
	// HDOP is the dimensionless horizontal dilution of precision in the GGA sentence.
	HDOP float64
	// CoordinateDecimals is the number of decimal places emitted for coordinates.
	CoordinateDecimals uint8
}

// DefaultNMEAGGAOptions returns the sibling-interface VRS defaults. The
// caller supplies Position and UTCSecondsOfDay before calling WriteNMEAGGA.
func DefaultNMEAGGAOptions() NMEAGGAOptions {
	return NMEAGGAOptions{Talker: "GP", Quality: 1, SatellitesUsed: 10, HDOP: 1, CoordinateDecimals: 7}
}

// NMEAAccumulator owns an incremental native parser. Mutating Push and Finish
// calls are serialized with read-only Summary, Epochs, RetainedLength, and
// Close calls, so the handle may be shared by goroutines.
type NMEAAccumulator struct {
	_      noCopy
	handle *native.NMEAAccumulator
}

// NewNMEAAccumulator creates an empty incremental NMEA parser.
func NewNMEAAccumulator() (*NMEAAccumulator, error) {
	handle, err := native.NewNMEAAccumulator()
	if err != nil {
		return nil, publicError(err)
	}
	if handle == nil {
		return nil, errNilNativeHandle
	}
	return &NMEAAccumulator{handle: handle}, nil
}

// Close releases the native accumulator. It is idempotent.
func (accumulator *NMEAAccumulator) Close() error {
	if accumulator == nil || accumulator.handle == nil {
		return nil
	}
	return publicError(accumulator.handle.Close())
}

// Push parses one arbitrary byte chunk and retains an incomplete trailing
// line for a later call. The input is copied into C-owned memory for the call.
func (accumulator *NMEAAccumulator) Push(data []byte) (NMEAChunkSummary, error) {
	if accumulator == nil || accumulator.handle == nil {
		return NMEAChunkSummary{}, ErrClosed
	}
	value, err := accumulator.handle.Push(data)
	return publicNMEAChunkSummary(value), publicError(err)
}

// Finish flushes the pending epoch after the final byte chunk. A
// CompletedEpochCount of zero means there was no pending epoch.
func (accumulator *NMEAAccumulator) Finish() (NMEAChunkSummary, error) {
	if accumulator == nil || accumulator.handle == nil {
		return NMEAChunkSummary{}, ErrClosed
	}
	value, err := accumulator.handle.Finish()
	return publicNMEAChunkSummary(value), publicError(err)
}

// Summary returns cumulative counts for all chunks accepted so far.
func (accumulator *NMEAAccumulator) Summary() (NMEASummary, error) {
	if accumulator == nil || accumulator.handle == nil {
		return NMEASummary{}, ErrClosed
	}
	value, err := accumulator.handle.Summary()
	return NMEASummary{
		SentenceCount: value.SentenceCount,
		EpochCount:    value.EpochCount,
		SkipCount:     value.SkipCount,
		WarningCount:  value.WarningCount,
	}, publicError(err)
}

// RetainedLength returns the number of partial-line bytes waiting for a later
// Push or Finish call.
func (accumulator *NMEAAccumulator) RetainedLength() (uint64, error) {
	if accumulator == nil || accumulator.handle == nil {
		return 0, ErrClosed
	}
	value, err := accumulator.handle.RetainedLength()
	return value, publicError(err)
}

// Epochs copies every completed accumulated epoch into Go-owned values.
func (accumulator *NMEAAccumulator) Epochs() ([]NMEAEpoch, error) {
	if accumulator == nil || accumulator.handle == nil {
		return nil, ErrClosed
	}
	values, err := accumulator.handle.Epochs()
	if err != nil {
		return nil, publicError(err)
	}
	return publicNMEAEpochs(values), nil
}

// SentenceRecords returns every accepted sentence in input order with one of
// the eight NMEA body variants decoded into its typed record.
func (accumulator *NMEAAccumulator) SentenceRecords() ([]NMEASentenceRecord, error) {
	if accumulator == nil || accumulator.handle == nil {
		return nil, ErrClosed
	}
	values, err := accumulator.handle.SentenceRecords()
	if err != nil {
		return nil, publicError(err)
	}
	return decodeNMEASentences(values)
}

// EpochRecords returns every retained epoch with all singleton, GSA, and GSV
// fields decoded into detached Go values.
func (accumulator *NMEAAccumulator) EpochRecords() ([]NMEAEpochRecord, error) {
	if accumulator == nil || accumulator.handle == nil {
		return nil, ErrClosed
	}
	values, err := accumulator.handle.EpochRecords()
	if err != nil {
		return nil, publicError(err)
	}
	return decodeNMEAEpochs(values)
}

// Diagnostics returns parser and epoch-assembly skips and warnings in native order.
func (accumulator *NMEAAccumulator) Diagnostics() ([]NMEADiagnostic, error) {
	if accumulator == nil || accumulator.handle == nil {
		return nil, ErrClosed
	}
	values, err := accumulator.handle.Diagnostics()
	if err != nil {
		return nil, publicError(err)
	}
	return publicNMEADiagnostics(values), nil
}

// EpochDiagnostics returns detached assembly diagnostics for one retained epoch.
func (accumulator *NMEAAccumulator) EpochDiagnostics(index int) ([]NMEADiagnostic, error) {
	if accumulator == nil || accumulator.handle == nil {
		return nil, ErrClosed
	}
	values, err := accumulator.handle.EpochDiagnostics(index)
	if err != nil {
		return nil, publicError(err)
	}
	return publicNMEADiagnostics(values), nil
}

// WriteNMEAGGA formats one checksummed NMEA 0183 GGA sentence, including its
// CRLF terminator, through the engine's validated writer.
func WriteNMEAGGA(options NMEAGGAOptions) ([]byte, error) {
	value, err := native.WriteNMEAGGA(native.NMEAGGAOptions{
		Talker:             options.Talker,
		UTCSecondsOfDay:    options.UTCSecondsOfDay,
		Position:           native.Geodetic{LatitudeRad: options.Position.LatitudeRad, LongitudeRad: options.Position.LongitudeRad, HeightM: options.Position.HeightM},
		Quality:            options.Quality,
		SatellitesUsed:     options.SatellitesUsed,
		HDOP:               options.HDOP,
		CoordinateDecimals: options.CoordinateDecimals,
	})
	return value, publicError(err)
}

func publicNMEAChunkSummary(value native.NMEAChunkSummary) NMEAChunkSummary {
	return NMEAChunkSummary{
		SentenceCount:       value.SentenceCount,
		CompletedEpochCount: value.CompletedEpochCount,
		SkipCount:           value.SkipCount,
		WarningCount:        value.WarningCount,
		RetainedLength:      value.RetainedLength,
	}
}

// NMEALog owns a parsed C NMEA log. A log may be shared by goroutines for
// read-only Summary and Epochs calls. Close may run concurrently with reads;
// it waits for an active read, clears the native pointer, and is idempotent.
type NMEALog struct {
	_      noCopy
	handle *native.NMEALog
}

// ParseNMEA parses data through the C ABI and returns an owning log handle.
// The input is copied only for the duration of the C call; the C library does
// not retain a Go pointer.
func ParseNMEA(data []byte) (*NMEALog, error) {
	handle, err := native.ParseNMEA(data)
	if err != nil {
		return nil, publicError(err)
	}
	if handle == nil {
		return nil, errNilNativeHandle
	}
	return &NMEALog{handle: handle}, nil
}

// Close releases the native log. It is safe to call more than once and from a
// goroutine concurrent with read-only methods.
func (l *NMEALog) Close() error {
	if l == nil || l.handle == nil {
		return nil
	}
	err := l.handle.Close()
	return publicError(err)
}

// Summary copies aggregate log counts into Go-owned values.
func (l *NMEALog) Summary() (NMEASummary, error) {
	if l == nil || l.handle == nil {
		return NMEASummary{}, ErrClosed
	}
	summary, err := l.handle.Summary()
	return NMEASummary{
		SentenceCount: uint64(summary.SentenceCount),
		EpochCount:    uint64(summary.EpochCount),
		SkipCount:     uint64(summary.SkipCount),
		WarningCount:  uint64(summary.WarningCount),
	}, publicError(err)
}

// Epochs performs the C size query and copy, returning Go-owned summaries.
func (l *NMEALog) Epochs() ([]NMEAEpoch, error) {
	if l == nil || l.handle == nil {
		return nil, ErrClosed
	}
	epochs, err := l.handle.Epochs()
	if err != nil {
		return nil, publicError(err)
	}
	return publicNMEAEpochs(epochs), nil
}

// SentenceRecords returns all accepted sentence bodies in input order.
func (l *NMEALog) SentenceRecords() ([]NMEASentenceRecord, error) {
	if l == nil || l.handle == nil {
		return nil, ErrClosed
	}
	values, err := l.handle.SentenceRecords()
	if err != nil {
		return nil, publicError(err)
	}
	return decodeNMEASentences(values)
}

// EpochRecords returns all singleton, GSA, and GSV fields for each log epoch.
func (l *NMEALog) EpochRecords() ([]NMEAEpochRecord, error) {
	if l == nil || l.handle == nil {
		return nil, ErrClosed
	}
	values, err := l.handle.EpochRecords()
	if err != nil {
		return nil, publicError(err)
	}
	return decodeNMEAEpochs(values)
}

// Diagnostics returns parser and epoch-assembly skips and warnings in native order.
func (l *NMEALog) Diagnostics() ([]NMEADiagnostic, error) {
	if l == nil || l.handle == nil {
		return nil, ErrClosed
	}
	values, err := l.handle.Diagnostics()
	if err != nil {
		return nil, publicError(err)
	}
	return publicNMEADiagnostics(values), nil
}

// EpochDiagnostics returns detached assembly diagnostics for one log epoch.
func (l *NMEALog) EpochDiagnostics(index int) ([]NMEADiagnostic, error) {
	if l == nil || l.handle == nil {
		return nil, ErrClosed
	}
	values, err := l.handle.EpochDiagnostics(index)
	if err != nil {
		return nil, publicError(err)
	}
	return publicNMEADiagnostics(values), nil
}

func decodeNMEASentences(values [][]byte) ([]NMEASentenceRecord, error) {
	out := make([]NMEASentenceRecord, len(values))
	for i, raw := range values {
		if err := json.Unmarshal(raw, &out[i]); err != nil {
			return nil, err
		}
		out[i].Raw = append(json.RawMessage(nil), raw...)
	}
	return out, nil
}

func decodeNMEAEpochs(values [][]byte) ([]NMEAEpochRecord, error) {
	out := make([]NMEAEpochRecord, len(values))
	for i, raw := range values {
		if err := json.Unmarshal(raw, &out[i]); err != nil {
			return nil, err
		}
		out[i].Raw = append(json.RawMessage(nil), raw...)
	}
	return out, nil
}

func publicNMEADiagnostics(values []native.NMEADiagnostic) []NMEADiagnostic {
	out := make([]NMEADiagnostic, len(values))
	for i, value := range values {
		out[i] = NMEADiagnostic{Source: NMEADiagnosticSource(value.Info.Source), Kind: NMEADiagnosticKind(value.Info.Kind), HasEpochIndex: value.Info.HasEpochIndex, EpochIndex: value.Info.EpochIndex, Payload: append(json.RawMessage(nil), value.Payload...)}
		switch out[i].Kind {
		case NMEADiagnosticSkip:
			out[i].Skip = new(NMEASkipDiagnostic)
			out[i].DecodeError = json.Unmarshal(value.Payload, out[i].Skip)
		case NMEADiagnosticWarning:
			out[i].Warning = new(NMEAWarningDiagnostic)
			out[i].DecodeError = json.Unmarshal(value.Payload, out[i].Warning)
		}
	}
	return out
}

func publicNMEAEpochs(values []native.NMEAEpoch) []NMEAEpoch {
	out := make([]NMEAEpoch, len(values))
	for i, epoch := range values {
		out[i] = NMEAEpoch{
			HasCalendarEpoch:   epoch.HasCalendarEpoch,
			CalendarEpoch:      calendarEpoch(epoch.CalendarEpoch),
			HasPosition:        epoch.HasPosition,
			LatitudeRad:        epoch.LatitudeRad,
			LongitudeRad:       epoch.LongitudeRad,
			HeightM:            epoch.HeightM,
			HasInstantJ2000S:   epoch.HasInstantJ2000S,
			InstantJ2000S:      epoch.InstantJ2000S,
			HasPDOP:            epoch.HasPDOP,
			PDOP:               epoch.PDOP,
			HasHDOP:            epoch.HasHDOP,
			HDOP:               epoch.HDOP,
			HasVDOP:            epoch.HasVDOP,
			VDOP:               epoch.VDOP,
			SentenceCount:      epoch.SentenceCount,
			UsedSatelliteCount: epoch.UsedSatelliteCount,
			SatellitesInView:   epoch.SatellitesInView,
			SkipCount:          epoch.SkipCount,
			WarningCount:       epoch.WarningCount,
			HasGGA:             epoch.HasGGA,
			HasRMC:             epoch.HasRMC,
			HasGLL:             epoch.HasGLL,
			GSACount:           epoch.GSACount,
			GSVGroupCount:      epoch.GSVGroupCount,
		}
	}
	return out
}
