package sidereon

import (
	"errors"
	"fmt"

	"github.com/neilberkman/sidereon-go/internal/native"
)

// StatusCode is the numeric status returned by the C ABI.
type StatusCode int

const (
	StatusOK              StatusCode = 0
	StatusNullPointer     StatusCode = 1
	StatusInvalidArgument StatusCode = 2
	StatusInvalidToken    StatusCode = 3
	StatusSP3Parse        StatusCode = 4
	StatusSolve           StatusCode = 5
	StatusPanic           StatusCode = 6
	StatusTimeout         StatusCode = 7
)

// StatusError reports one failed C call. Text is the stable status name from
// sidereon_status_message; Detail is the thread-local reason captured during
// the same call.
type StatusError struct {
	Code         StatusCode
	Text         string
	Detail       string
	TerrainDatum *TerrainDatumError
	TerrainStore *TerrainStoreError
}

// FallbackStatus is the distinct status namespace returned by the
// precise/broadcast fallback solver.
type FallbackStatus uint32

const (
	FallbackOK              FallbackStatus = 0
	FallbackNullPointer     FallbackStatus = 1
	FallbackInvalidArgument FallbackStatus = 2
	FallbackInvalidToken    FallbackStatus = 3
	FallbackPanic           FallbackStatus = 4
	FallbackPreciseSolve    FallbackStatus = 5
	FallbackBroadcastSolve  FallbackStatus = 6
)

func (s FallbackStatus) name() string {
	switch s {
	case FallbackOK:
		return "ok"
	case FallbackNullPointer:
		return "null pointer"
	case FallbackInvalidArgument:
		return "invalid argument"
	case FallbackInvalidToken:
		return "invalid token"
	case FallbackPanic:
		return "panic"
	case FallbackPreciseSolve:
		return "precise solve"
	case FallbackBroadcastSolve:
		return "broadcast solve"
	default:
		return "unknown"
	}
}

// FallbackError reports one failed fallback solve. Detail is the native
// thread-local reason captured during the same call.
type FallbackError struct {
	Status FallbackStatus
	Detail string
}

func (e *FallbackError) Error() string {
	if e == nil {
		return "sidereon fallback solve failed"
	}
	message := fmt.Sprintf("sidereon fallback solve failed: %s (status %d)", e.Status.name(), e.Status)
	if e.Detail != "" {
		message += ": " + e.Detail
	}
	return message
}

func (e *StatusError) Error() string {
	if e.Detail == "" {
		return e.Text
	}
	return fmt.Sprintf("%s: %s", e.Text, e.Detail)
}

func (e *StatusError) Unwrap() error {
	var details []error
	if e.TerrainDatum != nil {
		details = append(details, e.TerrainDatum)
	}
	if e.TerrainStore != nil {
		details = append(details, e.TerrainStore)
	}
	return errors.Join(details...)
}

// ErrClosed is returned when an operation uses a handle after Close.
var ErrClosed = errors.New("sidereon: handle is closed")

func nativeCountToInt(value uint64, field string) (int, error) {
	if value > uint64(^uint(0)>>1) {
		return 0, fmt.Errorf("sidereon: native %s %d does not fit in int", field, value)
	}
	return int(value), nil
}

var errNilNativeHandle = errors.New("sidereon: native constructor returned no handle")

func (e *TerrainDatumError) Error() string {
	if e == nil {
		return "sidereon: terrain datum error"
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Remediation != "" {
		return e.Remediation
	}
	if e.Path != "" {
		return e.Path
	}
	return "sidereon: terrain datum error"
}

func (e *TerrainStoreError) Error() string {
	if e == nil {
		return "sidereon: terrain store error"
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Reason != "" {
		return e.Reason
	}
	if e.Path != "" {
		return e.Path
	}
	return "sidereon: terrain store error"
}

func publicError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, native.ErrClosed) {
		return ErrClosed
	}
	var selectionErr *native.SelectionError
	if errors.As(err, &selectionErr) {
		return &SelectionError{
			Status: SelectionStatus(selectionErr.Status),
			Text:   selectionErr.Text,
			Detail: selectionErr.Detail,
		}
	}
	var statusErr *native.StatusError
	if errors.As(err, &statusErr) {
		result := &StatusError{
			Code:   StatusCode(statusErr.Code),
			Text:   statusErr.Text,
			Detail: statusErr.Detail,
		}
		if statusErr.TerrainDatum != nil {
			value := *statusErr.TerrainDatum
			result.TerrainDatum = &TerrainDatumError{
				Kind:        TerrainDatumErrorKind(value.Kind),
				Path:        value.Path,
				Message:     value.Message,
				Remediation: value.Remediation,
			}
		}
		if statusErr.TerrainStore != nil {
			value := *statusErr.TerrainStore
			result.TerrainStore = &TerrainStoreError{
				Kind:             TerrainStoreErrorKind(value.Kind),
				Path:             value.Path,
				Message:          value.Message,
				Reason:           value.Reason,
				Version:          value.Version,
				Tag:              value.Tag,
				LatIndex:         value.LatIndex,
				LonIndex:         value.LonIndex,
				ExpectedChecksum: value.ExpectedChecksum,
				FoundChecksum:    value.FoundChecksum,
			}
		}
		return result
	}
	var fallbackErr *native.FallbackError
	if errors.As(err, &fallbackErr) {
		return &FallbackError{Status: FallbackStatus(fallbackErr.Status), Detail: fallbackErr.Detail}
	}
	return err
}

func joinPublicErrors(errs ...error) error {
	translated := make([]error, 0, len(errs))
	for _, err := range errs {
		if err != nil {
			translated = append(translated, publicError(err))
		}
	}
	return errors.Join(translated...)
}
