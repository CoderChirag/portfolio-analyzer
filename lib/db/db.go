package db

import (
	"context"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type TimezoneInfo struct {
	SessionTimeZone string    `db:"session_time_zone"`
	Now             time.Time `db:"now"`
	UTCNow          time.Time `db:"utc_now"`
}

func Connect(dsn string, poolSize int) (*sqlx.DB, error) {
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("db connect: %w", err)
	}
	db.SetMaxOpenConns(poolSize)
	db.SetMaxIdleConns(poolSize / 2)
	return db, nil
}

// VerifyUTC checks that the MySQL session runs in UTC, so DB-generated
// timestamps (CURRENT_TIMESTAMP defaults, NOW()) match Go-written ones.
func VerifyUTC(ctx context.Context, db *sqlx.DB) (TimezoneInfo, error) {
	var tz TimezoneInfo
	err := db.GetContext(ctx, &tz, `
					SELECT @@session.time_zone AS session_time_zone,
								 NOW(3)              AS now,
								 UTC_TIMESTAMP(3)    AS utc_now`)
	if err != nil {
		return tz, fmt.Errorf("verify timezone: %w", err)
	}
	if !tz.Now.Equal(tz.UTCNow) {
		return tz, fmt.Errorf(
			"mysql session time_zone is %q: NOW()=%s but UTC_TIMESTAMP()=%s; add time_zone='+00:00' to the DSN",
			tz.SessionTimeZone, tz.Now.Format(time.RFC3339Nano), tz.UTCNow.Format(time.RFC3339Nano))
	}
	return tz, nil
}
