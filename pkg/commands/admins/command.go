package admins

import (
	"devzat/pkg/interfaces"
)

const (
	name     = "admins"
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

func (c *Command) IsRest() bool {
	return false
}

func (c *Command) IsSecret() bool {
	return false
}

func (c *Command) Fn(_ string, u interfaces.User) error {
	admins, err := u.Room().Server().GetAdmins()
	if err != nil {
		return err
	}

	msg := "Admins:  \n"
	for i := range admins {
		msg += admins[i] + ": " + i + "  \n"
	}

	u.Room().BotCast(msg)

	return nil
}
