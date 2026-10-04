package sidereon

import (
	"errors"
	"unsafe"

	"sidereon.dev/go/v3/internal/native"
)

// ANTEXErrorKind identifies a typed ANTEX parse or write refusal.
type ANTEXErrorKind uint32

const (
	// ANTEXErrorNone means no failure was recorded.
	ANTEXErrorNone ANTEXErrorKind = 0
	// ANTEXErrorInvalidDateTime identifies an invalid ANTEX GPS calendar value.
	ANTEXErrorInvalidDateTime ANTEXErrorKind = 1
	// ANTEXErrorInvalidField identifies malformed record-field content.
	ANTEXErrorInvalidField ANTEXErrorKind = 2
	// ANTEXErrorRepeatedRecord identifies conflicting repeated record content.
	ANTEXErrorRepeatedRecord ANTEXErrorKind = 3
	// ANTEXErrorDegenerateGrid identifies PCV coordinates that cannot form a grid.
	ANTEXErrorDegenerateGrid ANTEXErrorKind = 4
	// ANTEXErrorInvalidInput identifies refused caller input.
	ANTEXErrorInvalidInput ANTEXErrorKind = 5
	// ANTEXErrorUnknownFrequency identifies a missing frequency section.
	ANTEXErrorUnknownFrequency ANTEXErrorKind = 6
	// ANTEXErrorAmbiguousFrequency identifies differing duplicate frequency sections.
	ANTEXErrorAmbiguousFrequency ANTEXErrorKind = 7
	// ANTEXErrorMissingPCO identifies a frequency section without its required PCO.
	ANTEXErrorMissingPCO ANTEXErrorKind = 8
	// ANTEXErrorEmptyPCVGrid identifies a PCV grid without samples.
	ANTEXErrorEmptyPCVGrid ANTEXErrorKind = 9
	// ANTEXErrorUnwritable identifies content that cannot be encoded exactly.
	ANTEXErrorUnwritable ANTEXErrorKind = 10
)

// ANTEXError preserves the typed failure and each optional text detail.
type ANTEXError struct {
	// Kind identifies the native ANTEX refusal; unknown numeric values are retained.
	Kind ANTEXErrorKind
	// HasAntennaID reports whether AntennaID is present.
	HasAntennaID bool
	// HasRecord reports whether Record is present.
	HasRecord bool
	// HasField reports whether Field is present.
	HasField bool
	// HasValue reports whether Value is present.
	HasValue bool
	// HasFrequency reports whether Frequency is present.
	HasFrequency bool
	// HasReason reports whether Reason is present.
	HasReason bool
	// HasSections reports whether Sections is present.
	HasSections bool
	// Sections is the affected section count when HasSections is true.
	Sections int
	// AntennaID is the affected antenna identifier when HasAntennaID is true.
	AntennaID string
	// Record identifies the affected record when HasRecord is true.
	Record string
	// Field identifies the affected field when HasField is true.
	Field string
	// Value is the rejected field value when HasValue is true.
	Value string
	// Frequency identifies the affected frequency when HasFrequency is true.
	Frequency string
	// Reason contains the native refusal reason when HasReason is true.
	Reason string
	// Message contains the native human-readable summary.
	Message string
}

// ANTEXOutcome records the status of one typed parse or encode attempt.
type ANTEXOutcome struct {
	// IsOK reports whether the native operation succeeded.
	IsOK bool
	// Status is the numeric C ABI status for the operation.
	Status uint32
	// Error contains detached typed failure details, even when IsOK is false.
	Error ANTEXError
}

// ANTEXHeaderTextPart selects one retained reference-antenna header string.
type ANTEXHeaderTextPart uint32

const (
	// ANTEXHeaderTextType selects the antenna type text.
	ANTEXHeaderTextType ANTEXHeaderTextPart = iota
	// ANTEXHeaderTextSerial selects the antenna serial text.
	ANTEXHeaderTextSerial
	// ANTEXHeaderTextReference selects the effective reference antenna text.
	ANTEXHeaderTextReference
)

// ANTEXPCVType identifies absolute or relative calibration values.
type ANTEXPCVType uint32

const (
	// ANTEXPCVAbsolute identifies absolute phase-center values.
	ANTEXPCVAbsolute ANTEXPCVType = 0
	// ANTEXPCVRelative identifies values relative to a reference antenna.
	ANTEXPCVRelative ANTEXPCVType = 1
)

// ANTEXHeader contains the retained fixed-width product header fields.
type ANTEXHeader struct {
	// HasVersion reports whether ANTEX VERSION / SYST was present.
	HasVersion bool
	// Version is the version number when HasVersion is true.
	Version float64
	// HasSystem reports whether the system column was nonblank.
	HasSystem bool
	// System is the system code point when HasSystem is true.
	System uint32
	// HasPCVType reports whether PCV TYPE / REFANT was present.
	HasPCVType bool
	// PCVType is the absolute or relative type when HasPCVType is true.
	PCVType ANTEXPCVType
	// HasReferenceAntenna reports whether a reference antenna was retained.
	HasReferenceAntenna bool
	// HeaderCommentCount is the number of header COMMENT records.
	HeaderCommentCount int
	// EndOfHeader reports whether END OF HEADER was present.
	EndOfHeader bool
}

// ANTEXDateTime is an exact GPS calendar instant used to select validity
// intervals. FractionDigits/FractionScale represents FractionDigits divided
// by 10^FractionScale seconds.
type ANTEXDateTime struct {
	// Year is the four-digit calendar year.
	Year int32
	// Month is the calendar month, 1 through 12.
	Month uint8
	// Day is the calendar day of month.
	Day uint8
	// Hour is the hour of day, 0 through 23.
	Hour uint8
	// Minute is the minute of hour, 0 through 59.
	Minute uint8
	// Second is the whole second, 0 through 59.
	Second uint8
	// FractionDigits are the significant fractional-second digits.
	FractionDigits uint64
	// FractionScale is the power of ten dividing FractionDigits.
	FractionScale uint64
}

func publicANTEXOutcome(value native.AntexOutcome) ANTEXOutcome {
	e := value.Error
	return ANTEXOutcome{IsOK: value.IsOK, Status: value.Status, Error: ANTEXError{Kind: ANTEXErrorKind(e.Kind), HasAntennaID: e.HasAntennaID, HasRecord: e.HasRecord, HasField: e.HasField, HasValue: e.HasValue, HasFrequency: e.HasFrequency, HasReason: e.HasReason, HasSections: e.HasSections, Sections: e.Sections, AntennaID: e.AntennaID, Record: e.Record, Field: e.Field, Value: e.Value, Frequency: e.Frequency, Reason: e.Reason, Message: e.Message}}
}

// ParseANTEXWithOutcome parses bytes and returns a typed refusal as an outcome.
// Structural call failures are returned as error; a refused ANTEX document
// returns a nil product, a failed outcome, and a nil error.
func ParseANTEXWithOutcome(data []byte) (*ANTEX, ANTEXOutcome, error) {
	value, out, err := native.ParseANTEXWithOutcome(append([]byte(nil), data...))
	if err != nil {
		return nil, ANTEXOutcome{}, publicError(err)
	}
	var product *ANTEX
	if value != nil {
		product = &ANTEX{native: value}
	}
	return product, publicANTEXOutcome(out), nil
}

