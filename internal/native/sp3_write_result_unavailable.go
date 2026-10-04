//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

type SP3WriteError struct {
	Kind                                                                                uint32
	HasField, HasTextValue, HasSatelliteID, HasEpochIndex, HasCommentIndex              bool
	HasColumns, HasDecimals, HasIntegerValue, HasNumber, HasYear                        bool
	HasFieldSeconds, HasResidualS, HasEpochTimeScale, HasHeaderTimeScale, HasTimeSystem bool
	HasDeclaredEpochs, HasEpochs, HasEntries, HasSatellites, HasCodes                   bool
	HasStored, HasNative, HasColumnValue, HasExponent                                   bool
	SatelliteID, TimeSystem                                                             string
	EpochIndex, CommentIndex, Columns, Decimals                                         int
	IntegerValue                                                                        uint64
	Number                                                                              float64
	Year                                                                                int64
	FieldSeconds, ResidualS                                                             float64
	EpochTimeScale, HeaderTimeScale                                                     uint32
	DeclaredEpochs                                                                      uint64
	Epochs, Entries, Satellites, Codes                                                  int
	Stored, Native, ColumnValue                                                         float64
	Exponent                                                                            int16
	Message, Field, TextValue                                                           string
}
type SP3WriteOutcome struct {
	IsOK   bool
	Status uint32
	Error  SP3WriteError
}

func (*SP3) TextResult() ([]byte, SP3WriteOutcome, error) {
	return nil, SP3WriteOutcome{}, unavailable()
}
