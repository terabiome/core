package querybuilder

import "fmt"

type deleteStatementWrapper struct {
	stmt *deleteStatement
}

func Delete(t Node) *deleteStatementWrapper {
	return &deleteStatementWrapper{stmt: &deleteStatement{Table: t}}
}

func (s *deleteStatementWrapper) Where(n Node) *deleteStatementWrapper {
	s.stmt.Where = n
	return s
}

func (s *deleteStatementWrapper) AndWhere(n Node) *deleteStatementWrapper {
	if s.stmt.Where == nil {
		return s.Where(n)
	}
	s.stmt.Where = BoolOp{Op: "AND", Kids: []Node{s.stmt.Where, n}}
	return s
}

func (s *deleteStatementWrapper) OrWhere(n Node) *deleteStatementWrapper {
	if s.stmt.Where == nil {
		return s.Where(n)
	}
	s.stmt.Where = BoolOp{Op: "OR", Kids: []Node{s.stmt.Where, n}}
	return s
}

func (s *deleteStatementWrapper) Build() (string, []any, error) {
	if s.stmt.Table == nil {
		return "", nil, fmt.Errorf("Delete: table required")
	}
	b := &Builder{}
	s.stmt.WriteTo(b)
	return b.sql.String(), b.args, nil
}

// Shim matching the rest of your API
func (s *deleteStatementWrapper) Parameterize() (string, error) {
	q, _, err := s.Build()
	return q, err
}
