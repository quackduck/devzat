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
	users := make([]string, 0, 5)
	red := color.New(color.FgHiRed)
	green := color.New(color.FgHiGreen)
	cyan := color.New(color.FgHiCyan)
	magenta := color.New(color.FgHiMagenta)
	yellow := color.New(color.FgHiYellow)

	colorArr := []*color.Color{red, green, cyan, magenta, yellow}

	ssh.Handle(func(s ssh.Session) {
		myColor := *colorArr[rand.Intn(len(colorArr))]
		username := myColor.Sprint(s.User())
		term := terminal.NewTerminal(s, "> ")
		term.SetSize(2000, 2000) // disable any formatting done by term
		rand.Seed(time.Now().Unix())

		writeln := func(a ...interface{}) {
			msg := fmt.Sprintln(a...)
			if !strings.HasPrefix(msg, username) {
				term.Write([]byte(time.Now().Format(time.Kitchen) + " " + msg))
			}
		}
		sendMsg := func(a ...interface{}) {
			for i := range writers {
				writers[i](a...)
			}
		}

		writers = append(writers, writeln)
		//write := func(a ...interface{}) {
		//	term.Write([]byte(fmt.Sprint(a...)))
		//}
		userDuplicate := func(a string) bool {
			for _, u := range users {
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
		users = append(users, username)
		defer func() { users = remove(users, username) }()

		//reader := bufio.NewReader(s)
		switch len(users) - 1 {
		case 0:
			writeln(cyan.Sprint("Welcome to the chat. There are no more users"))
		case 1:
			writeln(cyan.Sprint("Welcome to the chat. There is one more user"))
		default:
			writeln(cyan.Sprint("Welcome to the chat. There are ", len(users)-1, " more users"))
		}

		sendMsg(username + green.Sprint(" has joined the chat"))
		//b := make([]byte, 12)
		//_, err := s.Read(b)
		//if err != nil {
		//	log.Println(err)
		//}
		//fmt.Println(string(b))
		//go func() {
		//	for msg := range msgs {
		//		if !strings.HasPrefix(msg, username) {
		//			fmt.Println(msg)
		//			writeln(msg)
		//		}
		//	}
		//}()
		for {
			line, err := term.ReadLine()
			line = strings.TrimSpace(line)

			if err == io.EOF {
				sendMsg(username + red.Sprint(" has left the chat."))
				return
			}
			if err != nil {
				writeln(err)
				fmt.Println(username, err)
				continue
			}
			//msg, err := reader.ReadString('\n')
			//if err == io.EOF {
			//	fmt.Println("EOF")
			//	break
			//}
			//if err != nil {
			//	log.Println(err)
			//	continue
			//}
			if !(line == "") {
				sendMsg(username + ": " + line)
			}
			if line == "/users" {
				sendMsg(users)
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
