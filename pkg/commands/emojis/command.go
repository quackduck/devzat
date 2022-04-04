package emojis

import (
	i "devzat/pkg/interfaces"
	"devzat/pkg/models"
)

const (
	name     = "emojis"
	argsInfo = ""
	info     = "see a list of emojis"
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
	return models.CommandVisNormal
}

func (c *Command) Fn(_ string, u i.User) error {
	u.Room().BotCast("Check out https\\://github.com/ikatyang/emoji-cheat-sheet")

	return nil
}
