package validation

import (
	"errors"
	"strings"
)

// FieldError is one invalid field. It wraps a sentinel so callers can use errors.Is.
type FieldError struct {
	Field string // JSON field name, e.g. "email"
	Err   error  // sentinel, e.g. models.ErrUserEmailInvalid
}

func (e *FieldError) Error() string {
	return e.Field + ": " + e.Err.Error()
}

func (e *FieldError) Unwrap() error {
	return e.Err
}

// Errors collects every invalid field so the API can report them all at once.
type Errors []*FieldError

func (es Errors) Error() string {
	msgs := make([]string, len(es))
	for i, e := range es {
		msgs[i] = e.Error()
	}
	return "validation failed: " + strings.Join(msgs, "; ")
}

// Unwrap lets errors.Is / errors.As see every field error (Go 1.20+).
func (es Errors) Unwrap() []error {
	out := make([]error, len(es))
	for i, e := range es {
		out[i] = e
	}
	return out
}

// Add records a field error.
func (es *Errors) Add(field string, err error) {
	*es = append(*es, &FieldError{Field: field, Err: err})
}

// Err returns nil when nothing was added, so callers can `return errs.Err()`.
func (es Errors) Err() error {
	if len(es) == 0 {
		return nil
	}
	return es
}

// IsValidation reports whether err is (or wraps) a validation failure.
func IsValidation(err error) bool {
	var es Errors
	var fe *FieldError
	return errors.As(err, &es) || errors.As(err, &fe)
}
