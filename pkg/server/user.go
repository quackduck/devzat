package server

import (
	"devzat/pkg/commands/clear"
	"devzat/pkg/user"
	"fmt"
	"github.com/acarl005/stripansi"
	"strconv"
	"strings"
	"time"

	"devzat/pkg/colors"
	i "devzat/pkg/interfaces"
	"devzat/pkg/models"
	"devzat/pkg/util"
	"github.com/gliderlabs/ssh"
)

func (s *Server) NewUser(session ssh.Session) (i.User, error) {
	u, errUserFromSession := user.FromSession(session)
	if errUserFromSession != nil {
		return nil, fmt.Errorf("could not create user from ssh session: %v", errUserFromSession)
	}

	if err := u.SetBackgroundColor(colors.NoBackground); err != nil {
		// the FG will be set randomly
		return nil, fmt.Errorf("could not set user background color: %v", err)
	}

	s.Info().Msgf("Connected %v [%v]", u.Name(), u.ID())

	if s.BansContains(u.Addr(), u.ID()) {
		s.Info().Msgf("banned user '%s'@'%s' tried to connect", u.ID(), u.Addr())

		banResponse := fmt.Sprintf(fmtDefaultBannedLoginResponse, u.ID())
		botName := s.mainRoom.Bot().Name()

		s.Info().Msgf("Rejected %v [%v]", u.Name(), u.Addr())
		u.Writeln(botName, banResponse)

		u.CloseQuietly()

		return nil, nil
	}

	s.userLoginSpamCounts[u.ID()]++
	time.AfterFunc(60*time.Second, func() {
		s.userLoginSpamCounts[u.ID()]--
	})

	if s.userLoginSpamCounts[u.ID()] > tooManyLogins {
		s.bans = append(s.bans, models.Ban{Addr: u.Addr(), ID: u.ID()})
		msg := fmt.Sprintf("`%v` has been banned automatically. ID: %v", u.Name(), u.ID())

		u.Room().BotCast(msg)

		return nil, nil
	}

	// always clear the screen on connect
	if c, found := s.commands[clear.Name]; found {
		err := c.Fn("", u)

		if err != nil {
			return nil, err
		}
	}

	s.ChangeRoom(u, s.MainRoom().Name())
	s.handleValentinesDay(u)

	if len(s.backlog) > 0 {
		lastStamp := s.backlog[0].Time
		u.RWriteln(util.PrintPrettyDuration(u.JoinTime().Sub(lastStamp)) + " earlier")

		for i := range s.backlog {
			if s.backlog[i].Time.Sub(lastStamp) > time.Minute {
				lastStamp = s.backlog[i].Time
				u.RWriteln(util.PrintPrettyDuration(u.JoinTime().Sub(lastStamp)) + " earlier")
			}
			u.Writeln(s.backlog[i].SenderName, s.backlog[i].Text)
		}
	}

	if err := u.Room().PickUsername(session.User()); err != nil { // User exited or had some error
		s.Error().Err(err)
		return nil, session.Close()
	}

	u.Term().SetBracketedPasteMode(true) // experimental paste bracketing support
	u.Term().AutoCompleteCallback = func(line string, pos int, key rune) (string, int, bool) {
		return s.autocompleteCallback(u, line, pos, key)
	}

	currentUsers := s.mainRoom.AllUsers()

	switch len(currentUsers) - 1 {
	case 0:
		u.Writeln("", s.Colors().Blue.Paint("Welcome to the chat. There are no more users"))
	case 1:
		u.Writeln("", s.Colors().Yellow.Paint("Welcome to the chat. There is one more User"))
	default:
		u.Writeln("", s.Colors().Green.Paint("Welcome to the chat. There are", strconv.Itoa(len(currentUsers)-1), "more users"))
	}

	s.mainRoom.BotCast(fmt.Sprintf("%s has joined the chat", u.Name()))

	return u, nil
}

func (s *Server) Disconnect(user i.User) {
	user.Disconnect()
}

func (s *Server) Kick(u i.User, reason string) {
	u.Close(reason)
}

func (s *Server) AllUsers() []i.User {
	users := make([]i.User, 0)

	for _, r := range s.rooms {
		users = append(users, r.AllUsers()...)
	}

	return users
}

func (s *Server) SetUserColor(u i.User, colorName string) error {
	style, errGet := s.GetStyle(colorName)
	if errGet != nil {
		return errGet
	}

	if strings.HasPrefix(colorName, "bg-") {
		if err := u.SetBackgroundColor(style.Name); err != nil {
			return err
		}
	} else {
		if err := u.SetForegroundColor(style.Name); err != nil {
			return err
		}
	}

	// TODO: having savebans here is wildly incoherent, but this was noticed during a refactor.
	// it stays until i determine something else to do with it.
	if errGet = s.SaveBans(); errGet != nil {
		return fmt.Errorf("could not save the bans file: %v", errGet)
	}

	return nil
}

func (s *Server) NewUserFromSSH(session ssh.Session) (i.User, error) {
	return s.NewUser(session)
}

func (s *Server) UserDuplicate(a string) (i.User, bool) {
	for _, u := range s.AllUsers() {
		name := u.Name()
		if stripansi.Strip(name) == stripansi.Strip(a) {
			return u, true
		}
	}

	return nil, false
}

func (s *Server) FindUserByName(name string) (user i.User, found bool) {
	for _, u := range s.AllUsers() {
		if u.Name() == name {
			return u, true
		}
	}

	return nil, false
}

func (s *Server) PrintUsers() string {
	const fmtAdmin = "@%s"
	names := make([]string, 0)

	for _, u := range s.AllUsers() {
		name := u.Name()
		if s.IsAdmin(u) {
			name = fmt.Sprintf(fmtAdmin, name)
		}

		names = append(names, name)
	}

	return fmt.Sprintf("[ %s ]", strings.Join(names, " "))
}
