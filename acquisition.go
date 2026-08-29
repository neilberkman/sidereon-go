package sidereon

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	defaultMaxArchiveBytes = 64 << 20
	defaultMaxProductBytes = 500 << 20
	maxHTTPRetries         = 32
)

// AcquireRequest combines an exact catalog query with an HTTP-supported
// distribution source.
type AcquireRequest struct {
	CatalogRequest
	Source DistributionSource
}

// AcquisitionProvenance records only Go-owned, secret-free acquisition facts.
// RequestedIdentity is the exact identity derived from the caller's query
// before any publication-listing selection. ResolvedIdentity is the exact
// C-catalog identity used to obtain the returned bytes. They are equal for
// Acquire and may differ for AcquireLatest.
type AcquisitionProvenance struct {
	RequestedIdentity ProductIdentity
	ResolvedIdentity  ProductIdentity
	Source            DistributionSource
	// OfficialFilename is the provider's canonical product filename.
	OfficialFilename string
	// OriginalURL is the URL requested from the provider.
	OriginalURL string
	// FinalURL is the final URL after redirects.
	FinalURL string
	// RetrievedAt is the UTC time at which the product bytes were retrieved.
	RetrievedAt time.Time
	ByteLength  int
	// SHA256 is the lowercase SHA-256 digest of the returned product bytes.
	SHA256             string
	ArchiveCompression ArchiveCompression
	ArchiveByteLength  int
	// ArchiveSHA256 is the lowercase SHA-256 digest of the compressed archive bytes.
	ArchiveSHA256 string
}

// AcquiredProduct contains decompressed product bytes and their provenance.
// Bytes are independent of any response body and may be retained by callers.
type AcquiredProduct struct {
	Bytes      []byte
	Provenance AcquisitionProvenance
}

// HTTPAcquirer owns HTTP transport policy. The supplied Client is cloned for
// each operation so its redirect policy cannot bypass the catalog allowlist.
type HTTPAcquirer struct {
	// Client is the HTTP client to clone for requests; nil uses http.DefaultClient.
	Client *http.Client
	// MaxArchiveBytes is the maximum archive size in bytes.
	MaxArchiveBytes int64
	// MaxProductBytes is the maximum extracted product size in bytes.
	MaxProductBytes int64
	// Retries is capped at maxHTTPRetries (32) to keep exponential backoff
	// configuration bounded and predictable.
	Retries int
	// Backoff is the base delay between retry attempts.
	Backoff time.Duration
	// Now supplies the retry/retrieval clock; nil defaults to time.Now.
	Now func() time.Time
}

// NewHTTPAcquirer creates an acquirer with bounded archive/product reads and
// no retry delay. A nil client uses http.DefaultClient's transport settings.
func NewHTTPAcquirer(client *http.Client) *HTTPAcquirer {
	return &HTTPAcquirer{Client: client, MaxArchiveBytes: defaultMaxArchiveBytes, MaxProductBytes: defaultMaxProductBytes, Retries: 1, Now: time.Now}
}

// HTTPStatusError reports a non-success HTTP status.
type HTTPStatusError struct {
	Status int
	// URL is the URL whose HTTP status caused the error.
	URL string
}

// Error returns the HTTP status code and rejected URL.
func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("sidereon: HTTP status %d for %s", e.Status, e.URL)
}

// SizeLimitError reports an archive or decompressed product exceeding a bound.
type SizeLimitError struct {
	// Kind identifies the resource whose byte limit was exceeded.
	Kind string
	// Limit is the byte limit that was exceeded.
	Limit int64
}

// Error returns the limited resource kind and maximum byte count.
func (e *SizeLimitError) Error() string {
	return fmt.Sprintf("sidereon: %s exceeds %d bytes", e.Kind, e.Limit)
}

// RedirectPolicyError reports a redirect to a host not returned by the C
// catalog for this operation.
type RedirectPolicyError struct{ URL string }

// Error returns the redirect URL rejected by the catalog allowlist.
func (e *RedirectPolicyError) Error() string {
	return fmt.Sprintf("sidereon: redirect target is not catalog-allowlisted: %s", e.URL)
}

// UnsupportedHTTPSourceError reports a C-approved source that this net/http
// acquirer cannot serve, such as a local-file or non-HTTP endpoint.
type UnsupportedHTTPSourceError struct {
	Source DistributionSource
}

// Error returns the unsupported distribution source.
func (e *UnsupportedHTTPSourceError) Error() string {
	return fmt.Sprintf("sidereon: source %s is not an HTTP source", e.Source)
}

