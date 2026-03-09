package crypto

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TxFn is a function that executes within a DB transaction.
type TxFn func(ctx context.Context, tx pgx.Tx) error

// WithTx executes fn within a DB transaction at ReadCommitted isolation.
//
// Guarantees:
//   - If fn returns an error → transaction is rolled back, error is returned.
//   - If fn panics → deferred rollback is attempted before the panic propagates.
//   - Only pool.Query/pool.Exec methods that accept pgx.Tx should be called inside fn.
//     Passing the pool directly inside fn bypasses the transaction — do not do that.
func WithTx(ctx context.Context, db *pgxpool.Pool, fn TxFn) (retErr error) {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	// Deferred rollback catches both error returns and panics.
	defer func() {
		if p := recover(); p != nil {
			// Attempt rollback — ignore secondary error on panic path.
			_ = tx.Rollback(ctx)
			// Re-panic so the caller's recover (if any) can handle it.
			panic(p)
		}
		if retErr != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	if err := fn(ctx, tx); err != nil {
		return err // deferred rollback runs
	}

	return tx.Commit(ctx)
}
