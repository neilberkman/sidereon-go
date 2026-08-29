package sidereon

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func exactCacheTestIdentity(t *testing.T) ProductIdentity {
	t.Helper()
	identity, err := ResolveProductIdentity(gfzUltraRequest())
	if err != nil {
		t.Fatal(err)
	}
	return identity
}

func exactCacheTestPath(t *testing.T, identity ProductIdentity) string {
	t.Helper()
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("durable native exact-cache publication is supported on macOS and Linux")
	}
	return filepath.Join(t.TempDir(), identity.OfficialFilename)
}

func assertExactCacheFiles(t *testing.T, files *CacheFiles, product, archive, provenance []byte) {
	t.Helper()
	if files == nil {
		t.Fatal("cache files are nil")
	}
	if len(files.EntryID) != 32 {
		t.Fatalf("entry id length = %d, want 32", len(files.EntryID))
	}
	if filepath.Base(files.ProductPath) == "" || filepath.Base(files.ArchivePath) == "" || filepath.Base(files.ProvenancePath) == "" {
		t.Fatalf("cache paths are incomplete: %+v", files)
	}
	if !bytes.Equal(files.ProductBytes, product) || !bytes.Equal(files.ArchiveBytes, archive) || !bytes.Equal(files.ProvenanceBytes, provenance) {
		t.Fatalf("cache bytes = %q, %q, %q", files.ProductBytes, files.ArchiveBytes, files.ProvenanceBytes)
	}
}

func TestExactProductCachePublishReadCleanupAndCorruption(t *testing.T) {
	identity := exactCacheTestIdentity(t)
	stablePath := exactCacheTestPath(t, identity)
	product := []byte("validated exact product")
	archive := []byte("distributor archive")
	provenance := []byte(`{"source":"fixture"}`)

	if files, err := ReadExactProductCache(stablePath, identity, DistributionSourceDirect); err != nil || files != nil {
		t.Fatalf("initial unlocked read = %+v, %v", files, err)
	}

	cache, err := NewExactProductCache(stablePath, identity, DistributionSourceDirect)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := cache.Close(); err != nil {
			t.Error(err)
		}
	}()
	if files, err := cache.Read(); err != nil || files != nil {
		t.Fatalf("initial locked read = %+v, %v", files, err)
	}
	first, err := cache.Publish(product, archive, provenance)
	if err != nil {
		t.Fatal(err)
	}
	assertExactCacheFiles(t, &first, product, archive, provenance)

	secondProduct := []byte("replacement validated exact product")
	second, err := cache.Publish(secondProduct, archive, provenance)
	if err != nil {
		t.Fatal(err)
	}
	if second.EntryID == first.EntryID {
		t.Fatal("replacement publication reused immutable entry id")
	}
	if err := cache.CleanupAbandoned(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Dir(first.ProductPath)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("abandoned entry still exists: %v", err)
	}
	read, err := cache.Read()
	if err != nil {
		t.Fatal(err)
	}
	assertExactCacheFiles(t, read, secondProduct, archive, provenance)

	if err := os.WriteFile(second.ProductPath, []byte("corrupt"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err = cache.Read()
	var formatErr *CacheFormatError
	if !errors.As(err, &formatErr) {
		t.Fatalf("corrupt read error = %T %v, want CacheFormatError", err, err)
	}
}

func TestExactCacheLockTimeout(t *testing.T) {
	identity := exactCacheTestIdentity(t)
	stablePath := exactCacheTestPath(t, identity)
	first, err := NewExactProductCache(stablePath, identity, DistributionSourceDirect)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := first.Close(); err != nil {
			t.Error(err)
		}
	}()

	_, err = OpenExactProductCache(stablePath, identity, DistributionSourceDirect, 2*time.Millisecond)
	var timeoutErr *CacheLockTimeoutError
	if !errors.As(err, &timeoutErr) {
		t.Fatalf("second lock error = %T %v, want CacheLockTimeoutError", err, err)
	}
}

