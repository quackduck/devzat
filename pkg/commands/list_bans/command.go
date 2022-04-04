package list_bans

import (
	i "devzat/pkg/interfaces"
	"devzat/pkg/models"
	"strconv"
)

const (
	name     = "lsbans"
	argsInfo = ""
	info     = "List banned IDs"
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
	bans := u.Room().Server().GetBanList()

	msg := "Printing bans by ID:  \n"
	for i := 0; i < len(bans); i++ {
		cyan := u.Room().Colors().Yellow
		msg += cyan.Cyan(strconv.Itoa(i+1)) + ". " + bans[i].ID + "  \n"
	}

	u.Room().BotCast(msg)

	return nil
}
