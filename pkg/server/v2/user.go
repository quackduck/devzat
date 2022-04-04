package v2

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/gliderlabs/ssh"
	terminal "github.com/quackduck/term"

	"devzat/pkg/colors"
	"devzat/pkg/interfaces"
	"devzat/pkg/models"
	user "devzat/pkg/user/v2"
	"devzat/pkg/util"
)

func (s *Server) NewUser(session ssh.Session) (interfaces.User, error) {
	term := terminal.NewTerminal(session, "> ")

	pty, winChan, accepted := session.Pty()
	if !accepted {
		return nil, fmt.Errorf("PTY for ssh session not accepted: %v")
	}

	// disable any formatting done by term
	_ = term.SetSize(hugeTerminalSize, hugeTerminalSize)

	// definitely should not give an err
	host, _, _ := net.SplitHostPort(session.RemoteAddr().String())

	toHash := host // If we can't get the public key fall back to the IP.
	if pubKey := session.PublicKey(); pubKey != nil {
		toHash = string(pubKey.Marshal())
	}

	u := &user.User{
		name:          "",
		pronouns:      []string{"unset"},
		session:       session,
		term:          term,
		id:            shasum(toHash),
		addr:          host,
		Window:        pty.Window,
		LastTimestamp: time.Now(),
		JoinTime:      time.Now(),
		room:          s.mainRoom,
	}

	u.SetBell(true)
	u.Color.Background = colors.NoBackground // the FG will be set randomly

	go func() {
		for u.Window = range winChan {
		}
	}()

	s.Log.Printf("Connected %v [%v]", u.Name, u.ID)

	if s.BansContains(u.Addr(), u.ID()) {
		banResponse := fmt.Sprintf(fmtDefaultBannedLoginResponse, u.ID)
		botName := s.mainRoom.Bot.Name()

		s.Log.Printf("Rejected %v [%v]", u.Name, host)
		u.Writeln(botName, banResponse)

		u.CloseQuietly()

		return nil, nil
	}

	s.idsInMinToTimes[u.ID()]++
	time.AfterFunc(60*time.Second, func() {
		s.idsInMinToTimes[u.ID()]--
	})

	if s.idsInMinToTimes[u.ID()] > tooManyLogins {
		s.Bans = append(s.Bans, models.Ban{u.Addr(), u.ID()})
		msg := fmt.Sprintf("`%v` has been banned automatically. ID: %v", u.Name, u.ID)

		s.mainRoom.Broadcast(s.mainRoom.Bot.Name(), msg)

		return nil, nil
	}

	// always clear the screen on connect
	if err := s.commands["Clear"]("", u); err != nil {
		return nil, err
	}

	s.handleValentinesDay(u)

	if len(s.Backlog) > 0 {
		lastStamp := s.Backlog[0].Time
		u.RWriteln(util.PrintPrettyDuration(u.JoinTime.Sub(lastStamp)) + " earlier")

		for i := range s.Backlog {
			if s.Backlog[i].Time.Sub(lastStamp) > time.Minute {
				lastStamp = s.Backlog[i].Time
				u.RWriteln(util.PrintPrettyDuration(u.JoinTime.Sub(lastStamp)) + " earlier")
			}
			u.Writeln(s.Backlog[i].SenderName, s.Backlog[i].Text)
		}
	}

	if err := u.PickUsernameQuietly(session.User()); err != nil { // User exited or had some error
		s.Log.Println(err)
		return nil, session.Close()
	}

	s.mainRoom.UsersMutex.Lock()
	s.mainRoom.Users = append(s.mainRoom.Users, u)
	go s.SendCurrentUsersTwitterMessage()
	s.mainRoom.UsersMutex.Unlock()

	u.Term().SetBracketedPasteMode(true) // experimental paste bracketing support
	term.AutoCompleteCallback = func(line string, pos int, key rune) (string, int, bool) {
		return s.autocompleteCallback(u, line, pos, key)
	}

	switch len(s.mainRoom.Users) - 1 {
	case 0:
		u.Writeln("", s.mainRoom.Colors.Blue.Paint("Welcome to the chat. There are no more users"))
	case 1:
		u.Writeln("", s.mainRoom.Colors.Yellow.Paint("Welcome to the chat. There is one more User"))
	default:
		u.Writeln("", s.mainRoom.Colors.Green.Paint("Welcome to the chat. There are", strconv.Itoa(len(s.mainRoom.Users)-1), "more users"))
	}

	botName := s.mainRoom.Bot.Name()
	s.mainRoom.Broadcast(botName, fmt.Sprintf("%s has joined the chat", u.Name))

	return u, nil
}
