package man

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

func (c *Command) Fn(rest string, u pkg.User) error {
	devbot := u.Room().Bot().Name()
	if rest == "" {
		u.Room().BotCast("What command do you want help with?")
		return
	}

	for _, c := range allcmds {
		if c.Name == rest {
			u.Room().BotCast("Usage: " + c.Name + " " + c.argsInfo + "  \n" + c.info)
			return
		}
	}
	u.Room().Broadcast("", "This system has been minimized by removing packages and content that are not required on a system that users do not log into.\n\nTo restore this content, including manpages, you can run the 'unminimize' command. You will still need to ensure the 'man-db' package is installed.")
}
