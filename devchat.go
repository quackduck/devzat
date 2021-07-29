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
	"net/http"
	_ "net/http/pprof"
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
	port        = 22
	scrollback  = 16
	profilePort = 5555
	// should this instance run offline?
	offline = os.Getenv("DEVZAT_OFFLINE") != ""

	mainRoom         = &room{"#main", make([]*user, 0, 10), sync.Mutex{}}
	rooms            = map[string]*room{mainRoom.name: mainRoom}
	backlog          = make([]backlogMessage, 0, scrollback)
	bans             = make([]string, 0, 10)
	idsInMinToTimes  = make(map[string]int, 10)
	antispamMessages = make(map[string]int)

	logfile, _  = os.OpenFile("log.txt", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	l           = log.New(io.MultiWriter(logfile, os.Stdout), "", log.Ldate|log.Ltime|log.Lshortfile)
	devbot      = "" // initialized in main
	startupTime = time.Now()
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
	pingEverytime bool
	color         string
	id            string
	addr          string
	win           ssh.Window
	closeOnce     sync.Once
	lastTimestamp time.Time
	joinTime      time.Time
	timezone      *time.Location
	formatTime24  bool
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
	go func() {
		http.ListenAndServe(fmt.Sprintf(":%d", profilePort), nil)
	}()
	devbot = green.Paint("devbot")
	rand.Seed(time.Now().Unix())
	readBans()
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		<-c
		fmt.Println("Shutting down...")
		saveBans()
		logfile.Close()
		time.AfterFunc(time.Second, func() {
			l.Println("Broadcast taking too long, exiting server early.")
			os.Exit(4)
		})
		universeBroadcast(devbot, "Server going down! This is probably because it is being updated. Try joining back immediately.  \n"+
			"If you still can't join, try joining back in 2 minutes. If you _still_ can't join, make an issue at github.com/quackduck/devzat/issues")
		os.Exit(0)
	}()
	ssh.Handle(func(s ssh.Session) {
		u := newUser(s)
		if u == nil {
			s.Close()
			return
		}
		defer func() { // crash protection
			if i := recover(); i != nil {
				mainRoom.broadcast(devbot, "Slap the developers in the face for me, the server almost crashed, also tell them this: "+fmt.Sprint(i))
			}
		}()
		u.repl()
	})
	var err error
	if os.Getenv("PORT") != "" {
		port, err = strconv.Atoi(os.Getenv("PORT"))
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	fmt.Printf("Starting chat server on port %d and profiling on port %d\n", port, profilePort)
	go getMsgsFromSlack()
	go func() {
		if port == 22 {
			fmt.Println("Also starting chat server on port 443")
			err = ssh.ListenAndServe(":443", nil, ssh.HostKeyFile(os.Getenv("HOME")+"/.ssh/id_rsa"))
			if err != nil {
				fmt.Println(err)
			}
		}
	}()
	err = ssh.ListenAndServe(fmt.Sprintf(":%d", port), nil, ssh.HostKeyFile(os.Getenv("HOME")+"/.ssh/id_rsa"))
	if err != nil {
		fmt.Println(err)
	}
}

func universeBroadcast(senderName, msg string) {
	for _, r := range rooms {
		r.broadcast(senderName, msg)
	}
}

func (r *room) broadcast(senderName, msg string) {
	if msg == "" {
		return
	}
	if senderName != "" {
		slackChan <- "[" + r.name + "] " + senderName + ": " + msg
	} else {
		slackChan <- "[" + r.name + "] " + msg
	}
	r.broadcastNoSlack(senderName, msg)
}

