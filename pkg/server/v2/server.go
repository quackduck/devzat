package v2

import (
	"devzat/pkg/bot"
	"fmt"
	"strings"
	"time"

	goaway "github.com/TwiN/go-away"
	"github.com/gliderlabs/ssh"

	"devzat/pkg/colors"
	i "devzat/pkg/interfaces"
)

type adminsMap = map[string]interface{} // we just use the map for lookup

type Server struct {
	goaway.ProfanityDetector
	colors.Formatter

	commands    map[string]i.Command
	startupTime time.Time

	adminsMap
	settings
	serverLogs
	chatHistory
	roomManagement
	banManagement
	spamManagement
}

func (s *Server) Init() error {
	s.initCommands()
	s.chatHistory.init()
	s.spamManagement.init()

	if err := s.settings.init(); err != nil {
		return fmt.Errorf("could not init server settings: %s", err)
	}

	if err := s.roomManagement.init(s); err != nil {
		return fmt.Errorf("could not init rooms: %s", err)
	}

	if err := s.serverLogs.init(); err != nil {
		return fmt.Errorf("could not init server logging: %s", err)
	}

	if err := s.banManagement.init(s); err != nil {
		return fmt.Errorf("could not init server bans: %s", err)
	}

	b := &bot.DevBot{}
	b.SetRoom(s.mainRoom)
	if err := b.Init(); err != nil {
		return fmt.Errorf("there was an error initializing the bot: %v", err)
	}

	s.startupTime = time.Now()

	s.loadTwitterClient()

	return nil
}

func (s *Server) UniverseBroadcast(senderName, msg string) {
	for _, r := range s.rooms {
		r.Broadcast(senderName, msg)
	}
}

func (s *Server) Disconnect(user i.User) {
	//TODO implement me
	panic("implement me")
}

func (s *Server) Kick(u i.User, reason string) {
	//TODO implement me
	panic("implement me")
}

func (s *Server) AllUsers() []i.User {
	//TODO implement me
	panic("implement me")
}

func (s *Server) GiveAdmin(user i.User) error {
	if s.adminsMap == nil {
		s.adminsMap = make(adminsMap)
	}

	if _, alreadyAdmin := s.adminsMap[user.ID()]; alreadyAdmin {
		return fmt.Errorf("user '%s' is already an admin", user.Name())
	}

	s.adminsMap[user.ID()] = nil

	return nil
}

func (s *Server) RevokeAdmin(user i.User) error {
	if s.adminsMap == nil {
		s.adminsMap = make(adminsMap)
	}

	if _, isAdmin := s.adminsMap[user.ID()]; !isAdmin {
		return fmt.Errorf("user '%s' is not an admin", user.Name())
	}

	delete(s.adminsMap, user.ID())

	return nil
}

func (s *Server) Antispam(u i.User) {
	const fmtWarning = "%s, stop spamming or you could get banned."

	uid := u.ID()

	s.userLoginSpamCounts[uid]++
	time.AfterFunc(5*time.Second, func() {
		s.userLoginSpamCounts[uid]--
	})

	if s.userLoginSpamCounts[uid] >= s.settings.Antispam.LimitWarn {
		u.Room().BotCast(fmt.Sprintf(fmtWarning, u.Name()))
	}

	if s.userLoginSpamCounts[uid] >= s.settings.Antispam.LimitBan {
		botName := u.Room().Bot().Name()

		if !s.BansContains(u.Addr(), uid) {
			s.BanUserForDuration(botName, u, s.settings.Antispam.BanDuration)
			_ = s.SaveBans()
		}

		u.Writeln(botName, "anti-spam triggered")
		u.Close(s.Colors.Red.Paint(u.Name() + " has been banned for spamming"))

		return
	}
}

func (s *Server) IsAdmin(user i.User) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Server) AddRoom(room i.Room) {
	//TODO implement me
	panic("implement me")
}

func (s *Server) DeleteRoom(room i.Room) {
	//TODO implement me
	panic("implement me")
}

func (s *Server) AddCommand(command i.Command) {
	//TODO implement me
	panic("implement me")
}

func (s *Server) Bot() i.Bot {
	//TODO implement me
	panic("implement me")
}

func (s *Server) BotCast(msg string) {
	//TODO implement me
	panic("implement me")
}

func (s *Server) IsOfflineSlack() bool {
	//TODO implement me
	panic("implement me")
}

func (s *Server) GetSendToSlackChan() chan string {
	//TODO implement me
	panic("implement me")
}

func (s *Server) GetMsgsFromSlack() {
	//TODO implement me
	panic("implement me")
}

func (s *Server) PickUsername(possibleName string) error {
	//TODO implement me
	panic("implement me")
}

func (s *Server) PickUsernameQuietly(possibleName string) error {
	//TODO implement me
	panic("implement me")
}

func (s *Server) GetColorNames() []string {
	//TODO implement me
	panic("implement me")
}

func (s *Server) SetUserColor(u i.User, colorName string) error {
	style, err := s.GetStyle(colorName)
	if err != nil {
		return err
	}

	if strings.HasPrefix(colorName, "bg-") {
		return u.SetBackgroundColor(style.Name)
	} else {
		return u.SetForegroundColor(style.Name)
	}

	// TODO: having savebans here is wildly incoherent, but this was noticed during a refactor.
	// it stays until i determine something else to do with it.
	if err = s.SaveBans(); err != nil {
		return fmt.Errorf("could not save the bans file: %v", err)
	}

	return nil
}

func (s *Server) NewUserFromSSH(session ssh.Session) (i.User, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Server) UserDuplicate(a string) (i.User, bool) {
	//TODO implement me
	panic("implement me")
}

func (s *Server) FindUserByName(name string) (i.User, bool) {
	//TODO implement me
	panic("implement me")
}

func (s *Server) PrintUsersInRoom() string {
	//TODO implement me
	panic("implement me")
}
