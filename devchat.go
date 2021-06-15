package main

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/acarl005/stripansi"
	"github.com/gliderlabs/ssh"
	terminal "golang.org/x/term"
)

var (
	port = 22

	devbot = "" // initialized in main

	startupTime = time.Now()

	mainRoom = &room{"#main", make([]*user, 0, 10), sync.Mutex{}}
	rooms    = map[string]*room{mainRoom.name: mainRoom}

	allUsers      = make(map[string]string, 400) //map format is u.id => u.name
	allUsersMutex = sync.Mutex{}

	backlog      = make([]backlogMessage, 0, scrollback)
	backlogMutex = sync.Mutex{}

	logfile, _ = os.OpenFile("log.txt", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	l          = log.New(io.MultiWriter(logfile, os.Stdout), "", log.Ldate|log.Ltime|log.Lshortfile)

	bans            = make([]string, 0, 10)
	bansMutex       = sync.Mutex{}
	idsInMinToTimes = make(map[string]int, 10)
	idsInMinMutex   = sync.Mutex{}

	antispamMessages = make(map[string]int)
	antispamMutex    = sync.Mutex{}
)

type room struct {
	name       string
	users      []*user
	usersMutex sync.Mutex
}

type user struct {
	name          string
	session       ssh.Session
	term          *terminal.Terminal
	bell          bool
	color         string
	id            string
	addr          string
	win           ssh.Window
	closeOnce     sync.Once
	lastTimestamp time.Time
	joinTime      time.Time
	timezone      *time.Location
	room          *room
	messaging     *user
}

type backlogMessage struct {
	timestamp  time.Time
	senderName string
	text       string
}

// TODO: have a web dashboard that shows logs
func main() {
	devbot = green.Paint("devbot")
	var err error
	rand.Seed(time.Now().Unix())
	readBansAndUsers()
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGKILL)
	go func() {
		<-c
		fmt.Println("Shutting down...")
		saveBansAndUsers()
		logfile.Close()
		mainRoom.broadcast(devbot, "Server going down! This is probably because it is being updated. Try joining back immediately.  \n"+
			"If you still can't join, try joining back in 2 minutes. If you _still_ can't join, make an issue at github.com/quackduck/devzat/issues", true)
		os.Exit(0)
	}()
	ssh.Handle(func(s ssh.Session) {
		u := newUser(s)
		if u == nil {
			return
		}
		u.repl()
	})
	if os.Getenv("PORT") != "" {
		port, err = strconv.Atoi(os.Getenv("PORT"))
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	fmt.Printf("Starting chat server on port %d\n", port)
	go getMsgsFromSlack()
	go func() {
		if port == 22 {
			fmt.Println("Also starting chat server on port 443")
			err = ssh.ListenAndServe(
				":443",
				nil,
				ssh.HostKeyFile(os.Getenv("HOME")+"/.ssh/id_rsa"))
		}
	}()
	err = ssh.ListenAndServe(
		fmt.Sprintf(":%d", port),
		nil,
		ssh.HostKeyFile(os.Getenv("HOME")+"/.ssh/id_rsa"))
	if err != nil {
		fmt.Println(err)
	}
}

func (r *room) broadcast(senderName, msg string, toSlack bool) {
	if msg == "" {
		return
	}
	if toSlack {
		if senderName != "" {
			slackChan <- "[" + r.name + "] " + senderName + ": " + msg
		} else {
			slackChan <- "[" + r.name + "] " + msg
		}
	}
	msg = strings.ReplaceAll(msg, "@everyone", green.Paint("everyone\a"))
	r.usersMutex.Lock()
	for i := range r.users {
		msg = strings.ReplaceAll(msg, "@"+stripansi.Strip(r.users[i].name), r.users[i].name)
		msg = strings.ReplaceAll(msg, `\`+r.users[i].name, "@"+stripansi.Strip(r.users[i].name)) // allow escaping
	}
	for i := range r.users {
		r.users[i].writeln(senderName, msg)
	}
	r.usersMutex.Unlock()
	if r.name == "#main" {
		backlogMutex.Lock()
		backlog = append(backlog, backlogMessage{time.Now(), senderName, msg + "\n"})
		backlogMutex.Unlock()
		for len(backlog) > scrollback { // for instead of if just in case
			backlog = backlog[1:]
		}
	}
}

func newUser(s ssh.Session) *user {
	term := terminal.NewTerminal(s, "> ")
	_ = term.SetSize(10000, 10000) // disable any formatting done by term
	pty, winChan, _ := s.Pty()
	w := pty.Window
	host, _, err := net.SplitHostPort(s.RemoteAddr().String()) // definitely should not give an err
	if err != nil {
		term.Write([]byte(err.Error() + "\n"))
		s.Close()
		return nil
	}
	hash := sha256.New()
	hash.Write([]byte(host))
	u := &user{s.User(), s, term, true, "", hex.EncodeToString(hash.Sum(nil)), host, w, sync.Once{}, time.Now(), time.Now(), nil, mainRoom, nil}
	go func() {
		for u.win = range winChan {
		}
	}()
	l.Println("Connected " + u.name + " [" + u.id + "]")
	for i := range bans {
		if u.addr == bans[i] || u.id == bans[i] { // allow banning by ID
			if u.id == bans[i] { // then replace the ID in the ban with the actual IP
				bans[i] = u.addr
				saveBansAndUsers()
			}
			l.Println("Rejected " + u.name + " [" + u.addr + "]")
			u.writeln(devbot, "**You are banned**. If you feel this was done wrongly, please reach out at github.com/quackduck/devzat/issues. Please include the following information: [ID "+u.id+"]")
			u.close("")
			return nil
		}
	}
	idsInMinMutex.Lock()
	idsInMinToTimes[u.id]++
	idsInMinMutex.Unlock()
	time.AfterFunc(60*time.Second, func() {
		idsInMinMutex.Lock()
		idsInMinToTimes[u.id]--
		idsInMinMutex.Unlock()
	})
	if idsInMinToTimes[u.id] > 6 {
		bansMutex.Lock()
		bans = append(bans, u.addr)
		bansMutex.Unlock()
		mainRoom.broadcast(devbot, u.name+" has been banned automatically. IP: "+u.addr, true)
		return nil
	}
	if len(backlog) > 0 {
		lastStamp := backlog[0].timestamp
		u.rWriteln(printPrettyDuration(u.joinTime.Sub(lastStamp)) + " earlier")
		for i := range backlog {
			if backlog[i].timestamp.Sub(lastStamp) > time.Minute {
				lastStamp = backlog[i].timestamp
				u.rWriteln(printPrettyDuration(u.joinTime.Sub(lastStamp)) + " earlier")
			}
			u.writeln(backlog[i].senderName, backlog[i].text)
		}
	}

	u.pickUsername(s.User())
	if _, ok := allUsers[u.id]; !ok {
		mainRoom.broadcast(devbot, "You seem to be new here "+u.name+". Welcome to Devzat! Run ./help to see what you can do.", true)
	}
	mainRoom.usersMutex.Lock()
	mainRoom.users = append(mainRoom.users, u)
	go sendCurrentUsersTwitterMessage()
	mainRoom.usersMutex.Unlock()
	switch len(mainRoom.users) - 1 {
	case 0:
		u.writeln("", blue.Paint("Welcome to the chat. There are no more users"))
	case 1:
		u.writeln("", yellow.Paint("Welcome to the chat. There is one more user"))
	default:
		u.writeln("", green.Paint("Welcome to the chat. There are", strconv.Itoa(len(mainRoom.users)-1), "more users"))
	}
	//_, _ = term.Write([]byte(strings.Join(backlog, ""))) // print out backlog
	mainRoom.broadcast(devbot, u.name+green.Paint(" has joined the chat"), true)
	return u
}

func (u *user) close(msg string) {
	u.closeOnce.Do(func() {
		u.room.usersMutex.Lock()
		u.room.users = remove(u.room.users, u)
		u.room.usersMutex.Unlock()
		go sendCurrentUsersTwitterMessage()
		u.room.broadcast(devbot, msg, true)
		if time.Since(u.joinTime) > time.Minute/2 {
			u.room.broadcast(devbot, u.name+" stayed on for "+printPrettyDuration(time.Since(u.joinTime)), true)
		}
		u.session.Close()
	})
}

func (u *user) writeln(senderName string, msg string) {
	if u.bell {
		if strings.Contains(msg, u.name) { // is a ping
			msg += "\a"
		}
	}
	msg = strings.ReplaceAll(msg, `\n`, "\n")
	msg = strings.ReplaceAll(msg, `\`+"\n", `\n`) // let people escape newlines
	if senderName != "" {
		//msg = strings.TrimSpace(mdRender(msg, len(stripansi.Strip(senderName))+2, u.win.Width))
		if strings.HasSuffix(senderName, " <- ") || strings.HasSuffix(senderName, " -> ") { // kinda hacky
			msg = strings.TrimSpace(mdRender(msg, len(stripansi.Strip(senderName)), u.win.Width))
			msg = senderName + msg + "\a"
		} else {
			msg = strings.TrimSpace(mdRender(msg, len(stripansi.Strip(senderName))+2, u.win.Width))
			msg = senderName + ": " + msg
		}
	} else {
		msg = strings.TrimSpace(mdRender(msg, 0, u.win.Width)) // No sender
	}
	if time.Since(u.lastTimestamp) > time.Minute {
		if u.timezone == nil {
			u.rWriteln(printPrettyDuration(time.Since(u.joinTime)) + " in")
		} else {
			u.rWriteln(time.Now().In(u.timezone).Format("3:04 pm"))
		}
		u.lastTimestamp = time.Now()
	}
	u.term.Write([]byte(msg + "\n"))
}

// Write to the right of the user's window
func (u *user) rWriteln(msg string) {
	if u.win.Width-len([]rune(msg)) > 0 {
		u.term.Write([]byte(strings.Repeat(" ", u.win.Width-len([]rune(msg))) + msg + "\n"))
	} else {
		u.term.Write([]byte(msg + "\n"))
	}
}

func (u *user) pickUsername(possibleName string) {
	possibleName = cleanName(possibleName)
	var err error
	for userDuplicate(u.room, possibleName) || possibleName == "" || possibleName == "devbot" {
		u.writeln("", "Pick a different username")
		u.term.SetPrompt("> ")
		possibleName, err = u.term.ReadLine()
		if err != nil {
			l.Println(err)
			return
		}
		possibleName = cleanName(possibleName)
	}
	u.name = possibleName
	u.changeColor(styles[rand.Intn(len(styles))].name) // also sets prompt
}

func (u *user) changeRoom(r *room, toSlack bool) {
	u.room.users = remove(u.room.users, u)
	u.room.broadcast("", u.name+" is joining "+blue.Paint(r.name), toSlack) // tell the old room
	if u.room != mainRoom && len(u.room.users) == 0 {
		delete(rooms, u.room.name)
	}
	u.room = r
	if userDuplicate(u.room, u.name) {
		u.pickUsername("")
	}
	u.room.users = append(u.room.users, u)
	u.room.broadcast(devbot, u.name+" has joined "+blue.Paint(u.room.name), toSlack)
}

func (u *user) repl() {
	for {
		line, err := u.term.ReadLine()
		line = strings.TrimSpace(line)

		if err == io.EOF {
			u.close(u.name + red.Paint(" has left the chat"))
			return
		}
		if err != nil {
			l.Println(u.name, err)
			continue
		}
		u.term.Write([]byte(strings.Repeat("\033[A\033[2K", int(math.Ceil(float64(len([]rune(u.name+line))+2)/(float64(u.win.Width))))))) // basically, ceil(length of line divided by term width)

		antispamMutex.Lock()
		antispamMessages[u.id]++
		antispamMutex.Unlock()
		time.AfterFunc(5*time.Second, func() {
			antispamMutex.Lock()
			antispamMessages[u.id]--
			antispamMutex.Unlock()
		}
		if currentMessages >= 50 {
			bans = append(bans, u.addr)
			u.writeln(devbot, "Anti-Spam triggered")
			u.close(red.Paint(u.name + " has been banned for spamming"))
		runCommands(line, u, false)
	}
}
