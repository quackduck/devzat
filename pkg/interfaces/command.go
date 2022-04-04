package interfaces

import "devzat/pkg/models"

type CommandFunc = func(rest string, u User) error

type Command interface {
	Name() string
	Fn(rest string, u User) error
	ArgsInfo() string
	Info() string
	Visibility() models.CommandVisibility
}
