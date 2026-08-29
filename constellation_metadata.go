package sidereon

import (
	"fmt"
	"time"

	"github.com/neilberkman/sidereon-go/internal/native"
)

func metadataUnixMicroseconds(value time.Time) (int64, error) {
	utc := value.UTC()
	seconds := utc.Unix()
	fractionalMicroseconds := int64(utc.Nanosecond() / 1_000)
	const (
		maxMicroseconds       = int64(1<<63 - 1)
		minMicroseconds       = -maxMicroseconds - 1
		microsecondsPerSecond = int64(1_000_000)
	)
	maxSeconds := maxMicroseconds / microsecondsPerSecond
	minSeconds := minMicroseconds / microsecondsPerSecond
	if seconds > maxSeconds || seconds < minSeconds-1 {
		return 0, fmt.Errorf("sidereon: time does not fit in Unix microseconds")
	}
	if seconds == maxSeconds && fractionalMicroseconds > maxMicroseconds%microsecondsPerSecond {
		return 0, fmt.Errorf("sidereon: time does not fit in Unix microseconds")
	}
	minimumBoundaryFraction := microsecondsPerSecond + minMicroseconds%microsecondsPerSecond
	if seconds == minSeconds-1 {
		if fractionalMicroseconds < minimumBoundaryFraction {
			return 0, fmt.Errorf("sidereon: time does not fit in Unix microseconds")
		}
		return minMicroseconds + fractionalMicroseconds - minimumBoundaryFraction, nil
	}
	wholeMicroseconds := seconds * microsecondsPerSecond
	if wholeMicroseconds > maxMicroseconds-fractionalMicroseconds {
		return 0, fmt.Errorf("sidereon: time does not fit in Unix microseconds")
	}
	return wholeMicroseconds + fractionalMicroseconds, nil
}

// ConstellationBoolStyle selects boolean rendering in constellation CSV.
type ConstellationBoolStyle uint32

const (
	ConstellationBoolStyleLower ConstellationBoolStyle = ConstellationBoolStyle(native.ConstellationBoolStyleLowerValue)
	ConstellationBoolStyleTitle ConstellationBoolStyle = ConstellationBoolStyle(native.ConstellationBoolStyleTitleValue)
)

// ConstellationPRN identifies a system and PRN pair.
type ConstellationPRN struct {
	System GNSSSystem
	PRN    uint16
}

// ConstellationDiffCounts reports the number of each diff category.
type ConstellationDiffCounts struct{ Added, Removed, NORADReassigned, SP3IDChanged, SVNChanged, FDMAChannelChanged, ActivityChanged, UsabilityChanged int }

// ConstellationBoolChange records a boolean field change.
type ConstellationBoolChange struct {
	System   GNSSSystem
	PRN      uint16
	From, To bool
}

// ConstellationOptionalU16Change records an optional uint16 field change.
type ConstellationOptionalU16Change struct {
	System      GNSSSystem
	PRN         uint16
	FromPresent bool
	From        uint16
	ToPresent   bool
	To          uint16
}

// ConstellationOptionalI8Change records an optional int8 field change.
type ConstellationOptionalI8Change struct {
	System      GNSSSystem
	PRN         uint16
	FromPresent bool
	From        int8
	ToPresent   bool
	To          int8
}

// ConstellationU32Change records a uint32 field change.
type ConstellationU32Change struct {
	System   GNSSSystem
	PRN      uint16
	From, To uint32
}

// ConstellationStringChangeMeta records changed string lengths.
type ConstellationStringChangeMeta struct {
	System         GNSSSystem
	PRN            uint16
	FromLen, ToLen int
}

// ConstellationStringChange records a changed constellation string.
type ConstellationStringChange struct {
	Meta     ConstellationStringChangeMeta
	From, To string
}

// Constellation owns a C-derived satellite catalog.
type Constellation struct {
	_      noCopy
	handle *native.Constellation
}

// ConstellationDiff owns a C-derived catalog comparison.
type ConstellationDiff struct {
	_      noCopy
	handle *native.ConstellationDiff
}

// ConstellationValidation owns a C-derived catalog validation report.
type ConstellationValidation struct {
	_      noCopy
	handle *native.ConstellationValidation
}

func publicConstellationRecord(v native.ConstellationRecord) ConstellationRecord {
	return ConstellationRecord{System: GNSSSystem(v.System), PRN: v.PRN, SVNPresent: v.SVNPresent, SVN: v.SVN, NORADID: v.NORADID, FDMAChannelPresent: v.FDMAChannelPresent, FDMAChannel: v.FDMAChannel, Active: v.Active, Usable: v.Usable}
}

