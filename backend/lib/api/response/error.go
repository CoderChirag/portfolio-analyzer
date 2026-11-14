package response

import (
	"coderchirag/folio-tracker/lib/apperr"
	"errors"
	"net/http"
)

var (
	ErrBadRequest       = errors.New("malformed request body")
	ErrEntityTooLarge   = errors.New("entity too large")
	ErrNotFound         = errors.New("not found")
	ErrMethodNotAllowed = errors.New("method not allowed")
)

type fieldError struct {
	Field string `json:"field"`
	Msg   string `json:"msg"`
}

type ErrorBody struct {
	Success bool         `json:"success"`
	Code    string       `json:"code"`
	Msg     string       `json:"msg"`
	Fields  []fieldError `json:"fields,omitempty"`
}

func Classify(err error) (int, *ErrorBody) {
	var verrs apperr.ValidationError
	var dup *apperr.DuplicateError
	switch {
	case errors.Is(err, ErrBadRequest):
		return http.StatusBadRequest, &ErrorBody{
			Code: "BAD_REQUEST",
			Msg:  err.Error(),
		}
	case errors.Is(err, apperr.ErrNotFound):
		return http.StatusNotFound, &ErrorBody{
			Code: "NOT_FOUND",
			Msg:  apperr.ErrNotFound.Error(),
		}
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound, &ErrorBody{
			Code: "NOT_FOUND",
			Msg:  ErrNotFound.Error(),
		}
	case errors.Is(err, ErrMethodNotAllowed):
		return http.StatusMethodNotAllowed, &ErrorBody{
			Code: "METHOD_NOT_ALLOWED",
			Msg:  ErrMethodNotAllowed.Error(),
		}
	case errors.As(err, &dup):
		b := &ErrorBody{Code: "DUPLICATE", Msg: "already exists"}
		if dup.Reason != nil {
			b.Fields = []fieldError{{Field: dup.Field, Msg: dup.Reason.Error()}}
		}
		return http.StatusConflict, b
	case errors.Is(err, apperr.ErrConflict):
		return http.StatusConflict, &ErrorBody{
			Code: "CONFLICT",
			Msg:  apperr.ErrConflict.Error(),
		}
	case errors.Is(err, apperr.ErrReferenced):
		return http.StatusConflict, &ErrorBody{
			Code: "IN_USE",
			Msg:  apperr.ErrReferenced.Error(),
		}
	case errors.Is(err, ErrEntityTooLarge):
		return http.StatusRequestEntityTooLarge, &ErrorBody{
			Code: "ENTITY_TOO_LARGE",
			Msg:  ErrEntityTooLarge.Error(),
		}
	case errors.As(err, &verrs):
		fields := make([]fieldError, 0, len(verrs))
		for _, ve := range verrs {
			fields = append(fields, fieldError{Field: ve.Field, Msg: ve.Err.Error()})
		}
		return http.StatusUnprocessableEntity, &ErrorBody{
			Code:   "VALIDATION_FAILED",
			Msg:    "validation failed",
			Fields: fields,
		}
	case errors.Is(err, apperr.ErrInvalidRef):
		return http.StatusUnprocessableEntity, &ErrorBody{
			Code: "INVALID_REFERENCE",
			Msg:  apperr.ErrInvalidRef.Error(),
		}
	case errors.Is(err, apperr.ErrRetryable):
		return http.StatusServiceUnavailable, &ErrorBody{
			Code: "RETRYABLE",
			Msg:  apperr.ErrRetryable.Error(),
		}
	default:
		return http.StatusInternalServerError, &ErrorBody{
			Code: "INTERNAL",
			Msg:  "internal server error",
		}
	}
}
