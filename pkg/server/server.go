package server

import (
	"devzat/pkg/bot"
	"fmt"
	"os"
	"time"

	"devzat/pkg/colors"
	i "devzat/pkg/interfaces"
	goaway "github.com/TwiN/go-away"
)

type adminsMap = map[string]interface{} // we just use the map for lookup

type Server struct {
	goaway.ProfanityDetector
	colors.Formatter

	startupTime time.Time
	bot         i.Bot

	adminsMap

	commandRegistry
	serverSettings
	serverLogs
	chatHistory
	roomManagement
	banManagement
	spamManagement
	twitter twitterIntegration
	slack   slackIntegration
}

func (s *Server) Init() error {
	s.startupTime = time.Now()
	s.adminsMap = make(adminsMap)

	if err := s.serverLogs.init(); err != nil {
		return fmt.Errorf("could not init server logging: %s", err)
	}

	if err := s.roomManagement.init(s); err != nil {
		return fmt.Errorf("could not init rooms: %s", err)
	}

	s.chatHistory.init()
	s.spamManagement.init()
	s.commandRegistry.init()
	s.Formatter.Init()

	if err := s.serverSettings.init(); err != nil {
		return fmt.Errorf("could not init server settings: %s", err)
	}

	if err := s.banManagement.init(s); err != nil {
		return fmt.Errorf("could not init server bans: %s", err)
	}

	b := &bot.DevBot{}
	b.SetRoom(s.mainRoom)
	if err := b.Init(); err != nil {
		return fmt.Errorf("there was an error initializing the bot: %v", err)
	}

	s.SetBot(b)
	s.bot.SetRoom(s.mainRoom)

	if !s.serverSettings.Twitter.Offline {
		if err := s.twitter.init(); err != nil && os.IsNotExist(err) {
			s.serverSettings.Twitter.Offline = true
			s.Log().Println("Did not find twitter-creds.json. Enabling offline mode.")
		} else if err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) UniverseBroadcast(senderName, msg string) {
	for _, r := range s.rooms {
		r.Broadcast(senderName, msg)
	}
}

func (s *Server) Bot() i.Bot {
	return s.bot
}

func (s *Server) SetBot(b i.Bot) {
	s.bot = b
}

func (s *Server) BotCast(msg string) {
	for _, r := range s.rooms {
		r.BotCast(msg)
	}
}
