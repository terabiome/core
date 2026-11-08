package postgres

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/terabiome/core/internal/repositories"
)

func loadTxFromCtx(ctx context.Context) (pgx.Tx, error) {
	tx := ctx.Value(repositories.TxKey)
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

	return context.WithValue(ctx, repositories.TxKey, tx), nil
}

// trust the SELECT base query
func transformSelectQuery(
	baseQ string,
	queryParameters repositories.QueryParameters,
	lockControlParameters repositories.LockControlParameters,
) (string, error) {
	var q string = baseQ

	re, err := regexp.Compile("^[a-zA-Z][.]{0,1}[a-zA-Z0-9_]*")
	if err != nil {
		return "", fmt.Errorf("invalid [ORDER BY] field pattern: %w", err)
	}

	orderFields := make([]string, 0, len(queryParameters.SortBy))
	for _, sortBy := range queryParameters.SortBy {
		if !re.MatchString(sortBy.Name) {
			err = errors.Join(err, fmt.Errorf("name does not match pattern: %s", sortBy.Name))
			continue
		}

		switch string(sortBy.Order) {
		case string(repositories.SortOrderAscending):
		case string(repositories.SortOrderDescending):
		case "":
			sortBy.Order = repositories.SortOrderAscending
		default:
			err = errors.Join(err, fmt.Errorf("sort order does not match pattern: %s %s", sortBy.Name, sortBy.Order))
			continue
		}

		orderFields = append(orderFields, fmt.Sprintf("%s %s", sortBy.Name, sortBy.Order))
	}
	if err != nil {
		return "", fmt.Errorf("invalid [ORDER BY] field pattern: %w", err)
	}
	if len(orderFields) > 0 {
		q = fmt.Sprintf("%s ORDER BY %s", q, strings.Join(orderFields, ", "))
	}

	if queryParameters.Limit > 0 {
		q = fmt.Sprintf("%s LIMIT %d", q, queryParameters.Limit)
	}
	if queryParameters.Offset > 0 {
		q = fmt.Sprintf("%s OFFSET %d", q, queryParameters.Offset)
	}

	// can move to generic method
	switch lockControlParameters.Type {
	case repositories.LockTypeLockForUpdate:
		q = fmt.Sprintf("%s FOR UPDATE", q)
	case repositories.LockTypeLockForNoKeyUpdate:
		q = fmt.Sprintf("%s FOR NO KEY UPDATE", q)
	case repositories.LockTypeNoLock:
	default:
		return "", fmt.Errorf("invalid lock type: %s", lockControlParameters.Type)
	}

	// default would do nothing, as expected
	switch lockControlParameters.Behavior {
	case repositories.CascadeBehaviorNoWait:
		q = fmt.Sprintf("%s NOWAIT", q)
	case repositories.CascadeBehaviorSkipLocked:
		q = fmt.Sprintf("%s SKIP LOCKED", q)
	}

	return q, nil
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
