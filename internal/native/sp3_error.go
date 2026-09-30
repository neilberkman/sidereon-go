package native

import "fmt"

// SP3Error preserves the native typed summary and its lossless JSON payload.
// Integer tick values remain in PayloadJSON as decimal strings.
type SP3Error struct {
	Kind               uint32
	Field              uint32
	Reason             uint32
	HasValue           bool
	Value              float64
	HasRequestedTick   bool
	HasDeclaredTick    bool
	HasRequestedJ2000S bool
	HasDeclaredJ2000S  bool
	RequestedJ2000S    float64
	DeclaredJ2000S     float64
	PayloadJSON        string
}

func (e *SP3Error) Error() string {
	if e == nil {
		return "sidereon: SP3 validation error"
	}
	return fmt.Sprintf("sidereon: SP3 validation kind=%d field=%d reason=%d", e.Kind, e.Field, e.Reason)
}
