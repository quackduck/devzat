package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/acarl005/stripansi"

	"github.com/slack-go/slack"
)

var (
	slackChan chan string
	api       *slack.Client
	rtm       *slack.RTM
)

func getMsgsFromSlack() {
	if Integrations.Slack == nil {
		return
	}

	go rtm.ManageConnection()
	uslack := new(user)
	uslack.room = mainRoom
	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			msg := ev.Msg
			text := strings.TrimSpace(msg.Text)
			if msg.SubType != "" {
				break // We're only handling normal messages.
			}
			u, _ := api.GetUserInfo(msg.User)
			if !strings.HasPrefix(text, "./hide") {
				h := sha1.Sum([]byte(u.ID))
				i, _ := strconv.ParseInt(hex.EncodeToString(h[:2]), 16, 0) // two bytes as an int
				uslack.name = yellow.Paint(Integrations.Slack.Prefix+" ") + (styles[int(i)%len(styles)]).apply(strings.Fields(u.RealName)[0])
				uslack.isSlack = true
				runCommands(text, uslack)
			}
		case *slack.ConnectedEvent:
			l.Println("Connected to Slack")
		case *slack.InvalidAuthEvent:
			l.Println("Invalid token")
			return
		}
	}
}

func slackInit() { // called by init() in config.go
	if Integrations.Slack == nil {
		slackChan = make(chan string, 2)
		go func() {
			for range slackChan {
			}
		}()
		return
	}

	fmt.Println(Integrations)
	api = slack.New(Integrations.Slack.Token)
	rtm = api.NewRTM()
	slackChan = make(chan string, 100)
	go func() {
		for msg := range slackChan {
			msg = strings.ReplaceAll(stripansi.Strip(msg), `\n`, "\n")
			rtm.SendMessage(rtm.NewOutgoingMessage(msg, Integrations.Slack.ChannelID))
		}
	}()
}
