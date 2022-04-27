package main

import (
	_ "embed"
	"encoding/json"
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
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/acarl005/stripansi"
	"github.com/gliderlabs/ssh"
	terminal "github.com/quackduck/term"
)

var (
	Scrollback = 16

	MainRoom         = &Room{"#main", make([]*User, 0, 10), sync.Mutex{}}
	Rooms            = map[string]*Room{MainRoom.name: MainRoom}
	Backlog          = make([]backlogMessage, 0, Scrollback)
	Bans             = make([]Ban, 0, 10)
	IDsInMinToTimes  = make(map[string]int, 10) // TODO: maybe add some IP-based factor to disallow rapid key-gen attempts
	AntispamMessages = make(map[string]int)

	Log    *log.Logger
	Devbot = "" // initialized in main
)

const (
	maxMsgLen = 5120
)

type Ban struct {
	Addr string
	ID   string
}

type Room struct {
	name       string
	users      []*User
	usersMutex sync.Mutex
}

// User represents a user connected to the SSH server.
// Exported fields represent ones saved to disk. (see also: User.savePrefs())
type User struct {
	Name     string
	Pronouns []string
	session  ssh.Session
	term     *terminal.Terminal

	room      *Room
	messaging *User // currently messaging this User in a DM

	Bell          bool
	PingEverytime bool
	isSlack       bool
	FormatTime24  bool

	Color   string
	ColorBG string
	id      string
	addr    string

	win           ssh.Window
	closeOnce     sync.Once
	lastTimestamp time.Time
	joinTime      time.Time
	Timezone      tz
}

type tz struct {
	*time.Location
}

func (t *tz) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	if s == "" { // empty string means timezone agnostic format
		t.Location = nil
		return nil
	}
	loc, err := time.LoadLocation(s)
	if err != nil {
		return err
	}
	t.Location = loc
	return nil
}

func (t *tz) MarshalJSON() ([]byte, error) {
	if t.Location == nil {
		return json.Marshal("")
	}
	return json.Marshal(t.Location.String())
}

type backlogMessage struct {
	timestamp  time.Time
	senderName string
	text       string
}

// TODO: have a web dashboard that shows logs
func main() {
	logfile, err := os.OpenFile(Config.DataDir+string(os.PathSeparator)+"log.txt", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println(err) // can't log yet so just print
		return
	}
	Log = log.New(io.MultiWriter(logfile, os.Stdout), "", log.Ldate|log.Ltime|log.Lshortfile)
	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", Config.ProfilePort), nil)
		if err != nil {
			Log.Println(err)
		}
	}()
	Devbot = Green.Paint("devbot")
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
			Log.Println("Broadcast taking too long, exiting server early.")
			os.Exit(4)
		})
		for _, r := range Rooms {
			r.broadcast(Devbot, "Server going down! This is probably because it is being updated. Try joining back immediately.  \n"+
				"If you still can't join, try joining back in 2 minutes. If you _still_ can't join, make an issue at github.com/quackduck/devzat/issues")
		}
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
				MainRoom.broadcast(Devbot, "Slap the developers in the face for me, the server almost crashed, also tell them this: "+fmt.Sprint(i)+", stack: "+string(debug.Stack()))
			}
		}()
		u.repl()
	})

	fmt.Printf("Starting chat server on port %d and profiling on port %d\n", Config.Port, Config.ProfilePort)
	go getMsgsFromSlack()
	go func() {
		fmt.Println("Also starting chat server on port", Config.AltPort)
		err := ssh.ListenAndServe(fmt.Sprintf(":%d", Config.AltPort), nil, ssh.HostKeyFile(Config.KeyFile))
		if err != nil {
			fmt.Println(err)
		}
	}()
	err = ssh.ListenAndServe(fmt.Sprintf(":%d", Config.Port), nil, ssh.HostKeyFile(Config.KeyFile), ssh.PublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
		return true // allow all keys, this lets us hash pubkeys later
	}))
	if err != nil {
		fmt.Println(err)
	}
}

