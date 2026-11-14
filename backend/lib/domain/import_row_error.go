package domain

import (
	"coderchirag/folio-tracker/lib/db/models/types"
	"coderchirag/folio-tracker/lib/utils"
	"coderchirag/folio-tracker/lib/validation"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

// MaxImportRowErrors caps stored errors per import; rows_failed keeps counting past it.
const MaxImportRowErrors = 1000

// Limits mirror the import_row_errors table columns.
const (
	rowErrorFieldMaxLen   = 64   // VARCHAR(64), counted in characters
	rowErrorCodeMaxLen    = 64   // VARCHAR(64) ascii
	rowErrorMessageMaxLen = 1024 // VARCHAR(1024), counted in characters
)

var (
	ErrRowErrorImportRequired  = errors.New("import is required")
	ErrRowErrorLineInvalid     = errors.New("line number must be at least 1")
	ErrRowErrorFieldTooLong    = errors.New("field must be at most 64 characters")
	ErrRowErrorCodeInvalid     = errors.New("code must be at most 64 characters of A-Z, 0-9 and '_'")
	ErrRowErrorMessageRequired = errors.New("message is required")
	ErrRowErrorMessageTooLong  = errors.New("message must be at most 1024 characters")
	ErrRowErrorRawRowInvalid   = errors.New("raw row must be a JSON object")
)

// ImportRowError records why one row of an imported file was skipped.
// A row can have several, e.g. an unknown ISIN and a bad quantity.
type ImportRowError struct {
	ID         types.ID
	ImportID   types.ID
	LineNumber uint32
	Field      string
	Code       string
	Message    string
	RawRow     types.JSON
	CreatedAt  time.Time
}

type ImportRowErrorParams struct {
	ImportID   types.ID
	LineNumber uint32
	Field      string // file column, e.g. "ISIN"; empty if it concerns the whole row
	Code       string // stable machine code, e.g. "UNKNOWN_ISIN"
	Message    string // cut to the column width rather than rejected
	// Header → value. MySQL's JSON type doesn't keep key order, so the UI
	// should lay columns out by the file's header, not by this object.
	RawRow map[string]string
}

// NewImportRowError normalizes and validates input and assigns a fresh ID.
// On failure it returns validation.Errors listing every invalid field.
func NewImportRowError(p ImportRowErrorParams) (*ImportRowError, error) {
	e := &ImportRowError{
		ID:         types.NewID(),
		ImportID:   p.ImportID,
		LineNumber: p.LineNumber,
		Field:      strings.TrimSpace(p.Field),
		Code:       strings.ToUpper(strings.TrimSpace(p.Code)),
		Message:    utils.TruncateRunes(strings.TrimSpace(p.Message), rowErrorMessageMaxLen),
	}

	if p.RawRow != nil {
		raw, err := json.Marshal(p.RawRow) // a map[string]string always marshals
		if err != nil {
			return nil, err
		}
		e.RawRow = raw
	}
	if err := e.Validate(); err != nil {
		return nil, err
	}

	return e, nil
}

// Validate mirrors the table's column limits.
func (e *ImportRowError) Validate() error {
	var errs validation.Errors

	if e.ImportID.IsZero() {
		errs.Add("import", ErrRowErrorImportRequired)
	}

	if e.LineNumber == 0 {
		errs.Add("lineNumber", ErrRowErrorLineInvalid)
	}

	if e.Field != "" && utf8.RuneCountInString(e.Field) > rowErrorFieldMaxLen {
		errs.Add("field", ErrRowErrorFieldTooLong)
	}

	if len(e.Code) > rowErrorCodeMaxLen || !errorCodePattern.MatchString(e.Code) {
		errs.Add("code", ErrRowErrorCodeInvalid)
	}

	switch n := utf8.RuneCountInString(e.Message); {
	case n == 0:
		errs.Add("message", ErrRowErrorMessageRequired)
	case n > rowErrorMessageMaxLen:
		errs.Add("message", ErrRowErrorMessageTooLong)
	}

	if len(e.RawRow) > 0 && !e.RawRow.IsObject() {
		errs.Add("rawRow", ErrRowErrorRawRowInvalid)
	}

	return errs.Err()
}
