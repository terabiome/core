package querybuilder

import (
	"strings"
	"testing"
)

func TestSelectQueryBuilder(t *testing.T) {
	q, err := Select(
		StringExpr("uuid"),
		StringExpr("name"),
		StringExpr("config"),
		StringExpr("created_at"),
		StringExpr("staged_at"),
	).
		FromExpr(TableWithAliasExpr("units_of_work", "uow")).
		Where(NullCheckExpr(StringExpr("staged_at"))).
		Lock(RowLockForNoKeyUpdateExpr()).
		OrderBy(DescSortExpr(StringExpr("uuid"))).
		Limit(3).
		Parameterize()

	if err != nil {
		t.Fatal(err)
	}

	refQ := `
		SELECT uuid, name, config, created_at, staged_at
		FROM units_of_work AS uow
		WHERE staged_at IS NULL
		ORDER BY uuid DESC
		FOR NO KEY UPDATE
		LIMIT 3
	`

	fn := func(r rune) bool {
		if r == rune('\n') {
			return true
		}

		if r == rune('\r') {
			return true
		}

		if r == rune(' ') {
			return true
		}

		return false
	}

	refQ = strings.ReplaceAll(refQ, "\n", " ")
	refQ = strings.ReplaceAll(refQ, "\r", " ")
	refQ = strings.ReplaceAll(refQ, "\t", "")
	refQ = strings.TrimFunc(refQ, fn)

	if q != refQ {
		t.Fatalf("output mismatch: '%s' != '%s'", q, refQ)
	}
}
