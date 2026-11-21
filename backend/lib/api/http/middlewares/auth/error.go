package httpauthmiddleware

import (
	"errors"
)

var (
	ErrMissingXAuthToken = errors.New("missing x-auth-token")
)
