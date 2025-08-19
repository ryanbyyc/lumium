package store

import (
	"context"

	lumErrors "lumium/lib/errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Helpers for turning query results into slices of structs
// These centralize pgx.CollectRows usage, ensure consistent error wrapping, and make handlers much smaller

// CollectStructsByName collects all rows into a slice of T, mapping column names to struct fields by name.
// * Column names must match struct field names (case-insensitive), or use SQL aliases to match them exactly
// * The caller is responsible for closing the rows
func CollectStructsByName[T any](rows pgx.Rows) ([]T, error) {
	out, err := pgx.CollectRows(rows, pgx.RowToStructByName[T])
	if err != nil {
		return nil, lumErrors.WrapErrorf(err, lumErrors.ErrorCodeDB, "failed to collect rows")
	}
	return out, nil
}

// QueryAndCollectByName runs a query and returns a slice of T, mapping by column name as with CollectStructsByName
// * Closes rows automagically
// * Wraps errors with DB error code
func QueryAndCollectByName[T any](ctx context.Context, db *pgxpool.Pool, sql string, args ...any) ([]T, error) {
	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, lumErrors.WrapErrorf(err, lumErrors.ErrorCodeDB, "query failed")
	}
	defer rows.Close()
	return CollectStructsByName[T](rows)
}
