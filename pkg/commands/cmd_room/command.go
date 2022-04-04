package cmd_room

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
	u.Writeln(u.DMTarget.Name+" <- ", line)
	if u == u.DMTarget {
		devbotRespond(u.Room, []string{"You must be really lonely, DMing yourself.",
			"Don't worry, I won't judge :wink:",
			"srsly?",
			"what an idiot"}, 30)
		return
	}
	u.DMTarget.writeln(u.Name+" -> ", line)
}
