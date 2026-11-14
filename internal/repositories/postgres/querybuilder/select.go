package querybuilder

type selectStatement struct {
	table Expr
	exprs []Expr
	joins []joinClause
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

func (stmt *selectStatement) OrderBy(exprs ...orderByExpr) *selectStatement {

	return stmt
}

func (stmt *selectStatement) Limit() *selectStatement {
	return stmt
}

func (stmt *selectStatement) Offset() *selectStatement {
	return stmt
}

type joinClause struct {
	table string
}

type orderByClause struct {
}
