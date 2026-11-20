package postgresql

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/terabiome/core/internal/models"
	querybuilder "github.com/terabiome/core/pkg/querybuilder/postgresql"
)

type UnitOfWorkPostgresRepository struct {
	basePostgresRepository
}

func NewUnitOfWorkPostgresRepository(pool *pgxpool.Pool) *UnitOfWorkPostgresRepository {
	base := basePostgresRepository{pool}
	return &UnitOfWorkPostgresRepository{base}
}

// Create creates just an UOW record.
func (r *UnitOfWorkPostgresRepository) Create(ctx context.Context, uow models.UnitOfWork) error {
	q, err := querybuilder.
		Insert(querybuilder.StringExpr("units_of_work")).
		Columns("name", "config").
		Parameterize()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	tx, err := loadTxFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("failed to load transaction from ctx: %w", err)
	}

	if _, err := r.wrappedExec(ctx, tx, q, uow.Name, uow.Config); err != nil {
		return r.Errorf("failed to create unit-of-work: %w", err)
	}

	return nil
}

func (r *UnitOfWorkPostgresRepository) UpdateByUUID(ctx context.Context, uow models.UnitOfWork) error {
	const q = `
		UPDATE units_of_work
		SET
			name = $2,
			config = $3,
			staged_at = $4
		WHERE uuid = $1
	`

	tx, err := loadTxFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("failed to load transaction from ctx: %w", err)
	}

	if _, err := r.wrappedExec(ctx, tx, q, uow.UUID, uow.Name, uow.Config, uow.StagedAt); err != nil {
		return r.Errorf("failed to update unit-of-work: %w", err)
	}

	return nil
}

func (r *UnitOfWorkPostgresRepository) FilterByUnstaged(ctx context.Context, limit int) ([]models.UnitOfWork, error) {
	q, err := querybuilder.
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
		Parameterize()

	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	tx, err := loadTxFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load transaction from ctx: %w", err)
	}

	rows, err := r.wrappedQuery(ctx, tx, q)
	// smh this err is usually ignore, but good to check
	if err != nil {
		return nil, fmt.Errorf("failed to query units_of_work: %w", err)
	}

	// don't care about capacity, but a good heuristic is about 8 coz I like it
	records := make([]models.UnitOfWork, 0, 8)

	fn := func(tx pgx.Tx, r pgx.Rows) error {
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
	if err := r.execScanCallback(tx, rows, fn); err != nil {
		return nil, fmt.Errorf("failed to scan units_of_work: %w", err)
	}

	return records, nil
}

// Ensure the repository satisfies the interface at compile time.
// var _ [interface] = (*struct)(nil)