// ErrProductNotPublished is returned when all C-approved listing URLs are
// readable but contain no matching published product.
var ErrProductNotPublished = errors.New("sidereon: product is not published")

func (a *HTTPAcquirer) normalized() (*HTTPAcquirer, error) {
	if a == nil {
		a = NewHTTPAcquirer(nil)
	}
	copy := *a
	if copy.Client == nil {
		copy.Client = http.DefaultClient
	}
	if copy.MaxArchiveBytes == 0 {
		copy.MaxArchiveBytes = defaultMaxArchiveBytes
	}
	if copy.MaxProductBytes == 0 {
		copy.MaxProductBytes = defaultMaxProductBytes
	}
	if copy.MaxArchiveBytes < 0 || copy.MaxProductBytes < 0 {
		return nil, errors.New("sidereon: byte bounds must not be negative")
	}
	if copy.Retries == 0 {
		copy.Retries = 1
	}
	if copy.Retries < 0 {
		return nil, errors.New("sidereon: retries must not be negative")
	}
	if copy.Retries > maxHTTPRetries {
		return nil, fmt.Errorf("sidereon: retries must not exceed %d", maxHTTPRetries)
	}
	if copy.Backoff < 0 {
		return nil, errors.New("sidereon: backoff must not be negative")
	}
	if copy.Now == nil {
		copy.Now = time.Now
	}
	return &copy, nil
}

// Acquire resolves the exact identity and URL through C before making any
// request, then downloads, bounds, decompresses, and returns Go-owned bytes.
func (a *HTTPAcquirer) Acquire(ctx context.Context, request AcquireRequest) (AcquiredProduct, error) {
	if ctx == nil {
		return AcquiredProduct{}, errors.New("sidereon: nil context")
	}
	acquirer, err := a.normalized()
	if err != nil {
		return AcquiredProduct{}, err
	}
	identity, err := ResolveProductIdentity(request.CatalogRequest)
	if err != nil {
		return AcquiredProduct{}, err
	}
	location, err := DistributionLocationFor(request.CatalogRequest, request.Source)
	if err != nil {
		return AcquiredProduct{}, err
	}
	if request.Source != DistributionSourceDirect && request.Source != DistributionSourceNASACDDIS {
		return AcquiredProduct{}, &UnsupportedHTTPSourceError{Source: request.Source}
	}
	return acquirer.acquireAt(ctx, request.Source, identity, identity, location, nil)
}

// AcquireLatest obtains at most the bounded listing URL set returned by C,
// parses each body through C, and then downloads only the selected C-catalog
// identity. It never polls or guesses a neighboring date.
func (a *HTTPAcquirer) AcquireLatest(ctx context.Context, request CatalogRequest, source DistributionSource) (AcquiredProduct, error) {
	if ctx == nil {
		return AcquiredProduct{}, errors.New("sidereon: nil context")
	}
	acquirer, err := a.normalized()
	if err != nil {
		return AcquiredProduct{}, err
	}
	urls, err := PublicationListingURLs(request)
	if err != nil {
		return AcquiredProduct{}, err
	}
	if source != DistributionSourceDirect && source != DistributionSourceNASACDDIS {
		return AcquiredProduct{}, &UnsupportedHTTPSourceError{Source: source}
	}
	requested, err := ResolveProductIdentity(request)
	if err != nil {
		return AcquiredProduct{}, err
	}
	if len(urls) > 2 {
		return AcquiredProduct{}, errors.New("sidereon: C returned more than two listing URLs")
	}
	allowed := make(map[string]bool, len(urls))
	for _, value := range urls {
		parsed, parseErr := catalogHTTPURL(value)
		if parseErr != nil {
			return AcquiredProduct{}, parseErr
		}
		allowed[parsed.Scheme+"://"+parsed.Host] = true
	}
	var lastErr error
	for _, listingURL := range urls {
		body, _, fetchErr := acquirer.fetch(ctx, listingURL, allowed, 1)
		if fetchErr != nil {
			lastErr = fetchErr
			continue
		}
		published, parseErr := NewestPublishedProduct(request.Center, request.Family, body)
		if parseErr != nil {
			return AcquiredProduct{}, parseErr
		}
		if published == nil {
			continue
		}
		candidate := request
		candidate.Date = published.Date
		candidate.Issue = published.Issue
		identity, identityErr := ResolveProductIdentity(candidate)
		if identityErr != nil {
			return AcquiredProduct{}, identityErr
		}
		if identity.OfficialFilename != published.Filename {
			return AcquiredProduct{}, fmt.Errorf("sidereon: listing filename %q is not the C-catalog filename %q", published.Filename, identity.OfficialFilename)
		}
		location, locationErr := DistributionLocationFor(candidate, source)
		if locationErr != nil {
			return AcquiredProduct{}, locationErr
		}
		if parsed, parseErr := catalogHTTPURL(location.OriginalURL); parseErr != nil {
			return AcquiredProduct{}, parseErr
		} else {
			allowed[parsed.Scheme+"://"+parsed.Host] = true
		}
		return acquirer.acquireAt(ctx, source, requested, identity, location, allowed)
	}
	if lastErr != nil {
		return AcquiredProduct{}, lastErr
	}
	return AcquiredProduct{}, ErrProductNotPublished
}

