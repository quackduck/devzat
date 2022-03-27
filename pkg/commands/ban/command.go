package ban

import (
	"strings"
	"time"

	"devzat/pkg/user"
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
	if !checkIsAdmin(u) {
		u.Room.Broadcast(devbot, "Not authorized")
		return
	}
	split := strings.Split(line, " ")
	if len(split) == 0 {
		u.Room.Broadcast(devbot, "Which User do you want to ban?")
		return
	}
	victim, ok := u.Room.FindUserByName(split[0])
	if !ok {
		u.Room.Broadcast("", "User not found")
		return
	}
	// check if the ban is for a certain duration
	if len(split) > 1 {
		dur, err := time.ParseDuration(split[1])
		if err != nil {
			u.Room.Broadcast(devbot, "I couldn't parse that as a duration")
			return
		}
		bans = append(bans, server.ban{victim.addr, victim.id})
		victim.close(victim.Name + " has been banned by " + u.Name + " for " + dur.String())
		go func(id string) {
			time.Sleep(dur)
			unbanIDorIP(id)
		}(victim.id) // evaluate id now, call unban with that value later
		return
	}
	banUser(u.Name, victim)
}
