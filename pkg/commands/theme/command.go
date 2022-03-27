package theme

import (
	"strings"

	chromastyles "github.com/alecthomas/chroma/styles"
	markdown "github.com/quackduck/go-term-markdown"

	"devzat/pkg/user"
)

const (
	name     = ""
	argsInfo = ""
	info     = ""
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

func (c *Command) IsRest() bool {
	return false
}

func (c *Command) IsSecret() bool {
	return false
}

func (c *Command) Fn(line string, u *user.User) error {
	if line == "list" {
		u.Room.Broadcast(devbot, "Available themes: "+strings.Join(chromastyles.Names(), ", "))
		return
	}
	for _, name := range chromastyles.Names() {
		if name == line {
			markdown.CurrentTheme = chromastyles.Get(name)
			u.Room.Broadcast(devbot, "Theme set to "+name)
			return
		}
	}
	u.Room.Broadcast(devbot, "What theme is that? Use theme list to see what's available.")
}
