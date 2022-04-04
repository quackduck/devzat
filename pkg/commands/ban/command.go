package ban

import (
	"devzat/pkg/interfaces"
	"devzat/pkg/models"
	"strings"
	"time"
)

const (
	name     = "ban"
	argsInfo = "<user>"
	info     = "ban a user"
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
	return models.CommandVisSecret
}

func (c *Command) Fn(line string, u interfaces.User) error {
	if !u.IsAdmin() {
		u.Room().BotCast("Not authorized")
		return nil
	}

	bannerName := u.Name()

	split := strings.Split(line, " ")
	if len(split) == 0 {
		u.Room().BotCast("Which User do you want to ban?")
		return nil
	}

	victim, found := u.Room().FindUserByName(split[0])
	if !found {
		u.Room().Broadcast("", "User not found")
		return nil
	}

	// check if the ban is for a certain duration
	hasDuration := len(split) > 1
	if !hasDuration {
		u.Room().Server().BanUser(bannerName, victim)
	}

	dur, err := time.ParseDuration(split[1])
	if err != nil {
		u.Room().BotCast("I couldn't parse that as a duration")
		return nil
	}

	u.Room().Server().BanUserForDuration(bannerName, victim, dur)

	return nil
}
