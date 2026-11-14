package api

import (
	"coderchirag/folio-tracker/lib/api/handlers"
	"coderchirag/folio-tracker/lib/api/response"
	"coderchirag/folio-tracker/lib/config"
	"coderchirag/folio-tracker/lib/db"
	"coderchirag/folio-tracker/lib/db/repository"
	"coderchirag/folio-tracker/lib/registry"
	"coderchirag/folio-tracker/lib/service/accounts"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func Main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	})))
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		fatal("config parsing failed", err)
	}

	sqlDB := ensureDB(cfg)
	defer sqlDB.Close()
	verifyDB(sqlDB)

	_, err = registry.Load(ctx, sqlDB) // To be used later when Handler is implemented
	if err != nil {
		fatal("registry load", err)
	}
	users := repository.NewUserRepo(sqlDB, db.SystemClock)
	accountsHandler := handlers.NewAccountsHandler(accounts.NewService(users))

	r := chi.NewRouter()
	r.Use(
		middleware.RequestID,
		middleware.Logger,
		middleware.Recoverer,
	)

	r.Post("/signup", accountsHandler.SignUp)
	r.Get("/users/{id}", accountsHandler.GetUserByID)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		response.JSONError(w, r, response.ErrNotFound, nil)
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		response.JSONError(w, r, response.ErrMethodNotAllowed, nil)
	})

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		slog.Info("server listening", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fatal("server error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down server")
	shutCtx, sc := context.WithTimeout(context.Background(), 30*time.Second)
	defer sc()
	if err := srv.Shutdown(shutCtx); err != nil {
		slog.Error("shutdown error", "err", err)
	}
}

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

func fatal(msg string, err error) {
	slog.Error(msg, "err", err)
	os.Exit(1)
}
