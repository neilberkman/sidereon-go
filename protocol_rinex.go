package sidereon

import (
	"errors"

	"github.com/neilberkman/sidereon-go/internal/native"
)

// RINEXClockInstantRepresentation identifies how a clock instant is stored.
type RINEXClockInstantRepresentation uint32

const (
	RINEXClockInstantJulianDate RINEXClockInstantRepresentation = RINEXClockInstantRepresentation(native.RINEXClockInstantJulianDateValue)
	RINEXClockInstantNanos      RINEXClockInstantRepresentation = RINEXClockInstantRepresentation(native.RINEXClockInstantNanosValue)
)

// GNSSWeekTow is a scale-tagged GNSS week and time-of-week value.
type GNSSWeekTow struct {
	System     TimeScale
	Week       uint32
	TOWSeconds float64
}

// ClockEpoch preserves the C ABI's lossless clock-epoch representation.
type ClockEpoch struct {
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
	AF0    float64
	AF1    float64
	AF2    float64
	TocSOW float64
}

// BroadcastGroupDelays contains optional broadcast group-delay values. A
// Has... field distinguishes an absent value from a present zero.
type BroadcastGroupDelays struct {
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

// BroadcastRecord is one complete raw RINEX NAV record. Optional values are
// retained with explicit presence flags, matching the C value model.
type BroadcastRecord struct {
	SatelliteID    string
	Message        uint32
	Issue          uint32
	IssueMessage   uint32
	Week           uint32
	Toe            GNSSWeekTow
	Toc            GNSSWeekTow
	Elements       KeplerianElements
	Clock          ClockPolynomial
	GroupDelays    BroadcastGroupDelays
	CNAV           BroadcastCNAV
	SVHealth       float64
	SVAccuracyM    float64
	HasFitInterval bool
	FitIntervalS   float64
}

func broadcastRecordFromNative(value native.NativeBroadcastRecord) BroadcastRecord {
	return BroadcastRecord{
		SatelliteID: value.SatelliteID, Message: value.Message, Issue: value.Issue, IssueMessage: value.IssueMessage, Week: value.Week,
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
		SVHealth: value.SVHealth, SVAccuracyM: value.SVAccuracyM, HasFitInterval: value.HasFitInterval, FitIntervalS: value.FitIntervalS,
	}
}

func broadcastRecordToNative(value BroadcastRecord) native.NativeBroadcastRecord {
	return native.NativeBroadcastRecord{
		SatelliteID: value.SatelliteID, Message: value.Message, Issue: value.Issue, IssueMessage: value.IssueMessage, Week: value.Week,
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
		SVHealth: value.SVHealth, SVAccuracyM: value.SVAccuracyM, HasFitInterval: value.HasFitInterval, FitIntervalS: value.FitIntervalS,
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

// ParseRINEXClock strictly parses a RINEX clock product.
func ParseRINEXClock(data []byte) (*RINEXClock, error) { return parseRINEXClock(data, false) }

// ParseRINEXClockLossy skips malformed and non-AS rows according to C.
func ParseRINEXClockLossy(data []byte) (*RINEXClock, error) { return parseRINEXClock(data, true) }

func parseRINEXClock(data []byte, lossy bool) (*RINEXClock, error) {
	handle, err := native.ParseRinexClock(data, lossy)
	if err != nil {
		return nil, publicError(err)
	}
	return &RINEXClock{handle: handle}, nil
}

func (clock *RINEXClock) Close() error {
	if clock == nil || clock.handle == nil {
		return nil
	}
	return publicError(clock.handle.Close())
}

func (clock *RINEXClock) Satellites() ([]string, error) {
	if clock == nil || clock.handle == nil {
		return nil, ErrClosed
	}
	values, err := clock.handle.Satellites()
	return values, publicError(err)
}

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

func (clock *RINEXClock) SampleCount() (int, error) {
	if clock == nil || clock.handle == nil {
		return 0, ErrClosed
	}
	value, err := clock.handle.SampleCount()
	return value, publicError(err)
}

// Read-only methods may run concurrently with Close; Close waits for active
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

// ToText serializes the parsed product through the C writer.
func (clock *RINEXClock) ToText() ([]byte, error) {
	if clock == nil || clock.handle == nil {
		return nil, ErrClosed
	}
	value, err := clock.handle.Text()
	return value, publicError(err)
}

// Close releases a child clock series and is safe to repeat.
func (series *RINEXClockSeries) Close() error {
	if series == nil || series.handle == nil {
		return nil
	}
	return publicError(series.handle.Close())
}

func (series *RINEXClockSeries) Satellite() (string, error) {
	if series == nil || series.handle == nil {
		return "", ErrClosed
	}
	value, err := series.handle.Satellite()
	return value, publicError(err)
}

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

func (series *RINEXClockSeries) SampleCount() (int, error) {
	if series == nil || series.handle == nil {
		return 0, ErrClosed
	}
	value, err := series.handle.SampleCount()
	return value, publicError(err)
}
