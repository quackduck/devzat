package room

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"devzat/pkg/colors"
	i "devzat/pkg/interfaces"
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
	server i.Server

	name  string
	users []i.User
	mux   sync.Mutex
	bot   i.Bot

	slackIntegration
}

func (r *Room) PickUsername(possibleName string) error {
	if r.Server().IsProfane(possibleName) {
		return errors.New("name can not be profane")
	}

	return nil
}

func (r *Room) ReplaceSlackEmoji(s string) string {
	return r.Server().ReplaceSlackEmoji(s)
}

func (r *Room) init() error {
	if err := r.slackIntegration.init(); err != nil {
		return fmt.Errorf("could not init slack integration: %v", err)
	}

	return nil
}

func (r *Room) Name() string { return r.name }

func (r *Room) Server() i.Server {
	return r.server
}

func (r *Room) SetServer(server i.Server) {
	r.server = server
}

func (r *Room) Cleanup() {
	if r.Name() != r.Server().MainRoom().Name() && len(r.users) == 0 {
		r.Server().DeleteRoom(r)
	}
}
