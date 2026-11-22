package jobsrepo

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
	logPrefix             = "repo:Job"
	createLogPrefix       = "Create"
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
