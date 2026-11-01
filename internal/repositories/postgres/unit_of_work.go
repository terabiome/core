package repositories

import (
	"context"
	"fmt"

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
func (r *UnitOfWorkPostgresRepository) Create(ctx context.Context, uow models.UnitOfWork) error {
	const q = `
		INSERT INTO units_of_work (name, status, config_json)
		VALUES ($1, $2, $3, $4)
	`
	if _, err := r.wrappedExec(ctx, nil, q, uow.Name, uow.Status, uow.ConfigJSON); err != nil {
		return r.Errorf("failed to create unit-of-work: %w", err)
	}

	return nil
}

// GetByStatus fetches by status.
func (r *UnitOfWorkPostgresRepository) GetByStatus(ctx context.Context, status models.UnitOfWorkStatus) ([]models.UnitOfWork, error) {
	const q = `
		SELECT uuid, name, status, config_json, created_at
		FROM units_of_work
		WHERE status = $1
	`

	rows, err := r.wrappedQuery(ctx, nil, q, status)
	// smh this err is usually ignore, but good to check
	if err != nil {
		return nil, fmt.Errorf("failed to query units_of_work: %w", err)
	}
	defer rows.Close()

	// don't care about capacity, but a good heuristic is about 8 coz I like it
	records := make([]models.UnitOfWork, 0, 8)

	// can use pgx.* but i don't like it, rawdogging [pgx.Rows] instead
	for rows.Next() {
		var record models.UnitOfWork
		if err = rows.Scan(record.UUID, record.Name, record.Status, record.ConfigJSON, record.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan records of units_of_work, aborting the rest of the set: %w", err)
		}

		records = append(records, record)
	}

	// something is wrong and i don't like it
	if rows.Err() != nil {
		return nil, fmt.Errorf("something went wrong when querying units_of_work (before scanning): %w", rows.Err())
	}

	return records, nil
}

// --- Private SQL Helpers ---

// Ensure the repository satisfies the interface at compile time.
// var _ [interface] = (*struct)(nil)
