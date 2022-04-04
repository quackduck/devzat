package util

import (
	"math"
	"strings"

	markdown "github.com/quackduck/go-term-markdown"
)

func MarkdownRender(a string, beforeMessageLen int, lineWidth int) string {
	if strings.Contains(a, "![") && strings.Contains(a, "](") {
		lineWidth = int(math.Min(float64(lineWidth/2), 200)) // max image width is 200
	}

	md := string(markdown.Render(a, lineWidth-(beforeMessageLen), 0))
	md = strings.TrimSuffix(md, "\n")
	split := strings.Split(md, "\n")

	for i := range split {
		if i == 0 {
			continue // the first line will automatically be padded
		}

		split[i] = strings.Repeat(" ", beforeMessageLen) + split[i]
	}

	if len(split) == 1 {
		return md
	}

	return strings.Join(split, "\n")
}
