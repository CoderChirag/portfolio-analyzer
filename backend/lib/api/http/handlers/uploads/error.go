package httpuploadshandler

import "errors"

var (
	ErrDataSourceRequired = errors.New("datasource required in query params")
	ErrJobIDRequired      = errors.New("job_id required in query params")
	ErrImportIDRequired   = errors.New("import_id required in query params")
	ErrJobIDInvalid       = errors.New("invalid job_id")
	ErrImportIDInvalid    = errors.New("invalid import_id")
)
