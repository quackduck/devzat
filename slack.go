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
	SlackChan chan string
	API       *slack.Client
	RTM       *slack.RTM
)

func getMsgsFromSlack() {
	if Integrations.Slack == nil {
		return
	}

	go RTM.ManageConnection()
	uslack := new(User)
	uslack.isSlack = true
	devnull, _ := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	uslack.term = term.NewTerminal(devnull, "")
	uslack.room = MainRoom
	for msg := range RTM.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			msg := ev.Msg
			text := strings.TrimSpace(msg.Text)
			if msg.SubType != "" {
				break // We're only handling normal messages.
			}
			u, _ := API.GetUserInfo(msg.User)
			if !strings.HasPrefix(text, "./hide") {
				h := sha1.Sum([]byte(u.ID))
				i, _ := strconv.ParseInt(hex.EncodeToString(h[:2]), 16, 0) // two bytes as an int
				uslack.Name = Yellow.Paint(Integrations.Slack.Prefix+" ") + (Styles[int(i)%len(Styles)]).apply(strings.Fields(u.RealName)[0])
				runCommands(text, uslack)
			}
		case *slack.ConnectedEvent:
			Log.Println("Connected to Slack")
		case *slack.InvalidAuthEvent:
			Log.Println("Invalid token")
			return
		}
	}
}

func slackInit() { // called by init() in config.go
	if Integrations.Slack == nil {
		SlackChan = make(chan string, 2)
		go func() {
			for range SlackChan {
			}
		}()
		return
	}

	API = slack.New(Integrations.Slack.Token)
	RTM = API.NewRTM()
	SlackChan = make(chan string, 100)
	go func() {
		for msg := range SlackChan {
			msg = strings.ReplaceAll(stripansi.Strip(msg), `\n`, "\n")
			RTM.SendMessage(RTM.NewOutgoingMessage(msg, Integrations.Slack.ChannelID))
		}
	}()
}
