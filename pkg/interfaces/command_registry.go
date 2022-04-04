package interfaces

type commandRegistry interface {
	AddCommand(Command)
	GetCommand(name string) (Command, bool)
	Commands() map[string]Command
}
