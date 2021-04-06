package main

import (
	_ "embed"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/acarl005/stripansi"
	"github.com/fatih/color"
	"github.com/gliderlabs/ssh"
	terminal "golang.org/x/term"
)

var (
	//go:embed slackAPIURL.txt
	slackAPIURL []byte
	red         = color.New(color.FgHiRed)
	green       = color.New(color.FgHiGreen)
	cyan        = color.New(color.FgHiCyan)
	magenta     = color.New(color.FgHiMagenta)
	yellow      = color.New(color.FgHiYellow)
	colorArr    = []*color.Color{red, green, cyan, magenta, yellow}
	writers     = make([]func(string), 0, 5)
	usernames   = make([]string, 0, 5)
	backlog     = make([]string, 0, 10)
)

func broadcast(msg string) {
	backlog = append(backlog, msg+"\n")
	sendToSlack(msg)
	for len(backlog) > 16 { // for instead of if just in case
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
	const PORT = 2222
	var err error
	ssh.Handle(func(s ssh.Session) {
		myColor := *colorArr[rand.Intn(len(colorArr))]
		username := myColor.Sprint(s.User())

		term := terminal.NewTerminal(s, "> ")
		_ = term.SetSize(10000, 10000) // disable any formatting done by term
		rand.Seed(time.Now().Unix())

		writeln := func(msg string) {
			if !strings.HasPrefix(msg, username+":") { // ignore messages sent by same person
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
		term.SetPrompt(username + ": ")

		writers = append(writers, writeln)
		usernames = append(usernames, username)

		defer func() {
			usernames = remove(usernames, username)
			broadcast(username + red.Sprint(" has left the chat."))
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

		broadcast(username + green.Sprint(" has joined the chat"))
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
			if !(line == "") {
				broadcast(username + ": " + line)
			}
			if line == "/users" {
				broadcast(fmt.Sprint(usernames))
			}
			if line == "/exit" {
				return
			}
			if line == "/help" {
				broadcast(`Available commands:
   /users   list users
   /exit    leave the chat
   /help    show this help message
Made by Ishan G (@quackduck)`)
			}
		}
	})

	fmt.Println(fmt.Sprintf("Starting chat server on port %d", PORT))
	log.Fatal(ssh.ListenAndServe(fmt.Sprintf(":%d", PORT), nil, ssh.HostKeyFile(os.Getenv("HOME")+"/.ssh/id_rsa")))
}

func sendToSlack(msg string) {
	go func() {
		msg = stripansi.Strip(msg)
		r, err := http.Post(string(slackAPIURL), "application/json", strings.NewReader(fmt.Sprintf(`{"text":"%s"}`, msg)))
		if err != nil {
			fmt.Println(err)
			return
		}
		_ = r.Body.Close()
	}()
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
