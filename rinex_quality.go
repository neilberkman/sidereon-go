package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// RINEXLintSeverity classifies a native RINEX lint finding.
type RINEXLintSeverity uint32

const (
	// RINEXLintFatal marks an unrecoverable input error.
	RINEXLintFatal RINEXLintSeverity = iota
	// RINEXLintError marks an error that prevents a clean product.
	RINEXLintError
	// RINEXLintWarning marks a recoverable quality issue.
	RINEXLintWarning
	// RINEXLintInfo marks informational output.
	RINEXLintInfo
)

// RINEXLintSummary contains detached finding counts and CRINEX provenance.
type RINEXLintSummary struct {
	// FindingCount is the total number of lint findings.
	FindingCount int
	// FatalCount is the number of fatal findings.
	FatalCount int
	// ErrorCount is the number of error-severity findings.
	ErrorCount int
	// WarningCount is the number of warning-severity findings.
	WarningCount int
	// InfoCount is the number of informational findings.
	InfoCount int
	// IsClean reports whether the native linter found no fatal or error finding.
	IsClean bool
	// DecodedFromCRINEX reports whether the input was decoded from CRINEX compression.
	DecodedFromCRINEX bool
}

// RINEXLintFinding is one detached lint finding. Optional fields are guarded by
// their Has* values; EpochIndex is an epoch ordinal, not a time in seconds.
type RINEXLintFinding struct {
	// Code is the observable code.
	Code string
	// Severity classifies the finding as fatal, error, warning, or informational.
	Severity RINEXLintSeverity
	// Repairable reports whether the finding can be repaired.
	Repairable bool
	// HasEpochIndex reports whether EpochIndex identifies the affected epoch.
	HasEpochIndex bool
	// EpochIndex is the zero-based RINEX epoch ordinal when HasEpochIndex is true.
	EpochIndex int
	// HasSatellite reports whether Satellite identifies the affected GNSS satellite.
	HasSatellite bool
	// Satellite is the GNSS satellite identifier when HasSatellite is true.
	Satellite string
	// HasField reports whether Field identifies the affected observation field.
	HasField bool
	// Field names the RINEX header or observation field when HasField is true.
	Field string
}

// RINEXLintReport owns a C-backed lint result.
type RINEXLintReport struct {
	_      noCopy
	handle *native.RinexLintReport
}

// LintRINEXObservation lints RINEX observation bytes through the C ABI.
func LintRINEXObservation(data []byte) (*RINEXLintReport, error) {
	h, e := native.ParseRINEXLint(data, true)
	if e != nil {
		return nil, publicError(e)
	}
	return &RINEXLintReport{handle: h}, nil
}

// LintRINEXNav lints RINEX navigation bytes through the C ABI.
func LintRINEXNav(data []byte) (*RINEXLintReport, error) {
	h, e := native.ParseRINEXLint(data, false)
	if e != nil {
		return nil, publicError(e)
	}
	return &RINEXLintReport{handle: h}, nil
}

// Close releases the lint report; repeated calls are safe.
func (r *RINEXLintReport) Close() error {
	if r == nil || r.handle == nil {
		return nil
	}
	return publicError(r.handle.Close())
}

// Summary returns a detached lint summary.
func (r *RINEXLintReport) Summary() (RINEXLintSummary, error) {
	if r == nil || r.handle == nil {
		return RINEXLintSummary{}, ErrClosed
	}
	v, e := r.handle.Summary()
	return RINEXLintSummary{FindingCount: v.FindingCount, FatalCount: v.FatalCount, ErrorCount: v.ErrorCount, WarningCount: v.WarningCount, InfoCount: v.InfoCount, IsClean: v.IsClean, DecodedFromCRINEX: v.DecodedFromCRINEX}, publicError(e)
}

// Findings returns detached lint findings.
func (r *RINEXLintReport) Findings() ([]RINEXLintFinding, error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	v, e := r.handle.Findings()
	if e != nil {
		return nil, publicError(e)
	}
	out := make([]RINEXLintFinding, len(v))
	for i, x := range v {
		out[i] = RINEXLintFinding{Code: x.Code, Severity: RINEXLintSeverity(x.Severity), Repairable: x.Repairable, HasEpochIndex: x.HasEpochIndex, EpochIndex: x.EpochIndex, HasSatellite: x.HasSatellite, Satellite: x.Satellite, HasField: x.HasField, Field: x.Field}
	}
	return out, nil
}

// RINEXRepairOptions controls native RINEX repair transforms. File-stamp
// strings are optional and remain owned by the caller.
type RINEXRepairOptions struct {
	// HasFileStamp reports whether FileStamp identifies the affected file location.
	HasFileStamp bool
	// Program is the program name written to the RINEX header when HasFileStamp is true.
	Program string
	// RunBy is the operator or agency written to the RINEX header when HasFileStamp is true.
	RunBy string
	// Date is the file-stamp date written to the RINEX header when HasFileStamp is true.
	Date string
	// SetInterval enables replacement of the observation interval header.
	SetInterval bool
	// SetTimeOfLastObs enables replacement of the time-of-last-observation header.
	SetTimeOfLastObs bool
	// SetObservationCounts enables recomputation of RINEX observation counts.
	SetObservationCounts bool
	// DropEmptyRecords enables removal of records with no usable observations.
	DropEmptyRecords bool
	// SortRecords enables chronological sorting of repaired records.
	SortRecords bool
	// DropUnsupported enables removal of unsupported observation records.
	DropUnsupported bool
}

