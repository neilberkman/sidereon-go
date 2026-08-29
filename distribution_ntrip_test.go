package sidereon

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func gfzUltraRequest() CatalogRequest {
	return CatalogRequest{Center: "gfz_ult", Family: ProductFamilySP3, Date: time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC), Issue: "0000"}
}

func gzipBytes(t *testing.T, data []byte) []byte {
	t.Helper()
	var out bytes.Buffer
	writer := gzip.NewWriter(&out)
	if _, err := writer.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func routedClient(server *httptest.Server, seen *[]string) *http.Client {
	return &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		*seen = append(*seen, request.URL.String())
		copy := request.Clone(request.Context())
		path := request.URL.EscapedPath()
		if path == "" {
			path = "/"
		}
		parsed, err := url.Parse(server.URL + path)
		if err != nil {
			return nil, err
		}
		copy.URL = parsed
		copy.Host = copy.URL.Host
		return http.DefaultTransport.RoundTrip(copy)
	})}
}

func writeTestResponse(t *testing.T, response http.ResponseWriter, data []byte) {
	t.Helper()
	if _, err := response.Write(data); err != nil {
		t.Errorf("write test response: %v", err)
	}
}

func cleanupTestResource(t *testing.T, close func() error) {
	t.Helper()
	if err := close(); err != nil {
		t.Errorf("close test resource: %v", err)
	}
}

func nilContextForTest() context.Context { return nil }

func TestNTRIPDiscriminantsMatchCContract(t *testing.T) {
	tests := []struct {
		name string
		got  uint32
		want uint32
	}{
		{"version rev1", uint32(NTRIPVersionRev1), 1},
		{"version rev2", uint32(NTRIPVersionRev2), 2},
		{"state idle", uint32(NTRIPStateIdle), 0},
		{"state awaiting status", uint32(NTRIPStateAwaitingStatus), 1},
		{"state awaiting headers", uint32(NTRIPStateAwaitingHeaders), 2},
		{"state streaming", uint32(NTRIPStateStreaming), 3},
		{"state sourcetable", uint32(NTRIPStateSourcetable), 4},
		{"state closed", uint32(NTRIPStateClosed), 5},
		{"event connected", uint32(NTRIPEventConnected), 0},
		{"event payload", uint32(NTRIPEventPayload), 1},
		{"event sourcetable", uint32(NTRIPEventSourcetable), 2},
		{"event rejected", uint32(NTRIPEventRejected), 3},
		{"event stream corrupted", uint32(NTRIPEventStreamCorrupted), 4},
		{"event stream ended", uint32(NTRIPEventStreamEnded), 5},
		{"rejection none", uint32(NTRIPRejectionNone), 0},
		{"rejection unauthorized", uint32(NTRIPRejectionUnauthorized), 1},
		{"rejection mountpoint not found", uint32(NTRIPRejectionMountpointNotFound), 2},
		{"rejection digest required", uint32(NTRIPRejectionDigestRequired), 3},
		{"rejection caster error", uint32(NTRIPRejectionCasterError), 4},
		{"rejection content type", uint32(NTRIPRejectionUnexpectedContentType), 5},
		{"rejection HTTP error", uint32(NTRIPRejectionHTTPError), 6},
		{"rejection malformed handshake", uint32(NTRIPRejectionMalformedHandshake), 7},
		{"auth none", uint32(NTRIPSourcetableAuthNone), 0},
		{"auth basic", uint32(NTRIPSourcetableAuthBasic), 1},
		{"auth digest", uint32(NTRIPSourcetableAuthDigest), 2},
		{"auth other", uint32(NTRIPSourcetableAuthOther), 3},
	}
	for _, test := range tests {
		if test.got != test.want {
			t.Errorf("%s = %d, want %d", test.name, test.got, test.want)
		}
	}
}

func TestCatalogRoutesReturnCIdentityAndScheduleValues(t *testing.T) {
	request := gfzUltraRequest()
	identity, err := ResolveProductIdentity(request)
	if err != nil {
		t.Fatal(err)
	}
	if identity.AnalysisCenter != "gfz_ult" || identity.OfficialFilename != "GFZ0OPSULT_20262160000_02D_05M_ORB.SP3" {
		t.Fatalf("identity = %+v", identity)
	}
	if got, err := DefaultSampleForDate(request.Center, request.Family, request.Date); err != nil || got != "05M" {
		t.Fatalf("default sample = %q, %v", got, err)
	}
	samples, err := SupportedSamples(request)
	if err != nil || len(samples) == 0 {
		t.Fatalf("supported samples = %v, %v", samples, err)
	}
	if _, err := ProductSolutionClass(request.Center, request.Family); err != nil {
		t.Fatal(err)
	}
	if _, err := DistributionLocationFor(request, DistributionSourceDirect); err != nil {
		t.Fatal(err)
	}
	if _, err := NextIssueDue(request.Center, request.Family, time.Date(2026, 8, 4, 8, 21, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	candidates, err := PredictedIONEXLineCandidates(time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC), "01H")
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) == 0 {
		t.Fatal("C returned no predicted IONEX candidates")
	}
}

