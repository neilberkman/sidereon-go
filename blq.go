package sidereon

import "sidereon.dev/go/v3/internal/native"

// BLQCommentPlacement identifies where a retained BLQ comment appears in its block.
type BLQCommentPlacement uint32

const (
	// BLQBeforeStation places the line before the station name.
	BLQBeforeStation BLQCommentPlacement = 0
	// BLQBeforeRow places the line before the indexed coefficient row.
	BLQBeforeRow BLQCommentPlacement = 1
	// BLQAfterRows places the line after the sixth coefficient row.
	BLQAfterRows BLQCommentPlacement = 2
)

// BLQErrorKind identifies the family of a typed BLQ refusal.
type BLQErrorKind uint32

const (
	// BLQErrorNone means the operation succeeded.
	BLQErrorNone BLQErrorKind = 0
	// BLQErrorParse reports a parser refusal.
	BLQErrorParse BLQErrorKind = 1
	// BLQErrorWrite reports a writer refusal.
	BLQErrorWrite BLQErrorKind = 2
	// BLQErrorOther reports another engine failure.
	BLQErrorOther BLQErrorKind = 999
)

// BLQComment is one retained comment or column-order header line.
type BLQComment struct {
	// Placement is the location of the retained line in its block.
	Placement BLQCommentPlacement
	// Row is the zero-based coefficient row for BLQBeforeRow placements.
	Row int
	// Text is the retained line without its line terminator.
	Text string
}

// BLQParseErrorKind identifies the parser's typed refusal reason.
type BLQParseErrorKind uint32

const (
	// BLQParseNone denotes no parser refusal.
	BLQParseNone BLQParseErrorKind = 0
	// BLQParseEmpty means the input contains only whitespace.
	BLQParseEmpty BLQParseErrorKind = 1
	// BLQParseMissingStation means a coefficient row appeared before a station.
	BLQParseMissingStation BLQParseErrorKind = 2
	// BLQParseMissingCoefficientRows means a station ended before six rows.
	BLQParseMissingCoefficientRows BLQParseErrorKind = 3
	// BLQParseTooManyCoefficientRows means a station had more than six rows.
	BLQParseTooManyCoefficientRows BLQParseErrorKind = 4
	// BLQParseWrongColumnCount means a row or header had an unexpected number of values.
	BLQParseWrongColumnCount BLQParseErrorKind = 5
	// BLQParseInvalidNumber means a coefficient token was not numeric.
	BLQParseInvalidNumber BLQParseErrorKind = 6
	// BLQParseNonFiniteNumber means a coefficient was NaN or infinite.
	BLQParseNonFiniteNumber BLQParseErrorKind = 7
	// BLQParseUnsupportedConstituent means a column header named an unknown constituent.
	BLQParseUnsupportedConstituent BLQParseErrorKind = 8
	// BLQParseDuplicateConstituent means a column header repeated a constituent.
	BLQParseDuplicateConstituent BLQParseErrorKind = 9
	// BLQParseMultipleBlocks means a single-block parser encountered several blocks.
	BLQParseMultipleBlocks BLQParseErrorKind = 10
)

// BLQWriteErrorKind identifies the writer's typed refusal reason.
type BLQWriteErrorKind uint32

