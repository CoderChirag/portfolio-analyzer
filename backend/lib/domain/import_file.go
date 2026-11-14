package domain

import (
	"bytes"
	"coderchirag/folio-tracker/lib/apperr"
	"coderchirag/folio-tracker/lib/utils"
	"crypto/sha256"
	"errors"
	"regexp"
	"time"
)

/*
*

	MaxImportFileSize is the MEDIUMBLOB limit.

	The value is 16,777,215 bytes (16 MiB − 1), the maximum a MEDIUMBLOB column can hold. It is the hard limit the database itself imposes.

	The upload handler should enforce a lower limit for users, such as 5 MB via http.MaxBytesReader, and return 413 above it.

*
*/
const MaxImportFileSize = 1<<24 - 1

const importContentTypeMaxLen = 255 // VARCHAR(255) ascii

// Media types the server can detect. Which ones a kind accepts is decided by its importer.
type ContentType string

const (
	ContentTypeCSV  = "text/csv"
	ContentTypeXLSX = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
)

func (c ContentType) IsValid() bool {
	switch c {
	case ContentTypeCSV, ContentTypeXLSX:
		return true
	}
	return false
}

var (
	ErrImportFileImportRequired = errors.New("import is required")
	ErrImportFileContentType    = errors.New("content type must be a media type like text/csv")
	ErrImportFileMismatch       = errors.New("file content does not match the import's size and hash")
)

// ImportFile holds an import's uploaded bytes, 1:1 with imports.
// Kept apart so status polls and history lists never read the blob.
type ImportFile struct {
	ImportID    ID
	ContentType string
	Content     []byte
	CreatedAt   time.Time
}

type ImportFileParams struct {
	Import      *Import
	ContentType string // detected by the server from the bytes, not the client's header
	Content     []byte
}

// RFC 6838 type/subtype characters, lowercase.
var mediaTypePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9!#$&^_.+-]*/[a-z0-9][a-z0-9!#$&^_.+-]*$`)

// NewImportFile normalizes the media type and checks the bytes are the ones the import describes.
// On failure it returns apperr.ValidationError listing every invalid field.
func NewImportFile(p ImportFileParams) (*ImportFile, error) {
	if p.Import == nil {
		var errs apperr.ValidationError
		errs.Add("import", ErrImportFileImportRequired)
		return nil, errs.Err()
	}

	f := &ImportFile{
		ImportID:    p.Import.ID,
		ContentType: utils.NormalizeMediaType(p.ContentType),
		Content:     p.Content,
	}

	if err := f.Validate(p.Import); err != nil {
		return nil, err
	}

	return f, nil
}

// Validate mirrors the column limits and ties the bytes to the import's size and hash.
func (f *ImportFile) Validate(imp *Import) error {
	var errs apperr.ValidationError

	if f.ImportID.IsZero() || imp == nil || imp.ID != f.ImportID {
		errs.Add("import", ErrImportFileImportRequired)
	}

	if len(f.ContentType) > importContentTypeMaxLen || !mediaTypePattern.MatchString(f.ContentType) {
		errs.Add("contentType", ErrImportFileContentType)
	}

	if imp != nil {
		sum := sha256.Sum256(f.Content)
		if len(f.Content) != int(imp.FileSize) || !bytes.Equal(sum[:], imp.FileSHA256) {
			errs.Add("file", ErrImportFileMismatch)
		}
	}

	return errs.Err()
}
