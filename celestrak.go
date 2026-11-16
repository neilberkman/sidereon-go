package sidereon

import (
	"encoding/json"
	"errors"
	"reflect"
	"sidereon.dev/go/v3/internal/native"
)

// CelestrakGPFormat selects the GP response format.
type CelestrakGPFormat uint32

const (
	// CelestrakGPFormatTLE selects named TLE records.
	CelestrakGPFormatTLE CelestrakGPFormat = iota
	// CelestrakGPFormatThreeLine selects three-line records.
	CelestrakGPFormatThreeLine
	// CelestrakGPFormatTwoLine selects two-line records.
	CelestrakGPFormatTwoLine
	// CelestrakGPFormatXML selects XML.
	CelestrakGPFormatXML
	// CelestrakGPFormatKVN selects KVN.
	CelestrakGPFormatKVN
	// CelestrakGPFormatJSON selects JSON.
	CelestrakGPFormatJSON
	// CelestrakGPFormatJSONPretty selects formatted JSON.
	CelestrakGPFormatJSONPretty
	// CelestrakGPFormatCSV selects CSV.
	CelestrakGPFormatCSV
)

// CelestrakGPQuery is a CelesTrak GP selector.
type CelestrakGPQuery interface {
	gpURL(CelestrakGPFormat) (string, error)
}

// CelestrakGPGroup selects a CelesTrak group.
type CelestrakGPGroup string

// CelestrakGPCatalogNumber selects a catalog number.
type CelestrakGPCatalogNumber uint64

// CelestrakGPInternationalDesignator selects a launch designator.
type CelestrakGPInternationalDesignator string

// CelestrakGPName selects a partial object name.
type CelestrakGPName string

func (v CelestrakGPGroup) gpURL(f CelestrakGPFormat) (string, error) {
	s, e := native.CelestrakText(5, 0, string(v), 0, uint32(f))
	return s, publicCatalogError(e)
}
func (v CelestrakGPCatalogNumber) gpURL(f CelestrakGPFormat) (string, error) {
	s, e := native.CelestrakText(6, 0, "", uint64(v), uint32(f))
	return s, publicCatalogError(e)
}
func (v CelestrakGPInternationalDesignator) gpURL(f CelestrakGPFormat) (string, error) {
	s, e := native.CelestrakText(5, 2, string(v), 0, uint32(f))
	return s, publicCatalogError(e)
}
func (v CelestrakGPName) gpURL(f CelestrakGPFormat) (string, error) {
	s, e := native.CelestrakText(5, 3, string(v), 0, uint32(f))
	return s, publicCatalogError(e)
}
func (v CelestrakGPSpecial) gpURL(f CelestrakGPFormat) (string, error) {
	s, e := native.CelestrakText(7, uint32(v), "", 0, uint32(f))
	return s, publicCatalogError(e)
}

// CelestrakGPSpecial selects a special GP dataset.
type CelestrakGPSpecial uint32

const (
	// CelestrakGPSpecialGPZ selects the GEO protected zone.
	CelestrakGPSpecialGPZ CelestrakGPSpecial = iota
	// CelestrakGPSpecialGPZPlus selects the extended zone.
	CelestrakGPSpecialGPZPlus
	// CelestrakGPSpecialDecaying selects decaying objects.
	CelestrakGPSpecialDecaying
)

// CatalogSourceEntry identifies a source. Values compare by value.
type CatalogSourceEntry struct {
	// Protocol is the source protocol.
	Protocol string `json:"protocol"`
	// Host is the source host.
	Host string `json:"host"`
	// RootURL is the source root URL.
	RootURL string `json:"root_url"`
}

func catalogSourceEntry(operation uint32) (CatalogSourceEntry, error) {
	var entry CatalogSourceEntry
	s, err := native.CelestrakText(operation, 0, "", 0, 0)
	if err != nil {
		return entry, publicCatalogError(err)
	}
	err = json.Unmarshal([]byte(s), &entry)
	return entry, err
}

// CelestrakGPSourceEntry returns the GP source metadata.
func CelestrakGPSourceEntry() (CatalogSourceEntry, error) { return catalogSourceEntry(0) }

