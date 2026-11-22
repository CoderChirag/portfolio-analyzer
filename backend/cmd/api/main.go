package api

import (
	"coderchirag/folio-tracker/lib/config"
	"coderchirag/folio-tracker/lib/db"
	"coderchirag/folio-tracker/lib/db/repository"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
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

	repository.UseSchema("")

	repos := getRepos(sqlDB, db.SystemClock)
	registry := getRegistries(ctx, repos)

	services := getServices(sqlDB, repos)
	handlers := getHandlers(registry, services)

	r := chi.NewRouter()
	registerRoutes(r, *handlers)

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

func fatal(msg string, err error) {
	slog.Error(msg, "err", err)
	os.Exit(1)
}
