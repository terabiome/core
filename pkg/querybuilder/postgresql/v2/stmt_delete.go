package querybuilder

type deleteStatement struct {
	Table     Node // *Table
	Where     Node // nil = delete all rows
	Returning Node // stub for later
}

func (s *deleteStatement) WriteTo(b *Builder) {
	b.sql.WriteString("DELETE FROM ")
	s.Table.WriteTo(b)

	if s.Where != nil {
		b.sql.WriteString(" WHERE ")
		s.Where.WriteTo(b)
	}

	if s.Returning != nil {
		b.sql.WriteString(" ")
		s.Returning.WriteTo(b)
	}
}
