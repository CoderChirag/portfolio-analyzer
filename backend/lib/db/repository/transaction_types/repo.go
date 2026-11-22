package transactiontypesrepo

import (
	"coderchirag/folio-tracker/lib/db"
	"coderchirag/folio-tracker/lib/db/repository"
	"coderchirag/folio-tracker/lib/db/repository/internal/dberr"
	"coderchirag/folio-tracker/lib/db/repository/internal/jetgen/model"
	"coderchirag/folio-tracker/lib/db/repository/internal/jetgen/table"
	"coderchirag/folio-tracker/lib/domain"
	"context"
	"fmt"

	mysqldialect "github.com/go-jet/jet/v2/mysql"
)

const (
	logPrefix             = "repo:TransactionType"
	getAllPrefix          = "GetAll"
	queryContextLogPrefix = "QueryContext"
)

type TransactionType struct {
	db db.DBTX
}

func New(db db.DBTX) *TransactionType {
	return &TransactionType{
		db: db,
	}
}

func (r *TransactionType) GetAll(ctx context.Context) ([]domain.TransactionType, error) {
	q := mysqldialect.
		SELECT(table.TransactionTypes.AllColumns).
		FROM(table.TransactionTypes).
		ORDER_BY(table.TransactionTypes.ID)
	repository.LogQuery(ctx, fmt.Sprintf("%s.%s", logPrefix, getAllPrefix), q)

	var rows []model.TransactionTypes
	err := q.QueryContext(ctx, r.db, &rows)
	if err != nil {
		return nil, fmt.Errorf("%s.%s: %s: %w", logPrefix, getAllPrefix, queryContextLogPrefix, dberr.MapError(err))
	}

	out := make([]domain.TransactionType, len(rows))
	for i, r := range rows {
		out[i] = *ToTransactionTypeDomain(&r)
	}

	return out, nil
}
