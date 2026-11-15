package querybuilder

import (
	"errors"
	"fmt"
	"strings"
)

type SortOrder string

const (
	SortOrderAscending  SortOrder = "ASC"
	SortOrderDescending SortOrder = "DESC"
)

type stringExpr string

func (e stringExpr) Eval() (string, error) {
	return string(e), nil
}

type tableExpr struct {
	base  string
	alias string
}

func (e tableExpr) Eval() (string, error) {
	if !tableNameRegex.Match([]byte(e.base)) {
		return "", fmt.Errorf("invalid table name: value=%s pattern=%s", e.base, tableNameRegex.String())
	}

	if e.alias != "" && !aliasNameRegex.Match([]byte(e.alias)) {
		return "", fmt.Errorf("invalid alias name")
	}

	if e.alias == "" {
		return e.base, nil
	}

	return fmt.Sprintf("%s AS %s", e.base, e.alias), nil
}

type columnExpr struct {
	base  string
	alias string
}

func (e columnExpr) Eval() (string, error) {
	if !columnNameRegex.Match([]byte(e.base)) {
		return "", fmt.Errorf("invalid column name: value=%s pattern=%s", e.base, columnNameRegex.String())
	}

	if e.alias != "" && !aliasNameRegex.Match([]byte(e.alias)) {
		return "", fmt.Errorf("invalid alias name")
	}

	if e.alias == "" {
		return e.base, nil
	}

	return fmt.Sprintf("%s AS %s", e.base, e.alias), nil
}

type sortFieldExpr struct {
	sortExpr  Expr
	direction SortOrder
}

func (e sortFieldExpr) Eval() (string, error) {
	switch e.direction {
	case SortOrderAscending:
	case SortOrderDescending:
	default:
		return "", fmt.Errorf("invalid sort direction: %s", e.direction)
	}

	sortExpr, err := e.sortExpr.Eval()
	if err != nil {
		return "", fmt.Errorf("invalid sort expression: %w", err)
	}

	return fmt.Sprintf("%s %s", sortExpr, e.direction), nil
}

type comparisonExpr struct {
	lhs        Expr
	rhs        Expr
	comparator string
}

func (e comparisonExpr) Eval() (string, error) {
	lhs, err := e.lhs.Eval()
	if err != nil {
		return "", fmt.Errorf("failed to evaluate LHS: %w", err)
	}

	if lhs == "" || e.comparator == "" {
		return "", fmt.Errorf("missing LHS or comparator")
	}

	if e.rhs != nil {
		rhs, err := e.rhs.Eval()
		if err != nil {
			return "", fmt.Errorf("failed to evaluate RHS: %w", err)
		}

		if rhs != "" {
			return fmt.Sprintf("%s %s %s", lhs, e.comparator, rhs), nil
		}
	}

	return fmt.Sprintf("%s %s", lhs, e.comparator), nil
}

type comparisonGroupExpr struct {
	exprs []Expr
}

func (e *comparisonGroupExpr) And(expr Expr) *comparisonGroupExpr {
	e.exprs = append(e.exprs, And, expr)
	return e
}

func (e *comparisonGroupExpr) Or(expr Expr) *comparisonGroupExpr {
	e.exprs = append(e.exprs, Or, expr)
	return e
}

func (e comparisonGroupExpr) Eval() (string, error) {
	if len(e.exprs) == 0 {
		return "", fmt.Errorf("no expression")
	}

	if len(e.exprs) == 1 {
		return e.exprs[0].Eval()
	}

	segments, err := EvalFold(e.exprs)
	if err != nil {
		return "", fmt.Errorf("failed to eval expr group: %w", err)
	}

	return fmt.Sprintf("(%s)", strings.Join(segments, " ")), nil
}

type rowLockExpr struct {
	stringExpr
}

type whenLockedExpr struct {
	stringExpr
}

func StringExpr(val string) stringExpr {
	return stringExpr(val)
}

func DoubleQuotedStringExpr(val string) stringExpr {
	return stringExpr(fmt.Sprintf("\"%s\"", val))
}

func SingleQuotedStringExpr(val string) stringExpr {
	return stringExpr(fmt.Sprintf("'%s'", val))
}

func TableExpr(base string) tableExpr {
	return tableExpr{
		base: base,
	}
}

func TableWithAliasExpr(base string, alias string) tableExpr {
	return tableExpr{
		base:  base,
		alias: alias,
	}
}

func ColumnExpr(column string) columnExpr {
	return columnExpr{
		base: column,
	}
}

func ColumnExprWithAliasExpr(column string, alias string) columnExpr {
	return columnExpr{
		base:  column,
		alias: alias,
	}
}

func AscSortExpr(sortExpr Expr) sortFieldExpr {
	return sortFieldExpr{
		sortExpr:  sortExpr,
		direction: SortOrderAscending,
	}
}

func DescSortExpr(sortExpr Expr) sortFieldExpr {
	return sortFieldExpr{
		sortExpr:  sortExpr,
		direction: SortOrderDescending,
	}
}

func NullCheckExpr(lhs Expr) comparisonExpr {
	return comparisonExpr{
		lhs:        lhs,
		comparator: "IS NULL",
	}
}

func NotNullCheckExpr(lhs Expr) comparisonExpr {
	return comparisonExpr{
		lhs:        lhs,
		comparator: "IS NOT NULL",
	}
}

func EqualCheckExpr(lhs, rhs Expr) comparisonExpr {
	return comparisonExpr{
		lhs:        lhs,
		comparator: "=",
		rhs:        rhs,
	}
}

func NotEqualCheckExpr(lhs, rhs Expr) comparisonExpr {
	return comparisonExpr{
		lhs:        lhs,
		comparator: "!=",
		rhs:        rhs,
	}
}

func ComparisonGroupExpr(expr Expr) *comparisonGroupExpr {
	return &comparisonGroupExpr{
		exprs: []Expr{expr},
	}
}

func RowLockForUpdateExpr() rowLockExpr {
	return rowLockExpr{stringExpr("FOR UPDATE")}
}

func RowLockForNoKeyUpdateExpr() rowLockExpr {
	return rowLockExpr{stringExpr("FOR NO KEY UPDATE")}
}

func WhenLockedNoWaitExpr() whenLockedExpr {
	return whenLockedExpr{stringExpr("NOWAIT")}
}

func WhenLockedSkipLockedExpr() whenLockedExpr {
	return whenLockedExpr{stringExpr("SKIP LOCKED")}
}

func EvalFold(exprs []Expr) ([]string, error) {
	var err error
	segments := make([]string, 0, len(exprs))

	for idx, expr := range exprs {
		if expr == nil {
			continue
		}

		exprStr, subErr := expr.Eval()
		if subErr != nil {
			err = errors.Join(err, fmt.Errorf("failed to eval expr at idx %d: %w", idx, err))
			continue
		}
		segments = append(segments, exprStr)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to eval expr group: %w", err)
	}

	return segments, nil
}

var Wildcard = stringExpr("*")
var Or = stringExpr("OR")
var And = stringExpr("AND")
