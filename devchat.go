package main

import (
	_ "embed"
	"fmt"
	"io"
	"log"
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
	// should this instance run offline? (should it not connect to slack or twitter?)
	offline = os.Getenv("DEVZAT_OFFLINE") != ""

	mainRoom         = &room{"#main", make([]*user, 0, 10), sync.Mutex{}}
	rooms            = map[string]*room{mainRoom.name: mainRoom}
	backlog          = make([]backlogMessage, 0, scrollback)
	bans             = make([]ban, 0, 10)
	idsInMinToTimes  = make(map[string]int, 10) // TODO: maybe add some IP-based factor to disallow rapid key-gen attempts
	antispamMessages = make(map[string]int)

	logfile, _  = os.OpenFile("log.txt", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	l           = log.New(io.MultiWriter(logfile, os.Stdout), "", log.Ldate|log.Ltime|log.Lshortfile)
	devbot      = "" // initialized in main
	startupTime = time.Now()
)

type ban struct {
	Addr string
	ID   string
}

type room struct {
	name       string
	users      []*user
	usersMutex sync.Mutex
}

type user struct {
	name     string
	pronouns []string
	session  ssh.Session
	term     *terminal.Terminal

	room      *room
	messaging *user // currently messaging this user in a DM

	bell          bool
	pingEverytime bool
	isSlack       bool
	formatTime24  bool

	color   string
	colorBG string
	id      string
	addr    string

	win           ssh.Window
	closeOnce     sync.Once
	lastTimestamp time.Time
	joinTime      time.Time
	timezone      *time.Location
}

type backlogMessage struct {
	timestamp  time.Time
	senderName string
	text       string
}

// TODO: have a web dashboard that shows logs
func main() {
	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", profilePort), nil)
		if err != nil {
			l.Println(err)
		}
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
	err = ssh.ListenAndServe(fmt.Sprintf(":%d", port), nil, ssh.HostKeyFile(os.Getenv("HOME")+"/.ssh/id_rsa"), ssh.PublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
		return true // allow all keys, this lets us hash pubkeys later
	}))
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

	toHash := ""

	pubkey := s.PublicKey()
	if pubkey != nil {
		toHash = string(pubkey.Marshal())
	} else { // If we can't get the public key fall back to the IP.
		toHash = host
	}

	u := &user{
		name:          s.User(),
		pronouns:      []string{"unset"},
		session:       s,
		term:          term,
		bell:          true,
		id:            shasum(toHash),
		addr:          host,
		win:           w,
		lastTimestamp: time.Now(),
		joinTime:      time.Now(),
		room:          mainRoom}

	go func() {
		for u.win = range winChan {
		}
	}()

	l.Println("Connected " + u.name + " [" + u.id + "]")

	if bansContains(bans, u.addr, u.id) {
		l.Println("Rejected " + u.name + " [" + host + "]")
		u.writeln(devbot, "**You are banned**. If you feel this was a mistake, please reach out at github.com/quackduck/devzat/issues or email igoel.mail@gmail.com. Please include the following information: [ID "+u.id+"]")
		u.closeBackend()
		return nil
	}
	idsInMinToTimes[u.id]++
	time.AfterFunc(60*time.Second, func() {
		idsInMinToTimes[u.id]--
	})
	if idsInMinToTimes[u.id] > 6 {
		bans = append(bans, ban{u.addr, u.id})
		mainRoom.broadcast(devbot, u.name+" has been banned automatically. ID: "+u.id)
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

	if err := u.pickUsername(s.User()); err != nil { // user exited or had some error
		l.Println(err)
		s.Close()
		return nil
	}

	mainRoom.usersMutex.Lock()
	mainRoom.users = append(mainRoom.users, u)
	go sendCurrentUsersTwitterMessage()
	mainRoom.usersMutex.Unlock()

	u.term.SetBracketedPasteMode(true) // experimental paste bracketing support

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

// cleanupRoom deletes a room if it's empty and isn't the main room
func cleanupRoom(r *room) {
	if r != mainRoom && len(r.users) == 0 {
		delete(rooms, r.name)
	}
}

// Removes a user and prints Twitter and chat message
func (u *user) close(msg string) {
	u.closeOnce.Do(func() {
		u.closeBackend()
		go sendCurrentUsersTwitterMessage()
		u.room.broadcast(devbot, msg)
		if time.Since(u.joinTime) > time.Minute/2 {
			u.room.broadcast(devbot, u.name+" stayed on for "+printPrettyDuration(time.Since(u.joinTime)))
		}
		u.room.users = remove(u.room.users, u)
		cleanupRoom(u.room)
	})
}

// Removes a user silently, used to close banned users
func (u *user) closeBackend() {
	u.room.usersMutex.Lock()
	u.room.users = remove(u.room.users, u)
	u.room.usersMutex.Unlock()
	u.session.Close()
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
	if u.pingEverytime && senderName != u.name {
		msg += "\a"
	}
	if !u.bell {
		msg = strings.ReplaceAll(msg, "\a", "")
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

func (u *user) pickUsername(possibleName string) error {
	possibleName = cleanName(possibleName)
	var err error
	for {
		if possibleName == "" {
		} else if strings.HasPrefix(possibleName, "#") || possibleName == "devbot" {
			u.writeln("", "Your username is invalid. Pick a different one:")
		} else if userDuplicate(u.room, possibleName) {
			u.writeln("", "Your username is already in use. Pick a different one:")
		} else {
			possibleName = cleanName(possibleName)
			break
		}

		u.term.SetPrompt("> ")
		possibleName, err = u.term.ReadLine()
		if err != nil {
			return err
		}
		possibleName = cleanName(possibleName)
	}

	u.name = possibleName
	u.initColor()

	if rand.Float64() <= 0.4 { // 40% chance of being a random color
		u.changeColor("random") // also sets prompt
		return nil
	}
	u.changeColor(styles[rand.Intn(len(styles))].name)
	return nil
}

func (u *user) displayPronouns() string {
	result := ""
	for i := 0; i < len(u.pronouns); i++ {
		str, _ := applyColorToData(u.pronouns[i], u.color, u.colorBG)
		result += "/" + str
	}
	if result == "" {
		return result
	}
	return result[1:]
}

func (u *user) changeRoom(r *room) {
	if u.room == r {
		return
	}
	u.room.users = remove(u.room.users, u)
	u.room.broadcast("", u.name+" is joining "+blue.Paint(r.name)) // tell the old room
	cleanupRoom(u.room)
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
		if err == io.EOF {
			u.close(u.name + " has left the chat")
			return
		}
		line += "\n"
		//oldPrompt := u.name + ": "
		for err == terminal.ErrPasteIndicator {
			//u.term.SetPrompt(strings.Repeat(" ", lenString(u.name)+2))
			u.term.SetPrompt("")
			additionalLine := ""
			additionalLine, err = u.term.ReadLine()
			additionalLine = strings.ReplaceAll(additionalLine, `\n`, `\\n`)
			//additionalLine = strings.ReplaceAll(additionalLine, "\t", strings.Repeat(" ", 8))
			line += additionalLine + "\n"
		}
		u.term.SetPrompt(u.name + ": ")
		line = strings.TrimSpace(line)

		if err != nil {
			l.Println(u.name, err)
			u.close(u.name + " has left the chat due to an error: " + err.Error())
			return
		}

		//fmt.Println("window", u.win)
		u.term.Write([]byte(strings.Repeat("\033[A\033[2K", calculateLinesTaken(u.name+": "+line, u.win.Width))))

		if line == "" {
			continue
		}

		antispamMessages[u.id]++
		time.AfterFunc(5*time.Second, func() {
			antispamMessages[u.id]--
		})
		if antispamMessages[u.id] >= 30 {
			u.room.broadcast(devbot, u.name+", stop spamming or you could get banned.")
		}
		if antispamMessages[u.id] >= 50 {
			if !bansContains(bans, u.addr, u.id) {
				bans = append(bans, ban{u.addr, u.id})
				saveBans()
			}
			u.writeln(devbot, "anti-spam triggered")
			u.close(red.Paint(u.name + " has been banned for spamming"))
			return
		}
		runCommands(line, u)
	}
}

// may contain a bug (may because it could be the terminal's fault)
func calculateLinesTaken(s string, width int) int {
	s = stripansi.Strip(s)
	//fmt.Println("`"+s+"`", "width", width)
	pos := 0
	lines := 1
	currLine := ""
	for _, c := range s {
		pos++
		currLine += string(c)
		if c == '\t' {
			pos += 8
		}
		if c == '\n' || pos > width { // || (c == '\t' && pos+8 > width)
			pos = 1
			lines++
		}
		//fmt.Println(string(c), "`"+currLine+"`", "pos", pos, "lines", lines)
	}
	return lines
}

// bansContains reports if the addr or id is found in the bans list
func bansContains(b []ban, addr string, id string) bool {
	for i := 0; i < len(b); i++ {
		if b[i].Addr == addr || b[i].ID == id {
			return true
		}
	}
	return false
}
