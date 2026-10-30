package init_mod

import (
	"coderchirag/folio-tracker/lib/dbinit"
	"database/sql"
	"fmt"
	"log/slog"
)

func InitApp() error {
	env, err := newEnvCfg()
	if err != nil {
		return err
	}
	slog.Info("init-app starting", "db", env.dbName, "host", env.dbHost)

	// Admin connection (no default DB) for database + user/grant creation
	adminDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/", env.dbRootUser, env.dbRootPassword, env.dbHost, env.dbPort)
	adminDB, err := sql.Open("mysql", adminDSN)
	if err != nil {
		return fmt.Errorf("open admin connection: %w", err)
	}
	defer adminDB.Close()
	if err := adminDB.Ping(); err != nil {
		return fmt.Errorf("ping admin connection: %w", err)
	}

	if err := dbinit.EnsureDatabase(adminDB, env.dbName); err != nil {
		return err
	}
	if err = dbinit.EnsureUserGrants(adminDB, env.dbName, env.dbFolioUser, env.dbFolioPassword); err != nil {
		return err
	}

	// folio user connection scoped to the schema, multistatements for the migrations.
	migrateDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?multiStatements=true&parseTime=true",
		env.dbFolioUser, env.dbFolioPassword, env.dbHost, env.dbPort, env.dbName)
	migrateDB, err := sql.Open("mysql", migrateDSN)
	if err != nil {
		return fmt.Errorf("open migrate connection: %w", err)
	}
	defer migrateDB.Close()
	if err := migrateDB.Ping(); err != nil {
		return fmt.Errorf("ping migrate connection: %w", err)
	}

	return dbinit.Migrate(migrateDB, env.dbName, env.migrationsDir)
}
