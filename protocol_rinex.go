package sidereon

import (
	"errors"

	"sidereon.dev/go/v3/internal/native"
)

// RINEXClockInstantRepresentation identifies how a clock instant is stored.
type RINEXClockInstantRepresentation uint32

const (
	// RINEXClockInstantJulianDate stores clock epochs as split Julian dates.
	RINEXClockInstantJulianDate RINEXClockInstantRepresentation = RINEXClockInstantRepresentation(native.RINEXClockInstantJulianDateValue)
	// RINEXClockInstantNanos stores clock epochs as nanoseconds from J2000.
	RINEXClockInstantNanos RINEXClockInstantRepresentation = RINEXClockInstantRepresentation(native.RINEXClockInstantNanosValue)
)

// GNSSWeekTow is a scale-tagged GNSS week and time-of-week value.
type GNSSWeekTow struct {
	// System identifies the time scale; Week and TOWSeconds are native GNSS time.
	System     TimeScale
	Week       uint32
	TOWSeconds float64
}

// ClockEpoch preserves the C ABI's lossless clock-epoch representation.
type ClockEpoch struct {
	// Scale and Representation identify the lossless epoch encoding.
	Scale          TimeScale
	Representation RINEXClockInstantRepresentation
	JulianWhole    float64
	JulianFraction float64
	NanosHigh      int64
	NanosLow       uint64
}

// KeplerianElements contains one complete broadcast orbit in SI units and
// radians. C performs all parsing and numerical interpretation.
type KeplerianElements struct {
	// All fields are detached broadcast orbital elements in native engineering units.
	SqrtA    float64
	E        float64
	M0       float64
	DeltaN   float64
	Omega0   float64
	I0       float64
	Omega    float64
	OmegaDot float64
	IDot     float64
	CUC      float64
	CUS      float64
	CRC      float64
	CRS      float64
	CIC      float64
	CIS      float64
	ToeSOW   float64
}

// ClockPolynomial is the broadcast satellite-clock polynomial about TocSOW.
type ClockPolynomial struct {
	// AF0, AF1, and AF2 are native clock polynomial coefficients; TocSOW is seconds of week.
	AF0    float64
	AF1    float64
	AF2    float64
	TocSOW float64
}

// BroadcastGroupDelays contains optional broadcast group-delay values. A
// Has... field distinguishes an absent value from a present zero.
type BroadcastGroupDelays struct {
	// Each Has* flag controls the corresponding optional group-delay coefficient.
	HasGPSTGD          bool
	GPSTGD             float64
	HasGalileoBGDE5AE1 bool
	GalileoBGDE5AE1    float64
	HasGalileoBGDE5BE1 bool
	GalileoBGDE5BE1    float64
	HasBeiDouTGD1      bool
	BeiDouTGD1         float64
	HasBeiDouTGD2      bool
	BeiDouTGD2         float64
	HasCNAVISCL1CA     bool
	CNAVISCL1CA        float64
	HasCNAVISCL2C      bool
	CNAVISCL2C         float64
	HasCNAVISCL5I5     bool
	CNAVISCL5I5        float64
	HasCNAVISCL5Q5     bool
	CNAVISCL5Q5        float64
	HasCNAVISCL1CD     bool
	CNAVISCL1CD        float64
	HasCNAVISCL1CP     bool
	CNAVISCL1CP        float64
}

// BroadcastCNAV contains the optional CNAV/CNAV-2 extension.
type BroadcastCNAV struct {
	// Present controls the optional CNAV fields.
	Present             bool
	ADOTMPerS           float64
	DeltaN0DotRadPerS2  float64
	Top                 GNSSWeekTow
	URAEDIndex          int8
	URANED0Index        int8
	URANED1Index        uint8
	URANED2Index        uint8
	TransmissionTimeSOW float64
	HasFlags            bool
	Flags               uint32
}

// StatedNavFields retains legacy NAV columns that are not part of orbit or
// clock calculations. Each Has* flag distinguishes a blank column from zero.
type StatedNavFields struct {
	HasOrbit5Field2        bool
	Orbit5Field2           float64
	HasOrbit5Field4        bool
	Orbit5Field4           float64
	HasOrbit6Field4        bool
	Orbit6Field4           float64
	HasTransmissionTimeSOW bool
	TransmissionTimeSOW    float64
	HasOrbit7Field2        bool
	Orbit7Field2           float64
	HasOrbit7Field3        bool
	Orbit7Field3           float64
	HasOrbit7Field4        bool
	Orbit7Field4           float64
}

// BroadcastRecord is one complete raw RINEX NAV record. Optional values are
// retained with explicit presence flags, matching the C value model.
type BroadcastRecord struct {
	// SatelliteID and message/issue fields identify the detached broadcast record.
	SatelliteID string
	Message     uint32
	// HasIssue marks a present issue field, including a present zero. For
	// compatibility, encoding treats a nonzero Issue as present even if false.
	HasIssue     bool
	Issue        uint32
	IssueMessage uint32
	Week         uint32
	Toe          GNSSWeekTow
	Toc          GNSSWeekTow
	Elements     KeplerianElements
	Clock        ClockPolynomial
	GroupDelays  BroadcastGroupDelays
	CNAV         BroadcastCNAV
	SVHealth     float64
	// HasSVAccuracyM distinguishes a present zero from unavailable accuracy.
	// Encoding also treats a nonzero SVAccuracyM as present for compatibility.
	HasSVAccuracyM bool
	SVAccuracyM    float64
	HasFitInterval bool
	FitIntervalS   float64
	Stated         StatedNavFields
}

func broadcastRecordFromNative(value native.NativeBroadcastRecord) BroadcastRecord {
	return BroadcastRecord{
		SatelliteID: value.SatelliteID, Message: value.Message, HasIssue: value.HasIssue, Issue: value.Issue, IssueMessage: value.IssueMessage, Week: value.Week,
		Toe: GNSSWeekTow{System: TimeScale(value.Toe.System), Week: value.Toe.Week, TOWSeconds: value.Toe.TOWSeconds},
		Toc: GNSSWeekTow{System: TimeScale(value.Toc.System), Week: value.Toc.Week, TOWSeconds: value.Toc.TOWSeconds},
		Elements: KeplerianElements{
			SqrtA: value.Elements.SqrtA, E: value.Elements.E, M0: value.Elements.M0, DeltaN: value.Elements.DeltaN,
			Omega0: value.Elements.Omega0, I0: value.Elements.I0, Omega: value.Elements.Omega, OmegaDot: value.Elements.OmegaDot,
			IDot: value.Elements.IDot, CUC: value.Elements.CUC, CUS: value.Elements.CUS, CRC: value.Elements.CRC,
			CRS: value.Elements.CRS, CIC: value.Elements.CIC, CIS: value.Elements.CIS, ToeSOW: value.Elements.ToeSOW,
		},
		Clock: ClockPolynomial{AF0: value.Clock.AF0, AF1: value.Clock.AF1, AF2: value.Clock.AF2, TocSOW: value.Clock.TocSOW},
		GroupDelays: BroadcastGroupDelays{
			HasGPSTGD: value.GroupDelays.HasGPSTGD, GPSTGD: value.GroupDelays.GPSTGD,
			HasGalileoBGDE5AE1: value.GroupDelays.HasGalileoBGDE5AE1, GalileoBGDE5AE1: value.GroupDelays.GalileoBGDE5AE1,
			HasGalileoBGDE5BE1: value.GroupDelays.HasGalileoBGDE5BE1, GalileoBGDE5BE1: value.GroupDelays.GalileoBGDE5BE1,
			HasBeiDouTGD1: value.GroupDelays.HasBeiDouTGD1, BeiDouTGD1: value.GroupDelays.BeiDouTGD1,
			HasBeiDouTGD2: value.GroupDelays.HasBeiDouTGD2, BeiDouTGD2: value.GroupDelays.BeiDouTGD2,
			HasCNAVISCL1CA: value.GroupDelays.HasCNAVISCL1CA, CNAVISCL1CA: value.GroupDelays.CNAVISCL1CA,
			HasCNAVISCL2C: value.GroupDelays.HasCNAVISCL2C, CNAVISCL2C: value.GroupDelays.CNAVISCL2C,
			HasCNAVISCL5I5: value.GroupDelays.HasCNAVISCL5I5, CNAVISCL5I5: value.GroupDelays.CNAVISCL5I5,
			HasCNAVISCL5Q5: value.GroupDelays.HasCNAVISCL5Q5, CNAVISCL5Q5: value.GroupDelays.CNAVISCL5Q5,
			HasCNAVISCL1CD: value.GroupDelays.HasCNAVISCL1CD, CNAVISCL1CD: value.GroupDelays.CNAVISCL1CD,
			HasCNAVISCL1CP: value.GroupDelays.HasCNAVISCL1CP, CNAVISCL1CP: value.GroupDelays.CNAVISCL1CP,
		},
		CNAV: BroadcastCNAV{
			Present: value.CNAV.Present, ADOTMPerS: value.CNAV.ADOTMPerS, DeltaN0DotRadPerS2: value.CNAV.DeltaN0DotRadPerS2,
			Top:        GNSSWeekTow{System: TimeScale(value.CNAV.Top.System), Week: value.CNAV.Top.Week, TOWSeconds: value.CNAV.Top.TOWSeconds},
			URAEDIndex: value.CNAV.URAEDIndex, URANED0Index: value.CNAV.URANED0Index, URANED1Index: value.CNAV.URANED1Index,
			URANED2Index: value.CNAV.URANED2Index, TransmissionTimeSOW: value.CNAV.TransmissionTimeSOW,
			HasFlags: value.CNAV.HasFlags, Flags: value.CNAV.Flags,
		},
		SVHealth: value.SVHealth, HasSVAccuracyM: value.HasSVAccuracyM, SVAccuracyM: value.SVAccuracyM, HasFitInterval: value.HasFitInterval, FitIntervalS: value.FitIntervalS,
		Stated: StatedNavFields{
			HasOrbit5Field2: value.Stated.HasOrbit5Field2, Orbit5Field2: value.Stated.Orbit5Field2,
			HasOrbit5Field4: value.Stated.HasOrbit5Field4, Orbit5Field4: value.Stated.Orbit5Field4,
			HasOrbit6Field4: value.Stated.HasOrbit6Field4, Orbit6Field4: value.Stated.Orbit6Field4,
			HasTransmissionTimeSOW: value.Stated.HasTransmissionTimeSOW, TransmissionTimeSOW: value.Stated.TransmissionTimeSOW,
			HasOrbit7Field2: value.Stated.HasOrbit7Field2, Orbit7Field2: value.Stated.Orbit7Field2,
			HasOrbit7Field3: value.Stated.HasOrbit7Field3, Orbit7Field3: value.Stated.Orbit7Field3,
			HasOrbit7Field4: value.Stated.HasOrbit7Field4, Orbit7Field4: value.Stated.Orbit7Field4,
		},
	}
}