// BuildConstellation builds a catalog from OMM JSON and NAVCEN HTML.
func BuildConstellation(system GNSSSystem, ommJSON, navcenHTML []byte) (*Constellation, error) {
	v, e := native.BuildConstellation(uint32(system), ommJSON, navcenHTML)
	if e != nil {
		return nil, publicError(e)
	}
	if v == nil {
		return nil, errNilNativeHandle
	}
	return &Constellation{handle: v}, nil
}

// BuildConstellationAt builds a catalog evaluated at a specified instant.
func BuildConstellationAt(system GNSSSystem, ommJSON, navcenHTML []byte, evaluatedAt time.Time) (*Constellation, error) {
	us, e := metadataUnixMicroseconds(evaluatedAt)
	if e != nil {
		return nil, e
	}
	v, e := native.BuildConstellationAt(uint32(system), ommJSON, navcenHTML, us)
	if e != nil {
		return nil, publicError(e)
	}
	if v == nil {
		return nil, errNilNativeHandle
	}
	return &Constellation{handle: v}, nil
}

// Close releases the catalog; it is safe to call more than once.
func (c *Constellation) Close() error {
	if c == nil || c.handle == nil {
		return nil
	}
	return publicError(c.handle.Close())
}

// RecordCount returns the number of catalog records.
func (c *Constellation) RecordCount() (int, error) {
	if c == nil || c.handle == nil {
		return 0, ErrClosed
	}
	v, e := c.handle.RecordCount()
	return v, publicError(e)
}

// Record returns one catalog record by index.
func (c *Constellation) Record(index int) (ConstellationRecord, error) {
	if c == nil || c.handle == nil {
		return ConstellationRecord{}, ErrClosed
	}
	v, e := c.handle.Record(index)
	return publicConstellationRecord(v), publicError(e)
}

// Records returns detached catalog records.
func (c *Constellation) Records() ([]ConstellationRecord, error) {
	if c == nil || c.handle == nil {
		return nil, ErrClosed
	}
	v, e := c.handle.Records()
	if e != nil {
		return nil, publicError(e)
	}
	out := make([]ConstellationRecord, len(v))
	for i := range v {
		out[i] = publicConstellationRecord(v[i])
	}
	return out, nil
}

// GNSSSP3ID returns the canonical SP3 token for a constellation member.
func (c *Constellation) GNSSSP3ID(system GNSSSystem, prn uint16) (string, error) {
	if c == nil || c.handle == nil {
		return "", ErrClosed
	}
	v, e := c.handle.GNSSSP3ID(uint32(system), prn)
	return v, publicError(e)
}

// GNSSSP3ID returns the canonical SP3 token for a system and PRN.
func GNSSSP3ID(system GNSSSystem, prn uint16) (string, error) {
	v, e := native.GNSSSP3ID(uint32(system), prn)
	return v, publicError(e)
}

// ToCSV serializes the catalog using the requested boolean style.
func (c *Constellation) ToCSV(boolStyle ConstellationBoolStyle) ([]byte, error) {
	if c == nil || c.handle == nil {
		return nil, ErrClosed
	}
	v, e := c.handle.ToCSV(uint32(boolStyle))
	return v, publicError(e)
}

// Validate returns a native catalog validation report.
func (c *Constellation) Validate() (*ConstellationValidation, error) {
	if c == nil || c.handle == nil {
		return nil, ErrClosed
	}
	v, e := c.handle.Validate()
	if e != nil {
		return nil, publicError(e)
	}
	return &ConstellationValidation{handle: v}, nil
}

// ValidateAgainstSP3 validates the catalog against an SP3 product.
func (c *Constellation) ValidateAgainstSP3(sp3 *SP3) (*ConstellationValidation, error) {
	if c == nil || c.handle == nil || sp3 == nil || sp3.handle == nil {
		return nil, ErrClosed
	}
	v, e := c.handle.ValidateAgainstSP3(sp3.handle)
	if e != nil {
		return nil, publicError(e)
	}
	return &ConstellationValidation{handle: v}, nil
}

// ValidateAgainstSP3IDs validates the catalog against SP3 tokens.
func (c *Constellation) ValidateAgainstSP3IDs(ids []string) (*ConstellationValidation, error) {
	if c == nil || c.handle == nil {
		return nil, ErrClosed
	}
	v, e := c.handle.ValidateAgainstSP3IDs(ids)
	if e != nil {
		return nil, publicError(e)
	}
	return &ConstellationValidation{handle: v}, nil
}

