package assettypes

import (
	"coderchirag/folio-tracker/lib/db/repository/internal/jetgen/model"
	"coderchirag/folio-tracker/lib/domain"
)

func ToAssetTypeDomain(m *model.AssetTypes) *domain.AssetType {
	return &domain.AssetType{
		ID:        m.ID,
		Code:      m.Code,
		Name:      m.Name,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
