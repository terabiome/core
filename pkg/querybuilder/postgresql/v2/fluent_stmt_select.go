package querybuilder

type selectStatementWrapper struct{ stmt *selectStatement }

func Select(cols ...Node) *selectStatementWrapper {
	return &selectStatementWrapper{stmt: &selectStatement{Cols: cols}}
}

func (s *selectStatementWrapper) From(t string) *selectStatementWrapper {
	s.stmt.From = Table{Base: t}
	return s
}
func (s *selectStatementWrapper) FromExpr(t Node) *selectStatementWrapper {
	s.stmt.From = t
	return s
}
func (s *selectStatementWrapper) Where(n Node) *selectStatementWrapper {
	s.stmt.Where = n
	return s
}
func (s *selectStatementWrapper) AndWhere(n Node) *selectStatementWrapper {
	if s.stmt.Where == nil {
		return s.Where(n)
	}
	s.stmt.Where = BoolOp{Op: "AND", Kids: []Node{s.stmt.Where, n}}
	return s
}
func (s *selectStatementWrapper) OrWhere(n Node) *selectStatementWrapper {
	if s.stmt.Where == nil {
		return s.Where(n)
	}
	s.stmt.Where = BoolOp{Op: "OR", Kids: []Node{s.stmt.Where, n}}
	return s
}
func (s *selectStatementWrapper) OrderBy(o ...Node) *selectStatementWrapper {
	s.stmt.Order = o
	return s
}
func (s *selectStatementWrapper) Limit(n int) *selectStatementWrapper {
	s.stmt.Limit = n
	return s
}
func (s *selectStatementWrapper) Offset(n int) *selectStatementWrapper {
	s.stmt.Offset = n
	return s
}
func (s *selectStatementWrapper) Lock(n Node) *selectStatementWrapper {
	s.stmt.Lock = n
	return s
}
func (s *selectStatementWrapper) WhenLocked(n Node) *selectStatementWrapper {
	s.stmt.When = n
	return s
}

func (s *selectStatementWrapper) Build() (string, []any) {
	b := &Builder{}
	s.stmt.WriteTo(b)
	return b.sql.String(), b.args
}
