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

type stringExpr struct {
	val string
}

func (e stringExpr) Eval() (string, error) {
	return e.val, nil
}

type tableExpr struct {
	base  string
	alias string
}

func (e tableExpr) Eval() (string, error) {
	if !tableNameRegex.Match([]byte(e.base)) {
		return "", fmt.Errorf("invalid table name")
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

	rhs, err := e.rhs.Eval()
	if err != nil {
		return "", fmt.Errorf("failed to evaluate RHS: %w", err)
	}

	if lhs == "" || e.comparator == "" {
		return "", fmt.Errorf("missing LHS or comparator")
	}

	if rhs != "" {
		return fmt.Sprintf("%s %s %s", lhs, e.comparator, rhs), nil
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

func StringExpr(val string) stringExpr {
	return stringExpr{
		val: val,
	}
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

func EvalFold(exprs []Expr) ([]string, error) {
	var err error
	segments := make([]string, 0, len(exprs))
	for idx, expr := range exprs {
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

var Wildcard = stringExpr{"*"}
var Or = stringExpr{"OR"}
var And = stringExpr{"AND"}
