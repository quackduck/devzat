package list_bans

import (
	"strconv"
)

const (
	name     = ""
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

func (c *Command) Fn(_ string, u pkg.User) error {
	msg := "Printing bans by ID:  \n"
	for i := 0; i < len(bans); i++ {
		msg += cyan.Cyan(strconv.Itoa(i+1)) + ". " + bans[i].ID + "  \n"
	}
	u.Room().BotCast(msg)
}
