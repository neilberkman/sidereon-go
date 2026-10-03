package sidereon

import "fmt"

// TDMErrorDetail is the lossless public detail attached to a generic engine
// error from the TDM family. Fields retains every JSON value and number token
// supplied by the C ABI, including fields from future error variants.
type TDMErrorDetail struct {
	// Kind is the stable TDM error discriminator.
	Kind string
	// Fields is the recursive, lossless field object for Kind.
	Fields map[string]EngineJSONValue
	// Display is the native diagnostic text captured with the same C call.
	Display string
}

// Error returns native display text when present, preserving the C diagnostic;
// otherwise it returns a stable summary for manually constructed values.
func (e *TDMErrorDetail) Error() string {
	if e == nil {
		return "sidereon: TDM error"
	}
	if e.Display != "" {
		return e.Display
	}
	if e.Kind == "" {
		return "sidereon: TDM error"
	}
	return fmt.Sprintf("sidereon: TDM error kind=%s", e.Kind)
}
