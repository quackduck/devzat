package example_code

import (
	i "devzat/pkg/interfaces"
	"devzat/pkg/models"
	"fmt"
)

const (
	name     = "eg-code"
	argsInfo = "[big]"
	info     = "Example syntax-highlighted code"
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
	const fmtCode = "```%s```"
	msg := ""

	switch line {
	case "big":
		msg = fmt.Sprintf(fmtCode, exampleBig)
	default:
		msg = fmt.Sprintf(fmtCode, exampleSmall)
	}

	u.Room().BotCast(msg)
	return nil
}
