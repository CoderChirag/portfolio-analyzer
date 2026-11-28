package jobsrepo

import (
	"coderchirag/folio-tracker/lib/db"
	"coderchirag/folio-tracker/lib/db/repository"
	"coderchirag/folio-tracker/lib/db/repository/internal/dberr"
	"coderchirag/folio-tracker/lib/db/repository/internal/jetgen/model"
	"coderchirag/folio-tracker/lib/db/repository/internal/jetgen/table"
	"coderchirag/folio-tracker/lib/domain"
	"coderchirag/folio-tracker/lib/utils"
	"context"
	"fmt"
	"time"

	mysqldialect "github.com/go-jet/jet/v2/mysql"
)

const (
	logPrefix             = "repo:Job"
	createLogPrefix       = "Create"
	getStatusLogPrefix    = "GetStatus"
	claimLogPrefix        = "Claim"
	execFencedLogPrefix   = "execFenced"
	completeLogPrefix     = "Complete"
	releaseLogPrefix      = "Release"
	retryLogPrefix        = "Retry"
	failLogPrefix         = "Fail"
	reapExpiredLogPrefix  = "ReapExpired"
	execContextLogPrefix  = "ExecContext"
	queryContextLogPrefix = "QueryContext"

	maxLastErrorLen = 2048 // matches last_error VARCHAR(2048)
)