// NAVCENSourceEntry returns NAVCEN source metadata.
func NAVCENSourceEntry() (CatalogSourceEntry, error) { return catalogSourceEntry(1) }

// NAVCENConstellationURL returns the NAVCEN constellation URL.
func NAVCENConstellationURL() (string, error) {
	s, e := native.CelestrakText(2, 0, "", 0, 0)
	return s, publicCatalogError(e)
}

// CelestrakGroup returns the GP group for a constellation.
func CelestrakGroup(system GNSSSystem) (string, error) {
	s, e := native.CelestrakText(3, uint32(system), "", 0, 0)
	return s, publicCatalogError(e)
}

// CelestrakConstellationURL returns a constellation GP URL.
func CelestrakConstellationURL(system GNSSSystem, format CelestrakGPFormat) (string, error) {
	s, e := native.CelestrakText(4, uint32(system), "", 0, uint32(format))
	return s, publicCatalogError(e)
}

// CelestrakGPURL builds a URL for the supplied GP selector.
func CelestrakGPURL(query CelestrakGPQuery, format CelestrakGPFormat) (string, error) {
	if query == nil || (reflect.ValueOf(query).Kind() == reflect.Ptr && reflect.ValueOf(query).IsNil()) {
		return "", &StatusError{Code: StatusInvalidArgument, Text: "invalid argument", Detail: "GP query is nil"}
	}
	return query.gpURL(format)
}

// CelestrakGPResponse owns parsed GP records.
type CelestrakGPResponse struct {
	_      noCopy
	handle *native.CelestrakGpResponse
}

// ParseCelestrakGPResponse parses a GP response in the specified format.
func ParseCelestrakGPResponse(data []byte, format CelestrakGPFormat) (*CelestrakGPResponse, error) {
	value, err := native.ParseCelestrakGpResponse(data, uint32(format))
	if err != nil {
		return nil, publicCatalogError(err)
	}
	return &CelestrakGPResponse{handle: value}, nil
}

// Close releases the parsed response.
func (r *CelestrakGPResponse) Close() error {
	if r == nil || r.handle == nil {
		return nil
	}
	return publicError(r.handle.Close())
}

// Counts returns the accepted and rejected record counts.
func (r *CelestrakGPResponse) Counts() (int, int, error) {
	if r == nil || r.handle == nil {
		return 0, 0, ErrClosed
	}
	a, b, e := r.handle.Counts()
	return a, b, publicCatalogError(e)
}

// CelestrakGPPayload describes a native GP payload.
type CelestrakGPPayload struct {
	// Kind identifies the payload format.
	Kind string `json:"kind"`
	// TLE holds the TLE payload when Kind is tle.
	TLE *CelestrakTLERecord `json:"-"`
}

// CelestrakTLERecord holds a detached TLE payload. Its fields compare by value with reflect.DeepEqual.
type CelestrakTLERecord struct {
	// Elements contains the parsed TLE elements.
	Elements TLEMetadata
	// Line1 is the first TLE line.
	Line1 string
	// Line2 is the second TLE line.
	Line2 string
	// MeanMotionDoubleDotText retains the TLE field text.
	MeanMotionDoubleDotText *string
	// BStarText retains the TLE field text.
	BStarText *string
	// ChecksumWarnings retains checksum details.
	ChecksumWarnings []CelestrakChecksumWarning
}

type celestrakPayloadWire struct {
	Kind                    string                     `json:"kind"`
	Line1                   string                     `json:"line1"`
	Line2                   string                     `json:"line2"`
	MeanMotionDoubleDotText *string                    `json:"mean_motion_double_dot_text"`
	BStarText               *string                    `json:"bstar_text"`
	ChecksumWarnings        []CelestrakChecksumWarning `json:"checksum_warnings"`
}

// CelestrakChecksumWarning retains a TLE checksum finding.
type CelestrakChecksumWarning struct {
	// LineLabel identifies the TLE line.
	LineLabel string `json:"line_label"`
	// Computed is the calculated checksum.
	Computed uint8 `json:"computed"`
	// Detail identifies the checksum finding.
	Detail CelestrakChecksumDetail `json:"detail"`
}

