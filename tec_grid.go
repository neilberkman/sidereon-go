package sidereon

import "sidereon.dev/go/v3/internal/native"

// TECGridInput contains the axes and flat values for an independently queryable
// TEC grid. Values use epoch-latitude-longitude order, with longitude varying
// fastest. A nil Presence slice marks every value present; otherwise false
// entries mark missing nodes and their corresponding values are ignored.
type TECGridInput struct {
	// EpochsUnixNanos is a strictly increasing axis of Unix nanoseconds stored
	// as float64; adjacent nanoseconds may lose distinction at large values.
	EpochsUnixNanos []float64
	// LatitudesDeg is a strictly increasing latitude axis in degrees.
	LatitudesDeg []float64
	// LongitudesDeg is a strictly increasing longitude axis in degrees.
	LongitudesDeg []float64
	// ValuesTECU contains epoch-latitude-longitude values, with longitude
	// varying fastest. Entries marked absent by Presence are ignored.
	ValuesTECU []float64
	// Presence is nil (all values present) or the same length as ValuesTECU.
	// False marks a missing node; an explicit zero remains present.
	Presence []bool
}

// TECGridErrorKind identifies a rejected grid or query while preserving
// unknown native values for forward compatibility.
type TECGridErrorKind uint32

const (
	// TECGridErrorNone indicates no typed failure.
	TECGridErrorNone TECGridErrorKind = iota
	// TECGridErrorAxesTooShort means an axis has fewer than two nodes.
	TECGridErrorAxesTooShort
	// TECGridErrorAxesNotIncreasing means an axis is not strictly increasing.
	TECGridErrorAxesNotIncreasing
	// TECGridErrorDimensionsOverflow means the axis-length product overflowed.
	TECGridErrorDimensionsOverflow
	// TECGridErrorValueCountMismatch means the value count differs from the product.
	TECGridErrorValueCountMismatch
	// TECGridErrorInvalidField means shared validation rejected a named field.
	TECGridErrorInvalidField
	// TECGridErrorNodesNotAvailable means the query weights missing nodes.
	TECGridErrorNodesNotAvailable
	// TECGridErrorOutOfBounds means the query lies outside an axis.
	TECGridErrorOutOfBounds
	// TECGridErrorValueNotFinite means a present input value was non-finite.
	TECGridErrorValueNotFinite
	// TECGridErrorUnknown preserves an unrecognized native error kind.
	TECGridErrorUnknown TECGridErrorKind = 999
)

// TECGridAxis identifies the axis named by an out-of-bounds query.
type TECGridAxis uint32

const (
	// TECGridAxisNone indicates that no axis is named.
	TECGridAxisNone TECGridAxis = iota
	// TECGridAxisEpoch names the Unix-nanosecond epoch axis.
	TECGridAxisEpoch
	// TECGridAxisLatitude names the latitude axis in degrees.
	TECGridAxisLatitude
	// TECGridAxisLongitude names the longitude axis in degrees.
	TECGridAxisLongitude
	// TECGridAxisUnknown preserves an unrecognized native axis value.
	TECGridAxisUnknown TECGridAxis = 999
)

// TECGridError retains the complete typed detail reported by native grid
// construction or evaluation.
type TECGridError struct {
	// Kind identifies the failure reported by the native engine.
	Kind TECGridErrorKind
	// Status is the native status tag paired with this failure.
	Status uint32
	// Axis identifies the coordinate named by an out-of-bounds failure.
	Axis TECGridAxis
	// HasGap indicates whether Gap contains affected interpolation corners.
	HasGap bool
	// Gap retains missing-corner masks from the earlier and later epochs.
	Gap IONEXNodeGap
	// HasAxisValue indicates whether AxisValue contains the rejected coordinate.
	HasAxisValue bool
	// AxisValue uses the named axis unit: nanoseconds or degrees.
	AxisValue float64
	// ValueCount and ExpectedValueCount retain supplied and required flat counts.
	ValueCount, ExpectedValueCount int
	// HasValueIndex indicates whether ValueIndex names a rejected input value.
	HasValueIndex bool
	// ValueIndex is the zero-based flat index of a non-finite present value.
	ValueIndex uint64
	// HasField and HasReason indicate which InvalidField text parts are present.
	HasField, HasReason bool
	// Field and Reason preserve the exact native InvalidField label and reason.
	Field, Reason string
	// Message retains the full native failure message.
	Message string
}

// Error returns the native message or a stable fallback if it is empty.
func (e *TECGridError) Error() string {
	if e == nil {
		return "sidereon: TEC grid operation failed"
	}
	if e.Message != "" {
		return e.Message
	}
	return "sidereon: TEC grid operation failed"
}

// TECGridInfo reports axis lengths and the flat value/presence count.
type TECGridInfo struct {
	// EpochCount is the number of Unix-nanosecond axis entries.
	EpochCount int
	// LatitudeCount is the number of latitude entries.
	LatitudeCount int
	// LongitudeCount is the number of longitude entries.
	LongitudeCount int
	// ValueCount is the product of the three axis lengths.
	ValueCount int
}

