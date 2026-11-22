package domain

import (
	"coderchirag/folio-tracker/lib/apperr"
	"coderchirag/folio-tracker/lib/utils"
	"crypto/sha256"
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// Limits mirror the imports table columns.
const (
	importFileNameMaxLen  = 255  // VARCHAR(255), counted in characters
	importErrorCodeMaxLen = 64   // VARCHAR(64) ascii
	importErrorMsgMaxLen  = 1024 // VARCHAR(1024), counted in characters
)

// ImportKind picks the parser for a file. One data source can export several kinds.
type ImportKind struct {
	datasource DataSourceCode
	assetType  AssetTypeCode
	importKind string
}

func (k ImportKind) DataSource() DataSourceCode {
	return k.datasource
}

func (k ImportKind) AssetType() AssetTypeCode {
	return k.assetType
}

func (k ImportKind) String() string {
	return k.importKind
}

func (k ImportKind) Value() (driver.Value, error) {
	if !k.IsValid() {
		return "", ErrImportKindInvalid
	}
	return k.String(), nil
}

func (k *ImportKind) Scan(src any) error {
	var importKind string
	switch s := src.(type) {
	case string:
		importKind = s
	case []byte:
		importKind = string(s)
	default:
		return fmt.Errorf("domain.ImportKind: cannot scan %T", src)
	}

	switch importKind {
	case ImportGrowwStockOrders.String():
		k.datasource = ImportGrowwStockOrders.datasource
		k.assetType = ImportGrowwStockOrders.assetType
		k.importKind = importKind
	default:
		return ErrImportKindInvalid
	}
	return nil
}

func (k ImportKind) IsValid() bool {
	switch k.importKind {
	case ImportGrowwStockOrders.importKind:
		if k.datasource != SourceGroww || k.assetType != AssetStock {
			return false
		}
		return true
	}
	return false
}

var (
	ImportGrowwStockOrders = &ImportKind{
		datasource: SourceGroww,
		assetType:  AssetStock,
		importKind: fmt.Sprintf("%s_%s_orders",
			string(SourceGroww),
			strings.ToLower(string(AssetStock)),
		),
	}
)

type ImportStatus string

const (
	ImportQueued              ImportStatus = "queued"
	ImportProcessing          ImportStatus = "processing"
	ImportSucceeded           ImportStatus = "succeeded"
	ImportCompletedWithErrors ImportStatus = "completed_with_errors"
	ImportFailed              ImportStatus = "failed"
)

func (s ImportStatus) Value() (driver.Value, error) {
	return string(s), nil
}

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
	ErrImportFileHashInvalid  = errors.New("file hash must be a SHA-256 digest")
	ErrImportStatusInvalid    = errors.New("invalid import status")
	ErrImportRowsInvalid      = errors.New("processed rows cannot exceed total rows")
	ErrImportFinishedInvalid  = errors.New("finish time cannot be before start time")
	ErrImportErrorCodeInvalid = errors.New("error code must be at most 64 characters of A-Z, 0-9 and '_'")
	ErrImportErrorMsgTooLong  = errors.New("error message must be at most 1024 characters")
	ErrDuplicateFile          = errors.New("duplicate file")
)

// Import is one uploaded file and the outcome of processing it.
type Import struct {
	ID            ID
	UserID        ID
	DataSourceID  DataSourceID
	Kind          ImportKind
	FileName      string
	FileSize      uint32
	FileSHA256    []byte
	Status        ImportStatus
	RowsTotal     uint32
	RowsProcessed uint32
	RowsImported  uint32
	RowsDuplicate uint32
	RowsFailed    uint32
	ErrorCode     string
	ErrorMsg      string
	StartedAt     time.Time
	FinishedAt    time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type ImportParams struct {
	UserID       ID
	DataSourceID DataSourceID
	Kind         ImportKind
	FileName     string // as sent by the client; any directory part is dropped
	Content      []byte // size and hash are computed from it, so they can't disagree
}

var errorCodePattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)

// NewImport normalizes and validates input, assigns a fresh ID and starts it as queued.
// On failure it returns apperr.ValidationError listing every invalid field.
func NewImport(p *ImportParams) (*Import, error) {
	if len(p.Content) > MaxImportFileSize {
		return nil, apperr.ErrFileTooLarge
	}

	sum := sha256.Sum256(p.Content)
	imp := &Import{
		ID:           NewID(),
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
	var errs apperr.ValidationError

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
		errs.Add("file", apperr.ErrFileTooLarge)
	}

	if len(i.FileSHA256) != sha256.Size {
		errs.Add("file", ErrImportFileHashInvalid)
	}

	if !i.Status.IsValid() {
		errs.Add("status", ErrImportStatusInvalid)
	}

	if i.RowsTotal != 0 && i.RowsProcessed > i.RowsTotal {
		errs.Add("rowsProcessed", ErrImportRowsInvalid)
	}

	if i.FinishedAt.Before(i.StartedAt) {
		errs.Add("finishedAt", ErrImportFinishedInvalid)
	}

	if i.ErrorCode != "" &&
		(len(i.ErrorCode) > importErrorCodeMaxLen || !errorCodePattern.MatchString(i.ErrorCode)) {
		errs.Add("errorCode", ErrImportErrorCodeInvalid)
	}

	if i.ErrorMsg != "" && utf8.RuneCountInString(i.ErrorMsg) > importErrorMsgMaxLen {
		errs.Add("errorMsg", ErrImportErrorMsgTooLong)
	}

	return errs.Err()
}
