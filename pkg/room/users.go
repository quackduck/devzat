package room

import (
	i "devzat/pkg/interfaces"
	"fmt"
	"runtime/debug"
	"strings"
)

const (
	fmtRecover = "Slap the developers in the face for me, the server almost crashed, also tell them this: %v, stack: %v"
)

func (r *Room) ParseUserInput(line string, u i.User) error {
	if r.Server().IsProfane(line) {
		r.Server().BanUser("devbot [grow up]", u)
		return nil
	}

	if line == "" {
		return nil
	}

	defer func() { // crash protection
		if err := recover(); err != nil {
			u.Room().BotCast(fmt.Sprintf(fmtRecover, err, debug.Stack()))
		}
	}()

	currCmd := strings.Fields(line)[0]

	if u.DMTarget() != nil &&
		currCmd != "=" &&
		currCmd != "cd" &&
		currCmd != "exit" &&
		currCmd != "pwd" {
		// the commands allowed in a private dm room
		if c, found := r.Server().GetCommand("roomCMD"); found {
			return c.Fn(line, u)
		}
	}

	if strings.HasPrefix(line, "=") && !u.IsSlack() {
		if c, found := r.Server().GetCommand("DirectMessage"); found {
			return c.Fn(strings.TrimSpace(strings.TrimPrefix(line, "=")), u)
		}
	}

	switch currCmd {
	case "hang":
		if c, found := r.Server().GetCommand("Hang"); found {
			return c.Fn(strings.TrimSpace(strings.TrimPrefix(line, "hang")), u)
		}
	case "cd":
		if c, found := r.Server().GetCommand("CMD"); found {
			return c.Fn(strings.TrimSpace(strings.TrimPrefix(line, "cd")), u)
		}
	case "shrug":
		if c, found := r.Server().GetCommand("Shrug"); found {
			return c.Fn(strings.TrimSpace(strings.TrimPrefix(line, "shrug")), u)
		}
	}

	if u.IsSlack() {
		u.Room().BroadcastNoSlack(u.Name(), line)
	} else {
		u.Room().Broadcast(u.Name(), line)
	}

	r.Bot().RespondTo(line)

	for name, c := range r.Server().Commands() {
		if name == currCmd {
			return c.Fn(strings.TrimSpace(strings.TrimPrefix(line, name)), u)
		}
	}

	return nil
}

func (r *Room) AllUsers() []i.User {
	all := make([]i.User, len(r.users))

	idx := 0
	for _, u := range r.users {
		all[idx] = u
		idx++
	}

	return all
}

func (r *Room) Kick(user i.User, reason string) {
	const (
		fmtKick = "you were kicked for %s"
	)

	user.RWriteln(fmt.Sprintf(fmtKick, reason))

	r.Disconnect(user)
}

func (r *Room) Disconnect(user i.User) {
	if r.Name() == r.Server().MainRoom().Name() {
		user.CloseQuietly()

		return
	}

	user.SetRoom(r.Server().MainRoom())
}

func (r *Room) ChangeRoom(u i.User, roomName string) {
	r.Server().ChangeRoom(u, roomName)
}

func (r *Room) Join(u i.User) {
	r.mux.Lock()
	defer r.mux.Unlock()

	r.users = append(r.users, u)
}

func (r *Room) UserDuplicate(a string) (i.User, bool) {
	//TODO implement me
	panic("implement me")
}

func (r *Room) FindUserByName(name string) (i.User, bool) {
	for _, u := range r.users {
		if u.Name() == name {
			return u, true
		}
	}

	for _, u := range r.users {
		if u.Nick() == name {
			return u, true
		}
	}

	return nil, false
}

func (r *Room) PrintUsers() string {
	const fmtAdmin = "@%s"
	names := make([]string, 0)

	for _, u := range r.AllUsers() {
		name := u.Name()
		if r.IsAdmin(u) {
			name = fmt.Sprintf(fmtAdmin, name)
		}

		names = append(names, name)
	}

	return fmt.Sprintf("[ %s ]", strings.Join(names, " "))
}
