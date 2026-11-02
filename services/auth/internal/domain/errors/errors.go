package errors

import apperror "github.com/umohsamuel/eba/pkg/errors"

var (
	ErrBadRequest = apperror.New(
		apperror.KindValidation,
		"bad request",
	)

	ErrInternal = apperror.New(
		apperror.KindInternal,
		"internal server error",
	)

	ErrUserNotFound = apperror.New(
		apperror.KindNotFound,
		"user not found",
	)

	ErrEmailAlreadyExists = apperror.New(
		apperror.KindConflict,
		"email already registered",
	)

	ErrInvalidCredentials = apperror.New(
		apperror.KindUnauthorized,
		"invalid credentials",
	)

	ErrAccountSuspended = apperror.New(
		apperror.KindForbidden,
		"account suspended",
	)

	ErrInvalidEmail = apperror.New(
		apperror.KindValidation,
		"provided email address is invalid",
	)

	ErrInvalidPassword = apperror.New(
		apperror.KindValidation,
		"password must be 8-32 characters long and contain at least one uppercase letter, one lowercase letter, one digit, and one special character",
	)
)
