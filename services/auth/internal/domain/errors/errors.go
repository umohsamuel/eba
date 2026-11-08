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

	ErrNotFound = apperror.New(
		apperror.KindNotFound,
		"not found",
	)

	ErrUserNotFound = apperror.New(
		apperror.KindNotFound,
		"user not found",
	)

	ErrRefreshTokenNotFound = apperror.New(
		apperror.KindNotFound,
		"refresh token not found",
	)

	ErrTokenExpired = apperror.New(apperror.KindUnauthorized, "token expired")
	ErrTokenInvalid = apperror.New(apperror.KindUnauthorized, "invalid token")

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

	ErrNoMigrationToApply = apperror.New(
		apperror.KindConflict,
		"no migration to apply",
	)

	ErrMigrationVersionNotFound = apperror.New(
		apperror.KindNotFound,
		"migration version not found",
	)

	ErrInvalidPassword = apperror.New(
		apperror.KindValidation,
		"password must be 8-32 characters long and contain at least one uppercase letter, one lowercase letter, one digit, and one special character",
	)
)
