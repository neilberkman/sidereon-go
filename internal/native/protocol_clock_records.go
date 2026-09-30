package native

type NativeClockSurplusValue struct {
	Position uint64
	Value    float64
}
type NativeClockRecord struct {
	RecordType                                               uint32
	Name                                                     string
	HasSatellite                                             bool
	Satellite                                                string
	CivilEpoch                                               CivilDateTime
	HasEpoch                                                 bool
	Epoch                                                    NativeClockEpoch
	Values                                                   []float64
	Surplus                                                  []NativeClockSurplusValue
	HasLine                                                  bool
	Line, LineCount                                          uint64
	Reading                                                  uint32
	HasContinuationReading                                   bool
	ContinuationReading                                      uint32
	ReadingUnknownVariant, ContinuationReadingUnknownVariant string
}
type NativeClockRecordValuesEdit struct {
	Index  int
	Values []float64
}
type NativeRinexClockInfo struct {
	HasVersion                                                                     bool
	Version                                                                        float64
	HasLayout                                                                      bool
	Layout                                                                         uint32
	HasSatelliteSystem                                                             bool
	SatelliteSystem                                                                uint32
	HasTimeSystem                                                                  bool
	TimeSystem                                                                     uint32
	TimeSystemStatus                                                               uint32
	TimeSystemLabelCount, HeaderRecordCount, RecordCount, SeriesCount, SampleCount int
	SkippedRecordCount, DiagnosticCount, NoticeCount                               int
	HasTimeScale                                                                   bool
	TimeScale                                                                      uint32
	TimeSystemUnknownVariant, TimeSystemStatusUnknownVariant                       string
}
type NativeClockHeaderRecord struct {
	HasLine                                        bool
	Line, LabelColumn                              uint64
	Reading, FieldKind                             uint32
	TextParts                                      []string
	HasVersion                                     bool
	Version                                        float64
	HasSystemCode                                  bool
	SystemCode                                     uint32
	HasCount                                       bool
	Count                                          uint64
	HasInteger                                     bool
	Integer                                        int64
	HasStart, HasStop                              bool
	Start, Stop                                    CivilDateTime
	HasConstraintS                                 bool
	ConstraintS                                    float64
	HasXYZMM                                       bool
	XYZMM                                          [3]int64
	ReadingUnknownVariant, FieldKindUnknownVariant string
	LineText, Label, Payload                       string
}
type NativeClockWriteDeparture struct {
	Kind                          uint32
	Record                        uint64
	HasEpoch                      bool
	Epoch                         NativeClockEpoch
	UnknownVariant, Name, Written string
}
type NativeClockWriteResult struct {
	Text       []byte
	Departures []NativeClockWriteDeparture
}

type NativeClockWriteFailure struct {
	Kind           uint32
	HasLine        bool
	Line           uint64
	HasTimeScale   bool
	TimeScale      uint32
	HasField       bool
	Field          string
	HasReason      bool
	Reason         string
	HasRecord      bool
	Record         string
	HasRecordType  bool
	RecordType     string
	HasValue       bool
	Value          string
	Message        string
	UnknownVariant string
}

func (failure *NativeClockWriteFailure) Error() string {
	if failure == nil {
		return "sidereon: RINEX clock write failed"
	}
	return failure.Message
}
