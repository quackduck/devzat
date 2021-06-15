package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/acarl005/stripansi"
)

func registerCommands() {
	commands = append(commands, commandInfo{"clear", "Clears your terminal", clearCommand, false, false})
	commands = append(commands, commandInfo{"msg", "Sends a private message to someone", messageCommand, false, false})
	commands = append(commands, commandInfo{"users", "Gets a list of the active users", usersCommand, true, false})
	commands = append(commands, commandInfo{"all", "Gets a list of all users who has ever connected", allCommand, true, false})
	commands = append(commands, commandInfo{"exit", "Kicks you out of the chat incase your client was bugged", exitCommand, false, false})
	commands = append(commands, commandInfo{"bell", "Toggles notifications when you get pinged", bellCommand, true, false})
}

func clearCommand(u *user, _ []string) {
	u.term.Write([]byte("\033[H\033[2J"))
}

func messageCommand(u *user, args []string) {
	if len(args) < 1 {
		u.writeln(devbot, "Please provide a user to send a message to")
		return
	}

	if len(args) < 2 {
		u.writeln(devbot, "Please provide a message to send")
	}
	peer, ok := findUserByName(u.room, args[0])
	if !ok {
		u.writeln(devbot, "The user was not found, maybe they are in another room?")
		return
	}
	message := strings.Join(append(args[:0], args[1:]...), " ")
	peer.writeln(u.name+" -> ", message)
	u.writeln(u.name+" <- ", message)
}

func usersCommand(u *user, _ []string) {
	u.writeln(devbot, printUsersInRoom(u.room))
}

func allCommand(u *user, _ []string) {
	names := make([]string, 0, len(allUsers))
	for _, name := range allUsers {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		return strings.ToLower(stripansi.Strip(names[i])) < strings.ToLower(stripansi.Strip(names[j]))
	})
	u.writeln(devbot, fmt.Sprint(names))
}

func exitCommand(u *user, _ []string) {
	u.close(u.name + red.Paint(" has left the chat"))
}

func bellCommand(u *user, _ []string) {
	u.bell = !u.bell
	if u.bell {
		u.writeln(devbot, "Turned bell ON")
	} else {
		u.writeln(devbot, "Turned bell OFF")
	}
}
