package postgresql

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/terabiome/core/internal/models"
	"github.com/terabiome/core/internal/repositories"
	qb "github.com/terabiome/core/pkg/querybuilder/postgresql"
)

type DispatchPostgresRepository struct {
	basePostgresRepository
}

func NewDispatchPostgresRepository(pool *pgxpool.Pool) *DispatchPostgresRepository {
	base := basePostgresRepository{pool}
	return &DispatchPostgresRepository{base}
}

func (r *DispatchPostgresRepository) BulkInsert(ctx context.Context, dispatches []models.Dispatch) error {
	q, err := qb.
		Insert(qb.StringExpr("dispatches")).
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

func (r *DispatchPostgresRepository) FilterByStatus(ctx context.Context, status models.DispatchStatus, limit int) (*repositories.DispatchFilterOutput, error) {
	q, err := qb.
		Select(
			qb.ColumnExpr("uuid"),
			qb.ColumnExpr("uow_uuid"),
			qb.ColumnExpr("target"),
			qb.ColumnExpr("metadata"),
			qb.ColumnExpr("payload"),
		).
		FromExpr(qb.TableExpr("dispatches")).
		Where(qb.EqualCheckExpr(
			qb.StringExpr("status"),
			qb.SingleQuotedStringExpr(string(status))),
		).
		OrderBy(qb.DescSortExpr(qb.StringExpr("uuid"))).
		Limit(limit).
		Lock(qb.RowLockForNoKeyUpdateExpr()).
		WhenLocked(qb.WhenLockedSkipLockedExpr()).
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

	projection := repositories.DispatchFilterOutput{
		// don't care about capacity, but a good heuristic is about 8 coz I like it
		Records: make([]repositories.DispatchFilterRecord, 0, 8),
	}

	fn := func(tx pgx.Tx, r pgx.Rows) error {
		for r.Next() {
			var (
				dbRecord models.Dispatch
			)
			subErr := r.Scan(
				&dbRecord.UUID,
				&dbRecord.UOWUUID,
				&dbRecord.Target,
				&dbRecord.Metadata,
				&dbRecord.Payload,
			)
			if subErr != nil {
				return subErr
			}

			projection.Records = append(projection.Records, repositories.DispatchFilterRecord{
				UUID:     dbRecord.UUID,
				UOWUUID:  dbRecord.UOWUUID,
				Target:   dbRecord.Target,
				Metadata: dbRecord.Metadata,
				Payload:  dbRecord.Payload,
			})
		}
		return nil
	}
	if err := r.execScanCallback(tx, rows, fn); err != nil {
		return nil, fmt.Errorf("failed to scan units_of_work: %w", err)
	}

	return &projection, nil
}