func broadcastRecordToNative(value BroadcastRecord) native.NativeBroadcastRecord {
	return native.NativeBroadcastRecord{
		SatelliteID: value.SatelliteID, Message: value.Message, HasIssue: value.HasIssue, Issue: value.Issue, IssueMessage: value.IssueMessage, Week: value.Week,
		Toe: native.NativeGnssWeekTow{System: uint32(value.Toe.System), Week: value.Toe.Week, TOWSeconds: value.Toe.TOWSeconds},
		Toc: native.NativeGnssWeekTow{System: uint32(value.Toc.System), Week: value.Toc.Week, TOWSeconds: value.Toc.TOWSeconds},
		Elements: native.NativeKeplerianElements{
			SqrtA: value.Elements.SqrtA, E: value.Elements.E, M0: value.Elements.M0, DeltaN: value.Elements.DeltaN,
			Omega0: value.Elements.Omega0, I0: value.Elements.I0, Omega: value.Elements.Omega, OmegaDot: value.Elements.OmegaDot,
			IDot: value.Elements.IDot, CUC: value.Elements.CUC, CUS: value.Elements.CUS, CRC: value.Elements.CRC,
			CRS: value.Elements.CRS, CIC: value.Elements.CIC, CIS: value.Elements.CIS, ToeSOW: value.Elements.ToeSOW,
		},
		Clock: native.NativeClockPolynomial{AF0: value.Clock.AF0, AF1: value.Clock.AF1, AF2: value.Clock.AF2, TocSOW: value.Clock.TocSOW},
		GroupDelays: native.NativeBroadcastGroupDelays{
			HasGPSTGD: value.GroupDelays.HasGPSTGD, GPSTGD: value.GroupDelays.GPSTGD,
			HasGalileoBGDE5AE1: value.GroupDelays.HasGalileoBGDE5AE1, GalileoBGDE5AE1: value.GroupDelays.GalileoBGDE5AE1,
			HasGalileoBGDE5BE1: value.GroupDelays.HasGalileoBGDE5BE1, GalileoBGDE5BE1: value.GroupDelays.GalileoBGDE5BE1,
			HasBeiDouTGD1: value.GroupDelays.HasBeiDouTGD1, BeiDouTGD1: value.GroupDelays.BeiDouTGD1,
			HasBeiDouTGD2: value.GroupDelays.HasBeiDouTGD2, BeiDouTGD2: value.GroupDelays.BeiDouTGD2,
			HasCNAVISCL1CA: value.GroupDelays.HasCNAVISCL1CA, CNAVISCL1CA: value.GroupDelays.CNAVISCL1CA,
			HasCNAVISCL2C: value.GroupDelays.HasCNAVISCL2C, CNAVISCL2C: value.GroupDelays.CNAVISCL2C,
			HasCNAVISCL5I5: value.GroupDelays.HasCNAVISCL5I5, CNAVISCL5I5: value.GroupDelays.CNAVISCL5I5,
			HasCNAVISCL5Q5: value.GroupDelays.HasCNAVISCL5Q5, CNAVISCL5Q5: value.GroupDelays.CNAVISCL5Q5,
			HasCNAVISCL1CD: value.GroupDelays.HasCNAVISCL1CD, CNAVISCL1CD: value.GroupDelays.CNAVISCL1CD,
			HasCNAVISCL1CP: value.GroupDelays.HasCNAVISCL1CP, CNAVISCL1CP: value.GroupDelays.CNAVISCL1CP,
		},
		CNAV: native.NativeBroadcastCNAV{
			Present: value.CNAV.Present, ADOTMPerS: value.CNAV.ADOTMPerS, DeltaN0DotRadPerS2: value.CNAV.DeltaN0DotRadPerS2,
			Top:        native.NativeGnssWeekTow{System: uint32(value.CNAV.Top.System), Week: value.CNAV.Top.Week, TOWSeconds: value.CNAV.Top.TOWSeconds},
			URAEDIndex: value.CNAV.URAEDIndex, URANED0Index: value.CNAV.URANED0Index, URANED1Index: value.CNAV.URANED1Index,
			URANED2Index: value.CNAV.URANED2Index, TransmissionTimeSOW: value.CNAV.TransmissionTimeSOW, HasFlags: value.CNAV.HasFlags, Flags: value.CNAV.Flags,
		},
		SVHealth: value.SVHealth, HasSVAccuracyM: value.HasSVAccuracyM, SVAccuracyM: value.SVAccuracyM, HasFitInterval: value.HasFitInterval, FitIntervalS: value.FitIntervalS,
		Stated: native.NativeStatedNavFields{
			HasOrbit5Field2: value.Stated.HasOrbit5Field2, Orbit5Field2: value.Stated.Orbit5Field2,
			HasOrbit5Field4: value.Stated.HasOrbit5Field4, Orbit5Field4: value.Stated.Orbit5Field4,
			HasOrbit6Field4: value.Stated.HasOrbit6Field4, Orbit6Field4: value.Stated.Orbit6Field4,
			HasTransmissionTimeSOW: value.Stated.HasTransmissionTimeSOW, TransmissionTimeSOW: value.Stated.TransmissionTimeSOW,
			HasOrbit7Field2: value.Stated.HasOrbit7Field2, Orbit7Field2: value.Stated.Orbit7Field2,
			HasOrbit7Field3: value.Stated.HasOrbit7Field3, Orbit7Field3: value.Stated.Orbit7Field3,
			HasOrbit7Field4: value.Stated.HasOrbit7Field4, Orbit7Field4: value.Stated.Orbit7Field4,
		},
	}
}

// EncodeRINEXNav delegates validation and serialization to C. The input
// records are copied into temporary native-compatible storage.
func EncodeRINEXNav(records []BroadcastRecord) ([]byte, error) {
	values := make([]native.NativeBroadcastRecord, len(records))
	for index, record := range records {
		values[index] = broadcastRecordToNative(record)
	}
	output, err := native.EncodeRinexNav(values)
	return output, publicError(err)
}

// RINEXNavRecords owns the complete raw NAV record list returned by C.
// Read-only methods may run concurrently with Close; Close waits for active
// calls to finish before releasing the native resource.
type RINEXNavRecords struct {
	_      noCopy
	handle *native.RinexNavRecords
}

// ParseRINEXNavRecords strictly parses all supported raw NAV records.
func ParseRINEXNavRecords(data []byte) (*RINEXNavRecords, error) {
	handle, err := native.ParseRinexNavRecords(data)
	if err != nil {
		return nil, publicError(err)
	}
	return &RINEXNavRecords{handle: handle}, nil
}

// Close releases parsed NAV records and is idempotent.
func (records *RINEXNavRecords) Close() error {
	if records == nil || records.handle == nil {
		return nil
	}
	return publicError(records.handle.Close())
}

// Count returns the number of records without exposing native storage.
func (records *RINEXNavRecords) Count() (int, error) {
	if records == nil || records.handle == nil {
		return 0, ErrClosed
	}
	value, err := records.handle.Count()
	return value, publicError(err)
}

