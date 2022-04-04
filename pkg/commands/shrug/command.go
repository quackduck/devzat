package shrug

import (
	i "devzat/pkg/interfaces"
	"devzat/pkg/models"
	"fmt"
)

const (
	name     = "shrug"
	argsInfo = "[message]"
	info     = `¯\_(ツ)_/¯`
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
	const fmtShrug = `%s ¯\_(ツ)_/¯`
	u.Room().Broadcast(u.Name(), fmt.Sprintf(fmtShrug, line))

	return nil
}
