package server

import (
	"fmt"
	"io"
	"os"
	"time"

	goaway "github.com/TwiN/go-away"
	"github.com/rs/zerolog"

	"devzat/pkg/bot"
	"devzat/pkg/colors"
	i "devzat/pkg/interfaces"
	"devzat/pkg/models"
)

type adminsMap = map[string]interface{} // we just use the map for lookup

type Server struct {
	goaway.ProfanityDetector
	colors.Formatter

	configDir   string
	startupTime time.Time
	bot         i.Bot
	logFile     io.WriteCloser
	zerolog.Logger

	adminsMap

	commandRegistry
	models.ServerSettings
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

	if err := s.initLogs(); err != nil {
		return fmt.Errorf("could not init server logging: %s", err)
	}

	if err := s.roomManagement.init(s); err != nil {
		return fmt.Errorf("could not init rooms: %s", err)
	}

	s.chatHistory.init()
	s.spamManagement.init()
	s.commandRegistry.init()
	s.Formatter.Init()

	if err := s.ServerSettings.Init(); err != nil {
		return fmt.Errorf("could not init server settings: %s", err)
	}

	if err := s.banManagement.init(s); err != nil {
		return fmt.Errorf("could not init server bans: %s", err)
	}

	b := &bot.DevBot{}
	b.SetRoom(s.mainRoom)
	s.mainRoom.SetBot(b)
	if err := b.Init(); err != nil {
		return fmt.Errorf("there was an error initializing the bot: %v", err)
	}

	s.SetBot(b)
	s.bot.SetRoom(s.mainRoom)

	if !s.ServerSettings.Twitter.Offline {
		if err := s.twitter.init(); err != nil && os.IsNotExist(err) {
			s.ServerSettings.Twitter.Offline = true
			s.Info().Msg("Did not find twitter-creds.json. Enabling offline mode.")
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
