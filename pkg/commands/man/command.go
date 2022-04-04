package man

import (
	i "devzat/pkg/interfaces"
	"devzat/pkg/models"
	"fmt"
)

const (
	name     = "man"
	argsInfo = "<command>"
	info     = "show usage info for a command"
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
	const fmtUsage = "Usage: %s %s\n%s\n"

	if line == "" {
		u.Room().BotCast("What command do you want help with?")

		return nil
	}

	for _, command := range u.Room().Server().Commands() {
		if command.Name() == line {
			u.Room().BotCast(fmt.Sprintf(fmtUsage, command.Name(), command.ArgsInfo(), command.Info()))
			return nil
		}
	}

	return nil
}
