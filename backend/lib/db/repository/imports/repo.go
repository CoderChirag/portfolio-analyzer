package importsrepo

import (
	"coderchirag/folio-tracker/lib/db"
	"coderchirag/folio-tracker/lib/db/repository"
	"coderchirag/folio-tracker/lib/db/repository/internal/dberr"
	"coderchirag/folio-tracker/lib/db/repository/internal/jetgen/model"
	"coderchirag/folio-tracker/lib/db/repository/internal/jetgen/table"
	"coderchirag/folio-tracker/lib/domain"
	"context"
	"fmt"

	mysqldialect "github.com/go-jet/jet/v2/mysql"
)

const (
	logPrefix             = "repo:Import"
	createLogPrefix       = "Create"
	getStatusLogPrefix    = "GetStatus"
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

func (r *Import) GetStatus(ctx context.Context, p *domain.IDAndUserID) (*domain.ImportIDUserAndStatus, error) {
	q := table.Imports.
		SELECT(
			table.Imports.ID,
			table.Imports.UserID,
			table.Imports.Status,
		).
		FROM(table.Imports).
		WHERE(mysqldialect.AND(
			table.Imports.ID.EQ(mysqldialect.Blob(p.ID.Binary())),
			table.Imports.UserID.EQ(mysqldialect.Blob(p.UserID.Binary())),
		))
	repository.LogQuery(ctx, fmt.Sprintf("%s.%s", logPrefix, getStatusLogPrefix), q)

	var out model.Imports
	err := q.QueryContext(ctx, r.db, &out)
	if err != nil {
		return nil, fmt.Errorf("%s.%s: %s: %w", logPrefix, getStatusLogPrefix, queryContextLogPrefix, dberr.MapError(err))
	}

	return toIDUserAndStatusDomain(&out), nil
}
