package server

import (
	"fmt"
	"time"

	i "devzat/pkg/interfaces"
)

const (
	fmtWarning = "%s, stop spamming or you could get banned."
)

type spamManagement struct {
	userChatSpamCounts  map[string]int
	userLoginSpamCounts map[string]int
}

func (sm *spamManagement) init() {
	// TODO: maybe add some IP-based factor to disallow rapid key-gen attempts
	sm.userChatSpamCounts = make(map[string]int)
	sm.userLoginSpamCounts = make(map[string]int)
}

func (s *Server) Antispam(u i.User) {
	uid := u.ID()

	s.userLoginSpamCounts[uid]++
	time.AfterFunc(5*time.Second, func() {
		s.userLoginSpamCounts[uid]--
	})

	if s.userLoginSpamCounts[uid] >= s.ServerSettings.Antispam.LimitBan {
		botName := u.Room().Bot().Name()

		if !s.BansContains(u.Addr(), uid) {
			s.BanUserForDuration(botName, u, s.ServerSettings.Antispam.BanDuration)
			_ = s.SaveBans()
		}

		u.Writeln(botName, "anti-spam triggered")
		u.Close(s.Colors().Red.Paint(u.Name() + " has been banned for spamming"))

		return
	}

	if s.userLoginSpamCounts[uid] >= s.ServerSettings.Antispam.LimitWarn {
		u.Room().BotCast(fmt.Sprintf(fmtWarning, u.Name()))
	}
}