// TECGridVTEC contains an interpolated value and the missing corners around
// which a renormalized result was computed. A strict missing-node refusal is
// returned as TECGridError with its Gap field populated instead.
type TECGridVTEC struct {
	// HasVTEC reports whether ValueTECU contains a computed result.
	HasVTEC bool
	// ValueTECU is the interpolated vertical total electron content.
	ValueTECU float64
	// Degraded identifies missing nodes skipped by renormalizing interpolation.
	Degraded IONEXNodeGap
}

// TECGrid owns a standalone time-latitude-longitude TEC surface.
type TECGrid struct {
	_      noCopy
	handle *native.TecGrid
}

func (g *TECGrid) native() *native.TecGrid {
	if g == nil {
		return nil
	}
	return g.handle
}

// NewTECGrid validates and copies an independent grid. Native input failures
// are returned as typed detail in TECGridError; marshalling failures are the
// final error.
func NewTECGrid(input TECGridInput) (*TECGrid, *TECGridError, error) {
	h, failure, err := native.NewTecGrid(native.TecGridInput{EpochsUnixNanos: input.EpochsUnixNanos, LatitudesDeg: input.LatitudesDeg, LongitudesDeg: input.LongitudesDeg, ValuesTECU: input.ValuesTECU, Presence: input.Presence})
	if err != nil {
		return nil, nil, publicError(err)
	}
	if failure != nil {
		return nil, publicTECGridError(failure), nil
	}
	return &TECGrid{handle: h}, nil, nil
}

func publicTECGridError(v *native.NativeTecGridError) *TECGridError {
	if v == nil {
		return nil
	}
	return &TECGridError{Kind: TECGridErrorKind(v.Kind), Status: v.Status, Axis: TECGridAxis(v.Axis), HasGap: v.HasGap,
		Gap: IONEXNodeGap{HasGap: v.Gap.HasGap, Earlier: v.Gap.Earlier, Later: v.Gap.Later}, HasAxisValue: v.HasAxisValue,
		AxisValue: v.AxisValue, ValueCount: v.ValueCount, ExpectedValueCount: v.ExpectedValueCount, HasValueIndex: v.HasValueIndex,
		ValueIndex: v.ValueIndex, HasField: v.HasField, HasReason: v.HasReason, Field: v.Field, Reason: v.Reason, Message: v.Message}
}

// Close releases the native grid. It is safe to call more than once.
func (g *TECGrid) Close() error { return publicError(g.native().Close()) }

// Dimensions reports axis lengths and the product-sized data count.
func (g *TECGrid) Dimensions() (TECGridInfo, error) {
	v, e := g.native().Dimensions()
	return TECGridInfo{v.EpochCount, v.LatitudeCount, v.LongitudeCount, v.ValueCount}, publicError(e)
}

// EpochsUnixNanos returns a detached copy of the epoch axis.
func (g *TECGrid) EpochsUnixNanos() ([]float64, error) {
	v, e := g.native().Epochs()
	return v, publicError(e)
}

// LatitudesDeg returns a detached copy of the latitude axis.
func (g *TECGrid) LatitudesDeg() ([]float64, error) {
	v, e := g.native().Latitudes()
	return v, publicError(e)
}

// LongitudesDeg returns a detached copy of the longitude axis.
func (g *TECGrid) LongitudesDeg() ([]float64, error) {
	v, e := g.native().Longitudes()
	return v, publicError(e)
}

// ValuesTECU returns detached values in epoch-latitude-longitude order.
// Missing nodes are represented by NaN; use ValuePresence as the authority.
func (g *TECGrid) ValuesTECU() ([]float64, error) {
	v, e := g.native().Values()
	return v, publicError(e)
}

// ValuePresence returns a detached presence mask in the same order as values.
func (g *TECGrid) ValuePresence() ([]bool, error) {
	v, e := g.native().Presence()
	return v, publicError(e)
}

// VTECAtPiercePoint interpolates at an exact Unix-nanosecond epoch and a
// latitude/longitude in degrees. Latitude is clamped by the native engine to
// [-87.5, 87.5]. Strict mode refuses weighted missing nodes; renormalizing
// mode returns a value and identifies the omitted corners in Degraded.
func (g *TECGrid) VTECAtPiercePoint(epochUnixNanos int64, latitudeDeg, longitudeDeg float64, policy IONEXMissingNodePolicy) (TECGridVTEC, *TECGridError, error) {
	v, e := g.native().VTECAt(epochUnixNanos, latitudeDeg, longitudeDeg, uint32(policy))
	if failure, ok := e.(*native.NativeTecGridError); ok {
		return TECGridVTEC{}, publicTECGridError(failure), nil
	}
	if e != nil {
		return TECGridVTEC{}, nil, publicError(e)
	}
	return TECGridVTEC{HasVTEC: v.HasVTEC, ValueTECU: v.ValueTECU, Degraded: IONEXNodeGap{HasGap: v.Degraded.HasGap, Earlier: v.Degraded.Earlier, Later: v.Degraded.Later}}, nil, nil
}
