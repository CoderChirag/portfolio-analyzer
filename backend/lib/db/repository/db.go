package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// DBTX is satisfied by both *sqlx.DB and *sqlx.Tx, so a repository
// works inside or outside a transaction.
type DBTX interface {
	sqlx.ExtContext
	GetContext(ctx context.Context, dest any, query string, args ...any) error
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
}

// WithTx runs fn in a transaction: commit if fn returns nil, otherwise roll back.
// A panic also rolls back, then continues.
func WithTx(ctx context.Context, db *sqlx.DB, fn func(tx *sqlx.Tx) error) (err error) {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
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

	if err = fn(tx); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