// ValidateAgainstSP3IDsStrict returns an error when any token mismatches.
func (c *Constellation) ValidateAgainstSP3IDsStrict(ids []string) error {
	if c == nil || c.handle == nil {
		return ErrClosed
	}
	return publicError(c.handle.ValidateAgainstSP3IDsStrict(ids))
}

// Close releases the validation report; it is safe to call more than once.
func (v *ConstellationValidation) Close() error {
	if v == nil || v.handle == nil {
		return nil
	}
	return publicError(v.handle.Close())
}

// IsValid reports whether the native validation passed.
func (v *ConstellationValidation) IsValid() (bool, error) {
	if v == nil || v.handle == nil {
		return false, ErrClosed
	}
	x, e := v.handle.IsValid()
	return x, publicError(e)
}

// InactiveUnusablePRNs returns inactive and unusable PRNs.
func (v *ConstellationValidation) InactiveUnusablePRNs() ([]ConstellationPRN, error) {
	if v == nil || v.handle == nil {
		return nil, ErrClosed
	}
	x, e := v.handle.InactiveUnusablePRNs()
	return mapConstellationPRNs(x), publicError(e)
}

// DuplicatePRNs returns duplicate system/PRN pairs.
func (v *ConstellationValidation) DuplicatePRNs() ([]ConstellationPRN, error) {
	if v == nil || v.handle == nil {
		return nil, ErrClosed
	}
	x, e := v.handle.DuplicatePRNs()
	return mapConstellationPRNs(x), publicError(e)
}

// DuplicateNORADIDs returns duplicate NORAD identifiers.
func (v *ConstellationValidation) DuplicateNORADIDs() ([]uint32, error) {
	if v == nil || v.handle == nil {
		return nil, ErrClosed
	}
	x, e := v.handle.DuplicateNORADIDs()
	return x, publicError(e)
}

// MissingSP3IDs returns catalog tokens absent from the SP3 input.
func (v *ConstellationValidation) MissingSP3IDs() ([]string, error) {
	if v == nil || v.handle == nil {
		return nil, ErrClosed
	}
	x, e := v.handle.MissingSP3IDs()
	return x, publicError(e)
}

// ExtraSP3IDs returns SP3 tokens absent from the catalog.
func (v *ConstellationValidation) ExtraSP3IDs() ([]string, error) {
	if v == nil || v.handle == nil {
		return nil, ErrClosed
	}
	x, e := v.handle.ExtraSP3IDs()
	return x, publicError(e)
}
func mapConstellationPRNs(x []native.ConstellationPRN) []ConstellationPRN {
	out := make([]ConstellationPRN, len(x))
	for i, v := range x {
		out[i] = ConstellationPRN{System: GNSSSystem(v.System), PRN: v.PRN}
	}
	return out
}

// Diff compares this catalog with another catalog.
func (c *Constellation) Diff(other *Constellation) (*ConstellationDiff, error) {
	if c == nil || c.handle == nil || other == nil || other.handle == nil {
		return nil, ErrClosed
	}
	v, e := c.handle.Diff(other.handle)
	if e != nil {
		return nil, publicError(e)
	}
	return &ConstellationDiff{handle: v}, nil
}

// Close releases the diff; it is safe to call more than once.
func (d *ConstellationDiff) Close() error {
	if d == nil || d.handle == nil {
		return nil
	}
	return publicError(d.handle.Close())
}

// Changed reports whether any catalog field differs.
func (d *ConstellationDiff) Changed() (bool, error) {
	if d == nil || d.handle == nil {
		return false, ErrClosed
	}
	v, e := d.handle.Changed()
	return v, publicError(e)
}

// Counts returns per-category diff counts.
func (d *ConstellationDiff) Counts() (ConstellationDiffCounts, error) {
	if d == nil || d.handle == nil {
		return ConstellationDiffCounts{}, ErrClosed
	}
	v, e := d.handle.Counts()
	return ConstellationDiffCounts{Added: v.Added, Removed: v.Removed, NORADReassigned: v.NORADReassigned, SP3IDChanged: v.SP3IDChanged, SVNChanged: v.SVNChanged, FDMAChannelChanged: v.FDMAChannelChanged, ActivityChanged: v.ActivityChanged, UsabilityChanged: v.UsabilityChanged}, publicError(e)
}

