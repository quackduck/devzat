package pronouns

import (
	"strings"
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

func (c *Command) Fn(linestring, u pkg.User) error {
	args := strings.Fields(line)

	if line == "" {
		u.Room().BotCast("Set pronouns by providing em or query a User's pronouns!")
		return
	}

	if len(args) == 1 && strings.HasPrefix(args[0], "@") {
		victim, ok := u.Room().FindUserByName(args[0][1:])
		if !ok {
			u.Room().BotCast("Who's that?")
			return
		}
		u.Room().BotCast(victim.Name + "'s pronouns are " + victim.displayPronouns())
		return
	}

	u.pronouns = strings.Fields(strings.ReplaceAll(strings.ToLower(line), "\n", ""))
	//u.changeColor(u.color) // refresh pronouns
	u.Room().BotCast(u.Name + " now goes by " + u.displayPronouns())
}
