package postgres

import (
	"fmt"
	"regexp"
	"strings"
)

// intrinsic patterns
const (
	numericPattern            = "[0-9]"
	lowercaseCharacterPattern = "[a-z]"
	uppercaseCharacterPattern = "[A-Z]"
	underscorePattern         = "[_]"
)

const (
	characterPattern                  = (lowercaseCharacterPattern + uppercaseCharacterPattern)
	characterWithUnderscorePattern    = (underscorePattern + characterPattern)
	alphanumericPattern               = (characterPattern + numericPattern)
	alphanumericWithUnderscorePattern = (underscorePattern + alphanumericPattern)
	numericWithUnderscorePattern      = (underscorePattern + numericPattern)
)

var (
	tableNameRegex = regexp.MustCompile(fmt.Sprintf("%s+%s*", characterPattern, alphanumericWithUnderscorePattern))
)

// did not factor in <x>.<column-name>.
// make it work first.
var (
	columnNameRegex          = regexp.MustCompile(fmt.Sprintf("%s+%s*", characterPattern, alphanumericWithUnderscorePattern))
	columnNameWithAliasRegex = regexp.MustCompile(fmt.Sprintf("%s AS %s", columnNameRegex, columnNameRegex))
)

type InsertQueryBuilder struct {
	table   string
	columns []string
}

func (b *InsertQueryBuilder) Build() (string, error) {
	if !tableNameRegex.Match([]byte(b.table)) {
		return "", fmt.Errorf("invalid table name: %s", b.table)
	}

	if len(b.columns) < 1 {
		return "", fmt.Errorf("no column")
	}

	var columnMap = make(map[string]bool, len(b.columns))

	for _, col := range b.columns {
		if !columnNameWithAliasRegex.Match([]byte(col)) {
			return "", fmt.Errorf("invalid column name: %s", col)
		}

		if _, ok := columnMap[col]; ok {
			return "", fmt.Errorf("repeated column name: %s", col)
		}

		columnMap[col] = true
	}

	colPlaceholders := make([]string, len(b.columns))
	for i := range len(b.columns) {
		colPlaceholders = append(colPlaceholders, fmt.Sprintf("$%d", i+1))
	}

	var q = fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		b.table,
		strings.Join(b.columns, ", "),
		strings.Join(colPlaceholders, ", "),
	)

	return q, nil
}
