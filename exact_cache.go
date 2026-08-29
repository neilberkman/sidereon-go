package sidereon

import (
	"errors"
	"fmt"
	"time"

	"github.com/neilberkman/sidereon-go/internal/native"
)

// ExactCacheControlDirectory is the engine-defined directory containing
// commit and coordination records beside one stable exact-product path.
const ExactCacheControlDirectory = native.ExactCacheControlDirectory

const defaultExactCacheLockTimeout = 30 * time.Second

// CacheLockTimeoutError reports that a bounded exact-cache writer lock could
// not be acquired. The underlying StatusError is available through Unwrap.
type CacheLockTimeoutError struct{ Err error }

func (e *CacheLockTimeoutError) Error() string {
	return fmt.Sprintf("sidereon: timed out waiting for the exact-cache lock: %v", e.Err)
}

func (e *CacheLockTimeoutError) Unwrap() error { return e.Err }

// CacheSingleFlightTimeoutError reports that a live owner did not publish
// before a single-flight wait deadline.
type CacheSingleFlightTimeoutError struct{ Err error }

func (e *CacheSingleFlightTimeoutError) Error() string {
	return fmt.Sprintf("sidereon: timed out waiting for the exact-cache in-flight owner: %v", e.Err)
}

func (e *CacheSingleFlightTimeoutError) Unwrap() error { return e.Err }

// CacheFormatError reports a malformed, incomplete, corrupt, or mismatched
// immutable cache transaction. Such an error is never treated as a miss.
type CacheFormatError struct{ Err error }

func (e *CacheFormatError) Error() string {
	return fmt.Sprintf("sidereon: invalid exact-cache transaction: %v", e.Err)
}

func (e *CacheFormatError) Unwrap() error { return e.Err }

// CacheFiles contains paths and independent Go-owned copies of all bytes in
// one digest-verified immutable exact-cache transaction.
type CacheFiles struct {
	ProductPath     string
	ArchivePath     string
	ProvenancePath  string
	EntryID         string
	ProductBytes    []byte
	ArchiveBytes    []byte
	ProvenanceBytes []byte
}

// ExactCacheSingleFlightOptions is the bounded timing policy for miss
// coalescing. Every duration must be positive and HeartbeatInterval must be
// shorter than LivenessTimeout.
type ExactCacheSingleFlightOptions struct {
	PollInterval      time.Duration
	HeartbeatInterval time.Duration
	LivenessTimeout   time.Duration
	WaitTimeout       time.Duration
}

// DefaultExactCacheSingleFlightOptions returns the engine defaults: 50 ms
// polling, 5 s heartbeats, 30 s liveness, and a 30 minute total wait.
func DefaultExactCacheSingleFlightOptions() (ExactCacheSingleFlightOptions, error) {
	value, err := native.ExactCacheSingleFlightOptionsInit()
	if err != nil {
		return ExactCacheSingleFlightOptions{}, publicError(err)
	}
	return ExactCacheSingleFlightOptions{
		PollInterval:      time.Duration(value.PollIntervalMS) * time.Millisecond,
		HeartbeatInterval: time.Duration(value.HeartbeatIntervalMS) * time.Millisecond,
		LivenessTimeout:   time.Duration(value.LivenessTimeoutMS) * time.Millisecond,
		WaitTimeout:       time.Duration(value.WaitTimeoutMS) * time.Millisecond,
	}, nil
}

// ExactProductCache owns the engine's bounded cross-process writer lock until
// Close. Methods serialize access and are safe to call from multiple
// goroutines, although Close causes later operations to return ErrClosed.
type ExactProductCache struct {
	native *native.ExactCache
}

// NewExactProductCache opens a lock-owning cache transaction using the engine
// parity default of a 30-second lock timeout.
func NewExactProductCache(stablePath string, identity ProductIdentity, source DistributionSource) (*ExactProductCache, error) {
	return OpenExactProductCache(stablePath, identity, source, defaultExactCacheLockTimeout)
}

// OpenExactProductCache opens a lock-owning cache transaction with a bounded
// non-negative timeout.
func OpenExactProductCache(stablePath string, identity ProductIdentity, source DistributionSource, timeout time.Duration) (*ExactProductCache, error) {
	timeoutMS, err := exactCacheDurationMS(timeout, true, "cache lock timeout")
	if err != nil {
		return nil, err
	}
	nativeIdentity, err := nativeProductIdentity(identity)
	if err != nil {
		return nil, err
	}
	value, err := native.ExactCacheOpen(stablePath, nativeIdentity, native.DistributionSource(source), timeoutMS)
	if err != nil {
		return nil, exactCacheLockError(err)
	}
	return &ExactProductCache{native: value}, nil
}

// Close releases the cross-process writer lock. It is idempotent.
func (cache *ExactProductCache) Close() error {
	if cache == nil || cache.native == nil {
		return nil
	}
	return publicError(cache.native.Close())
}

