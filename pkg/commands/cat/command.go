package cat

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

func (c *Command) Fn(line string, u *user.User) error {
	if line == "" {
		u.Room.Broadcast("", "usage: cat [-benstuv] [file ...]")
	} else if line == "README.md" {
		helpCMD(line, u)
	} else {
		u.Room.Broadcast("", "cat: "+line+": Permission denied")
	}
}
