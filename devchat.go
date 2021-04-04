package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	//"golang.org/x/term"
	"io"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/gliderlabs/ssh"
	terminal "golang.org/x/term"
)

func main() {
	writers := make([]func(a ...interface{}), 0, 5)
	usernames := make([]string, 0, 5)
	red := color.New(color.FgHiRed)
	green := color.New(color.FgHiGreen)
	cyan := color.New(color.FgHiCyan)
	magenta := color.New(color.FgHiMagenta)
	yellow := color.New(color.FgHiYellow)

	colorArr := []*color.Color{red, green, cyan, magenta, yellow}

	messagesBuffer := make([]string, 0, 10)

	ssh.Handle(func(s ssh.Session) {
		myColor := *colorArr[rand.Intn(len(colorArr))]
		username := myColor.Sprint(s.User())
		term := terminal.NewTerminal(s, "> ")
		//_, w, ptyReq := s.Pty()
		//if !ptyReq {
		term.SetSize(10000, 10000) // disable any formatting done by term
		//} else { // resize term on win resize
		//go func() {
		//	for win := range w {
		//		term.SetSize(win.Width, win.Height)
		//	}
		//}()
		//}
		rand.Seed(time.Now().Unix())
		writeln := func(a ...interface{}) {
			msg := fmt.Sprintln(a...)
			if !strings.HasPrefix(msg, username+":") {
				term.Write([]byte("\a"+msg)) // "\a" is beep
			}
		}
		broadcast := func(a ...interface{}) {
			messagesBuffer = append(messagesBuffer, fmt.Sprintln(a...))
			for len(messagesBuffer) > 16 { // for instead of if just in case
				fmt.Println(messagesBuffer)
				//time.Sleep(time.Second)
				messagesBuffer = messagesBuffer[1:]
			}
			for i := range writers {
				writers[i](a...)
			}
		}
		writers = append(writers, writeln)
		userDuplicate := func(a string) bool {
			for _, u := range usernames {
				if u == a {
					return true
				}
			}
			return false
		}

		for userDuplicate(username) {
			var err error
			writeln("Pick a different username")
			username, err = term.ReadLine()
			if err != nil {
				writeln(err)
				fmt.Println(username, err)
			}
		}

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
		term.Write([]byte(strings.Join(messagesBuffer, "")))

		broadcast(username + green.Sprint(" has joined the chat"))
		for {
			line, err := term.ReadLine()
			line = strings.TrimSpace(line)

			if err == io.EOF {
				return
			}
			if err != nil {
				writeln(err)
				fmt.Println(username, err)
				continue
			}
			if !(line == "") {
				broadcast(username + ": " + line)
			}
			if line == "/users" {
				broadcast(usernames)
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
	fmt.Println("Starting chat server on port 2222")
	log.Fatal(ssh.ListenAndServe(":2222", nil, ssh.HostKeyFile(os.Getenv("HOME")+"/.ssh/id_rsa")))
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
//
//func removeI(s []string, i int) []string {
//	fmt.Println("appending" , s[:i], s[i+1:])
//	return append(s[:i], s[i+1:]...)
//}
