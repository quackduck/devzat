package main

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
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
	"unicode"

	markdown "github.com/MichaelMure/go-term-markdown"
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
	port       = 22
	scrollback = 16

	slackChan = getSendToSlackChan()
	api       = slack.New(string(slackAPI))
	rtm       = api.NewRTM()

	red      = color.New(color.FgHiRed)
	green    = color.New(color.FgHiGreen)
	cyan     = color.New(color.FgHiCyan)
	magenta  = color.New(color.FgHiMagenta)
	yellow   = color.New(color.FgHiYellow)
	blue     = color.New(color.FgHiBlue)
	black    = color.New(color.FgHiBlack)
	white    = color.New(color.FgHiWhite)
	colorArr = []*color.Color{green, cyan, magenta, yellow, white, blue}

	users      = make([]*user, 0, 10)
	usersMutex = sync.Mutex{}

	allUsers      = make(map[string]string, 100) //map format is u.id => u.name
	allUsersMutex = sync.Mutex{}

	backlog      = make([]message, 0, scrollback)
	backlogMutex = sync.Mutex{}

	logfile, _ = os.OpenFile("log.txt", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	l          = log.New(io.MultiWriter(logfile, os.Stdout), "", log.Ldate|log.Ltime|log.Lshortfile)

	bans      = make([]string, 0, 10)
	bansMutex = sync.Mutex{}

	// stores the ids which have joined in 20 seconds and how many times this happened
	idsIn20ToTimes = make(map[string]int, 10)
	idsIn20Mutex   = sync.Mutex{}
)

func broadcast(senderName, msg string, toSlack bool) {
	if msg == "" {
		return
	}
	backlogMutex.Lock()
	backlog = append(backlog, message{senderName, msg + "\n"})
	backlogMutex.Unlock()
	if toSlack {
		if senderName != "" {
			slackChan <- senderName + ": " + msg
		} else {
			slackChan <- msg
		}
	}
	for len(backlog) > scrollback { // for instead of if just in case
		backlog = backlog[1:]
	}
	for i := range users {
		users[i].writeln(senderName, msg)
	}
}

type user struct {
	name      string
	session   ssh.Session
	term      *terminal.Terminal
	bell      bool
	color     color.Color
	id        string
	addr      string
	win       ssh.Window
	closeOnce sync.Once
}

type message struct {
	senderName string
	text       string
}

func newUser(s ssh.Session) *user {
	term := terminal.NewTerminal(s, "> ")
	_ = term.SetSize(10000, 10000) // disable any formatting done by term
	pty, winchan, _ := s.Pty()
	w := pty.Window
	host, _, err := net.SplitHostPort(s.RemoteAddr().String()) // definitely should not give an err
	if err != nil {
		term.Write([]byte(fmt.Sprintln(err) + "\n"))
		s.Close()
		return nil
	}
	hash := sha256.New()
	hash.Write([]byte(host))
	u := &user{s.User(), s, term, true, color.Color{}, hex.EncodeToString(hash.Sum(nil)), host, w, sync.Once{}}
	go func() {
		for u.win = range winchan {
		}
	}()
	l.Println("Connected " + u.name)
	for _, banAddr := range bans {
		if u.addr == banAddr {
			l.Println("Rejected " + u.addr)
			u.writeln("", "**You have been banned**. If you feel this was done wrongly, please reach out at https://github.com/quackduck/devzat/issues")
			u.close("")
			return nil
		}
	}
	idsIn20Mutex.Lock()
	idsIn20ToTimes[u.id]++
	idsIn20Mutex.Unlock()
	time.AfterFunc(20*time.Second, func() {
		idsIn20Mutex.Lock()
		idsIn20ToTimes[u.id]--
		idsIn20Mutex.Unlock()
	})
	if idsIn20ToTimes[u.id] > 3 { // 10 minute ban
		bansMutex.Lock()
		bans = append(bans, u.addr)
		bansMutex.Unlock()
		broadcast("", u.name+" has been banned automatically. IP: "+u.addr, true)
		return nil
	}
	u.pickUsername(s.User())
	usersMutex.Lock()
	users = append(users, u)
	usersMutex.Unlock()
	saveBansAndUsers()
	switch len(users) - 1 {
	case 0:
		u.writeln("", "**"+cyan.Sprint("Welcome to the chat. There are no more users")+"**")
	case 1:
		u.writeln("", "**"+cyan.Sprint("Welcome to the chat. There is one more user")+"**")
	default:
		u.writeln("", "**"+cyan.Sprint("Welcome to the chat. There are ", len(users)-1, " more users")+"**")
	}
	//_, _ = term.Write([]byte(strings.Join(backlog, ""))) // print out backlog
	for i := range backlog {
		u.writeln(backlog[i].senderName, backlog[i].text)
	}
	broadcast("", "**"+u.name+"** **"+green.Sprint("has joined the chat")+"**", true)
	return u
}

