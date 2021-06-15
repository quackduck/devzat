package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/acarl005/stripansi"
)

func registerCommands() {
	commands = append(commands, commandInfo{"clear", "Clears your terminal", clearCommand, false, false})
	commands = append(commands, commandInfo{"msg", "Sends a private message to someone", messageCommand, false, false})
	commands = append(commands, commandInfo{"users", "Gets a list of the active users", usersCommand, true, false})
	commands = append(commands, commandInfo{"all", "Gets a list of all users who has ever connected", allCommand, true, false})
	commands = append(commands, commandInfo{"exit", "Kicks you out of the chat incase your client was bugged", exitCommand, false, false})
	commands = append(commands, commandInfo{"bell", "Toggles notifications when you get pinged", bellCommand, true, false})
	commands = append(commands, commandInfo{"room", "Changes which room you are currently in", roomCommand, false, false})
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

func roomCommand(u *user, args []string) {
	if len(args) == 0 {
		if u.messaging != nil {
			u.writeln(devbot, fmt.Sprintf("You are currently private messaging %s, and in room %s", u.messaging.name, u.room.name))
		} else {
			u.writeln(devbot, fmt.Sprintf("You are currently in %s", u.room.name))
		}
		return
	}
	if args[0] == "leave" {
		if u.messaging == nil {
			if u.room != mainRoom {
				u.changeRoom(mainRoom, true)
				u.writeln(devbot, "You are now in the main room!")
			} else {
				u.writeln(devbot, "You are not messaging someone or in a room") // TODO: This should probably be more clear that they can leave the room that they are in if they are not in the mainroom or if they are messaging someone
			}
			return
		}
		// They are messaging someone
		u.messaging = nil
		u.writeln(devbot, fmt.Sprintf("You are now in %s", u.room.name))

		return
	}

	if strings.HasPrefix(args[0], "#") {
		// It's a normal room

		roomName := strings.TrimPrefix(args[0], "#")
		if len(roomName) == 0 {
			u.writeln(devbot, "You need to give me a channel name to move you to!")
			return
		}
		newRoom, exists := rooms[roomName]
		if !exists {
			newRoom = &room{roomName, make([]*user, 0, 10), sync.Mutex{}}
			rooms[roomName] = newRoom
		}
		u.changeRoom(newRoom, true)
		u.writeln(devbot, fmt.Sprintf("Moved you to %s", roomName))
		return
	}
	if strings.HasPrefix(args[0], "@") {
		userName := strings.TrimPrefix(args[0], "@")
		if len(userName) == 0 {
			u.writeln(devbot, "You have to tell me who you want to message")
			return
		}
		peer, ok := findUserByName(u.room, userName)
		if !ok {
			u.writeln(devbot, "No person in your room found with that name")
			return
		}
		u.messaging = peer
		u.system(fmt.Sprintf("Now messaging %s. To leave use\n>./room leave", u.messaging.name))
		return
	}
	u.system("Invalid usage. Valid usage: ./room leave|#room-name|@user-name")
}
