package querybuilder

type Expr interface {
	Eval() (string, error)
}
