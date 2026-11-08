package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TxManager struct {
	pool *pgxpool.Pool
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