// CelestrakChecksumDetail describes a checksum finding. Values compare by value.
type CelestrakChecksumDetail struct {
	// Kind identifies the finding.
	Kind string `json:"kind"`
	// Expected is the expected checksum for a mismatch.
	Expected uint8 `json:"expected"`
	// Found is the non-digit character, when present.
	Found string `json:"found"`
}

// CelestrakGPRecord references an accepted record in a live response.
type CelestrakGPRecord struct {
	owner *CelestrakGPResponse
	index int
}

// CelestrakGPRecordData is a detached accepted record. Its fields compare by value with reflect.DeepEqual.
type CelestrakGPRecordData struct {
	// RecordIndex is the source record index.
	RecordIndex uint64 `json:"record_index"`
	// LineNumber is the source line number.
	LineNumber *uint64 `json:"line_number"`
	// ObjectName is the optional object name.
	ObjectName *string `json:"object_name"`
	// Payload describes the source payload.
	Payload CelestrakGPPayload `json:"payload"`
	// Raw holds the exact source bytes.
	Raw []byte `json:"-"`
}

// Record returns a reference to an accepted record.
func (r *CelestrakGPResponse) Record(index int) (CelestrakGPRecord, error) {
	value := CelestrakGPRecord{owner: r, index: index}
	_, err := value.Snapshot()
	if err != nil {
		return CelestrakGPRecord{}, err
	}
	return value, nil
}

// Snapshot copies all accepted record metadata and exact source bytes.
func (r CelestrakGPRecord) Snapshot() (CelestrakGPRecordData, error) {
	var value CelestrakGPRecordData
	if r.owner == nil || r.owner.handle == nil {
		return value, ErrClosed
	}
	data, err := r.owner.handle.RecordBytes(r.index, false, false)
	if err != nil {
		return value, publicCatalogError(err)
	}
	if err = json.Unmarshal(data, &value); err != nil {
		return value, err
	}
	if value.Payload.Kind == "tle" {
		var wire struct {
			Payload celestrakPayloadWire `json:"payload"`
		}
		if err = json.Unmarshal(data, &wire); err != nil {
			return value, err
		}
		tle, tleErr := r.TLEWithOpsModeAndPolicy(OpsModeAFSPC, TLEFilePolicyLenient)
		if tleErr != nil {
			return value, tleErr
		}
		elements, metadataErr := tle.Metadata()
		closeErr := tle.Close()
		if metadataErr != nil || closeErr != nil {
			return value, errors.Join(metadataErr, closeErr)
		}
		value.Payload.TLE = &CelestrakTLERecord{Elements: elements, Line1: wire.Payload.Line1, Line2: wire.Payload.Line2, MeanMotionDoubleDotText: wire.Payload.MeanMotionDoubleDotText, BStarText: wire.Payload.BStarText, ChecksumWarnings: wire.Payload.ChecksumWarnings}
	}
	value.Raw, err = r.owner.handle.RecordBytes(r.index, false, true)
	return value, publicCatalogError(err)
}

// ToOMM converts an accepted record to an OMM.
func (r CelestrakGPRecord) ToOMM() (*OMM, error) {
	if r.owner == nil || r.owner.handle == nil {
		return nil, ErrClosed
	}
	value, err := r.owner.handle.RecordOMM(r.index)
	if err != nil {
		return nil, publicCatalogError(err)
	}
	return publicOMM(value), nil
}

// Equal compares responses using core value equality.
func (r *CelestrakGPResponse) Equal(other *CelestrakGPResponse) (bool, error) {
	if r == nil || r.handle == nil || other == nil || other.handle == nil {
		return false, ErrClosed
	}
	equal, err := r.handle.Equal(other.handle)
	return equal, publicCatalogError(err)
}

// CelestrakGPRejectionReason describes a rejected source record.
type CelestrakGPRejectionReason struct {
	// Kind identifies the reason.
	Kind string `json:"kind"`
	// Cause preserves the typed native error, when present.
	Cause json.RawMessage `json:"cause"`
	// TLE contains the TLE parse error, when present.
	TLE *CelestrakTLEError `json:"-"`
	// OMM contains the OMM parse error, when present.
	OMM *OMMParseError `json:"-"`
	// ElementSet contains the SGP4 element conversion error, when present.
	ElementSet *CelestrakElementSetError `json:"-"`
}

