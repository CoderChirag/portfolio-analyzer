package assetshandler

import (
	"coderchirag/folio-tracker/lib/domain"
	"context"
	"fmt"
	"slices"
	"time"
)

const (
	logPrefix                = "service:AssetsHandlerService"
	importLogPrefix          = "Import"
	getImportStatusPrefix    = "GetImportStatus"
	getImportStatusLogPrefix = "GetImportStatus"
)

type Service struct {
	instruments []Instrument
	repos       *TxRepos
	transactor  Transactor
}

func New(instruments []Instrument, repos *TxRepos, transactor Transactor) *Service {
	return &Service{
		instruments: instruments,
		repos:       repos,
		transactor:  transactor,
	}
}

func (s *Service) getInstrument(a domain.AssetTypeCode) Instrument {
	for _, i := range s.instruments {
		if slices.Contains(i.AssetType(), a) {
			return i
		}
	}
	return nil
}

func (s *Service) Import(ctx context.Context, p *domain.ImportParams, contentType domain.ContentType) (*domain.ImportJobStatus, error) {
	imp, err := domain.NewImport(p)
	if err != nil {
		return nil, fmt.Errorf("%s.%s: new import: %w", logPrefix, importLogPrefix, err)
	}

	impFile, err := domain.NewImportFile(
		&domain.ImportFileParams{
			Import:      imp,
			ContentType: contentType,
			Content:     p.Content,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%s.%s: new import_file: %w", logPrefix, importLogPrefix, err)
	}

	job, err := domain.NewImportJob(imp, time.Now().UTC().Truncate(time.Millisecond))
	if err != nil {
		// Unwapping error to make every error a 500 here
		return nil, fmt.Errorf("%s.%s: new import_job: %s", logPrefix, importLogPrefix, err.Error())
	}

	if err := s.transactor.WithinTx(ctx, func(ctx context.Context, r *TxRepos) error {
		if err := r.Imports.Create(ctx, imp); err != nil {
			return fmt.Errorf("%s.%s: import create: %w", logPrefix, importLogPrefix, err)
		}
		if err := r.ImportFiles.Create(ctx, imp, impFile); err != nil {
			return fmt.Errorf("%s.%s: import_file create: %w", logPrefix, importLogPrefix, err)
		}
		if err := r.Jobs.Create(ctx, job); err != nil {
			// Unwapping error to make every error a 500 here
			return fmt.Errorf("%s.%s: job create: %s", logPrefix, importLogPrefix, err.Error())
		}
		return nil
	}, fmt.Sprintf("%s_%s", p.Kind.String(), p.UserID)); err != nil {
		return nil, fmt.Errorf("%s.%s: WithinTx: %w", logPrefix, importLogPrefix, err)
	}

	return &domain.ImportJobStatus{
		UserID:       job.UserID,
		JobID:        job.ID,
		ImportID:     imp.ID,
		JobStatus:    job.Status,
		ImportStatus: imp.Status,
	}, nil
}

func (s *Service) GetImportStatus(ctx context.Context, p *domain.GetImportStatusParams) (*domain.ImportJobStatus, error) {
	var job *domain.JobIDUserAndStatus
	var imp *domain.ImportIDUserAndStatus
	var err error
	if err := s.transactor.WithinTx(ctx, func(ctx context.Context, r *TxRepos) error {
		job, err = r.Jobs.GetStatus(ctx, &domain.IDAndUserID{
			UserID: p.UserID,
			ID:     p.JobID,
		})
		if err != nil {
			return fmt.Errorf("%s.%s: get job: %w", logPrefix, getImportStatusLogPrefix, err)
		}

		imp, err = r.Imports.GetStatus(ctx, &domain.IDAndUserID{
			UserID: p.UserID,
			ID:     p.ImportID,
		})
		if err != nil {
			return fmt.Errorf("%s.%s: get import: %w", logPrefix, getImportStatusLogPrefix, err)
		}

		return nil
	}, fmt.Sprintf("%s_%s_%s", "get_import_status", p.UserID, p.JobID)); err != nil {
		return nil, fmt.Errorf("%s.%s: WithinTx: %w", logPrefix, getImportStatusLogPrefix, err)
	}

	return &domain.ImportJobStatus{
		UserID:       imp.UserID,
		JobID:        job.ID,
		ImportID:     imp.ID,
		JobStatus:    job.Status,
		ImportStatus: imp.Status,
	}, nil
}
