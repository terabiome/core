package querybuilder

type updateStatement struct {
	Table     Node  // *Table
	Sets      []Set // at least one required
	Where     Node  // nil = update all rows
	Returning Node  // stub
}

type Set struct {
	Col   string // validated column name
	Value Node   // Lit, Val, or an expression like Col+Val
}

func (s *updateStatement) WriteTo(b *Builder) {
	b.sql.WriteString("UPDATE ")
	s.Table.WriteTo(b)
	b.sql.WriteString(" SET ")

	for i, set := range s.Sets {
		if i > 0 {
			b.sql.WriteString(", ")
		}
		b.sql.WriteString(set.Col)
		b.sql.WriteString(" = ")
		set.Value.WriteTo(b)
	}

	if s.Where != nil {
		b.sql.WriteString(" WHERE ")
		s.Where.WriteTo(b)
	}

	if s.Returning != nil {
		b.sql.WriteString(" ")
		s.Returning.WriteTo(b)
	}
}