func TestHTTPAcquirerAllowsCatalogURLAndOwnsDecompressedBytes(t *testing.T) {
	var seen []string
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		writeTestResponse(t, response, gzipBytes(t, []byte("catalog-owned product")))
	}))
	t.Cleanup(server.Close)

	acquirer := NewHTTPAcquirer(routedClient(server, &seen))
	acquirer.Now = func() time.Time { return time.Date(2026, 8, 28, 1, 2, 3, 0, time.UTC) }
	product, err := acquirer.Acquire(context.Background(), AcquireRequest{CatalogRequest: gfzUltraRequest(), Source: DistributionSourceDirect})
	if err != nil {
		t.Fatal(err)
	}
	if string(product.Bytes) != "catalog-owned product" || len(seen) != 1 {
		t.Fatalf("bytes=%q requests=%d", product.Bytes, len(seen))
	}
	if product.Provenance.ArchiveCompression != ArchiveCompressionGZIP || product.Provenance.ByteLength != len(product.Bytes) {
		t.Fatalf("provenance = %+v", product.Provenance)
	}
	product.Bytes[0] = 'C'
	if product.Provenance.ResolvedIdentity.OfficialFilename != "GFZ0OPSULT_20262160000_02D_05M_ORB.SP3" {
		t.Fatal("provenance did not retain the C-owned identity copy")
	}
}

func TestHTTPAcquirerFailsBeforeTransportAndRejectsRedirects(t *testing.T) {
	var requests int
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		requests++
		return nil, errors.New("transport should not run")
	})}
	acquirer := NewHTTPAcquirer(client)
	_, err := acquirer.Acquire(context.Background(), AcquireRequest{CatalogRequest: CatalogRequest{Center: "unsupported", Family: ProductFamilySP3, Date: time.Now()}, Source: DistributionSourceDirect})
	if err == nil || requests != 0 {
		t.Fatalf("unsupported catalog result=%v requests=%d", err, requests)
	}

	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set("Location", "http://not-allowlisted.invalid/object")
		response.WriteHeader(http.StatusFound)
	}))
	t.Cleanup(server.Close)
	var seen []string
	redirectAcquirer := NewHTTPAcquirer(routedClient(server, &seen))
	_, err = redirectAcquirer.Acquire(context.Background(), AcquireRequest{CatalogRequest: gfzUltraRequest(), Source: DistributionSourceDirect})
	var redirectErr *RedirectPolicyError
	if !errors.As(err, &redirectErr) {
		t.Fatalf("redirect error = %v", err)
	}
}

func TestHTTPAcquirerStatusSizeAndCancellation(t *testing.T) {
	var mode string
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch mode {
		case "status":
			response.WriteHeader(http.StatusBadGateway)
		case "size":
			writeTestResponse(t, response, []byte("too large"))
		case "cancel":
			<-request.Context().Done()
		}
	}))
	t.Cleanup(server.Close)
	var seen []string
	acquirer := NewHTTPAcquirer(routedClient(server, &seen))
	acquirer.Retries = 1
	mode = "status"
	if _, err := acquirer.Acquire(context.Background(), AcquireRequest{CatalogRequest: gfzUltraRequest(), Source: DistributionSourceDirect}); err == nil {
		t.Fatal("status response unexpectedly succeeded")
	}
	mode = "size"
	acquirer.MaxArchiveBytes = 4
	if _, err := acquirer.Acquire(context.Background(), AcquireRequest{CatalogRequest: gfzUltraRequest(), Source: DistributionSourceDirect}); err == nil {
		t.Fatal("oversized response unexpectedly succeeded")
	}
	mode = "cancel"
	acquirer.MaxArchiveBytes = defaultMaxArchiveBytes
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if _, err := acquirer.Acquire(ctx, AcquireRequest{CatalogRequest: gfzUltraRequest(), Source: DistributionSourceDirect}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("cancellation error = %v", err)
	}
}

func TestBoundedReadsAndRetryBackoffHandleExtremeValues(t *testing.T) {
	const maxInt64 = int64(1<<63 - 1)
	content, err := readBounded(strings.NewReader("small"), maxInt64, "test")
	if err != nil || string(content) != "small" {
		t.Fatalf("maximum bounded read = %q, %v", content, err)
	}
	if _, err := readBounded(strings.NewReader("too long"), 3, "test"); err == nil {
		t.Fatal("bounded read accepted data beyond its limit")
	}
	if got := retryDelay(time.Nanosecond, int(^uint(0)>>1)); got != time.Duration(1<<63-1) {
		t.Fatalf("saturated retry delay = %v", got)
	}
}

func TestNTRIPClientValidatesTransportOptionsBeforeDial(t *testing.T) {
	tests := []struct {
		name     string
		readSize int
		want     int
		wantErr  bool
	}{
		{name: "negative", readSize: -1, wantErr: true},
		{name: "zero selects default", readSize: 0, want: defaultNTRIPReadSize},
		{name: "positive is preserved", readSize: 4096, want: 4096},
		{name: "oversized", readSize: maxNTRIPReadSize + 1, wantErr: true},
	}
	for _, test := range tests {
		client := NTRIPClient{ReadSize: test.readSize}
		t.Run(test.name, func(t *testing.T) {
			got, err := client.validate()
			if test.wantErr {
				if err == nil {
					t.Fatal("invalid client options were accepted")
				}
				return
			}
			if err != nil || got != test.want {
				t.Fatalf("validated read size = %d, %v; want %d, nil", got, err, test.want)
			}
			connection := newNTRIPConnection(nil, nil, 0, got)
			if connection.state.readSize != test.want {
				t.Fatalf("connection read size = %d; want %d", connection.state.readSize, test.want)
			}
			if err := connection.Close(); err != nil {
				t.Fatalf("close validated connection: %v", err)
			}
		})
	}

	var dials int
	dialer := func(context.Context, string, string) (net.Conn, error) {
		dials++
		return nil, errors.New("dial should not run")
	}
	for _, readSize := range []int{-1, maxNTRIPReadSize + 1} {
		client := NTRIPClient{ReadSize: readSize, Dialer: dialer}
		if _, err := client.Connect(context.Background()); err == nil {
			t.Fatalf("invalid read size %d connected", readSize)
		}
	}
	if dials != 0 {
		t.Fatalf("invalid options dialed %d times", dials)
	}
}

