package dberr

import (
	"coderchirag/folio-tracker/lib/apperr"
	"database/sql"
	"errors"
	"fmt"
	"regexp"

	"github.com/go-jet/jet/v2/qrm"
	"github.com/go-sql-driver/mysql"
)

const (
	mysqlErrDuplicateEntry  = 1062
	mysqlErrLockWaitTimeout = 1205
	mysqlErrDeadlock        = 1213
	mysqlErrRowIsReferenced = 1451 // delete/update blocked by a child row
	mysqlErrNoReferencedRow = 1452 // insert/update points at a missing parent
)

// UniqueKeyField describes which field and sentinel a unique key represents.
type UniqueKeyField struct {
	Field  string
	Reason error
}

// MySQL 8 reports the key as 'table.key_name'.
var duplicateKeyPattern = regexp.MustCompile(`for key '(?:[^'.]+\.)?([^']+)'`)

// MapError converts driver errors into repository errors. Other errors pass through unchanged.
func MapError(err error) error {
	if errors.Is(err, qrm.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
		return apperr.ErrNotFound
	}

	var me *mysql.MySQLError
	if !errors.As(err, &me) {
		return err
	}
	switch me.Number {
	case mysqlErrDuplicateEntry:
		dup := &apperr.DuplicateError{Err: err}
		if m := duplicateKeyPattern.FindStringSubmatch(me.Message); m != nil {
			dup.Key = m[1]
		}
		return dup
	case mysqlErrRowIsReferenced:
		return fmt.Errorf("%w: %w", apperr.ErrReferenced, err)
	case mysqlErrNoReferencedRow:
		return fmt.Errorf("%w: %w", apperr.ErrInvalidRef, err)
	case mysqlErrDeadlock, mysqlErrLockWaitTimeout:
		return fmt.Errorf("%w: %w", apperr.ErrRetryable, err)
	}

	return err
}

// MapWriteError is mapError plus unique key → field translation for one table.
func MapWriteError(err error, keys map[string]UniqueKeyField) error {
	err = MapError(err)
	var dup *apperr.DuplicateError
	if errors.As(err, &dup) {
		if f, ok := keys[dup.Key]; ok {
			dup.Field, dup.Reason = f.Field, f.Reason
		}
	}
	return err
}