func (a *HTTPAcquirer) acquireAt(ctx context.Context, source DistributionSource, requested, resolved ProductIdentity, location DistributionLocation, allowed map[string]bool) (AcquiredProduct, error) {
	if location.OriginalURL == "" {
		return AcquiredProduct{}, &UnsupportedHTTPSourceError{Source: source}
	}
	parsed, err := catalogHTTPURL(location.OriginalURL)
	if err != nil {
		return AcquiredProduct{}, err
	}
	if allowed == nil {
		allowed = map[string]bool{parsed.Scheme + "://" + parsed.Host: true}
	}
	var body []byte
	finalURL := parsed.String()
	body, response, err := a.fetch(ctx, location.OriginalURL, allowed, a.Retries)
	if err != nil {
		return AcquiredProduct{}, err
	}
	if response != nil && response.Request != nil {
		finalURL = response.Request.URL.String()
	}
	content, err := decompressArchive(body, location.Compression, a.MaxProductBytes)
	if err != nil {
		return AcquiredProduct{}, err
	}
	now := a.Now().UTC()
	archiveDigest := sha256.Sum256(body)
	productDigest := sha256.Sum256(content)
	return AcquiredProduct{Bytes: append([]byte(nil), content...), Provenance: AcquisitionProvenance{
		RequestedIdentity:  requested,
		ResolvedIdentity:   resolved,
		Source:             source,
		OfficialFilename:   resolved.OfficialFilename,
		OriginalURL:        location.OriginalURL,
		FinalURL:           finalURL,
		RetrievedAt:        now,
		ByteLength:         len(content),
		SHA256:             hex.EncodeToString(productDigest[:]),
		ArchiveCompression: location.Compression,
		ArchiveByteLength:  len(body),
		ArchiveSHA256:      hex.EncodeToString(archiveDigest[:]),
	}}, nil
}

func (a *HTTPAcquirer) fetch(ctx context.Context, rawURL string, allowed map[string]bool, retries int) ([]byte, *http.Response, error) {
	parsed, err := catalogHTTPURL(rawURL)
	if err != nil {
		return nil, nil, err
	}
	client := *a.Client
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 8 {
			return &RedirectPolicyError{URL: req.URL.String()}
		}
		host := req.URL.Scheme + "://" + req.URL.Host
		if !allowed[host] {
			return &RedirectPolicyError{URL: req.URL.String()}
		}
		return nil
	}
	if retries <= 0 {
		retries = 1
	}
	for attempt := 0; attempt < retries; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
		req, requestErr := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
		if requestErr != nil {
			return nil, nil, requestErr
		}
		response, requestErr := client.Do(req)
		if requestErr != nil {
			closeErr := closeResponseBody(response)
			if ctx.Err() != nil {
				return nil, nil, errors.Join(ctx.Err(), closeErr)
			}
			if attempt+1 < retries {
				if closeErr != nil {
					return nil, nil, errors.Join(requestErr, closeErr)
				}
				if waitErr := waitContext(ctx, a.Backoff, attempt); waitErr != nil {
					return nil, nil, errors.Join(requestErr, waitErr)
				}
				continue
			}
			return nil, nil, errors.Join(requestErr, closeErr)
		}
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			statusURL := parsed.String()
			if response.Request != nil {
				statusURL = response.Request.URL.String()
			}
			statusErr := &HTTPStatusError{Status: response.StatusCode, URL: statusURL}
			if closeErr := closeResponseBody(response); closeErr != nil {
				return nil, nil, errors.Join(statusErr, closeErr)
			}
			if response.StatusCode >= 500 && attempt+1 < retries {
				if waitErr := waitContext(ctx, a.Backoff, attempt); waitErr != nil {
					return nil, nil, waitErr
				}
				continue
			}
			return nil, nil, statusErr
		}
		body, readErr := readBounded(response.Body, a.MaxArchiveBytes, "archive")
		if closeErr := closeResponseBody(response); closeErr != nil {
			readErr = errors.Join(readErr, closeErr)
		}
		if readErr != nil {
			return nil, nil, readErr
		}
		return body, response, nil
	}
	return nil, nil, errors.New("sidereon: HTTP retry loop exhausted")
}

