package example_code

const (
	name     = "=<user>"
	argsInfo = "<msg>"
	info     = "DirectMessage <User> with <msg>"
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
	if line == "big" {
		u.Room().BotCast("```go\npackage mainRoom\n\nimport \"fmt\"\n\nfunc sum(nums ...int) {\n    fmt.Print(nums, \" \")\n    total := 0\n    for _, num := range nums {\n        total += num\n    }\n    fmt.Println(total)\n}\n\nfunc mainRoom() {\n\n    sum(1, 2)\n    sum(1, 2, 3)\n\n    nums := []int{1, 2, 3, 4}\n    sum(nums...)\n}\n```")
		return
	}
	u.Room().BotCast("\n```go\npackage mainRoom\nimport \"fmt\"\nfunc mainRoom() {\n   fmt.Println(\"Example!\")\n}\n```")
}
