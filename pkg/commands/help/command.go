package help

import (
	i "devzat/pkg/interfaces"
	"devzat/pkg/models"
	_ "embed"
)

//go embed:help.txt
var defaultHelpMessage string

const (
	name     = "help"
	argsInfo = "<msg>"
	info     = "DirectMessage <User> with <msg>"
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
	u.Room().Broadcast("", defaultHelpMessage)

	return nil
}
