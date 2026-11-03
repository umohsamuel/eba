package apperror

import "errors"

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
	}
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
