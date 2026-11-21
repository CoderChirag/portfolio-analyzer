package httpauthmiddleware

import (
	httpresponse "coderchirag/folio-tracker/lib/api/http/response"
	"errors"
	"net/http"
)

var (
	ErrMissingXAuthToken = errors.New("missing x-auth-token")
)

func classifyErrors(err error) (int, *httpresponse.ErrorBody) {
	switch {
	case errors.Is(err, ErrMissingXAuthToken):
		return http.StatusUnauthorized, &httpresponse.ErrorBody{
			Code: "MISSING_X_AUTH_TOKEN",
			Msg:  ErrMissingXAuthToken.Error(),
		}
	default:
		return httpresponse.Classify(err)
	}
}
