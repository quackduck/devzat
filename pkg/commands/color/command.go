package color

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

func (c *Command) Fn(rest string, u *user.User) error {
	devbot := u.Room.Bot.Name()
	if rest == "which" {
		u.Room.Broadcast(devbot, "fg: "+u.color+" & bg: "+u.colorBG)
	} else if err := u.changeColor(rest); err != nil {
		u.Room.Broadcast(devbot, err.Error())
	}
}
