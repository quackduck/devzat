package cd

import (
	"sort"
	"strings"
	"sync"

	"devzat/pkg/room"
	"devzat/pkg/user"
)

const (
	name     = ""
	argsInfo = ""
	info     = ""
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

func (c *Command) Fn(rest string, u *user.User) error {
	devbot := u.Room.Bot.Name()
	if u.Messaging != nil {
		u.Messaging = nil
		u.Writeln(devbot, "Left private chat")
		if rest == "" || rest == ".." {
			return
		}
	}

	if rest == ".." { // cd back into the main room
		if u.Room != u.Room.Server.MainRoom {
			u.ChangeRoom(u.Room.Server.MainRoom)
		}
		return
	}

	if strings.HasPrefix(rest, "#") {
		u.Room.Broadcast(u.Name, "cd "+rest)
		if len(rest) > maxLengthRoomName {
			rest = rest[0:maxLengthRoomName]
			u.Room.Broadcast(devbot, "room name lengths are limited, so I'm shortening it to "+rest+".")
		}
		if v, ok := rooms[rest]; ok {
			u.ChangeRoom(v)
		} else {
			rooms[rest] = &room.Room{rest, make([]*user.User, 0, 10), sync.Mutex{}}
			u.ChangeRoom(rooms[rest])
		}
		return
	}

	if rest == "" {
		u.Room.Broadcast(u.Name, "cd "+rest)
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
		roomsInfo := ""
		for _, room := range ss {
			roomsInfo += blue.Paint(room.roomName) + ": " + u.Room.PrintUsersInRoom() + "  \n"
		}
		u.Room.Broadcast("", "Rooms and users  \n"+strings.TrimSpace(roomsInfo))
		return
	}
	name := strings.Fields(rest)[0]
	if len(name) == 0 {
		u.Writeln(devbot, "You think people have empty names?")
		return
	}
	peer, ok := u.Room.FindUserByName(name)
	if !ok {
		u.Writeln(devbot, "No such person lol, who do you want to dm? (you might be in the wrong room)")
		return
	}
	u.Messaging = peer
	u.Writeln(devbot, "Now in DMs with "+peer.Name+". To leave use cd ..")
}
