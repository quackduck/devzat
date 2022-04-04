package v2

import "github.com/acarl005/stripansi"

func lenString(a string) int {
	return len([]rune(stripansi.Strip(a)))
}
