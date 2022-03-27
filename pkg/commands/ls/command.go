package ls

import "devzat/pkg/user"

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

func (c *Command) Fn(rest string, u *user.User) error {
	devbot := u.Room.Bot.Name()
	if len(rest) > 0 && rest[0] == '#' {
		if r, ok := rooms[rest]; ok {
			usersList := ""
			for _, us := range r.users {
				usersList += us.Name + blue.Paint("/ ")
			}
			u.Room.broadcast("", usersList)
			return
		}
	}
	if rest != "" {
		u.Room.Broadcast("", "ls: "+rest+" Permission denied")
		return
	}
	roomList := ""
	for _, r := range rooms {
		roomList += blue.Paint(r.Name + "/ ")
	}
	usersList := ""
	for _, us := range u.Room.users {
		usersList += us.Name + blue.Paint("/ ")
	}
	u.Room.Broadcast("", "README.md "+usersList+roomList)
}
