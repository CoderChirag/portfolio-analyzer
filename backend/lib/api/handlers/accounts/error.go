package accountshandler

import (
	"coderchirag/folio-tracker/lib/api/response"
	"errors"
	"net/http"
)

var (
	ErrEmailRequired = errors.New("email required in query params.")
	ErrUserForbidden = errors.New("user not allowed to access requested resource")
)

func classifyErrors(err error) (int, *response.ErrorBody) {
	switch {
	case errors.Is(err, ErrUserForbidden):
		return http.StatusForbidden, &response.ErrorBody{
			Code: "USER_FORBIDDEN",
			Msg:  ErrUserForbidden.Error(),
		}
	default:
		return response.Classify(err)
	}
}
