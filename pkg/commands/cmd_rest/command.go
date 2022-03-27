package cmd_rest

import "devzat/pkg/user"

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

func (c *Command) Fn(_ string, u *user.User) error {
	u.Room.Broadcast("", "The rest  \n"+autogenCommands(cmdsRest))
}
