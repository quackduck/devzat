package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"syscall"
	"time"

	"github.com/gliderlabs/ssh"

	"devzat/pkg"
	"devzat/pkg/server"
)

const (
	errCodeAllGood = iota
	errCodeNoListenPort
	_
	_
	errCodeBroadcastTimeout
)

const (
	msgGoingDown = `
	Server going down!

	This is probably because it is being updated. Try joining back immediately.

	If you still can't join, try joining back in 2 minutes. If you _still_ can't join, make an issue at github.com/quackduck/devzat/issues
	`

	fmtRecover = "Slap the developers in the face for me, the server almost crashed, also tell them this: %v, stack: %v"
)

// TODO: have a web dashboard that shows logs
func main() {
	rand.Seed(time.Now().Unix())

	server := server.Server{}
	if err := server.Init(); err != nil {
		fmt.Printf("could not init chat server: %v", err)

		return
	}

	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", server.ProfilePort), nil)
		if err != nil {
			fmt.Println(err)
		}
	}()

	server.ReadBans()

	c := make(chan os.Signal, 2)

	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		<-c
		fmt.Println("Shutting down...")

		_ = server.SaveBans()
		_ = server.Logfile.Close()

		time.AfterFunc(time.Second, func() {
			server.Log.Println("Broadcast taking too long, exiting server early.")
			os.Exit(errCodeBroadcastTimeout)
		})

		botName := server.MainRoom.Bot.Name()

		server.UniverseBroadcast(botName, msgGoingDown)

		os.Exit(errCodeAllGood)
	}()

	ssh.Handle(func(s ssh.Session) {
		u, err := server.NewUser(s)
		if err != nil {
			server.Log.Printf("could not create user: %v", err)

			return
		}

		if u == nil {
			_ = s.Close()

			return
		}

		defer func() { // crash protection
			if i := recover(); i != nil {
				botName := server.MainRoom.Bot.Name()

				server.MainRoom.Broadcast(botName, fmt.Sprintf(fmtRecover, i, debug.Stack()))
			}
		}()

		u.Repl()
	})

	var err error

	envPort := os.Getenv(pkg.EnvServerPort)
	if envPort != "" {
		if server.Port, err = strconv.Atoi(envPort); err != nil {
			fmt.Println(err)
			os.Exit(errCodeNoListenPort)
		}
	}

	// Check for global offline for backwards compatibility
	if os.Getenv(pkg.EnvOffline) != "" {
		server.OfflineSlack = true
		server.OfflineTwitter = true
	}

	fmt.Printf("Starting chat server on port %d and profiling on port %d\n", server.Port, server.ProfilePort)

	go server.MainRoom.GetMsgsFromSlack()

	go func() {
		if server.Port == 22 {
			fmt.Println("Also starting chat server on port 443")
			err = ssh.ListenAndServe(":443", nil, ssh.HostKeyFile(os.Getenv("HOME")+"/.ssh/id_rsa"))
			if err != nil {
				fmt.Println(err)
			}
		}
	}()

	err = ssh.ListenAndServe(fmt.Sprintf(":%d", server.Port), nil, ssh.HostKeyFile(os.Getenv("HOME")+"/.ssh/id_rsa"), ssh.PublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
		return true // allow all keys, this lets us hash pubkeys later
	}))

	if err != nil {
		fmt.Println(err)
	}
}
