package scheduler

import (
	"coderchirag/folio-tracker/lib/domain"
	"context"
	"time"
)

type Handler interface {
	Handle(ctx context.Context, job *domain.Job) error
	RecordFailures(ctx context.Context, jobs []*domain.Job) error
}

type JobRepo interface {
	Claim(ctx context.Context, p *domain.JobClaim) ([]*domain.Job, error)
	Release(ctx context.Context, j *domain.Job) error
	ReQueue(ctx context.Context, j *domain.Job, delay time.Duration, cause string) error
	Fail(ctx context.Context, j *domain.Job, cause string) error
	ReapExpired(ctx context.Context, types []domain.JobType, limit int) (requeued []*domain.Job, failed []*domain.Job, err error)
}

type Transactor interface {
	// WithinTx runs fn in one transaction: it commits if fn returns nil and rolls back otherwise.
	// fn may run more than once (retried on deadlock), so it must contain only database work.
	WithinTx(ctx context.Context, fn func(ctx context.Context, r *TxRepos) error, txName string) error
}
