package server

import (
	"crypto/sha1"
	"devzat/pkg/user"
	"encoding/hex"
	"github.com/acarl005/stripansi"
	"github.com/slack-go/slack"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type slackIntegration struct {
	channel chan string
	api     *slack.Client
	rtm     *slack.RTM
}

func (s *slackIntegration) init() error {
	go s.rtm.ManageConnection()

	return nil
}

func (s *Server) IsOfflineSlack() bool {
	return s.Slack.Offline
}

func (s *Server) GetMsgsFromSlack() {
	uslack := user.SlackUser{}
	uslack.SetRoom(s.MainRoom())

	styles := s.Formatter.Styles.Normal
	yellow := s.Formatter.Colors().Yellow

	for e := range s.slack.rtm.IncomingEvents {
		switch data := e.Data.(type) {
		case *slack.MessageEvent:
			msg := data.Msg
			text := strings.TrimSpace(msg.Text)
			if msg.SubType != "" {
				break // We're only handling normal messages.
			}

			u, _ := s.slack.api.GetUserInfo(msg.User)
			if !strings.HasPrefix(text, "./hide") {
				h := sha1.Sum([]byte(u.ID))
				i, _ := strconv.ParseInt(hex.EncodeToString(h[:2]), 16, 0) // two bytes as an int
				coloredName := yellow.Paint("HC ") + (styles[int(i)%len(styles)]).Apply(strings.Fields(u.RealName)[0])
				uslack.PickUsername(coloredName)
				uslack.IsSlackUser = true
				_ = s.MainRoom().ParseUserInput(text, &uslack)
			}

		case *slack.ConnectedEvent:
			s.Info().Msg("Connected to Slack")
			return
		case *slack.InvalidAuthEvent:
			s.Info().Msg("Invalid token")
			return
		}
	}
}

func (s *Server) GetSendToSlackChan() chan string {
	s.slack.channel = s.GetSendToSlackChan()
	slackChannelID := "C01T5J557AA" // todo: generalize
	slackAPI, err := ioutil.ReadFile("slackAPI.txt")

	if os.IsNotExist(err) {
		s.Slack.Offline = true
		s.Info().Msg("Did not find slackAPI.txt. Enabling offline mode.")
	} else if err != nil {
		s.Fatal().Err(err)
	}

	if s.Slack.Offline {
		msgs := make(chan string, 2)
		go func() {
			for range msgs {
			}
		}()
		return msgs
	}

	s.slack.api = slack.New(string(slackAPI))
	s.slack.rtm = s.slack.api.NewRTM()

	msgs := make(chan string, 100)
	go func() {
		for msg := range msgs {
			msg = strings.ReplaceAll(stripansi.Strip(msg), `\n`, "\n")
			//if strings.HasPrefix(msg, "sshchat: ") { // just in case
			//	continue
			//}
			s.slack.rtm.SendMessage(s.slack.rtm.NewOutgoingMessage(msg, slackChannelID))
		}
	}()

	return msgs
}