// Record returns one copied file-order record.
func (records *RINEXNavRecords) Record(index int) (BroadcastRecord, error) {
	if records == nil || records.handle == nil {
		return BroadcastRecord{}, ErrClosed
	}
	value, err := records.handle.Record(index)
	if err != nil {
		return BroadcastRecord{}, publicError(err)
	}
	return broadcastRecordFromNative(value), nil
}

// Records returns Go-owned copies of every record in file order.
func (records *RINEXNavRecords) Records() ([]BroadcastRecord, error) {
	if records == nil || records.handle == nil {
		return nil, ErrClosed
	}
	values, err := records.handle.Records()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]BroadcastRecord, len(values))
	for index := range values {
		out[index] = broadcastRecordFromNative(values[index])
	}
	return out, nil
}

// RINEXNavParse owns a lenient parse result and its skipped-block diagnostics.
// Read-only methods may run concurrently with Close; Close waits for active
// calls to finish before releasing the native resource.
type RINEXNavParse struct {
	_      noCopy
	handle *native.RinexNavParse
}

// ParseRINEXNavLenient parses supported records while retaining malformed
// block diagnostics.
func ParseRINEXNavLenient(data []byte) (*RINEXNavParse, error) {
	handle, err := native.ParseRinexNavLenient(data)
	if err != nil {
		return nil, publicError(err)
	}
	return &RINEXNavParse{handle: handle}, nil
}

// Close releases lenient NAV parse results and is idempotent.
func (parse *RINEXNavParse) Close() error {
	if parse == nil || parse.handle == nil {
		return nil
	}
	return publicError(parse.handle.Close())
}

// Records returns the successfully parsed records from a lenient result.
func (parse *RINEXNavParse) Records() ([]BroadcastRecord, error) {
	if parse == nil || parse.handle == nil {
		return nil, ErrClosed
	}
	values, err := parse.handle.Records()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]BroadcastRecord, len(values))
	for index := range values {
		out[index] = broadcastRecordFromNative(values[index])
	}
	return out, nil
}

// RecordCount returns the number of successfully parsed records.
func (parse *RINEXNavParse) RecordCount() (int, error) {
	if parse == nil || parse.handle == nil {
		return 0, ErrClosed
	}
	value, err := parse.handle.RecordCount()
	return value, publicError(err)
}

// SkippedCount returns the number of malformed blocks retained as diagnostics.
func (parse *RINEXNavParse) SkippedCount() (int, error) {
	if parse == nil || parse.handle == nil {
		return 0, ErrClosed
	}
	value, err := parse.handle.SkippedCount()
	return value, publicError(err)
}

// SkippedNavBlock identifies a malformed block retained by the lenient parser.
type SkippedNavBlock struct {
	// SatelliteID identifies the skipped block; Message contains native diagnostics.
	SatelliteID string
	Message     string
}

// Skipped returns copied diagnostics in file order.
func (parse *RINEXNavParse) Skipped() ([]SkippedNavBlock, error) {
	if parse == nil || parse.handle == nil {
		return nil, ErrClosed
	}
	values, err := parse.handle.Skipped()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]SkippedNavBlock, len(values))
	for index := range values {
		out[index] = SkippedNavBlock{SatelliteID: values[index].SatelliteID, Message: values[index].Message}
	}
	return out, nil
}

// IONOCorrections contains optional NAV-header ionosphere parameters.
type IONOCorrections struct {
	// GPSPresent, BeiDouPresent, and GalileoPresent control corresponding correction fields.
	GPSPresent     bool
	GPSAlpha       [4]float64
	GPSBeta        [4]float64
	BeiDouPresent  bool
	BeiDouAlpha    [4]float64
	BeiDouBeta     [4]float64
	GalileoPresent bool
	GalileoAI0     float64
	GalileoAI1     float64
	GalileoAI2     float64
}

// ParseRINEXIONOCorrections parses optional NAV-header ionosphere fields.
func ParseRINEXIONOCorrections(data []byte) (IONOCorrections, error) {
	value, err := native.ParseRinexIonoCorrections(data)
	return IONOCorrections{GPSPresent: value.GPSPresent, GPSAlpha: value.GPSAlpha, GPSBeta: value.GPSBeta, BeiDouPresent: value.BeiDouPresent, BeiDouAlpha: value.BeiDouAlpha, BeiDouBeta: value.BeiDouBeta, GalileoPresent: value.GalileoPresent, GalileoAI0: value.GalileoAI0, GalileoAI1: value.GalileoAI1, GalileoAI2: value.GalileoAI2}, publicError(err)
}

// ParseRINEXLeapSeconds returns the optional GPS-minus-UTC header value.
func ParseRINEXLeapSeconds(data []byte) (float64, bool, error) {
	value, present, err := native.ParseRinexLeapSeconds(data)
	return value, present, publicError(err)
}

// ClockPoint is one complete copied RINEX clock sample.
type ClockPoint struct {
	// Epoch is the lossless clock epoch; BiasS is bias in seconds.
	Epoch ClockEpoch
	BiasS float64
}

// RINEXClock owns a parsed RINEX clock product.
// Read-only methods may run concurrently with Close; Close waits for active
// calls to finish before releasing the native resource.
type RINEXClock struct {
	_      noCopy
	handle *native.RinexClock
}

// RINEXClockErrorKind identifies a RINEX clock parse or write failure.
type RINEXClockErrorKind uint32

const (
	// RINEXClockErrorNone means no failure is recorded.
	RINEXClockErrorNone RINEXClockErrorKind = 0
	// RINEXClockErrorMalformedASRecord reports a short satellite clock record.
	RINEXClockErrorMalformedASRecord RINEXClockErrorKind = 1
	// RINEXClockErrorMissingContinuation reports a missing declared continuation.
	RINEXClockErrorMissingContinuation RINEXClockErrorKind = 2
	// RINEXClockErrorMalformedContinuation reports a malformed continuation line.
	RINEXClockErrorMalformedContinuation RINEXClockErrorKind = 3
	// RINEXClockErrorBadField reports an unreadable or out-of-range field.
	RINEXClockErrorBadField RINEXClockErrorKind = 4
	// RINEXClockErrorInvalidInput reports refused caller input or unwritable exact values.
	RINEXClockErrorInvalidInput RINEXClockErrorKind = 5
	// RINEXClockErrorUnsupportedTimeScale reports a time scale the format cannot name.
	RINEXClockErrorUnsupportedTimeScale RINEXClockErrorKind = 6
)

// RINEXClockTimeSystem identifies the system declared by a clock file.
type RINEXClockTimeSystem uint32

// RINEXClockLayout identifies the record-column layout.
type RINEXClockLayout uint32

const (
	// RINEXClockLayoutV300 is the 80-column layout used through version 3.03.
	RINEXClockLayoutV300 RINEXClockLayout = 0
	// RINEXClockLayoutV304 is the 85-column layout used by version 3.04.
	RINEXClockLayoutV304 RINEXClockLayout = 1
)

// RINEXClockTimeSystemStatus records how the product's time system was set.
type RINEXClockTimeSystemStatus uint32

const (
	// RINEXClockTimeSystemDeclared means a TIME SYSTEM ID row supplied the value.
	RINEXClockTimeSystemDeclared RINEXClockTimeSystemStatus = 0
	// RINEXClockTimeSystemDefaulted means the parser applied its format default.
	RINEXClockTimeSystemDefaulted RINEXClockTimeSystemStatus = 1
	// RINEXClockTimeSystemUnrecognized means the file contained an unknown label.
	RINEXClockTimeSystemUnrecognized RINEXClockTimeSystemStatus = 2
	// RINEXClockTimeSystemConflicting means retained declarations disagree.
	RINEXClockTimeSystemConflicting RINEXClockTimeSystemStatus = 3
	// RINEXClockTimeSystemConstructed means the value came from a constructor.
	RINEXClockTimeSystemConstructed RINEXClockTimeSystemStatus = 4
	// RINEXClockTimeSystemStatusUnknown preserves a future status value.
	RINEXClockTimeSystemStatusUnknown RINEXClockTimeSystemStatus = 999
)

const (
	// RINEXClockTimeSystemGPS is GPS system time.
	RINEXClockTimeSystemGPS RINEXClockTimeSystem = 0
	// RINEXClockTimeSystemGLO is GLONASS time as reported by RINEX.
	RINEXClockTimeSystemGLO RINEXClockTimeSystem = 1
	// RINEXClockTimeSystemGAL is Galileo system time.
	RINEXClockTimeSystemGAL RINEXClockTimeSystem = 2
	// RINEXClockTimeSystemQZS is QZSS system time.
	RINEXClockTimeSystemQZS RINEXClockTimeSystem = 3
	// RINEXClockTimeSystemBDS is BeiDou system time.
	RINEXClockTimeSystemBDS RINEXClockTimeSystem = 4
	// RINEXClockTimeSystemIRN is IRNSS system time.
	RINEXClockTimeSystemIRN RINEXClockTimeSystem = 5
	// RINEXClockTimeSystemUTC is Coordinated Universal Time.
	RINEXClockTimeSystemUTC RINEXClockTimeSystem = 6
	// RINEXClockTimeSystemTAI is International Atomic Time.
	RINEXClockTimeSystemTAI RINEXClockTimeSystem = 7
	// RINEXClockTimeSystemUnknown preserves a future native system code.
	RINEXClockTimeSystemUnknown RINEXClockTimeSystem = 999
)

