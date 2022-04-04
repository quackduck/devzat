package nick

import (
	i "devzat/pkg/interfaces"
	"devzat/pkg/models"
)

const (
	name     = "nick"
	argsInfo = "<foobar>"
	info     = "set your nickname (no profanity)"
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

func (c *Command) Fn(line string, u i.User) error {
	return u.SetNick(line)
}
