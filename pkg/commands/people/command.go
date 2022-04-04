package people

import (
	_ "embed"

	i "devzat/pkg/interfaces"
	"devzat/pkg/models"
)

const (
	name     = ""
	argsInfo = ""
	info     = ""
)

//go:embed banner.txt
var banner string

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
	return models.CommandVisLow
}

func (c *Command) Fn(_ string, u i.User) error {
	u.Room().Broadcast("", banner)

	return nil
}
