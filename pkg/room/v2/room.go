package v2

import (
	"devzat/pkg/colors"
	i "devzat/pkg/interfaces"
	"fmt"
	"strings"
)

const (
	usersInitSize = 10
	roomPrefix    = "#"
)

func New(name string) i.Room {
	r := &Room{}

	if !strings.HasPrefix(name, roomPrefix) {
		name = fmt.Sprintf("%s%s", roomPrefix, name)
	}

	r.users = make([]i.User, 0, usersInitSize)

	r.Formatter = colors.NewFormatter()

	return r
}

type Room struct {
	*colors.Formatter
	name  string
	users []i.User
}

func (r *Room) Name() string { return r.name }

func (r *Room) Server() i.Server {
	//TODO implement me
	panic("implement me")
}

func (r *Room) SetServer(server i.Server) {
	//TODO implement me
	panic("implement me")
}

func (r *Room) Cleanup() {
	//TODO implement me
	panic("implement me")
}