// RINEXClockRecordType identifies a RINEX clock Table A16 record.
type RINEXClockRecordType uint32

const (
	// RINEXClockRecordAR is a receiver analysis result.
	RINEXClockRecordAR RINEXClockRecordType = 0
	// RINEXClockRecordAS is a satellite analysis result.
	RINEXClockRecordAS RINEXClockRecordType = 1
	// RINEXClockRecordCR is a receiver calibration measurement.
	RINEXClockRecordCR RINEXClockRecordType = 2
	// RINEXClockRecordDR is a receiver discontinuity measurement.
	RINEXClockRecordDR RINEXClockRecordType = 3
	// RINEXClockRecordMS is a broadcast-clock monitor measurement.
	RINEXClockRecordMS RINEXClockRecordType = 4
)

// RINEXClockRecordReading describes how a data row was interpreted.
type RINEXClockRecordReading uint32

const (
	// RINEXClockRecordReadingColumnsV300 is the 80-column layout.
	RINEXClockRecordReadingColumnsV300 RINEXClockRecordReading = 0
	// RINEXClockRecordReadingColumnsV304 is the 85-column layout.
	RINEXClockRecordReadingColumnsV304 RINEXClockRecordReading = 1
	// RINEXClockRecordReadingWhitespace is whitespace-separated.
	RINEXClockRecordReadingWhitespace RINEXClockRecordReading = 2
	// RINEXClockRecordReadingEdited was created or changed through the typed API.
	RINEXClockRecordReadingEdited RINEXClockRecordReading = 3
	// RINEXClockRecordReadingUnknown preserves a future reading value.
	RINEXClockRecordReadingUnknown RINEXClockRecordReading = 999
)

// RINEXClockHeaderReading describes how a header row was interpreted.
type RINEXClockHeaderReading uint32

const (
	// RINEXClockHeaderReadingColumns uses columns for the declared version.
	RINEXClockHeaderReadingColumns RINEXClockHeaderReading = 0
	// RINEXClockHeaderReadingOtherVersionColumns uses the other supported layout.
	RINEXClockHeaderReadingOtherVersionColumns RINEXClockHeaderReading = 1
	// RINEXClockHeaderReadingWhitespace uses whitespace-separated fields.
	RINEXClockHeaderReadingWhitespace RINEXClockHeaderReading = 2
	// RINEXClockHeaderReadingUninterpreted has a known label and unread fields.
	RINEXClockHeaderReadingUninterpreted RINEXClockHeaderReading = 3
	// RINEXClockHeaderReadingUnknownLabel has an unrecognized label.
	RINEXClockHeaderReadingUnknownLabel RINEXClockHeaderReading = 4
	// RINEXClockHeaderReadingUnknown preserves a future reading value.
	RINEXClockHeaderReadingUnknown RINEXClockHeaderReading = 999
)

// RINEXClockHeaderFieldKind identifies the typed fields carried by a header row.
type RINEXClockHeaderFieldKind uint32

const (
	// RINEXClockHeaderFieldNone has no typed field value.
	RINEXClockHeaderFieldNone RINEXClockHeaderFieldKind = 0
	// RINEXClockHeaderFieldVersionType is the version/type row.
	RINEXClockHeaderFieldVersionType RINEXClockHeaderFieldKind = 1
	// RINEXClockHeaderFieldProgramRunByDate is the program/agency/date row.
	RINEXClockHeaderFieldProgramRunByDate RINEXClockHeaderFieldKind = 2
	// RINEXClockHeaderFieldComment is a comment row.
	RINEXClockHeaderFieldComment RINEXClockHeaderFieldKind = 3
	// RINEXClockHeaderFieldObservationTypes declares observation types.
	RINEXClockHeaderFieldObservationTypes RINEXClockHeaderFieldKind = 4
	// RINEXClockHeaderFieldTimeSystem declares the time system.
	RINEXClockHeaderFieldTimeSystem RINEXClockHeaderFieldKind = 5
	// RINEXClockHeaderFieldLeapSeconds is the leap-second count.
	RINEXClockHeaderFieldLeapSeconds RINEXClockHeaderFieldKind = 6
	// RINEXClockHeaderFieldLeapSecondsGNSS is the GNSS leap-second count.
	RINEXClockHeaderFieldLeapSecondsGNSS RINEXClockHeaderFieldKind = 7
	// RINEXClockHeaderFieldDCBSApplied records differential code-bias corrections.
	RINEXClockHeaderFieldDCBSApplied RINEXClockHeaderFieldKind = 8
	// RINEXClockHeaderFieldPCVSApplied records phase-center variation corrections.
	RINEXClockHeaderFieldPCVSApplied RINEXClockHeaderFieldKind = 9
	// RINEXClockHeaderFieldTypesOfData declares the data types.
	RINEXClockHeaderFieldTypesOfData RINEXClockHeaderFieldKind = 10
	// RINEXClockHeaderFieldStationNameNum identifies a station.
	RINEXClockHeaderFieldStationNameNum RINEXClockHeaderFieldKind = 11
	// RINEXClockHeaderFieldStationClockRef identifies a station clock reference.
	RINEXClockHeaderFieldStationClockRef RINEXClockHeaderFieldKind = 12
	// RINEXClockHeaderFieldAnalysisCenter identifies the analysis center.
	RINEXClockHeaderFieldAnalysisCenter RINEXClockHeaderFieldKind = 13
	// RINEXClockHeaderFieldClockRefCount declares clock-reference coverage.
	RINEXClockHeaderFieldClockRefCount RINEXClockHeaderFieldKind = 14
	// RINEXClockHeaderFieldAnalysisClockRef identifies the analysis clock reference.
	RINEXClockHeaderFieldAnalysisClockRef RINEXClockHeaderFieldKind = 15
	// RINEXClockHeaderFieldSolutionStationCount declares solution stations.
	RINEXClockHeaderFieldSolutionStationCount RINEXClockHeaderFieldKind = 16
	// RINEXClockHeaderFieldSolutionStation stores a solution station.
	RINEXClockHeaderFieldSolutionStation RINEXClockHeaderFieldKind = 17
	// RINEXClockHeaderFieldSolutionSatelliteCount declares solution satellites.
	RINEXClockHeaderFieldSolutionSatelliteCount RINEXClockHeaderFieldKind = 18
	// RINEXClockHeaderFieldPRNList lists satellite identifiers.
	RINEXClockHeaderFieldPRNList RINEXClockHeaderFieldKind = 19
	// RINEXClockHeaderFieldEndOfHeader marks the header end.
	RINEXClockHeaderFieldEndOfHeader RINEXClockHeaderFieldKind = 20
	// RINEXClockHeaderFieldUnknown preserves a future field-kind value.
	RINEXClockHeaderFieldUnknown RINEXClockHeaderFieldKind = 999
)

// RINEXClockSurplusValue is a value beyond a record's declared count.
type RINEXClockSurplusValue struct {
	// Position is the zero-based position in the ordered value sequence.
	Position uint64
	// Value is the retained extra numeric value.
	Value float64
}

// RINEXClockRecord is one detached data record, including its typed and raw identity.
type RINEXClockRecord struct {
	// RecordType identifies the AR, AS, CR, DR, or MS record.
	RecordType RINEXClockRecordType
	// Name is the trimmed receiver or satellite name as written.
	Name string
	// HasSatellite reports whether the record has a canonical satellite identifier.
	HasSatellite bool
	// SatelliteID is the canonical satellite identifier when present.
	SatelliteID string
	// CivilEpoch retains calendar fields in the product time system.
	CivilEpoch CivilDateTime
	// HasEpoch reports whether the civil epoch converts to a continuous instant.
	HasEpoch bool
	// Epoch is the exact native instant representation when HasEpoch is true.
	Epoch ClockEpoch
	// Values contains declared clock values in their original order.
	Values []float64
	// Surplus contains retained values beyond the declared count.
	Surplus []RINEXClockSurplusValue
	// HasLine reports whether Line identifies a source row.
	HasLine bool
	// Line is the one-based first source line.
	Line uint64
	// LineCount is the number of physical lines occupied by the record.
	LineCount uint64
	// Reading identifies the input layout used for the record.
	Reading RINEXClockRecordReading
	// HasContinuationReading reports whether a continuation row was interpreted.
	HasContinuationReading bool
	// ContinuationReading identifies the continuation-row input layout.
	ContinuationReading RINEXClockRecordReading
	// ReadingUnknownVariant retains a future reading name.
	ReadingUnknownVariant string
	// ContinuationReadingUnknownVariant retains a future continuation reading name.
	ContinuationReadingUnknownVariant string
}

