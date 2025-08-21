package store

import (
	"context"

	lumErrors "lumium/lib/errors"
)

// WithTx begins a transaction, runs fn, and commits/rolls back appropriately.
// Pass a *pgxpool.Pool (it implements Beginner). fn receives the tx as a Queryer.
func WithTx(ctx context.Context, b Beginner, fn func(q Queryer) error) error {
	tx, err := b.Begin(ctx)
	if err != nil {
		return lumErrors.WrapErrorf(err, lumErrors.ErrorCodeDB, "begin tx")
	}
	defer tx.Rollback(ctx) // safe if already committed

	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return lumErrors.WrapErrorf(err, lumErrors.ErrorCodeDB, "commit tx")
	}
	return nil
}
