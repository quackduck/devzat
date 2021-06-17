package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/acarl005/stripansi"
)

func registerCommands() {
	commands = append(commands, commandInfo{"clear", "Clears your terminal", clearCommand, false, false, nil})
	commands = append(commands, commandInfo{"message", "Sends a private message to someone", messageCommand, false, false, []string{"msg", "="}})
	commands = append(commands, commandInfo{"users", "Gets a list of the active users", usersCommand, true, false, nil})
	commands = append(commands, commandInfo{"all", "Gets a list of all users who has ever connected", allCommand, true, false, nil})
	commands = append(commands, commandInfo{"exit", "Kicks you out of the chat incase your client was bugged", exitCommand, false, false, nil})
	commands = append(commands, commandInfo{"bell", "Toggles notifications when you get pinged", bellCommand, true, false, nil})
	commands = append(commands, commandInfo{"room", "Changes which room you are currently in", roomCommand, false, false, nil})
	commands = append(commands, commandInfo{"kick", "Kicks a user", kickCommand, true, true, nil})
	commands = append(commands, commandInfo{"ban", "Bans a user", banCommand, true, true, nil})
	commands = append(commands, commandInfo{"id", "Gets the hashed IP of the user", idCommand, true, false, nil})
	commands = append(commands, commandInfo{"help", "Get a list of commands", helpCommand, false, false, []string{"commands"}})
	commands = append(commands, commandInfo{"nick", "Change your display name", nickCommand, false, false, nil})
	commands = append(commands, commandInfo{"color", "Change your display name color", colorCommand, false, false, nil})
	commands = append(commands, commandInfo{"timezone", "Change how you view time", timezoneCommand, false, false, []string{"tz"}})
	commands = append(commands, commandInfo{"emojis", "Get a list of emojis you can use", emojisCommand, false, false, nil})
	commands = append(commands, commandInfo{"unban", "Unban a user", unbanCommand, false, true, nil})
}

func clearCommand(u *user, _ []string) {
	u.term.Write([]byte("\033[H\033[2J"))
}

func messageCommand(u *user, args []string) {
	if len(args) < 1 {
		u.system("Please provide a user to send a message to")
		return
	}

	if len(args) < 2 {
		u.system("Please provide a message to send")
	}
	peer, ok := findUserByName(u.room, args[0])
	if !ok {
		u.system("The user was not found, maybe they are in another room?")
		return
	}
	message := strings.Join(append(args[:0], args[1:]...), " ")
	peer.writeln(u.name+" -> ", message)
	u.writeln(u.name+" <- ", message)
}

func usersCommand(u *user, _ []string) {
	u.system(printUsersInRoom(u.room))
}

func allCommand(u *user, _ []string) {
	names := make([]string, 0, len(allUsers))
	for _, name := range allUsers {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		return strings.ToLower(stripansi.Strip(names[i])) < strings.ToLower(stripansi.Strip(names[j]))
	})
	u.system(fmt.Sprint(names))
}

func exitCommand(u *user, _ []string) {
	u.close(u.name + red.Paint(" has left the chat"))
}

func bellCommand(u *user, _ []string) {
	u.bell = !u.bell
	if u.bell {
		u.system("Turned bell ON")
	} else {
		u.system("Turned bell OFF")
	}
}

func roomCommand(u *user, args []string) {
	if len(args) == 0 {
		if u.messaging != nil {
			u.system(fmt.Sprintf("You are currently private messaging %s, and in room %s", u.messaging.name, u.room.name))
		} else {
			u.system(fmt.Sprintf("You are currently in %s", u.room.name))
		}

		// Send a list of rooms and the rooms users.
		type kv struct {
			roomName   string
			numOfUsers int
		}
		var ss []kv
		for k, v := range rooms {
			ss = append(ss, kv{k, len(v.users)})
		}
		sort.Slice(ss, func(i, j int) bool {
			return ss[i].numOfUsers > ss[j].numOfUsers
		})

		u.system("Rooms and users")
		for _, kv := range ss {
			u.system(blue.Paint(kv.roomName) + ": " + printUsersInRoom(rooms[kv.roomName]))
		}

		return
	}
	if args[0] == "leave" {
		if u.messaging == nil {
			if u.room != mainRoom {
				u.changeRoom(mainRoom, true)
				u.system("You are now in the main room!")
			} else {
				u.system("You are not messaging someone or in a room") // TODO: This should probably be more clear that they can leave the room that they are in if they are not in the mainroom or if they are messaging someone
			}
			return
		}
		// They are messaging someone
		u.messaging = nil
		u.system(fmt.Sprintf("You are now in %s", u.room.name))

		return
	}

	if strings.HasPrefix(args[0], "#") {
		// It's a normal room

		roomName := strings.TrimPrefix(args[0], "#")
		if len(roomName) == 0 {
			u.system("You need to give me a channel name to move you to!")
			return
		}
		newRoom, exists := rooms[roomName]
		if !exists {
			newRoom = &room{roomName, make([]*user, 0, 10), sync.Mutex{}}
			rooms[roomName] = newRoom
		}
		u.changeRoom(newRoom, true)
		u.system(fmt.Sprintf("Moved you to %s", roomName))
		return
	}
	if strings.HasPrefix(args[0], "@") {
		userName := strings.TrimPrefix(args[0], "@")
		if len(userName) == 0 {
			u.system("You have to tell me who you want to message")
			return
		}
		peer, ok := findUserByName(u.room, userName)
		if !ok {
			u.system("No person in your room found with that name")
			return
		}
		u.messaging = peer
		u.system(fmt.Sprintf("Now messaging %s. To leave use\n>./room leave", u.messaging.name))
		return
	}
	u.system("Invalid usage. Valid usage: ./room leave|#room-name|@user-name")
}

