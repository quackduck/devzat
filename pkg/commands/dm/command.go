package dm

import (
	"devzat/pkg/interfaces"
	"fmt"
	"strings"
)

const (
	name     = "=<user>"
	argsInfo = "<msg>"
	info     = "DirectMessage <User> with <msg>"
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

func (c *Command) Fn(rest string, u interfaces.User) error {
	bot := u.Room().Bot

	restSplit := strings.Fields(rest)
	if len(restSplit) < 2 {
		u.Writeln(bot.Name(), "You gotta have a message, mate")
		return nil
	}

	peer, ok := u.Room().FindUserByName(restSplit[0])
	if !ok {
		u.Writeln(bot.Name(), "No such person lol, who you wanna dm? (you might be in the wrong room)")
		return nil
	}

	msg := strings.TrimSpace(strings.TrimPrefix(rest, restSplit[0]))
	u.Writeln(peer.Name()+" <- ", msg)

	if u == peer {
		foreverAlone := []string{
			"You must be really lonely, DMing yourself.",
			"Don't worry, I won't judge :wink:",
			"srsly?",
			"what an idiot",
		}

		u.Room().Bot.Respond(foreverAlone, 30)

		return nil
	}

	peer.Writeln(fmt.Sprintf("%s -> ", u.Name), msg)

	return nil
}
