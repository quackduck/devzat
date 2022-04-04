package pronouns

import (
	i "devzat/pkg/interfaces"
	"devzat/pkg/models"
	"fmt"
	"strings"
)

const (
	name     = "pronouns"
	argsInfo = "<[@user] []string>"
	info     = "set pronouns (admin can set others pronouns"
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
	args := strings.Fields(line)

	if line == "" {
		u.Room().BotCast("Set pronouns by providing em or query a User's pronouns!")

		return nil
	}

	if u.IsAdmin() && len(args) > 1 && strings.HasPrefix(args[0], "@") {
		target, ok := u.Room().FindUserByName(args[0][1:])
		if !ok {
			u.Room().BotCast("Who's that?")

			return nil
		}

		target.SetPronouns(strings.Fields(strings.ReplaceAll(strings.ToLower(line), "\n", ""))...)
		msg := fmt.Sprintf("%s now goes by %s", target.Name(), target.DisplayPronouns())
		u.Room().BotCast(msg)

		return nil
	}

	u.SetPronouns(strings.Fields(strings.ReplaceAll(strings.ToLower(line), "\n", ""))...)
	msg := fmt.Sprintf("%s now goes by %s", u.Name(), u.DisplayPronouns())

	u.Room().BotCast(msg)

	return nil
}
