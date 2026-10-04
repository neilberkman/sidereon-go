package sidereon

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

type diagnosticRoundTripFunc func(*http.Request) (*http.Response, error)

func (f diagnosticRoundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

type diagnosticSecretError string

func (e diagnosticSecretError) Error() string { return "transport secret: " + string(e) }

type diagnosticSecretPanic struct{ secret string }

func (p diagnosticSecretPanic) String() string { return p.secret }

func TestHTTPAcquirerTransportDiagnosticsPreserveReturnedError(t *testing.T) {
	const secret = "transport-password-should-not-enter-diagnostic"
	newAcquirer := func(observer func(HTTPTransportDiagnostic)) (*HTTPAcquirer, *int) {
		calls := new(int)
		transport := diagnosticRoundTripFunc(func(*http.Request) (*http.Response, error) {
			(*calls)++
			return nil, diagnosticSecretError(secret)
		})
		acquirer := NewHTTPAcquirer(&http.Client{Transport: transport})
		acquirer.Retries = 2
		acquirer.Backoff = 0
		acquirer.OnTransportDiagnostic = observer
		return acquirer, calls
	}

	baseline, baselineCalls := newAcquirer(nil)
	_, baselineErr := baseline.Acquire(context.Background(), AcquireRequest{
		CatalogRequest: gfzUltraRequest(), Source: DistributionSourceDirect,
	})
	if baselineErr == nil || *baselineCalls != 2 {
		t.Fatalf("baseline err=%v calls=%d, want terminal error after 2 attempts", baselineErr, *baselineCalls)
	}

	var events []HTTPTransportDiagnostic
	observed, observedCalls := newAcquirer(func(event HTTPTransportDiagnostic) {
		events = append(events, event)
		panic(secret)
	})
	_, observedErr := observed.Acquire(context.Background(), AcquireRequest{
		CatalogRequest: gfzUltraRequest(), Source: DistributionSourceDirect,
	})
	if observedErr == nil || observedErr.Error() != baselineErr.Error() || *observedCalls != *baselineCalls {
		t.Fatalf("observer changed returned error/retries: got err=%v calls=%d; baseline err=%v calls=%d", observedErr, *observedCalls, baselineErr, *baselineCalls)
	}
	if len(events) != 2 {
		t.Fatalf("diagnostic count=%d, want one event per failed transport attempt", len(events))
	}
	for _, event := range events {
		if event.Kind != "round_trip_error" || !strings.HasSuffix(event.TransportType, "diagnosticRoundTripFunc") || !strings.HasSuffix(event.FailureType, "diagnosticSecretError") {
			t.Fatalf("diagnostic types = %+v", event)
		}
		if len(event.CallSites) == 0 || len(event.CallSites) > maxTransportDiagnosticCallSites {
			t.Fatalf("diagnostic call sites = %v", event.CallSites)
		}
		foundRoundTrip := false
		for _, name := range event.CallSites {
			if strings.Contains(name, "diagnosticRoundTripFunc.RoundTrip") {
				foundRoundTrip = true
			}
			if strings.Contains(name, secret) || strings.Contains(name, ".go:") {
				t.Fatalf("diagnostic contains secret or source location: %q", name)
			}
		}
		if !foundRoundTrip {
			t.Fatalf("custom RoundTripper method missing from call sites: %v", event.CallSites)
		}
		if strings.Contains(event.TransportType, secret) || strings.Contains(event.FailureType, secret) {
			t.Fatalf("diagnostic type contains secret: %+v", event)
		}
	}
	if !strings.Contains(observedErr.Error(), secret) {
		t.Fatalf("original returned error no longer exposes its unchanged wrapped cause: %v", observedErr)
	}
}

func TestHTTPAcquirerTransportDiagnosticsRepanicOriginalValueWithoutFallback(t *testing.T) {
	listingURLs, err := PublicationListingURLs(gfzUltraRequest())
	if err != nil {
		t.Fatal(err)
	}
	if len(listingURLs) != 2 {
		t.Fatalf("C listing URL count = %d, want two fallback candidates", len(listingURLs))
	}
	const secret = "panic-password-should-not-enter-diagnostic"
	panicValue := diagnosticSecretPanic{secret: secret}
	calls := 0
	transport := diagnosticRoundTripFunc(func(*http.Request) (*http.Response, error) {
		calls++
		panic(panicValue)
	})
	var events []HTTPTransportDiagnostic
	acquirer := NewHTTPAcquirer(&http.Client{Transport: transport})
	acquirer.OnTransportDiagnostic = func(event HTTPTransportDiagnostic) {
		events = append(events, event)
		panic("observer panic is ignored")
	}

	var recovered any
	func() {
		defer func() { recovered = recover() }()
		_, _ = acquirer.AcquireLatest(context.Background(), gfzUltraRequest(), DistributionSourceDirect)
	}()
	if !reflect.DeepEqual(recovered, panicValue) || calls != 1 {
		t.Fatalf("recovered=%#v calls=%d, want original panic and no retry/fallback", recovered, calls)
	}
	if len(events) != 1 {
		t.Fatalf("diagnostic count=%d, want 1", len(events))
	}
	event := events[0]
	if event.Kind != "round_trip_panic" || !strings.HasSuffix(event.TransportType, "diagnosticRoundTripFunc") || !strings.HasSuffix(event.FailureType, "diagnosticSecretPanic") {
		t.Fatalf("diagnostic = %+v", event)
	}
	foundRoundTrip := false
	for _, name := range event.CallSites {
		if strings.Contains(name, "diagnosticRoundTripFunc.RoundTrip") {
			foundRoundTrip = true
		}
		if strings.Contains(name, secret) || strings.Contains(name, ".go:") {
			t.Fatalf("diagnostic contains secret or source location: %q", name)
		}
	}
	if !foundRoundTrip || len(event.CallSites) == 0 || len(event.CallSites) > maxTransportDiagnosticCallSites {
		t.Fatalf("bounded call sites do not identify custom transport: %v", event.CallSites)
	}
}

func TestHTTPAcquirerTransportDiagnosticsDoNotLeakAnonymousPanicTypeTag(t *testing.T) {
	const tagSecret = "transport-diagnostic-tag-secret"
	const valueSecret = "transport-diagnostic-value-secret"
	panicValue := struct {
		Payload string `diagnostic:"transport-diagnostic-tag-secret"`
	}{Payload: valueSecret}
	transport := diagnosticRoundTripFunc(func(*http.Request) (*http.Response, error) {
		panic(panicValue)
	})
	var event HTTPTransportDiagnostic
	acquirer := NewHTTPAcquirer(&http.Client{Transport: transport})
	acquirer.OnTransportDiagnostic = func(got HTTPTransportDiagnostic) { event = got }
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		_, _ = acquirer.Acquire(context.Background(), AcquireRequest{
			CatalogRequest: gfzUltraRequest(), Source: DistributionSourceDirect,
		})
	}()
	if !reflect.DeepEqual(recovered, panicValue) {
		t.Fatalf("recovered panic = %#v, want original value", recovered)
	}
	if event.Kind != "round_trip_panic" || event.FailureType != "struct" {
		t.Fatalf("anonymous panic diagnostic = %+v", event)
	}
	for _, text := range append([]string{event.Kind, event.TransportType, event.FailureType}, event.CallSites...) {
		if strings.Contains(text, tagSecret) || strings.Contains(text, valueSecret) {
			t.Fatalf("diagnostic leaked panic tag/value: %q", text)
		}
	}
}

func TestHTTPAcquirerTransportDiagnosticsDoNotObserveHTTPStatus(t *testing.T) {
	calls := 0
	transport := diagnosticRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		calls++
		return &http.Response{
			StatusCode: http.StatusBadGateway,
			Header:     make(http.Header),
			Body:       http.NoBody,
			Request:    request,
		}, nil
	})
	var events []HTTPTransportDiagnostic
	acquirer := NewHTTPAcquirer(&http.Client{Transport: transport})
	acquirer.OnTransportDiagnostic = func(event HTTPTransportDiagnostic) { events = append(events, event) }
	_, err := acquirer.Acquire(context.Background(), AcquireRequest{
		CatalogRequest: gfzUltraRequest(), Source: DistributionSourceDirect,
	})
	var statusErr *HTTPStatusError
	if !errors.As(err, &statusErr) || calls != 1 || len(events) != 0 {
		t.Fatalf("status err=%v calls=%d diagnostics=%v", err, calls, events)
	}
}
