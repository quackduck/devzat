package room

import (
	"crypto/sha1"
	"devzat/pkg/user"
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/slack-go/slack"
)

func (r *Room) GetMsgsFromSlack() {
	if r.Server().IsOfflineSlack() {
		return
	}

	go r.slack.rtm.ManageConnection()

	uslack := new(user.User)
	uslack.Room = r.Server().MainRoom()

	styles := r.Formatter.Styles.Normal
	yellow := r.Formatter.Colors.Yellow

	for msg := range r.slack.rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			msg := ev.Msg
			text := strings.TrimSpace(msg.Text)
			if msg.SubType != "" {
				break // We're only handling normal messages.
			}

			u, _ := r.slack.api.GetUserInfo(msg.User)
			if !strings.HasPrefix(text, "./hide") {
				h := sha1.Sum([]byte(u.ID))
				i, _ := strconv.ParseInt(hex.EncodeToString(h[:2]), 16, 0) // two bytes as an int
				uslack.Name = yellow.Paint("HC ") + (styles[int(i)%len(styles)]).Apply(strings.Fields(u.RealName)[0])
				uslack.IsSlack = true
				r.ParseUserInput(text, uslack)
			}

		case *slack.ConnectedEvent:
			r.Server().Log.Println("Connected to Slack")
			return
		case *slack.InvalidAuthEvent:
			r.Server().Log.Println("Invalid token")
			return
		}
	}
}
