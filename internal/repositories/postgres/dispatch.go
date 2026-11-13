package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/terabiome/core/internal/models"
)

type DispatchPostgresRepository struct {
	basePostgresRepository
}

func NewDispatchPostgresRepository(pool *pgxpool.Pool) *DispatchPostgresRepository {
	base := basePostgresRepository{pool}
	return &DispatchPostgresRepository{base}
}

func (r *DispatchPostgresRepository) BulkInsert(ctx context.Context, dispatches []models.Dispatch) error {
	b := InsertQueryBuilder{
		table:   "dispatches",
		columns: []string{"uow_uuid", "status", "target", "payload"},
	}

	q, err := b.Build()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	tx, err := loadTxFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("failed to load transaction from ctx: %w", err)
	}

	batch := pgx.Batch{}
	for _, d := range dispatches {
		batch.Queue(q, d.UOWUUID, d.Status, d.Target, d.Payload)
	}

	if err := r.wrappedSendBatch(ctx, tx, &batch).Close(); err != nil {
		return fmt.Errorf("failed to insert dispatch batch: %w", err)
	}

	return nil
}