// RINEXClockInfo is a detached product summary. Optional fields are governed
// by their Has flags; counts include every record type unless named otherwise.
type RINEXClockInfo struct {
	// HasVersion reports whether Version is meaningful.
	HasVersion bool
	// Version is the declared or constructed format version.
	Version float64
	// HasLayout reports whether Layout is meaningful.
	HasLayout bool
	// Layout is the product's column layout.
	Layout RINEXClockLayout
	// HasSatelliteSystem reports whether SatelliteSystem is meaningful.
	HasSatelliteSystem bool
	// SatelliteSystem is the RINEX system code point.
	SatelliteSystem uint32
	// HasTimeSystem reports whether TimeSystem is present.
	HasTimeSystem bool
	// TimeSystem is the declared, defaulted, or constructed time system.
	TimeSystem RINEXClockTimeSystem
	// TimeSystemStatus explains how TimeSystem was established.
	TimeSystemStatus RINEXClockTimeSystemStatus
	// TimeSystemLabelCount counts retained unrecognized/conflicting labels.
	TimeSystemLabelCount int
	// HasTimeScale reports whether records have a resolved continuous scale.
	HasTimeScale bool
	// TimeScale is the resolved scale when HasTimeScale is true.
	TimeScale TimeScale
	// HeaderRecordCount counts source header rows; built products have zero.
	HeaderRecordCount int
	// RecordCount counts all readable rows of every record type.
	RecordCount int
	// SeriesCount counts satellites with AS clock series.
	SeriesCount int
	// SampleCount counts samples across all series.
	SampleCount int
	// SkippedRecordCount counts non-series records retained as skipped rows.
	SkippedRecordCount int
	// DiagnosticCount counts lossy parse diagnostics.
	DiagnosticCount int
	// NoticeCount counts non-fatal parse notices.
	NoticeCount int
	// TimeSystemUnknownVariant preserves a future time-system name.
	TimeSystemUnknownVariant string
	// TimeSystemStatusUnknownVariant preserves a future status name.
	TimeSystemStatusUnknownVariant string
}

// RINEXClockRecordValuesEdit identifies one record and its replacement values.
type RINEXClockRecordValuesEdit struct {
	// Index is the zero-based record-order index.
	Index int
	// Values replaces the declared numeric fields, with bias first.
	Values []float64
}

// RINEXClockHeaderRecord is a detached typed header row with retained text.
type RINEXClockHeaderRecord struct {
	// HasLine reports whether Line identifies a source row.
	HasLine bool
	// Line is the one-based original source line.
	Line uint64
	// LabelColumn is the zero-based label start column.
	LabelColumn uint64
	// Reading describes how the source fields were interpreted.
	Reading RINEXClockHeaderReading
	// FieldKind identifies the typed reading kind.
	FieldKind RINEXClockHeaderFieldKind
	// TextParts contains copied typed values in native field order.
	TextParts []string
	// HasVersion reports whether Version is present.
	HasVersion bool
	// Version is the format version when this is a version row.
	Version float64
	// HasSystemCode reports whether SystemCode is present.
	HasSystemCode bool
	// SystemCode is the observation system code point.
	SystemCode uint32
	// HasCount reports whether Count is present.
	HasCount bool
	// Count is the typed count value.
	Count uint64
	// HasInteger reports whether Integer is present.
	HasInteger bool
	// Integer is a retained integral header value.
	Integer int64
	// HasStart reports whether Start is present.
	HasStart bool
	// Start is the start civil epoch.
	Start CivilDateTime
	// HasStop reports whether Stop is present.
	HasStop bool
	// Stop is the stop civil epoch.
	Stop CivilDateTime
	// HasConstraintS reports whether ConstraintS is present.
	HasConstraintS bool
	// ConstraintS is the analysis clock-reference constraint in seconds.
	ConstraintS float64
	// HasXYZMM reports whether XYZMM is present.
	HasXYZMM bool
	// XYZMM is the three geocentric station coordinates in millimeters.
	XYZMM [3]int64
	// ReadingUnknownVariant retains a future reading name.
	ReadingUnknownVariant string
	// FieldKindUnknownVariant retains a future field-kind name.
	FieldKindUnknownVariant string
	// LineText is the complete source row without its terminator.
	LineText string
	// Label is the source header label.
	Label string
	// Payload is the source text preceding the label.
	Payload string
}

// RINEXClockWritePolicy controls the explicitly allowed writer departure.
type RINEXClockWritePolicy struct {
	// AllowNearestMicrosecondEpoch permits nearest-microsecond epoch text when exact text is impossible.
	AllowNearestMicrosecondEpoch bool
}

// RINEXClockWriteDepartureKind identifies a reported writer departure.
type RINEXClockWriteDepartureKind uint32

const (
	// RINEXClockWriteDepartureNearestMicrosecond reports a rounded epoch.
	RINEXClockWriteDepartureNearestMicrosecond RINEXClockWriteDepartureKind = 1
	// RINEXClockWriteDepartureUnknown preserves a future native departure kind.
	RINEXClockWriteDepartureUnknown RINEXClockWriteDepartureKind = 999
)

// RINEXClockWriteDeparture records a change explicitly permitted by policy.
type RINEXClockWriteDeparture struct {
	// Kind identifies the departure.
	Kind RINEXClockWriteDepartureKind
	// Record is the zero-based data-record index.
	Record uint64
	// HasEpoch reports whether Epoch identifies the original instant.
	HasEpoch bool
	// Epoch is the original exact epoch.
	Epoch ClockEpoch
	// UnknownVariant preserves a future native kind name.
	UnknownVariant string
	// Name is the record name the writer retained.
	Name string
	// Written is the epoch text the writer emitted.
	Written string
}

// RINEXClockWriteResult contains serialized text and every allowed departure.
type RINEXClockWriteResult struct {
	// Text is the serialized RINEX clock product.
	Text []byte
	// Departures contains each explicitly allowed writer departure.
	Departures []RINEXClockWriteDeparture
}

// RINEXClockFailure retains the native typed detail for a parse or write refusal.
type RINEXClockFailure struct {
	// Kind identifies the native failure; unknown values are preserved.
	Kind RINEXClockErrorKind
	// HasLine reports whether Line identifies the failing input row.
	HasLine bool
	// Line is the one-based failing input line.
	Line uint64
	// HasTimeScale reports whether TimeScale identifies the refused scale.
	HasTimeScale bool
	// TimeScale is the refused time scale.
	TimeScale TimeScale
	// HasField reports whether Field names a failed field.
	HasField bool
	// Field is the failed field name.
	Field string
	// HasReason reports whether Reason retains a native explanation.
	HasReason bool
	// Reason is the native failure explanation.
	Reason string
	// HasRecord reports whether Record retains source record text.
	HasRecord bool
	// Record is the source record text.
	Record string
	// HasRecordType reports whether RecordType names a failed record type.
	HasRecordType bool
	// RecordType is the source record type text.
	RecordType string
	// HasValue reports whether Value retains the rejected value.
	HasValue bool
	// Value is the rejected field value.
	Value string
	// Message is the complete native failure text.
	Message string
}

// RINEXClockWriteError names a failure returned while writing a clock product.
type RINEXClockWriteError = RINEXClockFailure

// Error returns the complete native failure text.
func (failure *RINEXClockFailure) Error() string {
	if failure == nil {
		return "sidereon: RINEX clock operation failed"
	}
	return failure.Message
}

func publicClockWriteError(err error) error {
	if err == nil {
		return nil
	}
	var failure *native.NativeClockWriteFailure
	if errors.As(err, &failure) {
		return &RINEXClockWriteError{Kind: RINEXClockErrorKind(failure.Kind), HasLine: failure.HasLine, Line: failure.Line, HasTimeScale: failure.HasTimeScale, TimeScale: TimeScale(failure.TimeScale), HasField: failure.HasField, Field: failure.Field, HasReason: failure.HasReason, Reason: failure.Reason, HasRecord: failure.HasRecord, Record: failure.Record, HasRecordType: failure.HasRecordType, RecordType: failure.RecordType, HasValue: failure.HasValue, Value: failure.Value, Message: failure.Message}
	}
	return publicError(err)
}

// RINEXClockNoticeKind identifies a non-fatal clock parser finding.
type RINEXClockNoticeKind uint32

