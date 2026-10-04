package sidereon

import "fmt"

// Error returns the retained antenna refusal message or its numeric category.
func (e *ANTEXError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("sidereon: ANTEX error %d", e.Kind)
}
