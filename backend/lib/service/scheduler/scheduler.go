package scheduler

import (
	"coderchirag/folio-tracker/lib/apperr"
	"coderchirag/folio-tracker/lib/domain"
	"coderchirag/folio-tracker/lib/utils"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

type Scheduler struct {
	tx             Transactor
	jobs           JobRepo
	queues         []*Queue
	workerID       string
	minPoll        time.Duration
	maxPoll        time.Duration
	shutdownGrace  time.Duration
	releaseTimeout time.Duration
	reapInterval   time.Duration
	reapLimit      int
}

func New(p *SchedulerParams) (*Scheduler, error) {
	s := &Scheduler{
		tx:             p.Tx,
		jobs:           p.Jobs,
		queues:         p.Queues,
		workerID:       p.WorkerID,
		minPoll:        p.MinPoll,
		maxPoll:        p.MaxPoll,
		shutdownGrace:  p.ShutdownGrace,
		releaseTimeout: p.ReleaseTimeout,
		reapInterval:   p.ReapInterval,
		reapLimit:      p.ReapLimit,
	}
	if err := s.Validate(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Scheduler) Validate() error {
	var errs apperr.ValidationError

	if len(s.queues) == 0 {
		errs.Add("Queues", ErrNoQueueGiven)
	}

	m := map[domain.JobType]bool{}
	for i, q := range s.queues {
		if q == nil {
			errs.Add(fmt.Sprintf("queue[%d]", i), ErrNilQueue)
			continue
		}
		if !q.Type.IsValid() {
			errs.Add(fmt.Sprintf("queue[%d].Type", i), fmt.Errorf("%w: %s", domain.ErrJobTypeInvalid, string(q.Type)))
		}
		if m[q.Type] {
			errs.Add(fmt.Sprintf("queue[%d].Type", i), fmt.Errorf("%w: %s", ErrDuplicateJobType, string(q.Type)))
		} else {
			m[q.Type] = true
		}

		if q.Handler == nil {
			errs.Add(fmt.Sprintf("queue[%d].Handler", i), ErrNilHandler)
		}
		if q.Concurrency <= 0 {
			errs.Add(fmt.Sprintf("queue[%d].Concurrency", i), fmt.Errorf("%w: %d", ErrConcurrencyInvalid, q.Concurrency))
		}
		if q.Lease <= 0 {
			errs.Add(fmt.Sprintf("queue[%d].Lease", i), ErrLeaseInvalid)
		}
		if q.LeaseMargin <= 0 || q.Lease < 2*q.LeaseMargin {
			errs.Add(fmt.Sprintf("queue[%d].LeaseMargin", i), ErrLeaseMarginInvalid)
		}
		if q.MinRetry <= 0 {
			errs.Add(fmt.Sprintf("queue[%d].MinRetry", i), ErrMinRetryInvalid)
		}
		if q.MaxRetry < q.MinRetry {
			errs.Add(fmt.Sprintf("queue[%d].MaxRetry", i), ErrMaxRetryInvalid)
		}
	}

	if s.tx == nil {
		errs.Add("Tx", ErrNilTx)
	}
	if s.jobs == nil {
		errs.Add("Jobs", ErrNilJobRepo)
	}
	if s.workerID == "" {
		errs.Add("WorkerID", ErrWorkerIDInvalid)
	}
	if s.minPoll <= 0 {
		errs.Add("MinPoll", ErrMinPollInvalid)
	}
	if s.maxPoll < s.minPoll {
		errs.Add("MaxPoll", ErrMaxPollInvalid)
	}
	if s.shutdownGrace <= 0 {
		errs.Add("ShutdownGrace", ErrShutdownGraceInvalid)
	}
	if s.releaseTimeout <= 0 {
		errs.Add("ReleaseTimeout", ErrReleaseTimeoutInvalid)
	}
	if s.reapInterval <= 0 {
		errs.Add("ReapInterval", ErrReapIntervalInvalid)
	}
	if s.reapLimit <= 0 {
		errs.Add("ReapLimit", ErrReapLimitInvalid)
	}

	return errs.Err()
}

// Run blocks until ctx is cancelled and every in-flight job is cancelled or finished.
func (s *Scheduler) Run(ctx context.Context) {
	jobsCtx, cancelJobs := context.WithCancel(context.WithoutCancel(ctx))
	defer cancelJobs()
	stopGrace := context.AfterFunc(ctx, func() {
		time.AfterFunc(s.shutdownGrace, cancelJobs)
	})
	defer stopGrace()

	var wg sync.WaitGroup
	for _, q := range s.queues {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.runQueue(ctx, jobsCtx, q)
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.reapLoop(ctx, jobsCtx)
	}()

	wg.Wait() // bounded by shutdownGrace + releaseTimeout
}

func (s *Scheduler) runQueue(ctx context.Context, jobsCtx context.Context, q *Queue) {
	slots := make(chan struct{}, q.Concurrency)
	freed := make(chan struct{}, 1) // capacity 1, so repeated signals collapse into one

	var running sync.WaitGroup
	defer running.Wait() // on shutdown, wait for in-flight jobs

	idle := s.minPoll
	for {
		free := q.Concurrency - len(slots)
		if free > 0 {
			claimStart := time.Now()
			jobs, err := s.claim(ctx, q, free)
			if err != nil && ctx.Err() == nil {
				slog.Error("claim failed", "type", q.Type, "err", err)
			}

			for _, job := range jobs {
				slots <- struct{}{} // never blocks: we claimed at most `free` jobs
				running.Add(1)
				go func() {
					defer running.Done()
					defer func() {
						<-slots
						select {
						case freed <- struct{}{}:
						default:
						}
					}()
					s.runJob(jobsCtx, q, job, claimStart)
				}()
			}

			if len(jobs) > 0 {
				idle = s.minPoll // reacts faster to the next job
			}
			if len(jobs) == free {
				continue // a full batch suggests more are waiting
			}

		}

		t := time.NewTimer(utils.Jitter(idle))
		select {
		case <-ctx.Done():
			t.Stop()
			return
		case <-freed:
			// a slot freed up; try to claim right away
		case <-t.C:
			idle = min(idle*2, s.maxPoll)
		}
		t.Stop()
	}
}

func (s *Scheduler) claim(ctx context.Context, q *Queue, n int) ([]*domain.Job, error) {
	var jobs []*domain.Job
	err := s.tx.WithinTx(ctx, func(ctx context.Context, r *TxRepos) error {
		var err error
		jobs, err = r.Jobs.Claim(ctx, &domain.JobClaim{
			WorkerID: s.workerID,
			Type:     q.Type,
			Limit:    n,
			Lease:    q.Lease,
		})
		return err // jobs is reassigned on each attempt, so a transactor retry starts clean
	}, "scheduler.claim")
	return jobs, err
}

func (s *Scheduler) runJob(jobsCtx context.Context, q *Queue, j *domain.Job, claimStart time.Time) {
	jctx, cancel := context.WithDeadline(jobsCtx, claimStart.Add(q.Lease-q.LeaseMargin))
	defer cancel()

	err := s.safeHandle(jctx, q, j)

	// canceled from the parent (shutdown) before any deadline fired
	stopped := errors.Is(jctx.Err(), context.Canceled)
	fctx, fcancel := context.WithTimeout(context.WithoutCancel(jobsCtx), s.releaseTimeout)
	defer fcancel()

	if ferr := s.finish(fctx, q, j, err, stopped); ferr != nil {
		slog.Error(
			"finish failed",
			"type", q.Type,
			"jobID", j.ID.String(),
			"attempt", j.Attempts,
			"err", ferr,
		)
	}
}

func (s *Scheduler) safeHandle(jctx context.Context, q *Queue, j *domain.Job) error {
	err := func() (err error) {
		defer func() {
			if p := recover(); p != nil {
				err = fmt.Errorf("panic: %v", p) // one bad job must not crash the process
				slog.Error("panic during Handle", "type", q.Type, "jobID", j.ID.String(), "stack", string(debug.Stack()))
			}
		}()
		return q.Handler.Handle(jctx, j)
	}()
	return err
}

func (s *Scheduler) finish(fctx context.Context, q *Queue, j *domain.Job, err error, stopped bool) error {
	switch {
	case err == nil:
		return nil // the handler called Complete in its transaction
	case errors.Is(err, domain.ErrLeaseLost):
		slog.Warn(
			"lease lost",
			"type", q.Type,
			"jobID", j.ID.String(),
		)
		return nil
	case stopped:
		// Shutdown interrupted the job before its deadline: requeue it without counting the attempt.
		slog.Info(
			"shutdown event: releasing job",
			"type", q.Type,
			"jobID", j.ID.String(),
		)
		return s.jobs.Release(fctx, j)
	case j.CanRetry() && !errors.Is(err, apperr.ErrPermanent):
		slog.Warn(
			"job failed, re-queueing",
			"type", q.Type,
			"jobID", j.ID.String(),
			"attempt", j.Attempts,
			"err", err,
		)
		return s.jobs.ReQueue(
			fctx,
			j,
			utils.Backoff(q.MinRetry, q.MaxRetry, j.Attempts),
			err.Error(),
		)
	default:
		slog.Error(
			"job failed, marking as permanently failed",
			"type", q.Type,
			"jobID", j.ID.String(),
			"attempt", j.Attempts,
			"err", err,
		)
		if err := s.jobs.Fail(fctx, j, err.Error()); err != nil {
			return err
		}
		return s.recordFailures(fctx, q, []*domain.Job{j})
	}
}

func (s *Scheduler) reapLoop(ctx context.Context, jobsCtx context.Context) {
	t := time.NewTimer(0) // reap immediately at startup, to recover jobs left by a crash
	defer t.Stop()

	types := make([]domain.JobType, len(s.queues))
	for i, q := range s.queues {
		types[i] = q.Type
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}

		var requeued, failed []*domain.Job
		if err := s.tx.WithinTx(ctx, func(ctx context.Context, r *TxRepos) error {
			var err error
			requeued, failed, err = r.Jobs.ReapExpired(ctx, types, s.reapLimit)
			return err
		}, "scheduler.reapLoop"); err != nil {
			slog.Error("reap failed", "type", "ReapLoop", "err", err)
		}

		if len(requeued) > 0 {
			ids := make([]string, len(requeued))
			for i, j := range requeued {
				ids[i] = j.ID.String()
			}
			slog.Info("requeued expired jobs", "type", "ReapLoop", "ids", strings.Join(ids, ", "))
		}
		if len(failed) > 0 {
			ids := make([]string, len(failed))
			m := map[domain.JobType][]*domain.Job{}
			for i, j := range failed {
				ids[i] = j.ID.String()
				m[j.Type] = append(m[j.Type], j)
			}

			slog.Warn("jobs failed after lease expiry on final attempt", "type", "ReapLoop", "ids", strings.Join(ids, ", "))

			var wg sync.WaitGroup
			fctx, fcancel := context.WithTimeout(jobsCtx, s.releaseTimeout)
			for _, q := range s.queues {
				if jobs, exists := m[q.Type]; exists {
					wg.Add(1)
					go func() {
						defer wg.Done()
						if err := s.recordFailures(fctx, q, jobs); err != nil {
							ids := make([]string, len(jobs))
							for i, j := range jobs {
								ids[i] = j.ID.String()
							}
							slog.Error(
								"record failures",
								"type",
								q.Type,
								"ids",
								strings.Join(ids, ", "),
								"err",
								err,
							)
						}
					}()
				}
			}
			wg.Wait()
			fcancel()
		}

		if len(requeued)+len(failed) == s.reapLimit {
			t.Reset(0) //full batch suggests more expired jobs are present
		} else {
			t.Reset(utils.Jitter(s.reapInterval))
		}
	}
}

func (s *Scheduler) recordFailures(fctx context.Context, q *Queue, jobs []*domain.Job) error {
	return func() (err error) {
		defer func() {
			if p := recover(); p != nil {
				err = fmt.Errorf("panic: %v", p)
				slog.Error(
					"panic",
					"type",
					q.Type,
					"method",
					"RecordFailures",
					"stack",
					string(debug.Stack()),
				)
			}
		}()
		return q.Handler.RecordFailures(fctx, jobs)
	}()
}
