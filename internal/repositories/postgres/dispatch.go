package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/terabiome/core/internal/models"
	"github.com/terabiome/core/internal/repositories/postgres/querybuilder"
)

type DispatchPostgresRepository struct {
	basePostgresRepository
}

func NewDispatchPostgresRepository(pool *pgxpool.Pool) *DispatchPostgresRepository {
	base := basePostgresRepository{pool}
	return &DispatchPostgresRepository{base}
}

func (r *DispatchPostgresRepository) BulkInsert(ctx context.Context, dispatches []models.Dispatch) error {
	q, err := querybuilder.
		Insert(querybuilder.StringExpr("dispatches")).
		Columns("uow_uuid", "status", "target", "payload").
		Parameterize()

	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	batch := pgx.Batch{}
	for _, d := range dispatches {
		batch.Queue(q, d.UOWUUID, d.Status, d.Target, d.Payload)
	}

	tx, err := loadTxFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("failed to load transaction from ctx: %w", err)
	}

	if err := r.wrappedSendBatch(ctx, tx, &batch).Close(); err != nil {
		return fmt.Errorf("failed to insert dispatch batch: %w", err)
	}

	return nil
}

func (r *DispatchPostgresRepository) FilterByStatus(ctx context.Context, status models.DispatchStatus, limit int) ([]models.Dispatch, error) {
	q, err := querybuilder.
		Select(
			querybuilder.ColumnExpr("uuid"),
			querybuilder.ColumnExpr("uow_uuid"),
			querybuilder.ColumnExpr("target"),
			querybuilder.ColumnExpr("metadata"),
			querybuilder.ColumnExpr("payload"),
		).
		FromExpr(querybuilder.TableExpr("dispatches")).
		Where(querybuilder.EqualCheckExpr(
			querybuilder.StringExpr("status"),
			querybuilder.SingleQuotedStringExpr(string(status))),
		).
		OrderBy(querybuilder.DescSortExpr(querybuilder.StringExpr("uuid"))).
		Limit(limit).
		Lock(querybuilder.RowLockForNoKeyUpdateExpr()).
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
		return nil, fmt.Errorf("failed to query dispatches: %w", err)
	}

	// don't care about capacity, but a good heuristic is about 8 coz I like it
	records := make([]models.Dispatch, 0, 8)

	fn := func(tx pgx.Tx, r pgx.Rows) error {
		for r.Next() {
			var record models.Dispatch
			subErr := r.Scan(
				&record.UUID,
				&record.UOWUUID,
				&record.Target,
				&record.Metadata,
				&record.Payload,
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
