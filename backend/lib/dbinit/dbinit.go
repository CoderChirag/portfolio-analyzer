package dbinit

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/file"
)

// identRe restricts database/user names to a safe identifier set, since they
// cannot be passed as bound parameters in CREATE DATABASE / CREATE USER / GRANT
var identRe = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

// EnsureDatabase creates the named database if it doeas not already exist. Idempotent.
func EnsureDatabase(adminDB *sql.DB, name string) error {
	if !identRe.MatchString(name) {
		return fmt.Errorf("invalid database name %q", name)
	}
	var n int
	if err := adminDB.QueryRow(
		"SELECT COUNT(*) FROM information_schema.SCHEMATA WHERE SCHEMA_NAME = ?", name,
	).Scan(&n); err != nil {
		return fmt.Errorf("check database %s: %w", name, err)
	}
	if n > 0 {
		slog.Info("dbinit: database already exists", "db", name)
		return nil
	}
	if _, err := adminDB.Exec("CREATE DATABASE `" + name + "`"); err != nil {
		return fmt.Errorf("create database %s: %w", name, err)
	}
	slog.Info("dbinit: created database", "db", name)
	return nil
}

// EnsureUserGrants creates the application user and grants it ALL PRIVILEGES on
// the schema for hosts localhost, '%', and the server's @@hostname.
func EnsureUserGrants(adminDB *sql.DB, dbName, user, pass string) error {
	if !identRe.MatchString(dbName) {
		return fmt.Errorf("invalid database name %q", dbName)
	}
	if !identRe.MatchString(user) {
		return fmt.Errorf("invalid user name %q", user)
	}
	var hostname string
	if err := adminDB.QueryRow("SELECT @@hostname").Scan(&hostname); err != nil {
		return fmt.Errorf("get hostname: %w", err)
	}
	for _, host := range []string{"localhost", "%", hostname} {
		acct := fmt.Sprintf("'%s'@'%s'", user, host)
		if _, err := adminDB.Exec(fmt.Sprintf(
			"CREATE USER IF NOT EXISTS %s IDENTIFIED BY '%s'", acct, escapeSQLString(pass),
		)); err != nil {
			return fmt.Errorf("create user %s: %w", acct, err)
		}
		if _, err := adminDB.Exec(fmt.Sprintf(
			"GRANT ALL PRIVILEGES ON `%s`.* TO %s", dbName, acct,
		)); err != nil {
			return fmt.Errorf("grant on %s to %s: %w", dbName, acct, err)
		}
	}
	slog.Info("dbinit: ensured user grants", "user", user, "db", dbName)
	return nil
}

// Migrate applies the schema to dbName. db must be connected to dbName with
// multiStatements = true. The detect-and stamp sequence is guarded
// by a MySQL advisory lock so concurrent runs across replicas serialize. Idempotent
func Migrate(db *sql.DB, dbName, migrationsDir string) error {
	ctx := context.Background()

	// Holds an advisory lock for the duration on a single pinned connection so
	// the sequence cannot race across replicas.
	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("acquire lock conn: %w", err)
	}
	defer conn.Close()
	var locked sql.NullInt64
	if err := conn.QueryRowContext(ctx, "SELECT GET_LOCK('folio_init', 120)").Scan(&locked); err != nil {
		return fmt.Errorf("get_lock: %w", err)
	}
	if !locked.Valid || locked.Int64 != 1 {
		return fmt.Errorf("could not acquire folio_init advisory lock")
	}
	defer conn.ExecContext(ctx, "SELECT RELEASE_LOCK('folio_init')") //nolint:errcheck

	absDir, err := filepath.Abs(migrationsDir)
	if err != nil {
		return fmt.Errorf("resolve migrations dir: %w", err)
	}
	src, err := (&file.File{}).Open("file://" + absDir)
	if err != nil {
		return fmt.Errorf("open migration source: %w", err)
	}
	defer src.Close()

	dbDrv, err := mysql.WithInstance(db, &mysql.Config{DatabaseName: dbName})
	if err != nil {
		return fmt.Errorf("migrate driver: %w", err)
	}
	m, err := migrate.NewWithInstance("file", src, dbName, dbDrv)
	if err != nil {
		return fmt.Errorf("migrator: %w", err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("dbinit: schema up to date", "db", dbName)
			return nil
		}
		return fmt.Errorf("migrate up: %w", err)
	}
	slog.Info("dbinit: migrations applied", "db", dbName)
	return nil
}

// escapeSQLString escapes a value for use inside a single-quoted SQL string
// literal (backslash and single quote), for the CREATE USER password which
// cannot be a bound parameter alongside the account identifier.
func escapeSQLString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	return strings.ReplaceAll(s, `'`, `\'`)
}
