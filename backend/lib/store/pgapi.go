package store

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Queryer is implemented by *pgxpool.Pool and pgx.Tx.
// Use this in all repos so they can accept either a pool or a tx.
type Queryer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// Beginner is implemented by *pgxpool.Pool (and anything that can Begin a tx).
type Beginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}
