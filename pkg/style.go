package pkg

type Style struct {
	Name  string
	Apply func(string) string
}
