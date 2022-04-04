package room

import (
	"devzat/pkg/interfaces"
	"devzat/pkg/models"
	server2 "devzat/pkg/server"
	"devzat/pkg/user"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/acarl005/stripansi"
	"github.com/slack-go/slack"

	"devzat/pkg/colors"
)

const (
	fmtRecover = "Slap the developers in the face for me, the server almost crashed, also tell them this: %v, stack: %v"
)

type Room struct {
	server interfaces.Server
	name   string
	Users  []*user.User
	*colors.Formatter
	UsersMutex   sync.Mutex
	slackChan    chan string
	offlineSlack bool
	bot          interfaces.Bot
	slack        struct {
		api     *slack.Client
		rtm     *slack.RTM
		channel chan string
	}
}

func (r *Room) GetAdmins() (map[string]string, error) {
	//TODO implement me
	panic("implement me")
}

func (r *Room) IsAdmin(user interfaces.User) (bool, error) {
	return r.Server().IsAdmin(user)
}

func (r *Room) Server() interfaces.Server {
	return r.server
}

func (r *Room) SetServer(s interfaces.Server) {
	r.server = s
}

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
	r.Broadcast(r.bot.Name(), msg)
}

func (r *Room) BroadcastNoSlack(senderName, msg string) {
	if msg == "" {
		return
	}

	msg = strings.ReplaceAll(msg, "@everyone", r.Formatter.Colors.Green.Paint("everyone\a"))

	r.UsersMutex.Lock()

	for i := range r.Users {
		name := r.Users[i].Name()
		msg = strings.ReplaceAll(msg, "@"+stripansi.Strip(name), name)
		msg = strings.ReplaceAll(msg, `\`+name, "@"+stripansi.Strip(name)) // allow escaping
	}

	for i := range r.Users {
		r.Users[i].Writeln(senderName, msg)
	}

	r.UsersMutex.Unlock()

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
	if len(bl) > server2.Scrollback {
		bl = bl[len(bl)-server2.Scrollback:]
	}

	r.Server().SetBacklogs(bl)
}

// Cleanup deletes a Room if it's empty and isn't the mainRoom Room
func (r *Room) Cleanup() {
	if r != r.Server().MainRoom() && len(r.Users) == 0 {
		r.Server().DeleteRoom(r.name)
	}
}

func (r *Room) UserDuplicate(a string) (interfaces.User, bool) {
	for i := range r.Users {
		name := r.Users[i].Name()
		if stripansi.Strip(name) == stripansi.Strip(a) {
			return r.Users[i], true
		}
	}

	return nil, false
}

func (r *Room) FindUserByName(name string) (interfaces.User, bool) {
	r.UsersMutex.Lock()
	defer r.UsersMutex.Unlock()
	for _, u := range r.Users {
		if stripansi.Strip(u.Name()) == name {
			return u, true
		}
	}
	return nil, false
}

func (r *Room) PrintUsersInRoom() string {
	userNames, adminNames := make([]string, 0), make([]string, 0)

	for _, us := range r.Users {
		if isAdmin, _ := r.CheckIsAdmin(us); isAdmin {
			adminNames = append(adminNames, us.Name())
			continue
		}

		userNames = append(userNames, us.Name())
	}

	users := formatName(userNames)
	admins := formatName(adminNames)
	fromatted := fmt.Sprintf("%s\n\nAdmins: %s", users, admins)

	return fromatted
}

func (r *Room) CheckIsAdmin(u interfaces.User) (bool, error) {
	adminList, err := r.Server().GetAdmins()
	if err != nil {
		return false, err
	}

	_, ok := adminList[u.ID()]
	return ok, nil
}

func (r *Room) GetSendToSlackChan() chan string {
	r.slack.channel = r.GetSendToSlackChan()
	slackChannelID := "C01T5J557AA" // todo: generalize
	slackAPI, err := ioutil.ReadFile("slackAPI.txt")

	if os.IsNotExist(err) {
		r.offlineSlack = true
		r.Server().Log.Println("Did not find slackAPI.txt. Enabling offline mode.")
	} else if err != nil {
		panic(err)
	}

	if r.offlineSlack {
		msgs := make(chan string, 2)
		go func() {
			for range msgs {
			}
		}()
		return msgs
	}

	r.slack.api = slack.New(string(slackAPI))
	r.slack.rtm = r.slack.api.NewRTM()

	msgs := make(chan string, 100)
	go func() {
		for msg := range msgs {
			msg = strings.ReplaceAll(stripansi.Strip(msg), `\n`, "\n")
			//if strings.HasPrefix(msg, "sshchat: ") { // just in case
			//	continue
			//}
			r.slack.rtm.SendMessage(r.slack.rtm.NewOutgoingMessage(msg, slackChannelID))
		}
	}()

	return msgs
}

// ParseUserInput parses a line of raw input from a User and sends a message as
// required, running any commands the User may have called.
// It also accepts a boolean indicating if the line of input is from slack, in
// which case some commands will not be run (such as ./tz and ./exit)
func (r *Room) ParseUserInput(line string, u interfaces.User) error {
	return nil
}

func formatName(names []string) string {
	joined := strings.Join(names, " ")

	return fmt.Sprintf("[%s]", joined)
}

func (r *Room) AllUsers() []interfaces.User {
	res := make([]interfaces.User, len(r.Users))

	for idx := range r.Users {
		res[idx] = r.Users[idx]
	}

	return res
}

func (r *Room) Bot() interfaces.Bot {
	return r.bot
}
