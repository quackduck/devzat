package main

import (
	_ "embed"
	"fmt"
	"github.com/fatih/color"
	"github.com/gliderlabs/ssh"
	terminal "golang.org/x/term"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

var (
	red       = color.New(color.FgHiRed)
	green     = color.New(color.FgHiGreen)
	cyan      = color.New(color.FgHiCyan)
	magenta   = color.New(color.FgHiMagenta)
	yellow    = color.New(color.FgHiYellow)
	colorArr  = []*color.Color{red, green, cyan, magenta, yellow}
	writers   = make([]func(string), 0, 5)
	usernames = make([]string, 0, 5)
	backlog   = make([]string, 0, 10)
)

// Closure which returns a user-specific writeln func
func getWriteLnFunc(username string, term *terminal.Terminal) func(msg string) {
	return func(msg string) {
		if !strings.HasPrefix(msg, username+":") { // ignore messages sent by same person
			_, _ = term.Write([]byte("\a" + msg + "\n")) // "\a" is beep
		}
	}
}

func broadcast(msg string) {
	backlog = append(backlog, msg+"\n")
	for len(backlog) > 16 { // for instead of if just in case
		backlog = backlog[1:]
	}
	for i := range writers {
		writers[i](msg)
	}
}

// Returns true if the username is taken, false otherwise
func isUsernameDuplicate(a string) bool {
	for _, u := range usernames {
		if u == a {
			return true
		}
	}
	return false
}

// Handles a duplicate username
func checkForDuplicateName(username string, writeln func(string), term *terminal.Terminal) {
	var err error
	for isUsernameDuplicate(username) {
		writeln("Pick a different username")
		username, err = term.ReadLine()
		if err != nil {
			writeln(fmt.Sprint(err))
			fmt.Println(username, err)
		}
	}
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

		// Uses getWriteLnFunc closure to get a user-specific writeln func
		writeln := getWriteLnFunc(username, term)
		writers = append(writers, writeln)

		checkForDuplicateName(username, writeln, term)

		term.SetPrompt(username + ": ")
		usernames = append(usernames, username)
		defer func() { usernames = remove(usernames, username) }()
		defer broadcast(username + red.Sprint(" has left the chat."))

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

func remove(s []string, a string) []string {
	var i int
	for i = range s {
		if s[i] == a {
			break // i is now where it is
		}
	}
	return append(s[:i], s[i+1:]...)
}
