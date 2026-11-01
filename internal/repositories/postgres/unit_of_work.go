package repositories

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/terabiome/terabiome/internal/models"
)

type UnitOfWorkPostgresRepository struct {
	basePostgresRepository
}

func NewUnitOfWorkPostgresRepository(pool *pgxpool.Pool) *UnitOfWorkPostgresRepository {
	base := basePostgresRepository{pool}
	return &UnitOfWorkPostgresRepository{base}
}

// Create creates just an UOW record.
func (r *UnitOfWorkPostgresRepository) Create(ctx context.Context, plan models.CreationPlan) error {
	if err := r.insertUOW(ctx, nil, plan.UOW); err != nil {
		return r.Errorf("failed to create unit-of-work: %w", err)
	}

	return nil
}

// --- Private SQL Helpers ---

// tx can be nil -> use [pgxpool.Pool] directly
func (r *UnitOfWorkPostgresRepository) insertUOW(ctx context.Context, tx pgx.Tx, uow models.UnitOfWork) error {
	var err error

	const q = `INSERT INTO units_of_work (name, status, config_json) VALUES ($1, $2, $3, $4)`
	if tx == nil {
		_, err = r.pool.Exec(ctx, q, uow.Name, uow.Status, uow.ConfigJSON)
	} else {
		_, err = tx.Exec(ctx, q, uow.Name, uow.Status, uow.ConfigJSON)
	}

	return err
}

// Ensure the repository satisfies the interface at compile time.
// var _ [interface] = (*struct)(nil)
