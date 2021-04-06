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
	colorArr   = []*color.Color{red, green, cyan, magenta, yellow}
	writers    = make([]func(string), 0, 5)
	usernames  = make([]string, 0, 5)
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
	for i := range writers {
		writers[i](msg)
	}
}

// Returns true if the username is taken, false otherwise
func userDuplicate(a string) bool {
	for _, u := range usernames {
		if stripansi.Strip(u) == stripansi.Strip(a) {
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

		writeln := func(msg string) {
			if !strings.HasPrefix(msg, username+": ") { // ignore messages sent by same person
				_, _ = term.Write([]byte("\a" + msg + "\n")) // "\a" is beep
			}
		}
		for userDuplicate(username) {
			writeln("Pick a different username")
			username, err = term.ReadLine()
			if err != nil {
				writeln(fmt.Sprint(err))
				fmt.Println(username, err)
			}
		}
		username = myColor.Sprint(username)
		term.SetPrompt(username + ": ")

		writers = append(writers, writeln)
		usernames = append(usernames, username)

		defer func() {
			usernames = remove(usernames, username)
			broadcast(username+red.Sprint(" has left the chat."), true)
			//sendToSlack(username + red.Sprint(" has left the chat."))
		}()
		switch len(usernames) - 1 {
		case 0:
			writeln(cyan.Sprint("Welcome to the chat. There are no more users"))
		case 1:
			writeln(cyan.Sprint("Welcome to the chat. There is one more user"))
		default:
			writeln(cyan.Sprint("Welcome to the chat. There are ", len(usernames)-1, " more users"))
		}
		_, _ = term.Write([]byte(strings.Join(backlog, ""))) // print out backlog

		broadcast(username+green.Sprint(" has joined the chat"), true)
		//sendToSlack(username + green.Sprint(" has joined the chat"))
		var line string
		for {
			line, err = term.ReadLine()
			line = strings.TrimSpace(line)

			if err == io.EOF {
				return
			}
			if err != nil {
				writeln(fmt.Sprint(err))
				fmt.Println(username, err)
				continue
			}
			toSlack := true

			if strings.HasPrefix(line, "/hide") {
				toSlack = false
			}
			if !(line == "") {
				broadcast(username+": "+line, toSlack)
			}
			if line == "/users" {
				broadcast(fmt.Sprint(usernames), toSlack)
			}
			if line == "easter" {
				broadcast("eggs?", toSlack)
			}
			if line == "/exit" {
				return
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

func remove(s []string, a string) []string {
	var i int
	for i = range s {
		if s[i] == a {
			break // i is now where it is
		}
	}
	return append(s[:i], s[i+1:]...)
}
