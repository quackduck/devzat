package id

import (
	i "devzat/pkg/interfaces"
	"devzat/pkg/models"
)

const (
	name     = "id"
	argsInfo = "<user>"
	info     = "Get a unique ID for a user (hashed key)"
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
	return models.CommandVisLow
}

func (c *Command) Fn(line string, u i.User) error {
	victim, ok := u.Room().FindUserByName(line)
	if !ok {
		u.Room().Broadcast("", "User not found")
		return nil
	}

	u.Room().Broadcast("", victim.ID())

	return nil
}
