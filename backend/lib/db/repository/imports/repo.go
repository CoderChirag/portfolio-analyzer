package importsrepo

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
	logPrefix             = "repo:Import"
	createLogPrefix       = "Create"
	execContextLogPrefix  = "ExecContext"
	queryContextLogPrefix = "QueryContext"
)

var importUniqueKeys = map[string]dberr.UniqueKeyField{
	"uq_imports_user_file": {Field: "file", Reason: domain.ErrDuplicateFile},
}

type Import struct {
	db  db.DBTX
	now db.Clock
}

func New(db db.DBTX, now db.Clock) *Import {
	return &Import{
		db:  db,
		now: now,
	}
}

func (r *Import) WithTx(tx db.DBTX) *Import {
	return &Import{
		db:  tx,
		now: r.now,
	}
}

func (r *Import) Create(ctx context.Context, imp *domain.Import) error {
	if err := imp.Validate(); err != nil {
		return err
	}

	now := r.now()
	row := toImportModel(imp)
	row.CreatedAt, row.UpdatedAt = now, now

	q := table.Imports.INSERT(table.Imports.AllColumns).MODEL(row)
	repository.LogQuery(ctx, fmt.Sprintf("%s.%s", logPrefix, createLogPrefix), q)

	if _, err := q.ExecContext(ctx, r.db); err != nil {
		return fmt.Errorf("%s.%s: %s: %w", logPrefix, createLogPrefix, execContextLogPrefix, dberr.MapWriteError(err, importUniqueKeys))
	}

	imp.CreatedAt, imp.UpdatedAt = now, now
	return nil
}
