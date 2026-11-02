package gin

import (
	"net/http"

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
	kind := apperror.ErrorKind(err)

	statusCode := mapErrorKindToHTTP(kind)

	return GinResponder[any]{
		ctx:        c,
		statusCode: statusCode,
		envelope:   response.BuildError(string(kind), err.Error()),
	}
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
	if r.envelope.Meta == nil {
		r.envelope.Meta = &response.Meta{}
	}

	if r.envelope.Meta.RequestID == "" {
		r.envelope.Meta.RequestID = r.ctx.GetString("X-Request-ID")
	}

	r.ctx.JSON(r.statusCode, r.envelope)
}
