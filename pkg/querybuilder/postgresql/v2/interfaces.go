package querybuilder

type Node interface {
	WriteTo(b *Builder)
}
