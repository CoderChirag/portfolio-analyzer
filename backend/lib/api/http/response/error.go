package httpresponse

import (
	"coderchirag/folio-tracker/lib/apperr"
	"errors"
	"net/http"
)

var (
	ErrBadRequest        = errors.New("malformed request body")
	ErrMultipartParse    = errors.New("error parsing multipart form")
	ErrUnauthorized      = errors.New("user unauthorized")
	ErrMissingXAuthToken = errors.New("missing x-auth-token")
	ErrForbidden         = errors.New("access forbidden")
	ErrMultipartFormFile = errors.New("error getting multipart form file")
	ErrEntityTooLarge    = errors.New("entity too large")
	ErrMethodNotAllowed  = errors.New("method not allowed")
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
	var verrs apperr.ValidationError
	var dup *apperr.DuplicateError
	var mbe *http.MaxBytesError
	switch {
	case errors.Is(err, ErrBadRequest):
		return http.StatusBadRequest, &errorBody{
			Code: "BAD_REQUEST",
			Msg:  ErrBadRequest.Error(),
		}
	case errors.Is(err, ErrMultipartParse):
		return http.StatusBadRequest, &errorBody{
			Code: "BAD_REQUEST",
			Msg:  ErrMultipartParse.Error(),
		}
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized, &errorBody{
			Code: "UNAUTHORIZED",
			Msg:  ErrUnauthorized.Error(),
		}
	case errors.Is(err, ErrMissingXAuthToken):
		return http.StatusUnauthorized, &errorBody{
			Code: "UNAUTHORIZED",
			Msg:  ErrMissingXAuthToken.Error(),
		}
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden, &errorBody{
			Code: "FORBIDDEN",
			Msg:  ErrForbidden.Error(),
		}
	case errors.Is(err, apperr.ErrNotFound):
		return http.StatusNotFound, &errorBody{
			Code: "NOT_FOUND",
			Msg:  apperr.ErrNotFound.Error(),
		}
	case errors.Is(err, ErrMethodNotAllowed):
		return http.StatusMethodNotAllowed, &errorBody{
			Code: "METHOD_NOT_ALLOWED",
			Msg:  ErrMethodNotAllowed.Error(),
		}
	case errors.As(err, &dup):
		b := &errorBody{Code: "DUPLICATE", Msg: "already exists"}
		if dup.Reason != nil {
			b.Fields = []fieldError{{Field: dup.Field, Msg: dup.Reason.Error()}}
		}
		return http.StatusConflict, b
	case errors.Is(err, apperr.ErrConflict):
		return http.StatusConflict, &errorBody{
			Code: "CONFLICT",
			Msg:  apperr.ErrConflict.Error(),
		}
	case errors.Is(err, apperr.ErrReferenced):
		return http.StatusConflict, &errorBody{
			Code: "IN_USE",
			Msg:  apperr.ErrReferenced.Error(),
		}
	case errors.Is(err, ErrEntityTooLarge):
		return http.StatusRequestEntityTooLarge, &errorBody{
			Code: "ENTITY_TOO_LARGE",
			Msg:  ErrEntityTooLarge.Error(),
		}
	case errors.Is(err, apperr.ErrFileTooLarge):
		return http.StatusRequestEntityTooLarge, &errorBody{
			Code: "FILE_TOO_LARGE",
			Msg:  apperr.ErrFileTooLarge.Error(),
		}
	case errors.As(err, &mbe):
		return http.StatusRequestEntityTooLarge, &errorBody{
			Code: "FILE_TOO_LARGE",
			Msg:  apperr.ErrFileTooLarge.Error(),
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
	case errors.Is(err, apperr.ErrInvalidRef):
		return http.StatusUnprocessableEntity, &errorBody{
			Code: "INVALID_REFERENCE",
			Msg:  apperr.ErrInvalidRef.Error(),
		}
	case errors.Is(err, ErrMultipartFormFile):
		return http.StatusUnprocessableEntity, &errorBody{
			Code: "UNPROCESSABLE_ENTITY",
			Msg:  ErrMultipartFormFile.Error(),
		}
	case errors.Is(err, apperr.ErrInvalidFile):
		return http.StatusUnprocessableEntity, &errorBody{
			Code: "INVALID_FILE",
			Msg:  apperr.ErrInvalidFile.Error(),
		}
	case errors.Is(err, apperr.ErrInvalidFileName):
		return http.StatusUnprocessableEntity, &errorBody{
			Code: "INVALID_FILE",
			Msg:  apperr.ErrInvalidFileName.Error(),
		}
	case errors.Is(err, apperr.ErrRetryable):
		return http.StatusServiceUnavailable, &errorBody{
			Code: "RETRYABLE",
			Msg:  apperr.ErrRetryable.Error(),
		}
	default:
		return http.StatusInternalServerError, &errorBody{
			Code: "INTERNAL",
			Msg:  "internal server error",
		}
	}
}
