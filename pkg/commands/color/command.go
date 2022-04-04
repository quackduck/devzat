package color

import (
	i "devzat/pkg/interfaces"
	"devzat/pkg/models"
	"fmt"
)

const (
	name     = "color"
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

func (c *Command) Visibility() models.CommandVisibility {
	return models.CommandVisHidden
}

func (c *Command) Fn(line string, u i.User) error {
	if line == "which" {
		fmtMsg := "fg: %s & bg: %s"
		msg := fmt.Sprintf(fmtMsg, u.ForegroundColor(), u.BackgroundColor())

		u.Room().BotCast(msg)

		return nil
	}

	if err := u.ChangeColor(line); err != nil {
		u.Room().BotCast(err.Error())
	}

	return nil
}
