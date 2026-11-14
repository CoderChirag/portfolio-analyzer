package registry

import (
	"coderchirag/folio-tracker/lib/domain"
	"coderchirag/folio-tracker/lib/lookup"
	"context"
	"fmt"
)

type Registry struct {
	dataSourceStore RegistryStore[domain.DataSource]
	dataSource      *lookup.Registry[domain.DataSourceID, domain.DataSourceCode, domain.DataSource]

	assetTypeStore RegistryStore[domain.AssetType]
	assetType      *lookup.Registry[domain.AssetTypeID, domain.AssetTypeCode, domain.AssetType]

	transactionTypeStore RegistryStore[domain.TransactionType]
	transactionType      *lookup.Registry[domain.TransactionTypeID, domain.TransactionTypeCode, domain.TransactionType]
}

type RegistryLoadParams struct {
	DataSourceStore      RegistryStore[domain.DataSource]
	AssetTypeStore       RegistryStore[domain.AssetType]
	TransactionTypeStore RegistryStore[domain.TransactionType]
}

func Load(ctx context.Context, p *RegistryLoadParams) (*Registry, error) {
	datasources, err := p.DataSourceStore.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("registry.Load: lookup datasources: %w", err)
	}
	assetTypes, err := p.AssetTypeStore.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("registry.Load: lookup assettypes: %w", err)
	}
	transactionTypes, err := p.TransactionTypeStore.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("registry.Load: lookup transactiontypes: %w", err)
	}

	ds, err := lookup.New[domain.DataSourceID, domain.DataSourceCode, domain.DataSource](
		"DataSource", datasources, domain.SourceGroww,
	)
	if err != nil {
		return nil, fmt.Errorf("registry.Load: %w", err)
	}
	at, err := lookup.New[domain.AssetTypeID, domain.AssetTypeCode, domain.AssetType](
		"AssetType", assetTypes,
		domain.AssetStock, domain.AssetETF, domain.AssetMF,
	)
	if err != nil {
		return nil, fmt.Errorf("registry.Load: %w", err)
	}
	tt, err := lookup.New[domain.TransactionTypeID, domain.TransactionTypeCode, domain.TransactionType](
		"TransactionType", transactionTypes,
		domain.TxnBuy, domain.TxnSell,
	)
	if err != nil {
		return nil, fmt.Errorf("registry.Load: %w", err)
	}

	return &Registry{
		dataSourceStore: p.DataSourceStore,
		dataSource:      ds,

		assetTypeStore: p.AssetTypeStore,
		assetType:      at,

		transactionTypeStore: p.TransactionTypeStore,
		transactionType:      tt,
	}, nil
}

func (r *Registry) DataSource() *lookup.Registry[domain.DataSourceID, domain.DataSourceCode, domain.DataSource] {
	return r.dataSource
}

func (r *Registry) AssetType() *lookup.Registry[domain.AssetTypeID, domain.AssetTypeCode, domain.AssetType] {
	return r.assetType
}

func (r *Registry) TransactionType() *lookup.Registry[domain.TransactionTypeID, domain.TransactionTypeCode, domain.TransactionType] {
	return r.transactionType
}
