package querybuilder

import "github.com/terabiome/core/internal/repositories"

type selectStatement struct {
	table  Expr
	exprs  []Expr
	joins  []joinClause
	limit  int
	offset int

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
	return stmt
}

func (stmt *selectStatement) OrderBy(exprs ...sortFieldExpr) *selectStatement {
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
	return "", nil
}

type joinClause struct {
	table string
}

type orderByClause struct {
}