// Added returns records present only in the new catalog.
func (d *ConstellationDiff) Added() ([]ConstellationRecord, error) {
	if d == nil || d.handle == nil {
		return nil, ErrClosed
	}
	x, e := d.handle.Added()
	out := make([]ConstellationRecord, len(x))
	for i := range x {
		out[i] = publicConstellationRecord(x[i])
	}
	return out, publicError(e)
}

// Removed returns records present only in the old catalog.
func (d *ConstellationDiff) Removed() ([]ConstellationRecord, error) {
	if d == nil || d.handle == nil {
		return nil, ErrClosed
	}
	x, e := d.handle.Removed()
	out := make([]ConstellationRecord, len(x))
	for i := range x {
		out[i] = publicConstellationRecord(x[i])
	}
	return out, publicError(e)
}

// ActivityChanged returns activity changes.
func (d *ConstellationDiff) ActivityChanged() ([]ConstellationBoolChange, error) {
	if d == nil || d.handle == nil {
		return nil, ErrClosed
	}
	x, e := d.handle.ActivityChanged()
	return mapBoolChanges(x), publicError(e)
}

// UsabilityChanged returns usability changes.
func (d *ConstellationDiff) UsabilityChanged() ([]ConstellationBoolChange, error) {
	if d == nil || d.handle == nil {
		return nil, ErrClosed
	}
	x, e := d.handle.UsabilityChanged()
	return mapBoolChanges(x), publicError(e)
}

// FDMAChannelChanged returns GLONASS FDMA changes.
func (d *ConstellationDiff) FDMAChannelChanged() ([]ConstellationOptionalI8Change, error) {
	if d == nil || d.handle == nil {
		return nil, ErrClosed
	}
	x, e := d.handle.FDMAChannelChanged()
	out := make([]ConstellationOptionalI8Change, len(x))
	for i := range x {
		out[i] = ConstellationOptionalI8Change{System: GNSSSystem(x[i].System), PRN: x[i].PRN, FromPresent: x[i].FromPresent, From: x[i].From, ToPresent: x[i].ToPresent, To: x[i].To}
	}
	return out, publicError(e)
}

// SVNChanged returns SVN changes.
func (d *ConstellationDiff) SVNChanged() ([]ConstellationOptionalU16Change, error) {
	if d == nil || d.handle == nil {
		return nil, ErrClosed
	}
	x, e := d.handle.SVNChanged()
	out := make([]ConstellationOptionalU16Change, len(x))
	for i := range x {
		out[i] = ConstellationOptionalU16Change{System: GNSSSystem(x[i].System), PRN: x[i].PRN, FromPresent: x[i].FromPresent, From: x[i].From, ToPresent: x[i].ToPresent, To: x[i].To}
	}
	return out, publicError(e)
}

// NORADReassigned returns NORAD reassignment changes.
func (d *ConstellationDiff) NORADReassigned() ([]ConstellationU32Change, error) {
	if d == nil || d.handle == nil {
		return nil, ErrClosed
	}
	x, e := d.handle.NORADReassigned()
	out := make([]ConstellationU32Change, len(x))
	for i := range x {
		out[i] = ConstellationU32Change{System: GNSSSystem(x[i].System), PRN: x[i].PRN, From: x[i].From, To: x[i].To}
	}
	return out, publicError(e)
}

// SP3IDChanged returns SP3 token changes.
func (d *ConstellationDiff) SP3IDChanged() ([]ConstellationStringChange, error) {
	if d == nil || d.handle == nil {
		return nil, ErrClosed
	}
	x, e := d.handle.SP3IDChanged()
	out := make([]ConstellationStringChange, len(x))
	for i := range x {
		out[i] = ConstellationStringChange{Meta: ConstellationStringChangeMeta{System: GNSSSystem(x[i].Meta.System), PRN: x[i].Meta.PRN, FromLen: x[i].Meta.FromLen, ToLen: x[i].Meta.ToLen}, From: x[i].From, To: x[i].To}
	}
	return out, publicError(e)
}
func mapBoolChanges(x []native.ConstellationBoolChange) []ConstellationBoolChange {
	out := make([]ConstellationBoolChange, len(x))
	for i := range x {
		out[i] = ConstellationBoolChange{System: GNSSSystem(x[i].System), PRN: x[i].PRN, From: x[i].From, To: x[i].To}
	}
	return out
}

// GalileoPRNForGSAT maps a Galileo GSAT number to PRN when defined.
func GalileoPRNForGSAT(gsat uint16) (bool, uint16, error) {
	v, p, e := native.GalileoPRNForGSAT(gsat)
	return v, p, publicError(e)
}