func kickCommand(u *user, args []string) {
	if len(args) != 1 {
		u.system("Please provide a user to kick!")
		return
	}
	target, ok := findUserByName(u.room, args[0])
	if !ok {
		u.system("User not found!")
		return
	}
	target.system(fmt.Sprintf("You have been kicked by %s", u.name))
	target.close(fmt.Sprintf(red.Paint("%s was kicked by %s"), target.name, u.name))
	u.system("Kicked!")
}

func banCommand(u *user, args []string) {
	if len(args) == 0 {
		u.system("Please provide a user to ban!")
		return
	}
	target, ok := findUserByName(u.room, args[0])
	if !ok {
		u.system("User not found!")
		return
	}
	if len(args) > 1 {
		target.ban(u, strings.Join(args[1:], " "))
	} else {
		target.ban(u, "")
	}
	u.system("Banned!")
}

func idCommand(u *user, args []string) {
	if len(args) == 0 {
		u.system(u.id)
		return
	}

	target, ok := findUserByName(u.room, args[0])
	if !ok {
		u.system("User not found!")
		return
	}
	u.system(target.id)
}

func helpCommand(u *user, args []string) {
	u.system("**Commands**")
	for _, command := range commands {
		if command.requiresAdmin {
			if auth(u) {
				u.system(fmt.Sprintf("%s - %s %s", green.Paint(command.name), command.description, red.Paint("(ADMIN ONLY)")))
			}
		} else {
			u.system(fmt.Sprintf("%s - %s", green.Paint(command.name), command.description))
		}
	}
}

func nickCommand(u *user, args []string) {
	if len(args) > 0 {
		u.pickUsername(strings.Join(args, " "))
	} else {
		u.pickUsername("")
	}
	u.system(fmt.Sprintf("Nick changed to %s", u.name))
}

func colorCommand(u *user, args []string) {
	if len(args) == 0 {
		u.system("Syntax: ./color <which>|#HEX|0-5,0-5,0-5")
		return
	}
	if args[0] == "which" {
		u.system(fmt.Sprintf("Your nickname color is %s", u.color))
		return
	}
	err := u.changeColor(strings.Join(args, " "))
	if err != nil {
		u.system(err.Error())
		return
	}
	u.system("Your display name color has been changed.")
}

func timezoneCommand(u *user, args []string) {
	if len(args) == 0 {
		if u.timezone == nil {
			u.system("You need to send a timezone!")
			return
		}
		u.timezone = nil
		u.system("Cleared your timezone! You will now see relative timestamps (x minutes in)")
		return
	}

	tz, err := time.LoadLocation(args[0])
	if err != nil {
		u.system("Weird timezone you have there, use Continent/City, EST, PST or see nodatime.org/TimeZones!")
		return
	}
	u.timezone = tz
	u.system("Timezone updated!")
}

func emojisCommand(u *user, _ []string) {
	u.system("Check out github.com/ikatyang/emoji-cheat-sheet")
}

func unbanCommand(u *user, args []string) {
	if len(args) == 0 {
		u.system("Who should I unban?")
		return
	}

	for banId, ban := range bans {
		if ban == args[0] {
			bans = append(bans[:banId], bans[banId+1:]...)
			saveBansAndUsers()
			u.system("User has been unbanned")
			return
		}
	}
	u.system("No user found!")

}
