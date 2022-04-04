package clear

import (
	"devzat/pkg/interfaces"
	"devzat/pkg/models"
)

const (
	Name     = "clear"
	argsInfo = ""
	info     = ""
)

type Command struct{}

func (c *Command) Name() string {
	return Name
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

func (c *Command) Fn(_ string, u interfaces.User) error {
	_, err := u.Term().Write([]byte("\033[H\033[2J"))

	return err
}