// EncodeWithOutcome serializes the product and retains any typed refusal.
func (a *ANTEX) EncodeWithOutcome() ([]byte, ANTEXOutcome, error) {
	if a == nil || a.native == nil {
		return nil, ANTEXOutcome{}, ErrClosed
	}
	text, out, err := a.native.EncodeWithOutcome()
	return text, publicANTEXOutcome(out), publicError(err)
}

// Header returns the typed header fields retained by the ANTEX parser.
func (a *ANTEX) Header() (ANTEXHeader, error) {
	if a == nil || a.native == nil {
		return ANTEXHeader{}, ErrClosed
	}
	v, err := a.native.Header()
	return ANTEXHeader{HasVersion: v.HasVersion, Version: v.Version, HasSystem: v.HasSystem, System: v.System, HasPCVType: v.HasPcvType, PCVType: ANTEXPCVType(v.PcvType), HasReferenceAntenna: v.HasReferenceAntenna, HeaderCommentCount: v.CommentCount, EndOfHeader: v.EndOfHeader}, publicError(err)
}

// HeaderText returns one retained reference-antenna text field: part 0 is the
// type, part 1 is the serial, and part 2 is the effective reference antenna.
func (a *ANTEX) HeaderText(part ANTEXHeaderTextPart) (string, error) {
	if a == nil || a.native == nil {
		return "", ErrClosed
	}
	value, err := a.native.HeaderText(uint32(part))
	return value, publicError(err)
}

// HeaderComment returns one retained header COMMENT line in file order.
func (a *ANTEX) HeaderComment(index int) (string, error) {
	if a == nil || a.native == nil {
		return "", ErrClosed
	}
	value, err := a.native.HeaderComment(index)
	return value, publicError(err)
}

// OuterCommentCount returns the number of COMMENT lines outside antenna blocks.
func (a *ANTEX) OuterCommentCount() (int, error) {
	if a == nil || a.native == nil {
		return 0, ErrClosed
	}
	value, err := a.native.OuterCommentCount()
	return value, publicError(err)
}

// OuterComment returns one between-block COMMENT line and the number of blocks
// before it in file order.
func (a *ANTEX) OuterComment(index int) (string, int, error) {
	if a == nil || a.native == nil {
		return "", 0, ErrClosed
	}
	text, before, err := a.native.OuterComment(index)
	return text, before, publicError(err)
}

// SkippedRecords returns the number of malformed or inconsistent records the
// forgiving ANTEX parser skipped.
func (a *ANTEX) SkippedRecords() (int, error) {
	if a == nil || a.native == nil {
		return 0, ErrClosed
	}
	value, err := a.native.SkippedRecords()
	return value, publicError(err)
}

// BlockCount returns the number of antenna validity blocks, including repeated
// ids with different validity intervals.
func (a *ANTEX) BlockCount() (int, error) {
	if a == nil || a.native == nil {
		return 0, ErrClosed
	}
	value, err := a.native.BlockCount()
	return value, publicError(err)
}

// Block returns an owned antenna handle for the zero-based block in file order.
func (a *ANTEX) Block(index int) (*Antenna, error) {
	if a == nil || a.native == nil {
		return nil, ErrClosed
	}
	value, err := a.native.Block(index)
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errors.New("sidereon: native ANTEX block lookup returned no handle")
	}
	return &Antenna{native: value}, nil
}

// AntennaAt returns an owned block for the exact antenna id at epoch. found is
// false when no validity interval covers epoch.
func (a *ANTEX) AntennaAt(id string, epoch ANTEXDateTime) (*Antenna, bool, error) {
	if a == nil || a.native == nil {
		return nil, false, ErrClosed
	}
	value, found, err := a.native.AntennaAt(id, native.AntexDateTime{Year: epoch.Year, Month: epoch.Month, Day: epoch.Day, Hour: epoch.Hour, Minute: epoch.Minute, Second: epoch.Second, FractionDigits: epoch.FractionDigits, FractionScale: epoch.FractionScale})
	if err != nil {
		return nil, false, publicError(err)
	}
	if !found {
		return nil, false, nil
	}
	if value == nil {
		return nil, false, errors.New("sidereon: native ANTEX time lookup returned no handle")
	}
	return &Antenna{native: value}, true, nil
}

// SatelliteAntenna returns an owned satellite calibration block valid at epoch.
// found is false when the ANTEX product has no matching validity interval.
func (a *ANTEX) SatelliteAntenna(prn string, epoch ANTEXDateTime) (*Antenna, bool, error) {
	if a == nil || a.native == nil {
		return nil, false, ErrClosed
	}
	value, found, err := a.native.SatelliteAntenna(prn, native.AntexDateTime{Year: epoch.Year, Month: epoch.Month, Day: epoch.Day, Hour: epoch.Hour, Minute: epoch.Minute, Second: epoch.Second, FractionDigits: epoch.FractionDigits, FractionScale: epoch.FractionScale})
	if err != nil {
		return nil, false, publicError(err)
	}
	if !found {
		return nil, false, nil
	}
	if value == nil {
		return nil, false, errors.New("sidereon: native ANTEX satellite lookup returned no handle")
	}
	return &Antenna{native: value}, true, nil
}

// AntennaPCO is a north/east/up phase-center offset in metres.
type AntennaPCO struct {
	// NorthM is the north m in metres.
	NorthM float64
	// EastM is the east m in metres.
	EastM float64
	// UpM is the up m in metres.
	UpM float64
}

// AntennaInfo retains fixed fields and presence/count metadata for one block.
type AntennaInfo struct {
	// Kind is the receiver or satellite role code.
	Kind ANTEXAntennaKind
	// HasDAZI reports whether the block declares an azimuth increment.
	HasDAZI bool
	// DAZIDeg is the declared azimuth increment in degrees.
	DAZIDeg float64
	// HasZenithGrid reports whether ZEN1/ZEN2/DZEN are present.
	HasZenithGrid bool
	// ZenithStartDeg is ZEN1 in degrees.
	ZenithStartDeg float64
	// ZenithEndDeg is ZEN2 in degrees.
	ZenithEndDeg float64
	// ZenithStepDeg is DZEN in degrees.
	ZenithStepDeg float64
	// HasFrequencyCountRecord reports whether # OF FREQUENCIES was present.
	HasFrequencyCountRecord bool
	// HasSINEXCode reports whether SINEX CODE was present.
	HasSINEXCode bool
	// HasValidFrom reports whether VALID FROM was present.
	HasValidFrom bool
	// ValidFrom is the exact lower validity bound when HasValidFrom is true.
	ValidFrom ANTEXDateTime
	// HasValidUntil reports whether VALID UNTIL was present.
	HasValidUntil bool
	// ValidUntil is the exact upper validity bound when HasValidUntil is true.
	ValidUntil ANTEXDateTime
	// CalibrationCount is the number of calibration records.
	CalibrationCount int
	// LeadingCommentCount counts comments before TYPE / SERIAL NO.
	LeadingCommentCount int
	// CommentCount counts comments after TYPE / SERIAL NO.
	CommentCount int
	// FrequencyCount counts all frequency sections in file order.
	FrequencyCount int
}

// ANTEXAntennaKind identifies the role of an antenna block.
type ANTEXAntennaKind uint32

