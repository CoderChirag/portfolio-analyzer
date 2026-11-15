package auth

import (
	"coderchirag/folio-tracker/lib/api/response"
	"errors"
	"net/http"
)

var (
	ErrMissingXAuthToken = errors.New("missing x-auth-token")
)

func classifyErrors(err error) (int, *response.ErrorBody) {
	switch {
	case errors.Is(err, ErrMissingXAuthToken):
		return http.StatusUnauthorized, &response.ErrorBody{
			Code: "MISSING_X_AUTH_TOKEN",
			Msg:  ErrMissingXAuthToken.Error(),
		}
	default:
		return response.Classify(err)
	}
}