func (u *user) repl() {
	for {
		line, err := u.term.ReadLine()
		line = clean(line)

		if err == io.EOF {
			return
		}
		if err != nil {
			l.Println(u.name, err)
			continue
		}
		inputLine := line
		u.term.Write([]byte(strings.Repeat("\033[A\033[2K", int(math.Ceil(float64(len([]rune(u.name+inputLine))+2)/(float64(u.win.Width))))))) // basically, ceil(length of line divided by term width)

		toSlack := true
		if strings.HasPrefix(line, "/hide") {
			toSlack = false
		}
		if !(line == "") {
			broadcast(u.name, line, toSlack)
		} else {
			u.writeln("", "An empty message? Send some content!")
			continue
		}
		if line == "/users" {
			names := make([]string, 0, len(users))
			for _, us := range users {
				names = append(names, us.name)
			}
			broadcast("", fmt.Sprint(names), toSlack)
		}
		if line == "/all" {
			names := make([]string, 0, len(allUsers))
			for _, name := range allUsers {
				names = append(names, name)
			}
			broadcast("", fmt.Sprint(names), toSlack)
		}
		if line == "easter" {
			broadcast("", "eggs?", toSlack)
		}
		if line == "/exit" {
			return
		}
		if line == "/bell" {
			u.bell = !u.bell
			if u.bell {
				broadcast("", fmt.Sprint("bell on"), toSlack)
			} else {
				broadcast("", fmt.Sprint("bell off"), toSlack)
			}
		}
		if strings.HasPrefix(line, "/id") {
			victim, ok := findUserByName(strings.TrimSpace(strings.TrimPrefix(line, "/id")))
			if !ok {
				broadcast("", "User not found", toSlack)
			} else {
				broadcast("", victim.id, toSlack)
			}
		}
		if strings.HasPrefix(line, "/nick") {
			u.pickUsername(strings.TrimSpace(strings.TrimPrefix(line, "/nick")))
		}
		if strings.HasPrefix(line, "/banIP") {
			var pass string
			pass, err = u.term.ReadPassword("Admin password: ")
			if err != nil {
				l.Println(u.name, err)
			}
			if strings.TrimSpace(pass) == strings.TrimSpace(string(adminPass)) {
				bansMutex.Lock()
				bans = append(bans, strings.TrimSpace(strings.TrimPrefix(line, "/banIP")))
				bansMutex.Unlock()
				saveBansAndUsers()
			} else {
				u.writeln("", "Incorrect password")
			}
		} else if strings.HasPrefix(line, "/ban") {
			victim, ok := findUserByName(strings.TrimSpace(strings.TrimPrefix(line, "/ban")))
			if !ok {
				broadcast("", "User not found", toSlack)
			} else {
				var pass string
				pass, err = u.term.ReadPassword("Admin password: ")
				if err != nil {
					l.Println(u.name, err)
				}
				if strings.TrimSpace(pass) == strings.TrimSpace(string(adminPass)) {
					bansMutex.Lock()
					bans = append(bans, victim.addr)
					bansMutex.Unlock()
					saveBansAndUsers()
					victim.close(victim.name + " has been banned by " + u.name)
				} else {
					u.writeln("", "Incorrect password")
				}
			}
		}
		if strings.HasPrefix(line, "/kick") {
			victim, ok := findUserByName(strings.TrimSpace(strings.TrimPrefix(line, "/kick")))
			if !ok {
				broadcast("", "User not found", toSlack)
			} else {
				var pass string
				pass, err = u.term.ReadPassword("Admin password: ")
				if err != nil {
					l.Println(u.name, err)
				}
				if strings.TrimSpace(pass) == strings.TrimSpace(string(adminPass)) {
					victim.close(victim.name + red.Sprint(" has been kicked by ") + u.name)
				} else {
					u.writeln("", "Incorrect password")
				}
			}
		}
		if strings.HasPrefix(line, "/color") {
			colorMsg := "Which color? Choose from green, cyan, blue, red/orange, magenta/purple/pink, yellow/beige, white/cream and black/gray/grey.  \nThere's also a few secret colors :)"
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
				// secret colors
			case "easter":
				u.changeColor(*color.New(color.BgMagenta, color.FgHiYellow))
			case "baby":
				u.changeColor(*color.New(color.BgBlue, color.FgHiMagenta))
			case "l33t":
				u.changeColor(*u.color.Add(color.BgHiBlack))
			case "whiten":
				u.changeColor(*u.color.Add(color.BgWhite))
			case "hacker":
				u.changeColor(*color.New(color.FgHiGreen, color.BgBlack))
			default:
				broadcast("", colorMsg, toSlack)
			}
		}
		if line == "/people" {
			broadcast("", `
**Hack Club members**  
Zach - Founder of Hack Club  
Zachary - HC Game Designer  
Caleb, Safin, Eleeza, Jubril  
Sarthak, Anghe, Tommy, Sam  
_Possibly more people_


**From my school:**  
Kiyan, Riya


**From Twitter:**  
Ayush @ayshptk  
Srushti @srushtiuniverse  
Arav @tregsthedev  
Bereket @heybereket


**And many more have joined! (especially from HN)**`, toSlack)
		}
		if line == "/help" {
			broadcast("", `**Available commands**  
   /users   list users  
   /nick    change your name  
   /color   change your name color  
   /exit    leave the chat  
   /hide    hide messages from HC Slack  
   /bell    toggle the ansi bell  
   /id      get a unique identifier for a user  
   /all     get a list of all unique users ever  
   /people  see info about nice people who joined  
   /ban     ban a user, requires an admin pass  
   /kick    kick a user, requires an admin pass  
   /help    show this help message  
Made by Ishan Goel with feature ideas from Hack Club members.  
Thanks to Caleb Denio for lending his server!`, toSlack)
		}
	}
}

