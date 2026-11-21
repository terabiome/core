package querybuilder

import "fmt"

type insertStatementWrapper struct {
	stmt *insertStatement
}

func Insert(t Node) *insertStatementWrapper {
	return &insertStatementWrapper{stmt: &insertStatement{Table: t}}
}

func (s *insertStatementWrapper) Columns(cols ...string) *insertStatementWrapper {
	s.stmt.Cols = cols
	return s
}

func (s *insertStatementWrapper) Build() (string, []any, error) {
	if err := s.validate(); err != nil {
		return "", nil, err
	}
	b := &Builder{}
	s.stmt.WriteTo(b)
	return b.sql.String(), b.args, nil
}

func (s *insertStatementWrapper) validate() error {
	if s.stmt.Table == nil {
		return fmt.Errorf("Insert: table required")
	}
	if len(s.stmt.Cols) == 0 {
		return fmt.Errorf("Insert: no columns")
	}
	seen := make(map[string]bool, len(s.stmt.Cols))
	for _, c := range s.stmt.Cols {
		if !columnNameRegex.MatchString(c) {
			return fmt.Errorf("invalid column name: %q", c)
		}
		if seen[c] {
			return fmt.Errorf("repeated column name: %q", c)
		}
		seen[c] = true
	}
	return nil
}

// Backwards-compatible shim.
func (s *insertStatementWrapper) Parameterize() (string, error) {
	q, _, err := s.Build()
	return q, err
}
