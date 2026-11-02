package response

func BuildError(errCode string, message string) Envelope[any] {
	return Envelope[any]{
		Status: "error",
		Error: &APIError{
			Code:    errCode,
			Message: message,
		},
	}
}

func BuildSuccess[T any](data T, message string) Envelope[T] {
	return Envelope[T]{
		Status:  "success",
		Message: message,
		Data:    &data,
	}
}

func BuildMessage(message string) Envelope[any] {
	return Envelope[any]{
		Status:  "success",
		Message: message,
	}
}

func (e Envelope[T]) WithMeta(requestID string) Envelope[T] {
	e.Meta = &Meta{RequestID: requestID}
	return e
}