const (
	// RINEXClockNoticeTimeSystemDefaulted reports use of the format default.
	RINEXClockNoticeTimeSystemDefaulted RINEXClockNoticeKind = 1
	// RINEXClockNoticeTimeSystemMissing reports a missing required system label.
	RINEXClockNoticeTimeSystemMissing RINEXClockNoticeKind = 2
	// RINEXClockNoticeTimeSystemWithoutScale reports a declared system with no core scale.
	RINEXClockNoticeTimeSystemWithoutScale RINEXClockNoticeKind = 3
	// RINEXClockNoticeHeaderNonconforming reports a header line read at alternate columns.
	RINEXClockNoticeHeaderNonconforming RINEXClockNoticeKind = 4
	// RINEXClockNoticeHeaderUninterpreted reports a known label whose fields failed to parse.
	RINEXClockNoticeHeaderUninterpreted RINEXClockNoticeKind = 5
	// RINEXClockNoticeHeaderUnknownLabel reports an unrecognized header label.
	RINEXClockNoticeHeaderUnknownLabel RINEXClockNoticeKind = 6
	// RINEXClockNoticeSurplusValues reports records with values beyond their declaration.
	RINEXClockNoticeSurplusValues RINEXClockNoticeKind = 7
	// RINEXClockNoticeOtherLayoutRecords reports records read at alternate columns.
	RINEXClockNoticeOtherLayoutRecords RINEXClockNoticeKind = 8
	// RINEXClockNoticeWhitespaceRecords reports records read as whitespace-separated values.
	RINEXClockNoticeWhitespaceRecords RINEXClockNoticeKind = 9
	// RINEXClockNoticeUnknown preserves a future native notice kind.
	RINEXClockNoticeUnknown RINEXClockNoticeKind = 999
)

// RINEXClockDiagnostic retains one lossy-read or header-time-system finding.
type RINEXClockDiagnostic struct {
	// Kind is the native error-kind discriminant.
	Kind RINEXClockErrorKind
	// HasLine reports whether Line identifies a source row.
	HasLine bool
	// Line is the one-based source line when present.
	Line uint64
	// HasErrorLine reports whether the nested typed error carries its own line.
	HasErrorLine bool
	// ErrorLine is the line carried by the nested typed error.
	ErrorLine uint64
	// HasTimeScale reports whether TimeScale identifies a refused scale.
	HasTimeScale bool
	// TimeScale is the refused scale when present.
	TimeScale TimeScale
	// HasField reports whether Field names the rejected field.
	HasField bool
	// HasReason reports whether Reason is available.
	HasReason bool
	// HasRecord reports whether Record retains the source record.
	HasRecord bool
	// HasRecordType reports whether RecordType identifies a missing continuation type.
	HasRecordType bool
	// HasValue reports whether Value retains the rejected field value.
	HasValue bool
	// Message is the complete native diagnostic text.
	Message string
	// Field is the field name when HasField is true.
	Field string
	// Reason is the parser reason when HasReason is true.
	Reason string
	// Record is the original record text when HasRecord is true.
	Record string
	// RecordType names the expected continuation when present.
	RecordType string
	// Value is the rejected field value when HasValue is true.
	Value string
}

// RINEXClockNotice preserves a non-fatal parser finding and future enum names.
type RINEXClockNotice struct {
	// Kind is the native notice-kind discriminant.
	Kind RINEXClockNoticeKind
	// HasTimeSystem reports whether TimeSystem applies.
	HasTimeSystem bool
	// TimeSystem is the native time-system discriminant.
	TimeSystem RINEXClockTimeSystem
	// HasLine reports whether Line identifies a header line.
	HasLine bool
	// Line is the one-based header line when present.
	Line uint64
	// HasRecords reports whether Records and FirstLine are meaningful.
	HasRecords bool
	// Records is the number of records covered by the notice.
	Records uint64
	// FirstLine is the one-based first line of the affected records.
	FirstLine uint64
	// KindUnknownVariant preserves a future notice name.
	KindUnknownVariant string
	// TimeSystemUnknownVariant preserves a future time-system name.
	TimeSystemUnknownVariant string
}

// RINEXClockSkippedRecord identifies a non-AS record retained by the parser.
type RINEXClockSkippedRecord struct {
	// Line is the one-based source line.
	Line uint64
	// RecordType is the native clock-record discriminant.
	RecordType RINEXClockRecordType
}

// ParseRINEXClock strictly parses a RINEX clock product.
func ParseRINEXClock(data []byte) (*RINEXClock, error) { return parseRINEXClock(data, false) }

// RINEXClockParseOutcome retains the typed result of a well-formed strict parse
// attempt. Structural input errors are returned as Go errors separately.
type RINEXClockParseOutcome struct {
	// IsOK reports whether the text produced a clock product.
	IsOK bool
	// Failure contains the exact parser fields and message when IsOK is false.
	Failure *RINEXClockFailure
}

// ParseRINEXClockWithOutcome strictly parses text and returns its owned typed
// refusal separately from structural call errors.
func ParseRINEXClockWithOutcome(data []byte) (*RINEXClock, RINEXClockParseOutcome, error) {
	handle, failure, err := native.ParseRinexClockWithOutcome(data)
	if err != nil {
		return nil, RINEXClockParseOutcome{}, publicError(err)
	}
	if failure != nil {
		converted, _ := publicClockWriteError(failure).(*RINEXClockFailure)
		return nil, RINEXClockParseOutcome{Failure: converted}, nil
	}
	if handle == nil {
		return nil, RINEXClockParseOutcome{}, errors.New("sidereon: native RINEX clock parser returned no outcome")
	}
	return &RINEXClock{handle: handle}, RINEXClockParseOutcome{IsOK: true}, nil
}

// ParseRINEXClockLossy skips malformed and non-AS rows according to C.
func ParseRINEXClockLossy(data []byte) (*RINEXClock, error) { return parseRINEXClock(data, true) }

func parseRINEXClock(data []byte, lossy bool) (*RINEXClock, error) {
	handle, err := native.ParseRinexClock(data, lossy)
	if err != nil {
		return nil, publicError(err)
	}
	return &RINEXClock{handle: handle}, nil
}

// Close releases the RINEX clock and is idempotent.
func (clock *RINEXClock) Close() error {
	if clock == nil || clock.handle == nil {
		return nil
	}
	return publicError(clock.handle.Close())
}

// Satellites returns detached satellite identifiers in the clock product.
func (clock *RINEXClock) Satellites() ([]string, error) {
	if clock == nil || clock.handle == nil {
		return nil, ErrClosed
	}
	values, err := clock.handle.Satellites()
	return values, publicError(err)
}

// SatelliteCount returns the number of clock-product satellites.
func (clock *RINEXClock) SatelliteCount() (int, error) {
	if clock == nil || clock.handle == nil {
		return 0, ErrClosed
	}
	value, err := clock.handle.SatelliteCount()
	return value, publicError(err)
}

// SeriesCount returns the exact number of satellite series in the product.
func (clock *RINEXClock) SeriesCount() (int, error) {
	if clock == nil || clock.handle == nil {
		return 0, ErrClosed
	}
	value, err := clock.handle.SeriesCount()
	return value, publicError(err)
}

// SampleCount returns the number of clock samples.
func (clock *RINEXClock) SampleCount() (int, error) {
	if clock == nil || clock.handle == nil {
		return 0, ErrClosed
	}
	value, err := clock.handle.SampleCount()
	return value, publicError(err)
}

// RINEXClockSeries owns one C-backed satellite clock series. Read-only methods may run concurrently with Close; Close waits for active
// calls to finish before releasing the native resource.
type RINEXClockSeries struct {
	_      noCopy
	handle *native.ClockSeries
}

// Series returns each satellite series as an owning child handle. Callers
// should Close every returned series when finished.
func (clock *RINEXClock) Series() ([]*RINEXClockSeries, error) {
	if clock == nil || clock.handle == nil {
		return nil, ErrClosed
	}
	count, err := clock.SeriesCount()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]*RINEXClockSeries, count)
	for index := range out {
		out[index], err = clock.SeriesAt(index)
		if err != nil {
			cleanupErrs := []error{err}
			for _, value := range out[:index] {
				if closeErr := value.Close(); closeErr != nil {
					cleanupErrs = append(cleanupErrs, closeErr)
				}
			}
			return nil, errors.Join(cleanupErrs...)
		}
	}
	return out, nil
}

// SeriesAt returns one series by deterministic satellite order.
func (clock *RINEXClock) SeriesAt(index int) (*RINEXClockSeries, error) {
	if clock == nil || clock.handle == nil {
		return nil, ErrClosed
	}
	value, err := clock.handle.Series(index)
	if err != nil {
		return nil, publicError(err)
	}
	return &RINEXClockSeries{handle: value}, nil
}

// SeriesFor returns nil, nil when the satellite has no parsed AS series.
func (clock *RINEXClock) SeriesFor(satelliteID string) (*RINEXClockSeries, error) {
	if clock == nil || clock.handle == nil {
		return nil, ErrClosed
	}
	value, err := clock.handle.SeriesFor(satelliteID)
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, nil
	}
	return &RINEXClockSeries{handle: value}, nil
}

// BiasAtGPSSeconds interpolates one satellite clock bias. Available is false
// when C has no usable value at the requested epoch.
func (clock *RINEXClock) BiasAtGPSSeconds(satelliteID string, gpsSeconds float64) (bias float64, available bool, err error) {
	if clock == nil || clock.handle == nil {
		return 0, false, ErrClosed
	}
	bias, available, err = clock.handle.BiasAtGPSSeconds(satelliteID, gpsSeconds)
	return bias, available, publicError(err)
}

