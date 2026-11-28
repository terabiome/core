package postgresql

import (
	"fmt"

	"github.com/jackc/pgx/v5"
)

func ScanRows(rows pgx.Rows, callback func(pgx.Rows) error) error {
	defer rows.Close()

	if err := callback(rows); err != nil {
		return fmt.Errorf("failed to scan, abort the set: %w", err)
	}

	if rows.Err() != nil {
		return fmt.Errorf("something went wrong when querying: %w", rows.Err())
	}

	return nil
}
