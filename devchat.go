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
	slackAPI   []byte
	slackChan  = getSendToSlackChan()
	api        = slack.New(string(slackAPI))
	rtm        = api.NewRTM()
	red        = color.New(color.FgHiRed)
	green      = color.New(color.FgHiGreen)
	cyan       = color.New(color.FgHiCyan)
	magenta    = color.New(color.FgHiMagenta)
	yellow     = color.New(color.FgHiYellow)
	blue       = color.New(color.FgHiBlue)
	black      = color.New(color.FgHiBlack)
	white      = color.New(color.FgHiWhite)
	colorArr   = []*color.Color{green, cyan, magenta, yellow, white}
	users      = make([]*user, 0, 10)
	port       = 2222
	scrollback = 16
	backlog    = make([]string, 0, scrollback)
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
func (u *user) pickNewUsername(possibleName string) {
	var err error
	for userDuplicate(possibleName) {
		u.writeln("Pick a different username")
		u.term.SetPrompt(">")
		possibleName, err = u.term.ReadLine()
		if err != nil {
			fmt.Println(err)
		}
	}
	possibleName = (*colorArr[rand.Intn(len(colorArr))]).Sprint(possibleName)
	u.term.SetPrompt(possibleName + ": ")
	u.name = possibleName
}

func (u *user) changeColor(color color.Color) {
	u.name = color.Sprint(stripansi.Strip(u.name))
	u.term.SetPrompt(u.name + ": ")
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
		users = append(users, u)

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
			} else {
				continue
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
				u.pickNewUsername(strings.TrimSpace(strings.TrimPrefix(line, "/nick")))
			}
			if strings.HasPrefix(line, "/color") {
				parsed := strings.Fields(line)
				colorMsg := "Which color? Choose from green, cyan, blue, red (orange), magenta (purple, pink), yellow (beige), white (cream) and black (gray, grey)."
				if len(parsed) < 2 {
					u.writeln(colorMsg)
				} else {
					switch parsed[1] {
					case "green":
						u.changeColor(*green)
					case "cyan", "blue":
						u.changeColor(*cyan)
					case "red", "orange":
						u.changeColor(*red)
					case "magenta", "purple", "pink":
						u.changeColor(*magenta)
					case "yellow", "beige":
						u.changeColor(*yellow)
					case "white", "cream":
						u.changeColor(*white)
					case "black", "gray", "grey":
						u.changeColor(*black)
					default:
						u.writeln(colorMsg)
					}
				}
			}
			if line == "/help" {
				broadcast(`Available commands:
   /users   list users
   /nick    change your name
   /color   change your name color
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
