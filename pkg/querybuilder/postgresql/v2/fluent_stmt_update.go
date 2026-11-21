package querybuilder

import "fmt"

type updateStatementWrapper struct {
	stmt *updateStatement
}

func Update(t Node) *updateStatementWrapper {
	return &updateStatementWrapper{stmt: &updateStatement{Table: t}}
}

func (s *updateStatementWrapper) Set(col string, val Node) *updateStatementWrapper {
	s.stmt.Sets = append(s.stmt.Sets, Set{Col: col, Value: val})
	return s
}

func (s *updateStatementWrapper) Where(n Node) *updateStatementWrapper {
	s.stmt.Where = n
	return s
}

func (s *updateStatementWrapper) AndWhere(n Node) *updateStatementWrapper {
	if s.stmt.Where == nil {
		return s.Where(n)
	}
	s.stmt.Where = BoolOp{Op: "AND", Kids: []Node{s.stmt.Where, n}}
	return s
}

func (s *updateStatementWrapper) OrWhere(n Node) *updateStatementWrapper {
	if s.stmt.Where == nil {
		return s.Where(n)
	}
	s.stmt.Where = BoolOp{Op: "OR", Kids: []Node{s.stmt.Where, n}}
	return s
}

func (s *updateStatementWrapper) Build() (string, []any, error) {
	if s.stmt.Table == nil {
		return "", nil, fmt.Errorf("Update: table required")
	}
	if len(s.stmt.Sets) == 0 {
		return "", nil, fmt.Errorf("Update: at least one Set required")
	}
	for _, set := range s.stmt.Sets {
		if !columnNameRegex.MatchString(set.Col) {
			return "", nil, fmt.Errorf("invalid column: %q", set.Col)
		}
	}
	b := &Builder{}
	s.stmt.WriteTo(b)
	return b.sql.String(), b.args, nil
}

func (s *updateStatementWrapper) Parameterize() (string, error) {
	q, _, err := s.Build()
	return q, err
}
