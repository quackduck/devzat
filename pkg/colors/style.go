package colors

type Style struct {
	Name  string
	Apply func(string) string
}
