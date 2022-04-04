package rm

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
	if line == "" {
		u.Room().Broadcast("", `usage: rm [-f | -i] [-dPRrvW] file ...
unlink file`)
	} else {
		u.Room().Broadcast("", "rm: "+line+": Permission denied, sucker")
	}
}