func (u *user) close(msg string) {
	u.closeOnce.Do(func() {
		usersMutex.Lock()
		users = remove(users, u)
		usersMutex.Unlock()
		broadcast("", msg, true)
		u.session.Close()
	})
}

func (u *user) writeln(senderName string, msg string) {
	msg = strings.ReplaceAll(msg, `\n`, "\n")
	if senderName != "" {
		msg = strings.TrimSpace(mdRender(msg, len([]rune(stripansi.Strip(senderName))), u.win.Width))
		msg = senderName + ": " + msg
	} else {
		msg = strings.TrimSpace(mdRender(msg, -2, u.win.Width)) // -2 so linewidth is used as is
	}
	if u.bell {
		u.term.Write([]byte("\a" + msg + "\n")) // "\a" is beep
	} else {
		u.term.Write([]byte(msg + "\n"))
	}
}

func (u *user) pickUsername(possibleName string) {
	possibleName = cleanName(possibleName)
	var err error
	for userDuplicate(possibleName) || possibleName == "" {
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
	u.changeColor(*colorArr[rand.Intn(len(colorArr))])
	allUsersMutex.Lock()
	allUsers[u.id] = u.name
	allUsersMutex.Unlock()
	saveBansAndUsers()
}

func cleanName(name string) string {
	var s string
	s = ""
	name = strings.TrimSpace(name)
	name = strings.Split(name, "\n")[0] // use only one line
	for _, r := range name {
		if unicode.IsGraphic(r) {
			s += string(r)
		}
	}
	return s
}

func mdRender(a string, nameLen int, lineWidth int) string {
	md := string(markdown.Render(a, lineWidth-(nameLen+2), 0))
	md = strings.TrimSuffix(md, "\n")
	split := strings.Split(md, "\n")
	for i := range split {
		if i == 0 {
			continue // the first line will automatically be padded
		}
		split[i] = strings.Repeat(" ", nameLen+2) + split[i]
	}
	if len(split) == 1 {
		return md
	}
	return strings.Join(split, "\n")
}

// trims space and invisible characters
func clean(a string) string {
	var s string
	s = ""
	a = strings.TrimSpace(a)
	for _, r := range a {
		if unicode.IsGraphic(r) {
			s += string(r)
		}
	}
	return s
}

func (u *user) changeColor(color color.Color) {
	u.name = color.Sprint(stripansi.Strip(u.name))
	u.color = color
	u.term.SetPrompt(u.name + ": ")
	saveBansAndUsers()
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

func main() {
	color.NoColor = false
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
		broadcast("", "Server going down! This is probably because it is being updated. Try joining in ~5 minutes.  \nIf you still can't join, make an issue at github.com/quackduck/devzat/issues", true)
		os.Exit(0)
	}()

	ssh.Handle(func(s ssh.Session) {
		u := newUser(s)
		if u == nil {
			return
		}
		u.repl()
		u.close("**" + u.name + "** **" + red.Sprint("has left the chat") + "**")
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
	go func() {
		if port == 22 {
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

func saveBansAndUsers() {
	f, err := os.Create("allusers.json")
	if err != nil {
		l.Println(err)
		return
	}
	j := json.NewEncoder(f)
	j.SetIndent("", "   ")
	j.Encode(allUsers)
	f.Close()

	f, err = os.Create("bans.json")
	if err != nil {
		l.Println(err)
		return
	}
	j = json.NewEncoder(f)
	j.SetIndent("", "   ")
	j.Encode(bans)
	f.Close()
}

func readBansAndUsers() {
	f, err := os.Open("allusers.json")
	if err != nil {
		l.Println(err)
		return
	}
	allUsersMutex.Lock()
	json.NewDecoder(f).Decode(&allUsers)
	allUsersMutex.Unlock()
	f.Close()

	f, err = os.Open("bans.json")
	if err != nil {
		l.Println(err)
		return
	}
	bansMutex.Lock()
	json.NewDecoder(f).Decode(&bans)
	bansMutex.Unlock()
	f.Close()
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
				broadcast("", "slack: "+u.RealName+": "+msg.Text, false)
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
			msg = strings.ReplaceAll(stripansi.Strip(msg), `\n`, "\n")
			rtm.SendMessage(rtm.NewOutgoingMessage(msg, "C01T5J557AA"))
		}
	}()
	return msgs
}

func findUserByName(name string) (*user, bool) {
	usersMutex.Lock()
	defer usersMutex.Unlock()
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
