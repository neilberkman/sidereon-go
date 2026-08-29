//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

const ExactCacheControlDirectory = ".sidereon-cache-v3"

type ExactCacheSingleFlightOptions struct {
	PollIntervalMS      uint64
	HeartbeatIntervalMS uint64
	LivenessTimeoutMS   uint64
	WaitTimeoutMS       uint64
}

type ExactCacheFiles struct {
	ProductPath    string
	ArchivePath    string
	ProvenancePath string
	EntryID        string
	Product        []byte
	Archive        []byte
	Provenance     []byte
}

type ExactCache struct{}
type ExactCacheEntry struct{}
type ExactCacheOwner struct{}

func ExactCacheSingleFlightOptionsInit() (ExactCacheSingleFlightOptions, error) {
	return ExactCacheSingleFlightOptions{}, unavailable()
}

func ExactCacheOpen(string, ProductIdentity, DistributionSource, uint64) (*ExactCache, error) {
	return nil, unavailable()
}

func ExactCacheOpenSingleFlight(string, ProductIdentity, DistributionSource, *ExactCacheSingleFlightOptions) (*ExactCacheEntry, *ExactCacheOwner, error) {
	return nil, nil, unavailable()
}

func ExactCacheReadUnlocked(string, ProductIdentity, DistributionSource) (*ExactCacheEntry, bool, error) {
	return nil, false, unavailable()
}

func (*ExactCache) Close() error                          { return nil }
func (*ExactCache) Cleanup() error                        { return unavailable() }
func (*ExactCache) Read() (*ExactCacheEntry, bool, error) { return nil, false, unavailable() }
func (*ExactCache) Publish([]byte, []byte, []byte) (*ExactCacheEntry, error) {
	return nil, unavailable()
}

func (*ExactCacheEntry) Close() error { return nil }
func (*ExactCacheEntry) Files() (ExactCacheFiles, error) {
	return ExactCacheFiles{}, unavailable()
}

func (*ExactCacheOwner) Close() error     { return nil }
func (*ExactCacheOwner) Heartbeat() error { return unavailable() }
func (*ExactCacheOwner) Publish([]byte, []byte, []byte) (*ExactCacheEntry, error) {
	return nil, unavailable()
}
