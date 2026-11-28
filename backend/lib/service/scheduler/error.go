package scheduler

import "errors"

var (
	ErrNoQueueGiven       = errors.New("no queue given, at least 1 queue is required")
	ErrNilQueue           = errors.New("queue cannot be nil")
	ErrDuplicateJobType   = errors.New("duplicate job type for more than one queue")
	ErrNilHandler         = errors.New("handler cannot be nil")
	ErrConcurrencyInvalid = errors.New("invalid concurrency")
	ErrLeaseInvalid       = errors.New("invalid lease interval")
	ErrLeaseMarginInvalid = errors.New("invalid lease margin interval")
	ErrMinRetryInvalid    = errors.New("invalid min retry interval")
	ErrMaxRetryInvalid    = errors.New("invalid max retry interval")

	ErrNilTx                 = errors.New("transactor cannot be nil")
	ErrNilJobRepo            = errors.New("jobs repo cannot be nil")
	ErrWorkerIDInvalid       = errors.New("invalid worker id")
	ErrMinPollInvalid        = errors.New("invalid min poll interval")
	ErrMaxPollInvalid        = errors.New("invalid max poll interval")
	ErrShutdownGraceInvalid  = errors.New("invalid shutdown grace")
	ErrReleaseTimeoutInvalid = errors.New("invalid release timeout")
	ErrReapIntervalInvalid   = errors.New("invalid reap interval")
	ErrReapLimitInvalid      = errors.New("reap limit invalid")
)
