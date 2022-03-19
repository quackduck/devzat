package main

import (
	"crypto/sha1"
	"encoding/hex"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/acarl005/stripansi"

	"github.com/slack-go/slack"
)

var (
	slackChan      = getSendToSlackChan()
	slackChannelID = "C01T5J557AA" // todo: generalize
	api            *slack.Client
	rtm            *slack.RTM
)

func getMsgsFromSlack() {
	if offlineSlack {
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
				uslack.Name = yellow.Paint("HC ") + (styles[int(i)%len(styles)]).apply(strings.Fields(u.RealName)[0])
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

func getSendToSlackChan() chan string {
	slackAPI, err := ioutil.ReadFile("slackAPI.txt")

	if os.IsNotExist(err) {
		offlineSlack = true
		l.Println("Did not find slackAPI.txt. Enabling offline mode.")
	} else if err != nil {
		panic(err)
	}

	if offlineSlack {
		msgs := make(chan string, 2)
		go func() {
			for range msgs {
			}
		}()
		return msgs
	}

	api = slack.New(string(slackAPI))
	rtm = api.NewRTM()
	msgs := make(chan string, 100)
	go func() {
		for msg := range msgs {
			msg = strings.ReplaceAll(stripansi.Strip(msg), `\n`, "\n")
			//if strings.HasPrefix(msg, "sshchat: ") { // just in case
			//	continue
			//}
			rtm.SendMessage(rtm.NewOutgoingMessage(msg, slackChannelID))
		}
	}()
	return msgs
}
