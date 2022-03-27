package pronouns

import (
	"devzat/pkg/user"
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

func (c *Command) Fn(line string, u *user.User) error {
	args := strings.Fields(line)

	if line == "" {
		u.Room.Broadcast(devbot, "Set pronouns by providing em or query a User's pronouns!")
		return
	}

	if len(args) == 1 && strings.HasPrefix(args[0], "@") {
		victim, ok := u.Room.FindUserByName(args[0][1:])
		if !ok {
			u.Room.Broadcast(devbot, "Who's that?")
			return
		}
		u.Room.Broadcast(devbot, victim.Name+"'s pronouns are "+victim.displayPronouns())
		return
	}

	u.pronouns = strings.Fields(strings.ReplaceAll(strings.ToLower(line), "\n", ""))
	//u.changeColor(u.color) // refresh pronouns
	u.Room.Broadcast(devbot, u.Name+" now goes by "+u.displayPronouns())
}