// BiasAtCivil interpolates in the product's time system at an exact civil epoch.
func (clock *RINEXClock) BiasAtCivil(satelliteID string, epoch CivilDateTime) (bias float64, available bool, err error) {
	if clock == nil || clock.handle == nil {
		return 0, false, ErrClosed
	}
	bias, available, err = clock.handle.BiasAtCivil(satelliteID, native.CivilDateTime{Year: epoch.Year, Month: epoch.Month, Day: epoch.Day, Hour: epoch.Hour, Minute: epoch.Minute, Second: epoch.Second})
	return bias, available, publicError(err)
}

// BiasAtEpoch interpolates at a scale-tagged continuous epoch without converting
// the supplied instant through GPS seconds.
func (clock *RINEXClock) BiasAtEpoch(satelliteID string, epoch ClockEpoch) (bias float64, available bool, err error) {
	if clock == nil || clock.handle == nil {
		return 0, false, ErrClosed
	}
	bias, available, err = clock.handle.BiasAtEpoch(satelliteID, nativeClockEpoch(epoch))
	return bias, available, publicError(err)
}

// ToText serializes the parsed product through the C writer.
func (clock *RINEXClock) ToText() ([]byte, error) {
	if clock == nil || clock.handle == nil {
		return nil, ErrClosed
	}
	value, err := clock.handle.Text()
	return value, publicError(err)
}

// Records returns detached, typed snapshots of every readable clock data row.
func (clock *RINEXClock) Records() ([]RINEXClockRecord, error) {
	if clock == nil || clock.handle == nil {
		return nil, ErrClosed
	}
	values, err := clock.handle.Records()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]RINEXClockRecord, len(values))
	for i, v := range values {
		surplus := make([]RINEXClockSurplusValue, len(v.Surplus))
		for j, value := range v.Surplus {
			surplus[j] = RINEXClockSurplusValue{Position: value.Position, Value: value.Value}
		}
		out[i] = RINEXClockRecord{RecordType: RINEXClockRecordType(v.RecordType), Name: v.Name, HasSatellite: v.HasSatellite, SatelliteID: v.Satellite, CivilEpoch: CivilDateTime{Year: v.CivilEpoch.Year, Month: v.CivilEpoch.Month, Day: v.CivilEpoch.Day, Hour: v.CivilEpoch.Hour, Minute: v.CivilEpoch.Minute, Second: v.CivilEpoch.Second}, HasEpoch: v.HasEpoch, Epoch: ClockEpoch{Scale: TimeScale(v.Epoch.Scale), Representation: RINEXClockInstantRepresentation(v.Epoch.Representation), JulianWhole: v.Epoch.JulianWhole, JulianFraction: v.Epoch.JulianFraction, NanosHigh: v.Epoch.NanosHigh, NanosLow: v.Epoch.NanosLow}, Values: append([]float64(nil), v.Values...), Surplus: surplus, HasLine: v.HasLine, Line: v.Line, LineCount: v.LineCount, Reading: RINEXClockRecordReading(v.Reading), HasContinuationReading: v.HasContinuationReading, ContinuationReading: RINEXClockRecordReading(v.ContinuationReading), ReadingUnknownVariant: v.ReadingUnknownVariant, ContinuationReadingUnknownVariant: v.ContinuationReadingUnknownVariant}
	}
	return out, nil
}

// RecordCount returns the number of readable data rows of every record type.
func (clock *RINEXClock) RecordCount() (int, error) {
	if clock == nil || clock.handle == nil {
		return 0, ErrClosed
	}
	n, err := clock.handle.RecordCount()
	return n, publicError(err)
}

// Info returns the fixed summary and optional-field/count metadata for a clock.
func (clock *RINEXClock) Info() (RINEXClockInfo, error) {
	if clock == nil || clock.handle == nil {
		return RINEXClockInfo{}, ErrClosed
	}
	v, err := clock.handle.Info()
	if err != nil {
		return RINEXClockInfo{}, publicError(err)
	}
	return RINEXClockInfo{HasVersion: v.HasVersion, Version: v.Version, HasLayout: v.HasLayout, Layout: RINEXClockLayout(v.Layout), HasSatelliteSystem: v.HasSatelliteSystem, SatelliteSystem: v.SatelliteSystem, HasTimeSystem: v.HasTimeSystem, TimeSystem: RINEXClockTimeSystem(v.TimeSystem), TimeSystemStatus: RINEXClockTimeSystemStatus(v.TimeSystemStatus), TimeSystemLabelCount: v.TimeSystemLabelCount, HasTimeScale: v.HasTimeScale, TimeScale: TimeScale(v.TimeScale), HeaderRecordCount: v.HeaderRecordCount, RecordCount: v.RecordCount, SeriesCount: v.SeriesCount, SampleCount: v.SampleCount, SkippedRecordCount: v.SkippedRecordCount, DiagnosticCount: v.DiagnosticCount, NoticeCount: v.NoticeCount, TimeSystemUnknownVariant: v.TimeSystemUnknownVariant, TimeSystemStatusUnknownVariant: v.TimeSystemStatusUnknownVariant}, nil
}

// SourceLine returns a detached original line and whether the parsed input
// contained that one-based line number. Constructed products report false.
func (clock *RINEXClock) SourceLine(line uint64) (string, bool, error) {
	if clock == nil || clock.handle == nil {
		return "", false, ErrClosed
	}
	text, present, err := clock.handle.SourceLine(line)
	return text, present, publicError(err)
}

// TimeSystemLabel returns a retained label from an unrecognized or conflicting
// file time-system declaration, in source order.
func (clock *RINEXClock) TimeSystemLabel(index int) (string, error) {
	if clock == nil || clock.handle == nil {
		return "", ErrClosed
	}
	text, err := clock.handle.TimeSystemLabel(index)
	return text, publicError(err)
}

// HeaderRecords returns detached typed rows and original text from the clock header.
func (clock *RINEXClock) HeaderRecords() ([]RINEXClockHeaderRecord, error) {
	if clock == nil || clock.handle == nil {
		return nil, ErrClosed
	}
	values, err := clock.handle.HeaderRecords()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]RINEXClockHeaderRecord, len(values))
	for i, v := range values {
		out[i] = RINEXClockHeaderRecord{HasLine: v.HasLine, Line: v.Line, LabelColumn: v.LabelColumn, Reading: RINEXClockHeaderReading(v.Reading), FieldKind: RINEXClockHeaderFieldKind(v.FieldKind), TextParts: append([]string(nil), v.TextParts...), HasVersion: v.HasVersion, Version: v.Version, HasSystemCode: v.HasSystemCode, SystemCode: v.SystemCode, HasCount: v.HasCount, Count: v.Count, HasInteger: v.HasInteger, Integer: v.Integer, HasStart: v.HasStart, Start: CivilDateTime{Year: v.Start.Year, Month: v.Start.Month, Day: v.Start.Day, Hour: v.Start.Hour, Minute: v.Start.Minute, Second: v.Start.Second}, HasStop: v.HasStop, Stop: CivilDateTime{Year: v.Stop.Year, Month: v.Stop.Month, Day: v.Stop.Day, Hour: v.Stop.Hour, Minute: v.Stop.Minute, Second: v.Stop.Second}, HasConstraintS: v.HasConstraintS, ConstraintS: v.ConstraintS, HasXYZMM: v.HasXYZMM, XYZMM: v.XYZMM, ReadingUnknownVariant: v.ReadingUnknownVariant, FieldKindUnknownVariant: v.FieldKindUnknownVariant, LineText: v.LineText, Label: v.Label, Payload: v.Payload}
	}
	return out, nil
}

// InsertRecord inserts one typed data row before index; index may equal RecordCount.
func (clock *RINEXClock) InsertRecord(index int, recordType RINEXClockRecordType, name string, epoch CivilDateTime, values []float64) error {
	if clock == nil || clock.handle == nil {
		return ErrClosed
	}
	return publicClockWriteError(clock.handle.InsertRecord(index, uint32(recordType), name, native.CivilDateTime{Year: epoch.Year, Month: epoch.Month, Day: epoch.Day, Hour: epoch.Hour, Minute: epoch.Minute, Second: epoch.Second}, append([]float64(nil), values...)))
}

// SetRecordValues replaces all declared numeric values for one record.
func (clock *RINEXClock) SetRecordValues(index int, values []float64) error {
	if clock == nil || clock.handle == nil {
		return ErrClosed
	}
	return publicClockWriteError(clock.handle.SetRecordValues(index, append([]float64(nil), values...)))
}

// SetRecordsValues atomically replaces numeric fields for several records.
// Duplicate indices with conflicting values are refused without changing rows.
func (clock *RINEXClock) SetRecordsValues(edits []RINEXClockRecordValuesEdit) (int, error) {
	if clock == nil || clock.handle == nil {
		return 0, ErrClosed
	}
	owned := make([]native.NativeClockRecordValuesEdit, len(edits))
	for i, edit := range edits {
		owned[i] = native.NativeClockRecordValuesEdit{Index: edit.Index, Values: append([]float64(nil), edit.Values...)}
	}
	n, err := clock.handle.SetRecordsValues(owned)
	return n, publicClockWriteError(err)
}

