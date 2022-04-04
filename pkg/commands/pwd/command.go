package pwd

import (
	i "devzat/pkg/interfaces"
	"devzat/pkg/models"
)

const (
	name     = "pwd"
	argsInfo = ""
	info     = "show all users in the current room"
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

func (c *Command) Fn(_ string, u i.User) error {
	if u.DMTarget() == nil {
		u.Room().Broadcast("", u.Room().Name())
		return nil
	}

	u.Writeln("", u.DMTarget().Name())
	u.DMTarget().Writeln("", u.DMTarget().Name())

	return nil
}