// GLONASSSlotForNumber maps a GLONASS satellite number to its slot.
func GLONASSSlotForNumber(number uint16) (bool, uint16, error) {
	v, p, e := native.GLONASSSlotForNumber(number)
	return v, p, publicError(e)
}

// GLONASSFDMAChannel maps a slot to its FDMA channel when defined.
func GLONASSFDMAChannel(slot uint16) (bool, int8, error) {
	v, p, e := native.GLONASSFDMAChannel(slot)
	return v, p, publicError(e)
}

// NAVCENAssessment is one time-aware navigation advisory assessment.
type NAVCENAssessment struct {
	System                GNSSSystem
	PRN                   uint16
	SVNPresent            bool
	SVN                   uint16
	Usable                bool
	ActiveNANU            bool
	EvaluatedAt           time.Time
	EvaluatedAtUnixUS     int64
	Timing                uint32
	EffectiveStartPresent bool
	EffectiveStart        time.Time
	EffectiveStartUnixUS  int64
	EffectiveEndPresent   bool
	EffectiveEnd          time.Time
	EffectiveEndUnixUS    int64
}

// NAVCENAssessments owns parsed NAVCEN assessments.
type NAVCENAssessments struct {
	_      noCopy
	handle *native.NavcenAssessments
}

// ParseNAVCENAt parses NAVCEN HTML with an explicit evaluation instant.
func ParseNAVCENAt(html []byte, evaluatedAt time.Time) (*NAVCENAssessments, error) {
	us, e := metadataUnixMicroseconds(evaluatedAt)
	if e != nil {
		return nil, e
	}
	v, e := native.ParseNAVCENAt(html, us)
	if e != nil {
		return nil, publicError(e)
	}
	return &NAVCENAssessments{handle: v}, nil
}

// Close releases NAVCEN assessments; it is safe to call more than once.
func (n *NAVCENAssessments) Close() error {
	if n == nil || n.handle == nil {
		return nil
	}
	return publicError(n.handle.Close())
}

// Count returns the number of parsed assessments.
func (n *NAVCENAssessments) Count() (int, error) {
	if n == nil || n.handle == nil {
		return 0, ErrClosed
	}
	v, e := n.handle.Count()
	return v, publicError(e)
}

// Assessment returns one parsed assessment by index.
func (n *NAVCENAssessments) Assessment(i int) (NAVCENAssessment, error) {
	if n == nil || n.handle == nil {
		return NAVCENAssessment{}, ErrClosed
	}
	v, e := n.handle.Assessment(i)
	if e != nil {
		return NAVCENAssessment{}, publicError(e)
	}
	out := NAVCENAssessment{System: GNSSSystem(v.System), PRN: v.PRN, SVNPresent: v.SVNPresent, SVN: v.SVN, Usable: v.Usable, ActiveNANU: v.ActiveNANU, EvaluatedAtUnixUS: v.EvaluatedAtUnixUS, Timing: v.Timing, EffectiveStartPresent: v.EffectiveStartPresent, EffectiveStartUnixUS: v.EffectiveStartUnixUS, EffectiveEndPresent: v.EffectiveEndPresent, EffectiveEndUnixUS: v.EffectiveEndUnixUS}
	out.EvaluatedAt = time.UnixMicro(v.EvaluatedAtUnixUS).UTC()
	if out.EffectiveStartPresent {
		out.EffectiveStart = time.UnixMicro(v.EffectiveStartUnixUS).UTC()
	}
	if out.EffectiveEndPresent {
		out.EffectiveEnd = time.UnixMicro(v.EffectiveEndUnixUS).UTC()
	}
	return out, nil
}

// NANUType returns the advisory type at index i.
func (n *NAVCENAssessments) NANUType(i int) (string, error) {
	if n == nil || n.handle == nil {
		return "", ErrClosed
	}
	v, e := n.handle.NANUType(i)
	return v, publicError(e)
}

// NANUSubject returns the advisory subject at index i.
func (n *NAVCENAssessments) NANUSubject(i int) (string, error) {
	if n == nil || n.handle == nil {
		return "", ErrClosed
	}
	v, e := n.handle.NANUSubject(i)
	return v, publicError(e)
}

// OutageStart returns the advisory outage start at index i.
func (n *NAVCENAssessments) OutageStart(i int) (string, error) {
	if n == nil || n.handle == nil {
		return "", ErrClosed
	}
	v, e := n.handle.OutageStart(i)
	return v, publicError(e)
}
