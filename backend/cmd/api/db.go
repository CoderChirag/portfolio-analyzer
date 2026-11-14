package api

import (
	"coderchirag/folio-tracker/lib/config"
	"coderchirag/folio-tracker/lib/db"
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func ensureDB(cfg *config.Config) *sqlx.DB {
	dsn := (&mysql.Config{
		User:                 cfg.DbFolioUser,
		Passwd:               cfg.DbFolioPassword,
		Net:                  "tcp",
		Addr:                 net.JoinHostPort(cfg.DbHost, cfg.DbPort),
		DBName:               cfg.DbName,
		ParseTime:            true,
		Loc:                  time.UTC,
		Params:               map[string]string{"time_zone": "'+00:00'"},
		AllowNativePasswords: true,
	}).FormatDSN()
	sqlDB, err := db.Connect(dsn, cfg.DbPoolSize)
	if err != nil {
		fatal("open db connection", err)
	}
	return sqlDB
}

func verifyDB(sqlDB *sqlx.DB) {
	tz, err := db.VerifyUTC(context.Background(), sqlDB)
	if err != nil {
		fatal("verify db utc", err)
	}

	slog.Info(fmt.Sprintf("mysql time_zone=%s now=%s utc=%s",
		tz.SessionTimeZone, tz.Now.Format(time.RFC3339Nano), tz.UTCNow.Format(time.RFC3339Nano)))

	if err := sqlDB.Ping(); err != nil {
		fatal("ping db connection", err)
	}
}
