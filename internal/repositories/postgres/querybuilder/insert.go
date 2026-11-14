package querybuilder

import (
	"fmt"
	"strings"
)

type insert struct {
	table          string
	columns        []string
	returningExprs []Expr
}

func (b *insert) Columns(columns ...string) *insert {
	b.columns = columns
	return b
}

func (b *insert) Parameterize() (string, error) {
	if !tableNameRegex.Match([]byte(b.table)) {
		return "", fmt.Errorf("invalid table name: %s", b.table)
	}

	var q = fmt.Sprintf("INSERT INTO %s", b.table)

	if len(b.columns) < 1 {
		return "", fmt.Errorf("no column")
	}

	var columnMap = make(map[string]bool, len(b.columns))

	for _, col := range b.columns {
		if !columnNameRegex.Match([]byte(col)) {
			return "", fmt.Errorf("invalid column name: %s", col)
		}

		if _, ok := columnMap[col]; ok {
			return "", fmt.Errorf("repeated column name/alias: %s", col)
		}

		columnMap[col] = true
	}

	colPlaceholders := make([]string, 0, len(b.columns))
	for i := range len(b.columns) {
		colPlaceholders = append(colPlaceholders, fmt.Sprintf("$%d", i+1))
	}

	q = fmt.Sprintf("%s (%s) VALUES (%s)",
		q,
		strings.Join(b.columns, ", "),
		strings.Join(colPlaceholders, ", "),
	)

	// TODO: implement ON CONFLICT
	// TODO: implement RETURNING

	return q, nil
}
