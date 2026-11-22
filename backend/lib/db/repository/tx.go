package repository

import (
	"coderchirag/folio-tracker/lib/apperr"
	"coderchirag/folio-tracker/lib/db"
	"coderchirag/folio-tracker/lib/db/repository/internal/dberr"
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
)

type Transactor[R any] struct {
	db         *sqlx.DB
	bind       func(tx db.DBTX) R
	maxRetries int
}

func NewTransactor[R any](db *sqlx.DB, bind func(tx db.DBTX) R, maxRetries int) *Transactor[R] {
	return &Transactor[R]{
		db:         db,
		bind:       bind,
		maxRetries: maxRetries,
	}
}

func (t *Transactor[R]) WithinTx(ctx context.Context, fn func(ctx context.Context, repos R) error, txName string) error {
	var err error
	for attempt := 0; ; attempt++ {
		err = t.once(ctx, fn)
		if !errors.Is(err, apperr.ErrRetryable) || attempt == t.maxRetries {
			return err
		} else if slog.Default().Enabled(ctx, slog.LevelDebug) {
			slog.DebugContext(ctx, "transaction failed", "tx name", txName, "err", err)
		}

		// Back off with jitter before retrying, and stop if ctx is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(attempt+1) * 20 * time.Millisecond):
		}
	}
}

func (t *Transactor[R]) once(ctx context.Context, fn func(ctx context.Context, repos R) error) (err error) {
	tx, err := t.db.BeginTxx(ctx, nil)
	if err != nil {
		return dberr.MapError(err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = fn(ctx, t.bind(tx)); err != nil {
		// No need of wrapping with dberr.MapError() here
		// As all db calls inside fn would be served by repos
		// Which would have wrapped already
		return err
	}
	return dberr.MapError(tx.Commit())
}
