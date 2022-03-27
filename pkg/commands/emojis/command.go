package emojis

import (
	"devzat/pkg/user"
)

const (
	name     = "=<user>"
	argsInfo = "<msg>"
	info     = "DirectMessage <User> with <msg>"
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
	u.Room.Broadcast(devbot, "Check out https\\://github.com/ikatyang/emoji-cheat-sheet")
}
