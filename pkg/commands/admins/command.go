package admins

import (
	"devzat/pkg/interfaces"
	"devzat/pkg/models"
	"fmt"
)

const (
	name     = "admins"
	argsInfo = ""
	info     = "Print the ID (hashed key) for all admins"
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

const (
	fmtConcat = "%s%s"
	fmtAdmin  = "%s: %s\n"
)

func (c *Command) Fn(_ string, u interfaces.User) error {
	admins, err := u.Room().Server().GetAdmins()
	if err != nil {
		return err
	}

	strAdmins := ""
	for adminName := range admins {
		line := fmt.Sprintf(fmtAdmin, admins[adminName], adminName)
		strAdmins = fmt.Sprintf(fmtConcat, strAdmins, line)
	}

	msg := fmt.Sprintf("Admins:\n%s", strAdmins)

	u.Room().BotCast(msg)

	return nil
}
