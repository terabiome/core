package querybuilder

import (
	"fmt"
	"strings"
)

type Builder struct {
	sql  strings.Builder
	args []any
}

func (b *Builder) Arg(v any) string {
	b.args = append(b.args, v)
	return fmt.Sprintf("$%d", len(b.args))
}
