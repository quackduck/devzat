package room

import (
	"crypto/sha1"
	"devzat/pkg/user"
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/slack-go/slack"
)

type slackIntegration struct {
	slackChan chan string
	slack     struct {
		api     *slack.Client
		rtm     *slack.RTM
		channel chan string
	}
}

func (s *slackIntegration) init() error {
	go s.slack.rtm.ManageConnection()

	return nil
}

func (r *Room) IsOfflineSlack() bool {
	return r.Server().IsOfflineSlack()
}

func (r *Room) GetSendToSlackChan() chan string {
	return r.Server().GetSendToSlackChan()
}

func (r *Room) GetMsgsFromSlack() {
	uslack := user.SlackUser{}
	uslack.SetRoom(r.Server().MainRoom())

	styles := r.Formatter.Styles.Normal
	yellow := r.Formatter.Colors().Yellow

	for e := range r.slack.rtm.IncomingEvents {
		switch data := e.Data.(type) {
		case *slack.MessageEvent:
			msg := data.Msg
			text := strings.TrimSpace(msg.Text)
			if msg.SubType != "" {
				break // We're only handling normal messages.
			}

			u, _ := r.slack.api.GetUserInfo(msg.User)
			if !strings.HasPrefix(text, "./hide") {
				h := sha1.Sum([]byte(u.ID))
				i, _ := strconv.ParseInt(hex.EncodeToString(h[:2]), 16, 0) // two bytes as an int
				coloredName := yellow.Paint("HC ") + (styles[int(i)%len(styles)]).Apply(strings.Fields(u.RealName)[0])
				_ = uslack.PickUsername(coloredName)
				uslack.IsSlackUser = true
				_ = r.ParseUserInput(text, &uslack)
			}

		case *slack.ConnectedEvent:
			r.Server().Log().Info().Msg("Connected to Slack")
			return
		case *slack.InvalidAuthEvent:
			r.Server().Log().Info().Msg("Invalid token")
			return
		}
	}
}
