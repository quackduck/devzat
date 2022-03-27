package pwd

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
	if u.Messaging != nil {
		u.Writeln("", u.Messaging.Name)
		u.Messaging.writeln("", u.Messaging.Name)
	} else {
		u.Room.Broadcast("", u.Room.Name)
	}
}
