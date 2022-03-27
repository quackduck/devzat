package pkg

type CommandFunc = func(rest string, u *User) error

type CommandRegistration interface {
	Name() string
	Command(rest string, u *User) error
	ArgsInfo() string
	Info() string
	IsRest() bool // whatever the fuck this means...
	IsSecret() bool
}