// CelestrakTLEError retains a typed TLE refusal.
type CelestrakTLEError struct {
	// Kind identifies the refusal.
	Kind string `json:"kind"`
	// Fields holds the refusal details.
	Fields CelestrakTLEErrorFields `json:"fields"`
}

// CelestrakTLEErrorFields contains the TLE error detail fields.
type CelestrakTLEErrorFields struct {
	// Value is the offending text, when present.
	Value *string `json:"value"`
	// Reason explains an invalid value, when present.
	Reason *string `json:"reason"`
	// Field names the invalid field, when present.
	Field *string `json:"field"`
	// CatalogNumber is an out-of-range catalog number, when present.
	CatalogNumber *json.Number `json:"catalog_number"`
	// LineLabel identifies the checksum line, when present.
	LineLabel *string `json:"line_label"`
	// Expected is the stated checksum, when present.
	Expected *json.Number `json:"expected"`
	// Computed is the computed checksum, when present.
	Computed *json.Number `json:"computed"`
	// Found is the non-digit checksum character, when present.
	Found *string `json:"found"`
}

// CelestrakElementSetError retains a typed SGP4 element conversion refusal.
type CelestrakElementSetError struct {
	// Kind identifies the refusal.
	Kind string `json:"kind"`
	// Fields contains its detail fields.
	Fields CelestrakElementSetErrorFields `json:"fields"`
}

// CelestrakElementSetErrorFields contains SGP4 error details.
type CelestrakElementSetErrorFields struct {
	// Field identifies the field, when present.
	Field *string `json:"field"`
	// Kind identifies the input issue, when present.
	Kind *string `json:"kind"`
	// Message contains the refusal text, when present.
	Message *string `json:"message"`
	// Code is the SGP4 code, when present.
	Code *json.Number `json:"code"`
	// Budget is the resonance step budget, when present.
	Budget *json.Number `json:"budget"`
}

// UnmarshalJSON decodes the typed rejection cause.
func (reason *CelestrakGPRejectionReason) UnmarshalJSON(data []byte) error {
	type wire CelestrakGPRejectionReason
	var value wire
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*reason = CelestrakGPRejectionReason(value)
	switch reason.Kind {
	case "tle":
		reason.TLE = new(CelestrakTLEError)
		return json.Unmarshal(reason.Cause, reason.TLE)
	case "omm":
		reason.OMM = new(OMMParseError)
		return json.Unmarshal(reason.Cause, reason.OMM)
	case "invalid_element_set":
		reason.ElementSet = new(CelestrakElementSetError)
		return json.Unmarshal(reason.Cause, reason.ElementSet)
	}
	return nil
}

// CelestrakGPRejectedRecord references a rejected record in a live response.
type CelestrakGPRejectedRecord struct {
	owner *CelestrakGPResponse
	index int
}

// CelestrakGPRejectedRecordData is detached rejection data. Its fields compare by value with reflect.DeepEqual.
type CelestrakGPRejectedRecordData struct {
	// RecordIndex is the source record index.
	RecordIndex uint64 `json:"record_index"`
	// LineNumber is the source line number.
	LineNumber *uint64 `json:"line_number"`
	// Reason describes the rejection.
	Reason CelestrakGPRejectionReason `json:"reason"`
	// Raw holds the exact source bytes.
	Raw []byte `json:"-"`
}

// RejectedRecord returns a reference to a rejected record.
func (r *CelestrakGPResponse) RejectedRecord(index int) (CelestrakGPRejectedRecord, error) {
	value := CelestrakGPRejectedRecord{owner: r, index: index}
	_, err := value.Snapshot()
	if err != nil {
		return CelestrakGPRejectedRecord{}, err
	}
	return value, nil
}

