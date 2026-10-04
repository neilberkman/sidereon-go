package sidereon

import (
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"unicode/utf8"
)

const (
	maxTransportDiagnosticCallSites = 8
	maxTransportDiagnosticNameBytes = 256
)

// HTTPTransportDiagnostic describes a RoundTripper failure without carrying
// its error or panic value, request data, source locations, or response data.
// Kind is "round_trip_error" or "round_trip_panic". TransportType and
// FailureType are named-type labels, or kind labels for unnamed composites.
// CallSites contains at most eight function names and no file paths or line
// numbers. Function symbols with composite-type syntax are omitted to avoid
// exposing embedded struct tags.
type HTTPTransportDiagnostic struct {
	Kind          string
	TransportType string
	FailureType   string
	CallSites     []string
}

type diagnosticRoundTripper struct {
	next     http.RoundTripper
	observer func(HTTPTransportDiagnostic)
}

// RoundTrip forwards the request and reports failures through the optional observer.
func (t diagnosticRoundTripper) RoundTrip(request *http.Request) (response *http.Response, err error) {
	defer func() {
		if failure := recover(); failure != nil {
			t.observe("round_trip_panic", failure)
			panic(failure)
		}
	}()

	response, err = t.next.RoundTrip(request)
	if err != nil {
		t.observe("round_trip_error", err)
	}
	return response, err
}

// observe deliberately treats all diagnostic construction and callback work as
// best-effort so it cannot alter a transport result or the original panic.
func (t diagnosticRoundTripper) observe(kind string, failure any) {
	defer func() { _ = recover() }()
	failureType := diagnosticTypeName(reflect.TypeOf(failure))
	transportType := diagnosticTypeName(reflect.TypeOf(t.next))
	t.observer(HTTPTransportDiagnostic{
		Kind:          kind,
		TransportType: transportType,
		FailureType:   failureType,
		CallSites:     diagnosticCallSites(t.next),
	})
}

func diagnosticCallSites(transport http.RoundTripper) []string {
	var names []string
	if typ := reflect.TypeOf(transport); typ != nil {
		names = appendDiagnosticName(names, diagnosticTypeName(typ)+".RoundTrip")
	}
	var pcs [24]uintptr
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	for len(names) < maxTransportDiagnosticCallSites {
		frame, more := frames.Next()
		if safeDiagnosticFunction(frame.Function) {
			names = appendDiagnosticName(names, frame.Function)
		}
		if !more {
			break
		}
	}
	return names
}

func diagnosticTypeName(typ reflect.Type) string {
	if typ == nil {
		return "<nil>"
	}
	if typ.Name() != "" {
		name := typ.Name()
		if generic := strings.IndexByte(name, '['); generic >= 0 {
			name = name[:generic]
		}
		if packagePath := typ.PkgPath(); packagePath != "" {
			return boundedDiagnosticName(packagePath + "." + name)
		}
		return boundedDiagnosticName(name)
	}
	if typ.Kind() == reflect.Pointer {
		return boundedDiagnosticName("*" + diagnosticTypeName(typ.Elem()))
	}
	return typ.Kind().String()
}

func safeDiagnosticFunction(name string) bool {
	// Runtime symbols for instantiated generic or anonymous-composite types may
	// include type syntax and struct tags. The concrete transport method is
	// recorded separately using diagnosticTypeName, so omit such frames.
	return name != "" && !strings.ContainsAny(name, "\"`{}[]")
}

func appendDiagnosticName(names []string, name string) []string {
	name = boundedDiagnosticName(name)
	for _, existing := range names {
		if existing == name {
			return names
		}
	}
	if len(names) < maxTransportDiagnosticCallSites {
		return append(names, name)
	}
	return names
}

func boundedDiagnosticName(name string) string {
	if len(name) <= maxTransportDiagnosticNameBytes {
		return name
	}
	name = name[:maxTransportDiagnosticNameBytes]
	for !utf8.ValidString(name) {
		name = name[:len(name)-1]
	}
	return name
}
