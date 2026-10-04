package sidereon

// SP3MergeAgreementMetric is agreement metadata for one accepted output cell.
type SP3MergeAgreementMetric struct {
	// Epoch and EpochJ2000S identify this cell's epoch in exact and approximate form.
	Epoch       ClockEpoch
	EpochJ2000S float64
	// Satellite is the cell's satellite token.
	Satellite string
	// PositionMembers is the accepted position consensus size.
	PositionMembers int
	// PositionRMSPresent reports whether PositionRMSM is present.
	PositionRMSPresent bool
	// PositionRMSM is the position consensus RMS in metres.
	PositionRMSM float64
	// PositionMaxPresent reports whether PositionMaxM is present.
	PositionMaxPresent bool
	// PositionMaxM is the maximum position difference in metres.
	PositionMaxM float64
	// ClockMembers is the accepted clock consensus size.
	ClockMembers int
	// ClockRMSPresent reports whether ClockRMSS is present.
	ClockRMSPresent bool
	// ClockRMSS is the clock consensus RMS in seconds.
	ClockRMSS float64
	// ClockMaxPresent reports whether ClockMaxS is present.
	ClockMaxPresent bool
	// ClockMaxS is the maximum clock difference in seconds.
	ClockMaxS float64
}

// SP3MergeClockOmission records one source clock omitted from the merged output.
type SP3MergeClockOmission struct {
	// Epoch and EpochJ2000S identify the omitted clock's epoch.
	Epoch       ClockEpoch
	EpochJ2000S float64
	// Satellite is the satellite token for the omitted clock.
	Satellite string
	// Source is the omitted clock contributor's zero-based input index.
	Source int
	// Reason is the native omission reason code.
	Reason SP3MergeClockOmissionReason
	// HasPreferred reports whether Preferred names the selected contributor.
	HasPreferred bool
	// Preferred is the selected clock contributor's zero-based input index.
	Preferred int
	// CellHasClock reports whether the output cell retained a clock value.
	CellHasClock bool
}

// SP3MergeClockOmissionReason identifies why a source clock was omitted.
type SP3MergeClockOmissionReason uint32

const (
	// SP3ClockOmissionDatumNotObservable means the datum offset could not be estimated.
	SP3ClockOmissionDatumNotObservable SP3MergeClockOmissionReason = 0
	// SP3ClockOmissionPreferredSourceWithoutClock means precedence selected a source without a clock.
	SP3ClockOmissionPreferredSourceWithoutClock SP3MergeClockOmissionReason = 1
	// SP3ClockOmissionNoConsensus means no agreeing clock subset met the consensus rule.
	SP3ClockOmissionNoConsensus SP3MergeClockOmissionReason = 2
)

// SP3DroppedInputEpoch records an input epoch that took no part in the merge.
type SP3DroppedInputEpoch struct {
	// Source is the zero-based input index.
	Source int
	// EpochIndex is the source epoch's zero-based index.
	EpochIndex int
	// Epoch and EpochJ2000S identify the exact and approximate source epoch.
	Epoch       ClockEpoch
	EpochJ2000S float64
	// Reason is the native drop reason code.
	Reason SP3DroppedEpochReason
}

// SP3DroppedEpochReason identifies why an input epoch did not take part.
type SP3DroppedEpochReason uint32

const (
	// SP3DroppedEpochOffTargetGrid means the epoch was absent from the explicit target grid.
	SP3DroppedEpochOffTargetGrid SP3DroppedEpochReason = 0
	// SP3DroppedEpochNotOnTickAxis means the epoch was not on the SP3 10 ns tick axis.
	SP3DroppedEpochNotOnTickAxis SP3DroppedEpochReason = 1
)

// SP3MergeOmittedEpoch records a union-grid epoch with no accepted cell.
type SP3MergeOmittedEpoch struct {
	// Epoch and EpochJ2000S identify an omitted union-grid epoch.
	Epoch       ClockEpoch
	EpochJ2000S float64
}

