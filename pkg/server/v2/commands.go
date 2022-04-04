package v2

import (
	"devzat/pkg/commands/admins"
	"devzat/pkg/commands/ascii_art"
	"devzat/pkg/commands/ban"
	"devzat/pkg/commands/bell"
	"devzat/pkg/commands/cat"
	"devzat/pkg/commands/cd"
	"devzat/pkg/commands/clear"
	"devzat/pkg/commands/cmd"
	"devzat/pkg/commands/cmd_rest"
	"devzat/pkg/commands/cmd_room"
	"devzat/pkg/commands/color"
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
	"devzat/pkg/interfaces"
)

func (s *Server) initCommands() {
	commands := []interfaces.Command{
		&dm.Command{},
		&admins.Command{},
		&ascii_art.Command{},
		&ban.Command{},
		&bell.Command{},
		&cat.Command{},
		&cd.Command{},
		&clear.Command{},
		&cmd.Command{},
		&cmd_rest.Command{},
		&cmd_room.Command{},
		&color.Command{},
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

	commandMap := make(map[string]interfaces.Command)

	for _, c := range commands {
		commandMap[c.Name()] = c
	}
}

func (s *Server) Commands() map[string]interfaces.Command {
	return s.commands
}

func (s *Server) GetCommand(name string) (interfaces.Command, bool) {
	c, found := s.commands[name]
	return c, found
}
