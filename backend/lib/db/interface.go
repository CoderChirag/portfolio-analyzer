package db

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// DBTX is satisfied by both *sqlx.DB and *sqlx.Tx, so a repository
// works inside or outside a transaction.
type DBTX interface {
	sqlx.ExtContext
	GetContext(ctx context.Context, dest any, query string, args ...any) error
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
}