const (
	// ANTEXAntennaReceiver identifies a receiver antenna.
	ANTEXAntennaReceiver ANTEXAntennaKind = 0
	// ANTEXAntennaSatellite identifies a satellite antenna.
	ANTEXAntennaSatellite ANTEXAntennaKind = 1
)

// ANTEXAntennaText selects a retained antenna text field.
type ANTEXAntennaText uint32

const (
	// ANTEXAntennaTextID selects the TYPE / SERIAL NO identifier.
	ANTEXAntennaTextID ANTEXAntennaText = iota
	// ANTEXAntennaTextType selects the antenna type field.
	ANTEXAntennaTextType
	// ANTEXAntennaTextSerial selects the serial number field.
	ANTEXAntennaTextSerial
	// ANTEXAntennaTextSINEX selects the SINEX code field.
	ANTEXAntennaTextSINEX
)

// ANTEXAntennaCommentList selects a retained comment list.
type ANTEXAntennaCommentList uint32

const (
	// ANTEXAntennaLeadingComments selects comments before TYPE / SERIAL NO.
	ANTEXAntennaLeadingComments ANTEXAntennaCommentList = iota
	// ANTEXAntennaBlockComments selects comments after TYPE / SERIAL NO.
	ANTEXAntennaBlockComments
)

// ANTEXCalibrationText selects a field from METH / BY / # / DATE.
type ANTEXCalibrationText uint32

const (
	// ANTEXCalibrationMethod selects the calibration method.
	ANTEXCalibrationMethod ANTEXCalibrationText = iota
	// ANTEXCalibrationAgency selects the calibration agency.
	ANTEXCalibrationAgency
	// ANTEXCalibrationDate selects the calibration date text.
	ANTEXCalibrationDate
)

// ANTEXCalibration contains the typed count from one calibration record.
type ANTEXCalibration struct {
	// HasAntennasCalibrated reports whether the I6 count is present.
	HasAntennasCalibrated bool
	// AntennasCalibrated is the number of antennas calibrated when present.
	AntennasCalibrated uint32
}

// ANTEXFrequencyInfo contains the numeric fields of one frequency section.
type ANTEXFrequencyInfo struct {
	// PCOM contains north, east, and up offsets for receiver antennas, or
	// X, Y, and Z body-frame offsets for satellite antennas, in metres.
	PCOM [3]float64
	// PCVSampleCount is the number of ordinary PCV samples.
	PCVSampleCount int
	// HasRMS reports whether an RMS section exists.
	HasRMS bool
	// HasRMSPCOM reports whether that section has an RMS PCO row.
	HasRMSPCOM bool
	// RMSPCOM contains RMS north, east, and up offsets for receiver antennas,
	// or X, Y, and Z body-frame offsets for satellite antennas, in metres.
	RMSPCOM [3]float64
	// RMSPCVSampleCount is the number of RMS PCV samples.
	RMSPCVSampleCount int
}

// ANTEXPCVSample is one sample in the retained frequency pattern.
type ANTEXPCVSample struct {
	// Grid identifies NOAZI or azimuth-dependent row kind.
	Grid ANTEXPCVGrid
	// HasAzimuth reports whether AzimuthDeg is meaningful.
	HasAzimuth bool
	// AzimuthDeg is the row azimuth in degrees.
	AzimuthDeg float64
	// ZenithDeg is the receiver zenith or satellite nadir angle in degrees.
	ZenithDeg float64
	// ValueM is the PCV value in metres.
	ValueM float64
}

// ANTEXPCVGrid identifies a retained PCV sample row form.
type ANTEXPCVGrid uint32

const (
	// ANTEXPCVGridNoAzimuth identifies a NOAZI row.
	ANTEXPCVGridNoAzimuth ANTEXPCVGrid = iota
	// ANTEXPCVGridAzimuth identifies a numeric-azimuth row.
	ANTEXPCVGridAzimuth
)

// Antenna owns one parsed ANTEX antenna block. Read calls may run
// concurrently; Close waits for active calls and is idempotent. The value
// must not be copied after first use.
type Antenna struct {
	_      noCopy
	native *native.Antenna
}

// ANTEX owns a parsed ANTEX 1.4 product. Read calls may run concurrently;
// Close waits for active calls and is idempotent. The value must not be copied
// after first use.
type ANTEX struct {
	_      noCopy
	native *native.ANTEX
}

// ParseANTEX parses ANTEX bytes with the native parser.
func ParseANTEX(data []byte) (*ANTEX, error) {
	value, err := native.ParseANTEX(data)
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errors.New("sidereon: native ANTEX constructor returned no handle")
	}
	return &ANTEX{native: value}, nil
}

// Close releases the parsed ANTEX product and is idempotent.
func (a *ANTEX) Close() error {
	if a == nil || a.native == nil {
		return nil
	}
	return publicError(a.native.Close())
}

// AntennaCount returns the number of antenna blocks in the product.
func (a *ANTEX) AntennaCount() (int, error) {
	if a == nil || a.native == nil {
		return 0, ErrClosed
	}
	value, err := a.native.AntennaCount()
	return value, publicError(err)
}

// Antenna looks up an exact ANTEX TYPE / SERIAL identifier. found is false
// when the C product has no matching block.
func (a *ANTEX) Antenna(id string) (*Antenna, bool, error) {
	if a == nil || a.native == nil {
		return nil, false, ErrClosed
	}
	value, found, err := a.native.Antenna(id)
	if err != nil {
		return nil, false, publicError(err)
	}
	if !found {
		return nil, false, nil
	}
	if value == nil {
		return nil, false, errors.New("sidereon: native ANTEX antenna lookup returned no handle")
	}
	return &Antenna{native: value}, true, nil
}

// Encode serializes the parsed ANTEX product into newly owned bytes.
func (a *ANTEX) Encode() ([]byte, error) {
	if a == nil || a.native == nil {
		return nil, ErrClosed
	}
	value, err := a.native.Encode()
	return value, publicError(err)
}

// Close releases the antenna block and is idempotent.
func (a *Antenna) Close() error {
	if a == nil || a.native == nil {
		return nil
	}
	return publicError(a.native.Close())
}

// Info returns detached fixed fields and source-presence/count information.
func (a *Antenna) Info() (AntennaInfo, error) {
	if a == nil || a.native == nil {
		return AntennaInfo{}, ErrClosed
	}
	v, err := a.native.Info()
	return AntennaInfo{Kind: ANTEXAntennaKind(v.Kind), HasDAZI: v.HasDazi, DAZIDeg: v.DaziDeg, HasZenithGrid: v.HasZenithGrid, ZenithStartDeg: v.ZenithStartDeg, ZenithEndDeg: v.ZenithEndDeg, ZenithStepDeg: v.ZenithStepDeg, HasFrequencyCountRecord: v.HasFrequencyCountRecord, HasSINEXCode: v.HasSinexCode, HasValidFrom: v.HasValidFrom, ValidFrom: ANTEXDateTime(v.ValidFrom), HasValidUntil: v.HasValidUntil, ValidUntil: ANTEXDateTime(v.ValidUntil), CalibrationCount: v.CalibrationCount, LeadingCommentCount: v.LeadingCommentCount, CommentCount: v.CommentCount, FrequencyCount: v.FrequencyCount}, publicError(err)
}

