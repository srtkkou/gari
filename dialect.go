package gari

type (
	Dialect interface {
		QuerySuffix() string
		BindVar(i int) string
	}
)
