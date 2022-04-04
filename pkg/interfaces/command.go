package interfaces

type CommandFunc = func(rest string, u User) error

type Command interface {
	Name() string
	Fn(rest string, u User) error
	ArgsInfo() string
	Info() string
	IsRest() bool // whatever the fuck this means...
	IsSecret() bool
}
