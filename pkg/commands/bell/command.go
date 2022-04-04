package bell

import (
	"devzat/pkg/interfaces"
	"devzat/pkg/models"
	"strings"
)

const (
	name     = ""
	argsInfo = "[on|off|all]"
	info     = "ANSI bell on pings (on), never (off) or for every message (all)"
)

type bellLevel = int

const (
	off bellLevel = iota
	on
	all
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

func (c *Command) Fn(rest string, u interfaces.User) error {
	lookup := map[string]bellLevel{
		"off": off,
		"on":  on,
		"all": all,
	}

	switch lookup[strings.ToLower(rest)] {
	case off:
		u.SetBell(false)
		u.SetPingEverytime(false)
		u.Room().Broadcast("", "bell off (never)")
	case on:
		u.SetBell(true)
		u.SetPingEverytime(false)
		u.Room().Broadcast("", "bell on (pings)")
	case all:
		u.SetPingEverytime(true)
		u.Room().Broadcast("", "bell all (every message)")
	default:
		if u.Bell() {
			u.Room().Broadcast("", "bell on (pings)")
		} else if u.PingEverytime() {
			u.Room().Broadcast("", "bell all (every message)")
		} else { // bell is off
			u.Room().Broadcast("", "bell off (never)")
		}
		u.Room().BotCast("your options are off, on and all")
	}

	return nil
}
