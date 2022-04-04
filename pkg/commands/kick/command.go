package kick

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
	victim, ok := u.Room().FindUserByName(line)
	if !ok {
		u.Room().Broadcast("", "User not found")
		return
	}
	if !checkIsAdmin(u) {
		u.Room().BotCast("Not authorized")
		return
	}
	victim.close(victim.Name + red.Paint(" has been kicked by ") + u.Name)
}
