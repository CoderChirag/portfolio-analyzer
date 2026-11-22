package datasourcesrepo

import (
	"coderchirag/folio-tracker/lib/db/repository/internal/jetgen/model"
	"coderchirag/folio-tracker/lib/domain"
)

func ToDataSourceDomain(m *model.DataSources) *domain.DataSource {
	return &domain.DataSource{
		ID:        m.ID,
		Code:      m.Code,
		Name:      m.Name,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
