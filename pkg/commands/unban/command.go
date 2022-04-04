package unban

import (
	"devzat/pkg/interfaces"
	"devzat/pkg/models"
	"fmt"
)

const (
	Name     = "unban"
	argsInfo = "<user>"
	info     = "unban a user"
)

const (
	fmtNotAuthorized = "%s is not authorized to use the %s command"
)

type Command struct{}

func (c *Command) Name() string {
	return Name
}

func (c *Command) ArgsInfo() string {
	return argsInfo
}

func (c *Command) Info() string {
	return info
}

func (c *Command) Visibility() models.CommandVisibility {
	return models.CommandVisSecret
}

func (c *Command) Fn(toUnban string, u interfaces.User) error {
	if !u.IsAdmin() {
		u.Room().BotCast(fmt.Sprintf(fmtNotAuthorized, u.Name(), Name))

		return nil
	}

	if err := u.Room().Server().UnbanUser(toUnban); err != nil {
		return err
	}

	u.Room().BotCast("Unbanned person: " + toUnban)

	return nil
}
