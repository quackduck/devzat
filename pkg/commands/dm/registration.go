package dm

import (
	"devchat/pkg"
	"fmt"
	"strings"
)

var _ pkg.CommandRegistration = &Command{} // static check that we implement the interface

type Command struct{}

func (r *Command) Name() string {
	return "=<User>"
}

func (r *Command) Command(rest string, u *pkg.User) error {
	bot := u.Room.Bot

	restSplit := strings.Fields(rest)
	if len(restSplit) < 2 {
		u.Writeln(bot.Name(), "You gotta have a message, mate")
		return nil
	}

	peer, ok := u.Room.FindUserByName(restSplit[0])
	if !ok {
		u.Writeln(bot.Name(), "No such person lol, who you wanna dm? (you might be in the wrong room)")
		return nil
	}

	msg := strings.TrimSpace(strings.TrimPrefix(rest, restSplit[0]))
	u.Writeln(peer.Name+" <- ", msg)

	if u == peer {
		foreverAlone := []string{
			"You must be really lonely, DMing yourself.",
			"Don't worry, I won't judge :wink:",
			"srsly?",
			"what an idiot",
		}

		u.Room.Bot.Respond(foreverAlone, 30)

		return nil
	}

	peer.Writeln(fmt.Sprintf("%s -> ", u.Name), msg)

	return nil
}

func (r *Command) ArgsInfo() string {
	return "<msg>"
}

func (r *Command) Info() string {
	return "DM <User> with <msg>"
}

func (r *Command) IsRest() bool {
	return false
}

func (r *Command) IsSecret() bool {
	return false
}
