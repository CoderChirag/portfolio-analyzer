package assetshandler

import (
	"coderchirag/folio-tracker/lib/domain"
	"context"
	"fmt"
	"slices"
	"time"
)

const (
	logPrefix       = "service:AssetsHandlerService"
	importLogPrefix = "Import"
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

func (s *Service) Import(ctx context.Context, p *domain.ImportParams, contentType domain.ContentType) (*domain.Job, error) {
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

	return job, nil
}
