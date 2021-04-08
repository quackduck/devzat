package main

import (
	"crypto/sha256"
	_ "embed"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	gossh "golang.org/x/crypto/ssh"
	"io"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/acarl005/stripansi"
	"github.com/fatih/color"
	"github.com/gliderlabs/ssh"
	"github.com/slack-go/slack"
	terminal "golang.org/x/term"
)

var (
	//go:embed slackAPI.txt
	slackAPI []byte
	//go:embed adminPass.txt
	adminPass  []byte
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
	colorArr   = []*color.Color{green, cyan, magenta, yellow, white, blue}
	users      = make([]*user, 0, 10)
	allUsers   = make(map[string]string, 100) //map format is u.id => u.name
	port       = 22
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
	name      string
	session   ssh.Session
	term      *terminal.Terminal
	bell      bool
	color     color.Color
	id        string
	closeOnce sync.Once
}

func newUser(s ssh.Session) *user {
	term := terminal.NewTerminal(s, "> ")
	_ = term.SetSize(10000, 10000) // disable any formatting done by term

	hash := sha256.New()
	hash.Write([]byte(strings.TrimSpace(string(gossh.MarshalAuthorizedKey(s.PublicKey())))))
	u := &user{"", s, term, true, color.Color{}, hex.EncodeToString(hash.Sum(nil)), sync.Once{}}
	u.pickUsername(s.User())
	users = append(users, u)
	allUsers[u.id] = u.name
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
	return u
}

func (u *user) repl() {
	for {
		line, err := u.term.ReadLine()
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
			u.writeln("An empty message? Send some content!")
			continue
		}
		if line == "/users" {
			names := make([]string, 0, len(users))
			for _, us := range users {
				names = append(names, us.name)
			}
			broadcast(fmt.Sprint(names), toSlack)
		}
		if line == "/all" {
			names := make([]string, 0, len(allUsers))
			for _, name := range allUsers {
				names = append(names, name)
			}
			broadcast(fmt.Sprint(names), toSlack)
		}
		if line == "easter" {
			broadcast("eggs?", toSlack)
		}
		if line == "/exit" {
			return
		}
		if line == "/bell" {
			u.bell = !u.bell
			broadcast(fmt.Sprint("bell: ", u.bell), toSlack)
		}
		if strings.HasPrefix(line, "/id") {
			victim, ok := findUserByName(strings.TrimSpace(strings.TrimPrefix(line, "/id")))
			if !ok {
				broadcast("User not found", toSlack)
			} else {
				broadcast(victim.id, toSlack)
			}
		}
		if strings.HasPrefix(line, "/nick") {
			u.pickUsername(strings.TrimSpace(strings.TrimPrefix(line, "/nick")))
		}
		if strings.HasPrefix(line, "/kick") {
			victim, ok := findUserByName(strings.TrimSpace(strings.TrimPrefix(line, "/kick")))
			if !ok {
				broadcast("User not found", toSlack)
			} else {
				var pass string
				pass, err = u.term.ReadPassword("Admin password: ")
				if err != nil {
					fmt.Println(u.name, err)
				}
				if strings.TrimSpace(pass) == strings.TrimSpace(string(adminPass)) {
					victim.close(true, u)
				} else {
					broadcast("Incorrect password", toSlack)
				}
			}
		}
		if strings.HasPrefix(line, "/color") {
			colorMsg := "Which color? Choose from green, cyan, blue, red (orange), magenta (purple, pink), yellow (beige), white (cream) and black (gray, grey)."
			switch strings.TrimSpace(strings.TrimPrefix(line, "/color")) {
			case "green":
				u.changeColor(*green)
			case "cyan":
				u.changeColor(*cyan)
			case "blue":
				u.changeColor(*blue)
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
			case "easter":
				u.changeColor(*color.New(color.BgHiMagenta, color.FgHiGreen))
			case "baby":
				u.changeColor(*color.New(color.BgBlue, color.FgHiMagenta))
			case "l33t":
				u.changeColor(*u.color.Add(color.BgHiBlack))
			default:
				broadcast(colorMsg, toSlack)
			}
		}
		if line == "/help" {
			broadcast(`Available commands:
   /users   list users
   /nick    change your name
   /color   change your name color
   /exit    leave the chat
   /hide    hide messages from HC Slack
   /bell    toggle the ascii bell
   /id      get a unique identifier for a user
   /all     get a list of all unique users ever
   /kick    kick a user, requires an admin pass
   /help    show this help message
Made by Ishan Goel with feature ideas from Hack Club members.
Thanks to Caleb Denio for lending me his server!`, toSlack)
		}
	}
}

func (u *user) close(kicked bool, kicker *user) {
	u.closeOnce.Do(func() {
		users = remove(users, u)
		if kicked {
			broadcast(u.name+red.Sprint(" has been kicked by ")+kicker.name, true)
		} else {
			broadcast(u.name+red.Sprint(" has left the chat"), true)
		}
		u.session.Close()
	})
}

func (u *user) writeln(msg string) {
	if !strings.HasPrefix(msg, u.name+": ") { // ignore messages sent by same person
		if u.bell {
			u.term.Write([]byte("\a" + msg + "\n")) // "\a" is beep
		} else {
			u.term.Write([]byte(msg + "\n"))
		}
	}
}

func (u *user) pickUsername(possibleName string) {
	possibleName = strings.TrimSpace(possibleName)
	var err error
	for userDuplicate(possibleName) {
		u.writeln("Pick a different username")
		u.term.SetPrompt("> ")
		possibleName, err = u.term.ReadLine()
		if err != nil {
			fmt.Println(err)
		}
		possibleName = strings.TrimSpace(possibleName)
	}
	u.color = *colorArr[rand.Intn(len(colorArr))]
	possibleName = u.color.Sprint(possibleName)
	u.term.SetPrompt(possibleName + ": ")
	u.name = possibleName
}

func (u *user) changeColor(color color.Color) {
	u.name = color.Sprint(stripansi.Strip(u.name))
	u.color = color
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

func saveAllUsersToLog() {
	f, err := os.Create("allUsers.gob")
	if err != nil {
		fmt.Println(err)
		return
	}
	g := gob.NewEncoder(f)
	g.Encode(allUsers)
	f.Close()
}

func main() {
	rand.Seed(time.Now().Unix())
	f, err := os.Open("allUsers.gob")
	if err != nil {
		fmt.Println(err)
		return
	}
	d := gob.NewDecoder(f)
	d.Decode(&allUsers)
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		saveAllUsersToLog()
		os.Exit(0)
	}()
	ssh.Handle(func(s ssh.Session) {
		u := newUser(s)
		u.repl()
		u.close(false, nil)
	})
	if os.Getenv("PORT") != "" {
		port, err = strconv.Atoi(os.Getenv("PORT"))
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	fmt.Println(fmt.Sprintf("Starting chat server on port %d", port))
	go getMsgsFromSlack()
	err = ssh.ListenAndServe(
		fmt.Sprintf(":%d", port),
		nil,
		ssh.HostKeyFile(os.Getenv("HOME")+"/.ssh/id_rsa"),
		ssh.PublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool { return true }))
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

func findUserByName(name string) (*user, bool) {
	for _, u := range users {
		if stripansi.Strip(u.name) == name {
			return u, true
		}
	}
	return nil, false
}

func remove(s []*user, a *user) []*user {
	var i int
	for i = range s {
		if s[i] == a {
			break // i is now where it is
		}
	}
	if i == 0 {
		return make([]*user, 0)
	}
	return append(s[:i], s[i+1:]...)
}
