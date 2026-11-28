package txmanager

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txContextKey struct{}

func ContextWithTx(ctx context.Context, tx pgx.Tx) (context.Context, error) {
	if tx == nil {
		return ctx, errors.New("injecting nil tx, ignored")
	}
	return context.WithValue(ctx, txContextKey{}, tx), nil
}

func TxFromContext(ctx context.Context) (pgx.Tx, error) {
	tx := ctx.Value(txContextKey{})
	if tx == nil {
		return nil, nil
	}

	pgxTx, ok := tx.(pgx.Tx)
	if !ok {
		return nil, fmt.Errorf("type mismatch: %T (instead of [pgx.Tx])", tx)
	}
	return pgxTx, nil
}

type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

func (m *TxManager) Wrap(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := TxFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to load existing transaction: %w", err)
	}
	if tx != nil {
		return fn(ctx)
	}

	tx, err = m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start a transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	txCtx, err := ContextWithTx(ctx, tx)
	if err != nil {
		return fmt.Errorf("failed to set tx in context: %w", err)
	}

	if err := fn(txCtx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

type PgxPoolWrapper struct {
	pool *pgxpool.Pool
	tx   pgx.Tx
}

func NewPgxPoolWrapper(ctx context.Context, pool *pgxpool.Pool) (*PgxPoolWrapper, error) {
	tx, err := TxFromContext(ctx)
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
