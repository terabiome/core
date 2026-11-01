package repositories

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

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

// Ensure the repository satisfies the interface at compile time.
// var _ [interface] = (*struct)(nil)
