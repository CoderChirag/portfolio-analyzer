package main

import (
	"coderchirag/folio-tracker/lib/config"
	"coderchirag/folio-tracker/lib/db"
	"coderchirag/folio-tracker/lib/db/models"
	"coderchirag/folio-tracker/lib/init_mod"
	"coderchirag/folio-tracker/lib/lookup"
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	cfg, err := config.Load()
	if err != nil {
		fatal("config parsing failed", err)
	}

	err = init_mod.InitApp(cfg)
	if err != nil {
		fatal("init-app failed", err)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=UTC&time_zone=%%27%%2B00%%3A00%%27",
		cfg.DbFolioUser, cfg.DbFolioPassword, cfg.DbHost, cfg.DbPort, cfg.DbName)
	sqlDB, err := db.Connect(dsn, cfg.DbPoolSize)
	if err != nil {
		fatal("open db connection failed", err)
	}
	defer sqlDB.Close()
	tz, err := db.VerifyUTC(context.Background(), sqlDB)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("mysql time_zone=%s now=%s utc=%s",
		tz.SessionTimeZone, tz.Now.Format(time.RFC3339Nano), tz.UTCNow.Format(time.RFC3339Nano))
	if err := sqlDB.Ping(); err != nil {
		fatal("ping db connection failed", err)
	}

	_, err = lookup.Load[models.DataSourceID, models.DataSourceCode, models.DataSource](
		context.Background(), sqlDB, models.DataSourceTable, models.SourceGroww,
	)
	if err != nil {
		fatal(fmt.Sprintf("lookup %q failed", models.DataSourceTable), err)
	}
	_, err = lookup.Load[models.AssetTypeID, models.AssetTypeCode, models.AssetType](
		context.Background(), sqlDB, models.AssetTypeTable,
		models.AssetStock, models.AssetETF, models.AssetMF,
	)
	if err != nil {
		fatal(fmt.Sprintf("lookup %q failed", models.AssetTypeTable), err)
	}
	_, err = lookup.Load[models.TransactionTypeID, models.TransactionTypeCode, models.TransactionType](
		context.Background(), sqlDB, models.TransactionTypeTable,
		models.TxnBuy, models.TxnSell,
	)
	if err != nil {
		fatal(fmt.Sprintf("lookup %q failed", models.TransactionTypeTable), err)
	}

}

func fatal(msg string, err error) {
	slog.Error(msg, "err", err)
	os.Exit(1)
}
