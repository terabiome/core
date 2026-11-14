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

type orderByExpr struct {
	sortExpr  Expr
	direction SortOrder
}

func (e orderByExpr) Eval() (string, error) {
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

func OrderByExpr(sortExpr Expr, direction SortOrder) orderByExpr {
	return orderByExpr{
		sortExpr:  sortExpr,
		direction: direction,
	}
}

var Wildcard = stringExpr{"*"}