func TestNTRIPRejectsEmbeddedNULConfigurationFields(t *testing.T) {
	tests := []struct {
		name  string
		apply func(*NTRIPConfig)
	}{
		{name: "host", apply: func(config *NTRIPConfig) { config.Host = "caster\x00.invalid" }},
		{name: "mountpoint", apply: func(config *NTRIPConfig) { config.Mountpoint = "MOUNT\x00POINT" }},
		{name: "username", apply: func(config *NTRIPConfig) { config.Username = "user\x00name" }},
		{name: "password", apply: func(config *NTRIPConfig) { config.Password = "pass\x00word" }},
		{name: "user agent", apply: func(config *NTRIPConfig) { config.UserAgent = "agent\x00test" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := NTRIPConfig{Host: "caster.example.test", Mountpoint: "MOUNT", HasCredentials: true}
			test.apply(&config)
			if _, err := NTRIPRequestBytes(config); err == nil {
				t.Fatal("embedded NUL was accepted")
			}
		})
	}
	config := NTRIPConfig{Host: "caster.example.test", Mountpoint: "MOUNT", Username: "user\x00name"}
	if _, err := NTRIPRequestBytes(config); err == nil {
		t.Fatal("embedded NUL in an unused credential was accepted")
	}
	var dials int
	client := NTRIPClient{
		Config: NTRIPConfig{Host: "caster\x00.invalid", Mountpoint: "MOUNT"},
		Dialer: func(context.Context, string, string) (net.Conn, error) {
			dials++
			return nil, errors.New("dial should not run")
		},
	}
	if _, err := client.Connect(context.Background()); err == nil {
		t.Fatal("embedded NUL host was accepted by Connect")
	}
	if dials != 0 {
		t.Fatalf("embedded NUL host dialed %d times", dials)
	}
}

type noProgressConn struct{}

func (noProgressConn) Read([]byte) (int, error)         { return 0, nil }
func (noProgressConn) Write([]byte) (int, error)        { return 0, nil }
func (noProgressConn) Close() error                     { return nil }
func (noProgressConn) LocalAddr() net.Addr              { return nil }
func (noProgressConn) RemoteAddr() net.Addr             { return nil }
func (noProgressConn) SetDeadline(time.Time) error      { return nil }
func (noProgressConn) SetReadDeadline(time.Time) error  { return nil }
func (noProgressConn) SetWriteDeadline(time.Time) error { return nil }

