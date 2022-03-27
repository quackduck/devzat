package pkg

import (
	"devchat/pkg/colors"
	"fmt"
	"github.com/acarl005/stripansi"
	"github.com/slack-go/slack"
	"io/ioutil"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

type Room struct {
	Server *Server
	Name   string
	users  []*User
	*colors.Formatter
	usersMutex   sync.Mutex
	slackChan    chan string
	offlineSlack bool
	Bot          Bot
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
	r.usersMutex.Lock()
	for i := range r.users {
		msg = strings.ReplaceAll(msg, "@"+stripansi.Strip(r.users[i].Name), r.users[i].Name)
		msg = strings.ReplaceAll(msg, `\`+r.users[i].Name, "@"+stripansi.Strip(r.users[i].Name)) // allow escaping
	}
	for i := range r.users {
		r.users[i].Writeln(senderName, msg)
	}
	r.usersMutex.Unlock()
	if r == r.Server.MainRoom {
		r.Server.backlog = append(r.Server.backlog, BacklogMessage{time.Now(), senderName, msg + "\n"})
		if len(r.Server.backlog) > scrollback {
			r.Server.backlog = r.Server.backlog[len(r.Server.backlog)-scrollback:]
		}
	}
}

// cleanupRoom deletes a Room if it's empty and isn't the MainRoom Room
func (r *Room) cleanupRoom() {
	if r != r.Server.MainRoom && len(r.users) == 0 {
		delete(r.Server.Rooms, r.Name)
	}
}

func (r *Room) UserDuplicate(a string) (*User, bool) {
	for i := range r.users {
		if stripansi.Strip(r.users[i].Name) == stripansi.Strip(a) {
			return r.users[i], true
		}
	}
	return nil, false
}

func (r *Room) FindUserByName(name string) (*User, bool) {
	r.usersMutex.Lock()
	defer r.usersMutex.Unlock()
	for _, u := range r.users {
		if stripansi.Strip(u.Name) == name {
			return u, true
		}
	}
	return nil, false
}

func (r *Room) printUsersInRoom() string {
	userNames, adminNames := make([]string, 0), make([]string, 0)

	for _, us := range r.users {
		if isAdmin, _ := r.checkIsAdmin(us); isAdmin {
			adminNames = append(adminNames, us.Name)
			continue
		}

		userNames = append(userNames, us.Name)
	}

	users := formatNames(userNames)
	admins := formatNames(adminNames)
	fromatted := fmt.Sprintf("%s\n\nAdmins: %s", users, admins)

	return fromatted
}

func (r *Room) checkIsAdmin(u *User) (bool, error) {
	adminList, err := r.Server.getAdmins()
	if err != nil {
		return false, err
	}

	_, ok := adminList[u.id]
	return ok, nil
}

func (r *Room) getSendToSlackChan() chan string {
	r.slack.channel = r.getSendToSlackChan()
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
func (r *Room) runCommands(line string, u *User) error {
	if detectBadWords(line) {
		banUser("devbot [grow up]", u)
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
		dmRoomCMD(line, u)
		return nil
	}

	if strings.HasPrefix(line, "=") && !u.isSlack {
		dmCMD(strings.TrimSpace(strings.TrimPrefix(line, "=")), u)
		return nil
	}

	switch currCmd {
	case "hang":
		hangCMD(strings.TrimSpace(strings.TrimPrefix(line, "hang")), u)
		return nil
	case "cd":
		cdCMD(strings.TrimSpace(strings.TrimPrefix(line, "cd")), u)
		return nil
	case "shrug":
		shrugCMD(strings.TrimSpace(strings.TrimPrefix(line, "shrug")), u)
		return nil
	}

	if u.isSlack {
		u.Room.BroadcastNoSlack(u.name, line)
	} else {
		u.Room.Broadcast(u.name, line)
	}

	devbotChat(u.Room, line)

	for _, c := range allcmds {
		if c.name == currCmd {
			c.run(strings.TrimSpace(strings.TrimPrefix(line, c.name)), u)
			return
		}
	}
}