// Snapshot copies rejection metadata and exact source bytes.
func (r CelestrakGPRejectedRecord) Snapshot() (CelestrakGPRejectedRecordData, error) {
	var value CelestrakGPRejectedRecordData
	if r.owner == nil || r.owner.handle == nil {
		return value, ErrClosed
	}
	data, err := r.owner.handle.RecordBytes(r.index, true, false)
	if err != nil {
		return value, publicCatalogError(err)
	}
	if err = json.Unmarshal(data, &value); err != nil {
		return value, err
	}
	value.Raw, err = r.owner.handle.RecordBytes(r.index, true, true)
	return value, publicCatalogError(err)
}

// AsStr returns the GP wire value for the format.
func (format CelestrakGPFormat) AsStr() (string, error) {
	value, err := native.CelestrakText(8, 0, "", 0, uint32(format))
	return value, publicCatalogError(err)
}

// AsStr returns the GP wire value for the special selector.
func (special CelestrakGPSpecial) AsStr() (string, error) {
	value, err := native.CelestrakText(9, uint32(special), "", 0, 0)
	return value, publicCatalogError(err)
}

// TLE copies an accepted TLE using AFSPC mode and strict checksum policy.
func (r CelestrakGPRecord) TLE() (*TLE, error) {
	return r.TLEWithOpsModeAndPolicy(OpsModeAFSPC, TLEFilePolicyStrict)
}

// TLEWithOpsMode copies an accepted TLE using the supplied operations mode.
func (r CelestrakGPRecord) TLEWithOpsMode(mode OpsMode) (*TLE, error) {
	return r.TLEWithOpsModeAndPolicy(mode, TLEFilePolicyStrict)
}

// TLEWithPolicy copies an accepted TLE using the supplied checksum policy.
func (r CelestrakGPRecord) TLEWithPolicy(policy TLEFilePolicy) (*TLE, error) {
	return r.TLEWithOpsModeAndPolicy(OpsModeAFSPC, policy)
}

// TLEWithOpsModeAndPolicy copies an accepted TLE with the supplied mode and policy.
func (r CelestrakGPRecord) TLEWithOpsModeAndPolicy(mode OpsMode, policy TLEFilePolicy) (*TLE, error) {
	if r.owner == nil || r.owner.handle == nil {
		return nil, ErrClosed
	}
	value, err := r.owner.handle.RecordTLE(r.index, uint32(mode), uint32(policy))
	if err != nil {
		return nil, publicCatalogError(err)
	}
	return &TLE{handle: value}, nil
}

// AllowedHosts returns the catalog's allowed hosts.
func AllowedHosts() ([]string, error) {
	value, err := native.CelestrakText(10, 0, "", 0, 0)
	if err != nil {
		return nil, publicCatalogError(err)
	}
	var hosts []string
	err = json.Unmarshal([]byte(value), &hosts)
	return hosts, err
}

// ToElementSet converts an accepted record into SGP4 elements.
func (r CelestrakGPRecord) ToElementSet() (OMMElementSet, error) {
	if r.owner == nil || r.owner.handle == nil {
		return OMMElementSet{}, ErrClosed
	}
	value, err := r.owner.handle.RecordElementSet(r.index)
	if err != nil {
		return OMMElementSet{}, publicCatalogError(err)
	}
	return publicOMMElementSet(value), nil
}

// DataCatalogError retains a catalog refusal and its core-provided details.
// Equal compares detached fields because the C API has no error equality call.
type DataCatalogError struct {
	// Kind identifies the refusal.
	Kind string
	// Selector identifies an invalid selector, when present.
	Selector string `json:"selector"`
	// Value is the invalid selector value.
	Value string `json:"value"`
	// Query is the repeated-download query, when present.
	Query *string `json:"query"`
	// LastSuccessfulDownload is the reported download time.
	LastSuccessfulDownload *string `json:"last_successful_download"`
	// Reason describes an invalid response.
	Reason string `json:"reason"`
	// Line is the optional source line.
	Line *uint64 `json:"line"`
	// Column is the optional source column.
	Column *uint64 `json:"column"`
	cause  error
}

// Error returns the native refusal text.
func (e *DataCatalogError) Error() string {
	if e.cause != nil {
		return e.cause.Error()
	}
	return "sidereon: catalog error " + e.Kind
}

// Unwrap exposes the native error.
func (e *DataCatalogError) Unwrap() error { return e.cause }