func TestNTRIPTransportRejectsNoProgressAndStopsCancellationInterrupter(t *testing.T) {
	connection := noProgressConn{}
	if _, err := readWithContext(context.Background(), connection, make([]byte, 1), 0); !errors.Is(err, io.ErrNoProgress) {
		t.Fatalf("zero-progress read error = %v", err)
	}
	if err := writeWithContext(context.Background(), connection, []byte("x"), 0); !errors.Is(err, io.ErrNoProgress) {
		t.Fatalf("zero-progress write error = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	stop := interruptConnOnCancel(ctx, connection, false)
	if err := stop(); err != nil {
		t.Fatalf("stop cancellation interrupter: %v", err)
	}
	if err := stop(); err != nil {
		t.Fatalf("stop cancellation interrupter twice: %v", err)
	}
	cancel()
}

type blockedDeadlineConn struct {
	mu               sync.Mutex
	readDeadline     time.Time
	writeDeadline    time.Time
	readStarted      chan struct{}
	writeStarted     chan struct{}
	readReleased     chan struct{}
	writeReleased    chan struct{}
	forcedReadSet    chan struct{}
	forcedWriteSet   chan struct{}
	readStartOnce    sync.Once
	writeStartOnce   sync.Once
	readReleaseOnce  sync.Once
	writeReleaseOnce sync.Once
	forcedReadOnce   sync.Once
	forcedWriteOnce  sync.Once
	readCalls        int
	writeCalls       int
}

func newBlockedDeadlineConn() *blockedDeadlineConn {
	return &blockedDeadlineConn{
		readStarted:    make(chan struct{}),
		writeStarted:   make(chan struct{}),
		readReleased:   make(chan struct{}),
		writeReleased:  make(chan struct{}),
		forcedReadSet:  make(chan struct{}),
		forcedWriteSet: make(chan struct{}),
	}
}

func (c *blockedDeadlineConn) Read(buffer []byte) (int, error) {
	c.mu.Lock()
	c.readCalls++
	call := c.readCalls
	c.mu.Unlock()
	if call == 1 {
		c.readStartOnce.Do(func() { close(c.readStarted) })
		<-c.readReleased
		return 0, errors.New("blocked read interrupted")
	}
	return copy(buffer, []byte("read reuse")), nil
}

func (c *blockedDeadlineConn) Write(data []byte) (int, error) {
	c.mu.Lock()
	c.writeCalls++
	call := c.writeCalls
	c.mu.Unlock()
	if call == 1 {
		c.writeStartOnce.Do(func() { close(c.writeStarted) })
		<-c.writeReleased
		return 0, errors.New("blocked write interrupted")
	}
	return len(data), nil
}

func (c *blockedDeadlineConn) Close() error         { return nil }
func (c *blockedDeadlineConn) LocalAddr() net.Addr  { return nil }
func (c *blockedDeadlineConn) RemoteAddr() net.Addr { return nil }
func (c *blockedDeadlineConn) SetDeadline(time.Time) error {
	return nil
}

func (c *blockedDeadlineConn) SetReadDeadline(deadline time.Time) error {
	c.mu.Lock()
	c.readDeadline = deadline
	call := c.readCalls
	c.mu.Unlock()
	if call == 1 && !deadline.IsZero() && !deadline.After(time.Now()) {
		c.forcedReadOnce.Do(func() {
			close(c.forcedReadSet)
			c.readReleaseOnce.Do(func() { close(c.readReleased) })
		})
	}
	return nil
}

func (c *blockedDeadlineConn) SetWriteDeadline(deadline time.Time) error {
	c.mu.Lock()
	c.writeDeadline = deadline
	call := c.writeCalls
	c.mu.Unlock()
	if call == 1 && !deadline.IsZero() && !deadline.After(time.Now()) {
		c.forcedWriteOnce.Do(func() {
			close(c.forcedWriteSet)
			c.writeReleaseOnce.Do(func() { close(c.writeReleased) })
		})
	}
	return nil
}

func (c *blockedDeadlineConn) deadlines() (time.Time, time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.readDeadline, c.writeDeadline
}

func TestNTRIPContextCancellationClearsForcedReadDeadlineAndReusesConn(t *testing.T) {
	conn := newBlockedDeadlineConn()
	ctx, cancel := context.WithCancel(context.Background())
	readResult := make(chan error, 1)
	go func() {
		_, err := readWithContext(ctx, conn, make([]byte, 32), 0)
		readResult <- err
	}()
	<-conn.readStarted
	cancel()
	if err := <-readResult; !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled read error = %v", err)
	}
	readDeadline, _ := conn.deadlines()
	if !readDeadline.IsZero() {
		t.Fatalf("forced read deadline remained %v", readDeadline)
	}

	buffer := make([]byte, 32)
	n, err := readWithContext(context.Background(), conn, buffer, 0)
	if err != nil || string(buffer[:n]) != "read reuse" {
		t.Fatalf("reused read = %q, %v", buffer[:n], err)
	}
	readDeadline, _ = conn.deadlines()
	if !readDeadline.IsZero() {
		t.Fatalf("reused read left deadline %v", readDeadline)
	}

	laterContext, laterCancel := context.WithCancel(context.Background())
	laterConn := newBlockedDeadlineConn()
	stop := interruptConnOnCancel(laterContext, laterConn, false)
	if err := stop(); err != nil {
		t.Fatalf("stop read interrupter: %v", err)
	}
	if err := stop(); err != nil {
		t.Fatalf("stop read interrupter twice: %v", err)
	}
	laterCancel()
	select {
	case <-laterConn.forcedReadSet:
		t.Fatal("stopped read interrupter set a later deadline")
	default:
	}
}

func TestNTRIPContextCancellationClearsForcedWriteDeadlineAndReusesConn(t *testing.T) {
	conn := newBlockedDeadlineConn()
	ctx, cancel := context.WithCancel(context.Background())
	writeResult := make(chan error, 1)
	go func() {
		writeResult <- writeWithContext(ctx, conn, []byte("blocked"), 0)
	}()
	<-conn.writeStarted
	cancel()
	if err := <-writeResult; !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled write error = %v", err)
	}
	_, writeDeadline := conn.deadlines()
	if !writeDeadline.IsZero() {
		t.Fatalf("forced write deadline remained %v", writeDeadline)
	}

	if err := writeWithContext(context.Background(), conn, []byte("reuse"), 0); err != nil {
		t.Fatalf("reused write = %v", err)
	}
	_, writeDeadline = conn.deadlines()
	if !writeDeadline.IsZero() {
		t.Fatalf("reused write left deadline %v", writeDeadline)
	}

	laterContext, laterCancel := context.WithCancel(context.Background())
	laterConn := newBlockedDeadlineConn()
	stop := interruptConnOnCancel(laterContext, laterConn, true)
	if err := stop(); err != nil {
		t.Fatalf("stop write interrupter: %v", err)
	}
	laterCancel()
	select {
	case <-laterConn.forcedWriteSet:
		t.Fatal("stopped write interrupter set a later deadline")
	default:
	}
}

type scriptedRead struct {
	data []byte
	err  error
}

type scriptedConn struct {
	mu    sync.Mutex
	reads []scriptedRead
}

func (c *scriptedConn) Read(buffer []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.reads) == 0 {
		return 0, io.EOF
	}
	read := c.reads[0]
	c.reads = c.reads[1:]
	return copy(buffer, read.data), read.err
}

