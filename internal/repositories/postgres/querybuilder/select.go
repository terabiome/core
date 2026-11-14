package querybuilder

import (
	"fmt"
	"strings"

	"github.com/terabiome/core/internal/repositories"
)

// no JOIN yet
type selectStatement struct {
	table    Expr
	exprs    []Expr
	where    []Expr
	orderBy  []sortFieldExpr
	limit    int
	offset   int
	lockType repositories.LockType
}

func Select(exprs ...Expr) *selectStatement {
	return &selectStatement{exprs: exprs}
}

func (stmt *selectStatement) FromExpr(tableExpr Expr) *selectStatement {
	stmt.table = tableExpr
	return stmt
}

func (stmt *selectStatement) From(table string) *selectStatement {
	return stmt.FromExpr(StringExpr(table))
}

func (stmt *selectStatement) Where(expr Expr) *selectStatement {
	stmt.where = []Expr{expr}
	return stmt
}

func (stmt *selectStatement) OrWhere(expr Expr) *selectStatement {
	if len(stmt.exprs) < 1 {
		return stmt.Where(expr)
	}

	stmt.exprs = append(stmt.exprs, Or, expr)
	return stmt
}

func (stmt *selectStatement) AndWhere(expr Expr) *selectStatement {
	if len(stmt.exprs) < 1 {
		return stmt.Where(expr)
	}

	stmt.exprs = append(stmt.exprs, And, expr)
	return stmt
}

func (stmt *selectStatement) OrderBy(exprs ...sortFieldExpr) *selectStatement {
	stmt.orderBy = exprs
	return stmt
}

func (stmt *selectStatement) Limit(val int) *selectStatement {
	stmt.limit = val
	return stmt
}

func (stmt *selectStatement) Offset(val int) *selectStatement {
	stmt.offset = val
	return stmt
}

func (stmt *selectStatement) Lock(lockType repositories.LockType) *selectStatement {
	stmt.lockType = lockType
	return stmt
}

func (stmt *selectStatement) Parameterize() (string, error) {
	// 1. Validate and evaluate the table
	if stmt.table == nil {
		return "", fmt.Errorf("table expression is required")
	}
	tableStr, err := stmt.table.Eval()
	if err != nil {
		return "", fmt.Errorf("failed to evaluate table: %w", err)
	}

	// 2. Evaluate SELECT columns
	var selectCols []string
	if len(stmt.exprs) == 0 {
		selectCols = append(selectCols, "*")
	} else {
		for _, expr := range stmt.exprs {
			colStr, err := expr.Eval()
			if err != nil {
				return "", fmt.Errorf("failed to evaluate select expression: %w", err)
			}
			selectCols = append(selectCols, colStr)
		}
	}

	q := fmt.Sprintf("SELECT %s FROM %s", strings.Join(selectCols, ", "), tableStr)

	// 3. Evaluate WHERE clause
	if stmt.where != nil {
		segments, err := EvalFold(stmt.where)
		if err != nil {
			return "", fmt.Errorf("failed to eval expression group in WHERE: %w", err)
		}

		if len(segments) > 1 {
			q = fmt.Sprintf("%s WHERE %s", q, strings.Join(segments, " "))
		}
	}

	// 4. Evaluate ORDER BY clause
	if len(stmt.orderBy) > 0 {
		var orderByStrs []string
		for _, ob := range stmt.orderBy {
			obStr, err := ob.Eval()
			if err != nil {
				return "", fmt.Errorf("failed to evaluate order by expression: %w", err)
			}
			orderByStrs = append(orderByStrs, obStr)
		}
		q += fmt.Sprintf(" ORDER BY %s", strings.Join(orderByStrs, ", "))
	}

	// 5. Append Lock clause (e.g., FOR NO KEY UPDATE)
	// Note: Adjust the condition below if repositories.LockType is an int enum (use != 0)
	// instead of a string alias (use != ""). fmt.Sprint is used as a safe fallback.
	if lockStr := fmt.Sprint(stmt.lockType); lockStr != "" && lockStr != "0" {
		q += fmt.Sprintf(" %s", stmt.lockType)
	}

	// 6. Append LIMIT and OFFSET
	if stmt.limit > 0 {
		q += fmt.Sprintf(" LIMIT %d", stmt.limit)
	}

	if stmt.offset > 0 {
		q += fmt.Sprintf(" OFFSET %d", stmt.offset)
	}

	return q, nil
}