// Equal compares the detached catalog error fields.
func (e *DataCatalogError) Equal(other *DataCatalogError) bool {
	if e == nil || other == nil {
		return e == other
	}
	left, right := *e, *other
	left.cause, right.cause = nil, nil
	return reflect.DeepEqual(left, right)
}
func publicCatalogError(err error) error {
	result := publicError(err)
	if result == nil {
		return nil
	}
	var engine *EngineError
	if errors.As(result, &engine) && engine.Family == EngineErrorFamilyCatalog {
		if engine.Kind == "tle" || engine.Kind == "omm" {
			var fields struct {
				Cause json.RawMessage `json:"cause"`
			}
			if json.Unmarshal(engine.Fields, &fields) == nil && len(fields.Cause) != 0 {
				value := &CelestrakGPRecordError{Kind: engine.Kind, cause: result}
				if engine.Kind == "tle" {
					value.TLE = new(CelestrakTLEError)
					if json.Unmarshal(fields.Cause, value.TLE) != nil {
						return result
					}
				} else {
					value.OMM = new(OMMParseError)
					if json.Unmarshal(fields.Cause, value.OMM) != nil {
						return result
					}
				}
				return value
			}
		}
		value := &DataCatalogError{Kind: engine.Kind, cause: result}
		if json.Unmarshal(engine.Fields, value) == nil {
			return value
		}
	}
	return result
}

// CelestrakGPRecordError retains a record conversion refusal.
// Equal compares detached fields because the C API has no error equality call.
type CelestrakGPRecordError struct {
	// Kind identifies the native cause.
	Kind string
	// TLE is the TLE cause, when present.
	TLE *CelestrakTLEError
	// OMM is the OMM cause, when present.
	OMM   *OMMParseError
	cause error
}

// Error returns the native refusal text.
func (e *CelestrakGPRecordError) Error() string {
	if e.cause != nil {
		return e.cause.Error()
	}
	return "sidereon: GP record error " + e.Kind
}

// Unwrap exposes the native refusal.
func (e *CelestrakGPRecordError) Unwrap() error { return e.cause }

// Equal compares the detached conversion error fields.
func (e *CelestrakGPRecordError) Equal(other *CelestrakGPRecordError) bool {
	if e == nil || other == nil {
		return e == other
	}
	left, right := *e, *other
	left.cause, right.cause = nil, nil
	return reflect.DeepEqual(left, right)
}

// Equal compares accepted records using core equality.
func (r CelestrakGPRecord) Equal(other CelestrakGPRecord) (bool, error) {
	return compareGpRecords(r.owner, other.owner, r.index, other.index, 0)
}

// PayloadEqual compares accepted native payloads using core equality.
func (r CelestrakGPRecord) PayloadEqual(other CelestrakGPRecord) (bool, error) {
	return compareGpRecords(r.owner, other.owner, r.index, other.index, 2)
}

// TLEEqual compares TLE payloads using core equality.
func (r CelestrakGPRecord) TLEEqual(other CelestrakGPRecord) (bool, error) {
	return compareGpRecords(r.owner, other.owner, r.index, other.index, 3)
}

// Equal compares rejected records using core equality.
func (r CelestrakGPRejectedRecord) Equal(other CelestrakGPRejectedRecord) (bool, error) {
	return compareGpRecords(r.owner, other.owner, r.index, other.index, 1)
}

// ReasonEqual compares rejection reasons using core equality.
func (r CelestrakGPRejectedRecord) ReasonEqual(other CelestrakGPRejectedRecord) (bool, error) {
	return compareGpRecords(r.owner, other.owner, r.index, other.index, 4)
}

// StructuralErrorEqual compares structural rejection causes using core equality.
func (r CelestrakGPRejectedRecord) StructuralErrorEqual(other CelestrakGPRejectedRecord) (bool, error) {
	return compareGpRecords(r.owner, other.owner, r.index, other.index, 5)
}
func compareGpRecords(a, b *CelestrakGPResponse, i, j int, mode uint32) (bool, error) {
	if a == nil || b == nil || a.handle == nil || b.handle == nil {
		return false, ErrClosed
	}
	equal, err := a.handle.CompareRecord(b.handle, i, j, mode)
	return equal, publicCatalogError(err)
}
