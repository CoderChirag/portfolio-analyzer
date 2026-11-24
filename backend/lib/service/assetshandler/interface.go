package assetshandler

import (
	"coderchirag/folio-tracker/lib/domain"
	"context"
)

type Instrument interface {
	AssetType() []domain.AssetTypeCode
}

type ImportRepo interface {
	Create(ctx context.Context, imp *domain.Import) error
	GetStatus(ctx context.Context, p *domain.IDAndUserID) (*domain.ImportIDUserAndStatus, error)
}

type ImportFileRepo interface {
	Create(ctx context.Context, imp *domain.Import, impFile *domain.ImportFile) error
}

type JobRepo interface {
	Create(ctx context.Context, job *domain.Job) error
	GetStatus(ctx context.Context, p *domain.IDAndUserID) (*domain.JobIDUserAndStatus, error)
}

type Transactor interface {
	// WithinTx runs fn in one transaction: it commits if fn returns nil and rolls back otherwise.
	// fn may run more than once (retried on deadlock), so it must contain only database work.
	WithinTx(ctx context.Context, fn func(ctx context.Context, R *TxRepos) error, txName string) error
}
