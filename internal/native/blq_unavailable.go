//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

type BLQComment struct {
	Placement uint32
	Row       int
	Text      string
}
type BLQError struct {
	Kind, ParseKind, WriteKind                                                         uint32
	HasLine, HasExpected, HasFound, HasText, HasBlock, HasCoefficient, HasCommentIndex bool
	Line, Expected, Found, Block, Row, CommentIndex                                    int
	Constituent                                                                        uint32
	Message, Text                                                                      string
}
type BLQOutcome struct {
	IsOK  bool
	Error BLQError
}
type BLQBlocks struct{}

func ParseBLQ([]byte) (*BLQBlocks, BLQOutcome, error) {
	return nil, BLQOutcome{}, protocolUnavailable()
}
func NewBLQBlocks() (*BLQBlocks, error)        { return nil, protocolUnavailable() }
func (*BLQBlocks) Close() error                { return nil }
func (*BLQBlocks) Count() (int, error)         { return 0, protocolUnavailable() }
func (*BLQBlocks) Station(int) (string, error) { return "", protocolUnavailable() }
func (*BLQBlocks) Coefficients(int) (OceanLoadingBLQ, error) {
	return OceanLoadingBLQ{}, protocolUnavailable()
}
func (*BLQBlocks) CommentCount(int) (int, error)        { return 0, protocolUnavailable() }
func (*BLQBlocks) Comment(int, int) (BLQComment, error) { return BLQComment{}, protocolUnavailable() }
func (*BLQBlocks) Push(string, OceanLoadingBLQ) error   { return protocolUnavailable() }
func (*BLQBlocks) PushComment(int, BLQComment) error    { return protocolUnavailable() }
func (*BLQBlocks) Encode() ([]byte, BLQOutcome, error) {
	return nil, BLQOutcome{}, protocolUnavailable()
}
func (*BLQBlocks) BlockToText(int) ([]byte, BLQOutcome, error) {
	return nil, BLQOutcome{}, protocolUnavailable()
}
