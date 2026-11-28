package postgresql

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/terabiome/core/internal/repositories"
)

type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

// check this weird thing later about CHECKPOINT: https://pkg.go.dev/github.com/avito-tech/go-transaction-manager/trm/v2/drivers#SavePointCounter
func (m *TxManager) Wrap(ctx context.Context, fn func(ctx context.Context) error) error {
	// For now, assume there can be 1 transaction at a time. If exist -> reuse.

	var err error
	var (
		tx    pgx.Tx
		txCtx context.Context
	)

	tx, err = loadTxFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("failed to load existing transaction: %w", err)
	} else {
		// original context has [pgx.Tx] injected
		txCtx = ctx
	}

	if tx == nil {
		tx, err = m.pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to start a transaction: %w", err)
		}
		defer tx.Rollback(ctx)

		txCtx, err = withTxInCtx(ctx, tx)
		if err != nil {
			return fmt.Errorf("failed to set tx in context: %w", err)
		}
	}

	if err := fn(txCtx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func loadTxFromCtx(ctx context.Context) (pgx.Tx, error) {
	tx := ctx.Value(repositories.TxContextKey)
	if tx == nil {
		return nil, nil
	}

	pgxTx, ok := tx.(pgx.Tx)
	if !ok {
		return nil, fmt.Errorf("type mismatch: %T (instead of [pgx.Tx])", tx)
	}

	return pgxTx, nil
}

func withTxInCtx(ctx context.Context, tx pgx.Tx) (context.Context, error) {
	if tx == nil {
		return ctx, fmt.Errorf("injecting nil tx, ignored")
	}

	return context.WithValue(ctx, repositories.TxContextKey, tx), nil
}

type PgxPoolWrapper struct {
	pool *pgxpool.Pool
	tx   pgx.Tx
}

func NewPgxPoolWrapper(ctx context.Context, pool *pgxpool.Pool) (*PgxPoolWrapper, error) {
	tx, err := loadTxFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	return &PgxPoolWrapper{pool: pool, tx: tx}, nil
}

func (w *PgxPoolWrapper) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if w.tx != nil {
		return w.tx.Exec(ctx, sql, args...)
	}
	return w.pool.Exec(ctx, sql, args...)
}

func (w *PgxPoolWrapper) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if w.tx != nil {
		return w.tx.Query(ctx, sql, args...)
	}
	return w.pool.Query(ctx, sql, args...)
}

func (w *PgxPoolWrapper) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	if w.tx != nil {
		return w.tx.SendBatch(ctx, b)
	}
	return w.pool.SendBatch(ctx, b)
}

func scanRows(rows pgx.Rows, callback func(pgx.Rows) error) error {
	defer rows.Close()

	if err := callback(rows); err != nil {
		return fmt.Errorf("failed to scan, abort the set: %w", err)
	}

	if rows.Err() != nil {
		return fmt.Errorf("something went wrong when querying: %w", rows.Err())
	}

	return nil
}