// Text returns one retained antenna identifier, type, serial, or SINEX code.
func (a *Antenna) Text(part ANTEXAntennaText) (string, error) {
	if a == nil || a.native == nil {
		return "", ErrClosed
	}
	v, e := a.native.Text(uint32(part))
	return v, publicError(e)
}

// Comment returns one detached antenna comment from the selected ordered list.
func (a *Antenna) Comment(list ANTEXAntennaCommentList, index int) (string, error) {
	if a == nil || a.native == nil {
		return "", ErrClosed
	}
	v, e := a.native.Comment(uint32(list), index)
	return v, publicError(e)
}

// Calibration returns fixed numeric fields from a calibration record.
func (a *Antenna) Calibration(index int) (ANTEXCalibration, error) {
	if a == nil || a.native == nil {
		return ANTEXCalibration{}, ErrClosed
	}
	v, e := a.native.Calibration(index)
	return ANTEXCalibration{HasAntennasCalibrated: v.HasAntennasCalibrated, AntennasCalibrated: v.AntennasCalibrated}, publicError(e)
}

// CalibrationText returns method, agency, or date text from a calibration record.
func (a *Antenna) CalibrationText(index int, part ANTEXCalibrationText) (string, error) {
	if a == nil || a.native == nil {
		return "", ErrClosed
	}
	v, e := a.native.CalibrationText(index, uint32(part))
	return v, publicError(e)
}

// Frequency returns numeric PCO/RMS fields and sample counts for one section.
func (a *Antenna) Frequency(index int) (ANTEXFrequencyInfo, error) {
	if a == nil || a.native == nil {
		return ANTEXFrequencyInfo{}, ErrClosed
	}
	v, e := a.native.Frequency(index)
	return ANTEXFrequencyInfo{PCOM: v.PCOM, PCVSampleCount: v.PCVSampleCount, HasRMS: v.HasRMS, HasRMSPCOM: v.HasRMSPCOM, RMSPCOM: v.RMSPCOM, RMSPCVSampleCount: v.RMSPCVSampleCount}, publicError(e)
}

// FrequencyLabel returns the exact retained label for one frequency section.
func (a *Antenna) FrequencyLabel(index int) (string, error) {
	if a == nil || a.native == nil {
		return "", ErrClosed
	}
	v, e := a.native.FrequencyLabel(index)
	return v, publicError(e)
}

// FrequencyPCVSamples returns detached samples in source row and token order.
func (a *Antenna) FrequencyPCVSamples(index int, rms bool) ([]ANTEXPCVSample, error) {
	if a == nil || a.native == nil {
		return nil, ErrClosed
	}
	v, e := a.native.FrequencyPCVSamples(index, rms)
	if e != nil {
		return nil, publicError(e)
	}
	out := make([]ANTEXPCVSample, len(v))
	for i, s := range v {
		out[i] = ANTEXPCVSample{Grid: ANTEXPCVGrid(s.Grid), HasAzimuth: s.HasAzimuth, AzimuthDeg: s.AzimuthDeg, ZenithDeg: s.ZenithDeg, ValueM: s.ValueM}
	}
	return out, nil
}

// ValidAt reports whether the exact GPS calendar instant is inside this block's validity interval.
func (a *Antenna) ValidAt(epoch ANTEXDateTime) (bool, error) {
	if a == nil || a.native == nil {
		return false, ErrClosed
	}
	v, e := a.native.ValidAt(native.AntexDateTime(epoch))
	return v, publicError(e)
}

// PCO returns the frequency-dependent phase-center offset in metres.
func (a *Antenna) PCO(frequency string) (AntennaPCO, error) {
	if a == nil || a.native == nil {
		return AntennaPCO{}, ErrClosed
	}
	value, err := a.native.PCO(frequency)
	return AntennaPCO{NorthM: value.NorthM, EastM: value.EastM, UpM: value.UpM}, publicError(err)
}

// PCV returns the frequency-dependent phase-center variation in metres.
func (a *Antenna) PCV(frequency string, zenithDeg float64, azimuthDeg *float64) (float64, error) {
	if a == nil || a.native == nil {
		return 0, ErrClosed
	}
	var azimuth float64
	if azimuthDeg != nil {
		azimuth = *azimuthDeg
	}
	value, err := a.native.PCV(frequency, zenithDeg, azimuthDeg != nil, azimuth)
	return value, publicError(err)
}

// SpaceWeather contains the three native solar/geomagnetic indices.
type SpaceWeather struct {
	// F107 is the 10.7 cm solar radio flux.
	F107 float64
	// F107A is the adjusted 10.7 cm solar radio flux.
	F107A float64
	// Ap is the planetary geomagnetic index.
	Ap float64
}

// SpaceWeatherObservationClass identifies the provenance of a table record.
// The values mirror the C source enum: observed, interpolated, daily
// predicted, and monthly predicted.
type SpaceWeatherObservationClass uint32

const (
	// SpaceWeatherObservationObserved identifies the space weather observation observed case.
	SpaceWeatherObservationObserved SpaceWeatherObservationClass = SpaceWeatherObservationClass(native.SpaceWeatherObservationObservedValue)
	// SpaceWeatherObservationInterpolated identifies the space weather observation interpolated case.
	SpaceWeatherObservationInterpolated SpaceWeatherObservationClass = SpaceWeatherObservationClass(native.SpaceWeatherObservationInterpolatedValue)
	// SpaceWeatherObservationDailyPredicted identifies the space weather observation daily predicted case.
	SpaceWeatherObservationDailyPredicted SpaceWeatherObservationClass = SpaceWeatherObservationClass(native.SpaceWeatherObservationDailyPredictedValue)
	// SpaceWeatherObservationMonthlyPredicted identifies the space weather observation monthly predicted case.
	SpaceWeatherObservationMonthlyPredicted SpaceWeatherObservationClass = SpaceWeatherObservationClass(native.SpaceWeatherObservationMonthlyPredictedValue)
)

// DefaultSpaceWeather returns the native quiet-Sun defaults.
func DefaultSpaceWeather() (SpaceWeather, error) {
	value, err := native.DefaultSpaceWeather()
	return SpaceWeather{F107: value.F107, F107A: value.F107A, Ap: value.Ap}, publicError(err)
}

// SpaceWeatherCoverage describes observed and predicted coverage boundaries.
type SpaceWeatherCoverage struct {
	// FirstJ2000S is the earliest table epoch in seconds from J2000.
	FirstJ2000S float64
	// HasLastObservedJ2000S reports whether LastObservedJ2000S contains an observed epoch.
	HasLastObservedJ2000S bool
	// LastObservedJ2000S is the last observed sample epoch in seconds from J2000 when HasLastObservedJ2000S is true.
	LastObservedJ2000S float64
	// HasLastDailyPredictedJ2000S reports whether LastDailyPredictedJ2000S contains a daily-predicted epoch.
	HasLastDailyPredictedJ2000S bool
	// LastDailyPredictedJ2000S is the last daily-predicted sample epoch in seconds from J2000 when HasLastDailyPredictedJ2000S is true.
	LastDailyPredictedJ2000S float64
	// EndJ2000S is the end of table coverage in seconds from J2000.
	EndJ2000S float64
}

