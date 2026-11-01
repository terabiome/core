package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/terabiome/terabiome/internal/models"
)

func loadTxFromCtx(ctx context.Context) (pgx.Tx, error) {
	tx := ctx.Value(models.RepositoriesTxKey)
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
		return ctx, errors.New("injecting nil tx, ignored")
	}

	return context.WithValue(ctx, models.RepositoriesTxKey, tx), nil
}

// tx can be nil -> use [pgxpool.Pool] directly
// TODO: add a tx helper to start a transaction, but not passing the tx explicitly in parameter
type basePostgresRepository struct {
	pool *pgxpool.Pool
}

func NewBasePostgresRepository(pool *pgxpool.Pool) *basePostgresRepository {
	return &basePostgresRepository{pool: pool}
}

// shortcut for [fmt.Errorf], with prefix or something
func (r basePostgresRepository) Errorf(format string, a ...any) error {
	format = fmt.Sprintf("[repository] %s", format)
	return fmt.Errorf(format, a)
}

func (r *basePostgresRepository) wrappedExec(ctx context.Context, tx pgx.Tx, sql string, arguments ...any) (pgconn.CommandTag, error) {
	if tx == nil {
		return r.pool.Exec(ctx, sql, arguments...)
	}
	return tx.Exec(ctx, sql, arguments...)
}

func (r *basePostgresRepository) wrappedQuery(ctx context.Context, tx pgx.Tx, sql string, arguments ...any) (pgx.Rows, error) {
	if tx == nil {
		return r.pool.Query(ctx, sql, arguments...)
	}
	return tx.Query(ctx, sql, arguments...)
}

// can use pgx.* but i don't like it, rawdogging [pgx.Rows] instead
// no need to call [pgx.Rows.Close] or check [pgx.Rows.Err]
func (r *basePostgresRepository) execScanCallback(tx pgx.Tx, rows pgx.Rows, callback func(pgx.Tx, pgx.Rows) error) error {
	defer rows.Close()

	if err := callback(tx, rows); err != nil {
		return fmt.Errorf("failed to scan, abort the set: %w", err)
	}

	// something is wrong and i don't like it
	if rows.Err() != nil {
		return fmt.Errorf("something went wrong when querying: %w", rows.Err())
	}

	return nil
}

// Ensure the repository satisfies the interface at compile time.
// var _ [interface] = (*struct)(nil)
