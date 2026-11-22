package transactiontypesrepo

import (
	"coderchirag/folio-tracker/lib/db/repository/internal/jetgen/model"
	"coderchirag/folio-tracker/lib/domain"
)

func ToTransactionTypeDomain(m *model.TransactionTypes) *domain.TransactionType {
	return &domain.TransactionType{
		ID:        m.ID,
		Code:      m.Code,
		Name:      m.Name,
		Direction: m.Direction,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
