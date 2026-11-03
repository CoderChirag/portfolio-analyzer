package repository

import (
	"database/sql"
	"errors"
	"regexp"

	"github.com/go-sql-driver/mysql"
)

const mysqlErrDuplicateEntry = 1062

var (
	ErrNotFound  = errors.New("not found")
	ErrConflict  = errors.New("modified by someone else; reload and retry")
	ErrDuplicate = errors.New("duplicate")
)

// MySQL 8 reports the key as 'table.key_name'.
var duplicateKeyPattern = regexp.MustCompile(`for key '(?:[^'.]+\.)?([^']+)'`)

// DuplicateError is a unique-key violation. errors.Is matches ErrDuplicate
// and, when the key is known, the model sentinel (e.g. models.ErrUserEmailTaken).
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

// uniqueKeyField describes which field and sentinel a unique key represents.
type uniqueKeyField struct {
	Field  string
	Reason error
}

// mapError converts driver errors into repository errors. Other errors pass through unchanged.
func mapError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}

	var me *mysql.MySQLError
	if errors.As(err, &me) && me.Number == mysqlErrDuplicateEntry {
		dup := &DuplicateError{Err: err}
		if m := duplicateKeyPattern.FindStringSubmatch(me.Message); m != nil && len(m) >= 2 {
			dup.Key = m[1]
		}
		return dup
	}
	return err
}

// mapWriteError is mapError plus unique key → field translation for one table.
func mapWriteError(err error, keys map[string]uniqueKeyField) error {
	err = mapError(err)
	var dup *DuplicateError
	if errors.As(err, &dup) {
		if f, ok := keys[dup.Key]; ok {
			dup.Field, dup.Reason = f.Field, f.Reason
		}
	}
	return err
}