// Read returns the current digest-verified immutable entry. A missing commit
// returns (nil, nil); corrupt or mismatched data returns CacheFormatError.
func (cache *ExactProductCache) Read() (*CacheFiles, error) {
	if cache == nil || cache.native == nil {
		return nil, ErrClosed
	}
	entry, hit, err := cache.native.Read()
	if err != nil {
		translated := publicError(err)
		if errors.Is(translated, ErrClosed) {
			return nil, translated
		}
		return nil, &CacheFormatError{Err: translated}
	}
	if !hit {
		return nil, nil
	}
	return copyAndCloseExactCacheEntry(entry)
}

// Publish atomically commits product, distributor archive, and provenance
// bytes. Callers must validate product semantics before publication.
func (cache *ExactProductCache) Publish(product, archive, provenance []byte) (CacheFiles, error) {
	if cache == nil || cache.native == nil {
		return CacheFiles{}, ErrClosed
	}
	entry, err := cache.native.Publish(product, archive, provenance)
	if err != nil {
		return CacheFiles{}, publicError(err)
	}
	files, err := copyAndCloseExactCacheEntry(entry)
	if err != nil {
		return CacheFiles{}, err
	}
	return *files, nil
}

// CleanupAbandoned removes unreferenced immutable transactions and temporary
// commit artifacts while this cache holds the writer lock.
func (cache *ExactProductCache) CleanupAbandoned() error {
	if cache == nil || cache.native == nil {
		return ErrClosed
	}
	return publicError(cache.native.Cleanup())
}

// ExactCacheOpenResult contains exactly one non-nil field: Files for a
// verified hit, or Owner for exclusive ownership of a cache miss.
type ExactCacheOpenResult struct {
	Files *CacheFiles
	Owner *ExactCacheOwner
}

// ExactCacheOwner is the exclusive right to fetch, validate, and publish one
// single-flight miss. Close abandons an unpublished acquisition.
type ExactCacheOwner struct {
	native *native.ExactCacheOwner
}

// OpenExactCacheSingleFlight returns a verified hit or ownership of a miss.
// A nil options pointer selects engine defaults.
func OpenExactCacheSingleFlight(stablePath string, identity ProductIdentity, source DistributionSource, options *ExactCacheSingleFlightOptions) (ExactCacheOpenResult, error) {
	var nativeOptions *native.ExactCacheSingleFlightOptions
	if options != nil {
		converted, err := nativeExactCacheSingleFlightOptions(*options)
		if err != nil {
			return ExactCacheOpenResult{}, err
		}
		nativeOptions = &converted
	}
	nativeIdentity, err := nativeProductIdentity(identity)
	if err != nil {
		return ExactCacheOpenResult{}, err
	}
	entry, owner, err := native.ExactCacheOpenSingleFlight(stablePath, nativeIdentity, native.DistributionSource(source), nativeOptions)
	if err != nil {
		return ExactCacheOpenResult{}, exactCacheSingleFlightError(err)
	}
	if entry != nil {
		files, copyErr := copyAndCloseExactCacheEntry(entry)
		if copyErr != nil {
			return ExactCacheOpenResult{}, copyErr
		}
		return ExactCacheOpenResult{Files: files}, nil
	}
	if owner == nil {
		return ExactCacheOpenResult{}, errors.New("sidereon: exact-cache open returned neither a hit nor an owner")
	}
	return ExactCacheOpenResult{Owner: &ExactCacheOwner{native: owner}}, nil
}

// ReadExactProductCache reads without acquiring the writer lock. A missing
// commit returns (nil, nil); invalid committed data returns CacheFormatError.
func ReadExactProductCache(stablePath string, identity ProductIdentity, source DistributionSource) (*CacheFiles, error) {
	nativeIdentity, err := nativeProductIdentity(identity)
	if err != nil {
		return nil, err
	}
	entry, hit, err := native.ExactCacheReadUnlocked(stablePath, nativeIdentity, native.DistributionSource(source))
	if err != nil {
		return nil, &CacheFormatError{Err: publicError(err)}
	}
	if !hit {
		return nil, nil
	}
	return copyAndCloseExactCacheEntry(entry)
}

// Heartbeat refreshes this owner's liveness immediately. The engine also
// maintains the configured automatic heartbeat while the owner is open.
func (owner *ExactCacheOwner) Heartbeat() error {
	if owner == nil || owner.native == nil {
		return ErrClosed
	}
	return publicError(owner.native.Heartbeat())
}

// Publish atomically commits validated bytes and consumes this owner. The
// owner is closed after a native publication attempt, whether it succeeds or
// fails.
func (owner *ExactCacheOwner) Publish(product, archive, provenance []byte) (CacheFiles, error) {
	if owner == nil || owner.native == nil {
		return CacheFiles{}, ErrClosed
	}
	entry, err := owner.native.Publish(product, archive, provenance)
	if err != nil {
		return CacheFiles{}, publicError(err)
	}
	files, err := copyAndCloseExactCacheEntry(entry)
	if err != nil {
		return CacheFiles{}, err
	}
	return *files, nil
}

