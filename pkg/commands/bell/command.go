package bell

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
	switch rest {
	case "off":
		u.Bell = false
		u.PingEverytime = false
		u.Room.Broadcast("", "bell off (never)")
	case "on":
		u.Bell = true
		u.PingEverytime = false
		u.Room.Broadcast("", "bell on (pings)")
	case "all":
		u.PingEverytime = true
		u.Room.Broadcast("", "bell all (every message)")
	case "", "status":
		if u.Bell {
			u.Room.Broadcast("", "bell on (pings)")
		} else if u.PingEverytime {
			u.Room.Broadcast("", "bell all (every message)")
		} else { // bell is off
			u.Room.Broadcast("", "bell off (never)")
		}
	default:
		u.Room.Broadcast(devbot, "your options are off, on and all")
	}
}