func (c *scriptedConn) Write(data []byte) (int, error) { return len(data), nil }
func (c *scriptedConn) Close() error                   { return nil }
func (c *scriptedConn) LocalAddr() net.Addr            { return nil }
func (c *scriptedConn) RemoteAddr() net.Addr           { return nil }
func (c *scriptedConn) SetDeadline(time.Time) error    { return nil }
func (c *scriptedConn) SetReadDeadline(time.Time) error {
	return nil
}
func (c *scriptedConn) SetWriteDeadline(time.Time) error {
	return nil
}

type scriptedNTRIPMachine struct {
	pushErr      error
	finishErr    error
	finishEvents []NTRIPEvent
}

func (m *scriptedNTRIPMachine) Push([]byte) ([]NTRIPEvent, error) {
	return nil, m.pushErr
}
func (m *scriptedNTRIPMachine) Finish() ([]NTRIPEvent, error) {
	return m.finishEvents, m.finishErr
}
func (*scriptedNTRIPMachine) TryGGAMessage(float64, float64, NTRIPGGAPosition) ([]byte, bool, error) {
	return nil, false, nil
}
func (*scriptedNTRIPMachine) Close() error { return nil }

func TestNTRIPRunPropagatesMachineFailuresAndAcceptsCleanEOF(t *testing.T) {
	tests := []struct {
		name           string
		read           scriptedRead
		pushErr        error
		finishErr      error
		transportError error
	}{
		{name: "push failure", read: scriptedRead{data: []byte("x"), err: io.EOF}, pushErr: errors.New("push failure")},
		{name: "finish failure", read: scriptedRead{err: io.EOF}, finishErr: errors.New("finish failure")},
		{name: "clean EOF", read: scriptedRead{err: io.EOF}},
		{name: "finish and transport failure", read: scriptedRead{err: io.EOF}, finishErr: errors.New("finish failure"), transportError: errors.New("transport failure")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.transportError != nil {
				test.read.err = errors.Join(test.read.err, test.transportError)
			}
			connection := newNTRIPConnection(
				&scriptedConn{reads: []scriptedRead{test.read}},
				&scriptedNTRIPMachine{pushErr: test.pushErr, finishErr: test.finishErr},
				0,
				32,
			)
			err := (&NTRIPClient{}).runConnection(context.Background(), func(NTRIPEvent) error { return nil }, connection)
			if test.pushErr == nil && test.finishErr == nil {
				if err != nil {
					t.Fatalf("clean EOF Run error = %v", err)
				}
				return
			}
			wantErr := test.pushErr
			if wantErr == nil {
				wantErr = test.finishErr
			}
			if !errors.Is(err, wantErr) {
				t.Fatalf("Run error = %v; want %v", err, wantErr)
			}
			if test.transportError != nil && !errors.Is(err, test.transportError) {
				t.Fatalf("Run error = %v; want transport error %v", err, test.transportError)
			}
			if errors.Is(err, io.EOF) {
				t.Fatalf("machine failure was classified as EOF: %v", err)
			}
		})
	}
}

func TestSendGGAMessageValidatesContextBeforeNativeState(t *testing.T) {
	connection := &NTRIPConnection{state: &ntripConnectionState{}}
	if _, err := connection.SendGGAMessage(nilContextForTest(), 0, 0, NTRIPGGAPosition{}); err == nil {
		t.Fatal("nil context was accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := connection.SendGGAMessage(ctx, 0, 0, NTRIPGGAPosition{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled context error = %v", err)
	}
}

func TestUnixCompressDecompressionCrossesWidthBoundaryAndClear(t *testing.T) {
	archive, err := os.ReadFile("testdata/unix-compress-clear.Z")
	if err != nil {
		t.Fatal(err)
	}
	expected := unixCompressClearFixtureExpected()
	content, err := decompressUnixCompress(archive, int64(len(expected)))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(content, expected) {
		t.Fatalf("CLEAR fixture content mismatch: got %d bytes, want %d", len(content), len(expected))
	}
	if _, err := decompressUnixCompress(archive, int64(len(expected)-1)); err == nil {
		t.Fatal("Unix-compress bound was not enforced")
	}
}

func unixCompressClearFixtureExpected() []byte {
	const (
		pattern       = "GNSS-NTRIP-CLEAR-BOUNDARY\n"
		textBytes     = 16000
		binaryBytes   = 32000
		fixtureLength = textBytes + binaryBytes + textBytes
	)
	textSegment := func() []byte {
		result := bytes.Repeat([]byte(pattern), textBytes/len(pattern))
		return append(result, []byte(pattern)[:textBytes%len(pattern)]...)
	}
	result := make([]byte, 0, fixtureLength)
	result = append(result, textSegment()...)

	state := uint64(0x243f6a8885a308d3)
	for i := 0; i < binaryBytes; i++ {
		state ^= state << 13
		state ^= state >> 7
		state ^= state << 17
		result = append(result, byte(state>>56))
	}
	return append(result, textSegment()...)
}