func TestExactCacheSingleFlightOwnerPublishAndHit(t *testing.T) {
	identity := exactCacheTestIdentity(t)
	stablePath := exactCacheTestPath(t, identity)
	options, err := DefaultExactCacheSingleFlightOptions()
	if err != nil {
		t.Fatal(err)
	}
	options.PollInterval = time.Millisecond
	options.HeartbeatInterval = 2 * time.Millisecond
	options.LivenessTimeout = 20 * time.Millisecond
	options.WaitTimeout = 100 * time.Millisecond

	opened, err := OpenExactCacheSingleFlight(stablePath, identity, DistributionSourceDirect, &options)
	if err != nil {
		t.Fatal(err)
	}
	if opened.Files != nil || opened.Owner == nil {
		t.Fatalf("first open = %+v, want owner", opened)
	}
	if err := opened.Owner.Heartbeat(); err != nil {
		t.Fatal(err)
	}
	product := []byte("single-flight validated product")
	archive := []byte("single-flight archive")
	provenance := []byte(`{"single_flight":true}`)
	files, err := opened.Owner.Publish(product, archive, provenance)
	if err != nil {
		t.Fatal(err)
	}
	assertExactCacheFiles(t, &files, product, archive, provenance)
	if err := opened.Owner.Heartbeat(); !errors.Is(err, ErrClosed) {
		t.Fatalf("heartbeat after publish = %v, want ErrClosed", err)
	}

	hit, err := OpenExactCacheSingleFlight(stablePath, identity, DistributionSourceDirect, nil)
	if err != nil {
		t.Fatal(err)
	}
	if hit.Owner != nil {
		t.Fatal("warm single-flight open returned an owner")
	}
	assertExactCacheFiles(t, hit.Files, product, archive, provenance)
}

func TestExactCacheSingleFlightTimeoutAndValidation(t *testing.T) {
	identity := exactCacheTestIdentity(t)
	stablePath := exactCacheTestPath(t, identity)
	ownerOptions := ExactCacheSingleFlightOptions{
		PollInterval:      time.Millisecond,
		HeartbeatInterval: 2 * time.Millisecond,
		LivenessTimeout:   30 * time.Millisecond,
		WaitTimeout:       100 * time.Millisecond,
	}
	opened, err := OpenExactCacheSingleFlight(stablePath, identity, DistributionSourceDirect, &ownerOptions)
	if err != nil {
		t.Fatal(err)
	}
	if opened.Owner == nil {
		t.Fatal("first open did not return owner")
	}
	defer func() {
		if err := opened.Owner.Close(); err != nil {
			t.Error(err)
		}
	}()

	waiterOptions := ownerOptions
	waiterOptions.WaitTimeout = 8 * time.Millisecond
	_, err = OpenExactCacheSingleFlight(stablePath, identity, DistributionSourceDirect, &waiterOptions)
	var timeoutErr *CacheSingleFlightTimeoutError
	if !errors.As(err, &timeoutErr) {
		t.Fatalf("waiter error = %T %v, want CacheSingleFlightTimeoutError", err, err)
	}

	invalid := ownerOptions
	invalid.HeartbeatInterval = invalid.LivenessTimeout
	_, err = OpenExactCacheSingleFlight(stablePath, identity, DistributionSourceDirect, &invalid)
	if err == nil {
		t.Fatal("equal heartbeat and liveness intervals were accepted")
	}
}

func TestExactCacheCloseIsIdempotentAndOperationsFailClosed(t *testing.T) {
	identity := exactCacheTestIdentity(t)
	stablePath := exactCacheTestPath(t, identity)
	cache, err := NewExactProductCache(stablePath, identity, DistributionSourceDirect)
	if err != nil {
		t.Fatal(err)
	}
	if err := cache.Close(); err != nil {
		t.Fatal(err)
	}
	if err := cache.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := cache.Read(); !errors.Is(err, ErrClosed) {
		t.Fatalf("read after close = %v, want ErrClosed", err)
	}
}
