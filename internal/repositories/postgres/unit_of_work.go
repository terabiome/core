package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/terabiome/terabiome/internal/models"
)

type UnitOfWorkPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewUnitOfWorkPostgresRepository(pool *pgxpool.Pool) *UnitOfWorkPostgresRepository {
	return &UnitOfWorkPostgresRepository{pool: pool}
}

// CreateWithPlan executes a single atomic transaction persisting the entire execution hierarchy.
func (r *UnitOfWorkPostgresRepository) CreateWithPlan(ctx context.Context, plan models.CreationPlan) error {
	// Explicit transaction management
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	// Rollback is a no-op if Commit succeeds
	defer tx.Rollback(ctx)

	// 1. Persist Master UOW
	if err := r.insertUOW(ctx, tx, plan.UOW); err != nil {
		return fmt.Errorf("insert uow: %w", err)
	}

	// 2. Persist Dispatches -> Sequences -> Tasks
	for _, d := range plan.Dispatches {
		if err := r.insertDispatch(ctx, tx, d.Dispatch); err != nil {
			return fmt.Errorf("insert dispatch %s: %w", d.Dispatch.UUID, err)
		}

		for _, s := range d.Sequences {
			if err := r.insertSequence(ctx, tx, s.Sequence); err != nil {
				return fmt.Errorf("insert sequence %s: %w", s.Sequence.UUID, err)
			}

			for _, t := range s.Tasks {
				if err := r.insertTask(ctx, tx, t); err != nil {
					return fmt.Errorf("insert task %s: %w", t.UUID, err)
				}
			}
		}
	}

	// 3. Commit
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// --- Private SQL Helpers ---

func (r *UnitOfWorkPostgresRepository) insertUOW(ctx context.Context, tx pgx.Tx, uow models.UnitOfWork) error {
	const q = `INSERT INTO units_of_work (uuid, name, status, config_json) VALUES ($1, $2, $3, $4)`
	_, err := tx.Exec(ctx, q, uow.UUID, uow.Name, uow.Status, uow.ConfigJSON)
	return err
}

func (r *UnitOfWorkPostgresRepository) insertDispatch(ctx context.Context, tx pgx.Tx, d models.Dispatch) error {
	const q = `INSERT INTO unit_dispatches (uuid, unit_of_work_uuid, status, metadata) VALUES ($1, $2, $3, $4)`
	_, err := tx.Exec(ctx, q, d.UUID, d.UOWUUID, d.Status, d.Metadata)
	return err
}

func (r *UnitOfWorkPostgresRepository) insertSequence(ctx context.Context, tx pgx.Tx, s models.Sequence) error {
	const q = `INSERT INTO dispatched_sequences (uuid, dispatch_uuid, status, metadata) VALUES ($1, $2, $3, $4)`
	_, err := tx.Exec(ctx, q, s.UUID, s.DispatchUUID, s.Status, s.Metadata)
	return err
}

func (r *UnitOfWorkPostgresRepository) insertTask(ctx context.Context, tx pgx.Tx, t models.Task) error {
	const q = `INSERT INTO dispatched_tasks (uuid, sequence_uuid, status, metadata) VALUES ($1, $2, $3, $4)`
	_, err := tx.Exec(ctx, q, t.UUID, t.SequenceUUID, t.Status, t.Metadata)
	return err
}

// Ensure the repository satisfies the interface at compile time.
// var _ [interface] = (*struct)(nil)
