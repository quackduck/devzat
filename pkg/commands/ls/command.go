package ls

import (
	"strings"

	i "devzat/pkg/interfaces"
	"devzat/pkg/models"
)

const (
	name     = "ls"
	argsInfo = "<#room|user>"
	info     = "list all rooms, users, and/or users in a room"
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

func (c *Command) Visibility() models.CommandVisibility {
	return models.CommandVisSecret
}

func (c *Command) Fn(line string, u i.User) error {
	if len(line) > 1 && line[0] == '#' {
		return fnListRoom(line, u)
	}

	if line != "" && !u.IsAdmin() {
		u.Room().Broadcast("", "ls: "+line+" Permission denied")
		return nil
	}

	if err := fnListRooms(u); err != nil {
		return err
	}

	if err := fnListUsers(u.Room().Name(), u); err != nil {
		return err
	}

	return nil
}

func fnListRoom(line string, u i.User) error {
	roomMap := make(map[string]i.Room)
	rooms := u.Room().Server().AllRooms()
	blue := u.Room().Colors().Blue

	for _, r := range rooms {
		roomMap[r.Name()] = r
	}

	r, ok := roomMap[line]
	if !ok {
		return nil
	}

	names := make([]string, 0)
	for _, us := range r.AllUsers() {
		names = append(names, us.Name())
	}

	u.Room().Broadcast("", blue.Paint(strings.Join(names, " ")))

	return nil
}

func fnListRooms(u i.User) error {
	rooms := u.Room().Server().AllRooms()
	blue := u.Room().Colors().Blue

	names := make([]string, 0)
	for _, r := range rooms {
		names = append(names, r.Name())
	}

	u.Room().Broadcast("", blue.Paint(strings.Join(names, " ")))

	return nil
}

func fnListUsers(room string, u i.User) error {
	rooms := u.Room().Server().AllRooms()
	blue := u.Room().Colors().Blue

	names := make([]string, 0)
	for _, r := range rooms {
		if r.Name() == room {
			for _, roomUser := range r.AllUsers() {
				names = append(names, roomUser.Name())
			}
		}
	}

	u.Room().Broadcast("", blue.Paint(strings.Join(names, " ")))

	return nil
}