func (r *room) broadcastNoSlack(senderName, msg string) {
	if msg == "" {
		return
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
	if r == mainRoom {
		backlog = append(backlog, backlogMessage{time.Now(), senderName, msg + "\n"})
		if len(backlog) > scrollback {
			backlog = backlog[len(backlog)-scrollback:]
		}
	}
}

func newUser(s ssh.Session) *user {
	term := terminal.NewTerminal(s, "> ")
	_ = term.SetSize(10000, 10000) // disable any formatting done by term
	pty, winChan, _ := s.Pty()
	w := pty.Window
	host, _, _ := net.SplitHostPort(s.RemoteAddr().String()) // definitely should not give an err
	hash := sha256.New()
	pubkey := s.PublicKey()
	if pubkey != nil {
		fmt.Printf("key: %s\n", pubkey.Marshal())
		hash.Write([]byte(pubkey.Marshal()))
	} else { // If we can't get the public key fall back to the IP.
		hash.Write([]byte(host))
	}

	u := &user{
		name:          s.User(),
		session:       s,
		term:          term,
		bell:          true,
		id:            hex.EncodeToString(hash.Sum(nil)),
		addr:          host,
		win:           w,
		lastTimestamp: time.Now(),
		joinTime:      time.Now(),
		timezone:      nil,
		formatTime24:  false,
		room:          mainRoom}

	go func() {
		for u.win = range winChan {
		}
	}()

	l.Println("Connected " + u.name + " [" + u.id + "]")

	for i := range bans {
		if u.addr == bans[i] || u.id == bans[i] { // allow banning by ID
			if u.id == bans[i] { // then replace the ID in the ban with the actual IP
				bans[i] = u.addr
				saveBans()
			}
			l.Println("Rejected " + u.name + " [" + u.addr + "]")
			u.writeln(devbot, "**You are banned**. If you feel this was done wrongly, please reach out at github.com/quackduck/devzat/issues. Please include the following information: [ID "+u.id+"]")
			u.close("")
			return nil
		}
	}
	idsInMinToTimes[u.id]++
	time.AfterFunc(60*time.Second, func() {
		idsInMinToTimes[u.id]--
	})
	if idsInMinToTimes[u.id] > 6 {
		bans = append(bans, u.addr)
		mainRoom.broadcast(devbot, u.name+" has been banned automatically. IP: "+u.addr)
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

	if !u.pickUsername(s.User()) { // user exited / had err
		s.Close()
		return nil
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
	mainRoom.broadcast(devbot, u.name+" has joined the chat")
	return u
}

func (u *user) close(msg string) {
	u.closeOnce.Do(func() {
		u.room.usersMutex.Lock()
		u.room.users = remove(u.room.users, u)
		u.room.usersMutex.Unlock()
		go sendCurrentUsersTwitterMessage()
		u.room.broadcast(devbot, msg)
		if time.Since(u.joinTime) > time.Minute/2 {
			u.room.broadcast(devbot, u.name+" stayed on for "+printPrettyDuration(time.Since(u.joinTime)))
		}
		u.session.Close()
	})
}

func (u *user) writeln(senderName string, msg string) {
	if strings.Contains(msg, u.name) { // is a ping
		msg += "\a"
	}
	msg = strings.ReplaceAll(msg, `\n`, "\n")
	msg = strings.ReplaceAll(msg, `\`+"\n", `\n`) // let people escape newlines
	if senderName != "" {
		if strings.HasSuffix(senderName, " <- ") || strings.HasSuffix(senderName, " -> ") { // kinda hacky DM detection
			msg = strings.TrimSpace(mdRender(msg, lenString(senderName), u.win.Width))
			msg = senderName + msg + "\a"
		} else {
			msg = strings.TrimSpace(mdRender(msg, lenString(senderName)+2, u.win.Width))
			msg = senderName + ": " + msg
		}
	} else {
		msg = strings.TrimSpace(mdRender(msg, 0, u.win.Width)) // No sender
	}
	if time.Since(u.lastTimestamp) > time.Minute {
		if u.timezone == nil {
			u.rWriteln(printPrettyDuration(time.Since(u.joinTime)) + " in")
		} else {
			if u.formatTime24 {
				u.rWriteln(time.Now().In(u.timezone).Format("15:04"))
			} else {
				u.rWriteln(time.Now().In(u.timezone).Format("3:04 pm"))
			}
		}
		u.lastTimestamp = time.Now()
	}
	if !u.bell {
		msg = strings.ReplaceAll(msg, "\a", "")
	}
	if u.pingEverytime && senderName != u.name {
		msg += "\a"
	}
	u.term.Write([]byte(msg + "\n"))
}

// Write to the right of the user's window
func (u *user) rWriteln(msg string) {
	if u.win.Width-lenString(msg) > 0 {
		u.term.Write([]byte(strings.Repeat(" ", u.win.Width-lenString(msg)) + msg + "\n"))
	} else {
		u.term.Write([]byte(msg + "\n"))
	}
}

func (u *user) pickUsername(possibleName string) (ok bool) {
	possibleName = cleanName(possibleName)
	var err error
	for possibleName == "" || possibleName == "devbot" || strings.HasPrefix(possibleName, "#") || userDuplicate(u.room, possibleName) {
		u.writeln("", "Your username is already in use. Pick a different one:")
		u.term.SetPrompt("> ")
		possibleName, err = u.term.ReadLine()
		if err != nil {
			l.Println(err)
			return false
		}
		possibleName = cleanName(possibleName)
	}
	u.name = possibleName
	idx := rand.Intn(len(styles) * 140 / 100) // 40% chance of a random color
	if idx >= len(styles) {                   // allow the possibility of having a completely random RGB color
		u.changeColor("random")
		return true
	}
	u.changeColor(styles[idx].name) // also sets prompt
	return true
}

func (u *user) changeRoom(r *room) {
	if u.room == r {
		return
	}
	u.room.users = remove(u.room.users, u)
	u.room.broadcast("", u.name+" is joining "+blue.Paint(r.name)) // tell the old room
	if u.room != mainRoom && len(u.room.users) == 0 {
		delete(rooms, u.room.name)
	}
	u.room = r
	if userDuplicate(u.room, u.name) {
		u.pickUsername("")
	}
	u.room.users = append(u.room.users, u)
	u.room.broadcast(devbot, u.name+" has joined "+blue.Paint(u.room.name))
}

func (u *user) repl() {
	for {
		line, err := u.term.ReadLine()
		line = strings.TrimSpace(line)

		if err == io.EOF {
			u.close(u.name + " has left the chat")
			return
		}
		if err != nil {
			l.Println(u.name, err)
			u.close(u.name + " has left the chat due to an error")
			return
		}
		u.term.Write([]byte(strings.Repeat("\033[A\033[2K", int(math.Ceil(float64(lenString(u.name+line)+2)/(float64(u.win.Width))))))) // basically, ceil(length of line divided by term width)

		antispamMessages[u.id]++
		time.AfterFunc(5*time.Second, func() {
			antispamMessages[u.id]--
		})
		if antispamMessages[u.id] >= 50 {
			if !stringsContain(bans, u.addr) {
				bans = append(bans, u.addr)
				saveBans()
			}
			u.writeln(devbot, "anti-spam triggered")
			u.close(red.Paint(u.name + " has been banned for spamming"))
			return
		}
		runCommands(line, u, false)
	}
}
