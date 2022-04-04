package ban

import (
	"devzat/pkg/interfaces"
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

func (c *Command) Fn(line string, u interfaces.User) error {
	isAdmin, err := u.Room().IsAdmin(u)
	if err != nil {
		return err
	}

	bannerName := u.Name()

	if !isAdmin {
		u.Room().BotCast("Not authorized")
		return nil
	}

	split := strings.Split(line, " ")
	if len(split) == 0 {
		u.Room().BotCast("Which User do you want to ban?")
		return nil
	}
	victim, ok := u.Room().FindUserByName(split[0])
	if !ok {
		u.Room().Broadcast("", "User not found")
		return nil
	}

	srv := u.Room().Server()

	// check if the ban is for a certain duration
	hasDuration := len(split) > 1
	if !hasDuration {
		srv.BanUser(bannerName, victim)
	}

	dur, err := time.ParseDuration(split[1])
	if err != nil {
		u.Room().BotCast("I couldn't parse that as a duration")
		return nil
	}

	srv.BanUserForDuration(bannerName, victim, dur)

	return nil
}
