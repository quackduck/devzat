package main

import (
	"crypto/sha1"
	"encoding/hex"
	"os"
	"strconv"
	"strings"

	"github.com/acarl005/stripansi"
	"github.com/quackduck/term"
	"github.com/slack-go/slack"
)

var (
	SlackChan  chan string
	SlackAPI   *slack.Client
	SlackRTM   *slack.RTM
	SlackBotID string
)

func getMsgsFromSlack() {
	if Integrations.Slack == nil {
		return
	}

	go SlackRTM.ManageConnection()

	uslack := new(User)
	uslack.isBridge = true
	devnull, _ := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	uslack.term = term.NewTerminal(devnull, "")
	uslack.room = MainRoom
	for msg := range SlackRTM.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			msg := ev.Msg
			text := strings.TrimSpace(msg.Text)
			if msg.SubType != "" {
				break // We're only handling normal messages.
			}
			u, _ := SlackAPI.GetUserInfo(msg.User)
			if u == nil || u.ID == SlackBotID {
				break
			}
			h := sha1.Sum([]byte(u.ID))
			i, _ := strconv.ParseInt(hex.EncodeToString(h[:2]), 16, 0) // two bytes as an int
			name := strings.Fields(u.RealName)[0]
			uslack.Name = Yellow.Paint(Integrations.Slack.Prefix+" ") + (Styles[int(i)%len(Styles)]).apply(name)
			if Integrations.Discord != nil {
				DiscordChan <- Integrations.Slack.Prefix + " " + name + ": " + text // send this discord message to slack
			}
			runCommands(text, uslack)
		case *slack.ConnectedEvent:
			SlackBotID = ev.Info.User.ID
			Log.Println("Connected to Slack with bot ID", SlackBotID, "as", ev.Info.User.Name)
		case *slack.InvalidAuthEvent:
			Log.Println("Invalid Slack authentication")
			return
		}
	}
}

func slackInit() { // called by init() in config.go
	if Integrations.Slack == nil {
		return
	}

	SlackAPI = slack.New(Integrations.Slack.Token)
	SlackRTM = SlackAPI.NewRTM()
	SlackChan = make(chan string, 100)
	go func() {
		for msg := range SlackChan {
			msg = strings.ReplaceAll(stripansi.Strip(msg), `\n`, "\n")
			SlackRTM.SendMessage(SlackRTM.NewOutgoingMessage(msg, Integrations.Slack.ChannelID))
		}
	}()
}
