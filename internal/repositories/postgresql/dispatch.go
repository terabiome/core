package postgresql

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/terabiome/core/internal/models"
	"github.com/terabiome/core/internal/repositories"
	querybuilder "github.com/terabiome/core/pkg/querybuilder/postgresql/v2"
)

type DispatchPostgresRepository struct {
	basePostgresRepository
}

func NewDispatchPostgresRepository(pool *pgxpool.Pool) *DispatchPostgresRepository {
	base := basePostgresRepository{pool}
	return &DispatchPostgresRepository{base}
}

func (r *DispatchPostgresRepository) BulkInsert(ctx context.Context, dispatches []models.Dispatch) error {
	q, _, err := querybuilder.
		Insert(querybuilder.TableExpr("dispatches")).
		Columns("uow_uuid", "status", "target", "payload").
		Build()

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
	q, _ := querybuilder.
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
		WhenLocked(querybuilder.WhenLockedSkipLockedExpr()).
		Build()

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

func (r *DispatchPostgresRepository) Update(ctx context.Context, input repositories.DispatchUpdateInput) error {
	q, args, err := querybuilder.Update(querybuilder.TableExpr("dispatches")).
		SetWhen("status", querybuilder.Val{V: input.Fields.Status}, func() bool {
			return input.Fields.Status != ""
		}).
		SetWhen("metadata", querybuilder.Val{V: input.Fields.Metadata}, func() bool {
			return input.Fields.Metadata != nil
		}).
		SetWhen("target", querybuilder.Val{V: input.Fields.Target}, func() bool {
			return input.Fields.Target != nil
		}).
		SetWhen("payload", querybuilder.Val{V: input.Fields.Payload}, func() bool {
			return input.Fields.Payload != nil
		}).
		WhereWhen(querybuilder.EqualCheckExpr(querybuilder.ColumnExpr("uuid"), querybuilder.Val{V: input.Filters.UUID}),
			func() bool {
				return input.Filters.UUID != nil
			},
		).Build()

	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	tx, err := loadTxFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("failed to load transaction from ctx: %w", err)
	}

	if _, err := r.wrappedExec(ctx, tx, q, args...); err != nil {
		return fmt.Errorf("failed to update dispatch: %w", err)
	}

	return nil
}

func (r *DispatchPostgresRepository) BulkUpdate(ctx context.Context, inputs []repositories.DispatchUpdateInput) error {
	if len(inputs) == 0 {
		return nil
	}

	batch := pgx.Batch{}
	for _, input := range inputs {
		q, args, err := querybuilder.Update(querybuilder.TableExpr("dispatches")).
			Set("updated_at", querybuilder.Val{V: time.Now().UTC()}).
			SetWhen("status", querybuilder.Val{V: input.Fields.Status}, func() bool {
				return input.Fields.Status != ""
			}).
			SetWhen("metadata", querybuilder.Val{V: input.Fields.Metadata}, func() bool {
				return input.Fields.Metadata != nil
			}).
			SetWhen("target", querybuilder.Val{V: input.Fields.Target}, func() bool {
				return input.Fields.Target != nil
			}).
			SetWhen("payload", querybuilder.Val{V: input.Fields.Payload}, func() bool {
				return input.Fields.Payload != nil
			}).
			WhereWhen(querybuilder.EqualCheckExpr(querybuilder.ColumnExpr("uuid"), querybuilder.Val{V: input.Filters.UUID}),
				func() bool {
					return input.Filters.UUID != nil
				},
			).Build()

		if err != nil {
			return fmt.Errorf("failed to build update query: %w", err)
		}

		batch.Queue(q, args...)
	}

	tx, err := loadTxFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("failed to load transaction from ctx: %w", err)
	}

	if err := r.wrappedSendBatch(ctx, tx, &batch).Close(); err != nil {
		return fmt.Errorf("failed to bulk update dispatches: %w", err)
	}

	return nil
}