func closeResponseBody(response *http.Response) error {
	if response == nil || response.Body == nil {
		return nil
	}
	return response.Body.Close()
}

func catalogHTTPURL(raw string) (*url.URL, error) {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "http" && parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil {
		return nil, fmt.Errorf("sidereon: catalog returned a non-HTTP URL %q", raw)
	}
	return parsed, nil
}

func waitContext(ctx context.Context, delay time.Duration, attempt int) error {
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(retryDelay(delay, attempt))
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

const maxDuration = time.Duration(1<<63 - 1)

func retryDelay(delay time.Duration, attempt int) time.Duration {
	if delay <= 0 || attempt <= 0 {
		return delay
	}
	if attempt >= 63 {
		return maxDuration
	}
	multiplier := time.Duration(1) << uint(attempt)
	if delay > maxDuration/multiplier {
		return maxDuration
	}
	return delay * multiplier
}

func readBounded(reader io.Reader, limit int64, kind string) ([]byte, error) {
	if reader == nil {
		return nil, errors.New("sidereon: response has no body")
	}
	if limit < 0 {
		return nil, errors.New("sidereon: byte bound must not be negative")
	}
	data := make([]byte, 0, minInt64(limit, 32<<10))
	buffer := make([]byte, 32<<10)
	for int64(len(data)) < limit {
		readBuffer := buffer
		remaining := limit - int64(len(data))
		if remaining < int64(len(readBuffer)) {
			readBuffer = readBuffer[:int(remaining)]
		}
		n, err := reader.Read(readBuffer)
		if n < 0 || n > len(readBuffer) {
			return nil, io.ErrShortBuffer
		}
		if n > 0 {
			data = append(data, readBuffer[:n]...)
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				return data, nil
			}
			return data, err
		}
		if n == 0 {
			return nil, io.ErrNoProgress
		}
	}
	var probe [1]byte
	n, probeErr := reader.Read(probe[:])
	if n > 0 {
		return nil, &SizeLimitError{Kind: kind, Limit: limit}
	}
	if n < 0 || n > len(probe) {
		return nil, io.ErrShortBuffer
	}
	if probeErr != nil && !errors.Is(probeErr, io.EOF) {
		return nil, probeErr
	}
	if n == 0 && probeErr == nil {
		return nil, io.ErrNoProgress
	}
	return data, nil
}

func decompressArchive(archive []byte, compression ArchiveCompression, limit int64) (result []byte, err error) {
	if limit < 0 {
		return nil, errors.New("sidereon: product byte bound must not be negative")
	}
	switch compression {
	case ArchiveCompressionNone:
		if int64(len(archive)) > limit {
			return nil, &SizeLimitError{Kind: "product", Limit: limit}
		}
		return append([]byte(nil), archive...), nil
	case ArchiveCompressionGZIP:
		reader, err := gzip.NewReader(bytes.NewReader(archive))
		if err != nil {
			return nil, fmt.Errorf("sidereon: gzip decompression: %w", err)
		}
		result, err = readBounded(reader, limit, "decompressed product")
		if closeErr := reader.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("sidereon: gzip close: %w", closeErr))
		}
		return result, err
	case ArchiveCompressionUnixCompress:
		return decompressUnixCompress(archive, limit)
	default:
		return nil, fmt.Errorf("sidereon: unsupported archive compression %d", compression)
	}
}