// Close abandons an unpublished miss and releases its in-flight marker. It is
// idempotent.
func (owner *ExactCacheOwner) Close() error {
	if owner == nil || owner.native == nil {
		return nil
	}
	return publicError(owner.native.Close())
}

func copyAndCloseExactCacheEntry(entry *native.ExactCacheEntry) (*CacheFiles, error) {
	if entry == nil {
		return nil, errNilNativeHandle
	}
	value, copyErr := entry.Files()
	closeErr := entry.Close()
	if err := joinPublicErrors(copyErr, closeErr); err != nil {
		return nil, err
	}
	return &CacheFiles{
		ProductPath:     value.ProductPath,
		ArchivePath:     value.ArchivePath,
		ProvenancePath:  value.ProvenancePath,
		EntryID:         value.EntryID,
		ProductBytes:    value.Product,
		ArchiveBytes:    value.Archive,
		ProvenanceBytes: value.Provenance,
	}, nil
}

func nativeProductIdentity(value ProductIdentity) (native.ProductIdentity, error) {
	year, month, day := value.Date.UTC().Date()
	if year < -1<<31 || year > 1<<31-1 {
		return native.ProductIdentity{}, fmt.Errorf("sidereon: product identity year %d does not fit in the C ABI", year)
	}
	return native.ProductIdentity{
		Family:                   native.ProductFamily(value.Family),
		AnalysisCenter:           value.AnalysisCenter,
		Publisher:                native.ProductPublisher(value.Publisher),
		SolutionClass:            native.SolutionClass(value.SolutionClass),
		Campaign:                 native.ProductCampaign(value.Campaign),
		FilenameVersion:          value.FilenameVersion,
		Year:                     int32(year),
		Month:                    uint8(month),
		Day:                      uint8(day),
		HasIssue:                 value.HasIssue,
		Issue:                    value.Issue,
		Span:                     value.Span,
		Sample:                   value.Sample,
		OfficialFilename:         value.OfficialFilename,
		Format:                   native.ProductFormat(value.Format),
		HasFormatVersion:         value.HasFormatVersion,
		FormatVersion:            value.FormatVersion,
		HasPredictionHorizonDays: value.HasPredictionHorizonDays,
		PredictionHorizonDays:    value.PredictionHorizonDays,
	}, nil
}

func nativeExactCacheSingleFlightOptions(value ExactCacheSingleFlightOptions) (native.ExactCacheSingleFlightOptions, error) {
	poll, err := exactCacheDurationMS(value.PollInterval, false, "single-flight poll interval")
	if err != nil {
		return native.ExactCacheSingleFlightOptions{}, err
	}
	heartbeat, err := exactCacheDurationMS(value.HeartbeatInterval, false, "single-flight heartbeat interval")
	if err != nil {
		return native.ExactCacheSingleFlightOptions{}, err
	}
	liveness, err := exactCacheDurationMS(value.LivenessTimeout, false, "single-flight liveness timeout")
	if err != nil {
		return native.ExactCacheSingleFlightOptions{}, err
	}
	wait, err := exactCacheDurationMS(value.WaitTimeout, false, "single-flight wait timeout")
	if err != nil {
		return native.ExactCacheSingleFlightOptions{}, err
	}
	if heartbeat >= liveness {
		return native.ExactCacheSingleFlightOptions{}, errors.New("sidereon: single-flight heartbeat interval must be shorter than liveness timeout")
	}
	return native.ExactCacheSingleFlightOptions{
		PollIntervalMS:      poll,
		HeartbeatIntervalMS: heartbeat,
		LivenessTimeoutMS:   liveness,
		WaitTimeoutMS:       wait,
	}, nil
}

func exactCacheDurationMS(value time.Duration, allowZero bool, name string) (uint64, error) {
	if value < 0 || (!allowZero && value == 0) {
		qualifier := "positive"
		if allowZero {
			qualifier = "non-negative"
		}
		return 0, fmt.Errorf("sidereon: %s must be %s", name, qualifier)
	}
	milliseconds := uint64(value / time.Millisecond)
	if value%time.Millisecond != 0 {
		milliseconds++
	}
	return milliseconds, nil
}

func exactCacheLockError(err error) error {
	translated := publicError(err)
	var status *StatusError
	if errors.As(translated, &status) && status.Code == StatusTimeout {
		return &CacheLockTimeoutError{Err: translated}
	}
	return translated
}

func exactCacheSingleFlightError(err error) error {
	translated := publicError(err)
	var status *StatusError
	if errors.As(translated, &status) && status.Code == StatusTimeout {
		return &CacheSingleFlightTimeoutError{Err: translated}
	}
	return translated
}
