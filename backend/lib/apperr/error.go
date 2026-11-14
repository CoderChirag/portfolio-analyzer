package apperr

import (
	"errors"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrConflict   = errors.New("modified by someone else; reload and retry")
	ErrDuplicate  = errors.New("duplicate")
	ErrReferenced = errors.New("still referenced by other records")
	ErrInvalidRef = errors.New("referenced record does not exist")
	ErrRetryable  = errors.New("transient database error; retry")
)

// DuplicateError is a unique-key violation. errors.Is matches ErrDuplicate
// and, when the key is known, the model sentinel (e.g. domain.ErrUserEmailTaken).
type DuplicateError struct {
	Key    string // index name, e.g. "uq_users_email"
	Field  string // JSON field, e.g. "email" (empty if the key isn't mapped)
	Reason error  // model sentinel (nil if the key isn't mapped)
	Err    error  // original driver error
}

func (e *DuplicateError) Error() string {
	if e.Reason != nil {
		return e.Field + ": " + e.Reason.Error()
	}
	return "duplicate entry for key: " + e.Key
}

func (e *DuplicateError) Unwrap() []error {
	errs := []error{ErrDuplicate, e.Err}
	if e.Reason != nil {
		errs = append(errs, e.Reason)
	}
	return errs
}
