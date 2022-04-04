package tz

import (
	i "devzat/pkg/interfaces"
	"devzat/pkg/models"
	"strings"
)

const (
	name     = "tz"
	argsInfo = "<PST|PDT|CST|CDT|EST|EDT|MT>"
	info     = "set your timezone"
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

func (c *Command) Fn(line string, u i.User) error {
	if line == "" {
		return nil
	}

	tzArgList := strings.Fields(line)

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

	u.SetTimeZone(tz)

	if len(tzArgList) == 2 {
		u.SetFormatTime24(tzArgList[1] == "24h")
	}

	u.Room().BotCast("Changed your timezone")

	return nil
}
