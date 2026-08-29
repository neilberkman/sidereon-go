//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

type NTRIPVersion uint32

const (
	NTRIPVersionRev1 NTRIPVersion = 1
	NTRIPVersionRev2 NTRIPVersion = 2
)

type NTRIPState uint32

const (
	NTRIPStateIdle            NTRIPState = 0
	NTRIPStateAwaitingStatus  NTRIPState = 1
	NTRIPStateAwaitingHeaders NTRIPState = 2
	NTRIPStateStreaming       NTRIPState = 3
	NTRIPStateSourcetable     NTRIPState = 4
	NTRIPStateClosed          NTRIPState = 5
)

type NTRIPEventKind uint32

const (
	NTRIPEventConnected       NTRIPEventKind = 0
	NTRIPEventPayload         NTRIPEventKind = 1
	NTRIPEventSourcetable     NTRIPEventKind = 2
	NTRIPEventRejected        NTRIPEventKind = 3
	NTRIPEventStreamCorrupted NTRIPEventKind = 4
	NTRIPEventStreamEnded     NTRIPEventKind = 5
)

type NTRIPRejectionKind uint32

const (
	NTRIPRejectionNone                  NTRIPRejectionKind = 0
	NTRIPRejectionUnauthorized          NTRIPRejectionKind = 1
	NTRIPRejectionMountpointNotFound    NTRIPRejectionKind = 2
	NTRIPRejectionDigestRequired        NTRIPRejectionKind = 3
	NTRIPRejectionCasterError           NTRIPRejectionKind = 4
	NTRIPRejectionUnexpectedContentType NTRIPRejectionKind = 5
	NTRIPRejectionHTTPError             NTRIPRejectionKind = 6
	NTRIPRejectionMalformedHandshake    NTRIPRejectionKind = 7
)

type NTRIPSourcetableAuth uint32

const (
	NTRIPSourcetableAuthNone   NTRIPSourcetableAuth = 0
	NTRIPSourcetableAuthBasic  NTRIPSourcetableAuth = 1
	NTRIPSourcetableAuthDigest NTRIPSourcetableAuth = 2
	NTRIPSourcetableAuthOther  NTRIPSourcetableAuth = 3
)

type NTRIPConfig struct {
	Host           string
	Port           uint16
	Mountpoint     string
	Version        NTRIPVersion
	Username       string
	Password       string
	HasCredentials bool
	UserAgent      string
	GGAIntervalS   float64
	HasGGAInterval bool
}

type NTRIPGGAPosition struct {
	LatitudeDeg  float64
	LongitudeDeg float64
	HeightM      float64
	FixQuality   uint8
	Satellites   uint8
	HDOP         float64
}

type NTRIPEventInfo struct {
	Kind               NTRIPEventKind
	Version            NTRIPVersion
	Chunked            bool
	HeaderCount        int
	PayloadLength      int
	SourcetableRecords int
	Rejection          NTRIPRejectionKind
	HTTPStatus         uint16
}

type NTRIPStreamInfo struct {
	Mountpoint         string
	Identifier         string
	Format             string
	FormatDetails      string
	HasCarrier         bool
	Carrier            uint8
	NavSystem          string
	Network            string
	Country            string
	HasLatitudeDeg     bool
	LatitudeDeg        float64
	HasLongitudeDeg    bool
	LongitudeDeg       float64
	HasNMEARequired    bool
	NMEARequired       bool
	HasNetworkSolution bool
	NetworkSolution    bool
	Generator          string
	Compression        string
	Authentication     NTRIPSourcetableAuth
	HasFee             bool
	Fee                bool
	HasBitrate         bool
	Bitrate            uint32
	Misc               string
}

type NTRIPMachine struct{}
type NTRIPBytes struct{}
type NTRIPEvents struct{}
type NTRIPSourcetable struct{}

func NewNTRIPMachine(NTRIPConfig) (*NTRIPMachine, error) { return nil, unavailable() }
func NTRIPRequestBytes(NTRIPConfig) ([]byte, error)      { return nil, unavailable() }
func ParseNTRIPSourcetable([]byte) (*NTRIPSourcetable, error) {
	return nil, unavailable()
}

func (*NTRIPMachine) Close() {}
func (*NTRIPMachine) ConnectionRequest() (*NTRIPBytes, error) {
	return nil, unavailable()
}
func (*NTRIPMachine) Push([]byte) (*NTRIPEvents, error) { return nil, unavailable() }
func (*NTRIPMachine) Finish() (*NTRIPEvents, error)     { return nil, unavailable() }
func (*NTRIPMachine) Reset() error                      { return unavailable() }
func (*NTRIPMachine) State() (NTRIPState, error)        { return 0, unavailable() }
func (*NTRIPMachine) TryGGAMessage(float64, NTRIPGGAPosition, float64) (*NTRIPBytes, bool, error) {
	return nil, false, unavailable()
}

func (*NTRIPBytes) Close()                 {}
func (*NTRIPBytes) Bytes() ([]byte, error) { return nil, unavailable() }
func (*NTRIPEvents) Close()                {}
func (*NTRIPEvents) Count() (int, error)   { return 0, unavailable() }
func (*NTRIPEvents) Event(int) (NTRIPEventInfo, error) {
	return NTRIPEventInfo{}, unavailable()
}
func (*NTRIPEvents) Payload(int) ([]byte, error) { return nil, unavailable() }
func (*NTRIPEvents) Detail(int) ([]byte, error)  { return nil, unavailable() }
func (*NTRIPEvents) Sourcetable(int) (*NTRIPSourcetable, error) {
	return nil, unavailable()
}
func (*NTRIPSourcetable) Close() {}
func (*NTRIPSourcetable) Summary() (int, int, error) {
	return 0, 0, unavailable()
}
func (*NTRIPSourcetable) Streams() ([]NTRIPStreamInfo, error) { return nil, unavailable() }
func (*NTRIPSourcetable) Text() ([]byte, error)               { return nil, unavailable() }