// SP3MergeProvenanceInfo reports whether and how the merge recorded provenance.
type SP3MergeProvenanceInfo struct {
	// Recorded reports whether any provenance was retained.
	Recorded bool
	// Mode is the requested provenance retention level.
	Mode SP3ProvenanceMode
	// CellCount is the retained per-cell selection count.
	CellCount int
	// TransitionCount is the retained source-transition count.
	TransitionCount int
	// CoverageCount is the retained per-source coverage count.
	CoverageCount int
}

// SP3MergeCellSelection identifies how one output-cell channel was selected.
type SP3MergeCellSelection struct {
	// Kind identifies single-source, precedence, or combined selection.
	Kind SP3CellSelectionKind
	// HasSource reports whether Source identifies the value's supplier.
	HasSource bool
	// Source is the zero-based input index supplying the value.
	Source int
	// HasRule reports whether Rule identifies a combining rule.
	HasRule bool
	// Rule is the rule used to combine accepted members.
	Rule SP3MergeCombine
	// MemberCount is the number of accepted consensus sources.
	MemberCount int
}

// SP3CellSelectionKind identifies how a merge selected one cell channel.
type SP3CellSelectionKind uint32

const (
	// SP3CellSelectionSingleSource means one source supplied the value.
	SP3CellSelectionSingleSource SP3CellSelectionKind = 0
	// SP3CellSelectionPrecedence means precedence chose one source from the consensus.
	SP3CellSelectionPrecedence SP3CellSelectionKind = 1
	// SP3CellSelectionCombined means the value combines consensus members.
	SP3CellSelectionCombined SP3CellSelectionKind = 2
)

// SP3MergeCellProvenance records exact channel selections for one output cell.
type SP3MergeCellProvenance struct {
	// Epoch and EpochJ2000S identify the output cell's epoch.
	Epoch       ClockEpoch
	EpochJ2000S float64
	// Satellite is the output cell's satellite token.
	Satellite string
	// HasPosition reports whether the cell has a position value.
	HasPosition bool
	// Position describes the selected position value.
	Position SP3MergeCellSelection
	// HasClock reports whether the cell has a clock value.
	HasClock bool
	// Clock describes the selected clock value.
	Clock SP3MergeCellSelection
}

// SP3ContributorCoverage reports accepted-cell contribution for one source.
type SP3ContributorCoverage struct {
	// Source is the zero-based input index.
	Source int
	// CellsContributed counts accepted cells with any contribution from Source.
	CellsContributed int
	// CellsSelected counts accepted cells whose position Source alone supplied.
	CellsSelected int
	// HasFirstEpoch reports whether FirstEpoch is present.
	HasFirstEpoch bool
	// FirstEpoch and FirstEpochJ2000S identify the first contributed epoch.
	FirstEpoch       ClockEpoch
	FirstEpochJ2000S float64
	// HasLastEpoch reports whether LastEpoch is present.
	HasLastEpoch bool
	// LastEpoch and LastEpochJ2000S identify the last contributed epoch.
	LastEpoch       ClockEpoch
	LastEpochJ2000S float64
	// CellsAbsent counts accepted cells with no contribution from Source.
	CellsAbsent int
}

// SP3PrecedenceTransition records a change in the source supplying a satellite position.
type SP3PrecedenceTransition struct {
	// Satellite identifies the transitioned satellite.
	Satellite string
	// Epoch and EpochJ2000S identify the transition epoch.
	Epoch       ClockEpoch
	EpochJ2000S float64
	// HasFromSource reports whether FromSource is present.
	HasFromSource bool
	// FromSource is the previous zero-based source index.
	FromSource int
	// HasToSource reports whether ToSource is present.
	HasToSource bool
	// ToSource is the new zero-based source index.
	ToSource int
	// Reason explains why selection changed.
	Reason SP3TransitionReason
}

// SP3TransitionReason explains why the source supplying a satellite position changed.
type SP3TransitionReason uint32

const (
	// SP3TransitionSoleAvailability means the prior source no longer carried the cell.
	SP3TransitionSoleAvailability SP3TransitionReason = 0
	// SP3TransitionPrecedence means precedence selected another available source.
	SP3TransitionPrecedence SP3TransitionReason = 1
	// SP3TransitionOutlierRejection means the prior source became an outlier.
	SP3TransitionOutlierRejection SP3TransitionReason = 2
	// SP3TransitionConsensusChange means the cell selection changed between consensus modes.
	SP3TransitionConsensusChange SP3TransitionReason = 3
)

