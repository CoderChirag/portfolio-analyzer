package accountshandler

import (
	"coderchirag/folio-tracker/lib/api/response"
	"errors"
)

var (
	ErrEmailRequired = errors.New("email required in query params.")
)

func classifyErrors(err error) (int, *response.ErrorBody) {
	switch {
	default:
		return response.Classify(err)
	}
}