// SpaceWeatherDay preserves one complete parsed daily or monthly record,
// including explicit presence for every optional field.
type SpaceWeatherDay struct {
	// Year is the calendar year.
	Year int
	// Month is the calendar month.
	Month uint8
	// Day is the calendar day.
	Day uint8
	// Class is the product or solution class.
	Class SpaceWeatherObservationClass
	// HasBSRN reports whether the has bsrn field is present.
	HasBSRN bool
	// BSRN is the Bartels solar rotation number.
	BSRN uint16
	// HasND reports whether the has nd field is present.
	HasND bool
	// ND is the geomagnetic storm-day count.
	ND uint8
	// HasKp reports whether the has kp field is present.
	HasKp [8]bool
	// Kp10 is the eight-slot 3-hour Kp index array, ordered from 00:00 through 21:00 UTC.
	Kp10 [8]uint16
	// HasKpSum10 reports whether the has kp sum10 field is present.
	HasKpSum10 bool
	// KpSum10 is the 10-index geomagnetic sum.
	KpSum10 uint16
	// HasAp reports whether the has ap field is present.
	HasAp [8]bool
	// Ap8 is the eight-slot 3-hour Ap index array, ordered from 00:00 through 21:00 UTC.
	Ap8 [8]uint16
	// HasApAvg reports whether the has ap avg field is present.
	HasApAvg bool
	// ApAvg is the average planetary geomagnetic index.
	ApAvg uint16
	// HasCP10 reports whether the has cp10 field is present.
	HasCP10 bool
	// CP10 is the ten-centimetre solar flux.
	CP10 uint8
	// HasC9 reports whether the has c9 field is present.
	HasC9 bool
	// C9 is the nine-centimetre solar flux.
	C9 uint8
	// HasISN reports whether the has isn field is present.
	HasISN bool
	// ISN is the international sunspot number.
	ISN uint16
	// HasFluxQualifier reports whether the has flux qualifier field is present.
	HasFluxQualifier bool
	// FluxQualifier is the solar-flux qualifier.
	FluxQualifier uint8
	// HasF107Obs reports whether the has f107 obs field is present.
	HasF107Obs bool
	// F107Obs is the observed daily 10.7 cm solar flux in solar flux units (SFU), when HasF107Obs is true.
	F107Obs float64
	// HasF107Adj reports whether the has f107 adj field is present.
	HasF107Adj bool
	// F107Adj is the adjusted daily 10.7 cm solar flux in SFU, when HasF107Adj is true.
	F107Adj float64
	// HasF107ObsCenter81 reports whether the has f107 obs center81 field is present.
	HasF107ObsCenter81 bool
	// F107ObsCenter81 is the 81-day centered mean of observed 10.7 cm flux in SFU, when HasF107ObsCenter81 is true.
	F107ObsCenter81 float64
	// HasF107ObsLast81 reports whether the has f107 obs last81 field is present.
	HasF107ObsLast81 bool
	// F107ObsLast81 is the preceding 81-day mean of observed 10.7 cm flux in SFU, when HasF107ObsLast81 is true.
	F107ObsLast81 float64
	// HasF107AdjCenter81 reports whether the has f107 adj center81 field is present.
	HasF107AdjCenter81 bool
	// F107AdjCenter81 is the 81-day centered mean of adjusted 10.7 cm flux in SFU, when HasF107AdjCenter81 is true.
	F107AdjCenter81 float64
	// HasF107AdjLast81 reports whether the has f107 adj last81 field is present.
	HasF107AdjLast81 bool
	// F107AdjLast81 is the preceding 81-day mean of adjusted 10.7 cm flux in SFU, when HasF107AdjLast81 is true.
	F107AdjLast81 float64
}

// SpaceWeatherSample is one policy-selected space-weather sample.
type SpaceWeatherSample struct {
	// Weather is the space-weather table.
	Weather SpaceWeather
	// Class is the product or solution class.
	Class SpaceWeatherObservationClass
	// ApDefaulted reports that the native parser supplied a default Ap value.
	ApDefaulted bool
}

// SpaceWeatherPolicy controls which observation classes a sample query accepts
// and whether a missing geomagnetic index can be substituted.
type SpaceWeatherPolicy struct {
	// AllowInterpolated permits interpolation between observed daily samples.
	AllowInterpolated bool
	// AllowDailyPredicted permits daily predicted samples when observed data are unavailable.
	AllowDailyPredicted bool
	// AllowMonthlyPredicted permits monthly predicted samples when observed and daily predicted data are unavailable.
	AllowMonthlyPredicted bool
	// RequireGeomagnetic requires a geomagnetic index for a successful sample query.
	RequireGeomagnetic bool
	// AllowNotObserved permits rows explicitly classified as having no observation.
	AllowNotObserved bool
}

// DefaultSpaceWeatherPolicy returns the native strict policy, which rejects rows marked not observed.
func DefaultSpaceWeatherPolicy() (SpaceWeatherPolicy, error) {
	value, err := native.DefaultSpaceWeatherPolicy()
	return SpaceWeatherPolicy{AllowNotObserved: value.AllowNotObserved, AllowInterpolated: value.AllowInterpolated, AllowDailyPredicted: value.AllowDailyPredicted, AllowMonthlyPredicted: value.AllowMonthlyPredicted, RequireGeomagnetic: value.RequireGeomagnetic}, publicError(err)
}

// LenientSpaceWeatherPolicy returns the native policy that accepts every row class and substitutes missing geomagnetic values.
func LenientSpaceWeatherPolicy() (SpaceWeatherPolicy, error) {
	value, err := native.LenientSpaceWeatherPolicy()
	return SpaceWeatherPolicy{AllowNotObserved: value.AllowNotObserved, AllowInterpolated: value.AllowInterpolated, AllowDailyPredicted: value.AllowDailyPredicted, AllowMonthlyPredicted: value.AllowMonthlyPredicted, RequireGeomagnetic: value.RequireGeomagnetic}, publicError(err)
}

// SpaceWeatherTableSummary contains parser record and diagnostic counts.
type SpaceWeatherTableSummary struct {
	// DayCount is the number of daily records loaded.
	DayCount int
	// MonthlyCount is the number of monthly records loaded.
	MonthlyCount int
	// SkipCount is the number of records skipped by the parser.
	SkipCount int
	// WarningCount is the number of parser warnings recorded.
	WarningCount int
}

// SpaceWeatherTable owns a parsed space-weather table. Read methods may run
// concurrently; Close waits for active methods and is idempotent. The value
// must not be copied after first use.
type SpaceWeatherTable struct {
	_      noCopy
	native *native.SpaceWeatherTable
}

func spaceWeatherTable(value *native.SpaceWeatherTable, err error) (*SpaceWeatherTable, error) {
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errors.New("sidereon: native space-weather constructor returned no handle")
	}
	return &SpaceWeatherTable{native: value}, nil
}

// ParseSpaceWeatherTable parses the native table format from bytes.
func ParseSpaceWeatherTable(data []byte) (*SpaceWeatherTable, error) {
	return spaceWeatherTable(native.ParseSpaceWeatherTable(data))
}

// ParseSpaceWeatherCSV parses the CSV table format from bytes.
func ParseSpaceWeatherCSV(data []byte) (*SpaceWeatherTable, error) {
	return spaceWeatherTable(native.ParseSpaceWeatherCSV(data))
}

// ParseSpaceWeatherTXT parses the text table format from bytes.
func ParseSpaceWeatherTXT(data []byte) (*SpaceWeatherTable, error) {
	return spaceWeatherTable(native.ParseSpaceWeatherTXT(data))
}