func TestAcquireLatestUsesAtMostTwoCListingsAndCIdentitySelection(t *testing.T) {
	var listingRequests int
	var seen []string
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if strings.Contains(request.URL.Path, "GFZ0OPSULT") {
			writeTestResponse(t, response, gzipBytes(t, []byte("latest bytes")))
			return
		}
		listingRequests++
		if listingRequests == 1 {
			writeTestResponse(t, response, []byte(`<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 3.2 Final//EN">
<html><head><title>Index of /listing</title></head><body><h1>Index of /listing</h1>
<pre><hr><a href="README">README</a> 2026-08-04 00:00 1<hr></pre></body></html>`))
			return
		}
		writeTestResponse(t, response, []byte(`<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 3.2 Final//EN">
<html><head><title>Index of /listing</title></head><body><h1>Index of /listing</h1>
<pre><hr><a href="GFZ0OPSULT_20262150000_02D_05M_ORB.SP3.gz">GFZ0OPSULT_20262150000_02D_05M_ORB.SP3.gz</a> 2026-08-04 04:26 1.3M<hr></pre></body></html>`))
	}))
	t.Cleanup(server.Close)
	acquirer := NewHTTPAcquirer(routedClient(server, &seen))
	acquirer.Retries = 1
	product, err := acquirer.AcquireLatest(context.Background(), gfzUltraRequest(), DistributionSourceDirect)
	if err != nil {
		t.Fatal(err)
	}
	if string(product.Bytes) != "latest bytes" || listingRequests > 2 {
		t.Fatalf("bytes=%q listing requests=%d all requests=%d", product.Bytes, listingRequests, len(seen))
	}
	if product.Provenance.ResolvedIdentity.OfficialFilename != "GFZ0OPSULT_20262150000_02D_05M_ORB.SP3" {
		t.Fatalf("selected identity = %+v", product.Provenance.ResolvedIdentity)
	}
	if product.Provenance.RequestedIdentity.Date.Equal(product.Provenance.ResolvedIdentity.Date) {
		t.Fatalf("requested and resolved identities were not distinguished: %+v", product.Provenance)
	}
}

const testSourcetable = "STR;MOUNT;ID;RTCM 3;1004;2;GPS;NET;USA;40.0;-105.0;1;0;gen;none;B;N;9600;misc\r\nENDSOURCETABLE\r\n"

func TestNTRIPRequestSourcetableAndChunkedMachine(t *testing.T) {
	config := NTRIPConfig{Host: "caster.example.test", Port: 2101, Mountpoint: "MOUNT", Version: NTRIPVersionRev2, Username: "user", Password: "pass", HasCredentials: true, UserAgent: "sidereon-test/1"}
	request, err := NTRIPRequestBytes(config)
	if err != nil {
		t.Fatal(err)
	}
	requestText := string(request)
	for _, expected := range []string{"GET /MOUNT HTTP/1.1\r\n", "Host: caster.example.test:2101\r\n", "Ntrip-Version: Ntrip/2.0\r\n", "User-Agent: NTRIP sidereon-test/1\r\n", "Authorization: Basic dXNlcjpwYXNz\r\n", "Connection: close\r\n"} {
		if !strings.Contains(requestText, expected) {
			t.Fatalf("request %q lacks %q", requestText, expected)
		}
	}
	if !strings.HasSuffix(requestText, "\r\n\r\n") {
		t.Fatalf("request does not end with an empty line: %q", requestText)
	}

	table, err := ParseNTRIPSourcetable([]byte(testSourcetable))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cleanupTestResource(t, table.Close) })
	streams, err := table.Streams()
	if err != nil || len(streams) != 1 {
		t.Fatalf("streams=%v err=%v", streams, err)
	}
	if streams[0].Mountpoint != "MOUNT" || !streams[0].NMEARequired || streams[0].Authentication != NTRIPSourcetableAuthBasic {
		t.Fatalf("stream = %+v", streams[0])
	}
	if summary, err := table.Summary(); err != nil || summary.StreamCount != 1 {
		t.Fatalf("summary = %+v, %v", summary, err)
	}
	text, err := table.Text()
	if err != nil || !bytes.HasSuffix(text, []byte("ENDSOURCETABLE\r\n")) {
		t.Fatalf("serialized table=%q err=%v", text, err)
	}

	machine, err := NewNTRIPMachine(NTRIPConfig{Host: "caster.example.test", Mountpoint: "MOUNT", Version: NTRIPVersionRev2})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cleanupTestResource(t, machine.Close) })
	if _, err := machine.ConnectionRequest(); err != nil {
		t.Fatal(err)
	}
	wire := []byte("HTTP/1.1 200 OK\r\nContent-Type: gnss/data\r\nTransfer-Encoding: chunked\r\n\r\n3\r\nabc\r\n0\r\n\r\n")
	var events []NTRIPEvent
	for _, value := range wire {
		batch, pushErr := machine.Push([]byte{value})
		if pushErr != nil {
			t.Fatal(pushErr)
		}
		events = append(events, batch...)
	}
	var payload []byte
	var connected, ended bool
	for _, event := range events {
		connected = connected || event.Kind == NTRIPEventConnected && event.Chunked
		ended = ended || event.Kind == NTRIPEventStreamEnded
		if event.Kind == NTRIPEventPayload {
			payload = append(payload, event.Payload...)
		}
	}
	if !connected || !ended || string(payload) != "abc" {
		t.Fatalf("C events=%+v", events)
	}

	ggaMachine, err := NewNTRIPMachine(NTRIPConfig{Host: "caster.example.test", Mountpoint: "MOUNT", Version: NTRIPVersionRev1, HasGGAInterval: true, GGAIntervalS: 10})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cleanupTestResource(t, ggaMachine.Close) })
	if _, err := ggaMachine.ConnectionRequest(); err != nil {
		t.Fatal(err)
	}
	if _, err := ggaMachine.Push([]byte("ICY 200 OK\r\n")); err != nil {
		t.Fatal(err)
	}
	position := NTRIPGGAPosition{LatitudeDeg: 40, LongitudeDeg: -105, HeightM: 1600, FixQuality: 1, Satellites: 10, HDOP: 1}
	gga, present, err := ggaMachine.TryGGAMessage(5, 3661.239, position)
	if err != nil || !present {
		t.Fatalf("first GGA = %q present=%v err=%v", gga, present, err)
	}
	if string(gga) != "$GPGGA,010101.23,4000.0000000,N,10500.0000000,W,1,10,1.00,1600.0,M,,,,*2A\r\n" {
		t.Fatalf("C GGA = %q", gga)
	}
	if _, present, err := ggaMachine.TryGGAMessage(14, 2, position); err != nil || present {
		t.Fatalf("early GGA present=%v err=%v", present, err)
	}
}

