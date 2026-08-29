package sidereon

import "fmt"

// SelectionError reports a failed product-selection operation. Status is the
// SidereonSelectionStatus value, which has a namespace distinct from
// StatusCode; Detail is the native thread-local diagnostic when available.
type SelectionError struct {
	// Status is the native status code.
	Status SelectionStatus
	// Text is the stable human-readable name for Status.
	Text string
	// Detail is the same-call native thread-local diagnostic, when available.
	Detail string
}

// Error formats the selection status code, stable name, and optional native detail.
func (e *SelectionError) Error() string {
	if e == nil {
		return "sidereon: selection error"
	}
	if e.Detail == "" {
		return fmt.Sprintf("%s (%d)", e.Text, e.Status)
	}
	return fmt.Sprintf("%s (%d): %s", e.Text, e.Status, e.Detail)
}
