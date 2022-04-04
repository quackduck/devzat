package bell

import (
	"devzat/pkg/interfaces"
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

func (c *Command) Fn(rest string, u interfaces.User) error {
	switch rest {
	case "off":
		u.SetBell(false)
		u.SetPingEverytime(false)
		u.Room().Broadcast("", "bell off (never)")
	case "on":
		u.SetBell(true)
		u.SetPingEverytime(false)
		u.Room().Broadcast("", "bell on (pings)")
	case "all":
		u.SetPingEverytime(true)
		u.Room().Broadcast("", "bell all (every message)")
	case "", "status":
		if u.Bell() {
			u.Room().Broadcast("", "bell on (pings)")
		} else if u.PingEverytime() {
			u.Room().Broadcast("", "bell all (every message)")
		} else { // bell is off
			u.Room().Broadcast("", "bell off (never)")
		}
	default:
		u.Room().BotCast("your options are off, on and all")
	}

	return nil
}