func TestNTRIPInputsAreCopiedAcrossTheNativeBoundary(t *testing.T) {
	machine, err := NewNTRIPMachine(NTRIPConfig{Host: "caster.example.test", Mountpoint: "MOUNT", Version: NTRIPVersionRev1})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cleanupTestResource(t, machine.Close) })
	payloadInput := []byte("ICY 200 OK\r\n\r\ncopied payload")
	events, err := machine.Push(payloadInput)
	if err != nil {
		t.Fatal(err)
	}
	for i := range payloadInput {
		payloadInput[i] = 0
	}
	copy(payloadInput, []byte("caller buffer reused"))
	var payload []byte
	for _, event := range events {
		if event.Kind == NTRIPEventPayload {
			payload = append(payload, event.Payload...)
		}
	}
	if string(payload) != "copied payload" {
		t.Fatalf("payload after caller-buffer reuse = %q", payload)
	}

	detailMachine, err := NewNTRIPMachine(NTRIPConfig{Host: "caster.example.test", Mountpoint: "MOUNT", Version: NTRIPVersionRev1})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cleanupTestResource(t, detailMachine.Close) })
	detailInput := []byte("HTTP/1.1 500 Nope\r\n\r\n")
	detailEvents, err := detailMachine.Push(detailInput)
	if err != nil {
		t.Fatal(err)
	}
	if len(detailEvents) == 0 {
		detailEvents, err = detailMachine.Finish()
		if err != nil {
			t.Fatal(err)
		}
	}
	for i := range detailInput {
		detailInput[i] = 'x'
	}
	var detail []byte
	for _, event := range detailEvents {
		if event.Kind == NTRIPEventRejected {
			detail = append(detail, event.Detail...)
		}
	}
	if string(detail) != "Nope" {
		t.Fatalf("detail after caller-buffer mutation = %q events=%+v", detail, detailEvents)
	}

	tableInput := []byte(testSourcetable)
	table, err := ParseNTRIPSourcetable(tableInput)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cleanupTestResource(t, table.Close) })
	for i := range tableInput {
		tableInput[i] = 'x'
	}
	copy(tableInput, []byte("reused sourcetable buffer"))
	summary, err := table.Summary()
	if err != nil {
		t.Fatal(err)
	}
	if summary != (NTRIPSourcetableSummary{RecordCount: 1, StreamCount: 1}) {
		t.Fatalf("summary after caller-buffer reuse = %+v", summary)
	}
	streams, err := table.Streams()
	if err != nil {
		t.Fatal(err)
	}
	wantStream := NTRIPStream{Mountpoint: "MOUNT", Identifier: "ID", Format: "RTCM 3", FormatDetails: "1004", HasCarrier: true, Carrier: 2, NavSystem: "GPS", Network: "NET", Country: "USA", HasLatitudeDeg: true, LatitudeDeg: 40, HasLongitudeDeg: true, LongitudeDeg: -105, HasNMEARequired: true, NMEARequired: true, HasNetworkSolution: true, NetworkSolution: false, Generator: "gen", Compression: "none", Authentication: NTRIPSourcetableAuthBasic, HasFee: true, Fee: false, HasBitrate: true, Bitrate: 9600, Misc: "misc"}
	if len(streams) != 1 || streams[0] != wantStream {
		t.Fatalf("streams after caller-buffer reuse = %+v", streams)
	}
	text, err := table.Text()
	wantText := "STR;MOUNT;ID;RTCM 3;1004;2;GPS;NET;USA;40;-105;1;0;gen;none;B;N;9600;misc\r\nENDSOURCETABLE\r\n"
	if err != nil || !bytes.Equal(text, []byte(wantText)) {
		t.Fatalf("text after caller-buffer reuse = %q, %v", text, err)
	}
}

func TestNTRIPCloseResetUseAfterCloseAndConcurrentClose(t *testing.T) {
	machine, err := NewNTRIPMachine(NTRIPConfig{Host: "caster.example.test", Mountpoint: "MOUNT"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cleanupTestResource(t, machine.Close) })
	if err := machine.Reset(); err != nil {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	for i := 0; i < 16; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			if closeErr := machine.Close(); closeErr != nil {
				t.Errorf("concurrent machine close: %v", closeErr)
			}
			if _, stateErr := machine.State(); stateErr != nil && !errors.Is(stateErr, ErrClosed) {
				t.Errorf("state after concurrent close: %v", stateErr)
			}
		}()
	}
	group.Wait()
	if _, err := machine.Push([]byte("after-close")); !errors.Is(err, ErrClosed) {
		t.Fatalf("Push after close = %v", err)
	}
	if _, err := machine.ConnectionRequest(); !errors.Is(err, ErrClosed) {
		t.Fatalf("ConnectionRequest after close = %v", err)
	}

	table, err := ParseNTRIPSourcetable([]byte(testSourcetable))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cleanupTestResource(t, table.Close) })
	if err := table.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := table.Text(); !errors.Is(err, ErrClosed) {
		t.Fatalf("Text after close = %v", err)
	}
}

