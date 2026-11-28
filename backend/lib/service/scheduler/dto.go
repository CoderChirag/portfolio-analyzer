package scheduler

import (
	"coderchirag/folio-tracker/lib/domain"
	"time"
)

type TxRepos struct {
	Jobs JobRepo
}

type Queue struct {
	Type        domain.JobType
	Handler     Handler
	Concurrency int           // maximum jobs of this type running at once in this process
	Lease       time.Duration // must exceed the handler's worst-case run time, or a heartbeat must extend it
	LeaseMargin time.Duration
	MinRetry    time.Duration
	MaxRetry    time.Duration
}

type SchedulerParams struct {
	Tx             Transactor
	Jobs           JobRepo
	Queues         []*Queue
	WorkerID       string
	MinPoll        time.Duration
	MaxPoll        time.Duration
	ShutdownGrace  time.Duration
	ReleaseTimeout time.Duration
	ReapInterval   time.Duration
	ReapLimit      int
}
