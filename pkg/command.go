package pkg

import "devzat/pkg/user"

type CommandFunc = func(rest string, u *user.User) error

type Command struct {
	name     string
	run      CommandFunc
	argsInfo string
	info     string
}

type CommandRegistration interface {
	Name() string
	Fn(rest string, u *user.User) error
	ArgsInfo() string
	Info() string
	IsRest() bool // whatever the fuck this means...
	IsSecret() bool
}
