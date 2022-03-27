package room

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"devzat/pkg"
	"devzat/pkg/colors"
	"devzat/pkg/server"
	"devzat/pkg/user"
	"github.com/acarl005/stripansi"
	"github.com/slack-go/slack"
)

const (
	maxLengthRoomName = 30
)

const (
	fmtRecover = "Slap the developers in the face for me, the server almost crashed, also tell them this: %v, stack: %v"
)

type Room struct {
	*server.Server
	Name  string
	Users []*user.User
	*colors.Formatter
	UsersMutex   sync.Mutex
	slackChan    chan string
	offlineSlack bool
	Bot          pkg.Bot
	slack        struct {
		api     *slack.Client
		rtm     *slack.RTM
		channel chan string
	}
}

func (r *Room) Broadcast(senderName, msg string) {
	if msg == "" {
		return
	}

	r.BroadcastNoSlack(senderName, msg)

	if r.slackChan == nil {
		return
	}

	slackMsg := fmt.Sprintf("[%s] %s", r.Name, msg)

	if senderName != "" {
		slackMsg = fmt.Sprintf("[%s] %s: %s", r.Name, senderName, msg)
	}

	r.slackChan <- slackMsg
}

func (r *Room) BroadcastNoSlack(senderName, msg string) {
	if msg == "" {
		return
	}

	msg = strings.ReplaceAll(msg, "@everyone", r.Formatter.Colors.Green.Paint("everyone\a"))

	r.UsersMutex.Lock()

	for i := range r.Users {
		msg = strings.ReplaceAll(msg, "@"+stripansi.Strip(r.Users[i].Name), r.Users[i].Name)
		msg = strings.ReplaceAll(msg, `\`+r.Users[i].Name, "@"+stripansi.Strip(r.Users[i].Name)) // allow escaping
	}

	for i := range r.Users {
		r.Users[i].Writeln(senderName, msg)
	}

	r.UsersMutex.Unlock()

	if r != r.Server.MainRoom {
		return
	}

	backlogMsg := pkg.BacklogMessage{
		Time:       time.Now(),
		SenderName: senderName,
		Text:       msg + "\n",
	}

	r.Server.Backlog = append(r.Server.Backlog, backlogMsg)
	if len(r.Server.Backlog) > server.Scrollback {
		r.Server.Backlog = r.Server.Backlog[len(r.Server.Backlog)-server.Scrollback:]
	}
}

// Cleanup deletes a Room if it's empty and isn't the MainRoom Room
func (r *Room) Cleanup() {
	if r != r.Server.MainRoom && len(r.Users) == 0 {
		delete(r.Server.Rooms, r.Name)
	}
}

func (r *Room) UserDuplicate(a string) (*user.User, bool) {
	for i := range r.Users {
		if stripansi.Strip(r.Users[i].Name) == stripansi.Strip(a) {
			return r.Users[i], true
		}
	}
	return nil, false
}

func (r *Room) FindUserByName(name string) (*user.User, bool) {
	r.UsersMutex.Lock()
	defer r.UsersMutex.Unlock()
	for _, u := range r.Users {
		if stripansi.Strip(u.Name) == name {
			return u, true
		}
	}
	return nil, false
}

func (r *Room) PrintUsersInRoom() string {
	userNames, adminNames := make([]string, 0), make([]string, 0)

	for _, us := range r.Users {
		if isAdmin, _ := r.CheckIsAdmin(us); isAdmin {
			adminNames = append(adminNames, us.Name)
			continue
		}

		userNames = append(userNames, us.Name)
	}

	users := formatName(userNames)
	admins := formatName(adminNames)
	fromatted := fmt.Sprintf("%s\n\nAdmins: %s", users, admins)

	return fromatted
}

func (r *Room) CheckIsAdmin(u *user.User) (bool, error) {
	adminList, err := r.Server.GetAdmins()
	if err != nil {
		return false, err
	}

	_, ok := adminList[u.ID]
	return ok, nil
}

func (r *Room) GetSendToSlackChan() chan string {
	r.slack.channel = r.GetSendToSlackChan()
	slackChannelID := "C01T5J557AA" // todo: generalize
	slackAPI, err := ioutil.ReadFile("slackAPI.txt")

	if os.IsNotExist(err) {
		r.offlineSlack = true
		r.Server.Log.Println("Did not find slackAPI.txt. Enabling offline mode.")
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

// runCommands parses a line of raw input from a User and sends a message as
// required, running any commands the User may have called.
// It also accepts a boolean indicating if the line of input is from slack, in
// which case some commands will not be run (such as ./tz and ./exit)
func (r *Room) RunCommands(line string, u *user.User) error {
	if r.Server.IsProfane(line) {
		r.BanUser("devbot [grow up]", u)
		return nil
	}

	if line == "" {
		return nil
	}

	defer func() { // crash protection
		if i := recover(); i != nil {
			botName := u.Room.Server.MainRoom.Bot.Name()
			u.Room.Server.MainRoom.Broadcast(botName, fmt.Sprintf(fmtRecover, i, debug.Stack()))
		}
	}()

	currCmd := strings.Fields(line)[0]
	if u.Messaging != nil && currCmd != "=" && currCmd != "cd" && currCmd != "exit" && currCmd != "pwd" { // the commands allowed in a private dm room
		return r.Commands["roomCMD"](line, u)
	}

	if strings.HasPrefix(line, "=") && !u.IsSlack {
		return r.Commands["DirectMessage"](strings.TrimSpace(strings.TrimPrefix(line, "=")), u)
	}

	switch currCmd {
	case "hang":
		return r.Commands["Hang"](strings.TrimSpace(strings.TrimPrefix(line, "hang")), u)
	case "cd":
		return r.Commands["CMD"](strings.TrimSpace(strings.TrimPrefix(line, "cd")), u)
	case "shrug":
		return r.Commands["Shrug"](strings.TrimSpace(strings.TrimPrefix(line, "shrug")), u)
	}

	if u.IsSlack {
		u.Room.BroadcastNoSlack(u.Name, line)
	} else {
		u.Room.Broadcast(u.Name, line)
	}

	r.Bot.Chat(line)

	for name, c := range r.Commands {
		if name == currCmd {
			return c(strings.TrimSpace(strings.TrimPrefix(line, name)), u)
		}
	}

	return nil
}

func formatName(names []string) string {
	joined := strings.Join(names, " ")

	return fmt.Sprintf("[%s]", joined)
}
