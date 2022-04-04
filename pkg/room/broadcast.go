package room

import (
	"fmt"
	"strings"
	"time"

	"github.com/acarl005/stripansi"

	"devzat/pkg/models"
)

func (r *Room) Broadcast(senderName, msg string) {
	if msg == "" {
		return
	}

	r.BroadcastNoSlack(senderName, msg)

	if r.slackChan == nil {
		return
	}

	slackMsg := fmt.Sprintf("[%s] %s", r.name, msg)

	if senderName != "" {
		slackMsg = fmt.Sprintf("[%s] %s: %s", r.name, senderName, msg)
	}

	r.slackChan <- slackMsg
}

func (r *Room) BotCast(msg string) {
	r.Broadcast(r.Bot().Name(), msg)
}

func (r *Room) BroadcastNoSlack(senderName, msg string) {
	if msg == "" {
		return
	}

	msg = strings.ReplaceAll(msg, "@everyone", r.Formatter.Colors().Green.Paint("everyone\a"))

	r.mux.Lock()
	defer r.mux.Unlock()

	users := r.AllUsers()

	for _, u := range users {
		name := u.Name()
		msg = strings.ReplaceAll(msg, "@"+stripansi.Strip(name), name)
		msg = strings.ReplaceAll(msg, `\`+name, "@"+stripansi.Strip(name)) // allow escaping
	}

	for _, u := range users {
		u.Writeln(senderName, msg)
	}

	if r.name != r.Server().MainRoom().Name() {
		return
	}

	backlogMsg := models.BacklogMessage{
		Time:       time.Now(),
		SenderName: senderName,
		Text:       msg + "\n",
	}

	bl := r.Server().Backlogs()
	bl = append(bl, backlogMsg)

	r.Server().SetBacklogs(bl)
}