func (r *Room) broadcast(senderName, msg string) {
	if msg == "" {
		return
	}
	if senderName != "" {
		SlackChan <- "[" + r.name + "] " + senderName + ": " + msg
	} else {
		SlackChan <- "[" + r.name + "] " + msg
	}
	r.broadcastNoSlack(senderName, msg)
}

func (r *Room) broadcastNoSlack(senderName, msg string) {
	if msg == "" {
		return
	}
	msg = strings.ReplaceAll(msg, "@everyone", Green.Paint("everyone\a"))
	r.usersMutex.Lock()
	for i := range r.users {
		msg = strings.ReplaceAll(msg, "@"+stripansi.Strip(r.users[i].Name), r.users[i].Name)
		msg = strings.ReplaceAll(msg, `\`+r.users[i].Name, "@"+stripansi.Strip(r.users[i].Name)) // allow escaping
	}
	for i := range r.users {
		r.users[i].writeln(senderName, msg)
	}
	r.usersMutex.Unlock()
	if r == MainRoom {
		Backlog = append(Backlog, backlogMessage{time.Now(), senderName, msg + "\n"})
		if len(Backlog) > Scrollback {
			Backlog = Backlog[len(Backlog)-Scrollback:]
		}
	}
}

func autocompleteCallback(u *User, line string, pos int, key rune) (string, int, bool) {
	if key == '\t' {
		// Autocomplete a username

		// Split the input string to look for @<name>
		words := strings.Fields(line)

		toAdd := userMentionAutocomplete(u, words)
		if toAdd != "" {
			return line + toAdd, pos + len(toAdd), true
		}
		toAdd = roomAutocomplete(u, words)
		if toAdd != "" {
			return line + toAdd, pos + len(toAdd), true
		}
	}
	return "", pos, false
}

func userMentionAutocomplete(u *User, words []string) string {
	if len(words) < 1 {
		return ""
	}
	// Check the last word and see if it's trying to refer to a User
	if words[len(words)-1][0] == '@' || (len(words)-1 == 0 && words[0][0] == '=') { // mentioning someone or dm-ing someone
		inputWord := words[len(words)-1][1:] // slice the @ or = off
		for i := range u.room.users {
			strippedName := stripansi.Strip(u.room.users[i].Name)
			toAdd := strings.TrimPrefix(strippedName, inputWord)
			if toAdd != strippedName { // there was a match, and some text got trimmed!
				return toAdd + " "
			}
		}
	}
	return ""
}

func roomAutocomplete(_ *User, words []string) string {
	// trying to refer to a room?
	if len(words) > 0 && words[len(words)-1][0] == '#' {
		// don't slice the # off, since the room name includes it
		for name := range Rooms {
			toAdd := strings.TrimPrefix(name, words[len(words)-1])
			if toAdd != name { // there was a match, and some text got trimmed!
				return toAdd + " "
			}
		}
	}
	return ""
}

func newUser(s ssh.Session) *User {
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

	u := &User{
		Name:          s.User(),
		Pronouns:      []string{"unset"},
		session:       s,
		term:          term,
		ColorBG:       "bg-off",
		Bell:          true,
		id:            shasum(toHash),
		addr:          host,
		win:           w,
		lastTimestamp: time.Now(),
		joinTime:      time.Now(),
		room:          MainRoom}

	go func() {
		for u.win = range winChan {
		}
	}()

	Log.Println("Connected " + u.Name + " [" + u.id + "]")

	if bansContains(Bans, u.addr, u.id) {
		Log.Println("Rejected " + u.Name + " [" + host + "]")
		u.writeln(Devbot, "**You are banned**. If you feel this was a mistake, please reach out at github.com/quackduck/devzat/issues or email igoel.mail@gmail.com. Please include the following information: [ID "+u.id+"]")
		u.closeQuietly()
		return nil
	}
	IDsInMinToTimes[u.id]++
	time.AfterFunc(60*time.Second, func() {
		IDsInMinToTimes[u.id]--
	})
	if IDsInMinToTimes[u.id] > 6 {
		Bans = append(Bans, Ban{u.addr, u.id})
		MainRoom.broadcast(Devbot, "`"+s.User()+"` has been banned automatically. ID: "+u.id)
		return nil
	}

	clearCMD("", u) // always clear the screen on connect
	holidaysCheck(u)

	if len(Backlog) > 0 {
		lastStamp := Backlog[0].timestamp
		u.rWriteln(printPrettyDuration(u.joinTime.Sub(lastStamp)) + " earlier")
		for i := range Backlog {
			if Backlog[i].timestamp.Sub(lastStamp) > time.Minute {
				lastStamp = Backlog[i].timestamp
				u.rWriteln(printPrettyDuration(u.joinTime.Sub(lastStamp)) + " earlier")
			}
			u.writeln(Backlog[i].senderName, Backlog[i].text)
		}
	}

	if rand.Float64() <= 0.4 { // 40% chance of being a random color
		u.changeColor("random") //nolint:errcheck // we know "random" is a valid color
	} else {
		u.changeColor(Styles[rand.Intn(len(Styles))].name) //nolint:errcheck // we know this is a valid color
	}
	if rand.Float64() <= 0.1 { // 10% chance of a random bg color
		u.changeColor("bg-random") //nolint:errcheck // we know "bg-random" is a valid color
	}

	if err := u.pickUsernameQuietly(s.User()); err != nil { // User exited or had some error
		Log.Println(err)
		s.Close()
		return nil
	}

	err := u.loadPrefs() // since we are loading for the first time, respect the saved value
	if err != nil {
		Log.Println("Could not load user:", err)
	}
	MainRoom.usersMutex.Lock()
	MainRoom.users = append(MainRoom.users, u)
	go sendCurrentUsersTwitterMessage()
	MainRoom.usersMutex.Unlock()

	u.term.SetBracketedPasteMode(true) // experimental paste bracketing support
	term.AutoCompleteCallback = func(line string, pos int, key rune) (string, int, bool) {
		return autocompleteCallback(u, line, pos, key)
	}

	switch len(MainRoom.users) - 1 {
	case 0:
		u.writeln("", Blue.Paint("Welcome to the chat. There are no more users"))
	case 1:
		u.writeln("", Yellow.Paint("Welcome to the chat. There is one more user"))
	default:
		u.writeln("", Green.Paint("Welcome to the chat. There are", strconv.Itoa(len(MainRoom.users)-1), "more users"))
	}
	MainRoom.broadcast(Devbot, u.Name+" has joined the chat")
	return u
}

// cleanupRoom deletes a room if it's empty and isn't the main room
func cleanupRoom(r *Room) {
	if r != MainRoom && len(r.users) == 0 {
		delete(Rooms, r.name)
	}
}

// Removes a User and prints Twitter and chat message
func (u *User) close(msg string) {
	u.closeOnce.Do(func() {
		u.closeQuietly()
		err := u.savePrefs()
		if err != nil {
			Log.Println(err) // not much else we can do
		}
		go sendCurrentUsersTwitterMessage()
		if time.Since(u.joinTime) > time.Minute/2 {
			msg += ". They were online for " + printPrettyDuration(time.Since(u.joinTime))
		}
		u.room.broadcast(Devbot, msg)
		u.room.users = remove(u.room.users, u)
		cleanupRoom(u.room)
	})
}

// Removes a User silently, used to close banned users
func (u *User) closeQuietly() {
	u.room.usersMutex.Lock()
	u.room.users = remove(u.room.users, u)
	u.room.usersMutex.Unlock()
	u.session.Close()
}

func (u *User) writeln(senderName string, msg string) {
	if strings.Contains(msg, u.Name) { // is a ping
		msg += "\a"
	}
	msg = strings.ReplaceAll(msg, `\n`, "\n")
	msg = strings.ReplaceAll(msg, `\`+"\n", `\n`) // let people escape newlines
	if senderName != "" {
		if strings.HasSuffix(senderName, " <- ") || strings.HasSuffix(senderName, " -> ") { // TODO: kinda hacky DM detection
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
		if u.Timezone.Location == nil {
			u.rWriteln(printPrettyDuration(time.Since(u.joinTime)) + " in")
		} else {
			if u.FormatTime24 {
				u.rWriteln(time.Now().In(u.Timezone.Location).Format("15:04"))
			} else {
				u.rWriteln(time.Now().In(u.Timezone.Location).Format("3:04 pm"))
			}
		}
		u.lastTimestamp = time.Now()
	}
	if u.PingEverytime && senderName != u.Name {
		msg += "\a"
	}
	if !u.Bell {
		msg = strings.ReplaceAll(msg, "\a", "")
	}
	_, err := u.term.Write([]byte(msg + "\n"))
	if err != nil {
		u.close(u.Name + "has left the chat because of an error writing to their terminal: " + err.Error())
	}
}

// Write to the right of the User's window
func (u *User) rWriteln(msg string) {
	if u.win.Width-lenString(msg) > 0 {
		u.term.Write([]byte(strings.Repeat(" ", u.win.Width-lenString(msg)) + msg + "\n"))
	} else {
		u.term.Write([]byte(msg + "\n"))
	}
}

// pickUsernameQuietly changes the User's username, broadcasting a name change notification if needed.
// An error is returned if the username entered had a bad word or reading input failed.
func (u *User) pickUsername(possibleName string) error {
	oldName := u.Name
	err := u.pickUsernameQuietly(possibleName)
	if err != nil {
		return err
	}
	if stripansi.Strip(u.Name) != stripansi.Strip(oldName) && stripansi.Strip(u.Name) != possibleName { // did the name change, and is it not what the User entered?
		u.room.broadcast(Devbot, oldName+" is now called "+u.Name)
	}
	return nil
}

// pickUsernameQuietly is like pickUsername but does not broadcast a name change notification.
func (u *User) pickUsernameQuietly(possibleName string) error {
	possibleName = cleanName(possibleName)
	var err error
	for {
		if possibleName == "" {
		} else if strings.HasPrefix(possibleName, "#") || possibleName == "devbot" {
			u.writeln("", "Your username is invalid. Pick a different one:")
		} else if otherUser, dup := userDuplicate(u.room, possibleName); dup {
			if otherUser == u {
				break // allow selecting the same name as before
			}
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

	possibleName = rmBadWords(possibleName)

	u.Name, _ = applyColorToData(possibleName, u.Color, u.ColorBG) //nolint:errcheck // we haven't changed the color so we know it's valid
	u.term.SetPrompt(u.Name + ": ")
	return nil
}

func (u *User) displayPronouns() string {
	result := ""
	for i := 0; i < len(u.Pronouns); i++ {
		str, _ := applyColorToData(u.Pronouns[i], u.Color, u.ColorBG)
		result += "/" + str
	}
	if result == "" {
		return result
	}
	return result[1:]
}

func (u *User) savePrefs() error {
	oldname := u.Name
	u.Name = stripansi.Strip(u.Name)
	data, err := json.Marshal(u)
	u.Name = oldname
	if err != nil {
		return err
	}
	saveTo := filepath.Join(Config.DataDir, "user-prefs")
	err = os.MkdirAll(saveTo, 0755)
	if err != nil {
		return err
	}
	saveTo = filepath.Join(saveTo, u.id+".json")
	err = os.WriteFile(saveTo, data, 0644)
	return err
}

func (u *User) loadPrefs() error {
	save := filepath.Join(Config.DataDir, "user-prefs", u.id+".json")

	data, err := os.ReadFile(save)
	if err != nil {
		if os.IsNotExist(err) { // new user, nothing saved yet
			return nil
		}
		return err
	}

	oldUser := *u //nolint:govet // complains because of a lock copy. We may need that exact lock value later on

	err = json.Unmarshal(data, u) // won't overwrite private fields
	if err != nil {
		return err
	}

	newName := u.Name
	u.Name = oldUser.Name

	newColor, newColorBG := u.Color, u.ColorBG // needed because pickUsername changes these. TODO: stop it from doing that

	err = u.pickUsername(newName)
	if err != nil {
		return err
	}
	err = u.changeColor(newColor)
	if err != nil {
		return err
	}
	err = u.changeColor(newColorBG)
	if err != nil {
		return err
	}
	return nil
}

func (u *User) changeRoom(r *Room) {
	if u.room == r {
		return
	}
	u.room.users = remove(u.room.users, u)
	u.room.broadcast("", u.Name+" is joining "+Blue.Paint(r.name)) // tell the old room
	cleanupRoom(u.room)
	u.room = r
	if _, dup := userDuplicate(u.room, u.Name); dup {
		u.pickUsername("") //nolint:errcheck // if reading input failed the next repl will err out
	}
	u.room.users = append(u.room.users, u)
	u.room.broadcast(Devbot, u.Name+" has joined "+Blue.Paint(u.room.name))
}

func (u *User) repl() {
	for {
		line, err := u.term.ReadLine()
		if err == io.EOF {
			u.close(u.Name + " has left the chat")
			return
		}

		line = getMiddlewareResult(u, line)

		line += "\n"
		hasNewlines := false
		//oldPrompt := u.Name + ": "
		for err == terminal.ErrPasteIndicator {
			hasNewlines = true
			//u.term.SetPrompt(strings.Repeat(" ", lenString(u.Name)+2))
			u.term.SetPrompt("")
			additionalLine := ""
			additionalLine, err = u.term.ReadLine()
			additionalLine = strings.ReplaceAll(additionalLine, `\n`, `\\n`)
			//additionalLine = strings.ReplaceAll(additionalLine, "\t", strings.Repeat(" ", 8))
			line += additionalLine + "\n"
		}
		if err != nil {
			Log.Println(u.Name, err)
			u.close(u.Name + " has left the chat due to an error: " + err.Error())
			return
		}
		if len(line) > maxMsgLen { // limit msg len as early as possible.
			line = line[0:maxMsgLen]
		}
		line = strings.TrimSpace(line)

		u.term.SetPrompt(u.Name + ": ")

		if hasNewlines {
			calculateLinesTaken(u, u.Name+": "+line, u.win.Width)
		} else {
			u.term.Write([]byte(strings.Repeat("\033[A\033[2K", int(math.Ceil(float64(lenString(u.Name+line)+2)/(float64(u.win.Width))))))) // basically, ceil(length of line divided by term width)
		}
		//u.term.Write([]byte(strings.Repeat("\033[A\033[2K", calculateLinesTaken(u.Name+": "+line, u.win.Width))))

		if line == "" {
			continue
		}

		AntispamMessages[u.id]++
		time.AfterFunc(5*time.Second, func() {
			AntispamMessages[u.id]--
		})
		if AntispamMessages[u.id] >= 30 {
			u.room.broadcast(Devbot, u.Name+", stop spamming or you could get banned.")
		}
		if AntispamMessages[u.id] >= 50 {
			if !bansContains(Bans, u.addr, u.id) {
				Bans = append(Bans, Ban{u.addr, u.id})
				saveBans()
			}
			u.writeln(Devbot, "anti-spam triggered")
			u.close(Red.Paint(u.Name + " has been banned for spamming"))
			return
		}
		runCommands(line, u)
	}
}

// may contain a bug ("may" because it could be the terminal's fault)
func calculateLinesTaken(u *User, s string, width int) {
	s = stripansi.Strip(s)
	//fmt.Println("`"+s+"`", "width", width)
	pos := 0
	//lines := 1
	u.term.Write([]byte("\033[A\033[2K"))
	currLine := ""
	for _, c := range s {
		pos++
		currLine += string(c)
		if c == '\t' {
			pos += 8
		}
		if c == '\n' || pos > width {
			pos = 1
			//lines++
			u.term.Write([]byte("\033[A\033[2K"))
		}
		//fmt.Println(string(c), "`"+currLine+"`", "pos", pos, "lines", lines)
	}
	//return lines
}

// bansContains reports if the addr or id is found in the bans list
func bansContains(b []Ban, addr string, id string) bool {
	for i := 0; i < len(b); i++ {
		if b[i].Addr == addr || b[i].ID == id {
			return true
		}
	}
	return false
}
