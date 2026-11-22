package httpuploadshandler

import "errors"

var (
	ErrDataSourceRequired = errors.New("datasource required in query params")
)
