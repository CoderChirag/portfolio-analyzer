package api

import (
	httpaccountshandler "coderchirag/folio-tracker/lib/api/http/handlers/accounts"
	"coderchirag/folio-tracker/lib/db"
	assettypes "coderchirag/folio-tracker/lib/db/repository/asset_types"
	datasources "coderchirag/folio-tracker/lib/db/repository/data_sources"
	transactiontypes "coderchirag/folio-tracker/lib/db/repository/transaction_types"
	"coderchirag/folio-tracker/lib/db/repository/users"
	"coderchirag/folio-tracker/lib/registry"
	"coderchirag/folio-tracker/lib/service/accounts"
)

type repoStore struct {
	DataSource      *datasources.DataSource
	AssetType       *assettypes.AssetType
	TransactionType *transactiontypes.TransactionType

	User *users.User
}

type svcStore struct {
	Accounts *accounts.Service
}

type handlerStore struct {
	Accounts *httpaccountshandler.HTTPHandler
}

func getRepos(db db.DBTX, clock db.Clock) *repoStore {
	return &repoStore{
		DataSource:      datasources.New(db),
		AssetType:       assettypes.New(db),
		TransactionType: transactiontypes.New(db),

		User: users.New(db, clock),
	}
}

func getRegistryLoadParams(repos *repoStore) *registry.RegistryLoadParams {
	return &registry.RegistryLoadParams{
		DataSourceStore:      repos.DataSource,
		AssetTypeStore:       repos.AssetType,
		TransactionTypeStore: repos.TransactionType,
	}
}

func getServices(repos *repoStore) *svcStore {
	return &svcStore{
		Accounts: accounts.New(repos.User),
	}
}

func getHandlers(services *svcStore) *handlerStore {
	return &handlerStore{
		Accounts: httpaccountshandler.New(services.Accounts),
	}
}