// Close releases the table and is idempotent.
func (t *SpaceWeatherTable) Close() error {
	if t == nil || t.native == nil {
		return nil
	}
	return publicError(t.native.Close())
}

// Days returns copied daily records in native table order.
func (t *SpaceWeatherTable) Days() ([]SpaceWeatherDay, error) {
	if t == nil || t.native == nil {
		return nil, ErrClosed
	}
	value, err := t.native.Days()
	return convertSpaceWeatherDays(value), publicError(err)
}

// Monthly returns copied monthly records in native table order.
func (t *SpaceWeatherTable) Monthly() ([]SpaceWeatherDay, error) {
	if t == nil || t.native == nil {
		return nil, ErrClosed
	}
	value, err := t.native.Monthly()
	return convertSpaceWeatherDays(value), publicError(err)
}

func convertSpaceWeatherDays(values []native.SpaceWeatherDay) []SpaceWeatherDay {
	result := make([]SpaceWeatherDay, len(values))
	for i, value := range values {
		result[i] = SpaceWeatherDay{Year: value.Year, Month: value.Month, Day: value.Day, Class: SpaceWeatherObservationClass(value.Class), HasBSRN: value.HasBSRN, BSRN: value.BSRN, HasND: value.HasND, ND: value.ND, HasKp: value.HasKp, Kp10: value.Kp10, HasKpSum10: value.HasKpSum10, KpSum10: value.KpSum10, HasAp: value.HasAp, Ap8: value.Ap8, HasApAvg: value.HasApAvg, ApAvg: value.ApAvg, HasCP10: value.HasCP10, CP10: value.CP10, HasC9: value.HasC9, C9: value.C9, HasISN: value.HasISN, ISN: value.ISN, HasFluxQualifier: value.HasFluxQualifier, FluxQualifier: value.FluxQualifier, HasF107Obs: value.HasF107Obs, F107Obs: value.F107Obs, HasF107Adj: value.HasF107Adj, F107Adj: value.F107Adj, HasF107ObsCenter81: value.HasF107ObsCenter81, F107ObsCenter81: value.F107ObsCenter81, HasF107ObsLast81: value.HasF107ObsLast81, F107ObsLast81: value.F107ObsLast81, HasF107AdjCenter81: value.HasF107AdjCenter81, F107AdjCenter81: value.F107AdjCenter81, HasF107AdjLast81: value.HasF107AdjLast81, F107AdjLast81: value.F107AdjLast81}
	}
	return result
}

// Day returns a copied record and whether the requested date is present.
func (t *SpaceWeatherTable) Day(year int, month, day uint8) (SpaceWeatherDay, bool, error) {
	if t == nil || t.native == nil {
		return SpaceWeatherDay{}, false, ErrClosed
	}
	value, present, err := t.native.Day(year, month, day)
	return convertSpaceWeatherDays([]native.SpaceWeatherDay{value})[0], present, publicError(err)
}

// Coverage returns table coverage with explicit presence for optional bounds.
func (t *SpaceWeatherTable) Coverage() (SpaceWeatherCoverage, error) {
	if t == nil || t.native == nil {
		return SpaceWeatherCoverage{}, ErrClosed
	}
	value, err := t.native.Coverage()
	return SpaceWeatherCoverage{FirstJ2000S: value.FirstJ2000S, HasLastObservedJ2000S: value.HasLastObservedJ2000S, LastObservedJ2000S: value.LastObservedJ2000S, HasLastDailyPredictedJ2000S: value.HasLastDailyPredictedS, LastDailyPredictedJ2000S: value.LastDailyPredictedJ2000S, EndJ2000S: value.EndJ2000S}, publicError(err)
}

// Summary returns day, monthly, skipped-record, and warning counts.
func (t *SpaceWeatherTable) Summary() (SpaceWeatherTableSummary, error) {
	if t == nil || t.native == nil {
		return SpaceWeatherTableSummary{}, ErrClosed
	}
	value, err := t.native.Summary()
	return SpaceWeatherTableSummary{DayCount: value.DayCount, MonthlyCount: value.MonthlyCount, SkipCount: value.SkipCount, WarningCount: value.WarningCount}, publicError(err)
}

// APArrayAt returns the seven native Ap-history values at an epoch.
func (t *SpaceWeatherTable) APArrayAt(epochJ2000S float64) ([7]float64, error) {
	if t == nil || t.native == nil {
		return [7]float64{}, ErrClosed
	}
	value, err := t.native.APArrayAt(epochJ2000S)
	return value, publicError(err)
}

// SampleAt samples the native table using default policy.
func (t *SpaceWeatherTable) SampleAt(epochJ2000S float64) (SpaceWeatherSample, error) {
	if t == nil || t.native == nil {
		return SpaceWeatherSample{}, ErrClosed
	}
	value, err := t.native.SampleAt(epochJ2000S)
	return SpaceWeatherSample{Weather: SpaceWeather{F107: value.Weather.F107, F107A: value.Weather.F107A, Ap: value.Weather.Ap}, Class: SpaceWeatherObservationClass(value.Class), ApDefaulted: value.ApDefaulted}, publicError(err)
}

// SampleAtWithPolicy samples the table under explicit acceptance policy.
func (t *SpaceWeatherTable) SampleAtWithPolicy(epochJ2000S float64, policy SpaceWeatherPolicy) (SpaceWeatherSample, error) {
	if t == nil || t.native == nil {
		return SpaceWeatherSample{}, ErrClosed
	}
	value, err := t.native.SampleAtWithPolicy(epochJ2000S, native.SpaceWeatherPolicy{AllowNotObserved: policy.AllowNotObserved, AllowInterpolated: policy.AllowInterpolated, AllowDailyPredicted: policy.AllowDailyPredicted, AllowMonthlyPredicted: policy.AllowMonthlyPredicted, RequireGeomagnetic: policy.RequireGeomagnetic})
	return SpaceWeatherSample{Weather: SpaceWeather{F107: value.Weather.F107, F107A: value.Weather.F107A, Ap: value.Weather.Ap}, Class: SpaceWeatherObservationClass(value.Class), ApDefaulted: value.ApDefaulted}, publicError(err)
}

// SpaceWeatherAt samples only the three weather indices.
func (t *SpaceWeatherTable) SpaceWeatherAt(epochJ2000S float64) (SpaceWeather, error) {
	if t == nil || t.native == nil {
		return SpaceWeather{}, ErrClosed
	}
	value, err := t.native.SpaceWeatherAt(epochJ2000S)
	return SpaceWeather{F107: value.F107, F107A: value.F107A, Ap: value.Ap}, publicError(err)
}

// ToCSV serializes the table to newly owned CSV bytes.
func (t *SpaceWeatherTable) ToCSV() ([]byte, error) {
	if t == nil || t.native == nil {
		return nil, ErrClosed
	}
	value, err := t.native.ToCSV()
	return value, publicError(err)
}

// ToTXT serializes the table to newly owned text bytes.
func (t *SpaceWeatherTable) ToTXT() ([]byte, error) {
	if t == nil || t.native == nil {
		return nil, ErrClosed
	}
	value, err := t.native.ToTXT()
	return value, publicError(err)
}

// GeofenceErrorKind identifies the native diagnostic category accompanying a
// geofence operation.
type GeofenceErrorKind uint32