func TestNTRIPSourcetableConcurrentReadsAndClose(t *testing.T) {
	table, err := ParseNTRIPSourcetable([]byte(testSourcetable))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cleanupTestResource(t, table.Close) })
	const readers = 32
	const iterations = 64
	start := make(chan struct{})
	errCh := make(chan error, readers*iterations*3)
	var group sync.WaitGroup
	for i := 0; i < readers; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			for j := 0; j < iterations; j++ {
				summary, summaryErr := table.Summary()
				if summaryErr != nil && !errors.Is(summaryErr, ErrClosed) {
					errCh <- summaryErr
				} else if summaryErr == nil && summary.StreamCount != 1 {
					errCh <- errors.New("concurrent summary returned an unexpected stream count")
				}
				if _, streamsErr := table.Streams(); streamsErr != nil && !errors.Is(streamsErr, ErrClosed) {
					errCh <- streamsErr
				}
				if _, textErr := table.Text(); textErr != nil && !errors.Is(textErr, ErrClosed) {
					errCh <- textErr
				}
			}
		}()
	}
	closeResults := make(chan error, 2)
	go func() {
		<-start
		closeResults <- table.Close()
		closeResults <- table.Close()
	}()
	close(start)
	group.Wait()
	for i := 0; i < cap(closeResults); i++ {
		if err := <-closeResults; err != nil {
			t.Fatalf("concurrent sourcetable close = %v", err)
		}
	}
	close(errCh)
	for err := range errCh {
		t.Errorf("concurrent sourcetable read: %v", err)
	}
	if _, err := table.Summary(); !errors.Is(err, ErrClosed) {
		t.Fatalf("Summary after sourcetable close = %v", err)
	}
	if _, err := table.Streams(); !errors.Is(err, ErrClosed) {
		t.Fatalf("Streams after sourcetable close = %v", err)
	}
	if _, err := table.Text(); !errors.Is(err, ErrClosed) {
		t.Fatalf("Text after sourcetable close = %v", err)
	}
}

func TestNTRIPConnectionUsesInjectedNetPipeAndCloseInterruptsRead(t *testing.T) {
	clientSide, serverSide := net.Pipe()
	requestReady := make(chan struct{})
	go func() {
		defer func() { cleanupTestResource(t, serverSide.Close) }()
		buffer := make([]byte, 4096)
		var request []byte
		for !bytes.Contains(request, []byte("\r\n\r\n")) {
			n, err := serverSide.Read(buffer)
			if err != nil {
				return
			}
			request = append(request, buffer[:n]...)
		}
		close(requestReady)
		if _, err := serverSide.Write([]byte("ICY 200 OK\r\n\r\nabc")); err != nil {
			t.Errorf("write NTRIP response: %v", err)
		}
	}()
	client := NTRIPClient{Config: NTRIPConfig{Host: "caster.example.test", Mountpoint: "MOUNT", Version: NTRIPVersionRev1}, Dialer: func(context.Context, string, string) (net.Conn, error) { return clientSide, nil }, ReadSize: 2}
	connection, err := client.Connect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cleanupTestResource(t, connection.Close) })
	select {
	case <-requestReady:
	case <-time.After(time.Second):
		t.Fatal("dialed peer did not receive the C request")
	}
	var payload []byte
	for len(payload) < 3 {
		events, readErr := connection.Read(context.Background())
		for _, event := range events {
			if event.Kind == NTRIPEventPayload {
				payload = append(payload, event.Payload...)
			}
		}
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			t.Fatal(readErr)
		}
	}
	if string(payload) != "abc" {
		t.Fatalf("payload=%q", payload)
	}
	for {
		_, readErr := connection.Read(context.Background())
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			t.Fatalf("end-of-stream read = %v", readErr)
		}
	}

	blockedClient, blockedServer := net.Pipe()
	t.Cleanup(func() { cleanupTestResource(t, blockedServer.Close) })
	blockedReady := make(chan struct{})
	go func() {
		buffer := make([]byte, 4096)
		if _, err := blockedServer.Read(buffer); err != nil {
			t.Errorf("read blocked NTRIP request: %v", err)
			return
		}
		close(blockedReady)
	}()
	blocked := NTRIPClient{Config: NTRIPConfig{Host: "caster.example.test", Mountpoint: "MOUNT"}, Dialer: func(context.Context, string, string) (net.Conn, error) { return blockedClient, nil }, ReadSize: 32 << 10}
	blockedConnection, err := blocked.Connect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	<-blockedReady
	readDone := make(chan error, 1)
	go func() {
		_, readErr := blockedConnection.Read(context.Background())
		readDone <- readErr
	}()
	if err := blockedConnection.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-readDone:
		if !errors.Is(err, ErrClosed) {
			t.Fatalf("interrupted read = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Close did not interrupt the read")
	}
}
