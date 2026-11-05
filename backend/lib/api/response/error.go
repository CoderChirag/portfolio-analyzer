package response

import (
	"coderchirag/folio-tracker/lib/db/repository"
	"coderchirag/folio-tracker/lib/validation"
	"errors"
	"net/http"
)

var (
	ErrBadRequest     = errors.New("malformed request body")
	ErrEntityTooLarge = errors.New("entity too large")
)

type fieldError struct {
	Field string `json:"field"`
	Msg   string `json:"msg"`
}

type errorBody struct {
	Success bool         `json:"success"`
	Code    string       `json:"code"`
	Msg     string       `json:"msg"`
	Fields  []fieldError `json:"fields,omitempty"`
}

func classify(err error) (int, *errorBody) {
	var verrs validation.Errors
	var dup *repository.DuplicateError
	switch {
	case errors.Is(err, ErrBadRequest):
		return http.StatusBadRequest, &errorBody{
			Code: "BAD_REQUEST",
			Msg:  err.Error(),
		}
	case errors.Is(err, repository.ErrNotFound):
		return http.StatusNotFound, &errorBody{
			Code: "NOT_FOUND",
			Msg:  repository.ErrNotFound.Error(),
		}
	case errors.As(err, &dup):
		b := &errorBody{Code: "DUPLICATE", Msg: "already exists"}
		if dup.Reason != nil {
			b.Fields = []fieldError{{Field: dup.Field, Msg: dup.Reason.Error()}}
		}
		return http.StatusConflict, b
	case errors.Is(err, repository.ErrConflict):
		return http.StatusConflict, &errorBody{
			Code: "CONFLICT",
			Msg:  repository.ErrConflict.Error(),
		}
	case errors.Is(err, ErrEntityTooLarge):
		return http.StatusRequestEntityTooLarge, &errorBody{
			Code: "ENTITY_TOO_LARGE",
			Msg:  ErrEntityTooLarge.Error(),
		}
	case errors.As(err, &verrs):
		fields := make([]fieldError, 0, len(verrs))
		for _, ve := range verrs {
			fields = append(fields, fieldError{Field: ve.Field, Msg: ve.Err.Error()})
		}
		return http.StatusUnprocessableEntity, &errorBody{
			Code:   "VALIDATION_FAILED",
			Msg:    "validation failed",
			Fields: fields,
		}
	default:
		return http.StatusInternalServerError, &errorBody{
			Code: "INTERNAL",
			Msg:  "internal server error",
		}
	}
}