const (
	// BLQWriteNone denotes no writer refusal.
	BLQWriteNone BLQWriteErrorKind = 0
	// BLQWriteEmptyStation means the station identifier is empty.
	BLQWriteEmptyStation BLQWriteErrorKind = 1
	// BLQWriteStationLineBreak means the station identifier contains a line break.
	BLQWriteStationLineBreak BLQWriteErrorKind = 2
	// BLQWriteStationWhitespace means the station has surrounding whitespace.
	BLQWriteStationWhitespace BLQWriteErrorKind = 3
	// BLQWriteStationComment means the station would parse as a comment.
	BLQWriteStationComment BLQWriteErrorKind = 4
	// BLQWriteStationHeader means the station would parse as a column header.
	BLQWriteStationHeader BLQWriteErrorKind = 5
	// BLQWriteStationCoefficientRow means the station would parse as a coefficient row.
	BLQWriteStationCoefficientRow BLQWriteErrorKind = 6
	// BLQWriteNonFiniteCoefficient means a coefficient is NaN or infinite.
	BLQWriteNonFiniteCoefficient BLQWriteErrorKind = 7
	// BLQWriteCommentLineBreak means a retained comment contains a line break.
	BLQWriteCommentLineBreak BLQWriteErrorKind = 8
	// BLQWriteNotCommentLine means a retained line is blank or not a comment/header.
	BLQWriteNotCommentLine BLQWriteErrorKind = 9
	// BLQWriteCommentPlacement means a retained line has an invalid row placement.
	BLQWriteCommentPlacement BLQWriteErrorKind = 10
	// BLQWriteInvalidHeader means a retained column header is invalid.
	BLQWriteInvalidHeader BLQWriteErrorKind = 11
	// BLQWriteAfterRowsBeforeBlock means trailing comments precede another block.
	BLQWriteAfterRowsBeforeBlock BLQWriteErrorKind = 12
	// BLQWriteCommentsOutOfOrder means retained lines are not in file order.
	BLQWriteCommentsOutOfOrder BLQWriteErrorKind = 13
)

// BLQTideConstituent identifies one constituent column in a BLQ coefficient table.
type BLQTideConstituent uint32

const (
	// BLQTideM2 identifies M2.
	BLQTideM2 BLQTideConstituent = iota
	// BLQTideS2 identifies S2.
	BLQTideS2
	// BLQTideN2 identifies N2.
	BLQTideN2
	// BLQTideK2 identifies K2.
	BLQTideK2
	// BLQTideK1 identifies K1.
	BLQTideK1
	// BLQTideO1 identifies O1.
	BLQTideO1
	// BLQTideP1 identifies P1.
	BLQTideP1
	// BLQTideQ1 identifies Q1.
	BLQTideQ1
	// BLQTideMf identifies Mf.
	BLQTideMf
	// BLQTideMm identifies Mm.
	BLQTideMm
	// BLQTideSsa identifies Ssa.
	BLQTideSsa
)

// BLQError contains the typed parser or writer refusal and all available detail.
type BLQError struct {
	// Kind identifies the error family.
	Kind BLQErrorKind
	// ParseKind identifies a parser refusal or nested invalid-header refusal.
	ParseKind BLQParseErrorKind
	// WriteKind identifies a writer refusal.
	WriteKind BLQWriteErrorKind
	// Presence flags specify which associated fields carry values.
	HasLine, HasExpected, HasFound, HasText, HasBlock, HasCoefficient, HasCommentIndex bool
	// Line, Expected, Found, Block, Row, and CommentIndex retain indexed details.
	Line, Expected, Found, Block, Row, CommentIndex int
	// Constituent identifies the non-finite coefficient's tide constituent.
	Constituent BLQTideConstituent
	// Message is the full native refusal; Text is its station, token, or label.
	Message, Text string
}

// BLQOutcome records whether a parse or write operation succeeded and, when refused, why.
type BLQOutcome struct {
	// IsOK reports whether the parse or write operation succeeded.
	IsOK bool
	// Error contains typed failure detail; its kind is zero on success.
	Error BLQError
}

func publicBLQOutcome(value native.BLQOutcome) BLQOutcome {
	e := value.Error
	return BLQOutcome{IsOK: value.IsOK, Error: BLQError{Kind: BLQErrorKind(e.Kind), ParseKind: BLQParseErrorKind(e.ParseKind), WriteKind: BLQWriteErrorKind(e.WriteKind), HasLine: e.HasLine, HasExpected: e.HasExpected, HasFound: e.HasFound, HasText: e.HasText, HasBlock: e.HasBlock, HasCoefficient: e.HasCoefficient, HasCommentIndex: e.HasCommentIndex, Line: e.Line, Expected: e.Expected, Found: e.Found, Block: e.Block, Row: e.Row, CommentIndex: e.CommentIndex, Constituent: BLQTideConstituent(e.Constituent), Message: e.Message, Text: e.Text}}
}

// BLQBlocks owns an ordered list of BLQ station blocks and retained comment lines.
type BLQBlocks struct {
	_     noCopy
	inner *native.BLQBlocks
}

