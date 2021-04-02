package main

import (
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"strings"

	//"golang.org/x/term"
	"io"
	"log"
	"os"

	"github.com/gliderlabs/ssh"
)

func main() {
	msgs := make(chan string, 5)
	users := make([]string, 0, 5)

	ssh.Handle(func(s ssh.Session) {
		username := s.User()
		term := terminal.NewTerminal(s, "> ")
		writeln := func(a ...interface{}) {
			term.Write([]byte(fmt.Sprintln(a...)))
		}
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
			writeln("Pick a different username:")
			username, err = term.ReadLine()
			if err != nil {
				writeln(err)
				fmt.Println(username, err)
			}
		}

		term = terminal.NewTerminal(s, username+" ")
		users = append(users, username)
		defer func() { users = remove(users, username) }()
		fmt.Println(users)

		//reader := bufio.NewReader(s)
		writeln("Welcome to the chat", "There are", len(users)-1, "more users")
		msgs <- username + " has joined the chat"
		//b := make([]byte, 12)
		//_, err := s.Read(b)
		//if err != nil {
		//	log.Println(err)
		//}
		//fmt.Println(string(b))
		go func() {
			for msg := range msgs {
				if !strings.HasPrefix(msg, username) {
					writeln(msg)
				}
			}
		}()
		for {
			line, err := term.ReadLine()
			line = strings.TrimSpace(line)

			if err == io.EOF {
				msgs <- username + " has left the chat."
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
			msgs <- username + line
		}
	})
	fmt.Println("Starting chat server on port 2222")
	log.Fatal(ssh.ListenAndServe(":2222", nil, ssh.HostKeyFile(os.Getenv("HOME")+"/.ssh/id_rsa")))
}

func remove(slice []string, a string) []string {
	var i int
	for i = range slice {
		if slice[i] == a {
			break // i is now where it is
		}
	}
	fmt.Println(i)
	slice[i] = slice[len(slice)-1]
	return slice[:len(slice)-1]
}