// decompressUnixCompress decodes the classic .Z LZW stream. The decoder is
// deliberately here, beside the Go-owned HTTP/archive transport, rather than
// in the catalog ABI. Output is checked before every append so a corrupt or
// adversarial archive cannot grow past the caller's product bound.
func decompressUnixCompress(archive []byte, limit int64) ([]byte, error) {
	if limit < 0 {
		return nil, errors.New("sidereon: product byte bound must not be negative")
	}
	if len(archive) < 3 || archive[0] != 0x1f || archive[1] != 0x9d {
		return nil, errors.New("sidereon: invalid Unix-compress header")
	}
	flags := archive[2]
	maxBits := int(flags & 0x1f)
	if maxBits < 9 || maxBits > 16 {
		return nil, fmt.Errorf("sidereon: invalid Unix-compress code width %d", maxBits)
	}
	blockMode := flags&0x80 != 0
	maxEntries := 1 << maxBits
	const initialBits = 9
	const clearCode = 256

	prefix := make([]int, maxEntries)
	suffix := make([]byte, maxEntries)
	stack := make([]byte, maxEntries)
	for i := range prefix {
		prefix[i] = -1
	}

	bitOffset := 24
	nBits := initialBits
	maxCode := (1 << nBits) - 1
	freeEntry := 256
	alignmentOrigin := bitOffset
	if blockMode {
		freeEntry++
	}
	readCode := func() (int, bool) {
		if bitOffset+nBits > len(archive)*8 {
			return 0, false
		}
		byteOffset := bitOffset / 8
		shift := uint(bitOffset % 8)
		value := uint64(archive[byteOffset])
		if byteOffset+1 < len(archive) {
			value |= uint64(archive[byteOffset+1]) << 8
		}
		if byteOffset+2 < len(archive) {
			value |= uint64(archive[byteOffset+2]) << 16
		}
		code := int((value >> shift) & uint64((1<<nBits)-1))
		bitOffset += nBits
		return code, true
	}
	appendBytes := func(output *[]byte, data []byte) error {
		if int64(len(data)) > limit-int64(len(*output)) {
			return &SizeLimitError{Kind: "decompressed product", Limit: limit}
		}
		*output = append(*output, data...)
		return nil
	}

	oldCode, ok := readCode()
	if !ok || oldCode > 255 {
		return nil, errors.New("sidereon: Unix-compress stream has no initial literal")
	}
	finChar := byte(oldCode)
	output := make([]byte, 0, minInt64(limit, 32<<10))
	if err := appendBytes(&output, []byte{finChar}); err != nil {
		return nil, err
	}

	for {
		if freeEntry > maxCode && nBits < maxBits {
			bitOffset = alignUnixCompressBlock(bitOffset, alignmentOrigin, nBits)
			nBits++
			maxCode = (1 << nBits) - 1
			alignmentOrigin = bitOffset
		}
		code, ok := readCode()
		if !ok {
			break
		}
		if blockMode && code == clearCode {
			clearWidth := nBits
			for i := clearCode + 1; i < len(prefix); i++ {
				prefix[i] = -1
			}
			bitOffset = alignUnixCompressBlock(bitOffset, alignmentOrigin, clearWidth)
			nBits = initialBits
			maxCode = (1 << nBits) - 1
			freeEntry = clearCode
			alignmentOrigin = bitOffset
			continue
		}

		inputCode := code
		stackLen := 0
		if code >= freeEntry {
			if code != freeEntry || stackLen == len(stack) {
				return nil, errors.New("sidereon: invalid Unix-compress code")
			}
			stack[stackLen] = finChar
			stackLen++
			code = oldCode
		}
		for code >= 256 {
			if code >= freeEntry || code >= len(prefix) || stackLen == len(stack) {
				return nil, errors.New("sidereon: invalid Unix-compress dictionary reference")
			}
			stack[stackLen] = suffix[code]
			stackLen++
			code = prefix[code]
			if code < 0 {
				return nil, errors.New("sidereon: invalid Unix-compress dictionary chain")
			}
		}
		if code > 255 || stackLen == len(stack) {
			return nil, errors.New("sidereon: invalid Unix-compress literal")
		}
		finChar = byte(code)
		stack[stackLen] = finChar
		stackLen++
		for left, right := 0, stackLen-1; left < right; left, right = left+1, right-1 {
			stack[left], stack[right] = stack[right], stack[left]
		}
		if err := appendBytes(&output, stack[:stackLen]); err != nil {
			return nil, err
		}
		if freeEntry < maxEntries {
			prefix[freeEntry] = oldCode
			suffix[freeEntry] = finChar
			freeEntry++
		}
		oldCode = inputCode
	}
	return output, nil
}

// alignUnixCompressBlock advances a compressed code-stream offset to the next
// complete block of eight codes at the current width. alignmentOrigin is 24
// for the initial stream after the three-byte header and becomes the aligned
// start after each width transition or CLEAR reset. Byte alignment is not
// sufficient when a CLEAR code lands at an arbitrary bit position.
func alignUnixCompressBlock(bitOffset, alignmentOrigin, nBits int) int {
	blockBits := 8 * nBits
	blockBitsUsed := bitOffset - alignmentOrigin
	if remainder := blockBitsUsed % blockBits; remainder != 0 {
		bitOffset += blockBits - remainder
	}
	return bitOffset
}

func minInt64(value int64, maximum int) int {
	if value < int64(maximum) {
		return int(value)
	}
	return maximum
}
