package importfilesrepo

import (
	"coderchirag/folio-tracker/lib/db"
	"coderchirag/folio-tracker/lib/db/repository"
	"coderchirag/folio-tracker/lib/db/repository/internal/dberr"
	"coderchirag/folio-tracker/lib/db/repository/internal/jetgen/table"
	"coderchirag/folio-tracker/lib/domain"
	"context"
	"fmt"
)

const (
	logPrefix             = "repo:ImportFile"
	createLogPrefix       = "Create"
	execContextLogPrefix  = "ExecContext"
	queryContextLogPrefix = "QueryContext"
)

var importFileUniqueKeys = map[string]dberr.UniqueKeyField{}

type ImportFile struct {
	db  db.DBTX
	now db.Clock
}

func New(db db.DBTX, now db.Clock) *ImportFile {
	return &ImportFile{
		db:  db,
		now: now,
	}
}

func (r *ImportFile) WithTx(tx db.DBTX) *ImportFile {
	return &ImportFile{
		db:  tx,
		now: r.now,
	}
}

func (r *ImportFile) Create(ctx context.Context, imp *domain.Import, impFile *domain.ImportFile) error {
	if err := imp.Validate(); err != nil {
		return err
	}
	if err := impFile.Validate(imp); err != nil {
		return err
	}

	now := r.now()
	row := toImportFileModel(impFile)
	row.CreatedAt = now

	q := table.ImportFiles.INSERT(table.ImportFiles.AllColumns).MODEL(row)
	repository.LogQuery(ctx, fmt.Sprintf("%s.%s", logPrefix, createLogPrefix), q)

	if _, err := q.ExecContext(ctx, r.db); err != nil {
		return fmt.Errorf("%s.%s: %s: %w", logPrefix, createLogPrefix, execContextLogPrefix, dberr.MapWriteError(err, importFileUniqueKeys))
	}

	impFile.CreatedAt = now
	return nil
}
