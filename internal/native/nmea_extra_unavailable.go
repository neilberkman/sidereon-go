//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

type NMEAChunkSummary struct {
	SentenceCount       uint64
	CompletedEpochCount uint64
	SkipCount           uint64
	WarningCount        uint64
	RetainedLength      uint64
}

type NMEAGGAOptions struct {
	Talker             string
	UTCSecondsOfDay    float64
	Position           Geodetic
	Quality            uint32
	SatellitesUsed     uint8
	HDOP               float64
	CoordinateDecimals uint8
}

type NMEAAccumulator struct{}

func NewNMEAAccumulator() (*NMEAAccumulator, error) { return nil, unavailable() }
func (*NMEAAccumulator) Close() error               { return nil }
func (*NMEAAccumulator) Push([]byte) (NMEAChunkSummary, error) {
	return NMEAChunkSummary{}, unavailable()
}
func (*NMEAAccumulator) Finish() (NMEAChunkSummary, error) {
	return NMEAChunkSummary{}, unavailable()
}
func (*NMEAAccumulator) Summary() (NMEASummary, error)   { return NMEASummary{}, unavailable() }
func (*NMEAAccumulator) RetainedLength() (uint64, error) { return 0, unavailable() }
func (*NMEAAccumulator) Epochs() ([]NMEAEpoch, error)    { return nil, unavailable() }
func WriteNMEAGGA(NMEAGGAOptions) ([]byte, error)        { return nil, unavailable() }