// SP3MergeChannel identifies a provenance cell channel.
type SP3MergeChannel uint32

const (
	// SP3MergePositionChannel selects position provenance members.
	SP3MergePositionChannel SP3MergeChannel = iota
	// SP3MergeClockChannel selects clock provenance members.
	SP3MergeClockChannel
)

// AgreementMetrics returns one detached agreement row per accepted cell.
func (r *SP3MergeReport) AgreementMetrics() ([]SP3MergeAgreementMetric, error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	v, e := r.handle.AgreementMetrics()
	if e != nil {
		return nil, publicError(e)
	}
	out := make([]SP3MergeAgreementMetric, len(v))
	for i, x := range v {
		out[i] = SP3MergeAgreementMetric{Epoch: publicClockEpoch(x.Epoch), EpochJ2000S: x.EpochJ2000S, Satellite: x.Satellite, PositionMembers: x.PositionMembers, PositionRMSPresent: x.PositionRMSPresent, PositionRMSM: x.PositionRMSM, PositionMaxPresent: x.PositionMaxPresent, PositionMaxM: x.PositionMaxM, ClockMembers: x.ClockMembers, ClockRMSPresent: x.ClockRMSPresent, ClockRMSS: x.ClockRMSS, ClockMaxPresent: x.ClockMaxPresent, ClockMaxS: x.ClockMaxS}
	}
	return out, nil
}

// ClockOmissions returns every source clock the merge did not write.
func (r *SP3MergeReport) ClockOmissions() ([]SP3MergeClockOmission, error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	v, e := r.handle.ClockOmissions()
	if e != nil {
		return nil, publicError(e)
	}
	out := make([]SP3MergeClockOmission, len(v))
	for i, x := range v {
		out[i] = SP3MergeClockOmission{Epoch: publicClockEpoch(x.Epoch), EpochJ2000S: x.EpochJ2000S, Satellite: x.Satellite, Source: x.Source, Reason: SP3MergeClockOmissionReason(x.Reason), HasPreferred: x.HasPreferred, Preferred: x.Preferred, CellHasClock: x.CellHasClock}
	}
	return out, nil
}

// DroppedInputEpochs returns source epochs that contributed no merged cell.
func (r *SP3MergeReport) DroppedInputEpochs() ([]SP3DroppedInputEpoch, error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	v, e := r.handle.DroppedInputEpochs()
	if e != nil {
		return nil, publicError(e)
	}
	out := make([]SP3DroppedInputEpoch, len(v))
	for i, x := range v {
		out[i] = SP3DroppedInputEpoch{Source: x.Source, EpochIndex: x.EpochIndex, Epoch: publicClockEpoch(x.Epoch), EpochJ2000S: x.EpochJ2000S, Reason: SP3DroppedEpochReason(x.Reason)}
	}
	return out, nil
}

// OmittedEpochs returns union-grid epochs at which no cell was accepted.
func (r *SP3MergeReport) OmittedEpochs() ([]SP3MergeOmittedEpoch, error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	v, e := r.handle.OmittedEpochs()
	if e != nil {
		return nil, publicError(e)
	}
	out := make([]SP3MergeOmittedEpoch, len(v))
	for i, x := range v {
		out[i] = SP3MergeOmittedEpoch{Epoch: publicClockEpoch(x.Epoch), EpochJ2000S: x.EpochJ2000S}
	}
	return out, nil
}

// ProvenanceInfo returns whether provenance was recorded and the sizes of its lists.
func (r *SP3MergeReport) ProvenanceInfo() (SP3MergeProvenanceInfo, error) {
	if r == nil || r.handle == nil {
		return SP3MergeProvenanceInfo{}, ErrClosed
	}
	v, e := r.handle.ProvenanceInfo()
	if e != nil {
		return SP3MergeProvenanceInfo{}, publicError(e)
	}
	return SP3MergeProvenanceInfo{Recorded: v.Recorded, Mode: SP3ProvenanceMode(v.Mode), CellCount: v.CellCount, TransitionCount: v.TransitionCount, CoverageCount: v.CoverageCount}, nil
}