// ParseBLQ parses station blocks while retaining comments and column-order headers.
func ParseBLQ(data []byte) (*BLQBlocks, BLQOutcome, error) {
	value, outcome, err := native.ParseBLQ(data)
	publicOutcome := publicBLQOutcome(outcome)
	if err != nil || value == nil {
		return nil, publicOutcome, publicError(err)
	}
	return &BLQBlocks{inner: value}, publicOutcome, nil
}

// NewBLQBlocks creates an empty list that can be populated with Push and PushComment.
func NewBLQBlocks() (*BLQBlocks, error) {
	value, err := native.NewBLQBlocks()
	if err != nil {
		return nil, publicError(err)
	}
	return &BLQBlocks{inner: value}, nil
}

// Close releases the native block list. It is safe to call more than once.
func (b *BLQBlocks) Close() error {
	if b == nil || b.inner == nil {
		return nil
	}
	return publicError(b.inner.Close())
}

// Count returns the number of station blocks.
func (b *BLQBlocks) Count() (int, error) {
	if b == nil || b.inner == nil {
		return 0, ErrClosed
	}
	v, e := b.inner.Count()
	return v, publicError(e)
}

// Station returns the trimmed station identifier at index.
func (b *BLQBlocks) Station(index int) (string, error) {
	if b == nil || b.inner == nil {
		return "", ErrClosed
	}
	v, e := b.inner.Station(index)
	return v, publicError(e)
}

// Coefficients returns one block's 3-row amplitude and phase table.
func (b *BLQBlocks) Coefficients(index int) (OceanLoadingBLQ, error) {
	if b == nil || b.inner == nil {
		return OceanLoadingBLQ{}, ErrClosed
	}
	v, err := b.inner.Coefficients(index)
	return OceanLoadingBLQ{AmplitudeM: v.AmplitudeM, PhaseDeg: v.PhaseDeg}, publicError(err)
}

// CommentCount returns the number of retained lines for a block.
func (b *BLQBlocks) CommentCount(index int) (int, error) {
	if b == nil || b.inner == nil {
		return 0, ErrClosed
	}
	v, e := b.inner.CommentCount(index)
	return v, publicError(e)
}

// Comment returns a detached retained comment line in input order.
func (b *BLQBlocks) Comment(index, commentIndex int) (BLQComment, error) {
	if b == nil || b.inner == nil {
		return BLQComment{}, ErrClosed
	}
	v, err := b.inner.Comment(index, commentIndex)
	return BLQComment{Placement: BLQCommentPlacement(v.Placement), Row: v.Row, Text: v.Text}, publicError(err)
}

// Push appends a station block. Writer refusals are reported by Encode's outcome.
func (b *BLQBlocks) Push(station string, coefficients OceanLoadingBLQ) error {
	if b == nil || b.inner == nil {
		return ErrClosed
	}
	return publicError(b.inner.Push(station, native.OceanLoadingBLQ{AmplitudeM: coefficients.AmplitudeM, PhaseDeg: coefficients.PhaseDeg}))
}

// PushComment appends a retained line at its declared placement.
func (b *BLQBlocks) PushComment(index int, comment BLQComment) error {
	if b == nil || b.inner == nil {
		return ErrClosed
	}
	return publicError(b.inner.PushComment(index, native.BLQComment{Placement: uint32(comment.Placement), Row: comment.Row, Text: comment.Text}))
}

// Encode writes every block, returning a typed refusal separately from call errors.
func (b *BLQBlocks) Encode() ([]byte, BLQOutcome, error) {
	if b == nil || b.inner == nil {
		return nil, BLQOutcome{}, ErrClosed
	}
	v, o, e := b.inner.Encode()
	return v, publicBLQOutcome(o), publicError(e)
}

// BlockToText writes one block as six coefficient rows.
func (b *BLQBlocks) BlockToText(index int) ([]byte, BLQOutcome, error) {
	if b == nil || b.inner == nil {
		return nil, BLQOutcome{}, ErrClosed
	}
	v, o, e := b.inner.BlockToText(index)
	return v, publicBLQOutcome(o), publicError(e)
}
