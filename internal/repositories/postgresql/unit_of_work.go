package postgresql

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/terabiome/core/internal/models"
	pgdb "github.com/terabiome/core/pkg/database/postgresql"
	"github.com/terabiome/core/pkg/database/postgresql/txmanager"
	querybuilder "github.com/terabiome/core/pkg/querybuilder/postgresql/v2"
)

type UnitOfWorkPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewUnitOfWorkPostgresRepository(pool *pgxpool.Pool) *UnitOfWorkPostgresRepository {
	return &UnitOfWorkPostgresRepository{pool: pool}
}

// Create creates just an UOW record.
func (r *UnitOfWorkPostgresRepository) Create(ctx context.Context, uow models.UnitOfWork) error {
	q, _, err := querybuilder.
		Insert(querybuilder.StringExpr("units_of_work")).
		Columns("name", "config").
		Build()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	db, err := txmanager.NewPgxPoolWrapper(ctx, r.pool)
	if err != nil {
		return fmt.Errorf("failed to load transaction from ctx: %w", err)
	}

	if _, err := db.Exec(ctx, q, uow.Name, uow.Config); err != nil {
		return fmt.Errorf("failed to create unit-of-work: %w", err)
	}

	return nil
}

func (r *UnitOfWorkPostgresRepository) UpdateByUUID(ctx context.Context, uow models.UnitOfWork) error {
	q, args, err := querybuilder.
		Update(querybuilder.TableExpr("units_of_work")).
		Set("name", querybuilder.Val{V: uow.Name}).
		Set("config", querybuilder.Val{V: uow.Config}).
		Set("staged_at", querybuilder.Val{V: uow.StagedAt}).
		Where(querybuilder.EqualCheckExpr(querybuilder.ColumnExpr("uuid"), querybuilder.Val{V: uow.UUID})).
		Build()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	db, err := txmanager.NewPgxPoolWrapper(ctx, r.pool)
	if err != nil {
		return fmt.Errorf("failed to load transaction from ctx: %w", err)
	}

	if _, err := db.Exec(ctx, q, args...); err != nil {
		return fmt.Errorf("failed to update unit-of-work: %w", err)
	}

	return nil
}

func (r *UnitOfWorkPostgresRepository) FilterByUnstaged(ctx context.Context, limit int) ([]models.UnitOfWork, error) {
	q, _ := querybuilder.
		Select(
			querybuilder.StringExpr("uuid"),
			querybuilder.StringExpr("name"),
			querybuilder.StringExpr("config"),
			querybuilder.StringExpr("created_at"),
			querybuilder.StringExpr("staged_at"),
		).
		FromExpr(querybuilder.TableWithAliasExpr("units_of_work", "uow")).
		Where(querybuilder.NullCheckExpr(querybuilder.StringExpr("staged_at"))).
		Lock(querybuilder.RowLockForNoKeyUpdateExpr()).
		OrderBy(querybuilder.DescSortExpr(querybuilder.StringExpr("uuid"))).
		Limit(limit).
		Build()

	db, err := txmanager.NewPgxPoolWrapper(ctx, r.pool)
	if err != nil {
		return nil, fmt.Errorf("failed to load transaction from ctx: %w", err)
	}

	rows, err := db.Query(ctx, q)
	// smh this err is usually ignore, but good to check
	if err != nil {
		return nil, fmt.Errorf("failed to query units_of_work: %w", err)
	}

	// don't care about capacity, but a good heuristic is about 8 coz I like it
	records := make([]models.UnitOfWork, 0, 8)

	fn := func(r pgx.Rows) error {
		for r.Next() {
			var record models.UnitOfWork
			subErr := r.Scan(
				&record.UUID,
				&record.Name,
				&record.Config,
				&record.CreatedAt,
				&record.StagedAt,
			)
			if subErr != nil {
				return subErr
			}

			records = append(records, record)
		}
		return nil
	}
	if err := pgdb.ScanRows(rows, fn); err != nil {
		return nil, fmt.Errorf("failed to scan units_of_work: %w", err)
	}

	return records, nil
}

// Ensure the repository satisfies the interface at compile time.
// var _ [interface] = (*struct)(nil)
