package tz

import (
	"devzat/pkg/user"
	"strings"
	"time"
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

func (c *Command) Fn(tzArg string, u *user.User) error {
	var err error
	if tzArg == "" {
		u.timezone = nil
		return
	}
	tzArgList := strings.Fields(tzArg)
	tz := tzArgList[0]

	switch strings.ToUpper(tz) {
	case "PST", "PDT":
		tz = "PST8PDT"
	case "CST", "CDT":
		tz = "CST6CDT"
	case "EST", "EDT":
		tz = "EST5EDT"
	case "MT":
		tz = "America/Phoenix"
	}
	u.timezone, err = time.LoadLocation(tz)
	if err != nil {
		u.Room.Broadcast(devbot, "Weird timezone you have there, use the format Continent/City, the usual US timezones (PST, PDT, EST, EDT...) or check nodatime.org/TimeZones!")
		return
	}
	if len(tzArgList) == 2 {
		u.FormatTime24 = tzArgList[1] == "24h"
	} else {
		u.FormatTime24 = false
	}
	u.Room.Broadcast(devbot, "Changed your timezone!")
}
