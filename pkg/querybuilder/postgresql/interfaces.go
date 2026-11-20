package postgresql

type Expr interface {
	Eval() (string, error)
}
