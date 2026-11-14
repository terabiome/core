package querybuilder

import (
	"fmt"
	"strings"
)

type insertStatement struct {
	table      Expr
	columns    []string
	onConflict *onConflictClause
	returning  *returningClause
}

func Insert(table Expr) *insertStatement {
	return &insertStatement{table: table}
}

func (stmt *insertStatement) Columns(columns ...string) *insertStatement {
	stmt.columns = columns
	return stmt
}

func (stmt *insertStatement) Parameterize() (string, error) {
	table, err := stmt.table.Eval()
	if err != nil {
		return "", err
	}

	if !tableNameRegex.Match([]byte(table)) {
		return "", fmt.Errorf("invalid table name: %s", table)
	}

	var q = fmt.Sprintf("INSERT INTO %s", table)

	if len(stmt.columns) < 1 {
		return "", fmt.Errorf("no column")
	}

	var columnMap = make(map[string]bool, len(stmt.columns))

	for _, col := range stmt.columns {
		if !columnNameRegex.Match([]byte(col)) {
			return "", fmt.Errorf("invalid column name: %s", col)
		}

		if _, ok := columnMap[col]; ok {
			return "", fmt.Errorf("repeated column name/alias: %s", col)
		}

		columnMap[col] = true
	}

	colPlaceholders := make([]string, 0, len(stmt.columns))
	for i := range len(stmt.columns) {
		colPlaceholders = append(colPlaceholders, fmt.Sprintf("$%d", i+1))
	}

	q = fmt.Sprintf("%s (%s) VALUES (%s)",
		q,
		strings.Join(stmt.columns, ", "),
		strings.Join(colPlaceholders, ", "),
	)

	// TODO: implement ON CONFLICT
	// TODO: implement RETURNING

	return q, nil
}

type returningClause struct {
}

type onConflictClause struct {
}
