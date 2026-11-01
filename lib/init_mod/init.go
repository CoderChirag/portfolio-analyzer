package init_mod

import (
	"coderchirag/folio-tracker/lib/config"
	"coderchirag/folio-tracker/lib/dbinit"
	"database/sql"
	"fmt"
	"log/slog"
)

func InitApp(cfg *config.Config) error {

	slog.Info("init-app starting", "db", cfg.DbName, "host", cfg.DbHost)

	// Admin connection (no default DB) for database + user/grant creation
	adminDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/", cfg.DbRootUser, cfg.DbRootPassword, cfg.DbHost, cfg.DbPort)
	adminDB, err := sql.Open("mysql", adminDSN)
	if err != nil {
		return fmt.Errorf("open admin connection: %w", err)
	}
	defer adminDB.Close()
	if err := adminDB.Ping(); err != nil {
		return fmt.Errorf("ping admin connection: %w", err)
	}

	if err := dbinit.EnsureDatabase(adminDB, cfg.DbName); err != nil {
		return err
	}
	if err = dbinit.EnsureUserGrants(adminDB, cfg.DbName, cfg.DbFolioUser, cfg.DbFolioPassword); err != nil {
		return err
	}

	// folio user connection scoped to the schema, multistatements for the migrations.
	migrateDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?multiStatements=true&parseTime=true&loc=UTC&time_zone=%%27%%2B00%%3A00%%27",
		cfg.DbFolioUser, cfg.DbFolioPassword, cfg.DbHost, cfg.DbPort, cfg.DbName)
	migrateDB, err := sql.Open("mysql", migrateDSN)
	if err != nil {
		return fmt.Errorf("open migrate connection: %w", err)
	}
	defer migrateDB.Close()
	if err := migrateDB.Ping(); err != nil {
		return fmt.Errorf("ping migrate connection: %w", err)
	}

	return dbinit.Migrate(migrateDB, cfg.DbName, cfg.MigrationsDir)
}
