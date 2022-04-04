package server

import (
	"devzat/pkg/commands/admins"
	"devzat/pkg/commands/ascii_art"
	"devzat/pkg/commands/ban"
	"devzat/pkg/commands/bell"
	"devzat/pkg/commands/cat"
	"devzat/pkg/commands/cd"
	"devzat/pkg/commands/clear"
	"devzat/pkg/commands/cmd"
	"devzat/pkg/commands/color"
	"devzat/pkg/commands/colour"
	"devzat/pkg/commands/dm"
	"devzat/pkg/commands/emojis"
	"devzat/pkg/commands/example_code"
	"devzat/pkg/commands/exit"
	"devzat/pkg/commands/hang"
	"devzat/pkg/commands/help"
	"devzat/pkg/commands/id"
	"devzat/pkg/commands/kick"
	"devzat/pkg/commands/list_bans"
	"devzat/pkg/commands/ls"
	"devzat/pkg/commands/man"
	"devzat/pkg/commands/nick"
	"devzat/pkg/commands/people"
	"devzat/pkg/commands/pronouns"
	"devzat/pkg/commands/pwd"
	"devzat/pkg/commands/rm"
	"devzat/pkg/commands/shrug"
	"devzat/pkg/commands/theme"
	"devzat/pkg/commands/tic"
	"devzat/pkg/commands/tz"
	"devzat/pkg/commands/unban"
	"devzat/pkg/commands/users"

	i "devzat/pkg/interfaces"
)

type commandRegistry struct {
	commands map[string]i.Command
}

func (cr *commandRegistry) init() {
	commands := []i.Command{
		&dm.Command{},
		&admins.Command{},
		&ascii_art.Command{},
		&ban.Command{},
		&bell.Command{},
		&cat.Command{},
		&cd.Command{},
		&clear.Command{},
		&cmd.Command{},
		&color.Command{},
		&colour.Command{},
		&emojis.Command{},
		&example_code.Command{},
		&exit.Command{},
		&hang.Command{},
		&help.Command{},
		&id.Command{},
		&kick.Command{},
		&list_bans.Command{},
		&ls.Command{},
		&man.Command{},
		&nick.Command{},
		&people.Command{},
		&pronouns.Command{},
		&pwd.Command{},
		&rm.Command{},
		&shrug.Command{},
		&theme.Command{},
		&tic.Command{},
		&tz.Command{},
		&unban.Command{},
		&users.Command{},
	}

	cr.commands = make(map[string]i.Command)

	for _, c := range commands {
		cr.AddCommand(c)
	}
}

func (cr *commandRegistry) AddCommand(c i.Command) {
	if _, found := cr.commands[c.Name()]; !found {
		cr.commands[c.Name()] = c
	}
}

func (s *Server) Commands() map[string]i.Command {
	dupe := make(map[string]i.Command)

	for name := range s.commands {
		dupe[name] = s.commands[name]
	}

	return dupe
}

func (s *Server) GetCommand(name string) (i.Command, bool) {
	c, found := s.commands[name]
	return c, found
}
