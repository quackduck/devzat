package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
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
	MainRoom                   = &Room{"#main", make([]*User, 0, 10), sync.RWMutex{}}
	Rooms                      = map[string]*Room{MainRoom.name: MainRoom}
	Backlog                    []backlogMessage
	Bans                       = make([]Ban, 0, 10)
	IDandIPsToTimesJoinedInMin = make(map[string]int, 10) // ban type has addr and id
	AntispamMessages           = make(map[string]int)

	Devbot = Green.Paint("devbot")
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
	usersMutex sync.RWMutex
}

// User represents a user connected to the SSH server.
// Exported fields represent ones saved to disk. (see also: User.savePrefs())
type User struct {
	Name     string
	Pronouns []string
	Bio      string
	session  ssh.Session
	term     *terminal.Terminal

	room      *Room
	messaging *User // currently messaging this User in a DM

	Bell          bool
	PingEverytime bool
	isBridge      bool
	FormatTime24  bool

	Color   string
	ColorBG string
	id      string
	addr    string

	winWidth      int
	lastTimestamp time.Time
	joinTime      time.Time
	lastInteract  time.Time
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
	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", Config.ProfilePort), nil)
		if err != nil {
			Log.Println(err)
		}
	}()
	readBans()
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		<-c
		fmt.Println("Shutting down...")
		saveBans()
		time.AfterFunc(time.Second, func() {
			Log.Println("Broadcast taking too long, exiting server early.")
			os.Exit(4)
		})
		for _, r := range Rooms {
			r.broadcast(Devbot, "Server going down! This is probably because it is being updated. Try joining back immediately.  \n"+
				"If you still can't join, try joining back in 2 minutes. If you _still_ can't join, make an issue at github.com/quackduck/devzat/issues")
			for _, u := range r.users {
				u.savePrefs() //nolint:errcheck
			}
		}
		os.Exit(0)
	}()
	ssh.Handle(func(s ssh.Session) {
		go keepSessionAlive(s)
		u := newUser(s)
		if u == nil {
			s.Close()
			return
		}
		defer protectFromPanic()
		u.repl()
	})

	if Config.Private {
		Log.Printf("Starting a private Devzat server on port %d and profiling on port %d\n Edit your config to change who's allowed entry.", Config.Port, Config.ProfilePort)
	} else {
		Log.Printf("Starting a Devzat server on port %d and profiling on port %d\n", Config.Port, Config.ProfilePort)
	}
	go getMsgsFromSlack()
	if !Config.Private { // allow non-sshkey logins on a non-private server
		go func() {
			fmt.Println("Also serving on port", Config.AltPort)
			err := ssh.ListenAndServe(fmt.Sprintf(":%d", Config.AltPort), nil, ssh.HostKeyFile(Config.KeyFile))
			if err != nil {
				fmt.Println(err)
			}
		}()
	}
	err := ssh.ListenAndServe(fmt.Sprintf(":%d", Config.Port), nil, ssh.HostKeyFile(Config.KeyFile),
		ssh.PublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			return true // allow all keys, this lets us hash pubkeys later
		}),
		ssh.WrapConn(func(s ssh.Context, conn net.Conn) net.Conn { // doesn't actually work for some reason?
			conn.(*net.TCPConn).SetKeepAlive(true)              //nolint:errcheck
			conn.(*net.TCPConn).SetKeepAlivePeriod(time.Minute) //nolint:errcheck
			return conn
		}),
	)
	if err != nil {
		fmt.Println(err)
	}
}

func (r *Room) broadcast(senderName, msg string) {
	if msg == "" {
		return
	}
	if Integrations.Slack != nil || Integrations.Discord != nil {
		var toSendS string
		if senderName != "" {
			if Integrations.Slack != nil {
				toSendS = "[" + r.name + "] *" + senderName + "*: " + msg
			}
		} else {
			toSendS = "[" + r.name + "] " + msg
		}
		if Integrations.Slack != nil {
			SlackChan <- toSendS
		}
		if Integrations.Discord != nil {
			DiscordChan <- DiscordMsg{
				senderName: senderName,
				msg:        msg,
				channel:    r.name,
			}
		}
	}
	r.broadcastNoBridges(senderName, msg)
}

