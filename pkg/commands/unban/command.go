package unban

import (
	"devzat/pkg/interfaces"
	"fmt"
)

const (
	Name     = "unban"
	argsInfo = ""
	info     = ""
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

func (c *Command) IsRest() bool {
	return false
}

func (c *Command) IsSecret() bool {
	return false
}

func (c *Command) Fn(toUnban string, u interfaces.User) error {
	isAdmin, errCheckAdmin := u.Room().CheckIsAdmin(u)
	if errCheckAdmin != nil {
		return fmt.Errorf("could not unban: %v", errCheckAdmin)
	}

	if !isAdmin {
		u.Room().BotCast(fmt.Sprintf(fmtNotAuthorized, u.Name(), Name))

		return nil
	}

	if err := u.Room().Server.UnbanUser(toUnban); err != nil {
		return err
	}

	u.Room().BotCast("Unbanned person: " + toUnban)

	return nil
}
