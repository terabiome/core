package querybuilder

import (
	"fmt"
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

type aliasTableExpr struct {
	base  string
	alias string
}

func (e aliasTableExpr) Eval() (string, error) {
	if !tableNameRegex.Match([]byte(e.base)) {
		return "", fmt.Errorf("invalid table name")
	}

	if !aliasNameRegex.Match([]byte(e.alias)) {
		return "", fmt.Errorf("invalid alias name")
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

func StringExpr(val string) stringExpr {
	return stringExpr{
		val: val,
	}
}

func AliasTableExpr(base string, alias string) aliasTableExpr {
	return aliasTableExpr{
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

var Wildcard = stringExpr{"*"}
