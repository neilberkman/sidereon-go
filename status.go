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
	Code   StatusCode
	Text   string
	Detail string
}

func (e *StatusError) Error() string {
	if e.Detail == "" {
		return e.Text
	}
	return fmt.Sprintf("%s: %s", e.Text, e.Detail)
}

// ErrClosed is returned when an operation uses a handle after Close.
var ErrClosed = errors.New("sidereon: handle is closed")

var errNilNativeHandle = errors.New("sidereon: native constructor returned no handle")

func publicError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, native.ErrClosed) {
		return ErrClosed
	}
	var statusErr *native.StatusError
	if errors.As(err, &statusErr) {
		return &StatusError{
			Code:   StatusCode(statusErr.Code),
			Text:   statusErr.Text,
			Detail: statusErr.Detail,
		}
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