// NewRINEXRepairOptions returns the C ABI defaults.
func NewRINEXRepairOptions() (RINEXRepairOptions, error) {
	v, e := native.NewRINEXRepairOptions()
	return RINEXRepairOptions{HasFileStamp: v.HasFileStamp, Program: v.Program, RunBy: v.RunBy, Date: v.Date, SetInterval: v.SetInterval, SetTimeOfLastObs: v.SetTimeOfLastObs, SetObservationCounts: v.SetObservationCounts, DropEmptyRecords: v.DropEmptyRecords, SortRecords: v.SortRecords, DropUnsupported: v.DropUnsupported}, publicError(e)
}

// RINEXRepairAction describes one detached repair operation.
type RINEXRepairAction struct {
	// ID identifies the repair action in the native action log.
	ID string
	// Message is the diagnostic text describing the repair action.
	Message string
}

// RINEXRepair owns a C-backed repaired RINEX product.
type RINEXRepair struct {
	_      noCopy
	handle *native.RinexRepair
}

// RepairRINEXObservation repairs RINEX observation bytes through the C ABI.
func RepairRINEXObservation(data []byte, options *RINEXRepairOptions) (*RINEXRepair, error) {
	var o *native.NativeRINEXRepairOptions
	if options != nil {
		o = &native.NativeRINEXRepairOptions{HasFileStamp: options.HasFileStamp, Program: options.Program, RunBy: options.RunBy, Date: options.Date, SetInterval: options.SetInterval, SetTimeOfLastObs: options.SetTimeOfLastObs, SetObservationCounts: options.SetObservationCounts, DropEmptyRecords: options.DropEmptyRecords, SortRecords: options.SortRecords, DropUnsupported: options.DropUnsupported}
	}
	h, e := native.ParseRINEXRepair(data, true, o)
	if e != nil {
		return nil, publicError(e)
	}
	return &RINEXRepair{handle: h}, nil
}

// RepairRINEXNav repairs RINEX navigation bytes through the C ABI.
func RepairRINEXNav(data []byte, options *RINEXRepairOptions) (*RINEXRepair, error) {
	var o *native.NativeRINEXRepairOptions
	if options != nil {
		o = &native.NativeRINEXRepairOptions{HasFileStamp: options.HasFileStamp, Program: options.Program, RunBy: options.RunBy, Date: options.Date, SetInterval: options.SetInterval, SetTimeOfLastObs: options.SetTimeOfLastObs, SetObservationCounts: options.SetObservationCounts, DropEmptyRecords: options.DropEmptyRecords, SortRecords: options.SortRecords, DropUnsupported: options.DropUnsupported}
	}
	h, e := native.ParseRINEXRepair(data, false, o)
	if e != nil {
		return nil, publicError(e)
	}
	return &RINEXRepair{handle: h}, nil
}

// Close releases the repair result; repeated calls are safe.
func (r *RINEXRepair) Close() error {
	if r == nil || r.handle == nil {
		return nil
	}
	return publicError(r.handle.Close())
}

// Summary returns the repair lint summary.
func (r *RINEXRepair) Summary() (RINEXLintSummary, error) {
	if r == nil || r.handle == nil {
		return RINEXLintSummary{}, ErrClosed
	}
	v, e := r.handle.Summary()
	return RINEXLintSummary{FindingCount: v.FindingCount, FatalCount: v.FatalCount, ErrorCount: v.ErrorCount, WarningCount: v.WarningCount, InfoCount: v.InfoCount, IsClean: v.IsClean, DecodedFromCRINEX: v.DecodedFromCRINEX}, publicError(e)
}

// Actions returns detached repair actions.
func (r *RINEXRepair) Actions() ([]RINEXRepairAction, error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	v, e := r.handle.Actions()
	if e != nil {
		return nil, publicError(e)
	}
	out := make([]RINEXRepairAction, len(v))
	for i, x := range v {
		out[i] = RINEXRepairAction{ID: x.ID, Message: x.Message}
	}
	return out, nil
}

// RINEXText returns detached repaired RINEX representation.
func (r *RINEXRepair) RINEXText() ([]byte, error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	v, e := r.handle.Text()
	return v, publicError(e)
}

// CRINEXText returns detached repaired CRINEX representation.
func (r *RINEXRepair) CRINEXText() ([]byte, error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	v, e := r.handle.CRINEXText()
	return v, publicError(e)
}

// DecodeCRINEX expands CRINEX bytes into a newly allocated RINEX byte slice.
func DecodeCRINEX(data []byte) ([]byte, error) {
	v, e := native.DecodeCRINEX(data)
	return v, publicError(e)
}

// EncodeCRINEX compresses RINEX bytes into a newly allocated CRINEX slice.
func EncodeCRINEX(data []byte) ([]byte, error) {
	v, e := native.EncodeCRINEX(data)
	return v, publicError(e)
}