const (
	// GeofenceErrorNone identifies the geofence error none case.
	GeofenceErrorNone GeofenceErrorKind = GeofenceErrorKind(native.GeofenceErrorNoneValue)
	// GeofenceErrorTooFewVertices identifies the geofence error too few vertices case.
	GeofenceErrorTooFewVertices GeofenceErrorKind = GeofenceErrorKind(native.GeofenceErrorTooFewVerticesValue)
	// GeofenceErrorInvalidInput identifies the geofence error invalid input case.
	GeofenceErrorInvalidInput GeofenceErrorKind = GeofenceErrorKind(native.GeofenceErrorInvalidInputValue)
	// GeofenceErrorGeodesic identifies the geofence error geodesic case.
	GeofenceErrorGeodesic GeofenceErrorKind = GeofenceErrorKind(native.GeofenceErrorGeodesicValue)
	// GeofenceErrorDOP identifies a DOP calculation error.
	GeofenceErrorDOP GeofenceErrorKind = GeofenceErrorKind(native.GeofenceErrorDOPValue)
	// GeofenceErrorMetrics identifies the geofence error metrics case.
	GeofenceErrorMetrics GeofenceErrorKind = GeofenceErrorKind(native.GeofenceErrorMetricsValue)
)

// GeofenceUncertaintyKind identifies an uncertainty representation.
type GeofenceUncertaintyKind uint32

const (
	// GeofenceENUCovarianceM2 identifies covariance in the local ENU frame, in square metres.
	GeofenceENUCovarianceM2 GeofenceUncertaintyKind = GeofenceUncertaintyKind(native.GeofenceENUCovarianceM2Value)
	// GeofenceECEFCovarianceM2 identifies covariance in the ECEF frame, in square metres.
	GeofenceECEFCovarianceM2 GeofenceUncertaintyKind = GeofenceUncertaintyKind(native.GeofenceECEFCovarianceM2Value)
	// GeofenceCEPRadiusM identifies a circular error probable (CEP) radius in metres.
	GeofenceCEPRadiusM GeofenceUncertaintyKind = GeofenceUncertaintyKind(native.GeofenceCEPRadiusMValue)
)

// GeofenceProbabilityMethod selects native probability integration.
type GeofenceProbabilityMethod uint32

const (
	// GeofenceBoundaryNormal identifies the geofence boundary normal case.
	GeofenceBoundaryNormal GeofenceProbabilityMethod = GeofenceProbabilityMethod(native.GeofenceBoundaryNormalValue)
	// GeofencePlanarQuadrature identifies the geofence planar quadrature case.
	GeofencePlanarQuadrature GeofenceProbabilityMethod = GeofenceProbabilityMethod(native.GeofencePlanarQuadratureValue)
)

// GeofenceCrossingKind identifies an entered or left event.
type GeofenceCrossingKind uint32

const (
	// GeofenceEntered identifies the geofence entered case.
	GeofenceEntered GeofenceCrossingKind = GeofenceCrossingKind(native.GeofenceEnteredValue)
	// GeofenceLeft identifies the geofence left case.
	GeofenceLeft GeofenceCrossingKind = GeofenceCrossingKind(native.GeofenceLeftValue)
)

// GeofenceUncertainty describes covariance or CEP-radius uncertainty.
type GeofenceUncertainty struct {
	// Kind is the event or record kind.
	Kind GeofenceUncertaintyKind
	// CovarianceM2 is the covariance m2 in square metres.
	CovarianceM2 Matrix3
	// RadiusM is the radius m in metres.
	RadiusM float64
}

// GeofenceProbabilityOptions controls probability integration.
type GeofenceProbabilityOptions struct {
	// Method is the selected method.
	Method GeofenceProbabilityMethod
}

// GeofencePositionEstimate is one probabilistic trajectory sample.
type GeofencePositionEstimate struct {
	// Position is the position value in the containing frame.
	Position Geodetic
	// Uncertainty contains the position uncertainty representation for this geofence sample.
	Uncertainty GeofenceUncertainty
}

// GeofenceHysteresis contains entered and left confidence thresholds.
type GeofenceHysteresis struct {
	// EnterConfidence is the probability threshold for declaring entry.
	EnterConfidence float64
	// LeaveConfidence is the probability threshold for declaring exit.
	LeaveConfidence float64
}

// GeofenceCrossingEvent is one copied native crossing event.
type GeofenceCrossingEvent struct {
	// SampleIndex is the zero-based sample index associated with this crossing.
	SampleIndex int
	// Kind is the event or record kind.
	Kind GeofenceCrossingKind
	// InsideProbability is the probability that the sampled position is inside the geofence.
	InsideProbability float64
}

// Geofence owns a native WGS 84 geodesic polygon. Read calls may run
// concurrently; Close waits for active calls and is idempotent. The value
// must not be copied after first use.
type Geofence struct {
	_      noCopy
	native *native.Geofence
}

// NewGeofence creates a geodesic polygon and returns its native diagnostic
// category as well as any status error.
func NewGeofence(vertices []Geodetic) (*Geofence, GeofenceErrorKind, error) {
	if err := checkedEnvironmentAllocation(len(vertices), unsafe.Sizeof(native.Geodetic{})); err != nil {
		return nil, GeofenceErrorNone, err
	}
	nativeVertices := make([]native.Geodetic, len(vertices))
	for i, vertex := range vertices {
		nativeVertices[i] = native.Geodetic{LatitudeRad: vertex.LatitudeRad, LongitudeRad: vertex.LongitudeRad, HeightM: vertex.HeightM}
	}
	value, detail, err := native.GeofenceCreate(nativeVertices)
	if err != nil {
		return nil, GeofenceErrorKind(detail), publicError(err)
	}
	if value == nil {
		return nil, GeofenceErrorKind(detail), errors.New("sidereon: native geofence constructor returned no handle")
	}
	return &Geofence{native: value}, GeofenceErrorKind(detail), nil
}

// Close releases the native polygon and is idempotent.
func (f *Geofence) Close() error {
	if f == nil || f.native == nil {
		return nil
	}
	return publicError(f.native.Close())
}

// Contains reports whether a position is inside and returns the native
// diagnostic category.
func (f *Geofence) Contains(position Geodetic) (bool, GeofenceErrorKind, error) {
	if f == nil || f.native == nil {
		return false, GeofenceErrorNone, ErrClosed
	}
	value, detail, err := f.native.Contains(native.Geodetic{LatitudeRad: position.LatitudeRad, LongitudeRad: position.LongitudeRad, HeightM: position.HeightM})
	return value, GeofenceErrorKind(detail), publicError(err)
}

// DistanceToBoundary returns signed metres, positive inside, and the native
// diagnostic category.
func (f *Geofence) DistanceToBoundary(position Geodetic) (float64, GeofenceErrorKind, error) {
	if f == nil || f.native == nil {
		return 0, GeofenceErrorNone, ErrClosed
	}
	value, detail, err := f.native.DistanceToBoundary(native.Geodetic{LatitudeRad: position.LatitudeRad, LongitudeRad: position.LongitudeRad, HeightM: position.HeightM})
	return value, GeofenceErrorKind(detail), publicError(err)
}

