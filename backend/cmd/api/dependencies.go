package api

import (
	httpaccountshandler "coderchirag/folio-tracker/lib/api/http/handlers/accounts"
	httpuploadshandler "coderchirag/folio-tracker/lib/api/http/handlers/uploads"
	"coderchirag/folio-tracker/lib/db"
	"coderchirag/folio-tracker/lib/db/repository"
	assettypes "coderchirag/folio-tracker/lib/db/repository/asset_types"
	assettypesrepo "coderchirag/folio-tracker/lib/db/repository/asset_types"
	datasourcesrepo "coderchirag/folio-tracker/lib/db/repository/data_sources"
	importfilesrepo "coderchirag/folio-tracker/lib/db/repository/import_files"
	importsrepo "coderchirag/folio-tracker/lib/db/repository/imports"
	jobsrepo "coderchirag/folio-tracker/lib/db/repository/jobs"
	transactiontypes "coderchirag/folio-tracker/lib/db/repository/transaction_types"
	transactiontypesrepo "coderchirag/folio-tracker/lib/db/repository/transaction_types"
	usersrepo "coderchirag/folio-tracker/lib/db/repository/users"
	"coderchirag/folio-tracker/lib/registry"
	"coderchirag/folio-tracker/lib/service/accounts"
	"coderchirag/folio-tracker/lib/service/assetshandler"
	"coderchirag/folio-tracker/lib/service/excel"
	"coderchirag/folio-tracker/lib/service/instruments/stocks"
	stocksgroww "coderchirag/folio-tracker/lib/service/instruments/stocks/providers/groww"
	"context"

	"github.com/jmoiron/sqlx"
)

type repoStore struct {
	DataSource      *datasourcesrepo.DataSource
	AssetType       *assettypesrepo.AssetType
	TransactionType *transactiontypesrepo.TransactionType

	User       *usersrepo.User
	Import     *importsrepo.Import
	ImportFile *importfilesrepo.ImportFile
	Job        *jobsrepo.Job
}

type svcStore struct {
	Accounts      *accounts.Service
	AssetsHandler *assetshandler.Service

	Stocks      *stocks.Instrument
	StocksGroww *stocksgroww.Provider

	ExcelParser *excel.Parser
}

type handlerStore struct {
	Accounts *httpaccountshandler.HTTPHandler
	Uploads  *httpuploadshandler.HTTPHandler
}

func getRepos(db db.DBTX, clock db.Clock) *repoStore {
	return &repoStore{
		DataSource:      datasourcesrepo.New(db),
		AssetType:       assettypes.New(db),
		TransactionType: transactiontypes.New(db),

		User:       usersrepo.New(db, clock),
		Import:     importsrepo.New(db, clock),
		ImportFile: importfilesrepo.New(db, clock),
		Job:        jobsrepo.New(db, clock),
	}
}

func getRegistries(ctx context.Context, repos *repoStore) *registry.Registry {
	p := &registry.RegistryLoadParams{
		DataSourceStore:      repos.DataSource,
		AssetTypeStore:       repos.AssetType,
		TransactionTypeStore: repos.TransactionType,
	}
	reg, err := registry.Load(ctx, p) // To be used later when Handler is implemented
	if err != nil {
		fatal("registry load", err)
	}
	return reg
}

func getServices(sqlDB *sqlx.DB, repos *repoStore) *svcStore {
	xl := excel.New()

	stocksGroww := stocksgroww.New(xl)
	stocks := stocks.New([]stocks.Provider{stocksGroww})

	assetsHandler := assetshandler.New(
		[]assetshandler.Instrument{stocks},
		&assetshandler.TxRepos{
			Imports:     repos.Import,
			ImportFiles: repos.ImportFile,
			Jobs:        repos.Job,
		},
		repository.NewTransactor(
			sqlDB,
			func(tx db.DBTX) *assetshandler.TxRepos {
				return &assetshandler.TxRepos{
					Imports:     repos.Import.WithTx(tx),
					ImportFiles: repos.ImportFile.WithTx(tx),
					Jobs:        repos.Job.WithTx(tx),
				}
			},
			5, // maxRetries
		),
	)

	return &svcStore{
		Accounts:      accounts.New(repos.User),
		AssetsHandler: assetsHandler,

		Stocks:      stocks,
		StocksGroww: stocksGroww,

		ExcelParser: xl,
	}
}

func getHandlers(reg *registry.Registry, services *svcStore) *handlerStore {
	return &handlerStore{
		Accounts: httpaccountshandler.New(services.Accounts),
		Uploads:  httpuploadshandler.New(services.AssetsHandler, reg.DataSource()),
	}
}
