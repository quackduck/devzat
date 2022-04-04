package cd

import (
	"devzat/pkg/interfaces"
	"fmt"
	"sort"
	"strings"
)

const (
	name     = "cd"
	argsInfo = ""
	info     = ""
)

const (
	maxLengthRoomName = 30
)

const (
	parent     = ".."
	roomPrefix = "#"
	noArgs     = ""
)

const (
	fmtUserLeftChat    = "%s has left private chat"
	fmtRoomNameTooLong = "room name lengths are limited, so I'm shortening it to %s."
	fmtJoinPrivateChat = "Now in DMs with %s. To leave use cd .."

	msgUserNotFound   = "No such person, who do you want to dm? (you might be in the wrong room)"
	msgEmptyNameGiven = "You think people have empty names?"
)

const (
	argTargetUserName = iota
)

type Command struct{}

func (c *Command) Name() string {
	return name
}

func (c *Command) ArgsInfo() string {
	return argsInfo
}

func (c *Command) Info() string {
	return info
}

func (c *Command) IsRest() bool {
	return false
}

func (c *Command) IsSecret() bool {
	return false
}

func (c *Command) Fn(strArgs string, u interfaces.User) error {
	argv := strings.Fields(strArgs)

	if u.DMTarget() != nil {
		if earlyReturn := fnExitPrivateChat(u, strArgs); earlyReturn {
			return nil
		}
	}

	if strArgs == parent {
		fnReturnMainRoom(u)
		return nil
	}

	if strings.HasPrefix(strArgs, roomPrefix) {
		fnChangeRoom(u, strArgs)
		return nil
	}

	if strArgs == noArgs {
		fnPrintAllRoomsAndUserCounts(u)
		return nil
	}

	fnStartPrivateChat(u, argv[argTargetUserName])

	return nil
}

func fnExitPrivateChat(u interfaces.User, strArgs string) (returnEarly bool) {
	devbot := u.Room().Bot().Name()

	u.SetDMTarget(nil)
	u.Writeln(devbot, fmt.Sprintf(fmtUserLeftChat, u.Name()))

	return strArgs == "" || strArgs == ".."
}

func fnReturnMainRoom(u interfaces.User) {
	if u.Room() == u.Room().Server().MainRoom() {
		return
	}

	u.ChangeRoom(u.Room().Server().MainRoom())
}

func fnChangeRoom(u interfaces.User, strArgs string) {
	if len(strArgs) > maxLengthRoomName {
		strArgs = strArgs[0:maxLengthRoomName]
		u.Room().BotCast(fmt.Sprintf(fmtRoomNameTooLong, strArgs))
	}

	command := fmt.Sprintf("%s %s", name, strArgs)
	u.Room().Broadcast(u.Name(), command)
}

func fnPrintAllRoomsAndUserCounts(u interfaces.User) {
	rooms := u.Room().Server().AllRooms()
	names := make([]string, len(rooms))

	for roomName := range rooms {
		names = append(names, roomName)
	}

	sortByUserCount := func(i, j int) bool {
		return len(rooms[names[i]].Users) > len(rooms[names[j]].Users)
	}

	sort.Slice(names, sortByUserCount)

	roomsInfo := ""
	blue := u.Room().Formatter.Colors.Blue
	for _, room := range names {
		roomsInfo += blue.Paint(room) + ": " + u.Room().PrintUsersInRoom() + "  \n"
	}

	u.Room().Broadcast(u.Name(), name)
	u.Room().Broadcast("", "Rooms and users  \n"+strings.TrimSpace(roomsInfo))
}

func fnStartPrivateChat(u interfaces.User, target string) {
	devbot := u.Room().Bot().Name()

	if len(target) == 0 {
		u.Writeln(devbot, msgEmptyNameGiven)
		return
	}

	peer, ok := u.Room().FindUserByName(target)
	if !ok {
		u.Writeln(devbot, msgUserNotFound)
		return
	}

	u.SetDMTarget(peer)
	u.Writeln(devbot, fmt.Sprintf(fmtJoinPrivateChat, peer.Name()))
}