func geofenceUncertainty(value GeofenceUncertainty) native.GeofenceUncertainty {
	var covariance [9]float64
	for row := range value.CovarianceM2 {
		for column := range value.CovarianceM2[row] {
			covariance[row*3+column] = value.CovarianceM2[row][column]
		}
	}
	return native.GeofenceUncertainty{Kind: uint32(value.Kind), Covariance: covariance, RadiusM: value.RadiusM}
}

// ContainmentProbability returns the native probability and diagnostic.
func (f *Geofence) ContainmentProbability(position Geodetic, uncertainty GeofenceUncertainty) (float64, GeofenceErrorKind, error) {
	return f.containmentProbability(position, uncertainty, nil)
}

// ContainmentProbabilityWithOptions returns probability under explicit
// integration options.
func (f *Geofence) ContainmentProbabilityWithOptions(position Geodetic, uncertainty GeofenceUncertainty, options GeofenceProbabilityOptions) (float64, GeofenceErrorKind, error) {
	return f.containmentProbability(position, uncertainty, &options)
}

func (f *Geofence) containmentProbability(position Geodetic, uncertainty GeofenceUncertainty, options *GeofenceProbabilityOptions) (float64, GeofenceErrorKind, error) {
	if f == nil || f.native == nil {
		return 0, GeofenceErrorNone, ErrClosed
	}
	if !validGeofenceUncertaintyKind(uncertainty.Kind) {
		return 0, GeofenceErrorInvalidInput, errors.New("sidereon: invalid geofence uncertainty kind")
	}
	if options != nil && !validGeofenceProbabilityMethod(options.Method) {
		return 0, GeofenceErrorInvalidInput, errors.New("sidereon: invalid geofence probability method")
	}
	nativePosition := native.Geodetic{LatitudeRad: position.LatitudeRad, LongitudeRad: position.LongitudeRad, HeightM: position.HeightM}
	nativeUncertainty := geofenceUncertainty(uncertainty)
	var value float64
	var detail uint32
	var err error
	if options == nil {
		value, detail, err = f.native.ContainmentProbability(nativePosition, nativeUncertainty)
	} else {
		value, detail, err = f.native.ContainmentProbabilityWithOptions(nativePosition, nativeUncertainty, native.GeofenceProbabilityOptions{Method: uint32(options.Method)})
	}
	return value, GeofenceErrorKind(detail), publicError(err)
}

func geofenceEvents(values []native.GeofenceCrossingEvent) []GeofenceCrossingEvent {
	if len(values) == 0 {
		return nil
	}
	result := make([]GeofenceCrossingEvent, len(values))
	for i, value := range values {
		result[i] = GeofenceCrossingEvent{SampleIndex: value.SampleIndex, Kind: GeofenceCrossingKind(value.Kind), InsideProbability: value.InsideProbability}
	}
	return result
}

// DefaultGeofenceProbabilityOptions returns native probability defaults.
func DefaultGeofenceProbabilityOptions() (GeofenceProbabilityOptions, error) {
	value, err := native.DefaultGeofenceProbabilityOptions()
	return GeofenceProbabilityOptions{Method: GeofenceProbabilityMethod(value.Method)}, publicError(err)
}

// DefaultGeofenceHysteresis returns native hysteresis defaults.
func DefaultGeofenceHysteresis() (GeofenceHysteresis, error) {
	value, err := native.DefaultGeofenceHysteresis()
	return GeofenceHysteresis{EnterConfidence: value.EnterConfidence, LeaveConfidence: value.LeaveConfidence}, publicError(err)
}

// CrossingProbability returns copied native crossing events and a diagnostic.
func (f *Geofence) CrossingProbability(samples []GeofencePositionEstimate, hysteresis GeofenceHysteresis) ([]GeofenceCrossingEvent, GeofenceErrorKind, error) {
	return f.crossingProbability(samples, hysteresis, nil)
}

// CrossingProbabilityWithOptions returns crossing events under explicit
// probability options.
func (f *Geofence) CrossingProbabilityWithOptions(samples []GeofencePositionEstimate, hysteresis GeofenceHysteresis, options GeofenceProbabilityOptions) ([]GeofenceCrossingEvent, GeofenceErrorKind, error) {
	return f.crossingProbability(samples, hysteresis, &options)
}

func (f *Geofence) crossingProbability(samples []GeofencePositionEstimate, hysteresis GeofenceHysteresis, options *GeofenceProbabilityOptions) ([]GeofenceCrossingEvent, GeofenceErrorKind, error) {
	if f == nil || f.native == nil {
		return nil, GeofenceErrorNone, ErrClosed
	}
	if err := checkedEnvironmentAllocation(len(samples), unsafe.Sizeof(native.GeofencePositionEstimate{})); err != nil {
		return nil, GeofenceErrorInvalidInput, err
	}
	if options != nil && !validGeofenceProbabilityMethod(options.Method) {
		return nil, GeofenceErrorInvalidInput, errors.New("sidereon: invalid geofence probability method")
	}
	nativeSamples := make([]native.GeofencePositionEstimate, len(samples))
	for i, sample := range samples {
		if !validGeofenceUncertaintyKind(sample.Uncertainty.Kind) {
			return nil, GeofenceErrorInvalidInput, errors.New("sidereon: invalid geofence uncertainty kind")
		}
		nativeSamples[i] = native.GeofencePositionEstimate{Position: native.Geodetic{LatitudeRad: sample.Position.LatitudeRad, LongitudeRad: sample.Position.LongitudeRad, HeightM: sample.Position.HeightM}, Uncertainty: geofenceUncertainty(sample.Uncertainty)}
	}
	nativeHysteresis := native.GeofenceHysteresis{EnterConfidence: hysteresis.EnterConfidence, LeaveConfidence: hysteresis.LeaveConfidence}
	var value []native.GeofenceCrossingEvent
	var detail uint32
	var err error
	if options == nil {
		value, detail, err = f.native.CrossingProbability(nativeSamples, nativeHysteresis)
	} else {
		value, detail, err = f.native.CrossingProbabilityWithOptions(nativeSamples, nativeHysteresis, native.GeofenceProbabilityOptions{Method: uint32(options.Method)})
	}
	return geofenceEvents(value), GeofenceErrorKind(detail), publicError(err)
}

// ObservabilityTier identifies a geometry-quality tier.
type ObservabilityTier uint32

const (
	// ObservabilityRankDeficient identifies the observability rank deficient case.
	ObservabilityRankDeficient ObservabilityTier = ObservabilityTier(native.ObservabilityRankDeficientValue)
	// ObservabilityZeroRedundancy identifies the observability zero redundancy case.
	ObservabilityZeroRedundancy ObservabilityTier = ObservabilityTier(native.ObservabilityZeroRedundancyValue)
	// ObservabilityWeak identifies the observability weak case.
	ObservabilityWeak ObservabilityTier = ObservabilityTier(native.ObservabilityWeakValue)
	// ObservabilityNominal identifies the observability nominal case.
	ObservabilityNominal ObservabilityTier = ObservabilityTier(native.ObservabilityNominalValue)
)

// ObservabilityTierLabel returns the native stable lowercase tier label.
func ObservabilityTierLabel(tier ObservabilityTier) (string, error) {
	if !validObservabilityTier(tier) {
		return "", errors.New("sidereon: invalid observability tier")
	}
	value, err := native.ObservabilityTierLabel(uint32(tier))
	return string(value), publicError(err)
}
