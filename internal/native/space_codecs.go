package native

type NativeBiasLookup struct {
	Status                  uint32
	Value                   float64
	RecordIndices           []uint64
	OverriddenRecordIndices []uint64
	RecordIndex             uint64
	HasProductTimeScale     bool
	ProductTimeScale        uint32
	HasQueryTimeScale       bool
	QueryTimeScale          uint32
	Observable              string
	UnknownVariant          string
}

type NativeBiasModeInfo struct {
	Mode         uint32
	HasTimeScale bool
	TimeScale    uint32
}