// findMention finds mentions and colors them
func (r *Room) findMention(msg string) string {
	if len(msg) == 0 {
		return msg
	}
	maxLen := 0
	indexMax := -1

	if msg[0] == '@' {
		for i := range r.users {
			rawName := stripansi.Strip(r.users[i].Name)
			if strings.HasPrefix(msg, "@"+rawName) {
				if len(rawName) > maxLen {
					maxLen = len(rawName)
					indexMax = i
				}
			}
		}
		if indexMax != -1 { // found a mention
			return r.users[indexMax].Name + r.findMention(msg[maxLen+1:])
		}
	}

	posAt := strings.IndexByte(msg, '@')
	if posAt < 0 { // no mention
		return msg
	}
	if posAt == 0 { // if the message starts with "@" but it isn't a valid mention, we don't want to create an infinite loop
		return "@" + r.findMention(msg[1:])
	}

	if msg[posAt-1] == '\\' { // if the "@" is escaped
		return msg[0:posAt-1] + "@" + r.findMention(msg[posAt+1:])
	}

	return msg[0:posAt] + r.findMention(msg[posAt:])
}

func (r *Room) broadcastNoBridges(senderName, msg string) {
	if msg == "" {
		return
	}
	msg = r.findMention(strings.ReplaceAll(msg, "@everyone", Green.Paint("everyone\a")))
	//go func() {
	//r.usersMutex.RLock()
	for i := 0; i < len(r.users); i++ { // updates when new users join or old users leave. it is okay to read concurrently.
		r.users[i].writeln(senderName, msg)
	}
	//r.usersMutex.RUnlock()
	//}()
	if r == MainRoom && len(Backlog) > 0 {
		Backlog = Backlog[1:]
		Backlog = append(Backlog, backlogMessage{time.Now(), senderName, msg + "\n"})
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
	// remove @, =, or =@ from the start of the last word
	lastWord := words[len(words)-1]
	if len(lastWord) > 1 && lastWord[0] == '=' && lastWord[1] == '@' {
		lastWord = lastWord[2:]
	} else if lastWord[0] == '@' || lastWord[0] == '=' {
		lastWord = lastWord[1:]
	} else { // No prefix match
		return ""
	}
	// check the last word and see if it's trying to refer to a user
	for i := range u.room.users {
		strippedName := stripansi.Strip(u.room.users[i].Name)
		toAdd := strings.TrimPrefix(strippedName, lastWord)
		if toAdd != strippedName { // there was a match, and some text got trimmed!
			return toAdd + " "
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
	pty, winChan, isPty := s.Pty()
	w := pty.Window.Width
	if !isPty { // only support pty joins
		term.Write([]byte("Devzat does not allow non-pty joins. What are you trying to pull here?"))
		return nil
	}
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
		Bio:           "(none set)",
		id:            shasum(toHash),
		addr:          host,
		winWidth:      w,
		lastTimestamp: time.Now(),
		lastInteract:  time.Now(),
		joinTime:      time.Now(),
		room:          MainRoom}

	go func() {
		for win := range winChan {
			u.winWidth = win.Width
		}
	}()

	Log.Println("Connected " + u.Name + " [" + u.id + "]")

	if bansContains(Bans, u.addr, u.id) {
		Log.Println("Rejected " + u.Name + " [" + host + "] (banned)")
		u.writeln(Devbot, "**You are banned**. If you feel this was a mistake, please reach out to the server admin. Include the following information: [ID "+u.id+"]")
		s.Close()
		return nil
	}

	if Config.Private {
		_, isOnAllowlist := Config.Allowlist[u.id]
		_, isAdmin := Config.Admins[u.id]
		if !(isAdmin || isOnAllowlist) {
			Log.Println("Rejected " + u.Name + " [" + u.id + "] (not on allowlist)")
			u.writeln(Devbot, "You are not on the allowlist of this private server. If this is a mistake, send your id ("+u.id+") to the admin so that they can add you.")
			s.Close()
			return nil
		}
	}

	IDandIPsToTimesJoinedInMin[u.addr]++
	IDandIPsToTimesJoinedInMin[u.id]++
	time.AfterFunc(60*time.Second, func() {
		IDandIPsToTimesJoinedInMin[u.addr]--
		IDandIPsToTimesJoinedInMin[u.id]--
	})
	if IDandIPsToTimesJoinedInMin[u.addr] > 6 || IDandIPsToTimesJoinedInMin[u.id] > 6 {
		u.ban("")
		MainRoom.broadcast(Devbot, u.Name+" has been banned automatically. ID: "+u.id)
		return nil
	}

	clearCMD("", u) // always clear the screen on connect
	holidaysCheck(u)

	if rand.Float64() <= 0.4 { // 40% chance of being a random color
		u.changeColor("random") //nolint:errcheck // we know "random" is a valid color
	} else {
		u.changeColor(Styles[rand.Intn(len(Styles))].name) //nolint:errcheck // we know this is a valid color
	}
	if rand.Float64() <= 0.1 { // 10% chance of a random bg color
		u.changeColor("bg-random") //nolint:errcheck // we know "bg-random" is a valid color
	}

	timeoutChan := make(chan bool)
	timedOut := false
	go func() { // timeout to minimize inactive connections
		err := u.loadPrefs()
		if err != nil && !timedOut {
			Log.Println("Could not load user:", err)
			return
		}
		if timedOut {
			return
		}
		if err = u.pickUsernameQuietly(stripansi.Strip(u.Name)); err != nil && !timedOut {
			Log.Println(err)
			s.Close()
			s = nil // marker so we know to exit
		}
		timeoutChan <- true
	}()

	select {
	case <-time.After(time.Minute):
		Log.Println("Timeout for user", stripansi.Strip(u.Name), "with ID", u.id)
		timedOut = true
		s.Close()
		return nil
	case <-timeoutChan:
		if s == nil {
			return nil
		}
	}

	if !Config.Private { // sensitive info might be shared on a private server
		var lastStamp time.Time
		for i := range Backlog {
			if Backlog[i].text == "" { // skip empty entries
				continue
			}
			if i == 0 || Backlog[i].timestamp.Sub(lastStamp) > time.Minute {
				lastStamp = Backlog[i].timestamp
				u.rWriteln(fmtTime(u, lastStamp))
			}
			u.writeln(Backlog[i].senderName, Backlog[i].text)
		}
		if time.Since(lastStamp) > time.Minute && u.Timezone.Location != nil {
			u.rWriteln(fmtTime(u, time.Now()))
		}
	}

	MainRoom.usersMutex.Lock()
	MainRoom.users = append(MainRoom.users, u)
	MainRoom.usersMutex.Unlock()
	go sendCurrentUsersTwitterMessage()

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
	MainRoom.broadcast("", Green.Paint(" --> ")+u.Name+" has joined the chat")
	return u
}

// cleanupRoomInstant deletes a room if it's empty and isn't the main room
func cleanupRoomInstant(r *Room) {
	if r != MainRoom && r != nil && len(r.users) == 0 {
		delete(Rooms, r.name)
	}
}

var cleanupMap = make(map[*Room]chan bool, 5)

func cleanupRoom(r *Room) {
	if ch, ok := cleanupMap[r]; ok {
		ch <- true // reset timer
		return
	}
	go func() {
		ch := make(chan bool) // no buffer needed
		cleanupMap[r] = ch
		timer := time.NewTimer(time.Hour * 24)
		for {
			select {
			case <-ch: // need a reset?
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(time.Hour * 24)
				// no return, carry on to the next select
			case <-timer.C:
				delete(cleanupMap, r)
				timer.Stop()
				cleanupRoomInstant(r)
				return // done!
			}
		}
	}()
}

// Removes a User and prints a chat message
func (u *User) close(msg string) {
	u.room.usersMutex.Lock()
	u.room.users = remove(u.room.users, u)
	u.room.usersMutex.Unlock()
	cleanupRoom(u.room)
	if u.isBridge {
		return
	}
	u.session.Close()
	u.session = nil
	err := u.savePrefs()
	if err != nil {
		Log.Println(err) // not much else we can do
	}
	if msg == "" {
		return
	}
	if time.Since(u.joinTime) > time.Minute/2 {
		msg += ". They were online for " + printPrettyDuration(time.Since(u.joinTime))
	}
	u.room.broadcast("", Red.Paint(" <-- ")+msg)
}

func (u *User) ban(banner string) {
	Bans = append(Bans, Ban{u.addr, u.id})
	saveBans()
	uid := u.id
	u.close(banner)
	for i := range Rooms { // close all users that have this id (including this user)
		for j := 0; j < len(Rooms[i].users); j++ {
			if Rooms[i].users[j].id == uid {
				Rooms[i].users[j].close("")
				j--
			}
		}
	}
}

func (u *User) writeln(senderName string, msg string) {
	if strings.Contains(msg, u.Name) { // is a ping
		msg += "\a"
	}
	msg = strings.ReplaceAll(msg, `\n`, "\n")
	msg = strings.ReplaceAll(msg, `\`+"\n", `\n`) // let people escape newlines
	thisUserIsDMSender := strings.HasSuffix(senderName, " <- ")
	if senderName != "" {
		if thisUserIsDMSender || strings.HasSuffix(senderName, " -> ") { // TODO: kinda hacky DM detection
			msg = strings.TrimSpace(mdRender(msg, lenString(senderName), u.winWidth))
			msg = senderName + msg
			if !thisUserIsDMSender {
				msg += "\a"
			}
		} else {
			msg = strings.TrimSpace(mdRender(msg, lenString(senderName)+2, u.winWidth))
			msg = senderName + ": " + msg
		}
	} else {
		msg = strings.TrimSpace(mdRender(msg, 0, u.winWidth)) // No sender
	}
	if time.Since(u.lastTimestamp) > time.Minute {
		u.lastTimestamp = time.Now()
		u.rWriteln(fmtTime(u, u.lastTimestamp))
	}
	if u.PingEverytime && senderName != u.Name && !thisUserIsDMSender {
		msg += "\a"
	}
	if !u.Bell {
		msg = strings.ReplaceAll(msg, "\a", "")
	}
	_, err := u.term.Write([]byte(msg + "\n"))
	if err != nil {
		u.close(u.Name + " has left the chat because of an error writing to their terminal: " + err.Error())
	}
}

// Write to the right of the User's window
func (u *User) rWriteln(msg string) {
	if u.winWidth-lenString(msg) > 0 {
		u.term.Write([]byte(strings.Repeat(" ", u.winWidth-lenString(msg)) + msg + "\n"))
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
		if possibleName == "" || strings.HasPrefix(possibleName, "#") || possibleName == "devbot" || strings.HasPrefix(possibleName, "@") {
			u.writeln("", "Your username is invalid. Pick a different one:")
		} else if otherUser, dup := userDuplicate(u.room, possibleName); dup {
			if otherUser == u {
				break // allow selecting the same name as before the user tried to change it
			}
			u.writeln("", "Your username is already in use. Pick a different one:")
		} else { // valid name
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

	err = u.pickUsernameQuietly(newName)
	if err != nil {
		return err
	}
	err = u.changeColor(u.Color)
	if err != nil {
		return err
	}
	err = u.changeColor(u.ColorBG)
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
	u.room.broadcast("", Green.Paint(" --> ")+u.Name+" has joined "+Blue.Paint(u.room.name))
}

func (u *User) repl() {
	for {
		u.lastInteract = time.Now()
		line, err := u.term.ReadLine()
		if err == io.EOF {
			if u.session != nil {
				u.close(u.Name + " has left the chat")
			}
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
			calculateLinesTaken(u, u.Name+": "+line, u.winWidth)
		} else {
			u.term.Write([]byte(strings.Repeat("\033[A\033[2K", int(math.Ceil(float64(lenString(u.Name+line)+2)/(float64(u.winWidth))))))) // basically, ceil(length of line divided by term width)
		}

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
