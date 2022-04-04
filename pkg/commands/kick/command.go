package kick

import (
	i "devzat/pkg/interfaces"
	"devzat/pkg/models"
	"fmt"
)

const (
	name     = "kick"
	argsInfo = "<user>"
	info     = "kick a user from chat"
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
	if !u.IsAdmin() {
		u.Room().BotCast("Not authorized")
		return nil
	}

	victim, ok := u.Room().FindUserByName(line)
	if !ok {
		u.Room().Broadcast("", "User not found")
		return nil
	}

	wasKickedBy := fmt.Sprintf("has been kicked by %s", u.Name())
	wasKickedBy = u.Room().Colors().Red.Paint(wasKickedBy)
	msg := fmt.Sprintf("%s %s", victim.Name(), wasKickedBy)
	victim.Close(msg)

	return nil
}
