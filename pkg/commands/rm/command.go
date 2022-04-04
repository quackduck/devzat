package rm

import (
	i "devzat/pkg/interfaces"
	"devzat/pkg/models"
)

const (
	name     = "rm"
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
	return models.CommandVisHidden
}

func (c *Command) Fn(line string, u i.User) error {
	if line == "" {
		u.Room().Broadcast("", `usage: rm [-f | -i] [-dPRrvW] file ...
unlink file`)
	} else {
		u.Room().Broadcast("", "rm: "+line+": Permission denied, sucker")
	}

	return nil
}
