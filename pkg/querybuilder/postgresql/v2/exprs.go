package querybuilder

import (
	"fmt"
	"strings"
)

type SortOrder string

const (
	SortOrderAscending  SortOrder = "ASC"
	SortOrderDescending SortOrder = "DESC"
)

// Raw: trusted SQL syntax. Never user input.
type Raw string

func (r Raw) WriteTo(b *Builder) {
	b.sql.WriteString(string(r))
}

// Column: column name, optional alias.
type Column struct {
	Base  string
	Alias string
}

func (c Column) WriteTo(b *Builder) {
	b.sql.WriteString(c.Base)
	if c.Alias != "" {
		b.sql.WriteString(" AS ")
		b.sql.WriteString(c.Alias)
	}
}

// Table: table name, optional alias.
type Table struct {
	Base  string
	Alias string
}

func (t Table) WriteTo(b *Builder) {
	b.sql.WriteString(t.Base)
	if t.Alias != "" {
		b.sql.WriteString(" AS ")
		b.sql.WriteString(t.Alias)
	}
}

// Lit: inlined, escaped value.
type Lit struct {
	V any
}

func (l Lit) WriteTo(b *Builder) {
	switch v := l.V.(type) {
	case string:
		b.sql.WriteString("'")
		b.sql.WriteString(strings.ReplaceAll(v, "'", "''"))
		b.sql.WriteString("'")
	case nil:
		b.sql.WriteString("NULL")
	default:
		fmt.Fprintf(&b.sql, "'%v'", v) // covers your enum types
	}
}

// Val: bind parameter.
type Val struct {
	V any
}

func (v Val) WriteTo(b *Builder) {
	b.sql.WriteString(b.Arg(v.V))
}

// Cmp: comparison. R may be nil for IS NULL.
type Cmp struct {
	L  Node
	Op string
	R  Node
}

func (c Cmp) WriteTo(b *Builder) {
	c.L.WriteTo(b)
	b.sql.WriteString(" ")
	b.sql.WriteString(c.Op)
	if c.R != nil {
		b.sql.WriteString(" ")
		c.R.WriteTo(b)
	}
}

// BoolOp: AND/OR group.
type BoolOp struct {
	Op   string // "AND" or "OR"
	Kids []Node
}

func (g BoolOp) WriteTo(b *Builder) {
	for i, k := range g.Kids {
		if i > 0 {
			b.sql.WriteString(" ")
			b.sql.WriteString(g.Op)
			b.sql.WriteString(" ")
		}
		if needsParens(k, g.Op) {
			b.sql.WriteString("(")
		}
		k.WriteTo(b)
		if needsParens(k, g.Op) {
			b.sql.WriteString(")")
		}
	}
}

// Different operator → always parens. Same operator → associative, no parens.
func needsParens(child Node, parentOp string) bool {
	g, ok := child.(BoolOp)
	return ok && g.Op != parentOp
}

// Sort
type Sort struct {
	Expr Node
	Dir  SortOrder
}

func (s Sort) WriteTo(b *Builder) {
	s.Expr.WriteTo(b)
	b.sql.WriteString(" ")
	b.sql.WriteString(string(s.Dir))
}

// QualifiedTable: schema.table, optionally with an alias.
type QualifiedTable struct {
	Schema string
	Table  Table
}

func (q QualifiedTable) WriteTo(b *Builder) {
	b.sql.WriteString(q.Schema)
	b.sql.WriteString(".")
	q.Table.WriteTo(b) // Table.WriteTo handles the base name + optional alias
}
