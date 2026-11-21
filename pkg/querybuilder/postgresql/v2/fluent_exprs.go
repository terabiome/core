package querybuilder

import "fmt"

func ColumnExpr(s string) Column {
	return Column{Base: s}
}
func ColumnExprWithAliasExpr(s, a string) Column {
	return Column{Base: s, Alias: a}
}
func TableExpr(s string) Table {
	return Table{Base: s}
}
func TableWithAliasExpr(s, a string) Table {
	return Table{Base: s, Alias: a}
}
func SingleQuotedStringExpr(s string) Lit {
	return Lit{V: s}
}
func StringExpr(s string) Raw {
	return Raw(s)
}

func EqualCheckExpr(l, r Node) Cmp {
	return Cmp{L: l, Op: "=", R: r}
}
func NotEqualCheckExpr(l, r Node) Cmp {
	return Cmp{L: l, Op: "!=", R: r}
}
func NullCheckExpr(l Node) Cmp {
	return Cmp{L: l, Op: "IS NULL"}
}
func NotNullCheckExpr(l Node) Cmp {
	return Cmp{L: l, Op: "IS NOT NULL"}
}

func AscSortExpr(e Node) Sort {
	return Sort{Expr: e, Dir: SortOrderAscending}
}
func DescSortExpr(e Node) Sort {
	return Sort{Expr: e, Dir: SortOrderDescending}
}

func RowLockForUpdateExpr() Raw {
	return Raw("FOR UPDATE")
}
func RowLockForNoKeyUpdateExpr() Raw {
	return Raw("FOR NO KEY UPDATE")
}
func WhenLockedNoWaitExpr() Raw {
	return Raw("NOWAIT")
}
func WhenLockedSkipLockedExpr() Raw {
	return Raw("SKIP LOCKED")
}

func QualifiedTableExpr(schema, table string) QualifiedTable {
	if !tableNameRegex.MatchString(schema) {
		panic(fmt.Sprintf("invalid schema name: %q", schema))
	}
	return QualifiedTable{
		Schema: schema,
		Table:  Table{Base: table},
	}
}

func QualifiedTableWithAliasExpr(schema, table, alias string) QualifiedTable {
	if !tableNameRegex.MatchString(schema) {
		panic(fmt.Sprintf("invalid schema name: %q", schema))
	}
	return QualifiedTable{
		Schema: schema,
		Table:  Table{Base: table, Alias: alias},
	}
}

var Wildcard = Raw("*")
