package handlers

import (
	"coderchirag/folio-tracker/lib/api/response"
)

func classifyErrors(err error) (int, *response.ErrorBody) {
	switch {
	default:
		return response.Classify(err)
	}
}
