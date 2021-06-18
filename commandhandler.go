package main

import (
	"fmt"
	"strings"
)

type commandInfo struct {
	name          string
	description   string
	callable      func(user *user, args []string)
	echoLevel     int
	requiresAdmin bool
	aliases       []string
}

var (
	commands = make([]commandInfo, 0)
)

func processMessage(u *user, message string) {
	message = strings.TrimSpace(message)

	if message == "" {
		return // Don't allow empty messages
	}
	splitted := strings.Split(message, " ")
	if strings.HasPrefix(splitted[0], "./") {
		if u.slack {
			u.room.broadcast(devbot, "Slack users can't use commands", false)
			return
		}
		commandName := strings.TrimPrefix(splitted[0], "./")
		for _, command := range commands {
			if command.name == commandName {
				// Command found
				runCommand(u, command, splitted, message)
				return
			}
			if command.aliases != nil {
				for _, alias := range command.aliases {
					if alias == commandName {
						// Command found by alias
						runCommand(u, command, splitted, message)
						return
					}
				}
			}

		}
		u.writeln(u.name, message)
		u.system("Command not found..? Check ./help for a list of commands")
		return
	}
	devbotChat(u.room, message, true)
	if !u.slack {
		// Slack already sends their messages, this would cause 2 messages to be sent
		u.sendMessage(message)
	}
	triggerEasterEggs(u, message)
}
func handleCommandCrash(u *user) {
	err := recover()
	if err != nil {
		u.system(fmt.Sprintf("Oh no, something borked! Please create a issue on github.com/quackduck/devzat/issues and include exactly what command you ran. Some developer jumbo: %s", err))
		fmt.Print(err)
	}

}
func runCommand(u *user, command commandInfo, splitted []string, message string) {
	defer handleCommandCrash(u)
	if command.echoLevel == 1 {
		u.writeln(u.name, message)
	}
	if command.echoLevel == 2 {
		u.sendMessage(message)
	}

	if command.requiresAdmin && !auth(u) {
		u.writeln(devbot, "This command can only be ran by admins")
		return
	}
	commandArgs := append(splitted[:0], splitted[1:]...)
	command.callable(u, commandArgs)
}
