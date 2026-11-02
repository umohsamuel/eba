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

func ErrorKind(err error) Kind {
	if err == nil {
		return ""
	}

	var customErr *Error
	if errors.As(err, &customErr) {
		return customErr.Kind
	}

	return KindInternal
}