// TODO: Do a better handling of dedupe_key duplication error
var jobUniqueKeys = map[string]dberr.UniqueKeyField{
	"uq_jobs_dedupe_key": {Field: "dedupe_key", Reason: domain.ErrDuplicateJob},
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

func (r *Job) ownedBy(j *domain.Job) mysqldialect.BoolExpression {
	return mysqldialect.AND(
		table.Jobs.ID.EQ(mysqldialect.Blob(j.ID.Binary())),
		table.Jobs.Status.EQ(mysqldialect.String(string(domain.JobRunning))),
		table.Jobs.LockedBy.EQ(mysqldialect.String(j.LockedBy)),
		table.Jobs.Attempts.EQ(mysqldialect.Int(int64(j.Attempts))),
	)
}

func (r *Job) expired(types []domain.JobType) mysqldialect.BoolExpression {
	typeExprs := make([]mysqldialect.Expression, len(types))
	for i, t := range types {
		typeExprs[i] = mysqldialect.String(string(t))
	}

	return mysqldialect.AND(
		table.Jobs.Status.EQ(mysqldialect.String(string(domain.JobRunning))),
		table.Jobs.JobType.IN(typeExprs...),
		table.Jobs.AvailableAt.LT_EQ(mysqldialect.NOW(3)),
	)
}

func (r *Job) execFenced(ctx context.Context, log string, q mysqldialect.UpdateStatement) error {
	repository.LogQuery(ctx, fmt.Sprintf("%s: %s.%s", log, logPrefix, execFencedLogPrefix), q)

	res, err := q.ExecContext(ctx, r.db)
	if err != nil {
		return fmt.Errorf("%s: %s.%s: %w", log, logPrefix, execFencedLogPrefix, dberr.MapWriteError(err, jobUniqueKeys))
	}

	if n, err := res.RowsAffected(); err != nil {
		return fmt.Errorf("%s: %s.%s: RowsAffected: %w", log, logPrefix, execFencedLogPrefix, dberr.MapWriteError(err, jobUniqueKeys))
	} else if n == 0 {
		return fmt.Errorf("%s: %s.%s: %w", log, logPrefix, execFencedLogPrefix, domain.ErrLeaseLost)
	}

	return nil
}

// Claim locks and leases up to p.Limit due jobs of p.Type.
// It must run inside a transaction, or the FOR UPDATE locks are released immediately.
func (r *Job) Claim(ctx context.Context, p *domain.JobClaim) ([]*domain.Job, error) {
	q1 := table.Jobs.
		SELECT(table.Jobs.ID).
		FROM(table.Jobs).
		WHERE(mysqldialect.AND(
			table.Jobs.Status.EQ(mysqldialect.String(string(domain.JobQueued))),
			table.Jobs.JobType.EQ(mysqldialect.String(string(p.Type))),
			table.Jobs.AvailableAt.LT_EQ(mysqldialect.NOW(3)),
			table.Jobs.Attempts.LT(table.Jobs.MaxAttempts),
		)).
		ORDER_BY(table.Jobs.AvailableAt).
		LIMIT(int64(p.Limit)).
		FOR(mysqldialect.UPDATE().SKIP_LOCKED())
	repository.LogQuery(ctx, fmt.Sprintf("%s.%s:select", logPrefix, claimLogPrefix), q1)

	var locked []model.Jobs
	if err := q1.QueryContext(ctx, r.db, &locked); err != nil {
		return nil, fmt.Errorf("%s.%s: select: %w", logPrefix, claimLogPrefix, dberr.MapError(err))
	}

	if len(locked) == 0 {
		return []*domain.Job{}, nil
	}

	ids := make([]mysqldialect.Expression, len(locked))
	for i, j := range locked {
		ids[i] = mysqldialect.Blob(j.ID.Binary())
	}

	q2 := table.Jobs.
		UPDATE().
		SET(
			table.Jobs.Status.SET(mysqldialect.String(string(domain.JobRunning))),
			table.Jobs.Attempts.SET(table.Jobs.Attempts.ADD(mysqldialect.Int(1))),
			table.Jobs.LockedBy.SET(mysqldialect.String(p.WorkerID)),
			table.Jobs.AvailableAt.SET(mysqldialect.NOW(3).ADD(
				mysqldialect.INTERVALe(mysqldialect.Int(p.Lease.Microseconds()), mysqldialect.MICROSECOND),
			)),
			table.Jobs.UpdatedAt.SET(mysqldialect.DateTimeT(r.now())),
		).
		WHERE(table.Jobs.ID.IN(ids...))
	repository.LogQuery(ctx, fmt.Sprintf("%s.%s:lease", logPrefix, claimLogPrefix), q2)

	if _, err := q2.ExecContext(ctx, r.db); err != nil {
		return nil, fmt.Errorf("%s.%s: lease: %w", logPrefix, claimLogPrefix, dberr.MapWriteError(err, jobUniqueKeys))
	}

	q3 := table.Jobs.
		SELECT(table.Jobs.AllColumns).
		FROM(table.Jobs).
		WHERE(table.Jobs.ID.IN(ids...)).
		ORDER_BY(table.Jobs.AvailableAt)
	repository.LogQuery(ctx, fmt.Sprintf("%s.%s:reload", logPrefix, claimLogPrefix), q3)

	var rows []model.Jobs
	if err := q3.QueryContext(ctx, r.db, &rows); err != nil {
		return nil, fmt.Errorf("%s.%s: reload: %w", logPrefix, claimLogPrefix, dberr.MapError(err))
	}

	out := make([]*domain.Job, len(rows))
	for i := range rows {
		out[i] = toJobDomain(&rows[i])
	}
	return out, nil
}

// Complete marks the job succeeded. Call it inside the handler's transaction, as its first statement.
func (r *Job) Complete(ctx context.Context, j *domain.Job) error {
	// `locked_by` records which worker finished the job, which is useful for debugging.
	// `locked_by` together with `finished_at`` points straight to that pod's logs.
	// Hence we don't clear `locked_by` on terminal states
	return r.execFenced(
		ctx,
		fmt.Sprintf("%s.%s", logPrefix, completeLogPrefix),
		table.Jobs.
			UPDATE().
			SET(
				table.Jobs.Status.SET(mysqldialect.String(string(domain.JobSucceeded))),
				table.Jobs.FinishedAt.SET(mysqldialect.NOW(3)),
				table.Jobs.UpdatedAt.SET(mysqldialect.DateTimeT(r.now())),
			).
			WHERE(r.ownedBy(j)),
	)
}

func (r *Job) Release(ctx context.Context, j *domain.Job) error {
	return r.execFenced(
		ctx,
		fmt.Sprintf("%s.%s", logPrefix, releaseLogPrefix),
		table.Jobs.
			UPDATE().
			SET(
				table.Jobs.Status.SET(mysqldialect.String(string(domain.JobQueued))),
				table.Jobs.LockedBy.SET(mysqldialect.StringExp(mysqldialect.NULL)),
				// ownedBy requires attempts = j.Attempts, which is >= 1 after a claim,
				// so this can't underflow the UNSIGNED column
				table.Jobs.Attempts.SET(table.Jobs.Attempts.SUB(mysqldialect.Int(1))),
				table.Jobs.AvailableAt.SET(mysqldialect.NOW(3)),
				table.Jobs.UpdatedAt.SET(mysqldialect.DateTimeT(r.now())),
			).
			WHERE(r.ownedBy(j)),
	)
}

func (r *Job) ReQueue(ctx context.Context, j *domain.Job, delay time.Duration, cause string) error {
	return r.execFenced(
		ctx,
		fmt.Sprintf("%s.%s", logPrefix, retryLogPrefix),
		table.Jobs.
			UPDATE().
			SET(
				table.Jobs.Status.SET(mysqldialect.String(string(domain.JobQueued))),
				table.Jobs.LockedBy.SET(mysqldialect.StringExp(mysqldialect.NULL)),
				table.Jobs.AvailableAt.SET(mysqldialect.NOW(3).ADD(
					mysqldialect.INTERVALe(
						mysqldialect.Int(delay.Microseconds()),
						mysqldialect.MICROSECOND,
					),
				)),
				table.Jobs.LastError.SET(mysqldialect.String(
					utils.TruncateRunes(cause, maxLastErrorLen),
				)),
				table.Jobs.UpdatedAt.SET(mysqldialect.DateTimeT(r.now())),
			).
			WHERE(r.ownedBy(j)),
	)
}

func (r *Job) Fail(ctx context.Context, j *domain.Job, cause string) error {
	// `locked_by` records which worker finished the job, which is useful for debugging.
	// `locked_by` together with `finished_at`` points straight to that pod's logs.
	// Hence we don't clear `locked_by` on terminal states
	return r.execFenced(
		ctx,
		fmt.Sprintf("%s.%s", logPrefix, failLogPrefix),
		table.Jobs.
			UPDATE().
			SET(
				table.Jobs.Status.SET(mysqldialect.String(string(domain.JobFailed))),
				table.Jobs.FinishedAt.SET(mysqldialect.NOW(3)),
				table.Jobs.LastError.SET(mysqldialect.String(utils.TruncateRunes(cause, maxLastErrorLen))),
				table.Jobs.UpdatedAt.SET(mysqldialect.DateTimeT(r.now())),
			).
			WHERE(r.ownedBy(j)),
	)
}

// ReapExpired locks expired jobs to requeue/fail them
// It must run inside a transaction, or the FOR UPDATE locks are released immediately.
func (r *Job) ReapExpired(ctx context.Context, types []domain.JobType, limit int) (requeued []*domain.Job, failed []*domain.Job, err error) {
	if len(types) == 0 {
		return nil, nil, fmt.Errorf("no job types provided")
	}

	q1 := table.Jobs.
		SELECT(table.Jobs.AllColumns).
		WHERE(r.expired(types)).
		ORDER_BY(table.Jobs.AvailableAt).
		LIMIT(int64(limit)).
		FOR(mysqldialect.UPDATE().SKIP_LOCKED())
	repository.LogQuery(ctx, fmt.Sprintf("%s.%s:select", logPrefix, reapExpiredLogPrefix), q1)

	var locked []model.Jobs
	if err := q1.QueryContext(ctx, r.db, &locked); err != nil {
		return nil, nil, fmt.Errorf("%s.%s: select: %w", logPrefix, reapExpiredLogPrefix, dberr.MapError(err))
	}

	if len(locked) == 0 {
		return []*domain.Job{}, []*domain.Job{}, nil
	}

	requeued = make([]*domain.Job, 0)
	failed = make([]*domain.Job, 0)
	requeueIDs := make([]mysqldialect.Expression, 0)
	failIDs := make([]mysqldialect.Expression, 0)
	for _, j := range locked {
		job := toJobDomain(&j)
		if job.CanRetry() {
			requeued = append(requeued, job)
			requeueIDs = append(requeueIDs, mysqldialect.Blob(job.ID.Binary()))
		} else {
			failIDs = append(failIDs, mysqldialect.Blob(job.ID.Binary()))
		}
	}

	if len(requeueIDs) > 0 {
		now := r.now()
		requeue := table.Jobs.
			UPDATE().
			SET(
				table.Jobs.Status.SET(mysqldialect.String(string(domain.JobQueued))),
				table.Jobs.LastError.SET(mysqldialect.String(domain.ErrLeaseExpired.Error())),
				table.Jobs.LockedBy.SET(mysqldialect.StringExp(mysqldialect.NULL)),
				table.Jobs.UpdatedAt.SET(mysqldialect.DateTimeT(now)),
				// available_at is already in the past, so the job is claimable again immediately
			).
			WHERE(table.Jobs.ID.IN(requeueIDs...))
		repository.LogQuery(ctx, fmt.Sprintf("%s.%s:requeue", logPrefix, reapExpiredLogPrefix), requeue)

		if _, err := requeue.ExecContext(ctx, r.db); err != nil {
			return nil, nil, fmt.Errorf("%s.%s: requeue: %w", logPrefix, reapExpiredLogPrefix, dberr.MapWriteError(err, jobUniqueKeys))
		}

		for _, j := range requeued {
			j.Status = domain.JobQueued
			j.LastError = domain.ErrLeaseExpired.Error()
			j.LockedBy = ""
			j.UpdatedAt = now
		}
	}

	if len(failIDs) > 0 {
		now := r.now()
		fail := table.Jobs.
			UPDATE().
			SET(
				table.Jobs.Status.SET(mysqldialect.String(string(domain.JobFailed))),
				table.Jobs.LastError.SET(mysqldialect.String(domain.ErrLeaseExpired.Error())),
				table.Jobs.FinishedAt.SET(mysqldialect.NOW(3)),
				table.Jobs.UpdatedAt.SET(mysqldialect.DateTimeT(now)),
			).
			WHERE(table.Jobs.ID.IN(failIDs...))
		repository.LogQuery(ctx, fmt.Sprintf("%s.%s:fail", logPrefix, reapExpiredLogPrefix), fail)

		if _, err := fail.ExecContext(ctx, r.db); err != nil {
			return nil, nil, fmt.Errorf("%s.%s: fail: %w", logPrefix, reapExpiredLogPrefix, dberr.MapWriteError(err, jobUniqueKeys))
		}

		failedUpdated := table.Jobs.
			SELECT(table.Jobs.AllColumns).
			WHERE(table.Jobs.ID.IN(failIDs...))
		repository.LogQuery(ctx, fmt.Sprintf("%s.%s:fail_reload", logPrefix, reapExpiredLogPrefix), failedUpdated)

		var rows []model.Jobs
		if err := failedUpdated.QueryContext(ctx, r.db, &rows); err != nil {
			return nil, nil, fmt.Errorf("%s.%s: fail reload: %w", logPrefix, reapExpiredLogPrefix, dberr.MapError(err))
		}

		failed = make([]*domain.Job, len(rows))
		for i, row := range rows {
			failed[i] = toJobDomain(&row)
		}
	}

	return requeued, failed, nil
}
