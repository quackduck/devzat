package exit

import (
	i "devzat/pkg/interfaces"
	"devzat/pkg/models"
	"fmt"
)

const (
	name     = "exit"
	argsInfo = ""
	info     = "leave the chat"
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

func (c *Command) Fn(_ string, u i.User) error {
	hasLeft := u.Room().Colors().Red.Paint("has left the chat")
	u.Close(fmt.Sprintf("%s %s", u.Name(), hasLeft))

	return nil
}
