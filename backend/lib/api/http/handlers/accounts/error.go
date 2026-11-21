package httpaccountshandler

import (
	"errors"
)

var (
	ErrEmailRequired = errors.New("email required in query params.")
)
