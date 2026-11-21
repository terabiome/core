package querybuilder

import "fmt"

type selectStatement struct {
	Cols   []Node
	From   Node
	Where  Node
	Order  []Node
	Limit  int
	Offset int
	Lock   Node
	When   Node
}

func (s *selectStatement) WriteTo(b *Builder) {
	b.sql.WriteString("SELECT ")
	if len(s.Cols) == 0 {
		b.sql.WriteString("*")
	} else {
		for i, c := range s.Cols {
			if i > 0 {
				b.sql.WriteString(", ")
			}
			c.WriteTo(b)
		}
	}

	b.sql.WriteString(" FROM ")
	s.From.WriteTo(b)

	if s.Where != nil {
		b.sql.WriteString(" WHERE ")
		s.Where.WriteTo(b)
	}

	if len(s.Order) > 0 {
		b.sql.WriteString(" ORDER BY ")
		for i, o := range s.Order {
			if i > 0 {
				b.sql.WriteString(", ")
			}
			o.WriteTo(b)
		}
	}

	if s.Lock != nil {
		b.sql.WriteString(" ")
		s.Lock.WriteTo(b)
	}
	if s.When != nil {
		b.sql.WriteString(" ")
		s.When.WriteTo(b)
	}
	if s.Limit > 0 {
		fmt.Fprintf(&b.sql, " LIMIT %d", s.Limit)
	}
	if s.Offset > 0 {
		fmt.Fprintf(&b.sql, " OFFSET %d", s.Offset)
	}
}
