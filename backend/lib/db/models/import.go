package models

import (
	"coderchirag/folio-tracker/lib/db/models/types"
	"coderchirag/folio-tracker/lib/utils"
	"coderchirag/folio-tracker/lib/validation"
	"crypto/sha256"
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const ImportTable = "imports"

// Limits mirror the imports table columns.
const (
	importFileNameMaxLen  = 255  // VARCHAR(255), counted in characters
	importErrorCodeMaxLen = 64   // VARCHAR(64) ascii
	importErrorMsgMaxLen  = 1024 // VARCHAR(1024), counted in characters
)

// ImportKind picks the parser for a file. One data source can export several kinds.
type ImportKind string

const (
	ImportGrowwStockOrders ImportKind = "groww_stock_orders"
)

func (k ImportKind) IsValid() bool {
	return k == ImportGrowwStockOrders
}

type ImportStatus string

const (
	ImportQueued              ImportStatus = "queued"
	ImportProcessing          ImportStatus = "processing"
	ImportSucceeded           ImportStatus = "succeeded"
	ImportCompletedWithErrors ImportStatus = "completed_with_errors"
	ImportFailed              ImportStatus = "failed"
)

func (s ImportStatus) IsValid() bool {
	switch s {
	case ImportQueued, ImportProcessing, ImportSucceeded, ImportCompletedWithErrors, ImportFailed:
		return true
	}
	return false
}

// IsTerminal reports whether the import has finished; clients stop polling then.
func (s ImportStatus) IsTerminal() bool {
	switch s {
	case ImportCompletedWithErrors, ImportFailed, ImportSucceeded:
		return true
	}
	return false
}

var (
	ErrImportUserRequired     = errors.New("user is required")
	ErrImportSourceRequired   = errors.New("data source is required")
	ErrImportKindInvalid      = errors.New("unsupported import kind")
	ErrImportFileNameRequired = errors.New("file name is required")
	ErrImportFileNameTooLong  = errors.New("file name must be at most 255 characters")
	ErrImportFileNameInvalid  = errors.New("file name contains invalid characters")
	ErrImportFileEmpty        = errors.New("file is empty")
	ErrImportFileTooLarge     = errors.New("file is too large")
	ErrImportFileHashInvalid  = errors.New("file hash must be a SHA-256 digest")
	ErrImportStatusInvalid    = errors.New("invalid import status")
	ErrImportRowsInvalid      = errors.New("processed rows cannot exceed total rows")
	ErrImportFinishedInvalid  = errors.New("finish time cannot be before start time")
	ErrImportErrorCodeInvalid = errors.New("error code must be at most 64 characters of A-Z, 0-9 and '_'")
	ErrImportErrorMsgTooLong  = errors.New("error message must be at most 1024 characters")
)

// Import is one uploaded file and the outcome of processing it.
type Import struct {
	ID            types.ID     `db:"id"              json:"id"`
	UserID        types.ID     `db:"user_id"         json:"-"`
	DataSourceID  DataSourceID `db:"data_source_id"  json:"-"`
	Kind          ImportKind   `db:"kind"            json:"kind"`
	FileName      string       `db:"file_name"       json:"fileName"`
	FileSize      uint32       `db:"file_size"       json:"fileSize"`
	FileSHA256    []byte       `db:"file_sha256"     json:"-"`
	Status        ImportStatus `db:"status"          json:"status"`
	RowsTotal     *uint32      `db:"rows_total"      json:"rowsTotal"` // nil until the file is parsed
	RowsProcessed uint32       `db:"rows_processed"  json:"rowsProcessed"`
	RowsImported  uint32       `db:"rows_imported"   json:"rowsImported"`
	RowsDuplicate uint32       `db:"rows_duplicate"  json:"rowsDuplicate"`
	RowsFailed    uint32       `db:"rows_failed"     json:"rowsFailed"`
	ErrorCode     *string      `db:"error_code"      json:"errorCode,omitempty"`
	ErrorMsg      *string      `db:"error_msg"       json:"erroMsg,omitempty"`
	StartedAt     *time.Time   `db:"started_at"      json:"startedAt,omitempty"`
	FinishedAt    *time.Time   `db:"finished_at"     json:"finishedAt,omitempty"`
	CreatedAt     time.Time    `db:"created_at"      json:"createdAt"`
	UpdatedAt     time.Time    `db:"updated_at"      json:"updatedAt"`
}

type ImportParams struct {
	UserID       types.ID
	DataSourceID DataSourceID
	Kind         ImportKind
	FileName     string // as sent by the client; any directory part is dropped
	Content      []byte // size and hash are computed from it, so they can't disagree
}

var errorCodePattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)

