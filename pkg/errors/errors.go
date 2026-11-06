package apperror

import (
	"errors"
	"fmt"
	"runtime"
)

type Kind string

const (
	KindValidation   Kind = "VALIDATION_ERROR"
	KindNotFound     Kind = "NOT_FOUND"
	KindConflict     Kind = "CONFLICT"
	KindUnauthorized Kind = "UNAUTHORIZED"
	KindForbidden    Kind = "FORBIDDEN"
	KindInternal     Kind = "INTERNAL_ERROR"
)

type Error struct {
	Kind    Kind
	Message string
	Cause   error

	// pc is the program counter of the Wrap call that created this error.
	// It is zero for errors made with New, which are usually package-level
	// sentinels whose creation site says nothing about where they were returned.
	pc uintptr
}

func New(kind Kind, message string) *Error {
	return &Error{
		Kind:    kind,
		Message: message,
	}
}

func Wrap(kind Kind, message string, cause error) *Error {
	return &Error{
		Kind:    kind,
		Message: message,
		Cause:   cause,
		pc:      callerPC(),
	}
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Cause
}

// Wrap returns a copy of e with cause attached. The receiver is never
// modified, so it is safe to call on shared package-level errors.
func (e *Error) Wrap(cause error) *Error {
	return &Error{
		Kind:    e.Kind,
		Message: e.Message,
		Cause:   cause,
		pc:      callerPC(),
	}
}

// callerPC returns the program counter of the function that called the
// Wrap that called callerPC.
func callerPC() uintptr {
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:]) // skip runtime.Callers, callerPC and Wrap
	return pcs[0]
}

// Origin returns "file:line" of the deepest Wrap call in err's chain, which
// is the point closest to where the failure happened. It returns "" when no
// *Error in the chain was created by Wrap.
func Origin(err error) string {
	var pc uintptr
	for e := err; e != nil; e = errors.Unwrap(e) {
		if appErr, ok := e.(*Error); ok && appErr.pc != 0 {
			pc = appErr.pc
		}
	}

	if pc == 0 {
		return ""
	}

	frame, _ := runtime.CallersFrames([]uintptr{pc}).Next()
	return fmt.Sprintf("%s:%d", frame.File, frame.Line)
}

// Is reports whether target is an *Error with the same Kind and Message,
// so errors created with Wrap still match their sentinel via errors.Is.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	return ok && e.Kind == t.Kind && e.Message == t.Message
}

func ErrorKind(err error) Kind {
	if err == nil {
		return ""
	}

	if appErr, ok := AsAppError(err); ok {
		return appErr.Kind
	}

	return KindInternal
}

func AsAppError(err error) (*Error, bool) {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