// ProvenanceCells returns detached full-mode per-cell selection records.
func (r *SP3MergeReport) ProvenanceCells() ([]SP3MergeCellProvenance, error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	v, e := r.handle.ProvenanceCells()
	if e != nil {
		return nil, publicError(e)
	}
	out := make([]SP3MergeCellProvenance, len(v))
	for i, x := range v {
		out[i] = SP3MergeCellProvenance{Epoch: publicClockEpoch(x.Epoch), EpochJ2000S: x.EpochJ2000S, Satellite: x.Satellite, HasPosition: x.HasPosition, Position: SP3MergeCellSelection{Kind: SP3CellSelectionKind(x.Position.Kind), HasSource: x.Position.HasSource, Source: x.Position.Source, HasRule: x.Position.HasRule, Rule: SP3MergeCombine(x.Position.Rule), MemberCount: x.Position.MemberCount}, HasClock: x.HasClock, Clock: SP3MergeCellSelection{Kind: SP3CellSelectionKind(x.Clock.Kind), HasSource: x.Clock.HasSource, Source: x.Clock.Source, HasRule: x.Clock.HasRule, Rule: SP3MergeCombine(x.Clock.Rule), MemberCount: x.Clock.MemberCount}}
	}
	return out, nil
}

// ProvenanceCellMembers returns sorted source indices in one cell channel's consensus.
func (r *SP3MergeReport) ProvenanceCellMembers(index int, channel SP3MergeChannel) ([]int, error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	v, e := r.handle.ProvenanceCellMembers(index, uint32(channel))
	if e != nil {
		return nil, publicError(e)
	}
	return append([]int(nil), v...), nil
}

// ProvenanceCoverage returns one detached contribution summary per input source.
func (r *SP3MergeReport) ProvenanceCoverage() ([]SP3ContributorCoverage, error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	v, e := r.handle.ProvenanceCoverage()
	if e != nil {
		return nil, publicError(e)
	}
	out := make([]SP3ContributorCoverage, len(v))
	for i, x := range v {
		out[i] = SP3ContributorCoverage{Source: x.Source, CellsContributed: x.CellsContributed, CellsSelected: x.CellsSelected, HasFirstEpoch: x.HasFirstEpoch, FirstEpoch: publicClockEpoch(x.FirstEpoch), FirstEpochJ2000S: x.FirstEpochJ2000S, HasLastEpoch: x.HasLastEpoch, LastEpoch: publicClockEpoch(x.LastEpoch), LastEpochJ2000S: x.LastEpochJ2000S, CellsAbsent: x.CellsAbsent}
	}
	return out, nil
}

// ProvenanceTransitions returns source-selection changes in output order.
func (r *SP3MergeReport) ProvenanceTransitions() ([]SP3PrecedenceTransition, error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	v, e := r.handle.ProvenanceTransitions()
	if e != nil {
		return nil, publicError(e)
	}
	out := make([]SP3PrecedenceTransition, len(v))
	for i, x := range v {
		out[i] = SP3PrecedenceTransition{Satellite: x.Satellite, Epoch: publicClockEpoch(x.Epoch), EpochJ2000S: x.EpochJ2000S, HasFromSource: x.HasFromSource, FromSource: x.FromSource, HasToSource: x.HasToSource, ToSource: x.ToSource, Reason: SP3TransitionReason(x.Reason)}
	}
	return out, nil
}

// ContinuityJSON returns the merge's optional continuity post-condition report.
func (r *SP3MergeReport) ContinuityJSON() ([]byte, error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	v, e := r.handle.ContinuityJSON()
	return append([]byte(nil), v...), publicError(e)
}

// ContinuitySelectedNodes returns merged interpolation nodes and whether the report checked them.
func (r *SP3MergeReport) ContinuitySelectedNodes(satellite string, fromJ2000S, throughJ2000S float64) (bool, []float64, error) {
	if r == nil || r.handle == nil {
		return false, nil, ErrClosed
	}
	ok, v, e := r.handle.ContinuitySelectedNodes(satellite, fromJ2000S, throughJ2000S)
	return ok, append([]float64(nil), v...), publicError(e)
}
