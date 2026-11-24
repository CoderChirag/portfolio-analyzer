package jobsrepo

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
	logPrefix             = "repo:Job"
	createLogPrefix       = "Create"
	getStatusLogPrefix    = "GetStatus"
	execContextLogPrefix  = "ExecContext"
	queryContextLogPrefix = "QueryContext"
)

// TODO: Do a better handling of dedupe_key duplication error
var jobUniqueKeys = map[string]dberr.UniqueKeyField{
	"uq_job_dedupe_key": {Field: "dedupe_key", Reason: domain.ErrDuplicateJob},
}

type Job struct {
	db  db.DBTX
	now db.Clock
}

func New(db db.DBTX, now db.Clock) *Job {
	return &Job{
		db:  db,
		now: now,
	}
}

func (r *Job) WithTx(tx db.DBTX) *Job {
	return &Job{
		db:  tx,
		now: r.now,
	}
}

func (r *Job) Create(ctx context.Context, job *domain.Job) error {
	if err := job.Validate(); err != nil {
		return err
	}

	now := r.now()
	row := toJobModel(job)
	row.CreatedAt, row.UpdatedAt = now, now

	q := table.Jobs.INSERT(table.Jobs.AllColumns).MODEL(row)
	repository.LogQuery(ctx, fmt.Sprintf("%s.%s", logPrefix, createLogPrefix), q)

	if _, err := q.ExecContext(ctx, r.db); err != nil {
		return fmt.Errorf("%s.%s: %s: %w", logPrefix, createLogPrefix, execContextLogPrefix, dberr.MapWriteError(err, jobUniqueKeys))
	}

	job.CreatedAt, job.UpdatedAt = now, now
	return nil
}

func (r *Job) GetStatus(ctx context.Context, p *domain.IDAndUserID) (*domain.JobIDUserAndStatus, error) {
	q := table.Jobs.
		SELECT(
			table.Jobs.ID,
			table.Jobs.UserID,
			table.Jobs.Status,
		).
		FROM(table.Jobs).
		WHERE(mysqldialect.AND(
			table.Jobs.ID.EQ(mysqldialect.Blob(p.ID.Binary())),
			table.Jobs.UserID.EQ(mysqldialect.Blob(p.UserID.Binary())),
		))
	repository.LogQuery(ctx, fmt.Sprintf("%s.%s", logPrefix, getStatusLogPrefix), q)

	var out model.Jobs
	err := q.QueryContext(ctx, r.db, &out)
	if err != nil {
		return nil, fmt.Errorf("%s.%s: %s: %w", logPrefix, getStatusLogPrefix, queryContextLogPrefix, dberr.MapError(err))
	}

	return toIDUserAndStatusDomain(&out), nil
}
