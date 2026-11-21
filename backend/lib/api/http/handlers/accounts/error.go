package httpaccountshandler

import (
	httpresponse "coderchirag/folio-tracker/lib/api/http/response"
	"errors"
)

var (
	ErrEmailRequired = errors.New("email required in query params.")
)

func classifyErrors(err error) (int, *httpresponse.ErrorBody) {
	switch {
	default:
		return httpresponse.Classify(err)
	}
}