// NewImport normalizes and validates input, assigns a fresh ID and starts it as queued.
// On failure it returns validation.Errors listing every invalid field.
func NewImport(p ImportParams) (*Import, error) {
	if len(p.Content) > MaxImportFileSize {
		var errs validation.Errors
		errs.Add("file", ErrImportFileTooLarge)
		return nil, errs.Err()
	}

	sum := sha256.Sum256(p.Content)
	imp := &Import{
		ID:           types.NewID(),
		UserID:       p.UserID,
		DataSourceID: p.DataSourceID,
		Kind:         p.Kind,
		FileName:     utils.BaseFileName(p.FileName),
		FileSize:     uint32(len(p.Content)),
		FileSHA256:   sum[:],
		Status:       ImportQueued,
	}
	if err := imp.Validate(); err != nil {
		return nil, err
	}

	return imp, nil
}

// Validate mirrors the table's CHECK constraints and column limits.
func (i *Import) Validate() error {
	var errs validation.Errors

	if i.UserID.IsZero() {
		errs.Add("user", ErrImportUserRequired)
	}

	if i.DataSourceID.IsZero() {
		errs.Add("dataSource", ErrImportSourceRequired)
	}

	if !i.Kind.IsValid() {
		errs.Add("kind", ErrImportKindInvalid)
	}

	switch {
	case i.FileName == "":
		errs.Add("fileName", ErrImportFileNameRequired)
	case utf8.RuneCountInString(i.FileName) > importFileNameMaxLen:
		errs.Add("fileName", ErrImportFileNameTooLong)
	case !utf8.ValidString(i.FileName) || strings.IndexFunc(i.FileName, unicode.IsControl) >= 0:
		errs.Add("fileName", ErrImportFileNameInvalid)
	}

	switch {
	case i.FileSize == 0:
		errs.Add("file", ErrImportFileEmpty)
	case i.FileSize > MaxImportFileSize:
		errs.Add("file", ErrImportFileTooLarge)
	}

	if len(i.FileSHA256) != sha256.Size {
		errs.Add("file", ErrImportFileHashInvalid)
	}

	if !i.Status.IsValid() {
		errs.Add("status", ErrImportStatusInvalid)
	}

	if i.RowsTotal != nil && i.RowsProcessed > *i.RowsTotal {
		errs.Add("rowsProcessed", ErrImportRowsInvalid)
	}

	if i.StartedAt != nil && i.FinishedAt != nil && i.FinishedAt.Before(*i.StartedAt) {
		errs.Add("finishedAt", ErrImportFinishedInvalid)
	}

	if i.ErrorCode != nil &&
		(len(*i.ErrorCode) > importErrorCodeMaxLen || !errorCodePattern.MatchString(*i.ErrorCode)) {
		errs.Add("errorCode", ErrImportErrorCodeInvalid)
	}

	if i.ErrorMsg != nil && utf8.RuneCountInString(*i.ErrorMsg) > importErrorMsgMaxLen {
		errs.Add("errorMsg", ErrImportErrorMsgTooLong)
	}

	return errs.Err()
}
