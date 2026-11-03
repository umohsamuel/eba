package gin

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	apperror "github.com/umohsamuel/eba/pkg/errors"
	"github.com/umohsamuel/eba/pkg/response"
)

func mapErrorKindToHTTP(kind apperror.Kind) int {
	switch kind {
	case apperror.KindValidation:
		return http.StatusBadRequest
	case apperror.KindNotFound:
		return http.StatusNotFound
	case apperror.KindConflict:
		return http.StatusConflict
	case apperror.KindUnauthorized:
		return http.StatusUnauthorized
	case apperror.KindForbidden:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

type GinResponder[T any] struct {
	ctx        *gin.Context
	statusCode int
	envelope   response.Envelope[T]
}

func HandleError(c *gin.Context, err error) GinResponder[any] {
	kind := apperror.KindInternal
	message := "internal server error"

	if appErr, ok := apperror.AsAppError(err); ok {
		kind = appErr.Kind
		message = appErr.Message
	}

	statusCode := mapErrorKindToHTTP(kind)

	// Full error chain goes to the logs, never to the client.
	level := slog.LevelWarn
	if statusCode >= http.StatusInternalServerError {
		level = slog.LevelError
	}

	slog.Log(c.Request.Context(), level, "request failed",
		"method", c.Request.Method,
		"path", c.FullPath(),
		"status", statusCode,
		"kind", kind,
		"error", errorChain(err),
	)

	return GinResponder[any]{
		ctx:        c,
		statusCode: statusCode,
		envelope:   response.BuildError(string(kind), message),
	}
}

// errorChain renders the full error for logging. *apperror.Error.Error()
// returns only its client-safe Message, it walk through each *apperror.Error
// and append its Cause. A non-apperror error's Error() already includes
// everything it wraps, so it stops there.
func errorChain(err error) string {
	var parts []string
	for e := err; e != nil; e = errors.Unwrap(e) {
		parts = append(parts, e.Error())
		if _, ok := e.(*apperror.Error); !ok {
			break
		}
	}
	return strings.Join(parts, ": ")
}

func HandleSuccess[T any](c *gin.Context, statusCode int, data T, message string) GinResponder[T] {
	return GinResponder[T]{
		ctx:        c,
		statusCode: statusCode,
		envelope:   response.BuildSuccess(data, message),
	}
}

func HandleMessage(c *gin.Context, statusCode int, message string) GinResponder[any] {
	return GinResponder[any]{
		ctx:        c,
		statusCode: statusCode,
		envelope:   response.BuildMessage(message),
	}
}

func (r GinResponder[T]) WithMeta(meta *response.Meta) GinResponder[T] {
	r.envelope.Meta = meta
	return r
}

func (r GinResponder[T]) Send() {
	if requestID := r.ctx.GetString("X-Request-ID"); requestID != "" {
		if r.envelope.Meta == nil {
			r.envelope.Meta = &response.Meta{}
		}

		r.envelope.Meta.RequestID = requestID
	}

	r.ctx.JSON(r.statusCode, r.envelope)
}

func (r GinResponder[T]) Abort() {
	if requestID := r.ctx.GetString("X-Request-ID"); requestID != "" {
		if r.envelope.Meta == nil {
			r.envelope.Meta = &response.Meta{}
		}

		r.envelope.Meta.RequestID = requestID
	}

	r.ctx.AbortWithStatusJSON(r.statusCode, r.envelope)
}
