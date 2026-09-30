package native

type TecGridInput struct {
	EpochsUnixNanos []float64
	LatitudesDeg    []float64
	LongitudesDeg   []float64
	ValuesTECU      []float64
	Presence        []bool
}

type NativeTecGridInfo struct{ EpochCount, LatitudeCount, LongitudeCount, ValueCount int }
type NativeTecGridError struct {
	Kind, Status, Axis             uint32
	HasGap                         bool
	Gap                            IonexNodeGap
	HasAxisValue                   bool
	AxisValue                      float64
	ValueCount, ExpectedValueCount int
	HasValueIndex                  bool
	ValueIndex                     uint64
	HasField, HasReason            bool
	Field, Reason, Message         string
}

func (e *NativeTecGridError) Error() string {
	if e == nil || e.Message == "" {
		return "sidereon: TEC grid operation failed"
	}
	return e.Message
}

type NativeTecGridVTEC struct {
	HasVTEC   bool
	ValueTECU float64
	Degraded  IonexNodeGap
}
