package querybuilder

type insertStatement struct {
	Table      Node
	Cols       []string
	OnConflict Node // nil for now
	Returning  Node // nil for now
}

func (s *insertStatement) WriteTo(b *Builder) {
	b.sql.WriteString("INSERT INTO ")
	s.Table.WriteTo(b)

	if len(s.Cols) == 0 {
		panic("Insert: no columns")
	}

	b.sql.WriteString(" (")
	for i, c := range s.Cols {
		if i > 0 {
			b.sql.WriteString(", ")
		}
		b.sql.WriteString(c)
	}
	b.sql.WriteString(") VALUES (")

	for i := range s.Cols {
		if i > 0 {
			b.sql.WriteString(", ")
		}
		b.sql.WriteString(b.Arg(nil)) // $1, $2, ...
	}
	b.sql.WriteString(")")

	if s.OnConflict != nil {
		b.sql.WriteString(" ")
		s.OnConflict.WriteTo(b)
	}
	if s.Returning != nil {
		b.sql.WriteString(" ")
		s.Returning.WriteTo(b)
	}
}
