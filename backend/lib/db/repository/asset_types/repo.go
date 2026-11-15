package assettypes

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
	logPrefix             = "repo:AssetType"
	getAllPrefix          = "GetAll"
	queryContextLogPrefix = "QueryContext"
)

type AssetType struct {
	db db.DBTX
}

func New(db db.DBTX) *AssetType {
	return &AssetType{
		db: db,
	}
}

func (r *AssetType) GetAll(ctx context.Context) ([]domain.AssetType, error) {
	q := mysqldialect.
		SELECT(table.AssetTypes.AllColumns).
		FROM(table.AssetTypes).
		ORDER_BY(table.AssetTypes.ID)
	repository.LogQuery(ctx, fmt.Sprintf("%s.%s", logPrefix, getAllPrefix), q)

	var rows []model.AssetTypes
	err := q.QueryContext(ctx, r.db, &rows)
	if err != nil {
		return nil, fmt.Errorf("%s.%s: %s: %w", logPrefix, getAllPrefix, queryContextLogPrefix, dberr.MapError(err))
	}

	out := make([]domain.AssetType, len(rows))
	for i, r := range rows {
		out[i] = *ToAssetTypeDomain(&r)
	}

	return out, nil
}
