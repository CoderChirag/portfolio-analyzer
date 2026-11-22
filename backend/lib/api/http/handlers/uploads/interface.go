package httpuploadshandler

import (
	"coderchirag/folio-tracker/lib/domain"
	"context"
)

type AssetsHandlerService interface {
	Import(ctx context.Context, p *domain.ImportParams, contentType domain.ContentType) (*domain.Job, error)
}
