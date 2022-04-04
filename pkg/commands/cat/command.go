package cat

import (
	"devzat/pkg/interfaces"
	"devzat/pkg/models"
)

const (
	name     = "cat"
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

func (c *Command) Visibility() models.CommandVisibility {
	return models.CommandVisNormal
}

func (c *Command) Fn(line string, u interfaces.User) error {
	switch line {
	case "":
		u.Room().Broadcast("", "usage: cat [-benstuv] [file ...]")
	case "README.md":
		if cmd, found := u.Room().Server().GetCommand("help"); found {
			return cmd.Fn(line, u)
		}
	default:
		u.Room().Broadcast("", "cat: "+line+": Permission denied")
	}

	return nil
}
