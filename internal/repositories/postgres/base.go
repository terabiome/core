package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

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

// Ensure the repository satisfies the interface at compile time.
// var _ [interface] = (*struct)(nil)