// RemoveRecord removes a data row and every physical line it spans.
func (clock *RINEXClock) RemoveRecord(index int) error {
	if clock == nil || clock.handle == nil {
		return ErrClosed
	}
	return publicClockWriteError(clock.handle.RemoveRecord(index))
}

// RemoveRecordWithValue removes a data row and returns its detached pre-removal
// value, including original epoch, numeric fields, line span, and record name.
func (clock *RINEXClock) RemoveRecordWithValue(index int) (RINEXClockRecord, bool, error) {
	if clock == nil || clock.handle == nil {
		return RINEXClockRecord{}, false, ErrClosed
	}
	v, present, err := clock.handle.RemoveRecordWithValue(index)
	if err != nil {
		return RINEXClockRecord{}, false, publicClockWriteError(err)
	}
	surplus := make([]RINEXClockSurplusValue, len(v.Surplus))
	for i, item := range v.Surplus {
		surplus[i] = RINEXClockSurplusValue{Position: item.Position, Value: item.Value}
	}
	return RINEXClockRecord{RecordType: RINEXClockRecordType(v.RecordType), Name: v.Name, HasSatellite: v.HasSatellite, SatelliteID: v.Satellite, CivilEpoch: CivilDateTime{Year: v.CivilEpoch.Year, Month: v.CivilEpoch.Month, Day: v.CivilEpoch.Day, Hour: v.CivilEpoch.Hour, Minute: v.CivilEpoch.Minute, Second: v.CivilEpoch.Second}, HasEpoch: v.HasEpoch, Epoch: ClockEpoch{Scale: TimeScale(v.Epoch.Scale), Representation: RINEXClockInstantRepresentation(v.Epoch.Representation), JulianWhole: v.Epoch.JulianWhole, JulianFraction: v.Epoch.JulianFraction, NanosHigh: v.Epoch.NanosHigh, NanosLow: v.Epoch.NanosLow}, Values: append([]float64(nil), v.Values...), Surplus: surplus, HasLine: v.HasLine, Line: v.Line, LineCount: v.LineCount, Reading: RINEXClockRecordReading(v.Reading), HasContinuationReading: v.HasContinuationReading, ContinuationReading: RINEXClockRecordReading(v.ContinuationReading), ReadingUnknownVariant: v.ReadingUnknownVariant, ContinuationReadingUnknownVariant: v.ContinuationReadingUnknownVariant}, present, nil
}

// RemoveRecords atomically removes record-order indices. Duplicate indices
// remove one row once, as in the native API.
func (clock *RINEXClock) RemoveRecords(indices []int) (int, error) {
	if clock == nil || clock.handle == nil {
		return 0, ErrClosed
	}
	n, err := clock.handle.RemoveRecords(append([]int(nil), indices...))
	return n, publicClockWriteError(err)
}

// SetTimeSystem replaces the header time system after validating every row epoch.
func (clock *RINEXClock) SetTimeSystem(system RINEXClockTimeSystem) error {
	if clock == nil || clock.handle == nil {
		return ErrClosed
	}
	return publicClockWriteError(clock.handle.SetTimeSystem(uint32(system)))
}

// Write serializes the product and returns all departures allowed by policy.
func (clock *RINEXClock) Write(policy RINEXClockWritePolicy) (RINEXClockWriteResult, error) {
	if clock == nil || clock.handle == nil {
		return RINEXClockWriteResult{}, ErrClosed
	}
	v, err := clock.handle.TextWithPolicy(policy.AllowNearestMicrosecondEpoch)
	if err != nil {
		return RINEXClockWriteResult{}, publicClockWriteError(err)
	}
	out := RINEXClockWriteResult{Text: append([]byte(nil), v.Text...), Departures: make([]RINEXClockWriteDeparture, len(v.Departures))}
	for i, d := range v.Departures {
		out.Departures[i] = RINEXClockWriteDeparture{Kind: RINEXClockWriteDepartureKind(d.Kind), Record: d.Record, HasEpoch: d.HasEpoch, Epoch: ClockEpoch{Scale: TimeScale(d.Epoch.Scale), Representation: RINEXClockInstantRepresentation(d.Epoch.Representation), JulianWhole: d.Epoch.JulianWhole, JulianFraction: d.Epoch.JulianFraction, NanosHigh: d.Epoch.NanosHigh, NanosLow: d.Epoch.NanosLow}, UnknownVariant: d.UnknownVariant, Name: d.Name, Written: d.Written}
	}
	return out, nil
}

// Diagnostics returns detached structured details for retained lossy findings.
func (clock *RINEXClock) Diagnostics() ([]RINEXClockDiagnostic, error) {
	if clock == nil || clock.handle == nil {
		return nil, ErrClosed
	}
	values, err := clock.handle.Diagnostics()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]RINEXClockDiagnostic, len(values))
	for i, v := range values {
		out[i] = RINEXClockDiagnostic{Kind: RINEXClockErrorKind(v.Kind), HasLine: true, Line: v.Line, HasErrorLine: v.HasErrorLine, ErrorLine: v.ErrorLine, HasTimeScale: v.HasTimeScale, TimeScale: TimeScale(v.TimeScale), HasField: v.HasField, HasReason: v.HasReason, HasRecord: v.HasRecord, HasRecordType: v.HasRecordType, HasValue: v.HasValue, Message: v.Message, Field: v.Field, Reason: v.Reason, Record: v.Record, RecordType: v.RecordType, Value: v.Value}
	}
	return out, nil
}

// Notices returns detached non-fatal parser findings in source order.
func (clock *RINEXClock) Notices() ([]RINEXClockNotice, error) {
	if clock == nil || clock.handle == nil {
		return nil, ErrClosed
	}
	values, err := clock.handle.Notices()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]RINEXClockNotice, len(values))
	for i, v := range values {
		out[i] = RINEXClockNotice{Kind: RINEXClockNoticeKind(v.Kind), HasTimeSystem: v.HasTimeSystem, TimeSystem: RINEXClockTimeSystem(v.TimeSystem), HasLine: v.HasLine, Line: v.Line, HasRecords: v.HasRecords, Records: v.Records, FirstLine: v.FirstLine, KindUnknownVariant: v.KindUnknownVariant, TimeSystemUnknownVariant: v.TimeSystemUnknownVariant}
	}
	return out, nil
}

// SkippedRecords returns detached non-AS data-record locations and types.
func (clock *RINEXClock) SkippedRecords() ([]RINEXClockSkippedRecord, error) {
	if clock == nil || clock.handle == nil {
		return nil, ErrClosed
	}
	values, err := clock.handle.SkippedRecords()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]RINEXClockSkippedRecord, len(values))
	for i, v := range values {
		out[i] = RINEXClockSkippedRecord{Line: v.Line, RecordType: RINEXClockRecordType(v.RecordType)}
	}
	return out, nil
}

// Close releases a child clock series and is safe to repeat.
func (series *RINEXClockSeries) Close() error {
	if series == nil || series.handle == nil {
		return nil
	}
	return publicError(series.handle.Close())
}

// Satellite returns the detached satellite identifier.
func (series *RINEXClockSeries) Satellite() (string, error) {
	if series == nil || series.handle == nil {
		return "", ErrClosed
	}
	value, err := series.handle.Satellite()
	return value, publicError(err)
}

// Samples returns detached clock samples for the series.
func (series *RINEXClockSeries) Samples() ([]ClockPoint, error) {
	if series == nil || series.handle == nil {
		return nil, ErrClosed
	}
	values, err := series.handle.Samples()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]ClockPoint, len(values))
	for index, value := range values {
		out[index] = ClockPoint{Epoch: ClockEpoch{Scale: TimeScale(value.Epoch.Scale), Representation: RINEXClockInstantRepresentation(value.Epoch.Representation), JulianWhole: value.Epoch.JulianWhole, JulianFraction: value.Epoch.JulianFraction, NanosHigh: value.Epoch.NanosHigh, NanosLow: value.Epoch.NanosLow}, BiasS: value.BiasS}
	}
	return out, nil
}

// SampleCount returns the number of samples in the series.
func (series *RINEXClockSeries) SampleCount() (int, error) {
	if series == nil || series.handle == nil {
		return 0, ErrClosed
	}
	value, err := series.handle.SampleCount()
	return value, publicError(err)
}

// JulianWhole and JulianFraction are used for JulianDate representation.
// NanosHigh and NanosLow are used for nanosecond representation.
// ADOTMPerS and DeltaN0DotRadPerS2 are CNAV orbital rates in named units.
// Top is the CNAV reference time; URA indices and transmission time preserve native values.
// HasFlags controls Flags.
// Toe and Toc are GNSS week/time tags; Elements and Clock contain orbital data.
// GroupDelays and CNAV contain optional signal-specific fields.
// HasFitInterval controls FitIntervalS, which is in seconds.
// Alpha and Beta arrays contain native ionospheric coefficients.
