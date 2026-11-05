package registry

import (
	"coderchirag/folio-tracker/lib/db/models"
	"coderchirag/folio-tracker/lib/lookup"
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Registry struct {
	DataSource      *lookup.Registry[models.DataSourceID, models.DataSourceCode, models.DataSource]
	AssetType       *lookup.Registry[models.AssetTypeID, models.AssetTypeCode, models.AssetType]
	TransactionType *lookup.Registry[models.TransactionTypeID, models.TransactionTypeCode, models.TransactionType]
}

func Load(ctx context.Context, sqlDB *sqlx.DB) (*Registry, error) {
	ds, err := lookup.Load[models.DataSourceID, models.DataSourceCode, models.DataSource](
		ctx, sqlDB, models.DataSourceTable, models.SourceGroww,
	)
	if err != nil {
		return nil, fmt.Errorf("lookup %q: %w", models.DataSourceTable, err)
	}

	at, err := lookup.Load[models.AssetTypeID, models.AssetTypeCode, models.AssetType](
		ctx, sqlDB, models.AssetTypeTable,
		models.AssetStock, models.AssetETF, models.AssetMF,
	)
	if err != nil {
		return nil, fmt.Errorf("lookup %q: %w", models.AssetTypeTable, err)
	}

	tt, err := lookup.Load[models.TransactionTypeID, models.TransactionTypeCode, models.TransactionType](
		ctx, sqlDB, models.TransactionTypeTable,
		models.TxnBuy, models.TxnSell,
	)
	if err != nil {
		return nil, fmt.Errorf("lookup %q: %w", models.TransactionTypeTable, err)
	}

	return &Registry{
		DataSource:      ds,
		AssetType:       at,
		TransactionType: tt,
	}, nil
}
