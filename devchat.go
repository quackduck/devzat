package main

import (
	_ "embed"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/acarl005/stripansi"
	"github.com/fatih/color"
	"github.com/gliderlabs/ssh"
	"github.com/slack-go/slack"
	terminal "golang.org/x/term"
)

var (
	//go:embed slackAPI.txt
	slackAPI  []byte
	slackChan = getSendToSlackChan()
	api       = slack.New(string(slackAPI))
	rtm       = api.NewRTM()
	red       = color.New(color.FgHiRed)
	green     = color.New(color.FgHiGreen)
	cyan      = color.New(color.FgHiCyan)
	magenta   = color.New(color.FgHiMagenta)
	yellow    = color.New(color.FgHiYellow)
	colorArr  = []*color.Color{red, green, cyan, magenta, yellow}
	//writers    = make([]func(string), 0, 5) // TODO: make a new type called user that has the writer, session and username
	//usernames  = make([]string, 0, 5)
	users      = make([]*user, 0, 10)
	backlog    = make([]string, 0, 10)
	port       = 2222
	scrollback = 16
)

func broadcast(msg string, toSlack bool) {
	backlog = append(backlog, msg+"\n")
	if toSlack {
		slackChan <- msg
	}
	for len(backlog) > scrollback { // for instead of if just in case
		backlog = backlog[1:]
	}
	for i := range users {
		users[i].writeln(msg)
	}
}

type user struct {
	name    string
	session ssh.Session
	term    *terminal.Terminal
}

func (u *user) writeln(msg string) {
	if !strings.HasPrefix(msg, u.name+": ") { // ignore messages sent by same person
		_, _ = u.term.Write([]byte("\a" + msg + "\n")) // "\a" is beep
	}
}
func (u *user) pickNewUsername(possibleName string, color color.Color) {
	var err error
	for userDuplicate(possibleName) {
		u.writeln("Pick a different username")
		possibleName, err = u.term.ReadLine()
		if err != nil {
			fmt.Println(err)
		}
	}
	possibleName = color.Sprint(possibleName)
	u.term.SetPrompt(possibleName + ": ")
	u.name = possibleName
}

// Returns true if the username is taken, false otherwise
func userDuplicate(a string) bool {
	for i := range users {
		if stripansi.Strip(users[i].name) == stripansi.Strip(a) {
			return true
		}
	}
	return false
}

func main() {
	rand.Seed(time.Now().Unix())
	var err error
	ssh.Handle(func(s ssh.Session) {
		myColor := *colorArr[rand.Intn(len(colorArr))]
		username := s.User()

		term := terminal.NewTerminal(s, "> ")
		_ = term.SetSize(10000, 10000) // disable any formatting done by term

		for userDuplicate(username) {
			term.Write([]byte("Pick a different username\n"))
			username, err = term.ReadLine()
			if err != nil {
				fmt.Println(err)
			}
		}
		username = myColor.Sprint(username)
		term.SetPrompt(username + ": ")

		u := &user{username, s, term}
		fmt.Println(*u, u)
		users = append(users, u)
		fmt.Println(users)

		defer func() {
			users = remove(users, u)
			broadcast(u.name+red.Sprint(" has left the chat."), true)
		}()
		switch len(users) - 1 {
		case 0:
			u.writeln(cyan.Sprint("Welcome to the chat. There are no more users"))
		case 1:
			u.writeln(cyan.Sprint("Welcome to the chat. There is one more user"))
		default:
			u.writeln(cyan.Sprint("Welcome to the chat. There are ", len(users)-1, " more users"))
		}
		_, _ = term.Write([]byte(strings.Join(backlog, ""))) // print out backlog

		broadcast(u.name+green.Sprint(" has joined the chat"), true)
		var line string
		for {
			line, err = term.ReadLine()
			line = strings.TrimSpace(line)

			if err == io.EOF {
				return
			}
			if err != nil {
				u.writeln(fmt.Sprint(err))
				fmt.Println(u.name, err)
				continue
			}
			toSlack := true

			if strings.HasPrefix(line, "/hide") {
				toSlack = false
			}
			if !(line == "") {
				broadcast(u.name+": "+line, toSlack)
			}
			if line == "/users" {
				names := make([]string, 0, len(users))
				for _, us := range users {
					names = append(names, us.name)
				}
				broadcast(fmt.Sprint(names), toSlack)
			}
			if line == "easter" {
				broadcast("eggs?", toSlack)
			}
			if line == "/exit" {
				return
			}
			if strings.HasPrefix(line, "/nick") {
				u.pickNewUsername(strings.TrimSpace(strings.TrimPrefix(line, "/nick")), myColor)
			}
			if line == "/help" {
				broadcast(`Available commands:
   /users   list users
   /exit    leave the chat
   /help    show this help message
   /hide    hide messages from HC Slack
Made by Ishan Goel (@quackduck)`, toSlack)
			}
		}
	})

	fmt.Println(fmt.Sprintf("Starting chat server on port %d", port))
	go getMsgsFromSlack()
	err = ssh.ListenAndServe(fmt.Sprintf(":%d", port), nil, ssh.HostKeyFile(os.Getenv("HOME")+"/.ssh/id_rsa"))
	if err != nil {
		fmt.Println(err)
	}
}

func getMsgsFromSlack() {
	go rtm.ManageConnection()
	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			msg := ev.Msg
			if msg.SubType != "" {
				break // We're only handling normal messages.
			}
			u, _ := api.GetUserInfo(msg.User)
			if !strings.HasPrefix(msg.Text, "hide") {
				broadcast("slack: "+u.RealName+": "+msg.Text, false)
			}
		case *slack.ConnectedEvent:
			fmt.Println("Connected to Slack")
		case *slack.InvalidAuthEvent:
			fmt.Println("Invalid token")
			return
		}
	}
}

func getSendToSlackChan() chan string {
	msgs := make(chan string, 100)
	go func() {
		for msg := range msgs {
			if strings.HasPrefix(msg, "slack: ") { // just in case
				continue
			}
			msg = stripansi.Strip(msg)
			rtm.SendMessage(rtm.NewOutgoingMessage(msg, "C01T5J557AA"))
		}
	}()
	return msgs
}

func remove(s []*user, a *user) []*user {
	var i int
	for i = range s {
		if s[i] == a {
			break // i is now where it is
		}
	}
	return append(s[:i], s[i+1:]...)
}
