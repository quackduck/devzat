package theme

import (
	i "devzat/pkg/interfaces"
	"devzat/pkg/models"
	"strings"

	chromastyles "github.com/alecthomas/chroma/styles"
	markdown "github.com/quackduck/go-term-markdown"
)

const (
	name     = "theme"
	argsInfo = "<list|name>"
	info     = "list or set the markdown theme"
)

type Command struct{}

func (c *Command) Name() string {
	return name
}

func (c *Command) ArgsInfo() string {
	return argsInfo
}

func (c *Command) Info() string {
	return info
}

func (c *Command) Visibility() models.CommandVisibility {
	return models.CommandVisSecret
}

func (c *Command) Fn(line string, u i.User) error {
	if line == "list" {
		u.Room().BotCast("Available themes: " + strings.Join(chromastyles.Names(), ", "))
		return nil
	}

	if !u.IsAdmin() {
		u.Room().BotCast("Not authorized")
		return nil
	}

	for _, cnames := range chromastyles.Names() {
		if cnames == line {
			markdown.CurrentTheme = chromastyles.Get(cnames)
			u.Room().BotCast("Theme set to " + cnames)
			return nil
		}
	}

	u.Room().BotCast("What theme is that? Use theme list to see what's available.")

	return nil
}
